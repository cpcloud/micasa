// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveDBPath_ExplicitPath(t *testing.T) {
	c := cli{DBPath: "/custom/path.db"}
	got, err := resolveDBPath(c)
	require.NoError(t, err)
	assert.Equal(t, "/custom/path.db", got)
}

func TestResolveDBPath_ExplicitPathWithDemo(t *testing.T) {
	// Explicit path takes precedence even when --demo is set.
	c := cli{DBPath: "/tmp/demo.db", Demo: true}
	got, err := resolveDBPath(c)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/demo.db", got)
}

func TestResolveDBPath_DemoNoPath(t *testing.T) {
	c := cli{Demo: true}
	got, err := resolveDBPath(c)
	require.NoError(t, err)
	assert.Equal(t, ":memory:", got)
}

func TestResolveDBPath_Default(t *testing.T) {
	// With no flags, resolveDBPath falls through to DefaultDBPath.
	// Clear the env override so the platform default is used.
	t.Setenv("MICASA_DB_PATH", "")
	c := cli{}
	got, err := resolveDBPath(c)
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.True(
		t,
		strings.HasSuffix(got, "micasa.db"),
		"expected path ending in micasa.db, got %q",
		got,
	)
}

func TestResolveDBPath_EnvOverride(t *testing.T) {
	// MICASA_DB_PATH env var is honored when no positional arg is given.
	t.Setenv("MICASA_DB_PATH", "/env/override.db")
	c := cli{}
	got, err := resolveDBPath(c)
	require.NoError(t, err)
	assert.Equal(t, "/env/override.db", got)
}

func TestResolveDBPath_ExplicitPathBeatsEnv(t *testing.T) {
	// Positional arg takes precedence over env var.
	t.Setenv("MICASA_DB_PATH", "/env/override.db")
	c := cli{DBPath: "/explicit/wins.db"}
	got, err := resolveDBPath(c)
	require.NoError(t, err)
	assert.Equal(t, "/explicit/wins.db", got)
}

// Version tests use exec.Command("go", "build") because debug.ReadBuildInfo()
// only embeds VCS revision info in binaries built with go build, not go test,
// and -ldflags -X injection likewise requires a real build step.

func buildTestBinary(t *testing.T) string {
	t.Helper()
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	bin := filepath.Join(t.TempDir(), "micasa"+ext)
	cmd := exec.Command( //nolint:gosec // test helper with constant args
		"go",
		"build",
		"-o",
		bin,
		".",
	)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "build failed:\n%s", out)
	return bin
}

func TestVersion_DevShowsCommitHash(t *testing.T) {
	// Skip when there is no .git directory (e.g. Nix sandbox builds from a
	// source tarball), since Go won't embed VCS info without one.
	if _, err := os.Stat(".git"); err != nil {
		t.Skip("no .git directory; VCS info unavailable (e.g. Nix sandbox)")
	}
	bin := buildTestBinary(t)
	verCmd := exec.Command(bin, "--version") //nolint:gosec // test binary path from buildTestBinary
	out, err := verCmd.Output()
	require.NoError(t, err, "--version failed")
	got := strings.TrimSpace(string(out))
	// Built inside a git repo: expect a hex hash, possibly with -dirty.
	assert.NotEqual(t, "dev", got, "expected commit hash, got bare dev")
	assert.Regexp(t, `^[0-9a-f]+(-dirty)?$`, got, "expected hex hash, got %q", got)
}

func TestVersion_Injected(t *testing.T) {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	bin := filepath.Join(t.TempDir(), "micasa"+ext)
	cmd := exec.Command("go", "build", //nolint:gosec // test with constant args
		"-ldflags", "-X main.version=1.2.3",
		"-o", bin, ".")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "build failed:\n%s", out)
	verCmd := exec.Command(bin, "--version") //nolint:gosec // test binary path from above
	verOut, err := verCmd.Output()
	require.NoError(t, err, "--version failed")
	assert.Equal(t, "1.2.3", strings.TrimSpace(string(verOut)))
}
