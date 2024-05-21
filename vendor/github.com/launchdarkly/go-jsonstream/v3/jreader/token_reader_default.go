//go:build !launchdarkly_easyjson
// +build !launchdarkly_easyjson

package jreader

// This file defines the default implementation of the low-level JSON tokenizer. If the launchdarkly_easyjson
// build tag is enabled, we use the easyjson adapter in token_reader_easyjson.go instead. These have the same
// methods so the Reader code does not need to know which implementation we're using; however, we don't
// actually define an interface for these, because calling the methods through an interface would limit
// performance.

import (
	"bytes"
	"io"
	"strconv"
	"unicode"
	"unicode/utf8"
)

var (
	tokenNull  = []byte("null")  //nolint:gochecknoglobals
	tokenTrue  = []byte("true")  //nolint:gochecknoglobals
	tokenFalse = []byte("false") //nolint:gochecknoglobals
)

type token struct {
	kind        tokenKind
	boolValue   bool
	numberValue float64
	stringValue []byte
	delimiter   byte
}

type tokenKind int

const (
	nullToken      tokenKind = iota
	boolToken      tokenKind = iota
	numberToken    tokenKind = iota
	stringToken    tokenKind = iota
	delimiterToken tokenKind = iota
)

func (t token) valueKind() ValueKind {
	if t.kind == delimiterToken {
		if t.delimiter == '[' {
			return ArrayValue
		}
		if t.delimiter == '{' {
			return ObjectValue
		}
	}
	return valueKindFromTokenKind(t.kind)
}

func (t token) description() string {
	if t.kind == delimiterToken && t.delimiter != '[' && t.delimiter != '{' {
		return "'" + string(t.delimiter) + "'"
	}
	return t.valueKind().String()
}

type tokenReader struct {
	data        []byte
	pos         int
	len         int
	hasUnread   bool
	unreadToken token
	lastPos     int
}

func newTokenReader(data []byte) tokenReader {
	tr := tokenReader{
		data: data,
		pos:  0,
		len:  len(data),
	}
	return tr
}

// EOF returns true if we are at the end of the input (not counting whitespace).
func (r *tokenReader) EOF() bool {
	if r.hasUnread {
		return false
	}
	_, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return true
	}
	r.unreadByte()
	return false
}

// LastPos returns the byte offset within the input where we most recently started parsing a token.
func (r *tokenReader) LastPos() int {
	return r.lastPos
}

func (r *tokenReader) getPos() int {
	if r.hasUnread {
		return r.lastPos
	}
	return r.pos
}

// Null returns (true, nil) if the next token is a null (consuming the token); (false, nil) if the next
// token is not a null (not consuming the token); or (false, error) if the next token is not a valid
// JSON value.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) Null() (bool, error) {
	t, err := r.next()
	if err != nil {
		return false, err
	}
	if t.kind == nullToken {
		return true, nil
	}
	r.putBack(t)
	if t.kind == delimiterToken && t.delimiter != '[' && t.delimiter != '{' {
		return false, SyntaxError{Message: errMsgUnexpectedChar, Value: string(t.delimiter), Offset: r.getPos()}
	}
	return false, nil
}

// Bool requires that the next token is a JSON boolean, returning its value if successful (consuming
// the token), or an error if the next token is anything other than a JSON boolean.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) Bool() (bool, error) {
	t, err := r.consumeScalar(boolToken)
	return t.boolValue, err
}

// Bool requires that the next token is a JSON number, returning its value if successful (consuming
// the token), or an error if the next token is anything other than a JSON number.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) Number() (float64, error) {
	t, err := r.consumeScalar(numberToken)
	return t.numberValue, err
}

// String requires that the next token is a JSON string, returning its value if successful (consuming
// the token), or an error if the next token is anything other than a JSON string.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) String() (string, error) {
	t, err := r.consumeScalar(stringToken)
	return string(t.stringValue), err
}

// PropertyName requires that the next token is a JSON string and the token after that is a colon,
// returning the string as a byte slice if successful, or an error otherwise.
//
// Returning the string as a byte slice avoids the overhead of allocating a string, since normally
// the names of properties will not be retained as strings but are only compared to constants while
// parsing an object.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) PropertyName() ([]byte, error) {
	t, err := r.consumeScalar(stringToken)
	if err != nil {
		return nil, err
	}
	b, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return nil, io.EOF
	}
	if b != ':' {
		r.unreadByte()
		return nil, r.syntaxErrorOnNextToken(errMsgExpectedColon)
	}
	return t.stringValue, nil
}

// Delimiter checks whether the next token is the specified ASCII delimiter character. If so, it
// returns (true, nil) and consumes the token. If it is a delimiter, but not the same one, it
// returns (false, nil) and does not consume the token. For anything else, it returns an error.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) Delimiter(delimiter byte) (bool, error) {
	if r.hasUnread {
		if r.unreadToken.kind == delimiterToken && r.unreadToken.delimiter == delimiter {
			r.hasUnread = false
			return true, nil
		}
		return false, nil
	}
	b, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return false, nil
	}
	if b == delimiter {
		return true, nil
	}
	r.unreadByte() // we'll back up and try to parse a token, to see if it's valid JSON or not
	token, err := r.next()
	if err != nil {
		return false, err // it was malformed JSON
	}
	r.putBack(token) // it was valid JSON, we just haven't hit that delimiter
	return false, nil
}

// EndDelimiterOrComma checks whether the next token is the specified ASCII delimiter character
// or a comma. If it is the specified delimiter, it returns (true, nil) and consumes the token.
// If it is a comma, it returns (false, nil) and consumes the token. For anything else, it
// returns an error. The delimiter parameter will always be either '}' or ']'.
func (r *tokenReader) EndDelimiterOrComma(delimiter byte) (bool, error) {
	if r.hasUnread {
		if r.unreadToken.kind == delimiterToken &&
			(r.unreadToken.delimiter == delimiter || r.unreadToken.delimiter == ',') {
			r.hasUnread = false
			return r.unreadToken.delimiter == delimiter, nil
		}
		return false, SyntaxError{Message: badArrayOrObjectItemMessage(delimiter == '}'),
			Value: r.unreadToken.description(), Offset: r.lastPos}
	}
	b, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return false, io.EOF
	}
	if b == delimiter || b == ',' {
		return b == delimiter, nil
	}
	r.unreadByte()
	t, err := r.next()
	if err != nil {
		return false, err
	}
	return false, SyntaxError{Message: badArrayOrObjectItemMessage(delimiter == '}'),
		Value: t.description(), Offset: r.lastPos}
}

func badArrayOrObjectItemMessage(isObject bool) string {
	if isObject {
		return errMsgBadObjectItem
	}
	return errMsgBadArrayItem
}

// Any checks whether the next token is either a valid JSON scalar value or the opening delimiter of
// an array or object value. If so, it returns (AnyValue, nil) and consumes the token; if not, it
// returns an error. Unlike Reader.Any(), for array and object values it does not create an
// ArrayState or ObjectState.
func (r *tokenReader) Any() (AnyValue, error) {
	t, err := r.next()
	if err != nil {
		return AnyValue{}, err
	}
	switch t.kind {
	case boolToken:
		return AnyValue{Kind: BoolValue, Bool: t.boolValue}, nil
	case numberToken:
		return AnyValue{Kind: NumberValue, Number: t.numberValue}, nil
	case stringToken:
		return AnyValue{Kind: StringValue, String: string(t.stringValue)}, nil
	case delimiterToken:
		if t.delimiter == '[' {
			return AnyValue{Kind: ArrayValue}, nil
		}
		if t.delimiter == '{' {
			return AnyValue{Kind: ObjectValue}, nil
		}
		return AnyValue{},
			SyntaxError{Message: errMsgUnexpectedChar, Value: string(t.delimiter), Offset: r.lastPos}
	default:
		return AnyValue{Kind: NullValue}, nil
	}
}

// Attempts to parse and consume the next token, ignoring whitespace. A token is either a valid JSON scalar
// value or an ASCII delimiter character. If a token was previously unread using putBack, it consumes that
// instead.
func (r *tokenReader) next() (token, error) {
	if r.hasUnread {
		r.hasUnread = false
		return r.unreadToken, nil
	}
	b, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return token{}, io.EOF
	}

	switch {
	// We can get away with reading bytes instead of runes because the JSON spec doesn't allow multi-byte
	// characters except within a string literal.
	case b >= 'a' && b <= 'z':
		n := r.consumeASCIILowercaseAlphabeticChars() + 1
		id := r.data[r.lastPos : r.lastPos+n]
		if b == 'f' && bytes.Equal(id, tokenFalse) {
			return token{kind: boolToken, boolValue: false}, nil
		}
		if b == 't' && bytes.Equal(id, tokenTrue) {
			return token{kind: boolToken, boolValue: true}, nil
		}
		if b == 'n' && bytes.Equal(id, tokenNull) {
			return token{kind: nullToken}, nil
		}
		return token{}, SyntaxError{Message: errMsgUnexpectedSymbol, Value: string(id), Offset: r.lastPos}
	case (b >= '0' && b <= '9') || b == '-':
		if n, ok := r.readNumber(b); ok {
			return token{kind: numberToken, numberValue: n}, nil
		}
		return token{}, SyntaxError{Message: errMsgInvalidNumber, Offset: r.lastPos}
	case b == '"':
		s, err := r.readString()
		if err != nil {
			return token{}, err
		}
		return token{kind: stringToken, stringValue: s}, nil
	case b == '[', b == ']', b == '{', b == '}', b == ':', b == ',':
		return token{kind: delimiterToken, delimiter: b}, nil
	}

	return token{}, SyntaxError{Message: errMsgUnexpectedChar, Value: string(b), Offset: r.lastPos}
}

func (r *tokenReader) putBack(token token) {
	r.unreadToken = token
	r.hasUnread = true
}

func (r *tokenReader) consumeScalar(kind tokenKind) (token, error) {
	t, err := r.next()
	if err != nil {
		return token{}, err
	}
	if t.kind == kind {
		return t, nil
	}
	if t.kind == delimiterToken && t.delimiter != '[' && t.delimiter != '{' {
		return token{}, SyntaxError{Message: errMsgUnexpectedChar, Value: string(t.delimiter), Offset: r.LastPos()}
	}
	return token{}, TypeError{Expected: valueKindFromTokenKind(kind),
		Actual: t.valueKind(), Offset: r.LastPos()}
}

func (r *tokenReader) readByte() (byte, bool) {
	if r.pos >= r.len {
		return 0, false
	}
	b := r.data[r.pos]
	r.pos++
	return b, true
}

func (r *tokenReader) unreadByte() {
	r.pos--
}

func (r *tokenReader) skipWhitespaceAndReadByte() (byte, bool) {
	for {
		ch, ok := r.readByte()
		if !ok {
			return 0, false
		}
		if !unicode.IsSpace(rune(ch)) {
			r.lastPos = r.pos - 1
			return ch, true
		}
	}
}

func (r *tokenReader) consumeASCIILowercaseAlphabeticChars() int {
	n := 0
	for {
		ch, ok := r.readByte()
		if !ok {
			break
		}
		if ch < 'a' || ch > 'z' {
			r.unreadByte()
			break
		}
		n++
	}
	return n
}

func (r *tokenReader) readNumber(first byte) (float64, bool) { //nolint:unparam
	startPos := r.lastPos
	isFloat := false
	var ch byte
	var ok bool
	for {
		ch, ok = r.readByte()
		if !ok {
			break
		}
		if (ch < '0' || ch > '9') && !(ch == '.' && !isFloat) {
			break
		}
		if ch == '.' {
			isFloat = true
		}
	}
	hasExponent := false
	if ch == 'e' || ch == 'E' {
		// exponent must match this regex: [eE][-+]?[0-9]+
		ch, ok = r.readByte()
		if !ok {
			return 0, false
		}
		if ch == '+' || ch == '-' { //nolint:gocritic
		} else if ch >= '0' && ch <= '9' {
			r.unreadByte()
		} else {
			return 0, false
		}
		for {
			ch, ok = r.readByte()
			if !ok {
				break
			}
			if ch < '0' || ch > '9' {
				r.unreadByte()
				break
			}
			hasExponent = true
		}
		if !hasExponent {
			return 0, false
		}
		isFloat = true
	} else { //nolint:gocritic
		if ok {
			r.unreadByte()
		}
	}
	chars := r.data[startPos:r.pos]
	if isFloat {
		// Unfortunately, strconv.ParseFloat requires a string - there is no []byte equivalent. This means we can't
		// avoid a heap allocation here. Easyjson works around this by creating an unsafe string that points directly
		// at the existing bytes, but in our default implementation we can't use unsafe.
		n, err := strconv.ParseFloat(string(chars), 64)
		return n, err == nil
	} else { //nolint:revive
		n, ok := parseIntFromBytes(chars)
		return float64(n), ok
	}
}

func (r *tokenReader) readString() ([]byte, error) {
	startPos := r.pos // the opening quote mark has already been read
	var chars []byte
	haveEscaped := false
	var reader bytes.Reader // bytes.Reader understands multi-byte characters
	reader.Reset(r.data)
	_, _ = reader.Seek(int64(r.pos), io.SeekStart)

	for {
		ch, _, err := reader.ReadRune()
		if err != nil {
			return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
		}
		if ch == '"' {
			break
		}
		if ch != '\\' {
			if haveEscaped {
				chars = appendRune(chars, ch)
			}
			continue
		}
		if !haveEscaped {
			pos := (r.len - reader.Len()) - 1 // don't include the backslash we just read
			chars = make([]byte, pos-startPos, pos-startPos+20)
			if pos > startPos {
				copy(chars, r.data[startPos:pos])
			}
			haveEscaped = true
		}
		ch, _, err = reader.ReadRune()
		if err != nil {
			return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
		}
		switch ch {
		case '"', '\\', '/':
			chars = appendRune(chars, ch)
		case 'b':
			chars = appendRune(chars, '\b')
		case 'f':
			chars = appendRune(chars, '\f')
		case 'n':
			chars = appendRune(chars, '\n')
		case 'r':
			chars = appendRune(chars, '\r')
		case 't':
			chars = appendRune(chars, '\t')
		case 'u':
			if ch, ok := readHexChar(&reader); ok {
				chars = appendRune(chars, ch)
			} else {
				return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
			}
		default:
			return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
		}
	}
	r.pos = r.len - reader.Len()
	if haveEscaped {
		if len(chars) == 0 {
			return nil, nil
		}
		return chars, nil
	} else { //nolint:revive
		pos := r.pos - 1
		if pos <= startPos {
			return nil, nil
		}
		return r.data[startPos:pos], nil
	}
}

func readHexChar(reader *bytes.Reader) (rune, bool) {
	var digits [4]byte
	for i := 0; i < 4; i++ {
		ch, err := reader.ReadByte()
		if err != nil || !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return 0, false
		}
		digits[i] = ch
	}
	n, _ := strconv.ParseUint(string(digits[:]), 16, 32)
	return rune(n), true
}

func (r *tokenReader) syntaxErrorOnLastToken(msg string) error { //nolint:unparam
	return SyntaxError{Message: msg, Offset: r.LastPos()}
}

func (r *tokenReader) syntaxErrorOnNextToken(msg string) error {
	t, err := r.next()
	if err != nil {
		return err
	}
	return SyntaxError{Message: msg, Value: t.description(), Offset: r.LastPos()}
}

// This is faster than creating a string to pass to strconv.Atoi.
func parseIntFromBytes(chars []byte) (int64, bool) {
	negate := false
	p := 0
	var ret int64
	if len(chars) == 0 {
		return 0, false
	}
	if chars[0] == '-' {
		negate = true
		p++
		if p == len(chars) {
			return 0, false
		}
	}
	for p < len(chars) {
		ret = ret*10 + int64(chars[p]-'0')
		p++
	}
	if negate {
		ret = -ret
	}
	return ret, true
}

func appendRune(out []byte, ch rune) []byte {
	var encodedRune [10]byte
	n := utf8.EncodeRune(encodedRune[0:10], ch)
	return append(out, encodedRune[0:n]...)
}

func valueKindFromTokenKind(k tokenKind) ValueKind {
	switch k {
	case nullToken:
		return NullValue
	case boolToken:
		return BoolValue
	case numberToken:
		return NumberValue
	case stringToken:
		return StringValue
	}
	return -1
}
