package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

// ApplyBodyPhysical generates per-body physical characteristics
// (composition / density / gravity / mass) for terrestrial planets and
// moons. Stage-3 orchestrator entry per WBH pp.69-78. Mutates Body.Physical
// and Body.MassEarth in place.
//
// Per anti-pattern A.1, every terrestrial moon is walked alongside its
// planet via AllBodies — body.Host() supplies the parent for moons and
// the body itself for planets so HZ inheritance flows uniformly.
// GG-cascade moons (Body.SizeCode == "G", GGClass != NotGasGiant)
// already have MassEarth from Stage 2's gasGiantSpecialMoon and skip
// physical generation. Belts (Size 0) handled by ApplyBeltDetails.
func ApplyBodyPhysical(r roller.Roller, u *Universe) error {
	for body := range u.AllBodies() {
		if err := generateBodyPhysicalIfTerrestrial(r, body, body.Host(), u.System); err != nil {
			return err
		}
	}
	return nil
}

// generateBodyPhysicalIfTerrestrial runs the body-physical procedure for
// a terrestrial body (or moon). hostForHZ supplies the orbit and HZ flag
// — moons inherit them from the parent. No-op for non-terrestrial bodies.
func generateBodyPhysicalIfTerrestrial(r roller.Roller, body, hostForHZ *Body, sys stars.System) error {
	if body.GGClass != NotGasGiant {
		return nil
	}
	switch body.SizeCode {
	case "", "0", "R":
		return nil
	}
	hzco := hostForHZ.Group.HZCO()
	beyondHZCO := 0
	if offset := hostForHZ.Orbit - hzco; offset > 0 {
		beyondHZCO = int(offset)
	}
	dms := BodyPhysicalDMs{
		SizeCode:       body.SizeCode,
		AtHZCOOrCloser: hostForHZ.HZ,
		BeyondHZCO:     beyondHZCO,
		SystemAgeGyr:   sys.Primary.AgeGyr,
	}
	bp, err := GenerateBodyPhysical(r, body.SizeCode, int(body.DiameterKm), dms)
	if err != nil {
		return fmt.Errorf("worlds: stage3 body physical %s: %w", body.Designation, err)
	}
	body.Physical = &bp
	body.MassEarth = DeriveMass(bp.Density, body.DiameterKm)
	return nil
}

// ApplyBeltDetails generates per-belt details (span, composition,
// bulk, resource rating, significant-size body counts) for every belt
// body. Stage-3 orchestrator entry per WBH pp.91-93.
//
// Gates on Kind, not SizeCode: belts carry SizeCode "" per the SizeCode
// contract (sizing_terrestrial.go); the historical `SizeCode == "0"`
// gate matched no body, so belt details were never generated.
func ApplyBeltDetails(r roller.Roller, u *Universe) error {
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if body.Kind != BodyPlanetoidBelt {
			continue
		}
		adjacentToGG, outermostSlot := beltSpanFlags(u.Detail.Bodies, i)
		hzco := body.Group.HZCO()
		bd, err := GenerateBeltDetails(r, body.Orbit, u.Placement.SystemSpread, hzco, u.System.Primary.AgeGyr, adjacentToGG, outermostSlot)
		if err != nil {
			return fmt.Errorf("worlds: stage3 belt %s: %w", body.Designation, err)
		}
		body.Belt = &bd
	}
	return nil
}

// beltSpanFlags computes the WBH p.72 belt-span DM inputs for the belt
// at bodies[i]:
//
//   - adjacentToGG: the directly neighboring orbit slot on either side
//     (within the same star group, empty slots included as slots) holds
//     a gas giant. A gas giant beyond an intervening empty orbit is not
//     adjacent.
//   - outermostSlot: no non-empty body of the same group sits at a
//     higher orbit. Trailing empty slots do not disqualify — the DM's
//     intent is "nothing beyond the belt", and an empty slot is not a
//     body. (Book unavailable to confirm slot-vs-body wording; this
//     interpretation is asserted by TestBeltSpanFlags.)
//
// Neighbors are found by orbit value within the belt's group (nearest
// slot below / above by Orbit), not by slice position, so the result
// does not depend on Detail.Bodies' storage layout.
func beltSpanFlags(bodies []Body, i int) (adjacentToGG, outermostSlot bool) {
	belt := &bodies[i]
	var inner, outer *Body // nearest same-group slots by orbit
	outermostSlot = true
	for j := range bodies {
		b := &bodies[j]
		if j == i || b.Group.Designation != belt.Group.Designation {
			continue
		}
		switch {
		case b.Orbit < belt.Orbit:
			if inner == nil || b.Orbit > inner.Orbit {
				inner = b
			}
		case b.Orbit > belt.Orbit:
			if outer == nil || b.Orbit < outer.Orbit {
				outer = b
			}
			if b.Kind != BodyEmpty {
				outermostSlot = false
			}
		}
	}
	adjacentToGG = (inner != nil && inner.Kind == BodyGasGiant) ||
		(outer != nil && outer.Kind == BodyGasGiant)
	return adjacentToGG, outermostSlot
}

// ApplyMoonRefinement runs the WBH pp.75-77 moon refinement pass:
// Hill-sphere computation, moon-removal check, per-moon orbit + period.
// Mutates each parent's Children slice in place.
func ApplyMoonRefinement(r roller.Roller, u *Universe) error {
	for i := range u.Detail.Bodies {
		parent := &u.Detail.Bodies[i]
		if len(parent.Children) == 0 {
			continue
		}
		refineParentMoons(r, parent)
	}
	return nil
}

// refineParentMoons applies WBH pp.75-77 to one parent body. No-op
// for parents without resolvable mass.
func refineParentMoons(r roller.Roller, parent *Body) {
	planetMass := parent.MassOrDerived()
	if planetMass == 0 {
		return
	}
	planetDiameter := parent.DiameterKm
	if parent.GGClass != NotGasGiant && parent.DiameterEarth > 0 {
		planetDiameter = parent.DiameterEarth * DiameterTerra
	}
	sumStellarMass := sumStellarMassInterior(parent)
	au := stars.OrbitToAU(parent.Orbit)
	_, pd := HillSphere(au, parent.Eccentricity, planetMass, sumStellarMass, planetDiameter)
	limit := HillSphereMoonLimit(pd)
	if MoonRemovalCheck(limit) {
		// p.76: all significant moons removed, first promoted to a ring.
		parent.Children = nil
		parent.Ring = true
		return
	}
	mor := MoonOrbitRange(limit, len(parent.Children))
	// Effective parent size for moon period: terrestrials use Size code;
	// gas giants use diameter ratio (DiameterEarth × 8 → Earth-units).
	effSize := SizeAsInt(parent.SizeCode)
	if parent.GGClass != NotGasGiant && parent.DiameterEarth > 0 {
		effSize = int(parent.DiameterEarth * 8)
	}
	for j := range parent.Children {
		orbit, mr := RollMoonOrbit(r, mor)
		parent.Children[j].OrbitPD = orbit
		parent.Children[j].OrbitKm = orbit * planetDiameter
		parent.Children[j].Retrograde = RollMoonRetrograde(r, mr, orbit > float64(mor))
		if effSize > 0 {
			parent.Children[j].PeriodHours = MoonPeriodHours(orbit, effSize, planetMass)
		}
	}
}
