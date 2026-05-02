// Package stars implements WBH star generation procedures.
package stars

import (
	"fmt"
	"math"

	"wbh/roller"
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
