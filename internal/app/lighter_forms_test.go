// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.Containsf(t, view, want, "add project form should contain %q", want)
	}
	// Optional fields should be absent.
	for _, absent := range []string{"Budget", "Actual cost", "Start date", "End date", "Description"} {
		assert.NotContainsf(t, view, absent, "add project form should NOT contain %q", absent)
	}
}

func TestEditProjectFormHasMoreFieldsThanAdd(t *testing.T) {
	m := newTestModelWithStore(t)
	// Create a project so we can edit it.
	m.startProjectForm()
	m.form.Init()
	values, ok := m.formData.(*projectFormData)
	require.True(t, ok, "unexpected form data type")
	values.Title = testProjectTitle
	require.NoError(t, m.submitProjectForm())
	m.exitForm()
	m.reloadAll()

	require.NoError(t, m.startEditProjectForm(1))
	// The edit form's first group includes Budget and Actual cost,
	// which are absent from the add form.
	view := formFieldLabels(m)
	for _, want := range []string{"Title", "Status", "Budget", "Actual cost"} {
		assert.Containsf(t, view, want, "edit project form should contain %q", want)
	}
}

func TestAddVendorFormHasOnlyName(t *testing.T) {
	m := newTestModelWithStore(t)
	m.startVendorForm()

	view := formFieldLabels(m)
	assert.Contains(t, view, "Name")
	for _, absent := range []string{"Contact name", "Email", "Phone", "Website"} {
		assert.NotContainsf(t, view, absent, "add vendor form should NOT contain %q", absent)
	}
}

func TestEditVendorFormHasAllFields(t *testing.T) {
	m := newTestModelWithStore(t)
	m.startVendorForm()
	m.form.Init()
	values, ok := m.formData.(*vendorFormData)
	require.True(t, ok, "unexpected form data type")
	values.Name = "Test Vendor"
	require.NoError(t, m.submitVendorForm())
	m.exitForm()
	m.reloadAll()

	require.NoError(t, m.startEditVendorForm(1))
	view := formFieldLabels(m)
	for _, want := range []string{"Name", "Contact name", "Email", "Phone", "Website"} {
		assert.Containsf(t, view, want, "edit vendor form should contain %q", want)
	}
}

func TestAddApplianceFormHasOnlyName(t *testing.T) {
	m := newTestModelWithStore(t)
	m.startApplianceForm()

	view := formFieldLabels(m)
	assert.Contains(t, view, "Name")
	for _, absent := range []string{"Brand", "Model number", "Serial number", "Location", "Purchase date", "Warranty expiry", "Cost"} {
		assert.NotContainsf(t, view, absent, "add appliance form should NOT contain %q", absent)
	}
}

func TestAddMaintenanceFormHasOnlyEssentialFields(t *testing.T) {
	m := newTestModelWithStore(t)
	m.startMaintenanceForm()

	view := formFieldLabels(m)
	for _, want := range []string{"Item", "Category", "Interval months"} {
		assert.Containsf(t, view, want, "add maintenance form should contain %q", want)
	}
	for _, absent := range []string{"Manual URL", "Manual notes", "Cost", "Last serviced"} {
		assert.NotContainsf(t, view, absent, "add maintenance form should NOT contain %q", absent)
	}
}

func TestAddQuoteFormHasOnlyEssentialFields(t *testing.T) {
	m := newTestModelWithStore(t)
	// Need a project first.
	m.startProjectForm()
	m.form.Init()
	values, ok := m.formData.(*projectFormData)
	require.True(t, ok, "unexpected form data type")
	values.Title = testProjectTitle
	require.NoError(t, m.submitProjectForm())
	m.exitForm()
	m.reloadAll()

	require.NoError(t, m.startQuoteForm())
	view := formFieldLabels(m)
	for _, want := range []string{"Project", "Vendor name", "Total"} {
		assert.Containsf(t, view, want, "add quote form should contain %q", want)
	}
	for _, absent := range []string{"Contact name", "Email", "Phone", "Labor", "Materials", "Other", "Received date"} {
		assert.NotContainsf(t, view, absent, "add quote form should NOT contain %q", absent)
	}
}

func TestAddServiceLogFormHasOnlyEssentialFields(t *testing.T) {
	m := newTestModelWithStore(t)
	require.NoError(t, m.startServiceLogForm(0))
	view := formFieldLabels(m)
	for _, want := range []string{"Date serviced", "Performed by"} {
		assert.Containsf(t, view, want, "add service log form should contain %q", want)
	}
	for _, absent := range []string{"Cost", "Notes"} {
		assert.NotContainsf(t, view, absent, "add service log form should NOT contain %q", absent)
	}
}
