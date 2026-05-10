package worlds

import (
	"fmt"

	"wbh/roller"
	"wbh/stars"
)

// Generate constructs a Seeded Roller from seed and delegates to
// GenerateWithRoller. The convenience entry for production callers
// (cmd/wbh and end-users with a seed in hand).
func Generate(seed int64) (Universe, error) {
	return GenerateWithRoller(roller.NewSeeded(seed))
}

// GenerateWithRoller runs the entire pass-2 pipeline against any
// Roller. Tests use it with a Scripted roller for narrow per-procedure
// fixtures and with a Seeded roller for façade end-to-end fixtures
// (per docs/pass-2/harness.md § Façade end-to-end). cmd/wbh and
// Generate use it via the seed convenience.
//
// All other entry points (GenerateSystem, GenerateSystemPlacement,
// individual Apply* stages) remain available for callers that need
// finer control.
func GenerateWithRoller(r roller.Roller) (Universe, error) {
	sys, err := stars.GenerateSystem(r, stars.GenerateSystemOpts{
		WithVariance: true,
		Accuracy:     2,
	})
	if err != nil {
		return Universe{}, fmt.Errorf("worlds: stars: %w", err)
	}
	sp, err := GenerateSystemPlacement(r, sys)
	if err != nil {
		return Universe{}, fmt.Errorf("worlds: placement: %w", err)
	}
	u := Universe{System: sys, Placement: sp}
	for _, step := range []func(roller.Roller, *Universe) error{
		ApplyDetailFrontEnd,
		ApplyBodyPhysical,
		ApplyBeltDetails,
		ApplyMoonRefinement,
		ApplyRotationTilt,
		ApplyClimate,
		ApplyTaintTypology,
		ApplySurfaceDistribution,
		ApplyGeology,
		ApplyBiology,
	} {
		if err := step(r, &u); err != nil {
			return u, err
		}
	}
	ApplyHabitability(&u)
	AggregateSystem(&u)
	BuildIISSForms(&u)
	return u, nil
}
