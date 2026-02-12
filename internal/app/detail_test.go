// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/cpcloud/micasa/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDetailSetsContext(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	require.Nil(t, m.detail)

	require.NoError(t, m.openServiceLogDetail(42, "Test Item"))
	require.NotNil(t, m.detail)
	assert.Equal(t, uint(42), m.detail.ParentRowID)
	assert.Equal(t, "Maintenance > Test Item", m.detail.Breadcrumb)
}

func TestCloseDetailRestoresParent(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(42, "Test Item")

	m.closeDetail()
	assert.Nil(t, m.detail)
	assert.Equal(t, tabIndex(tabMaintenance), m.active)
}

func TestEffectiveTabReturnsDetailWhenOpen(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	mainTab := m.effectiveTab()
	require.NotNil(t, mainTab)
	assert.Equal(t, tabMaintenance, mainTab.Kind)

	_ = m.openServiceLogDetail(1, "Test")
	detailTab := m.effectiveTab()
	require.NotNil(t, detailTab)
	require.NotNil(t, detailTab.Handler)
	assert.Equal(t, formServiceLog, detailTab.Handler.FormKind())
}

func TestEffectiveTabFallsBackToMainTab(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabProjects)
	tab := m.effectiveTab()
	require.NotNil(t, tab)
	assert.Equal(t, tabProjects, tab.Kind)
}

func TestEscInNormalModeClosesDetail(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")
	require.NotNil(t, m.detail)
	sendKey(m, "esc")
	assert.Nil(t, m.detail)
}

func TestEscInEditModeDoesNotCloseDetail(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	sendKey(m, "i") // enter edit mode
	require.Equal(t, modeEdit, m.mode)
	sendKey(m, "esc") // should go to normal, not close detail
	assert.Equal(t, modeNormal, m.mode)
	assert.NotNil(t, m.detail, "expected detail still open after edit-mode esc")
}

func TestTabSwitchBlockedInDetailView(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	before := m.active
	sendKey(m, "f")
	assert.Equal(t, before, m.active, "tab switch should be blocked while in detail view")
}

func TestColumnNavWorksInDetailView(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	tab := m.effectiveTab()
	require.NotNil(t, tab)
	initial := tab.ColCursor
	sendKey(m, "l")
	if len(tab.Specs) > 1 {
		assert.NotEqual(
			t,
			initial,
			tab.ColCursor,
			"expected column cursor to advance in detail view",
		)
	}
}

func TestDetailTabHasServiceLogSpecs(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	tab := m.effectiveTab()
	require.Len(t, tab.Specs, 5)
	expected := []string{"ID", "Date", "Performed By", "Cost", "Notes"}
	for i, want := range expected {
		assert.Equalf(t, want, tab.Specs[i].Title, "column %d", i)
	}
}

func TestHandlerForFormKindFindsDetailHandler(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	handler := m.handlerForFormKind(formServiceLog)
	require.NotNil(t, handler)
	assert.Equal(t, formServiceLog, handler.FormKind())
}

func TestServiceLogHandlerFormKind(t *testing.T) {
	h := serviceLogHandler{maintenanceItemID: 5}
	assert.Equal(t, formServiceLog, h.FormKind())
}

func TestMaintenanceColumnsIncludeLog(t *testing.T) {
	specs := maintenanceColumnSpecs()
	last := specs[len(specs)-1]
	assert.Equal(t, "Log", last.Title)
	assert.Equal(t, cellDrilldown, last.Kind)
}

func TestApplianceColumnsIncludeMaint(t *testing.T) {
	specs := applianceColumnSpecs()
	last := specs[len(specs)-1]
	assert.Equal(t, "Maint", last.Title)
	assert.Equal(t, cellDrilldown, last.Kind)
}

func TestVendorOptions(t *testing.T) {
	m := newTestModel()
	opts := vendorOptions(m.vendors)
	require.NotEmpty(t, opts, "expected at least 1 vendor option (Self)")
	assert.Equal(t, uint(0), opts[0].Value, "expected first vendor option value=0 (Self)")
}

func TestServiceLogColumnSpecs(t *testing.T) {
	specs := serviceLogColumnSpecs()
	require.Len(t, specs, 5)
	// Verify the "Performed By" column is flex and linked to vendors.
	pb := specs[2]
	assert.True(t, pb.Flex)
	require.NotNil(t, pb.Link)
	assert.Equal(t, tabVendors, pb.Link.TargetTab)
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
	require.Len(t, cellRows, 1)
	assert.Equal(t, "Self", cellRows[0][2].Value)
	assert.Equal(t, uint(1), meta[0].ID)
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
	assert.Equal(t, "Acme Plumbing", cellRows[0][2].Value)
	assert.Equal(t, uint(5), cellRows[0][2].LinkID)
}

func TestServiceLogRowsSelfHasNoLink(t *testing.T) {
	entries := []data.ServiceLogEntry{
		{
			ID:         1,
			ServicedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	_, _, cellRows := serviceLogRows(entries)
	assert.Zero(t, cellRows[0][2].LinkID)
}

func TestMaintenanceLogColumnReplacedManual(t *testing.T) {
	specs := maintenanceColumnSpecs()
	for _, s := range specs {
		assert.NotEqual(t, "Manual", s.Title, "expected 'Manual' column to be replaced by 'Log'")
	}
}

func TestNewTestModelDetailNil(t *testing.T) {
	m := newTestModel()
	assert.Nil(t, m.detail)
}

func TestResizeTablesIncludesDetail(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	m.resizeTables()
	assert.Greater(t, m.detail.Tab.Table.Height(), 0)
}

func TestSortWorksInDetailView(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabMaintenance)
	_ = m.openServiceLogDetail(1, "Test")

	tab := m.effectiveTab()
	tab.ColCursor = 1 // Date column

	sendKey(m, "s")
	assert.NotEmpty(t, tab.Sorts, "expected sort entry after 's' in detail view")
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
	require.True(t, ok)
	assert.Equal(t, uint(1), meta.ID)
}

func TestSelectedCellUsesDetailTab(t *testing.T) {
	m := newTestModelWithDetailRows()
	c, ok := m.selectedCell(2)
	require.True(t, ok)
	assert.Equal(t, "Self", c.Value)
}

func TestApplianceMaintenanceDetailOpens(t *testing.T) {
	m := newTestModel()
	m.active = tabIndex(tabAppliances)
	require.NoError(t, m.openApplianceMaintenanceDetail(5, "Dishwasher"))
	require.NotNil(t, m.detail)
	assert.Equal(t, "Appliances > Dishwasher", m.detail.Breadcrumb)
	assert.Equal(t, "Maintenance", m.detail.Tab.Name)
	assert.Equal(t, tabAppliances, m.detail.Tab.Kind)
}

func TestApplianceMaintenanceHandlerFormKind(t *testing.T) {
	h := applianceMaintenanceHandler{applianceID: 1}
	assert.Equal(t, formMaintenance, h.FormKind())
}

func TestApplianceMaintenanceColumnSpecsNoApplianceOrLog(t *testing.T) {
	specs := applianceMaintenanceColumnSpecs()
	for _, s := range specs {
		assert.NotEqual(
			t,
			"Appliance",
			s.Title,
			"appliance maintenance detail should not include Appliance column",
		)
		assert.NotEqual(
			t,
			"Log",
			s.Title,
			"appliance maintenance detail should not include Log column",
		)
	}
	last := specs[len(specs)-1]
	assert.Equal(t, "Every", last.Title)
}

func TestApplianceMaintColumnIsDrilldown(t *testing.T) {
	specs := applianceColumnSpecs()
	last := specs[len(specs)-1]
	assert.Equal(t, "Maint", last.Title)
	assert.Equal(t, cellDrilldown, last.Kind)
}
