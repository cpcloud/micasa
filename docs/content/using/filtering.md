+++
title = "Row Filtering"
weight = 3
description = "Pin cell values to filter rows interactively."
linkTitle = "Row Filtering"
+++

micasa lets you filter table rows by pinning specific cell values. The
mechanism is two-step: **pin** values to preview which rows match, then
**activate** to hide the rest.

## Quick start

1. Navigate to a cell whose value you want to filter by (e.g., "Plan" in the
   Status column)
2. Press `n` to pin it -- matching rows stay bright, others dim
3. Press `N` to activate -- non-matching rows disappear
4. Press `N` again to deactivate (rows return, dimming resumes)
5. Press `n` on each pinned cell to unpin, or hide the column (`c`) to clear
   its pins

## Pin logic

- **OR within a column**: pinning "Plan" and "Active" in the Status column
  matches rows with *either* value
- **AND across columns**: pinning Status = "Plan" and Vendor = "Bob's Plumbing"
  matches rows where *both* conditions hold

Matching is case-insensitive and exact (the full cell value, not a substring).

## Visual states

| State | Matching rows | Non-matching rows | Pinned cells |
|-------|--------------|-------------------|--------------|
| No pins | Normal | Normal | Normal |
| Preview (pins, filter off) | Normal | Dimmed | Mauve foreground |
| Active (filter on) | Normal | Hidden | Mauve foreground |

Pinned cell values render in mauve to make it easy to see what you've selected.
Non-matching rows in preview mode are dimmed but still visible so you can verify
the filter before committing.

## Eager filter mode

You can press `N` to arm the filter *before* pinning anything. The status bar
shows a **FILTER** badge. Subsequent `n` presses immediately filter (no
preview step) because the filter is already active.

## Per-tab persistence

Pins and filter state are stored per tab. Switching tabs preserves your filter
exactly as you left it -- switch away to check another tab and come back
without losing your selection.

## Magnitude mode interaction

When magnitude mode (`m`) is active, pins operate on the magnitude value
(e.g., the rounded `log10` representation) rather than the underlying number.
Toggling magnitude mode translates existing pins between representations:

- **Enabling mag mode**: a pin on "$1,250" becomes a pin on the corresponding
  magnitude (e.g., "3")
- **Disabling mag mode**: a magnitude pin expands to all raw values that share
  that magnitude band

This means your filter stays meaningful across display modes without manual
re-pinning.

## Keybindings

| Key | Action |
|-----|--------|
| `n` | Toggle pin on current cell value |
| `N` | Toggle filter activation (preview <-> active) |

## Edge cases

- **Empty cells**: pinning an empty cell matches all rows with empty values in
  that column
- **Hidden columns**: hiding a column with `c` clears any pins on that column
- **Sorting**: sorts apply to whatever rows are visible (filtered or full)
- **Project status toggles** (`z`/`a`/`t`): these filters apply first at the
  data level; pins then filter within that subset
- **All rows filtered**: shows "No matches." instead of the table
