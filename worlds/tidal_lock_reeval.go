package worlds

import (
	"fmt"

	"wbh/roller"
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
	for body, parent := range u.AllBodiesWithParent() {
		if body.Kind == BodyEmpty || parent == nil {
			continue
		}
		if err := reEvalBody(r, u, body, parent); err != nil {
			return err
		}
	}
	// Pass 2: planets and belts. Parent is nil for top-level bodies.
	for body, parent := range u.AllBodiesWithParent() {
		if body.Kind == BodyEmpty || parent != nil {
			continue
		}
		if err := reEvalBody(r, u, body, parent); err != nil {
			return err
		}
	}
	return nil
}

// reEvalBody performs the tidal-lock re-evaluation for a single body.
// Skipped if the body has no pre-tidal-lock snapshot or pressure ≤ 2.5
// bar. Factored out of ApplyTidalLockReEval so both passes call the
// same logic.
func reEvalBody(r roller.Roller, u *Universe, body, parent *Body) error {
	if body.Atmosphere == nil || body.Atmosphere.Pressure <= 2.5 {
		return nil
	}
	if body.preTidalLockSnapshot == nil {
		return nil // Stage 4 didn't run tidal lock for this body
	}

	sys := u.System

	// Restore pre-tidal-lock state. Atmosphere stays as-is so
	// commonTidalLockDMs can see the pressure on the re-roll.
	body.preTidalLockSnapshot.RestoreInto(body)
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
