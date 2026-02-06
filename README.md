<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# `micasa`

A terminal UI for tracking everything about your home. Single SQLite file. No cloud. No account. No subscriptions. Just your house.

```
      ▄▓▄
    ▄▓▓▓▓▓▄
  ▄▓▓▓▓▓▓▓▓▓▄
  ██ ░░ ░░ ██
  ██  ████ ██
  ██  █  █ ██
  ▀▀▀▀▀▀▀▀▀▀▀
```

## Features

- **When did I last change the furnace filter?** Maintenance tracking with intervals, due dates, and service history.
- **How much did that roof quote come in at?** Compare vendor quotes per project with labor/materials breakdowns.
- **Is the dishwasher still under warranty?** Appliance tracking with purchase dates, costs, and linked maintenance.
- **Who replaced the water heater last time?** Service log per maintenance item with vendor tracking.
- **What's the status of all my projects?** Color-coded statuses from ideating through completion.
- **Can I just use my keyboard like a normal person?** Vim-style modal editing. `hjkl` navigation. Undo/redo. Multi-column sort. No mouse required.
- **Will this look like garbage on my terminal?** Colorblind-safe Wong palette with auto light/dark detection.
- **Where does it store my data?** `~/.local/share/micasa/micasa.db`. Back it up with `cp`. You're welcome.

## Install

Requires Go 1.24+ and CGO (for SQLite).

```sh
go install github.com/micasa/micasa/cmd/micasa@latest
```

## Quick start

```sh
micasa --demo   # explore with sample data
micasa          # start fresh
```

## Keybindings

### Normal mode

| Key | Action |
|-----|--------|
| `j` / `k` | Row up / down |
| `h` / `l` | Column left / right |
| `g` / `G` | First / last row |
| `d` / `u` | Half-page down / up |
| `tab` / `shift+tab` | Next / previous tab |
| `s` / `S` | Cycle sort / clear all sorts |
| `enter` | Drilldown or follow link |
| `i` | Enter Edit mode |
| `H` | Toggle house profile |
| `?` | Help |
| `q` | Quit |

### Edit mode

| Key | Action |
|-----|--------|
| `a` | Add entry |
| `e` | Edit cell (full form on ID) |
| `d` | Delete / restore |
| `x` | Show / hide deleted |
| `p` | Edit house profile |
| `u` / `r` | Undo / redo |
| `1`-`9` | Jump to Nth select option |
| `esc` | Back to Normal |

## Tech stack

[Bubble Tea](https://github.com/charmbracelet/bubbletea) +
[Bubbles](https://github.com/charmbracelet/bubbles) +
[huh](https://github.com/charmbracelet/huh) +
[Lip Gloss](https://github.com/charmbracelet/lipgloss) +
[GORM](https://gorm.io)/SQLite

## Contributing

PRs welcome. `go test ./...` before submitting. Pre-commit hooks handle formatting, linting, and tests.

## License

Apache-2.0 -- see [LICENSE](LICENSE).
