# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Source material

Every spec, plan, and implementation references Mongoose Publishing's _World Builder's Handbook_ (Geir Lanesskog, 2023) as the canonical authority. Cite WBH page numbers in doc-comments next to tables and procedures. To work locally, place a copy of the PDF at `docs/World Builders Handbook.pdf` (gitignored — copyright).

When the book contradicts itself (worked example vs formula box, table A vs table B), surface the divergence rather than silently picking one. Several known book inconsistencies are documented in design specs.

## Common commands

```bash
task              # default → check + test
task check        # modernize gate → gofumpt + go vet + golangci-lint
task test         # go test -race ./...
task fmt          # gofumpt -l -w -extra .
task tidy         # go mod tidy

go test ./worlds/ -run TestZed_FullDetail_3A2b   # single test
go test ./stars/ -run TestSolTerra_p35           # worked-example test
go run ./cmd/world-builder -seed 42 -format short    # run the CLI
```

`task check` runs `go fix ./...` first (modernizer pass) and **fails if it produces any diff** — modernizer hints are mandatory, not advisory. If `task check` complains, review the diff with `git diff` and commit it before continuing.

`gofumpt` is enforced via the CLI (not via golangci-lint's bundled copy) because the two disagreed on import grouping. Use `gofumpt -l -w -extra .` for formatting; do not enable golangci-lint's gofumpt linter.

## Architecture

**Before adding a pipeline stage:** verify moons are visited (including moons of gas-giant parents). Per-stage planet/moon divergence is the highest-frequency Critical bug in this codebase — see _Moons mirror planets_ below.

### Deterministic-dice pipeline

Every dice roll passes through `roller.Roller`. There are no package-level RNG calls. A seed plus a sequence of options fully determines a system. This is the core invariant — anything that calls `math/rand` or `crypto/rand` outside `roller/` is a bug.

```text
dice → roller → stars → worlds → iiss → cmd/world-builder
```

- `dice/` — parses WBH dice notation (`"2D"`, `"2D-7"`, `"D3-1"`); returns `Spec{Count, Sides, Modifier}`.
- `roller/` — `Roller` interface with three impls: `Seeded` (production), `Scripted` (replays book values for worked-example tests; **panics on exhaustion** — that always indicates a test bug), `Fixed`.
- `stars/` — WBH pp. 14–35 (Stars chapter). **Stage 0.** Public entry: `stars.GenerateSystem(r, opts)`.
- `worlds/` — WBH pp. 36–146. **Stages 1–10.** `GenerateSystemPlacement` produces Stage 1; `Apply*` stages 2–10 walk a `*Universe` and write to bodies in place.
- `iiss/` — IISS form structs (Class 0/I, Class II/III, Class IV-P) and their renderers. Split out from `worlds/` so renderer concerns are separately replaceable.
- `cmd/world-builder/` — thin CLI wrapper; emits Markdown (default), JSON, or short profile. See _Output_ below.

The function-naming discipline (`Generate*`, `Roll*`, `Compute*`, `Derive*`, `Apply*`, `Render*`) is documented in detail at `docs/api-surface.md` § Naming. The short version: `Generate*` rolls a sub-system from a Roller; `Apply*` mutates `*Universe`; `Compute*`/`Derive*`/`Roll*` return values without mutation. No package-level state.

### `worlds` package — Stage pipeline

`Generate(seed)` (and `GenerateWithRoller(r)` for tests) is the top-level entry. It builds a `Universe` from `stars.GenerateSystem` + `GenerateSystemPlacement`, then walks the `Apply*` stages in fixed order. See `worlds/generate.go`:

```text
Stage 0    stars.GenerateSystem                          pp. 14–35
Stage 1    GenerateSystemPlacement                       pp. 36–68
Stage 2    ApplyDetailFrontEnd                           sizing, moons, designations, periods
Stage 3    ApplyBodyPhysical + ApplyBeltDetails          composition/density/gravity/mass; belt details
Stage 4    ApplyMoonRefinement + ApplyRotationTilt       rotation, axial tilt, surface tidal effects
Stage 5    ApplyClimate                                  temperature, atmosphere, hydrographics
Stage 5'   ApplyTidalLockReEval                          tidal-lock cascade per WBH p.106
Stage 6    ApplyTaintTypology + ApplySurfaceDistribution
Stage 7    ApplyGeology                                  pp. 125–127
Stage 8    ApplyBiology                                  pp. 127–131
Stage 9    ApplyHabitability                             pp. 132–138
Stage 10   AggregateSystem + BuildIISSForms              per-allocation BaselineN, profiles, mainworld pick, the IISS Class IV Survey
```

Orchestrators live in role-named files: cross-cutting passes get their own (`detail_frontend.go`, `physical_detail.go`, `rotation_tilt.go`, `climate.go`, `taint_surface.go`, `aggregate.go`), and the single-feature passes live with their procedures (`geology.go`, `biology.go`, `habitability.go` each hold both the `Apply*` orchestrator and the feature's `Roll*`/`Compute*` procedures). Other per-feature files (`atmosphere.go`, `tidal_lock.go`, …) hold procedures only.

Bodies carry nullable pointer fields (`*Atmosphere`, `*Geology`, `*Biology`, `*Habitability`, …). Pointer = nil means "not applicable to this body type". Use the `Has*()` predicates — don't deref blindly.

Bodies are walked in ascending-orbit order within each group; `LongProfile` and `AssignPlanetDesignations` rely on this. Preserve the ordering when modifying the pipeline.

### Moons mirror planets

Per WBH, moons run through nearly the same physical pipeline as their parent planet. The implementation seam is `Universe.AllBodies()` — a single unified iterator that yields every body (planet or moon, including gas-giant moons) exactly once, with `body.Host()` supplying the parent for moons and the body itself for planets so HZ inheritance flows uniformly. A new stage that calls `for body := range u.AllBodies()` is moon-correct by construction.

**Anti-pattern alert:** historically, `Apply*` stage additions have added planet logic without iterating moons, producing silent-zero bugs. See `docs/anti-patterns.md` § A.1. When adding a new stage, verify moons are visited (including moons of gas-giant parents).

### Tables as Go literals

Heterogeneous WBH tables live as typed Go literals in `stars/tables.go` and per-feature files under `worlds/`. Conventions:

- `*float64` for cells the book leaves blank (the book's "—").
- Doc-comment every table with its WBH page and table name.
- Coarse subtype grids (O0, O5, B0, …) interpolate via a shared helper; the book endorses interpolation ("a G7 is 2/5 of the difference between G5 and K0").
- No external YAML/TOML data files. Tables compile in.

## Testing strategy — fidelity to the book

Three concentric layers:

1. **Worked-example regression tests** (`*_worked_examples_test.go`, `Test*_p<page>`) — the book threads in-line examples like the G7 V star "Zed" through every step. Encode each example with `roller.NewScripted(...)` driven by the exact dice the book describes; assert every output matches the book to the digit. **These tests are the proof of fidelity.** Breaking one breaks `go test`.
2. **Table integrity tests** — completeness, monotonicity claims, column totals.
3. **Property tests** — for arbitrary seeds, generated values fall in valid ranges.

When the book is inconsistent, the test asserts the implementation's chosen interpretation and the comment cites the divergent source. Don't paper over book inconsistencies.

## Conventions

- Module path: `github.com/philoserf/world-builder`. Imports: `"github.com/philoserf/world-builder/stars"`, `"github.com/philoserf/world-builder/worlds"`, etc.
- Go version: see `go.mod`. Modernizer hints (`go fix`) reflect Go 1.21+/1.22+ idioms (`min`/`max`, range-over-int, `new(value)`) and are enforced.
- Doc-comments cite WBH page numbers next to procedures and tables — that's the project's traceability mechanism (no runtime metadata wrapper).
- The library is the artifact; `cmd/world-builder` is one screen of code and stays that way.
- No CI. Local `task` (or `task check && task test`) is the gate.

## Scope

The project is **done** when:

1. The book's physical star-system rules (WBH pp. 14–146) are encoded faithfully in code, and
2. `cmd/world-builder` emits a complete description of the generated system as Markdown — a single IISS Class IV Survey document (PART 1 system census + a per-body PART P / PART P.B for every notable world).

WBH pp. 147–234 (World Social Characteristics, Special Circumstances) are **out of scope** for current and near-term purposes. Do not start work in those chapters; do not add code that anticipates them.

The rules half is complete on `main`: Stars (pp. 14–35), System Worlds and Orbits (pp. 36–68), and the full World Physical chapter (pp. 69–146) — including 3B-final habitability, mainworld pick, and the IISS Class IV Survey (PART 1 census + per-body PART P/P.B). Evergreen design + reference docs live at `docs/` root (api-surface, anti-patterns, design-intent, dependency-graph, harness, next-steps, summary, wbh-inconsistencies). Historical artifacts — pass-1 specs/plans/retrospective and pass-2 plans/retrospective — live under `docs/history/`.

### Output

`cmd/world-builder -format markdown` is the default. Output is a single **IISS Class IV Survey** document to stdout: an H1 title, the Notable Features summary, then **PART 1 — System Census** (system scalars, stellar roster with full orbital data, and the body roster — this folds in what the Class 0/I and Class II/III "short forms" used to show), followed by a **PART P** (planet/moon/gas-giant) or **PART P.B** (belt) per non-empty body in orbit order. The auto-picked mainworld's part is suffixed "— mainworld". Class 0/I and Class II/III are no longer emitted as standalone forms (they were earlier survey stages); the `Class0IForm`/`Class23Form` structs survive only as PART 1's data carriers.

`-format json` and `-format short` remain available for tooling and at-a-glance use; a future webservice may expose JSON over HTTP. The library is the source of truth — the CLI is a thin renderer.

This project is Unix-only. Do not add Windows-conditional code or filesystem logic.
