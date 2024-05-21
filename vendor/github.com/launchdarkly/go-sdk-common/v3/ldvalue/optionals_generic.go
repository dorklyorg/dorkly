package ldvalue

// This generic internal representation of optional values reduces the amount of boilerplate
// in the implementation of OptionalBool, OptionalInt, and OptionalString.

type optional[T any] struct {
	defined bool
	value   T
}

func newOptional[T any](value T) optional[T] {
	return optional[T]{defined: true, value: value}
}

func newOptionalFromPointer[T any](valuePointer *T) optional[T] {
	if valuePointer == nil {
		return optional[T]{}
	}
	return optional[T]{defined: true, value: *valuePointer}
}

func (o optional[T]) isDefined() bool {
	return o.defined
}

func (o optional[T]) getOrZeroValue() T {
	return o.value
}

func (o optional[T]) getOrElse(defaultValue T) T {
	if o.defined {
		return o.value
	}
	return defaultValue
}

func (o optional[T]) get() (T, bool) {
	return o.value, o.defined
}

func (o optional[T]) getAsPointer() *T {
	if !o.defined {
		return nil
	}
	return &o.value
}
