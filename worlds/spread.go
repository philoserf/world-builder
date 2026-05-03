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
	// When baselineN < 1 the spread degenerates (no baseline denominator);
	// clamp to 1 and return the base formula without cap checking — there
	// is no meaningful "outermost slot" to test against Orbit# 20.
	if baselineN < 1 {
		return (baselineOrbit - primary.MAO) / 1.0
	}
	base := (baselineOrbit - primary.MAO) / float64(baselineN)
	// Outermost slot would be at MAO + primaryAllocated × base.
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
	if secondaryAllocated+1 == 0 {
		return 0
	}
	return (outer - secondary.MAO) / float64(secondaryAllocated+1)
}
