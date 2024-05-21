package ldevents

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldlog"
)

type eventsOutbox struct {
	events           []anyEventOutput
	summarizer       eventSummarizer
	capacity         int
	capacityExceeded bool
	droppedEvents    int
	loggers          ldlog.Loggers
}

func newEventsOutbox(capacity int, loggers ldlog.Loggers) *eventsOutbox {
	return &eventsOutbox{
		events:     make([]anyEventOutput, 0, capacity),
		summarizer: newEventSummarizer(),
		capacity:   capacity,
		loggers:    loggers,
	}
}

func (b *eventsOutbox) addEvent(event anyEventInput) {
	if len(b.events) >= b.capacity {
		if !b.capacityExceeded {
			b.capacityExceeded = true
			b.loggers.Warn("Exceeded event queue capacity. Increase capacity to avoid dropping events.")
		}
		b.droppedEvents++
		return
	}
	b.capacityExceeded = false
	b.events = append(b.events, event)
}

func (b *eventsOutbox) addToSummary(ed EvaluationData) {
	b.summarizer.summarizeEvent(ed)
}

func (b *eventsOutbox) getPayload() flushPayload {
	var copied []anyEventOutput
	if len(b.events) > 0 {
		copied = make([]anyEventOutput, len(b.events))
		copy(copied, b.events)
	}
	return flushPayload{
		events:  copied,
		summary: b.summarizer.snapshot(),
	}
}

func (b *eventsOutbox) clear() {
	for i := range b.events {
		b.events[i] = nil
	}
	b.events = b.events[0:0]
	b.summarizer.reset()
}
