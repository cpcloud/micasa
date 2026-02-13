// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
	"time"

	"github.com/cpcloud/micasa/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.Len(t, rows, 2)

	// First row: overdue item.
	assert.Equal(t, "Replace Filter", rows[0].Cells[0].Text)
	assert.Equal(t, "Furnace", rows[0].Cells[1].Text)
	require.NotNil(t, rows[0].Target)
	assert.Equal(t, tabMaintenance, rows[0].Target.Tab)

	// Second row: upcoming item, no appliance.
	assert.Equal(t, "Check Pump", rows[1].Cells[0].Text)
	assert.Empty(t, rows[1].Cells[1].Text)
}

func TestDashMaintRowsEmpty(t *testing.T) {
	m := newTestModel()
	m.dashboard = dashboardData{}
	rows := m.dashMaintRows()
	assert.Nil(t, rows)
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
	require.Len(t, rows, 1)
	assert.Equal(t, "2025-12-25", rows[0].Cells[3].Text)
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
	require.Len(t, rows, 1)
	assert.NotEmpty(t, rows[0].Cells[2].Text, "expected budget text")
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
	require.Len(t, rows, 1)
	assert.NotEmpty(t, rows[0].Cells[2].Text, "expected budget-only text")
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
	require.Len(t, rows, 1)
	assert.Empty(t, rows[0].Cells[2].Text)
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
	require.Len(t, rows, 2)
	assert.Equal(t, "Fridge warranty", rows[0].Cells[0].Text)
	assert.Equal(t, "Oven warranty", rows[1].Cells[0].Text)
	// Both should have nav targets.
	require.NotNil(t, rows[0].Target)
	assert.Equal(t, tabAppliances, rows[0].Target.Tab)
}

func TestDashExpiringRowsEmpty(t *testing.T) {
	m := newTestModel()
	m.dashboard = dashboardData{}
	rows := m.dashExpiringRows()
	assert.Nil(t, rows)
}

// ---------------------------------------------------------------------------
// overhead
// ---------------------------------------------------------------------------

func TestOverhead(t *testing.T) {
	tests := []struct {
		name string
		s    dashSection
		want int
	}{
		{
			"single section",
			dashSection{title: "Projects", rows: make([]dashRow, 3)},
			1,
		},
		{
			"two non-empty sub-sections",
			dashSection{
				title: "Maintenance", subTitles: []string{"Overdue", "Upcoming"},
				subCounts: []int{3, 2},
			},
			3, // 2 sub-headers + 1 blank separator
		},
		{
			"one empty sub-section",
			dashSection{
				title: "Maintenance", subTitles: []string{"Overdue", "Upcoming"},
				subCounts: []int{3, 0},
			},
			1,
		},
		{
			"all sub-sections empty",
			dashSection{
				title: "Maintenance", subTitles: []string{"Overdue", "Upcoming"},
				subCounts: []int{0, 0},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.s.overhead())
		})
	}
}
