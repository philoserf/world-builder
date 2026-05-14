package worlds

import (
	"math"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

// CheckRunawayGreenhouse evaluates and applies WBH p.79 Optional Runaway
// Greenhouse. Triggered when:
//   - body.Atmosphere is non-nil AND atm.Code is 2 or above
//   - body.Temperature is non-nil AND MeanK > 303K
//   - 2D + DMs ≥ 12
//
// DMs:
//   - +1 per System Age Gyr (round up)
//   - +4 if mean T ≥ 388K (boiling temperature, 12+ on basic temp table)
//   - +1 if originally tainted (codes 2, 4, 7, 9)
//   - -2 if Size 2-5
//
// On trigger, the outcome depends on the original atm code:
//
//   - atm 2-9, D (13), E (14): mutate body.Atmosphere.Code via 1D table:
//     1   → A (10)
//     2-4 → B (11)
//     5+  → C (12)
//
//   - atm A (10), B (11), C (12), F+ (15+): no mutation. WBH: "the only
//     effect of a runaway greenhouse is to consider the world to be
//     boiling." The caller treats the bool return as "consider boiling"
//     and applies hydro DM-6 instead of DM-2.
//
// Returns true iff the trigger fired (regardless of outcome path).
// Caller distinguishes the mutation vs boiling-only paths by comparing
// the pre-call atm.Code to the post-call value.
func CheckRunawayGreenhouse(r roller.Roller, body *Body, sys stars.System) bool {
	if body.Atmosphere == nil || body.Temperature == nil {
		return false
	}
	if body.Temperature.MeanK <= 303 {
		return false
	}
	code := body.Atmosphere.Code
	// Atm 0 (None) and 1 (Trace) are not in the WBH p.79 runaway table.
	if code < 2 {
		return false
	}

	// Trigger roll: 2D + DMs.
	dm := 0
	dm += int(math.Ceil(sys.Primary.AgeGyr))
	if body.Temperature.MeanK >= 388 {
		dm += 4
	}
	if code == 2 || code == 4 || code == 7 || code == 9 {
		dm++
	}
	si := SizeAsInt(body.SizeCode)
	if si >= 2 && si <= 5 {
		dm -= 2
	}

	roll := r.Roll("2D")
	if roll+dm < 12 {
		return false
	}

	// Trigger fired. WBH p.79: for atm A, B, C, or F+, the only effect
	// is the "consider boiling" hydro DM (handled by the caller). No
	// atm code mutation, no subtype/pressure re-roll.
	if code == 10 || code == 11 || code == 12 || code >= 15 {
		return true
	}

	// Atm 2-9, D, E: mutate code via 1D table.
	atmRoll := r.Roll("1D")
	switch {
	case atmRoll == 1:
		body.Atmosphere.Code = 10 // A
	case atmRoll <= 4:
		body.Atmosphere.Code = 11 // B
	default:
		body.Atmosphere.Code = 12 // C
	}
	return true
}
