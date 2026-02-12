// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/cpcloud/micasa/internal/data"
)

func TestBuildViewShowsFullHouseBox(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.hasHouse = true
	m.house = data.HouseProfile{Nickname: "Test House"}

	output := m.buildView()
	lines := strings.Split(output, "\n")

	// The rounded border top-left corner must be on the first line.
	if len(lines) == 0 {
		t.Fatal("buildView returned empty output")
	}
	if !strings.Contains(lines[0], "╭") {
		t.Fatalf("first line should contain the top border (╭), got: %q", lines[0])
	}
}

func TestNaturalWidthsIgnoreMax(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID", Min: 4, Max: 6},
		{Title: "Name", Min: 8, Max: 12},
	}
	rows := [][]cell{
		{{Value: "1"}, {Value: "A very long name indeed"}},
	}
	natural := naturalWidths(specs, rows)
	// "A very long name indeed" is 23 chars, well past Max of 12.
	if natural[1] <= 12 {
		t.Fatalf("expected natural width > Max (12), got %d", natural[1])
	}
}

func TestColumnWidthsNoTruncationWhenRoomAvailable(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID", Min: 4, Max: 6},
		{Title: "Name", Min: 8, Max: 12},
	}
	rows := [][]cell{
		{{Value: "1"}, {Value: "A long name here"}},
	}
	// "A long name here" = 16 chars, exceeds Max=12.
	// With 200 width and 3 separator, natural widths should fit.
	widths := columnWidths(specs, rows, 200, 3)
	if widths[1] < 16 {
		t.Fatalf(
			"expected Name column >= 16 (content width), got %d",
			widths[1],
		)
	}
}

func TestColumnWidthsTruncatesWhenTerminalNarrow(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID", Min: 4, Max: 6},
		{Title: "Name", Min: 8, Max: 12, Flex: true},
	}
	rows := [][]cell{
		{{Value: "1"}, {Value: "A very long name indeed"}},
	}
	// Very narrow terminal: 20 total - 3 separator = 17 available.
	widths := columnWidths(specs, rows, 20, 3)
	total := widths[0] + widths[1]
	if total > 17 {
		t.Fatalf("expected total widths <= 17, got %d", total)
	}
}

func TestColumnWidthsTruncatedColumnsGetExtraFirst(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID", Min: 4, Max: 6},
		{Title: "Name", Min: 8, Max: 10},
		{Title: "Desc", Min: 8, Max: 10, Flex: true},
	}
	rows := [][]cell{
		{{Value: "1"}, {Value: "Fifteen chars!!"}, {Value: "short"}},
	}
	// Natural: ID=4, Name=15, Desc=8 = 27 total.
	// Available: 60 - 6 (two separators of 3) = 54.
	// Natural fits (27 < 54), so no truncation needed.
	widths := columnWidths(specs, rows, 60, 3)
	if widths[1] < 15 {
		t.Fatalf(
			"expected Name >= 15 (no truncation when room available), got %d",
			widths[1],
		)
	}
}

func TestWidenTruncated(t *testing.T) {
	widths := []int{4, 10, 8}
	natural := []int{4, 15, 8}
	remaining := widenTruncated(widths, natural, 3)
	// Should widen column 1 from 10 to 13 (3 extra given).
	if widths[1] != 13 {
		t.Fatalf("expected widths[1]=13 after widening, got %d", widths[1])
	}
	if remaining != 0 {
		t.Fatalf("expected 0 remaining, got %d", remaining)
	}
}

func TestWidenTruncatedCapsAtNatural(t *testing.T) {
	widths := []int{4, 10, 8}
	natural := []int{4, 12, 8}
	remaining := widenTruncated(widths, natural, 5)
	// Column 1 needs 2 more to reach natural. 5 - 2 = 3 remaining.
	if widths[1] != 12 {
		t.Fatalf("expected widths[1]=12 (natural), got %d", widths[1])
	}
	if remaining != 3 {
		t.Fatalf("expected 3 remaining, got %d", remaining)
	}
}

// --- Column visibility tests ---

func TestNextVisibleColForward(t *testing.T) {
	specs := []columnSpec{
		{Title: "A"}, {Title: "B", HideOrder: 1}, {Title: "C"}, {Title: "D"},
	}
	got := nextVisibleCol(specs, 0, true)
	if got != 2 {
		t.Fatalf("expected 2 (skip hidden B), got %d", got)
	}
}

func TestNextVisibleColBackward(t *testing.T) {
	specs := []columnSpec{
		{Title: "A"}, {Title: "B", HideOrder: 1}, {Title: "C"}, {Title: "D"},
	}
	got := nextVisibleCol(specs, 2, false)
	if got != 0 {
		t.Fatalf("expected 0 (skip hidden B), got %d", got)
	}
}

func TestNextVisibleColClampsAtEdge(t *testing.T) {
	specs := []columnSpec{
		{Title: "A"}, {Title: "B", HideOrder: 1}, {Title: "C", HideOrder: 2},
	}
	// Only A is visible; forward from A should stay at A (clamp).
	got := nextVisibleCol(specs, 0, true)
	if got != 0 {
		t.Fatalf("expected 0 (clamp, only A visible), got %d", got)
	}
	// Backward from A should also stay at A.
	got = nextVisibleCol(specs, 0, false)
	if got != 0 {
		t.Fatalf("expected 0 (clamp backward, only A visible), got %d", got)
	}
}

func TestNextVisibleColAllVisible(t *testing.T) {
	specs := []columnSpec{{Title: "A"}, {Title: "B"}, {Title: "C"}}
	got := nextVisibleCol(specs, 1, true)
	if got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestNextVisibleColClampsRight(t *testing.T) {
	specs := []columnSpec{{Title: "A"}, {Title: "B"}, {Title: "C"}}
	got := nextVisibleCol(specs, 2, true)
	if got != 2 {
		t.Fatalf("expected 2 (clamp at right edge), got %d", got)
	}
}

func TestNextVisibleColClampsLeftAllVisible(t *testing.T) {
	specs := []columnSpec{{Title: "A"}, {Title: "B"}, {Title: "C"}}
	got := nextVisibleCol(specs, 0, false)
	if got != 0 {
		t.Fatalf("expected 0 (clamp at left edge), got %d", got)
	}
}

func TestFirstVisibleCol(t *testing.T) {
	specs := []columnSpec{
		{Title: "A", HideOrder: 1}, {Title: "B"}, {Title: "C"}, {Title: "D"},
	}
	if got := firstVisibleCol(specs); got != 1 {
		t.Fatalf("expected 1 (A hidden), got %d", got)
	}
}

func TestLastVisibleCol(t *testing.T) {
	specs := []columnSpec{
		{Title: "A"}, {Title: "B"}, {Title: "C"}, {Title: "D", HideOrder: 1},
	}
	if got := lastVisibleCol(specs); got != 2 {
		t.Fatalf("expected 2 (D hidden), got %d", got)
	}
}

func TestVisibleCount(t *testing.T) {
	specs := []columnSpec{
		{Title: "A"}, {Title: "B", HideOrder: 1}, {Title: "C"},
	}
	if n := visibleCount(specs); n != 2 {
		t.Fatalf("expected 2 visible, got %d", n)
	}
}

func TestVisibleProjectionSkipsHidden(t *testing.T) {
	tab := &Tab{
		Specs: []columnSpec{
			{Title: "ID"}, {Title: "Name", HideOrder: 1}, {Title: "Status"},
		},
		CellRows: [][]cell{
			{{Value: "1"}, {Value: "Test"}, {Value: "active"}},
		},
		ColCursor: 2,
		Sorts:     []sortEntry{{Col: 2, Dir: sortAsc}},
	}
	specs, cells, cursor, sorts, visToFull := visibleProjection(tab)
	if len(specs) != 2 {
		t.Fatalf("expected 2 visible specs, got %d", len(specs))
	}
	if specs[0].Title != "ID" || specs[1].Title != "Status" {
		t.Fatalf("unexpected spec titles: %v, %v", specs[0].Title, specs[1].Title)
	}
	if len(cells[0]) != 2 {
		t.Fatalf("expected 2 visible cells per row, got %d", len(cells[0]))
	}
	if cells[0][0].Value != "1" || cells[0][1].Value != "active" {
		t.Fatalf("unexpected cell values: %v, %v", cells[0][0].Value, cells[0][1].Value)
	}
	if cursor != 1 {
		t.Fatalf("expected visible cursor 1 (Status), got %d", cursor)
	}
	if len(sorts) != 1 || sorts[0].Col != 1 {
		t.Fatalf("expected remapped sort on vis col 1, got %v", sorts)
	}
	if len(visToFull) != 2 || visToFull[0] != 0 || visToFull[1] != 2 {
		t.Fatalf("unexpected visToFull: %v", visToFull)
	}
}

func TestVisibleProjectionHiddenCursor(t *testing.T) {
	tab := &Tab{
		Specs:     []columnSpec{{Title: "A"}, {Title: "B", HideOrder: 1}},
		CellRows:  [][]cell{{{Value: "1"}, {Value: "2"}}},
		ColCursor: 1,
	}
	_, _, cursor, _, _ := visibleProjection(tab)
	if cursor != -1 {
		t.Fatalf("expected cursor -1 for hidden column, got %d", cursor)
	}
}

func TestVisibleProjectionHiddenSortOmitted(t *testing.T) {
	tab := &Tab{
		Specs:    []columnSpec{{Title: "A"}, {Title: "B", HideOrder: 1}},
		CellRows: [][]cell{{{Value: "1"}, {Value: "2"}}},
		Sorts:    []sortEntry{{Col: 1, Dir: sortAsc}},
	}
	_, _, _, sorts, _ := visibleProjection(tab)
	if len(sorts) != 0 {
		t.Fatalf("expected hidden sort to be omitted, got %v", sorts)
	}
}

func TestHideCurrentColumnPreventsLastVisible(t *testing.T) {
	m := newTestModel()
	m.mode = modeNormal
	tab := m.effectiveTab()
	// Hide all but one
	for i := 1; i < len(tab.Specs); i++ {
		tab.Specs[i].HideOrder = i
	}
	tab.ColCursor = 0
	m.hideCurrentColumn()
	if tab.Specs[0].HideOrder > 0 {
		t.Fatal("should not hide the last visible column")
	}
	if m.status.Kind != statusError {
		t.Fatal("expected error status when hiding last column")
	}
}

func TestHideCurrentColumnMovesToNext(t *testing.T) {
	m := newTestModel()
	m.mode = modeNormal
	tab := m.effectiveTab()
	tab.ColCursor = 0
	m.hideCurrentColumn()
	if tab.Specs[0].HideOrder == 0 {
		t.Fatal("expected column 0 to be hidden")
	}
	if tab.Specs[tab.ColCursor].HideOrder > 0 {
		t.Fatal("cursor should be on a visible column after hiding")
	}
}

func TestShowAllColumns(t *testing.T) {
	m := newTestModel()
	tab := m.effectiveTab()
	tab.Specs[1].HideOrder = 1
	tab.Specs[2].HideOrder = 2
	m.showAllColumns()
	for i, s := range tab.Specs {
		if s.HideOrder > 0 {
			t.Fatalf("expected column %d to be visible", i)
		}
	}
}

func TestJoinCellsPerGapSeparators(t *testing.T) {
	cells := []string{"A", "B", "C"}
	seps := []string{" | ", " ⋯ "}
	got := joinCells(cells, seps)
	want := "A | B ⋯ C"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestJoinCellsFallbackSeparator(t *testing.T) {
	cells := []string{"A", "B", "C"}
	seps := []string{" | "}
	got := joinCells(cells, seps)
	want := "A | B | C"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestGapSeparatorsDetectsCollapsedGaps(t *testing.T) {
	// visToFull [0, 3, 4]: gap between 0→3 has hidden cols, 3→4 doesn't.
	visToFull := []int{0, 3, 4}
	normal := "│"
	styles := DefaultStyles()
	plainSeps, collapsedSeps := gapSeparators(visToFull, 5, normal, styles)
	if len(collapsedSeps) != 2 {
		t.Fatalf("expected 2 gap separators, got %d", len(collapsedSeps))
	}
	// First gap should be collapsed (contains ⋯), second normal.
	if collapsedSeps[0] == normal {
		t.Fatal("first gap should be collapsed separator")
	}
	if collapsedSeps[1] != normal {
		t.Fatal("second gap should be normal separator")
	}
	// Plain seps should all be normal.
	if plainSeps[0] != normal || plainSeps[1] != normal {
		t.Fatal("plain seps should all be normal")
	}
}

func TestGapSeparatorsSingleColumn(t *testing.T) {
	plainSeps, collapsedSeps := gapSeparators([]int{2}, 5, "│", DefaultStyles())
	if len(plainSeps) != 0 || len(collapsedSeps) != 0 {
		t.Fatal("single visible column should have no gap separators")
	}
}

func TestHiddenColumnNames(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID"},
		{Title: "Name", HideOrder: 1},
		{Title: "Status"},
		{Title: "Cost", HideOrder: 2},
	}
	names := hiddenColumnNames(specs)
	if len(names) != 2 || names[0] != "Name" || names[1] != "Cost" {
		t.Fatalf("expected [Name, Cost], got %v", names)
	}
}

func TestHiddenColumnNamesNoneHidden(t *testing.T) {
	specs := []columnSpec{{Title: "A"}, {Title: "B"}}
	names := hiddenColumnNames(specs)
	if len(names) != 0 {
		t.Fatalf("expected empty, got %v", names)
	}
}

func TestNextHideOrder(t *testing.T) {
	specs := []columnSpec{
		{Title: "A", HideOrder: 3},
		{Title: "B"},
		{Title: "C", HideOrder: 1},
	}
	got := nextHideOrder(specs)
	if got != 4 {
		t.Fatalf("expected 4, got %d", got)
	}
}

func TestRenderHiddenBadgesEmpty(t *testing.T) {
	specs := []columnSpec{{Title: "A"}, {Title: "B"}}
	out := renderHiddenBadges(specs, 0, DefaultStyles())
	if out != "" {
		t.Fatalf("expected empty when nothing hidden, got %q", out)
	}
}

func TestRenderHiddenBadgesLeftOnly(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID", HideOrder: 1},
		{Title: "Name"},
		{Title: "Status"},
	}
	out := renderHiddenBadges(specs, 2, DefaultStyles())
	if !strings.Contains(out, "ID") {
		t.Fatalf("expected 'ID' in left badges, got %q", out)
	}
}

func TestRenderHiddenBadgesRightOnly(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID"},
		{Title: "Name"},
		{Title: "Cost", HideOrder: 1},
	}
	out := renderHiddenBadges(specs, 0, DefaultStyles())
	if !strings.Contains(out, "Cost") {
		t.Fatalf("expected 'Cost' in right badges, got %q", out)
	}
}

func TestRenderHiddenBadgesBothSides(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID", HideOrder: 1},
		{Title: "Name"},
		{Title: "Cost", HideOrder: 2},
	}
	out := renderHiddenBadges(specs, 1, DefaultStyles())
	if !strings.Contains(out, "ID") || !strings.Contains(out, "Cost") {
		t.Fatalf("expected both 'ID' and 'Cost', got %q", out)
	}
}

func TestColumnWidthsFixedValuesStillStabilize(t *testing.T) {
	specs := []columnSpec{
		{Title: "Status", Min: 8, Max: 12, FixedValues: []string{
			"ideating", "planned", "underway", "completed", "abandoned",
		}},
	}
	rows := [][]cell{
		{{Value: "planned"}},
	}
	// Even with only "planned" displayed, column should be wide enough
	// for the longest fixed value ("abandoned" = 9, "completed" = 9).
	widths := columnWidths(specs, rows, 80, 3)
	if widths[0] < 9 {
		t.Fatalf("expected width >= 9 (longest fixed value), got %d", widths[0])
	}
}

// --- Line clamping tests ---

func TestClampLinesBasic(t *testing.T) {
	got := clampLines("hello world", 5)
	if got != "hell…" {
		t.Fatalf("expected %q, got %q", "hell…", got)
	}
}

func TestClampLinesMultiline(t *testing.T) {
	input := "short\na very long line here\nok"
	got := clampLines(input, 8)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "short" {
		t.Fatalf("short line should be unchanged, got %q", lines[0])
	}
	if lines[2] != "ok" {
		t.Fatalf("short line should be unchanged, got %q", lines[2])
	}
	// The middle line should be truncated.
	if len(lines[1]) >= len("a very long line here") {
		t.Fatalf("long line should be truncated, got %q", lines[1])
	}
}

func TestClampLinesNoopWhenFits(t *testing.T) {
	input := "fits"
	got := clampLines(input, 100)
	if got != input {
		t.Fatalf("expected unchanged %q, got %q", input, got)
	}
}

func TestTruncateLeftBasic(t *testing.T) {
	got := truncateLeft("/home/user/long/path/to/data.db", 15)
	if !strings.HasPrefix(got, "…") {
		t.Fatalf("expected leading …, got %q", got)
	}
	if !strings.HasSuffix(got, "data.db") {
		t.Fatalf("expected to end with data.db, got %q", got)
	}
	if lipgloss.Width(got) > 15 {
		t.Fatalf("expected width <= 15, got %q (width %d)", got, lipgloss.Width(got))
	}
}

func TestTruncateLeftNoopWhenFits(t *testing.T) {
	input := "short.db"
	got := truncateLeft(input, 20)
	if got != input {
		t.Fatalf("expected unchanged %q, got %q", input, got)
	}
}

// --- Viewport tests ---

func TestViewportAllColumnsFit(t *testing.T) {
	widths := []int{10, 15, 10}
	start, end, hasL, hasR := viewportRange(widths, 3, 50, 0, 0)
	if start != 0 || end != 3 {
		t.Fatalf("expected full range [0,3), got [%d,%d)", start, end)
	}
	if hasL || hasR {
		t.Fatal("should have no scroll indicators when everything fits")
	}
}

func TestViewportScrollsRight(t *testing.T) {
	// 5 columns of 10 each + 4 seps of 3 = 62 total.
	// Terminal width 30 fits ~2 columns.
	widths := []int{10, 10, 10, 10, 10}
	// Cursor on column 3, offset starts at 0.
	start, end, hasL, hasR := viewportRange(widths, 3, 30, 0, 3)
	if start > 3 {
		t.Fatalf("start should be <= cursor (3), got %d", start)
	}
	if end <= 3 {
		t.Fatalf("end should be > cursor (3), got %d", end)
	}
	if !hasL {
		t.Fatal("expected left indicator when scrolled right")
	}
	_ = hasR // may or may not have right depending on fit
}

func TestViewportScrollsLeftOnCursorMove(t *testing.T) {
	// Simulate: user was scrolled to offset 3, then moved cursor left to 1.
	// ensureCursorVisible should pull ViewOffset back to cursor.
	tab := &Tab{ViewOffset: 3}
	ensureCursorVisible(tab, 1, 5)
	widths := []int{10, 10, 10, 10, 10}
	start, end, _, _ := viewportRange(widths, 3, 30, tab.ViewOffset, 1)
	if start > 1 {
		t.Fatalf("start should be <= cursor (1), got %d", start)
	}
	if end <= 1 {
		t.Fatalf("end should be > cursor (1), got %d", end)
	}
}

func TestEnsureCursorVisibleClamps(t *testing.T) {
	tab := &Tab{ViewOffset: 5}
	ensureCursorVisible(tab, 2, 4)
	if tab.ViewOffset > 2 {
		t.Fatalf("expected ViewOffset <= cursor (2), got %d", tab.ViewOffset)
	}
}

func TestEnsureCursorVisibleNoopWhenVisible(t *testing.T) {
	tab := &Tab{ViewOffset: 0}
	ensureCursorVisible(tab, 2, 5)
	if tab.ViewOffset != 0 {
		t.Fatalf("expected ViewOffset unchanged at 0, got %d", tab.ViewOffset)
	}
}

func TestViewportSortsAdjustsOffset(t *testing.T) {
	sorts := []sortEntry{{Col: 3, Dir: sortAsc}, {Col: 5, Dir: sortDesc}}
	adjusted := viewportSorts(sorts, 2)
	if adjusted[0].Col != 1 || adjusted[1].Col != 3 {
		t.Fatalf("expected cols [1,3], got [%d,%d]", adjusted[0].Col, adjusted[1].Col)
	}
}

func TestViewportSortsNoOffset(t *testing.T) {
	sorts := []sortEntry{{Col: 1, Dir: sortAsc}}
	adjusted := viewportSorts(sorts, 0)
	// When offset is 0, indices should be unchanged.
	if adjusted[0].Col != 1 {
		t.Fatalf("expected col 1, got %d", adjusted[0].Col)
	}
	if adjusted[0].Dir != sortAsc {
		t.Fatalf("expected sortAsc, got %v", adjusted[0].Dir)
	}
}

func TestApplianceAge(t *testing.T) {
	now := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		purchase *time.Time
		want     string
	}{
		{"nil purchase", nil, ""},
		{"less than a month", ptr(time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)), "<1 mo"},
		{"a few months", ptr(time.Date(2025, 10, 5, 0, 0, 0, 0, time.UTC)), "4 mo"},
		{"one year exact", ptr(time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC)), "1y"},
		{"years and months", ptr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)), "2y 7m"},
		{"future date", ptr(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applianceAge(tt.purchase, now)
			if got != tt.want {
				t.Errorf("applianceAge(%v, %v) = %q, want %q", tt.purchase, now, got, tt.want)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }

func TestNavBadgeLabel(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	status := m.statusView()
	if !strings.Contains(status, "NAV") {
		t.Error("expected NAV badge in status bar")
	}
	if strings.Contains(status, "NORMAL") {
		t.Error("status bar should not contain NORMAL anymore")
	}
}

func TestStatusViewProjectStatusSummaryOnlyOnProjectsTab(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40

	status := m.statusView()
	for _, label := range []string{"status", "all"} {
		if !strings.Contains(status, label) {
			t.Fatalf("expected %q summary on projects tab", label)
		}
	}

	m.active = tabIndex(tabQuotes)
	status = m.statusView()
	for _, label := range []string{"status", "all", "filters"} {
		if strings.Contains(status, label) {
			t.Fatalf("did not expect %q project status summary on non-project tab", label)
		}
	}
}

func TestStatusViewProjectStatusSummaryReflectsActiveFilters(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	tab := m.activeTab()
	if tab == nil || tab.Kind != tabProjects {
		t.Fatal("expected projects tab to be active")
	}

	tab.HideCompleted = true
	status := m.statusView()
	if !strings.Contains(status, "no completed") {
		t.Fatalf("expected no completed summary when completed is hidden, got %q", status)
	}
	if strings.Contains(status, "no abandoned") {
		t.Fatalf("did not expect no abandoned summary, got %q", status)
	}

	tab.HideAbandoned = true
	status = m.statusView()
	if !strings.Contains(status, "settled") {
		t.Fatalf("expected settled summary when both filters are active, got %q", status)
	}
}

func TestStatusViewUsesMoreLabelWhenHintsCollapse(t *testing.T) {
	m := newTestModel()
	m.width = 70
	m.height = 40
	status := m.statusView()
	if !strings.Contains(status, "more") {
		t.Fatalf("expected collapsed hint label to include more, got %q", status)
	}
	if strings.Contains(status, "find col") {
		t.Fatalf("did not expect low-priority find hint after collapse, got %q", status)
	}
}

func TestHelpContentIncludesProjectStatusFilterShortcuts(t *testing.T) {
	m := newTestModel()
	help := m.helpContent()
	for _, snippet := range []string{
		"Hide/show completed projects",
		"Hide/show abandoned projects",
		"Hide/show settled projects",
	} {
		if !strings.Contains(help, snippet) {
			t.Fatalf("expected help content to include %q", snippet)
		}
	}
}

func TestHeaderTitleWidthLink(t *testing.T) {
	spec := columnSpec{
		Title: "Project",
		Link:  &columnLink{TargetTab: tabProjects},
	}
	w := headerTitleWidth(spec)
	// "Project" (7) + " " (1) + "→" (1) = 9
	expected := lipgloss.Width("Project") + 1 + lipgloss.Width(linkArrow)
	if w != expected {
		t.Fatalf("expected width %d, got %d", expected, w)
	}
}

func TestHeaderTitleWidthDrilldown(t *testing.T) {
	spec := columnSpec{
		Title: "Log",
		Kind:  cellDrilldown,
	}
	w := headerTitleWidth(spec)
	expected := lipgloss.Width("Log") + 1 + lipgloss.Width(drilldownArrow)
	if w != expected {
		t.Fatalf("expected width %d, got %d", expected, w)
	}
}

func TestHeaderTitleWidthPlain(t *testing.T) {
	spec := columnSpec{Title: "Name"}
	w := headerTitleWidth(spec)
	if w != lipgloss.Width("Name") {
		t.Fatalf("expected width %d, got %d", lipgloss.Width("Name"), w)
	}
}
