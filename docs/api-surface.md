# Pass 2 — API Surface Design

This document specifies the public API design for pass 2: types, key signatures, conventions, and the delta from pass 1. It is the contract that procedures honor, the shape callers depend on, and the rationale for each non-obvious choice.

The actual stub commitments — every public signature compiling but unimplemented — land in code as the first implementation step (per `design-intent.md` § Implementation order). When stubs are written, they cross-reference this document by section. Until then, this document is the source of truth for "what should the public surface look like."

Pass 1's public surface was 1048 symbols. Pass 2 targets a smaller surface: roughly 200–300 public symbols, with the rest hidden behind procedure boundaries. The reduction comes from (a) the cuts list (variance/accuracy options gone, optional rules gone), (b) the unified body iterator (replaces `buildMoonPlacementView`, `parentInfoOf`, and several helper layers), and (c) hiding internal table-lookups that pass 1 exposed unnecessarily.

## Conventions

### Package boundaries

```text
wbh/
  dice/        — WBH dice notation parser; verbatim port from pass 1
  roller/      — Roller interface + Seeded/Scripted/Fixed; verbatim port
  stars/       — Stage 0
  worlds/      — Stages 1–10 (everything after stars)
  iiss/        — IISS form structs + renderers (split out from worlds/)
  cmd/wbh/     — thin CLI wrapper
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
    Bodies               []Body
    Allocations          []StarAllocation
    BaselineN            int
    BaselineOrbit        float64
    EmptyOrbits          int
    SystemSpread         float64
    ShortProfile         string
    LongProfile          string
    Class0I              iiss.Class0IForm
    Class23              iiss.Class23Form
    Class4P              iiss.Class4PForm     // mainworld only
    MainworldDesignation string
}
```

**Pass-1 delta.** Pass 1 had `SystemPlacement → DetailedPlacement[]` with `DetailedPlacement{Placement, ...}` embedding. Pass 2 flattens to a single `Body` type (next section) and groups everything system-wide into `SystemDetail`. The Universe wrapper is the top-level handoff to renderers and the CLI.

## The Body — unified per-body type

The single most important pass-2 design change. Pass 1 had `DetailedPlacement` (planet) and `Moon` as different types with `buildMoonPlacementView` to coerce one into the other. The moon-path silent-zero anti-pattern recurred four times because procedures could iterate the planet path without iterating the moon path.

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

    // Stage 5 (climate; populated post-ConvergeClimate)
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
// leaves room for future order-agnostic callers (per-body climate convergence
// does not need ordering and could parallelize).
func (u *Universe) Bodies(filter func(*Body) bool) iter.Seq[*Body] { ... }
```

Procedures take `*Body` directly. The "moon path" is not a separate iteration — moons are yielded by `AllBodies` alongside their parents.

**Pass-1 delta.** Removes `Moon`, `DetailedPlacement`, `buildMoonPlacementView`, `parentInfoOf`. Adds `Body`, `BodyKind.BodyMoon`, the iterator. Reduces the moon-path silent-zero attack surface to zero.

**Trade-off.** A moon's `Orbit` (around the parent planet) and its parent's `Orbit` (around the star) overlap in name. Pass 2 resolves: for moons, `Body.Orbit` is unset (or zero), `OrbitPD`/`OrbitKm` carry the moon's orbit around its parent. Procedures that need "the orbit around the star" call `body.StellarOrbit()` which returns `body.Parent.Orbit` for moons or `body.Orbit` for planets. This explicit indirection prevents the pass-1 confusion where moon code "happened to work" because some procedures aliased through `dp.Orbit` and others didn't.

## The Climate solver — ConvergeClimate

Replaces pass-1's 5A-atm/hydro + 5C-temp + 5D-rederive 2-pass loop + 5E TSS-temperature-back-edge with a single explicit per-body entry that folds partial-geology (Residual + TSF + THF) into each rederive pass.

The original pass-2 design called this a "fixed-point solver" with formal N-iteration convergence assertion. Empirical investigation post-cycle-17 showed the climate cluster is NOT a fixed point in the strict mathematical sense — `RederiveAtmosphereHydrographics` calls `RollHydroDigit`, which consumes fresh dice from the Roller each call. Every iteration is a fresh stochastic sample of hydro (and via albedo, temperature), not a convergence step. The name "ConvergeClimate" is retained for continuity but is a misnomer; pass-2 runs exactly 2 passes (matching pass-1's behaviour) and accepts the second sample. See `lessons-learned.md` § L13.

```go
package worlds

// Climate is the convergence variable for the atmosphere ↔ hydrographics
// ↔ temperature fixed-point cluster (WBH pp.79, 81, 96-99, 102, 108-126).
// It is local to ConvergeClimate; the result is unpacked back onto Body.
type Climate struct {
    Atmosphere    *Atmosphere
    Hydrographics *Hydrographics
    Temperature   *Temperature
    PartialGeology *PartialGeology   // residual + TSF + THF; excludes
                                     // tectonic plates (post-converge)
}

// ConvergeClimate runs the atm/hydro/temp/TSS cluster for the given
// body. Mutates body.Atmosphere, body.Hydrographics, body.Temperature,
// and body.Geology (partial — TectonicPlates added in Stage 7) on
// return.
//
// Eligibility: HZ-orbit terrestrials and HZ-planet moons get a full
// climate. Non-HZ terrestrials and atm-less bodies short-circuit (no
// atm / hydro / temp populated).
//
// Behaviour: runs exactly 2 passes of (temp → partial-geology → TSS
// apply → rederive). Matches pass-1's 2-rederive flow. The "convergence"
// framing of the original pass-2 design proved unrealizable — see
// lessons-learned.md § L13 (the cluster isn't a fixed point because
// hydro is re-sampled per pass).
//
// Dice consumption: 2 × (atm + hydro + temp + partial-geology inner-roll
// counts) plus the initial atm + hydro rolls. Determined entirely by
// the Roller's sequence.
func ConvergeClimate(r roller.Roller, body *Body, sys stars.System) error
```

**Why a struct, not just successive mutations on Body.** The convergence variable is local to the loop. Exposing it externally would let callers reach into a half-converged state. The Climate type is internal-by-design; only ConvergeClimate constructs it. Post-parity, a sibling `ConvergeClimateWithTrace` may expose iteration history for debugging unexpected atm flips; the type signature leaves room.

**Why the partial-geology fold-in.** Per `dependency-graph.md` § Stage 7, the TSS back-edge into Temperature is real. Pass 1 ignored it; pass 2 includes it. The geology factors that depend only on body physical / orbital parameters (Residual, TSF, THF) are computed inside the loop. The factors that depend on stable TSS (Tectonic Plates, GG residual heat propagation) are computed after.

**Pass-1 delta.** Removes `RederiveAtmosphereHydrographics`, the 2-pass loop in pass-1's `runStep5D`, and the 1-pass forward update of `ApplyInherentTempAddition` in `runStep5E`'s temp section. All folded into ConvergeClimate.

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

**Surface distribution moves to Stage 6** (post-climate). Pass 1 ran it here against preliminary hydro; pass 2 defers.

### Stage 5: ConvergeClimate

```go
func ConvergeClimate(r roller.Roller, body *Body, sys stars.System) error
```

(Detailed contract above.)

System-wide entry: `ApplyClimate(r, u, sys)` walks all bodies and calls ConvergeClimate per body.

### Stage 6: Atmosphere taint typology + post-climate followups

```go
func ApplyTaintTypology(r roller.Roller, u *Universe) error
func ApplySurfaceDistribution(r roller.Roller, u *Universe) error
```

`ApplyTaintTypology` mutates `Body.Atmosphere` in place (oxygen-taint promotion can change `atm.Code`). Surface distribution runs after climate has converged; uses post-converge `Body.Hydrographics`.

### Stage 7: Geology follow-ups

```go
func ApplyTectonicPlates(r roller.Roller, u *Universe) error
func ApplyGGResidualHeat(r roller.Roller, u *Universe) error
```

The TSS factors and partial geology are computed inside ConvergeClimate. These remaining geology procedures are forward-only and depend on stable post-climate state.

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

func MarkdownSystem(u *worlds.Universe) string   // concatenates all three under H1/H2

func PlainTextClass0I(f Class0IForm) string
func PlainTextClass23(f Class23Form) string
func PlainTextClass4P(f Class4PForm) string
```

No `Renderer` interface; no `(f Class0IForm) Markdown() string` method. Direct functions are simpler, do not impose an interface, and are trivially extended to a fourth format if one ever appears.

**Pass-1 delta.** Removes plain-text-only `RenderIISSClass4P`. Splits each form's renderers into per-format files (`iiss/class4p_markdown.go`, `iiss/class4p_json.go`, `iiss/class4p_plaintext.go`).

## The top-level façade

```go
package worlds

// Generate constructs a Seeded Roller from seed and delegates to
// GenerateWithRoller. The convenience entry for production callers
// (cmd/wbh and end-users with a seed in hand).
func Generate(seed int64) (Universe, error)

// GenerateWithRoller runs the entire pipeline against any Roller.
// Tests use it with a Scripted roller to drive end-to-end fixtures
// through one entry point; cmd/wbh and Generate use it via the seed
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

`cmd/wbh/main.go` becomes:

```go
u, err := worlds.Generate(*seedFlag)
if err != nil { /* ... */ }
fmt.Print(iiss.MarkdownSystem(&u))
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

## Pass-1 → pass-2 delta summary

Removed:

- **Embedded chain depth.** `Slot → AnomalousSlot → Placement → DetailedPlacement` (4 levels) collapses to `Body` (1 level, with `Group` embedded for shared fields).
- **`buildMoonPlacementView`, `parentInfoOf`** and several helper layers — replaced by unified `Body` + iterator.
- **`Moon` struct** — moons become `Body{Kind: BodyMoon, Parent: <planet>}`.
- **`RederiveAtmosphereHydrographics` and the 2-pass loop** — folded into `ConvergeClimate`.
- **`*Opts.WithVariance`, `*Opts.Accuracy`, `DetailOpts.OxygenAtmBiomassFloor`** — cut per `design-intent.md`.
- **`SurveyForm.ClassI bool`** dead field — pass-2 fields are added with both writer and reader simultaneously.
- **Plain-text-only `RenderIISSClass4P`** — replaced by struct + per-format renderers.
- **`IISSClass23Header.SectorLocation` (single string)** — replaced by `Sector` + `Location` fields.

Renamed:

- **`ComputeAlbedo` → `RollAlbedo`** (consumes Roller).
- **`ComputeGreenhouseFactor` → `RollGreenhouseFactor`** (consumes Roller).
- **`atmosphereSubtype` parameter (in `RollGasMix`) → `atmosphereColumnLetter`** with `AtmosphereColumnLetter` typed string.

Added:

- **`Body`, `BodyKind`, `Universe`** — top-level model.
- **`Climate`, `ConvergeClimate`** — the fixed-point solver.
- **`Universe.AllBodies() iter.Seq[*Body]`** — Go 1.23+ unified iterator.
- **`iiss/` package** — IISS forms + renderers split out.
- **`Class4PForm` (struct)** — pass-1 plain-text renderer becomes a struct + format renderers.
- **`Class0IForm`, `Class23Form`** updated headers (`Sector`, `Location` as separate fields).

## Stub commitment scope

The stubs are the API surface in code. They include:

- All public types declared.
- All public functions declared with bodies that `panic("unimplemented: see docs/api-surface.md § <section>")`.
- All `Has*()` predicates declared.
- All renderer functions declared.
- The Universe / Body / Climate types fully fielded (pass-2 tests can construct empty values).

After stubs land, the harness fixtures are written against them (red). After harness is red, implementation proceeds, driven by which fixture is closest to green next, in dependency-graph order.

The stub commit is a single PR-sized change, ~30 source files, no implementation logic. It is reviewable as one artifact. Subsequent implementation cycles touch a small subset of those files per cycle.

## Open questions, decided

These decisions resolve the stub line-items they blocked. Cycle 0 (stub commit) honors them.

- **Where does pass-2's `Group` live? — `stars/`.** Group is stellar-group geometry (HZCO, MAO, parent-star references). Pass-1 had `worlds.Group`; pass-2 moves to `stars.Group`. `worlds/` imports `stars/` and uses `stars.Group` as a foreign type at boundaries. The `Body` struct's `Group` field is typed as `stars.Group`.
- **`OrbitToAU` and HZCO accessors — methods on `stars.Star`.** Pass 1 mixed methods and free functions; pass 2 picks methods. `star.HZCOOrbit() float64`, `star.OrbitToAU(orbit float64) float64`. No sibling free functions; no two ways to compute the same value.
- **Belt member detail — omitted from cycle-0 stubs.** `BeltDetails` does not declare `Members []BeltMember`. `Class4PPartPB` is an empty struct shell — `Variant == Class4PBelt` on the parent `Class4PForm` is the only marker that the belt variant exists. Both are fleshed out post-parity, when the belt-mainworld fixture in `harness.md` § Class4P/PartPB lands. Per anti-pattern A.4 (dead fields), declaring a field without a writer/reader is forbidden; the omission is the right call.
- **`ConvergeClimate` panic-on-overflow stance — deferred to cycle-6 (Climate) spec.** Per `spike-findings.md` § Finding 6a; the production-vs-test trade-off (raise N? log and degrade? introspect Roller type?) is not pre-stub work. Cycle 0 stubs `ConvergeClimate` as `panic("unimplemented: see api-surface.md § ConvergeClimate")` — the convergence behavior is irrelevant until cycle 6.

## What this document commits pass 2 to

1. A unified `Body` type with one iterator, eliminating the moon-path silent-zero anti-pattern at the type level.
2. A consolidated `ConvergeClimate` per-body solver that folds partial-geology into the rederive flow. (Originally framed as a fixed-point solver; post-cycle-17 investigation reclassified it as deterministic 2-pass sampling — see `lessons-learned.md` § L13.)
3. Three IISS form structs from day one, with per-form per-format renderers in a separate `iiss/` package.
4. A smaller public surface (target ~200–300 symbols) by hiding internal helpers behind procedure boundaries.
5. Misuse-path contract tests as a mandatory class of fixture for every public function.
6. The pass-1 → pass-2 delta is enumerated; stubs are written against this document; tests are written against the stubs; implementation follows.

The next document, `harness.md`, catalogs the worked-example fixtures that gate the implementation.
