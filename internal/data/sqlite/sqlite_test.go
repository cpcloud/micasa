// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

// Based on github.com/glebarez/sqlite v1.11.0.
// Original code copyright (c) 2013-NOW Jinzhu <wosmvp@gmail.com>,
// licensed under the MIT License. See LICENSE-glebarez-sqlite for the
// full MIT text. Inlined because the upstream package is unmaintained.

package sqlite

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	modernsqlite "modernc.org/sqlite"
)

const inMemoryDSN = "file:testdatabase?mode=memory&cache=shared"

func TestDialector(t *testing.T) {
	const customDriverName = "test_custom_driver"

	sql.Register(customDriverName, &modernsqlite.Driver{})

	tests := []struct {
		description  string
		dialector    *Dialector
		openSuccess  bool
		query        string
		querySuccess bool
	}{
		{
			description:  "default_driver",
			dialector:    &Dialector{DSN: inMemoryDSN},
			openSuccess:  true,
			query:        "SELECT 1",
			querySuccess: true,
		},
		{
			description: "explicit_default_driver",
			dialector: &Dialector{
				DriverName: DriverName,
				DSN:        inMemoryDSN,
			},
			openSuccess:  true,
			query:        "SELECT 1",
			querySuccess: true,
		},
		{
			description: "bad_driver",
			dialector: &Dialector{
				DriverName: "not-a-real-driver",
				DSN:        inMemoryDSN,
			},
			openSuccess: false,
		},
		{
			description: "custom_driver",
			dialector: &Dialector{
				DriverName: customDriverName,
				DSN:        inMemoryDSN,
			},
			openSuccess:  true,
			query:        "SELECT 1",
			querySuccess: true,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d/%s", i, tt.description), func(t *testing.T) {
			db, err := gorm.Open(tt.dialector, &gorm.Config{})
			if !tt.openSuccess {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, db)

			if tt.query != "" {
				err = db.Exec(tt.query).Error
				if !tt.querySuccess {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
			}
		})
	}
}

func TestErrorTranslator(t *testing.T) {
	type Article struct {
		ArticleNumber string `gorm:"unique"`
	}

	db, err := gorm.Open(&Dialector{DSN: inMemoryDSN}, &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})
	require.NoError(t, err)
	require.NotNil(t, db)

	require.NoError(t, db.AutoMigrate(&Article{}))

	err = db.Create(&Article{ArticleNumber: "A00000XX"}).Error
	require.NoError(t, err)

	err = db.Create(&Article{ArticleNumber: "A00000XX"}).Error
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrDuplicatedKey)
}

func TestSQLiteVersion(t *testing.T) {
	db, err := sql.Open(DriverName, ":memory:")
	require.NoError(t, err)

	var version string
	require.NoError(t, db.QueryRow("select sqlite_version()").Scan(&version))
	assert.NotEmpty(t, version)
	t.Logf("SQLite version: %s", version)
}

func TestCompareVersion(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   int
	}{
		{"3.35.0", "3.35.0", 0},
		{"3.35.1", "3.35.0", 1},
		{"3.35.0", "3.35.1", -1},
		{"3.45.0", "3.35.0", 1},
		{"3.9.0", "3.35.0", -1},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tt.v1, tt.v2), func(t *testing.T) {
			assert.Equal(t, tt.want, compareVersion(tt.v1, tt.v2))
		})
	}
}
