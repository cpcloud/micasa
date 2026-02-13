// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusBarHiddenWhenDashboardActive(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.showDashboard = true

	status := m.statusView()

	// Main tab keybindings should be hidden.
	assert.NotContains(t, status, "NAV")
	assert.NotContains(t, status, "switch")
	assert.NotContains(t, status, "sort")
}

func TestStatusBarHiddenWhenHelpActive(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	sendKey(m, "?")
	require.NotNil(t, m.helpViewport)

	status := m.statusView()

	// Main tab keybindings should be hidden.
	assert.NotContains(t, status, "NAV")
	assert.NotContains(t, status, "switch")
	assert.NotContains(t, status, "sort")
}

func TestStatusBarHiddenWhenNotePreviewActive(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.showNotePreview = true
	m.notePreviewText = "test note"

	status := m.statusView()

	// Main tab keybindings should be hidden.
	assert.NotContains(t, status, "NAV")
	assert.NotContains(t, status, "switch")
	assert.NotContains(t, status, "sort")
}

func TestStatusBarHiddenWhenColumnFinderActive(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	sendKey(m, "/")
	require.NotNil(t, m.columnFinder)

	status := m.statusView()

	// Main tab keybindings should be hidden.
	assert.NotContains(t, status, "NAV")
	assert.NotContains(t, status, "switch")
	assert.NotContains(t, status, "sort")
}

func TestStatusBarHiddenWhenCalendarActive(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	fieldValue := ""
	m.openCalendar(&fieldValue, nil)
	require.NotNil(t, m.calendar)

	status := m.statusView()

	// Main tab keybindings should be hidden.
	assert.NotContains(t, status, "NAV")
	assert.NotContains(t, status, "switch")
	assert.NotContains(t, status, "sort")
}

func TestStatusBarShownWhenNoOverlayActive(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.showDashboard = false
	m.showNotePreview = false
	m.helpViewport = nil
	m.columnFinder = nil
	m.calendar = nil

	status := m.statusView()

	// Main tab keybindings should be visible.
	assert.Contains(t, status, "NAV")
}

func TestHasActiveOverlayDetectsAllOverlays(t *testing.T) {
	m := newTestModel()

	// No overlays
	assert.False(t, m.hasActiveOverlay())

	// Dashboard
	m.showDashboard = true
	assert.True(t, m.hasActiveOverlay())
	m.showDashboard = false

	// Help
	m.openHelp()
	assert.True(t, m.hasActiveOverlay())
	m.helpViewport = nil

	// Note preview
	m.showNotePreview = true
	assert.True(t, m.hasActiveOverlay())
	m.showNotePreview = false

	// Column finder
	m.openColumnFinder()
	assert.True(t, m.hasActiveOverlay())
	m.columnFinder = nil

	// Calendar
	fieldValue := ""
	m.openCalendar(&fieldValue, nil)
	assert.True(t, m.hasActiveOverlay())
	m.calendar = nil

	// Still no overlays
	assert.False(t, m.hasActiveOverlay())
}
