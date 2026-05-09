package worlds

import (
	"wbh/roller"
	"wbh/stars"
)

// runStep5DPrime applies the atmosphere taint typology pass per WBH
// pp.81-90:
//   - Pre-existing oxygen taint promotion (atm 5/6/8 → 4/7/9)
//   - Multi-roll Taint Subtype (max 3, with reroll on result 10)
//   - Severity + Persistence per taint
//   - ppO2 adjustment when fresh L/H rolled
//   - Insidious Atmosphere Hazard for atm C
//
// Slots between Step 5D (3A2b-rederive) and Step 5E (3B-geology). Visits
// every body and every moon. Mutates Atmosphere fields in place.
//
//nolint:unparam // error return kept for pipeline consistency with runStep5A-5G.
func runStep5DPrime(r roller.Roller, detailed []DetailedPlacement, _ stars.System) error {
	for i := range detailed {
		dp := &detailed[i]
		computeBodyTaints(r, dp)
		for j := range dp.Moons {
			m := &dp.Moons[j]
			moonDP := buildMoonPlacementView(m, dp)
			computeBodyTaints(r, moonDP)
		}
	}
	return nil
}

// computeBodyTaints applies the full taint pipeline to a body. Returns
// without rolling when body has no Atmosphere.
//
// Clears any stale Taints / InsidiousHazard before rolling so that
// re-running the pipeline (or running it on a body that previously held
// data from a different code) cannot leak old entries.
func computeBodyTaints(r roller.Roller, dp *DetailedPlacement) {
	if dp == nil || dp.Atmosphere == nil {
		return
	}
	atm := dp.Atmosphere
	atm.Taints = nil
	atm.InsidiousHazard = nil

	// Step 1: Promotion (atm 5/6/8 → 4/7/9 based on ppO2).
	newCode, preseed := PromoteOxygenTaint(atm.Code, atm.OxygenPartialPressure)
	atm.Code = newCode

	// Step 2: Multi-roll taints. RollAllTaints gates on eligible codes
	// (2/4/7/9, A/B/C) and fills severity/persistence for the preseeded
	// slot too. Returns nil for any other code.
	atm.Taints = RollAllTaints(r, dp, preseed)

	// Step 3: Insidious hazard for atm C only.
	//
	// isExtremelyDense currently never fires: insidious atms model
	// pressure as "Varies" (AtmospherePressureRange returns 0/0), so
	// RollTotalPressure leaves atm.Pressure == 0 and the threshold is
	// unreachable. The book's "extremely dense" criterion likely maps
	// to one of the high subtype letters, but the WBH text doesn't
	// pin down which — see the follow-up issue tracking that mapping.
	if atm.Code == 12 {
		isExtremelyDense := atm.Pressure >= 10.0
		hazardCode := RollInsidiousHazard(r, isExtremelyDense)
		atm.InsidiousHazard = &Hazard{Code: hazardCode}
	}
}
