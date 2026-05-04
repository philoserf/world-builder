package worlds

import (
	"math"

	"wbh/roller"
)

// ComputeGreenhouseFactor returns the world's effective greenhouse factor
// at mean baseline altitude per WBH p.110.
//
// Initial Greenhouse Factor = 0.5 × √(pressure_bar). Then a modifier is
// applied based on atmosphere code:
//   - Atm 1-9 or D or E: + 3D × 0.01
//   - Atm A or F:        × max(1D-1, 0.5) factor
//   - Atm B, C, G, H:    1D=1-5 → × that result; 1D=6 → × 3D
//   - Atm 0 (vacuum):    G = 0 (per p.110 "Vacuum worlds have a greenhouse factor of 0 by definition")
//
// Returned value is the unclamped G; (1+G) clamping to [0.001, 1.999] per
// WBH p.111 thumb-rule-two is applied at MeanTemperatureK.
func ComputeGreenhouseFactor(r roller.Roller, atm *Atmosphere) float64 {
	if atm == nil || atm.Code == 0 {
		return 0
	}

	initial := 0.5 * math.Sqrt(atm.Pressure)
	var g float64

	switch atm.Code {
	case 1, 2, 3, 4, 5, 6, 7, 8, 9, 13, 14: // 1-9 or D (13) or E (14)
		g = initial + float64(r.Roll("3D"))*0.01
	case 10, 15: // A (10) or F (15)
		factor := float64(r.Roll("1D") - 1)
		if factor < 0.5 {
			factor = 0.5
		}
		g = initial * factor
	case 11, 12, 16, 17: // B, C, G, H
		first := r.Roll("1D")
		if first == 6 {
			g = initial * float64(r.Roll("3D"))
		} else {
			g = initial * float64(first)
		}
	default:
		// Codes outside the table: leave at initial (defensive).
		g = initial
	}

	return g
}
