package worlds

import (
	"wbh/roller"
	"wbh/stars"
)

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

// ConvergeClimate finds the fixed point of the atmosphere / hydrographics
// / temperature / TSS cluster for the given body. Mutates body's climate
// pointer fields on return.
//
// Eligibility: HZ-orbit terrestrials and HZ-planet moons get a full
// climate. Non-HZ terrestrials and atm-less bodies receive a degenerate
// Climate with nil pointers.
//
// Convergence: iterates until atm.Code, hydro.Code, and temp.MeanK are
// stable. Cap N = 3 iterations. Asserts convergence; the panic-vs-error
// stance for production Seeded rollers is deferred to the cycle-6
// (Climate) spec per docs/pass-2/spike-findings.md § Finding 6a.
//
// Dice consumption: bounded by N * (atm + hydro + temp + partial-geology
// inner-roll counts). The Roller must support the worst case.
func ConvergeClimate(r roller.Roller, body *Body, sys stars.System) error {
	panic("unimplemented: see docs/pass-2/api-surface.md § The Climate solver — ConvergeClimate")
}
