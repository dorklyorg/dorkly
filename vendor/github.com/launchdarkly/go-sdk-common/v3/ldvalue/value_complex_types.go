package ldvalue

// This file contains types and methods that are only used for complex data structures (array and
// object), in the fully immutable model where no slices, maps, or interface{} values are exposed.

// ArrayBuilder is a builder created by [ArrayBuild], for creating immutable JSON arrays.
//
// An ArrayBuilder should not be accessed by multiple goroutines at once.
//
// Other ways to create a JSON array include:
//
//   - Specifying all of its elements at once with [ArrayOf].
//   - Copying it from a Go slice with [CopyArbitraryValue].
//   - Creating it from the JSON serialization of an arbitrary type with [FromJSONMarshal].
//   - Parsing it from JSON data with [Parse].
type ArrayBuilder struct {
	builder ValueArrayBuilder
}

// ObjectBuilder is a builder created by [ObjectBuild], for creating immutable JSON objects.
//
// An ObjectBuilder should not be accessed by multiple goroutines at once.
//
// Other ways to create a JSON object include:
//
//   - Copying it from a Go map with [CopyArbitraryValue].
//   - Creating it from the JSON serialization of an arbitrary type with [FromJSONMarshal].
//   - Parsing it from JSON data with [Parse].
type ObjectBuilder struct {
	builder ValueMapBuilder
}

// Add appends an element to the array builder.
func (b *ArrayBuilder) Add(value Value) *ArrayBuilder {
	if b != nil {
		b.builder.Add(value)
	}
	return b
}

// Build creates a Value containing the previously added array elements. Continuing to modify the
// same builder by calling [ArrayBuilder.Add] after that point does not affect the returned array.
func (b *ArrayBuilder) Build() Value {
	if b == nil {
		return Null()
	}
	return Value{valueType: ArrayType, arrayValue: b.builder.Build()}
}

// ArrayOf creates an array Value from a list of Values.
//
// This requires a slice copy to ensure immutability; otherwise, an existing slice could be passed
// using the spread operator, and then modified. However, since Value is itself immutable, it does
// not need to deep-copy each item.
func ArrayOf(items ...Value) Value {
	return Value{valueType: ArrayType, arrayValue: ValueArrayOf(items...)}
}

// ArrayBuild creates a builder for constructing an immutable array [Value].
//
//	arrayValue := ldvalue.ArrayBuild().Add(ldvalue.Int(100)).Add(ldvalue.Int(200)).Build()
func ArrayBuild() *ArrayBuilder {
	return ArrayBuildWithCapacity(1)
}

// ArrayBuildWithCapacity creates a builder for constructing an immutable array [Value]. This is
// the same as [ArrayBuild], but preallocates capacity for the underlying slice.
//
//	arrayValue := ldvalue.ArrayBuildWithCapacity(2).Add(ldvalue.Int(100)).Add(ldvalue.Int(200)).Build()
func ArrayBuildWithCapacity(capacity int) *ArrayBuilder {
	return &ArrayBuilder{ValueArrayBuilder{output: make([]Value, 0, capacity)}}
}

// CopyObject creates a Value by copying an existing map[string]Value.
//
// If you want to copy a map[string]interface{} instead, use [CopyArbitraryValue].
func CopyObject(m map[string]Value) Value {
	return Value{valueType: ObjectType, objectValue: CopyValueMap(m)}
}

// ObjectBuild creates a builder for constructing an immutable JSON object [Value].
//
//	objValue := ldvalue.ObjectBuild().Set("a", ldvalue.Int(100)).Set("b", ldvalue.Int(200)).Build()
func ObjectBuild() *ObjectBuilder {
	return ObjectBuildWithCapacity(1)
}

// ObjectBuildWithCapacity creates a builder for constructing an immutable JSON object [Value].
// This is the same as [ObjectBuild], but preallocates capacity for the underlying map.
//
//	objValue := ldvalue.ObjectBuildWithCapacity(2).Set("a", ldvalue.Int(100)).Set("b", ldvalue.Int(200)).Build()
func ObjectBuildWithCapacity(capacity int) *ObjectBuilder {
	return &ObjectBuilder{ValueMapBuilder{output: make(map[string]Value, capacity)}}
}

// Set sets a key-value pair in the object builder to a value of any JSON type.
func (b *ObjectBuilder) Set(key string, value Value) *ObjectBuilder {
	if b != nil {
		b.builder.Set(key, value)
	}
	return b
}

// SetBool sets a key-value pair in the object builder to a boolean value.
//
// This is exactly equivalent to b.Set(key, ldvalue.Bool(value)).
func (b *ObjectBuilder) SetBool(key string, value bool) *ObjectBuilder {
	return b.Set(key, Bool(value))
}

// SetInt sets a key-value pair in the object builder to an integer numeric value.
//
// This is exactly equivalent to b.Set(key, ldvalue.Int(value)).
func (b *ObjectBuilder) SetInt(key string, value int) *ObjectBuilder {
	return b.Set(key, Int(value))
}

// SetFloat64 sets a key-value pair in the object builder to a float64 numeric value.
//
// This is exactly equivalent to b.Set(key, ldvalue.Float64(value)).
func (b *ObjectBuilder) SetFloat64(key string, value float64) *ObjectBuilder {
	return b.Set(key, Float64(value))
}

// SetString sets a key-value pair in the object builder to a string value.
//
// This is exactly equivalent to b.Set(key, ldvalue.String(value)).
func (b *ObjectBuilder) SetString(key string, value string) *ObjectBuilder {
	return b.Set(key, String(value))
}

// Remove removes a key from the builder if it exists.
func (b *ObjectBuilder) Remove(key string) *ObjectBuilder {
	if b != nil {
		b.builder.Remove(key)
	}
	return b
}

// Build creates a Value containing the previously specified key-value pairs. Continuing to modify
// the same builder by calling [ObjectBuilder.Set] after that point does not affect the returned object.
func (b *ObjectBuilder) Build() Value {
	if b == nil {
		return Null()
	}
	return Value{valueType: ObjectType, objectValue: b.builder.Build()}
}

// Count returns the number of elements in an array or JSON object.
//
// For values of any other type, it returns zero.
//
// If the value is a JSON array or object created from unparsed JSON with [Raw], this method
// first parses the JSON, which can be inefficient.
func (v Value) Count() int {
	switch v.valueType {
	case ArrayType:
		return v.arrayValue.Count()
	case ObjectType:
		return v.objectValue.Count()
	case RawType:
		return v.parseIfRaw().Count()
	}
	return 0
}

// GetByIndex gets an element of an array by index.
//
// If the value is not an array, or if the index is out of range, it returns [Null]().
//
// If the value is a JSON array or object created from unparsed JSON with [Raw], this method
// first parses the JSON, which can be inefficient.
func (v Value) GetByIndex(index int) Value {
	ret, _ := v.TryGetByIndex(index)
	return ret
}

// TryGetByIndex gets an element of an array by index, with a second return value of true if
// successful.
//
// If the value is not an array, or if the index is out of range, it returns ([Null](), false).
//
// If the value is a JSON array or object created from unparsed JSON with [Raw], this method
// first parses the JSON, which can be inefficient.
func (v Value) TryGetByIndex(index int) (Value, bool) {
	if v.valueType == RawType {
		return v.parseIfRaw().TryGetByIndex(index)
	}
	return v.arrayValue.TryGet(index)
	// This is always safe because if v isn't an array, arrayValue is an empty ValueArray{}
	// and TryGet will always return Null(), false.
}

// Keys returns the keys of a JSON object as a slice.
//
// If a non-nil slice is passed in, it will be reused to hold the return values if it has enough capacity.
// Otherwise, a new slice is allocated if there are any keys.
//
// The ordering of the keys is undefined.
//
// If the value is a JSON array or object created from unparsed JSON with [Raw], this method
// first parses the JSON, which can be inefficient.
func (v Value) Keys(sliceIn []string) []string {
	if v.Type() == RawType {
		return Parse(v.rawValue).Keys(sliceIn)
	}
	if v.valueType == ObjectType {
		return v.objectValue.Keys(sliceIn)
	}
	return nil
}

// GetByKey gets a value from a JSON object by key.
//
// If the value is not an object, or if the key is not found, it returns [Null]().
//
// If the value is a JSON array or object created from unparsed JSON with [Raw], this method
// first parses the JSON, which can be inefficient.
func (v Value) GetByKey(name string) Value {
	ret, _ := v.TryGetByKey(name)
	return ret
	// This is always safe because if v isn't an object, objectValue is an empty ValueMap{}
	// and keys will never be found.
}

// TryGetByKey gets a value from a JSON object by key, with a second return value of true if
// successful.
//
// If the value is not an object, or if the key is not found, it returns ([Null](), false).
//
// If the value is a JSON array or object created from unparsed JSON with [Raw], this method
// first parses the JSON, which can be inefficient.
func (v Value) TryGetByKey(name string) (Value, bool) {
	if v.Type() == RawType {
		return Parse(v.rawValue).TryGetByKey(name)
	}
	return v.objectValue.TryGet(name)
}

// Transform applies a transformation function to a Value, returning a new Value.
//
// The behavior is as follows:
//
// If the input value is [Null](), the return value is always Null() and the function is not called.
//
// If the input value is an array, fn is called for each element, with the element's index in the
// first parameter, "" in the second, and the element value in the third. The return values of fn
// can be either a transformed value and true, or any value and false to remove the element.
//
//	ldvalue.ArrayOf(ldvalue.Int(1), ldvalue.Int(2), ldvalue.Int(3)).Build().
//	    Transform(func(index int, key string, value Value) (Value, bool) {
//	        if value.IntValue() == 2 {
//	            return ldvalue.Null(), false
//	        }
//	        return ldvalue.Int(value.IntValue() * 10), true
//	    })
//	// returns [10, 30]
//
// If the input value is an object, fn is called for each key-value pair, with 0 in the first
// parameter, the key in the second, and the value in the third. Again, fn can choose to either
// transform or drop the value.
//
//	ldvalue.ObjectBuild().Set("a", ldvalue.Int(1)).Set("b", ldvalue.Int(2)).Set("c", ldvalue.Int(3)).Build().
//	    Transform(func(index int, key string, value Value) (Value, bool) {
//	        if key == "b" {
//	            return ldvalue.Null(), false
//	        }
//	        return ldvalue.Int(value.IntValue() * 10), true
//	    })
//	// returns {"a": 10, "c": 30}
//
// For any other value type, fn is called once for that value; if it provides a transformed
// value and true, the transformed value is returned, otherwise Null().
//
//	ldvalue.Int(2).Transform(func(index int, key string, value Value) (Value, bool) {
//	    return ldvalue.Int(value.IntValue() * 10), true
//	})
//	// returns numeric value of 20
//
// For array and object values, if the function does not modify or drop any values, the exact
// same instance is returned without allocating a new slice or map.
//
// If the value is a JSON array or object created from unparsed JSON with [Raw], this method
// first parses the JSON, which can be inefficient.
func (v Value) Transform(fn func(index int, key string, value Value) (Value, bool)) Value {
	switch v.valueType {
	case NullType:
		return v
	case ArrayType:
		return Value{valueType: ArrayType, arrayValue: v.arrayValue.Transform(
			func(index int, value Value) (Value, bool) {
				return fn(index, "", value)
			},
		)}
	case ObjectType:
		return Value{valueType: ObjectType, objectValue: v.objectValue.Transform(
			func(key string, value Value) (string, Value, bool) {
				resultValue, ok := fn(0, key, value)
				return key, resultValue, ok
			},
		)}
	case RawType:
		return v.parseIfRaw().Transform(fn)
	default:
		if transformedValue, ok := fn(0, "", v); ok {
			return transformedValue
		}
		return Null()
	}
}

// AsValueArray converts the Value to the immutable ValueArray type if it is a JSON array.
// Otherwise it returns a ValueArray in an uninitialized (nil) state. This is an efficient operation
// that does not allocate a new slice.
//
// If the value is a JSON array or object created from unparsed JSON with [Raw], this method
// first parses the JSON, which can be inefficient.
func (v Value) AsValueArray() ValueArray {
	return v.parseIfRaw().arrayValue
}

// AsValueMap converts the Value to the immutable ValueMap type if it is a JSON object. Otherwise
// it returns a ValueMap in an uninitialized (nil) state. This is an efficient operation that does
// not allocate a new map.
//
// If the value is a JSON array or object created from unparsed JSON with [Raw], this method
// first parses the JSON, which can be inefficient.
func (v Value) AsValueMap() ValueMap {
	return v.parseIfRaw().objectValue
}
