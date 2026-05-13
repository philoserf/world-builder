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
//  4. Clears Stage-5 output (ClearStage5Output) and re-runs
//     ApplyClimatePasses to regenerate atmosphere/hydrographics/
//     temperature/geology from the new tidal-lock outputs.
//
// Bodies without a captured snapshot (no tidal-lock case fired in Stage 4)
// are skipped. Bodies with pressure ≤ 2.5 bar are skipped (the DM
// wouldn't fire anyway).
//
// Insert in the Generate pipeline immediately after ApplyClimate.
func ApplyTidalLockReEval(r roller.Roller, u *Universe) error {
	sys := u.System
	for body, parent := range u.AllBodiesWithParent() {
		if body.Kind == BodyEmpty {
			continue
		}
		if body.Atmosphere == nil || body.Atmosphere.Pressure <= 2.5 {
			continue
		}
		if body.preTidalLockSnapshot == nil {
			continue // Stage 4 didn't run tidal lock for this body
		}

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

		// Clear Stage-5 output AFTER the re-roll has consumed the pressure
		// from body.Atmosphere. Then re-run climate from scratch with the
		// new tidal-lock outputs.
		ClearStage5Output(body)
		if err := ApplyClimatePasses(r, body, sys); err != nil {
			return fmt.Errorf("worlds: tidal-lock re-eval climate %s%s: %w", moonTag(parent), body.Designation, err)
		}
	}
	return nil
}
