<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# mi casa

A terminal UI for tracking everything about your home -- projects, quotes,
maintenance schedules, appliances, and service history. All in a single
SQLite file that lives on your machine.

```
      ▄▓▄
    ▄▓▓▓▓▓▄
  ▄▓▓▓▓▓▓▓▓▓▄
  ██ ░░ ░░ ██
  ██  ████ ██
  ██  █  █ ██
  ▀▀▀▀▀▀▀▀▀▀▀
```

<!-- TODO: Replace with a real screenshot or asciinema recording -->
<!-- ![screenshot](docs/screenshot.png) -->

## Why

Spreadsheets get messy. Notion is overkill. You just want to know when the
furnace filter was last changed, how much the roof quotes came in at, and
whether the dishwasher is still under warranty.

mi casa keeps it all in one place with vim-style navigation, zero cloud
dependencies, and a data file you can back up with `cp`.

## Features

- **House profile** -- structure details, utilities, insurance, taxes, HOA, all
  on one screen with a retro pixel-art house
- **Projects** -- track home improvement work from ideating through completion
  with color-coded statuses
- **Quotes** -- compare vendor quotes per project with cost breakdowns
  (labor, materials, other)
- **Maintenance** -- recurring tasks with intervals, due dates, and
  linked appliances
- **Appliances** -- warranty tracking, purchase dates, costs, and a drilldown
  into related maintenance items
- **Service log** -- time-ordered history of maintenance events per item,
  with vendor or self-performed tracking
- **Drilldown detail views** -- enter on pill-badge columns opens a scoped
  sub-table with full editing, sorting, and undo support
- **Cross-tab FK navigation** -- linked columns show relationship indicators
  (m:1); enter follows the link to the target row
- **Vim-style modal editing** -- Normal mode for navigation, Edit mode for
  mutations; keybindings adapt per mode
- **Multi-column sorting** -- `s` cycles asc/desc/none per column with
  priority numbers; type-aware comparators for dates, money, and text
- **Inline cell editing** -- edit individual cells without opening a full form
- **Undo/redo** -- multi-level undo (`u`) and redo (`r`) for all edits
- **Colorblind-safe** -- full Wong palette with adaptive light/dark detection
- **Single SQLite file** -- no server, no config, just `~/.local/share/micasa/micasa.db`

## Install

### From source

Requires Go 1.24+ and CGO (for SQLite).

```sh
go install github.com/micasa/micasa/cmd/micasa@latest
```

### Build locally

```sh
git clone https://github.com/micasa/micasa.git
cd micasa
go build -o micasa ./cmd/micasa
```

## Quick start

```sh
# Launch with sample data to explore the UI
micasa --demo

# Or start fresh with your own data
micasa
```

Data is stored at `~/.local/share/micasa/micasa.db` by default. Override with
a positional argument or `MICASA_DB_PATH` environment variable.

## Keybindings

### Normal mode

| Key | Action |
|-----|--------|
| `j` / `k` | Row up / down |
| `h` / `l` | Column left / right |
| `g` / `G` | First / last row |
| `d` / `u` | Half-page down / up |
| `tab` | Next tab |
| `shift+tab` | Previous tab |
| `s` | Cycle sort on current column |
| `S` | Clear all sorts |
| `enter` | Drilldown or follow FK link |
| `i` | Enter Edit mode |
| `H` | Toggle house profile |
| `?` | Help overlay |
| `q` | Quit |

### Edit mode

| Key | Action |
|-----|--------|
| `a` | Add new entry |
| `e` | Edit cell (full form on ID column) |
| `d` | Delete / restore toggle |
| `x` | Show / hide deleted items |
| `p` | Edit house profile |
| `u` | Undo |
| `r` | Redo |
| `1`-`9` | Jump to Nth option in select fields |
| `esc` | Return to Normal mode |

## Architecture

```
cmd/micasa/          CLI entrypoint, arg parsing
internal/
  app/               Bubble Tea model, views, forms, handlers
    model.go         Core update loop, key dispatch, detail views
    handlers.go      TabHandler interface + per-entity implementations
    view.go          All rendering: house, tabs, tables, status bar
    forms.go         huh-based form builders and submit logic
    tables.go        Column specs, row builders, cell types
    sort.go          Multi-column sort engine
    undo.go          Undo/redo stack
    styles.go        Wong colorblind-safe palette, all lipgloss styles
  data/
    models.go        GORM models (House, Project, Quote, Maintenance, ...)
    store.go         CRUD operations, migrations, seed data
    validation.go    Shared parsing and formatting helpers
```

**Key design decisions:**

- **TabHandler interface** -- each entity type implements a 10-method interface,
  so adding a new data type requires zero changes to the core dispatch logic
- **effectiveTab()** -- detail views are just another Tab; all table operations
  (sort, edit, undo) work transparently through this single indirection
- **Pill badge drilldown** -- count columns with accent-colored badges signal
  interactive drill-in; the same `cellDrilldown` kind drives both rendering
  and key dispatch

## Tech stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) -- terminal UI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) -- table component
- [huh](https://github.com/charmbracelet/huh) -- form framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) -- styling
- [GORM](https://gorm.io) + SQLite -- data layer

## Contributing

PRs welcome. Run `go test ./...` before submitting. The project uses
`pre-commit` hooks for formatting (`golines`), linting (`golangci-lint`),
and tests.

## License

Apache-2.0 -- see [LICENSE](LICENSE) for details.
