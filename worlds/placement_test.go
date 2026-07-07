package worlds

import (
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

func TestPlaceWorlds_Order(t *testing.T) {
	t.Parallel()
	primary := Group{Designation: "A", Members: []stars.Star{{}}}
	slots := []AnomalousSlot{
		{Slot: Slot{StarSlot: "A1", Group: primary, Orbit: 1.0}},
		{Slot: Slot{StarSlot: "A2", Group: primary, Orbit: 2.0}},
		{Slot: Slot{StarSlot: "A3", Group: primary, Orbit: 3.0}},
		{Slot: Slot{StarSlot: "A4", Group: primary, Orbit: 4.0}},
	}
	counts := Counts{GasGiants: 1, PlanetoidBelts: 1, Terrestrials: 2, Total: 4}
	// 4 slots → 1D prefix (≤6). Each placement: (prefix, right).
	// GG: 1,1 → A1; Belt: 1,2 → A2; Terr: 1,3 → A3; Terr: 1,4 → A4.
	rolls := []int{1, 1, 1, 2, 1, 3, 1, 4}
	got, err := PlaceWorlds(roller.NewScripted(rolls...), slots, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(got) != 4 {
		t.Fatalf("len = %d, want 4", len(got))
	}
	if got[0].Body != BodyGasGiant {
		t.Errorf("A1 = %v, want GasGiant", got[0].Body)
	}
	if got[1].Body != BodyPlanetoidBelt {
		t.Errorf("A2 = %v, want PlanetoidBelt", got[1].Body)
	}
	if got[2].Body != BodyTerrestrial || got[3].Body != BodyTerrestrial {
		t.Errorf("A3,A4 should be Terrestrials")
	}
}

func TestPlaceWorlds_Collision_PlusOne(t *testing.T) {
	t.Parallel()
	primary := Group{Designation: "A", Members: []stars.Star{{}}}
	slots := []AnomalousSlot{
		{Slot: Slot{StarSlot: "A1", Group: primary, Orbit: 1.0}},
		{Slot: Slot{StarSlot: "A2", Group: primary, Orbit: 2.0}},
	}
	counts := Counts{GasGiants: 1, Terrestrials: 1, Total: 2}
	// GG → 1:1 → A1. Terrestrial → 1:1 (collision) → +1 → 1:2 → A2.
	rolls := []int{1, 1, 1, 1}
	got, err := PlaceWorlds(roller.NewScripted(rolls...), slots, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got[0].Body != BodyGasGiant {
		t.Errorf("A1 = %v, want GasGiant", got[0].Body)
	}
	if got[1].Body != BodyTerrestrial {
		t.Errorf("A2 = %v, want Terrestrial (after +1 collision)", got[1].Body)
	}
}

func TestPlaceWorlds_PrefixDieSelection(t *testing.T) {
	t.Parallel()
	primary := Group{Designation: "A", Members: []stars.Star{{}}}
	var slots []AnomalousSlot
	for i := 1; i <= 8; i++ {
		slots = append(slots, AnomalousSlot{Slot: Slot{Group: primary, Orbit: float64(i)}})
	}
	counts := Counts{Terrestrials: 8, Total: 8}
	// 8 slots → D2 prefix (7-12). Cover all 8 slots: prefix 1 → 1:1..1:6, prefix 2 → 2:1..2:2.
	rolls := []int{1, 1, 1, 2, 1, 3, 1, 4, 1, 5, 1, 6, 2, 1, 2, 2}
	got, err := PlaceWorlds(roller.NewScripted(rolls...), slots, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(got) != 8 {
		t.Errorf("placements = %d, want 8", len(got))
	}
}

func TestPlaceWorlds_IncludesEmptyBody(t *testing.T) {
	t.Parallel()
	primary := Group{Designation: "A", Members: []stars.Star{{}}}
	slots := []AnomalousSlot{
		{Slot: Slot{StarSlot: "A1", Group: primary, Orbit: 1.0}},
		{Slot: Slot{StarSlot: "A2", Group: primary, Orbit: 2.0}},
		{Slot: Slot{StarSlot: "A3", Group: primary, Orbit: 3.0}},
	}
	counts := Counts{Terrestrials: 2, Total: 2} // counts.Total < len(slots) → one empty
	// 1 empty (placed first) → 1:1 → A1. 2 terrestrials → 1:2, 1:3 → A2, A3.
	rolls := []int{1, 1, 1, 2, 1, 3}
	got, err := PlaceWorlds(roller.NewScripted(rolls...), slots, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got[0].Body != BodyEmpty {
		t.Errorf("A1 = %v, want Empty", got[0].Body)
	}
}

// TestPlaceWorlds_EmptySlotsErrors asserts the input-contract guard:
// an empty slots slice with a non-zero world count must return an
// error, not hang in rollSlot's rejection-sampling loop forever.
func TestPlaceWorlds_EmptySlotsErrors(t *testing.T) {
	t.Parallel()
	_, err := PlaceWorlds(roller.NewSeeded(1), nil, Counts{Terrestrials: 3, Total: 3})
	if err == nil {
		t.Fatal("PlaceWorlds(nil slots, Total=3) = nil error, want slot-capacity error")
	}
}
