package worlds

import (
	"fmt"

	"wbh/roller"
)

// BodyType classifies the body type assigned to an orbit slot in Step 8.
type BodyType int

const (
	// BodyEmpty marks an orbit slot with no world assigned.
	BodyEmpty BodyType = iota
	// BodyTerrestrial marks a terrestrial planet slot.
	BodyTerrestrial
	// BodyGasGiant marks a gas giant slot.
	BodyGasGiant
	// BodyPlanetoidBelt marks a planetoid belt slot.
	BodyPlanetoidBelt
)

// Placement is one fully-resolved orbit slot after Step 8.
type Placement struct {
	AnomalousSlot
	Body       BodyType
	PrefixRoll string // "1:6", "2:3" — audit trail
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
	out := make([]Placement, len(slots))
	for i, s := range slots {
		out[i] = Placement{AnomalousSlot: s}
	}
	assigned := make([]bool, len(slots))
	n := len(slots)

	prefixDie, prefixMax := prefixSpec(n)

	placeOne := func(body BodyType) error {
		prefix := r.Roll(prefixDie)
		right := r.Roll("1D")
		idx := slotIndex(prefix, right, n)
		// Collision-handling: +1 to right die; if right > 6, advance prefix
		// and wrap; if every prefix is exhausted, fall back to first
		// unassigned slot.
		attempts := 0
		for assigned[idx] {
			attempts++
			if attempts > n*6 {
				// Defensive fallback: scan for first unassigned slot.
				idx = -1
				for j, a := range assigned {
					if !a {
						idx = j
						break
					}
				}
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
			idx = slotIndex(prefix, right, n)
		}
		if idx >= n || idx < 0 {
			return fmt.Errorf("worlds: rolled slot index %d out of range (n=%d)", idx, n)
		}
		out[idx].Body = body
		out[idx].PrefixRoll = fmt.Sprintf("%d:%d", prefix, right)
		assigned[idx] = true
		return nil
	}

	// Order: empty → GG → belts → terrestrials.
	emptyCount := n - counts.Total
	if emptyCount < 0 {
		emptyCount = 0
	}
	for i := 0; i < emptyCount; i++ {
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

// slotIndex maps a (prefix, right) pair to a flat slot index, clamped to
// [0, n-1].
func slotIndex(prefix, right, n int) int {
	idx := (prefix-1)*6 + (right - 1)
	if idx >= n {
		idx = n - 1
	}
	if idx < 0 {
		idx = 0
	}
	return idx
}
