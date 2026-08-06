package stars

import (
	"errors"
	"math"
	"testing"
)

func TestMAO_ZedAa(t *testing.T) {
	t.Parallel()

	// G7 V should interpolate between G5 V (0.02) and K0 V (0.02);
	// expected MAO 0.02 (book uses 0.03 for Sol G2 V — different cell).
	zedAa := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: V,
		Mass:            0.929,
		Diameter:        0.967,
		Temperature:     5440,
	})

	got, err := MAO(zedAa)
	if err != nil {
		t.Fatalf("MAO: %v", err)
	}

	if math.Abs(got-0.02) > 1e-9 {
		t.Errorf("MAO(G7 V) = %v, want 0.02", got)
	}
}

func TestMAO_ZedB(t *testing.T) {
	t.Parallel()

	// K8 V interpolates K5 V (0.02) → M0 V (0.02); expected 0.02.
	zedB := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: V,
		Mass:            0.626,
		Diameter:        0.777,
		Temperature:     3980,
	})

	got, err := MAO(zedB)
	if err != nil {
		t.Fatalf("MAO: %v", err)
	}

	if math.Abs(got-0.02) > 1e-9 {
		t.Errorf("MAO(K8 V) = %v, want 0.02", got)
	}
}

func TestMAO_Sol(t *testing.T) {
	t.Parallel()

	// G2 V: G0 V (0.03) → G5 V (0.02); 2/5 between → 0.03 - (0.01 × 2/5) = 0.026.
	// Book reports 0.03 for Sol on the worked example survey form.
	// Allow small interpolation variance.
	sol := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: V,
		Mass:            1.000,
		Diameter:        1.000,
		Temperature:     5772,
	})

	got, err := MAO(sol)
	if err != nil {
		t.Fatalf("MAO: %v", err)
	}

	if math.Abs(got-0.03) > 0.005 {
		t.Errorf("MAO(G2 V) = %v, want ~0.03", got)
	}
}

func TestMAO_NoEntry(t *testing.T) {
	t.Parallel()

	// A0 VI is "—" in the book — no entry.
	a0vi := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'A', Subtype: 0},
		LuminosityClass: VI,
		Mass:            0.5,
		Diameter:        0.5,
		Temperature:     9000,
	})

	_, err := MAO(a0vi)
	if !errors.Is(err, ErrNoMAOForStar) {
		t.Errorf("MAO(A0 VI) error = %v, want ErrNoMAOForStar", err)
	}
}

func TestMAO_PostStellar(t *testing.T) {
	t.Parallel()

	bd := Star{Kind: KindBrownDwarf}

	_, err := MAO(bd)
	if !errors.Is(err, ErrPostStellarPrimaryUnsupported) {
		t.Errorf("MAO(BD) error = %v, want ErrPostStellarPrimaryUnsupported", err)
	}
}

// TestMAO_Protostar covers the seed 6724 regression: a protostar
// primary has no spectral type, so MAO previously fell through to the
// p.39 table lookup with a zero-valued SpectralType (Letter == 0),
// producing `no MAO row for "\x000"`. Protostars belong in the
// Special Circumstances bucket with the other no-spectral-type kinds.
func TestMAO_Protostar(t *testing.T) {
	t.Parallel()

	proto := Star{Kind: KindProtostar}

	_, err := MAO(proto)
	if !errors.Is(err, ErrPostStellarPrimaryUnsupported) {
		t.Errorf("MAO(protostar) error = %v, want ErrPostStellarPrimaryUnsupported", err)
	}
}

func TestMAO_CrossLetterInterpolation(t *testing.T) {
	t.Parallel()

	// O7 V brackets O5 V (0.30) → B0 V (0.18); frac = (7-5)/5 = 0.4.
	// Expected = 0.30 + (0.18 - 0.30) × 0.4 = 0.252.
	o7v := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'O', Subtype: 7},
		LuminosityClass: V,
		Mass:            30.0, Diameter: 6.6, Temperature: 36000,
	})

	got, err := MAO(o7v)
	if err != nil {
		t.Fatalf("MAO: %v", err)
	}

	want := 0.252
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("MAO(O7 V) = %v, want %v", got, want)
	}
}

// TestMAO_AggregateKindsClassifiable asserts Nebula, Star Cluster, and
// Anomaly primaries return the classifiable ErrPostStellarPrimaryUnsupported
// (wrapping ErrSpecialCircumstances) rather than a raw internal
// table-miss error — a regression guard for specialObjectAge letting
// these kinds survive the age step and reach the MAO lookup.
func TestMAO_AggregateKindsClassifiable(t *testing.T) {
	t.Parallel()

	for _, k := range []StarKind{KindNebula, KindStarCluster, KindAnomaly} {
		_, err := MAO(Star{Kind: k})
		if !errors.Is(err, ErrPostStellarPrimaryUnsupported) {
			t.Errorf("MAO(%s) = %v, want ErrPostStellarPrimaryUnsupported", k, err)
		}

		if !errors.Is(err, ErrSpecialCircumstances) {
			t.Errorf("MAO(%s) error does not wrap ErrSpecialCircumstances: %v", k, err)
		}
	}
}
