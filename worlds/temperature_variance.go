// Package worlds — temperature variance methods (latitude, season, time of day,
// altitude) per WBH pp.115-124. Split from temperature.go for review ergonomics;
// all symbols remain in the worlds package.
package worlds

import "math"

// SunlightPortion computes the fraction of a solar day with sunlight at a
// given latitude, axial tilt, and date per WBH p.118. Returns 1.0 for polar
// day, 0.0 for polar night, otherwise the sunrise-angle fraction.
//
// Callers wanting daylight hours multiply the result by their solarDayHours.
func SunlightPortion(latDeg, axialTiltDeg, daysSinceSolstice, localYearDays float64) float64 {
	if localYearDays <= 0 {
		return 0
	}
	// Solar declination (in radians) = axial_tilt × cos(360° × date / year).
	declRad := axialTiltDeg * math.Cos(2*math.Pi*daysSinceSolstice/localYearDays) * math.Pi / 180.0

	tanLat := math.Tan(latDeg * math.Pi / 180.0)
	tanDecl := math.Tan(declRad)
	cosSunrise := -tanLat * tanDecl

	switch {
	case cosSunrise > 1:
		return 0 // polar night
	case cosSunrise < -1:
		return 1.0 // polar day
	default:
		return math.Acos(cosSunrise) / math.Pi
	}
}

// MeanByLatitude returns the annual mean temperature at a specific latitude
// per WBH p.116-117, ignoring season and time of day.
//
// Twilight worlds short-circuit to TwilightK (no meaningful latitude variation
// when one hemisphere is in perpetual day and the other in perpetual night;
// caller should use BrightSideK/DarkSideK directly for hemisphere selection).
func (t *Temperature) MeanByLatitude(latDeg float64) float64 {
	if t.IsTwilight {
		return t.TwilightK
	}
	zoneTiltFactor := t.zoneTiltAdjustment(latDeg)
	lumMod := zoneTiltFactor / t.AtmosphericFactor
	// Clamp to [-1, 1]: zoneTiltAdjustment can return negative for arctic
	// latitudes where sin(45° - lat) < 0, producing legitimate cooling.
	if lumMod > 1 {
		lumMod = 1
	}
	if lumMod < -1 {
		lumMod = -1
	}
	latLum := t.Luminosity * (1 + lumMod)
	if latLum < 0 {
		latLum = 0
	}
	return MeanTemperatureK(latLum, t.Albedo, t.GreenhouseFactor, t.AU)
}

// tropicalLatitudeBoundary returns the latitude (degrees, [0, 90]) at which
// the no-seasonal-swing zone ends. WBH p.116-117 reorganizes between the
// two parts:
//
//   - Part A (tilt < 45°): the tropical band runs |lat| ≤ axial_tilt.
//   - Part B (tilt ≥ 45°): the inner equatorial-tropical band runs
//     |lat| ≤ (90° − axial_tilt). Outside that band, the world enters
//     the arctic zone directly (no middle zone exists).
//
// NaN from Asin (when |AxialTiltFactor| > 1) clamps to 90; negative
// AxialTiltFactor takes the absolute value so the comparison stays
// meaningful when callers construct Temperature directly with a sign-
// flipped factor.
func (t *Temperature) tropicalLatitudeBoundary() float64 {
	tiltDeg := math.Asin(t.AxialTiltFactor) * 180.0 / math.Pi
	if math.IsNaN(tiltDeg) {
		return 90
	}
	if tiltDeg < 0 {
		tiltDeg = -tiltDeg
	}
	if tiltDeg >= 45 {
		return 90 - tiltDeg
	}
	return tiltDeg
}

// zoneTiltAdjustment returns the latitude-zone-adjusted axial-tilt-equivalent
// factor per WBH p.116-117 three-zone classification (tropical / middle /
// arctic). The structure differs between the book's two parts:
//
//	Part A (tilt < 45°):
//	  |lat| ≤ tilt           → sin(45° − tilt)        (tropical)
//	  |lat| > tilt           → sin(45° − lat)         (middle/arctic)
//
//	Part B (tilt ≥ 45°): the middle zone disappears; the inner equatorial-
//	tropical band uses the arctic-edge result, the rest uses arctic.
//	  |lat| ≤ (90° − tilt)   → sin(tilt − 45°)        (inner band)
//	  |lat| > (90° − tilt)   → sin(45° − lat)         (arctic)
//
// Returns are continuous at the Part A / Part B boundary (tilt = 45°): both
// formulas evaluate to 0 there.
//
// Reads raw axial tilt directly rather than calling tropicalLatitudeBoundary —
// this function discriminates Part A vs Part B (raw tilt < 45 vs ≥ 45),
// while tropicalLatitudeBoundary returns the part-aware no-seasonal-swing
// boundary (which is 90 − tilt under Part B, the wrong value here).
func (t *Temperature) zoneTiltAdjustment(latDeg float64) float64 {
	rawTilt := math.Asin(t.AxialTiltFactor) * 180.0 / math.Pi
	if math.IsNaN(rawTilt) {
		rawTilt = 90
	}
	if rawTilt < 0 {
		rawTilt = -rawTilt
	}
	if latDeg < 0 {
		latDeg = -latDeg
	}
	if latDeg > 90 {
		latDeg = 90
	}

	if rawTilt >= 45 {
		// WBH p.117 Part B: middle zone disappears.
		if latDeg <= 90.0-rawTilt {
			return math.Sin((rawTilt - 45.0) * math.Pi / 180.0)
		}
		return math.Sin((45.0 - latDeg) * math.Pi / 180.0)
	}

	// WBH p.116 Part A: tropical band is |lat| ≤ tilt; rest is middle/arctic.
	if latDeg <= rawTilt {
		return math.Sin((45.0 - rawTilt) * math.Pi / 180.0)
	}
	return math.Sin((45.0 - latDeg) * math.Pi / 180.0)
}

// MeanBySeason returns the mean temperature on a specific day at a specific
// latitude, ignoring time of day, per WBH p.115-116.
//
// daysSinceSolstice: 0 = summer solstice in the relevant hemisphere; year/2 = winter solstice.
// localYearDays: caller decides — for moons, use parent's stellar year (moons co-orbit star with planet).
//
// Composition rule per WBH p.116-117 — the book splits behavior at 45°:
//
//   - Part A (tilt < 45°): tropical band is |lat| ≤ tilt — no seasonal
//     swing per "tropical temperatures have little seasonal variation
//     (from axial tilt)." Outside the band (middle/arctic zone), the
//     seasonal axial-tilt factor is added to the zone latitude
//     adjustment per "the zone's latitude adjustment is added to the
//     axial tilt factor for that time period."
//
//   - Part B (tilt ≥ 45°): the middle zone disappears. The inner
//     equatorial-tropical band is |lat| ≤ (90 − tilt) with no seasonal
//     swing; outside that band the world is fully arctic with the
//     seasonal swing applied.
//
// In both parts, tropicalLatitudeBoundary returns the correct no-
// seasonal-swing boundary, and zoneTiltAdjustment returns the
// part-aware zone latitude adjustment — see those helpers for the
// per-band formulas.
//
// Twilight worlds always return TwilightK (band centerline). Hemisphere-
// aware selection is the caller's responsibility: read t.BrightSideK /
// t.TwilightK / t.DarkSideK directly. WBH pp.120-122 model a 1:1-locked
// world as longitude-partitioned (substellar / terminator great circle /
// antistellar) — bright vs. dark is fundamentally a longitude decision,
// not a latitude one, so the lat-only scenario API cannot pick a
// hemisphere on a twilight world. Callers wanting hemisphere-aware
// values use the three cached scalars.
func (t *Temperature) MeanBySeason(latDeg, daysSinceSolstice, localYearDays float64) float64 {
	if t.IsTwilight {
		return t.TwilightK
	}
	if localYearDays <= 0 {
		return t.MeanByLatitude(latDeg)
	}

	// Zone latitude adjustment from WBH p.116-117. Sole variance
	// contributor inside the tropical band; added to the seasonal axial
	// tilt outside it. Note: zoneTiltAdjustment returns a single value
	// for the entire tropical band (sin(45° - tilt) for Part A,
	// sin(tilt - 45°) for Part B's inner band) — by book design, every
	// in-band latitude gets the same temperature regardless of where it
	// sits inside the band. That uniformity is intentional, not a bug.
	zoneAdj := t.zoneTiltAdjustment(latDeg)
	absLat := latDeg
	if absLat < 0 {
		absLat = -absLat
	}
	if absLat > 90 {
		absLat = 90
	}
	// "No seasonal swing" applies inside the tropical band — for Part A
	// (tilt < 45°) that band is |lat| ≤ tilt; for Part B (tilt ≥ 45°) it
	// is |lat| ≤ (90 − tilt). tropicalLatitudeBoundary returns the correct
	// boundary for each part per WBH p.116-117.
	isTropical := absLat <= t.tropicalLatitudeBoundary()

	variance := zoneAdj
	if !isTropical {
		// Adjusted Fractional Year per WBH p.115.
		stdYearDays := hoursPerYear / 24.0 // 365.25
		lagDays := 0.1 * math.Min(stdYearDays, localYearDays)
		adjFracYear := (daysSinceSolstice - 0.1*lagDays) / localYearDays

		// Seasonal axial tilt factor: cos(adjFracYear × 360°) × AxialTiltFactor.
		// +AxialTiltFactor at summer solstice; -AxialTiltFactor at winter.
		seasonalTilt := math.Cos(adjFracYear*2*math.Pi) * t.AxialTiltFactor
		variance += seasonalTilt
	}

	lumMod := variance / t.AtmosphericFactor
	if lumMod > 1 {
		lumMod = 1
	}
	if lumMod < -1 {
		lumMod = -1
	}
	latLum := t.Luminosity * (1 + lumMod)
	if latLum < 0 {
		latLum = 0
	}
	return MeanTemperatureK(latLum, t.Albedo, t.GreenhouseFactor, t.AU)
}

// AtMoment returns the instantaneous temperature at a specific moment per WBH p.117.
//
// Uses Method 1 (even-length days) internally; callers wanting Method 2
// precision should call SunlightPortion separately and modulate hoursSinceDawn.
//
// Hourly variation follows a cosine curve centered on the peak-heat time, which
// lags 15% of the solar day past solar noon (i.e., peak at 65% of the day from
// dawn). Dawn is the coolest point, peak is shortly after noon.
//
// Twilight worlds always return TwilightK (band centerline). Hemisphere-
// aware selection is the caller's responsibility: read t.BrightSideK /
// t.TwilightK / t.DarkSideK directly. WBH pp.120-122 model a 1:1-locked
// world as longitude-partitioned (substellar / terminator great circle /
// antistellar) — bright vs. dark is fundamentally a longitude decision,
// not a latitude one, so the lat-only scenario API cannot pick a
// hemisphere on a twilight world. Callers wanting hemisphere-aware
// values use the three cached scalars.
func (t *Temperature) AtMoment(latDeg, daysSinceSolstice, localYearDays, hoursSinceDawn, solarDayHours float64) float64 {
	if t.IsTwilight {
		return t.TwilightK
	}
	if solarDayHours <= 0 {
		return t.MeanBySeason(latDeg, daysSinceSolstice, localYearDays)
	}

	// Seasonal contribution.
	seasonalK := t.MeanBySeason(latDeg, daysSinceSolstice, localYearDays)

	// Hourly rotation factor: cosine centered on peak-heat time (65% of solar day).
	// cos(0) = 1 at the hottest point; cos(π) = -1 at dawn (coolest).
	// peakFrac = 0.65 (noon 0.50 + lag 0.15).
	peakFrac := 0.65
	fracDay := hoursSinceDawn / solarDayHours
	hourlyRot := math.Cos(2*math.Pi*(fracDay-peakFrac)) * t.RotationFactor

	// Apply hourly rotation as a luminosity-modifier delta scaled into K via fourth-root.
	delta := hourlyRot / t.AtmosphericFactor
	scale := math.Pow(1+delta, 0.25)
	if scale <= 0 {
		scale = 0.01
	}
	return seasonalK * scale
}

// AdjustedForAltitude returns a temperature adjusted for altitude per WBH
// p.123-124. The greenhouse factor is reduced because atmospheric pressure
// drops with altitude (exp(-altitude/scale_height)). The implementation
// recomputes the equation with the modified greenhouse factor, then scales
// baseTempK by the ratio newRefK/storedMeanK to preserve any caller-applied
// scenario adjustments (latitude, season, etc.).
//
// Uses the cached t.ScaleHeight (populated from body.Atmosphere.ScaleHeight
// at GenerateTemperature time). Returns baseTempK unchanged when altitudeKm
// is zero or t.ScaleHeight is zero (e.g., vacuum world or atmosphere with
// no scale-height data).
//
// This is an approximation — the book's full altitude treatment includes
// lapse-rate and density-gradient effects deferred beyond 3A2b-temp.
func (t *Temperature) AdjustedForAltitude(baseTempK, altitudeKm float64) float64 {
	if altitudeKm <= 0 || t.ScaleHeight <= 0 {
		return baseTempK
	}
	// Pressure scales as e^(-h/H). Greenhouse factor scales with √(pressure).
	pressureRatio := math.Exp(-altitudeKm / t.ScaleHeight)
	gAtAlt := t.GreenhouseFactor * math.Sqrt(pressureRatio)

	// Recompute mean equation with modified G, using stored A/L/AU. Then scale
	// baseTempK by the ratio to preserve caller-applied scenario adjustments.
	newRefK := MeanTemperatureK(t.Luminosity, t.Albedo, gAtAlt, t.AU)
	storedRefK := MeanTemperatureK(t.Luminosity, t.Albedo, t.GreenhouseFactor, t.AU)
	if storedRefK == 0 {
		return baseTempK
	}
	return baseTempK * (newRefK / storedRefK)
}
