<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

Run the full pre-commit verification bag before committing. All steps must
pass before proceeding to `git commit`. Fix any failures inline.

Run these steps in this order (parallelize where independent):

1. `go mod tidy` -- clean up go.mod/go.sum
2. `nix run '.#pre-commit'` -- linters, formatters, golines (includes govet)
3. `nix run '.#deadcode'` -- whole-program reachability analysis
4. `nix run '.#osv-scanner'` -- security vulnerabilities
5. `go test -shuffle=on ./...` -- all packages, shuffled, no `-v`

Steps 2-4 can run in parallel. Step 5 depends on nothing but is usually the
slowest -- start it in the background while fixing issues from earlier steps.

If `osv-scanner` finds vulnerabilities, try updating the dependency first. Only
add `[[IgnoredVulns]]` to `osv-scanner.toml` if the vuln is genuinely
unreachable in micasa's code paths (with a reason explaining why).

If `deadcode` finds unreachable exports, remove them.

If pre-commit reformats files, re-stage them and re-run.

If pre-commit fails in a worktree with environment or cache errors, recover
with: `direnv allow`, then `git clean -fdx`, then `direnv reload`, then
retry.

Do not proceed to `git commit` until every step passes cleanly. Never use
`--no-verify`.
