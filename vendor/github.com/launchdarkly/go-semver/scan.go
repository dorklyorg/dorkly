package semver

import "unicode"

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

func (s *simpleASCIIScanner) eof() bool {
	return s.pos >= s.length
}

func (s *simpleASCIIScanner) peek() int8 {
	if s.pos >= s.length {
		return scannerEOF
	}
	var ch uint8 = s.source[s.pos]
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
