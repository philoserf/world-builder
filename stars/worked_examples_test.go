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

func TestSolTerra_SurveyForm_p35(t *testing.T) {
	// WBH p.35 — Terra/Sol IISS Class 0/I Survey form. Sol is fully
	// specified; we Compose it directly and run BuildSurveyForm.
	sol := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.000,
		Diameter:        1.000,
		Temperature:     5772,
		AgeGyr:          4.568,
	})
	sys := stars.System{
		Primary:            sol,
		PrimaryDesignation: "A",
		AgeGyr:             4.568,
	}
	form := stars.BuildSurveyForm(sys, stars.SurveyMetadata{
		Sector:        "Solomani Rim",
		Location:      "1827",
		Designation:   "Terra",
		InitialSurvey: "001-(-2500)",
		LastUpdated:   "001-(-2498)",
	})

	// Header fields.
	if form.Sector != "Solomani Rim" {
		t.Errorf("Sector = %q want %q", form.Sector, "Solomani Rim")
	}
	if form.Location != "1827" {
		t.Errorf("Location = %q want %q", form.Location, "1827")
	}
	if form.IISSDesig != "Terra" {
		t.Errorf("IISSDesig = %q want %q", form.IISSDesig, "Terra")
	}
	if form.InitialSurvey != "001-(-2500)" {
		t.Errorf("InitialSurvey = %q", form.InitialSurvey)
	}
	if form.LastUpdated != "001-(-2498)" {
		t.Errorf("LastUpdated = %q", form.LastUpdated)
	}
	if math.Abs(form.SystemAgeGyr-4.568) > 1e-9 {
		t.Errorf("SystemAgeGyr = %v want 4.568", form.SystemAgeGyr)
	}
	if form.StellarCount != 1 {
		t.Errorf("StellarCount = %d want 1", form.StellarCount)
	}

	// Stars table: should have exactly one row.
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
	checks := []struct {
		name string
		got  float64
		want float64
	}{
		{"Mass", row.Mass, 1.000},
		{"Temperature", row.Temperature, 5772},
		{"Diameter", row.Diameter, 1.000},
		{"Luminosity", row.Luminosity, 1.000},
	}
	for _, c := range checks {
		if math.Abs(c.got-c.want) > 1e-9 {
			t.Errorf("%s = %v want %v", c.name, c.got, c.want)
		}
	}
}

func TestCorella_SurveyForm_p35(t *testing.T) {
	// WBH p.35 Corella binary: G2 V + G8 V (Companion-class).
	// Constructed directly via Compose; the book's roll sequence isn't
	// specified for Corella so we don't drive GenerateSystem here.

	a := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.224,
		Diameter:        0.998,
		Temperature:     5840,
		AgeGyr:          4.9,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.974,
		Diameter:        0.957,
		Temperature:     5360,
		AgeGyr:          4.9,
	})

	// Per WBH p.30 Kepler formula: P = sqrt(au^3 / (M+m)).
	au := 0.120
	totalMass := a.Mass + b.Mass
	period := stars.OrbitPeriodYears(au, a.Mass, b.Mass)

	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{
				Star:         b,
				OrbitClass:   stars.OrbitCompanion,
				ParentIndex:  -1,
				OrbitNumber:  0.30,
				AU:           au,
				Eccentricity: 0.010,
				PeriodYears:  period,
			},
		},
		AgeGyr: 4.9,
	}
	stars.AssignDesignations(&sys)

	form := stars.BuildSurveyForm(sys, stars.SurveyMetadata{
		Sector:      "The Beyond",
		Location:    "0314",
		Designation: "Corella",
	})

	if form.StellarCount != 2 {
		t.Fatalf("StellarCount = %d want 2", form.StellarCount)
	}

	// Expect 3 rows: Aa (primary), Ab (companion), Aab (A) (composite).
	if len(form.Stars) != 3 {
		t.Fatalf("Stars rows = %d want 3 (got: %v)", len(form.Stars), form.Stars)
	}

	// Row 0: primary Aa.
	row0 := form.Stars[0]
	if row0.Component != "Aa" {
		t.Errorf("[0].Component = %q want Aa", row0.Component)
	}
	if row0.Class != "G2 V" {
		t.Errorf("[0].Class = %q want G2 V", row0.Class)
	}
	primaryChecks := []struct {
		name      string
		got, want float64
	}{
		{"Mass", row0.Mass, 1.224},
		{"Temperature", row0.Temperature, 5840},
		{"Diameter", row0.Diameter, 0.998},
		{"Luminosity", row0.Luminosity, 1.045},
	}
	for _, c := range primaryChecks {
		// Luminosity is computed from the formula; allow small rel-tol.
		tol := 1e-9
		if c.name == "Luminosity" {
			tol = 5e-3
		}
		if math.Abs(c.got-c.want) > tol {
			t.Errorf("[0].%s = %v want %v (tol %v)", c.name, c.got, c.want, tol)
		}
	}

	// Row 1: companion Ab.
	row1 := form.Stars[1]
	if row1.Component != "Ab" {
		t.Errorf("[1].Component = %q want Ab", row1.Component)
	}
	if row1.Class != "G8 V" {
		t.Errorf("[1].Class = %q want G8 V", row1.Class)
	}
	companionChecks := []struct {
		name           string
		got, want, tol float64
	}{
		{"Mass", row1.Mass, 0.974, 1e-9},
		{"Temperature", row1.Temperature, 5360, 1e-9},
		{"Diameter", row1.Diameter, 0.957, 1e-9},
		{"Luminosity", row1.Luminosity, 0.681, 5e-3},
		{"Orbit", row1.Orbit, 0.30, 1e-9},
		{"AU", row1.AU, 0.120, 1e-3},
		{"Eccentricity", row1.Eccentricity, 0.010, 1e-9},
		{"Period", row1.PeriodYears, 0.028, 1e-3},
	}
	for _, c := range companionChecks {
		if math.Abs(c.got-c.want) > c.tol {
			t.Errorf("[1].%s = %v want %v (tol %v)", c.name, c.got, c.want, c.tol)
		}
	}

	// Row 2: composite Aab (A).
	row2 := form.Stars[2]
	if row2.Component != "Aab (A)" {
		t.Errorf("[2].Component = %q want Aab (A)", row2.Component)
	}
	if row2.Class != "—" {
		t.Errorf("[2].Class = %q want —", row2.Class)
	}
	if math.Abs(row2.Mass-totalMass) > 1e-9 {
		t.Errorf("[2].Mass = %v want %v", row2.Mass, totalMass)
	}
	// Composite luminosity = primary + companion luminosity (book reports 1.725).
	wantLum := row0.Luminosity + row1.Luminosity
	if math.Abs(row2.Luminosity-wantLum) > 1e-9 {
		t.Errorf("[2].Luminosity = %v want %v (sum)", row2.Luminosity, wantLum)
	}
	if math.Abs(row2.PeriodYears-period) > 1e-9 {
		t.Errorf("[2].Period = %v want %v", row2.PeriodYears, period)
	}
}
