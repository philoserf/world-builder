package worlds

import "math"

// StarAllocation is one star group's slice of the per-system world budget.
type StarAllocation struct {
	Group           Group
	TotalStarOrbits int // floor(group.Total() + (1 if no companion and prior allowable > 0))
	AllocatedWorlds int
	BaselineN       int // per-star baseline number (1-based slot index of system baseline orbit within this group; 0 if out of range). Backfilled by computeBaselineN via AggregateSystem (stage10.go).
}

// AllocateOrbitsByStar implements WBH pp. 43–44 Step 1.
//
// For each Group in avail.Groups:
//   - Total Star Orbits = floor(group.Total()), plus 1 if the group has no
//     companion AND group.Total() > 0.
//
// The system Total = sum of TotalStarOrbits across groups.
//
// World allocation per group:
//   - First (primary) group: ceil(counts.Total × TotalStarOrbits / sysTotal)
//   - Middle groups:         floor(...)
//   - Last group:            counts.Total - sum(others)
//
// Single-star systems return a single allocation with all worlds.
func AllocateOrbitsByStar(avail Result, counts Counts) ([]StarAllocation, error) {
	if len(avail.Groups) == 0 {
		return nil, nil
	}

	out := make([]StarAllocation, len(avail.Groups))
	sysTotal := 0

	for i, g := range avail.Groups {
		out[i].Group = g
		priorAllowable := g.Total()

		tso := int(math.Floor(priorAllowable))
		if len(g.Members) == 1 && priorAllowable > 0 {
			tso++
		}

		out[i].TotalStarOrbits = tso
		sysTotal += tso
	}

	// Single-star fast path.
	if len(avail.Groups) == 1 {
		out[0].AllocatedWorlds = counts.Total

		return out, nil
	}

	// Multi-star: distribute counts.Total proportionally. sysTotal cannot
	// be zero in practice — a multi-star System produced by AvailableOrbits
	// always has at least one group (the primary) with non-zero allowable
	// orbits. The guard below catches the contrived edge case so the
	// proportional formula doesn't divide by zero.
	if sysTotal == 0 {
		out[len(out)-1].AllocatedWorlds = counts.Total

		return out, nil
	}

	assigned := 0

	for i := range out {
		switch {
		case i == 0:
			// Primary: round up.
			v := float64(counts.Total*out[i].TotalStarOrbits) / float64(sysTotal)
			out[i].AllocatedWorlds = int(math.Ceil(v))
		case i == len(out)-1:
			// Last: remainder.
			out[i].AllocatedWorlds = counts.Total - assigned
		default:
			// Middle: round down.
			v := float64(counts.Total*out[i].TotalStarOrbits) / float64(sysTotal)
			out[i].AllocatedWorlds = int(math.Floor(v))
		}

		if i < len(out)-1 {
			assigned += out[i].AllocatedWorlds
		}
	}

	return out, nil
}
