package app

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/micasa/micasa/internal/data"
	"gorm.io/gorm"
)

type Model struct {
	store                 *data.Store
	dbPath                string
	styles                Styles
	tabs                  []Tab
	active                int
	width                 int
	height                int
	showHelp              bool
	showHouse             bool
	hasHouse              bool
	house                 data.HouseProfile
	mode                  Mode
	formKind              FormKind
	form                  *huh.Form
	formData              any
	formSnapshot          string
	formDirty             bool
	editID                *uint
	status                statusMsg
	log                   logState
	search                searchState
	projectTypes          []data.ProjectType
	maintenanceCategories []data.MaintenanceCategory
}

func NewModel(store *data.Store, options Options) (*Model, error) {
	styles := DefaultStyles()
	model := &Model{
		store:     store,
		dbPath:    options.DBPath,
		styles:    styles,
		tabs:      NewTabs(styles),
		active:    0,
		showHouse: false,
		mode:      modeTable,
		log:       newLogState(),
		search:    newSearchState(),
	}
	if err := model.loadLookups(); err != nil {
		return nil, err
	}
	if err := model.loadHouse(); err != nil {
		return nil, err
	}
	if err := model.reloadAllTabs(); err != nil {
		return nil, err
	}
	if !model.hasHouse {
		model.startHouseForm()
	}
	return model, nil
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.formInitCmd(), m.startSearchIndexBuild())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		m.resizeTables()
	case tea.KeyMsg:
		if typed.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Help overlay: esc or ? dismisses it.
	if m.showHelp {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc", "?":
				m.showHelp = false
			}
		}
		return m, nil
	}

	// Log filter typing mode: only esc escapes back to log browsing.
	if m.log.enabled && m.log.focus && m.mode != modeForm {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				m.log.focus = false
				m.log.input.Blur()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.log.input, cmd = m.log.input.Update(msg)
		m.log.setFilter(m.log.input.Value())
		return m, cmd
	}

	// Log browsing mode: !, /, esc handled here.
	if m.log.enabled && m.mode != modeForm {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc":
				return m, m.toggleLog()
			case "!":
				m.log.cycleLevel()
				m.logInfo(fmt.Sprintf("Log level: %s", m.log.levelLabel()))
				return m, nil
			case "/":
				m.log.focus = true
				return m, m.log.input.Focus()
			}
		}
		return m, nil
	}

	if m.mode == modeSearch {
		return m.updateSearch(msg)
	}

	if m.mode == modeForm && m.form != nil {
		// ctrl+s saves the form immediately from any field.
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "ctrl+s" {
			return m, m.saveForm()
		}
		// Don't let WindowSizeMsg resize the house form (it has a fixed width).
		if _, isResize := msg.(tea.WindowSizeMsg); isResize && m.formKind == formHouse {
			return m, nil
		}
		updated, cmd := m.form.Update(msg)
		form, ok := updated.(*huh.Form)
		if ok {
			m.form = form
		}
		m.checkFormDirty()
		switch m.form.State {
		case huh.StateCompleted:
			return m, m.saveForm()
		case huh.StateAborted:
			if m.formKind == formHouse && !m.hasHouse {
				m.setStatusError("House profile required.")
				m.startHouseForm()
				return m, m.formInitCmd()
			}
			m.exitForm()
		}
		return m, cmd
	}

	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch typed.String() {
		case "q":
			return m, tea.Quit
		case "?":
			m.showHelp = true
			return m, nil
		case "/":
			return m, m.openSearch()
		case "l":
			return m, m.toggleLog()
		case "tab":
			m.nextTab()
			return m, nil
		case "shift+tab":
			m.prevTab()
			return m, nil
		case "h":
			m.showHouse = !m.showHouse
			m.resizeTables()
			return m, nil
		case "p":
			m.startHouseForm()
			return m, m.formInitCmd()
		case "a":
			m.startAddForm()
			return m, m.formInitCmd()
		case "left":
			if tab := m.activeTab(); tab != nil {
				tab.ColCursor--
				if tab.ColCursor < 0 {
					tab.ColCursor = len(tab.Specs) - 1
				}
			}
			return m, nil
		case "right":
			if tab := m.activeTab(); tab != nil {
				tab.ColCursor++
				if tab.ColCursor >= len(tab.Specs) {
					tab.ColCursor = 0
				}
			}
			return m, nil
		case "e", "enter":
			if err := m.startCellOrFormEdit(); err != nil {
				m.setStatusError(err.Error())
				return m, nil
			}
			return m, m.formInitCmd()
		case "d":
			m.deleteSelected()
			return m, nil
		case "u", "U":
			m.restoreSelected()
			return m, nil
		case "x":
			m.toggleShowDeleted()
			return m, nil
		case "esc":
			m.status = statusMsg{}
			return m, nil
		}
	}

	tab := m.activeTab()
	if tab == nil {
		return m, nil
	}
	var cmd tea.Cmd
	tab.Table, cmd = tab.Table.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	return m.buildView()
}

func (m *Model) activeTab() *Tab {
	if m.active < 0 || m.active >= len(m.tabs) {
		return nil
	}
	return &m.tabs[m.active]
}

func (m *Model) nextTab() {
	if len(m.tabs) == 0 {
		return
	}
	m.active = (m.active + 1) % len(m.tabs)
	m.status = statusMsg{}
	_ = m.reloadActiveTab()
}

func (m *Model) prevTab() {
	if len(m.tabs) == 0 {
		return
	}
	m.active--
	if m.active < 0 {
		m.active = len(m.tabs) - 1
	}
	m.status = statusMsg{}
	_ = m.reloadActiveTab()
}

func (m *Model) startAddForm() {
	tab := m.activeTab()
	if tab == nil {
		return
	}
	switch tab.Kind {
	case tabProjects:
		m.startProjectForm()
	case tabQuotes:
		if err := m.startQuoteForm(); err != nil {
			m.setStatusError(err.Error())
		}
	case tabMaintenance:
		m.startMaintenanceForm()
	case tabAppliances:
		m.startApplianceForm()
	}
}

func (m *Model) startEditForm() error {
	tab := m.activeTab()
	if tab == nil {
		return fmt.Errorf("no active tab")
	}
	meta, ok := m.selectedRowMeta()
	if !ok {
		return fmt.Errorf("nothing selected")
	}
	if meta.Deleted {
		return fmt.Errorf("cannot edit a deleted item")
	}
	switch tab.Kind {
	case tabProjects:
		return m.startEditProjectForm(meta.ID)
	case tabQuotes:
		return m.startEditQuoteForm(meta.ID)
	case tabMaintenance:
		return m.startEditMaintenanceForm(meta.ID)
	case tabAppliances:
		return m.startEditApplianceForm(meta.ID)
	default:
		return fmt.Errorf("unknown tab")
	}
}

func (m *Model) startCellOrFormEdit() error {
	tab := m.activeTab()
	if tab == nil {
		return fmt.Errorf("no active tab")
	}
	meta, ok := m.selectedRowMeta()
	if !ok {
		return fmt.Errorf("nothing selected")
	}
	if meta.Deleted {
		return fmt.Errorf("cannot edit a deleted item")
	}
	col := tab.ColCursor
	if col < 0 || col >= len(tab.Specs) {
		col = 0
	}
	spec := tab.Specs[col]

	// If the column is linked and the cell has a target ID, navigate cross-tab.
	if spec.Link != nil {
		if c, ok := m.selectedCell(col); ok && c.LinkID > 0 {
			return m.navigateToLink(spec.Link, c.LinkID)
		}
	}

	if spec.Kind == cellReadonly {
		return m.startEditForm()
	}
	return m.startInlineCellEdit(meta.ID, tab.Kind, col)
}

// navigateToLink switches to the target tab and selects the row matching the FK.
func (m *Model) navigateToLink(link *columnLink, targetID uint) error {
	targetIdx := tabIndex(link.TargetTab)
	m.active = targetIdx
	_ = m.reloadActiveTab()
	tab := m.activeTab()
	if tab == nil {
		return fmt.Errorf("target tab not found")
	}
	if selectRowByID(tab, targetID) {
		m.setStatusInfo(fmt.Sprintf("Followed %s link to ID %d.", link.Relation, targetID))
		return nil
	}
	m.setStatusError(fmt.Sprintf("Linked item %d not found (deleted?).", targetID))
	return nil
}

// selectedCell returns the cell at the given column for the currently selected row.
func (m *Model) selectedCell(col int) (cell, bool) {
	tab := m.activeTab()
	if tab == nil {
		return cell{}, false
	}
	cursor := tab.Table.Cursor()
	if cursor < 0 || cursor >= len(tab.CellRows) {
		return cell{}, false
	}
	row := tab.CellRows[cursor]
	if col < 0 || col >= len(row) {
		return cell{}, false
	}
	return row[col], true
}

func (m *Model) deleteSelected() {
	tab := m.activeTab()
	if tab == nil {
		return
	}
	meta, ok := m.selectedRowMeta()
	if !ok {
		m.setStatusError("Nothing selected.")
		return
	}
	if meta.Deleted {
		m.setStatusError("Already deleted.")
		return
	}
	var err error
	switch tab.Kind {
	case tabProjects:
		err = m.store.DeleteProject(meta.ID)
	case tabQuotes:
		err = m.store.DeleteQuote(meta.ID)
	case tabMaintenance:
		err = m.store.DeleteMaintenance(meta.ID)
	case tabAppliances:
		err = m.store.DeleteAppliance(meta.ID)
	}
	if err != nil {
		m.setStatusError(err.Error())
		return
	}
	tab.LastDeleted = &meta.ID
	m.search.dirty = true
	m.logDebug(fmt.Sprintf("Deleted %s %d", tab.Name, meta.ID))
	m.setStatusInfo("Deleted. Press u to undo.")
	_ = m.reloadActiveTab()
}

func (m *Model) restoreSelected() {
	tab := m.activeTab()
	if tab == nil {
		return
	}
	meta, ok := m.selectedRowMeta()
	if !ok {
		if tab.LastDeleted != nil {
			if err := m.restoreByTab(tab.Kind, *tab.LastDeleted); err != nil {
				m.setStatusError(err.Error())
				return
			}
			tab.LastDeleted = nil
			m.setStatusInfo("Restored last deleted.")
			_ = m.reloadActiveTab()
			return
		}
		entity := deletionEntityForTab(tab.Kind)
		record, err := m.store.LastDeletion(entity)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			m.setStatusError("Nothing to undo.")
			return
		}
		if err != nil {
			m.setStatusError(err.Error())
			return
		}
		if err := m.restoreByTab(tab.Kind, record.TargetID); err != nil {
			m.setStatusError(err.Error())
			return
		}
		m.setStatusInfo("Restored last deleted.")
		_ = m.reloadActiveTab()
		return
	}
	if !meta.Deleted {
		m.setStatusError("Selected item is not deleted.")
		return
	}
	if err := m.restoreByTab(tab.Kind, meta.ID); err != nil {
		m.setStatusError(err.Error())
		return
	}
	if tab.LastDeleted != nil && *tab.LastDeleted == meta.ID {
		tab.LastDeleted = nil
	}
	m.search.dirty = true
	m.logDebug(fmt.Sprintf("Restored %s %d", tab.Name, meta.ID))
	m.setStatusInfo("Restored.")
	_ = m.reloadActiveTab()
}

func (m *Model) restoreByTab(kind TabKind, id uint) error {
	switch kind {
	case tabProjects:
		return m.store.RestoreProject(id)
	case tabQuotes:
		return m.store.RestoreQuote(id)
	case tabMaintenance:
		return m.store.RestoreMaintenance(id)
	case tabAppliances:
		return m.store.RestoreAppliance(id)
	default:
		return nil
	}
}

func deletionEntityForTab(kind TabKind) string {
	switch kind {
	case tabProjects:
		return data.DeletionEntityProject
	case tabQuotes:
		return data.DeletionEntityQuote
	case tabMaintenance:
		return data.DeletionEntityMaintenance
	case tabAppliances:
		return data.DeletionEntityAppliance
	default:
		return ""
	}
}

func (m *Model) toggleShowDeleted() {
	tab := m.activeTab()
	if tab == nil {
		return
	}
	tab.ShowDeleted = !tab.ShowDeleted
	_ = m.reloadActiveTab()
}

func (m *Model) selectedRowMeta() (rowMeta, bool) {
	tab := m.activeTab()
	if tab == nil || len(tab.Rows) == 0 {
		return rowMeta{}, false
	}
	cursor := tab.Table.Cursor()
	if cursor < 0 || cursor >= len(tab.Rows) {
		return rowMeta{}, false
	}
	return tab.Rows[cursor], true
}

func (m *Model) reloadActiveTab() error {
	tab := m.activeTab()
	if tab == nil {
		return nil
	}
	return m.reloadTab(tab)
}

func (m *Model) reloadAllTabs() error {
	for i := range m.tabs {
		if err := m.reloadTab(&m.tabs[i]); err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) reloadTab(tab *Tab) error {
	var rows []table.Row
	var meta []rowMeta
	var err error
	switch tab.Kind {
	case tabProjects:
		var projects []data.Project
		projects, err = m.store.ListProjects(tab.ShowDeleted)
		if err != nil {
			return err
		}
		var cellRows [][]cell
		rows, meta, cellRows = projectRows(projects)
		tab.CellRows = cellRows
		m.logDebug(fmt.Sprintf("Loaded %d projects", len(projects)))
	case tabQuotes:
		var quotes []data.Quote
		quotes, err = m.store.ListQuotes(tab.ShowDeleted)
		if err != nil {
			return err
		}
		var cellRows [][]cell
		rows, meta, cellRows = quoteRows(quotes)
		tab.CellRows = cellRows
		m.logDebug(fmt.Sprintf("Loaded %d quotes", len(quotes)))
	case tabMaintenance:
		var items []data.MaintenanceItem
		items, err = m.store.ListMaintenance(tab.ShowDeleted)
		if err != nil {
			return err
		}
		var cellRows [][]cell
		rows, meta, cellRows = maintenanceRows(items)
		tab.CellRows = cellRows
		m.logDebug(fmt.Sprintf("Loaded %d maintenance items", len(items)))
	case tabAppliances:
		var appliances []data.Appliance
		appliances, err = m.store.ListAppliances(tab.ShowDeleted)
		if err != nil {
			return err
		}
		var cellRows [][]cell
		rows, meta, cellRows = applianceRows(appliances)
		tab.CellRows = cellRows
		m.logDebug(fmt.Sprintf("Loaded %d appliances", len(appliances)))
	}
	tab.Table.SetRows(rows)
	tab.Rows = meta
	return nil
}

func (m *Model) loadHouse() error {
	profile, err := m.store.HouseProfile()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		m.hasHouse = false
		return nil
	}
	if err != nil {
		return err
	}
	m.house = profile
	m.hasHouse = true
	return nil
}

func (m *Model) loadLookups() error {
	var err error
	m.projectTypes, err = m.store.ProjectTypes()
	if err != nil {
		return err
	}
	m.maintenanceCategories, err = m.store.MaintenanceCategories()
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) resizeTables() {
	// Chrome: 1 blank after house + 1 tab row + 1 tab underline = 3
	height := m.height - m.houseLines() - 3 - m.statusLines() - m.logLines()
	if height < 4 {
		height = 4
	}
	tableHeight := height - 1
	if tableHeight < 2 {
		tableHeight = 2
	}
	for i := range m.tabs {
		m.tabs[i].Table.SetHeight(tableHeight)
		m.tabs[i].Table.SetWidth(m.width)
	}
	m.resizeLog()
}

func (m *Model) houseLines() int {
	return lipgloss.Height(m.houseView())
}

func (m *Model) statusLines() int {
	if m.status.Text == "" {
		return 1
	}
	return 2
}

func (m *Model) logLines() int {
	if !m.log.enabled {
		return 0
	}
	return m.logBodyLines() + 6
}

func (m *Model) resizeLog() {
	if !m.log.enabled {
		return
	}
	width := m.width - 24
	if width < 16 {
		width = 16
	}
	m.log.input.Width = width
}

func (m *Model) logBodyLines() int {
	body := 4
	if m.height > 0 {
		body = m.height / 4
		if body < 3 {
			body = 3
		}
		if body > 8 {
			body = 8
		}
	}
	return body
}

func (m *Model) saveForm() tea.Cmd {
	err := m.handleFormSubmit()
	if err != nil {
		m.setStatusError(err.Error())
		return nil
	}
	m.exitForm()
	m.setStatusInfo("Saved.")
	m.search.dirty = true
	_ = m.loadLookups()
	_ = m.loadHouse()
	_ = m.reloadAllTabs()
	return nil
}

func (m *Model) snapshotForm() {
	m.formSnapshot = fmt.Sprintf("%v", m.formData)
	m.formDirty = false
}

func (m *Model) checkFormDirty() {
	m.formDirty = fmt.Sprintf("%v", m.formData) != m.formSnapshot
}

func (m *Model) exitForm() {
	m.mode = modeTable
	m.formKind = formNone
	m.form = nil
	m.formData = nil
	m.formSnapshot = ""
	m.formDirty = false
	m.editID = nil
}

func (m *Model) setStatusInfo(text string) {
	m.status = statusMsg{Text: text, Kind: statusInfo}
	m.logInfo(text)
}

func (m *Model) setStatusError(text string) {
	m.status = statusMsg{Text: text, Kind: statusError}
	m.logError(text)
}

func (m *Model) formInitCmd() tea.Cmd {
	if m.mode == modeForm && m.form != nil {
		return m.form.Init()
	}
	return nil
}

func (m *Model) toggleLog() tea.Cmd {
	if m.log.enabled {
		m.log.enabled = false
		m.log.focus = false
		m.log.input.Blur()
		m.resizeTables()
		return nil
	}
	m.log.enabled = true
	m.resizeTables()
	return nil
}

const (
	defaultWidth  = 80
	defaultHeight = 24
)

func (m *Model) effectiveWidth() int {
	if m.width > 0 {
		return m.width
	}
	return defaultWidth
}

func (m *Model) effectiveHeight() int {
	if m.height > 0 {
		return m.height
	}
	return defaultHeight
}

func (m *Model) logInfo(message string) {
	m.log.append(logInfo, message)
}

func (m *Model) logError(message string) {
	m.log.append(logError, message)
}

func (m *Model) logDebug(message string) {
	m.log.append(logDebug, message)
}
