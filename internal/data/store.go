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
	db := s.db.Preload("Category").Order("updated_at desc")
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
	err := s.db.Preload("Category").First(&item, id).Error
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

func (s *Store) DeleteProject(id uint) error {
	result := s.db.Delete(&Project{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return s.logDeletion(DeletionEntityProject, id)
}

func (s *Store) DeleteQuote(id uint) error {
	result := s.db.Delete(&Quote{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return s.logDeletion(DeletionEntityQuote, id)
}

func (s *Store) DeleteMaintenance(id uint) error {
	result := s.db.Delete(&MaintenanceItem{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return s.logDeletion(DeletionEntityMaintenance, id)
}

func (s *Store) RestoreProject(id uint) error {
	if err := s.restoreByID(&Project{}, id); err != nil {
		return err
	}
	return s.markDeletionRestored(DeletionEntityProject, id)
}

func (s *Store) RestoreQuote(id uint) error {
	if err := s.restoreByID(&Quote{}, id); err != nil {
		return err
	}
	return s.markDeletionRestored(DeletionEntityQuote, id)
}

func (s *Store) RestoreMaintenance(id uint) error {
	if err := s.restoreByID(&MaintenanceItem{}, id); err != nil {
		return err
	}
	return s.markDeletionRestored(DeletionEntityMaintenance, id)
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
