// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/micasa-dev/micasa/internal/data"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func maintenanceEntityDef() entityDef[data.MaintenanceItem] {
	return entityDef[data.MaintenanceItem]{
		name:        "maintenance",
		singular:    "maintenance item",
		tableHeader: "MAINTENANCE",
		cols:        maintenanceCols,
		toMap:       maintenanceToMap,
		list: func(s *data.Store, deleted bool) ([]data.MaintenanceItem, error) {
			return s.ListMaintenance(deleted)
		},
		get: func(s *data.Store, id string) (data.MaintenanceItem, error) {
			return s.GetMaintenance(id)
		},
		decodeAndCreate: maintenanceCreate,
		decodeAndUpdate: maintenanceUpdate,
		del: func(s *data.Store, id string) error {
			return s.DeleteMaintenance(id)
		},
		restore: func(s *data.Store, id string) error {
			return s.RestoreMaintenance(id)
		},
		deletedAt: func(m data.MaintenanceItem) gorm.DeletedAt {
			return m.DeletedAt
		},
	}
}

func newMaintenanceCmd() *cobra.Command {
	return buildEntityCmd(maintenanceEntityDef())
}

func maintenanceCreate(store *data.Store, raw json.RawMessage) (data.MaintenanceItem, error) {
	var m data.MaintenanceItem
	if err := json.Unmarshal(raw, &m); err != nil {
		return data.MaintenanceItem{}, fmt.Errorf("invalid JSON: %w", err)
	}
	if m.Name == "" {
		return data.MaintenanceItem{}, errors.New("name is required")
	}
	if m.CategoryID == "" {
		return data.MaintenanceItem{}, errors.New("category_id is required")
	}
	if err := store.CreateMaintenance(&m); err != nil {
		return data.MaintenanceItem{}, err
	}
	return store.GetMaintenance(m.ID)
}

func maintenanceUpdate(
	store *data.Store,
	id string,
	raw json.RawMessage,
) (data.MaintenanceItem, error) {
	existing, err := store.GetMaintenance(id)
	if err != nil {
		return data.MaintenanceItem{}, fmt.Errorf("get maintenance: %w", err)
	}

	fields, err := parseFields(raw)
	if err != nil {
		return data.MaintenanceItem{}, err
	}

	for _, pair := range []struct {
		key string
		dst any
	}{
		{"name", &existing.Name},
		{"category_id", &existing.CategoryID},
		{"appliance_id", &existing.ApplianceID},
		{"season", &existing.Season},
		{"interval_months", &existing.IntervalMonths},
		{"notes", &existing.Notes},
		{"cost_cents", &existing.CostCents},
	} {
		if err := mergeField(fields, pair.key, pair.dst); err != nil {
			return data.MaintenanceItem{}, err
		}
	}

	for _, datePair := range []struct {
		key string
		dst **time.Time
	}{
		{"last_serviced_at", &existing.LastServicedAt},
		{"due_date", &existing.DueDate},
	} {
		if dateStr, ok := stringField(fields, datePair.key); ok {
			parsed, dateErr := data.ParseOptionalDate(dateStr)
			if dateErr != nil {
				return data.MaintenanceItem{}, fmt.Errorf("%s: %w", datePair.key, dateErr)
			}
			*datePair.dst = parsed
		} else if _, present := fields[datePair.key]; present {
			*datePair.dst = nil
		}
	}

	if err := store.UpdateMaintenance(existing); err != nil {
		return data.MaintenanceItem{}, err
	}
	return store.GetMaintenance(id)
}
