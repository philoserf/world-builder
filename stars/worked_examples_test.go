package stars_test

import (
	"math"
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestSolTerra_p35(t *testing.T) {
	// WBH p. 35 — Terra/Sol example (fully specified, not rolled).
	sol := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.000,
		Diameter:        1.000,
		Temperature:     5772,
		AgeGyr:          4.568,
	})
	if sol.SpectralType != (stars.SpectralType{Letter: 'G', Subtype: 2}) {
		t.Fatalf("spectral type wrong: %v", sol.SpectralType)
	}
	if sol.LuminosityClass != stars.V {
		t.Fatalf("class wrong: %v", sol.LuminosityClass)
	}
	checks := []struct {
		name string
		got  float64
		want float64
	}{
		{"mass", sol.Mass, 1.0},
		{"diameter", sol.Diameter, 1.0},
		{"temperature", sol.Temperature, 5772},
		{"luminosity", sol.Luminosity, 1.0},
		{"age", sol.AgeGyr, 4.568},
	}
	for _, c := range checks {
		if math.Abs(c.got-c.want) > 1e-9 {
			t.Errorf("%s: got %v want %v", c.name, c.got, c.want)
		}
	}
}

func TestZedPrimaryOnly_p17_p21(t *testing.T) {
	// WBH pp. 16–21 — Zed (G7 V) primary star, no companions.
	// Drive rolls verbatim from the book:
	//   2D=9 -> "G" type
	//   2D=6 -> Numeric subtype 7 (G7)
	//   2D-7=+2 mass variance -> 0.929
	//   2D-7=+1 diameter variance -> 0.967
	//   1D=3, D3=2, d10=3 -> 6.3 Gyr
	r := roller.NewScripted(
		9, // primary type 2D
		6, // subtype 2D
		2, // mass variance 2D-7
		1, // diameter variance 2D-7
		3, // age 1D
		2, // age D3
		3, // age d10
	)
	star, err := stars.GenerateMainSequenceStar(r, stars.GenerateOpts{
		WithVariance: true,
		Accuracy:     2,
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if star.SpectralType != (stars.SpectralType{Letter: 'G', Subtype: 7}) {
		t.Fatalf("spectral type: got %v want G7", star.SpectralType)
	}
	if star.LuminosityClass != stars.V {
		t.Fatalf("class: got %v want V", star.LuminosityClass)
	}
	checks := []struct {
		name string
		got  float64
		want float64
		tol  float64
	}{
		{"mass", star.Mass, 0.929, 2e-3},
		{"diameter", star.Diameter, 0.967, 2e-3},
		{"temperature", star.Temperature, 5440, 2e-3},
		{"luminosity", star.Luminosity, 0.738, 2e-3},
		{"age", star.AgeGyr, 6.3, 1e-9},
	}
	for _, c := range checks {
		if math.Abs(c.got-c.want) > c.tol {
			t.Errorf("%s: got %v want %v (tol %v)", c.name, c.got, c.want, c.tol)
		}
	}
}
