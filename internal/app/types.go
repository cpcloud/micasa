package app

import "github.com/charmbracelet/bubbles/table"

type Mode int

const (
	modeTable Mode = iota
	modeForm
)

type FormKind int

const (
	formNone FormKind = iota
	formHouse
	formProject
	formQuote
	formMaintenance
)

type TabKind int

const (
	tabProjects TabKind = iota
	tabQuotes
	tabMaintenance
)

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
)

type cell struct {
	Value string
	Kind  cellKind
}

type columnSpec struct {
	Title string
	Min   int
	Max   int
	Flex  bool
	Align alignKind
	Kind  cellKind
}
