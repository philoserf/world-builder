package worlds

import (
	"wbh/roller"
	"wbh/stars"
)

// runStep5F applies the 3B-biology pass: native lifeform ratings (Biomass,
// Biocomplexity, Sophont bools, Biodiversity, Compatibility) + Resource
// Rating. Mutates detailed in place. WBH pp.127-131.
//
// Body filter: terrestrials with Atmosphere; HZ-planet moons with
// Atmosphere. Skip GGs, belts (Size 0), empty placements, and terrestrials
// without atmosphere data.
//
// Per-body dice budget: up to 7 × 2D when life is present; 2 × 2D when
// biomass = 0 (Biomass + Resource); 0 × 2D when body is skipped.
//
//nolint:unparam // matches sibling runStep5* signatures (always-nil error)
func runStep5F(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error {
	for i := range detailed {
		dp := &detailed[i]
		if !biologyApplies(dp) {
			continue
		}
		dp.Biology = computeBiology(r, dp, sys.Primary.AgeGyr)

		for j := range dp.Moons {
			m := &dp.Moons[j]
			if m.Atmosphere == nil {
				continue
			}
			moonDP := buildMoonPlacementView(m, dp)
			// buildMoonPlacementView does not copy Temperature; set it
			// explicitly so RollBiomass can read m.Temperature for the
			// temperature DMs (precedent: same pattern in computeMoonGeology
			// for TidalEffects per 3B-geology Task 8).
			moonDP.Temperature = m.Temperature
			m.Biology = computeBiology(r, moonDP, sys.Primary.AgeGyr)
		}
	}
	return nil
}

// biologyApplies reports whether dp should receive a Biology struct.
// True for terrestrial bodies with Atmosphere; false for empty, belts,
// gas giants, or atmosphere-less terrestrials.
func biologyApplies(dp *DetailedPlacement) bool {
	if dp == nil || dp.Body == BodyEmpty {
		return false
	}
	if dp.Body == BodyGasGiant || dp.Body == BodyPlanetoidBelt {
		return false
	}
	if dp.SizeCode == "0" {
		return false
	}
	return dp.Atmosphere != nil
}

// computeBiology populates a Biology for the given body. Caller has
// already verified biologyApplies(dp).
func computeBiology(r roller.Roller, dp *DetailedPlacement, ageGyr float64) *Biology {
	bio := &Biology{}
	bio.Biomass = RollBiomass(r, dp, ageGyr)
	if bio.Biomass > 0 {
		bio.Biocomplexity = RollBiocomplexity(r, dp, bio.Biomass, ageGyr)
		if bio.Biocomplexity >= 8 {
			bio.HasNativeSophont = RollNativeSophont(r, bio.Biocomplexity)
			bio.HadExtinctSophont = RollExtinctSophont(r, bio.Biocomplexity, ageGyr)
		}
		bio.Biodiversity = RollBiodiversity(r, bio.Biomass, bio.Biocomplexity)
		bio.Compatibility = RollCompatibility(r, dp, bio.Biocomplexity, ageGyr)
	}
	bio.ResourceRating = RollTerrestrialResourceRating(r, dp, bio)
	return bio
}
