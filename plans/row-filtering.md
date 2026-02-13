<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Row Filtering (#20)

## Problem

Users can sort and hide columns but can't filter rows by content. On a table
with 30+ maintenance items or vendors, finding a specific entry means scrolling
manually.

## Design: Pin-and-Filter

A two-step, mechanical filter: **pin** cell values to preview matches, then
**activate** to hide non-matching rows.

### Interaction

1. **Navigate** to a cell whose value you want to filter by (e.g., the "Plan"
   cell in the Status column).
2. **Pin** (`n`): toggles a pin on the current cell's column+value pair.
   Matching rows are immediately highlighted; non-matching rows are dimmed. All
   rows remain visible (preview mode).
3. **Pin more** (optional): navigate to other cells and press `n` again.
   - Multiple pins in the **same column** = OR (e.g., Status = "Plan" OR
     "Active").
   - Pins across **different columns** = AND (e.g., Status = "Plan" AND
     Vendor = "Bob's Plumbing").
4. **Activate** (`N`): commits the filter. Non-matching rows are hidden. Only
   matching rows remain in the table.
5. **Clear** (`N` again while filtered, or `esc` then `N`): removes all pins
   and restores the full row set.

### Visual states

| State | Matching rows | Non-matching rows | Pinned cell values | Status bar |
|-------|--------------|-------------------|--------------------|------------|
| No pins | Normal | Normal | Normal | (nothing) |
| Pinned (preview) | Normal | Dimmed | `muted` foreground | pin indicators |
| Filtered (active) | Normal | Hidden | `muted` foreground | filter active indicator |

### Pin highlighting

- **Pinned cell values**: Every cell in a pinned column whose value matches a
  pin renders with the `muted` foreground color (`#AA4499` light / `#CC79A7`
  dark -- Wong palette mauve). This overrides the cell's normal semantic color
  (money, status, urgency) while pins are active. Original colors return when
  pins are cleared.
- **Status bar**: Shows pinned values in `muted` color, e.g.,
  `Status: Plan, Active · Vendor: Bob's` with `n` pin/unpin and `N` to
  activate/clear.
- **Non-matching rows (preview)**: Rendered with ANSI faint/dim -- visible but
  clearly secondary.
- **Styles**: Add a `Pinned` style to the `Styles` struct:
  `Pinned: lipgloss.NewStyle().Foreground(muted)`

### Key bindings

| Key | Context | Action |
|-----|---------|--------|
| `n` | Normal, no pins | Pin current cell value; enter preview mode |
| `n` | Normal, preview | Toggle pin on current cell value |
| `n` | Normal, filtered | Toggle pin and immediately re-filter |
| `N` | Normal, preview | Activate filter (hide non-matching) |
| `N` | Normal, filtered | Clear all pins, restore full row set |

### Data architecture

New fields on `Tab`:

```go
type filterPin struct {
    Col    int               // index in tab.Specs
    Values map[string]bool   // pinned values (OR / IN semantics within column)
}

type Tab struct {
    // ... existing fields ...

    // Pin-and-filter state.
    Pins         []filterPin  // active pins; AND across columns, OR within
    FilterActive bool         // true = non-matching rows hidden; false = preview only

    // Full data (pre-row-filter). Populated by reloadTab after project status
    // filtering. Row filter operates on these without hitting the DB.
    FullRows     []table.Row
    FullMeta     []rowMeta
    FullCellRows [][]cell
}
```

### Matching logic

```
matchesAllPins(row) = for each pinned column:
    row's cell value (lowercased) IN pinned values set for that column
```

Case-insensitive **exact** match (not substring) because we're pinning
specific cell values, not searching for text fragments. Pin values are stored
in a `map[string]bool` (lowercased keys) for O(1) membership checks -- the
equivalent of a SQL `IN` clause, evaluated client-side on already-loaded data.

### Data flow

```
reloadTab:
  DB load → project status filter → store in Full* → applyRowFilter → sort → viewport

applyRowFilter(tab):
  if no pins or not FilterActive:
    displayed = Full* (but set dim flags for preview)
  else:
    displayed = only matching rows from Full*
```

For the preview (pins exist, filter not active), we don't remove rows -- we
just pass dim metadata to the renderer. This means `tab.Rows`, `tab.CellRows`,
and `tab.Table` contain ALL rows, and rendering checks pin-match to decide
dimming.

For the filtered state, we actually remove non-matching rows from the displayed
set (same pattern as `filterProjectRowsByStatusFlags`).

### Edge cases

- **Pin on empty cell**: Pins the empty string. Matches all rows with empty
  values in that column. Valid but might be surprising -- could show a status
  message "Pinned: (empty)" so it's clear.
- **Pin on hidden column**: Not possible -- `n` operates on `ColCursor` which
  only points to visible columns.
- **Pin + sort**: Sort operates on whatever rows are displayed (filtered or
  full). Natural.
- **Pin + hide column**: If a column with pins is hidden, the pins on that
  column are cleared (the user can't see them anymore). Status message notes it.
- **Pin + tab switch**: Pins and filter state are per-tab and persist across
  tab switches. Switching away and back preserves your filter exactly as you
  left it.
- **Pin + detail drilldown**: Detail tabs have their own `Tab` with independent
  pins.
- **Pin + reload (mutation)**: `reloadTab` refreshes `Full*` then re-applies
  pins/filter. If a pinned value no longer exists in the data, that pin
  effectively matches nothing (AND with other pins may produce empty result).
- **All rows filtered out**: Show "No matches" placeholder, same style as
  "No entries yet."
- **Pin + project status filter (z/a/t)**: Project status filters apply first
  (at the DB/load level). Pins then filter within that subset. They compose
  naturally.

### Implementation order

1. Add `filterPin`, `Pins`, `FilterActive`, `Full*` fields to `Tab`.
2. Implement `matchesAllPins(cellRow []cell, pins []filterPin) bool`.
3. Implement `applyRowFilter(tab)` -- filtered mode removes rows, preview mode
   sets dim flags.
4. Wire into `reloadTab` between project filter and sort.
5. Handle `n` key in `handleNormalKeys`: toggle pin, set preview dim state.
6. Handle `N` key in `handleNormalKeys`: activate/clear filter.
7. Rendering: dim non-matching rows in preview mode (add `Dimmed` flag to
   `rowMeta` or pass pin-match state to `renderRows`).
8. Status bar: show pin indicators in `normalModeStatusHints`.
9. Help overlay: add `n`/`N` to the keybinding reference.
10. Block tab switching (`b`/`f`) when pins exist or filter is active.
11. Clear column pins when hiding that column.
12. Tests: pin single value, pin OR (same column), pin AND (cross-column),
    activate, clear, pin + sort, pin + hide column, pin + reload, pin cleared
    on tab switch.
