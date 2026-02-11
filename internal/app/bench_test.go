// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/cpcloud/micasa/internal/data"
	"github.com/cpcloud/micasa/internal/fake"
)

// benchModel returns a Model populated with demo data, sized for a
// realistic terminal. Reusable across benchmarks.
func benchModel(b *testing.B) *Model {
	b.Helper()
	path := filepath.Join(b.TempDir(), "bench.db")
	store, err := data.Open(path)
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
	if err := store.SeedDemoDataFrom(fake.New(42)); err != nil {
		b.Fatal(err)
	}
	m, err := NewModel(store, Options{DBPath: path})
	if err != nil {
		b.Fatal(err)
	}
	m.width = 120
	m.height = 40
	m.showDashboard = false
	if err := m.reloadAllTabs(); err != nil {
		b.Fatal(err)
	}
	return m
}

func BenchmarkView(b *testing.B) {
	m := benchModel(b)
	b.ResetTimer()
	for b.Loop() {
		_ = m.View()
	}
}

func BenchmarkViewDashboard(b *testing.B) {
	m := benchModel(b)
	m.showDashboard = true
	if err := m.loadDashboardAt(time.Now()); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for b.Loop() {
		_ = m.View()
	}
}

func BenchmarkReloadAll(b *testing.B) {
	m := benchModel(b)
	b.ResetTimer()
	for b.Loop() {
		m.reloadAll()
	}
}

func BenchmarkReloadActiveTab(b *testing.B) {
	m := benchModel(b)
	b.ResetTimer()
	for b.Loop() {
		_ = m.reloadActiveTab()
	}
}

func BenchmarkReloadAfterMutation(b *testing.B) {
	m := benchModel(b)
	b.ResetTimer()
	for b.Loop() {
		m.reloadAfterMutation()
	}
}

func BenchmarkReloadAfterMutationWithDashboard(b *testing.B) {
	m := benchModel(b)
	m.showDashboard = true
	if err := m.loadDashboardAt(time.Now()); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for b.Loop() {
		m.reloadAfterMutation()
	}
}

func BenchmarkLoadDashboard(b *testing.B) {
	m := benchModel(b)
	now := time.Now()
	b.ResetTimer()
	for b.Loop() {
		_ = m.loadDashboardAt(now)
	}
}

func BenchmarkColumnWidths(b *testing.B) {
	m := benchModel(b)
	tab := m.activeTab()
	visSpecs, visCells, _, _, _ := visibleProjection(tab)
	sepW := 3
	b.ResetTimer()
	for b.Loop() {
		_ = columnWidths(visSpecs, visCells, 120, sepW)
	}
}

func BenchmarkNaturalWidths(b *testing.B) {
	m := benchModel(b)
	tab := m.activeTab()
	visSpecs, visCells, _, _, _ := visibleProjection(tab)
	b.ResetTimer()
	for b.Loop() {
		_ = naturalWidths(visSpecs, visCells)
	}
}

func BenchmarkVisibleProjection(b *testing.B) {
	m := benchModel(b)
	tab := m.activeTab()
	b.ResetTimer()
	for b.Loop() {
		_, _, _, _, _ = visibleProjection(tab)
	}
}

func BenchmarkComputeTableViewport(b *testing.B) {
	m := benchModel(b)
	tab := m.activeTab()
	sep := m.styles.TableSeparator.Render(" │ ")
	b.ResetTimer()
	for b.Loop() {
		_ = computeTableViewport(tab, 120, sep, m.styles)
	}
}

func BenchmarkTableView(b *testing.B) {
	m := benchModel(b)
	tab := m.activeTab()
	b.ResetTimer()
	for b.Loop() {
		_ = m.tableView(tab)
	}
}

func BenchmarkDashboardView(b *testing.B) {
	m := benchModel(b)
	m.showDashboard = true
	if err := m.loadDashboardAt(time.Now()); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for b.Loop() {
		_ = m.dashboardView(30)
	}
}

func BenchmarkBuildBaseView(b *testing.B) {
	m := benchModel(b)
	b.ResetTimer()
	for b.Loop() {
		_ = m.buildBaseView()
	}
}

func BenchmarkDimBackground(b *testing.B) {
	m := benchModel(b)
	base := m.buildBaseView()
	b.ResetTimer()
	for b.Loop() {
		_ = dimBackground(base)
	}
}
