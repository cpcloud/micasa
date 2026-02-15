+++
title = "Building from Source"
weight = 1
description = "How to build micasa from source."
linkTitle = "Building"
+++

## Prerequisites

- **Go 1.25+** (the only hard requirement)
- **Nix** (optional, but provides the full dev environment)

## Quick build

```sh
git clone https://github.com/cpcloud/micasa.git
cd micasa
CGO_ENABLED=0 go build ./cmd/micasa
./micasa --demo
```

micasa uses a pure-Go SQLite driver, so `CGO_ENABLED=0` works and produces a
fully static binary.

## Nix dev shell

The recommended development environment uses Nix flakes:

```sh
nix develop
```

This gives you:

- Go compiler
- gopls (language server)
- golangci-lint
- golines + gofumpt (formatting)
- ripgrep, fd, tokei (search and stats)
- Pre-commit hooks (auto-installed on first shell entry)

Everything is pinned to a consistent version. No system dependency surprises.

## Build commands

From within the dev shell (or with Go installed):

```sh
# Build the binary
go build ./cmd/micasa

# Run directly
go run ./cmd/micasa -- --demo

# Run tests
go test -shuffle=on -v ./...
```

## Nix build

To build the binary via Nix (reproducible, hermetic):

```sh
nix build
./result/bin/micasa --demo
```

## Nix flake apps

The flake exposes several convenience apps:

| Command | Description |
|---------|-------------|
| `nix run` | Run micasa directly |
| `nix run '.#website'` | Serve the website locally |
| `nix run '.#build-docs'` | Build the Hugo docs into `website/docs/` |
| `nix run '.#record-demo'` | Record the demo GIF |

## Container image

Build the OCI container image via Nix:

```sh
nix build '.#micasa-container'
docker load < result
docker run -it --rm micasa:latest --demo
```
