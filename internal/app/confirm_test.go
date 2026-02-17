// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/cpcloud/micasa/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedProject creates a project and reloads the active tab, positioning the
// cursor on the first row so key-driven operations can target it.
func seedProject(t *testing.T, m *Model, title string) {
	t.Helper()
	h := projectHandler{}
	m.formData = &projectFormData{
		Title:         title,
		ProjectTypeID: m.projectTypes[0].ID,
		Status:        data.ProjectStatusPlanned,
	}
	require.NoError(t, h.SubmitForm(m))
	require.NoError(t, m.reloadActiveTab())
	tab := m.activeTab()
	require.NotNil(t, tab)
	tab.Table.SetCursor(0)
	tab.Table.Focus()
}

// TestDeleteConfirmThenAccept simulates a user pressing d on a project row,
// seeing the confirmation overlay, then pressing y to confirm the delete.
func TestDeleteConfirmThenAccept(t *testing.T) {
	m := newTestModelWithStore(t)
	seedProject(t, m, "Deck Build")

	tab := m.activeTab()
	require.Len(t, tab.Rows, 1, "expected one project row after creation")

	// Enter edit mode and press d.
	sendKey(m, "i")
	require.Equal(t, modeEdit, m.mode)
	sendKey(m, "d")

	// Confirmation overlay should be open — item NOT yet deleted.
	require.NotNil(t, m.confirm, "expected confirmation overlay after pressing d")
	assert.Contains(t, m.confirm.Prompt, "Deck Build")

	h := projectHandler{}
	rows, _, _, err := h.Load(m.store, false)
	require.NoError(t, err)
	assert.Len(t, rows, 1, "project should still exist before confirming")

	// Confirm with y.
	sendKey(m, "y")
	assert.Nil(t, m.confirm, "confirmation overlay should close after y")
	assert.Contains(t, m.status.Text, "Deleted")

	rows, _, _, err = h.Load(m.store, false)
	require.NoError(t, err)
	assert.Empty(t, rows, "project should be soft-deleted after confirming")
}

// TestDeleteConfirmThenCancel simulates a user pressing d, seeing the
// confirmation overlay, then pressing esc to cancel — the item survives.
func TestDeleteConfirmThenCancel(t *testing.T) {
	m := newTestModelWithStore(t)
	seedProject(t, m, "Paint Fence")

	sendKey(m, "i")
	sendKey(m, "d")
	require.NotNil(t, m.confirm)

	// Cancel with esc.
	sendKey(m, "esc")
	assert.Nil(t, m.confirm, "confirmation overlay should close after esc")
	assert.Contains(t, m.status.Text, "Cancelled")

	h := projectHandler{}
	rows, _, _, err := h.Load(m.store, false)
	require.NoError(t, err)
	assert.Len(t, rows, 1, "project should survive after cancelling delete")
}

// TestDeleteConfirmEnterAlsoConfirms verifies that enter works as an
// alternative to y for confirming a delete.
func TestDeleteConfirmEnterAlsoConfirms(t *testing.T) {
	m := newTestModelWithStore(t)
	seedProject(t, m, "Fix Roof")

	sendKey(m, "i")
	sendKey(m, "d")
	require.NotNil(t, m.confirm)

	sendKey(m, "enter")
	assert.Nil(t, m.confirm)

	h := projectHandler{}
	rows, _, _, err := h.Load(m.store, false)
	require.NoError(t, err)
	assert.Empty(t, rows, "project should be deleted after enter confirmation")
}

// TestRestoreSkipsConfirmation verifies that restoring a soft-deleted item
// is immediate — no confirmation dialog shown.
func TestRestoreSkipsConfirmation(t *testing.T) {
	m := newTestModelWithStore(t)
	seedProject(t, m, "Install Blinds")

	// Delete via the confirmation flow.
	sendKey(m, "i")
	sendKey(m, "d")
	require.NotNil(t, m.confirm)
	sendKey(m, "y")
	require.Nil(t, m.confirm)

	// Show deleted items so we can select the deleted row.
	tab := m.activeTab()
	require.NotNil(t, tab)
	tab.ShowDeleted = true
	require.NoError(t, m.reloadActiveTab())
	tab.Table.SetCursor(0)
	require.Len(t, tab.Rows, 1)
	require.True(t, tab.Rows[0].Deleted)

	// Press d to restore — should NOT open confirmation.
	sendKey(m, "d")
	assert.Nil(t, m.confirm, "restore should not show confirmation dialog")
	assert.Contains(t, m.status.Text, "Restored")

	h := projectHandler{}
	rows, _, _, err := h.Load(m.store, false)
	require.NoError(t, err)
	assert.Len(t, rows, 1, "project should be restored")
}

// TestConfirmOverlayRendered verifies the confirmation overlay appears
// in the rendered view with the expected prompt and key hints.
func TestConfirmOverlayRendered(t *testing.T) {
	m := newTestModelWithStore(t)
	seedProject(t, m, "Add Deck")

	sendKey(m, "i")
	sendKey(m, "d")
	require.NotNil(t, m.confirm)

	view := m.buildView()
	assert.Contains(t, view, "Add Deck")
	assert.Contains(t, view, "confirm")
	assert.Contains(t, view, "cancel")
}

// TestConfirmOverlayAbsorbsKeys verifies that keys other than y/enter
// also dismiss the overlay (treating them as cancel).
func TestConfirmOverlayAbsorbsKeys(t *testing.T) {
	m := newTestModelWithStore(t)
	seedProject(t, m, "Storm Door")

	sendKey(m, "i")
	sendKey(m, "d")
	require.NotNil(t, m.confirm)

	// Any non-confirm key should cancel.
	sendKey(m, "n")
	assert.Nil(t, m.confirm)
	assert.Contains(t, m.status.Text, "Cancelled")

	h := projectHandler{}
	rows, _, _, err := h.Load(m.store, false)
	require.NoError(t, err)
	assert.Len(t, rows, 1, "project should survive after pressing n")
}
