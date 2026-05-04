package worlds

import (
	"math"

	"wbh/roller"
)

// DayLength — rotation periods per WBH pp.103-104.
type DayLength struct {
	SiderealHours         float64 // post-lock final value
	SolarHours            float64 // 0 when YearDays == 0 (e.g. caller leaves YearDays unset on 1:1 star-lock twilight zones)
	YearDays              float64 // local solar days = year_h / sidereal_h - 1
	BaselineSiderealHours float64 // raw roll result, pre-tidal-lock
}

// maxCascadeIterations bounds the 40+ reroll cascade per WBH p.103. Each
// iteration has a ~1/3 probability of terminating (1D < 5), so the expected
// length is ~3 iterations and the chance of reaching 20 iterations is ~10^-8.
// The cap is purely defensive against pathological scripted-dice or seeds.
const maxCascadeIterations = 20

// DayLengthDMs accumulates DMs for the basic rotation roll, WBH p.103.
type DayLengthDMs struct {
	SystemAgeGyr float64 // DM+1 per 2 Gyrs (round down)
	IsGGOrSizeS  bool    // multiplies result by 2
}

// systemAgeDM computes DM+1 per 2 Gyrs (round down).
func systemAgeDM(ageGyr float64) int {
	if ageGyr <= 0 {
		return 0
	}
	return int(math.Floor(ageGyr / 2.0))
}

// rollOneBasic computes one (2D-2)×4 + 2 + 1D + DMs increment.
func rollOneBasic(r roller.Roller, dm int) float64 {
	twoD := r.Roll("2D")
	oneD := r.Roll("1D")
	return float64((twoD-2)*4 + 2 + oneD + dm)
}

// RollBasicSiderealHours per WBH p.103: (2D-2) × 4 + 2 + 1D + DMs.
//
// If the result is 40 or greater, roll 1D: on a 5+, add another basic roll
// (consuming a fresh pair of 2D + 1D), then roll 1D again to check for further
// additions. Repeat until 1D < 5 or the result has no further additions.
//
// For gas giant or small body (Size 0 or S) rotation, multiply the FINAL
// (post-cascade) hours by 2.
func RollBasicSiderealHours(r roller.Roller, dms DayLengthDMs) (float64, error) {
	dm := systemAgeDM(dms.SystemAgeGyr)
	hours := rollOneBasic(r, dm)
	for i := 0; i < maxCascadeIterations && hours >= 40; i++ {
		check := r.Roll("1D")
		if check < 5 {
			break
		}
		hours += rollOneBasic(r, dm)
	}
	if dms.IsGGOrSizeS {
		hours *= 2
	}
	return hours, nil
}

// ComputeYearDays per WBH p.104: year_hours / sidereal_hours - 1.
//
// Tidal-locked-to-star worlds have sidereal day equal to year length, so the
// formula yields 0 (which is correct — there are zero solar days in a locked
// world's year). Callers detecting the twilight-zone case should leave
// YearDays unset rather than relying on this function. The siderealHours==0
// guard is purely a divide-by-zero safety.
func ComputeYearDays(yearHours, siderealHours float64) float64 {
	if siderealHours == 0 {
		return 0
	}
	return yearHours/siderealHours - 1
}

// ComputeSolarHours per WBH p.104: year_hours / year_days. Returns 0 if year_days == 0.
func ComputeSolarHours(yearHours, yearDays float64) float64 {
	if yearDays == 0 {
		return 0
	}
	return yearHours / yearDays
}
