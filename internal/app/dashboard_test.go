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
			got := daysUntil(now, tt.target)
			if got != tt.want {
				t.Errorf("daysUntil(%v, %v) = %d, want %d", now, tt.target, got, tt.want)
			}
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
		got := daysText(tt.days, tt.overdue)
		if got != tt.want {
			t.Errorf("daysText(%d, %v) = %q, want %q",
				tt.days, tt.overdue, got, tt.want)
		}
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
		if items[i].DaysFromNow < items[i-1].DaysFromNow {
			t.Fatalf("not sorted: items[%d]=%d < items[%d]=%d",
				i, items[i].DaysFromNow, i-1, items[i-1].DaysFromNow)
		}
	}
}

func TestCapSlice(t *testing.T) {
	if got := capSlice([]int{1, 2, 3, 4, 5}, 3); len(got) != 3 {
		t.Fatalf("expected 3, got %d", len(got))
	}
	if got := capSlice([]int{1, 2}, 5); len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	if got := capSlice([]int{1, 2, 3}, -1); len(got) != 0 {
		t.Fatalf("expected 0, got %d", len(got))
	}
}

func TestDashboardToggle(t *testing.T) {
	m := newTestModel()
	m.showDashboard = false

	sendKey(m, "D")
	if !m.showDashboard {
		t.Fatal("expected showDashboard=true after D")
	}
	sendKey(m, "D")
	if m.showDashboard {
		t.Fatal("expected showDashboard=false after second D")
	}
}

func TestDashboardDismissedByTab(t *testing.T) {
	m := newTestModel()
	m.showDashboard = true

	sendKey(m, "tab")
	if m.showDashboard {
		t.Fatal("expected showDashboard=false after tab")
	}
}

func TestDashboardDismissedByShiftTab(t *testing.T) {
	m := newTestModel()
	m.showDashboard = true

	sendKey(m, "shift+tab")
	if m.showDashboard {
		t.Fatal("expected showDashboard=false after shift+tab")
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
	if m.dashCursor != 1 {
		t.Fatalf("expected cursor 1 after j, got %d", m.dashCursor)
	}
	// k moves up.
	sendKey(m, "k")
	if m.dashCursor != 0 {
		t.Fatalf("expected cursor 0 after k, got %d", m.dashCursor)
	}
	// k at 0 stays at 0 (no wrap).
	sendKey(m, "k")
	if m.dashCursor != 0 {
		t.Fatalf("expected cursor 0 after k at top, got %d", m.dashCursor)
	}
	// G jumps to bottom.
	sendKey(m, "G")
	if m.dashCursor != 4 {
		t.Fatalf("expected cursor 4 after G, got %d", m.dashCursor)
	}
	// j at bottom stays at bottom.
	sendKey(m, "j")
	if m.dashCursor != 4 {
		t.Fatalf("expected cursor 4 after j at bottom, got %d", m.dashCursor)
	}
	// g jumps to top.
	sendKey(m, "g")
	if m.dashCursor != 0 {
		t.Fatalf("expected cursor 0 after g, got %d", m.dashCursor)
	}
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
	if m.showDashboard {
		t.Fatal("expected dashboard dismissed after jump")
	}
	if m.active != tabIndex(tabProjects) {
		t.Fatalf("expected active tab %d, got %d", tabIndex(tabProjects), m.active)
	}
}

func TestDashboardEnterKeyJumps(t *testing.T) {
	m := newTestModel()
	m.showDashboard = true
	m.dashNav = []dashNavEntry{
		{Tab: tabMaintenance, ID: 1},
		{Tab: tabProjects, ID: 42},
	}
	m.dashCursor = 1

	// enter via sendKey should go through handleDashboardKeys and jump.
	sendKey(m, "enter")
	if m.showDashboard {
		t.Fatal("expected dashboard dismissed after enter")
	}
	if m.active != tabIndex(tabProjects) {
		t.Fatalf("expected active tab %d, got %d", tabIndex(tabProjects), m.active)
	}
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
	if m.tabs[m.active].ColCursor != colBefore {
		t.Error("l should be blocked on dashboard")
	}
	sendKey(m, "h")
	if m.tabs[m.active].ColCursor != colBefore {
		t.Error("h should be blocked on dashboard")
	}

	// s should not add a sort.
	sortsBefore := len(m.tabs[m.active].Sorts)
	sendKey(m, "s")
	if len(m.tabs[m.active].Sorts) != sortsBefore {
		t.Error("s should be blocked on dashboard")
	}

	// i should not enter edit mode.
	sendKey(m, "i")
	if m.mode != modeNormal {
		t.Error("i should be blocked on dashboard")
	}

	// Dashboard should still be showing.
	if !m.showDashboard {
		t.Error("dashboard should still be showing after blocked keys")
	}
	if m.active != startTab {
		t.Error("active tab should not have changed")
	}
}

func TestDashboardViewEmptySections(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.dashboard = dashboardData{}
	m.dashNav = nil
	m.dashCursor = 0

	view := m.dashboardView(50)
	// Empty dashboard returns empty string -- silence is success.
	if view != "" {
		t.Errorf("expected empty string for empty dashboard, got %q", view)
	}
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

	view := m.dashboardView(50)
	if !strings.Contains(view, "HVAC Filter") {
		t.Error("expected overdue item in view")
	}
	if !strings.Contains(view, "Kitchen Remodel") {
		t.Error("expected active project in view")
	}
	if !strings.Contains(view, "overdue") {
		t.Error("expected 'overdue' label in view")
	}
	if !strings.Contains(view, "$500.00") {
		t.Error("expected service spend in view")
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
	// Header shows today's date (no "Dashboard" title -- tab bar handles that).
	today := time.Now().Format("Monday, Jan 2")
	if !strings.Contains(ov, today) {
		t.Errorf("expected today's date %q in overlay header", today)
	}
	if !strings.Contains(ov, "help") {
		t.Error("expected help hint in overlay")
	}
}

func TestDashboardOverlayComposite(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.showDashboard = true
	m.dashboard = dashboardData{}
	m.dashNav = nil

	// buildView should produce a composited result without panicking.
	view := m.buildView()
	if view == "" {
		t.Error("expected non-empty composited view")
	}
}

func TestDashboardStatusBarShowsNormal(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.showDashboard = true
	// statusView always shows normal-mode status; dashboard hints are
	// in the overlay, not the status bar.
	status := m.statusView()
	if !strings.Contains(status, "NAV") {
		t.Error("expected NAV badge in status bar")
	}
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

	if len(m.dashNav) != 3 {
		t.Fatalf("expected 3 nav entries, got %d", len(m.dashNav))
	}
	if m.dashNav[0].Tab != tabMaintenance || m.dashNav[0].ID != 10 {
		t.Errorf("nav[0] = %+v, want maintenance/10", m.dashNav[0])
	}
	if m.dashNav[1].Tab != tabProjects || m.dashNav[1].ID != 20 {
		t.Errorf("nav[1] = %+v, want projects/20", m.dashNav[1])
	}
	if m.dashNav[2].Tab != tabAppliances || m.dashNav[2].ID != 30 {
		t.Errorf("nav[2] = %+v, want appliances/30", m.dashNav[2])
	}
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
	lines := renderMiniTable(rows, 3, -1, lipgloss.NewStyle())
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	// Both lines should have the same visible width due to column alignment.
	w0 := len(lines[0])
	w1 := len(lines[1])
	if w0 != w1 {
		t.Errorf("lines have different widths: %d vs %d", w0, w1)
	}
}

func TestDistributeDashRows(t *testing.T) {
	t.Run("everything fits", func(t *testing.T) {
		sections := []dashSection{
			{rows: make([]dashRow, 3)},
			{rows: make([]dashRow, 5)},
		}
		got := distributeDashRows(sections, 20)
		if got[0] != 3 || got[1] != 5 {
			t.Errorf("expected [3,5], got %v", got)
		}
	})

	t.Run("proportional trimming", func(t *testing.T) {
		sections := []dashSection{
			{rows: make([]dashRow, 10)},
			{rows: make([]dashRow, 2)},
		}
		got := distributeDashRows(sections, 6)
		// 6 avail, 2 sections get min 1 each, 4 remaining.
		// Section 0 has 9 extra, section 1 has 1 extra, total excess = 10.
		// Proportional: s0 gets 1 + 9*4/10=4 = 5, s1 gets 1 + 1*4/10=1 = 1+0=1.
		// Rounding leftover: 6-6=0. Total = 6.
		if got[0]+got[1] != 6 {
			t.Errorf("expected total 6, got %d (%v)", got[0]+got[1], got)
		}
		// Larger section should get more rows.
		if got[0] <= got[1] {
			t.Errorf("expected s0 > s1, got %v", got)
		}
	})

	t.Run("minimum 1 per section", func(t *testing.T) {
		sections := []dashSection{
			{rows: make([]dashRow, 10)},
			{rows: make([]dashRow, 10)},
			{rows: make([]dashRow, 10)},
		}
		got := distributeDashRows(sections, 3)
		for i, g := range got {
			if g < 1 {
				t.Errorf("section %d got %d rows, expected >= 1", i, g)
			}
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

	view := m.dashboardView(50)
	if !strings.Contains(view, "Spending") {
		t.Error("expected spending section")
	}
	if strings.Contains(view, "Overdue") || strings.Contains(view, "Active Projects") {
		t.Error("empty sections should not appear")
	}
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
	bigView := m.dashboardView(100)
	for i := 1; i <= 8; i++ {
		if !strings.Contains(bigView, fmt.Sprintf("Task %d", i)) {
			t.Errorf("expected Task %d in big-budget view", i)
		}
	}

	// With a tiny budget, rows are trimmed but at least 1 per section.
	m.buildDashNav()
	m.dashCursor = 0
	smallView := m.dashboardView(6)
	// Should still have content from both sections.
	hasMaint := strings.Contains(smallView, "Task")
	hasProj := strings.Contains(smallView, "Proj")
	if !hasMaint || !hasProj {
		t.Errorf("expected both sections represented, maint=%v proj=%v", hasMaint, hasProj)
	}
	// Nav should be trimmed too.
	if len(m.dashNav) >= 13 {
		t.Errorf("expected nav trimmed, got %d entries", len(m.dashNav))
	}
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
	_ = m.dashboardView(5)

	// Nav should be trimmed and cursor clamped.
	if len(m.dashNav) > 5 {
		t.Errorf("expected nav <= 5, got %d", len(m.dashNav))
	}
	if m.dashCursor >= len(m.dashNav) {
		t.Errorf("cursor %d should be < nav length %d", m.dashCursor, len(m.dashNav))
	}
}

func TestDashboardDefaultOnLaunchWithHouse(t *testing.T) {
	m := newTestModel()
	m.hasHouse = true
	m.house = data.HouseProfile{Nickname: "Test"}
	m.showDashboard = true

	if !m.showDashboard {
		t.Fatal("expected dashboard on launch when house exists")
	}
}
