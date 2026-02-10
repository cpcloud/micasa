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
	RecentActivity     []data.ServiceLogEntry
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
			if w := len(c.Text); w > widths[i] {
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
			textWidth := len(c.Text)
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

	// Recent activity.
	d.RecentActivity, err = m.store.ListRecentServiceLogs(5)
	if err != nil {
		return fmt.Errorf("load recent activity: %w", err)
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
	for _, e := range m.dashboard.RecentActivity {
		nav = append(nav, dashNavEntry{Tab: tabMaintenance, ID: e.MaintenanceItemID})
	}
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

func (m *Model) dashboardView() string {
	sel := m.styles.TableSelected
	d := m.dashboard
	cursor := m.dashCursor // running cursor into the flat nav list
	colGap := 3

	type sectionResult struct {
		lines []string
		count int // navigable rows consumed from the cursor
	}

	renderSection := func(
		header, emptyMsg string,
		rows []dashRow,
	) sectionResult {
		hdr := m.styles.DashSection.Render(header)
		if len(rows) == 0 {
			msg := m.styles.DashAllClear.Render("  " + emptyMsg)
			return sectionResult{lines: []string{hdr, msg}}
		}
		localCursor := -1
		if cursor >= 0 && cursor < len(rows) {
			localCursor = cursor
		}
		rendered := renderMiniTable(rows, colGap, localCursor, sel)
		cursor -= len(rows)
		return sectionResult{
			lines: append([]string{hdr}, rendered...),
			count: len(rows),
		}
	}

	var allLines []string

	// Maintenance (overdue + upcoming as one navigable block).
	maintRows := m.dashMaintRows()
	if len(maintRows) > 0 {
		// Split into overdue/upcoming sub-headers.
		nOverdue := len(d.Overdue)
		nUpcoming := len(d.Upcoming)
		var mLines []string

		if nOverdue > 0 {
			subhdr := m.styles.DashSection.Render("Overdue")
			mLines = append(mLines, subhdr)
			localCursor := -1
			if cursor >= 0 && cursor < nOverdue {
				localCursor = cursor
			}
			mLines = append(
				mLines,
				renderMiniTable(maintRows[:nOverdue], colGap, localCursor, sel)...)
		}
		if nUpcoming > 0 {
			if nOverdue > 0 {
				mLines = append(mLines, "")
			}
			subhdr := m.styles.DashSection.Render("Upcoming")
			mLines = append(mLines, subhdr)
			localCursor := -1
			upIdx := cursor - nOverdue
			if upIdx >= 0 && upIdx < nUpcoming {
				localCursor = upIdx
			}
			mLines = append(
				mLines,
				renderMiniTable(maintRows[nOverdue:], colGap, localCursor, sel)...)
		}
		if nOverdue == 0 && nUpcoming == 0 {
			mLines = append(mLines,
				m.styles.DashSection.Render("Maintenance"),
				m.styles.DashAllClear.Render(
					"  Nothing overdue or upcoming -- nice work!"),
			)
		}
		cursor -= len(maintRows)
		allLines = append(allLines, strings.Join(mLines, "\n"))
	} else {
		allLines = append(allLines, strings.Join([]string{
			m.styles.DashSection.Render("Maintenance"),
			m.styles.DashAllClear.Render(
				"  Nothing overdue or upcoming -- nice work!"),
		}, "\n"))
	}

	// Projects.
	projResult := renderSection(
		"Active Projects",
		"No active projects. Time to start something?",
		m.dashProjectRows(),
	)
	allLines = append(allLines, strings.Join(projResult.lines, "\n"))

	// Expiring.
	expRows := m.dashExpiringRows()
	expResult := renderSection(
		"Expiring Soon",
		"All clear for the next 90 days.",
		expRows,
	)
	allLines = append(allLines, strings.Join(expResult.lines, "\n"))

	// Recent activity.
	actResult := renderSection(
		"Recent Activity",
		"No service history yet.",
		m.dashActivityRows(),
	)
	allLines = append(allLines, strings.Join(actResult.lines, "\n"))

	// Spending (not navigable, no cursor tracking needed).
	if spend := m.dashSpendingLine(); spend != "" {
		header := m.styles.DashSection.Render("Spending (YTD)")
		allLines = append(allLines, header+"\n  "+spend)
	}

	return strings.Join(allLines, "\n\n")
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

func (m *Model) dashActivityRows() []dashRow {
	d := m.dashboard
	rows := make([]dashRow, 0, len(d.RecentActivity))
	for _, e := range d.RecentActivity {
		who := "Self"
		if e.VendorID != nil && e.Vendor.Name != "" {
			who = e.Vendor.Name
		}
		costText := ""
		if e.CostCents != nil {
			costText = data.FormatCents(*e.CostCents)
		}
		rows = append(rows, dashRow{
			Cells: []dashCell{
				{Text: e.ServicedAt.Format(data.DateLayout), Style: m.styles.DashLabel},
				{Text: e.MaintenanceItem.Name, Style: m.styles.DashValue},
				{Text: who, Style: m.styles.DashLabel},
				{Text: costText, Style: m.styles.Money, Align: alignRight},
			},
			Target: &dashNavEntry{Tab: tabMaintenance, ID: e.MaintenanceItemID},
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
	_ = m.reloadActiveTab()
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
