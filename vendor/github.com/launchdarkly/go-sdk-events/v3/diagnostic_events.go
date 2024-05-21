package ldevents

import (
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/launchdarkly/go-sdk-common/v3/ldtime"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
)

type diagnosticStreamInitInfo struct {
	timestamp      ldtime.UnixMillisecondTime
	failed         bool
	durationMillis uint64
}

// DiagnosticsManager is an object that maintains state for diagnostic events and produces JSON data.
//
// The format of the JSON event data is subject to change. Diagnostic events are represented opaquely with the
// Value type.
type DiagnosticsManager struct {
	id                ldvalue.Value
	configData        ldvalue.Value
	sdkData           ldvalue.Value
	startTime         ldtime.UnixMillisecondTime
	dataSinceTime     ldtime.UnixMillisecondTime
	streamInits       []diagnosticStreamInitInfo
	periodicEventGate <-chan struct{}
	lock              sync.Mutex
}

// NewDiagnosticID creates a unique identifier for this SDK instance.
func NewDiagnosticID(sdkKey string) ldvalue.Value {
	uuid, _ := uuid.NewRandom()
	var sdkKeySuffix string
	if len(sdkKey) > 6 {
		sdkKeySuffix = sdkKey[len(sdkKey)-6:]
	} else {
		sdkKeySuffix = sdkKey
	}
	return ldvalue.ObjectBuild().
		SetString("diagnosticId", uuid.String()).
		SetString("sdkKeySuffix", sdkKeySuffix).
		Build()
}

// NewDiagnosticsManager creates an instance of DiagnosticsManager.
func NewDiagnosticsManager(
	id ldvalue.Value,
	configData ldvalue.Value,
	sdkData ldvalue.Value,
	startTime time.Time,
	periodicEventGate <-chan struct{}, // periodicEventGate is test instrumentation - see CanSendStatsEvent
) *DiagnosticsManager {
	timestamp := ldtime.UnixMillisFromTime(startTime)
	m := &DiagnosticsManager{
		id:                id,
		configData:        configData,
		sdkData:           sdkData,
		startTime:         timestamp,
		dataSinceTime:     timestamp,
		periodicEventGate: periodicEventGate,
	}
	return m
}

// RecordStreamInit is called by the stream processor when a stream connection has either succeeded or failed.
func (m *DiagnosticsManager) RecordStreamInit(
	timestamp ldtime.UnixMillisecondTime,
	failed bool,
	durationMillis uint64,
) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.streamInits = append(m.streamInits, diagnosticStreamInitInfo{
		timestamp:      timestamp,
		failed:         failed,
		durationMillis: durationMillis,
	})
}

// CreateInitEvent is called by DefaultEventProcessor to create the initial diagnostics event that includes the
// configuration.
func (m *DiagnosticsManager) CreateInitEvent() ldvalue.Value {
	// Notes on platformData
	// - osArch: in Go, GOARCH is set at compile time, not at runtime (unlike GOOS, whiich is runtime).
	// - osVersion: Go provides no portable way to get this property.
	platformData := ldvalue.ObjectBuild().
		SetString("name", "Go").
		SetString("goVersion", runtime.Version()).
		SetString("osName", normalizeOSName(runtime.GOOS)).
		SetString("osArch", runtime.GOARCH).
		Build()
		// osVersion is not available, see above
	return ldvalue.ObjectBuild().
		SetString("kind", "diagnostic-init").
		Set("id", m.id).
		SetFloat64("creationDate", float64(m.startTime)).
		Set("sdk", m.sdkData).
		Set("configuration", m.configData).
		Set("platform", platformData).
		Build()
}

// CanSendStatsEvent is strictly for test instrumentation. In unit tests, we need to be able to stop
// DefaultEventProcessor from constructing the periodic event until the test has finished setting up its
// preconditions. This is done by passing in a periodicEventGate channel which the test will push to when it's
// ready.
func (m *DiagnosticsManager) CanSendStatsEvent() bool {
	if m.periodicEventGate != nil {
		select {
		case <-m.periodicEventGate: // non-blocking receive
			return true
		default:
			return false // COVERAGE: this path never executes in unit tests
		}
	}
	return true
}

// CreateStatsEventAndReset is called by DefaultEventProcessor to create the periodic event containing
// usage statistics. Some of the statistics are passed in as parameters because DefaultEventProcessor owns
// them and can more easily keep track of them internally - pushing them into DiagnosticsManager would
// require frequent lock usage.
func (m *DiagnosticsManager) CreateStatsEventAndReset(
	droppedEvents int,
	deduplicatedUsers int,
	eventsInLastBatch int,
) ldvalue.Value {
	m.lock.Lock()
	defer m.lock.Unlock()
	timestamp := ldtime.UnixMillisNow()
	streamInitsBuilder := ldvalue.ArrayBuildWithCapacity(len(m.streamInits))
	for _, si := range m.streamInits {
		streamInitsBuilder.Add(ldvalue.ObjectBuild().
			SetFloat64("timestamp", float64(si.timestamp)).
			SetBool("failed", si.failed).
			SetFloat64("durationMillis", float64(si.durationMillis)).
			Build())
	}
	event := ldvalue.ObjectBuild().
		SetString("kind", "diagnostic").
		Set("id", m.id).
		SetFloat64("creationDate", float64(timestamp)).
		SetFloat64("dataSinceDate", float64(m.dataSinceTime)).
		SetInt("droppedEvents", droppedEvents).
		SetInt("deduplicatedUsers", deduplicatedUsers).
		SetInt("eventsInLastBatch", eventsInLastBatch).
		Set("streamInits", streamInitsBuilder.Build()).
		Build()
	m.streamInits = nil
	m.dataSinceTime = timestamp
	return event
}

func normalizeOSName(osName string) string {
	switch osName {
	case "darwin": // COVERAGE: unit tests can only test one of these
		return "MacOS"
	case "windows": // COVERAGE: unit tests can only test one of these
		return "Windows"
	case "linux": // COVERAGE: unit tests can only test one of these
		return "Linux"
	}
	return osName // COVERAGE: unit tests can only test one of these
}
