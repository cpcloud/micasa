// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"testing"
	"time"
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
	if err != nil {
		t.Fatalf("ListMaintenanceWithSchedule: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Name != "With Interval" {
		t.Fatalf("expected 'With Interval', got %q", items[0].Name)
	}
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
	if err != nil {
		t.Fatalf("ListActiveProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 active projects, got %d", len(projects))
	}
	names := map[string]bool{}
	for _, p := range projects {
		names[p.Title] = true
	}
	if !names["A"] || !names["B"] {
		t.Fatalf("expected projects A and B, got %v", names)
	}
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
	if err != nil {
		t.Fatalf("ListExpiringWarranties: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 expiring, got %d", len(apps))
	}
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
	if err != nil {
		t.Fatalf("ListRecentServiceLogs: %v", err)
	}
	if len(entries) != 5 {
		t.Fatalf("expected 5, got %d", len(entries))
	}
	// Most recent should be first.
	if entries[0].ServicedAt.Month() != 10 {
		t.Fatalf("expected most recent (Oct), got %v", entries[0].ServicedAt)
	}
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
	if err != nil {
		t.Fatalf("YTDServiceSpendCents: %v", err)
	}
	if spend != 5000 {
		t.Fatalf("expected 5000, got %d", spend)
	}

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

	projSpend, err := store.YTDProjectSpendCents()
	if err != nil {
		t.Fatalf("YTDProjectSpendCents: %v", err)
	}
	if projSpend != 30000 {
		t.Fatalf("expected 30000, got %d", projSpend)
	}
}
