// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
	"time"

	"github.com/cpcloud/micasa/internal/data"
)

func TestLoadDashboardAtClassifiesOverdueAndUpcoming(t *testing.T) {
	m := newTestModelWithStore(t)

	// Create an appliance and a maintenance item with a known last-serviced
	// date and interval so we can predict overdue vs upcoming.
	app := data.Appliance{Name: "Furnace"}
	if err := m.store.CreateAppliance(app); err != nil {
		t.Fatalf("CreateAppliance: %v", err)
	}
	apps, err := m.store.ListAppliances(false)
	if err != nil {
		t.Fatalf("ListAppliances: %v", err)
	}
	appID := apps[0].ID

	cats, err := m.store.MaintenanceCategories()
	if err != nil {
		t.Fatal(err)
	}

	// Item serviced 4 months ago, interval 3 months -> 1 month overdue.
	fourMonthsAgo := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	overdue := data.MaintenanceItem{
		Name:           "Replace Filter",
		CategoryID:     cats[0].ID,
		ApplianceID:    &appID,
		LastServicedAt: &fourMonthsAgo,
		IntervalMonths: 3,
	}
	if err := m.store.CreateMaintenance(overdue); err != nil {
		t.Fatalf("CreateMaintenance overdue: %v", err)
	}

	// Item serviced 1 month ago, interval 3 months -> due in ~2 months (upcoming).
	oneMonthAgo := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)
	upcoming := data.MaintenanceItem{
		Name:           "Clean Coils",
		CategoryID:     cats[0].ID,
		LastServicedAt: &oneMonthAgo,
		IntervalMonths: 3,
	}
	if err := m.store.CreateMaintenance(upcoming); err != nil {
		t.Fatalf("CreateMaintenance upcoming: %v", err)
	}

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if err := m.loadDashboardAt(now); err != nil {
		t.Fatalf("loadDashboardAt: %v", err)
	}

	if len(m.dashboard.Overdue) != 1 {
		t.Fatalf("expected 1 overdue, got %d", len(m.dashboard.Overdue))
	}
	if m.dashboard.Overdue[0].Item.Name != "Replace Filter" {
		t.Errorf("overdue item = %q, want Replace Filter", m.dashboard.Overdue[0].Item.Name)
	}
	if m.dashboard.Overdue[0].ApplianceName != "Furnace" {
		t.Errorf("appliance name = %q, want Furnace", m.dashboard.Overdue[0].ApplianceName)
	}
	if m.dashboard.Overdue[0].DaysFromNow >= 0 {
		t.Errorf("overdue DaysFromNow = %d, expected negative", m.dashboard.Overdue[0].DaysFromNow)
	}

	// "Clean Coils" is due in ~2 months — not within 30 days, so not upcoming.
	if len(m.dashboard.Upcoming) != 0 {
		t.Fatalf("expected 0 upcoming (due in ~2 months), got %d", len(m.dashboard.Upcoming))
	}
}

func TestLoadDashboardAtUpcomingWithin30Days(t *testing.T) {
	m := newTestModelWithStore(t)
	cats, _ := m.store.MaintenanceCategories()

	// Serviced 2.5 months ago with 3-month interval -> due in ~2 weeks.
	lastSrv := time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC)
	item := data.MaintenanceItem{
		Name:           "Check Sump Pump",
		CategoryID:     cats[0].ID,
		LastServicedAt: &lastSrv,
		IntervalMonths: 3,
	}
	if err := m.store.CreateMaintenance(item); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if err := m.loadDashboardAt(now); err != nil {
		t.Fatal(err)
	}

	if len(m.dashboard.Upcoming) != 1 {
		t.Fatalf("expected 1 upcoming, got %d", len(m.dashboard.Upcoming))
	}
	if m.dashboard.Upcoming[0].DaysFromNow < 0 || m.dashboard.Upcoming[0].DaysFromNow > 30 {
		t.Errorf("unexpected DaysFromNow = %d", m.dashboard.Upcoming[0].DaysFromNow)
	}
}

func TestLoadDashboardAtActiveProjects(t *testing.T) {
	m := newTestModelWithStore(t)
	types, _ := m.store.ProjectTypes()

	if err := m.store.CreateProject(data.Project{
		Title:         "Kitchen Remodel",
		ProjectTypeID: types[0].ID,
		Status:        data.ProjectStatusInProgress,
	}); err != nil {
		t.Fatal(err)
	}
	if err := m.store.CreateProject(data.Project{
		Title:         "Done Project",
		ProjectTypeID: types[0].ID,
		Status:        data.ProjectStatusCompleted,
	}); err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	if err := m.loadDashboardAt(now); err != nil {
		t.Fatal(err)
	}

	// Only in-progress projects should appear.
	if len(m.dashboard.ActiveProjects) != 1 {
		t.Fatalf("expected 1 active project, got %d", len(m.dashboard.ActiveProjects))
	}
	if m.dashboard.ActiveProjects[0].Title != "Kitchen Remodel" {
		t.Errorf("active project = %q", m.dashboard.ActiveProjects[0].Title)
	}
}

func TestLoadDashboardAtExpiringWarranties(t *testing.T) {
	m := newTestModelWithStore(t)

	expiry := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	if err := m.store.CreateAppliance(data.Appliance{
		Name:           "Dishwasher",
		WarrantyExpiry: &expiry,
	}); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if err := m.loadDashboardAt(now); err != nil {
		t.Fatal(err)
	}

	if len(m.dashboard.ExpiringWarranties) != 1 {
		t.Fatalf("expected 1 expiring warranty, got %d", len(m.dashboard.ExpiringWarranties))
	}
	if m.dashboard.ExpiringWarranties[0].Appliance.Name != "Dishwasher" {
		t.Errorf("wrong appliance: %q", m.dashboard.ExpiringWarranties[0].Appliance.Name)
	}
}

func TestLoadDashboardAtInsuranceRenewal(t *testing.T) {
	m := newTestModelWithStore(t)

	renewal := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	m.house.InsuranceCarrier = "State Farm"
	m.house.InsuranceRenewal = &renewal
	m.hasHouse = true

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if err := m.loadDashboardAt(now); err != nil {
		t.Fatal(err)
	}

	if m.dashboard.InsuranceRenewal == nil {
		t.Fatal("expected insurance renewal data")
	}
	if m.dashboard.InsuranceRenewal.Carrier != "State Farm" {
		t.Errorf("carrier = %q", m.dashboard.InsuranceRenewal.Carrier)
	}
	if m.dashboard.InsuranceRenewal.DaysFromNow != 28 {
		t.Errorf("days = %d, want 28", m.dashboard.InsuranceRenewal.DaysFromNow)
	}
}

func TestLoadDashboardAtInsuranceRenewalOutOfRange(t *testing.T) {
	m := newTestModelWithStore(t)

	// Renewal 6 months away — outside the -30..+90 window.
	renewal := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	m.house.InsuranceCarrier = "Allstate"
	m.house.InsuranceRenewal = &renewal
	m.hasHouse = true

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if err := m.loadDashboardAt(now); err != nil {
		t.Fatal(err)
	}

	if m.dashboard.InsuranceRenewal != nil {
		t.Error("expected no insurance renewal when 6 months away")
	}
}

func TestLoadDashboardAtSpending(t *testing.T) {
	m := newTestModelWithStore(t)
	cats, _ := m.store.MaintenanceCategories()

	// Create a maintenance item + service log with a cost.
	item := data.MaintenanceItem{
		Name:       "Oil Change",
		CategoryID: cats[0].ID,
	}
	if err := m.store.CreateMaintenance(item); err != nil {
		t.Fatal(err)
	}
	items, _ := m.store.ListMaintenance(false)
	cost := int64(5000)
	entry := data.ServiceLogEntry{
		MaintenanceItemID: items[0].ID,
		ServicedAt:        time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		CostCents:         &cost,
	}
	if err := m.store.CreateServiceLog(entry, data.Vendor{}); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if err := m.loadDashboardAt(now); err != nil {
		t.Fatal(err)
	}

	if m.dashboard.ServiceSpendCents != 5000 {
		t.Errorf("service spend = %d, want 5000", m.dashboard.ServiceSpendCents)
	}
}

func TestLoadDashboardAtBuildsNav(t *testing.T) {
	m := newTestModelWithStore(t)
	cats, _ := m.store.MaintenanceCategories()

	// Create an overdue item so nav has at least one entry.
	fourMonthsAgo := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	if err := m.store.CreateMaintenance(data.MaintenanceItem{
		Name:           "Check Gutters",
		CategoryID:     cats[0].ID,
		LastServicedAt: &fourMonthsAgo,
		IntervalMonths: 3,
	}); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if err := m.loadDashboardAt(now); err != nil {
		t.Fatal(err)
	}

	if len(m.dashNav) == 0 {
		t.Error("expected dashNav to be populated")
	}
	if m.dashNav[0].Tab != tabMaintenance {
		t.Errorf("first nav entry tab = %d, want tabMaintenance", m.dashNav[0].Tab)
	}
}
