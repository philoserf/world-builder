// Package worlds — temperature characteristics per WBH pp.108-126.
package worlds

import (
	"math"

	"wbh/roller"
	"wbh/stars"
)

// MeanTemperatureK computes the world's mean temperature in Kelvin per WBH p.111:
//
//	T = 279 × ⁴√(L × (1 - A) × (1 + G) / d²)
//
// Per p.111 thumb-rule-two: "(1+G) should never use a factor of more than
// luminosity × 1.999 and a low no less than luminosity × 0.001". Clamp (1+G)
// to [0.001, 1.999].
func MeanTemperatureK(luminosity, albedo, greenhouse, au float64) float64 {
	if au <= 0 {
		return 0
	}
	gFactor := 1 + greenhouse
	if gFactor < 0.001 {
		gFactor = 0.001
	}
	if gFactor > 1.999 {
		gFactor = 1.999
	}
	core := luminosity * (1 - albedo) * gFactor / (au * au)
	if core <= 0 {
		return 0
	}
	return 279.0 * math.Pow(core, 0.25)
}

// CombineTemperatures combines independent temperature sources per WBH p.109:
//
//	T_total = ⁴√(T₁⁴ + T₂⁴ + …)
//
// Used to add a moon's parent-body IR contribution to its stellar temperature
// (p.125-126), and to combine separate stellar groups for a body orbiting a
// barycenter with multiple non-close-binary stars.
func CombineTemperatures(temps ...float64) float64 {
	if len(temps) == 0 {
		return 0
	}
	if len(temps) == 1 {
		return temps[0]
	}
	sumOf4ths := 0.0
	for _, t := range temps {
		sumOf4ths += math.Pow(t, 4)
	}
	return math.Pow(sumOf4ths, 0.25)
}

// basicMeanTemperatureK maps modified roll → Kelvin per WBH p.109 table.
var basicMeanTemperatureK = map[int]float64{
	0: 178, 1: 198, 2: 218, 3: 238, 4: 263, 5: 278, 6: 283, 7: 288,
	8: 293, 9: 298, 10: 313, 11: 338, 12: 388,
}

// BasicTemperatureRoll rolls 2D + DMs and returns the modified roll plus the
// Kelvin value from the WBH p.109 Basic Mean Temperature table.
//
// DMs (p.109):
//   - Atmosphere DM from p.47 table (via HZRegionAtmosphereDM)
//   - +4 +1 per 0.5 Orbit# below HZCO-1 if Orbit# < HZCO-1
//   - -4 -1 per 0.5 Orbit# above HZCO+1 if Orbit# > HZCO+1
//
// Modified roll above 12: per book "another +50° per result above 12".
// Modified roll below 0: per book "another -5° per result below 0", with
// special recompute "as 1D+5" if value would be < 10K.
func BasicTemperatureRoll(r roller.Roller, body *DetailedPlacement, sys stars.System) (modifiedRoll int, kelvin float64) {
	raw := r.Roll("2D")
	dm := 0

	if body.Atmosphere != nil {
		dm += HZRegionAtmosphereDM(body.Atmosphere.Code)
	}

	hzco := sys.Primary.HZCO()
	if len(body.Group.Members) > 0 {
		hzco = body.Group.HZCO()
	}
	orbit := body.Orbit
	if orbit < hzco-1 {
		dm += 4 + int(math.Floor((hzco-1-orbit)/0.5))
	} else if orbit > hzco+1 {
		dm -= 4 + int(math.Floor((orbit-(hzco+1))/0.5))
	}

	mod := raw + dm

	switch {
	case mod >= 13:
		// Each step above 12 adds +50K to the 388K table top.
		kelvin = 388 + float64(mod-12)*50
	case mod >= 0 && mod <= 12:
		kelvin = basicMeanTemperatureK[mod]
	default:
		// mod < 0 → 178K + 5K per step below 0.
		kelvin = 178 + float64(mod)*5 // mod negative → subtracts
		if kelvin < 10 {
			// Recompute as 1D+5 per p.109 footnote.
			kelvin = float64(r.Roll("1D") + 5)
		}
		if kelvin < 3 {
			kelvin = 3
		}
	}

	return mod, kelvin
}
