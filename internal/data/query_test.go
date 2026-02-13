// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTableNames(t *testing.T) {
	store := newTestStore(t)

	names, err := store.TableNames()
	require.NoError(t, err)

	// Should include our core tables.
	assert.Contains(t, names, "house_profiles")
	assert.Contains(t, names, "projects")
	assert.Contains(t, names, "vendors")
	assert.Contains(t, names, "maintenance_items")
	assert.Contains(t, names, "appliances")

	// Should not include sqlite internals.
	for _, name := range names {
		assert.NotContains(t, name, "sqlite_")
	}
}

func TestTableColumns(t *testing.T) {
	store := newTestStore(t)

	cols, err := store.TableColumns("projects")
	require.NoError(t, err)
	assert.NotEmpty(t, cols)

	// Check that the id column is a PK.
	var foundID bool
	for _, col := range cols {
		if col.Name == "id" {
			foundID = true
			assert.Greater(t, col.PK, 0)
		}
	}
	assert.True(t, foundID, "expected to find id column")
}

func TestTableColumnsInvalidName(t *testing.T) {
	store := newTestStore(t)

	_, err := store.TableColumns("'; DROP TABLE projects; --")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestReadOnlyQuerySelect(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.SeedDefaults())

	cols, rows, err := store.ReadOnlyQuery("SELECT name FROM project_types ORDER BY name LIMIT 3")
	require.NoError(t, err)
	assert.Equal(t, []string{"name"}, cols)
	assert.Len(t, rows, 3)
}

func TestReadOnlyQueryRejectsInsert(t *testing.T) {
	store := newTestStore(t)
	_, _, err := store.ReadOnlyQuery("INSERT INTO projects (title) VALUES ('hack')")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only SELECT")
}

func TestReadOnlyQueryRejectsDelete(t *testing.T) {
	store := newTestStore(t)
	_, _, err := store.ReadOnlyQuery("DELETE FROM projects WHERE id = 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only SELECT")
}

func TestReadOnlyQueryRejectsMultiStatement(t *testing.T) {
	store := newTestStore(t)
	_, _, err := store.ReadOnlyQuery("SELECT * FROM projects; DROP TABLE projects")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multiple statements")
}

func TestReadOnlyQueryRejectsAttach(t *testing.T) {
	store := newTestStore(t)
	_, _, err := store.ReadOnlyQuery("SELECT * FROM (SELECT 1) ATTACH DATABASE '/tmp/x' AS x")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disallowed keyword: ATTACH")
}

func TestReadOnlyQueryRejectsPragma(t *testing.T) {
	store := newTestStore(t)
	_, _, err := store.ReadOnlyQuery("SELECT * FROM pragma_table_info('projects') WHERE 1=1 PRAGMA journal_mode")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disallowed keyword: PRAGMA")
}

func TestReadOnlyQueryEmpty(t *testing.T) {
	store := newTestStore(t)
	_, _, err := store.ReadOnlyQuery("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestReadOnlyQueryAllowsDeletedAtColumn(t *testing.T) {
	store := newTestStore(t)
	// "deleted_at" contains "DELETE" as a substring but should be allowed.
	cols, _, err := store.ReadOnlyQuery(
		"SELECT id FROM projects WHERE deleted_at IS NULL LIMIT 1",
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"id"}, cols)
}

func TestContainsWord(t *testing.T) {
	assert.True(t, containsWord("SELECT * DELETE FROM", "DELETE"))
	assert.False(t, containsWord("WHERE DELETED_AT IS NULL", "DELETE"))
	assert.True(t, containsWord("DROP TABLE x", "DROP"))
	assert.False(t, containsWord("BACKDROP", "DROP"))
}

func TestIsSafeIdentifier(t *testing.T) {
	assert.True(t, isSafeIdentifier("projects"))
	assert.True(t, isSafeIdentifier("house_profiles"))
	assert.True(t, isSafeIdentifier("table123"))
	assert.False(t, isSafeIdentifier(""))
	assert.False(t, isSafeIdentifier("table; DROP"))
	assert.False(t, isSafeIdentifier("table'name"))
}

func TestDataDump(t *testing.T) {
	store := newTestStoreWithDemoData(t, testSeed)

	dump := store.DataDump()
	assert.NotEmpty(t, dump)
	// Should contain table headers with row counts.
	assert.Contains(t, dump, "rows)")
	// Should include actual data rows as bullet points.
	assert.Contains(t, dump, "- ")
}
