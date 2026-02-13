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
	YearBuilt        int
	SquareFeet       int
	LotSquareFeet    int
	Bedrooms         int
	Bathrooms        float64
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
	InsuranceRenewal *time.Time
	PropertyTaxCents *int64
	HOAName          string
	HOAFeeCents      *int64
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
	BudgetCents       *int64
	ActualCents       *int64
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
	TotalCents     int64
	LaborCents     *int64
	MaterialsCents *int64
	OtherCents     *int64
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
	ModelNumber    string
	SerialNumber   string
	PurchaseDate   *time.Time
	WarrantyExpiry *time.Time
	Location       string
	CostCents      *int64
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
	Appliance      Appliance `gorm:"constraint:OnDelete:SET NULL;"`
	LastServicedAt *time.Time
	IntervalMonths int
	ManualURL      string
	ManualText     string
	Notes          string
	CostCents      *int64
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
	CostCents         *int64
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
