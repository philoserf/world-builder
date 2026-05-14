package worlds

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

func TestRollPlanetEccentricities_AppliesAnomalyDM(t *testing.T) {
	t.Parallel()
	placements := []Placement{
		{AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "A1", Orbit: 1.0}}, Body: BodyTerrestrial},
		{AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "A+", Orbit: 5.0}, Anomaly: AnomalyEccentric, EccentricityDM: 5}, Body: BodyTerrestrial},
		{AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "A2", Orbit: 2.0}}, Body: BodyEmpty},
		{AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "A3", Orbit: 3.0}}, Body: BodyPlanetoidBelt},
	}
	// Two real RollEccentricity calls (skips Empty + Belt). Each consumes 2 rolls
	// (2D for the row + a second roll for the value).
	// Placement 0: natural=7, no DM → row 7 → Base 0.00 + 1/200 = 0.005.
	// Placement 1: natural=7, ExtraDM=5 → row 12 → Base 0.30 + 1/20 = 0.35.
	out, err := RollPlanetEccentricities(roller.NewScripted(7, 1, 7, 1), placements, 0.0)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(out) != len(placements) {
		t.Fatalf("len = %d, want %d", len(out), len(placements))
	}
	// Placement 0: row 7 → 0.00 + 1/200 = 0.005.
	if want := 0.005; math.Abs(out[0].Eccentricity-want) > 1e-9 {
		t.Errorf("placement 0 eccentricity = %v, want %v (row 7, no DM)", out[0].Eccentricity, want)
	}
	// Placement 1: row 12 (ExtraDM=5) → 0.30 + 1/20 = 0.35.
	if want := 0.35; math.Abs(out[1].Eccentricity-want) > 1e-9 {
		t.Errorf("placement 1 eccentricity = %v, want %v (row 12, ExtraDM=5)", out[1].Eccentricity, want)
	}
	// Empty and belt placements: eccentricity remains 0 (no roll consumed for them).
	if out[2].Eccentricity != 0 || out[3].Eccentricity != 0 {
		t.Errorf("Empty/Belt eccentricities = %v, %v; want 0,0", out[2].Eccentricity, out[3].Eccentricity)
	}
}

func TestRollPlanetEccentricities_TrojanInheritsParent(t *testing.T) {
	t.Parallel()
	placements := []Placement{
		// Parent slot at A1 with some rolled eccentricity.
		{AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "A1", Orbit: 3.0}}, Body: BodyTerrestrial},
		// Trojan shadowing A1.
		{AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "A+", Orbit: 3.0}, Anomaly: AnomalyTrojan, TrojanOf: "A1"}, Body: BodyTerrestrial},
	}
	// Only one RollEccentricity call (for A1). 2 rolls (2D for row + value).
	out, err := RollPlanetEccentricities(roller.NewScripted(7, 1), placements, 0.0)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if math.Abs(out[0].Eccentricity-out[1].Eccentricity) > 1e-9 {
		t.Errorf("Trojan eccentricity %v should equal parent %v", out[1].Eccentricity, out[0].Eccentricity)
	}
}

func TestRollPlanetEccentricities_NestingDepthForSecondary(t *testing.T) {
	t.Parallel()
	primary := Group{Designation: "A", Members: []stars.Star{{}}}
	nearCompanion := stars.CompanionStar{OrbitClass: stars.OrbitNear}
	secondary := Group{Designation: "B", Members: []stars.Star{{}}}
	secondary.sourceCompanion = &nearCompanion

	placements := []Placement{
		{AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "A1", Group: primary, Orbit: 3.0}}, Body: BodyTerrestrial},
		{AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "B1", Group: secondary, Orbit: 1.0}}, Body: BodyTerrestrial},
	}
	// Same scripted rolls for both. NestingDepth differs (0 vs 1) → results should differ.
	out, err := RollPlanetEccentricities(roller.NewScripted(7, 1, 7, 1), placements, 0.0)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if out[0].Eccentricity == out[1].Eccentricity {
		t.Errorf("primary (depth 0) and secondary (depth 1) eccentricities both = %v; NestingDepth had no effect", out[0].Eccentricity)
	}
}

// TestRollPlanetEccentricities_AgeDMApplies asserts that
// RollPlanetEccentricities passes Orbit and SystemAgeGyr through to
// stars.RollEccentricity, so the WBH p.27 sub-1.0/age>1Gyr DM-1
// actually shifts the EccentricityValues table-lookup index.
//
// Setup: one terrestrial placement at orbit 0.5 with ageGyr 5.0.
// EccentricityValues (WBH p.27):
//
//	Row 7: Base 0.00 + 1D/200
//	Row 8: Base 0.03 + 1D/100
//
// With DM-1 (ageGyr=5.0 > 1.0): natural=8 → row 7 → 0.00 + 4/200 = 0.02.
// Without DM   (ageGyr=0.5 ≤ 1.0): natural=8 → row 8 → 0.03 + 4/100 = 0.07.
func TestRollPlanetEccentricities_AgeDMApplies(t *testing.T) {
	t.Parallel()

	p := Placement{
		AnomalousSlot: AnomalousSlot{
			Slot: Slot{Orbit: 0.5},
		},
		Body: BodyTerrestrial,
	}

	// With ageGyr > 1, DM-1 → row 7 → Base 0.00 + 4/200 = 0.02.
	scripted := roller.NewScripted(8, 4)
	out, err := RollPlanetEccentricities(scripted, []Placement{p}, 5.0)
	if err != nil {
		t.Fatalf("RollPlanetEccentricities returned error: %v", err)
	}
	if got, want := out[0].Eccentricity, 0.02; math.Abs(got-want) > 1e-9 {
		t.Errorf("Eccentricity with ageGyr=5.0 = %v, want %v (DM-1 should shift to row 7)", got, want)
	}

	// Without the age DM, the same dice produce row 8's value.
	scripted = roller.NewScripted(8, 4)
	out, err = RollPlanetEccentricities(scripted, []Placement{p}, 0.5)
	if err != nil {
		t.Fatalf("RollPlanetEccentricities (no age DM) returned error: %v", err)
	}
	if got, want := out[0].Eccentricity, 0.07; math.Abs(got-want) > 1e-9 {
		t.Errorf("Eccentricity with ageGyr=0.5 = %v, want %v (no DM)", got, want)
	}
}
