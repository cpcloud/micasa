// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

// openFileResultMsg carries the outcome of an OS-viewer launch back to the
// Bubble Tea event loop so the status bar can surface errors.
type openFileResultMsg struct{ Err error }

// openSelectedDocument extracts the selected document to the cache and
// launches the OS-appropriate viewer. Only operates on the Documents tab;
// returns nil (no-op) on other tabs.
func (m *Model) openSelectedDocument() tea.Cmd {
	tab := m.effectiveTab()
	if tab == nil || tab.Kind != tabDocuments {
		return nil
	}

	meta, ok := m.selectedRowMeta()
	if !ok || meta.Deleted {
		return nil
	}

	cachePath, err := m.store.ExtractDocument(meta.ID)
	if err != nil {
		m.setStatusError(fmt.Sprintf("extract: %s", err))
		return nil
	}

	return openFileCmd(cachePath)
}

// openFileCmd returns a tea.Cmd that opens the given path with the OS viewer.
// The command runs to completion so exit-status errors (e.g. no handler for
// the MIME type) are captured and returned as an openFileResultMsg.
//
// Only called from openSelectedDocument with a path returned by
// Store.ExtractDocument (always under the XDG cache directory).
func openFileCmd(path string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		var openerName string
		switch runtime.GOOS {
		case "darwin":
			openerName = "open"
			cmd = exec.Command("open", path) //nolint:gosec // path from trusted cache directory
		case "windows":
			openerName = "cmd"
			cmd = exec.Command(
				"cmd",
				"/c",
				"start",
				"",
				path,
			) //nolint:gosec // path from trusted cache directory
		default:
			openerName = "xdg-open"
			cmd = exec.Command("xdg-open", path) //nolint:gosec // path from trusted cache directory
		}
		err := cmd.Run()
		if err != nil {
			err = wrapOpenerError(err, openerName)
		}
		return openFileResultMsg{Err: err}
	}
}

// wrapOpenerError adds an actionable hint when the OS file-opener command
// is missing or cannot be found.
func wrapOpenerError(err error, openerName string) error {
	if !errors.Is(err, exec.ErrNotFound) {
		return err
	}
	switch openerName {
	case "xdg-open":
		return fmt.Errorf(
			"%s not found -- install xdg-utils (e.g. apt install xdg-utils)",
			openerName,
		)
	case "open":
		return fmt.Errorf(
			"%s not found -- expected on macOS; is this a headless environment?",
			openerName,
		)
	default:
		return fmt.Errorf("%s not found -- no file opener available", openerName)
	}
}
