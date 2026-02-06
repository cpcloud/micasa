package data

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Store struct {
	db *gorm.DB
}

func Open(path string) (*Store, error) {
	db, err := gorm.Open(
		sqlite.Open(path),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) AutoMigrate() error {
	return s.db.AutoMigrate(
		&HouseProfile{},
		&ProjectType{},
		&Vendor{},
		&Project{},
		&Quote{},
		&MaintenanceCategory{},
		&Appliance{},
		&MaintenanceItem{},
		&DeletionRecord{},
	)
}

func (s *Store) SeedDefaults() error {
	if err := s.seedProjectTypes(); err != nil {
		return err
	}
	if err := s.seedMaintenanceCategories(); err != nil {
		return err
	}
	return nil
}

// SeedDemoData populates the database with plausible sample data for testing.
// It is idempotent: if a house profile already exists, it returns immediately.
func (s *Store) SeedDemoData() error {
	var count int64
	if err := s.db.Model(&HouseProfile{}).Count(&count).Error; err != nil {
		return fmt.Errorf("check existing data: %w", err)
	}
	if count > 0 {
		return nil
	}

	// Helper for pointer values.
	ptr := func(v int64) *int64 { return &v }
	ptrTime := func(y, m, d int) *time.Time {
		t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
		return &t
	}

	// House profile
	house := HouseProfile{
		Nickname:         "Elm Street",
		AddressLine1:     "742 Elm Street",
		City:             "Springfield",
		State:            "IL",
		PostalCode:       "62704",
		YearBuilt:        1987,
		SquareFeet:       2400,
		LotSquareFeet:    8500,
		Bedrooms:         4,
		Bathrooms:        2.5,
		FoundationType:   "Poured Concrete",
		WiringType:       "Copper",
		RoofType:         "Asphalt Shingle",
		ExteriorType:     "Vinyl Siding",
		HeatingType:      "Forced Air Gas",
		CoolingType:      "Central AC",
		WaterSource:      "Municipal",
		SewerType:        "Municipal",
		ParkingType:      "Attached 2-Car",
		BasementType:     "Finished",
		InsuranceCarrier: "Acme Insurance",
		InsurancePolicy:  "HO-00-0000000",
		InsuranceRenewal: ptrTime(2026, 8, 15),
		PropertyTaxCents: ptr(485000),
		HOAName:          "Elm Street HOA",
		HOAFeeCents:      ptr(15000),
	}
	if err := s.db.Create(&house).Error; err != nil {
		return fmt.Errorf("seed house: %w", err)
	}

	// Look up seeded project types and maintenance categories by name.
	typeID := func(name string) uint {
		var pt ProjectType
		s.db.Where("name = ?", name).First(&pt)
		return pt.ID
	}
	catID := func(name string) uint {
		var mc MaintenanceCategory
		s.db.Where("name = ?", name).First(&mc)
		return mc.ID
	}

	// Vendors (all fictitious: 555 numbers, example.com emails)
	vendors := []Vendor{
		{
			Name:        "Hartley Plumbing",
			ContactName: "Dave Hartley",
			Phone:       "555-555-0142",
			Email:       "dave@example.com",
		},
		{
			Name:        "Greenleaf Landscaping",
			ContactName: "Maria Santos",
			Phone:       "555-555-0198",
			Email:       "maria@example.com",
		},
		{
			Name:        "Sparks Electric",
			ContactName: "Tom Nguyen",
			Phone:       "555-555-0231",
			Email:       "tom@example.com",
		},
		{
			Name:        "Premier Roofing",
			ContactName: "Jake Miller",
			Phone:       "555-555-0307",
			Website:     "https://example.com/premier",
		},
		{
			Name:        "Central HVAC Services",
			ContactName: "Lisa Park",
			Phone:       "555-555-0415",
			Email:       "lisa@example.com",
		},
		{Name: "Bright Window Co", ContactName: "Sam Torres", Phone: "555-555-0523"},
	}
	for i := range vendors {
		if err := s.db.Create(&vendors[i]).Error; err != nil {
			return fmt.Errorf("seed vendor %s: %w", vendors[i].Name, err)
		}
	}

	// Projects
	projects := []Project{
		{
			Title:             "Replace water heater",
			ProjectTypeID:     typeID("Plumbing"),
			Status:            ProjectStatusCompleted,
			Description:       "50-gal tank water heater failed, replacing with tankless unit",
			StartDate:         ptrTime(2025, 3, 10),
			EndDate:           ptrTime(2025, 3, 14),
			BudgetCents:       ptr(350000),
			ActualCents:       ptr(328500),
			PreferredVendorID: &vendors[0].ID,
		},
		{
			Title:             "Front yard landscaping",
			ProjectTypeID:     typeID("Landscaping"),
			Status:            ProjectStatusInProgress,
			Description:       "Remove dead bushes, plant native perennials, add mulch beds",
			StartDate:         ptrTime(2025, 10, 1),
			BudgetCents:       ptr(450000),
			PreferredVendorID: &vendors[1].ID,
		},
		{
			Title:         "Upgrade electrical panel",
			ProjectTypeID: typeID("Electrical"),
			Status:        ProjectStatusQuoted,
			Description:   "Upgrade from 100A to 200A service for EV charger support",
			BudgetCents:   ptr(600000),
		},
		{
			Title:         "Replace back deck boards",
			ProjectTypeID: typeID("Exterior"),
			Status:        ProjectStatusPlanned,
			Description:   "Composite decking to replace rotting pressure-treated lumber",
			BudgetCents:   ptr(800000),
		},
		{
			Title:             "Kitchen faucet replacement",
			ProjectTypeID:     typeID("Plumbing"),
			Status:            ProjectStatusCompleted,
			Description:       "Pulldown faucet, brushed nickel finish",
			StartDate:         ptrTime(2025, 6, 20),
			EndDate:           ptrTime(2025, 6, 20),
			BudgetCents:       ptr(45000),
			ActualCents:       ptr(42000),
			PreferredVendorID: &vendors[0].ID,
		},
		{
			Title:         "Paint master bedroom",
			ProjectTypeID: typeID("Painting"),
			Status:        ProjectStatusIdeating,
			Description:   "Walls in off-white, trim in bright white, two coats",
		},
		{
			Title:         "Add attic insulation",
			ProjectTypeID: typeID("Exterior"),
			Status:        ProjectStatusDelayed,
			Description:   "Blow-in cellulose, waiting on contractor availability",
			BudgetCents:   ptr(200000),
		},
		{
			Title:         "Convert garage to gym",
			ProjectTypeID: typeID("Exterior"),
			Status:        ProjectStatusAbandoned,
			Description:   "Decided against it; would lose parking",
		},
	}
	for i := range projects {
		if err := s.db.Create(&projects[i]).Error; err != nil {
			return fmt.Errorf("seed project %s: %w", projects[i].Title, err)
		}
	}

	// Quotes
	quotes := []struct {
		projectIdx int
		vendorIdx  int
		total      int64
		labor      *int64
		materials  *int64
		received   *time.Time
		notes      string
	}{
		{2, 2, 575000, ptr(375000), ptr(200000), ptrTime(2025, 11, 5), "Includes permit fees"},
		{
			2,
			0,
			620000,
			ptr(400000),
			ptr(220000),
			ptrTime(2025, 11, 12),
			"Would need to subcontract",
		},
		{
			3,
			4,
			780000,
			ptr(450000),
			ptr(330000),
			ptrTime(2025, 12, 1),
			"Composite boards, includes railing",
		},
		{1, 1, 420000, ptr(280000), ptr(140000), ptrTime(2025, 9, 15), "Phase 1 front beds only"},
		{1, 1, 680000, ptr(400000), ptr(280000), ptrTime(2025, 9, 15), "Full front and side yards"},
	}
	for _, q := range quotes {
		quote := Quote{
			ProjectID:      projects[q.projectIdx].ID,
			VendorID:       vendors[q.vendorIdx].ID,
			TotalCents:     q.total,
			LaborCents:     q.labor,
			MaterialsCents: q.materials,
			ReceivedDate:   q.received,
			Notes:          q.notes,
		}
		if err := s.db.Create(&quote).Error; err != nil {
			return fmt.Errorf("seed quote: %w", err)
		}
	}

	// Appliances (all fictitious brands)
	appliances := []Appliance{
		{
			Name:           "Kitchen Refrigerator",
			Brand:          "Frostline",
			ModelNumber:    "FR-2400X",
			SerialNumber:   "FL-00-849271",
			Location:       "Kitchen",
			PurchaseDate:   ptrTime(2020, 6, 15),
			WarrantyExpiry: ptrTime(2025, 6, 15),
			CostCents:      ptr(189900),
		},
		{
			Name:           "Washer",
			Brand:          "CleanWave",
			ModelNumber:    "CW-850F",
			SerialNumber:   "CW-00-331045",
			Location:       "Laundry Room",
			PurchaseDate:   ptrTime(2021, 3, 10),
			WarrantyExpiry: ptrTime(2024, 3, 10),
			CostCents:      ptr(89900),
		},
		{
			Name:           "Dryer",
			Brand:          "CleanWave",
			ModelNumber:    "CW-850D",
			SerialNumber:   "CW-00-331046",
			Location:       "Laundry Room",
			PurchaseDate:   ptrTime(2021, 3, 10),
			WarrantyExpiry: ptrTime(2024, 3, 10),
			CostCents:      ptr(79900),
		},
		{
			Name:         "Dishwasher",
			Brand:        "Frostline",
			ModelNumber:  "FR-DW550",
			SerialNumber: "FL-00-220318",
			Location:     "Kitchen",
			PurchaseDate: ptrTime(2019, 11, 20),
			CostCents:    ptr(64900),
		},
		{
			Name:           "Tankless Water Heater",
			Brand:          "AquaMax",
			ModelNumber:    "AM-TL200",
			SerialNumber:   "AQ-00-558102",
			Location:       "Utility Closet",
			PurchaseDate:   ptrTime(2025, 3, 14),
			WarrantyExpiry: ptrTime(2035, 3, 14),
			CostCents:      ptr(328500),
			Notes:          "Installed during water heater replacement project",
		},
		{
			Name:           "Central AC / Furnace",
			Brand:          "AirComfort",
			ModelNumber:    "AC-4800DX",
			SerialNumber:   "AF-00-994520",
			Location:       "Basement",
			PurchaseDate:   ptrTime(2018, 8, 1),
			WarrantyExpiry: ptrTime(2028, 8, 1),
			CostCents:      ptr(650000),
		},
		{
			Name:         "Garage Door Opener",
			Brand:        "LiftRight",
			ModelNumber:  "LR-3400",
			SerialNumber: "LR-00-117890",
			Location:     "Garage",
			PurchaseDate: ptrTime(2017, 5, 22),
			CostCents:    ptr(35000),
		},
	}
	for i := range appliances {
		if err := s.db.Create(&appliances[i]).Error; err != nil {
			return fmt.Errorf("seed appliance %s: %w", appliances[i].Name, err)
		}
	}

	// Maintenance items
	maintItems := []MaintenanceItem{
		{
			Name:           "HVAC filter replacement",
			CategoryID:     catID("HVAC"),
			LastServicedAt: ptrTime(2025, 12, 1),
			NextDueAt:      ptrTime(2026, 3, 1),
			IntervalMonths: 3,
			Notes:          "20x25x1 MERV 13, buy in bulk",
			CostCents:      ptr(2500),
		},
		{
			Name:           "Gutter cleaning",
			CategoryID:     catID("Exterior"),
			LastServicedAt: ptrTime(2025, 11, 10),
			NextDueAt:      ptrTime(2026, 5, 10),
			IntervalMonths: 6,
			Notes:          "Front and back, check downspout screens",
			CostCents:      ptr(15000),
		},
		{
			Name:           "Smoke detector batteries",
			CategoryID:     catID("Safety"),
			LastServicedAt: ptrTime(2025, 11, 1),
			NextDueAt:      ptrTime(2026, 11, 1),
			IntervalMonths: 12,
			Notes:          "6 detectors, use 9V lithium",
			CostCents:      ptr(3000),
		},
		{
			Name:           "Water softener salt",
			CategoryID:     catID("Plumbing"),
			LastServicedAt: ptrTime(2026, 1, 15),
			NextDueAt:      ptrTime(2026, 3, 15),
			IntervalMonths: 2,
			Notes:          "40lb bag solar salt",
			CostCents:      ptr(800),
		},
		{
			Name:           "Furnace annual inspection",
			CategoryID:     catID("HVAC"),
			LastServicedAt: ptrTime(2025, 9, 20),
			NextDueAt:      ptrTime(2026, 9, 20),
			IntervalMonths: 12,
			Notes:          "Central HVAC Services, schedule in August",
			CostCents:      ptr(18000),
		},
		{
			Name:           "Lawn mower blade sharpening",
			CategoryID:     catID("Landscaping"),
			LastServicedAt: ptrTime(2025, 4, 1),
			NextDueAt:      ptrTime(2026, 4, 1),
			IntervalMonths: 12,
			Notes:          "Or replace blade if nicked",
			CostCents:      ptr(2000),
		},
		{
			Name:           "Refrigerator coil cleaning",
			CategoryID:     catID("Appliance"),
			LastServicedAt: ptrTime(2025, 7, 1),
			NextDueAt:      ptrTime(2026, 1, 1),
			IntervalMonths: 6,
			ManualURL:      "https://example.com/fridge-manual",
			Notes:          "Pull out, vacuum coils underneath",
		},
		{
			Name:           "Sump pump test",
			CategoryID:     catID("Plumbing"),
			LastServicedAt: ptrTime(2025, 10, 1),
			NextDueAt:      ptrTime(2026, 4, 1),
			IntervalMonths: 6,
			Notes:          "Pour 5 gallons in pit, confirm auto-start and drain",
		},
	}
	for i := range maintItems {
		if err := s.db.Create(&maintItems[i]).Error; err != nil {
			return fmt.Errorf("seed maintenance %s: %w", maintItems[i].Name, err)
		}
	}

	// Link some maintenance items to appliances.
	appLinks := map[string]uint{
		"HVAC filter replacement":    appliances[5].ID,
		"Furnace annual inspection":  appliances[5].ID,
		"Refrigerator coil cleaning": appliances[0].ID,
	}
	for name, appID := range appLinks {
		s.db.Model(&MaintenanceItem{}).Where("name = ?", name).Update("appliance_id", appID)
	}

	return nil
}

func (s *Store) HouseProfile() (HouseProfile, error) {
	var profile HouseProfile
	err := s.db.First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return HouseProfile{}, gorm.ErrRecordNotFound
	}
	return profile, err
}

func (s *Store) CreateHouseProfile(profile HouseProfile) error {
	var count int64
	if err := s.db.Model(&HouseProfile{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count house profiles: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("house profile already exists")
	}
	return s.db.Create(&profile).Error
}

func (s *Store) UpdateHouseProfile(profile HouseProfile) error {
	var existing HouseProfile
	if err := s.db.First(&existing).Error; err != nil {
		return err
	}
	profile.ID = existing.ID
	profile.CreatedAt = existing.CreatedAt
	return s.db.Model(&existing).Select("*").Updates(profile).Error
}

func (s *Store) ProjectTypes() ([]ProjectType, error) {
	var types []ProjectType
	if err := s.db.Order("name").Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

func (s *Store) MaintenanceCategories() ([]MaintenanceCategory, error) {
	var categories []MaintenanceCategory
	if err := s.db.Order("name").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (s *Store) ListProjects(includeDeleted bool) ([]Project, error) {
	var projects []Project
	db := s.db.Preload("ProjectType").Preload("PreferredVendor")
	db = db.Order("updated_at desc")
	if includeDeleted {
		db = db.Unscoped()
	}
	if err := db.Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (s *Store) ListQuotes(includeDeleted bool) ([]Quote, error) {
	var quotes []Quote
	db := s.db.Preload("Vendor")
	db = db.Preload("Project", func(q *gorm.DB) *gorm.DB {
		return q.Unscoped().Preload("ProjectType")
	})
	db = db.Order("updated_at desc")
	if includeDeleted {
		db = db.Unscoped()
	}
	if err := db.Find(&quotes).Error; err != nil {
		return nil, err
	}
	return quotes, nil
}

func (s *Store) ListMaintenance(includeDeleted bool) ([]MaintenanceItem, error) {
	var items []MaintenanceItem
	db := s.db.Preload("Category")
	db = db.Preload("Appliance", func(q *gorm.DB) *gorm.DB {
		return q.Unscoped()
	})
	db = db.Order("updated_at desc")
	if includeDeleted {
		db = db.Unscoped()
	}
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) GetProject(id uint) (Project, error) {
	var project Project
	err := s.db.Preload("ProjectType").Preload("PreferredVendor").First(&project, id).Error
	return project, err
}

func (s *Store) CreateProject(project Project) error {
	return s.db.Create(&project).Error
}

func (s *Store) UpdateProject(project Project) error {
	return s.db.Model(&Project{}).Where("id = ?", project.ID).
		Select("*").
		Omit("id", "created_at", "deleted_at").
		Updates(project).Error
}

func (s *Store) GetQuote(id uint) (Quote, error) {
	var quote Quote
	err := s.db.Preload("Vendor").
		Preload("Project", func(q *gorm.DB) *gorm.DB {
			return q.Unscoped().Preload("ProjectType")
		}).
		First(&quote, id).Error
	return quote, err
}

func (s *Store) CreateQuote(quote Quote, vendor Vendor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		foundVendor, err := findOrCreateVendor(tx, vendor)
		if err != nil {
			return err
		}
		quote.VendorID = foundVendor.ID
		if err := tx.Create(&quote).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) UpdateQuote(quote Quote, vendor Vendor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		foundVendor, err := findOrCreateVendor(tx, vendor)
		if err != nil {
			return err
		}
		quote.VendorID = foundVendor.ID
		return tx.Model(&Quote{}).Where("id = ?", quote.ID).
			Select("*").
			Omit("id", "created_at", "deleted_at").
			Updates(quote).Error
	})
}

func (s *Store) GetMaintenance(id uint) (MaintenanceItem, error) {
	var item MaintenanceItem
	err := s.db.Preload("Category").
		Preload("Appliance", func(q *gorm.DB) *gorm.DB {
			return q.Unscoped()
		}).
		First(&item, id).Error
	return item, err
}

func (s *Store) CreateMaintenance(item MaintenanceItem) error {
	return s.db.Create(&item).Error
}

func (s *Store) UpdateMaintenance(item MaintenanceItem) error {
	return s.db.Model(&MaintenanceItem{}).Where("id = ?", item.ID).
		Select("*").
		Omit("id", "created_at", "deleted_at").
		Updates(item).Error
}

func (s *Store) ListAppliances(includeDeleted bool) ([]Appliance, error) {
	var items []Appliance
	db := s.db.Order("updated_at desc")
	if includeDeleted {
		db = db.Unscoped()
	}
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) GetAppliance(id uint) (Appliance, error) {
	var item Appliance
	err := s.db.First(&item, id).Error
	return item, err
}

func (s *Store) CreateAppliance(item Appliance) error {
	return s.db.Create(&item).Error
}

func (s *Store) UpdateAppliance(item Appliance) error {
	return s.db.Model(&Appliance{}).Where("id = ?", item.ID).
		Select("*").
		Omit("id", "created_at", "deleted_at").
		Updates(item).Error
}

func (s *Store) DeleteProject(id uint) error {
	return s.softDelete(&Project{}, DeletionEntityProject, id)
}

func (s *Store) DeleteQuote(id uint) error {
	return s.softDelete(&Quote{}, DeletionEntityQuote, id)
}

func (s *Store) DeleteMaintenance(id uint) error {
	return s.softDelete(&MaintenanceItem{}, DeletionEntityMaintenance, id)
}

func (s *Store) DeleteAppliance(id uint) error {
	return s.softDelete(&Appliance{}, DeletionEntityAppliance, id)
}

func (s *Store) RestoreProject(id uint) error {
	return s.restoreEntity(&Project{}, DeletionEntityProject, id)
}

func (s *Store) RestoreQuote(id uint) error {
	return s.restoreEntity(&Quote{}, DeletionEntityQuote, id)
}

func (s *Store) RestoreMaintenance(id uint) error {
	return s.restoreEntity(&MaintenanceItem{}, DeletionEntityMaintenance, id)
}

func (s *Store) RestoreAppliance(id uint) error {
	return s.restoreEntity(&Appliance{}, DeletionEntityAppliance, id)
}

func (s *Store) softDelete(model any, entity string, id uint) error {
	result := s.db.Delete(model, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return s.logDeletion(entity, id)
}

func (s *Store) restoreEntity(model any, entity string, id uint) error {
	if err := s.restoreByID(model, id); err != nil {
		return err
	}
	return s.markDeletionRestored(entity, id)
}

func (s *Store) LastDeletion(entity string) (DeletionRecord, error) {
	var record DeletionRecord
	err := s.db.
		Where("entity = ? AND restored_at IS NULL", entity).
		Order("id desc").
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DeletionRecord{}, gorm.ErrRecordNotFound
	}
	return record, err
}

func (s *Store) seedProjectTypes() error {
	types := []ProjectType{
		{Name: "Appliance"},
		{Name: "Electrical"},
		{Name: "Exterior"},
		{Name: "Flooring"},
		{Name: "HVAC"},
		{Name: "Landscaping"},
		{Name: "Painting"},
		{Name: "Plumbing"},
		{Name: "Remodel"},
		{Name: "Roof"},
		{Name: "Structural"},
		{Name: "Windows"},
	}
	for _, projectType := range types {
		if err := s.db.FirstOrCreate(&projectType, "name = ?", projectType.Name).
			Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) seedMaintenanceCategories() error {
	categories := []MaintenanceCategory{
		{Name: "Appliance"},
		{Name: "Electrical"},
		{Name: "Exterior"},
		{Name: "HVAC"},
		{Name: "Interior"},
		{Name: "Landscaping"},
		{Name: "Plumbing"},
		{Name: "Safety"},
		{Name: "Structural"},
	}
	for _, category := range categories {
		if err := s.db.FirstOrCreate(&category, "name = ?", category.Name).
			Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) restoreByID(model any, id uint) error {
	return s.db.Unscoped().Model(model).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

func (s *Store) logDeletion(entity string, id uint) error {
	record := DeletionRecord{
		Entity:    entity,
		TargetID:  id,
		DeletedAt: time.Now(),
	}
	return s.db.Create(&record).Error
}

func (s *Store) markDeletionRestored(entity string, id uint) error {
	restoredAt := time.Now()
	return s.db.Model(&DeletionRecord{}).
		Where("entity = ? AND target_id = ? AND restored_at IS NULL", entity, id).
		Update("restored_at", restoredAt).Error
}

func findOrCreateVendor(tx *gorm.DB, vendor Vendor) (Vendor, error) {
	if strings.TrimSpace(vendor.Name) == "" {
		return Vendor{}, fmt.Errorf("vendor name is required")
	}
	var existing Vendor
	err := tx.Where("name = ?", vendor.Name).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := tx.Create(&vendor).Error; err != nil {
			return Vendor{}, err
		}
		return vendor, nil
	}
	if err != nil {
		return Vendor{}, err
	}
	updates := map[string]any{}
	if vendor.ContactName != "" {
		updates["contact_name"] = vendor.ContactName
	}
	if vendor.Email != "" {
		updates["email"] = vendor.Email
	}
	if vendor.Phone != "" {
		updates["phone"] = vendor.Phone
	}
	if vendor.Website != "" {
		updates["website"] = vendor.Website
	}
	if vendor.Notes != "" {
		updates["notes"] = vendor.Notes
	}
	if len(updates) > 0 {
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return Vendor{}, err
		}
	}
	return existing, nil
}
