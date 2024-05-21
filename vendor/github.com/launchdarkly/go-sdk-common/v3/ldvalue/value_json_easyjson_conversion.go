//go:build launchdarkly_easyjson

package ldvalue

import (
	"github.com/mailru/easyjson/jlexer"
	ej_jwriter "github.com/mailru/easyjson/jwriter"
)

// This conditionally-compiled file provides custom marshal/unmarshal functions for ldvalue types
// in EasyJSON.
//
// EasyJSON's code generator does recognize the same MarshalJSON and UnmarshalJSON methods used by
// encoding/json, and will call them if present. But this mechanism is inefficient: when marshaling
// it requires the allocation of intermediate byte slices, and when unmarshaling it causes the
// JSON object to be parsed twice. It is preferable to have our marshal/unmarshal methods write to
// and read from the EasyJSON Writer/Lexer directly. Also, since deserialization is a high-traffic
// path in some LaunchDarkly code on the service side, the extra overhead of the go-jsonstream
// abstraction is undesirable.

// For more information, see: https://github.com/launchdarkly/go-jsonstream/v3

func (v Value) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	switch v.valueType {
	case NullType:
		writer.Raw(nullAsJSONBytes, nil)
	case BoolType:
		writer.Bool(v.boolValue)
	case NumberType:
		writer.Float64(v.numberValue)
	case StringType:
		writer.String(v.stringValue)
	case ArrayType:
		v.arrayValue.MarshalEasyJSON(writer)
	case ObjectType:
		v.objectValue.MarshalEasyJSON(writer)
	case RawType:
		writer.Raw(v.rawValue, nil)
	}
}

func (v *Value) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsDelim('[') {
		var va ValueArray
		va.UnmarshalEasyJSON(lexer)
		*v = Value{valueType: ArrayType, arrayValue: va}
	} else if lexer.IsDelim('{') {
		var vm ValueMap
		vm.UnmarshalEasyJSON(lexer)
		*v = Value{valueType: ObjectType, objectValue: vm}
	} else {
		*v = CopyArbitraryValue(lexer.Interface())
	}
}

func (a ValueArray) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if a.data == nil {
		writer.Raw(nullAsJSONBytes, nil)
		return
	}
	writer.RawByte('[')
	for i, value := range a.data {
		if i != 0 {
			writer.RawByte(',')
		}
		value.MarshalEasyJSON(writer)
	}
	writer.RawByte(']')
}

func (a *ValueArray) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Null()
		*a = ValueArray{}
		return
	}
	lexer.Delim('[')
	a.data = make([]Value, 0, 4)
	for !lexer.IsDelim(']') {
		var value Value
		value.UnmarshalEasyJSON(lexer)
		a.data = append(a.data, value)
		lexer.WantComma()
	}
	lexer.Delim(']')
}

func (m ValueMap) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if m.data == nil {
		writer.Raw(nullAsJSONBytes, nil) //COVERAGE: EasyJSON optimizations may prevent us from reaching this line
		return
	}
	writer.RawByte('{')
	first := true
	for key, value := range m.data {
		if !first {
			writer.RawByte(',')
		}
		first = false
		writer.String(key)
		writer.RawByte(':')
		value.MarshalEasyJSON(writer)
	}
	writer.RawByte('}')
}

func (m *ValueMap) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Null()
		*m = ValueMap{}
		return
	}
	m.data = make(map[string]Value)
	lexer.Delim('{')
	for !lexer.IsDelim('}') {
		key := string(lexer.String())
		lexer.WantColon()
		var value Value
		value.UnmarshalEasyJSON(lexer)
		m.data[key] = value
		lexer.WantComma()
	}
	lexer.Delim('}')
}
