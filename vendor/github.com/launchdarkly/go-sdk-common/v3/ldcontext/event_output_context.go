package ldcontext

// EventOutputContext is a specialization of Context that uses the LaunchDarkly event schema.
//
// Applications will not normally need to read event data; this type is provided for use in
// LaunchDarkly service code and other tools. In JSON event data, contexts appear slightly
// differently. Marshaling or unmarshaling this type rather than Context causes this variant
// JSON schema to be used.
//
// The wrapped Context can have all of the same properties as a regular Context, except that
// the meaning of "private attributes" is slightly different: in event data, these are the
// attributes that were redacted due to private attribute configuration, whereas in a regular
// Context they are the attributes that should be redacted.
type EventOutputContext struct {
	Context
}
