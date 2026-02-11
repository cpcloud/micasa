// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
	"time"

	"github.com/cpcloud/micasa/internal/data"
)

func TestDashMaintRowsOverdueAndUpcoming(t *testing.T) {
	m := newTestModel()
	m.styles = DefaultStyles()

	lastSrv := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	m.dashboard = dashboardData{
		Overdue: []maintenanceUrgency{{
			Item: data.MaintenanceItem{
				ID:             1,
				Name:           "Replace Filter",
				LastServicedAt: &lastSrv,
			},
			ApplianceName: "Furnace",
			DaysFromNow:   -14,
		}},
		Upcoming: []maintenanceUrgency{{
			Item:        data.MaintenanceItem{ID: 2, Name: "Check Pump"},
			DaysFromNow: 10,
		}},
	}

	rows := m.dashMaintRows()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// First row: overdue item.
	if rows[0].Cells[0].Text != "Replace Filter" {
		t.Errorf("row[0] name = %q", rows[0].Cells[0].Text)
	}
	if rows[0].Cells[1].Text != "Furnace" {
		t.Errorf("row[0] appliance = %q", rows[0].Cells[1].Text)
	}
	if rows[0].Target == nil || rows[0].Target.Tab != tabMaintenance {
		t.Error("expected target pointing to maintenance tab")
	}

	// Second row: upcoming item, no appliance.
	if rows[1].Cells[0].Text != "Check Pump" {
		t.Errorf("row[1] name = %q", rows[1].Cells[0].Text)
	}
	if rows[1].Cells[1].Text != "" {
		t.Errorf("row[1] appliance = %q, want empty", rows[1].Cells[1].Text)
	}
}

func TestDashMaintRowsEmpty(t *testing.T) {
	m := newTestModel()
	m.dashboard = dashboardData{}
	rows := m.dashMaintRows()
	if rows != nil {
		t.Errorf("expected nil rows for empty maintenance, got %d", len(rows))
	}
}

func TestDashMaintRowsLastServicedAt(t *testing.T) {
	m := newTestModel()
	m.styles = DefaultStyles()

	lastSrv := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
	m.dashboard = dashboardData{
		Overdue: []maintenanceUrgency{{
			Item: data.MaintenanceItem{
				ID:             1,
				Name:           "Task",
				LastServicedAt: &lastSrv,
			},
			DaysFromNow: -5,
		}},
	}

	rows := m.dashMaintRows()
	if len(rows) != 1 {
		t.Fatal("expected 1 row")
	}
	// Last column should show the formatted date.
	if rows[0].Cells[3].Text != "2025-12-25" {
		t.Errorf("last serviced = %q, want 2025-12-25", rows[0].Cells[3].Text)
	}
}

func TestDashProjectRowsBudgetFormatting(t *testing.T) {
	m := newTestModel()
	m.styles = DefaultStyles()

	budget := int64(100000) // $1,000.00
	actual := int64(120000) // $1,200.00 — over budget
	m.dashboard = dashboardData{
		ActiveProjects: []data.Project{
			{
				Title:       "Over Budget Project",
				Status:      data.ProjectStatusInProgress,
				BudgetCents: &budget,
				ActualCents: &actual,
			},
		},
	}

	rows := m.dashProjectRows()
	if len(rows) != 1 {
		t.Fatal("expected 1 row")
	}

	budgetCell := rows[0].Cells[2]
	// Should show "actual / budget" format.
	if budgetCell.Text == "" {
		t.Error("expected budget text")
	}
}

func TestDashProjectRowsBudgetOnly(t *testing.T) {
	m := newTestModel()
	m.styles = DefaultStyles()

	budget := int64(50000)
	m.dashboard = dashboardData{
		ActiveProjects: []data.Project{{
			Title:       "Budget Only",
			Status:      data.ProjectStatusPlanned,
			BudgetCents: &budget,
		}},
	}

	rows := m.dashProjectRows()
	if len(rows) != 1 {
		t.Fatal("expected 1 row")
	}
	// No actual, so just budget.
	if rows[0].Cells[2].Text == "" {
		t.Error("expected budget-only text")
	}
}

func TestDashProjectRowsNoBudget(t *testing.T) {
	m := newTestModel()
	m.styles = DefaultStyles()

	m.dashboard = dashboardData{
		ActiveProjects: []data.Project{{
			Title:  "No Budget",
			Status: data.ProjectStatusIdeating,
		}},
	}

	rows := m.dashProjectRows()
	if len(rows) != 1 {
		t.Fatal("expected 1 row")
	}
	if rows[0].Cells[2].Text != "" {
		t.Errorf("expected empty budget, got %q", rows[0].Cells[2].Text)
	}
}

func TestDashExpiringRowsOverdueAndUpcoming(t *testing.T) {
	m := newTestModel()
	m.styles = DefaultStyles()

	expiredDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	upcomingDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	m.dashboard = dashboardData{
		ExpiringWarranties: []warrantyStatus{
			{
				Appliance:   data.Appliance{ID: 1, Name: "Fridge", WarrantyExpiry: &expiredDate},
				DaysFromNow: -20,
			},
			{
				Appliance:   data.Appliance{ID: 2, Name: "Oven", WarrantyExpiry: &upcomingDate},
				DaysFromNow: 55,
			},
		},
	}

	rows := m.dashExpiringRows()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].Cells[0].Text != "Fridge warranty" {
		t.Errorf("row[0] = %q", rows[0].Cells[0].Text)
	}
	if rows[1].Cells[0].Text != "Oven warranty" {
		t.Errorf("row[1] = %q", rows[1].Cells[0].Text)
	}
	// Both should have nav targets.
	if rows[0].Target == nil || rows[0].Target.Tab != tabAppliances {
		t.Error("expected appliance nav target on row 0")
	}
}

func TestDashExpiringRowsEmpty(t *testing.T) {
	m := newTestModel()
	m.dashboard = dashboardData{}
	rows := m.dashExpiringRows()
	if rows != nil {
		t.Error("expected nil rows for no expiring warranties")
	}
}

// ---------------------------------------------------------------------------
// overhead
// ---------------------------------------------------------------------------

func TestOverheadSingleSection(t *testing.T) {
	s := dashSection{title: "Projects", rows: make([]dashRow, 3)}
	if got := s.overhead(); got != 1 {
		t.Errorf("overhead = %d, want 1", got)
	}
}

func TestOverheadSubSections(t *testing.T) {
	s := dashSection{
		title:     "Maintenance",
		subTitles: []string{"Overdue", "Upcoming"},
		subCounts: []int{3, 2},
	}
	// 2 sub-headers + 1 blank separator = 3
	if got := s.overhead(); got != 3 {
		t.Errorf("overhead = %d, want 3", got)
	}
}

func TestOverheadSubSectionsOneEmpty(t *testing.T) {
	s := dashSection{
		title:     "Maintenance",
		subTitles: []string{"Overdue", "Upcoming"},
		subCounts: []int{3, 0},
	}
	// Only 1 non-empty sub-section -> overhead = 1.
	if got := s.overhead(); got != 1 {
		t.Errorf("overhead = %d, want 1", got)
	}
}

func TestOverheadAllSubSectionsEmpty(t *testing.T) {
	s := dashSection{
		title:     "Maintenance",
		subTitles: []string{"Overdue", "Upcoming"},
		subCounts: []int{0, 0},
	}
	if got := s.overhead(); got != 1 {
		t.Errorf("overhead = %d, want 1", got)
	}
}
