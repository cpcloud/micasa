// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"path/filepath"
	"testing"

	"github.com/cpcloud/micasa/internal/data"
	"github.com/cpcloud/micasa/internal/fake"
)

// newTestModelWithDemoData creates a Model backed by a real SQLite store,
// seeded with randomized demo data from the given HomeFaker. This provides
// richer test scenarios than newTestModelWithStore (which has only defaults).
func newTestModelWithDemoData(t *testing.T, seed uint64) *Model {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := data.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	if err := store.SeedDefaults(); err != nil {
		t.Fatalf("SeedDefaults: %v", err)
	}
	h := fake.New(seed)
	if err := store.SeedDemoDataFrom(h); err != nil {
		t.Fatalf("SeedDemoData: %v", err)
	}
	m, err := NewModel(store, Options{DBPath: path})
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	m.width = 120
	m.height = 40
	if m.mode == modeForm {
		m.exitForm()
	}
	m.showDashboard = false
	return m
}
