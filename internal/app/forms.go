// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/cpcloud/micasa/internal/data"
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
	ApplianceID    uint // 0 means none
	LastServiced   string
	IntervalMonths string
	ManualURL      string
	ManualText     string
	Cost           string
	Notes          string
}

type serviceLogFormData struct {
	MaintenanceItemID uint
	ServicedAt        string
	VendorID          uint // 0 = self
	Cost              string
	Notes             string
}

type vendorFormData struct {
	Name        string
	ContactName string
	Email       string
	Phone       string
	Website     string
	Notes       string
}

type applianceFormData struct {
	Name           string
	Brand          string
	ModelNumber    string
	SerialNumber   string
	PurchaseDate   string
	WarrantyExpiry string
	Location       string
	Cost           string
	Notes          string
}

func (m *Model) startHouseForm() {
	values := &houseFormData{}
	if m.hasHouse {
		values = houseFormValues(m.house)
	}
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
	formWidth := 60
	if m.width > 0 && m.width < formWidth+10 {
		formWidth = m.width - 10
	}
	form.WithWidth(formWidth)
	m.activateForm(formHouse, form, values)
}

func (m *Model) startProjectForm() {
	values := &projectFormData{
		Status: data.ProjectStatusPlanned,
	}
	options := projectTypeOptions(m.projectTypes)
	if len(options) > 0 {
		values.ProjectTypeID = options[0].Value
	}
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
		),
	)
	m.activateForm(formProject, form, values)
}

func (m *Model) startEditProjectForm(id uint) error {
	project, err := m.store.GetProject(id)
	if err != nil {
		return fmt.Errorf("load project: %w", err)
	}
	values := projectFormValues(project)
	options := projectTypeOptions(m.projectTypes)
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
	m.activateForm(formProject, form, values)
}

func (m *Model) startQuoteForm() error {
	projects, err := m.store.ListProjects(false)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return fmt.Errorf("add a project before adding quotes")
	}
	values := &quoteFormData{}
	options := projectOptions(projects)
	values.ProjectID = options[0].Value
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[uint]().
				Title("Project").
				Options(options...).
				Value(&values.ProjectID),
			huh.NewInput().
				Title("Vendor name").
				Value(&values.VendorName).
				Validate(requiredText("vendor name")),
			huh.NewInput().
				Title("Total").
				Placeholder("3250.00").
				Value(&values.Total).
				Validate(requiredMoney("total")),
		),
	)
	m.activateForm(formQuote, form, values)
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
	values := quoteFormValues(quote)
	options := projectOptions(projects)
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
	m.activateForm(formQuote, form, values)
}

func (m *Model) startMaintenanceForm() {
	values := &maintenanceFormData{}
	catOptions := maintenanceOptions(m.maintenanceCategories)
	if len(catOptions) > 0 {
		values.CategoryID = catOptions[0].Value
	}
	appliances, _ := m.store.ListAppliances(false)
	appOpts := applianceOptions(appliances)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Item").
				Value(&values.Name).
				Validate(requiredText("item")),
			huh.NewSelect[uint]().
				Title("Category").
				Options(catOptions...).
				Value(&values.CategoryID),
			huh.NewSelect[uint]().
				Title("Appliance").
				Options(appOpts...).
				Value(&values.ApplianceID),
			huh.NewInput().
				Title("Interval months").
				Placeholder("6").
				Value(&values.IntervalMonths).
				Validate(optionalInt("interval months")),
		),
	)
	m.activateForm(formMaintenance, form, values)
}

func (m *Model) startEditMaintenanceForm(id uint) error {
	item, err := m.store.GetMaintenance(id)
	if err != nil {
		return fmt.Errorf("load maintenance item: %w", err)
	}
	values := maintenanceFormValues(item)
	options := maintenanceOptions(m.maintenanceCategories)
	appliances, _ := m.store.ListAppliances(false)
	appOpts := applianceOptions(appliances)
	m.editID = &id
	m.openMaintenanceForm(values, options, appOpts)
	return nil
}

func (m *Model) openMaintenanceForm(
	values *maintenanceFormData,
	catOptions []huh.Option[uint],
	appOptions []huh.Option[uint],
) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Item").
				Value(&values.Name).
				Validate(requiredText("item")),
			huh.NewSelect[uint]().
				Title("Category").
				Options(catOptions...).
				Value(&values.CategoryID),
			huh.NewSelect[uint]().
				Title("Appliance").
				Options(appOptions...).
				Value(&values.ApplianceID),
			huh.NewInput().
				Title("Last serviced (YYYY-MM-DD)").
				Value(&values.LastServiced).
				Validate(optionalDate("last serviced")),
			huh.NewInput().
				Title("Interval months").
				Placeholder("6").
				Value(&values.IntervalMonths).
				Validate(optionalInt("interval months")),
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
	m.activateForm(formMaintenance, form, values)
}

func (m *Model) startApplianceForm() {
	values := &applianceFormData{}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("Kitchen Refrigerator").
				Value(&values.Name).
				Validate(requiredText("name")),
		),
	)
	m.activateForm(formAppliance, form, values)
}

func (m *Model) startEditApplianceForm(id uint) error {
	item, err := m.store.GetAppliance(id)
	if err != nil {
		return fmt.Errorf("load appliance: %w", err)
	}
	values := applianceFormValues(item)
	m.editID = &id
	m.openApplianceForm(values)
	return nil
}

func (m *Model) openApplianceForm(values *applianceFormData) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("Kitchen Refrigerator").
				Value(&values.Name).
				Validate(requiredText("name")),
			huh.NewInput().Title("Brand").Value(&values.Brand),
			huh.NewInput().Title("Model number").Value(&values.ModelNumber),
			huh.NewInput().Title("Serial number").Value(&values.SerialNumber),
			huh.NewInput().Title("Location").Placeholder("Kitchen").Value(&values.Location),
		).Title("Identity"),
		huh.NewGroup(
			huh.NewInput().
				Title("Purchase date (YYYY-MM-DD)").
				Value(&values.PurchaseDate).
				Validate(optionalDate("purchase date")),
			huh.NewInput().
				Title("Warranty expiry (YYYY-MM-DD)").
				Value(&values.WarrantyExpiry).
				Validate(optionalDate("warranty expiry")),
			huh.NewInput().
				Title("Cost").
				Placeholder("899.00").
				Value(&values.Cost).
				Validate(optionalMoney("cost")),
			huh.NewText().Title("Notes").Value(&values.Notes),
		).Title("Details"),
	)
	m.activateForm(formAppliance, form, values)
}

func (m *Model) submitApplianceForm() error {
	item, err := m.parseApplianceFormData()
	if err != nil {
		return err
	}
	return m.store.CreateAppliance(item)
}

func (m *Model) submitEditApplianceForm(id uint) error {
	item, err := m.parseApplianceFormData()
	if err != nil {
		return err
	}
	item.ID = id
	return m.store.UpdateAppliance(item)
}

func (m *Model) parseApplianceFormData() (data.Appliance, error) {
	values, ok := m.formData.(*applianceFormData)
	if !ok {
		return data.Appliance{}, fmt.Errorf("unexpected appliance form data")
	}
	purchaseDate, err := data.ParseOptionalDate(values.PurchaseDate)
	if err != nil {
		return data.Appliance{}, err
	}
	warrantyExpiry, err := data.ParseOptionalDate(values.WarrantyExpiry)
	if err != nil {
		return data.Appliance{}, err
	}
	cost, err := data.ParseOptionalCents(values.Cost)
	if err != nil {
		return data.Appliance{}, err
	}
	return data.Appliance{
		Name:           strings.TrimSpace(values.Name),
		Brand:          strings.TrimSpace(values.Brand),
		ModelNumber:    strings.TrimSpace(values.ModelNumber),
		SerialNumber:   strings.TrimSpace(values.SerialNumber),
		PurchaseDate:   purchaseDate,
		WarrantyExpiry: warrantyExpiry,
		Location:       strings.TrimSpace(values.Location),
		CostCents:      cost,
		Notes:          strings.TrimSpace(values.Notes),
	}, nil
}

func (m *Model) startVendorForm() {
	values := &vendorFormData{}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("Acme Plumbing").
				Value(&values.Name).
				Validate(requiredText("name")),
		),
	)
	m.activateForm(formVendor, form, values)
}

func (m *Model) startEditVendorForm(id uint) error {
	vendor, err := m.store.GetVendor(id)
	if err != nil {
		return fmt.Errorf("load vendor: %w", err)
	}
	values := vendorFormValues(vendor)
	m.editID = &id
	m.openVendorForm(values)
	return nil
}

func (m *Model) openVendorForm(values *vendorFormData) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("Acme Plumbing").
				Value(&values.Name).
				Validate(requiredText("name")),
			huh.NewInput().Title("Contact name").Value(&values.ContactName),
			huh.NewInput().Title("Email").Value(&values.Email),
			huh.NewInput().Title("Phone").Value(&values.Phone),
			huh.NewInput().Title("Website").Value(&values.Website),
			huh.NewText().Title("Notes").Value(&values.Notes),
		),
	)
	m.activateForm(formVendor, form, values)
}

func (m *Model) submitVendorForm() error {
	vendor, err := m.parseVendorFormData()
	if err != nil {
		return err
	}
	return m.store.CreateVendor(vendor)
}

func (m *Model) submitEditVendorForm(id uint) error {
	vendor, err := m.parseVendorFormData()
	if err != nil {
		return err
	}
	vendor.ID = id
	return m.store.UpdateVendor(vendor)
}

func (m *Model) parseVendorFormData() (data.Vendor, error) {
	values, ok := m.formData.(*vendorFormData)
	if !ok {
		return data.Vendor{}, fmt.Errorf("unexpected vendor form data")
	}
	return data.Vendor{
		Name:        strings.TrimSpace(values.Name),
		ContactName: strings.TrimSpace(values.ContactName),
		Email:       strings.TrimSpace(values.Email),
		Phone:       strings.TrimSpace(values.Phone),
		Website:     strings.TrimSpace(values.Website),
		Notes:       strings.TrimSpace(values.Notes),
	}, nil
}

func (m *Model) inlineEditVendor(id uint, col int) error {
	vendor, err := m.store.GetVendor(id)
	if err != nil {
		return fmt.Errorf("load vendor: %w", err)
	}
	values := vendorFormValues(vendor)
	// Column mapping: 0=ID, 1=Name, 2=Contact, 3=Email, 4=Phone, 5=Website, 6=Quotes(ro), 7=Jobs(ro)
	var field huh.Field
	switch col {
	case 1:
		field = huh.NewInput().Title("Name").Value(&values.Name).Validate(requiredText("name"))
	case 2:
		field = huh.NewInput().Title("Contact name").Value(&values.ContactName)
	case 3:
		field = huh.NewInput().Title("Email").Value(&values.Email)
	case 4:
		field = huh.NewInput().Title("Phone").Value(&values.Phone)
	case 5:
		field = huh.NewInput().Title("Website").Value(&values.Website)
	default:
		return m.startEditVendorForm(id)
	}
	m.openInlineEdit(id, formVendor, field, values)
	return nil
}

func vendorFormValues(vendor data.Vendor) *vendorFormData {
	return &vendorFormData{
		Name:        vendor.Name,
		ContactName: vendor.ContactName,
		Email:       vendor.Email,
		Phone:       vendor.Phone,
		Website:     vendor.Website,
		Notes:       vendor.Notes,
	}
}

func (m *Model) inlineEditProject(id uint, col int) error {
	project, err := m.store.GetProject(id)
	if err != nil {
		return fmt.Errorf("load project: %w", err)
	}
	values := projectFormValues(project)
	options := projectTypeOptions(m.projectTypes)
	// Column mapping: 0=ID, 1=Type, 2=Title, 3=Status, 4=Budget, 5=Actual, 6=Start, 7=End
	var field huh.Field
	switch col {
	case 1:
		field = huh.NewSelect[uint]().Title("Project type").
			Options(options...).
			Value(&values.ProjectTypeID)
	case 2:
		field = huh.NewInput().Title("Title").Value(&values.Title).Validate(requiredText("title"))
	case 3:
		field = huh.NewSelect[string]().Title("Status").
			Options(statusOptions()...).
			Value(&values.Status)
	case 4:
		field = huh.NewInput().
			Title("Budget").
			Placeholder("1250.00").
			Value(&values.Budget).
			Validate(optionalMoney("budget"))
	case 5:
		field = huh.NewInput().
			Title("Actual cost").
			Placeholder("1400.00").
			Value(&values.Actual).
			Validate(optionalMoney("actual cost"))
	case 6:
		m.openDatePicker(id, formProject, &values.StartDate, values)
		return nil
	case 7:
		m.openDatePicker(id, formProject, &values.EndDate, values)
		return nil
	default:
		return m.startEditProjectForm(id)
	}
	m.openInlineEdit(id, formProject, field, values)
	return nil
}

func (m *Model) inlineEditQuote(id uint, col int) error {
	quote, err := m.store.GetQuote(id)
	if err != nil {
		return fmt.Errorf("load quote: %w", err)
	}
	projects, err := m.store.ListProjects(false)
	if err != nil {
		return err
	}
	values := quoteFormValues(quote)
	projectOpts := projectOptions(projects)
	// Column mapping: 0=ID, 1=Project, 2=Vendor, 3=Total, 4=Labor, 5=Mat, 6=Other, 7=Recv
	var field huh.Field
	switch col {
	case 1:
		field = huh.NewSelect[uint]().Title("Project").
			Options(projectOpts...).
			Value(&values.ProjectID)
	case 2:
		field = huh.NewInput().
			Title("Vendor name").
			Value(&values.VendorName).
			Validate(requiredText("vendor name"))
	case 3:
		field = huh.NewInput().
			Title("Total").
			Placeholder("3250.00").
			Value(&values.Total).
			Validate(requiredMoney("total"))
	case 4:
		field = huh.NewInput().
			Title("Labor").
			Placeholder("2000.00").
			Value(&values.Labor).
			Validate(optionalMoney("labor"))
	case 5:
		field = huh.NewInput().
			Title("Materials").
			Placeholder("1000.00").
			Value(&values.Materials).
			Validate(optionalMoney("materials"))
	case 6:
		field = huh.NewInput().
			Title("Other").
			Placeholder("250.00").
			Value(&values.Other).
			Validate(optionalMoney("other costs"))
	case 7:
		m.openDatePicker(id, formQuote, &values.ReceivedDate, values)
		return nil
	default:
		return m.startEditQuoteForm(id)
	}
	m.openInlineEdit(id, formQuote, field, values)
	return nil
}

func (m *Model) inlineEditMaintenance(id uint, col int) error {
	item, err := m.store.GetMaintenance(id)
	if err != nil {
		return fmt.Errorf("load maintenance item: %w", err)
	}
	values := maintenanceFormValues(item)
	catOptions := maintenanceOptions(m.maintenanceCategories)
	// Column mapping: 0=ID, 1=Item, 2=Category, 3=Appliance, 4=Last, 5=Next(computed), 6=Every, 7=Log
	var field huh.Field
	switch col {
	case 1:
		field = huh.NewInput().Title("Item").Value(&values.Name).Validate(requiredText("item"))
	case 2:
		field = huh.NewSelect[uint]().Title("Category").
			Options(catOptions...).
			Value(&values.CategoryID)
	case 3:
		appliances, loadErr := m.store.ListAppliances(false)
		if loadErr != nil {
			return loadErr
		}
		appOpts := applianceOptions(appliances)
		field = huh.NewSelect[uint]().Title("Appliance").
			Options(appOpts...).
			Value(&values.ApplianceID)
	case 4:
		m.openDatePicker(id, formMaintenance, &values.LastServiced, values)
		return nil
	case 6:
		field = huh.NewInput().
			Title("Interval months").
			Placeholder("6").
			Value(&values.IntervalMonths).
			Validate(optionalInt("interval months"))
	default:
		// Col 0 (ID), 5 (Next Due, computed), 7 (Log) are readonly.
		return m.startEditMaintenanceForm(id)
	}
	m.openInlineEdit(id, formMaintenance, field, values)
	return nil
}

func (m *Model) inlineEditAppliance(id uint, col int) error {
	item, err := m.store.GetAppliance(id)
	if err != nil {
		return fmt.Errorf("load appliance: %w", err)
	}
	values := applianceFormValues(item)
	// Column mapping: 0=ID, 1=Name, 2=Brand, 3=Model, 4=Serial, 5=Location, 6=Purchased, 7=Age(readonly), 8=Warranty, 9=Cost, 10=Maint(readonly)
	var field huh.Field
	switch col {
	case 1:
		field = huh.NewInput().Title("Name").Value(&values.Name).Validate(requiredText("name"))
	case 2:
		field = huh.NewInput().Title("Brand").Value(&values.Brand)
	case 3:
		field = huh.NewInput().Title("Model number").Value(&values.ModelNumber)
	case 4:
		field = huh.NewInput().Title("Serial number").Value(&values.SerialNumber)
	case 5:
		field = huh.NewInput().Title("Location").Value(&values.Location)
	case 6:
		m.openDatePicker(id, formAppliance, &values.PurchaseDate, values)
		return nil
	case 8:
		m.openDatePicker(id, formAppliance, &values.WarrantyExpiry, values)
		return nil
	case 9:
		field = huh.NewInput().
			Title("Cost").
			Placeholder("899.00").
			Value(&values.Cost).
			Validate(optionalMoney("cost"))
	default:
		return m.startEditApplianceForm(id)
	}
	m.openInlineEdit(id, formAppliance, field, values)
	return nil
}

func (m *Model) startServiceLogForm(maintenanceItemID uint) error {
	values := &serviceLogFormData{
		MaintenanceItemID: maintenanceItemID,
		ServicedAt:        time.Now().Format(data.DateLayout),
	}
	vendorOpts := vendorOptions(m.vendors)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Date serviced (YYYY-MM-DD)").
				Value(&values.ServicedAt).
				Validate(requiredDate("date serviced")),
			huh.NewSelect[uint]().
				Title("Performed by").
				Options(vendorOpts...).
				Value(&values.VendorID),
		),
	)
	m.activateForm(formServiceLog, form, values)
	return nil
}

func (m *Model) startEditServiceLogForm(id uint) error {
	entry, err := m.store.GetServiceLog(id)
	if err != nil {
		return fmt.Errorf("load service log: %w", err)
	}
	values := serviceLogFormValues(entry)
	vendorOpts := vendorOptions(m.vendors)
	m.editID = &id
	m.openServiceLogForm(values, vendorOpts)
	return nil
}

func (m *Model) openServiceLogForm(
	values *serviceLogFormData,
	vendorOpts []huh.Option[uint],
) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Date serviced (YYYY-MM-DD)").
				Value(&values.ServicedAt).
				Validate(requiredDate("date serviced")),
			huh.NewSelect[uint]().
				Title("Performed by").
				Options(vendorOpts...).
				Value(&values.VendorID),
			huh.NewInput().
				Title("Cost").
				Placeholder("125.00").
				Value(&values.Cost).
				Validate(optionalMoney("cost")),
			huh.NewText().Title("Notes").Value(&values.Notes),
		),
	)
	m.activateForm(formServiceLog, form, values)
}

func (m *Model) submitServiceLogForm() error {
	entry, vendor, err := m.parseServiceLogFormData()
	if err != nil {
		return err
	}
	return m.store.CreateServiceLog(entry, vendor)
}

func (m *Model) submitEditServiceLogForm(id uint) error {
	entry, vendor, err := m.parseServiceLogFormData()
	if err != nil {
		return err
	}
	entry.ID = id
	return m.store.UpdateServiceLog(entry, vendor)
}

func (m *Model) parseServiceLogFormData() (data.ServiceLogEntry, data.Vendor, error) {
	values, ok := m.formData.(*serviceLogFormData)
	if !ok {
		return data.ServiceLogEntry{}, data.Vendor{}, fmt.Errorf("unexpected service log form data")
	}
	servicedAt, err := data.ParseRequiredDate(values.ServicedAt)
	if err != nil {
		return data.ServiceLogEntry{}, data.Vendor{}, err
	}
	cost, err := data.ParseOptionalCents(values.Cost)
	if err != nil {
		return data.ServiceLogEntry{}, data.Vendor{}, err
	}
	entry := data.ServiceLogEntry{
		MaintenanceItemID: values.MaintenanceItemID,
		ServicedAt:        servicedAt,
		CostCents:         cost,
		Notes:             strings.TrimSpace(values.Notes),
	}
	var vendor data.Vendor
	if values.VendorID > 0 {
		// Look up the vendor to pass to CreateServiceLog/UpdateServiceLog.
		for _, v := range m.vendors {
			if v.ID == values.VendorID {
				vendor = v
				break
			}
		}
	}
	return entry, vendor, nil
}

func (m *Model) inlineEditServiceLog(id uint, col int) error {
	entry, err := m.store.GetServiceLog(id)
	if err != nil {
		return fmt.Errorf("load service log: %w", err)
	}
	values := serviceLogFormValues(entry)
	vendorOpts := vendorOptions(m.vendors)
	// Column mapping: 0=ID, 1=Date, 2=Performed By, 3=Cost, 4=Notes
	var field huh.Field
	switch col {
	case 1:
		m.openDatePicker(id, formServiceLog, &values.ServicedAt, values)
		return nil
	case 2:
		field = huh.NewSelect[uint]().
			Title("Performed by").
			Options(vendorOpts...).
			Value(&values.VendorID)
	case 3:
		field = huh.NewInput().
			Title("Cost").
			Placeholder("125.00").
			Value(&values.Cost).
			Validate(optionalMoney("cost"))
	case 4:
		field = huh.NewText().Title("Notes").Value(&values.Notes)
	default:
		return m.startEditServiceLogForm(id)
	}
	m.openInlineEdit(id, formServiceLog, field, values)
	return nil
}

func serviceLogFormValues(entry data.ServiceLogEntry) *serviceLogFormData {
	var vendorID uint
	if entry.VendorID != nil {
		vendorID = *entry.VendorID
	}
	return &serviceLogFormData{
		MaintenanceItemID: entry.MaintenanceItemID,
		ServicedAt:        entry.ServicedAt.Format(data.DateLayout),
		VendorID:          vendorID,
		Cost:              data.FormatOptionalCents(entry.CostCents),
		Notes:             entry.Notes,
	}
}

func vendorOptions(vendors []data.Vendor) []huh.Option[uint] {
	options := make([]huh.Option[uint], 0, len(vendors)+1)
	options = append(options, huh.NewOption("Self (homeowner)", uint(0)))
	for _, v := range vendors {
		label := v.Name
		if v.ContactName != "" {
			label = fmt.Sprintf("%s (%s)", v.Name, v.ContactName)
		}
		options = append(options, huh.NewOption(label, v.ID))
	}
	return withOrdinals(options)
}

func requiredDate(label string) func(string) error {
	return func(input string) error {
		if strings.TrimSpace(input) == "" {
			return fmt.Errorf("%s is required", label)
		}
		if _, err := data.ParseRequiredDate(input); err != nil {
			return fmt.Errorf("%s should be YYYY-MM-DD", label)
		}
		return nil
	}
}

func applianceOptions(appliances []data.Appliance) []huh.Option[uint] {
	options := make([]huh.Option[uint], 0, len(appliances)+1)
	options = append(options, huh.NewOption("(none)", uint(0)))
	for _, appliance := range appliances {
		label := appliance.Name
		if appliance.Brand != "" {
			label = fmt.Sprintf("%s (%s)", appliance.Name, appliance.Brand)
		}
		options = append(options, huh.NewOption(label, appliance.ID))
	}
	return withOrdinals(options)
}

// openDatePicker opens the calendar picker for an inline date edit.
// When the user picks a date, the form data is saved via the handler.
func (m *Model) openDatePicker(
	id uint,
	kind FormKind,
	dateField *string,
	values any,
) {
	m.editID = &id
	m.formKind = kind
	m.formData = values
	m.openCalendar(dateField, func() {
		m.snapshotForUndo()
		if err := m.handleFormSubmit(); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.setStatusInfo("Saved.")
			m.reloadAll()
		}
		m.formKind = formNone
		m.formData = nil
		m.editID = nil
	})
}

// activateForm applies defaults and switches the model into form mode.
// All form-opening paths should call this instead of duplicating the epilogue.
func (m *Model) activateForm(kind FormKind, form *huh.Form, values any) {
	applyFormDefaults(form)
	m.prevMode = m.mode
	m.mode = modeForm
	m.formKind = kind
	m.form = form
	m.formData = values
	m.snapshotForm()
}

// openInlineEdit sets up a single-field inline edit form.
func (m *Model) openInlineEdit(id uint, kind FormKind, field huh.Field, values any) {
	m.editID = &id
	m.activateForm(kind, huh.NewForm(huh.NewGroup(field)), values)
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
	if m.formKind == formHouse {
		return m.submitHouseForm()
	}
	handler := m.handlerForFormKind(m.formKind)
	if handler == nil {
		return nil
	}
	return handler.SubmitForm(m)
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
	project, err := m.parseProjectFormData()
	if err != nil {
		return err
	}
	project.ID = id
	return m.store.UpdateProject(project)
}

func (m *Model) submitProjectForm() error {
	project, err := m.parseProjectFormData()
	if err != nil {
		return err
	}
	return m.store.CreateProject(project)
}

func (m *Model) parseProjectFormData() (data.Project, error) {
	values, ok := m.formData.(*projectFormData)
	if !ok {
		return data.Project{}, fmt.Errorf("unexpected project form data")
	}
	budget, err := data.ParseOptionalCents(values.Budget)
	if err != nil {
		return data.Project{}, err
	}
	actual, err := data.ParseOptionalCents(values.Actual)
	if err != nil {
		return data.Project{}, err
	}
	startDate, err := data.ParseOptionalDate(values.StartDate)
	if err != nil {
		return data.Project{}, err
	}
	endDate, err := data.ParseOptionalDate(values.EndDate)
	if err != nil {
		return data.Project{}, err
	}
	return data.Project{
		Title:         strings.TrimSpace(values.Title),
		ProjectTypeID: values.ProjectTypeID,
		Status:        values.Status,
		Description:   strings.TrimSpace(values.Description),
		StartDate:     startDate,
		EndDate:       endDate,
		BudgetCents:   budget,
		ActualCents:   actual,
	}, nil
}

func (m *Model) submitEditQuoteForm(id uint) error {
	quote, vendor, err := m.parseQuoteFormData()
	if err != nil {
		return err
	}
	quote.ID = id
	return m.store.UpdateQuote(quote, vendor)
}

func (m *Model) submitQuoteForm() error {
	quote, vendor, err := m.parseQuoteFormData()
	if err != nil {
		return err
	}
	return m.store.CreateQuote(quote, vendor)
}

func (m *Model) parseQuoteFormData() (data.Quote, data.Vendor, error) {
	values, ok := m.formData.(*quoteFormData)
	if !ok {
		return data.Quote{}, data.Vendor{}, fmt.Errorf("unexpected quote form data")
	}
	total, err := data.ParseRequiredCents(values.Total)
	if err != nil {
		return data.Quote{}, data.Vendor{}, err
	}
	labor, err := data.ParseOptionalCents(values.Labor)
	if err != nil {
		return data.Quote{}, data.Vendor{}, err
	}
	materials, err := data.ParseOptionalCents(values.Materials)
	if err != nil {
		return data.Quote{}, data.Vendor{}, err
	}
	other, err := data.ParseOptionalCents(values.Other)
	if err != nil {
		return data.Quote{}, data.Vendor{}, err
	}
	received, err := data.ParseOptionalDate(values.ReceivedDate)
	if err != nil {
		return data.Quote{}, data.Vendor{}, err
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
	return quote, vendor, nil
}

func (m *Model) submitEditMaintenanceForm(id uint) error {
	item, err := m.parseMaintenanceFormData()
	if err != nil {
		return err
	}
	item.ID = id
	return m.store.UpdateMaintenance(item)
}

func (m *Model) submitMaintenanceForm() error {
	item, err := m.parseMaintenanceFormData()
	if err != nil {
		return err
	}
	return m.store.CreateMaintenance(item)
}

func (m *Model) parseMaintenanceFormData() (data.MaintenanceItem, error) {
	values, ok := m.formData.(*maintenanceFormData)
	if !ok {
		return data.MaintenanceItem{}, fmt.Errorf("unexpected maintenance form data")
	}
	lastServiced, err := data.ParseOptionalDate(values.LastServiced)
	if err != nil {
		return data.MaintenanceItem{}, err
	}
	interval, err := data.ParseOptionalInt(values.IntervalMonths)
	if err != nil {
		return data.MaintenanceItem{}, err
	}
	cost, err := data.ParseOptionalCents(values.Cost)
	if err != nil {
		return data.MaintenanceItem{}, err
	}
	var appID *uint
	if values.ApplianceID > 0 {
		appID = &values.ApplianceID
	}
	return data.MaintenanceItem{
		Name:           strings.TrimSpace(values.Name),
		CategoryID:     values.CategoryID,
		ApplianceID:    appID,
		LastServicedAt: lastServiced,
		IntervalMonths: interval,
		ManualURL:      strings.TrimSpace(values.ManualURL),
		ManualText:     strings.TrimSpace(values.ManualText),
		CostCents:      cost,
		Notes:          strings.TrimSpace(values.Notes),
	}, nil
}

func projectTypeOptions(types []data.ProjectType) []huh.Option[uint] {
	options := make([]huh.Option[uint], 0, len(types))
	for _, projectType := range types {
		options = append(options, huh.NewOption(projectType.Name, projectType.ID))
	}
	return withOrdinals(options)
}

func maintenanceOptions(
	categories []data.MaintenanceCategory,
) []huh.Option[uint] {
	options := make([]huh.Option[uint], 0, len(categories))
	for _, category := range categories {
		options = append(options, huh.NewOption(category.Name, category.ID))
	}
	return withOrdinals(options)
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
	return withOrdinals(options)
}

func statusOptions() []huh.Option[string] {
	type entry struct {
		label string
		value string
		color lipgloss.AdaptiveColor
	}
	statuses := []entry{
		{"ideating", data.ProjectStatusIdeating, muted},
		{"planned", data.ProjectStatusPlanned, accent},
		{"quoted", data.ProjectStatusQuoted, secondary},
		{"underway", data.ProjectStatusInProgress, success},
		{"delayed", data.ProjectStatusDelayed, warning},
		{"completed", data.ProjectStatusCompleted, textDim},
		{"abandoned", data.ProjectStatusAbandoned, danger},
	}
	opts := make([]huh.Option[string], len(statuses))
	for i, s := range statuses {
		colored := lipgloss.NewStyle().Foreground(s.color).Render(s.label)
		opts[i] = huh.NewOption(colored, s.value)
	}
	return withOrdinals(opts)
}

// withOrdinals prefixes each option label with its 1-based position so users
// can see which number key jumps to which option.
func withOrdinals[T comparable](opts []huh.Option[T]) []huh.Option[T] {
	for i := range opts {
		opts[i].Key = fmt.Sprintf("%d. %s", i+1, opts[i].Key)
	}
	return opts
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

func projectFormValues(project data.Project) *projectFormData {
	return &projectFormData{
		Title:         project.Title,
		ProjectTypeID: project.ProjectTypeID,
		Status:        project.Status,
		Budget:        data.FormatOptionalCents(project.BudgetCents),
		Actual:        data.FormatOptionalCents(project.ActualCents),
		StartDate:     data.FormatDate(project.StartDate),
		EndDate:       data.FormatDate(project.EndDate),
		Description:   project.Description,
	}
}

func quoteFormValues(quote data.Quote) *quoteFormData {
	return &quoteFormData{
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
}

func maintenanceFormValues(item data.MaintenanceItem) *maintenanceFormData {
	var appID uint
	if item.ApplianceID != nil {
		appID = *item.ApplianceID
	}
	return &maintenanceFormData{
		Name:           item.Name,
		CategoryID:     item.CategoryID,
		ApplianceID:    appID,
		LastServiced:   data.FormatDate(item.LastServicedAt),
		IntervalMonths: intToString(item.IntervalMonths),
		ManualURL:      item.ManualURL,
		ManualText:     item.ManualText,
		Cost:           data.FormatOptionalCents(item.CostCents),
		Notes:          item.Notes,
	}
}

func applianceFormValues(item data.Appliance) *applianceFormData {
	return &applianceFormData{
		Name:           item.Name,
		Brand:          item.Brand,
		ModelNumber:    item.ModelNumber,
		SerialNumber:   item.SerialNumber,
		PurchaseDate:   data.FormatDate(item.PurchaseDate),
		WarrantyExpiry: data.FormatDate(item.WarrantyExpiry),
		Location:       item.Location,
		Cost:           data.FormatOptionalCents(item.CostCents),
		Notes:          item.Notes,
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
		Bathrooms:        formatFloat(profile.Bathrooms),
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
