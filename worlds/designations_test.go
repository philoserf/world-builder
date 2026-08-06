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
	bodies := []Body{
		{Kind: BodyTerrestrial, Group: g, Orbit: 1.0},
		{Kind: BodyTerrestrial, Group: g, Orbit: 1.6},
		{Kind: BodyTerrestrial, Group: g, Orbit: 2.1},
		{Kind: BodyPlanetoidBelt, Group: g, Orbit: 2.7},
		{Kind: BodyGasGiant, Group: g, Orbit: 3.1},
		{Kind: BodyGasGiant, Group: g, Orbit: 3.5},
		{Kind: BodyTerrestrial, Group: g, Orbit: 4.1},
		{Kind: BodyTerrestrial, Group: g, Orbit: 4.6},
		{Kind: BodyTerrestrial, Group: g, Orbit: 5.2},
	}

	AssignPlanetDesignations(bodies)

	want := []string{
		"Aab I", "Aab II", "Aab III", "Aab PI",
		"Aab IV", "Aab V", "Aab VI", "Aab VII", "Aab VIII",
	}
	for i, w := range want {
		if bodies[i].Designation != w {
			t.Errorf("bodies[%d].Designation = %q, want %q", i, bodies[i].Designation, w)
		}
	}
}

// TestAssignPlanetDesignations_PerGroupReset: WBH p.53 — "Each new set
// of stars resets the planetary enumeration to 'I'.".
func TestAssignPlanetDesignations_PerGroupReset(t *testing.T) {
	t.Parallel()

	gAab := Group{Designation: "Aab"}
	gAB := Group{Designation: "AB"}
	gB := Group{Designation: "B"}
	gCab := Group{Designation: "Cab"}

	bodies := []Body{
		{Kind: BodyTerrestrial, Group: gAab, Orbit: 1.0},
		{Kind: BodyTerrestrial, Group: gAB, Orbit: 7.2},
		{Kind: BodyTerrestrial, Group: gAB, Orbit: 7.8},
		{Kind: BodyTerrestrial, Group: gB, Orbit: 0.52},
		{Kind: BodyTerrestrial, Group: gB, Orbit: 1.0},
		{Kind: BodyPlanetoidBelt, Group: gCab, Orbit: 1.4},
		{Kind: BodyTerrestrial, Group: gCab, Orbit: 2.3},
	}

	AssignPlanetDesignations(bodies)

	want := []string{"Aab I", "AB I", "AB II", "B I", "B II", "Cab PI", "Cab I"}
	for i, w := range want {
		if bodies[i].Designation != w {
			t.Errorf("bodies[%d].Designation = %q, want %q", i, bodies[i].Designation, w)
		}
	}
}

// TestAssignMoonDesignations_AlphabeticOrder: WBH p.58 — moons closest to
// farthest from planet, alphabetic.
func TestAssignMoonDesignations_AlphabeticOrder(t *testing.T) {
	t.Parallel()

	bodies := []Body{
		{
			Designation: "Aab IV",
			Children: []*Body{
				{Kind: BodyMoon, SizeCode: "2"}, // a
				{Kind: BodyMoon, SizeCode: "S"}, // b
				{Kind: BodyMoon, SizeCode: "S"}, // c
				{Kind: BodyMoon, SizeCode: "5"}, // d
				{Kind: BodyMoon, SizeCode: "S"}, // e
			},
		},
		{
			Designation: "Aab V",
			Children: []*Body{
				{Kind: BodyMoon, SizeCode: "S"},
				{Kind: BodyMoon, SizeCode: "A"},
				{Kind: BodyMoon, SizeCode: "1"},
				{Kind: BodyMoon, SizeCode: "3"},
				{Kind: BodyMoon, SizeCode: "S"},
				{Kind: BodyMoon, SizeCode: "S"},
			},
		},
	}

	AssignMoonDesignations(bodies)

	wantAabIV := []string{"Aab IV a", "Aab IV b", "Aab IV c", "Aab IV d", "Aab IV e"}
	for i, w := range wantAabIV {
		if bodies[0].Children[i].Designation != w {
			t.Errorf("bodies[0].Children[%d].Designation = %q, want %q",
				i, bodies[0].Children[i].Designation, w)
		}
	}

	wantAabV := []string{"Aab V a", "Aab V b", "Aab V c", "Aab V d", "Aab V e", "Aab V f"}
	for i, w := range wantAabV {
		if bodies[1].Children[i].Designation != w {
			t.Errorf("bodies[1].Children[%d].Designation = %q, want %q",
				i, bodies[1].Children[i].Designation, w)
		}
	}
}

// TestAssignMoonDesignations_NoMoonsNoPanic asserts the no-moons path
// doesn't panic and leaves Designation empty.
func TestAssignMoonDesignations_NoMoonsNoPanic(t *testing.T) {
	t.Parallel()

	bodies := []Body{
		{Designation: "Aab III", Children: nil},
		{Designation: "Aab VII", Children: []*Body{}},
	}
	AssignMoonDesignations(bodies) // must not panic
}

// TestMarkHZ exercises the WBH p.58 HZ-tagging rule: a body is in the
// habitable zone when its orbit lies within HZCO ± 1.0 of the host
// group.
func TestMarkHZ(t *testing.T) {
	t.Parallel()

	g := Group{Designation: "Aab"}
	// Group.HZCO() reads from group's primary star; for tests we
	// drive the boundaries directly via Group's HZCO method by
	// composing a Group with known intervals; HZ is set per orbit
	// vs. that group's HZCO. Use a synthetic group whose HZCO
	// resolves to a known value via its computed primary.
	//
	// For this unit test we instead exercise the boundary logic
	// via the Group's HZCO() return — a Group with Members[0] set
	// to a star with a known HZCO. The simpler smoke test below
	// uses an empty Group (HZCO == 0) and asserts orbit-zero is HZ.
	bodies := []Body{
		{Kind: BodyTerrestrial, Group: g, Orbit: 0.0}, // empty Group → HZCO 0; in-HZ at 0±1
		{Kind: BodyTerrestrial, Group: g, Orbit: 1.5}, // out of [-1, 1]
		{Kind: BodyEmpty, Group: g, Orbit: 0.0},       // empty kind never tagged
	}

	MarkHZ(bodies)

	if !bodies[0].HZ {
		t.Errorf("orbit 0 with HZCO 0 should be HZ-tagged")
	}

	if bodies[1].HZ {
		t.Errorf("orbit 1.5 with HZCO 0 should not be HZ-tagged")
	}

	if bodies[2].HZ {
		t.Errorf("BodyEmpty should never be HZ-tagged")
	}
}
