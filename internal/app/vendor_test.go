// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/cpcloud/micasa/internal/data"
)

func TestVendorTabExists(t *testing.T) {
	m := newTestModel()
	found := false
	for _, tab := range m.tabs {
		if tab.Kind == tabVendors {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected Vendors tab to exist")
	}
}

func TestVendorTabIndex(t *testing.T) {
	idx := tabIndex(tabVendors)
	if idx != 4 {
		t.Fatalf("expected tabIndex(tabVendors)=4, got %d", idx)
	}
}

func TestVendorTabKindString(t *testing.T) {
	if tabVendors.String() != "Vendors" {
		t.Fatalf("expected Vendors, got %s", tabVendors.String())
	}
}

func TestVendorColumnSpecs(t *testing.T) {
	specs := vendorColumnSpecs()
	if len(specs) != 8 {
		t.Fatalf("expected 8 columns, got %d", len(specs))
	}
	titles := make([]string, len(specs))
	for i, s := range specs {
		titles[i] = s.Title
	}
	expected := []string{"ID", "Name", "Contact", "Email", "Phone", "Website", "Quotes", "Jobs"}
	for i, want := range expected {
		if titles[i] != want {
			t.Errorf("column %d: expected %q, got %q", i, want, titles[i])
		}
	}
}

func TestVendorRows(t *testing.T) {
	rows, meta, cells := vendorRows(
		sampleVendors(),
		map[uint]int{1: 3},
		map[uint]int{2: 5},
	)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if meta[0].ID != 1 || meta[1].ID != 2 {
		t.Fatalf("unexpected meta IDs: %+v", meta)
	}
	// Vendor 1 has 3 quotes, 0 jobs.
	if cells[0][6].Value != "3" {
		t.Errorf("vendor 1 quotes: expected '3', got %q", cells[0][6].Value)
	}
	if cells[0][7].Value != "" {
		t.Errorf("vendor 1 jobs: expected '', got %q", cells[0][7].Value)
	}
	// Vendor 2 has 0 quotes, 5 jobs.
	if cells[1][6].Value != "" {
		t.Errorf("vendor 2 quotes: expected '', got %q", cells[1][6].Value)
	}
	if cells[1][7].Value != "5" {
		t.Errorf("vendor 2 jobs: expected '5', got %q", cells[1][7].Value)
	}
}

func TestVendorHandlerFormKind(t *testing.T) {
	h := vendorHandler{}
	if h.FormKind() != formVendor {
		t.Fatalf("expected formVendor, got %d", h.FormKind())
	}
}

func TestVendorHandlerDeleteRestore(t *testing.T) {
	m := newTestModelWithStore(t)
	h := vendorHandler{}
	if err := m.store.CreateVendor(data.Vendor{Name: "Test Co"}); err != nil {
		t.Fatalf("CreateVendor: %v", err)
	}
	vendors, _ := m.store.ListVendors(false)
	id := vendors[0].ID

	if err := h.Delete(m.store, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	vendors, _ = m.store.ListVendors(false)
	if len(vendors) != 0 {
		t.Fatalf("expected 0 vendors after delete, got %d", len(vendors))
	}
	if err := h.Restore(m.store, id); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	vendors, _ = m.store.ListVendors(false)
	if len(vendors) != 1 {
		t.Fatalf("expected 1 vendor after restore, got %d", len(vendors))
	}
}

func TestVendorTabNavigable(t *testing.T) {
	m := newTestModel()
	// Navigate to vendor tab.
	m.active = tabIndex(tabVendors)
	tab := m.activeTab()
	if tab == nil {
		t.Fatal("vendor tab is nil")
	}
	if tab.Kind != tabVendors {
		t.Fatalf("expected tabVendors, got %v", tab.Kind)
	}
}

func TestVendorColumnSpecKinds(t *testing.T) {
	specs := vendorColumnSpecs()
	// ID (0), Quotes (6), Jobs (7) are readonly.
	readonlyCols := []int{0, 6, 7}
	for _, col := range readonlyCols {
		if specs[col].Kind != cellReadonly {
			t.Errorf(
				"col %d (%s): expected cellReadonly, got %v",
				col,
				specs[col].Title,
				specs[col].Kind,
			)
		}
	}
	// Editable columns: Name, Contact, Email, Phone, Website.
	editableCols := []int{1, 2, 3, 4, 5}
	for _, col := range editableCols {
		if specs[col].Kind != cellText {
			t.Errorf(
				"col %d (%s): expected cellText (editable), got %v",
				col,
				specs[col].Title,
				specs[col].Kind,
			)
		}
	}
}

func TestQuoteVendorColumnLinksToVendorTab(t *testing.T) {
	specs := quoteColumnSpecs()
	vendorSpec := specs[2] // Vendor column
	if vendorSpec.Link == nil {
		t.Fatal("expected Vendor column to have a Link")
	}
	if vendorSpec.Link.TargetTab != tabVendors {
		t.Fatalf("expected link to tabVendors, got %v", vendorSpec.Link.TargetTab)
	}
	// Relation field removed -- TargetTab is sufficient
}

func TestVendorFormData(t *testing.T) {
	v := vendorFormValues(sampleVendors()[0])
	if v.Name != "Acme Plumbing" {
		t.Fatalf("expected 'Acme Plumbing', got %q", v.Name)
	}
	if v.ContactName != "Jo Smith" {
		t.Fatalf("expected 'Jo Smith', got %q", v.ContactName)
	}
	if v.Email != "jo@example.com" {
		t.Fatalf("expected 'jo@example.com', got %q", v.Email)
	}
}

func sampleVendors() []data.Vendor {
	return []data.Vendor{
		{
			ID:          1,
			Name:        "Acme Plumbing",
			ContactName: "Jo Smith",
			Email:       "jo@example.com",
			Phone:       "555-0142",
		},
		{ID: 2, Name: "Sparks Electric", ContactName: "Tom", Phone: "555-0231"},
	}
}
