// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Compact intervals
// ---------------------------------------------------------------------------

func TestFormatInterval(t *testing.T) {
	tests := []struct {
		name   string
		months int
		want   string
	}{
		{"zero", 0, ""},
		{"negative", -3, ""},
		{"one month", 1, "1m"},
		{"three months", 3, "3m"},
		{"six months", 6, "6m"},
		{"eleven months", 11, "11m"},
		{"one year", 12, "1y"},
		{"two years", 24, "2y"},
		{"year and a half", 18, "1y 6m"},
		{"complex", 27, "2y 3m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatInterval(tt.months))
		})
	}
}

// ---------------------------------------------------------------------------
// Status icons
// ---------------------------------------------------------------------------

func TestStatusIconsAreDefined(t *testing.T) {
	styles := DefaultStyles()
	statuses := []string{
		"ideating", "planned", "quoted",
		"underway", "delayed", "completed", "abandoned",
	}
	for _, s := range statuses {
		icon, ok := styles.StatusIcons[s]
		assert.True(t, ok, "expected icon for status %q", s)
		assert.NotEmpty(t, icon, "icon for %q should not be empty", s)
	}
}

func TestStatusIconsAreDistinct(t *testing.T) {
	styles := DefaultStyles()
	seen := make(map[string]string) // icon -> status
	for status, icon := range styles.StatusIcons {
		if prev, ok := seen[icon]; ok {
			t.Errorf("duplicate icon %q shared by %q and %q", icon, prev, status)
		}
		seen[icon] = status
	}
}

func TestStatusIconsMatchStyleKeys(t *testing.T) {
	styles := DefaultStyles()
	// Every status that has a style should have an icon, and vice versa.
	for status := range styles.StatusStyles {
		_, ok := styles.StatusIcons[status]
		assert.True(t, ok, "StatusStyles has %q but StatusIcons does not", status)
	}
	for status := range styles.StatusIcons {
		_, ok := styles.StatusStyles[status]
		assert.True(t, ok, "StatusIcons has %q but StatusStyles does not", status)
	}
}
