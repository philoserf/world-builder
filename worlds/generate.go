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

// ApplyTaintTypology mutates Body.Atmosphere in place — oxygen-taint
// promotion can change atm.Code; corrosive / insidious typology may
// add taints. Runs after ConvergeClimate. Stage-6 entry point.
func ApplyTaintTypology(r roller.Roller, u *Universe) error {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 6")
}

// ApplySurfaceDistribution computes the hydrosphere distribution
// across surface zones for every HZ terrestrial. Runs after climate
// converges (so hydrographics is final) and after taint typology.
func ApplySurfaceDistribution(r roller.Roller, u *Universe) error {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 6")
}

// ApplyTectonicPlates rolls tectonic-plate count for terrestrial
// bodies whose total seismic stress is stable post-climate.
// Stage-7 procedure (forward-only after ConvergeClimate's TSS fold-in).
func ApplyTectonicPlates(r roller.Roller, u *Universe) error {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 7")
}

// ApplyGGResidualHeat computes residual heat for gas giants. No
// climate-cluster dependence; lives in Stage 7 alongside terrestrial
// geology follow-ups.
func ApplyGGResidualHeat(r roller.Roller, u *Universe) error {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 7")
}

// ApplyBiology walks atmosphere-bearing terrestrials (and HZ-planet
// moons with atmosphere) and rolls biomass → biocomplexity → sophonts
// → biodiversity → compatibility → terrestrial-resource rating in the
// strict order specified by dependency-graph.md § Stage 8.
func ApplyBiology(r roller.Roller, u *Universe) error {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 8: Biology")
}

// ApplyHabitability computes per-body habitability ratings for
// terrestrials. Pure function; no rolls. Stage-9 entry point.
func ApplyHabitability(u *Universe) {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 9: Habitability")
}

// AggregateSystem computes the system-wide aggregations after every
// body has converged: BaselineN backfill per allocation, ShortProfile,
// LongProfile, the three IISS form structs, and the auto-picked
// mainworld designation. Pure function; no rolls. Stage-10 entry point.
func AggregateSystem(u *Universe) {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 10")
}

// GenerateBodyPhysical lives in body_physical.go (real implementation).
// GenerateBeltDetails lives in belt_details.go (real implementation).
// Stage-3 orchestration (the mass-derivation, moon-refinement walks)
// lives in stage3.go.
