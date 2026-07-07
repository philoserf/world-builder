// Package worlds implements WBH world-placement procedures atop
// stars.System.
//
// Sub-project 2A covers Available Orbits and the Habitable Zone
// Centre Orbit# (HZCO). HZCO lives in the stars package as a stellar
// property; available-orbits computation lives here as a system-level
// constraint over those stars.
//
// Source: WBH pp. 38–43 (System Worlds and Orbits chapter).
package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/stars"
)

// Interval is a closed Orbit# range [Min, Max].
type Interval struct {
	Min, Max float64
}

// Group is one body or barycentric pair sharing an orbit set.
//
// Single-star group: Members has one Star, Designation is "A"/"B"/"C"/"D".
// Pair group:        Members has two Stars (parent first, companion
//
//	second), Designation is "Aab"/"Cab"/...
type Group struct {
	Designation string
	Members     []stars.Star
	MAO         float64    // p. 39 table; for pairs, raised by rule 2 if applicable
	Intervals   []Interval // disjoint, sorted ascending

	// companionEcc records the companion's eccentricity for pair
	// groups. Set by identifyGroups; read by AvailableOrbits's rule 2
	// pass. Unexported because it's an implementation detail of the
	// rule pipeline, not part of the public API.
	companionEcc float64

	// sourceCompanion is the CompanionStar that gave rise to this
	// secondary group (nil for the primary group). Set by identifyGroups
	// so rules 9–11 can look up the secondary's orbit class and
	// eccentricity directly instead of walking a parallel index. Also
	// used by sub-project 2B steps that need the secondary's
	// orbit-around-primary. The pointer is valid only for the lifetime of
	// the stars.System passed to AvailableOrbits; callers must not append
	// to System.Companions while a Result is in scope.
	sourceCompanion *stars.CompanionStar
}

// Total returns the sum of (Max - Min) over all intervals — the value
// the book calls "total Orbit#s" used in placement allocation
// (sub-project 2B).
func (g Group) Total() float64 {
	var t float64
	for _, iv := range g.Intervals {
		t += iv.Max - iv.Min
	}
	return t
}

// Contains reports whether orbit is inside any of g.Intervals.
// Endpoints count as inside.
func (g Group) Contains(orbit float64) bool {
	for _, iv := range g.Intervals {
		if orbit >= iv.Min && orbit <= iv.Max {
			return true
		}
	}
	return false
}

// Result is the per-group available orbits for an entire system.
type Result struct {
	Groups []Group // ordered by ascending stellar Orbit# of the group's outer member
}

// subtract removes the exclusion range (exMin, exMax) from a sorted
// disjoint interval list and returns the remaining intervals.
//
// Tolerates exclusions that fully cover or only partially overlap
// existing intervals. Endpoints are inclusive on the kept side
// (i.e., an interval [a, b] minus exclusion (exMin, exMax) yields
// [a, exMin] and [exMax, b] — the exclusion is open).
func subtract(intervals []Interval, exMin, exMax float64) []Interval {
	if exMin >= exMax {
		return intervals
	}
	out := make([]Interval, 0, len(intervals)+1)
	for _, iv := range intervals {
		// No overlap.
		if exMax <= iv.Min || exMin >= iv.Max {
			out = append(out, iv)
			continue
		}
		// Left remainder.
		if exMin > iv.Min {
			out = append(out, Interval{Min: iv.Min, Max: exMin})
		}
		// Right remainder.
		if exMax < iv.Max {
			out = append(out, Interval{Min: exMax, Max: iv.Max})
		}
	}
	return out
}

// hasOrbitClass reports whether sys has any non-companion CompanionStar
// in the given orbit class with ParentIndex == -1.
func hasOrbitClass(sys stars.System, oc stars.OrbitClass) bool {
	for _, c := range sys.Companions {
		if c.OrbitClass == oc && c.ParentIndex == -1 {
			return true
		}
	}
	return false
}

// adjacenciesFor returns the orbit classes adjacent to self per WBH p. 39:
//   - Close adjacent to Near
//   - Near adjacent to Close and Far
//   - Far adjacent to Near
//
// Note: Close+Far without Near does NOT count (book p. 39 explicit).
func adjacenciesFor(self stars.OrbitClass) []stars.OrbitClass {
	switch self {
	case stars.OrbitClose:
		return []stars.OrbitClass{stars.OrbitNear}
	case stars.OrbitNear:
		return []stars.OrbitClass{stars.OrbitClose, stars.OrbitFar}
	case stars.OrbitFar:
		return []stars.OrbitClass{stars.OrbitNear}
	}
	return nil
}

// hasAdjacentZone reports whether secondary in zone `self` has a
// populated adjacent zone, per rule 9.
func hasAdjacentZone(sys stars.System, self stars.OrbitClass) bool {
	for _, oc := range adjacenciesFor(self) {
		if hasOrbitClass(sys, oc) {
			return true
		}
	}
	return false
}

// adjacentEccGT02 reports whether any star in an adjacent zone has
// eccentricity > 0.2.
func adjacentEccGT02(sys stars.System, self stars.OrbitClass) bool {
	wanted := adjacenciesFor(self)
	for _, c := range sys.Companions {
		if c.ParentIndex != -1 {
			continue
		}
		for _, oc := range wanted {
			if c.OrbitClass == oc && c.Eccentricity > 0.2 {
				return true
			}
		}
	}
	return false
}

// AvailableOrbits applies the WBH pp. 38–40 simplified rules to a
// stars.System and returns per-group allowed Orbit# intervals.
//
// Implementation walks rules 1–11 in order, mutating each group's
// interval set. See spec for the rule list.
//
// Returns stars.ErrPostStellarPrimaryUnsupported if the primary star is a
// Brown Dwarf, White Dwarf, Neutron Star, Black Hole, Pulsar, or
// Protostar (their MAO is in the Special Circumstances chapter, not
// yet encoded).
func AvailableOrbits(sys stars.System) (Result, error) {
	if stars.LacksP39MAORow(sys.Primary.Kind) {
		return Result{}, stars.ErrPostStellarPrimaryUnsupported
	}

	groups := identifyGroups(sys)

	// Rule 1: MAO from p. 39 table for each group.
	for i := range groups {
		// Pair groups use the parent (first member) MAO; rule 2 may
		// raise it later. Post-stellar group representatives (BD/D/NS/
		// BH/Pulsar/Protostar) have no p.39 row — per WBH that's
		// Special Circumstances territory. Referee call: such bodies
		// exist in the system but contribute zero MAO (they don't push
		// out the parent's orbital exclusion zone). See
		// docs/history/plan-clean-every-run.md Phase 2f.
		m := groups[i].Members[0]
		if stars.LacksP39MAORow(m.Kind) {
			groups[i].MAO = 0
			continue
		}
		mao, err := stars.MAO(m)
		if err != nil {
			return Result{}, fmt.Errorf("worlds: MAO for group %s: %w",
				groups[i].Designation, err)
		}
		groups[i].MAO = mao
	}

	// Rule 2: companion eccentricity raises pair lower bound.
	// For pair groups, the lower bound becomes 0.50 + companion_ecc.
	// If the larger star's MAO > 0.2, add it to the unavailable lower zone.
	for i := range groups {
		if len(groups[i].Members) < 2 {
			continue
		}
		// Rule 1 already populated groups[i].MAO from stars.MAO(Members[0]).
		// Reuse that value as the larger-star MAO (parent is always the
		// first member per identifyGroups; WBH treats the parent as the
		// more massive/luminous star).
		largerMAO := groups[i].MAO
		floor := 0.50 + groups[i].companionEcc
		if largerMAO > 0.2 {
			floor += largerMAO
		}
		if floor > groups[i].MAO {
			groups[i].MAO = floor
		}
	}

	// Rule 3: primary group can have Orbit#s up to 20.
	groups[0].Intervals = []Interval{{Min: groups[0].MAO, Max: 20.0}}

	// Rules 5+6: each Close/Near/Far secondary excludes a range centred
	// on its Orbit# from the primary's intervals.
	//
	// Rule 5 (WBH p. 38): base exclusion width ±1.
	// Rule 6 (WBH p. 38): if companion eccentricity > 0.2, widen by ±1.
	// Secondary MAO > 0.2: additionally subtract the secondary's MAO on
	// each side.
	//
	// Rule 4 (companions occupy same orbit as parent) is handled by
	// identifyGroups via group folding; from here on, OrbitCompanion
	// entries are ignored.
	for _, c := range sys.Companions {
		if c.OrbitClass == stars.OrbitCompanion {
			continue
		}
		s := c.OrbitNumber
		// Rules 5+6+7: base ±1, widened by ±1 more if ecc > 0.2,
		// widened by another ±1 if ecc > 0.5 AND Close or Near (not Far).
		width := 1.0
		if c.Eccentricity > 0.2 {
			width += 1.0 // rule 6
		}
		if c.Eccentricity > 0.5 && (c.OrbitClass == stars.OrbitClose || c.OrbitClass == stars.OrbitNear) {
			width += 1.0 // rule 7 (Close/Near only, not Far)
		}
		exLow := s - width
		exHigh := s + width
		secMAO, _ := stars.MAO(c.Star)
		if secMAO > 0.2 {
			exLow -= secMAO
			exHigh += secMAO
		}
		groups[0].Intervals = subtract(groups[0].Intervals, exLow, exHigh)
	}

	// Rule 8: each secondary (Close/Near/Far) has its own orbit range.
	// Lower bound is the secondary's MAO (Rule 1); upper bound is
	// (Orbit# − 3). Rules 9–11 reduce maxOffset further:
	//   - Rule 9: -1 if adjacent zone populated (max once per secondary).
	//   - Rule 10: -1 if self ecc > 0.2 OR any adjacent zone star has ecc > 0.2 (max once).
	//   - Rule 11: -1 if self ecc > 0.5 (max once).
	for i := range groups {
		if groups[i].sourceCompanion == nil {
			continue // primary group
		}
		sc := groups[i].sourceCompanion
		maxOffset := sc.OrbitNumber - 3 // rule 8
		if hasAdjacentZone(sys, sc.OrbitClass) {
			maxOffset-- // rule 9
		}
		if sc.Eccentricity > 0.2 || adjacentEccGT02(sys, sc.OrbitClass) {
			maxOffset-- // rule 10
		}
		if sc.Eccentricity > 0.5 {
			maxOffset-- // rule 11
		}
		if maxOffset < 0 {
			maxOffset = 0
		}
		if maxOffset < groups[i].MAO {
			groups[i].Intervals = nil
		} else {
			groups[i].Intervals = []Interval{{Min: groups[i].MAO, Max: maxOffset}}
		}
	}

	return Result{Groups: groups}, nil
}

// identifyGroups partitions a System into its barycentric orbit groups.
// See package doc comment for the rules.
//
// Pairing uses CompanionStar.ParentIndex: -1 means "child of the
// primary"; otherwise it is an index into sys.Companions.
func identifyGroups(sys stars.System) []Group {
	groups := []Group{}

	// findCompanionOf returns the Star and its eccentricity for a
	// Companion-class entry whose ParentIndex matches parentIdx, or
	// (Star{}, 0, false) if none.
	findCompanionOf := func(parentIdx int) (stars.Star, float64, bool) {
		for _, c := range sys.Companions {
			if c.ParentIndex == parentIdx && c.OrbitClass == stars.OrbitCompanion {
				return c.Star, c.Eccentricity, true
			}
		}
		return stars.Star{}, 0, false
	}

	// Primary group: primary plus its companion (parent index -1, class
	// Companion), if any.
	primaryGroup := Group{Members: []stars.Star{sys.Primary}}
	if companion, ecc, ok := findCompanionOf(-1); ok {
		primaryGroup.Members = append(primaryGroup.Members, companion)
		primaryGroup.companionEcc = ecc
		primaryGroup.Designation = "Aab"
	} else {
		primaryGroup.Designation = "A"
	}
	groups = append(groups, primaryGroup)

	// Secondary groups: each Close/Near/Far companion of the primary
	// becomes its own group (with its own companion folded in if any).
	// Walk Close, then Near, then Far in canonical order so designations
	// are assigned positionally (B, C, D) skipping absent slots.
	letters := []string{"A", "B", "C", "D"}
	letterIdx := 1
	for _, oc := range []stars.OrbitClass{stars.OrbitClose, stars.OrbitNear, stars.OrbitFar} {
		for i := range sys.Companions {
			c := sys.Companions[i]
			if c.OrbitClass != oc || c.ParentIndex != -1 {
				continue
			}
			if letterIdx >= len(letters) {
				// WBH structurally caps secondaries at three (Close/Near/Far);
				// any additional secondaries are silently dropped rather than
				// panicking on letter assignment. Reaching this branch indicates
				// a hand-constructed System outside the WBH generator's output.
				break
			}
			group := Group{
				Members:         []stars.Star{c.Star},
				sourceCompanion: &sys.Companions[i],
			}
			if companion, ecc, ok := findCompanionOf(i); ok {
				group.Members = append(group.Members, companion)
				group.companionEcc = ecc
				group.Designation = letters[letterIdx] + "ab"
			} else {
				group.Designation = letters[letterIdx]
			}
			letterIdx++
			groups = append(groups, group)
		}
	}

	return groups
}
