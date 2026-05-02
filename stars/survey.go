package stars

import (
	"fmt"
	"strings"
)

// SurveyComponent is one row of the IISS Class 0/I Survey form's Stars
// table (WBH p.33). For composite barycentre rows (e.g. "Aab", "AB",
// "Cab", "ABC"), Class is "—", and Temperature/Diameter are 0; only
// Mass, Luminosity, and the orbital fields apply.
type SurveyComponent struct {
	Component    string
	Class        string // e.g. "G7 V", "D", or "—" for composites
	Mass         float64
	Temperature  float64 // 0 for composites
	Diameter     float64 // 0 for composites
	Luminosity   float64
	Orbit        float64
	AU           float64
	Eccentricity float64
	PeriodYears  float64
	HZCO         float64 // deferred to Plan 3 (orbits chapter); always 0 in Plan 2
}

// SurveyMetadata holds the form-header fields not derivable from the System.
type SurveyMetadata struct {
	Sector, Location, Designation string
	InitialSurvey, LastUpdated    string
}

// SurveyForm is the WBH p.33 Form 0421B-0I structure.
type SurveyForm struct {
	Sector        string
	Location      string
	IISSDesig     string
	InitialSurvey string
	LastUpdated   string
	SystemAgeGyr  float64
	StellarCount  int
	ClassI        bool
	Stars         []SurveyComponent
	Notes         string
	Comments      string
}

// BuildSurveyForm assembles a SurveyForm from a System. Stars rows are
// emitted in book order: per-orbit-class blocks, with the orbit class
// star and its OrbitCompanion-class child grouped together followed by
// their barycentre composite row, then the running outer-system composite.
//
// For Zed (Aa+Ab Companion → Aab; B alone; Ca+Cb → Cab) the rows are:
//
//	Aa, Ab, Aab (A), B, AB, Ca, Cb, Cab (C), ABC
//
// The orbital fields (Orbit, AU, Eccentricity, PeriodYears) on a
// barycentre composite row are inherited from the OUTER orbit class
// (e.g. Cab inherits from Cb's parent Ca's orbit; AB inherits from B's
// orbit; ABC inherits from the outermost present orbit class).
func BuildSurveyForm(sys System, meta SurveyMetadata) SurveyForm {
	form := SurveyForm{
		Sector:        meta.Sector,
		Location:      meta.Location,
		IISSDesig:     meta.Designation,
		InitialSurvey: meta.InitialSurvey,
		LastUpdated:   meta.LastUpdated,
		SystemAgeGyr:  sys.AgeGyr,
		StellarCount:  1 + len(sys.Companions),
	}

	primaryDesig := sys.PrimaryDesignation
	if primaryDesig == "" {
		primaryDesig = "A"
	}

	// Find primary's OrbitCompanion child (Ab).
	primaryCompanionIdx := -1
	for i, c := range sys.Companions {
		if c.ParentIndex == -1 && c.OrbitClass == OrbitCompanion {
			primaryCompanionIdx = i
			break
		}
	}

	// Emit the primary row.
	form.Stars = append(form.Stars, componentFromStar(primaryDesig, sys.Primary))

	if primaryCompanionIdx >= 0 {
		comp := sys.Companions[primaryCompanionIdx]
		form.Stars = append(form.Stars, componentFromCompanion(comp))
		// Aab composite row.
		bary := buildBarycentre(
			compositeName(primaryDesig, comp.Designation),
			[]float64{sys.Primary.Mass, comp.Star.Mass},
			[]float64{sys.Primary.Luminosity, comp.Star.Luminosity},
			comp.OrbitNumber,
			comp.AU,
			comp.Eccentricity,
			comp.PeriodYears,
		)
		// "Aab (A)" parenthetical: simplified outer designation = "A".
		bary.Component = bary.Component + " (A)"
		form.Stars = append(form.Stars, bary)
	}

	// Track the running outer-composite mass and luminosity (starts as
	// primary + its companion if present).
	runningMass := sys.Primary.Mass
	runningLum := sys.Primary.Luminosity
	if primaryCompanionIdx >= 0 {
		runningMass += sys.Companions[primaryCompanionIdx].Star.Mass
		runningLum += sys.Companions[primaryCompanionIdx].Star.Luminosity
	}
	// runningDesig is the current outermost composite label ("A", "AB", "ABC").
	runningDesig := outerLetter(primaryDesig)

	// Walk Close, Near, Far in order. For each present orbit class, emit
	// the star row, then (if it has a companion) the companion row and
	// the inner-pair composite ("Cab"), then advance the outer composite
	// ("AB", "ABC").
	for _, oc := range []OrbitClass{OrbitClose, OrbitNear, OrbitFar} {
		mainIdx := -1
		for i, c := range sys.Companions {
			if c.ParentIndex == -1 && c.OrbitClass == oc {
				mainIdx = i
				break
			}
		}
		if mainIdx < 0 {
			continue
		}
		main := sys.Companions[mainIdx]
		form.Stars = append(form.Stars, componentFromCompanion(main))

		// Companion of this orbit-class star?
		innerIdx := -1
		for i, c := range sys.Companions {
			if c.ParentIndex == mainIdx && c.OrbitClass == OrbitCompanion {
				innerIdx = i
				break
			}
		}

		var addedMass, addedLum float64
		if innerIdx >= 0 {
			inner := sys.Companions[innerIdx]
			form.Stars = append(form.Stars, componentFromCompanion(inner))
			// Inner-pair composite (e.g. Cab).
			pair := buildBarycentre(
				compositeName(main.Designation, inner.Designation),
				[]float64{main.Star.Mass, inner.Star.Mass},
				[]float64{main.Star.Luminosity, inner.Star.Luminosity},
				inner.OrbitNumber,
				inner.AU,
				inner.Eccentricity,
				inner.PeriodYears,
			)
			// Append simplified parenthetical: e.g. "Cab (C)".
			pair.Component = pair.Component + " (" + outerLetter(main.Designation) + ")"
			form.Stars = append(form.Stars, pair)

			addedMass = main.Star.Mass + inner.Star.Mass
			addedLum = main.Star.Luminosity + inner.Star.Luminosity
		} else {
			addedMass = main.Star.Mass
			addedLum = main.Star.Luminosity
		}

		runningMass += addedMass
		runningLum += addedLum
		runningDesig = runningDesig + outerLetter(main.Designation)

		// Outer composite for the running aggregate (e.g. AB, ABC).
		outer := SurveyComponent{
			Component:    runningDesig,
			Class:        "—",
			Mass:         runningMass,
			Luminosity:   runningLum,
			Orbit:        main.OrbitNumber,
			AU:           main.AU,
			Eccentricity: main.Eccentricity,
			PeriodYears:  main.PeriodYears,
		}
		form.Stars = append(form.Stars, outer)
	}

	return form
}

func componentFromStar(designation string, s Star) SurveyComponent {
	return SurveyComponent{
		Component:   designation,
		Class:       formatClass(s),
		Mass:        s.Mass,
		Temperature: s.Temperature,
		Diameter:    s.Diameter,
		Luminosity:  s.Luminosity,
	}
}

func componentFromCompanion(c CompanionStar) SurveyComponent {
	return SurveyComponent{
		Component:    c.Designation,
		Class:        formatClass(c.Star),
		Mass:         c.Star.Mass,
		Temperature:  c.Star.Temperature,
		Diameter:     c.Star.Diameter,
		Luminosity:   c.Star.Luminosity,
		Orbit:        c.OrbitNumber,
		AU:           c.AU,
		Eccentricity: c.Eccentricity,
		PeriodYears:  c.PeriodYears,
	}
}

func buildBarycentre(name string, masses, lums []float64, orbit, au, ecc, period float64) SurveyComponent {
	totalMass := 0.0
	totalLum := 0.0
	for _, m := range masses {
		totalMass += m
	}
	for _, l := range lums {
		totalLum += l
	}
	return SurveyComponent{
		Component:    name,
		Class:        "—",
		Mass:         totalMass,
		Luminosity:   totalLum,
		Orbit:        orbit,
		AU:           au,
		Eccentricity: ecc,
		PeriodYears:  period,
	}
}

// formatClass renders e.g. "G7 V" for stellar entries, "D" for white dwarf,
// "BD" for brown dwarf.
func formatClass(s Star) string {
	switch s.Kind {
	case KindWhiteDwarf:
		return "D"
	case KindBrownDwarf:
		return "BD"
	case KindNeutronStar:
		return "NS"
	case KindBlackHole:
		return "BH"
	case KindPulsar:
		return "PSR"
	case KindNebula:
		return "Nebula"
	case KindProtostar:
		return "Protostar"
	case KindStarCluster:
		return "Cluster"
	case KindAnomaly:
		return "Anomaly"
	}
	if s.SpectralType.Letter == 0 {
		return "—"
	}
	if s.LuminosityClass == "" {
		return s.SpectralType.String()
	}
	return s.SpectralType.String() + " " + string(s.LuminosityClass)
}

// compositeName returns "Aab" given "Aa" and "Ab", or "Cab" given "Ca"
// and "Cb". Combines the first character + concatenation of the lower
// suffixes.
func compositeName(parent, child string) string {
	if len(parent) < 2 || len(child) < 2 {
		return parent + child
	}
	return parent + string(child[1])
}

// outerLetter returns the uppercase letter of a designation ("Aa"->"A", "B"->"B").
func outerLetter(d string) string {
	if d == "" {
		return ""
	}
	return string(d[0])
}

// ShortProfile renders the WBH p.31 shorthand for a System.
// Format: "#-T# C-M-D-L-A" for single star;
// multi-star appends ":<desig>-<orbit#>-<ecc>-T# C-M-D-L" per companion.
func ShortProfile(sys System) string {
	var b strings.Builder
	count := 1 + len(sys.Companions)
	fmt.Fprintf(&b, "%d-%s-%.3f-%.3f-%.3f-%.3f",
		count, formatClass(sys.Primary),
		sys.Primary.Mass, sys.Primary.Diameter, sys.Primary.Luminosity, sys.AgeGyr)
	for _, c := range sys.Companions {
		fmt.Fprintf(&b, ":%s-%.2f-%.2f-%s-%.3f-%.3f-%.3f",
			c.Designation, c.OrbitNumber, c.Eccentricity,
			formatClass(c.Star),
			c.Star.Mass, c.Star.Diameter, c.Star.Luminosity)
	}
	return b.String()
}
