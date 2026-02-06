package app

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
)

type Mode int

const (
	modeTable Mode = iota
	modeForm
	modeSearch
)

type FormKind int

const (
	formNone FormKind = iota
	formHouse
	formProject
	formQuote
	formMaintenance
	formAppliance
)

type TabKind int

const (
	tabProjects TabKind = iota
	tabQuotes
	tabMaintenance
	tabAppliances
)

func (k TabKind) String() string {
	switch k {
	case tabProjects:
		return "Projects"
	case tabQuotes:
		return "Quotes"
	case tabMaintenance:
		return "Maintenance"
	case tabAppliances:
		return "Appliances"
	default:
		return "Unknown"
	}
}

type rowMeta struct {
	ID      uint
	Deleted bool
}

type Tab struct {
	Kind        TabKind
	Name        string
	Table       table.Model
	Rows        []rowMeta
	Specs       []columnSpec
	CellRows    [][]cell
	ColCursor   int
	LastDeleted *uint
	ShowDeleted bool
}

type statusKind int

const (
	statusInfo statusKind = iota
	statusError
)

type statusMsg struct {
	Text string
	Kind statusKind
}

type searchEntry struct {
	Tab        TabKind
	ID         uint
	Title      string
	Summary    string
	Searchable string
}

type searchState struct {
	active    bool
	indexing  bool
	dirty     bool
	input     textinput.Model
	spinner   spinner.Model
	entries   []searchEntry
	results   []searchEntry
	cursor    int
	lastQuery string
}

type searchIndexMsg struct {
	Entries []searchEntry
}

type searchIndexErrMsg struct {
	Err error
}

type Options struct {
	DBPath string
}

type alignKind int

const (
	alignLeft alignKind = iota
	alignRight
)

type cellKind int

const (
	cellText cellKind = iota
	cellMoney
	cellReadonly
	cellDate
	cellStatus
)

type cell struct {
	Value  string
	Kind   cellKind
	LinkID uint // FK target ID for cross-tab navigation; 0 = no link
}

// columnLink describes a foreign-key relationship to another tab.
type columnLink struct {
	TargetTab TabKind
	Relation  string // "m:1", "1:m", "1:1", "m:n"
}

type columnSpec struct {
	Title string
	Min   int
	Max   int
	Flex  bool
	Align alignKind
	Kind  cellKind
	Link  *columnLink // non-nil if this column references another tab
}
