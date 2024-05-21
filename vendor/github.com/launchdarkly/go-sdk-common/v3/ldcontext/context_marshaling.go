package ldcontext

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldattr"

	"github.com/launchdarkly/go-jsonstream/v3/jwriter"
)

// JSONString returns the JSON representation of the Context.
//
// This is equivalent to calling [Context.MarshalJSON] and converting the result to a string.
// An invalid Context cannot be represented in JSON and produces an empty string.
func (c Context) JSONString() string {
	bytes, _ := c.MarshalJSON()
	return string(bytes)
}

// MarshalJSON provides JSON serialization for Context when using [encoding/json.MarshalJSON].
//
// LaunchDarkly's JSON schema for contexts is standardized across SDKs. There are two output formats,
// depending on whether it is a single context or a multi-context. Unlike the unmarshaler,
// the marshaler never uses the old-style user context schema from older SDKs.
//
// If the Context is invalid (that is, it has a non-nil [Context.Err]) then marshaling fails with the
// same error.
func (c Context) MarshalJSON() ([]byte, error) {
	w := jwriter.NewWriter()
	ContextSerialization.MarshalToJSONWriter(&w, &c)
	return w.Bytes(), w.Error()
}

func (c *Context) writeToJSONWriterInternalSingle(w *jwriter.Writer, withinKind Kind, usingEventFormat bool) {
	obj := w.Object()
	if withinKind == "" {
		obj.Name(ldattr.KindAttr).String(string(c.kind))
	}

	obj.Name(ldattr.KeyAttr).String(c.key)
	if c.name.IsDefined() {
		obj.Name(ldattr.NameAttr).String(c.name.StringValue())
	}
	keys := make([]string, 0, 50) // arbitrary size to preallocate on stack
	for _, k := range c.attributes.Keys(keys) {
		obj.Name(k)
		c.attributes.Get(k).WriteToJSONWriter(w)
	}
	if c.anonymous {
		obj.Name(ldattr.AnonymousAttr).Bool(true)
	}

	needMeta := len(c.privateAttrs) != 0
	if needMeta {
		metaJSON := obj.Name(jsonPropMeta).Object()
		if len(c.privateAttrs) != 0 {
			name := jsonPropPrivate
			if usingEventFormat {
				name = jsonPropRedacted
			}
			privateAttrsJSON := metaJSON.Name(name).Array()
			for _, a := range c.privateAttrs {
				privateAttrsJSON.String(a.String())
			}
			privateAttrsJSON.End()
		}
		metaJSON.End()
	}

	obj.End()
}

func (c Context) writeToJSONWriterInternalMulti(w *jwriter.Writer, usingEventFormat bool) {
	obj := w.Object()
	obj.Name(ldattr.KindAttr).String(string(MultiKind))

	for _, mc := range c.multiContexts {
		obj.Name(string(mc.Kind()))
		mc.writeToJSONWriterInternalSingle(w, mc.Kind(), usingEventFormat)
	}

	obj.End()
}
