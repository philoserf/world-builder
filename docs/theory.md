# Theory of the codebase

This is a Naur-style theory of `world-builder`: the mental model a competent maintainer must hold to change this system without damaging its conceptual integrity. It is for the engineer inheriting the project — someone expected to make decisions, not just follow the docs. The companion documents in `docs/` (`design-intent.md`, `dependency-graph.md`, `api-surface.md`, `wbh-inconsistencies.md`, `anti-patterns.md`, `harness.md`) are the policy artefacts; this is the picture to hold in your head while reading them.

## What this system is for

The system encodes a published roleplaying-game supplement — Mongoose's _World Builder's Handbook_ (Geir Lanesskog, 2023) — as an executable Go program. WBH is a procedure book: pp.14–146 describe, with worked examples, how a referee constructs a star system at the table — roll some dice, read a table, apply a modifier, copy a value to a form. The thing being modelled is not "outer space" and not "Mongoose Traveller" the game; it is _the act of generating a system by the book's procedures_. Every entity in the code corresponds to something a referee would write on paper while running those procedures.

The vocabulary is the book's, not Go's and not astronomy's. `HZCO` is a star group's habitable-zone central orbit. `SAH` is the Size-Atmosphere-Hydrographics triplet that becomes one cell of a survey form. `MAO` is the maximum allowed orbit for an interior body. `TSS` is total seismic stress — residual plus tidal-stress factor plus tidal-heating factor — which feeds an inherent-temperature addition. `IISS` is the fictional in-setting survey service whose printed forms are the output the program produces. `Zed`, `Zed Prime`, `Corella`, `Sol`/`Terra` are the worked examples the book threads through its chapters — and they are first-class fixtures in the test suite, by name, driven by dice scripts that reproduce the values the book prints.

The output is one document in three renderings of the same content:

- **The default and canonical output is a single IISS Class IV Survey document in Markdown**: an H1 title, a Notable Features summary, **PART 1 — System Census** (system scalars, the stellar roster with full orbital data, and the body roster), then a **PART P** (planet / moon / gas giant) or **PART P.B** (belt) section per non-empty body in orbit order, the auto-picked mainworld's part suffixed "— mainworld".
- A JSON dump of the same aggregate (`iiss.SystemForms`).
- A `G-P-T-N-S` short profile string (gas giants, belts, terrestrials, baseline number, spread) per WBH p.58.

A single integer seed plus the procedures determine the system fully. Two `world-builder -seed 42 -format markdown` invocations on different machines produce byte-identical output.

This is not a general-purpose star-system generator. It is a faithful executable reading of one specific book, and that choice is load-bearing: it fixes the test gate (the book's printed worked examples), the doc conventions (WBH page numbers in doc-comments), and the scope edge (pp.14–146 in, pp.147+ out). If you want to add a generator option that isn't in WBH, stop and re-read `design-intent.md` § "What the project commits to" before you write it.

## The two axioms

Two ideas decide nearly everything downstream.

**The book is the specification.** Where the book says "roll 2D and apply DM+1 if X", the code rolls 2D and applies DM+1 if X — through the same path, in the same order, with the same modifiers. Where the book disagrees with itself (six known cases, catalogued in `wbh-inconsistencies.md`), the code commits to one interpretation, a test fixture asserts that interpretation, and a doc-comment cites which page wins and why. No runtime toggles between readings. When you read a procedure you should be able to lay the book beside it and walk the lines together; the page-number doc-comments make that mechanical. The acceptance gate is therefore not "do the tests pass" in a Go sense but "does the program reproduce the book's printed worked examples to the digit." Breaking a worked-example fixture is breaking fidelity to the book, not breaking a test.

**The seed is identity — and identity is now a tree, not a line.** Every random draw passes through `roller.Roller`. There is no `math/rand` anywhere outside `roller/` (verified: the only import is inside `roller.go`), and there is no package-level RNG state. `Seeded` is production; `Scripted` replays an exact sequence and panics on exhaustion (an exhausted Scripted always means a fixture bug, never a runtime problem); `Fixed` returns one value for property tests that pin a variable.

The subtle, load-bearing part — and the piece most likely to be missing from an out-of-date mental model — is _how_ the seed reaches each procedure. It is not a single global dice stream consumed in pipeline order. `Roller` has a second method, `Fork(key string) Roller`, and the whole per-body half of the pipeline runs on forked substreams:

```
r.Fork(bodyID).Fork(family)     // worlds/subroller.go: bodySub()
bodyID = parent.Designation + "/" + body.Designation   // moon; else body.Designation
family = "climate" | "geology" | "rotation" | ...
```

`Seeded.Fork` hashes the roller's **immutable construction seed** (not its live RNG position) with the key, so a fork is a pure function of `(seed, key)` — independent of how many draws earlier stages took from the parent. The consequence is the central architectural fact of the current codebase: **each body's output is a function of `(seed, body-identity, procedure-family)` only.** Reordering two independent stages, or re-rolling one body mid-pipeline, leaves every other body byte-identical. `Scripted` and `Fixed` implement `Fork` transparently (they return themselves), so worked-example fixtures that feed one flat narrated dice list are blind to the fork topology — the tree is invisible to a scripted stream.

Only the **structure prefix** — `stars.GenerateSystem`, `GenerateSystemPlacement`, and `ApplyDetailFrontEnd` — stays on the shared stream. That prefix is where body identity is _created_, so there is nothing stable to key a fork on yet; and none of the pain that motivated forking lives before identity exists. Once `ApplyDetailFrontEnd` has assigned designations, every suffix stage forks per body.

This is the deepest expression of "seed is identity," and it is not decorative. It is what makes three otherwise-dangerous things safe, and you will see all three below: free reordering of independent stages, a single-body re-roll in the middle of the pipeline, and a full-pipeline gold master. If you touch the pipeline, hold this invariant first.

## Pagination is not architecture

The hardest-earned lesson in the project: the book's chapter order is not the data's dependency order. An earlier implementation ("pass 1") followed WBH pagination as if it were architecture, and paid for it three ways — a cyclic climate cluster discovered mid-flight and patched awkwardly; a recurring bug where new per-body procedures iterated planets but not moons (the "moon-path silent-zero," logged four separate times); and renderer asymmetries because no chapter owned the API surface as a whole.

The current implementation ("pass 2", plus a "C1–C5" cleanup arc) inverts the relationship: **the data dependency graph determines structure; worked-example fixtures determine the acceptance gate; the book becomes a citation system, not an architecture.** Stages are numbered by where their values sit in the dependency graph (`dependency-graph.md`), not by which chapter introduces them. That graph is almost entirely acyclic — one cycle, the climate cluster, plus one deliberate back-edge, the tidal-lock re-evaluation. Nearly every design decision follows from where in the graph you stand:

- **Acyclic regions** are forward-only `Apply*` orchestrators that walk every body once.
- **The cycle** gets one explicit per-body solver (next section).
- **The back-edge** gets one explicit re-evaluation stage that re-rolls affected bodies (the section after).
- **One iterator** walks every body — planets, belts, and moons interleaved. The moon-vs-planet distinction is a `Kind` value on a single `Body` struct, not a separate type or a separate iteration.

The pipeline, read top-to-bottom from `worlds/generate.go`, is:

```
Stage 0   stars.GenerateSystem                              pp.14–35   (structure prefix, shared stream)
Stage 1   GenerateSystemPlacement                           pp.36–68   (structure prefix, shared stream)
Stage 2   ApplyDetailFrontEnd                               sizing, moons, designations, periods
                                                                       ↑ last shared-stream stage; forks below
Stage 3   ApplyBodyPhysical, ApplyBeltDetails               composition / density / gravity / mass; belt detail
Stage 4   ApplyMoonRefinement, ApplyRotationTilt            rotation, axial tilt, surface tidal effects
Stage 5   ApplyClimate                                      the climate cluster (two-pass solver)
Stage 5'  ApplyTidalLockReEval                              the deliberate back-edge (WBH p.106)
Stage 6   ApplyTaintTypology, ApplySurfaceDistribution
Stage 7   ApplyGeology                                      pp.125–127
Stage 8   ApplyBiology                                      pp.127–131
Stage 9   ApplyHabitability                                 pp.132–138  (no roller — pure)
Stage 10  AggregateSystem (pure), BuildIISSForms (pure)     profiles, mainworld pick, the survey document
```

If you internalise one habit: when you change something, ask first what its inputs are in the graph and what depends on it. The file the procedure lives in is a hint; the graph is the contract. (Orchestrator files are named for their role — `climate.go`, `rotation_tilt.go`, `taint_surface.go`, `detail_frontend.go`, `physical_detail.go`, `aggregate.go` — or live beside their feature's procedures: `geology.go`, `biology.go`, `habitability.go`. The old `stageN.go` numbering was retired in C5. The stage _numbers_ persist in doc-comments and test names, where they correctly denote graph indices.)

## The climate cluster, and why "convergence" would lie

WBH atmosphere, hydrographics, and temperature are mutually dependent. Hydrographics depends on temperature range; temperature depends on albedo, and albedo reads hydrographics; atmosphere depends on temperature through runaway-greenhouse mutation, and temperature depends on atmosphere through a greenhouse factor. There is no single-pass evaluation order that gives the right answer. The book implicitly assumes the referee derives provisional values, then refines.

The solver is `ApplyClimatePasses` (`worlds/climate.go`). It runs exactly **two** passes of: compute temperature from current atm/hydro → compute the atm/hydro-independent partial-geology factors → apply the TSS inherent-temperature addition (`T' = ⁴√(T⁴ + TSS⁴)`) and refresh scale height → rederive atmosphere and hydrographics from the post-TSS temperature. The second pass's result is trusted.

The name matters because an earlier design told a different story. Pass-2's original framing was a **fixed-point solver**: iterate until atmosphere, hydrographics, and mean temperature stabilise; cap iterations; assert convergence. That version (cycle 17) hit overflow on common seeds, and an instrumentation spike found why: `RederiveAtmosphereHydrographics` calls `RollHydroDigit`, which **consumes fresh dice on every call**. Each pass is a fresh stochastic sample, not a convergence step — there is no fixed point to find, because hydrographics is a probability distribution over a band and successive samples can legitimately disagree while nothing else changed. The code reverted to two passes and renamed `ConvergeClimate` → `ApplyClimatePasses`; the evergreen docs and code comments were swept (C2) to describe two-pass sampling, and the dead `Climate` convergence-variable struct was removed.

The standing instruction for a maintainer: **wherever you meet "convergence" or "fixed point," mentally substitute "two stochastic samples, second one wins."** That language now survives only in `history/` (deliberately — `history/lessons-learned.md` § L13 needs the word to tell the story). If you find it in an evergreen doc or a code comment, treat it as a regression and fix it. And if you ever add a third pass or assert iteration stability, you will reproduce cycle 17's overflow.

One structural note that trips people: by the time climate finishes, `body.Geology` is already partly populated — the TSS factors and the inherent-temperature addition are folded into each pass via `computePartialGeology`, because temperature needs them. Tectonic plates and gas-giant residual heat are filled in later by `ApplyGeology` (Stage 7). So "geology runs at Stage 7" is only half true; the seismic-stress half runs inside climate.

## The tidal-lock back-edge

`ApplyTidalLockReEval` (Stage 5', between climate and Stage 6) is the graph's one deliberate back-edge, and it is the clearest payoff of the Fork tree. WBH p.106 says a dense-enough atmosphere can break a tidal lock that Stage 4 assigned before atmosphere existed. So Stage 4 stashes a `preTidalLockSnapshot` on the body just before applying the lock; after climate has produced an atmosphere, Stage 5' consults the atmosphere DM and, where warranted, restores the pre-lock state and re-rolls the body's rotation cascade.

Re-rolling one body mid-pipeline is exactly the operation a single shared dice stream cannot survive — under pass-1's stream, the extra draws would shift every downstream body. Under C1, the re-rolled body draws from its own `(identity, family)` substream, so its neighbours are untouched. The back-edge is only affordable because of the fork topology. If you add a second back-edge, this is the pattern to copy: snapshot the pre-state, re-derive from a forked substream, and confirm with a property test that unrelated bodies are byte-identical before and after.

## The unified Body and the iterator

Every placed thing — terrestrial planet, gas giant, planetoid belt, moon — is a `Body` (`worlds/body.go`). Moons live in `Body.Children` and point back via `Body.Parent`; the kind is a `BodyKind` enum, not a separate type. `BodyEmpty` (the zero value) is the sentinel for an orbit slot with no world assigned — it is a `Kind`, not a nil `*Body`, so procedures that walk the iterator must recognise and skip it rather than deref-guarding. `Universe.AllBodies()` yields each body and then its children in placement order; `AllBodiesWithParent()` adds the parent for procedures that need parent context. The "moon path" is not a separate code path — the same iterator yields it, and the silent-zero bug class is closed at the type level.

The trade-off this unification introduces, and the thing a new maintainer must hold: **a moon's `Orbit` field is unset.** For planets and belts, `Body.Orbit` is the orbit around the star; for moons, `Body.OrbitPD` / `Body.OrbitKm` carry the orbit around the _parent_, and the orbit-around-the-star is the parent's. Three helpers exist precisely so you never re-implement that branch:

- `Body.StellarOrbit()` — orbit around the host star for any body (parent's orbit for a moon).
- `Body.Host()` — the body that orbits a star directly (the parent for a moon, the body itself otherwise). Reach for this whenever you need "the thing whose HZ flag, stellar orbit, or parent mass governs this procedure." Climate's eligibility check and HZCO offset both route through it.
- `Body.MassOrDerived()` — mass in Earth masses, preferring `MassEarth`, falling back to density × volume. It subsumes three previously open-coded copies.

Pointer-typed fields (`*Atmosphere`, `*Geology`, `*Biology`, …) are nil when the stage that populates them did not apply to this body; the `Has*()` predicates wrap the nil checks. Use them; don't deref blindly. The pointer-vs-nil shape is the project's chosen way of saying "this kind of body doesn't get this kind of state."

The litmus test for any new procedure: does it run for every body the iterator yields, including moons of gas-giant parents? If yes, iterate `u.AllBodies()` (or `AllBodiesWithParent()`) and switch on `Body.Kind` where the rules differ. If no, make the exclusion criterion explicit. The pre-flight checklist is `anti-patterns.md` § A.1 and it is non-negotiable — it is the type-level closure of pass-1's four-times silent-zero recurrence, and `TestProperty_MoonsHaveBodies` is its sentinel.

## The Apply mutation discipline

Mutation lives at the stage level; procedures are pure. `Apply*` walks the universe and writes to `*Body` fields in place. `Compute*`, `Derive*`, and `Roll*` take values, return values, and never reach into shared state. Pass 2 considered returning a fresh `Universe` per stage and rejected it: the universe carries mutable per-body state through ten stages, and copy-per-stage buys nothing semantic. The naming protocol is binding, not aesthetic, and the discriminator is dice consumption:

- `Generate*` — rolls a sub-system from a Roller; returns a value and an error.
- `Roll*` — rolls a single value; returns it and an error (the error is dice-exhaustion for Scripted tests, mainly).
- `Compute*` — deterministic; value, no error.
- `Derive*` — pure formula; no Roller, no error.
- `Apply*` — mutates a target; the mutation point is documented, the post-condition is tested.

If you write a function whose name doesn't fit, the likelier diagnosis is that the function does too much, not that the protocol is wrong.

## The iiss boundary

`iiss/` (the package, not the in-fiction service) is the seam between system construction and rendering. The form types (`Class0IForm`, `Class23Form`, `Class4PForm` with its `Class4PPartP` / `Class4PPartPB` bodies, `SystemForms`, `FormHeader`), the Markdown renderer (`MarkdownClass4Survey` and its part helpers in `render.go`), and the Class IV-P body structs and their `RenderBody` methods (`class4p.go`) all live there. **`iiss/` does not import `worlds/` in production** (the only such import is in `iiss/regression_test.go`, which drives the full pipeline to snapshot its output — a test dependency, not a production one). The boundary type crossing the seam is `iiss.SystemForms`, embedded in `worlds.SystemDetail`, populated by `worlds.BuildIISSForms`, and consumed by the renderer. `worlds` **builds** the forms; `iiss` **renders** them.

A subtlety in the current output surface, since it drifts from older descriptions: the program no longer emits three separate survey forms. It emits **one** IISS Class IV Survey document. `Class0IForm` and `Class23Form` survive not as rendered outputs but as **PART 1 data carriers** — their star rows and object/count tables are folded into the census that PART 1 prints. Class IV-P's body used to be the one asymmetry (a `RenderBody` closure and `any`-typed fields pointing back into worlds); C3 moved those structs and renderers into `iiss/` as concrete types, so all three form types now marshal to JSON as concrete structs and the worlds→iiss boundary is uniform. If you read a doc that speaks of "three IISS forms rendered out of one aggregate," read it as "one Class IV Survey document, whose PART 1 folds in what the 0/I and II/III forms used to show standalone."

## What sits outside the gate, and how it says so

The project commits to WBH pp.14–146. It does not commit to pp.147+ (World Social Characteristics, Special Circumstances), nor to detailed special-object physics (white-dwarf interiors, neutron-star fields). Those cuts are encoded in the type system, not merely documented. When a generation path hits an out-of-scope condition it does **not** silently degrade or substitute a placeholder: it returns a typed error wrapping `stars.ErrSpecialCircumstances` (see `stars/errors.go`; `ErrCompanionOfGiantMAO` is the archetype). Property tests and bulk sweeps classify the whole family with a single `errors.Is(err, stars.ErrSpecialCircumstances)` — the `isSpecialCircumstances` helper in `worlds/property_test.go` — and skip those seeds; a few percent of random seeds hit one. The 10 000-seed bulk sweep run during the "clean every run" project reports all-successes precisely because the WBH-provided Referee options for the common edge cases were adopted, narrowing the out-of-scope hits to genuinely out-of-scope categories.

To add a new guard: `errors.New` a sentinel (or reuse one), wrap at the call site with `fmt.Errorf("%w: …", ErrSpecialCircumstances, …)`, and let it propagate. Never pattern-match on error strings; never catch-and-continue.

## The six committed inconsistencies

`wbh-inconsistencies.md` enumerates six places where the book contradicts itself. Each is resolved by one interpretation, encoded in code, asserted by a fixture marked ⚠️ in the harness, and explained in the doc:

1. **HZCO formula vs p.42 table** — five Class VI cells disagree by >5%. Code follows the formula; tests skip those cells.
2. **Aab IV-d (Zed Prime) sizing** — p.58 table says S; p.63 form says Size 5. Code follows the form.
3. **Three temperature-chapter divergences** — albedo at Hyd 6+ (follow the table), Terra reference greenhouse (follow the book's 0.36 even though it mismatches real Earth), Zed Prime WorstLow (compute consistently, accept divergence from the sidebar).
4. **Residual seismic stress density DM** — table says +2, worked example uses +1. Code follows the table.
5. **Compatibility "+3" addend** — the worked example shows an unsourced +3; the formula box doesn't. Code follows the formula. This is the _only_ case where the chosen interpretation does **not** reproduce the canonical Zed Prime form value; encoding the +3 would mean a magic number with no procedural justification.
6. **Habitability gravity DM overlap** — two bands overlap; the footnote says "use worst at edges" but the worked example uses the narrower band. Code follows the worked example (it reproduces the form's Habitability=7).

The decision heuristic across the six is "follow whichever interpretation reproduces the canonical Zed Prime IISS form on pp.141–142, unless doing so requires an unsourced magic number." Entry 5 is the exception that makes this a heuristic, not a rule. When you surface a seventh inconsistency, the discipline is: do not pick a side in code without writing the seventh entry, citing both sources. The audit trail _is_ the discipline.

## The test discipline — and the gold master's return

The suite has layers that catch different failures; do not collapse them.

- **Per-procedure dice-script tests.** One per `Roll*` / `Generate*` / `Compute*`, driven by `Scripted` rollers with the book's narrated dice, asserting outputs to the digit. These are the proof of fidelity. Many predate pass 2 and were ported verbatim, because the architecture changed how procedures are orchestrated, not what they compute.
- **Property tests** (`worlds/property_test.go`, eight invariants × ~1000 seeds). They catch silent-zero / silent-skip across the population — the class per-procedure tests miss because each sees one fixture. `TestProperty_MoonsHaveBodies` guards anti-pattern A.1; `TestProperty_HZBodyHasClimate` guards "did atmosphere/hydro/temp populate for HZ moons too."
- **Façade fixtures** (`Sol/Generate`, `Zed/Generate` over ~100 seeds) assert _shape_ — no `?` in SAH triplets, every body has rotation state, a mainworld got picked.
- **Markdown regression baselines** (`iiss/testdata/seed_*.md`) catch unintended renderer drift, refreshed deliberately with an update flag after review.

And then the piece that reverses an older lesson, so read carefully. Pass 1 kept a full-pipeline dice-script gold (`TestZed_FullDetail`) that asserted intermediate values in pipeline order; it died the moment the pipeline reordered, because on a single shared stream any reorder shifts every downstream value. `history/spike-findings.md` § Finding 2 records the conclusion that such a fixture is "anti-gold" — it claims fidelity but actually asserts pipeline ordering. Early theory documents took from this that pass 2 should keep _no_ full-pipeline gold.

**That conclusion no longer holds, and the code says so.** C1's fork tree made each body's output a pure function of `(seed, identity, family)`, which dissolved the exact reason full-pipeline gold was fragile. So the current suite reintroduces one, deliberately: `TestZed_GoldMaster` (`worlds/zed_gold_test.go`) runs the named Zed system end-to-end at a pinned seed, renders the complete Class IV Survey Markdown, and diffs it against `worlds/testdata/zed_gold.md`. Its companion `TestZed_GoldSurvivesStageReorder` runs the pipeline twice with the two mutually-independent post-climate stages (taint typology, surface distribution) swapped, and asserts the entire system — every `Body` and the rendered document — is byte-identical. That test _is_ the executable proof of the C1 invariant, and it is what makes the gold master a trustworthy regression anchor rather than an ordering assertion.

The distinction to keep straight: pass 1's anti-gold was a **dice-script** gold asserting **intermediate values in order**; the current gold is a **rendered-Markdown** snapshot of the **final document**, protected by a reorder-survival test. They are different animals, and only the second is safe. If you propose a full-pipeline gold, it must be the second kind, and it must ship with a reorder-survival companion — otherwise you are re-walking pass 1's mistake.

For a new procedure: the per-procedure test is mandatory (it reproduces the book); the property test is mandatory if the procedure is per-body (it catches the silent-zero); the regression and Zed-gold baselines are refreshed if visible output changes.

## What changes easily, what doesn't

**Shaped to accommodate:**

- _Adding a procedure inside a WBH section._ Write it in the role-named or feature file with page citations; add a per-procedure dice-script test; if per-body, slot it into the relevant `Apply*` body loop, forking a `bodySub` for its dice; if per-system, into an aggregation. The unified `Body` plus the iterator make the per-body case mostly mechanical.
- _Fixing a misinterpretation._ Change the function, change the fixture, run `task`. Worked-example tests catch correctness regressions immediately.
- _Reordering two independent stages._ The fork tree makes this byte-safe; `TestZed_GoldSurvivesStageReorder` is the proof, and C5 exploited exactly this to rename and reorder the orchestrator files.
- _Adding a seventh committed inconsistency._ Entry in `wbh-inconsistencies.md` with both sources, interpretation in code, fixture flipped to ⚠️.
- _Adjusting the survey rendering._ Edit `iiss/render.go` / `class4p.go`; refresh the regression and Zed-gold baselines if the change is intended.
- _Adding an output format._ New sibling function in `iiss/`. No interface required — `design-intent.md` § "Stop rules" rejects one.

**Require rethinking something fundamental:**

- _Introducing a genuine graph back-edge._ A new edge from a later stage into an earlier one is the territory of the climate cluster and the tidal-lock re-eval. You must decide whether to fold it into the climate solver or model it as an explicit snapshot-and-re-derive stage like Stage 5' — and, either way, prove with a property test that unrelated bodies are unperturbed.
- _Replacing the Roller seam._ Every procedure consumes dice through `roller.Roller`, and the `Fork` contract is now load-bearing, not just `Roll`. The `Scripted`-with-book-dice strategy and the position-independence dividend both die if you swap this for an effect system, monadic IO, or context-threaded RNG. This is the highest-leverage seam in the codebase; touching it is a different project.
- _Splitting moons back out of `Body`._ The unified iterator and `Has*()` shape are the type-level closure of the silent-zero class. Pass 1's four-times recurrence is the documented cost of a separate moon path. Don't.
- _Bringing WBH pp.147+ into scope._ The `ErrSpecialCircumstances` family becomes a generation path; new domain entities (governments, trade codes, sophont social characteristics) appear; the rendering surface (a fourth part? a society sheet?) is open design. This is the most plausible "next major scope" and is post-v1 per `next-steps.md`.

A maintainer _with_ the theory reaches first for the dependency graph and the iterator-plus-`Has`-pointer shape, and forks a substream for new dice. A maintainer _without_ it reaches for the book chapter, finds the file by name match, and writes code that iterates planets only, or reads a pre-climate value where it needs a post-climate one, or takes dice off the shared stream and quietly breaks reorder-invariance. `anti-patterns.md` catalogues the first two mistakes (A.1 moon-path silent-zero, A.8 stale pre-climate inputs); the third is the new one the fork tree introduces and the reorder-survival test defends against.

## Where the theory is thin, and where I'm inferring

Built from the code, the design docs, and `main`'s history. Places I'm less than certain, or where the text and the code aren't yet in lockstep:

- **Docs still trailing the C1 dividend.** The fork tree is documented at the code level (`roller.go`, `subroller.go`, the `zed_gold_test.go` comments), but not every evergreen doc frames determinism as `(seed, identity, family)` position-independence rather than "single stream, order fixed socially." Where a doc still says the latter, or still says pass 2 keeps no full-pipeline gold, treat it as lagging the code — the gold master and its reorder test are in the tree and passing.
- **The working tree is mid-refresh.** At the time of writing, `worlds/testdata/zed_gold.md` and `docs/walkthrough.md` carry uncommitted regenerations for the "HZ breadth in PART 1" feature (recent commits #111–#114 reconcile PART P fields, ring centre/span, gas-giant PART P, and HZ breadth). The gold and walkthrough files are generated artefacts; regenerate them with the update flags after a reviewed output change rather than hand-editing.
- **The bulk-sweep tool isn't in `main`.** `harness.md` describes a `cmd/world-builder-bulk` runner producing 10 000 successes; that binary lived on a branch during the clean-every-run project and is not in the tree (`cmd/world-builder` is the only command). To re-verify "is the system clean today," recreate the sweep from history (commit `328b37c`) or write it fresh — `task test`'s ~1000-seed property runs are a sample, not the sweep.
- **The Zed-Prime heuristic is a rule of last resort, not a mechanical one.** I infer the authors decide the formula-vs-form judgment case by case and consult "reproduce Zed Prime" only when a principled reading doesn't settle it. If a seventh inconsistency arrives where each interpretation reproduces some of the form and neither reproduces all, the project will have to make a judgment, not run the heuristic.
- **`BodyEmpty` leakage.** It's confirmed a `Kind` sentinel (the zero value), skipped by procedures rather than filtered out of the slice. I have not exhaustively traced every renderer's handling; if an empty slot ever reaches a future renderer, the likely behaviour is "skip silently," which is fine for current renderers but could surprise a new one.

## A maintainer's posture

Hold the whole thing as one shape:

- A book that prescribes procedures; a generator that runs them, citing page numbers as it goes.
- A seed that names the run, reaching each body through a fork tree so that identity is `(seed, body, family)` — which is what lets stages reorder, one body re-roll, and a full-pipeline gold master all be safe at once.
- A graph of values, forward-only except for one cycle (climate, sampled twice, second wins) and one back-edge (tidal-lock re-eval, snapshot and re-derive).
- One `Body` type walked by one iterator; pointers and `Has*()` for stage-conditional state; `Host()` / `StellarOrbit()` for the moon indirection.
- One IISS Class IV Survey document rendered from one shared aggregate, fully owned by `iiss/` as concrete structs, with Class 0/I and II/III surviving only as PART 1 data carriers.
- A test suite whose primary assertion is "the book is reproduced," backed by population invariants, a rendered-Markdown gold master, and its reorder-survival proof.
- A scope edge encoded as a typed error family, never a silent fallback.

When you change something, ask in order: where is this value in the graph; does it run for every body the iterator yields, including gas-giant moons; from which forked substream does it draw its dice; what does the book say at the cited pages; which fixture proves the chosen interpretation; what changes downstream. If you can answer those, the change is safe. If you can't, `dependency-graph.md` and `design-intent.md` will tell you which question you're not yet positioned to answer.

The two substitutions that cover most of where pass 1 went wrong and where the cleanup arc spent its budget: when you see "convergence," think "two stochastic samples, second wins"; when you see "for each planet," write "for each body" instead. And one more, newer than those: when you reach for dice in a per-body stage, fork a substream — never draw from the shared stream past `ApplyDetailFrontEnd`, or you silently break the invariant the gold master depends on.
