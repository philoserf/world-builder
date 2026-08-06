package worlds

import "github.com/philoserf/world-builder/roller"

// SPIKE/PLAN C1 (docs/rebuild-spec.md § C1, docs/c1-subroller-plan.md):
// every per-body suffix stage draws its dice from a substream keyed by
// stable body identity and procedure family, rather than from the single
// shared stream. Because roller.Seeded.Fork derives from the immutable
// construction seed, a stage can fork off the r it is already handed and
// get the same substream regardless of how much earlier stages consumed —
// so each body's outputs are invariant to stage reordering and to another
// body being re-rolled (the tidal-lock cascade). With a Scripted roller
// Fork is transparent, so worked-example fixtures are unaffected.
//
// The structure prefix (stars, placement, ApplyDetailFrontEnd) stays on
// the shared stream: body identity is created there, so there is nothing
// stable to key on, and none of the C1 pain lives before it.

// bodyForkID returns a stable, unique fork key for a body's substream.
// Designations are unique among top-level bodies; a moon's designation is
// prefixed with its parent's to stay unique system-wide. Assigned in
// ApplyDetailFrontEnd (Stage 2), so it is stable for every suffix stage.
func bodyForkID(body, parent *Body) string {
	if parent != nil {
		return parent.Designation + "/" + body.Designation
	}

	return body.Designation
}

// bodySub forks the per-body substream for a given procedure family off
// the shared roller r. The two-level key (identity, family) matches the
// tree in docs/rebuild-spec.md § C1: r.Fork(bodyID).Fork(family).
func bodySub(r roller.Roller, body, parent *Body, family string) roller.Roller {
	return r.Fork(bodyForkID(body, parent)).Fork(family)
}
