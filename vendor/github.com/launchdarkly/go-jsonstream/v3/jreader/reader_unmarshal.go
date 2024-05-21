package jreader

// UnmarshalJSONWithReader is a convenience method for implementing json.Marshaler to unmarshal from
// a byte slice with the default TokenReader implementation. If an error occurs, it is converted to
// the corresponding error type defined by the encoding/json package when applicable.
//
// This method will generally be less efficient than writing the exact same logic inline in a custom
// UnmarshalJSON method for the object's specific type, because casting a pointer to an interface
// (Readable) will force the object to be allocated on the heap if it was not already.
func UnmarshalJSONWithReader(data []byte, readable Readable) error {
	r := NewReader(data)
	readable.ReadFromJSONReader(&r)
	if err := r.Error(); err != nil {
		return ToJSONError(err, readable)
	}
	return r.RequireEOF()
}
