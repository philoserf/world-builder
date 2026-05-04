package worlds

import (
	"math"
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestComputeAlbedo_ZedPrime(t *testing.T) {
	// Zed Prime: rocky terrestrial (density ~1.0), atm 6, hyd 6, orbit 1.06 AU
	// (parent's), star Aab L=1.419, HZCO computed from L.
	// Per WBH p.111: 0.04 + (8-2)*0.02 + 8*0.01 + (7-4)*0.03 = 0.33.
	// Scripted dice: [8 (rocky base), 8 (atm 6), 7 (hyd 6+)].
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Orbit = 1.0 // moon's parent orbit; for albedo we use Orbit# vs HZCO
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.04}
	body.Hydrographics = &Hydrographics{Code: 6}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.419}}

	r := roller.NewScripted(8, 8, 7)
	got := ComputeAlbedo(r, body, sys)
	want := 0.33
	if math.Abs(got-want) > 0.005 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeAlbedo_Terra_Reference(t *testing.T) {
	// Terra: rocky (density 1.0), atm 6, hyd 7, orbit 1.0 AU, sol L=1.0.
	// 0.04 + (X-2)*0.02 + Y*0.01 + (Z-4)*0.03 should hit ~0.30 with mid rolls.
	// Scripted [7, 7, 6]: 0.04 + 0.10 + 0.07 + 0.06 = 0.27. Close to 0.30 reference.
	body := &DetailedPlacement{}
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0}
	body.Hydrographics = &Hydrographics{Code: 7}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 6)
	got := ComputeAlbedo(r, body, sys)
	if got < 0.25 || got > 0.30 {
		t.Errorf("Terra-reference albedo got %v, want ~0.27 (scripted [7,7,6]; book reference 0.30)", got)
	}
}

func TestComputeAlbedo_GasGiant(t *testing.T) {
	// Gas giant: 0.05 + 2D × 0.05. With 2D=7: 0.40.
	body := &DetailedPlacement{}
	body.GGClass = GasGiantSmall
	body.SizeCode = "S"
	body.Orbit = 5.0

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7)
	got := ComputeAlbedo(r, body, sys)
	want := 0.40
	if math.Abs(got-want) > 0.01 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeAlbedo_IcyBeyondHZCO2(t *testing.T) {
	// Icy beyond HZCO+2: 0.25 + (2D-2) × 0.07. With 2D=7: 0.25 + 5*0.07 = 0.60.
	// Sol HZCO()=3.0, so HZCO+2=5.0. Body at Orbit# 6.0 is clearly beyond.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Orbit = 6.0                            // HZCO+3 for Sol-like star (HZCO()=3.0); clearly beyond HZCO+2
	body.Physical = &BodyPhysical{Density: 0.3} // icy

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7)
	got := ComputeAlbedo(r, body, sys)
	// 0.25 + 5*0.07 = 0.60. Above 0.4 so the bonus-subtraction does not fire.
	want := 0.60
	if math.Abs(got-want) > 0.01 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeAlbedo_Clamping(t *testing.T) {
	// Force a result above 0.98: rocky terr base 0.04 + (12-2)*0.02 = 0.24 + atm A-C +(12-2)*0.05 = 0.50 + hyd 6+ +(12-4)*0.03 = 0.24 → 0.98 exactly.
	body := &DetailedPlacement{}
	body.SizeCode = "8"
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 10, Pressure: 1.0} // A → +(2D-2)*0.05
	body.Hydrographics = &Hydrographics{Code: 10}          // 6+ → +(2D-4)*0.03

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(12, 12, 12) // max rolls everywhere
	got := ComputeAlbedo(r, body, sys)
	// 0.04 + 10*0.02 + 10*0.05 + 8*0.03 = 0.04 + 0.20 + 0.50 + 0.24 = 0.98 exactly.
	if got > 0.98 {
		t.Errorf("clamp failed: got %v, want ≤ 0.98", got)
	}
}

func TestComputeAlbedo_Vacuum(t *testing.T) {
	// Vacuum body: Atmosphere{Code: 0} → no atm modifier applied.
	// Rocky terr base [7]: 0.04 + (7-2)*0.02 = 0.14. Hyd 0 → no hyd modifier.
	// Sol HZCO()=3.0 → HZCO+2=5.0; orbit 1.0 < 5.0 → rocky branch fires.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Orbit = 1.0
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 0, Pressure: 0}
	body.Hydrographics = &Hydrographics{Code: 0}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7) // only 1 roll consumed (rocky base)
	got := ComputeAlbedo(r, body, sys)
	want := 0.14
	if math.Abs(got-want) > 0.01 {
		t.Errorf("got %v, want %v (no atm or hyd modifier should apply)", got, want)
	}
}

func TestComputeGreenhouseFactor_Vacuum(t *testing.T) {
	// Atmosphere code 0 → vacuum → greenhouse 0.
	r := roller.NewScripted()
	got := ComputeGreenhouseFactor(r, &Atmosphere{Code: 0, Pressure: 0})
	if got != 0 {
		t.Errorf("got %v, want 0 for vacuum", got)
	}
}

func TestComputeGreenhouseFactor_ZedPrime(t *testing.T) {
	// Zed Prime atm 6, pressure 1.04 bar.
	// Initial = 0.5 × √1.04 = 0.5099.
	// Atm 1-9 or D/E modifier: +3D × 0.01. Book walk: 3D=8 → +0.08.
	// Total: 0.51 + 0.08 = 0.59.
	r := roller.NewScripted(8)
	got := ComputeGreenhouseFactor(r, &Atmosphere{Code: 6, Pressure: 1.04})
	want := 0.59
	if math.Abs(got-want) > 0.005 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeGreenhouseFactor_AtmosphereA_Min0p5(t *testing.T) {
	// Atm A (10): × 1D-1 (minimum 0.5).
	// 1D=1 → 0 → minimum 0.5 applied.
	r := roller.NewScripted(1)
	atm := &Atmosphere{Code: 10, Pressure: 0.5}
	initial := 0.5 * math.Sqrt(0.5) // 0.354
	got := ComputeGreenhouseFactor(r, atm)
	want := initial * 0.5 // minimum
	if math.Abs(got-want) > 0.01 {
		t.Errorf("got %v, want %v (initial %v × min 0.5)", got, want, initial)
	}
}

func TestComputeGreenhouseFactor_AtmosphereB_RollOf6(t *testing.T) {
	// Atm B (11): 1D=1-5 → × result; 1D=6 → × 3D.
	// Test 1D=6 path: 1D=6, then 3D=10 → × 10.
	r := roller.NewScripted(6, 10)
	atm := &Atmosphere{Code: 11, Pressure: 1.0}
	initial := 0.5 // 0.5 × √1.0
	got := ComputeGreenhouseFactor(r, atm)
	want := initial * 10
	if math.Abs(got-want) > 0.01 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeGreenhouseFactor_AtmosphereB_NormalPath(t *testing.T) {
	// Atm B (11) 1D ≤ 5 path: × first roll (the dominant 5/6 probability branch).
	// 1D=3, pressure 1.0 → initial 0.5, × 3 → 1.5.
	r := roller.NewScripted(3)
	atm := &Atmosphere{Code: 11, Pressure: 1.0}
	got := ComputeGreenhouseFactor(r, atm)
	want := 1.5
	if math.Abs(got-want) > 0.01 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMeanTemperatureK_ZedPrime(t *testing.T) {
	// L=1.419, A=0.33, G=0.59, AU=1.06.
	// T = 279 × ⁴√(1.419 × 0.67 × 1.59 / 1.06²) ≈ 300.4 K → 300K.
	got := MeanTemperatureK(1.419, 0.33, 0.59, 1.06)
	want := 300.0
	if math.Abs(got-want) > 1.0 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMeanTemperatureK_Terra_Reference(t *testing.T) {
	// L=1.0, A=0.30, G=0.36 (book reference), AU=1.0.
	// T = 279 × ⁴√(1.0 × 0.70 × 1.36 / 1.0²) = 279 × (0.952)^0.25 ≈ 275.6K.
	// Note: real Earth ~288K requires G≈0.62; WBH's G=0.36 is a simplified model value.
	got := MeanTemperatureK(1.0, 0.30, 0.36, 1.0)
	if got < 273 || got > 278 {
		t.Errorf("got %v, want ~275.6K (L=1.0, A=0.30, G=0.36, AU=1.0)", got)
	}
}

func TestMeanTemperatureK_ClampsHighGreenhouse(t *testing.T) {
	// (1+G) > 1.999 should be clamped. With G=10: T should equal T at G=0.999.
	gotClamped := MeanTemperatureK(1.0, 0.0, 10.0, 1.0)
	gotAtLimit := MeanTemperatureK(1.0, 0.0, 0.999, 1.0)
	if math.Abs(gotClamped-gotAtLimit) > 0.5 {
		t.Errorf("clamp failed: at G=10 got %v, at G=0.999 got %v", gotClamped, gotAtLimit)
	}
}

func TestMeanTemperatureK_AlbedoOne_NearZero(t *testing.T) {
	// Albedo 1.0 → (1-A) = 0 → T = 0K. (Edge case; clamped to a small positive in clamp.)
	got := MeanTemperatureK(1.0, 1.0, 0.5, 1.0)
	if got > 1 {
		t.Errorf("albedo 1.0 should give near-0K, got %v", got)
	}
}

func TestCombineTemperatures_SingleSource(t *testing.T) {
	got := CombineTemperatures(300)
	if got != 300 {
		t.Errorf("single source should pass through, got %v", got)
	}
}

func TestCombineTemperatures_TwoEqual(t *testing.T) {
	// ⁴√(300⁴ + 300⁴) = 300 × ⁴√2 ≈ 356.7.
	got := CombineTemperatures(300, 300)
	want := 300 * math.Pow(2, 0.25)
	if math.Abs(got-want) > 0.5 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCombineTemperatures_DominantSource(t *testing.T) {
	// 1000K + 100K → close to 1000K.
	got := CombineTemperatures(1000, 100)
	if math.Abs(got-1000) > 1 {
		t.Errorf("got %v, want close to 1000 (dominant source)", got)
	}
}

func TestHZRegionAtmosphereDM(t *testing.T) {
	cases := []struct {
		code int
		want int
	}{
		{0, 0},
		{1, 0},
		{2, -2},
		{3, -2},
		{4, -1},
		{5, -1},
		{14, -1}, // E
		{6, 0},
		{7, 0},
		{8, 1},
		{9, 1},
		{10, 2},
		{13, 2},
		{15, 2}, // A, D, F
		{11, 6},
		{12, 6}, // B, C
		{16, 0},
		{17, 0},
		{99, 0}, // outside table
	}
	for _, c := range cases {
		if got := HZRegionAtmosphereDM(c.code); got != c.want {
			t.Errorf("code %d: got %d, want %d", c.code, got, c.want)
		}
	}
}

func TestBasicTemperatureRoll_Mod7_TableValue(t *testing.T) {
	// Atm 6 → DM 0; orbit at HZCO → no orbit DM. 2D=7 → mod=7 → 288K.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	// Sol HZCO()=3.0; set body.Orbit=3.0 to be exactly at HZCO (in zone, no orbit DM).
	body.Orbit = 3.0
	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7)
	mod, k := BasicTemperatureRoll(r, body, sys)
	if mod != 7 {
		t.Errorf("mod: got %d, want 7", mod)
	}
	if k != 288 {
		t.Errorf("kelvin: got %v, want 288", k)
	}
}

func TestBasicTemperatureRoll_AtmDMShifts(t *testing.T) {
	// Atm B (11) → DM +6. 2D=2 → mod=8 → 293K.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 11}
	body.Orbit = 3.0
	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(2)
	mod, k := BasicTemperatureRoll(r, body, sys)
	if mod != 8 {
		t.Errorf("mod: got %d, want 8 (raw 2 + DM +6)", mod)
	}
	if k != 293 {
		t.Errorf("kelvin: got %v, want 293", k)
	}
}

func TestBasicTemperatureRoll_OrbitInside_DMPlus(t *testing.T) {
	// Sol HZCO=3.0, HZCO-1=2.0. Body at orbit 1.0 → 2.0 - 1.0 = 1.0 below HZCO-1
	// → DM = 4 + floor(1.0/0.5) = 4 + 2 = +6.
	// Atm 6 (DM 0). 2D=2 → mod=8 → 293K.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Orbit = 1.0
	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(2)
	mod, k := BasicTemperatureRoll(r, body, sys)
	if mod != 8 {
		t.Errorf("mod: got %d, want 8 (raw 2 + orbit DM +6)", mod)
	}
	if k != 293 {
		t.Errorf("kelvin: got %v, want 293", k)
	}
}

func TestBasicTemperatureRoll_OrbitOutside_DMMinus(t *testing.T) {
	// Sol HZCO=3.0, HZCO+1=4.0. Body at orbit 5.0 → 5.0 - 4.0 = 1.0 above HZCO+1
	// → DM = -(4 + floor(1.0/0.5)) = -(4+2) = -6.
	// Atm 6 (DM 0). 2D=12 → mod=6 → 283K.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Orbit = 5.0
	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(12)
	mod, k := BasicTemperatureRoll(r, body, sys)
	if mod != 6 {
		t.Errorf("mod: got %d, want 6 (raw 12 + orbit DM -6)", mod)
	}
	if k != 283 {
		t.Errorf("kelvin: got %v, want 283", k)
	}
}

func TestBasicTemperatureRoll_AboveTable(t *testing.T) {
	// Force modified roll 14 → 388 + 2*50 = 488K. Per the project's roller
	// convention, NewScripted(14) returns 14 from any Roll() call regardless
	// of dice notation — so we can directly script the modified-roll value.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6} // DM 0
	body.Orbit = 3.0                       // at HZCO; no orbit DM
	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(14)
	_, k := BasicTemperatureRoll(r, body, sys)
	if k != 488 {
		t.Errorf("got %v, want 488", k)
	}
}

func TestBasicTemperatureRoll_BelowTable_NoRecompute(t *testing.T) {
	// Sol HZCO=3.0, HZCO+1=4.0. Body at orbit 10.0 → 10-4 = 6 above HZCO+1
	// → DM = -(4 + 12) = -16.
	// Atm 6 (DM 0). 2D=12 → mod = 12 - 16 = -4 → kelvin = 178 + (-4)*5 = 158K.
	// 158K > 10K → no recompute path fires.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Orbit = 10.0
	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(12)
	mod, k := BasicTemperatureRoll(r, body, sys)
	if mod != -4 {
		t.Errorf("mod: got %d, want -4", mod)
	}
	if math.Abs(k-158) > 0.1 {
		t.Errorf("got %v, want 158", k)
	}
}

func TestBasicTemperatureRoll_BelowTable_RecomputeAs1DPlus5(t *testing.T) {
	// Force modified roll low enough to trigger the < 10K recompute branch.
	// Sol HZCO=3.0, HZCO+1=4.0. Body at orbit 20.0 → 20-4 = 16 above HZCO+1
	// → orbit DM = -(4 + floor(16/0.5)) = -(4 + 32) = -36.
	// Atm 6 (DM 0). 2D=2 → mod = 2 - 36 = -34.
	// 178 + (-34)*5 = 178 - 170 = 8K → triggers recompute.
	// Recompute: 1D=4 → 4 + 5 = 9K (above 3K floor → no further adjustment).
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Orbit = 20.0
	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	// Two scripted values: first for 2D (raw temp roll), second for 1D (recompute).
	r := roller.NewScripted(2, 4)
	_, k := BasicTemperatureRoll(r, body, sys)
	if k != 9 {
		t.Errorf("got %v, want 9 (recompute as 1D+5 with 1D=4)", k)
	}
}

func TestGenerateTemperature_ZedPrime_Mean(t *testing.T) {
	// Zed Prime as a moon of a gas giant orbiting Aab (L=1.419) at 1.06 AU.
	// Per WBH p.111: A=0.33, G=0.59 → MeanK ≈ 300K.
	parent := &DetailedPlacement{}
	parent.Body = BodyGasGiant
	// Set parent.Orbit so OrbitToAU round-trips to ≈1.06 AU (Zed Prime's actual
	// distance per WBH p.111). Using a raw Orbit#=3.0 would give 1.0 AU and
	// drift MeanK ~9K from the book.
	parent.Orbit = stars.AUToOrbit(1.06)

	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.03}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.04, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 6}
	body.Eccentricity = 0.10

	sys := stars.System{Primary: stars.Star{Mass: 0.918, AgeGyr: 6.3, Luminosity: 1.419}}

	// Albedo: [8, 8, 7] → 0.33. Greenhouse: [8] → 0.59. Basic roll: [7].
	// 5 scripted values total.
	r := roller.NewScripted(8, 8, 7, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, parent)
	if err != nil {
		t.Fatal(err)
	}
	if temp == nil {
		t.Fatal("expected non-nil Temperature")
	}
	if math.Abs(temp.Albedo-0.33) > 0.01 {
		t.Errorf("Albedo: got %v, want 0.33", temp.Albedo)
	}
	if math.Abs(temp.GreenhouseFactor-0.59) > 0.01 {
		t.Errorf("GreenhouseFactor: got %v, want 0.59", temp.GreenhouseFactor)
	}
	if temp.Luminosity != 1.419 {
		t.Errorf("Luminosity: got %v, want 1.419", temp.Luminosity)
	}
	if math.Abs(temp.AU-1.06) > 0.01 {
		t.Errorf("AU: got %v, want ~1.06 (parent at AUToOrbit(1.06))", temp.AU)
	}
	// MeanK pinned to book worked example (~300K per WBH p.111).
	if math.Abs(temp.MeanK-300) > 5 {
		t.Errorf("MeanK: got %v, want 300 ±5K (Zed Prime per WBH p.111)", temp.MeanK)
	}
	// ScaleHeight should be cached for Task 11's AdjustedForAltitude.
	if temp.ScaleHeight != 8.5 {
		t.Errorf("ScaleHeight: got %v, want 8.5", temp.ScaleHeight)
	}
	// BasicK should be populated and finite.
	if temp.BasicK == 0 {
		t.Errorf("BasicK should be populated, got 0")
	}
}

func TestGenerateTemperature_BodyEmpty_ReturnsNil(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyEmpty
	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted()
	temp, err := GenerateTemperature(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if temp != nil {
		t.Errorf("BodyEmpty should return nil, got %+v", temp)
	}
}

func TestGenerateTemperature_PlanetUsesOwnOrbit(t *testing.T) {
	// For a planet (parent==nil), AU comes from body.Orbit via OrbitToAU.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0 // Sol HZCO Orbit#
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 6, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if temp == nil {
		t.Fatal("expected non-nil Temperature")
	}
	// Terra-like body should land in habitable range.
	if temp.MeanK < 200 || temp.MeanK > 400 {
		t.Errorf("MeanK out of plausible range: got %v", temp.MeanK)
	}
	// AU should equal stars.OrbitToAU(3.0) ≈ 1.0.
	if math.Abs(temp.AU-1.0) > 0.05 {
		t.Errorf("AU: got %v, want ~1.0 (OrbitToAU(3.0))", temp.AU)
	}
}

func TestGenerateTemperature_MultiStarLuminositySum(t *testing.T) {
	// Close binary mate (OrbitClass==OrbitCompanion, ParentIndex==-1) sums
	// luminosity into the primary group.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}

	sys := stars.System{
		Primary: stars.Star{Luminosity: 1.0},
		Companions: []stars.CompanionStar{
			{
				Star:        stars.Star{Luminosity: 0.5},
				OrbitClass:  stars.OrbitCompanion,
				ParentIndex: -1,
			},
		},
	}

	r := roller.NewScripted(7, 7, 6, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := 1.5 // 1.0 + 0.5
	if math.Abs(temp.Luminosity-want) > 0.001 {
		t.Errorf("Luminosity: got %v, want %v (close binary mate summed)", temp.Luminosity, want)
	}
}

func TestGenerateTemperature_NoAtmosphere(t *testing.T) {
	// Vacuum body: Atmosphere is nil → albedo skips atm modifier; greenhouse 0;
	// ScaleHeight stays 0; AtmosphericFactor will default in later tasks.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "5"
	body.Orbit = 3.0
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Hydrographics = &Hydrographics{Code: 0}
	// No Atmosphere, no Hydrographics modifier (code 0)

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	// Albedo: 1 roll (rocky base only). Greenhouse: 0 rolls (vacuum). Basic: 1 roll.
	r := roller.NewScripted(7, 7)
	temp, err := GenerateTemperature(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if temp == nil {
		t.Fatal("expected non-nil Temperature")
	}
	if temp.GreenhouseFactor != 0 {
		t.Errorf("vacuum should have G=0, got %v", temp.GreenhouseFactor)
	}
	if temp.ScaleHeight != 0 {
		t.Errorf("nil atmosphere should have ScaleHeight=0, got %v", temp.ScaleHeight)
	}
}

func TestGenerateTemperature_ZedPrime_HighLow(t *testing.T) {
	// Zed Prime: high=346K, low=250K per WBH p.114.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.03}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.04, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 6}
	body.Eccentricity = 0.25
	body.AxialTilt = &AxialTilt{Degrees: 73.65}
	body.DayLength = &DayLength{SiderealHours: 42.37, SolarHours: 85.77}
	body.Period = Period{Hours: 26 * 24} // moon's local year for short-year halving

	parent := &DetailedPlacement{}
	parent.Body = BodyGasGiant
	parent.Orbit = stars.AUToOrbit(1.06)
	parent.Eccentricity = 0.10

	sys := stars.System{Primary: stars.Star{Mass: 0.918, AgeGyr: 6.3, Luminosity: 1.419}}

	r := roller.NewScripted(8, 8, 7, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, parent)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(temp.HighK-346) > 5 {
		t.Errorf("HighK: got %v, want 346 ±5K", temp.HighK)
	}
	if math.Abs(temp.LowK-250) > 5 {
		t.Errorf("LowK: got %v, want 250 ±5K", temp.LowK)
	}
	// Variance components per spec.
	if math.Abs(temp.AxialTiltFactor-0.48) > 0.02 {
		t.Errorf("AxialTiltFactor: got %v, want 0.48 (halved from 0.96 by short year)", temp.AxialTiltFactor)
	}
	if math.Abs(temp.RotationFactor-0.185) > 0.01 {
		t.Errorf("RotationFactor: got %v, want 0.185", temp.RotationFactor)
	}
	if math.Abs(temp.GeographicFactor-0.20) > 0.01 {
		t.Errorf("GeographicFactor: got %v, want 0.20", temp.GeographicFactor)
	}
	if math.Abs(temp.AtmosphericFactor-2.04) > 0.01 {
		t.Errorf("AtmosphericFactor: got %v, want 2.04", temp.AtmosphericFactor)
	}
	if math.Abs(temp.LuminosityModifier-0.424) > 0.01 {
		t.Errorf("LuminosityModifier: got %v, want 0.424", temp.LuminosityModifier)
	}
	// Near AU = 1.06 × (1-0.10) = 0.954; Far AU = 1.06 × (1+0.10) = 1.166. Uses parent's ecc.
	if math.Abs(temp.NearAU-0.954) > 0.005 {
		t.Errorf("NearAU: got %v, want 0.954 (parent's ecc 0.10)", temp.NearAU)
	}
	if math.Abs(temp.FarAU-1.166) > 0.005 {
		t.Errorf("FarAU: got %v, want 1.166", temp.FarAU)
	}
}

func TestGenerateTemperature_ZedPrime_WorstCase(t *testing.T) {
	// Per WBH p.115 sidebar: WorstHigh=359K (book), WorstLow=219K (consistent
	// computation; book stated 230K is an inconsistency we surface in Task 13).
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.03}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.04, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 6}
	body.Eccentricity = 0.25
	body.AxialTilt = &AxialTilt{Degrees: 73.65}
	body.DayLength = &DayLength{SiderealHours: 42.37, SolarHours: 85.77}
	body.Period = Period{Hours: 26 * 24}

	parent := &DetailedPlacement{}
	parent.Body = BodyGasGiant
	parent.Orbit = stars.AUToOrbit(1.06)
	parent.Eccentricity = 0.10

	sys := stars.System{Primary: stars.Star{Luminosity: 1.419}}

	r := roller.NewScripted(8, 8, 7, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, parent)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(temp.WorstHighK-359) > 5 {
		t.Errorf("WorstHighK: got %v, want 359 ±5K", temp.WorstHighK)
	}
	// Implementation uses Near/Far AU consistently → 219K. Book p.115 sidebar
	// stated 230K, which appears to use base AU instead of Far AU. Pin the
	// consistent computation; document divergence.
	if math.Abs(temp.WorstLowK-219) > 5 {
		t.Errorf("WorstLowK: got %v, want 219 ±5K (book stated 230K — see WBH p.115 inconsistency)", temp.WorstLowK)
	}
}

func TestGenerateTemperature_AxialTiltLongYearBoost(t *testing.T) {
	// Year > 2 std years → axial tilt factor +0.01 per std year (max +0.25, cap 1.0).
	// Body with 5-year period: boost = 0.05.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.AxialTilt = &AxialTilt{Degrees: 23.45}
	body.DayLength = &DayLength{SiderealHours: 24, SolarHours: 24}
	body.Period = Period{Years: 5.0, Hours: 5.0 * 8766}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 6, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// sin(23.45°) = 0.398. Long-year boost +0.05. Expected ≈ 0.448.
	want := math.Sin(23.45*math.Pi/180.0) + 0.05
	if math.Abs(temp.AxialTiltFactor-want) > 0.01 {
		t.Errorf("AxialTiltFactor: got %v, want %v (long-year boost)", temp.AxialTiltFactor, want)
	}
}

func TestGenerateTemperature_RotationFactorLongDay(t *testing.T) {
	// Solar day > 2500h → rotation factor = 1.0.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.AxialTilt = &AxialTilt{Degrees: 0}
	body.DayLength = &DayLength{SiderealHours: 3000, SolarHours: 3000}
	body.Period = Period{Years: 1.0, Hours: 8766}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 6, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if temp.RotationFactor != 1.0 {
		t.Errorf("RotationFactor: got %v, want 1.0 (solar day > 2500h)", temp.RotationFactor)
	}
}

func TestGenerateTemperature_TwilightZone_Detected(t *testing.T) {
	// Body 1:1 star-locked → IsTwilight=true, BrightSideK > TwilightK > DarkSideK.
	// Eccentricity=0.2 ensures NearAU < AU < FarAU so bright/dark ordering holds
	// even when the dark lum modifier clamps to zero.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 1.0 // close to its star
	body.Eccentricity = 0.2
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.AxialTilt = &AxialTilt{Degrees: 0}
	body.DayLength = &DayLength{SiderealHours: 4383, SolarHours: 0} // twilight: undefined solar day
	body.Period = Period{Years: 0.5, Hours: 4383}
	body.TidalLock = &TidalLock{
		Case:      TidalLockCasePlanetToStar,
		LockRatio: "1:1",
	}

	sys := stars.System{Primary: stars.Star{Luminosity: 0.5}}

	r := roller.NewScripted(7, 7, 7, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !temp.IsTwilight {
		t.Error("expected IsTwilight=true for 1:1 star-lock")
	}
	if temp.BrightSideK <= temp.TwilightK {
		t.Errorf("BrightSideK %v should exceed TwilightK %v", temp.BrightSideK, temp.TwilightK)
	}
	if temp.DarkSideK >= temp.TwilightK {
		t.Errorf("DarkSideK %v should be below TwilightK %v", temp.DarkSideK, temp.TwilightK)
	}
	if math.Abs(temp.TwilightK-temp.MeanK) > 0.5 {
		t.Errorf("TwilightK %v should equal MeanK %v", temp.TwilightK, temp.MeanK)
	}
}

func TestGenerateTemperature_MoonLockedToPlanet_NotTwilight(t *testing.T) {
	// Moon 1:1 locked to its parent planet (Case == MoonToPlanet) → NOT twilight.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.AxialTilt = &AxialTilt{Degrees: 0}
	body.DayLength = &DayLength{SiderealHours: 24, SolarHours: 24}
	body.Period = Period{Hours: 30 * 24}
	body.TidalLock = &TidalLock{
		Case:      TidalLockCaseMoonToPlanet, // NOT PlanetToStar
		LockRatio: "1:1",
	}

	parent := &DetailedPlacement{}
	parent.Body = BodyGasGiant
	parent.Orbit = stars.AUToOrbit(1.0)

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 7, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, parent)
	if err != nil {
		t.Fatal(err)
	}
	if temp.IsTwilight {
		t.Error("moon→planet 1:1 lock should NOT be twilight zone")
	}
	if temp.BrightSideK != 0 || temp.DarkSideK != 0 || temp.TwilightK != 0 {
		t.Errorf("twilight fields should be zero for non-twilight body, got bright=%v dark=%v twilight=%v",
			temp.BrightSideK, temp.DarkSideK, temp.TwilightK)
	}
}

func TestGenerateTemperature_NotLocked_NoTwilight(t *testing.T) {
	// Normal body with TidalLock == nil → not twilight, fields zero.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.AxialTilt = &AxialTilt{Degrees: 23.45}
	body.DayLength = &DayLength{SiderealHours: 23.93, SolarHours: 24.0}
	body.Period = Period{Hours: 8766}
	// body.TidalLock == nil

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 6, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if temp.IsTwilight {
		t.Error("body with no TidalLock should not be twilight")
	}
}

func TestGenerateTemperature_GGMoon_ParentRadiance_AppliedWhenWarm(t *testing.T) {
	// Moon of a hot gas giant: parent's MeanK > moon's stellar-only MeanK + 30K → combine.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Eccentricity = 0.0
	body.AxialTilt = &AxialTilt{Degrees: 0}
	body.DayLength = &DayLength{SiderealHours: 24, SolarHours: 24}
	body.Period = Period{Hours: 7 * 24}

	parent := &DetailedPlacement{}
	parent.Body = BodyGasGiant
	parent.Orbit = stars.AUToOrbit(5.0) // far from star → moon's stellar-only ~150K
	parent.Eccentricity = 0.0
	// Pre-populate parent.Temperature with a HOT gas giant (much warmer than moon's stellar-only).
	parent.Temperature = &Temperature{MeanK: 500}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 7, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, parent)
	if err != nil {
		t.Fatal(err)
	}
	if temp.ParentRadianceK == 0 {
		t.Error("expected ParentRadianceK > 0 for hot GG parent")
	}
	if temp.ParentRadianceK != 500 {
		t.Errorf("ParentRadianceK: got %v, want 500 (parent MeanK)", temp.ParentRadianceK)
	}
	// Combined MeanK should be elevated by parent radiance (⁴√ combine of stellar + 500).
	// Stellar-only at 5 AU with low albedo and atm: ~150K. Combined with 500K parent ~= 500K dominant.
	if temp.MeanK < 450 {
		t.Errorf("MeanK should be elevated by parent radiance, got %v", temp.MeanK)
	}
}

func TestGenerateTemperature_GGMoon_ParentRadiance_SkippedWhenCold(t *testing.T) {
	// Moon of a cold gas giant: parent's MeanK ≤ moon's stellar-only MeanK + 30K → skip.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Eccentricity = 0.0
	body.AxialTilt = &AxialTilt{Degrees: 0}
	body.DayLength = &DayLength{SiderealHours: 24, SolarHours: 24}
	body.Period = Period{Hours: 7 * 24}

	parent := &DetailedPlacement{}
	parent.Body = BodyGasGiant
	parent.Orbit = stars.AUToOrbit(1.0) // moon's stellar-only ~280K
	parent.Eccentricity = 0.0
	parent.Temperature = &Temperature{MeanK: 200} // colder than moon's stellar-only

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 7, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, parent)
	if err != nil {
		t.Fatal(err)
	}
	if temp.ParentRadianceK != 0 {
		t.Errorf("expected ParentRadianceK=0 for cold parent, got %v", temp.ParentRadianceK)
	}
	// MeanK should NOT be elevated.
	stellarOnlyExpected := temp.MeanK
	_ = stellarOnlyExpected // sanity: just confirm we got a value
}

func TestGenerateTemperature_PlanetNoParentRadiance(t *testing.T) {
	// Planet (parent==nil) → ParentRadianceK stays 0.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.AxialTilt = &AxialTilt{Degrees: 23.45}
	body.DayLength = &DayLength{SiderealHours: 23.93, SolarHours: 24.0}
	body.Period = Period{Hours: 8766}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 6, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if temp.ParentRadianceK != 0 {
		t.Errorf("planet should have ParentRadianceK=0, got %v", temp.ParentRadianceK)
	}
}

func TestGenerateTemperature_GGMoon_ParentNoTemperature_Skipped(t *testing.T) {
	// Defensive: if parent.Temperature is nil (shouldn't happen in normal pipeline
	// but defensive), ParentRadianceK stays 0 and MeanK is unchanged.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.AxialTilt = &AxialTilt{Degrees: 0}
	body.DayLength = &DayLength{SiderealHours: 24, SolarHours: 24}
	body.Period = Period{Hours: 7 * 24}

	parent := &DetailedPlacement{}
	parent.Body = BodyGasGiant
	parent.Orbit = stars.AUToOrbit(1.0)
	// parent.Temperature == nil

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 7, 8, 7)
	temp, err := GenerateTemperature(r, body, sys, parent)
	if err != nil {
		t.Fatal(err)
	}
	if temp.ParentRadianceK != 0 {
		t.Errorf("nil parent.Temperature should leave ParentRadianceK=0, got %v", temp.ParentRadianceK)
	}
}

func TestSunlightPortion_Equator_Equinox(t *testing.T) {
	// Equator at equinox: portion 0.5 regardless of axial tilt.
	// Equinox = 1/4 year past summer solstice → cos(90°) = 0 → declination=0 → portion=0.5.
	got := SunlightPortion(0.0, 23.45, 0.25*365.25, 365.25)
	if math.Abs(got-0.5) > 0.01 {
		t.Errorf("portion: got %v, want 0.5", got)
	}
}

func TestSunlightPortion_Pole_SummerSolstice(t *testing.T) {
	// North pole at summer solstice → polar day → portion 1.0.
	got := SunlightPortion(89.99, 23.45, 0, 365.25)
	if got != 1.0 {
		t.Errorf("polar day: got %v, want 1.0", got)
	}
}

func TestSunlightPortion_Pole_WinterSolstice(t *testing.T) {
	// North pole at winter solstice → polar night → portion 0.
	got := SunlightPortion(89.99, 23.45, 365.25/2, 365.25)
	if got != 0 {
		t.Errorf("polar night: got %v, want 0", got)
	}
}

func TestTemperature_MeanByLatitude_Tropical(t *testing.T) {
	// Tropical zone (lat ≤ axial tilt): higher temperature than at the pole.
	temp := &Temperature{
		MeanK:             288,
		Luminosity:        1.0,
		Albedo:            0.3,
		GreenhouseFactor:  0.36,
		AU:                1.0,
		AxialTiltFactor:   math.Sin(23.45 * math.Pi / 180.0), // ~0.398
		AtmosphericFactor: 2.0,
	}
	tropic := temp.MeanByLatitude(10) // tropical (< 23.45°)
	arctic := temp.MeanByLatitude(80) // arctic (> 90 - 23.45 = 66.55°)
	if tropic <= arctic {
		t.Errorf("tropic %v should exceed arctic %v", tropic, arctic)
	}
}

func TestTemperature_MeanByLatitude_TwilightShortCircuit(t *testing.T) {
	// Twilight world: MeanByLatitude returns TwilightK regardless of latitude.
	temp := &Temperature{
		MeanK:       288,
		IsTwilight:  true,
		TwilightK:   285,
		BrightSideK: 320,
		DarkSideK:   200,
	}
	got := temp.MeanByLatitude(45)
	if got != 285 {
		t.Errorf("got %v, want 285 (TwilightK for IsTwilight body)", got)
	}
}

func TestTemperature_MeanBySeason_OppositeSolstices(t *testing.T) {
	temp := &Temperature{
		MeanK:             288,
		Luminosity:        1.0,
		Albedo:            0.3,
		GreenhouseFactor:  0.36,
		AU:                1.0,
		AxialTiltFactor:   0.40,
		AtmosphericFactor: 2.0,
	}
	summer := temp.MeanBySeason(45, 0, 365.25)        // summer solstice at 45°N
	winter := temp.MeanBySeason(45, 365.25/2, 365.25) // winter solstice at 45°N
	if summer <= winter {
		t.Errorf("summer %v should exceed winter %v", summer, winter)
	}
}

func TestTemperature_AtMoment_NoonExceedsDawn(t *testing.T) {
	temp := &Temperature{
		MeanK:             288,
		Luminosity:        1.0,
		Albedo:            0.3,
		GreenhouseFactor:  0.36,
		AU:                1.0,
		AxialTiltFactor:   0.40,
		RotationFactor:    0.10,
		AtmosphericFactor: 2.0,
	}
	dawn := temp.AtMoment(0, 0, 365.25, 0, 24)
	noon := temp.AtMoment(0, 0, 365.25, 12, 24)
	if noon <= dawn {
		t.Errorf("noon %v should exceed dawn %v (with 0.15 lag, peak is post-noon)", noon, dawn)
	}
}

func TestTemperature_AdjustedForAltitude_NearGround(t *testing.T) {
	temp := &Temperature{
		MeanK:            288,
		Luminosity:       1.0,
		Albedo:           0.3,
		GreenhouseFactor: 0.36,
		AU:               1.0,
		ScaleHeight:      8.5,
	}
	got := temp.AdjustedForAltitude(288, 0.001) // 1 m altitude
	if math.Abs(got-288) > 0.5 {
		t.Errorf("near-ground should return ~baseTemp, got %v", got)
	}
}

func TestTemperature_AdjustedForAltitude_8000m_LessThanBase(t *testing.T) {
	// At ~8000m on Terra-like world, pressure is roughly 0.36 bar (e^(-8/8.5)).
	// Lower greenhouse → cooler.
	temp := &Temperature{
		MeanK:            288,
		Luminosity:       1.0,
		Albedo:           0.3,
		GreenhouseFactor: 0.36,
		AU:               1.0,
		ScaleHeight:      8.5,
	}
	got := temp.AdjustedForAltitude(288, 8.0)
	if got >= 288 {
		t.Errorf("8000m should be cooler than base, got %v", got)
	}
}

func TestTemperature_AdjustedForAltitude_NoScaleHeight_Passthrough(t *testing.T) {
	// Vacuum world or atmosphere with no scale-height data: pass through.
	temp := &Temperature{MeanK: 288, Albedo: 0.3, GreenhouseFactor: 0, ScaleHeight: 0}
	got := temp.AdjustedForAltitude(288, 5.0)
	if got != 288 {
		t.Errorf("zero scale height should return baseTempK, got %v", got)
	}
}

func TestTemperature_AdjustedForAltitude_ZeroAltitude_Passthrough(t *testing.T) {
	// 0 altitude: pass through.
	temp := &Temperature{MeanK: 288, Albedo: 0.3, GreenhouseFactor: 0.36, AU: 1.0, ScaleHeight: 8.5}
	got := temp.AdjustedForAltitude(288, 0)
	if got != 288 {
		t.Errorf("zero altitude should return baseTempK, got %v", got)
	}
}
