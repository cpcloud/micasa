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
