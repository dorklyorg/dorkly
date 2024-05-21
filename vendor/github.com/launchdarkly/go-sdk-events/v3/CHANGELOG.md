# Change log

All notable changes to the project will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## [3.2.0] - 2024-03-13
### Changed:
- Redact anonymous attributes within feature events
- Always inline contexts for feature events

## [3.1.0] - 2023-10-23
### Added:
- Add new `ForceSampling` field. This ensures an event is sent regardless of the provided `SamplingRatio`.

## [3.0.0] - 2023-10-11
### Added:
- `EventProcessor` interface now supports recording migration related events.
- Event sampling and summary exclusion controls are now supported.

## [2.0.2] - 2023-05-11
### Fixed:
- Do not queue subsequent events after unrecoverable error in the event processor.
- HTTP status code 413 will no longer trigger an event processor shutdown.

## [2.0.1] - 2023-03-01
### Changed:
- Bumped go-sdk-common to v3.0.1.

## [2.0.0] - 2022-12-01
This major version release of `go-sdk-events` corresponds to the upcoming v6.0.0 release of the LaunchDarkly Go SDK (`go-server-sdk`), and cannot be used with earlier SDK versions. As before, this package is intended for internal use by the Go SDK, and by LaunchDarkly services; other use is unsupported.

### Added:
- `EventProcessor.FlushBlocking`
- `EventProcessor.RecordRawEvent`
- `EventInputContext`
- `NewServerSideEventSender`
- `PreserializedContext`
- `SendEventDataWithRetry`

### Changed:
- The minimum Go version is now 1.18.
- The package now uses a regular import path (`github.com/launchdarkly/go-sdk-events/v2`) rather than a `gopkg.in` path (`gopkg.in/launchdarkly/go-sdk-events.v1`).
- The dependency on `gopkg.in/launchdarkly/go-sdk-common.v2` has been changed to `github.com/launchdarkly/go-sdk-common/v3`.
- Events now use the `ldcontext.Context` type rather than `lduser.User`.
- Private attributes can now be designated with the `ldattr.Ref` type, which allows redaction of either a full attribute or a property within a JSON object value.
- There is a new JSON schema for analytics events. The HTTP headers for event payloads now report the schema version as 4.
- Renamed `FeatureRequestEvent`, `IdentifyEvent`, and `CustomEvent` to `EvaluationData`, `IdentifyEventData`, and `CustomEventData`, to clarify that they are inputs affecting events rather than the events themselves.

### Removed:
- All alias event functionality
- `EventsConfiguration.InlineUsersInEvents`
- `FlagEventProperties`
- `NewDefaultEventSender`

## [1.1.1] - 2021-06-03
### Fixed:
- Updated `go-jsonstream` and `go-sdk-common` dependencies to latest patch versions for JSON parsing fixes. Those patches should not affect `go-sdk-events` since it does not _parse_ JSON, but this ensures that the latest release has the most correct transitive dependencies.

## [1.1.0] - 2021-01-21
### Added:
- Added support for a new analytics event type, &#34;alias&#34;, which will be used in a future version of the SDK.

## [1.0.1] - 2020-12-17
### Changed:
- The library now uses [`go-jsonstream`](https://github.com/launchdarkly/go-jsonstream) for generating JSON output.

## [1.0.0] - 2020-09-18
Initial release of this analytics event support code that will be used with versions 5.0.0 and above of the LaunchDarkly Server-Side SDK for Go.
