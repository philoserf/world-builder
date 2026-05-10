package worlds

// Per-body component types referenced from Body via nullable pointer.
// Cycle 0 declares each as an empty struct; fields land cycle-by-cycle
// as their respective Stage's procedures are implemented. The cycle
// schedule is in docs/pass-2/dependency-graph.md.

// Atmosphere — Stage 5 climate. Populated post-ConvergeClimate. Cycle 4
// reads Pressure for the dense-atmosphere tidal-lock DM; the rest of
// the fields land in cycle 5 (ConvergeClimate).
type Atmosphere struct {
	Code                  int
	Subtype               string
	Pressure              float64
	OxygenPartialPressure float64
	ScaleHeight           float64
	Taints                []string
}

// Hydrographics — Stage 5 climate. Code plus surface distribution
// inputs. Populated post-ConvergeClimate.
type Hydrographics struct{}

// Temperature — Stage 5 climate. MeanK plus per-zone variants
// (high/low/worst). Populated post-ConvergeClimate.
type Temperature struct{}

// SurfaceDistribution — Stage 6. Hydrosphere distribution across zones.
// Computed against converged hydrographics.
type SurfaceDistribution struct{}

// Geology — Stages 5–7. TSS components and tectonic plates.
type Geology struct{}

// Biology — Stage 8. Biomass, biocomplexity, sophonts, biodiversity,
// compatibility, terrestrial-resource rating.
type Biology struct{}

// Habitability — Stage 9. Rating plus contributing-DM notes.
type Habitability struct{}
