package ldvalue

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/launchdarkly/go-jsonstream/v3/jreader"
	"github.com/launchdarkly/go-jsonstream/v3/jwriter"
)

// This file contains methods for converting Value to and from JSON.

// Parse returns a Value parsed from a JSON string, or Null if it cannot be parsed.
//
// This is simply a shortcut for calling json.Unmarshal and disregarding errors. It is meant for
// use in test scenarios where malformed data is not a concern.
func Parse(jsonData []byte) Value {
	var v Value
	if err := v.UnmarshalJSON(jsonData); err != nil {
		return Null()
	}
	return v
}

// JSONString returns the JSON representation of the value.
//
// This is equivalent to calling [Value.MarshalJSON] and converting the result to a string.
// Since all Values by definition can be represented in JSON, this method does not need to
// return an error value so it can be easily used within an expression.
func (v Value) JSONString() string {
	// The following is somewhat redundant with json.Marshal, but it avoids the overhead of
	// converting between byte arrays and strings.
	switch v.valueType {
	case NullType:
		return nullAsJSON
	case BoolType:
		if v.boolValue {
			return trueString
		}
		return falseString
	case NumberType:
		if v.IsInt() {
			return strconv.Itoa(int(v.numberValue))
		}
		return strconv.FormatFloat(v.numberValue, 'f', -1, 64)
	}
	// For all other types, we rely on our custom marshaller.
	bytes, _ := json.Marshal(v)
	// It shouldn't be possible for marshalling to fail, because Value can only contain
	// JSON-compatible types. But if it somehow did fail, bytes will be nil and we'll return
	// an empty string.
	return string(bytes)
}

// MarshalJSON converts the Value to its JSON representation.
//
// Note that the "omitempty" tag for a struct field will not cause an empty Value field to be
// omitted; it will be output as null. If you want to completely omit a JSON property when there
// is no value, it must be a pointer; use AsPointer().
func (v Value) MarshalJSON() ([]byte, error) {
	switch v.valueType {
	case NullType:
		return nullAsJSONBytes, nil
	case BoolType:
		if v.boolValue {
			return trueBytes, nil
		}
		return falseBytes, nil
	case NumberType:
		if v.IsInt() {
			return []byte(strconv.Itoa(int(v.numberValue))), nil
		}
		return []byte(strconv.FormatFloat(v.numberValue, 'f', -1, 64)), nil
	case StringType:
		return json.Marshal(v.stringValue)
	case ArrayType:
		return v.arrayValue.MarshalJSON()
	case ObjectType:
		return v.objectValue.MarshalJSON()
	case RawType:
		return v.rawValue, nil
	}
	return nil, errors.New("unknown data type") // should not be possible
}

// UnmarshalJSON parses a Value from JSON.
func (v *Value) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, v)
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Unmarshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (v *Value) ReadFromJSONReader(r *jreader.Reader) {
	a := r.Any()
	if r.Error() != nil {
		return
	}
	switch a.Kind {
	case jreader.BoolValue:
		*v = Bool(a.Bool)
	case jreader.NumberValue:
		*v = Float64(a.Number)
	case jreader.StringValue:
		*v = String(a.String)
	case jreader.ArrayValue:
		var va ValueArray
		if va.readFromJSONArray(r, &a.Array); r.Error() == nil {
			*v = Value{valueType: ArrayType, arrayValue: va}
		}
	case jreader.ObjectValue:
		var vm ValueMap
		if vm.readFromJSONObject(r, &a.Object); r.Error() == nil {
			*v = Value{valueType: ObjectType, objectValue: vm}
		}
	default:
		*v = Null()
	}
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Marshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (v Value) WriteToJSONWriter(w *jwriter.Writer) {
	switch v.valueType {
	case NullType:
		w.Null()
	case BoolType:
		w.Bool(v.boolValue)
	case NumberType:
		w.Float64(v.numberValue)
	case StringType:
		w.String(v.stringValue)
	case ArrayType:
		v.arrayValue.WriteToJSONWriter(w)
	case ObjectType:
		v.objectValue.WriteToJSONWriter(w)
	case RawType:
		w.Raw(v.rawValue)
	}
}

// JSONString returns the JSON representation of the array.
func (a ValueArray) JSONString() string {
	bytes, _ := a.MarshalJSON()
	// It shouldn't be possible for marshalling to fail, because Value can only contain
	// JSON-compatible types. But if it somehow did fail, bytes will be nil and we'll return
	// an empty tring.
	return string(bytes)
}

// MarshalJSON converts the ValueArray to its JSON representation.
//
// Like a Go slice, a ValueArray in an uninitialized/nil state produces a JSON null rather than an empty [].
func (a ValueArray) MarshalJSON() ([]byte, error) {
	if a.data == nil {
		return nullAsJSONBytes, nil
	}
	return json.Marshal(a.data)
}

// UnmarshalJSON parses a ValueArray from JSON.
func (a *ValueArray) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, a)
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Unmarshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (a *ValueArray) ReadFromJSONReader(r *jreader.Reader) {
	arr := r.ArrayOrNull()
	a.readFromJSONArray(r, &arr)
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Marshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
//
// Like a Go slice, a ValueArray in an uninitialized/nil state produces a JSON null rather than an empty [].
func (a ValueArray) WriteToJSONWriter(w *jwriter.Writer) {
	if a.data == nil {
		w.Null()
		return
	}
	arr := w.Array()
	for _, v := range a.data {
		v.WriteToJSONWriter(w)
	}
	arr.End()
}

func (a *ValueArray) readFromJSONArray(r *jreader.Reader, arr *jreader.ArrayState) {
	if r.Error() != nil {
		return
	}
	if !arr.IsDefined() {
		*a = ValueArray{}
		return
	}
	var ab ValueArrayBuilder
	for arr.Next() {
		var vv Value
		vv.ReadFromJSONReader(r)
		ab.Add(vv)
	}
	if r.Error() == nil {
		*a = ab.Build()
	}
}

// String converts the value to a map representation, equivalent to JSONString().
//
// This method is provided because it is common to use the Stringer interface as a quick way to
// summarize the contents of a value. The simplest way to do so in this case is to use the JSON
// representation.
func (m ValueMap) String() string {
	return m.JSONString()
}

// JSONString returns the JSON representation of the map.
func (m ValueMap) JSONString() string {
	bytes, _ := m.MarshalJSON()
	// It shouldn't be possible for marshalling to fail, because Value can only contain
	// JSON-compatible types. But if it somehow did fail, bytes will be nil and we'll return
	// an empty tring.
	return string(bytes)
}

// MarshalJSON converts the ValueMap to its JSON representation.
//
// Like a Go map, a ValueMap in an uninitialized/nil state produces a JSON null rather than an empty {}.
func (m ValueMap) MarshalJSON() ([]byte, error) {
	return jwriter.MarshalJSONWithWriter(m)
}

// UnmarshalJSON parses a ValueMap from JSON.
func (m *ValueMap) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, m)
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Unmarshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
func (m *ValueMap) ReadFromJSONReader(r *jreader.Reader) {
	obj := r.ObjectOrNull()
	m.readFromJSONObject(r, &obj)
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than [json.Marshal].
// See [github.com/launchdarkly/go-jsonstream/v3] for more details.
//
// Like a Go map, a ValueMap in an uninitialized/nil state produces a JSON null rather than an empty {}.
func (m ValueMap) WriteToJSONWriter(w *jwriter.Writer) {
	if m.data == nil {
		w.Null()
		return
	}
	obj := w.Object()
	for k, vv := range m.data {
		vv.WriteToJSONWriter(obj.Name(k))
	}
	obj.End()
}

func (m *ValueMap) readFromJSONObject(r *jreader.Reader, obj *jreader.ObjectState) {
	if r.Error() != nil {
		return
	}
	if !obj.IsDefined() {
		*m = ValueMap{}
		return
	}
	var mb ValueMapBuilder
	for obj.Next() {
		name := obj.Name()
		var vv Value
		vv.ReadFromJSONReader(r)
		mb.Set(string(name), vv)
	}
	if r.Error() == nil {
		*m = mb.Build()
	}
}
