<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# FTS5 Full-Text Document Search

Issue: #690

## Problem

Every document stores `ExtractedText` but there is no way to search across
documents. "Find all receipts mentioning 'plumber'" currently requires the LLM
chat. FTS5 makes it instant and deterministic.

## Design

### Data Layer

**FTS5 virtual table** (`documents_fts`) as a content-sync (external content)
table backed by the `documents` table:

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
    title,
    notes,
    extracted_text,
    content=documents,
    content_rowid=id,
    tokenize='porter unicode61'
);
```

- `porter` stemmer so "plumbing" matches "plumber"
- `unicode61` for case-folding and diacritic removal
- External content mode: FTS reads from `documents` on demand, no data duplication

**Sync triggers**: INSERT/UPDATE/DELETE on `documents` automatically maintain
the FTS index via standard SQLite external-content triggers.

**Rebuild on open**: On `AutoMigrate`, create the virtual table if missing,
install triggers if missing, then `INSERT INTO documents_fts(documents_fts) VALUES('rebuild')`
to catch any documents that were created before FTS existed.

**Soft-delete aware**: Since we use GORM soft-delete (`deleted_at`), the FTS
table naturally includes soft-deleted rows. The search query JOINs back to
`documents` and filters `deleted_at IS NULL`.

**Search method**:
```go
func (s *Store) SearchDocuments(query string) ([]DocumentSearchResult, error)
```

Returns results with:
- Document ID, Title, FileName, EntityKind, EntityID
- Snippet of matched text (via FTS5 `snippet()` function)
- BM25 rank score

### UI Layer

**Search overlay**: A new overlay (similar to column finder but richer) that
provides instant-as-you-type document search.

**Keybinding**: `ctrl+f` -- universally recognized "find" shortcut, not yet
used in the app.

**State struct**:
```go
type docSearchState struct {
    Input    textinput.Model
    Results  []data.DocumentSearchResult
    Cursor   int
    Err      error
}
```

**Overlay rendering**:
- Title: " Search Documents "
- Input with search icon prompt
- Results list showing:
  - Document title (bold when selected)
  - Entity association (muted, e.g., "P Kitchen Reno")
  - Matched text snippet with highlighted matches
- Empty states: "type to search" (no query), "no matches" (query with no results)
- Hints: enter=open, esc=close

**Behavior**:
- Type to search: re-query FTS on every keystroke (FTS5 is fast enough)
- Up/down to navigate results
- Enter to navigate: switch to Documents tab and select the matched document
- Esc to close
- Works from any tab in normal mode (global search)

**Overlay position**: Between columnFinder and extraction in the overlay stack.

**Mouse support**: Zone-mark each result row with `search-N` prefix for click
navigation.

### Help Integration

Add `ctrl+f` to the Nav Mode help section as "Search docs".
Add to status bar hints at priority 3 (same as "ask").

## Implementation Order

1. Data: FTS5 table creation, triggers, rebuild, search method
2. UI: search state, overlay rendering, key dispatch
3. Navigation: enter to jump to document
4. Tests: user-interaction tests via keypresses
5. Help/status bar integration
6. Mouse click support

## Non-goals

- Searching non-document entities (future extension)
- Advanced query syntax help (FTS5 handles this natively)
- Search history persistence
