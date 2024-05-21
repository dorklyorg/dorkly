package ldevents

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/launchdarkly/go-jsonstream/v3/jwriter"
	"github.com/launchdarkly/go-sdk-common/v3/ldlog"
	"github.com/launchdarkly/go-sdk-common/v3/ldsampling"
	"github.com/launchdarkly/go-sdk-common/v3/ldtime"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
)

// anyEventInput and anyEventOutput only exist to make it a little clearer in the code whether we're referring
// to something in the inbox or the outbox.
type anyEventInput interface{}
type anyEventOutput interface{}

type defaultEventProcessor struct {
	inboxCh       chan eventDispatcherMessage
	inboxFullOnce sync.Once
	closeOnce     sync.Once
	loggers       ldlog.Loggers
}

type eventDispatcher struct {
	config               EventsConfiguration
	outbox               *eventsOutbox
	flushCh              chan *flushPayload
	senderResultCh       chan EventSenderResult
	workersGroup         *sync.WaitGroup
	userKeys             lruCache
	lastKnownPastTime    ldtime.UnixMillisecondTime
	deduplicatedContexts int
	eventsInLastBatch    int
	disabled             bool
	currentTimestampFn   func() ldtime.UnixMillisecondTime
	sampler              *ldsampling.RatioSampler
}

type flushPayload struct {
	diagnosticEvent ldvalue.Value
	events          []anyEventOutput
	summary         eventSummary
}

// Payload of the inboxCh channel.
type eventDispatcherMessage interface{}

type sendEventMessage struct {
	event anyEventInput
}

type flushEventsMessage struct {
	replyCh chan struct{}
}

type shutdownEventsMessage struct {
	replyCh chan struct{}
}

type syncEventsMessage struct {
	replyCh chan struct{}
}

const (
	maxFlushWorkers = 5
)

// NewDefaultEventProcessor creates an instance of the default implementation of analytics event processing.
func NewDefaultEventProcessor(config EventsConfiguration) EventProcessor {
	inboxCh := make(chan eventDispatcherMessage, config.Capacity)
	startEventDispatcher(config, inboxCh)
	return &defaultEventProcessor{
		inboxCh: inboxCh,
		loggers: config.Loggers,
	}
}

func (ep *defaultEventProcessor) RecordEvaluation(ed EvaluationData) {
	ep.postNonBlockingMessageToInbox(sendEventMessage{event: ed})
}

func (ep *defaultEventProcessor) RecordIdentifyEvent(e IdentifyEventData) {
	ep.postNonBlockingMessageToInbox(sendEventMessage{event: e})
}

func (ep *defaultEventProcessor) RecordCustomEvent(e CustomEventData) {
	ep.postNonBlockingMessageToInbox(sendEventMessage{event: e})
}

func (ep *defaultEventProcessor) RecordMigrationOpEvent(e MigrationOpEventData) {
	ep.postNonBlockingMessageToInbox(sendEventMessage{event: e})
}

func (ep *defaultEventProcessor) RecordRawEvent(data json.RawMessage) {
	ep.postNonBlockingMessageToInbox(sendEventMessage{event: rawEvent{data: data}})
}

func (ep *defaultEventProcessor) Flush() {
	ep.postNonBlockingMessageToInbox(flushEventsMessage{})
}

func (ep *defaultEventProcessor) FlushBlocking(timeout time.Duration) bool {
	replyCh := make(chan struct{}, 1)
	m := flushEventsMessage{replyCh: replyCh}
	ep.inboxCh <- m
	var deadline <-chan time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		deadline = timer.C
	}
	select {
	case <-m.replyCh:
		return true
	case <-deadline:
		return false
	}
}

func (ep *defaultEventProcessor) postNonBlockingMessageToInbox(e eventDispatcherMessage) {
	select {
	case ep.inboxCh <- e:
		return
	default: // COVERAGE: no way to simulate this condition in unit tests
	}
	// If the inbox is full, it means the eventDispatcher is seriously backed up with not-yet-processed events.
	// This is unlikely, but if it happens, it means the application is probably doing a ton of flag evaluations
	// across many goroutines-- so if we wait for a space in the inbox, we risk a very serious slowdown of the
	// app. To avoid that, we'll just drop the event. The log warning about this will only be shown once.
	ep.inboxFullOnce.Do(func() { // COVERAGE: no way to simulate this condition in unit tests
		ep.loggers.Warn("Events are being produced faster than they can be processed; some events will be dropped")
	})
}

func (ep *defaultEventProcessor) Close() error {
	ep.closeOnce.Do(func() {
		// We put the flush and shutdown messages directly into the channel instead of calling
		// postNonBlockingMessageToInbox, because we *do* want to block to make sure there is room in the channel;
		// these aren't analytics events, they are messages that are necessary for an orderly shutdown.
		ep.inboxCh <- flushEventsMessage{}
		m := shutdownEventsMessage{replyCh: make(chan struct{})}
		ep.inboxCh <- m
		<-m.replyCh
	})
	return nil
}

func startEventDispatcher(
	config EventsConfiguration,
	inboxCh <-chan eventDispatcherMessage,
) {
	ed := &eventDispatcher{
		config:             config,
		outbox:             newEventsOutbox(config.Capacity, config.Loggers),
		flushCh:            make(chan *flushPayload, 1),
		senderResultCh:     make(chan EventSenderResult, maxFlushWorkers),
		workersGroup:       &sync.WaitGroup{},
		userKeys:           newLruCache(config.UserKeysCapacity),
		currentTimestampFn: config.currentTimeProvider,
		sampler:            ldsampling.NewSampler(),
	}

	if ed.currentTimestampFn == nil {
		ed.currentTimestampFn = ldtime.UnixMillisNow
	}

	formatter := &eventOutputFormatter{
		contextFormatter: newEventContextFormatter(config),
		config:           config,
	}

	// Start a fixed-size pool of workers that wait on flushTriggerCh. This is the
	// maximum number of flushes we can do concurrently.
	for i := 0; i < maxFlushWorkers; i++ {
		go runFlushTask(config, formatter, ed.flushCh, ed.workersGroup, ed.senderResultCh)
	}
	if config.DiagnosticsManager != nil {
		event := config.DiagnosticsManager.CreateInitEvent()
		ed.sendDiagnosticsEvent(event)
	}
	go ed.runMainLoop(inboxCh)
}

func (ed *eventDispatcher) runMainLoop(
	inboxCh <-chan eventDispatcherMessage,
) {
	if err := recover(); err != nil { // COVERAGE: no way to simulate this condition in unit tests
		ed.config.Loggers.Errorf("Unexpected panic in event processing thread: %+v", err)
	}

	flushInterval := ed.config.FlushInterval
	if flushInterval <= 0 { // COVERAGE: no way to test this logic in unit tests
		flushInterval = DefaultFlushInterval
	}
	userKeysFlushInterval := ed.config.UserKeysFlushInterval
	if userKeysFlushInterval <= 0 { // COVERAGE: no way to test this logic in unit tests
		userKeysFlushInterval = DefaultUserKeysFlushInterval
	}
	flushTicker := time.NewTicker(flushInterval)
	usersResetTicker := time.NewTicker(userKeysFlushInterval)

	var diagnosticsTicker *time.Ticker
	var diagnosticsTickerCh <-chan time.Time
	diagnosticsManager := ed.config.DiagnosticsManager
	if diagnosticsManager != nil {
		interval := ed.config.DiagnosticRecordingInterval
		if interval > 0 {
			if interval < MinimumDiagnosticRecordingInterval { // COVERAGE: no way to test this logic in unit tests
				interval = DefaultDiagnosticRecordingInterval
			}
		} else {
			if ed.config.forceDiagnosticRecordingInterval > 0 {
				interval = ed.config.forceDiagnosticRecordingInterval
			} else {
				interval = DefaultDiagnosticRecordingInterval
			}
		}
		diagnosticsTicker = time.NewTicker(interval)
		diagnosticsTickerCh = diagnosticsTicker.C
	}

	for {
		// Drain the response channel with a higher priority than anything else
		// to ensure that the flush workers don't get blocked.
		select {
		case message := <-inboxCh:
			switch m := message.(type) {
			case sendEventMessage:
				ed.processEvent(m.event)
			case flushEventsMessage:
				ed.triggerFlush()
				if m.replyCh != nil {
					ed.workersGroup.Wait() // Wait for all in-progress flushes to complete
					m.replyCh <- struct{}{}
				}
			case syncEventsMessage:
				ed.workersGroup.Wait()
				m.replyCh <- struct{}{}
			case shutdownEventsMessage:
				flushTicker.Stop()
				usersResetTicker.Stop()
				if diagnosticsTicker != nil {
					diagnosticsTicker.Stop()
				}
				ed.workersGroup.Wait() // Wait for all in-progress flushes to complete
				close(ed.flushCh)      // Causes all idle flush workers to terminate
				close(ed.senderResultCh)
				m.replyCh <- struct{}{}
				return
			}
		case result := <-ed.senderResultCh:
			switch {
			case ed.disabled: // COVERAGE: no way to simulate in unit tests
				continue
			case result.MustShutDown:
				ed.disabled = true
				ed.outbox.clear()
			case result.TimeFromServer > 0:
				ed.lastKnownPastTime = result.TimeFromServer
			}
		case <-flushTicker.C:
			ed.triggerFlush()
		case <-usersResetTicker.C:
			ed.userKeys.clear()
		case <-diagnosticsTickerCh:
			if diagnosticsManager == nil || !diagnosticsManager.CanSendStatsEvent() {
				// COVERAGE: no way to test this logic in unit tests
				break
			}
			event := diagnosticsManager.CreateStatsEventAndReset(
				ed.outbox.droppedEvents,
				ed.deduplicatedContexts,
				ed.eventsInLastBatch,
			)
			ed.outbox.droppedEvents = 0
			ed.deduplicatedContexts = 0
			ed.eventsInLastBatch = 0
			ed.sendDiagnosticsEvent(event)
		}
	}
}

func (ed *eventDispatcher) processEvent(evt anyEventInput) {
	if ed.disabled {
		return
	}

	var samplingRatio ldvalue.OptionalInt

	// Decide whether to add the event to the payload. Feature events may be added twice, once for
	// the event (if tracked) and once for debugging.
	willAddFullEvent := true
	var debugEvent anyEventInput
	inlinedUser := false
	var eventContext EventInputContext
	var creationDate ldtime.UnixMillisecondTime
	switch evt := evt.(type) {
	case EvaluationData:
		samplingRatio = evt.SamplingRatio
		if evt.ForceSampling {
			samplingRatio = ldvalue.NewOptionalInt(1)
		}

		eventContext = evt.Context
		creationDate = evt.CreationDate

		// add all feature events to summaries, provided we aren't specifically
		// excluding them.
		if !evt.ExcludeFromSummaries {
			ed.outbox.addToSummary(evt)
		}

		willAddFullEvent = evt.RequireFullEvent
		if ed.shouldDebugEvent(&evt) {
			de := evt
			de.debug = true
			debugEvent = de
		}
	case IdentifyEventData:
		samplingRatio = evt.SamplingRatio
		if evt.ForceSampling {
			samplingRatio = ldvalue.NewOptionalInt(1)
		}

		eventContext = evt.Context
		creationDate = evt.CreationDate
		inlinedUser = true
	case CustomEventData:
		samplingRatio = evt.SamplingRatio
		if evt.ForceSampling {
			samplingRatio = ldvalue.NewOptionalInt(1)
		}

		eventContext = evt.Context
		creationDate = evt.CreationDate
	case MigrationOpEventData:
		samplingRatio = evt.SamplingRatio
		if evt.ForceSampling {
			samplingRatio = ldvalue.NewOptionalInt(1)
		}

		if ed.shouldSample(samplingRatio) {
			ed.outbox.addEvent(evt)
		}
		// We can halt execution here as a migration event shouldn't generate an index or debug event.
		return
	default:
		ed.outbox.addEvent(evt)
		return
	}
	// For each context we haven't seen before, we add an index event before the event that referenced
	// the context - unless the original event will contain an inline context (e.g. an identify event).
	alreadySeenUser := ed.userKeys.add(eventContext.context.FullyQualifiedKey())
	if !(willAddFullEvent && inlinedUser) {
		if alreadySeenUser {
			ed.deduplicatedContexts++
		} else {
			indexEvent := indexEvent{
				BaseEvent{CreationDate: creationDate, Context: eventContext},
			}
			ed.outbox.addEvent(indexEvent)
		}
	}
	if willAddFullEvent && ed.shouldSample(samplingRatio) {
		ed.outbox.addEvent(evt)
	}
	if debugEvent != nil && ed.shouldSample(samplingRatio) {
		ed.outbox.addEvent(debugEvent)
	}
}

func (ed *eventDispatcher) shouldDebugEvent(evt *EvaluationData) bool {
	if evt.DebugEventsUntilDate == 0 {
		return false
	}
	// The "last known past time" comes from the last HTTP response we got from the server.
	// In case the client's time is set wrong, at least we know that any expiration date
	// earlier than that point is definitely in the past.  If there's any discrepancy, we
	// want to err on the side of cutting off event debugging sooner.
	return evt.DebugEventsUntilDate > ed.lastKnownPastTime &&
		evt.DebugEventsUntilDate > ed.currentTimestampFn()
}

// Signal that we would like to do a flush as soon as possible.
func (ed *eventDispatcher) triggerFlush() {
	if ed.disabled {
		return
	}
	// Is there anything to flush?
	payload := ed.outbox.getPayload()
	totalEventCount := len(payload.events)
	if payload.summary.hasCounters() {
		totalEventCount++
	}
	if totalEventCount == 0 {
		ed.eventsInLastBatch = 0
		return
	}
	ed.workersGroup.Add(1) // Increment the count of active flushes
	select {
	case ed.flushCh <- &payload:
		// If the channel wasn't full, then there is a worker available who will pick up
		// this flush payload and send it. The event outbox and summary state can now be
		// cleared from the main goroutine.
		ed.eventsInLastBatch = totalEventCount
		ed.outbox.clear()
	default:
		// We can't start a flush right now because we're waiting for one of the workers
		// to pick up the last one.  Do not reset the event outbox or summary state.
		ed.workersGroup.Done()
	}
}

func (ed *eventDispatcher) sendDiagnosticsEvent(
	event ldvalue.Value,
) {
	payload := flushPayload{diagnosticEvent: event}
	ed.workersGroup.Add(1) // Increment the count of active flushes
	select {
	case ed.flushCh <- &payload:
		// If the channel wasn't full, then there is a worker available who will pick up
		// this flush payload and send it.
	default:
		// We can't start a flush right now because we're waiting for one of the workers
		// to pick up the last one. We'll just discard this diagnostic event - presumably
		// we'll send another one later anyway, and we don't want this kind of nonessential
		// data to cause any kind of back-pressure.
		ed.workersGroup.Done() // COVERAGE: no way to simulate this condition in unit tests
	}
}

func (ed *eventDispatcher) shouldSample(ratio ldvalue.OptionalInt) bool {
	if ed.config.forceSampling {
		return true
	}

	return ed.sampler.Sample(ratio.OrElse(1))
}

func runFlushTask(config EventsConfiguration, formatter *eventOutputFormatter, flushCh <-chan *flushPayload,
	workersGroup *sync.WaitGroup, senderResultCh chan<- EventSenderResult) {
	for {
		payload, more := <-flushCh
		if !more {
			// Channel has been closed - we're shutting down
			break
		}
		if !payload.diagnosticEvent.IsNull() {
			w := jwriter.NewWriter()
			payload.diagnosticEvent.WriteToJSONWriter(&w)
			bytes := w.Bytes()
			_ = config.EventSender.SendEventData(DiagnosticEventDataKind, bytes, 1)
		} else {
			bytes, count := formatter.makeOutputEvents(payload.events, payload.summary)
			if len(bytes) > 0 {
				result := config.EventSender.SendEventData(AnalyticsEventDataKind, bytes, count)
				senderResultCh <- result
			}
		}
		workersGroup.Done() // Decrement the count of in-progress flushes
	}
}
