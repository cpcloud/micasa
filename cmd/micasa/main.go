// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cpcloud/micasa/internal/app"
	"github.com/cpcloud/micasa/internal/data"
)

type cli struct {
	DBPath string `arg:"" optional:"" help:"SQLite database path. Pass with --demo to persist demo data." env:"MICASA_DB_PATH"`
	Demo   bool   `                   help:"Launch with sample data in an in-memory database."`
}

func main() {
	var c cli
	kong.Parse(&c,
		kong.Name(data.AppName),
		kong.Description("A terminal UI for tracking everything about your home."),
		kong.UsageOnError(),
	)

	dbPath, err := resolveDBPath(c)
	if err != nil {
		fail("resolve db path", err)
	}
	store, err := data.Open(dbPath)
	if err != nil {
		fail("open database", err)
	}
	if err := store.AutoMigrate(); err != nil {
		fail("migrate database", err)
	}
	if err := store.SeedDefaults(); err != nil {
		fail("seed defaults", err)
	}
	if c.Demo {
		if err := store.SeedDemoData(); err != nil {
			fail("seed demo data", err)
		}
	}
	model, err := app.NewModel(store, app.Options{DBPath: dbPath})
	if err != nil {
		fail("initialize app", err)
	}
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			os.Exit(130)
		}
		fail("run app", err)
	}
}

func resolveDBPath(c cli) (string, error) {
	if c.DBPath != "" {
		return c.DBPath, nil
	}
	if c.Demo {
		return ":memory:", nil
	}
	return data.DefaultDBPath()
}

func fail(context string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s: %v\n", data.AppName, context, err)
	os.Exit(1)
}
