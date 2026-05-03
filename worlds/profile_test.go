package worlds

import (
	"strings"
	"testing"
)

// TestShortProfile_Zed asserts the WBH p.58 short form for the Zed
// system: "4-2-12-5-0.5".
func TestShortProfile_Zed(t *testing.T) {
	t.Parallel()

	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			Counts:       Counts{GasGiants: 4, PlanetoidBelts: 2, Terrestrials: 12, Total: 18},
			BaselineN:    5,
			SystemSpread: 0.5,
		},
	}
	got := ShortProfile(sd)
	want := "4-2-12-5-0.5"
	if got != want {
		t.Errorf("ShortProfile = %q, want %q", got, want)
	}
}

// TestShortProfile_BaselineNFloor asserts BaselineN < 0 renders as 0.
func TestShortProfile_BaselineNFloor(t *testing.T) {
	t.Parallel()
	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			Counts:       Counts{GasGiants: 1, PlanetoidBelts: 0, Terrestrials: 3, Total: 4},
			BaselineN:    -2,
			SystemSpread: 0.7,
		},
	}
	got := ShortProfile(sd)
	want := "1-0-3-0-0.7"
	if got != want {
		t.Errorf("ShortProfile = %q, want %q", got, want)
	}
}

// TestLongProfile_Zed asserts the WBH p.58 long form for Zed:
// "Aab-5-T-T-T-P-G-G-T-T-T-0.5:B-2-T-T-0.5:AB-0-T-T-G-0.5:Cab-0-P-G-T-T-0.5"
func TestLongProfile_Zed(t *testing.T) {
	t.Parallel()

	gAab := Group{Designation: "Aab"}
	gB := Group{Designation: "B"}
	gAB := Group{Designation: "AB"}
	gCab := Group{Designation: "Cab"}

	dps := []DetailedPlacement{
		// Aab: T-T-T-P-G-G-T-T-T (9 slots in orbit order)
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.6, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.1, Group: gAab}}}},
		{Placement: Placement{Body: BodyPlanetoidBelt, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.7, Group: gAab}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.1, Group: gAab}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.5, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.1, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.6, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 5.2, Group: gAab}}}},
		// B: T-T (2 slots)
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 0.52, Group: gB}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: gB}}}},
		// AB: T-T-G (3 slots)
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 7.2, Group: gAB}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 7.8, Group: gAB}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 8.3, Group: gAB}}}},
		// Cab: P-G-T-T (4 slots)
		{Placement: Placement{Body: BodyPlanetoidBelt, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.4, Group: gCab}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.3, Group: gCab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.9, Group: gCab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.3, Group: gCab}}}},
	}

	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			SystemSpread: 0.5,
			Allocations: []StarAllocation{
				{Group: gAab, BaselineN: 5},
				{Group: gB, BaselineN: 2},
				{Group: gAB, BaselineN: 0},
				{Group: gCab, BaselineN: 0},
			},
		},
		Detailed: dps,
	}

	got := LongProfile(sd)
	want := "Aab-5-T-T-T-P-G-G-T-T-T-0.5:B-2-T-T-0.5:AB-0-T-T-G-0.5:Cab-0-P-G-T-T-0.5"
	if got != want {
		t.Errorf("LongProfile mismatch\n got: %q\nwant: %q", got, want)
	}

	if c := strings.Count(got, ":"); c != 3 {
		t.Errorf("LongProfile colon count = %d, want 3 (4 stars separated by :)", c)
	}
}
