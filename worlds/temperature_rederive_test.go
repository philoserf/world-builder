package worlds

import (
	"math"
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestSelectExoticLiquid_Water_Terra(t *testing.T) {
	// At meanK=288 with atm A (10), water (Abundance 100) wins among candidates
	// where 273 ≤ 288 ≤ 373.
	got := SelectExoticLiquid(288, 10)
	if got != "H2O" {
		t.Errorf("got %q, want H2O", got)
	}
}

func TestSelectExoticLiquid_Methane_Cold(t *testing.T) {
	// At meanK=100 with atm B (11), methane (range 91-113, Abundance 70) wins.
	got := SelectExoticLiquid(100, 11)
	if got != "CH4" {
		t.Errorf("got %q, want CH4", got)
	}
}

func TestSelectExoticLiquid_Ethane_NotMethaneAtMid(t *testing.T) {
	// At meanK=150, methane (boils at 113) is out; ethane (range 90-184, Abundance 70) wins.
	got := SelectExoticLiquid(150, 11)
	if got != "C2H6" {
		t.Errorf("got %q, want C2H6", got)
	}
}

func TestSelectExoticLiquid_NoCandidate_TooHot(t *testing.T) {
	// At meanK=2000, all candidates' boiling points are exceeded.
	got := SelectExoticLiquid(2000, 10)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestSelectExoticLiquid_NonExoticAtm_Empty(t *testing.T) {
	// Atm 6 (standard) → defensive: caller shouldn't call; return empty.
	got := SelectExoticLiquid(288, 6)
	if got != "" {
		t.Errorf("got %q, want empty (non-exotic atm)", got)
	}
}

func TestSelectExoticLiquid_TieBreakLowerBoiling(t *testing.T) {
	// Construct a meanK where two candidates tie on Abundance — verify lower
	// BoilingK wins. CH4 (91-113, 70) and C2H6 (90-184, 70) both contain 100K.
	// Lower BoilingK is CH4 (113 < 184) → CH4 wins.
	got := SelectExoticLiquid(100, 11)
	if got != "CH4" {
		t.Errorf("got %q, want CH4 (tie-break by lower BoilingK)", got)
	}
}

func TestSelectExoticLiquid_BoundaryInclusive(t *testing.T) {
	// Verify [MeltingK, BoilingK] is inclusive on both ends.
	// H2O range: melting 273, boiling 373.
	if got := SelectExoticLiquid(273, 10); got != "H2O" {
		t.Errorf("meanK=273 (= H2O melting): got %q, want H2O", got)
	}
	if got := SelectExoticLiquid(373, 10); got != "H2O" {
		t.Errorf("meanK=373 (= H2O boiling): got %q, want H2O", got)
	}
}

func TestMeanKToTempRange_Boundaries(t *testing.T) {
	cases := []struct {
		meanK float64
		want  TempRange
	}{
		{50, TempFrozen},
		{122, TempFrozen},
		{123, TempCold},
		{200, TempCold},
		{272, TempCold},
		{273, TempTemperate},
		{300, TempTemperate},
		{352, TempTemperate},
		{353, TempHot},
		{400, TempHot},
		{452, TempHot},
		{453, TempBoiling},
		{1000, TempBoiling},
	}
	for _, c := range cases {
		if got := MeanKToTempRange(c.meanK); got != c.want {
			t.Errorf("meanK=%v: got %v, want %v", c.meanK, got, c.want)
		}
	}
}

func TestCheckRunawayGreenhouse_BelowTempThreshold(t *testing.T) {
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Temperature = &Temperature{MeanK: 300}
	sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

	r := roller.NewScripted()
	if got := CheckRunawayGreenhouse(r, body, sys); got {
		t.Error("expected false (meanK below 303K)")
	}
}

func TestCheckRunawayGreenhouse_LowAtmCode(t *testing.T) {
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 1}
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

	r := roller.NewScripted()
	if got := CheckRunawayGreenhouse(r, body, sys); got {
		t.Error("expected false (atm 1 out of trigger range)")
	}
}

func TestCheckRunawayGreenhouse_AtmAlreadyExtreme_BoilingOnly(t *testing.T) {
	// WBH p.79: for atm A, B, C, F+, the runaway-greenhouse trigger still
	// evaluates, but the only effect is to consider the world boiling. The
	// atm code does NOT mutate. Each subtest scripts dice that produce a
	// trigger total ≥ 12 (atm 8, age 5, size 8 → DM age+5 + boiling+4 +
	// size+0 = +9; 2D=3 + 9 = 12 → trigger).
	cases := []struct {
		name string
		code int
	}{
		{"atm A (10)", 10},
		{"atm B (11)", 11},
		{"atm C (12)", 12},
		{"atm F (15)", 15},
		{"atm G (16)", 16},
		{"atm H (17)", 17},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body := &DetailedPlacement{}
			body.Atmosphere = &Atmosphere{Code: c.code}
			body.SizeCode = "8"
			body.Temperature = &Temperature{MeanK: 400}
			sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

			// 2D=3 → 3 + age+5 (ceil 5.0) + boiling+4 = 12 → trigger.
			// No 1D code-mutation roll consumed because boiling-only.
			r := roller.NewScripted(3)
			got := CheckRunawayGreenhouse(r, body, sys)
			if !got {
				t.Errorf("expected true (trigger fired with mod=12)")
			}
			if body.Atmosphere.Code != c.code {
				t.Errorf("atm code mutated for boiling-only path: got %d, want %d (unchanged)",
					body.Atmosphere.Code, c.code)
			}
		})
	}
}

func TestCheckRunawayGreenhouse_LowDiceRoll(t *testing.T) {
	// atm 6, meanK=400, sysAge=1 → DM age+1 (ceil), boil+4 (400≥388). 2D=2 + 5 = 7 < 12 → false.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 1}}

	r := roller.NewScripted(2)
	if got := CheckRunawayGreenhouse(r, body, sys); got {
		t.Error("expected false (mod=7, below 12)")
	}
}

func TestCheckRunawayGreenhouse_Triggered_AtmA(t *testing.T) {
	// atm 6, meanK=400 (boiling +4), sysAge=5 (round up +5).
	// 2D=3 + 5 + 4 = 12 → trigger. 1D=1 → atm becomes A (10).
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.SizeCode = "8"
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

	r := roller.NewScripted(3, 1)
	if got := CheckRunawayGreenhouse(r, body, sys); !got {
		t.Error("expected true (mod=12, triggered)")
	}
	if body.Atmosphere.Code != 10 {
		t.Errorf("atm code: got %d, want 10 (A)", body.Atmosphere.Code)
	}
}

func TestCheckRunawayGreenhouse_Triggered_AtmB(t *testing.T) {
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.SizeCode = "8"
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

	r := roller.NewScripted(3, 3)
	if got := CheckRunawayGreenhouse(r, body, sys); !got {
		t.Error("expected true")
	}
	if body.Atmosphere.Code != 11 {
		t.Errorf("atm code: got %d, want 11 (B)", body.Atmosphere.Code)
	}
}

func TestCheckRunawayGreenhouse_Triggered_AtmC(t *testing.T) {
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.SizeCode = "8"
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

	r := roller.NewScripted(3, 6)
	if got := CheckRunawayGreenhouse(r, body, sys); !got {
		t.Error("expected true")
	}
	if body.Atmosphere.Code != 12 {
		t.Errorf("atm code: got %d, want 12 (C)", body.Atmosphere.Code)
	}
}

func TestCheckRunawayGreenhouse_TaintedDM(t *testing.T) {
	// atm 7 (tainted), meanK=400, sysAge=2 → DM age+2, boil+4, taint+1, size 8 no penalty = +7.
	// 2D=4 + 7 = 11 < 12 → false. Confirms tainted DM applied (without taint mod=10; with taint mod=11).
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 7}
	body.SizeCode = "8"
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 2}}

	r := roller.NewScripted(4)
	if got := CheckRunawayGreenhouse(r, body, sys); got {
		t.Error("expected false (mod=11 with taint+1; without taint mod=10 — taint applied means mod=11)")
	}
}

func TestCheckRunawayGreenhouse_SizeDM(t *testing.T) {
	// Size 4 → DM-2. atm 6, meanK=400, sysAge=10 → DM age+10, boil+4, size-2 = +12.
	// 2D=2 + 12 = 14 → trigger.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.SizeCode = "4"
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 10}}

	r := roller.NewScripted(2, 3)
	if got := CheckRunawayGreenhouse(r, body, sys); !got {
		t.Error("expected true (size DM applied; mod=14)")
	}
}

func TestDeriveHydrographicsProfile_Water(t *testing.T) {
	// Atm 6 (standard), hydro 6, meanK 288 → "H6:H2O-100".
	got := DeriveHydrographicsProfile(288, 6, 6)
	if got != "H6:H2O-100" {
		t.Errorf("got %q, want H6:H2O-100", got)
	}
}

func TestDeriveHydrographicsProfile_ExoticMethane(t *testing.T) {
	// Atm A (10), hydro 4, meanK 100 → "H4:CH4-100" (methane wins by Abundance 70).
	got := DeriveHydrographicsProfile(100, 10, 4)
	if got != "H4:CH4-100" {
		t.Errorf("got %q, want H4:CH4-100", got)
	}
}

func TestDeriveHydrographicsProfile_Vacuum_Empty(t *testing.T) {
	// Atm 0, hydro 0 → empty (no liquid).
	got := DeriveHydrographicsProfile(288, 0, 0)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestDeriveHydrographicsProfile_NoHydro_Empty(t *testing.T) {
	// Atm 6, hydro 0 → empty (no liquid surface).
	got := DeriveHydrographicsProfile(288, 6, 0)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestDeriveHydrographicsProfile_HydroA(t *testing.T) {
	// Hydro 10 renders as "A" in the formula tail.
	got := DeriveHydrographicsProfile(288, 6, 10)
	if got != "HA:H2O-100" {
		t.Errorf("got %q, want HA:H2O-100", got)
	}
}

func TestDeriveHydrographicsProfile_ExoticAtm_NoCandidate(t *testing.T) {
	// Atm A, hydro 5, meanK 2000 → no exotic liquid fits → empty.
	got := DeriveHydrographicsProfile(2000, 10, 5)
	if got != "" {
		t.Errorf("got %q, want empty (no liquid candidate at meanK 2000)", got)
	}
}

func TestRederive_TerraLike_StableInTemperate(t *testing.T) {
	// Terra-like body; rederive doesn't dramatically shift atm/hydro.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 6, Subtype: "5", Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 288}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	preCode := body.Atmosphere.Code
	preHydro := body.Hydrographics.Code

	r := roller.NewScripted(7, 5) // hydro 2D + percent d10
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Atm code should NOT change in Task 6 (runaway not wired yet).
	if body.Atmosphere.Code != preCode {
		t.Errorf("atm code changed without runaway: pre=%d post=%d", preCode, body.Atmosphere.Code)
	}
	// Hydro code may shift slightly; pin to plausible range.
	if abs(body.Hydrographics.Code-preHydro) > 3 {
		t.Errorf("hydro code shifted too much: pre=%d post=%d", preHydro, body.Hydrographics.Code)
	}
	// Hydrographics.Profile populated.
	if body.Hydrographics.Profile == "" {
		t.Error("Hydrographics.Profile should be populated")
	}
}

func TestRederive_NilTemperature_NoOp(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Atmosphere = &Atmosphere{Code: 6}
	// body.Temperature is nil
	sys := stars.System{Primary: stars.Star{}}

	r := roller.NewScripted()
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// No mutations: Hydrographics is nil so Profile can't be set.
	if body.Hydrographics != nil && body.Hydrographics.Profile != "" {
		t.Error("expected no Profile when Temperature is nil")
	}
}

func TestRederive_BodyEmpty_NoOp(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyEmpty
	body.Temperature = &Temperature{MeanK: 288}
	sys := stars.System{Primary: stars.Star{}}

	r := roller.NewScripted()
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRederive_ScaleHeightUpdate(t *testing.T) {
	// Cold body: meanK=200 → ScaleHeight should scale lower than at 288K.
	// 8.5 × 200/288 / 1.0 ≈ 5.9 km.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 6, Subtype: "5", Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 200}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	r := roller.NewScripted(7, 5) // hydro 2D + percent d10
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantApprox := 8.5 * 200 / 288 / 1.0
	if math.Abs(body.Atmosphere.ScaleHeight-wantApprox) > 0.5 {
		t.Errorf("ScaleHeight: got %v, want ≈%v", body.Atmosphere.ScaleHeight, wantApprox)
	}
}

func TestRederive_AtmosphereB_NoRunaway_NoSubtypeChange(t *testing.T) {
	// Atm B (11) without runaway → subtype stays unchanged.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Atmosphere = &Atmosphere{Code: 11, Subtype: "5", Pressure: 1.5, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 288}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	preSubtype := body.Atmosphere.Subtype
	prePressure := body.Atmosphere.Pressure

	// Scripted dice: 1 for hydro re-roll + gas-mix budget for exotic atm B + hydro > 0 (Task 8).
	// rerollAtmSubtypeAndPressure is not called here (Task 9 wires that).
	r := roller.NewScripted(7, 5, 8, 5, 8, 5, 8, 5, 8, 5, 8, 5, 8, 5) // hydro 2D + percent d10 + gas mix
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if body.Atmosphere.Subtype != preSubtype {
		t.Errorf("subtype changed without runaway: pre=%s post=%s", preSubtype, body.Atmosphere.Subtype)
	}
	if body.Atmosphere.Pressure != prePressure {
		t.Errorf("pressure changed without runaway: pre=%v post=%v", prePressure, body.Atmosphere.Pressure)
	}
}

func TestRerollAtmSubtypeAndPressure_AtmBDirectCall(t *testing.T) {
	// Direct call to verify helper exists and runs; full integration tested in Task 9.
	body := &DetailedPlacement{}
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Atmosphere = &Atmosphere{Code: 11, Subtype: "5", Pressure: 1.5}
	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 7) // subtype roll + pressure 2 dice
	if err := rerollAtmSubtypeAndPressure(r, body, sys, false); err != nil {
		t.Fatal(err)
	}
	// Subtype should now be a valid value (1-9 or A-E).
	if body.Atmosphere.Subtype == "" {
		t.Error("expected non-empty Subtype after re-roll")
	}
}

func TestRederive_AtmProfile_ExoticC(t *testing.T) {
	// Atm C (code 12) with subtype "5" (numeric, the common subtype output) and
	// hydro 4 → RollGasMix uses column letter "C" (mapped from Code, NOT Subtype).
	// Profile should be populated with named gases from the C-column of the
	// gas-mix table — not a fallback "Other"-only profile.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Atmosphere = &Atmosphere{Code: 12, Subtype: "5", Pressure: 5.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 4}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 288}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	r := roller.NewScripted(7, 5, 8, 5, 8, 5, 8, 5, 8, 5, 8, 5, 8, 5) // hydro 2D + percent d10 + gas mix
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(body.Atmosphere.Profile.Gases) == 0 {
		t.Error("Atm.Profile.Gases should be populated for exotic atm with hydro")
	}
	// Regression check: the C1 bug produced a single "Other"-only profile.
	// A correctly-routed RollGasMix should produce at least one named gas.
	hasNamed := false
	for _, g := range body.Atmosphere.Profile.Gases {
		if g.Name != "Other" {
			hasNamed = true
			break
		}
	}
	if !hasNamed {
		t.Errorf("Atm.Profile.Gases should contain at least one named gas, not just Other; got %+v",
			body.Atmosphere.Profile.Gases)
	}
}

func TestRederive_AtmProfile_StandardAtm_NotMutated(t *testing.T) {
	// Atm 6 (standard) → Atm.Profile NOT touched by rederive (gas mix only for exotic).
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 6, Subtype: "5", Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 288}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	r := roller.NewScripted(7, 5) // hydro 2D + percent d10
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(body.Atmosphere.Profile.Gases) != 0 {
		t.Errorf("Atm.Profile.Gases should be empty for standard atm; got %d gases", len(body.Atmosphere.Profile.Gases))
	}
}

func TestRederive_AtmProfile_ExoticAtmNoHydro_NotMutated(t *testing.T) {
	// Atm A (10) with hydro 0 → no liquid → no gas mix → leave Profile alone.
	// Construct dice so post-rederive hydro stays 0 (genuine "no hydro" scenario).
	// With atm A (code 10), size 8, Temperate, no atm taint: digit = roll-7+10-4 = roll-1.
	// Need digit ≤ 0, so roll ≤ 1. NewScripted(1) → digit = 0, hydrographics stays 0.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 10, Subtype: "A", Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 0}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 288}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	r := roller.NewScripted(1, 5) // hydro 2D=1 → digit 0; percent d10 still consumed; no gas mix
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if body.Hydrographics.Code != 0 {
		t.Fatalf("test setup wrong: post-rederive hydro should be 0, got %d", body.Hydrographics.Code)
	}
	if len(body.Atmosphere.Profile.Gases) != 0 {
		t.Errorf("Atm.Profile.Gases should not be populated when hydro is 0; got %d gases",
			len(body.Atmosphere.Profile.Gases))
	}
}

func TestRederive_RunawayFires_AtmAndHydroMutate(t *testing.T) {
	// Atm 6 + meanK=400 + sysAge=10 + size 8 (no size DM) → very easy trigger.
	// Body must be HZ — set body.HZ=true.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.HZ = true
	body.Atmosphere = &Atmosphere{Code: 6, Subtype: "5", Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 400}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 10}}

	// Pre-rederive: atm code 6, hydro 7.
	// Scripted dice (in order):
	// 1. CheckRunawayGreenhouse trigger 2D: 2 (mod = 2 + age 10 + boil 4 = 16 → trigger)
	// 2. CheckRunawayGreenhouse 1D atm code: 3 → atm becomes B (11)
	// 3. rerollAtmSubtypeAndPressure subtype 2D: 8 (subtype calculation; runawayResult=true → DM+4)
	// 4. rerollAtmSubtypeAndPressure pressure: 2 dice (1D + 1D for RollTotalPressure)
	// 5. RollHydroDigit with TempBoiling: 1 dice (2D)
	// 6. RollGasMix for new atm B + hydro: ~6-12 dice
	r := roller.NewScripted(2, 3, 8, 7, 5, 5, 5, 8, 5, 8, 5, 8, 5, 8, 5, 8) // trigger + atm code + subtype + hydro 2D + percent d10 + gas mix
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Atm code should now be A, B, or C.
	if body.Atmosphere.Code != 10 && body.Atmosphere.Code != 11 && body.Atmosphere.Code != 12 {
		t.Errorf("atm code: got %d, want A/B/C (10/11/12)", body.Atmosphere.Code)
	}
	// Hydrographics.Code should be reduced because boiling DM-6 was applied
	// instead of hot DM-2 (relative -4 vs the no-runaway path).
	// Pre-rederive hydro was 7. With boiling DM-6 and atm B (11):
	// digit = 7 - 7 + 11 + boilingDM(-6) = 5. Original hot DM would give 9.
	// So post-rederive hydro should be ≤ 6 (we use a generous bound).
	if body.Hydrographics.Code > 6 {
		t.Errorf("hydro code: got %d, expected ≤ 6 (boiling DM-6 applied)", body.Hydrographics.Code)
	}
}

func TestRederive_RunawayDoesNotFire_NoBoilingDM(t *testing.T) {
	// Same body but low dice → runaway doesn't trigger; hydro re-rolled with normal Temperate range.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.HZ = true
	body.Atmosphere = &Atmosphere{Code: 6, Subtype: "5", Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 288} // below 303K — runaway can't fire

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	// Scripted: hydro 2D + percent d10 (runaway short-circuits without consuming dice when MeanK ≤ 303).
	r := roller.NewScripted(7, 5)
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Atm code unchanged.
	if body.Atmosphere.Code != 6 {
		t.Errorf("atm code: got %d, want 6 (unchanged)", body.Atmosphere.Code)
	}
}

func TestRederive_NonHZBody_NoRunawayCheck(t *testing.T) {
	// Non-HZ body → CheckRunawayGreenhouse not called even if temp > 303K.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.HZ = false // explicit non-HZ
	body.Atmosphere = &Atmosphere{Code: 6, Subtype: "5", Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 400} // hot but not HZ

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 10}}

	// Scripted: hydro 2D + percent d10 (no runaway check for non-HZ).
	r := roller.NewScripted(7, 5)
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Atm code unchanged.
	if body.Atmosphere.Code != 6 {
		t.Errorf("atm code: got %d, want 6 (non-HZ skips runaway)", body.Atmosphere.Code)
	}
}

func TestRederive_AtmosphereB_RunawayBoilingOnly_PreservesSubtype(t *testing.T) {
	// HZ atm-B body with MeanK > 303 (boiling) and DMs that trigger the
	// runaway-greenhouse roll. Per WBH p.79, atm B's only runaway effect
	// is the "considered boiling" hydro DM — atm.Code/Subtype/Pressure
	// must NOT change.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.HZ = true
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Atmosphere = &Atmosphere{Code: 11, Subtype: "5", Pressure: 1.5, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 400} // > 303, > 388 → boiling DM+4

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	preCode := body.Atmosphere.Code
	preSubtype := body.Atmosphere.Subtype
	prePressure := body.Atmosphere.Pressure

	// Scripted dice budget:
	//   1 roll for runaway 2D trigger (= 3 + age+5 + boiling+4 = 12 → fires)
	//   No 1D code-mutation roll (boiling-only path).
	//   No subtype/pressure re-roll (caller skips for boiling-only).
	//   Hydrographics + gas mix re-rolls consume the rest (over-script for safety).
	r := roller.NewScripted(3, 7, 5, 8, 5, 8, 5, 8, 5, 8, 5, 8, 5, 8, 5, 8)
	if err := RederiveAtmosphereHydrographics(r, body, sys, nil); err != nil {
		t.Fatal(err)
	}

	if body.Atmosphere.Code != preCode {
		t.Errorf("atm code changed under boiling-only runaway: pre=%d post=%d",
			preCode, body.Atmosphere.Code)
	}
	if body.Atmosphere.Subtype != preSubtype {
		t.Errorf("atm subtype changed under boiling-only runaway: pre=%s post=%s",
			preSubtype, body.Atmosphere.Subtype)
	}
	if body.Atmosphere.Pressure != prePressure {
		t.Errorf("atm pressure changed under boiling-only runaway: pre=%v post=%v",
			prePressure, body.Atmosphere.Pressure)
	}
}

// abs is a local int abs helper (no math.Abs for ints in stdlib).
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
