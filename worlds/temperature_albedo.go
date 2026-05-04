package worlds

import (
	"wbh/roller"
	"wbh/stars"
)

// ComputeAlbedo returns the body's Bond/bolometric albedo per WBH p.110.
// Result clamped to [0.02, 0.98].
//
// Roll budget per body type:
//   - Gas giant: 1 roll (2D base)
//   - Rocky terrestrial (density > 0.4): 1 roll (2D base) + atmosphere modifier
//     (1 roll if atm code active) + hydrographics modifier (1 roll if hyd ≥ 2)
//   - Icy terrestrial up to HZCO+2: same as rocky but different base formula
//   - Icy terrestrial beyond HZCO+2: 1 roll (2D base); if base ≤ 0.4, +1 roll
//     (1D-1 × 0.05 subtract) + atmosphere/hydrographics modifiers
func ComputeAlbedo(r roller.Roller, body *DetailedPlacement, sys stars.System) float64 {
	var albedo float64

	if body.GGClass != NotGasGiant {
		// Gas giant: 0.05 + 2D × 0.05.
		albedo = 0.05 + float64(r.Roll("2D"))*0.05
	} else {
		density := 0.0
		if body.Physical != nil {
			density = body.Physical.Density
		}

		// HZCO in Orbit# — use the group's HZCO when available, otherwise
		// fall back to sys.Primary.HZCO() (also in Orbit#, never Luminosity).
		hzco := sys.Primary.HZCO()
		if len(body.Group.Members) > 0 {
			hzco = body.Group.HZCO()
		}
		beyondIcyLimit := body.Orbit > hzco+2.0

		switch {
		case beyondIcyLimit:
			// Icy beyond HZCO+2: 0.25 + (2D-2) × 0.07.
			albedo = 0.25 + float64(r.Roll("2D")-2)*0.07
			// Footnote p.110: on any result of 0.4 or less, subtract 1D-1 × 0.05
			// to lower toward the 0.02 minimum.
			if albedo <= 0.4 {
				albedo -= float64(r.Roll("1D")-1) * 0.05
			}
		case density > 0.4:
			// Rocky terrestrial: 0.04 + (2D-2) × 0.02.
			albedo = 0.04 + float64(r.Roll("2D")-2)*0.02
		default:
			// Icy terrestrial up to HZCO+2: 0.2 + (2D-3) × 0.05.
			albedo = 0.2 + float64(r.Roll("2D")-3)*0.05
		}

		// Atmosphere modifier — mutually exclusive bands per WBH p.110 table.
		if body.Atmosphere != nil {
			switch body.Atmosphere.Code {
			case 1, 2, 3, 14: // 1–3 or E
				albedo += float64(r.Roll("2D")-3) * 0.01
			case 4, 5, 6, 7, 8, 9:
				albedo += float64(r.Roll("2D")) * 0.01
			case 10, 11, 12, 15, 16, 17: // A–C or F+
				albedo += float64(r.Roll("2D")-2) * 0.05
			case 13: // D
				albedo += float64(r.Roll("2D")) * 0.03
			}
			// Atm 0 and unrecognised codes add no modifier.
		}

		// Hydrographics modifier — mutually exclusive bands per WBH p.110 table.
		if body.Hydrographics != nil {
			hyd := body.Hydrographics.Code
			switch {
			case hyd >= 2 && hyd <= 5:
				albedo += float64(r.Roll("2D")-2) * 0.02
			case hyd >= 6:
				albedo += float64(r.Roll("2D")-4) * 0.03
			}
			// Hyd 0–1 adds no modifier.
		}
	}

	// Clamp [0.02, 0.98] per p.110.
	if albedo < 0.02 {
		albedo = 0.02
	}
	if albedo > 0.98 {
		albedo = 0.98
	}
	return albedo
}
