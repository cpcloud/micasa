// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"strings"

	"github.com/cpcloud/micasa/internal/data"
)

// compactMoneyCells returns a copy of the cell grid with money values
// replaced by their compact representation (e.g. "$1.2k"). The original
// cells are not modified so sorting continues to work on full-precision
// values.
func compactMoneyCells(rows [][]cell) [][]cell {
	out := make([][]cell, len(rows))
	for i, row := range rows {
		transformed := make([]cell, len(row))
		for j, c := range row {
			if c.Kind == cellMoney {
				transformed[j] = cell{
					Value:  compactMoneyValue(c.Value),
					Kind:   c.Kind,
					LinkID: c.LinkID,
				}
			} else {
				transformed[j] = c
			}
		}
		out[i] = transformed
	}
	return out
}

// compactMoneyValue converts a full-precision money string like "$1,234.56"
// or "1,234.56" to compact form using data.FormatCompactCents. Returns the
// original value unchanged if it can't be parsed.
func compactMoneyValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "—" {
		return v
	}
	cents, err := data.ParseRequiredCents(v)
	if err != nil {
		return v
	}
	return data.FormatCompactCents(cents)
}
