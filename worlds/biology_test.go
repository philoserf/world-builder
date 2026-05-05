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
