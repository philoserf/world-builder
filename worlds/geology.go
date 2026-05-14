// Package worlds — geology (seismic + GG residual heat + tectonic plates)
// per WBH pp.125-127 (sub-project 3B-geology).
package worlds

import (
	"math"

	"github.com/philoserf/world-builder/roller"
)

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
func ComputeResidualSeismicStress(body *Body, ageGyr float64, isMoon bool) int {
	if body == nil {
		return 0
	}
	size := SizeAsInt(body.SizeCode)
	dm := 0
	if isMoon {
		dm++
	}
	moonDM := 0
	for _, m := range body.Children {
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
func ComputeTidalStressFactor(body *Body) int {
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

// ApplyInherentTempAddition mutates each populated temperature field on
// temp via the WBH p.125 addition equation:
//
//	NewT = ⁴√(OldT⁴ + addedK⁴)
//
// Idempotent in shape — same equation applied to every field. Safe on nil
// temp or zero addedK (no-op).
//
// Fields touched: MeanK, HighK, LowK, BasicK, WorstHighK, WorstLowK.
// When IsTwilight is true, also: TwilightK, BrightSideK, DarkSideK.
//
// Equation inputs (Luminosity, Albedo, GreenhouseFactor, AU, ScaleHeight)
// and variance components (AxialTiltFactor, RotationFactor, etc.) are
// NOT touched — those are inputs, not output temperatures.
//
// A field with value 0 is treated as "not populated" and skipped.
func ApplyInherentTempAddition(temp *Temperature, addedK float64) {
	if temp == nil || addedK == 0 {
		return
	}
	addPow4 := math.Pow(addedK, 4)
	add := func(v *float64) {
		if *v == 0 {
			return
		}
		*v = math.Pow(math.Pow(*v, 4)+addPow4, 0.25)
	}
	add(&temp.MeanK)
	add(&temp.HighK)
	add(&temp.LowK)
	add(&temp.BasicK)
	add(&temp.WorstHighK)
	add(&temp.WorstLowK)
	if temp.IsTwilight {
		add(&temp.TwilightK)
		add(&temp.BrightSideK)
		add(&temp.DarkSideK)
	}
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

// RollTectonicPlates per WBH p.127:
//
//	Plates = Size + Hydrographics − 2D + DMs
//
// Prerequisites: tss > 0 AND body.Hydrographics.Code ≥ 1. If either
// fails, returns 0 without consuming dice.
//
// DMs: tss in [10, 100] → +1; tss > 100 → +2.
//
// If the rolled result ≤ 1, returns 0 (no tectonic activity).
//
// Worked: Zed Prime (S=5, Hydro=6, TSS=17, 2D=8) → 5 + 6 − 8 + 1 = 4.
func RollTectonicPlates(r roller.Roller, body *Body, tss int) int {
	if body == nil || body.Hydrographics == nil {
		return 0
	}
	if tss <= 0 || body.Hydrographics.Code < 1 {
		return 0
	}
	dm := 0
	switch {
	case tss > 100:
		dm = 2
	case tss >= 10:
		dm = 1
	}
	roll := r.Roll("2D")
	result := SizeAsInt(body.SizeCode) + body.Hydrographics.Code - roll + dm
	if result <= 1 {
		return 0
	}
	return result
}

// TotalSeismicStressLabel returns a banded label for total seismic stress
// anchored on WBH p.126 prose: "<1 = essentially geologically dead" (book
// quote), "10-100 = active (DM+1 plates)" (book band), ">100 = at least one
// ongoing volcanic eruption + near-constant earthquake activity" (book quote).
func TotalSeismicStressLabel(s int) string {
	switch {
	case s <= 0:
		return "Geologically dead" // book quote
	case s < 10:
		return "Quiet" // interpolated (between "dead" and "active")
	case s <= 100:
		return "Active" // book: DM+1 band, "tectonic activity"
	default: // > 100
		return "Hyperactive (constant eruptions)" // book quote (paraphrased)
	}
}

// TidalHeatingFactorLabel returns a banded label for tidal heating factor
// anchored on WBH p.126 reference points: "ignore values less than 1",
// "Enceladus ≈ 11", "Io ≈ 101".
func TidalHeatingFactorLabel(f int) string {
	switch {
	case f < 1:
		return "Negligible" // book: "ignore values less than 1"
	case f <= 10:
		return "Mild" // interpolated (below Enceladus)
	case f <= 100:
		return "Significant (Enceladus-class)" // book anchor
	default: // > 100
		return "Extreme (Io-class)" // book anchor
	}
}
