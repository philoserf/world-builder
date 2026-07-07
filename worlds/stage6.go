package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
)

// ApplyTaintTypology applies the atmosphere taint typology pass per
// WBH pp.81-90 to every body with an atmosphere. Stage-6 orchestrator.
//
// Sub-stages per body:
//
//  1. Pre-existing oxygen taint promotion (atm 5/6/8 → 4/7/9 based on
//     OxygenPartialPressure).
//  2. Multi-roll Taint Subtype (max 3, reroll-on-10), with severity +
//     persistence rolls.
//  3. Insidious Atmosphere Hazard typology for atm C (Insidious).
//
// Per anti-pattern A.1, every moon is walked alongside its parent —
// AllBodies descends into Children automatically.
func ApplyTaintTypology(r roller.Roller, u *Universe) error {
	for body, parent := range u.AllBodiesWithParent() {
		// Guard before forking, mirroring ApplySurfaceDistribution and
		// ApplyBiology: applyBodyTaints is a no-op without an atmosphere,
		// so skip constructing an unused sub-roller. Behavior-preserving —
		// eligible bodies key the same "taint" sub-stream regardless.
		if !body.HasAtmosphere() {
			continue
		}
		applyBodyTaints(bodySub(r, body, parent, "taint"), body)
	}
	return nil
}

// applyBodyTaints runs the taint pipeline for one body. No-op when the
// body has no Atmosphere. Clears stale taints/hazards before rolling.
func applyBodyTaints(r roller.Roller, body *Body) {
	if body == nil || body.Atmosphere == nil {
		return
	}
	atm := body.Atmosphere
	atm.Taints = nil
	atm.InsidiousHazards = nil

	newCode, preseed := PromoteOxygenTaint(atm.Code, atm.OxygenPartialPressure)
	atm.Code = newCode

	atm.Taints = RollAllTaints(r, body, preseed)

	if atm.Code == 12 {
		isExtremelyDense := isExtremelyDenseSubtype(atm.Subtype)
		atm.InsidiousHazards = RollInsidiousHazards(r, atm.Subtype, isExtremelyDense)
	}
}

// ApplySurfaceDistribution computes the hydrosphere distribution
// across surface zones for every body that has Hydrographics. Stage-6
// orchestrator. Runs after ApplyClimate so hydrographics is final
// (per dependency-graph.md § Stage 6).
//
// Per anti-pattern A.1, every moon is walked alongside its parent —
// AllBodies descends into Children automatically.
func ApplySurfaceDistribution(r roller.Roller, u *Universe) error {
	for body, parent := range u.AllBodiesWithParent() {
		if !body.HasHydrographics() {
			continue
		}
		sd, err := GenerateSurfaceDistribution(bodySub(r, body, parent, "surface"), body.Hydrographics)
		if err != nil {
			return fmt.Errorf("worlds: stage6 surface distribution %s: %w", body.Designation, err)
		}
		body.SurfaceDistribution = sd
	}
	return nil
}
