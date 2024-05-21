package ldevents

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/launchdarkly/go-sdk-common/v3/ldlog"
	"github.com/launchdarkly/go-sdk-common/v3/ldtime"

	"github.com/google/uuid"
)

const (
	defaultEventsURI   = "https://events.launchdarkly.com"
	eventSchemaHeader  = "X-LaunchDarkly-Event-Schema"
	payloadIDHeader    = "X-LaunchDarkly-Payload-ID"
	currentEventSchema = "4"
	defaultRetryDelay  = time.Second
)

// EventSenderConfiguration contains parameters for event delivery that do not vary from one event payload to another.
type EventSenderConfiguration struct {
	// Client is the HTTP client instance to use, or nil to use http.DefaultClient.
	Client *http.Client
	// BaseURI is the base URI to which the event endpoint paths will be added.
	BaseURI string
	// BaseHeaders contains any headers that should be added to the HTTP request, other than the schema version.
	// The event delivery logic will never modify this map; it will clone it if necessary.
	BaseHeaders func() http.Header
	// SchemaVersion specifies the value for the X-LaunchDarkly-Event-Schema header, or 0 to use the latest version.
	SchemaVersion int
	// Loggers is used for logging event delivery status.
	Loggers ldlog.Loggers
	// RetryDelay is the length of time to wait for a retry, or 0 to use the default delay (1 second).
	RetryDelay time.Duration
}

type defaultEventSender struct {
	config EventSenderConfiguration
}

// NewServerSideEventSender creates the standard implementation of EventSender for server-side SDKs.
//
// The underlying behavior is mostly provided by SendEventData. It adds the behavior of putting the SDK key in an
// Authorization header (in addition to any other headers in config.BaseHeaders), and it forces the schema version
// to be the latest schema version (since, in the regular use case of EventSender being used within a
// DefaultEventProcessor, the latter is only ever going to generate output in the current schema).
//
// This object maintains no state other than its configuration, so discarding it does not require any special
// cleanup.
func NewServerSideEventSender(
	config EventSenderConfiguration,
	sdkKey string,
) EventSender {
	realConfig := config
	realConfig.SchemaVersion = 0 // defaults to current
	realConfig.BaseHeaders = func() http.Header {
		var base http.Header
		if config.BaseHeaders != nil {
			base = config.BaseHeaders()
		}
		ret := make(http.Header, len(base)+1)
		for k, vv := range base {
			ret[k] = vv
		}
		ret.Set("Authorization", sdkKey)
		return ret
	}
	return &defaultEventSender{
		config: realConfig,
	}
}

func (s *defaultEventSender) SendEventData(kind EventDataKind, data []byte, eventCount int) EventSenderResult {
	return SendEventDataWithRetry(
		s.config,
		kind,
		"",
		data,
		eventCount,
	)
}

// SendEventDataWithRetry provides an entry point to the same event delivery logic that is used by DefaultEventSender.
// This is exported separately for convenience in code such as the Relay Proxy which needs to implement the same
// behavior in situations where EventProcessor and EventSender are not relevant. The behavior provided is specifically:
//
// 1. Add headers as follows, besides config.BaseHeaders: Content-Type (application/json); X-LaunchDarkly-Schema-Version
// (based on config.Schema Version; omitted for diagnostic events); and X-LaunchDarkly-Payload-ID (a UUID value).
// Unlike NewServerSideEventSender, it does not add an Authorization header.
//
// 2. If delivery fails with a recoverable error, such as an HTTP 503 or an I/O error, retry exactly once after a delay
// configured by config.RetryDelay. This is done synchronously. If the retry fails, return Success: false.
//
// 3. If delivery fails with an unrecoverable error, such as an HTTP 401, return Success: false and MustShutDown: true.
//
// 4. If the response has a Date header, parse it into TimeFromServer.
//
// The overridePath parameter only needs to be set if you need to customize the URI path. If it is empty, the standard
// path of /bulk or /diagnostic will be used as appropriate.
func SendEventDataWithRetry(
	config EventSenderConfiguration,
	kind EventDataKind,
	overridePath string,
	data []byte,
	eventCount int,
) EventSenderResult {
	headers := make(http.Header)
	if config.BaseHeaders != nil {
		for k, vv := range config.BaseHeaders() {
			headers[k] = vv
		}
	}
	headers.Set("Content-Type", "application/json")

	var uri, path, description string

	switch kind {
	case AnalyticsEventDataKind:
		path = "/bulk"
		description = fmt.Sprintf("%d events", eventCount)
		if config.SchemaVersion == 0 {
			headers.Add(eventSchemaHeader, currentEventSchema)
		} else {
			headers.Add(eventSchemaHeader, strconv.Itoa(config.SchemaVersion))
		}
		payloadUUID, _ := uuid.NewRandom()
		headers.Add(payloadIDHeader, payloadUUID.String())
		// if NewRandom somehow failed, we'll just proceed with an empty string
	case DiagnosticEventDataKind:
		path = "/diagnostic"
		description = "diagnostic event"
	default:
		return EventSenderResult{}
	}

	if overridePath != "" {
		path = "/" + strings.TrimLeft(overridePath, "/")
	}
	baseURI := strings.TrimRight(config.BaseURI, "/")
	if baseURI == "" {
		baseURI = defaultEventsURI
	}
	uri = baseURI + path

	config.Loggers.Debugf("Sending %s: %s", description, data)

	var resp *http.Response
	var respErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			delay := config.RetryDelay
			if delay == 0 {
				delay = defaultRetryDelay // COVERAGE: unit tests always set a short delay
			}
			config.Loggers.Warnf("Will retry posting events after %f second", float64(delay/time.Second))
			time.Sleep(delay)
		}
		req, reqErr := http.NewRequest("POST", uri, bytes.NewReader(data))
		if reqErr != nil { // COVERAGE: no way to simulate this condition in unit tests
			config.Loggers.Errorf("Unexpected error while creating event request: %+v", reqErr)
			return EventSenderResult{}
		}
		req.Header = headers

		client := config.Client
		if client == nil {
			client = http.DefaultClient
		}
		resp, respErr = client.Do(req)

		if resp != nil && resp.Body != nil {
			_, _ = ioutil.ReadAll(resp.Body)
			_ = resp.Body.Close()
		}

		if respErr != nil {
			config.Loggers.Warnf("Unexpected error while sending events: %+v", respErr)
			continue
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			result := EventSenderResult{Success: true}
			t, err := http.ParseTime(resp.Header.Get("Date"))
			if err == nil {
				result.TimeFromServer = ldtime.UnixMillisFromTime(t)
			}
			return result
		}
		if isHTTPErrorRecoverable(resp.StatusCode) {
			maybeRetry := "will retry"
			if attempt == 1 {
				maybeRetry = "some events were dropped"
			}
			config.Loggers.Warnf(httpErrorMessage(resp.StatusCode, "sending events", maybeRetry))
		} else {
			config.Loggers.Warnf(httpErrorMessage(resp.StatusCode, "sending events", ""))
			// Large payloads mean this particular request is a failure, but
			// that doesn't mean subsequent payloads won't be small enough to
			// succeed.
			tooLarge := resp.StatusCode == http.StatusRequestEntityTooLarge
			return EventSenderResult{MustShutDown: !tooLarge}
		}
	}
	return EventSenderResult{}
}
