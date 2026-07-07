package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
)

// GasGiantClass identifies the WBH p.55 gas-giant size category.
type GasGiantClass int

const (
	// NotGasGiant is the zero value; the body is not a gas giant.
	NotGasGiant GasGiantClass = iota
	// GasGiantSmall is a GS Neptune-analogue (D3+D3 diameter, 2-6⊕).
	GasGiantSmall
	// GasGiantMedium is a GM Jupiter-analogue (1D+6 diameter, 7-12⊕; WBH p.55 prints 6-12⊕ but the formula minimum is 7).
	GasGiantMedium
	// GasGiantLarge is a GL Superjovian (2D+6 diameter, 8-18⊕).
	GasGiantLarge
)

// GasGiantSize is the result of RollGasGiantSize.
type GasGiantSize struct {
	Class         GasGiantClass
	DiameterCode  string  // "2"-"F" (or "G"/"H"/"J" for Large GGs ≥16⊕ via eHex)
	DiameterEarth float64 // in Terra diameters (Size 8 = 1.0)
	MassEarth     float64 // in Terra masses
}

// gasGiantDiameterCode converts an integer diameter (Terra diameters) to
// its eHex code per WBH p.55: 2-9 → "2"-"9", 10-15 → "A"-"F",
// 16 → "G", 17 → "H", 18 → "J" (skips "I" per Traveller eHex convention).
func gasGiantDiameterCode(n int) string {
	// Safety-net clamps: unreachable given the current dice formulas
	// (D3+D3, 1D+6, 2D+6 all have min ≥ 2 and max ≤ 18), kept only as
	// defense against future formula edits.
	if n < 2 {
		n = 2
	}
	if n > 18 {
		n = 18
	}
	switch {
	case n < 10:
		return fmt.Sprintf("%d", n)
	case n <= 15:
		return string(rune('A' + n - 10))
	case n == 16:
		return "G"
	case n == 17:
		return "H"
	default: // 18
		return "J"
	}
}

// RollGasGiantSize implements the WBH p.55 procedure:
//
//  1. Roll 1D + dms; the result selects the size category:
//     2-  → Small  (GS, D3+D3 diameter, 5×(1D+1) mass)
//     3-4 → Medium (GM, 1D+6 diameter, 20×(3D-1) mass)
//     5+  → Large  (GL, 2D+6 diameter, D3×50×(3D+4) mass)
//
//  2. Roll the second-roll diameter formula for the chosen class.
//
//  3. Roll the third-roll mass formula. For Large GGs whose initial
//     mass ≥3,000⊕ (3D third-roll ≥15), substitute mass = 4000 - 200×(2D-2).
//
// Caller-supplied dms accumulates per WBH p.55:
//   - Brown Dwarf primary, M-V star, or any Class VI star: DM-1
//   - System Spread < 0.1: DM-1
func RollGasGiantSize(r roller.Roller, dms int) (GasGiantSize, error) {
	selectorRaw := r.Roll("1D")
	selector := selectorRaw + dms

	var class GasGiantClass
	switch {
	case selector <= 2:
		class = GasGiantSmall
	case selector <= 4:
		class = GasGiantMedium
	default: // 5+
		class = GasGiantLarge
	}

	var diameter int
	switch class {
	case GasGiantSmall:
		// D3+D3 is two distinct rolls; each scripted separately.
		diameter = r.Roll("D3") + r.Roll("D3") // 2-6
	case GasGiantMedium:
		diameter = r.Roll("1D+6") // 7-12
	case GasGiantLarge:
		diameter = r.Roll("2D+6") // 8-18
	}

	var mass float64
	switch class {
	case GasGiantSmall:
		mass = float64(5 * (r.Roll("1D") + 1)) // 10-35
	case GasGiantMedium:
		threeD := r.Roll("3D")            // 3-18
		mass = float64(20 * (threeD - 1)) // 40-340
	case GasGiantLarge:
		d3 := r.Roll("D3")
		threeD := r.Roll("3D")                 // 3-18
		mass = float64(d3 * 50 * (threeD + 4)) // 350-4000+
		if mass >= 3000 {
			// WBH p.55 footnote: initial mass ≥3,000⊕ → roll 2D-2,
			// substitute mass = 4000 - 200 × (2D-2).
			twoD := r.Roll("2D")
			mass = float64(4000 - 200*(twoD-2))
		}
	}

	return GasGiantSize{
		Class:         class,
		DiameterCode:  gasGiantDiameterCode(diameter),
		DiameterEarth: float64(diameter),
		MassEarth:     mass,
	}, nil
}
