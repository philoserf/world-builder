// Package worlds — per-body habitability rating per WBH p.132-133
// (sub-project 3B-final).
package worlds

// Habitability — a per-body habitability rating for Terragens per WBH
// p.132-133. Computed by Step 5G for any non-empty terrestrial body
// (and HZ-planet moons).
//
// Range: 0-12. The book theoretically allows higher but treats 12 as
// "very unlikely" and clamps negative results to 0.
//
// Ratings interpretation (WBH p.133):
//
//	0       — Actively hostile world: not survivable without specialised equipment
//	1-2     — Barely habitable: full protective equipment needed
//	3-5     — Marginally survivable with proper equipment
//	6-7     — Regionally habitable: may require acclimation
//	8-9     — Suitable for human habitation with minimal equipment or acclimation
//	10-12   — Terra-equivalent garden world (10/A is the Terran baseline)
type Habitability struct {
	Rating int

	// Notes is a referee-color string visible in the Class IV-P form's
	// Habitability section (e.g., "High temperatures hinder habitability").
	// Currently always empty — populated by future referee-feature carry-forward.
	Notes string
}

// ComputeHabitability per WBH p.132: 10 + DMs, clamped to [0, 12].
// Deterministic — no dice. Operates on body's current Atmosphere /
// Hydrographics / Temperature / Physical / SizeCode / TidalLock fields.
//
// Returns Habitability{Rating: 0} if body is nil. For bodies with
// missing pointer fields, the corresponding DMs are skipped (treated
// as 0) — defensive but documented as caller's responsibility.
//
// Skipped: low-oxygen-taint DM-2 deferred per spec Q3-a (taint
// typology not yet modeled).
func ComputeHabitability(body *DetailedPlacement) Habitability {
	if body == nil {
		return Habitability{Rating: 0}
	}
	dm := habitabilitySizeDM(SizeAsInt(body.SizeCode))
	dm += habitabilityAtmDM(body)
	dm += habitabilityHydroDM(body)
	dm += habitabilityTidalLockDM(body)

	// Temperature + Gravity DMs added in Task 3.

	rating := min(max(10+dm, 0), 12)
	return Habitability{Rating: rating}
}

// habitabilitySizeDM per WBH p.132 size-DM table.
func habitabilitySizeDM(size int) int {
	switch {
	case size <= 4:
		return -1
	case size >= 9:
		return +1
	}
	return 0
}

// habitabilityAtmDM per WBH p.132 atmosphere-DM table.
// nil Atmosphere is treated as atm code 0 (vacuum) → DM-8.
func habitabilityAtmDM(body *DetailedPlacement) int {
	atmCode := 0
	if body.Atmosphere != nil {
		atmCode = body.Atmosphere.Code
	}
	switch atmCode {
	case 0, 1, 10: // 0, 1, A
		return -8
	case 2, 14: // 2, E
		return -4
	case 3, 13: // 3, D
		return -3
	case 4, 9:
		return -2
	case 5, 7, 8:
		return -1
	case 6:
		return 0 // baseline
	case 11: // B
		return -10
	case 12, 15: // C, F+
		return -12
	}
	return 0
}

// habitabilityHydroDM per WBH p.132 hydrographics-DM table.
// nil Hydrographics is treated as Hydro code 0 → DM-4.
func habitabilityHydroDM(body *DetailedPlacement) int {
	hydroCode := 0
	if body.Hydrographics != nil {
		hydroCode = body.Hydrographics.Code
	}
	switch {
	case hydroCode == 0:
		return -4
	case hydroCode >= 1 && hydroCode <= 3:
		return -2
	case hydroCode == 9:
		return -1
	case hydroCode >= 10:
		return -2
	}
	return 0 // 4-8
}

// habitabilityTidalLockDM per WBH p.132: "Solar tidally locked (1:1)
// world" → DM-2. Detection: TidalLock.IsTwilightZone (which is true
// only when Case == PlanetToStar AND LockRatio == "1:1").
func habitabilityTidalLockDM(body *DetailedPlacement) int {
	if body.TidalLock == nil {
		return 0
	}
	if body.TidalLock.IsTwilightZone {
		return -2
	}
	return 0
}
