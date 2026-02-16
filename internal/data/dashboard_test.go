// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListMaintenanceWithSchedule(t *testing.T) {
	store := newTestStore(t)
	cat := MaintenanceCategory{Name: "TestCat"}
	store.db.Create(&cat)

	ptrTime := func(y, m, d int) *time.Time {
		t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
		return &t
	}
	// Item with interval > 0 should appear.
	store.db.Create(&MaintenanceItem{
		Name: "With Interval", CategoryID: cat.ID,
		IntervalMonths: 3, LastServicedAt: ptrTime(2025, 6, 1),
	})
	// Item with interval = 0 should NOT appear.
	store.db.Create(&MaintenanceItem{
		Name: "No Interval", CategoryID: cat.ID, IntervalMonths: 0,
	})

	items, err := store.ListMaintenanceWithSchedule()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "With Interval", items[0].Name)
}

func TestListActiveProjects(t *testing.T) {
	store := newTestStore(t)
	var pt ProjectType
	store.db.First(&pt)
	store.db.Create(&Project{Title: "A", ProjectTypeID: pt.ID, Status: ProjectStatusInProgress})
	store.db.Create(&Project{Title: "B", ProjectTypeID: pt.ID, Status: ProjectStatusDelayed})
	store.db.Create(&Project{Title: "C", ProjectTypeID: pt.ID, Status: ProjectStatusCompleted})
	store.db.Create(&Project{Title: "D", ProjectTypeID: pt.ID, Status: ProjectStatusIdeating})

	projects, err := store.ListActiveProjects()
	require.NoError(t, err)
	require.Len(t, projects, 2)
	names := map[string]bool{}
	for _, p := range projects {
		names[p.Title] = true
	}
	assert.True(t, names["A"] && names["B"], "expected projects A and B, got %v", names)
}

func TestListExpiringWarranties(t *testing.T) {
	store := newTestStore(t)
	now := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	ptrTime := func(y, m, d int) *time.Time {
		t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
		return &t
	}
	// Expiring in 30 days -- should appear.
	store.db.Create(&Appliance{Name: "Soon", WarrantyExpiry: ptrTime(2026, 3, 10)})
	// Expired 10 days ago -- should appear (within lookBack).
	store.db.Create(&Appliance{Name: "Recent", WarrantyExpiry: ptrTime(2026, 1, 29)})
	// Expired 60 days ago -- should NOT appear.
	store.db.Create(&Appliance{Name: "Old", WarrantyExpiry: ptrTime(2025, 12, 1)})
	// Expiring in 120 days -- should NOT appear.
	store.db.Create(&Appliance{Name: "Far", WarrantyExpiry: ptrTime(2026, 6, 8)})
	// No warranty -- should NOT appear.
	store.db.Create(&Appliance{Name: "None"})

	apps, err := store.ListExpiringWarranties(now, 30*24*time.Hour, 90*24*time.Hour)
	require.NoError(t, err)
	require.Len(t, apps, 2)
}

func TestListRecentServiceLogs(t *testing.T) {
	store := newTestStore(t)
	cat := MaintenanceCategory{Name: "SLCat"}
	store.db.Create(&cat)
	item := MaintenanceItem{Name: "SL Item", CategoryID: cat.ID, IntervalMonths: 6}
	store.db.Create(&item)

	for i := 0; i < 10; i++ {
		store.db.Create(&ServiceLogEntry{
			MaintenanceItemID: item.ID,
			ServicedAt:        time.Date(2025, 1+time.Month(i), 1, 0, 0, 0, 0, time.UTC),
		})
	}

	entries, err := store.ListRecentServiceLogs(5)
	require.NoError(t, err)
	require.Len(t, entries, 5)
	// Most recent should be first.
	assert.Equal(t, time.October, entries[0].ServicedAt.Month())
}

func TestYTDSpending(t *testing.T) {
	store := newTestStore(t)
	ptr := func(v int64) *int64 { return &v }

	cat := MaintenanceCategory{Name: "SpendCat"}
	store.db.Create(&cat)
	item := MaintenanceItem{Name: "Spend Item", CategoryID: cat.ID, IntervalMonths: 6}
	store.db.Create(&item)

	// This year.
	store.db.Create(&ServiceLogEntry{
		MaintenanceItemID: item.ID,
		ServicedAt:        time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		CostCents:         ptr(5000),
	})
	// Last year -- should not count.
	store.db.Create(&ServiceLogEntry{
		MaintenanceItemID: item.ID,
		ServicedAt:        time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		CostCents:         ptr(9999),
	})

	yearStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	spend, err := store.YTDServiceSpendCents(yearStart)
	require.NoError(t, err)
	assert.Equal(t, int64(5000), spend)

	// Projects.
	var pt ProjectType
	store.db.First(&pt)
	store.db.Create(&Project{
		Title: "P1", ProjectTypeID: pt.ID, Status: ProjectStatusCompleted,
		ActualCents: ptr(20000),
	})
	store.db.Create(&Project{
		Title: "P2", ProjectTypeID: pt.ID, Status: ProjectStatusInProgress,
		ActualCents: ptr(10000),
	})
	// Project updated last year -- should not count.
	oldProj := Project{
		Title: "P3", ProjectTypeID: pt.ID, Status: ProjectStatusCompleted,
		ActualCents: ptr(7777),
	}
	store.db.Create(&oldProj)
	store.db.Exec(
		"UPDATE projects SET updated_at = ? WHERE title = ?",
		time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), "P3",
	)

	projSpend, err := store.YTDProjectSpendCents(yearStart)
	require.NoError(t, err)
	assert.Equal(t, int64(30000), projSpend)
}
