package jreader

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	errMsgBadArrayItem     = "expected comma or end of array"
	errMsgBadObjectItem    = "expected comma or end of object"
	errMsgDataAfterEnd     = "unexpected data after end of JSON value"
	errMsgExpectedColon    = "expected colon after property name"
	errMsgInvalidNumber    = "invalid numeric value"
	errMsgInvalidString    = "unterminated or invalid string value"
	errMsgUnexpectedChar   = "unexpected character"
	errMsgUnexpectedSymbol = "unexpected symbol"
)

// SyntaxError is returned by Reader if the input is not well-formed JSON.
type SyntaxError struct {
	// Message is a descriptive message.
	Message string

	// Offset is the approximate character index within the input where the error occurred.
	Offset int

	// Value, if not empty, is the token that caused the error.
	Value string
}

// TypeError is returned by Reader if the type of JSON value that was read did not
// match what the caller requested.
type TypeError struct {
	// Expected is the type of JSON value that the caller requested.
	Expected ValueKind

	// Actual is the type of JSON value that was found.
	Actual ValueKind

	// Nullable is true if the caller indicated that a null value was acceptable in this context.
	Nullable bool

	// Offset is the approximate character index within the input where the error occurred.
	Offset int
}

// RequiredPropertyError is returned by Reader if a JSON object did not contain a property that
// was designated as required (by using ObjectState.WithRequiredProperties).
type RequiredPropertyError struct {
	// Name is the name of a required object property that was not found.
	Name string

	// Offset is the approximate character index within the input where the error occurred
	// (at or near the end of the JSON object).
	Offset int
}

// Error returns a description of the error.
func (e SyntaxError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("%s at position %d (%q)", e.Message, e.Offset, e.Value)
	}
	return fmt.Sprintf("%s at position %d", e.Message, e.Offset)
}

// Error returns a description of the error.
func (e TypeError) Error() string {
	if e.Nullable {
		return fmt.Sprintf("expected %s or null, got %s at position %d", e.Expected, e.Actual, e.Offset)
	}
	return fmt.Sprintf("expected %s, got %s at position %d", e.Expected, e.Actual, e.Offset)
}

// Error returns a description of the error.
func (e RequiredPropertyError) Error() string {
	return fmt.Sprintf("a required property %q was missing from a JSON object at position %d", e.Name, e.Offset)
}

// ToJSONError converts errors defined by the jreader package into the corresponding error types defined
// by the encoding/json package, if any. The target parameter, if not nil, is used to determine the
// target value type for json.UnmarshalTypeError.
func ToJSONError(err error, target interface{}) error {
	switch e := err.(type) {
	case SyntaxError:
		return &json.SyntaxError{
			Offset: int64(e.Offset),
		}
	case TypeError:
		return &json.UnmarshalTypeError{
			Value:  e.Expected.String(),
			Type:   reflect.TypeOf(target),
			Offset: int64(e.Offset),
		}
	}
	return err
}
