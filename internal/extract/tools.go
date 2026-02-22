// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package extract

import (
	"os/exec"
	"sync"
)

var (
	tesseractOnce  sync.Once
	tesseractFound bool
	pdftoppmOnce   sync.Once
	pdftoppmFound  bool
	pdftotextOnce  sync.Once
	pdftotextFound bool
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

// HasPDFToPPM reports whether the pdftoppm binary (from poppler-utils)
// is on PATH. The result is cached for the process lifetime.
func HasPDFToPPM() bool {
	pdftoppmOnce.Do(func() {
		_, err := exec.LookPath("pdftoppm")
		pdftoppmFound = err == nil
	})
	return pdftoppmFound
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

// OCRAvailable reports whether both tesseract and pdftoppm are available,
// which is the minimum needed to OCR scanned PDFs.
func OCRAvailable() bool {
	return HasTesseract() && HasPDFToPPM()
}

// ImageOCRAvailable reports whether tesseract is available for direct
// image OCR (no pdftoppm needed for image files).
func ImageOCRAvailable() bool {
	return HasTesseract()
}
