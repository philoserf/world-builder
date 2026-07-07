package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

// SystemPlacement is the full audit trail produced by
// GenerateSystemPlacement. Each named field corresponds to one Step in
// the WBH placement procedure.
type SystemPlacement struct {
	Counts        Counts
	Allocations   []StarAllocation
	BaselineN     int
	BaselineOrbit float64
	EmptyOrbits   int
	SystemSpread  float64
	Placements    []Placement
}

// GenerateSystemPlacement runs the WBH 9-step pipeline end-to-end:
//
//  0. GenerateCounts (pp. 36-38)
//  1. AllocateOrbitsByStar (pp. 43-44)
//  2. RollBaselineNumber (pp. 44-45)
//  3. BaselineOrbit (pp. 45-46)
//  4. RollEmptyOrbits (p. 48)
//  5. Spread (pp. 48-49)
//  6. PlaceOrbitSlots (pp. 49-50)
//  7. AddAnomalous (pp. 50-51)
//  8. PlaceWorlds (pp. 51-52)
//  9. RollPlanetEccentricities (p. 52)
func GenerateSystemPlacement(r roller.Roller, sys stars.System) (SystemPlacement, error) {
	counts, err := GenerateCounts(r, sys, CountsOpts{})
	if err != nil {
		return SystemPlacement{}, fmt.Errorf("worlds: counts: %w", err)
	}
	avail, err := AvailableOrbits(sys)
	if err != nil {
		return SystemPlacement{}, fmt.Errorf("worlds: available-orbits: %w", err)
	}
	allocs, err := AllocateOrbitsByStar(avail, counts)
	if err != nil {
		return SystemPlacement{}, fmt.Errorf("worlds: allocations: %w", err)
	}
	baselineN := RollBaselineNumber(r, sys, counts)
	primary := allocs[0].Group
	baselineOrbit := BaselineOrbit(r, primary, primary.HZCO(), baselineN, counts.Total)
	emptyOrbits, err := RollEmptyOrbits(r)
	if err != nil {
		return SystemPlacement{}, fmt.Errorf("worlds: empty-orbits: %w", err)
	}
	totalStars := 1 + secondaryStarCount(sys)
	spread := Spread(primary, allocs[0].AllocatedWorlds, baselineOrbit, baselineN, totalStars)
	slots, err := PlaceOrbitSlots(r, allocs, baselineN, baselineOrbit, spread, emptyOrbits)
	if err != nil {
		return SystemPlacement{}, fmt.Errorf("worlds: orbit-slots: %w", err)
	}
	anomSlots, newCounts, err := AddAnomalous(r, slots, allocs, counts)
	if err != nil {
		return SystemPlacement{}, fmt.Errorf("worlds: anomalous: %w", err)
	}
	placements, err := PlaceWorlds(r, anomSlots, newCounts)
	if err != nil {
		return SystemPlacement{}, fmt.Errorf("worlds: place-worlds: %w", err)
	}
	placements, err = RollPlanetEccentricities(r, placements, sys.AgeGyr)
	if err != nil {
		return SystemPlacement{}, fmt.Errorf("worlds: planet-eccentricity: %w", err)
	}
	return SystemPlacement{
		Counts:        newCounts,
		Allocations:   allocs,
		BaselineN:     baselineN,
		BaselineOrbit: baselineOrbit,
		EmptyOrbits:   emptyOrbits,
		SystemSpread:  spread,
		Placements:    placements,
	}, nil
}
