// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package extract

import (
	"os"
	"runtime"
	"testing"
)

// ciStrictOS returns true when running in CI on an OS where all
// extraction tools (tesseract, pdftoppm, pdftotext, magick) are
// expected to be installed. Currently that's Linux and macOS.
func ciStrictOS() bool {
	return os.Getenv("CI") != "" && runtime.GOOS != "windows"
}

// skipOrFatalCI skips the test when tools/fixtures are missing
// locally, but fails hard in CI on Linux/macOS where everything
// should be available.
func skipOrFatalCI(t *testing.T, msg string) {
	t.Helper()
	if ciStrictOS() {
		t.Fatalf("CI: %s", msg)
	}
	t.Skip(msg)
}
