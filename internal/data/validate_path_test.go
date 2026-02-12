// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDBPath(t *testing.T) {
	tests := []struct {
		path    string
		wantErr string // substring of error, "" means no error
	}{
		// Valid paths.
		{path: ":memory:"},
		{path: "/home/user/micasa.db"},
		{path: "relative/path.db"},
		{path: "./local.db"},
		{path: "../parent/db.sqlite"},
		{path: "/tmp/micasa test.db"},
		{path: "C:\\Users\\me\\micasa.db"},

		// URI schemes -- must be rejected.
		{path: "https://evil.com/db", wantErr: "looks like a URI"},
		{path: "http://localhost/db", wantErr: "looks like a URI"},
		{path: "ftp://files.example.com/data.db", wantErr: "looks like a URI"},
		{path: "file://localhost/tmp/test.db", wantErr: "looks like a URI"},

		// file: without // -- SQLite still interprets this as URI.
		{path: "file:/tmp/test.db", wantErr: "file: scheme"},
		{path: "file:test.db", wantErr: "file: scheme"},
		{path: "file:test.db?mode=ro", wantErr: "file: scheme"},

		// Query parameters -- trigger url.ParseQuery in driver.
		{path: "/tmp/test.db?_pragma=journal_mode(wal)", wantErr: "contains '?'"},
		{path: "test.db?cache=shared", wantErr: "contains '?'"},

		// Empty path.
		{path: "", wantErr: "must not be empty"},

		// Not a scheme: no letters before "://".
		{path: "/path/with://in/middle"},

		// Numeric prefix before :// is not a scheme.
		{path: "123://not-a-scheme"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			err := ValidateDBPath(tt.path)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr,
				"ValidateDBPath(%q) = %v, want error containing %q", tt.path, err, tt.wantErr)
		})
	}
}

func TestValidateDBPathRejectsRandomURLs(t *testing.T) {
	f := gofakeit.New(testSeed)
	for i := 0; i < 100; i++ {
		u := f.URL()
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			assert.Error(t, ValidateDBPath(u), "ValidateDBPath(%q) should reject", u)
		})
	}
}

func TestValidateDBPathRejectsRandomURLsWithQueryParams(t *testing.T) {
	f := gofakeit.New(testSeed)
	for i := 0; i < 50; i++ {
		u := fmt.Sprintf("%s?%s=%s", f.URL(), f.Word(), f.Word())
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			assert.Error(t, ValidateDBPath(u), "ValidateDBPath(%q) should reject", u)
		})
	}
}

func TestOpenRejectsURIs(t *testing.T) {
	f := gofakeit.New(testSeed)
	for i := 0; i < 10; i++ {
		u := f.URL()
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := Open(u)
			require.Error(t, err, "Open(%q) should reject URI paths", u)
		})
	}
}
