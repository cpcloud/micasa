// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotePreviewOpensOnEnter(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	// Open service log detail (has Notes column).
	_ = m.openServiceLogDetail(1, "Test")
	tab := m.effectiveTab()
	require.NotNil(t, tab, "expected detail tab")

	// Seed a row with a note.
	tab.Table.SetRows(
		[]table.Row{
			{"1", "2026-01-15", "Self", "$50.00", "Changed the filter and checked pressure"},
		},
	)
	tab.Rows = []rowMeta{{ID: 1}}
	tab.CellRows = [][]cell{
		{
			{Value: "1", Kind: cellReadonly},
			{Value: "2026-01-15", Kind: cellDate},
			{Value: "Self", Kind: cellText},
			{Value: "$50.00", Kind: cellMoney},
			{Value: "Changed the filter and checked pressure", Kind: cellNotes},
		},
	}

	// Move cursor to Notes column (col 4).
	tab.ColCursor = 4

	// Press enter in Normal mode.
	sendKey(m, "enter")

	require.True(t, m.showNotePreview)
	assert.Equal(t, "Changed the filter and checked pressure", m.notePreviewText)
	assert.Equal(t, "Notes", m.notePreviewTitle)
	// Note preview overlay should be visible in the rendered view.
	view := m.buildView()
	assert.Contains(t, view, "Changed the filter and checked pressure")
	assert.Contains(t, view, "Notes")
}

func TestNotePreviewDismissesOnAnyKey(t *testing.T) {
	m := newTestModel()
	m.showNotePreview = true
	m.notePreviewText = "some note"
	m.notePreviewTitle = "Notes"

	sendKey(m, "q")

	assert.False(t, m.showNotePreview)
	assert.Empty(t, m.notePreviewText)
	// After dismissal, the note overlay should not be in the view and
	// the normal tab hints should be visible.
	view := m.buildView()
	assert.NotContains(t, view, "Press any key to close")
	assert.Contains(t, m.statusView(), "NAV")
}

func TestNotePreviewDoesNotOpenOnEmptyNote(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")
	tab := m.effectiveTab()

	tab.Table.SetRows([]table.Row{{"1", "2026-01-15", "Self", "", ""}})
	tab.Rows = []rowMeta{{ID: 1}}
	tab.CellRows = [][]cell{
		{
			{Value: "1", Kind: cellReadonly},
			{Value: "2026-01-15", Kind: cellDate},
			{Value: "Self", Kind: cellText},
			{Value: "", Kind: cellMoney},
			{Value: "", Kind: cellNotes},
		},
	}
	tab.ColCursor = 4

	sendKey(m, "enter")

	assert.False(t, m.showNotePreview)
	// Tab hints should still be visible (no overlay opened).
	assert.Contains(t, m.statusView(), "NAV")
}

func TestNotePreviewRendersInView(t *testing.T) {
	m := newTestModel()
	m.showNotePreview = true
	m.notePreviewText = "This is a test note with some content."
	m.notePreviewTitle = "Notes"

	view := m.buildView()
	assert.Contains(t, view, "This is a test note")
	assert.Contains(t, view, "Press any key to close")
}

func TestNotePreviewBlocksOtherKeys(t *testing.T) {
	m := newTestModel()
	m.showNotePreview = true
	m.notePreviewText = "test"
	initialTab := m.active

	// These should all be absorbed by the note preview.
	sendKey(m, "j")
	assert.Equal(t, initialTab, m.active, "expected tab not to change while note preview is open")
}

func TestWordWrap(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{"empty", "", 40, ""},
		{"fits", "hello world", 40, "hello world"},
		{"wraps", "hello world foo bar", 11, "hello world\nfoo bar"},
		{
			"long word",
			"superlongword fits",
			20,
			"superlongword fits",
		},
		{
			"preserves newlines",
			"line one\nline two",
			40,
			"line one\nline two",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wordWrap(tt.input, tt.width)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEnterHintShowsPreviewOnNotesColumn(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")
	tab := m.effectiveTab()
	tab.ColCursor = 4 // Notes column
	tab.Table.SetRows([]table.Row{{"1", "2026-01-15", "Self", "", "some note"}})
	tab.Rows = []rowMeta{{ID: 1}}
	tab.CellRows = [][]cell{
		{
			{Value: "1", Kind: cellReadonly},
			{Value: "2026-01-15", Kind: cellDate},
			{Value: "Self", Kind: cellText},
			{Value: "", Kind: cellMoney},
			{Value: "some note", Kind: cellNotes},
		},
	}

	hint := m.enterHint()
	assert.Equal(t, "preview", hint)
}

func TestFirstLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no newlines", "hello world", "hello world"},
		{"leading/trailing space", "  hello  ", "hello"},
		{"single newline", "line one\nline two", "line one..."},
		{"crlf", "line one\r\nline two", "line one..."},
		{"multiple newlines", "a\n\nb\n\nc", "a..."},
		{"tabs and newlines", "a\t\nb", "a..."},
		{"only whitespace", "  \n\t  ", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, firstLine(tt.input))
		})
	}
}

func TestMultilineNotesRenderedAsSingleLineInTable(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")
	tab := m.effectiveTab()
	require.NotNil(t, tab)

	multilineNote := "Changed the filter\nand checked pressure"

	tab.Table.SetRows([]table.Row{
		{"1", "2026-01-15", "Self", "$50.00", multilineNote},
	})
	tab.Rows = []rowMeta{{ID: 1}}
	tab.CellRows = [][]cell{
		{
			{Value: "1", Kind: cellReadonly},
			{Value: "2026-01-15", Kind: cellDate},
			{Value: "Self", Kind: cellText},
			{Value: "$50.00", Kind: cellMoney},
			{Value: multilineNote, Kind: cellNotes},
		},
	}

	view := m.buildView()

	// The table should NOT show a raw newline in the rendered note.
	// Instead, the two lines should be collapsed into a single line.
	assert.NotContains(t, view, "filter\nand",
		"table should not render literal newlines in notes cells")
	assert.Contains(t, view, "Changed the filter...",
		"table should show the first line of the note with ellipsis")
	assert.NotContains(t, view, "and checked pressure",
		"table should not show subsequent lines of a multi-line note")
}

func TestMultilineNotesPreservedInPreviewOverlay(t *testing.T) {
	m := newTestModel()
	m.showNotePreview = true
	m.notePreviewText = "Changed the filter\nand checked pressure"
	m.notePreviewTitle = "Notes"

	view := m.buildView()
	assert.Contains(t, view, "Changed the filter")
	assert.Contains(t, view, "and checked pressure")
}

func TestNaturalWidthsMultilineNotesFirstLine(t *testing.T) {
	specs := []columnSpec{
		{Title: "Notes", Min: 5, Max: 40, Flex: true, Kind: cellNotes},
	}
	rows := [][]cell{
		{{Value: "short\nvery long second line here", Kind: cellNotes}},
	}
	widths := naturalWidths(specs, rows)
	// Width should reflect the first line + "...", not the longer second line.
	require.Len(t, widths, 1)
	assert.Equal(t, len("short..."), widths[0])
}
