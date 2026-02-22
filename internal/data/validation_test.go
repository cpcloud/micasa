// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOptionalDate(t *testing.T) {
	date, err := ParseOptionalDate("2025-06-11")
	require.NoError(t, err)
	require.NotNil(t, date)
	assert.Equal(t, "2025-06-11", date.Format(DateLayout))

	_, err = ParseOptionalDate("06/11/2025")
	assert.Error(t, err)
}

func TestParseOptionalInt(t *testing.T) {
	value, err := ParseOptionalInt("12")
	require.NoError(t, err)
	assert.Equal(t, 12, value)

	_, err = ParseOptionalInt("-1")
	assert.Error(t, err)
}

func TestParseOptionalFloat(t *testing.T) {
	value, err := ParseOptionalFloat("2.5")
	require.NoError(t, err)
	assert.Equal(t, 2.5, value)

	_, err = ParseOptionalFloat("-1.2")
	assert.Error(t, err)
}

func TestParseRequiredDate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2025-06-11", "2025-06-11"},
		{" 2025-06-11 ", "2025-06-11"},
	}
	for _, tt := range tests {
		got, err := ParseRequiredDate(tt.input)
		require.NoError(t, err, "input=%q", tt.input)
		assert.Equal(t, tt.want, got.Format(DateLayout), "input=%q", tt.input)
	}
}

func TestParseRequiredDateInvalid(t *testing.T) {
	for _, input := range []string{"", "06/11/2025", "not-a-date", "2025-13-01"} {
		_, err := ParseRequiredDate(input)
		assert.Error(t, err, "input=%q", input)
	}
}

func TestFormatDate(t *testing.T) {
	assert.Empty(t, FormatDate(nil))
	d := time.Date(2025, 6, 11, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, "2025-06-11", FormatDate(&d))
}

func TestParseRequiredInt(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"42", 42},
		{" 7 ", 7},
		{"0", 0},
	}
	for _, tt := range tests {
		got, err := ParseRequiredInt(tt.input)
		require.NoError(t, err, "input=%q", tt.input)
		assert.Equal(t, tt.want, got, "input=%q", tt.input)
	}
}

func TestParseRequiredIntInvalid(t *testing.T) {
	for _, input := range []string{"", "abc", "-5", "1.5"} {
		_, err := ParseRequiredInt(input)
		assert.Error(t, err, "input=%q", input)
	}
}

func TestParseRequiredFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"2.5", 2.5},
		{" 0 ", 0},
		{"100", 100},
	}
	for _, tt := range tests {
		got, err := ParseRequiredFloat(tt.input)
		require.NoError(t, err, "input=%q", tt.input)
		assert.Equal(t, tt.want, got, "input=%q", tt.input)
	}
}

func TestParseRequiredFloatInvalid(t *testing.T) {
	for _, input := range []string{"", "abc", "-1.5"} {
		_, err := ParseRequiredFloat(input)
		assert.Error(t, err, "input=%q", input)
	}
}

func TestParseOptionalIntEmpty(t *testing.T) {
	got, err := ParseOptionalInt("")
	require.NoError(t, err)
	assert.Zero(t, got)
}

func TestParseOptionalFloatEmpty(t *testing.T) {
	got, err := ParseOptionalFloat("")
	require.NoError(t, err)
	assert.Zero(t, got)
}

func TestParseOptionalDateEmpty(t *testing.T) {
	got, err := ParseOptionalDate("")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestComputeNextDue(t *testing.T) {
	last := time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)
	next := ComputeNextDue(&last, 6)
	require.NotNil(t, next)
	assert.Equal(t, "2025-04-10", next.Format(DateLayout))
}

func TestComputeNextDueNilDate(t *testing.T) {
	assert.Nil(t, ComputeNextDue(nil, 6))
}

func TestComputeNextDueZeroInterval(t *testing.T) {
	d := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Nil(t, ComputeNextDue(&d, 0))
}

func TestAddMonths(t *testing.T) {
	tests := []struct {
		name   string
		start  time.Time
		months int
		want   string
	}{
		{
			"Jan 31 + 1 month = Feb 28 (non-leap year)",
			time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), 1,
			"2025-02-28",
		},
		{
			"Jan 31 + 1 month = Feb 29 (leap year)",
			time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), 1,
			"2024-02-29",
		},
		{
			"Mar 31 + 1 month = Apr 30",
			time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC), 1,
			"2025-04-30",
		},
		{
			"normal case: Jan 15 + 1 month = Feb 15",
			time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), 1,
			"2025-02-15",
		},
		{
			"multiple months: Jan 31 + 3 months = Apr 30",
			time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), 3,
			"2025-04-30",
		},
		{
			"year wrap: Nov 30 + 3 months = Feb 28",
			time.Date(2024, 11, 30, 0, 0, 0, 0, time.UTC), 3,
			"2025-02-28",
		},
		{
			"Feb 29 (leap) + 12 months = Feb 28 (non-leap)",
			time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), 12,
			"2025-02-28",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddMonths(tt.start, tt.months)
			assert.Equal(t, tt.want, got.Format(DateLayout))
		})
	}
}

func TestParseIntervalMonths(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		// bare integers
		{"12", 12},
		{"0", 0},
		{"  7  ", 7},
		// month suffix
		{"6m", 6},
		{"6M", 6},
		{" 3m ", 3},
		// year suffix
		{"1y", 12},
		{"2Y", 24},
		{" 1y ", 12},
		// combined
		{"2y 6m", 30},
		{"1y6m", 18},
		{"1Y 3M", 15},
		{"  2y  6m  ", 30},
		// empty
		{"", 0},
		{"   ", 0},
	}
	for _, tt := range tests {
		got, err := ParseIntervalMonths(tt.input)
		require.NoError(t, err, "input=%q", tt.input)
		assert.Equal(t, tt.want, got, "input=%q", tt.input)
	}
}

func TestParseIntervalMonthsInvalid(t *testing.T) {
	for _, input := range []string{"abc", "-1", "1.5m", "1x", "m", "y", "6m 1y"} {
		_, err := ParseIntervalMonths(input)
		assert.Error(t, err, "input=%q should be rejected", input)
	}
}

func TestComputeNextDueMonthEndClamping(t *testing.T) {
	last := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	next := ComputeNextDue(&last, 1)
	require.NotNil(t, next)
	assert.Equal(t, "2025-02-28", next.Format(DateLayout))
}
