package jreader

// ObjectState is returned by Reader's Object and ObjectOrNull methods. Use it in conjunction with
// Reader to iterate through a JSON object. To read the value of each object property, you will
// still use the Reader's methods. Properties may appear in any order.
//
// This example reads an object whose values are strings; if there is a null instead of an object,
// it behaves the same as for an empty object. Note that it is not necessary to check for an error
// result before iterating over the ObjectState, or to break out of the loop if String causes an
// error, because the ObjectState's Next method will return false if the Reader has had any errors.
//
//	values := map[string]string
//	for obj := r.ObjectOrNull(); obj.Next(); {
//	    key := string(obj.Name())
//	    if s := r.String(); r.Error() == nil {
//	        values[key] = s
//	    }
//	}
//
// The next example reads an object with two expected property names, "a" and "b". Any unrecognized
// properties are ignored.
//
//	var result struct {
//	    a int
//	    b int
//	}
//	for obj := r.ObjectOrNull(); obj.Next(); {
//	    switch string(obj.Name()) {
//	    case "a":
//	        result.a = r.Int()
//	    case "b":
//	        result.b = r.Int()
//	    }
//	}
//
// If the schema requires certain properties to always be present, the WithRequiredProperties method is
// a convenient way to enforce this.
type ObjectState struct {
	r                     *Reader
	afterFirst            bool
	name                  []byte
	requiredProps         []string
	requiredPropsFound    []bool
	requiredPropsPrealloc [20]bool // used as initial base array for requiredPropsFound to avoid allocation
}

// WithRequiredProperties adds a requirement that the specified JSON property name(s) must appear
// in the JSON object at some point before it ends.
//
// This method returns a new, modified ObjectState. It should be called before the first time you
// call Next. For instance:
//
//	requiredProps := []string{"key", "name"}
//	for obj := reader.Object().WithRequiredProperties(requiredProps); obj.Next(); {
//	    switch string(obj.Name()) { ... }
//	}
//
// When the end of the object is reached (and Next() returns false), if one of the required
// properties has not yet been seen, and no other error has occurred, the Reader's error state
// will be set to a RequiredPropertyError.
//
// For efficiency, it is best to preallocate the list of property names globally rather than creating
// it inline.
func (obj ObjectState) WithRequiredProperties(requiredProps []string) ObjectState {
	ret := obj
	if len(requiredProps) > 0 {
		ret.requiredProps = requiredProps
	}
	return ret
}

// IsDefined returns true if the ObjectState represents an actual object, or false if it was
// parsed from a null value or was the result of an error. If IsDefined is false, Next will
// always return false. The zero value ObjectState{} returns false for IsDefined.
func (obj *ObjectState) IsDefined() bool {
	return obj.r != nil
}

// Next checks whether an object property is available and returns true if so. It returns false
// if the Reader has reached the end of the object, or if any previous Reader operation failed,
// or if the object was empty or null.
//
// If Next returns true, you can then get the property name with Name, and use Reader methods
// such as Bool or String to read the property value. If you do not care about the value, simply
// calling Next again without calling a Reader method will discard the value, just as if you had
// called SkipValue on the reader.
//
// See ObjectState for example code.
func (obj *ObjectState) Next() bool {
	if obj.r == nil || obj.r.err != nil {
		return false
	}
	var isEnd bool
	var err error
	if !obj.afterFirst && len(obj.requiredProps) != 0 {
		// Initialize the bool slice that we'll use to keep track of what properties we found.
		// See comment on requiredPropsFoundSlice().
		if len(obj.requiredProps) > len(obj.requiredPropsPrealloc) {
			obj.requiredPropsFound = make([]bool, len(obj.requiredProps))
		}
	}

	if obj.afterFirst {
		if obj.r.awaitingReadValue {
			if err := obj.r.SkipValue(); err != nil {
				return false
			}
		}
		isEnd, err = obj.r.tr.EndDelimiterOrComma('}')
	} else {
		obj.afterFirst = true
		isEnd, err = obj.r.tr.Delimiter('}')
	}
	if err != nil {
		obj.r.AddError(err)
		return false
	}
	if isEnd {
		obj.name = nil
		if obj.requiredProps != nil {
			found := obj.requiredPropsFoundSlice()
			for i, requiredName := range obj.requiredProps {
				if !found[i] {
					obj.r.AddError(RequiredPropertyError{Name: requiredName, Offset: obj.r.tr.LastPos()})
					break
				}
			}
		}
		return false
	}
	name, err := obj.r.tr.PropertyName()
	if err != nil {
		obj.r.AddError(err)
		return false
	}
	obj.name = name
	obj.r.awaitingReadValue = true
	if obj.requiredProps != nil {
		found := obj.requiredPropsFoundSlice()
		for i, requiredName := range obj.requiredProps {
			if requiredName == string(name) {
				found[i] = true
				break
			}
		}
	}
	return true
}

// Name returns the name of the current object property, or nil if there is no current property
// (that is, if Next returned false or if Next was never called).
//
// For efficiency, to avoid allocating a string for each property name, the name is returned as a
// byte slice which may refer directly to the source data. Casting this to a string within a simple
// comparison expression or switch statement should not cause a string allocation; the Go compiler
// optimizes these into direct byte-slice comparisons.
func (obj *ObjectState) Name() []byte {
	return obj.name
}

// This technique of using either a preallocated fixed-length array or a slice (where we have
// only set the slice to a non-nil value if we determined that the array wasn't big enough) is a
// way to avoid unnecessary heap allocations: if the ObjectState is on the stack, the fixed-length
// array can stay on the stack too. In order for this to work, we *cannot* set the slice to refer
// to the array (obj.requiredProps = obj.requiredPropsFound[0:len(obj.requiredProps)]); the Go
// compiler can't prove that that's safe, so it will make everything escape to the heap. Instead
// we have to conditionally reference one or the other here.
func (obj *ObjectState) requiredPropsFoundSlice() []bool {
	if obj.requiredPropsFound != nil {
		return obj.requiredPropsFound
	}
	return obj.requiredPropsPrealloc[0:len(obj.requiredProps)]
}
