// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMagFormatMoneyWithUnit(t *testing.T) {
	// Used by magCents for dashboard (input still has $ from FormatCents).
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"thousands", "$5,234.23", "$ \U0001F8214"},
		{"hundreds", "$500.00", "$ \U0001F8213"},
		{"millions", "$1,000,000.00", "$ \U0001F8216"},
		{"zero", "$0.00", "$ \U0001F8210"},
		{"negative", "-$5.00", "-$ \U0001F8211"},
		{"negative large", "-$12,345.00", "-$ \U0001F8214"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cell{Value: tt.value, Kind: cellMoney}
			assert.Equal(t, tt.want, magFormat(c, true))
		})
	}
}

func TestMagFormatBareMoney(t *testing.T) {
	// Table cells now carry $ from FormatCents. With includeUnit=false
	// the mag output strips the $ (header carries the unit instead).
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"thousands", "$5,234.23", "\U0001F8214"},
		{"hundreds", "$500.00", "\U0001F8213"},
		{"millions", "$1,000,000.00", "\U0001F8216"},
		{"tens", "$42.00", "\U0001F8212"},
		{"single digit", "$7.50", "\U0001F8211"},
		{"sub-dollar", "$0.50", "\U0001F8210"},
		{"zero", "$0.00", "\U0001F8210"},
		{"negative", "-$5.00", "-\U0001F8211"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cell{Value: tt.value, Kind: cellMoney}
			assert.Equal(t, tt.want, magFormat(c, false))
		})
	}
}

func TestMagFormatDrilldown(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"count", "42", "\U0001F8212"},
		{"zero", "0", "\U0001F8210"},
		{"large", "1000", "\U0001F8213"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cell{Value: tt.value, Kind: cellDrilldown}
			assert.Equal(t, tt.want, magFormat(c, false))
		})
	}
}

func TestMagFormatSkipsReadonly(t *testing.T) {
	c := cell{Value: "42", Kind: cellReadonly}
	assert.Equal(t, "42", magFormat(c, false))
}

func TestMagFormatSkipsNonNumericKinds(t *testing.T) {
	tests := []struct {
		name  string
		value string
		kind  cellKind
	}{
		{"text name", "Kitchen Remodel", cellText},
		{"status", "underway", cellStatus},
		{"date", "2026-02-12", cellDate},
		{"warranty date", "2027-06-15", cellWarranty},
		{"urgency date", "2026-03-01", cellUrgency},
		{"notes", "Some long note", cellNotes},
		{"empty text", "", cellText},
		{"dash money", "\u2014", cellMoney},
		{"readonly id", "7", cellReadonly},

		// Numeric-looking cellText values that must NOT be transformed:
		// phone numbers, serial numbers, model numbers, zip codes.
		{"phone number", "5551234567", cellText},
		{"formatted phone", "(555) 123-4567", cellText},
		{"serial number", "123456789", cellText},
		{"model number", "12345", cellText},
		{"zip code", "90210", cellText},
		{"interval", "3m", cellText},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cell{Value: tt.value, Kind: tt.kind}
			assert.Equal(t, tt.value, magFormat(c, false), "value should be unchanged")
		})
	}
}

func TestMagTransformCells(t *testing.T) {
	rows := [][]cell{
		{
			{Value: "1", Kind: cellReadonly},
			{Value: "Kitchen Remodel", Kind: cellText},
			{Value: "$5,234.23", Kind: cellMoney},
			{Value: "3", Kind: cellDrilldown},
		},
		{
			{Value: "2", Kind: cellReadonly},
			{Value: "Deck", Kind: cellText},
			{Value: "$100.00", Kind: cellMoney},
			{Value: "0", Kind: cellDrilldown},
		},
	}
	out := magTransformCells(rows)

	// ID cells unchanged.
	assert.Equal(t, "1", out[0][0].Value)
	assert.Equal(t, "2", out[1][0].Value)

	// Text cells unchanged.
	assert.Equal(t, "Kitchen Remodel", out[0][1].Value)
	assert.Equal(t, "Deck", out[1][1].Value)

	// Money cells: magnitude only ($ stripped by transform).
	assert.Equal(t, "\U0001F8214", out[0][2].Value)
	assert.Equal(t, "\U0001F8212", out[1][2].Value)

	// Drilldown counts transformed.
	assert.Equal(t, "\U0001F8210", out[0][3].Value)
	assert.Equal(t, "\U0001F8210", out[1][3].Value)

	// Original rows are not modified.
	assert.Equal(t, "$5,234.23", rows[0][2].Value)
}

func TestMagModeToggle(t *testing.T) {
	m := newTestModel()
	assert.False(t, m.magMode)
	sendKey(m, "ctrl+m")
	assert.True(t, m.magMode)
	sendKey(m, "ctrl+m")
	assert.False(t, m.magMode)
}

func TestMagModeWorksInEditMode(t *testing.T) {
	m := newTestModel()
	m.enterEditMode()
	assert.False(t, m.magMode)
	sendKey(m, "ctrl+m")
	assert.True(t, m.magMode)
}

func TestMagCentsIncludesUnit(t *testing.T) {
	assert.Equal(t, "$ \U0001F8214", magCents(523423))
	assert.Equal(t, "$ \U0001F8213", magCents(50000))
	assert.Equal(t, "$ \U0001F8210", magCents(100))
}

func TestMagOptionalCentsNil(t *testing.T) {
	assert.Equal(t, "", magOptionalCents(nil))
}

func TestMagOptionalCentsPresent(t *testing.T) {
	cents := int64(100000)
	assert.Equal(t, "$ \U0001F8213", magOptionalCents(&cents))
}
