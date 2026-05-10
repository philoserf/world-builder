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

	// Sub-stage 1: day length.
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if body.Kind == BodyEmpty {
			continue
		}
		dl, err := GenerateDayLength(r, body, sys)
		if err != nil {
			return fmt.Errorf("worlds: stage4 day length %s: %w", body.Designation, err)
		}
		body.DayLength = dl

		for _, child := range body.Children {
			child.Period.Hours = body.Period.Hours
			dl, err := GenerateDayLength(r, child, sys)
			if err != nil {
				return fmt.Errorf("worlds: stage4 moon day length %s: %w", child.Designation, err)
			}
			child.DayLength = dl
		}
	}

	// Sub-stage 2: axial tilt.
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if body.Kind == BodyEmpty {
			continue
		}
		at, err := GenerateAxialTilt(r, body)
		if err != nil {
			return fmt.Errorf("worlds: stage4 axial tilt %s: %w", body.Designation, err)
		}
		body.AxialTilt = at

		for _, child := range body.Children {
			at, err := GenerateAxialTilt(r, child)
			if err != nil {
				return fmt.Errorf("worlds: stage4 moon axial tilt %s: %w", child.Designation, err)
			}
			child.AxialTilt = at
		}
	}

	// Sub-stage 3: tidal lock. Planet uses its own Period.Hours; moon
	// uses its PeriodHours (orbit around planet).
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if body.Kind == BodyEmpty {
			continue
		}
		tl, err := GenerateTidalLock(r, body, nil, sys, nil, body.Period.Hours)
		if err != nil {
			return fmt.Errorf("worlds: stage4 tidal lock %s: %w", body.Designation, err)
		}
		body.TidalLock = tl

		for _, child := range body.Children {
			tl, err := GenerateTidalLock(r, child, child, sys, body, child.PeriodHours)
			if err != nil {
				return fmt.Errorf("worlds: stage4 moon tidal lock %s: %w", child.Designation, err)
			}
			child.TidalLock = tl
		}
	}

	// Sub-stage 4: surface tidal effects.
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if body.Kind == BodyEmpty {
			continue
		}
		ste, err := GenerateSurfaceTidalEffects(body, nil, sys, nil)
		if err != nil {
			return fmt.Errorf("worlds: stage4 surface tidal %s: %w", body.Designation, err)
		}
		body.TidalEffects = ste

		for _, child := range body.Children {
			ste, err := GenerateSurfaceTidalEffects(child, child, sys, body)
			if err != nil {
				return fmt.Errorf("worlds: stage4 moon surface tidal %s: %w", child.Designation, err)
			}
			child.TidalEffects = ste
		}
	}

	return nil
}
