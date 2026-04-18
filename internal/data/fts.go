// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	sqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// FTS5 virtual table and trigger names.
const (
	tableFTS         = "documents_fts"
	triggerFTSInsert = "documents_fts_ai"
	triggerFTSDelete = "documents_fts_ad"
	triggerFTSUpdate = "documents_fts_au"
)

// DocumentSearchResult holds a single FTS5 match with metadata for display.
type DocumentSearchResult struct {
	ID         string
	Title      string
	FileName   string
	EntityKind string
	EntityID   string
	Snippet    string
	UpdatedAt  time.Time
}

// setupFTS creates the FTS5 virtual table and sync triggers if they do not
// already exist, then rebuilds the index to catch any documents that were
// created before FTS was added.
func (s *Store) setupFTS() error {
	// Create the external-content FTS5 virtual table. Porter stemmer
	// enables "plumbing" matching "plumber"; unicode61 handles case
	// folding and diacritics.
	createTable := fmt.Sprintf(`
		CREATE VIRTUAL TABLE IF NOT EXISTS %s USING fts5(
			title,
			notes,
			extracted_text,
			content=%s,
			content_rowid=rowid,
			tokenize='porter unicode61'
		)`, tableFTS, TableDocuments)
	if err := s.db.Exec(createTable).Error; err != nil {
		return fmt.Errorf("create FTS table: %w", err)
	}

	// Install triggers to keep the FTS index in sync with the documents
	// table. Triggers are dropped and recreated on every Open so that
	// definition changes (e.g., soft-delete awareness) apply to existing DBs.
	triggers := []struct {
		name string
		sql  string
	}{
		{
			name: triggerFTSInsert,
			sql: fmt.Sprintf(`
				CREATE TRIGGER %s AFTER INSERT ON %s BEGIN
					INSERT INTO %s(rowid, title, notes, extracted_text)
					SELECT new.rowid, new.title, new.notes, new.extracted_text
					WHERE new.deleted_at IS NULL;
				END`, triggerFTSInsert, TableDocuments, tableFTS),
		},
		{
			name: triggerFTSDelete,
			sql: fmt.Sprintf(`
				CREATE TRIGGER %s AFTER DELETE ON %s BEGIN
					INSERT INTO %s(%s, rowid, title, notes, extracted_text)
					VALUES ('delete', old.rowid, old.title, old.notes, old.extracted_text);
				END`, triggerFTSDelete, TableDocuments, tableFTS, tableFTS),
		},
		{
			name: triggerFTSUpdate,
			sql: fmt.Sprintf(`
				CREATE TRIGGER %s AFTER UPDATE ON %s BEGIN
					-- Remove old FTS entry only when it was previously indexed.
					INSERT INTO %s(%s, rowid, title, notes, extracted_text)
					SELECT 'delete', old.rowid, old.title, old.notes, old.extracted_text
					WHERE old.deleted_at IS NULL;
					-- Re-index only when the row is alive (not soft-deleted).
					INSERT INTO %s(rowid, title, notes, extracted_text)
					SELECT new.rowid, new.title, new.notes, new.extracted_text
					WHERE new.deleted_at IS NULL;
				END`, triggerFTSUpdate, TableDocuments, tableFTS, tableFTS, tableFTS),
		},
	}
	for _, t := range triggers {
		// Drop first so updated trigger definitions take effect on existing DBs.
		drop := "DROP TRIGGER IF EXISTS " + t.name
		if err := s.db.Exec(drop).Error; err != nil {
			return fmt.Errorf("drop trigger %s: %w", t.name, err)
		}
		if err := s.db.Exec(t.sql).Error; err != nil {
			return fmt.Errorf("create trigger %s: %w", t.name, err)
		}
	}

	// Rebuild to index any documents created before FTS was set up.
	rebuild := fmt.Sprintf(
		`INSERT INTO %s(%s) VALUES('rebuild')`, tableFTS, tableFTS,
	)
	if err := s.db.Exec(rebuild).Error; err != nil {
		return fmt.Errorf("rebuild FTS index: %w", err)
	}

	return nil
}

// SearchDocuments performs a full-text search across document titles, notes,
// and extracted text. Returns results ranked by BM25 relevance with text
// snippets showing matched context. Only non-deleted documents are returned.
func (s *Store) SearchDocuments(query string) ([]DocumentSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	safeQuery := prepareFTSQuery(query)
	if safeQuery == "" {
		return nil, nil
	}

	var results []DocumentSearchResult
	err := s.db.Raw(fmt.Sprintf(`
		SELECT
			d.id,
			d.title,
			d.file_name,
			d.entity_kind,
			d.entity_id,
			snippet(%s, -1, '>>>', '<<<', '...', 32) AS snippet,
			d.updated_at
		FROM %s
		JOIN %s d ON d.rowid = %s.rowid
		WHERE %s MATCH ?
			AND d.deleted_at IS NULL
		ORDER BY rank
		LIMIT 50
	`, tableFTS, tableFTS, TableDocuments, tableFTS, tableFTS), safeQuery).
		Scan(&results).Error
	if err != nil {
		// FTS syntax errors should not crash the app. Return empty
		// results so the UI can show "no matches" gracefully.
		if isFTSSyntaxError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("search documents: %w", err)
	}
	return results, nil
}

// prepareFTSQuery transforms a user query into a valid FTS5 expression.
//
// Simple queries get prefix matching: "plumb" -> "plumb*". Queries that
// look like FTS5 operator usage (AND/OR/NOT, quotes, parens, *) pass
// through when their quotes and parens are balanced; unbalanced ones
// fall back to safe word search so a stray quote doesn't propagate as
// a syntax error from SQLite.
//
// Returns "" when the input contains no usable terms after sanitization
// (e.g., only punctuation). Callers should treat this as "no results".
func prepareFTSQuery(query string) string {
	upper := strings.ToUpper(query)
	hasFTSOps := strings.ContainsAny(query, `"*()`) ||
		strings.Contains(upper, " AND ") ||
		strings.Contains(upper, " OR ") ||
		strings.Contains(upper, " NOT ") ||
		strings.HasPrefix(upper, "NOT ")

	if hasFTSOps && ftsDelimitersBalanced(query) && hasFTSTerms(query) {
		return query
	}
	return ftsSafeWordsQuery(query)
}

// hasFTSTerms reports whether q contains at least one letter or digit
// that could serve as a search term. Operator-only inputs like "***"
// have no terms and would error at SQLite ("unknown special query").
func hasFTSTerms(q string) bool {
	for _, r := range q {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// ftsDelimitersBalanced reports whether double-quote and parenthesis
// pairs in q are balanced. Inside a quoted phrase, "" escapes a literal
// double-quote per FTS5 grammar.
func ftsDelimitersBalanced(q string) bool {
	parens := 0
	inQuote := false
	for i := 0; i < len(q); i++ {
		c := q[i]
		if inQuote {
			if c == '"' {
				if i+1 < len(q) && q[i+1] == '"' {
					i++
					continue
				}
				inQuote = false
			}
			continue
		}
		switch c {
		case '"':
			inQuote = true
		case '(':
			parens++
		case ')':
			parens--
			if parens < 0 {
				return false
			}
		}
	}
	return !inQuote && parens == 0
}

// ftsSafeWordsQuery splits q into whitespace-separated words, strips
// FTS5 special characters from each, and joins them with a trailing *
// for prefix matching. Words that become empty after stripping are
// dropped; an all-special-chars input yields "".
func ftsSafeWordsQuery(q string) string {
	words := strings.Fields(q)
	out := make([]string, 0, len(words))
	for _, w := range words {
		clean := strings.Map(func(r rune) rune {
			switch r {
			case '"', '*', '(', ')', ':', '+', '-', '^':
				return -1
			}
			return r
		}, w)
		if clean != "" {
			out = append(out, clean+"*")
		}
	}
	return strings.Join(out, " ")
}

// isFTSSyntaxError reports whether err is a SQLite FTS5 syntax error.
// Narrows by error type and result code before inspecting the message,
// so unrelated errors that happen to mention "fts5" can't false-positive.
func isFTSSyntaxError(err error) bool {
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}
	if sqliteErr.Code() != sqlite3.SQLITE_ERROR {
		return false
	}
	return strings.Contains(sqliteErr.Error(), "fts5")
}

// RebuildFTSIndex forces a full rebuild of the FTS5 index. Useful after
// bulk imports or data recovery.
func (s *Store) RebuildFTSIndex() error {
	rebuild := fmt.Sprintf(
		`INSERT INTO %s(%s) VALUES('rebuild')`, tableFTS, tableFTS,
	)
	return s.db.Exec(rebuild).Error
}

// hasFTSTable checks whether the FTS virtual table exists.
func (s *Store) hasFTSTable() bool {
	var count int64
	s.db.Raw(
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`,
		tableFTS,
	).Scan(&count)
	return count > 0
}
