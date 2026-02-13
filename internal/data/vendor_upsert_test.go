// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.AutoMigrate())
	return store.db
}

func TestFindOrCreateVendorNewVendor(t *testing.T) {
	db := openTestDB(t)
	v, err := findOrCreateVendor(db, Vendor{Name: "New Plumber"})
	require.NoError(t, err)
	assert.NotZero(t, v.ID)
	assert.Equal(t, "New Plumber", v.Name)
}

func TestFindOrCreateVendorExistingClearsFields(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, db.Create(&Vendor{Name: "Existing Co", Phone: "555-0000"}).Error)

	// Passing empty contact fields clears them on the existing vendor.
	v, err := findOrCreateVendor(db, Vendor{Name: "Existing Co"})
	require.NoError(t, err)

	var reloaded Vendor
	require.NoError(t, db.First(&reloaded, v.ID).Error)
	assert.Empty(t, reloaded.Phone, "empty phone should clear existing value")
}

func TestFindOrCreateVendorExistingPreservesWhenPassedThrough(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, db.Create(&Vendor{
		Name: "Preserve Co", Phone: "555-0000", Notes: "keep me",
	}).Error)

	// Passing the existing values back preserves them.
	v, err := findOrCreateVendor(db, Vendor{
		Name:  "Preserve Co",
		Phone: "555-0000",
		Notes: "keep me",
	})
	require.NoError(t, err)

	var reloaded Vendor
	require.NoError(t, db.First(&reloaded, v.ID).Error)
	assert.Equal(t, "555-0000", reloaded.Phone)
	assert.Equal(t, "keep me", reloaded.Notes)
}

func TestFindOrCreateVendorExistingWithUpdates(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, db.Create(&Vendor{Name: "Update Co"}).Error)

	v, err := findOrCreateVendor(db, Vendor{
		Name:        "Update Co",
		ContactName: "Alice",
		Email:       "alice@update.co",
		Phone:       "555-1111",
		Website:     "https://update.co",
		Notes:       "great vendor",
	})
	require.NoError(t, err)

	var reloaded Vendor
	require.NoError(t, db.First(&reloaded, v.ID).Error)
	assert.Equal(t, "Alice", reloaded.ContactName)
	assert.Equal(t, "alice@update.co", reloaded.Email)
	assert.Equal(t, "555-1111", reloaded.Phone)
	assert.Equal(t, "https://update.co", reloaded.Website)
	assert.Equal(t, "great vendor", reloaded.Notes)
}

func TestFindOrCreateVendorEmptyNameReturnsError(t *testing.T) {
	db := openTestDB(t)
	_, err := findOrCreateVendor(db, Vendor{Name: ""})
	assert.Error(t, err)
}

func TestFindOrCreateVendorWhitespaceNameReturnsError(t *testing.T) {
	db := openTestDB(t)
	_, err := findOrCreateVendor(db, Vendor{Name: "   "})
	assert.Error(t, err)
}
