// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
	"time"
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

func TestDashJumpClearsStaleFlag(t *testing.T) {
	m := newTestModelWithDemoData(t, 42)
	m.width = 120
	m.height = 40

	// Open the dashboard and load data so we have nav entries.
	m.showDashboard = true
	if err := m.loadDashboardAt(time.Now()); err != nil {
		t.Fatal(err)
	}
	if m.dashNavCount() == 0 {
		t.Skip("no dashboard nav entries in demo data")
	}

	// Mark all tabs stale.
	for i := range m.tabs {
		m.tabs[i].Stale = true
	}

	// Jump to the first dashboard entry.
	m.dashCursor = 0
	targetTab := m.dashNav[0].Tab
	m.dashJump()

	// The target tab should be fresh after the jump.
	idx := tabIndex(targetTab)
	if m.tabs[idx].Stale {
		t.Errorf("tab %d (%s) should not be stale after dashJump", idx, m.tabs[idx].Name)
	}

	// A different tab should still be stale.
	otherIdx := (idx + 1) % len(m.tabs)
	if !m.tabs[otherIdx].Stale {
		t.Errorf("tab %d (%s) should still be stale after jumping to tab %d",
			otherIdx, m.tabs[otherIdx].Name, idx)
	}
}

func TestNavigateToLinkClearsStaleFlag(t *testing.T) {
	m := newTestModelWithDemoData(t, 42)
	m.width = 120
	m.height = 40

	// Mark all tabs stale.
	for i := range m.tabs {
		m.tabs[i].Stale = true
	}

	// Navigate to the Vendors tab via a link. The target ID doesn't need
	// to match an actual row — we just verify the tab reload happens.
	link := &columnLink{TargetTab: tabVendors}
	_ = m.navigateToLink(link, 1)

	vendorIdx := tabIndex(tabVendors)
	if m.tabs[vendorIdx].Stale {
		t.Errorf("vendors tab should not be stale after navigateToLink")
	}

	// A different tab should still be stale.
	projIdx := tabIndex(tabProjects)
	if !m.tabs[projIdx].Stale {
		t.Errorf("projects tab should still be stale after navigating to vendors")
	}
}

func TestCloseDetailClearsStaleParentTab(t *testing.T) {
	m := newTestModelWithDemoData(t, 42)
	m.width = 120
	m.height = 40

	// Switch to Maintenance tab and open a service log detail view.
	m.active = tabIndex(tabMaintenance)
	_ = m.reloadActiveTab()

	// We need a maintenance item ID to open the detail.
	tab := m.activeTab()
	if tab == nil || len(tab.Rows) == 0 {
		t.Skip("no maintenance rows in demo data")
	}
	itemID := tab.Rows[0].ID

	if err := m.openServiceLogDetail(itemID, "Test Item"); err != nil {
		t.Fatal(err)
	}
	if m.detail == nil {
		t.Fatal("expected detail view to be open")
	}

	// Mark the parent (Maintenance) tab stale while in the detail view.
	maintIdx := tabIndex(tabMaintenance)
	m.tabs[maintIdx].Stale = true

	// Close the detail — should reload the stale parent tab.
	m.closeDetail()

	if m.tabs[maintIdx].Stale {
		t.Error("maintenance tab should not be stale after closeDetail")
	}
}
