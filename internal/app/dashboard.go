// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
// lines. colGap is the space between columns. maxWidth caps the total line
// width; the first column is truncated with an ellipsis when rows would
// otherwise wrap. Pass 0 to disable width capping.
func renderMiniTable(
	rows []dashRow, colGap, maxWidth, cursor int, selected lipgloss.Style,
) []string {
	if len(rows) == 0 {
		return nil
	}

	// Compute max display width per column.
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

	// If total width exceeds maxWidth, shrink the first column.
	const indent = 2
	if maxWidth > 0 && nCols > 0 {
		total := indent
		for i, w := range widths {
			total += w
			if i > 0 {
				total += colGap
			}
		}
		if overflow := total - maxWidth; overflow > 0 {
			minFirst := 6
			newFirst := widths[0] - overflow
			if newFirst < minFirst {
				newFirst = minFirst
			}
			widths[0] = newFirst
		}
	}

	gap := strings.Repeat(" ", colGap)
	lines := make([]string, 0, len(rows))
	for rowIdx, r := range rows {
		parts := make([]string, len(r.Cells))
		for i, c := range r.Cells {
			text := c.Text
			// Truncate text that exceeds its column width.
			if tw := lipgloss.Width(text); tw > widths[i] {
				text = truncateToWidth(text, widths[i])
			}
			styled := c.Style.Render(text)
			tw := lipgloss.Width(text)
			pad := widths[i] - tw
			if pad < 0 {
				pad = 0
			}
			if c.Align == alignRight {
				parts[i] = strings.Repeat(" ", pad) + styled
			} else {
				parts[i] = styled + strings.Repeat(" ", pad)
			}
		}
		prefix := "  "
		if rowIdx == cursor {
			prefix = "\u25b8 "
		}
		line := prefix + strings.Join(parts, gap)
		if rowIdx == cursor {
			line = selected.Render(line)
		}
		lines = append(lines, line)
	}
	return lines
}

// truncateToWidth trims text to fit within maxW display columns, appending
// an ellipsis if truncation occurs. Delegates to ansi.Truncate for correct
// grapheme-cluster and wide-character handling.
func truncateToWidth(text string, maxW int) string {
	return ansi.Truncate(text, maxW, "\u2026")
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
// lines and maxWidth display columns. Empty sections are skipped; rows are
// trimmed proportionally when the terminal is short. maxWidth prevents rows
// from exceeding the overlay's inner width (which would cause wrapping and
// eat vertical space).
func (m *Model) dashboardView(budget, maxWidth int) string {
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
		fixedLines += 3 // rule + header + data
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
			sLines = m.renderMaintSection(s, limit, cursor, colGap, maxWidth, sel)
		} else {
			hdr := m.styles.DashSection.Render(s.title)
			localCursor := -1
			if cursor >= 0 && cursor < len(rows) {
				localCursor = cursor
			}
			sLines = append([]string{hdr},
				renderMiniTable(rows, colGap, maxWidth, localCursor, sel)...)
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
		rule := m.styles.DashRule.Render(strings.Repeat("─", maxWidth))
		header := m.styles.DashSection.Render("Spending (YTD)")
		allLines = append(allLines, rule+"\n"+header+"\n  "+spendLine)
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
	s dashSection, limit, cursor, colGap, maxWidth int, sel lipgloss.Style,
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
		sectionStyle := m.styles.DashSection
		if subTitle == "Overdue" {
			sectionStyle = m.styles.DashSectionWarn
		}
		lines = append(lines, sectionStyle.Render(subTitle))
		subRows := capSlice(s.rows[offset:offset+s.subCounts[si]], subN)
		localCursor := -1
		if cursor >= 0 && cursor < subN {
			localCursor = cursor
		}
		lines = append(lines,
			renderMiniTable(subRows, colGap, maxWidth, localCursor, sel)...)
		cursor -= subN
		offset += s.subCounts[si]
		rendered++
	}
	return lines
}

// distributeProportional allocates avail slots across buckets proportionally,
// giving each non-empty bucket at least 1. Used for both section-level and
// sub-section-level row budgeting.
func distributeProportional(counts []int, avail int) []int {
	n := len(counts)
	if n == 0 {
		return nil
	}
	result := make([]int, n)
	total := 0
	for _, c := range counts {
		total += c
	}
	if avail >= total {
		copy(result, counts)
		return result
	}

	// Give each non-empty bucket at least 1.
	nonEmpty := 0
	remaining := avail
	for i, c := range counts {
		if c > 0 && remaining > 0 {
			result[i] = 1
			remaining--
			nonEmpty++
		}
	}

	// Distribute the rest proportionally to each bucket's raw count.
	excess := total - nonEmpty
	if excess > 0 && remaining > 0 {
		for i, c := range counts {
			extra := c - 1
			if extra <= 0 {
				continue
			}
			share := extra * remaining / excess
			result[i] += share
		}
	}

	// Rounding may leave us short; give leftover to largest buckets first.
	allocated := 0
	for _, r := range result {
		allocated += r
	}
	for allocated < avail {
		best := -1
		bestGap := 0
		for i, c := range counts {
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

// distributeDashRows allocates avail row slots across sections proportionally.
func distributeDashRows(sections []dashSection, avail int) []int {
	counts := make([]int, len(sections))
	for i, s := range sections {
		counts[i] = len(s.rows)
	}
	return distributeProportional(counts, avail)
}

// distributeSubLimits splits limit rows across sub-sections proportionally.
func distributeSubLimits(subCounts []int, limit int) []int {
	return distributeProportional(subCounts, limit)
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
					Style: m.daysStyle(e.DaysFromNow, overdue),
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
		statusText := statusLabel(p.Status)
		budgetText := ""
		budgetStyle := m.styles.Money
		if p.BudgetCents != nil {
			fmtCents := data.FormatCompactCents
			fmtOpt := data.FormatCompactOptionalCents
			if m.magMode {
				fmtCents = magCents
				fmtOpt = magOptionalCents
			}
			act := fmtOpt(p.ActualCents)
			bud := fmtCents(*p.BudgetCents)
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
				{Text: statusText, Style: statusStyle},
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
					Style: m.daysStyle(w.DaysFromNow, overdue),
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
					Style: m.daysStyle(ins.DaysFromNow, overdue),
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
	fmtCents := data.FormatCompactCents
	if m.magMode {
		fmtCents = magCents
	}
	var parts []string
	if d.ServiceSpendCents > 0 {
		parts = append(parts,
			m.styles.DashLabel.Render("Maintenance ")+
				m.styles.Money.Render(fmtCents(d.ServiceSpendCents)))
	}
	if d.ProjectSpendCents > 0 {
		parts = append(parts,
			m.styles.DashLabel.Render("Projects ")+
				m.styles.Money.Render(fmtCents(d.ProjectSpendCents)))
	}
	sep := m.styles.DashLabel.Render(" \u00b7 ")
	line := strings.Join(parts, sep)
	if d.ServiceSpendCents > 0 && d.ProjectSpendCents > 0 {
		line += sep + m.styles.DashLabel.Render("Total ") +
			m.styles.Money.Render(fmtCents(total))
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
	m.switchToTab(tabIndex(entry.Tab))
	if tab := m.activeTab(); tab != nil {
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

// daysStyle returns the appropriate style for a timing label, using the
// Styles struct to stay consistent with the colorblind-safe palette.
func (m *Model) daysStyle(days int, overdue bool) lipgloss.Style {
	if days == 0 || overdue {
		return m.styles.DashOverdue
	}
	return m.styles.DashUpcoming
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
	slices.SortFunc(items, func(a, b maintenanceUrgency) int {
		return cmp.Compare(a.DaysFromNow, b.DaysFromNow)
	})
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
