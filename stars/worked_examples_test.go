package stars_test

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
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
	// HZCO on solo primary row (WBH p.35 — Sol/Terra survey form).
	if math.Abs(row.HZCO-3.0) > 0.05 {
		t.Errorf("HZCO = %v want 3.0±0.05", row.HZCO)
	}
}

func TestZed_SurveyForm_p34(t *testing.T) {
	// WBH p.34 — Zed quintuple system. The book gives only final survey-form
	// values; we Compose each star directly and verify BuildSurveyForm
	// produces the 9-row layout with correct cumulative-barycentre masses
	// and luminosities.
	aa := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass:            0.929,
		Diameter:        0.967,
		Temperature:     5440,
		AgeGyr:          6.3,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.907,
		Diameter:        0.957,
		Temperature:     5360,
		AgeGyr:          6.3,
	})
	bStar := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.626,
		Diameter:        0.777,
		Temperature:     3980,
		AgeGyr:          6.3,
	})
	ca := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'M', Subtype: 0},
		LuminosityClass: stars.V,
		Mass:            0.510,
		Diameter:        0.728,
		Temperature:     3700,
		AgeGyr:          6.3,
	})
	// Cb is a white dwarf. Compose still computes luminosity from D and T;
	// the book reports 0.000525, so set D and T to match.
	cb := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindWhiteDwarf,
		LuminosityClass: stars.D,
		// White dwarfs don't have a SpectralType in the conventional sense.
		Mass:        0.490,
		Diameter:    0.017,
		Temperature: 6700,
		AgeGyr:      4.635, // book p.30
	})

	sys := stars.System{
		Primary: aa,
		AgeGyr:  6.3,
		Companions: []stars.CompanionStar{
			// 0: Ab — primary's companion (orbit class Companion)
			{
				Star: ab, OrbitClass: stars.OrbitCompanion, ParentIndex: -1,
				OrbitNumber: 0.09, AU: 0.036, Eccentricity: 0.11,
				PeriodYears: stars.OrbitPeriodYears(0.036, aa.Mass, ab.Mass),
			},
			// 1: B — Near (alone, no companion)
			{
				Star: bStar, OrbitClass: stars.OrbitNear, ParentIndex: -1,
				OrbitNumber: 6.1, AU: 5.68, Eccentricity: 0.08,
				PeriodYears: stars.OrbitPeriodYears(5.68, aa.Mass+ab.Mass, bStar.Mass),
			},
			// 2: Ca — Far (with companion)
			{
				Star: ca, OrbitClass: stars.OrbitFar, ParentIndex: -1,
				OrbitNumber: 12.1, AU: 338, Eccentricity: 0.47,
				PeriodYears: stars.OrbitPeriodYears(338, aa.Mass+ab.Mass+bStar.Mass, ca.Mass),
			},
			// 3: Cb — Ca's companion (orbit class Companion, parent = index 2)
			{
				Star: cb, OrbitClass: stars.OrbitCompanion, ParentIndex: 2,
				OrbitNumber: 0.21, AU: 0.084, Eccentricity: 0.24,
				PeriodYears: stars.OrbitPeriodYears(0.084, ca.Mass, cb.Mass),
			},
		},
	}
	stars.AssignDesignations(&sys)

	// Sanity-check designations.
	if sys.PrimaryDesignation != "Aa" {
		t.Errorf("primary: got %q want Aa", sys.PrimaryDesignation)
	}
	wantDesignations := []string{"Ab", "B", "Ca", "Cb"}
	for i, w := range wantDesignations {
		if sys.Companions[i].Designation != w {
			t.Errorf("companion[%d]: got %q want %q", i, sys.Companions[i].Designation, w)
		}
	}

	form := stars.BuildSurveyForm(sys, stars.SurveyMetadata{
		Sector: "Storr", Location: "0602",
		Designation:   "Zed",
		InitialSurvey: "207-568", LastUpdated: "218-1061",
	})

	if form.StellarCount != 5 {
		t.Errorf("StellarCount = %d want 5", form.StellarCount)
	}
	if math.Abs(form.SystemAgeGyr-6.3) > 1e-9 {
		t.Errorf("SystemAgeGyr = %v want 6.3", form.SystemAgeGyr)
	}

	wantComponents := []string{"Aa", "Ab", "Aab (A)", "B", "AB", "Ca", "Cb", "Cab (C)", "ABC"}
	if len(form.Stars) != len(wantComponents) {
		t.Fatalf("rows = %d want %d (%v)", len(form.Stars), len(wantComponents), form.Stars)
	}
	for i, w := range wantComponents {
		if form.Stars[i].Component != w {
			t.Errorf("[%d].Component = %q want %q", i, form.Stars[i].Component, w)
		}
	}

	// Per-row numeric assertions.
	type rowExpect struct {
		idx                                                int
		mass, temp, diameter, lum, orbit, au, ecc, period  float64
		massTol, tempTol, diamTol, lumTol, orbitTol, auTol float64
		eccTol, periodTol                                  float64
	}
	expected := []rowExpect{
		// Aa
		{
			idx: 0, mass: 0.929, temp: 5440, diameter: 0.967, lum: 0.738,
			massTol: 1e-9, tempTol: 1e-9, diamTol: 1e-9, lumTol: 5e-3,
		},
		// Ab
		{
			idx: 1, mass: 0.907, temp: 5360, diameter: 0.957, lum: 0.681,
			orbit: 0.09, au: 0.036, ecc: 0.11, period: 0.005,
			massTol: 1e-9, tempTol: 1e-9, diamTol: 1e-9, lumTol: 5e-3,
			orbitTol: 1e-9, auTol: 1e-9, eccTol: 1e-9, periodTol: 5e-4,
		},
		// Aab (A) — composite of Aa + Ab
		{
			idx: 2, mass: 1.836, lum: 0.738 + 0.681,
			orbit: 0.09, au: 0.036, ecc: 0.11, period: 0.005,
			massTol: 1e-9, lumTol: 5e-3,
			orbitTol: 1e-9, auTol: 1e-9, eccTol: 1e-9, periodTol: 5e-4,
		},
		// B (K8 V Near)
		{
			idx: 3, mass: 0.626, temp: 3980, diameter: 0.777, lum: 0.136,
			orbit: 6.1, au: 5.68, ecc: 0.08, period: 8.627,
			massTol: 1e-9, tempTol: 1e-9, diamTol: 1e-9, lumTol: 5e-2,
			orbitTol: 1e-9, auTol: 1e-9, eccTol: 1e-9, periodTol: 5e-2,
		},
		// AB — composite of Aa + Ab + B
		{
			idx: 4, mass: 2.462, lum: 0.738 + 0.681 + 0.136,
			orbit: 6.1, au: 5.68, ecc: 0.08, period: 8.627,
			massTol: 1e-9, lumTol: 5e-2,
			orbitTol: 1e-9, auTol: 1e-9, eccTol: 1e-9, periodTol: 5e-2,
		},
		// Ca (M0 V Far)
		{
			idx: 5, mass: 0.510, temp: 3700, diameter: 0.728, lum: 0.0895,
			orbit: 12.1, au: 338, ecc: 0.47, period: 3598,
			massTol: 1e-9, tempTol: 1e-9, diamTol: 1e-9, lumTol: 5e-2,
			orbitTol: 1e-9, auTol: 1, eccTol: 1e-9, periodTol: 50,
		},
		// Cb (white dwarf)
		{
			idx: 6, mass: 0.490, temp: 6700, diameter: 0.017, lum: 0.000525,
			orbit: 0.21, au: 0.084, ecc: 0.24, period: 0.024,
			massTol: 1e-9, tempTol: 1e-9, diamTol: 1e-9, lumTol: 5e-4,
			orbitTol: 1e-9, auTol: 1e-9, eccTol: 1e-9, periodTol: 5e-3,
		},
		// Cab (C) — composite of Ca + Cb
		{
			idx: 7, mass: 0.510 + 0.490, lum: 0.0895 + 0.000525,
			orbit: 0.21, au: 0.084, ecc: 0.24, period: 0.024,
			massTol: 1e-9, lumTol: 5e-2,
			orbitTol: 1e-9, auTol: 1e-9, eccTol: 1e-9, periodTol: 5e-3,
		},
		// ABC — outer composite of Aa + Ab + B + Ca + Cb
		{
			idx: 8, mass: 0.929 + 0.907 + 0.626 + 0.510 + 0.490,
			lum:   0.738 + 0.681 + 0.136 + 0.0895 + 0.000525,
			orbit: 12.1, au: 338, ecc: 0.47, period: 3598,
			massTol: 1e-9, lumTol: 5e-2,
			orbitTol: 1e-9, auTol: 1, eccTol: 1e-9, periodTol: 50,
		},
	}
	for _, exp := range expected {
		row := form.Stars[exp.idx]
		check := func(name string, got, want, tol float64) {
			if tol == 0 {
				return // unset = skip
			}
			if math.Abs(got-want) > tol {
				t.Errorf("[%d:%s].%s = %v, want %v (tol %v)", exp.idx, row.Component, name, got, want, tol)
			}
		}
		check("Mass", row.Mass, exp.mass, exp.massTol)
		check("Temperature", row.Temperature, exp.temp, exp.tempTol)
		check("Diameter", row.Diameter, exp.diameter, exp.diamTol)
		check("Luminosity", row.Luminosity, exp.lum, exp.lumTol)
		check("Orbit", row.Orbit, exp.orbit, exp.orbitTol)
		check("AU", row.AU, exp.au, exp.auTol)
		check("Eccentricity", row.Eccentricity, exp.ecc, exp.eccTol)
		check("Period", row.PeriodYears, exp.period, exp.periodTol)
	}

	// HZCO assertions per WBH p.34 Zed survey form. Per the book, HZCO
	// is published only on rows that act as a single HZCO source:
	//   Aab (A) → 3.3 (Aa+Ab combined luminosity)
	//   B       → 0.92 (K8 V solo)
	//   Cab (C) → 0.75 (Ca+Cb combined luminosity)
	// Pair-member rows (Aa, Ab, Ca, Cb) and outer running composites
	// (AB, ABC) leave HZCO=0 in the book.
	hzcoChecks := []struct {
		idx  int
		want float64
		tol  float64
	}{
		{idx: 0, want: 0, tol: 1e-9},    // Aa
		{idx: 1, want: 0, tol: 1e-9},    // Ab
		{idx: 2, want: 3.3, tol: 0.05},  // Aab (A)
		{idx: 3, want: 0.92, tol: 0.05}, // B
		{idx: 4, want: 0, tol: 1e-9},    // AB
		{idx: 5, want: 0, tol: 1e-9},    // Ca
		{idx: 6, want: 0, tol: 1e-9},    // Cb
		{idx: 7, want: 0.75, tol: 0.05}, // Cab (C)
		{idx: 8, want: 0, tol: 1e-9},    // ABC
	}
	for _, c := range hzcoChecks {
		row := form.Stars[c.idx]
		if math.Abs(row.HZCO-c.want) > c.tol {
			t.Errorf("[%d:%s].HZCO = %v, want %v (tol %v)", c.idx, row.Component, row.HZCO, c.want, c.tol)
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
	// HZCO on the Aab composite row (WBH p.62 Corella — combined L 1.725 → 3.5).
	if math.Abs(row2.HZCO-3.5) > 0.05 {
		t.Errorf("[2].HZCO = %v want 3.5±0.05", row2.HZCO)
	}
}
