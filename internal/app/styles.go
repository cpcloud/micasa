// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Header          lipgloss.Style
	HeaderBox       lipgloss.Style
	HeaderTitle     lipgloss.Style
	HeaderHint      lipgloss.Style
	HeaderBadge     lipgloss.Style
	HeaderSection   lipgloss.Style
	HeaderLabel     lipgloss.Style
	HeaderValue     lipgloss.Style
	Keycap          lipgloss.Style
	TabActive       lipgloss.Style
	TabInactive     lipgloss.Style
	TabLocked       lipgloss.Style
	TabUnderline    lipgloss.Style
	TableHeader     lipgloss.Style
	TableSelected   lipgloss.Style
	TableSeparator  lipgloss.Style
	ColActiveHeader lipgloss.Style
	FormClean       lipgloss.Style
	FormDirty       lipgloss.Style
	ModeNormal      lipgloss.Style
	ModeEdit        lipgloss.Style
	Money           lipgloss.Style
	Readonly        lipgloss.Style
	Drilldown       lipgloss.Style
	Empty           lipgloss.Style
	Error           lipgloss.Style
	Info            lipgloss.Style
	DeletedLabel    lipgloss.Style
	LinkIndicator   lipgloss.Style
	Breadcrumb      lipgloss.Style
	BreadcrumbArrow lipgloss.Style
	HiddenLeft      lipgloss.Style // hidden cols to the left of cursor
	HiddenRight     lipgloss.Style // hidden cols to the right of cursor
	DashSubtitle    lipgloss.Style // dashboard subtitle (house name, date)
	DashSection     lipgloss.Style // dashboard section header
	DashSectionWarn lipgloss.Style // dashboard section header (overdue/warning)
	DashRule        lipgloss.Style // dashboard horizontal rule
	DashLabel       lipgloss.Style // dashboard dim label text
	DashValue       lipgloss.Style // dashboard bright value text
	DashOverdue     lipgloss.Style // overdue maintenance item
	DashUpcoming    lipgloss.Style // upcoming maintenance (within 30 days)
	CalCursor       lipgloss.Style // calendar: cursor day
	CalSelected     lipgloss.Style // calendar: previously selected day
	CalToday        lipgloss.Style // calendar: today marker
	StatusStyles    map[string]lipgloss.Style
	StatusIcons     map[string]string
}

// Colorblind-safe palette (Wong) with adaptive light/dark variants.
//
// Each color uses lipgloss.AdaptiveColor{Light, Dark} so the UI looks
// correct on both dark and light terminal backgrounds. The Light values
// are darkened/saturated versions of the Dark values to maintain contrast
// on white backgrounds.
//
// Chromatic roles:
//   Primary accent:   sky blue     Dark #56B4E9  Light #0072B2
//   Secondary accent: orange       Dark #E69F00  Light #D55E00
//   Success/positive: bluish green Dark #009E73  Light #007A5A
//   Warning:          yellow       Dark #F0E442  Light #B8860B
//   Error/danger:     vermillion   Dark #D55E00  Light #CC3311
//   Muted accent:     rose         Dark #CC79A7  Light #AA4499
//
// Neutral roles:
//   Text bright:      Dark #E5E7EB  Light #1F2937
//   Text mid:         Dark #9CA3AF  Light #4B5563
//   Text dim:         Dark #6B7280  Light #6B7280
//   Surface:          Dark #1F2937  Light #F3F4F6
//   Surface deep:     Dark #111827  Light #E5E7EB
//   On-accent text:   Dark #0F172A  Light #FFFFFF

var (
	accent    = lipgloss.AdaptiveColor{Light: "#0072B2", Dark: "#56B4E9"}
	secondary = lipgloss.AdaptiveColor{Light: "#D55E00", Dark: "#E69F00"}
	success   = lipgloss.AdaptiveColor{Light: "#007A5A", Dark: "#009E73"}
	warning   = lipgloss.AdaptiveColor{Light: "#B8860B", Dark: "#F0E442"}
	danger    = lipgloss.AdaptiveColor{Light: "#CC3311", Dark: "#D55E00"}
	muted     = lipgloss.AdaptiveColor{Light: "#AA4499", Dark: "#CC79A7"}

	textBright = lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#E5E7EB"}
	textMid    = lipgloss.AdaptiveColor{Light: "#4B5563", Dark: "#9CA3AF"}
	textDim    = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#6B7280"}
	surface    = lipgloss.AdaptiveColor{Light: "#F3F4F6", Dark: "#1F2937"}
	onAccent   = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#0F172A"}
	border     = lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#374151"}
)

func DefaultStyles() Styles {
	return Styles{
		Header: lipgloss.NewStyle().Bold(true),
		HeaderBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(0, 1),
		HeaderTitle: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(accent).
			Padding(0, 1).
			Bold(true),
		HeaderHint: lipgloss.NewStyle().
			Foreground(textMid),
		HeaderBadge: lipgloss.NewStyle().
			Foreground(textBright).
			Background(surface).
			Padding(0, 1),
		HeaderSection: lipgloss.NewStyle().
			Foreground(textBright).
			Background(border).
			Padding(0, 1).
			Bold(true),
		HeaderLabel: lipgloss.NewStyle().
			Foreground(textDim),
		HeaderValue: lipgloss.NewStyle().
			Foreground(secondary),
		Keycap: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(textBright).
			Padding(0, 1).
			Bold(true),
		TabActive: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(accent).
			Padding(0, 1).
			Bold(true),
		TabInactive: lipgloss.NewStyle().
			Foreground(textMid).
			Padding(0, 1),
		TabLocked: lipgloss.NewStyle().
			Foreground(textDim).
			Padding(0, 1).
			Strikethrough(true),
		TabUnderline: lipgloss.NewStyle().
			Foreground(accent),
		TableHeader: lipgloss.NewStyle().
			Foreground(textDim).
			Bold(true),
		TableSelected: lipgloss.NewStyle().
			Background(surface).
			Bold(true),
		TableSeparator: lipgloss.NewStyle().
			Foreground(border),
		ColActiveHeader: lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true),
		FormClean: lipgloss.NewStyle().
			Foreground(textMid),
		FormDirty: lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true),
		ModeNormal: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(accent).
			Padding(0, 1).
			Bold(true),
		ModeEdit: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(secondary).
			Padding(0, 1).
			Bold(true),
		Money: lipgloss.NewStyle().
			Foreground(success),
		Readonly: lipgloss.NewStyle().
			Foreground(textDim),
		Drilldown: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(accent).
			Bold(true).
			Padding(0, 1),
		Empty: lipgloss.NewStyle().
			Foreground(textDim),
		Error: lipgloss.NewStyle().
			Foreground(danger).
			Bold(true),
		Info: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),
		DeletedLabel: lipgloss.NewStyle().
			Foreground(danger),
		LinkIndicator: lipgloss.NewStyle().
			Foreground(muted),
		Breadcrumb: lipgloss.NewStyle().
			Foreground(textBright).
			Bold(true),
		BreadcrumbArrow: lipgloss.NewStyle().
			Foreground(accent),
		HiddenLeft: lipgloss.NewStyle().
			Foreground(secondary).
			Italic(true),
		HiddenRight: lipgloss.NewStyle().
			Foreground(accent).
			Italic(true),
		DashSubtitle: lipgloss.NewStyle().
			Foreground(textDim).
			Italic(true),
		DashSection: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(accent).
			Padding(0, 1).
			Bold(true),
		DashSectionWarn: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(danger).
			Padding(0, 1).
			Bold(true),
		DashRule: lipgloss.NewStyle().
			Foreground(border),
		DashLabel: lipgloss.NewStyle().
			Foreground(textDim),
		DashValue: lipgloss.NewStyle().
			Foreground(textBright),
		DashOverdue: lipgloss.NewStyle().
			Foreground(danger).
			Bold(true),
		DashUpcoming: lipgloss.NewStyle().
			Foreground(warning),
		CalCursor: lipgloss.NewStyle().
			Background(accent).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"}).
			Bold(true),
		CalSelected: lipgloss.NewStyle().
			Foreground(secondary).
			Underline(true),
		CalToday: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),
		StatusStyles: map[string]lipgloss.Style{
			"ideating":  lipgloss.NewStyle().Foreground(muted),
			"planned":   lipgloss.NewStyle().Foreground(accent),
			"quoted":    lipgloss.NewStyle().Foreground(secondary),
			"underway":  lipgloss.NewStyle().Foreground(success),
			"delayed":   lipgloss.NewStyle().Foreground(warning),
			"completed": lipgloss.NewStyle().Foreground(textDim),
			"abandoned": lipgloss.NewStyle().Foreground(danger),
		},
		StatusIcons: map[string]string{
			"ideating":  "○", // open circle: idea, not concrete yet
			"planned":   "●", // filled circle: solid plan
			"quoted":    "◆", // diamond: value/cost attached
			"underway":  "▸", // right triangle: in motion
			"delayed":   "‖", // double bar: paused
			"completed": "✓", // check: done
			"abandoned": "✗", // cross: cancelled
		},
	}
}
