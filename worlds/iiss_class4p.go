// Package worlds — IISS Class IV-P Survey Form rendering per WBH pp.138-146
// (sub-project 3B-final).
package worlds

import (
	"fmt"
	"strings"

	"wbh/stars"
)

// RenderIISSClass4P renders the WBH p.138 IISS Class IV Survey form
// (FORM 0407F-IV PART P) for a single body. Plain-text output with
// section headers matching the book's form layout.
//
// mainworldDesignation is the SystemDetail.MainworldDesignation; used
// only to mark whether THIS body is the mainworld in the Comments section.
//
// Returns "" if body is nil or body.Body == BodyEmpty.
//
// Belt bodies (Size 0) dispatch to renderIISS4PBelt for Form 0407K-IV
// PART P.B (WBH p.139).
func RenderIISSClass4P(body *DetailedPlacement, sys stars.System, mainworldDesignation string) string {
	if body == nil || body.Body == BodyEmpty {
		return ""
	}
	// Belt path — Form 0407K-IV PART P.B.
	if body.SizeCode == "0" {
		return renderIISS4PBelt(body, sys, mainworldDesignation)
	}

	var sb strings.Builder
	sb.WriteString("IISS CLASS IV SURVEY — FORM 0407F-IV PART P\n\n")
	renderIISS4PWorld(&sb, body, sys)
	renderIISS4POrbit(&sb, body)
	renderIISS4PSize(&sb, body)
	renderIISS4PAtmosphere(&sb, body)
	renderIISS4PHydrographics(&sb, body)
	renderIISS4PRotation(&sb, body)
	renderIISS4PTemperature(&sb, body)
	renderIISS4PSeismic(&sb, body)
	renderIISS4PLife(&sb, body)
	renderIISS4PResources(&sb, body)
	renderIISS4PHabitability(&sb, body)
	renderIISS4PSubordinates(&sb, body)
	renderIISS4PComments(&sb, body, mainworldDesignation)

	return sb.String()
}

func renderIISS4PWorld(sb *strings.Builder, body *DetailedPlacement, sys stars.System) {
	fmt.Fprintf(sb, "WORLD: %s\n", body.Designation)
	fmt.Fprintf(sb, "SECTOR | LOCATION:    INITIAL SURVEY:    LAST UPDATED:\n")
	fmt.Fprintf(sb, "PRIMARY OBJECT(S):    SYSTEM AGE (Gyr): %.3f    TRAVEL ZONE:\n\n",
		sys.Primary.AgeGyr)
}

func renderIISS4POrbit(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("ORBIT\n")
	fmt.Fprintf(sb, "  AU: %.2f   Eccentricity: %.2f   Period (h): %.2f\n\n",
		stars.OrbitToAU(body.Orbit), body.Eccentricity, body.Period.Hours)
}

func renderIISS4PSize(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("SIZE\n")
	density := 0.0
	gravity := 0.0
	if body.Physical != nil {
		density = body.Physical.Density
		gravity = body.Physical.Gravity
	}
	fmt.Fprintf(sb, "  Diameter (km): %.0f   Density: %.2f   Gravity: %.2f   Mass: %.2f\n\n",
		body.DiameterKm, density, gravity, body.MassEarth)
}

func renderIISS4PAtmosphere(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("ATMOSPHERE\n")
	if body.Atmosphere == nil {
		sb.WriteString("  (none — vacuum)\n\n")
		return
	}
	atm := body.Atmosphere
	fmt.Fprintf(sb, "  Code: %d   Pressure (bar): %.3f   O2 (bar): %.3f   Scale Height: %.2f\n\n",
		atm.Code, atm.Pressure, atm.OxygenPartialPressure, atm.ScaleHeight)
}

func renderIISS4PHydrographics(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("HYDROGRAPHICS\n")
	if body.Hydrographics == nil {
		sb.WriteString("  (none)\n\n")
		return
	}
	hydro := body.Hydrographics
	fmt.Fprintf(sb, "  Code: %d   Coverage (%%): %d   Profile: %s\n\n",
		hydro.Code, hydro.Percent, hydro.Profile)
}

func renderIISS4PRotation(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("ROTATION\n")
	if body.DayLength != nil {
		fmt.Fprintf(sb, "  Sidereal (h): %.2f   Solar (h): %.2f   Solar days/year: %.2f\n",
			body.DayLength.SiderealHours, body.DayLength.SolarHours, body.DayLength.YearDays)
	}
	if body.AxialTilt != nil {
		fmt.Fprintf(sb, "  Axial Tilt: %.2f°\n", body.AxialTilt.Degrees)
	}
	tidalLockText := "no"
	if body.TidalLock != nil && body.TidalLock.LockRatio != "" {
		tidalLockText = body.TidalLock.LockRatio
	}
	tidesM := 0.0
	if body.TidalEffects != nil {
		tidesM = body.TidalEffects.Total
	}
	fmt.Fprintf(sb, "  Tidal lock: %s   Tides (m): %.2f\n\n", tidalLockText, tidesM)
}

func renderIISS4PTemperature(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("TEMPERATURE\n")
	if body.Temperature == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	t := body.Temperature
	// LowK == 0 is MeanTemperatureK's degenerate-model sentinel (lowL ≤ 0
	// when LuminosityModifier hits its 1.0 clamp). Render as "—" so users
	// don't read a literal "0 K" claim. See issue #1.
	if t.LowK <= 0 && t.MeanK > 0 {
		fmt.Fprintf(sb, "  High (K): %.1f   Mean (K): %.1f   Low (K): —\n",
			t.HighK, t.MeanK)
	} else {
		fmt.Fprintf(sb, "  High (K): %.1f   Mean (K): %.1f   Low (K): %.1f\n",
			t.HighK, t.MeanK, t.LowK)
	}
	fmt.Fprintf(sb, "  Luminosity: %.3f   Albedo: %.2f   Greenhouse: %.2f\n\n",
		t.Luminosity, t.Albedo, t.GreenhouseFactor)
}

func renderIISS4PSeismic(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("SEISMIC\n")
	if body.Geology == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	g := body.Geology
	fmt.Fprintf(sb, "  TSS: %d   Residual: %d   TidalStress: %d   TidalHeating: %d   Plates: %d\n\n",
		g.TotalSeismicStress, g.ResidualSeismicStress, g.TidalStressFactor, g.TidalHeatingFactor, g.TectonicPlates)
}

func renderIISS4PLife(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("LIFE\n")
	if body.Biology == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	b := body.Biology
	sophontStr := "no"
	if b.HasNativeSophont {
		sophontStr = "yes"
	} else if b.HadExtinctSophont {
		sophontStr = "extinct"
	}
	fmt.Fprintf(sb, "  Biomass: %s   Biocomplexity: %d   Sophonts?: %s   Biodiversity: %d   Compatibility: %d\n\n",
		string(eHexDigit(b.Biomass)), b.Biocomplexity, sophontStr, b.Biodiversity, b.Compatibility)
}

func renderIISS4PResources(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("RESOURCES\n")
	if body.Biology == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	fmt.Fprintf(sb, "  Rating: %s\n\n", string(eHexDigit(body.Biology.ResourceRating)))
}

func renderIISS4PHabitability(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("HABITABILITY\n")
	if body.Habitability == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	fmt.Fprintf(sb, "  Rating: %d\n\n", body.Habitability.Rating)
}

func renderIISS4PSubordinates(sb *strings.Builder, body *DetailedPlacement) {
	if len(body.Moons) == 0 {
		return
	}
	sb.WriteString("SUBORDINATES\n")
	sb.WriteString("  Designation   SizeCode   Diameter (km)   Orbit (km)   Ecc   Period (h)\n")
	for _, m := range body.Moons {
		fmt.Fprintf(sb, "  %s   %s   %.0f   %d   %.3f   %.2f\n",
			m.Designation, m.SizeCode, m.DiameterKm, m.OrbitKm, m.Eccentricity, m.PeriodHours)
	}
	sb.WriteString("\n")
}

func renderIISS4PComments(sb *strings.Builder, body *DetailedPlacement, mainworldDesignation string) {
	sb.WriteString("COMMENTS\n")
	if mainworldDesignation != "" && body.Designation == mainworldDesignation {
		sb.WriteString("  This is the system mainworld.\n")
	}
}
