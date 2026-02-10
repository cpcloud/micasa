+++
title = "Configuration"
weight = 2
description = "CLI flags, environment variables, and database path resolution."
linkTitle = "Configuration"
+++

micasa has minimal configuration -- it's designed to work out of the box.

## CLI flags

```
Usage: micasa [<db-path>] [flags]

A terminal UI for tracking everything about your home.

Arguments:
  [<db-path>]    SQLite database path. Pass with --demo to persist demo data.

Flags:
  -h, --help          Show help.
      --demo          Launch with sample data in an in-memory database.
      --print-path    Print the resolved database path and exit.
```

### `<db-path>`

Optional positional argument. When provided, micasa uses this path for the
SQLite database instead of the default location.

When combined with `--demo`, the demo data is written to this file (instead
of in-memory), so you can restart with the same demo state:

```sh
micasa --demo /tmp/my-demo.db   # creates and populates
micasa /tmp/my-demo.db          # reopens with the demo data
```

### `--demo`

Launches with fictitious sample data: a house profile, several projects,
maintenance items, appliances, service log entries, and quotes. Without a
`<db-path>`, the database lives in memory and disappears when you quit.

### `--print-path`

Prints the resolved database path to stdout and exits. Useful for scripting
and backup:

```sh
micasa --print-path                               # platform default
MICASA_DB_PATH=/tmp/foo.db micasa --print-path    # /tmp/foo.db
micasa --print-path /custom/path.db               # /custom/path.db
micasa --demo --print-path                        # :memory:
micasa --demo --print-path /tmp/d.db              # /tmp/d.db
cp "$(micasa --print-path)" backup.db             # backup the database
```

## Environment variables

### `MICASA_DB_PATH`

Sets the default database path when no positional argument is given. Equivalent
to passing the path as an argument:

```sh
export MICASA_DB_PATH=/path/to/my/house.db
micasa   # uses /path/to/my/house.db
```

### Platform data directory

micasa uses platform-aware data directories (via
[adrg/xdg](https://github.com/adrg/xdg)). When no path is specified (via
argument or `MICASA_DB_PATH`), the database is stored at:

| Platform | Default path |
|----------|-------------|
| Linux    | `$XDG_DATA_HOME/micasa/micasa.db` (default `~/.local/share/micasa/micasa.db`) |
| macOS    | `~/Library/Application Support/micasa/micasa.db` |
| Windows  | `%LOCALAPPDATA%\micasa\micasa.db` |

On Linux, `XDG_DATA_HOME` is respected per the [XDG Base Directory
Specification](https://specifications.freedesktop.org/basedir-spec/latest/).

## Resolution order

The database path is resolved in this order:

1. Positional CLI argument, if provided
2. `MICASA_DB_PATH` environment variable, if set
3. Platform data directory (see table above)

In `--demo` mode without a path argument, an in-memory database (`:memory:`)
is used.
