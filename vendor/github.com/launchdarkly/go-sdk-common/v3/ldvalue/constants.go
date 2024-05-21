package ldvalue

// string literals should be defined here if they are referenced in multiple files

const (
	trueString      = "true"
	falseString     = "false"
	nullAsJSON      = "null"
	noneDescription = "[none]"
)

var (
	nullAsJSONBytes = []byte("null")  //nolint:gochecknoglobals
	trueBytes       = []byte("true")  //nolint:gochecknoglobals
	falseBytes      = []byte("false") //nolint:gochecknoglobals
)

// ValueType is defined in value_base.go

const (
	// NullType describes a null value. Note that the zero value of ValueType is NullType, so the
	// zero of Value is a null value. See [Null].
	NullType ValueType = iota
	// BoolType describes a boolean value. See [Bool].
	BoolType ValueType = iota
	// NumberType describes a numeric value. JSON does not have separate types for int and float, but
	// you can convert to either. See [Int] and [Float64].
	NumberType ValueType = iota
	// StringType describes a string value. See [String].
	StringType ValueType = iota
	// ArrayType describes an array value. See [ArrayOf] and [ArrayBuild].
	ArrayType ValueType = iota
	// ObjectType describes an object (a.k.a. map). See [ObjectBuild].
	ObjectType ValueType = iota
	// RawType describes a json.RawMessage value. See [Raw].
	RawType ValueType = iota
)
