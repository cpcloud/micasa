// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ExtractDocument writes the document's BLOB content to the XDG cache
// directory and returns the resulting filesystem path. If the cached file
// already exists and has the expected size, the extraction is skipped.
func (s *Store) ExtractDocument(id uint) (string, error) {
	var doc Document
	err := s.db.Select("data", "file_name", "sha256", "size_bytes").
		First(&doc, id).Error
	if err != nil {
		return "", fmt.Errorf("load document content: %w", err)
	}
	if len(doc.Data) == 0 {
		return "", fmt.Errorf("document has no content")
	}

	cacheDir, err := DocumentCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve cache dir: %w", err)
	}

	name := doc.ChecksumSHA256 + "-" + filepath.Base(doc.FileName)
	cachePath := filepath.Join(cacheDir, name)

	// Cache hit: file exists with correct size. Touch the ModTime so the
	// TTL-based eviction in EvictStaleCache treats it as recently used.
	if info, statErr := os.Stat(cachePath); statErr == nil && info.Size() == doc.SizeBytes {
		now := time.Now()
		_ = os.Chtimes(
			cachePath,
			now,
			now,
		) // best-effort; stale ModTime just means earlier re-extraction
		return cachePath, nil
	}

	if err := os.WriteFile(cachePath, doc.Data, 0o600); err != nil {
		return "", fmt.Errorf("write cached document: %w", err)
	}
	return cachePath, nil
}

// EvictStaleCache removes cached document files from dir that haven't been
// modified in the given number of days. A ttlDays of 0 disables eviction.
// Returns the number of files removed and any error encountered while listing
// the directory (individual file removal errors are skipped).
func EvictStaleCache(dir string, ttlDays int) (int, error) {
	if ttlDays <= 0 || dir == "" {
		return 0, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil // cache dir doesn't exist yet; nothing to evict
		}
		return 0, fmt.Errorf("list cache dir: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -ttlDays)
	removed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if os.Remove(filepath.Join(dir, entry.Name())) == nil {
				removed++
			}
		}
	}
	return removed, nil
}
