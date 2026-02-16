// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cpcloud/micasa/internal/app"
	"github.com/cpcloud/micasa/internal/config"
	"github.com/cpcloud/micasa/internal/data"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

type cli struct {
	DBPath    string           `arg:"" optional:"" help:"SQLite database path. Pass with --demo to persist demo data." env:"MICASA_DB_PATH"`
	Demo      bool             `                   help:"Launch with sample data in an in-memory database."`
	PrintPath bool             `                   help:"Print the resolved database path and exit."`
	Version   kong.VersionFlag `                   help:"Show version and exit."`
}

func main() {
	var c cli
	kong.Parse(&c,
		kong.Name(data.AppName),
		kong.Description("A terminal UI for tracking everything about your home."),
		kong.UsageOnError(),
		kong.Vars{"version": versionString()},
	)

	dbPath, err := resolveDBPath(c)
	if err != nil {
		fail("resolve db path", err)
	}
	if c.PrintPath {
		fmt.Println(dbPath)
		return
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

	cfg, err := config.Load()
	if err != nil {
		fail("load config", err)
	}
	if err := store.SetMaxDocumentSize(cfg.Documents.MaxFileSize); err != nil {
		fail("configure document size limit", err)
	}
	cacheDir, err := data.DocumentCacheDir()
	if err != nil {
		fail("resolve document cache directory", err)
	}
	if _, err := data.EvictStaleCache(cacheDir, cfg.Documents.CacheTTLDays); err != nil {
		fail("evict stale cache", err)
	}

	opts := app.Options{
		DBPath:     dbPath,
		ConfigPath: config.Path(),
	}
	opts.SetLLM(cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.ExtraContext)

	model, err := app.NewModel(store, opts)
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

// versionString returns the version for display. Release builds return
// the version set via ldflags. Dev builds return the short git commit hash
// (with a -dirty suffix if the tree was modified), or "dev" as a last resort.
func versionString() string {
	if version != "dev" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}
	var revision string
	var dirty bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if revision == "" {
		return version
	}
	if dirty {
		return revision + "-dirty"
	}
	return revision
}

func fail(context string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s: %v\n", data.AppName, context, err)
	os.Exit(1)
}
