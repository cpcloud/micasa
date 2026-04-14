<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# CLI `status` Command

Design spec for `micasa status` -- a non-TUI command that prints
overdue and upcoming maintenance items, open incidents, and active
projects to stdout and exits non-zero when items need attention.

Addresses [#692](https://github.com/micasa-dev/micasa/issues/692).

## Motivation

The TUI dashboard surfaces overdue/upcoming items, open incidents, and
active projects, but only when the full interactive app is running.
A headless command enables:

- Cron jobs ("email me when maintenance is overdue")
- Shell prompt integration (`PS1` / `starship`)
- Status bar widgets (Polybar, Waybar, i3blocks)
- Desktop notification wrappers (`notify-send`)
- CI-style health checks ("is my house okay?")

## Command

```
micasa status [--json] [--days N] [database-path]
```

### Flags

| Flag       | Default | Description                                       |
|------------|---------|---------------------------------------------------|
| `--json`   | false   | Output JSON instead of human-readable text         |
| `--days N` | 30      | Look-ahead window for upcoming items (calendar days)|

Database path is a positional argument, consistent with `micasa show`,
`micasa backup`, and `micasa pro status`. Defaults to platform default,
honors `MICASA_DB_PATH`.

## Exit codes

| Code | Meaning                                         |
|------|--------------------------------------------------|
| 0    | No attention needed                              |
| 1    | Items need attention, or an error occurred        |

Exit code 1 (attention needed) fires when any of these are true:
- At least one maintenance item is overdue (next due date < today)
- At least one incident is open or in-progress
- At least one project has status "delayed"

Errors (DB not found, query failure) also exit 1, consistent with
how every other `micasa` subcommand behaves -- `main()` calls
`os.Exit(1)` on any error. Callers that need to distinguish
attention-needed from error should use `--json` and check the
`needs_attention` field.

"Upcoming" items (due within `--days` window but not yet overdue) are
informational and do NOT trigger exit code 1 by themselves.

## Data sources

Reuses existing `data.Store` queries from the TUI dashboard:

| Section    | Store method                    | Filter                         |
|------------|----------------------------------|--------------------------------|
| Overdue    | `ListMaintenanceWithSchedule()` | `ComputeNextDue` < today       |
| Upcoming   | `ListMaintenanceWithSchedule()` | today <= `ComputeNextDue` <= today + `--days` |
| Incidents  | `ListOpenIncidents()`           | Status open or in-progress     |
| Projects   | `ListActiveProjects()`          | Status underway or delayed     |

`ComputeNextDue` and `dateDiffDays` already exist. `dateDiffDays` is
currently in `internal/app/table.go` -- it will be moved to
`internal/data/` so both the TUI and CLI can use it without importing
the `app` package.

## Text output format

Sections are printed only when non-empty. Each section has a header
and a tab-aligned table. Example:

```
=== OVERDUE ===
NAME                  OVERDUE
Replace HVAC filter   15d
Clean gutters         3d

=== UPCOMING ===
NAME                  DUE
Inspect roof          12d
Service water heater  28d

=== INCIDENTS ===
TITLE                 SEVERITY  REPORTED
Leaking faucet        urgent    2d
Garage door stuck     soon      5d

=== ACTIVE PROJECTS ===
TITLE                 STATUS    STARTED
Kitchen remodel       delayed   3mo
Fence repair          underway  14d
```

Duration formatting reuses the existing `shortDur` / `daysText` helpers,
which will be moved to `internal/data/` alongside `dateDiffDays`.

## JSON output format

```json
{
  "overdue": [
    {
      "id": "01JQ...",
      "name": "Replace HVAC filter",
      "category": "HVAC",
      "appliance": "Central AC",
      "next_due": "2026-04-02",
      "days_overdue": 15
    }
  ],
  "upcoming": [
    {
      "id": "01JQ...",
      "name": "Inspect roof",
      "category": "Exterior",
      "appliance": "",
      "next_due": "2026-04-26",
      "days_until_due": 12
    }
  ],
  "incidents": [
    {
      "id": "01JQ...",
      "title": "Leaking faucet",
      "status": "open",
      "severity": "urgent",
      "date_noticed": "2026-04-12"
    }
  ],
  "active_projects": [
    {
      "id": "01JQ...",
      "title": "Kitchen remodel",
      "status": "delayed",
      "start_date": "2026-01-15"
    }
  ],
  "needs_attention": true
}
```

The top-level `needs_attention` boolean mirrors exit code semantics:
true when exit code would be 1. Empty arrays are included (not omitted)
for predictable `jq` usage.

## Implementation plan

### Move shared helpers to `internal/data/`

Move `dateDiffDays`, `shortDur`, `daysText` from `internal/app/` to
`internal/data/`. These are pure functions with no TUI dependencies.
Update `internal/app/` call sites to use the moved versions.

### New file: `cmd/micasa/status.go`

- `newStatusCmd()` -- cobra command, registers `--json` and `--days`
- `statusOpts` struct with `asJSON bool`, `days int`
- `runStatus(w io.Writer, opts *statusOpts, dbPath string) error`
- Registered in `newRootCmd()` via `root.AddCommand(newStatusCmd())`

### Exit code plumbing

`runStatus` returns a typed `exitError{code int}` when items need
attention. `fang.Execute` would normally print this via its error
handler (which calls `err.Error()` on stderr). To suppress output
for the sentinel, register `fang.WithErrorHandler` in `main()` that
skips printing when the error is an `exitError`. `main()` then
extracts the exit code from the error and calls `os.Exit` with it.

For JSON mode, `needs_attention` is embedded in the output so
callers can programmatically distinguish attention-needed from
errors without relying on exit codes.

### Tests

- `cmd/micasa/status_test.go` with seeded demo data
- Verify text output format (sections present/absent)
- Verify JSON output structure (roundtrip via `json.Unmarshal`)
- Verify exit code 0 when nothing overdue
- Verify exit code 1 when items are overdue
- Verify `--days` flag controls upcoming window

## Relationship to other work

- **#920 (CLI CRUD)**: Independent. Status command is read-only and uses
  existing store queries. No conflict.
- **#453 (non-TUI commands)**: Status command is one step toward the
  broader non-TUI CLI surface. Uses the same patterns as `micasa show`.
- **TUI dashboard**: Shares data loading logic but not rendering. The
  `loadDashboardAt` method in `internal/app/` has TUI-specific state
  management (nav entries, cursor, scroll) that is not reused here.

## Non-goals

- No color output (plain text for maximum portability).
- No `--watch` / continuous mode (use `watch micasa status` instead).
- No warranty or insurance sections (keep scope tight; add later if
  requested).
- No spending summary (orthogonal to "needs attention" semantics).
