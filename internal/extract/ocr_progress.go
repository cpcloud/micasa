// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package extract

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// OCRProgress reports incremental progress from OCRWithProgress.
type OCRProgress struct {
	Phase string // "rasterize" or "ocr"
	Page  int    // current page (1-indexed)
	Total int    // total pages (0 until known)
	Done  bool   // all phases finished
	Text  string // accumulated text (set on Done)
	TSV   []byte // accumulated TSV (set on Done)
	Err   error  // set on failure
}

// OCRWithProgress runs OCR with per-page progress updates sent on the
// returned channel. The channel closes when processing completes.
// Only PDF and image MIME types are supported; unsupported types produce
// a single Done message with empty text.
func OCRWithProgress(
	ctx context.Context,
	data []byte,
	mime string,
	maxPages int,
) <-chan OCRProgress {
	ch := make(chan OCRProgress, 8)
	go func() {
		defer close(ch)
		if IsImageMIME(mime) {
			ocrImageWithProgress(ctx, data, ch)
		} else {
			ocrPDFWithProgress(ctx, data, maxPages, ch)
		}
	}()
	return ch
}

// ocrImageWithProgress runs tesseract directly on an image file.
func ocrImageWithProgress(ctx context.Context, data []byte, ch chan<- OCRProgress) {
	if len(data) == 0 {
		ch <- OCRProgress{Done: true}
		return
	}

	tmpDir, err := os.MkdirTemp("", "micasa-ocr-*")
	if err != nil {
		ch <- OCRProgress{Err: fmt.Errorf("create temp dir: %w", err), Done: true}
		return
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // best-effort cleanup

	imgPath := filepath.Join(tmpDir, "input.png")
	if err := os.WriteFile(imgPath, data, 0o600); err != nil {
		ch <- OCRProgress{Err: fmt.Errorf("write temp image: %w", err), Done: true}
		return
	}

	select {
	case ch <- OCRProgress{Phase: "ocr", Page: 1, Total: 1}:
	case <-ctx.Done():
		ch <- OCRProgress{Err: ctx.Err(), Done: true}
		return
	}

	text, tsv, err := ocrImageFile(ctx, imgPath)
	if err != nil {
		ch <- OCRProgress{Err: fmt.Errorf("tesseract: %w", err), Done: true}
		return
	}

	ch <- OCRProgress{
		Done: true,
		Text: normalizeWhitespace(text),
		TSV:  tsv,
	}
}

func ocrPDFWithProgress(
	ctx context.Context,
	data []byte,
	maxPages int,
	ch chan<- OCRProgress,
) {
	if len(data) == 0 {
		ch <- OCRProgress{Done: true}
		return
	}
	if maxPages <= 0 {
		maxPages = DefaultMaxOCRPages
	}

	tmpDir, err := os.MkdirTemp("", "micasa-ocr-*")
	if err != nil {
		ch <- OCRProgress{Err: fmt.Errorf("create temp dir: %w", err), Done: true}
		return
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // best-effort cleanup

	pdfPath := filepath.Join(tmpDir, "input.pdf")
	if err := os.WriteFile(pdfPath, data, 0o600); err != nil {
		ch <- OCRProgress{Err: fmt.Errorf("write temp pdf: %w", err), Done: true}
		return
	}

	// Rasterize.
	outputPrefix := filepath.Join(tmpDir, "page")
	if err := rasterize(ctx, pdfPath, outputPrefix, maxPages); err != nil {
		ch <- OCRProgress{Err: fmt.Errorf("pdftoppm: %w", err), Done: true}
		return
	}

	images, err := filepath.Glob(outputPrefix + "*.png")
	if err != nil {
		ch <- OCRProgress{Err: fmt.Errorf("glob page images: %w", err), Done: true}
		return
	}
	sort.Strings(images)

	if len(images) == 0 {
		ch <- OCRProgress{Done: true}
		return
	}

	total := len(images)

	// Send rasterize complete.
	select {
	case ch <- OCRProgress{Phase: "rasterize", Page: total, Total: total}:
	case <-ctx.Done():
		ch <- OCRProgress{Err: ctx.Err(), Done: true}
		return
	}

	// OCR each page.
	var allText bytes.Buffer
	var allTSV bytes.Buffer
	headerWritten := false

	for i, img := range images {
		if ctx.Err() != nil {
			ch <- OCRProgress{Err: ctx.Err(), Done: true}
			return
		}

		select {
		case ch <- OCRProgress{Phase: "ocr", Page: i + 1, Total: total}:
		case <-ctx.Done():
			ch <- OCRProgress{Err: ctx.Err(), Done: true}
			return
		}

		pageText, pageTSV, ocrErr := ocrImageFile(ctx, img)
		if ocrErr != nil {
			continue // skip pages that fail
		}
		if pageText != "" {
			if allText.Len() > 0 {
				allText.WriteString("\n\n")
			}
			allText.WriteString(pageText)
		}
		if len(pageTSV) > 0 {
			lines := bytes.SplitN(pageTSV, []byte("\n"), 2)
			if !headerWritten {
				allTSV.Write(pageTSV)
				headerWritten = true
			} else if len(lines) > 1 {
				allTSV.Write(lines[1])
			}
		}
	}

	ch <- OCRProgress{
		Done: true,
		Text: normalizeWhitespace(allText.String()),
		TSV:  allTSV.Bytes(),
	}
}

// rasterize calls pdftoppm to convert PDF pages to PNG images.
func rasterize(ctx context.Context, pdfPath, outputPrefix string, maxPages int) error {
	args := []string{
		"-png",
		"-r", "300",
		"-l", fmt.Sprintf("%d", maxPages),
		pdfPath,
		outputPrefix,
	}
	cmd := exec.CommandContext( //nolint:gosec // args are constructed internally
		ctx,
		"pdftoppm",
		args...,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(stderr.String()), err)
	}
	return nil
}
