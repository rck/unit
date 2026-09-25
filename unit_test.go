// Copyright 2017 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unit

import "testing"

var m = map[string]int64{
	"B":   1,
	"KiB": 1024,
	"MiB": 1024 * 1024,
}

var mapTests = []struct {
	description string
	mapping     map[string]int64
	expectFail  bool
}{
	{
		description: "Valid map 1",
		mapping: map[string]int64{
			"B": 1,
		},
		expectFail: false,
	}, {
		description: "Valid map 2",
		mapping: map[string]int64{
			"B":   1,
			"KiB": 1024,
			"MiB": 1024 * 1024,
			"GiB": 1024 * 1024 * 1024,
			"TiB": 1024 * 1024 * 1024 * 1024,
		},
		expectFail: false,
	}, {
		description: "Valid map 3",
		mapping: map[string]int64{
			"B":   1,
			"KiB": 1024,
			"MiB": 1024 * 1024,
			"":    1024 * 1024,
			"GiB": 1024 * 1024 * 1024,
			"TiB": 1024 * 1024 * 1024 * 1024,
		},
		expectFail: false,
	}, {
		description: "Invalid map 1 (empty map)",
		mapping:     map[string]int64{},
		expectFail:  true,
	}, {
		description: "Invalid map 2 (no mapping to multiplier 1)",
		mapping: map[string]int64{
			"KB": 1000,
			"MB": 1000 * 1000,
			"GB": 1000 * 1000 * 1000,
		},
		expectFail: true,
	}, {
		description: "Invalid map 3 (contains mapping to multiplier 0)",
		mapping: map[string]int64{
			"Bad": 0,
			"B":   1,
			"KiB": 1024,
			"MiB": 1024 * 1024,
			"GiB": 1024 * 1024 * 1024,
			"TiB": 1024 * 1024 * 1024 * 1024,
		},
		expectFail: true,
	}, {
		description: "Invalid map 4 (default unit does not match a named unit)",
		mapping: map[string]int64{
			"":  1024 * 1024,
			"B": 1,
		},
		expectFail: true,
	}, {
		description: "Invalid map 5 (default unit is the only mapping to multiplier 1)",
		mapping: map[string]int64{
			"":  1,
			"K": 1024,
		},
		expectFail: true,
	}, {
		description: "Invalid map 6 (default unit maps to multiplier 0)",
		mapping: map[string]int64{
			"":  0,
			"B": 1,
		},
		expectFail: true,
	},
}

var unitTests = []struct {
	in, out      string
	value        int64
	explicitSign Sign
	err          bool
}{
	{
		// Invalid
		in:  "",
		err: true,
	}, {
		// With invalid unit
		in:  "23K",
		err: true,
	}, {
		// Out of range value * unit multiplication
		in:  "18014398509482034KiB",
		err: true,
	}, {
		// With valid unit
		in:    "23KiB",
		out:   "23KiB",
		value: 23 * 1024,
	}, {
		// With valid without unit
		in:    "23",
		out:   "23B",
		value: 23,
	}, {
		// With valid unit, multiple of another unit
		in:    "1024KiB",
		out:   "1MiB",
		value: 1024 * 1024,
	}, {
		// With valid unit, multiple of another unit, explicit +
		in:           "+1024KiB",
		out:          "+1MiB",
		value:        1024 * 1024,
		explicitSign: Positive,
	}, {
		// With valid unit, multiple of another unit, explicit -
		in:           "-1024KiB",
		out:          "-1MiB",
		value:        -1024 * 1024,
		explicitSign: Negative,
	},
}

// TestNewUnit tests the NewUnit function
func TestNewUnit(t *testing.T) {
	for _, test := range mapTests {
		_, err := NewUnit(test.mapping)
		if test.expectFail && err == nil {
			t.Fatalf("Test case \"%s\" did not return an error as expected\n", test.description)
		} else if !test.expectFail && err != nil {
			t.Fatalf("Test case \"%s\" return an unexpected error: %v\n", test.description, err)
		}
	}
}

// TestUnit implements a table-driven test.
func TestUnit(t *testing.T) {
	u, err := NewUnit(m)
	if err != nil {
		t.Fatalf("Expected that a valid unit map does not return an error: %v\n", err)
	}

	for _, test := range unitTests {
		v, err := u.ValueFromString(test.in)
		if test.err && err == nil {
			t.Fatalf("Expected that test '%s' returns an error\n", test.in)
		} else if !test.err && err != nil {
			t.Fatalf("Expected that test '%s' does not return an error\n", test.in)
		} else if test.err && err != nil {
			continue
		}

		if v.Value != test.value {
			t.Fatalf("Expected that test '%s' returns %d as bytes, but got %d\n", test.in, test.value, v.Value)
		}

		if v.String() != test.out {
			t.Fatalf("Expected that test '%s' returns %s as string, but got %s\n", test.in, test.out, v)
		}

		if v.ExplicitSign != test.explicitSign {
			t.Fatalf("Expected that test '%s' has explicit Sign %d, but got %d\n", test.in, test.explicitSign, v.ExplicitSign)
		}
	}
}

// TestStringDeterministic checks that String() picks the same unit
// regardless of map iteration order and that the result parses back.
func TestStringDeterministic(t *testing.T) {
	tests := []struct {
		description string
		mapping     map[string]int64
		in, out     string
	}{
		{
			description: "named unit wins the tie against the default unit",
			mapping:     map[string]int64{"": G, "B": 1, "GiB": G},
			in:          "10",
			out:         "10GiB",
		}, {
			description: "default unit is never printed",
			mapping:     map[string]int64{"": M, "B": 1, "M": M},
			in:          "3",
			out:         "3M",
		}, {
			description: "default unit with multiplier 1",
			mapping:     map[string]int64{"": 1, "B": 1, "K": K},
			in:          "1500",
			out:         "1500B",
		}, {
			description: "equal multipliers with equal length: lexicographically smaller name",
			mapping:     DefaultUnits,
			in:          "2000",
			out:         "2KB",
		}, {
			description: "equal multipliers with different length: shorter name",
			mapping:     map[string]int64{"B": 1, "K": K, "KiB": K},
			in:          "2K",
			out:         "2K",
		}, {
			description: "explicit sign is kept",
			mapping:     map[string]int64{"": M, "B": 1, "M": M},
			in:          "-3",
			out:         "-3M",
		},
	}

	for _, test := range tests {
		u := MustNewUnit(test.mapping)
		v, err := u.ValueFromString(test.in)
		if err != nil {
			t.Fatalf("Test case \"%s\": '%s' returned an unexpected error: %v\n", test.description, test.in, err)
		}
		// map iteration order is random, so repeat to catch order dependence
		for i := 0; i < 100; i++ {
			if s := v.String(); s != test.out {
				t.Fatalf("Test case \"%s\": expected %s, but got %s\n", test.description, test.out, s)
			}
		}
		back, err := u.ValueFromString(v.String())
		if err != nil {
			t.Fatalf("Test case \"%s\": String() result %q does not parse back: %v\n", test.description, v.String(), err)
		}
		if back.Value != v.Value {
			t.Fatalf("Test case \"%s\": round trip changed value from %d to %d\n", test.description, v.Value, back.Value)
		}
	}
}

// TestNewUnitCopiesMapping checks that later changes to the input map do
// not affect the Unit.
func TestNewUnitCopiesMapping(t *testing.T) {
	m := map[string]int64{"B": 1, "K": K}
	u := MustNewUnit(m)
	m["M"] = M
	if _, err := u.ValueFromString("1M"); err == nil {
		t.Fatalf("Expected that unit 'M' added after NewUnit is not valid\n")
	}
}
