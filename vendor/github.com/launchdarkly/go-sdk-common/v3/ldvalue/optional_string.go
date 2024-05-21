package ldvalue

// OptionalString represents a string that may or may not have a value. This is similar to using a
// string pointer to distinguish between an empty string and nil, but it is safer because it does
// not expose a pointer to any mutable value.
//
// Unlike [Value], which can contain a value of any JSON-compatible type, OptionalString either
// contains a string or nothing. To create an instance with a string value, use [NewOptionalString].
// There is no corresponding method for creating an instance with no value; simply use the empty
// literal OptionalString{}.
//
//	os1 := NewOptionalString("this has a value")
//	os2 := NewOptionalString("") // this has a value which is an empty string
//	os2 := OptionalString{} // this does not have a value
//
// This can also be used as a convenient way to construct a string pointer within an expression.
// For instance, this example causes myStringPointer to point to the string "x":
//
//	var myStringPointer *string = NewOptionalString("x").AsPointer()
//
// The reason LaunchDarkly code uses this specific type instead of a generic Optional[T] is for
// efficiency in JSON marshaling/unmarshaling. A generic type would have to use reflection and
// dynamic typecasting for its marshal/unmarshal methods.
type OptionalString struct {
	optValue optional[string]
}

// NewOptionalString constructs an OptionalString that has a string value.
//
// There is no corresponding method for creating an OptionalString with no value; simply use
// the empty literal OptionalString{}.
func NewOptionalString(value string) OptionalString {
	return OptionalString{optValue: newOptional(value)}
}

// NewOptionalStringFromPointer constructs an OptionalString from a string pointer. If the pointer
// is non-nil, then the OptionalString copies its value; otherwise the OptionalString has no value.
func NewOptionalStringFromPointer(valuePointer *string) OptionalString {
	return OptionalString{optValue: newOptionalFromPointer(valuePointer)}
}

// IsDefined returns true if the OptionalString contains a string value, or false if it has no value.
func (o OptionalString) IsDefined() bool {
	return o.optValue.isDefined()
}

// StringValue returns the OptionalString's value, or an empty string if it has no value.
func (o OptionalString) StringValue() string {
	return o.optValue.getOrZeroValue()
}

// Get is a combination of StringValue and IsDefined. If the OptionalString contains a string value,
// it returns that value and true; otherwise it returns an empty string and false.
func (o OptionalString) Get() (string, bool) {
	return o.optValue.get()
}

// OrElse returns the OptionalString's value if it has one, or else the specified fallback value.
func (o OptionalString) OrElse(valueIfEmpty string) string {
	return o.optValue.getOrElse(valueIfEmpty)
}

// OnlyIfNonEmptyString returns the same OptionalString unless it contains an empty string (""), in
// which case it returns an OptionalString that has no value.
func (o OptionalString) OnlyIfNonEmptyString() OptionalString {
	if value, ok := o.optValue.get(); ok && value == "" {
		return OptionalString{}
	}
	return o
}

// AsPointer returns the OptionalString's value as a string pointer if it has a value, or
// nil otherwise.
//
// The string value, if any, is copied rather than returning to a pointer to the internal field.
func (o OptionalString) AsPointer() *string {
	return o.optValue.getAsPointer()
}

// AsValue converts the OptionalString to a [Value], which is either [Null]() or a string value.
func (o OptionalString) AsValue() Value {
	if value, ok := o.optValue.get(); ok {
		return String(value)
	}
	return Null()
}
