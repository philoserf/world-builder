package worlds

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
