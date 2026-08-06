// Package stars implements WBH star generation procedures.
package stars

import (
	"fmt"
	"math"

	"github.com/philoserf/world-builder/roller"
)

// MainSequenceLifespan returns the main-sequence lifespan in Gyr (WBH p. 20):
//
//	Lifespan = 10 / Mass^2.5
func MainSequenceLifespan(mass float64) float64 {
	return 10.0 / math.Pow(mass, 2.5)
}

// SmallStarAge returns a random age for a small star in Gyr (WBH p. 21).
//
// accuracy=1: 1D × 2 + D3 - 1
// accuracy=2: 1D × 2 + D3 - 2 + d10/10
//
// Higher accuracy is referenced in the book ("Adding additional digits
// of accuracy requires additional d10 rolls"). Only 1 and 2 are
// implemented here.
func SmallStarAge(r roller.Roller, accuracy int) (float64, error) {
	if accuracy != 1 && accuracy != 2 {
		return 0, fmt.Errorf("stars: accuracy must be 1 or 2, got %d", accuracy)
	}

	oneD := r.Roll("1D")

	d3 := r.Roll("D3")
	if accuracy == 1 {
		return float64(oneD*2 + d3 - 1), nil
	}

	d10 := r.Roll("d10")

	return float64(oneD*2+d3-2) + float64(d10)/10.0, nil
}

// SubgiantLifespan returns the subgiant phase lifespan in Gyr (WBH p. 21):
//
//	Subgiant Lifespan = Main Sequence Lifespan / (4 + Mass)
func SubgiantLifespan(mainSequenceLifespanGyr, mass float64) float64 {
	return mainSequenceLifespanGyr / (4.0 + mass)
}

// GiantLifespan returns the giant phase lifespan in Gyr (WBH p. 22):
//
//	Giant Lifespan = Main Sequence Lifespan / (10 × Mass^3)
func GiantLifespan(mainSequenceLifespanGyr, mass float64) float64 {
	return mainSequenceLifespanGyr / (10.0 * mass * mass * mass)
}

// FinalAgeProgenitor returns the total elapsed lifespan of a star up to
// its post-stellar transition (WBH p. 22):
//
//	Star Final Age = (10 / Mass^2.5) × (1 + 1/(4+Mass) + 1/(10×Mass^3))
//
// progenitorMass is the original star's mass (NOT dead-star mass).
func FinalAgeProgenitor(progenitorMass float64) float64 {
	msl := MainSequenceLifespan(progenitorMass)

	return msl * (1.0 +
		1.0/(4.0+progenitorMass) +
		1.0/(10.0*progenitorMass*progenitorMass*progenitorMass))
}

// AgeSpecialObject returns the total age in Gyr for a special or
// post-stellar object per WBH p.22.
//
// Roll order:
//  1. Base formula:
//     - "small_star":    SmallStarAge(r, accuracy=1) — consumes 1D, D3
//     - "100m_per_2d10": 100 / 2d10 (in millions of years; convert to Gyr).
//     Consumes 2 d10 rolls.
//     - "10m_per_2d10":  10 / 2d10 (millions of years).
//  2. If AddProgenitorAge: roll D3, compute progenitor_mass = (2+D3) ×
//     deadStarMass, then add FinalAgeProgenitor(progenitor_mass).
//
// `deadStarMass` is the remnant's current mass (used only for objects
// where AddProgenitorAge is true). Pass 0 for BD and Protostar.
func AgeSpecialObject(r roller.Roller, kind StarKind, deadStarMass float64) (float64, error) {
	row, ok := SpecialObjectAgeByType[kind]
	if !ok {
		return 0, fmt.Errorf("stars: kind %q has no age formula: %w", kind, ErrSpecialCircumstances)
	}

	var base float64

	switch row.BaseFormula {
	case "small_star":
		v, err := SmallStarAge(r, 1)
		if err != nil {
			return 0, err
		}

		base = v
	case "100m_per_2d10":
		// 100 million years / 2d10. 2d10 = sum of two d10 rolls.
		d1 := r.Roll("d10")
		d2 := r.Roll("d10")

		sum := d1 + d2
		if sum == 0 {
			sum = 1
		}

		base = 0.1 / float64(sum) // 100 Myr = 0.1 Gyr
	case "10m_per_2d10":
		d1 := r.Roll("d10")
		d2 := r.Roll("d10")

		sum := d1 + d2
		if sum == 0 {
			sum = 1
		}

		base = 0.01 / float64(sum) // 10 Myr = 0.01 Gyr
	default:
		return 0, fmt.Errorf("stars: unknown base formula %q", row.BaseFormula)
	}

	if !row.AddProgenitorAge {
		return base, nil
	}

	d3 := r.Roll("D3")
	progenitorMass := float64(2+d3) * deadStarMass
	progenitorAge := FinalAgeProgenitor(progenitorMass)

	return base + progenitorAge, nil
}
