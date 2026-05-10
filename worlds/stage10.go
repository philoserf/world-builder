package worlds

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// AggregateSystem computes the system-wide aggregations after every
// body has converged: BaselineN backfill per allocation, ShortProfile,
// LongProfile, and the auto-picked mainworld designation. Pure
// function — no rolls. Stage-10 entry point per docs/pass-2/api-
// surface.md § Stage 10.
//
// IISS form structs (Class0I / Class23 / Class4P) are populated in
// cycle 11 alongside the iiss/ renderers; cycle-10 leaves them at
// zero values.
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
	mainworld := pickMainworld(u.Detail.Bodies)
	// SystemDetail embeds iiss.SystemForms — promoted access:
	u.Detail.MainworldDesignation = mainworld
}

// computeBaselineN returns the per-star baseline number for a group
// per WBH p.58: the 1-based index of the slot closest to the group's
// own HZCO, except:
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

	for _, body := range u.Detail.Bodies {
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
//  4. First terrestrial-or-belt body in iteration order.
//
// Walks both Body and Children (moons). Returns "" only when the
// system has no terrestrial / belt bodies whatsoever.
func pickMainworld(bodies []Body) string {
	type candidate struct {
		designation  string
		habitability int
		resource     int
		hasSophont   bool
	}
	var candidates []candidate
	collect := func(designation string, kind BodyKind, h *Habitability, b *Biology, belt *BeltDetails) {
		if kind != BodyTerrestrial && kind != BodyMoon && kind != BodyPlanetoidBelt {
			return
		}
		c := candidate{designation: designation}
		if h != nil {
			c.habitability = h.Rating
		}
		if b != nil {
			c.resource = b.ResourceRating
			c.hasSophont = b.HasNativeSophont || b.HadExtinctSophont
		}
		if kind == BodyPlanetoidBelt && belt != nil {
			c.resource = belt.ResourceRating
		}
		candidates = append(candidates, c)
	}

	for i := range bodies {
		body := &bodies[i]
		collect(body.Designation, body.Kind, body.Habitability, body.Biology, body.Belt)
		for _, child := range body.Children {
			collect(child.Designation, child.Kind, child.Habitability, child.Biology, nil)
		}
	}

	if len(candidates) == 0 {
		return ""
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
	if best != -1 {
		return candidates[best].designation
	}

	for i, c := range candidates {
		if c.habitability == 0 {
			continue
		}
		if best == -1 || better(i, best) {
			best = i
		}
	}
	if best != -1 {
		return candidates[best].designation
	}

	for i, c := range candidates {
		if c.resource == 0 {
			continue
		}
		if best == -1 || candidates[i].resource > candidates[best].resource {
			best = i
		}
	}
	if best != -1 {
		return candidates[best].designation
	}

	return candidates[0].designation
}
