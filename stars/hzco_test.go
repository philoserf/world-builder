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

// hzcoTablePage42 is the WBH p. 42 HZCO# table, used as a verification
// fixture for the formula-based Star.HZCO(). Cells with no value (the
// book's "—") are omitted from the map.
//
// Outer key: spectral type ("O0", "B5", "G2", ...). Inner map: luminosity
// class to expected HZCO# value.
var hzcoTablePage42 = map[string]map[LuminosityClass]float64{
	"O0": {Ia: 14.5, Ib: 14.4, II: 14.3, III: 14.3, V: 14.2, VI: 7.3},
	"O5": {Ia: 13.7, Ib: 13.5, II: 13.4, III: 13.2, V: 12.9, VI: 6.7},
	"B0": {Ia: 12.8, Ib: 12.2, II: 12.0, III: 11.7, IV: 11.4, V: 11.2, VI: 6.0},
	"B5": {Ia: 12.3, Ib: 11.1, II: 10.2, III: 9.0, IV: 8.6, V: 8.2, VI: 5.2},
	"A0": {Ia: 12.2, Ib: 10.9, II: 10.2, III: 7.5, IV: 7.2, V: 6.3},
	"A5": {Ia: 12.1, Ib: 10.8, II: 10.1, III: 6.9, IV: 6.1, V: 5.5},
	"F0": {Ia: 12.1, Ib: 10.8, II: 10.1, III: 6.7, IV: 5.9, V: 5.0},
	"F5": {Ia: 12.1, Ib: 10.8, II: 10.1, III: 6.2, IV: 4.7, V: 4.2},
	"G0": {Ia: 12.1, Ib: 10.8, II: 10.1, III: 7.1, IV: 5.2, V: 3.3},
	"G5": {Ia: 12.1, Ib: 10.8, II: 10.1, III: 7.4, IV: 5.4, V: 2.6, VI: 2.5},
	"K0": {Ia: 12.1, Ib: 10.8, II: 10.2, III: 7.6, IV: 5.8, V: 2.1, VI: 1.9},
	"K5": {Ia: 12.1, Ib: 10.9, II: 10.2, III: 8.1, V: 1.2, VI: 1.3},
	"M0": {Ia: 12.2, Ib: 11.0, II: 10.2, III: 8.2, V: 0.72, VI: 0.40},
	"M5": {Ia: 12.1, Ib: 11.1, II: 10.2, III: 8.4, V: 0.13, VI: 0.07},
	"M9": {Ia: 12.0, Ib: 10.8, II: 10.1, III: 8.8, V: 0.04, VI: 0.03},
}

// hzcoTableP42Inconsistent lists cells where the WBH p.42 HZCO# table is
// internally inconsistent with the WBH p.19 luminosity table. For these
// cells, applying the formula to the p.19 luminosity produces an HZCO#
// that diverges by >5% from p.42. The formula itself is correct; the
// discrepancy is in the book's own published tables.
//
// Investigated 2026-05-02: p.19 luminosities required to reproduce the
// p.42 values differ substantially from the encoded StarLuminosity values:
//
//	G5 VI: p.19 L=0.43, needs L≈0.72 (67% higher) → rel err 26%
//	K0 VI: p.19 L=0.23, needs L≈0.45 (96% higher) → rel err 33%
//	K5 VI: p.19 L=0.083, needs L≈0.24 (189% higher) → rel err 45%
//	M9 V:  p.19 L=0.00029, needs L≈0.00025 (14% lower) → rel err 6%
//	M9 VI: p.19 L=0.00019, needs L≈0.00014 (26% lower) → rel err 15%
var hzcoTableP42Inconsistent = map[string]map[LuminosityClass]bool{
	"G5": {VI: true},
	"K0": {VI: true},
	"K5": {VI: true},
	"M9": {V: true, VI: true},
}

func TestStar_HZCO_TableFidelity(t *testing.T) {
	t.Parallel()

	const tolerance = 0.05 // ±5%

	for typeStr, row := range hzcoTablePage42 {
		st, err := ParseSpectralType(typeStr)
		if err != nil {
			t.Fatalf("ParseSpectralType(%q): %v", typeStr, err)
		}
		for lc, want := range row {
			if hzcoTableP42Inconsistent[typeStr][lc] {
				continue // known p.19 vs p.42 inter-table discrepancy; see comment above
			}
			lum, err := ComputeLuminosityFromTable(st, lc)
			if err != nil {
				continue // book-blank cells already excluded above; defensive
			}
			s := Star{
				SpectralType:    st,
				LuminosityClass: lc,
				Luminosity:      lum,
			}
			got := s.HZCO()
			rel := math.Abs(got-want) / want
			if rel > tolerance {
				t.Errorf("HZCO(%s %s) = %.4f, want %.4f (rel err %.3f > %.3f)",
					typeStr, lc, got, want, rel, tolerance)
			}
		}
	}
}
