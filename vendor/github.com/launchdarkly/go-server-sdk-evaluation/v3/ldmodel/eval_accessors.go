package ldmodel

import (
	"regexp"
	"time"

	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	"github.com/launchdarkly/go-semver"
)

// EvaluatorAccessorMethods contains functions that are used by the evaluation engine in the
// parent package to perform certain lookup operations on data model structs.
//
// These are defined in the ldmodel package because they take advantage of the preprocessing
// behavior that is defined for these types, which populates additional data structures to
// speed up such lookups. Those data structures are implementation details of this package,
// so they are not exported. Instead, these methods provide a more abstract way for the
// evaluation engine to perform simple lookups regardless of whether the preprocessed data
// is available. (Normally preprocessed data is always available, because the preprocessing
// step is done every time we unmarshal data from JSON; but the evaluator must be able to
// work even if it receives inputs that were constructed in some other way.)
//
// For efficiency, all of these methods expect structs to be passed by address rather than
// by value. They are guaranteed not to modify any fields.
//
// Defining these as methods of EvaluatorAccessorMethods (accessed via the global variable
// EvaluatorAccessors), rather than simple functions or methods of other types, keeps this
// functionality clearly grouped together and allows data model types like FlagRule to be
// simple structs without methods.
type EvaluatorAccessorMethods struct{}

// EvaluatorAccessors is the global entry point for EvaluatorAccessorMethods.
var EvaluatorAccessors EvaluatorAccessorMethods //nolint:gochecknoglobals

// ClauseFindValue returns true if the specified value is deeply equal to any of the Clause's
// Values, or false otherwise. It also returns false if the value is a JSON array, a JSON
// object, or a JSON null (since equality tests are not valid for these in the LaunchDarkly
// model), or if the clause parameter is nil.
//
// If preprocessing has been done, this is a fast map lookup (as long as the Clause's operator
// is "in", which is the only case where it makes sense to create a map). Otherwise it iterates
// the list.
func (e EvaluatorAccessorMethods) ClauseFindValue(clause *Clause, contextValue ldvalue.Value) bool {
	if clause == nil {
		return false
	}
	if clause.preprocessed.valuesMap != nil {
		if key := asPrimitiveValueKey(contextValue); key.isValid() {
			_, found := clause.preprocessed.valuesMap[key]
			return found
		}
	}
	switch contextValue.Type() {
	case ldvalue.BoolType, ldvalue.NumberType, ldvalue.StringType:
		for _, clauseValue := range clause.Values {
			if contextValue.Equal(clauseValue) {
				return true
			}
		}
	default:
		break
	}
	return false
}

// ClauseGetValueAsRegexp returns one of the Clause's values as a Regexp, if the value is a string
// that represents a valid regular expression.
//
// It returns nil if the value is not a string or is not valid as a regular expression; if the
// index is out of range; or if the clause parameter is nil.
//
// If preprocessing has been done, this is a fast slice lookup. Otherwise it calls regexp.Compile.
func (e EvaluatorAccessorMethods) ClauseGetValueAsRegexp(clause *Clause, index int) *regexp.Regexp {
	if clause == nil {
		return nil
	}
	if clause.preprocessed.values != nil {
		if index < 0 || index >= len(clause.preprocessed.values) {
			return nil
		}
		return clause.preprocessed.values[index].parsedRegexp
	}
	if index >= 0 && index < len(clause.Values) {
		return parseRegexp(clause.Values[index])
	}
	return nil
}

// ClauseGetValueAsSemanticVersion returns one of the Clause's values as a semver.Version, if the
// value is a string in the correct format. Any other type is invalid.
//
// The second return value is true for success or false for failure. It also returns failure if the
// index is out of range, or if the clause parameter is nil.
//
// If preprocessing has been done, this is a fast slice lookup. Otherwise it calls
// TypeConversions.ValueToSemanticVersion.
func (e EvaluatorAccessorMethods) ClauseGetValueAsSemanticVersion(clause *Clause, index int) (semver.Version, bool) {
	if clause == nil {
		return semver.Version{}, false
	}
	if clause.preprocessed.values != nil {
		if index < 0 || index >= len(clause.preprocessed.values) {
			return semver.Version{}, false
		}
		p := clause.preprocessed.values[index]
		return p.parsedSemver, p.valid
	}
	if index >= 0 && index < len(clause.Values) {
		return TypeConversions.ValueToSemanticVersion(clause.Values[index])
	}
	return semver.Version{}, false
}

// ClauseGetValueAsTimestamp returns one of the Clause's values as a time.Time, if the value is a
// string or number in the correct format. Any other type is invalid.
//
// The second return value is true for success or false for failure.. It also returns failure if the
// index is out of range, or if the clause parameter is nil.
//
// If preprocessing has been done, this is a fast slice lookup. Otherwise it calls
// TypeConversions.ValueToTimestamp.
func (e EvaluatorAccessorMethods) ClauseGetValueAsTimestamp(clause *Clause, index int) (time.Time, bool) {
	if clause == nil {
		return time.Time{}, false
	}
	if clause.preprocessed.values != nil {
		if index < 0 || index >= len(clause.preprocessed.values) {
			return time.Time{}, false
		}
		t := clause.preprocessed.values[index].parsedTime
		return t, !t.IsZero()
	}
	if index >= 0 && index < len(clause.Values) {
		return TypeConversions.ValueToTimestamp(clause.Values[index])
	}
	return time.Time{}, false
}

// SegmentFindKeyInExcluded returns true if the specified key is in this Segment's
// Excluded list, or false otherwise. It also returns false if the segment parameter is nil.
//
// If preprocessing has been done, this is a fast map lookup. Otherwise it iterates the list.
func (e EvaluatorAccessorMethods) SegmentFindKeyInExcluded(segment *Segment, key string) bool {
	if segment == nil {
		return false
	}
	return findValueInMapOrStrings(key, segment.Excluded, segment.preprocessed.excludeMap)
}

// SegmentFindKeyInIncluded returns true if the specified key is in this Segment's
// Included list, or false otherwise. It also returns false if the segment parameter is nil.
//
// If preprocessing has been done, this is a fast map lookup. Otherwise it iterates the list.
func (e EvaluatorAccessorMethods) SegmentFindKeyInIncluded(segment *Segment, key string) bool {
	if segment == nil {
		return false
	}
	return findValueInMapOrStrings(key, segment.Included, segment.preprocessed.includeMap)
}

// SegmentTargetFindKey returns true if the specified key is in this SegmentTarget's
// Values list, or false otherwise. It also returns false if the target parameter is nil.
//
// If preprocessing has been done, this is a fast map lookup. Otherwise it iterates the list.
func (e EvaluatorAccessorMethods) SegmentTargetFindKey(target *SegmentTarget, key string) bool {
	if target == nil {
		return false
	}
	return findValueInMapOrStrings(key, target.Values, target.preprocessed.valuesMap)
}

// TargetFindKey returns true if the specified key is in this Target's Values list, or false
// otherwise. It also returns false if the target parameter is nil.
//
// If preprocessing has been done, this is a fast map lookup. Otherwise it iterates the list.
func (e EvaluatorAccessorMethods) TargetFindKey(target *Target, key string) bool {
	if target == nil {
		return false
	}
	return findValueInMapOrStrings(key, target.Values, target.preprocessed.valuesMap)
}

func findValueInMapOrStrings(value string, values []string, valuesMap map[string]struct{}) bool {
	if valuesMap != nil {
		_, found := valuesMap[value]
		return found
	}
	for _, v := range values {
		if value == v {
			return true
		}
	}
	return false
}
