// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// gapSeparators computes a per-gap separator for the header/data and divider.
// Gaps between visible columns that have hidden columns in between use ⋯ to
// signal a collapsed region. Returns one separator per gap (len(visToFull)-1).
func gapSeparators(
	visToFull []int,
	totalCols int,
	normalSep string,
	styles Styles,
) (plainSeps, collapsedSeps []string) {
	n := len(visToFull)
	if n <= 1 {
		return nil, nil
	}
	collapsedSep := styles.TableSeparator.Render(" ") +
		lipgloss.NewStyle().Foreground(secondary).Render("⋯") +
		styles.TableSeparator.Render(" ")

	plainSeps = make([]string, n-1)
	collapsedSeps = make([]string, n-1)
	for i := 0; i < n-1; i++ {
		plainSeps[i] = normalSep
		if visToFull[i+1] > visToFull[i]+1 {
			collapsedSeps[i] = collapsedSep
		} else {
			collapsedSeps[i] = normalSep
		}
	}
	return
}

// hiddenColumnNames returns the titles of all hidden columns.
func hiddenColumnNames(specs []columnSpec) []string {
	var names []string
	for _, s := range specs {
		if s.HideOrder > 0 {
			names = append(names, s.Title)
		}
	}
	return names
}

// renderHiddenBadges renders a single left-aligned line of hidden column
// names. Color indicates position relative to the cursor: HiddenLeft for
// columns to the left (with a trailing left-arrow), HiddenRight for columns
// to the right (with a leading right-arrow).
func renderHiddenBadges(
	specs []columnSpec,
	colCursor int,
	styles Styles,
) string {
	sep := styles.HeaderHint.Render(" · ")

	var parts []string
	for i, spec := range specs {
		if spec.HideOrder == 0 {
			continue
		}
		if i < colCursor {
			parts = append(parts, styles.HiddenLeft.Render(spec.Title+" <"))
		} else {
			parts = append(parts, styles.HiddenRight.Render("> "+spec.Title))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, sep)
}
