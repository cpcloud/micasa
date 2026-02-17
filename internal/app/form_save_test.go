// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openHouseForm enters edit mode and presses p to open the house profile form,
// the same way a user would.
func openHouseForm(m *Model) {
	sendKey(m, "i") // enter edit mode
	sendKey(m, "p") // open house form
}

// openAddForm enters edit mode and presses a to open an add form for the
// active tab, the same way a user would.
func openAddForm(m *Model) {
	sendKey(m, "i") // enter edit mode
	sendKey(m, "a") // add entry
}

func TestUserEditsHouseProfileAndSavesWithCtrlS(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)
	require.Contains(t, m.statusView(), "saved", "user should be in form mode")

	// User changes the nickname field.
	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Beach House"
	m.checkFormDirty()
	require.Contains(t, m.statusView(), "unsaved", "form should be dirty after editing")

	// User presses Ctrl+S to save.
	sendKey(m, "ctrl+s")

	// User sees the form is still open and dirty indicator resets.
	status := m.statusView()
	assert.Contains(t, status, "saved", "form should remain open after ctrl+s")
	assert.NotContains(t, status, "unsaved", "dirty indicator should reset after save")

	// Data actually persisted to the database.
	require.NoError(t, m.loadHouse())
	assert.Equal(t, "Beach House", m.house.Nickname)
}

func TestUserEditsHouseProfileThenSavesThenEditsAgain(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	// First edit + save.
	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Lake House"
	m.checkFormDirty()
	sendKey(m, "ctrl+s")

	// After save, user continues editing in the same form.
	assert.Contains(t, m.statusView(), "saved", "form should remain open after save")
	values.City = "Tahoe"
	m.checkFormDirty()
	assert.Contains(t, m.statusView(), "unsaved", "form should be dirty again after further edits")

	// Second save.
	sendKey(m, "ctrl+s")
	status := m.statusView()
	assert.Contains(t, status, "saved")
	assert.NotContains(t, status, "unsaved")

	// Both values persisted.
	require.NoError(t, m.loadHouse())
	assert.Equal(t, "Lake House", m.house.Nickname)
	assert.Equal(t, "Tahoe", m.house.City)
}

func TestUserAddsProjectAndSavesWithCtrlS(t *testing.T) {
	m := newTestModelWithStore(t)
	openAddForm(m)
	require.Contains(t, m.statusView(), "saved", "user should be in form mode")

	values, ok := m.formData.(*projectFormData)
	require.True(t, ok)
	values.Title = "New Deck"
	m.checkFormDirty()

	sendKey(m, "ctrl+s")

	// User is still in the form.
	status := m.statusView()
	assert.Contains(t, status, "saved", "form should remain open after save")
	assert.NotContains(t, status, "unsaved")
}

func TestUserSeesStatusBarTransitionOnSave(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	// Initially the status bar shows "saved" (clean state).
	view := m.statusView()
	assert.Contains(t, view, "saved")
	assert.NotContains(t, view, "unsaved")

	// User edits a field — status bar flips to "unsaved".
	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Updated"
	m.checkFormDirty()

	view = m.statusView()
	assert.Contains(t, view, "unsaved")

	// User presses Ctrl+S — status bar flips back to "saved".
	sendKey(m, "ctrl+s")

	view = m.statusView()
	assert.Contains(t, view, "saved")
	assert.NotContains(t, view, "unsaved")
}

func TestUserCancelsFormWithEscAfterSaving(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Saved Then Cancelled"
	m.checkFormDirty()

	// Save in place.
	sendKey(m, "ctrl+s")
	assert.Contains(t, m.statusView(), "saved", "form should still be open after save")

	// Esc closes the form, returning to the previous mode.
	sendKey(m, "esc")
	assert.Contains(t, m.statusView(), "EDIT", "esc should close the form and return to edit mode")

	// Data from the save is still persisted.
	require.NoError(t, m.loadHouse())
	assert.Equal(t, "Saved Then Cancelled", m.house.Nickname)
}
