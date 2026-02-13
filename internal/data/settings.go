// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Setting is a simple key-value store for app preferences that persist
// across sessions (e.g. last-used LLM model). Stored in SQLite so a
// single "cp micasa.db backup.db" captures everything.
type Setting struct {
	Key       string `gorm:"primaryKey"`
	Value     string
	UpdatedAt time.Time
}

// ChatInput stores a single chat prompt for cross-session history.
// Ordered by creation time, newest last.
type ChatInput struct {
	ID        uint   `gorm:"primaryKey"`
	Input     string `gorm:"not null"`
	CreatedAt time.Time
}

const (
	settingLLMModel = "llm.model"

	// chatHistoryMax is the maximum number of chat inputs retained.
	chatHistoryMax = 200
)

// GetSetting retrieves a setting by key. Returns ("", nil) if not found.
func (s *Store) GetSetting(key string) (string, error) {
	var setting Setting
	err := s.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return setting.Value, nil
}

// PutSetting upserts a setting.
func (s *Store) PutSetting(key, value string) error {
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&Setting{Key: key, Value: value, UpdatedAt: time.Now()}).Error
}

// GetLastModel returns the persisted LLM model name, or "" if none.
func (s *Store) GetLastModel() (string, error) {
	return s.GetSetting(settingLLMModel)
}

// PutLastModel persists the LLM model name.
func (s *Store) PutLastModel(model string) error {
	return s.PutSetting(settingLLMModel, model)
}

// AppendChatInput adds a prompt to the persistent history, deduplicating
// consecutive repeats. Trims old entries beyond chatHistoryMax.
func (s *Store) AppendChatInput(input string) error {
	// Deduplicate: skip if the most recent entry matches.
	var last ChatInput
	if err := s.db.Order("id DESC").First(&last).Error; err == nil {
		if last.Input == input {
			return nil
		}
	}

	if err := s.db.Create(&ChatInput{Input: input}).Error; err != nil {
		return err
	}

	// Trim old entries.
	var count int64
	s.db.Model(&ChatInput{}).Count(&count)
	if count > chatHistoryMax {
		excess := count - chatHistoryMax
		// Delete the oldest N rows.
		s.db.Exec(
			"DELETE FROM chat_inputs WHERE id IN (SELECT id FROM chat_inputs ORDER BY id ASC LIMIT ?)",
			excess,
		)
	}
	return nil
}

// LoadChatHistory returns all persisted chat inputs, oldest first.
func (s *Store) LoadChatHistory() ([]string, error) {
	var entries []ChatInput
	if err := s.db.Order("id ASC").Find(&entries).Error; err != nil {
		return nil, err
	}
	result := make([]string, len(entries))
	for i, e := range entries {
		result[i] = e.Input
	}
	return result, nil
}
