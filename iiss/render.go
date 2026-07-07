package iiss

import (
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
// FORM 0407K-IV PART P.B. The body is rendered from the concrete
// PartP / PartPB struct on the form.
func MarkdownClass4P(f Class4PForm) string {
	var b strings.Builder
	switch f.Variant {
	case Class4PBelt:
		fmt.Fprintf(&b, "## Class IV-P PART P.B — %s (Belt mainworld)\n\n", f.IISSDesig)
	case Class4PMoon:
		fmt.Fprintf(&b, "## Class IV-P PART P — %s (Moon mainworld)\n\n", f.IISSDesig)
	default:
		fmt.Fprintf(&b, "## Class IV-P PART P — %s\n\n", f.IISSDesig)
	}
	switch {
	case f.PartPB != nil:
		f.PartPB.RenderBody(&b, f.FormHeader)
	case f.PartP != nil:
		f.PartP.RenderBody(&b, f.FormHeader)
	}
	return b.String()
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
	if sf.NotableFeatures != "" {
		b.WriteString(sf.NotableFeatures)
		b.WriteString("\n")
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
