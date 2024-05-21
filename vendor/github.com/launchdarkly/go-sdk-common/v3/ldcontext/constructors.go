package ldcontext

// New creates a Context with a Kind of [DefaultKind] and the specified key.
//
// To specify additional properties, use [NewBuilder]. To create a multi-context, use
// [NewMulti] or [NewMultiBuilder]. To create a single Context of a different kind than
// [DefaultKind], use [NewWithKind]; New is simply a shortcut for calling NewWithKind(DefaultKind, key).
func New(key string) Context {
	return NewWithKind(DefaultKind, key)
}

// NewWithKind creates a Context with only the Kind and Key properties specified.
//
// To specify additional properties, use [NewBuilder]. To create a multi-context, use
// [NewMulti] or [NewMultiBuilder]. As a shortcut if the Kind is [DefaultKind], you can
// use [New].
func NewWithKind(kind Kind, key string) Context {
	// Here we'll use Builder rather than directly constructing the Context struct. That
	// allows us to take advantage of logic in Builder like the setting of FullyQualifiedKey.
	// We avoid the heap allocation overhead of NewBuilder by declaring a Builder locally.
	var b Builder
	b.Kind(kind)
	b.Key(key)
	return b.Build()
}

// NewMulti creates a multi-context out of the specified Contexts.
//
// To create a single [Context], use [New], [NewWithKind], or [NewBuilder].
//
// For the returned Context to be valid, the contexts list must not be empty, and all of its
// elements must be valid Contexts. Otherwise, the returned Context will be invalid as reported
// by [Context.Err].
//
// If only one context parameter is given, NewMulti returns that same context.
//
// If one of the nested contexts is a multi-context, this is exactly equivalent to adding each
// of the individual contexts from it separately. For instance, in the following example,
// "multi1" and "multi2" end up being exactly the same:
//
//	c1 := ldcontext.NewWithKind("kind1", "key1")
//	c2 := ldcontext.NewWithKind("kind2", "key2")
//	c3 := ldcontext.NewWithKind("kind3", "key3")
//
//	multi1 := ldcontext.NewMulti(c1, c2, c3)
//
//	c1plus2 := ldcontext.NewMulti(c1, c2)
//	multi2 := ldcontext.NewMulti(c1plus2, c3)
func NewMulti(contexts ...Context) Context {
	// Same rationale as for New/NewWithKey of using the builder instead of constructing directly.
	var m MultiBuilder
	for _, c := range contexts {
		m.Add(c)
	}
	return m.Build()
}
