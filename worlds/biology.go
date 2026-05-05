// Package worlds — biology (native lifeform ratings + resource rating)
// per WBH pp.127-131 (sub-project 3B-biology).
package worlds

import "wbh/roller"

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

// RollBiomass per WBH p.127-128: 2D + DMs (combined DM sum clamped to
// [-12, +4]). Includes the exotic-atm bonus (Special Case 2): if the
// rolled biomass is ≥ 1 AND atm code ∈ {0, 1, 10, 11, 12, 15}, add
// (|atm DM| − 1) to the result.
//
// Returns 0 if body or body.Atmosphere is nil. nil Hydrographics is
// treated as Hydro 0 (DM-4). nil Temperature contributes no temp DMs.
//
// Skipped: Special Case 1 (biologic-taint biomass=0 promotion) requires
// Atmosphere taint typology not yet modeled — deferred per spec Q3-a.
func RollBiomass(r roller.Roller, body *DetailedPlacement, ageGyr float64) int {
	if body == nil || body.Atmosphere == nil {
		return 0
	}
	atmDM := biomassAtmDM(body.Atmosphere.Code)
	hydroCode := 0
	if body.Hydrographics != nil {
		hydroCode = body.Hydrographics.Code
	}
	dm := atmDM + biomassHydroDM(hydroCode) + biomassAgeDM(ageGyr)
	if body.Temperature != nil {
		dm += biomassTempDM(body.Temperature.MeanK, body.Temperature.HighK)
	}
	dm = min(max(dm, -12), 4)

	roll := r.Roll("2D")
	biomass := max(roll+dm, 0)

	// Exotic-atm bonus (rolled biomass ≥ 1 on atm 0/1/A/B/C/F+).
	if biomass >= 1 && exoticBiomassBonusApplies(body.Atmosphere.Code) {
		biomass += exoticBiomassBonus(body.Atmosphere.Code)
	}
	return biomass
}

// biomassAtmDM per WBH p.128 atmosphere-DM table.
func biomassAtmDM(atmCode int) int {
	switch atmCode {
	case 0:
		return -6
	case 1:
		return -4
	case 2, 3, 14: // E
		return -3
	case 4, 5:
		return -2
	case 8, 9, 13: // D
		return +2
	case 10: // A
		return -3
	case 11: // B
		return -5
	case 12: // C
		return -7
	case 15: // F+
		return -5
	}
	return 0 // atm 6, 7, or unmapped
}

// biomassHydroDM per WBH p.128 hydrographics-DM table.
func biomassHydroDM(hydroCode int) int {
	switch {
	case hydroCode == 0:
		return -4
	case hydroCode >= 1 && hydroCode <= 3:
		return -2
	case hydroCode >= 6 && hydroCode <= 8:
		return +1
	case hydroCode >= 9:
		return +2
	}
	return 0 // 4-5
}

// biomassAgeDM per WBH p.128 age-DM table.
func biomassAgeDM(ageGyr float64) int {
	switch {
	case ageGyr < 0.2:
		return -6
	case ageGyr < 1:
		return -2
	case ageGyr > 4:
		return +1
	}
	return 0
}

// biomassTempDM per WBH p.128 temperature-DM table. Returns 0 when both
// MeanK and HighK are 0 (defensive — caller already gated on nil temp).
func biomassTempDM(meanK, highK float64) int {
	dm := 0
	if highK > 353 {
		dm += -2
	} else if highK > 0 && highK < 273 {
		dm += -4
	}
	if meanK > 353 {
		dm += -4
	} else if meanK > 0 && meanK < 273 {
		dm += -2
	}
	if meanK >= 279 && meanK <= 303 {
		dm += +2
	}
	return dm
}

// exoticBiomassBonusApplies reports whether atm code is in {0, 1, A, B, C, F+}.
func exoticBiomassBonusApplies(atmCode int) bool {
	return atmCode == 0 || atmCode == 1 ||
		atmCode == 10 || atmCode == 11 || atmCode == 12 || atmCode == 15
}

// exoticBiomassBonus returns |atmDM| − 1 for the exotic atm codes per WBH
// Special Case 2: "Add one less than the negative Atmosphere DM".
func exoticBiomassBonus(atmCode int) int {
	switch atmCode {
	case 0:
		return 5
	case 1:
		return 3
	case 10:
		return 2
	case 11:
		return 4
	case 12:
		return 6
	case 15:
		return 4
	}
	return 0
}

// RollBiocomplexity per WBH p.129: 2D - 7 + min(Biomass, 9) + DMs.
// Returns 0 without consuming dice if biomass <= 0.
//
// DMs:
//   - Atmosphere not 4-9: -2
//   - Age 3-4 Gyrs: -2
//   - Age 2-3 Gyrs: -4
//   - Age 1-2 Gyrs: -8
//   - Age < 1 Gyr: -10
//
// At age boundaries (e.g., age=2.0), uses the worst (more negative) DM
// per WBH "If the system age is exactly at a limit between two DMs, use
// the worst DM." Implemented via ordered-case switch with inclusive
// upper bounds.
//
// Result < 1 promoted to 1 (when biomass > 0).
//
// Skipped: low-oxygen-taint DM-2 deferred per spec Q3-a (taint typology
// not yet modeled).
func RollBiocomplexity(r roller.Roller, body *DetailedPlacement, biomass int, ageGyr float64) int {
	if biomass <= 0 {
		return 0
	}
	dm := biocomplexityAgeDM(ageGyr)
	if !atmIs4to9(body) {
		dm += -2
	}
	roll := r.Roll("2D")
	bx := min(biomass, 9)
	result := roll - 7 + bx + dm
	if result < 1 {
		return 1
	}
	return result
}

// biocomplexityAgeDM per WBH p.129. Uses the worst DM at boundaries.
func biocomplexityAgeDM(ageGyr float64) int {
	switch {
	case ageGyr <= 1:
		return -10
	case ageGyr <= 2:
		return -8
	case ageGyr <= 3:
		return -4
	case ageGyr <= 4:
		return -2
	}
	return 0
}

// RollNativeSophont per WBH p.130: extant native sophont exists if
// 2D + min(Biocomplexity, 9) - 7 ≥ 13.
//
// Returns false without consuming dice if biocomplexity < 8.
func RollNativeSophont(r roller.Roller, biocomplexity int) bool {
	if biocomplexity < 8 {
		return false
	}
	bx := min(biocomplexity, 9)
	roll := r.Roll("2D")
	return roll+bx-7 >= 13
}

// atmIs4to9 reports whether body.Atmosphere.Code is in [4, 9]. Returns
// false on nil atmosphere.
func atmIs4to9(body *DetailedPlacement) bool {
	if body == nil || body.Atmosphere == nil {
		return false
	}
	c := body.Atmosphere.Code
	return c >= 4 && c <= 9
}
