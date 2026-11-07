package main

import "testing"

func TestNormalizeSalaryRange(t *testing.T) {
	tests := []struct {
		name             string
		minimum, maximum int
		wantMin, wantMax int
	}{
		{name: "both zero", minimum: 0, maximum: 0, wantMin: 5000, wantMax: 10000},
		{name: "zero minimum", minimum: 0, maximum: 8000, wantMin: 5000, wantMax: 10000},
		{name: "zero maximum", minimum: 5000, maximum: 0, wantMin: 5000, wantMax: 10000},
		{name: "valid range", minimum: 6000, maximum: 9000, wantMin: 6000, wantMax: 9000},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			minimum, maximum := normalizeSalaryRange(test.minimum, test.maximum)
			if minimum != test.wantMin || maximum != test.wantMax {
				t.Fatalf("normalizeSalaryRange(%d, %d) = (%d, %d), want (%d, %d)", test.minimum, test.maximum, minimum, maximum, test.wantMin, test.wantMax)
			}
		})
	}
}
