package worlds

import (
	"testing"

	"wbh/roller"
)

func TestRollSurfaceDistribution_TableValues(t *testing.T) {
	cases := []struct {
		twoD int
		want int
	}{
		{2, 0},   // 2D=2 → 2-2=0
		{4, 2},   // 2D=4 → 2-2=2
		{7, 5},   // 2D=7 → 2-2=5 (Mixed)
		{12, 10}, // 2D=12 → 2-2=10 (cap at A)
	}
	for _, c := range cases {
		r := roller.NewScripted(c.twoD)
		got, err := RollSurfaceDistribution(r)
		if err != nil {
			t.Fatalf("twoD=%d: unexpected error: %v", c.twoD, err)
		}
		if got != c.want {
			t.Errorf("twoD=%d: got %d, want %d", c.twoD, got, c.want)
		}
	}
}

func TestDescribeSurfaceDistribution(t *testing.T) {
	cases := []struct {
		code int
		want string
	}{
		{0, "Extremely Dispersed"},
		{1, "Very Dispersed"},
		{2, "Dispersed"},
		{3, "Scattered"},
		{4, "Slightly Scattered"},
		{5, "Mixed"},
		{6, "Slightly Skewed"},
		{7, "Skewed"},
		{8, "Concentrated"},
		{9, "Very Concentrated"},
		{10, "Extremely Concentrated"},
	}
	for _, c := range cases {
		got := DescribeSurfaceDistribution(c.code)
		if got != c.want {
			t.Errorf("code=%d: got %q, want %q", c.code, got, c.want)
		}
	}
}

func TestDetermineFundamentalGeography(t *testing.T) {
	t.Run("hydro_6_plus_is_ocean", func(t *testing.T) {
		for h := 6; h <= 10; h++ {
			r := roller.NewScripted() // no rolls needed
			got, err := DetermineFundamentalGeography(r, h)
			if err != nil {
				t.Fatalf("hydro=%d: %v", h, err)
			}
			if got != GeographyOcean {
				t.Errorf("hydro=%d: got %v, want GeographyOcean", h, got)
			}
		}
	})
	t.Run("hydro_4_minus_is_land", func(t *testing.T) {
		for h := 0; h <= 4; h++ {
			r := roller.NewScripted() // no rolls needed
			got, err := DetermineFundamentalGeography(r, h)
			if err != nil {
				t.Fatalf("hydro=%d: %v", h, err)
			}
			if got != GeographyLand {
				t.Errorf("hydro=%d: got %v, want GeographyLand", h, got)
			}
		}
	})
	t.Run("hydro_5_rolls_1D", func(t *testing.T) {
		oceanCases := []int{1, 2, 3}
		for _, d := range oceanCases {
			r := roller.NewScripted(d)
			got, err := DetermineFundamentalGeography(r, 5)
			if err != nil {
				t.Fatalf("hydro=5 1D=%d: %v", d, err)
			}
			if got != GeographyOcean {
				t.Errorf("hydro=5 1D=%d: got %v, want GeographyOcean", d, got)
			}
		}
		landCases := []int{4, 5, 6}
		for _, d := range landCases {
			r := roller.NewScripted(d)
			got, err := DetermineFundamentalGeography(r, 5)
			if err != nil {
				t.Fatalf("hydro=5 1D=%d: %v", d, err)
			}
			if got != GeographyLand {
				t.Errorf("hydro=5 1D=%d: got %v, want GeographyLand", d, got)
			}
		}
	})
}

func TestGenerateSurfaceDistribution_NilForMissingHydro(t *testing.T) {
	r := roller.NewScripted() // should not roll
	got, err := GenerateSurfaceDistribution(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("expected nil for nil hydro, got %+v", got)
	}
}

func TestGenerateSurfaceDistribution_ZedPrimeMixed(t *testing.T) {
	hydro := &Hydrographics{Code: 6}
	r := roller.NewScripted(7) // 2D for surface dist
	got, err := GenerateSurfaceDistribution(r, hydro)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil SurfaceDistribution")
	}
	if got.Code != 5 {
		t.Errorf("Code: got %d, want 5", got.Code)
	}
	if got.Description != "Mixed" {
		t.Errorf("Description: got %q, want %q", got.Description, "Mixed")
	}
	if got.Geography != GeographyOcean {
		t.Errorf("Geography: got %v, want GeographyOcean", got.Geography)
	}
}
