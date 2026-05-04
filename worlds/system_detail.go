package worlds

import (
	"fmt"
	"math"

	"wbh/roller"
	"wbh/stars"
)

// DetailSystem composes the full WBH pp. 53-67 procedure on top of a
// SystemPlacement (2B output). Returns a SystemDetail with sizes,
// moons, designations, periods, HZ tags, profiles, and the IISS
// Class II/III form.
//
// Pipeline:
//
//  1. Per placement: roll Size (terrestrial or gas giant); attach diameter/mass.
//  2. Per non-belt non-empty placement: roll moon count + per-moon sizes.
//  3. AssignPlanetDesignations + AssignMoonDesignations.
//  4. Compute Period per placement.
//  5. MarkHZ.
//  6. Backfill StarAllocation.BaselineN.
//  7. ShortProfile + LongProfile.
//  8. RenderIISSClass23.
func DetailSystem(r roller.Roller, sys stars.System, sp SystemPlacement, h IISSClass23Header) (SystemDetail, error) {
	detailed := make([]DetailedPlacement, len(sp.Placements))
	for i := range sp.Placements {
		detailed[i] = DetailedPlacement{Placement: sp.Placements[i]}
	}

	// Step 1 — sizing
	gasGiantDM := gasGiantSizingDM(sys, sp)
	for i := range detailed {
		switch detailed[i].Body {
		case BodyTerrestrial:
			ts, err := RollTerrestrialSize(r)
			if err != nil {
				return SystemDetail{}, fmt.Errorf("worlds: detail size terrestrial[%d]: %w", i, err)
			}
			detailed[i].SizeCode = ts.SizeCode
			detailed[i].DiameterKm = ts.DiameterKm
		case BodyGasGiant:
			gs, err := RollGasGiantSize(r, gasGiantDM)
			if err != nil {
				return SystemDetail{}, fmt.Errorf("worlds: detail size gas-giant[%d]: %w", i, err)
			}
			detailed[i].GGClass = gs.Class
			detailed[i].GGDiameterCode = gs.DiameterCode
			detailed[i].DiameterEarth = gs.DiameterEarth
			detailed[i].MassEarth = gs.MassEarth
		}
	}

	// Step 2 — moons (skip belts and empty)
	for i := range detailed {
		if detailed[i].Body == BodyEmpty || detailed[i].Body == BodyPlanetoidBelt {
			continue
		}
		parent := parentInfoOf(detailed[i])
		moonDM := moonCountDM(detailed[i], sp)
		count, err := CountMoons(r, parent, moonDM)
		if err != nil {
			return SystemDetail{}, fmt.Errorf("worlds: detail moon-count[%d]: %w", i, err)
		}
		moons := make([]Moon, 0, count)
		for j := 0; j < count; j++ {
			m, err := SizeMoon(r, parent)
			if err != nil {
				return SystemDetail{}, fmt.Errorf("worlds: detail moon-size[%d/%d]: %w", i, j, err)
			}
			moons = append(moons, m)
		}
		detailed[i].Moons = moons
	}

	// Step 3 — designations
	AssignPlanetDesignations(detailed)
	AssignMoonDesignations(detailed)

	// Step 4 — periods
	for i := range detailed {
		if detailed[i].Body == BodyEmpty {
			continue
		}
		au := stars.OrbitToAU(detailed[i].Orbit)
		sumMass := sumStellarMassInterior(detailed[i])
		bodyMassEarth := 0.0
		if detailed[i].Body == BodyGasGiant && detailed[i].MassEarth >= 100 {
			bodyMassEarth = detailed[i].MassEarth
		}
		detailed[i].Period = PeriodFor(au, sumMass, bodyMassEarth)
	}

	// Step 5 — HZ tagging
	if err := MarkHZ(detailed); err != nil {
		return SystemDetail{}, fmt.Errorf("worlds: detail mark-hz: %w", err)
	}

	// Step 5A — 3A1 passes: body physical, belt details, atmosphere,
	// hydrographics, moon refinement.
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
				return SystemDetail{}, fmt.Errorf("worlds: body physical %s: %w", dp.Designation, err)
			}
			dp.Physical = &bp
		}

		// Belt details (Size 0 only).
		if dp.SizeCode == "0" {
			bd, err := GenerateBeltDetails(r, dp.Orbit, sp.SystemSpread, hzco, sys.Primary.AgeGyr, false, false)
			if err != nil {
				return SystemDetail{}, fmt.Errorf("worlds: belt %s: %w", dp.Designation, err)
			}
			dp.Belt = &bd
		}

		// Atmosphere (HZ-orbit terrestrials only).
		if dp.HZ && dp.GGClass == NotGasGiant && dp.SizeCode != "0" && dp.SizeCode != "R" && dp.SizeCode != "" {
			offset := dp.Orbit - hzco
			atmoCode, err := RollAtmoCode(r, dp.SizeCode, offset)
			if err != nil {
				return SystemDetail{}, fmt.Errorf("worlds: atmosphere %s: %w", dp.Designation, err)
			}
			atmo := Atmosphere{Code: atmoCode}
			if atmoCode == 11 || atmoCode == 12 {
				st, serr := RollCorrosiveInsidiousSubtype(r, dp.SizeCode, dp.Orbit, hzco, atmoCode == 12, false)
				if serr != nil {
					return SystemDetail{}, fmt.Errorf("worlds: atmo subtype %s: %w", dp.Designation, serr)
				}
				atmo.Subtype = st
			}
			press, perr := RollTotalPressure(r, atmoCode)
			if perr != nil {
				return SystemDetail{}, fmt.Errorf("worlds: pressure %s: %w", dp.Designation, perr)
			}
			atmo.Pressure = press
			if atmoCode >= 2 && atmoCode <= 9 {
				frac, ferr := RollOxygenFraction(r, sys.Primary.AgeGyr)
				if ferr != nil {
					return SystemDetail{}, fmt.Errorf("worlds: oxygen %s: %w", dp.Designation, ferr)
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
				return SystemDetail{}, fmt.Errorf("worlds: hydro %s: %w", dp.Designation, herr)
			}
			dp.Hydrographics = &hydro
		}

		// Moon refinement (any planet with moons).
		if len(dp.Moons) > 0 {
			refinePlacementMoons(r, dp)
		}
	}

	// Step 6 — backfill StarAllocation.BaselineN
	allocs := make([]StarAllocation, len(sp.Allocations))
	copy(allocs, sp.Allocations)
	for i := range allocs {
		allocs[i].BaselineN = computeBaselineN(allocs[i].Group, detailed)
	}

	// Step 7 — profiles
	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			Counts:        sp.Counts,
			Allocations:   allocs,
			BaselineN:     sp.BaselineN,
			BaselineOrbit: sp.BaselineOrbit,
			EmptyOrbits:   sp.EmptyOrbits,
			SystemSpread:  sp.SystemSpread,
			Placements:    sp.Placements,
		},
		Detailed: detailed,
	}
	sd.ShortProfile = ShortProfile(sd)
	sd.LongProfile = LongProfile(sd)

	// Step 8 — IISS Class II/III form
	sd.Survey = RenderIISSClass23(sd, sys, h)

	return sd, nil
}

// tempRangeMidpointK returns a representative mean temperature in Kelvin
// for the given TempRange band. Used for scale-height calculation in 3A1
// before the full temperature roll lands in 3A2.
func tempRangeMidpointK(t TempRange) float64 {
	switch t {
	case TempBoiling:
		return 600
	case TempHot:
		return 400
	case TempTemperate:
		return 313
	case TempCold:
		return 200
	case TempFrozen:
		return 100
	}
	return 288
}

// refinePlacementMoons applies WBH pp.75-77 moon refinements: Hill sphere,
// moon-removal check, per-moon orbit + period. Mutates dp in place.
// No-op for bodies without resolvable mass.
func refinePlacementMoons(r roller.Roller, dp *DetailedPlacement) {
	planetMass := dp.MassEarth
	if planetMass == 0 && dp.Physical != nil {
		planetMass = DeriveMass(dp.Physical.Density, dp.DiameterKm)
	}
	if planetMass == 0 {
		return
	}
	planetDiameter := dp.DiameterKm
	if dp.GGClass != NotGasGiant && dp.DiameterEarth > 0 {
		planetDiameter = dp.DiameterEarth * DiameterTerra
	}
	sumStellarMass := sumStellarMassInterior(*dp)
	au := stars.OrbitToAU(dp.Orbit)
	_, pd := HillSphere(au, dp.Eccentricity, planetMass, sumStellarMass, planetDiameter)
	limit := HillSphereMoonLimit(pd)
	if removeAll, _ := MoonRemovalCheck(limit); removeAll {
		dp.Moons = nil
		return
	}
	mor := MoonOrbitRange(limit, len(dp.Moons))
	// MoonPeriodHours uses (PD × effectiveSize) where effectiveSize ≈ parent
	// diameter in 1600 km units. For terrestrials, parent Size code is the
	// multiplier; for gas giants, use diameterEarth × 8.
	effSize := SizeAsInt(dp.SizeCode)
	if dp.GGClass != NotGasGiant && dp.DiameterEarth > 0 {
		effSize = int(dp.DiameterEarth * 8)
	}
	for j := range dp.Moons {
		orbit, _ := RollMoonOrbit(r, mor)
		dp.Moons[j].OrbitPD = orbit
		if effSize > 0 {
			dp.Moons[j].PeriodHours = MoonPeriodHours(orbit, effSize, planetMass)
		}
	}
}

// parentInfoOf builds a ParentInfo from a DetailedPlacement.
func parentInfoOf(dp DetailedPlacement) ParentInfo {
	if dp.Body == BodyGasGiant {
		return ParentInfo{IsGasGiant: true, GGClass: dp.GGClass}
	}
	return ParentInfo{SizeCode: dp.SizeCode}
}

// gasGiantSizingDM derives the WBH p.55 Gas Giant Sizing DM.
func gasGiantSizingDM(sys stars.System, sp SystemPlacement) int {
	dm := 0
	primary := sys.Primary
	if primary.Kind == stars.KindBrownDwarf ||
		(primary.SpectralType.Letter == 'M' && primary.LuminosityClass == stars.V) ||
		primary.LuminosityClass == stars.VI {
		dm--
	}
	if sp.SystemSpread < 0.1 {
		dm--
	}
	return dm
}

// moonCountDM derives the WBH p.55 per-die moon-count DM.
// Returns -1 when any of the following apply:
//   - Planet's Orbit# < 1.0
//   - Planet's orbit is adjacent to a companion interval edge (within spread/2)
func moonCountDM(dp DetailedPlacement, sp SystemPlacement) int {
	if dp.Orbit < 1.0 {
		return -1
	}
	w := sp.SystemSpread / 2
	for _, iv := range dp.Group.Intervals {
		if math.Abs(dp.Orbit-iv.Min) < w || math.Abs(dp.Orbit-iv.Max) < w {
			return -1
		}
	}
	return 0
}

// sumStellarMassInterior returns the sum of stellar masses in the
// placement's host group per WBH p.53.
func sumStellarMassInterior(dp DetailedPlacement) float64 {
	sum := 0.0
	for _, m := range dp.Group.Members {
		sum += m.Mass
	}
	return sum
}

// computeBaselineN returns the per-star baseline number for a group
// per WBH p.58: the 1-based index of the slot closest to the group's
// own HZCO, except:
//   - if the group has no Members or no non-empty slots: return 0
//   - if all slots are outside the HZ (HZCO ± 1.0): return 0
//   - if all slots are inside the HZ: return total slot count (N = X)
func computeBaselineN(g Group, detailed []DetailedPlacement) int {
	if len(g.Members) == 0 {
		return 0
	}
	hzco := g.HZCO()

	idx := 0
	inHZCount := 0
	totalCount := 0
	bestDelta := math.Inf(1)
	bestIdx := 0
	for _, dp := range detailed {
		if dp.Group.Designation != g.Designation || dp.Body == BodyEmpty {
			continue
		}
		idx++
		totalCount++
		if dp.Orbit >= hzco-1.0 && dp.Orbit <= hzco+1.0 {
			inHZCount++
		}
		d := math.Abs(dp.Orbit - hzco)
		if d < bestDelta {
			bestDelta = d
			bestIdx = idx
		}
	}
	if totalCount == 0 {
		return 0
	}
	if inHZCount == 0 {
		return 0
	}
	if inHZCount == totalCount {
		return totalCount
	}
	return bestIdx
}

// DetailedPlacement extends 2B's Placement with the WBH pp. 53-67
// per-body data (Size, moons, period, HZ flag, designation).
//
// Embeds Placement, continuing the existing chain:
//
//	Slot → AnomalousSlot → Placement → DetailedPlacement
//
// 2B types are unchanged.
type DetailedPlacement struct {
	Placement // 2B fields: Body, PrefixRoll, Eccentricity, AnomalousSlot, Slot

	// Terrestrial fields — set when Body == BodyTerrestrial.
	SizeCode   SizeCode
	DiameterKm float64

	// Gas-giant fields — set when Body == BodyGasGiant.
	GGClass        GasGiantClass
	GGDiameterCode string
	DiameterEarth  float64
	MassEarth      float64

	// All non-empty bodies:
	Designation string // "Aab I", "Aab PI" — assigned by AssignPlanetDesignations
	Period      Period
	HZ          bool // within HZCO ± 1.0 — set by MarkHZ
	Moons       []Moon

	// 3A1 additions — pointer = nil means "not applicable to this body type"
	Physical      *BodyPhysical
	Belt          *BeltDetails
	Atmosphere    *Atmosphere
	Hydrographics *Hydrographics

	// 3A2a additions
	DayLength *DayLength
	AxialTilt *AxialTilt
}

// HasPhysical reports whether body-physical data has been generated for this placement.
func (dp *DetailedPlacement) HasPhysical() bool { return dp.Physical != nil }

// HasAtmosphere reports whether atmosphere data has been generated for this placement.
func (dp *DetailedPlacement) HasAtmosphere() bool { return dp.Atmosphere != nil }

// HasHydrographics reports whether hydrographics data has been generated for this placement.
func (dp *DetailedPlacement) HasHydrographics() bool { return dp.Hydrographics != nil }

// RenderSAH returns the 3-character SAH triplet for the IISS form.
// HZ bodies get the full triplet; non-HZ bodies render as "<Size>??".
func (dp *DetailedPlacement) RenderSAH() string {
	size := string(dp.SizeCode)
	if size == "" {
		size = "?"
	}
	if !dp.HasAtmosphere() || !dp.HasHydrographics() {
		return size + "??"
	}
	atmoChar := atmosphereCodeChar(dp.Atmosphere.Code)
	hydroChar := fmt.Sprintf("%d", dp.Hydrographics.Code)
	if dp.Hydrographics.Code == 10 {
		hydroChar = "A"
	}
	return size + atmoChar + hydroChar
}

// SystemDetail is the DetailSystem façade output, layered atop 2B's
// SystemPlacement.
type SystemDetail struct {
	SystemPlacement // 2B: Counts, Allocations, BaselineN, BaselineOrbit, EmptyOrbits, SystemSpread, Placements

	// Detailed mirrors SystemPlacement.Placements 1:1, with 2C per-body
	// detail attached. Ordered by ascending orbit within each group,
	// matching SystemPlacement.Placements (which itself follows
	// PlaceOrbitSlots' ascending-orbit walk). LongProfile and
	// AssignPlanetDesignations both rely on this ordering — the T14
	// DetailSystem façade must preserve it when building Detailed.
	Detailed []DetailedPlacement

	ShortProfile string          // "G-P-T-N-S" form per WBH p.58
	LongProfile  string          // "St-N-W-W-S:..." form per WBH p.58
	Survey       IISSClass23Form // IISS Class II/III survey form (Task 13)
}
