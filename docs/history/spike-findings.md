# Pass 2 — Spike: Cycle-1 Pre-Flight Findings

**Spike goal.** Validate two claims from `design-intent.md` before cycle 1 begins:

1. Pass-1 dice scripts can port verbatim to pass-2's reordered pipeline.
2. Implementation-order steps 2 (fixtures red) and 3 (stub signatures) — fixtures can land before stubs.

**Worktree.** `.worktrees/spike-cycle1` on branch `spike/cycle1`. Pass-1 implementation is the in-tree reference. No new generation code was written; the spike is read-and-reason against existing source.

**Headline.** The "scripts are gold; ported verbatim" claim is correct for narrow per-procedure fixtures (Stages 0–1) and false for any full-pipeline fixture (`Sol/Generate`, `Zed/Generate`). Pass-1 itself proved the latter — its `TestZed_FullDetail` was abandoned mid-pass-1 when new pipeline passes added dice the gold script didn't provide. Pass-2's harness must reflect this.

## Finding 1: `Sol/Terra/p35` does not exercise dice at all

`stars/worked_examples_test.go:11` — `TestSolTerra_p35` builds Sol via `stars.Compose(ComposeOpts{...})` with explicit `Mass: 1.000, Diameter: 1.000, Temperature: 5772, AgeGyr: 4.568`. There is no `Roller`. The book gives Sol's values directly on p.35; nothing is rolled. Same applies to `TestSolTerra_SurveyForm_p35`, `TestZed_SurveyForm_p34`, and `TestCorella_SurveyForm_p35` — all build via `Compose` and `OrbitPeriodYears`.

**Implication.** The harness entries for `Sol/Terra/p35`, `Sol/SurveyForm`, `Corella/SurveyForm`, `Zed/SurveyForm` (catalog rows in `harness.md` § Stars) are not "dice-scripted fixtures." They're compose-and-assert fixtures. The "port dice scripts verbatim" rule does not apply to them. They port trivially because pass-2 keeps `stars.Compose` and `BuildSurveyForm` verbatim.

The only Stage-0 fixture that exercises the Roller is `TestZedPrimaryOnly_p17_p21` (7 rolls scripted from the book's narration: type=9, subtype=6, mass-variance=2, diameter-variance=1, age={3,2,3}). Pass-2 keeps stars/ verbatim, so this one ports as-is.

## Finding 2: Pass-1 abandoned its full-pipeline gold script

`worlds/worked_examples_test.go:587` — `TestZed_FullDetail` is **`t.Skip()`**:

> 3A1 added six new pipeline passes to DetailSystem (body physical, belt details, atmosphere, hydrographics, moon refinement) that consume additional dice not present in `composeZedDetailScript`. Task 15 of the 3A1 plan replaces this with `TestZed_FullDetail_3A1` using a free-dice (Seeded) roller and shape-only assertions.

The successor `TestZed_FullDetail_3A2b` (`worlds/worked_examples_test.go:917`) drives 100 iterations of `roller.NewSeeded(seed)` and asserts shape-only invariants:

- Every HZ-orbit terrestrial has a 3-char SAH (no `?`).
- Every body has DayLength + AxialTilt + TidalEffects.
- Every HZ terrestrial with hydrographics has SurfaceDistribution.
- TidalLock case is valid when present.
- The survey form rendered without error.

This is a property test, not a worked-example regression. There is no end-to-end gold script.

**Implication.** Pass-2's `harness.md` § Façade end-to-end currently describes:

| ID             | Status | Asserts                                                                          |
| -------------- | :----: | -------------------------------------------------------------------------------- |
| `Sol/Generate` |   🔴   | `GenerateWithRoller(scriptedSol)` returns Universe with non-empty Stars, ...     |
| `Zed/Generate` |   🔴   | `GenerateWithRoller(scriptedZed)` returns Universe matching the per-stage Zed... |

The `scriptedSol` / `scriptedZed` framing is wrong. Pass-1 already learned that a full-pipeline `Scripted` roller cannot survive a pipeline change — and pass-2's pipeline is a pipeline change. Two corrections:

- Rename to `seededSol` / `seededZed` (or just `Sol/Smoke`, `Zed/Smoke`).
- Assert shape-only invariants (mainworld picked, IISS forms non-empty, every HZ terrestrial has SAH-without-`?`, etc.). Mirror `TestZed_FullDetail_3A2b`'s assertion catalog.

The phrase "matching the per-stage Zed assertions composed end-to-end" cannot be honored. You cannot concatenate per-stage Scripted dice scripts and feed them through the pipeline because the pipeline order determines the consumption order, and any reorder breaks the concatenation. Pass-1 hit this exact wall.

## Finding 3: Per-procedure gold scripts port verbatim — but only as long as procedure signatures don't change

Per-procedure tests are narrow:

- `TestZed_RollBaselineNumber` — 1 die.
- `TestZed_BaselineOrbit` — 1 die.
- `TestZed_PlaceOrbitSlots_Aab` — 10 dice (one per non-baseline slot variance).
- `TestZed_GenerateCounts` — small handful.
- `TestZed_AddAnomalous` — small handful.
- `TestZed_FullPlacement` — Stage-1-end-to-end with a longer concatenated script, but stays inside Stage 1's deterministic order.

Every test is `worlds.Procedure(roller.NewScripted(...), ...)` where the procedure's dice consumption is fixed by its body, not by the surrounding pipeline.

**Implication.** Pass-2 keeps Stage 0 and Stage 1 procedures verbatim per `dependency-graph.md`. These per-procedure scripts port as-is — they are gold. The harness entries for Stage 0 (`Zed/PrimaryOnly`) and Stage 1 (`Zed/Counts`, `Zed/BaselineN`, `Zed/BaselineOrbit`, `Zed/Spread`, `Zed/AddAnomalous`, `Zed/PlaceOrbitSlots_Aab`, `Zed/FullPlacement`) port verbatim.

Stage 2+ procedures may have different signatures in pass-2 (notably `ConvergeClimate` replacing 5A/5C/5D, `Apply*` operating on `*Body` instead of `*DetailedPlacement`). Their narrow per-procedure scripts can still be ported, but the test setup (constructing the input Body) changes. Mechanical port, not "verbatim."

`TestZed_FullPlacement` (a Stage-1-end-to-end concatenated script) is the most ambitious gold script that actually still works. It works because Stage 1's call order is fixed by `GenerateSystemPlacement`'s internal sequence and pass-2 doesn't reorder it. Stage 1's gold scripts are safe.

## Finding 4: Implementation-order steps 2 and 3 are reversed

`design-intent.md` § Implementation order:

```text
1. Port roller/ and dice/ verbatim from pass 1.
2. Port the worked-example dice scripts as failing fixtures (compiles, tests red).
3. Stub every public signature per api-surface.md (compiles, tests still red).
```

Go fixtures must compile. A test calling `worlds.GenerateSystemPlacement(r, sys)` cannot compile until `GenerateSystemPlacement` exists as a stub. Step 2 cannot precede step 3.

**Trivial fix.** Swap the order:

```text
1. Port roller/ and dice/ verbatim from pass 1.
2. Stub every public signature per api-surface.md (compiles).
3. Port the worked-example dice scripts as failing fixtures (compiles, tests red).
```

Or: collapse 2 and 3 into a single "stub commit + harness commit" pair. The stub commit's PR-sized scope (api-surface.md § Stub commitment scope) is fine; the harness commit lands immediately after.

## Finding 5: Cadence math doesn't close

`harness.md` lists ~50 worked-example fixtures, ~14 misuse-path entries, and ~5 property tests. At "1–3 fixtures green per cycle, 3–7 days each" (`design-intent.md` § Cadence), that is 17–50 cycles and 8–50 weeks. The "soft cap: revisit at week 3 if parity is not in sight" is wildly optimistic.

**Two paths:**

- **Bundle harder.** A cycle = one stage's fixtures (10–15 fixtures), 3–7 days. That gives ~10 cycles (30–70 days). Cap fits at week 8–10.
- **Accept the timeline.** Pass-2 is genuinely a 2–3 month build. Set the soft cap at week 8 and the hard "is this still worth it?" review at week 12.

I prefer bundling harder. The fixture catalog is naturally stage-grouped; pass-1's sub-projects were stage-sized; the dependency graph is stage-sized. Cycle = stage is the natural unit.

## Finding 6: Risks not in `design-intent.md`

Two risks the spike surfaced that aren't named in `design-intent.md` § Risks named:

- **`ConvergeClimate` panic-on-overflow on production seeds.** `api-surface.md` says "Cap N = 3 iterations. Asserts convergence; panics on overflow (a fixture failure)." For `Scripted` rollers in test, panic is right (a fixture bug). For `Seeded` rollers running arbitrary referee seeds, a non-converging body crashes `cmd/wbh`. Decision needed: cap higher with a soft warning, panic-only-in-test (introspect the Roller type), or prove formal convergence in N=3. None of these land for free.
- **Per-procedure scripts may need re-derivation when signatures change.** Pass-2 reshapes signatures: `*DetailedPlacement` → `*Body`, `Moon` → `Body{Kind: BodyMoon}`, climate procedures consolidate into `ConvergeClimate`. The per-procedure dice scripts survive, but the test scaffolding (input construction, output assertions) does not. "Verbatim port" is wrong for Stage 2+ tests; the right word is "mechanical port" — same dice, different scaffolding.

## Recommendations to apply before cycle 1

In rough order of urgency:

1. **`design-intent.md` § Implementation order** — swap steps 2 and 3 (stubs precede fixtures).
2. **`harness.md` § Façade end-to-end** — replace `scriptedSol` / `scriptedZed` with `seededSol` / `seededZed`; assertion catalog mirrors `TestZed_FullDetail_3A2b` shape invariants.
3. **`design-intent.md` § Cadence and § Risks named** — bundle to "cycle = stage" sizing; soft cap at week 8 not week 3.
4. **`design-intent.md` § What pass 1 got right** — the "Worked-example regression tests" entry must distinguish _per-procedure_ gold scripts (gold and ported verbatim) from _full-pipeline_ gold scripts (don't exist; pass-1 abandoned them; pass-2 uses Seeded + shape invariants for façade).
5. **`api-surface.md` § ConvergeClimate** — decide panic-vs-error stance for production `Seeded` rollers before stubs commit. Recommend: cap N=10, log-and-degrade-on-overflow for `Seeded`, panic-on-overflow for `Scripted` (introspect via type assertion or by passing a `convergenceMode` flag on the Roller construction).
6. **`api-surface.md` § The Body / Stage 2+ tests** — name the "mechanical port" reality for Stage-2+ test scaffolding. Per-procedure scripts survive; wrappers don't.

## What this spike did not do

- **Did not gut pass-2 down to greenfield.** That's cycle-1 work, not spike work.
- **Did not write any pass-2 stubs.** All findings are reachable from existing source + design docs.
- **Did not run a forced reorder experiment.** Pass-1's own `TestZed_FullDetail` skip-and-supersede event is sufficient empirical evidence; manufacturing a synthetic reorder would prove the same point at higher cost.

## Disposition

The spike branch (`spike/cycle1`) holds this findings document and the `.gitignore` update for `.worktrees/`. The findings doc moves to `pass-2` (or `main`) for visibility; the branch can be deleted once findings are merged.

Cycle 1 cannot start cleanly until findings 1–5 are reflected in the design docs. After that, cycle 1 = "port `roller/` and `dice/` verbatim; stub every public signature per `api-surface.md`; harness fixtures land red against stubs." That is one cycle's worth of work, ~3–7 days.
