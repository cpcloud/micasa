// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenInlineInputSetsState(t *testing.T) {
	m := newTestModel()
	var field string
	m.openInlineInput(42, formVendor, "Name", "Acme", &field, nil, &vendorFormData{})

	require.NotNil(t, m.inlineInput)
	assert.Equal(t, "Name", m.inlineInput.Title)
	assert.Equal(t, uint(42), m.inlineInput.EditID)
	assert.Equal(t, formVendor, m.formKind)
	require.NotNil(t, m.editID)
	assert.Equal(t, uint(42), *m.editID)
}

func TestInlineInputEscCloses(t *testing.T) {
	m := newTestModel()
	var field string
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	sendKey(m, "esc")

	assert.Nil(t, m.inlineInput)
	assert.Equal(t, formNone, m.formKind)
	assert.Nil(t, m.editID)
}

func TestInlineInputAbsorbsKeys(t *testing.T) {
	m := newTestModel()
	var field string
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	// Keys that would normally toggle house profile or switch tabs should be absorbed.
	showHouseBefore := m.showHouse
	sendKey(m, "tab")
	assert.Equal(t, showHouseBefore, m.showHouse, "tab should be absorbed by inline input")

	// 'q' should not quit -- inline input should still be active.
	sendKey(m, "q")
	assert.NotNil(t, m.inlineInput, "inline input should still be active after pressing q")
}

func TestInlineInputTypingUpdatesValue(t *testing.T) {
	m := newTestModel()
	var field string
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	// Type some characters.
	for _, ch := range "hello" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	assert.Equal(t, "hello", m.inlineInput.Input.Value())
}

func TestInlineInputValidationBlocksSubmit(t *testing.T) {
	m := newTestModel()
	var field string
	validate := func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("name is required")
		}
		return nil
	}
	m.openInlineInput(1, formVendor, "Name", "", &field, validate, &vendorFormData{})

	// Try to submit with empty value -- should fail validation.
	sendKey(m, "enter")

	// Inline input should still be open (validation failed).
	require.NotNil(t, m.inlineInput)
	assert.Equal(t, statusError, m.status.Kind)
	assert.Contains(t, m.status.Text, "required")
}

func TestInlineInputStatusViewRendersPrompt(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	var field string
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	status := m.statusView()
	assert.Contains(t, status, "Name:")
}

func TestInlineInputPreservesExistingValue(t *testing.T) {
	m := newTestModel()
	field := "existing value"
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	assert.Equal(t, "existing value", m.inlineInput.Input.Value())
}

func TestInlineInputPlaceholder(t *testing.T) {
	m := newTestModel()
	var field string
	m.openInlineInput(1, formAppliance, "Cost", "899.00", &field, nil, &applianceFormData{})

	assert.Equal(t, "899.00", m.inlineInput.Input.Placeholder)
}

func TestInlineInputTableStaysVisible(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	var field string
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	// The model should NOT be in modeForm, so the table stays visible.
	assert.NotEqual(t, modeForm, m.mode, "inline input should not switch to modeForm")

	// buildBaseView should render the table, not a form.
	view := m.buildBaseView()
	// The table view includes column headers from the active tab.
	tab := m.activeTab()
	if tab != nil && len(tab.Specs) > 0 {
		assert.Contains(
			t,
			view,
			tab.Specs[0].Title,
			"expected table to be visible during inline input",
		)
	}
}
