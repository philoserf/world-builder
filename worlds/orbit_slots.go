package worlds

import (
	"fmt"

	"wbh/roller"
	"wbh/stars"
)

// Slot is one placed orbit position in a system, before world-body
// assignment (Step 8).
type Slot struct {
	StarSlot string // "A1", "A2", ..., "B+", "C5". The "+" suffix marks
	// the extra slot added to a star by Step 4's empty-bump.
	// Step 7 anomalous slots get their own "+" suffix when
	// placed; the AnomalousSlot.Anomaly field disambiguates.
	Group Group
	Orbit float64
}

// PlaceOrbitSlots implements WBH Step 6 (pp. 49-50).
//
// Walks each StarAllocation in order, placing slots outward from MAO by
// (spread + (2D-7) * spread/10). The baseline-N-th slot of the primary
// group is fixed at baselineOrbit (overrides variance). When the next
// slot would land in an exclusion zone, widens the spread by the zone
// width (mandatory, p. 49 narrative on Zed orbits 9-11).
//
// emptyOrbits (from Step 4) are distributed before placement: Close star
// first, then Near, then Far, then the primary; each receiving star gets
// one extra slot named with a "+" suffix.
//
// baselineN is the explicit slot index (1-based) produced by Step 2's
// RollBaselineNumber. It is clamped to [1, primary slot count].
func PlaceOrbitSlots(
	r roller.Roller,
	allocs []StarAllocation,
	baselineN int,
	baselineOrbit, spread float64,
	emptyOrbits int,
) ([]Slot, error) {
	if len(allocs) == 0 {
		return nil, nil
	}

	// Distribute empty orbits: Close → Near → Far → primary.
	extraSlots := make([]int, len(allocs))
	remaining := emptyOrbits
	for _, oc := range []stars.OrbitClass{stars.OrbitClose, stars.OrbitNear, stars.OrbitFar} {
		if remaining == 0 {
			break
		}
		for i := 1; i < len(allocs); i++ {
			if remaining == 0 {
				break
			}
			sc := allocs[i].Group.sourceCompanion
			if sc != nil && sc.OrbitClass == oc {
				extraSlots[i]++
				remaining--
			}
		}
	}
	// Any leftover goes to the primary.
	extraSlots[0] += remaining

	// Clamp baselineN to the valid slot range for the primary group.
	primarySlotCount := allocs[0].AllocatedWorlds + extraSlots[0]
	if baselineN < 1 {
		baselineN = 1
	}
	if primarySlotCount > 0 && baselineN > primarySlotCount {
		baselineN = primarySlotCount
	}

	var out []Slot
	for i, alloc := range allocs {
		slotCount := alloc.AllocatedWorlds + extraSlots[i]
		if slotCount == 0 {
			continue
		}
		cur := alloc.Group.MAO
		for j := 0; j < slotCount; j++ {
			label := slotLabel(alloc.Group.Designation, j, alloc.AllocatedWorlds, extraSlots[i])
			isBaselineSlot := i == 0 && j == baselineN-1

			var orbit float64
			switch {
			case isBaselineSlot:
				orbit = baselineOrbit
			case j == 0:
				v := r.Roll("2D")
				orbit = cur + spread + float64(v-7)*spread/10.0
			default:
				v := r.Roll("2D")
				proposed := cur + spread + float64(v-7)*spread/10.0
				// Exclusion-zone widening applies only to the primary group.
				if i == 0 {
					if zone := excludedZoneAt(alloc.Group, proposed); zone != 0 {
						proposed += zone
					}
				}
				orbit = proposed
			}
			out = append(out, Slot{
				StarSlot: label,
				Group:    alloc.Group,
				Orbit:    orbit,
			})
			cur = orbit
		}
	}
	return out, nil
}

// slotLabel returns the StarSlot id for a regular or bumped slot.
//
// regularCount is the AllocatedWorlds for this group (regular slots
// numbered 1..regularCount). extraCount is the count of bumped slots
// from Step 4 (each named with "+" suffix). The bumped slot is placed
// last in the sequence.
func slotLabel(designation string, indexInGroup, regularCount, _ int) string {
	prefix := "?"
	if len(designation) > 0 {
		prefix = string(designation[0])
	}
	if indexInGroup < regularCount {
		return fmt.Sprintf("%s%d", prefix, indexInGroup+1)
	}
	return prefix + "+"
}

// excludedZoneAt returns the width of the exclusion zone at orbit, if
// orbit lands in a gap between intervals; otherwise 0.
func excludedZoneAt(g Group, orbit float64) float64 {
	if g.Contains(orbit) {
		return 0
	}
	for i := 0; i < len(g.Intervals)-1; i++ {
		gapStart := g.Intervals[i].Max
		gapEnd := g.Intervals[i+1].Min
		if orbit >= gapStart && orbit <= gapEnd {
			return gapEnd - gapStart
		}
	}
	return 0
}
