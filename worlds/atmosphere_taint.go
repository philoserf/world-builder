// Package worlds — atmosphere taint typology per WBH pp.81-90
// (post-3B follow-up: closes Q3-a deferrals).
package worlds

import (
	"wbh/roller"
)

// Taint — one taint or irritant condition per WBH p.82-84.
//
// Code values per WBH p.82 Taint Subtype table:
//
//	L = Low Oxygen
//	R = Radioactivity
//	B = Biologic
//	G = Gas Mix
//	P = Particulates
//	S = Sulphur Compounds
//	H = High Oxygen
//
// Severity (1-9) per WBH p.84 Taint Severity table.
// Persistence (2-9) per WBH p.84 Taint Persistence table.
//
// On atms outside 4-9 (A/B/C/F+), the Taint Subtype table is used for
// "irritants" with the same fields. Renderers distinguish T.S.P (taint)
// from I.S.P (irritant) by atm code, not by Taint type.
type Taint struct {
	Code        string
	Severity    int
	Persistence int
}

// Hazard — Insidious Atmosphere inherent hazard per WBH p.90.
//
// Code values:
//
//	B = Biologic
//	R = Radioactivity
//	G = Gas Mix
//	T = Temperature
//
// Hazards are inherently lethal and constant per WBH p.89; severity and
// persistence are not rolled.
type Hazard struct {
	Code string
}

// HasTaintCode reports whether any Taint in the slice has the given code.
// Used by RollBiomass (Code "B"), RollBiocomplexity (Code "L"), and
// RollCompatibility (any taint present → "otherwise tainted" -2).
func HasTaintCode(taints []Taint, code string) bool {
	for _, t := range taints {
		if t.Code == code {
			return true
		}
	}
	return false
}

// HasAnyTaint reports whether the slice contains at least one Taint.
// Used by RollCompatibility for the "or otherwise tainted" qualifier.
func HasAnyTaint(taints []Taint) bool {
	return len(taints) > 0
}

// taintSubtypeFromTotal maps a 2D+DM total to a Taint Subtype code per
// WBH p.82 Taint Subtype table. Values below 2 clamp to L; values above
// 12 clamp to H.
func taintSubtypeFromTotal(total int) string {
	switch {
	case total <= 2:
		return "L"
	case total == 3:
		return "R"
	case total == 4:
		return "B"
	case total == 5:
		return "G"
	case total == 6:
		return "P"
	case total == 7:
		return "G"
	case total == 8:
		return "S"
	case total == 9:
		return "B"
	case total == 10:
		return "P"
	case total == 11:
		return "R"
	default: // 12+
		return "H"
	}
}

// taintSubtypeAtmDM returns the WBH p.82 atmosphere DM applied to the
// Taint Subtype roll: atm 4 → -2, atm 9 → +2, others → 0.
func taintSubtypeAtmDM(atmCode int) int {
	switch atmCode {
	case 4:
		return -2
	case 9:
		return 2
	}
	return 0
}

// RollTaintSeverity rolls 2D + DMs on WBH p.84 Taint Severity table.
//
// DMs:
//   - Atm C (Insidious, code 12): DM+6
//   - L/H taints: ppO2-specific overrides (p.84 footnote) take precedence —
//     no dice roll is made.
//     For L: severity = 2 if ppO2 ≥ 0.09; 3 if ≥ 0.08; otherwise 8.
//     For H: severity = 2 if ppO2 < 0.6; 7 if < 0.7; otherwise 8.
//
// Returns severity 1-9.
func RollTaintSeverity(r roller.Roller, taintCode string, atmCode int, ppO2 float64) int {
	switch taintCode {
	case "L":
		switch {
		case ppO2 >= 0.09:
			return 2
		case ppO2 >= 0.08:
			return 3
		default:
			return 8
		}
	case "H":
		switch {
		case ppO2 < 0.6:
			return 2
		case ppO2 < 0.7:
			return 7
		default:
			return 8
		}
	}
	roll := r.Roll("2D")
	dm := 0
	if atmCode == 12 {
		dm += 6
	}
	return severityFromTotal(roll + dm)
}

func severityFromTotal(total int) int {
	switch {
	case total <= 4:
		return 1
	case total == 5:
		return 2
	case total == 6:
		return 3
	case total == 7:
		return 4
	case total == 8:
		return 5
	case total == 9:
		return 6
	case total == 10:
		return 7
	case total == 11:
		return 8
	default:
		return 9
	}
}

// RollTaintPersistence rolls 2D + DMs on WBH p.84 Taint Persistence table.
//
// DMs:
//   - Atm C (Insidious, code 12): DM+6
//   - L/H taints: DM+4
//   - Severity ≥ 8: DM+6
//
// Returns persistence 2-9.
func RollTaintPersistence(r roller.Roller, taintCode string, atmCode, severity int) int {
	roll := r.Roll("2D")
	dm := 0
	if taintCode == "L" || taintCode == "H" {
		dm += 4
	}
	if atmCode == 12 {
		dm += 6
	}
	if severity >= 8 {
		dm += 6
	}
	return persistenceFromTotal(roll + dm)
}

func persistenceFromTotal(total int) int {
	switch {
	case total <= 2:
		return 2
	case total == 3:
		return 3
	case total == 4:
		return 4
	case total == 5:
		return 5
	case total == 6:
		return 6
	case total == 7:
		return 7
	case total == 8:
		return 8
	default:
		return 9
	}
}

// RollInsidiousHazard rolls 2D + DM on WBH p.90 Insidious Atmosphere
// Hazard table. Returns hazard code from {B, R, G, T}.
//
// DMs:
//   - Atmosphere is extremely dense: DM+2
//
// Table:
//
//	4-: B (Biologic)
//	5:  R (Radioactivity)
//	6,7: G (Gas Mix)
//	8:  T (Temperature)
//	9:  G
//	10: T
//	11: R
//	12+: T
//
// The "T hazard auto on subtype D/E + reroll for additional hazard"
// rule from p.90 is handled by the runStep5DPrime orchestrator, not here.
func RollInsidiousHazard(r roller.Roller, isExtremelyDense bool) string {
	roll := r.Roll("2D")
	dm := 0
	if isExtremelyDense {
		dm += 2
	}
	return hazardFromTotal(roll + dm)
}

func hazardFromTotal(total int) string {
	switch {
	case total <= 4:
		return "B"
	case total == 5:
		return "R"
	case total == 6, total == 7:
		return "G"
	case total == 8:
		return "T"
	case total == 9:
		return "G"
	case total == 10:
		return "T"
	case total == 11:
		return "R"
	default:
		return "T"
	}
}

// RollTaintSubtype rolls 2D + atm DM on the WBH p.82 Taint Subtype
// table. Applies the L/H suppression rule (treat as G):
//   - When atmCode is outside the 4-9 band (e.g., A/B/C/F+ atms rolling
//     for irritants).
//   - When isSecondOrLater is true (2nd/3rd taint rolls per p.83).
//
// Returns the subtype code letter.
func RollTaintSubtype(r roller.Roller, atmCode int, isSecondOrLater bool) string {
	roll := r.Roll("2D")
	dm := taintSubtypeAtmDM(atmCode)
	code := taintSubtypeFromTotal(roll + dm)

	if (code == "L" || code == "H") && (isSecondOrLater || atmCode < 4 || atmCode > 9) {
		return "G"
	}
	return code
}

// RollAllTaints rolls the full taint profile for a body's atmosphere
// per WBH pp.81-84. Caller passes any pre-seeded taint from
// PromoteOxygenTaint (or nil if the atm wasn't promoted). Pre-seeded
// taints have Severity and Persistence == 0; this function fills them
// using the severity/persistence rolls so callers don't have to special-
// case pre-seeded entries.
//
// Multi-roll behavior per WBH p.83:
//   - Result of 10 (rawRoll + dm == 10) → particulates AND reroll for next taint.
//   - Maximum 3 taint conditions per world.
//   - L/H suppression on 2nd/3rd rolls (handled by RollTaintSubtype).
//
// ppO2 adjustment per WBH p.83 (only on fresh L/H rolls, not pre-seeded):
//   - L: ppO2 -= 1D/100; total pressure unchanged (gas replaced with N₂).
//   - H: ppO2 += 1D/10; same.
//
// Mutates body.Atmosphere.OxygenPartialPressure when ppO2 adjustment fires.
//
// Returns the populated []Taint slice for assignment to body.Atmosphere.Taints.
func RollAllTaints(r roller.Roller, body *DetailedPlacement, preseeded *Taint) []Taint {
	if body == nil || body.Atmosphere == nil {
		return nil
	}
	atmCode := body.Atmosphere.Code
	taints := make([]Taint, 0, 3)

	// Slot 1: pre-seeded if present. Fill severity + persistence.
	if preseeded != nil {
		ppO2 := body.Atmosphere.OxygenPartialPressure
		sev := RollTaintSeverity(r, preseeded.Code, atmCode, ppO2)
		pers := RollTaintPersistence(r, preseeded.Code, atmCode, sev)
		taints = append(taints, Taint{Code: preseeded.Code, Severity: sev, Persistence: pers})
	}

	// Multi-roll loop. isSecondOrLater = anything appended already.
	for len(taints) < 3 {
		isSecondOrLater := len(taints) > 0
		rawRoll := r.Roll("2D")
		dm := taintSubtypeAtmDM(atmCode)
		total := rawRoll + dm
		code := taintSubtypeFromTotal(total)
		if (code == "L" || code == "H") && (isSecondOrLater || atmCode < 4 || atmCode > 9) {
			code = "G"
		}

		// Fresh L/H only (no pre-seed, first roll): adjust ppO2.
		if (code == "L" || code == "H") && !isSecondOrLater && preseeded == nil {
			adjust := r.Roll("1D")
			if code == "L" {
				body.Atmosphere.OxygenPartialPressure -= float64(adjust) / 100.0
			} else {
				body.Atmosphere.OxygenPartialPressure += float64(adjust) / 10.0
			}
		}

		ppO2 := body.Atmosphere.OxygenPartialPressure
		sev := RollTaintSeverity(r, code, atmCode, ppO2)
		pers := RollTaintPersistence(r, code, atmCode, sev)
		taints = append(taints, Taint{Code: code, Severity: sev, Persistence: pers})

		// Reroll trigger: rawRoll + dm == 10.
		if total != 10 {
			break
		}
	}
	return taints
}
