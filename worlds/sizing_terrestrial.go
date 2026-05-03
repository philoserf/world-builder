package worlds

import (
	"fmt"

	"wbh/roller"
)

// SizeCode is the WBH terrestrial Size character (pp. 54, 56):
//
//	"0" — planetoid (0 km)
//	"R" — ring (0 km)
//	"S" — small body (600 km)
//	"1"-"9", "A"-"F" — Sizes 1-15 (1600 km × Size code)
//
// Empty string means "not a size-having body" (belt or empty slot).
type SizeCode string

// TerrestrialSize is the result of RollTerrestrialSize.
type TerrestrialSize struct {
	SizeCode   SizeCode
	DiameterKm float64
}

// basicTerrestrialDiameterTable is the WBH p.54 Basic Terrestrial
// World Size table (km per Size code).
var basicTerrestrialDiameterTable = map[SizeCode]float64{
	"0": 0, "R": 0, "S": 600,
	"1": 1600, "2": 3200, "3": 4800, "4": 6400, "5": 8000,
	"6": 9600, "7": 11200, "8": 12800, "9": 14400,
	"A": 16000, "B": 17600, "C": 19200, "D": 20800, "E": 22400, "F": 24000,
}

// BasicTerrestrialDiameter returns the diameter in km for a SizeCode
// per WBH p.54. Returns 0 for unknown codes (callers should validate
// SizeCode comes from a known source).
func BasicTerrestrialDiameter(code SizeCode) float64 {
	return basicTerrestrialDiameterTable[code]
}

// nForSizeCode (the inverse of sizeCodeForN) is intentionally deferred
// to T8 (SizeMoon), where it is first required. Adding it here would
// fail golangci-lint's unused-function check.

// sizeCodeForN converts an integer Size 0-15 to its hex code (1=1, 9=9, 10=A, ..., 15=F).
// Used internally by RollTerrestrialSize and SizeMoon.
func sizeCodeForN(n int) SizeCode {
	if n < 0 {
		return "0"
	}
	if n > 15 {
		n = 15
	}
	switch {
	case n < 10:
		return SizeCode(fmt.Sprintf("%d", n))
	default:
		return SizeCode(fmt.Sprintf("%c", 'A'+n-10))
	}
}

// RollTerrestrialSize rolls the WBH p.54 Terrestrial World Sizing
// procedure: a 1D selector chooses one of three second-roll formulas:
//
//	1-2 → second roll 1D            (range 1-6)
//	3-4 → second roll 2D            (range 2-C/12)
//	5-6 → second roll 2D+3          (range 5-F/15; clamped to F)
//
// Returns the resulting SizeCode and its book diameter in km.
func RollTerrestrialSize(r roller.Roller) (TerrestrialSize, error) {
	selector := r.Roll("1D")
	if selector < 1 || selector > 6 {
		return TerrestrialSize{}, fmt.Errorf("worlds: terrestrial size selector out of range: %d", selector)
	}

	var n int
	switch {
	case selector <= 2:
		n = r.Roll("1D")
	case selector <= 4:
		n = r.Roll("2D")
	default: // 5-6
		n = r.Roll("2D") + 3
	}
	if n > 15 {
		n = 15
	}
	if n < 1 {
		n = 1
	}
	code := sizeCodeForN(n)
	return TerrestrialSize{
		SizeCode:   code,
		DiameterKm: basicTerrestrialDiameterTable[code],
	}, nil
}
