package worlds

import (
	"errors"
	"math"
	"testing"

	"github.com/philoserf/world-builder/stars"
)

func TestGroup_Total_SingleInterval(t *testing.T) {
	t.Parallel()

	g := Group{
		Designation: "A",
		MAO:         0.03,
		Intervals:   []Interval{{Min: 0.03, Max: 20.0}},
	}
	got := g.Total()
	want := 19.97
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("Total = %v, want %v", got, want)
	}
}

func TestGroup_Total_MultiInterval(t *testing.T) {
	t.Parallel()

	// Zed Aab from WBH p. 40: 0.61–5.10, 7.10–10.10, 14.10–20.00 → 13.39.
	g := Group{
		Designation: "Aab",
		MAO:         0.61,
		Intervals: []Interval{
			{Min: 0.61, Max: 5.10},
			{Min: 7.10, Max: 10.10},
			{Min: 14.10, Max: 20.00},
		},
	}
	got := g.Total()
	want := 13.39
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("Total = %v, want %v", got, want)
	}
}

func TestGroup_Total_Empty(t *testing.T) {
	t.Parallel()

	g := Group{Intervals: nil}
	if got := g.Total(); got != 0 {
		t.Errorf("Total = %v, want 0", got)
	}
}

func TestGroup_Contains(t *testing.T) {
	t.Parallel()

	g := Group{
		Intervals: []Interval{
			{Min: 0.61, Max: 5.10},
			{Min: 7.10, Max: 10.10},
		},
	}

	tests := []struct {
		name  string
		orbit float64
		want  bool
	}{
		{"below first interval", 0.5, false},
		{"first interval lower endpoint", 0.61, true},
		{"first interval middle", 3.0, true},
		{"first interval upper endpoint", 5.10, true},
		{"between intervals (lower)", 6.0, false},
		{"between intervals (upper)", 7.0, false},
		{"second interval lower endpoint", 7.10, true},
		{"second interval upper endpoint", 10.10, true},
		{"above all intervals", 15.0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := g.Contains(tc.orbit); got != tc.want {
				t.Errorf("Contains(%v) = %v, want %v", tc.orbit, got, tc.want)
			}
		})
	}
}

func TestIdentifyGroups_SinglePrimary(t *testing.T) {
	t.Parallel()

	sys := stars.System{
		Primary: stars.Compose(stars.ComposeOpts{
			Kind:            stars.KindMainSequence,
			SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
			LuminosityClass: stars.V,
			Mass:            1.0, Diameter: 1.0, Temperature: 5772,
		}),
		Companions: nil,
	}
	groups := identifyGroups(sys)
	if len(groups) != 1 {
		t.Fatalf("groups = %d, want 1", len(groups))
	}
	if groups[0].Designation != "A" {
		t.Errorf("Designation = %q, want \"A\"", groups[0].Designation)
	}
	if len(groups[0].Members) != 1 {
		t.Errorf("Members = %d, want 1", len(groups[0].Members))
	}
}

func TestIdentifyGroups_PrimaryWithCompanion(t *testing.T) {
	t.Parallel()

	primary := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass:            0.929, Diameter: 0.967, Temperature: 5440,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.907, Diameter: 0.957, Temperature: 5360,
	})
	sys := stars.System{
		Primary: primary,
		Companions: []stars.CompanionStar{
			{Star: ab, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.09, Eccentricity: 0.11, ParentIndex: -1},
		},
	}
	groups := identifyGroups(sys)
	if len(groups) != 1 {
		t.Fatalf("groups = %d, want 1 (companion folds into primary)", len(groups))
	}
	if groups[0].Designation != "Aab" {
		t.Errorf("Designation = %q, want \"Aab\"", groups[0].Designation)
	}
	if len(groups[0].Members) != 2 {
		t.Errorf("Members = %d, want 2", len(groups[0].Members))
	}
}

func TestAvailableOrbits_SoleMainSequence(t *testing.T) {
	t.Parallel()

	// Single G2 V primary, no companions. Expected: one group "A" with
	// MAO ~0.03 (interpolated G0→G5) and intervals [[MAO, 20.0]].
	sol := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.0, Diameter: 1.0, Temperature: 5772,
	})
	sys := stars.System{Primary: sol}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	if len(got.Groups) != 1 {
		t.Fatalf("groups = %d, want 1", len(got.Groups))
	}
	g := got.Groups[0]
	if g.Designation != "A" {
		t.Errorf("Designation = %q, want \"A\"", g.Designation)
	}
	if math.Abs(g.MAO-0.03) > 0.005 {
		t.Errorf("MAO = %v, want ~0.03", g.MAO)
	}
	if len(g.Intervals) != 1 {
		t.Fatalf("intervals = %d, want 1", len(g.Intervals))
	}
	iv := g.Intervals[0]
	if math.Abs(iv.Min-g.MAO) > 1e-9 {
		t.Errorf("Min = %v, want MAO %v", iv.Min, g.MAO)
	}
	if math.Abs(iv.Max-20.0) > 1e-9 {
		t.Errorf("Max = %v, want 20.0", iv.Max)
	}
}

func TestAvailableOrbits_PostStellarPrimary(t *testing.T) {
	t.Parallel()

	sys := stars.System{
		Primary: stars.Star{Kind: stars.KindWhiteDwarf, Mass: 0.5},
	}
	_, err := AvailableOrbits(sys)
	if !errors.Is(err, stars.ErrPostStellarPrimaryUnsupported) {
		t.Errorf("err = %v, want stars.ErrPostStellarPrimaryUnsupported", err)
	}
}

func TestAvailableOrbits_Rule2_CompanionEccentricity(t *testing.T) {
	t.Parallel()

	// Aab pair: primary G7 V, companion G8 V with eccentricity 0.11.
	// Expected: pair group MAO becomes 0.50 + 0.11 = 0.61.
	aa := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass:            0.929, Diameter: 0.967, Temperature: 5440,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.907, Diameter: 0.957, Temperature: 5360,
	})
	sys := stars.System{
		Primary: aa,
		Companions: []stars.CompanionStar{
			{Star: ab, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.09, Eccentricity: 0.11, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	g := got.Groups[0]
	if math.Abs(g.MAO-0.61) > 1e-9 {
		t.Errorf("Aab MAO = %v, want 0.61", g.MAO)
	}
	if math.Abs(g.Intervals[0].Min-0.61) > 1e-9 {
		t.Errorf("Aab interval Min = %v, want 0.61", g.Intervals[0].Min)
	}
}

func TestAvailableOrbits_Rule5_SecondaryExclusion(t *testing.T) {
	t.Parallel()

	// Construct: primary G2 V, single secondary at Orbit# 6.10 (Near).
	// Expected: primary intervals = [[MAO, 5.10], [7.10, 20.0]].
	a := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.0, Diameter: 1.0, Temperature: 5772,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.626, Diameter: 0.777, Temperature: 3980,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.0, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	primary := got.Groups[0]
	if len(primary.Intervals) != 2 {
		t.Fatalf("primary intervals = %d, want 2: %+v", len(primary.Intervals), primary.Intervals)
	}
	if math.Abs(primary.Intervals[0].Max-5.10) > 1e-9 {
		t.Errorf("first interval Max = %v, want 5.10", primary.Intervals[0].Max)
	}
	if math.Abs(primary.Intervals[1].Min-7.10) > 1e-9 {
		t.Errorf("second interval Min = %v, want 7.10", primary.Intervals[1].Min)
	}
	if math.Abs(primary.Intervals[1].Max-20.0) > 1e-9 {
		t.Errorf("second interval Max = %v, want 20.0", primary.Intervals[1].Max)
	}
}

func TestAvailableOrbits_Rule6_EccentricSecondary(t *testing.T) {
	t.Parallel()

	// Zed Cab pair (Far) at Orbit# 12.10 with eccentricity 0.47 (> 0.2).
	// Rule 5 base width = ±1: exclusion is (11.10, 13.10).
	// Rule 6 widens by ±1 because ecc > 0.2: exclusion is (10.10, 14.10).
	// M0 V MAO < 0.2, so no MAO expansion. Final primary intervals:
	// [[MAO, 10.10], [14.10, 20.0]] — matches WBH p. 40.
	a := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass:            0.929, Diameter: 0.967, Temperature: 5440,
	})
	ca := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'M', Subtype: 0},
		LuminosityClass: stars.V,
		Mass:            0.510, Diameter: 0.728, Temperature: 3700,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: ca, OrbitClass: stars.OrbitFar, OrbitNumber: 12.10, Eccentricity: 0.47, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	primary := got.Groups[0]
	if len(primary.Intervals) != 2 {
		t.Fatalf("primary intervals = %d, want 2: %+v", len(primary.Intervals), primary.Intervals)
	}
	if math.Abs(primary.Intervals[0].Max-10.10) > 1e-9 {
		t.Errorf("first interval Max = %v, want 10.10", primary.Intervals[0].Max)
	}
	if math.Abs(primary.Intervals[1].Min-14.10) > 1e-9 {
		t.Errorf("second interval Min = %v, want 14.10", primary.Intervals[1].Min)
	}
}

func TestAvailableOrbits_Rule7_NearEccentricityGT05(t *testing.T) {
	t.Parallel()

	// Near at Orbit# 6.0 with ecc 0.6 (> 0.5).
	// Rule 5: ±1, Rule 6: ±1 more (= ±2), Rule 7: ±1 more (= ±3).
	// Expected exclusion: (3.0, 9.0).
	a := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.0, Diameter: 1.0, Temperature: 5772,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.626, Diameter: 0.777, Temperature: 3980,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.0, Eccentricity: 0.6, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	primary := got.Groups[0]
	if math.Abs(primary.Intervals[0].Max-3.0) > 1e-9 {
		t.Errorf("first Max = %v, want 3.0", primary.Intervals[0].Max)
	}
	if math.Abs(primary.Intervals[1].Min-9.0) > 1e-9 {
		t.Errorf("second Min = %v, want 9.0", primary.Intervals[1].Min)
	}
}

func TestAvailableOrbits_Rule7_FarDoesNotTrigger(t *testing.T) {
	t.Parallel()

	// Far at Orbit# 12 with ecc 0.6: rule 7 must NOT apply.
	// Expected exclusion: ±2 only (rules 5+6).
	a := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.0, Diameter: 1.0, Temperature: 5772,
	})
	far := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'M', Subtype: 0},
		LuminosityClass: stars.V,
		Mass:            0.510, Diameter: 0.728, Temperature: 3700,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: far, OrbitClass: stars.OrbitFar, OrbitNumber: 12.0, Eccentricity: 0.6, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	primary := got.Groups[0]
	// Exclusion (10, 14): primary intervals [[MAO, 10.0], [14.0, 20.0]].
	if math.Abs(primary.Intervals[0].Max-10.0) > 1e-9 {
		t.Errorf("first Max = %v, want 10.0", primary.Intervals[0].Max)
	}
	if math.Abs(primary.Intervals[1].Min-14.0) > 1e-9 {
		t.Errorf("second Min = %v, want 14.0", primary.Intervals[1].Min)
	}
}

func TestAvailableOrbits_Rule8_SecondaryOwnRange(t *testing.T) {
	t.Parallel()

	// Zed B at Orbit# 6.10: own range = [B's MAO, 6.10 - 3] = [0.02, 3.10].
	a := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.0, Diameter: 1.0, Temperature: 5772,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.626, Diameter: 0.777, Temperature: 3980,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.08, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	if len(got.Groups) != 2 {
		t.Fatalf("groups = %d, want 2", len(got.Groups))
	}
	bg := got.Groups[1]
	if bg.Designation != "B" {
		t.Errorf("Designation = %q, want \"B\"", bg.Designation)
	}
	if len(bg.Intervals) != 1 {
		t.Fatalf("intervals = %d, want 1", len(bg.Intervals))
	}
	if math.Abs(bg.Intervals[0].Min-bg.MAO) > 1e-9 {
		t.Errorf("Min = %v, want MAO %v", bg.Intervals[0].Min, bg.MAO)
	}
	if math.Abs(bg.Intervals[0].Max-3.10) > 1e-9 {
		t.Errorf("Max = %v, want 3.10", bg.Intervals[0].Max)
	}
}

func TestAvailableOrbits_Rules9to11_ZedB(t *testing.T) {
	t.Parallel()

	// Zed B reductions (WBH p. 40):
	//  Rule 8: own range to 3.10.
	//  Rule 9: Far is adjacent → -1 → 2.10.
	//  Rule 10: Far ecc 0.47 > 0.2 → -1 → 1.10.
	//  Rule 11: B ecc 0.08 (not > 0.5) → no reduction.
	// Final B max = 1.10.
	a := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass:            0.929, Diameter: 0.967, Temperature: 5440,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.626, Diameter: 0.777, Temperature: 3980,
	})
	ca := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'M', Subtype: 0},
		LuminosityClass: stars.V,
		Mass:            0.510, Diameter: 0.728, Temperature: 3700,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.08, ParentIndex: -1},
			{Star: ca, OrbitClass: stars.OrbitFar, OrbitNumber: 12.10, Eccentricity: 0.47, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	if len(got.Groups) != 3 {
		t.Fatalf("groups = %d, want 3", len(got.Groups))
	}
	bg := got.Groups[1]
	if math.Abs(bg.Intervals[0].Max-1.10) > 1e-9 {
		t.Errorf("B Max = %v, want 1.10", bg.Intervals[0].Max)
	}
}

func TestSubtract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input []Interval
		exMin float64
		exMax float64
		want  []Interval
	}{
		{
			name:  "exclusion fully contained",
			input: []Interval{{Min: 0, Max: 10}},
			exMin: 3, exMax: 7,
			want: []Interval{{Min: 0, Max: 3}, {Min: 7, Max: 10}},
		},
		{
			name:  "exclusion straddles left edge",
			input: []Interval{{Min: 0, Max: 10}},
			exMin: -1, exMax: 5,
			want: []Interval{{Min: 5, Max: 10}},
		},
		{
			name:  "exclusion straddles right edge",
			input: []Interval{{Min: 0, Max: 10}},
			exMin: 5, exMax: 15,
			want: []Interval{{Min: 0, Max: 5}},
		},
		{
			name:  "exclusion fully covers interval",
			input: []Interval{{Min: 2, Max: 8}},
			exMin: 0, exMax: 10,
			want: []Interval{},
		},
		{
			name:  "zero-width exclusion is no-op",
			input: []Interval{{Min: 0, Max: 10}},
			exMin: 5, exMax: 5,
			want: []Interval{{Min: 0, Max: 10}},
		},
		{
			name:  "inverted exclusion is no-op",
			input: []Interval{{Min: 0, Max: 10}},
			exMin: 7, exMax: 3,
			want: []Interval{{Min: 0, Max: 10}},
		},
		{
			name:  "no overlap (before)",
			input: []Interval{{Min: 5, Max: 10}},
			exMin: 0, exMax: 3,
			want: []Interval{{Min: 5, Max: 10}},
		},
		{
			name:  "no overlap (after)",
			input: []Interval{{Min: 0, Max: 5}},
			exMin: 7, exMax: 10,
			want: []Interval{{Min: 0, Max: 5}},
		},
		{
			name:  "exclusion spans multiple intervals",
			input: []Interval{{Min: 0, Max: 5}, {Min: 8, Max: 15}, {Min: 20, Max: 25}},
			exMin: 3, exMax: 22,
			want: []Interval{{Min: 0, Max: 3}, {Min: 22, Max: 25}},
		},
		{
			name:  "empty input",
			input: nil,
			exMin: 0, exMax: 10,
			want: []Interval{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := subtract(tc.input, tc.exMin, tc.exMax)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d intervals %+v, want %d %+v", len(got), got, len(tc.want), tc.want)
			}
			for i := range got {
				if math.Abs(got[i].Min-tc.want[i].Min) > 1e-9 || math.Abs(got[i].Max-tc.want[i].Max) > 1e-9 {
					t.Errorf("interval[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestIdentifyGroups_SourceCompanion(t *testing.T) {
	t.Parallel()

	aa := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V, Mass: 0.929, Diameter: 0.967, Temperature: 5440,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V, Mass: 0.907, Diameter: 0.957, Temperature: 5360,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V, Mass: 0.626, Diameter: 0.777, Temperature: 3980,
	})
	sys := stars.System{
		Primary: aa,
		Companions: []stars.CompanionStar{
			{Star: ab, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.09, Eccentricity: 0.11, ParentIndex: -1},
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.08, ParentIndex: -1},
		},
	}

	groups := identifyGroups(sys)
	if len(groups) != 2 {
		t.Fatalf("groups = %d, want 2", len(groups))
	}
	if groups[0].sourceCompanion != nil {
		t.Errorf("primary group sourceCompanion = %+v, want nil", groups[0].sourceCompanion)
	}
	if groups[1].sourceCompanion == nil {
		t.Fatalf("secondary group sourceCompanion = nil, want non-nil")
	}
	if groups[1].sourceCompanion.OrbitClass != stars.OrbitNear {
		t.Errorf("secondary sourceCompanion.OrbitClass = %v, want OrbitNear", groups[1].sourceCompanion.OrbitClass)
	}
	if math.Abs(groups[1].sourceCompanion.OrbitNumber-6.10) > 1e-9 {
		t.Errorf("secondary sourceCompanion.OrbitNumber = %v, want 6.10", groups[1].sourceCompanion.OrbitNumber)
	}
}

func TestIdentifyGroups_ZedQuintuple(t *testing.T) {
	t.Parallel()

	// Construct the Zed quintuple in pieces and verify three groups.
	// Aa + Ab → Aab; B alone → B; Ca + Cb → Cab.
	aa := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass:            0.929, Diameter: 0.967, Temperature: 5440,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.907, Diameter: 0.957, Temperature: 5360,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.626, Diameter: 0.777, Temperature: 3980,
	})
	ca := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'M', Subtype: 0},
		LuminosityClass: stars.V,
		Mass:            0.510, Diameter: 0.728, Temperature: 3700,
	})
	cb := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindWhiteDwarf,
		Mass: 0.490, Diameter: 0.017, Temperature: 6700,
	})
	sys := stars.System{
		Primary: aa,
		Companions: []stars.CompanionStar{
			// Index 0: Ab is companion of primary.
			{Star: ab, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.09, Eccentricity: 0.11, ParentIndex: -1},
			// Index 1: B is Near secondary.
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.08, ParentIndex: -1},
			// Index 2: Ca is Far secondary.
			{Star: ca, OrbitClass: stars.OrbitFar, OrbitNumber: 12.10, Eccentricity: 0.47, ParentIndex: -1},
			// Index 3: Cb is companion of Ca (parent index 2).
			{Star: cb, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.21, Eccentricity: 0.24, ParentIndex: 2},
		},
	}
	// Zed: Close is absent. Designations: primary group "Aab",
	// Near (only Close-relative slot left) becomes "B",
	// Far becomes "Cab" (with its own companion folded in).
	groups := identifyGroups(sys)
	if len(groups) != 3 {
		t.Fatalf("groups = %d, want 3", len(groups))
	}
	wantDesig := []string{"Aab", "B", "Cab"}
	for i, g := range groups {
		if g.Designation != wantDesig[i] {
			t.Errorf("groups[%d].Designation = %q, want %q", i, g.Designation, wantDesig[i])
		}
	}
	// Verify pair groups have companionEcc set.
	if math.Abs(groups[0].companionEcc-0.11) > 1e-9 {
		t.Errorf("Aab companionEcc = %v, want 0.11", groups[0].companionEcc)
	}
	if math.Abs(groups[2].companionEcc-0.24) > 1e-9 {
		t.Errorf("Cab companionEcc = %v, want 0.24", groups[2].companionEcc)
	}
	// Solo groups have companionEcc = 0.
	if groups[1].companionEcc != 0 {
		t.Errorf("B companionEcc = %v, want 0", groups[1].companionEcc)
	}

	// Verify sourceCompanion pointer identity for all three groups.
	// groups[0] is the primary (Aab) — must be nil.
	if groups[0].sourceCompanion != nil {
		t.Errorf("Aab sourceCompanion = %+v, want nil", groups[0].sourceCompanion)
	}
	// groups[1] is B (Near, Companions index 1).
	if groups[1].sourceCompanion == nil {
		t.Fatalf("B sourceCompanion = nil, want non-nil")
	}
	if groups[1].sourceCompanion.OrbitClass != stars.OrbitNear {
		t.Errorf("B sourceCompanion.OrbitClass = %v, want OrbitNear", groups[1].sourceCompanion.OrbitClass)
	}
	if math.Abs(groups[1].sourceCompanion.OrbitNumber-6.10) > 1e-9 {
		t.Errorf("B sourceCompanion.OrbitNumber = %v, want 6.10", groups[1].sourceCompanion.OrbitNumber)
	}
	// groups[2] is Cab (Far, Companions index 2).
	if groups[2].sourceCompanion == nil {
		t.Fatalf("Cab sourceCompanion = nil, want non-nil")
	}
	if groups[2].sourceCompanion.OrbitClass != stars.OrbitFar {
		t.Errorf("Cab sourceCompanion.OrbitClass = %v, want OrbitFar", groups[2].sourceCompanion.OrbitClass)
	}
	if math.Abs(groups[2].sourceCompanion.OrbitNumber-12.10) > 1e-9 {
		t.Errorf("Cab sourceCompanion.OrbitNumber = %v, want 12.10", groups[2].sourceCompanion.OrbitNumber)
	}
}
