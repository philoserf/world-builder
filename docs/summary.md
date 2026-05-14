# Project Summary

A one-page overview of what this project is and how it's organised. For deeper context see `design-intent.md` (why the architecture looks this way), `api-surface.md` (public API), and `dependency-graph.md` (data dependencies).

## What this is

`go run ./cmd/world-builder -seed N -format {markdown,json,short}` generates a complete Mongoose Traveller _World Builder's Handbook_ star system end-to-end. Default Markdown output emits all three IISS forms (Class 0/I, Class II/III, Class IV-P) plus a referee-facing Notable Features summary. Seed determinism preserved; every invocation produces a real, fully-formed system.

WBH pp.14–146 (Stars + System Worlds and Orbits + World Physical Characteristics) are implemented to book fidelity. WBH pp.147+ (World Social Characteristics, Special Circumstances) are out of scope.

## Architectural shape

**Unified `Body` type.** Moons are `Body{Kind: BodyMoon, Parent: <planet>}` walked by the same iterator (`Universe.AllBodies() iter.Seq[*Body]`) as planets. The moon-path silent-zero anti-pattern is prevented at the type level — there is no separate moon code path.

**`iiss/` package boundary.** All IISS form structs (`Class0IForm`, `Class23Form`, `Class4PForm`) and their Markdown renderers live in a separate package. `iiss/` does not import `worlds/`; `iiss.SystemForms` is the boundary type. Form-building (`worlds.BuildIISSForms`) lives in `worlds/`; rendering lives in `iiss/`.

**`ApplyClimatePasses` per-body solver.** Folds partial geology (Residual + TSF + THF) into each rederive pass. Two passes, second is trusted — the climate cluster is not a fixed point (`RederiveAtmosphereHydrographics` consumes fresh dice per call).

**Stage orchestrators.** `ApplyDetailFrontEnd`, `ApplyBodyPhysical`, `ApplyBeltDetails`, `ApplyMoonRefinement`, `ApplyRotationTilt`, `ApplyClimate`, `ApplyTaintTypology`, `ApplySurfaceDistribution`, `ApplyGeology`, `ApplyBiology`, `ApplyHabitability`, `AggregateSystem`. Each walks `Body` + `Body.Children` to drive per-procedure work.

**Top-level façade.** `Generate(seed)` and `GenerateWithRoller(r)` run every Apply\* step in dependency-graph order, populate `Universe.Detail.SystemForms`, return. `cmd/world-builder/main.go` is three lines of pipeline plus format dispatch.

## Test coverage

Four layers (see `harness.md` for the full catalog):

- **Per-procedure tests.** Every `Roll*` / `Generate*` / `Compute*` function has at least one test that drives book-narrated dice and asserts the expected output. Bulk of coverage.
- **Named worked-example fixtures.** Where the book threads a multi-procedure example (Sol, Corella, Zed, Zed Prime), the fixture re-walks that chain end-to-end.
- **Property tests.** 8 invariants × 1000 seeds each (`worlds/property_test.go`). Smoke tests for systemic correctness that catch silent-zero / silent-skip bugs.
- **Markdown regression baseline.** 5 seeds × full Markdown output at `iiss/testdata/seed_*.md`. Refreshed with `go test ./iiss/... -update.regression -run TestRegression`.
- **Bulk-sweep verification.** 10 000-seed sweep produces 10 000 real systems with zero errors. See `history/generator-error-catalog.md`.

## Book-faithfulness levers

The project commits to specific interpretations of six WBH internal contradictions (`wbh-inconsistencies.md`), encodes them in code, and asserts the chosen values in tests. No runtime toggles for the inconsistencies.

For Referee-level book options (WBH p.15 column toggle, p.16 fallbacks), the project defaults to the simpler book-endorsed Special-column choice — every roll produces a real system. Users who want Unusual-column primaries (BD/D admitted, peculiar dispatch) can opt in via `GenerateSystemOpts.PeculiarColumn`.

## Out of scope

- WBH pp.147+ (World Social Characteristics + Special Circumstances). Explicitly out of scope per `CLAUDE.md`.
- Optional Referee knobs from `design-intent.md` § Post-parity work (Rare Earth variant, oxygen-atm biomass floor, Insidious DE optional branch, `-mainworld` override flag) — open polish, see `next-steps.md`.
