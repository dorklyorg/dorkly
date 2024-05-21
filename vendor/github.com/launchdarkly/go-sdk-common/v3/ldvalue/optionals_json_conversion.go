package ldvalue

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/launchdarkly/go-jsonstream/v3/jreader"
	"github.com/launchdarkly/go-jsonstream/v3/jwriter"
)

// JSONString returns the JSON representation of the value as a string. This is
// guaranteed to be logically equivalent to calling [json.Marshal] and converting the
// first return value to a string.
//
// Since types that support this method are by definition always convertible to JSON,
// it does not need to return an error value so it can be easily used within an
// expression.
func (o OptionalBool) JSONString() string {
	return o.AsValue().JSONString()
}

// JSONString returns the JSON representation of the value as a string. This is
// guaranteed to be logically equivalent to calling [json.Marshal] and converting the
// first return value to a string.
//
// Since types that support this method are by definition always convertible to JSON,
// it does not need to return an error value so it can be easily used within an
// expression.
func (o OptionalInt) JSONString() string {
	return o.AsValue().JSONString()
}

// JSONString returns the JSON representation of the value as a string. This is
// guaranteed to be logically equivalent to calling [json.Marshal] and converting the
// first return value to a string.
//
// Since types that support this method are by definition always convertible to JSON,
// it does not need to return an error value so it can be easily used within an
// expression.
func (o OptionalString) JSONString() string {
	return o.AsValue().JSONString()
}

// MarshalJSON converts the OptionalBool to its JSON representation.
//
// The output will be either a JSON boolean or null. Note that the "omitempty" tag for a struct
// field will not cause an empty OptionalBool field to be omitted; it will be output as null.
// If you want to completely omit a JSON property when there is no value, it must be a bool
// pointer instead of an OptionalBool; use [OptionalBool.AsPointer] to get a pointer.
func (o OptionalBool) MarshalJSON() ([]byte, error) {
	return o.AsValue().MarshalJSON()
}

// MarshalJSON converts the OptionalInt to its JSON representation.
//
// The output will be either a JSON number or null. Note that the "omitempty" tag for a struct
// field will not cause an empty OptionalInt field to be omitted; it will be output as null.
// If you want to completely omit a JSON property when there is no value, it must be an int
// pointer instead of an OptionalInt; use [OptionalInt.AsPointer] method to get a pointer.
func (o OptionalInt) MarshalJSON() ([]byte, error) {
	return o.AsValue().MarshalJSON()
}

// MarshalJSON converts the OptionalString to its JSON representation.
//
// The output will be either a JSON string or null. Note that the "omitempty" tag for a struct
// field will not cause an empty OptionalString field to be omitted; it will be output as null.
// If you want to completely omit a JSON property when there is no value, it must be a string
// pointer instead of an OptionalString; use [OptionalString.AsPointer] method to get a pointer.
func (o OptionalString) MarshalJSON() ([]byte, error) {
	return o.AsValue().MarshalJSON()
}

// UnmarshalJSON parses an OptionalBool from JSON.
//
// The input must be either a JSON boolean or null.
func (o *OptionalBool) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*o = OptionalBool{}
		return nil
	}
	if bytes.Equal(data, []byte("true")) {
		*o = NewOptionalBool(true)
		return nil
	}
	if bytes.Equal(data, []byte("false")) {
		*o = NewOptionalBool(false)
		return nil
	}
	return &json.UnmarshalTypeError{Value: string(data), Type: reflect.TypeOf(o)}
}

// UnmarshalJSON parses an OptionalInt from JSON.
//
// The input must be either a JSON number that is an integer or null.
func (o *OptionalInt) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, o)
}

// UnmarshalJSON parses an OptionalString from JSON.
//
// The input must be either a JSON string or null.
func (o *OptionalString) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, o)
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Unmarshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (o *OptionalBool) ReadFromJSONReader(r *jreader.Reader) {
	val, nonNull := r.BoolOrNull()
	if r.Error() == nil {
		if nonNull {
			*o = NewOptionalBool(val)
		} else {
			*o = OptionalBool{}
		}
	}
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Unmarshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (o *OptionalInt) ReadFromJSONReader(r *jreader.Reader) {
	val, nonNull := r.IntOrNull()
	if r.Error() == nil {
		if nonNull {
			*o = NewOptionalInt(val)
		} else {
			*o = OptionalInt{}
		}
	}
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Unmarshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (o *OptionalString) ReadFromJSONReader(r *jreader.Reader) {
	val, nonNull := r.StringOrNull()
	if r.Error() == nil {
		if nonNull {
			*o = NewOptionalString(val)
		} else {
			*o = OptionalString{}
		}
	}
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Marshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (o OptionalBool) WriteToJSONWriter(w *jwriter.Writer) {
	o.AsValue().WriteToJSONWriter(w)
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Marshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (o OptionalInt) WriteToJSONWriter(w *jwriter.Writer) {
	o.AsValue().WriteToJSONWriter(w)
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Marshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (o OptionalString) WriteToJSONWriter(w *jwriter.Writer) {
	o.AsValue().WriteToJSONWriter(w)
}
