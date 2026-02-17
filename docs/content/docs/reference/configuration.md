+++
title = "Configuration"
weight = 2
description = "CLI flags, environment variables, config file, and LLM setup."
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
      --version       Show version and exit.
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

### `OLLAMA_HOST`

Sets the LLM API base URL, overriding the config file value. If the URL
doesn't end with `/v1`, it's appended automatically:

```sh
export OLLAMA_HOST=http://192.168.1.50:11434
micasa   # connects to http://192.168.1.50:11434/v1
```

### `MICASA_LLM_MODEL`

Sets the LLM model name, overriding the config file value:

```sh
export MICASA_LLM_MODEL=llama3.3
micasa   # uses llama3.3 instead of the default qwen3
```

### `MICASA_LLM_TIMEOUT`

Sets the LLM timeout for quick operations (ping, model listing), overriding
the config file value. Uses Go duration syntax:

```sh
export MICASA_LLM_TIMEOUT=15s
micasa   # waits up to 15s for LLM server responses
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

## Database path resolution order

The database path is resolved in this order:

1. Positional CLI argument, if provided
2. `MICASA_DB_PATH` environment variable, if set
3. Platform data directory (see table above)

In `--demo` mode without a path argument, an in-memory database (`:memory:`)
is used.

## Config file

micasa reads a TOML config file from your platform's config directory:

| Platform | Default path |
|----------|-------------|
| Linux    | `$XDG_CONFIG_HOME/micasa/config.toml` (default `~/.config/micasa/config.toml`) |
| macOS    | `~/Library/Application Support/micasa/config.toml` |
| Windows  | `%APPDATA%\micasa\config.toml` |

The config file is optional. If it doesn't exist, all settings use their
defaults. Unset fields fall back to defaults -- you only need to specify the
values you want to change.

### Example config

```toml
# micasa configuration

[llm]
# Base URL for an OpenAI-compatible API endpoint.
# Ollama (default): http://localhost:11434/v1
# llama.cpp:        http://localhost:8080/v1
# LM Studio:        http://localhost:1234/v1
base_url = "http://localhost:11434/v1"

# Model name passed in chat requests.
model = "qwen3"

# Optional: custom context appended to all system prompts.
# Use this to inject domain-specific details about your house, currency, etc.
# extra_context = "My house is a 1920s craftsman in Portland, OR. All budgets are in CAD."

# Timeout for quick LLM server operations (ping, model listing).
# Go duration syntax: "5s", "10s", "500ms", etc. Default: "5s".
# Increase if your LLM server is slow to respond.
# timeout = "5s"
```

### `[llm]` section

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `base_url` | string | `http://localhost:11434/v1` | Root URL of an OpenAI-compatible API. micasa appends `/chat/completions`, `/models`, etc. |
| `model` | string | `qwen3` | Model identifier sent in chat requests. Must be available on the server. |
| `extra_context` | string | (empty) | Free-form text appended to all LLM system prompts. Useful for telling the model about your house, preferred currency, or regional conventions. |
| `timeout` | string | `"5s"` | Max wait time for quick LLM operations (ping, model listing). Go duration syntax, e.g. `"10s"`, `"500ms"`. Increase for slow servers. |

### Supported LLM backends

micasa talks to any server that implements the OpenAI chat completions API
with streaming (SSE). [Ollama](https://ollama.com) is the primary tested
backend:

| Backend | Default URL | Notes |
|---------|-------------|-------|
| [Ollama](https://ollama.com) | `http://localhost:11434/v1` | Default and tested. Models are pulled automatically if not present. |
| [llama.cpp server](https://github.com/ggml-org/llama.cpp) | `http://localhost:8080/v1` | Should work (untested). Pass `--host` and `--port` when starting the server. |
| [LM Studio](https://lmstudio.ai) | `http://localhost:1234/v1` | Should work (untested). Enable the local server in LM Studio settings. |

### Override precedence

Environment variables override config file values. The full precedence order
(highest to lowest):

1. `OLLAMA_HOST` / `MICASA_LLM_MODEL` / `MICASA_LLM_TIMEOUT` environment variables
2. Config file values
3. Built-in defaults

### `extra_context` examples

The `extra_context` field is injected into every system prompt sent to the
LLM, giving it persistent knowledge about your situation:

```toml
[llm]
extra_context = """
My house is a 1920s craftsman bungalow in Portland, OR.
All costs are in USD. Property tax is assessed annually in November.
The HVAC system is a heat pump (Mitsubishi hyper-heat) -- no gas furnace.
"""
```

This helps the model give more relevant answers without you repeating context
in every question.

## Persistent preferences

Some preferences are stored in the SQLite database and persist across
restarts. These are controlled through the UI rather than config files:

| Preference | Default | How to change |
|------------|---------|---------------|
| Dashboard on startup | Shown | Press `D` to toggle; your choice is remembered |
| LLM model | From config | Changed automatically when you switch models in the chat interface |
