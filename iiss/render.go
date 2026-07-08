package iiss

import (
	"fmt"
	"strings"
)

// MarkdownClass4Survey renders the full IISS Class IV Survey document as
// Markdown: PART 1 (the system census) followed by a per-body PART P /
// PART P.B for every surveyed body. This is the sole Markdown output — the
// Class 0/I and Class II/III "short forms" (earlier survey stages) are
// folded into PART 1.
func MarkdownClass4Survey(sf SystemForms) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — IISS Class IV Survey\n\n", sf.Class0I.IISSDesig)
	if sf.MainworldDesignation != "" {
		fmt.Fprintf(&b, "**Mainworld:** %s\n\n", sf.MainworldDesignation)
	}
	if sf.ShortProfile != "" {
		fmt.Fprintf(&b, "Short profile: `%s`\n\n", sf.ShortProfile)
	}
	if sf.LongProfile != "" {
		fmt.Fprintf(&b, "Long profile: `%s`\n\n", sf.LongProfile)
	}
	if sf.NotableFeatures != "" {
		b.WriteString(sf.NotableFeatures)
		b.WriteString("\n")
	}
	b.WriteString(markdownClass4Part1(sf))
	for _, f := range sf.Class4PForms {
		b.WriteString("\n")
		b.WriteString(markdownClass4Part(f))
	}
	return b.String()
}

// markdownClass4Part1 renders PART 1 — the system census: system-level
// scalars, world counts, the stellar roster (with full orbital data), and
// the body roster. Its data source is the retained Class0I star rows and
// Class23 object rows (the content the old short forms carried).
func markdownClass4Part1(sf SystemForms) string {
	var b strings.Builder
	c0 := sf.Class0I
	b.WriteString("## PART 1 — System Census\n\n")

	fmt.Fprintf(&b, "- System: %s\n", c0.SystemName)
	fmt.Fprintf(&b, "- Sector / Location: %s / %s\n", c0.Sector, c0.Location)
	fmt.Fprintf(&b, "- Survey: initial %s, last updated %s\n", c0.InitialSurvey, c0.LastUpdated)
	fmt.Fprintf(&b, "- System age: %.3f Gyr\n", c0.SystemAgeGyr)
	fmt.Fprintf(&b, "- Stellar count: %d\n", c0.StellarCount)
	fmt.Fprintf(&b, "- Worlds: %d gas giants, %d belts, %d terrestrials (total %d)\n",
		sf.Class23.Counts.GasGiants, sf.Class23.Counts.PlanetoidBelts,
		sf.Class23.Counts.Terrestrials, sf.Class23.Counts.Total)
	fmt.Fprintf(&b, "- Baseline: number %d, Orbit# %.2f; spread %.2f; empty orbits %d\n\n",
		sf.Census.BaselineNumber, sf.Census.BaselineOrbit, sf.Census.Spread, sf.Census.EmptyOrbits)

	if len(c0.Stars) > 0 {
		b.WriteString("### Stars\n\n")
		b.WriteString("| Component | Class | Mass | Diameter | Temp (K) | Luminosity | Orbit | AU | Ecc | Period (y) | MAO | HZCO | HZ Orbit# |\n")
		b.WriteString("| --------- | ----- | ---- | -------- | -------- | ---------- | ----- | --- | --- | ---------- | --- | ---- | --------- |\n")
		for _, s := range c0.Stars {
			fmt.Fprintf(&b, "| %s | %s | %.3f | %.3f | %.0f | %.4f | %s | %s | %s | %s | %.2f | %.2f | %s |\n",
				s.Component, s.Class, s.Mass, s.Diameter, s.Temperature, s.Luminosity,
				blankFloat(s.Orbit), blankFloat(s.AU), blankFloat(s.Eccentricity),
				blankFloat(s.PeriodYears), s.MAO, s.HZCO, hzRange(s.HZCO))
		}
		b.WriteString("\nHabitable zone breadth: ±1.0 Orbit# from HZCO (WBH p.43).\n\n")
	}

	if len(sf.Class23.Objects) > 0 {
		b.WriteString("### Bodies\n\n")
		b.WriteString("| Primary | Designation | Orbit | AU | Ecc | Period | SAH | Sub | Notes |\n")
		b.WriteString("| ------- | ----------- | ----- | --- | --- | ------ | --- | --- | ----- |\n")
		for _, o := range sf.Class23.Objects {
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
				o.Primary, o.Designation, blankFloat(o.Orbit), blankFloat(o.AU),
				blankFloat(o.Ecc), o.PeriodStr, o.SAH, o.Sub, o.Notes)
		}
	}
	return b.String()
}

// markdownClass4Part renders one per-body PART P / PART P.B, titled by the
// body's own Designation. The body content comes from the concrete
// PartP / PartPB RenderBody methods.
func markdownClass4Part(f Class4PForm) string {
	var b strings.Builder
	switch f.Variant {
	case Class4PBelt:
		fmt.Fprintf(&b, "## PART P.B — %s (Belt)%s\n\n", f.Designation, mainworldSuffix(f))
		if f.PartPB != nil {
			f.PartPB.RenderBody(&b, f.FormHeader)
		}
	case Class4PMoon:
		fmt.Fprintf(&b, "## PART P — %s (Moon)%s\n\n", f.Designation, mainworldSuffix(f))
		if f.PartP != nil {
			f.PartP.RenderBody(&b, f.FormHeader)
		}
	default:
		fmt.Fprintf(&b, "## PART P — %s%s\n\n", f.Designation, mainworldSuffix(f))
		if f.PartP != nil {
			f.PartP.RenderBody(&b, f.FormHeader)
		}
	}
	return b.String()
}

// mainworldSuffix returns " — mainworld" for the auto-picked mainworld's
// part, empty otherwise.
func mainworldSuffix(f Class4PForm) string {
	if (f.PartP != nil && f.PartP.IsMainworld) || (f.PartPB != nil && f.PartPB.IsMainworld) {
		return " — mainworld"
	}
	return ""
}

// blankFloat formats a float for a table cell, blanking a zero value.
func blankFloat(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%.2f", v)
}

// hzRange renders the habitable-zone Orbit# breadth (HZCO ± 1.0, WBH p.43)
// for a star row that carries an HZCO; blank otherwise.
func hzRange(hzco float64) string {
	if hzco <= 0 {
		return ""
	}
	return fmt.Sprintf("%.2f–%.2f", hzco-1, hzco+1)
}
