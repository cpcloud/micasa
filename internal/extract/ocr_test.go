// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package extract

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextFromTSV(t *testing.T) {
	// Simulated tesseract TSV output with header + data rows.
	// Columns: level page_num block_num par_num line_num word_num left top width height conf text
	tsv := []byte(
		"level\tpage_num\tblock_num\tpar_num\tline_num\tword_num\tleft\ttop\twidth\theight\tconf\ttext\n" +
			"5\t1\t1\t1\t1\t1\t100\t200\t50\t12\t96\tHello\n" +
			"5\t1\t1\t1\t1\t2\t160\t200\t50\t12\t95\tworld\n" +
			"5\t1\t1\t1\t2\t1\t100\t220\t50\t12\t94\tSecond\n" +
			"5\t1\t1\t1\t2\t2\t160\t220\t50\t12\t93\tline\n" +
			"5\t1\t2\t1\t1\t1\t100\t300\t80\t12\t92\tNew\n" +
			"5\t1\t2\t1\t1\t2\t190\t300\t80\t12\t91\tblock\n",
	)

	text := textFromTSV(tsv)
	assert.Equal(t, "Hello world\nSecond line\n\nNew block", text)
}

func TestTextFromTSV_Empty(t *testing.T) {
	assert.Empty(t, textFromTSV(nil))
	assert.Empty(t, textFromTSV([]byte("")))
	assert.Empty(t, textFromTSV([]byte("header\n")))
}

func TestTextFromTSV_EmptyWords(t *testing.T) {
	tsv := []byte(
		"level\tpage_num\tblock_num\tpar_num\tline_num\tword_num\tleft\ttop\twidth\theight\tconf\ttext\n" +
			"5\t1\t1\t1\t1\t1\t100\t200\t50\t12\t96\t\n" +
			"5\t1\t1\t1\t1\t2\t160\t200\t50\t12\t95\tword\n",
	)
	text := textFromTSV(tsv)
	assert.Equal(t, "word", text)
}

func TestTextFromTSV_ParagraphBreaks(t *testing.T) {
	tsv := []byte(
		"level\tpage_num\tblock_num\tpar_num\tline_num\tword_num\tleft\ttop\twidth\theight\tconf\ttext\n" +
			"5\t1\t1\t1\t1\t1\t100\t200\t50\t12\t96\tPar1\n" +
			"5\t1\t1\t2\t1\t1\t100\t250\t50\t12\t95\tPar2\n",
	)
	text := textFromTSV(tsv)
	assert.Equal(t, "Par1\n\nPar2", text)
}

func TestAtoi(t *testing.T) {
	tests := []struct {
		input  string
		expect int
	}{
		{"0", 0},
		{"1", 1},
		{"42", 42},
		{"100", 100},
		{"abc", 0},
		{"", 0},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expect, atoi([]byte(tt.input)), "input: %q", tt.input)
	}
}

func TestIsImageMIME(t *testing.T) {
	assert.True(t, IsImageMIME("image/png"))
	assert.True(t, IsImageMIME("image/jpeg"))
	assert.True(t, IsImageMIME("image/tiff"))
	assert.True(t, IsImageMIME("image/bmp"))
	assert.True(t, IsImageMIME("image/webp"))
	assert.False(t, IsImageMIME("image/svg+xml"))
	assert.False(t, IsImageMIME("application/pdf"))
	assert.False(t, IsImageMIME("text/plain"))
}

func TestOCR_UnsupportedMIME(t *testing.T) {
	text, tsv, err := OCR(context.Background(), []byte("data"), "application/json", 20)
	require.NoError(t, err)
	assert.Empty(t, text)
	assert.Empty(t, tsv)
}

func TestOCR_EmptyData(t *testing.T) {
	text, tsv, err := OCR(context.Background(), nil, "application/pdf", 20)
	require.NoError(t, err)
	assert.Empty(t, text)
	assert.Empty(t, tsv)
}

func TestOCR_PDF_Integration(t *testing.T) {
	if !OCRAvailable() {
		t.Skip("tesseract and/or pdftoppm not available")
	}

	pdfPath := filepath.Join("testdata", "sample.pdf")
	data, err := os.ReadFile(pdfPath) //nolint:gosec // test fixture path
	if err != nil {
		t.Skipf("test fixture not found: %s", pdfPath)
	}

	text, tsv, err := OCR(context.Background(), data, "application/pdf", 20)
	require.NoError(t, err)
	// The sample PDF has digital text, so tesseract should find something.
	assert.NotEmpty(t, text)
	assert.NotEmpty(t, tsv)
	assert.Contains(t, text, "Invoice")
}

func TestOCR_Image_Integration(t *testing.T) {
	if !ImageOCRAvailable() {
		t.Skip("tesseract not available")
	}

	// Create a simple test image with text using ImageMagick if available.
	imgPath := filepath.Join("testdata", "sample-text.png")
	if _, err := os.Stat(imgPath); err != nil {
		t.Skipf("test image fixture not found: %s", imgPath)
	}

	data, err := os.ReadFile(imgPath) //nolint:gosec // test fixture path
	require.NoError(t, err)

	text, tsv, err := OCR(context.Background(), data, "image/png", 20)
	require.NoError(t, err)
	assert.NotEmpty(t, text)
	assert.NotEmpty(t, tsv)
}
