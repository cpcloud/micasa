// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"testing"
)

// testSeed is the base seed for all faker instances in this package's tests.
// Set via MICASA_TEST_SEED env var, or generated randomly if unset.
var testSeed uint64

func TestMain(m *testing.M) {
	if s := os.Getenv("MICASA_TEST_SEED"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid MICASA_TEST_SEED=%q: %v\n", s, err)
			os.Exit(2)
		}
		testSeed = v
	} else {
		testSeed = rand.Uint64() //nolint:gosec // test seed, not crypto
	}
	fmt.Fprintf(os.Stderr, "MICASA_TEST_SEED=%d\n", testSeed)
	os.Exit(m.Run())
}
