// Package worlds — biology (native lifeform ratings + resource rating)
// per WBH pp.127-131 (sub-project 3B-biology).
package worlds

// Biology — native lifeform ratings + resource rating per WBH pp.127-131.
// Populated by Step 5F for terrestrial bodies (and their HZ-planet moons)
// that have Atmosphere data.
//
// Conditional applicability:
//   - Bodies with Biomass == 0: only ResourceRating populated; biology
//     ratings (Biocomplexity, Biodiversity, Compatibility) stay 0; sophont
//     bools stay false.
//   - Bodies with Biomass >= 1 but Biocomplexity < 8: sophont bools stay
//     false (prerequisite for sophont rolls is Biocomplexity >= 8).
//   - Belts (Size 0), gas giants, empty placements: biology not generated;
//     dp.Biology stays nil.
type Biology struct {
	// 2D + DMs, with combined-DM sum clamped to [-12, +4] per WBH p.127.
	// Range 0-15 (eHex 0-F); 0 = no native life.
	Biomass int

	// 2D - 7 + Biomass + DMs per WBH p.129. Zero if Biomass == 0.
	// Result < 1 promoted to 1 (when biomass > 0). Range 0-15.
	Biocomplexity int

	// True if extant native sophont species exists; 2D + Biocomplexity - 7
	// >= 13 per WBH p.130. False if Biocomplexity < 8.
	HasNativeSophont bool

	// True if evidence of an extinct native sophont species; 2D + Biocomplexity
	// - 7 + DMs >= 13 per WBH p.130. False if Biocomplexity < 8. Independent
	// of HasNativeSophont — both can be true.
	HadExtinctSophont bool

	// ceil(2D - 7 + (Biomass + Biocomplexity) / 2) per WBH p.130.
	// Zero if Biomass == 0. Result < 1 promoted to 1 (when biomass > 0).
	Biodiversity int

	// floor(2D - Biocomplexity/2 + DMs) per WBH p.130-131. Zero if Biomass == 0
	// or if rolled result <= 0. Range 0-15+ (10 = full Terran, > 10 possible).
	Compatibility int

	// 2D - 7 + Size + DMs per WBH p.131. Computed for ALL terrestrial
	// bodies regardless of biology (biology DMs only apply when applicable).
	// Range [2, 12] per WBH lower/upper bounds.
	ResourceRating int
}
