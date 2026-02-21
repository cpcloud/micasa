// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapOpenerError_NotFound(t *testing.T) {
	tests := []struct {
		name     string
		opener   string
		wantSub  string // substring that must appear
		wantHint string // actionable hint substring
	}{
		{
			name:     "xdg-open",
			opener:   "xdg-open",
			wantSub:  "xdg-open not found",
			wantHint: "xdg-utils",
		},
		{
			name:     "open (macOS)",
			opener:   "open",
			wantSub:  "open not found",
			wantHint: "headless",
		},
		{
			name:     "unknown opener",
			opener:   "something-else",
			wantSub:  "something-else not found",
			wantHint: "no file opener available",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapOpenerError(exec.ErrNotFound, tt.opener)
			assert.ErrorContains(t, got, tt.wantSub)
			assert.ErrorContains(t, got, tt.wantHint)
		})
	}
}

func TestWrapOpenerError_OtherError(t *testing.T) {
	other := errors.New("exit status 1")
	got := wrapOpenerError(other, "xdg-open")
	assert.Equal(t, other, got, "non-ErrNotFound errors should pass through unchanged")
}
