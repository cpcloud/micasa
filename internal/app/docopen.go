// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

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
func openFileCmd(path string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", path)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", "", path)
		default:
			cmd = exec.Command("xdg-open", path)
		}
		_ = cmd.Start()
		return nil
	}
}
