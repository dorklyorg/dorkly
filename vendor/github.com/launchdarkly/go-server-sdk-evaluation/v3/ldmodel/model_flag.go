package ldmodel

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	"github.com/launchdarkly/go-sdk-common/v3/ldtime"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
)

// FeatureFlag describes an individual feature flag.
//
// The fields of this struct are exported for use by LaunchDarkly internal components. Application code
// should normally not reference FeatureFlag fields directly; flag data normally comes from LaunchDarkly
// SDK endpoints in JSON form and can be deserialized using the DataModelSerialization interface.
type FeatureFlag struct {
	// Key is the unique string key of the feature flag.
	Key string
	// On is true if targeting is turned on for this flag.
	//
	// If On is false, the evaluator always uses OffVariation and ignores all other fields.
	On bool
	// Prerequisites is a list of feature flag conditions that are prerequisites for this flag.
	//
	// If any prerequisite is not met, the flag behaves as if targeting is turned off.
	Prerequisites []Prerequisite
	// Targets contains sets of individually targeted users for the default context kind (user).
	//
	// Targets take precedence over Rules: if a user is matched by any Target, the Rules are ignored.
	// Targets are ignored if targeting is turned off.
	Targets []Target
	// ContextTargets contains sets of individually targeted users for specific context kinds.
	//
	// Targets take precedence over Rules: if a user is matched by any Target, the Rules are ignored.
	// Targets are ignored if targeting is turned off.
	ContextTargets []Target
	// Rules is a list of rules that may match a user.
	//
	// If a user is matched by a Rule, all subsequent Rules in the list are skipped. Rules are ignored
	// if targeting is turned off.
	Rules []FlagRule
	// Fallthrough defines the flag's behavior if targeting is turned on but the user is not matched
	// by any Target or Rule.
	Fallthrough VariationOrRollout
	// OffVariation specifies the variation index to use if targeting is turned off.
	//
	// If this is undefined (ldvalue.OptionalInt{}), Evaluate returns undefined for the variation
	// index and ldvalue.Null() for the value.
	OffVariation ldvalue.OptionalInt
	// Variations is the list of all allowable variations for this flag. The variation index in a
	// Target or Rule is a zero-based index to this list.
	Variations []ldvalue.Value
	// ClientSideAvailability indicates whether a flag is available using each of the client-side
	// authentication methods.
	ClientSideAvailability ClientSideAvailability
	// Salt is a randomized value assigned to this flag when it is created.
	//
	// The hash function used for calculating percentage rollouts uses this as a salt to ensure that
	// rollouts are consistent within each flag but not predictable from one flag to another.
	Salt string
	// TrackEvents is used internally by the SDK analytics event system.
	//
	// This field is true if the current LaunchDarkly account has data export enabled, and has turned on
	// the "send detailed event information for this flag" option for this flag. This tells the SDK to
	// send full event data for each flag evaluation, rather than only aggregate data in a summary event.
	//
	// The go-server-sdk-evaluation package does not implement that behavior; it is only in the data
	// model for use by the SDK.
	TrackEvents bool
	// TrackEventsFallthrough is used internally by the SDK analytics event system.
	//
	// This field is true if the current LaunchDarkly account has experimentation enabled, has associated
	// this flag with an experiment, and has enabled "default rule" for the experiment. This tells the
	// SDK to send full event data for any evaluation where this flag had targeting turned on but the
	// user did not match any targets or rules.
	//
	// The go-server-sdk-evaluation package does not implement that behavior; it is only in the data
	// model for use by the SDK.
	TrackEventsFallthrough bool
	// DebugEventsUntilDate is used internally by the SDK analytics event system.
	//
	// This field is non-zero if debugging for this flag has been turned on temporarily in the
	// LaunchDarkly dashboard. Debugging always is for a limited time, so the field specifies a Unix
	// millisecond timestamp when this mode should expire. Until then, the SDK will send full event data
	// for each evaluation of this flag.
	//
	// The go-server-sdk-evaluation package does not implement that behavior; it is only in the data
	// model for use by the SDK.
	DebugEventsUntilDate ldtime.UnixMillisecondTime
	// Version is an integer that is incremented by LaunchDarkly every time the configuration of the flag is
	// changed.
	Version int
	// Deleted is true if this is not actually a feature flag but rather a placeholder (tombstone) for a
	// deleted flag. This is only relevant in data store implementations. The SDK does not evaluate
	// deleted flags.
	Deleted bool
	// Migration contains migration-related flag parameters. If this flag is
	// for migration purposes, this property is guaranteed to be set.
	Migration *MigrationFlagParameters
	// SamplingRatio controls the rate at which feature and debug events are emitted from the SDK for
	// this particular flag. If this value is not defined, it is assumed to be 1. Non-positive
	// values will disable emission entirely.
	//
	// LaunchDarkly may affect this flag to prevent poorly performing applications from adversely
	// affecting upstream service health.
	SamplingRatio ldvalue.OptionalInt
	// ExcludeFromSummaries determines whether or not this flag will be excluded from the event
	// summarization process.
	//
	// LaunchDarkly may affect this flag to prevent poorly performing applications from adversely
	// affecting upstream service health.
	ExcludeFromSummaries bool
}

// MigrationFlagParameters are used to control flag-specific migration
// configuration.
type MigrationFlagParameters struct {
	// CheckRatio controls the rate at which consistency checks are performing during a
	// migration-influenced read or write operation. This value can be controlled through the
	// LaunchDarkly UI and propagated downstream to the SDKs.
	CheckRatio ldvalue.OptionalInt
}

// FlagRule describes a single rule within a feature flag.
//
// A rule consists of a set of ANDed matching conditions (Clause) for a user, along with either a fixed
// variation or a set of rollout percentages to use if the user matches all of the clauses.
type FlagRule struct {
	// VariationRollout properties for a FlagRule define what variation to return if the user matches
	// this rule.
	VariationOrRollout
	// ID is a randomized identifier assigned to each rule when it is created.
	//
	// This is used to populate the RuleID property of ldreason.EvaluationReason.
	ID string
	// Clauses is a list of test conditions that make up the rule. These are ANDed: every Clause must
	// match in order for the FlagRule to match.
	Clauses []Clause
	// TrackEvents is used internally by the SDK analytics event system.
	//
	// This field is true if the current LaunchDarkly account has experimentation enabled, has associated
	// this flag with an experiment, and has enabled this rule for the experiment. This tells the SDK to
	// send full event data for any evaluation that matches this rule.
	//
	// The go-server-sdk-evaluation package does not implement that behavior; it is only in the data
	// model for use by the SDK.
	TrackEvents bool
}

// RolloutKind describes whether a rollout is a simple percentage rollout or represents an experiment. Experiments have
// different behaviour for tracking and variation bucketing.
type RolloutKind string

const (
	// RolloutKindRollout represents a simple percentage rollout. This is the default rollout kind, and will be assumed if
	// not otherwise specified.
	RolloutKindRollout RolloutKind = "rollout"
	// RolloutKindExperiment represents an experiment. Experiments have different behaviour for tracking and variation
	// bucketing.
	RolloutKindExperiment RolloutKind = "experiment"
)

// VariationOrRollout desscribes either a fixed variation or a percentage rollout.
//
// There is a VariationOrRollout for every FlagRule, and also one in FeatureFlag.Fallthrough which is
// used if no rules match.
//
// Invariant: one of the variation or rollout must be non-nil.
type VariationOrRollout struct {
	// Variation specifies the index of the variation to return. It is undefined (ldvalue.OptionalInt{})
	// if no specific variation is defined.
	Variation ldvalue.OptionalInt
	// Rollout specifies a percentage rollout to be used instead of a specific variation. A rollout is
	// only defined if it has a non-empty Variations list.
	Rollout Rollout
}

// Rollout describes how users will be bucketed into variations during a percentage rollout.
type Rollout struct {
	// Kind specifies whether this rollout is a simple percentage rollout or represents an experiment. Experiments have
	// different behaviour for tracking and variation bucketing.
	Kind RolloutKind
	// ContextKind is the context kind that this rollout will use to get any necessary context attributes.
	//
	// LaunchDarkly will normally always set this property, but if it is empty/omitted, it should be
	// treated as ldcontext.DefaultKind. An empty string value here represents the property being unset
	// (so it will be omitted in serialization).
	ContextKind ldcontext.Kind
	// Variations is a list of the variations in the percentage rollout and what percentage of users
	// to include in each.
	//
	// The Weight values of all elements in this list should add up to 100000 (100%). If they do not,
	// the last element in the list will behave as if it includes any leftover percentage (that is, if
	// the weights are [1000, 1000, 1000] they will be treated as if they were [1000, 1000, 98000]).
	Variations []WeightedVariation
	// BucketBy specifies which user attribute should be used to distinguish between users in a rollout.
	// This only works for simple rollouts; it is ignored for experiments.
	//
	// The default (when BucketBy is empty) is ldattr.KeyAttr, the user's primary key. If you wish to
	// treat users with different keys as the same for rollout purposes as long as they have the same
	// "country" attribute, you would set this to "country".
	//
	// Simple rollouts always take the user's "secondary key" attribute into account as well if the user
	// has one. Experiments ignore the secondary key.
	BucketBy ldattr.Ref
	// Seed, if present, specifies the seed for the hashing algorithm this rollout will use to bucket users, so that
	// rollouts with the same Seed will assign the same users to the same buckets.
	// If unspecified, the seed will default to a combination of the flag key and flag-level Salt.
	Seed ldvalue.OptionalInt
}

// IsExperiment returns whether this rollout represents an experiment.
func (r Rollout) IsExperiment() bool {
	return r.Kind == RolloutKindExperiment
}

// Clause describes an individual clause within a FlagRule or SegmentRule.
type Clause struct {
	// ContextKind is the context kind that this clause applies to.
	//
	// LaunchDarkly will normally always set this property, but if it is empty/omitted, it should be
	// treated as ldcontext.DefaultKind. An empty string value here represents the property being unset (so
	// it will be omitted in serialization).
	//
	// If the value of Attribute is "kind", then ContextKind is ignored because the nature of the context kind
	// test is described in a richer way by Operator and Values.
	ContextKind ldcontext.Kind
	// Attribute specifies the context attribute that is being tested.
	//
	// This is required for all Operator types except SegmentMatch. If Op is SegmentMatch then Attribute
	// is ignored (and will normally be an empty ldattr.Ref{}).
	//
	// If the context's value for this attribute is a JSON array, then the test specified in the Clause is
	// repeated for each value in the array until a match is found or there are no more values.
	Attribute ldattr.Ref
	// Op specifies the type of test to perform.
	Op Operator
	// Values is a list of values to be compared to the user attribute.
	//
	// This is interpreted as an OR: if the user attribute matches any of these values with the specified
	// operator, the Clause matches the user.
	//
	// In the special case where Op is OperatorSegmentMtach, there should only be a single Value, which
	// must be a string: the key of the user segment.
	//
	// If the user does not have a value for the specified attribute, the Values are ignored and the
	// Clause is always treated as a non-match.
	Values []ldvalue.Value
	// Negate is true if the specified Operator should be inverted.
	//
	// For instance, this would cause OperatorIn to mean "not equal" rather than "equal". Note that if no
	// tests are performed for this Clause because the user does not have a value for the specified
	// attribute, then Negate will not come into effect (the Clause will just be treated as a non-match).
	Negate bool
	// preprocessed is created by PreprocessFlag() to speed up clause evaluation in scenarios like
	// regex matching.
	preprocessed clausePreprocessedData
}

// WeightedVariation describes a fraction of users who will receive a specific variation.
type WeightedVariation struct {
	// Variation is the index of the variation to be returned if the user is in this bucket. This is
	// always a real variation index; it cannot be undefined.
	Variation int
	// Weight is the proportion of users who should go into this bucket, as an integer from 0 to 100000.
	Weight int
	// Untracked means that users allocated to this variation should not have tracking events sent.
	Untracked bool
}

// Target describes a set of users who will receive a specific variation.
type Target struct {
	// ContextKind is the context kind that this target list applies to.
	//
	// LaunchDarkly will normally always set this property, but if it is empty/omitted, it should be
	// treated as ldcontext.DefaultKind. An empty string value here represents the property being unset (so
	// it will be omitted in serialization).
	ContextKind ldcontext.Kind
	// Values is the set of user keys included in this Target.
	Values []string
	// Variation is the index of the variation to be returned if the user matches one of these keys. This
	// is always a real variation index; it cannot be undefined.
	Variation int
	// preprocessed is created by PreprocessFlag() to speed up target matching.
	preprocessed targetPreprocessedData
}

// Prerequisite describes a requirement that another feature flag return a specific variation.
//
// A prerequisite condition is met if the specified prerequisite flag has targeting turned on and
// returns the specified variation.
type Prerequisite struct {
	// Key is the unique key of the feature flag to be evaluated as a prerequisite.
	Key string
	// Variation is the index of the variation that the prerequisite flag must return in order for
	// the prerequisite condition to be met. If the prerequisite flag has targeting turned on, then
	// the condition is not met even if the flag's OffVariation matches this value. This is always a
	// real variation index; it cannot be undefined.
	Variation int
}

// ClientSideAvailability describes whether a flag is available to client-side SDKs.
//
// This field can be used by a server-side client to determine whether to include an individual flag in
// bootstrapped set of flag data (see https://docs.launchdarkly.com/sdk/client-side/javascript#bootstrapping).
type ClientSideAvailability struct {
	// UsingMobileKey indicates that this flag is available to clients using the mobile key for authorization
	// (includes most desktop and mobile clients).
	UsingMobileKey bool
	// UsingEnvironmentID indicates that this flag is available to clients using the environment id to identify an
	// environment (includes client-side javascript clients).
	UsingEnvironmentID bool
	// Explicit is true if, when serializing this flag, all of the ClientSideAvailability properties should
	// be included. If it is false, then an older schema is used in which this object is entirely omitted,
	// UsingEnvironmentID is stored in a deprecated property, and UsingMobileKey is assumed to be true.
	//
	// This field exists to ensure that flag representations remain consistent when sent and received
	// even though the clientSideAvailability property may not be present in the JSON data. It is false
	// if the flag was deserialized from an older JSON schema that did not include that property.
	//
	// Similarly, when deserializing a flag, if it used the older schema then Explicit will be false and
	// UsingMobileKey will be true.
	Explicit bool
}
