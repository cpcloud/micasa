package app

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/micasa/micasa/internal/data"
)

type houseFormData struct {
	Nickname         string
	AddressLine1     string
	AddressLine2     string
	City             string
	State            string
	PostalCode       string
	YearBuilt        string
	SquareFeet       string
	LotSquareFeet    string
	Bedrooms         string
	Bathrooms        string
	FoundationType   string
	WiringType       string
	RoofType         string
	ExteriorType     string
	HeatingType      string
	CoolingType      string
	WaterSource      string
	SewerType        string
	ParkingType      string
	BasementType     string
	InsuranceCarrier string
	InsurancePolicy  string
	InsuranceRenewal string
	PropertyTax      string
	HOAName          string
	HOAFee           string
}

type projectFormData struct {
	Title         string
	ProjectTypeID uint
	Status        string
	Budget        string
	Actual        string
	StartDate     string
	EndDate       string
	Description   string
}

type quoteFormData struct {
	ProjectID    uint
	VendorName   string
	ContactName  string
	Email        string
	Phone        string
	Website      string
	Total        string
	Labor        string
	Materials    string
	Other        string
	ReceivedDate string
	Notes        string
}

type maintenanceFormData struct {
	Name           string
	CategoryID     uint
	LastServiced   string
	NextDue        string
	IntervalMonths string
	ManualURL      string
	ManualText     string
	Cost           string
	Notes          string
}

func (m *Model) startHouseForm() {
	values := &houseFormData{}
	if m.hasHouse {
		values = houseFormValues(m.house)
	}
	m.logDebug("Opening house profile form.")
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Nickname").
				Description("Ex: Primary Residence").
				Value(&values.Nickname).
				Validate(requiredText("nickname")),
			huh.NewInput().Title("Address line 1").Value(&values.AddressLine1),
			huh.NewInput().Title("Address line 2").Value(&values.AddressLine2),
			huh.NewInput().Title("City").Value(&values.City),
			huh.NewInput().Title("State").Value(&values.State),
			huh.NewInput().Title("Postal code").Value(&values.PostalCode),
		).Title("Basics"),
		huh.NewGroup(
			huh.NewInput().
				Title("Year built").
				Placeholder("1998").
				Value(&values.YearBuilt).
				Validate(optionalInt("year built")),
			huh.NewInput().
				Title("Square feet").
				Placeholder("1820").
				Value(&values.SquareFeet).
				Validate(optionalInt("square feet")),
			huh.NewInput().
				Title("Lot square feet").
				Placeholder("7000").
				Value(&values.LotSquareFeet).
				Validate(optionalInt("lot square feet")),
			huh.NewInput().
				Title("Bedrooms").
				Placeholder("3").
				Value(&values.Bedrooms).
				Validate(optionalInt("bedrooms")),
			huh.NewInput().
				Title("Bathrooms").
				Placeholder("2.5").
				Value(&values.Bathrooms).
				Validate(optionalFloat("bathrooms")),
			huh.NewInput().Title("Foundation type").Value(&values.FoundationType),
			huh.NewInput().Title("Wiring type").Value(&values.WiringType),
			huh.NewInput().Title("Roof type").Value(&values.RoofType),
			huh.NewInput().Title("Exterior type").Value(&values.ExteriorType),
			huh.NewInput().Title("Basement type").Value(&values.BasementType),
		).Title("Structure"),
		huh.NewGroup(
			huh.NewInput().Title("Heating type").Value(&values.HeatingType),
			huh.NewInput().Title("Cooling type").Value(&values.CoolingType),
			huh.NewInput().Title("Water source").Value(&values.WaterSource),
			huh.NewInput().Title("Sewer type").Value(&values.SewerType),
			huh.NewInput().Title("Parking type").Value(&values.ParkingType),
		).Title("Utilities"),
		huh.NewGroup(
			huh.NewInput().Title("Insurance carrier").Value(&values.InsuranceCarrier),
			huh.NewInput().Title("Insurance policy").Value(&values.InsurancePolicy),
			huh.NewInput().
				Title("Insurance renewal (YYYY-MM-DD)").
				Value(&values.InsuranceRenewal).
				Validate(optionalDate("insurance renewal")),
			huh.NewInput().
				Title("Property tax (annual)").
				Placeholder("4200.00").
				Value(&values.PropertyTax).
				Validate(optionalMoney("property tax")),
			huh.NewInput().Title("HOA name").Value(&values.HOAName),
			huh.NewInput().
				Title("HOA fee (monthly)").
				Placeholder("250.00").
				Value(&values.HOAFee).
				Validate(optionalMoney("HOA fee")),
		).Title("Financial"),
	)
	applyFormDefaults(form)
	m.mode = modeForm
	m.formKind = formHouse
	m.form = form
	m.formData = values
	m.snapshotForm()
}

func (m *Model) startProjectForm() {
	values := &projectFormData{
		Status: data.ProjectStatusPlanned,
	}
	options := projectTypeOptions(m.projectTypes)
	if len(options) > 0 {
		values.ProjectTypeID = options[0].Value
	}
	m.logDebug("Opening project form.")
	m.openProjectForm(values, options)
}

func (m *Model) startEditProjectForm(id uint) error {
	project, err := m.store.GetProject(id)
	if err != nil {
		return fmt.Errorf("load project: %w", err)
	}
	values := &projectFormData{
		Title:         project.Title,
		ProjectTypeID: project.ProjectTypeID,
		Status:        project.Status,
		Budget:        data.FormatOptionalCents(project.BudgetCents),
		Actual:        data.FormatOptionalCents(project.ActualCents),
		StartDate:     data.FormatDate(project.StartDate),
		EndDate:       data.FormatDate(project.EndDate),
		Description:   project.Description,
	}
	options := projectTypeOptions(m.projectTypes)
	m.logDebug(fmt.Sprintf("Editing project %d.", id))
	m.editID = &id
	m.openProjectForm(values, options)
	return nil
}

func (m *Model) openProjectForm(values *projectFormData, options []huh.Option[uint]) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Value(&values.Title).
				Validate(requiredText("title")),
			huh.NewSelect[uint]().
				Title("Project type").
				Options(options...).
				Value(&values.ProjectTypeID),
			huh.NewSelect[string]().
				Title("Status").
				Options(statusOptions()...).
				Value(&values.Status),
			huh.NewInput().
				Title("Budget").
				Placeholder("1250.00").
				Value(&values.Budget).
				Validate(optionalMoney("budget")),
			huh.NewInput().
				Title("Actual cost").
				Placeholder("1400.00").
				Value(&values.Actual).
				Validate(optionalMoney("actual cost")),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Start date (YYYY-MM-DD)").
				Value(&values.StartDate).
				Validate(optionalDate("start date")),
			huh.NewInput().
				Title("End date (YYYY-MM-DD)").
				Value(&values.EndDate).
				Validate(optionalDate("end date")),
			huh.NewText().
				Title("Description").
				Value(&values.Description),
		).Title("Timeline"),
	)
	applyFormDefaults(form)
	m.mode = modeForm
	m.formKind = formProject
	m.form = form
	m.formData = values
	m.snapshotForm()
}

func (m *Model) startQuoteForm() error {
	projects, err := m.store.ListProjects(false)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return fmt.Errorf("add a project before adding quotes")
	}
	m.logDebug("Opening quote form.")
	values := &quoteFormData{}
	options := projectOptions(projects)
	values.ProjectID = options[0].Value
	m.openQuoteForm(values, options)
	return nil
}

func (m *Model) startEditQuoteForm(id uint) error {
	quote, err := m.store.GetQuote(id)
	if err != nil {
		return fmt.Errorf("load quote: %w", err)
	}
	projects, err := m.store.ListProjects(false)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return fmt.Errorf("no projects available")
	}
	values := &quoteFormData{
		ProjectID:    quote.ProjectID,
		VendorName:   quote.Vendor.Name,
		ContactName:  quote.Vendor.ContactName,
		Email:        quote.Vendor.Email,
		Phone:        quote.Vendor.Phone,
		Website:      quote.Vendor.Website,
		Total:        data.FormatCents(quote.TotalCents),
		Labor:        data.FormatOptionalCents(quote.LaborCents),
		Materials:    data.FormatOptionalCents(quote.MaterialsCents),
		Other:        data.FormatOptionalCents(quote.OtherCents),
		ReceivedDate: data.FormatDate(quote.ReceivedDate),
		Notes:        quote.Notes,
	}
	options := projectOptions(projects)
	m.logDebug(fmt.Sprintf("Editing quote %d.", id))
	m.editID = &id
	m.openQuoteForm(values, options)
	return nil
}

func (m *Model) openQuoteForm(values *quoteFormData, projectOpts []huh.Option[uint]) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[uint]().
				Title("Project").
				Options(projectOpts...).
				Value(&values.ProjectID),
			huh.NewInput().
				Title("Vendor name").
				Value(&values.VendorName).
				Validate(requiredText("vendor name")),
			huh.NewInput().Title("Contact name").Value(&values.ContactName),
			huh.NewInput().Title("Email").Value(&values.Email),
			huh.NewInput().Title("Phone").Value(&values.Phone),
			huh.NewInput().Title("Website").Value(&values.Website),
		).Title("Vendor"),
		huh.NewGroup(
			huh.NewInput().
				Title("Total").
				Placeholder("3250.00").
				Value(&values.Total).
				Validate(requiredMoney("total")),
			huh.NewInput().
				Title("Labor").
				Placeholder("2000.00").
				Value(&values.Labor).
				Validate(optionalMoney("labor")),
			huh.NewInput().
				Title("Materials").
				Placeholder("1000.00").
				Value(&values.Materials).
				Validate(optionalMoney("materials")),
			huh.NewInput().
				Title("Other").
				Placeholder("250.00").
				Value(&values.Other).
				Validate(optionalMoney("other costs")),
			huh.NewInput().
				Title("Received date (YYYY-MM-DD)").
				Value(&values.ReceivedDate).
				Validate(optionalDate("received date")),
			huh.NewText().Title("Notes").Value(&values.Notes),
		).Title("Quote"),
	)
	applyFormDefaults(form)
	m.mode = modeForm
	m.formKind = formQuote
	m.form = form
	m.formData = values
	m.snapshotForm()
}

func (m *Model) startMaintenanceForm() {
	values := &maintenanceFormData{}
	options := maintenanceOptions(m.maintenanceCategories)
	if len(options) > 0 {
		values.CategoryID = options[0].Value
	}
	m.logDebug("Opening maintenance form.")
	m.openMaintenanceForm(values, options)
}

func (m *Model) startEditMaintenanceForm(id uint) error {
	item, err := m.store.GetMaintenance(id)
	if err != nil {
		return fmt.Errorf("load maintenance item: %w", err)
	}
	values := &maintenanceFormData{
		Name:           item.Name,
		CategoryID:     item.CategoryID,
		LastServiced:   data.FormatDate(item.LastServicedAt),
		NextDue:        data.FormatDate(item.NextDueAt),
		IntervalMonths: intToString(item.IntervalMonths),
		ManualURL:      item.ManualURL,
		ManualText:     item.ManualText,
		Cost:           data.FormatOptionalCents(item.CostCents),
		Notes:          item.Notes,
	}
	options := maintenanceOptions(m.maintenanceCategories)
	m.logDebug(fmt.Sprintf("Editing maintenance item %d.", id))
	m.editID = &id
	m.openMaintenanceForm(values, options)
	return nil
}

func (m *Model) openMaintenanceForm(values *maintenanceFormData, options []huh.Option[uint]) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Item").
				Value(&values.Name).
				Validate(requiredText("item")),
			huh.NewSelect[uint]().
				Title("Category").
				Options(options...).
				Value(&values.CategoryID),
			huh.NewInput().
				Title("Last serviced (YYYY-MM-DD)").
				Value(&values.LastServiced).
				Validate(optionalDate("last serviced")),
			huh.NewInput().
				Title("Interval months").
				Placeholder("6").
				Value(&values.IntervalMonths).
				Validate(optionalInt("interval months")),
			huh.NewInput().
				Title("Next due (YYYY-MM-DD)").
				Value(&values.NextDue).
				Validate(optionalDate("next due")),
		).Title("Schedule"),
		huh.NewGroup(
			huh.NewInput().Title("Manual URL").Value(&values.ManualURL),
			huh.NewText().Title("Manual notes").Value(&values.ManualText),
			huh.NewInput().
				Title("Cost").
				Placeholder("125.00").
				Value(&values.Cost).
				Validate(optionalMoney("cost")),
			huh.NewText().Title("Notes").Value(&values.Notes),
		).Title("Details"),
	)
	applyFormDefaults(form)
	m.mode = modeForm
	m.formKind = formMaintenance
	m.form = form
	m.formData = values
	m.snapshotForm()
}

func applyFormDefaults(form *huh.Form) {
	form.WithShowErrors(true)
	form.WithKeyMap(formKeyMap())
}

func formKeyMap() *huh.KeyMap {
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit.SetKeys("esc")
	keymap.Quit.SetHelp("esc", "cancel")
	return keymap
}

func (m *Model) handleFormSubmit() error {
	switch m.formKind {
	case formHouse:
		return m.submitHouseForm()
	case formProject:
		if m.editID != nil {
			return m.submitEditProjectForm(*m.editID)
		}
		return m.submitProjectForm()
	case formQuote:
		if m.editID != nil {
			return m.submitEditQuoteForm(*m.editID)
		}
		return m.submitQuoteForm()
	case formMaintenance:
		if m.editID != nil {
			return m.submitEditMaintenanceForm(*m.editID)
		}
		return m.submitMaintenanceForm()
	default:
		return nil
	}
}

func (m *Model) submitHouseForm() error {
	values, ok := m.formData.(*houseFormData)
	if !ok {
		return fmt.Errorf("unexpected house form data")
	}
	yearBuilt, err := data.ParseOptionalInt(values.YearBuilt)
	if err != nil {
		return err
	}
	sqft, err := data.ParseOptionalInt(values.SquareFeet)
	if err != nil {
		return err
	}
	lotSqft, err := data.ParseOptionalInt(values.LotSquareFeet)
	if err != nil {
		return err
	}
	bedrooms, err := data.ParseOptionalInt(values.Bedrooms)
	if err != nil {
		return err
	}
	bathrooms, err := data.ParseOptionalFloat(values.Bathrooms)
	if err != nil {
		return err
	}
	insuranceRenewal, err := data.ParseOptionalDate(values.InsuranceRenewal)
	if err != nil {
		return err
	}
	propertyTax, err := data.ParseOptionalCents(values.PropertyTax)
	if err != nil {
		return err
	}
	hoaFee, err := data.ParseOptionalCents(values.HOAFee)
	if err != nil {
		return err
	}
	profile := data.HouseProfile{
		Nickname:         strings.TrimSpace(values.Nickname),
		AddressLine1:     strings.TrimSpace(values.AddressLine1),
		AddressLine2:     strings.TrimSpace(values.AddressLine2),
		City:             strings.TrimSpace(values.City),
		State:            strings.TrimSpace(values.State),
		PostalCode:       strings.TrimSpace(values.PostalCode),
		YearBuilt:        yearBuilt,
		SquareFeet:       sqft,
		LotSquareFeet:    lotSqft,
		Bedrooms:         bedrooms,
		Bathrooms:        bathrooms,
		FoundationType:   strings.TrimSpace(values.FoundationType),
		WiringType:       strings.TrimSpace(values.WiringType),
		RoofType:         strings.TrimSpace(values.RoofType),
		ExteriorType:     strings.TrimSpace(values.ExteriorType),
		HeatingType:      strings.TrimSpace(values.HeatingType),
		CoolingType:      strings.TrimSpace(values.CoolingType),
		WaterSource:      strings.TrimSpace(values.WaterSource),
		SewerType:        strings.TrimSpace(values.SewerType),
		ParkingType:      strings.TrimSpace(values.ParkingType),
		BasementType:     strings.TrimSpace(values.BasementType),
		InsuranceCarrier: strings.TrimSpace(values.InsuranceCarrier),
		InsurancePolicy:  strings.TrimSpace(values.InsurancePolicy),
		InsuranceRenewal: insuranceRenewal,
		PropertyTaxCents: propertyTax,
		HOAName:          strings.TrimSpace(values.HOAName),
		HOAFeeCents:      hoaFee,
	}
	if m.hasHouse {
		if err := m.store.UpdateHouseProfile(profile); err != nil {
			return err
		}
	} else {
		if err := m.store.CreateHouseProfile(profile); err != nil {
			return err
		}
	}
	m.house = profile
	m.hasHouse = true
	return nil
}

func (m *Model) submitEditProjectForm(id uint) error {
	values, ok := m.formData.(*projectFormData)
	if !ok {
		return fmt.Errorf("unexpected project form data")
	}
	budget, err := data.ParseOptionalCents(values.Budget)
	if err != nil {
		return err
	}
	actual, err := data.ParseOptionalCents(values.Actual)
	if err != nil {
		return err
	}
	startDate, err := data.ParseOptionalDate(values.StartDate)
	if err != nil {
		return err
	}
	endDate, err := data.ParseOptionalDate(values.EndDate)
	if err != nil {
		return err
	}
	project := data.Project{
		ID:            id,
		Title:         strings.TrimSpace(values.Title),
		ProjectTypeID: values.ProjectTypeID,
		Status:        values.Status,
		Description:   strings.TrimSpace(values.Description),
		StartDate:     startDate,
		EndDate:       endDate,
		BudgetCents:   budget,
		ActualCents:   actual,
	}
	return m.store.UpdateProject(project)
}

func (m *Model) submitProjectForm() error {
	values, ok := m.formData.(*projectFormData)
	if !ok {
		return fmt.Errorf("unexpected project form data")
	}
	budget, err := data.ParseOptionalCents(values.Budget)
	if err != nil {
		return err
	}
	actual, err := data.ParseOptionalCents(values.Actual)
	if err != nil {
		return err
	}
	startDate, err := data.ParseOptionalDate(values.StartDate)
	if err != nil {
		return err
	}
	endDate, err := data.ParseOptionalDate(values.EndDate)
	if err != nil {
		return err
	}
	project := data.Project{
		Title:         strings.TrimSpace(values.Title),
		ProjectTypeID: values.ProjectTypeID,
		Status:        values.Status,
		Description:   strings.TrimSpace(values.Description),
		StartDate:     startDate,
		EndDate:       endDate,
		BudgetCents:   budget,
		ActualCents:   actual,
	}
	return m.store.CreateProject(project)
}

func (m *Model) submitEditQuoteForm(id uint) error {
	values, ok := m.formData.(*quoteFormData)
	if !ok {
		return fmt.Errorf("unexpected quote form data")
	}
	total, err := data.ParseRequiredCents(values.Total)
	if err != nil {
		return err
	}
	labor, err := data.ParseOptionalCents(values.Labor)
	if err != nil {
		return err
	}
	materials, err := data.ParseOptionalCents(values.Materials)
	if err != nil {
		return err
	}
	other, err := data.ParseOptionalCents(values.Other)
	if err != nil {
		return err
	}
	received, err := data.ParseOptionalDate(values.ReceivedDate)
	if err != nil {
		return err
	}
	quote := data.Quote{
		ID:             id,
		ProjectID:      values.ProjectID,
		TotalCents:     total,
		LaborCents:     labor,
		MaterialsCents: materials,
		OtherCents:     other,
		ReceivedDate:   received,
		Notes:          strings.TrimSpace(values.Notes),
	}
	vendor := data.Vendor{
		Name:        strings.TrimSpace(values.VendorName),
		ContactName: strings.TrimSpace(values.ContactName),
		Email:       strings.TrimSpace(values.Email),
		Phone:       strings.TrimSpace(values.Phone),
		Website:     strings.TrimSpace(values.Website),
	}
	return m.store.UpdateQuote(quote, vendor)
}

func (m *Model) submitQuoteForm() error {
	values, ok := m.formData.(*quoteFormData)
	if !ok {
		return fmt.Errorf("unexpected quote form data")
	}
	total, err := data.ParseRequiredCents(values.Total)
	if err != nil {
		return err
	}
	labor, err := data.ParseOptionalCents(values.Labor)
	if err != nil {
		return err
	}
	materials, err := data.ParseOptionalCents(values.Materials)
	if err != nil {
		return err
	}
	other, err := data.ParseOptionalCents(values.Other)
	if err != nil {
		return err
	}
	received, err := data.ParseOptionalDate(values.ReceivedDate)
	if err != nil {
		return err
	}
	quote := data.Quote{
		ProjectID:      values.ProjectID,
		TotalCents:     total,
		LaborCents:     labor,
		MaterialsCents: materials,
		OtherCents:     other,
		ReceivedDate:   received,
		Notes:          strings.TrimSpace(values.Notes),
	}
	vendor := data.Vendor{
		Name:        strings.TrimSpace(values.VendorName),
		ContactName: strings.TrimSpace(values.ContactName),
		Email:       strings.TrimSpace(values.Email),
		Phone:       strings.TrimSpace(values.Phone),
		Website:     strings.TrimSpace(values.Website),
	}
	return m.store.CreateQuote(quote, vendor)
}

func (m *Model) submitEditMaintenanceForm(id uint) error {
	values, ok := m.formData.(*maintenanceFormData)
	if !ok {
		return fmt.Errorf("unexpected maintenance form data")
	}
	lastServiced, err := data.ParseOptionalDate(values.LastServiced)
	if err != nil {
		return err
	}
	nextDue, err := data.ParseOptionalDate(values.NextDue)
	if err != nil {
		return err
	}
	interval, err := data.ParseOptionalInt(values.IntervalMonths)
	if err != nil {
		return err
	}
	if nextDue == nil {
		nextDue = data.ComputeNextDue(lastServiced, interval)
	}
	cost, err := data.ParseOptionalCents(values.Cost)
	if err != nil {
		return err
	}
	item := data.MaintenanceItem{
		ID:             id,
		Name:           strings.TrimSpace(values.Name),
		CategoryID:     values.CategoryID,
		LastServicedAt: lastServiced,
		NextDueAt:      nextDue,
		IntervalMonths: interval,
		ManualURL:      strings.TrimSpace(values.ManualURL),
		ManualText:     strings.TrimSpace(values.ManualText),
		CostCents:      cost,
		Notes:          strings.TrimSpace(values.Notes),
	}
	return m.store.UpdateMaintenance(item)
}

func (m *Model) submitMaintenanceForm() error {
	values, ok := m.formData.(*maintenanceFormData)
	if !ok {
		return fmt.Errorf("unexpected maintenance form data")
	}
	lastServiced, err := data.ParseOptionalDate(values.LastServiced)
	if err != nil {
		return err
	}
	nextDue, err := data.ParseOptionalDate(values.NextDue)
	if err != nil {
		return err
	}
	interval, err := data.ParseOptionalInt(values.IntervalMonths)
	if err != nil {
		return err
	}
	if nextDue == nil {
		nextDue = data.ComputeNextDue(lastServiced, interval)
	}
	cost, err := data.ParseOptionalCents(values.Cost)
	if err != nil {
		return err
	}
	item := data.MaintenanceItem{
		Name:           strings.TrimSpace(values.Name),
		CategoryID:     values.CategoryID,
		LastServicedAt: lastServiced,
		NextDueAt:      nextDue,
		IntervalMonths: interval,
		ManualURL:      strings.TrimSpace(values.ManualURL),
		ManualText:     strings.TrimSpace(values.ManualText),
		CostCents:      cost,
		Notes:          strings.TrimSpace(values.Notes),
	}
	return m.store.CreateMaintenance(item)
}

func projectTypeOptions(types []data.ProjectType) []huh.Option[uint] {
	options := make([]huh.Option[uint], 0, len(types))
	for _, projectType := range types {
		options = append(options, huh.NewOption(projectType.Name, projectType.ID))
	}
	return options
}

func maintenanceOptions(
	categories []data.MaintenanceCategory,
) []huh.Option[uint] {
	options := make([]huh.Option[uint], 0, len(categories))
	for _, category := range categories {
		options = append(options, huh.NewOption(category.Name, category.ID))
	}
	return options
}

func projectOptions(projects []data.Project) []huh.Option[uint] {
	options := make([]huh.Option[uint], 0, len(projects))
	for _, project := range projects {
		label := project.Title
		if label == "" {
			label = fmt.Sprintf("Project %d", project.ID)
		}
		options = append(options, huh.NewOption(label, project.ID))
	}
	return options
}

func statusOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("planned", data.ProjectStatusPlanned),
		huh.NewOption("quoted", data.ProjectStatusQuoted),
		huh.NewOption("in progress", data.ProjectStatusInProgress),
		huh.NewOption("completed", data.ProjectStatusCompleted),
	}
}

func requiredText(label string) func(string) error {
	return func(input string) error {
		if strings.TrimSpace(input) == "" {
			return fmt.Errorf("%s is required", label)
		}
		return nil
	}
}

func optionalInt(label string) func(string) error {
	return func(input string) error {
		if _, err := data.ParseOptionalInt(input); err != nil {
			return fmt.Errorf("%s should be a whole number", label)
		}
		return nil
	}
}

func optionalFloat(label string) func(string) error {
	return func(input string) error {
		if _, err := data.ParseOptionalFloat(input); err != nil {
			return fmt.Errorf("%s should be a number like 2.5", label)
		}
		return nil
	}
}

func optionalDate(label string) func(string) error {
	return func(input string) error {
		if _, err := data.ParseOptionalDate(input); err != nil {
			return fmt.Errorf("%s should be YYYY-MM-DD", label)
		}
		return nil
	}
}

func optionalMoney(label string) func(string) error {
	return func(input string) error {
		if _, err := data.ParseOptionalCents(input); err != nil {
			return fmt.Errorf("%s should look like 1250.00", label)
		}
		return nil
	}
}

func requiredMoney(label string) func(string) error {
	return func(input string) error {
		if _, err := data.ParseRequiredCents(input); err != nil {
			return fmt.Errorf("%s should look like 1250.00", label)
		}
		return nil
	}
}

func houseFormValues(profile data.HouseProfile) *houseFormData {
	return &houseFormData{
		Nickname:         profile.Nickname,
		AddressLine1:     profile.AddressLine1,
		AddressLine2:     profile.AddressLine2,
		City:             profile.City,
		State:            profile.State,
		PostalCode:       profile.PostalCode,
		YearBuilt:        intToString(profile.YearBuilt),
		SquareFeet:       intToString(profile.SquareFeet),
		LotSquareFeet:    intToString(profile.LotSquareFeet),
		Bedrooms:         intToString(profile.Bedrooms),
		Bathrooms:        floatToString(profile.Bathrooms),
		FoundationType:   profile.FoundationType,
		WiringType:       profile.WiringType,
		RoofType:         profile.RoofType,
		ExteriorType:     profile.ExteriorType,
		HeatingType:      profile.HeatingType,
		CoolingType:      profile.CoolingType,
		WaterSource:      profile.WaterSource,
		SewerType:        profile.SewerType,
		ParkingType:      profile.ParkingType,
		BasementType:     profile.BasementType,
		InsuranceCarrier: profile.InsuranceCarrier,
		InsurancePolicy:  profile.InsurancePolicy,
		InsuranceRenewal: data.FormatDate(profile.InsuranceRenewal),
		PropertyTax:      data.FormatOptionalCents(profile.PropertyTaxCents),
		HOAName:          profile.HOAName,
		HOAFee:           data.FormatOptionalCents(profile.HOAFeeCents),
	}
}

func intToString(value int) string {
	if value == 0 {
		return ""
	}
	return strconv.Itoa(value)
}

func floatToString(value float64) string {
	if value == 0 {
		return ""
	}
	if value == math.Trunc(value) {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.1f", value)
}
