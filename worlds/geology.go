// Package worlds — geology (seismic + GG residual heat + tectonic plates)
// per WBH pp.125-127 (sub-project 3B-geology).
package worlds

import "math"

// Geology — seismic activity and inherent temperature contribution per
// WBH pp.125-127. Populated by Step 5E for any non-empty, non-belt body.
//
// Conditional applicability:
//   - Terrestrials: ResidualSeismicStress, TidalStressFactor,
//     TidalHeatingFactor, TotalSeismicStress, TectonicPlates populated;
//     InherentTemperatureK == float64(TotalSeismicStress).
//   - Gas giants: only InherentTemperatureK populated (from the GG
//     residual heat formula); seismic fields and TectonicPlates remain 0.
//   - Belts (Size 0): geology not generated; dp.Geology stays nil.
type Geology struct {
	// Terrestrial-only seismic factors (0 for gas giants).
	ResidualSeismicStress int // (Size − Age + DMs)² per WBH p.125
	TidalStressFactor     int // Σ tidal effects ÷ 10 per WBH p.126
	TidalHeatingFactor    int // primary-mass formula ÷ 3000 per WBH p.126
	TotalSeismicStress    int // sum of the three above

	// Terrestrial-only tectonic plate count. Zero if prerequisites failed
	// (TSS ≤ 0 or Hydro < 1) or if the dice roll produced ≤ 1.
	TectonicPlates int

	// Inherent temperature addition in Kelvin, used in the temperature
	// recompute equation: New T = ⁴√(T⁴ + InherentTemperatureK⁴).
	// For terrestrials: equals float64(TotalSeismicStress).
	// For gas giants: equals 80 × ⁴√(MassEarth) ÷ √(AgeGyr), zero if
	// the formula produces a negligible value (< 1K).
	InherentTemperatureK float64
}

// ComputeResidualSeismicStress computes the Residual Seismic Stress component
// per WBH p.125: floor(Size − Age(Gyr) + DMs)², clamped so values < 1
// before squaring yield 0 (e.g. −1.5 → 0, not 2.25).
//
// DMs applied:
//   - isMoon is true → +1
//   - body has Size-1+ moons (counted) → +1 per moon, max +12
//   - body.Physical.Density > 1.0 → +2
//   - body.Physical.Density < 0.5 → −1
//
// ageGyr is the system age in billions of years.
func ComputeResidualSeismicStress(body *DetailedPlacement, ageGyr float64, isMoon bool) int {
	if body == nil {
		return 0
	}
	size := SizeAsInt(body.SizeCode)
	dm := 0
	if isMoon {
		dm++
	}
	moonDM := 0
	for _, m := range body.Moons {
		if SizeAsInt(m.SizeCode) >= 1 {
			moonDM++
		}
	}
	dm += min(moonDM, 12)
	if body.Physical != nil {
		switch {
		case body.Physical.Density > 1.0:
			dm += 2
		case body.Physical.Density < 0.5:
			dm--
		}
	}
	inner := math.Floor(float64(size) - ageGyr + float64(dm))
	if inner < 1 {
		return 0
	}
	return int(inner) * int(inner)
}
