package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/micasa/micasa/internal/app"
	"github.com/micasa/micasa/internal/data"
)

func main() {
	dbOverride, showHelp, err := parseArgs(os.Args[1:])
	if err != nil {
		fail("parse args", err)
	}
	if showHelp {
		printHelp()
		return
	}
	dbPath, err := resolveDBPath(dbOverride)
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
	model, err := app.NewModel(store, app.Options{})
	if err != nil {
		fail("initialize app", err)
	}
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fail("run app", err)
	}
}

func parseArgs(args []string) (string, bool, error) {
	var dbPath string
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			return "", true, nil
		default:
			if strings.HasPrefix(arg, "-") {
				return "", false, fmt.Errorf("unknown flag: %s", arg)
			}
			if dbPath != "" {
				return "", false, fmt.Errorf("too many arguments")
			}
			dbPath = arg
		}
	}
	return dbPath, false, nil
}

func resolveDBPath(override string) (string, error) {
	if override != "" {
		return override, nil
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
