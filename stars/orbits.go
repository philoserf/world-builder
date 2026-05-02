package stars

import (
	"fmt"
	"math"

	"wbh/roller"
)

// ----- P2-4: Stellar Orbit# Ranges (WBH p.27) -----

// RollStellarOrbit rolls the Orbit# for a star at the given orbit class.
//
// Formulas (WBH p.27):
//   - Close:     1D - 1 (a result of 0 means Orbit# 0.5, i.e. 0.2 AU)
//   - Near:      1D + 5
//   - Far:       1D + 11
//   - Companion: 1D / 10 + (2D - 7) / 100  (range 0.05 to 0.65)
//
// Companions of Class Ia/Ib/II/III primaries use 1D × MAO of the primary
// (page 39); MAO is a Plan 3+ concept, so this function returns an error
// if asked for a Companion orbit with a giant-class primary.
func RollStellarOrbit(r roller.Roller, oc OrbitClass, primaryClass LuminosityClass) (float64, error) {
	switch oc {
	case OrbitClose:
		v := r.Roll("1D-1")
		if v == 0 {
			return 0.5, nil
		}
		return float64(v), nil
	case OrbitNear:
		v := r.Roll("1D+5")
		return float64(v), nil
	case OrbitFar:
		v := r.Roll("1D+11")
		return float64(v), nil
	case OrbitCompanion:
		switch primaryClass {
		case Ia, Ib, II, III:
			return 0, fmt.Errorf("stars: companion of giant primary requires MAO (Plan 3+)")
		}
		first := r.Roll("1D")
		second := r.Roll("2D-7")
		return float64(first)/10.0 + float64(second)/100.0, nil
	}
	return 0, fmt.Errorf("stars: unknown orbit class: %q", oc)
}

// ----- P2-5: Eccentricity rolls and DMs (WBH p.27) -----

// EccentricityOpts collects the contextual flags for the WBH p.27 DMs.
type EccentricityOpts struct {
	IsStar       bool    // adds DM+2 (star eccentricities)
	NestingDepth int     // count of bodies this object orbits beyond the first; adds NestingDepth as DM
	Orbit        float64 // for the sub-1.0 / age>1Gyr DM-1 rule
	SystemAgeGyr float64
	IsBeltMember bool // adds DM+1
}

// RollEccentricity rolls 2D + DMs into the EccentricityValues table,
// clamps the row to [5, 12], adds a second-roll term, and returns the
// final eccentricity value clamped to [0, 0.999].
//
// DMs (WBH p.27):
//   - Star eccentricities: +2
//   - For each object an object directly orbits beyond the first: +1
//     per nesting depth (passed via opts.NestingDepth)
//   - For Orbit#s below 1.0 if SystemAgeGyr > 1 Gyr: -1
//   - Object is a significant body in an asteroid or planetoid belt: +1
func RollEccentricity(r roller.Roller, opts EccentricityOpts) (float64, error) {
	dm := 0
	if opts.IsStar {
		dm += 2
	}
	dm += opts.NestingDepth
	if opts.Orbit < 1.0 && opts.SystemAgeGyr > 1.0 {
		dm--
	}
	if opts.IsBeltMember {
		dm++
	}
	natural := r.Roll("2D")
	row := max(5, min(12, natural+dm))
	rowData, ok := EccentricityValues[row]
	if !ok {
		return 0, fmt.Errorf("stars: eccentricity row %d missing", row)
	}
	second := r.Roll(rowData.SecondRoll)
	v := rowData.Base + float64(second)/rowData.Divisor
	if v < 0 {
		v = 0
	}
	if v > 0.999 {
		v = 0.999
	}
	return v, nil
}

// ----- P2-6: Inclination (WBH p.28) -----

// RollInclination rolls 2D and applies the WBH p.28 severity-keyed
// formula. Returns degrees (0-180) plus the severity name.
//
// Severity formulas:
//   - 2-6  Very Low:    1D / 2
//   - 7    Low:         1D
//   - 8    Moderate:    2D
//   - 9    High:        (2D × 3) + 1D
//   - 10   Very High:   (1D + 1) × 5 + 1D
//   - 11   Extreme:     (3D × 5) - 1D
//   - 12   Retrograde:  180 - <recursive call>
func RollInclination(r roller.Roller) (degrees float64, severity string, err error) {
	natural := r.Roll("2D")
	switch {
	case natural <= 6:
		return float64(r.Roll("1D")) / 2.0, "VeryLow", nil
	case natural == 7:
		return float64(r.Roll("1D")), "Low", nil
	case natural == 8:
		return float64(r.Roll("2D")), "Moderate", nil
	case natural == 9:
		return float64(r.Roll("2D")*3 + r.Roll("1D")), "High", nil
	case natural == 10:
		return float64((r.Roll("1D")+1)*5 + r.Roll("1D")), "VeryHigh", nil
	case natural == 11:
		return float64(r.Roll("3D")*5 - r.Roll("1D")), "Extreme", nil
	default: // 12+
		inner, _, ierr := RollInclination(r)
		if ierr != nil {
			return 0, "", ierr
		}
		return 180.0 - inner, "Retrograde", nil
	}
}

// ----- P2-7: Star orbit period (Kepler's third law, WBH p.30) -----

// OrbitPeriodYears returns the orbital period in years for two masses
// orbiting a common barycentre at semi-major axis auSemiMajor.
//
// Kepler's third law: P (years) = sqrt(AU^3 / (M + m))
func OrbitPeriodYears(auSemiMajor, primaryMass, companionMass float64) float64 {
	return math.Sqrt(auSemiMajor * auSemiMajor * auSemiMajor / (primaryMass + companionMass))
}
