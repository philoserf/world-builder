// Package worlds — biology (native lifeform ratings + resource rating)
// per WBH pp.127-131 (sub-project 3B-biology).
package worlds

import (
	"math"

	"wbh/roller"
)

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
//   - Terrestrial bodies without Atmosphere data: biology not generated;
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
// Applies WBH p.127 Special Case 1: a body with biologic-taint
// atmosphere (Taints contains a "B" entry) and rolled biomass = 0
// is promoted to biomass = 1.
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

	// Special Case 1 (WBH p.127): biologic-taint forces biomass ≥ 1.
	// Track whether biomass was rolled (not promoted) for the SC2 gate below.
	rolledPositive := biomass >= 1
	if biomass == 0 && HasTaintCode(body.Atmosphere.Taints, "B") {
		biomass = 1
	}

	// Exotic-atm bonus (rolled biomass ≥ 1 on atm 0/1/A/B/C/F+).
	// Only fires when the roll itself produced a positive result (SC2 says
	// "rolled biomass ≥ 1"); a biomass that was promoted from 0 by SC1 does
	// not qualify.
	if rolledPositive && exoticBiomassBonusApplies(body.Atmosphere.Code) {
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
//   - Low-oxygen taint ("L"): -2
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
// Applies WBH p.129 low-oxygen-taint DM-2 when atm has an "L" taint.
func RollBiocomplexity(r roller.Roller, body *DetailedPlacement, biomass int, ageGyr float64) int {
	if biomass <= 0 {
		return 0
	}
	dm := biocomplexityAgeDM(ageGyr)
	if !atmIs4to9(body) {
		dm += -2
	}
	// Low-oxygen-taint DM-2 per WBH p.129.
	if body != nil && body.Atmosphere != nil && HasTaintCode(body.Atmosphere.Taints, "L") {
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

// RollExtinctSophont per WBH p.130: evidence of an extinct native sophont
// existed if 2D + min(Biocomplexity, 9) - 7 + DMs ≥ 13.
//
// DMs: +1 if system age > 5 Gyrs.
//
// Returns false without consuming dice if biocomplexity < 8. Independent
// of RollNativeSophont — both can be true (extant species with extinct
// ancestors).
func RollExtinctSophont(r roller.Roller, biocomplexity int, ageGyr float64) bool {
	if biocomplexity < 8 {
		return false
	}
	dm := 0
	if ageGyr > 5 {
		dm = 1
	}
	bx := min(biocomplexity, 9)
	roll := r.Roll("2D")
	return roll+bx-7+dm >= 13
}

// RollBiodiversity per WBH p.130: ceil(2D - 7 + (Biomass + Biocomplexity) / 2).
//
// Returns 0 without consuming dice if biomass <= 0.
// Result < 1 promoted to 1 (when biomass > 0).
//
// The (B + X) / 2 is float division; the final ceil rounds the entire
// expression. So Biomass=10 + Biocomplexity=5 with 2D=6 produces
// 6 - 7 + 7.5 = 6.5 → ceil → 7.
func RollBiodiversity(r roller.Roller, biomass, biocomplexity int) int {
	if biomass <= 0 {
		return 0
	}
	roll := r.Roll("2D")
	v := float64(roll) - 7 + float64(biomass+biocomplexity)/2.0
	result := int(math.Ceil(v))
	if result < 1 {
		return 1
	}
	return result
}

// RollCompatibility per WBH p.130-131:
//
//	floor(2D - Biocomplexity/2 + DMs)
//
// Result clamped to ≥ 0. Caller is responsible for the biomass > 0
// prerequisite — this function returns the formula result for any inputs.
//
// DMs:
//   - Atmosphere 0, 1, B: -8
//   - Atmosphere 2, 4, 7, 9: -2
//   - Atmosphere 3, 5, 8: +1
//   - Atmosphere 6: +2
//   - Atmosphere A, F: -6
//   - Atmosphere C: -10
//   - Atmosphere D, E: -1
//   - Age > 8 Gyrs: -2
//
// NOTE: WBH p.131 worked example shows 7 + 3 - 2.5 + 2 = 9.5 for Zed Prime
// (giving compatibility 9), but the formula box has no "+3" addend.
// Implementation follows the formula → Zed Prime compatibility = 6.
// This divergence is documented as a feedback memory after merge.
//
// Atm codes G/H mentioned in the book DM table don't exist in the WBH
// 0-F atm system. They cannot be produced by RollAtmoCode — no DM applied.
//
// Skipped: "or otherwise tainted" qualifier on the -2 row deferred per
// spec Q3-a (Atmosphere taint typology not yet modeled).
func RollCompatibility(r roller.Roller, body *DetailedPlacement, biocomplexity int, ageGyr float64) int {
	dm := compatibilityAtmDM(body)
	if ageGyr > 8 {
		dm += -2
	}
	roll := r.Roll("2D")
	v := float64(roll) - float64(biocomplexity)/2.0 + float64(dm)
	result := int(math.Floor(v))
	if result < 0 {
		return 0
	}
	return result
}

// compatibilityAtmDM per WBH p.131 atmosphere-DM table.
func compatibilityAtmDM(body *DetailedPlacement) int {
	if body == nil || body.Atmosphere == nil {
		return 0
	}
	switch body.Atmosphere.Code {
	case 0, 1, 11: // 0, 1, B (G and H don't exist in our system)
		return -8
	case 2, 4, 7, 9:
		return -2
	case 3, 5, 8:
		return +1
	case 6:
		return +2
	case 10, 15: // A, F
		return -6
	case 12: // C
		return -10
	case 13, 14: // D, E
		return -1
	}
	return 0
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

// Profile returns the WBH p.131 native-lifeform-profile MXDC eHex string
// (Biomass / Biocomplexity / Biodiversity / Compatibility). Returns ""
// when receiver is nil or Biomass is 0 (no native life to profile).
//
// eHex encoding: 0-9 → "0"-"9"; 10-15 → "A"-"F"; values > 15 saturate to "F".
func (b *Biology) Profile() string {
	if b == nil || b.Biomass == 0 {
		return ""
	}
	return string([]byte{
		eHexDigit(b.Biomass),
		eHexDigit(b.Biocomplexity),
		eHexDigit(b.Biodiversity),
		eHexDigit(b.Compatibility),
	})
}

// eHexDigit converts an int to a single eHex character byte. 0-9 → '0'-'9';
// 10-15 → 'A'-'F'; values > 15 saturate to 'F'; negative values clamp to '0'.
func eHexDigit(n int) byte {
	if n < 0 {
		return '0'
	}
	if n > 15 {
		return 'F'
	}
	if n < 10 {
		return byte('0' + n)
	}
	return byte('A' + (n - 10))
}

// RollTerrestrialResourceRating per WBH p.131:
//
//	2D - 7 + Size + DMs, clamped to [2, 12]
//
// Runs for ALL terrestrial bodies regardless of biology — biology DMs
// only apply when ratings are non-zero.
//
// DMs:
//   - Density > 1.12: +2
//   - Density < 0.5: -2
//   - Biomass ≥ 3: +2
//   - Biodiversity 8-10 (8-A): +1
//   - Biodiversity ≥ 11 (B+): +2
//   - Compatibility 0-3: -1 (only if Biomass ≥ 1)
//   - Compatibility ≥ 8: +2
//
// bio may be a zero Biology{} for bodies without life.
func RollTerrestrialResourceRating(r roller.Roller, body *DetailedPlacement, bio *Biology) int {
	dm := 0
	if body != nil && body.Physical != nil {
		switch {
		case body.Physical.Density > 1.12:
			dm += 2
		case body.Physical.Density < 0.5:
			dm += -2
		}
	}
	if bio != nil {
		if bio.Biomass >= 3 {
			dm += 2
		}
		switch {
		case bio.Biodiversity >= 11:
			dm += 2
		case bio.Biodiversity >= 8:
			dm++
		}
		if bio.Compatibility >= 0 && bio.Compatibility <= 3 && bio.Biomass >= 1 {
			dm += -1
		}
		if bio.Compatibility >= 8 {
			dm += 2
		}
	}
	roll := r.Roll("2D")
	size := 0
	if body != nil {
		size = SizeAsInt(body.SizeCode)
	}
	result := roll - 7 + size + dm
	return min(max(result, 2), 12)
}

// BiocomplexityName returns the WBH p.129 biocomplexity descriptor.
// Source: WBH p.129 "Biocomplexity Rating" table — book-quoted descriptions
// for ratings 1-9; rating 10+ uses the book's "A (10)+" row description.
// Returns "" for rating 0 (no biomass → no biocomplexity rendered).
func BiocomplexityName(c int) string {
	switch {
	case c <= 0:
		return ""
	case c == 1:
		return "Primitive single-cell organisms"
	case c == 2:
		return "Advanced cellular organisms"
	case c == 3:
		return "Primitive multicellular organisms"
	case c == 4:
		return "Differentiated multicellular organisms"
	case c == 5:
		return "Complex multicellular organisms"
	case c == 6:
		return "Advanced multicellular organisms"
	case c == 7:
		return "Socially advanced organisms"
	case c == 8:
		return "Mentally advanced organisms"
	case c == 9:
		return "Extant or extinct sophonts"
	default: // 10+
		return "Ecosystem-wide superorganisms"
	}
}

// BiomassLabel returns a project-supplied descriptor anchored on WBH pp.127-128
// prose anchors: "0 = no native life" and "A (10)+ = healthy garden world".
// Intermediate bands (1, 2-3, 4-5, 6-7, 8-9) are interpolated to give a
// graded sense of life density on the book's exponential scale.
func BiomassLabel(b int) string {
	switch {
	case b <= 0:
		return "None (sterile)" // book: "no native life"
	case b == 1:
		return "Trace" // interpolated
	case b <= 3:
		return "Sparse" // interpolated
	case b <= 5:
		return "Moderate" // interpolated
	case b <= 7:
		return "Abundant" // interpolated
	case b <= 9:
		return "Lush" // interpolated
	default: // 10+
		return "Garden world" // book: "healthy garden world"
	}
}

// BiodiversityLabel returns a project-supplied descriptor anchored on WBH p.130
// prose anchors: "<3 = very uniform biosphere with limited species dominating"
// and "A (10)+ = pre-human Terra-like complexity".
func BiodiversityLabel(b int) string {
	switch {
	case b <= 0:
		return "None"
	case b <= 2:
		return "Uniform (few dominant species)" // book: "very uniform biosphere"
	case b <= 5:
		return "Limited" // interpolated
	case b <= 7:
		return "Moderate" // interpolated
	case b <= 9:
		return "Diverse" // interpolated
	default: // 10+
		return "Terra-like complexity" // book: "pre-human Terra-like"
	}
}

// CompatibilityLabel returns a project-supplied descriptor anchored on WBH
// pp.130-131 prose anchors: "0 = incompatible" and "A (10) = full Terran
// compatibility". Values >10 are possible and represent specially nutritious
// or beneficial native biology per book footnote.
func CompatibilityLabel(c int) string {
	switch {
	case c <= 0:
		return "Incompatible" // book: "all native life is incompatible"
	case c <= 3:
		return "Marginal" // interpolated
	case c <= 6:
		return "Partial" // interpolated
	case c <= 9:
		return "Mostly compatible" // interpolated
	case c == 10:
		return "Fully Terran-compatible" // book: "equivalent to full Terran compatibility"
	default: // 11+
		return "Beyond Terran" // book footnote: "greater values are possible"
	}
}
