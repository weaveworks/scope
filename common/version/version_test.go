package version

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	for input, expected := range map[string]*Version{
		"0.0.0":     {0, 0, 0},
		"0.0.1":     {0, 0, 1},
		"0.10.0":    {0, 10, 0},
		"0.9.0":     {0, 9, 0},
		"0.10.1":    {0, 10, 1},
		"1.10.1":    {1, 10, 1},
		"1.1100.1":  {1, 1100, 1},
		"23.11.657": {23, 11, 657},
		"v0.0.0":    {0, 0, 0},
		"v0.0.1":    {0, 0, 1},
		"v0.10.0":   {0, 10, 0},
		"v0.9.0":    {0, 9, 0},
		"v0.10.1":   {0, 10, 1},
		"v1.10.1":   {1, 10, 1},
		"v1.1100.1": {1, 1100, 1},
		"0.0":       nil,
		"0.9":       nil,
		"0.10":      nil,
		"v0.0":      nil,
		"v0.10":     nil,
		"v1.10":     nil,
		"1.10.1-b":  nil,
		"v1.10.1-b": nil,
		"1.10.1-r":  nil,
		"v1.10.1-r": nil,
		"":          nil,
		"a":         nil,
		"abc":       nil,
		"1":         nil,
		"12":        nil,
		"v1":        nil,
		"v12":       nil,
		"1669ff8e":  nil,
		"1669ff8e285641526c3e56e942e9f50cbe6b03b6": nil,
	} {
		v, err := ParseVersion(input)
		if err == nil && expected == nil {
			t.Errorf("%v parsed to %v, but did not expect to parse", input, v)
		} else if err == nil && *expected != *v {
			t.Errorf("%v parsed to %v, but expected %v", input, v, expected)
		} else if err != nil && expected != nil {
			t.Errorf("%v did not parse, but expected to parse to %v", input, expected)
		}
	}
}

func parse(t *testing.T, version string) Version {
	v, err := ParseVersion(version)
	if err != nil {
		t.Errorf("%v did not parse", version)
	}
	return *v
}

func TestVerionLower(t *testing.T) {
	successes := map[string]string{
		"1.30.0": "2.0.0",
		"1.31.1": "1.32.1",
	}
	failures := map[string]string{
		"2.65.130":    "2.65.130",
		"2.123.15565": "0.23123.2111",
	}

	for lhs, rhs := range successes {
		vLHS := parse(t, lhs)
		vRHS := parse(t, rhs)
		if !vLHS.Lower(vRHS) {
			t.Errorf("expected %v < %v", vLHS, vRHS)
		}
	}
	for lhs, rhs := range failures {
		vLHS := parse(t, lhs)
		vRHS := parse(t, rhs)
		if vLHS.Lower(vRHS) {
			t.Errorf("expected %v >= %v", vLHS, vRHS)
		}
	}
}
