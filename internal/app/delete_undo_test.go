// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// docTabIndex returns the index of the documents tab.
func docTabIndex(m *Model) int {
	for i, tab := range m.tabs {
		if tab.Kind == tabDocuments {
			return i
		}
	}
	return -1
}

func TestDeleteUndoPersistsAcrossTabSwitch(t *testing.T) {
	m := newTestModelWithDemoData(t, testSeed)
	m.resizeTables()

	// Navigate to documents tab -- DeleteDocument has no FK guards.
	di := docTabIndex(m)
	require.GreaterOrEqual(t, di, 0)
	m.switchToTab(di)
	tab := m.activeTab()
	require.NotEmpty(t, tab.Rows, "documents tab should have rows")

	// Directly invoke delete (simulates user pressing d in edit mode).
	m.mode = modeEdit
	m.toggleDeleteSelected()
	require.NotNil(t, m.lastDeleted,
		"lastDeleted should be set after delete (status: %s)", m.status.Text)
	deletedID := m.lastDeleted.ID
	assert.Equal(t, tabDocuments, m.lastDeleted.Tab)

	// Switch away and back.
	m.mode = modeNormal
	m.prevTab()
	require.NotEqual(t, di, m.active)
	m.nextTab()
	require.Equal(t, di, m.active)

	// The deletion ref must still be present after returning.
	require.NotNil(t, m.lastDeleted,
		"lastDeleted must persist across tab switches")
	assert.Equal(t, deletedID, m.lastDeleted.ID)
	assert.Equal(t, tabDocuments, m.lastDeleted.Tab)
}

func TestDeleteUndoClearedOnRestore(t *testing.T) {
	m := newTestModelWithDemoData(t, testSeed)
	m.resizeTables()

	di := docTabIndex(m)
	m.switchToTab(di)
	tab := m.activeTab()
	tab.ShowDeleted = true

	// Delete the selected document.
	m.mode = modeEdit
	m.toggleDeleteSelected()
	require.NotNil(t, m.lastDeleted,
		"lastDeleted should be set after delete (status: %s)", m.status.Text)

	// Restore by toggling again on the same (now-deleted) row.
	m.toggleDeleteSelected()
	assert.Nil(t, m.lastDeleted,
		"lastDeleted should be cleared after restoring the same item")
}
