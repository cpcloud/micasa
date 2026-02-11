// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"strings"
	"testing"
)

func TestInlineEditProjectTextColumnOpensInlineInput(t *testing.T) {
	m := newTestModelWithStore(t)
	// Create a project.
	m.startProjectForm()
	m.form.Init()
	values, ok := m.formData.(*projectFormData)
	if !ok {
		t.Fatal("unexpected form data type")
	}
	values.Title = "Test Project"
	if err := m.submitProjectForm(); err != nil {
		t.Fatalf("create project: %v", err)
	}
	m.exitForm()
	m.reloadAll()

	// Inline edit the Title column (col 2) -- should open inline input.
	if err := m.inlineEditProject(1, 2); err != nil {
		t.Fatalf("inlineEditProject: %v", err)
	}
	if m.inlineInput == nil {
		t.Fatal("expected inline input for text column (Title)")
	}
	if m.inlineInput.Title != "Title" {
		t.Fatalf("expected title 'Title', got %q", m.inlineInput.Title)
	}
	if m.mode == modeForm {
		t.Fatal("inline input should not switch to modeForm")
	}
}

func TestInlineEditProjectSelectColumnOpensFormOverlay(t *testing.T) {
	m := newTestModelWithStore(t)
	m.startProjectForm()
	m.form.Init()
	values, ok := m.formData.(*projectFormData)
	if !ok {
		t.Fatal("unexpected form data type")
	}
	values.Title = "Test Project"
	if err := m.submitProjectForm(); err != nil {
		t.Fatalf("create project: %v", err)
	}
	m.exitForm()
	m.reloadAll()

	// Inline edit the Status column (col 3) -- should open form overlay.
	if err := m.inlineEditProject(1, 3); err != nil {
		t.Fatalf("inlineEditProject: %v", err)
	}
	if m.inlineInput != nil {
		t.Fatal("select column should NOT open inline input")
	}
	if m.mode != modeForm {
		t.Fatal("select column should open form overlay")
	}
}

func TestInlineEditVendorTextColumnsUseInlineInput(t *testing.T) {
	m := newTestModelWithStore(t)
	m.startVendorForm()
	m.form.Init()
	values, ok := m.formData.(*vendorFormData)
	if !ok {
		t.Fatal("unexpected form data type")
	}
	values.Name = "Test Vendor"
	if err := m.submitVendorForm(); err != nil {
		t.Fatalf("create vendor: %v", err)
	}
	m.exitForm()
	m.reloadAll()

	// All editable vendor columns are text, so they should all use inline input.
	cases := []struct {
		col   int
		title string
	}{
		{1, "Name"},
		{2, "Contact name"},
		{3, "Email"},
		{4, "Phone"},
		{5, "Website"},
	}
	for _, tc := range cases {
		m.closeInlineInput()
		if err := m.inlineEditVendor(1, tc.col); err != nil {
			t.Fatalf("inlineEditVendor col %d: %v", tc.col, err)
		}
		if m.inlineInput == nil {
			t.Fatalf("col %d (%s) should open inline input", tc.col, tc.title)
		}
		if m.inlineInput.Title != tc.title {
			t.Fatalf("col %d: expected title %q, got %q", tc.col, tc.title, m.inlineInput.Title)
		}
	}
}

func TestInlineEditAppliaceDateColumnOpensCalendar(t *testing.T) {
	m := newTestModelWithStore(t)
	m.startApplianceForm()
	m.form.Init()
	values, ok := m.formData.(*applianceFormData)
	if !ok {
		t.Fatal("unexpected form data type")
	}
	values.Name = "Test Fridge"
	if err := m.submitApplianceForm(); err != nil {
		t.Fatalf("create appliance: %v", err)
	}
	m.exitForm()
	m.reloadAll()

	// Purchase date column (col 6) should open calendar picker.
	if err := m.inlineEditAppliance(1, 6); err != nil {
		t.Fatalf("inlineEditAppliance: %v", err)
	}
	if m.calendar == nil {
		t.Fatal("date column should open calendar picker")
	}
	if m.inlineInput != nil {
		t.Fatal("date column should NOT open inline input")
	}
}

func TestEditKeyDispatchesInlineEditInEditMode(t *testing.T) {
	m := newTestModelWithStore(t)
	// Create a vendor so there's data to edit.
	m.startVendorForm()
	m.form.Init()
	values, ok := m.formData.(*vendorFormData)
	if !ok {
		t.Fatal("unexpected form data type")
	}
	values.Name = "Test Vendor"
	if err := m.submitVendorForm(); err != nil {
		t.Fatalf("create vendor: %v", err)
	}
	m.exitForm()
	m.reloadAll()

	// Switch to vendor tab.
	for i, tab := range m.tabs {
		if tab.Kind == tabVendors {
			m.active = i
			break
		}
	}
	if err := m.reloadActiveTab(); err != nil {
		t.Fatalf("reload vendor tab: %v", err)
	}
	tab := m.activeTab()
	if tab == nil || len(tab.Rows) == 0 {
		t.Skip("no vendor rows to test")
	}

	// Ensure table cursor is on the first row.
	tab.Table.SetCursor(0)

	// Enter edit mode and position cursor on Name column.
	m.enterEditMode()
	tab.ColCursor = 1 // Name column

	// Press 'e' to trigger inline edit.
	sendKey(m, "e")

	// Should have opened inline input for the Name field.
	if m.inlineInput == nil && m.mode != modeForm {
		t.Fatal("pressing 'e' in edit mode should open inline input or form for the current cell")
	}

	// Verify the status bar shows the inline prompt.
	status := m.statusView()
	if m.inlineInput != nil && !strings.Contains(status, "Name") {
		t.Fatalf("expected status bar to show 'Name' prompt, got %q", status)
	}
}
