// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"path/filepath"
	"testing"

	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	return store.db
}

func TestFindOrCreateVendorNewVendor(t *testing.T) {
	db := openTestDB(t)
	v, err := findOrCreateVendor(db, Vendor{Name: "New Plumber"})
	if err != nil {
		t.Fatalf("findOrCreateVendor: %v", err)
	}
	if v.ID == 0 {
		t.Error("expected non-zero ID for new vendor")
	}
	if v.Name != "New Plumber" {
		t.Errorf("name = %q", v.Name)
	}
}

func TestFindOrCreateVendorExistingNoUpdates(t *testing.T) {
	db := openTestDB(t)

	// Create first.
	if err := db.Create(&Vendor{Name: "Existing Co", Phone: "555-0000"}).Error; err != nil {
		t.Fatal(err)
	}

	// Find with same name, no new fields.
	v, err := findOrCreateVendor(db, Vendor{Name: "Existing Co"})
	if err != nil {
		t.Fatal(err)
	}
	if v.Phone != "555-0000" {
		t.Errorf("phone = %q, want 555-0000 (should keep original)", v.Phone)
	}
}

func TestFindOrCreateVendorExistingWithUpdates(t *testing.T) {
	db := openTestDB(t)

	if err := db.Create(&Vendor{Name: "Update Co"}).Error; err != nil {
		t.Fatal(err)
	}

	// Pass new contact info alongside the name match.
	v, err := findOrCreateVendor(db, Vendor{
		Name:        "Update Co",
		ContactName: "Alice",
		Email:       "alice@update.co",
		Phone:       "555-1111",
		Website:     "https://update.co",
		Notes:       "great vendor",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Reload to verify persistence.
	var reloaded Vendor
	if err := db.First(&reloaded, v.ID).Error; err != nil {
		t.Fatal(err)
	}
	if reloaded.ContactName != "Alice" {
		t.Errorf("contact = %q", reloaded.ContactName)
	}
	if reloaded.Email != "alice@update.co" {
		t.Errorf("email = %q", reloaded.Email)
	}
	if reloaded.Phone != "555-1111" {
		t.Errorf("phone = %q", reloaded.Phone)
	}
	if reloaded.Website != "https://update.co" {
		t.Errorf("website = %q", reloaded.Website)
	}
	if reloaded.Notes != "great vendor" {
		t.Errorf("notes = %q", reloaded.Notes)
	}
}

func TestFindOrCreateVendorEmptyNameReturnsError(t *testing.T) {
	db := openTestDB(t)
	_, err := findOrCreateVendor(db, Vendor{Name: ""})
	if err == nil {
		t.Error("expected error for empty vendor name")
	}
}

func TestFindOrCreateVendorWhitespaceNameReturnsError(t *testing.T) {
	db := openTestDB(t)
	_, err := findOrCreateVendor(db, Vendor{Name: "   "})
	if err == nil {
		t.Error("expected error for whitespace-only vendor name")
	}
}
