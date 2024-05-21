package interfaces

// ServiceEndpoints allow configuration of custom service URIs.
//
// If you want to set non-default values for any of these fields,
// set the ServiceEndpoints field in the SDK's
// [github.com/launchdarkly/go-server-sdk/v7.Config] struct.
// You may set individual values such as Streaming, or use the helper method
// [github.com/launchdarkly/go-server-sdk/v7/ldcomponents.RelayProxyEndpoints].
//
// Important note: if one or more URI is set to a custom value, then
// all URIs should be set to custom values. Otherwise, the SDK will emit
// an error-level log to surface this potential misconfiguration, while
// using default values for the unset URIs.
//
// There are some scenarios where it is desirable to set only some of the
// fields, but this is not recommended for general usage. If your scenario
// requires it, you can call [WithPartialSpecification] to suppress the
// error message.
//
// See Config.ServiceEndpoints for more details.
type ServiceEndpoints struct {
	Streaming                 string
	Polling                   string
	Events                    string
	allowPartialSpecification bool
}

// WithPartialSpecification returns a copy of this ServiceEndpoints that will
// not trigger an error-level log message if only some, but not all the fields
// are set to custom values. This is an advanced configuration and likely not
// necessary for most use-cases.
func (s ServiceEndpoints) WithPartialSpecification() ServiceEndpoints {
	s.allowPartialSpecification = true
	return s
}

// PartialSpecificationRequested returns true if this ServiceEndpoints should not
// be treated as malformed if some, but not all fields are set.
func (s ServiceEndpoints) PartialSpecificationRequested() bool {
	return s.allowPartialSpecification
}
