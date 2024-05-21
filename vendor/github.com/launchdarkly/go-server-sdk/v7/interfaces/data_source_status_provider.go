package interfaces

import (
	"fmt"
	"time"
)

// DataSourceStatusProvider is an interface for querying the status of a DataSource. The data source is the
// component that receives updates to feature flag data; normally this is a streaming connection, but it
// could be polling or file data depending on your configuration.
//
// An implementation of this interface is returned by
// [github.com/launchdarkly/go-server-sdk/v7.LDClient.GetDataSourceStatusProvider()].
// Application code should not implement this interface.
//
// There are three ways to interact with the data source status. One is to simply get the current status;
// if its State property is DataSourceStateValid, then the SDK is able to receive feature flag updates.
//
//	status := client.GetDataSourceStatusProvider().GetStatus()
//	isValid = status.State == interfaces.DataSourceStateValid
//
// Second, you can use AddStatusListener to get a channel that provides a status update whenever the
// connection has an error or starts working again.
//
//	statusCh := client.GetDataSourceStatusProvider().AddStatusListener()
//	go func() {
//	    for newStatus := range statusCh {
//	        log.Printf("data source status is now: %+v", newStatus)
//	    }
//	}()
//
// Third, you can use WaitFor to block until the data source has the desired status. For instance, if you
// did not want to wait for a connection when you originally created the client, you could set the
// timeout to zero so that the connection happens in the background. Then, when you need to do something
// that requires a valid connection (possibly on another goroutine), you can wait until it is valid.
//
//	client, _ := ld.MakeCustomClient(sdkKey, config, 0)
//
//	// later...
//	inited := client.GetDataSourceStatusProvider().WaitFor(interfaces.DataSourceStateValid, 10 * time.Second)
//	if !inited {
//	    // do whatever is appropriate if initialization has timed out
//	}
type DataSourceStatusProvider interface {
	// GetStatus returns the current status of the data source.
	//
	// All of the built-in data source implementations are guaranteed to update this status whenever they
	// successfully initialize, encounter an error, or recover after an error.
	GetStatus() DataSourceStatus

	// AddStatusListener subscribes for notifications of status changes. The returned channel will receive a
	// new DataSourceStatus value for any change in status.
	//
	// The listener will be notified whenever any property of the status has changed. See DataSourceStatus for
	// an explanation of the meaning of each property and what could cause it to change.
	//
	// It is the caller's responsibility to consume values from the channel. Allowing values to accumulate in
	// the channel can cause an SDK goroutine to be blocked. If you no longer need the channel, call
	// RemoveStatusListener.
	AddStatusListener() <-chan DataSourceStatus

	// RemoveStatusListener unsubscribes from notifications of status changes. The specified channel must be
	// one that was previously returned by AddStatusListener(); otherwise, the method has no effect.
	RemoveStatusListener(listener <-chan DataSourceStatus)

	// WaitFor is a synchronous method for waiting for a desired connection state.
	//
	// If the current state is already desiredState when this method is called, it immediately returns.
	// Otherwise, it blocks until 1. the state has become desiredState, 2. the state has become
	// DataSourceStateOff (since that is a permanent condition), or 3. the specified timeout elapses.
	//
	// A scenario in which this might be useful is if you want to create the LDClient without waiting
	// for it to initialize, and then wait for initialization at a later time or on a different goroutine:
	//
	//     // create the client but do not wait
	//     client = ld.MakeCustomClient(sdkKey, config, 0)
	//
	//     // later, possibly on another goroutine:
	//     inited := client.GetDataSourceStatusProvider().WaitFor(DataSourceStateValid, 10 * time.Second)
	//     if !inited {
	//         // do whatever is appropriate if initialization has timed out
	//     }
	WaitFor(desiredState DataSourceState, timeout time.Duration) bool
}

// DataSourceStatus is information about the data source's status and the last status change.
//
// See [DataSourceStatusProvider].
type DataSourceStatus struct {
	// State represents the overall current state of the data source. It will always be one of the
	// DataSourceState constants such as DataSourceStateValid.
	State DataSourceState

	// StateSince is the date/time that the value of State most recently changed.
	//
	// The meaning of this depends on the current State:
	//   - For DataSourceStateInitializing, it is the time that the SDK started initializing.
	//   - For DataSourceStateValid, it is the time that the data source most recently entered a valid
	//     state, after previously having been either Initializing or Interrupted.
	//   - For DataSourceStateInterrupted, it is the time that the data source most recently entered an
	//     error state, after previously having been Valid.
	//   - For DataSourceStateOff, it is the time that the data source encountered an unrecoverable error
	//     or that the SDK was explicitly shut down.
	StateSince time.Time

	// LastError is information about the last error that the data source encountered, if any.
	//
	// This property should be updated whenever the data source encounters a problem, even if it does
	// not cause State to change. For instance, if a stream connection fails and the
	// state changes to DataSourceStateInterrupted, and then subsequent attempts to restart the
	// connection also fail, the state will remain Interrupted but the error information
	// will be updated each time-- and the last error will still be reported in this property even if
	// the state later becomes Valid.
	//
	// If no error has ever occurred, this field will be an empty DataSourceErrorInfo{}.
	LastError DataSourceErrorInfo
}

// String returns a simple string representation of the status.
func (e DataSourceStatus) String() string {
	return fmt.Sprintf("Status(%s,%s,%s)", e.State, e.StateSince.Format(time.RFC3339), e.LastError)
}

// DataSourceState is any of the allowable values for [DataSourceStatus].State.
//
// See [DataSourceStatusProvider].
type DataSourceState string

const (
	// DataSourceStateInitializing is the initial state of the data source when the SDK is being
	// initialized.
	//
	// If it encounters an error that requires it to retry initialization, the state will remain at
	// Initializing until it either succeeds and becomes DataSourceStateValid, or permanently fails and
	// becomes DataSourceStateOff.
	DataSourceStateInitializing DataSourceState = "INITIALIZING"

	// DataSourceStateValid indicates that the data source is currently operational and has not had
	// any problems since the last time it received data.
	//
	// In streaming mode, this means that there is currently an open stream connection and that at least
	// one initial message has been received on the stream. In polling mode, it means that the last poll
	// request succeeded.
	DataSourceStateValid DataSourceState = "VALID"

	// DataSourceStateInterrupted indicates that the data source encountered an error that it will
	// attempt to recover from.
	//
	// In streaming mode, this means that the stream connection failed, or had to be dropped due to some
	// other error, and will be retried after a backoff delay. In polling mode, it means that the last poll
	// request failed, and a new poll request will be made after the configured polling interval.
	DataSourceStateInterrupted DataSourceState = "INTERRUPTED"

	// DataSourceStateOff indicates that the data source has been permanently shut down.
	//
	// This could be because it encountered an unrecoverable error (for instance, the LaunchDarkly service
	// rejected the SDK key; an invalid SDK key will never become valid), or because the SDK client was
	// explicitly shut down.
	DataSourceStateOff DataSourceState = "OFF"
)

// DataSourceErrorInfo is a description of an error condition that the data source encountered.
//
// See [DataSourceStatusProvider].
type DataSourceErrorInfo struct {
	// Kind is the general category of the error. It will always be one of the DataSourceErrorKind
	// constants such as DataSourceErrorKindNetworkError, or "" if there have not been any errors.
	Kind DataSourceErrorKind

	// StatusCode is the HTTP status code if the error was DataSourceErrorKindErrorResponse, or zero
	// otherwise.
	StatusCode int

	// Message is any any additional human-readable information relevant to the error. The format of
	// this message is subject to change and should not be relied on programmatically.
	Message string

	// Time is the date/time that the error occurred.
	Time time.Time
}

// String returns a simple string representation of the error.
func (e DataSourceErrorInfo) String() string {
	ret := string(e.Kind)
	if e.StatusCode > 0 || e.Message != "" {
		ret += "("
		if e.StatusCode > 0 {
			ret += fmt.Sprintf("%d", e.StatusCode)
		}
		if e.Message != "" {
			if e.StatusCode > 0 {
				ret += ","
			}
			ret += e.Message
		}
		ret += ")"
	}
	if !e.Time.IsZero() {
		ret += fmt.Sprintf("@%s", e.Time.Format(time.RFC3339))
	}
	return ret
}

// DataSourceErrorKind is any of the allowable values for [DataSourceErrorInfo].Kind.
//
// See [DataSourceStatusProvider].
type DataSourceErrorKind string

const (
	// DataSourceErrorKindUnknown indicates an unexpected error, such as an uncaught exception,
	// further described by DataSourceErrorInfo.Message.
	DataSourceErrorKindUnknown DataSourceErrorKind = "UNKNOWN"

	// DataSourceErrorKindNetworkError represents an I/O error such as a dropped connection.
	DataSourceErrorKindNetworkError DataSourceErrorKind = "NETWORK_ERROR"

	// DataSourceErrorKindErrorResponse means the LaunchDarkly service returned an HTTP response
	// with an error status, available in DataSourceErrorInfo.StatusCode.
	DataSourceErrorKindErrorResponse DataSourceErrorKind = "ERROR_RESPONSE"

	// DataSourceErrorKindInvalidData means the SDK received malformed data from the LaunchDarkly
	// service.
	DataSourceErrorKindInvalidData DataSourceErrorKind = "INVALID_DATA"

	// DataSourceErrorKindStoreError means the data source itself is working, but when it tried
	// to put an update into the data store, the data store failed (so the SDK may not have the
	// latest data).
	//
	// Data source implementations do not need to report this kind of error; it will be
	// automatically reported by the SDK whenever one of the update methods of DataSourceUpdateSink
	// encounters a failure.
	DataSourceErrorKindStoreError DataSourceErrorKind = "STORE_ERROR"
)
