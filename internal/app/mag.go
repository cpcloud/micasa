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

// magValue converts a numeric cell value to order-of-magnitude notation.
// Non-numeric values are returned unchanged. Dollar prefixes are preserved.
func magValue(c cell) string {
	value := strings.TrimSpace(c.Value)
	if value == "" || value == "\u2014" {
		return value
	}

	// Only transform kinds that carry meaningful numeric data.
	// Skip cellReadonly (IDs, ages, counts) and non-numeric kinds.
	switch c.Kind {
	case cellText, cellMoney, cellDrilldown:
		// Potentially numeric; continue to parsing below.
	case cellReadonly, cellDate, cellWarranty, cellUrgency, cellNotes, cellStatus:
		return value
	}

	prefix := ""
	numStr := value

	// Handle negative money: "-$123.45"
	if strings.HasPrefix(numStr, "-$") {
		prefix = "-$ "
		numStr = numStr[2:]
	} else if strings.HasPrefix(numStr, "$") {
		prefix = "$ "
		numStr = numStr[1:]
	}

	numStr = strings.ReplaceAll(numStr, ",", "")

	f, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return value
	}

	if f == 0 {
		return prefix + magArrow + "0"
	}

	mag := int(math.Floor(math.Log10(math.Abs(f))))
	return fmt.Sprintf("%s%s%d", prefix, magArrow, mag)
}

// magCents converts a cent amount to magnitude notation (e.g. 523423 → "$🠡 3").
func magCents(cents int64) string {
	return magValue(cell{Value: data.FormatCents(cents), Kind: cellMoney})
}

// magOptionalCents converts an optional cent amount to magnitude notation.
func magOptionalCents(cents *int64) string {
	if cents == nil {
		return ""
	}
	return magCents(*cents)
}

// magTransformCells returns a copy of the cell grid with numeric values
// replaced by their order-of-magnitude representation.
func magTransformCells(rows [][]cell) [][]cell {
	out := make([][]cell, len(rows))
	for i, row := range rows {
		transformed := make([]cell, len(row))
		for j, c := range row {
			transformed[j] = cell{
				Value:  magValue(c),
				Kind:   c.Kind,
				LinkID: c.LinkID,
			}
		}
		out[i] = transformed
	}
	return out
}
