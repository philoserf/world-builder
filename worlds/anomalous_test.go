package worlds

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

func TestAddAnomalous_None(t *testing.T) {
	t.Parallel()

	primary := Group{Designation: "A", Members: []stars.Star{{}}, MAO: 0.5, Intervals: []Interval{{Min: 0.5, Max: 20.0}}}
	allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 5}}
	counts := Counts{Total: 5}
	slots := []Slot{{StarSlot: "A1", Group: primary, Orbit: 1.0}}

	out, newCounts, err := AddAnomalous(roller.NewScripted(5), slots, allocs, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(out) != 1 {
		t.Errorf("slots = %d, want 1", len(out))
	}

	if newCounts.Total != counts.Total {
		t.Errorf("counts unchanged should hold, got Total %d", newCounts.Total)
	}

	if out[0].Anomaly != AnomalyNone {
		t.Errorf("Anomaly = %v, want None", out[0].Anomaly)
	}
}

func TestAddAnomalous_Retrograde_SingleStar(t *testing.T) {
	t.Parallel()

	primary := Group{Designation: "A", Members: []stars.Star{{}}, MAO: 0.5, Intervals: []Interval{{Min: 0.5, Max: 20.0}}}
	allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 5}}
	counts := Counts{Terrestrials: 4, Total: 5}
	slots := []Slot{{StarSlot: "A1", Group: primary, Orbit: 1.0}}
	// Anomalous count: 10 → 1.
	// Type: 10 → Retrograde.
	// Random orbit: 2D-2 = 5 (raw 7); d10 = 2 → orbit 5.2.
	rolls := []int{10, 10, 7, 2}

	out, newCounts, err := AddAnomalous(roller.NewScripted(rolls...), slots, allocs, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(out) != 2 {
		t.Fatalf("slots = %d, want 2", len(out))
	}

	last := out[1]
	if last.Anomaly != AnomalyRetrograde {
		t.Errorf("Anomaly = %v, want Retrograde", last.Anomaly)
	}

	if math.Abs(last.Orbit-5.2) > 0.01 {
		t.Errorf("Orbit = %v, want 5.2", last.Orbit)
	}

	if last.EccentricityDM != 2 {
		t.Errorf("EccentricityDM = %d, want +2", last.EccentricityDM)
	}

	if newCounts.Terrestrials != counts.Terrestrials+1 {
		t.Errorf("Terrestrials = %d, want %d", newCounts.Terrestrials, counts.Terrestrials+1)
	}

	if newCounts.Total != counts.Total+1 {
		t.Errorf("Total = %d, want %d", newCounts.Total, counts.Total+1)
	}
}

func TestAddAnomalous_TypeTable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		roll int
		want AnomalyType
	}{
		{2, AnomalyRandom},
		{7, AnomalyRandom},
		{8, AnomalyEccentric},
		{9, AnomalyInclined},
		{10, AnomalyRetrograde},
		{11, AnomalyRetrograde},
		{12, AnomalyTrojan},
	}
	primary := Group{Designation: "A", Members: []stars.Star{{}}, MAO: 0.5, Intervals: []Interval{{Min: 0.5, Max: 20.0}}}
	allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 5}}
	counts := Counts{Total: 5}
	slots := []Slot{{StarSlot: "A1", Group: primary, Orbit: 3.0}}

	for _, tc := range cases {
		// Anomalous count = 10 (1). Type = tc.roll. Random orbit: 2D-2 = 5 (raw 7), d10 = 2.
		rolls := []int{10, tc.roll, 7, 2}
		if tc.want == AnomalyInclined {
			rolls = append(rolls, 4, 5) // 1D=4, d10=5
		}

		out, _, err := AddAnomalous(roller.NewScripted(rolls...), slots, allocs, counts)
		if err != nil {
			t.Fatalf("type %d: %v", tc.roll, err)
		}

		if out[len(out)-1].Anomaly != tc.want {
			t.Errorf("type %d: Anomaly = %v, want %v", tc.roll, out[len(out)-1].Anomaly, tc.want)
		}
	}
}

func TestAddAnomalous_RandomClampsToMAO(t *testing.T) {
	t.Parallel()

	primary := Group{Designation: "A", Members: []stars.Star{{}}, MAO: 1.0, Intervals: []Interval{{Min: 1.0, Max: 20.0}}}
	allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 5}}
	counts := Counts{Total: 5}
	// Anomalous = 1, type = Random (7), orbit 0.1 (below MAO 1.0).
	// Retry: 2D-2 = 5 (raw 7), d10 = 5 → 5.5.
	rolls := []int{10, 7, 2, 1, 7, 5}

	out, _, err := AddAnomalous(roller.NewScripted(rolls...), nil, allocs, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if math.Abs(out[len(out)-1].Orbit-5.5) > 0.01 {
		t.Errorf("Orbit = %v, want 5.5 (after retry)", out[len(out)-1].Orbit)
	}
}
