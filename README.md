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
- **Is the dishwasher still under warranty?** Appliance tracking with purchase dates, warranty windows, and linked maintenance.

## Install

Requires Go 1.25+:

```sh
go install github.com/cpcloud/micasa/cmd/micasa@latest
```

Or grab a binary from the [latest release](https://github.com/cpcloud/micasa/releases/latest).

```sh
micasa --demo         # poke around with sample data
micasa                # start fresh with your own house
micasa --print-path   # show where the database lives
```

> Your data stays yours. One SQLite file in your platform's data directory. Back it up with `cp "$(micasa --print-path)" backup.db`.

## Documentation

Full docs at [micasa.dev/docs](https://micasa.dev/docs/) -- installation, user guide, keybinding reference, configuration, and development setup.

## Development

Built with the [Charmbracelet](https://github.com/charmbracelet) TUI toolkit, [GORM](https://gorm.io), and [SQLite](https://sqlite.org). [Pure Go](https://go.dev), zero CGO.

PRs welcome. The repo uses a [Nix](https://nixos.org) dev shell with pre-commit hooks for formatting, linting, and tests:

```sh
nix develop          # enter dev shell
go test ./...        # run tests
```

## License

Apache-2.0 -- see [LICENSE](LICENSE).
