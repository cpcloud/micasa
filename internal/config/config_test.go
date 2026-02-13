// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
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
}

func TestMalformedConfigReturnsError(t *testing.T) {
	path := writeConfig(t, "{{not toml")

	_, err := LoadFromPath(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}
