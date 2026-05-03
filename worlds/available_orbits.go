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
