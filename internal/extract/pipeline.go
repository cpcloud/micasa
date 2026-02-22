// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package extract

import (
	"context"
	"fmt"
	"time"

	"github.com/cpcloud/micasa/internal/llm"
)

// Pipeline orchestrates the document extraction layers: text extraction,
// OCR, and LLM-powered structured extraction. Each layer is independent
// and gracefully degrades when its dependencies are unavailable.
type Pipeline struct {
	LLMClient     *llm.Client   // nil = skip LLM extraction
	MaxOCRPages   int           // 0 = DefaultMaxOCRPages
	TextTimeout   time.Duration // 0 = DefaultTextTimeout
	EntityContext EntityContext // existing entities for LLM matching
}

// Result holds the output of a pipeline run.
type Result struct {
	ExtractedText string // best available text (pdftotext or OCR, merged)
	PdfText       string // raw pdftotext output (PDFs only)
	OCRText       string // raw OCR output (scanned PDFs and images)
	OCRData       []byte
	Hints         *ExtractionHints // nil if LLM unavailable or failed
	OCRUsed       bool
	LLMUsed       bool
	Err           error // non-fatal extraction error; document still saves
}

// Run executes the extraction pipeline on the given document data.
// It never returns a Go error -- all failures are captured in Result.Err
// so the caller can save the document regardless.
func (p *Pipeline) Run(
	ctx context.Context,
	data []byte,
	filename string,
	mime string,
) *Result {
	r := &Result{}
	if len(data) == 0 {
		return r
	}

	maxPages := p.MaxOCRPages
	if maxPages <= 0 {
		maxPages = DefaultMaxOCRPages
	}

	isPDF := mime == "application/pdf"
	isImage := IsImageMIME(mime)

	// Layer 1: text extraction (PDFs and text files).
	text, textErr := ExtractText(data, mime, p.TextTimeout)
	if textErr != nil {
		r.Err = fmt.Errorf("text extraction: %w", textErr)
	} else {
		r.ExtractedText = text
		if isPDF {
			r.PdfText = text
		}
	}

	// Layer 2: OCR. For PDFs, always run (captures scanned pages that
	// pdftotext misses). For images, always run. For text files, skip.
	needsOCR := isPDF || isImage
	if needsOCR {
		ocrText, ocrData, ocrErr := p.tryOCR(ctx, data, mime, maxPages)
		if ocrErr != nil {
			r.Err = fmt.Errorf("ocr: %w", ocrErr)
		} else if ocrText != "" {
			r.OCRText = ocrText
			r.OCRData = ocrData
			r.OCRUsed = true
			// For images (no pdftotext), OCR is the only text source.
			if r.ExtractedText == "" {
				r.ExtractedText = ocrText
			}
		}
	}

	// Layer 3: LLM extraction if client configured and any text available.
	if p.LLMClient != nil && (r.PdfText != "" || r.OCRText != "" || r.ExtractedText != "") {
		hints, llmErr := p.extractWithLLM(
			ctx,
			r.PdfText,
			r.OCRText,
			r.ExtractedText,
			filename,
			mime,
			int64(len(data)),
		)
		if llmErr != nil {
			r.Err = fmt.Errorf("llm extraction: %w", llmErr)
		} else if hints != nil {
			r.Hints = hints
			r.LLMUsed = true
		}
	}

	return r
}

// tryOCR runs OCR if the required tools are available. Returns empty
// values (not an error) when tools are missing.
func (p *Pipeline) tryOCR(
	ctx context.Context,
	data []byte,
	mime string,
	maxPages int,
) (string, []byte, error) {
	isPDF := mime == "application/pdf"
	isImage := IsImageMIME(mime)

	if isPDF && !OCRAvailable() {
		return "", nil, nil
	}
	if isImage && !ImageOCRAvailable() {
		return "", nil, nil
	}
	if !isPDF && !isImage {
		return "", nil, nil
	}

	return OCR(ctx, data, mime, maxPages)
}

// extractWithLLM runs the LLM extraction model on both text sources.
func (p *Pipeline) extractWithLLM(
	ctx context.Context,
	pdfText string,
	ocrText string,
	extractedText string,
	filename string,
	mime string,
	sizeBytes int64,
) (*ExtractionHints, error) {
	messages := BuildExtractionPrompt(ExtractionPromptInput{
		Filename:  filename,
		MIME:      mime,
		SizeBytes: sizeBytes,
		Entities:  p.EntityContext,
		PdfText:   pdfText,
		OCRText:   ocrText,
		Text:      extractedText,
	})

	raw, err := p.LLMClient.ChatComplete(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm chat: %w", err)
	}

	hints, err := ParseExtractionResponse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse llm response: %w", err)
	}

	return &hints, nil
}
