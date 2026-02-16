// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"

	"github.com/cpcloud/micasa/internal/data"
)

// Config is the top-level application configuration, loaded from a TOML file.
type Config struct {
	LLM       LLM       `toml:"llm"`
	Documents Documents `toml:"documents"`
}

// LLM holds settings for the local LLM inference backend.
type LLM struct {
	// BaseURL is the root of an OpenAI-compatible API.
	// The client appends /chat/completions, /models, etc.
	// Default: http://localhost:11434/v1 (Ollama)
	BaseURL string `toml:"base_url"`

	// Model is the model identifier passed in chat requests.
	// Default: qwen3
	Model string `toml:"model"`

	// ExtraContext is custom text appended to all system prompts.
	// Useful for domain-specific details: house style, currency, location, etc.
	// Optional; defaults to empty.
	ExtraContext string `toml:"extra_context"`
}

// Documents holds settings for document attachments.
type Documents struct {
	// MaxFileSize is the largest file (in bytes) that can be imported as a
	// document attachment. Default: 50 MiB.
	MaxFileSize int64 `toml:"max_file_size"`

	// CacheTTLDays is the number of days an extracted document cache entry
	// is kept before being evicted on the next startup. Set to 0 to disable
	// eviction. Default: 30.
	CacheTTLDays int `toml:"cache_ttl_days"`
}

const (
	DefaultBaseURL      = "http://localhost:11434/v1"
	DefaultModel        = "qwen3"
	DefaultCacheTTLDays = 30
	configRelPath       = "micasa/config.toml"
)

// defaults returns a Config with all default values populated.
func defaults() Config {
	return Config{
		LLM: LLM{
			BaseURL: DefaultBaseURL,
			Model:   DefaultModel,
		},
		Documents: Documents{
			MaxFileSize:  data.MaxDocumentSize,
			CacheTTLDays: DefaultCacheTTLDays,
		},
	}
}

// Path returns the expected config file path (XDG_CONFIG_HOME/micasa/config.toml).
func Path() string {
	return filepath.Join(xdg.ConfigHome, configRelPath)
}

// Load reads the TOML config file from the default path if it exists, falls
// back to defaults for any unset fields, and applies environment variable
// overrides last.
func Load() (Config, error) {
	return LoadFromPath(Path())
}

// LoadFromPath reads the TOML config file at the given path if it exists,
// falls back to defaults for any unset fields, and applies environment
// variable overrides last.
func LoadFromPath(path string) (Config, error) {
	cfg := defaults()

	if _, err := os.Stat(path); err == nil {
		if _, err := toml.DecodeFile(path, &cfg); err != nil {
			return cfg, fmt.Errorf("parse %s: %w", path, err)
		}
	}

	applyEnvOverrides(&cfg)

	// Normalize: strip trailing slash from base URL.
	cfg.LLM.BaseURL = strings.TrimRight(cfg.LLM.BaseURL, "/")

	if cfg.Documents.MaxFileSize <= 0 {
		return cfg, fmt.Errorf(
			"documents.max_file_size must be positive, got %d",
			cfg.Documents.MaxFileSize,
		)
	}

	if cfg.Documents.CacheTTLDays < 0 {
		return cfg, fmt.Errorf(
			"documents.cache_ttl_days must be non-negative, got %d",
			cfg.Documents.CacheTTLDays,
		)
	}

	return cfg, nil
}

// applyEnvOverrides lets environment variables override config-file values.
// OLLAMA_HOST sets the base URL (with /v1 appended if missing).
// MICASA_LLM_MODEL sets the model.
func applyEnvOverrides(cfg *Config) {
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		host = strings.TrimRight(host, "/")
		if !strings.HasSuffix(host, "/v1") {
			host += "/v1"
		}
		cfg.LLM.BaseURL = host
	}
	if model := os.Getenv("MICASA_LLM_MODEL"); model != "" {
		cfg.LLM.Model = model
	}
	if maxSize := os.Getenv("MICASA_MAX_DOCUMENT_SIZE"); maxSize != "" {
		if n, err := strconv.ParseInt(maxSize, 10, 64); err == nil {
			cfg.Documents.MaxFileSize = n
		}
	}
	if ttl := os.Getenv("MICASA_CACHE_TTL_DAYS"); ttl != "" {
		if n, err := strconv.Atoi(ttl); err == nil {
			cfg.Documents.CacheTTLDays = n
		}
	}
}

// ExampleTOML returns a commented config file suitable for writing as a
// starter config. Not written automatically -- offered to the user on demand.
func ExampleTOML() string {
	return `# micasa configuration
# Place this file at: ` + Path() + `

[llm]
# Base URL for an OpenAI-compatible API endpoint.
# Ollama (default): http://localhost:11434/v1
# llama.cpp:        http://localhost:8080/v1
# LM Studio:        http://localhost:1234/v1
base_url = "` + DefaultBaseURL + `"

# Model name passed in chat requests.
model = "` + DefaultModel + `"

# Optional: custom context appended to all system prompts.
# Use this to inject domain-specific details about your house, currency, etc.
# extra_context = "My house is a 1920s craftsman in Portland, OR. All budgets are in CAD."

[documents]
# Maximum file size (in bytes) for document imports. Default: 50 MiB.
# max_file_size = 52428800

# Days to keep extracted document cache entries before evicting on startup.
# Set to 0 to disable eviction. Default: 30.
# cache_ttl_days = 30
`
}
