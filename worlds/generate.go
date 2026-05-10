package worlds

import (
	"wbh/roller"
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
	panic("unimplemented: see docs/pass-2/api-surface.md § The top-level façade")
}

// ApplyRotationTilt is implemented in stage4.go.

// ApplyClimate is implemented in stage5.go.

// ApplyTaintTypology and ApplySurfaceDistribution are implemented in
// stage6.go.

// ApplyGeology (single Stage-7 entry covering residual seismic, TSF,
// THF, GG residual heat, post-TSS temperature update, scale-height
// recompute, and tectonic plates) is implemented in stage7.go.

// ApplyBiology is implemented in stage8.go.

// ApplyHabitability is implemented in stage9.go.

// AggregateSystem is implemented in stage10.go.

// GenerateBodyPhysical lives in body_physical.go (real implementation).
// GenerateBeltDetails lives in belt_details.go (real implementation).
// Stage-3 orchestration (the mass-derivation, moon-refinement walks)
// lives in stage3.go.
