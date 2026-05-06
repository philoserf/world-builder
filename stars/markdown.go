package stars

import (
	"fmt"
	"strings"
)

// RenderClass0IMarkdown renders the IISS Class 0/I Survey form (WBH p.33,
// Form 0421B-0I) as a Markdown section. Output starts with an H2 form
// heading and contains H3 sub-sections for Header and Stars. Empty/zero
// fields render as em-dash.
func RenderClass0IMarkdown(form SurveyForm) string {
	var sb strings.Builder
	sb.WriteString("## IISS Class 0/I Survey — Form 0421B-0I\n\n")
	writeClass0IHeader(&sb, form)
	writeClass0IStars(&sb, form.Stars)
	if form.Notes != "" {
		fmt.Fprintf(&sb, "### Notes\n\n%s\n\n", form.Notes)
	}
	if form.Comments != "" {
		fmt.Fprintf(&sb, "### Comments\n\n%s\n\n", form.Comments)
	}
	return sb.String()
}

func writeClass0IHeader(sb *strings.Builder, form SurveyForm) {
	sb.WriteString("### Header\n\n")
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("|---|---|\n")
	fmt.Fprintf(sb, "| Sector | %s |\n", emDashIfEmpty(form.Sector))
	fmt.Fprintf(sb, "| Location | %s |\n", emDashIfEmpty(form.Location))
	fmt.Fprintf(sb, "| IISS Designation | %s |\n", emDashIfEmpty(form.IISSDesig))
	fmt.Fprintf(sb, "| Initial Survey | %s |\n", emDashIfEmpty(form.InitialSurvey))
	fmt.Fprintf(sb, "| Last Updated | %s |\n", emDashIfEmpty(form.LastUpdated))
	fmt.Fprintf(sb, "| System Age (Gyr) | %.3f |\n", form.SystemAgeGyr)
	fmt.Fprintf(sb, "| Stellar Count | %d |\n", form.StellarCount)
	sb.WriteString("\n")
}

func writeClass0IStars(sb *strings.Builder, stars []SurveyComponent) {
	sb.WriteString("### Stars\n\n")
	sb.WriteString("| Component | Class | Mass | Temperature | Diameter | Luminosity | Orbit | AU | Eccentricity | Period (y) | HZCO |\n")
	sb.WriteString("|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, c := range stars {
		// Mass and Luminosity always render with a value (composites sum to non-zero);
		// other numerics use formatFloatNonZero so 0 → em-dash for "not applicable".
		fmt.Fprintf(
			sb, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			emDashIfEmpty(c.Component),
			emDashIfEmpty(c.Class),
			formatFloat(c.Mass, 3),
			formatFloatNonZero(c.Temperature, 0),
			formatFloatNonZero(c.Diameter, 3),
			formatFloat(c.Luminosity, 3),
			formatFloatNonZero(c.Orbit, 2),
			formatFloatNonZero(c.AU, 3),
			formatFloatNonZero(c.Eccentricity, 2),
			formatFloatNonZero(c.PeriodYears, 3),
			formatFloatNonZero(c.HZCO, 2),
		)
	}
	sb.WriteString("\n")
}

// emDashIfEmpty returns "—" for empty strings.
func emDashIfEmpty(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// formatFloat renders a float to the given precision; never returns em-dash.
func formatFloat(v float64, prec int) string {
	return fmt.Sprintf("%.*f", prec, v)
}

// formatFloatNonZero renders a float to the given precision, but returns
// em-dash if the value is zero. Use for fields where 0 means "not applicable"
// (e.g., Orbit on the primary's row, Temperature on a barycentre composite).
func formatFloatNonZero(v float64, prec int) string {
	if v == 0 {
		return "—"
	}
	return fmt.Sprintf("%.*f", prec, v)
}
