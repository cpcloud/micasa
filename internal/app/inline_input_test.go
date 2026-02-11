// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOpenInlineInputSetsState(t *testing.T) {
	m := newTestModel()
	var field string
	m.openInlineInput(42, formVendor, "Name", "Acme", &field, nil, &vendorFormData{})

	if m.inlineInput == nil {
		t.Fatal("expected inlineInput to be set")
	}
	if m.inlineInput.Title != "Name" {
		t.Fatalf("expected title 'Name', got %q", m.inlineInput.Title)
	}
	if m.inlineInput.EditID != 42 {
		t.Fatalf("expected editID 42, got %d", m.inlineInput.EditID)
	}
	if m.formKind != formVendor {
		t.Fatalf("expected formKind formVendor, got %d", m.formKind)
	}
	if m.editID == nil || *m.editID != 42 {
		t.Fatal("expected model editID to be set to 42")
	}
}

func TestInlineInputEscCloses(t *testing.T) {
	m := newTestModel()
	var field string
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	sendKey(m, "esc")

	if m.inlineInput != nil {
		t.Fatal("expected inlineInput to be nil after esc")
	}
	if m.formKind != formNone {
		t.Fatalf("expected formKind reset, got %d", m.formKind)
	}
	if m.editID != nil {
		t.Fatal("expected editID to be nil after esc")
	}
}

func TestInlineInputAbsorbsKeys(t *testing.T) {
	m := newTestModel()
	var field string
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	// Keys that would normally switch tabs or enter edit mode should be absorbed.
	initialActive := m.active
	sendKey(m, "tab")
	if m.active != initialActive {
		t.Fatal("tab should be absorbed by inline input")
	}

	// 'q' should not quit -- inline input should still be active.
	sendKey(m, "q")
	if m.inlineInput == nil {
		t.Fatal("inline input should still be active after pressing q")
	}
}

func TestInlineInputTypingUpdatesValue(t *testing.T) {
	m := newTestModel()
	var field string
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	// Type some characters.
	for _, ch := range "hello" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	got := m.inlineInput.Input.Value()
	if got != "hello" {
		t.Fatalf("expected input value 'hello', got %q", got)
	}
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
	if m.inlineInput == nil {
		t.Fatal("expected inlineInput to remain open after validation failure")
	}
	if m.status.Kind != statusError {
		t.Fatalf("expected error status after validation failure, got %v", m.status.Kind)
	}
	if !strings.Contains(m.status.Text, "required") {
		t.Fatalf("expected error about 'required', got %q", m.status.Text)
	}
}

func TestInlineInputStatusViewRendersPrompt(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	var field string
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	status := m.statusView()
	if !strings.Contains(status, "Name:") {
		t.Fatalf("expected status to contain 'Name:', got %q", status)
	}
}

func TestInlineInputPreservesExistingValue(t *testing.T) {
	m := newTestModel()
	field := "existing value"
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	got := m.inlineInput.Input.Value()
	if got != "existing value" {
		t.Fatalf("expected input to show existing value, got %q", got)
	}
}

func TestInlineInputPlaceholder(t *testing.T) {
	m := newTestModel()
	var field string
	m.openInlineInput(1, formAppliance, "Cost", "899.00", &field, nil, &applianceFormData{})

	placeholder := m.inlineInput.Input.Placeholder
	if placeholder != "899.00" {
		t.Fatalf("expected placeholder '899.00', got %q", placeholder)
	}
}

func TestInlineInputTableStaysVisible(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	var field string
	m.openInlineInput(1, formVendor, "Name", "", &field, nil, &vendorFormData{})

	// The model should NOT be in modeForm, so the table stays visible.
	if m.mode == modeForm {
		t.Fatal("inline input should not switch to modeForm")
	}

	// buildBaseView should render the table, not a form.
	view := m.buildBaseView()
	// The table view includes column headers from the active tab.
	tab := m.activeTab()
	if tab != nil && len(tab.Specs) > 0 {
		if !strings.Contains(view, tab.Specs[0].Title) {
			t.Fatal("expected table to be visible during inline input")
		}
	}
}
