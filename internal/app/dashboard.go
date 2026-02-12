// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/cpcloud/micasa/internal/data"
)

// ---------------------------------------------------------------------------
// Dashboard header
// ---------------------------------------------------------------------------

func (m *Model) dashboardHeader(width int) string {
	var parts []string
	if m.hasHouse && m.house.Nickname != "" {
		parts = append(parts, m.house.Nickname)
	}
	parts = append(parts, time.Now().Format("Monday, Jan 2"))

	text := m.styles.DashSubtitle.Render(strings.Join(parts, " \u00b7 "))
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(text)
}

// ---------------------------------------------------------------------------
// Dashboard data types
// ---------------------------------------------------------------------------

// dashboardData holds pre-computed dashboard content. Loaded fresh each time
// the dashboard is opened or returned to.
type dashboardData struct {
	Overdue            []maintenanceUrgency
	Upcoming           []maintenanceUrgency
	ActiveProjects     []data.Project
	ExpiringWarranties []warrantyStatus
	InsuranceRenewal   *insuranceStatus
	ServiceSpendCents  int64
	ProjectSpendCents  int64
}

type maintenanceUrgency struct {
	Item          data.MaintenanceItem
	NextDue       time.Time
	DaysFromNow   int // negative = overdue, positive = upcoming
	ApplianceName string
}

type warrantyStatus struct {
	Appliance   data.Appliance
	DaysFromNow int // negative = recently expired, positive = expiring soon
}

type insuranceStatus struct {
	Carrier     string
	RenewalDate time.Time
	DaysFromNow int
}

// dashNavEntry maps a dashboard cursor position to a target row in a tab.
type dashNavEntry struct {
	Tab TabKind
	ID  uint
}

// ---------------------------------------------------------------------------
// Mini-table rendering (aligned columns per dashboard section)
// ---------------------------------------------------------------------------

// dashCell is one cell in a dashboard mini-table row.
type dashCell struct {
	Text  string         // raw (unstyled) text for width measurement
	Style lipgloss.Style // applied when rendering
	Align alignKind
}

// dashRow is one navigable row in a section.
type dashRow struct {
	Cells  []dashCell
	Target *dashNavEntry // nil = not navigable (e.g. summary lines)
}

// renderMiniTable renders rows with aligned columns and returns the rendered
// lines. colGap is the space between columns.
func renderMiniTable(rows []dashRow, colGap int, cursor int, selected lipgloss.Style) []string {
	if len(rows) == 0 {
		return nil
	}
	// Compute max width per column.
	nCols := 0
	for _, r := range rows {
		if len(r.Cells) > nCols {
			nCols = len(r.Cells)
		}
	}
	widths := make([]int, nCols)
	for _, r := range rows {
		for i, c := range r.Cells {
			if w := lipgloss.Width(c.Text); w > widths[i] {
				widths[i] = w
			}
		}
	}

	gap := strings.Repeat(" ", colGap)
	lines := make([]string, 0, len(rows))
	for rowIdx, r := range rows {
		parts := make([]string, len(r.Cells))
		for i, c := range r.Cells {
			styled := c.Style.Render(c.Text)
			textWidth := lipgloss.Width(c.Text)
			pad := widths[i] - textWidth
			if pad < 0 {
				pad = 0
			}
			if c.Align == alignRight {
				parts[i] = strings.Repeat(" ", pad) + styled
			} else {
				parts[i] = styled + strings.Repeat(" ", pad)
			}
		}
		line := "  " + strings.Join(parts, gap)
		if rowIdx == cursor {
			line = selected.Render(line)
		}
		lines = append(lines, line)
	}
	return lines
}

// ---------------------------------------------------------------------------
// Data loading
// ---------------------------------------------------------------------------

func (m *Model) loadDashboard() error {
	return m.loadDashboardAt(time.Now())
}

func (m *Model) loadDashboardAt(now time.Time) error {
	if m.store == nil {
		return nil
	}
	var d dashboardData

	// Maintenance urgency.
	items, err := m.store.ListMaintenanceWithSchedule()
	if err != nil {
		return fmt.Errorf("load maintenance: %w", err)
	}
	for _, item := range items {
		nextDue := data.ComputeNextDue(item.LastServicedAt, item.IntervalMonths)
		if nextDue == nil {
			continue
		}
		days := daysUntil(now, *nextDue)
		appName := ""
		if item.ApplianceID != nil && item.Appliance.Name != "" {
			appName = item.Appliance.Name
		}
		entry := maintenanceUrgency{
			Item:          item,
			NextDue:       *nextDue,
			DaysFromNow:   days,
			ApplianceName: appName,
		}
		if days < 0 {
			d.Overdue = append(d.Overdue, entry)
		} else if days <= 30 {
			d.Upcoming = append(d.Upcoming, entry)
		}
	}
	sortByDays(d.Overdue)
	sortByDays(d.Upcoming)
	d.Overdue = capSlice(d.Overdue, 10)
	d.Upcoming = capSlice(d.Upcoming, 10-len(d.Overdue))

	// Active projects.
	d.ActiveProjects, err = m.store.ListActiveProjects()
	if err != nil {
		return fmt.Errorf("load active projects: %w", err)
	}

	// Expiring warranties (expired within 30 days or expiring within 90).
	appliances, err := m.store.ListExpiringWarranties(
		now, 30*24*time.Hour, 90*24*time.Hour,
	)
	if err != nil {
		return fmt.Errorf("load warranties: %w", err)
	}
	for _, a := range appliances {
		if a.WarrantyExpiry == nil {
			continue
		}
		days := daysUntil(now, *a.WarrantyExpiry)
		d.ExpiringWarranties = append(d.ExpiringWarranties, warrantyStatus{
			Appliance:   a,
			DaysFromNow: days,
		})
	}

	// Insurance renewal.
	if m.hasHouse && m.house.InsuranceRenewal != nil {
		days := daysUntil(now, *m.house.InsuranceRenewal)
		if days >= -30 && days <= 90 {
			d.InsuranceRenewal = &insuranceStatus{
				Carrier:     m.house.InsuranceCarrier,
				RenewalDate: *m.house.InsuranceRenewal,
				DaysFromNow: days,
			}
		}
	}

	// Spending snapshot (YTD).
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	d.ServiceSpendCents, err = m.store.YTDServiceSpendCents(yearStart)
	if err != nil {
		return fmt.Errorf("load service spend: %w", err)
	}
	d.ProjectSpendCents, err = m.store.YTDProjectSpendCents()
	if err != nil {
		return fmt.Errorf("load project spend: %w", err)
	}

	m.dashboard = d
	m.buildDashNav()
	return nil
}

// ---------------------------------------------------------------------------
// Navigation index
// ---------------------------------------------------------------------------

// buildDashNav builds the flat navigation list from dashboard data. Each
// navigable item maps cursor position -> (tab, id).
func (m *Model) buildDashNav() {
	var nav []dashNavEntry
	for _, e := range m.dashboard.Overdue {
		nav = append(nav, dashNavEntry{Tab: tabMaintenance, ID: e.Item.ID})
	}
	for _, e := range m.dashboard.Upcoming {
		nav = append(nav, dashNavEntry{Tab: tabMaintenance, ID: e.Item.ID})
	}
	for _, p := range m.dashboard.ActiveProjects {
		nav = append(nav, dashNavEntry{Tab: tabProjects, ID: p.ID})
	}
	for _, w := range m.dashboard.ExpiringWarranties {
		nav = append(nav, dashNavEntry{Tab: tabAppliances, ID: w.Appliance.ID})
	}
	// Insurance renewal is not navigable (it's in the house profile).
	m.dashNav = nav
	if m.dashCursor >= len(nav) {
		m.dashCursor = max(0, len(nav)-1)
	}
}

func (m *Model) dashNavCount() int {
	return len(m.dashNav)
}

// ---------------------------------------------------------------------------
// Dashboard view (main entry)
// ---------------------------------------------------------------------------

// dashSection holds one dashboard section's data for height budgeting.
type dashSection struct {
	title string
	rows  []dashRow
	// Maintenance uses sub-sections (Overdue / Upcoming), adding extra header
	// lines. Other sections have subCounts == nil.
	subTitles []string
	subCounts []int // rows per sub-section (sum == len(rows))
}

// overhead returns the number of non-data lines this section occupies
// (headers, blanks between sub-sections).
func (s dashSection) overhead() int {
	if len(s.subCounts) <= 1 {
		return 1 // just the section header
	}
	// N sub-headers + (N-1) blanks between them.
	n := 0
	for _, c := range s.subCounts {
		if c > 0 {
			n++
		}
	}
	if n <= 1 {
		return 1
	}
	return n + (n - 1) // sub-headers + separator blanks
}

// dashboardView renders the dashboard content, fitting within budget content
// lines. Empty sections are skipped; rows are trimmed proportionally when the
// terminal is short.
func (m *Model) dashboardView(budget int) string {
	d := m.dashboard

	// Collect non-empty sections.
	var sections []dashSection

	// Maintenance: split into overdue / upcoming sub-sections.
	if nO, nU := len(d.Overdue), len(d.Upcoming); nO+nU > 0 {
		allRows := m.dashMaintRows()
		var subTitles []string
		var subCounts []int
		if nO > 0 {
			subTitles = append(subTitles, "Overdue")
			subCounts = append(subCounts, nO)
		}
		if nU > 0 {
			subTitles = append(subTitles, "Upcoming")
			subCounts = append(subCounts, nU)
		}
		sections = append(sections, dashSection{
			title:     "Maintenance",
			rows:      allRows,
			subTitles: subTitles,
			subCounts: subCounts,
		})
	}

	if projRows := m.dashProjectRows(); len(projRows) > 0 {
		sections = append(sections, dashSection{
			title: "Active Projects", rows: projRows,
		})
	}

	if expRows := m.dashExpiringRows(); len(expRows) > 0 {
		sections = append(sections, dashSection{
			title: "Expiring Soon", rows: expRows,
		})
	}

	spendLine := m.dashSpendingLine()

	if len(sections) == 0 && spendLine == "" {
		return ""
	}

	// Compute fixed overhead: section headers, sub-headers, inter-section
	// blanks, and the spending footer.
	fixedLines := 0
	for i, s := range sections {
		fixedLines += s.overhead()
		if i > 0 {
			fixedLines++ // blank between sections
		}
	}
	if spendLine != "" {
		if len(sections) > 0 {
			fixedLines++ // blank before spending
		}
		fixedLines += 2 // header + data
	}

	// Distribute remaining budget among data rows.
	totalRaw := 0
	for _, s := range sections {
		totalRaw += len(s.rows)
	}
	avail := budget - fixedLines
	if avail < len(sections) {
		avail = len(sections) // at least 1 row per section
	}

	limits := distributeDashRows(sections, avail)

	// Trim section rows and sub-counts to limits, then render.
	sel := m.styles.TableSelected
	colGap := 3
	cursor := m.dashCursor
	var nav []dashNavEntry
	var allLines []string

	for i, s := range sections {
		limit := limits[i]
		rows := capSlice(s.rows, limit)

		var sLines []string
		if len(s.subCounts) > 0 {
			sLines = m.renderMaintSection(s, limit, cursor, colGap, sel)
		} else {
			hdr := m.styles.DashSection.Render(s.title)
			localCursor := -1
			if cursor >= 0 && cursor < len(rows) {
				localCursor = cursor
			}
			sLines = append([]string{hdr},
				renderMiniTable(rows, colGap, localCursor, sel)...)
		}
		cursor -= len(rows)

		// Collect nav entries from the rendered rows.
		for _, r := range rows {
			if r.Target != nil {
				nav = append(nav, *r.Target)
			}
		}

		allLines = append(allLines, strings.Join(sLines, "\n"))
	}

	if spendLine != "" {
		header := m.styles.DashSection.Render("Spending (YTD)")
		allLines = append(allLines, header+"\n  "+spendLine)
	}

	// Update nav to match the trimmed view.
	m.dashNav = nav
	if m.dashCursor >= len(nav) {
		m.dashCursor = max(0, len(nav)-1)
	}

	return strings.Join(allLines, "\n\n")
}

// renderMaintSection renders the maintenance section with Overdue/Upcoming
// sub-headers, distributing limit rows across sub-sections proportionally.
func (m *Model) renderMaintSection(
	s dashSection, limit, cursor, colGap int, sel lipgloss.Style,
) []string {
	// Distribute limit among sub-sections proportionally.
	subLimits := distributeSubLimits(s.subCounts, limit)

	var lines []string
	offset := 0
	rendered := 0
	for si, subTitle := range s.subTitles {
		subN := subLimits[si]
		if subN == 0 {
			offset += s.subCounts[si]
			continue
		}
		if rendered > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, m.styles.DashSection.Render(subTitle))
		subRows := capSlice(s.rows[offset:offset+s.subCounts[si]], subN)
		localCursor := -1
		if cursor >= 0 && cursor < subN {
			localCursor = cursor
		}
		lines = append(lines,
			renderMiniTable(subRows, colGap, localCursor, sel)...)
		cursor -= subN
		offset += s.subCounts[si]
		rendered++
	}
	return lines
}

// distributeDashRows allocates avail row slots across sections proportionally,
// giving each section at least 1 row. Returns per-section limits.
func distributeDashRows(sections []dashSection, avail int) []int {
	n := len(sections)
	if n == 0 {
		return nil
	}
	limits := make([]int, n)
	totalRaw := 0
	for _, s := range sections {
		totalRaw += len(s.rows)
	}

	if avail >= totalRaw {
		// Everything fits.
		for i, s := range sections {
			limits[i] = len(s.rows)
		}
		return limits
	}

	// Give each section at least 1 row.
	remaining := avail
	for i := range sections {
		limits[i] = 1
		remaining--
	}
	if remaining < 0 {
		remaining = 0
	}

	// Distribute the rest proportionally to each section's raw count.
	excess := totalRaw - n // total "extra" rows beyond the 1 minimum
	if excess > 0 && remaining > 0 {
		for i, s := range sections {
			extra := len(s.rows) - 1
			share := extra * remaining / excess
			limits[i] += share
		}
	}

	// Rounding may leave us short; give leftover to largest sections first.
	allocated := 0
	for _, l := range limits {
		allocated += l
	}
	for allocated < avail {
		best := -1
		bestGap := 0
		for i, s := range sections {
			gap := len(s.rows) - limits[i]
			if gap > bestGap {
				bestGap = gap
				best = i
			}
		}
		if best < 0 {
			break
		}
		limits[best]++
		allocated++
	}

	return limits
}

// distributeSubLimits splits limit rows across sub-sections proportionally.
func distributeSubLimits(subCounts []int, limit int) []int {
	n := len(subCounts)
	if n == 0 {
		return nil
	}
	total := 0
	for _, c := range subCounts {
		total += c
	}
	result := make([]int, n)
	if limit >= total {
		copy(result, subCounts)
		return result
	}
	// Give each non-empty sub-section at least 1.
	remaining := limit
	for i, c := range subCounts {
		if c > 0 && remaining > 0 {
			result[i] = 1
			remaining--
		}
	}
	if remaining > 0 && total > n {
		for i, c := range subCounts {
			extra := c - 1
			share := extra * remaining / (total - n)
			result[i] += share
		}
	}
	// Distribute rounding leftover.
	allocated := 0
	for _, r := range result {
		allocated += r
	}
	for allocated < limit {
		best := -1
		bestGap := 0
		for i, c := range subCounts {
			gap := c - result[i]
			if gap > bestGap {
				bestGap = gap
				best = i
			}
		}
		if best < 0 {
			break
		}
		result[best]++
		allocated++
	}
	return result
}

// ---------------------------------------------------------------------------
// Row builders (produce dashRow slices for mini-table rendering)
// ---------------------------------------------------------------------------

func (m *Model) dashMaintRows() []dashRow {
	d := m.dashboard
	all := make([]maintenanceUrgency, 0, len(d.Overdue)+len(d.Upcoming))
	all = append(all, d.Overdue...)
	all = append(all, d.Upcoming...)
	if len(all) == 0 {
		return nil
	}
	rows := make([]dashRow, 0, len(all))
	for _, e := range all {
		overdue := e.DaysFromNow < 0
		nameStyle := m.styles.DashUpcoming
		if overdue {
			nameStyle = m.styles.DashOverdue
		}
		appliance := ""
		if e.ApplianceName != "" {
			appliance = e.ApplianceName
		}
		lastSrv := ""
		if e.Item.LastServicedAt != nil {
			lastSrv = e.Item.LastServicedAt.Format(data.DateLayout)
		}
		rows = append(rows, dashRow{
			Cells: []dashCell{
				{Text: e.Item.Name, Style: nameStyle},
				{Text: appliance, Style: m.styles.DashLabel},
				{
					Text:  daysText(e.DaysFromNow, overdue),
					Style: daysStyle(e.DaysFromNow, overdue),
					Align: alignRight,
				},
				{Text: lastSrv, Style: m.styles.DashLabel, Align: alignRight},
			},
			Target: &dashNavEntry{Tab: tabMaintenance, ID: e.Item.ID},
		})
	}
	return rows
}

func (m *Model) dashProjectRows() []dashRow {
	d := m.dashboard
	rows := make([]dashRow, 0, len(d.ActiveProjects))
	for _, p := range d.ActiveProjects {
		statusStyle := m.styles.StatusStyles[p.Status]
		budgetText := ""
		budgetStyle := m.styles.Money
		if p.BudgetCents != nil {
			act := data.FormatOptionalCents(p.ActualCents)
			bud := data.FormatCents(*p.BudgetCents)
			if act != "" {
				budgetText = act + " / " + bud
				if p.ActualCents != nil && *p.ActualCents > *p.BudgetCents {
					budgetStyle = m.styles.DashOverdue
				}
			} else {
				budgetText = bud
			}
		}
		rows = append(rows, dashRow{
			Cells: []dashCell{
				{Text: p.Title, Style: m.styles.DashValue},
				{Text: p.Status, Style: statusStyle},
				{Text: budgetText, Style: budgetStyle, Align: alignRight},
			},
			Target: &dashNavEntry{Tab: tabProjects, ID: p.ID},
		})
	}
	return rows
}

func (m *Model) dashExpiringRows() []dashRow {
	d := m.dashboard
	var rows []dashRow
	for _, w := range d.ExpiringWarranties {
		overdue := w.DaysFromNow < 0
		nameStyle := m.styles.DashUpcoming
		if overdue {
			nameStyle = m.styles.DashOverdue
		}
		rows = append(rows, dashRow{
			Cells: []dashCell{
				{Text: w.Appliance.Name + " warranty", Style: nameStyle},
				{
					Text:  w.Appliance.WarrantyExpiry.Format(data.DateLayout),
					Style: m.styles.DashLabel,
					Align: alignRight,
				},
				{
					Text:  daysText(w.DaysFromNow, overdue),
					Style: daysStyle(w.DaysFromNow, overdue),
					Align: alignRight,
				},
			},
			Target: &dashNavEntry{Tab: tabAppliances, ID: w.Appliance.ID},
		})
	}
	// Insurance renewal is not navigable (no tab row to jump to).
	if d.InsuranceRenewal != nil {
		ins := d.InsuranceRenewal
		overdue := ins.DaysFromNow < 0
		nameStyle := m.styles.DashUpcoming
		if overdue {
			nameStyle = m.styles.DashOverdue
		}
		label := "Insurance renewal"
		if ins.Carrier != "" {
			label += " (" + ins.Carrier + ")"
		}
		rows = append(rows, dashRow{
			Cells: []dashCell{
				{Text: label, Style: nameStyle},
				{
					Text:  ins.RenewalDate.Format(data.DateLayout),
					Style: m.styles.DashLabel,
					Align: alignRight,
				},
				{
					Text:  daysText(ins.DaysFromNow, overdue),
					Style: daysStyle(ins.DaysFromNow, overdue),
					Align: alignRight,
				},
			},
			// Not navigable: no Target.
		})
	}
	return rows
}

func (m *Model) dashSpendingLine() string {
	d := m.dashboard
	total := d.ServiceSpendCents + d.ProjectSpendCents
	if total == 0 {
		return ""
	}
	var parts []string
	if d.ServiceSpendCents > 0 {
		parts = append(parts,
			m.styles.DashLabel.Render("Maintenance ")+
				m.styles.Money.Render(data.FormatCents(d.ServiceSpendCents)))
	}
	if d.ProjectSpendCents > 0 {
		parts = append(parts,
			m.styles.DashLabel.Render("Projects ")+
				m.styles.Money.Render(data.FormatCents(d.ProjectSpendCents)))
	}
	sep := m.styles.DashLabel.Render(" \u00b7 ")
	line := strings.Join(parts, sep)
	if d.ServiceSpendCents > 0 && d.ProjectSpendCents > 0 {
		line += sep + m.styles.DashLabel.Render("Total ") +
			m.styles.Money.Render(data.FormatCents(total))
	}
	return line
}

// ---------------------------------------------------------------------------
// Dashboard keyboard navigation
// ---------------------------------------------------------------------------

func (m *Model) dashDown() {
	n := m.dashNavCount()
	if n == 0 {
		return
	}
	m.dashCursor++
	if m.dashCursor >= n {
		m.dashCursor = n - 1
	}
}

func (m *Model) dashUp() {
	m.dashCursor--
	if m.dashCursor < 0 {
		m.dashCursor = 0
	}
}

func (m *Model) dashTop() {
	m.dashCursor = 0
}

func (m *Model) dashBottom() {
	n := m.dashNavCount()
	if n == 0 {
		return
	}
	m.dashCursor = n - 1
}

func (m *Model) dashJump() {
	nav := m.dashNav
	if m.dashCursor < 0 || m.dashCursor >= len(nav) {
		return
	}
	entry := nav[m.dashCursor]
	m.showDashboard = false
	m.active = tabIndex(entry.Tab)
	tab := m.activeTab()
	if tab != nil && tab.Stale {
		_ = m.reloadIfStale(tab)
	} else {
		_ = m.reloadActiveTab()
	}
	if tab != nil {
		selectRowByID(tab, entry.ID)
	}
}

// ---------------------------------------------------------------------------
// Utility helpers
// ---------------------------------------------------------------------------

// daysText returns the human-readable timing string without ANSI styling.
func daysText(days int, overdue bool) string {
	if days == 0 {
		return "today"
	}
	abs := days
	if abs < 0 {
		abs = -abs
	}
	unit := "days"
	if abs == 1 {
		unit = "day"
	}
	if overdue {
		return fmt.Sprintf("%d %s overdue", abs, unit)
	}
	return fmt.Sprintf("in %d %s", abs, unit)
}

// daysStyle returns the appropriate style for a timing label.
func daysStyle(days int, overdue bool) lipgloss.Style {
	if days == 0 || overdue {
		return lipgloss.NewStyle().Foreground(danger).Bold(true)
	}
	return lipgloss.NewStyle().Foreground(warning)
}

// daysUntil returns the number of whole days from now to target. Negative
// means target is in the past.
func daysUntil(now, target time.Time) int {
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	tgtDate := time.Date(
		target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, time.UTC,
	)
	return int(math.Round(tgtDate.Sub(nowDate).Hours() / 24))
}

func sortByDays(items []maintenanceUrgency) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].DaysFromNow < items[j-1].DaysFromNow; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}

func capSlice[T any](s []T, maxLen int) []T {
	if maxLen < 0 {
		maxLen = 0
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
