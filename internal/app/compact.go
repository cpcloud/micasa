// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"strings"

	"github.com/cpcloud/micasa/internal/data"
)

// statusLabels maps full status names to short display labels.
var statusLabels = map[string]string{
	"ideating":  "idea",
	"planned":   "plan",
	"quoted":    "bid",
	"underway":  "wip",
	"delayed":   "hold",
	"completed": "done",
	"abandoned": "drop",
}

// statusLabel returns the short display label for a status value.
func statusLabel(status string) string {
	if label, ok := statusLabels[status]; ok {
		return label
	}
	return status
}

// annotateMoneyHeaders returns a copy of specs with a styled green "$"
// appended to money column titles. The unit lives in the header so cell
// values can be bare numbers.
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

// compactMoneyCells returns a copy of the cell grid with money values
// replaced by their compact representation (e.g. "1.2k") without the $
// prefix (the $ lives in the column header). The original cells are not
// modified so sorting continues to work on full-precision values.
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
// to compact form without the $ prefix (e.g. "5.2k", "100.00"). The $
// prefix is handled by the column header annotation instead.
func compactMoneyValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "—" {
		return v
	}
	cents, err := data.ParseRequiredCents(v)
	if err != nil {
		return v
	}
	compact := data.FormatCompactCents(cents)
	return strings.TrimPrefix(compact, "$")
}
