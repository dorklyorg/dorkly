package ldmodel

// List of available operators
const (
	// OperatorIn matches a user value and clause value if the two values are equal (including their type).
	OperatorIn Operator = "in"
	// OperatorEndsWith matches a user value and clause value if they are both strings and the former ends with
	// the latter.
	OperatorEndsWith Operator = "endsWith"
	// OperatorStartsWith matches a user value and clause value if they are both strings and the former starts
	// with the latter.
	OperatorStartsWith Operator = "startsWith"
	// OperatorMatches matches a user value and clause value if they are both strings and the latter is a valid
	// regular expression that matches the former.
	OperatorMatches Operator = "matches"
	// OperatorContains matches a user value and clause value if they are both strings and the former contains
	// the latter.
	OperatorContains Operator = "contains"
	// OperatorLessThan matches a user value and clause value if they are both numbers and the former < the
	// latter.
	OperatorLessThan Operator = "lessThan"
	// OperatorLessThanOrEqual matches a user value and clause value if they are both numbers and the former
	// <= the latter.
	OperatorLessThanOrEqual Operator = "lessThanOrEqual"
	// OperatorGreaterThan matches a user value and clause value if they are both numbers and the former > the
	// latter.
	OperatorGreaterThan Operator = "greaterThan"
	// OperatorGreaterThanOrEqual matches a user value and clause value if they are both numbers and the former
	// >= the latter.
	OperatorGreaterThanOrEqual Operator = "greaterThanOrEqual"
	// OperatorBefore matches a user value and clause value if they are both timestamps and the former < the
	// latter.
	//
	// A valid timestamp is either a string in RFC3339/ISO8601 format, or a number which is treated as Unix
	// milliseconds.
	OperatorBefore Operator = "before"
	// OperatorAfter matches a user value and clause value if they are both timestamps and the former > the
	// latter.
	//
	// A valid timestamp is either a string in RFC3339/ISO8601 format, or a number which is treated as Unix
	// milliseconds.
	OperatorAfter Operator = "after"
	// OperatorSegmentMatch matches a user if the user is included in the user segment whose key is the clause
	// value.
	OperatorSegmentMatch Operator = "segmentMatch"
	// OperatorSemVerEqual matches a user value and clause value if they are both semantic versions and they
	// are equal.
	//
	// A semantic version is a string that either follows the Semantic Versions 2.0 spec, or is an abbreviated
	// version consisting of digits and optional periods in the form "m" (equivalent to m.0.0) or "m.n"
	// (equivalent to m.n.0).
	OperatorSemVerEqual Operator = "semVerEqual"
	// OperatorSemVerLessThan matches a user value and clause value if they are both semantic versions and the
	// former < the latter.
	//
	// A semantic version is a string that either follows the Semantic Versions 2.0 spec, or is an abbreviated
	// version consisting of digits and optional periods in the form "m" (equivalent to m.0.0) or "m.n"
	// (equivalent to m.n.0).
	OperatorSemVerLessThan Operator = "semVerLessThan"
	// OperatorSemVerGreaterThan matches a user value and clause value if they are both semantic versions and
	// the former > the latter.
	//
	// A semantic version is a string that either follows the Semantic Versions 2.0 spec, or is an abbreviated
	// version consisting of digits and optional periods in the form "m" (equivalent to m.0.0) or "m.n"
	// (equivalent to m.n.0).
	OperatorSemVerGreaterThan Operator = "semVerGreaterThan"
)

// Operator describes an operator for a clause.
type Operator string
