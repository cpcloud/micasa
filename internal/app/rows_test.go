// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
	"time"

	"github.com/cpcloud/micasa/internal/data"
	"gorm.io/gorm"
)

func TestDateValue(t *testing.T) {
	if got := dateValue(nil); got != "" {
		t.Fatalf("dateValue(nil) = %q, want empty", got)
	}
	d := time.Date(2025, 6, 11, 0, 0, 0, 0, time.UTC)
	if got := dateValue(&d); got != "2025-06-11" {
		t.Fatalf("dateValue = %q, want 2025-06-11", got)
	}
}

func TestCentsValue(t *testing.T) {
	if got := centsValue(nil); got != "" {
		t.Fatalf("centsValue(nil) = %q, want empty", got)
	}
	c := int64(123456)
	if got := centsValue(&c); got != "$1,234.56" {
		t.Fatalf("centsValue = %q, want $1,234.56", got)
	}
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
	rows, meta, cells := projectRows(projects)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if meta[0].ID != 1 || meta[0].Deleted {
		t.Fatalf("unexpected meta: %+v", meta[0])
	}
	// Title is at col 2.
	if cells[0][2].Value != "Kitchen" {
		t.Fatalf("title cell = %q", cells[0][2].Value)
	}
	// Budget is at col 4.
	if cells[0][4].Value != "$1,000.00" {
		t.Fatalf("budget cell = %q", cells[0][4].Value)
	}
	// Start date at col 6.
	if cells[0][6].Value != "2025-03-01" {
		t.Fatalf("start date cell = %q", cells[0][6].Value)
	}
	// Rows and cells should have the same values.
	if rows[0][2] != "Kitchen" {
		t.Fatalf("row title = %q", rows[0][2])
	}
}

func TestProjectRowsDeleted(t *testing.T) {
	projects := []data.Project{
		{
			ID:        1,
			Title:     "Old Project",
			DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
		},
	}
	_, meta, _ := projectRows(projects)
	if !meta[0].Deleted {
		t.Fatal("expected deleted=true")
	}
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
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if meta[0].ID != 1 {
		t.Fatalf("unexpected ID: %d", meta[0].ID)
	}
	// Project name at col 1.
	if cells[0][1].Value != "Kitchen" {
		t.Fatalf("project cell = %q", cells[0][1].Value)
	}
	if cells[0][1].LinkID != 1 {
		t.Fatalf("project linkID = %d, want 1", cells[0][1].LinkID)
	}
	// Vendor name at col 2.
	if cells[0][2].Value != "ContractorCo" {
		t.Fatalf("vendor cell = %q", cells[0][2].Value)
	}
	// Total at col 3.
	if cells[0][3].Value != "$500.00" {
		t.Fatalf("total cell = %q", cells[0][3].Value)
	}
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
	if cells[0][1].Value != "Project 42" {
		t.Fatalf("expected fallback project name, got %q", cells[0][1].Value)
	}
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
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if meta[0].ID != 1 {
		t.Fatalf("unexpected ID: %d", meta[0].ID)
	}
	// Name at col 1.
	if cells[0][1].Value != "HVAC Filter" {
		t.Fatalf("name cell = %q", cells[0][1].Value)
	}
	// Category at col 2.
	if cells[0][2].Value != "HVAC" {
		t.Fatalf("category cell = %q", cells[0][2].Value)
	}
	// Appliance at col 3.
	if cells[0][3].Value != "AC Unit" || cells[0][3].LinkID != 5 {
		t.Fatalf("appliance cell = %q linkID=%d", cells[0][3].Value, cells[0][3].LinkID)
	}
	// Interval at col 6.
	if cells[0][6].Value != "3 mo" {
		t.Fatalf("interval cell = %q", cells[0][6].Value)
	}
	// Log count at col 7 (drilldown).
	if cells[0][7].Value != "4" {
		t.Fatalf("log count cell = %q", cells[0][7].Value)
	}
}

func TestMaintenanceRowsNoAppliance(t *testing.T) {
	items := []data.MaintenanceItem{
		{ID: 1, Name: "Gutters", Category: data.MaintenanceCategory{Name: "Exterior"}},
	}
	_, _, cells := maintenanceRows(items, nil)
	if cells[0][3].Value != "" {
		t.Fatalf("expected empty appliance cell, got %q", cells[0][3].Value)
	}
	if cells[0][3].LinkID != 0 {
		t.Fatalf("expected zero linkID, got %d", cells[0][3].LinkID)
	}
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
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if meta[0].ID != 1 {
		t.Fatalf("unexpected ID: %d", meta[0].ID)
	}
	// Name at col 1.
	if cells[0][1].Value != "Fridge" {
		t.Fatalf("name = %q", cells[0][1].Value)
	}
	// Brand at col 2.
	if cells[0][2].Value != "Samsung" {
		t.Fatalf("brand = %q", cells[0][2].Value)
	}
	// Purchase date at col 6.
	if cells[0][6].Value != "2023-06-15" {
		t.Fatalf("purchase date = %q", cells[0][6].Value)
	}
	// Age at col 7.
	if cells[0][7].Value != "2y" {
		t.Fatalf("age = %q, want '2y'", cells[0][7].Value)
	}
	// Cost at col 9.
	if cells[0][9].Value != "$899.00" {
		t.Fatalf("cost = %q", cells[0][9].Value)
	}
	// Maint count at col 10 (drilldown).
	if cells[0][10].Value != "2" {
		t.Fatalf("maint count = %q", cells[0][10].Value)
	}
}

func TestApplianceRowsNoOptionalFields(t *testing.T) {
	now := time.Now()
	items := []data.Appliance{
		{ID: 1, Name: "Lamp"},
	}
	_, _, cells := applianceRows(items, nil, now)
	if cells[0][6].Value != "" {
		t.Fatalf("expected empty purchase date, got %q", cells[0][6].Value)
	}
	if cells[0][7].Value != "" {
		t.Fatalf("expected empty age, got %q", cells[0][7].Value)
	}
	if cells[0][9].Value != "" {
		t.Fatalf("expected empty cost, got %q", cells[0][9].Value)
	}
	if cells[0][10].Value != "" {
		t.Fatalf("expected empty maint count, got %q", cells[0][10].Value)
	}
}

func TestBuildRowsEmpty(t *testing.T) {
	rows, meta, cells := projectRows(nil)
	if len(rows) != 0 || len(meta) != 0 || len(cells) != 0 {
		t.Fatal("expected empty slices for nil input")
	}
}

func TestCellsToRow(t *testing.T) {
	cells := []cell{
		{Value: "1"},
		{Value: "hello"},
		{Value: "$100.00"},
	}
	row := cellsToRow(cells)
	if len(row) != 3 || row[0] != "1" || row[1] != "hello" || row[2] != "$100.00" {
		t.Fatalf("unexpected row: %v", row)
	}
}
