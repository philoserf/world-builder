package worlds

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/stars"
)

func TestAllocateOrbitsByStar_SingleStar(t *testing.T) {
	t.Parallel()

	g := Group{
		Designation: "A",
		Members:     []stars.Star{{LuminosityClass: stars.V}},
		MAO:         0.03,
		Intervals:   []Interval{{Min: 0.03, Max: 20.0}},
	}
	avail := Result{Groups: []Group{g}}
	counts := Counts{GasGiants: 4, PlanetoidBelts: 1, Terrestrials: 4, Total: 9}

	got, err := AllocateOrbitsByStar(avail, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(got) != 1 {
		t.Fatalf("allocations = %d, want 1", len(got))
	}
	// Single-star: TotalStarOrbits = floor(19.97 + 1 (no companion)) = 20.
	if got[0].TotalStarOrbits != 20 {
		t.Errorf("TotalStarOrbits = %d, want 20", got[0].TotalStarOrbits)
	}

	if got[0].AllocatedWorlds != 9 {
		t.Errorf("AllocatedWorlds = %d, want 9", got[0].AllocatedWorlds)
	}
}

func TestAllocateOrbitsByStar_NoPriorAllowable_NoPlusOne(t *testing.T) {
	t.Parallel()
	// A no-companion star with zero prior allowable orbits gets no +1
	// (the book footnote: "Only if the star had more than zero previously
	// computed Allowable Orbits").
	g := Group{
		Designation: "A",
		Members:     []stars.Star{{LuminosityClass: stars.V}},
		Intervals:   nil, // 0 prior allowable
	}
	avail := Result{Groups: []Group{g}}
	counts := Counts{Total: 5}

	got, err := AllocateOrbitsByStar(avail, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if got[0].TotalStarOrbits != 0 {
		t.Errorf("TotalStarOrbits = %d, want 0", got[0].TotalStarOrbits)
	}

	if got[0].AllocatedWorlds != 5 {
		t.Errorf("AllocatedWorlds = %d, want 5 (last star gets remainder)", got[0].AllocatedWorlds)
	}
}

func TestAllocateOrbitsByStar_PairGetsNoPlusOne(t *testing.T) {
	t.Parallel()
	// A pair group (Members has 2 stars) does NOT get the +1 even if it
	// has prior allowable orbits, because it has a companion.
	g := Group{
		Designation: "Aab",
		Members:     []stars.Star{{}, {}},
		Intervals:   []Interval{{Min: 0.61, Max: 5.10}, {Min: 7.10, Max: 10.10}, {Min: 14.10, Max: 20.0}},
	}
	avail := Result{Groups: []Group{g}}
	// Total should floor (4.49 + 3.0 + 5.9 = 13.39) → 13.
	counts := Counts{Total: 13}

	got, err := AllocateOrbitsByStar(avail, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if got[0].TotalStarOrbits != 13 {
		t.Errorf("TotalStarOrbits = %d, want 13 (pair gets no +1)", got[0].TotalStarOrbits)
	}

	if got[0].AllocatedWorlds != 13 {
		t.Errorf("AllocatedWorlds = %d, want 13", got[0].AllocatedWorlds)
	}

	if math.Abs(g.Total()-13.39) > 0.01 {
		t.Errorf("(sanity) g.Total() = %v, want 13.39", g.Total())
	}
}
