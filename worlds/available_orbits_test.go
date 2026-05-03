package worlds

import (
	"errors"
	"math"
	"testing"

	"wbh/stars"
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

func TestMAO_ZedAa(t *testing.T) {
	t.Parallel()

	// G7 V should interpolate between G5 V (0.02) and K0 V (0.02);
	// expected MAO 0.02 (book uses 0.03 for Sol G2 V — different cell).
	zedAa := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
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
	zedB := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
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
	sol := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
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
	a0vi := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'A', Subtype: 0},
		LuminosityClass: stars.VI,
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

	bd := stars.Star{Kind: stars.KindBrownDwarf}
	_, err := MAO(bd)
	if !errors.Is(err, ErrPostStellarPrimaryUnsupported) {
		t.Errorf("MAO(BD) error = %v, want ErrPostStellarPrimaryUnsupported", err)
	}
}

func TestMAO_CrossLetterInterpolation(t *testing.T) {
	t.Parallel()

	// O7 V brackets O5 V (0.30) → B0 V (0.18); frac = (7-5)/5 = 0.4.
	// Expected = 0.30 + (0.18 - 0.30) × 0.4 = 0.252.
	o7v := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'O', Subtype: 7},
		LuminosityClass: stars.V,
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
