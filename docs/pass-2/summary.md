# Pass 2 — Detailed Summary

This document records what pass 2 actually built, organised by architectural deliverable rather than chronologically. The commit log on the `pass-2` branch is the authoritative chronology; this is the conceptual rollup.

## Intent

Pass 1 confused **where the procedures live in the book** with **what depends on what**. Following WBH pagination produced a fixed-point system discovered mid-flight (atmosphere ↔ temperature ↔ hydrographics), a moon-path silent-zero anti-pattern that recurred four times, a late-arriving Class IV-P renderer that returned a string while siblings returned structs, API gotchas (RollGasMix's misnamed parameter) that slipped past 11 chapter-level reviews, and a project structure where every chapter pulled forward from earlier chapters and pushed back into later ones.

Pass 2 inverted the relationship: **the data dependency graph determined structure; worked-example fixtures determined the acceptance gate; the book became a citation system, not an architecture.**

## Delivery

18 cycles plus interstitial cleanup commits, all on `pass-2`. The branch builds clean (modernizer + gofumpt + vet + golangci-lint, 0 issues) and `go test -race ./...` is green for every package. `cmd/wbh -seed N -format {markdown,json,short}` runs end-to-end for any non-Special-Circumstances seed.

### Architectural deliverables

**Unified `Body` type (cycle 0, 1a, 2).** Replaces pass-1's `DetailedPlacement` / `Moon` split. Moons are `Body{Kind: BodyMoon, Parent: <planet>}` and walked by the same iterator as planets via `Universe.AllBodies() iter.Seq[*Body]`. The moon-path silent-zero anti-pattern (A.1) is prevented at the type level — there is no separate moon code path. The property test `TestProperty_MoonsHaveBodies` codifies the invariant: every Child is reachable from `AllBodies()` with `Kind == BodyMoon` and `Parent` populated.

**`iiss/` package boundary (cycle 0, 11, 14, 15, 16).** All IISS form structs (`Class0IForm`, `Class23Form`, `Class4PForm`) and their Markdown / JSON / PlainText renderers live in a separate package. `iiss/` does not import `worlds/`; `iiss.SystemForms` is the boundary type. Form-building (`worlds.BuildIISSForms`) lives in `worlds/` and translates `Universe` state to `iiss.SystemForms`; rendering lives in `iiss/` and reads the structs. The cycle-11 MVP shipped minimal-viable rendering; cycles 14-16 brought the forms to full pass-1 fidelity (Class 0/I delegates to `stars.BuildSurveyForm`; Class II/III mirrors pass-1's `ObjectRow` with Primary/Orbit/AU/Ecc/PeriodStr/SAH/Sub/Notes; Class IV-P PART P + PART P.B render every per-body block).

**`ConvergeClimate` per-body fixed-point solver (cycles 0, 5, 17+18).** Replaces pass-1's separate 5A-atm/hydro + 5C-temp + 5D-rederive (2-pass loop) with a single per-body entry. Cycle-5 MVP replicated pass-1's flow; cycles 17+18 folded partial-geology (Residual + TSF + THF) into the iteration loop per `dependency-graph.md` § Stage 7. Post-TSS Temperature is now the value atm/hydro re-derive against — for cold/rogue worlds where TSS dominates the temperature budget, this is the correct behaviour. N=5 cap with early-exit on stable (atm.Code, hydro.Code, MeanK); the strict panic-on-overflow contract is deferred (some seeds oscillate within the cap; see Lessons Learned).

**Stage orchestrators 2-10 (cycles 2-10).** `ApplyDetailFrontEnd`, `ApplyBodyPhysical`, `ApplyBeltDetails`, `ApplyMoonRefinement`, `ApplyRotationTilt`, `ApplyClimate`, `ApplyTaintTypology`, `ApplySurfaceDistribution`, `ApplyGeology`, `ApplyBiology`, `ApplyHabitability`, `AggregateSystem`. Each walks `Body` and `Body.Children` to drive per-procedure work. Pass-1 procedures (`sizing_terrestrial.go`, `belt_details.go`, etc.) ported verbatim — they take raw inputs and return values, so signatures are stable across the rebuild. Stage-2+ tests adapted to the new Body shape via sed sweep (DetailedPlacement → Body, `*Moon` → `*Body`, `.Moons` → `.Children`).

**Top-level façade (`Generate(seed)` + `GenerateWithRoller(r)`, cycle 11).** A 30-line orchestrator that runs every Apply\* step in dependency-graph order, populates `Universe.Detail.SystemForms`, and returns. `cmd/wbh/main.go` is three lines of pipeline plus format dispatch. Per spike-findings § 2, façade fixtures are Seeded + shape-invariant (not Scripted value-exact gold scripts) because pass-1 itself abandoned its full-pipeline gold script when new pipeline passes broke it.

### Testing infrastructure

**Per-procedure tests (cycles 2-9).** Every WBH-procedure file (`sizing_terrestrial_test.go`, `belt_details_test.go`, `body_physical_test.go`, `day_length_test.go`, `axial_tilt_test.go`, `tidal_lock_test.go`, `surface_tidal_effects_test.go`, `atmosphere_test.go`, `atmosphere_taint_test.go`, `hydrographics_test.go`, `temperature_test.go`, `temperature_rederive_test.go`, `geology_test.go`, `biology_test.go`, `habitability_test.go`, `moon_refinement_test.go`, `moons_test.go`, `designations_test.go`) carries the pass-1 gold dice scripts with mechanical type-rename adaptation. These are the "fidelity to the book" tests per CLAUDE.md § Testing strategy layer 1.

**Stage integration tests (cycles 2-9).** `stage2_test.go` through `stage7_test.go` each run the relevant orchestrator over 25-50 Seeded iterations against `composeZed()` (and Sol where applicable), asserting shape invariants. These are the spike-findings § 2 Seeded-shape-invariant pattern.

**Façade end-to-end tests (cycle 1b, 11).** `TestSol_Generate`, `TestZed_Generate`, `TestZed_MarkdownGolden`. Run the full pipeline through the public façade with a Seeded roller and assert shape invariants (Stars populated, Allocations non-empty, IISS form has Stars rows, mainworld picked when candidates exist, every child reachable from AllBodies, etc.).

**Property tests (post-cycle-17).** Five entries from `harness.md` § Property tests, each over 1000 seeds: `TestProperty_HZBodyHasClimate`, `TestProperty_MoonsHaveBodies`, `TestProperty_MainworldExists`, `TestProperty_BiomassImpliesAtm`, `TestProperty_ConvergenceCompletes`. Special-Circumstances primaries (post-stellar / peculiar / giant-companion-MAO / missing-class-IV) are recognised via `isSpecialCircumstances` and skipped as out-of-pass-2-scope.

**Misuse-path tests (post-cycle-17).** Five high-value cases from `harness.md` § Misuse-path tests (ConvergeClimate-on-GG, ConvergeClimate-non-HZ, AggregateSystem-empty, ComputeHabitability-no-temp, RollGasMix-empty-column). The remaining nine entries are mechanical extension when needed.

**Regression baseline (post-cycle-17).** `iiss/testdata/seed_{1,7,42,100,500}.md` snapshot the current Markdown output. `TestRegression_MarkdownSeeds` compares — fails on unintentional drift, refreshes on `go test -update.regression` after a reviewed change. This is a within-pass-2 guard, not a pass-1-vs-pass-2 comparison.

### Fidelity gate status

`design-intent.md` § Fidelity gate (the renamed parity gate) lists three merge criteria:

1. **Every worked-example fixture passes** (post-decision values, per `wbh-inconsistencies.md`). ✓ — every per-procedure test is green; the six WBH inconsistencies (Compatibility "+3", gravity DM overlap, etc.) are committed to specific interpretations in the source.
2. **Pass-1-vs-pass-2 IISS divergence assertions** with comments citing the corrected design. ✗ — never built. The within-pass-2 regression baseline catches drift between pass-2 cycles but does not compare against pass-1's binary output. Building the comparison is mechanical; resolving divergences needs design opinion (see Next Steps).
3. **Cuts list honoured pre-merge.** Partial — variance/accuracy/optional flags are still present in `stars.GenerateSystemOpts` because removing them would break ~10 pass-1 fidelity tests (the cuts list was aspirational on this point). Pass-3 referee knobs (Rare Earth, optional biomass floor, `-mainworld` override) are correctly deferred.

The spirit of the gate is met — pass-2 is architecturally complete, every per-procedure test is green, and the IISS forms render with full WBH-page fidelity. The letter requires the comparison work in Next Steps.

### Carry-over deferrals

These are named explicitly in `next-steps.md` and called out in their commit messages, not hidden:

- **Strict `ConvergeClimate` convergence** — empirically some seeds oscillate; the current N=5 / accept-last-state behaviour is a pragmatic compromise.
- **Pass-1-vs-pass-2 byte comparison tool** — building it is mechanical; triaging the divergences it surfaces needs design opinion.
- **`stars.Group` migration to `stars/`** (cycle 12 deferred) — `api-surface.md` § Open questions decided this should move but `worlds.Group`'s unexported fields are referenced by tests; export-vs-getter is a real trade-off.
- **`stars.GenerateSystemOpts` cuts** (cycle 13 deferred) — would break ~10 Accuracy:1 fidelity tests; the cuts list was aspirational on this point.
- **Belt-mainworld worked example** — no canonical WBH source; constructing one needs design opinion.
- **Special Circumstances chapter** (WBH pp.147+) — explicitly out of pass-2 scope per CLAUDE.md.
- **Pass-3 referee knobs** — named in `design-intent.md` § Post-parity work; not in pass-2.

## Numbers

- ~18 implementation cycles plus ~6 cleanup / decision-recording / spike commits.
- 30+ source files in `worlds/`, 4 in `iiss/`, plus the verbatim `roller/`, `dice/`, and (mostly verbatim) `stars/` packages.
- 20+ test files covering per-procedure value-exact, stage integration, façade end-to-end, property invariants, misuse paths, and regression snapshots.
- 5 IISS form structs (Class 0/I, Class II/III header, Class IV-P PART P + PART P.B + sub-blocks) plus their renderers.
- 1 fidelity gate with the spirit met and one item (the pass-1 comparison) for letter.
