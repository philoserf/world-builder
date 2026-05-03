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
