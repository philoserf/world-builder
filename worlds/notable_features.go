package worlds

import (
	"fmt"
	"strings"
)

// Notable-feature thresholds. Tuned for "what would a Referee want at a
// glance" — borrowed from habitability.go's DM bands where applicable
// so the flags align with the book's hazard rhetoric.
const (
	notableGravityHigh    = 1.4 // > p.132 "uncomfortably high" boundary
	notablePressureHigh   = 2.5 // bar; structures need reinforcement, breathing problematic
	notableWorstLowK      = 233 // ~ -40°C; severe frostbite, sustained cold lethal to unprotected humans
	notableMeanLivableK   = 250 // ~ -23°C; below this the mean is already too cold for a "snap" to be the story
	notableTaintChainSize = 2   // 2+ taints = "chain"
)

// NotableFeatures scans the universe and returns a referee-facing
// Markdown block flagging tidal locks, cold-snap scenarios, crush
// worlds (high gravity + high pressure), taint chains, and the
// mainworld's habitability rationale. Returns "" if nothing notable.
//
// Inserted by BuildIISSForms above the IISS forms in cmd/world-builder's
// Markdown output. The block is informational only — every fact
// surfaced is already present in the IISS forms, just dispersed
// across them.
func NotableFeatures(u *Universe) string {
	if u == nil {
		return ""
	}

	tidals := collectTidalLocks(u)
	colds := collectColdSnaps(u)
	crushes := collectCrushWorlds(u)
	taints := collectTaintChains(u)
	mainworldNote := mainworldHabitabilityNote(u)

	if len(tidals) == 0 && len(colds) == 0 && len(crushes) == 0 && len(taints) == 0 && mainworldNote == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Notable Features\n\n")

	if len(tidals) > 0 {
		b.WriteString("### Tidal locks\n")

		for _, line := range tidals {
			fmt.Fprintf(&b, "- %s\n", line)
		}

		b.WriteString("\n")
	}

	if len(colds) > 0 {
		b.WriteString("### Cold snaps\n")

		for _, line := range colds {
			fmt.Fprintf(&b, "- %s\n", line)
		}

		b.WriteString("\n")
	}

	if len(crushes) > 0 {
		b.WriteString("### Crush worlds\n")

		for _, line := range crushes {
			fmt.Fprintf(&b, "- %s\n", line)
		}

		b.WriteString("\n")
	}

	if len(taints) > 0 {
		b.WriteString("### Taint chains\n")

		for _, line := range taints {
			fmt.Fprintf(&b, "- %s\n", line)
		}

		b.WriteString("\n")
	}

	if mainworldNote != "" {
		b.WriteString("### Mainworld habitability\n")
		fmt.Fprintf(&b, "- %s\n\n", mainworldNote)
	}

	return b.String()
}

func collectTidalLocks(u *Universe) []string {
	var out []string

	for body := range u.AllBodies() {
		if !body.HasTidalLock() || body.TidalLock.LockRatio == "" {
			continue
		}

		out = append(out, formatTidalLock(body))
	}

	return out
}

func formatTidalLock(body *Body) string {
	caseLabel := tidalCaseLabel(body.TidalLock.Case)

	suffix := ""
	if body.TidalLock.IsTwilightZone {
		suffix = ", twilight zone"
	}

	return fmt.Sprintf("%s: %s, %s%s", body.Designation, caseLabel, body.TidalLock.LockRatio, suffix)
}

func tidalCaseLabel(c TidalLockCase) string {
	switch c {
	case TidalLockCasePlanetToStar:
		return "planet → star"
	case TidalLockCaseMoonToPlanet:
		return "moon → planet"
	case TidalLockCasePlanetToMoon:
		return "planet → moon"
	}

	return "unlocked" // unreachable when LockRatio != ""
}

func collectColdSnaps(u *Universe) []string {
	// A "cold snap" is a body whose mean is otherwise livable but whose
	// worst-low scenario is dangerous — the bullet a Referee actually
	// wants flagged. Bodies that are always cold (low mean too) aren't
	// snaps; their cold is the normal state and is already shown in the
	// IISS Class II/III temperature column.
	var out []string

	for body := range u.AllBodies() {
		if !body.HasTemperature() {
			continue
		}

		t := body.Temperature
		if t.WorstLowK <= 0 || t.WorstLowK >= notableWorstLowK {
			continue
		}

		if t.MeanK < notableMeanLivableK {
			continue
		}

		out = append(out, fmt.Sprintf("%s: WorstLow %.0f K (mean %.0f K)",
			body.Designation, t.WorstLowK, t.MeanK))
	}

	return out
}

func collectCrushWorlds(u *Universe) []string {
	var out []string

	for body := range u.AllBodies() {
		if !body.HasPhysical() || !body.HasAtmosphere() {
			continue
		}

		g := body.Physical.Gravity

		p := body.Atmosphere.Pressure
		if g <= notableGravityHigh || p <= notablePressureHigh {
			continue
		}

		out = append(out, fmt.Sprintf("%s: gravity %.2f G, pressure %.2f bar",
			body.Designation, g, p))
	}

	return out
}

func collectTaintChains(u *Universe) []string {
	var out []string

	for body := range u.AllBodies() {
		if !body.HasAtmosphere() || len(body.Atmosphere.Taints) < notableTaintChainSize {
			continue
		}

		codes := make([]string, 0, len(body.Atmosphere.Taints))
		for _, t := range body.Atmosphere.Taints {
			codes = append(codes, t.Code)
		}

		out = append(out, fmt.Sprintf("%s: %s", body.Designation, strings.Join(codes, ", ")))
	}

	return out
}

func mainworldHabitabilityNote(u *Universe) string {
	m := u.Detail.Mainworld
	if m == nil || !m.HasHabitability() {
		return ""
	}

	h := m.Habitability

	label := HabitabilityRatingName(h.Rating)
	if h.Notes == "" {
		return fmt.Sprintf("%s — Rating %d/12 (%s)", m.Designation, h.Rating, label)
	}

	return fmt.Sprintf("%s — Rating %d/12 (%s): %s", m.Designation, h.Rating, label, h.Notes)
}
