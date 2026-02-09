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
  -h, --help    Show help.
      --demo    Launch with sample data in an in-memory database.
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

## Environment variables

### `MICASA_DB_PATH`

Sets the default database path when no positional argument is given. Equivalent
to passing the path as an argument:

```sh
export MICASA_DB_PATH=/path/to/my/house.db
micasa   # uses /path/to/my/house.db
```

### `XDG_DATA_HOME`

micasa follows the [XDG Base Directory
Specification](https://specifications.freedesktop.org/basedir-spec/latest/).
When no path is specified (via argument or `MICASA_DB_PATH`), the database is
stored at:

```
$XDG_DATA_HOME/micasa/micasa.db
```

If `XDG_DATA_HOME` is not set, it defaults to `~/.local/share`:

```
~/.local/share/micasa/micasa.db
```

## Resolution order

The database path is resolved in this order:

1. Positional CLI argument, if provided
2. `MICASA_DB_PATH` environment variable, if set
3. `$XDG_DATA_HOME/micasa/micasa.db` (or `~/.local/share/micasa/micasa.db`)

In `--demo` mode without a path argument, an in-memory database (`:memory:`)
is used.
