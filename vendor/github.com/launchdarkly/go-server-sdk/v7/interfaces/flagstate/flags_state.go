package flagstate

import (
	"fmt"

	"github.com/launchdarkly/go-jsonstream/v3/jwriter"
	"github.com/launchdarkly/go-sdk-common/v3/ldreason"
	"github.com/launchdarkly/go-sdk-common/v3/ldtime"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"golang.org/x/exp/maps"
)

// AllFlags is a snapshot of the state of multiple feature flags with regard to a specific evaluation
// context. This is the return type of LDClient.AllFlagsState().
//
// Serializing this object to JSON using json.Marshal() will produce the appropriate data structure for
// bootstrapping the LaunchDarkly JavaScript client.
type AllFlags struct {
	flags map[string]FlagState
	valid bool
}

// AllFlagsBuilder is a builder that creates AllFlags instances. This is normally done only by the SDK, but
// it may also be used in test code.
//
// AllFlagsBuilder methods should not be used concurrently from multiple goroutines.
type AllFlagsBuilder struct {
	state   AllFlags
	options allFlagsOptions
}

type allFlagsOptions struct {
	withReasons          bool
	detailsOnlyIfTracked bool
}

// FlagState represents the state of an individual feature flag, with regard to a specific evaluation
// context, at the time when LDClient.AllFlagsState() was called.
type FlagState struct {
	// Value is the result of evaluating the flag for the specified evaluation context.
	Value ldvalue.Value

	// Variation is the variation index that was selected for the specified evaluation context.
	Variation ldvalue.OptionalInt

	// Version is the flag's version number when it was evaluated. This is an int rather than an OptionalInt
	// because a flag always has a version and nonexistent flag keys are not included in AllFlags.
	Version int

	// Reason is the evaluation reason from evaluating the flag.
	Reason ldreason.EvaluationReason

	// TrackEvents is true if a full feature event must be sent whenever evaluating this flag. This will be
	// true if tracking was explicitly enabled for this flag for data export, or if the evaluation involved
	// an experiment, or both.
	TrackEvents bool

	// TrackReason is true if the evaluation reason should always be included in any full feature event
	// created for this flag, regardless of whether variationDetail was called. This will be true if the
	// evaluation involved an experiment.
	TrackReason bool

	// DebugEventsUntilDate is non-zero if event debugging is enabled for this flag until the specified time.
	DebugEventsUntilDate ldtime.UnixMillisecondTime

	// OmitDetails is true if, based on the options passed to AllFlagsState and the flag state, some of the
	// metadata can be left out of the JSON representation.
	OmitDetails bool
}

// Option is the interface for optional parameters that can be passed to LDClient.AllFlagsState.
type Option interface {
	fmt.Stringer
	apply(*allFlagsOptions)
}

type clientSideOnlyOption struct{}
type withReasonsOption struct{}
type detailsOnlyForTrackedFlagsOption struct{}

// OptionClientSideOnly is an option that can be passed to LDClient.AllFlagsState().
//
// It specifies that only flags marked for use with the client-side SDK should be included in the state
// object. By default, all flags are included.
func OptionClientSideOnly() Option {
	return clientSideOnlyOption{}
}

// OptionWithReasons is an option that can be passed to LDClient.AllFlagsState(). It specifies that
// evaluation reasons should be included in the state object. By default, they are not.
func OptionWithReasons() Option {
	return withReasonsOption{}
}

// OptionDetailsOnlyForTrackedFlags is an option that can be passed to LDClient.AllFlagsState(). It
// specifies that any flag metadata that is normally only used for event generation - such as flag versions
// and evaluation reasons - should be omitted for any flag that does not have event tracking or debugging
// turned on. This reduces the size of the JSON data if you are passing the flag state to the front end.
func OptionDetailsOnlyForTrackedFlags() Option {
	return detailsOnlyForTrackedFlagsOption{}
}

// IsValid returns true if the call to LDClient.AllFlagsState() succeeded. It returns false if there was an
// error (such as the data store not being available), in which case no flag data is in this object.
func (a AllFlags) IsValid() bool {
	return a.valid
}

// GetFlag looks up information for a specific flag by key. The returned FlagState struct contains the flag
// flag evaluation result and flag metadata that was recorded when LDClient.AllFlagsState() was called. The
// second return value is true if successful, or false if there was no such flag.
func (a AllFlags) GetFlag(flagKey string) (FlagState, bool) {
	f, ok := a.flags[flagKey]
	return f, ok
}

// GetValue returns the value of an individual feature flag at the time the state was recorded. The return
// value will be ldvalue.Null() if the flag returned the default value, or if there was no such flag.
//
// This is equivalent to calling GetFlag for the flag and then getting the Value property.
func (a AllFlags) GetValue(flagKey string) ldvalue.Value {
	return a.flags[flagKey].Value
}

// ToValuesMap returns a map of flag keys to flag values. If a flag would have evaluated to the default
// value, its value will be ldvalue.Null().
//
// Do not use this method if you are passing data to the front end to "bootstrap" the JavaScript client.
// Instead, convert the state object to JSON using json.Marshal.
func (a AllFlags) ToValuesMap() map[string]ldvalue.Value {
	ret := make(map[string]ldvalue.Value, len(a.flags))
	for k, v := range a.flags {
		ret[k] = v.Value
	}
	return ret
}

// MarshalJSON implements a custom JSON serialization for AllFlags, to produce the correct data structure
// for "bootstrapping" the LaunchDarkly JavaScript client.
func (a AllFlags) MarshalJSON() ([]byte, error) {
	w := jwriter.NewWriter()
	obj := w.Object()
	obj.Name("$valid").Bool(a.valid)
	for key, flag := range a.flags {
		flag.Value.WriteToJSONWriter(obj.Name(key))
	}
	stateObj := obj.Name("$flagsState").Object()
	for key, flag := range a.flags {
		flagObj := stateObj.Name(key).Object()
		flagObj.Maybe("variation", flag.Variation.IsDefined()).Int(flag.Variation.IntValue())
		flagObj.Maybe("version", !flag.OmitDetails).Int(flag.Version)
		if flag.Reason.IsDefined() && !flag.OmitDetails {
			flag.Reason.WriteToJSONWriter(flagObj.Name("reason"))
		}
		flagObj.Maybe("trackEvents", flag.TrackEvents).Bool(flag.TrackEvents)
		flagObj.Maybe("trackReason", flag.TrackReason).Bool(flag.TrackReason)
		flagObj.Maybe("debugEventsUntilDate", flag.DebugEventsUntilDate > 0).Float64(float64(flag.DebugEventsUntilDate))
		flagObj.End()
	}
	stateObj.End()
	obj.End()
	return w.Bytes(), w.Error()
}

// NewAllFlagsBuilder creates a builder for constructing an AllFlags instance. This is normally done only by
// the SDK, but it may also be used in test code.
func NewAllFlagsBuilder(options ...Option) *AllFlagsBuilder {
	b := &AllFlagsBuilder{
		state: AllFlags{
			flags: make(map[string]FlagState),
			valid: true,
		},
	}
	for _, o := range options {
		o.apply(&b.options)
	}
	return b
}

// Build returns an immutable State instance copied from the current builder data.
func (b *AllFlagsBuilder) Build() AllFlags {
	return AllFlags{valid: b.state.valid, flags: maps.Clone(b.state.flags)}
}

// AddFlag adds information about a flag.
//
// The Reason property in the FlagState may or may not be recorded in the State, depending on the builder
// options.
func (b *AllFlagsBuilder) AddFlag(flagKey string, flag FlagState) *AllFlagsBuilder {
	// To save bandwidth, we include evaluation reasons only if 1. the application explicitly said to
	// include them or 2. they must be included because of experimentation
	if b.options.detailsOnlyIfTracked {
		if !flag.TrackEvents && !flag.TrackReason &&
			!(flag.DebugEventsUntilDate != 0 && flag.DebugEventsUntilDate > ldtime.UnixMillisNow()) {
			flag.OmitDetails = true
		}
	}
	if !b.options.withReasons && !flag.TrackReason {
		flag.Reason = ldreason.EvaluationReason{}
	}
	b.state.flags[flagKey] = flag
	return b
}

func (o clientSideOnlyOption) String() string {
	return "ClientSideOnly"
}

func (o clientSideOnlyOption) apply(options *allFlagsOptions) {
}

func (o withReasonsOption) String() string {
	return "WithReasons"
}

func (o withReasonsOption) apply(options *allFlagsOptions) {
	options.withReasons = true
}

func (o detailsOnlyForTrackedFlagsOption) String() string {
	return "DetailsOnlyForTrackedFlags"
}

func (o detailsOnlyForTrackedFlagsOption) apply(options *allFlagsOptions) {
	options.detailsOnlyIfTracked = true
}
