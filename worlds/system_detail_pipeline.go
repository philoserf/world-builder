package worlds

import (
	"fmt"

	"wbh/roller"
	"wbh/stars"
)

// runDetailPipeline runs the per-body passes that populate `detailed`
// in place: Steps 1-5 (sizing, moons, designations, periods, HZ) and
// Steps 5A-5G (the 3A1, 3A2a, 3A2b-temp, 3A2b-rederive, 3B-geology,
// 3B-biology, 3B-final passes). Caller owns construction of the
// detailed slice and post-pipeline assembly (baseline, profiles, IISS
// form, mainworld pick).
func runDetailPipeline(r roller.Roller, detailed []DetailedPlacement, sys stars.System, sp SystemPlacement) error {
	// Step 1 — sizing
	gasGiantDM := gasGiantSizingDM(sys, sp)
	for i := range detailed {
		switch detailed[i].Body {
		case BodyTerrestrial:
			ts, err := RollTerrestrialSize(r)
			if err != nil {
				return fmt.Errorf("worlds: detail size terrestrial[%d]: %w", i, err)
			}
			detailed[i].SizeCode = ts.SizeCode
			detailed[i].DiameterKm = ts.DiameterKm
		case BodyGasGiant:
			gs, err := RollGasGiantSize(r, gasGiantDM)
			if err != nil {
				return fmt.Errorf("worlds: detail size gas-giant[%d]: %w", i, err)
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
			return fmt.Errorf("worlds: detail moon-count[%d]: %w", i, err)
		}
		moons := make([]Moon, 0, count)
		for j := range count {
			m, err := SizeMoon(r, parent)
			if err != nil {
				return fmt.Errorf("worlds: detail moon-size[%d/%d]: %w", i, j, err)
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
		return fmt.Errorf("worlds: detail mark-hz: %w", err)
	}

	// Step 5A — 3A1 passes (body physical, belt, atmosphere, hydrographics, moon refinement).
	if err := runStep5A(r, detailed, sys, sp); err != nil {
		return err
	}

	// Step 5B — 3A2a passes (surface distribution, day length, axial tilt, tidal lock, surface tidal effects).
	if err := runStep5B(r, detailed, sys); err != nil {
		return err
	}

	// Step 5C — 3A2b-temp temperature pass.
	if err := runStep5C(r, detailed, sys); err != nil {
		return err
	}

	// Step 5D — 3A2b-rederive 2-pass iteration.
	if err := runStep5D(r, detailed, sys); err != nil {
		return err
	}

	// Step 5E — 3B-geology pass: seismic + GG residual heat + temp recompute + tectonic plates.
	if err := runStep5E(r, detailed, sys); err != nil {
		return err
	}

	// Step 5F — 3B-biology pass: native lifeform ratings + resource rating.
	if err := runStep5F(r, detailed, sys); err != nil {
		return err
	}

	// Step 5G — 3B-final pass: per-body habitability rating.
	if err := runStep5G(r, detailed, sys); err != nil {
		return err
	}

	return nil
}
