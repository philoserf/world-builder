package worlds

// Per-body component types referenced from Body via nullable pointer.
// Cycle 0 declares each as an empty struct; fields land cycle-by-cycle
// as their respective Stage's procedures are implemented. The cycle
// schedule is in docs/pass-2/dependency-graph.md.

// Atmosphere, Hydrographics, Temperature live in atmosphere.go,
// hydrographics.go, temperature.go (cycle 5).

// SurfaceDistribution lives in surface_distribution.go (cycle 5/6).

// Geology lives in geology.go (cycle 7).

// Biology lives in biology.go (cycle 8).

// Habitability — Stage 9. Rating plus contributing-DM notes.
type Habitability struct{}
