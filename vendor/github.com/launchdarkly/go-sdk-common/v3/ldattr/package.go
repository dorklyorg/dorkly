// Package ldattr defines the model for context attribute references used by the LaunchDarkly SDK.
//
// This includes the [Ref] type, which provides a syntax similar to JSON Pointer for
// referencing values either of a top-level context attribute, or of a value within a JSON object
// or JSON array. It also includes constants for the names of some built-in attributes.
//
// These types and constants are mainly intended to be used internally by LaunchDarkly SDK and
// service code. Applications are unlikely to need to use them directly; context attributes are
// normally set via methods in [github.com/launchdarkly/go-sdk-common/v3/ldcontext].
package ldattr
