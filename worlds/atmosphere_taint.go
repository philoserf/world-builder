// Package worlds — atmosphere taint typology per WBH pp.81-90
// (post-3B follow-up: closes Q3-a deferrals).
package worlds

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
