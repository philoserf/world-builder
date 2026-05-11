package worlds

// Climate is the convergence variable for the atmosphere ↔ hydrographics
// ↔ temperature fixed-point cluster (WBH pp.79, 81, 96-99, 102, 108-126).
// It is local to ConvergeClimate; the result is unpacked back onto Body.
//
// Per docs/pass-2/api-surface.md § The Climate solver, exposing Climate
// externally would let callers reach into a half-converged state. The
// type is internal-by-design: only ConvergeClimate constructs it.
//
// Post-parity, a sibling ConvergeClimateWithTrace may expose iteration
// history for debugging unexpected atm flips; the type signature leaves
// room.
type Climate struct {
	Atmosphere     *Atmosphere
	Hydrographics  *Hydrographics
	Temperature    *Temperature
	PartialGeology *Geology // residual + TSF + THF; tectonic plates land post-converge
}

// ConvergeClimate is implemented in stage5.go.
