package worlds

import (
	"fmt"

	"wbh/roller"
	"wbh/stars"
)

// runStep5B applies 3A2a passes (surface distribution, day length, axial
// tilt, tidal lock, surface tidal effects) to each placement in detailed.
// Mutates detailed in place. WBH pp.100-108.
func runStep5B(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error {
	// 5B.1 — surface feature distribution (per terrestrial + per HZ-planet moon).
	for i := range detailed {
		dp := &detailed[i]
		if dp.HasHydrographics() {
			sd, err := GenerateSurfaceDistribution(r, dp.Hydrographics)
			if err != nil {
				return fmt.Errorf("worlds: surface distribution %s: %w", dp.Designation, err)
			}
			dp.SurfaceDistribution = sd
		}
		// Per-moon surface distribution for HZ-planet moons.
		if dp.HZ {
			for j := range dp.Moons {
				m := &dp.Moons[j]
				if m.Hydrographics != nil {
					sd, err := GenerateSurfaceDistribution(r, m.Hydrographics)
					if err != nil {
						return fmt.Errorf("worlds: moon surface distribution %s: %w", m.Designation, err)
					}
					m.SurfaceDistribution = sd
				}
			}
		}
	}

	// 5B.2 — day length (per body + per moon).
	for i := range detailed {
		dp := &detailed[i]
		if dp.Body == BodyEmpty {
			continue
		}
		dl, err := GenerateDayLength(r, dp, sys)
		if err != nil {
			return fmt.Errorf("worlds: day length %s: %w", dp.Designation, err)
		}
		dp.DayLength = dl

		for j := range dp.Moons {
			m := &dp.Moons[j]
			// Build a synthetic DetailedPlacement view for the moon. YearDays /
			// SolarHours are calendar quantities — a moon's year is its parent's
			// year around the star (they co-orbit), NOT the moon's orbit around
			// the planet (m.PeriodHours, which is a synodic month).
			moonDP := &DetailedPlacement{
				Period:   Period{Hours: dp.Period.Hours},
				SizeCode: m.SizeCode,
				GGClass:  m.GGClass,
			}
			moonDP.Body = BodyTerrestrial
			if m.GGClass != NotGasGiant {
				moonDP.Body = BodyGasGiant
			}
			dl, err := GenerateDayLength(r, moonDP, sys)
			if err != nil {
				return fmt.Errorf("worlds: moon day length %s: %w", m.Designation, err)
			}
			m.DayLength = dl
		}
	}

	// 5B.3 — axial tilt (per body + per moon).
	for i := range detailed {
		dp := &detailed[i]
		if dp.Body == BodyEmpty {
			continue
		}
		at, err := GenerateAxialTilt(r, dp)
		if err != nil {
			return fmt.Errorf("worlds: axial tilt %s: %w", dp.Designation, err)
		}
		dp.AxialTilt = at

		for j := range dp.Moons {
			m := &dp.Moons[j]
			moonDP := &DetailedPlacement{SizeCode: m.SizeCode}
			moonDP.Body = BodyTerrestrial
			if m.GGClass != NotGasGiant {
				moonDP.Body = BodyGasGiant
			}
			at, err := GenerateAxialTilt(r, moonDP)
			if err != nil {
				return fmt.Errorf("worlds: moon axial tilt %s: %w", m.Designation, err)
			}
			m.AxialTilt = at
		}
	}

	// 5B.4 — tidal lock (per body + per moon).
	for i := range detailed {
		dp := &detailed[i]
		if dp.Body == BodyEmpty {
			continue
		}
		tl, err := GenerateTidalLock(r, dp, nil, sys, nil, dp.Period.Hours)
		if err != nil {
			return fmt.Errorf("worlds: tidal lock %s: %w", dp.Designation, err)
		}
		dp.TidalLock = tl

		for j := range dp.Moons {
			m := &dp.Moons[j]
			// Build a moon-side DetailedPlacement view that carries the moon's
			// own size/eccentricity/axial-tilt/atmosphere/etc. for DM evaluation.
			moonDP := buildMoonPlacementView(m, dp)
			tl, err := GenerateTidalLock(r, moonDP, m, sys, dp, m.PeriodHours)
			if err != nil {
				return fmt.Errorf("worlds: moon tidal lock %s: %w", m.Designation, err)
			}
			m.TidalLock = tl
			// DayLength and AxialTilt are aliased through buildMoonPlacementView's
			// pointer copies, so mutations inside ApplyTidalLockEffect already reach
			// the Moon. Eccentricity is a value field on embedded Placement, so it
			// must be written back explicitly when a 1:1 lock rerolled it.
			m.Eccentricity = moonDP.Eccentricity
		}
	}

	// 5B.5 — surface tidal effects (per body + per moon).
	for i := range detailed {
		dp := &detailed[i]
		if dp.Body == BodyEmpty {
			continue
		}
		te, err := GenerateSurfaceTidalEffects(dp, nil, sys, nil)
		if err != nil {
			return fmt.Errorf("worlds: tidal effects %s: %w", dp.Designation, err)
		}
		dp.TidalEffects = te

		for j := range dp.Moons {
			m := &dp.Moons[j]
			moonDP := buildMoonPlacementView(m, dp)
			te, err := GenerateSurfaceTidalEffects(moonDP, m, sys, dp)
			if err != nil {
				return fmt.Errorf("worlds: moon tidal effects %s: %w", m.Designation, err)
			}
			m.TidalEffects = te
		}
	}
	return nil
}
