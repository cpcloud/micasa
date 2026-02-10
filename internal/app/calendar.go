// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// calendarState tracks the date picker overlay.
type calendarState struct {
	Cursor    time.Time // the date the cursor is on
	Selected  time.Time // the date the field currently has (dim highlight)
	HasValue  bool      // whether Selected is meaningful
	FieldPtr  *string   // pointer to the form field's value string
	OnConfirm func()    // called after writing the picked date to FieldPtr
}

// calendarGrid renders a single month calendar with the cursor highlighted.
func calendarGrid(cal calendarState, styles Styles) string {
	cursor := cal.Cursor
	year, month := cursor.Year(), cursor.Month()

	// Header: month name + year.
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(accent).
		Render(fmt.Sprintf(" %s %d ", month.String(), year))

	// Day-of-week labels.
	dayLabels := lipgloss.NewStyle().
		Foreground(textDim).
		Render("Su Mo Tu We Th Fr Sa")

	// Build the day grid.
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	startDow := int(first.Weekday()) // 0=Sun
	daysInMonth := daysIn(year, month)

	var grid strings.Builder
	// Leading blanks.
	for i := 0; i < startDow; i++ {
		grid.WriteString("   ")
	}

	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		label := fmt.Sprintf("%2d", day)

		isCursor := sameDay(date, cursor)
		isSelected := cal.HasValue && sameDay(date, cal.Selected)
		isToday := sameDay(date, time.Now())

		var style lipgloss.Style
		switch {
		case isCursor:
			style = styles.CalCursor
		case isSelected:
			style = styles.CalSelected
		case isToday:
			style = styles.CalToday
		default:
			style = lipgloss.NewStyle()
		}

		grid.WriteString(style.Render(label))

		dow := (startDow + day - 1) % 7
		if dow == 6 && day < daysInMonth {
			grid.WriteString("\n")
		} else if day < daysInMonth {
			grid.WriteString(" ")
		}
	}

	// Navigation hints split into two lines to keep the box compact.
	hintStyle := lipgloss.NewStyle().Foreground(textDim)
	hints := lipgloss.JoinVertical(lipgloss.Left,
		hintStyle.Render("h/l day · j/k week · H/L month · ctrl+shift+h/l year"),
		hintStyle.Render("enter pick · esc cancel"),
	)

	// The grid is 20 chars wide ("Su Mo Tu We Th Fr Sa"); center everything
	// to the wider of the grid and hints so the box looks balanced.
	calW := lipgloss.Width(dayLabels)
	hintW := lipgloss.Width(hints)
	boxW := calW
	if hintW > boxW {
		boxW = hintW
	}

	center := func(s string) string {
		return lipgloss.PlaceHorizontal(boxW, lipgloss.Center, s)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		center(header),
		"",
		center(dayLabels),
		center(grid.String()),
		"",
		center(hints),
	)
}

func daysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// calendarMove adjusts the calendar cursor by the given number of days.
func calendarMove(cal *calendarState, days int) {
	cal.Cursor = cal.Cursor.AddDate(0, 0, days)
}

// calendarMoveMonth adjusts the calendar cursor by the given number of months.
func calendarMoveMonth(cal *calendarState, months int) {
	cal.Cursor = cal.Cursor.AddDate(0, months, 0)
}

// calendarMoveYear adjusts the calendar cursor by the given number of years.
func calendarMoveYear(cal *calendarState, years int) {
	cal.Cursor = cal.Cursor.AddDate(years, 0, 0)
}
