package ldmodel

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
)

// Segment describes a group of contexts based on context keys and/or matching rules.
type Segment struct {
	// Key is the unique key of the user segment.
	Key string
	// Included is a list of context keys that are always matched by this segment, for the default context kind
	// ("user"). Other context kinds are covered by IncludedContexts.
	Included []string
	// Excluded is a list of user keys that are never matched by this segment, unless the key is also in Included.
	Excluded []string
	// IncludedContexts contains sets of individually included contexts for specific context kinds.
	//
	// For backward compatibility, the targeting lists are divided up as follows: for the default kind ("user"),
	// we list the keys in Included, but for all other context kinds we use IncludedContexts.
	IncludedContexts []SegmentTarget
	// ExcludedContexts contains sets of individually excluded contexts for specific context kinds.
	//
	// For backward compatibility, the targeting lists are divided up as follows: for the default kind ("user"),
	// we list the keys in Excluded, but for all other context kinds we use ExcludedContexts.
	ExcludedContexts []SegmentTarget
	// Salt is a randomized value assigned to this segment when it is created.
	//
	// The hash function used for calculating percentage rollouts uses this as a salt to ensure that
	// rollouts are consistent within each segment but not predictable from one segment to another.
	Salt string
	// Rules is a list of rules that may match a context.
	//
	// If a context is matched by a Rule, all subsequent Rules in the list are skipped. Rules are ignored
	// if the context's key was matched by Included, Excluded, IncludedContexts, or ExcludedContexts.
	Rules []SegmentRule
	// Unbounded is true if this is a segment whose included/excluded key lists are stored separately and are not
	// limited in size.
	//
	// The name is historical: "unbounded segments" was an earlier name for the product feature that is currently
	// known as "Big Segments". If Unbounded is true, this is a Big Segment.
	Unbounded bool
	// UnboundedContextKind is the context kind associated with the included/excluded key lists if this segment
	// is a Big Segment. If it is empty, we assume ldcontext.DefaultKind. This field is ignored if Unbounded is
	// false.
	//
	// An empty string value here represents the property being unset (so it will be omitted in
	// serialization).
	UnboundedContextKind ldcontext.Kind
	// Version is an integer that is incremented by LaunchDarkly every time the configuration of the segment is
	// changed.
	Version int
	// Generation is an integer that indicates which set of big segment data is currently active for this segment
	// key. LaunchDarkly increments it if a segment is deleted and recreated. This value is only meaningful for big
	// segments. If this field is unset, it means the segment representation used an older schema so the generation
	// is unknown, in which case matching a big segment is not possible.
	Generation ldvalue.OptionalInt
	// Deleted is true if this is not actually a user segment but rather a placeholder (tombstone) for a
	// deleted segment. This is only relevant in data store implementations.
	Deleted bool
	// preprocessedData is created by Segment.Preprocess() to speed up target matching.
	preprocessed segmentPreprocessedData
}

// SegmentTarget describes a target list within a segment, for a specific context kind.
type SegmentTarget struct {
	// ContextKind is the context kind that this target list applies to.
	//
	// LaunchDarkly will normally always set this property, but if it is empty/omitted, it should be
	// treated as ldcontext.DefaultKind. An empty string value here represents the property being unset
	// (so it will be omitted in serialization).
	ContextKind ldcontext.Kind
	// Values is the set of context keys included in this Target.
	Values []string
	// preprocessed is created by PreprocessSegment() to speed up target matching.
	preprocessed targetPreprocessedData
}

// SegmentRule describes a single rule within a segment.
type SegmentRule struct {
	// ID is a randomized identifier assigned to each rule when it is created.
	ID string
	// Clauses is a list of test conditions that make up the rule. These are ANDed: every Clause must
	// match in order for the SegmentRule to match.
	Clauses []Clause
	// Weight, if defined, specifies a percentage rollout in which only a subset of contexts matching this
	// rule are included in the segment. This is specified as an integer from 0 (0%) to 100000 (100%).
	Weight ldvalue.OptionalInt
	// BucketBy specifies which context attribute should be used to distinguish between contexts in a rollout.
	// This property is ignored if Weight is undefined.
	//
	// The default (when BucketBy is empty) is ldattr.KeyAttr, the context's primary key. If you wish to
	// treat contexts with different keys as the same for rollout purposes as long as they have the same
	// "country" attribute, you would set this to "country".
	//
	// Rollouts always take the context's "secondary key" attribute into account as well if there is one.
	//
	// An empty ldattr.Ref{} value here represents the property being unset (so it will be omitted in
	// serialization). That is different from setting it explicitly to "", which is an invalid attribute
	// reference.
	BucketBy ldattr.Ref
	// RolloutContextKind specifies what kind of context the key (or other attribute if BucketBy is set)
	// should be used to get attributes when computing a rollout. This property is ignored if Weight is
	// undefined. If unset, it defaults to ldcontext.DefaultKind.
	//
	// An empty string value here represents the property being unset (so it will be omitted in
	// serialization).
	RolloutContextKind ldcontext.Kind
}
