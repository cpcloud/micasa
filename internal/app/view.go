// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

func (m *Model) buildView() string {
	if m.mode == modeForm && m.form != nil && m.formKind == formHouse {
		return m.formFullScreen()
	}

	base := m.buildBaseView()

	if m.showDashboard {
		fg := cancelFaint(m.buildDashboardOverlay())
		base = overlay.Composite(fg, dimBackground(base), overlay.Center, overlay.Center, 0, 0)
	}

	if m.showHelp {
		fg := cancelFaint(m.buildHelpOverlay())
		base = overlay.Composite(fg, dimBackground(base), overlay.Center, overlay.Center, 0, 0)
	}

	return base
}

// buildBaseView renders the normal table/detail/form view with house, tabs,
// content area, and status bar. Used as the background when the dashboard
// overlay is active.
func (m *Model) buildBaseView() string {
	house := m.houseView()

	var tabs, tabLine, content string
	if m.detail != nil {
		tabs = m.breadcrumbView()
		tabLine = m.tabUnderline()
		if m.mode == modeForm && m.form != nil {
			content = m.form.View()
		} else if tab := m.effectiveTab(); tab != nil {
			content = m.tableView(tab)
		}
	} else {
		tabs = m.tabsView()
		tabLine = m.tabUnderline()
		if m.mode == modeForm && m.form != nil {
			content = m.form.View()
		} else if tab := m.effectiveTab(); tab != nil {
			content = m.tableView(tab)
		}
	}
	status := m.statusView()

	// Right-align db path on the tab row when set.
	if m.dbPath != "" {
		dbLabel := m.styles.HeaderLabel.Render("db")
		tabsW := lipgloss.Width(tabs)
		labelW := lipgloss.Width(dbLabel) + 1 // "db "
		minGap := 2                           // breathing room between tabs and path
		available := m.effectiveWidth() - tabsW - labelW - minGap
		if available > 5 {
			path := truncateLeft(m.dbPath, available)
			label := dbLabel + " " + m.styles.HeaderHint.Render(path)
			gap := m.effectiveWidth() - tabsW - lipgloss.Width(label)
			if gap > 0 {
				tabs += strings.Repeat(" ", gap) + label
			}
		}
	}

	// Assemble upper portion with intentional spacing.
	upper := lipgloss.JoinVertical(lipgloss.Left, house, "", tabs, tabLine)
	if content != "" {
		upper = lipgloss.JoinVertical(lipgloss.Left, upper, content)
	}

	// Anchor the status bar to the terminal bottom.
	upperH := lipgloss.Height(upper)
	statusH := lipgloss.Height(status)
	gap := m.height - upperH - statusH + 1
	if gap < 1 {
		gap = 1
	}

	var b strings.Builder
	b.WriteString(upper)
	b.WriteString(strings.Repeat("\n", gap))
	b.WriteString(status)
	return clampLines(b.String(), m.effectiveWidth())
}

// buildDashboardOverlay renders the dashboard content inside a bordered box
// with navigation hints, suitable for compositing over the base view.
func (m *Model) buildDashboardOverlay() string {
	title := m.styles.DashSection.Render("Dashboard")
	content := m.dashboardView()

	// Navigation hints inside the overlay.
	items := []string{m.helpItem("j/k", "navigate")}
	if m.dashNavCount() > 0 {
		items = append(items, m.helpItem("enter", "jump to"))
	}
	items = append(items,
		m.helpItem("D", "close"),
		m.helpItem("?", "help"),
	)
	hints := joinWithSeparator(m.helpSeparator(), items...)

	boxContent := lipgloss.JoinVertical(
		lipgloss.Left, title, "", content, "", hints,
	)

	// Content width: terminal minus border (2) + padding (4) + breathing room (6).
	contentW := m.effectiveWidth() - 12
	if contentW > 72 {
		contentW = 72
	}
	if contentW < 30 {
		contentW = 30
	}
	maxH := m.effectiveHeight() - 4
	if maxH < 10 {
		maxH = 10
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 2).
		Width(contentW).
		MaxHeight(maxH).
		Render(boxContent)
}

func (m *Model) tabsView() string {
	pinned := m.mode == modeEdit
	tabs := make([]string, 0, len(m.tabs))
	for i, tab := range m.tabs {
		if i == m.active {
			tabs = append(tabs, m.styles.TabActive.Render(tab.Name))
		} else if pinned {
			tabs = append(tabs, m.styles.TabLocked.Render(tab.Name))
		} else {
			tabs = append(tabs, m.styles.TabInactive.Render(tab.Name))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (m *Model) breadcrumbView() string {
	if m.detail == nil {
		return ""
	}
	arrow := m.styles.BreadcrumbArrow.Render(" > ")
	parts := strings.Split(m.detail.Breadcrumb, " > ")
	rendered := make([]string, len(parts))
	for i, p := range parts {
		if i < len(parts)-1 {
			rendered[i] = m.styles.HeaderHint.Render(p)
		} else {
			rendered[i] = m.styles.Breadcrumb.Render(p)
		}
	}
	crumb := strings.Join(rendered, arrow)
	back := m.styles.HeaderHint.Render(" (")
	back += m.keycap("esc")
	back += m.styles.HeaderHint.Render(" back)")
	return crumb + back
}

func (m *Model) tabUnderline() string {
	return m.styles.TabUnderline.Render(strings.Repeat("━", m.effectiveWidth()))
}

func (m *Model) statusView() string {
	if m.mode == modeForm {
		dirtyIndicator := m.styles.FormClean.Render("○ saved")
		if m.formDirty {
			dirtyIndicator = m.styles.FormDirty.Render("● unsaved")
		}
		help := joinWithSeparator(
			m.helpSeparator(),
			dirtyIndicator,
			m.helpItem("ctrl+s", "save"),
			m.helpItem("esc", "cancel"),
			m.helpItem("ctrl+c", "quit"),
		)
		return m.withStatusMessage(help)
	}

	// Both badges render at the same width to prevent layout shift.
	badgeWidth := lipgloss.Width(m.styles.ModeNormal.Render("NORMAL"))
	modeBadge := m.styles.ModeNormal.Render("NORMAL")
	if m.mode == modeEdit {
		modeBadge = m.styles.ModeEdit.
			Width(badgeWidth).
			Align(lipgloss.Center).
			Render("EDIT")
	}

	var help string
	if m.mode == modeNormal {
		items := []string{modeBadge}
		if m.detail == nil {
			items = append(items, m.helpItem("tab", "switch"))
		}
		items = append(items,
			m.helpItem("h/l", "col"),
			m.helpItem("s", "sort"),
		)
		if hint := m.enterHint(); hint != "" {
			items = append(items, m.helpItem("enter", hint))
		}
		items = append(items,
			m.helpItem("c", "hide col"),
		)
		if hint := m.hiddenHint(); hint != "" {
			items = append(items, hint)
		}
		items = append(items,
			m.helpItem("i", "edit"),
			m.helpItem("D", "dashboard"),
			m.helpItem("H", "house"),
			m.helpItem("?", "help"),
		)
		if m.detail != nil {
			items = append(items, m.helpItem("esc", "back"))
		} else {
			items = append(items, m.helpItem("q", "quit"))
		}
		help = joinWithSeparator(m.helpSeparator(), items...)
	} else {
		help = joinWithSeparator(
			m.helpSeparator(),
			modeBadge,
			m.helpItem("a", "add"),
			m.helpItem("e", m.editHint()),
			m.helpItem("d", "del/restore"),
			m.helpItem("u", "undo"),
			m.helpItem("r", "redo"),
			m.deletedHint(m.effectiveTab()),
			m.helpItem("p", "profile"),
			m.helpItem("esc", "normal"),
		)
	}

	return m.withStatusMessage(help)
}

// hiddenHint returns a status bar label listing hidden column names when any
// are hidden, or "" when all are visible.
func (m *Model) hiddenHint() string {
	tab := m.effectiveTab()
	if tab == nil {
		return ""
	}
	names := hiddenColumnNames(tab.Specs)
	if len(names) == 0 {
		return ""
	}
	label := "hidden: " + strings.Join(names, ", ")
	hint := m.keycap("C") + " " + m.styles.HeaderHint.Render("show all")
	return m.styles.HeaderHint.Render(label) + "  " + hint
}

func (m *Model) deletedHint(tab *Tab) string {
	key := m.keycap("x")
	label := m.styles.HeaderHint.Render("deleted")
	if tab != nil && tab.ShowDeleted {
		label = m.styles.DeletedLabel.Render("deleted")
	}
	return strings.TrimSpace(fmt.Sprintf("%s %s", key, label))
}

// withStatusMessage renders the help line, prepending the status message if set.
func (m *Model) withStatusMessage(helpLine string) string {
	if m.status.Text == "" {
		return helpLine
	}
	style := m.styles.Info
	if m.status.Kind == statusError {
		style = m.styles.Error
	}
	return lipgloss.JoinVertical(lipgloss.Left, style.Render(m.status.Text), helpLine)
}

func (m *Model) editHint() string {
	tab := m.effectiveTab()
	if tab == nil {
		return "edit"
	}
	col := tab.ColCursor
	if col < 0 || col >= len(tab.Specs) {
		return "edit"
	}
	spec := tab.Specs[col]
	// Show "follow link" hint when on a linked cell with a target.
	if spec.Link != nil {
		if c, ok := m.selectedCell(col); ok && c.LinkID > 0 {
			return "follow " + spec.Link.Relation
		}
	}
	if spec.Kind == cellReadonly {
		return "edit"
	}
	return "edit: " + spec.Title
}

// enterHint returns a contextual label for the enter key in Normal mode,
// or "" if enter has no action on the current column.
func (m *Model) enterHint() string {
	tab := m.effectiveTab()
	if tab == nil {
		return ""
	}
	col := tab.ColCursor
	if col < 0 || col >= len(tab.Specs) {
		return ""
	}
	spec := tab.Specs[col]
	if spec.Kind == cellDrilldown && m.detail == nil {
		switch tab.Kind {
		case tabMaintenance:
			return "service log"
		case tabAppliances:
			return "maintenance"
		}
	}
	if spec.Link != nil {
		if c, ok := m.selectedCell(col); ok && c.LinkID > 0 {
			return "follow " + spec.Link.Relation
		}
	}
	return ""
}

func (m *Model) formFullScreen() string {
	formContent := m.form.View()
	status := m.statusView()
	panel := lipgloss.JoinVertical(lipgloss.Left, formContent, "", status)
	return m.centerPanel(panel, 1)
}

func (m *Model) buildHelpOverlay() string {
	// helpView() already renders a bordered box with padding.
	return m.helpView()
}

// centerPanel centers a rendered panel within the terminal dimensions.
// minPadTop sets the minimum top padding (e.g. 1 to keep a gap above forms).
func (m *Model) centerPanel(panel string, minPadTop int) string {
	width := m.effectiveWidth()
	height := m.effectiveHeight()
	panelH := lipgloss.Height(panel)
	panelW := lipgloss.Width(panel)
	padTop := (height - panelH) / 2
	if padTop < minPadTop {
		padTop = minPadTop
	}
	padLeft := (width - panelW) / 2
	if padLeft < 0 {
		padLeft = 0
	}
	lines := strings.Split(panel, "\n")
	var b strings.Builder
	for i := 0; i < padTop; i++ {
		b.WriteString("\n")
	}
	indent := strings.Repeat(" ", padLeft)
	for i, line := range lines {
		b.WriteString(indent)
		b.WriteString(line)
		if i < len(lines)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m *Model) helpView() string {
	type binding struct {
		key  string
		desc string
	}
	sections := []struct {
		title    string
		bindings []binding
	}{
		{
			title: "Normal Mode",
			bindings: []binding{
				{"j/k", "Move through rows"},
				{"h/l", "Move through columns"},
				{"^/$", "Jump to first/last column"},
				{"g/G", "Jump to first/last row"},
				{"d/u", "Half page down/up"},
				{"tab/shift+tab", "Switch tabs"},
				{"H", "Toggle house profile"},
				{"s", "Sort by column (cycle asc/desc/off)"},
				{"S", "Clear all sorts"},
				{"enter", "Open detail / follow link"},
				{"c", "Hide current column"},
				{"C", "Show all columns"},
				{"D", "Toggle dashboard"},
				{"i", "Enter Edit mode"},
				{"?", "Help"},
				{"q", "Quit"},
			},
		},
		{
			title: "Edit Mode",
			bindings: []binding{
				{"a", "Add new entry"},
				{"e", "Edit cell (or full row on ID column)"},
				{"d", "Toggle delete/restore"},
				{"u", "Undo last edit"},
				{"r", "Redo last undo"},
				{"x", "Toggle showing deleted items"},
				{"p", "Edit house profile"},
				{"esc", "Back to Normal mode"},
			},
		},
		{
			title: "Forms",
			bindings: []binding{
				{"1-9", "Jump to Nth option (selects only)"},
				{"ctrl+s", "Save immediately"},
				{"esc", "Cancel and discard"},
			},
		},
	}

	var b strings.Builder
	b.WriteString(m.styles.HeaderTitle.Render(" Keyboard Shortcuts "))
	b.WriteString("\n\n")
	for i, section := range sections {
		b.WriteString(m.styles.HeaderSection.Render(" " + section.title + " "))
		b.WriteString("\n")
		for _, bind := range section.bindings {
			keys := m.renderKeys(bind.key)
			desc := m.styles.HeaderHint.Render(bind.desc)
			b.WriteString(fmt.Sprintf("  %s  %s\n", keys, desc))
		}
		if i < len(sections)-1 {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(m.styles.HeaderHint.Render("Press "))
	b.WriteString(m.renderKeys("esc"))
	b.WriteString(m.styles.HeaderHint.Render(" or "))
	b.WriteString(m.renderKeys("?"))
	b.WriteString(m.styles.HeaderHint.Render(" to close"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 2).
		Render(b.String())
	return box
}

// tableView orchestrates the full table rendering: visible projection,
// column sizing, horizontal scroll viewport, header/divider/rows, and
// hidden-column badge line.
func (m *Model) tableView(tab *Tab) string {
	if tab == nil || len(tab.Specs) == 0 {
		return ""
	}

	width := m.effectiveWidth()
	normalSep := m.styles.TableSeparator.Render(" │ ")
	normalDiv := m.styles.TableSeparator.Render("─┼─")
	sepW := lipgloss.Width(normalSep)

	vp := computeTableViewport(tab, width, normalSep, m.styles)
	if len(vp.Specs) == 0 {
		return ""
	}
	header := renderHeaderRow(
		vp.Specs,
		vp.Widths,
		vp.CollapsedSeps,
		vp.Cursor,
		vp.Sorts,
		vp.HasLeft,
		vp.HasRight,
		m.styles,
	)
	divider := renderDivider(vp.Widths, vp.PlainSeps, normalDiv, m.styles.TableSeparator)

	// Badge line accounts for 1 row of vertical space when visible.
	badges := renderHiddenBadges(tab.Specs, tab.ColCursor, m.styles)
	badgeChrome := 0
	if badges != "" {
		badgeChrome = 1
	}

	effectiveHeight := tab.Table.Height() - badgeChrome
	if effectiveHeight < 2 {
		effectiveHeight = 2
	}
	rows := renderRows(
		vp.Specs,
		vp.Cells,
		tab.Rows,
		vp.Widths,
		vp.PlainSeps,
		vp.CollapsedSeps,
		tab.Table.Cursor(),
		vp.Cursor,
		effectiveHeight,
		m.styles,
	)

	// Assemble body (header + divider + data rows).
	bodyParts := []string{header, divider}
	if len(rows) == 0 {
		bodyParts = append(bodyParts, m.styles.Empty.Render("No entries yet."))
	} else {
		bodyParts = append(bodyParts, strings.Join(rows, "\n"))
	}
	if badges != "" {
		tableWidth := sumInts(vp.Widths)
		if len(vp.Widths) > 1 {
			tableWidth += (len(vp.Widths) - 1) * sepW
		}
		centered := lipgloss.PlaceHorizontal(tableWidth, lipgloss.Center, badges)
		bodyParts = append(bodyParts, centered)
	}
	return joinVerticalNonEmpty(bodyParts...)
}

// dimBackground applies ANSI faint (dim) to an already-styled string. It
// replaces full resets with reset+faint so the dim survives through existing
// color codes.
func dimBackground(s string) string {
	dimmed := strings.ReplaceAll(s, "\033[0m", "\033[0;2m")
	return "\033[2m" + dimmed + "\033[0m"
}

// cancelFaint prepends each line with the ANSI "normal intensity" code so that
// faint state inherited from a composited background does not bleed into the
// foreground overlay.
func cancelFaint(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "\033[22m" + line
	}
	return strings.Join(lines, "\n")
}

// --- Keycap rendering ---

func (m *Model) helpItem(keys, label string) string {
	keycaps := m.renderKeys(keys)
	desc := m.styles.HeaderHint.Render(label)
	return strings.TrimSpace(fmt.Sprintf("%s %s", keycaps, desc))
}

func (m *Model) helpSeparator() string {
	return m.styles.HeaderHint.Render(" · ")
}

func (m *Model) renderKeys(keys string) string {
	// A bare "/" is a single key, not a separator between two keys.
	if strings.TrimSpace(keys) == "/" {
		return m.keycap("/")
	}
	parts := strings.Split(keys, "/")
	rendered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		rendered = append(rendered, m.keycap(part))
	}
	return joinWithSeparator(m.styles.HeaderHint.Render(" · "), rendered...)
}

func (m *Model) keycap(value string) string {
	// Detect bare uppercase letters and render as SHIFT+X.
	if len(value) == 1 && value[0] >= 'A' && value[0] <= 'Z' {
		return m.styles.Keycap.Render("SHIFT+" + value)
	}
	return m.styles.Keycap.Render(strings.ToUpper(value))
}

// --- General view utilities ---

// filterNonBlank returns only the values that have visible content.
func filterNonBlank(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func joinInline(values ...string) string {
	if f := filterNonBlank(values); len(f) > 0 {
		return lipgloss.JoinHorizontal(lipgloss.Center, f...)
	}
	return ""
}

func joinVerticalNonEmpty(values ...string) string {
	if f := filterNonBlank(values); len(f) > 0 {
		return lipgloss.JoinVertical(lipgloss.Left, f...)
	}
	return ""
}

func joinWithSeparator(sep string, values ...string) string {
	return strings.Join(filterNonBlank(values), sep)
}

func joinNonEmpty(values []string, sep string) string {
	return joinWithSeparator(sep, values...)
}

// clampLines truncates each line in s to maxW visible columns, appending "…"
// when truncation occurs. ANSI escape sequences are preserved.
func clampLines(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) > maxW {
			lines[i] = ansi.Truncate(line, maxW, "…")
		}
	}
	return strings.Join(lines, "\n")
}

// truncateLeft trims s from the left so the result fits within maxW visible
// columns, prepending "…" when truncation occurs.
func truncateLeft(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxW {
		return s
	}
	ellipsis := "…"
	ellW := lipgloss.Width(ellipsis)
	if maxW <= ellW {
		return ansi.Truncate(s, maxW, "")
	}

	// Walk runes from the end, accumulating visible width, until adding another
	// rune would exceed maxW once we include the ellipsis.
	runes := []rune(s)
	keptW := 0
	cut := len(runes)
	for i := len(runes) - 1; i >= 0; i-- {
		cw := lipgloss.Width(string(runes[i]))
		if keptW+cw+ellW > maxW {
			break
		}
		keptW += cw
		cut = i
	}
	if cut >= len(runes) {
		return ansi.Truncate(s, maxW, "")
	}
	return ellipsis + string(runes[cut:])
}
