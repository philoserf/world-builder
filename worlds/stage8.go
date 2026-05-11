package worlds

import (
	"wbh/roller"
)

// ApplyBiology runs the WBH pp.127-131 biology pass: per-body biomass,
// biocomplexity, sophonts, biodiversity, compatibility, terrestrial
// resource rating. Stage-8 orchestrator.
//
// Per dependency-graph.md § Stage 8, the optional any-oxygen-atm
// biomass floor is cut (design-intent.md cuts list). Per anti-pattern
// A.1, every body is walked uniformly via Universe.AllBodies() — a
// terrestrial moon with atmosphere gets biology regardless of its
// parent type (Zed Prime canonical example: moon of Aab IV gas giant).
func ApplyBiology(r roller.Roller, u *Universe) error {
	ageGyr := u.System.Primary.AgeGyr
	for body := range u.AllBodies() {
		if !biologyApplies(body) {
			continue
		}
		body.Biology = computeBiology(r, body, ageGyr)
	}
	return nil
}

// biologyApplies reports whether body should receive a Biology struct.
// True for terrestrial bodies with Atmosphere; false for empty, belts,
// gas giants, or atmosphereless terrestrials.
func biologyApplies(body *Body) bool {
	if body == nil || body.Kind == BodyEmpty {
		return false
	}
	if body.Kind == BodyGasGiant || body.Kind == BodyPlanetoidBelt {
		return false
	}
	if body.SizeCode == "0" {
		return false
	}
	return body.Atmosphere != nil
}

// computeBiology populates a Biology for the given body. Caller has
// already verified biologyApplies(body).
func computeBiology(r roller.Roller, body *Body, ageGyr float64) *Biology {
	bio := &Biology{}
	bio.Biomass = RollBiomass(r, body, ageGyr)
	if bio.Biomass > 0 {
		bio.Biocomplexity = RollBiocomplexity(r, body, bio.Biomass, ageGyr)
		if bio.Biocomplexity >= 8 {
			bio.HasNativeSophont = RollNativeSophont(r, bio.Biocomplexity)
			bio.HadExtinctSophont = RollExtinctSophont(r, bio.Biocomplexity, ageGyr)
		}
		bio.Biodiversity = RollBiodiversity(r, bio.Biomass, bio.Biocomplexity)
		bio.Compatibility = RollCompatibility(r, body, bio.Biocomplexity, ageGyr)
	}
	bio.ResourceRating = RollTerrestrialResourceRating(r, body, bio)
	return bio
}
