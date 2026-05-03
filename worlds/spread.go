package worlds

// Spread implements WBH Step 5 (pp. 48-49). Returns the average Orbit#
// separation between slots inside the system.
//
// Base formula: (baselineOrbit - primary.MAO) / max(baselineN, 1).
//
// When applying the base spread would place the outermost primary slot
// past Orbit# 20, the mandatory Maximum Spread cap (p. 48) is used:
//
//	primary.Total() / (primaryAllocated + totalStars)
//
// totalStars counts primary + Close/Near/Far secondaries. Companions do
// not count.
func Spread(primary Group, primaryAllocated int, baselineOrbit float64, baselineN, totalStars int) float64 {
	n := baselineN
	if n < 1 {
		n = 1
	}
	base := (baselineOrbit - primary.MAO) / float64(n)
	// Outermost slot would be at MAO + primaryAllocated × base. The cap
	// fires when this exceeds Orbit# 20, including the baselineN < 1 case
	// the book explicitly calls out on p. 48.
	if primary.MAO+float64(primaryAllocated)*base <= 20.0 {
		return base
	}
	if primaryAllocated+totalStars == 0 {
		return base
	}
	return primary.Total() / float64(primaryAllocated+totalStars)
}

// MaximumSecondarySpread implements the WBH p. 49 per-secondary cap:
//
//	(secondary.Outermost - secondary.MAO) / (secondaryAllocated + 1)
//
// secondary.Outermost is the maximum endpoint across secondary.Intervals.
// PlaceOrbitSlots applies this when the system spread would push the
// secondary's outermost slot past its outer allowable bound.
func MaximumSecondarySpread(secondary Group, secondaryAllocated int) float64 {
	var outer float64
	for _, iv := range secondary.Intervals {
		if iv.Max > outer {
			outer = iv.Max
		}
	}
	return (outer - secondary.MAO) / float64(secondaryAllocated+1)
}
