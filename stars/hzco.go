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

// CompositeHZCO returns the HZCO# for a circumbinary group of stars
// orbiting a shared barycentre. Per WBH p. 42, the luminosities of all
// stars interior to the planet's orbit are summed, then the formula
// applies to the combined luminosity.
//
// Empty input returns 0.
func CompositeHZCO(starsInterior ...Star) float64 {
	var totalL float64
	for _, s := range starsInterior {
		totalL += s.Luminosity
	}

	if totalL <= 0 {
		return 0
	}

	return AUToOrbit(math.Sqrt(totalL))
}
