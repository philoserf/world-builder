package worlds

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
)

// TestBasicTerrestrialDiameter verifies the WBH p.54 Basic Terrestrial
// World Size table: every code maps to its book diameter in km.
func TestBasicTerrestrialDiameter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		code SizeCode
		km   float64
	}{
		{"0", 0},
		{"R", 0},
		{"S", 600},
		{"1", 1600},
		{"2", 3200},
		{"3", 4800},
		{"4", 6400},
		{"5", 8000},
		{"6", 9600},
		{"7", 11200},
		{"8", 12800},
		{"9", 14400},
		{"A", 16000},
		{"B", 17600},
		{"C", 19200},
		{"D", 20800},
		{"E", 22400},
		{"F", 24000},
	}
	for _, tc := range cases {
		t.Run(string(tc.code), func(t *testing.T) {
			t.Parallel()
			got := BasicTerrestrialDiameter(tc.code)
			if math.Abs(got-tc.km) > 1e-9 {
				t.Errorf("BasicTerrestrialDiameter(%q) = %v, want %v", tc.code, got, tc.km)
			}
		})
	}
}

// TestRollTerrestrialSize_Branches covers each of the three 1D
// selector branches per WBH p.54 Terrestrial World Sizing table:
//
//	1-2 → second roll 1D    range 1-6
//	3-4 → second roll 2D    range 2-C(12)
//	5-6 → second roll 2D+3  range 5-F(15)
func TestRollTerrestrialSize_Branches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		dice     []int
		wantCode SizeCode
	}{
		{"selector1 second1D=4", []int{1, 4}, "4"},
		{"selector3 second2D=7", []int{3, 7}, "7"},
		{"selector4 second2D=12", []int{4, 12}, "C"},
		{"selector5 second2D+3=11", []int{5, 8}, "B"},
		{"selector6 second2D+3=15", []int{6, 12}, "F"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roller.NewScripted(tc.dice...)
			got, err := RollTerrestrialSize(r)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got.SizeCode != tc.wantCode {
				t.Errorf("SizeCode = %q, want %q", got.SizeCode, tc.wantCode)
			}
			if got.DiameterKm != BasicTerrestrialDiameter(tc.wantCode) {
				t.Errorf("DiameterKm = %v, want %v", got.DiameterKm, BasicTerrestrialDiameter(tc.wantCode))
			}
		})
	}
}

// TestRollTerrestrialSize_ZedAabII reproduces a Zed terrestrial size from p.56:
//
//	Aab II  Terrestrial  Size Rolls "1: 6 = 6"  Code 6
//
// Selector 1D=1 → branch 1-2 → second 1D=6 → Size 6.
func TestRollTerrestrialSize_ZedAabII(t *testing.T) {
	t.Parallel()
	r := roller.NewScripted(1, 6)
	got, err := RollTerrestrialSize(r)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "6" {
		t.Errorf("SizeCode = %q, want \"6\"", got.SizeCode)
	}
	if got.DiameterKm != 9600 {
		t.Errorf("DiameterKm = %v, want 9600", got.DiameterKm)
	}
}
