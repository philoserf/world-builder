package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

// Generate constructs a Seeded Roller from seed and delegates to
// GenerateWithRoller. The convenience entry for production callers
// (cmd/world-builder and end-users with a seed in hand).
func Generate(seed int64) (Universe, error) {
	return GenerateWithRoller(roller.NewSeeded(seed))
}

// GenerateOpts carries caller-tunable generation options for the
// opts-taking entry points. The zero value reproduces Generate /
// GenerateWithRoller exactly.
type GenerateOpts struct {
	// PeculiarColumn selects which WBH p.15-16 column resolves a
	// "Special" (2D=2) primary roll: Special (zero value), Unusual, or
	// Peculiar. Only Unusual/Peculiar can yield BD, D, Pulsar, Neutron
	// Star, Black Hole, Nebula, Protostar, Star Cluster, or Anomaly
	// primaries.
	PeculiarColumn stars.PeculiarPath
}

// GenerateWithOpts is Generate with options.
func GenerateWithOpts(seed int64, opts GenerateOpts) (Universe, error) {
	return GenerateWithRollerOpts(roller.NewSeeded(seed), opts)
}

// GenerateWithRoller runs the entire pass-2 pipeline against any
// Roller. Tests use it with a Scripted roller for narrow per-procedure
// fixtures and with a Seeded roller for façade end-to-end fixtures
// (per docs/harness.md § Façade end-to-end). cmd/world-builder and
// Generate use it via the seed convenience.
//
// All other entry points (GenerateSystem, GenerateSystemPlacement,
// individual Apply* stages) remain available for callers that need
// finer control.
func GenerateWithRoller(r roller.Roller) (Universe, error) {
	return GenerateWithRollerOpts(r, GenerateOpts{})
}

// GenerateWithRollerOpts is GenerateWithRoller with options.
func GenerateWithRollerOpts(r roller.Roller, opts GenerateOpts) (Universe, error) {
	sys, err := stars.GenerateSystem(r, stars.GenerateSystemOpts{
		WithVariance:   true,
		Accuracy:       2,
		MAO:            MAO,
		PeculiarColumn: opts.PeculiarColumn,
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
		ApplyTidalLockReEval,
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
