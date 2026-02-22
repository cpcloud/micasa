// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"

var (
	ErrInvalidDate     = errors.New("invalid date value")
	ErrInvalidInt      = errors.New("invalid integer value")
	ErrInvalidFloat    = errors.New("invalid decimal value")
	ErrInvalidInterval = errors.New("invalid interval value")
)

func ParseRequiredDate(input string) (time.Time, error) {
	parsed, err := time.Parse(DateLayout, strings.TrimSpace(input))
	if err != nil {
		return time.Time{}, ErrInvalidDate
	}
	return parsed, nil
}

func ParseOptionalDate(input string) (*time.Time, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse(DateLayout, trimmed)
	if err != nil {
		return nil, ErrInvalidDate
	}
	return &parsed, nil
}

func FormatDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(DateLayout)
}

func ParseOptionalInt(input string) (int, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(trimmed)
	if err != nil || value < 0 {
		return 0, ErrInvalidInt
	}
	return value, nil
}

func ParseRequiredInt(input string) (int, error) {
	value, err := ParseOptionalInt(input)
	if err != nil || strings.TrimSpace(input) == "" {
		return 0, ErrInvalidInt
	}
	return value, nil
}

// intervalRe matches duration strings like "1y", "6m", "2y 6m", "1y6m".
var intervalRe = regexp.MustCompile(
	`(?i)^\s*(?:(\d+)\s*y)?\s*(?:(\d+)\s*m)?\s*$`,
)

// ParseIntervalMonths parses a human-friendly interval into months.
// Accepts bare integers ("12"), month suffix ("6m"), year suffix ("1y"),
// or combined ("2y 6m", "1y6m"). Case-insensitive, whitespace-flexible.
// Returns (0, nil) for empty/blank input (non-recurring).
func ParseIntervalMonths(input string) (int, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0, nil
	}
	// Try bare integer first.
	if value, err := strconv.Atoi(trimmed); err == nil {
		if value < 0 {
			return 0, ErrInvalidInterval
		}
		return value, nil
	}
	matches := intervalRe.FindStringSubmatch(trimmed)
	if matches == nil {
		return 0, ErrInvalidInterval
	}
	yearStr, monthStr := matches[1], matches[2]
	// Regex matched but both groups empty means the pattern matched
	// zero-length content -- reject.
	if yearStr == "" && monthStr == "" {
		return 0, ErrInvalidInterval
	}
	var total int
	if yearStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err != nil {
			return 0, ErrInvalidInterval
		}
		total += y * 12
	}
	if monthStr != "" {
		m, err := strconv.Atoi(monthStr)
		if err != nil {
			return 0, ErrInvalidInterval
		}
		total += m
	}
	if total < 0 {
		return 0, ErrInvalidInterval
	}
	return total, nil
}

func ParseOptionalFloat(input string) (float64, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0, nil
	}
	value, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || value < 0 {
		return 0, ErrInvalidFloat
	}
	return value, nil
}

func ParseRequiredFloat(input string) (float64, error) {
	value, err := ParseOptionalFloat(input)
	if err != nil || strings.TrimSpace(input) == "" {
		return 0, ErrInvalidFloat
	}
	return value, nil
}

func ComputeNextDue(last *time.Time, intervalMonths int) *time.Time {
	if last == nil || intervalMonths <= 0 {
		return nil
	}
	next := AddMonths(*last, intervalMonths)
	return &next
}

// AddMonths adds the given number of months to t, clamping the day to the
// last day of the target month. This avoids the time.AddDate gotcha where
// Jan 31 + 1 month = March 3 instead of Feb 28.
func AddMonths(t time.Time, months int) time.Time {
	y, m, d := t.Date()
	targetMonth := m + time.Month(months)
	// Day 0 of the NEXT month gives the last day of the target month.
	lastDay := time.Date(y, targetMonth+1, 0, 0, 0, 0, 0, t.Location()).Day()
	if d > lastDay {
		d = lastDay
	}
	return time.Date(y, targetMonth, d,
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}
