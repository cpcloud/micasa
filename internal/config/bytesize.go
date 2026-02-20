// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package config

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// ByteSize represents a size in bytes, parseable from unitized strings
// like "50 MiB" or bare integers (interpreted as bytes).
type ByteSize int64

// Bytes returns the size in bytes.
func (b ByteSize) Bytes() int64 { return int64(b) }

// String returns a human-readable representation (e.g. "50 MiB").
func (b ByteSize) String() string {
	n := int64(b)
	switch {
	case n >= 1<<30 && n%(1<<30) == 0:
		return fmt.Sprintf("%d GiB", n/(1<<30))
	case n >= 1<<20 && n%(1<<20) == 0:
		return fmt.Sprintf("%d MiB", n/(1<<20))
	case n >= 1<<10 && n%(1<<10) == 0:
		return fmt.Sprintf("%d KiB", n/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// UnmarshalTOML implements toml.Unmarshaler for ByteSize,
// accepting both TOML integers (bytes) and strings ("50 MiB").
func (b *ByteSize) UnmarshalTOML(v any) error {
	switch val := v.(type) {
	case int64:
		*b = ByteSize(val)
		return nil
	case string:
		parsed, err := ParseByteSize(val)
		if err != nil {
			return err
		}
		*b = parsed
		return nil
	default:
		return fmt.Errorf("max_file_size: expected integer or string, got %T", v)
	}
}

var byteSizeRe = regexp.MustCompile(
	`(?i)^\s*([0-9]+(?:\.[0-9]+)?)\s*(b|kib|kb|mib|mb|gib|gb|tib|tb)?\s*$`,
)

// byteSizeMultipliers maps unit suffixes (lowercase) to their byte multiplier.
var byteSizeMultipliers = map[string]float64{
	"":    1,
	"b":   1,
	"kb":  1e3,
	"kib": 1 << 10,
	"mb":  1e6,
	"mib": 1 << 20,
	"gb":  1e9,
	"gib": 1 << 30,
	"tb":  1e12,
	"tib": 1 << 40,
}

// ParseByteSize parses a size string like "50 MiB", "1.5 GiB", or "1024".
// A bare integer is interpreted as bytes.
func ParseByteSize(s string) (ByteSize, error) {
	m := byteSizeRe.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf(
			"invalid byte size %q -- use a number with optional unit "+
				"(B, KiB, MiB, GiB, TiB, KB, MB, GB, TB), e.g. \"50 MiB\"",
			s,
		)
	}

	num, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid byte size number %q: %w", m[1], err)
	}

	unit := strings.ToLower(m[2])
	mult, ok := byteSizeMultipliers[unit]
	if !ok {
		return 0, fmt.Errorf("unknown byte size unit %q", m[2])
	}

	result := num * mult
	if result > math.MaxInt64 {
		return 0, fmt.Errorf("byte size %q overflows int64", s)
	}

	return ByteSize(int64(math.Round(result))), nil
}
