// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	t.Run("tilde slash prefix", func(t *testing.T) {
		assert.Equal(t, filepath.Join(home, "foo.pdf"), expandHome("~/foo.pdf"))
	})
	t.Run("nested path", func(t *testing.T) {
		assert.Equal(
			t,
			filepath.Join(home, "docs", "invoice.pdf"),
			expandHome("~/docs/invoice.pdf"),
		)
	})
	t.Run("bare tilde", func(t *testing.T) {
		assert.Equal(t, home, expandHome("~"))
	})
	t.Run("absolute path unchanged", func(t *testing.T) {
		assert.Equal(t, "/tmp/foo.pdf", expandHome("/tmp/foo.pdf"))
	})
	t.Run("relative path unchanged", func(t *testing.T) {
		assert.Equal(t, "foo.pdf", expandHome("foo.pdf"))
	})
	t.Run("empty string unchanged", func(t *testing.T) {
		assert.Equal(t, "", expandHome(""))
	})
	t.Run("tilde other user unchanged", func(t *testing.T) {
		// ~otheruser/foo is a different expansion we don't support
		assert.Equal(t, "~otheruser/foo", expandHome("~otheruser/foo"))
	})
}

func TestOptionalFilePathExpandsTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	// Create a temp file inside home to test with a real tilde path.
	tmp := filepath.Join(home, ".micasa-test-file")
	require.NoError(t, os.WriteFile(tmp, []byte("test"), 0o600))
	t.Cleanup(func() { _ = os.Remove(tmp) })

	validate := optionalFilePath()
	assert.NoError(t, validate("~/.micasa-test-file"))
	assert.NoError(t, validate(tmp))
	assert.NoError(t, validate(""))
	assert.Error(t, validate("~/nonexistent-file-abc123"))
}
