package worlds

import (
	"iter"

	"wbh/iiss"
	"wbh/stars"
)

// Universe is the top-level container produced by the full pipeline.
// Per docs/pass-2/api-surface.md § The Universe model, this is the
// handoff to renderers and to cmd/wbh.
type Universe struct {
	System    stars.System
	Placement SystemPlacement
	Detail    SystemDetail
}

// SystemDetail aggregates the per-body bodies and the system-wide
// aggregations (profiles, IISS forms, mainworld pick).
//
// The IISS-form fields are carried via an embedded iiss.SystemForms so
// iiss/ can render the system without importing worlds/. cmd/wbh
// extracts u.Detail.SystemForms and passes it to iiss.MarkdownSystem.
type SystemDetail struct {
	Bodies        []Body
	Allocations   []StarAllocation
	BaselineN     int
	BaselineOrbit float64
	EmptyOrbits   int
	SystemSpread  float64

	// Mainworld is the auto-picked mainworld body, set by AggregateSystem
	// alongside MainworldDesignation. nil when the system has no
	// terrestrial / moon / belt bodies (the body kinds pickMainworld
	// admits as candidates).
	Mainworld *Body

	iiss.SystemForms
}

// AllBodies yields every Body in the universe (planets, moons, belts) in
// ascending-orbit order within each star group, with each body's children
// yielded immediately after the parent. This order is contract — the
// LongProfile and AssignPlanetDesignations procedures rely on it.
func (u *Universe) AllBodies() iter.Seq[*Body] {
	return func(yield func(*Body) bool) {
		for i := range u.Detail.Bodies {
			body := &u.Detail.Bodies[i]
			if !yield(body) {
				return
			}
			for _, child := range body.Children {
				if !yield(child) {
					return
				}
			}
		}
	}
}

// Bodies filters AllBodies to a predicate. Iteration order is not
// contract — consumers that need a specific order use AllBodies and
// filter inline. This leaves room for future order-agnostic callers
// (per-body climate convergence does not need ordering).
func (u *Universe) Bodies(filter func(*Body) bool) iter.Seq[*Body] {
	return func(yield func(*Body) bool) {
		for body := range u.AllBodies() {
			if !filter(body) {
				continue
			}
			if !yield(body) {
				return
			}
		}
	}
}
