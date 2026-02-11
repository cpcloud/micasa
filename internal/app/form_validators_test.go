// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
	"time"

	"github.com/cpcloud/micasa/internal/data"
)

func TestRequiredTextRejectsEmpty(t *testing.T) {
	validate := requiredText("title")
	if err := validate(""); err == nil {
		t.Fatal("expected error for empty string")
	}
	if err := validate("  "); err == nil {
		t.Fatal("expected error for whitespace-only string")
	}
}

func TestRequiredTextAcceptsNonEmpty(t *testing.T) {
	validate := requiredText("title")
	if err := validate("hello"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOptionalIntAcceptsValid(t *testing.T) {
	validate := optionalInt("months")
	for _, input := range []string{"", "0", "12", "  7  "} {
		if err := validate(input); err != nil {
			t.Fatalf("optionalInt(%q) unexpected error: %v", input, err)
		}
	}
}

func TestOptionalIntRejectsInvalid(t *testing.T) {
	validate := optionalInt("months")
	for _, input := range []string{"abc", "-1", "1.5"} {
		if err := validate(input); err == nil {
			t.Fatalf("optionalInt(%q) expected error", input)
		}
	}
}

func TestOptionalFloatAcceptsValid(t *testing.T) {
	validate := optionalFloat("bathrooms")
	for _, input := range []string{"", "0", "2.5", "  3  "} {
		if err := validate(input); err != nil {
			t.Fatalf("optionalFloat(%q) unexpected error: %v", input, err)
		}
	}
}

func TestOptionalFloatRejectsInvalid(t *testing.T) {
	validate := optionalFloat("bathrooms")
	for _, input := range []string{"abc", "-1.5"} {
		if err := validate(input); err == nil {
			t.Fatalf("optionalFloat(%q) expected error", input)
		}
	}
}

func TestOptionalDateAcceptsValid(t *testing.T) {
	validate := optionalDate("start date")
	for _, input := range []string{"", "2025-06-11"} {
		if err := validate(input); err != nil {
			t.Fatalf("optionalDate(%q) unexpected error: %v", input, err)
		}
	}
}

func TestOptionalDateRejectsInvalid(t *testing.T) {
	validate := optionalDate("start date")
	for _, input := range []string{"06/11/2025", "not-a-date"} {
		if err := validate(input); err == nil {
			t.Fatalf("optionalDate(%q) expected error", input)
		}
	}
}

func TestOptionalMoneyAcceptsValid(t *testing.T) {
	validate := optionalMoney("budget")
	for _, input := range []string{"", "100", "1250.00", "$5,000.50"} {
		if err := validate(input); err != nil {
			t.Fatalf("optionalMoney(%q) unexpected error: %v", input, err)
		}
	}
}

func TestOptionalMoneyRejectsInvalid(t *testing.T) {
	validate := optionalMoney("budget")
	if err := validate("abc"); err == nil {
		t.Fatal("expected error for invalid money")
	}
}

func TestRequiredMoneyAcceptsValid(t *testing.T) {
	validate := requiredMoney("total")
	if err := validate("1250.00"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequiredMoneyRejectsEmpty(t *testing.T) {
	validate := requiredMoney("total")
	if err := validate(""); err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestIntToString(t *testing.T) {
	if got := intToString(0); got != "" {
		t.Fatalf("intToString(0) = %q, want empty", got)
	}
	if got := intToString(42); got != "42" {
		t.Fatalf("intToString(42) = %q, want 42", got)
	}
}

func TestProjectFormValues(t *testing.T) {
	budget := int64(500000)
	start := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	project := data.Project{
		Title:         "Kitchen Remodel",
		ProjectTypeID: 1,
		Status:        data.ProjectStatusInProgress,
		BudgetCents:   &budget,
		StartDate:     &start,
		Description:   "Full gut renovation",
	}
	got := projectFormValues(project)
	if got.Title != "Kitchen Remodel" {
		t.Fatalf("Title = %q", got.Title)
	}
	if got.Budget != "$5,000.00" {
		t.Fatalf("Budget = %q", got.Budget)
	}
	if got.StartDate != "2025-03-01" {
		t.Fatalf("StartDate = %q", got.StartDate)
	}
	if got.Actual != "" {
		t.Fatalf("Actual should be empty, got %q", got.Actual)
	}
}

func TestVendorFormValues(t *testing.T) {
	vendor := data.Vendor{
		Name:        "HVAC Pros",
		ContactName: "Alice",
		Email:       "alice@hvac.com",
		Phone:       "555-1234",
		Website:     "https://hvac.com",
	}
	got := vendorFormValues(vendor)
	if got.Name != "HVAC Pros" || got.ContactName != "Alice" {
		t.Fatalf("got %+v", got)
	}
}

func TestQuoteFormValues(t *testing.T) {
	labor := int64(10000)
	quote := data.Quote{
		ProjectID:  1,
		TotalCents: 50000,
		LaborCents: &labor,
		Vendor:     data.Vendor{Name: "ContractorCo"},
	}
	got := quoteFormValues(quote)
	if got.Total != "$500.00" {
		t.Fatalf("Total = %q", got.Total)
	}
	if got.Labor != "$100.00" {
		t.Fatalf("Labor = %q", got.Labor)
	}
	if got.Materials != "" {
		t.Fatalf("Materials should be empty, got %q", got.Materials)
	}
	if got.VendorName != "ContractorCo" {
		t.Fatalf("VendorName = %q", got.VendorName)
	}
}

func TestMaintenanceFormValues(t *testing.T) {
	appID := uint(3)
	item := data.MaintenanceItem{
		Name:           "HVAC Filter",
		CategoryID:     1,
		ApplianceID:    &appID,
		IntervalMonths: 3,
	}
	got := maintenanceFormValues(item)
	if got.Name != "HVAC Filter" {
		t.Fatalf("Name = %q", got.Name)
	}
	if got.ApplianceID != 3 {
		t.Fatalf("ApplianceID = %d, want 3", got.ApplianceID)
	}
	if got.IntervalMonths != "3" {
		t.Fatalf("IntervalMonths = %q", got.IntervalMonths)
	}
}

func TestMaintenanceFormValuesNoAppliance(t *testing.T) {
	item := data.MaintenanceItem{
		Name:       "Smoke Detectors",
		CategoryID: 1,
	}
	got := maintenanceFormValues(item)
	if got.ApplianceID != 0 {
		t.Fatalf("ApplianceID = %d, want 0", got.ApplianceID)
	}
}

func TestApplianceFormValues(t *testing.T) {
	cost := int64(89900)
	purchase := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)
	appliance := data.Appliance{
		Name:         "Fridge",
		Brand:        "Samsung",
		ModelNumber:  "RF28R7351SR",
		PurchaseDate: &purchase,
		CostCents:    &cost,
	}
	got := applianceFormValues(appliance)
	if got.Name != "Fridge" || got.Brand != "Samsung" {
		t.Fatalf("got %+v", got)
	}
	if got.PurchaseDate != "2023-06-15" {
		t.Fatalf("PurchaseDate = %q", got.PurchaseDate)
	}
	if got.Cost != "$899.00" {
		t.Fatalf("Cost = %q", got.Cost)
	}
}

func TestHouseFormValues(t *testing.T) {
	profile := data.HouseProfile{
		Nickname:  "Home",
		YearBuilt: 1995,
		Bedrooms:  3,
		Bathrooms: 2.5,
	}
	got := houseFormValues(profile)
	if got.Nickname != "Home" {
		t.Fatalf("Nickname = %q", got.Nickname)
	}
	if got.YearBuilt != "1995" {
		t.Fatalf("YearBuilt = %q", got.YearBuilt)
	}
	if got.Bedrooms != "3" {
		t.Fatalf("Bedrooms = %q", got.Bedrooms)
	}
	if got.Bathrooms != "2.5" {
		t.Fatalf("Bathrooms = %q", got.Bathrooms)
	}
}

func TestServiceLogFormValues(t *testing.T) {
	cost := int64(15000)
	vendorID := uint(1)
	entry := data.ServiceLogEntry{
		ServicedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		VendorID:   &vendorID,
		CostCents:  &cost,
		Notes:      "replaced filter",
	}
	got := serviceLogFormValues(entry)
	if got.ServicedAt != "2025-01-15" {
		t.Fatalf("ServicedAt = %q", got.ServicedAt)
	}
	if got.Cost != "$150.00" {
		t.Fatalf("Cost = %q", got.Cost)
	}
	if got.VendorID != 1 {
		t.Fatalf("VendorID = %d, want 1", got.VendorID)
	}
}

func TestServiceLogFormValuesNoVendor(t *testing.T) {
	entry := data.ServiceLogEntry{
		ServicedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	got := serviceLogFormValues(entry)
	if got.VendorID != 0 {
		t.Fatalf("VendorID = %d, want 0", got.VendorID)
	}
	if got.Cost != "" {
		t.Fatalf("Cost = %q, want empty", got.Cost)
	}
}
