// Package worlds — temperature characteristics per WBH pp.108-126.
package worlds

import (
	"math"
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
