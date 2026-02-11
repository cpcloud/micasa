// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestSeedDefaults(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	if err != nil {
		t.Fatalf("ProjectTypes error: %v", err)
	}
	if len(types) == 0 {
		t.Fatalf("ProjectTypes empty")
	}
	categories, err := store.MaintenanceCategories()
	if err != nil {
		t.Fatalf("MaintenanceCategories error: %v", err)
	}
	if len(categories) == 0 {
		t.Fatalf("MaintenanceCategories empty")
	}
}

func TestHouseProfileSingle(t *testing.T) {
	store := newTestStore(t)
	profile := HouseProfile{Nickname: "Primary Residence"}
	if err := store.CreateHouseProfile(profile); err != nil {
		t.Fatalf("CreateHouseProfile error: %v", err)
	}
	if _, err := store.HouseProfile(); err != nil {
		t.Fatalf("HouseProfile error: %v", err)
	}
	if err := store.CreateHouseProfile(profile); err == nil {
		t.Fatalf("expected error on second profile")
	}
}

func TestUpdateHouseProfile(t *testing.T) {
	store := newTestStore(t)
	profile := HouseProfile{Nickname: "Primary Residence", City: "Portland"}
	if err := store.CreateHouseProfile(profile); err != nil {
		t.Fatalf("CreateHouseProfile error: %v", err)
	}
	updated := HouseProfile{Nickname: "Primary Residence", City: "Seattle"}
	if err := store.UpdateHouseProfile(updated); err != nil {
		t.Fatalf("UpdateHouseProfile error: %v", err)
	}
	fetched, err := store.HouseProfile()
	if err != nil {
		t.Fatalf("HouseProfile error: %v", err)
	}
	if fetched.City != "Seattle" {
		t.Fatalf("expected city Seattle, got %q", fetched.City)
	}
}

func TestSoftDeleteRestoreProject(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	if err != nil {
		t.Fatalf("ProjectTypes error: %v", err)
	}
	project := Project{
		Title:         "Test Project",
		ProjectTypeID: types[0].ID,
		Status:        ProjectStatusPlanned,
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("CreateProject error: %v", err)
	}
	projects, err := store.ListProjects(false)
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects expected 1, got %d err %v", len(projects), err)
	}
	if err := store.DeleteProject(projects[0].ID); err != nil {
		t.Fatalf("DeleteProject error: %v", err)
	}
	projects, err = store.ListProjects(false)
	if err != nil || len(projects) != 0 {
		t.Fatalf("ListProjects expected 0, got %d err %v", len(projects), err)
	}
	projects, err = store.ListProjects(true)
	if err != nil || len(projects) != 1 || !projects[0].DeletedAt.Valid {
		t.Fatalf("ListProjects expected deleted row")
	}
	if err := store.RestoreProject(projects[0].ID); err != nil {
		t.Fatalf("RestoreProject error: %v", err)
	}
	projects, err = store.ListProjects(false)
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects after restore expected 1, got %d err %v", len(projects), err)
	}
}

func TestLastDeletionRecord(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	if err != nil {
		t.Fatalf("ProjectTypes error: %v", err)
	}
	project := Project{
		Title:         "Test Project",
		ProjectTypeID: types[0].ID,
		Status:        ProjectStatusPlanned,
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("CreateProject error: %v", err)
	}
	projects, err := store.ListProjects(false)
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects expected 1, got %d err %v", len(projects), err)
	}
	if err := store.DeleteProject(projects[0].ID); err != nil {
		t.Fatalf("DeleteProject error: %v", err)
	}
	record, err := store.LastDeletion(DeletionEntityProject)
	if err != nil {
		t.Fatalf("LastDeletion error: %v", err)
	}
	if record.TargetID != projects[0].ID {
		t.Fatalf("LastDeletion target %d != %d", record.TargetID, projects[0].ID)
	}
	if err := store.RestoreProject(record.TargetID); err != nil {
		t.Fatalf("RestoreProject error: %v", err)
	}
	_, err = store.LastDeletion(DeletionEntityProject)
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestUpdateProject(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	if err != nil {
		t.Fatalf("ProjectTypes error: %v", err)
	}
	project := Project{
		Title:         "Original Title",
		ProjectTypeID: types[0].ID,
		Status:        ProjectStatusPlanned,
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("CreateProject error: %v", err)
	}
	projects, err := store.ListProjects(false)
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects expected 1, got %d err %v", len(projects), err)
	}
	id := projects[0].ID

	fetched, err := store.GetProject(id)
	if err != nil {
		t.Fatalf("GetProject error: %v", err)
	}
	if fetched.Title != "Original Title" {
		t.Fatalf("expected 'Original Title', got %q", fetched.Title)
	}

	updated := Project{
		ID:            id,
		Title:         "Updated Title",
		ProjectTypeID: types[0].ID,
		Status:        ProjectStatusInProgress,
	}
	if err := store.UpdateProject(updated); err != nil {
		t.Fatalf("UpdateProject error: %v", err)
	}

	fetched, err = store.GetProject(id)
	if err != nil {
		t.Fatalf("GetProject after update error: %v", err)
	}
	if fetched.Title != "Updated Title" {
		t.Fatalf("expected 'Updated Title', got %q", fetched.Title)
	}
	if fetched.Status != ProjectStatusInProgress {
		t.Fatalf("expected status %q, got %q", ProjectStatusInProgress, fetched.Status)
	}
}

func TestUpdateQuote(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	if err != nil {
		t.Fatalf("ProjectTypes error: %v", err)
	}
	project := Project{
		Title:         "Test Project",
		ProjectTypeID: types[0].ID,
		Status:        ProjectStatusPlanned,
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("CreateProject error: %v", err)
	}
	projects, err := store.ListProjects(false)
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects expected 1, got %d err %v", len(projects), err)
	}
	vendor := Vendor{Name: "Acme Corp"}
	quote := Quote{
		ProjectID:  projects[0].ID,
		TotalCents: 100000,
	}
	if err := store.CreateQuote(quote, vendor); err != nil {
		t.Fatalf("CreateQuote error: %v", err)
	}
	quotes, err := store.ListQuotes(false)
	if err != nil || len(quotes) != 1 {
		t.Fatalf("ListQuotes expected 1, got %d err %v", len(quotes), err)
	}
	id := quotes[0].ID

	updatedQuote := Quote{
		ID:         id,
		ProjectID:  projects[0].ID,
		TotalCents: 200000,
	}
	updatedVendor := Vendor{Name: "Acme Corp", ContactName: "John Doe"}
	if err := store.UpdateQuote(updatedQuote, updatedVendor); err != nil {
		t.Fatalf("UpdateQuote error: %v", err)
	}

	fetched, err := store.GetQuote(id)
	if err != nil {
		t.Fatalf("GetQuote after update error: %v", err)
	}
	if fetched.TotalCents != 200000 {
		t.Fatalf("expected total 200000, got %d", fetched.TotalCents)
	}
	if fetched.Vendor.ContactName != "John Doe" {
		t.Fatalf("expected contact 'John Doe', got %q", fetched.Vendor.ContactName)
	}
}

func TestUpdateMaintenance(t *testing.T) {
	store := newTestStore(t)
	categories, err := store.MaintenanceCategories()
	if err != nil {
		t.Fatalf("MaintenanceCategories error: %v", err)
	}
	item := MaintenanceItem{
		Name:       "Filter Change",
		CategoryID: categories[0].ID,
	}
	if err := store.CreateMaintenance(item); err != nil {
		t.Fatalf("CreateMaintenance error: %v", err)
	}
	items, err := store.ListMaintenance(false)
	if err != nil || len(items) != 1 {
		t.Fatalf("ListMaintenance expected 1, got %d err %v", len(items), err)
	}
	id := items[0].ID

	fetched, err := store.GetMaintenance(id)
	if err != nil {
		t.Fatalf("GetMaintenance error: %v", err)
	}
	if fetched.Name != "Filter Change" {
		t.Fatalf("expected 'Filter Change', got %q", fetched.Name)
	}

	updated := MaintenanceItem{
		ID:             id,
		Name:           "HVAC Filter Change",
		CategoryID:     categories[0].ID,
		IntervalMonths: 3,
	}
	if err := store.UpdateMaintenance(updated); err != nil {
		t.Fatalf("UpdateMaintenance error: %v", err)
	}

	fetched, err = store.GetMaintenance(id)
	if err != nil {
		t.Fatalf("GetMaintenance after update error: %v", err)
	}
	if fetched.Name != "HVAC Filter Change" {
		t.Fatalf("expected 'HVAC Filter Change', got %q", fetched.Name)
	}
	if fetched.IntervalMonths != 3 {
		t.Fatalf("expected interval 3, got %d", fetched.IntervalMonths)
	}
}

func TestServiceLogCRUD(t *testing.T) {
	store := newTestStore(t)
	categories, err := store.MaintenanceCategories()
	if err != nil {
		t.Fatalf("MaintenanceCategories error: %v", err)
	}
	item := MaintenanceItem{
		Name:       "Test Maintenance",
		CategoryID: categories[0].ID,
	}
	if err := store.CreateMaintenance(item); err != nil {
		t.Fatalf("CreateMaintenance error: %v", err)
	}
	items, err := store.ListMaintenance(false)
	if err != nil || len(items) != 1 {
		t.Fatalf("ListMaintenance expected 1, got %d err %v", len(items), err)
	}
	maintID := items[0].ID

	// Create a service log entry (self-performed, no vendor).
	entry := ServiceLogEntry{
		MaintenanceItemID: maintID,
		ServicedAt:        time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Notes:             "did it myself",
	}
	if err := store.CreateServiceLog(entry, Vendor{}); err != nil {
		t.Fatalf("CreateServiceLog error: %v", err)
	}

	entries, err := store.ListServiceLog(maintID, false)
	if err != nil || len(entries) != 1 {
		t.Fatalf("ListServiceLog expected 1, got %d err %v", len(entries), err)
	}
	if entries[0].VendorID != nil {
		t.Fatalf("expected nil VendorID for self-performed entry")
	}
	if entries[0].Notes != "did it myself" {
		t.Fatalf("expected notes 'did it myself', got %q", entries[0].Notes)
	}

	// Create a vendor-performed entry.
	vendorEntry := ServiceLogEntry{
		MaintenanceItemID: maintID,
		ServicedAt:        time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		CostCents:         func() *int64 { v := int64(15000); return &v }(),
		Notes:             "vendor did it",
	}
	vendor := Vendor{Name: "Test Plumber", Phone: "555-555-0001"}
	if err := store.CreateServiceLog(vendorEntry, vendor); err != nil {
		t.Fatalf("CreateServiceLog with vendor error: %v", err)
	}

	entries, err = store.ListServiceLog(maintID, false)
	if err != nil || len(entries) != 2 {
		t.Fatalf("ListServiceLog expected 2, got %d err %v", len(entries), err)
	}
	// Most recent first (2026-02-01 before 2026-01-15).
	if entries[0].VendorID == nil {
		t.Fatalf("expected vendor on first entry")
	}
	if entries[0].Vendor.Name != "Test Plumber" {
		t.Fatalf("expected vendor 'Test Plumber', got %q", entries[0].Vendor.Name)
	}

	// Update: change vendor entry to self-performed.
	updated := entries[0]
	updated.Notes = "actually did it myself"
	if err := store.UpdateServiceLog(updated, Vendor{}); err != nil {
		t.Fatalf("UpdateServiceLog error: %v", err)
	}
	fetched, err := store.GetServiceLog(updated.ID)
	if err != nil {
		t.Fatalf("GetServiceLog error: %v", err)
	}
	if fetched.VendorID != nil {
		t.Fatalf("expected nil VendorID after update")
	}
	if fetched.Notes != "actually did it myself" {
		t.Fatalf("expected updated notes, got %q", fetched.Notes)
	}

	// Delete and restore.
	if err := store.DeleteServiceLog(fetched.ID); err != nil {
		t.Fatalf("DeleteServiceLog error: %v", err)
	}
	entries, err = store.ListServiceLog(maintID, false)
	if err != nil || len(entries) != 1 {
		t.Fatalf("after delete expected 1, got %d err %v", len(entries), err)
	}
	entries, err = store.ListServiceLog(maintID, true)
	if err != nil || len(entries) != 2 {
		t.Fatalf("with deleted expected 2, got %d err %v", len(entries), err)
	}
	if err := store.RestoreServiceLog(fetched.ID); err != nil {
		t.Fatalf("RestoreServiceLog error: %v", err)
	}
	entries, err = store.ListServiceLog(maintID, false)
	if err != nil || len(entries) != 2 {
		t.Fatalf("after restore expected 2, got %d err %v", len(entries), err)
	}

	// CountServiceLogs.
	counts, err := store.CountServiceLogs([]uint{maintID})
	if err != nil {
		t.Fatalf("CountServiceLogs error: %v", err)
	}
	if counts[maintID] != 2 {
		t.Fatalf("expected count 2, got %d", counts[maintID])
	}
}

func TestSoftDeletePersistsAcrossRuns(t *testing.T) {
	// Verify that soft-deleted items stay hidden after closing and reopening
	// the database, confirming cross-session persistence.
	path := filepath.Join(t.TempDir(), "persist.db")

	// Session 1: create a project, then soft-delete it.
	store1, err := Open(path)
	if err != nil {
		t.Fatalf("Open session 1: %v", err)
	}
	if err := store1.AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate session 1: %v", err)
	}
	if err := store1.SeedDefaults(); err != nil {
		t.Fatalf("SeedDefaults session 1: %v", err)
	}
	types, _ := store1.ProjectTypes()
	if err := store1.CreateProject(Project{Title: "Persist Test", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	// Find the created project by listing.
	projects, _ := store1.ListProjects(false)
	var projectID uint
	for _, p := range projects {
		if p.Title == "Persist Test" {
			projectID = p.ID
			break
		}
	}
	if projectID == 0 {
		t.Fatal("could not find created project")
	}
	if err := store1.DeleteProject(projectID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
	_ = store1.Close()

	// Session 2: reopen and verify the project is still soft-deleted.
	store2, err := Open(path)
	if err != nil {
		t.Fatalf("Open session 2: %v", err)
	}
	t.Cleanup(func() { _ = store2.Close() })
	if err := store2.AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate session 2: %v", err)
	}

	// Normal list should not include the deleted project.
	projects2, err := store2.ListProjects(false)
	if err != nil {
		t.Fatalf("ListProjects(false): %v", err)
	}
	for _, p := range projects2 {
		if p.ID == projectID {
			t.Fatal("soft-deleted project should not appear in normal listing after reopen")
		}
	}

	// Unscoped list should include it.
	projectsAll, err := store2.ListProjects(true)
	if err != nil {
		t.Fatalf("ListProjects(true): %v", err)
	}
	found := false
	for _, p := range projectsAll {
		if p.ID == projectID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("soft-deleted project should appear in unscoped listing after reopen")
	}

	// Restore should still work.
	if err := store2.RestoreProject(projectID); err != nil {
		t.Fatalf("RestoreProject: %v", err)
	}
	projects3, err := store2.ListProjects(false)
	if err != nil {
		t.Fatalf("ListProjects after restore: %v", err)
	}
	found = false
	for _, p := range projects3 {
		if p.ID == projectID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("restored project should appear in normal listing")
	}
}

func TestVendorCRUD(t *testing.T) {
	store := newTestStore(t)

	// Create.
	v := Vendor{
		Name:        "Test Vendor",
		ContactName: "Alice",
		Email:       "alice@example.com",
		Phone:       "555-0001",
	}
	if err := store.CreateVendor(v); err != nil {
		t.Fatalf("CreateVendor: %v", err)
	}

	// List.
	vendors, err := store.ListVendors()
	if err != nil {
		t.Fatalf("ListVendors: %v", err)
	}
	if len(vendors) != 1 {
		t.Fatalf("expected 1 vendor, got %d", len(vendors))
	}
	got := vendors[0]
	if got.Name != "Test Vendor" || got.ContactName != "Alice" {
		t.Fatalf("unexpected vendor: %+v", got)
	}

	// Get.
	fetched, err := store.GetVendor(got.ID)
	if err != nil {
		t.Fatalf("GetVendor: %v", err)
	}
	if fetched.Email != "alice@example.com" {
		t.Fatalf("expected email alice@example.com, got %s", fetched.Email)
	}

	// Update.
	fetched.Phone = "555-9999"
	fetched.Website = "https://example.com"
	if err := store.UpdateVendor(fetched); err != nil {
		t.Fatalf("UpdateVendor: %v", err)
	}
	updated, _ := store.GetVendor(fetched.ID)
	if updated.Phone != "555-9999" || updated.Website != "https://example.com" {
		t.Fatalf("update didn't stick: %+v", updated)
	}
}

func TestCountQuotesByVendor(t *testing.T) {
	store := newTestStore(t)

	// Seed a vendor and a project.
	v := Vendor{Name: "Quote Vendor"}
	if err := store.CreateVendor(v); err != nil {
		t.Fatalf("CreateVendor: %v", err)
	}
	vendors, _ := store.ListVendors()
	vendorID := vendors[0].ID

	types, _ := store.ProjectTypes()
	p := Project{Title: "Test", ProjectTypeID: types[0].ID, Status: ProjectStatusPlanned}
	if err := store.CreateProject(p); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	projects, _ := store.ListProjects(false)
	projectID := projects[0].ID

	// Create two quotes for this vendor.
	for i := 0; i < 2; i++ {
		q := Quote{ProjectID: projectID, TotalCents: 100000}
		if err := store.CreateQuote(q, Vendor{Name: "Quote Vendor"}); err != nil {
			t.Fatalf("CreateQuote: %v", err)
		}
	}

	counts, err := store.CountQuotesByVendor([]uint{vendorID})
	if err != nil {
		t.Fatalf("CountQuotesByVendor: %v", err)
	}
	if counts[vendorID] != 2 {
		t.Fatalf("expected 2 quotes, got %d", counts[vendorID])
	}

	// Empty input.
	empty, err := store.CountQuotesByVendor(nil)
	if err != nil {
		t.Fatalf("CountQuotesByVendor(nil): %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty map, got %v", empty)
	}
}

func TestCountServiceLogsByVendor(t *testing.T) {
	store := newTestStore(t)

	v := Vendor{Name: "Job Vendor"}
	if err := store.CreateVendor(v); err != nil {
		t.Fatalf("CreateVendor: %v", err)
	}
	vendors, _ := store.ListVendors()
	vendorID := vendors[0].ID

	cats, _ := store.MaintenanceCategories()
	m := MaintenanceItem{Name: "Filter", CategoryID: cats[0].ID}
	if err := store.CreateMaintenance(m); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	now := time.Now()
	entry := ServiceLogEntry{MaintenanceItemID: maintID, ServicedAt: now}
	if err := store.CreateServiceLog(entry, Vendor{Name: "Job Vendor"}); err != nil {
		t.Fatalf("CreateServiceLog: %v", err)
	}

	counts, err := store.CountServiceLogsByVendor([]uint{vendorID})
	if err != nil {
		t.Fatalf("CountServiceLogsByVendor: %v", err)
	}
	if counts[vendorID] != 1 {
		t.Fatalf("expected 1 job, got %d", counts[vendorID])
	}
}

func TestDeleteProjectBlockedByQuotes(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()
	project := Project{
		Title: "Blocked Project", ProjectTypeID: types[0].ID,
		Status: ProjectStatusPlanned,
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	projects, _ := store.ListProjects(false)
	projID := projects[0].ID

	// Attach a quote.
	quote := Quote{ProjectID: projID, TotalCents: 1000}
	if err := store.CreateQuote(quote, Vendor{Name: "V1"}); err != nil {
		t.Fatalf("CreateQuote: %v", err)
	}

	// Delete should be refused.
	err := store.DeleteProject(projID)
	if err == nil {
		t.Fatal("expected error deleting project with active quotes")
	}
	if !strings.Contains(err.Error(), "active quote") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Soft-delete the quote, then project deletion should succeed.
	quotes, _ := store.ListQuotes(false)
	if err := store.DeleteQuote(quotes[0].ID); err != nil {
		t.Fatalf("DeleteQuote: %v", err)
	}
	if err := store.DeleteProject(projID); err != nil {
		t.Fatalf("DeleteProject after quote removed: %v", err)
	}
}

func TestRestoreQuoteBlockedByDeletedProject(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()
	project := Project{
		Title: "Doomed Project", ProjectTypeID: types[0].ID,
		Status: ProjectStatusPlanned,
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	projects, _ := store.ListProjects(false)
	projID := projects[0].ID

	quote := Quote{ProjectID: projID, TotalCents: 500}
	if err := store.CreateQuote(quote, Vendor{Name: "V2"}); err != nil {
		t.Fatalf("CreateQuote: %v", err)
	}
	quotes, _ := store.ListQuotes(false)
	quoteID := quotes[0].ID

	// Delete quote, then project.
	if err := store.DeleteQuote(quoteID); err != nil {
		t.Fatalf("DeleteQuote: %v", err)
	}
	if err := store.DeleteProject(projID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	// Restoring the quote should be refused while project is deleted.
	err := store.RestoreQuote(quoteID)
	if err == nil {
		t.Fatal("expected error restoring quote with deleted project")
	}
	if !strings.Contains(err.Error(), "project is deleted") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Restore the project, then quote restore should succeed.
	if err := store.RestoreProject(projID); err != nil {
		t.Fatalf("RestoreProject: %v", err)
	}
	if err := store.RestoreQuote(quoteID); err != nil {
		t.Fatalf("RestoreQuote after project restored: %v", err)
	}
}

func TestRestoreServiceLogBlockedByDeletedMaintenance(t *testing.T) {
	store := newTestStore(t)
	cats, _ := store.MaintenanceCategories()
	item := MaintenanceItem{
		Name: "Doomed Maint", CategoryID: cats[0].ID, IntervalMonths: 6,
	}
	if err := store.CreateMaintenance(item); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	now := time.Now()
	entry := ServiceLogEntry{MaintenanceItemID: maintID, ServicedAt: now}
	if err := store.CreateServiceLog(entry, Vendor{Name: "SL2"}); err != nil {
		t.Fatalf("CreateServiceLog: %v", err)
	}
	logs, _ := store.ListServiceLog(maintID, false)
	logID := logs[0].ID

	// Delete service log, then maintenance.
	if err := store.DeleteServiceLog(logID); err != nil {
		t.Fatalf("DeleteServiceLog: %v", err)
	}
	if err := store.DeleteMaintenance(maintID); err != nil {
		t.Fatalf("DeleteMaintenance: %v", err)
	}

	// Restoring the service log should be refused.
	err := store.RestoreServiceLog(logID)
	if err == nil {
		t.Fatal("expected error restoring service log with deleted maintenance")
	}
	if !strings.Contains(err.Error(), "maintenance item is deleted") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Restore maintenance, then service log restore should succeed.
	if err := store.RestoreMaintenance(maintID); err != nil {
		t.Fatalf("RestoreMaintenance: %v", err)
	}
	if err := store.RestoreServiceLog(logID); err != nil {
		t.Fatalf("RestoreServiceLog after maintenance restored: %v", err)
	}
}

func TestDeleteMaintenanceBlockedByServiceLogs(t *testing.T) {
	store := newTestStore(t)
	cats, _ := store.MaintenanceCategories()
	item := MaintenanceItem{
		Name: "Blocked Maint", CategoryID: cats[0].ID, IntervalMonths: 3,
	}
	if err := store.CreateMaintenance(item); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	// Attach a service log.
	now := time.Now()
	entry := ServiceLogEntry{MaintenanceItemID: maintID, ServicedAt: now}
	if err := store.CreateServiceLog(entry, Vendor{Name: "SL Vendor"}); err != nil {
		t.Fatalf("CreateServiceLog: %v", err)
	}

	// Delete should be refused.
	err := store.DeleteMaintenance(maintID)
	if err == nil {
		t.Fatal("expected error deleting maintenance with active service logs")
	}
	if !strings.Contains(err.Error(), "service log") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Soft-delete the service log, then maintenance deletion should succeed.
	logs, _ := store.ListServiceLog(maintID, false)
	if err := store.DeleteServiceLog(logs[0].ID); err != nil {
		t.Fatalf("DeleteServiceLog: %v", err)
	}
	if err := store.DeleteMaintenance(maintID); err != nil {
		t.Fatalf("DeleteMaintenance after logs removed: %v", err)
	}
}

func TestPartialQuoteDeletionStillBlocksProjectDelete(t *testing.T) {
	store := newTestStore(t)
	types, _ := store.ProjectTypes()
	project := Project{
		Title: "Multi-Quote", ProjectTypeID: types[0].ID,
		Status: ProjectStatusPlanned,
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	projects, _ := store.ListProjects(false)
	projID := projects[0].ID

	// Attach two quotes.
	for _, name := range []string{"Vendor A", "Vendor B"} {
		q := Quote{ProjectID: projID, TotalCents: 1000}
		if err := store.CreateQuote(q, Vendor{Name: name}); err != nil {
			t.Fatalf("CreateQuote: %v", err)
		}
	}
	quotes, _ := store.ListQuotes(false)
	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}

	// Delete one quote; project delete should still be blocked.
	if err := store.DeleteQuote(quotes[0].ID); err != nil {
		t.Fatalf("DeleteQuote: %v", err)
	}
	err := store.DeleteProject(projID)
	if err == nil {
		t.Fatal("expected error: one active quote remains")
	}
	if !strings.Contains(err.Error(), "1 active quote") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Delete the second quote; project delete should now succeed.
	if err := store.DeleteQuote(quotes[1].ID); err != nil {
		t.Fatalf("DeleteQuote: %v", err)
	}
	if err := store.DeleteProject(projID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
}

func TestRestoreMaintenanceBlockedByDeletedAppliance(t *testing.T) {
	store := newTestStore(t)
	appliance := Appliance{Name: "Doomed Fridge"}
	if err := store.CreateAppliance(appliance); err != nil {
		t.Fatalf("CreateAppliance: %v", err)
	}
	appliances, _ := store.ListAppliances(false)
	appID := appliances[0].ID

	cats, _ := store.MaintenanceCategories()
	item := MaintenanceItem{
		Name: "Coil Cleaning", CategoryID: cats[0].ID, IntervalMonths: 6,
		ApplianceID: &appID,
	}
	if err := store.CreateMaintenance(item); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	// Delete maintenance then appliance.
	if err := store.DeleteMaintenance(maintID); err != nil {
		t.Fatalf("DeleteMaintenance: %v", err)
	}
	if err := store.DeleteAppliance(appID); err != nil {
		t.Fatalf("DeleteAppliance: %v", err)
	}

	// Restore maintenance should be blocked -- the link existed, so the
	// appliance must be alive.
	err := store.RestoreMaintenance(maintID)
	if err == nil {
		t.Fatal("expected error restoring maintenance with deleted appliance")
	}
	if !strings.Contains(err.Error(), "appliance is deleted") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Restore appliance, then maintenance restore should succeed.
	if err := store.RestoreAppliance(appID); err != nil {
		t.Fatalf("RestoreAppliance: %v", err)
	}
	if err := store.RestoreMaintenance(maintID); err != nil {
		t.Fatalf("RestoreMaintenance after appliance restored: %v", err)
	}
}

func TestRestoreMaintenanceAllowedWithoutAppliance(t *testing.T) {
	// Maintenance items with no appliance link (nil ApplianceID) should
	// restore freely -- the nullable FK means "no link", not "broken link".
	store := newTestStore(t)
	cats, _ := store.MaintenanceCategories()
	item := MaintenanceItem{
		Name: "Gutter Cleaning", CategoryID: cats[0].ID, IntervalMonths: 6,
	}
	if err := store.CreateMaintenance(item); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	if err := store.DeleteMaintenance(maintID); err != nil {
		t.Fatalf("DeleteMaintenance: %v", err)
	}
	if err := store.RestoreMaintenance(maintID); err != nil {
		t.Fatalf("RestoreMaintenance should succeed with nil ApplianceID: %v", err)
	}
}

func TestThreeLevelDeleteRestoreChain(t *testing.T) {
	// Full Appliance → Maintenance → ServiceLog lifecycle exercising guards
	// at every level.
	store := newTestStore(t)

	// Set up the three-level chain.
	appliance := Appliance{Name: "HVAC Unit"}
	if err := store.CreateAppliance(appliance); err != nil {
		t.Fatalf("CreateAppliance: %v", err)
	}
	appliances, _ := store.ListAppliances(false)
	appID := appliances[0].ID

	cats, _ := store.MaintenanceCategories()
	item := MaintenanceItem{
		Name: "Filter Change", CategoryID: cats[0].ID, IntervalMonths: 3,
		ApplianceID: &appID,
	}
	if err := store.CreateMaintenance(item); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}
	items, _ := store.ListMaintenance(false)
	maintID := items[0].ID

	entry := ServiceLogEntry{
		MaintenanceItemID: maintID,
		ServicedAt:        time.Now(),
	}
	if err := store.CreateServiceLog(entry, Vendor{}); err != nil {
		t.Fatalf("CreateServiceLog: %v", err)
	}
	logs, _ := store.ListServiceLog(maintID, false)
	logID := logs[0].ID

	// --- Delete bottom-up ---
	// Can't delete maintenance while service log is active.
	err := store.DeleteMaintenance(maintID)
	if err == nil {
		t.Fatal("expected error: active service log blocks maintenance delete")
	}

	if err := store.DeleteServiceLog(logID); err != nil {
		t.Fatalf("DeleteServiceLog: %v", err)
	}
	if err := store.DeleteMaintenance(maintID); err != nil {
		t.Fatalf("DeleteMaintenance: %v", err)
	}
	if err := store.DeleteAppliance(appID); err != nil {
		t.Fatalf("DeleteAppliance: %v", err)
	}

	// --- Attempt wrong-order restores ---
	// Can't restore service log while maintenance is deleted.
	err = store.RestoreServiceLog(logID)
	if err == nil {
		t.Fatal("expected error: maintenance is deleted")
	}
	if !strings.Contains(err.Error(), "maintenance item is deleted") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Can't restore maintenance while its linked appliance is deleted.
	err = store.RestoreMaintenance(maintID)
	if err == nil {
		t.Fatal("expected error: appliance is deleted")
	}
	if !strings.Contains(err.Error(), "appliance is deleted") {
		t.Fatalf("unexpected error: %v", err)
	}

	// --- Restore correct order: appliance → maintenance → service log ---
	if err := store.RestoreAppliance(appID); err != nil {
		t.Fatalf("RestoreAppliance: %v", err)
	}
	if err := store.RestoreMaintenance(maintID); err != nil {
		t.Fatalf("RestoreMaintenance: %v", err)
	}
	if err := store.RestoreServiceLog(logID); err != nil {
		t.Fatalf("RestoreServiceLog: %v", err)
	}

	// Verify everything is alive and linked.
	fetched, err := store.GetMaintenance(maintID)
	if err != nil {
		t.Fatalf("GetMaintenance: %v", err)
	}
	if fetched.ApplianceID == nil || *fetched.ApplianceID != appID {
		t.Fatal("maintenance should still reference restored appliance")
	}
	restoredLogs, err := store.ListServiceLog(maintID, false)
	if err != nil || len(restoredLogs) != 1 {
		t.Fatalf("expected 1 service log, got %d err %v", len(restoredLogs), err)
	}
}

func TestDeleteApplianceAllowedWithMaintenance(t *testing.T) {
	store := newTestStore(t)
	appliance := Appliance{Name: "Deletable Fridge"}
	if err := store.CreateAppliance(appliance); err != nil {
		t.Fatalf("CreateAppliance: %v", err)
	}
	appliances, _ := store.ListAppliances(false)
	appID := appliances[0].ID

	// Attach a maintenance item.
	cats, _ := store.MaintenanceCategories()
	item := MaintenanceItem{
		Name: "Filter", CategoryID: cats[0].ID, IntervalMonths: 6,
		ApplianceID: &appID,
	}
	if err := store.CreateMaintenance(item); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}

	// Appliance deletion should succeed (SET NULL semantics).
	if err := store.DeleteAppliance(appID); err != nil {
		t.Fatalf("DeleteAppliance should be allowed: %v", err)
	}
}

func TestGetAppliance(t *testing.T) {
	store := newTestStore(t)
	if err := store.CreateAppliance(Appliance{Name: "Fridge"}); err != nil {
		t.Fatalf("CreateAppliance: %v", err)
	}
	got, err := store.GetAppliance(1)
	if err != nil {
		t.Fatalf("GetAppliance: %v", err)
	}
	if got.Name != "Fridge" {
		t.Fatalf("expected name 'Fridge', got %q", got.Name)
	}
}

func TestGetApplianceNotFound(t *testing.T) {
	store := newTestStore(t)
	_, err := store.GetAppliance(9999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestUpdateAppliance(t *testing.T) {
	store := newTestStore(t)
	if err := store.CreateAppliance(Appliance{Name: "Fridge"}); err != nil {
		t.Fatalf("CreateAppliance: %v", err)
	}
	got, _ := store.GetAppliance(1)
	got.Brand = "Samsung"
	if err := store.UpdateAppliance(got); err != nil {
		t.Fatalf("UpdateAppliance: %v", err)
	}
	updated, _ := store.GetAppliance(1)
	if updated.Brand != "Samsung" {
		t.Fatalf("expected brand 'Samsung', got %q", updated.Brand)
	}
}

func TestListMaintenanceByAppliance(t *testing.T) {
	store := newTestStore(t)
	categories, _ := store.MaintenanceCategories()
	catID := categories[0].ID

	if err := store.CreateAppliance(Appliance{Name: "Fridge"}); err != nil {
		t.Fatalf("CreateAppliance: %v", err)
	}
	appID := uint(1)
	if err := store.CreateMaintenance(MaintenanceItem{
		Name:        "Clean coils",
		CategoryID:  catID,
		ApplianceID: &appID,
	}); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}
	// Create a maintenance item without this appliance.
	if err := store.CreateMaintenance(MaintenanceItem{
		Name:       "Check smoke detectors",
		CategoryID: catID,
	}); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}

	items, err := store.ListMaintenanceByAppliance(appID, false)
	if err != nil {
		t.Fatalf("ListMaintenanceByAppliance: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Name != "Clean coils" {
		t.Fatalf("expected 'Clean coils', got %q", items[0].Name)
	}
}

func TestCountMaintenanceByAppliance(t *testing.T) {
	store := newTestStore(t)
	categories, _ := store.MaintenanceCategories()
	catID := categories[0].ID

	if err := store.CreateAppliance(Appliance{Name: "Fridge"}); err != nil {
		t.Fatalf("CreateAppliance: %v", err)
	}
	appID := uint(1)
	for _, name := range []string{"Clean coils", "Replace filter"} {
		if err := store.CreateMaintenance(MaintenanceItem{
			Name:        name,
			CategoryID:  catID,
			ApplianceID: &appID,
		}); err != nil {
			t.Fatalf("CreateMaintenance: %v", err)
		}
	}

	counts, err := store.CountMaintenanceByAppliance([]uint{appID})
	if err != nil {
		t.Fatalf("CountMaintenanceByAppliance: %v", err)
	}
	if counts[appID] != 2 {
		t.Fatalf("expected count 2, got %d", counts[appID])
	}
}

func TestUpdateServiceLog(t *testing.T) {
	store := newTestStore(t)
	categories, _ := store.MaintenanceCategories()
	catID := categories[0].ID

	if err := store.CreateMaintenance(MaintenanceItem{
		Name:       "HVAC filter",
		CategoryID: catID,
	}); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}
	now := time.Now().Truncate(time.Second)
	entry := ServiceLogEntry{
		MaintenanceItemID: 1,
		ServicedAt:        now,
		Notes:             "initial",
	}
	if err := store.CreateServiceLog(entry, Vendor{}); err != nil {
		t.Fatalf("CreateServiceLog: %v", err)
	}

	// Update with a new vendor.
	created, _ := store.GetServiceLog(1)
	created.Notes = "updated"
	vendor := Vendor{Name: "HVAC Pros"}
	if err := store.UpdateServiceLog(created, vendor); err != nil {
		t.Fatalf("UpdateServiceLog: %v", err)
	}

	updated, _ := store.GetServiceLog(1)
	if updated.Notes != "updated" {
		t.Fatalf("expected 'updated', got %q", updated.Notes)
	}
	if updated.VendorID == nil {
		t.Fatal("expected vendor to be set after update")
	}
}

func TestUpdateServiceLogClearVendor(t *testing.T) {
	store := newTestStore(t)
	categories, _ := store.MaintenanceCategories()
	catID := categories[0].ID

	if err := store.CreateMaintenance(MaintenanceItem{
		Name:       "HVAC filter",
		CategoryID: catID,
	}); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}
	now := time.Now().Truncate(time.Second)
	entry := ServiceLogEntry{
		MaintenanceItemID: 1,
		ServicedAt:        now,
	}
	vendor := Vendor{Name: "HVAC Pros"}
	if err := store.CreateServiceLog(entry, vendor); err != nil {
		t.Fatalf("CreateServiceLog: %v", err)
	}

	// Update with empty vendor name -- should clear the vendor.
	created, _ := store.GetServiceLog(1)
	if err := store.UpdateServiceLog(created, Vendor{}); err != nil {
		t.Fatalf("UpdateServiceLog: %v", err)
	}
	updated, _ := store.GetServiceLog(1)
	if updated.VendorID != nil {
		t.Fatal("expected vendor to be cleared")
	}
}

func TestListMaintenanceByApplianceIncludeDeleted(t *testing.T) {
	store := newTestStore(t)
	categories, _ := store.MaintenanceCategories()
	catID := categories[0].ID

	if err := store.CreateAppliance(Appliance{Name: "Fridge"}); err != nil {
		t.Fatalf("CreateAppliance: %v", err)
	}
	appID := uint(1)
	if err := store.CreateMaintenance(MaintenanceItem{
		Name:        "Clean coils",
		CategoryID:  catID,
		ApplianceID: &appID,
	}); err != nil {
		t.Fatalf("CreateMaintenance: %v", err)
	}
	if err := store.DeleteMaintenance(1); err != nil {
		t.Fatalf("DeleteMaintenance: %v", err)
	}

	// Without deleted.
	items, _ := store.ListMaintenanceByAppliance(appID, false)
	if len(items) != 0 {
		t.Fatalf("expected 0 items without deleted, got %d", len(items))
	}

	// With deleted.
	items, _ = store.ListMaintenanceByAppliance(appID, true)
	if len(items) != 1 {
		t.Fatalf("expected 1 item with deleted, got %d", len(items))
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate error: %v", err)
	}
	if err := store.SeedDefaults(); err != nil {
		t.Fatalf("SeedDefaults error: %v", err)
	}
	return store
}
