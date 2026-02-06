// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/micasa/micasa/internal/app"
	"github.com/micasa/micasa/internal/data"
)

func main() {
	dbOverride, demo, showHelp, err := parseArgs(os.Args[1:])
	if err != nil {
		fail("parse args", err)
	}
	if showHelp {
		printHelp()
		return
	}
	dbPath, err := resolveDBPath(dbOverride, demo)
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
	if demo {
		if err := store.SeedDemoData(); err != nil {
			fail("seed demo data", err)
		}
	}
	model, err := app.NewModel(store, app.Options{DBPath: dbPath})
	if err != nil {
		fail("initialize app", err)
	}
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fail("run app", err)
	}
}

func parseArgs(args []string) (string, bool, bool, error) {
	var dbPath string
	var demo bool
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			return "", false, true, nil
		case "--demo":
			demo = true
		default:
			if strings.HasPrefix(arg, "-") {
				return "", false, false, fmt.Errorf("unknown flag: %s", arg)
			}
			if dbPath != "" {
				return "", false, false, fmt.Errorf("too many arguments")
			}
			dbPath = arg
		}
	}
	return dbPath, demo, false, nil
}

func resolveDBPath(override string, demo bool) (string, error) {
	if override != "" {
		return override, nil
	}
	if demo {
		return filepath.Join(os.TempDir(), "micasa-demo.db"), nil
	}
	return data.DefaultDBPath()
}

func printHelp() {
	lines := []string{
		"micasa - home improvement tracker",
		"",
		"Usage:",
		"  micasa [db-path] [--help]",
		"",
		"Options:",
		"  -h, --help  Show help and exit.",
		"  --demo      Launch with sample data in a temporary database.",
		"",
		"Args:",
		"  db-path     Override default sqlite path.",
		"",
		"Environment:",
		"  MICASA_DB_PATH  Override default sqlite path.",
	}
	_, _ = fmt.Fprintln(os.Stdout, strings.Join(lines, "\n"))
}

func fail(context string, err error) {
	fmt.Fprintf(os.Stderr, "micasa: %s: %v\n", context, err)
	os.Exit(1)
}
