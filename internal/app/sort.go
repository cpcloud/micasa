// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/cpcloud/micasa/internal/data"
)

// toggleSort cycles the sort on colIdx: none -> asc -> desc -> none.
// If the column is already in the sort stack, it advances its direction
// or removes it. If not present, it appends as ascending.
func toggleSort(tab *Tab, colIdx int) {
	for i, entry := range tab.Sorts {
		if entry.Col == colIdx {
			if entry.Dir == sortAsc {
				tab.Sorts[i].Dir = sortDesc
			} else {
				// Was desc; remove from stack.
				tab.Sorts = append(tab.Sorts[:i], tab.Sorts[i+1:]...)
			}
			return
		}
	}
	tab.Sorts = append(tab.Sorts, sortEntry{Col: colIdx, Dir: sortAsc})
}

// clearSorts removes all sort entries from the tab.
func clearSorts(tab *Tab) {
	tab.Sorts = nil
}

// applySorts sorts the tab's rows in place using the current sort stack.
// PK (column 0) is always appended as an implicit ascending tiebreaker
// unless it's already in the stack. When the sort stack is empty this
// means rows sort by PK asc by default.
func applySorts(tab *Tab) {
	if len(tab.CellRows) <= 1 {
		return
	}

	sorts := withPKTiebreaker(tab.Sorts)

	indices := make([]int, len(tab.CellRows))
	for i := range indices {
		indices[i] = i
	}

	sort.SliceStable(indices, func(a, b int) bool {
		for _, entry := range sorts {
			va := cellValueAt(tab, indices[a], entry.Col)
			vb := cellValueAt(tab, indices[b], entry.Col)

			// Empty values always sort last, regardless of direction.
			if va == "" && vb == "" {
				continue
			}
			if va == "" {
				return false
			}
			if vb == "" {
				return true
			}

			cmp := compareCells(tab, entry.Col, indices[a], indices[b])
			if cmp == 0 {
				continue
			}
			if entry.Dir == sortDesc {
				return cmp > 0
			}
			return cmp < 0
		}
		return false
	})

	reorderTab(tab, indices)
}

// withPKTiebreaker appends a PK (col 0) ascending entry if col 0 is not
// already in the stack, ensuring a stable deterministic order.
func withPKTiebreaker(sorts []sortEntry) []sortEntry {
	for _, e := range sorts {
		if e.Col == 0 {
			return sorts
		}
	}
	return append(sorts, sortEntry{Col: 0, Dir: sortAsc})
}

// compareCells returns -1, 0, or 1 comparing row a vs row b at the given
// column. Uses type-aware comparison based on the column's cellKind.
// Empty values are handled by the caller (applySorts) to ensure they
// always sort last regardless of direction.
func compareCells(tab *Tab, col, a, b int) int {
	va := cellValueAt(tab, a, col)
	vb := cellValueAt(tab, b, col)

	if va == vb {
		return 0
	}

	kind := cellText
	if col >= 0 && col < len(tab.Specs) {
		kind = tab.Specs[col].Kind
	}

	switch kind {
	case cellMoney:
		return compareMoney(va, vb)
	case cellDate, cellUrgency, cellWarranty:
		return compareDates(va, vb)
	case cellReadonly:
		// ID columns are numeric.
		return compareNumeric(va, vb)
	case cellDrilldown:
		return compareNumeric(va, vb)
	default:
		return compareStrings(va, vb)
	}
}

func cellValueAt(tab *Tab, row, col int) string {
	if row < 0 || row >= len(tab.CellRows) {
		return ""
	}
	cells := tab.CellRows[row]
	if col < 0 || col >= len(cells) {
		return ""
	}
	return strings.TrimSpace(cells[col].Value)
}

func compareMoney(a, b string) int {
	pa := parseMoney(a)
	pb := parseMoney(b)
	if pa < pb {
		return -1
	}
	if pa > pb {
		return 1
	}
	return 0
}

func compareDates(a, b string) int {
	ta, errA := time.Parse(data.DateLayout, a)
	tb, errB := time.Parse(data.DateLayout, b)
	if errA != nil || errB != nil {
		return compareStrings(a, b)
	}
	if ta.Before(tb) {
		return -1
	}
	if ta.After(tb) {
		return 1
	}
	return 0
}

func compareNumeric(a, b string) int {
	na, errA := strconv.ParseFloat(a, 64)
	nb, errB := strconv.ParseFloat(b, 64)
	if errA != nil || errB != nil {
		return compareStrings(a, b)
	}
	if na < nb {
		return -1
	}
	if na > nb {
		return 1
	}
	return 0
}

func compareStrings(a, b string) int {
	la := strings.ToLower(a)
	lb := strings.ToLower(b)
	if la < lb {
		return -1
	}
	if la > lb {
		return 1
	}
	return 0
}

// parseMoney strips $, commas, and parses as float64. All money cell
// values come from FormatCents (always valid) and empty cells are
// filtered out before compareMoney is called, so the parse cannot fail
// in practice. Form-level validators (optionalMoney / requiredMoney)
// reject invalid input before it reaches the data layer.
func parseMoney(s string) float64 {
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// reorderTab rearranges CellRows, Rows (meta), and the table's rows
// according to the given index permutation.
func reorderTab(tab *Tab, indices []int) {
	n := len(indices)
	newCellRows := make([][]cell, n)
	newMeta := make([]rowMeta, n)
	tableRows := tab.Table.Rows()
	newTableRows := make([]table.Row, n)

	for i, idx := range indices {
		newCellRows[i] = tab.CellRows[idx]
		newMeta[i] = tab.Rows[idx]
		if idx < len(tableRows) {
			newTableRows[i] = tableRows[idx]
		}
	}

	tab.CellRows = newCellRows
	tab.Rows = newMeta
	tab.Table.SetRows(newTableRows)
}
