package dice

import (
	"testing"
)

func TestParse_Valid(t *testing.T) {
	cases := []struct {
		notation string
		want     Spec
	}{
		{"2D", Spec{Count: 2, Sides: 6, Modifier: 0}},
		{"1D", Spec{Count: 1, Sides: 6, Modifier: 0}},
		{"2D-7", Spec{Count: 2, Sides: 6, Modifier: -7}},
		{"1D+5", Spec{Count: 1, Sides: 6, Modifier: 5}},
		{"3D+2", Spec{Count: 3, Sides: 6, Modifier: 2}},
		{"D3", Spec{Count: 1, Sides: 3, Modifier: 0}},
		{"D3-1", Spec{Count: 1, Sides: 3, Modifier: -1}},
		{"d10", Spec{Count: 1, Sides: 10, Modifier: 0}},
		{"d100", Spec{Count: 1, Sides: 100, Modifier: 0}},
		{"2d10", Spec{Count: 2, Sides: 10, Modifier: 0}},
	}
	for _, tc := range cases {
		t.Run(tc.notation, func(t *testing.T) {
			got, err := Parse(tc.notation)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.notation, err)
			}
			if got != tc.want {
				t.Fatalf("Parse(%q) = %+v, want %+v", tc.notation, got, tc.want)
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	bad := []string{"", "2", "2X", "Dx", "2D-", "2D+", "D1", "0D"}
	for _, n := range bad {
		t.Run(n, func(t *testing.T) {
			if _, err := Parse(n); err == nil {
				t.Fatalf("Parse(%q) succeeded; want error", n)
			}
		})
	}
}
