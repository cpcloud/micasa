// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
)

func TestReloadAfterMutationMarksOtherTabsStale(t *testing.T) {
	m := newTestModelWithDemoData(t, 42)
	m.width = 120
	m.height = 40

	// Start on the Projects tab (index 0).
	m.active = 0
	m.reloadAfterMutation()

	// Active tab (0) should NOT be stale.
	if m.tabs[0].Stale {
		t.Error("active tab should not be stale after reloadAfterMutation")
	}

	// All other tabs should be stale.
	for i := 1; i < len(m.tabs); i++ {
		if !m.tabs[i].Stale {
			t.Errorf("tab %d (%s) should be stale after mutation on tab 0", i, m.tabs[i].Name)
		}
	}
}

func TestNavigatingToStaleTabClearsStaleFlag(t *testing.T) {
	m := newTestModelWithDemoData(t, 42)
	m.width = 120
	m.height = 40

	// Simulate a mutation on tab 0 to mark others stale.
	m.active = 0
	m.reloadAfterMutation()

	// Navigate to the next tab.
	m.nextTab()
	if m.active != 1 {
		t.Fatalf("expected active=1, got %d", m.active)
	}

	// After navigation, the new active tab should not be stale.
	if m.tabs[1].Stale {
		t.Error("tab 1 should not be stale after navigating to it")
	}

	// But tab 2 should still be stale (we haven't visited it).
	if !m.tabs[2].Stale {
		t.Error("tab 2 should still be stale")
	}
}

func TestPrevTabClearsStaleFlag(t *testing.T) {
	m := newTestModelWithDemoData(t, 42)
	m.width = 120
	m.height = 40

	// Start on tab 2, mutate to mark others stale.
	m.active = 2
	m.reloadAfterMutation()

	// Navigate backward.
	m.prevTab()
	if m.active != 1 {
		t.Fatalf("expected active=1, got %d", m.active)
	}
	if m.tabs[1].Stale {
		t.Error("tab 1 should not be stale after navigating to it via prevTab")
	}
}

func TestReloadAllClearsAllStaleFlags(t *testing.T) {
	m := newTestModelWithDemoData(t, 42)
	m.width = 120
	m.height = 40

	// Mark tabs stale.
	for i := range m.tabs {
		m.tabs[i].Stale = true
	}

	// reloadAllTabs resets all data, and reloadIfStale clears per-tab.
	m.reloadAll()

	// After reloadAll, no tabs should be stale (they were all freshly loaded).
	for i := range m.tabs {
		if m.tabs[i].Stale {
			t.Errorf("tab %d (%s) should not be stale after reloadAll", i, m.tabs[i].Name)
		}
	}
}
