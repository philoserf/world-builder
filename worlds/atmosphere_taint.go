// Package worlds — atmosphere taint typology per WBH pp.81-90
// (post-3B follow-up: closes Q3-a deferrals).
package worlds

import "wbh/roller"

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
