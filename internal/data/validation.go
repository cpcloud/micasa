// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"

var (
	ErrInvalidMoney = errors.New("invalid money value")
	ErrInvalidDate  = errors.New("invalid date value")
	ErrInvalidInt   = errors.New("invalid integer value")
	ErrInvalidFloat = errors.New("invalid decimal value")
)

func ParseRequiredCents(input string) (int64, error) {
	cents, err := parseCents(strings.TrimSpace(input))
	if err != nil {
		return 0, ErrInvalidMoney
	}
	return cents, nil
}

func ParseOptionalCents(input string) (*int64, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, nil
	}
	cents, err := parseCents(trimmed)
	if err != nil {
		return nil, ErrInvalidMoney
	}
	return &cents, nil
}

func FormatCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	dollars := cents / 100
	remainder := cents % 100
	return fmt.Sprintf("%s$%s.%02d", sign, formatWithCommas(dollars), remainder)
}

func FormatOptionalCents(cents *int64) string {
	if cents == nil {
		return ""
	}
	return FormatCents(*cents)
}

// FormatCompactCents formats cents using abbreviated notation for large
// values: $1.2k, $45k, $1.3M. Values under $1,000 use full precision.
func FormatCompactCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	dollars := float64(cents) / 100.0
	switch {
	case dollars >= 1_000_000:
		return sign + compactDollars(dollars/1_000_000, "M")
	case dollars >= 1_000:
		return sign + compactDollars(dollars/1_000, "k")
	default:
		d := cents / 100
		r := cents % 100
		return fmt.Sprintf("%s$%d.%02d", sign, d, r)
	}
}

// FormatCompactOptionalCents formats optional cents compactly.
func FormatCompactOptionalCents(cents *int64) string {
	if cents == nil {
		return ""
	}
	return FormatCompactCents(*cents)
}

func compactDollars(v float64, suffix string) string {
	// Drop the decimal when it's a whole number (e.g., $45k not $45.0k).
	if v == float64(int64(v)) {
		return fmt.Sprintf("$%.0f%s", v, suffix)
	}
	return fmt.Sprintf("$%.1f%s", v, suffix)
}

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
	next := last.AddDate(0, intervalMonths, 0)
	return &next
}

func parseCents(input string) (int64, error) {
	clean := strings.ReplaceAll(input, ",", "")
	clean = strings.TrimPrefix(clean, "$")
	if clean == "" {
		return 0, ErrInvalidMoney
	}
	parts := strings.Split(clean, ".")
	if len(parts) > 2 {
		return 0, ErrInvalidMoney
	}
	wholePart, err := parseDigits(parts[0], true)
	if err != nil {
		return 0, ErrInvalidMoney
	}
	frac := int64(0)
	if len(parts) == 2 {
		if len(parts[1]) > 2 {
			return 0, ErrInvalidMoney
		}
		frac, err = parseDigits(parts[1], false)
		if err != nil {
			return 0, ErrInvalidMoney
		}
		if len(parts[1]) == 1 {
			frac *= 10
		}
	}
	return wholePart*100 + frac, nil
}

func parseDigits(input string, allowEmpty bool) (int64, error) {
	if input == "" {
		if allowEmpty {
			return 0, nil
		}
		return 0, ErrInvalidMoney
	}
	for _, r := range input {
		if r < '0' || r > '9' {
			return 0, ErrInvalidMoney
		}
	}
	return strconv.ParseInt(input, 10, 64)
}

func formatWithCommas(value int64) string {
	raw := strconv.FormatInt(value, 10)
	if len(raw) <= 3 {
		return raw
	}
	var out strings.Builder
	mod := len(raw) % 3
	if mod > 0 {
		out.WriteString(raw[:mod])
		if len(raw) > mod {
			out.WriteString(",")
		}
	}
	for i := mod; i < len(raw); i += 3 {
		out.WriteString(raw[i : i+3])
		if i+3 < len(raw) {
			out.WriteString(",")
		}
	}
	return out.String()
}
