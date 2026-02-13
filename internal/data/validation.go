// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

const DateLayout = "2006-01-02"

var (
	ErrInvalidMoney  = errors.New("invalid money value")
	ErrNegativeMoney = errors.New("negative money value")
	ErrInvalidDate   = errors.New("invalid date value")
	ErrInvalidInt    = errors.New("invalid integer value")
	ErrInvalidFloat  = errors.New("invalid decimal value")
)

func ParseRequiredCents(input string) (int64, error) {
	cents, err := parseCents(strings.TrimSpace(input))
	if err != nil {
		return 0, err
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
		return nil, err
	}
	return &cents, nil
}

func FormatCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		// Special case: math.MinInt64 cannot be negated without overflow.
		// Treat it as math.MaxInt64 for display purposes (off by one cent is
		// acceptable for an impossibly large negative value).
		if cents == math.MinInt64 {
			cents = math.MaxInt64
		} else {
			cents = -cents
		}
	}
	dollars := cents / 100
	remainder := cents % 100
	return fmt.Sprintf("%s$%s.%02d", sign, humanize.Comma(dollars), remainder)
}

func FormatOptionalCents(cents *int64) string {
	if cents == nil {
		return ""
	}
	return FormatCents(*cents)
}

// FormatCompactCents formats cents using abbreviated notation for large
// values: $1.2k, $45k, $1.3M. Values under $1,000 use full precision.
// Uses go-humanize for SI prefix formatting.
func FormatCompactCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		// Special case: math.MinInt64 cannot be negated without overflow.
		if cents == math.MinInt64 {
			cents = math.MaxInt64
		} else {
			cents = -cents
		}
	}
	dollars := float64(cents) / 100.0
	if dollars < 1000 {
		return fmt.Sprintf(
			"%s$%s.%02d",
			sign,
			humanize.Comma(cents/100),
			cents%100,
		)
	}
	// SIWithDigits produces "1.2 k" -- strip the space between number and suffix.
	si := humanize.SIWithDigits(dollars, 1, "")
	si = strings.Replace(si, " ", "", 1)
	return sign + "$" + si
}

// FormatCompactOptionalCents formats optional cents compactly.
func FormatCompactOptionalCents(cents *int64) string {
	if cents == nil {
		return ""
	}
	return FormatCompactCents(*cents)
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
	// Reject negative values -- all money fields are costs/fees/budgets.
	if strings.HasPrefix(clean, "-") {
		return 0, ErrNegativeMoney
	}
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
	// Guard against overflow: wholePart*100 + frac must fit in int64.
	const maxDollars = math.MaxInt64 / 100
	if wholePart > maxDollars {
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
	cents := wholePart*100 + frac
	// Final overflow check: frac can push past MaxInt64 when wholePart == maxDollars.
	if cents < 0 {
		return 0, ErrInvalidMoney
	}
	return cents, nil
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
