// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import "testing"

func TestAllTabsHaveHandlers(t *testing.T) {
	m := newTestModel()
	for i, tab := range m.tabs {
		if tab.Handler == nil {
			t.Fatalf("tab %d (%s) has nil handler", i, tab.Name)
		}
	}
}

func TestHandlerForFormKind(t *testing.T) {
	m := newTestModel()
	cases := []struct {
		kind FormKind
		name string
	}{
		{formProject, "project"},
		{formQuote, "quote"},
		{formMaintenance, "maintenance"},
		{formAppliance, "appliance"},
	}

	for _, tc := range cases {
		handler := m.handlerForFormKind(tc.kind)
		if handler == nil {
			t.Fatalf("expected handler for %s, got nil", tc.name)
		}
		if handler.FormKind() != tc.kind {
			t.Fatalf(
				"handler for %s returned FormKind %d, want %d",
				tc.name,
				handler.FormKind(),
				tc.kind,
			)
		}
	}
}

func TestHandlerForFormKindHouseReturnsNil(t *testing.T) {
	m := newTestModel()
	if h := m.handlerForFormKind(formHouse); h != nil {
		t.Fatal("expected nil handler for formHouse")
	}
}

func TestHandlerForFormKindUnknownReturnsNil(t *testing.T) {
	m := newTestModel()
	if h := m.handlerForFormKind(formNone); h != nil {
		t.Fatal("expected nil handler for formNone")
	}
}

func TestHandlerFormKindMatchesTabKind(t *testing.T) {
	m := newTestModel()
	expected := map[TabKind]FormKind{
		tabProjects:    formProject,
		tabQuotes:      formQuote,
		tabMaintenance: formMaintenance,
		tabAppliances:  formAppliance,
	}
	for _, tab := range m.tabs {
		want, ok := expected[tab.Kind]
		if !ok {
			continue
		}
		if tab.Handler.FormKind() != want {
			t.Fatalf(
				"tab %s handler FormKind() = %d, want %d",
				tab.Name,
				tab.Handler.FormKind(),
				want,
			)
		}
	}
}
