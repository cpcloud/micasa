// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestModel creates a minimal Model for mode tests (no database).
func newTestModel() *Model {
	styles := DefaultStyles()
	m := &Model{
		styles: styles,
		tabs:   NewTabs(styles),
		active: 0,
		mode:   modeNormal,
	}
	// Seed minimal rows so cursor operations don't panic.
	for i := range m.tabs {
		m.tabs[i].Table.SetRows([]table.Row{{"1", "test"}})
		m.tabs[i].Rows = []rowMeta{{ID: 1}}
	}
	return m
}

func sendKey(m *Model, key string) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	// Some keys need special types.
	switch key {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEscape}
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	case "ctrl+c":
		msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	case keyCtrlN:
		msg = tea.KeyMsg{Type: tea.KeyCtrlN}
	case "ctrl+o":
		msg = tea.KeyMsg{Type: tea.KeyCtrlO}
	}
	m.Update(msg)
}

func TestStartsInNormalMode(t *testing.T) {
	m := newTestModel()
	assert.Equal(t, modeNormal, m.mode)
}

func TestEnterEditMode(t *testing.T) {
	m := newTestModel()
	sendKey(m, "i")
	assert.Equal(t, modeEdit, m.mode)
}

func TestExitEditModeWithEsc(t *testing.T) {
	m := newTestModel()
	sendKey(m, "i")
	sendKey(m, "esc")
	assert.Equal(t, modeNormal, m.mode)
}

func TestTableKeyMapNormalMode(t *testing.T) {
	m := newTestModel()
	// In normal mode, HalfPageDown should include "d".
	tab := m.activeTab()
	require.NotNil(t, tab)
	keys := tab.Table.KeyMap.HalfPageDown.Keys()
	assert.True(
		t,
		containsKey(keys, "d"),
		"expected 'd' in HalfPageDown keys for normal mode, got %v",
		keys,
	)
}

func TestTableKeyMapEditMode(t *testing.T) {
	m := newTestModel()
	sendKey(m, "i")
	tab := m.activeTab()
	require.NotNil(t, tab)
	keys := tab.Table.KeyMap.HalfPageDown.Keys()
	assert.False(
		t,
		containsKey(keys, "d"),
		"'d' should not be in HalfPageDown keys in edit mode, got %v",
		keys,
	)
	assert.True(
		t,
		containsKey(keys, "ctrl+d"),
		"expected 'ctrl+d' in HalfPageDown keys for edit mode, got %v",
		keys,
	)
}

func TestTableKeyMapRestoredOnNormalReturn(t *testing.T) {
	m := newTestModel()
	sendKey(m, "i")
	sendKey(m, "esc")
	tab := m.activeTab()
	require.NotNil(t, tab)
	keys := tab.Table.KeyMap.HalfPageDown.Keys()
	assert.True(
		t,
		containsKey(keys, "d"),
		"expected 'd' restored in HalfPageDown after returning to normal, got %v",
		keys,
	)
}

func TestColumnNavH(t *testing.T) {
	m := newTestModel()
	tab := m.activeTab()
	initial := tab.ColCursor
	sendKey(m, "l")
	if len(tab.Specs) > 1 {
		assert.NotEqual(t, initial, tab.ColCursor, "expected column cursor to advance on 'l'")
	}
}

func TestColumnNavClampsLeft(t *testing.T) {
	m := newTestModel()
	tab := m.activeTab()
	tab.ColCursor = 0
	sendKey(m, "h")
	assert.Equal(t, 0, tab.ColCursor)
}

func TestCaretJumpsToFirstColumn(t *testing.T) {
	m := newTestModel()
	tab := m.activeTab()
	tab.ColCursor = len(tab.Specs) - 1
	sendKey(m, "^")
	assert.Equal(t, 0, tab.ColCursor)
}

func TestDollarJumpsToLastColumn(t *testing.T) {
	m := newTestModel()
	tab := m.activeTab()
	tab.ColCursor = 0
	sendKey(m, "$")
	assert.Equal(t, len(tab.Specs)-1, tab.ColCursor)
}

func TestNextTabAdvances(t *testing.T) {
	m := newTestModel()
	// Verify mode transitions via sendKey don't reset the active tab.
	m.active = 0
	sendKey(m, "i")
	assert.Equal(t, modeEdit, m.mode)
	assert.Equal(t, 0, m.active, "entering edit mode should not change active tab")
	m.active = 2
	sendKey(m, "esc")
	assert.Equal(t, modeNormal, m.mode)
	assert.Equal(t, 2, m.active, "entering normal mode should not change active tab")
}

func TestQuitOnlyInNormalMode(t *testing.T) {
	m := newTestModel()

	// In edit mode, 'ctrl+q' should quit (returns tea.Quit).
	sendKey(m, "i")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	assert.NotNil(t, cmd, "'ctrl+q' should quit even in edit mode")
}

func TestIKeyDoesNothingInEditMode(t *testing.T) {
	m := newTestModel()
	sendKey(m, "i")
	require.Equal(t, modeEdit, m.mode)
	// Press 'i' again — should not switch mode or do anything unexpected.
	sendKey(m, "i")
	assert.Equal(t, modeEdit, m.mode, "expected to stay in modeEdit")
}

func TestHouseToggle(t *testing.T) {
	m := newTestModel()
	m.hasHouse = true
	assert.False(t, m.showHouse, "expected house hidden initially")
	// Tab toggles house in both modes.
	sendKey(m, "tab")
	assert.True(t, m.showHouse, "expected house shown after tab")
	sendKey(m, "tab")
	assert.False(t, m.showHouse, "expected house hidden after second tab")
}

func TestHelpToggle(t *testing.T) {
	m := newTestModel()
	sendKey(m, "?")
	assert.NotNil(t, m.helpViewport, "expected help visible after '?'")
	sendKey(m, "?")
	assert.Nil(t, m.helpViewport, "expected help hidden after second '?'")
}

func TestHelpViewportScrolling(t *testing.T) {
	m := newTestModel()
	sendKey(m, "?")
	require.NotNil(t, m.helpViewport, "expected help visible")

	// Scroll down and verify offset moves.
	sendKey(m, "j")
	if m.helpViewport.TotalLineCount() > m.helpViewport.Height {
		assert.NotZero(t, m.helpViewport.YOffset, "expected viewport to scroll down on 'j'")
	}

	// Scroll back up.
	sendKey(m, "k")
	assert.Equal(t, 0, m.helpViewport.YOffset, "expected viewport at top after scrolling back up")

	// Go to bottom with G.
	sendKey(m, "G")
	if m.helpViewport.TotalLineCount() > m.helpViewport.Height {
		assert.True(t, m.helpViewport.AtBottom(), "expected viewport at bottom after 'G'")
	}

	// Go to top with g.
	sendKey(m, "g")
	assert.True(t, m.helpViewport.AtTop(), "expected viewport at top after 'g'")

	// Esc dismisses.
	sendKey(m, "esc")
	assert.Nil(t, m.helpViewport, "expected help hidden after esc")
}

func TestHelpOverlayFixedWidthOnScroll(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 20 // Small height forces scrolling.
	sendKey(m, "?")
	require.NotNil(t, m.helpViewport, "expected help visible")
	if m.helpViewport.TotalLineCount() <= m.helpViewport.Height {
		t.Skip("help content fits without scrolling at this height")
	}

	// Measure width at top.
	widthAtTop := lipgloss.Width(m.helpView())

	// Scroll to middle.
	for i := 0; i < 5; i++ {
		sendKey(m, "j")
	}
	widthAtMiddle := lipgloss.Width(m.helpView())

	// Scroll to bottom.
	sendKey(m, "G")
	widthAtBottom := lipgloss.Width(m.helpView())

	assert.Equal(t, widthAtTop, widthAtMiddle, "help width changed from top to middle")
	assert.Equal(t, widthAtTop, widthAtBottom, "help width changed from top to bottom")
}

func TestHelpScrollIndicatorChanges(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 20
	sendKey(m, "?")
	require.NotNil(t, m.helpViewport, "expected help visible")
	if m.helpViewport.TotalLineCount() <= m.helpViewport.Height {
		t.Skip("help content fits without scrolling at this height")
	}

	viewAtTop := m.helpView()
	assert.Contains(t, viewAtTop, "Top")

	sendKey(m, "G")
	viewAtBottom := m.helpView()
	assert.Contains(t, viewAtBottom, "Bot")

	// Scroll back up one line from bottom -- should show percentage.
	sendKey(m, "k")
	viewAtMiddle := m.helpView()
	assert.NotContains(t, viewAtMiddle, "Top")
	assert.NotContains(t, viewAtMiddle, "Bot")
	assert.Contains(t, viewAtMiddle, "%")
}

func TestHelpAbsorbsOtherKeys(t *testing.T) {
	m := newTestModel()
	sendKey(m, "?")
	require.NotNil(t, m.helpViewport, "expected help visible")

	// Keys that would normally affect the model should be absorbed.
	sendKey(m, "i")
	assert.Equal(t, modeNormal, m.mode, "'i' should not switch to edit mode while help is open")
}

func TestDeleteRequiresEditMode(t *testing.T) {
	m := newTestModel()
	// In normal mode, 'd' is half-page-down (table handles it).
	// It should NOT trigger delete.
	sendKey(m, "d")
	assert.Empty(t, m.status.Text, "'d' in normal mode should not set status")
}

func TestEscClearsStatusInNormalMode(t *testing.T) {
	m := newTestModel()
	m.status = statusMsg{Text: "something", Kind: statusInfo}
	sendKey(m, "esc")
	assert.Empty(t, m.status.Text)
}

func TestProjectStatusFilterToggleKeys(t *testing.T) {
	m := newTestModel()
	tab := m.activeTab()
	require.NotNil(t, tab)
	require.Equal(t, tabProjects, tab.Kind, "expected projects tab to be active")
	assert.False(t, tab.HideCompleted, "project status filters should start disabled")
	assert.False(t, tab.HideAbandoned, "project status filters should start disabled")

	sendKey(m, "z")
	assert.True(t, tab.HideCompleted, "expected HideCompleted enabled after first z")
	assert.Contains(t, m.status.Text, "Completed")
	assert.Contains(t, m.status.Text, "hidden")

	sendKey(m, "z")
	assert.False(t, tab.HideCompleted, "expected HideCompleted disabled after second z")
	assert.Contains(t, m.status.Text, "Completed")
	assert.Contains(t, m.status.Text, "shown")

	sendKey(m, "a")
	assert.True(t, tab.HideAbandoned, "expected HideAbandoned enabled after first a")
	assert.Contains(t, m.status.Text, "Abandoned")
	assert.Contains(t, m.status.Text, "hidden")

	sendKey(m, "a")
	assert.False(t, tab.HideAbandoned, "expected HideAbandoned disabled after second a")

	sendKey(m, "t")
	assert.True(t, tab.HideCompleted, "expected settled toggle to enable completed filter")
	assert.True(t, tab.HideAbandoned, "expected settled toggle to enable abandoned filter")
	assert.Contains(t, m.status.Text, "Settled")
	assert.Contains(t, m.status.Text, "hidden")

	sendKey(m, "t")
	assert.False(t, tab.HideCompleted, "expected settled toggle to disable completed filter")
	assert.False(t, tab.HideAbandoned, "expected settled toggle to disable abandoned filter")
}

func TestProjectStatusFilterToggleIgnoredOutsideProjects(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabQuotes)
	tab := m.activeTab()
	require.NotNil(t, tab)
	require.Equal(t, tabQuotes, tab.Kind, "expected quotes tab to be active")

	for _, key := range []string{"z", "a", "t"} {
		sendKey(m, key)
	}
	assert.False(t, tab.HideCompleted,
		"project status filters should remain disabled on non-project tabs")
	assert.False(t, tab.HideAbandoned,
		"project status filters should remain disabled on non-project tabs")
	assert.Empty(t, m.status.Text)
}

func TestKeyDispatchEditModeOnly(t *testing.T) {
	m := newTestModel()

	// 'p' should not change mode in normal mode.
	sendKey(m, "p")
	assert.Equal(t, modeNormal, m.mode, "'p' should not change mode in normal mode")

	// 'esc' should be handled in edit mode (back to normal).
	sendKey(m, "i")
	require.Equal(t, modeEdit, m.mode)
	sendKey(m, "esc")
	assert.Equal(t, modeNormal, m.mode)
}

func TestModeAfterFormExit(t *testing.T) {
	m := newTestModel()
	// Enter edit mode via key, open a form, then exit.
	sendKey(m, "i")
	require.Equal(t, modeEdit, m.mode)
	m.prevMode = m.mode
	m.mode = modeForm
	// Simulate exitForm (no key to close a form without a database).
	m.exitForm()
	assert.Equal(t, modeEdit, m.mode, "expected modeEdit after exitForm (was in edit before form)")

	// Return to normal mode via key, then form again.
	sendKey(m, "esc")
	require.Equal(t, modeNormal, m.mode)
	m.prevMode = m.mode
	m.mode = modeForm
	m.exitForm()
	assert.Equal(
		t,
		modeNormal,
		m.mode,
		"expected modeNormal after exitForm (was in normal before form)",
	)
}

func TestTabTogglesHouseInEditMode(t *testing.T) {
	m := newTestModel()
	m.hasHouse = true
	sendKey(m, "i")
	require.Equal(t, modeEdit, m.mode)
	assert.False(t, m.showHouse)
	sendKey(m, "tab")
	assert.True(t, m.showHouse, "tab should toggle house in edit mode")
}

func TestTabSwitchKeysBlockedInEditMode(t *testing.T) {
	m := newTestModel()
	sendKey(m, "i")
	require.Equal(t, modeEdit, m.mode)
	startTab := m.active
	// b/f (tab-switch keys) should not switch tabs in edit mode.
	sendKey(m, "b")
	assert.Equal(t, startTab, m.active, "b should not switch tabs in edit mode")
	sendKey(m, "f")
	assert.Equal(t, startTab, m.active, "f should not switch tabs in edit mode")
}

func TestModeBadgeFixedWidth(t *testing.T) {
	styles := DefaultStyles()
	normalBadge := styles.ModeNormal.Render("NAV")
	normalWidth := lipgloss.Width(normalBadge)

	editBadge := styles.ModeEdit.
		Width(normalWidth).
		Align(lipgloss.Center).
		Render("EDIT")
	editWidth := lipgloss.Width(editBadge)

	assert.Equal(t, normalWidth, editWidth, "badge widths should match")
}

func TestShiftPrefixOnUppercaseKeycap(t *testing.T) {
	m := newTestModel()
	// Uppercase "H" should produce a keycap containing "SHIFT+H".
	rendered := m.keycap("H")
	assert.Contains(t, rendered, "SHIFT+H")
	// Lowercase "h" should produce "H" (uppercased), not "SHIFT+H".
	rendered = m.keycap("h")
	assert.NotContains(t, rendered, "SHIFT")
}

func containsKey(keys []string, target string) bool {
	for _, k := range keys {
		if k == target {
			return true
		}
	}
	return false
}
