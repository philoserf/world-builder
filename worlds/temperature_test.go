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
