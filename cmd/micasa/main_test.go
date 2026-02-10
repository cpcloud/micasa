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

// Integration tests that invoke the built binary with --print-path.
// These exercise the full CLI parsing + resolveDBPath + print-and-exit path.

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

func TestPrintPath_Default(t *testing.T) {
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "--print-path")
	cmd.Env = append(os.Environ(), "MICASA_DB_PATH=")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--print-path failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	if got == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.HasSuffix(got, "micasa.db") {
		t.Errorf("expected path ending in micasa.db, got %q", got)
	}
}

func TestPrintPath_ExplicitPath(t *testing.T) {
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "--print-path", "/custom/path.db")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--print-path failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	if got != "/custom/path.db" {
		t.Errorf("got %q, want /custom/path.db", got)
	}
}

func TestPrintPath_EnvOverride(t *testing.T) {
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "--print-path")
	cmd.Env = append(os.Environ(), "MICASA_DB_PATH=/env/test.db")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--print-path failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	if got != "/env/test.db" {
		t.Errorf("got %q, want /env/test.db", got)
	}
}

func TestPrintPath_DemoNoPath(t *testing.T) {
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "--print-path", "--demo")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--print-path --demo failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	if got != ":memory:" {
		t.Errorf("got %q, want :memory:", got)
	}
}

func TestPrintPath_DemoWithPath(t *testing.T) {
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "--print-path", "--demo", "/tmp/demo.db")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--print-path --demo /path failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	if got != "/tmp/demo.db" {
		t.Errorf("got %q, want /tmp/demo.db", got)
	}
}

func TestPrintPath_ExitCodeZero(t *testing.T) {
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "--print-path")
	if err := cmd.Run(); err != nil {
		t.Errorf("expected exit 0, got %v", err)
	}
}

func TestVersion_DevShowsCommitHash(t *testing.T) {
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
