// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"testing"
	"time"
)

func TestParseRequiredCents(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"100", 10000},
		{"100.5", 10050},
		{"100.05", 10005},
		{"$1,234.56", 123456},
		{".75", 75},
		{"0.99", 99},
	}
	for _, test := range tests {
		got, err := ParseRequiredCents(test.input)
		if err != nil {
			t.Fatalf("ParseRequiredCents(%q) returned error: %v", test.input, err)
		}
		if got != test.want {
			t.Fatalf("ParseRequiredCents(%q) = %d, want %d", test.input, got, test.want)
		}
	}
}

func TestParseRequiredCentsInvalid(t *testing.T) {
	inputs := []string{"", "12.345", "abc", "1.2.3"}
	for _, input := range inputs {
		if _, err := ParseRequiredCents(input); err == nil {
			t.Fatalf("ParseRequiredCents(%q) expected error", input)
		}
	}
}

func TestParseOptionalCents(t *testing.T) {
	value, err := ParseOptionalCents("")
	if err != nil {
		t.Fatalf("ParseOptionalCents empty returned error: %v", err)
	}
	if value != nil {
		t.Fatalf("ParseOptionalCents empty expected nil, got %v", *value)
	}
	value, err = ParseOptionalCents("5")
	if err != nil {
		t.Fatalf("ParseOptionalCents returned error: %v", err)
	}
	if value == nil || *value != 500 {
		t.Fatalf("ParseOptionalCents = %v, want 500", value)
	}
}

func TestFormatCents(t *testing.T) {
	got := FormatCents(123456)
	if got != "$1,234.56" {
		t.Fatalf("FormatCents = %q, want $1,234.56", got)
	}
}

func TestParseOptionalDate(t *testing.T) {
	date, err := ParseOptionalDate("2025-06-11")
	if err != nil {
		t.Fatalf("ParseOptionalDate error: %v", err)
	}
	if date == nil || date.Format(DateLayout) != "2025-06-11" {
		t.Fatalf("ParseOptionalDate got %v", date)
	}
	if _, err := ParseOptionalDate("06/11/2025"); err == nil {
		t.Fatalf("ParseOptionalDate expected error")
	}
}

func TestParseOptionalInt(t *testing.T) {
	value, err := ParseOptionalInt("12")
	if err != nil || value != 12 {
		t.Fatalf("ParseOptionalInt got %d err %v", value, err)
	}
	if _, err := ParseOptionalInt("-1"); err == nil {
		t.Fatalf("ParseOptionalInt expected error for negative")
	}
}

func TestParseOptionalFloat(t *testing.T) {
	value, err := ParseOptionalFloat("2.5")
	if err != nil || value != 2.5 {
		t.Fatalf("ParseOptionalFloat got %v err %v", value, err)
	}
	if _, err := ParseOptionalFloat("-1.2"); err == nil {
		t.Fatalf("ParseOptionalFloat expected error for negative")
	}
}

func TestFormatOptionalCents(t *testing.T) {
	if got := FormatOptionalCents(nil); got != "" {
		t.Fatalf("FormatOptionalCents(nil) = %q, want empty", got)
	}
	cents := int64(123456)
	if got := FormatOptionalCents(&cents); got != "$1,234.56" {
		t.Fatalf("FormatOptionalCents = %q, want $1,234.56", got)
	}
}

func TestFormatCentsNegative(t *testing.T) {
	got := FormatCents(-500)
	if got != "-$5.00" {
		t.Fatalf("FormatCents(-500) = %q, want -$5.00", got)
	}
}

func TestFormatCentsZero(t *testing.T) {
	got := FormatCents(0)
	if got != "$0.00" {
		t.Fatalf("FormatCents(0) = %q, want $0.00", got)
	}
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
		if err != nil {
			t.Fatalf("ParseRequiredDate(%q) error: %v", tt.input, err)
		}
		if got.Format(DateLayout) != tt.want {
			t.Fatalf(
				"ParseRequiredDate(%q) = %s, want %s",
				tt.input,
				got.Format(DateLayout),
				tt.want,
			)
		}
	}
}

func TestParseRequiredDateInvalid(t *testing.T) {
	for _, input := range []string{"", "06/11/2025", "not-a-date", "2025-13-01"} {
		if _, err := ParseRequiredDate(input); err == nil {
			t.Fatalf("ParseRequiredDate(%q) expected error", input)
		}
	}
}

func TestFormatDate(t *testing.T) {
	if got := FormatDate(nil); got != "" {
		t.Fatalf("FormatDate(nil) = %q, want empty", got)
	}
	d := time.Date(2025, 6, 11, 0, 0, 0, 0, time.UTC)
	if got := FormatDate(&d); got != "2025-06-11" {
		t.Fatalf("FormatDate = %q, want 2025-06-11", got)
	}
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
		if err != nil {
			t.Fatalf("ParseRequiredInt(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("ParseRequiredInt(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseRequiredIntInvalid(t *testing.T) {
	for _, input := range []string{"", "abc", "-5", "1.5"} {
		if _, err := ParseRequiredInt(input); err == nil {
			t.Fatalf("ParseRequiredInt(%q) expected error", input)
		}
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
		if err != nil {
			t.Fatalf("ParseRequiredFloat(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("ParseRequiredFloat(%q) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestParseRequiredFloatInvalid(t *testing.T) {
	for _, input := range []string{"", "abc", "-1.5"} {
		if _, err := ParseRequiredFloat(input); err == nil {
			t.Fatalf("ParseRequiredFloat(%q) expected error", input)
		}
	}
}

func TestParseOptionalIntEmpty(t *testing.T) {
	got, err := ParseOptionalInt("")
	if err != nil {
		t.Fatalf("ParseOptionalInt empty error: %v", err)
	}
	if got != 0 {
		t.Fatalf("ParseOptionalInt empty = %d, want 0", got)
	}
}

func TestParseOptionalFloatEmpty(t *testing.T) {
	got, err := ParseOptionalFloat("")
	if err != nil {
		t.Fatalf("ParseOptionalFloat empty error: %v", err)
	}
	if got != 0 {
		t.Fatalf("ParseOptionalFloat empty = %f, want 0", got)
	}
}

func TestParseOptionalDateEmpty(t *testing.T) {
	got, err := ParseOptionalDate("")
	if err != nil {
		t.Fatalf("ParseOptionalDate empty error: %v", err)
	}
	if got != nil {
		t.Fatalf("ParseOptionalDate empty = %v, want nil", got)
	}
}

func TestParseOptionalCentsInvalid(t *testing.T) {
	if _, err := ParseOptionalCents("abc"); err == nil {
		t.Fatal("expected error for invalid money")
	}
}

func TestComputeNextDue(t *testing.T) {
	last := time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)
	next := ComputeNextDue(&last, 6)
	if next == nil {
		t.Fatalf("ComputeNextDue returned nil")
	}
	if next.Format(DateLayout) != "2025-04-10" {
		t.Fatalf("ComputeNextDue got %s", next.Format(DateLayout))
	}
}

func TestComputeNextDueNilDate(t *testing.T) {
	if got := ComputeNextDue(nil, 6); got != nil {
		t.Fatalf("expected nil for nil date, got %v", got)
	}
}

func TestComputeNextDueZeroInterval(t *testing.T) {
	d := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := ComputeNextDue(&d, 0); got != nil {
		t.Fatalf("expected nil for zero interval, got %v", got)
	}
}
