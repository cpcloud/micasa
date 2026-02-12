// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cpcloud/micasa/internal/data"
)

const magArrow = "\U0001F821" // 🠡

// magFormat converts a numeric cell value to order-of-magnitude notation.
// When includeUnit is false the dollar prefix is stripped (table cells get the
// unit from the column header instead). Non-numeric values are returned unchanged.
func magFormat(c cell, includeUnit bool) string {
	value := strings.TrimSpace(c.Value)
	if value == "" || value == "\u2014" {
		return value
	}

	// Only transform kinds that carry meaningful numeric data.
	// cellText is excluded because it covers phone numbers, serial numbers,
	// model numbers, and other identifiers that happen to look numeric.
	switch c.Kind {
	case cellMoney, cellDrilldown:
		// Definitely numeric; continue to parsing below.
	default:
		return value
	}

	sign := ""
	numStr := value

	// Strip dollar sign and detect negative.
	if strings.HasPrefix(numStr, "-$") {
		sign = "-"
		numStr = numStr[2:]
	} else if strings.HasPrefix(numStr, "$") {
		numStr = numStr[1:]
	}

	numStr = strings.ReplaceAll(numStr, ",", "")

	f, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return value
	}

	if f < 0 {
		sign = "-"
	}

	unit := ""
	if includeUnit && c.Kind == cellMoney {
		unit = "$ "
	}

	if f == 0 {
		return fmt.Sprintf("%s%s%s0", sign, unit, magArrow)
	}

	mag := int(math.Floor(math.Log10(math.Abs(f))))
	return fmt.Sprintf("%s%s%s%d", sign, unit, magArrow, mag)
}

// magCents converts a cent amount to magnitude notation with the dollar
// prefix included (for use outside of table columns, e.g. dashboard).
func magCents(cents int64) string {
	return magFormat(cell{Value: data.FormatCents(cents), Kind: cellMoney}, true)
}

// magOptionalCents converts an optional cent amount to magnitude notation.
func magOptionalCents(cents *int64) string {
	if cents == nil {
		return ""
	}
	return magCents(*cents)
}

// magTransformCells returns a copy of the cell grid with numeric values
// replaced by their order-of-magnitude representation. Dollar prefixes are
// stripped because the column header carries the unit annotation instead.
func magTransformCells(rows [][]cell) [][]cell {
	out := make([][]cell, len(rows))
	for i, row := range rows {
		transformed := make([]cell, len(row))
		for j, c := range row {
			transformed[j] = cell{
				Value:  magFormat(c, false),
				Kind:   c.Kind,
				LinkID: c.LinkID,
			}
		}
		out[i] = transformed
	}
	return out
}

// annotateMoneyHeaders returns a copy of specs with a styled "$" suffix
// on money column titles. The unit lives in the header so cell values
// can be bare numbers.
func annotateMoneyHeaders(specs []columnSpec, styles Styles) []columnSpec {
	out := make([]columnSpec, len(specs))
	copy(out, specs)
	for i, spec := range out {
		if spec.Kind == cellMoney {
			out[i].Title = spec.Title + " " + styles.Money.Render("$")
		}
	}
	return out
}
