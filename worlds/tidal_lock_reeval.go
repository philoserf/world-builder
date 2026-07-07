package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
)

// ClearStage5Output zeroes the per-body fields populated by
// ApplyClimatePasses (Stage 5). Called before re-running Stage 5 for a
// body whose tidal-lock outputs changed during the atmosphere-DM
// re-evaluation cascade (WBH p.106).
//
// Stage 5 writes four fields on Body:
//
//   - Atmosphere    — initial roll + passes
//   - Hydrographics — initial roll + passes
//   - Temperature   — per climatePass
//   - Geology       — partial geology (Residual + TSF + THF); Stage 7
//     extends this if non-nil, recomputes from scratch if nil.
//     Setting to nil here is safe: Stage 7 handles the nil case cleanly
//     via the computePartialGeology → RollTectonicPlates path.
//
// Stage 4 fields (DayLength, AxialTilt, TidalLock, TidalEffects,
// Eccentricity) are NOT cleared — they are either restored by
// PreTidalLockSnapshot.RestoreInto or re-set by the re-eval's own
// GenerateTidalLock call.
func ClearStage5Output(body *Body) {
	body.Atmosphere = nil
	body.Hydrographics = nil
	body.Temperature = nil
	body.Geology = nil
}

// ApplyTidalLockReEval performs the WBH p.106 atmosphere-DM re-evaluation
// cascade. After ApplyClimate has set body.Atmosphere, this pass:
//  1. Identifies bodies with atmospheric pressure > 2.5 bar (the only
//     pressure that triggers the -2 atmosphere DM in commonTidalLockDMs).
//  2. Restores the pre-tidal-lock snapshot of Eccentricity, AxialTilt,
//     and DayLength.
//  3. Clears body.TidalLock and re-runs GenerateTidalLock — now with
//     body.Atmosphere set, so commonTidalLockDMs sees the pressure and
//     applies the -2 DM.
//  4. Recomputes body.TidalEffects from the new TidalLock (mirror of
//     Stage 4 sub-stage 4). TidalEffects depends on the lock ratio; the
//     Stage-4 value is stale after the re-roll.
//  5. Clears Stage-5 output (ClearStage5Output) and re-runs
//     ApplyClimatePasses to regenerate atmosphere/hydrographics/
//     temperature/geology from the new tidal-lock outputs.
//
// Walk order mirrors Stage 4 sub-stage 3: moons first, then
// planets/belts. This ensures that when a parent planet's re-eval
// evaluates hasLockedMoon (which reads moon.TidalLock), the moon's
// TidalLock has already been updated by the moon's own re-eval pass.
//
// Bodies without a captured snapshot (no tidal-lock case fired in Stage 4)
// are skipped. Bodies with pressure ≤ 2.5 bar are skipped (the DM
// wouldn't fire anyway).
//
// Insert in the Generate pipeline immediately after ApplyClimate.
func ApplyTidalLockReEval(r roller.Roller, u *Universe) error {
	// Pass 1: moons. Re-eval moons first so their TidalLock is current
	// by the time the parent planet's re-eval reads hasLockedMoon.
	// Parents whose moons change lock state are recorded: a terrestrial
	// planet's PlanetToMoon case (its DM and day length) is derived from
	// a child moon's lock, so a moon-side change staleness-forces the
	// parent's re-eval even when the parent's own pressure gate wouldn't
	// fire — both when the parent is already PlanetToMoon (its derived
	// data went stale) and when a newly-locked moon makes PlanetToMoon a
	// candidate that could now win over the parent's Stage-4 case.
	changedParents := map[*Body]bool{}
	for body, parent := range u.AllBodiesWithParent() {
		if body.Kind == BodyEmpty || parent == nil {
			continue
		}
		before := lockFingerprint(body)
		if err := reEvalBody(r, u, body, parent, false); err != nil {
			return err
		}
		if lockFingerprint(body) != before {
			changedParents[parent] = true
		}
	}
	// Pass 2: planets and belts. Parent is nil for top-level bodies.
	for body, parent := range u.AllBodiesWithParent() {
		if body.Kind == BodyEmpty || parent != nil {
			continue
		}
		// Force re-eval of any PlanetToMoon-candidate parent (terrestrial,
		// significant moon) whose moon-set changed lock state — not just
		// parents already in that case. reEvalBody re-runs GenerateTidalLock,
		// which re-picks the winning case, so a newly-locked moon can
		// promote the parent to PlanetToMoon and a newly-unlocked moon can
		// demote it.
		force := changedParents[body] && isTerrestrial(body) && hasSignificantMoon(body)
		if err := reEvalBody(r, u, body, parent, force); err != nil {
			return err
		}
	}
	return nil
}

// lockFingerprint summarizes a body's tidal-lock outcome for
// change detection across the re-eval passes.
func lockFingerprint(body *Body) string {
	if body.TidalLock == nil {
		return ""
	}
	return fmt.Sprintf("%d|%s", body.TidalLock.Case, body.TidalLock.LockRatio)
}

// reEvalBody performs the tidal-lock re-evaluation for a single body.
// Skipped if the body has no pre-tidal-lock snapshot or (unless force)
// pressure ≤ 2.5 bar. force is set for a planet in a PlanetToMoon lock
// whose referenced moon changed lock state in pass 1 — its DM and day
// length are derived from that moon and would otherwise go stale.
// Factored out of ApplyTidalLockReEval so both passes call the same
// logic.
//
// Known, accepted asymmetry: the -2 atmosphere DM is applied from the
// pre-re-eval pressure, and the climate re-run afterwards can settle on
// a pressure ≤ 2.5 bar that would no longer justify it. This mirrors
// stage5.go's documented non-convergence: each climate pass is a
// stochastic sample, not a fixed point, and the book's p.106 cascade
// re-rolls forward only — it does not iterate to consistency.
func reEvalBody(r roller.Roller, u *Universe, body, parent *Body, force bool) error {
	if !force && (body.Atmosphere == nil || body.Atmosphere.Pressure <= 2.5) {
		return nil
	}
	if body.preTidalLockSnapshot == nil {
		return nil // Stage 4 didn't run tidal lock for this body
	}

	sys := u.System

	// Restore pre-tidal-lock state. Atmosphere stays as-is so
	// commonTidalLockDMs can see the pressure on the re-roll. The
	// snapshot's lifetime is "set once by Stage 4, consumed at most
	// once here" — clear it so an accidental second invocation fails
	// closed (skips) instead of reusing stale data.
	body.preTidalLockSnapshot.RestoreInto(body)
	body.preTidalLockSnapshot = nil
	body.TidalLock = nil

	// Re-run tidal lock — now with atmosphere DM active.
	// Moons pass themselves as moonRef and use PeriodHours (orbit
	// around planet); planets pass nil and use stellar Period.Hours.
	var moonRef *Body
	var yearHours float64
	if parent != nil {
		moonRef = body
		yearHours = body.PeriodHours
	} else {
		yearHours = body.Period.Hours
	}
	tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, yearHours)
	if err != nil {
		return fmt.Errorf("worlds: tidal-lock re-eval %s%s: %w", moonTag(parent), body.Designation, err)
	}
	body.TidalLock = tl

	// Recompute surface tidal effects from the new lock ratio. The
	// Stage-4 value in body.TidalEffects is stale: GenerateSurfaceTidalEffects
	// depends on body.TidalLock.LockRatio (e.g. a 1:1 locked moon
	// produces no planet→moon tide). Must happen BEFORE ClearStage5Output
	// so that ApplyClimatePasses (ComputeTidalStressFactor) reads fresh data.
	ste, err := GenerateSurfaceTidalEffects(body, moonRef, sys, parent)
	if err != nil {
		return fmt.Errorf("worlds: tidal-lock re-eval surface tidal %s%s: %w", moonTag(parent), body.Designation, err)
	}
	body.TidalEffects = ste

	// Clear Stage-5 output AFTER the re-roll has consumed the pressure
	// from body.Atmosphere. Then re-run climate from scratch with the
	// new tidal-lock outputs.
	ClearStage5Output(body)
	if err := ApplyClimatePasses(r, body, sys); err != nil {
		return fmt.Errorf("worlds: tidal-lock re-eval climate %s%s: %w", moonTag(parent), body.Designation, err)
	}
	return nil
}
