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

// ComputeTidalStressFactor per WBH p.126: floor(ΣTidalEffects / 10).
// Reads body.TidalEffects.Total (metres, populated in Step 5B.5).
// Returns 0 if body or TidalEffects is nil.
func ComputeTidalStressFactor(body *DetailedPlacement) int {
	if body == nil || body.TidalEffects == nil {
		return 0
	}
	return int(math.Floor(body.TidalEffects.Total / 10.0))
}

// TidalHeatingInputs are the inputs to the WBH p.126 tidal-heating-factor
// formula. The book specifies units explicitly: Distance in millions of
// kilometres (Mkm) and Period in days. Caller must convert AU→Mkm
// (multiply by 149.6) and hours→days (divide by 24) for planet bodies;
// moons natively use Mkm and days after dividing m.OrbitKm by 1_000_000
// and m.PeriodHours by 24.
type TidalHeatingInputs struct {
	PrimaryMassEarth float64 // primary body's mass in Earth masses
	SizeN            int     // body size 0-15 (numeric)
	Eccentricity     float64 // body's orbital eccentricity
	DistanceMkm      float64 // distance from primary in millions of km
	PeriodDays       float64 // orbital period around primary, in days
	WorldMassEarth   float64 // body's mass in Earth masses
}

// ComputeTidalHeatingFactor per WBH p.126:
//
//	(PrimaryMass⊕)² × SizeN⁵ × ecc²
//	─────────────────────────────────────────────────────────────────
//	3,000 × DistanceMkm⁵ × PeriodDays × WorldMass⊕
//
// Floor the result; values < 1 are treated as 0. Returns 0 if any divisor
// component (DistanceMkm, PeriodDays, WorldMassEarth) is zero.
//
// Worked references from the book (WBH p.126):
//   - Io ≈ 101 (book's calibration point)
//   - Enceladus ≈ 11
//   - Zed Prime ≈ 14
func ComputeTidalHeatingFactor(in TidalHeatingInputs) int {
	if in.DistanceMkm == 0 || in.PeriodDays == 0 || in.WorldMassEarth == 0 {
		return 0
	}
	num := in.PrimaryMassEarth * in.PrimaryMassEarth *
		math.Pow(float64(in.SizeN), 5) *
		in.Eccentricity * in.Eccentricity
	den := 3000.0 * math.Pow(in.DistanceMkm, 5) * in.PeriodDays * in.WorldMassEarth
	v := num / den
	if v < 1 {
		return 0
	}
	return int(math.Floor(v))
}

// ComputeGGResidualHeat per WBH p.125 sidebar:
//
//	T(K) = 80 × ⁴√(MassEarth) ÷ √(AgeGyr)
//
// Returns 0 if mass ≤ 0 or age ≤ 0; returns 0 if the formula produces < 1K
// (negligible). Used for gas giants only — terrestrial bodies do not
// receive this contribution.
//
// Worked: Zed Prime's GG (MassEarth=1200, AgeGyr=6.336) ≈ 187K.
func ComputeGGResidualHeat(massEarth, ageGyr float64) float64 {
	if massEarth <= 0 || ageGyr <= 0 {
		return 0
	}
	v := 80.0 * math.Pow(massEarth, 0.25) / math.Sqrt(ageGyr)
	if v < 1 {
		return 0
	}
	return v
}
