package jreader

// NewReader creates a Reader that consumes the specified JSON input data.
//
// This function returns the struct by value (Reader, not *Reader). This avoids the overhead of a
// heap allocation since, in typical usage, the Reader will not escape the scope in which it was
// declared and can remain on the stack.
func NewReader(data []byte) Reader {
	return Reader{
		tr: newTokenReader(data),
	}
}
