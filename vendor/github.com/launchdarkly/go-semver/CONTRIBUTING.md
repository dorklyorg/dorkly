# Contributing to this project

## Submitting bug reports and feature requests

The LaunchDarkly SDK team monitors the [issue tracker](https://github.com/launchdarkly/go-semver/issues) in tis repository. Bug reports and feature requests specific to this project should be filed in this issue tracker. The SDK team will respond to all newly filed issues within two business days.
 
## Submitting pull requests
 
We encourage pull requests and other contributions from the community. Before submitting pull requests, ensure that all temporary or unintended code is removed. Don't worry about adding reviewers to the pull request; the LaunchDarkly SDK team will add themselves. The SDK team will acknowledge all pull requests within two business days.
 
## Build instructions
 
### Prerequisites
 
This project should be built against Go 1.13 or newer.

### Building

To build the project without running any tests:
```
make
```

If you wish to clean your working directory between builds, you can clean it by running:
```
make clean
```

To run the linter:
```
make lint
```

### Testing
 
To build and run all unit tests:
```
make test
```

## Coding best practices

### Test coverage

It is important to keep unit test coverage as close to 100% as possible in this project. You can view the latest code coverage report in CircleCI, as `coverage.html` and `coverage.txt` in the artifacts. You can also generate this information locally with `make test-coverage`.

The build will fail if there are any uncovered blocks of code, unless you explicitly add an override by placing a comment that starts with `// COVERAGE` somewhere within that block. Sometimes a gap in coverage is unavoidable, usually because the compiler requires us to provide a code path for some condition that in practice can't happen and can't be tested. Exclude these paths with a `// COVERAGE` comment.

### Avoid heap allocations

A major design goal in this implementation is to maximize performance and avoid unwanted heap churn. No operations in this package should allocate any data on the heap; everything should be passed by value, and slices or maps should not be used. The `make benchmarks` target will fail if any benchmark shows heap allocations.
