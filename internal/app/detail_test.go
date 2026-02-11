// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/cpcloud/micasa/internal/data"
)

func TestOpenDetailSetsContext(t *testing.T) {
	m := newTestModel()
	// Navigate to the maintenance tab.
	m.active = tabIndex(tabMaintenance)
	if m.detail != nil {
		t.Fatal("expected nil detail before open")
	}

	err := m.openServiceLogDetail(42, "Test Item")
	if err != nil {
		t.Fatalf("openServiceLogDetail error: %v", err)
	}
	if m.detail == nil {
		t.Fatal("expected non-nil detail after open")
	}
	if m.detail.ParentRowID != 42 {
		t.Fatalf("expected ParentRowID=42, got %d", m.detail.ParentRowID)
	}
	if m.detail.Breadcrumb != "Maintenance > Test Item" {
		t.Fatalf("unexpected breadcrumb: %q", m.detail.Breadcrumb)
	}
}

func TestCloseDetailRestoresParent(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(42, "Test Item")

	m.closeDetail()
	if m.detail != nil {
		t.Fatal("expected nil detail after close")
	}
	if m.active != tabIndex(tabMaintenance) {
		t.Fatalf("expected active=%d, got %d", tabIndex(tabMaintenance), m.active)
	}
}

func TestEffectiveTabReturnsDetailWhenOpen(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	mainTab := m.effectiveTab()
	if mainTab == nil || mainTab.Kind != tabMaintenance {
		t.Fatal("expected maintenance tab before detail open")
	}

	_ = m.openServiceLogDetail(1, "Test")
	detailTab := m.effectiveTab()
	if detailTab == nil {
		t.Fatal("expected non-nil effective tab in detail view")
	}
	if detailTab.Handler == nil {
		t.Fatal("expected handler on detail tab")
	}
	if detailTab.Handler.FormKind() != formServiceLog {
		t.Fatalf("expected formServiceLog, got %d", detailTab.Handler.FormKind())
	}
}

func TestEffectiveTabFallsBackToMainTab(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabProjects)
	tab := m.effectiveTab()
	if tab == nil || tab.Kind != tabProjects {
		t.Fatal("expected projects tab when no detail")
	}
}

func TestEscInNormalModeClosesDetail(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")
	if m.detail == nil {
		t.Fatal("expected detail open")
	}
	sendKey(m, "esc")
	if m.detail != nil {
		t.Fatal("expected detail closed after esc in normal mode")
	}
}

func TestEscInEditModeDoesNotCloseDetail(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	sendKey(m, "i") // enter edit mode
	if m.mode != modeEdit {
		t.Fatal("expected edit mode")
	}
	sendKey(m, "esc") // should go to normal, not close detail
	if m.mode != modeNormal {
		t.Fatal("expected normal mode after esc in edit mode")
	}
	if m.detail == nil {
		t.Fatal("expected detail still open after edit-mode esc")
	}
}

func TestTabSwitchBlockedInDetailView(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	before := m.active
	sendKey(m, "tab")
	if m.active != before {
		t.Fatal("tab switch should be blocked while in detail view")
	}
}

func TestColumnNavWorksInDetailView(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	tab := m.effectiveTab()
	if tab == nil {
		t.Fatal("no effective tab")
	}
	initial := tab.ColCursor
	sendKey(m, "l")
	if tab.ColCursor == initial && len(tab.Specs) > 1 {
		t.Fatal("expected column cursor to advance in detail view")
	}
}

func TestDetailTabHasServiceLogSpecs(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	tab := m.effectiveTab()
	specs := tab.Specs
	// Expect: ID, Date, Performed By, Cost, Notes
	if len(specs) != 5 {
		t.Fatalf("expected 5 service log columns, got %d", len(specs))
	}
	titles := make([]string, len(specs))
	for i, s := range specs {
		titles[i] = s.Title
	}
	expected := []string{"ID", "Date", "Performed By", "Cost", "Notes"}
	for i, want := range expected {
		if titles[i] != want {
			t.Fatalf("column %d: expected %q, got %q", i, want, titles[i])
		}
	}
}

func TestHandlerForFormKindFindsDetailHandler(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	handler := m.handlerForFormKind(formServiceLog)
	if handler == nil {
		t.Fatal("expected to find service log handler via handlerForFormKind")
	}
	if handler.FormKind() != formServiceLog {
		t.Fatalf("expected formServiceLog, got %d", handler.FormKind())
	}
}

func TestServiceLogHandlerFormKind(t *testing.T) {
	h := serviceLogHandler{maintenanceItemID: 5}
	if h.FormKind() != formServiceLog {
		t.Fatalf("expected formServiceLog, got %d", h.FormKind())
	}
}

func TestMaintenanceColumnsIncludeLog(t *testing.T) {
	specs := maintenanceColumnSpecs()
	last := specs[len(specs)-1]
	if last.Title != "Log" {
		t.Fatalf("expected last maintenance column to be 'Log', got %q", last.Title)
	}
	if last.Kind != cellDrilldown {
		t.Fatal("expected Log column to be drilldown")
	}
}

func TestApplianceColumnsIncludeMaint(t *testing.T) {
	specs := applianceColumnSpecs()
	last := specs[len(specs)-1]
	if last.Title != "Maint" {
		t.Fatalf("expected last appliance column to be 'Maint', got %q", last.Title)
	}
	if last.Kind != cellDrilldown {
		t.Fatal("expected Maint column to be drilldown")
	}
}

func TestVendorOptions(t *testing.T) {
	m := newTestModel()
	// No vendors loaded -- should just have "Self (homeowner)".
	opts := vendorOptions(m.vendors)
	if len(opts) < 1 {
		t.Fatal("expected at least 1 vendor option (Self)")
	}
	// First option value should be 0 (self).
	if opts[0].Value != 0 {
		t.Fatalf("expected first vendor option value=0 (Self), got %d", opts[0].Value)
	}
}

func TestServiceLogColumnSpecs(t *testing.T) {
	specs := serviceLogColumnSpecs()
	if len(specs) != 5 {
		t.Fatalf("expected 5 columns, got %d", len(specs))
	}
	// Verify the "Performed By" column is flex and linked to vendors.
	pb := specs[2]
	if !pb.Flex {
		t.Fatal("expected 'Performed By' column to be flex")
	}
	if pb.Link == nil {
		t.Fatal("expected 'Performed By' column to link to vendors")
	}
	if pb.Link.TargetTab != tabVendors {
		t.Fatalf("expected link to tabVendors, got %v", pb.Link.TargetTab)
	}
}

func TestServiceLogRowsSelfPerformed(t *testing.T) {
	entries := []data.ServiceLogEntry{
		{
			ID:         1,
			ServicedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			Notes:      "test note",
		},
	}
	_, meta, cellRows := serviceLogRows(entries)
	if len(cellRows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(cellRows))
	}
	if cellRows[0][2].Value != "Self" {
		t.Fatalf("expected 'Self' for performed by, got %q", cellRows[0][2].Value)
	}
	if meta[0].ID != 1 {
		t.Fatalf("expected meta ID=1, got %d", meta[0].ID)
	}
}

func TestServiceLogRowsVendorPerformed(t *testing.T) {
	vendorID := uint(5)
	entries := []data.ServiceLogEntry{
		{
			ID:         2,
			ServicedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			VendorID:   &vendorID,
			Vendor:     data.Vendor{Name: "Acme Plumbing"},
		},
	}
	_, _, cellRows := serviceLogRows(entries)
	if cellRows[0][2].Value != "Acme Plumbing" {
		t.Fatalf("expected 'Acme Plumbing', got %q", cellRows[0][2].Value)
	}
	if cellRows[0][2].LinkID != 5 {
		t.Fatalf("expected LinkID=5 for vendor, got %d", cellRows[0][2].LinkID)
	}
}

func TestServiceLogRowsSelfHasNoLink(t *testing.T) {
	entries := []data.ServiceLogEntry{
		{
			ID:         1,
			ServicedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	_, _, cellRows := serviceLogRows(entries)
	if cellRows[0][2].LinkID != 0 {
		t.Fatalf("expected LinkID=0 for self-performed, got %d", cellRows[0][2].LinkID)
	}
}

func TestMaintenanceLogColumnReplacedManual(t *testing.T) {
	specs := maintenanceColumnSpecs()
	// The old "Manual" column should be gone, replaced by "Log".
	for _, s := range specs {
		if s.Title == "Manual" {
			t.Fatal("expected 'Manual' column to be replaced by 'Log'")
		}
	}
}

func TestNewTestModelDetailNil(t *testing.T) {
	m := newTestModel()
	if m.detail != nil {
		t.Fatal("new test model should have nil detail")
	}
}

func TestResizeTablesIncludesDetail(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	m.resizeTables()
	detailH := m.detail.Tab.Table.Height()
	if detailH <= 0 {
		t.Fatalf("expected detail table height > 0, got %d", detailH)
	}
}

func TestSortWorksInDetailView(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	tab := m.effectiveTab()
	tab.ColCursor = 1 // Date column

	sendKey(m, "s")
	if len(tab.Sorts) == 0 {
		t.Fatal("expected sort entry after 's' in detail view")
	}
}

// newTestModelWithDetailRows creates a model with detail open and seeded rows.
func newTestModelWithDetailRows() *Model {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	tab := m.effectiveTab()
	// Seed a couple rows.
	tab.Table.SetRows([]table.Row{
		{"1", "2026-01-15", "Self", "", "first"},
		{"2", "2026-02-01", "Acme", "$150.00", "second"},
	})
	tab.Rows = []rowMeta{{ID: 1}, {ID: 2}}
	tab.CellRows = [][]cell{
		{
			{Value: "1", Kind: cellReadonly},
			{Value: "2026-01-15", Kind: cellDate},
			{Value: "Self", Kind: cellText},
			{Value: "", Kind: cellMoney},
			{Value: "first", Kind: cellText},
		},
		{
			{Value: "2", Kind: cellReadonly},
			{Value: "2026-02-01", Kind: cellDate},
			{Value: "Acme", Kind: cellText},
			{Value: "$150.00", Kind: cellMoney},
			{Value: "second", Kind: cellText},
		},
	}
	return m
}

func TestSelectedRowMetaUsesDetailTab(t *testing.T) {
	m := newTestModelWithDetailRows()
	meta, ok := m.selectedRowMeta()
	if !ok {
		t.Fatal("expected row meta from detail tab")
	}
	if meta.ID != 1 {
		t.Fatalf("expected ID=1, got %d", meta.ID)
	}
}

func TestSelectedCellUsesDetailTab(t *testing.T) {
	m := newTestModelWithDetailRows()
	c, ok := m.selectedCell(2)
	if !ok {
		t.Fatal("expected cell from detail tab")
	}
	if c.Value != "Self" {
		t.Fatalf("expected 'Self', got %q", c.Value)
	}
}

func TestApplianceMaintenanceDetailOpens(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabAppliances)
	err := m.openApplianceMaintenanceDetail(5, "Dishwasher")
	if err != nil {
		t.Fatalf("openApplianceMaintenanceDetail error: %v", err)
	}
	if m.detail == nil {
		t.Fatal("detail should be set")
	}
	if m.detail.Breadcrumb != "Appliances > Dishwasher" {
		t.Fatalf("unexpected breadcrumb: %q", m.detail.Breadcrumb)
	}
	if m.detail.Tab.Name != "Maintenance" {
		t.Fatalf("unexpected detail tab name: %q", m.detail.Tab.Name)
	}
	if m.detail.Tab.Kind != tabAppliances {
		t.Fatalf("expected detail tab kind=tabAppliances, got %d", m.detail.Tab.Kind)
	}
}

func TestApplianceMaintenanceHandlerFormKind(t *testing.T) {
	h := applianceMaintenanceHandler{applianceID: 1}
	if h.FormKind() != formMaintenance {
		t.Fatal("expected formMaintenance")
	}
}

func TestApplianceMaintenanceColumnSpecsNoApplianceOrLog(t *testing.T) {
	specs := applianceMaintenanceColumnSpecs()
	for _, s := range specs {
		if s.Title == "Appliance" {
			t.Fatal("appliance maintenance detail should not include Appliance column")
		}
		if s.Title == "Log" {
			t.Fatal(
				"appliance maintenance detail should not include Log column (no nested drilldown yet)",
			)
		}
	}
	last := specs[len(specs)-1]
	if last.Title != "Every" {
		t.Fatalf("expected last column to be 'Every', got %q", last.Title)
	}
}

func TestApplianceMaintColumnIsDrilldown(t *testing.T) {
	specs := applianceColumnSpecs()
	last := specs[len(specs)-1]
	if last.Title != "Maint" {
		t.Fatalf("expected last column to be 'Maint', got %q", last.Title)
	}
	if last.Kind != cellDrilldown {
		t.Fatal("expected Maint column to be drilldown")
	}
}
