package worlds

import (
	"wbh/roller"
	"wbh/stars"
)

// RollPlanetEccentricities implements WBH Step 9 (p. 52).
//
// For each non-empty non-belt placement, calls stars.RollEccentricity
// with the placement's anomaly DM (AnomalousSlot.EccentricityDM) passed
// through stars.EccentricityOpts.ExtraDM. Stores the result on
// Placement.Eccentricity.
//
// Belts and empty slots are skipped (no roll consumed).
//
// Trojan slots have EccentricityDM == 0 by Task 12 design and inherit
// their eccentricity from the slot they shadow; for 2B we still call
// the eccentricity roll (since the 2B procedure does not actually look
// up the parent slot's eccentricity here). 2C should special-case
// Trojan to copy the parent's eccentricity instead.
func RollPlanetEccentricities(r roller.Roller, ps []Placement) ([]Placement, error) {
	out := make([]Placement, len(ps))
	copy(out, ps)
	for i := range out {
		if out[i].Body == BodyEmpty || out[i].Body == BodyPlanetoidBelt {
			continue
		}
		ecc, err := stars.RollEccentricity(r, stars.EccentricityOpts{
			ExtraDM: out[i].EccentricityDM,
			// We don't have the system age plumbed through to this layer
			// in 2B; the sub-1.0/age>1Gyr DM is therefore not applied
			// here. 2C should plumb sys.Primary.AgeGyr through.
		})
		if err != nil {
			return nil, err
		}
		out[i].Eccentricity = ecc
	}
	return out, nil
}
