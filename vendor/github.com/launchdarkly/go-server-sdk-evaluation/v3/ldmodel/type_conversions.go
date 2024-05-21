package ldmodel

import (
	"time"

	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	"github.com/launchdarkly/go-semver"
)

// TypeConversionMethods contains type conversion functions that are used by the evaluation
// engine in the parent package. They implement the defined behavior for JSON value types
// when used with LaunchDarkly feature flag operators that apply to a more specialized
// logical type, such as timestamps and semantic versions.
//
// These are defined in the ldmodel package, rather than in the parent package with the
// evaluation engine, for two reasons:
//
// 1. They are logically part of the LaunchDarkly data model. For instance, when a clause
// uses a date/time operator, that implies that the clause values must be strings or numbers
// in the specific format that LaunchDarkly uses for timestamps.
//
// 2. The preprocessing logic in ldmodel uses the same conversions to parse clause values as
// the appropriate types ahead of time. EvaluatorAccessorMethods will use the pre-parsed
// values if available, or else apply the conversions on the fly.
type TypeConversionMethods struct{}

// TypeConversions is the global entry point for TypeConversionMethods.
var TypeConversions TypeConversionMethods //nolint:gochecknoglobals

// ValueToTimestamp attempts to convert a JSON value to a time.Time, using the standard
// LaunchDarkly rules for timestamp values.
//
// If the value is a string, it is parsed according to RFC3339. If the value is a number,
// it is treated as integer epoch milliseconds. Any other type is invalid.
//
// The second return value is true for success or false for failure.
func (e TypeConversionMethods) ValueToTimestamp(value ldvalue.Value) (time.Time, bool) {
	switch value.Type() {
	case ldvalue.StringType:
		return parseRFC3339TimeUTC(value.StringValue())
	case ldvalue.NumberType:
		unixMillis := int64(value.Float64Value())
		return time.Unix(0, unixMillis*int64(time.Millisecond)).UTC(), true
	}
	return time.Time{}, false
}

// ValueToSemanticVersion attempts to convert a JSON value to a semver.Version.
//
// If the value is a string, it is parsed with the parser defined in the semver package. Any
// other type is invalid.
//
// The second return value is true for success or false for failure.
func (e TypeConversionMethods) ValueToSemanticVersion(value ldvalue.Value) (semver.Version, bool) {
	return parseSemVer(value)
}
