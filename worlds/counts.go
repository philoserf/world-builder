package worlds

import (
	"wbh/roller"
	"wbh/stars"
)

// Counts is the per-system count of bodies a Referee will place.
type Counts struct {
	GasGiants      int // 0–6
	PlanetoidBelts int // 0–3
	Terrestrials   int // 0–13 (cap from Step 7 narrative)
	Total          int // GasGiants + PlanetoidBelts + Terrestrials
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

	// Belts and terrestrials are filled in by later tasks. For now,
	// delegate to placeholders so the test can reach the gas-giant
	// assertions. They will be implemented in tasks 4 and 5.
	c.PlanetoidBelts = 0
	c.Terrestrials = 0
	c.Total = c.GasGiants + c.PlanetoidBelts + c.Terrestrials
	return c, nil
}

// gasGiantExistenceDMs computes the WBH p. 37 DM stack for the
// gas-giant existence roll only. The single-Class-V DM+1 does not
// apply to existence (it applies to quantity only).
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
