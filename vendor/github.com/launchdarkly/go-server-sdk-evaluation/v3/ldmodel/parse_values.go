package ldmodel

import (
	"regexp"
	"time"

	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	"github.com/launchdarkly/go-semver"
)

func parseDateTime(value ldvalue.Value) (time.Time, bool) {
	switch value.Type() {
	case ldvalue.StringType:
		return parseRFC3339TimeUTC(value.StringValue())
	case ldvalue.NumberType:
		return unixMillisToUtcTime(value.Float64Value()), true
	}
	return time.Time{}, false
}

func unixMillisToUtcTime(unixMillis float64) time.Time {
	return time.Unix(0, int64(unixMillis)*int64(time.Millisecond)).UTC()
}

func parseRegexp(value ldvalue.Value) *regexp.Regexp {
	if value.IsString() {
		if r, err := regexp.Compile(value.StringValue()); err == nil {
			return r
		}
	}
	return nil
}

func parseSemVer(value ldvalue.Value) (semver.Version, bool) {
	if value.IsString() {
		versionStr := value.StringValue()
		if sv, err := semver.ParseAs(versionStr, semver.ParseModeAllowMissingMinorAndPatch); err == nil {
			return sv, true
		}
	}
	return semver.Version{}, false
}
