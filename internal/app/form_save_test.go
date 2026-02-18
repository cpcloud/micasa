// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

func TestUserCreatesMaintenanceWithDurationInterval(t *testing.T) {
	m := newTestModelWithStore(t)

	// User navigates to the Maintenance tab, then opens the add form.
	m.active = tabIndex(tabMaintenance)
	openAddForm(m)
	require.Contains(t, m.statusView(), "saved", "user should be in form mode")

	values, ok := m.formData.(*maintenanceFormData)
	require.True(t, ok)
	values.Name = "HVAC Filter"
	values.IntervalMonths = "1y"
	m.checkFormDirty()

	// User presses Ctrl+S to save.
	sendKey(m, "ctrl+s")
	status := m.statusView()
	assert.Contains(t, status, "saved", "form should stay open after save")
	assert.NotContains(t, status, "unsaved")

	// Verify the interval was stored as 12 months in the database.
	items, err := m.store.ListMaintenance(false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, 12, items[0].IntervalMonths)

	// Verify round-trip: editing shows compact format "1y", not "12".
	got := maintenanceFormValues(items[0])
	assert.Equal(t, "1y", got.IntervalMonths)
}

func TestUserCreatesMaintenanceWithCombinedInterval(t *testing.T) {
	m := newTestModelWithStore(t)
	m.active = tabIndex(tabMaintenance)
	openAddForm(m)
	require.Contains(t, m.statusView(), "saved", "user should be in form mode")

	values, ok := m.formData.(*maintenanceFormData)
	require.True(t, ok)
	values.Name = "Gutter Cleaning"
	values.IntervalMonths = "2y 6m"
	sendKey(m, "ctrl+s")

	items, err := m.store.ListMaintenance(false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, 30, items[0].IntervalMonths)
	assert.Equal(t, "2y 6m", maintenanceFormValues(items[0]).IntervalMonths)
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

func TestUserEscDirtyFormShowsConfirmation(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	// User edits a field, making the form dirty.
	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Unsaved Change"
	m.checkFormDirty()
	require.True(t, m.formDirty, "form should be dirty after edit")

	// User presses ESC — should see confirmation instead of exiting.
	sendKey(m, "esc")
	assert.Equal(t, modeForm, m.mode, "should still be in form mode")
	assert.True(t, m.confirmDiscard, "confirm dialog should be active")
	status := m.statusView()
	assert.Contains(t, status, "Discard unsaved changes?")
	assert.Contains(t, status, "discard")
	assert.Contains(t, status, "keep editing")
}

func TestUserConfirmsDiscardWithY(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Will Be Discarded"
	m.checkFormDirty()

	// ESC triggers confirmation, y discards.
	sendKey(m, "esc")
	require.True(t, m.confirmDiscard)
	sendKey(m, "y")
	assert.False(t, m.confirmDiscard, "confirm dialog should be dismissed")
	assert.NotEqual(t, modeForm, m.mode, "should have exited form mode")
	assert.Nil(t, m.form, "form should be nil after discard")

	// Database should still have the original value, not the discarded edit.
	require.NoError(t, m.loadHouse())
	assert.Equal(t, "Test House", m.house.Nickname)
}

func TestUserCancelsDiscardWithN(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Keep This Edit"
	m.checkFormDirty()

	// ESC triggers confirmation, n cancels it.
	sendKey(m, "esc")
	require.True(t, m.confirmDiscard)
	sendKey(m, "n")
	assert.False(t, m.confirmDiscard, "confirm dialog should be dismissed")
	assert.Equal(t, modeForm, m.mode, "should remain in form mode")
	assert.NotNil(t, m.form, "form should still be open")

	// The unsaved edit should still be in the form data.
	assert.Equal(t, "Keep This Edit", values.Nickname)
}

func TestUserCancelsDiscardWithEsc(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Keep Editing"
	m.checkFormDirty()

	// ESC triggers confirmation, a second ESC cancels it.
	sendKey(m, "esc")
	require.True(t, m.confirmDiscard)
	sendKey(m, "esc")
	assert.False(t, m.confirmDiscard, "confirm dialog should be dismissed")
	assert.Equal(t, modeForm, m.mode, "should remain in form mode")
}

func TestCleanFormExitsImmediatelyOnEsc(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	// Form is clean (no edits), so ESC should exit immediately.
	require.False(t, m.formDirty)
	sendKey(m, "esc")
	assert.False(t, m.confirmDiscard, "confirm should not appear for clean forms")
	assert.NotEqual(t, modeForm, m.mode, "should exit form on ESC when clean")
}

func TestConfirmDiscardSwallowsOtherKeys(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Dirty"
	m.checkFormDirty()

	sendKey(m, "esc")
	require.True(t, m.confirmDiscard)

	// Keys other than y/n/esc should be swallowed.
	sendKey(m, "a")
	assert.True(t, m.confirmDiscard, "confirm should still be active after 'a'")
	sendKey(m, "x")
	assert.True(t, m.confirmDiscard, "confirm should still be active after 'x'")
	assert.Equal(t, modeForm, m.mode, "should remain in form mode")
}

func TestSavedFormExitsImmediatelyOnEsc(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	// Edit, save in place, then ESC — form is clean after save.
	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Saved Edit"
	m.checkFormDirty()
	sendKey(m, "ctrl+s")
	require.False(t, m.formDirty, "form should be clean after ctrl+s")

	// ESC should exit immediately since form is no longer dirty.
	sendKey(m, "esc")
	assert.False(t, m.confirmDiscard, "no confirm needed after save")
	assert.NotEqual(t, modeForm, m.mode, "should exit form mode")

	// Saved data should persist.
	require.NoError(t, m.loadHouse())
	assert.Equal(t, "Saved Edit", m.house.Nickname)
}

func TestCtrlQDirtyFormShowsConfirmation(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Unsaved Quit"
	m.checkFormDirty()
	require.True(t, m.formDirty)

	// ctrl+q on a dirty form should show confirmation, not quit.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	assert.Nil(t, cmd, "should not quit immediately")
	assert.True(t, m.confirmDiscard, "confirm dialog should be active")
	assert.Equal(t, modeForm, m.mode, "should still be in form mode")
}

func TestCtrlQDirtyFormConfirmQuits(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Will Quit"
	m.checkFormDirty()

	// ctrl+q triggers confirmation, y quits.
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	require.True(t, m.confirmDiscard)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	assert.NotNil(t, cmd, "y after ctrl+q should return quit command")

	// Database should have the original value, not the unsaved edit.
	require.NoError(t, m.loadHouse())
	assert.Equal(t, "Test House", m.house.Nickname)
}

func TestCtrlQDirtyFormCancelStaysInForm(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)

	values, ok := m.formData.(*houseFormData)
	require.True(t, ok)
	values.Nickname = "Keep Editing"
	m.checkFormDirty()

	// ctrl+q triggers confirmation, n cancels.
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	require.True(t, m.confirmDiscard)
	sendKey(m, "n")
	assert.False(t, m.confirmDiscard, "confirm should be dismissed")
	assert.False(t, m.confirmQuit, "quit flag should be cleared")
	assert.Equal(t, modeForm, m.mode, "should remain in form mode")
	assert.NotNil(t, m.form, "form should still be open")
}

func TestCtrlQCleanFormQuitsImmediately(t *testing.T) {
	m := newTestModelWithStore(t)
	openHouseForm(m)
	require.False(t, m.formDirty)

	// ctrl+q on a clean form should quit immediately.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	assert.NotNil(t, cmd, "clean form ctrl+q should quit immediately")
	assert.False(t, m.confirmDiscard, "no confirm needed for clean form")
}
