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

// MarkdownClass4P renders the Class IV-P form. Cycle-11 MVP emits a
// stub block per variant; full cell-by-cell rendering lands as a
// post-parity sub-project.
func MarkdownClass4P(f Class4PForm) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Class IV-P — %s\n\n", f.IISSDesig)
	switch f.Variant {
	case Class4PPlanet:
		b.WriteString("Variant: Planet\n")
	case Class4PMoon:
		b.WriteString("Variant: Moon\n")
	case Class4PBelt:
		b.WriteString("Variant: Belt\n")
	}
	if f.PartP != nil {
		b.WriteString("\n_(Class IV-P PART P body detail — full layout deferred post-parity.)_\n")
	}
	if f.PartPB != nil {
		b.WriteString("\n_(Class IV-P PART P.B belt detail — full layout deferred post-parity.)_\n")
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
