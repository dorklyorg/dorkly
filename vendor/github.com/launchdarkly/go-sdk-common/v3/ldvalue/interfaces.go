package ldvalue

// JSONStringer is a common interface defining the JSONString() method, a convenience for
// marshaling a value to JSON and getting the result as a string.
type JSONStringer interface {
	// JSONString returns the JSON representation of the value as a string. This is
	// guaranteed to be logically equivalent to calling json.Marshal() and converting
	// the first return value to a string.
	//
	// Since types that support this method are by definition always convertible to JSON,
	// it does not need to return an error value so it can be easily used within an
	// expression.
	JSONString() string
}
