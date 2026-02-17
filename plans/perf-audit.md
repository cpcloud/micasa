<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Performance Audit

Issue: #312

## Executive Summary

The codebase has no catastrophic performance issues, but the TUI render loop
does significant redundant work per frame. The main waste categories are:

1. **Per-frame recomputation** of column widths, visible projections, and cell
   styles that only change on data load/sort/filter
2. **Glamour renderer allocation** on every LLM streaming tick
3. **String parsing in hot loops** (dates, money, styles) that could use
   pre-computed values

The biggest wins come from caching computed values that only change on data
mutations, not on every render frame.

## Findings

### P0 -- Per-frame waste (highest impact)

#### 1. Cache glamour renderer in chat state

**File:** `internal/app/chat.go:1329-1332`

`renderMarkdown` creates a new `glamour.NewTermRenderer` on every call.
During LLM streaming, `refreshChatViewport` fires ~10x/sec, each time
re-rendering all assistant messages. Glamour initialization parses a JSON
stylesheet and allocates rendering buffers.

**Fix:** Cache the renderer on `chatState`, keyed by width. Invalidate on
resize only.

**Estimate:** Eliminates the dominant per-tick allocation during streaming.

#### 2. Cache column natural widths on Tab

**Files:** `internal/app/table.go:719-746`, `table.go:142,169`

`naturalWidths` iterates every cell of every row calling `lipgloss.Width`
(ANSI-aware string measurement). This runs via `columnWidths` which is called
twice per render: once in `updateTabViewport` (Update phase) and once in
`computeTableViewport` (View phase). Each call also recomputes
`visibleProjection`.

Cell values are stable between data loads -- widths only change on
load/sort/filter/column-hide, not on cursor movement or scrolling.

**Fix:** Cache `naturalWidths` result on the Tab struct. Invalidate on data
change (reload, sort, filter, column toggle). Remove duplicate computation
between Update and View phases.

**Estimate:** Eliminates O(rows * cols) string measurements per frame.

#### 3. Cache lipgloss.Style for default cells

**File:** `internal/app/table.go:564`

`cellStyle` returns `lipgloss.NewStyle()` for every non-special cell on every
render. Most cells hit this default path.

**Fix:** Package-level `var defaultCellStyle = lipgloss.NewStyle()`.

**Estimate:** Small per-cell savings, large aggregate (applies to most cells).

#### 4. Pre-compute urgency/warranty styles at load time

**Files:** `internal/app/table.go:571-605`

`urgencyStyle` and `warrantyStyle` parse date strings and create new
`lipgloss.Style` values per cell per render. The underlying dates don't
change between loads, and the urgency bucket only changes daily.

**Fix:** Compute and store the style per cell at data-load time (in a parallel
`[]lipgloss.Style` slice or on the cell struct).

**Estimate:** Eliminates date parsing + style allocation per urgency/warranty
cell per frame.

### P1 -- Per-interaction waste

#### 5. Pre-compute sort keys

**Files:** `internal/app/sort.go:152-205`

Sort comparisons re-parse dates (`time.Parse`) and money
(`strings.ReplaceAll` + `ParseFloat`) from display strings on every
comparison. With N rows, that's O(N log N) parse operations per sort.
`compareStrings` also allocates via `strings.ToLower` per comparison.

**Fix:** Build a parallel sort-key slice (parsed dates, cents, lowered strings)
before sorting. Compare pre-computed keys.

**Estimate:** Eliminates O(N log N) parse+alloc operations per sort.

#### 6. Cache compact/mag cell transformations

**Files:** `internal/app/compact.go:49-67`, `internal/app/mag.go:107-121`

Both allocate a full N*C cell grid copy on every render when their mode is
active. `compactMoneyValue` also round-trips through string parsing
(formatted string -> cents -> compact string).

**Fix:** Cache transformed grids. Invalidate on data change or mode toggle.

**Estimate:** Eliminates full grid copy per frame in compact/mag modes.

### P2 -- Infrequent but worth noting

#### 7. Cache schema info for LLM chat

**File:** `internal/app/chat.go:1173-1199`

`buildTableInfoFrom` runs N+1 queries (one `TableNames` + one
`TableColumns` per table) on every chat query. Schema doesn't change at
runtime.

**Fix:** Cache schema on first use. Lazily populated, never invalidated.

**Estimate:** Eliminates ~11 queries per chat interaction.

#### 8. reflect.DeepEqual for form dirty check

**File:** `internal/app/model.go:1471`

Called on every keystroke in a form. `reflect.DeepEqual` is expensive.

**Fix:** Low priority. Could use typed comparison or hash, but forms are small
structs and keystroke frequency is human-limited (~10/sec max).

## Implementation Order

1. P0-3: Default cell style constant -- **Done**. Lifted urgency/warranty
   styles and default cell style to package-level vars.
2. P0-1: Cache glamour renderer -- **Done**. Cached on `chatState`, keyed
   by width. Invalidated only on resize.
3. P0-2: Deduplicate natural widths in `computeTableViewport` -- **Done**.
   Compute once as a local variable and reuse for both `columnWidths` calls.
   No cross-frame caching (avoids invalidation bugs).
4. P0-4: Pre-compute urgency/warranty styles -- **Skipped**. Style
   allocation was already fixed by #1. Remaining `time.Parse` per cell is
   ~1.5us/frame, not worth structural changes.
5. P1-5: Pre-compute sort keys -- **Done**. Build a `sortKey` slice once
   (O(N)) before sorting instead of re-parsing on every O(N log N)
   comparison.
6. P1-6: Cache compact/mag transformations -- **Skipped**. These operate
   on viewport-sliced cells (~300-400 cells), and the transforms are cheap.
   Not worth invalidation complexity.

Items P2-7 and P2-8 are deferred -- schema caching is a nice-to-have, and
form dirty checking is human-rate-limited.

## Results

Baseline → optimized (3-run median, 64 cores):

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| ComputeTableViewport | 15,730 → 12,400 (-21%) | 12,592 → 10,912 (-13%) | 253 → 152 (-40%) |
| View | 466,000 → 454,000 (-3%) | 96,300 → 95,400 (-1%) | 2,064 → 1,968 (-5%) |
| TableView | 321,000 → 320,000 (~0%) | 56,200 → 54,500 (-3%) | 1,466 → 1,365 (-7%) |
| BuildBaseView | 459,000 → 493,000 (noise) | 96,900 → 95,400 (-2%) | 2,068 → 1,964 (-5%) |

## Non-findings

- **Bubble Tea Update loop**: Clean message dispatch, no unnecessary work
- **GORM queries**: Properly batched (CountQuotesByProject, etc.), no N+1 in
  the main data path
- **Filter logic**: Efficient in-memory filtering, reasonable allocation pattern
- **Dashboard rendering**: Only runs when visible, acceptable complexity
- **String building in view.go**: Uses strings.Builder where it matters
