package worlds

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
)

// TestRollGasGiantSize_Classes covers the three category branches per
// WBH p.55 Gas Giant Sizing table (no DMs).
func TestRollGasGiantSize_Classes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		dice          []int
		dms           int
		wantClass     GasGiantClass
		wantDiamCode  string
		wantDiamEarth float64
		wantMassEarth float64
	}{
		{
			// selector 1D=2 → Small; D3=1, D3=2 → diam=3; 1D=4 → mass=5×(4+1)=25
			name: "Small GS smallest", dice: []int{2, 1, 2, 4}, dms: 0,
			wantClass: GasGiantSmall, wantDiamCode: "3", wantDiamEarth: 3, wantMassEarth: 25,
		},
		{
			// selector 1D=3 → Medium; 1D+6=11 → diam=11; 3D=12 → mass=20×(12-1)=220
			name: "Medium GM mid", dice: []int{3, 11, 12}, dms: 0,
			wantClass: GasGiantMedium, wantDiamCode: "B", wantDiamEarth: 11, wantMassEarth: 220,
		},
		{
			// selector 1D=5 → Large; 2D+6=14 → diam=14; D3=2; 3D=12 → mass=2×50×(12+4)=1600
			name: "Large GL big", dice: []int{5, 14, 2, 12}, dms: 0,
			wantClass: GasGiantLarge, wantDiamCode: "E", wantDiamEarth: 14, wantMassEarth: 1600,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := roller.NewScripted(tc.dice...)

			got, err := RollGasGiantSize(r, tc.dms)
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			if got.Class != tc.wantClass {
				t.Errorf("Class = %v, want %v", got.Class, tc.wantClass)
			}

			if got.DiameterCode != tc.wantDiamCode {
				t.Errorf("DiameterCode = %q, want %q", got.DiameterCode, tc.wantDiamCode)
			}

			if math.Abs(got.DiameterEarth-tc.wantDiamEarth) > 1e-9 {
				t.Errorf("DiameterEarth = %v, want %v", got.DiameterEarth, tc.wantDiamEarth)
			}

			if math.Abs(got.MassEarth-tc.wantMassEarth) > 1e-9 {
				t.Errorf("MassEarth = %v, want %v", got.MassEarth, tc.wantMassEarth)
			}
		})
	}
}

// TestRollGasGiantSize_DMs verifies that dms shifts the selector roll.
func TestRollGasGiantSize_DMs(t *testing.T) {
	t.Parallel()

	// selector 1D=3, dms=-1 → 2 → Small; D3=1, D3=2 → diam=3; 1D=4 → mass=25
	r := roller.NewScripted(3, 1, 2, 4)

	got, err := RollGasGiantSize(r, -1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if got.Class != GasGiantSmall {
		t.Errorf("Class with dms=-1 = %v, want GasGiantSmall (selector 3-1=2 → Small)", got.Class)
	}

	// selector 1D=4, dms=-2 → 2 → Small; D3=1, D3=2 → diam=3; 1D=4 → mass=25
	r = roller.NewScripted(4, 1, 2, 4)

	got, err = RollGasGiantSize(r, -2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if got.Class != GasGiantSmall {
		t.Errorf("Class with dms=-2 = %v, want GasGiantSmall (selector 4-2=2 → Small)", got.Class)
	}
}

// TestRollGasGiantSize_LargeMassClamp covers the WBH p.55 footnote:
// if initial Large GG mass ≥3,000⊕ (3D ≥15), substitute mass = 4000 - 200×(2D-2).
//
// selector 1D=5 → Large; 2D+6=18 → diam=18 (eHex "J"); D3=3; 3D=18 → mass=3×50×22=3300 ≥3000;
// clamp: 2D=7 → mass=4000-200×(7-2)=3000.
func TestRollGasGiantSize_LargeMassClamp(t *testing.T) {
	t.Parallel()

	r := roller.NewScripted(5, 18, 3, 18, 7)

	got, err := RollGasGiantSize(r, 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if got.Class != GasGiantLarge {
		t.Errorf("Class = %v, want GasGiantLarge", got.Class)
	}

	if got.DiameterCode != "J" {
		t.Errorf("DiameterCode = %q, want \"J\" (eHex 18)", got.DiameterCode)
	}

	if math.Abs(got.MassEarth-3000) > 1e-9 {
		t.Errorf("MassEarth = %v, want 3000 (clamped from 3300 via 4000-200×5)", got.MassEarth)
	}
}

// TestRollGasGiantSize_ZedExamples reproduces the four Zed gas giants:
//
//	Aab IV  GLE  diameter 14⊕, mass 1,200⊕
//	Aab V   GLC  diameter 12⊕, mass 800⊕
//	AB III  GMB  diameter 11⊕, mass 180⊕
//	Cab I   GS4  diameter 4⊕,  mass 10⊕
//
// Zed dice come from p.56 Size Rolls column.
func TestRollGasGiantSize_ZedExamples(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		dice          []int
		dms           int
		wantClass     GasGiantClass
		wantDiamCode  string
		wantDiamEarth float64
		wantMassEarth float64
	}{
		{
			// selector 1D=5 → Large; 2D+6=14 (4+4+6) → diam=14; D3=2; 3D=8 (3+3+2) → mass=2×50×12=1200
			"Aab IV GLE",
			[]int{5, 14, 2, 8},
			0, GasGiantLarge, "E", 14, 1200,
		},
		{
			// selector 1D=6 → Large; 2D+6=12 (2+4+6) → diam=12; D3=1; 3D=12 (4+4+4) → mass=1×50×16=800
			"Aab V GLC",
			[]int{6, 12, 1, 12},
			0, GasGiantLarge, "C", 12, 800,
		},
		{
			// selector 1D=3 → Medium; 1D+6=11 (5+6) → diam=11; 3D=10 (4+3+3) → mass=20×9=180
			"AB III GMB",
			[]int{3, 11, 10},
			0, GasGiantMedium, "B", 11, 180,
		},
		{
			// selector 1D=1 → Small; D3=1, D3=3 → diam=4; 1D=1 → mass=5×(1+1)=10
			"Cab I GS4",
			[]int{1, 1, 3, 1},
			0, GasGiantSmall, "4", 4, 10,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := roller.NewScripted(tc.dice...)

			got, err := RollGasGiantSize(r, tc.dms)
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			if got.Class != tc.wantClass {
				t.Errorf("Class = %v, want %v", got.Class, tc.wantClass)
			}

			if got.DiameterCode != tc.wantDiamCode {
				t.Errorf("DiameterCode = %q, want %q", got.DiameterCode, tc.wantDiamCode)
			}

			if math.Abs(got.DiameterEarth-tc.wantDiamEarth) > 1e-9 {
				t.Errorf("DiameterEarth = %v, want %v", got.DiameterEarth, tc.wantDiamEarth)
			}

			if math.Abs(got.MassEarth-tc.wantMassEarth) > 1e-9 {
				t.Errorf("MassEarth = %v, want %v", got.MassEarth, tc.wantMassEarth)
			}
		})
	}
}
