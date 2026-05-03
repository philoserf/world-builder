package stars

import (
	"math"
	"testing"
)

func TestStar_HZCO_Sol(t *testing.T) {
	t.Parallel()

	sol := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: V,
		Mass:            1.000, Diameter: 1.000, Temperature: 5772, AgeGyr: 4.568,
	})
	got := sol.HZCO()
	if math.Abs(got-3.0) > 0.05 {
		t.Errorf("Sol HZCO = %.4f, want 3.0±0.05", got)
	}
}

func TestStar_HZCO_ZedB(t *testing.T) {
	t.Parallel()

	// Zed B: K8 V, L=0.136 → HZCO ≈ 0.92 (WBH p. 43).
	b := Star{Luminosity: 0.136}
	got := b.HZCO()
	if math.Abs(got-0.92) > 0.05 {
		t.Errorf("Zed B HZCO = %.4f, want 0.92±0.05", got)
	}
}
