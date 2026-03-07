// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package extract

import (
	"os/exec"
	"sync"
)

var (
	tesseractOnce   sync.Once
	tesseractFound  bool
	pdftocairoOnce  sync.Once
	pdftocairoFound bool
	pdftotextOnce   sync.Once
	pdftotextFound  bool
	pdfinfoOnce     sync.Once
	pdfinfoFound    bool
)

// HasTesseract reports whether the tesseract binary is on PATH.
// The result is cached for the process lifetime.
func HasTesseract() bool {
	tesseractOnce.Do(func() {
		_, err := exec.LookPath("tesseract")
		tesseractFound = err == nil
	})
	return tesseractFound
}

// HasPDFToCairo reports whether the pdftocairo binary (from poppler-utils)
// is on PATH. The result is cached for the process lifetime.
func HasPDFToCairo() bool {
	pdftocairoOnce.Do(func() {
		_, err := exec.LookPath("pdftocairo")
		pdftocairoFound = err == nil
	})
	return pdftocairoFound
}

// HasPDFToText reports whether the pdftotext binary (from poppler-utils)
// is on PATH. The result is cached for the process lifetime.
func HasPDFToText() bool {
	pdftotextOnce.Do(func() {
		_, err := exec.LookPath("pdftotext")
		pdftotextFound = err == nil
	})
	return pdftotextFound
}

// HasPDFInfo reports whether the pdfinfo binary (from poppler-utils)
// is on PATH. The result is cached for the process lifetime.
func HasPDFInfo() bool {
	pdfinfoOnce.Do(func() {
		_, err := exec.LookPath("pdfinfo")
		pdfinfoFound = err == nil
	})
	return pdfinfoFound
}

// OCRAvailable reports whether tesseract and pdftocairo (with pdfinfo
// for page count discovery) are available.
func OCRAvailable() bool {
	return HasTesseract() && HasPDFToCairo() && HasPDFInfo()
}

// ImageOCRAvailable reports whether tesseract is available for direct
// image OCR (no PDF tools needed for image files).
func ImageOCRAvailable() bool {
	return HasTesseract()
}
