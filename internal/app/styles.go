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
	TabUnderline    lipgloss.Style
	TableHeader     lipgloss.Style
	TableSelected   lipgloss.Style
	TableSeparator  lipgloss.Style
	CellActive      lipgloss.Style
	ColActiveHeader lipgloss.Style
	LogTitle        lipgloss.Style
	LogFocus        lipgloss.Style
	LogBlur         lipgloss.Style
	LogValid        lipgloss.Style
	LogInvalid      lipgloss.Style
	LogLevelInfo    lipgloss.Style
	LogLevelError   lipgloss.Style
	LogLevelDebug   lipgloss.Style
	LogHighlight    lipgloss.Style
	FormClean       lipgloss.Style
	FormDirty       lipgloss.Style
	SearchBox       lipgloss.Style
	SearchTitle     lipgloss.Style
	SearchHint      lipgloss.Style
	SearchResult    lipgloss.Style
	SearchSelected  lipgloss.Style
	Money           lipgloss.Style
	Readonly        lipgloss.Style
	Empty           lipgloss.Style
	Error           lipgloss.Style
	Info            lipgloss.Style
	Deleted         lipgloss.Style
	DeletedLabel    lipgloss.Style
	DBHint          lipgloss.Style
	LinkIndicator   lipgloss.Style
	StatusStyles    map[string]lipgloss.Style
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
		TabUnderline: lipgloss.NewStyle().
			Foreground(accent),
		TableHeader: lipgloss.NewStyle().
			Foreground(textDim).
			Bold(true),
		TableSelected: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(textBright).
			Bold(true),
		TableSeparator: lipgloss.NewStyle().
			Foreground(border),
		CellActive: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(secondary).
			Bold(true),
		ColActiveHeader: lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true),
		LogTitle: lipgloss.NewStyle().
			Foreground(textBright).
			Background(surface).
			Padding(0, 1).
			Bold(true),
		LogFocus: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),
		LogBlur: lipgloss.NewStyle().
			Foreground(textMid),
		LogValid: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),
		LogInvalid: lipgloss.NewStyle().
			Foreground(danger).
			Bold(true),
		LogLevelInfo: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		LogLevelError: lipgloss.NewStyle().
			Foreground(danger).
			Bold(true),
		LogLevelDebug: lipgloss.NewStyle().
			Foreground(muted).
			Bold(true),
		LogHighlight: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(warning).
			Bold(true),
		FormClean: lipgloss.NewStyle().
			Foreground(textMid),
		FormDirty: lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true),
		SearchBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(0, 1),
		SearchTitle: lipgloss.NewStyle().
			Foreground(textBright).
			Background(surface).
			Padding(0, 1).
			Bold(true),
		SearchHint: lipgloss.NewStyle().
			Foreground(textMid),
		SearchResult: lipgloss.NewStyle().
			Foreground(textBright),
		SearchSelected: lipgloss.NewStyle().
			Foreground(onAccent).
			Background(textBright).
			Bold(true),
		Money: lipgloss.NewStyle().
			Foreground(success),
		Readonly: lipgloss.NewStyle().
			Foreground(textDim),
		Empty: lipgloss.NewStyle().
			Foreground(textDim),
		Error: lipgloss.NewStyle().
			Foreground(danger).
			Bold(true),
		Info: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),
		Deleted: lipgloss.NewStyle().
			Foreground(danger).
			Strikethrough(true),
		DeletedLabel: lipgloss.NewStyle().
			Foreground(danger),
		DBHint: lipgloss.NewStyle().
			Foreground(textBright),
		LinkIndicator: lipgloss.NewStyle().
			Foreground(muted),
		StatusStyles: map[string]lipgloss.Style{
			"ideating":    lipgloss.NewStyle().Foreground(muted),
			"planned":     lipgloss.NewStyle().Foreground(accent),
			"quoted":      lipgloss.NewStyle().Foreground(secondary),
			"in_progress": lipgloss.NewStyle().Foreground(success),
			"delayed":     lipgloss.NewStyle().Foreground(warning),
			"completed":   lipgloss.NewStyle().Foreground(textDim),
			"abandoned":   lipgloss.NewStyle().Foreground(danger),
		},
	}
}
