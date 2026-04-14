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

func projectEntityDef() entityDef[data.Project] {
	return entityDef[data.Project]{
		name:        "project",
		singular:    "project",
		tableHeader: "PROJECTS",
		cols:        projectCols,
		toMap:       projectToMap,
		list: func(s *data.Store, deleted bool) ([]data.Project, error) {
			return s.ListProjects(deleted)
		},
		get: func(s *data.Store, id string) (data.Project, error) {
			return s.GetProject(id)
		},
		decodeAndCreate: projectCreate,
		decodeAndUpdate: projectUpdate,
		del: func(s *data.Store, id string) error {
			return s.DeleteProject(id)
		},
		restore: func(s *data.Store, id string) error {
			return s.RestoreProject(id)
		},
		deletedAt: func(p data.Project) gorm.DeletedAt {
			return p.DeletedAt
		},
	}
}

func newProjectCmd() *cobra.Command {
	return buildEntityCmd(projectEntityDef())
}

func projectCreate(store *data.Store, raw json.RawMessage) (data.Project, error) {
	fields, err := parseFields(raw)
	if err != nil {
		return data.Project{}, err
	}

	var p data.Project
	for _, pair := range []struct {
		key string
		dst any
	}{
		{"title", &p.Title},
		{"project_type_id", &p.ProjectTypeID},
		{"status", &p.Status},
		{"description", &p.Description},
		{"budget_cents", &p.BudgetCents},
		{"actual_cents", &p.ActualCents},
	} {
		if err := mergeField(fields, pair.key, pair.dst); err != nil {
			return data.Project{}, err
		}
	}

	for _, datePair := range []struct {
		key string
		dst **time.Time
	}{
		{"start_date", &p.StartDate},
		{"end_date", &p.EndDate},
	} {
		if dateStr, ok := stringField(fields, datePair.key); ok {
			parsed, dateErr := data.ParseOptionalDate(dateStr)
			if dateErr != nil {
				return data.Project{}, fmt.Errorf("%s: %w", datePair.key, dateErr)
			}
			*datePair.dst = parsed
		}
	}

	if p.Title == "" {
		return data.Project{}, errors.New("title is required")
	}
	if p.ProjectTypeID == "" {
		return data.Project{}, errors.New("project_type_id is required")
	}
	if p.Status == "" {
		p.Status = data.ProjectStatusPlanned
	}
	if err := store.CreateProject(&p); err != nil {
		return data.Project{}, err
	}
	return store.GetProject(p.ID)
}

func projectUpdate(store *data.Store, id string, raw json.RawMessage) (data.Project, error) {
	existing, err := store.GetProject(id)
	if err != nil {
		return data.Project{}, fmt.Errorf("get project: %w", err)
	}

	fields, err := parseFields(raw)
	if err != nil {
		return data.Project{}, err
	}

	for _, pair := range []struct {
		key string
		dst any
	}{
		{"title", &existing.Title},
		{"project_type_id", &existing.ProjectTypeID},
		{"status", &existing.Status},
		{"description", &existing.Description},
		{"budget_cents", &existing.BudgetCents},
		{"actual_cents", &existing.ActualCents},
	} {
		if err := mergeField(fields, pair.key, pair.dst); err != nil {
			return data.Project{}, err
		}
	}

	for _, datePair := range []struct {
		key string
		dst **time.Time
	}{
		{"start_date", &existing.StartDate},
		{"end_date", &existing.EndDate},
	} {
		if dateStr, ok := stringField(fields, datePair.key); ok {
			parsed, dateErr := data.ParseOptionalDate(dateStr)
			if dateErr != nil {
				return data.Project{}, fmt.Errorf("%s: %w", datePair.key, dateErr)
			}
			*datePair.dst = parsed
		} else if _, present := fields[datePair.key]; present {
			*datePair.dst = nil
		}
	}

	if err := store.UpdateProject(existing); err != nil {
		return data.Project{}, err
	}
	return store.GetProject(id)
}
