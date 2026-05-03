package worlds

import "wbh/stars"

// Period — orbital period; both Years and Days are populated and the
// renderer picks based on magnitude (form p.63 uses days for periods
// shorter than ~0.05y, otherwise years).
//
// WBH p.53 ("Length of 'Years'") gives three forms:
//
//	Single star:        P = sqrt(AU^3 / M☉)
//	Multiple stars:     P = sqrt(AU^3 / Σ M☉)
//	Large planet:       P = sqrt(AU^3 / (Σ M☉ + m⊕ × 0.000003))
//
// All three reduce to one call to stars.OrbitPeriodYears(au, sumMass, m)
// where m = bodyMassEarth × 0.000003 (or 0 for the standard cases).
type Period struct {
	Years float64 // primary representation; from Kepler's 3rd
	Days  float64 // = Years * 365.25
}

// massSolarPerEarth is the WBH p.53 "Large Planet" mass-conversion
// factor: 1 Terra mass in solar units.
const massSolarPerEarth = 0.000003

// PeriodFor computes a Period for a body at orbit (au) given the sum
// of stellar masses interior to that orbit (sumStellarMassSolar) and
// the body's mass in Terra masses (bodyMassEarth, 0 for the standard
// formula). Wraps stars.OrbitPeriodYears.
func PeriodFor(au, sumStellarMassSolar, bodyMassEarth float64) Period {
	years := stars.OrbitPeriodYears(au, sumStellarMassSolar, bodyMassEarth*massSolarPerEarth)
	return Period{
		Years: years,
		Days:  years * 365.25,
	}
}
