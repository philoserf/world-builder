package worlds

import (
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
	panic("unimplemented: see docs/pass-2/api-surface.md § The top-level façade")
}

// ApplyDetailFrontEnd populates Body.SizeCode, DiameterKm, MassEarth
// (gas giants only — terrestrial mass is derived during Stage 3),
// Designation, Period, HZ, and Children (moons), for every Body in the
// universe. Belt details and body physical follow in Stage 3.
//
// Operates on the universe in place (per docs/pass-2/api-surface.md §
// Mutability — the pipeline is mutator-shaped).
func ApplyDetailFrontEnd(r roller.Roller, u *Universe) error {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 2: Detail front-end")
}

// ApplyRotationTilt populates DayLength, AxialTilt, TidalLock, and
// SurfaceTidalEffects for every body in the universe. Surface
// distribution is deferred to ApplySurfaceDistribution (Stage 6) so
// it runs against converged hydrographics.
func ApplyRotationTilt(r roller.Roller, u *Universe) error {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 4: Rotation/Tilt/Tide")
}

// ApplyClimate walks the universe and calls ConvergeClimate per body.
// Stage-5 entry point. Per dependency-graph.md § Stage 5, the climate
// fixed-point cluster (atm/hydro/temp + partial geology) converges
// per body before downstream stages run.
func ApplyClimate(r roller.Roller, u *Universe) error {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 5: ConvergeClimate")
}

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

// GenerateBodyPhysical rolls composition / density / gravity / mass
// for one terrestrial body. Stage-3 leaf procedure.
func GenerateBodyPhysical(r roller.Roller, body *Body, ageGyr float64) (BodyPhysical, error) {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 3")
}

// GenerateBeltDetails rolls span / composition / bulk / resource /
// significant-size counts for one belt. Stage-3 leaf procedure.
func GenerateBeltDetails(r roller.Roller, body *Body, sys stars.System, sp SystemPlacement) (BeltDetails, error) {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 3")
}

// RefineMoons computes Hill-sphere moon orbit limit for parent, may
// remove moons exceeding the limit, then per-moon orbit and period.
// Stage-3 procedure for terrestrial parents and gas giants alike.
func RefineMoons(r roller.Roller, parent *Body) error {
	panic("unimplemented: see docs/pass-2/api-surface.md § Stage 3")
}
