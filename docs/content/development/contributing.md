+++
title = "Contributing"
weight = 4
description = "How to contribute to micasa."
linkTitle = "Contributing"
+++

PRs welcome! Here's how to get set up and what to expect.

## Setup

1. Fork and clone the repo
2. Enter the dev shell: `nix develop` (or install Go 1.24+ manually)
3. The dev shell auto-installs pre-commit hooks on first entry

## Pre-commit hooks

The repo uses pre-commit hooks that run automatically on `git commit`:

- **golines** + **gofumpt**: code formatting (max 100 chars/line)
- **golangci-lint**: static analysis
- **license-header**: ensures every source file has the Apache-2.0 header

If a hook fails, fix the issue and commit again. The hooks auto-fix formatting
where possible.

## Commit conventions

micasa uses [conventional commits](https://www.conventionalcommits.org/) with
scopes. Examples:

```
feat(dashboard): add spending summary section
fix(maintenance): correct next-due computation for edge case
refactor(handlers): extract shared inline edit logic
test(sort): add multi-column comparator tests
docs(website): update feature list
```

Use `docs(website):` (not `feat(website):`) for website changes to avoid
triggering version bumps.

## Code style

- **Run `go mod tidy` before committing** to keep dependencies clean
- Follow existing patterns: check how similar features are implemented
- Use the Wong colorblind-safe palette for any new colors (see `styles.go`)
- Always provide both Light and Dark variants in `lipgloss.AdaptiveColor`
- Keep type safety: avoid `any` casts, use proper types and guards
- DRY: search for existing helpers before adding new ones

## Tests

- Write tests for new features
- Don't test implementation details -- test behavior
- Run `go test -shuffle=on -v ./...` to verify
- All tests must pass on Linux, macOS, and Windows

## License

By contributing, you agree that your contributions will be licensed under the
Apache License 2.0. All source files must include the copyright header (the
pre-commit hook handles this automatically).
