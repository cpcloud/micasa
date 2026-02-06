package app

import (
	"errors"
	"fmt"

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
	detail                *detailContext // non-nil when viewing a detail sub-table
	width                 int
	height                int
	showHelp              bool
	showHouse             bool
	hasHouse              bool
	house                 data.HouseProfile
	mode                  Mode
	prevMode              Mode // mode to restore after form closes
	formKind              FormKind
	form                  *huh.Form
	formData              any
	formSnapshot          string
	formDirty             bool
	editID                *uint
	undoStack             []undoEntry
	redoStack             []undoEntry
	status                statusMsg
	projectTypes          []data.ProjectType
	maintenanceCategories []data.MaintenanceCategory
	vendors               []data.Vendor
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
		mode:      modeNormal,
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

	if m.mode == modeForm && m.form != nil {
		return m.updateForm(msg)
	}

	switch typed := msg.(type) {
	case tea.KeyMsg:
		if cmd, handled := m.handleCommonKeys(typed); handled {
			return m, cmd
		}
		if m.mode == modeNormal {
			if cmd, handled := m.handleNormalKeys(typed); handled {
				return m, cmd
			}
		}
		if m.mode == modeEdit {
			if cmd, handled := m.handleEditKeys(typed); handled {
				return m, cmd
			}
		}
	}

	// Pass unhandled messages to the effective table (handles j/k, g/G, etc.).
	tab := m.effectiveTab()
	if tab == nil {
		return m, nil
	}
	var cmd tea.Cmd
	tab.Table, cmd = tab.Table.Update(msg)
	return m, cmd
}

// updateForm handles input while a form is active.
func (m *Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "ctrl+s" {
		return m, m.saveForm()
	}
	if _, isResize := msg.(tea.WindowSizeMsg); isResize && m.formKind == formHouse {
		return m, nil
	}
	// Intercept 1-9 on Select fields to jump to the Nth option.
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if n, isOrdinal := selectOrdinal(keyMsg); isOrdinal && isSelectField(m.form) {
			m.jumpSelectToOrdinal(n)
			return m, nil
		}
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

// handleCommonKeys processes keys available in both Normal and Edit modes.
func (m *Model) handleCommonKeys(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.String() {
	case "?":
		m.showHelp = true
		return nil, true
	case "H":
		m.showHouse = !m.showHouse
		m.resizeTables()
		return nil, true
	case "h", "left":
		if tab := m.effectiveTab(); tab != nil {
			tab.ColCursor--
			if tab.ColCursor < 0 {
				tab.ColCursor = len(tab.Specs) - 1
			}
		}
		return nil, true
	case "l", "right":
		if tab := m.effectiveTab(); tab != nil {
			tab.ColCursor++
			if tab.ColCursor >= len(tab.Specs) {
				tab.ColCursor = 0
			}
		}
		return nil, true
	}
	return nil, false
}

// handleNormalKeys processes keys unique to Normal mode.
func (m *Model) handleNormalKeys(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.String() {
	case "tab":
		if m.detail == nil {
			m.nextTab()
		}
		return nil, true
	case "shift+tab":
		if m.detail == nil {
			m.prevTab()
		}
		return nil, true
	case "s":
		if tab := m.effectiveTab(); tab != nil {
			toggleSort(tab, tab.ColCursor)
			applySorts(tab)
		}
		return nil, true
	case "S":
		if tab := m.effectiveTab(); tab != nil {
			clearSorts(tab)
			applySorts(tab)
		}
		return nil, true
	case "i":
		m.enterEditMode()
		return nil, true
	case "enter":
		if err := m.handleNormalEnter(); err != nil {
			m.setStatusError(err.Error())
			return nil, true
		}
		if m.mode == modeForm {
			return m.formInitCmd(), true
		}
		return nil, true
	case "q":
		return tea.Quit, true
	case "esc":
		if m.detail != nil {
			m.closeDetail()
			return nil, true
		}
		m.status = statusMsg{}
		return nil, true
	}
	return nil, false
}

// handleNormalEnter handles enter in Normal mode: drill into detail views
// on maintenance rows, follow FK links, or fall through to cell edit.
func (m *Model) handleNormalEnter() error {
	tab := m.effectiveTab()
	if tab == nil {
		return fmt.Errorf("no active tab")
	}
	meta, ok := m.selectedRowMeta()
	if !ok {
		return fmt.Errorf("nothing selected")
	}

	// On the maintenance tab (not already in detail), open service log.
	if m.detail == nil && tab.Kind == tabMaintenance {
		item, err := m.store.GetMaintenance(meta.ID)
		if err != nil {
			return fmt.Errorf("load maintenance item: %w", err)
		}
		return m.openDetail(meta.ID, item.Name)
	}

	// Otherwise: follow FK link or fall through to edit.
	return m.startCellOrFormEdit()
}

// handleEditKeys processes keys unique to Edit mode.
func (m *Model) handleEditKeys(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.String() {
	case "a":
		m.startAddForm()
		return m.formInitCmd(), true
	case "e", "enter":
		if err := m.startCellOrFormEdit(); err != nil {
			m.setStatusError(err.Error())
			return nil, true
		}
		return m.formInitCmd(), true
	case "d":
		m.toggleDeleteSelected()
		return nil, true
	case "u":
		if err := m.popUndo(); err != nil {
			m.setStatusError(err.Error())
		} else if m.store != nil {
			_ = m.loadLookups()
			_ = m.loadHouse()
			_ = m.reloadAllTabs()
			if m.detail != nil {
				_ = m.reloadDetailTab()
			}
		}
		return nil, true
	case "r":
		if err := m.popRedo(); err != nil {
			m.setStatusError(err.Error())
		} else if m.store != nil {
			_ = m.loadLookups()
			_ = m.loadHouse()
			_ = m.reloadAllTabs()
			if m.detail != nil {
				_ = m.reloadDetailTab()
			}
		}
		return nil, true
	case "x":
		m.toggleShowDeleted()
		return nil, true
	case "p":
		m.startHouseForm()
		return m.formInitCmd(), true
	case "esc":
		m.enterNormalMode()
		return nil, true
	}
	return nil, false
}

func (m *Model) View() string {
	return m.buildView()
}

func (m *Model) enterNormalMode() {
	m.mode = modeNormal
	m.setAllTableKeyMaps(normalTableKeyMap())
}

func (m *Model) enterEditMode() {
	m.mode = modeEdit
	m.setAllTableKeyMaps(editTableKeyMap())
}

func (m *Model) activeTab() *Tab {
	if m.active < 0 || m.active >= len(m.tabs) {
		return nil
	}
	return &m.tabs[m.active]
}

// effectiveTab returns the detail tab when a detail view is open, otherwise
// the main active tab. All interaction code should use this.
func (m *Model) effectiveTab() *Tab {
	if m.detail != nil {
		return &m.detail.Tab
	}
	return m.activeTab()
}

func (m *Model) openDetail(maintID uint, maintName string) error {
	specs := serviceLogColumnSpecs()
	tab := Tab{
		Kind:    tabMaintenance, // inherits parent kind for styling
		Name:    "Service Log",
		Handler: serviceLogHandler{maintenanceItemID: maintID},
		Specs:   specs,
		Table:   newTable(specsToColumns(specs), m.styles),
	}
	m.detail = &detailContext{
		ParentTabIndex: m.active,
		ParentRowID:    maintID,
		Breadcrumb:     "Maintenance > " + maintName,
		Tab:            tab,
	}
	if err := m.reloadDetailTab(); err != nil {
		m.detail = nil
		return err
	}
	m.resizeTables()
	m.status = statusMsg{}
	return nil
}

func (m *Model) closeDetail() {
	if m.detail == nil {
		return
	}
	parentIdx := m.detail.ParentTabIndex
	parentRowID := m.detail.ParentRowID
	m.detail = nil
	m.active = parentIdx
	if m.store != nil {
		_ = m.reloadActiveTab()
	}
	// Restore cursor to the parent row.
	if tab := m.activeTab(); tab != nil {
		selectRowByID(tab, parentRowID)
	}
	m.resizeTables()
	m.status = statusMsg{}
}

func (m *Model) reloadDetailTab() error {
	if m.detail == nil || m.store == nil {
		return nil
	}
	tab := &m.detail.Tab
	rows, meta, cellRows, err := tab.Handler.Load(m.store, tab.ShowDeleted)
	if err != nil {
		return err
	}
	tab.CellRows = cellRows
	tab.Table.SetRows(rows)
	tab.Rows = meta
	applySorts(tab)
	return nil
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
	tab := m.effectiveTab()
	if tab == nil {
		return
	}
	if err := tab.Handler.StartAddForm(m); err != nil {
		m.setStatusError(err.Error())
	}
}

func (m *Model) startEditForm() error {
	tab := m.effectiveTab()
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
	return tab.Handler.StartEditForm(m, meta.ID)
}

func (m *Model) startCellOrFormEdit() error {
	tab := m.effectiveTab()
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
	return tab.Handler.InlineEdit(m, meta.ID, col)
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
	tab := m.effectiveTab()
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

func (m *Model) reloadEffectiveTab() error {
	if m.detail != nil {
		return m.reloadDetailTab()
	}
	return m.reloadActiveTab()
}

func (m *Model) toggleDeleteSelected() {
	tab := m.effectiveTab()
	if tab == nil {
		return
	}
	meta, ok := m.selectedRowMeta()
	if !ok {
		m.setStatusError("Nothing selected.")
		return
	}
	if meta.Deleted {
		if err := tab.Handler.Restore(m.store, meta.ID); err != nil {
			m.setStatusError(err.Error())
			return
		}
		if tab.LastDeleted != nil && *tab.LastDeleted == meta.ID {
			tab.LastDeleted = nil
		}
		m.setStatusInfo("Restored.")
		_ = m.reloadEffectiveTab()
		return
	}
	if err := tab.Handler.Delete(m.store, meta.ID); err != nil {
		m.setStatusError(err.Error())
		return
	}
	tab.LastDeleted = &meta.ID
	m.setStatusInfo("Deleted. Press d to restore.")
	_ = m.reloadEffectiveTab()
}

func (m *Model) toggleShowDeleted() {
	tab := m.effectiveTab()
	if tab == nil {
		return
	}
	tab.ShowDeleted = !tab.ShowDeleted
	_ = m.reloadEffectiveTab()
}

func (m *Model) selectedRowMeta() (rowMeta, bool) {
	tab := m.effectiveTab()
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
	rows, meta, cellRows, err := tab.Handler.Load(m.store, tab.ShowDeleted)
	if err != nil {
		return err
	}
	tab.CellRows = cellRows
	tab.Table.SetRows(rows)
	tab.Rows = meta
	applySorts(tab)
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
	m.vendors, err = m.store.ListVendors()
	if err != nil {
		return err
	}
	m.syncFixedValues()
	return nil
}

// syncFixedValues updates FixedValues on columns that reference dynamic lookup
// tables so columnWidths stays stable regardless of which values are displayed.
func (m *Model) syncFixedValues() {
	for i := range m.tabs {
		tab := &m.tabs[i]
		tab.Handler.SyncFixedValues(m, tab.Specs)
	}
}

func setFixedValues(specs []columnSpec, title string, values []string) {
	for i := range specs {
		if specs[i].Title == title {
			specs[i].FixedValues = values
			return
		}
	}
}

func (m *Model) resizeTables() {
	// Chrome: 1 blank after house + 1 tab/breadcrumb row + 1 underline = 3
	height := m.height - m.houseLines() - 3 - m.statusLines()
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
	if m.detail != nil {
		m.detail.Tab.Table.SetHeight(tableHeight)
		m.detail.Tab.Table.SetWidth(m.width)
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

func (m *Model) saveForm() tea.Cmd {
	m.snapshotForUndo()
	err := m.handleFormSubmit()
	if err != nil {
		m.setStatusError(err.Error())
		return nil
	}
	m.exitForm()
	m.setStatusInfo("Saved.")
	_ = m.loadLookups()
	_ = m.loadHouse()
	_ = m.reloadAllTabs()
	if m.detail != nil {
		_ = m.reloadDetailTab()
	}
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
	m.mode = m.prevMode
	// Restore correct table key bindings for the returning mode.
	if m.mode == modeEdit {
		m.setAllTableKeyMaps(editTableKeyMap())
	} else {
		m.setAllTableKeyMaps(normalTableKeyMap())
	}
	m.formKind = formNone
	m.form = nil
	m.formData = nil
	m.formSnapshot = ""
	m.formDirty = false
	m.editID = nil
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

func selectRowByID(tab *Tab, id uint) bool {
	for idx, meta := range tab.Rows {
		if meta.ID == id {
			tab.Table.SetCursor(idx)
			return true
		}
	}
	return false
}

func tabIndex(kind TabKind) int {
	switch kind {
	case tabProjects:
		return 0
	case tabQuotes:
		return 1
	case tabMaintenance:
		return 2
	case tabAppliances:
		return 3
	default:
		return 0
	}
}
