package worlds

import (
	"wbh/roller"
	"wbh/stars"
)

// RollPlanetEccentricities implements WBH Step 9 (p. 52).
//
// For each non-empty non-belt placement, calls stars.RollEccentricity
// with the placement's Orbit and the system's ageGyr passed through
// stars.EccentricityOpts so the WBH p.27 sub-1.0/age>1Gyr DM-1 applies.
// The placement's anomaly DM (AnomalousSlot.EccentricityDM) is also
// forwarded via stars.EccentricityOpts.ExtraDM. Stores the result on
// Placement.Eccentricity.
//
// Pass ageGyr=0 (or any value ≤1) to suppress the age DM.
//
// Belts and empty slots are skipped (no roll consumed).
//
// Trojan slots (Anomaly == AnomalyTrojan) are handled specially per
// WBH p. 51: they inherit the orbit, eccentricity, and inclination of
// the slot they shadow. We do not roll a fresh eccentricity for them;
// we copy from the slot whose StarSlot matches TrojanOf in a second pass.
func RollPlanetEccentricities(r roller.Roller, ps []Placement, ageGyr float64) ([]Placement, error) {
	out := make([]Placement, len(ps))
	copy(out, ps)
	// First pass: roll for non-Trojan, non-empty, non-belt placements.
	for i := range out {
		if out[i].Body == BodyEmpty || out[i].Body == BodyPlanetoidBelt {
			continue
		}
		if out[i].Anomaly == AnomalyTrojan {
			continue // handled in pass 2
		}
		ecc, err := stars.RollEccentricity(r, stars.EccentricityOpts{
			ExtraDM:      out[i].EccentricityDM,
			NestingDepth: nestingDepthFor(out[i]),
			Orbit:        out[i].Orbit,
			SystemAgeGyr: ageGyr,
		})
		if err != nil {
			return nil, err
		}
		out[i].Eccentricity = ecc
	}
	// Second pass: Trojans inherit from their TrojanOf parent.
	for i := range out {
		if out[i].Anomaly != AnomalyTrojan {
			continue
		}
		for j := range out {
			if out[j].StarSlot == out[i].TrojanOf {
				out[i].Eccentricity = out[j].Eccentricity
				break
			}
		}
	}
	return out, nil
}

// nestingDepthFor returns the WBH p. 27 NestingDepth for a placement.
// Planets in the primary group orbit at depth 0; planets in a Close,
// Near, or Far secondary group orbit the secondary which orbits the
// primary, so depth 1.
func nestingDepthFor(p Placement) int {
	if p.Group.sourceCompanion == nil {
		return 0 // primary group
	}
	return 1 // secondary group
}
