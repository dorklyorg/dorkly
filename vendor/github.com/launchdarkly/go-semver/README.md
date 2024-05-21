# LaunchDarkly Semantic Version Package

[![Circle CI](https://circleci.com/gh/launchdarkly/go-semver.svg?style=shield)](https://circleci.com/gh/launchdarkly/go-semver) [![Documentation](https://godoc.org/github.com/launchdarkly/go-semver?status.svg)](https://godoc.org/github.com/launchdarkly/go-semver)

## Overview

This Go package implements parsing and comparison of semantic version (semver) strings, as defined by the [Semantic Versioning 2.0.0 specification](https://semver.org/).

Several semver implementations exist for Go. This implementation was designed for high performance in applications where semver operations may be done frequently, such as in the [LaunchDarkly Go SDK](https://github.com/launchdarkly/go-server-sdk). To that end, it does not use regular expressions and it never allocates data on the heap.

It does not include any additional functionality beyond what is defined in the Semantic Versioning 2.0.0 specification, such as comparison against range/wildcard expressions like ">=1.0.0" or "2.5.x".

This package has no external dependencies other than the regular Go runtime.

## Supported Go versions

This version of the project has been tested with Go 1.14 and above.

## Contributing

We encourage pull requests and other contributions from the community. Check out our [contributing guidelines](CONTRIBUTING.md) for instructions on how to contribute to this project.
