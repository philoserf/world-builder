# Pass 2 — Next Steps

Three categories: items requiring user judgement (the trade-offs are real and pass-2-side automation can't resolve them); mechanical items (need labor, no design decision); strategic / scope items (merge planning, pass-3 outlook).

The ordering within each category roughly reflects priority — higher items unblock or de-risk lower ones.

## A. Items requiring your judgement

**A0 (dropped).** A pass-1-vs-pass-2 byte comparison was originally listed here as the unfinished fidelity-gate item. It is now explicitly **dropped**: pass-2's design intentionally diverges from pass-1 on multiple axes (TSS fold-into-climate, surface-distribution-after-converge, narrower-band-wins for gravity DM, etc.), and the within-pass-2 Markdown regression baseline (`iiss/testdata/seed_*.md` + `TestRegression_MarkdownSeeds`) is the working substitute. Future drift between pass-2 cycles is caught by the baseline; pass-1's outputs are no longer authoritative.

### A1. Strict `ConvergeClimate` convergence — RESOLVED

**Status.** Closed via option (b) investigation. The oscillation root cause was identified empirically: `RederiveAtmosphereHydrographics` calls `RollHydroDigit`, which consumes fresh dice from the Roller each call. Each iteration is a stochastic sample of hydro (and, via albedo, temperature), not a convergence step. There is no fixed point to find because the system isn't deterministic in the convergence variable.

**Resolution.** Reverted `ConvergeClimate` to exactly 2 passes (matching pass-1's flow with TSS folded into each pass). Removed the N-iteration loop, early-exit check, and convergence assertion — they were over-engineering of a stochastic-sampling pattern. The name "ConvergeClimate" is retained for continuity but is a misnomer.

**Cascade.**

- `api-surface.md` § The Climate solver updated to reflect 2-pass sampling behaviour.
- `lessons-learned.md` § L13 documents the architectural finding.
- `iiss/testdata/seed_*.md` regression baseline refreshed to capture the new 2-pass output (one seed drifted; magnitudes were small and sensible).
- No external API change: `ConvergeClimate(r, body, sys) error` signature is unchanged.

### A2. `stars.Group` migration to `stars/` package (cycle 12 deferred)

**Status.** `api-surface.md` § Open questions, decided says Group should live in `stars/`. Cycle 12 was deferred because `worlds.Group` has unexported fields (`companionEcc`, `sourceCompanion`) accessed by tests in the same package; moving requires either exporting them or threading getters.

**Why your call.** Three sub-decisions:

- **Export-vs-getter.** Renaming `companionEcc` → `CompanionEcc` and `sourceCompanion` → `SourceCompanion` is one line of edits but breaks encapsulation. Adding `Group.CompanionEcc()` getter methods preserves encapsulation but adds 4 method definitions.
- **Test refactor.** `worlds/available_orbits_test.go` accesses the unexported fields. Either export them or rewrite the tests to use the public API.
- **Worth it at all?** The migration is pure cleanliness — there's no functional benefit. The code already works. Cycle 12's deferral may be the right answer indefinitely.

**Recommendation.** Defer indefinitely unless `stars.Group` starts attracting unrelated coupling. The api-surface.md decision was right in principle; the cost in practice doesn't justify it for a working system.

### A3. `stars.GenerateSystemOpts` cuts — RESOLVED (option b: keep load-bearing)

**Status.** Closed. The `WithVariance` and `Accuracy` fields stay in `stars.GenerateSystemOpts` indefinitely. The original `design-intent.md` cuts-list claim that these would be cut is reclassified as aspirational — these toggles are load-bearing for ~10 pass-1 worked-example tests that drive the book's narrated dice scripts (e.g. `Accuracy: 1` selects the simple-1D age path).

**Resolution.** No code change. The cuts list in `design-intent.md` updated to mark these as preserved-not-cut, with the worlds-side variance fields and `AccurateAlbedo` / oxygen-atm biomass floor still genuinely cut.

**Why this is honest, not a regression.** Pass-2's cuts-list intent was to reduce the API surface by removing toggles that gated speculative variants. The variance and accuracy toggles on `stars.GenerateSystemOpts` are book-fidelity load-bearing, not speculative variants — they choose between two equally-correct pass-1 procedures specified by WBH. Removing them would erase book-fidelity tests; preserving them adds zero ambiguity to callers (production sets `WithVariance: true, Accuracy: 2` and is done).

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
