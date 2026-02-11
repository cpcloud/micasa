// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"errors"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/cpcloud/micasa/internal/data"
	"gorm.io/gorm"
)

const (
	keyEsc   = "esc"
	keyEnter = "enter"
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
	helpScroll            int
	showHouse             bool
	showDashboard         bool
	showNotePreview       bool
	notePreviewText       string
	notePreviewTitle      string
	calendar              *calendarState
	columnFinder          *columnFinderState
	dashboard             dashboardData
	dashCursor            int
	dashNav               []dashNavEntry
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
	inlineInput           *inlineInputState
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
	} else {
		model.showDashboard = true
		_ = model.loadDashboard()
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
		m.updateAllViewports()
	case tea.KeyMsg:
		if typed.String() == "ctrl+c" {
			return m, tea.Interrupt
		}
	}

	// Help overlay: scrollable, esc or ? dismisses it.
	if m.showHelp {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case keyEsc, "?":
				m.showHelp = false
				m.helpScroll = 0
			case "j", "down":
				m.helpScroll++
			case "k", "up":
				if m.helpScroll > 0 {
					m.helpScroll--
				}
			case "g":
				m.helpScroll = 0
			case "G":
				// Jump to bottom; clamped in view rendering.
				m.helpScroll = 9999
			case "d":
				m.helpScroll += 10
			case "u":
				m.helpScroll -= 10
				if m.helpScroll < 0 {
					m.helpScroll = 0
				}
			}
		}
		return m, nil
	}

	// Note preview overlay: any key dismisses it.
	if m.showNotePreview {
		if _, ok := msg.(tea.KeyMsg); ok {
			m.showNotePreview = false
			m.notePreviewText = ""
			m.notePreviewTitle = ""
		}
		return m, nil
	}

	// Calendar date picker: absorb all keys.
	if m.calendar != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m.handleCalendarKey(keyMsg)
		}
		return m, nil
	}

	// Column finder overlay: absorb all keys.
	if m.columnFinder != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m, m.handleColumnFinderKey(keyMsg)
		}
		return m, nil
	}

	// Inline text input: absorb all keys.
	if m.inlineInput != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m.handleInlineInputKey(keyMsg)
		}
		return m, nil
	}

	if m.mode == modeForm && m.form != nil {
		return m.updateForm(msg)
	}

	switch typed := msg.(type) {
	case tea.KeyMsg:
		// Dashboard intercepts nav keys before other handlers.
		if m.showDashboard {
			if cmd, handled := m.handleDashboardKeys(typed); handled {
				return m, cmd
			}
		}
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

	// Dashboard absorbs remaining keys so they don't reach the table.
	if m.showDashboard {
		return m, nil
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
	default:
	}
	return m, cmd
}

// handleDashboardKeys intercepts keys that belong to the dashboard (j/k
// navigation, enter to jump, h/l blocked). Keys like D, tab, ?, q fall
// through to the normal handlers.
func (m *Model) handleDashboardKeys(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.String() {
	case "j", "down":
		m.dashDown()
		return nil, true
	case "k", "up":
		m.dashUp()
		return nil, true
	case "g":
		m.dashTop()
		return nil, true
	case "G":
		m.dashBottom()
		return nil, true
	case keyEnter:
		m.dashJump()
		return nil, true
	case "h", "l", "left", "right":
		// Block column movement on dashboard.
		return nil, true
	case "s", "S", "c", "C", "i", "/":
		// Block table-specific keys on dashboard.
		return nil, true
	}
	return nil, false
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
			tab.ColCursor = nextVisibleCol(tab.Specs, tab.ColCursor, false)
			m.updateTabViewport(tab)
		}
		return nil, true
	case "l", "right":
		if tab := m.effectiveTab(); tab != nil {
			tab.ColCursor = nextVisibleCol(tab.Specs, tab.ColCursor, true)
			m.updateTabViewport(tab)
		}
		return nil, true
	case "^":
		if tab := m.effectiveTab(); tab != nil {
			tab.ColCursor = firstVisibleCol(tab.Specs)
			m.updateTabViewport(tab)
		}
		return nil, true
	case "$":
		if tab := m.effectiveTab(); tab != nil {
			tab.ColCursor = lastVisibleCol(tab.Specs)
			m.updateTabViewport(tab)
		}
		return nil, true
	}
	return nil, false
}

// handleNormalKeys processes keys unique to Normal mode.
func (m *Model) handleNormalKeys(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.String() {
	case "D":
		m.toggleDashboard()
		return nil, true
	case "tab":
		if m.detail == nil {
			if m.showDashboard {
				m.showDashboard = false
			}
			m.nextTab()
		}
		return nil, true
	case "shift+tab":
		if m.detail == nil {
			if m.showDashboard {
				m.showDashboard = false
			}
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
	case "c":
		m.hideCurrentColumn()
		return nil, true
	case "C":
		m.showAllColumns()
		return nil, true
	case "/":
		m.openColumnFinder()
		return nil, true
	case "i":
		m.enterEditMode()
		return nil, true
	case keyEnter:
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
	case keyEsc:
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
// on drilldown columns, or follow FK links.
func (m *Model) handleNormalEnter() error {
	tab := m.effectiveTab()
	if tab == nil {
		return nil
	}
	meta, ok := m.selectedRowMeta()
	if !ok {
		return nil
	}

	col := tab.ColCursor
	if col < 0 || col >= len(tab.Specs) {
		return nil
	}
	spec := tab.Specs[col]

	// On a notes column, show the note preview overlay.
	if spec.Kind == cellNotes {
		if c, ok := m.selectedCell(col); ok && c.Value != "" {
			m.notePreviewTitle = spec.Title
			m.notePreviewText = c.Value
			m.showNotePreview = true
		}
		return nil
	}

	// On a drilldown column, open the detail view for that row.
	if spec.Kind == cellDrilldown && m.detail == nil {
		switch tab.Kind {
		case tabMaintenance:
			item, err := m.store.GetMaintenance(meta.ID)
			if err != nil {
				return fmt.Errorf("load maintenance item: %w", err)
			}
			return m.openServiceLogDetail(meta.ID, item.Name)
		case tabAppliances:
			appliance, err := m.store.GetAppliance(meta.ID)
			if err != nil {
				return fmt.Errorf("load appliance: %w", err)
			}
			return m.openApplianceMaintenanceDetail(meta.ID, appliance.Name)
		default:
		}
		return nil
	}

	// On a linked column with a target, follow the FK.
	if spec.Link != nil {
		if c, ok := m.selectedCell(col); ok && c.LinkID > 0 {
			return m.navigateToLink(spec.Link, c.LinkID)
		}
	}

	_ = meta // suppress unused warning
	return nil
}

// handleEditKeys processes keys unique to Edit mode.
func (m *Model) handleEditKeys(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.String() {
	case "a":
		m.startAddForm()
		return m.formInitCmd(), true
	case "e":
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
		} else {
			m.reloadAll()
		}
		return nil, true
	case "r":
		if err := m.popRedo(); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.reloadAll()
		}
		return nil, true
	case "x":
		m.toggleShowDeleted()
		return nil, true
	case "p":
		m.startHouseForm()
		return m.formInitCmd(), true
	case keyEsc:
		m.enterNormalMode()
		return nil, true
	}
	return nil, false
}

func (m *Model) handleCalendarKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "h", "left":
		calendarMove(m.calendar, -1)
	case "l", "right":
		calendarMove(m.calendar, 1)
	case "j", "down":
		calendarMove(m.calendar, 7)
	case "k", "up":
		calendarMove(m.calendar, -7)
	case "H":
		calendarMoveMonth(m.calendar, -1)
	case "L":
		calendarMoveMonth(m.calendar, 1)
	case "[":
		calendarMoveYear(m.calendar, -1)
	case "]":
		calendarMoveYear(m.calendar, 1)
	case keyEnter:
		m.confirmCalendar()
	case keyEsc:
		m.calendar = nil
	}
	return m, nil
}

func (m *Model) confirmCalendar() {
	if m.calendar == nil {
		return
	}
	dateStr := m.calendar.Cursor.Format("2006-01-02")
	if m.calendar.FieldPtr != nil {
		*m.calendar.FieldPtr = dateStr
	}
	if m.calendar.OnConfirm != nil {
		m.calendar.OnConfirm()
	}
	m.calendar = nil
}

// openCalendar opens the date picker for a form field value pointer.
func (m *Model) openCalendar(fieldPtr *string, onConfirm func()) {
	cursor := time.Now()
	var selected time.Time
	hasValue := false
	if fieldPtr != nil && *fieldPtr != "" {
		if t, err := time.Parse("2006-01-02", *fieldPtr); err == nil {
			cursor = t
			selected = t
			hasValue = true
		}
	}
	m.calendar = &calendarState{
		Cursor:    cursor,
		Selected:  selected,
		HasValue:  hasValue,
		FieldPtr:  fieldPtr,
		OnConfirm: onConfirm,
	}
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

func (m *Model) openServiceLogDetail(maintID uint, maintName string) error {
	specs := serviceLogColumnSpecs()
	return m.openDetailWith(detailContext{
		ParentTabIndex: m.active,
		ParentRowID:    maintID,
		Breadcrumb:     "Maintenance > " + maintName,
		Tab: Tab{
			Kind:    tabMaintenance,
			Name:    "Service Log",
			Handler: serviceLogHandler{maintenanceItemID: maintID},
			Specs:   specs,
			Table:   newTable(specsToColumns(specs), m.styles),
		},
	})
}

func (m *Model) openApplianceMaintenanceDetail(applianceID uint, applianceName string) error {
	specs := applianceMaintenanceColumnSpecs()
	return m.openDetailWith(detailContext{
		ParentTabIndex: m.active,
		ParentRowID:    applianceID,
		Breadcrumb:     "Appliances > " + applianceName,
		Tab: Tab{
			Kind:    tabAppliances,
			Name:    "Maintenance",
			Handler: applianceMaintenanceHandler{applianceID: applianceID},
			Specs:   specs,
			Table:   newTable(specsToColumns(specs), m.styles),
		},
	})
}

func (m *Model) openDetailWith(dc detailContext) error {
	m.detail = &dc
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
	return m.reloadTab(&m.detail.Tab)
}

func (m *Model) handleInlineInputKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case keyEsc:
		m.closeInlineInput()
		return m, nil
	case keyEnter:
		m.submitInlineInput()
		return m, nil
	}
	var cmd tea.Cmd
	m.inlineInput.Input, cmd = m.inlineInput.Input.Update(key)
	return m, cmd
}

func (m *Model) submitInlineInput() {
	ii := m.inlineInput
	if ii == nil {
		return
	}
	val := ii.Input.Value()
	if ii.Validate != nil {
		if err := ii.Validate(val); err != nil {
			m.setStatusError(err.Error())
			return
		}
	}
	*ii.FieldPtr = val
	m.snapshotForUndo()
	if err := m.handleFormSubmit(); err != nil {
		m.setStatusError(err.Error())
	} else {
		m.setStatusInfo("Saved.")
		m.reloadAll()
	}
	m.closeInlineInput()
}

func (m *Model) closeInlineInput() {
	m.inlineInput = nil
	m.formKind = formNone
	m.formData = nil
	m.editID = nil
}

// reloadAll refreshes lookups, house profile, all tabs, detail tab, and
// dashboard data. Called after any data mutation.
func (m *Model) reloadAll() {
	if m.store == nil {
		return
	}
	_ = m.loadLookups()
	_ = m.loadHouse()
	_ = m.reloadAllTabs()
	if m.detail != nil {
		_ = m.reloadDetailTab()
	}
	if m.showDashboard {
		_ = m.loadDashboard()
	}
}

func (m *Model) toggleDashboard() {
	m.showDashboard = !m.showDashboard
	if m.showDashboard {
		_ = m.loadDashboard()
		// Close any open detail view when returning to dashboard.
		if m.detail != nil {
			m.closeDetail()
		}
	}
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

	if spec.Kind == cellReadonly || spec.Kind == cellDrilldown {
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
		m.setStatusInfo(fmt.Sprintf("Followed link to ID %d.", targetID))
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
	if m.store == nil {
		return nil
	}
	tab := m.activeTab()
	if tab == nil {
		return nil
	}
	return m.reloadTab(tab)
}

func (m *Model) reloadAllTabs() error {
	if m.store == nil {
		return nil
	}
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
	m.updateTabViewport(tab)
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
	m.reloadAll()
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

// nextVisibleCol returns the next visible column index from current, clamping
// at boundaries. If forward is true it searches right; otherwise left. Returns
// current if already at the edge or no other visible columns exist.
func nextVisibleCol(specs []columnSpec, current int, forward bool) int {
	n := len(specs)
	if n == 0 {
		return 0
	}
	step := 1
	if !forward {
		step = -1
	}
	for i := current + step; i >= 0 && i < n; i += step {
		if specs[i].HideOrder == 0 {
			return i
		}
	}
	return current
}

// firstVisibleCol returns the index of the leftmost visible column.
func firstVisibleCol(specs []columnSpec) int {
	for i, s := range specs {
		if s.HideOrder == 0 {
			return i
		}
	}
	return 0
}

// lastVisibleCol returns the index of the rightmost visible column.
func lastVisibleCol(specs []columnSpec) int {
	for i := len(specs) - 1; i >= 0; i-- {
		if specs[i].HideOrder == 0 {
			return i
		}
	}
	return 0
}

// visibleCount returns the number of non-hidden columns.
func visibleCount(specs []columnSpec) int {
	count := 0
	for _, s := range specs {
		if s.HideOrder == 0 {
			count++
		}
	}
	return count
}

// nextHideOrder returns the next sequence number for hiding a column.
func nextHideOrder(specs []columnSpec) int {
	max := 0
	for _, s := range specs {
		if s.HideOrder > max {
			max = s.HideOrder
		}
	}
	return max + 1
}

func (m *Model) hideCurrentColumn() {
	tab := m.effectiveTab()
	if tab == nil {
		return
	}
	col := tab.ColCursor
	if col < 0 || col >= len(tab.Specs) {
		return
	}
	if tab.Specs[col].HideOrder > 0 {
		return
	}
	if visibleCount(tab.Specs) <= 1 {
		m.setStatusError("Cannot hide the last visible column.")
		return
	}
	tab.Specs[col].HideOrder = nextHideOrder(tab.Specs)
	// Try forward first; if at the right edge fall back to backward.
	next := nextVisibleCol(tab.Specs, col, true)
	if next == col {
		next = nextVisibleCol(tab.Specs, col, false)
	}
	tab.ColCursor = next
	m.updateTabViewport(tab)
	m.setStatusInfo(
		fmt.Sprintf("Hidden: %s. Press C to show all.", tab.Specs[col].Title),
	)
}

func (m *Model) showAllColumns() {
	tab := m.effectiveTab()
	if tab == nil {
		return
	}
	any := false
	for i := range tab.Specs {
		if tab.Specs[i].HideOrder > 0 {
			tab.Specs[i].HideOrder = 0
			any = true
		}
	}
	if any {
		m.updateTabViewport(tab)
		m.setStatusInfo("All columns visible.")
	}
}

func (m *Model) updateAllViewports() {
	if tab := m.activeTab(); tab != nil {
		m.updateTabViewport(tab)
	}
	if m.detail != nil {
		m.updateTabViewport(&m.detail.Tab)
	}
}

func (m *Model) updateTabViewport(tab *Tab) {
	if tab == nil {
		return
	}
	visSpecs, visCells, visColCursor, _, _ := visibleProjection(tab)
	if len(visSpecs) == 0 || visColCursor < 0 {
		tab.ViewOffset = 0
		return
	}
	width := m.effectiveWidth()
	sepW := lipgloss.Width(" │ ")
	fullWidths := columnWidths(visSpecs, visCells, width, sepW)
	ensureCursorVisible(tab, visColCursor, len(visSpecs))
	vpStart, _, _, _ := viewportRange(
		fullWidths, sepW, width, tab.ViewOffset, visColCursor,
	)
	tab.ViewOffset = vpStart
}

// tabIndex returns the position of the given TabKind in the canonical tab
// ordering defined by NewTabs. Derived from the actual slice at init time
// so adding a tab to NewTabs automatically keeps this in sync.
func tabIndex(kind TabKind) int {
	idx, ok := tabKindIndex[kind]
	if !ok {
		return 0
	}
	return idx
}

// tabKindIndex maps each TabKind to its position in the canonical tab slice.
// Populated once at init from NewTabs.
var tabKindIndex = func() map[TabKind]int {
	tabs := NewTabs(DefaultStyles())
	m := make(map[TabKind]int, len(tabs))
	for i, tab := range tabs {
		m[tab.Kind] = i
	}
	return m
}()
