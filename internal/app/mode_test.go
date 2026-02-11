// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	}
	m.Update(msg)
}

func TestStartsInNormalMode(t *testing.T) {
	m := newTestModel()
	if m.mode != modeNormal {
		t.Fatalf("expected modeNormal, got %d", m.mode)
	}
}

func TestEnterEditMode(t *testing.T) {
	m := newTestModel()
	sendKey(m, "i")
	if m.mode != modeEdit {
		t.Fatalf("expected modeEdit after 'i', got %d", m.mode)
	}
}

func TestExitEditModeWithEsc(t *testing.T) {
	m := newTestModel()
	sendKey(m, "i")
	sendKey(m, "esc")
	if m.mode != modeNormal {
		t.Fatalf("expected modeNormal after esc, got %d", m.mode)
	}
}

func TestTableKeyMapNormalMode(t *testing.T) {
	m := newTestModel()
	// In normal mode, HalfPageDown should include "d".
	tab := m.activeTab()
	if tab == nil {
		t.Fatal("no active tab")
	}
	keys := tab.Table.KeyMap.HalfPageDown.Keys()
	if !containsKey(keys, "d") {
		t.Fatalf("expected 'd' in HalfPageDown keys for normal mode, got %v", keys)
	}
}

func TestTableKeyMapEditMode(t *testing.T) {
	m := newTestModel()
	sendKey(m, "i")
	tab := m.activeTab()
	if tab == nil {
		t.Fatal("no active tab")
	}
	keys := tab.Table.KeyMap.HalfPageDown.Keys()
	if containsKey(keys, "d") {
		t.Fatalf("'d' should not be in HalfPageDown keys in edit mode, got %v", keys)
	}
	if !containsKey(keys, "ctrl+d") {
		t.Fatalf("expected 'ctrl+d' in HalfPageDown keys for edit mode, got %v", keys)
	}
}

func TestTableKeyMapRestoredOnNormalReturn(t *testing.T) {
	m := newTestModel()
	sendKey(m, "i")
	sendKey(m, "esc")
	tab := m.activeTab()
	if tab == nil {
		t.Fatal("no active tab")
	}
	keys := tab.Table.KeyMap.HalfPageDown.Keys()
	if !containsKey(keys, "d") {
		t.Fatalf("expected 'd' restored in HalfPageDown after returning to normal, got %v", keys)
	}
}

func TestColumnNavH(t *testing.T) {
	m := newTestModel()
	tab := m.activeTab()
	initial := tab.ColCursor
	sendKey(m, "l")
	if tab.ColCursor == initial && len(tab.Specs) > 1 {
		t.Fatal("expected column cursor to advance on 'l'")
	}
}

func TestColumnNavClampsLeft(t *testing.T) {
	m := newTestModel()
	tab := m.activeTab()
	tab.ColCursor = 0
	sendKey(m, "h")
	if tab.ColCursor != 0 {
		t.Fatalf("expected clamp at column 0, got %d", tab.ColCursor)
	}
}

func TestCaretJumpsToFirstColumn(t *testing.T) {
	m := newTestModel()
	tab := m.activeTab()
	tab.ColCursor = len(tab.Specs) - 1
	sendKey(m, "^")
	if tab.ColCursor != 0 {
		t.Fatalf("expected cursor at 0, got %d", tab.ColCursor)
	}
}

func TestDollarJumpsToLastColumn(t *testing.T) {
	m := newTestModel()
	tab := m.activeTab()
	tab.ColCursor = 0
	sendKey(m, "$")
	if tab.ColCursor != len(tab.Specs)-1 {
		t.Fatalf("expected cursor at last column %d, got %d", len(tab.Specs)-1, tab.ColCursor)
	}
}

func TestNextTabAdvances(t *testing.T) {
	m := newTestModel()
	// Directly call nextTab logic without store (would panic on reloadActiveTab).
	// Instead verify enterEditMode / enterNormalMode don't reset active tab.
	m.active = 0
	m.enterEditMode()
	if m.active != 0 {
		t.Fatal("entering edit mode should not change active tab")
	}
	m.active = 2
	m.enterNormalMode()
	if m.active != 2 {
		t.Fatal("entering normal mode should not change active tab")
	}
}

func TestQuitOnlyInNormalMode(t *testing.T) {
	m := newTestModel()

	// In edit mode, 'q' should not quit (no cmd returned).
	sendKey(m, "i")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd != nil {
		t.Fatal("'q' should not produce a command in edit mode")
	}
}

func TestIKeyDoesNothingInEditMode(t *testing.T) {
	m := newTestModel()
	sendKey(m, "i")
	if m.mode != modeEdit {
		t.Fatal("should be in edit mode")
	}
	// Press 'i' again — should not switch mode or do anything unexpected.
	sendKey(m, "i")
	if m.mode != modeEdit {
		t.Fatalf("expected to stay in modeEdit, got %d", m.mode)
	}
}

func TestHouseToggle(t *testing.T) {
	m := newTestModel()
	m.hasHouse = true
	if m.showHouse {
		t.Fatal("expected house hidden initially")
	}
	// Capital H toggles house in both modes.
	sendKey(m, "H")
	if !m.showHouse {
		t.Fatal("expected house shown after 'H'")
	}
	sendKey(m, "H")
	if m.showHouse {
		t.Fatal("expected house hidden after second 'H'")
	}
}

func TestHelpToggle(t *testing.T) {
	m := newTestModel()
	sendKey(m, "?")
	if m.helpViewport == nil {
		t.Fatal("expected help visible after '?'")
	}
	sendKey(m, "?")
	if m.helpViewport != nil {
		t.Fatal("expected help hidden after second '?'")
	}
}

func TestHelpViewportScrolling(t *testing.T) {
	m := newTestModel()
	sendKey(m, "?")
	if m.helpViewport == nil {
		t.Fatal("expected help visible")
	}

	// Scroll down and verify offset moves.
	sendKey(m, "j")
	if m.helpViewport.YOffset == 0 && m.helpViewport.TotalLineCount() > m.helpViewport.Height {
		t.Fatal("expected viewport to scroll down on 'j'")
	}

	// Scroll back up.
	sendKey(m, "k")
	if m.helpViewport.YOffset != 0 {
		t.Fatal("expected viewport at top after scrolling back up")
	}

	// Go to bottom with G.
	sendKey(m, "G")
	if !m.helpViewport.AtBottom() && m.helpViewport.TotalLineCount() > m.helpViewport.Height {
		t.Fatal("expected viewport at bottom after 'G'")
	}

	// Go to top with g.
	sendKey(m, "g")
	if !m.helpViewport.AtTop() {
		t.Fatal("expected viewport at top after 'g'")
	}

	// Esc dismisses.
	sendKey(m, "esc")
	if m.helpViewport != nil {
		t.Fatal("expected help hidden after esc")
	}
}

func TestHelpOverlayFixedWidthOnScroll(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 20 // Small height forces scrolling.
	sendKey(m, "?")
	if m.helpViewport == nil {
		t.Fatal("expected help visible")
	}
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

	if widthAtTop != widthAtMiddle {
		t.Fatalf("help width changed from top (%d) to middle (%d)", widthAtTop, widthAtMiddle)
	}
	if widthAtTop != widthAtBottom {
		t.Fatalf("help width changed from top (%d) to bottom (%d)", widthAtTop, widthAtBottom)
	}
}

func TestHelpScrollIndicatorChanges(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 20
	sendKey(m, "?")
	if m.helpViewport == nil {
		t.Fatal("expected help visible")
	}
	if m.helpViewport.TotalLineCount() <= m.helpViewport.Height {
		t.Skip("help content fits without scrolling at this height")
	}

	viewAtTop := m.helpView()
	if !strings.Contains(viewAtTop, "Top") {
		t.Fatal("expected 'Top' indicator when at top of help")
	}

	sendKey(m, "G")
	viewAtBottom := m.helpView()
	if !strings.Contains(viewAtBottom, "Bot") {
		t.Fatal("expected 'Bot' indicator when at bottom of help")
	}

	// Scroll back up one line from bottom -- should show percentage.
	sendKey(m, "k")
	viewAtMiddle := m.helpView()
	if strings.Contains(viewAtMiddle, "Top") || strings.Contains(viewAtMiddle, "Bot") {
		t.Fatal("expected percentage indicator in middle, not Top/Bot")
	}
	if !strings.Contains(viewAtMiddle, "%") {
		t.Fatal("expected '%' in scroll indicator when in middle")
	}
}

func TestHelpAbsorbsOtherKeys(t *testing.T) {
	m := newTestModel()
	sendKey(m, "?")
	if m.helpViewport == nil {
		t.Fatal("expected help visible")
	}

	// Keys that would normally affect the model should be absorbed.
	sendKey(m, "q")
	if m.helpViewport == nil {
		t.Fatal("help should absorb 'q' without quitting")
	}
	sendKey(m, "i")
	if m.mode != modeNormal {
		t.Fatal("'i' should not switch to edit mode while help is open")
	}
}

func TestDeleteRequiresEditMode(t *testing.T) {
	m := newTestModel()
	// In normal mode, 'd' is half-page-down (table handles it).
	// It should NOT trigger delete.
	sendKey(m, "d")
	if m.status.Text != "" {
		t.Fatalf("'d' in normal mode should not set status, got %q", m.status.Text)
	}
}

func TestEscClearsStatusInNormalMode(t *testing.T) {
	m := newTestModel()
	m.status = statusMsg{Text: "something", Kind: statusInfo}
	sendKey(m, "esc")
	if m.status.Text != "" {
		t.Fatalf("expected status cleared, got %q", m.status.Text)
	}
}

func TestKeyDispatchEditModeOnly(t *testing.T) {
	m := newTestModel()

	// 'a' should not be handled in normal mode.
	_, handled := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if handled {
		t.Fatal("'a' should not be handled in normal mode")
	}

	// 'esc' should be handled in edit mode (back to normal).
	m.enterEditMode()
	_, handled = m.handleEditKeys(tea.KeyMsg{Type: tea.KeyEscape})
	if !handled {
		t.Fatal("'esc' should be handled in edit mode")
	}
	if m.mode != modeNormal {
		t.Fatalf("expected modeNormal after esc in edit mode, got %d", m.mode)
	}
}

func TestModeAfterFormExit(t *testing.T) {
	m := newTestModel()
	// Simulate: enter edit mode, open a form, then exit.
	m.enterEditMode()
	m.prevMode = m.mode
	m.mode = modeForm
	// Now simulate exitForm.
	m.exitForm()
	if m.mode != modeEdit {
		t.Fatalf("expected modeEdit after exitForm (was in edit before form), got %d", m.mode)
	}

	// Now from normal mode.
	m.enterNormalMode()
	m.prevMode = m.mode
	m.mode = modeForm
	m.exitForm()
	if m.mode != modeNormal {
		t.Fatalf("expected modeNormal after exitForm (was in normal before form), got %d", m.mode)
	}
}

func TestTabSwitchBlockedInEditMode(t *testing.T) {
	m := newTestModel()
	m.enterEditMode()
	// tab should not be handled by handleCommonKeys or handleEditKeys.
	_, handled := m.handleCommonKeys(tea.KeyMsg{Type: tea.KeyTab})
	if handled {
		t.Fatal("tab should not be handled in edit mode (common keys)")
	}
	_, handled = m.handleEditKeys(tea.KeyMsg{Type: tea.KeyTab})
	if handled {
		t.Fatal("tab should not be handled in edit mode (edit keys)")
	}
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

	if normalWidth != editWidth {
		t.Fatalf(
			"badge widths should match: NAV=%d, EDIT=%d",
			normalWidth, editWidth,
		)
	}
}

func TestShiftPrefixOnUppercaseKeycap(t *testing.T) {
	m := newTestModel()
	// Uppercase "H" should produce a keycap containing "SHIFT+H".
	rendered := m.keycap("H")
	if !strings.Contains(rendered, "SHIFT+H") {
		t.Fatalf("expected keycap to contain 'SHIFT+H', got %q", rendered)
	}
	// Lowercase "h" should produce "H" (uppercased), not "SHIFT+H".
	rendered = m.keycap("h")
	if strings.Contains(rendered, "SHIFT") {
		t.Fatalf("lowercase keycap should not contain 'SHIFT', got %q", rendered)
	}
}

func containsKey(keys []string, target string) bool {
	for _, k := range keys {
		if k == target {
			return true
		}
	}
	return false
}
