// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
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
func openFileCmd(path string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", path) //nolint:gosec // path from trusted cache directory
		case "windows":
			cmd = exec.Command(
				"rundll32",
				"url.dll,FileProtocolHandler",
				path,
			) //nolint:gosec // path from trusted cache directory
		default:
			cmd = exec.Command("xdg-open", path) //nolint:gosec // path from trusted cache directory
		}
		return openFileResultMsg{Err: cmd.Run()}
	}
}
