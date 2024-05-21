package ldvalue

import (
	"encoding/json"

	"golang.org/x/exp/slices"
)

// Value represents any of the data types supported by JSON, all of which can be used for a LaunchDarkly
// feature flag variation, or for an attribute in an evaluation context. Value instances are immutable.
//
// # Uses of JSON types in LaunchDarkly
//
// LaunchDarkly feature flags can have variations of any JSON type other than null. If you want to
// evaluate a feature flag in a general way that does not have expectations about the variation type,
// or if the variation value is a complex data structure such as an array or object, you can use the
// SDK method [github.com/launchdarkly/go-server-sdk/v6/LDClient.JSONVariationDetail] to get the value
// and then use Value methods to examine it.
//
// Similarly, attributes of an evaluation context ([github.com/launchdarkly/go-sdk-common/v3/ldcontext.Context])
// can have variations of any JSON type other than null. If you want to set a context attribute in a general
// way that will accept any type, or set the attribute value to a complex data structure such as an array
// or object, you can use the builder method [github.com/launchdarkly/go-sdk-common/v3/ldcontext.Builder.SetValue].
//
// Arrays and objects have special meanings in LaunchDarkly flag evaluation:
//   - An array of values means "try to match any of these values to the targeting rule."
//   - An object allows you to match a property within the object to the targeting rule. For instance,
//     in the example above, a targeting rule could reference /objectAttr1/color to match the value
//     "green". Nested property references like /objectAttr1/address/street are allowed if a property
//     contains another JSON object.
//
// # Constructors and builders
//
// The most efficient way to create a Value is with typed constructors such as [Bool] and [String], or, for
// complex data, the builders [ArrayBuild] and [ObjectBuild]. However, any value that can be represented in
// JSON can be converted to a Value with [CopyArbitraryValue] or [FromJSONMarshal], or you can parse a
// Value from JSON. The following code examples all produce the same value:
//
//	value := ldvalue.ObjectBuild().SetString("thing", "x").Build()
//
//	type MyStruct struct {
//		Thing string `json:"thing"`
//	}
//	value := ldvalue.FromJSONMarshal(MyStruct{Thing: "x"})
//
//	value := ldvalue.Parse([]byte(`{"thing": "x"}`))
//
// # Comparisons
//
// You cannot compare Value instances with the == operator, because the struct may contain a slice. or a
// map. Value has the [Value.Equal] method for this purpose; [reflect.DeepEqual] will also work.
type Value struct {
	valueType ValueType
	// Used when the value is a boolean.
	boolValue bool
	// Used when the value is a number.
	numberValue float64
	// Used when the value is a string.
	stringValue string
	// Used when the value is an array, zero-valued otherwise.
	arrayValue ValueArray
	// Used when the value is an object, zero-valued otherwise.
	objectValue ValueMap
	rawValue    []byte
}

// ValueType indicates which JSON type is contained in a [Value].
type ValueType int

// String returns the name of the value type.
func (t ValueType) String() string {
	switch t {
	case NullType:
		return nullAsJSON
	case BoolType:
		return "bool"
	case NumberType:
		return "number"
	case StringType:
		return "string"
	case ArrayType:
		return "array"
	case ObjectType:
		return "object"
	case RawType:
		return "raw"
	default:
		return "unknown"
	}
}

// Null creates a null Value.
func Null() Value {
	return Value{valueType: NullType}
}

// Bool creates a boolean Value.
func Bool(value bool) Value {
	return Value{valueType: BoolType, boolValue: value}
}

// Int creates a numeric Value from an integer.
//
// Note that all numbers are represented internally as the same type (float64), so Int(2) is
// exactly equal to Float64(2).
func Int(value int) Value {
	return Float64(float64(value))
}

// Float64 creates a numeric Value from a float64.
func Float64(value float64) Value {
	return Value{valueType: NumberType, numberValue: value}
}

// String creates a string Value.
func String(value string) Value {
	return Value{valueType: StringType, stringValue: value}
}

// Raw creates an unparsed JSON Value.
//
// This constructor stores a copy of the [json.RawMessage] value as-is without syntax validation, and
// sets the type of the Value to [RawType]. The advantage of this is that if you have some
// data that you do not expect to do any computations with, but simply want to include in JSON data
// to be sent to LaunchDarkly, the original data will be output as-is and does not need to be parsed.
//
// However, if you do anything that involves inspecting the value (such as comparing it to another
// value, or evaluating a feature flag that references the value in a context attribute), the JSON
// data will be parsed automatically each time that happens, so there will be no efficiency gain.
// Therefore, if you expect any such operations to happen, it is better to use [Parse] instead to
// parse the JSON immediately, or use value builder methods such as [ObjectBuild].
//
// If you pass malformed data that is not valid JSON, you will get malformed data if it is re-encoded
// to JSON. It is the caller's responsibility to make sure the json.RawMessage really is valid JSON.
// However, since it is easy to mistakenly write json.RawMessage(nil) (an invalid zero-length value)
// when what is really meant is a JSON null value, this constructor transparently converts both
// json.RawMessage(nil) and json.RawMessage([]byte("")) to Null().
func Raw(value json.RawMessage) Value {
	if len(value) == 0 {
		return Null()
	}
	return Value{valueType: RawType, rawValue: slices.Clone(value)}
}

// FromJSONMarshal creates a Value from the JSON representation of any Go value.
//
// This is based on [encoding/json.Marshal], so it can be used with any value that can be passed to that
// function, and follows the usual behavior of json.Marshal with regard to types and field names.
// For instance, you could use it with a custom struct type:
//
//	type MyStruct struct {
//		Thing string `json:"thing"`
//	}
//	value := ldvalue.FromJSONMarshal(MyStruct{Thing: "x"})
//
// It is equivalent to calling json.Marshal followed by [Parse], so it incurs the same overhead of
// first marshaling the value and then parsing the resulting JSON.
func FromJSONMarshal(anyValue any) Value {
	jsonBytes, err := json.Marshal(anyValue)
	if err != nil {
		return Null()
	}
	return Parse(jsonBytes)
}

// CopyArbitraryValue creates a Value from an arbitrary value of any type.
//
// If the value is nil, a boolean, an integer, a floating-point number, or a string, it becomes the
// corresponding JSON primitive value type. If it is a slice of values ([]any or []Value), it is
// deep-copied to an array value. If it is a map of strings to values (map[string]any or
// map[string]Value), it is deep-copied to an object value.
//
// If it is a pointer to any of the above types, then it is dereferenced and treated the same as above,
// unless the pointer is nil, in which case it becomes [Null]().
//
// For all other types, the value is marshaled to JSON and then converted to the corresponding Value
// type, just as if you had called [FromJSONMarshal]. The difference is only that CopyArbitraryValue is
// more efficient for primitive types, slices, and maps, which it can copy without first marshaling
// them to JSON.
func CopyArbitraryValue(anyValue any) Value { //nolint:gocyclo // yes, we know it's a long function
	if anyValue == nil {
		return Null()
		// Note that an interface value can be nil in two ways: nil with no type at all, which is this case,
		// or a nil pointer of some specific pointer type, which we have to check for separately below.
	}
	switch o := anyValue.(type) {
	case Value:
		return o
	case *Value:
		if o == nil {
			return Null()
		}
		return *o
	case OptionalString:
		return o.AsValue()
	case *OptionalString:
		if o == nil {
			return Null()
		}
		return o.AsValue()
	case bool:
		return Bool(o)
	case *bool:
		if o == nil {
			return Null()
		}
		return Bool(*o)
	case int8:
		return Float64(float64(o))
	case *int8:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case uint8:
		return Float64(float64(o))
	case *uint8:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case int16:
		return Float64(float64(o))
	case *int16:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case uint16:
		return Float64(float64(o))
	case *uint16:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case int:
		return Float64(float64(o))
	case *int:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case uint:
		return Float64(float64(o))
	case *uint:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case int32:
		return Float64(float64(o))
	case *int32:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case uint32:
		return Float64(float64(o))
	case *uint32:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case float32:
		return Float64(float64(o))
	case *float32:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case float64:
		return Float64(o)
	case *float64:
		if o == nil {
			return Null()
		}
		return Float64(*o)
	case string:
		return String(o)
	case *string:
		if o == nil {
			return Null()
		}
		return String(*o)
	case []any:
		return copyArbitraryValueArray(o)
	case *[]any:
		if o == nil {
			return Null()
		}
		return copyArbitraryValueArray(*o)
	case []Value:
		return ArrayOf(o...)
	case *[]Value:
		if o == nil {
			return Null()
		}
		return ArrayOf((*o)...)
	case map[string]any:
		return copyArbitraryValueMap(o)
	case *map[string]any:
		if o == nil {
			return Null()
		}
		return copyArbitraryValueMap(*o)
	case map[string]Value:
		return CopyObject(o)
	case *map[string]Value:
		if o == nil {
			return Null()
		}
		return CopyObject(*o)
	case json.RawMessage:
		return Raw(o)
	case *json.RawMessage:
		if o == nil {
			return Null()
		}
		return Raw(*o)
	default:
		return FromJSONMarshal(anyValue)
	}
}

func copyArbitraryValueArray(o []any) Value {
	return Value{valueType: ArrayType, arrayValue: CopyArbitraryValueArray(o)}
}

func copyArbitraryValueMap(o map[string]any) Value {
	return Value{valueType: ObjectType, objectValue: CopyArbitraryValueMap(o)}
}

// Type returns the ValueType of the Value.
func (v Value) Type() ValueType {
	return v.valueType
}

// IsNull returns true if the Value is a null.
func (v Value) IsNull() bool {
	return v.valueType == NullType || (v.valueType == RawType && v.parseIfRaw().IsNull())
}

// IsDefined returns true if the Value is anything other than null.
//
// This is exactly equivalent to !v.IsNull(), but is provided as a separate method for consistency
// with other types that have an IsDefined() method.
func (v Value) IsDefined() bool {
	return !v.IsNull()
}

// IsBool returns true if the Value is a boolean.
func (v Value) IsBool() bool {
	return v.valueType == BoolType || (v.valueType == RawType && v.parseIfRaw().IsBool())
}

// IsNumber returns true if the Value is numeric.
func (v Value) IsNumber() bool {
	return v.valueType == NumberType || (v.valueType == RawType && v.parseIfRaw().IsNumber())
}

// IsInt returns true if the Value is an integer.
//
// JSON does not have separate types for integer and floating-point values; they are both just numbers.
// IsInt returns true if and only if the actual numeric value has no fractional component, so
// Int(2).IsInt() and Float64(2.0).IsInt() are both true.
func (v Value) IsInt() bool {
	return (v.valueType == NumberType && v.numberValue == float64(int(v.numberValue))) ||
		(v.valueType == RawType && v.parseIfRaw().IsInt())
}

// IsString returns true if the Value is a string.
func (v Value) IsString() bool {
	return v.valueType == StringType || (v.valueType == RawType && v.parseIfRaw().IsString())
}

// BoolValue returns the Value as a boolean.
//
// If the Value is not a boolean, it returns false.
func (v Value) BoolValue() bool {
	switch v.valueType {
	case BoolType:
		return v.boolValue
	case RawType:
		return v.parseIfRaw().BoolValue()
	default:
		return false
	}
}

// IntValue returns the value as an int.
//
// If the Value is not numeric, it returns zero. If the value is a number but not an integer, it is
// rounded toward zero (truncated).
func (v Value) IntValue() int {
	return int(v.Float64Value())
}

// Float64Value returns the value as a float64.
//
// If the Value is not numeric, it returns zero.
func (v Value) Float64Value() float64 {
	switch v.valueType {
	case NumberType:
		return v.numberValue
	case RawType:
		return v.parseIfRaw().Float64Value()
	default:
		return 0
	}
}

// StringValue returns the value as a string.
//
// If the value is not a string, it returns an empty string.
//
// This is different from [String], which returns a string representation of any value type,
// including any necessary JSON delimiters.
func (v Value) StringValue() string {
	switch v.valueType {
	case StringType:
		return v.stringValue
	case RawType:
		return v.parseIfRaw().StringValue()
	default:
		return ""
	}
}

// AsOptionalString converts the value to the OptionalString type, which contains either a string
// value or nothing if the original value was not a string.
func (v Value) AsOptionalString() OptionalString {
	switch v.valueType {
	case StringType:
		return NewOptionalString(v.stringValue)
	case RawType:
		return v.parseIfRaw().AsOptionalString()
	default:
		return OptionalString{}
	}
}

// AsRaw returns the value as a [json.RawMessage].
//
// If the value was originally created with [Raw], it returns a copy of the original value. For all
// other values, it converts the value to its JSON representation and returns that representation.
//
// Note that the [Raw] constructor does not do any syntax validation, so if you create a Value from
// a malformed string such as ldvalue.Raw(json.RawMessage("{{{")), you will get back the same string
// from AsRaw().
func (v Value) AsRaw() json.RawMessage {
	if v.valueType == RawType {
		return v.rawValue
	}
	bytes, err := json.Marshal(v)
	if err == nil {
		return json.RawMessage(bytes)
	}
	return nil
}

// AsArbitraryValue returns the value in its simplest Go representation, typed as "any".
//
// This is nil for a null value; for primitive types, it is bool, float64, or string (all numbers
// are represented as float64 because that is Go's default when parsing from JSON). For unparsed
// JSON data created with [Raw], it returns a [json.RawMessage].
//
// Arrays and objects are represented as []any and map[string]any. They are deep-copied, which
// preserves immutability of the Value but may be an expensive operation. To examine array and
// object values without copying the whole data structure, use getter methods: [Value.Count],
// [Value.Keys], [Value.GetByIndex], [Value.TryGetByIndex], [Value.GetByKey], [Value.TryGetByKey].
func (v Value) AsArbitraryValue() any {
	switch v.valueType {
	case NullType:
		return nil
	case BoolType:
		return v.boolValue
	case NumberType:
		return v.numberValue
	case StringType:
		return v.stringValue
	case ArrayType:
		return v.arrayValue.AsArbitraryValueSlice()
	case ObjectType:
		return v.objectValue.AsArbitraryValueMap()
	case RawType:
		return v.AsRaw()
	default:
		return nil
	}
}

// String converts the value to a string representation, equivalent to [Value.JSONString].
//
// This is different from [Value.StringValue], which returns the actual string for a string value or an empty
// string for anything else. For instance, Int(2).StringValue() returns "2" and String("x").StringValue()
// returns "\"x\"", whereas Int(2).AsString() returns "" and String("x").AsString() returns
// "x".
//
// This method is provided because it is common to use the [fmt.Stringer] interface as a quick way to
// summarize the contents of a value. The simplest way to do so in this case is to use the JSON
// representation.
func (v Value) String() string {
	return v.JSONString()
}

// Equal tests whether this Value is equal to another, in both type and value.
//
// For arrays and objects, this is a deep equality test. This method behaves the same as
// [reflect.DeepEqual], but is slightly more efficient.
//
// Unparsed JSON values created with [Raw] will be parsed in order to do this comparison.
func (v Value) Equal(other Value) bool {
	if v.valueType == RawType || other.valueType == RawType {
		return v.parseIfRaw().Equal(other.parseIfRaw())
	}
	if v.valueType == other.valueType {
		switch v.valueType {
		case NullType:
			return true
		case BoolType:
			return v.boolValue == other.boolValue
		case NumberType:
			return v.numberValue == other.numberValue
		case StringType, RawType:
			return v.stringValue == other.stringValue
		case ArrayType:
			return v.arrayValue.Equal(other.arrayValue)
		case ObjectType:
			return v.objectValue.Equal(other.objectValue)
		}
	}
	return false
}

// AsPointer returns either a pointer to a copy of this Value, or nil if it is a null value.
//
// This may be desirable if you are serializing a struct that contains a Value, and you want
// that field to be completely omitted if the Value is null; since the "omitempty" tag only
// works for pointers, you can declare the field as a *Value like so:
//
//	type MyJsonStruct struct {
//	    AnOptionalField *Value `json:"anOptionalField,omitempty"`
//	}
//	s := MyJsonStruct{AnOptionalField: someValue.AsPointer()}
func (v Value) AsPointer() *Value {
	if v.IsNull() {
		return nil
	}
	return &v
}

func (v Value) parseIfRaw() Value {
	if v.valueType != RawType {
		return v
	}
	return Parse(v.rawValue)
}
