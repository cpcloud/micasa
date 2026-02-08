<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# `micasa`

A terminal UI for tracking everything about your home. Single SQLite file. No cloud. No account. No subscriptions. Just your house.

<div align="center"><pre><code>    ▄▓▄
  ▄▓▓▓▓▓▄
▄▓▓▓▓▓▓▓▓▓▄
██ ░░ ░░ ██
██  ████ ██
██  █  █ ██
▀▀▀▀▀▀▀▀▀▀▀
</code></pre></div>

## Features

- **When did I last change the furnace filter?** Maintenance schedules, due dates, service history.
- **How much was that roof quote?** Vendor quotes with cost breakdowns.
- **Is the dishwasher still under warranty?** Appliances with linked maintenance.
- **Who replaced the water heater?** Service log with vendor tracking.
- **What's the status of everything?** Projects with color-coded statuses.
- **Do I need a mouse?** No. Vim-style `hjkl`, undo/redo, multi-column sort.
- **Where's my data?** `~/.local/share/micasa/micasa.db`. Back it up with `cp`.

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
