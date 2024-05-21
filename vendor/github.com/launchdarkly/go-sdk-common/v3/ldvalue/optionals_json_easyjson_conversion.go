//go:build launchdarkly_easyjson

package ldvalue

import (
	"github.com/mailru/easyjson/jlexer"
	ej_jwriter "github.com/mailru/easyjson/jwriter"
)

// MarshalEasyJSON implements the easyjson.Marshaler interface.
//
// This method is only defined if the launchdarkly_easyjson build tag is set. For more information,
// see the package documentation for ldvalue.
func (v OptionalBool) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if value, ok := v.optValue.get(); ok {
		writer.Bool(value)
	} else {
		writer.Raw(nullAsJSONBytes, nil)
	}
}

// MarshalEasyJSON implements the easyjson.Marshaler interface.
//
// This method is only defined if the launchdarkly_easyjson build tag is set. For more information,
// see the package documentation for ldvalue.
func (v OptionalInt) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if value, ok := v.optValue.get(); ok {
		writer.Int(value)
	} else {
		writer.Raw(nullAsJSONBytes, nil)
	}
}

// MarshalEasyJSON implements the easyjson.Marshaler interface.
//
// This method is only defined if the launchdarkly_easyjson build tag is set. For more information,
// see the package documentation for ldvalue.
func (v OptionalString) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if value, ok := v.optValue.get(); ok {
		writer.String(value)
	} else {
		writer.Raw(nullAsJSONBytes, nil)
	}
}

// UnmarshalEasyJSON implements the easyjson.Unmarshaler interface.
//
// This method is only defined if the launchdarkly_easyjson build tag is set. For more information,
// see the package documentation for ldvalue.
func (v *OptionalBool) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Null()
		*v = OptionalBool{}
		return
	}
	*v = NewOptionalBool(lexer.Bool())
}

// UnmarshalEasyJSON implements the easyjson.Unmarshaler interface.
//
// This method is only defined if the launchdarkly_easyjson build tag is set. For more information,
// see the package documentation for ldvalue.
func (v *OptionalInt) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Null()
		*v = OptionalInt{}
		return
	}
	*v = NewOptionalInt(lexer.Int())
}

// UnmarshalEasyJSON implements the easyjson.Unmarshaler interface.
//
// This method is only defined if the launchdarkly_easyjson build tag is set. For more information,
// see the package documentation for ldvalue.
func (v *OptionalString) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Null()
		*v = OptionalString{}
		return
	}
	*v = NewOptionalString(lexer.String())
}
