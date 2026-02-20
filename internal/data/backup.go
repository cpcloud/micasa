// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"context"
	"fmt"

	"modernc.org/sqlite"
)

// backuper is the interface exposed by the modernc.org/sqlite driver
// connection for online backups.
type backuper interface {
	NewBackup(string) (*sqlite.Backup, error)
}

// Backup creates a consistent snapshot of the database at destPath using
// SQLite's Online Backup API. The destination must not already exist.
func (s *Store) Backup(ctx context.Context, destPath string) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("get underlying db: %w", err)
	}

	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	return conn.Raw(func(driverConn any) error {
		b, ok := driverConn.(backuper)
		if !ok {
			return fmt.Errorf(
				"SQLite driver does not support the backup API -- please report this as a bug",
			)
		}

		bck, err := b.NewBackup(destPath)
		if err != nil {
			return fmt.Errorf("init backup: %w", err)
		}

		// Step(-1) copies all remaining pages. The loop drains the
		// backup even if multiple iterations are needed.
		for more := true; more; {
			more, err = bck.Step(-1)
			if err != nil {
				_ = bck.Finish()
				return fmt.Errorf("backup step: %w", err)
			}
		}

		if err := bck.Finish(); err != nil {
			return fmt.Errorf("finish backup: %w", err)
		}
		return nil
	})
}
