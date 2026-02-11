// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
	"time"
)

func TestModelWithDemoDataLoadsAllTabs(t *testing.T) {
	m := newTestModelWithDemoData(t, 77)

	for i, tab := range m.tabs {
		if len(tab.Table.Rows()) == 0 {
			t.Errorf("tab %d (%s) has no rows after demo data seed", i, tab.Name)
		}
	}
}

func TestModelWithDemoDataDashboard(t *testing.T) {
	m := newTestModelWithDemoData(t, 77)
	m.showDashboard = true
	if err := m.loadDashboardAt(time.Now()); err != nil {
		t.Fatalf("loadDashboard: %v", err)
	}

	if m.dashNavCount() == 0 {
		t.Error("expected dashboard nav entries after demo data seed")
	}
}

func TestModelWithDemoDataVariedSeeds(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3, 4, 5} {
		m := newTestModelWithDemoData(t, seed)
		if m == nil {
			t.Fatalf("seed %d: nil model", seed)
		}
		totalRows := 0
		for _, tab := range m.tabs {
			totalRows += len(tab.Table.Rows())
		}
		if totalRows == 0 {
			t.Errorf("seed %d: no rows in any tab", seed)
		}
	}
}
