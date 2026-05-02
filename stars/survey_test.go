package stars

import (
	"math"
	"testing"
)

func TestBuildSurveyForm_SinglePrimary(t *testing.T) {
	primary := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: V,
		Mass:            1.000, Diameter: 1.000, Temperature: 5772, AgeGyr: 4.568,
	})
	sys := System{
		Primary:            primary,
		PrimaryDesignation: "A",
		AgeGyr:             4.568,
	}
	form := BuildSurveyForm(sys, SurveyMetadata{
		Sector:      "Solomani Rim",
		Location:    "1827",
		Designation: "Terra",
	})
	if form.StellarCount != 1 {
		t.Fatalf("StellarCount = %d want 1", form.StellarCount)
	}
	if len(form.Stars) != 1 {
		t.Fatalf("Stars rows = %d want 1", len(form.Stars))
	}
	row := form.Stars[0]
	if row.Component != "A" {
		t.Errorf("Component = %q want A", row.Component)
	}
	if row.Class != "G2 V" {
		t.Errorf("Class = %q want G2 V", row.Class)
	}
	if math.Abs(row.Mass-1.0) > 1e-9 {
		t.Errorf("Mass = %v want 1.0", row.Mass)
	}
	if math.Abs(row.Luminosity-1.0) > 1e-9 {
		t.Errorf("Luminosity = %v want 1.0", row.Luminosity)
	}
}

func TestBuildSurveyForm_PrimaryWithCompanion(t *testing.T) {
	primary := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: V,
		Mass:            0.929, Diameter: 0.967, Temperature: 5440, AgeGyr: 6.3,
	})
	companion := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: V,
		Mass:            0.907, Diameter: 0.957, Temperature: 5360, AgeGyr: 6.3,
	})
	sys := System{
		Primary:            primary,
		PrimaryDesignation: "Aa",
		Companions: []CompanionStar{
			{
				Star:         companion,
				Designation:  "Ab",
				OrbitClass:   OrbitCompanion,
				ParentIndex:  -1,
				OrbitNumber:  0.09,
				AU:           0.036,
				Eccentricity: 0.11,
				PeriodYears:  0.005,
			},
		},
		AgeGyr: 6.3,
	}
	form := BuildSurveyForm(sys, SurveyMetadata{})
	// Expect 3 rows: Aa, Ab, Aab (A).
	if len(form.Stars) != 3 {
		t.Fatalf("Stars rows = %d want 3 (got: %v)", len(form.Stars), form.Stars)
	}
	if form.Stars[0].Component != "Aa" {
		t.Errorf("[0].Component = %q want Aa", form.Stars[0].Component)
	}
	if form.Stars[1].Component != "Ab" {
		t.Errorf("[1].Component = %q want Ab", form.Stars[1].Component)
	}
	if form.Stars[2].Component != "Aab (A)" {
		t.Errorf("[2].Component = %q want Aab (A)", form.Stars[2].Component)
	}
	if math.Abs(form.Stars[2].Mass-(0.929+0.907)) > 1e-9 {
		t.Errorf("[2].Mass = %v want 1.836", form.Stars[2].Mass)
	}
}

func TestBuildSurveyForm_Metadata(t *testing.T) {
	primary := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'K', Subtype: 0},
		LuminosityClass: V,
		Mass:            0.78, Diameter: 0.82, Temperature: 5250, AgeGyr: 3.0,
	})
	sys := System{
		Primary:            primary,
		PrimaryDesignation: "A",
		AgeGyr:             3.0,
	}
	meta := SurveyMetadata{
		Sector:        "Spinward Marches",
		Location:      "0304",
		Designation:   "Regina",
		InitialSurvey: "001-993",
		LastUpdated:   "200-1105",
	}
	form := BuildSurveyForm(sys, meta)
	if form.Sector != "Spinward Marches" {
		t.Errorf("Sector = %q want Spinward Marches", form.Sector)
	}
	if form.Location != "0304" {
		t.Errorf("Location = %q want 0304", form.Location)
	}
	if form.IISSDesig != "Regina" {
		t.Errorf("IISSDesig = %q want Regina", form.IISSDesig)
	}
	if form.InitialSurvey != "001-993" {
		t.Errorf("InitialSurvey = %q want 001-993", form.InitialSurvey)
	}
	if form.LastUpdated != "200-1105" {
		t.Errorf("LastUpdated = %q want 200-1105", form.LastUpdated)
	}
	if math.Abs(form.SystemAgeGyr-3.0) > 1e-9 {
		t.Errorf("SystemAgeGyr = %v want 3.0", form.SystemAgeGyr)
	}
}

func TestBuildSurveyForm_BinaryNoCompanions(t *testing.T) {
	// A + B (no companions); expect 3 rows: A, B, AB.
	primary := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'G', Subtype: 5},
		LuminosityClass: V,
		Mass:            0.92, Diameter: 0.95, Temperature: 5510, AgeGyr: 5.0,
	})
	secondary := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'K', Subtype: 2},
		LuminosityClass: V,
		Mass:            0.72, Diameter: 0.76, Temperature: 5100, AgeGyr: 5.0,
	})
	sys := System{
		Primary:            primary,
		PrimaryDesignation: "A",
		Companions: []CompanionStar{
			{
				Star:         secondary,
				Designation:  "B",
				OrbitClass:   OrbitNear,
				ParentIndex:  -1,
				OrbitNumber:  3.5,
				AU:           2.8,
				Eccentricity: 0.20,
				PeriodYears:  4.2,
			},
		},
		AgeGyr: 5.0,
	}
	form := BuildSurveyForm(sys, SurveyMetadata{})
	// Rows: A, B, AB.
	if len(form.Stars) != 3 {
		t.Fatalf("Stars rows = %d want 3 (got: %v)", len(form.Stars), form.Stars)
	}
	if form.Stars[0].Component != "A" {
		t.Errorf("[0].Component = %q want A", form.Stars[0].Component)
	}
	if form.Stars[1].Component != "B" {
		t.Errorf("[1].Component = %q want B", form.Stars[1].Component)
	}
	if form.Stars[2].Component != "AB" {
		t.Errorf("[2].Component = %q want AB", form.Stars[2].Component)
	}
	// AB composite inherits B's orbital fields.
	if math.Abs(form.Stars[2].Orbit-3.5) > 1e-9 {
		t.Errorf("[2].Orbit = %v want 3.5", form.Stars[2].Orbit)
	}
	if math.Abs(form.Stars[2].Mass-(0.92+0.72)) > 1e-9 {
		t.Errorf("[2].Mass = %v want 1.64", form.Stars[2].Mass)
	}
}

func TestShortProfile_SinglePrimary(t *testing.T) {
	primary := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: V,
		Mass:            1.0, Diameter: 1.0, Temperature: 5772, AgeGyr: 4.568,
	})
	sys := System{Primary: primary, PrimaryDesignation: "A", AgeGyr: 4.568}
	got := ShortProfile(sys)
	want := "1-G2 V-1.000-1.000-1.000-4.568"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestShortProfile_Binary(t *testing.T) {
	primary := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: V,
		Mass:            0.929, Diameter: 0.967, Temperature: 5440, AgeGyr: 6.3,
	})
	secondary := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'K', Subtype: 1},
		LuminosityClass: V,
		Mass:            0.82, Diameter: 0.87, Temperature: 5150, AgeGyr: 6.3,
	})
	sys := System{
		Primary:            primary,
		PrimaryDesignation: "A",
		Companions: []CompanionStar{
			{
				Star:         secondary,
				Designation:  "B",
				OrbitClass:   OrbitFar,
				ParentIndex:  -1,
				OrbitNumber:  7.0,
				AU:           40.0,
				Eccentricity: 0.30,
				PeriodYears:  220.0,
			},
		},
		AgeGyr: 6.3,
	}
	got := ShortProfile(sys)
	// Should start with "2-G7 V-..." and contain ":B-7.00-0.30-K1 V-..."
	if len(got) == 0 {
		t.Fatal("empty profile")
	}
	// Check count prefix.
	if got[:2] != "2-" {
		t.Errorf("profile should start with 2-, got %q", got[:10])
	}
}

func TestFormatClass_WhiteDwarf(t *testing.T) {
	s := Star{Kind: KindWhiteDwarf, LuminosityClass: D}
	if got := formatClass(s); got != "D" {
		t.Fatalf("got %q want D", got)
	}
}

func TestFormatClass_BrownDwarf(t *testing.T) {
	s := Star{Kind: KindBrownDwarf, LuminosityClass: BD}
	if got := formatClass(s); got != "BD" {
		t.Fatalf("got %q want BD", got)
	}
}

func TestFormatClass_Stellar(t *testing.T) {
	s := Star{Kind: KindMainSequence, SpectralType: SpectralType{Letter: 'G', Subtype: 7}, LuminosityClass: V}
	if got := formatClass(s); got != "G7 V" {
		t.Fatalf("got %q want G7 V", got)
	}
}

func TestFormatClass_NoSpectralType(t *testing.T) {
	s := Star{Kind: KindMainSequence}
	if got := formatClass(s); got != "—" {
		t.Fatalf("got %q want —", got)
	}
}

func TestCompositeName(t *testing.T) {
	tests := []struct {
		parent, child, want string
	}{
		{"Aa", "Ab", "Aab"},
		{"Ca", "Cb", "Cab"},
		{"A", "B", "AB"}, // short designations fall back to concatenation
	}
	for _, tc := range tests {
		got := compositeName(tc.parent, tc.child)
		if got != tc.want {
			t.Errorf("compositeName(%q, %q) = %q want %q", tc.parent, tc.child, got, tc.want)
		}
	}
}

func TestOuterLetter(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Aa", "A"},
		{"B", "B"},
		{"Ca", "C"},
		{"", ""},
	}
	for _, tc := range tests {
		got := outerLetter(tc.in)
		if got != tc.want {
			t.Errorf("outerLetter(%q) = %q want %q", tc.in, got, tc.want)
		}
	}
}
