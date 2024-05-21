//go:build launchdarkly_easyjson

package ldcontext

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"github.com/launchdarkly/go-jsonstream/v3/jwriter"

	"github.com/mailru/easyjson/jlexer"
	ej_jwriter "github.com/mailru/easyjson/jwriter"
)

// This conditionally-compiled file provides custom marshal and unmarshal functions for the Context
// type in EasyJSON.
//
// EasyJSON's code generator does recognize the same MarshalJSON and UnmarshalJSON methods used by
// encoding/json, and will call them if present. But this mechanism is inefficient: when marshaling
// it requires the allocation of intermediate byte slices, and when unmarshaling it causes the
// JSON object to be parsed twice. It is preferable to have our marshal/unmarshal methods write to
// and read from the EasyJSON Writer/Lexer directly.
//
// Unmarshaling is the most performance-critical code path, because the client-side endpoints of
// the LD back-end use this implementation to get the context parameters for every request. So,
// rather than using an adapter to delegate jsonstream operations to EasyJSON, as we do for many
// other types-- which is preferred if performance is a bit less critical, because then we only
// have to write the logic once-- the Context unmarshaler is fully reimplemented here with direct
// calls to EasyJSON lexer methods. This allows us to take full advantage of EasyJSON optimizations
// that are available in our service code but may not be available in customer application code,
// such as the use of the unsafe package for direct []byte-to-string conversion.
//
// This means that if we make changes to the schema or the unmarshaling logic, we will need to
// update both context_unmarshaling.go and context_easyjson.go. Our unit tests run the same test
// data against both implementations to verify that they are in sync.
//
// For more information, see: https://github.com/launchdarkly/go-jsonstream/v3

// Arbitrary preallocation size that's likely to be longer than we will need for private/redacted
// attribute lists, to minimize reallocations during unmarshaling.
const initialAttrListAllocSize = 10

// MarshalEasyJSON is the marshaler method for Context when using the EasyJSON API. Because
// marshaling of contexts is not a requirement in high-traffic LaunchDarkly services, the
// current implementation delegates to the default non-EasyJSON marshaler.
//
// This method is only available when compiling with the build tag "launchdarkly_easyjson".
func (c Context) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if err := c.Err(); err != nil {
		writer.Error = err
		return
	}
	wrappedWriter := jwriter.NewWriterFromEasyJSONWriter(writer)
	ContextSerialization.MarshalToJSONWriter(&wrappedWriter, &c)
}

// MarshalEasyJSON is the marshaler method for EventOutputContext when using the EasyJSON API.
// Because marshaling of contexts is not a requirement in high-traffic LaunchDarkly services,
// the current implementation delegates to the default non-EasyJSON marshaler.
//
// This method is only available when compiling with the build tag "launchdarkly_easyjson".
func (c EventOutputContext) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if err := c.Err(); err != nil {
		writer.Error = err
		return
	}
	wrappedWriter := jwriter.NewWriterFromEasyJSONWriter(writer)
	ContextSerialization.MarshalToJSONWriterEventOutput(&wrappedWriter, &c)
}

// UnmarshalEasyJSON is the unmarshaler method for Context when using the EasyJSON API. Because
// unmarshaling of contexts is a requirement in high-traffic LaunchDarkly services, this
// implementation is optimized for speed and memory usage and does not share code with the default
// unmarshaler.
//
// This method is only available when compiling with the build tag "launchdarkly_easyjson".
func (c *Context) UnmarshalEasyJSON(in *jlexer.Lexer) {
	ContextSerialization.UnmarshalFromEasyJSONLexer(in, c)
}

// UnmarshalEasyJSON is the unmarshaler method for Context when using the EasyJSON API. Because
// unmarshaling of contexts is a requirement in high-traffic LaunchDarkly services, this
// implementation is optimized for speed and memory usage and does not share code with the default
// unmarshaler.
//
// This method is only available when compiling with the build tag "launchdarkly_easyjson".
func (c *EventOutputContext) UnmarshalEasyJSON(in *jlexer.Lexer) {
	ContextSerialization.UnmarshalFromEasyJSONLexerEventOutput(in, c)
}

// Note: other ContextSerialization methods are defined in context_serialization.go.

// UnmarshalFromEasyJSONLexer unmarshals a Context with the EasyJSON API. Because unmarshaling
// of contexts is a requirement in high-traffic LaunchDarkly services, this implementation is
// optimized for speed and memory usage and does not share code with the default unmarshaler.
//
// This method is only available when compiling with the build tag "launchdarkly_easyjson".
func (s ContextSerializationMethods) UnmarshalFromEasyJSONLexer(in *jlexer.Lexer, c *Context) {
	unmarshalFromEasyJSONLexer(in, c, false)
}

// UnmarshalFromEasyJSONLexerEventOutput unmarshals an EventContext with the EasyJSON API.
// Because unmarshaling of contexts in event data is a requirement in high-traffic LaunchDarkly
// services, this implementation is optimized for speed and memory usage and does not share code
// with the default unmarshaler.
//
// This method is only available when compiling with the build tag "launchdarkly_easyjson".
func (s ContextSerializationMethods) UnmarshalFromEasyJSONLexerEventOutput(in *jlexer.Lexer, c *EventOutputContext) {
	unmarshalFromEasyJSONLexer(in, &c.Context, true)
}

func unmarshalFromEasyJSONLexer(in *jlexer.Lexer, c *Context, usingEventFormat bool) {
	if in.IsNull() {
		in.Delim('{') // to trigger an "expected an object, got null" error
		return
	}

	// Do a first pass where we just check for the "kind" property, because that determines what
	// schema we use to parse everything else.
	kind, hasKind, err := parseKindOnlyEasyJSON(in)
	if err != nil {
		in.AddError(err)
		return
	}

	switch {
	case !hasKind:
		unmarshalOldUserSchemaEasyJSON(c, in, usingEventFormat)
	case kind == MultiKind:
		unmarshalMultiKindEasyJSON(c, in, usingEventFormat)
	default:
		unmarshalSingleKindEasyJSON(c, in, "", usingEventFormat)
	}
}

func unmarshalSingleKindEasyJSON(c *Context, in *jlexer.Lexer, knownKind Kind, usingEventFormat bool) {
	c.defined = true
	if knownKind != "" {
		c.kind = Kind(knownKind)
	}
	hasKey := false
	var attributes ldvalue.ValueMapBuilder
	in.Delim('{')
	for !in.IsDelim('}') {
		// Because the field name will often be a literal that we won't be retaining, we don't want the overhead
		// of allocating a string for it every time. So we call UnsafeBytes(), which still reads a JSON string
		// like String(), but returns the data as a subslice of the existing byte slice if possible-- allocating
		// a new byte slice only in the unlikely case that there were escape sequences. Go's switch statement is
		// optimized so that doing "switch string(key)" does *not* allocate a string, but just uses the bytes.
		key := in.UnsafeBytes()
		in.WantColon()
		switch string(key) {
		case ldattr.KindAttr:
			c.kind = Kind(in.String())
		case ldattr.KeyAttr:
			c.key = in.String()
			hasKey = true
		case ldattr.NameAttr:
			c.name = readOptStringEasyJSON(in)
		case ldattr.AnonymousAttr:
			c.anonymous = in.Bool()
		case jsonPropMeta:
			if in.IsNull() {
				in.Skip()
				break
			}
			in.Delim('{')
			for !in.IsDelim('}') {
				key := in.UnsafeBytes() // see comment above
				in.WantColon()
				switch {
				case string(key) == jsonPropPrivate && !usingEventFormat:
					readPrivateAttributesEasyJSON(in, c, false)
				case string(key) == jsonPropRedacted && usingEventFormat:
					readPrivateAttributesEasyJSON(in, c, false)
				default:
					// Unrecognized property names within _meta are ignored. Calling SkipRecursive makes the Lexer
					// consume and discard the property value so we can advance to the next object property.
					in.SkipRecursive()
				}
				in.WantComma()
			}
			in.Delim('}')
		default:
			if in.IsNull() {
				in.Skip()
			} else {
				var v ldvalue.Value
				v.UnmarshalEasyJSON(in)
				attributes.Set(internAttributeNameIfPossible(key), v)
			}
		}
		in.WantComma()
	}
	in.Delim('}')
	if in.Error() != nil {
		return
	}
	if !hasKey {
		in.AddError(lderrors.ErrContextKeyMissing{})
		return
	}
	c.kind, c.err = validateSingleKind(c.kind)
	if c.err != nil {
		in.AddError(c.err)
		return
	}
	if c.key == "" {
		c.err = lderrors.ErrContextKeyEmpty{}
		in.AddError(c.err)
	} else {
		c.fullyQualifiedKey = makeFullyQualifiedKeySingleKind(c.kind, c.key, true)
		c.attributes = attributes.Build()
	}
}

func unmarshalMultiKindEasyJSON(c *Context, in *jlexer.Lexer, usingEventFormat bool) {
	var b MultiBuilder
	in.Delim('{')
	for !in.IsDelim('}') {
		name := in.String()
		in.WantColon()
		if name == ldattr.KindAttr {
			in.SkipRecursive()
		} else {
			var subContext Context
			unmarshalSingleKindEasyJSON(&subContext, in, Kind(name), usingEventFormat)
			b.Add(subContext)
		}
		in.WantComma()
	}
	in.Delim('}')
	if in.Error() == nil {
		*c = b.Build()
		if err := c.Err(); err != nil {
			in.AddError(err)
		}
	}
}

func unmarshalOldUserSchemaEasyJSON(c *Context, in *jlexer.Lexer, usingEventFormat bool) {
	c.defined = true
	c.kind = DefaultKind
	hasKey := false
	var attributes ldvalue.ValueMapBuilder
	in.Delim('{')
	for !in.IsDelim('}') {
		// See comment about UnsafeBytes in unmarshalSingleKindEasyJSON.
		key := in.UnsafeBytes()
		in.WantColon()
		switch string(key) {
		case ldattr.KeyAttr:
			c.key = in.String()
			hasKey = true
		case ldattr.NameAttr:
			c.name = readOptStringEasyJSON(in)
		case jsonPropOldUserSecondary:
			c.secondary = readOptStringEasyJSON(in)
		case ldattr.AnonymousAttr:
			if in.IsNull() {
				in.Skip()
				c.anonymous = false
			} else {
				c.anonymous = in.Bool()
			}
		case jsonPropOldUserCustom:
			if in.IsNull() {
				in.Skip()
				attributes = ldvalue.ValueMapBuilder{}
				break
			}
			in.Delim('{')
			for !in.IsDelim('}') {
				name := in.String()
				in.WantColon()
				if in.IsNull() {
					in.Skip()
				} else {
					var value ldvalue.Value
					value.UnmarshalEasyJSON(in)
					if isOldUserCustomAttributeNameAllowed(name) {
						attributes.Set(name, value)
					}
				}
				in.WantComma()
			}
			in.Delim('}')
		case jsonPropOldUserPrivate:
			if usingEventFormat {
				in.SkipRecursive()
				break
			}
			readPrivateAttributesEasyJSON(in, c, true)
			// The "true" here means to interpret the strings as literal attribute names, since the
			// attribute reference path syntax was not used in the old user schema.
		case jsonPropOldUserRedacted:
			if !usingEventFormat {
				in.SkipRecursive()
				break
			}
			readPrivateAttributesEasyJSON(in, c, true)
		case "firstName", "lastName", "email", "country", "avatar", "ip":
			if in.IsNull() {
				in.Skip()
			} else {
				value := ldvalue.String(in.String())
				attributes.Set(internAttributeNameIfPossible(key), value)
			}
		default:
			// In the old user schema, unrecognized top-level property names are ignored. Calling SkipRecursive
			// makes the Lexer consume and discard the property value so we can advance to the next object property.
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if in.Error() != nil {
		return
	}
	if !hasKey {
		in.AddError(lderrors.ErrContextKeyMissing{})
		return
	}
	c.fullyQualifiedKey = c.key
	c.attributes = attributes.Build()
}

func parseKindOnlyEasyJSON(originalLexer *jlexer.Lexer) (Kind, bool, error) {
	// Make an exact copy of the original lexer so that changes in its state will not
	// affect the original lexer; both point to the same []byte array, but each has its
	// own "current position" and "next token" fields.
	in := *originalLexer
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if key == ldattr.KindAttr {
			kind := in.String()
			if in.Error() == nil && kind == "" {
				return "", false, lderrors.ErrContextKindEmpty{}
			}
			return Kind(kind), true, in.Error()
		}
		in.SkipRecursive()
		in.WantComma()
	}
	in.Delim('}')
	return "", false, in.Error()
}

func readOptStringEasyJSON(in *jlexer.Lexer) ldvalue.OptionalString {
	if in.IsNull() {
		in.Skip()
		return ldvalue.OptionalString{}
	} else {
		return ldvalue.NewOptionalString(in.String())
	}
}

func readPrivateAttributesEasyJSON(in *jlexer.Lexer, c *Context, asLiterals bool) {
	c.privateAttrs = nil
	if in.IsNull() {
		in.SkipRecursive()
		return
	}
	in.Delim('[')
	for !in.IsDelim(']') {
		if c.privateAttrs == nil {
			c.privateAttrs = make([]ldattr.Ref, 0, initialAttrListAllocSize)
		}
		c.privateAttrs = append(c.privateAttrs, refOrLiteralRef(in.String(), asLiterals))
		in.WantComma()
	}
	in.Delim(']')
}
