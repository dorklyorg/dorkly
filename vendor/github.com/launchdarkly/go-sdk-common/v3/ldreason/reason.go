package ldreason

import (
	"fmt"

	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"github.com/launchdarkly/go-jsonstream/v3/jreader"
	"github.com/launchdarkly/go-jsonstream/v3/jwriter"
)

// EvalReasonKind defines the possible values of [EvaluationReason.GetKind].
type EvalReasonKind string

const (
	// EvalReasonOff indicates that the flag was off and therefore returned its configured off value.
	EvalReasonOff EvalReasonKind = "OFF"
	// EvalReasonTargetMatch indicates that the context key was specifically targeted for this flag.
	EvalReasonTargetMatch EvalReasonKind = "TARGET_MATCH"
	// EvalReasonRuleMatch indicates that the context matched one of the flag's rules.
	EvalReasonRuleMatch EvalReasonKind = "RULE_MATCH"
	// EvalReasonPrerequisiteFailed indicates that the flag was considered off because it had at
	// least one prerequisite flag that either was off or did not return the desired variation.
	EvalReasonPrerequisiteFailed EvalReasonKind = "PREREQUISITE_FAILED"
	// EvalReasonFallthrough indicates that the flag was on but the context did not match any targets
	// or rules.
	EvalReasonFallthrough EvalReasonKind = "FALLTHROUGH"
	// EvalReasonError indicates that the flag could not be evaluated, e.g. because it does not
	// exist or due to an unexpected error. In this case the result value will be the default value
	// that the caller passed to the client.
	EvalReasonError EvalReasonKind = "ERROR"
)

// EvalErrorKind defines the possible values of EvaluationReason.GetErrorKind().
type EvalErrorKind string

const (
	// EvalErrorClientNotReady indicates that the caller tried to evaluate a flag before the client
	// had successfully initialized.
	EvalErrorClientNotReady EvalErrorKind = "CLIENT_NOT_READY"
	// EvalErrorFlagNotFound indicates that the caller provided a flag key that did not match any
	// known flag.
	EvalErrorFlagNotFound EvalErrorKind = "FLAG_NOT_FOUND"
	// EvalErrorMalformedFlag indicates that there was an internal inconsistency in the flag data,
	// e.g. a rule specified a nonexistent variation.
	EvalErrorMalformedFlag EvalErrorKind = "MALFORMED_FLAG"
	// EvalErrorUserNotSpecified indicates that the caller passed an invalid or uninitialized
	// context. The name and value of this constant refer to "user" rather than "context" only for
	// backward compatibility with older SDKs that used the term "user".
	EvalErrorUserNotSpecified EvalErrorKind = "USER_NOT_SPECIFIED"
	// EvalErrorWrongType indicates that the result value was not of the requested type, e.g. you
	// called BoolVariationDetail but the value was an integer.
	EvalErrorWrongType EvalErrorKind = "WRONG_TYPE"
	// EvalErrorException indicates that an unexpected error stopped flag evaluation; check the
	// log for details.
	EvalErrorException EvalErrorKind = "EXCEPTION"
)

// BigSegmentsStatus defines the possible values of [EvaluationReason.GetBigSegmentsStatus].
//
// "Big segments" are a specific type of segments. For more information, read the LaunchDarkly
// documentation: https://docs.launchdarkly.com/home/contexts/big-segments
type BigSegmentsStatus string

const (
	// BigSegmentsHealthy indicates that the Big Segment query involved in the flag
	// evaluation was successful, and that the segment state is considered up to date.
	BigSegmentsHealthy BigSegmentsStatus = "HEALTHY"

	// BigSegmentsStale indicates that the Big Segment query involved in the flag
	// evaluation was successful, but that the segment state may not be up to date.
	BigSegmentsStale BigSegmentsStatus = "STALE"

	// BigSegmentsNotConfigured indicates that Big Segments could not be queried for the
	// flag evaluation because the SDK configuration did not include a big segment store.
	BigSegmentsNotConfigured BigSegmentsStatus = "NOT_CONFIGURED"

	// BigSegmentsStoreError indicates that the Big Segment query involved in the flag
	// evaluation failed, for instance due to a database error.
	BigSegmentsStoreError BigSegmentsStatus = "STORE_ERROR"
)

// EvaluationReason describes the reason that a flag evaluation producted a particular value.
//
// This struct is immutable; its properties can be accessed only via getter methods.
type EvaluationReason struct {
	kind              EvalReasonKind
	ruleIndex         ldvalue.OptionalInt
	ruleID            string
	prerequisiteKey   string
	inExperiment      bool
	errorKind         EvalErrorKind
	bigSegmentsStatus BigSegmentsStatus
}

// IsDefined returns true if this EvaluationReason has a non-empty [EvaluationReason.GetKind]. It is
// false for a zero value of EvaluationReason{}.
func (r EvaluationReason) IsDefined() bool {
	return r.kind != ""
}

// String returns a concise string representation of the reason. Examples: "OFF", "ERROR(WRONG_TYPE)".
//
// This value is intended only for convenience in logging or debugging. Application code should not
// rely on its specific format.
func (r EvaluationReason) String() string {
	switch r.kind {
	case EvalReasonRuleMatch:
		return fmt.Sprintf("%s(%d,%s)", r.kind, r.ruleIndex.OrElse(0), r.ruleID)
	case EvalReasonPrerequisiteFailed:
		return fmt.Sprintf("%s(%s)", r.kind, r.prerequisiteKey)
	case EvalReasonError:
		return fmt.Sprintf("%s(%s)", r.kind, r.errorKind)
	default:
		return string(r.GetKind())
	}
}

// GetKind describes the general category of the reason.
func (r EvaluationReason) GetKind() EvalReasonKind {
	return r.kind
}

// GetRuleIndex provides the index of the rule that was matched (0 being the first), if
// the Kind is [EvalReasonRuleMatch]. Otherwise it returns -1.
func (r EvaluationReason) GetRuleIndex() int {
	return r.ruleIndex.OrElse(-1)
}

// GetRuleID provides the unique identifier of the rule that was matched, if the Kind is
// [EvalReasonRuleMatch]. Otherwise it returns an empty string. Unlike the rule index, this
// identifier will not change if other rules are added or deleted.
func (r EvaluationReason) GetRuleID() string {
	return r.ruleID
}

// GetPrerequisiteKey provides the flag key of the prerequisite that failed, if the Kind
// is [EvalReasonPrerequisiteFailed]. Otherwise it returns an empty string.
func (r EvaluationReason) GetPrerequisiteKey() string {
	return r.prerequisiteKey
}

// IsInExperiment describes whether the evaluation was part of an experiment. It returns
// true if the evaluation resulted in an experiment rollout *and* served one of the
// variations in the experiment.  Otherwise it returns false.
func (r EvaluationReason) IsInExperiment() bool {
	return r.inExperiment
}

// GetErrorKind describes the general category of the error, if the Kind is [EvalReasonError].
// Otherwise it returns an empty string.
func (r EvaluationReason) GetErrorKind() EvalErrorKind {
	return r.errorKind
}

// GetBigSegmentsStatus describes the validity of Big Segment information, if and only if the flag
// evaluation required querying at least one Big Segment. Otherwise it returns an empty string.
//
// "Big segments" are a specific kind of segments. For more information, read the LaunchDarkly
// documentation: https://docs.launchdarkly.com/home/contexts/big-segments
func (r EvaluationReason) GetBigSegmentsStatus() BigSegmentsStatus {
	return r.bigSegmentsStatus
}

// NewEvalReasonOff returns an EvaluationReason whose Kind is [EvalReasonOff].
func NewEvalReasonOff() EvaluationReason {
	return EvaluationReason{kind: EvalReasonOff}
}

// NewEvalReasonFallthrough returns an EvaluationReason whose Kind is [EvalReasonFallthrough].
func NewEvalReasonFallthrough() EvaluationReason {
	return EvaluationReason{kind: EvalReasonFallthrough}
}

// NewEvalReasonFallthroughExperiment returns an EvaluationReason whose Kind is
// [EvalReasonFallthrough]. The inExperiment parameter represents whether the evaluation was
// part of an experiment.
func NewEvalReasonFallthroughExperiment(inExperiment bool) EvaluationReason {
	return EvaluationReason{kind: EvalReasonFallthrough, inExperiment: inExperiment}
}

// NewEvalReasonTargetMatch returns an EvaluationReason whose Kind is [EvalReasonTargetMatch].
func NewEvalReasonTargetMatch() EvaluationReason {
	return EvaluationReason{kind: EvalReasonTargetMatch}
}

// NewEvalReasonRuleMatch returns an EvaluationReason whose Kind is [EvalReasonRuleMatch].
func NewEvalReasonRuleMatch(ruleIndex int, ruleID string) EvaluationReason {
	return EvaluationReason{kind: EvalReasonRuleMatch,
		ruleIndex: ldvalue.NewOptionalInt(ruleIndex), ruleID: ruleID}
}

// NewEvalReasonRuleMatchExperiment returns an EvaluationReason whose Kind is
// [EvalReasonRuleMatch]. The inExperiment parameter represents whether the evaluation was
// part of an experiment.
func NewEvalReasonRuleMatchExperiment(ruleIndex int, ruleID string, inExperiment bool) EvaluationReason {
	return EvaluationReason{
		kind:         EvalReasonRuleMatch,
		ruleIndex:    ldvalue.NewOptionalInt(ruleIndex),
		ruleID:       ruleID,
		inExperiment: inExperiment,
	}
}

// NewEvalReasonPrerequisiteFailed returns an EvaluationReason whose Kind is [EvalReasonPrerequisiteFailed].
func NewEvalReasonPrerequisiteFailed(prereqKey string) EvaluationReason {
	return EvaluationReason{kind: EvalReasonPrerequisiteFailed, prerequisiteKey: prereqKey}
}

// NewEvalReasonError returns an EvaluationReason whose Kind is [EvalReasonError].
func NewEvalReasonError(errorKind EvalErrorKind) EvaluationReason {
	return EvaluationReason{kind: EvalReasonError, errorKind: errorKind}
}

// NewEvalReasonFromReasonWithBigSegmentsStatus returns a copy of an EvaluationReason
// with a specific [BigSegmentsStatus] value added.
func NewEvalReasonFromReasonWithBigSegmentsStatus(
	reason EvaluationReason,
	bigSegmentsStatus BigSegmentsStatus,
) EvaluationReason {
	reason.bigSegmentsStatus = bigSegmentsStatus
	return reason
}

// MarshalJSON implements custom JSON serialization for EvaluationReason.
func (r EvaluationReason) MarshalJSON() ([]byte, error) {
	return jwriter.MarshalJSONWithWriter(r)
}

// UnmarshalJSON implements custom JSON deserialization for EvaluationReason.
func (r *EvaluationReason) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, r)
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [encoding/json.Unmarshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (r *EvaluationReason) ReadFromJSONReader(reader *jreader.Reader) {
	var ret EvaluationReason
	for obj := reader.ObjectOrNull(); obj.Next(); {
		switch string(obj.Name()) {
		case "kind":
			ret.kind = EvalReasonKind(reader.String())
		case "ruleId":
			ret.ruleID = reader.String()
		case "ruleIndex":
			ret.ruleIndex = ldvalue.NewOptionalInt(reader.Int())
		case "errorKind":
			ret.errorKind = EvalErrorKind(reader.String())
		case "prerequisiteKey":
			ret.prerequisiteKey = reader.String()
		case "inExperiment":
			ret.inExperiment = reader.Bool()
		case "bigSegmentsStatus":
			ret.bigSegmentsStatus = BigSegmentsStatus(reader.String())
		}
	}
	if reader.Error() == nil {
		*r = ret
	}
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [encoding/json.Marshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (r EvaluationReason) WriteToJSONWriter(w *jwriter.Writer) {
	if r.kind == "" {
		w.Null()
		return
	}
	obj := w.Object()
	obj.Name("kind").String(string(r.kind))
	if r.ruleIndex.IsDefined() {
		obj.Name("ruleIndex").Int(r.ruleIndex.OrElse(0))
		obj.Maybe("ruleId", r.ruleID != "").String(r.ruleID)
	}
	obj.Maybe("inExperiment", r.inExperiment).Bool(r.inExperiment)
	if r.kind == EvalReasonPrerequisiteFailed {
		obj.Name("prerequisiteKey").String(r.prerequisiteKey)
	}
	if r.kind == EvalReasonError {
		obj.Name("errorKind").String(string(r.errorKind))
	}
	if r.bigSegmentsStatus != "" {
		obj.Name("bigSegmentsStatus").String(string(r.bigSegmentsStatus))
	}
	obj.End()
}
