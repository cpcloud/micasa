// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package locale

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveUSD(t *testing.T) {
	c, err := Resolve("USD")
	require.NoError(t, err)
	assert.Equal(t, "USD", c.Code())
	assert.Equal(t, "$", c.Symbol())
}

func TestResolveEUR(t *testing.T) {
	c, err := Resolve("EUR")
	require.NoError(t, err)
	assert.Equal(t, "EUR", c.Code())
	assert.Equal(t, "\u20ac", c.Symbol()) // euro sign
}

func TestResolveGBP(t *testing.T) {
	c, err := Resolve("GBP")
	require.NoError(t, err)
	assert.Equal(t, "GBP", c.Code())
	assert.Equal(t, "\u00a3", c.Symbol()) // pound sign
}

func TestResolveJPY(t *testing.T) {
	c, err := Resolve("JPY")
	require.NoError(t, err)
	assert.Equal(t, "JPY", c.Code())
	// CLDR narrow symbol for JPY in Japanese locale is fullwidth yen.
	assert.Equal(t, "\uffe5", c.Symbol()) // fullwidth yen sign
}

func TestResolveInvalid(t *testing.T) {
	_, err := Resolve("NOPE")
	assert.Error(t, err)
}

func TestResolveCaseInsensitive(t *testing.T) {
	c, err := Resolve("eur")
	require.NoError(t, err)
	assert.Equal(t, "EUR", c.Code())
}

func TestResolveEmpty(t *testing.T) {
	c, err := Resolve("")
	require.NoError(t, err)
	assert.Equal(t, "USD", c.Code())
}

func TestDefaultCurrency(t *testing.T) {
	c := DefaultCurrency()
	assert.Equal(t, "USD", c.Code())
	assert.Equal(t, "$", c.Symbol())
}

func TestFormatCentsUSD(t *testing.T) {
	c := MustResolve("USD")
	tests := []struct {
		name  string
		cents int64
		want  string
	}{
		{"zero", 0, "$0.00"},
		{"small", 99, "$0.99"},
		{"even dollars", 10000, "$100.00"},
		{"typical", 123456, "$1,234.56"},
		{"large", 100000000, "$1,000,000.00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, c.FormatCents(tt.cents))
		})
	}
}

func TestFormatCentsNegative(t *testing.T) {
	c := MustResolve("USD")
	assert.Equal(t, "-$5.00", c.FormatCents(-500))
}

func TestFormatCentsMinInt64(t *testing.T) {
	c := MustResolve("USD")
	formatted := c.FormatCents(math.MinInt64)
	assert.Contains(t, formatted, "-")
	assert.Contains(t, formatted, "$")
}

func TestFormatOptionalCentsNil(t *testing.T) {
	c := MustResolve("USD")
	assert.Empty(t, c.FormatOptionalCents(nil))
}

func TestFormatOptionalCentsNonNil(t *testing.T) {
	c := MustResolve("USD")
	cents := int64(123456)
	assert.Equal(t, "$1,234.56", c.FormatOptionalCents(&cents))
}

func TestFormatCentsEUR(t *testing.T) {
	c := MustResolve("EUR")
	formatted := c.FormatCents(123456)
	// EUR with German locale uses comma as decimal, period as grouping,
	// and symbol after the number.
	assert.Contains(t, formatted, "\u20ac")
	assert.Contains(t, formatted, "1.234,56")
}

func TestFormatCentsGBP(t *testing.T) {
	c := MustResolve("GBP")
	formatted := c.FormatCents(123456)
	assert.Contains(t, formatted, "\u00a3")
	assert.Contains(t, formatted, "1,234.56")
}

func TestFormatCompactCentsUSD(t *testing.T) {
	c := MustResolve("USD")
	tests := []struct {
		name  string
		cents int64
		want  string
	}{
		{"zero", 0, "$0.00"},
		{"small", 999, "$9.99"},
		{"just under 1k", 99999, "$999.99"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, c.FormatCompactCents(tt.cents))
		})
	}
}

func TestFormatCompactCentsLargeUSD(t *testing.T) {
	c := MustResolve("USD")
	assert.Equal(t, "$1k", c.FormatCompactCents(100000))
	assert.Equal(t, "$1.2k", c.FormatCompactCents(123456))
	assert.Equal(t, "$45k", c.FormatCompactCents(4500000))
	assert.Equal(t, "$1M", c.FormatCompactCents(100000000))
}

func TestFormatCompactCentsEUR(t *testing.T) {
	c := MustResolve("EUR")
	// EUR uses comma as decimal separator in compact notation too.
	assert.Contains(t, c.FormatCompactCents(123456), "1,2k")
	assert.Contains(t, c.FormatCompactCents(130000000), "1,3M")
	assert.Contains(t, c.FormatCompactCents(100000), "1k")
}

func TestFormatCompactOptionalCentsNil(t *testing.T) {
	c := MustResolve("USD")
	assert.Empty(t, c.FormatCompactOptionalCents(nil))
}

func TestParseRequiredCentsUSD(t *testing.T) {
	c := MustResolve("USD")
	tests := []struct {
		input string
		want  int64
	}{
		{"100", 10000},
		{"100.5", 10050},
		{"100.05", 10005},
		{"$1,234.56", 123456},
		{".75", 75},
		{"0.99", 99},
	}
	for _, tt := range tests {
		got, err := c.ParseRequiredCents(tt.input)
		require.NoError(t, err, "input=%q", tt.input)
		assert.Equal(t, tt.want, got, "input=%q", tt.input)
	}
}

func TestParseRequiredCentsInvalid(t *testing.T) {
	c := MustResolve("USD")
	for _, input := range []string{"", "12.345", "abc", "1.2.3"} {
		_, err := c.ParseRequiredCents(input)
		assert.Error(t, err, "input=%q", input)
	}
}

func TestParseCentsRejectsNegative(t *testing.T) {
	c := MustResolve("USD")
	for _, input := range []string{"-$5.00", "-5.00", "-$1,234.56"} {
		_, err := c.ParseRequiredCents(input)
		assert.ErrorIs(t, err, ErrNegativeMoney, "input=%q", input)
	}
}

func TestParseCentsRoundtripUSD(t *testing.T) {
	c := MustResolve("USD")
	values := []int64{0, 1, 99, 100, 123456}
	for _, cents := range values {
		formatted := c.FormatCents(cents)
		parsed, err := c.ParseRequiredCents(formatted)
		require.NoError(t, err, "roundtrip failed for %d (formatted=%q)", cents, formatted)
		assert.Equal(t, cents, parsed, "roundtrip mismatch for %d (formatted=%q)", cents, formatted)
	}
}

func TestParseCentsRoundtripEUR(t *testing.T) {
	c := MustResolve("EUR")
	values := []int64{0, 1, 99, 100, 123456}
	for _, cents := range values {
		formatted := c.FormatCents(cents)
		parsed, err := c.ParseRequiredCents(formatted)
		require.NoError(t, err, "roundtrip failed for %d (formatted=%q)", cents, formatted)
		assert.Equal(t, cents, parsed, "roundtrip mismatch for %d (formatted=%q)", cents, formatted)
	}
}

func TestParseOptionalCentsEmpty(t *testing.T) {
	c := MustResolve("USD")
	val, err := c.ParseOptionalCents("")
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestParseOptionalCentsValid(t *testing.T) {
	c := MustResolve("USD")
	val, err := c.ParseOptionalCents("5")
	require.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, int64(500), *val)
}

func TestParseCentsEURFormat(t *testing.T) {
	c := MustResolve("EUR")
	// EUR-formatted input with comma as decimal, period as grouping.
	cents, err := c.ParseRequiredCents("1.234,56")
	require.NoError(t, err)
	assert.Equal(t, int64(123456), cents)
}

func TestCurrencyFromLocaleString(t *testing.T) {
	tests := []struct {
		locale string
		want   string
	}{
		{"en_US.UTF-8", "USD"},
		{"de_DE.UTF-8", "EUR"},
		{"ja_JP.UTF-8", "JPY"},
		{"en_GB.UTF-8", "GBP"},
		{"", ""},
		{"C", ""},
		{"POSIX", ""},
	}
	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			assert.Equal(t, tt.want, currencyFromLocaleString(tt.locale))
		})
	}
}

func TestParseCentsOverflow(t *testing.T) {
	c := MustResolve("USD")
	tests := []struct {
		name  string
		input string
	}{
		{"one dollar over", "$92233720368547759.00"},
		{"way over", "$999999999999999999999.99"},
		{"frac overflow at boundary", "$92233720368547758.08"},
		{"frac overflow .99 at boundary", "$92233720368547758.99"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.ParseRequiredCents(tt.input)
			assert.Error(t, err, "should reject overflow: %s", tt.input)
		})
	}
}

func TestParseCentsAtMaxSafeValue(t *testing.T) {
	c := MustResolve("USD")
	cents, err := c.ParseRequiredCents("$92233720368547758.00")
	require.NoError(t, err)
	assert.Equal(t, int64(9223372036854775800), cents)

	cents, err = c.ParseRequiredCents("$92233720368547758.07")
	require.NoError(t, err)
	assert.Equal(t, int64(9223372036854775807), cents)
}
