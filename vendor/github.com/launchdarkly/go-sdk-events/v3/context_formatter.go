package ldevents

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"github.com/launchdarkly/go-jsonstream/v3/jwriter"
)

// eventContextFormatter provides the special JSON serialization format that is used when including Context
// data in analytics events. In this format, some attribute values may be redacted based on the SDK's
// events configuration and/or the per-Context setting of ldcontext.Builder.Private().
type eventContextFormatter struct {
	allAttributesPrivate bool
	privateAttributes    map[string]*privateAttrLookupNode
}

type privateAttrLookupNode struct {
	attribute *ldattr.Ref
	children  map[string]*privateAttrLookupNode
}

// newEventContextFormatter creates an eventContextFormatter.
//
// An instance of this type is owned by the eventOutputFormatter that is responsible for writing all
// JSON event data. It is created at SDK initialization time based on the SDK configuration.
func newEventContextFormatter(config EventsConfiguration) eventContextFormatter {
	ret := eventContextFormatter{allAttributesPrivate: config.AllAttributesPrivate}
	if len(config.PrivateAttributes) != 0 {
		// Reformat the list of private attributes into a map structure that will allow
		// for faster lookups.
		ret.privateAttributes = makePrivateAttrLookupData(config.PrivateAttributes)
	}
	return ret
}

func makePrivateAttrLookupData(attrRefList []ldattr.Ref) map[string]*privateAttrLookupNode {
	// This function transforms a list of AttrRefs into a data structure that allows for more efficient
	// implementation of eventContextFormatter.checkGloballyPrivate().
	//
	// For instance, if the original AttrRefs were "/name", "/address/street", and "/address/city",
	// it would produce the following map:
	//
	// "name": {
	//   attribute: NewAttrRef("/name"),
	// },
	// "address": {
	//   children: {
	//     "street": {
	//       attribute: NewAttrRef("/address/street/"),
	//     },
	//     "city": {
	//       attribute: NewAttrRef("/address/city/"),
	//     },
	//   },
	// }
	ret := make(map[string]*privateAttrLookupNode)
	for _, a := range attrRefList {
		parentMap := &ret
		for i := 0; i < a.Depth(); i++ {
			name := a.Component(i)
			if *parentMap == nil {
				*parentMap = make(map[string]*privateAttrLookupNode)
			}
			nextNode := (*parentMap)[name]
			if nextNode == nil {
				nextNode = &privateAttrLookupNode{}
				if i == a.Depth()-1 {
					aa := a
					nextNode.attribute = &aa
				}
				(*parentMap)[name] = nextNode
			}
			parentMap = &nextNode.children
		}
	}
	return ret
}

// WriteContext serializes a Context in the format appropriate for an analytics event, redacting
// private attributes if necessary.
func (f *eventContextFormatter) WriteContext(w *jwriter.Writer, ec *EventInputContext) {
	f.writeContext(w, ec, false)
}

// WriteContextRedactAnonymous serializes a Context in the format appropriate for an analytics event, redacting
// private attributes if necessary. If the context is anonymous, ALL attributes will be redacted except key,
// kind, and anonymous.
func (f *eventContextFormatter) WriteContextRedactAnonymous(w *jwriter.Writer, ec *EventInputContext) {
	f.writeContext(w, ec, true)
}

func (f *eventContextFormatter) writeContext(w *jwriter.Writer, ec *EventInputContext, redactAnonymous bool) {
	if ec.preserialized != nil {
		w.Raw(ec.preserialized)
		return
	}
	if ec.context.Err() != nil {
		w.AddError(ec.context.Err())
		return
	}
	if ec.context.Multiple() {
		f.writeContextInternalMulti(w, ec, redactAnonymous)
	} else {
		f.writeContextInternalSingle(w, &ec.context, true, redactAnonymous)
	}
}

func (f *eventContextFormatter) writeContextInternalSingle(
	w *jwriter.Writer,
	c *ldcontext.Context,
	includeKind,
	redactAnonymous bool,
) {
	redactAll := f.allAttributesPrivate || (c.Anonymous() && redactAnonymous)
	obj := w.Object()
	if includeKind {
		obj.Name(ldattr.KindAttr).String(string(c.Kind()))
	}

	obj.Name(ldattr.KeyAttr).String(c.Key())

	optionalAttrNames := make([]string, 0, 50) // arbitrary capacity, expanded if necessary by GetOptionalAttributeNames
	redactedAttrs := make([]string, 0, 20)

	optionalAttrNames = c.GetOptionalAttributeNames(optionalAttrNames)

	for _, key := range optionalAttrNames {
		if value := c.GetValue(key); value.IsDefined() {
			if redactAll {
				// If redactAll is true, then there's no complex filtering or recursing to be done: all of
				// these values are by definition private, so just add their names to the redacted list. Since the
				// redacted list uses the attribute reference syntax, we may need to escape the value if the name of
				// this individual attribute happens to be something like "/a/b"; the easiest way to do that is to
				// call NewLiteralRef and then convert the Ref to an attribute reference string.
				escapedAttrName := ldattr.NewLiteralRef(key).String()
				redactedAttrs = append(redactedAttrs, escapedAttrName)
				continue
			}
			path := make([]string, 0, 10)
			f.writeFilteredAttribute(w, c, &obj, path, key, value, &redactedAttrs)
		}
	}

	if c.Anonymous() {
		obj.Name(ldattr.AnonymousAttr).Bool(true)
	}

	anyRedacted := len(redactedAttrs) != 0
	if anyRedacted {
		metaJSON := obj.Name("_meta").Object()
		privateAttrsJSON := metaJSON.Name("redactedAttributes").Array()
		for _, a := range redactedAttrs {
			privateAttrsJSON.String(a)
		}
		privateAttrsJSON.End()
		metaJSON.End()
	}

	obj.End()
}

func (f *eventContextFormatter) writeContextInternalMulti(w *jwriter.Writer, ec *EventInputContext, redactAnonymous bool) {
	obj := w.Object()
	obj.Name(ldattr.KindAttr).String(string(ldcontext.MultiKind))

	for i := 0; i < ec.context.IndividualContextCount(); i++ {
		if ic := ec.context.IndividualContextByIndex(i); ic.IsDefined() {
			obj.Name(string(ic.Kind()))
			f.writeContextInternalSingle(w, &ic, false, redactAnonymous)
		}
	}

	obj.End()
}

// writeFilteredAttribute checks whether a given value should be considered private, and then
// either writes the attribute to the output JSON object if it is *not* private, or adds the
// corresponding attribute reference to the redactedAttrs list if it is private.
//
// The parentPath parameter indicates where we are in the context data structure. If it is empty,
// we are at the top level and "key" is an attribute name. If it is not empty, we are recursing
// into the properties of an attribute value that is a JSON object: for instance, if parentPath
// is ["billing", "address"] and key is "street", then the top-level attribute is "billing" and
// has a value in the form {"address": {"street": ...}} and we are now deciding whether to
// write the "street" property. See maybeRedact() for the logic involved in that decision.
//
// If allAttributesPrivate is true, this method is never called.
func (f *eventContextFormatter) writeFilteredAttribute(
	w *jwriter.Writer,
	c *ldcontext.Context,
	parentObj *jwriter.ObjectState,
	parentPath []string,
	key string,
	value ldvalue.Value,
	redactedAttrs *[]string,
) {
	path := append(parentPath, key) //nolint:gocritic // purposely not assigning to same slice

	isRedacted, nestedPropertiesAreRedacted := f.maybeRedact(c, path, value.Type(), redactedAttrs)

	if value.Type() != ldvalue.ObjectType {
		// For all value types except object, the question is only "is there a private attribute
		// reference that directly points to this property", since there are no nested properties.
		if !isRedacted {
			parentObj.Name(key)
			value.WriteToJSONWriter(w)
		}
		return
	}

	// If the value is an object, then there are three possible outcomes: 1. this value is
	// completely redacted, so drop it and do not recurse; 2. the value is not redacted, and
	// and neither are any subproperties within it, so output the whole thing as-is; 3. the
	// value itself is not redacted, but some subproperties within it are, so we'll need to
	// recurse through it and filter as we go.
	if isRedacted {
		return // outcome 1
	}
	parentObj.Name(key)
	if !nestedPropertiesAreRedacted {
		value.WriteToJSONWriter(w) // writes the whole value unchanged
		return                     // outcome 2
	}
	subObj := w.Object() // writes the opening brace for the output object

	objectKeys := make([]string, 0, 50) // arbitrary capacity, expanded if necessary by value.Keys()
	for _, subKey := range value.Keys(objectKeys) {
		subValue := value.GetByKey(subKey)
		// recurse to write or not write each property - outcome 3
		f.writeFilteredAttribute(w, c, &subObj, path, subKey, subValue, redactedAttrs)
	}
	subObj.End() // writes the closing brace for the output object
}

// maybeRedact is called by writeFilteredAttribute to decide whether or not a given value (or,
// possibly, properties within it) should be considered private, based on the private attribute
// references in either 1. the eventContextFormatter configuration or 2. this specific Context.
//
// If the value should be private, then the first return value is true, and also the attribute
// reference is added to redactedAttrs.
//
// The second return value indicates whether there are any private attribute references
// designating properties *within* this value. That is, if attrPath is ["address"], and the
// configuration says that "/address/street" is private, then the second return value will be
// true, which tells us that we can't just dump the value of the "address" object directly into
// the output but will need to filter its properties.
//
// Note that even though an AttrRef can contain numeric path components to represent an array
// element lookup, for the purposes of flag evaluations (like "/animals/0" which conceptually
// represents context.animals[0]), those will not work as private attribute references since
// we do not recurse to redact anything within an array value. A reference like "/animals/0"
// would only work if context.animals were an object with a property named "0".
//
// If allAttributesPrivate is true, this method is never called.
func (f *eventContextFormatter) maybeRedact(
	c *ldcontext.Context,
	attrPath []string,
	valueType ldvalue.ValueType,
	redactedAttrs *[]string,
) (bool, bool) {
	// First check against the eventContextFormatter configuration.
	redactedAttrRef, nestedPropertiesAreRedacted := f.checkGlobalPrivateAttrRefs(attrPath)
	if redactedAttrRef != nil {
		*redactedAttrs = append(*redactedAttrs, redactedAttrRef.String())
		return true, false
		// true, false = "this attribute itself is redacted, never mind its children"
	}

	shouldCheckForNestedProperties := valueType == ldvalue.ObjectType

	// Now check the per-Context configuration. Unlike the eventContextFormatter configuration, this
	// does not have a lookup map, just a list of AttrRefs.
	for i := 0; i < c.PrivateAttributeCount(); i++ {
		a, _ := c.PrivateAttributeByIndex(i)
		depth := a.Depth()
		if depth < len(attrPath) {
			// If the attribute reference is shorter than the current path, then it can't possibly be a match,
			// because if it had matched the first part of our path, we wouldn't have recursed this far.
			continue
		}
		if !shouldCheckForNestedProperties && depth > len(attrPath) {
			continue
		}
		match := true
		for j := 0; j < len(attrPath); j++ {
			name := a.Component(j)
			if name != attrPath[j] {
				match = false
				break
			}
		}
		if match {
			if depth == len(attrPath) {
				*redactedAttrs = append(*redactedAttrs, a.String())
				return true, false
				// true, false = "this attribute itself is redacted, never mind its children"
			}
			nestedPropertiesAreRedacted = true
		}
	}
	return false, nestedPropertiesAreRedacted // false = "this attribute itself is not redacted"
}

// Checks whether the given attribute or subproperty matches any AttrRef that was designated as
// private in the SDK options given to newEventContextFormatter.
//
// If attrPath has just one element, it is the name of a top-level attribute. If it has multiple
// elements, it is a path to a property within a custom object attribute: for instance, if you
// represented the overall context as a JSON object, the attrPath ["billing", "address", "street"]
// would refer to the street property within something like {"billing": {"address": {"street": "x"}}}.
//
// The first return value is nil if the attribute does not need to be redacted; otherwise it is the
// specific attribute reference that was matched.
//
// The second return value is true if and only if there's at least one configured private
// attribute reference for *children* of attrPath (and there is not one for attrPath itself, since if
// there was, we would not bother recursing to write the children). See comments on writeFilteredAttribute.
func (f eventContextFormatter) checkGlobalPrivateAttrRefs(attrPath []string) (
	redactedAttrRef *ldattr.Ref, nestedPropertiesAreRedacted bool,
) {
	redactedAttrRef = nil
	nestedPropertiesAreRedacted = false
	lookup := f.privateAttributes
	if lookup == nil {
		return
	}
	for i, pathComponent := range attrPath {
		nextNode := lookup[pathComponent]
		if nextNode == nil {
			break
		}
		if i == len(attrPath)-1 {
			if nextNode.attribute != nil {
				redactedAttrRef = nextNode.attribute
				return
			}
			nestedPropertiesAreRedacted = true
			return
		} else if nextNode.children != nil {
			lookup = nextNode.children
			continue
		}
	}
	return
}
