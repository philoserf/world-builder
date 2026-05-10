package worlds

import (
	"fmt"
	"math"

	"wbh/roller"
	"wbh/stars"
)

// DetailOpts gates opt-in WBH rules and Referee-discretion variants for
// the per-body detail pipeline. The zero value disables every opt — so
// DetailSystem(...) (the no-opts wrapper) preserves canonical-book
// behavior for all existing callers.
type DetailOpts struct {
	// OxygenAtmBiomassFloor enables WBH p.128 Optional Rule: any world
	// whose Atmosphere.Code is in the oxygen-bearing set {2-9, D, E}
	// gets a biomass floor of 1 (the rolled value is clamped up if it
	// came in below). Off by default — the book describes it as a
	// Referee opt-in.
	OxygenAtmBiomassFloor bool
}

// DetailSystem composes the full WBH pp. 53-67 procedure on top of a
// SystemPlacement (2B output). Returns a SystemDetail with sizes,
// moons, designations, periods, HZ tags, profiles, and the IISS
// Class II/III form.
//
// Equivalent to DetailSystemWithOpts with a zero-valued DetailOpts —
// canonical-book behavior, no opt-in rules.
//
// Pipeline:
//
//  1. runDetailPipeline: per-body Steps 1-5 + 5A-5G (sizing, moons,
//     designations, periods, HZ, then 3A1/3A2/3B passes).
//  2. Backfill StarAllocation.BaselineN.
//  3. ShortProfile + LongProfile.
//  4. RenderIISSClass23.
//  5. pickMainworld.
func DetailSystem(r roller.Roller, sys stars.System, sp SystemPlacement, h IISSClass23Header) (SystemDetail, error) {
	return DetailSystemWithOpts(r, sys, sp, h, DetailOpts{})
}

// DetailSystemWithOpts is the opt-aware variant of DetailSystem. See
// DetailOpts for the available opt-in rules.
func DetailSystemWithOpts(r roller.Roller, sys stars.System, sp SystemPlacement, h IISSClass23Header, opts DetailOpts) (SystemDetail, error) {
	detailed := make([]DetailedPlacement, len(sp.Placements))
	for i := range sp.Placements {
		detailed[i] = DetailedPlacement{Placement: sp.Placements[i]}
	}

	if err := runDetailPipeline(r, detailed, sys, sp, opts); err != nil {
		return SystemDetail{}, err
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

	sd.MainworldDesignation = pickMainworld(detailed)

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

// buildMoonPlacementView constructs a DetailedPlacement-shaped view of a moon
// for the purpose of feeding it into 3A2a's Generate* functions. The moon's
// orbital fields (Eccentricity, etc.) come from the Moon struct; star-relative
// fields (Orbit, HZ) are inherited from the parent placement.
func buildMoonPlacementView(m *Moon, parent *DetailedPlacement) *DetailedPlacement {
	dp := &DetailedPlacement{
		SizeCode:      m.SizeCode,
		DiameterKm:    m.DiameterKm,
		GGClass:       m.GGClass,
		MassEarth:     m.MassEarth,
		Designation:   m.Designation,
		Period:        Period{Hours: m.PeriodHours},
		HZ:            parent.HZ,
		Atmosphere:    m.Atmosphere,
		Hydrographics: m.Hydrographics,
		Physical:      m.Physical,
	}
	dp.Eccentricity = m.Eccentricity
	dp.AxialTilt = m.AxialTilt
	dp.DayLength = m.DayLength
	dp.Body = BodyTerrestrial
	if m.GGClass != NotGasGiant {
		dp.Body = BodyGasGiant
	}
	dp.Orbit = parent.Orbit
	dp.Group = parent.Group
	return dp
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
	DayLength           *DayLength
	AxialTilt           *AxialTilt
	SurfaceDistribution *SurfaceDistribution
	TidalLock           *TidalLock
	TidalEffects        *SurfaceTidalEffects

	// 3A2b-temp additions
	Temperature *Temperature

	// 3B-geology additions
	Geology *Geology

	// 3B-biology additions
	Biology *Biology

	// 3B-final additions
	Habitability *Habitability
}

// HasPhysical reports whether body-physical data has been generated for this placement.
func (dp *DetailedPlacement) HasPhysical() bool { return dp.Physical != nil }

// HasAtmosphere reports whether atmosphere data has been generated for this placement.
func (dp *DetailedPlacement) HasAtmosphere() bool { return dp.Atmosphere != nil }

// HasHydrographics reports whether hydrographics data has been generated for this placement.
func (dp *DetailedPlacement) HasHydrographics() bool { return dp.Hydrographics != nil }

// HasDayLength reports whether day-length data has been generated for this placement.
func (dp *DetailedPlacement) HasDayLength() bool { return dp.DayLength != nil }

// HasAxialTilt reports whether axial-tilt data has been generated for this placement.
func (dp *DetailedPlacement) HasAxialTilt() bool { return dp.AxialTilt != nil }

// HasSurfaceDistribution reports whether surface-distribution data has been generated.
func (dp *DetailedPlacement) HasSurfaceDistribution() bool { return dp.SurfaceDistribution != nil }

// HasTidalLock reports whether tidal-lock data has been generated for this placement.
func (dp *DetailedPlacement) HasTidalLock() bool { return dp.TidalLock != nil }

// HasTidalEffects reports whether surface tidal-effects data has been generated.
func (dp *DetailedPlacement) HasTidalEffects() bool { return dp.TidalEffects != nil }

// HasTemperature reports whether 5C ran for this placement.
func (dp *DetailedPlacement) HasTemperature() bool { return dp.Temperature != nil }

// HasGeology reports whether 5E ran for this placement.
func (dp *DetailedPlacement) HasGeology() bool { return dp.Geology != nil }

// HasBiology reports whether biology data has been generated for this placement.
func (dp *DetailedPlacement) HasBiology() bool { return dp.Biology != nil }

// HasHabitability reports whether habitability data has been generated for this placement.
func (dp *DetailedPlacement) HasHabitability() bool { return dp.Habitability != nil }

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

	// MainworldDesignation is the auto-picked mainworld's designation per
	// WBH p.134. Priority chain: bodies with native sophonts → highest
	// habitability → highest resource → first in iteration order.
	// Empty string if no terrestrial body qualifies.
	//
	// The book explicitly says the Referee may override this pick. A future
	// sub-project may add a Referee-override mechanism; for now the
	// auto-pick is the only source.
	MainworldDesignation string
}
