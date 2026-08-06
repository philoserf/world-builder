package worlds

import (
	"math"

	"github.com/philoserf/world-builder/roller"
)

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

	return auResult, pd
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
// The two outcomes never diverge in the book's rule, so a single bool
// covers both; the caller sets Body.Ring when it fires.
func MoonRemovalCheck(hillSphereMoonLimit float64) bool {
	return hillSphereMoonLimit < 1.5
}

// MoonOrbitRange returns the Moon Orbit Range (MOR) per WBH p.77:
//
//	MOR = floor(hillSphereMoonLimit) - 2
//
// If MOR > 200, it is clamped to 200 + nMoons (the number of moons to be placed).
// MOR is floored at 0.
func MoonOrbitRange(hillSphereMoonLimit float64, nMoons int) int {
	mor := int(math.Floor(hillSphereMoonLimit)) - 2
	if mor > 200 {
		mor = 200 + nMoons
	}

	if mor < 0 {
		mor = 0
	}

	return mor
}

// MoonPeriodHours computes a moon's orbital period in hours per WBH p.78:
//
//	Period (hours) = 0.176927 × √((PD × effectiveSize)³ / Mp)
//
// effectiveSize is the parent body's diameter expressed in units of 1600 km
// (i.e. diameterMultiplier × sizeCode for gas giants, or just sizeCode for
// terrestrial worlds). For example, Aab IV is a Size-8 GG with diameter
// multiplier 14, so effectiveSize = 14 × 8 = 112.
//
// Mp is the parent mass in Earth masses.
// Returns 0 if Mp ≤ 0.
func MoonPeriodHours(orbitPD float64, effectiveSize int, parentMassEarth float64) float64 {
	if parentMassEarth <= 0 {
		return 0
	}

	d := orbitPD * float64(effectiveSize)
	cube := d * d * d

	return 0.176927 * math.Sqrt(cube/parentMassEarth)
}

// RollRingProfile rolls a significant ring's centre location and span in
// planet-diameters (PD) per WBH p.77:
//
//	Ring Centre Location (PD) = 0.4 + 2D ÷ 8
//	Ring Span (PD)            = 3D ÷ 100 + 0.07
//
// The model produces a single ring, so the book's multi-ring overlap
// adjustments (two-ring push-out, surface-intersection narrowing) do not
// apply. Rendered as the ring profile R01:centre-span.
func RollRingProfile(r roller.Roller) (centrePD, spanPD float64) {
	centrePD = 0.4 + float64(r.Roll("2D"))/8.0
	spanPD = float64(r.Roll("3D"))/100.0 + 0.07

	return centrePD, spanPD
}

// MoonRange categorizes a moon by its orbit relative to MOR.
type MoonRange int

// Moon orbit range bands relative to the Moon Orbit Range (MOR).
const (
	MoonRangeInner  MoonRange = iota // inner sixth of the MOR
	MoonRangeMiddle                  // middle third of the MOR
	MoonRangeOuter                   // outer half of the MOR
)

// RollMoonOrbit determines a moon's orbital distance per WBH p.77
// Moon Orbit Location table. Roll 1D (DM+1 if MOR < 60) for range,
// then 2D for the specific orbit in Planetary Diameters (PD):
//
//	1–3 Inner:  (2D-2) × MOR ÷ 60 + 2
//	4–5 Middle: (2D-2) × MOR ÷ 30 + MOR ÷ 6 + 3
//	6+  Outer:  (2D-2) × MOR ÷ 20 + MOR ÷ 2 + 4
//
// Returns orbit in PD and the range classification.
func RollMoonOrbit(r roller.Roller, mor int) (orbitPD float64, mr MoonRange) {
	dm := 0
	if mor < 60 {
		dm = 1
	}

	rng := r.Roll("1D") + dm
	v := r.Roll("2D")

	switch {
	case rng <= 3:
		mr = MoonRangeInner
		orbitPD = float64(v-2)*float64(mor)/60 + 2
	case rng <= 5:
		mr = MoonRangeMiddle
		orbitPD = float64(v-2)*float64(mor)/30 + float64(mor)/6 + 3
	default:
		mr = MoonRangeOuter
		orbitPD = float64(v-2)*float64(mor)/20 + float64(mor)/2 + 4
	}

	return orbitPD, mr
}

// RollMoonRetrograde determines whether a moon has a retrograde orbit per WBH p.78.
// Roll 2D and apply DM by range: inner −1, middle +1, outer +4.
// Add DM+6 if the moon's orbit exceeds the MOR.
// An orbit is retrograde on a result of 10+.
func RollMoonRetrograde(r roller.Roller, mr MoonRange, exceedsMOR bool) bool {
	dm := 0

	switch mr {
	case MoonRangeInner:
		dm = -1
	case MoonRangeMiddle:
		dm = 1
	case MoonRangeOuter:
		dm = 4
	}

	if exceedsMOR {
		dm += 6
	}

	return r.Roll("2D")+dm >= 10
}
