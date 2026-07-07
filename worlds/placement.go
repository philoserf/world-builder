package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
)

// Placement is one fully-resolved orbit slot after Step 8.
//
// BodyKind (planet, belt, gas giant, empty) is defined in body.go and
// shared with the pass-2 Body type so a single set of constants applies
// across Stage 1 (placement) and Stage 2+ (per-body procedures).
type Placement struct {
	AnomalousSlot
	Body         BodyKind
	PrefixRoll   string  // "1:6", "2:3" — audit trail
	Eccentricity float64 // populated by Step 9 (RollPlanetEccentricities)
}

// PlaceWorlds implements WBH Step 8 (pp. 51-52). Order: empty → gas
// giants → planetoid belts → terrestrials. Uses 1D:1D rolling with a
// prefix die selected by total slot count (≤6 → 1D, 7-12 → D2, 13-18 →
// D3, >18 → 1D with reroll-above-N).
//
// Collision handling: if rolled slot already has a body, +1 to the right
// die (within the same prefix), then advance to next slot id, then wrap
// to first unassigned slot.
//
// Mainworld and Continuation-only branches (moon of GG, size-1 in belt,
// atmosphere-DM raw-temp reverse-engineering) are out of scope.
func PlaceWorlds(r roller.Roller, slots []AnomalousSlot, counts Counts) ([]Placement, error) {
	// Input contract: enough slots for the requested worlds. Without
	// this, rollSlot's rejection-sampling loop can never produce a
	// valid index for an empty slots slice and spins forever.
	if counts.Total > 0 && len(slots) < counts.Total {
		return nil, fmt.Errorf("worlds: PlaceWorlds: %d slots cannot hold %d worlds", len(slots), counts.Total)
	}
	out := make([]Placement, len(slots))
	for i, s := range slots {
		out[i] = Placement{AnomalousSlot: s}
	}
	assigned := make([]bool, len(slots))
	n := len(slots)

	prefixDie, prefixMax := prefixSpec(n)

	// rollSlot does rejection sampling on both prefix and right per WBH
	// p. 51 ("rolls above the top prefix number ... are rerolled"):
	// keep rerolling prefix until ≤ prefixMax; keep rerolling right until
	// the resulting (prefix, right) pair maps to a valid in-range slot.
	rollSlot := func() (int, int) {
		for {
			p := r.Roll(prefixDie)
			if p > prefixMax {
				continue
			}
			right := r.Roll("1D")
			idx := (p-1)*6 + (right - 1)
			if idx >= 0 && idx < n {
				return p, right
			}
		}
	}

	placeOne := func(body BodyKind) error {
		prefix, right := rollSlot()
		idx := (prefix-1)*6 + (right - 1)
		// Collision: +1 to right; if right > 6 advance prefix and wrap to 1
		// (skipping prefixes whose minimum-right slot is already past n).
		// Bail to a fallback scan if every (prefix, right) pair is taken.
		attempts := 0
		for assigned[idx] {
			attempts++
			if attempts > prefixMax*6 {
				idx = firstUnassigned(assigned)
				if idx == -1 {
					return fmt.Errorf("worlds: no unassigned slots")
				}
				break
			}
			right++
			if right > 6 {
				right = 1
				prefix++
				if prefix > prefixMax {
					prefix = 1
				}
			}
			candidate := (prefix-1)*6 + (right - 1)
			if candidate >= n {
				continue // skip out-of-range advance steps
			}
			idx = candidate
		}
		out[idx].Body = body
		out[idx].PrefixRoll = fmt.Sprintf("%d:%d", prefix, right)
		assigned[idx] = true
		return nil
	}

	// Order: empty → GG → belts → terrestrials.
	emptyCount := max(n-counts.Total, 0)
	for range emptyCount {
		if err := placeOne(BodyEmpty); err != nil {
			return nil, err
		}
	}
	for i := 0; i < counts.GasGiants; i++ {
		if err := placeOne(BodyGasGiant); err != nil {
			return nil, err
		}
	}
	for i := 0; i < counts.PlanetoidBelts; i++ {
		if err := placeOne(BodyPlanetoidBelt); err != nil {
			return nil, err
		}
	}
	for i := 0; i < counts.Terrestrials; i++ {
		if err := placeOne(BodyTerrestrial); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// prefixSpec returns the prefix-die notation and max valid prefix value
// for n total slots, per WBH Step 8.
func prefixSpec(n int) (notation string, maxValid int) {
	switch {
	case n <= 6:
		return "1D", 1
	case n <= 12:
		return "D2", 2
	case n <= 18:
		return "D3", 3
	default:
		return "1D", (n + 5) / 6
	}
}

// firstUnassigned returns the index of the first false entry in assigned,
// or -1 if all are true.
func firstUnassigned(assigned []bool) int {
	for j, a := range assigned {
		if !a {
			return j
		}
	}
	return -1
}
