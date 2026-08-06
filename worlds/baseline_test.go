package worlds

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

func TestRollBaselineNumber_NoMods(t *testing.T) {
	t.Parallel()

	sys := stars.System{Primary: stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, LuminosityClass: stars.V,
	})}
	counts := Counts{Total: 17}

	got := RollBaselineNumber(roller.NewScripted(7), sys, counts)
	if got != 7 {
		t.Errorf("baseline = %d, want 7", got)
	}
}

func TestRollBaselineNumber_DMTable(t *testing.T) {
	t.Parallel()

	type tc struct {
		name string
		sys  stars.System
		tot  int
		roll int
		want int
	}

	cases := []tc{
		{
			name: "primary has companion",
			sys: stars.System{
				Primary: stars.Compose(stars.ComposeOpts{Kind: stars.KindMainSequence, LuminosityClass: stars.V}),
				Companions: []stars.CompanionStar{
					{Star: stars.Star{}, OrbitClass: stars.OrbitCompanion, ParentIndex: -1},
				},
			},
			tot: 17, roll: 9, want: 9 - 2,
		},
		{
			name: "Class III primary",
			sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
				Kind: stars.KindMainSequence, LuminosityClass: stars.III,
			})},
			tot: 17, roll: 7, want: 7 + 2,
		},
		{
			name: "Class IV primary",
			sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
				Kind: stars.KindMainSequence, LuminosityClass: stars.IV,
			})},
			tot: 17, roll: 7, want: 7 + 1,
		},
		{
			name: "Class VI primary",
			sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
				Kind: stars.KindMainSequence, LuminosityClass: stars.VI,
			})},
			tot: 17, roll: 7, want: 7 - 1,
		},
		{
			name: "post-stellar primary",
			sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
				Kind: stars.KindWhiteDwarf, LuminosityClass: stars.D,
			})},
			tot: 17, roll: 9, want: 9 - 2,
		},
		{
			name: "total worlds < 6",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  3, roll: 9, want: 9 - 4,
		},
		{
			name: "total worlds 6-9",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  7, roll: 9, want: 9 - 3,
		},
		{
			name: "total worlds 10-12",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  11, roll: 9, want: 9 - 2,
		},
		{
			name: "total worlds 13-15",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  14, roll: 9, want: 9 - 1,
		},
		{
			name: "total worlds 18-20",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  19, roll: 9, want: 9 + 1,
		},
		{
			name: "total worlds > 20",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  25, roll: 9, want: 9 + 2,
		},
		{
			name: "secondary star DM",
			sys: stars.System{
				Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V}),
				Companions: []stars.CompanionStar{
					{Star: stars.Star{}, OrbitClass: stars.OrbitNear, ParentIndex: -1},
					{Star: stars.Star{}, OrbitClass: stars.OrbitFar, ParentIndex: -1},
				},
			},
			tot: 17, roll: 9, want: 9 - 2,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := RollBaselineNumber(roller.NewScripted(c.roll), c.sys, Counts{Total: c.tot})
			if got != c.want {
				t.Errorf("baseline = %d, want %d", got, c.want)
			}
		})
	}
}

func TestBaselineOrbit_3a_HZCOGTE1(t *testing.T) {
	t.Parallel()

	primary := Group{
		Designation: "A",
		Members:     []stars.Star{{Luminosity: 1.0}}, // HZCO = 3.0
		MAO:         0.03,
		Intervals:   []Interval{{Min: 0.03, Max: 20.0}},
	}
	// 2D = 5 → (5-7)/10 = -0.2. HZCO 3.0 + (-0.2) = 2.8.
	got := BaselineOrbit(roller.NewScripted(5), primary, primary.HZCO(), 5, 17)
	if math.Abs(got-2.8) > 0.01 {
		t.Errorf("BaselineOrbit = %v, want 2.8", got)
	}
}

func TestBaselineOrbit_3a_HZCOLT1(t *testing.T) {
	t.Parallel()

	primary := Group{
		Designation: "A",
		Members:     []stars.Star{{Luminosity: 0.04}},
		MAO:         0.01,
		Intervals:   []Interval{{Min: 0.01, Max: 20.0}},
	}
	hzco := primary.HZCO()
	// 2D = 9 → (9-7)/100 = 0.02 → hzco + 0.02
	got := BaselineOrbit(roller.NewScripted(9), primary, hzco, 5, 17)
	if math.Abs(got-(hzco+0.02)) > 0.005 {
		t.Errorf("BaselineOrbit = %v, want %v", got, hzco+0.02)
	}
}

func TestBaselineOrbit_3b_ColdSystem_MinGTE1(t *testing.T) {
	t.Parallel()
	// baselineN = -2 (< 1), HZCO = 3.0, totalWorlds = 5, MAO = 1.5
	primary := Group{
		Members:   []stars.Star{{Luminosity: 1.0}},
		MAO:       1.5,
		Intervals: []Interval{{Min: 1.5, Max: 20.0}},
	}
	// BaselineOrbit = 3.0 - (-2) + 5 + (2D-2)/10 = 10.0 + (7-2)/10 = 10.5
	got := BaselineOrbit(roller.NewScripted(7), primary, primary.HZCO(), -2, 5)
	if math.Abs(got-10.5) > 0.01 {
		t.Errorf("BaselineOrbit = %v, want 10.5", got)
	}
}

func TestBaselineOrbit_3c_HotSystem(t *testing.T) {
	t.Parallel()
	// baselineN = 8 > totalWorlds = 5, HZCO = 3.0
	// 3.0 - 8 + 5 = 0 (< 1.0 path):
	//   = 3.0 - (8 + 5 + (2D-7)/5)/10 = 3.0 - (13 + (10-7)/5)/10 = 3.0 - (13 + 0.6)/10 = 3.0 - 1.36 = 1.64
	primary := Group{
		Members:   []stars.Star{{Luminosity: 1.0}},
		MAO:       0.03,
		Intervals: []Interval{{Min: 0.03, Max: 20.0}},
	}

	got := BaselineOrbit(roller.NewScripted(10), primary, primary.HZCO(), 8, 5)
	if math.Abs(got-1.64) > 0.05 {
		t.Errorf("BaselineOrbit = %v, want ~1.64", got)
	}
}

func TestBaselineOrbit_SnapToAvailable(t *testing.T) {
	t.Parallel()
	// primary with hole around 3.0, roll formula into hole, assert snapping.
	primary := Group{
		Members:   []stars.Star{{Luminosity: 1.0}},
		MAO:       0.03,
		Intervals: []Interval{{Min: 0.03, Max: 2.5}, {Min: 3.5, Max: 20.0}},
	}
	// 3a path with roll 7 → variance 0 → BaselineOrbit = 3.0 (in hole 2.5-3.5).
	// Snap: nearest endpoint is 2.5 (snapped DOWN from 3.0), so snap direction
	// moves further DOWN into the lower interval. Snap roll: 2D=5 → magnitude
	// |5-7|/10 = 0.2 → 2.5 + (-1)*0.2 = 2.3.
	got := BaselineOrbit(roller.NewScripted(7, 5), primary, primary.HZCO(), 3, 17)
	if math.Abs(got-2.3) > 0.05 {
		t.Errorf("BaselineOrbit = %v, want ~2.3 (snap to 2.5, variance 0.2 into lower interval)", got)
	}
}

func TestBaselineOrbit_SnapToAvailable_HighRollStaysInZone(t *testing.T) {
	t.Parallel()
	// Same hole as above. With a 2D=10 snap roll under a naive signed-variance
	// implementation, 2.5 + (10-7)/10 = 2.8 would land BACK INSIDE the hole.
	// Per WBH p. 45 the variance must "always move the orbit into the
	// allowable zone" — magnitude |10-7|/10 = 0.3 in the snap direction (down).
	// Expected: 2.5 - 0.3 = 2.2.
	primary := Group{
		Members:   []stars.Star{{Luminosity: 1.0}},
		MAO:       0.03,
		Intervals: []Interval{{Min: 0.03, Max: 2.5}, {Min: 3.5, Max: 20.0}},
	}

	got := BaselineOrbit(roller.NewScripted(7, 10), primary, primary.HZCO(), 3, 17)
	if math.Abs(got-2.2) > 0.05 {
		t.Errorf("BaselineOrbit = %v, want ~2.2 (high snap roll must still move into zone)", got)
	}
}
