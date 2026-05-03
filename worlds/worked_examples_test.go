package worlds_test

import (
	"math"
	"testing"

	"wbh/stars"
	"wbh/worlds"
)

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
