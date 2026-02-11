// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/cpcloud/micasa/internal/fake"
)

func benchStore(b *testing.B, seed uint64) *Store {
	b.Helper()
	path := filepath.Join(b.TempDir(), "bench.db")
	store, err := Open(path)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = store.Close() })
	if err := store.AutoMigrate(); err != nil {
		b.Fatal(err)
	}
	if err := store.SeedDefaults(); err != nil {
		b.Fatal(err)
	}
	if err := store.SeedDemoDataFrom(fake.New(seed)); err != nil {
		b.Fatal(err)
	}
	return store
}

func BenchmarkListMaintenanceWithSchedule(b *testing.B) {
	store := benchStore(b, 42)
	b.ResetTimer()
	for b.Loop() {
		_, _ = store.ListMaintenanceWithSchedule()
	}
}

func BenchmarkListActiveProjects(b *testing.B) {
	store := benchStore(b, 42)
	b.ResetTimer()
	for b.Loop() {
		_, _ = store.ListActiveProjects()
	}
}

func BenchmarkListProjects(b *testing.B) {
	store := benchStore(b, 42)
	b.ResetTimer()
	for b.Loop() {
		_, _ = store.ListProjects(false)
	}
}

func BenchmarkListMaintenance(b *testing.B) {
	store := benchStore(b, 42)
	b.ResetTimer()
	for b.Loop() {
		_, _ = store.ListMaintenance(false)
	}
}

func BenchmarkListVendors(b *testing.B) {
	store := benchStore(b, 42)
	b.ResetTimer()
	for b.Loop() {
		_, _ = store.ListVendors(false)
	}
}

func BenchmarkListExpiringWarranties(b *testing.B) {
	store := benchStore(b, 42)
	now := time.Now()
	b.ResetTimer()
	for b.Loop() {
		_, _ = store.ListExpiringWarranties(now, 30*24*time.Hour, 90*24*time.Hour)
	}
}

func BenchmarkYTDServiceSpendCents(b *testing.B) {
	store := benchStore(b, 42)
	yearStart := time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	b.ResetTimer()
	for b.Loop() {
		_, _ = store.YTDServiceSpendCents(yearStart)
	}
}
