# API Surface

This document is the public-API reference: types, key signatures, conventions, and the rationale behind non-obvious choices. The actual signatures live in code; this document is the index and the design context.

The public surface is intentionally compact — roughly 200–300 public symbols — with internal helpers hidden behind procedure boundaries. The compactness comes from (a) cutting toggles for speculative variants, (b) the unified body iterator (a single `Universe.AllBodies()` replaces the earlier `buildMoonPlacementView` / `parentInfoOf` helper sprawl), and (c) keeping internal table lookups unexported.

## Conventions

### Package boundaries

```text
world-builder/
  dice/        — WBH dice notation parser; verbatim port from pass 1
  roller/      — Roller interface + Seeded/Scripted/Fixed; verbatim port
  stars/       — Stage 0
  worlds/      — Stages 1–10 (everything after stars)
  iiss/        — IISS form structs + renderers (split out from worlds/)
  cmd/world-builder/     — thin CLI wrapper
```

`iiss/` is a new package in pass 2. Pass 1 mixed renderers, form structs, and pipeline machinery in `worlds/`; splitting them out makes the renderer concerns explicitly separable and shrinks `worlds/` to its actual concern (per-body procedures + system-wide aggregations).

### Naming

- **`Generate*`** functions roll a complete sub-system from a Roller. They take a Roller and any upstream context. They return a value type plus an error. Examples: `GenerateSystem`, `GenerateUniverse`, `GenerateBodyPhysical`.
- **`Roll*`** functions roll a single value via dice. They take a Roller and any DMs/inputs. They return the rolled value plus an error. Examples: `RollAtmoCode`, `RollBiomass`.
- **`Compute*`** functions compute a deterministic value with no rolls. They take inputs and return a value with no error. Examples: `ComputeAlbedo` (roll moves out — see below), `ComputeHabitability`, `ComputeMeanTemperatureK`.
- **`Derive*`** functions derive one quantity from others as a pure formula. No Roller, no error. Examples: `DeriveMass`, `DeriveScaleHeight`.
- **`Apply*`** functions mutate a passed-in target. Document the mutation explicitly. Examples: `ApplyInherentTempAddition`. Mutators are minimized; pass 2 prefers returning replacement values.
- **`Render*`** functions return one of the three IISS form structs. Format-specific output (Markdown, JSON, plain text) lives in `iiss.MarkdownClass23(form)`, etc. — separate functions, not methods on the form.

`Compute*` previously coexisted with `Roll*` confusingly: pass 1's `ComputeAlbedo` reads from a Roller. Pass 2 renames roll-bearing computations to `RollAlbedo` for consistency. The discriminator is dice consumption, not whether the underlying procedure is "computational."

### Error semantics

- Roll-bearing functions return `(T, error)` because dice exhaustion is a real failure mode for `Scripted` rollers (panics for production `Seeded`).
- Pure-formula functions (`Derive*`, deterministic `Compute*`) return `T` with no error.
- Validation errors at trust boundaries (CLI input parsing) return wrapped errors with context.
- Misuse-path errors (illegal subtype, unknown atm code) panic in production, return error in test-friendly procedures. Each public function declares panic-or-error in `harness.md` § Misuse-path tests (Stance column) before stubs land; no procedure does both depending on caller.

### Mutability

The pipeline is mutator-shaped. `ApplyDetailFrontEnd`, `ApplyRotationTilt`, `ApplyClimate`, `ApplyTaintTypology`, etc., walk the universe and write to bodies in place. This is deliberate: the universe carries mutable per-body state through ten stages; returning a fresh `Universe` per stage would copy hundreds of bodies for no semantic gain.

Within procedures, value-typed inputs/outputs (atomic computations, table lookups) stay immutable. The boundary is at the stage-application level: `Apply*` mutates, `Compute*`/`Derive*`/`Roll*` returns. The doc-comment of every `Apply*` function cites the WBH page or the `dependency-graph.md` stage that requires the mutation.

### Types vs interfaces

- `Roller` is the only interface. It abstracts dice consumption.
- All other public types are concrete structs (or named primitives like `SizeCode`, `TempRange`).
- No "Renderer" interface, no "Procedure" interface, no "Body" interface. The unified body iterator yields concrete `*Body`; procedures take `*Body` directly.

## The Universe model

Top-level container produced by the full pipeline.

```go
package worlds

type Universe struct {
    System    stars.System          // Stage 0
    Placement SystemPlacement       // Stage 1
    Detail    SystemDetail          // Stages 2–10
}

// SystemDetail aggregates the per-body bodies and the system-wide
// aggregations (profiles, IISS forms, mainworld pick).
type SystemDetail struct {
    Bodies      []Body
    Allocations []StarAllocation
    Mainworld   *Body
    iiss.SystemForms   // Class0I + Class23 (PART 1 data carriers),
                       // Class4PForms []Class4PForm (one PART P/P.B per
                       // body), Census, profiles, NotableFeatures
}
// placement scalars (baseline number/orbit, spread, empty orbits) live on
// Universe.Placement; SystemForms.Census carries them for PART 1 rendering.
```

## The Body — unified per-body type

A single `Body` type holds every placed object (planet, moon, belt member). The unified iterator (`Universe.AllBodies()`) yields every body, eliminating the moon-path silent-zero anti-pattern at the type level: there is no separate moon code path that procedures can forget to walk.

Pass 2 unifies:

```go
type BodyKind int

const (
    BodyEmpty BodyKind = iota
    BodyTerrestrial
    BodyGasGiant
    BodyPlanetoidBelt
    BodyMoon            // moons are bodies; their parent is non-nil
)

type Body struct {
    // Identity
    Designation string
    Kind        BodyKind
    Parent      *Body   // nil for planets and belts; non-nil for moons

    // Stage 1 (placement)
    Group        Group
    Orbit        float64
    Eccentricity float64
    HZ           bool

    // Stage 2 (sizing/period)
    SizeCode      SizeCode
    DiameterKm    float64
    GGClass       GasGiantClass    // NotGasGiant for non-GGs
    DiameterEarth float64          // GG only
    MassEarth     float64
    Period        Period

    // Moon-specific (Kind == BodyMoon)
    OrbitPD     float64    // moon's orbit in planet-diameters
    OrbitKm     float64
    PeriodHours float64

    // Stage 3 (body physical / belt details)
    Physical *BodyPhysical    // nil for non-applicable kinds
    Belt     *BeltDetails     // belts only

    // Stage 4 (rotation/tilt/tide)
    DayLength           *DayLength
    AxialTilt           *AxialTilt
    TidalLock           *TidalLock
    TidalEffects        *SurfaceTidalEffects

    // Stage 5 (climate; populated by ApplyClimatePasses)
    Atmosphere    *Atmosphere
    Hydrographics *Hydrographics
    Temperature  *Temperature

    // Stage 6 (post-climate, post-taint)
    SurfaceDistribution *SurfaceDistribution

    // Stage 7 (geology)
    Geology *Geology

    // Stage 8 (biology)
    Biology *Biology

    // Stage 9 (habitability)
    Habitability *Habitability

    // Sub-bodies (moons) live in the children slice; iterator descends.
    Children []*Body
}
```

**Has\* predicates.** As pass 1: `body.HasAtmosphere()`, `body.HasGeology()`, etc. wrap nil-pointer checks. The list is canonical and shared.

**Iteration.** A single iterator walks every body in dependency-friendly order:

```go
// AllBodies yields every Body in the universe (planets, moons, belts) in
// ascending-orbit order within each star group, with each body's children
// yielded immediately after the parent. This order is contract — LongProfile
// and AssignPlanetDesignations rely on it.
func (u *Universe) AllBodies() iter.Seq[*Body] { ... }

// Bodies filters AllBodies to a predicate. Iteration order is *not* contract —
// consumers that need a specific order use AllBodies and filter inline. This
// leaves room for future order-agnostic callers (the per-body climate passes
// do not need ordering and could parallelize).
func (u *Universe) Bodies(filter func(*Body) bool) iter.Seq[*Body] { ... }
```

Procedures take `*Body` directly. The "moon path" is not a separate iteration — moons are yielded by `AllBodies` alongside their parents.

**Trade-off.** A moon's `Orbit` (around the parent planet) and its parent's `Orbit` (around the star) overlap in name. Resolution: for moons, `Body.Orbit` is unset (or zero), `OrbitPD`/`OrbitKm` carry the moon's orbit around its parent. Procedures that need "the orbit around the star" call `body.StellarOrbit()` which returns `body.Parent.Orbit` for moons or `body.Orbit` for planets. Explicit indirection.

## The climate solver — ApplyClimatePasses

Replaces pass-1's 5A-atm/hydro + 5C-temp + 5D-rederive 2-pass loop + 5E TSS-temperature-back-edge with a single explicit per-body entry that folds partial-geology (Residual + TSF + THF) into each rederive pass.

This is **not** a fixed-point solver. The original pass-2 design framed it as one — iterate until atm/hydro/temp stabilise, assert convergence within N — but investigation post-cycle-17 showed the climate cluster is not a fixed point: `RederiveAtmosphereHydrographics` calls `RollHydroDigit`, which consumes fresh dice each call, so every pass is a fresh stochastic sample of hydro (and via albedo, temperature), not a convergence step. There is no fixed point to converge to. The code runs **exactly 2 passes** (matching pass-1) and trusts the second sample; the function was renamed `ConvergeClimate → ApplyClimatePasses` and the `Climate` convergence-variable struct was removed as dead code. See `lessons-learned.md` §§ L13, L14.

```go
package worlds

// ApplyClimatePasses runs the atmosphere ↔ hydrographics ↔ temperature
// cluster (WBH pp.79, 81, 96-99, 102, 108-126) for the given body,
// folding partial geology (Residual + TSF + THF) into each pass so the
// post-TSS Temperature re-derives atm/hydro consistently. Mutates
// body.Atmosphere, body.Hydrographics, body.Temperature, and body.Geology
// (partial — TectonicPlates added in Stage 7) on return.
//
// Eligibility: HZ-orbit terrestrials and HZ-planet moons get a full
// climate. Non-HZ terrestrials and atm-less bodies short-circuit (no
// atm / hydro / temp populated).
//
// Behaviour: exactly 2 passes of (temp → partial-geology → TSS apply →
// rederive atm/hydro); the second sample is trusted. Not a fixed point —
// hydro is re-sampled from fresh dice each pass (lessons-learned.md § L13).
//
// Dice consumption: 2 × (temp + partial-geology + rederive atm/hydro
// inner-roll counts) plus the initial atm + hydro rolls, all drawn from
// the body's per-body "climate" sub-roller (docs/c1-subroller-plan.md).
func ApplyClimatePasses(r roller.Roller, body *Body, sys stars.System) error
```

**No convergence-variable struct.** Pass-2's original design proposed a `Climate` value type to hold the loop's working state; cycle-17's revert to 2-pass sampling made it dead code, and it was removed (`lessons-learned.md` § L14). `ApplyClimatePasses` mutates the `body` fields directly across the two passes.

**Why the partial-geology fold-in.** Per `dependency-graph.md` § Stage 7, the TSS back-edge into Temperature is real. The geology factors that depend only on body physical / orbital parameters (Residual, TSF, THF) are computed inside the loop. The factors that depend on stable TSS (Tectonic Plates, GG residual heat propagation) are computed after.

## Stage-by-stage signature design

Cross-references `dependency-graph.md` for the dependency rationale per stage. Each stage's procedures share the conventions above.

### Stage 0: Stars (`stars/`)

Verbatim port of pass 1, with cuts:

```go
type GenerateSystemOpts struct{}   // empty after cuts; kept for future evolution

func GenerateSystem(r roller.Roller, opts GenerateSystemOpts) (System, error)
```

`WithVariance` and `Accuracy` are gone (cut list § "All Opts variance fields", "All Opts accuracy fields"). The single canonical path uses the formula-table interpretation throughout.

Internal procedures (`RollPrimaryTypeAndClass`, `RollSubtype`, `ComputeMass`, etc.) keep their pass-1 signatures except for naming consistency (`Compute` → `Derive` where the procedure has no roll).

### Stage 1: SystemPlacement (`worlds/`)

```go
func GenerateSystemPlacement(r roller.Roller, sys stars.System) (SystemPlacement, error)
```

Same as pass 1. The 9-sub-stage pipeline (counts → allocations → baseline → spread → slots → anomalous → place → eccentricities) is preserved internally; the public face is the single function.

### Stage 2: Detail front-end

```go
// ApplyDetailFrontEnd populates Body.SizeCode, DiameterKm, MassEarth (for
// gas giants only — terrestrial mass is derived during Stage 3),
// Designation, Period, HZ, and Children (moons), for every Body in the
// universe. Belt details and body physical follow in Stage 3.
func ApplyDetailFrontEnd(r roller.Roller, u *Universe) error
```

Operates on the universe in place. Internally walks `u.AllBodies()` for each sub-step.

### Stage 3: Body Physical + Belt Details + Moon Refinement

```go
func GenerateBodyPhysical(r roller.Roller, body *Body, ageGyr float64) (BodyPhysical, error)
func GenerateBeltDetails(r roller.Roller, body *Body, sys stars.System, sp SystemPlacement) (BeltDetails, error)
func RefineMoons(r roller.Roller, parent *Body) error
```

`GenerateBodyPhysical` consumes `*Body` directly; the moon/planet distinction is a parameter to the procedure (read from `body.Kind`), not a separate code path.

### Stage 4: Rotation/Tilt/Tide (3A2a)

```go
func ApplyRotationTilt(r roller.Roller, u *Universe, sys stars.System) error
```

Internal sub-procedures: `GenerateDayLength`, `GenerateAxialTilt`, `GenerateTidalLock`, `GenerateSurfaceTidalEffects`. All take `*Body`; moons are walked by the unified iterator.

**Surface distribution moves to Stage 6** (post-climate). The project defers.

### Stage 5: ApplyClimatePasses

```go
func ApplyClimatePasses(r roller.Roller, body *Body, sys stars.System) error
```

(Detailed contract above.)

System-wide entry: `ApplyClimate(r, u)` walks all bodies and calls ApplyClimatePasses per body.

### Stage 6: Atmosphere taint typology + post-climate followups

```go
func ApplyTaintTypology(r roller.Roller, u *Universe) error
func ApplySurfaceDistribution(r roller.Roller, u *Universe) error
```

`ApplyTaintTypology` mutates `Body.Atmosphere` in place (oxygen-taint promotion can change `atm.Code`). Surface distribution runs after the climate passes; uses the post-climate `Body.Hydrographics`.

### Stage 7: Geology follow-ups

```go
func ApplyTectonicPlates(r roller.Roller, u *Universe) error
func ApplyGGResidualHeat(r roller.Roller, u *Universe) error
```

The TSS factors and partial geology are computed inside ApplyClimatePasses. These remaining geology procedures are forward-only and depend on stable post-climate state.

### Stage 8: Biology

```go
func ApplyBiology(r roller.Roller, u *Universe) error
```

Internal: `RollBiomass`, `RollBiocomplexity`, `RollNativeSophont`, `RollExtinctSophont`, `RollBiodiversity`, `RollCompatibility`, `RollTerrestrialResourceRating`. Strictly ordered per `dependency-graph.md` § Stage 8. The optional oxygen-atm biomass floor is cut.

### Stage 9: Habitability

```go
func ApplyHabitability(u *Universe)   // no Roller, no error — deterministic
```

Internal: `ComputeHabitability(*Body) Habitability`.

### Stage 10: System aggregations

```go
func AggregateSystem(u *Universe)
```

Computes BaselineN backfill, ShortProfile, LongProfile, the three IISS form structs, and pickMainworld. No rolls; deterministic.

## The IISS forms

Three sibling structs in `iiss/`. All renderers return structs from day one; no plain-text-only renderer exists in production code.

```go
package iiss

type Class0IForm struct { ... }
type Class23Form struct { ... }
type Class4PForm struct {
    Variant     Class4PVariant   // Planet, Moon, or Belt
    PartP       *Class4PPartP    // populated for Variant == Planet | Moon
    PartPB      *Class4PPartPB   // populated for Variant == Belt
    // shared header fields
}

type Class4PVariant int
const (
    Class4PPlanet Class4PVariant = iota
    Class4PMoon
    Class4PBelt
)
```

**Header type sharing.** Pass 1's `IISSClass23Header.SectorLocation string` (split on " | ") is replaced by `Sector string; Location string` — two fields per anti-pattern A.5.

```go
type FormHeader struct {
    SystemName string
    Sector     string
    Location   string
    // ...
}
```

`Class0IForm`, `Class23Form`, `Class4PForm` all embed `FormHeader`.

**Rendering.** Per-form per-format functions in `iiss/`:

```go
func MarkdownClass0I(f Class0IForm) string
func MarkdownClass23(f Class23Form) string
func MarkdownClass4P(f Class4PForm) string

func JSONClass0I(f Class0IForm) ([]byte, error)
func JSONClass23(f Class23Form) ([]byte, error)
func JSONClass4P(f Class4PForm) ([]byte, error)

func MarkdownClass4Survey(sf SystemForms) string // PART 1 census + a per-body PART P/P.B, under H1/H2

func PlainTextClass0I(f Class0IForm) string
func PlainTextClass23(f Class23Form) string
func PlainTextClass4P(f Class4PForm) string
```

No `Renderer` interface; no `(f Class0IForm) Markdown() string` method. Direct functions are simpler, do not impose an interface, and are trivially extended to a fourth format if one ever appears.

## The top-level façade

```go
package worlds

// Generate constructs a Seeded Roller from seed and delegates to
// GenerateWithRoller. The convenience entry for production callers
// (cmd/world-builder and end-users with a seed in hand).
func Generate(seed int64) (Universe, error)

// GenerateWithRoller runs the entire pipeline against any Roller.
// Tests use it with a Scripted roller to drive end-to-end fixtures
// through one entry point; cmd/world-builder and Generate use it via the seed
// convenience. All other entry points (GenerateSystem,
// GenerateSystemPlacement, individual Apply* stages) remain available
// for callers that need finer control.
func GenerateWithRoller(r roller.Roller) (Universe, error)
```

`GenerateWithRoller(r)` is the pipeline. It:

1. Calls `GenerateSystem(r, GenerateSystemOpts{})`.
2. Calls `GenerateSystemPlacement(r, sys)`.
3. Constructs `Universe` with empty Detail.
4. Calls `ApplyDetailFrontEnd`, `GenerateBodyPhysical` (per body), `RefineMoons`, `ApplyRotationTilt`, `ApplyClimate`, `ApplyTaintTypology`, `ApplySurfaceDistribution`, `ApplyTectonicPlates`, `ApplyGGResidualHeat`, `ApplyBiology`, `ApplyHabitability`, `AggregateSystem`.
5. Returns the universe.

`Generate(seed)` constructs a `roller.NewSeeded(seed)` and delegates. Splitting the two means harness fixtures (`harness.md` § Façade end-to-end) can drive the full pipeline through the public API with a Scripted roller — without the seed convenience getting in the way.

`cmd/world-builder/main.go` becomes:

```go
u, err := worlds.Generate(*seedFlag)
if err != nil { /* ... */ }
fmt.Print(iiss.MarkdownClass4Survey(u.Detail.SystemForms))
```

Three lines of pipeline. Same shape as pass 1's CLI.

## Roller and dice

Verbatim port from pass 1. The `Roller` interface is the project's keystone.

```go
package roller

type Roller interface {
    Roll(notation string) int
}

func NewSeeded(seed int64) Roller
func NewScripted(rolls ...int) Roller   // panics on exhaustion
func NewFixed(value int) Roller
```

**Pass-1 lesson honored.** `NewScripted` takes Roll _results_, not per-die values (per `feedback_world_builder_scripted_takes_results` memory). This counterintuitive shape is preserved because pass-1's worked-example dice scripts encode it.

```go
package dice

type Spec struct {
    Count    int
    Sides    int
    Modifier int
}

func Parse(notation string) (Spec, error)
```

## Errors and misuse contracts

Two failure-mode categories:

### 1. Roller exhaustion (Scripted only)

Production `Seeded` rolls indefinitely; `Scripted` panics on exhaustion. This is by design — exhaustion always indicates a fixture bug, never a runtime issue. Pass 2 keeps the panic.

### 2. Misuse paths (illegal inputs)

Pass 1's RollGasMix bug — passing a Subtype where a column letter was expected — must be impossible in pass 2. Two prevention mechanisms:

- **Type system encoding.** Where the input has discrete enumerable values, encode them as a typed string or named integer. `AtmosphereColumnLetter` distinct from `AtmosphereSubtype` distinct from raw `string`. Misuse becomes a compile error.
- **Contract tests.** For every public function, a test verifies that plausible misuse paths return an error or compile-fail. The harness catalogs misuse-path tests as a class of fixture (`harness.md` § Misuse-path tests). They run alongside worked-example tests.

### Misuse-path test pattern

For each public function, the spec lists the plausible misuse paths and the corresponding test. Example:

```text
RollGasMix(r, columnLetter, atmCode) returns (string, error)

Misuse paths:
1. Pass a Subtype letter ("c") instead of a column letter ("A"/"B"/"C")
   → contract test asserts error or compile-fail.
2. Pass an atmCode outside [2..9] ∪ {D, E}
   → contract test asserts error.
3. Pass an empty columnLetter
   → contract test asserts error.

Implementation: TBD per stub commitment.
```

The pattern is mandatory. No public function ships without misuse-path test coverage.

## What this document commits to

1. A unified `Body` type with one iterator (`Universe.AllBodies()`), eliminating the moon-path silent-zero anti-pattern at the type level.
2. A consolidated `ApplyClimatePasses` per-body solver that folds partial-geology into the rederive flow. (Originally framed as a fixed-point solver; investigation reclassified it as deterministic 2-pass sampling — see `history/lessons-learned.md` § L13.)
3. Three IISS form structs with per-form Markdown renderers in a separate `iiss/` package; `iiss/` does not import `worlds/`.
4. A compact public surface (~200–300 symbols) with internal helpers hidden behind procedure boundaries.
5. Misuse-path contract tests as a mandatory class of fixture for every public function.

The companion `harness.md` catalogs the worked-example fixtures; `dependency-graph.md` maps every value to its inputs; `wbh-inconsistencies.md` consolidates the six book-internal divergences.
