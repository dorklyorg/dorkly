package ldcontext

import (
	"fmt"
	"strings"

	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"golang.org/x/exp/slices"
)

// Builder is a mutable object that uses the builder pattern to specify properties for a Context.
//
// Use this type if you need to construct a [Context] that has only a single [Context.Kind]. To
// define a multi-context, use [MultiBuilder] instead.
//
// Obtain an instance of Builder by calling [NewBuilder]. Then, call setter methods such as
// [Builder.Kind] or [Builder.Name] to specify any additional attributes; all of the Builder
// setters return a reference to the same builder, so they can be chained together. Then, call
// [Builder.Build] to produce the immutable [Context].
//
//	context := ldcontext.NewBuilder("user-key").
//		Name("my-name").
//		SetString("country", "us").
//		Build()
//
// A Builder should not be accessed by multiple goroutines at once. Once you have called
// [Builder.Build], the resulting Context is immutable and is safe to use from multiple
// goroutines.
//
// # Context attributes
//
// There are several built-in attribute names with special meaning in LaunchDarkly, and
// restrictions on the type of their value. These have their own builder methods: see
// [Builder.Key], [Builder.Kind], [Builder.Name], and [Builder.Anonymous].
//
// You may also set any number of other attributes with whatever names are useful for your
// application (subject to validation constraints; see [Builder.SetValue] for rules regarding
// attribute names). These attributes can have any data type that is supported in JSON:
// boolean, number, string, array, or object.
//
// # Setting attributes with simple value types
//
// For convenience, there are setter methods for simple types:
//
//	context := ldcontext.NewBuilder("user-key").
//		SetBool("a", true).    // this attribute has a boolean value
//		SetString("b", "xyz"). // this attribute has a string value
//		SetInt("c", 3).        // this attribute has an integer numeric value
//		SetFloat64("d", 4.5).  // this attribute has a floating-point numeric value
//		Build()
//
// # Setting attributes with complex value types
//
// JSON arrays and objects are represented by the [ldvalue.Value] type. The [Builder.SetValue]
// method takes a value of this type.
//
// The [ldvalue] package provides several ways to construct such values. Here are some examples;
// for more information, see [ldvalue.Value].
//
//	context := ldcontext.NewBuilder("user-key").
//		SetValue("arrayAttr1",
//			ldvalue.ArrayOf(ldvalue.String("a"), ldvalue.String("b"))).
//		SetValue("arrayAttr2",
//			ldvalue.CopyArbitraryValue([]string{"a", "b"})).
//		SetValue("objectAttr1",
//			ldvalue.ObjectBuild().SetString("color", "green").Build()).
//		SetValue("objectAttr2",
//			ldvalue.FromJSONMarshal(MyStructType{Color: "green"})).
//		Build()
//
// Arrays and objects have special meanings in LaunchDarkly flag evaluation:
//   - An array of values means "try to match any of these values to the targeting rule."
//   - An object allows you to match a property within the object to the targeting rule. For instance,
//     in the example above, a targeting rule could reference /objectAttr1/color to match the value
//     "green". Nested property references like /objectAttr1/address/street are allowed if a property
//     contains another JSON object.
//
// # Private attributes
//
// You may designate certain attributes, or values within them, as "private", meaning that their
// values are not included in analytics data sent to LaunchDarkly. See [Builder.Private].
//
//	context := ldcontext.NewBuilder("user-key").
//		SetString("email", "test@example.com").
//		Private("email").
//		Build()
type Builder struct {
	kind               Kind
	key                string
	allowEmptyKey      bool
	name               ldvalue.OptionalString
	attributes         ldvalue.ValueMapBuilder
	anonymous          bool
	privateAttrs       []ldattr.Ref
	privateCopyOnWrite bool
}

// NewBuilder creates a Builder for building a Context, initializing its Key property and
// setting Kind to DefaultKind.
//
// You may use [Builder] methods to set additional attributes and/or change the [Builder.Kind]
// before calling [Builder.Build]. If you do not change any values, the defaults for the
// [Context] are that its [Builder.Kind] is [DefaultKind] ("user"), its [Builder.Key] is set
// to whatever value you passed to [NewBuilder], its [Builder.Anonymous] attribute is false,
// and it has no values for any other attributes.
//
// This method is for building a Context that has only a single Kind. To define a
// multi-Context, use [NewMultiBuilder] instead.
//
// If the key parameter is an empty string, there is no default. A Context must have a
// non-empty key, so if you call [Builder.Build] in this state without using [Builder.Key] to
// set the key, you will get an invalid Context.
//
// An empty Builder{} is valid as long as you call [Builder.Key] to set a non-empty key. This
// means that in in performance-critical code paths where you want to minimize heap allocations,
// if you do not want to allocate a Builder on the heap with NewBuilder, you can declare one
// locally instead:
//
//	var b ldcontext.Builder
//	c := b.Kind("org").Key("my-key").Name("my-name").Build()
func NewBuilder(key string) *Builder {
	b := &Builder{}
	return b.Key(key)
}

// NewBuilderFromContext creates a Builder whose properties are the same as an existing
// single context.
//
// You may then change the Builder's state in any way and call [Builder.Build] to create
// a new independent [Context].
//
// If fromContext is a multi-context created with [NewMulti] or [MultiBuilder], this method is
// not applicable and returns an uninitialized [Builder].
func NewBuilderFromContext(fromContext Context) *Builder {
	b := &Builder{}
	b.copyFrom(fromContext)
	return b
}

// Build creates a Context from the current Builder properties.
//
// The [Context] is immutable and will not be affected by any subsequent actions on the [Builder].
//
// It is possible to specify invalid attributes for a Builder, such as an empty [Builder.Key].
// Instead of returning two values (Context, error), the Builder always returns a Context and you
// can call [Context.Err] to see if it has an error. Using a single-return-value syntax is more
// convenient for application code, since in normal usage an application will never build an
// invalid Context. If you pass an invalid Context to an SDK method, the SDK will detect this and
// will generally log a description of the error.
//
// You may call [Builder.TryBuild] instead of Build if you prefer to use two-value return semantics,
// but the validation behavior is the same for both.
func (b *Builder) Build() Context {
	if b == nil {
		return Context{}
	}
	actualKind, err := validateSingleKind(b.kind)
	if err != nil {
		return Context{defined: true, err: err, kind: b.kind}
	}
	if b.key == "" && !b.allowEmptyKey {
		return Context{defined: true, err: lderrors.ErrContextKeyEmpty{}, kind: b.kind}
	}
	// We set the kind in the error cases above because that improves error reporting if this
	// context is used within a multi-context.

	ret := Context{
		defined:   true,
		kind:      actualKind,
		key:       b.key,
		name:      b.name,
		anonymous: b.anonymous,
	}

	ret.fullyQualifiedKey = makeFullyQualifiedKeySingleKind(actualKind, ret.key, true)
	ret.attributes = b.attributes.Build()
	if b.privateAttrs != nil {
		ret.privateAttrs = b.privateAttrs
		b.privateCopyOnWrite = true
		// The ___CopyOnWrite fields allow us to avoid the overhead of cloning maps/slices in
		// the typical case where Builder properties do not get modified after calling Build().
		// To guard against concurrent modification if someone does continue to modify the
		// Builder after calling Build(), we will clone the data later if and only if someone
		// tries to modify it when ___CopyOnWrite is true. That is safe as long as no one is
		// trying to modify Builder from two goroutines at once, which (per our documentation)
		// is not supported anyway.
	}

	return ret
}

// TryBuild is an alternative to Build that returns any validation errors as a second value.
//
// As described in [Builder.Build], there are several ways the state of a [Context] could be
// invalid. Since in normal usage it is possible to be confident that these will not occur,
// the Build method is designed for convenient use within expressions by returning a single
// Context value, and any validation problems are contained within that value where they can
// be detected by calling [Context.Err]. But, if you prefer to use the two-value pattern
// that is common in Go, you can call TryBuild instead:
//
//	c, err := ldcontext.NewBuilder("my-key").
//		Name("my-name").
//		TryBuild()
//	if err != nil {
//		// do whatever is appropriate if building the context failed
//	}
//
// The two return values are the same as to 1. the Context that would be returned by Build(),
// and 2. the result of calling Err() on that Context. So, the above example is exactly
// equivalent to:
//
//	c := ldcontext.NewBuilder("my-key").
//		Name("my-name").
//		Build()
//	if c.Err() != nil {
//		// do whatever is appropriate if building the context failed
//	}
//
// Note that unlike some Go methods where the first return value is normally an
// uninitialized zero value if the error is non-nil, the Context returned by TryBuild in case
// of an error is not completely uninitialized: it does contain the error information as well,
// so that if it is mistakenly passed to an SDK method, the SDK can tell what the error was.
func (b *Builder) TryBuild() (Context, error) {
	c := b.Build()
	return c, c.Err()
}

// Kind sets the Context's kind attribute.
//
// Every [Context] has a kind. Setting it to an empty string is equivalent to the default kind of
// "user". This value is case-sensitive. Validation rules are as follows:
//
//   - It may only contain letters, numbers, and the characters ".", "_", and "-".
//   - It cannot equal the literal string "kind".
//   - It cannot equal the literal string "multi" ([MultiKind]).
//
// If the value is invalid at the time [Builder.Build] is called, you will receive an invalid Context
// whose [Context.Err] value will describe the problem.
func (b *Builder) Kind(kind Kind) *Builder {
	if b != nil {
		if kind == "" {
			b.kind = DefaultKind
		} else {
			b.kind = kind
		}
	}
	return b
}

// Key sets the Context's key attribute.
//
// Every [Context] has a key, which is always a string. There are no restrictions on its value except
// that it cannot be empty.
//
// The key attribute can be referenced by flag rules, flag target lists, and segments.
//
// If the key is empty at the time [Builder.Build] is called, you will receive an invalid Context
// whose [Context.Err] value will describe the problem.
func (b *Builder) Key(key string) *Builder {
	if b != nil {
		b.key = key
	}
	return b
}

// Used internally when we are deserializing an old-style user from JSON; otherwise an empty key is
// never allowed.
func (b *Builder) setAllowEmptyKey(value bool) *Builder {
	if b != nil {
		b.allowEmptyKey = value
	}
	return b
}

// Name sets the Context's name attribute.
//
// This attribute is optional. It has the following special rules:
//   - Unlike most other attributes, it is always a string if it is specified.
//   - The LaunchDarkly dashboard treats this attribute as the preferred display name for contexts.
func (b *Builder) Name(name string) *Builder {
	if b == nil {
		return b
	}
	return b.OptName(ldvalue.NewOptionalString(name))
}

// OptName sets or clears the Context's name attribute.
//
// Calling b.OptName(ldvalue.NewOptionalString("x")) is equivalent to b.Name("x"), but since it uses
// the OptionalString type, it also allows clearing a previously set name with
// b.OptName(ldvalue.OptionalString{}).
func (b *Builder) OptName(name ldvalue.OptionalString) *Builder {
	if b != nil {
		b.name = name
	}
	return b
}

// SetBool sets an attribute to a boolean value.
//
// For rules regarding attribute names and values, see [Builder.SetValue]. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.Bool(value)).
func (b *Builder) SetBool(attributeName string, value bool) *Builder {
	return b.SetValue(attributeName, ldvalue.Bool(value))
}

// SetFloat64 sets an attribute to a float64 numeric value.
//
// For rules regarding attribute names and values, see [Builder.SetValue]. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.Float64(value)).
//
// Note: the LaunchDarkly model for feature flags and user attributes is based on JSON types,
// and JSON does not distinguish between integer and floating-point types. Therefore,
// b.SetFloat64(name, float64(1.0)) is exactly equivalent to b.SetInt(name, 1).
func (b *Builder) SetFloat64(attributeName string, value float64) *Builder {
	return b.SetValue(attributeName, ldvalue.Float64(value))
}

// SetInt sets an attribute to an int numeric value.
//
// For rules regarding attribute names and values, see [Builder.SetValue]. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.Int(value)).
//
// Note: the LaunchDarkly model for feature flags and user attributes is based on JSON types,
// and JSON does not distinguish between integer and floating-point types. Therefore,
// b.SetFloat64(name, float64(1.0)) is exactly equivalent to b.SetInt(name, 1).
func (b *Builder) SetInt(attributeName string, value int) *Builder {
	return b.SetValue(attributeName, ldvalue.Int(value))
}

// SetString sets an attribute to a string value.
//
// For rules regarding attribute names and values, see [Builder.SetValue]. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.String(value)).
func (b *Builder) SetString(attributeName string, value string) *Builder {
	return b.SetValue(attributeName, ldvalue.String(value))
}

// SetValue sets the value of any attribute for the Context.
//
// This method uses the [ldvalue.Value] type to represent a value of any JSON type: boolean,
// number, string, array, or object. The [ldvalue] package provides several ways to construct
// values of each type.
//
// The return value is always the same [Builder], for convenience (to allow method chaining).
//
// # Allowable attribute names
//
// The attribute names "kind", "key", "name", and "anonymous" have special meaning in
// LaunchDarkly. You may use these names with SetValue, as an alternative to using the
// methods [Builder.Kind], [Builder.Key], [Builder.Name], and [Builder.Anonymous]. However,
// there are restrictions on the value type: "kind" and "key" must be a string, "name" must
// be a string or null, and "anonymous" must be a boolean. Any value of an unsupported type
// is ignored (leaving the attribute unchanged).
//
// The string "_meta" cannot be used as an attribute name.
//
// All other non-empty strings are valid as an attribute name, and have no special meaning
// in LaunchDarkly; their definition is up to you.
//
// Context metadata such as [Builder.Private], which is not addressable in evaluations, is not
// considered an attribute; if you define an attribute of your own with the name "private",
// it is simply an attribute like any other, unrelated to the context metadata.
//
// # Simple value types
//
// Passing a simple value constructed with [ldvalue.Bool], [ldvalue.Int], [ldvalue.Float64],
// or [ldvalue.String], is exactly equivalent to calling one of the typed setter methods
// [Builder.SetBool], [Builder.SetInt], [Builder.SetFloat64], or [Builder.SetString].
//
// Values of different JSON types are always treated as different values. For instance, the
// number 1 is not the same as the string "1".
//
// The null value, [ldvalue.Null](), is a special case: it is a valid value in JSON, but
// LaunchDarkly considers null to be the same as "no such attribute", so setting an
// attribute's value to null is the same as removing it.
//
// # Complex value types
//
// The ldvalue package provides several ways to construct JSON array or object values. Here
// are some examples; for more information, see [ldvalue.Value].
//
//	context := ldcontext.NewBuilder("user-key").
//		SetValue("arrayAttr1",
//			ldvalue.ArrayOf(ldvalue.String("a"), ldvalue.String("b"))).
//		SetValue("arrayAttr2",
//			ldvalue.CopyArbitraryValue([]string{"a", "b"})).
//		SetValue("objectAttr1",
//			ldvalue.ObjectBuild().SetString("color", "green").Build()).
//		SetValue("objectAttr2",
//			ldvalue.FromJSONMarshal(MyStructType{Color: "green"})).
//		Build()
//
// Arrays and objects have special meanings in LaunchDarkly flag evaluation:
//   - An array of values means "try to match any of these values to the targeting rule."
//   - An object allows you to match a property within the object to the targeting rule. For instance,
//     in the example above, a targeting rule could reference /objectAttr1/color to match the value
//     "green". Nested property references like /objectAttr1/address/street are allowed if a property
//     contains another JSON object.
func (b *Builder) SetValue(attributeName string, value ldvalue.Value) *Builder {
	_ = b.TrySetValue(attributeName, value)
	return b
}

// TrySetValue sets the value of any attribute for the Context.
//
// This is the same as [Builder.SetValue], except that it returns true for success, or false if the
// parameters violated one of the restrictions described for SetValue (for instance,
// attempting to set "key" to a value that was not a string).
func (b *Builder) TrySetValue(attributeName string, value ldvalue.Value) bool {
	if b == nil || attributeName == "" {
		return false
	}
	switch attributeName {
	case ldattr.KindAttr:
		if !value.IsString() {
			return false
		}
		b.Kind(Kind(value.StringValue()))
	case ldattr.KeyAttr:
		if !value.IsString() {
			return false
		}
		b.Key(value.StringValue())
	case ldattr.NameAttr:
		if !value.IsString() && !value.IsNull() {
			return false
		}
		b.OptName(value.AsOptionalString())
	case ldattr.AnonymousAttr:
		if !value.IsBool() {
			return false
		}
		b.Anonymous(value.BoolValue())
	case jsonPropMeta:
		return false
	default:
		if value.IsNull() {
			b.attributes.Remove(attributeName)
		} else {
			b.attributes.Set(attributeName, value)
		}
		return true
	}
	return true
}

// Anonymous sets whether the Context is only intended for flag evaluations and should not be indexed by
// LaunchDarkly.
//
// The default value is false. False means that this [Context] represents an entity such as a user that you
// want to be able to see on the LaunchDarkly dashboard.
//
// Setting Anonymous to true excludes this Context from the database that is used by the dashboard. It does
// not exclude it from analytics event data, so it is not the same as making attributes private; all
// non-private attributes will still be included in events and data export. There is no limitation on what
// other attributes may be included (so, for instance, Anonymous does not mean there is no [Builder.Name]).
//
// This value is also addressable in evaluations as the attribute name "anonymous". It is always treated as
// a boolean true or false in evaluations; it cannot be null/undefined.
func (b *Builder) Anonymous(value bool) *Builder {
	if b != nil {
		b.anonymous = value
	}
	return b
}

// Private designates any number of Context attributes, or properties within them, as private: that is,
// their values will not be sent to LaunchDarkly in analytics data.
//
// This action only affects analytics events that involve this particular [Context]. To mark some (or all)
// Context attributes as private for all context, use the overall event configuration for the SDK.
//
// In this example, firstName is marked as private, but lastName is not:
//
//	c := ldcontext.NewBuilder("org", "my-key").
//		SetString("firstName", "Pierre").
//		SetString("lastName", "Menard").
//		Private("firstName").
//		Build()
//
// The attributes "kind", "key", and "anonymous" cannot be made private.
//
// This is a metadata property, rather than an attribute that can be addressed in evaluations: that is,
// a rule clause that references the attribute name "private" will not use this value, but instead will
// use whatever value (if any) you have set for that name with a method such as [Builder.SetString].
//
// # Designating an entire attribute as private
//
// If the parameter is an attribute name such as "email" that does not start with a '/' character, the
// entire attribute is private.
//
// # Designating a property within a JSON object as private
//
// If the parameter starts with a '/' character, it is interpreted as a slash-delimited path to a
// property within a JSON object. The first path component is an attribute name, and each following
// component is a property name.
//
// For instance, suppose that the attribute "address" had the following JSON object value:
// {"street": {"line1": "abc", "line2": "def"}, "city": "ghi"}
//
//   - Calling either Private("address") or Private("/address") would cause the entire "address"
//     attribute to be private.
//   - Calling Private("/address/street") would cause the "street" property to be private, so that
//     only {"city": "ghi"} is included in analytics.
//   - Calling Private("/address/street/line2") would cause only "line2" within "street" to be private,
//     so that {"street": {"line1": "abc"}, "city": "ghi"} is included in analytics.
//
// This syntax deliberately resembles JSON Pointer, but other JSON Pointer features such as array
// indexing are not supported for Private.
//
// If an attribute's actual name starts with a '/' character, you must use the same escaping syntax as
// JSON Pointer: replace "~" with "~0", and "/" with "~1".
func (b *Builder) Private(attrRefStrings ...string) *Builder {
	refs := make([]ldattr.Ref, 0, 20) // arbitrary capacity that's likely greater than needed, to preallocate on stack
	for _, s := range attrRefStrings {
		refs = append(refs, ldattr.NewRef(s))
	}
	return b.PrivateRef(refs...)
}

// PrivateRef is equivalent to Private, but uses the ldattr.Ref type. It designates any number of
// Context attributes, or properties within them, as private: that is, their values will not be
// sent to LaunchDarkly.
//
// Application code is unlikely to need to use the ldattr.Ref type directly; however, in cases where
// you are constructing Contexts constructed repeatedly with the same set of private attributes, if
// you are also using complex private attribute path references such as "/address/street", converting
// this to an [ldattr.Ref] once and reusing it in many PrivateRef calls is slightly more efficient than
// calling [Builder.Private] (since it does not need to parse the path repeatedly).
func (b *Builder) PrivateRef(attrRefs ...ldattr.Ref) *Builder {
	if b == nil {
		return b
	}
	if b.privateAttrs == nil {
		b.privateAttrs = make([]ldattr.Ref, 0, len(attrRefs))
	} else if b.privateCopyOnWrite {
		// See note in Build() on ___CopyOnWrite
		b.privateAttrs = slices.Clone(b.privateAttrs)
		b.privateCopyOnWrite = false
	}
	b.privateAttrs = append(b.privateAttrs, attrRefs...)
	return b
}

// RemovePrivate removes any private attribute references previously added with [Builder.Private]
// or [Builder.PrivateRef] that exactly match any of the specified attribute references.
func (b *Builder) RemovePrivate(attrRefStrings ...string) *Builder {
	refs := make([]ldattr.Ref, 0, 20) // arbitrary capacity that's likely greater than needed, to preallocate on stack
	for _, s := range attrRefStrings {
		refs = append(refs, ldattr.NewRef(s))
	}
	return b.RemovePrivateRef(refs...)
}

// RemovePrivateRef removes any private attribute references previously added with [Builder.Private]
// or [Builder.PrivateRef] that exactly match that of any of the specified attribute references.
//
// Application code is unlikely to need to use the [ldattr.Ref] type directly, and can use
// RemovePrivate with a string parameter to accomplish the same thing. This method is mainly for
// use by internal LaunchDarkly SDK and service code which uses ldattr.Ref.
func (b *Builder) RemovePrivateRef(attrRefs ...ldattr.Ref) *Builder {
	if b == nil {
		return b
	}
	if b.privateCopyOnWrite {
		// See note in Build() on ___CopyOnWrite
		b.privateAttrs = slices.Clone(b.privateAttrs)
		b.privateCopyOnWrite = false
	}
	for _, attrRefToRemove := range attrRefs {
		for i := 0; i < len(b.privateAttrs); i++ {
			if b.privateAttrs[i].String() == attrRefToRemove.String() {
				b.privateAttrs = append(b.privateAttrs[0:i], b.privateAttrs[i+1:]...)
				i--
			}
		}
	}
	return b
}

func (b *Builder) copyFrom(fromContext Context) {
	if fromContext.Multiple() || b == nil {
		return
	}
	b.kind = fromContext.kind
	b.key = fromContext.key
	b.name = fromContext.name
	b.anonymous = fromContext.anonymous
	b.attributes = ldvalue.ValueMapBuilder{}
	b.attributes.SetAllFromValueMap(fromContext.attributes)
	b.privateAttrs = fromContext.privateAttrs
	b.privateCopyOnWrite = true
}

func makeFullyQualifiedKeySingleKind(kind Kind, key string, omitDefaultKind bool) string {
	// Per the users-to-contexts specification, the fully-qualified key for a single context is:
	// - equal to the regular "key" property, if the kind is "user" (a.k.a. DefaultKind)
	// - or, for any other kind, it's the kind plus ":" plus the result of partially URL-encoding the
	// "key" property ("partially URL-encoding" here means that ':' and '%' are percent-escaped; other
	// URL-encoding behaviors are inconsistent across platforms, so we do not use a library function).
	if omitDefaultKind && kind == DefaultKind {
		return key
	}
	escapedKey := strings.ReplaceAll(strings.ReplaceAll(key, "%", "%25"), ":", "%3A")
	return fmt.Sprintf("%s:%s", kind, escapedKey)
}
