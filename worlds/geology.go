// Package worlds — geology (seismic + GG residual heat + tectonic plates)
// per WBH pp.125-127 (sub-project 3B-geology).
package worlds

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
