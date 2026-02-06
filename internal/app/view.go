package app

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/micasa/micasa/internal/data"
)

func (m *Model) buildView() string {
	if m.showHelp {
		return m.helpFullScreen()
	}
	if m.mode == modeForm && m.form != nil && m.formKind == formHouse {
		return m.formFullScreen()
	}

	house := m.houseView()
	tabs := m.tabsView()
	tabLine := m.tabUnderline()
	content := ""
	if m.mode == modeForm && m.form != nil {
		content = m.form.View()
	} else if tab := m.activeTab(); tab != nil {
		content = m.tableView(tab)
	}
	status := m.statusView()

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
	return b.String()
}

func (m *Model) houseView() string {
	if !m.hasHouse {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			joinInline(
				m.styles.HeaderTitle.Render("House"),
				m.styles.HeaderBadge.Render("setup"),
				m.keycap("H"),
			),
			m.styles.HeaderHint.Render("Complete the form to add a house profile."),
		)
		return m.headerBox(content)
	}
	if m.showHouse {
		return m.headerBox(m.houseExpanded())
	}
	return m.headerBox(m.houseCollapsed())
}

func (m *Model) houseCollapsed() string {
	title := m.styles.HeaderTitle.Render("House")
	badge := m.styles.HeaderBadge.Render("▸")
	sep := m.styles.HeaderHint.Render(" · ")
	hint := m.styles.HeaderHint
	val := m.styles.HeaderValue
	stats := joinStyledParts(sep,
		styledPart(val, m.house.Nickname),
		styledPart(hint, formatCityState(m.house)),
		styledPart(hint, bedBathLabel(m.house)),
		styledPart(hint, sqftLabel(m.house.SquareFeet)),
		styledPart(hint, formatInt(m.house.YearBuilt)),
	)
	return joinInline(title, badge) + "  " + stats
}

func (m *Model) houseExpanded() string {
	title := m.styles.HeaderTitle.Render("House")
	badge := m.styles.HeaderBadge.Render("▾")
	hint := m.styles.HeaderHint
	val := m.styles.HeaderValue
	sep := hint.Render(" · ")

	identity := joinStyledParts(sep,
		styledPart(val, m.house.Nickname),
		styledPart(hint, formatAddress(m.house)),
	)
	titleLine := joinInline(title, badge) + "  " + identity

	structNums := joinStyledParts(sep,
		styledPart(val, formatInt(m.house.YearBuilt)),
		styledPart(val, sqftLabel(m.house.SquareFeet)),
		styledPart(val, lotLabel(m.house.LotSquareFeet)),
		styledPart(val, bedBathLabel(m.house)),
	)
	structMaterials := joinStyledParts(sep,
		m.hlv("fnd", m.house.FoundationType),
		m.hlv("wir", m.house.WiringType),
		m.hlv("roof", m.house.RoofType),
		m.hlv("ext", m.house.ExteriorType),
		m.hlv("bsmt", m.house.BasementType),
	)
	structure := m.houseSection("Structure", structNums, structMaterials)

	utilLine := joinStyledParts(sep,
		m.hlv("heat", m.house.HeatingType),
		m.hlv("cool", m.house.CoolingType),
		m.hlv("water", m.house.WaterSource),
		m.hlv("sewer", m.house.SewerType),
		m.hlv("park", m.house.ParkingType),
	)
	utilities := m.houseSection("Utilities", utilLine)

	finLine1 := joinStyledParts(sep,
		m.hlv("ins", m.house.InsuranceCarrier),
		m.hlv("policy", m.house.InsurancePolicy),
		m.hlv("renew", data.FormatDate(m.house.InsuranceRenewal)),
	)
	finLine2 := joinStyledParts(sep,
		m.hlv("tax", data.FormatOptionalCents(m.house.PropertyTaxCents)),
		m.hlv("hoa", hoaSummary(m.house)),
	)
	financial := m.houseSection("Financial", finLine1, finLine2)

	content := joinVerticalNonEmpty(titleLine, "", structure, utilities, financial)
	art := m.houseArt()
	if art == "" {
		return content
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, content, "   ", art)
}

func (m *Model) tabsView() string {
	tabs := make([]string, 0, len(m.tabs))
	for i, tab := range m.tabs {
		if i == m.active {
			tabs = append(tabs, m.styles.TabActive.Render(tab.Name))
		} else {
			tabs = append(tabs, m.styles.TabInactive.Render(tab.Name))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
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

	modeBadge := m.styles.ModeNormal.Render("NORMAL")
	if m.mode == modeEdit {
		modeBadge = m.styles.ModeEdit.Render("EDIT")
	}

	var help string
	if m.mode == modeNormal {
		help = joinWithSeparator(
			m.helpSeparator(),
			modeBadge,
			m.helpItem("h/l", "col"),
			m.helpItem("i", "edit mode"),
			m.helpItem("enter", "edit"),
			m.deletedHint(m.activeTab()),
			m.helpItem("H", "house"),
			m.helpItem("?", "help"),
			m.helpItem("q", "quit"),
		)
	} else {
		help = joinWithSeparator(
			m.helpSeparator(),
			modeBadge,
			m.helpItem("a", "add"),
			m.helpItem("e", m.editHint()),
			m.helpItem("d", "del"),
			m.helpItem("u", "undo"),
			m.deletedHint(m.activeTab()),
			m.helpItem("p", "profile"),
			m.helpItem("esc", "normal"),
		)
	}

	dbLabel := m.styles.DBHint.Render(m.dbPath)
	leftWidth := ansi.StringWidth(help)
	dbWidth := ansi.StringWidth(dbLabel)
	width := m.effectiveWidth()
	gap := width - leftWidth - dbWidth
	if gap < 2 {
		gap = 2
	}
	helpLine := help + strings.Repeat(" ", gap) + dbLabel
	return m.withStatusMessage(helpLine)
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
	tab := m.activeTab()
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

func (m *Model) formFullScreen() string {
	formContent := m.form.View()
	status := m.statusView()
	panel := lipgloss.JoinVertical(lipgloss.Left, formContent, "", status)
	return m.centerPanel(panel, 1)
}

func (m *Model) helpFullScreen() string {
	return m.centerPanel(m.helpView(), 0)
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
				{"g/G", "Jump to first/last row"},
				{"d/u", "Half page down/up"},
				{"tab/shift+tab", "Switch tabs"},
				{"H", "Toggle house profile"},
				{"x", "Toggle showing deleted items"},
				{"enter", "Edit current cell"},
				{"i", "Enter Edit mode"},
				{"?", "This help menu"},
				{"q", "Quit"},
			},
		},
		{
			title: "Edit Mode",
			bindings: []binding{
				{"a", "Add new entry"},
				{"e / enter", "Edit cell (or full row on ID column)"},
				{"d", "Delete selected entry"},
				{"u", "Restore last deleted entry"},
				{"p", "Edit house profile"},
				{"esc", "Back to Normal mode"},
			},
		},
		{
			title: "Forms",
			bindings: []binding{
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

func (m *Model) tableView(tab *Tab) string {
	if tab == nil || len(tab.Specs) == 0 {
		return ""
	}
	width := m.effectiveWidth()
	separator := m.styles.TableSeparator.Render(" │ ")
	dividerSep := m.styles.TableSeparator.Render("─┼─")
	widths := columnWidths(tab.Specs, tab.CellRows, width, lipgloss.Width(separator))
	header := renderHeaderRow(tab.Specs, widths, separator, tab.ColCursor, m.styles)
	divider := renderDivider(widths, dividerSep, m.styles.TableSeparator)
	rows := renderRows(
		tab.Specs,
		tab.CellRows,
		tab.Rows,
		widths,
		separator,
		tab.Table.Cursor(),
		tab.ColCursor,
		tab.Table.Height(),
		m.styles,
	)
	if len(rows) == 0 {
		empty := m.styles.Empty.Render("No entries yet.")
		rows = []string{empty}
	}
	return joinVerticalNonEmpty(header, divider, strings.Join(rows, "\n"))
}

func renderHeaderRow(
	specs []columnSpec,
	widths []int,
	separator string,
	colCursor int,
	styles Styles,
) string {
	cells := make([]string, 0, len(specs))
	for i, spec := range specs {
		width := safeWidth(widths, i)
		title := spec.Title
		if spec.Link != nil {
			title = title + " " + styles.LinkIndicator.Render(spec.Link.Relation)
		}
		text := formatCell(title, width, spec.Align)
		if i == colCursor {
			cells = append(cells, styles.ColActiveHeader.Render(text))
		} else {
			cells = append(cells, styles.TableHeader.Render(text))
		}
	}
	return strings.Join(cells, separator)
}

func renderDivider(
	widths []int,
	separator string,
	style lipgloss.Style,
) string {
	parts := make([]string, 0, len(widths))
	for _, width := range widths {
		if width < 1 {
			width = 1
		}
		parts = append(parts, style.Render(strings.Repeat("─", width)))
	}
	return strings.Join(parts, separator)
}

func renderRows(
	specs []columnSpec,
	rows [][]cell,
	meta []rowMeta,
	widths []int,
	separator string,
	cursor int,
	colCursor int,
	height int,
	styles Styles,
) []string {
	total := len(rows)
	if total == 0 {
		return nil
	}
	if height <= 0 {
		height = total
	}
	start, end := visibleRange(total, height, cursor)
	rendered := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		selected := i == cursor
		deleted := i < len(meta) && meta[i].Deleted
		row := renderRow(specs, rows[i], widths, separator, selected, deleted, colCursor, styles)
		rendered = append(rendered, row)
	}
	return rendered
}

func renderRow(
	specs []columnSpec,
	row []cell,
	widths []int,
	separator string,
	selected bool,
	deleted bool,
	colCursor int,
	styles Styles,
) string {
	cells := make([]string, 0, len(specs))
	for i, spec := range specs {
		width := safeWidth(widths, i)
		var cellValue cell
		if i < len(row) {
			cellValue = row[i]
		}
		rendered := renderCell(cellValue, spec, width, styles)
		if selected && i == colCursor {
			rendered = styles.CellActive.Render(rendered)
		} else if selected {
			rendered = styles.TableSelected.Render(rendered)
		}
		cells = append(cells, rendered)
	}
	rendered := strings.Join(cells, separator)
	if deleted {
		rendered = styles.Deleted.Render(rendered)
	}
	return rendered
}

func renderCell(
	cellValue cell,
	spec columnSpec,
	width int,
	styles Styles,
) string {
	if width < 1 {
		width = 1
	}
	value := strings.TrimSpace(cellValue.Value)
	style := cellStyle(cellValue.Kind, styles)
	if value == "" {
		value = "—"
		style = styles.Empty
	} else if cellValue.Kind == cellStatus {
		if s, ok := styles.StatusStyles[value]; ok {
			style = s
		}
	}
	aligned := formatCell(value, width, spec.Align)
	return style.Render(aligned)
}

func cellStyle(kind cellKind, styles Styles) lipgloss.Style {
	switch kind {
	case cellMoney:
		return styles.Money
	case cellReadonly:
		return styles.Readonly
	default:
		return lipgloss.NewStyle()
	}
}

func formatCell(value string, width int, align alignKind) string {
	if width < 1 {
		return ""
	}
	truncated := ansi.Truncate(value, width, "…")
	current := lipgloss.Width(truncated)
	if current >= width {
		return truncated
	}
	padding := width - current
	switch align {
	case alignRight:
		return strings.Repeat(" ", padding) + truncated
	default:
		return truncated + strings.Repeat(" ", padding)
	}
}

func visibleRange(total, height, cursor int) (int, int) {
	if total <= height {
		return 0, total
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > total {
		end = total
		start = end - height
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

func columnWidths(
	specs []columnSpec,
	rows [][]cell,
	width int,
	separatorWidth int,
) []int {
	columnCount := len(specs)
	if columnCount == 0 {
		return nil
	}
	available := width - separatorWidth*(columnCount-1)
	if available < columnCount {
		available = columnCount
	}
	widths := make([]int, columnCount)
	for i, spec := range specs {
		maxWidth := lipgloss.Width(spec.Title)
		if spec.Max > 0 && maxWidth > spec.Max {
			maxWidth = spec.Max
		}
		for _, row := range rows {
			if i >= len(row) {
				continue
			}
			value := strings.TrimSpace(row[i].Value)
			if value == "" {
				continue
			}
			cellWidth := lipgloss.Width(value)
			if cellWidth > maxWidth {
				maxWidth = cellWidth
			}
		}
		if maxWidth < spec.Min {
			maxWidth = spec.Min
		}
		if spec.Max > 0 && maxWidth > spec.Max {
			maxWidth = spec.Max
		}
		widths[i] = maxWidth
	}
	total := sumInts(widths)
	if total == available {
		return widths
	}
	if total < available {
		extra := available - total
		flex := flexColumns(specs)
		if len(flex) == 0 {
			flex = allColumns(specs)
		}
		distribute(widths, specs, flex, extra, true)
		return widths
	}
	deficit := total - available
	flex := flexColumns(specs)
	if len(flex) == 0 {
		flex = allColumns(specs)
	}
	distribute(widths, specs, flex, deficit, false)
	return widths
}

func distribute(
	widths []int,
	specs []columnSpec,
	indices []int,
	amount int,
	grow bool,
) {
	if amount <= 0 || len(indices) == 0 {
		return
	}
	for amount > 0 {
		changed := false
		for _, idx := range indices {
			if idx >= len(widths) {
				continue
			}
			if grow {
				if specs[idx].Max > 0 && widths[idx] >= specs[idx].Max {
					continue
				}
				widths[idx]++
			} else {
				if widths[idx] <= specs[idx].Min {
					continue
				}
				widths[idx]--
			}
			amount--
			changed = true
			if amount == 0 {
				break
			}
		}
		if !changed {
			return
		}
	}
}

func flexColumns(specs []columnSpec) []int {
	indices := make([]int, 0, len(specs))
	for i, spec := range specs {
		if spec.Flex {
			indices = append(indices, i)
		}
	}
	return indices
}

func allColumns(specs []columnSpec) []int {
	indices := make([]int, len(specs))
	for i := range specs {
		indices[i] = i
	}
	return indices
}

func sumInts(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func safeWidth(widths []int, idx int) int {
	if idx < 0 || idx >= len(widths) {
		return 1
	}
	if widths[idx] < 1 {
		return 1
	}
	return widths[idx]
}

func (m *Model) headerBox(content string) string {
	return m.styles.HeaderBox.Render(content)
}

// houseArt renders a retro pixel-art house for the expanded profile.
// Uses shade characters (░▒▓█) for a classic DOS/BBS aesthetic.
// Returns "" if the terminal is too narrow.
func (m *Model) houseArt() string {
	if m.effectiveWidth() < 80 {
		return ""
	}
	rf := lipgloss.NewStyle().Foreground(accent)    // roof
	wl := lipgloss.NewStyle().Foreground(textMid)   // walls
	wn := lipgloss.NewStyle().Foreground(warning)   // windows (lit)
	dr := lipgloss.NewStyle().Foreground(secondary) // door
	lines := []string{
		rf.Render("      ▄▓▄"),
		rf.Render("    ▄▓▓▓▓▓▄"),
		rf.Render("  ▄▓▓▓▓▓▓▓▓▓▄"),
		wl.Render("  ██ ") + wn.Render("░░") + wl.Render(" ") + wn.Render("░░") + wl.Render(" ██"),
		wl.Render("  ██  ") + dr.Render("████") + wl.Render(" ██"),
		wl.Render("  ██  ") + dr.Render("█  █") + wl.Render(" ██"),
		wl.Render("  ▀▀▀▀▀▀▀▀▀▀▀"),
	}
	return strings.Join(lines, "\n")
}

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
	return m.styles.Keycap.Render(strings.ToUpper(value))
}

// styledPart returns a styled value, or "" if the underlying value is blank.
func styledPart(style lipgloss.Style, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return style.Render(value)
}

// joinStyledParts joins pre-styled parts with a separator, skipping empty ones.
func joinStyledParts(sep string, parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	return strings.Join(filtered, sep)
}

// hlv renders a dim label followed by a bright value, or "" if the value is blank.
func (m *Model) hlv(label, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return m.styles.HeaderLabel.Render(label) + " " + m.styles.HeaderValue.Render(value)
}

// houseSection renders a section header with values, indenting continuation lines.
func (m *Model) houseSection(header string, lines ...string) string {
	label := m.styles.HeaderSection.Render(header)
	labelWidth := lipgloss.Width(label)
	indent := strings.Repeat(" ", labelWidth+1)
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			filtered = append(filtered, line)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	result := make([]string, len(filtered))
	for i, line := range filtered {
		if i == 0 {
			result[i] = label + " " + line
		} else {
			result[i] = indent + line
		}
	}
	return strings.Join(result, "\n")
}

func bedBathLabel(profile data.HouseProfile) string {
	var parts []string
	if profile.Bedrooms > 0 {
		parts = append(parts, fmt.Sprintf("%dbd", profile.Bedrooms))
	}
	if profile.Bathrooms > 0 {
		parts = append(parts, fmt.Sprintf("%sba", formatFloat(profile.Bathrooms)))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " / ")
}

func sqftLabel(sqft int) string {
	if sqft == 0 {
		return ""
	}
	return fmt.Sprintf("%d ft\u00B2", sqft)
}

func lotLabel(sqft int) string {
	if sqft == 0 {
		return ""
	}
	return fmt.Sprintf("%d ft\u00B2 lot", sqft)
}

func formatInt(value int) string {
	if value == 0 {
		return ""
	}
	return fmt.Sprintf("%d", value)
}

func formatFloat(value float64) string {
	if value == 0 {
		return ""
	}
	if value == math.Trunc(value) {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.1f", value)
}

func formatCityState(profile data.HouseProfile) string {
	parts := []string{
		strings.TrimSpace(profile.City),
		strings.TrimSpace(profile.State),
	}
	return joinNonEmpty(parts, ", ")
}

func formatAddress(profile data.HouseProfile) string {
	parts := []string{
		strings.TrimSpace(profile.AddressLine1),
		strings.TrimSpace(profile.AddressLine2),
		strings.TrimSpace(profile.City),
		strings.TrimSpace(profile.State),
		strings.TrimSpace(profile.PostalCode),
	}
	return joinNonEmpty(parts, ", ")
}

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

func hoaSummary(profile data.HouseProfile) string {
	if profile.HOAName == "" && profile.HOAFeeCents == nil {
		return ""
	}
	fee := data.FormatOptionalCents(profile.HOAFeeCents)
	if profile.HOAName == "" {
		return fee
	}
	if fee == "" {
		return profile.HOAName
	}
	return fmt.Sprintf("%s (%s)", profile.HOAName, fee)
}
