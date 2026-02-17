// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cpcloud/micasa/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestDefaultsApplied(t *testing.T) {
	// Point to a nonexistent file so no config is loaded.
	cfg, err := LoadFromPath(filepath.Join(t.TempDir(), "nope.toml"))
	require.NoError(t, err)
	assert.Equal(t, DefaultBaseURL, cfg.LLM.BaseURL)
	assert.Equal(t, DefaultModel, cfg.LLM.Model)
}

func TestLoadFromFile(t *testing.T) {
	path := writeConfig(t, `[llm]
base_url = "http://myhost:8080/v1"
model = "llama3"
extra_context = "My house is old."
`)
	cfg, err := LoadFromPath(path)
	require.NoError(t, err)
	assert.Equal(t, "http://myhost:8080/v1", cfg.LLM.BaseURL)
	assert.Equal(t, "llama3", cfg.LLM.Model)
	assert.Equal(t, "My house is old.", cfg.LLM.ExtraContext)
}

func TestPartialConfigUsesDefaults(t *testing.T) {
	path := writeConfig(t, `[llm]
model = "phi3"
`)
	cfg, err := LoadFromPath(path)
	require.NoError(t, err)
	assert.Equal(t, DefaultBaseURL, cfg.LLM.BaseURL)
	assert.Equal(t, "phi3", cfg.LLM.Model)
}

func TestEnvOverridesConfig(t *testing.T) {
	path := writeConfig(t, `[llm]
base_url = "http://file-host:1234/v1"
model = "from-file"
`)
	t.Setenv("OLLAMA_HOST", "http://env-host:5678")
	t.Setenv("MICASA_LLM_MODEL", "from-env")

	cfg, err := LoadFromPath(path)
	require.NoError(t, err)
	assert.Equal(t, "http://env-host:5678/v1", cfg.LLM.BaseURL)
	assert.Equal(t, "from-env", cfg.LLM.Model)
}

func TestOllamaHostAppendsV1(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "http://myhost:11434")

	cfg, err := LoadFromPath(filepath.Join(t.TempDir(), "nope.toml"))
	require.NoError(t, err)
	assert.Equal(t, "http://myhost:11434/v1", cfg.LLM.BaseURL)
}

func TestOllamaHostAlreadyHasV1(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "http://myhost:11434/v1")

	cfg, err := LoadFromPath(filepath.Join(t.TempDir(), "nope.toml"))
	require.NoError(t, err)
	assert.Equal(t, "http://myhost:11434/v1", cfg.LLM.BaseURL)
}

func TestTrailingSlashStripped(t *testing.T) {
	path := writeConfig(t, `[llm]
base_url = "http://localhost:11434/v1/"
`)
	cfg, err := LoadFromPath(path)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:11434/v1", cfg.LLM.BaseURL)
}

func TestExampleTOML(t *testing.T) {
	example := ExampleTOML()
	assert.Contains(t, example, "[llm]")
	assert.Contains(t, example, "base_url")
	assert.Contains(t, example, "model")
	assert.Contains(t, example, "timeout")
}

func TestMalformedConfigReturnsError(t *testing.T) {
	path := writeConfig(t, "{{not toml")

	_, err := LoadFromPath(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestDefaultMaxDocumentSize(t *testing.T) {
	cfg, err := LoadFromPath(filepath.Join(t.TempDir(), "nope.toml"))
	require.NoError(t, err)
	assert.Equal(t, data.MaxDocumentSize, cfg.Documents.MaxFileSize)
}

func TestMaxDocumentSizeFromFile(t *testing.T) {
	path := writeConfig(t, `[documents]
max_file_size = 1048576
`)
	cfg, err := LoadFromPath(path)
	require.NoError(t, err)
	assert.Equal(t, int64(1048576), cfg.Documents.MaxFileSize)
}

func TestMaxDocumentSizeEnvOverride(t *testing.T) {
	t.Setenv("MICASA_MAX_DOCUMENT_SIZE", "2097152")
	cfg, err := LoadFromPath(filepath.Join(t.TempDir(), "nope.toml"))
	require.NoError(t, err)
	assert.Equal(t, int64(2097152), cfg.Documents.MaxFileSize)
}

func TestMaxDocumentSizeRejectsInvalid(t *testing.T) {
	for _, val := range []string{"-1", "0"} {
		t.Run(val, func(t *testing.T) {
			path := writeConfig(t, "[documents]\nmax_file_size = "+val+"\n")
			_, err := LoadFromPath(path)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "must be positive")
		})
	}
}

func TestDefaultCacheTTLDays(t *testing.T) {
	cfg, err := LoadFromPath(filepath.Join(t.TempDir(), "nope.toml"))
	require.NoError(t, err)
	assert.Equal(t, DefaultCacheTTLDays, cfg.Documents.CacheTTLDays)
}

func TestCacheTTLDaysFromFile(t *testing.T) {
	path := writeConfig(t, "[documents]\ncache_ttl_days = 7\n")
	cfg, err := LoadFromPath(path)
	require.NoError(t, err)
	assert.Equal(t, 7, cfg.Documents.CacheTTLDays)
}

func TestCacheTTLDaysZeroDisables(t *testing.T) {
	path := writeConfig(t, "[documents]\ncache_ttl_days = 0\n")
	cfg, err := LoadFromPath(path)
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.Documents.CacheTTLDays)
}

func TestCacheTTLDaysEnvOverride(t *testing.T) {
	t.Setenv("MICASA_CACHE_TTL_DAYS", "14")
	cfg, err := LoadFromPath(filepath.Join(t.TempDir(), "nope.toml"))
	require.NoError(t, err)
	assert.Equal(t, 14, cfg.Documents.CacheTTLDays)
}

func TestCacheTTLDaysRejectsNegative(t *testing.T) {
	path := writeConfig(t, "[documents]\ncache_ttl_days = -1\n")
	_, err := LoadFromPath(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be non-negative")
}

func TestLLMTimeout(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		cfg, err := LoadFromPath(filepath.Join(t.TempDir(), "nope.toml"))
		require.NoError(t, err)
		assert.Equal(t, DefaultLLMTimeout, cfg.LLM.TimeoutDuration())
	})

	t.Run("from file", func(t *testing.T) {
		path := writeConfig(t, "[llm]\ntimeout = \"10s\"\n")
		cfg, err := LoadFromPath(path)
		require.NoError(t, err)
		assert.Equal(t, 10*time.Second, cfg.LLM.TimeoutDuration())
	})

	t.Run("sub-second", func(t *testing.T) {
		path := writeConfig(t, "[llm]\ntimeout = \"500ms\"\n")
		cfg, err := LoadFromPath(path)
		require.NoError(t, err)
		assert.Equal(t, 500*time.Millisecond, cfg.LLM.TimeoutDuration())
	})

	t.Run("env override", func(t *testing.T) {
		t.Setenv("MICASA_LLM_TIMEOUT", "15s")
		cfg, err := LoadFromPath(filepath.Join(t.TempDir(), "nope.toml"))
		require.NoError(t, err)
		assert.Equal(t, 15*time.Second, cfg.LLM.TimeoutDuration())
	})

	t.Run("rejects invalid", func(t *testing.T) {
		path := writeConfig(t, "[llm]\ntimeout = \"not-a-duration\"\n")
		_, err := LoadFromPath(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid duration")
	})

	t.Run("rejects negative", func(t *testing.T) {
		path := writeConfig(t, "[llm]\ntimeout = \"-1s\"\n")
		_, err := LoadFromPath(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})
}
