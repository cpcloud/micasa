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
	house := m.houseView()
	tabs := m.tabsView()
	content := ""
	if m.mode == modeForm && m.form != nil {
		content = m.form.View()
	} else if tab := m.activeTab(); tab != nil {
		content = m.tableView(tab)
	}
	logPane := m.logView()
	if logPane != "" {
		logPane = joinVerticalNonEmpty(
			m.logDivider(),
			logPane,
		)
	}
	status := m.statusView()
	return joinVerticalNonEmpty(house, tabs, content, logPane, status)
}

func (m *Model) houseView() string {
	if !m.hasHouse {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			m.houseTitleLine("setup"),
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
	line1 := m.houseTitleLine("▸")
	line2 := joinInline(
		m.chip("House", m.house.Nickname),
		m.chip("Loc", formatCityState(m.house)),
		m.chip("Yr", formatInt(m.house.YearBuilt)),
		m.chip("Sq Ft", formatInt(m.house.SquareFeet)),
		m.chip("Beds", formatInt(m.house.Bedrooms)),
		m.chip("Baths", formatFloat(m.house.Bathrooms)),
	)
	return joinVerticalNonEmpty(line1, line2)
}

func (m *Model) houseExpanded() string {
	address := formatAddress(m.house)
	line1 := m.houseTitleLine("▾")
	line2 := joinInline(
		m.chip("House", m.house.Nickname),
		m.chip("Addr", address),
	)
	line3 := m.sectionLine(
		"Structure",
		m.chip("Yr", formatInt(m.house.YearBuilt)),
		m.chip("Sq Ft", formatInt(m.house.SquareFeet)),
		m.chip("Lot", formatInt(m.house.LotSquareFeet)),
		m.chip("Beds", formatInt(m.house.Bedrooms)),
		m.chip("Baths", formatFloat(m.house.Bathrooms)),
		m.chip("Fnd", m.house.FoundationType),
		m.chip("Wir", m.house.WiringType),
		m.chip("Roof", m.house.RoofType),
		m.chip("Ext", m.house.ExteriorType),
		m.chip("Base", m.house.BasementType),
	)
	line4 := m.sectionLine(
		"Utilities",
		m.chip("Heat", m.house.HeatingType),
		m.chip("Cool", m.house.CoolingType),
		m.chip("Water", m.house.WaterSource),
		m.chip("Sewer", m.house.SewerType),
		m.chip("Park", m.house.ParkingType),
	)
	line5 := m.sectionLine(
		"Financial",
		m.chip("Ins", m.house.InsuranceCarrier),
		m.chip("Policy", m.house.InsurancePolicy),
		m.chip("Renew", data.FormatDate(m.house.InsuranceRenewal)),
		m.chip("Tax", data.FormatOptionalCents(m.house.PropertyTaxCents)),
		m.chip("HOA", hoaSummary(m.house)),
	)
	return joinVerticalNonEmpty(line1, line2, line3, line4, line5)
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

func (m *Model) statusView() string {
	if m.mode == modeForm {
		help := joinWithSeparator(
			m.helpSeparator(),
			m.helpItem("esc", "cancel"),
			m.helpItem("ctrl+c", "quit"),
		)
		if m.status.Text == "" {
			return help
		}
		style := m.styles.Info
		if m.status.Kind == statusError {
			style = m.styles.Error
		}
		return lipgloss.JoinVertical(
			lipgloss.Left,
			style.Render(m.status.Text),
			help,
		)
	}
	deleted := "deleted:off"
	tab := m.activeTab()
	if tab != nil && tab.ShowDeleted {
		deleted = "deleted:on"
	}
	help := joinWithSeparator(
		m.helpSeparator(),
		m.helpItem("tab/shift+tab", "switch"),
		m.helpItem("a", "add"),
		m.helpItem("d", "delete"),
		m.helpItem("u", "restore"),
		m.helpItem("x", "deleted"),
		m.helpItem("p", "profile"),
		m.helpItem("h", "house"),
		m.helpItem("q", "quit"),
	)
	if m.log.enabled {
		help = joinWithSeparator(m.helpSeparator(), help, m.helpItem("l", "log"))
	}
	deletedLabel := m.styles.HeaderHint.Render(deleted)
	helpLine := joinWithSeparator(
		m.helpSeparator(),
		help,
		deletedLabel,
	)
	if m.status.Text == "" {
		return helpLine
	}
	style := m.styles.Info
	if m.status.Kind == statusError {
		style = m.styles.Error
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.Render(m.status.Text),
		helpLine,
	)
}

func (m *Model) logView() string {
	if !m.log.enabled {
		return ""
	}
	title := m.styles.LogTitle.Render("Logs")
	level := m.styles.HeaderHint.Render(fmt.Sprintf("v%d", m.log.verbosity))
	focus := m.styles.HeaderHint.Render("blur")
	if m.log.focus {
		focus = m.styles.LogFocus.Render("focus")
	}
	header := joinInline(title, level, focus)

	filterStatus := m.log.validityLabel()
	statusStyle := m.styles.LogValid
	if m.log.filterErr != nil {
		statusStyle = m.styles.LogInvalid
	}
	filterLine := fmt.Sprintf(
		"%s /%s/ %s",
		m.styles.HeaderHint.Render("filter"),
		m.log.input.View(),
		statusStyle.Render(filterStatus),
	)

	width := m.width
	if width <= 0 {
		width = 80
	}
	header = ansi.Truncate(header, width, "…")
	filterLine = ansi.Truncate(filterLine, width, "…")
	bodyLines := m.logLines() - 2
	if bodyLines < 1 {
		bodyLines = 1
	}
	entries := m.filteredLogEntries()
	if len(entries) > bodyLines {
		entries = entries[len(entries)-bodyLines:]
	}
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		line := m.formatLogEntry(entry, width)
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		lines = []string{m.styles.Empty.Render("No log entries.")}
	}
	return joinVerticalNonEmpty(header, filterLine, strings.Join(lines, "\n"))
}

func (m *Model) logDivider() string {
	width := m.width
	if width <= 0 {
		width = 80
	}
	line := strings.Repeat("─", width)
	return m.styles.TableSeparator.Render(line)
}

func (m *Model) filteredLogEntries() []logEntry {
	entries := make([]logEntry, 0, len(m.log.entries))
	for _, entry := range m.log.entries {
		raw := fmt.Sprintf(
			"%s %s %s",
			entry.Time.Format("15:04:05"),
			entry.Level.String(),
			entry.Message,
		)
		if !m.log.matches(raw) {
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}

func (m *Model) formatLogEntry(entry logEntry, width int) string {
	levelStyle := m.styles.LogLevelInfo
	switch entry.Level {
	case logError:
		levelStyle = m.styles.LogLevelError
	case logDebug:
		levelStyle = m.styles.LogLevelDebug
	}
	level := levelStyle.Render(entry.Level.String())
	raw := fmt.Sprintf("%s %s %s", entry.Time.Format("15:04:05"), level, entry.Message)
	return ansi.Truncate(raw, width, "…")
}

func (m *Model) tableView(tab *Tab) string {
	if tab == nil || len(tab.Specs) == 0 {
		return ""
	}
	width := m.width
	if width <= 0 {
		width = 80
	}
	separator := m.styles.TableSeparator.Render(" │ ")
	dividerSep := m.styles.TableSeparator.Render("─┼─")
	widths := columnWidths(tab.Specs, tab.CellRows, width, lipgloss.Width(separator))
	header := renderHeaderRow(tab.Specs, widths, separator, m.styles.TableHeader)
	divider := renderDivider(widths, dividerSep, m.styles.TableSeparator)
	rows := renderRows(
		tab.Specs,
		tab.CellRows,
		tab.Rows,
		widths,
		separator,
		tab.Table.Cursor(),
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
	style lipgloss.Style,
) string {
	cells := make([]string, 0, len(specs))
	for i, spec := range specs {
		width := safeWidth(widths, i)
		text := formatCell(spec.Title, width, spec.Align)
		cells = append(cells, style.Render(text))
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
		row := renderRow(specs, rows[i], widths, separator, selected, deleted, styles)
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
	styles Styles,
) string {
	cells := make([]string, 0, len(specs))
	for i, spec := range specs {
		width := safeWidth(widths, i)
		var cellValue cell
		if i < len(row) {
			cellValue = row[i]
		}
		cells = append(cells, renderCell(cellValue, spec, width, styles))
	}
	rendered := strings.Join(cells, separator)
	if deleted {
		rendered = styles.Deleted.Render(rendered)
	}
	if selected {
		rendered = styles.TableSelected.Render(rendered)
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

func (m *Model) helpItem(keys, label string) string {
	keycaps := m.renderKeys(keys)
	desc := m.styles.HeaderHint.Render(label)
	return strings.TrimSpace(fmt.Sprintf("%s %s", keycaps, desc))
}

func (m *Model) helpSeparator() string {
	return m.styles.HeaderHint.Render(" · ")
}

func (m *Model) renderKeys(keys string) string {
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

func (m *Model) houseTitleLine(state string) string {
	title := m.styles.HeaderTitle.Render("House Profile")
	badge := ""
	if strings.TrimSpace(state) != "" {
		badge = m.styles.HeaderBadge.Render(state)
	}
	hint := m.styles.HeaderHint.Render("h toggle")
	return joinInline(title, badge, hint)
}

func (m *Model) chip(label, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	labelText := m.styles.HeaderLabel.Render(label)
	valueText := m.renderHouseValue(value)
	return m.styles.HeaderChip.Render(fmt.Sprintf("%s %s", labelText, valueText))
}

func (m *Model) sectionLine(label string, chips ...string) string {
	section := m.styles.HeaderSection.Render(label)
	parts := make([]string, 0, len(chips)+1)
	for _, chip := range chips {
		if strings.TrimSpace(chip) != "" {
			parts = append(parts, chip)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	parts = append([]string{section}, parts...)
	return joinInline(parts...)
}

func (m *Model) renderHouseValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return m.styles.Empty.Render("n/a")
	}
	return m.styles.HeaderValue.Render(value)
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

func joinInline(values ...string) string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filtered = append(filtered, value)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, filtered...)
}

func joinVerticalNonEmpty(values ...string) string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filtered = append(filtered, value)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	return lipgloss.JoinVertical(lipgloss.Left, filtered...)
}

func joinWithSeparator(sep string, values ...string) string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filtered = append(filtered, value)
		}
	}
	return strings.Join(filtered, sep)
}

func joinNonEmpty(values []string, sep string) string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filtered = append(filtered, value)
		}
	}
	return strings.Join(filtered, sep)
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
