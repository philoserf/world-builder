package worlds

import (
	"testing"

	"wbh/roller"
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
