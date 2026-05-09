package worlds

import (
	"fmt"
	"strings"

	"wbh/stars"
)

// RenderClass4PMarkdown renders the IISS Class IV-P Survey form for a
// single body (the system mainworld). Dispatches on SizeCode:
//   - "0"  → Form 0407K-IV PART P.B (belt variant, WBH p.139)
//   - else → Form 0407F-IV PART P (terrestrial/moon variant, WBH p.138)
//
// Returns "" if body is nil. mainworldDesignation is used only for the
// "this is the mainworld" marker in the Comments section.
func RenderClass4PMarkdown(body *DetailedPlacement, sys stars.System, mainworldDesignation string) string {
	if body == nil || body.Body == BodyEmpty {
		return ""
	}
	if body.SizeCode == "0" {
		return renderClass4PBeltMarkdown(body, sys, mainworldDesignation)
	}
	return renderClass4PTerrestrialMarkdown(body, sys, mainworldDesignation)
}

func renderClass4PTerrestrialMarkdown(body *DetailedPlacement, sys stars.System, mainworldDesignation string) string {
	var sb strings.Builder
	sb.WriteString("## IISS Class IV-P Survey — Form 0407F-IV PART P\n\n")
	writeClass4PHeader(&sb, body, sys)
	writeClass4POrbit(&sb, body)
	writeClass4PSize(&sb, body)
	writeClass4PAtmosphere(&sb, body)
	writeClass4PHydrographics(&sb, body)
	writeClass4PRotation(&sb, body)
	writeClass4PTemperature(&sb, body)
	writeClass4PSeismic(&sb, body)
	writeClass4PLife(&sb, body)
	writeClass4PResources(&sb, body)
	writeClass4PHabitability(&sb, body)
	writeClass4PSubordinates(&sb, body)
	writeClass4PComments(&sb, body, mainworldDesignation)
	return sb.String()
}

func renderClass4PBeltMarkdown(body *DetailedPlacement, sys stars.System, mainworldDesignation string) string {
	var sb strings.Builder
	sb.WriteString("## IISS Class IV-P Survey — Form 0407K-IV PART P.B\n\n")
	writeClass4PBeltHeader(&sb, body, sys)
	writeClass4PBeltOrbit(&sb, body)
	writeClass4PBeltComposition(&sb, body)
	writeClass4PBeltResources(&sb, body)
	writeClass4PBeltMajorBodies(&sb, body)
	writeClass4PComments(&sb, body, mainworldDesignation)
	return sb.String()
}

// --- Terrestrial/moon helpers ---

func writeClass4PHeader(sb *strings.Builder, body *DetailedPlacement, sys stars.System) {
	sb.WriteString("### Header\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| World | %s |\n", body.Designation)
	fmt.Fprintf(sb, "| Primary Object(s) | %s |\n", body.Group.Designation)
	fmt.Fprintf(sb, "| System Age (Gyr) | %.3f |\n\n", sys.Primary.AgeGyr)
}

func writeClass4POrbit(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Orbit\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| O# | %.2f |\n", body.Orbit)
	fmt.Fprintf(sb, "| AU | %.2f |\n", stars.OrbitToAU(body.Orbit))
	fmt.Fprintf(sb, "| Eccentricity | %.3f |\n", body.Eccentricity)
	fmt.Fprintf(sb, "| Period (h) | %.2f |\n\n", body.Period.Hours)
}

func writeClass4PSize(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Size\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Size Code | %s |\n", body.SizeCode)
	fmt.Fprintf(sb, "| Diameter (km) | %.0f |\n", body.DiameterKm)
	density, gravity := 0.0, 0.0
	if body.Physical != nil {
		density = body.Physical.Density
		gravity = body.Physical.Gravity
	}
	fmt.Fprintf(sb, "| Density | %.3f |\n", density)
	fmt.Fprintf(sb, "| Gravity | %.3f |\n", gravity)
	fmt.Fprintf(sb, "| Mass | %.3f |\n\n", body.MassEarth)
}

func writeClass4PAtmosphere(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Atmosphere\n\n")
	if body.Atmosphere == nil {
		writeNotGenerated(sb)
		return
	}
	a := body.Atmosphere
	sb.WriteString("| Field | Value |\n|---|---|\n")
	if name := AtmosphereCodeName(a.Code); name != "" {
		fmt.Fprintf(sb, "| Code | %d — %s |\n", a.Code, name)
	} else {
		fmt.Fprintf(sb, "| Code | %d |\n", a.Code)
	}
	fmt.Fprintf(sb, "| Pressure (bar) | %.3f |\n", a.Pressure)
	fmt.Fprintf(sb, "| O₂ (bar) | %.3f |\n", a.OxygenPartialPressure)
	fmt.Fprintf(sb, "| Scale Height | %.2f |\n", a.ScaleHeight)
	if sh := FormatAtmoProfileShorthand(*a, a.Profile); sh != "" {
		fmt.Fprintf(sb, "| Profile | %s |\n", sh)
	}
	sb.WriteString("\n")
}

func writeClass4PHydrographics(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Hydrographics\n\n")
	if body.Hydrographics == nil {
		writeNotGenerated(sb)
		return
	}
	h := body.Hydrographics
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Code | %d |\n", h.Code)
	fmt.Fprintf(sb, "| Coverage (%%) | %d |\n", h.Percent)
	fmt.Fprintf(sb, "| Profile | %s |\n\n", h.Profile)
}

func writeClass4PRotation(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Rotation\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	if body.DayLength != nil {
		fmt.Fprintf(sb, "| Sidereal (h) | %.2f |\n", body.DayLength.SiderealHours)
		fmt.Fprintf(sb, "| Solar (h) | %.2f |\n", body.DayLength.SolarHours)
		fmt.Fprintf(sb, "| Solar days/year | %.2f |\n", body.DayLength.YearDays)
	} else {
		sb.WriteString("| Day length | (not generated) |\n")
	}
	if body.AxialTilt != nil {
		fmt.Fprintf(sb, "| Axial Tilt (°) | %.2f |\n", body.AxialTilt.Degrees)
	} else {
		sb.WriteString("| Axial Tilt | (not generated) |\n")
	}
	tidalLockText := "no"
	if body.TidalLock != nil && body.TidalLock.LockRatio != "" {
		tidalLockText = body.TidalLock.LockRatio
	}
	fmt.Fprintf(sb, "| Tidal lock | %s |\n", tidalLockText)
	tidesM := 0.0
	if body.TidalEffects != nil {
		tidesM = body.TidalEffects.Total
	}
	fmt.Fprintf(sb, "| Tides (m) | %.2f |\n\n", tidesM)
}

func writeClass4PTemperature(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Temperature\n\n")
	if body.Temperature == nil {
		writeNotGenerated(sb)
		return
	}
	t := body.Temperature
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| High (K) | %.0f |\n", t.HighK)
	fmt.Fprintf(sb, "| Mean (K) | %.0f |\n", t.MeanK)
	// LowK == 0 is MeanTemperatureK's degenerate-model sentinel (lowL ≤ 0
	// when LuminosityModifier hits its 1.0 clamp). Suppress as em-dash so
	// users don't read a literal "0 K" claim. See issue #1.
	if t.LowK <= 0 && t.MeanK > 0 {
		sb.WriteString("| Low (K) | — |\n")
	} else {
		fmt.Fprintf(sb, "| Low (K) | %.0f |\n", t.LowK)
	}
	fmt.Fprintf(sb, "| Luminosity | %.3f |\n", t.Luminosity)
	fmt.Fprintf(sb, "| Albedo | %.2f |\n", t.Albedo)
	fmt.Fprintf(sb, "| Greenhouse | %.2f |\n\n", t.GreenhouseFactor)
}

func writeClass4PSeismic(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Seismic\n\n")
	if body.Geology == nil {
		writeNotGenerated(sb)
		return
	}
	g := body.Geology
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Total Seismic Stress | %d — %s |\n", g.TotalSeismicStress, TotalSeismicStressLabel(g.TotalSeismicStress))
	fmt.Fprintf(sb, "| Residual Stress | %d |\n", g.ResidualSeismicStress)
	fmt.Fprintf(sb, "| Tidal Stress | %d |\n", g.TidalStressFactor)
	fmt.Fprintf(sb, "| Tidal Heating | %d — %s |\n", g.TidalHeatingFactor, TidalHeatingFactorLabel(g.TidalHeatingFactor))
	fmt.Fprintf(sb, "| Tectonic Plates | %d |\n\n", g.TectonicPlates)
}

func writeClass4PLife(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Life\n\n")
	if body.Biology == nil {
		writeNotGenerated(sb)
		return
	}
	b := body.Biology
	sophontStr := "no"
	if b.HasNativeSophont {
		sophontStr = "yes"
	} else if b.HadExtinctSophont {
		sophontStr = "extinct"
	}
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Biomass | %s — %s |\n", string(eHexDigit(b.Biomass)), BiomassLabel(b.Biomass))
	if name := BiocomplexityName(b.Biocomplexity); name != "" {
		fmt.Fprintf(sb, "| Biocomplexity | %d — %s |\n", b.Biocomplexity, name)
	} else {
		fmt.Fprintf(sb, "| Biocomplexity | %d |\n", b.Biocomplexity)
	}
	fmt.Fprintf(sb, "| Sophonts? | %s |\n", sophontStr)
	fmt.Fprintf(sb, "| Biodiversity | %d — %s |\n", b.Biodiversity, BiodiversityLabel(b.Biodiversity))
	fmt.Fprintf(sb, "| Compatibility | %d — %s |\n\n", b.Compatibility, CompatibilityLabel(b.Compatibility))
}

func writeClass4PResources(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Resources\n\n")
	if body.Biology == nil {
		writeNotGenerated(sb)
		return
	}
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Rating | %s — %s |\n\n",
		string(eHexDigit(body.Biology.ResourceRating)),
		ResourceRatingName(body.Biology.ResourceRating))
}

func writeClass4PHabitability(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Habitability\n\n")
	if body.Habitability == nil {
		writeNotGenerated(sb)
		return
	}
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Rating | %d — %s |\n\n",
		body.Habitability.Rating, HabitabilityRatingName(body.Habitability.Rating))
}

func writeClass4PSubordinates(sb *strings.Builder, body *DetailedPlacement) {
	if len(body.Moons) == 0 {
		return
	}
	sb.WriteString("### Subordinates\n\n")
	sb.WriteString("| Designation | Size | Diameter (km) | Orbit (km) | Eccentricity | Period (h) |\n")
	sb.WriteString("|---|---|---|---|---|---|\n")
	for _, m := range body.Moons {
		fmt.Fprintf(sb, "| %s | %s | %.0f | %d | %.3f | %.2f |\n",
			m.Designation, m.SizeCode, m.DiameterKm, m.OrbitKm, m.Eccentricity, m.PeriodHours)
	}
	sb.WriteString("\n")
}

// --- Belt helpers ---

func writeClass4PBeltHeader(sb *strings.Builder, body *DetailedPlacement, sys stars.System) {
	sb.WriteString("### Header\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| World | %s |\n", body.Designation)
	sb.WriteString("| SAH/UWP | 000 |\n")
	fmt.Fprintf(sb, "| Primary Object(s) | %s |\n", body.Group.Designation)
	fmt.Fprintf(sb, "| System Age (Gyr) | %.3f |\n\n", sys.Primary.AgeGyr)
}

func writeClass4PBeltOrbit(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Orbit\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| O# | %.2f |\n", body.Orbit)
	fmt.Fprintf(sb, "| AU | %.2f |\n", stars.OrbitToAU(body.Orbit))
	if body.Belt != nil {
		fmt.Fprintf(sb, "| Span (Orbit#s) | %.3f |\n", body.Belt.Span)
	} else {
		sb.WriteString("| Span | (not available) |\n")
	}
	fmt.Fprintf(sb, "| Period (h) | %.2f |\n\n", body.Period.Hours)
}

func writeClass4PBeltComposition(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Composition\n\n")
	if body.Belt == nil {
		writeNotGenerated(sb)
		return
	}
	c := body.Belt.Composition
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| m-type (%%) | %d |\n", c.MTypePct)
	fmt.Fprintf(sb, "| s-type (%%) | %d |\n", c.STypePct)
	fmt.Fprintf(sb, "| c-type (%%) | %d |\n", c.CTypePct)
	fmt.Fprintf(sb, "| other (%%) | %d |\n", c.OtherPct)
	fmt.Fprintf(sb, "| Bulk | %d |\n", body.Belt.Bulk)
	fmt.Fprintf(sb, "| Major Bodies (Size 1) | %d |\n", body.Belt.SigSize1Bodies)
	fmt.Fprintf(sb, "| Major Bodies (Size S) | %d |\n\n", body.Belt.SigSizeSBodies)
}

func writeClass4PBeltResources(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Resources\n\n")
	if body.Belt == nil {
		writeNotGenerated(sb)
		return
	}
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Rating | %d |\n\n", body.Belt.ResourceRating)
}

func writeClass4PBeltMajorBodies(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Major Bodies\n\n")
	if body.Belt == nil {
		writeNotGenerated(sb)
		return
	}
	fmt.Fprintf(sb, "_Counts only: %d size-1 + %d size-S; per-body detail not generated._\n\n",
		body.Belt.SigSize1Bodies, body.Belt.SigSizeSBodies)
}

// --- Shared helpers ---

func writeClass4PComments(sb *strings.Builder, body *DetailedPlacement, mainworldDesignation string) {
	sb.WriteString("### Comments\n\n")
	if mainworldDesignation != "" && body.Designation == mainworldDesignation {
		sb.WriteString("This is the system mainworld.\n\n")
	}
}

func writeNotGenerated(sb *strings.Builder) {
	sb.WriteString("| Field | Value |\n|---|---|\n")
	sb.WriteString("| Status | (not generated) |\n\n")
}

// RenderClass23Markdown renders the IISS Class II/III Survey form (WBH
// pp.60-67, Form 0421D-II.III) as a Markdown section. Output starts
// with an H2 form heading and contains H3 sub-sections for Header,
// Object Counts, Stars, and Objects.
func RenderClass23Markdown(form IISSClass23Form) string {
	var sb strings.Builder
	sb.WriteString("## IISS Class II/III Survey — Form 0421D-II.III\n\n")
	writeClass23Header(&sb, form)
	writeClass23ObjectCounts(&sb, form)
	writeClass23Stars(&sb, form.Stars)
	writeClass23Objects(&sb, form.Objects)
	if form.Notes != "" {
		fmt.Fprintf(&sb, "### Notes\n\n%s\n\n", form.Notes)
	}
	if form.Comments != "" {
		fmt.Fprintf(&sb, "### Comments\n\n%s\n\n", form.Comments)
	}
	return sb.String()
}

func writeClass23Header(sb *strings.Builder, form IISSClass23Form) {
	sb.WriteString("### Header\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Sector | %s |\n", emDashIfEmpty23(form.Sector))
	fmt.Fprintf(sb, "| Location | %s |\n", emDashIfEmpty23(form.Location))
	fmt.Fprintf(sb, "| IISS Designation | %s |\n", emDashIfEmpty23(form.IISSDesig))
	fmt.Fprintf(sb, "| Initial Survey | %s |\n", emDashIfEmpty23(form.InitialSurvey))
	fmt.Fprintf(sb, "| Last Updated | %s |\n", emDashIfEmpty23(form.LastUpdated))
	fmt.Fprintf(sb, "| System Age (Gyr) | %.3f |\n", form.SystemAgeGyr)
	classIIIStr := "no"
	if form.ClassIIIStatus {
		classIIIStr = "yes"
	}
	fmt.Fprintf(sb, "| Class III Status | %s |\n\n", classIIIStr)
}

func writeClass23ObjectCounts(sb *strings.Builder, form IISSClass23Form) {
	sb.WriteString("### Object Counts\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Stellar | %d |\n", form.StellarCount)
	fmt.Fprintf(sb, "| Gas Giants | %d |\n", form.GasGiants)
	fmt.Fprintf(sb, "| Planetoid Belts | %d |\n", form.PlanetoidBelts)
	fmt.Fprintf(sb, "| Terrestrials | %d |\n\n", form.Terrestrials)
}

// writeClass23Stars renders the same Stars table data as Class 0/I, but
// in this package — the spec chose duplication over extracting a shared
// helper across packages.
func writeClass23Stars(sb *strings.Builder, components []stars.SurveyComponent) {
	sb.WriteString("### Stars\n\n")
	sb.WriteString("| Component | Class | Mass | Temperature | Diameter | Luminosity | Orbit | AU | Eccentricity | Period (y) | HZCO | MAO |\n")
	sb.WriteString("|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, c := range components {
		// Mass and Luminosity always render with a value (composites sum to non-zero);
		// other numerics use floatNonZero23 so 0 → em-dash for "not applicable".
		fmt.Fprintf(
			sb,
			"| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			emDashIfEmpty23(c.Component),
			emDashIfEmpty23(c.Class),
			fmt.Sprintf("%.3f", c.Mass),
			floatNonZero23(c.Temperature, 0),
			floatNonZero23(c.Diameter, 3),
			fmt.Sprintf("%.3f", c.Luminosity),
			floatNonZero23(c.Orbit, 2),
			floatNonZero23(c.AU, 3),
			floatNonZero23(c.Eccentricity, 2),
			floatNonZero23(c.PeriodYears, 3),
			floatNonZero23(c.HZCO, 2),
			floatNonZero23(c.MAO, 2),
		)
	}
	sb.WriteString("\n")
}

func writeClass23Objects(sb *strings.Builder, rows []ObjectRow) {
	sb.WriteString("### Objects\n\n")
	sb.WriteString("| Primary | Designation | Orbit | AU | Eccentricity | Period | SAH/UWP | Subs | Notes |\n")
	sb.WriteString("|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(
			sb,
			"| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			emDashIfEmpty23(r.Primary),
			emDashIfEmpty23(r.Designation),
			floatNonZero23(r.Orbit, 2),
			floatNonZero23(r.AU, 3),
			floatNonZero23(r.Ecc, 3),
			emDashIfEmpty23(r.PeriodStr),
			emDashIfEmpty23(r.SAH),
			emDashIfEmpty23(r.Sub),
			emDashIfEmpty23(r.Notes),
		)
	}
	sb.WriteString("\n")
}

// emDashIfEmpty23 is the worlds-package counterpart to stars/markdown.go's
// emDashIfEmpty — duplicated intentionally per spec (no shared helper across
// package boundaries).
func emDashIfEmpty23(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func floatNonZero23(v float64, prec int) string {
	if v == 0 {
		return "—"
	}
	return fmt.Sprintf("%.*f", prec, v)
}
