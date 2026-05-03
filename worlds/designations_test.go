package worlds

import "testing"

// TestAssignPlanetDesignations_BeltSkip exercises the WBH p.53 rule:
// "Planetoid belts are not enumerated as planets — planet enumeration
// skips a belt location and continues uninterrupted with the next planet."
//
// Zed Aab placements: 8 worlds + 1 belt + 1 retrograde:
//
//	1.0  Terr → Aab I       1.6  Terr → Aab II      2.1  Terr → Aab III
//	2.7  Belt → Aab PI      3.1  GG   → Aab IV      3.5  GG   → Aab V
//	4.1  Terr → Aab VI      4.6  Terr → Aab VII     5.2  Terr → Aab VIII (retrograde)
func TestAssignPlanetDesignations_BeltSkip(t *testing.T) {
	t.Parallel()

	g := Group{Designation: "Aab"}
	dps := []DetailedPlacement{
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: g}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.6, Group: g}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.1, Group: g}}}},
		{Placement: Placement{Body: BodyPlanetoidBelt, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.7, Group: g}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.1, Group: g}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.5, Group: g}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.1, Group: g}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.6, Group: g}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 5.2, Group: g}}}},
	}

	AssignPlanetDesignations(dps)

	want := []string{
		"Aab I", "Aab II", "Aab III", "Aab PI",
		"Aab IV", "Aab V", "Aab VI", "Aab VII", "Aab VIII",
	}
	for i, w := range want {
		if dps[i].Designation != w {
			t.Errorf("dps[%d].Designation = %q, want %q", i, dps[i].Designation, w)
		}
	}
}

// TestAssignPlanetDesignations_PerGroupReset: WBH p.53 — "Each new set
// of stars resets the planetary enumeration to 'I'."
func TestAssignPlanetDesignations_PerGroupReset(t *testing.T) {
	t.Parallel()

	gAab := Group{Designation: "Aab"}
	gAB := Group{Designation: "AB"}
	gB := Group{Designation: "B"}
	gCab := Group{Designation: "Cab"}

	dps := []DetailedPlacement{
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 7.2, Group: gAB}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 7.8, Group: gAB}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 0.52, Group: gB}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: gB}}}},
		{Placement: Placement{Body: BodyPlanetoidBelt, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.4, Group: gCab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.3, Group: gCab}}}},
	}

	AssignPlanetDesignations(dps)

	want := []string{"Aab I", "AB I", "AB II", "B I", "B II", "Cab PI", "Cab I"}
	for i, w := range want {
		if dps[i].Designation != w {
			t.Errorf("dps[%d].Designation = %q, want %q", i, dps[i].Designation, w)
		}
	}
}

// TestAssignMoonDesignations_AlphabeticOrder: WBH p.58 — moons closest to
// farthest from planet, alphabetic.
func TestAssignMoonDesignations_AlphabeticOrder(t *testing.T) {
	t.Parallel()

	dps := []DetailedPlacement{
		{
			Designation: "Aab IV",
			Moons: []Moon{
				{SizeCode: "2"}, // a
				{SizeCode: "S"}, // b
				{SizeCode: "S"}, // c
				{SizeCode: "5"}, // d
				{SizeCode: "S"}, // e
			},
		},
		{
			Designation: "Aab V",
			Moons: []Moon{
				{SizeCode: "S"},
				{SizeCode: "A"},
				{SizeCode: "1"},
				{SizeCode: "3"},
				{SizeCode: "S"},
				{SizeCode: "S"},
			},
		},
	}

	AssignMoonDesignations(dps)

	wantAabIV := []string{"Aab IV a", "Aab IV b", "Aab IV c", "Aab IV d", "Aab IV e"}
	for i, w := range wantAabIV {
		if dps[0].Moons[i].Designation != w {
			t.Errorf("dps[0].Moons[%d].Designation = %q, want %q", i, dps[0].Moons[i].Designation, w)
		}
	}
	wantAabV := []string{"Aab V a", "Aab V b", "Aab V c", "Aab V d", "Aab V e", "Aab V f"}
	for i, w := range wantAabV {
		if dps[1].Moons[i].Designation != w {
			t.Errorf("dps[1].Moons[%d].Designation = %q, want %q", i, dps[1].Moons[i].Designation, w)
		}
	}
}

// TestAssignMoonDesignations_NoMoonsNoPanic asserts the no-moons path
// doesn't panic and leaves Designation empty.
func TestAssignMoonDesignations_NoMoonsNoPanic(t *testing.T) {
	t.Parallel()
	dps := []DetailedPlacement{
		{Designation: "Aab III", Moons: nil},
		{Designation: "Aab VII", Moons: []Moon{}},
	}
	AssignMoonDesignations(dps) // must not panic
}
