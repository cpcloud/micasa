// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"time"

	"gorm.io/gorm"
)

const (
	ProjectStatusIdeating   = "ideating"
	ProjectStatusPlanned    = "planned"
	ProjectStatusQuoted     = "quoted"
	ProjectStatusInProgress = "underway"
	ProjectStatusDelayed    = "delayed"
	ProjectStatusCompleted  = "completed"
	ProjectStatusAbandoned  = "abandoned"
)

const (
	DeletionEntityProject     = "project"
	DeletionEntityQuote       = "quote"
	DeletionEntityMaintenance = "maintenance"
	DeletionEntityAppliance   = "appliance"
	DeletionEntityServiceLog  = "service_log"
	DeletionEntityVendor      = "vendor"
)

func ProjectStatuses() []string {
	return []string{
		ProjectStatusIdeating,
		ProjectStatusPlanned,
		ProjectStatusQuoted,
		ProjectStatusInProgress,
		ProjectStatusDelayed,
		ProjectStatusCompleted,
		ProjectStatusAbandoned,
	}
}

type HouseProfile struct {
	ID               uint `gorm:"primaryKey"`
	Nickname         string
	AddressLine1     string
	AddressLine2     string
	City             string
	State            string
	PostalCode       string
	YearBuilt        int     `gorm:"column:yr_blt"`
	SquareFeet       int     `gorm:"column:sq_ft"`
	LotSquareFeet    int     `gorm:"column:lot_sq_ft"`
	Bedrooms         int     `gorm:"column:br"`
	Bathrooms        float64 `gorm:"column:ba"`
	FoundationType   string  `gorm:"column:fnd_t"`
	WiringType       string  `gorm:"column:wir_t"`
	RoofType         string  `gorm:"column:roof_t"`
	ExteriorType     string  `gorm:"column:ext_t"`
	HeatingType      string  `gorm:"column:heat_t"`
	CoolingType      string  `gorm:"column:cool_t"`
	WaterSource      string  `gorm:"column:wtr_src"`
	SewerType        string  `gorm:"column:sewer_t"`
	ParkingType      string  `gorm:"column:park_t"`
	BasementType     string  `gorm:"column:base_t"`
	InsuranceCarrier string  `gorm:"column:ins_co"`
	InsurancePolicy  string  `gorm:"column:ins_pol"`
	InsuranceRenewal *time.Time
	PropertyTaxCents *int64 `gorm:"column:tax_ct"`
	HOAName          string `gorm:"column:hoa_nm"`
	HOAFeeCents      *int64 `gorm:"column:hoa_ct"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ProjectType struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Vendor struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex"`
	ContactName string
	Email       string
	Phone       string
	Website     string
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type Project struct {
	ID                uint `gorm:"primaryKey"`
	Title             string
	ProjectTypeID     uint
	ProjectType       ProjectType `gorm:"constraint:OnDelete:RESTRICT;"`
	Status            string
	Description       string
	StartDate         *time.Time
	EndDate           *time.Time
	BudgetCents       *int64 `gorm:"column:bud_ct"`
	ActualCents       *int64 `gorm:"column:act_ct"`
	PreferredVendorID *uint
	PreferredVendor   Vendor `gorm:"constraint:OnDelete:SET NULL;"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

type Quote struct {
	ID             uint `gorm:"primaryKey"`
	ProjectID      uint
	Project        Project `gorm:"constraint:OnDelete:RESTRICT;"`
	VendorID       uint
	Vendor         Vendor `gorm:"constraint:OnDelete:RESTRICT;"`
	TotalCents     int64  `gorm:"column:tot_ct"`
	LaborCents     *int64 `gorm:"column:lbr_ct"`
	MaterialsCents *int64 `gorm:"column:mat_ct"`
	OtherCents     *int64 `gorm:"column:oth_ct"`
	ReceivedDate   *time.Time
	Notes          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type MaintenanceCategory struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Appliance struct {
	ID             uint `gorm:"primaryKey"`
	Name           string
	Brand          string
	ModelNumber    string     `gorm:"column:model_no"`
	SerialNumber   string     `gorm:"column:serial_no"`
	PurchaseDate   *time.Time `gorm:"column:purch_dt"`
	WarrantyExpiry *time.Time `gorm:"column:warr_exp"`
	Location       string
	CostCents      *int64 `gorm:"column:cost_ct"`
	Notes          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type MaintenanceItem struct {
	ID             uint `gorm:"primaryKey"`
	Name           string
	CategoryID     uint
	Category       MaintenanceCategory `gorm:"constraint:OnDelete:RESTRICT;"`
	ApplianceID    *uint
	Appliance      Appliance  `gorm:"constraint:OnDelete:SET NULL;"`
	LastServicedAt *time.Time `gorm:"column:last_srv"`
	IntervalMonths int        `gorm:"column:int_mo"`
	ManualURL      string     `gorm:"column:man_url"`
	ManualText     string     `gorm:"column:man_txt"`
	Notes          string
	CostCents      *int64 `gorm:"column:cost_ct"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type ServiceLogEntry struct {
	ID                uint            `gorm:"primaryKey"`
	MaintenanceItemID uint            `gorm:"index"`
	MaintenanceItem   MaintenanceItem `gorm:"constraint:OnDelete:CASCADE;"`
	ServicedAt        time.Time
	VendorID          *uint
	Vendor            Vendor `gorm:"constraint:OnDelete:SET NULL;"`
	CostCents         *int64 `gorm:"column:cost_ct"`
	Notes             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

type DeletionRecord struct {
	ID         uint       `gorm:"primaryKey"`
	Entity     string     `gorm:"index:idx_entity_restored,priority:1"`
	TargetID   uint       `gorm:"index"`
	DeletedAt  time.Time  `gorm:"index"`
	RestoredAt *time.Time `gorm:"index:idx_entity_restored,priority:2"`
}
