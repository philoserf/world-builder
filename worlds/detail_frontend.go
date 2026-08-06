package worlds

import (
	"fmt"
	"math"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

// placementToBody constructs a Body from a Stage-1 Placement. Group,
// Orbit, and Eccentricity are copied via Placement's Slot embedding.
// Kind comes from Placement.Body. Stage-2+ fields are zero-valued and
// land cycle-by-cycle.
func placementToBody(p Placement) Body {
	return Body{
		Kind:         p.Body,
		Group:        p.Group,
		Orbit:        p.Orbit,
		Eccentricity: p.Eccentricity,
	}
}

// parentInfoOf extracts ParentInfo from a Body for the moon-counting
// and moon-sizing procedures.
func parentInfoOf(body *Body) ParentInfo {
	if body.Kind == BodyGasGiant {
		return ParentInfo{IsGasGiant: true, GGClass: body.GGClass}
	}

	return ParentInfo{SizeCode: body.SizeCode}
}

// gasGiantSizingDM derives the WBH p.55 Gas Giant Sizing DM from
// primary class and system spread.
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

// moonCountDM derives the WBH p.55 per-die moon-count DM. Returns -1
// when any of the following apply:
//   - Planet's Orbit# < 1.0
//   - Planet's orbit is adjacent to a companion interval edge (within spread/2)
func moonCountDM(body *Body, sp SystemPlacement) int {
	if body.Orbit < 1.0 {
		return -1
	}

	w := sp.SystemSpread / 2
	for _, iv := range body.Group.Intervals {
		if math.Abs(body.Orbit-iv.Min) < w || math.Abs(body.Orbit-iv.Max) < w {
			return -1
		}
	}

	return 0
}

// sumStellarMassInterior returns the sum of stellar masses in the
// body's host group per WBH p.53. Used by Kepler's third law via
// PeriodFor when computing a body's orbital period.
func sumStellarMassInterior(body *Body) float64 {
	sum := 0.0
	for _, m := range body.Group.Members {
		sum += m.Mass
	}

	return sum
}

// ApplyDetailFrontEnd populates SystemDetail.Bodies with sized,
// designated, periodic, HZ-tagged bodies from the universe's
// SystemPlacement. Stage 2 of the dependency graph.
//
// Mutates u.Detail.Bodies in place. Per docs/api-surface.md §
// Mutability — the pipeline is mutator-shaped.
//
// Sub-stages, in order:
//
//  1. placementToBody — build []Body from sp.Placements.
//  2. Sizing — RollTerrestrialSize / RollGasGiantSize per body kind.
//  3. Moons — CountMoons + SizeMoon per non-belt non-empty body;
//     attach moons as Children with Parent pointer set.
//  4. Designations — AssignPlanetDesignations + AssignMoonDesignations.
//  5. Periods — PeriodFor per body via Kepler's third law.
//  6. HZ tagging — MarkHZ flags bodies in HZCO ± 1.0.
//
// Per anti-pattern A.1 (moon-path silent-zero), every moon is walked
// alongside its planet via the Children slice. There is no separate
// moon code path.
func ApplyDetailFrontEnd(r roller.Roller, u *Universe) error {
	sp := u.Placement

	// Sub-stage 1: placement → body.
	bodies := make([]Body, len(sp.Placements))
	for i := range sp.Placements {
		bodies[i] = placementToBody(sp.Placements[i])
	}

	// Sub-stage 2: sizing.
	ggDM := gasGiantSizingDM(u.System, sp)

	for i := range bodies {
		switch bodies[i].Kind {
		case BodyTerrestrial:
			ts, err := RollTerrestrialSize(r)
			if err != nil {
				return fmt.Errorf("worlds: stage2 size terrestrial[%d]: %w", i, err)
			}

			bodies[i].SizeCode = ts.SizeCode
			bodies[i].DiameterKm = ts.DiameterKm
		case BodyGasGiant:
			gs, err := RollGasGiantSize(r, ggDM)
			if err != nil {
				return fmt.Errorf("worlds: stage2 size gas-giant[%d]: %w", i, err)
			}

			bodies[i].GGClass = gs.Class
			bodies[i].GGDiameterCode = gs.DiameterCode
			bodies[i].DiameterEarth = gs.DiameterEarth
			bodies[i].MassEarth = gs.MassEarth
		}
	}

	// Sub-stage 3: moons.
	for i := range bodies {
		if bodies[i].Kind == BodyEmpty || bodies[i].Kind == BodyPlanetoidBelt {
			continue
		}

		parent := parentInfoOf(&bodies[i])
		dm := moonCountDM(&bodies[i], sp)

		count, ring, err := CountMoons(r, parent, dm)
		if err != nil {
			return fmt.Errorf("worlds: stage2 moon-count[%d]: %w", i, err)
		}

		if ring {
			bodies[i].Ring = true
		}

		if count == 0 {
			continue
		}

		children := make([]*Body, 0, count)
		for j := range count {
			m, err := SizeMoon(r, parent)
			if err != nil {
				return fmt.Errorf("worlds: stage2 moon-size[%d/%d]: %w", i, j, err)
			}

			child := m // local copy; take pointer below
			child.Parent = &bodies[i]
			children = append(children, &child)
		}

		bodies[i].Children = children
	}

	// Sub-stage 4: designations.
	AssignPlanetDesignations(bodies)
	AssignMoonDesignations(bodies)

	// Sub-stage 5: periods.
	for i := range bodies {
		if bodies[i].Kind == BodyEmpty {
			continue
		}

		au := stars.OrbitToAU(bodies[i].Orbit)
		sumMass := sumStellarMassInterior(&bodies[i])

		bodyMassEarth := 0.0
		if bodies[i].Kind == BodyGasGiant && bodies[i].MassEarth >= 100 {
			bodyMassEarth = bodies[i].MassEarth
		}

		bodies[i].Period = PeriodFor(au, sumMass, bodyMassEarth)
	}

	// Sub-stage 6: HZ tagging.
	MarkHZ(bodies)

	u.Detail.Bodies = bodies

	return nil
}
