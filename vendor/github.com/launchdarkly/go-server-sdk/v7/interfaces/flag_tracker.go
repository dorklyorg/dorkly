package interfaces

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
)

// FlagTracker is an interface for tracking changes in feature flag configurations.
//
// An implementation of this interface is returned by
// [github.com/launchdarkly/go-server-sdk/v7.LDClient.GetFlagTracker]. Application code should not
// implement this interface.
type FlagTracker interface {
	// AddFlagChangeListener subscribes for notifications of feature flag changes in general.
	//
	// The returned channel will receive a new FlagChangeEvent value whenever the SDK receives any change to
	// any feature flag's configuration, or to a user segment that is referenced by a feature flag. If the
	// updated flag is used as a prerequisite for other flags, the SDK assumes that those flags may now
	// behave differently and sends flag change events for them as well.
	//
	// Note that this does not necessarily mean the flag's value has changed for any particular evaluation
	// context, only that some part of the flag configuration was changed so that it may return a different
	// value than it previously returned for some context. If you want to track flag value changes, use
	// AddFlagValueChangeListener instead.
	//
	// Change events only work if the SDK is actually connecting to LaunchDarkly (or using the file data source).
	// If the SDK is only reading flags from a database (ldcomponents.ExternalUpdatesOnly) then it cannot
	// know when there is a change, because flags are read on an as-needed basis.
	//
	// It is the caller's responsibility to consume values from the channel. Allowing values to accumulate in
	// the channel can cause an SDK goroutine to be blocked.
	AddFlagChangeListener() <-chan FlagChangeEvent

	// RemoveFlagChangeListener unsubscribes from notifications of feature flag changes. The specified channel
	// must be one that was previously returned by AddFlagChangeListener(); otherwise, the method has no effect.
	RemoveFlagChangeListener(listener <-chan FlagChangeEvent)

	// AddFlagValueChangeListener subscribes for notifications of changes in a specific feature flag's value
	// for a specific set of context properties.
	//
	// When you call this method, it first immediately evaluates the feature flag. It then starts listening
	// for feature flag configuration changes (using the same mechanism as AddFlagChangeListener), and whenever
	// the specified feature flag changes, it re-evaluates the flag for the same evaluation context. It then
	// pushes a new FlagValueChangeEvent to the channel if and only if the resulting value has changed.
	//
	// All feature flag evaluations require an instance of ldcontext.Context. If the feature flag you are tracking
	// tracking does not have any targeting rules, you must still pass a dummy context such as
	// ldcontext.New("for-global-flags"). If you do not want the context to appear on your dashboard, use
	// the Anonymous property: ldcontext.NewBuilder("for-global-flags").Anonymous(true).Build().
	//
	// The defaultValue parameter is used if the flag cannot be evaluated; it is the same as the corresponding
	// parameter in LDClient.JSONVariation().
	AddFlagValueChangeListener(
		flagKey string,
		context ldcontext.Context,
		defaultValue ldvalue.Value,
	) <-chan FlagValueChangeEvent

	// RemoveFlagValueChangeListener unsubscribes from notifications of feature flag value changes. The
	// specified channel must be one that was previously returned by AddFlagValueChangeListener(); otherwise,
	// the method has no effect.
	RemoveFlagValueChangeListener(listener <-chan FlagValueChangeEvent)
}

// FlagChangeEvent is a parameter type used with FlagTracker.AddFlagChangeListener().
//
// This is not an analytics event to be sent to LaunchDarkly; it is a notification to the application.
type FlagChangeEvent struct {
	// Key is the key of the feature flag whose configuration has changed.
	//
	// The specified flag may have been modified directly, or this may be an indirect change due to a change
	// in some other flag that is a prerequisite for this flag, or a user segment that is referenced in the
	// flag's rules.
	Key string
}

// FlagValueChangeEvent is a parameter type used with FlagTracker.AddFlagValueChangeListener().
//
// This is not an analytics event to be sent to LaunchDarkly; it is a notification to the application.
type FlagValueChangeEvent struct {
	// Key is the key of the feature flag whose configuration has changed.
	//
	// The specified flag may have been modified directly, or this may be an indirect change due to a change
	// in some other flag that is a prerequisite for this flag, or a user segment that is referenced in the
	// flag's rules.
	Key string

	// OldValue is the last known value of the flag for the specified evaluation context prior to the update.
	//
	// Since flag values can be of any JSON data type, this is represented as ldvalue.Value. That type has
	// methods for converting to a primitive Java type such as Value.BoolValue().
	//
	// If the flag did not exist before or could not be evaluated, this will be whatever value was
	// specified as the default with AddFlagValueChangeListener().
	OldValue ldvalue.Value

	// NewValue is the new value of the flag for the specified evaluation context.
	//
	// Since flag values can be of any JSON data type, this is represented as ldvalue.Value. That type has
	// methods for converting to a primitive Java type such Value.BoolValue().
	//
	// If the flag could not be evaluated or was deleted, this will be whatever value was specified as
	// the default with AddFlagValueChangeListener().
	NewValue ldvalue.Value
}
