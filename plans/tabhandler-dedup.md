<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# TabHandler Implementation Deduplication

Issue: #520

## Problem

The 8 `TabHandler` implementations in `handlers.go` shared near-identical
patterns in **Load() count-fetch blocks** -- ~15 instances of
`counts, err := fn(ids); if err != nil { counts = map[uint]int{} }` across
main handlers and scoped handler constructors.

> **Note:** This plan originally also covered `Snapshot()` deduplication via a
> `makeSnapshot[T any]` generic helper. The undo/redo feature was removed
> entirely in #572, so that section no longer applies.

## Approach

One targeted helper. No framework, no config structs, no interface changes.

### 1. `fetchCounts` helper

```go
func fetchCounts(fn func([]uint) (map[uint]int, error), ids []uint) map[uint]int
```

Returns the count map on success, empty map on error. Replaces the 3-line
pattern everywhere it appears (~15 call sites in main and scoped handlers).

Before:
```go
quoteCounts, err := store.CountQuotesByProject(ids)
if err != nil {
    quoteCounts = map[uint]int{}
}
```

After:
```go
quoteCounts := fetchCounts(store.CountQuotesByProject, ids)
```

### 3. Scoped handler constructors -- no structural change

The scoped constructors (`newApplianceMaintenanceHandler`, etc.) vary enough in
their overrides (inlineEditFn, startAddFn, submitFn) that a factory abstraction
would not materially reduce complexity. These stay as-is, but their internal
`loadFn` closures benefit from `fetchCounts`.

## Out of scope

- No changes to the `TabHandler` interface.
- No changes to `scopedHandler` struct.
- No changes to Load() structure beyond replacing count-fetch boilerplate.
- Scoped handler factory (approach 3 from the issue) -- deferred; the
  `fetchCounts` cleanup is sufficient for now.

## Verification

- `go build ./...`
- `go test -shuffle=on ./...`
- Existing tests cover Load paths through integration tests.
