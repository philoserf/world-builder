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
	"errors"
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

// ErrPostStellarPrimaryUnsupported indicates the primary star has no
// p.39 MAO row — Brown Dwarf, White Dwarf, Neutron Star, Black Hole,
// Pulsar, or Protostar. MAO for these kinds lives in the WBH Special
// Circumstances chapter and is not yet encoded. (The name predates
// protostar inclusion; pre-stellar kinds like nebulae and star
// clusters fail earlier at the age-formula step and never reach MAO.)
// Wraps stars.ErrSpecialCircumstances so callers can classify uniformly.
var ErrPostStellarPrimaryUnsupported = fmt.Errorf(
	"worlds: post-stellar primary MAO requires Special Circumstances chapter: %w",
	stars.ErrSpecialCircumstances,
)

// maoRow is one row of the WBH p. 39 Minimum Allowable Orbit# table,
// keyed by luminosity class.
type maoRow struct {
	Ia, Ib, II, III, IV, V, VI *float64
}

// maoTablePage39 is the WBH p. 39 MAO table, keyed by spectral-type
// short code ("O0", "B5", "G7", ...).
//
// nil pointer means the book leaves the cell as "—" (combination does
// not exist as a star).
var maoTablePage39 = map[string]maoRow{
	"O0": {Ia: new(0.63), Ib: new(0.60), II: new(0.55), III: new(0.53), V: new(0.5), VI: new(0.01)},
	"O5": {Ia: new(0.55), Ib: new(0.50), II: new(0.45), III: new(0.38), V: new(0.3), VI: new(0.01)},
	"B0": {Ia: new(0.50), Ib: new(0.35), II: new(0.30), III: new(0.25), IV: new(0.20), V: new(0.18), VI: new(0.01)},
	"B5": {Ia: new(1.67), Ib: new(0.63), II: new(0.35), III: new(0.15), IV: new(0.13), V: new(0.09), VI: new(0.01)},
	"A0": {Ia: new(3.34), Ib: new(1.40), II: new(0.75), III: new(0.13), IV: new(0.10), V: new(0.06)},
	"A5": {Ia: new(4.17), Ib: new(2.17), II: new(1.17), III: new(0.13), IV: new(0.07), V: new(0.05)},
	"F0": {Ia: new(4.42), Ib: new(2.50), II: new(1.33), III: new(0.13), IV: new(0.07), V: new(0.04)},
	"F5": {Ia: new(5.00), Ib: new(3.25), II: new(1.87), III: new(0.13), IV: new(0.06), V: new(0.03)},
	"G0": {Ia: new(5.21), Ib: new(3.59), II: new(2.24), III: new(0.25), IV: new(0.07), V: new(0.03), VI: new(0.02)},
	"G5": {Ia: new(5.34), Ib: new(3.84), II: new(2.67), III: new(0.38), IV: new(0.10), V: new(0.02), VI: new(0.02)},
	"K0": {Ia: new(5.59), Ib: new(4.17), II: new(3.17), III: new(0.50), IV: new(0.15), V: new(0.02), VI: new(0.02)},
	"K5": {Ia: new(6.17), Ib: new(4.84), II: new(4.00), III: new(1.00), V: new(0.02), VI: new(0.01)},
	"M0": {Ia: new(6.80), Ib: new(5.42), II: new(4.59), III: new(1.68), V: new(0.02), VI: new(0.01)},
	"M5": {Ia: new(7.20), Ib: new(6.17), II: new(5.30), III: new(3.00), V: new(0.01), VI: new(0.01)},
	"M9": {Ia: new(7.80), Ib: new(6.59), II: new(5.92), III: new(4.34), V: new(0.01), VI: new(0.01)},
}

// ErrNoMAOForStar reports a "—" cell in the p. 39 MAO table — the
// spectral type / class combination does not exist as a star.
var ErrNoMAOForStar = errors.New("worlds: spectral type / class combination has no MAO entry")

// isPostStellar reports whether a StarKind is post-stellar (BD, D, NS,
// BH, Pulsar). Used by the WBH p.37 DM stack in counts.go — semantic,
// not structural; do not broaden it. To gate the MAO lookup, use
// lacksP39MAORow instead, which also excludes protostars.
func isPostStellar(k stars.StarKind) bool {
	switch k {
	case stars.KindBrownDwarf, stars.KindWhiteDwarf,
		stars.KindNeutronStar, stars.KindBlackHole, stars.KindPulsar:
		return true
	}
	return false
}

// lacksP39MAORow reports whether a StarKind has no row in the WBH
// p.39 Minimum Allowable Orbit# table — every post-stellar kind plus
// the pre-stellar Protostar. MAO for these kinds lives in the
// Special Circumstances chapter (not yet encoded).
func lacksP39MAORow(k stars.StarKind) bool {
	return isPostStellar(k) || k == stars.KindProtostar
}

// maoCell reads the MAO cell for an exact spectral type key (e.g. "G5")
// at a given luminosity class. Returns ErrNoMAOForStar if the cell is
// the book's "—".
func maoCell(typeKey string, lc stars.LuminosityClass) (float64, error) {
	row, ok := maoTablePage39[typeKey]
	if !ok {
		return 0, fmt.Errorf("worlds: no MAO row for %q", typeKey)
	}
	var ptr *float64
	switch lc {
	case stars.Ia:
		ptr = row.Ia
	case stars.Ib:
		ptr = row.Ib
	case stars.II:
		ptr = row.II
	case stars.III:
		ptr = row.III
	case stars.IV:
		ptr = row.IV
	case stars.V:
		ptr = row.V
	case stars.VI:
		ptr = row.VI
	default:
		return 0, fmt.Errorf("worlds: unknown luminosity class %q", lc)
	}
	if ptr == nil {
		return 0, ErrNoMAOForStar
	}
	return *ptr, nil
}

// MAO returns the Minimum Allowable Orbit# for a star, interpolated by
// spectral type within its luminosity-class column per the WBH p. 39
// table.
//
// Kinds without a p.39 row (post-stellar, plus protostar) return
// ErrPostStellarPrimaryUnsupported. Combinations the book lists as
// "—" return ErrNoMAOForStar.
func MAO(s stars.Star) (float64, error) {
	if lacksP39MAORow(s.Kind) {
		return 0, ErrPostStellarPrimaryUnsupported
	}
	lower, upper, frac := bracketSpectralType(s.SpectralType)
	lo, errLo := maoCell(lower, s.LuminosityClass)
	hi, errHi := maoCell(upper, s.LuminosityClass)
	switch {
	case errors.Is(errLo, ErrNoMAOForStar) && errors.Is(errHi, ErrNoMAOForStar):
		return 0, ErrNoMAOForStar
	case errors.Is(errLo, ErrNoMAOForStar):
		return hi, nil
	case errors.Is(errHi, ErrNoMAOForStar):
		return lo, nil
	case errLo != nil:
		return 0, errLo
	case errHi != nil:
		return 0, errHi
	}
	return lo + (hi-lo)*frac, nil
}

// bracketSpectralType returns the two p. 39 grid keys bracketing st
// within st's letter, and the fractional position from lower to upper
// (0.0 at lower, 1.0 at upper).
//
// At exact grid points (O0, O5, ..., K5, M0, M5, M9) the function
// returns frac=0, so the upper key is unused in interpolation. Only
// M5 and M9 (the table's terminal rows) return lower == upper.
func bracketSpectralType(st stars.SpectralType) (lower, upper string, frac float64) {
	letter := string(st.Letter)
	switch {
	case st.Letter == 'M' && st.Subtype >= 5:
		// Bracket M5 → M9; subtypes 5..9.
		if st.Subtype <= 5 {
			return "M5", "M5", 0
		}
		if st.Subtype >= 9 {
			return "M9", "M9", 0
		}
		return "M5", "M9", float64(st.Subtype-5) / 4.0
	case st.Subtype < 5:
		lower = letter + "0"
		upper = letter + "5"
		return lower, upper, float64(st.Subtype) / 5.0
	default:
		// st.Subtype in [5, 9] but letter != M (handled above).
		// Bracket Letter5 → NextLetter0.
		next := nextSpectralLetter(st.Letter)
		lower = letter + "5"
		upper = string(next) + "0"
		return lower, upper, float64(st.Subtype-5) / 5.0
	}
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

// nextSpectralLetter returns the next cooler spectral letter in O B A F G K M order.
func nextSpectralLetter(l stars.SpectralLetter) stars.SpectralLetter {
	switch l {
	case 'O':
		return 'B'
	case 'B':
		return 'A'
	case 'A':
		return 'F'
	case 'F':
		return 'G'
	case 'G':
		return 'K'
	case 'K':
		return 'M'
	default:
		return l
	}
}

// AvailableOrbits applies the WBH pp. 38–40 simplified rules to a
// stars.System and returns per-group allowed Orbit# intervals.
//
// Implementation walks rules 1–11 in order, mutating each group's
// interval set. See spec for the rule list.
//
// Returns ErrPostStellarPrimaryUnsupported if the primary star is a
// Brown Dwarf, White Dwarf, Neutron Star, Black Hole, Pulsar, or
// Protostar (their MAO is in the Special Circumstances chapter, not
// yet encoded).
func AvailableOrbits(sys stars.System) (Result, error) {
	if lacksP39MAORow(sys.Primary.Kind) {
		return Result{}, ErrPostStellarPrimaryUnsupported
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
		if lacksP39MAORow(m.Kind) {
			groups[i].MAO = 0
			continue
		}
		mao, err := MAO(m)
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
		// Rule 1 already populated groups[i].MAO from MAO(Members[0]).
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
		secMAO, _ := MAO(c.Star)
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
