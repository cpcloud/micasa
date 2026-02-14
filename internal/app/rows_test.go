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
	"gorm.io/gorm"
)

func TestDateValue(t *testing.T) {
	assert.Empty(t, dateValue(nil))
	d := time.Date(2025, 6, 11, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, "2025-06-11", dateValue(&d))
}

func TestCentsValue(t *testing.T) {
	assert.Empty(t, centsValue(nil))
	c := int64(123456)
	assert.Equal(t, "$1,234.56", centsValue(&c))
}

func TestProjectRows(t *testing.T) {
	budget := int64(100000)
	start := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	projects := []data.Project{
		{
			ID:            1,
			Title:         "Kitchen",
			ProjectType:   data.ProjectType{Name: "Renovation"},
			ProjectTypeID: 1,
			Status:        data.ProjectStatusPlanned,
			BudgetCents:   &budget,
			StartDate:     &start,
		},
	}
	rows, meta, cells := projectRows(projects, nil)
	require.Len(t, rows, 1)
	assert.Equal(t, uint(1), meta[0].ID)
	assert.False(t, meta[0].Deleted)
	assert.Equal(t, "Kitchen", cells[0][2].Value)
	assert.Equal(t, "$1,000.00", cells[0][4].Value)
	assert.Equal(t, "2025-03-01", cells[0][6].Value)
	assert.Equal(t, "Kitchen", rows[0][2])
}

func TestProjectRowsDeleted(t *testing.T) {
	projects := []data.Project{
		{
			ID:        1,
			Title:     "Old Project",
			DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
		},
	}
	_, meta, _ := projectRows(projects, nil)
	assert.True(t, meta[0].Deleted)
}

func TestQuoteRows(t *testing.T) {
	labor := int64(20000)
	quotes := []data.Quote{
		{
			ID:         1,
			ProjectID:  1,
			Project:    data.Project{Title: "Kitchen"},
			Vendor:     data.Vendor{Name: "ContractorCo"},
			VendorID:   1,
			TotalCents: 50000,
			LaborCents: &labor,
		},
	}
	rows, meta, cells := quoteRows(quotes)
	require.Len(t, rows, 1)
	assert.Equal(t, uint(1), meta[0].ID)
	assert.Equal(t, "Kitchen", cells[0][1].Value)
	assert.Equal(t, uint(1), cells[0][1].LinkID)
	assert.Equal(t, "ContractorCo", cells[0][2].Value)
	assert.Equal(t, "$500.00", cells[0][3].Value)
}

func TestQuoteRowsFallbackProjectName(t *testing.T) {
	quotes := []data.Quote{
		{
			ID:         1,
			ProjectID:  42,
			TotalCents: 100,
		},
	}
	_, _, cells := quoteRows(quotes)
	assert.Equal(t, "Project 42", cells[0][1].Value)
}

func TestMaintenanceRows(t *testing.T) {
	appID := uint(5)
	lastServiced := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	items := []data.MaintenanceItem{
		{
			ID:             1,
			Name:           "HVAC Filter",
			Category:       data.MaintenanceCategory{Name: "HVAC"},
			ApplianceID:    &appID,
			Appliance:      data.Appliance{Name: "AC Unit"},
			LastServicedAt: &lastServiced,
			IntervalMonths: 3,
		},
	}
	logCounts := map[uint]int{1: 4}
	rows, meta, cells := maintenanceRows(items, logCounts)
	require.Len(t, rows, 1)
	assert.Equal(t, uint(1), meta[0].ID)
	assert.Equal(t, "HVAC Filter", cells[0][1].Value)
	assert.Equal(t, "HVAC", cells[0][2].Value)
	assert.Equal(t, "AC Unit", cells[0][3].Value)
	assert.Equal(t, uint(5), cells[0][3].LinkID)
	assert.Equal(t, "3m", cells[0][6].Value)
	assert.Equal(t, "4", cells[0][7].Value)
}

func TestMaintenanceRowsNoAppliance(t *testing.T) {
	items := []data.MaintenanceItem{
		{ID: 1, Name: "Gutters", Category: data.MaintenanceCategory{Name: "Exterior"}},
	}
	_, _, cells := maintenanceRows(items, nil)
	assert.Empty(t, cells[0][3].Value)
	assert.Zero(t, cells[0][3].LinkID)
}

func TestApplianceRows(t *testing.T) {
	cost := int64(89900)
	purchase := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	items := []data.Appliance{
		{
			ID:           1,
			Name:         "Fridge",
			Brand:        "Samsung",
			ModelNumber:  "RF28R",
			PurchaseDate: &purchase,
			CostCents:    &cost,
		},
	}
	maintCounts := map[uint]int{1: 2}
	rows, meta, cells := applianceRows(items, maintCounts, now)
	require.Len(t, rows, 1)
	assert.Equal(t, uint(1), meta[0].ID)
	assert.Equal(t, "Fridge", cells[0][1].Value)
	assert.Equal(t, "Samsung", cells[0][2].Value)
	assert.Equal(t, "2023-06-15", cells[0][6].Value)
	assert.Equal(t, "2y", cells[0][7].Value)
	assert.Equal(t, "$899.00", cells[0][9].Value)
	assert.Equal(t, "2", cells[0][10].Value)
}

func TestApplianceRowsNoOptionalFields(t *testing.T) {
	now := time.Now()
	items := []data.Appliance{
		{ID: 1, Name: "Lamp"},
	}
	_, _, cells := applianceRows(items, nil, now)
	assert.Empty(t, cells[0][6].Value, "expected empty purchase date")
	assert.Empty(t, cells[0][7].Value, "expected empty age")
	assert.Empty(t, cells[0][9].Value, "expected empty cost")
	assert.Empty(t, cells[0][10].Value, "expected empty maint count")
}

func TestBuildRowsEmpty(t *testing.T) {
	rows, meta, cells := projectRows(nil, nil)
	assert.Empty(t, rows)
	assert.Empty(t, meta)
	assert.Empty(t, cells)
}

func TestDocumentRows(t *testing.T) {
	docs := []data.Document{
		{
			ID:         1,
			Title:      "Invoice",
			EntityKind: data.DocumentEntityProject,
			EntityID:   42,
			MIMEType:   "application/pdf",
			SizeBytes:  2048,
			UpdatedAt:  time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	rows, meta, cells := documentRows(docs)
	require.Len(t, rows, 1)
	assert.Equal(t, uint(1), meta[0].ID)
	assert.Equal(t, "Invoice", cells[0][1].Value)
	assert.Equal(t, "project #42", cells[0][2].Value)
	assert.Equal(t, "application/pdf", cells[0][3].Value)
	assert.Equal(t, "2.0 KB", cells[0][4].Value)
}

func TestEntityDocumentRows(t *testing.T) {
	docs := []data.Document{
		{
			ID:        1,
			Title:     "Manual",
			MIMEType:  "application/pdf",
			SizeBytes: 1048576, // 1 MB
			UpdatedAt: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	rows, meta, cells := entityDocumentRows(docs)
	require.Len(t, rows, 1)
	assert.Equal(t, uint(1), meta[0].ID)
	assert.Equal(t, "Manual", cells[0][1].Value)
	assert.Equal(t, "application/pdf", cells[0][2].Value)
	assert.Equal(t, "1.0 MB", cells[0][3].Value)
}

func TestFormatFileSize(t *testing.T) {
	assert.Empty(t, formatFileSize(0))
	assert.Equal(t, "512 B", formatFileSize(512))
	assert.Equal(t, "1.0 KB", formatFileSize(1024))
	assert.Equal(t, "1.5 KB", formatFileSize(1536))
	assert.Equal(t, "1.0 MB", formatFileSize(1024*1024))
	assert.Equal(t, "1.0 GB", formatFileSize(1024*1024*1024))
}

func TestDocumentEntityLabel(t *testing.T) {
	assert.Empty(t, documentEntityLabel("", 0))
	assert.Equal(t, "project #5", documentEntityLabel("project", 5))
	assert.Equal(t, "appliance #12", documentEntityLabel("appliance", 12))
}

func TestCellsToRow(t *testing.T) {
	cells := []cell{
		{Value: "1"},
		{Value: "hello"},
		{Value: "$100.00"},
	}
	row := cellsToRow(cells)
	assert.Equal(t, table.Row{"1", "hello", "$100.00"}, row)
}
