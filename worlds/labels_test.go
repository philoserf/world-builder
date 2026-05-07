package worlds

import "testing"

// TestAtmosphereCodeName covers WBH p.79 atmosphere codes 0-H (0-17).
// Spot-checks several codes plus the unknown-code fallback.
func TestAtmosphereCodeName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		code int
		want string
	}{
		{0, "None (Vacuum)"},
		{1, "Trace"},
		{6, "Standard"},
		{10, "Exotic"}, // A
		{11, "Corrosive"},
		{12, "Insidious"},
		{15, "Unusual"}, // F
		{17, "Gas, Hydrogen"},
		{-1, ""}, // out of range
		{99, ""},
	}
	for _, tc := range cases {
		if got := AtmosphereCodeName(tc.code); got != tc.want {
			t.Errorf("AtmosphereCodeName(%d) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

// TestBiocomplexityName covers WBH p.129 biocomplexity ratings.
func TestBiocomplexityName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		c    int
		want string
	}{
		{0, ""}, // no biocomplexity rendered
		{1, "Primitive single-cell organisms"},
		{5, "Complex multicellular organisms"},
		{8, "Mentally advanced organisms"},
		{9, "Extant or extinct sophonts"},
		{10, "Ecosystem-wide superorganisms"},
		{15, "Ecosystem-wide superorganisms"}, // 10+ band
	}
	for _, tc := range cases {
		if got := BiocomplexityName(tc.c); got != tc.want {
			t.Errorf("BiocomplexityName(%d) = %q, want %q", tc.c, got, tc.want)
		}
	}
}

// TestBiomassLabel covers the project-supplied band scheme anchored on
// WBH pp.127-128 prose ("0 = no native life", "A+ = healthy garden world").
func TestBiomassLabel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		b    int
		want string
	}{
		{0, "None (sterile)"},
		{1, "Trace"},
		{2, "Sparse"},
		{4, "Moderate"},
		{6, "Abundant"},
		{8, "Lush"},
		{10, "Garden world"},
		{12, "Garden world"}, // 10+ stays at top band
	}
	for _, tc := range cases {
		if got := BiomassLabel(tc.b); got != tc.want {
			t.Errorf("BiomassLabel(%d) = %q, want %q", tc.b, got, tc.want)
		}
	}
}

// TestBiodiversityLabel covers anchored bands per WBH p.130.
func TestBiodiversityLabel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		b    int
		want string
	}{
		{0, "None"},
		{1, "Uniform (few dominant species)"},
		{2, "Uniform (few dominant species)"},
		{3, "Limited"},
		{6, "Moderate"},
		{8, "Diverse"},
		{10, "Terra-like complexity"},
	}
	for _, tc := range cases {
		if got := BiodiversityLabel(tc.b); got != tc.want {
			t.Errorf("BiodiversityLabel(%d) = %q, want %q", tc.b, got, tc.want)
		}
	}
}

// TestCompatibilityLabel covers anchored bands per WBH pp.130-131.
func TestCompatibilityLabel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		c    int
		want string
	}{
		{0, "Incompatible"},
		{1, "Marginal"},
		{4, "Partial"},
		{7, "Mostly compatible"},
		{10, "Fully Terran-compatible"},
		{12, "Beyond Terran"},
	}
	for _, tc := range cases {
		if got := CompatibilityLabel(tc.c); got != tc.want {
			t.Errorf("CompatibilityLabel(%d) = %q, want %q", tc.c, got, tc.want)
		}
	}
}

// TestHabitabilityRatingName covers WBH p.133 banded labels.
func TestHabitabilityRatingName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		r    int
		want string
	}{
		{0, "Actively hostile"},
		{1, "Barely habitable"},
		{4, "Marginally survivable"},
		{7, "Regionally habitable"},
		{9, "Suitable"},
		{10, "Garden world"},
		{12, "Garden world"},
	}
	for _, tc := range cases {
		if got := HabitabilityRatingName(tc.r); got != tc.want {
			t.Errorf("HabitabilityRatingName(%d) = %q, want %q", tc.r, got, tc.want)
		}
	}
}

// TestResourceRatingName covers WBH p.131 banded remarks.
func TestResourceRatingName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		r    int
		want string
	}{
		{2, "No economically extractable resources"},
		{4, "Marginal"},
		{7, "Worthwhile"},
		{9, "Priority target"},
		{10, "Priority target"},
		{11, "Resource rush"},
		{12, "Resource rush"},
	}
	for _, tc := range cases {
		if got := ResourceRatingName(tc.r); got != tc.want {
			t.Errorf("ResourceRatingName(%d) = %q, want %q", tc.r, got, tc.want)
		}
	}
}

// TestTotalSeismicStressLabel covers WBH p.126 prose-anchored bands.
func TestTotalSeismicStressLabel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s    int
		want string
	}{
		{0, "Geologically dead"},
		{5, "Quiet"},
		{50, "Active"},
		{100, "Active"},
		{144, "Hyperactive (constant eruptions)"},
		{500, "Hyperactive (constant eruptions)"},
	}
	for _, tc := range cases {
		if got := TotalSeismicStressLabel(tc.s); got != tc.want {
			t.Errorf("TotalSeismicStressLabel(%d) = %q, want %q", tc.s, got, tc.want)
		}
	}
}

// TestTidalHeatingFactorLabel covers WBH p.126 reference-anchored bands.
func TestTidalHeatingFactorLabel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		f    int
		want string
	}{
		{0, "Negligible"},
		{5, "Mild"},
		{11, "Significant (Enceladus-class)"},
		{100, "Significant (Enceladus-class)"},
		{101, "Extreme (Io-class)"},
		{500, "Extreme (Io-class)"},
	}
	for _, tc := range cases {
		if got := TidalHeatingFactorLabel(tc.f); got != tc.want {
			t.Errorf("TidalHeatingFactorLabel(%d) = %q, want %q", tc.f, got, tc.want)
		}
	}
}
