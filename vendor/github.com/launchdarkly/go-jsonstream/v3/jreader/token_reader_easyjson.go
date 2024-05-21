//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package jreader

// This file defines the easyjson-based implementation of the low-level JSON tokenizer, which is used instead
// of token_reader_default.go if the launchdarkly_easyjson build tag is enabled.
//
// For the contract governing the behavior of the exported methods in this type, see the comments on the
// corresponding methods in token_reader_default.go.

import (
	"fmt"
	"strings"

	"github.com/mailru/easyjson/jlexer"
)

type tokenReader struct {
	// We might be initialized either with a pointer to an existing Lexer, in which case we'll use that.
	pLexer *jlexer.Lexer
	// Or, we might be initialized with a byte slice so we must create our own Lexer. We'd like to avoid
	// allocating that on the heap, so we'll store it here.
	inlineLexer          jlexer.Lexer
	posBeforeCallingNull int
	posAfterCallingNull  int
	posBeforeValue       int
}

func newTokenReader(data []byte) tokenReader {
	return tokenReader{inlineLexer: jlexer.Lexer{Data: data}}
}

func newTokenReaderFromEasyjsonLexer(lexer *jlexer.Lexer) tokenReader {
	return tokenReader{pLexer: lexer}
}

func (tr *tokenReader) EOF() bool {
	pLexer := tr.pLexer
	if pLexer == nil {
		pLexer = &tr.inlineLexer
	}
	if pLexer.Error() != nil {
		return true
	}
	pLexer.Consumed()
	return pLexer.Error() == nil
}

func (tr *tokenReader) LastPos() int {
	if tr.pLexer == nil {
		return tr.inlineLexer.GetPos()
	}
	return tr.pLexer.GetPos()
}

func (tr *tokenReader) Null() (bool, error) {
	pLexer := tr.pLexer
	if pLexer == nil {
		pLexer = &tr.inlineLexer
	}
	// The "posBefore/posAfter" stuff is because it's not possible to rewind a Lexer. If we call Null(),
	// and the value isn't null, the Lexer will cache that token for the next read, but GetPos() will
	// still be pointing *after* the token and that'll screw up our error reporting.
	tr.posBeforeCallingNull = pLexer.GetPos()
	if pLexer.IsNull() {
		// Lexer.IsNull can return a misleading true value if there's a parsing error
		if err := pLexer.Error(); err != nil {
			return false, tr.translateLexerError()
		}
		pLexer.Null()
		return true, nil
	}
	tr.posAfterCallingNull = pLexer.GetPos()
	return false, tr.translateLexerError()
}

func (tr *tokenReader) Bool() (bool, error) {
	pLexer := tr.pLexer
	if pLexer == nil {
		pLexer = &tr.inlineLexer
	}
	tr.markPosBeforeValue()
	val := pLexer.Bool()
	if pLexer.Error() == nil {
		return val, nil
	}
	return false, tr.translateLexerErrorWithExpectedType(BoolValue)
}

func (tr *tokenReader) Number() (float64, error) {
	pLexer := tr.pLexer
	if pLexer == nil {
		pLexer = &tr.inlineLexer
	}
	tr.markPosBeforeValue()
	val := pLexer.Float64()
	if pLexer.Error() == nil {
		return val, nil
	}
	return 0, tr.translateLexerErrorWithExpectedType(NumberValue)
}

func (tr *tokenReader) String() (string, error) {
	pLexer := tr.pLexer
	if pLexer == nil {
		pLexer = &tr.inlineLexer
	}
	tr.markPosBeforeValue()
	val := pLexer.String()
	if pLexer.Error() == nil {
		return val, nil
	}
	return "", tr.translateLexerErrorWithExpectedType(StringValue)
}

func (tr *tokenReader) PropertyName() ([]byte, error) {
	pLexer := tr.pLexer
	if pLexer == nil {
		pLexer = &tr.inlineLexer
	}
	val := pLexer.UnsafeBytes()
	if err := pLexer.Error(); err != nil {
		return nil, tr.translateLexerError()
	}
	pLexer.WantColon()
	pLexer.FetchToken()
	return val, tr.translateLexerError()
}

func (tr *tokenReader) Delimiter(delim byte) (bool, error) {
	pLexer := tr.pLexer
	if pLexer == nil {
		pLexer = &tr.inlineLexer
	}
	if err := pLexer.Error(); err != nil {
		return false, tr.translateLexerError()
	}
	found := false
	if pLexer.IsDelim(byte(delim)) {
		pLexer.Delim(delim)
		found = true
	}
	// IsDelim can return a misleading true value if there's a parsing error
	if err := pLexer.Error(); err != nil {
		return false, tr.translateLexerError()
	}
	return found, nil
}

func (tr *tokenReader) EndDelimiterOrComma(delim byte) (bool, error) {
	pLexer := tr.pLexer
	if pLexer == nil {
		pLexer = &tr.inlineLexer
	}
	if pLexer.Error() != nil {
		return false, tr.translateLexerError()
	}
	pLexer.WantComma()
	if pLexer.IsDelim(delim) {
		pLexer.Delim(delim)
		return true, nil
	}
	return false, tr.translateLexerError()
}

func (tr *tokenReader) Any() (AnyValue, error) {
	pLexer := tr.pLexer
	if pLexer == nil {
		pLexer = &tr.inlineLexer
	}
	value, err := readAnyValue(pLexer)
	if err != nil {
		return AnyValue{}, tr.translateLexerError()
	}
	return value, nil
}

func (tr *tokenReader) lexerError() error {
	if tr.pLexer == nil {
		return tr.inlineLexer.Error()
	}
	return tr.pLexer.Error()
}

func readAnyValue(lexer *jlexer.Lexer) (AnyValue, error) {
	if lexer.IsDelim('[') {
		// IsDelim can return a misleading true value if there's a parsing error
		if err := lexer.Error(); err != nil {
			return AnyValue{}, err
		}
		lexer.Delim('[')
		return AnyValue{Kind: ArrayValue}, nil
	}
	if lexer.IsDelim('{') {
		if err := lexer.Error(); err != nil {
			return AnyValue{}, err
		}
		lexer.Delim('{')
		return AnyValue{Kind: ObjectValue}, nil
	}
	intf := lexer.Interface()
	if err := lexer.Error(); err != nil {
		return AnyValue{}, err
	}
	if intf == nil {
		return AnyValue{Kind: NullValue}, nil
	}
	switch v := intf.(type) {
	case bool:
		return AnyValue{Kind: BoolValue, Bool: v}, nil
	case int:
		return AnyValue{Kind: NumberValue, Number: float64(v)}, nil
	case float64:
		return AnyValue{Kind: NumberValue, Number: v}, nil
	case string:
		return AnyValue{Kind: StringValue, String: v}, nil
	}
	return AnyValue{}, fmt.Errorf("Lexer.Interface() returned unrecognized type %T", intf)
}

func (tr *tokenReader) markPosBeforeValue() {
	pos := tr.LastPos()
	if pos == tr.posAfterCallingNull {
		pos = tr.posBeforeCallingNull
	}
}

func (tr *tokenReader) translateLexerError() error {
	return tr.translateLexerErrorWithExpectedType(-1)
}

func (tr *tokenReader) translateLexerErrorWithExpectedType(expectedType ValueKind) error {
	pLexer := tr.pLexer
	if pLexer == nil {
		pLexer = &tr.inlineLexer
	}
	originalError := pLexer.Error()
	if originalError == nil {
		return nil
	}
	le, ok := originalError.(*jlexer.LexerError)
	if !ok {
		return originalError
	}
	if strings.HasPrefix(le.Reason, "expected ") && expectedType >= 0 {
		// LexerError is not very useful for determining what the invalid token was, because it tends to
		// leave the Data property empty and put the Offset property *after* the bad token. Fortunately,
		// it's very easy to create a new Lexer to re-parse from where we started.
		tempLexer := jlexer.Lexer{Data: pLexer.Data[tr.posBeforeValue:]}
		value, err := readAnyValue(&tempLexer)
		if err != nil {
			return translateLexerParseError(err)
		}
		return TypeError{Expected: expectedType, Actual: value.Kind, Offset: tr.posBeforeValue}
	}
	return translateLexerParseError(originalError)
}

func translateLexerParseError(err error) error {
	if le, ok := err.(*jlexer.LexerError); ok {
		return SyntaxError{Message: strings.TrimPrefix(le.Reason, "parse error: "), Offset: le.Offset}
	}
	return err
}
