// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"strings"
	"testing"
)

// formFieldLabels initializes the form and returns the rendered view text.
// Callers check for presence/absence of field labels.
func formFieldLabels(m *Model) string {
	if m.form == nil {
		return ""
	}
	m.form.Init()
	return m.form.View()
}

func TestAddProjectFormHasOnlyEssentialFields(t *testing.T) {
	m := newTestModelWithStore(t)
	m.startProjectForm()

	view := formFieldLabels(m)
	// Essential fields should be present.
	for _, want := range []string{"Title", "Project type", "Status"} {
		if !strings.Contains(view, want) {
			t.Errorf("add project form should contain %q", want)
		}
	}
	// Optional fields should be absent.
	for _, absent := range []string{"Budget", "Actual cost", "Start date", "End date", "Description"} {
		if strings.Contains(view, absent) {
			t.Errorf("add project form should NOT contain %q", absent)
		}
	}
}

func TestEditProjectFormHasMoreFieldsThanAdd(t *testing.T) {
	m := newTestModelWithStore(t)
	// Create a project so we can edit it.
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

	if err := m.startEditProjectForm(1); err != nil {
		t.Fatalf("start edit: %v", err)
	}
	// The edit form's first group includes Budget and Actual cost,
	// which are absent from the add form.
	view := formFieldLabels(m)
	for _, want := range []string{"Title", "Status", "Budget", "Actual cost"} {
		if !strings.Contains(view, want) {
			t.Errorf("edit project form should contain %q", want)
		}
	}
}

func TestAddVendorFormHasOnlyName(t *testing.T) {
	m := newTestModelWithStore(t)
	m.startVendorForm()

	view := formFieldLabels(m)
	if !strings.Contains(view, "Name") {
		t.Error("add vendor form should contain 'Name'")
	}
	for _, absent := range []string{"Contact name", "Email", "Phone", "Website"} {
		if strings.Contains(view, absent) {
			t.Errorf("add vendor form should NOT contain %q", absent)
		}
	}
}

func TestEditVendorFormHasAllFields(t *testing.T) {
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

	if err := m.startEditVendorForm(1); err != nil {
		t.Fatalf("start edit: %v", err)
	}
	view := formFieldLabels(m)
	for _, want := range []string{"Name", "Contact name", "Email", "Phone", "Website"} {
		if !strings.Contains(view, want) {
			t.Errorf("edit vendor form should contain %q", want)
		}
	}
}

func TestAddApplianceFormHasOnlyName(t *testing.T) {
	m := newTestModelWithStore(t)
	m.startApplianceForm()

	view := formFieldLabels(m)
	if !strings.Contains(view, "Name") {
		t.Error("add appliance form should contain 'Name'")
	}
	for _, absent := range []string{"Brand", "Model number", "Serial number", "Location", "Purchase date", "Warranty expiry", "Cost"} {
		if strings.Contains(view, absent) {
			t.Errorf("add appliance form should NOT contain %q", absent)
		}
	}
}

func TestAddMaintenanceFormHasOnlyEssentialFields(t *testing.T) {
	m := newTestModelWithStore(t)
	m.startMaintenanceForm()

	view := formFieldLabels(m)
	for _, want := range []string{"Item", "Category", "Interval months"} {
		if !strings.Contains(view, want) {
			t.Errorf("add maintenance form should contain %q", want)
		}
	}
	for _, absent := range []string{"Manual URL", "Manual notes", "Cost", "Last serviced"} {
		if strings.Contains(view, absent) {
			t.Errorf("add maintenance form should NOT contain %q", absent)
		}
	}
}

func TestAddQuoteFormHasOnlyEssentialFields(t *testing.T) {
	m := newTestModelWithStore(t)
	// Need a project first.
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

	if err := m.startQuoteForm(); err != nil {
		t.Fatalf("start quote form: %v", err)
	}
	view := formFieldLabels(m)
	for _, want := range []string{"Project", "Vendor name", "Total"} {
		if !strings.Contains(view, want) {
			t.Errorf("add quote form should contain %q", want)
		}
	}
	for _, absent := range []string{"Contact name", "Email", "Phone", "Labor", "Materials", "Other", "Received date"} {
		if strings.Contains(view, absent) {
			t.Errorf("add quote form should NOT contain %q", absent)
		}
	}
}

func TestAddServiceLogFormHasOnlyEssentialFields(t *testing.T) {
	m := newTestModelWithStore(t)
	if err := m.startServiceLogForm(0); err != nil {
		t.Fatalf("start service log form: %v", err)
	}
	view := formFieldLabels(m)
	for _, want := range []string{"Date serviced", "Performed by"} {
		if !strings.Contains(view, want) {
			t.Errorf("add service log form should contain %q", want)
		}
	}
	for _, absent := range []string{"Cost", "Notes"} {
		if strings.Contains(view, absent) {
			t.Errorf("add service log form should NOT contain %q", absent)
		}
	}
}
