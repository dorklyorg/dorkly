package ldtime

import "time"

// UnixMillisecondTime is a millisecond timestamp starting from the Unix epoch.
type UnixMillisecondTime uint64

// UnixMillisFromTime converts a Time value into UnixMillisecondTime.
func UnixMillisFromTime(t time.Time) UnixMillisecondTime {
	ms := time.Duration(t.UnixNano()) / time.Millisecond
	return UnixMillisecondTime(ms)
}

// UnixMillisNow returns the current date/time as a UnixMillisecondTime.
func UnixMillisNow() UnixMillisecondTime {
	return UnixMillisFromTime(time.Now())
}

// IsDefined returns true if the time value is non-zero.
//
// This can be used to treat a zero value as "undefined" as an alternative to using a pointer,
// assuming that the exact beginning of the Unix epoch itself is not a valid time in this context.
func (t UnixMillisecondTime) IsDefined() bool {
	return t > 0
}
