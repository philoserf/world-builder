package stars

import (
	"fmt"
	"math"

	"github.com/philoserf/world-builder/roller"
)

// ----- P2-4: Stellar Orbit# Ranges (WBH p.27) -----

// IsGiantClass reports whether the luminosity class is one of the
// giants (Ia/Ib/II/III) per WBH p.14 nomenclature.
func IsGiantClass(lc LuminosityClass) bool {
	switch lc {
	case Ia, Ib, II, III:
		return true
	}
	return false
}

// RollCompanionOrbitOfGiant computes the Orbit# for a companion of a
// giant primary per WBH p.27: Orbit# = 1D × MAO(primary). Callers pass
// in the primary's MAO (via MAO); this stays a pure roll.
func RollCompanionOrbitOfGiant(r roller.Roller, primaryMAO float64) float64 {
	return float64(r.Roll("1D")) * primaryMAO
}

// RollStellarOrbit rolls the Orbit# for a star at the given orbit class.
//
// Formulas (WBH p.27):
//   - Close:     1D - 1 (a result of 0 means Orbit# 0.5, i.e. 0.2 AU)
//   - Near:      1D + 5
//   - Far:       1D + 11
//   - Companion: 1D / 10 + (2D - 7) / 100  (range 0.05 to 0.65)
//
// Companions of Class Ia/Ib/II/III primaries use 1D × MAO of the primary
// (page 39); this function has only the luminosity class, not the full
// star that MAO needs, so it returns ErrCompanionOfGiantMAO for that
// case. Compute MAO(primary) and use RollCompanionOrbitOfGiant directly
// for giant-class companions.
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
			return 0, ErrCompanionOfGiantMAO
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
	ExtraDM      int  // additional DM applied externally (e.g., 2B Step 7 anomaly DMs)
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
	dm += opts.ExtraDM
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

// ----- P2-8: Orbit# ↔ AU conversion (WBH p.26) -----

// OrbitToAU converts a fractional Orbit# to AU distance from the central
// star (WBH p.26).
//
//	AU = DistanceAU(floor) + DifferenceAU(floor) × frac
//
// Where `floor` is the largest whole-number Orbit# whose DistanceAU is ≤
// the input. Orbit# 0 with frac is special: AU = 0 + 0.4 × frac (e.g.
// Orbit# 0.5 = 0.2 AU, the Companion-orbit reference).
//
// Negative inputs are treated as 0; inputs ≥ 20 return DistanceAU(20).
func OrbitToAU(orbit float64) float64 {
	if orbit < 0 {
		orbit = 0
	}
	if orbit >= 20 {
		return OrbitNumberTable[20].DistanceAU
	}
	floor := int(orbit)
	frac := orbit - float64(floor)
	row := OrbitNumberTable[floor]
	return row.DistanceAU + row.DifferenceAU*frac
}

// AUToOrbit converts an AU distance back to a fractional Orbit# (WBH p.26).
//
//	Orbit# = full + (AU - DistanceAU(full)) / DifferenceAU(full)
//
// Where `full` is the largest whole-number Orbit# whose DistanceAU ≤ AU.
//
// Returns 0 for AU < 0; returns 20 for AU ≥ DistanceAU(20).
func AUToOrbit(au float64) float64 {
	if au <= 0 {
		return 0
	}
	if au >= OrbitNumberTable[20].DistanceAU {
		return 20
	}
	// Find the largest whole Orbit# whose DistanceAU ≤ au.
	full := 0
	for n := 0; n <= 20; n++ {
		if OrbitNumberTable[n].DistanceAU <= au {
			full = n
		} else {
			break
		}
	}
	row := OrbitNumberTable[full]
	if row.DifferenceAU == 0 {
		return float64(full)
	}
	return float64(full) + (au-row.DistanceAU)/row.DifferenceAU
}
