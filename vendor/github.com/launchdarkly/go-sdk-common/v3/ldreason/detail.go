package ldreason

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
)

// EvaluationDetail is an object returned by the SDK's "detail" evaluation methods, such as
// [github.com/launchdarkly/go-server-sdk/v6.LDClient.BoolVariationDetail], combining the result of a
// flag evaluation with an explanation of how it was calculated.
type EvaluationDetail struct {
	// Value is the result of the flag evaluation. This will be either one of the flag's variations or
	// the default value that was passed to the evaluation method.
	Value ldvalue.Value
	// VariationIndex is the index of the returned value within the flag's list of variations, e.g.
	// 0 for the first variation. This is an ldvalue.OptionalInt rather than an int, because it is
	// possible for the value to be undefined (there is no variation index if the application default
	// value was returned due to an error in evaluation) which is different from a value of 0. See
	// ldvalue.OptionalInt for more about how to use this type.
	VariationIndex ldvalue.OptionalInt
	// Reason is an EvaluationReason object describing the main factor that influenced the flag
	// evaluation value.
	Reason EvaluationReason
}

// IsDefaultValue returns true if the result of the evaluation was the application default value.
// This means that an error prevented the flag from being evaluated; the Reason field should contain
// an error value such as NewEvalReasonError(EvalErrorFlagNotFound).
func (d EvaluationDetail) IsDefaultValue() bool {
	return !d.VariationIndex.IsDefined()
}

// NewEvaluationDetail constructs an EvaluationDetail, specifying all fields. This assumes that there
// is a defined value for variationIndex; if variationIndex is undefined, use [NewEvaluationDetailForError]
// or set the struct fields directly.
func NewEvaluationDetail(
	value ldvalue.Value,
	variationIndex int,
	reason EvaluationReason,
) EvaluationDetail {
	return EvaluationDetail{Value: value, VariationIndex: ldvalue.NewOptionalInt(variationIndex), Reason: reason}
}

// NewEvaluationDetailForError constructs an EvaluationDetail for an error condition.
func NewEvaluationDetailForError(errorKind EvalErrorKind, defaultValue ldvalue.Value) EvaluationDetail {
	return EvaluationDetail{Value: defaultValue, Reason: NewEvalReasonError(errorKind)}
}
