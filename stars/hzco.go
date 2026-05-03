package stars

import "math"

// HZCO returns the Habitable Zone Centre Orbit# for a single star,
// computed from its luminosity by the WBH p. 41 formula:
//
//	HZCO_AU    = sqrt(luminosity)
//	HZCO_Orbit = AUToOrbit(HZCO_AU)
//
// The p. 42 HZCO table is encoded as a test fixture only; this function
// uses the formula path which the book itself validates as the canonical
// computation.
func (s Star) HZCO() float64 {
	return AUToOrbit(math.Sqrt(s.Luminosity))
}
