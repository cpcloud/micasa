// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/micasa/micasa/internal/app"
	"github.com/micasa/micasa/internal/data"
)

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fail("parse args", err)
	}
	if opts.showHelp {
		printHelp()
		return
	}
	dbPath, err := resolveDBPath(opts)
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
	if opts.demo {
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

type cliOpts struct {
	dbPath   string
	demo     bool
	showHelp bool
}

func parseArgs(args []string) (cliOpts, error) {
	var opts cliOpts
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			opts.showHelp = true
			return opts, nil
		case "--demo":
			opts.demo = true
		default:
			if strings.HasPrefix(arg, "-") {
				return cliOpts{}, fmt.Errorf("unknown flag: %s", arg)
			}
			if opts.dbPath != "" {
				return cliOpts{}, fmt.Errorf("too many arguments")
			}
			opts.dbPath = arg
		}
	}
	return opts, nil
}

func resolveDBPath(opts cliOpts) (string, error) {
	if opts.dbPath != "" {
		return opts.dbPath, nil
	}
	if opts.demo {
		return ":memory:", nil
	}
	return data.DefaultDBPath()
}

func printHelp() {
	lines := []string{
		"micasa - home improvement tracker",
		"",
		"Usage:",
		"  micasa [db-path] [--demo] [--help]",
		"",
		"Options:",
		"  -h, --help  Show help and exit.",
		"  --demo      Launch with sample data in an in-memory database.",
		"",
		"Args:",
		"  db-path     SQLite path. Pass with --demo to persist demo data.",
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
