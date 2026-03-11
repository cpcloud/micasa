<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

Diagnose and fix failing CI jobs on the current branch's PR.

## 1. Identify failures

1. List check runs: `gh pr checks --json name,state,link`
2. If all checks pass, report that and stop.

## 2. Diagnose each failure

For each failing job:

1. Extract the run ID from the `link` URL returned in step 1.
2. Fetch failed logs: `gh run view <run_id> --log-failed`
3. Read the relevant source files to understand the root cause.
4. Common failure categories:
   - **Compile errors**: Fix the Go code.
   - **Test failures**: Read the failing test, understand the assertion,
     fix the code (not the test) unless the test itself is wrong.
   - **Lint/format**: Run the linter locally, apply fixes.
   - **Nix build**: Hash mismatches need `/update-vendor-hash`. Eval
     errors need Nix expression fixes.
   - **OSV scanner**: Use `/fix-osv-finding`.
   - **Cross-platform (Windows/PowerShell)**: No `&&` in commands, no
     bash-isms, use `-bench .` not `-bench=.`.

## 3. Verify locally

Run `/pre-commit-check` to confirm the fixes work before pushing.

## 4. Push and watch

1. `git push` (or `git push --force-with-lease` if you rebased)
2. Watch checks: `gh pr checks --watch --fail-fast`
3. If new failures appear, loop back to step 2.
