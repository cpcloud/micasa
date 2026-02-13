// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/cpcloud/micasa/internal/data"
	"github.com/cpcloud/micasa/internal/llm"
	"gorm.io/gorm"
)

const (
	keyEsc   = "esc"
	keyEnter = "enter"
)

// Key bindings for help viewport (g/G for top/bottom are not in the
// default viewport keymap).
var (
	helpGotoTop    = key.NewBinding(key.WithKeys("g"))
	helpGotoBottom = key.NewBinding(key.WithKeys("G"))
)

type Model struct {
	store                 *data.Store
	dbPath                string
	configPath            string
	llmClient             *llm.Client
	llmExtraContext       string     // user-provided context appended to prompts
	chat                  *chatState // non-nil when chat overlay is open
	styles                Styles
	tabs                  []Tab
	active                int
	detailStack           []*detailContext // drilldown stack; top is active detail view
	width                 int
	height                int
	helpViewport          *viewport.Model
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
	magMode               bool // easter egg: display numbers as order-of-magnitude
	status                statusMsg
	projectTypes          []data.ProjectType
	maintenanceCategories []data.MaintenanceCategory
	vendors               []data.Vendor
}

func NewModel(store *data.Store, options Options) (*Model, error) {
	styles := DefaultStyles()

	var client *llm.Client
	var extraContext string
	if options.LLMConfig != nil {
		model := options.LLMConfig.Model
		// Prefer the last-used model from the database if available.
		if persisted, err := store.GetLastModel(); err == nil && persisted != "" {
			model = persisted
		} else {
			// No persisted model -- try auto-detecting if the server has exactly one.
			tempClient := llm.NewClient(options.LLMConfig.BaseURL, model)
			if detected := autoDetectModel(tempClient); detected != "" {
				model = detected
				// Persist so we don't re-detect every startup.
				_ = store.PutLastModel(model)
			}
		}
		client = llm.NewClient(options.LLMConfig.BaseURL, model)
		extraContext = options.LLMConfig.ExtraContext
	}

	model := &Model{
		store:           store,
		dbPath:          options.DBPath,
		configPath:      options.ConfigPath,
		llmClient:       client,
		llmExtraContext: extraContext,
		styles:          styles,
		tabs:            NewTabs(styles),
		active:          0,
		showHouse:       false,
		mode:            modeNormal,
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
		if typed.String() == "ctrl+q" {
			m.cancelChatOperations()
			return m, tea.Interrupt
		}
		if typed.String() == "ctrl+c" {
			// Cancel any ongoing LLM operations but don't quit.
			m.cancelChatOperations()
			return m, nil
		}
	case chatChunkMsg:
		// Chunks arriving after chat is closed are harmlessly dropped.
		if m.chat != nil {
			return m, m.handleChatChunk(typed)
		}
		return m, nil
	case sqlResultMsg:
		if m.chat != nil {
			return m, m.handleSQLResult(typed)
		}
		return m, nil
	case pullProgressMsg:
		if m.chat != nil {
			return m, m.handlePullProgress(typed)
		}
		return m, nil
	case modelsListMsg:
		if m.chat != nil {
			m.handleModelsListMsg(typed)
		}
		return m, nil
	case spinner.TickMsg:
		if m.chat != nil && m.chat.Streaming {
			var cmd tea.Cmd
			m.chat.Spinner, cmd = m.chat.Spinner.Update(msg)
			m.refreshChatViewport()
			return m, cmd
		}
		return m, nil
	}

	// Help overlay: delegate scrolling to the viewport, esc or ? dismisses.
	if m.helpViewport != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case keyMsg.String() == keyEsc || keyMsg.String() == "?":
				m.helpViewport = nil
			case key.Matches(keyMsg, helpGotoTop):
				m.helpViewport.GotoTop()
			case key.Matches(keyMsg, helpGotoBottom):
				m.helpViewport.GotoBottom()
			default:
				vp, _ := m.helpViewport.Update(keyMsg)
				m.helpViewport = &vp
			}
		}
		return m, nil
	}

	// Chat overlay: absorb all keys when visible.
	if m.chat != nil && m.chat.Visible {
		if typed, ok := msg.(tea.KeyMsg); ok {
			return m.handleChatKey(typed)
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
// navigation, enter to jump) and blocks keys that affect backgrounded
// widgets. Keys like D, b/f, ?, q fall through to the normal handlers.
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
	case "tab":
		// Block house profile toggle on dashboard.
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
		m.openHelp()
		return nil, true
	case "tab":
		m.showHouse = !m.showHouse
		m.resizeTables()
		return nil, true
	case "m":
		m.magMode = !m.magMode
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
	case "f":
		if !m.inDetail() {
			if m.showDashboard {
				m.showDashboard = false
			}
			m.nextTab()
		}
		return nil, true
	case "b":
		if !m.inDetail() {
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
	case "z":
		if m.toggleHideCompletedProjects() {
			return nil, true
		}
	case "a":
		if m.toggleHideAbandonedProjects() {
			return nil, true
		}
	case "t":
		if m.toggleHideSettledProjects() {
			return nil, true
		}
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
	case "@":
		m.openChat()
		return nil, true
	case "q":
		return tea.Quit, true
	case keyEsc:
		if m.inDetail() {
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
	if spec.Kind == cellDrilldown {
		return m.openDetailForRow(tab, meta.ID, spec.Title)
	}

	// On a linked column with a target, follow the FK.
	if spec.Link != nil {
		if c, ok := m.selectedCell(col); ok && c.LinkID > 0 {
			return m.navigateToLink(spec.Link, c.LinkID)
		}
	}

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
			m.reloadAfterMutation()
		}
		return nil, true
	case "r":
		if err := m.popRedo(); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.reloadAfterMutation()
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

// detail returns the top of the drilldown stack, or nil if no detail view
// is active.
func (m *Model) detail() *detailContext {
	if len(m.detailStack) == 0 {
		return nil
	}
	return m.detailStack[len(m.detailStack)-1]
}

// inDetail reports whether a detail drilldown is active.
func (m *Model) inDetail() bool {
	return len(m.detailStack) > 0
}

// effectiveTab returns the detail tab when a detail view is open, otherwise
// the main active tab. All interaction code should use this.
func (m *Model) effectiveTab() *Tab {
	if dc := m.detail(); dc != nil {
		return &dc.Tab
	}
	return m.activeTab()
}

func (m *Model) openServiceLogDetail(maintID uint, maintName string) error {
	// When drilled from the top-level Maintenance tab, the breadcrumb starts
	// with "Maintenance"; when nested (e.g. Appliances > … > Maint item),
	// the parent tab context is already on the stack so we skip the prefix.
	bc := maintName + breadcrumbSep + "Service Log"
	if !m.inDetail() {
		bc = "Maintenance" + breadcrumbSep + bc
	}
	specs := serviceLogColumnSpecs()
	return m.openDetailWith(detailContext{
		ParentTabIndex: m.active,
		ParentRowID:    maintID,
		Breadcrumb:     bc,
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
		Breadcrumb:     "Appliances" + breadcrumbSep + applianceName,
		Tab: Tab{
			Kind:    tabAppliances,
			Name:    "Maintenance",
			Handler: applianceMaintenanceHandler{applianceID: applianceID},
			Specs:   specs,
			Table:   newTable(specsToColumns(specs), m.styles),
		},
	})
}

func (m *Model) openVendorQuoteDetail(vendorID uint, vendorName string) error {
	specs := vendorQuoteColumnSpecs()
	return m.openDetailWith(detailContext{
		ParentTabIndex: m.active,
		ParentRowID:    vendorID,
		Breadcrumb:     "Vendors" + breadcrumbSep + vendorName + breadcrumbSep + tabQuotes.String(),
		Tab: Tab{
			Kind:    tabVendors,
			Name:    tabQuotes.String(),
			Handler: vendorQuoteHandler{vendorID: vendorID},
			Specs:   specs,
			Table:   newTable(specsToColumns(specs), m.styles),
		},
	})
}

func (m *Model) openVendorJobsDetail(vendorID uint, vendorName string) error {
	specs := vendorJobsColumnSpecs()
	return m.openDetailWith(detailContext{
		ParentTabIndex: m.active,
		ParentRowID:    vendorID,
		Breadcrumb:     "Vendors" + breadcrumbSep + vendorName + breadcrumbSep + "Jobs",
		Tab: Tab{
			Kind:    tabVendors,
			Name:    "Jobs",
			Handler: vendorJobsHandler{vendorID: vendorID},
			Specs:   specs,
			Table:   newTable(specsToColumns(specs), m.styles),
		},
	})
}

func (m *Model) openProjectQuoteDetail(projectID uint, projectTitle string) error {
	specs := projectQuoteColumnSpecs()
	return m.openDetailWith(detailContext{
		ParentTabIndex: m.active,
		ParentRowID:    projectID,
		Breadcrumb:     "Projects" + breadcrumbSep + projectTitle + breadcrumbSep + tabQuotes.String(),
		Tab: Tab{
			Kind:    tabProjects,
			Name:    tabQuotes.String(),
			Handler: projectQuoteHandler{projectID: projectID},
			Specs:   specs,
			Table:   newTable(specsToColumns(specs), m.styles),
		},
	})
}

// openDetailForRow dispatches a drilldown based on the current tab kind and the
// column that was activated. Supports nested drilldowns (e.g. Appliance →
// Maintenance → Service Log).
func (m *Model) openDetailForRow(tab *Tab, rowID uint, colTitle string) error {
	switch {
	case tab.Kind == tabMaintenance && colTitle == "Log":
		item, err := m.store.GetMaintenance(rowID)
		if err != nil {
			return fmt.Errorf("load maintenance item: %w", err)
		}
		return m.openServiceLogDetail(rowID, item.Name)

	case tab.Kind == tabAppliances && colTitle == "Maint":
		appliance, err := m.store.GetAppliance(rowID)
		if err != nil {
			return fmt.Errorf("load appliance: %w", err)
		}
		return m.openApplianceMaintenanceDetail(rowID, appliance.Name)

	case tab.Kind == tabVendors && colTitle == tabQuotes.String():
		vendor, err := m.store.GetVendor(rowID)
		if err != nil {
			return fmt.Errorf("load vendor: %w", err)
		}
		return m.openVendorQuoteDetail(rowID, vendor.Name)

	case tab.Kind == tabVendors && colTitle == "Jobs":
		vendor, err := m.store.GetVendor(rowID)
		if err != nil {
			return fmt.Errorf("load vendor: %w", err)
		}
		return m.openVendorJobsDetail(rowID, vendor.Name)

	case tab.Kind == tabProjects && colTitle == tabQuotes.String():
		project, err := m.store.GetProject(rowID)
		if err != nil {
			return fmt.Errorf("load project: %w", err)
		}
		return m.openProjectQuoteDetail(rowID, project.Title)
	}
	return nil
}

func (m *Model) openDetailWith(dc detailContext) error {
	m.detailStack = append(m.detailStack, &dc)
	if err := m.reloadDetailTab(); err != nil {
		m.detailStack = m.detailStack[:len(m.detailStack)-1]
		return err
	}
	m.resizeTables()
	m.status = statusMsg{}
	return nil
}

func (m *Model) closeDetail() {
	if len(m.detailStack) == 0 {
		return
	}
	top := m.detailStack[len(m.detailStack)-1]
	m.detailStack = m.detailStack[:len(m.detailStack)-1]

	// If there's still a parent detail view on the stack, reload it and
	// restore the cursor to the row that opened the now-closed child.
	if parent := m.detail(); parent != nil {
		if m.store != nil {
			_ = m.reloadTab(&parent.Tab)
		}
		selectRowByID(&parent.Tab, top.ParentRowID)
	} else {
		// Back to a top-level tab.
		m.active = top.ParentTabIndex
		if m.store != nil {
			tab := m.activeTab()
			if tab != nil && tab.Stale {
				_ = m.reloadIfStale(tab)
			} else {
				_ = m.reloadActiveTab()
			}
		}
		if tab := m.activeTab(); tab != nil {
			selectRowByID(tab, top.ParentRowID)
		}
	}
	m.resizeTables()
	m.status = statusMsg{}
}

// closeAllDetails collapses the entire drilldown stack back to the top-level tab.
func (m *Model) closeAllDetails() {
	for len(m.detailStack) > 0 {
		m.closeDetail()
	}
}

func (m *Model) reloadDetailTab() error {
	dc := m.detail()
	if dc == nil || m.store == nil {
		return nil
	}
	return m.reloadTab(&dc.Tab)
}

// reloadAll refreshes lookups, house profile, all tabs, detail tab, and
// dashboard data. Used only at initialization; mutations should call
// reloadAfterMutation for targeted reload.
func (m *Model) reloadAll() {
	if m.store == nil {
		return
	}
	_ = m.loadLookups()
	_ = m.loadHouse()
	_ = m.reloadAllTabs()
	if m.inDetail() {
		_ = m.reloadDetailTab()
	}
	if m.showDashboard {
		_ = m.loadDashboard()
	}
}

// reloadAfterMutation refreshes only the tab the user is looking at and
// marks all other tabs as stale for lazy reload on navigation. Dashboard
// is refreshed only when visible. This avoids reloading 4 idle tabs on
// every save/undo/redo.
func (m *Model) reloadAfterMutation() {
	if m.store == nil {
		return
	}
	_ = m.reloadEffectiveTab()
	m.markNonEffectiveStale()
	if m.showDashboard {
		_ = m.loadDashboard()
	}
}

// markNonEffectiveStale marks all tabs except the effective (active or
// detail-parent) tab as needing a reload on next navigation.
func (m *Model) markNonEffectiveStale() {
	effectiveIdx := m.active
	for i := range m.tabs {
		if i != effectiveIdx {
			m.tabs[i].Stale = true
		}
	}
}

// reloadIfStale reloads a tab only if it is marked stale. The stale flag
// is cleared by reloadTab on success.
func (m *Model) reloadIfStale(tab *Tab) error {
	if tab == nil || !tab.Stale {
		return nil
	}
	return m.reloadTab(tab)
}

func (m *Model) toggleDashboard() {
	m.showDashboard = !m.showDashboard
	if m.showDashboard {
		_ = m.loadDashboard()
		// Close all drilldown levels when returning to dashboard.
		m.closeAllDetails()
	}
}

// switchToTab sets the active tab index, reloads it (lazy if stale), and
// clears the status message. Centralizes the reload-after-switch pattern.
func (m *Model) switchToTab(idx int) {
	m.active = idx
	m.status = statusMsg{}
	tab := m.activeTab()
	if tab != nil && tab.Stale {
		_ = m.reloadIfStale(tab)
	} else {
		_ = m.reloadActiveTab()
	}
}

func (m *Model) nextTab() {
	if len(m.tabs) == 0 {
		return
	}
	m.switchToTab((m.active + 1) % len(m.tabs))
}

func (m *Model) prevTab() {
	if len(m.tabs) == 0 {
		return
	}
	idx := m.active - 1
	if idx < 0 {
		idx = len(m.tabs) - 1
	}
	m.switchToTab(idx)
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

// navigateToLink closes any open drilldown stack, switches to the target tab,
// and selects the row matching the FK.
func (m *Model) navigateToLink(link *columnLink, targetID uint) error {
	m.closeAllDetails()
	m.switchToTab(tabIndex(link.TargetTab))
	tab := m.activeTab()
	if tab == nil {
		return fmt.Errorf("target tab not found")
	}
	if !selectRowByID(tab, targetID) {
		m.setStatusError(fmt.Sprintf("Linked item %d not found (deleted?).", targetID))
	}
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
	if m.inDetail() {
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

func (m *Model) toggleHideCompletedProjects() bool {
	tab := m.projectTabForStatusFilter()
	if tab == nil {
		return false
	}
	tab.HideCompleted = !tab.HideCompleted
	if tab.HideCompleted {
		m.setStatusInfo("Completed projects hidden.")
	} else {
		m.setStatusInfo("Completed projects shown.")
	}
	_ = m.reloadActiveTab()
	return true
}

func (m *Model) toggleHideAbandonedProjects() bool {
	tab := m.projectTabForStatusFilter()
	if tab == nil {
		return false
	}
	tab.HideAbandoned = !tab.HideAbandoned
	if tab.HideAbandoned {
		m.setStatusInfo("Abandoned projects hidden.")
	} else {
		m.setStatusInfo("Abandoned projects shown.")
	}
	_ = m.reloadActiveTab()
	return true
}

func (m *Model) toggleHideSettledProjects() bool {
	tab := m.projectTabForStatusFilter()
	if tab == nil {
		return false
	}
	settledHidden := !tab.HideCompleted || !tab.HideAbandoned
	tab.HideCompleted = settledHidden
	tab.HideAbandoned = settledHidden
	if settledHidden {
		m.setStatusInfo("Settled projects hidden.")
	} else {
		m.setStatusInfo("Settled projects shown.")
	}
	_ = m.reloadActiveTab()
	return true
}

func (m *Model) projectTabForStatusFilter() *Tab {
	if m.inDetail() {
		return nil
	}
	tab := m.activeTab()
	if tab == nil || tab.Kind != tabProjects {
		return nil
	}
	return tab
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
	if tab.Kind == tabProjects && (tab.HideCompleted || tab.HideAbandoned) {
		rows, meta, cellRows = filterProjectRowsByStatusFlags(
			rows,
			meta,
			cellRows,
			tab.HideCompleted,
			tab.HideAbandoned,
		)
	}
	tab.CellRows = cellRows
	tab.Table.SetRows(rows)
	tab.Rows = meta
	tab.Stale = false
	applySorts(tab)
	m.updateTabViewport(tab)
	return nil
}

func filterProjectRowsByStatusFlags(
	rows []table.Row,
	meta []rowMeta,
	cellRows [][]cell,
	hideCompleted bool,
	hideAbandoned bool,
) ([]table.Row, []rowMeta, [][]cell) {
	const projectStatusCol = 3
	if len(rows) != len(meta) || len(rows) != len(cellRows) {
		return rows, meta, cellRows
	}
	filteredRows := rows[:0]
	filteredMeta := meta[:0]
	filteredCells := cellRows[:0]
	for i := range rows {
		if len(cellRows[i]) > projectStatusCol {
			status := cellRows[i][projectStatusCol].Value
			if hideCompleted && status == data.ProjectStatusCompleted {
				continue
			}
			if hideAbandoned && status == data.ProjectStatusAbandoned {
				continue
			}
		}
		filteredRows = append(filteredRows, rows[i])
		filteredMeta = append(filteredMeta, meta[i])
		filteredCells = append(filteredCells, cellRows[i])
	}
	return filteredRows, filteredMeta, filteredCells
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
	m.vendors, err = m.store.ListVendors(false)
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
	if dc := m.detail(); dc != nil {
		dc.Tab.Table.SetHeight(tableHeight)
		dc.Tab.Table.SetWidth(m.width)
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
	kind := m.formKind
	err := m.handleFormSubmit()
	if err != nil {
		m.setStatusError(err.Error())
		return nil
	}
	m.exitForm()
	m.setStatusInfo("Saved.")
	m.reloadAfterFormSave(kind)
	return nil
}

// reloadAfterFormSave picks the minimal reload strategy based on which
// form was just saved. House and vendor mutations need broader refreshes;
// everything else uses the targeted reload.
func (m *Model) reloadAfterFormSave(kind FormKind) {
	if m.store == nil {
		return
	}
	switch kind {
	case formHouse:
		_ = m.loadHouse()
		m.reloadAfterMutation()
	case formVendor:
		_ = m.loadLookups()
		m.reloadAfterMutation()
	default:
		m.reloadAfterMutation()
	}
}

func (m *Model) snapshotForm() {
	m.formSnapshot = fmt.Sprintf("%v", m.formData)
	m.formDirty = false
}

func (m *Model) checkFormDirty() {
	m.formDirty = fmt.Sprintf("%v", m.formData) != m.formSnapshot
}

// openHelp creates a viewport sized to fit the terminal and populated with
// the help content. The viewport handles scroll state and key delegation.
func (m *Model) openHelp() {
	content := m.helpContent()
	lines := strings.Split(content, "\n")

	// Chrome: border (2) + padding (2) + gap + rule + hint (4) = 8 lines.
	maxH := m.effectiveHeight() - 2
	if maxH < 10 {
		maxH = 10
	}
	viewH := maxH - 8
	if viewH < 3 {
		viewH = 3
	}
	// If content fits, no scrolling needed.
	if len(lines) <= viewH {
		viewH = len(lines)
	}

	// Lock width to the widest content line so the overlay never resizes.
	maxW := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > maxW {
			maxW = w
		}
	}

	vp := viewport.New(maxW, viewH)
	vp.SetContent(content)
	// Disable horizontal scroll to avoid conflicts with table navigation.
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)
	m.helpViewport = &vp
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
	value := ii.Input.Value()
	if ii.Validate != nil {
		if err := ii.Validate(value); err != nil {
			m.setStatusError(err.Error())
			return
		}
	}
	*ii.FieldPtr = value
	kind := ii.FormKind
	m.snapshotForUndo()
	if err := m.handleFormSubmit(); err != nil {
		m.setStatusError(err.Error())
		return
	}
	m.closeInlineInput()
	m.setStatusInfo("Saved.")
	m.reloadAfterFormSave(kind)
}

func (m *Model) closeInlineInput() {
	m.inlineInput = nil
	m.formKind = formNone
	m.formData = nil
	m.editID = nil
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
	defaultWidth    = 80
	defaultHeight   = 24
	minUsableWidth  = 80
	minUsableHeight = 24
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

// overlayContentWidth returns the clamped content width for overlay boxes
// (dashboard, note preview). Accounts for border (2), padding (4), and
// breathing room (6) = 12 total, clamped to [30, 72].
func (m *Model) overlayContentWidth() int {
	w := m.effectiveWidth() - 12
	if w > 72 {
		w = 72
	}
	if w < 30 {
		w = 30
	}
	return w
}

func (m *Model) terminalTooSmall() bool {
	return m.effectiveWidth() < minUsableWidth || m.effectiveHeight() < minUsableHeight
}

// hasActiveOverlay returns true when any overlay is currently shown. Overlays
// include dashboard, calendar, note preview, column finder, and help. When
// true, main tab keybindings should be hidden from the status bar since they
// are not accessible.
func (m *Model) hasActiveOverlay() bool {
	return m.showDashboard ||
		m.calendar != nil ||
		m.showNotePreview ||
		m.columnFinder != nil ||
		m.helpViewport != nil
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
	if dc := m.detail(); dc != nil {
		m.updateTabViewport(&dc.Tab)
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

// autoDetectModel checks if the LLM server has exactly one model available
// and returns it. Returns "" if the server is unreachable or has zero/multiple
// models (ambiguous cases where manual config is safer).
func autoDetectModel(client *llm.Client) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	models, err := client.ListModels(ctx)
	if err != nil || len(models) != 1 {
		return ""
	}
	return models[0]
}
