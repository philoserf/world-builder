package worlds

import (
	"errors"

	"wbh/roller"
	"wbh/stars"
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

// ErrContinuationMethodUnsupported is returned when sys carries
// pre-existing mainworld data (no field today; placeholder for the
// future Continuation Method sub-project).
var ErrContinuationMethodUnsupported = errors.New(
	"worlds: pre-existing mainworld input requires Continuation Method, not yet encoded",
)

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
		return SystemPlacement{}, err
	}
	avail, err := AvailableOrbits(sys)
	if err != nil {
		return SystemPlacement{}, err
	}
	allocs, err := AllocateOrbitsByStar(avail, counts)
	if err != nil {
		return SystemPlacement{}, err
	}
	baselineN, err := RollBaselineNumber(r, sys, counts)
	if err != nil {
		return SystemPlacement{}, err
	}
	primary := allocs[0].Group
	baselineOrbit, err := BaselineOrbit(r, primary, primary.HZCO(), baselineN, counts.Total)
	if err != nil {
		return SystemPlacement{}, err
	}
	emptyOrbits, err := RollEmptyOrbits(r)
	if err != nil {
		return SystemPlacement{}, err
	}
	totalStars := 1 + secondaryStarCount(sys)
	spread := Spread(primary, allocs[0].AllocatedWorlds, baselineOrbit, baselineN, totalStars)
	slots, err := PlaceOrbitSlots(r, allocs, baselineN, baselineOrbit, spread, emptyOrbits)
	if err != nil {
		return SystemPlacement{}, err
	}
	anomSlots, newCounts, err := AddAnomalous(r, slots, allocs, counts)
	if err != nil {
		return SystemPlacement{}, err
	}
	placements, err := PlaceWorlds(r, anomSlots, newCounts)
	if err != nil {
		return SystemPlacement{}, err
	}
	placements, err = RollPlanetEccentricities(r, placements)
	if err != nil {
		return SystemPlacement{}, err
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
