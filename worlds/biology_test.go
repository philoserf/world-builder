package worlds

import (
	"testing"

	"wbh/roller"
)

func TestRollBiomass_ZedPrime(t *testing.T) {
	// Atm 6 (no DM), Hydro 6 (+1), Age 6.3 (+1), MeanK 300 (+2), HighK 346 (no DM).
	// DMs total +4 (at cap), 2D=6 → 6+4 = 10 → biomass A.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 6}
	body.Temperature = &Temperature{MeanK: 300, HighK: 346}
	r := roller.NewScripted(6)
	got := RollBiomass(r, body, 6.3)
	if got != 10 {
		t.Errorf("Zed Prime: got %d, want 10", got)
	}
}

func TestRollBiomass_DMCap_AtPositiveCeiling(t *testing.T) {
	// Atm 8 (+2) + Hydro A (+2) + Age > 4 (+1) + MeanK 290 (+2) = +7, clamp +4.
	// 2D=10 → 10+4 = 14 → biomass E.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 8}
	body.Hydrographics = &Hydrographics{Code: 10}
	body.Temperature = &Temperature{MeanK: 290}
	r := roller.NewScripted(10)
	got := RollBiomass(r, body, 5.0)
	if got != 14 {
		t.Errorf("got %d, want 14 (DM cap +4)", got)
	}
}

func TestRollBiomass_DMCap_AtNegativeFloor(t *testing.T) {
	// Vacuum atm 0 (-6) + Hydro 0 (-4) + Age < 0.2 (-6) + MeanK 100 (-2) + HighK 100 (-4)
	// = -22, clamp to -12.
	// 2D=2 → 2-12 = -10 → biomass clamped to 0.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 0}
	body.Hydrographics = &Hydrographics{Code: 0}
	body.Temperature = &Temperature{MeanK: 100, HighK: 100}
	r := roller.NewScripted(2)
	got := RollBiomass(r, body, 0.1)
	if got != 0 {
		t.Errorf("got %d, want 0 (DM clamp -12; result < 0 → 0)", got)
	}
}

func TestRollBiomass_ExoticAtm_BonusApplied_AtmB(t *testing.T) {
	// Atm B (-5) + Hydro 6 (+1) + Age 5 (+1) + MeanK 290 (+2) = -1, no clamp needed.
	// 2D=8 → 8 - 1 = 7 (biomass ≥ 1).
	// Exotic-atm bonus for atm B: |−5| − 1 = +4 → final 11.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 11}
	body.Hydrographics = &Hydrographics{Code: 6}
	body.Temperature = &Temperature{MeanK: 290}
	r := roller.NewScripted(8)
	got := RollBiomass(r, body, 5.0)
	if got != 11 {
		t.Errorf("got %d, want 11 (atm B bonus +4 applied)", got)
	}
}

func TestRollBiomass_ExoticAtm_BonusSkipped_AtmBZero(t *testing.T) {
	// Atm B (-5), no other DMs, 2D=2 → 2-5 = -3 → biomass 0.
	// Bonus NOT applied because rolled biomass is 0.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 11}
	body.Hydrographics = &Hydrographics{Code: 0}
	r := roller.NewScripted(2)
	got := RollBiomass(r, body, 2.0)
	if got != 0 {
		t.Errorf("got %d, want 0 (bonus skipped when biomass=0)", got)
	}
}

func TestRollBiomass_VacuumAtm_BonusApplied(t *testing.T) {
	// Atm 0 (-6) + Hydro 9 (+2) + Age 5.0 (+1) = -3, no clamp.
	// 2D=12 → 12 - 3 = 9 (biomass ≥ 1).
	// Bonus for atm 0: |−6| − 1 = +5 → final 14.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 0}
	body.Hydrographics = &Hydrographics{Code: 9}
	r := roller.NewScripted(12)
	got := RollBiomass(r, body, 5.0)
	if got != 14 {
		t.Errorf("got %d, want 14 (atm 0 bonus +5 applied)", got)
	}
}

func TestRollBiomass_NilAtmosphere_Zero(t *testing.T) {
	body := &DetailedPlacement{}
	body.Hydrographics = &Hydrographics{Code: 5}
	r := roller.NewScripted(7)
	got := RollBiomass(r, body, 5.0)
	if got != 0 {
		t.Errorf("got %d, want 0 (nil atmosphere)", got)
	}
}

func TestRollBiomass_NilHydrographics_HydroZeroDM(t *testing.T) {
	// Atm 6 (no DM), nil hydro treated as DM-4, MeanK 290 (+2), Age 5 (+1) = -1.
	// 2D=10 → 10 - 1 = 9.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Temperature = &Temperature{MeanK: 290}
	r := roller.NewScripted(10)
	got := RollBiomass(r, body, 5.0)
	if got != 9 {
		t.Errorf("got %d, want 9 (nil hydro → DM-4)", got)
	}
}

func TestRollBiomass_NilTemperature_NoTempDMs(t *testing.T) {
	// Atm 6 + Hydro 5 (no DM) + Age 5 (+1), no temp DMs (nil temp).
	// 2D=8 → 8 + 1 = 9.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	r := roller.NewScripted(8)
	got := RollBiomass(r, body, 5.0)
	if got != 9 {
		t.Errorf("got %d, want 9 (nil temp, no temp DMs)", got)
	}
}

func TestRollBiocomplexity_ZedPrime(t *testing.T) {
	// Biomass=10 (clamped to 9), Atm 6 (in 4-9 → no DM), Age 6.3 (no age DM since > 4).
	// 2D=3 → 3 - 7 + 9 = 5.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(3)
	got := RollBiocomplexity(r, body, 10, 6.3)
	if got != 5 {
		t.Errorf("Zed Prime: got %d, want 5", got)
	}
}

func TestRollBiocomplexity_BiomassZero_Zero(t *testing.T) {
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted() // empty — must NOT consume dice
	got := RollBiocomplexity(r, body, 0, 6.3)
	if got != 0 {
		t.Errorf("got %d, want 0 (Biomass=0 prerequisite fails)", got)
	}
}

func TestRollBiocomplexity_BiomassClamp_Above9(t *testing.T) {
	// Biomass=15 should be clamped to 9 in the formula.
	// 2D=2, Atm 6, Age > 4 → 2 - 7 + 9 + 0 = 4.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(2)
	got := RollBiocomplexity(r, body, 15, 5.0)
	if got != 4 {
		t.Errorf("got %d, want 4 (Biomass=15 → uses 9)", got)
	}
}

func TestRollBiocomplexity_AgeBoundary_Exactly4_UsesWorseDM(t *testing.T) {
	// Age = 4.0 exactly → 3-4 band → DM-2 (the worst at the boundary).
	// Biomass=9, Atm 6, 2D=10 → 10 - 7 + 9 - 2 = 10.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(10)
	got := RollBiocomplexity(r, body, 9, 4.0)
	if got != 10 {
		t.Errorf("age=4 boundary: got %d, want 10 (DM-2 worst)", got)
	}
}

func TestRollBiocomplexity_AgeBoundary_Exactly1_UsesWorseDM(t *testing.T) {
	// Age = 1.0 exactly → < 1 band → DM-10 (the worst at the boundary).
	// Biomass=9, Atm 6, 2D=12 → 12 - 7 + 9 - 10 = 4.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(12)
	got := RollBiocomplexity(r, body, 9, 1.0)
	if got != 4 {
		t.Errorf("age=1 boundary: got %d, want 4 (DM-10 worst)", got)
	}
}

func TestRollBiocomplexity_AtmNotIn4to9_DMMinus2(t *testing.T) {
	// Atm 11 (B) → not in 4-9 → DM-2. Biomass=9, Age 5.
	// 2D=10 → 10 - 7 + 9 - 2 = 10.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 11}
	r := roller.NewScripted(10)
	got := RollBiocomplexity(r, body, 9, 5.0)
	if got != 10 {
		t.Errorf("got %d, want 10 (atm not 4-9 → DM-2)", got)
	}
}

func TestRollBiocomplexity_ResultLessThanOne_PromotedToOne(t *testing.T) {
	// Force a result < 1 with biomass > 0: 2D=2, Biomass=1, Atm 11 (DM-2), Age 0.5 (DM-10 from biocomplexity table).
	// 2 - 7 + 1 - 2 - 10 = -16 → < 1 → 1.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 11}
	r := roller.NewScripted(2)
	got := RollBiocomplexity(r, body, 1, 0.5)
	if got != 1 {
		t.Errorf("got %d, want 1 (result < 1 → promoted)", got)
	}
}

func TestRollBiocomplexity_NilAtmosphere_NoAtmDM(t *testing.T) {
	// Defensive: nil atmosphere is not in 4-9, so still gets DM-2.
	body := &DetailedPlacement{}
	body.Atmosphere = nil
	r := roller.NewScripted(10)
	got := RollBiocomplexity(r, body, 9, 5.0)
	// 10 - 7 + 9 - 2 = 10
	if got != 10 {
		t.Errorf("got %d, want 10 (nil atm → DM-2)", got)
	}
}

func TestRollNativeSophont_BelowPrerequisite_False(t *testing.T) {
	// Biocomplexity=7 < 8 → no roll, no dice consumed.
	r := roller.NewScripted() // empty
	if got := RollNativeSophont(r, 7); got {
		t.Error("got true, want false (Biocomplexity<8)")
	}
}

func TestRollNativeSophont_Triggers_AtBiocomplexity9(t *testing.T) {
	// Biocomplexity=9, 2D=11 → 11+9-7=13 ≥ 13 → true.
	r := roller.NewScripted(11)
	if got := RollNativeSophont(r, 9); !got {
		t.Error("got false, want true (mod=13)")
	}
}

func TestRollNativeSophont_BelowThreshold(t *testing.T) {
	// Biocomplexity=8, 2D=11 → 11+8-7=12 < 13 → false.
	r := roller.NewScripted(11)
	if got := RollNativeSophont(r, 8); got {
		t.Error("got true, want false (mod=12)")
	}
}

func TestRollNativeSophont_BiocomplexityClamp_Above9(t *testing.T) {
	// Biocomplexity=15 should be clamped to 9 in the formula.
	// 2D=11 → 11+9-7=13 ≥ 13 → true.
	r := roller.NewScripted(11)
	if got := RollNativeSophont(r, 15); !got {
		t.Error("got false, want true (Biocomplexity=15 → uses 9)")
	}
}

func TestRollExtinctSophont_BelowPrerequisite_False(t *testing.T) {
	// Biocomplexity=7 → no roll, no dice consumed.
	r := roller.NewScripted()
	if got := RollExtinctSophont(r, 7, 6.0); got {
		t.Error("got true, want false (Biocomplexity<8)")
	}
}

func TestRollExtinctSophont_AgeOver5_DMPlusOne(t *testing.T) {
	// Biocomplexity=9, Age=6 (DM+1), 2D=10 → 10+9-7+1=13 ≥ 13 → true.
	// Without the +1: 10+9-7=12 < 13 → false. So this test verifies the DM applies.
	r := roller.NewScripted(10)
	if got := RollExtinctSophont(r, 9, 6.0); !got {
		t.Error("got false, want true (age>5 DM+1 makes mod=13)")
	}
}

func TestRollExtinctSophont_AgeUnder5_NoDM(t *testing.T) {
	// Biocomplexity=9, Age=4 (no DM), 2D=10 → 10+9-7=12 < 13 → false.
	r := roller.NewScripted(10)
	if got := RollExtinctSophont(r, 9, 4.0); got {
		t.Error("got true, want false (age≤5 no DM, mod=12)")
	}
}

func TestRollExtinctSophont_HighRoll_AlwaysTrue(t *testing.T) {
	// Biocomplexity=8, Age=4, 2D=12 → 12+8-7=13 ≥ 13 → true.
	r := roller.NewScripted(12)
	if got := RollExtinctSophont(r, 8, 4.0); !got {
		t.Error("got false, want true (mod=13 even without age DM)")
	}
}

func TestRollExtinctSophont_BiocomplexityClamp_Above9(t *testing.T) {
	// Biocomplexity=15 → clamped to 9. 2D=10, Age=4 → 10+9-7=12 < 13 → false.
	r := roller.NewScripted(10)
	if got := RollExtinctSophont(r, 15, 4.0); got {
		t.Error("got true, want false (Biocomplexity=15 → 9, mod=12)")
	}
}

func TestRollBiodiversity_ZedPrime(t *testing.T) {
	// Biomass=10, Biocomplexity=5 → (10+5)/2 = 7.5. 2D=6 → 6-7+7.5 = 6.5 → ceil → 7.
	r := roller.NewScripted(6)
	got := RollBiodiversity(r, 10, 5)
	if got != 7 {
		t.Errorf("Zed Prime: got %d, want 7", got)
	}
}

func TestRollBiodiversity_BiomassZero_Zero(t *testing.T) {
	r := roller.NewScripted()
	if got := RollBiodiversity(r, 0, 5); got != 0 {
		t.Errorf("got %d, want 0 (biomass=0 prerequisite fails)", got)
	}
}

func TestRollBiodiversity_RoundsUp(t *testing.T) {
	// Biomass=4, Biocomplexity=3 → (4+3)/2 = 3.5. 2D=8 → 8-7+3.5 = 4.5 → ceil → 5.
	r := roller.NewScripted(8)
	got := RollBiodiversity(r, 4, 3)
	if got != 5 {
		t.Errorf("got %d, want 5 (ceil semantics)", got)
	}
}

func TestRollBiodiversity_ResultLessThanOne_PromotedToOne(t *testing.T) {
	// Biomass=1, Biocomplexity=1 → (1+1)/2 = 1. 2D=2 → 2-7+1 = -4 → < 1 → 1.
	r := roller.NewScripted(2)
	got := RollBiodiversity(r, 1, 1)
	if got != 1 {
		t.Errorf("got %d, want 1 (result<1 promoted)", got)
	}
}

func TestRollBiodiversity_IntegerArithmetic_NoFractional(t *testing.T) {
	// Biomass=4, Biocomplexity=4 → (4+4)/2 = 4 (integer). 2D=7 → 7-7+4 = 4 (no rounding).
	r := roller.NewScripted(7)
	got := RollBiodiversity(r, 4, 4)
	if got != 4 {
		t.Errorf("got %d, want 4 (no rounding when integer)", got)
	}
}

func TestRollCompatibility_ZedPrime_FollowsFormula(t *testing.T) {
	// Per WBH p.131 formula box: 2D - Biocomplexity/2 + DMs.
	// Atm 6 (DM+2), Biocomplexity=5, Age 6.3, 2D=7 → 7 - 2.5 + 2 = 6.5 → floor → 6.
	//
	// NOTE: WBH p.131 worked example shows 7 + 3 - 2.5 + 2 = 9.5 → 9. The
	// "+3" has no source in the formula box. Implementation follows formula
	// → Zed Prime gets 6 (book says 9). Logged as feedback memory after merge.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(7)
	got := RollCompatibility(r, body, 5, 6.3)
	if got != 6 {
		t.Errorf("Zed Prime per formula: got %d, want 6 (book worked example says 9)", got)
	}
}

func TestRollCompatibility_BiomassDependsOnPrereq_NoDirectGate(t *testing.T) {
	// The Compatibility function itself doesn't gate on biomass — caller
	// should check biomass > 0 before calling. This test verifies the
	// function still returns the formula result for any inputs.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(7)
	got := RollCompatibility(r, body, 5, 5.0)
	if got != 6 {
		t.Errorf("got %d, want 6", got)
	}
}

func TestRollCompatibility_NegativeResult_ClampedToZero(t *testing.T) {
	// Atm C (DM-10), Biocomplexity=10, 2D=2 → 2 - 5 - 10 = -13 → ≤ 0 → 0.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 12}
	r := roller.NewScripted(2)
	got := RollCompatibility(r, body, 10, 5.0)
	if got != 0 {
		t.Errorf("got %d, want 0 (negative result clamped)", got)
	}
}

func TestRollCompatibility_AtmCRich_DMMinus10(t *testing.T) {
	// Atm C (DM-10), Biocomplexity=4, 2D=12 → 12 - 2 - 10 = 0 → 0.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 12}
	r := roller.NewScripted(12)
	got := RollCompatibility(r, body, 4, 5.0)
	if got != 0 {
		t.Errorf("got %d, want 0 (atm C heavy penalty)", got)
	}
}

func TestRollCompatibility_AgeOver8_DMMinus2(t *testing.T) {
	// Atm 6 (+2), Biocomplexity=4, Age=9 (DM-2), 2D=10 → 10 - 2 + 2 - 2 = 8.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(10)
	got := RollCompatibility(r, body, 4, 9.0)
	if got != 8 {
		t.Errorf("got %d, want 8 (age>8 DM-2)", got)
	}
}

func TestRollCompatibility_FloorRounding(t *testing.T) {
	// Biocomplexity=3 → 3/2 = 1.5. 2D=10, Atm 6 (+2) → 10 - 1.5 + 2 = 10.5 → floor → 10.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(10)
	got := RollCompatibility(r, body, 3, 5.0)
	if got != 10 {
		t.Errorf("got %d, want 10 (floor 10.5)", got)
	}
}

func TestRollCompatibility_NilAtmosphere_NoAtmDM(t *testing.T) {
	// Defensive: nil atm → no atm DM applied. Biocomplexity=4, Age 5, 2D=10 → 10 - 2 = 8.
	body := &DetailedPlacement{}
	r := roller.NewScripted(10)
	got := RollCompatibility(r, body, 4, 5.0)
	if got != 8 {
		t.Errorf("got %d, want 8 (nil atm)", got)
	}
}
