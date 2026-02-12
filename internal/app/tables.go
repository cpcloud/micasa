// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/cpcloud/micasa/internal/data"
)

// baseTableKeyMap returns the default table KeyMap with b/f removed from
// page-up/page-down so those keys can be used for tab navigation.
func baseTableKeyMap() table.KeyMap {
	km := table.DefaultKeyMap()
	km.PageDown.SetKeys("pgdown")
	km.PageDown.SetHelp("pgdn", "page down")
	km.PageUp.SetKeys("pgup")
	km.PageUp.SetHelp("pgup", "page up")
	return km
}

// normalTableKeyMap returns the table KeyMap for normal (nav) mode.
func normalTableKeyMap() table.KeyMap {
	return baseTableKeyMap()
}

// editTableKeyMap returns a table KeyMap with d/u stripped from half-page
// bindings so they can be used for delete/undo without conflicting.
func editTableKeyMap() table.KeyMap {
	km := baseTableKeyMap()
	km.HalfPageDown.SetKeys("ctrl+d")
	km.HalfPageDown.SetHelp("ctrl+d", "½ page down")
	km.HalfPageUp.SetKeys("ctrl+u")
	km.HalfPageUp.SetHelp("ctrl+u", "½ page up")
	return km
}

// setAllTableKeyMaps applies a KeyMap to every tab's table.
func (m *Model) setAllTableKeyMaps(km table.KeyMap) {
	for i := range m.tabs {
		m.tabs[i].Table.KeyMap = km
	}
	if dc := m.detail(); dc != nil {
		dc.Tab.Table.KeyMap = km
	}
}

func NewTabs(styles Styles) []Tab {
	projectSpecs := projectColumnSpecs()
	quoteSpecs := quoteColumnSpecs()
	maintenanceSpecs := maintenanceColumnSpecs()
	applianceSpecs := applianceColumnSpecs()
	vendorSpecs := vendorColumnSpecs()
	documentSpecs := documentColumnSpecs()
	return []Tab{
		{
			Kind:    tabProjects,
			Name:    "Projects",
			Handler: projectHandler{},
			Specs:   projectSpecs,
			Table:   newTable(specsToColumns(projectSpecs), styles),
		},
		{
			Kind:    tabQuotes,
			Name:    tabQuotes.String(),
			Handler: quoteHandler{},
			Specs:   quoteSpecs,
			Table:   newTable(specsToColumns(quoteSpecs), styles),
		},
		{
			Kind:    tabMaintenance,
			Name:    "Maintenance",
			Handler: maintenanceHandler{},
			Specs:   maintenanceSpecs,
			Table:   newTable(specsToColumns(maintenanceSpecs), styles),
		},
		{
			Kind:    tabAppliances,
			Name:    "Appliances",
			Handler: applianceHandler{},
			Specs:   applianceSpecs,
			Table:   newTable(specsToColumns(applianceSpecs), styles),
		},
		{
			Kind:    tabVendors,
			Name:    "Vendors",
			Handler: vendorHandler{},
			Specs:   vendorSpecs,
			Table:   newTable(specsToColumns(vendorSpecs), styles),
		},
		{
			Kind:    tabDocuments,
			Name:    "Documents",
			Handler: documentHandler{},
			Specs:   documentSpecs,
			Table:   newTable(specsToColumns(documentSpecs), styles),
		},
	}
}

func projectColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{Title: "Type", Min: 8, Max: 14, Flex: true},
		{Title: "Title", Min: 14, Max: 32, Flex: true},
		{Title: "Status", Min: 6, Max: 8, Kind: cellStatus},
		{Title: "Budget", Min: 10, Max: 14, Align: alignRight, Kind: cellMoney},
		{Title: "Actual", Min: 10, Max: 14, Align: alignRight, Kind: cellMoney},
		{Title: "Start", Min: 10, Max: 12, Kind: cellDate},
		{Title: "End", Min: 10, Max: 12, Kind: cellDate},
		{Title: tabQuotes.String(), Min: 6, Max: 8, Align: alignRight, Kind: cellDrilldown},
	}
}

func quoteColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{
			Title: "Project",
			Min:   12,
			Max:   24,
			Flex:  true,
			Link:  &columnLink{TargetTab: tabProjects},
		},
		{
			Title: "Vendor",
			Min:   12,
			Max:   20,
			Flex:  true,
			Link:  &columnLink{TargetTab: tabVendors},
		},
		{Title: "Total", Min: 10, Max: 14, Align: alignRight, Kind: cellMoney},
		{Title: "Labor", Min: 10, Max: 14, Align: alignRight, Kind: cellMoney},
		{Title: "Mat", Min: 8, Max: 12, Align: alignRight, Kind: cellMoney},
		{Title: "Other", Min: 8, Max: 12, Align: alignRight, Kind: cellMoney},
		{Title: "Recv", Min: 10, Max: 12, Kind: cellDate},
	}
}

func maintenanceColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{Title: "Item", Min: 12, Max: 26, Flex: true},
		{Title: "Category", Min: 10, Max: 14},
		{
			Title: "Appliance",
			Min:   10,
			Max:   18,
			Flex:  true,
			Link:  &columnLink{TargetTab: tabAppliances},
		},
		{Title: "Last", Min: 10, Max: 12, Kind: cellDate},
		{Title: "Next", Min: 10, Max: 12, Kind: cellUrgency},
		{Title: "Every", Min: 6, Max: 10},
		{Title: "Log", Min: 4, Max: 6, Align: alignRight, Kind: cellDrilldown},
	}
}

func applianceColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{Title: "Name", Min: 12, Max: 24, Flex: true},
		{Title: "Brand", Min: 8, Max: 16, Flex: true},
		{Title: "Model", Min: 8, Max: 16},
		{Title: "Serial", Min: 8, Max: 14},
		{Title: "Location", Min: 8, Max: 14},
		{Title: "Purchased", Min: 10, Max: 12, Kind: cellDate},
		{Title: "Age", Min: 5, Max: 8, Kind: cellReadonly},
		{Title: "Warranty", Min: 10, Max: 12, Kind: cellWarranty},
		{Title: "Cost", Min: 8, Max: 12, Align: alignRight, Kind: cellMoney},
		{Title: "Maint", Min: 5, Max: 6, Align: alignRight, Kind: cellDrilldown},
	}
}

// applianceMaintenanceColumnSpecs is like maintenanceColumnSpecs but without
// the Appliance column (already scoped by the parent appliance).
func applianceMaintenanceColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{Title: "Item", Min: 12, Max: 26, Flex: true},
		{Title: "Category", Min: 10, Max: 14},
		{Title: "Last", Min: 10, Max: 12, Kind: cellDate},
		{Title: "Next", Min: 10, Max: 12, Kind: cellUrgency},
		{Title: "Every", Min: 6, Max: 10},
		{Title: "Log", Min: 4, Max: 6, Align: alignRight, Kind: cellDrilldown},
	}
}

func applianceMaintenanceRows(
	items []data.MaintenanceItem,
	logCounts map[uint]int,
) ([]table.Row, []rowMeta, [][]cell) {
	return buildRows(items, func(item data.MaintenanceItem) rowSpec {
		interval := formatInterval(item.IntervalMonths)
		logCount := ""
		if n := logCounts[item.ID]; n > 0 {
			logCount = fmt.Sprintf("%d", n)
		}
		nextDue := data.ComputeNextDue(item.LastServicedAt, item.IntervalMonths)
		return rowSpec{
			ID:      item.ID,
			Deleted: item.DeletedAt.Valid,
			Cells: []cell{
				{Value: fmt.Sprintf("%d", item.ID), Kind: cellReadonly},
				{Value: item.Name, Kind: cellText},
				{Value: item.Category.Name, Kind: cellText},
				{Value: dateValue(item.LastServicedAt), Kind: cellDate},
				{Value: dateValue(nextDue), Kind: cellUrgency},
				{Value: interval, Kind: cellText},
				{Value: logCount, Kind: cellDrilldown},
			},
		}
	})
}

func serviceLogColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{Title: "Date", Min: 10, Max: 12, Kind: cellDate},
		{
			Title: "Performed By",
			Min:   12,
			Max:   22,
			Flex:  true,
			Link:  &columnLink{TargetTab: tabVendors},
		},
		{Title: "Cost", Min: 8, Max: 12, Align: alignRight, Kind: cellMoney},
		{Title: "Notes", Min: 12, Max: 40, Flex: true, Kind: cellNotes},
	}
}

func serviceLogRows(
	entries []data.ServiceLogEntry,
) ([]table.Row, []rowMeta, [][]cell) {
	return buildRows(entries, func(e data.ServiceLogEntry) rowSpec {
		performedBy := "Self"
		var vendorLinkID uint
		if e.VendorID != nil && e.Vendor.Name != "" {
			performedBy = e.Vendor.Name
			vendorLinkID = *e.VendorID
		}
		return rowSpec{
			ID:      e.ID,
			Deleted: e.DeletedAt.Valid,
			Cells: []cell{
				{Value: fmt.Sprintf("%d", e.ID), Kind: cellReadonly},
				{Value: e.ServicedAt.Format(data.DateLayout), Kind: cellDate},
				{Value: performedBy, Kind: cellText, LinkID: vendorLinkID},
				{Value: centsValue(e.CostCents), Kind: cellMoney},
				{Value: e.Notes, Kind: cellNotes},
			},
		}
	})
}

func applianceRows(
	items []data.Appliance,
	maintCounts map[uint]int,
	now time.Time,
) ([]table.Row, []rowMeta, [][]cell) {
	return buildRows(items, func(a data.Appliance) rowSpec {
		maintCount := ""
		if n := maintCounts[a.ID]; n > 0 {
			maintCount = fmt.Sprintf("%d", n)
		}
		return rowSpec{
			ID:      a.ID,
			Deleted: a.DeletedAt.Valid,
			Cells: []cell{
				{Value: fmt.Sprintf("%d", a.ID), Kind: cellReadonly},
				{Value: a.Name, Kind: cellText},
				{Value: a.Brand, Kind: cellText},
				{Value: a.ModelNumber, Kind: cellText},
				{Value: a.SerialNumber, Kind: cellText},
				{Value: a.Location, Kind: cellText},
				{Value: dateValue(a.PurchaseDate), Kind: cellDate},
				{Value: applianceAge(a.PurchaseDate, now), Kind: cellReadonly},
				{Value: dateValue(a.WarrantyExpiry), Kind: cellWarranty},
				{Value: centsValue(a.CostCents), Kind: cellMoney},
				{Value: maintCount, Kind: cellDrilldown},
			},
		}
	})
}

// formatInterval returns a compact interval string: "3m", "1y", "2y 6m".
// Returns empty for non-positive values.
func formatInterval(months int) string {
	if months <= 0 {
		return ""
	}
	y := months / 12
	m := months % 12
	if y == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dy", y)
	}
	return fmt.Sprintf("%dy %dm", y, m)
}

// applianceAge returns a human-readable age string from purchase date to now.
func applianceAge(purchased *time.Time, now time.Time) string {
	if purchased == nil {
		return ""
	}
	years := now.Year() - purchased.Year()
	months := int(now.Month()) - int(purchased.Month())
	if now.Day() < purchased.Day() {
		months--
	}
	if months < 0 {
		years--
		months += 12
	}
	if years < 0 {
		return ""
	}
	if years == 0 {
		if months <= 0 {
			return "<1m"
		}
		return fmt.Sprintf("%dm", months)
	}
	if months == 0 {
		return fmt.Sprintf("%dy", years)
	}
	return fmt.Sprintf("%dy %dm", years, months)
}

func vendorColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{Title: "Name", Min: 14, Max: 24, Flex: true},
		{Title: "Contact", Min: 10, Max: 20, Flex: true},
		{Title: "Email", Min: 12, Max: 24, Flex: true},
		{Title: "Phone", Min: 12, Max: 16},
		{Title: "Website", Min: 12, Max: 28, Flex: true},
		{Title: tabQuotes.String(), Min: 6, Max: 8, Align: alignRight, Kind: cellDrilldown},
		{Title: "Jobs", Min: 5, Max: 8, Align: alignRight, Kind: cellDrilldown},
	}
}

func documentColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{Title: "Title", Min: 14, Max: 24, Flex: true},
		{Title: "File", Min: 12, Max: 24, Flex: true},
		{Title: "Type", Min: 10, Max: 14},
		{Title: "Ref", Min: 8, Max: 16},
		{Title: "Size", Min: 8, Max: 12, Align: alignRight, Kind: cellReadonly},
	}
}

func vendorRows(
	vendors []data.Vendor,
	quoteCounts map[uint]int,
	jobCounts map[uint]int,
) ([]table.Row, []rowMeta, [][]cell) {
	return buildRows(vendors, func(v data.Vendor) rowSpec {
		quoteCount := ""
		if n := quoteCounts[v.ID]; n > 0 {
			quoteCount = fmt.Sprintf("%d", n)
		}
		jobCount := ""
		if n := jobCounts[v.ID]; n > 0 {
			jobCount = fmt.Sprintf("%d", n)
		}
		return rowSpec{
			ID:      v.ID,
			Deleted: v.DeletedAt.Valid,
			Cells: []cell{
				{Value: fmt.Sprintf("%d", v.ID), Kind: cellReadonly},
				{Value: v.Name, Kind: cellText},
				{Value: v.ContactName, Kind: cellText},
				{Value: v.Email, Kind: cellText},
				{Value: v.Phone, Kind: cellText},
				{Value: v.Website, Kind: cellText},
				{Value: quoteCount, Kind: cellDrilldown},
				{Value: jobCount, Kind: cellDrilldown},
			},
		}
	})
}

func documentRows(items []data.Document) ([]table.Row, []rowMeta, [][]cell) {
	return buildRows(items, func(d data.Document) rowSpec {
		ref := ""
		if d.EntityID != nil {
			ref = fmt.Sprintf("%s:%d", d.EntityKind, *d.EntityID)
		}
		return rowSpec{
			ID:      d.ID,
			Deleted: d.DeletedAt.Valid,
			Cells: []cell{
				{Value: fmt.Sprintf("%d", d.ID), Kind: cellReadonly},
				{Value: d.Title, Kind: cellText},
				{Value: d.FileName, Kind: cellText},
				{Value: displayDocumentKind(d.EntityKind), Kind: cellText},
				{Value: ref, Kind: cellText},
				{Value: formatSize(d.SizeBytes), Kind: cellReadonly},
			},
		}
	})
}

func displayDocumentKind(kind string) string {
	if kind == "" {
		return "none"
	}
	return kind
}

func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	kb := float64(size) / 1024.0
	if kb < 1024 {
		return fmt.Sprintf("%.1f KB", kb)
	}
	mb := kb / 1024.0
	return fmt.Sprintf("%.1f MB", mb)
}

func specsToColumns(specs []columnSpec) []table.Column {
	cols := make([]table.Column, 0, len(specs))
	for _, spec := range specs {
		width := spec.Min
		if width <= 0 {
			width = 6
		}
		cols = append(cols, table.Column{Title: spec.Title, Width: width})
	}
	return cols
}

func newTable(columns []table.Column, styles Styles) table.Model {
	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)
	tbl.SetStyles(table.Styles{
		Header:   styles.TableHeader,
		Selected: styles.TableSelected,
	})
	return tbl
}

func projectRows(
	projects []data.Project,
	quoteCounts map[uint]int,
) ([]table.Row, []rowMeta, [][]cell) {
	return buildRows(projects, func(p data.Project) rowSpec {
		quoteCount := ""
		if n := quoteCounts[p.ID]; n > 0 {
			quoteCount = fmt.Sprintf("%d", n)
		}
		return rowSpec{
			ID:      p.ID,
			Deleted: p.DeletedAt.Valid,
			Cells: []cell{
				{Value: fmt.Sprintf("%d", p.ID), Kind: cellReadonly},
				{Value: p.ProjectType.Name, Kind: cellText},
				{Value: p.Title, Kind: cellText},
				{Value: p.Status, Kind: cellStatus},
				{Value: centsValue(p.BudgetCents), Kind: cellMoney},
				{Value: centsValue(p.ActualCents), Kind: cellMoney},
				{Value: dateValue(p.StartDate), Kind: cellDate},
				{Value: dateValue(p.EndDate), Kind: cellDate},
				{Value: quoteCount, Kind: cellDrilldown},
			},
		}
	})
}

func quoteRows(
	quotes []data.Quote,
) ([]table.Row, []rowMeta, [][]cell) {
	return buildRows(quotes, func(q data.Quote) rowSpec {
		projectName := q.Project.Title
		if projectName == "" {
			projectName = fmt.Sprintf("Project %d", q.ProjectID)
		}
		return rowSpec{
			ID:      q.ID,
			Deleted: q.DeletedAt.Valid,
			Cells: []cell{
				{Value: fmt.Sprintf("%d", q.ID), Kind: cellReadonly},
				{Value: projectName, Kind: cellText, LinkID: q.ProjectID},
				{Value: q.Vendor.Name, Kind: cellText, LinkID: q.VendorID},
				{Value: data.FormatCents(q.TotalCents), Kind: cellMoney},
				{Value: centsValue(q.LaborCents), Kind: cellMoney},
				{Value: centsValue(q.MaterialsCents), Kind: cellMoney},
				{Value: centsValue(q.OtherCents), Kind: cellMoney},
				{Value: dateValue(q.ReceivedDate), Kind: cellDate},
			},
		}
	})
}

func maintenanceRows(
	items []data.MaintenanceItem,
	logCounts map[uint]int,
) ([]table.Row, []rowMeta, [][]cell) {
	return buildRows(items, func(item data.MaintenanceItem) rowSpec {
		interval := formatInterval(item.IntervalMonths)
		appName := ""
		var appLinkID uint
		if item.ApplianceID != nil {
			appName = item.Appliance.Name
			appLinkID = *item.ApplianceID
		}
		logCount := ""
		if n := logCounts[item.ID]; n > 0 {
			logCount = fmt.Sprintf("%d", n)
		}
		nextDue := data.ComputeNextDue(item.LastServicedAt, item.IntervalMonths)
		return rowSpec{
			ID:      item.ID,
			Deleted: item.DeletedAt.Valid,
			Cells: []cell{
				{Value: fmt.Sprintf("%d", item.ID), Kind: cellReadonly},
				{Value: item.Name, Kind: cellText},
				{Value: item.Category.Name, Kind: cellText},
				{Value: appName, Kind: cellText, LinkID: appLinkID},
				{Value: dateValue(item.LastServicedAt), Kind: cellDate},
				{Value: dateValue(nextDue), Kind: cellUrgency},
				{Value: interval, Kind: cellText},
				{Value: logCount, Kind: cellDrilldown},
			},
		}
	})
}

func cellsToRow(cells []cell) table.Row {
	row := make(table.Row, len(cells))
	for i, cell := range cells {
		row[i] = cell.Value
	}
	return row
}

// rowSpec describes one table row from an entity.
type rowSpec struct {
	ID      uint
	Deleted bool
	Cells   []cell
}

// buildRows converts a slice of entities into the three parallel slices that
// the table and sort systems consume. The toRow function maps each entity to
// its ID, deletion status, and cell values.
func buildRows[T any](items []T, toRow func(T) rowSpec) ([]table.Row, []rowMeta, [][]cell) {
	rows := make([]table.Row, 0, len(items))
	meta := make([]rowMeta, 0, len(items))
	cells := make([][]cell, 0, len(items))
	for _, item := range items {
		spec := toRow(item)
		rows = append(rows, cellsToRow(spec.Cells))
		cells = append(cells, spec.Cells)
		meta = append(meta, rowMeta{ID: spec.ID, Deleted: spec.Deleted})
	}
	return rows, meta, cells
}

// vendorQuoteColumnSpecs defines the columns for quotes scoped to a vendor.
// Omits the Vendor column since the parent context provides that.
func vendorQuoteColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{
			Title: "Project",
			Min:   12,
			Max:   24,
			Flex:  true,
			Link:  &columnLink{TargetTab: tabProjects},
		},
		{Title: "Total", Min: 10, Max: 14, Align: alignRight, Kind: cellMoney},
		{Title: "Labor", Min: 10, Max: 14, Align: alignRight, Kind: cellMoney},
		{Title: "Mat", Min: 8, Max: 12, Align: alignRight, Kind: cellMoney},
		{Title: "Other", Min: 8, Max: 12, Align: alignRight, Kind: cellMoney},
		{Title: "Recv", Min: 10, Max: 12, Kind: cellDate},
	}
}

func vendorQuoteRows(
	quotes []data.Quote,
) ([]table.Row, []rowMeta, [][]cell) {
	return buildRows(quotes, func(q data.Quote) rowSpec {
		projectName := q.Project.Title
		if projectName == "" {
			projectName = fmt.Sprintf("Project %d", q.ProjectID)
		}
		return rowSpec{
			ID:      q.ID,
			Deleted: q.DeletedAt.Valid,
			Cells: []cell{
				{Value: fmt.Sprintf("%d", q.ID), Kind: cellReadonly},
				{Value: projectName, Kind: cellText, LinkID: q.ProjectID},
				{Value: data.FormatCents(q.TotalCents), Kind: cellMoney},
				{Value: centsValue(q.LaborCents), Kind: cellMoney},
				{Value: centsValue(q.MaterialsCents), Kind: cellMoney},
				{Value: centsValue(q.OtherCents), Kind: cellMoney},
				{Value: dateValue(q.ReceivedDate), Kind: cellDate},
			},
		}
	})
}

// vendorJobsColumnSpecs defines the columns for service log entries scoped to
// a vendor. Omits the Vendor column since the parent context provides that.
func vendorJobsColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{
			Title: "Item",
			Min:   12,
			Max:   24,
			Flex:  true,
			Link:  &columnLink{TargetTab: tabMaintenance},
		},
		{Title: "Date", Min: 10, Max: 12, Kind: cellDate},
		{Title: "Cost", Min: 8, Max: 12, Align: alignRight, Kind: cellMoney},
		{Title: "Notes", Min: 12, Max: 40, Flex: true, Kind: cellNotes},
	}
}

func vendorJobsRows(
	entries []data.ServiceLogEntry,
) ([]table.Row, []rowMeta, [][]cell) {
	return buildRows(entries, func(e data.ServiceLogEntry) rowSpec {
		itemName := e.MaintenanceItem.Name
		return rowSpec{
			ID:      e.ID,
			Deleted: e.DeletedAt.Valid,
			Cells: []cell{
				{Value: fmt.Sprintf("%d", e.ID), Kind: cellReadonly},
				{Value: itemName, Kind: cellText, LinkID: e.MaintenanceItemID},
				{Value: e.ServicedAt.Format(data.DateLayout), Kind: cellDate},
				{Value: centsValue(e.CostCents), Kind: cellMoney},
				{Value: e.Notes, Kind: cellNotes},
			},
		}
	})
}

// projectQuoteColumnSpecs defines the columns for quotes scoped to a project.
// Omits the Project column since the parent context provides that.
func projectQuoteColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{
			Title: "Vendor",
			Min:   12,
			Max:   20,
			Flex:  true,
			Link:  &columnLink{TargetTab: tabVendors},
		},
		{Title: "Total", Min: 10, Max: 14, Align: alignRight, Kind: cellMoney},
		{Title: "Labor", Min: 10, Max: 14, Align: alignRight, Kind: cellMoney},
		{Title: "Mat", Min: 8, Max: 12, Align: alignRight, Kind: cellMoney},
		{Title: "Other", Min: 8, Max: 12, Align: alignRight, Kind: cellMoney},
		{Title: "Recv", Min: 10, Max: 12, Kind: cellDate},
	}
}

func projectQuoteRows(
	quotes []data.Quote,
) ([]table.Row, []rowMeta, [][]cell) {
	return buildRows(quotes, func(q data.Quote) rowSpec {
		return rowSpec{
			ID:      q.ID,
			Deleted: q.DeletedAt.Valid,
			Cells: []cell{
				{Value: fmt.Sprintf("%d", q.ID), Kind: cellReadonly},
				{Value: q.Vendor.Name, Kind: cellText, LinkID: q.VendorID},
				{Value: data.FormatCents(q.TotalCents), Kind: cellMoney},
				{Value: centsValue(q.LaborCents), Kind: cellMoney},
				{Value: centsValue(q.MaterialsCents), Kind: cellMoney},
				{Value: centsValue(q.OtherCents), Kind: cellMoney},
				{Value: dateValue(q.ReceivedDate), Kind: cellDate},
			},
		}
	})
}

func centsValue(cents *int64) string {
	if cents == nil {
		return ""
	}
	return data.FormatCents(*cents)
}

func dateValue(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(data.DateLayout)
}
