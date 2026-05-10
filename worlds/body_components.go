package worlds

// Per-body component types referenced from Body via nullable pointer.
// Cycle 0 declares each as an empty struct; fields land cycle-by-cycle
// as their respective Stage's procedures are implemented. The cycle
// schedule is in docs/pass-2/dependency-graph.md.

// BodyPhysical — Stage 3 (3A1). Composition, density, gravity, mass.
type BodyPhysical struct{}

// BeltDetails — Stage 3, belts only. Span, composition, bulk, resource,
// significant-size counts. Members []BeltMember is deferred per
// docs/pass-2/api-surface.md § Open questions, decided.
type BeltDetails struct{}

// DayLength — Stage 4 (3A2a). Sidereal hours, solar hours, year days.
type DayLength struct{}

// AxialTilt — Stage 4. Degrees plus extreme-tilt flag.
type AxialTilt struct{}

// TidalLock — Stage 4. Lock case (planet-to-star, moon-to-planet,
// planet-to-moon, none) plus the trigger metadata.
type TidalLock struct{}

// SurfaceTidalEffects — Stage 4. Per-zone tidal effects.
type SurfaceTidalEffects struct{}

// Atmosphere — Stage 5 climate. Code, pressure, taints, oxygen partial
// pressure. Populated post-ConvergeClimate.
type Atmosphere struct{}

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
