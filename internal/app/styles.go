package app

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Header         lipgloss.Style
	HeaderBox      lipgloss.Style
	HeaderTitle    lipgloss.Style
	HeaderHint     lipgloss.Style
	HeaderBadge    lipgloss.Style
	HeaderChip     lipgloss.Style
	HeaderSection  lipgloss.Style
	HeaderLabel    lipgloss.Style
	HeaderValue    lipgloss.Style
	Keycap         lipgloss.Style
	TabActive      lipgloss.Style
	TabInactive    lipgloss.Style
	TabUnderline   lipgloss.Style
	TableHeader    lipgloss.Style
	TableSelected  lipgloss.Style
	TableSeparator lipgloss.Style
	LogTitle       lipgloss.Style
	LogFocus       lipgloss.Style
	LogBlur        lipgloss.Style
	LogValid       lipgloss.Style
	LogInvalid     lipgloss.Style
	LogLevelInfo   lipgloss.Style
	LogLevelError  lipgloss.Style
	LogLevelDebug  lipgloss.Style
	LogHighlight   lipgloss.Style
	FormClean      lipgloss.Style
	FormDirty      lipgloss.Style
	SearchBox      lipgloss.Style
	SearchTitle    lipgloss.Style
	SearchHint     lipgloss.Style
	SearchResult   lipgloss.Style
	SearchSelected lipgloss.Style
	Money          lipgloss.Style
	Readonly       lipgloss.Style
	Empty          lipgloss.Style
	Error          lipgloss.Style
	Info           lipgloss.Style
	Deleted        lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Header: lipgloss.NewStyle().Bold(true),
		HeaderBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#2F3A59")).
			Padding(0, 1),
		HeaderTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8FAFC")).
			Background(lipgloss.Color("#5A56E0")).
			Padding(0, 1).
			Bold(true),
		HeaderHint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Italic(true),
		HeaderBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0")).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1),
		HeaderChip: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#334155")).
			Background(lipgloss.Color("#111827")).
			Padding(0, 1),
		HeaderSection: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8FAFC")).
			Background(lipgloss.Color("#334155")).
			Padding(0, 1).
			Bold(true),
		HeaderLabel: lipgloss.NewStyle().Foreground(lipgloss.Color("#7C8DB5")),
		HeaderValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E8A2A2")).
			Italic(true),
		Keycap: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F172A")).
			Background(lipgloss.Color("#E2E8F0")).
			Padding(0, 1).
			Bold(true),
		TabActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#5A56E0")).
			Padding(0, 1).
			Bold(true),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A1A1AA")).
			Padding(0, 1),
		TabUnderline: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5A56E0")),
		TableHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C8DB5")).
			Bold(true),
		TableSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1F2937")).
			Background(lipgloss.Color("#E2E8F0")).
			Bold(true),
		TableSeparator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151")),
		LogTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8FAFC")).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1).
			Bold(true),
		LogFocus: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#34D399")).
			Bold(true),
		LogBlur: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")),
		LogValid: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#34D399")).
			Bold(true),
		LogInvalid: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171")).
			Bold(true),
		LogLevelInfo: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#38BDF8")).
			Bold(true),
		LogLevelError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171")).
			Bold(true),
		LogLevelDebug: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			Bold(true),
		LogHighlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1F2937")).
			Background(lipgloss.Color("#FBBF24")).
			Bold(true),
		FormClean: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")),
		FormDirty: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24")).
			Bold(true),
		SearchBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(0, 1),
		SearchTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8FAFC")).
			Background(lipgloss.Color("#1E293B")).
			Padding(0, 1).
			Bold(true),
		SearchHint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")),
		SearchResult: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0")),
		SearchSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1F2937")).
			Background(lipgloss.Color("#E2E8F0")).
			Bold(true),
		Money: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C7A4FF")),
		Readonly: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E8A2A2")).
			Italic(true),
		Empty: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Italic(true),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171")).
			Bold(true),
		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#34D399")).
			Bold(true),
		Deleted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FCA5A5")).
			Italic(true),
	}
}
