package iiss

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MarkdownClass0I renders the Class 0/I form as Markdown — header,
// system age, stellar census table.
func MarkdownClass0I(f Class0IForm) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Class 0/I — %s\n\n", f.IISSDesig)
	fmt.Fprintf(&b, "- System: %s\n", f.SystemName)
	fmt.Fprintf(&b, "- Sector / Location: %s / %s\n", f.Sector, f.Location)
	fmt.Fprintf(&b, "- Survey: initial %s, last updated %s\n", f.InitialSurvey, f.LastUpdated)
	fmt.Fprintf(&b, "- System age: %.3f Gyr\n", f.SystemAgeGyr)
	fmt.Fprintf(&b, "- Stellar count: %d\n\n", f.StellarCount)
	if len(f.Stars) > 0 {
		b.WriteString("| Component | Class | Mass | Diameter | Temp (K) | Luminosity | HZCO |\n")
		b.WriteString("| --------- | ----- | ---- | -------- | -------- | ---------- | ---- |\n")
		for _, s := range f.Stars {
			fmt.Fprintf(&b, "| %s | %s | %.3f | %.3f | %.0f | %.4f | %.2f |\n",
				s.Component, s.Class, s.Mass, s.Diameter, s.Temperature, s.Luminosity, s.HZCO)
		}
	}
	return b.String()
}

// MarkdownClass23 renders the Class II/III form as Markdown —
// counts, stellar census with MAO, and the WBH p.61 Objects table.
func MarkdownClass23(f Class23Form) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Class II/III — %s\n\n", f.IISSDesig)
	fmt.Fprintf(&b, "- Gas giants: %d\n", f.Counts.GasGiants)
	fmt.Fprintf(&b, "- Belts: %d\n", f.Counts.PlanetoidBelts)
	fmt.Fprintf(&b, "- Terrestrials: %d\n", f.Counts.Terrestrials)
	fmt.Fprintf(&b, "- Total: %d\n\n", f.Counts.Total)

	if len(f.Stars) > 0 {
		b.WriteString("### Stars (with MAO)\n\n")
		b.WriteString("| Component | Class | Mass | Diameter | Temp (K) | Luminosity | HZCO | MAO |\n")
		b.WriteString("| --------- | ----- | ---- | -------- | -------- | ---------- | ---- | --- |\n")
		for _, s := range f.Stars {
			fmt.Fprintf(&b, "| %s | %s | %.3f | %.3f | %.0f | %.4f | %.2f | %.2f |\n",
				s.Component, s.Class, s.Mass, s.Diameter, s.Temperature, s.Luminosity, s.HZCO, s.MAO)
		}
		b.WriteString("\n")
	}

	if len(f.Objects) > 0 {
		b.WriteString("### Objects\n\n")
		b.WriteString("| Primary | Designation | Orbit | AU | Ecc | Period | SAH | Sub | Notes |\n")
		b.WriteString("| ------- | ----------- | ----- | --- | --- | ------ | --- | --- | ----- |\n")
		for _, o := range f.Objects {
			ecc := ""
			if o.Ecc > 0 {
				ecc = fmt.Sprintf("%.2f", o.Ecc)
			}
			orbit := ""
			if o.Orbit > 0 {
				orbit = fmt.Sprintf("%.2f", o.Orbit)
			}
			au := ""
			if o.AU > 0 {
				au = fmt.Sprintf("%.2f", o.AU)
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
				o.Primary, o.Designation, orbit, au, ecc, o.PeriodStr, o.SAH, o.Sub, o.Notes)
		}
	}
	return b.String()
}

// MarkdownClass4P renders the WBH p.138 IISS Class IV Survey Form
// (FORM 0407F-IV PART P) for planet/moon mainworlds; for belts,
// FORM 0407K-IV PART P.B.
func MarkdownClass4P(f Class4PForm) string {
	var b strings.Builder
	switch f.Variant {
	case Class4PBelt:
		fmt.Fprintf(&b, "## Class IV-P PART P.B — %s (Belt mainworld)\n\n", f.IISSDesig)
		writeClass4PPartPB(&b, f.PartPB, f.FormHeader)
	case Class4PMoon:
		fmt.Fprintf(&b, "## Class IV-P PART P — %s (Moon mainworld)\n\n", f.IISSDesig)
		writeClass4PPartP(&b, f.PartP, f.FormHeader)
	default:
		fmt.Fprintf(&b, "## Class IV-P PART P — %s\n\n", f.IISSDesig)
		writeClass4PPartP(&b, f.PartP, f.FormHeader)
	}
	return b.String()
}

func writeClass4PPartP(b *strings.Builder, p *Class4PPartP, h FormHeader) {
	if p == nil {
		return
	}
	fmt.Fprintf(b, "**WORLD:** %s\n", p.Designation)
	fmt.Fprintf(b, "**SECTOR | LOCATION:** %s | %s\n", h.Sector, h.Location)
	fmt.Fprintf(b, "**INITIAL SURVEY:** %s   **LAST UPDATED:** %s\n", h.InitialSurvey, h.LastUpdated)
	fmt.Fprintf(b, "**SYSTEM AGE (Gyr):** %.3f\n\n", p.SystemAgeGyr)

	b.WriteString("### ORBIT\n")
	fmt.Fprintf(b, "- AU: %.2f, Eccentricity: %.2f, Period (h): %.2f\n\n",
		p.AU, p.Eccentricity, p.PeriodHours)

	b.WriteString("### SIZE\n")
	fmt.Fprintf(b, "- Diameter (km): %.0f, Density: %.2f, Gravity: %.2f, Mass (Earth): %.3f\n\n",
		p.DiameterKm, p.Density, p.Gravity, p.MassEarth)

	b.WriteString("### ATMOSPHERE\n")
	if p.Atmosphere == nil {
		b.WriteString("- (none — vacuum)\n\n")
	} else {
		atm := p.Atmosphere
		fmt.Fprintf(b, "- Code: %d, Pressure (bar): %.3f, O₂ (bar): %.3f, Scale Height: %.2f\n",
			atm.Code, atm.Pressure, atm.OxygenPartialPressure, atm.ScaleHeight)
		if atm.ProfileShorthand != "" {
			fmt.Fprintf(b, "- Profile: %s\n", atm.ProfileShorthand)
		}
		b.WriteString("\n")
	}

	b.WriteString("### HYDROGRAPHICS\n")
	if p.Hydrographics == nil {
		b.WriteString("- (none)\n\n")
	} else {
		fmt.Fprintf(b, "- Code: %d, Coverage (%%): %d, Profile: %s\n\n",
			p.Hydrographics.Code, p.Hydrographics.Percent, p.Hydrographics.Profile)
	}

	b.WriteString("### ROTATION\n")
	fmt.Fprintf(b, "- Sidereal (h): %.2f, Solar (h): %.2f, Solar days/year: %.2f\n",
		p.SiderealHours, p.SolarHours, p.SolarDaysPerYear)
	fmt.Fprintf(b, "- Axial Tilt: %.2f°\n", p.AxialTiltDeg)
	fmt.Fprintf(b, "- Tidal lock: %s, Tides (m): %.2f\n\n", p.TidalLockRatio, p.TidesMeters)

	b.WriteString("### TEMPERATURE\n")
	if p.Temperature == nil {
		b.WriteString("- (not computed)\n\n")
	} else {
		t := p.Temperature
		low := fmt.Sprintf("%.1f", t.LowK)
		if t.LowK < 0 {
			low = "—"
		}
		fmt.Fprintf(b, "- High (K): %.1f, Mean (K): %.1f, Low (K): %s\n",
			t.HighK, t.MeanK, low)
		fmt.Fprintf(b, "- Luminosity: %.3f, Albedo: %.2f, Greenhouse: %.2f\n\n",
			t.Luminosity, t.Albedo, t.GreenhouseFactor)
	}

	b.WriteString("### SEISMIC\n")
	if p.Seismic == nil {
		b.WriteString("- (not computed)\n\n")
	} else {
		s := p.Seismic
		fmt.Fprintf(b, "- TSS: %d, Residual: %d, Tidal Stress: %d, Tidal Heating: %d, Plates: %d\n\n",
			s.TotalSeismicStress, s.ResidualSeismicStress, s.TidalStressFactor, s.TidalHeatingFactor, s.TectonicPlates)
	}

	b.WriteString("### LIFE\n")
	if p.Life == nil {
		b.WriteString("- (not computed)\n\n")
	} else {
		l := p.Life
		soph := "no"
		if l.HasSophont {
			soph = "yes"
		} else if l.HadExtinct {
			soph = "extinct"
		}
		fmt.Fprintf(b, "- Biomass: %d, Biocomplexity: %d, Sophonts: %s, Biodiversity: %d, Compatibility: %d\n\n",
			l.Biomass, l.Biocomplexity, soph, l.Biodiversity, l.Compatibility)
		fmt.Fprintf(b, "### RESOURCES\n- Rating: %d\n\n", l.ResourceRating)
	}

	b.WriteString("### HABITABILITY\n")
	fmt.Fprintf(b, "- Rating: %d\n", p.HabitabilityRating)
	if p.HabitabilityNotes != "" {
		fmt.Fprintf(b, "- Notes: %s\n", p.HabitabilityNotes)
	}
	b.WriteString("\n")

	if len(p.Subordinates) > 0 {
		b.WriteString("### SUBORDINATES\n")
		b.WriteString("| Designation | Size | Diameter (km) | Orbit (km) | Ecc | Period (h) |\n")
		b.WriteString("| ----------- | ---- | ------------- | ---------- | --- | ---------- |\n")
		for _, s := range p.Subordinates {
			fmt.Fprintf(b, "| %s | %s | %.0f | %d | %.3f | %.2f |\n",
				s.Designation, s.SizeCode, s.DiameterKm, s.OrbitKm, s.Eccentricity, s.PeriodHours)
		}
		b.WriteString("\n")
	}

	if p.IsMainworld {
		b.WriteString("### COMMENTS\n- This is the system mainworld.\n\n")
	}
}

func writeClass4PPartPB(b *strings.Builder, pb *Class4PPartPB, h FormHeader) {
	if pb == nil {
		return
	}
	fmt.Fprintf(b, "**WORLD:** %s   **SAH/UWP:** 000\n", pb.Designation)
	fmt.Fprintf(b, "**SECTOR | LOCATION:** %s | %s\n", h.Sector, h.Location)
	fmt.Fprintf(b, "**INITIAL SURVEY:** %s   **LAST UPDATED:** %s\n", h.InitialSurvey, h.LastUpdated)
	fmt.Fprintf(b, "**PRIMARY OBJECT(S):** %s   **SYSTEM AGE (Gyr):** %.3f\n\n",
		pb.PrimaryGroup, pb.SystemAgeGyr)

	b.WriteString("### ORBIT\n")
	fmt.Fprintf(b, "- O#: %.2f, AU: %.2f, Span: %.3f Orbit#s, Period (h): %.2f\n\n",
		pb.OrbitNumber, pb.AU, pb.SpanOrbits, pb.PeriodHours)

	b.WriteString("### COMPOSITION\n")
	fmt.Fprintf(b, "- m-type%%: %d, s-type%%: %d, c-type%%: %d, other%%: %d\n",
		pb.MTypePct, pb.STypePct, pb.CTypePct, pb.OtherPct)
	fmt.Fprintf(b, "- Bulk: %d\n", pb.Bulk)
	fmt.Fprintf(b, "- Major Bodies: Size 1 = %d, Size S = %d\n\n",
		pb.SigSize1Bodies, pb.SigSizeSBodies)

	b.WriteString("### RESOURCES\n")
	fmt.Fprintf(b, "- Rating: %d\n\n", pb.ResourceRating)

	b.WriteString("### MAJOR BODIES\n")
	fmt.Fprintf(b, "- Counts only: %d size-1 + %d size-S; per-body detail post-parity.\n\n",
		pb.SigSize1Bodies, pb.SigSizeSBodies)

	if pb.IsMainworld {
		b.WriteString("### COMMENTS\n- This is the system mainworld.\n\n")
	}
}

// MarkdownSystem concatenates all three forms under H1/H2 headings in
// book order. Class IV-P renders only for the auto-picked mainworld.
func MarkdownSystem(sf SystemForms) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — System Survey\n\n", sf.Class0I.IISSDesig)
	if sf.MainworldDesignation != "" {
		fmt.Fprintf(&b, "**Mainworld:** %s\n\n", sf.MainworldDesignation)
	}
	if sf.ShortProfile != "" {
		fmt.Fprintf(&b, "Short profile: `%s`\n\n", sf.ShortProfile)
	}
	if sf.LongProfile != "" {
		fmt.Fprintf(&b, "Long profile: `%s`\n\n", sf.LongProfile)
	}
	b.WriteString(MarkdownClass0I(sf.Class0I))
	b.WriteString("\n")
	b.WriteString(MarkdownClass23(sf.Class23))
	b.WriteString("\n")
	if sf.MainworldDesignation != "" {
		b.WriteString(MarkdownClass4P(sf.Class4P))
		b.WriteString("\n")
	}
	return b.String()
}

// JSONClass0I returns the Class 0/I form as JSON.
func JSONClass0I(f Class0IForm) ([]byte, error) {
	return json.MarshalIndent(f, "", "  ")
}

// JSONClass23 returns the Class II/III form as JSON.
func JSONClass23(f Class23Form) ([]byte, error) {
	return json.MarshalIndent(f, "", "  ")
}

// JSONClass4P returns the Class IV-P form as JSON.
func JSONClass4P(f Class4PForm) ([]byte, error) {
	return json.MarshalIndent(f, "", "  ")
}

// PlainTextClass0I renders the Class 0/I form as a stripped-down
// plain-text version of the Markdown rendering.
func PlainTextClass0I(f Class0IForm) string {
	return MarkdownClass0I(f)
}

// PlainTextClass23 renders the Class II/III form as plain text.
func PlainTextClass23(f Class23Form) string {
	return MarkdownClass23(f)
}

// PlainTextClass4P renders the Class IV-P form as plain text.
func PlainTextClass4P(f Class4PForm) string {
	return MarkdownClass4P(f)
}
