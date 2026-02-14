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
	require.NoError(t, h.SubmitForm(m))

	// Load should return the project.
	rows, meta, cells, err := h.Load(m.store, false)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Len(t, meta, 1)
	require.Len(t, cells, 1)
	id := meta[0].ID

	// Delete.
	require.NoError(t, h.Delete(m.store, id))

	// Load without deleted should be empty.
	rows, _, _, err = h.Load(m.store, false)
	require.NoError(t, err)
	assert.Empty(t, rows)

	// Load with deleted should show it.
	rows, _, _, err = h.Load(m.store, true)
	require.NoError(t, err)
	assert.Len(t, rows, 1)

	// Restore.
	require.NoError(t, h.Restore(m.store, id))
	rows, _, _, _ = h.Load(m.store, false)
	assert.Len(t, rows, 1)
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
	require.NoError(t, h.SubmitForm(m))
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
	require.NoError(t, h.SubmitForm(m))
	m.editID = nil

	project, err := m.store.GetProject(id)
	require.NoError(t, err)
	assert.Equal(t, "Paint Fence Red", project.Title)
	assert.Equal(t, data.ProjectStatusInProgress, project.Status)
	require.NotNil(t, project.BudgetCents)
	assert.Equal(t, int64(50000), *project.BudgetCents)
}

func TestProjectHandlerSnapshot(t *testing.T) {
	m := newTestModelWithStore(t)
	h := projectHandler{}

	m.formData = &projectFormData{
		Title:         "Roof Repair",
		ProjectTypeID: m.projectTypes[0].ID,
		Status:        data.ProjectStatusPlanned,
	}
	require.NoError(t, h.SubmitForm(m))
	_, meta, _, _ := h.Load(m.store, false)
	id := meta[0].ID

	entry, ok := h.Snapshot(m.store, id)
	require.True(t, ok)
	assert.Equal(t, formProject, entry.FormKind)
	assert.Equal(t, id, entry.EntityID)
}

func TestProjectTabStatusFiltersRows(t *testing.T) {
	m := newTestModelWithStore(t)
	types, _ := m.store.ProjectTypes()

	for _, p := range []data.Project{
		{
			Title:         "Kitchen Plan",
			ProjectTypeID: types[0].ID,
			Status:        data.ProjectStatusPlanned,
		},
		{
			Title:         "Fence Done",
			ProjectTypeID: types[0].ID,
			Status:        data.ProjectStatusCompleted,
		},
		{
			Title:         "Basement Work",
			ProjectTypeID: types[0].ID,
			Status:        data.ProjectStatusInProgress,
		},
		{
			Title:         "Old Patio Idea",
			ProjectTypeID: types[0].ID,
			Status:        data.ProjectStatusAbandoned,
		},
	} {
		require.NoError(t, m.store.CreateProject(p), "CreateProject(%q)", p.Title)
	}

	m.active = tabIndex(tabProjects)
	require.NoError(t, m.reloadActiveTab())
	tab := m.activeTab()
	require.NotNil(t, tab, "expected active projects tab")
	require.Len(t, tab.Rows, 4, "rows before filtering")

	tab.HideCompleted = true
	require.NoError(t, m.reloadActiveTab(), "reloadActiveTab with HideCompleted")
	assert.Len(t, tab.Rows, 3, "rows with completed hidden")
	for i, cells := range tab.CellRows {
		if len(cells) > 3 {
			assert.NotEqual(
				t,
				data.ProjectStatusCompleted,
				cells[3].Value,
				"row %d still has completed status",
				i,
			)
		}
	}

	tab.HideCompleted = false
	tab.HideAbandoned = true
	require.NoError(t, m.reloadActiveTab(), "reloadActiveTab with HideAbandoned")
	assert.Len(t, tab.Rows, 3, "rows with abandoned hidden")
	for i, cells := range tab.CellRows {
		if len(cells) > 3 {
			assert.NotEqual(
				t,
				data.ProjectStatusAbandoned,
				cells[3].Value,
				"row %d still has abandoned status",
				i,
			)
		}
	}

	tab.HideAbandoned = false
	tab.HideCompleted = true
	tab.HideAbandoned = true
	require.NoError(t, m.reloadActiveTab(), "reloadActiveTab with settled filters")
	assert.Len(t, tab.Rows, 2, "rows with settled hidden")
	for i, cells := range tab.CellRows {
		if len(cells) <= 3 {
			continue
		}
		status := cells[3].Value
		assert.NotEqual(
			t,
			data.ProjectStatusCompleted,
			status,
			"row %d still has settled status",
			i,
		)
		assert.NotEqual(
			t,
			data.ProjectStatusAbandoned,
			status,
			"row %d still has settled status",
			i,
		)
	}

	tab.HideCompleted = false
	tab.HideAbandoned = false
	require.NoError(t, m.reloadActiveTab(), "reloadActiveTab after clearing filters")
	assert.Len(t, tab.Rows, 4, "rows after showing all projects")
}

func TestProjectStatusFilterToggleKeysReloadRows(t *testing.T) {
	m := newTestModelWithStore(t)
	types, _ := m.store.ProjectTypes()

	require.NoError(t, m.store.CreateProject(data.Project{
		Title:         "Done Project",
		ProjectTypeID: types[0].ID,
		Status:        data.ProjectStatusCompleted,
	}), "CreateProject completed")
	require.NoError(t, m.store.CreateProject(data.Project{
		Title:         "Live Project",
		ProjectTypeID: types[0].ID,
		Status:        data.ProjectStatusInProgress,
	}), "CreateProject in-progress")
	require.NoError(t, m.store.CreateProject(data.Project{
		Title:         "Abandoned Project",
		ProjectTypeID: types[0].ID,
		Status:        data.ProjectStatusAbandoned,
	}), "CreateProject abandoned")

	m.active = tabIndex(tabProjects)
	require.NoError(t, m.reloadActiveTab())
	require.Len(t, m.activeTab().Rows, 3, "rows before toggles")

	sendKey(m, "t")
	assert.Len(t, m.activeTab().Rows, 1, "rows after hiding settled")
	assert.True(t, m.activeTab().HideCompleted, "settled toggle should enable completed")
	assert.True(t, m.activeTab().HideAbandoned, "settled toggle should enable abandoned")

	sendKey(m, "t")
	assert.Len(t, m.activeTab().Rows, 3, "rows after showing settled")
	assert.False(t, m.activeTab().HideCompleted, "settled toggle should disable completed")
	assert.False(t, m.activeTab().HideAbandoned, "settled toggle should disable abandoned")
}

// ---------------------------------------------------------------------------
// applianceHandler CRUD
// ---------------------------------------------------------------------------

func TestApplianceHandlerLoadDeleteRestoreRoundTrip(t *testing.T) {
	m := newTestModelWithStore(t)
	h := applianceHandler{}

	m.formData = &applianceFormData{Name: "Washer"}
	require.NoError(t, h.SubmitForm(m))

	rows, meta, _, err := h.Load(m.store, false)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	id := meta[0].ID

	require.NoError(t, h.Delete(m.store, id))
	rows, _, _, _ = h.Load(m.store, false)
	assert.Empty(t, rows)

	require.NoError(t, h.Restore(m.store, id))
	rows, _, _, _ = h.Load(m.store, false)
	assert.Len(t, rows, 1)
}

func TestApplianceHandlerEditRoundTrip(t *testing.T) {
	m := newTestModelWithStore(t)
	h := applianceHandler{}

	m.formData = &applianceFormData{Name: "Dryer"}
	require.NoError(t, h.SubmitForm(m))
	_, meta, _, _ := h.Load(m.store, false)
	id := meta[0].ID

	editID := id
	m.editID = &editID
	m.formData = &applianceFormData{
		Name:  "Dryer",
		Brand: "LG",
		Cost:  "800.00",
	}
	require.NoError(t, h.SubmitForm(m))
	m.editID = nil

	app, _ := m.store.GetAppliance(id)
	assert.Equal(t, "LG", app.Brand)
	require.NotNil(t, app.CostCents)
	assert.Equal(t, int64(80000), *app.CostCents)
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
	require.NoError(t, h.SubmitForm(m))

	rows, meta, _, err := h.Load(m.store, false)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	id := meta[0].ID

	require.NoError(t, h.Delete(m.store, id))
	rows, _, _, _ = h.Load(m.store, false)
	assert.Empty(t, rows)

	require.NoError(t, h.Restore(m.store, id))
	rows, _, _, _ = h.Load(m.store, false)
	assert.Len(t, rows, 1)
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
	require.NoError(t, h.SubmitForm(m))

	rows, meta, _, err := h.Load(m.store, false)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	// Edit vendor.
	editID := meta[0].ID
	m.editID = &editID
	m.formData = &vendorFormData{
		Name:  "Bob's Plumbing",
		Phone: "555-5678",
		Email: "bob@plumbing.com",
	}
	require.NoError(t, h.SubmitForm(m))
	m.editID = nil

	vendor, _ := m.store.GetVendor(editID)
	assert.Equal(t, "555-5678", vendor.Phone)
	assert.Equal(t, "bob@plumbing.com", vendor.Email)
}

// Vendor delete/restore tests moved to vendor_test.go (TestVendorHandlerDeleteRestore)
// -- they now require a real store.

// ---------------------------------------------------------------------------
// quoteHandler CRUD
// ---------------------------------------------------------------------------

func TestQuoteHandlerRoundTrip(t *testing.T) {
	m := newTestModelWithStore(t)
	h := quoteHandler{}

	// Need a project first.
	types, _ := m.store.ProjectTypes()
	require.NoError(t, m.store.CreateProject(data.Project{
		Title:         "Bathroom Reno",
		ProjectTypeID: types[0].ID,
		Status:        data.ProjectStatusQuoted,
	}))
	projects, _ := m.store.ListProjects(false)
	projID := projects[0].ID

	m.formData = &quoteFormData{
		ProjectID:  projID,
		VendorName: "Acme Contractors",
		Total:      "1,500.00",
	}
	require.NoError(t, h.SubmitForm(m))

	rows, meta, _, err := h.Load(m.store, false)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	id := meta[0].ID

	// Delete.
	require.NoError(t, h.Delete(m.store, id))
	rows, _, _, _ = h.Load(m.store, false)
	assert.Empty(t, rows)

	// Restore.
	require.NoError(t, h.Restore(m.store, id))
	rows, _, _, _ = h.Load(m.store, false)
	assert.Len(t, rows, 1)
}

func TestQuoteHandlerSnapshot(t *testing.T) {
	m := newTestModelWithStore(t)
	h := quoteHandler{}

	types, _ := m.store.ProjectTypes()
	require.NoError(t, m.store.CreateProject(data.Project{
		Title:         "Garage Door",
		ProjectTypeID: types[0].ID,
		Status:        data.ProjectStatusQuoted,
	}))
	projects, _ := m.store.ListProjects(false)

	m.formData = &quoteFormData{
		ProjectID:  projects[0].ID,
		VendorName: "QuoteCo",
		Total:      "200.00",
	}
	require.NoError(t, h.SubmitForm(m))
	_, meta, _, _ := h.Load(m.store, false)
	id := meta[0].ID

	entry, ok := h.Snapshot(m.store, id)
	require.True(t, ok)
	assert.Equal(t, formQuote, entry.FormKind)
}

// ---------------------------------------------------------------------------
// serviceLogHandler CRUD
// ---------------------------------------------------------------------------

func TestServiceLogHandlerRoundTrip(t *testing.T) {
	m := newTestModelWithStore(t)
	cats, _ := m.store.MaintenanceCategories()

	// Create a maintenance item to attach logs to.
	require.NoError(t, m.store.CreateMaintenance(data.MaintenanceItem{
		Name:       "Oil Furnace",
		CategoryID: cats[0].ID,
	}))
	items, _ := m.store.ListMaintenance(false)
	maintID := items[0].ID

	h := serviceLogHandler{maintenanceItemID: maintID}

	m.formData = &serviceLogFormData{
		MaintenanceItemID: maintID,
		ServicedAt:        "2026-01-15",
		Cost:              "75.00",
		Notes:             "routine service",
	}
	require.NoError(t, h.SubmitForm(m))

	rows, meta, _, err := h.Load(m.store, false)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	id := meta[0].ID

	// Delete.
	require.NoError(t, h.Delete(m.store, id))
	rows, _, _, _ = h.Load(m.store, false)
	assert.Empty(t, rows)

	// Restore.
	require.NoError(t, h.Restore(m.store, id))
	rows, _, _, _ = h.Load(m.store, false)
	assert.Len(t, rows, 1)
}

func TestServiceLogHandlerSnapshot(t *testing.T) {
	m := newTestModelWithStore(t)
	cats, _ := m.store.MaintenanceCategories()

	require.NoError(t, m.store.CreateMaintenance(data.MaintenanceItem{
		Name:       "Gutter Clean",
		CategoryID: cats[0].ID,
	}))
	items, _ := m.store.ListMaintenance(false)
	maintID := items[0].ID

	h := serviceLogHandler{maintenanceItemID: maintID}

	m.formData = &serviceLogFormData{
		MaintenanceItemID: maintID,
		ServicedAt:        "2026-01-20",
	}
	require.NoError(t, h.SubmitForm(m))
	_, meta, _, _ := h.Load(m.store, false)
	id := meta[0].ID

	entry, ok := h.Snapshot(m.store, id)
	require.True(t, ok)
	assert.Equal(t, formServiceLog, entry.FormKind)
	// Restore function should not error.
	assert.NoError(t, entry.Restore())
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

	assert.NotEmpty(t, specs[0].FixedValues, "expected FixedValues for Type column")
}

func TestMaintenanceHandlerSyncFixedValues(t *testing.T) {
	m := newTestModelWithStore(t)
	h := maintenanceHandler{}
	specs := []columnSpec{
		{Title: "Category"},
		{Title: "Item"},
	}
	h.SyncFixedValues(m, specs)

	assert.NotEmpty(t, specs[0].FixedValues, "expected FixedValues for Category column")
}

// ---------------------------------------------------------------------------
// Handler with non-existent IDs
// ---------------------------------------------------------------------------

func TestHandlerSnapshotNonExistent(t *testing.T) {
	m := newTestModelWithStore(t)
	handlers := []struct {
		name string
		h    TabHandler
	}{
		{"project", projectHandler{}},
		{"quote", quoteHandler{}},
		{"maintenance", maintenanceHandler{}},
		{"appliance", applianceHandler{}},
		{"vendor", vendorHandler{}},
		{"serviceLog", serviceLogHandler{maintenanceItemID: 1}},
	}
	for _, tc := range handlers {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tc.h.Snapshot(m.store, 99999)
			assert.False(t, ok)
		})
	}
}

// ---------------------------------------------------------------------------
// applianceMaintenanceHandler (detail view)
// ---------------------------------------------------------------------------

func TestApplianceMaintenanceHandlerLoad(t *testing.T) {
	m := newTestModelWithStore(t)
	cats, _ := m.store.MaintenanceCategories()

	// Create an appliance with maintenance items.
	require.NoError(t, m.store.CreateAppliance(data.Appliance{Name: "HVAC"}))
	apps, _ := m.store.ListAppliances(false)
	appID := apps[0].ID

	lastSrv := time.Now()
	require.NoError(t, m.store.CreateMaintenance(data.MaintenanceItem{
		Name:           "Replace Belt",
		CategoryID:     cats[0].ID,
		ApplianceID:    &appID,
		LastServicedAt: &lastSrv,
		IntervalMonths: 12,
	}))

	h := applianceMaintenanceHandler{applianceID: appID}
	rows, meta, _, err := h.Load(m.store, false)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.NotZero(t, meta[0].ID)
}
