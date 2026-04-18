// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchDocumentsBasic(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	// Create documents with extracted text.
	require.NoError(t, store.CreateDocument(&Document{
		Title:         "Plumber Receipt",
		FileName:      "receipt.pdf",
		ExtractedText: "Invoice from ABC Plumbing for kitchen sink repair",
		Notes:         "paid in full",
	}))
	require.NoError(t, store.CreateDocument(&Document{
		Title:         "HVAC Manual",
		FileName:      "manual.pdf",
		ExtractedText: "Installation guide for central air conditioning unit",
	}))

	results, err := store.SearchDocuments("plumb")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Plumber Receipt", results[0].Title)
	assert.Contains(t, results[0].Snippet, "Plumb")
}

func TestSearchDocumentsMatchesTitle(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	require.NoError(t, store.CreateDocument(&Document{
		Title:    "Kitchen Renovation Quote",
		FileName: "quote.pdf",
	}))

	results, err := store.SearchDocuments("kitchen")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Kitchen Renovation Quote", results[0].Title)
}

func TestSearchDocumentsMatchesNotes(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	require.NoError(t, store.CreateDocument(&Document{
		Title:    "Receipt",
		FileName: "r.pdf",
		Notes:    "emergency plumbing repair on Sunday",
	}))

	results, err := store.SearchDocuments("emergency")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Receipt", results[0].Title)
}

func TestSearchDocumentsExcludesSoftDeleted(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	require.NoError(t, store.CreateDocument(&Document{
		Title:         "Deleted Doc",
		FileName:      "deleted.pdf",
		ExtractedText: "plumber invoice",
	}))
	docs, err := store.ListDocuments(false)
	require.NoError(t, err)
	require.Len(t, docs, 1)

	require.NoError(t, store.DeleteDocument(docs[0].ID))

	results, err := store.SearchDocuments("plumber")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchDocumentsEmptyQuery(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	require.NoError(t, store.CreateDocument(&Document{
		Title:    "Something",
		FileName: "s.pdf",
	}))

	results, err := store.SearchDocuments("")
	require.NoError(t, err)
	assert.Nil(t, results)

	results, err = store.SearchDocuments("   ")
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchDocumentsMultipleMatches(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	require.NoError(t, store.CreateDocument(&Document{
		Title:         "Receipt 1",
		FileName:      "r1.pdf",
		ExtractedText: "plumber fixed the kitchen sink",
	}))
	require.NoError(t, store.CreateDocument(&Document{
		Title:         "Receipt 2",
		FileName:      "r2.pdf",
		ExtractedText: "plumber replaced bathroom faucet",
	}))
	require.NoError(t, store.CreateDocument(&Document{
		Title:    "Unrelated",
		FileName: "u.pdf",
	}))

	results, err := store.SearchDocuments("plumber")
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestSearchDocumentsPorterStemming(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	require.NoError(t, store.CreateDocument(&Document{
		Title:         "Painting Invoice",
		FileName:      "inv.pdf",
		ExtractedText: "Professional painting services rendered",
	}))

	// "painted" should match "painting" via porter stemmer (both stem to "paint").
	results, err := store.SearchDocuments("painted")
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchDocumentsUpdateReflected(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	require.NoError(t, store.CreateDocument(&Document{
		Title:         "Old Title",
		FileName:      "doc.pdf",
		ExtractedText: "original text about gardening",
	}))
	docs, err := store.ListDocuments(false)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	id := docs[0].ID

	// Update extraction text.
	require.NoError(t, store.UpdateDocumentExtraction(id, "new text about plumbing", nil, "", nil))

	results, err := store.SearchDocuments("plumbing")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, id, results[0].ID)

	// Old text should no longer match.
	results, err = store.SearchDocuments("gardening")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchDocumentsBadSyntaxGraceful(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	// Each input is one that previously surfaced an FTS5 syntax error
	// (or could in future regressions). Sanitization should keep them
	// from reaching SQLite as bad syntax.
	bad := []string{
		`"unclosed`,
		`unclosed"`,
		`(kitchen`,
		`kitchen)`,
		`((nested`,
		`"phrase with "" inside`,
	}
	for _, q := range bad {
		t.Run(q, func(t *testing.T) {
			results, err := store.SearchDocuments(q)
			require.NoError(t, err)
			assert.Empty(t, results)
		})
	}
}

// TestSearchDocumentsTrailingOperator covers the isFTSSyntaxError
// safety net: queries that prepareFTSQuery passes through (balanced
// delimiters, has terms) but still violate FTS5 grammar should be
// swallowed and returned as no-results.
func TestSearchDocumentsTrailingOperator(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	require.NoError(t, store.CreateDocument(&Document{
		Title:    "Anything",
		FileName: "a.pdf",
	}))

	// "(a AND)" passes hasFTSOps (paren), balanced (paren+quote), has
	// terms (a), but FTS5 errors on the dangling AND operator.
	results, err := store.SearchDocuments("(a AND)")
	require.NoError(t, err)
	assert.Empty(t, results)
}

// TestSearchDocumentsAllSpecialChars verifies that input consisting only
// of FTS5 metacharacters is treated as an empty search (no results, no
// error) rather than producing a malformed MATCH clause.
func TestSearchDocumentsAllSpecialChars(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	require.NoError(t, store.CreateDocument(&Document{
		Title:    "Something",
		FileName: "s.pdf",
	}))

	for _, q := range []string{`***`, `:::`, `+++---`} {
		t.Run(q, func(t *testing.T) {
			results, err := store.SearchDocuments(q)
			require.NoError(t, err)
			assert.Nil(t, results)
		})
	}
}

func TestSearchDocumentsEntityFields(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	require.NoError(t, store.CreateDocument(&Document{
		Title:         "Project Doc",
		FileName:      "pd.pdf",
		EntityKind:    DocumentEntityProject,
		EntityID:      "01JTEST00000000000000042",
		ExtractedText: "kitchen renovation details",
	}))

	results, err := store.SearchDocuments("kitchen")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, DocumentEntityProject, results[0].EntityKind)
	assert.Equal(t, "01JTEST00000000000000042", results[0].EntityID)
}

func TestSearchDocumentsSnippetFromBestColumn(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	// Match is in title only -- snippet should reflect the title.
	require.NoError(t, store.CreateDocument(&Document{
		Title:    "Plumber Receipt",
		FileName: "receipt.pdf",
	}))

	results, err := store.SearchDocuments("plumber")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(
		t,
		results[0].Snippet,
		"Plumb",
		"snippet should come from title when that's the matching column",
	)
}

func TestSearchDocumentsCaseInsensitive(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	require.NoError(t, store.CreateDocument(&Document{
		Title:         "HVAC Manual",
		FileName:      "hvac.pdf",
		ExtractedText: "Central Air Conditioning INSTALLATION Guide",
	}))

	// All case variants should match.
	for _, q := range []string{"hvac", "HVAC", "Hvac", "installation", "GUIDE"} {
		results, err := store.SearchDocuments(q)
		require.NoError(t, err, "query %q should not error", q)
		assert.Len(t, results, 1, "query %q should match", q)
	}
}

func TestPrepareFTSQuerySimple(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "hello*", prepareFTSQuery("hello"))
	assert.Equal(t, "hello* world*", prepareFTSQuery("hello world"))
}

func TestPrepareFTSQueryOperators(t *testing.T) {
	t.Parallel()
	// Balanced operator queries pass through unchanged.
	assert.Equal(t, `"exact phrase"`, prepareFTSQuery(`"exact phrase"`))
	assert.Equal(t, "plumb*", prepareFTSQuery("plumb*"))
	assert.Equal(t, "a AND b", prepareFTSQuery("a AND b"))
	assert.Equal(t, "a OR b", prepareFTSQuery("a OR b"))
	assert.Equal(t, "NOT bad", prepareFTSQuery("NOT bad"))
	assert.Equal(t, "(a OR b)", prepareFTSQuery("(a OR b)"))
}

func TestPrepareFTSQueryUnbalancedFallsBack(t *testing.T) {
	t.Parallel()
	// Unbalanced delimiters trigger the safe fallback: specials are
	// stripped and each remaining word gets prefix matching.
	assert.Equal(t, "unclosed*", prepareFTSQuery(`"unclosed`))
	assert.Equal(t, "kitchen*", prepareFTSQuery(`(kitchen`))
	assert.Equal(t, "kitchen*", prepareFTSQuery(`kitchen)`))
	assert.Equal(t, "a* b*", prepareFTSQuery(`"a b`))
}

func TestPrepareFTSQueryAllSpecials(t *testing.T) {
	t.Parallel()
	// No usable terms -> empty string. SearchDocuments treats this as
	// "no results" without issuing a MATCH against SQLite.
	assert.Empty(t, prepareFTSQuery(`***`))
	assert.Empty(t, prepareFTSQuery(`+ - :`))
}

func TestFTSDelimitersBalanced(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"", true},
		{"plain text", true},
		{`"phrase"`, true},
		{`"with "" escape"`, true},
		{`(a OR b)`, true},
		{`((nested))`, true},
		{`"unclosed`, false},
		{`unclosed"`, false},
		{`(open`, false},
		{`close)`, false},
		{`)early`, false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, ftsDelimitersBalanced(tt.in))
		})
	}
}

func TestRebuildFTSIndex(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	require.NoError(t, store.CreateDocument(&Document{
		Title:         "Test Doc",
		FileName:      "t.pdf",
		ExtractedText: "searchable content here",
	}))

	require.NoError(t, store.RebuildFTSIndex())

	results, err := store.SearchDocuments("searchable")
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestHasFTSTable(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	assert.True(t, store.hasFTSTable())
}
