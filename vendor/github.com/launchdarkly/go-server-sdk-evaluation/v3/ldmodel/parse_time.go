package ldmodel

import (
	"time"
	"unicode"
)

// A fast, zero-heap-allocations implementation of RFC3339 date/time parsing
// (https://tools.ietf.org/html/rfc3339). The returned Time always has UTC as its Location -
// it still respects any time zone specifier but simply adds that offset to the UTC time.
//
// The following deviations from the RFC3339 spec are intentional, in order to be consistent
// with the behavior of Go's time.Parse() with the time.RFC3339 format. They are marked in
// the code with "// NONSTANDARD".
//
// - There cannot be more than 9 digits of fractional seconds (the spec does not define any
// maximum length).
//
// - In a time zone offset of -hh:mm or +hh:mm, hh must be an integer but has no maximum
// (the spec defines a maximum of 23).
//
// - The hour can be either 1 digit or 2 digits (the spec requires 2).
func parseRFC3339TimeUTC(s string) (time.Time, bool) {
	scanner := newSimpleASCIIScanner(s)

	year, _, ok := parseDateTimeNumericField(&scanner, hyphenTerminator, false, 4, 4, 0, 9999)
	if !ok {
		return time.Time{}, false
	}
	month, _, ok := parseDateTimeNumericField(&scanner, hyphenTerminator, false, 2, 2, 1, 12)
	if !ok {
		return time.Time{}, false
	}
	day, _, ok := parseDateTimeNumericField(&scanner, tTerminator, false, 2, 2, 1, 31)
	if !ok {
		return time.Time{}, false
	}
	hour, _, ok := parseDateTimeNumericField(&scanner, colonTerminator, false, 1, 2, 0, 23)
	// NONSTANDARD: time.Parse allows 1-digit hour
	if !ok {
		return time.Time{}, false
	}
	minute, _, ok := parseDateTimeNumericField(&scanner, colonTerminator, false, 2, 2, 0, 59)
	if !ok {
		return time.Time{}, false
	}
	second, term, ok := parseDateTimeNumericField(&scanner, endOfSecondsTerminator, false, 2, 2, 0, 60)
	// note that second can be 60 sometimes due to leap seconds
	if !ok {
		return time.Time{}, false
	}

	var nanos int
	if term == '.' {
		var fractionStr string
		fractionStr, term = scanner.readUntil(endOfFractionalSecondsTerminator)
		if term < 0 || len(fractionStr) > 9 {
			// NONSTANDARD: time.Parse does not support more than 9 fractional digits
			return time.Time{}, false
		}
		n, ok := parsePositiveNumericString(fractionStr)
		if !ok {
			return time.Time{}, false
		}
		nanos = n
		for i := len(fractionStr); i < 9; i++ {
			nanos *= 10
		}
	}

	var tzOffsetSeconds int
	if term == '+' || term == '-' {
		offsetHours, _, ok := parseDateTimeNumericField(&scanner, colonTerminator, false, 2, 2, 0, 99)
		// NONSTANDARD: time.Parse imposes no maximum on the hour field, just a 2-digit length
		if !ok {
			return time.Time{}, false
		}
		offsetMinutes, _, ok := parseDateTimeNumericField(&scanner, noTerminator, true, 2, 2, 0, 59)
		if !ok {
			return time.Time{}, false
		}
		tzOffsetSeconds = (offsetMinutes + (offsetHours * 60)) * 60
		if term == '+' {
			tzOffsetSeconds = -tzOffsetSeconds
		}
	}

	t := time.Date(year, time.Month(month), day, hour, minute, second, nanos, time.UTC)
	if tzOffsetSeconds != 0 {
		t = t.Add(time.Second * time.Duration(tzOffsetSeconds))
	}

	return t, true
}

func hyphenTerminator(ch rune) bool { return ch == '-' }
func tTerminator(ch rune) bool      { return ch == 't' || ch == 'T' }
func colonTerminator(ch rune) bool  { return ch == ':' }
func endOfSecondsTerminator(ch rune) bool {
	return ch == '.' || ch == 'Z' || ch == 'z' || ch == '+' || ch == '-'
}
func endOfFractionalSecondsTerminator(ch rune) bool {
	return ch == 'Z' || ch == 'z' || ch == '+' || ch == '-'
}

func parseDateTimeNumericField(
	scanner *simpleASCIIScanner,
	terminatorFn func(rune) bool,
	eofOK bool,
	minLength, maxLength, minValue, maxValue int,
) (int, int8, bool) {
	s, term := scanner.readUntil(terminatorFn)
	if s == "" || (!eofOK && term < 0) {
		return 0, term, false
	}
	length := len(s)
	if length < minLength || length > maxLength {
		return 0, term, false
	}
	n, ok := parsePositiveNumericString(s)
	if !ok || n < minValue || n > maxValue {
		return 0, term, false
	}
	return n, term, true
}

// Attempts to parse a string as an integer greater than or equal to zero. Non-ASCII strings are not supported.
func parsePositiveNumericString(s string) (int, bool) {
	max := len(s)
	if max == 0 {
		return 0, false
	}
	n := 0
	for i := 0; i < max; i++ {
		ch := rune(s[i])
		if ch < '0' || ch > '9' {
			return 0, false
		}
		n = n*10 + int(ch-'0')
	}
	return n, true
}

// An extremely simple tokenizing helper that only handles ASCII strings.

type simpleASCIIScanner struct {
	source string
	length int
	pos    int
}

const (
	scannerEOF      int8 = -1
	scannerNonASCII int8 = -2
)

func noTerminator(rune) bool {
	return false
}

func newSimpleASCIIScanner(source string) simpleASCIIScanner {
	return simpleASCIIScanner{source: source, length: len(source)}
}

func (s *simpleASCIIScanner) peek() int8 {
	if s.pos >= s.length {
		return scannerEOF
	}
	var ch uint8 = s.source[s.pos] //nolint:stylecheck
	if ch == 0 || ch > unicode.MaxASCII {
		return scannerNonASCII
	}
	return int8(ch)
}

func (s *simpleASCIIScanner) next() int8 {
	ch := s.peek()
	if ch > 0 {
		s.pos++
	}
	return ch
}

func (s *simpleASCIIScanner) readUntil(terminatorFn func(rune) bool) (substring string, terminatedBy int8) {
	startPos := s.pos
	var ch int8
	for {
		ch = s.next()
		if ch < 0 || terminatorFn(rune(ch)) {
			break
		}
	}
	endPos := s.pos
	if ch > 0 {
		endPos--
	}
	return s.source[startPos:endPos], ch
}
