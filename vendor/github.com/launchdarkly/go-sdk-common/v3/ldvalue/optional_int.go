package ldvalue

// OptionalInt represents an int that may or may not have a value. This is similar to using an
// int pointer to distinguish between a zero value and nil, but it is safer because it does not
// expose a pointer to any mutable value.
//
// To create an instance with an int value, use [NewOptionalInt]. There is no corresponding method
// for creating an instance with no value; simply use the empty literal OptionalInt{}.
//
//	oi1 := NewOptionalInt(1)
//	oi2 := NewOptionalInt(0) // this has a value which is zero
//	oi3 := OptionalInt{}     // this does not have a value
//
// This can also be used as a convenient way to construct an int pointer within an expression.
// For instance, this example causes myIntPointer to point to the int value 2:
//
//	var myIntPointer *int = NewOptionalInt("x").AsPointer()
//
// This type is used in ldreason.EvaluationDetail.VariationIndex, and for other similar fields
// in the LaunchDarkly Go SDK where an int value may or may not be defined.
//
// The reason LaunchDarkly code uses this specific type instead of a generic Optional[T] is for
// efficiency in JSON marshaling/unmarshaling. A generic type would have to use reflection and
// dynamic typecasting for its marshal/unmarshal methods.
type OptionalInt struct {
	optValue optional[int]
}

// NewOptionalInt constructs an OptionalInt that has an int value.
//
// There is no corresponding method for creating an OptionalInt with no value; simply use the
// empty literal OptionalInt{}.
func NewOptionalInt(value int) OptionalInt {
	return OptionalInt{optValue: newOptional(value)}
}

// NewOptionalIntFromPointer constructs an OptionalInt from an int pointer. If the pointer is
// non-nil, then the OptionalInt copies its value; otherwise the OptionalInt has no value.
func NewOptionalIntFromPointer(valuePointer *int) OptionalInt {
	return OptionalInt{optValue: newOptionalFromPointer(valuePointer)}
}

// IsDefined returns true if the OptionalInt contains an int value, or false if it has no value.
func (o OptionalInt) IsDefined() bool {
	return o.optValue.isDefined()
}

// IntValue returns the OptionalInt's value, or zero if it has no value.
func (o OptionalInt) IntValue() int {
	return o.optValue.getOrZeroValue()
}

// Get is a combination of IntValue and IsDefined. If the OptionalInt contains an int value, it
// returns that value and true; otherwise it returns zero and false.
func (o OptionalInt) Get() (int, bool) {
	return o.optValue.get()
}

// OrElse returns the OptionalInt's value if it has one, or else the specified fallback value.
func (o OptionalInt) OrElse(valueIfEmpty int) int {
	return o.optValue.getOrElse(valueIfEmpty)
}

// AsPointer returns the OptionalInt's value as an int pointer if it has a value, or nil
// otherwise.
//
// The int value, if any, is copied rather than returning to a pointer to the internal field.
func (o OptionalInt) AsPointer() *int {
	return o.optValue.getAsPointer()
}

// AsValue converts the OptionalInt to a [Value], which is either [Null]() or a number value.
func (o OptionalInt) AsValue() Value {
	if value, ok := o.optValue.get(); ok {
		return Int(value)
	}
	return Null()
}
