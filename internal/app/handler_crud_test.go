// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
	"time"

	"github.com/cpcloud/micasa/internal/data"
)

// ---------------------------------------------------------------------------
// projectHandler CRUD
// ---------------------------------------------------------------------------

func TestProjectHandlerLoadDeleteRestoreRoundTrip(t *testing.T) {
	m := newTestModelWithStore(t)
	h := projectHandler{}

	// Create a project via form data.
	m.formData = &projectFormData{
		Title:         "Deck Build",
		ProjectTypeID: m.projectTypes[0].ID,
		Status:        data.ProjectStatusPlanned,
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatalf("SubmitForm (create): %v", err)
	}

	// Load should return the project.
	rows, meta, cells, err := h.Load(m.store, false)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(meta) != 1 || len(cells) != 1 {
		t.Fatalf("meta/cells length mismatch")
	}
	id := meta[0].ID

	// Delete.
	if err := h.Delete(m.store, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Load without deleted should be empty.
	rows, _, _, err = h.Load(m.store, false)
	if err != nil {
		t.Fatalf("Load after delete: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 rows after delete, got %d", len(rows))
	}

	// Load with deleted should show it.
	rows, _, _, err = h.Load(m.store, true)
	if err != nil {
		t.Fatalf("Load with deleted: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row with showDeleted, got %d", len(rows))
	}

	// Restore.
	if err := h.Restore(m.store, id); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	rows, _, _, _ = h.Load(m.store, false)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row after restore, got %d", len(rows))
	}
}

func TestProjectHandlerEditRoundTrip(t *testing.T) {
	m := newTestModelWithStore(t)
	h := projectHandler{}

	// Create.
	m.formData = &projectFormData{
		Title:         "Paint Fence",
		ProjectTypeID: m.projectTypes[0].ID,
		Status:        data.ProjectStatusIdeating,
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatal(err)
	}
	_, meta, _, _ := h.Load(m.store, false)
	id := meta[0].ID

	// Edit via form data.
	editID := id
	m.editID = &editID
	m.formData = &projectFormData{
		Title:         "Paint Fence Red",
		ProjectTypeID: m.projectTypes[0].ID,
		Status:        data.ProjectStatusInProgress,
		Budget:        "500.00",
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatalf("SubmitForm (edit): %v", err)
	}
	m.editID = nil

	project, err := m.store.GetProject(id)
	if err != nil {
		t.Fatal(err)
	}
	if project.Title != "Paint Fence Red" {
		t.Errorf("title = %q, want Paint Fence Red", project.Title)
	}
	if project.Status != data.ProjectStatusInProgress {
		t.Errorf("status = %q", project.Status)
	}
	if project.BudgetCents == nil || *project.BudgetCents != 50000 {
		t.Errorf("budget = %v, want 50000", project.BudgetCents)
	}
}

func TestProjectHandlerSnapshot(t *testing.T) {
	m := newTestModelWithStore(t)
	h := projectHandler{}

	m.formData = &projectFormData{
		Title:         "Roof Repair",
		ProjectTypeID: m.projectTypes[0].ID,
		Status:        data.ProjectStatusPlanned,
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatal(err)
	}
	_, meta, _, _ := h.Load(m.store, false)
	id := meta[0].ID

	entry, ok := h.Snapshot(m.store, id)
	if !ok {
		t.Fatal("Snapshot returned false")
	}
	if entry.FormKind != formProject {
		t.Errorf("FormKind = %d, want formProject", entry.FormKind)
	}
	if entry.EntityID != id {
		t.Errorf("EntityID = %d, want %d", entry.EntityID, id)
	}
}

// ---------------------------------------------------------------------------
// applianceHandler CRUD
// ---------------------------------------------------------------------------

func TestApplianceHandlerLoadDeleteRestoreRoundTrip(t *testing.T) {
	m := newTestModelWithStore(t)
	h := applianceHandler{}

	m.formData = &applianceFormData{Name: "Washer"}
	if err := h.SubmitForm(m); err != nil {
		t.Fatalf("SubmitForm: %v", err)
	}

	rows, meta, _, err := h.Load(m.store, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	id := meta[0].ID

	if err := h.Delete(m.store, id); err != nil {
		t.Fatal(err)
	}
	rows, _, _, _ = h.Load(m.store, false)
	if len(rows) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(rows))
	}

	if err := h.Restore(m.store, id); err != nil {
		t.Fatal(err)
	}
	rows, _, _, _ = h.Load(m.store, false)
	if len(rows) != 1 {
		t.Fatalf("expected 1 after restore, got %d", len(rows))
	}
}

func TestApplianceHandlerEditRoundTrip(t *testing.T) {
	m := newTestModelWithStore(t)
	h := applianceHandler{}

	m.formData = &applianceFormData{Name: "Dryer"}
	if err := h.SubmitForm(m); err != nil {
		t.Fatal(err)
	}
	_, meta, _, _ := h.Load(m.store, false)
	id := meta[0].ID

	editID := id
	m.editID = &editID
	m.formData = &applianceFormData{
		Name:  "Dryer",
		Brand: "LG",
		Cost:  "800.00",
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatal(err)
	}
	m.editID = nil

	app, _ := m.store.GetAppliance(id)
	if app.Brand != "LG" {
		t.Errorf("brand = %q", app.Brand)
	}
	if app.CostCents == nil || *app.CostCents != 80000 {
		t.Errorf("cost = %v, want 80000", app.CostCents)
	}
}

// ---------------------------------------------------------------------------
// maintenanceHandler CRUD
// ---------------------------------------------------------------------------

func TestMaintenanceHandlerLoadDeleteRestoreRoundTrip(t *testing.T) {
	m := newTestModelWithStore(t)
	h := maintenanceHandler{}
	cats, _ := m.store.MaintenanceCategories()

	m.formData = &maintenanceFormData{
		Name:           "Change Air Filter",
		CategoryID:     cats[0].ID,
		IntervalMonths: "3",
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatal(err)
	}

	rows, meta, _, err := h.Load(m.store, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1, got %d", len(rows))
	}
	id := meta[0].ID

	if err := h.Delete(m.store, id); err != nil {
		t.Fatal(err)
	}
	rows, _, _, _ = h.Load(m.store, false)
	if len(rows) != 0 {
		t.Fatal("expected 0 after delete")
	}

	if err := h.Restore(m.store, id); err != nil {
		t.Fatal(err)
	}
	rows, _, _, _ = h.Load(m.store, false)
	if len(rows) != 1 {
		t.Fatal("expected 1 after restore")
	}
}

// ---------------------------------------------------------------------------
// vendorHandler CRUD
// ---------------------------------------------------------------------------

func TestVendorHandlerLoadAndSubmit(t *testing.T) {
	m := newTestModelWithStore(t)
	h := vendorHandler{}

	m.formData = &vendorFormData{
		Name:  "Bob's Plumbing",
		Phone: "555-1234",
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatal(err)
	}

	rows, meta, _, err := h.Load(m.store, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 vendor, got %d", len(rows))
	}

	// Edit vendor.
	editID := meta[0].ID
	m.editID = &editID
	m.formData = &vendorFormData{
		Name:  "Bob's Plumbing",
		Phone: "555-5678",
		Email: "bob@plumbing.com",
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatal(err)
	}
	m.editID = nil

	vendor, _ := m.store.GetVendor(editID)
	if vendor.Phone != "555-5678" {
		t.Errorf("phone = %q", vendor.Phone)
	}
	if vendor.Email != "bob@plumbing.com" {
		t.Errorf("email = %q", vendor.Email)
	}
}

func TestVendorHandlerDeleteReturnsError(t *testing.T) {
	h := vendorHandler{}
	if err := h.Delete(nil, 1); err == nil {
		t.Error("expected error from vendor Delete")
	}
}

func TestVendorHandlerRestoreReturnsError(t *testing.T) {
	h := vendorHandler{}
	if err := h.Restore(nil, 1); err == nil {
		t.Error("expected error from vendor Restore")
	}
}

// ---------------------------------------------------------------------------
// quoteHandler CRUD
// ---------------------------------------------------------------------------

func TestQuoteHandlerRoundTrip(t *testing.T) {
	m := newTestModelWithStore(t)
	h := quoteHandler{}

	// Need a project first.
	types, _ := m.store.ProjectTypes()
	if err := m.store.CreateProject(data.Project{
		Title:         "Bathroom Reno",
		ProjectTypeID: types[0].ID,
		Status:        data.ProjectStatusQuoted,
	}); err != nil {
		t.Fatal(err)
	}
	projects, _ := m.store.ListProjects(false)
	projID := projects[0].ID

	m.formData = &quoteFormData{
		ProjectID:  projID,
		VendorName: "Acme Contractors",
		Total:      "1,500.00",
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatalf("SubmitForm: %v", err)
	}

	rows, meta, _, err := h.Load(m.store, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 quote, got %d", len(rows))
	}
	id := meta[0].ID

	// Delete.
	if err := h.Delete(m.store, id); err != nil {
		t.Fatal(err)
	}
	rows, _, _, _ = h.Load(m.store, false)
	if len(rows) != 0 {
		t.Fatal("expected 0 after delete")
	}

	// Restore.
	if err := h.Restore(m.store, id); err != nil {
		t.Fatal(err)
	}
	rows, _, _, _ = h.Load(m.store, false)
	if len(rows) != 1 {
		t.Fatal("expected 1 after restore")
	}
}

func TestQuoteHandlerSnapshot(t *testing.T) {
	m := newTestModelWithStore(t)
	h := quoteHandler{}

	types, _ := m.store.ProjectTypes()
	if err := m.store.CreateProject(data.Project{
		Title:         "Garage Door",
		ProjectTypeID: types[0].ID,
		Status:        data.ProjectStatusQuoted,
	}); err != nil {
		t.Fatal(err)
	}
	projects, _ := m.store.ListProjects(false)

	m.formData = &quoteFormData{
		ProjectID:  projects[0].ID,
		VendorName: "QuoteCo",
		Total:      "200.00",
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatal(err)
	}
	_, meta, _, _ := h.Load(m.store, false)
	id := meta[0].ID

	entry, ok := h.Snapshot(m.store, id)
	if !ok {
		t.Fatal("Snapshot returned false")
	}
	if entry.FormKind != formQuote {
		t.Errorf("FormKind = %d", entry.FormKind)
	}
}

// ---------------------------------------------------------------------------
// serviceLogHandler CRUD
// ---------------------------------------------------------------------------

func TestServiceLogHandlerRoundTrip(t *testing.T) {
	m := newTestModelWithStore(t)
	cats, _ := m.store.MaintenanceCategories()

	// Create a maintenance item to attach logs to.
	if err := m.store.CreateMaintenance(data.MaintenanceItem{
		Name:       "Oil Furnace",
		CategoryID: cats[0].ID,
	}); err != nil {
		t.Fatal(err)
	}
	items, _ := m.store.ListMaintenance(false)
	maintID := items[0].ID

	h := serviceLogHandler{maintenanceItemID: maintID}

	m.formData = &serviceLogFormData{
		MaintenanceItemID: maintID,
		ServicedAt:        "2026-01-15",
		Cost:              "75.00",
		Notes:             "routine service",
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatal(err)
	}

	rows, meta, _, err := h.Load(m.store, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 log, got %d", len(rows))
	}
	id := meta[0].ID

	// Delete.
	if err := h.Delete(m.store, id); err != nil {
		t.Fatal(err)
	}
	rows, _, _, _ = h.Load(m.store, false)
	if len(rows) != 0 {
		t.Fatal("expected 0 after delete")
	}

	// Restore.
	if err := h.Restore(m.store, id); err != nil {
		t.Fatal(err)
	}
	rows, _, _, _ = h.Load(m.store, false)
	if len(rows) != 1 {
		t.Fatal("expected 1 after restore")
	}
}

func TestServiceLogHandlerSnapshot(t *testing.T) {
	m := newTestModelWithStore(t)
	cats, _ := m.store.MaintenanceCategories()

	if err := m.store.CreateMaintenance(data.MaintenanceItem{
		Name:       "Gutter Clean",
		CategoryID: cats[0].ID,
	}); err != nil {
		t.Fatal(err)
	}
	items, _ := m.store.ListMaintenance(false)
	maintID := items[0].ID

	h := serviceLogHandler{maintenanceItemID: maintID}

	m.formData = &serviceLogFormData{
		MaintenanceItemID: maintID,
		ServicedAt:        "2026-01-20",
	}
	if err := h.SubmitForm(m); err != nil {
		t.Fatal(err)
	}
	_, meta, _, _ := h.Load(m.store, false)
	id := meta[0].ID

	entry, ok := h.Snapshot(m.store, id)
	if !ok {
		t.Fatal("Snapshot returned false")
	}
	if entry.FormKind != formServiceLog {
		t.Errorf("FormKind = %d", entry.FormKind)
	}
	// Restore function should not error.
	if err := entry.Restore(); err != nil {
		t.Errorf("Restore: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Handler SyncFixedValues
// ---------------------------------------------------------------------------

func TestProjectHandlerSyncFixedValues(t *testing.T) {
	m := newTestModelWithStore(t)
	h := projectHandler{}
	specs := []columnSpec{
		{Title: "Type"},
		{Title: "Status"},
	}
	h.SyncFixedValues(m, specs)

	typeSpec := specs[0]
	if len(typeSpec.FixedValues) == 0 {
		t.Error("expected FixedValues for Type column")
	}
}

func TestMaintenanceHandlerSyncFixedValues(t *testing.T) {
	m := newTestModelWithStore(t)
	h := maintenanceHandler{}
	specs := []columnSpec{
		{Title: "Category"},
		{Title: "Item"},
	}
	h.SyncFixedValues(m, specs)

	catSpec := specs[0]
	if len(catSpec.FixedValues) == 0 {
		t.Error("expected FixedValues for Category column")
	}
}

// ---------------------------------------------------------------------------
// Handler with non-existent IDs
// ---------------------------------------------------------------------------

func TestProjectHandlerSnapshotNonExistent(t *testing.T) {
	m := newTestModelWithStore(t)
	h := projectHandler{}
	_, ok := h.Snapshot(m.store, 99999)
	if ok {
		t.Error("expected false for non-existent project")
	}
}

func TestQuoteHandlerSnapshotNonExistent(t *testing.T) {
	m := newTestModelWithStore(t)
	h := quoteHandler{}
	_, ok := h.Snapshot(m.store, 99999)
	if ok {
		t.Error("expected false for non-existent quote")
	}
}

func TestMaintenanceHandlerSnapshotNonExistent(t *testing.T) {
	m := newTestModelWithStore(t)
	h := maintenanceHandler{}
	_, ok := h.Snapshot(m.store, 99999)
	if ok {
		t.Error("expected false for non-existent maintenance item")
	}
}

func TestApplianceHandlerSnapshotNonExistent(t *testing.T) {
	m := newTestModelWithStore(t)
	h := applianceHandler{}
	_, ok := h.Snapshot(m.store, 99999)
	if ok {
		t.Error("expected false for non-existent appliance")
	}
}

func TestVendorHandlerSnapshotNonExistent(t *testing.T) {
	m := newTestModelWithStore(t)
	h := vendorHandler{}
	_, ok := h.Snapshot(m.store, 99999)
	if ok {
		t.Error("expected false for non-existent vendor")
	}
}

func TestServiceLogHandlerSnapshotNonExistent(t *testing.T) {
	m := newTestModelWithStore(t)
	h := serviceLogHandler{maintenanceItemID: 1}
	_, ok := h.Snapshot(m.store, 99999)
	if ok {
		t.Error("expected false for non-existent service log")
	}
}

// ---------------------------------------------------------------------------
// applianceMaintenanceHandler (detail view)
// ---------------------------------------------------------------------------

func TestApplianceMaintenanceHandlerLoad(t *testing.T) {
	m := newTestModelWithStore(t)
	cats, _ := m.store.MaintenanceCategories()

	// Create an appliance with maintenance items.
	if err := m.store.CreateAppliance(data.Appliance{Name: "HVAC"}); err != nil {
		t.Fatal(err)
	}
	apps, _ := m.store.ListAppliances(false)
	appID := apps[0].ID

	lastSrv := time.Now()
	if err := m.store.CreateMaintenance(data.MaintenanceItem{
		Name:           "Replace Belt",
		CategoryID:     cats[0].ID,
		ApplianceID:    &appID,
		LastServicedAt: &lastSrv,
		IntervalMonths: 12,
	}); err != nil {
		t.Fatal(err)
	}

	h := applianceMaintenanceHandler{applianceID: appID}
	rows, meta, _, err := h.Load(m.store, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 maintenance item for appliance, got %d", len(rows))
	}
	if meta[0].ID == 0 {
		t.Error("expected non-zero ID")
	}
}
