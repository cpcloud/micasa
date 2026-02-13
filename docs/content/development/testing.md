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
| `internal/app/handler_crud_test.go` | Handler CRUD operations |
| `internal/app/detail_test.go` | Drilldown stack push/pop, nested drilldowns, breadcrumbs, vendor/project drilldowns |
| `internal/app/dashboard_test.go` | Dashboard navigation and view content |
| `internal/app/dashboard_load_test.go` | Dashboard data loading |
| `internal/app/dashboard_rows_test.go` | Dashboard row building |
| `internal/app/view_test.go` | View rendering, line clamping, viewport |
| `internal/app/undo_test.go` | Undo/redo stack, cross-stack snapshotting |
| `internal/app/form_select_test.go` | Select field ordinal jumping |
| `internal/app/form_validators_test.go` | Form validation helpers |
| `internal/app/lighter_forms_test.go` | Lighter-weight add forms |
| `internal/app/inline_edit_dispatch_test.go` | Inline edit column dispatch |
| `internal/app/inline_input_test.go` | Inline text input editing |
| `internal/app/calendar_test.go` | Date picker overlay |
| `internal/app/column_finder_test.go` | Fuzzy column finder |
| `internal/app/mag_test.go` | Mag mode (order-of-magnitude easter egg) |
| `internal/app/notes_test.go` | Note preview overlay |
| `internal/app/vendor_test.go` | Vendor tab operations |
| `internal/app/rows_test.go` | Row building helpers |
| `internal/app/compact_test.go` | Compact intervals, status abbreviation, compact money |
| `internal/app/demo_data_test.go` | Demo data seeding |
| `internal/app/model_with_store_test.go` | Model integration with store |
| `internal/app/model_with_demo_data_test.go` | Model with demo data |
| `internal/data/store_test.go` | CRUD operations, queries |
| `internal/data/dashboard_test.go` | Dashboard-specific queries |
| `internal/data/validation_test.go` | Parsing helpers |
| `internal/data/validate_path_test.go` | Database path validation |
| `internal/data/vendor_upsert_test.go` | Vendor upsert logic |
| `internal/data/seed_demo_test.go` | Demo data seeding |
| `internal/data/settings_test.go` | Settings and chat history storage |
| `internal/data/settings_integration_test.go` | Cross-session persistence |
| `internal/llm/client_test.go` | LLM client HTTP interactions, error parsing |
| `internal/llm/prompt_test.go` | Prompt building, date/context injection |
| `internal/llm/sqlfmt_test.go` | SQL pretty-printer |

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
