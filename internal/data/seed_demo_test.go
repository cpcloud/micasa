// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"testing"
)

func TestSeedDemoDataPopulatesAllEntities(t *testing.T) {
	store := newTestStoreWithDemoData(t, 99)

	house, err := store.HouseProfile()
	if err != nil {
		t.Fatalf("HouseProfile: %v", err)
	}
	if house.Nickname == "" {
		t.Error("house nickname empty")
	}
	if house.YearBuilt == 0 {
		t.Error("house year built not set")
	}

	vendors, err := store.ListVendors()
	if err != nil {
		t.Fatalf("ListVendors: %v", err)
	}
	if len(vendors) == 0 {
		t.Error("no vendors seeded")
	}

	projects, err := store.ListProjects(false)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) == 0 {
		t.Error("no projects seeded")
	}

	appliances, err := store.ListAppliances(false)
	if err != nil {
		t.Fatalf("ListAppliances: %v", err)
	}
	if len(appliances) == 0 {
		t.Error("no appliances seeded")
	}

	maint, err := store.ListMaintenance(false)
	if err != nil {
		t.Fatalf("ListMaintenance: %v", err)
	}
	if len(maint) == 0 {
		t.Error("no maintenance items seeded")
	}
}

func TestSeedDemoDataDeterministic(t *testing.T) {
	store1 := newTestStoreWithDemoData(t, 42)
	store2 := newTestStoreWithDemoData(t, 42)

	h1, _ := store1.HouseProfile()
	h2, _ := store2.HouseProfile()

	if h1.Nickname != h2.Nickname {
		t.Errorf("same seed produced different house names: %q vs %q", h1.Nickname, h2.Nickname)
	}
}

func TestSeedDemoDataVariety(t *testing.T) {
	names := make(map[string]bool)
	for seed := uint64(0); seed < 5; seed++ {
		store := newTestStoreWithDemoData(t, seed)
		h, _ := store.HouseProfile()
		names[h.Nickname] = true
	}
	if len(names) < 3 {
		t.Errorf("expected variety across seeds, got only %d unique house names", len(names))
	}
}

func TestSeedDemoDataSkipsIfDataExists(t *testing.T) {
	store := newTestStoreWithDemoData(t, 1)

	vendors1, _ := store.ListVendors()
	count1 := len(vendors1)

	// Call again -- should be a no-op.
	if err := store.SeedDemoData(); err != nil {
		t.Fatalf("second SeedDemoData: %v", err)
	}

	vendors2, _ := store.ListVendors()
	if len(vendors2) != count1 {
		t.Errorf("vendor count changed: %d -> %d", count1, len(vendors2))
	}
}
