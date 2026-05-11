package worlds

import (
	"fmt"

	"wbh/roller"
)

// ApplyRotationTilt runs the WBH pp.100-108 (3A2a) per-body rotation/
// tilt/tide pass: day length, axial tilt, tidal lock, surface tidal
// effects. Stage-4 orchestrator.
//
// Per dependency-graph.md § Stage 4, surface distribution moves to
// Stage 6 (post-climate); cycle-4 does not run it. Per anti-pattern
// A.1, every moon is walked alongside its parent.
//
// Notes:
//   - For moons, calendar quantities (YearDays, SolarHours) use the
//     parent's period (a moon's year around the star equals its
//     parent's). This procedure sets child.Period.Hours = parent's
//     before calling GenerateDayLength so the result is correct.
//   - Tidal lock for a moon uses the moon's PeriodHours (the moon's
//     orbit around its planet) as the relevant yearHours; for a
//     planet, it uses the planet's Period.Hours (year around star).
func ApplyRotationTilt(r roller.Roller, u *Universe) error {
	sys := u.System

	// Sub-stage 1: day length. Moons inherit the parent's stellar period
	// before the per-body call (their year around the star equals the
	// parent's).
	for body, parent := range u.AllBodiesWithParent() {
		if body.Kind == BodyEmpty {
			continue
		}
		if parent != nil {
			body.Period.Hours = parent.Period.Hours
		}
		dl, err := GenerateDayLength(r, body, sys)
		if err != nil {
			return fmt.Errorf("worlds: stage4 day length %s%s: %w", moonTag(parent), body.Designation, err)
		}
		body.DayLength = dl
	}

	// Sub-stage 2: axial tilt. GenerateAxialTilt itself is
	// parent-independent; the parent here is used only to keep the
	// "moon "-prefixed error message consistent with the other sub-stages.
	for body, parent := range u.AllBodiesWithParent() {
		if body.Kind == BodyEmpty {
			continue
		}
		at, err := GenerateAxialTilt(r, body)
		if err != nil {
			return fmt.Errorf("worlds: stage4 axial tilt %s%s: %w", moonTag(parent), body.Designation, err)
		}
		body.AxialTilt = at
	}

	// Sub-stage 3: tidal lock. Planet uses its own Period.Hours; moon
	// uses its PeriodHours (orbit around planet) and passes the parent
	// planet so GenerateTidalLock can resolve the moon-vs-planet branch.
	for body, parent := range u.AllBodiesWithParent() {
		if body.Kind == BodyEmpty {
			continue
		}
		var moonRef *Body
		hours := body.Period.Hours
		if parent != nil {
			moonRef = body
			hours = body.PeriodHours
		}
		tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, hours)
		if err != nil {
			return fmt.Errorf("worlds: stage4 tidal lock %s%s: %w", moonTag(parent), body.Designation, err)
		}
		body.TidalLock = tl
	}

	// Sub-stage 4: surface tidal effects.
	for body, parent := range u.AllBodiesWithParent() {
		if body.Kind == BodyEmpty {
			continue
		}
		var moonRef *Body
		if parent != nil {
			moonRef = body
		}
		ste, err := GenerateSurfaceTidalEffects(body, moonRef, sys, parent)
		if err != nil {
			return fmt.Errorf("worlds: stage4 surface tidal %s%s: %w", moonTag(parent), body.Designation, err)
		}
		body.TidalEffects = ste
	}

	return nil
}

// moonTag yields "moon " when the iterating body is a moon (parent
// non-nil), empty otherwise. Used to preserve the moon-vs-planet split
// in Stage 4 error message prefixes after the loop unification.
func moonTag(parent *Body) string {
	if parent != nil {
		return "moon "
	}
	return ""
}
