package worlds

import (
	"math"
	"testing"

	"wbh/stars"
)

func TestGroup_HZCO_SingleStar(t *testing.T) {
	t.Parallel()
	sol := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V, Mass: 1.000, Diameter: 1.000, Temperature: 5772,
	})
	g := Group{Members: []stars.Star{sol}}
	got := g.HZCO()
	if math.Abs(got-3.0) > 0.05 {
		t.Errorf("Sol single-star HZCO = %.4f, want 3.0±0.05", got)
	}
}

func TestGroup_HZCO_Pair_ZedAab(t *testing.T) {
	t.Parallel()
	// Zed Aab pair: combined luminosity 1.419 (treated as ~1.4 by book) → HZCO 3.3.
	aa := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V, Mass: 0.929, Diameter: 0.967, Temperature: 5440,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V, Mass: 0.907, Diameter: 0.957, Temperature: 5360,
	})
	g := Group{Members: []stars.Star{aa, ab}}
	got := g.HZCO()
	if math.Abs(got-3.3) > 0.05 {
		t.Errorf("Zed Aab pair HZCO = %.4f, want 3.3±0.05", got)
	}
}
