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
	if wantsHelp(os.Args[1:]) {
		printHelp()
		return
	}
	dbPath, err := data.DefaultDBPath()
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
	model, err := app.NewModel(store)
	if err != nil {
		fail("initialize app", err)
	}
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fail("run app", err)
	}
}

func wantsHelp(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func printHelp() {
	lines := []string{
		"micasa - home improvement tracker",
		"",
		"Usage:",
		"  micasa [--help]",
		"",
		"Options:",
		"  -h, --help    Show help and exit.",
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
