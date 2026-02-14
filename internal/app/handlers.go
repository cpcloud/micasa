// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/cpcloud/micasa/internal/data"
)

// TabHandler encapsulates entity-specific operations for a tab, eliminating
// TabKind/FormKind switch dispatch scattered across the codebase. Each entity
// type (projects, quotes, maintenance, appliances) implements this interface.
type TabHandler interface {
	// FormKind returns the FormKind that identifies this entity in forms and
	// undo entries.
	FormKind() FormKind

	// Load fetches entities and converts them to table rows.
	Load(store *data.Store, showDeleted bool) ([]table.Row, []rowMeta, [][]cell, error)

	// Delete soft-deletes the entity with the given ID.
	Delete(store *data.Store, id uint) error

	// Restore reverses a soft-delete.
	Restore(store *data.Store, id uint) error

	// StartAddForm opens a "new entity" form on the model.
	StartAddForm(m *Model) error

	// StartEditForm opens an "edit entity" form for the given ID.
	StartEditForm(m *Model, id uint) error

	// InlineEdit opens a single-field editor for the given column.
	InlineEdit(m *Model, id uint, col int) error

	// SubmitForm persists the current form data (create or update).
	SubmitForm(m *Model) error

	// Snapshot captures the current DB state of an entity for undo/redo.
	Snapshot(store *data.Store, id uint) (undoEntry, bool)

	// SyncFixedValues updates column specs with values from dynamic lookup
	// tables so column widths stay stable.
	SyncFixedValues(m *Model, specs []columnSpec)
}

// handlerForFormKind finds the tab handler that owns the given FormKind.
// Checks both main tabs and the detail tab (if active).
// Returns nil for formHouse (no tab) or unknown kinds.
func (m *Model) handlerForFormKind(kind FormKind) TabHandler {
	// Check the detail tab first since it may shadow a main tab's form kind.
	if dc := m.detail(); dc != nil && dc.Tab.Handler != nil &&
		dc.Tab.Handler.FormKind() == kind {
		return dc.Tab.Handler
	}
	for i := range m.tabs {
		if m.tabs[i].Handler != nil && m.tabs[i].Handler.FormKind() == kind {
			return m.tabs[i].Handler
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// projectHandler
// ---------------------------------------------------------------------------

type projectHandler struct{}

func (projectHandler) FormKind() FormKind { return formProject }

func (projectHandler) Load(
	store *data.Store,
	showDeleted bool,
) ([]table.Row, []rowMeta, [][]cell, error) {
	projects, err := store.ListProjects(showDeleted)
	if err != nil {
		return nil, nil, nil, err
	}
	ids := make([]uint, len(projects))
	for i, p := range projects {
		ids[i] = p.ID
	}
	quoteCounts, err := store.CountQuotesByProject(ids)
	if err != nil {
		quoteCounts = map[uint]int{}
	}
	rows, meta, cellRows := projectRows(projects, quoteCounts)
	return rows, meta, cellRows, nil
}

func (projectHandler) Delete(store *data.Store, id uint) error {
	return store.DeleteProject(id)
}

func (projectHandler) Restore(store *data.Store, id uint) error {
	return store.RestoreProject(id)
}

func (projectHandler) StartAddForm(m *Model) error {
	m.startProjectForm()
	return nil
}

func (projectHandler) StartEditForm(m *Model, id uint) error {
	return m.startEditProjectForm(id)
}

func (projectHandler) InlineEdit(m *Model, id uint, col int) error {
	return m.inlineEditProject(id, col)
}

func (projectHandler) SubmitForm(m *Model) error {
	return m.submitProjectForm()
}

func (projectHandler) Snapshot(store *data.Store, id uint) (undoEntry, bool) {
	project, err := store.GetProject(id)
	if err != nil {
		return undoEntry{}, false
	}
	return undoEntry{
		Description: fmt.Sprintf("project %q", project.Title),
		FormKind:    formProject,
		EntityID:    id,
		Restore: func() error {
			return store.UpdateProject(project)
		},
	}, true
}

func (projectHandler) SyncFixedValues(m *Model, specs []columnSpec) {
	typeNames := make([]string, len(m.projectTypes))
	for i, pt := range m.projectTypes {
		typeNames[i] = pt.Name
	}
	setFixedValues(specs, "Type", typeNames)
}

// ---------------------------------------------------------------------------
// quoteHandler
// ---------------------------------------------------------------------------

type quoteHandler struct{}

func (quoteHandler) FormKind() FormKind { return formQuote }

func (quoteHandler) Load(
	store *data.Store,
	showDeleted bool,
) ([]table.Row, []rowMeta, [][]cell, error) {
	quotes, err := store.ListQuotes(showDeleted)
	if err != nil {
		return nil, nil, nil, err
	}
	rows, meta, cellRows := quoteRows(quotes)
	return rows, meta, cellRows, nil
}

func (quoteHandler) Delete(store *data.Store, id uint) error {
	return store.DeleteQuote(id)
}

func (quoteHandler) Restore(store *data.Store, id uint) error {
	return store.RestoreQuote(id)
}

func (quoteHandler) StartAddForm(m *Model) error {
	return m.startQuoteForm()
}

func (quoteHandler) StartEditForm(m *Model, id uint) error {
	return m.startEditQuoteForm(id)
}

func (quoteHandler) InlineEdit(m *Model, id uint, col int) error {
	return m.inlineEditQuote(id, col)
}

func (quoteHandler) SubmitForm(m *Model) error {
	return m.submitQuoteForm()
}

func (quoteHandler) Snapshot(store *data.Store, id uint) (undoEntry, bool) {
	quote, err := store.GetQuote(id)
	if err != nil {
		return undoEntry{}, false
	}
	vendor := quote.Vendor
	return undoEntry{
		Description: fmt.Sprintf("quote from %s", vendor.Name),
		FormKind:    formQuote,
		EntityID:    id,
		Restore: func() error {
			return store.UpdateQuote(quote, vendor)
		},
	}, true
}

func (quoteHandler) SyncFixedValues(_ *Model, _ []columnSpec) {}

// ---------------------------------------------------------------------------
// maintenanceHandler
// ---------------------------------------------------------------------------

type maintenanceHandler struct{}

func (maintenanceHandler) FormKind() FormKind { return formMaintenance }

func (maintenanceHandler) Load(
	store *data.Store,
	showDeleted bool,
) ([]table.Row, []rowMeta, [][]cell, error) {
	items, err := store.ListMaintenance(showDeleted)
	if err != nil {
		return nil, nil, nil, err
	}
	// Batch-fetch service log counts for all items.
	ids := make([]uint, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	logCounts, err := store.CountServiceLogs(ids)
	if err != nil {
		logCounts = map[uint]int{} // non-fatal
	}
	rows, meta, cellRows := maintenanceRows(items, logCounts)
	return rows, meta, cellRows, nil
}

func (maintenanceHandler) Delete(store *data.Store, id uint) error {
	return store.DeleteMaintenance(id)
}

func (maintenanceHandler) Restore(store *data.Store, id uint) error {
	return store.RestoreMaintenance(id)
}

func (maintenanceHandler) StartAddForm(m *Model) error {
	m.startMaintenanceForm()
	return nil
}

func (maintenanceHandler) StartEditForm(m *Model, id uint) error {
	return m.startEditMaintenanceForm(id)
}

func (maintenanceHandler) InlineEdit(m *Model, id uint, col int) error {
	return m.inlineEditMaintenance(id, col)
}

func (maintenanceHandler) SubmitForm(m *Model) error {
	return m.submitMaintenanceForm()
}

func (maintenanceHandler) Snapshot(store *data.Store, id uint) (undoEntry, bool) {
	item, err := store.GetMaintenance(id)
	if err != nil {
		return undoEntry{}, false
	}
	return undoEntry{
		Description: fmt.Sprintf("maintenance %q", item.Name),
		FormKind:    formMaintenance,
		EntityID:    id,
		Restore: func() error {
			return store.UpdateMaintenance(item)
		},
	}, true
}

func (maintenanceHandler) SyncFixedValues(m *Model, specs []columnSpec) {
	catNames := make([]string, len(m.maintenanceCategories))
	for i, c := range m.maintenanceCategories {
		catNames[i] = c.Name
	}
	setFixedValues(specs, "Category", catNames)
}

// ---------------------------------------------------------------------------
// applianceHandler
// ---------------------------------------------------------------------------

type applianceHandler struct{}

func (applianceHandler) FormKind() FormKind { return formAppliance }

func (applianceHandler) Load(
	store *data.Store,
	showDeleted bool,
) ([]table.Row, []rowMeta, [][]cell, error) {
	items, err := store.ListAppliances(showDeleted)
	if err != nil {
		return nil, nil, nil, err
	}
	ids := make([]uint, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	maintCounts, err := store.CountMaintenanceByAppliance(ids)
	if err != nil {
		maintCounts = map[uint]int{}
	}
	rows, meta, cellRows := applianceRows(items, maintCounts, time.Now())
	return rows, meta, cellRows, nil
}

func (applianceHandler) Delete(store *data.Store, id uint) error {
	return store.DeleteAppliance(id)
}

func (applianceHandler) Restore(store *data.Store, id uint) error {
	return store.RestoreAppliance(id)
}

func (applianceHandler) StartAddForm(m *Model) error {
	m.startApplianceForm()
	return nil
}

func (applianceHandler) StartEditForm(m *Model, id uint) error {
	return m.startEditApplianceForm(id)
}

func (applianceHandler) InlineEdit(m *Model, id uint, col int) error {
	return m.inlineEditAppliance(id, col)
}

func (applianceHandler) SubmitForm(m *Model) error {
	return m.submitApplianceForm()
}

func (applianceHandler) Snapshot(store *data.Store, id uint) (undoEntry, bool) {
	item, err := store.GetAppliance(id)
	if err != nil {
		return undoEntry{}, false
	}
	return undoEntry{
		Description: fmt.Sprintf("appliance %q", item.Name),
		FormKind:    formAppliance,
		EntityID:    id,
		Restore: func() error {
			return store.UpdateAppliance(item)
		},
	}, true
}

func (applianceHandler) SyncFixedValues(_ *Model, _ []columnSpec) {}

// ---------------------------------------------------------------------------
// applianceMaintenanceHandler -- detail-view handler for maintenance items
// scoped to a single appliance.
// ---------------------------------------------------------------------------

type applianceMaintenanceHandler struct {
	applianceID uint
}

func (h applianceMaintenanceHandler) FormKind() FormKind { return formMaintenance }

func (h applianceMaintenanceHandler) Load(
	store *data.Store,
	showDeleted bool,
) ([]table.Row, []rowMeta, [][]cell, error) {
	items, err := store.ListMaintenanceByAppliance(h.applianceID, showDeleted)
	if err != nil {
		return nil, nil, nil, err
	}
	ids := make([]uint, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	logCounts, err := store.CountServiceLogs(ids)
	if err != nil {
		logCounts = map[uint]int{}
	}
	rows, meta, cellRows := applianceMaintenanceRows(items, logCounts)
	return rows, meta, cellRows, nil
}

func (h applianceMaintenanceHandler) Delete(store *data.Store, id uint) error {
	return store.DeleteMaintenance(id)
}

func (h applianceMaintenanceHandler) Restore(store *data.Store, id uint) error {
	return store.RestoreMaintenance(id)
}

func (h applianceMaintenanceHandler) StartAddForm(m *Model) error {
	m.startMaintenanceForm()
	return nil
}

func (h applianceMaintenanceHandler) StartEditForm(m *Model, id uint) error {
	return m.startEditMaintenanceForm(id)
}

func (h applianceMaintenanceHandler) InlineEdit(m *Model, id uint, col int) error {
	// Column mapping without Appliance: 0=ID, 1=Item, 2=Category, 3=Last,
	// 4=Next, 5=Every, 6=Log. Remap to the full maintenance column indices
	// (skip col 3=Appliance in the full spec).
	fullCol := col
	if col >= 3 {
		fullCol = col + 1
	}
	return m.inlineEditMaintenance(id, fullCol)
}

func (h applianceMaintenanceHandler) SubmitForm(m *Model) error {
	return m.submitMaintenanceForm()
}

func (applianceMaintenanceHandler) Snapshot(
	store *data.Store,
	id uint,
) (undoEntry, bool) {
	return maintenanceHandler{}.Snapshot(store, id)
}

func (applianceMaintenanceHandler) SyncFixedValues(m *Model, specs []columnSpec) {
	maintenanceHandler{}.SyncFixedValues(m, specs)
}

// ---------------------------------------------------------------------------
// serviceLogHandler -- detail-view handler for service log entries scoped to
// a single maintenance item.
// ---------------------------------------------------------------------------

type serviceLogHandler struct {
	maintenanceItemID uint
}

func (h serviceLogHandler) FormKind() FormKind { return formServiceLog }

func (h serviceLogHandler) Load(
	store *data.Store,
	showDeleted bool,
) ([]table.Row, []rowMeta, [][]cell, error) {
	entries, err := store.ListServiceLog(h.maintenanceItemID, showDeleted)
	if err != nil {
		return nil, nil, nil, err
	}
	rows, meta, cellRows := serviceLogRows(entries)
	return rows, meta, cellRows, nil
}

func (h serviceLogHandler) Delete(store *data.Store, id uint) error {
	return store.DeleteServiceLog(id)
}

func (h serviceLogHandler) Restore(store *data.Store, id uint) error {
	return store.RestoreServiceLog(id)
}

func (h serviceLogHandler) StartAddForm(m *Model) error {
	return m.startServiceLogForm(h.maintenanceItemID)
}

func (h serviceLogHandler) StartEditForm(m *Model, id uint) error {
	return m.startEditServiceLogForm(id)
}

func (h serviceLogHandler) InlineEdit(m *Model, id uint, col int) error {
	return m.inlineEditServiceLog(id, col)
}

func (h serviceLogHandler) SubmitForm(m *Model) error {
	return m.submitServiceLogForm()
}

func (h serviceLogHandler) Snapshot(store *data.Store, id uint) (undoEntry, bool) {
	entry, err := store.GetServiceLog(id)
	if err != nil {
		return undoEntry{}, false
	}
	vendor := entry.Vendor
	return undoEntry{
		Description: fmt.Sprintf("service log %s", entry.ServicedAt.Format("2006-01-02")),
		FormKind:    formServiceLog,
		EntityID:    id,
		Restore: func() error {
			return store.UpdateServiceLog(entry, vendor)
		},
	}, true
}

func (serviceLogHandler) SyncFixedValues(_ *Model, _ []columnSpec) {}

// ---------------------------------------------------------------------------
// vendorHandler
// ---------------------------------------------------------------------------

type vendorHandler struct{}

func (vendorHandler) FormKind() FormKind { return formVendor }

func (vendorHandler) Load(
	store *data.Store,
	showDeleted bool,
) ([]table.Row, []rowMeta, [][]cell, error) {
	vendors, err := store.ListVendors(showDeleted)
	if err != nil {
		return nil, nil, nil, err
	}
	ids := make([]uint, len(vendors))
	for i, v := range vendors {
		ids[i] = v.ID
	}
	quoteCounts, err := store.CountQuotesByVendor(ids)
	if err != nil {
		quoteCounts = map[uint]int{}
	}
	jobCounts, err := store.CountServiceLogsByVendor(ids)
	if err != nil {
		jobCounts = map[uint]int{}
	}
	rows, meta, cellRows := vendorRows(vendors, quoteCounts, jobCounts)
	return rows, meta, cellRows, nil
}

func (vendorHandler) Delete(store *data.Store, id uint) error {
	return store.DeleteVendor(id)
}

func (vendorHandler) Restore(store *data.Store, id uint) error {
	return store.RestoreVendor(id)
}

func (vendorHandler) StartAddForm(m *Model) error {
	m.startVendorForm()
	return nil
}

func (vendorHandler) StartEditForm(m *Model, id uint) error {
	return m.startEditVendorForm(id)
}

func (vendorHandler) InlineEdit(m *Model, id uint, col int) error {
	return m.inlineEditVendor(id, col)
}

func (vendorHandler) SubmitForm(m *Model) error {
	return m.submitVendorForm()
}

func (vendorHandler) Snapshot(store *data.Store, id uint) (undoEntry, bool) {
	vendor, err := store.GetVendor(id)
	if err != nil {
		return undoEntry{}, false
	}
	return undoEntry{
		Description: fmt.Sprintf("vendor %q", vendor.Name),
		FormKind:    formVendor,
		EntityID:    id,
		Restore: func() error {
			return store.UpdateVendor(vendor)
		},
	}, true
}

func (vendorHandler) SyncFixedValues(_ *Model, _ []columnSpec) {}

// ---------------------------------------------------------------------------
// vendorQuoteHandler -- detail-view handler for quotes scoped to a vendor.
// ---------------------------------------------------------------------------

type vendorQuoteHandler struct {
	vendorID uint
}

func (vendorQuoteHandler) FormKind() FormKind { return formQuote }

func (h vendorQuoteHandler) Load(
	store *data.Store,
	showDeleted bool,
) ([]table.Row, []rowMeta, [][]cell, error) {
	quotes, err := store.ListQuotesByVendor(h.vendorID, showDeleted)
	if err != nil {
		return nil, nil, nil, err
	}
	rows, meta, cellRows := vendorQuoteRows(quotes)
	return rows, meta, cellRows, nil
}

func (vendorQuoteHandler) Delete(store *data.Store, id uint) error {
	return store.DeleteQuote(id)
}

func (vendorQuoteHandler) Restore(store *data.Store, id uint) error {
	return store.RestoreQuote(id)
}

func (vendorQuoteHandler) StartAddForm(m *Model) error {
	return m.startQuoteForm()
}

func (vendorQuoteHandler) StartEditForm(m *Model, id uint) error {
	return m.startEditQuoteForm(id)
}

func (vendorQuoteHandler) InlineEdit(m *Model, id uint, col int) error {
	// Columns: 0=ID, 1=Project, 2=Total, 3=Labor, 4=Mat, 5=Other, 6=Recv.
	// Map to full quote columns (skip col 2=Vendor).
	fullCol := col
	if col >= 2 {
		fullCol = col + 1
	}
	return m.inlineEditQuote(id, fullCol)
}

func (vendorQuoteHandler) SubmitForm(m *Model) error {
	return m.submitQuoteForm()
}

func (vendorQuoteHandler) Snapshot(store *data.Store, id uint) (undoEntry, bool) {
	return quoteHandler{}.Snapshot(store, id)
}

func (vendorQuoteHandler) SyncFixedValues(_ *Model, _ []columnSpec) {}

// ---------------------------------------------------------------------------
// vendorJobsHandler -- detail-view handler for service log entries scoped to
// a vendor.
// ---------------------------------------------------------------------------

type vendorJobsHandler struct {
	vendorID uint
}

func (vendorJobsHandler) FormKind() FormKind { return formServiceLog }

func (h vendorJobsHandler) Load(
	store *data.Store,
	showDeleted bool,
) ([]table.Row, []rowMeta, [][]cell, error) {
	entries, err := store.ListServiceLogsByVendor(h.vendorID, showDeleted)
	if err != nil {
		return nil, nil, nil, err
	}
	rows, meta, cellRows := vendorJobsRows(entries)
	return rows, meta, cellRows, nil
}

func (vendorJobsHandler) Delete(store *data.Store, id uint) error {
	return store.DeleteServiceLog(id)
}

func (vendorJobsHandler) Restore(store *data.Store, id uint) error {
	return store.RestoreServiceLog(id)
}

func (h vendorJobsHandler) StartAddForm(m *Model) error {
	// Service log entries need a maintenance item ID; cannot add from
	// the vendor-scoped view since the parent maintenance item is unknown.
	return fmt.Errorf("add service log entries from the Maintenance tab")
}

func (vendorJobsHandler) StartEditForm(m *Model, id uint) error {
	return m.startEditServiceLogForm(id)
}

func (vendorJobsHandler) InlineEdit(m *Model, id uint, col int) error {
	// Columns: 0=ID, 1=Item, 2=Date, 3=Cost, 4=Notes.
	// Map to full service log columns: 0=ID, 1=Date, 2=Performed By, 3=Cost, 4=Notes.
	switch col {
	case 2:
		return m.inlineEditServiceLog(id, 1) // Date
	case 3:
		return m.inlineEditServiceLog(id, 3) // Cost
	default:
		return nil
	}
}

func (vendorJobsHandler) SubmitForm(m *Model) error {
	return m.submitServiceLogForm()
}

func (vendorJobsHandler) Snapshot(store *data.Store, id uint) (undoEntry, bool) {
	return serviceLogHandler{}.Snapshot(store, id)
}

func (vendorJobsHandler) SyncFixedValues(_ *Model, _ []columnSpec) {}

// ---------------------------------------------------------------------------
// projectQuoteHandler -- detail-view handler for quotes scoped to a project.
// ---------------------------------------------------------------------------

type projectQuoteHandler struct {
	projectID uint
}

func (projectQuoteHandler) FormKind() FormKind { return formQuote }

func (h projectQuoteHandler) Load(
	store *data.Store,
	showDeleted bool,
) ([]table.Row, []rowMeta, [][]cell, error) {
	quotes, err := store.ListQuotesByProject(h.projectID, showDeleted)
	if err != nil {
		return nil, nil, nil, err
	}
	rows, meta, cellRows := projectQuoteRows(quotes)
	return rows, meta, cellRows, nil
}

func (projectQuoteHandler) Delete(store *data.Store, id uint) error {
	return store.DeleteQuote(id)
}

func (projectQuoteHandler) Restore(store *data.Store, id uint) error {
	return store.RestoreQuote(id)
}

func (projectQuoteHandler) StartAddForm(m *Model) error {
	return m.startQuoteForm()
}

func (projectQuoteHandler) StartEditForm(m *Model, id uint) error {
	return m.startEditQuoteForm(id)
}

func (projectQuoteHandler) InlineEdit(m *Model, id uint, col int) error {
	// Columns: 0=ID, 1=Vendor, 2=Total, 3=Labor, 4=Mat, 5=Other, 6=Recv.
	// Map to full quote columns (skip col 1=Project).
	fullCol := col
	if col >= 1 {
		fullCol = col + 1
	}
	return m.inlineEditQuote(id, fullCol)
}

func (projectQuoteHandler) SubmitForm(m *Model) error {
	return m.submitQuoteForm()
}

func (projectQuoteHandler) Snapshot(store *data.Store, id uint) (undoEntry, bool) {
	return quoteHandler{}.Snapshot(store, id)
}

func (projectQuoteHandler) SyncFixedValues(_ *Model, _ []columnSpec) {}

// ---------------------------------------------------------------------------
// documentHandler -- top-level handler for the Documents tab.
// ---------------------------------------------------------------------------

type documentHandler struct{}

func (documentHandler) FormKind() FormKind { return formDocument }

func (documentHandler) Load(
	store *data.Store,
	showDeleted bool,
) ([]table.Row, []rowMeta, [][]cell, error) {
	docs, err := store.ListDocuments(showDeleted)
	if err != nil {
		return nil, nil, nil, err
	}
	rows, meta, cellRows := documentRows(docs)
	return rows, meta, cellRows, nil
}

func (documentHandler) Delete(store *data.Store, id uint) error {
	return store.DeleteDocument(id)
}

func (documentHandler) Restore(store *data.Store, id uint) error {
	return store.RestoreDocument(id)
}

func (documentHandler) StartAddForm(m *Model) error {
	return m.startDocumentForm("")
}

func (documentHandler) StartEditForm(m *Model, id uint) error {
	return m.startEditDocumentForm(id)
}

func (documentHandler) InlineEdit(m *Model, id uint, col int) error {
	return m.inlineEditDocument(id, col)
}

func (documentHandler) SubmitForm(m *Model) error {
	return m.submitDocumentForm()
}

func (documentHandler) Snapshot(store *data.Store, id uint) (undoEntry, bool) {
	doc, err := store.GetDocument(id)
	if err != nil {
		return undoEntry{}, false
	}
	return undoEntry{
		Description: fmt.Sprintf("document %q", doc.Title),
		FormKind:    formDocument,
		EntityID:    id,
		Restore: func() error {
			return store.UpdateDocument(doc)
		},
	}, true
}

func (documentHandler) SyncFixedValues(_ *Model, _ []columnSpec) {}
