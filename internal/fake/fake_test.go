// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package fake

import (
	"testing"
)

func TestNewDeterministicSeed(t *testing.T) {
	h1 := New(42)
	h2 := New(42)

	v1 := h1.Vendor()
	v2 := h2.Vendor()
	if v1.Name != v2.Name {
		t.Errorf("same seed produced different vendors: %q vs %q", v1.Name, v2.Name)
	}
}

func TestHouseProfile(t *testing.T) {
	h := New(1)
	house := h.HouseProfile()

	if house.Nickname == "" {
		t.Error("expected non-empty nickname")
	}
	if house.City == "" {
		t.Error("expected non-empty city")
	}
	if house.YearBuilt < 1920 || house.YearBuilt > 2024 {
		t.Errorf("year built %d out of range", house.YearBuilt)
	}
	if house.SquareFeet < 800 || house.SquareFeet > 4500 {
		t.Errorf("sqft %d out of range", house.SquareFeet)
	}
	if house.Bedrooms < 1 || house.Bedrooms > 6 {
		t.Errorf("bedrooms %d out of range", house.Bedrooms)
	}
	if house.InsuranceRenewal == nil {
		t.Error("expected insurance renewal date")
	}
}

func TestVendor(t *testing.T) {
	h := New(2)
	v := h.Vendor()

	if v.Name == "" {
		t.Error("expected non-empty vendor name")
	}
	if v.ContactName == "" {
		t.Error("expected non-empty contact name")
	}
	if v.Phone == "" {
		t.Error("expected non-empty phone")
	}
	if v.Email == "" {
		t.Error("expected non-empty email")
	}
}

func TestVendorForTrade(t *testing.T) {
	h := New(3)
	v := h.VendorForTrade("Plumbing")

	if v.Name == "" {
		t.Error("expected non-empty name")
	}
}

func TestProject(t *testing.T) {
	h := New(4)

	for _, typeName := range ProjectTypes() {
		p := h.Project(typeName)
		if p.Title == "" {
			t.Errorf("empty title for type %q", typeName)
		}
		if p.TypeName != typeName {
			t.Errorf("type name = %q, want %q", p.TypeName, typeName)
		}
		if p.Description == "" {
			t.Errorf("empty description for type %q", typeName)
		}
	}
}

func TestProjectCompletedHasEndDateAndActual(t *testing.T) {
	for seed := uint64(0); seed < 100; seed++ {
		h := New(seed)
		p := h.Project("Plumbing")
		if p.Status == StatusCompleted {
			if p.EndDate == nil {
				t.Error("completed project should have end date")
			}
			if p.ActualCents == nil {
				t.Error("completed project should have actual cost")
			}
			return
		}
	}
	t.Skip("never hit completed status in 100 seeds")
}

func TestProjectUnknownType(t *testing.T) {
	h := New(5)
	p := h.Project("Unknown")
	if p.Title == "" {
		t.Error("expected fallback title for unknown type")
	}
}

func TestAppliance(t *testing.T) {
	h := New(6)
	a := h.Appliance()

	if a.Name == "" {
		t.Error("expected non-empty name")
	}
	if a.Brand == "" {
		t.Error("expected non-empty brand")
	}
	if a.ModelNumber == "" {
		t.Error("expected non-empty model number")
	}
	if a.SerialNumber == "" {
		t.Error("expected non-empty serial number")
	}
	if a.Location == "" {
		t.Error("expected non-empty location")
	}
	if a.PurchaseDate == nil {
		t.Error("expected purchase date")
	}
	if a.CostCents == nil {
		t.Error("expected cost")
	}
}

func TestMaintenanceItem(t *testing.T) {
	h := New(7)

	for _, cat := range MaintenanceCategories() {
		m := h.MaintenanceItem(cat)
		if m.Name == "" {
			t.Errorf("empty name for category %q", cat)
		}
		if m.IntervalMonths <= 0 {
			t.Errorf("interval %d for category %q", m.IntervalMonths, cat)
		}
	}
}

func TestMaintenanceItemUnknownCategory(t *testing.T) {
	h := New(8)
	m := h.MaintenanceItem("Unknown")
	if m.Name == "" {
		t.Error("expected fallback name")
	}
	if m.IntervalMonths != 12 {
		t.Errorf("expected 12-month interval for unknown, got %d", m.IntervalMonths)
	}
}

func TestServiceLogEntry(t *testing.T) {
	h := New(9)
	e := h.ServiceLogEntry()

	if e.ServicedAt.IsZero() {
		t.Error("expected non-zero serviced at")
	}
	if e.CostCents == nil {
		t.Error("expected cost")
	}
	if e.Notes == "" {
		t.Error("expected notes")
	}
}

func TestQuote(t *testing.T) {
	h := New(10)
	q := h.Quote()

	if q.TotalCents <= 0 {
		t.Error("expected positive total")
	}
	if q.LaborCents == nil || q.MaterialsCents == nil {
		t.Error("expected labor and materials breakdown")
	}
	if *q.LaborCents+*q.MaterialsCents != q.TotalCents {
		t.Error("labor + materials should equal total")
	}
	if q.ReceivedDate == nil {
		t.Error("expected received date")
	}
}

func TestVarietyAcrossSeeds(t *testing.T) {
	names := map[string]bool{}
	for seed := uint64(0); seed < 20; seed++ {
		h := New(seed)
		v := h.Vendor()
		names[v.Name] = true
	}
	if len(names) < 10 {
		t.Errorf("expected variety, got only %d unique names from 20 seeds", len(names))
	}
}

func TestVendorTrades(t *testing.T) {
	trades := VendorTrades()
	if len(trades) == 0 {
		t.Error("expected non-empty trades list")
	}
}

func TestIntN(t *testing.T) {
	h := New(42)
	for i := 0; i < 100; i++ {
		v := h.IntN(5)
		if v < 0 || v >= 5 {
			t.Errorf("IntN(5) = %d, out of range", v)
		}
	}
}
