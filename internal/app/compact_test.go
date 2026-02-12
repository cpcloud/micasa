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

// ---------------------------------------------------------------------------
// Compact money
// ---------------------------------------------------------------------------

func TestCompactMoneyValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"small stays full", "$500.00", "500.00"},
		{"thousands", "$5,234.23", "5.2k"},
		{"round thousands", "$45,000.00", "45k"},
		{"millions", "$1,300,000.00", "1.3M"},
		{"empty", "", ""},
		{"dash", "—", "—"},
		{"unparseable", "not money", "not money"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, compactMoneyValue(tt.input))
		})
	}
}

func TestCompactMoneyCells(t *testing.T) {
	rows := [][]cell{
		{
			{Value: "1", Kind: cellReadonly},
			{Value: "Kitchen", Kind: cellText},
			{Value: "$5,234.23", Kind: cellMoney},
			{Value: "3", Kind: cellDrilldown},
		},
		{
			{Value: "2", Kind: cellReadonly},
			{Value: "Deck", Kind: cellText},
			{Value: "$100.00", Kind: cellMoney},
			{Value: "", Kind: cellMoney},
		},
	}
	out := compactMoneyCells(rows)

	// Non-money cells unchanged.
	assert.Equal(t, "1", out[0][0].Value)
	assert.Equal(t, "Kitchen", out[0][1].Value)
	assert.Equal(t, "3", out[0][3].Value)

	// Money cells compacted ($ stripped for bare display).
	assert.Equal(t, "5.2k", out[0][2].Value)
	assert.Equal(t, "100.00", out[1][2].Value)

	// Empty money cell stays empty.
	assert.Equal(t, "", out[1][3].Value)

	// Original rows not modified.
	assert.Equal(t, "$5,234.23", rows[0][2].Value)
}
