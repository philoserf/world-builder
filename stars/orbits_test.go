package stars

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
)

// ----- P2-4: RollStellarOrbit -----

func TestRollStellarOrbit_Close_NaturalOne(t *testing.T) {
	// 1D=1 -> 1D-1 = 0 -> mapped to Orbit# 0.5.
	r := roller.NewScripted(0)
	got, err := RollStellarOrbit(r, OrbitClose, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 0.5 {
		t.Fatalf("got %v want 0.5", got)
	}
}

func TestRollStellarOrbit_Close_OtherRolls(t *testing.T) {
	// 1D=4 -> 1D-1 = 3.
	r := roller.NewScripted(3)
	got, err := RollStellarOrbit(r, OrbitClose, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 3.0 {
		t.Fatalf("got %v want 3", got)
	}
}

func TestRollStellarOrbit_Near(t *testing.T) {
	// 1D=4 -> 1D+5 = 9.
	r := roller.NewScripted(9)
	got, err := RollStellarOrbit(r, OrbitNear, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 9.0 {
		t.Fatalf("got %v want 9", got)
	}
}

func TestRollStellarOrbit_Far(t *testing.T) {
	// 1D=1 -> 1D+11 = 12.
	r := roller.NewScripted(12)
	got, err := RollStellarOrbit(r, OrbitFar, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 12.0 {
		t.Fatalf("got %v want 12", got)
	}
}

func TestRollStellarOrbit_Companion(t *testing.T) {
	// 1D=1, then 2D-7=-5 (e.g. 2D=2): 1/10 + (-5)/100 = 0.1 - 0.05 = 0.05.
	r := roller.NewScripted(1, -5)
	got, err := RollStellarOrbit(r, OrbitCompanion, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-0.05) > 1e-9 {
		t.Fatalf("got %v want 0.05", got)
	}
}

func TestRollStellarOrbit_CompanionGiantErrors(t *testing.T) {
	r := roller.NewScripted(1, 0)
	for _, lc := range []LuminosityClass{Ia, Ib, II, III} {
		t.Run(string(lc), func(t *testing.T) {
			_, err := RollStellarOrbit(r, OrbitCompanion, lc)
			if err == nil {
				t.Fatal("expected error for giant primary")
			}
		})
	}
}

// ----- P2-5: RollEccentricity -----

func TestRollEccentricity_Star_BaseDM(t *testing.T) {
	// IsStar=true -> DM+2. 2D=6 -> row 8 -> Base 0.03 + 1D/100.
	// Then 1D=3 -> 0.03 + 0.03 = 0.06.
	r := roller.NewScripted(6, 3)
	got, err := RollEccentricity(r, EccentricityOpts{IsStar: true})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 0.06
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestRollEccentricity_LowClamp(t *testing.T) {
	// 2D=2 with no DMs -> row 5 (Base -0.001 + 1D/1000).
	// 1D=1 -> -0.001 + 0.001 = 0.0. Clamped to [0, 0.999].
	r := roller.NewScripted(2, 1)
	got, err := RollEccentricity(r, EccentricityOpts{})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got < 0 || got > 0.005 {
		t.Fatalf("got %v not in [0, 0.005]", got)
	}
}

func TestRollEccentricity_HighClamp(t *testing.T) {
	// 2D=12 -> row 12 -> Base 0.30 + 2D/20.
	// 2D=12 -> 0.30 + 0.6 = 0.9.
	r := roller.NewScripted(12, 12)
	got, err := RollEccentricity(r, EccentricityOpts{})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-0.9) > 1e-9 {
		t.Fatalf("got %v want 0.9", got)
	}
}

func TestRollEccentricity_NestingDM(t *testing.T) {
	// NestingDepth=2 -> DM+2. 2D=6 with DM+2 -> row 8 -> 0.03 + 1D/100.
	// 1D=2 -> 0.05.
	r := roller.NewScripted(6, 2)
	got, err := RollEccentricity(r, EccentricityOpts{NestingDepth: 2})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 0.05
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestRollEccentricity_ExtraDM(t *testing.T) {
	t.Parallel()
	// 2D=7, 1D=1 with no DM -> row 7 -> Base 0.00 + 1/200 = 0.005.
	// 2D=7, 1D=1 with ExtraDM=5 -> row 12 -> Base 0.30 + 1D(=1)/20 = 0.35.
	// (Row 12 SecondRoll is "2D" in notation but the scripted roller yields the
	// next scripted value regardless of the dice expression.)
	base, err := RollEccentricity(roller.NewScripted(7, 1), EccentricityOpts{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if want := 0.005; math.Abs(base-want) > 1e-9 {
		t.Errorf("base = %v, want %v (row 7: 0.00 + 1/200)", base, want)
	}
	bumped, err := RollEccentricity(roller.NewScripted(7, 1), EccentricityOpts{ExtraDM: 5})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if want := 0.35; math.Abs(bumped-want) > 1e-9 {
		t.Errorf("bumped = %v, want %v (row 12: 0.30 + 1/20)", bumped, want)
	}
}

// ----- P2-6: RollInclination -----

func TestRollInclination_VeryLow(t *testing.T) {
	// 2D=4 (in 2-6 range) -> Very Low: 1D/2. 1D=4 -> 2.0.
	r := roller.NewScripted(4, 4)
	deg, sev, err := RollInclination(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(deg-2.0) > 1e-9 {
		t.Fatalf("deg = %v want 2.0", deg)
	}
	if sev != "VeryLow" {
		t.Fatalf("sev = %q want VeryLow", sev)
	}
}

func TestRollInclination_Moderate(t *testing.T) {
	// 2D=8 -> Moderate: 2D. 2D=7 -> 7.
	r := roller.NewScripted(8, 7)
	deg, sev, err := RollInclination(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if deg != 7.0 {
		t.Fatalf("deg = %v want 7", deg)
	}
	if sev != "Moderate" {
		t.Fatalf("sev = %q want Moderate", sev)
	}
}

func TestRollInclination_High(t *testing.T) {
	// 2D=9 -> High: (2D × 3) + 1D. 2D=4, 1D=2 -> 12 + 2 = 14.
	r := roller.NewScripted(9, 4, 2)
	deg, sev, err := RollInclination(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if deg != 14.0 {
		t.Fatalf("deg = %v want 14", deg)
	}
	if sev != "High" {
		t.Fatalf("sev = %q want High", sev)
	}
}

func TestRollInclination_Retrograde(t *testing.T) {
	// 2D=12 -> Retrograde: 180 - recursive. Recursive 2D=8 (Moderate, 2D=7 -> 7).
	// Final: 180 - 7 = 173.
	r := roller.NewScripted(12, 8, 7)
	deg, sev, err := RollInclination(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if deg != 173.0 {
		t.Fatalf("deg = %v want 173", deg)
	}
	if sev != "Retrograde" {
		t.Fatalf("sev = %q want Retrograde", sev)
	}
}

// ----- P2-7: OrbitPeriodYears -----

func TestOrbitPeriodYears_Earth(t *testing.T) {
	// Earth: 1 AU, 1 + ~0 solar masses -> 1 year.
	got := OrbitPeriodYears(1.0, 1.0, 0.0)
	if math.Abs(got-1.0) > 1e-9 {
		t.Fatalf("got %v want 1.0", got)
	}
}

func TestOrbitPeriodYears_ZedAB(t *testing.T) {
	// Zed AB (book p.34): AU=5.68, M+m=2.462 -> sqrt(5.68^3 / 2.462) ≈ 8.627y.
	got := OrbitPeriodYears(5.68, 2.462, 0.0)
	want := 8.627
	if math.Abs(got-want) > 1e-2 {
		t.Fatalf("got %v want %v (within 1e-2)", got, want)
	}
}

func TestOrbitPeriodYears_ZedAabSubsystem(t *testing.T) {
	// Zed Aab inner pair (book p.34): AU=0.036, M+m=1.836 -> very short period.
	got := OrbitPeriodYears(0.036, 1.836, 0.0)
	want := 0.005
	if math.Abs(got-want) > 1e-3 {
		t.Fatalf("got %v want ~%v", got, want)
	}
}

// ----- P2-8: OrbitToAU and AUToOrbit -----

func TestOrbitToAU_Whole(t *testing.T) {
	cases := []struct {
		orbit, want float64
	}{
		{0, 0},
		{1, 0.4},
		{3, 1.0},
		{5, 2.8},
		{7, 10},
	}
	for _, tc := range cases {
		got := OrbitToAU(tc.orbit)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("OrbitToAU(%v) = %v, want %v", tc.orbit, got, tc.want)
		}
	}
}

func TestOrbitToAU_Fractional(t *testing.T) {
	// WBH p.26 example: Orbit# 4.3 = 1.6 + 1.2*0.3 = 1.96 AU.
	got := OrbitToAU(4.3)
	want := 1.96
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestOrbitToAU_HalfOrbit(t *testing.T) {
	// Orbit# 0.5 = 0 + 0.4*0.5 = 0.2 AU (Close-orbit special).
	got := OrbitToAU(0.5)
	if math.Abs(got-0.2) > 1e-9 {
		t.Fatalf("got %v want 0.2", got)
	}
}

func TestAUToOrbit_Whole(t *testing.T) {
	cases := []struct {
		au, want float64
	}{
		{0.4, 1},
		{1.0, 3},
		{2.8, 5},
		{5.2, 6},
	}
	for _, tc := range cases {
		got := AUToOrbit(tc.au)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("AUToOrbit(%v) = %v, want %v", tc.au, got, tc.want)
		}
	}
}

func TestAUToOrbit_BookExample(t *testing.T) {
	// WBH p.26 example: 3.4 AU -> 5.25 Orbit#.
	// 3.4 - 2.8 = 0.6; 0.6 / 2.4 = 0.25; full=5; result 5.25.
	got := AUToOrbit(3.4)
	want := 5.25
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestAUToOrbit_RoundTrip(t *testing.T) {
	values := []float64{0.5, 1.5, 2.5, 5.0, 7.5, 12.3, 18.0}
	for _, x := range values {
		au := OrbitToAU(x)
		back := AUToOrbit(au)
		if math.Abs(x-back) > 1e-9 {
			t.Errorf("round-trip %v -> %v -> %v", x, au, back)
		}
	}
}
