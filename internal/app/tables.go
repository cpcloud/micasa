package app

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/micasa/micasa/internal/data"
)

// normalTableKeyMap returns the default table KeyMap with full vim bindings.
func normalTableKeyMap() table.KeyMap {
	return table.DefaultKeyMap()
}

// editTableKeyMap returns a table KeyMap with d/u stripped from half-page
// bindings so they can be used for delete/undo without conflicting.
func editTableKeyMap() table.KeyMap {
	km := table.DefaultKeyMap()
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
}

func NewTabs(styles Styles) []Tab {
	projectSpecs := projectColumnSpecs()
	quoteSpecs := quoteColumnSpecs()
	maintenanceSpecs := maintenanceColumnSpecs()
	applianceSpecs := applianceColumnSpecs()
	return []Tab{
		{
			Kind:  tabProjects,
			Name:  "Projects",
			Specs: projectSpecs,
			Table: newTable(specsToColumns(projectSpecs), styles),
		},
		{
			Kind:  tabQuotes,
			Name:  "Quotes",
			Specs: quoteSpecs,
			Table: newTable(specsToColumns(quoteSpecs), styles),
		},
		{
			Kind:  tabMaintenance,
			Name:  "Maintenance",
			Specs: maintenanceSpecs,
			Table: newTable(specsToColumns(maintenanceSpecs), styles),
		},
		{
			Kind:  tabAppliances,
			Name:  "Appliances",
			Specs: applianceSpecs,
			Table: newTable(specsToColumns(applianceSpecs), styles),
		},
	}
}

func projectColumnSpecs() []columnSpec {
	return []columnSpec{
		{Title: "ID", Min: 4, Max: 6, Align: alignRight, Kind: cellReadonly},
		{Title: "Type", Min: 8, Max: 14, Flex: true},
		{Title: "Title", Min: 14, Max: 32, Flex: true},
		{Title: "Status", Min: 8, Max: 12, Kind: cellStatus},
		{Title: "Budget", Min: 10, Max: 14, Align: alignRight, Kind: cellMoney},
		{Title: "Actual", Min: 10, Max: 14, Align: alignRight, Kind: cellMoney},
		{Title: "Start", Min: 10, Max: 12, Kind: cellDate},
		{Title: "End", Min: 10, Max: 12, Kind: cellDate},
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
			Link:  &columnLink{TargetTab: tabProjects, Relation: "m:1"},
		},
		{Title: "Vendor", Min: 12, Max: 20, Flex: true},
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
			Link:  &columnLink{TargetTab: tabAppliances, Relation: "m:1"},
		},
		{Title: "Last", Min: 10, Max: 12, Kind: cellDate},
		{Title: "Next", Min: 10, Max: 12, Kind: cellDate},
		{Title: "Every", Min: 6, Max: 10},
		{Title: "Manual", Min: 8, Max: 14, Flex: true},
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
		{Title: "Warranty", Min: 10, Max: 12, Kind: cellDate},
		{Title: "Cost", Min: 8, Max: 12, Align: alignRight, Kind: cellMoney},
	}
}

func applianceRows(
	items []data.Appliance,
) ([]table.Row, []rowMeta, [][]cell) {
	rows := make([]table.Row, 0, len(items))
	meta := make([]rowMeta, 0, len(items))
	cells := make([][]cell, 0, len(items))
	for _, item := range items {
		deleted := item.DeletedAt.Valid
		rowCells := []cell{
			{Value: fmt.Sprintf("%d", item.ID), Kind: cellReadonly},
			{Value: item.Name, Kind: cellText},
			{Value: item.Brand, Kind: cellText},
			{Value: item.ModelNumber, Kind: cellText},
			{Value: item.SerialNumber, Kind: cellText},
			{Value: item.Location, Kind: cellText},
			{Value: dateValue(item.PurchaseDate), Kind: cellDate},
			{Value: dateValue(item.WarrantyExpiry), Kind: cellDate},
			{Value: centsValue(item.CostCents), Kind: cellMoney},
		}
		rows = append(rows, cellsToRow(rowCells))
		cells = append(cells, rowCells)
		meta = append(meta, rowMeta{
			ID:      item.ID,
			Deleted: deleted,
		})
	}
	return rows, meta, cells
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
) ([]table.Row, []rowMeta, [][]cell) {
	rows := make([]table.Row, 0, len(projects))
	meta := make([]rowMeta, 0, len(projects))
	cells := make([][]cell, 0, len(projects))
	for _, project := range projects {
		deleted := project.DeletedAt.Valid
		rowCells := []cell{
			{Value: fmt.Sprintf("%d", project.ID), Kind: cellReadonly},
			{Value: project.ProjectType.Name, Kind: cellText},
			{Value: project.Title, Kind: cellText},
			{Value: project.Status, Kind: cellStatus},
			{Value: centsValue(project.BudgetCents), Kind: cellMoney},
			{Value: centsValue(project.ActualCents), Kind: cellMoney},
			{Value: dateValue(project.StartDate), Kind: cellDate},
			{Value: dateValue(project.EndDate), Kind: cellDate},
		}
		rows = append(rows, cellsToRow(rowCells))
		cells = append(cells, rowCells)
		meta = append(meta, rowMeta{
			ID:      project.ID,
			Deleted: deleted,
		})
	}
	return rows, meta, cells
}

func quoteRows(
	quotes []data.Quote,
) ([]table.Row, []rowMeta, [][]cell) {
	rows := make([]table.Row, 0, len(quotes))
	meta := make([]rowMeta, 0, len(quotes))
	cells := make([][]cell, 0, len(quotes))
	for _, quote := range quotes {
		deleted := quote.DeletedAt.Valid
		projectName := quote.Project.Title
		if projectName == "" {
			projectName = fmt.Sprintf("Project %d", quote.ProjectID)
		}
		rowCells := []cell{
			{Value: fmt.Sprintf("%d", quote.ID), Kind: cellReadonly},
			{Value: projectName, Kind: cellText, LinkID: quote.ProjectID},
			{Value: quote.Vendor.Name, Kind: cellText},
			{Value: data.FormatCents(quote.TotalCents), Kind: cellMoney},
			{Value: centsValue(quote.LaborCents), Kind: cellMoney},
			{Value: centsValue(quote.MaterialsCents), Kind: cellMoney},
			{Value: centsValue(quote.OtherCents), Kind: cellMoney},
			{Value: dateValue(quote.ReceivedDate), Kind: cellDate},
		}
		rows = append(rows, cellsToRow(rowCells))
		cells = append(cells, rowCells)
		meta = append(meta, rowMeta{
			ID:      quote.ID,
			Deleted: deleted,
		})
	}
	return rows, meta, cells
}

func maintenanceRows(
	items []data.MaintenanceItem,
) ([]table.Row, []rowMeta, [][]cell) {
	rows := make([]table.Row, 0, len(items))
	meta := make([]rowMeta, 0, len(items))
	cells := make([][]cell, 0, len(items))
	for _, item := range items {
		deleted := item.DeletedAt.Valid
		manual := manualSummary(item)
		interval := ""
		if item.IntervalMonths > 0 {
			interval = fmt.Sprintf("%d mo", item.IntervalMonths)
		}
		appName := ""
		var appLinkID uint
		if item.ApplianceID != nil {
			appName = item.Appliance.Name
			appLinkID = *item.ApplianceID
		}
		rowCells := []cell{
			{Value: fmt.Sprintf("%d", item.ID), Kind: cellReadonly},
			{Value: item.Name, Kind: cellText},
			{Value: item.Category.Name, Kind: cellText},
			{Value: appName, Kind: cellText, LinkID: appLinkID},
			{Value: dateValue(item.LastServicedAt), Kind: cellDate},
			{Value: dateValue(item.NextDueAt), Kind: cellDate},
			{Value: interval, Kind: cellText},
			{Value: manual, Kind: cellText},
		}
		rows = append(rows, cellsToRow(rowCells))
		cells = append(cells, rowCells)
		meta = append(meta, rowMeta{
			ID:      item.ID,
			Deleted: deleted,
		})
	}
	return rows, meta, cells
}

func cellsToRow(cells []cell) table.Row {
	row := make(table.Row, len(cells))
	for i, cell := range cells {
		row[i] = cell.Value
	}
	return row
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

func manualSummary(item data.MaintenanceItem) string {
	if item.ManualText != "" {
		return "stored"
	}
	if item.ManualURL != "" {
		return "link"
	}
	return ""
}
