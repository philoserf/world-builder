package worlds

import (
	"wbh/roller"
	"wbh/stars"
)

// runStep5G applies the 3B-final habitability pass: per-body Habitability
// rating per WBH p.132. Mutates detailed in place.
//
// Body filter: terrestrials only (skip belts/GGs/empty). Atmosphere
// optional — vacuum worlds (nil atm) get DM-8 per the atm 0 row. Differs
// from Biology's filter (which required atm).
//
// Per-body dice budget: 0 dice. Habitability is deterministic. Same per moon.
//
//nolint:unparam // matches sibling runStep5* signatures (always-nil error; r unused)
func runStep5G(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error {
	_ = r   // unused — Habitability is deterministic
	_ = sys // unused
	for i := range detailed {
		dp := &detailed[i]
		if !habitabilityApplies(dp) {
			continue
		}
		h := ComputeHabitability(dp)
		dp.Habitability = &h

		for j := range dp.Moons {
			m := &dp.Moons[j]
			moonDP := buildMoonPlacementView(m, dp)
			// buildMoonPlacementView does not copy Temperature or TidalLock;
			// alias them so ComputeHabitability sees the moon's own values.
			// (Same pattern as 3B-biology Task 9's fix for moon biomass DMs.)
			moonDP.Temperature = m.Temperature
			moonDP.TidalLock = m.TidalLock
			if !habitabilityApplies(moonDP) {
				continue
			}
			mh := ComputeHabitability(moonDP)
			m.Habitability = &mh
		}
	}
	return nil
}

// habitabilityApplies reports whether dp should receive a Habitability struct.
// True for terrestrial bodies (atmosphere optional — vacuum worlds get a rating).
// False for empty, belts, gas giants.
func habitabilityApplies(dp *DetailedPlacement) bool {
	if dp == nil || dp.Body == BodyEmpty {
		return false
	}
	if dp.Body == BodyGasGiant || dp.Body == BodyPlanetoidBelt {
		return false
	}
	if dp.SizeCode == "0" {
		return false
	}
	return true
}
