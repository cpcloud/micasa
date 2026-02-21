<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Cross-platform audit (#402)

Audit of Linux/Unix assumptions that may break on macOS or Windows.

## Summary

The codebase is in good cross-platform shape overall. CI already tests on
all three platforms (ubuntu, macos, windows). Production code consistently
uses `filepath.Join`, `os.UserHomeDir`, and the `adrg/xdg` library for
platform-aware paths. The `runtime.GOOS` switch in `docopen.go` correctly
dispatches to platform-native file openers.

## Findings

### Silent failure

1. **`internal/app/docopen.go:62` -- non-actionable error when OS viewer
   missing.** On Linux, if `xdg-open` is not installed (containers, minimal
   systems, WSL without desktop), `cmd.Run()` returns a raw exec error like
   `exec: "xdg-open": executable file not found in $PATH`. The status bar
   shows `open: exec: "xdg-open": ...` which gives no remediation hint.
   Same issue applies on macOS (`open`) and Windows (`cmd /c start`) in
   degraded environments, though less likely there.

2. **`internal/data/doccache.go:48` -- file permission `0o600` ignored on
   Windows.** Cached document BLOBs are written with owner-only read/write
   permissions. Windows ignores Unix permission modes, so files end up with
   default ACLs (potentially readable by other local users). Severity
   depends on whether documents contain sensitive data.

3. **`internal/data/path.go:30` -- dir permission `0o700` ignored on
   Windows.** Same as above for the cache directory itself.

### Cosmetic

4. **`cmd/micasa/main_test.go:20-31,59,69` -- hardcoded `/custom/path.db`,
   `/tmp/demo.db`, `/env/override.db` in path-resolution tests.** These are
   pure string identity checks (resolveDBPath returns verbatim), not
   filesystem operations. They work on all platforms but look Unix-centric.

5. **`cmd/micasa/main_test.go:232` -- `/nonexistent/path.db` in error
   test.** Tests that a missing source is rejected. Works on all platforms
   since the path doesn't exist anywhere, but reads as Unix-only.

6. **`cmd/micasa/main_test.go:247` -- `file:///tmp/backup.db?mode=rwc` in
   validation test.** Tests that URI-style destinations are rejected. The
   validator catches the `file:` prefix before the path matters, so it
   works cross-platform. Reads as Unix-centric.

7. **`internal/app/view_test.go:430` -- `/home/user/long/path/to/data.db`
   test fixture.** Used to test `truncateLeft` string truncation. Pure
   string operation, works everywhere.

8. **`internal/app/view_test.go:462` -- `shortenHome("/tmp/other.db")`
   test.** Tests that non-home paths are unchanged. Works everywhere since
   `/tmp/other.db` won't match any user's home directory.

9. **`internal/data/query_test.go:91` -- `/tmp/x` in SQL injection test.**
   Tests that `ATTACH DATABASE` keyword is rejected. The validator catches
   the keyword before the path is evaluated. Works everywhere.

10. **No `//go:build` constraints.** The codebase uses `runtime.GOOS`
    switches instead of build tags. This is idiomatic Go and not a problem
    -- both approaches are valid. The runtime switch in `docopen.go` is
    simpler than three platform-specific files.

11. **`internal/app/view.go:1067-1069` -- `os.PathSeparator` in
    `shortenHome`.** Uses `string(os.PathSeparator)` instead of
    `filepath.Join`. The code is correct and cross-platform; `filepath.Join`
    would be marginally cleaner but the current form is explicit.

## What's already correct

- **CI matrix:** `ci.yml` tests on ubuntu-latest, macos-latest,
  windows-latest with Go 1.25.
- **Path construction:** All production code uses `filepath.Join`,
  `filepath.Base`, `filepath.Ext`, `filepath.Clean`, `filepath.Abs`.
- **Config/data paths:** `adrg/xdg` library handles Linux vs macOS vs
  Windows differences.
- **Home directory:** `os.UserHomeDir()` used everywhere; no `$HOME`
  hardcoding.
- **Test binary:** `main_test.go:82-84` correctly appends `.exe` on
  Windows.
- **Temp dirs:** Tests use `t.TempDir()` for filesystem operations.
- **No POSIX syscalls:** No `syscall` package usage, no signal handling,
  no `/proc` or `/sys` reads.
- **No shell-outs:** The only `exec.Command` calls are the platform-aware
  file opener and `go build` in tests.

## Fixes

### Fix 1 -- Actionable error in docopen.go

Wrap the `cmd.Run()` error with platform-specific remediation hints when
the opener command is not found (exec.ErrNotFound).

### Fix 2 -- Document Windows permission limitation

Add comments documenting the Windows ACL limitation on the `0o600` and
`0o700` modes. No code change needed -- this is the standard Go pattern
and there's no practical cross-platform alternative without pulling in
Windows ACL libraries (over-engineering for a TUI app's document cache).

### Non-fixes

Findings 4-11 are cosmetic. The hardcoded paths in tests are string
identity checks, not filesystem operations. Changing them to
`filepath.Join(...)` would make the tests harder to read without improving
correctness. No changes needed.
