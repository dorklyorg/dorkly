package ldcontext

import (
	"sort"

	"github.com/launchdarkly/go-sdk-common/v3/lderrors"

	"golang.org/x/exp/slices"
)

const defaultMultiBuilderCapacity = 3 // arbitrary value based on presumed likely use cases

// MultiBuilder is a mutable object that uses the builder pattern to create a multi-context,
// as an alternative to [NewMulti].
//
// Use this type if you need to construct a Context that has multiple Kind values, each with its
// own nested [Context]. To define a single context, use [Builder] instead.
//
// Obtain an instance of MultiBuilder by calling [NewMultiBuilder]; then, call [MultiBuilder.Add] to
// specify the nested [Context] for each Kind. Finally, call [MultiBuilder.Build]. MultiBuilder
// setters return a reference to the same builder, so they can be chained together:
//
//	context := ldcontext.NewMultiBuilder().
//		Add(ldcontext.New("my-user-key")).
//		Add(ldcontext.NewBuilder("my-org-key").Kind("organization").Name("Org1").Build()).
//		Build()
//
// A MultiBuilder should not be accessed by multiple goroutines at once. Once you have called
// [MultiBuilder.Build], the resulting Context is immutable and safe to use from multiple
// goroutines.
type MultiBuilder struct {
	contexts            []Context
	contextsCopyOnWrite bool
}

// NewMultiBuilder creates a MultiBuilder for building a multi-context.
//
// This method is for building a [Context] that has multiple [Context.Kind] values, each with its
// own nested Context. To define a single context, use [NewBuilder] instead.
func NewMultiBuilder() *MultiBuilder {
	return &MultiBuilder{contexts: make([]Context, 0, defaultMultiBuilderCapacity)}
}

// Build creates a Context from the current MultiBuilder properties.
//
// The [Context] is immutable and will not be affected by any subsequent actions on the MultiBuilder.
//
// It is possible for a MultiBuilder to represent an invalid state. Instead of returning two
// values (Context, error), the Builder always returns a Context and you can call Context.Err()
// to see if it has an error. See [Context.Err] for more information about invalid Context
// conditions. Using a single-return-value syntax is more convenient for application code, since
// in normal usage an application will never build an invalid Context.
//
// If only one context was added to the builder, Build returns that same context, rather than a
// multi-context-- since there is no logical difference in LaunchDarkly between a single context and
// a multi-context that only contains one context.
func (m *MultiBuilder) Build() Context {
	if len(m.contexts) == 0 {
		return Context{defined: true, err: lderrors.ErrContextKindMultiWithNoKinds{}}
	}

	if len(m.contexts) == 1 {
		// If only one context was added, the result is just the same as that one
		return m.contexts[0]
	}

	m.contextsCopyOnWrite = true // see note on ___CopyOnWrite in Builder.Build()

	// Sort the list by kind - this makes our output deterministic and will also be important when we
	// compute a fully qualified key.
	sort.Slice(m.contexts, func(i, j int) bool { return m.contexts[i].Kind() < m.contexts[j].Kind() })

	// Check for conditions that could make a multi-context invalid
	var individualErrors map[string]error
	duplicates := false
	for i, c := range m.contexts {
		err := c.Err()
		switch {
		case err != nil: // one of the individual contexts already had an error
			if individualErrors == nil {
				individualErrors = make(map[string]error)
			}
			individualErrors[string(c.Kind())] = err
		default:
			// duplicate check's correctness relies on m.contexts being sorted by kind.
			if i > 0 && m.contexts[i-1].Kind() == c.Kind() {
				duplicates = true
			}
		}
	}
	var err error
	switch {
	case duplicates:
		err = lderrors.ErrContextKindMultiDuplicates{}
	case len(individualErrors) != 0:
		err = lderrors.ErrContextPerKindErrors{Errors: individualErrors}
	}
	if err != nil {
		return Context{
			defined: true,
			err:     err,
		}
	}

	ret := Context{
		defined:       true,
		kind:          MultiKind,
		multiContexts: m.contexts,
	}

	// Fully-qualified key for multi-context is defined as "kind1:key1:kind2:key2" etc., where kinds are in
	// alphabetical order (we have already sorted them above) and keys are URL-encoded. In this case we
	// do _not_ omit a default kind of "user".
	for _, c := range m.contexts {
		if ret.fullyQualifiedKey != "" {
			ret.fullyQualifiedKey += ":"
		}
		ret.fullyQualifiedKey += makeFullyQualifiedKeySingleKind(c.kind, c.key, false)
	}

	return ret
}

// TryBuild is an alternative to Build that returns any validation errors as a second value.
//
// As described in [MultiBuilder.Build], there are several ways the state of a [Context] could
// be invalid. Since in normal usage it is possible to be confident that these will not occur,
// the Build method is designed for convenient use within expressions by returning a single
// Context value, and any validation problems are contained within that value where they can be
// detected by calling the context's [Context.Err] method. But, if you prefer to use the
// two-value pattern that is common in Go, you can call TryBuild instead:
//
//	c, err := ldcontext.NewMultiBuilder().
//		Add(context1).Add(context2).
//		TryBuild()
//	if err != nil {
//		// do whatever is appropriate if building the context failed
//	}
//
// The two return values are the same as to 1. the Context that would be returned by Build(),
// and 2. the result of calling [Context.Err] on that Context. So, the above example is exactly
// equivalent to:
//
//	c := ldcontext.NewMultiBuilder().
//		Add(context1).Add(context2).
//		Build()
//	if c.Err() != nil {
//		// do whatever is appropriate if building the context failed
//	}
//
// Note that unlike some Go methods where the first return value is normally an
// uninitialized zero value if the error is non-nil, the Context returned by TryBuild in case
// of an error is not completely uninitialized: it does contain the error information as well,
// so that if it is mistakenly passed to an SDK method, the SDK can tell what the error was.
func (m *MultiBuilder) TryBuild() (Context, error) {
	c := m.Build()
	return c, c.Err()
}

// Add adds a nested context for a specific Kind to a MultiBuilder.
//
// It is invalid to add more than one context with the same Kind. This error is detected
// when you call [MultiBuilder.Build] or [MultiBuilder.TryBuild].
//
// If the parameter is a multi-context, this is exactly equivalent to adding each of the
// individual kinds from it separately. For instance, in the following example, "multi1" and
// "multi2" end up being exactly the same:
//
//	c1 := ldcontext.NewWithKind("kind1", "key1")
//	c2 := ldcontext.NewWithKind("kind2", "key2")
//	c3 := ldcontext.NewWithKind("kind3", "key3")
//
//	multi1 := ldcontext.NewMultiBuilder().Add(c1).Add(c2).Add(c3).Build()
//
//	c1plus2 := ldcontext.NewMultiBuilder().Add(c1).Add(c2).Build()
//	multi2 := ldcontext.NewMultiBuilder().Add(c1plus2).Add(c3).Build()
func (m *MultiBuilder) Add(context Context) *MultiBuilder {
	if m.contextsCopyOnWrite {
		m.contexts = slices.Clone(m.contexts)
		m.contextsCopyOnWrite = true
	}
	if context.Multiple() {
		m.contexts = append(m.contexts, context.multiContexts...)
	} else {
		m.contexts = append(m.contexts, context)
	}
	return m
}
