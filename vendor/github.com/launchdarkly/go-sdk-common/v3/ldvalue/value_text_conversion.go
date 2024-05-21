package ldvalue

import (
	"encoding/json"
)

// MarshalText implements the encoding.TextMarshaler interface.
//
// This may be useful with packages that support describing arbitrary types via that interface.
//
// The behavior for Value is the same as Value.MarshalJSON. For instance, ldvalue.Bool(true)
// becomes the string "true", and ldvalue.String("x") becomes the string "\"x\"".
func (v Value) MarshalText() ([]byte, error) {
	return v.MarshalJSON()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// This may be useful with packages that support parsing arbitrary types via that interface,
// such as gcfg.
//
// The behavior for Value is that if the input is a valid JSON representation-- such as the
// string "null", "false", "1", or "[1, 2]"-- it is parsed identically to json.Unmarshal().
// Otherwise, it is assumed to be a string, and the result is the same as calling
// ldvalue.String() with the same value. This behavior is intended to be similar to the
// default behavior of YAML v2 parsers.
func (v *Value) UnmarshalText(text []byte) error {
	if err := json.Unmarshal(text, v); err == nil {
		return nil
	}
	*v = String(string(text))
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
//
// This may be useful with packages that support describing arbitrary types via that interface.
//
// The behavior for ValueArray is the same as ValueArray.MarshalJSON.
func (a ValueArray) MarshalText() ([]byte, error) {
	return a.MarshalJSON()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// This may be useful with packages that support parsing arbitrary types via that interface,
// such as gcfg.
//
// The behavior for ValueArray is the same as ValueArray.UnmarshalJSON.
func (a *ValueArray) UnmarshalText(text []byte) error {
	return a.UnmarshalJSON(text)
}

// MarshalText implements the encoding.TextMarshaler interface.
//
// This may be useful with packages that support describing arbitrary types via that interface.
//
// The behavior for ValueMap is the same as ValueMap.MarshalJSON.
func (m ValueMap) MarshalText() ([]byte, error) {
	return m.MarshalJSON()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// This may be useful with packages that support parsing arbitrary types via that interface,
// such as gcfg.
//
// The behavior for ValueMap is the same as ValueMap.UnmarshalJSON.
func (m *ValueMap) UnmarshalText(text []byte) error {
	return m.UnmarshalJSON(text)
}
