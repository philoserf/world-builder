package worlds

import (
	"math"

	"wbh/roller"
	"wbh/stars"
)

// CheckRunawayGreenhouse evaluates and applies WBH p.79 Optional Runaway
// Greenhouse. Triggered when:
//   - body.Atmosphere is non-nil AND atm.Code in {2-9, D=13, E=14}
//   - body.Temperature is non-nil AND MeanK > 303K
//   - 2D + DMs ≥ 12
//
// DMs:
//   - +1 per System Age Gyr (round up)
//   - +4 if mean T ≥ 388K (boiling temperature, 12+ on basic temp table)
//   - +1 if originally tainted (codes 2, 4, 7, 9)
//   - -2 if Size 2-5
//
// On trigger, mutates atm.Code via 1D table:
//
//	1   → A (10)
//	2-4 → B (11)
//	5+  → C (12)
//
// Returns true iff trigger fired. Caller re-rolls Hydrographics with DM-6
// (boiling) instead of DM-2 (hot) when this returns true.
//
// MVP simplification: atm A/B/C (10/11/12) and F+ (15+) skip this check.
// The book's "consider boiling" case for those codes (only flips hydrographics
// DM without mutating atm code) is deferred — see spec carry-forwards.
func CheckRunawayGreenhouse(r roller.Roller, body *DetailedPlacement, sys stars.System) bool {
	if body.Atmosphere == nil || body.Temperature == nil {
		return false
	}
	if body.Temperature.MeanK <= 303 {
		return false
	}
	code := body.Atmosphere.Code
	// Trigger range: atm 2-9, D (13), E (14). Skip A/B/C (10-12), F+ (15+), and 0/1.
	if code < 2 || code == 10 || code == 11 || code == 12 || code >= 15 {
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

	// Trigger fired: roll 1D for new atm code.
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
