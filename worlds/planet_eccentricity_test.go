package worlds

import (
	"math"
	"testing"

	"wbh/roller"
	"wbh/stars"
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
	out, err := RollPlanetEccentricities(roller.NewScripted(7, 1, 7, 1), placements)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(out) != len(placements) {
		t.Fatalf("len = %d, want %d", len(out), len(placements))
	}
	// First placement (no DM) and second (DM=5) should have DIFFERENT eccentricities.
	if out[0].Eccentricity == out[1].Eccentricity {
		t.Errorf("placements 0 and 1 have same eccentricity %v; ExtraDM had no effect", out[0].Eccentricity)
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
	out, err := RollPlanetEccentricities(roller.NewScripted(7, 1), placements)
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
	out, err := RollPlanetEccentricities(roller.NewScripted(7, 1, 7, 1), placements)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if out[0].Eccentricity == out[1].Eccentricity {
		t.Errorf("primary (depth 0) and secondary (depth 1) eccentricities both = %v; NestingDepth had no effect", out[0].Eccentricity)
	}
}
