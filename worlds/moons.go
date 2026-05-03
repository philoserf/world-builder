package worlds

import (
	"fmt"

	"wbh/roller"
)

// Moon — one significant moon. Insignificant moons (free-form Referee
// fiat per WBH p.58) are out of scope for 2C.
type Moon struct {
	Designation string // "Aab IV a", ... — assigned by AssignMoonDesignations
	SizeCode    SizeCode
	DiameterKm  float64

	// Set when the moon is itself gas-giant-sized (rare, GG Special row, p.57):
	GGClass        GasGiantClass
	GGDiameterCode string
	DiameterEarth  float64
	MassEarth      float64
}

// ParentInfo describes a moon's parent body. Only one of (terrestrial
// SizeCode) or (IsGasGiant + GGClass) should be populated.
type ParentInfo struct {
	IsGasGiant bool
	GGClass    GasGiantClass // NotGasGiant for terrestrial parents
	SizeCode   SizeCode      // for terrestrial parents (e.g. "5", "A")
}

// CountMoons rolls the WBH p.55 Significant Moon Quantity table:
//
//	Size 1-2 → 1D-5    Size 3-9 → 2D-8     Size A-F → 2D-6
//	Small GG → 3D-7    Medium/Large GG → 4D-6
//
// dms is the per-die DM (0 or -1) per the p.55 conditions:
//   - Planet's Orbit# < 1.0
//   - Planet is in orbital slot adjacent to a companion
//   - Planet's slot adjacent to Close/Near unavailability range
//   - Planet in adjacent slot to outermost Close/Near/Far range
//
// Per the book: only ONE DM applies regardless of how many conditions
// are met. Caller is responsible for evaluating conditions and passing
// dms = 0 or dms = -1.
//
// Negative result → returns 0 (no significant moons). Exactly 0 →
// returns 0 (caller treats as a planetary ring per p.55).
func CountMoons(r roller.Roller, parent ParentInfo, dms int) (int, error) {
	// Sub-1-Size terrestrial parents (SizeCode "0", "R", "S") cannot
	// host significant moons per WBH p.55 — the Quantity table starts
	// at Size 1-2. Short-circuit before consuming a die.
	if !parent.IsGasGiant {
		switch parent.SizeCode {
		case "0", "R", "S":
			return 0, nil
		}
	}

	notation, base, dieCount, err := moonQuantityFormula(parent)
	if err != nil {
		return 0, err
	}

	rawSum := r.Roll(notation)
	// dms is per-die: each of the dieCount dice gets dms applied.
	adjusted := rawSum + dms*dieCount
	result := adjusted + base
	if result < 0 {
		return 0, nil
	}
	return result, nil
}

// moonQuantityFormula returns the dice notation, additive base
// (negative because the book writes "1D-5", "2D-8", etc.), and the
// die count for the per-die DM application.
func moonQuantityFormula(p ParentInfo) (notation string, base, dieCount int, err error) {
	if p.IsGasGiant {
		switch p.GGClass {
		case GasGiantSmall:
			return "3D", -7, 3, nil
		case GasGiantMedium, GasGiantLarge:
			return "4D", -6, 4, nil
		default:
			return "", 0, 0, fmt.Errorf("worlds: CountMoons: unknown GGClass %v", p.GGClass)
		}
	}
	n := nForSizeCode(p.SizeCode)
	switch {
	case n >= 1 && n <= 2:
		return "1D", -5, 1, nil
	case n >= 3 && n <= 9:
		return "2D", -8, 2, nil
	case n >= 10 && n <= 15: // A-F
		return "2D", -6, 2, nil
	default:
		return "", 0, 0, fmt.Errorf("worlds: CountMoons: unsupported parent SizeCode %q", p.SizeCode)
	}
}
