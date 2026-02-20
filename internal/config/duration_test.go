// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDurationDaySuffix(t *testing.T) {
	d, err := ParseDuration("30d")
	require.NoError(t, err)
	assert.Equal(t, 30*24*time.Hour, d)
}

func TestParseDurationGoDuration(t *testing.T) {
	d, err := ParseDuration("720h")
	require.NoError(t, err)
	assert.Equal(t, 720*time.Hour, d)
}

func TestParseDurationBareInteger(t *testing.T) {
	d, err := ParseDuration("3600")
	require.NoError(t, err)
	assert.Equal(t, time.Hour, d)
}

func TestParseDurationZero(t *testing.T) {
	for _, input := range []string{"0", "0s", "0d"} {
		t.Run(input, func(t *testing.T) {
			d, err := ParseDuration(input)
			require.NoError(t, err)
			assert.Equal(t, time.Duration(0), d)
		})
	}
}

func TestParseDurationRejectsInvalid(t *testing.T) {
	for _, input := range []string{"", "abc", "30x"} {
		t.Run(input, func(t *testing.T) {
			_, err := ParseDuration(input)
			assert.Error(t, err)
		})
	}
}

func TestDurationUnmarshalTOMLInt(t *testing.T) {
	var d Duration
	require.NoError(t, d.UnmarshalTOML(int64(86400)))
	assert.Equal(t, 24*time.Hour, d.Duration)
}

func TestDurationUnmarshalTOMLString(t *testing.T) {
	var d Duration
	require.NoError(t, d.UnmarshalTOML("7d"))
	assert.Equal(t, 7*24*time.Hour, d.Duration)
}

func TestDurationUnmarshalTOMLRejectsOtherTypes(t *testing.T) {
	var d Duration
	assert.Error(t, d.UnmarshalTOML(3.14))
}
