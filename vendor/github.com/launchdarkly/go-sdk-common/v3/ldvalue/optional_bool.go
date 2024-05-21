package ldvalue

// OptionalBool represents a bool that may or may not have a value. This is similar to using a
// bool pointer to distinguish between a false value and nil, but it is safer because it does not
// expose a pointer to any mutable value.
//
// To create an instance with a bool value, use [NewOptionalBool]. There is no corresponding method
// for creating an instance with no value; simply use the empty literal OptionalBool{}.
//
//	ob1 := NewOptionalBool(1)
//	ob2 := NewOptionalBool(false) // this has a value which is false
//	ob3 := OptionalBool{}         // this does not have a value
//
// This can also be used as a convenient way to construct a bool pointer within an expression.
// For instance, this example causes myIntPointer to point to the bool value true:
//
//	var myBoolPointer *int = NewOptionalBool(true).AsPointer()
//
// The reason LaunchDarkly code uses this specific type instead of a generic Optional[T] is for
// efficiency in JSON marshaling/unmarshaling. A generic type would have to use reflection and
// dynamic typecasting for its marshal/unmarshal methods.
type OptionalBool struct {
	optValue optional[bool]
}

// NewOptionalBool constructs an OptionalBool that has a bool value.
//
// There is no corresponding method for creating an OptionalBool with no value; simply use the
// empty literal OptionalBool{}.
func NewOptionalBool(value bool) OptionalBool {
	return OptionalBool{optValue: newOptional(value)}
}

// NewOptionalBoolFromPointer constructs an OptionalBool from a bool pointer. If the pointer is
// non-nil, then the OptionalBool copies its value; otherwise the OptionalBool has no value.
func NewOptionalBoolFromPointer(valuePointer *bool) OptionalBool {
	return OptionalBool{optValue: newOptionalFromPointer(valuePointer)}
}

// IsDefined returns true if the OptionalBool contains a bool value, or false if it has no value.
func (o OptionalBool) IsDefined() bool {
	return o.optValue.isDefined()
}

// BoolValue returns the OptionalBool's value, or false if it has no value.
func (o OptionalBool) BoolValue() bool {
	return o.optValue.getOrZeroValue()
}

// Get is a combination of BoolValue and IsDefined. If the OptionalBool contains a bool value, it
// returns that value and true; otherwise it returns false and false.
func (o OptionalBool) Get() (bool, bool) {
	return o.optValue.get()
}

// OrElse returns the OptionalBool's value if it has one, or else the specified fallback value.
func (o OptionalBool) OrElse(valueIfEmpty bool) bool {
	return o.optValue.getOrElse(valueIfEmpty)
}

// AsPointer returns the OptionalBool's value as a bool pointer if it has a value, or nil
// otherwise.
//
// The bool value, if any, is copied rather than returning to a pointer to the internal field.
func (o OptionalBool) AsPointer() *bool {
	return o.optValue.getAsPointer()
}

// AsValue converts the OptionalBool to a [Value], which is either [Null]() or a boolean value.
func (o OptionalBool) AsValue() Value {
	if value, ok := o.optValue.get(); ok {
		return Bool(value)
	}
	return Null()
}
