//go:build launchdarkly_easyjson

package ldreason

import (
	"github.com/launchdarkly/go-jsonstream/v3/jreader"
	"github.com/launchdarkly/go-jsonstream/v3/jwriter"

	"github.com/mailru/easyjson/jlexer"
	ej_jwriter "github.com/mailru/easyjson/jwriter"
)

// This conditionally-compiled file provides custom marshal/unmarshal functions for the EvaluationReason
// type in EasyJSON.
//
// EasyJSON's code generator does recognize the same MarshalJSON and UnmarshalJSON methods used by
// encoding/json, and will call them if present. But this mechanism is inefficient: when marshaling
// it requires the allocation of intermediate byte slices, and when unmarshaling it causes the
// JSON object to be parsed twice. It is preferable to have our marshal/unmarshal methods write to
// and read from the EasyJSON Writer/Lexer directly. Our go-jsonstream library provides methods for
// doing this, if the launchdarkly_easyjson build tag is set.package ldmodel
//
// For more information, see: https://github.com/launchdarkly/go-jsonstream/v3

func (r EvaluationReason) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	wrappedWriter := jwriter.NewWriterFromEasyJSONWriter(writer)
	r.WriteToJSONWriter(&wrappedWriter)
}

func (r *EvaluationReason) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	wrappedReader := jreader.NewReaderFromEasyJSONLexer(lexer)
	r.ReadFromJSONReader(&wrappedReader)
	lexer.AddError(wrappedReader.Error())
}
