// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMagValueMoney(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"thousands", "$5,234.23", "$\U0001F821 3"},
		{"hundreds", "$500.00", "$\U0001F821 2"},
		{"millions", "$1,000,000.00", "$\U0001F821 6"},
		{"tens", "$42.00", "$\U0001F821 1"},
		{"single digit", "$7.50", "$\U0001F821 0"},
		{"sub-dollar", "$0.50", "$\U0001F821 -1"},
		{"zero", "$0.00", "$\U0001F821 0"},
		{"negative", "-$5.00", "-$\U0001F821 0"},
		{"negative large", "-$12,345.00", "-$\U0001F821 4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cell{Value: tt.value, Kind: cellMoney}
			assert.Equal(t, tt.want, magValue(c))
		})
	}
}

func TestMagValuePlainNumbers(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"integer", "42", "\U0001F821 1"},
		{"zero", "0", "\U0001F821 0"},
		{"large", "1000000", "\U0001F821 6"},
		{"decimal", "3.14", "\U0001F821 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cell{Value: tt.value, Kind: cellReadonly}
			assert.Equal(t, tt.want, magValue(c))
		})
	}
}

func TestMagValueSkipsNonNumeric(t *testing.T) {
	tests := []struct {
		name  string
		value string
		kind  cellKind
	}{
		{"text", "Kitchen Remodel", cellText},
		{"status", "underway", cellStatus},
		{"date", "2026-02-12", cellDate},
		{"warranty date", "2027-06-15", cellWarranty},
		{"urgency date", "2026-03-01", cellUrgency},
		{"notes", "Some long note", cellNotes},
		{"empty", "", cellText},
		{"dash", "\u2014", cellMoney},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cell{Value: tt.value, Kind: tt.kind}
			assert.Equal(t, tt.value, magValue(c), "non-numeric value should be unchanged")
		})
	}
}

func TestMagTransformCells(t *testing.T) {
	rows := [][]cell{
		{
			{Value: "Kitchen Remodel", Kind: cellText},
			{Value: "$5,234.23", Kind: cellMoney},
			{Value: "3", Kind: cellDrilldown},
		},
		{
			{Value: "Deck", Kind: cellText},
			{Value: "$100.00", Kind: cellMoney},
			{Value: "0", Kind: cellDrilldown},
		},
	}
	out := magTransformCells(rows)

	// Text cells unchanged.
	assert.Equal(t, "Kitchen Remodel", out[0][0].Value)
	assert.Equal(t, "Deck", out[1][0].Value)

	// Money cells transformed.
	assert.Equal(t, "$\U0001F821 3", out[0][1].Value)
	assert.Equal(t, "$\U0001F821 2", out[1][1].Value)

	// Drilldown counts transformed.
	assert.Equal(t, "\U0001F821 0", out[0][2].Value)
	assert.Equal(t, "\U0001F821 0", out[1][2].Value)

	// Original rows are not modified.
	assert.Equal(t, "$5,234.23", rows[0][1].Value)
}

func TestMagModeToggle(t *testing.T) {
	m := newTestModel()
	assert.False(t, m.magMode)
	sendKey(m, "m")
	assert.True(t, m.magMode)
	sendKey(m, "m")
	assert.False(t, m.magMode)
}

func TestMagModeWorksInEditMode(t *testing.T) {
	m := newTestModel()
	m.enterEditMode()
	assert.False(t, m.magMode)
	sendKey(m, "m")
	assert.True(t, m.magMode)
}

func TestMagCents(t *testing.T) {
	assert.Equal(t, "$\U0001F821 3", magCents(523423))
	assert.Equal(t, "$\U0001F821 2", magCents(50000))
	assert.Equal(t, "$\U0001F821 0", magCents(100))
}

func TestMagOptionalCentsNil(t *testing.T) {
	assert.Equal(t, "", magOptionalCents(nil))
}

func TestMagOptionalCentsPresent(t *testing.T) {
	cents := int64(100000)
	assert.Equal(t, "$\U0001F821 3", magOptionalCents(&cents))
}
