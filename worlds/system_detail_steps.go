package worlds

import (
	"fmt"

	"wbh/roller"
	"wbh/stars"
)

// runStep5A applies 3A1 passes (body physical, belt details, atmosphere,
// hydrographics, moon refinement) to each placement in detailed.
// Mutates detailed in place. WBH pp.69-100.
func runStep5A(r roller.Roller, detailed []DetailedPlacement, sys stars.System, sp SystemPlacement) error {
	for i := range detailed {
		dp := &detailed[i]

		hzco := dp.Group.HZCO()

		// Body physical (terrestrials only — not belts, not GGs).
		if dp.GGClass == NotGasGiant && dp.SizeCode != "" && dp.SizeCode != "0" && dp.SizeCode != "R" {
			beyondHZCO := 0
			offset := dp.Orbit - hzco
			if offset > 0 {
				beyondHZCO = int(offset)
			}
			dms := BodyPhysicalDMs{
				SizeCode:       dp.SizeCode,
				AtHZCOOrCloser: dp.HZ,
				BeyondHZCO:     beyondHZCO,
				SystemAgeGyr:   sys.Primary.AgeGyr,
			}
			bp, err := GenerateBodyPhysical(r, dp.SizeCode, int(dp.DiameterKm), dms)
			if err != nil {
				return fmt.Errorf("worlds: body physical %s: %w", dp.Designation, err)
			}
			dp.Physical = &bp
		}

		// Belt details (Size 0 only).
		if dp.SizeCode == "0" {
			bd, err := GenerateBeltDetails(r, dp.Orbit, sp.SystemSpread, hzco, sys.Primary.AgeGyr, false, false)
			if err != nil {
				return fmt.Errorf("worlds: belt %s: %w", dp.Designation, err)
			}
			dp.Belt = &bd
		}

		// Atmosphere (HZ-orbit terrestrials only).
		if dp.HZ && dp.GGClass == NotGasGiant && dp.SizeCode != "0" && dp.SizeCode != "R" && dp.SizeCode != "" {
			offset := dp.Orbit - hzco
			atmoCode, err := RollAtmoCode(r, dp.SizeCode, offset)
			if err != nil {
				return fmt.Errorf("worlds: atmosphere %s: %w", dp.Designation, err)
			}
			atmo := Atmosphere{Code: atmoCode}
			if atmoCode == 11 || atmoCode == 12 {
				st, serr := RollCorrosiveInsidiousSubtype(r, dp.SizeCode, dp.Orbit, hzco, atmoCode == 12, false)
				if serr != nil {
					return fmt.Errorf("worlds: atmo subtype %s: %w", dp.Designation, serr)
				}
				atmo.Subtype = st
			}
			press, perr := RollTotalPressure(r, atmoCode)
			if perr != nil {
				return fmt.Errorf("worlds: pressure %s: %w", dp.Designation, perr)
			}
			atmo.Pressure = press
			if atmoCode >= 2 && atmoCode <= 9 {
				frac, ferr := RollOxygenFraction(r, sys.Primary.AgeGyr)
				if ferr != nil {
					return fmt.Errorf("worlds: oxygen %s: %w", dp.Designation, ferr)
				}
				atmo.OxygenPartialPressure = frac * press
			}
			if dp.Physical != nil {
				meanT := tempRangeMidpointK(HZCOOffsetToTempRange(dp.Orbit, hzco))
				atmo.ScaleHeight = DeriveScaleHeight(meanT, dp.Physical.Gravity)
			}
			dp.Atmosphere = &atmo

			// Hydrographics (after atmosphere is known).
			hydro, herr := GenerateHydrographics(r, atmo, dp.SizeCode, HZCOOffsetToTempRange(dp.Orbit, hzco))
			if herr != nil {
				return fmt.Errorf("worlds: hydro %s: %w", dp.Designation, herr)
			}
			dp.Hydrographics = &hydro
		}

		// Moon refinement (any planet with moons).
		if len(dp.Moons) > 0 {
			refinePlacementMoons(r, dp)
		}
	}
	return nil
}

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

// runStep5C applies the 3A2b-temp temperature pass to each placement in
// detailed. Mutates detailed in place. WBH pp.108-126.
func runStep5C(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error {
	for i := range detailed {
		dp := &detailed[i]
		if dp.Body == BodyEmpty {
			continue
		}
		temp, err := GenerateTemperature(r, dp, sys, nil)
		if err != nil {
			return fmt.Errorf("worlds: temperature %s: %w", dp.Designation, err)
		}
		dp.Temperature = temp

		for j := range dp.Moons {
			m := &dp.Moons[j]
			moonDP := buildMoonPlacementView(m, dp)
			moonTemp, err := GenerateTemperature(r, moonDP, sys, dp)
			if err != nil {
				return fmt.Errorf("worlds: moon temperature %s: %w", m.Designation, err)
			}
			m.Temperature = moonTemp
		}
	}
	return nil
}

// runStep5D applies the 3A2b-rederive 2-pass iteration (rederive →
// re-run GenerateTemperature → rederive) to each placement in detailed.
// Mutates detailed in place. WBH pp.79, 81, 96-99, 102.
//
// buildMoonPlacementView copies Atmosphere, Hydrographics, and Physical as
// pointer aliases — mutations through moonDP propagate directly to m without
// any explicit write-back for those fields.
func runStep5D(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error {
	for i := range detailed {
		dp := &detailed[i]
		if dp.Body == BodyEmpty || !dp.HasTemperature() {
			continue
		}
		// Pass 1: rederive using 3A2b-temp's MeanK
		if err := RederiveAtmosphereHydrographics(r, dp, sys, nil); err != nil {
			return fmt.Errorf("worlds: rederive %s pass 1: %w", dp.Designation, err)
		}
		// Re-run temperature with corrected atm/hydro
		temp, err := GenerateTemperature(r, dp, sys, nil)
		if err != nil {
			return fmt.Errorf("worlds: temperature %s pass 2: %w", dp.Designation, err)
		}
		dp.Temperature = temp
		// Pass 2: rederive using corrected MeanK (final)
		if err := RederiveAtmosphereHydrographics(r, dp, sys, nil); err != nil {
			return fmt.Errorf("worlds: rederive %s pass 2: %w", dp.Designation, err)
		}

		// Same 2-pass for moons.
		for j := range dp.Moons {
			m := &dp.Moons[j]
			if !m.HasTemperature() {
				continue
			}
			moonDP := buildMoonPlacementView(m, dp)
			// Pass 1
			if err := RederiveAtmosphereHydrographics(r, moonDP, sys, dp); err != nil {
				return fmt.Errorf("worlds: moon rederive %s pass 1: %w", m.Designation, err)
			}
			// Re-run moon temperature
			moonTemp, err := GenerateTemperature(r, moonDP, sys, dp)
			if err != nil {
				return fmt.Errorf("worlds: moon temperature %s pass 2: %w", m.Designation, err)
			}
			m.Temperature = moonTemp
			// Pass 2
			if err := RederiveAtmosphereHydrographics(r, moonDP, sys, dp); err != nil {
				return fmt.Errorf("worlds: moon rederive %s pass 2: %w", m.Designation, err)
			}
		}
	}
	return nil
}
