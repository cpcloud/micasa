// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newSortTab() *Tab {
	return &Tab{
		Specs: []columnSpec{
			{Title: "ID", Kind: cellReadonly},
			{Title: "Name", Kind: cellText},
			{Title: "Cost", Kind: cellMoney},
			{Title: "Date", Kind: cellDate},
		},
		CellRows: [][]cell{
			{
				{Value: "3", Kind: cellReadonly},
				{Value: "Charlie", Kind: cellText},
				{Value: "$200.00", Kind: cellMoney},
				{Value: "2025-03-01", Kind: cellDate},
			},
			{
				{Value: "1", Kind: cellReadonly},
				{Value: "Alice", Kind: cellText},
				{Value: "$50.00", Kind: cellMoney},
				{Value: "2025-01-15", Kind: cellDate},
			},
			{
				{Value: "2", Kind: cellReadonly},
				{Value: "Bob", Kind: cellText},
				{Value: "$1,000.00", Kind: cellMoney},
				{Value: "2025-02-10", Kind: cellDate},
			},
		},
		Rows: []rowMeta{
			{ID: 3},
			{ID: 1},
			{ID: 2},
		},
	}
}

func TestToggleSortCycle(t *testing.T) {
	tab := &Tab{}

	// none -> asc
	toggleSort(tab, 1)
	if len(tab.Sorts) != 1 || tab.Sorts[0].Dir != sortAsc || tab.Sorts[0].Col != 1 {
		t.Fatalf("expected [col=1 asc], got %+v", tab.Sorts)
	}

	// asc -> desc
	toggleSort(tab, 1)
	if len(tab.Sorts) != 1 || tab.Sorts[0].Dir != sortDesc {
		t.Fatalf("expected [col=1 desc], got %+v", tab.Sorts)
	}

	// desc -> none (removed)
	toggleSort(tab, 1)
	if len(tab.Sorts) != 0 {
		t.Fatalf("expected empty sorts, got %+v", tab.Sorts)
	}
}

func TestToggleSortMultiColumn(t *testing.T) {
	tab := &Tab{}
	toggleSort(tab, 0) // col 0 asc
	toggleSort(tab, 2) // col 2 asc

	if len(tab.Sorts) != 2 {
		t.Fatalf("expected 2 sorts, got %d", len(tab.Sorts))
	}
	if tab.Sorts[0].Col != 0 || tab.Sorts[1].Col != 2 {
		t.Fatalf("expected cols [0, 2], got %+v", tab.Sorts)
	}

	// Toggle col 0 to desc; col 2 stays asc.
	toggleSort(tab, 0)
	if tab.Sorts[0].Dir != sortDesc {
		t.Fatalf("expected col 0 desc, got %+v", tab.Sorts[0])
	}
	if tab.Sorts[1].Dir != sortAsc {
		t.Fatalf("expected col 2 still asc, got %+v", tab.Sorts[1])
	}
}

func TestClearSorts(t *testing.T) {
	tab := &Tab{}
	toggleSort(tab, 0)
	toggleSort(tab, 1)
	clearSorts(tab)
	if len(tab.Sorts) != 0 {
		t.Fatalf("expected empty sorts after clear, got %+v", tab.Sorts)
	}
}

func TestApplySortsDefaultPK(t *testing.T) {
	tab := newSortTab()
	// No explicit sorts => default PK (col 0) asc.
	applySorts(tab)

	ids := collectIDs(tab)
	expected := []uint{1, 2, 3}
	if !equalIDs(ids, expected) {
		t.Fatalf("expected IDs %v, got %v", expected, ids)
	}
}

func TestApplySortsByNameAsc(t *testing.T) {
	tab := newSortTab()
	toggleSort(tab, 1) // Name asc
	applySorts(tab)

	names := collectCol(tab, 1)
	expected := []string{"Alice", "Bob", "Charlie"}
	if !equalStrings(names, expected) {
		t.Fatalf("expected %v, got %v", expected, names)
	}
}

func TestApplySortsByNameDesc(t *testing.T) {
	tab := newSortTab()
	toggleSort(tab, 1) // Name asc
	toggleSort(tab, 1) // Name desc
	applySorts(tab)

	names := collectCol(tab, 1)
	expected := []string{"Charlie", "Bob", "Alice"}
	if !equalStrings(names, expected) {
		t.Fatalf("expected %v, got %v", expected, names)
	}
}

func TestApplySortsByMoneyAsc(t *testing.T) {
	tab := newSortTab()
	toggleSort(tab, 2) // Cost asc
	applySorts(tab)

	costs := collectCol(tab, 2)
	expected := []string{"$50.00", "$200.00", "$1,000.00"}
	if !equalStrings(costs, expected) {
		t.Fatalf("expected %v, got %v", expected, costs)
	}
}

func TestApplySortsByDateDesc(t *testing.T) {
	tab := newSortTab()
	toggleSort(tab, 3) // Date asc
	toggleSort(tab, 3) // Date desc
	applySorts(tab)

	dates := collectCol(tab, 3)
	expected := []string{"2025-03-01", "2025-02-10", "2025-01-15"}
	if !equalStrings(dates, expected) {
		t.Fatalf("expected %v, got %v", expected, dates)
	}
}

func TestApplySortsEmptyLastRegardlessOfDirection(t *testing.T) {
	tab := &Tab{
		Specs: []columnSpec{
			{Title: "Name", Kind: cellText},
		},
		CellRows: [][]cell{
			{{Value: "", Kind: cellText}},
			{{Value: "Bravo", Kind: cellText}},
			{{Value: "Alpha", Kind: cellText}},
		},
		Rows: []rowMeta{{ID: 1}, {ID: 2}, {ID: 3}},
	}
	toggleSort(tab, 0) // asc
	applySorts(tab)

	names := collectCol(tab, 0)
	if names[2] != "" {
		t.Fatalf("expected empty value last, got %v", names)
	}

	// Now desc: empty should still be last.
	toggleSort(tab, 0) // desc
	applySorts(tab)

	names = collectCol(tab, 0)
	if names[2] != "" {
		t.Fatalf("expected empty value last in desc, got %v", names)
	}
}

func TestApplySortsMultiKey(t *testing.T) {
	tab := &Tab{
		Specs: []columnSpec{
			{Title: "Group", Kind: cellText},
			{Title: "Name", Kind: cellText},
		},
		CellRows: [][]cell{
			{{Value: "B", Kind: cellText}, {Value: "Zara", Kind: cellText}},
			{{Value: "A", Kind: cellText}, {Value: "Yuri", Kind: cellText}},
			{{Value: "A", Kind: cellText}, {Value: "Alex", Kind: cellText}},
			{{Value: "B", Kind: cellText}, {Value: "Mia", Kind: cellText}},
		},
		Rows: []rowMeta{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}},
	}
	toggleSort(tab, 0) // Group asc (primary)
	toggleSort(tab, 1) // Name asc (secondary)
	applySorts(tab)

	names := collectCol(tab, 1)
	expected := []string{"Alex", "Yuri", "Mia", "Zara"}
	if !equalStrings(names, expected) {
		t.Fatalf("expected %v, got %v", expected, names)
	}
}

func TestSortIndicatorSingle(t *testing.T) {
	sorts := []sortEntry{{Col: 2, Dir: sortAsc}}
	if got := sortIndicator(sorts, 2); got != "\u25b2" {
		t.Fatalf("expected ▲ (no number for single sort), got %q", got)
	}
}

func TestSortIndicatorMulti(t *testing.T) {
	sorts := []sortEntry{
		{Col: 2, Dir: sortAsc},
		{Col: 5, Dir: sortDesc},
	}
	if got := sortIndicator(sorts, 2); got != "\u25b21" {
		t.Fatalf("expected ▲1, got %q", got)
	}
	if got := sortIndicator(sorts, 5); got != "\u25bc2" {
		t.Fatalf("expected ▼2, got %q", got)
	}
	if got := sortIndicator(sorts, 0); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestPKTiebreaker(t *testing.T) {
	// Col 0 not in stack: gets appended.
	sorts := []sortEntry{{Col: 2, Dir: sortAsc}}
	result := withPKTiebreaker(sorts)
	if len(result) != 2 || result[1].Col != 0 || result[1].Dir != sortAsc {
		t.Fatalf("expected PK appended, got %+v", result)
	}

	// Col 0 already in stack: unchanged.
	sorts = []sortEntry{{Col: 0, Dir: sortDesc}, {Col: 3, Dir: sortAsc}}
	result = withPKTiebreaker(sorts)
	if len(result) != 2 {
		t.Fatalf("expected no append when PK present, got %+v", result)
	}
}

func TestSortKeyOnlyInNormalMode(t *testing.T) {
	m := newTestModel()
	m.enterEditMode()
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	_, handled := m.handleEditKeys(key)
	if handled {
		t.Fatal("s should not be handled in Edit mode")
	}
}

// helpers

func collectIDs(tab *Tab) []uint {
	ids := make([]uint, len(tab.Rows))
	for i, m := range tab.Rows {
		ids[i] = m.ID
	}
	return ids
}

func collectCol(tab *Tab, col int) []string {
	vals := make([]string, len(tab.CellRows))
	for i, row := range tab.CellRows {
		if col < len(row) {
			vals[i] = row[col].Value
		}
	}
	return vals
}

func equalIDs(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
