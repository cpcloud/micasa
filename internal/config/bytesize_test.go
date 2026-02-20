// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseByteSizeBareInteger(t *testing.T) {
	b, err := ParseByteSize("1024")
	require.NoError(t, err)
	assert.Equal(t, int64(1024), b.Bytes())
}

func TestParseByteSizeIECUnits(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"100 B", 100},
		{"1 KiB", 1 << 10},
		{"50 MiB", 50 << 20},
		{"2 GiB", 2 << 30},
		{"1 TiB", 1 << 40},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			b, err := ParseByteSize(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, b.Bytes())
		})
	}
}

func TestParseByteSizeSIUnits(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"1 KB", 1000},
		{"50 MB", 50_000_000},
		{"1 GB", 1_000_000_000},
		{"1 TB", 1_000_000_000_000},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			b, err := ParseByteSize(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, b.Bytes())
		})
	}
}

func TestParseByteSizeFractional(t *testing.T) {
	b, err := ParseByteSize("1.5 GiB")
	require.NoError(t, err)
	assert.Equal(t, int64(1.5*(1<<30)), b.Bytes())
}

func TestParseByteSizeCaseInsensitive(t *testing.T) {
	b, err := ParseByteSize("50 mib")
	require.NoError(t, err)
	assert.Equal(t, int64(50<<20), b.Bytes())
}

func TestParseByteSizeNoSpace(t *testing.T) {
	b, err := ParseByteSize("50MiB")
	require.NoError(t, err)
	assert.Equal(t, int64(50<<20), b.Bytes())
}

func TestParseByteSizeRejectsInvalid(t *testing.T) {
	for _, input := range []string{"", "abc", "50 XiB", "MiB"} {
		t.Run(input, func(t *testing.T) {
			_, err := ParseByteSize(input)
			assert.Error(t, err)
		})
	}
}

func TestByteSizeString(t *testing.T) {
	tests := []struct {
		size ByteSize
		want string
	}{
		{ByteSize(1 << 30), "1 GiB"},
		{ByteSize(50 << 20), "50 MiB"},
		{ByteSize(1 << 10), "1 KiB"},
		{ByteSize(500), "500 B"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.size.String())
		})
	}
}

func TestByteSizeUnmarshalTOMLInt(t *testing.T) {
	var b ByteSize
	require.NoError(t, b.UnmarshalTOML(int64(1024)))
	assert.Equal(t, int64(1024), b.Bytes())
}

func TestByteSizeUnmarshalTOMLString(t *testing.T) {
	var b ByteSize
	require.NoError(t, b.UnmarshalTOML("50 MiB"))
	assert.Equal(t, int64(50<<20), b.Bytes())
}

func TestByteSizeUnmarshalTOMLRejectsOtherTypes(t *testing.T) {
	var b ByteSize
	assert.Error(t, b.UnmarshalTOML(3.14))
}
