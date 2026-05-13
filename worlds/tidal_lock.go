package worlds

import (
	"fmt"
	"math"

	"wbh/roller"
	"wbh/stars"
)

// TidalLock — tidal lock state per WBH pp.105-107.
type TidalLock struct {
	Case              TidalLockCase
	InitialResult     int    // 2D + DM (pre-verification)
	FinalResult       int    // post-verification (= InitialResult when no verification fired)
	VerificationFired bool   // p.105 footnote: InitialResult ≥ 12 triggered 2D verification, natural 12 caused a no-DM reroll
	LockRatio         string // "" | "3:2" | "1:1"
	IsTwilightZone    bool   // true only if Case == PlanetToStar AND LockRatio == "1:1"

	// Effect descriptors — set based on FinalResult
	DayLengthMultiplier float64 // 1.5 / 2 / 3 / 5 for FinalResult 3-6
	NewSiderealHours    float64 // for prograde/retrograde reroll FinalResult 7-10
	BecomesRetrograde   bool    // FinalResult 9-10
	EccentricityMutated bool    // 1:1 lock with old ecc > 0.1
	AxialTiltMutated    bool    // 3:2 or 1:1 lock with old tilt > 3°
}

// TidalLockCase identifies which p.106 case fired (highest DM among applicable).
type TidalLockCase int

// TidalLockCase constants enumerate the three p.106 check scenarios.
// TidalLockCaseNone is returned when no applicable case has a DM above -10.
const (
	TidalLockCaseNone         TidalLockCase = iota // no roll (all DMs ≤ -10)
	TidalLockCasePlanetToStar                      // planet locked to its star(s)
	TidalLockCaseMoonToPlanet                      // moon locked to its parent planet
	TidalLockCasePlanetToMoon                      // planet locked to its moon
)

// EvaluateTidalLockDMs returns per-case DM totals per WBH p.106.
//
// Inputs:
//   - body:        the body being checked for tidal lock
//   - sys:         the star system (for primary mass and system age)
//   - parentPlanet: the body's parent planet if body is a moon; nil for planets
//   - moonRef:     the moon record if body is a moon; nil for planets
//
// All-cases-common DMs stack additively on top of each case's base DM and
// case-specific DMs. Cases that don't apply are absent from the returned map:
//   - moon→planet requires parentPlanet and moonRef to both be non-nil
//   - planet→moon requires the body to have at least one significant (Size 1+) moon
//
// Per WBH p.106: "In 'edge' conditions where a value corresponds to more than
// one DM or falls between two DMs, use the DM closer to 0."
func EvaluateTidalLockDMs(
	body *Body,
	sys stars.System,
	parentPlanet *Body,
	moonRef *Body,
) map[TidalLockCase]int {
	common := commonTidalLockDMs(body, sys)
	out := make(map[TidalLockCase]int, 3)

	// Planet → star: gated to non-moons (planets and belts — anything
	// that orbits a star directly). Moons are excluded because they
	// store their orbit around the parent in OrbitPD; body.Orbit is 0
	// for moons, which would spuriously trigger the orbit < 1 +14
	// close-orbit DM and drown out the proper moon → planet case.
	if parentPlanet == nil && moonRef == nil {
		out[TidalLockCasePlanetToStar] = common + planetToStarDMs(body, sys)
	}

	// Moon → planet: only applies if body is a moon (parentPlanet and moonRef provided).
	if parentPlanet != nil && moonRef != nil {
		out[TidalLockCaseMoonToPlanet] = common + moonToPlanetDMs(moonRef, parentPlanet)
	}

	// Planet → moon: only applies if body is a TERRESTRIAL planet with at
	// least one moon already in a 1:1 or 3:2 lock per WBH p.107. The
	// pre-locked-moon precondition requires moon-cases to be evaluated in
	// an earlier pass; the Stage 4 orchestrator is responsible for that
	// ordering.
	if parentPlanet == nil && moonRef == nil &&
		isTerrestrial(body) && hasSignificantMoon(body) && hasLockedMoon(body) {
		out[TidalLockCasePlanetToMoon] = common + planetToMoonDMs(body)
	}

	return out
}

// commonTidalLockDMs computes DMs that apply to all three cases per WBH p.106.
func commonTidalLockDMs(body *Body, sys stars.System) int {
	dm := 0

	// Size 1 or more: DM+Size÷3 (round up).
	if n := nForSizeCode(body.SizeCode); n >= 1 {
		dm += int(math.Ceil(float64(n) / 3.0))
	}

	// Eccentricity greater than 0.1: DM-Eccentricity×10 (round down).
	if body.Eccentricity > 0.1 {
		dm -= int(math.Floor(body.Eccentricity * 10.0))
	}

	// Axial tilt DMs (cumulative per p.106 table):
	//   Above 30°:           DM-2
	//   Between 60° and 120°: DM-4
	//   Between 80° and 100°: DM-4
	if body.AxialTilt != nil {
		t := body.AxialTilt.Degrees
		if t > 30 {
			dm -= 2
		}
		if t >= 60 && t <= 120 {
			dm -= 4
		}
		if t >= 80 && t <= 100 {
			dm -= 4
		}
	}

	// Atmospheric pressure above 2.5 bar: DM-2.
	if body.Atmosphere != nil && body.Atmosphere.Pressure > 2.5 {
		dm -= 2
	}

	// System age:
	switch {
	case sys.Primary.AgeGyr < 1:
		dm -= 2
	case sys.Primary.AgeGyr >= 5 && sys.Primary.AgeGyr <= 10:
		dm += 2
	case sys.Primary.AgeGyr > 10:
		dm += 4
	}

	return dm
}

// planetToStarDMs computes the planet→star case-specific DMs per WBH p.106.
func planetToStarDMs(body *Body, sys stars.System) int {
	dm := -4 // Base

	// Orbit# DM ladder.
	orbit := body.Orbit
	switch {
	case orbit < 1:
		// DM+4 + (10 × (1-Orbit# fraction, rounded down)).
		// "Orbit# fraction" is the decimal portion of the orbit number.
		// Example: orbit=0.5 → fraction=0.5, 1-0.5=0.5, ×10=5, floor=5 → DM+4+5=+9.
		fractionalPart := orbit - math.Floor(orbit)
		dm += 4 + int(math.Floor(10.0*(1.0-fractionalPart)))
	case orbit < 2:
		dm += 4
	case orbit < 3:
		dm++
	default:
		// Orbit# greater than 3: DM-Orbit# (rounded down) × 2.
		dm -= int(math.Floor(orbit)) * 2
	}

	// Star mass(es) DM ladder.
	starMass := totalStellarMass(sys)
	switch {
	case starMass < 0.5:
		dm -= 2
	case starMass <= 1.0:
		dm--
	case starMass >= 2 && starMass <= 5:
		dm++
	case starMass > 5:
		dm += 2
		// 1.0 < starMass < 2: no DM (falls between table rows; use DM closer to 0 = 0)
	}

	// Planet orbits more than one star: DM-total number of stars orbited.
	if numStars := countStarsOrbited(sys); numStars > 1 {
		dm -= numStars
	}

	// Planet has a significant moon (Size 1+): DM-total Size of all such moons.
	dm -= sumSignificantMoonSizes(body)

	return dm
}

// moonToPlanetDMs computes the moon→planet case-specific DMs per WBH p.106.
func moonToPlanetDMs(moonRef, parent *Body) int {
	dm := 6 // Base

	// Moon orbit greater than 20 PD: DM-PD÷20 (round down).
	if moonRef.OrbitPD > 20 {
		dm -= int(math.Floor(moonRef.OrbitPD / 20.0))
	}

	// Moon orbit is retrograde: DM-2.
	if moonRef.Retrograde {
		dm -= 2
	}

	// Planet mass DM ladder (Earth masses).
	mass := parent.MassOrDerived()
	switch {
	case mass >= 1 && mass < 10:
		dm += 2
	case mass >= 10 && mass < 100:
		dm += 4
	case mass >= 100 && mass < 1000:
		dm += 6
	case mass >= 1000:
		dm += 8
	}

	return dm
}

// planetToMoonDMs computes the planet→moon case-specific DMs per WBH p.106.
// The book does not publish a complete DM table for this case in the same
// detail as the other two; this implements the structure analogous to p.106's
// listed parameters. Only called when hasSignificantMoon is true.
func planetToMoonDMs(body *Body) int {
	dm := -10 // Base

	// Use the closest significant moon (smallest OrbitPD with Size 1+).
	var closest *Body
	for i := range body.Children {
		if nForSizeCode(body.Children[i].SizeCode) < 1 {
			continue
		}
		if closest == nil || body.Children[i].OrbitPD < closest.OrbitPD {
			closest = body.Children[i]
		}
	}
	if closest == nil {
		return dm // guard (should not happen given hasSignificantMoon gate)
	}

	// Moon Size 1 or above: DM+Size.
	dm += nForSizeCode(closest.SizeCode)

	// Moon orbit DM ladder (planetary diameters). The 40 < pd ≤ 60 range
	// receives no DM per the WBH p.106 table.
	pd := closest.OrbitPD
	switch {
	case pd < 5:
		dm += 5 + int(math.Ceil((5.0-pd)*5.0))
	case pd <= 10:
		dm += 4
	case pd <= 20:
		dm += 2
	case pd <= 40:
		dm++
	case pd > 60:
		dm -= 6
	}

	// Planet has more than one significant moon: DM-2 per additional moon.
	if count := countSignificantMoons(body); count > 1 {
		dm -= 2 * (count - 1)
	}

	return dm
}

// SelectHighestDMCase returns the case to roll, applying p.106 tiebreakers:
//   - Cases with DM ≤ -10 are filtered out (no roll required for those).
//   - On ties, moon-cases ordered first (MoonToPlanet before PlanetToStar).
//   - On ties between multiple moons (future: moonToPlanet for multiple moons),
//     closest moon first — handled at orchestration level via per-moon iteration.
//   - Returns TidalLockCaseNone if no case applies.
func SelectHighestDMCase(dms map[TidalLockCase]int, _ *Body) (TidalLockCase, int) {
	bestCase := TidalLockCaseNone
	bestDM := -10 // exclusive lower bound
	// Order: MoonToPlanet > PlanetToMoon > PlanetToStar.
	// (Moon cases first on tie per p.106.)
	priority := []TidalLockCase{
		TidalLockCaseMoonToPlanet,
		TidalLockCasePlanetToMoon,
		TidalLockCasePlanetToStar,
	}
	for _, kase := range priority {
		dm, ok := dms[kase]
		if !ok {
			continue
		}
		if dm <= -10 {
			continue
		}
		if dm > bestDM {
			bestDM = dm
			bestCase = kase
		}
	}
	if bestCase == TidalLockCaseNone {
		return TidalLockCaseNone, 0
	}
	return bestCase, bestDM
}

// RollTidalLockStatus rolls 2D + DM. Caller handles the natural-12-verification
// and effect application separately.
func RollTidalLockStatus(r roller.Roller, dm int) int {
	twoD := r.Roll("2D")
	return twoD + dm
}

// --- helpers ---

// isTerrestrial reports whether a body is eligible for the Planet→Moon
// tidal-lock case per WBH p.107 (terrestrial worlds, Size 1–F).
func isTerrestrial(body *Body) bool {
	return body.Kind == BodyTerrestrial
}

func hasSignificantMoon(body *Body) bool {
	return countSignificantMoons(body) > 0
}

// hasLockedMoon reports whether body has at least one moon already in a
// 1:1 or 3:2 tidal lock with the planet, per WBH p.107 Planet→Moon
// pre-condition. Relies on the Stage 4 two-pass ordering — moons must
// be evaluated before planets so their TidalLock fields are populated.
func hasLockedMoon(body *Body) bool {
	for _, c := range body.Children {
		if c.TidalLock == nil {
			continue
		}
		if c.TidalLock.LockRatio == "1:1" || c.TidalLock.LockRatio == "3:2" {
			return true
		}
	}
	return false
}

func countSignificantMoons(body *Body) int {
	n := 0
	for i := range body.Children {
		if nForSizeCode(body.Children[i].SizeCode) >= 1 {
			n++
		}
	}
	return n
}

func sumSignificantMoonSizes(body *Body) int {
	total := 0
	for i := range body.Children {
		if n := nForSizeCode(body.Children[i].SizeCode); n >= 1 {
			total += n
		}
	}
	return total
}

// totalStellarMass returns the summed mass (in solar units) of all stars in
// the primary group (primary + OrbitCompanion-class companion with ParentIndex
// == -1, i.e. the Ab in an Aab pair). Per WBH p.106: star mass(es) is the
// relevant gravitational mass for the planet→star DM.
func totalStellarMass(sys stars.System) float64 {
	total := sys.Primary.Mass
	for _, c := range sys.Companions {
		if c.OrbitClass == stars.OrbitCompanion && c.ParentIndex == -1 {
			total += c.Star.Mass
		}
	}
	return total
}

// countStarsOrbited returns the number of stars gravitationally bound to the
// primary group that the planet orbits. For a single primary this is 1; for
// an Aab close binary it is 2 (primary Aa + companion Ab).
func countStarsOrbited(sys stars.System) int {
	count := 1
	for _, c := range sys.Companions {
		if c.OrbitClass == stars.OrbitCompanion && c.ParentIndex == -1 {
			count++
		}
	}
	return count
}

// ApplyTidalLockEffect mutates body fields based on the rolled 2D+DM result,
// per WBH p.105 effect table.
//
// Natural-12 verification (p.105 footnote): when InitialResult ≥ 12, rolls 2D;
// on a natural 12, rerolls TidalLockStatus with DM=0 and uses that as FinalResult.
//
// On 3:2 or 1:1 lock with old tilt > 3°: rerolls AxialTilt.Degrees as (2D-2)/10.
// On 1:1 lock with old ecc > 0.1: rerolls eccentricity with DM-2 on the standard
// eccentricity table (stars.EccentricityValues), takes min of original/new.
//
// Recomputes YearDays + SolarHours after SiderealHours mutation.
func ApplyTidalLockEffect(
	r roller.Roller,
	body *Body,
	_ *Body,
	kase TidalLockCase,
	initialResult int,
	yearHours float64,
) (TidalLock, error) {
	tl := TidalLock{
		Case:          kase,
		InitialResult: initialResult,
		FinalResult:   initialResult,
	}

	// Natural-12 verification per p.105 footnote: result ≥ 12 may break the
	// 1:1 lock via a 2D verification roll; on natural 12 reroll status with no DMs.
	if initialResult >= 12 {
		verification := r.Roll("2D")
		if verification == 12 {
			tl.VerificationFired = true
			tl.FinalResult = RollTidalLockStatus(r, 0)
		}
	}

	// Apply effect based on FinalResult.
	switch {
	case tl.FinalResult <= 2:
		// No effect.
	case tl.FinalResult == 3:
		tl.DayLengthMultiplier = 1.5
	case tl.FinalResult == 4:
		tl.DayLengthMultiplier = 2
	case tl.FinalResult == 5:
		tl.DayLengthMultiplier = 3
	case tl.FinalResult == 6:
		tl.DayLengthMultiplier = 5
	case tl.FinalResult == 7:
		tl.NewSiderealHours = float64(r.Roll("1D") * 5 * 24)
	case tl.FinalResult == 8:
		tl.NewSiderealHours = float64(r.Roll("1D") * 20 * 24)
	case tl.FinalResult == 9:
		tl.NewSiderealHours = float64(r.Roll("1D") * 10 * 24)
		tl.BecomesRetrograde = true
	case tl.FinalResult == 10:
		tl.NewSiderealHours = float64(r.Roll("1D") * 50 * 24)
		tl.BecomesRetrograde = true
	case tl.FinalResult == 11:
		tl.LockRatio = "3:2"
	case tl.FinalResult >= 12:
		tl.LockRatio = "1:1"
		if kase == TidalLockCasePlanetToStar {
			tl.IsTwilightZone = true
		}
	}

	// Apply day-length effects.
	if body.DayLength != nil {
		switch {
		case tl.DayLengthMultiplier > 0:
			body.DayLength.SiderealHours *= tl.DayLengthMultiplier
		case tl.NewSiderealHours > 0:
			body.DayLength.SiderealHours = tl.NewSiderealHours
		case tl.LockRatio == "3:2":
			body.DayLength.SiderealHours = yearHours * 2.0 / 3.0
		case tl.LockRatio == "1:1":
			body.DayLength.SiderealHours = yearHours
		}

		// Recompute YearDays + SolarHours.
		if tl.LockRatio == "1:1" && tl.IsTwilightZone {
			body.DayLength.SolarHours = 0
			body.DayLength.YearDays = 0
		} else if body.DayLength.SiderealHours > 0 {
			body.DayLength.YearDays = ComputeYearDays(yearHours, body.DayLength.SiderealHours)
			body.DayLength.SolarHours = ComputeSolarHours(yearHours, body.DayLength.YearDays)
		}
	}

	// Apply retrograde flag (results 9-10).
	if tl.BecomesRetrograde && body.AxialTilt != nil {
		if body.AxialTilt.Degrees < 90 {
			body.AxialTilt.Degrees = 180 - body.AxialTilt.Degrees
		}
		body.AxialTilt.Retrograde = body.AxialTilt.Degrees > 90
	}

	// 3:2 or 1:1 lock axial-tilt reroll: if old tilt > 3°, reroll as (2D-2)/10.
	// LockRatio (results 11/12+) and BecomesRetrograde (results 9/10) are mutually
	// exclusive in the FinalResult switch, so body.AxialTilt.Degrees here is the
	// pre-flip value and the "old tilt" check is meaningful.
	if (tl.LockRatio == "3:2" || tl.LockRatio == "1:1") && body.AxialTilt != nil && body.AxialTilt.Degrees > 3 {
		twoD := r.Roll("2D")
		body.AxialTilt.Degrees = float64(twoD-2) / 10.0
		body.AxialTilt.Retrograde = false
		tl.AxialTiltMutated = true
	}

	// 1:1 lock eccentricity reroll: if old ecc > 0.1, reroll with DM-2, take min.
	if tl.LockRatio == "1:1" && body.Eccentricity > 0.1 {
		newEcc, err := rerollEccentricityDMMinus2(r)
		if err != nil {
			return tl, fmt.Errorf("worlds: ApplyTidalLockEffect: ecc reroll: %w", err)
		}
		if newEcc < body.Eccentricity {
			body.Eccentricity = newEcc
		}
		tl.EccentricityMutated = true
	}

	return tl, nil
}

// rerollEccentricityDMMinus2 rerolls eccentricity using stars.RollEccentricity
// with ExtraDM=-2 per WBH p.105 footnote. Called only for 1:1-locked bodies
// whose existing eccentricity exceeds 0.1. Consumes 2 roller calls:
// one 2D for the table row, one 1D (or 2D for rows 11-12) for the value.
func rerollEccentricityDMMinus2(r roller.Roller) (float64, error) {
	return stars.RollEccentricity(r, stars.EccentricityOpts{ExtraDM: -2})
}

// GenerateTidalLock orchestrates the per-body tidal-lock pipeline per WBH p.106.
// Returns nil (no error) for empty bodies or when no tidal-lock case applies.
// Mutates body's DayLength, AxialTilt, and Eccentricity in place when an effect applies.
func GenerateTidalLock(
	r roller.Roller,
	body *Body,
	moonRef *Body,
	sys stars.System,
	parentPlanet *Body,
	yearHours float64,
) (*TidalLock, error) {
	if body.Kind == BodyEmpty {
		return nil, nil
	}

	dms := EvaluateTidalLockDMs(body, sys, parentPlanet, moonRef)
	kase, dm := SelectHighestDMCase(dms, body)
	if kase == TidalLockCaseNone {
		return nil, nil // no case applies; body has no tidal lock pressure
	}

	initialResult := RollTidalLockStatus(r, dm)
	tl, err := ApplyTidalLockEffect(r, body, moonRef, kase, initialResult, yearHours)
	if err != nil {
		return nil, fmt.Errorf("worlds: GenerateTidalLock: %w", err)
	}
	return &tl, nil
}
