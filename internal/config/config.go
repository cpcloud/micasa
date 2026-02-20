// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

	// Timeout is the maximum time to wait for quick LLM server operations
	// (ping, model listing, auto-detect). Go duration string, e.g. "5s",
	// "10s", "500ms". Default: "5s".
	Timeout string `toml:"timeout"`
}

// TimeoutDuration returns the parsed LLM timeout, falling back to
// DefaultLLMTimeout if the value is empty or unparseable.
func (l LLM) TimeoutDuration() time.Duration {
	if l.Timeout == "" {
		return DefaultLLMTimeout
	}
	d, err := time.ParseDuration(l.Timeout)
	if err != nil {
		return DefaultLLMTimeout
	}
	return d
}

// Documents holds settings for document attachments.
type Documents struct {
	// MaxFileSize is the largest file that can be imported as a document
	// attachment. Accepts unitized strings ("50 MiB") or bare integers
	// (bytes). Default: 50 MiB.
	MaxFileSize ByteSize `toml:"max_file_size"`

	// CacheTTL is the preferred cache lifetime setting. Accepts unitized
	// strings ("30d", "720h") or bare integers (seconds). Default: 30d.
	CacheTTL *Duration `toml:"cache_ttl,omitempty"`

	// CacheTTLDays is deprecated; use CacheTTL instead. Kept for backward
	// compatibility. Bare integer interpreted as days.
	CacheTTLDays *int `toml:"cache_ttl_days,omitempty"`
}

// CacheTTLDuration returns the resolved cache TTL as a time.Duration.
// CacheTTL takes precedence over CacheTTLDays. Returns 0 to disable.
func (d Documents) CacheTTLDuration() time.Duration {
	if d.CacheTTL != nil {
		return d.CacheTTL.Duration
	}
	if d.CacheTTLDays != nil {
		return time.Duration(*d.CacheTTLDays) * 24 * time.Hour
	}
	return DefaultCacheTTL
}

const (
	DefaultBaseURL    = "http://localhost:11434/v1"
	DefaultModel      = "qwen3"
	DefaultLLMTimeout = 5 * time.Second
	DefaultCacheTTL   = 30 * 24 * time.Hour // 30 days
	configRelPath     = "micasa/config.toml"
)

// defaults returns a Config with all default values populated.
func defaults() Config {
	return Config{
		LLM: LLM{
			BaseURL: DefaultBaseURL,
			Model:   DefaultModel,
			Timeout: DefaultLLMTimeout.String(),
		},
		Documents: Documents{
			MaxFileSize: ByteSize(data.MaxDocumentSize),
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

	if cfg.LLM.Timeout != "" {
		d, err := time.ParseDuration(cfg.LLM.Timeout)
		if err != nil {
			return cfg, fmt.Errorf(
				"llm.timeout: invalid duration %q -- use Go syntax like \"5s\" or \"10s\"",
				cfg.LLM.Timeout,
			)
		}
		if d <= 0 {
			return cfg, fmt.Errorf("llm.timeout must be positive, got %s", cfg.LLM.Timeout)
		}
	}

	if cfg.Documents.MaxFileSize == 0 {
		return cfg, fmt.Errorf(
			"documents.max_file_size must be positive, got 0",
		)
	}

	if cfg.Documents.CacheTTL != nil && cfg.Documents.CacheTTLDays != nil {
		return cfg, fmt.Errorf(
			"documents.cache_ttl and documents.cache_ttl_days cannot both be set -- " +
				"remove cache_ttl_days (deprecated) and use cache_ttl instead",
		)
	}

	if cfg.Documents.CacheTTLDays != nil {
		log.Printf(
			"warning: documents.cache_ttl_days is deprecated -- " +
				"use documents.cache_ttl (e.g. \"30d\") instead",
		)
		if *cfg.Documents.CacheTTLDays < 0 {
			return cfg, fmt.Errorf(
				"documents.cache_ttl_days must be non-negative, got %d",
				*cfg.Documents.CacheTTLDays,
			)
		}
	}

	if cfg.Documents.CacheTTL != nil && cfg.Documents.CacheTTL.Duration < 0 {
		return cfg, fmt.Errorf(
			"documents.cache_ttl must be non-negative, got %s",
			cfg.Documents.CacheTTL.Duration,
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
	if timeout := os.Getenv("MICASA_LLM_TIMEOUT"); timeout != "" {
		cfg.LLM.Timeout = timeout
	}
	if maxSize := os.Getenv("MICASA_MAX_DOCUMENT_SIZE"); maxSize != "" {
		if parsed, err := ParseByteSize(maxSize); err == nil {
			cfg.Documents.MaxFileSize = parsed
		}
	}
	if ttl := os.Getenv("MICASA_CACHE_TTL"); ttl != "" {
		if parsed, err := ParseDuration(ttl); err == nil {
			d := Duration{parsed}
			cfg.Documents.CacheTTL = &d
		}
	}
	if ttl := os.Getenv("MICASA_CACHE_TTL_DAYS"); ttl != "" {
		if n, err := strconv.Atoi(ttl); err == nil {
			cfg.Documents.CacheTTLDays = &n
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

# Timeout for quick LLM server operations (ping, model listing).
# Go duration syntax: "5s", "10s", "500ms", etc. Default: "5s".
# Increase if your LLM server is slow to respond.
# timeout = "5s"

[documents]
# Maximum file size for document imports. Accepts unitized strings or bare
# integers (bytes). Default: 50 MiB.
# max_file_size = "50 MiB"

# How long to keep extracted document cache entries before evicting on startup.
# Accepts "30d", "720h", or bare integers (seconds). Set to "0s" to disable.
# Default: 30d.
# cache_ttl = "30d"
`
}
