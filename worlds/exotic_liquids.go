package worlds

import "math"

// ExoticLiquid is one row of the WBH p.102 Possible Exotic Liquids table.
type ExoticLiquid struct {
	Code      string // "H2O", "CH4", "NH3", etc.
	MeltingK  float64
	BoilingK  float64
	Abundance int // Relative Abundance, 1..100
}

// PossibleExoticLiquids — WBH p.102 table, 15 entries ordered by boiling point.
var PossibleExoticLiquids = []ExoticLiquid{
	{"F2", 53, 85, 2},
	{"O2", 54, 90, 50},
	{"CH4", 91, 113, 70},
	{"C2H6", 90, 184, 70},
	{"Cl2", 171, 239, 1},
	{"NH3", 195, 240, 30},
	{"SO2", 201, 263, 20},
	{"HF", 190, 293, 2},
	{"HCN", 260, 299, 30},
	{"HCl", 247, 321, 1},
	{"H2O", 273, 373, 100},
	{"CH2O2", 281, 374, 15},
	{"CH3NO", 275, 483, 15},
	{"H2CO3", 193, 607, 20},
	{"H2SO4", 388, 718, 20},
}

// isExoticAtmCode reports whether atmCode requires exotic-liquid selection
// (Atm A=10, B=11, C=12, F=15 per p.102).
func isExoticAtmCode(atmCode int) bool {
	return atmCode == 10 || atmCode == 11 || atmCode == 12 || atmCode == 15
}

// SelectExoticLiquid returns the dominant liquid for a body with exotic
// atmosphere (Atm A-C/F: codes 10, 11, 12, 15) and non-zero hydrographics
// at the given mean temperature.
//
// Deterministic: among molecules where MeltingK ≤ meanK ≤ BoilingK, returns
// the highest-Abundance candidate. Ties broken by lower BoilingK (more
// "stable" in range). Returns "" if no candidate fits or atmCode is not exotic.
func SelectExoticLiquid(meanK float64, atmCode int) string {
	if !isExoticAtmCode(atmCode) {
		return ""
	}

	bestCode := ""
	bestAbundance := -1
	bestBoiling := math.Inf(1) // sentinel makes tie-break direction self-evident

	for _, l := range PossibleExoticLiquids {
		if meanK < l.MeltingK || meanK > l.BoilingK {
			continue
		}

		if l.Abundance > bestAbundance ||
			(l.Abundance == bestAbundance && l.BoilingK < bestBoiling) {
			bestCode = l.Code
			bestAbundance = l.Abundance
			bestBoiling = l.BoilingK
		}
	}

	return bestCode
}
