// Package jreader provides an efficient mechanism for reading JSON data sequentially.
//
// The high-level API for this package, Writer, is designed to facilitate writing custom JSON
// marshaling logic concisely and reliably. Output is buffered in memory.
//
//	import (
//	    "gopkg.in/launchdarkly/jsonstream.v1/jreader"
//	)
//
//	type myStruct struct {
//	    value int
//	}
//
//	func (s *myStruct) ReadFromJSONReader(r *jreader.Reader) {
//	    // reading a JSON object structure like {"value":2}
//	    for obj := r.Object(); obj.Next; {
//	        if string(obj.Name()) == "value" {
//	            s.value = r.Int()
//	        }
//	    }
//	}
//
//	func ParseMyStructJSON() {
//	    var s myStruct
//	    r := jreader.NewReader([]byte(`{"value":2}`))
//	    s.ReadFromJSONReader(&r)
//	    fmt.Printf("%+v\n", s)
//	}
//
// The underlying low-level token parsing mechanism has two available implementations. The default
// implementation has no external dependencies. For interoperability with the easyjson library
// (https://github.com/mailru/easyjson), there is also an implementation that delegates to the
// easyjson streaming parser; this is enabled by setting the build tag "launchdarkly_easyjson".
// Be aware that by default, easyjson uses Go's "unsafe" package (https://pkg.go.dev/unsafe),
// which may not be available on all platforms.
//
// Setting the "launchdarkly_easyjson" tag also adds a new constructor function,
// NewReaderFromEasyJSONLexer, allowing Reader-based code to read directly from an existing
// EasyJSON jlexer.Lexer. This may be desirable in order to define common unmarshaling logic that
// may be used with or without EasyJSON. For example:
//
//	import (
//	    "github.com/mailru/easyjson/jlexer"
//	)
//
//	func (s *myStruct) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
//	    r := jreader.NewReaderFromEasyJSONLexer(lexer)
//	    s.ReadFromJSONReader(&r)
//	}
package jreader
