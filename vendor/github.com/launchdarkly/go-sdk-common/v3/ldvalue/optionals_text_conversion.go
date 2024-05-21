package ldvalue

import "strconv"

// Note: the implementations of encoding.TextMarshaler and encoding.TextUnmarshaler are provided
// for convenience for use with general-purpose parsing/configuration tools. LaunchDarkly SDK code
// does not use these methods.

// String returns a human-readable string representation of the value.
//
// Currently, this is defined as being either "true", "false, or "[none]" if it has no value.
// However, the specific representation format is subject to change and should not be relied on in
// code; it is intended for convenience in logging or debugging.
func (o OptionalBool) String() string {
	if value, ok := o.optValue.get(); ok {
		if value {
			return trueString
		}
		return falseString
	}
	return noneDescription
}

// String returns a human-readable string representation of the value.
//
// Currently, this is defined as being either a numeric string or "[none]" if it has no value.
// However, the specific representation format is subject to change and should not be relied on in
// code; it is intended for convenience in logging or debugging.
func (o OptionalInt) String() string {
	if value, ok := o.optValue.get(); ok {
		return strconv.Itoa(value)
	}
	return noneDescription
}

// String returns a human-readable string representation of the value.
//
// Currently, this is defined as being either the string value itself or "[none]" if it has no value.
// However, the specific representation format is subject to change and should not be relied on in
// code; it is intended for convenience in logging or debugging.
//
// Since String() is used by methods like fmt.Printf, if you want to use the actual string value
// of the OptionalString instead of this special representation, you should call StringValue():
//
//	s := OptionalString{}
//	fmt.Printf("it is '%s'", s) // prints "it is '[none]'"
//	fmt.Printf("it is '%s'", s.StringValue()) // prints "it is ''"
func (o OptionalString) String() string {
	if value, ok := o.optValue.get(); ok {
		if value == "" {
			return "[empty]"
		}
		return value
	}
	return noneDescription
}

// MarshalText implements the encoding.TextMarshaler interface.
//
// This may be useful with packages that support describing arbitrary types via that interface.
//
// The behavior for OptionalBool is that a true or false value produces the string "true" or
// "false", and an undefined value produces an empty string.
func (o OptionalBool) MarshalText() ([]byte, error) {
	if value, ok := o.optValue.get(); ok {
		if value {
			return trueBytes, nil
		}
		return falseBytes, nil
	}
	return []byte(""), nil
}

// MarshalText implements the encoding.TextMarshaler interface.
//
// This may be useful with packages that support describing arbitrary types via that interface.
//
// The behavior for OptionalBool is that a true or false value produces a decimal numeric
// string, and an undefined value produces an empty string.
func (o OptionalInt) MarshalText() ([]byte, error) {
	if value, ok := o.optValue.get(); ok {
		return []byte(strconv.Itoa(value)), nil
	}
	return []byte(""), nil
}

// MarshalText implements the encoding.TextMarshaler interface.
//
// This may be useful with packages that support describing arbitrary types via that interface.
//
// The behavior for OptionalString is that a defined string value produces the same string,
// and an undefined value produces nil.
func (o OptionalString) MarshalText() ([]byte, error) {
	if value, ok := o.optValue.get(); ok {
		return []byte(value), nil
	}
	return nil, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// This may be useful with packages that support parsing arbitrary types via that interface,
// such as gcfg.
//
// If the input data is nil or empty, the result is an empty OptionalBool{}. Otherwise, it
// recognizes "true" or "false" as valid inputs and returns an error for all other values.
//
// This allows OptionalBool to be used with packages that can parse text content, such as gcfg.
func (o *OptionalBool) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*o = OptionalBool{}
		return nil
	}
	return o.UnmarshalJSON(text)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// This may be useful with packages that support parsing arbitrary types via that interface,
// such as gcfg.
//
// If the input data is nil or empty, the result is an empty OptionalInt{}. Otherwise, it
// uses strconv.Atoi() to parse a numeric value.
//
// This allows OptionalInt to be used with packages that can parse text content, such as gcfg.
func (o *OptionalInt) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*o = OptionalInt{}
		return nil
	}
	n, err := strconv.Atoi(string(text))
	if err != nil {
		return err
	}
	*o = NewOptionalInt(n)
	return nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// This may be useful with packages that support parsing arbitrary types via that interface,
// such as gcfg.
//
// If the byte slice is nil, rather than zero-length, it will set the target value to an empty
// OptionalString{}. Otherwise, it will set it to a string value.
func (o *OptionalString) UnmarshalText(text []byte) error {
	if text == nil {
		*o = OptionalString{}
	} else {
		*o = NewOptionalString(string(text))
	}
	return nil
}
