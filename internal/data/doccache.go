// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Cache hit: file exists with correct size.
	if info, statErr := os.Stat(cachePath); statErr == nil && info.Size() == doc.SizeBytes {
		return cachePath, nil
	}

	if err := os.WriteFile(cachePath, doc.Data, 0o600); err != nil {
		return "", fmt.Errorf("write cached document: %w", err)
	}
	return cachePath, nil
}
