// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/cpcloud/micasa/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NotEmpty(t, lines, "buildView returned empty output")
	assert.Contains(t, lines[0], "╭", "first line should contain the top border")
}

func TestBuildViewShowsTerminalTooSmallMessage(t *testing.T) {
	m := newTestModel()
	m.width = minUsableWidth - 1
	m.height = minUsableHeight - 1
	m.showDashboard = true
	m.showNotePreview = true

	output := m.buildView()
	assert.Contains(t, output, "Terminal too small")
	assert.Contains(t, output, "need at least 80x24")
}

func TestBuildViewDoesNotShowTerminalTooSmallMessageAtMinimumSize(t *testing.T) {
	m := newTestModel()
	m.width = minUsableWidth
	m.height = minUsableHeight

	output := m.buildView()
	assert.NotContains(t, output, "Terminal too small")
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
	assert.Greater(t, natural[1], 12)
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
	assert.GreaterOrEqual(t, widths[1], 16)
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
	assert.LessOrEqual(t, total, 17)
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
	widths := columnWidths(specs, rows, 60, 3)
	assert.GreaterOrEqual(t, widths[1], 15)
}

func TestWidenTruncated(t *testing.T) {
	t.Run("distributes all extra space", func(t *testing.T) {
		widths := []int{4, 10, 8}
		remaining := widenTruncated(widths, []int{4, 15, 8}, 3)
		assert.Equal(t, 13, widths[1])
		assert.Equal(t, 0, remaining)
	})
	t.Run("caps at natural width", func(t *testing.T) {
		widths := []int{4, 10, 8}
		remaining := widenTruncated(widths, []int{4, 12, 8}, 5)
		assert.Equal(t, 12, widths[1])
		assert.Equal(t, 3, remaining)
	})
}

// --- Column visibility tests ---

func TestNextVisibleCol(t *testing.T) {
	tests := []struct {
		name    string
		specs   []columnSpec
		from    int
		forward bool
		want    int
	}{
		{
			"forward skips hidden",
			[]columnSpec{{Title: "A"}, {Title: "B", HideOrder: 1}, {Title: "C"}, {Title: "D"}},
			0,
			true,
			2,
		},
		{
			"backward skips hidden",
			[]columnSpec{{Title: "A"}, {Title: "B", HideOrder: 1}, {Title: "C"}, {Title: "D"}},
			2,
			false,
			0,
		},
		{
			"clamps forward at edge",
			[]columnSpec{{Title: "A"}, {Title: "B", HideOrder: 1}, {Title: "C", HideOrder: 2}},
			0,
			true,
			0,
		},
		{
			"clamps backward at edge",
			[]columnSpec{{Title: "A"}, {Title: "B", HideOrder: 1}, {Title: "C", HideOrder: 2}},
			0,
			false,
			0,
		},
		{"all visible forward", []columnSpec{{Title: "A"}, {Title: "B"}, {Title: "C"}}, 1, true, 2},
		{
			"clamps at right edge",
			[]columnSpec{{Title: "A"}, {Title: "B"}, {Title: "C"}},
			2,
			true,
			2,
		},
		{
			"clamps at left edge",
			[]columnSpec{{Title: "A"}, {Title: "B"}, {Title: "C"}},
			0,
			false,
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, nextVisibleCol(tt.specs, tt.from, tt.forward))
		})
	}
}

func TestFirstVisibleCol(t *testing.T) {
	specs := []columnSpec{
		{Title: "A", HideOrder: 1}, {Title: "B"}, {Title: "C"}, {Title: "D"},
	}
	assert.Equal(t, 1, firstVisibleCol(specs))
}

func TestLastVisibleCol(t *testing.T) {
	specs := []columnSpec{
		{Title: "A"}, {Title: "B"}, {Title: "C"}, {Title: "D", HideOrder: 1},
	}
	assert.Equal(t, 2, lastVisibleCol(specs))
}

func TestVisibleCount(t *testing.T) {
	specs := []columnSpec{
		{Title: "A"}, {Title: "B", HideOrder: 1}, {Title: "C"},
	}
	assert.Equal(t, 2, visibleCount(specs))
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
	require.Len(t, specs, 2)
	assert.Equal(t, "ID", specs[0].Title)
	assert.Equal(t, "Status", specs[1].Title)
	require.Len(t, cells[0], 2)
	assert.Equal(t, "1", cells[0][0].Value)
	assert.Equal(t, "active", cells[0][1].Value)
	assert.Equal(t, 1, cursor)
	require.Len(t, sorts, 1)
	assert.Equal(t, 1, sorts[0].Col)
	assert.Equal(t, []int{0, 2}, visToFull)
}

func TestVisibleProjectionHiddenCursor(t *testing.T) {
	tab := &Tab{
		Specs:     []columnSpec{{Title: "A"}, {Title: "B", HideOrder: 1}},
		CellRows:  [][]cell{{{Value: "1"}, {Value: "2"}}},
		ColCursor: 1,
	}
	_, _, cursor, _, _ := visibleProjection(tab)
	assert.Equal(t, -1, cursor)
}

func TestVisibleProjectionHiddenSortOmitted(t *testing.T) {
	tab := &Tab{
		Specs:    []columnSpec{{Title: "A"}, {Title: "B", HideOrder: 1}},
		CellRows: [][]cell{{{Value: "1"}, {Value: "2"}}},
		Sorts:    []sortEntry{{Col: 1, Dir: sortAsc}},
	}
	_, _, _, sorts, _ := visibleProjection(tab)
	assert.Empty(t, sorts)
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
	assert.Equal(t, 0, tab.Specs[0].HideOrder, "should not hide the last visible column")
	assert.Equal(t, statusError, m.status.Kind)
}

func TestHideCurrentColumnMovesToNext(t *testing.T) {
	m := newTestModel()
	m.mode = modeNormal
	tab := m.effectiveTab()
	tab.ColCursor = 0
	m.hideCurrentColumn()
	assert.NotZero(t, tab.Specs[0].HideOrder, "expected column 0 to be hidden")
	assert.Equal(
		t,
		0,
		tab.Specs[tab.ColCursor].HideOrder,
		"cursor should be on a visible column after hiding",
	)
}

func TestShowAllColumns(t *testing.T) {
	m := newTestModel()
	tab := m.effectiveTab()
	tab.Specs[1].HideOrder = 1
	tab.Specs[2].HideOrder = 2
	m.showAllColumns()
	for i, s := range tab.Specs {
		assert.Equalf(t, 0, s.HideOrder, "expected column %d to be visible", i)
	}
}

func TestJoinCells(t *testing.T) {
	t.Run("per-gap separators", func(t *testing.T) {
		assert.Equal(
			t,
			"A | B \u22ef C",
			joinCells([]string{"A", "B", "C"}, []string{" | ", " \u22ef "}),
		)
	})
	t.Run("fallback separator", func(t *testing.T) {
		assert.Equal(t, "A | B | C", joinCells([]string{"A", "B", "C"}, []string{" | "}))
	})
}

func TestGapSeparators(t *testing.T) {
	t.Run("detects collapsed gaps", func(t *testing.T) {
		normal := "\u2502"
		styles := DefaultStyles()
		plainSeps, collapsedSeps := gapSeparators([]int{0, 3, 4}, 5, normal, styles)
		require.Len(t, collapsedSeps, 2)
		assert.NotEqual(t, normal, collapsedSeps[0], "first gap should be collapsed separator")
		assert.Equal(t, normal, collapsedSeps[1], "second gap should be normal separator")
		assert.Equal(t, normal, plainSeps[0])
		assert.Equal(t, normal, plainSeps[1])
	})
	t.Run("single column returns empty", func(t *testing.T) {
		plainSeps, collapsedSeps := gapSeparators([]int{2}, 5, "\u2502", DefaultStyles())
		assert.Empty(t, plainSeps)
		assert.Empty(t, collapsedSeps)
	})
}

func TestHiddenColumnNames(t *testing.T) {
	t.Run("returns hidden names in order", func(t *testing.T) {
		specs := []columnSpec{
			{Title: "ID"},
			{Title: "Name", HideOrder: 1},
			{Title: "Status"},
			{Title: "Cost", HideOrder: 2},
		}
		assert.Equal(t, []string{"Name", "Cost"}, hiddenColumnNames(specs))
	})
	t.Run("empty when none hidden", func(t *testing.T) {
		assert.Empty(t, hiddenColumnNames([]columnSpec{{Title: "A"}, {Title: "B"}}))
	})
}

func TestNextHideOrder(t *testing.T) {
	specs := []columnSpec{
		{Title: "A", HideOrder: 3},
		{Title: "B"},
		{Title: "C", HideOrder: 1},
	}
	assert.Equal(t, 4, nextHideOrder(specs))
}

func TestRenderHiddenBadges(t *testing.T) {
	styles := DefaultStyles()
	t.Run("empty when none hidden", func(t *testing.T) {
		specs := []columnSpec{{Title: "A"}, {Title: "B"}}
		assert.Empty(t, renderHiddenBadges(specs, 0, styles))
	})
	t.Run("left only", func(t *testing.T) {
		specs := []columnSpec{{Title: "ID", HideOrder: 1}, {Title: "Name"}, {Title: "Status"}}
		assert.Contains(t, renderHiddenBadges(specs, 2, styles), "ID")
	})
	t.Run("right only", func(t *testing.T) {
		specs := []columnSpec{{Title: "ID"}, {Title: "Name"}, {Title: "Cost", HideOrder: 1}}
		assert.Contains(t, renderHiddenBadges(specs, 0, styles), "Cost")
	})
	t.Run("both sides", func(t *testing.T) {
		specs := []columnSpec{
			{Title: "ID", HideOrder: 1},
			{Title: "Name"},
			{Title: "Cost", HideOrder: 2},
		}
		out := renderHiddenBadges(specs, 1, styles)
		assert.Contains(t, out, "ID")
		assert.Contains(t, out, "Cost")
	})
}

func TestRenderHiddenBadgesStableWidthAcrossCursorMoves(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID", HideOrder: 1},
		{Title: "Name"},
		{Title: "Cost", HideOrder: 2},
		{Title: "Status"},
	}
	styles := DefaultStyles()

	start := renderHiddenBadges(specs, 0, styles)
	middle := renderHiddenBadges(specs, 1, styles)
	end := renderHiddenBadges(specs, 3, styles)

	startW := lipgloss.Width(start)
	middleW := lipgloss.Width(middle)
	endW := lipgloss.Width(end)
	assert.Equal(t, startW, middleW, "start vs middle badge width")
	assert.Equal(t, middleW, endW, "middle vs end badge width")
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
	widths := columnWidths(specs, rows, 80, 3)
	assert.GreaterOrEqual(t, widths[0], 9)
}

// --- Line clamping tests ---

func TestClampLines(t *testing.T) {
	t.Run("truncates long line", func(t *testing.T) {
		assert.Equal(t, "hell…", clampLines("hello world", 5))
	})
	t.Run("multiline truncates only long lines", func(t *testing.T) {
		got := clampLines("short\na very long line here\nok", 8)
		lines := strings.Split(got, "\n")
		require.Len(t, lines, 3)
		assert.Equal(t, "short", lines[0])
		assert.Equal(t, "ok", lines[2])
		assert.Less(t, len(lines[1]), len("a very long line here"))
	})
	t.Run("noop when fits", func(t *testing.T) {
		assert.Equal(t, "fits", clampLines("fits", 100))
	})
}

func TestTruncateLeft(t *testing.T) {
	t.Run("truncates long path", func(t *testing.T) {
		got := truncateLeft("/home/user/long/path/to/data.db", 15)
		assert.True(t, strings.HasPrefix(got, "\u2026"))
		assert.True(t, strings.HasSuffix(got, "data.db"))
		assert.LessOrEqual(t, lipgloss.Width(got), 15)
	})
	t.Run("noop when fits", func(t *testing.T) {
		assert.Equal(t, "short.db", truncateLeft("short.db", 20))
	})
	t.Run("grapheme clusters", func(t *testing.T) {
		got := truncateLeft("\U0001F1EF\U0001F1F5/path/to/file.db", 15)
		assert.LessOrEqual(t, lipgloss.Width(got), 15)
		assert.True(t, strings.HasPrefix(got, "\u2026"))
	})
	t.Run("zero and negative width", func(t *testing.T) {
		assert.Empty(t, truncateLeft("anything", 0))
		assert.Empty(t, truncateLeft("anything", -1))
	})
}

// --- Viewport tests ---

func TestViewportAllColumnsFit(t *testing.T) {
	widths := []int{10, 15, 10}
	start, end, hasL, hasR := viewportRange(widths, 3, 50, 0, 0)
	assert.Equal(t, 0, start)
	assert.Equal(t, 3, end)
	assert.False(t, hasL)
	assert.False(t, hasR)
}

func TestViewportScrollsRight(t *testing.T) {
	widths := []int{10, 10, 10, 10, 10}
	start, end, hasL, _ := viewportRange(widths, 3, 30, 0, 3)
	assert.LessOrEqual(t, start, 3, "start should be <= cursor")
	assert.Greater(t, end, 3, "end should be > cursor")
	assert.True(t, hasL, "expected left indicator when scrolled right")
}

func TestViewportScrollsLeftOnCursorMove(t *testing.T) {
	tab := &Tab{ViewOffset: 3}
	ensureCursorVisible(tab, 1, 5)
	widths := []int{10, 10, 10, 10, 10}
	start, end, _, _ := viewportRange(widths, 3, 30, tab.ViewOffset, 1)
	assert.LessOrEqual(t, start, 1)
	assert.Greater(t, end, 1)
}

func TestEnsureCursorVisibleClamps(t *testing.T) {
	tab := &Tab{ViewOffset: 5}
	ensureCursorVisible(tab, 2, 4)
	assert.LessOrEqual(t, tab.ViewOffset, 2)
}

func TestEnsureCursorVisibleNoopWhenVisible(t *testing.T) {
	tab := &Tab{ViewOffset: 0}
	ensureCursorVisible(tab, 2, 5)
	assert.Equal(t, 0, tab.ViewOffset)
}

func TestViewportSorts(t *testing.T) {
	t.Run("adjusts column indices by offset", func(t *testing.T) {
		adjusted := viewportSorts([]sortEntry{{Col: 3, Dir: sortAsc}, {Col: 5, Dir: sortDesc}}, 2)
		assert.Equal(t, 1, adjusted[0].Col)
		assert.Equal(t, 3, adjusted[1].Col)
	})
	t.Run("no offset passthrough", func(t *testing.T) {
		adjusted := viewportSorts([]sortEntry{{Col: 1, Dir: sortAsc}}, 0)
		assert.Equal(t, 1, adjusted[0].Col)
		assert.Equal(t, sortAsc, adjusted[0].Dir)
	})
}

func TestApplianceAge(t *testing.T) {
	now := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		purchase *time.Time
		want     string
	}{
		{"nil purchase", nil, ""},
		{"less than a month", ptr(time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)), "<1m"},
		{"a few months", ptr(time.Date(2025, 10, 5, 0, 0, 0, 0, time.UTC)), "4m"},
		{"one year exact", ptr(time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC)), "1y"},
		{"years and months", ptr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)), "2y 7m"},
		{"future date", ptr(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, applianceAge(tt.purchase, now))
		})
	}
}

func ptr[T any](v T) *T { return &v }

func TestNavBadgeLabel(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	status := m.statusView()
	assert.Contains(t, status, "NAV")
	assert.NotContains(t, status, "NORMAL")
}

func TestStatusViewProjectStatusSummaryOnlyOnProjectsTab(t *testing.T) {
	m := newTestModel()
	m.width = 160
	m.height = 40

	status := m.statusView()
	for _, label := range []string{"status", "all"} {
		assert.Contains(t, status, label, "expected %q summary on projects tab", label)
	}

	m.active = tabIndex(tabQuotes)
	status = m.statusView()
	for _, label := range []string{"status", "all", "filters"} {
		assert.NotContains(
			t,
			status,
			label,
			"did not expect %q project status summary on non-project tab",
			label,
		)
	}
}

func TestStatusViewProjectStatusSummaryReflectsActiveFilters(t *testing.T) {
	m := newTestModel()
	m.width = 160
	m.height = 40
	tab := m.activeTab()
	require.NotNil(t, tab)
	require.Equal(t, tabProjects, tab.Kind, "expected projects tab to be active")

	tab.HideCompleted = true
	status := m.statusView()
	assert.Contains(
		t,
		status,
		"no completed",
		"expected no completed summary when completed is hidden",
	)
	assert.NotContains(t, status, "no abandoned")

	tab.HideAbandoned = true
	status = m.statusView()
	assert.Contains(t, status, "settled", "expected settled summary when both filters are active")
}

func TestStatusViewUsesMoreLabelWhenHintsCollapse(t *testing.T) {
	m := newTestModel()
	m.width = 70
	m.height = 40
	status := m.statusView()
	assert.Contains(t, status, "more", "expected collapsed hint label to include more")
	assert.NotContains(
		t,
		status,
		"find col",
		"did not expect low-priority find hint after collapse",
	)
}

func TestHelpContentIncludesProjectStatusFilterShortcuts(t *testing.T) {
	m := newTestModel()
	help := m.helpContent()
	for _, snippet := range []string{
		"Hide/show completed projects",
		"Hide/show abandoned projects",
		"Hide/show settled projects",
	} {
		assert.Contains(t, help, snippet)
	}
}

func TestHeaderTitleWidth(t *testing.T) {
	tests := []struct {
		name string
		spec columnSpec
		want int
	}{
		{
			"link",
			columnSpec{Title: "Project", Link: &columnLink{TargetTab: tabProjects}},
			lipgloss.Width("Project") + 1 + lipgloss.Width(linkArrow),
		},
		{
			"drilldown",
			columnSpec{Title: "Log", Kind: cellDrilldown},
			lipgloss.Width("Log") + 1 + lipgloss.Width(drilldownArrow),
		},
		{"plain", columnSpec{Title: "Name"}, lipgloss.Width("Name")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, headerTitleWidth(tt.spec))
		})
	}
}

func TestDimBackgroundNeutralizesCancelFaint(t *testing.T) {
	// Simulate a composited overlay: cancelFaint injects \033[22m (normal
	// intensity) so the overlay content stays bright. A subsequent
	// dimBackground pass must neutralize those markers so the entire
	// background dims uniformly (nested overlay scenario).
	inner := cancelFaint("dashboard content")
	assert.Contains(t, inner, "\033[22m", "cancelFaint should inject normal-intensity")

	dimmed := dimBackground(inner)
	assert.NotContains(t, dimmed, "\033[22m",
		"dimBackground should neutralize cancel-faint markers from nested overlays")
	assert.Contains(t, dimmed, "\033[2m", "dimBackground should apply faint")
}
