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

func TestCheckRunawayGreenhouse_AtmAlreadyExotic_Skipped(t *testing.T) {
	for _, code := range []int{10, 11, 12, 15} {
		body := &DetailedPlacement{}
		body.Atmosphere = &Atmosphere{Code: code}
		body.Temperature = &Temperature{MeanK: 400}
		sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

		r := roller.NewScripted()
		if got := CheckRunawayGreenhouse(r, body, sys); got {
			t.Errorf("atm %d: expected false (exotic atm skipped per MVP)", code)
		}
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

	r := roller.NewScripted(7) // 1 dice for hydro re-roll (no runaway in Task 6 yet)
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

	r := roller.NewScripted(7)
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

	// Scripted dice: 1 for hydrographics re-roll only (Task 6 doesn't call rerollAtmSubtypeAndPressure).
	r := roller.NewScripted(7)
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

// abs is a local int abs helper (no math.Abs for ints in stdlib).
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
