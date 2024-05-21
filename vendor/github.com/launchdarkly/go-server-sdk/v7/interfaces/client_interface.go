package interfaces

import (
	"time"

	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	"github.com/launchdarkly/go-sdk-common/v3/ldmigration"
	"github.com/launchdarkly/go-sdk-common/v3/ldreason"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	ldevents "github.com/launchdarkly/go-sdk-events/v3"
	"github.com/launchdarkly/go-server-sdk/v7/interfaces/flagstate"
)

// LDClientEvaluations defines the basic feature flag evaluation methods implemented by LDClient.
type LDClientEvaluations interface {
	// BoolVariation returns the value of a boolean feature flag for a given evaluation context.
	//
	// Returns defaultVal if there is an error, if the flag doesn't exist, or the feature is turned off and
	// has no off variation.
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/evaluating#go
	BoolVariation(key string, context ldcontext.Context, defaultVal bool) (bool, error)

	// BoolVariationDetail is the same as [LDClientEvaluation.BoolVariation], but also returns further
	// information about how the value was calculated. The "reason" data will also be included in analytics events.
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/evaluation-reasons#go
	BoolVariationDetail(key string, context ldcontext.Context, defaultVal bool) (bool, ldreason.EvaluationDetail, error)

	// IntVariation returns the value of a feature flag (whose variations are integers) for the given evaluation
	// context.
	//
	// Returns defaultVal if there is an error, if the flag doesn't exist, or the feature is turned off and
	// has no off variation.
	//
	// If the flag variation has a numeric value that is not an integer, it is rounded toward zero (truncated).
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/evaluating#go
	IntVariation(key string, context ldcontext.Context, defaultVal int) (int, error)

	// IntVariationDetail is the same as [LDClientEvaluation.IntVariation], but also returns further information about how
	// the value was calculated. The "reason" data will also be included in analytics events.
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/evaluation-reasons#go
	IntVariationDetail(key string, context ldcontext.Context, defaultVal int) (int, ldreason.EvaluationDetail, error)

	// Float64Variation returns the value of a feature flag (whose variations are floats) for the given evaluation
	// context.
	//
	// Returns defaultVal if there is an error, if the flag doesn't exist, or the feature is turned off and
	// has no off variation.
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/evaluating#go
	Float64Variation(key string, context ldcontext.Context, defaultVal float64) (float64, error)

	// Float64VariationDetail is the same as [LDClientEvaluation.Float64Variation], but also returns further
	// information about how the value was calculated. The "reason" data will also be included in analytics events.
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/evaluation-reasons#go
	Float64VariationDetail(
		key string,
		context ldcontext.Context,
		defaultVal float64,
	) (float64, ldreason.EvaluationDetail, error)

	// StringVariation returns the value of a feature flag (whose variations are strings) for the given evaluation
	// context.
	//
	// Returns defaultVal if there is an error, if the flag doesn't exist, or the feature is turned off and has
	// no off variation.
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/evaluating#go
	StringVariation(key string, context ldcontext.Context, defaultVal string) (string, error)

	// StringVariationDetail is the same as [LDClientEvaluation.StringVariation], but also returns further
	// information about how the value was calculated. The "reason" data will also be included in analytics events.
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/evaluation-reasons#go
	StringVariationDetail(
		key string,
		context ldcontext.Context,
		defaultVal string,
	) (string, ldreason.EvaluationDetail, error)

	// MigrationVariation returns the migration stage of the migration feature flag for the given evaluation context.
	//
	// Returns defaultStage if there is an error or if the flag doesn't exist.
	MigrationVariation(
		key string,
		context ldcontext.Context,
		defaultStage ldmigration.Stage,
	) (ldmigration.Stage, LDMigrationOpTracker, error)

	// JSONVariation returns the value of a feature flag for the given evaluation context, allowing the value to
	// be of any JSON type.
	//
	// The value is returned as an ldvalue.Value, which can be inspected or converted to other types using
	// Value methods such as GetType() and BoolValue(). The defaultVal parameter also uses this type. For
	// instance, if the values for this flag are JSON arrays:
	//
	//     defaultValAsArray := ldvalue.BuildArray().
	//         Add(ldvalue.String("defaultFirstItem")).
	//         Add(ldvalue.String("defaultSecondItem")).
	//         Build()
	//     result, err := client.JSONVariation(flagKey, context, defaultValAsArray)
	//     firstItemAsString := result.GetByIndex(0).StringValue() // "defaultFirstItem", etc.
	//
	// You can also use unparsed json.RawMessage values:
	//
	//     defaultValAsRawJSON := ldvalue.Raw(json.RawMessage(`{"things":[1,2,3]}`))
	//     result, err := client.JSONVariation(flagKey, context, defaultValAsJSON
	//     resultAsRawJSON := result.AsRaw()
	//
	// Returns defaultVal if there is an error, if the flag doesn't exist, or the feature is turned off.
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/evaluating#go
	JSONVariation(key string, context ldcontext.Context, defaultVal ldvalue.Value) (ldvalue.Value, error)

	// JSONVariationDetail is the same as [LDClientEvaluation.JSONVariation], but also returns further
	// information about how the value was calculated. The "reason" data will also be included in analytics events.
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/evaluation-reasons#go
	JSONVariationDetail(key string, context ldcontext.Context, defaultVal ldvalue.Value) (
		ldvalue.Value, ldreason.EvaluationDetail, error)

	// AllFlagsState returns an object that encapsulates the state of all feature flags for a given evaluation
	// context.
	// This includes the flag values, and also metadata that can be used on the front end.
	//
	// The most common use case for this method is to bootstrap a set of client-side feature flags from a
	// back-end service.
	//
	// You may pass any combination of flagstate.ClientSideOnly, flagstate.WithReasons, and
	// flagstate.DetailsOnlyForTrackedFlags as optional parameters to control what data is included.
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/all-flags#go
	AllFlagsState(context ldcontext.Context, options ...flagstate.Option) flagstate.AllFlags
}

// LDClientEvents defines the methods implemented by LDClient that are specifically for generating
// analytics events. Events may also be generated as a side effect of the methods in LDClientEvaluations.
type LDClientEvents interface {
	// Identify reports details about an evaluation context.
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/identify#go
	Identify(context ldcontext.Context) error

	// TrackEvent reports an event associated with an evaluation context.
	//
	// The eventName parameter is defined by the application and will be shown in analytics reports;
	// it normally corresponds to the event name of a metric that you have created through the
	// LaunchDarkly dashboard. If you want to associate additional data with this event, use\
	// [LDClientEvents.TrackData] or [LDClientEvents.TrackMetric].
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/events#go
	TrackEvent(eventName string, context ldcontext.Context) error

	// TrackData reports an event associated with an evaluation context, and adds custom data.
	//
	// The eventName parameter is defined by the application and will be shown in analytics reports;
	// it normally corresponds to the event name of a metric that you have created through the
	// LaunchDarkly dashboard.
	//
	// The data parameter is a value of any JSON type, represented with the ldvalue.Value type, that
	// will be sent with the event. If no such value is needed, use [ldvalue.Null]() (or call
	// [LDClientEvents.TrackEvent] instead). To send a numeric value for experimentation, use
	// [LDClientEvents.TrackMetric].
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/events#go
	TrackData(eventName string, context ldcontext.Context, data ldvalue.Value) error

	// TrackMetric reports an event associated with an evaluation context, and adds a numeric value.
	// This value is used by the LaunchDarkly experimentation feature in numeric custom metrics, and will also
	// be returned as part of the custom event for Data Export.
	//
	// The eventName parameter is defined by the application and will be shown in analytics reports;
	// it normally corresponds to the event name of a metric that you have created through the
	// LaunchDarkly dashboard.
	//
	// The data parameter is a value of any JSON type, represented with the [ldvalue.Value] type, that
	// will be sent with the event. If no such value is needed, use [ldvalue.Null]().
	//
	// For more information, see the Reference Guide: https://docs.launchdarkly.com/sdk/features/events#go
	TrackMetric(eventName string, context ldcontext.Context, metricValue float64, data ldvalue.Value) error

	// TrackMigrationOp reports a migration operation event.
	//
	// The measurements included in the event are used by LaunchDarkly to enhance support and
	// visibility during migration-assisted technology migrations.
	//
	// Migration operation events can be created with a [LDMigrationOpTracker].
	TrackMigrationOp(event ldevents.MigrationOpEventData) error
}

// LDClientInterface defines the basic SDK client operations implemented by LDClient.
//
// This includes all methods for evaluating a feature flag or generating analytics events, as defined by
// LDEvaluations and LDEvents. It does not include general control operations like Flush(), Close(), or
// GetDataSourceStatusProvider().
type LDClientInterface interface {
	LDClientEvaluations
	LDClientEvents

	// WithEventsDisabled returns a decorator for the client that implements the same basic operations
	// but will not generate any analytics events.
	//
	// If events were already disabled, this is just the same object. Otherwise, it is an object whose
	// Variation methods use the same LDClient to evaluate feature flags, but without generating any
	// events, and whose Identify/Track/Custom methods do nothing. Neither evaluation counts nor context
	// properties will be sent to LaunchDarkly for any operations done with this object.
	//
	// You can use this to suppress events within some particular area of your code where you do not want
	// evaluations to affect your dashboard statistics, or do not want to incur the overhead of processing
	// the events.
	//
	// Note that if the original client configuration already had events disabled
	// (config.Events = ldcomponents.NoEvents()), you cannot re-enable them with this method. It is only
	// useful for temporarily disabling events on a client that had them enabled, or re-enabling them on
	// an LDClientInterface that was the result of WithEventsDisabled(true).
	//
	//     // Assuming you did not disable events when creating the client,
	//     // this evaluation generates an event:
	//     value, err := client.BoolVariation("flagkey1", context, false)
	//
	//     // Now we want to do some evaluations without events
	//     tempClient := client.WithEventsDisabled(true)
	//     value, err = tempClient.BoolVariation("flagkey2", context, false)
	//     value, err = tempClient.BoolVariation("flagkey3", context, false)
	WithEventsDisabled(eventsDisabled bool) LDClientInterface
}

// LDMigrationOpTracker defines the required operations implemented by [MigrationOpTracker].
//
// These methods allow incrementally constructing a migration operation event for later reporting to
// LaunchDarkly through the [LDClientEvents.TrackMigrationOp] method.
type LDMigrationOpTracker interface {
	// Operation sets the migration related operation associated with these tracking measurements.
	Operation(op ldmigration.Operation)

	// TrackInvoked allows recording which origins were called during a migration.
	TrackInvoked(origin ldmigration.Origin)

	// TrackConsistency allows recording the results of a consistency check.
	TrackConsistency(isConsistent func() bool)

	// TrackError allows recording whether an error occurred during the operation.
	TrackError(origin ldmigration.Origin)

	// TrackLatency allows tracking the recorded latency for an individual operation.
	TrackLatency(origin ldmigration.Origin, duration time.Duration)

	// Build creates an instance of [ldevents.MigrationOpEventData]. This event data can be provided to
	// the [LDClientEvents.TrackMigrationOp] method to rely this metric information upstream to LaunchDarkly
	// services.
	Build() (*ldevents.MigrationOpEventData, error)
}

// LDLoggers represents an interface that provides basic logging reporting methods.
type LDLoggers interface {
	Debug(values ...interface{})

	Debugf(format string, values ...interface{})

	Error(values ...interface{})

	Errorf(format string, values ...interface{})

	Warn(values ...interface{})

	Warnf(format string, values ...interface{})
}
