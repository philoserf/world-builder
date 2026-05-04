package worlds

import "math"

const auKm = 149597870.9 // km per AU

// HillSphere computes the Hill sphere of a planet around its primary stars,
// returning both the AU result and the conversion to planetary diameters (PD).
// Per WBH p.75:
//
//	Hill Sphere (AU) = AU × (1 - ecc) × ³√(planetMassSolar / (3 × stellarMassSolar))
//	  where planetMassSolar = planetMassEarth × 0.000003
//	Hill Sphere (PD) = HillSphereAU × 149,597,870.9 / planet_diameter_km
func HillSphere(au, ecc, planetMassEarth, sumStellarMassSolar, planetDiameterKm float64) (auResult, pd float64) {
	if sumStellarMassSolar <= 0 || planetDiameterKm <= 0 {
		return 0, 0
	}
	planetMassSolar := planetMassEarth * 0.000003
	ratio := planetMassSolar / (3 * sumStellarMassSolar)
	cube := math.Cbrt(ratio)
	auResult = au * (1 - ecc) * cube
	pd = auResult * auKm / planetDiameterKm
	return
}

// HillSphereMoonLimit returns the prograde-moon outer bound: HillSpherePD ÷ 2,
// rounded down (WBH p.75-76).
func HillSphereMoonLimit(hillSpherePD float64) float64 {
	return math.Floor(hillSpherePD / 2)
}

// RocheLimit computes the Roche limit per WBH p.76:
//
//	Roche Limit = 1.22 × planet_diameter × ³√(planet_density / moon_density)
//
// Returns 0 if moon_density is non-positive.
func RocheLimit(planetDiameterKm, planetDensityRel, moonDensityRel float64) float64 {
	if moonDensityRel <= 0 {
		return 0
	}
	return 1.22 * planetDiameterKm * math.Cbrt(planetDensityRel/moonDensityRel)
}

// MoonRemovalCheck implements WBH p.76: if HillSphereMoonLimit < 1.5 PD,
// all significant moons are removed and the first is promoted to a ring.
// Returns (removeAll, promoteFirstToRing).
func MoonRemovalCheck(hillSphereMoonLimit float64) (removeAll, promoteFirstToRing bool) {
	if hillSphereMoonLimit < 1.5 {
		return true, true
	}
	return false, false
}
