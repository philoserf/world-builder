package worlds

import (
	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

// Counts is the per-system count of bodies a Referee will place.
type Counts struct {
	GasGiants      int // 0–6
	PlanetoidBelts int // 0–3
	Terrestrials   int // 0–13 (cap from Step 7 narrative)
	Total          int // GasGiants + PlanetoidBelts + Terrestrials
}

// ComponentSum returns GasGiants + PlanetoidBelts + Terrestrials — the
// definition of Total, and the count of worlds actually placed. Callers
// that must not trust the separately-stored Total field (which nothing
// enforces) use this instead.
func (c Counts) ComponentSum() int {
	return c.GasGiants + c.PlanetoidBelts + c.Terrestrials
}

// CountsOpts is reserved for future knobs (e.g., the alternate "Gas Giant
// Exists on 2+: roll 1D" existence form). Empty for now; the standard
// CRB form is used.
type CountsOpts struct{}

// GenerateCounts implements WBH pp. 36–38. Returns the per-system counts
// of gas giants, planetoid belts, and terrestrial planets, and their
// total. Continuation Method (pre-existing mainworld) is out of scope.
func GenerateCounts(r roller.Roller, sys stars.System, _ CountsOpts) (Counts, error) {
	var c Counts

	// Existence: "Gas Giant Exists on 9-: roll 2D" (WBH p. 37).
	// Only the negative DMs (post-stellar, multi-star) apply to the
	// existence roll. The single-Class-V DM+1 applies to quantity only.
	existDM := gasGiantExistenceDMs(sys)
	if r.Roll("2D")+existDM <= 9 {
		// Present: roll 2D + all DMs for quantity.
		qtyDM := gasGiantQuantityDMs(sys)
		qty := r.Roll("2D") + qtyDM
		c.GasGiants = gasGiantQuantity(qty)
	}

	// Existence: "Planetoid Belt Exists on 8+: roll 2D" (WBH p. 37).
	// Per p. 37 sidebar: existence has no DMs unless Special Circumstances.
	beltExistDM := beltExistenceDMs(sys)
	if r.Roll("2D")+beltExistDM >= 8 {
		// Present: roll 2D + all DMs for quantity.
		beltQtyDM := beltQuantityDMs(sys, c.GasGiants)
		beltQty := r.Roll("2D") + beltQtyDM
		c.PlanetoidBelts = beltQuantity(beltQty)
	}

	// Terrestrials: WBH p. 38. 2D - 2 + DM-1 per post-stellar object.
	// If result < 3, reroll as D3+2. If result ≥ 3, add D3-1.
	terrDM := -postStellarCount(sys)
	raw := r.Roll("2D") - 2 + terrDM
	switch {
	case raw < 3:
		c.Terrestrials = r.Roll("D3") + 2
	default:
		c.Terrestrials = raw + r.Roll("D3") - 1
	}
	c.Total = c.ComponentSum()
	return c, nil
}

// gasGiantExistenceDMs computes the WBH p. 37 DM stack for the
// gas-giant existence roll only. Per WBH p. 37 sidebar: "There are no
// DMs to the gas giant existence roll unless the primary star is subject
// to rules covered in the Special Circumstances chapter." The single-
// Class-V DM+1 therefore applies to quantity only.
func gasGiantExistenceDMs(sys stars.System) int {
	dm := 0
	if sys.Primary.Kind == stars.KindBrownDwarf {
		dm -= 2
	}
	if isPostStellar(sys.Primary.Kind) {
		dm -= 2
	}
	dm -= postStellarCount(sys)
	if totalStarCount(sys) >= 4 {
		dm--
	}
	return dm
}

// gasGiantQuantityDMs computes the WBH p. 37 DM stack for the
// gas-giant quantity roll (all DMs including the single-Class-V DM+1).
func gasGiantQuantityDMs(sys stars.System) int {
	dm := 0
	if isSingleClassVSystem(sys) {
		dm++
	}
	if sys.Primary.Kind == stars.KindBrownDwarf {
		dm -= 2
	}
	if isPostStellar(sys.Primary.Kind) {
		dm -= 2
	}
	dm -= postStellarCount(sys)
	if totalStarCount(sys) >= 4 {
		dm--
	}
	return dm
}

// gasGiantQuantity maps a 2D+DMs result to the WBH p. 37 quantity table.
// Outputs 1–6.
func gasGiantQuantity(roll int) int {
	switch {
	case roll <= 4:
		return 1
	case roll <= 6:
		return 2
	case roll <= 8:
		return 3
	case roll <= 11:
		return 4
	case roll == 12:
		return 5
	default:
		return 6
	}
}

// isSingleClassVSystem reports whether the system has exactly one star
// (no companions of any kind) and that star is luminosity class V.
func isSingleClassVSystem(sys stars.System) bool {
	if len(sys.Companions) > 0 {
		return false
	}
	return sys.Primary.LuminosityClass == stars.V
}

// postStellarCount returns the count of post-stellar objects in the
// system, including the primary if it is post-stellar.
func postStellarCount(sys stars.System) int {
	n := 0
	if isPostStellar(sys.Primary.Kind) {
		n++
	}
	for _, c := range sys.Companions {
		if isPostStellar(c.Star.Kind) {
			n++
		}
	}
	return n
}

// totalStarCount returns the count of stellar bodies (primary +
// companions of every orbit class). Companions count.
func totalStarCount(sys stars.System) int {
	return 1 + len(sys.Companions)
}

// beltExistenceDMs computes the WBH p. 37 DM stack for the planetoid-
// belt existence roll only. Per WBH p. 37 sidebar: "There are no DMs to
// the planetoid belt existence roll unless the primary star is subject
// to rules covered in the Special Circumstances chapter." Only the
// Special-Circumstances DMs apply here; the GG-present and 2+-stars DMs
// apply to quantity only.
func beltExistenceDMs(sys stars.System) int {
	dm := 0
	if sys.Primary.Kind == stars.KindProtostar {
		dm += 3
	}
	if isPrimordial(sys.Primary) {
		dm += 2
	}
	if isPostStellar(sys.Primary.Kind) {
		dm++
	}
	// postStellarCount includes the primary; the spec lists a flat
	// post-stellar primary DM+1 AND a per-post-stellar-object DM+1, so a
	// lone post-stellar primary intentionally contributes +2 here.
	dm += postStellarCount(sys)
	return dm
}

// beltQuantityDMs computes the WBH p. 37 DM stack for the planetoid-belt
// quantity roll (all DMs).
//
// gasGiantsPresent is the freshly-computed count from the same call to
// GenerateCounts, used for the "system has ≥1 gas giants" DM.
func beltQuantityDMs(sys stars.System, gasGiantsPresent int) int {
	dm := beltExistenceDMs(sys)
	if gasGiantsPresent >= 1 {
		dm++
	}
	if totalStarCount(sys) >= 2 {
		dm++
	}
	return dm
}

// beltQuantity maps a 2D+DMs result to the WBH p. 37 belt quantity table.
func beltQuantity(roll int) int {
	switch {
	case roll <= 6:
		return 1
	case roll <= 11:
		return 2
	default:
		return 3
	}
}

// isPrimordial reports whether a star is "primordial" per WBH p. 14:
// any star with age below 0.1 Gyr. The AgeGyr > 0 guard avoids treating
// a default-zero Star{} (e.g., constructed without AgeGyr in tests) as
// primordial — the book defines primordial as a generation-time property,
// not "age unknown."
func isPrimordial(s stars.Star) bool {
	return s.AgeGyr > 0 && s.AgeGyr < 0.1
}
