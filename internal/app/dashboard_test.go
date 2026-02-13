// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/cpcloud/micasa/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaysUntil(t *testing.T) {
	now := time.Date(2026, 2, 8, 14, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		target time.Time
		want   int
	}{
		{"same day", time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC), 0},
		{"tomorrow", time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC), 1},
		{"yesterday", time.Date(2026, 2, 7, 0, 0, 0, 0, time.UTC), -1},
		{"30 days ahead", time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC), 30},
		{"10 days ago", time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC), -10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, daysUntil(now, tt.target))
		})
	}
}

func TestDaysText(t *testing.T) {
	tests := []struct {
		days    int
		overdue bool
		want    string
	}{
		{0, true, "today"},
		{0, false, "today"},
		{-5, true, "5 days overdue"},
		{-1, true, "1 day overdue"},
		{3, false, "in 3 days"},
		{1, false, "in 1 day"},
	}
	for _, tt := range tests {
		assert.Equalf(t, tt.want, daysText(tt.days, tt.overdue),
			"daysText(%d, %v)", tt.days, tt.overdue)
	}
}

func TestSortByDays(t *testing.T) {
	items := []maintenanceUrgency{
		{DaysFromNow: 10},
		{DaysFromNow: -5},
		{DaysFromNow: 2},
		{DaysFromNow: -20},
	}
	sortByDays(items)
	for i := 1; i < len(items); i++ {
		assert.GreaterOrEqualf(t, items[i].DaysFromNow, items[i-1].DaysFromNow,
			"not sorted: items[%d]=%d < items[%d]=%d",
			i, items[i].DaysFromNow, i-1, items[i-1].DaysFromNow)
	}
}

func TestCapSlice(t *testing.T) {
	assert.Len(t, capSlice([]int{1, 2, 3, 4, 5}, 3), 3)
	assert.Len(t, capSlice([]int{1, 2}, 5), 2)
	assert.Empty(t, capSlice([]int{1, 2, 3}, -1))
}

func TestDashboardToggle(t *testing.T) {
	m := newTestModel()
	m.showDashboard = false

	sendKey(m, "D")
	assert.True(t, m.showDashboard)
	sendKey(m, "D")
	assert.False(t, m.showDashboard)
}

func TestDashboardDismissedByTabSwitch(t *testing.T) {
	for _, key := range []string{"f", "b"} {
		t.Run(key, func(t *testing.T) {
			m := newTestModel()
			m.showDashboard = true
			sendKey(m, key)
			assert.False(t, m.showDashboard)
		})
	}
}

func TestDashboardNavigation(t *testing.T) {
	m := newTestModel()
	m.showDashboard = true
	// Populate nav with 5 entries.
	m.dashNav = []dashNavEntry{
		{Tab: tabMaintenance, ID: 1},
		{Tab: tabMaintenance, ID: 2},
		{Tab: tabProjects, ID: 3},
		{Tab: tabAppliances, ID: 4},
		{Tab: tabMaintenance, ID: 5},
	}
	m.dashCursor = 0

	// j moves down.
	sendKey(m, "j")
	assert.Equal(t, 1, m.dashCursor)
	// k moves up.
	sendKey(m, "k")
	assert.Equal(t, 0, m.dashCursor)
	// k at 0 stays at 0 (no wrap).
	sendKey(m, "k")
	assert.Equal(t, 0, m.dashCursor)
	// G jumps to bottom.
	sendKey(m, "G")
	assert.Equal(t, 4, m.dashCursor)
	// j at bottom stays at bottom.
	sendKey(m, "j")
	assert.Equal(t, 4, m.dashCursor)
	// g jumps to top.
	sendKey(m, "g")
	assert.Equal(t, 0, m.dashCursor)
}

func TestDashboardJump(t *testing.T) {
	m := newTestModel()
	m.showDashboard = true
	m.dashNav = []dashNavEntry{
		{Tab: tabMaintenance, ID: 1},
		{Tab: tabProjects, ID: 42},
	}
	m.dashCursor = 1

	m.dashJump()
	assert.False(t, m.showDashboard)
	assert.Equal(t, tabIndex(tabProjects), m.active)
}

func TestDashboardEnterKeyJumps(t *testing.T) {
	m := newTestModel()
	m.showDashboard = true
	m.dashNav = []dashNavEntry{
		{Tab: tabMaintenance, ID: 1},
		{Tab: tabProjects, ID: 42},
	}
	m.dashCursor = 1

	sendKey(m, "enter")
	assert.False(t, m.showDashboard)
	assert.Equal(t, tabIndex(tabProjects), m.active)
}

func TestDashboardBlocksTableKeys(t *testing.T) {
	m := newTestModel()
	m.showDashboard = true
	m.dashNav = []dashNavEntry{{Tab: tabMaintenance, ID: 1}}
	m.dashCursor = 0
	startTab := m.active

	// h/l should not move column cursor.
	colBefore := m.tabs[m.active].ColCursor
	sendKey(m, "l")
	assert.Equal(t, colBefore, m.tabs[m.active].ColCursor, "l should be blocked on dashboard")
	sendKey(m, "h")
	assert.Equal(t, colBefore, m.tabs[m.active].ColCursor, "h should be blocked on dashboard")

	// s should not add a sort.
	sortsBefore := len(m.tabs[m.active].Sorts)
	sendKey(m, "s")
	assert.Equal(t, sortsBefore, len(m.tabs[m.active].Sorts), "s should be blocked on dashboard")

	// i should not enter edit mode.
	sendKey(m, "i")
	assert.Equal(t, modeNormal, m.mode, "i should be blocked on dashboard")

	// tab should not toggle house profile.
	houseBefore := m.showHouse
	sendKey(m, "tab")
	assert.Equal(t, houseBefore, m.showHouse, "tab should be blocked on dashboard")

	// Dashboard should still be showing.
	assert.True(t, m.showDashboard)
	assert.Equal(t, startTab, m.active)
}

func TestDashboardViewEmptySections(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.dashboard = dashboardData{}
	m.dashNav = nil
	m.dashCursor = 0

	view := m.dashboardView(50, 120)
	// Empty dashboard returns empty string -- silence is success.
	assert.Empty(t, view)
}

func TestDashboardViewWithData(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	now := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	overdueDue := time.Date(2026, 1, 25, 0, 0, 0, 0, time.UTC)
	lastSrv := time.Date(2025, 10, 25, 0, 0, 0, 0, time.UTC)

	m.dashboard = dashboardData{
		Overdue: []maintenanceUrgency{{
			Item: data.MaintenanceItem{
				Name:           "HVAC Filter",
				LastServicedAt: &lastSrv,
			},
			NextDue:     overdueDue,
			DaysFromNow: daysUntil(now, overdueDue),
		}},
		ActiveProjects: []data.Project{{
			Title:  "Kitchen Remodel",
			Status: data.ProjectStatusInProgress,
		}},
		ServiceSpendCents: 50000,
		ProjectSpendCents: 100000,
	}
	m.buildDashNav()

	view := m.dashboardView(50, 120)
	assert.Contains(t, view, "HVAC Filter")
	assert.Contains(t, view, "Kitchen Remodel")
	assert.Contains(t, view, "overdue")
	assert.Contains(t, view, "$500.00")
}

func TestDashboardViewFitsOverlayWidth(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 40

	// Build data with long names that would wrap without truncation.
	now := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	overdueDue := time.Date(2026, 1, 25, 0, 0, 0, 0, time.UTC)
	lastSrv := time.Date(2025, 10, 25, 0, 0, 0, 0, time.UTC)

	m.dashboard = dashboardData{
		Overdue: []maintenanceUrgency{{
			Item: data.MaintenanceItem{
				Name:           "Refrigerator coil cleaning and deep inspection",
				LastServicedAt: &lastSrv,
			},
			NextDue:       overdueDue,
			DaysFromNow:   daysUntil(now, overdueDue),
			ApplianceName: "Kitchen Refrigerator",
		}},
		ActiveProjects: []data.Project{{
			Title:  "Kitchen countertop upgrade with premium materials",
			Status: data.ProjectStatusInProgress,
		}},
	}
	m.buildDashNav()

	innerW := m.overlayContentWidth() - 4

	view := m.dashboardView(50, innerW)
	for i, line := range strings.Split(view, "\n") {
		w := lipgloss.Width(line)
		assert.LessOrEqual(t, w, innerW,
			"line %d width %d exceeds overlay inner width %d: %q",
			i, w, innerW, line)
	}
}

func TestDashboardOverlay(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.showDashboard = true
	m.dashboard = dashboardData{}
	m.dashNav = nil

	ov := m.buildDashboardOverlay()
	today := time.Now().Format("Monday, Jan 2")
	assert.Contains(t, ov, today)
	assert.Contains(t, ov, "help")
}

func TestDashboardOverlayFitsHeight(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 30
	m.showDashboard = true

	// Lots of data to stress the height budget.
	overdue := make([]maintenanceUrgency, 12)
	for i := range overdue {
		overdue[i] = maintenanceUrgency{
			Item: data.MaintenanceItem{
				ID:   uint(i + 1),
				Name: fmt.Sprintf("Long maintenance task %d", i+1),
			},
			DaysFromNow:   -(i + 1),
			ApplianceName: "Big Appliance",
		}
	}
	projects := make([]data.Project, 8)
	for i := range projects {
		projects[i] = data.Project{
			Title:  fmt.Sprintf("Project with a fairly long name %d", i+1),
			Status: data.ProjectStatusInProgress,
		}
		projects[i].ID = uint(100 + i)
	}
	m.dashboard = dashboardData{
		Overdue:           overdue,
		ActiveProjects:    projects,
		ServiceSpendCents: 50000,
		ProjectSpendCents: 100000,
	}
	m.buildDashNav()

	overlay := m.buildDashboardOverlay()
	overlayH := lipgloss.Height(overlay)
	maxH := m.effectiveHeight() - 4
	assert.LessOrEqual(t, overlayH, maxH,
		"overlay height %d exceeds max %d", overlayH, maxH)
}

func TestDashboardOverlayDimsSurroundingContent(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.showDashboard = true
	m.dashboard = dashboardData{}
	m.dashNav = nil

	view := m.buildView()
	// Every line of the composited view that contains background content
	// (the tab underline, table headers, etc.) should carry the ANSI faint
	// attribute (\033[2m). Verify no line contains the tab underline
	// character without being wrapped in faint.
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "━") {
			assert.Contains(t, line, "\033[2m",
				"tab underline should be dimmed in overlay")
		}
	}
}

func TestDashboardStatusBarShowsNormal(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.showDashboard = true
	status := m.statusView()
	// With overlay active, main tab keybindings should be hidden.
	assert.NotContains(t, status, "NAV")
}

func TestBuildDashNav(t *testing.T) {
	m := newTestModel()
	now := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	overdueDue := time.Date(2026, 1, 25, 0, 0, 0, 0, time.UTC)

	m.dashboard = dashboardData{
		Overdue: []maintenanceUrgency{{
			Item:        data.MaintenanceItem{ID: 10, Name: "Filter"},
			NextDue:     overdueDue,
			DaysFromNow: daysUntil(now, overdueDue),
		}},
		ActiveProjects: []data.Project{{ID: 20, Title: "Deck"}},
		ExpiringWarranties: []warrantyStatus{{
			Appliance:   data.Appliance{ID: 30, Name: "Fridge"},
			DaysFromNow: 45,
		}},
	}
	m.buildDashNav()

	require.Len(t, m.dashNav, 3)
	assert.Equal(t, tabMaintenance, m.dashNav[0].Tab)
	assert.Equal(t, uint(10), m.dashNav[0].ID)
	assert.Equal(t, tabProjects, m.dashNav[1].Tab)
	assert.Equal(t, uint(20), m.dashNav[1].ID)
	assert.Equal(t, tabAppliances, m.dashNav[2].Tab)
	assert.Equal(t, uint(30), m.dashNav[2].ID)
}

func TestRenderMiniTable(t *testing.T) {
	rows := []dashRow{
		{Cells: []dashCell{
			{Text: "Short", Style: lipgloss.NewStyle()},
			{Text: "123", Style: lipgloss.NewStyle(), Align: alignRight},
		}},
		{Cells: []dashCell{
			{Text: "Longer name", Style: lipgloss.NewStyle()},
			{Text: "7", Style: lipgloss.NewStyle(), Align: alignRight},
		}},
	}
	lines := renderMiniTable(rows, 3, 0, -1, lipgloss.NewStyle())
	require.Len(t, lines, 2)
	// Both lines should have the same visible width due to column alignment.
	assert.Equal(t, lipgloss.Width(lines[0]), lipgloss.Width(lines[1]))
}

func TestRenderMiniTableUnicode(t *testing.T) {
	plain := lipgloss.NewStyle()

	t.Run("accented Latin", func(t *testing.T) {
		rows := []dashRow{
			{Cells: []dashCell{
				{Text: "Garc\u00eda Plumbing", Style: plain},
				{Text: "3 days overdue", Style: plain, Align: alignRight},
			}},
			{Cells: []dashCell{
				{Text: "HVAC Filter", Style: plain},
				{Text: "in 14 days", Style: plain, Align: alignRight},
			}},
		}
		lines := renderMiniTable(rows, 3, 0, -1, plain)
		require.Len(t, lines, 2)
		assert.Equal(t, lipgloss.Width(lines[0]), lipgloss.Width(lines[1]),
			"rows with accented characters should align")
	})

	t.Run("CJK wide characters", func(t *testing.T) {
		// CJK characters occupy 2 terminal cells each.
		rows := []dashRow{
			{Cells: []dashCell{
				{Text: "\u6771\u829d\u88fd\u54c1", Style: plain}, // 東芝製品 = 8 cells
				{Text: "$500", Style: plain, Align: alignRight},
			}},
			{Cells: []dashCell{
				{Text: "Short", Style: plain}, // 5 cells
				{Text: "$1,000", Style: plain, Align: alignRight},
			}},
		}
		lines := renderMiniTable(rows, 3, 0, -1, plain)
		require.Len(t, lines, 2)
		assert.Equal(t, lipgloss.Width(lines[0]), lipgloss.Width(lines[1]),
			"rows with CJK characters should align")
	})

	t.Run("emoji", func(t *testing.T) {
		rows := []dashRow{
			{Cells: []dashCell{
				{Text: "Check \u2705", Style: plain},
				{Text: "done", Style: plain},
			}},
			{Cells: []dashCell{
				{Text: "Long task name", Style: plain},
				{Text: "pending", Style: plain},
			}},
		}
		lines := renderMiniTable(rows, 3, 0, -1, plain)
		require.Len(t, lines, 2)
		assert.Equal(t, lipgloss.Width(lines[0]), lipgloss.Width(lines[1]),
			"rows with emoji should align")
	})
}

func TestRenderMiniTableTruncatesOnNarrowWidth(t *testing.T) {
	plain := lipgloss.NewStyle()
	rows := []dashRow{
		{Cells: []dashCell{
			{Text: "Very long maintenance item name here", Style: plain},
			{Text: "3 days overdue", Style: plain, Align: alignRight},
		}},
		{Cells: []dashCell{
			{Text: "Another long name for testing", Style: plain},
			{Text: "in 14 days", Style: plain, Align: alignRight},
		}},
	}

	// Without width cap, rows are as wide as content demands.
	uncapped := renderMiniTable(rows, 3, 0, -1, plain)
	require.Len(t, uncapped, 2)
	uncappedWidth := lipgloss.Width(uncapped[0])

	// With a tight width cap, rows should be truncated.
	capped := renderMiniTable(rows, 3, 40, -1, plain)
	require.Len(t, capped, 2)
	for i, line := range capped {
		w := lipgloss.Width(line)
		assert.LessOrEqual(t, w, 40,
			"line %d width %d exceeds maxWidth 40", i, w)
	}

	// Capped should be narrower than uncapped.
	assert.Less(t, lipgloss.Width(capped[0]), uncappedWidth,
		"capped lines should be narrower")

	// Truncated first column should contain an ellipsis.
	assert.Contains(t, capped[0], "\u2026", "truncated line should contain ellipsis")
}

func TestTruncateToWidth(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		maxW  int
		check func(t *testing.T, result string)
	}{
		{
			name: "fits within width",
			text: "hello",
			maxW: 10,
			check: func(t *testing.T, result string) {
				assert.Equal(t, "hello", result)
			},
		},
		{
			name: "truncated with ellipsis",
			text: "very long text here",
			maxW: 10,
			check: func(t *testing.T, result string) {
				assert.LessOrEqual(t, lipgloss.Width(result), 10)
				assert.Contains(t, result, "\u2026")
			},
		},
		{
			name: "CJK truncation",
			text: "\u6771\u829d\u88fd\u54c1\u682a\u5f0f\u4f1a\u793e", // 東芝製品株式会社
			maxW: 8,
			check: func(t *testing.T, result string) {
				assert.LessOrEqual(t, lipgloss.Width(result), 8)
				assert.Contains(t, result, "\u2026")
			},
		},
		{
			name: "width 1 returns ellipsis",
			text: "hello",
			maxW: 1,
			check: func(t *testing.T, result string) {
				assert.Equal(t, "\u2026", result)
			},
		},
		{
			name: "width 0 returns empty",
			text: "hello",
			maxW: 0,
			check: func(t *testing.T, result string) {
				assert.Empty(t, result)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateToWidth(tt.text, tt.maxW)
			tt.check(t, result)
		})
	}
}

func TestDistributeProportional(t *testing.T) {
	t.Run("everything fits", func(t *testing.T) {
		got := distributeProportional([]int{3, 5}, 20)
		assert.Equal(t, []int{3, 5}, got)
	})

	t.Run("proportional trimming", func(t *testing.T) {
		got := distributeProportional([]int{10, 2}, 6)
		assert.Equal(t, 6, got[0]+got[1])
		assert.Greater(t, got[0], got[1])
	})

	t.Run("minimum 1 per bucket", func(t *testing.T) {
		got := distributeProportional([]int{10, 10, 10}, 3)
		for i, g := range got {
			assert.GreaterOrEqualf(t, g, 1, "bucket %d got %d", i, g)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		assert.Nil(t, distributeProportional(nil, 10))
		assert.Nil(t, distributeProportional([]int{}, 10))
	})

	t.Run("single bucket", func(t *testing.T) {
		got := distributeProportional([]int{8}, 3)
		assert.Equal(t, []int{3}, got)
	})

	t.Run("never exceeds counts", func(t *testing.T) {
		got := distributeProportional([]int{2, 3}, 100)
		assert.Equal(t, []int{2, 3}, got)
	})
}

func TestDistributeDashRows(t *testing.T) {
	t.Run("everything fits", func(t *testing.T) {
		sections := []dashSection{
			{rows: make([]dashRow, 3)},
			{rows: make([]dashRow, 5)},
		}
		got := distributeDashRows(sections, 20)
		assert.Equal(t, 3, got[0])
		assert.Equal(t, 5, got[1])
	})

	t.Run("proportional trimming", func(t *testing.T) {
		sections := []dashSection{
			{rows: make([]dashRow, 10)},
			{rows: make([]dashRow, 2)},
		}
		got := distributeDashRows(sections, 6)
		assert.Equal(t, 6, got[0]+got[1])
		assert.Greater(t, got[0], got[1])
	})

	t.Run("minimum 1 per section", func(t *testing.T) {
		sections := []dashSection{
			{rows: make([]dashRow, 10)},
			{rows: make([]dashRow, 10)},
			{rows: make([]dashRow, 10)},
		}
		got := distributeDashRows(sections, 3)
		for i, g := range got {
			assert.GreaterOrEqualf(t, g, 1, "section %d got %d rows", i, g)
		}
	})
}

func TestDashboardViewSkipsEmptySections(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	// Only spending data, no navigable sections.
	m.dashboard = dashboardData{
		ServiceSpendCents: 10000,
	}
	m.buildDashNav()

	view := m.dashboardView(50, 120)
	assert.Contains(t, view, "Spending")
	assert.NotContains(t, view, "Overdue")
	assert.NotContains(t, view, "Active Projects")
}

func TestDashboardViewTrimRows(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40

	// Create enough data to exceed a small budget.
	overdue := make([]maintenanceUrgency, 8)
	for i := range overdue {
		overdue[i] = maintenanceUrgency{
			Item:        data.MaintenanceItem{ID: uint(i + 1), Name: fmt.Sprintf("Task %d", i+1)},
			DaysFromNow: -(i + 1),
		}
	}
	projects := make([]data.Project, 5)
	for i := range projects {
		projects[i] = data.Project{Title: fmt.Sprintf("Proj %d", i+1), Status: "underway"}
		projects[i].ID = uint(100 + i)
	}
	m.dashboard = dashboardData{
		Overdue:        overdue,
		ActiveProjects: projects,
	}
	m.buildDashNav()

	// With a generous budget, all rows appear.
	bigView := m.dashboardView(100, 120)
	for i := 1; i <= 8; i++ {
		assert.Containsf(
			t,
			bigView,
			fmt.Sprintf("Task %d", i),
			"expected Task %d in big-budget view",
			i,
		)
	}

	// With a tiny budget, rows are trimmed but at least 1 per section.
	m.buildDashNav()
	m.dashCursor = 0
	smallView := m.dashboardView(6, 120)
	assert.Contains(t, smallView, "Task")
	assert.Contains(t, smallView, "Proj")
	assert.Less(t, len(m.dashNav), 13, "expected nav trimmed")
}

func TestDashboardNavRebuiltFromTrimmedView(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40

	overdue := make([]maintenanceUrgency, 10)
	for i := range overdue {
		overdue[i] = maintenanceUrgency{
			Item:        data.MaintenanceItem{ID: uint(i + 1), Name: fmt.Sprintf("Item %d", i+1)},
			DaysFromNow: -(i + 1),
		}
	}
	m.dashboard = dashboardData{Overdue: overdue}
	m.buildDashNav()
	m.dashCursor = 9 // at the end

	// Render with a tiny budget: forces trimming.
	_ = m.dashboardView(5, 120)

	assert.LessOrEqual(t, len(m.dashNav), 5)
	assert.Less(t, m.dashCursor, len(m.dashNav))
}

func TestOverlayContentWidth(t *testing.T) {
	tests := []struct {
		name      string
		termWidth int
		want      int
	}{
		{"wide terminal caps at 72", 200, 72},
		{"normal terminal", 100, 72},
		{"narrow terminal", 60, 48},
		{"very narrow caps at 30", 30, 30},
		{"minimum clamp", 20, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.width = tt.termWidth
			assert.Equal(t, tt.want, m.overlayContentWidth())
		})
	}
}
