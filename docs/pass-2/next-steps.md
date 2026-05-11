# Pass 2 — Next Steps

Three categories: items requiring user judgement (the trade-offs are real and pass-2-side automation can't resolve them); mechanical items (need labor, no design decision); strategic / scope items (merge planning, pass-3 outlook).

The ordering within each category roughly reflects priority — higher items unblock or de-risk lower ones.

## A. Items requiring your judgement

**A0 (dropped).** A pass-1-vs-pass-2 byte comparison was originally listed here as the unfinished fidelity-gate item. It is now explicitly **dropped**: pass-2's design intentionally diverges from pass-1 on multiple axes (TSS fold-into-climate, surface-distribution-after-converge, narrower-band-wins for gravity DM, etc.), and the within-pass-2 Markdown regression baseline (`iiss/testdata/seed_*.md` + `TestRegression_MarkdownSeeds`) is the working substitute. Future drift between pass-2 cycles is caught by the baseline; pass-1's outputs are no longer authoritative.

### A1. Strict `ConvergeClimate` convergence

**Status.** Cycle 17+18 ships N=5 with early-exit on stable triple; if N exhausts, accept the last-iteration state silently. `api-surface.md` specced N=3 with panic-on-overflow; that contract is deferred.

**Why your call.** Empirical testing showed some seeds oscillate between adjacent atm.Code values within N=5. The oscillation is probably real (RederiveAtmosphereHydrographics's runaway-greenhouse check interacting with TSS-bumped ScaleHeight), not a numerical artifact. Three options:

- **(a) Accept the current relaxed contract.** Document "ConvergeClimate may produce a non-converged last-iteration state for some seeds; this is by design." Lossy but stable.
- **(b) Investigate the oscillation root cause.** Likely involves instrumenting one oscillating seed, reading the per-iteration atm/hydro/temp values, identifying the bistable attractor, and either (i) breaking the tie via a deterministic rule (e.g., "on oscillation, prefer the higher-MeanK state"), or (ii) widening some threshold so the oscillation collapses.
- **(c) Punt convergence to a documented invariant.** Don't iterate at all — do exactly N=2 like pass-1 — and accept that the result isn't formally "converged." Pass-2 ships this in practice (early-exit rarely fires; we're effectively doing N=5 fixed iterations).

**Recommendation.** Option (b) is most rigorous and produces a real engineering insight. Option (a) is what's currently in code and is "good enough" for the IISS output. Decide based on appetite.

**Time estimate.** (a): 30 minutes (docstring update). (b): 1-2 days of investigation. (c): not real work; relabel current behaviour.

### A2. `stars.Group` migration to `stars/` package (cycle 12 deferred)

**Status.** `api-surface.md` § Open questions, decided says Group should live in `stars/`. Cycle 12 was deferred because `worlds.Group` has unexported fields (`companionEcc`, `sourceCompanion`) accessed by tests in the same package; moving requires either exporting them or threading getters.

**Why your call.** Three sub-decisions:

- **Export-vs-getter.** Renaming `companionEcc` → `CompanionEcc` and `sourceCompanion` → `SourceCompanion` is one line of edits but breaks encapsulation. Adding `Group.CompanionEcc()` getter methods preserves encapsulation but adds 4 method definitions.
- **Test refactor.** `worlds/available_orbits_test.go` accesses the unexported fields. Either export them or rewrite the tests to use the public API.
- **Worth it at all?** The migration is pure cleanliness — there's no functional benefit. The code already works. Cycle 12's deferral may be the right answer indefinitely.

**Recommendation.** Defer indefinitely unless `stars.Group` starts attracting unrelated coupling. The api-surface.md decision was right in principle; the cost in practice doesn't justify it for a working system.

### A3. `stars.GenerateSystemOpts` cuts (cycle 13 deferred)

**Status.** `design-intent.md` cuts list says drop `WithVariance` / `Accuracy` from `GenerateSystemOpts`. Cycle 13 deferred because ~10 pass-1 fidelity tests rely on `Accuracy: 1` to drive the simple-1D age path that the book's worked examples use.

**Why your call.** Cutting requires re-deriving expected values for those tests under `Accuracy: 2`. Two options:

- **(a) Cut the toggle; update the tests.** Tests assert new values; book fidelity is preserved at the production-default behaviour, not at the historical-test-fixture behaviour.
- **(b) Keep the toggle as load-bearing.** The cuts list was aspirational; calling it now is a design refinement.

**Recommendation.** Option (b). The toggle is small (two fields on one struct); removing it doesn't simplify anything meaningfully; the test-update work is real. Defer indefinitely.

### A4. Belt-mainworld worked example

**Status.** `harness.md` § Class4P/PartPB lists `ZedPrime/Class4P/PartPB` as 🚧 deferred. No canonical WBH example exists. Cycle 16 shipped a functioning PART P.B renderer with structurally-correct content; the missing piece is a fixture that asserts specific values.

**Why your call.** Constructing the fixture requires deciding what "canonical belt mainworld" looks like:

- Composition (m/s/c percentages)?
- Resource rating?
- Significant-body counts?
- The IISS form's expected output?

Without a book reference, this is design fiat. Pass-2 chose to defer rather than invent.

**Recommendation.** Defer until a real belt-mainworld scenario surfaces in a campaign. The PART P.B renderer works; a fixture is nice-to-have, not critical.

## B. Mechanical items (no judgement needed)

### B1. Remaining 9 misuse-path tests

Post-cycle-17 added 5 of the 14 entries in `harness.md` § Misuse-path tests. The remaining 9:

- `RollAtmoCode`: SizeCode "0", negative offset
- `RollTotalPressure`: atmCode outside table, Subtype required when code 11/12
- `RollOxygenFraction`: negative ageGyr
- `RollCorrosiveInsidiousSubtype`: atmCode not 11/12, HZCO ≤ 0
- `GenerateBodyPhysical`: SizeCode "S", DiameterKm ≤ 0, negative ageGyr
- `GenerateBeltDetails`: SizeCode not "0", negative ageGyr
- `GenerateHydrographics`: atm.Code 0 with non-degenerate inputs, tempRange invalid
- `RollBiomass`: body without atmosphere, negative ageGyr
- `RollCompatibility`: biocomplexity 0, atm code not in DM table
- `MarkdownClass0I` / `MarkdownClass23` / `MarkdownClass4P`: zero-value form asserts deterministic empty output

**Effort.** ~3-5 lines per test × ~25 misuse cases = ~100 lines of test code, ~1-2 hours.

**Value.** Low — most assert current behaviour rather than enforce a contract. The high-value ones (the 5 already done) were the ones that found real edge cases.

### B2. Harness.md status update

`harness.md` lists 50+ fixture entries all marked 🔴. Many are now green via per-procedure tests in `worlds/*_test.go` and `stars/worked_examples_test.go`. The catalog should reflect actual status.

**Effort.** ~30 minutes of crossing 🔴 to 🟢 in the markdown tables, with a note pointing to the per-procedure test file for each.

**Value.** Medium — `harness.md` is the project's fixture catalog; if it's stale, future developers won't trust it.

### B3. cmd/wbh JSON output

Cycle 11 ships `cmd/wbh -format json` that emits only the Class II/III form. The full Universe is more useful for tooling integration (a future webservice per CLAUDE.md § Output).

**Effort.** Add a top-level JSON struct that aggregates Class0I, Class23, Class4P, ShortProfile, LongProfile, MainworldDesignation. Maybe 30 lines.

**Value.** Low until a downstream consumer wants it. Defer.

### B4. Property test expansion

Five property tests run over 1000 seeds. Could add more:

- Every body with `Habitability.Rating > 0` has an Atmosphere with `Pressure > 0`.
- Every GG has `MassEarth > 0`.
- Every body with `Children` has at least one child with `OrbitPD > 0`.
- ScaleHeight is positive for every body with Atmosphere + Physical.

**Effort.** ~5 lines per test, ~20 minutes.

**Value.** Low — the existing five cover the high-risk invariants. Diminishing returns.

## C. Strategic / scope items

### C1. Merge `pass-2` to `main`

**Status.** Pass-2 is on its own branch. Main has pass-1. With A0 (pass-1 comparison) dropped, the gating concern is gone.

**Plan.** Merge `pass-2` to `main` when ready. The pass-2 branch is shippable: every test green, end-to-end CLI works, full IISS form fidelity, the within-pass-2 regression baseline guards drift. Tag `pre-pass-2` on `main`'s current HEAD before the merge so the historical pass-1 binary stays buildable from a known reference. Then either fast-forward `main` to `pass-2`'s HEAD or use `git merge --no-ff` for a marked merge commit; the choice is preference. Time: ~30 minutes including the tag, the merge, the push, and a sanity-run of `cmd/wbh` from `main`.

### C2. Pass-3 referee knobs

`design-intent.md` § Post-parity work names four referee-facing items deferred from pass-2:

- Rare Earth Universe Variant
- Optional any-oxygen-atm biomass floor
- Optional Insidious DE hazard rule's optional branch
- `-mainworld <designation>` override flag

These reintroduce the variance/accuracy/optional-rule toggles the cuts list removed. Each is a small individual change (an `Opts` field on the relevant `Generate*` or `cmd/wbh` flag) but together they're a coherent "give the referee back the knobs WBH ships with" cycle.

**Effort.** ~1 day total for all four, including tests and a `cmd/wbh -help` update.

**Value.** Medium — referee facing. Whether to do this depends on whether actual referees will use the tool.

### C3. Notable Features Markdown block

`design-intent.md` § Post-parity work also names "a referee-facing summary above the IISS forms: tidal-lock zones, WorstLow cold snaps, high-gravity/high-atm crush worlds, taint chains, mainworld habitability rationale."

The data is all in `Universe.Detail.Bodies`. A Markdown block that scans the universe and surfaces flagged conditions (e.g., "Aab IV a is tidally locked", "B I has WorstLow 89K — frostbite hazard") would significantly improve the referee UX.

**Effort.** Half a day to design the conditions + implement + render.

**Value.** High for actual play. Pass-2's IISS-only output is canon-correct but not at-a-glance useful.

### C4. Special-object detail (Brown Dwarf, White Dwarf, Neutron Star, etc.)

`design-intent.md` § Post-parity work also lists special-object detail. Currently these companion types get minimum-useful values (type/mass/age) but no detailed physics. Detailed physics (accretion, degenerate-matter equations, jet behaviour) is "post-parity but pre-pass-3."

**Effort.** Unknown — depends on which special objects to detail and to what depth. WBH's Special Circumstances chapter has the rules.

**Value.** Variable. Black-hole companions are dramatic; neutron-star companions cause real campaign effects. White dwarfs are common in older systems.

### C5. Special Circumstances chapter (WBH pp.147+)

Explicitly out of pass-2 scope per CLAUDE.md. Includes social characteristics, government, technology, special-object population physics, etc.

**Effort.** Huge — comparable to all of pass-1's physical-rules work.

**Value.** Depends on project goal. CLAUDE.md says "Do not start work in those chapters; do not add code that anticipates them." Pass-2 honours this.

## Suggested ordering

If the goal is "close pass-2 fully":

1. **C1** (merge to main) — pass-2 becomes the trunk.
2. **B2** (harness.md status update) — final hygiene.
3. **A1** (strict convergence investigation) — only if appetite remains.

If the goal is "ship a useful tool":

1. **C3** (Notable Features Markdown block) — biggest user-facing win.
2. **C2** (Pass-3 referee knobs) — gives referees the variants.
3. **C1** (merge) — make it the default branch.

If the goal is "preserve pass-2 indefinitely as the v2 design":

1. **C1** (merge with regression baseline as the canonical output).
2. Defer everything else; pass-2 is the new baseline.

With A0 (pass-1 comparison) dropped, the close-pass-2-fully path is dramatically shorter. Pass 2 reached architectural completeness. What remains is choice of direction.
