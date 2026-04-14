// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/micasa-dev/micasa/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Vendor ---

func TestVendorCRUD(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := vendorEntityDef()

	// ADD
	created, err := def.decodeAndCreate(store,
		json.RawMessage(`{"name":"Acme Plumbing","phone":"5551234567"}`))
	require.NoError(t, err)
	m := def.toMap(created)
	assert.Equal(t, "Acme Plumbing", m["name"])
	assert.Equal(t, "5551234567", m["phone"])
	id, _ := m["id"].(string)
	require.NotEmpty(t, id)

	// LIST
	items, err := def.list(store, false)
	require.NoError(t, err)
	require.Len(t, items, 1)

	// GET
	got, err := def.get(store, id)
	require.NoError(t, err)
	gm := def.toMap(got)
	assert.Equal(t, "Acme Plumbing", gm["name"])
	assert.Equal(t, "5551234567", gm["phone"])

	// EDIT (partial)
	edited, err := def.decodeAndUpdate(store, id,
		json.RawMessage(`{"phone":"5559876543"}`))
	require.NoError(t, err)
	em := def.toMap(edited)
	assert.Equal(t, "Acme Plumbing", em["name"])
	assert.Equal(t, "5559876543", em["phone"])

	// DELETE
	require.NoError(t, def.del(store, id))
	afterDel, err := def.list(store, false)
	require.NoError(t, err)
	assert.Empty(t, afterDel)

	// RESTORE
	require.NoError(t, def.restore(store, id))
	afterRestore, err := def.list(store, false)
	require.NoError(t, err)
	require.Len(t, afterRestore, 1)
}

func TestVendorAddMissingName(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := vendorEntityDef()
	_, err := def.decodeAndCreate(store, json.RawMessage(`{"phone":"123"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestVendorDeleteWithDeps(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)

	vendor, err := store.FindOrCreateVendor(data.Vendor{Name: "DepVendor"})
	require.NoError(t, err)

	ptypes, err := store.ProjectTypes()
	require.NoError(t, err)
	require.NotEmpty(t, ptypes)

	p := &data.Project{
		Title:         "DepProj",
		ProjectTypeID: ptypes[0].ID,
		Status:        data.ProjectStatusPlanned,
	}
	require.NoError(t, store.CreateProject(p))

	q := &data.Quote{ProjectID: p.ID, TotalCents: 100}
	require.NoError(t, store.CreateQuote(q, vendor))

	def := vendorEntityDef()
	err = def.del(store, vendor.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "active quote")
}

func TestVendorListTable(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	require.NoError(t, store.CreateVendor(&data.Vendor{Name: "TableTest"}))

	def := vendorEntityDef()
	items, err := def.list(store, false)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, writeTable(&buf, def.tableHeader, items, def.cols))
	assert.Contains(t, buf.String(), "TableTest")
	assert.Contains(t, buf.String(), "NAME")
}

func TestVendorListDeleted(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := vendorEntityDef()

	created, err := def.decodeAndCreate(store, json.RawMessage(`{"name":"Ghost"}`))
	require.NoError(t, err)
	id, _ := def.toMap(created)["id"].(string)

	require.NoError(t, def.del(store, id))

	live, err := def.list(store, false)
	require.NoError(t, err)
	assert.Empty(t, live)

	all, err := def.list(store, true)
	require.NoError(t, err)
	require.Len(t, all, 1)
}

func TestVendorJSONOutput(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := vendorEntityDef()

	created, err := def.decodeAndCreate(store, json.RawMessage(`{"name":"JSONTest"}`))
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, writeJSON(&buf, []data.Vendor{created}, def.toMap))

	var result []map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 1)
	assert.Equal(t, "JSONTest", result[0]["name"])
}

func TestVendorCobraWiring(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)
	out, err := executeCLI("vendor", "list", dbPath)
	require.NoError(t, err)
	assert.Equal(t, "[]\n", out)
}

// --- Project ---

func TestProjectCRUD(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := projectEntityDef()

	ptypes, err := store.ProjectTypes()
	require.NoError(t, err)
	require.NotEmpty(t, ptypes)

	// ADD
	created, err := def.decodeAndCreate(store, json.RawMessage(
		`{"title":"Fence Build","project_type_id":"`+ptypes[0].ID+`","budget_cents":500000}`))
	require.NoError(t, err)
	m := def.toMap(created)
	assert.Equal(t, "Fence Build", m["title"])
	assert.Equal(t, "planned", m["status"])
	id, _ := m["id"].(string)

	// EDIT (status only)
	edited, err := def.decodeAndUpdate(store, id,
		json.RawMessage(`{"status":"completed"}`))
	require.NoError(t, err)
	em := def.toMap(edited)
	assert.Equal(t, "completed", em["status"])
	assert.Equal(t, "Fence Build", em["title"])

	// EDIT (clear nullable)
	cleared, err := def.decodeAndUpdate(store, id,
		json.RawMessage(`{"budget_cents":null}`))
	require.NoError(t, err)
	cm := def.toMap(cleared)
	assert.Nil(t, cm["budget_cents"])

	// DELETE + RESTORE
	require.NoError(t, def.del(store, id))
	require.NoError(t, def.restore(store, id))

	items, err := def.list(store, false)
	require.NoError(t, err)
	require.Len(t, items, 1)
}

func TestProjectAddMissingFields(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := projectEntityDef()

	_, err := def.decodeAndCreate(store, json.RawMessage(`{"title":"NoType"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project_type_id is required")

	_, err = def.decodeAndCreate(store, json.RawMessage(`{"project_type_id":"abc"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title is required")
}

func TestProjectCobraWiring(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)
	out, err := executeCLI("project", "list", dbPath)
	require.NoError(t, err)
	assert.Equal(t, "[]\n", out)
}

// --- Appliance ---

func TestApplianceCRUD(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := applianceEntityDef()

	cost := int64(120000)
	created, err := def.decodeAndCreate(store, json.RawMessage(
		`{"name":"Dishwasher","brand":"Bosch","cost_cents":120000}`))
	require.NoError(t, err)
	m := def.toMap(created)
	assert.Equal(t, "Dishwasher", m["name"])
	assert.Equal(t, "Bosch", m["brand"])
	assert.Equal(t, &cost, m["cost_cents"])
	id, _ := m["id"].(string)

	edited, err := def.decodeAndUpdate(store, id,
		json.RawMessage(`{"brand":"Samsung"}`))
	require.NoError(t, err)
	assert.Equal(t, "Samsung", def.toMap(edited)["brand"])
	assert.Equal(t, "Dishwasher", def.toMap(edited)["name"])

	require.NoError(t, def.del(store, id))
	require.NoError(t, def.restore(store, id))
}

func TestApplianceCobraWiring(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)
	out, err := executeCLI("appliance", "list", dbPath)
	require.NoError(t, err)
	assert.Equal(t, "[]\n", out)
}

// --- Incident ---

func TestIncidentCRUD(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := incidentEntityDef()

	created, err := def.decodeAndCreate(store,
		json.RawMessage(`{"title":"Pipe burst","location":"Basement"}`))
	require.NoError(t, err)
	m := def.toMap(created)
	assert.Equal(t, "Pipe burst", m["title"])
	assert.Equal(t, "open", m["status"])
	assert.Equal(t, "soon", m["severity"])
	id, _ := m["id"].(string)

	edited, err := def.decodeAndUpdate(store, id,
		json.RawMessage(`{"status":"in_progress"}`))
	require.NoError(t, err)
	assert.Equal(t, "in_progress", def.toMap(edited)["status"])

	require.NoError(t, def.del(store, id))
	require.NoError(t, def.restore(store, id))

	got, err := def.get(store, id)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", def.toMap(got)["status"])
}

func TestIncidentCobraWiring(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)
	out, err := executeCLI("incident", "list", dbPath)
	require.NoError(t, err)
	assert.Equal(t, "[]\n", out)
}

// --- Quote ---

func TestQuoteCRUD(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := quoteEntityDef()

	ptypes, err := store.ProjectTypes()
	require.NoError(t, err)
	p := &data.Project{
		Title:         "QProj",
		ProjectTypeID: ptypes[0].ID,
		Status:        data.ProjectStatusPlanned,
	}
	require.NoError(t, store.CreateProject(p))

	// ADD with vendor_name
	created, err := def.decodeAndCreate(store, json.RawMessage(
		`{"project_id":"`+p.ID+`","vendor_name":"QuoteVendor","total_cents":750000}`))
	require.NoError(t, err)
	m := def.toMap(created)
	assert.Equal(t, "QuoteVendor", m["vendor"])
	id, _ := m["id"].(string)

	// EDIT preserving vendor (omit both)
	edited, err := def.decodeAndUpdate(store, id,
		json.RawMessage(`{"total_cents":800000}`))
	require.NoError(t, err)
	em := def.toMap(edited)
	assert.Equal(t, "QuoteVendor", em["vendor"])
	assert.Equal(t, int64(800000), em["total_cents"])

	// EDIT changing vendor via vendor_name
	edited2, err := def.decodeAndUpdate(store, id,
		json.RawMessage(`{"vendor_name":"NewVendor"}`))
	require.NoError(t, err)
	assert.Equal(t, "NewVendor", def.toMap(edited2)["vendor"])

	require.NoError(t, def.del(store, id))
	require.NoError(t, def.restore(store, id))
}

func TestQuoteAddMissingFields(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := quoteEntityDef()

	_, err := def.decodeAndCreate(store, json.RawMessage(`{"total_cents":100}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project_id is required")

	ptypes, err := store.ProjectTypes()
	require.NoError(t, err)
	p := &data.Project{Title: "P", ProjectTypeID: ptypes[0].ID, Status: data.ProjectStatusPlanned}
	require.NoError(t, store.CreateProject(p))

	_, err = def.decodeAndCreate(store,
		json.RawMessage(`{"project_id":"`+p.ID+`","total_cents":100}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vendor_id or vendor_name is required")
}

func TestQuoteCobraWiring(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)
	out, err := executeCLI("quote", "list", dbPath)
	require.NoError(t, err)
	assert.Equal(t, "[]\n", out)
}

// --- Maintenance ---

func TestMaintenanceCRUD(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := maintenanceEntityDef()

	cats, err := store.MaintenanceCategories()
	require.NoError(t, err)
	require.NotEmpty(t, cats)

	created, err := def.decodeAndCreate(store, json.RawMessage(
		`{"name":"Change HVAC Filter","category_id":"`+cats[0].ID+`","interval_months":3}`))
	require.NoError(t, err)
	m := def.toMap(created)
	assert.Equal(t, "Change HVAC Filter", m["name"])
	id, _ := m["id"].(string)

	edited, err := def.decodeAndUpdate(store, id,
		json.RawMessage(`{"interval_months":6}`))
	require.NoError(t, err)
	assert.Equal(t, 6, def.toMap(edited)["interval_months"])

	require.NoError(t, def.del(store, id))
	require.NoError(t, def.restore(store, id))
}

func TestMaintenanceAddMissingCategory(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := maintenanceEntityDef()
	_, err := def.decodeAndCreate(store, json.RawMessage(`{"name":"NoCategory"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "category_id is required")
}

func TestMaintenanceCobraWiring(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)
	out, err := executeCLI("maintenance", "list", dbPath)
	require.NoError(t, err)
	assert.Equal(t, "[]\n", out)
}

// --- Service Log ---

func TestServiceLogCRUD(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := serviceLogEntityDef()

	cats, err := store.MaintenanceCategories()
	require.NoError(t, err)
	m := &data.MaintenanceItem{Name: "SLTest", CategoryID: cats[0].ID}
	require.NoError(t, store.CreateMaintenance(m))

	now := time.Now().Truncate(24 * time.Hour).Format(data.DateLayout)

	// ADD with vendor_name
	created, err := def.decodeAndCreate(store, json.RawMessage(
		`{"maintenance_item_id":"`+m.ID+`","serviced_at":"`+now+`","vendor_name":"SLVendor","cost_cents":5000}`,
	))
	require.NoError(t, err)
	cm := def.toMap(created)
	assert.Equal(t, "SLVendor", cm["vendor"])
	id, _ := cm["id"].(string)

	// EDIT preserving vendor
	edited, err := def.decodeAndUpdate(store, id,
		json.RawMessage(`{"cost_cents":6000}`))
	require.NoError(t, err)
	assert.Equal(t, "SLVendor", def.toMap(edited)["vendor"])

	require.NoError(t, def.del(store, id))
	require.NoError(t, def.restore(store, id))
}

func TestServiceLogAddMissingFields(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := serviceLogEntityDef()

	_, err := def.decodeAndCreate(store, json.RawMessage(`{"serviced_at":"2026-01-01"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maintenance_item_id is required")
}

func TestServiceLogCobraWiring(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)
	out, err := executeCLI("service-log", "list", dbPath)
	require.NoError(t, err)
	assert.Equal(t, "[]\n", out)
}

// --- Document ---

func TestDocumentCRUD(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	require.NoError(t, store.SetMaxDocumentSize(50*1024*1024))
	def := documentEntityDef()

	vendor, err := store.FindOrCreateVendor(data.Vendor{Name: "DocVendor"})
	require.NoError(t, err)

	// ADD metadata-only
	created, err := documentCreate(store,
		json.RawMessage(`{"title":"Test Doc","entity_kind":"vendor","entity_id":"`+vendor.ID+`"}`),
		"")
	require.NoError(t, err)
	m := def.toMap(created)
	assert.Equal(t, "Test Doc", m["title"])
	id, _ := m["id"].(string)

	// EDIT title
	edited, err := def.decodeAndUpdate(store, id,
		json.RawMessage(`{"title":"Updated Doc"}`))
	require.NoError(t, err)
	assert.Equal(t, "Updated Doc", def.toMap(edited)["title"])

	// DELETE + RESTORE
	require.NoError(t, def.del(store, id))
	require.NoError(t, def.restore(store, id))
}

func TestDocumentAddWithFile(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	require.NoError(t, store.SetMaxDocumentSize(50*1024*1024))

	vendor, err := store.FindOrCreateVendor(data.Vendor{Name: "FileVendor"})
	require.NoError(t, err)

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("hello world"), 0o600))

	created, err := documentCreate(store,
		json.RawMessage(`{"entity_kind":"vendor","entity_id":"`+vendor.ID+`"}`),
		tmpFile)
	require.NoError(t, err)

	def := documentEntityDef()
	m := def.toMap(created)
	assert.Equal(t, "Test", m["title"])
	assert.Equal(t, "test.txt", m["file_name"])
	assert.NotEmpty(t, m["sha256"])
}

func TestDocumentAddMissingFields(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	require.NoError(t, store.SetMaxDocumentSize(50*1024*1024))

	_, err := documentCreate(store, json.RawMessage(`{"title":"NoLink"}`), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "entity_kind is required")
}

func TestDocumentCobraWiring(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)
	out, err := executeCLI("document", "list", dbPath)
	require.NoError(t, err)
	assert.Equal(t, "[]\n", out)
}

// --- House ---

func TestHouseCRUD(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)

	// GET empty
	_, err := store.HouseProfile()
	require.Error(t, err) // not found

	// ADD
	created, err := houseCreate(store,
		json.RawMessage(`{"nickname":"Test House","city":"Springfield","state":"IL"}`))
	require.NoError(t, err)
	assert.Equal(t, "Test House", created.Nickname)
	assert.Equal(t, "Springfield", created.City)

	// GET populated
	h, err := store.HouseProfile()
	require.NoError(t, err)
	assert.Equal(t, "Test House", h.Nickname)

	// EDIT partial
	updated, err := houseUpdate(store,
		json.RawMessage(`{"nickname":"Updated House"}`))
	require.NoError(t, err)
	assert.Equal(t, "Updated House", updated.Nickname)
	assert.Equal(t, "Springfield", updated.City)

	// ADD duplicate
	_, err = houseCreate(store,
		json.RawMessage(`{"nickname":"Duplicate"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestHouseCobraWiring(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)
	out, err := executeCLI("house", "get", dbPath)
	require.NoError(t, err)
	assert.Equal(t, "{}\n", out)
}

// --- Lookup Tables ---

func TestProjectTypeList(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := projectTypeEntityDef()

	items, err := def.list(store, false)
	require.NoError(t, err)
	assert.NotEmpty(t, items)

	var buf bytes.Buffer
	require.NoError(t, writeJSON(&buf, items, def.toMap))
	var result []map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.NotEmpty(t, result)
	assert.NotEmpty(t, result[0]["name"])
}

func TestProjectTypeCobraWiring(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)
	out, err := executeCLI("project-type", "list", dbPath)
	require.NoError(t, err)
	assert.Contains(t, out, "name")
}

func TestMaintenanceCategoryList(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := maintenanceCategoryEntityDef()

	items, err := def.list(store, false)
	require.NoError(t, err)
	assert.NotEmpty(t, items)
}

func TestMaintenanceCategoryCobraWiring(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)
	out, err := executeCLI("maintenance-category", "list", dbPath)
	require.NoError(t, err)
	assert.Contains(t, out, "name")
}

// --- Input Validation ---

func TestReadInputDataMutualExclusion(t *testing.T) {
	t.Parallel()

	root := newRootCmd()
	addCmd := buildAddCmd(vendorEntityDef())
	root.AddCommand(addCmd)

	// Set both flags
	require.NoError(t, addCmd.Flags().Set("data", `{"name":"x"}`))
	require.NoError(t, addCmd.Flags().Set("data-file", "some.json"))

	_, err := readInputData(addCmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestReadInputDataFromFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "input.json")
	require.NoError(t, os.WriteFile(fp, []byte(`{"name":"FromFile"}`), 0o600))

	root := newRootCmd()
	addCmd := buildAddCmd(vendorEntityDef())
	root.AddCommand(addCmd)
	require.NoError(t, addCmd.Flags().Set("data-file", fp))

	raw, err := readInputData(addCmd)
	require.NoError(t, err)
	assert.Contains(t, string(raw), "FromFile")
}

func TestInvalidJSON(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := vendorEntityDef()

	_, err := def.decodeAndCreate(store, json.RawMessage(`{invalid`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestGetNonexistentID(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := vendorEntityDef()

	_, err := def.get(store, "nonexistent")
	require.Error(t, err)
}

func TestDeleteNonexistentID(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := vendorEntityDef()

	err := def.del(store, "nonexistent")
	require.Error(t, err)
}

func TestRestoreNonexistentID(t *testing.T) {
	t.Parallel()
	store := newTestStoreWithMigration(t)
	def := vendorEntityDef()

	// Restore of nonexistent ID is a no-op (GORM Update silently
	// succeeds with zero rows affected).
	err := def.restore(store, "nonexistent")
	require.NoError(t, err)
}

// --- Show Deprecation ---

func TestShowDeprecation(t *testing.T) {
	t.Parallel()
	dbPath := createTestDB(t)

	// Deprecated per-entity commands still work
	_, err := executeCLI("show", "vendors", dbPath)
	require.NoError(t, err)

	// "show all" is not deprecated -- still works
	_, err = executeCLI("show", "all", dbPath)
	require.NoError(t, err)
}
