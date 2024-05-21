// Package ldreason provides types that describe the outcome of a LaunchDarkly flag evaluation.
//
// You do not need to use these types to evaluate feature flags with the LaunchDarkly SDK. They are
// only required for the "detail" evaluation methods that allow you to determine whether the result
// of a flag evaluation was due to, for instance, context key targeting or a specific flag rule, such
// as [github.com/launchdarkly/go-server-sdk/v6.LDClient.BoolVariationDetail].
package ldreason
