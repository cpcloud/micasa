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

func TestLoadDashboardAtClassifiesOverdueAndUpcoming(t *testing.T) {
	m := newTestModelWithStore(t)

	app := data.Appliance{Name: "Furnace"}
	require.NoError(t, m.store.CreateAppliance(app))
	apps, err := m.store.ListAppliances(false)
	require.NoError(t, err)
	appID := apps[0].ID

	cats, err := m.store.MaintenanceCategories()
	require.NoError(t, err)

	// Item serviced 4 months ago, interval 3 months -> 1 month overdue.
	fourMonthsAgo := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	overdue := data.MaintenanceItem{
		Name:           "Replace Filter",
		CategoryID:     cats[0].ID,
		ApplianceID:    &appID,
		LastServicedAt: &fourMonthsAgo,
		IntervalMonths: 3,
	}
	require.NoError(t, m.store.CreateMaintenance(overdue))

	// Item serviced 1 month ago, interval 3 months -> due in ~2 months (upcoming).
	oneMonthAgo := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)
	upcoming := data.MaintenanceItem{
		Name:           "Clean Coils",
		CategoryID:     cats[0].ID,
		LastServicedAt: &oneMonthAgo,
		IntervalMonths: 3,
	}
	require.NoError(t, m.store.CreateMaintenance(upcoming))

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, m.loadDashboardAt(now))

	require.Len(t, m.dashboard.Overdue, 1)
	assert.Equal(t, "Replace Filter", m.dashboard.Overdue[0].Item.Name)
	assert.Equal(t, "Furnace", m.dashboard.Overdue[0].ApplianceName)
	assert.Less(t, m.dashboard.Overdue[0].DaysFromNow, 0)

	// "Clean Coils" is due in ~2 months — not within 30 days, so not upcoming.
	assert.Empty(t, m.dashboard.Upcoming)
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
	require.NoError(t, m.store.CreateMaintenance(item))

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, m.loadDashboardAt(now))

	require.Len(t, m.dashboard.Upcoming, 1)
	assert.GreaterOrEqual(t, m.dashboard.Upcoming[0].DaysFromNow, 0)
	assert.LessOrEqual(t, m.dashboard.Upcoming[0].DaysFromNow, 30)
}

func TestLoadDashboardAtActiveProjects(t *testing.T) {
	m := newTestModelWithStore(t)
	types, _ := m.store.ProjectTypes()

	require.NoError(t, m.store.CreateProject(data.Project{
		Title:         "Kitchen Remodel",
		ProjectTypeID: types[0].ID,
		Status:        data.ProjectStatusInProgress,
	}))
	require.NoError(t, m.store.CreateProject(data.Project{
		Title:         "Done Project",
		ProjectTypeID: types[0].ID,
		Status:        data.ProjectStatusCompleted,
	}))

	now := time.Now()
	require.NoError(t, m.loadDashboardAt(now))

	// Only in-progress projects should appear.
	require.Len(t, m.dashboard.ActiveProjects, 1)
	assert.Equal(t, "Kitchen Remodel", m.dashboard.ActiveProjects[0].Title)
}

func TestLoadDashboardAtExpiringWarranties(t *testing.T) {
	m := newTestModelWithStore(t)

	expiry := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	require.NoError(t, m.store.CreateAppliance(data.Appliance{
		Name:           "Dishwasher",
		WarrantyExpiry: &expiry,
	}))

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, m.loadDashboardAt(now))

	require.Len(t, m.dashboard.ExpiringWarranties, 1)
	assert.Equal(t, "Dishwasher", m.dashboard.ExpiringWarranties[0].Appliance.Name)
}

func TestLoadDashboardAtInsuranceRenewal(t *testing.T) {
	m := newTestModelWithStore(t)

	renewal := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	m.house.InsuranceCarrier = "State Farm"
	m.house.InsuranceRenewal = &renewal
	m.hasHouse = true

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, m.loadDashboardAt(now))

	require.NotNil(t, m.dashboard.InsuranceRenewal)
	assert.Equal(t, "State Farm", m.dashboard.InsuranceRenewal.Carrier)
	assert.Equal(t, 28, m.dashboard.InsuranceRenewal.DaysFromNow)
}

func TestLoadDashboardAtInsuranceRenewalOutOfRange(t *testing.T) {
	m := newTestModelWithStore(t)

	// Renewal 6 months away — outside the -30..+90 window.
	renewal := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	m.house.InsuranceCarrier = "Allstate"
	m.house.InsuranceRenewal = &renewal
	m.hasHouse = true

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, m.loadDashboardAt(now))

	assert.Nil(t, m.dashboard.InsuranceRenewal)
}

func TestLoadDashboardAtSpending(t *testing.T) {
	m := newTestModelWithStore(t)
	cats, _ := m.store.MaintenanceCategories()

	// Create a maintenance item + service log with a cost.
	item := data.MaintenanceItem{
		Name:       "Oil Change",
		CategoryID: cats[0].ID,
	}
	require.NoError(t, m.store.CreateMaintenance(item))
	items, _ := m.store.ListMaintenance(false)
	cost := int64(5000)
	entry := data.ServiceLogEntry{
		MaintenanceItemID: items[0].ID,
		ServicedAt:        time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		CostCents:         &cost,
	}
	require.NoError(t, m.store.CreateServiceLog(entry, data.Vendor{}))

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, m.loadDashboardAt(now))

	assert.Equal(t, int64(5000), m.dashboard.ServiceSpendCents)
}

func TestLoadDashboardAtBuildsNav(t *testing.T) {
	m := newTestModelWithStore(t)
	cats, _ := m.store.MaintenanceCategories()

	// Create an overdue item so nav has at least one entry.
	fourMonthsAgo := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, m.store.CreateMaintenance(data.MaintenanceItem{
		Name:           "Check Gutters",
		CategoryID:     cats[0].ID,
		LastServicedAt: &fourMonthsAgo,
		IntervalMonths: 3,
	}))

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, m.loadDashboardAt(now))

	assert.NotEmpty(t, m.dashNav)
	assert.Equal(t, tabMaintenance, m.dashNav[0].Tab)
}

func TestLoadDashboardExcludesAppliancesWithoutWarranty(t *testing.T) {
	m := newTestModelWithStore(t)

	// One appliance with warranty in range, one without any warranty.
	expiry := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	require.NoError(t, m.store.CreateAppliance(data.Appliance{
		Name:           "Fridge",
		WarrantyExpiry: &expiry,
	}))
	require.NoError(t, m.store.CreateAppliance(data.Appliance{
		Name: "Toaster",
	}))

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, m.loadDashboardAt(now))

	require.Len(t, m.dashboard.ExpiringWarranties, 1)
	assert.Equal(t, "Fridge", m.dashboard.ExpiringWarranties[0].Appliance.Name)
}
