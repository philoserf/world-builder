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

// runStep5E applies the 3B-geology pass: residual seismic stress, tidal
// stress factor, tidal heating factor, total seismic stress, GG residual
// heat, post-TSS temperature recompute, and tectonic plates. Mutates
// detailed in place. WBH pp.125-127.
//
// Per-body dice budget: 1 × 2D per terrestrial body that passes plate
// prerequisites (Hydro ≥ 1, TSS > 0). Otherwise zero. Moons add the same
// conditional 2D each.
//
//nolint:unparam // error return kept for pipeline consistency with runStep5A-5D.
func runStep5E(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error {
	for i := range detailed {
		dp := &detailed[i]
		if dp.Body == BodyEmpty {
			continue
		}
		// Belts (Size 0) get no Geology.
		if dp.SizeCode == "0" {
			continue
		}

		dp.Geology = computeGeology(r, dp, sys)

		// Temperature recompute (in place) per WBH p.125.
		if dp.Temperature != nil {
			ApplyInherentTempAddition(dp.Temperature, dp.Geology.InherentTemperatureK)
		}

		// Process moons.
		for j := range dp.Moons {
			m := &dp.Moons[j]
			m.Geology = computeMoonGeology(r, m, dp, sys)
			if m.Temperature != nil {
				ApplyInherentTempAddition(m.Temperature, m.Geology.InherentTemperatureK)
			}
		}
	}
	return nil
}

// computeGeology populates a Geology for the given body. Caller is
// responsible for applying ApplyInherentTempAddition afterwards.
func computeGeology(r roller.Roller, dp *DetailedPlacement, sys stars.System) *Geology {
	g := &Geology{}
	if dp.Body == BodyGasGiant {
		g.InherentTemperatureK = ComputeGGResidualHeat(dp.MassEarth, sys.Primary.AgeGyr)
		// GGs don't get seismic factors or plates.
		return g
	}
	// Terrestrial path.
	g.ResidualSeismicStress = ComputeResidualSeismicStress(dp, sys.Primary.AgeGyr, false)
	g.TidalStressFactor = ComputeTidalStressFactor(dp)
	g.TidalHeatingFactor = ComputeTidalHeatingFactor(planetTidalHeatingInputs(dp, sys))
	g.TotalSeismicStress = g.ResidualSeismicStress + g.TidalStressFactor + g.TidalHeatingFactor
	g.InherentTemperatureK = float64(g.TotalSeismicStress)
	g.TectonicPlates = RollTectonicPlates(r, dp, g.TotalSeismicStress)
	return g
}

// computeMoonGeology populates a Geology for a moon. Builds a synthetic
// DetailedPlacement view via buildMoonPlacementView; uses the parent
// planet's mass for tidal heating instead of the primary star's.
func computeMoonGeology(r roller.Roller, m *Moon, parent *DetailedPlacement, sys stars.System) *Geology {
	moonDP := buildMoonPlacementView(m, parent)
	// buildMoonPlacementView does not copy TidalEffects; set it explicitly
	// so ComputeTidalStressFactor can read it.
	moonDP.TidalEffects = m.TidalEffects
	g := &Geology{}
	if moonDP.Body == BodyGasGiant {
		// Rare moon-of-GG-class scenario; reuse the GG formula.
		g.InherentTemperatureK = ComputeGGResidualHeat(m.MassEarth, sys.Primary.AgeGyr)
		return g
	}
	g.ResidualSeismicStress = ComputeResidualSeismicStress(moonDP, sys.Primary.AgeGyr, true)
	g.TidalStressFactor = ComputeTidalStressFactor(moonDP)
	g.TidalHeatingFactor = ComputeTidalHeatingFactor(moonTidalHeatingInputs(m, parent))
	g.TotalSeismicStress = g.ResidualSeismicStress + g.TidalStressFactor + g.TidalHeatingFactor
	g.InherentTemperatureK = float64(g.TotalSeismicStress)
	g.TectonicPlates = RollTectonicPlates(r, moonDP, g.TotalSeismicStress)
	return g
}

// planetTidalHeatingInputs derives TidalHeatingInputs for a planet around
// its primary star. Converts AU → Mkm and hours → days.
func planetTidalHeatingInputs(dp *DetailedPlacement, sys stars.System) TidalHeatingInputs {
	const auMkm = 149.6 // Mkm per AU
	const solarMassEarth = 332946.0
	return TidalHeatingInputs{
		PrimaryMassEarth: sys.Primary.Mass * solarMassEarth,
		SizeN:            SizeAsInt(dp.SizeCode),
		Eccentricity:     dp.Eccentricity,
		DistanceMkm:      stars.OrbitToAU(dp.Orbit) * auMkm,
		PeriodDays:       dp.Period.Hours / 24.0,
		WorldMassEarth:   bodyMassEarth(dp),
	}
}

// moonTidalHeatingInputs derives TidalHeatingInputs for a moon around its
// parent planet. Distance and period are already in km/hours; convert.
func moonTidalHeatingInputs(m *Moon, parent *DetailedPlacement) TidalHeatingInputs {
	return TidalHeatingInputs{
		PrimaryMassEarth: bodyMassEarth(parent),
		SizeN:            SizeAsInt(m.SizeCode),
		Eccentricity:     m.Eccentricity,
		DistanceMkm:      float64(m.OrbitKm) / 1_000_000.0,
		PeriodDays:       m.PeriodHours / 24.0,
		WorldMassEarth:   moonMassEarth(m),
	}
}

// bodyMassEarth returns dp.MassEarth if populated, else falls back to
// DeriveMass(density, diameter) when Physical is available. Returns 0
// only when both paths are unavailable.
//
// Needed because DetailSystem only populates dp.MassEarth for gas giants;
// terrestrial bodies leave it 0 and rely on derived mass at call sites.
// Step 5E's tidal heating computation needs the actual mass for the WBH
// p.126 formula's WorldMass⊕ term.
func bodyMassEarth(dp *DetailedPlacement) float64 {
	if dp == nil {
		return 0
	}
	if dp.MassEarth != 0 {
		return dp.MassEarth
	}
	if dp.Physical != nil {
		return DeriveMass(dp.Physical.Density, dp.DiameterKm)
	}
	return 0
}

// moonMassEarth returns m.MassEarth if populated, else falls back to
// DeriveMass(density, diameter). Same rationale as bodyMassEarth but
// for moons (which use DiameterKm directly, not Earth-relative).
func moonMassEarth(m *Moon) float64 {
	if m == nil {
		return 0
	}
	if m.MassEarth != 0 {
		return m.MassEarth
	}
	if m.Physical != nil {
		return DeriveMass(m.Physical.Density, m.DiameterKm)
	}
	return 0
}

// runStep5F applies the 3B-biology pass: native lifeform ratings (Biomass,
// Biocomplexity, Sophont bools, Biodiversity, Compatibility) + Resource
// Rating. Mutates detailed in place. WBH pp.127-131.
//
// Body filter: terrestrials with Atmosphere; HZ-planet moons with
// Atmosphere. Skip GGs, belts (Size 0), empty placements, and terrestrials
// without atmosphere data.
//
// Per-body dice budget: up to 7 × 2D when life is present; 2 × 2D when
// biomass = 0 (Biomass + Resource); 0 × 2D when body is skipped.
//
//nolint:unparam // matches sibling runStep5* signatures (always-nil error)
func runStep5F(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error {
	for i := range detailed {
		dp := &detailed[i]
		if !biologyApplies(dp) {
			continue
		}
		dp.Biology = computeBiology(r, dp, sys.Primary.AgeGyr)

		for j := range dp.Moons {
			m := &dp.Moons[j]
			if m.Atmosphere == nil {
				continue
			}
			moonDP := buildMoonPlacementView(m, dp)
			// buildMoonPlacementView does not copy Temperature; set it
			// explicitly so RollBiomass can read m.Temperature for the
			// temperature DMs (precedent: same pattern in computeMoonGeology
			// for TidalEffects per 3B-geology Task 8).
			moonDP.Temperature = m.Temperature
			m.Biology = computeBiology(r, moonDP, sys.Primary.AgeGyr)
		}
	}
	return nil
}

// biologyApplies reports whether dp should receive a Biology struct.
// True for terrestrial bodies with Atmosphere; false for empty, belts,
// gas giants, or atmosphere-less terrestrials.
func biologyApplies(dp *DetailedPlacement) bool {
	if dp == nil || dp.Body == BodyEmpty {
		return false
	}
	if dp.Body == BodyGasGiant || dp.Body == BodyPlanetoidBelt {
		return false
	}
	if dp.SizeCode == "0" {
		return false
	}
	return dp.Atmosphere != nil
}

// computeBiology populates a Biology for the given body. Caller has
// already verified biologyApplies(dp).
func computeBiology(r roller.Roller, dp *DetailedPlacement, ageGyr float64) *Biology {
	bio := &Biology{}
	bio.Biomass = RollBiomass(r, dp, ageGyr)
	if bio.Biomass > 0 {
		bio.Biocomplexity = RollBiocomplexity(r, dp, bio.Biomass, ageGyr)
		if bio.Biocomplexity >= 8 {
			bio.HasNativeSophont = RollNativeSophont(r, bio.Biocomplexity)
			bio.HadExtinctSophont = RollExtinctSophont(r, bio.Biocomplexity, ageGyr)
		}
		bio.Biodiversity = RollBiodiversity(r, bio.Biomass, bio.Biocomplexity)
		bio.Compatibility = RollCompatibility(r, dp, bio.Biocomplexity, ageGyr)
	}
	bio.ResourceRating = RollTerrestrialResourceRating(r, dp, bio)
	return bio
}
