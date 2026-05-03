package worlds_test

import (
	"math"
	"testing"

	"wbh/stars"
	"wbh/worlds"
)

func TestZed_AvailableOrbits(t *testing.T) {
	t.Parallel()

	// WBH p. 40: Zed quintuple available orbits.
	//   Aab pair (G7 V + G8 V):  MAO 0.61, [[0.61, 5.10], [7.10, 10.10], [14.10, 20.00]], Total 13.39
	//   B (K8 V):                MAO 0.02, [[0.02, 1.10]], Total 1.08
	//   Cab pair (M0 V + D):     MAO 0.74, [[0.74, 7.10]], Total 6.36

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

	got, err := worlds.AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	if len(got.Groups) != 3 {
		t.Fatalf("groups = %d, want 3", len(got.Groups))
	}

	// Group Aab.
	aab := got.Groups[0]
	if aab.Designation != "Aab" {
		t.Errorf("groups[0].Designation = %q, want \"Aab\"", aab.Designation)
	}
	if math.Abs(aab.MAO-0.61) > 0.01 {
		t.Errorf("Aab MAO = %v, want 0.61", aab.MAO)
	}
	wantAab := []worlds.Interval{
		{Min: 0.61, Max: 5.10},
		{Min: 7.10, Max: 10.10},
		{Min: 14.10, Max: 20.00},
	}
	if !intervalsEqual(aab.Intervals, wantAab, 0.01) {
		t.Errorf("Aab intervals = %+v, want %+v", aab.Intervals, wantAab)
	}
	if math.Abs(aab.Total()-13.39) > 0.05 {
		t.Errorf("Aab Total = %v, want 13.39", aab.Total())
	}

	// Group B.
	bg := got.Groups[1]
	if bg.Designation != "B" {
		t.Errorf("groups[1].Designation = %q, want \"B\"", bg.Designation)
	}
	if math.Abs(bg.MAO-0.02) > 0.005 {
		t.Errorf("B MAO = %v, want 0.02", bg.MAO)
	}
	wantB := []worlds.Interval{{Min: 0.02, Max: 1.10}}
	if !intervalsEqual(bg.Intervals, wantB, 0.01) {
		t.Errorf("B intervals = %+v, want %+v", bg.Intervals, wantB)
	}
	if math.Abs(bg.Total()-1.08) > 0.01 {
		t.Errorf("B Total = %v, want 1.08", bg.Total())
	}

	// Group Cab.
	cab := got.Groups[2]
	if cab.Designation != "Cab" {
		t.Errorf("groups[2].Designation = %q, want \"Cab\"", cab.Designation)
	}
	if math.Abs(cab.MAO-0.74) > 0.01 {
		t.Errorf("Cab MAO = %v, want 0.74", cab.MAO)
	}
	wantCab := []worlds.Interval{{Min: 0.74, Max: 7.10}}
	if !intervalsEqual(cab.Intervals, wantCab, 0.01) {
		t.Errorf("Cab intervals = %+v, want %+v", cab.Intervals, wantCab)
	}
	if math.Abs(cab.Total()-6.36) > 0.01 {
		t.Errorf("Cab Total = %v, want 6.36", cab.Total())
	}
}

// intervalsEqual reports whether two interval slices match within tol on each bound.
func intervalsEqual(a, b []worlds.Interval, tol float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(a[i].Min-b[i].Min) > tol || math.Abs(a[i].Max-b[i].Max) > tol {
			return false
		}
	}
	return true
}

func TestSol_AvailableOrbits(t *testing.T) {
	t.Parallel()

	sol := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.000, Diameter: 1.000, Temperature: 5772,
	})
	sys := stars.System{Primary: sol}
	got, err := worlds.AvailableOrbits(sys)
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
	if math.Abs(g.Intervals[0].Max-20.0) > 1e-9 {
		t.Errorf("Max = %v, want 20.0", g.Intervals[0].Max)
	}

	// Sanity: full Total ≈ 19.97.
	if math.Abs(g.Total()-19.97) > 0.01 {
		t.Errorf("Total = %v, want ~19.97", g.Total())
	}
}
