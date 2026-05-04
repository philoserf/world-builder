package worlds

import (
	"wbh/roller"
)

// Atmosphere holds the UWP atmosphere code and WBH refinement fields (pp.79-95).
// Pressure, OxygenPartialPressure, ScaleHeight, Subtype, and Profile are populated
// by later tasks (T10-T12).
type Atmosphere struct {
	Code                  int
	Subtype               string
	Pressure              float64
	OxygenPartialPressure float64
	ScaleHeight           float64
	Profile               AtmosphereProfile
}

// AtmosphereProfile holds the gas-mix composition detail populated in Task 12.
type AtmosphereProfile struct {
	TempRange string
	Gases     []GasFraction
	Shorthand string
}

// GasFraction is one constituent gas in an AtmosphereProfile.
type GasFraction struct {
	Name      string
	PercentBP int
}

// TempRange is a provisional temperature class derived from a body's HZCO offset
// per WBH pp.94-98.
type TempRange int

// Temperature range constants keyed to HZCO offset bands per WBH pp.94-98.
const (
	TempBoiling   TempRange = iota // offset ≤ -2.01  (mean ≥ 453 K)
	TempHot                        // offset -2.0..-1.01 (353-453 K)
	TempTemperate                  // offset -1.0..+1.0  (273-353 K)
	TempCold                       // offset +1.01..+3.0 (123-273 K)
	TempFrozen                     // offset ≥ +3.01     (≤ 123 K)
)

// HZCOOffsetToTempRange converts a body's orbital position relative to its star's
// habitable-zone centre orbit (HZCO) into a temperature class per WBH pp.94-98.
func HZCOOffsetToTempRange(orbitNumber, hzco float64) TempRange {
	offset := orbitNumber - hzco
	switch {
	case offset <= -2.01:
		return TempBoiling
	case offset <= -1.01:
		return TempHot
	case offset <= 1.0:
		return TempTemperate
	case offset <= 3.0:
		return TempCold
	default:
		return TempFrozen
	}
}

// atmosphereLabels maps WBH p.79 Atmosphere Codes.
var atmosphereLabels = map[int]string{
	0:  "None",
	1:  "Trace",
	2:  "Very Thin, Tainted",
	3:  "Very Thin",
	4:  "Thin, Tainted",
	5:  "Thin",
	6:  "Standard",
	7:  "Standard, Tainted",
	8:  "Dense",
	9:  "Dense, Tainted",
	10: "Exotic",
	11: "Corrosive",
	12: "Insidious",
	13: "Very Dense",
	14: "Low",
	15: "Unusual",
	16: "Gas, Helium",
	17: "Gas, Hydrogen",
}

// AtmosphereCompositionLabel returns the human-readable label for a WBH p.79
// atmosphere code. Returns "" for unknown codes.
func AtmosphereCompositionLabel(code int) string {
	return atmosphereLabels[code]
}

// SizeAsInt converts a SizeCode to its integer equivalent for arithmetic.
// "S", "R", "", "0" → 0; "1"-"9" → 1-9; "A"-"F" → 10-15.
func SizeAsInt(s SizeCode) int {
	switch s {
	case "", "0", "S", "R":
		return 0
	}
	if len(s) != 1 {
		return 0
	}
	ch := s[0]
	switch {
	case ch >= '1' && ch <= '9':
		return int(ch - '0')
	case ch >= 'A' && ch <= 'F':
		return int(ch-'A') + 10
	}
	return 0
}

// RollAtmoCode rolls the unified WBH atmosphere digit formula: 2D-7+Size.
//
// Sizes 0, 1, and S return automatic atmo code 0 without consuming a roll.
// Results below 0 are clamped to 0.
//
// The third argument (hzcoOffset) is reserved for non-HZ Hot/Cold table column
// selection in Task 11; it does not affect the roll formula here.
func RollAtmoCode(r roller.Roller, sizeCode SizeCode, _ float64) (int, error) {
	if sizeCode == "0" || sizeCode == "1" || sizeCode == "S" {
		return 0, nil
	}
	roll := r.Roll("2D")
	code := roll - 7 + SizeAsInt(sizeCode)
	if code < 0 {
		code = 0
	}
	return code, nil
}
