package jreader

// AnyValue is returned by Reader.Any() to represent a JSON value of an arbitrary type.
type AnyValue struct {
	// Kind describes the type of the JSON value.
	Kind ValueKind

	// Bool is the value if the JSON value is a boolean, or false otherwise.
	Bool bool

	// Number is the value if the JSON value is a number, or zero otherwise.
	Number float64

	// String is the value if the JSON value is a string, or an empty string otherwise.
	String string

	// Array is an ArrayState that can be used to iterate through the array elements if the JSON
	// value is an array, or an uninitialized ArrayState{} otherwise.
	Array ArrayState

	// Object is an ObjectState that can be used to iterate through the object properties if the
	// JSON value is an object, or an uninitialized ObjectState{} otherwise.
	Object ObjectState
}

// ValueKind defines the allowable value types for Reader.Any.
type ValueKind int

const (
	// NullValue means the value is a null.
	NullValue ValueKind = iota

	// BoolValue means the value is a boolean.
	BoolValue ValueKind = iota

	// NumberValue means the value is a number.
	NumberValue ValueKind = iota

	// StringValue means the value is a string.
	StringValue ValueKind = iota

	// ArrayValue means the value is an array.
	ArrayValue ValueKind = iota

	// ObjectValue means the value is an object.
	ObjectValue ValueKind = iota
)

// String returns a description of the ValueKind.
func (k ValueKind) String() string {
	switch k {
	case NullValue:
		return "null"
	case BoolValue:
		return "boolean"
	case NumberValue:
		return "number"
	case StringValue:
		return "string"
	case ArrayValue:
		return "array"
	case ObjectValue:
		return "object"
	default:
		return "unknown token"
	}
}

// Readable is an interface for types that can read their data from a Reader.
type Readable interface {
	// ReadFromJSONReader attempts to read the object's state from a Reader.
	//
	// This method does not need to return an error value because Reader remembers when it
	// has encountered an error.
	ReadFromJSONReader(*Reader)
}
