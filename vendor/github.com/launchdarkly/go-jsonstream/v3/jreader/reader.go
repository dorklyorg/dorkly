package jreader

// Reader is a high-level API for reading JSON data sequentially.
//
// It is designed to make writing custom unmarshallers for application types as convenient as
// possible. The general usage pattern is as follows:
//
// - Values are parsed in the order that they appear.
//
// - In general, the caller should know what data type is expected. Since it is common for
// properties to be nullable, the methods for reading scalar types have variants for allowing
// a null instead of the specified type. If the type is completely unknown, use Any.
//
// - For reading array or object structures, the Array and Object methods return a struct that
// keeps track of additional reader state while that structure is being parsed.
//
// - If any method encounters an error (due to either malformed JSON, or well-formed JSON that
// did not match the caller's data type expectations), the Reader permanently enters a failed
// state and remembers that error; all subsequent method calls will return the same error and no
// more parsing will happen. This means that the caller does not necessarily have to check the
// error return value of any individual method, although it can.
type Reader struct {
	tr                tokenReader
	awaitingReadValue bool // used by ArrayState & ObjectState
	err               error
}

// Error returns the first error that the Reader encountered, if the Reader is in a failed state,
// or nil if it is still in a good state.
func (r *Reader) Error() error {
	return r.err
}

// RequireEOF returns nil if all of the input has been consumed (not counting whitespace), or an
// error if not.
func (r *Reader) RequireEOF() error {
	if !r.tr.EOF() {
		return SyntaxError{Message: errMsgDataAfterEnd, Offset: r.tr.LastPos()}
	}
	return nil
}

// AddError sets the Reader's error value and puts it into a failed state. If the parameter is nil
// or the Reader was already in a failed state, it does nothing.
func (r *Reader) AddError(err error) {
	if r.err == nil {
		r.err = err
	}
}

// ReplaceError sets the Reader's error value and puts it into a failed state, replacing any
// previously reported error. If the parameter is nil, it does nothing (a failed state cannot be
// changed to a non-failed state).
func (r *Reader) ReplaceError(err error) {
	if err != nil {
		r.err = err
	}
}

// Null attempts to read a null value, returning an error if the next token is not a null.
func (r *Reader) Null() error {
	r.awaitingReadValue = false
	if r.err != nil {
		return r.err
	}
	isNull, err := r.tr.Null()
	if isNull || err != nil {
		return err
	}
	return r.typeErrorForCurrentToken(NullValue, false)
}

// Bool attempts to read a boolean value.
//
// If there is a parsing error, or the next value is not a boolean, the return value is false
// and the Reader enters a failed state, which you can detect with Error().
func (r *Reader) Bool() bool {
	r.awaitingReadValue = false
	if r.err != nil {
		return false
	}
	val, err := r.tr.Bool()
	if err != nil {
		r.err = err
		return false
	}
	return val
}

// BoolOrNull attempts to read either a boolean value or a null. In the case of a boolean, the return
// values are (value, true); for a null, they are (false, false).
//
// If there is a parsing error, or the next value is neither a boolean nor a null, the return values
// are (false, false) and the Reader enters a failed state, which you can detect with Error().
func (r *Reader) BoolOrNull() (value bool, nonNull bool) {
	r.awaitingReadValue = false
	if r.err != nil {
		return false, false
	}
	isNull, err := r.tr.Null()
	if isNull || err != nil {
		r.err = err
		return false, false
	}
	val, err := r.tr.Bool()
	if err != nil {
		r.err = typeErrorForNullableValue(err)
		return false, false
	}
	return val, true
}

// Int attempts to read a numeric value and returns it as an int.
//
// If there is a parsing error, or the next value is not a number, the return value is zero and
// the Reader enters a failed state, which you can detect with Error(). Non-numeric types are never
// converted to numbers.
func (r *Reader) Int() int {
	return int(r.Float64())
}

// IntOrNull attempts to read either an integer numeric value or a null. In the case of a number, the
// return values are (value, true); for a null, they are (0, false).
//
// If there is a parsing error, or the next value is neither a number nor a null, the return values
// are (0, false) and the Reader enters a failed state, which you can detect with Error().
func (r *Reader) IntOrNull() (int, bool) {
	val, nonNull := r.Float64OrNull()
	return int(val), nonNull
}

// Float64 attempts to read a numeric value and returns it as a float64.
//
// If there is a parsing error, or the next value is not a number, the return value is zero and
// the Reader enters a failed state, which you can detect with Error(). Non-numeric types are never
// converted to numbers.
func (r *Reader) Float64() float64 {
	r.awaitingReadValue = false
	if r.err != nil {
		return 0
	}
	val, err := r.tr.Number()
	if err != nil {
		r.err = err
		return 0
	}
	return val
}

// Float64OrNull attempts to read either a numeric value or a null. In the case of a number, the
// return values are (value, true); for a null, they are (0, false).
//
// If there is a parsing error, or the next value is neither a number nor a null, the return values
// are (0, false) and the Reader enters a failed state, which you can detect with Error().
func (r *Reader) Float64OrNull() (float64, bool) {
	r.awaitingReadValue = false
	if r.err != nil {
		return 0, false
	}
	isNull, err := r.tr.Null()
	if isNull || err != nil {
		r.err = err
		return 0, false
	}
	val, err := r.tr.Number()
	if err != nil {
		r.err = typeErrorForNullableValue(err)
		return 0, false
	}
	return val, true
}

// String attempts to read a string value.
//
// If there is a parsing error, or the next value is not a string, the return value is "" and
// the Reader enters a failed state, which you can detect with Error(). Types other than string
// are never converted to strings.
func (r *Reader) String() string {
	r.awaitingReadValue = false
	if r.err != nil {
		return ""
	}
	val, err := r.tr.String()
	if err != nil {
		r.err = err
		return ""
	}
	return val
}

// StringOrNull attempts to read either a string value or a null. In the case of a string, the
// return values are (value, true); for a null, they are ("", false).
//
// If there is a parsing error, or the next value is neither a string nor a null, the return values
// are ("", false) and the Reader enters a failed state, which you can detect with Error().
func (r *Reader) StringOrNull() (string, bool) {
	r.awaitingReadValue = false
	if r.err != nil {
		return "", false
	}
	isNull, err := r.tr.Null()
	if isNull || err != nil {
		r.err = err
		return "", false
	}
	val, err := r.tr.String()
	if err != nil {
		r.err = typeErrorForNullableValue(err)
		return "", false
	}
	return val, true
}

// Array attempts to begin reading a JSON array value. If successful, the return value will be an
// ArrayState containing the necessary state for iterating through the array elements.
//
// The ArrayState is used only for the iteration state; to read the value of each array element, you
// will still use the Reader's methods.
//
// If there is a parsing error, or the next value is not an array, the returned ArrayState is a stub
// whose Next() method always returns false, and the Reader enters a failed state, which you can
// detect with Error().
//
// See ArrayState for example code.
func (r *Reader) Array() ArrayState {
	return r.tryArray(false)
}

// ArrayOrNull attempts to either begin reading an JSON array value, or read a null. In the case of an
// array, the return value will be an ArrayState containing the necessary state for iterating through
// the array elements; the ArrayState's IsDefined() method will return true. In the case of a null, the
// returned ArrayState will be a stub whose Next() and IsDefined() methods always returns false.
//
// The ArrayState is used only for the iteration state; to read the value of each array element, you
// will still use the Reader's methods.
//
// If there is a parsing error, or the next value is neither an array nor a null, the return value is
// the same as for a null but the Reader enters a failed state, which you can detect with Error().
//
// See ArrayState for example code.
func (r *Reader) ArrayOrNull() ArrayState {
	return r.tryArray(true)
}

func (r *Reader) tryArray(allowNull bool) ArrayState {
	r.awaitingReadValue = false
	if r.err != nil {
		return ArrayState{}
	}
	if allowNull {
		isNull, err := r.tr.Null()
		if err != nil {
			r.err = err
			return ArrayState{}
		}
		if isNull {
			return ArrayState{}
		}
	}
	gotDelim, err := r.tr.Delimiter('[')
	if err != nil {
		r.err = err
		return ArrayState{}
	}
	if gotDelim {
		return ArrayState{r: r}
	}
	r.err = r.typeErrorForCurrentToken(ArrayValue, allowNull)
	return ArrayState{}
}

// Object attempts to begin reading a JSON object value. If successful, the return value will be an
// ObjectState containing the necessary state for iterating through the object properties.
//
// The ObjectState is used only for the iteration state; to read the value of each property, you
// will still use the Reader's methods.
//
// If there is a parsing error, or the next value is not an object, the returned ObjectState is a stub
// whose Next() method always returns false, and the Reader enters a failed state, which you can
// detect with Error().
//
// See ObjectState for example code.
func (r *Reader) Object() ObjectState {
	return r.tryObject(false)
}

// ObjectOrNull attempts to either begin reading an JSON object value, or read a null. In the case of an
// object, the return value will be an ObjectState containing the necessary state for iterating through
// the object properties; the ObjectState's IsDefined() method will return true. In the case of a null,
// the returned ObjectState will be a stub whose Next() and IsDefined() methods always returns false.
//
// The ObjectState is used only for the iteration state; to read the value of each property, you
// will still use the Reader's methods.
//
// If there is a parsing error, or the next value is neither an object nor a null, the return value is
// the same as for a null but the Reader enters a failed state, which you can detect with Error().
//
// See ObjectState for example code.
func (r *Reader) ObjectOrNull() ObjectState {
	return r.tryObject(true)
}

func (r *Reader) tryObject(allowNull bool) ObjectState {
	r.awaitingReadValue = false
	if r.err != nil {
		return ObjectState{}
	}
	if allowNull {
		isNull, err := r.tr.Null()
		if err != nil || isNull {
			r.err = err
			return ObjectState{}
		}
	}
	gotDelim, err := r.tr.Delimiter('{')
	if err != nil {
		r.err = err
		return ObjectState{}
	}
	if gotDelim {
		return ObjectState{r: r}
	}
	r.err = r.typeErrorForCurrentToken(ObjectValue, allowNull)
	return ObjectState{}
}

// Any reads a single value of any type, if it is a scalar value or a null, or prepares to read
// the value if it is an array or object.
//
// The returned AnyValue's Kind field indicates the value type. If it is BoolValue, NumberValue,
// or StringValue, check the corresponding Bool, Number, or String property. If it is ArrayValue
// or ObjectValue, the AnyValue's Array or Object field has been initialized with an ArrayState or
// ObjectState just as if you had called the Reader's Array or Object method.
//
// If there is a parsing error, the return value is the same as for a null and the Reader enters
// a failed state, which you can detect with Error().
func (r *Reader) Any() AnyValue {
	r.awaitingReadValue = false
	if r.err != nil {
		return AnyValue{}
	}
	v, err := r.tr.Any()
	if err != nil {
		r.err = err
		return AnyValue{}
	}
	switch v.Kind {
	case BoolValue:
		return AnyValue{Kind: v.Kind, Bool: v.Bool}
	case NumberValue:
		return AnyValue{Kind: v.Kind, Number: v.Number}
	case StringValue:
		return AnyValue{Kind: v.Kind, String: v.String}
	case ArrayValue:
		return AnyValue{Kind: v.Kind, Array: ArrayState{r: r}}
	case ObjectValue:
		return AnyValue{Kind: v.Kind, Object: ObjectState{r: r}}
	default:
		return AnyValue{Kind: NullValue}
	}
}

// SkipValue consumes and discards the next JSON value of any type. For an array or object value, it
// recurses to also consume and discard all array elements or object properties.
func (r *Reader) SkipValue() error {
	r.awaitingReadValue = false
	if r.err != nil {
		return r.err
	}
	v := r.Any()
	if v.Kind == ArrayValue {
		for v.Array.Next() {
		}
	} else if v.Kind == ObjectValue {
		for v.Object.Next() {
		}
	}
	return r.err
}

func typeErrorForNullableValue(err error) error {
	if err != nil {
		switch e := err.(type) { //nolint:gocritic
		case TypeError:
			e.Nullable = true
			return e
		}
	}
	return err
}

func (r *Reader) typeErrorForCurrentToken(expected ValueKind, nullable bool) error {
	v, err := r.tr.Any()
	if err != nil {
		return err
	}
	return TypeError{Expected: expected, Actual: v.Kind, Offset: r.tr.LastPos(), Nullable: nullable}
}
