//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package jreader

import (
	"github.com/mailru/easyjson/jlexer"
)

// NewReaderFromEasyJSONLexer creates a Reader that consumes JSON input data from the specified easyjson
// jlexer.Lexer.
//
// This function is only available in code that was compiled with the build tag "launchdarkly_easyjson".
// Its purpose is to allow custom unmarshaling code that is based on the Reader API to be used as
// efficiently as possible within other data structures that are being unmarshaled with easyjson.
// Directly using the same Lexer that is already being used is more efficient than asking Lexer to
// scan the next object, return it as a byte slice, and then pass that byte slice to NewReader.
func NewReaderFromEasyJSONLexer(lexer *jlexer.Lexer) Reader {
	return Reader{
		tr: newTokenReaderFromEasyjsonLexer(lexer),
	}
}
