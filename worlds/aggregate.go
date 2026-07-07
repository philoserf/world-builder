package worlds

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// AggregateSystem computes the system-wide aggregations after every
// body's per-body pipeline is complete: BaselineN backfill per allocation, ShortProfile,
// LongProfile, and the auto-picked mainworld. Pure function — no
// rolls. Stage-10 entry point per docs/api-surface.md § Stage 10.
func AggregateSystem(u *Universe) {
	// Step 1 — backfill per-allocation BaselineN.
	allocs := make([]StarAllocation, len(u.Placement.Allocations))
	copy(allocs, u.Placement.Allocations)
	for i := range allocs {
		allocs[i].BaselineN = computeBaselineN(allocs[i].Group, u.Detail.Bodies)
	}
	u.Detail.Allocations = allocs

	// Step 2 — system-wide profile strings.
	u.Detail.ShortProfile = buildShortProfile(u)
	u.Detail.LongProfile = buildLongProfile(u)

	// Step 3 — mainworld pick.
	designation, body := pickMainworld(u)
	u.Detail.MainworldDesignation = designation
	u.Detail.Mainworld = body
}

// computeBaselineN returns the per-star baseline number for a group
// per WBH p.58: the 1-based index of the slot closest to the group's
// own HZCO. This Stage-10 profile concept is unrelated to Stage 1's
// Baseline Number/Baseline Orbit placement machinery (baseline.go).
// Exceptions:
//   - if the group has no Members or no non-empty bodies: 0
//   - if no body falls in HZ (HZCO ± 1.0): 0
//   - if every body is in HZ: total slot count
func computeBaselineN(g Group, bodies []Body) int {
	if len(g.Members) == 0 {
		return 0
	}
	hzco := g.HZCO()

	idx := 0
	inHZCount := 0
	totalCount := 0
	bestDelta := math.Inf(1)
	bestIdx := 0
	for _, body := range bodies {
		if body.Group.Designation != g.Designation || body.Kind == BodyEmpty {
			continue
		}
		idx++
		totalCount++
		if body.Orbit >= hzco-1.0 && body.Orbit <= hzco+1.0 {
			inHZCount++
		}
		d := math.Abs(body.Orbit - hzco)
		if d < bestDelta {
			bestDelta = d
			bestIdx = idx
		}
	}
	if totalCount == 0 || inHZCount == 0 {
		return 0
	}
	if inHZCount == totalCount {
		return totalCount
	}
	return bestIdx
}

// buildShortProfile renders the WBH p.58 short Planetary System
// Profile: "G-P-T-N-S" where G=gas giants, P=belts, T=terrestrials,
// N=baseline number (floored at 0), S=system spread (one decimal).
func buildShortProfile(u *Universe) string {
	n := max(0, u.Placement.BaselineN)
	return fmt.Sprintf(
		"%d-%d-%d-%d-%s",
		u.Placement.Counts.GasGiants,
		u.Placement.Counts.PlanetoidBelts,
		u.Placement.Counts.Terrestrials,
		n,
		formatSpread(u.Placement.SystemSpread),
	)
}

// buildLongProfile renders the WBH p.58 long Planetary System
// Profile: "St-N-W-W-W...-S:..." per star. Slot codes (in orbit
// order): G (gas giant), P (belt), T (terrestrial).
func buildLongProfile(u *Universe) string {
	type starSegment struct {
		designation string
		baselineN   int
		codes       []string
	}
	segments := []starSegment{}
	groupIdx := map[string]int{}

	for _, alloc := range u.Detail.Allocations {
		groupIdx[alloc.Group.Designation] = len(segments)
		segments = append(segments, starSegment{
			designation: alloc.Group.Designation,
			baselineN:   alloc.BaselineN,
		})
	}

	for body := range u.AllBodies() {
		idx, ok := groupIdx[body.Group.Designation]
		if !ok {
			continue
		}
		var code string
		switch body.Kind {
		case BodyTerrestrial:
			code = "T"
		case BodyGasGiant:
			code = "G"
		case BodyPlanetoidBelt:
			code = "P"
		default:
			continue
		}
		segments[idx].codes = append(segments[idx].codes, code)
	}

	parts := make([]string, 0, len(segments))
	for _, seg := range segments {
		fields := []string{seg.designation, strconv.Itoa(seg.baselineN)}
		fields = append(fields, seg.codes...)
		fields = append(fields, formatSpread(u.Placement.SystemSpread))
		parts = append(parts, strings.Join(fields, "-"))
	}
	return strings.Join(parts, ":")
}

func formatSpread(s float64) string {
	return strconv.FormatFloat(s, 'f', 1, 64)
}

// pickMainworld implements the WBH p.134 priority chain:
//
//  1. Body with native (extant or extinct) sophont, ranked by
//     Habitability then Resource.
//  2. Highest Habitability > 0.
//  3. Highest ResourceRating > 0 (admits belts).
//  4. First terrestrial / moon / belt body in iteration order.
//
// Walks every Body in the universe via AllBodies (planets, moons,
// belts). Returns ("", nil) only when the system has no terrestrial /
// moon / belt bodies whatsoever.
func pickMainworld(u *Universe) (string, *Body) {
	type candidate struct {
		body         *Body
		habitability int
		resource     int
		hasSophont   bool
	}
	var candidates []candidate

	for body := range u.AllBodies() {
		if body.Kind != BodyTerrestrial && body.Kind != BodyMoon && body.Kind != BodyPlanetoidBelt {
			continue
		}
		c := candidate{body: body}
		if body.Habitability != nil {
			c.habitability = body.Habitability.Rating
		}
		if body.Biology != nil {
			c.resource = body.Biology.ResourceRating
			c.hasSophont = body.Biology.HasNativeSophont || body.Biology.HadExtinctSophont
		}
		if body.Kind == BodyPlanetoidBelt && body.Belt != nil {
			c.resource = body.Belt.ResourceRating
		}
		candidates = append(candidates, c)
	}

	if len(candidates) == 0 {
		return "", nil
	}

	better := func(i, best int) bool {
		return candidates[i].habitability > candidates[best].habitability ||
			(candidates[i].habitability == candidates[best].habitability &&
				candidates[i].resource > candidates[best].resource)
	}

	best := -1
	for i, c := range candidates {
		if !c.hasSophont {
			continue
		}
		if best == -1 || better(i, best) {
			best = i
		}
	}
	if best == -1 {
		for i, c := range candidates {
			if c.habitability == 0 {
				continue
			}
			if best == -1 || better(i, best) {
				best = i
			}
		}
	}
	if best == -1 {
		for i, c := range candidates {
			if c.resource == 0 {
				continue
			}
			if best == -1 || candidates[i].resource > candidates[best].resource {
				best = i
			}
		}
	}
	if best == -1 {
		best = 0
	}
	return candidates[best].body.Designation, candidates[best].body
}
