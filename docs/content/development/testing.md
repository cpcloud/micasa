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

Tests live alongside the code they test:

| File | Tests |
|------|-------|
| `internal/app/mode_test.go` | Mode transitions, key dispatch, KeyMap switching |
| `internal/app/sort_test.go` | Sort cycling, comparators, multi-column ordering |
| `internal/app/handlers_test.go` | TabHandler implementations |
| `internal/app/detail_test.go` | Detail view open/close, breadcrumbs, scoping |
| `internal/app/dashboard_test.go` | Dashboard data loading, navigation, view content |
| `internal/app/view_test.go` | View rendering, line clamping, viewport |
| `internal/app/undo_test.go` | Undo/redo stack, cross-stack snapshotting |
| `internal/app/form_select_test.go` | Select field ordinal jumping |
| `internal/data/store_test.go` | CRUD operations, queries |
| `internal/data/dashboard_test.go` | Dashboard-specific queries |
| `internal/data/validation_test.go` | Parsing helpers |

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
