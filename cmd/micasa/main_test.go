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
)

func TestResolveDBPath_ExplicitPath(t *testing.T) {
	c := cli{DBPath: "/custom/path.db"}
	got, err := resolveDBPath(c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/custom/path.db" {
		t.Errorf("got %q, want /custom/path.db", got)
	}
}

func TestResolveDBPath_ExplicitPathWithDemo(t *testing.T) {
	// Explicit path takes precedence even when --demo is set.
	c := cli{DBPath: "/tmp/demo.db", Demo: true}
	got, err := resolveDBPath(c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/demo.db" {
		t.Errorf("got %q, want /tmp/demo.db", got)
	}
}

func TestResolveDBPath_DemoNoPath(t *testing.T) {
	c := cli{Demo: true}
	got, err := resolveDBPath(c)
	if err != nil {
		t.Fatal(err)
	}
	if got != ":memory:" {
		t.Errorf("got %q, want :memory:", got)
	}
}

func TestResolveDBPath_Default(t *testing.T) {
	// With no flags, resolveDBPath falls through to DefaultDBPath.
	// Clear the env override so the platform default is used.
	t.Setenv("MICASA_DB_PATH", "")
	c := cli{}
	got, err := resolveDBPath(c)
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Fatal("expected non-empty default path")
	}
	if !strings.HasSuffix(got, "micasa.db") {
		t.Errorf("expected path ending in micasa.db, got %q", got)
	}
}

func TestResolveDBPath_EnvOverride(t *testing.T) {
	// MICASA_DB_PATH env var is honored when no positional arg is given.
	t.Setenv("MICASA_DB_PATH", "/env/override.db")
	c := cli{}
	got, err := resolveDBPath(c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/env/override.db" {
		t.Errorf("got %q, want /env/override.db", got)
	}
}

func TestResolveDBPath_ExplicitPathBeatsEnv(t *testing.T) {
	// Positional arg takes precedence over env var.
	t.Setenv("MICASA_DB_PATH", "/env/override.db")
	c := cli{DBPath: "/explicit/wins.db"}
	got, err := resolveDBPath(c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/explicit/wins.db" {
		t.Errorf("got %q, want /explicit/wins.db", got)
	}
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
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func TestVersion_DevShowsCommitHash(t *testing.T) {
	// Skip when there is no .git directory (e.g. Nix sandbox builds from a
	// source tarball), since Go won't embed VCS info without one.
	if _, err := os.Stat(".git"); err != nil {
		t.Skip("no .git directory; VCS info unavailable (e.g. Nix sandbox)")
	}
	bin := buildTestBinary(t)
	out, err := exec.Command(bin, "--version").Output()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	// Built inside a git repo: expect a hex hash, possibly with -dirty.
	if got == "dev" {
		t.Error("expected commit hash, got bare dev")
	}
	hash := strings.TrimSuffix(got, "-dirty")
	for _, c := range hash {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("expected hex hash, got %q", got)
			break
		}
	}
}

func TestVersion_Injected(t *testing.T) {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	bin := filepath.Join(t.TempDir(), "micasa"+ext)
	cmd := exec.Command("go", "build",
		"-ldflags", "-X main.version=1.2.3",
		"-o", bin, ".")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	out, err := exec.Command(bin, "--version").Output()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	if got != "1.2.3" {
		t.Errorf("got %q, want 1.2.3", got)
	}
}
