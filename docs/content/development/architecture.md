+++
title = "Architecture"
weight = 2
description = "How micasa is built: Bubble Tea, TabHandler, overlays."
linkTitle = "Architecture"
+++

micasa is a [Bubble Tea](https://github.com/charmbracelet/bubbletea)
application following The Elm Architecture (TEA): Model, Update, View.

## Package layout

```
cmd/micasa/          CLI entry point (kong argument parsing)
internal/
  app/               Bubble Tea application layer
    model.go         Model struct, Init, Update, key dispatch
    types.go         Mode, Tab, cell, columnSpec, etc.
    handlers.go      TabHandler interface + entity implementations
    tables.go        Column specs, row builders, table construction
    forms.go         Form builders, validators, submit logic
    styles.go        Wong colorblind-safe palette, all lipgloss styles
    view.go          Main View() assembly, overlays
    table.go         Table rendering (headers, rows, viewport)
    collapse.go      Hidden column badges
    house.go         House profile rendering
    dashboard.go     Dashboard data loading + view
    sort.go          Multi-column sort logic
    undo.go          Undo/redo stack
    form_select.go   Select field ordinal jumping
  data/              Data access layer
    models.go        GORM models (HouseProfile, Project, etc.)
    store.go         Store struct, CRUD methods, queries
    dashboard.go     Dashboard-specific queries
    path.go          DB path resolution (XDG)
    validation.go    Parsing helpers (dates, money, ints)
```

## Key design decisions

### TabHandler interface

Entity-specific operations (load, delete, add form, edit form, inline edit,
submit, snapshot, etc.) are encapsulated in the `TabHandler` interface.
Each entity type (projects, quotes, maintenance, appliances) implements this
interface as a stateless struct.

This eliminates scattered `switch tab.Kind` dispatch. Adding a new entity type
means implementing one interface -- no shotgun surgery across the codebase.

Detail views (service log, appliance maintenance) also implement `TabHandler`,
so they get all the same capabilities (add, edit, delete, sort, undo) for
free.

### Modal key handling

micasa uses three modes: Normal, Edit, and Form. The key dispatch chain in
`Update()` is:

1. Window resize handling
2. `ctrl+c` always quits
3. Help overlay intercepts `esc`/`?` when open
4. Form mode delegates to `huh` form library
5. Dashboard intercepts nav keys when visible
6. Common keys (shared by Normal and Edit)
7. Mode-specific keys

The `bubbles/table` widget has its own vim keybindings. In Edit mode, `d` and
`u` are stripped from the table's KeyMap so they can be used for delete/undo
without conflicting with half-page navigation.

### Effective tab

The `effectiveTab()` method returns the detail tab when a detail view is open,
or the main active tab otherwise. All interaction code uses this method, so
detail views work identically to top-level tabs.

### Cell-based rendering

Table cells carry type information (`cellKind`): text, money, date, status,
readonly, drilldown. The renderer uses this to apply per-kind styling (green
for money, colored for status, accent for drilldown). Sort comparators are
also kind-aware.

### Colorblind-safe palette

All colors use the Wong palette with `lipgloss.AdaptiveColor{Light, Dark}`
variants, so the UI works on both dark and light terminal backgrounds. Color
roles are defined in `styles.go`.

## Data flow

```
User keystroke
  -> tea.KeyMsg
  -> Model.Update()
  -> key dispatch (mode-aware)
  -> data mutation (Store CRUD)
  -> reloadAll() (refreshes tabs, detail, dashboard)
  -> Model.View()
  -> rendered string to terminal
```

All data mutations go through the Store, which uses GORM for SQLite access.
After any mutation, `reloadAll()` refreshes lookups, house profile, all tabs,
the detail tab (if open), and the dashboard (if visible).

## Overlays

Dashboard and help are rendered as overlays using
[bubbletea-overlay](https://github.com/rmhubbert/bubbletea-overlay). They
composite on top of the live table view. Help can stack on top of the
dashboard (dimming the layer beneath).
