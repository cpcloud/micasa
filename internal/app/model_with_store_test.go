// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"path/filepath"
	"testing"

	"github.com/cpcloud/micasa/internal/data"
)

// newTestModelWithStore creates a Model backed by a real in-memory SQLite
// store with seeded defaults (project types, maintenance categories). The
// model is sized to 120x40 and starts in normal mode (dashboard and house
// form dismissed).
func newTestModelWithStore(t *testing.T) *Model {
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

	// Create a house profile so NewModel doesn't open the house form.
	if err := store.CreateHouseProfile(data.HouseProfile{
		Nickname: "Test House",
	}); err != nil {
		t.Fatalf("CreateHouseProfile: %v", err)
	}

	m, err := NewModel(store, Options{DBPath: path})
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	m.width = 120
	m.height = 40
	m.showDashboard = false
	return m
}
