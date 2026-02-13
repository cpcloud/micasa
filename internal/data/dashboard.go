// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"time"

	"gorm.io/gorm"
)

// ListMaintenanceWithSchedule returns all non-deleted maintenance items that
// have a positive interval, preloading Category and Appliance. These are the
// items eligible for overdue/upcoming computation.
func (s *Store) ListMaintenanceWithSchedule() ([]MaintenanceItem, error) {
	var items []MaintenanceItem
	err := s.db.
		Where("interval_months > 0").
		Preload("Category").
		Preload("Appliance", func(q *gorm.DB) *gorm.DB {
			return q.Unscoped()
		}).
		Order("updated_at desc").
		Find(&items).Error
	return items, err
}

// ListActiveProjects returns non-deleted projects with status "underway" or
// "delayed", preloading ProjectType.
func (s *Store) ListActiveProjects() ([]Project, error) {
	var projects []Project
	err := s.db.
		Where("status IN ?", []string{ProjectStatusInProgress, ProjectStatusDelayed}).
		Preload("ProjectType").
		Order("updated_at desc").
		Find(&projects).Error
	return projects, err
}

// ListExpiringWarranties returns non-deleted appliances whose warranty expires
// between (now - lookBack) and (now + horizon).
func (s *Store) ListExpiringWarranties(
	now time.Time,
	lookBack, horizon time.Duration,
) ([]Appliance, error) {
	var appliances []Appliance
	from := now.Add(-lookBack)
	to := now.Add(horizon)
	err := s.db.
		Where("warranty_expiry IS NOT NULL AND warranty_expiry BETWEEN ? AND ?", from, to).
		Order("warranty_expiry asc").
		Find(&appliances).Error
	return appliances, err
}

// ListRecentServiceLogs returns the most recent service log entries across all
// maintenance items, preloading MaintenanceItem and Vendor.
func (s *Store) ListRecentServiceLogs(limit int) ([]ServiceLogEntry, error) {
	var entries []ServiceLogEntry
	err := s.db.
		Preload("MaintenanceItem").
		Preload("Vendor", func(q *gorm.DB) *gorm.DB {
			return q.Unscoped()
		}).
		Order("serviced_at desc, id desc").
		Limit(limit).
		Find(&entries).Error
	return entries, err
}

// YTDServiceSpendCents returns the total cost of service log entries with
// ServicedAt on or after the given start-of-year.
func (s *Store) YTDServiceSpendCents(yearStart time.Time) (int64, error) {
	var total *int64
	err := s.db.Model(&ServiceLogEntry{}).
		Select("COALESCE(SUM(cost_cents), 0)").
		Where("serviced_at >= ?", yearStart).
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	if total == nil {
		return 0, nil
	}
	return *total, nil
}

// YTDProjectSpendCents returns the total actual spend across all non-deleted
// projects.
func (s *Store) YTDProjectSpendCents() (int64, error) {
	var total *int64
	err := s.db.Model(&Project{}).
		Select("COALESCE(SUM(actual_cents), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	if total == nil {
		return 0, nil
	}
	return *total, nil
}
