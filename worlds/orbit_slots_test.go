package worlds

import (
	"math"
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestPlaceOrbitSlots_SinglePrimary_NoExclusions(t *testing.T) {
	t.Parallel()
	primary := Group{
		Designation: "A",
		Members:     []stars.Star{{}},
		MAO:         0.5,
		Intervals:   []Interval{{Min: 0.5, Max: 20.0}},
	}
	allocs := []StarAllocation{{Group: primary, TotalStarOrbits: 5, AllocatedWorlds: 5}}
	// baselineOrbit=1.5, spread=1.0, MAO=0.5 → derived baselineN = round((1.5-0.5)/1.0)+1 = 2.
	// 5 slots: variance rolls all 7 (no variance):
	//   slot 1 (inner): MAO 0.5 + 1.0 = 1.5
	//   slot 2 (BASELINE): 1.5 (overrides variance)
	//   slot 3: 1.5 + 1.0 = 2.5
	//   slot 4: 2.5 + 1.0 = 3.5
	//   slot 5: 3.5 + 1.0 = 4.5
	// Variance roll consumed: slot 1 (1), slot 3 (1), slot 4 (1), slot 5 (1) = 4 rolls.
	got, err := PlaceOrbitSlots(roller.NewScripted(7, 7, 7, 7), allocs, 1.5, 1.0, 0)
	if err != nil {
		t.Fatalf("%v", err)
	}
	want := []float64{1.5, 1.5, 2.5, 3.5, 4.5}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		if math.Abs(got[i].Orbit-w) > 0.05 {
			t.Errorf("slot %d Orbit = %v, want %v", i, got[i].Orbit, w)
		}
	}
}

func TestPlaceOrbitSlots_BaselineFixedSlotIsAtBaselineOrbit(t *testing.T) {
	t.Parallel()
	primary := Group{
		Members:   []stars.Star{{}},
		MAO:       0.5,
		Intervals: []Interval{{Min: 0.5, Max: 20.0}},
	}
	allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 5}}
	// baselineOrbit=5.0, spread=1.0, MAO=0.5 → derived baselineN = round((5.0-0.5)/1.0)+1 = 6.
	// 5 slots → baselineN clamped to 5. So slot 5 (the last) is fixed at 5.0.
	// Slots 1-4 use variance rolls (7,7,7,7 → no variance):
	//   slot 1: 0.5 + 1.0 = 1.5
	//   slot 2: 1.5 + 1.0 = 2.5
	//   slot 3: 2.5 + 1.0 = 3.5
	//   slot 4: 3.5 + 1.0 = 4.5
	//   slot 5: BASELINE → 5.0
	got, err := PlaceOrbitSlots(roller.NewScripted(7, 7, 7, 7), allocs, 5.0, 1.0, 0)
	if err != nil {
		t.Fatalf("%v", err)
	}
	want := []float64{1.5, 2.5, 3.5, 4.5, 5.0}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		if math.Abs(got[i].Orbit-w) > 0.05 {
			t.Errorf("slot %d Orbit = %v, want %v", i, got[i].Orbit, w)
		}
	}
}

func TestPlaceOrbitSlots_ExclusionZoneWidens(t *testing.T) {
	t.Parallel()
	primary := Group{
		Members:   []stars.Star{{}},
		MAO:       0.5,
		Intervals: []Interval{{Min: 0.5, Max: 5.0}, {Min: 8.0, Max: 20.0}},
	}
	allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 4}}
	// Spread 2.0, baselineOrbit 2.5. Derived baselineN = round((2.5-0.5)/2.0)+1 = round(1.0)+1 = 2.
	//   slot 1 (inner): 0.5 + 2.0 + 0 (variance 7) = 2.5
	//   slot 2 (BASELINE): 2.5 (overrides)
	//   slot 3: 2.5 + 2.0 + 0 = 4.5 (still in lower interval, OK, no widening)
	//   slot 4: 4.5 + 2.0 + 0 = 6.5 — inside exclusion zone (5.0, 8.0)!
	//     Widen by zone width 3.0: slot 4 = 6.5 + 3.0 = 9.5
	// Variance rolls: slot 1, 3, 4 = 3 rolls.
	got, err := PlaceOrbitSlots(roller.NewScripted(7, 7, 7), allocs, 2.5, 2.0, 0)
	if err != nil {
		t.Fatalf("%v", err)
	}
	want := []float64{2.5, 2.5, 4.5, 9.5}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		if math.Abs(got[i].Orbit-w) > 0.05 {
			t.Errorf("slot %d Orbit = %v, want %v", i, got[i].Orbit, w)
		}
	}
}

func TestPlaceOrbitSlots_EmptyDistribution(t *testing.T) {
	t.Parallel()
	primary := Group{Designation: "A", Members: []stars.Star{{}}, MAO: 0.5, Intervals: []Interval{{Min: 0.5, Max: 20.0}}}
	nearCompanion := stars.CompanionStar{OrbitClass: stars.OrbitNear, OrbitNumber: 6.0, ParentIndex: -1}
	nearSec := Group{
		Designation: "B",
		Members:     []stars.Star{{}},
		MAO:         0.1,
		Intervals:   []Interval{{Min: 0.1, Max: 1.5}},
	}
	nearSec.sourceCompanion = &nearCompanion
	allocs := []StarAllocation{
		{Group: primary, AllocatedWorlds: 3},
		{Group: nearSec, AllocatedWorlds: 1},
	}
	// emptyOrbits = 1 → goes to first non-primary alloc with sourceCompanion in
	// Close/Near/Far order. Near matches → B grows from 1 to 2 slots.
	// Variance rolls: primary has 3 slots (one is baseline → 2 variance rolls);
	// B has 2 slots, baseline doesn't apply (B is not primary) → 2 variance rolls.
	got, err := PlaceOrbitSlots(roller.NewScripted(7, 7, 7, 7), allocs, 0.5, 0.5, 1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	primaryCount, secCount := 0, 0
	for _, s := range got {
		switch s.Group.Designation {
		case "A":
			primaryCount++
		case "B":
			secCount++
		}
	}
	if primaryCount != 3 {
		t.Errorf("primary slots = %d, want 3", primaryCount)
	}
	if secCount != 2 {
		t.Errorf("B slots = %d, want 2 (1 alloc + 1 empty bump)", secCount)
	}
	// The bumped B slot has the "+" suffix.
	foundPlus := false
	for _, s := range got {
		if s.Group.Designation == "B" && len(s.StarSlot) > 0 && s.StarSlot[len(s.StarSlot)-1] == '+' {
			foundPlus = true
		}
	}
	if !foundPlus {
		t.Errorf("expected a B+ slot, got %v", got)
	}
}
