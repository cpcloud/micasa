// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package main

import (
	"github.com/micasa-dev/micasa/internal/data"
	"github.com/spf13/cobra"
)

func projectTypeEntityDef() entityDef[data.ProjectType] {
	return entityDef[data.ProjectType]{
		name:        "project-type",
		singular:    "project type",
		tableHeader: "PROJECT TYPES",
		cols:        projectTypeCols,
		toMap:       projectTypeToMap,
		list: func(s *data.Store, _ bool) ([]data.ProjectType, error) {
			return s.ProjectTypes()
		},
	}
}

func newProjectTypeCmd() *cobra.Command {
	return buildEntityCmd(projectTypeEntityDef())
}

func maintenanceCategoryEntityDef() entityDef[data.MaintenanceCategory] {
	return entityDef[data.MaintenanceCategory]{
		name:        "maintenance-category",
		singular:    "maintenance category",
		tableHeader: "MAINTENANCE CATEGORIES",
		cols:        maintenanceCategoryCols,
		toMap:       maintenanceCategoryToMap,
		list: func(s *data.Store, _ bool) ([]data.MaintenanceCategory, error) {
			return s.MaintenanceCategories()
		},
	}
}

func newMaintenanceCategoryCmd() *cobra.Command {
	return buildEntityCmd(maintenanceCategoryEntityDef())
}
