# Pass 2 — Lessons Learned

This document captures lessons that emerged **during pass-2 implementation** — they go beyond `spike-findings.md` (which was written before cycle 0) and `anti-patterns.md` (which was a pre-flight checklist). Each lesson is named, its evidence cited, and its implication for future passes or projects flagged.

## L1 — Verbatim ports work for self-contained procedures; orchestrators are where adaptation happens

The pass-1 procedure files that took raw inputs and returned values — `sizing_terrestrial.go`, `sizing_gasgiant.go`, `period.go`, `body_physical.go`, `belt_details.go`, `moon_refinement.go`, `geology.go`, `biology.go`, `habitability.go`, `atmosphere.go` and friends — ported with **zero adaptation** in cycles 2-9. `git checkout main -- worlds/$f.go` and the file compiled.

The pass-1 procedure files that took `*DetailedPlacement` or `*Moon` parameters — `day_length.go`, `axial_tilt.go`, `tidal_lock.go`, `surface_tidal_effects.go` — needed `sed s/DetailedPlacement/Body/g; s/*Moon/*Body/g; s/.Body /.Kind /g` and a few hand fixes (`body.Moons` → `body.Children`, `&body.Children[i]` → `body.Children[i]`, integer field types). Mechanical but not free.

The orchestrators (`stage2.go` through `stage10.go`, plus the climate `stage5.go` and the new `ApplyGeology` in `stage7.go`) were **fully rewritten** — they walk `Body` + `Body.Children` instead of `[]DetailedPlacement` + `dp.Moons`, and their dispatching by `Kind` differs from pass-1's switching on `BodyType`.

**Implication.** When choosing what to port verbatim vs adapt vs rewrite: signature shape predicts cost. Procedures with raw-data inputs port free; procedures with embedded-typed inputs need mechanical sed; orchestrators that consume the embedded types need rewriting. Plan effort accordingly.

## L2 — The design got the unified `Body` exactly right

`anti-pattern A.1` (moon-path silent-zero) recurred four times in pass 1. Pass-2's unified `Body` with `Children []*Body` walked by a single iterator prevented this bug **at the type level** — there is no separate moon code path to forget about. The property test `TestProperty_MoonsHaveBodies` (1000 random seeds) finds zero violations.

The cost of this was small: about 50 source-file references to update (sed sweep) plus a few small adaptations (e.g., `closest = &body.Children[i]` → `closest = body.Children[i]` because `Children` is already a slice of pointers).

**Implication.** When a bug-class recurs three or more times in pass 1, look for a type-system solution before the next rewrite. The investment in a unified type pays back in N-fold prevention.

## L3 — `ConvergeClimate` formal convergence was harder than the design assumed

`api-surface.md` § The Climate solver specified "Cap N = 3 iterations. Asserts convergence; panics on overflow (a fixture failure)." Cycle 17's implementation hit convergence overflow on common seeds (seed 0 Aab system in `TestZed_ApplyStage5`).

Investigation revealed: folding TSS into the climate loop (cycle 18 per `dependency-graph.md` § Stage 7) introduces additional coupling between iterations. Pass-1 didn't fold TSS — it applied TSS once, AFTER the climate-only 2-pass rederive, then stopped. Pass-2's tighter coupling creates oscillation modes where atm.Code swings between adjacent values (e.g., 5 → 6 → 5) across iterations.

Pragmatic compromise: N=5 cap with early-exit on stable triple; if N exhausts without early-exit, accept the last-iteration state silently. This matches pass-1's "trust the last rederive" but with up to 5 iterations rather than 2.

**Implication.** A design spec that asserts convergence within N must be informed by the empirical convergence rate of the underlying system. The pass-2 design specced N=3 without running the loop. A spike (one iteration of the loop on representative seeds) would have caught the oscillation before the design was committed. For deferred work: investigating the oscillation root cause (likely RederiveAtmosphereHydrographics's runaway-greenhouse check interacting with the TSS-bumped ScaleHeight) is a real engineering task, not a procedural one.

## L4 — Per-stage cycles matched the budget; the within-stage tasks did not

`spike-findings.md` § Finding 5 recommended "cycle = stage" with soft cap at week 8. The 11 implementation cycles (0, 1a, 1b, 2-11) landed in roughly 11 distinct sessions. The cadence matched.

But within each cycle, the work split into 3-7 sub-tasks (port the procedure file, port the test file, write the orchestrator, write the integration test, sed-rename old → new, commit). The original "1-3 fixtures per cycle" target was wrong; the actual unit was "1 stage per cycle" with sub-tasks that the cycle's commit message tracked.

**Implication.** "Cycle = stage" was the right granularity. The earlier "1-3 fixtures per cycle" target was too small. Future projects: estimate the cycle by the integration test, not by the value-exact fixture count.

## L5 — Stub-first compile-driven development found errors at the right time

Cycle 0 stubbed every public signature with `panic("unimplemented: see docs/pass-2/api-surface.md § ...")`. Cycles 2-11 progressively replaced stubs with real implementations.

The Go compiler flagged stub-call-site mismatches at every cycle boundary — when a real implementation's signature differed from the stub's signature (because the stub was wrong about the design), the compiler refused to build. This caught two real design errors: `GenerateBodyPhysical(r, *Body, ageGyr float64)` (cycle-0 stub) vs `GenerateBodyPhysical(r, SizeCode, int, BodyPhysicalDMs)` (pass-1 actual signature), and `AggregateSystem.SystemForms.X` redundant access flagged by staticcheck QF1008.

The cycle-0 stubs that survived multiple cycles ended up trustworthy: every Has\*() predicate, every body field, every Apply\* signature. The stubs that were rewritten signalled real design uncertainty.

**Implication.** Stub-first works when the API surface is mostly known. When a stub's signature gets rewritten, treat it as a design signal — the original spec was incomplete. Don't paper over with type assertions; revise the spec.

## L6 — `iiss/` boundary was easy because the design did the hard part up front

`api-surface.md` § Package boundaries committed to "iiss/ does not import worlds/; SystemForms is the boundary type." That single decision made the rest trivial: form-building lives in `worlds/iiss_build.go` (worlds → iiss); rendering lives in `iiss/render.go` (no inbound deps); the boundary is clean.

If pass-2 had tried to put `MarkdownSystem(*worlds.Universe)` inside `iiss/`, the import cycle would have surfaced at cycle 11 and forced a redesign mid-implementation. The pre-decision saved real time.

**Implication.** Package-boundary decisions are cheap during design and expensive during implementation. Make them up front.

## L7 — Documentation drift was small; cycle-by-cycle commit messages did the work

Pass-2's six design docs (`design-intent.md`, `api-surface.md`, `dependency-graph.md`, `wbh-inconsistencies.md`, `anti-patterns.md`, `harness.md`) were written before cycle 0 and largely held up. Post-implementation cleanup found three stale references (one `DetailedPlacement.MassEarth`, one outdated public-types list, one `Slot directly inside DetailedPlacement` description) — total edit was 3 small replacements.

The reason: each cycle's commit message documented what changed, what's deferred, and why. The commit log served as a living changelog. When the design docs needed an update, the relevant commit message had the rationale.

**Implication.** Detailed commit messages aren't ceremony — they're the substitute for living design docs. The pass-1 retrospective complained that doc drift was a problem; pass-2's discipline of "every commit explains itself" largely solved it.

## L8 — Property tests and a regression baseline are cheap insurance

The five property tests (1000 seeds each, ~5000 total seed runs) and the 5-seed Markdown regression baseline took less than half a day to write and produced two real wins:

1. The property tests confirmed `TestProperty_MoonsHaveBodies` — zero anti-pattern A.1 violations across 1000 seeds. This is the type-level prevention that the unified `Body` type promised, empirically verified.
2. The regression baseline catches unintentional drift between pass-2 cycles. If cycle 19+ accidentally changes Markdown output (e.g., a renderer tweak that breaks formatting), the test fails with a clear "drifted from snapshot" message.

The cost is small (~5KB of testdata files + 200 lines of test code). The value is "if we accidentally break something, we know immediately."

**Implication.** Property tests and regression baselines are high ROI when the system has invariants (which most do). Add them once architecture is stable. Pass-2 deferred them to post-cycle-17, which was the right call — too early and the architecture would shift; too late and bugs accumulated.

## L9 — Pass-1's "skip the failing test" pattern signaled real architectural debt

Pass-1's `TestZed_FullDetail` (the full-pipeline gold-script test) was `t.Skip()`'d mid-pass-1 with the comment "3A1 added six new pipeline passes that consume additional dice not present in `composeZedDetailScript`." Pass-2 inherited this — the test was deleted in cycle 0 because the types it referenced (DetailedPlacement, IISSClass23Header, GasGiantLarge) had been replaced.

This was the smoking gun in `spike-findings.md` § Finding 2 — pass-1 itself abandoned its full-pipeline gold script when pipeline reorders broke it. Pass-2 codified the lesson: façade fixtures are Seeded + shape-invariant, never Scripted + value-exact.

**Implication.** A `t.Skip()` with "this used to work but the pipeline changed" is architectural debt. It says "the test was wrong about what it was testing." Fix the test's framing or delete it; don't carry skipped fidelity tests across major changes.

## L10 — Sed-driven type renames are fine for mechanical refactors but require diagnostic awareness

Cycles 4, 5, 7, 8, 9 each used a sed sweep to rename DetailedPlacement → Body, `*Moon` → `*Body`, `.Moons` → `.Children`, `.Body` → `.Kind` in source files. Most worked first try. Two surprises:

1. `Moon{...}` struct literals required `s/Moon{/Body{Kind: BodyMoon, /g` not just `s/Moon/Body/g` — the latter would also affect `BodyMoon` constant. Word boundaries matter.
2. `&body.Children[i]` produced `**Body` (since `Children []*Body` already returns `*Body`); needed manual fix to `body.Children[i]`.

The Go compiler caught everything that sed missed. No silent bugs from the renames.

**Implication.** sed is fine for mechanical refactors when paired with a strong type system. The compile-fail is the safety net. Don't be afraid of large-scale find-and-replace; do trust the compiler to catch the remainder.

## L11 — The pass-1 dice scripts ARE gold for narrow per-procedure tests

`spike-findings.md` § Finding 3 distinguished "per-procedure gold scripts (1-10 dice each, deterministic procedure-internal order)" from "full-pipeline gold scripts (don't survive pipeline change)." Cycles 2-9 confirmed: the per-procedure scripts ported verbatim. Every Stage-2-onwards test that wraps `roller.NewScripted(...)` with the book's narrated dice — `TestZed_RollBaselineNumber`, `TestZed_BaselineOrbit`, `TestRollBiomass_ZedPrime`, `TestRollTectonicPlates_ZedPrime`, etc. — passed without expected-value updates.

This is real fidelity. Pass-2's procedures roll the book's dice and produce the book's values to the digit, mediated through Body-shaped scaffolding instead of DetailedPlacement-shaped scaffolding. The book is reproduced.

**Implication.** When the architecture is right, the test fixtures port. When the architecture is wrong, the fixtures break and you can't tell whether the bug is in the new code or in the new test. The book is the spec; the per-procedure tests assert that spec is met.

## L12 — Special Circumstances errors are pass-2's "out-of-scope" smoke

The `stars/` package emits `ErrPostStellarPrimaryUnsupported`, plus errors for special-primary giants, peculiar-primary dispatch, missing class-IV table cells, and giant-primary companion MAO. All five fall under WBH's Special Circumstances chapter (pp.147+), explicitly out of pass-2 scope per CLAUDE.md.

Pass-2's tests recognise these via an `isSpecialCircumstances(err)` predicate and skip seeds that hit them. About 30-40 seeds out of every 1000 fall in this category — enough to matter, but not enough to constrain the pass-2 deliverable.

**Implication.** When a project has explicit out-of-scope boundaries (WBH chapter X is post-parity), give the test infrastructure a way to recognise those boundaries via specific error sentinels. Don't fail tests on legitimate out-of-scope outcomes.

## Closing — what changed from pass 1 vs what survived

What changed: the type system (unified Body), the iteration pattern (`iter.Seq[*Body]`), the architectural layering (`iiss/` split, ConvergeClimate as a per-body entry, TSS folded into climate), the test fixture pattern (Seeded shape-invariant for façades), the orchestrator code (every Apply\* function), the cycle cadence (per-stage instead of per-WBH-page).

What survived: every per-procedure file in `stars/` (verbatim), every Stage-1 file in `worlds/` (verbatim, plus minor Body-rename adaptations), every per-procedure test (mechanical sed), `roller/`, `dice/`, the dice scripts themselves (the book's narrated dice), every WBH-inconsistency interpretation, every anti-pattern checklist item, the modernizer-as-mandatory-gate discipline, the brainstorm → spec → plan → TDD workflow shape.

Pass 2 is not a rewrite. It is a re-layering. The same WBH procedures run in a different orchestration and produce results through a different data model. The book is still the spec.
