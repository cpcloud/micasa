// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newFilterTab() *Tab {
	specs := []columnSpec{
		{Title: "ID", Kind: cellReadonly},
		{Title: "Status", Kind: cellStatus},
		{Title: "Vendor", Kind: cellText},
	}
	cellRows := [][]cell{
		{{Value: "1"}, {Value: "Plan"}, {Value: "Alice"}},
		{{Value: "2"}, {Value: "Active"}, {Value: "Bob"}},
		{{Value: "3"}, {Value: "Plan"}, {Value: "Bob"}},
		{{Value: "4"}, {Value: "Done"}, {Value: "Alice"}},
	}
	rows := make([]table.Row, len(cellRows))
	meta := make([]rowMeta, len(cellRows))
	for i, cr := range cellRows {
		r := make(table.Row, len(cr))
		for j, c := range cr {
			r[j] = c.Value
		}
		rows[i] = r
		meta[i] = rowMeta{ID: uint(i + 1)} //nolint:gosec // i bounded by slice length
	}
	cols := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Status", Width: 8},
		{Title: "Vendor", Width: 8},
	}
	tbl := table.New(table.WithColumns(cols), table.WithRows(rows))
	return &Tab{
		Specs:        specs,
		CellRows:     cellRows,
		Rows:         meta,
		FullRows:     rows,
		FullMeta:     meta,
		FullCellRows: cellRows,
		Table:        tbl,
	}
}

func TestTogglePinAddsAndRemoves(t *testing.T) {
	tab := newFilterTab()

	pinned := togglePin(tab, 1, "Plan")
	assert.True(t, pinned, "first toggle should pin")
	assert.True(t, hasPins(tab))
	assert.True(t, isPinned(tab, 1, "Plan"))

	unpinned := togglePin(tab, 1, "Plan")
	assert.False(t, unpinned, "second toggle should unpin")
	assert.False(t, hasPins(tab), "pins should be empty after unpin")
}

func TestTogglePinMultipleValuesInColumn(t *testing.T) {
	tab := newFilterTab()
	togglePin(tab, 1, "Plan")
	togglePin(tab, 1, "Active")

	require.Len(t, tab.Pins, 1, "same column should have one pin entry")
	assert.True(t, isPinned(tab, 1, "Plan"))
	assert.True(t, isPinned(tab, 1, "Active"))
	assert.Len(t, tab.Pins[0].Values, 2)
}

func TestTogglePinCaseInsensitive(t *testing.T) {
	tab := newFilterTab()
	togglePin(tab, 1, "plan")
	assert.True(t, isPinned(tab, 1, "Plan"))
	assert.True(t, isPinned(tab, 1, "PLAN"))
}

func TestClearPins(t *testing.T) {
	tab := newFilterTab()
	togglePin(tab, 1, "Plan")
	tab.FilterActive = true
	clearPins(tab)
	assert.False(t, hasPins(tab))
	assert.False(t, tab.FilterActive)
}

func TestClearPinsForColumn(t *testing.T) {
	tab := newFilterTab()
	togglePin(tab, 1, "Plan")
	togglePin(tab, 2, "Alice")

	clearPinsForColumn(tab, 1)
	assert.True(t, hasPins(tab), "column 2 pin should remain")
	assert.False(t, isPinned(tab, 1, "Plan"))
	assert.True(t, isPinned(tab, 2, "Alice"))
}

func TestClearPinsForColumnClearsFilterWhenEmpty(t *testing.T) {
	tab := newFilterTab()
	togglePin(tab, 1, "Plan")
	tab.FilterActive = true
	clearPinsForColumn(tab, 1)
	assert.False(t, tab.FilterActive)
}

func TestMatchesAllPinsSingleColumn(t *testing.T) {
	pins := []filterPin{{Col: 1, Values: map[string]bool{"plan": true, "active": true}}}
	row1 := []cell{{Value: "1"}, {Value: "Plan"}, {Value: "Alice"}}
	row2 := []cell{{Value: "4"}, {Value: "Done"}, {Value: "Alice"}}

	assert.True(t, matchesAllPins(row1, pins, false), "Plan should match")
	assert.False(t, matchesAllPins(row2, pins, false), "Done should not match")
}

func TestMatchesAllPinsCrossColumn(t *testing.T) {
	pins := []filterPin{
		{Col: 1, Values: map[string]bool{"plan": true}},
		{Col: 2, Values: map[string]bool{"bob": true}},
	}
	// Row 3: Plan + Bob => match
	row3 := []cell{{Value: "3"}, {Value: "Plan"}, {Value: "Bob"}}
	// Row 1: Plan + Alice => no match (fails vendor pin)
	row1 := []cell{{Value: "1"}, {Value: "Plan"}, {Value: "Alice"}}

	assert.True(t, matchesAllPins(row3, pins, false))
	assert.False(t, matchesAllPins(row1, pins, false))
}

func TestApplyRowFilterNoPin(t *testing.T) {
	tab := newFilterTab()
	applyRowFilter(tab, false)
	assert.Len(t, tab.CellRows, 4)
	assert.Len(t, tab.Rows, 4)
	for _, m := range tab.Rows {
		assert.False(t, m.Dimmed)
	}
}

func TestApplyRowFilterPreview(t *testing.T) {
	tab := newFilterTab()
	togglePin(tab, 1, "Plan")
	applyRowFilter(tab, false)

	require.Len(t, tab.CellRows, 4, "preview keeps all rows")
	assert.False(t, tab.Rows[0].Dimmed, "row 0 matches Plan")
	assert.True(t, tab.Rows[1].Dimmed, "row 1 is Active, should be dimmed")
	assert.False(t, tab.Rows[2].Dimmed, "row 2 matches Plan")
	assert.True(t, tab.Rows[3].Dimmed, "row 3 is Done, should be dimmed")
}

func TestApplyRowFilterActive(t *testing.T) {
	tab := newFilterTab()
	togglePin(tab, 1, "Plan")
	tab.FilterActive = true
	applyRowFilter(tab, false)

	require.Len(t, tab.CellRows, 2, "active filter hides non-matching")
	assert.Equal(t, "1", tab.CellRows[0][0].Value)
	assert.Equal(t, "3", tab.CellRows[1][0].Value)
}

func TestApplyRowFilterActiveAcrossColumns(t *testing.T) {
	tab := newFilterTab()
	togglePin(tab, 1, "Plan")
	togglePin(tab, 2, "Bob")
	tab.FilterActive = true
	applyRowFilter(tab, false)

	require.Len(t, tab.CellRows, 1, "only row 3 matches Plan AND Bob")
	assert.Equal(t, "3", tab.CellRows[0][0].Value)
}

func TestPinSummary(t *testing.T) {
	tab := newFilterTab()
	togglePin(tab, 1, "Plan")
	s := pinSummary(tab)
	assert.Contains(t, s, "Status")
	assert.Contains(t, s, "plan")
}

func TestPinSummaryEmpty(t *testing.T) {
	tab := newFilterTab()
	assert.Equal(t, "", pinSummary(tab))
}

func TestEagerModeToggleWithNoPins(t *testing.T) {
	tab := newFilterTab()
	assert.False(t, tab.FilterActive)

	// Toggle on with no pins.
	tab.FilterActive = !tab.FilterActive
	assert.True(t, tab.FilterActive, "eager mode should be armed")

	// Now pin while eager mode is on -- filter should immediately hide rows.
	togglePin(tab, 1, "Plan")
	applyRowFilter(tab, false)
	require.Len(t, tab.CellRows, 2, "eager mode + pin should immediately filter")
	assert.Equal(t, "1", tab.CellRows[0][0].Value)
	assert.Equal(t, "3", tab.CellRows[1][0].Value)
}

func TestEagerModeToggleOff(t *testing.T) {
	tab := newFilterTab()
	tab.FilterActive = true
	togglePin(tab, 1, "Plan")
	applyRowFilter(tab, false)
	require.Len(t, tab.CellRows, 2)

	// Toggle off -- should restore all rows with preview dimming.
	tab.FilterActive = false
	applyRowFilter(tab, false)
	require.Len(t, tab.CellRows, 4, "toggling off should restore all rows")
	assert.True(t, tab.Rows[1].Dimmed, "non-matching rows should be dimmed in preview")
}

func TestPinsPersistAcrossTabSwitch(t *testing.T) {
	m := newTestModel()
	m.mode = modeNormal
	startTab := m.active
	tab := &m.tabs[startTab]

	// Set up Full* so applyRowFilter works.
	tab.FullRows = tab.Table.Rows()
	tab.FullMeta = tab.Rows
	tab.FullCellRows = tab.CellRows

	togglePin(tab, 0, "1")
	tab.FilterActive = true
	require.True(t, hasPins(tab))

	// Switch away and back.
	sendKey(m, "f")
	assert.NotEqual(t, startTab, m.active, "should switch tabs freely")
	sendKey(m, "b")
	assert.Equal(t, startTab, m.active)

	// Pins and filter state should still be there.
	tab = &m.tabs[startTab]
	assert.True(t, hasPins(tab), "pins should persist across tab switch")
	assert.True(t, tab.FilterActive, "filter mode should persist across tab switch")
}

func TestHideColumnClearsPinsOnThatColumn(t *testing.T) {
	m := newTestModel()
	m.mode = modeNormal
	tab := m.effectiveTab()
	tab.ColCursor = 0

	// Set up Full* fields so applyRowFilter doesn't panic.
	tab.FullRows = tab.Table.Rows()
	tab.FullMeta = tab.Rows
	tab.FullCellRows = tab.CellRows

	togglePin(tab, 0, "1")
	require.True(t, hasPins(tab))

	m.hideCurrentColumn()
	assert.False(t, isPinned(tab, 0, "1"), "hiding the column should clear its pins")
}

func TestMatchesAllPinsMagMode(t *testing.T) {
	// In mag mode, $50 -> round(log10(50)) = round(1.7) = 2
	// and $1,000 -> round(log10(1000)) = 3
	row50 := []cell{{Value: "$50.00", Kind: cellMoney}}
	row1k := []cell{{Value: "$1,000.00", Kind: cellMoney}}
	row200 := []cell{{Value: "$200.00", Kind: cellMoney}}

	// Pin on magnitude "2" (covers $50 and $200, both round to mag 2).
	pins := []filterPin{{Col: 0, Values: map[string]bool{magArrow + "2": true}}}

	assert.True(t, matchesAllPins(row50, pins, true), "$50 is mag 2")
	assert.True(t, matchesAllPins(row200, pins, true), "$200 is mag 2")
	assert.False(t, matchesAllPins(row1k, pins, true), "$1000 is mag 3, not 2")

	// Without mag mode, magnitude pins don't match raw values.
	assert.False(t, matchesAllPins(row50, pins, false), "mag pin shouldn't match raw value")
}

func TestTranslatePinsToMag(t *testing.T) {
	// Pin raw "$50.00", translate to mag mode -> should become mag "🠡2".
	tab := &Tab{
		Specs: []columnSpec{{Title: "Cost", Kind: cellMoney}},
		FullCellRows: [][]cell{
			{{Value: "$50.00", Kind: cellMoney}},
			{{Value: "$1,000.00", Kind: cellMoney}},
			{{Value: "$200.00", Kind: cellMoney}},
		},
	}
	togglePin(tab, 0, "$50.00")
	translatePins(tab, true) // switching TO mag

	require.True(t, hasPins(tab))
	// $50 -> mag 2
	assert.True(t, tab.Pins[0].Values[magArrow+"2"], "should have mag 2")
	assert.False(t, tab.Pins[0].Values["$50.00"], "raw value should be gone")
}

func TestTranslatePinsFromMag(t *testing.T) {
	// Pin mag "🠡3", translate from mag mode -> should expand to all mag-3 raw values.
	tab := &Tab{
		Specs: []columnSpec{{Title: "Cost", Kind: cellMoney}},
		FullCellRows: [][]cell{
			{{Value: "$50.00", Kind: cellMoney}},
			{{Value: "$1,000.00", Kind: cellMoney}},
			{{Value: "$2,000.00", Kind: cellMoney}},
		},
	}
	togglePin(tab, 0, magArrow+"3")
	translatePins(tab, false) // switching FROM mag

	require.True(t, hasPins(tab))
	// Both $1,000 and $2,000 are mag 3.
	assert.True(t, tab.Pins[0].Values["$1,000.00"], "$1,000 should match")
	assert.True(t, tab.Pins[0].Values["$2,000.00"], "$2,000 should match")
	assert.False(t, tab.Pins[0].Values["$50.00"], "$50 is mag 2, not 3")
}

func TestTranslatePinsRoundTrip(t *testing.T) {
	// Pin $1,000.00 -> toggle to mag (🠡3) -> toggle back -> should get
	// $1,000.00 AND $2,000.00 (both mag 3).
	tab := &Tab{
		Specs: []columnSpec{{Title: "Cost", Kind: cellMoney}},
		FullCellRows: [][]cell{
			{{Value: "$1,000.00", Kind: cellMoney}},
			{{Value: "$2,000.00", Kind: cellMoney}},
			{{Value: "$50.00", Kind: cellMoney}},
		},
	}
	togglePin(tab, 0, "$1,000.00")

	// To mag: $1,000 -> 🠡3
	translatePins(tab, true)
	require.Len(t, tab.Pins[0].Values, 1)
	assert.True(t, tab.Pins[0].Values[magArrow+"3"])

	// Back to raw: 🠡3 -> $1,000 and $2,000
	translatePins(tab, false)
	assert.True(t, tab.Pins[0].Values["$1,000.00"])
	assert.True(t, tab.Pins[0].Values["$2,000.00"])
	assert.Len(t, tab.Pins[0].Values, 2)
}

// seedTabForPinning sets up CellRows and Full* fields on the active tab so
// that sendKey("n") can pin the cell under the cursor.
func seedTabForPinning(m *Model) *Tab {
	m.showDashboard = false
	tab := m.effectiveTab()
	rows := tab.Table.Rows()
	cellRows := make([][]cell, len(rows))
	for i, r := range rows {
		cr := make([]cell, len(r))
		for j, v := range r {
			kind := cellText
			if j < len(tab.Specs) {
				kind = tab.Specs[j].Kind
			}
			cr[j] = cell{Value: v, Kind: kind}
		}
		cellRows[i] = cr
	}
	tab.CellRows = cellRows
	tab.FullRows = rows
	tab.FullMeta = tab.Rows
	tab.FullCellRows = cellRows
	return tab
}

func TestCtrlNClearsAllPinsAndFilter(t *testing.T) {
	m := newTestModel()
	m.mode = modeNormal
	tab := seedTabForPinning(m)

	// Pin via key press, then activate filter.
	sendKey(m, "n")
	require.True(t, hasPins(tab), "n should pin the cell under cursor")
	sendKey(m, "N")
	require.True(t, tab.FilterActive)

	// ctrl+n should clear everything.
	sendKey(m, keyCtrlN)
	assert.False(t, hasPins(tab), "ctrl+n should clear all pins")
	assert.False(t, tab.FilterActive, "ctrl+n should deactivate filter")
}

func TestCtrlNNoopWithoutPins(t *testing.T) {
	m := newTestModel()
	m.mode = modeNormal
	m.showDashboard = false

	sendKey(m, keyCtrlN)
	tab := m.effectiveTab()
	assert.False(t, hasPins(tab), "ctrl+n with no pins should be a no-op")
	assert.False(t, tab.FilterActive)
}

func TestPinOnDashboardBlocked(t *testing.T) {
	m := newTestModel()
	m.showDashboard = true
	startPins := len(m.effectiveTab().Pins)

	sendKey(m, "n")
	assert.Len(t, m.effectiveTab().Pins, startPins, "n should be blocked on dashboard")

	sendKey(m, "N")
	assert.False(t, m.effectiveTab().FilterActive, "N should be blocked on dashboard")
}
