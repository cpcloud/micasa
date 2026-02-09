+++
title = "Keybindings"
weight = 1
description = "Complete reference of every keybinding."
linkTitle = "Keybindings"
+++

Complete reference of every keybinding in micasa, organized by mode.

## Global (all modes)

| Key      | Action |
|----------|--------|
| `ctrl+c` | Force quit (exit code 130) |

## Normal mode

### Movement

| Key             | Action |
|-----------------|--------|
| `j` / `down`    | Move row down |
| `k` / `up`      | Move row up |
| `h` / `left`    | Move column left (skips hidden columns) |
| `l` / `right`   | Move column right (skips hidden columns) |
| `g`             | Jump to first row |
| `G`             | Jump to last row |
| `d` / `ctrl+d`  | Half-page down |
| `u` / `ctrl+u`  | Half-page up |
| `space`         | Page down |
| `b`             | Page up |

### Tabs and views

| Key             | Action |
|-----------------|--------|
| `tab`           | Next tab (dismisses dashboard if open) |
| `shift+tab`     | Previous tab |
| `H`             | Toggle house profile display |
| `D`             | Toggle dashboard overlay |

### Table operations

| Key | Action |
|-----|--------|
| `s` | Cycle sort on current column (none -> asc -> desc -> none) |
| `S` | Clear all sorts |
| `c` | Hide current column |
| `C` | Show all hidden columns |

### Actions

| Key     | Action |
|---------|--------|
| `enter` | Drilldown into detail view, or follow FK link |
| `i`     | Enter Edit mode |
| `?`     | Open help overlay |
| `q`     | Quit (exit code 0) |
| `esc`   | Close detail view, or clear status message |

## Edit mode

### Movement

Same as Normal mode, except `d` and `u` are rebound:

| Key            | Action |
|----------------|--------|
| `j`/`k`/`h`/`l`/`g`/`G` | Same as Normal |
| `ctrl+d`       | Half-page down |
| `ctrl+u`       | Half-page up |

### Data operations

| Key   | Action |
|-------|--------|
| `a`   | Add new entry to current tab |
| `e`   | Edit current cell inline, or full form if cell is read-only |
| `d`   | Toggle delete/restore on selected row |
| `x`   | Toggle visibility of soft-deleted rows |
| `p`   | Edit house profile |
| `u`   | Undo last edit |
| `r`   | Redo undone edit |
| `esc` | Return to Normal mode |

## Form mode

| Key       | Action |
|-----------|--------|
| `tab`     | Next field |
| `shift+tab` | Previous field |
| `ctrl+s`  | Save form |
| `esc`     | Cancel form (return to previous mode) |
| `1`-`9`   | Jump to Nth option in a select field |

## Dashboard

When the dashboard overlay is open:

| Key       | Action |
|-----------|--------|
| `j`/`k`   | Move cursor down/up through items |
| `g`/`G`   | Jump to first/last item |
| `enter`   | Jump to highlighted item in its tab |
| `D`       | Close dashboard |
| `tab`     | Dismiss dashboard and switch tab |
| `?`       | Open help overlay (stacks on dashboard) |
| `q`       | Quit |

## Help overlay

| Key       | Action |
|-----------|--------|
| `esc`     | Close help |
| `?`       | Close help |
