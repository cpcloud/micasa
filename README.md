<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

<div align="center">
  <img src="website/house.svg" alt="micasa" width="120">
</div>

# `micasa`

A terminal UI for tracking everything about your home. Single SQLite file. No cloud. No account. No subscriptions. Just your house.

Your house is quietly plotting to break while you sleep -- and you're dreaming about redoing the kitchen. `micasa` tracks both from your terminal.

## Features

- **When did I last change the furnace filter?** Maintenance schedules, auto-computed due dates, full service history.
- **What if we finally did the backyard?** Projects from napkin sketch to completion -- or graceful abandonment.
- **How much would it actually cost to...** Quotes, vendors, stare at the numbers, close the laptop, reopen the laptop.
- **Is the dishwasher still under warranty?** Appliance tracking with purchase dates, warranty windows, and linked maintenance.

## Install

```sh
go install github.com/cpcloud/micasa/cmd/micasa@latest
```

Or grab a binary from the [latest release](https://github.com/cpcloud/micasa/releases/latest).

```sh
micasa --demo   # poke around with sample data
micasa          # start fresh with your own house
```

Your data stays yours. `~/.local/share/micasa/micasa.db`. One file. Back it up with `cp`.

## Keybindings

No mouse required.

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
| `c` / `C` | Hide / show columns |
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

## Tech

Built with the [Charmbracelet](https://github.com/charmbracelet) TUI toolkit, [GORM](https://gorm.io), and [SQLite](https://sqlite.org). [Pure Go](https://go.dev), zero CGO.

## Contributing

PRs welcome. `go test ./...` before submitting. Pre-commit hooks handle formatting, linting, and tests.

## License

Apache-2.0 -- see [LICENSE](LICENSE).
