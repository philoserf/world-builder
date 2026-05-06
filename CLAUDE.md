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
go run ./cmd/wbh -seed 42 -format short          # run the CLI
```

`task check` runs `go fix ./...` first (modernizer pass) and **fails if it produces any diff** — modernizer hints are mandatory, not advisory. If `task check` complains, review the diff with `git diff` and commit it before continuing.

`gofumpt` is enforced via the CLI (not via golangci-lint's bundled copy) because the two disagreed on import grouping. Use `gofumpt -l -w -extra .` for formatting; do not enable golangci-lint's gofumpt linter.

## Architecture

### Pure-function pipeline with deterministic dice

Every dice roll passes through `roller.Roller`. There are no package-level RNG calls. A seed plus a sequence of options fully determines a system. This is the core invariant — anything that calls `math/rand` or `crypto/rand` outside `roller/` is a bug.

```text
dice → roller → stars → worlds → cmd/wbh
```

- `dice/` — parses WBH dice notation (`"2D"`, `"2D-7"`, `"D3-1"`); returns `Spec{Count, Sides, Modifier}`.
- `roller/` — `Roller` interface with three impls: `Seeded` (production), `Scripted` (replays book values for worked-example tests; **panics on exhaustion** — that always indicates a test bug), `Fixed`.
- `stars/` — WBH pp. 14–35 (Stars chapter). Public entry: `stars.GenerateSystem(r, opts)`.
- `worlds/` — WBH pp. 36–146. Layered façades: `worlds.SystemPlacement` → `worlds.SystemDetail`.
- `cmd/wbh/` — thin CLI wrapper; emits JSON IISS Survey form or short profile.

Each `Generate*` takes upstream results plus a `Roller`, returns immutable value types. No package-level state. Variance and accuracy options live on per-call `*Opts` structs, off by default.

### `worlds` package — DetailSystem pipeline

`DetailSystem` orchestrates the per-body procedures in `runDetailPipeline` (extracted from `DetailSystem` for testability):

```text
Steps 1–4   sizing → moons → designations → periods
Step 5      HZ tagging
Steps 5A–5G 3A1 (body-physical) → 3A2a (rotation/tilt/tidal)
            → 3A2b-temp (temperature) → 3A2b-rederive (atmosphere/hydro)
            → 3B-geology → 3B-biology → 3B-final (habitability)
Step 6      backfill StarAllocation.BaselineN
Step 7      Short/Long profiles
Step 8      RenderIISSClass23 (the survey form)
            pickMainworld
```

`DetailedPlacement` embeds `Placement` (2B) and adds nullable pointer fields (`*Atmosphere`, `*Geology`, `*Biology`, `*Habitability`, …). Pointer = nil means "not applicable to this body type". `Has*()` predicates wrap the nil checks. Use them — don't deref blindly.

Bodies are walked in ascending-orbit order within each group; `LongProfile` and `AssignPlanetDesignations` rely on this. Preserve the ordering when modifying the pipeline.

### Moons mirror planets

Per WBH, moons run through nearly the same physical pipeline as their parent planet. Implementation pattern: `buildMoonPlacementView(m, parent)` synthesizes a `*DetailedPlacement` from a `Moon` so per-body procedures (sizing, atmosphere, etc.) reuse one code path.

**Anti-pattern alert:** historically, `runStep5*` additions have added planet logic without iterating moons, producing silent-zero bugs. When adding a new step, verify moons are visited (including moons of gas-giant parents).

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

- Module path: `wbh` (local-only). Imports: `"wbh/stars"`, `"wbh/worlds"`, etc.
- Go 1.26.2 (`go.mod`). Modernizer hints (`go fix`) reflect Go 1.21+/1.22+ idioms (`min`/`max`, range-over-int, `new(value)`) and are enforced.
- Doc-comments cite WBH page numbers next to procedures and tables — that's the project's traceability mechanism (no runtime metadata wrapper).
- The library is the artifact; `cmd/wbh` is one screen of code and stays that way.
- No CI. Local `task` (or `task check && task test`) is the gate.

## Scope

The project is **done** when:

1. The book's physical star-system rules (WBH pp. 14–146) are encoded faithfully in code, and
2. `cmd/wbh` emits a complete description of the generated system as Markdown — all three IISS forms (Class 0/I, Class II/III, Class IV-P) covering whatever world type the mainworld turns out to be (planet, moon, or belt).

WBH pp. 147–234 (World Social Characteristics, Special Circumstances) are **out of scope** for current and near-term purposes. Do not start work in those chapters; do not add code that anticipates them.

The rules half is essentially complete on `main`: Stars (pp. 14–35), System Worlds and Orbits (pp. 36–68), and the full World Physical chapter (pp. 69–146) including 3B-final habitability, mainworld pick, and the IISS Class IV-P form. The one rules gap that remains is **Form 0407K-IV PART P.B** (belt-mainworld Class IV-P variant) — back in scope so Class IV-P works for every mainworld type. Specs/plans live in `docs/specs/` and `docs/plans/`, dated and named for the WBH section they cover.

### Output

`cmd/wbh -format markdown` is the default. Output is the full system as Markdown to stdout — concatenated IISS Class 0/I, Class II/III, and Class IV-P forms under H1/H2 section headings, in book order. Class IV-P renders **only for the auto-picked mainworld** (per book — not for every notable body), using whichever variant matches the mainworld's type.

`-format json` and `-format short` remain available for tooling and at-a-glance use; a future webservice may expose JSON over HTTP. The library is the source of truth — the CLI is a thin renderer.

This project is Unix-only. Do not add Windows-conditional code or filesystem logic.
