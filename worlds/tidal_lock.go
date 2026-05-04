package worlds

import (
	"math"

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
	body *DetailedPlacement,
	sys stars.System,
	parentPlanet *DetailedPlacement,
	moonRef *Moon,
) map[TidalLockCase]int {
	common := commonTidalLockDMs(body, sys)
	out := make(map[TidalLockCase]int, 3)

	// Planet → star: every body has a star to potentially lock to.
	out[TidalLockCasePlanetToStar] = common + planetToStarDMs(body, sys)

	// Moon → planet: only applies if body is a moon (parentPlanet and moonRef provided).
	if parentPlanet != nil && moonRef != nil {
		out[TidalLockCaseMoonToPlanet] = common + moonToPlanetDMs(moonRef, parentPlanet)
	}

	// Planet → moon: only applies if body is a planet with at least one Size-1+ moon.
	if parentPlanet == nil && moonRef == nil && hasSignificantMoon(body) {
		out[TidalLockCasePlanetToMoon] = common + planetToMoonDMs(body)
	}

	return out
}

// commonTidalLockDMs computes DMs that apply to all three cases per WBH p.106.
func commonTidalLockDMs(body *DetailedPlacement, sys stars.System) int {
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
func planetToStarDMs(body *DetailedPlacement, sys stars.System) int {
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
func moonToPlanetDMs(moonRef *Moon, parent *DetailedPlacement) int {
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
	mass := parentMassEarth(parent)
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
func planetToMoonDMs(body *DetailedPlacement) int {
	dm := -10 // Base

	// Use the closest significant moon (smallest OrbitPD with Size 1+).
	var closest *Moon
	for i := range body.Moons {
		if nForSizeCode(body.Moons[i].SizeCode) < 1 {
			continue
		}
		if closest == nil || body.Moons[i].OrbitPD < closest.OrbitPD {
			closest = &body.Moons[i]
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

// --- helpers ---

func hasSignificantMoon(body *DetailedPlacement) bool {
	return countSignificantMoons(body) > 0
}

func countSignificantMoons(body *DetailedPlacement) int {
	n := 0
	for i := range body.Moons {
		if nForSizeCode(body.Moons[i].SizeCode) >= 1 {
			n++
		}
	}
	return n
}

func sumSignificantMoonSizes(body *DetailedPlacement) int {
	total := 0
	for i := range body.Moons {
		if n := nForSizeCode(body.Moons[i].SizeCode); n >= 1 {
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

// parentMassEarth returns the parent planet's mass in Earth masses.
// For gas giants, reads MassEarth directly (set by the GG sizing step).
// For terrestrial parents with BodyPhysical, derives mass from density and diameter.
func parentMassEarth(parent *DetailedPlacement) float64 {
	if parent.Body == BodyGasGiant {
		return parent.MassEarth
	}
	if parent.Physical != nil {
		return DeriveMass(parent.Physical.Density, parent.DiameterKm)
	}
	return 0
}
