package app

import (
	"errors"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/micasa/micasa/internal/data"
	"gorm.io/gorm"
)

type Model struct {
	store                 *data.Store
	styles                Styles
	tabs                  []Tab
	active                int
	width                 int
	height                int
	showHouse             bool
	hasHouse              bool
	house                 data.HouseProfile
	mode                  Mode
	formKind              FormKind
	form                  *huh.Form
	formData              any
	status                statusMsg
	projectTypes          []data.ProjectType
	maintenanceCategories []data.MaintenanceCategory
}

func NewModel(store *data.Store) (*Model, error) {
	styles := DefaultStyles()
	model := &Model{
		store:     store,
		styles:    styles,
		tabs:      NewTabs(styles),
		active:    0,
		showHouse: false,
		mode:      modeTable,
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
	return m.formInitCmd()
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

	if m.mode == modeForm && m.form != nil {
		updated, cmd := m.form.Update(msg)
		form, ok := updated.(*huh.Form)
		if ok {
			m.form = form
		}
		switch m.form.State {
		case huh.StateCompleted:
			err := m.handleFormSubmit()
			m.exitForm()
			if err != nil {
				m.setStatusError(err.Error())
			} else {
				m.setStatusInfo("Saved.")
				_ = m.loadLookups()
				_ = m.loadHouse()
				_ = m.reloadAllTabs()
			}
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
		case "a":
			m.startAddForm()
			return m, m.formInitCmd()
		case "d":
			m.deleteSelected()
			return m, nil
		case "u":
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
	}
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
	}
	if err != nil {
		m.setStatusError(err.Error())
		return
	}
	m.setStatusInfo("Deleted.")
	_ = m.reloadActiveTab()
}

func (m *Model) restoreSelected() {
	tab := m.activeTab()
	if tab == nil {
		return
	}
	meta, ok := m.selectedRowMeta()
	if !ok {
		m.setStatusError("Nothing selected.")
		return
	}
	if !meta.Deleted {
		m.setStatusError("Selected item is not deleted.")
		return
	}
	var err error
	switch tab.Kind {
	case tabProjects:
		err = m.store.RestoreProject(meta.ID)
	case tabQuotes:
		err = m.store.RestoreQuote(meta.ID)
	case tabMaintenance:
		err = m.store.RestoreMaintenance(meta.ID)
	}
	if err != nil {
		m.setStatusError(err.Error())
		return
	}
	m.setStatusInfo("Restored.")
	_ = m.reloadActiveTab()
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
		rows, meta = projectRows(projects, m.styles)
	case tabQuotes:
		var quotes []data.Quote
		quotes, err = m.store.ListQuotes(tab.ShowDeleted)
		if err != nil {
			return err
		}
		rows, meta = quoteRows(quotes, m.styles)
	case tabMaintenance:
		var items []data.MaintenanceItem
		items, err = m.store.ListMaintenance(tab.ShowDeleted)
		if err != nil {
			return err
		}
		rows, meta = maintenanceRows(items, m.styles)
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
	height := m.height - m.houseLines() - 1 - m.statusLines()
	if height < 3 {
		height = 3
	}
	for i := range m.tabs {
		m.tabs[i].Table.SetHeight(height)
	}
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

func (m *Model) exitForm() {
	m.mode = modeTable
	m.formKind = formNone
	m.form = nil
	m.formData = nil
}

func (m *Model) setStatusInfo(text string) {
	m.status = statusMsg{Text: text, Kind: statusInfo}
}

func (m *Model) setStatusError(text string) {
	m.status = statusMsg{Text: text, Kind: statusError}
}

func (m *Model) formInitCmd() tea.Cmd {
	if m.mode == modeForm && m.form != nil {
		return m.form.Init()
	}
	return nil
}
