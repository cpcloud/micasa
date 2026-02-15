+++
title = "Testing"
weight = 3
description = "How to run and write tests."
linkTitle = "Testing"
+++

## Running tests

Always run all tests from the repo root with shuffle enabled:

```sh
go test -shuffle=on -v ./...
```

The `-shuffle=on` flag randomizes test execution order to catch accidental
order dependencies. Go picks and prints the seed automatically.

## Test organization

Tests live alongside the code they test in `*_test.go` files.

## Test philosophy

- **Black-box testing**: tests interact with exported behavior, not
  implementation details. They create a Model, send key messages, and assert
  on the resulting state or view output.
- **In-memory database**: data-layer tests use `:memory:` SQLite databases for
  speed and isolation.
- **No test order dependencies**: `-shuffle=on` ensures this.

## Writing tests

When adding a new feature:

1. Add data-layer tests if you touched Store methods
2. Add app-layer tests for key handling, state transitions, and view output
3. Use the existing test helpers (`newTestModel`, `newTestStore`, etc.)
4. Don't poke into unexported fields -- test through the public interface

## CI

Tests run in CI on every push to `main` and on pull requests, across Linux,
macOS, and Windows. The CI matrix uses `-shuffle=on` to match local behavior.
Pre-commit hooks catch formatting and lint issues before they reach CI.
