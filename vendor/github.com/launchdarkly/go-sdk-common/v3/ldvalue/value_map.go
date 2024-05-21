package ldvalue

import (
	"golang.org/x/exp/maps"
)

// we reuse this for all non-nil zero-length ValueMap instances
var emptyMap = map[string]Value{} //nolint:gochecknoglobals

// ValueMap is an immutable map of string keys to Value values.
//
// This is used internally to hold the properties of a JSON object in a Value. You can also use it
// separately in any context where you know that the data must be map-like, rather than any of the
// other types that a Value can be.
//
// The wrapped map is not directly accessible, so it cannot be modified. You can obtain a copy of
// it with AsMap() if necessary.
//
// Like a Go map, there is a distinction between a map in a nil state-- which is the zero value of
// ValueMap{}-- and a non-nil map that is empty. The former is represented in JSON as a null; the
// latter is an empty JSON object {}.
type ValueMap struct {
	data map[string]Value
}

// ValueMapBuilder is a builder created by ValueMapBuild(), for creating immutable JSON objects.
//
// A ValueMapBuilder should not be accessed by multiple goroutines at once.
type ValueMapBuilder struct {
	copyOnWrite bool
	output      map[string]Value
}

// Set sets a key-value pair in the map builder.
func (b *ValueMapBuilder) Set(key string, value Value) *ValueMapBuilder {
	if b == nil {
		return b
	}
	if b.copyOnWrite {
		b.output = maps.Clone(b.output)
		b.copyOnWrite = false
	}
	if b.output == nil {
		b.output = make(map[string]Value, 1)
	}
	b.output[key] = value
	return b
}

// SetAllFromValueMap copies all key-value pairs from an existing ValueMap.
func (b *ValueMapBuilder) SetAllFromValueMap(m ValueMap) *ValueMapBuilder {
	if b == nil {
		return b
	}
	if b.output == nil {
		b.output = m.data
		b.copyOnWrite = true
	} else {
		for k, v := range m.data {
			b.Set(k, v)
		}
	}
	return b
}

// Remove unsets a key if it was set.
func (b *ValueMapBuilder) Remove(key string) *ValueMapBuilder {
	if b == nil {
		return b
	}
	if b.output != nil {
		if b.copyOnWrite {
			b.output = maps.Clone(b.output)
			b.copyOnWrite = false
		}
		delete(b.output, key)
	}
	return b
}

// HasKey returns true if the specified key has been set in the builder.
func (b *ValueMapBuilder) HasKey(key string) bool {
	_, found := b.output[key]
	return found
}

// Build creates a ValueMap containing the previously specified key-value pairs. Continuing to
// modify the same builder by calling Set after that point does not affect the returned ValueMap.
func (b *ValueMapBuilder) Build() ValueMap {
	if b == nil {
		return ValueMap{}
	}
	if b.output == nil {
		return ValueMap{emptyMap}
	}
	b.copyOnWrite = true
	return ValueMap{b.output}
}

// ValueMapBuild creates a builder for constructing an immutable ValueMap.
//
//	valueMap := ldvalue.ValueMapBuild().Set("a", ldvalue.Int(100)).Set("b", ldvalue.Int(200)).Build()
func ValueMapBuild() *ValueMapBuilder {
	return &ValueMapBuilder{}
}

// ValueMapBuildWithCapacity creates a builder for constructing an immutable ValueMap.
//
// The capacity parameter is the same as the capacity of a map, allowing you to preallocate space
// if you know the number of elements; otherwise you can pass zero.
//
//	objValue := ldvalue.ObjectBuildWithCapacity(2).Set("a", ldvalue.Int(100)).Set("b", ldvalue.Int(200)).Build()
func ValueMapBuildWithCapacity(capacity int) *ValueMapBuilder {
	return &ValueMapBuilder{output: make(map[string]Value, capacity)}
}

// ValueMapBuildFromMap creates a builder for constructing an immutable ValueMap, initializing it
// from an existing ValueMap.
//
// The builder has copy-on-write behavior, so if you make no changes before calling Build(), the
// original map is used as-is.
func ValueMapBuildFromMap(m ValueMap) *ValueMapBuilder {
	return &ValueMapBuilder{output: m.data, copyOnWrite: true}
}

// CopyValueMap copies an existing ordinary map to a ValueMap.
//
// If the parameter is nil, an uninitialized ValueMap{} is returned instead of a zero-length map.
func CopyValueMap(data map[string]Value) ValueMap {
	if data == nil {
		return ValueMap{}
	}
	if len(data) == 0 {
		return ValueMap{emptyMap}
	}
	return ValueMap{maps.Clone(data)}
}

// CopyArbitraryValueMap copies an existing ordinary map of values of any type to a ValueMap. The
// behavior for each value is the same as CopyArbitraryValue.
//
// If the parameter is nil, an uninitialized ValueMap{} is returned instead of a zero-length map.
func CopyArbitraryValueMap(data map[string]any) ValueMap {
	if data == nil {
		return ValueMap{}
	}
	m := make(map[string]Value, len(data))
	for k, v := range data {
		m[k] = CopyArbitraryValue(v)
	}
	return ValueMap{data: m}
}

// IsDefined returns true if the map is non-nil.
func (m ValueMap) IsDefined() bool {
	return m.data != nil
}

// Count returns the number of keys in the map. For an uninitialized ValueMap{}, this is zero.
func (m ValueMap) Count() int {
	return len(m.data)
}

// AsValue converts the ValueMap to a Value which is either Null() or an object. This does not
// cause any new allocations.
func (m ValueMap) AsValue() Value {
	if m.data == nil {
		return Null()
	}
	return Value{valueType: ObjectType, objectValue: m}
}

// Get gets a value from the map by key.
//
// If the key is not found, it returns Null().
func (m ValueMap) Get(key string) Value {
	return m.data[key]
}

// TryGet gets a value from the map by key, with a second return value of true if successful.
//
// If the key is not found, it returns (Null(), false).
func (m ValueMap) TryGet(key string) (Value, bool) {
	ret, ok := m.data[key]
	return ret, ok
}

// Keys returns the keys of a the map as a slice.
//
// If a non-nil slice is passed in, it will be reused to hold the return values if it has enough capacity.
// Otherwise, a new slice is allocated if there are any keys.
//
// The ordering of the keys is undefined.
func (m ValueMap) Keys(sliceIn []string) []string {
	if len(m.data) == 0 {
		return sliceIn
	}
	ret := sliceIn[0:0]
	for key := range m.data {
		ret = append(ret, key)
	}
	return ret
}

// AsMap returns a copy of the wrapped data as a simple Go map whose values are of type Value.
//
// For an uninitialized ValueMap{}, this returns nil.
func (m ValueMap) AsMap() map[string]Value {
	return maps.Clone(m.data)
}

// AsArbitraryValueMap returns a copy of the wrapped data as a simple Go map whose values are of any
// type. The behavior for each value is the same as Value.AsArbitraryValue().
//
// For an uninitialized ValueMap{}, this returns nil.
func (m ValueMap) AsArbitraryValueMap() map[string]any {
	if m.data == nil {
		return nil
	}
	ret := make(map[string]any, len(m.data))
	for k, v := range m.data {
		ret[k] = v.AsArbitraryValue()
	}
	return ret
}

// Equal returns true if the two maps are deeply equal. Nil and zero-length maps are not considered
// equal to each other.
func (m ValueMap) Equal(other ValueMap) bool {
	if m.IsDefined() != other.IsDefined() {
		return false
	}
	return maps.EqualFunc(m.data, other.data, Value.Equal)
}

// Transform applies a transformation function to a ValueMap, returning a new ValueMap.
//
// The behavior is as follows:
//
// If the input value is nil or zero-length, the result is identical and the function is not called.
//
// Otherwise, fn is called for each key-value pair. It should return a transformed key-value pair
// and true, or else return false for the third return value if the property should be dropped.
func (m ValueMap) Transform(fn func(key string, value Value) (string, Value, bool)) ValueMap {
	if len(m.data) == 0 {
		return m
	}
	ret := m.data
	startedNewMap := false
	seenKeys := make([]string, 0, len(m.data))
	for k, v := range m.data {
		resultKey, resultValue, ok := fn(k, v)
		modified := !ok || resultKey != k || !resultValue.Equal(v)
		if modified && !startedNewMap {
			// This is the first change we've seen, so we should start building a new map and
			// retroactively add any values to it that already passed the test without changes.
			startedNewMap = true
			ret = make(map[string]Value, len(m.data))
			for _, seenKey := range seenKeys {
				ret[seenKey] = m.data[seenKey]
			}
		} else {
			seenKeys = append(seenKeys, k)
		}
		if startedNewMap && ok {
			ret[k] = resultValue
		}
	}
	return ValueMap{ret}
}
