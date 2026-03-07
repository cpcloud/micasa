<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Parallel Per-Page Rasterization with pdftocairo

Closes #640. Ref #639.

## Problem

The OCR pipeline rasterized PDF pages sequentially via a single `pdftoppm`
process. On large documents this was the primary bottleneck -- tesseract OCR
was already parallelized but had to wait for all pages to be rasterized first.

## Design

### Single-tool architecture

Replaced the old three-tool approach (`pdfimages` + `pdftohtml` + `pdftoppm`)
with a single tool: **pdftocairo** (Cairo renderer from poppler-utils).

- Better rendering quality than `pdftoppm` (Cairo handles transparency, color
  profiles, and complex vector graphics)
- Handles both raster and vector content (replacing both `pdftoppm` and
  `pdftohtml`)
- Supports per-page rendering with `-singlefile -f N -l N`
- Supports stdout output with `-` for zero-disk-IO piping

### Fused rasterize+OCR pipeline with process piping

Each page is processed by a goroutine that pipes `pdftocairo` stdout directly
into `tesseract` stdin -- no intermediate files touch disk:

```
goroutine per page (capped by semaphore at NumCPU):
  pdftocairo -png -r 300 -singlefile -f N -l N input.pdf - | tesseract stdin stdout tsv
```

Benefits:
- No waiting for all pages to rasterize before OCR starts
- Each CPU core is fully utilized (render + OCR interleaved)
- Zero intermediate disk I/O (works correctly even without tmpfs)
- Progress reporting works per-page from the start
- Simpler code (no multi-tool acquisition, no image merging)

### Page count discovery

Uses `pdfinfo` to get the total page count upfront. This drives the parallel
dispatch and progress reporting.

## Changes

### `internal/extract/tools.go`
- Removed `HasPDFToPPM()`, `HasPDFToHTML()`, `HasPDFImages()` and their
  `sync.Once` vars
- Added `HasPDFToCairo()` and `HasPDFInfo()`
- Simplified `OCRAvailable()`: `HasTesseract() && HasPDFToCairo() && HasPDFInfo()`

### `internal/extract/ocr.go`
- Removed `extractPDFImages()`, `extractPDFToHTMLImages()`,
  `acquireImages()`, `mergeAcquiredImages()`, `ocrPagesParallel()`,
  `isOCRWorthy()`, `rasterize()`
- Added `pdfPageCount()`: runs `pdfinfo` to get page count
- Added `ocrPage()`: fused `pdftocairo | tesseract` pipe per page
- Added `ocrPDFPages()`: semaphore-capped worker pool running `ocrPage`
- Simplified `ocrPDF()` to use the fused pipeline directly

### `internal/extract/ocr_progress.go`
- Removed `rasterize()`, `snapshotToolStates()`, multi-tool state tracking
- Simplified `ocrPDFWithProgress()` to single-tool progress with
  incrementing page count

### `internal/app/extraction.go`
- Added in-progress count display for running tools

### Tests
- All updated to reflect single-tool architecture
- `rasterizePage` tests replaced with `ocrPage` tests
- `mergeAcquiredImages` tests removed (no merging needed)
- Tool name references updated throughout
