// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import "testing"

func TestNaturalWidthsIgnoreMax(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID", Min: 4, Max: 6},
		{Title: "Name", Min: 8, Max: 12},
	}
	rows := [][]cell{
		{{Value: "1"}, {Value: "A very long name indeed"}},
	}
	natural := naturalWidths(specs, rows)
	// "A very long name indeed" is 23 chars, well past Max of 12.
	if natural[1] <= 12 {
		t.Fatalf("expected natural width > Max (12), got %d", natural[1])
	}
}

func TestColumnWidthsNoTruncationWhenRoomAvailable(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID", Min: 4, Max: 6},
		{Title: "Name", Min: 8, Max: 12},
	}
	rows := [][]cell{
		{{Value: "1"}, {Value: "A long name here"}},
	}
	// "A long name here" = 16 chars, exceeds Max=12.
	// With 200 width and 3 separator, natural widths should fit.
	widths := columnWidths(specs, rows, 200, 3)
	if widths[1] < 16 {
		t.Fatalf(
			"expected Name column >= 16 (content width), got %d",
			widths[1],
		)
	}
}

func TestColumnWidthsTruncatesWhenTerminalNarrow(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID", Min: 4, Max: 6},
		{Title: "Name", Min: 8, Max: 12, Flex: true},
	}
	rows := [][]cell{
		{{Value: "1"}, {Value: "A very long name indeed"}},
	}
	// Very narrow terminal: 20 total - 3 separator = 17 available.
	widths := columnWidths(specs, rows, 20, 3)
	total := widths[0] + widths[1]
	if total > 17 {
		t.Fatalf("expected total widths <= 17, got %d", total)
	}
}

func TestColumnWidthsTruncatedColumnsGetExtraFirst(t *testing.T) {
	specs := []columnSpec{
		{Title: "ID", Min: 4, Max: 6},
		{Title: "Name", Min: 8, Max: 10},
		{Title: "Desc", Min: 8, Max: 10, Flex: true},
	}
	rows := [][]cell{
		{{Value: "1"}, {Value: "Fifteen chars!!"}, {Value: "short"}},
	}
	// Natural: ID=4, Name=15, Desc=8 = 27 total.
	// Available: 60 - 6 (two separators of 3) = 54.
	// Natural fits (27 < 54), so no truncation needed.
	widths := columnWidths(specs, rows, 60, 3)
	if widths[1] < 15 {
		t.Fatalf(
			"expected Name >= 15 (no truncation when room available), got %d",
			widths[1],
		)
	}
}

func TestWidenTruncated(t *testing.T) {
	widths := []int{4, 10, 8}
	natural := []int{4, 15, 8}
	remaining := widenTruncated(widths, natural, 3)
	// Should widen column 1 from 10 to 13 (3 extra given).
	if widths[1] != 13 {
		t.Fatalf("expected widths[1]=13 after widening, got %d", widths[1])
	}
	if remaining != 0 {
		t.Fatalf("expected 0 remaining, got %d", remaining)
	}
}

func TestWidenTruncatedCapsAtNatural(t *testing.T) {
	widths := []int{4, 10, 8}
	natural := []int{4, 12, 8}
	remaining := widenTruncated(widths, natural, 5)
	// Column 1 needs 2 more to reach natural. 5 - 2 = 3 remaining.
	if widths[1] != 12 {
		t.Fatalf("expected widths[1]=12 (natural), got %d", widths[1])
	}
	if remaining != 3 {
		t.Fatalf("expected 3 remaining, got %d", remaining)
	}
}

func TestColumnWidthsFixedValuesStillStabilize(t *testing.T) {
	specs := []columnSpec{
		{Title: "Status", Min: 8, Max: 12, FixedValues: []string{
			"ideating", "planned", "underway", "completed", "abandoned",
		}},
	}
	rows := [][]cell{
		{{Value: "planned"}},
	}
	// Even with only "planned" displayed, column should be wide enough
	// for the longest fixed value ("abandoned" = 9, "completed" = 9).
	widths := columnWidths(specs, rows, 80, 3)
	if widths[0] < 9 {
		t.Fatalf("expected width >= 9 (longest fixed value), got %d", widths[0])
	}
}
