// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"strings"
	"testing"

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

func TestNextVisibleColWraps(t *testing.T) {
	specs := []columnSpec{
		{Title: "A"}, {Title: "B", HideOrder: 1}, {Title: "C", HideOrder: 2},
	}
	got := nextVisibleCol(specs, 0, true)
	if got != 0 {
		t.Fatalf("expected 0 (wrap around, only A visible), got %d", got)
	}
}

func TestNextVisibleColAllVisible(t *testing.T) {
	specs := []columnSpec{{Title: "A"}, {Title: "B"}, {Title: "C"}}
	got := nextVisibleCol(specs, 1, true)
	if got != 2 {
		t.Fatalf("expected 2, got %d", got)
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
