// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cpcloud/micasa/internal/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSeedDefaults(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	require.NoError(t, err)
	require.NotEmpty(t, types)
	categories, err := store.MaintenanceCategories()
	require.NoError(t, err)
	require.NotEmpty(t, categories)
}

func TestHouseProfileSingle(t *testing.T) {
	store := newTestStore(t)
	profile := HouseProfile{Nickname: "Primary Residence"}
	require.NoError(t, store.CreateHouseProfile(profile))
	_, err := store.HouseProfile()
	require.NoError(t, err)
	assert.Error(t, store.CreateHouseProfile(profile), "second profile should fail")
}

func TestUpdateHouseProfile(t *testing.T) {
	store := newTestStore(t)
	require.NoError(
		t,
		store.CreateHouseProfile(HouseProfile{Nickname: "Primary Residence", City: "Portland"}),
	)
	require.NoError(
		t,
		store.UpdateHouseProfile(HouseProfile{Nickname: "Primary Residence", City: "Seattle"}),
	)
	fetched, err := store.HouseProfile()
	require.NoError(t, err)
	assert.Equal(t, "Seattle", fetched.City)
}

func TestSoftDeleteRestoreProject(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	require.NoError(t, err)
	require.NoError(t, store.CreateProject(Project{
		Title: "Test Project", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))

	projects, err := store.ListProjects(false)
	require.NoError(t, err)
	require.Len(t, projects, 1)

	require.NoError(t, store.DeleteProject(projects[0].ID))

	projects, err = store.ListProjects(false)
	require.NoError(t, err)
	assert.Empty(t, projects)

	projects, err = store.ListProjects(true)
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.True(t, projects[0].DeletedAt.Valid)

	require.NoError(t, store.RestoreProject(projects[0].ID))
	projects, err = store.ListProjects(false)
	require.NoError(t, err)
	assert.Len(t, projects, 1)
}

func TestLastDeletionRecord(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	require.NoError(t, err)
	require.NoError(t, store.CreateProject(Project{
		Title: "Test Project", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, err := store.ListProjects(false)
	require.NoError(t, err)
	require.Len(t, projects, 1)

	require.NoError(t, store.DeleteProject(projects[0].ID))
	record, err := store.LastDeletion(DeletionEntityProject)
	require.NoError(t, err)
	assert.Equal(t, projects[0].ID, record.TargetID)

	require.NoError(t, store.RestoreProject(record.TargetID))
	_, err = store.LastDeletion(DeletionEntityProject)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUpdateProject(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	require.NoError(t, err)
	require.NoError(t, store.CreateProject(Project{
		Title: "Original Title", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, err := store.ListProjects(false)
	require.NoError(t, err)
	require.Len(t, projects, 1)
	id := projects[0].ID

	fetched, err := store.GetProject(id)
	require.NoError(t, err)
	assert.Equal(t, "Original Title", fetched.Title)

	require.NoError(t, store.UpdateProject(Project{
		ID: id, Title: "Updated Title", ProjectTypeID: types[0].ID,
		Status: ProjectStatusInProgress,
	}))

	fetched, err = store.GetProject(id)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", fetched.Title)
	assert.Equal(t, ProjectStatusInProgress, fetched.Status)
}

func TestUpdateQuote(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	require.NoError(t, err)
	require.NoError(t, store.CreateProject(Project{
		Title: "Test Project", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, err := store.ListProjects(false)
	require.NoError(t, err)
	require.Len(t, projects, 1)

	require.NoError(t, store.CreateQuote(
		Quote{ProjectID: projects[0].ID, TotalCents: 100000},
		Vendor{Name: "Acme Corp"},
	))
	quotes, err := store.ListQuotes(false)
	require.NoError(t, err)
	require.Len(t, quotes, 1)
	id := quotes[0].ID

	require.NoError(t, store.UpdateQuote(
		Quote{ID: id, ProjectID: projects[0].ID, TotalCents: 200000},
		Vendor{Name: "Acme Corp", ContactName: "John Doe"},
	))

	fetched, err := store.GetQuote(id)
	require.NoError(t, err)
	assert.Equal(t, int64(200000), fetched.TotalCents)
	assert.Equal(t, "John Doe", fetched.Vendor.ContactName)
}

func TestUpdateMaintenance(t *testing.T) {
	store := newTestStore(t)
	categories, err := store.MaintenanceCategories()
	require.NoError(t, err)
	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Filter Change", CategoryID: categories[0].ID,
	}))
	items, err := store.ListMaintenance(false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	id := items[0].ID

	fetched, err := store.GetMaintenance(id)
	require.NoError(t, err)
	assert.Equal(t, "Filter Change", fetched.Name)

	require.NoError(t, store.UpdateMaintenance(MaintenanceItem{
		ID: id, Name: "HVAC Filter Change", CategoryID: categories[0].ID, IntervalMonths: 3,
	}))

	fetched, err = store.GetMaintenance(id)
	require.NoError(t, err)
	assert.Equal(t, "HVAC Filter Change", fetched.Name)
	assert.Equal(t, 3, fetched.IntervalMonths)
}

func TestServiceLogCRUD(t *testing.T) {
	store := newTestStore(t)
	categories, err := store.MaintenanceCategories()
	require.NoError(t, err)
	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Test Maintenance", CategoryID: categories[0].ID,
	}))
	items, err := store.ListMaintenance(false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	maintID := items[0].ID

	// Create a service log entry (self-performed, no vendor).
	require.NoError(t, store.CreateServiceLog(ServiceLogEntry{
		MaintenanceItemID: maintID,
		ServicedAt:        time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Notes:             "did it myself",
	}, Vendor{}))

	entries, err := store.ListServiceLog(maintID, false)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Nil(t, entries[0].VendorID)
	assert.Equal(t, "did it myself", entries[0].Notes)

	// Create a vendor-performed entry.
	require.NoError(t, store.CreateServiceLog(ServiceLogEntry{
		MaintenanceItemID: maintID,
		ServicedAt:        time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		CostCents:         func() *int64 { v := int64(15000); return &v }(),
		Notes:             "vendor did it",
	}, Vendor{Name: "Test Plumber", Phone: "555-555-0001"}))

	entries, err = store.ListServiceLog(maintID, false)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	// Most recent first.
	require.NotNil(t, entries[0].VendorID)
	assert.Equal(t, "Test Plumber", entries[0].Vendor.Name)

	// Update: change vendor entry to self-performed.
	updated := entries[0]
	updated.Notes = "actually did it myself"
	require.NoError(t, store.UpdateServiceLog(updated, Vendor{}))

	fetched, err := store.GetServiceLog(updated.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched.VendorID)
	assert.Equal(t, "actually did it myself", fetched.Notes)

	// Delete and restore.
	require.NoError(t, store.DeleteServiceLog(fetched.ID))
	entries, err = store.ListServiceLog(maintID, false)
	require.NoError(t, err)
	assert.Len(t, entries, 1)

	entries, err = store.ListServiceLog(maintID, true)
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	require.NoError(t, store.RestoreServiceLog(fetched.ID))
	entries, err = store.ListServiceLog(maintID, false)
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	// CountServiceLogs.
	counts, err := store.CountServiceLogs([]uint{maintID})
	require.NoError(t, err)
	assert.Equal(t, 2, counts[maintID])
}

func TestSoftDeletePersistsAcrossRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persist.db")

	// Session 1: create a project, then soft-delete it.
	store1, err := Open(path)
	require.NoError(t, err)
	require.NoError(t, store1.AutoMigrate())
	require.NoError(t, store1.SeedDefaults())
	types, _ := store1.ProjectTypes()
	require.NoError(t, store1.CreateProject(Project{
		Title: "Persist Test", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store1.ListProjects(false)
	var projectID uint
	for _, p := range projects {
		if p.Title == "Persist Test" {
			projectID = p.ID
			break
		}
	}
	require.NotZero(t, projectID)
	require.NoError(t, store1.DeleteProject(projectID))
	_ = store1.Close()

	// Session 2: reopen and verify the project is still soft-deleted.
	store2, err := Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store2.Close() })
	require.NoError(t, store2.AutoMigrate())

	projects2, err := store2.ListProjects(false)
	require.NoError(t, err)
	for _, p := range projects2 {
		assert.NotEqual(
			t,
			projectID,
			p.ID,
			"soft-deleted project should not appear in normal listing after reopen",
		)
	}

	projectsAll, err := store2.ListProjects(true)
	require.NoError(t, err)
	found := false
	for _, p := range projectsAll {
		if p.ID == projectID {
			found = true
			break
		}
	}
	assert.True(t, found, "soft-deleted project should appear in unscoped listing after reopen")

	require.NoError(t, store2.RestoreProject(projectID))
	projects3, err := store2.ListProjects(false)
	require.NoError(t, err)
	found = false
	for _, p := range projects3 {
		if p.ID == projectID {
			found = true
			break
		}
	}
	assert.True(t, found, "restored project should appear in normal listing")
}

func TestVendorCRUD(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.CreateVendor(Vendor{
		Name: "Test Vendor", ContactName: "Alice",
		Email: "alice@example.com", Phone: "555-0001",
	}))

	vendors, err := store.ListVendors(false)
	require.NoError(t, err)
	require.Len(t, vendors, 1)
	got := vendors[0]
	assert.Equal(t, "Test Vendor", got.Name)
	assert.Equal(t, "Alice", got.ContactName)

	fetched, err := store.GetVendor(got.ID)
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", fetched.Email)

	fetched.Phone = "555-9999"
	fetched.Website = "https://example.com"
	require.NoError(t, store.UpdateVendor(fetched))
	updated, _ := store.GetVendor(fetched.ID)
	assert.Equal(t, "555-9999", updated.Phone)
	assert.Equal(t, "https://example.com", updated.Website)
}

func TestCountQuotesByVendor(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.CreateVendor(Vendor{Name: "Quote Vendor"}))
	vendors, _ := store.ListVendors(false)
	vendorID := vendors[0].ID

	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "Test", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projectID := projects[0].ID

	for i := 0; i < 2; i++ {
		require.NoError(t, store.CreateQuote(
			Quote{ProjectID: projectID, TotalCents: 100000},
			Vendor{Name: "Quote Vendor"},
		))
	}

	counts, err := store.CountQuotesByVendor([]uint{vendorID})
	require.NoError(t, err)
	assert.Equal(t, 2, counts[vendorID])

	empty, err := store.CountQuotesByVendor(nil)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestCountServiceLogsByVendor(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.CreateVendor(Vendor{Name: "Job Vendor"}))
	vendors, _ := store.ListVendors(false)
	vendorID := vendors[0].ID

	cats, _ := store.MaintenanceCategories()
	require.NoError(
		t,
		store.CreateMaintenance(MaintenanceItem{Name: "Filter", CategoryID: cats[0].ID}),
	)
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	require.NoError(t, store.CreateServiceLog(
		ServiceLogEntry{MaintenanceItemID: maintID, ServicedAt: time.Now()},
		Vendor{Name: "Job Vendor"},
	))

	counts, err := store.CountServiceLogsByVendor([]uint{vendorID})
	require.NoError(t, err)
	assert.Equal(t, 1, counts[vendorID])
}

func TestDeleteProjectBlockedByQuotes(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "Blocked Project", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projID := projects[0].ID

	require.NoError(
		t,
		store.CreateQuote(Quote{ProjectID: projID, TotalCents: 1000}, Vendor{Name: "V1"}),
	)

	require.ErrorContains(t, store.DeleteProject(projID), "active quote")

	quotes, _ := store.ListQuotes(false)
	require.NoError(t, store.DeleteQuote(quotes[0].ID))
	require.NoError(t, store.DeleteProject(projID))
}

func TestRestoreQuoteBlockedByDeletedProject(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "Doomed Project", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projID := projects[0].ID

	require.NoError(
		t,
		store.CreateQuote(Quote{ProjectID: projID, TotalCents: 500}, Vendor{Name: "V2"}),
	)
	quotes, _ := store.ListQuotes(false)
	quoteID := quotes[0].ID

	require.NoError(t, store.DeleteQuote(quoteID))
	require.NoError(t, store.DeleteProject(projID))

	require.ErrorContains(t, store.RestoreQuote(quoteID), "project is deleted")

	require.NoError(t, store.RestoreProject(projID))
	require.NoError(t, store.RestoreQuote(quoteID))
}

func TestRestoreServiceLogBlockedByDeletedMaintenance(t *testing.T) {
	store := newTestStore(t)
	cats, _ := store.MaintenanceCategories()
	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Doomed Maint", CategoryID: cats[0].ID, IntervalMonths: 6,
	}))
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	require.NoError(t, store.CreateServiceLog(
		ServiceLogEntry{MaintenanceItemID: maintID, ServicedAt: time.Now()},
		Vendor{Name: "SL2"},
	))
	logs, _ := store.ListServiceLog(maintID, false)
	logID := logs[0].ID

	require.NoError(t, store.DeleteServiceLog(logID))
	require.NoError(t, store.DeleteMaintenance(maintID))

	require.ErrorContains(t, store.RestoreServiceLog(logID), "maintenance item is deleted")

	require.NoError(t, store.RestoreMaintenance(maintID))
	require.NoError(t, store.RestoreServiceLog(logID))
}

func TestDeleteMaintenanceBlockedByServiceLogs(t *testing.T) {
	store := newTestStore(t)
	cats, _ := store.MaintenanceCategories()
	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Blocked Maint", CategoryID: cats[0].ID, IntervalMonths: 3,
	}))
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	require.NoError(t, store.CreateServiceLog(
		ServiceLogEntry{MaintenanceItemID: maintID, ServicedAt: time.Now()},
		Vendor{Name: "SL Vendor"},
	))

	require.ErrorContains(t, store.DeleteMaintenance(maintID), "service log")

	logs, _ := store.ListServiceLog(maintID, false)
	require.NoError(t, store.DeleteServiceLog(logs[0].ID))
	require.NoError(t, store.DeleteMaintenance(maintID))
}

func TestPartialQuoteDeletionStillBlocksProjectDelete(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "Multi-Quote", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projID := projects[0].ID

	for _, name := range []string{"Vendor A", "Vendor B"} {
		require.NoError(
			t,
			store.CreateQuote(Quote{ProjectID: projID, TotalCents: 1000}, Vendor{Name: name}),
		)
	}
	quotes, _ := store.ListQuotes(false)
	require.Len(t, quotes, 2)

	require.NoError(t, store.DeleteQuote(quotes[0].ID))
	require.ErrorContains(t, store.DeleteProject(projID), "1 active quote")

	require.NoError(t, store.DeleteQuote(quotes[1].ID))
	require.NoError(t, store.DeleteProject(projID))
}

func TestRestoreMaintenanceBlockedByDeletedAppliance(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.CreateAppliance(Appliance{Name: "Doomed Fridge"}))
	appliances, _ := store.ListAppliances(false)
	appID := appliances[0].ID

	cats, _ := store.MaintenanceCategories()
	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Coil Cleaning", CategoryID: cats[0].ID, IntervalMonths: 6, ApplianceID: &appID,
	}))
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	require.NoError(t, store.DeleteMaintenance(maintID))
	require.NoError(t, store.DeleteAppliance(appID))

	require.ErrorContains(t, store.RestoreMaintenance(maintID), "appliance is deleted")

	require.NoError(t, store.RestoreAppliance(appID))
	require.NoError(t, store.RestoreMaintenance(maintID))
}

func TestRestoreMaintenanceAllowedWithoutAppliance(t *testing.T) {
	store := newTestStore(t)
	cats, _ := store.MaintenanceCategories()
	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Gutter Cleaning", CategoryID: cats[0].ID, IntervalMonths: 6,
	}))
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	require.NoError(t, store.DeleteMaintenance(maintID))
	require.NoError(t, store.RestoreMaintenance(maintID))
}

func TestThreeLevelDeleteRestoreChain(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.CreateAppliance(Appliance{Name: "HVAC Unit"}))
	appliances, _ := store.ListAppliances(false)
	appID := appliances[0].ID

	cats, _ := store.MaintenanceCategories()
	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Filter Change", CategoryID: cats[0].ID, IntervalMonths: 3, ApplianceID: &appID,
	}))
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	require.NoError(t, store.CreateServiceLog(
		ServiceLogEntry{MaintenanceItemID: maintID, ServicedAt: time.Now()},
		Vendor{},
	))
	logs, _ := store.ListServiceLog(maintID, false)
	logID := logs[0].ID

	// --- Delete bottom-up ---
	assert.Error(t, store.DeleteMaintenance(maintID), "active service log should block")

	require.NoError(t, store.DeleteServiceLog(logID))
	require.NoError(t, store.DeleteMaintenance(maintID))
	require.NoError(t, store.DeleteAppliance(appID))

	// --- Attempt wrong-order restores ---
	require.ErrorContains(t, store.RestoreServiceLog(logID), "maintenance item is deleted")
	require.ErrorContains(t, store.RestoreMaintenance(maintID), "appliance is deleted")

	// --- Restore correct order ---
	require.NoError(t, store.RestoreAppliance(appID))
	require.NoError(t, store.RestoreMaintenance(maintID))
	require.NoError(t, store.RestoreServiceLog(logID))

	fetched, err := store.GetMaintenance(maintID)
	require.NoError(t, err)
	require.NotNil(t, fetched.ApplianceID)
	assert.Equal(t, appID, *fetched.ApplianceID)

	restoredLogs, err := store.ListServiceLog(maintID, false)
	require.NoError(t, err)
	assert.Len(t, restoredLogs, 1)
}

func TestDeleteApplianceAllowedWithMaintenance(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.CreateAppliance(Appliance{Name: "Deletable Fridge"}))
	appliances, _ := store.ListAppliances(false)
	appID := appliances[0].ID

	cats, _ := store.MaintenanceCategories()
	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Filter", CategoryID: cats[0].ID, IntervalMonths: 6, ApplianceID: &appID,
	}))

	require.NoError(t, store.DeleteAppliance(appID))
}

func TestGetAppliance(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.CreateAppliance(Appliance{Name: "Fridge"}))
	got, err := store.GetAppliance(1)
	require.NoError(t, err)
	assert.Equal(t, "Fridge", got.Name)
}

func TestGetApplianceNotFound(t *testing.T) {
	store := newTestStore(t)
	_, err := store.GetAppliance(9999)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUpdateAppliance(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.CreateAppliance(Appliance{Name: "Fridge"}))
	got, _ := store.GetAppliance(1)
	got.Brand = "Samsung"
	require.NoError(t, store.UpdateAppliance(got))
	updated, _ := store.GetAppliance(1)
	assert.Equal(t, "Samsung", updated.Brand)
}

func TestListMaintenanceByAppliance(t *testing.T) {
	store := newTestStore(t)
	categories, _ := store.MaintenanceCategories()
	catID := categories[0].ID

	require.NoError(t, store.CreateAppliance(Appliance{Name: "Fridge"}))
	appID := uint(1)
	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Clean coils", CategoryID: catID, ApplianceID: &appID,
	}))
	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Check smoke detectors", CategoryID: catID,
	}))

	items, err := store.ListMaintenanceByAppliance(appID, false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Clean coils", items[0].Name)
}

func TestCountMaintenanceByAppliance(t *testing.T) {
	store := newTestStore(t)
	categories, _ := store.MaintenanceCategories()
	catID := categories[0].ID

	require.NoError(t, store.CreateAppliance(Appliance{Name: "Fridge"}))
	appID := uint(1)
	for _, name := range []string{"Clean coils", "Replace filter"} {
		require.NoError(t, store.CreateMaintenance(MaintenanceItem{
			Name: name, CategoryID: catID, ApplianceID: &appID,
		}))
	}

	counts, err := store.CountMaintenanceByAppliance([]uint{appID})
	require.NoError(t, err)
	assert.Equal(t, 2, counts[appID])
}

func TestUpdateServiceLog(t *testing.T) {
	store := newTestStore(t)
	categories, _ := store.MaintenanceCategories()
	catID := categories[0].ID

	require.NoError(
		t,
		store.CreateMaintenance(MaintenanceItem{Name: "HVAC filter", CategoryID: catID}),
	)
	now := time.Now().Truncate(time.Second)
	require.NoError(t, store.CreateServiceLog(ServiceLogEntry{
		MaintenanceItemID: 1, ServicedAt: now, Notes: "initial",
	}, Vendor{}))

	created, _ := store.GetServiceLog(1)
	created.Notes = "updated"
	require.NoError(t, store.UpdateServiceLog(created, Vendor{Name: "HVAC Pros"}))

	updated, _ := store.GetServiceLog(1)
	assert.Equal(t, "updated", updated.Notes)
	assert.NotNil(t, updated.VendorID)
}

func TestUpdateServiceLogClearVendor(t *testing.T) {
	store := newTestStore(t)
	categories, _ := store.MaintenanceCategories()
	catID := categories[0].ID

	require.NoError(
		t,
		store.CreateMaintenance(MaintenanceItem{Name: "HVAC filter", CategoryID: catID}),
	)
	now := time.Now().Truncate(time.Second)
	require.NoError(t, store.CreateServiceLog(ServiceLogEntry{
		MaintenanceItemID: 1, ServicedAt: now,
	}, Vendor{Name: "HVAC Pros"}))

	created, _ := store.GetServiceLog(1)
	require.NoError(t, store.UpdateServiceLog(created, Vendor{}))
	updated, _ := store.GetServiceLog(1)
	assert.Nil(t, updated.VendorID)
}

func TestListMaintenanceByApplianceIncludeDeleted(t *testing.T) {
	store := newTestStore(t)
	categories, _ := store.MaintenanceCategories()
	catID := categories[0].ID

	require.NoError(t, store.CreateAppliance(Appliance{Name: "Fridge"}))
	appID := uint(1)
	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Clean coils", CategoryID: catID, ApplianceID: &appID,
	}))
	require.NoError(t, store.DeleteMaintenance(1))

	items, _ := store.ListMaintenanceByAppliance(appID, false)
	assert.Empty(t, items)

	items, _ = store.ListMaintenanceByAppliance(appID, true)
	assert.Len(t, items, 1)
}

func TestSoftDeleteRestoreVendor(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.CreateVendor(Vendor{Name: "Test Vendor"}))

	vendors, _ := store.ListVendors(false)
	require.Len(t, vendors, 1)
	id := vendors[0].ID

	require.NoError(t, store.DeleteVendor(id))
	vendors, _ = store.ListVendors(false)
	assert.Empty(t, vendors)

	vendors, _ = store.ListVendors(true)
	require.Len(t, vendors, 1)
	assert.True(t, vendors[0].DeletedAt.Valid)

	require.NoError(t, store.RestoreVendor(id))
	vendors, _ = store.ListVendors(false)
	assert.Len(t, vendors, 1)
}

func TestDeleteVendorBlockedByQuotes(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.CreateVendor(Vendor{Name: "Blocked Vendor"}))
	vendors, _ := store.ListVendors(false)
	vendorID := vendors[0].ID

	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "Test", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projID := projects[0].ID

	require.NoError(
		t,
		store.CreateQuote(
			Quote{ProjectID: projID, TotalCents: 1000},
			Vendor{Name: "Blocked Vendor"},
		),
	)

	require.ErrorContains(t, store.DeleteVendor(vendorID), "active quote")

	quotes, _ := store.ListQuotes(false)
	require.NoError(t, store.DeleteQuote(quotes[0].ID))
	require.NoError(t, store.DeleteVendor(vendorID))
}

func TestRestoreQuoteBlockedByDeletedVendor(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.CreateVendor(Vendor{Name: "Doomed Vendor"}))
	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "Test", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projID := projects[0].ID
	vendors, _ := store.ListVendors(false)
	vendorID := vendors[0].ID

	require.NoError(
		t,
		store.CreateQuote(Quote{ProjectID: projID, TotalCents: 500}, Vendor{Name: "Doomed Vendor"}),
	)
	quotes, _ := store.ListQuotes(false)
	quoteID := quotes[0].ID

	require.NoError(t, store.DeleteQuote(quoteID))
	require.NoError(t, store.DeleteVendor(vendorID))

	require.ErrorContains(t, store.RestoreQuote(quoteID), "vendor is deleted")

	require.NoError(t, store.RestoreVendor(vendorID))
	require.NoError(t, store.RestoreQuote(quoteID))
}

func TestRestoreServiceLogBlockedByDeletedVendor(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.CreateVendor(Vendor{Name: "Doomed SL Vendor"}))
	vendors, _ := store.ListVendors(false)
	vendorID := vendors[0].ID

	cats, _ := store.MaintenanceCategories()
	require.NoError(
		t,
		store.CreateMaintenance(MaintenanceItem{Name: "Test Maint", CategoryID: cats[0].ID}),
	)
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	require.NoError(t, store.CreateServiceLog(
		ServiceLogEntry{MaintenanceItemID: maintID, ServicedAt: time.Now()},
		Vendor{Name: "Doomed SL Vendor"},
	))
	logs, _ := store.ListServiceLog(maintID, false)
	logID := logs[0].ID

	require.NoError(t, store.DeleteServiceLog(logID))
	require.NoError(t, store.DeleteVendor(vendorID))

	require.ErrorContains(t, store.RestoreServiceLog(logID), "vendor is deleted")

	require.NoError(t, store.RestoreVendor(vendorID))
	require.NoError(t, store.RestoreServiceLog(logID))
}

func TestRestoreServiceLogAllowedWithoutVendor(t *testing.T) {
	store := newTestStore(t)
	cats, _ := store.MaintenanceCategories()
	require.NoError(
		t,
		store.CreateMaintenance(MaintenanceItem{Name: "Self Maint", CategoryID: cats[0].ID}),
	)
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	require.NoError(t, store.CreateServiceLog(
		ServiceLogEntry{MaintenanceItemID: maintID, ServicedAt: time.Now()},
		Vendor{},
	))
	logs, _ := store.ListServiceLog(maintID, false)
	logID := logs[0].ID

	require.NoError(t, store.DeleteServiceLog(logID))
	require.NoError(t, store.RestoreServiceLog(logID))
}

func TestRestoreProjectBlockedByDeletedPreferredVendor(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.CreateVendor(Vendor{Name: "Preferred Vendor"}))
	vendors, _ := store.ListVendors(false)
	vendorID := vendors[0].ID

	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "Vendor Project", ProjectTypeID: types[0].ID,
		Status: ProjectStatusPlanned, PreferredVendorID: &vendorID,
	}))
	projects, _ := store.ListProjects(false)
	projID := projects[0].ID

	require.NoError(t, store.DeleteProject(projID))
	require.NoError(t, store.DeleteVendor(vendorID))

	require.ErrorContains(t, store.RestoreProject(projID), "preferred vendor is deleted")

	require.NoError(t, store.RestoreVendor(vendorID))
	require.NoError(t, store.RestoreProject(projID))
}

func TestRestoreProjectAllowedWithoutPreferredVendor(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "No Vendor Project", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projID := projects[0].ID

	require.NoError(t, store.DeleteProject(projID))
	require.NoError(t, store.RestoreProject(projID))
}

func TestVendorQuoteProjectDeleteRestoreChain(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.CreateVendor(Vendor{Name: "Chain Vendor"}))
	vendors, _ := store.ListVendors(false)
	vendorID := vendors[0].ID

	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "Chain Project", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projID := projects[0].ID

	require.NoError(
		t,
		store.CreateQuote(Quote{ProjectID: projID, TotalCents: 1000}, Vendor{Name: "Chain Vendor"}),
	)
	quotes, _ := store.ListQuotes(false)
	quoteID := quotes[0].ID

	// --- Delete bottom-up ---
	assert.Error(t, store.DeleteVendor(vendorID), "active quote blocks vendor delete")
	assert.Error(t, store.DeleteProject(projID), "active quote blocks project delete")

	require.NoError(t, store.DeleteQuote(quoteID))
	require.NoError(t, store.DeleteProject(projID))
	require.NoError(t, store.DeleteVendor(vendorID))

	// --- Attempt wrong-order restores ---
	require.ErrorContains(t, store.RestoreQuote(quoteID), "project is deleted")

	require.NoError(t, store.RestoreProject(projID))
	require.ErrorContains(t, store.RestoreQuote(quoteID), "vendor is deleted")

	// --- Restore correct order ---
	require.NoError(t, store.RestoreVendor(vendorID))
	require.NoError(t, store.RestoreQuote(quoteID))

	vendors, _ = store.ListVendors(false)
	assert.Len(t, vendors, 1)
	quotes, _ = store.ListQuotes(false)
	assert.Len(t, quotes, 1)
}

func TestFindOrCreateVendorRestoresSoftDeleted(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.CreateVendor(Vendor{Name: "Revivable Vendor"}))
	vendors, _ := store.ListVendors(false)
	vendorID := vendors[0].ID

	require.NoError(t, store.DeleteVendor(vendorID))
	vendors, _ = store.ListVendors(false)
	assert.Empty(t, vendors)

	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "Test", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	require.NoError(t, store.CreateQuote(
		Quote{ProjectID: projects[0].ID, TotalCents: 500},
		Vendor{Name: "Revivable Vendor"},
	))

	vendors, _ = store.ListVendors(false)
	require.Len(t, vendors, 1)
	assert.Equal(t, vendorID, vendors[0].ID)
}

func TestVendorDeletionRecord(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.CreateVendor(Vendor{Name: "Record Vendor"}))
	vendors, _ := store.ListVendors(false)
	vendorID := vendors[0].ID

	require.NoError(t, store.DeleteVendor(vendorID))
	record, err := store.LastDeletion(DeletionEntityVendor)
	require.NoError(t, err)
	assert.Equal(t, vendorID, record.TargetID)

	require.NoError(t, store.RestoreVendor(vendorID))
	_, err = store.LastDeletion(DeletionEntityVendor)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUnicodeRoundTrip(t *testing.T) {
	store := newTestStore(t)

	tests := []struct {
		name     string
		nickname string
		city     string
	}{
		{"accented Latin", "Casa de Garc\u00eda", "San Jos\u00e9"},
		{"CJK characters", "\u6211\u7684\u5bb6", "\u6771\u4eac"},      // 我的家, 東京
		{"emoji", "Home \U0001f3e0", "City \u2605"},                   // 🏠, ★
		{"mixed scripts", "Haus M\u00fcller \u2014 \u6771\u4eac", ""}, // Haus Müller — 東京
		{"fraction and section", "\u00bd acre lot", "\u00a75 district"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Delete existing profile (if any) so we can create fresh.
			store.db.Where("1 = 1").Delete(&HouseProfile{})

			profile := HouseProfile{Nickname: tt.nickname, City: tt.city}
			require.NoError(t, store.CreateHouseProfile(profile))

			fetched, err := store.HouseProfile()
			require.NoError(t, err)
			assert.Equal(t, tt.nickname, fetched.Nickname, "nickname round-trip")
			assert.Equal(t, tt.city, fetched.City, "city round-trip")
		})
	}
}

func TestUnicodeRoundTripVendor(t *testing.T) {
	store := newTestStore(t)

	names := []string{
		"Garc\u00eda Plumbing",                 // García
		"M\u00fcller HVAC",                     // Müller
		"\u6771\u829d\u30b5\u30fc\u30d3\u30b9", // 東芝サービス
		"O'Brien & Sons",
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, store.CreateVendor(Vendor{Name: name}))
		})
	}

	vendors, err := store.ListVendors(false)
	require.NoError(t, err)
	vendorNames := make([]string, len(vendors))
	for i, v := range vendors {
		vendorNames[i] = v.Name
	}
	for _, name := range names {
		assert.Contains(t, vendorNames, name, "vendor %q should survive round-trip", name)
	}
}

func TestUnicodeRoundTripNotes(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	require.NoError(t, err)

	notes := "Technician Jos\u00e9 used \u00bd-inch fittings per \u00a75.2"
	require.NoError(t, store.CreateProject(Project{
		Title:         "Unicode notes test",
		ProjectTypeID: types[0].ID,
		Status:        ProjectStatusPlanned,
		Description:   notes,
	}))

	projects, err := store.ListProjects(false)
	require.NoError(t, err)
	require.NotEmpty(t, projects)
	assert.Equal(t, notes, projects[len(projects)-1].Description)
}

func TestDocumentCRUDAndMetadata(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "Doc Project", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projectID := projects[0].ID

	filePath := filepath.Join(t.TempDir(), "invoice.pdf")
	require.NoError(t, os.WriteFile(filePath, []byte("fake pdf content"), 0o600))

	require.NoError(t, store.CreateDocument(Document{
		Title:      "Quote PDF",
		EntityKind: DocumentEntityProject,
		EntityID:   &projectID,
		Notes:      "first draft",
	}, filePath))
	docs, err := store.ListDocuments(false)
	require.NoError(t, err)
	require.Len(t, docs, 1)

	doc := docs[0]
	assert.Equal(t, "Quote PDF", doc.Title)
	assert.Equal(t, "invoice.pdf", doc.FileName)
	assert.NotEmpty(t, doc.ChecksumSHA256)
	assert.NotEmpty(t, doc.MIMEType)
	assert.Equal(t, DocumentEntityProject, doc.EntityKind)
	require.NotNil(t, doc.EntityID)
	assert.Equal(t, projectID, *doc.EntityID)
	// Content is excluded from list queries.
	assert.Empty(t, doc.Content)

	require.NoError(t, store.DeleteDocument(doc.ID))
	docs, err = store.ListDocuments(false)
	require.NoError(t, err)
	assert.Empty(t, docs)

	require.NoError(t, store.RestoreDocument(doc.ID))
	docs, err = store.ListDocuments(false)
	require.NoError(t, err)
	require.Len(t, docs, 1)
}

func TestRestoreDocumentBlockedByDeletedTarget(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()
	require.NoError(t, store.CreateProject(Project{
		Title: "Doc Restore Project", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projectID := projects[0].ID

	filePath := filepath.Join(t.TempDir(), "note.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("note"), 0o600))

	require.NoError(t, store.CreateDocument(Document{
		Title:      "Project Note",
		EntityKind: DocumentEntityProject,
		EntityID:   &projectID,
	}, filePath))
	docs, _ := store.ListDocuments(false)
	docID := docs[0].ID

	require.NoError(t, store.DeleteDocument(docID))
	require.NoError(t, store.DeleteProject(projectID))
	require.ErrorContains(t, store.RestoreDocument(docID), "record not found")

	require.NoError(t, store.RestoreProject(projectID))
	require.NoError(t, store.RestoreDocument(docID))
}

func TestCreateDocumentRejectsInvalidEntityKind(t *testing.T) {
	store := newTestStore(t)
	filePath := filepath.Join(t.TempDir(), "doc.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("hello"), 0o600))
	id := uint(1)
	err := store.CreateDocument(Document{
		Title:      "Bad Link",
		EntityKind: "bogus",
		EntityID:   &id,
	}, filePath)
	require.ErrorContains(t, err, "invalid document entity kind")
}

func TestDocumentBLOBStorageAndExtract(t *testing.T) {
	store := newTestStore(t)

	content := []byte("this is a test PDF")
	filePath := filepath.Join(t.TempDir(), "report.pdf")
	require.NoError(t, os.WriteFile(filePath, content, 0o600))

	require.NoError(t, store.CreateDocument(Document{
		Title: "Test Report",
	}, filePath))

	docs, err := store.ListDocuments(false)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	assert.Equal(t, int64(len(content)), docs[0].SizeBytes)

	// ExtractDocument writes to cache and returns a path.
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	cachePath, err := store.ExtractDocument(docs[0].ID)
	require.NoError(t, err)
	assert.FileExists(t, cachePath)
	cached, err := os.ReadFile(cachePath) //nolint:gosec // test-only: path from ExtractDocument
	require.NoError(t, err)
	assert.Equal(t, content, cached)

	// Second call is a cache hit (same path, no error).
	cachePath2, err := store.ExtractDocument(docs[0].ID)
	require.NoError(t, err)
	assert.Equal(t, cachePath, cachePath2)
}

func TestCreateDocumentRequiresFile(t *testing.T) {
	store := newTestStore(t)
	err := store.CreateDocument(Document{Title: "No File"}, "")
	require.ErrorContains(t, err, "document file is required")
}

func TestCreateDocumentRejectsOversizedFile(t *testing.T) {
	store := newTestStore(t)
	filePath := filepath.Join(t.TempDir(), "huge.bin")
	// Create a sparse file that reports as larger than the limit without
	// actually allocating disk space.
	require.NoError(t, os.WriteFile(filePath, nil, 0o600))
	require.NoError(t, os.Truncate(filePath, MaxDocumentSize+1))

	err := store.CreateDocument(Document{Title: "Too Big"}, filePath)
	require.ErrorContains(t, err, "file is too large")
}

func TestTitleFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"invoice_2026_q1.pdf", "Invoice 2026 Q 1"},
		{"final-quote.PDF", "Final Quote"},
		{"README.md", "Readme"},
		{"my_great_project-notes.txt", "My Great Project Notes"},
		{"no-extension", "No Extension"},
		{"  spaced__out---file.txt  ", "Spaced Out File"},
		{"ALLCAPS.pdf", "Allcaps"},
		{"already Title Case.pdf", "Already Title Case"},
		{".hidden", "Hidden"},
		// Compound extensions stripped fully.
		{"archive.tar.gz", "Archive"},
		// CamelCase split into words.
		{"myGreatFile.pdf", "My Great File"},
		// Non-MIME dots preserved, stem split on dot boundaries.
		{"report.v2.final.pdf", "Report V 2 Final"},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			assert.Equal(t, tt.expected, titleFromFilename(tt.filename))
		})
	}
}

func TestDocumentTitleAutoFilledFromFilename(t *testing.T) {
	store := newTestStore(t)

	filePath := filepath.Join(t.TempDir(), "quarterly_budget-report.pdf")
	require.NoError(t, os.WriteFile(filePath, []byte("pdf content"), 0o600))

	require.NoError(t, store.CreateDocument(Document{}, filePath))
	docs, err := store.ListDocuments(false)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	assert.Equal(t, "Quarterly Budget Report", docs[0].Title)
}

func TestDocumentExplicitTitleNotOverwritten(t *testing.T) {
	store := newTestStore(t)

	filePath := filepath.Join(t.TempDir(), "boring_name.pdf")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o600))

	require.NoError(t, store.CreateDocument(Document{Title: "My Custom Title"}, filePath))
	docs, err := store.ListDocuments(false)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	assert.Equal(t, "My Custom Title", docs[0].Title)
}

func TestUpdateDocumentMetadataPreservesFile(t *testing.T) {
	store := newTestStore(t)

	content := []byte("important contract text")
	filePath := filepath.Join(t.TempDir(), "contract.pdf")
	require.NoError(t, os.WriteFile(filePath, content, 0o600))

	require.NoError(t, store.CreateDocument(Document{Title: "Contract"}, filePath))
	docs, err := store.ListDocuments(false)
	require.NoError(t, err)
	require.Len(t, docs, 1)

	original := docs[0]
	require.Equal(t, "contract.pdf", original.FileName)
	require.NotEmpty(t, original.ChecksumSHA256)
	require.Equal(t, int64(len(content)), original.SizeBytes)

	// Update only metadata -- no new file.
	require.NoError(t, store.UpdateDocument(Document{
		ID:    original.ID,
		Title: "Updated Contract",
		Notes: "added notes",
	}, ""))

	updated, err := store.GetDocument(original.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Contract", updated.Title)
	assert.Equal(t, "added notes", updated.Notes)
	// File metadata must be preserved.
	assert.Equal(t, original.FileName, updated.FileName)
	assert.Equal(t, original.MIMEType, updated.MIMEType)
	assert.Equal(t, original.SizeBytes, updated.SizeBytes)
	assert.Equal(t, original.ChecksumSHA256, updated.ChecksumSHA256)

	// Verify BLOB content is still intact.
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	cachePath, err := store.ExtractDocument(original.ID)
	require.NoError(t, err)
	cached, err := os.ReadFile(cachePath) //nolint:gosec // test-only path
	require.NoError(t, err)
	assert.Equal(t, content, cached)
}

func TestDocumentTitleRequiredOnEditWithoutFile(t *testing.T) {
	store := newTestStore(t)

	filePath := filepath.Join(t.TempDir(), "doc.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("hello"), 0o600))
	require.NoError(t, store.CreateDocument(Document{Title: "Original"}, filePath))

	docs, _ := store.ListDocuments(false)
	doc := docs[0]
	doc.Title = ""
	err := store.UpdateDocument(doc, "")
	require.ErrorContains(t, err, "document title is required")
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.AutoMigrate())
	require.NoError(t, store.SeedDefaults())
	return store
}

// newTestStoreWithDemoData creates a store pre-populated with randomized
// demo data from the given seed.
func newTestStoreWithDemoData(t *testing.T, seed uint64) *Store {
	t.Helper()
	store := newTestStore(t)
	require.NoError(t, store.SeedDemoDataFrom(fake.New(seed)))
	return store
}

func TestCountQuotesByProject(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()

	require.NoError(t, store.CreateProject(Project{
		Title: "P1", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projectID := projects[0].ID

	require.NoError(t, store.CreateQuote(
		Quote{ProjectID: projectID, TotalCents: 5000},
		Vendor{Name: "V1"},
	))
	require.NoError(t, store.CreateQuote(
		Quote{ProjectID: projectID, TotalCents: 7500},
		Vendor{Name: "V2"},
	))

	counts, err := store.CountQuotesByProject([]uint{projectID})
	require.NoError(t, err)
	assert.Equal(t, 2, counts[projectID])

	empty, err := store.CountQuotesByProject(nil)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestListQuotesByVendor(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()

	require.NoError(t, store.CreateVendor(Vendor{Name: "TestVendor"}))
	vendors, _ := store.ListVendors(false)
	vendorID := vendors[0].ID

	require.NoError(t, store.CreateProject(Project{
		Title: "P1", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	projectID := projects[0].ID

	require.NoError(t, store.CreateQuote(
		Quote{ProjectID: projectID, TotalCents: 1000},
		Vendor{Name: "TestVendor"},
	))
	require.NoError(t, store.CreateQuote(
		Quote{ProjectID: projectID, TotalCents: 2000},
		Vendor{Name: "OtherVendor"},
	))

	quotes, err := store.ListQuotesByVendor(vendorID, false)
	require.NoError(t, err)
	assert.Len(t, quotes, 1)
	assert.Equal(t, int64(1000), quotes[0].TotalCents)
}

func TestListQuotesByProject(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()

	require.NoError(t, store.CreateProject(Project{
		Title: "P1", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	require.NoError(t, store.CreateProject(Project{
		Title: "P2", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned,
	}))
	projects, _ := store.ListProjects(false)
	p1ID := projects[0].ID
	p2ID := projects[1].ID

	require.NoError(t, store.CreateQuote(
		Quote{ProjectID: p1ID, TotalCents: 1000},
		Vendor{Name: "V1"},
	))
	require.NoError(t, store.CreateQuote(
		Quote{ProjectID: p2ID, TotalCents: 5000},
		Vendor{Name: "V1"},
	))

	quotes, err := store.ListQuotesByProject(p1ID, false)
	require.NoError(t, err)
	assert.Len(t, quotes, 1)
	assert.Equal(t, int64(1000), quotes[0].TotalCents)
}

func TestListServiceLogsByVendor(t *testing.T) {
	store := newTestStore(t)
	cats, _ := store.MaintenanceCategories()

	require.NoError(t, store.CreateVendor(Vendor{Name: "LogVendor"}))
	vendors, _ := store.ListVendors(false)
	vendorID := vendors[0].ID

	require.NoError(t, store.CreateMaintenance(MaintenanceItem{
		Name: "Filter", CategoryID: cats[0].ID,
	}))
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	require.NoError(t, store.CreateServiceLog(
		ServiceLogEntry{MaintenanceItemID: maintID, ServicedAt: time.Now()},
		Vendor{Name: "LogVendor"},
	))
	require.NoError(t, store.CreateServiceLog(
		ServiceLogEntry{MaintenanceItemID: maintID, ServicedAt: time.Now()},
		Vendor{Name: "OtherVendor"},
	))

	entries, err := store.ListServiceLogsByVendor(vendorID, false)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "Filter", entries[0].MaintenanceItem.Name,
		"preloaded MaintenanceItem should be available")
}
