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

func TestCompositeHZCO_ZedAab(t *testing.T) {
	t.Parallel()

	// Zed Aab: combined luminosity 1.419 → HZCO ≈ 3.3 (WBH p. 42).
	aa := Star{Luminosity: 0.738}
	ab := Star{Luminosity: 0.681}
	got := CompositeHZCO(aa, ab)
	if math.Abs(got-3.3) > 0.05 {
		t.Errorf("Zed Aab CompositeHZCO = %.4f, want 3.3±0.05", got)
	}
}

func TestCompositeHZCO_ZedCab(t *testing.T) {
	t.Parallel()

	// Zed Cab: combined luminosity 0.0896 → HZCO ≈ 0.75 (WBH p. 43).
	ca := Star{Luminosity: 0.0895}
	cb := Star{Luminosity: 0.000525}
	got := CompositeHZCO(ca, cb)
	if math.Abs(got-0.75) > 0.05 {
		t.Errorf("Zed Cab CompositeHZCO = %.4f, want 0.75±0.05", got)
	}
}

func TestCompositeHZCO_CorellaAab(t *testing.T) {
	t.Parallel()

	// Corella Aab: combined luminosity 1.725 → HZCO ≈ 3.5 (WBH p. 62).
	a := Star{Luminosity: 1.045}
	b := Star{Luminosity: 0.681}
	got := CompositeHZCO(a, b)
	if math.Abs(got-3.5) > 0.05 {
		t.Errorf("Corella Aab CompositeHZCO = %.4f, want 3.5±0.05", got)
	}
}

func TestCompositeHZCO_Empty(t *testing.T) {
	t.Parallel()

	if got := CompositeHZCO(); got != 0 {
		t.Errorf("CompositeHZCO() = %v, want 0", got)
	}
}
