<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

<div align="center">
  <img src="images/house.svg" alt="micasa">
</div>

# `micasa`

Your house is quietly plotting to break while you sleep -- and you're dreaming about redoing the kitchen. `micasa` tracks both from your terminal.

> Single SQLite file. No cloud. No account. No subscriptions.

<div align="center">
  <img src="images/demo.gif" alt="micasa demo" width="800">
</div>

## Features

- **When did I last change the furnace filter?** Maintenance schedules, auto-computed due dates, full service history.
- **What if we finally did the backyard?** Projects from napkin sketch to completion -- or graceful abandonment.
- **How much would it actually cost to...** Quotes side by side, vendor history, and the math you need to actually decide.
- **Is the dishwasher still under warranty?** Appliance tracking with purchase dates, warranty status, and maintenance history tied to each one.
- **Who did we use last time?** A vendor directory with contact info, quote history, and every job they've done for you.

## Keyboard driven

Vim-style modal keys: `nav` mode to browse, `edit` mode to change things. Sort by any column, jump to columns with fuzzy search, hide what you don't need, and drill into related records.

See the full [keybinding reference](https://micasa.dev/docs/reference/keybindings/).

## Install

Requires Go 1.25+:

```sh
go install github.com/cpcloud/micasa/cmd/micasa@latest
```

Or grab a binary from the [latest release](https://github.com/cpcloud/micasa/releases/latest).

Linux, macOS, and Windows binaries are available for amd64 and arm64.

```sh
micasa --demo         # poke around with sample data
micasa                # start fresh with your own house
micasa --print-path   # show where the database lives
```

> One SQLite file. Your data, your machine. Back it up with `cp`.

Need Nix or container install options? Use the full [installation guide](https://micasa.dev/docs/getting-started/installation/).

## Documentation

Full docs at [micasa.dev/docs](https://micasa.dev/docs/) -- start with [Installation](https://micasa.dev/docs/getting-started/installation/) and [First Run](https://micasa.dev/docs/getting-started/first-run/), then use the [Guide](https://micasa.dev/docs/guide/) and [Reference](https://micasa.dev/docs/reference/).

## Development

[Pure Go](https://go.dev), zero CGO. [Charmbracelet](https://github.com/charmbracelet) + [GORM](https://gorm.io) + [SQLite](https://sqlite.org). Pair-programmed with [Claude](https://claude.ai) via [Cursor](https://cursor.com).

PRs welcome. The repo uses a [Nix](https://nixos.org) dev shell with pre-commit hooks for formatting, linting, and tests:

```sh
nix develop          # enter dev shell
go test -shuffle=on -v ./...        # run tests
```

## License

Apache-2.0 -- see [LICENSE](LICENSE).
