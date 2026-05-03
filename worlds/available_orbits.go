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

	"wbh/stars"
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

// ErrPostStellarPrimaryUnsupported indicates the primary star is a
// Brown Dwarf, White Dwarf, Neutron Star, Black Hole, or Pulsar —
// classes whose MAO is in the Special Circumstances chapter and not
// yet encoded.
var ErrPostStellarPrimaryUnsupported = errors.New(
	"worlds: post-stellar primary MAO requires Special Circumstances chapter",
)

// fp returns a pointer to the float64 — used for nil-able cells in tables.
func fp(x float64) *float64 { return &x }

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
	"O0": {Ia: fp(0.63), Ib: fp(0.60), II: fp(0.55), III: fp(0.53), V: fp(0.5), VI: fp(0.01)},
	"O5": {Ia: fp(0.55), Ib: fp(0.50), II: fp(0.45), III: fp(0.38), V: fp(0.3), VI: fp(0.01)},
	"B0": {Ia: fp(0.50), Ib: fp(0.35), II: fp(0.30), III: fp(0.25), IV: fp(0.20), V: fp(0.18), VI: fp(0.01)},
	"B5": {Ia: fp(1.67), Ib: fp(0.63), II: fp(0.35), III: fp(0.15), IV: fp(0.13), V: fp(0.09), VI: fp(0.01)},
	"A0": {Ia: fp(3.34), Ib: fp(1.40), II: fp(0.75), III: fp(0.13), IV: fp(0.10), V: fp(0.06)},
	"A5": {Ia: fp(4.17), Ib: fp(2.17), II: fp(1.17), III: fp(0.13), IV: fp(0.07), V: fp(0.05)},
	"F0": {Ia: fp(4.42), Ib: fp(2.50), II: fp(1.33), III: fp(0.13), IV: fp(0.07), V: fp(0.04)},
	"F5": {Ia: fp(5.00), Ib: fp(3.25), II: fp(1.87), III: fp(0.13), IV: fp(0.06), V: fp(0.03)},
	"G0": {Ia: fp(5.21), Ib: fp(3.59), II: fp(2.24), III: fp(0.25), IV: fp(0.07), V: fp(0.03), VI: fp(0.02)},
	"G5": {Ia: fp(5.34), Ib: fp(3.84), II: fp(2.67), III: fp(0.38), IV: fp(0.10), V: fp(0.02), VI: fp(0.02)},
	"K0": {Ia: fp(5.59), Ib: fp(4.17), II: fp(3.17), III: fp(0.50), IV: fp(0.15), V: fp(0.02), VI: fp(0.02)},
	"K5": {Ia: fp(6.17), Ib: fp(4.84), II: fp(4.00), III: fp(1.00), V: fp(0.02), VI: fp(0.01)},
	"M0": {Ia: fp(6.80), Ib: fp(5.42), II: fp(4.59), III: fp(1.68), V: fp(0.02), VI: fp(0.01)},
	"M5": {Ia: fp(7.20), Ib: fp(6.17), II: fp(5.30), III: fp(3.00), V: fp(0.01), VI: fp(0.01)},
	"M9": {Ia: fp(7.80), Ib: fp(6.59), II: fp(5.92), III: fp(4.34), V: fp(0.01), VI: fp(0.01)},
}

// ErrNoMAOForStar reports a "—" cell in the p. 39 MAO table — the
// spectral type / class combination does not exist as a star.
var ErrNoMAOForStar = errors.New("worlds: spectral type / class combination has no MAO entry")

// isPostStellar reports whether a StarKind is post-stellar (BD, D, NS,
// BH, Pulsar) — these have MAO defined in the Special Circumstances
// chapter, not yet encoded.
func isPostStellar(k stars.StarKind) bool {
	switch k {
	case stars.KindBrownDwarf, stars.KindWhiteDwarf,
		stars.KindNeutronStar, stars.KindBlackHole, stars.KindPulsar:
		return true
	}
	return false
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
// Post-stellar kinds return ErrPostStellarPrimaryUnsupported.
// Combinations the book lists as "—" return ErrNoMAOForStar.
func MAO(s stars.Star) (float64, error) {
	if isPostStellar(s.Kind) {
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
		for i, c := range sys.Companions {
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
			group := Group{Members: []stars.Star{c.Star}}
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
