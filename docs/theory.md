# Theory of the codebase

This is a Naur-style theory of `world-builder`: the mental model a competent maintainer must hold to change this system without damaging its conceptual integrity. It is meant for an engineer inheriting the project — someone who will be expected to make decisions, not just follow the docs.

The companion documents in `docs/` (`design-intent.md`, `dependency-graph.md`, `api-surface.md`, `wbh-inconsistencies.md`, `anti-patterns.md`, `harness.md`) are the policy artefacts. This document is the picture you should keep in your head while reading them.

## What this system is for

The system encodes a published roleplaying-game supplement — Mongoose's _World Builder's Handbook_ (Geir Lanesskog, 2023) — as an executable Go program. WBH is a procedure book: pp.14–146 describe, with worked examples, how a referee constructs a star system at the table — roll some dice, read a table, apply a modifier, copy a value to a form. The system being modelled is not "outer space" and not "Mongoose Traveller"; it is _the act of generating a system at the table by the book's procedures_. Every entity in the code corresponds to a thing a referee would write on a piece of paper while running the procedures.

The vocabulary is the book's, not Go's and not astronomy's. `HZCO` is the habitable-zone central orbit of a star. `SAH` is the Size-Atmosphere-Hydrographics triplet that becomes one cell of an IISS Class II/III form. `MAO` is the maximum allowed orbit for an interior body. `TSS` is total seismic stress: residual plus tidal-stress factor plus tidal-heating factor. `IISS` is the fictional in-setting survey service whose three printed forms (Class 0/I, Class II/III, Class IV-P) are the output the program is ultimately trying to produce. `Zed`, `Zed Prime`, `Corella`, `Sol`/`Terra` are the worked examples the book itself threads through its chapters — and they are first-class fixtures in the test suite, by name, with dice scripts that match what the book reports.

The output is one of three forms of the same content:

- A complete Markdown rendering of all three IISS forms for the generated system (default; the canonical output).
- A JSON dump of the same.
- A `G-P-T-N-S` short profile string (gas giants, belts, terrestrials, baseline number, spread) per WBH p.58.

A single integer seed plus the procedures determine the system fully. Two `world-builder -seed 42 -format markdown` invocations on different machines produce byte-identical output.

This is not a general-purpose star-system generator. It is a faithful executable reading of one specific book. That choice is load-bearing: it picks the test gate (the book's printed worked examples), the doc conventions (WBH page numbers in doc-comments), and the scope (pp.14–146 in, pp.147+ out). If you find yourself wanting to add a generator option that isn't in WBH, stop and re-read `design-intent.md` § "What the project commits to" before you write it.

## The book as spec, the seed as identity

Two ideas decide nearly everything.

**The book is the specification.** Where the book says "roll 2D and apply DM+1 if X", the code rolls 2D and applies DM+1 if X — through the same path, in the same order, with the same modifiers. Where the book disagrees with itself (six known cases consolidated in `wbh-inconsistencies.md`), the code commits to one interpretation, the test fixture asserts that interpretation, and the doc-comment cites which page wins and why. No runtime toggles. When you read a procedure, you should be able to put the book next to it and walk the lines together. The doc-comments on tables and procedures cite page numbers so this works mechanically.

The acceptance gate is therefore not "do the tests pass" in a Go sense — it is "does the program reproduce the book's printed worked examples to the digit." The harness (`harness.md`) catalogues those examples. Most are encoded with `roller.NewScripted(...)` driven by the dice the book narrates; the assertions compare against the values the book prints. When a fixture is marked ⚠️, it asserts the implementation's chosen interpretation of a book-internal inconsistency, with a comment naming what the book prints and what `wbh-inconsistencies.md` § N says. Breaking such a fixture is breaking the book's fidelity, not breaking a test.

**The seed is identity.** Every random draw passes through `roller.Roller`. There is no `math/rand` outside `roller/`, and there are no globals. `Seeded` is the production implementation; `Scripted` replays an exact sequence (and panics on exhaustion, because an exhausted Scripted always means a fixture bug, never a runtime problem); `Fixed` returns one value for property tests that need to pin one variable. The seed plus the sequence of procedures the pipeline runs determines the entire system bit-for-bit.

That sounds obvious for a deterministic generator. The thing it buys you is testability with the same fixtures the book uses. The book says "we roll a 9, then a 7"; the test feeds 9 then 7 to a `Scripted` roller and asserts the same downstream values the book prints. This works only because no procedure reaches for randomness outside the interface, and because the roll order inside each procedure is stable. Both invariants are enforced socially (no exceptions accepted), not mechanically — but the `Scripted`-with-known-dice tests would scream the moment you broke one.

## Pagination is not architecture

The hardest-earned lesson in the project is this: the book's chapter order is not the data's dependency order. An earlier implementation ("pass 1") followed WBH pagination as if it were architecture, and the architecture quietly accumulated three categories of cost: a cyclic climate cluster (atmosphere ↔ temperature ↔ hydrographics) that was discovered mid-flight and recovered awkwardly with a two-pass rederive; a recurring class of bug where new per-body procedures iterated planets but not moons (the "moon-path silent-zero," logged four times); and renderer asymmetries because no chapter owned the API surface as a whole.

The current implementation ("pass 2") inverts the relationship. **The data dependency graph determines structure. Worked-example fixtures determine the acceptance gate. The book becomes a citation system, not an architecture.** Stages are numbered by where their values sit in the dependency graph (`dependency-graph.md` § Stages 0–10), not by which chapter introduces them.

That graph has one cycle in it — the climate cluster — and the rest of the system is acyclic. Almost every design decision in pass 2 follows from where in the graph you are:

- **Acyclic regions** are forward-only `Apply*` orchestrators that walk bodies once. Sizing, day length, axial tilt, biology, habitability, system aggregations — each is a single pass.
- **The cycle** gets one explicit per-body solver (see next section).
- **One iterator** walks every body, planets and moons interleaved. The moon-vs-planet distinction is a `Kind` value on a single `Body` struct, not a separate type or a separate iteration. The silent-zero bug class is closed at the type level.

If you internalise nothing else from this document: when you make a change, ask first what its inputs are in the graph and what depends on it. The file the procedure lives in is a hint, not a contract. The graph is the contract.

## The climate cluster, and why the names lie a little

WBH atmosphere, hydrographics, and temperature are mutually dependent. Hydrographics depends on temperature range. Temperature depends on albedo, and albedo reads hydrographics. Atmosphere depends on temperature through runaway-greenhouse mutation, and temperature depends on atmosphere through greenhouse factor. There is no single-pass evaluation order that gives the right answer; the original handbook implicitly assumes the referee will derive provisional values, then refine.

The code calls this cluster's solver `ApplyClimatePasses` (`worlds/climate.go`). It runs exactly two passes of (compute temperature → compute partial-geology factors → apply the inherent-temperature addition from TSS → rederive atmosphere and hydrographics from the post-TSS temperature). The second pass's result is trusted.

The design history matters here because the names still partly reflect an earlier story. The original pass-2 design framed this as a **fixed-point solver**: iterate until atmosphere code, hydrographics code, and mean temperature stabilise; cap iterations and assert convergence. Cycle 17 implemented that, immediately hit overflow on common seeds, and an instrumentation spike found the reason: `RederiveAtmosphereHydrographics` calls `RollHydroDigit`, which **consumes fresh dice each time it is called**. Every pass is a fresh stochastic sample, not a convergence step. There is no fixed point to find — the hydrographics outcome is a probability distribution over a band, and successive samples can disagree even when nothing else has changed.

The code then reverted to the pass-1 behaviour (two passes, trust the second) and renamed `ConvergeClimate` to `ApplyClimatePasses` in commit `041704d`. The evergreen docs (`api-surface.md`, `dependency-graph.md`, `design-intent.md`, `harness.md`, and this file) and the code comments were swept to match in the C2 naming pass: they now describe two-pass sampling, not a fixed-point solver, and the `Climate` convergence-variable struct they once documented is gone (it was dead code, removed per L14). The old "converge / fixed-point" framing survives only in `history/`, deliberately — `history/lessons-learned.md` § L13 is the canonical account of what actually happened, and it needs the word to tell the story.

The implication for a maintainer: when you see "convergence" or "fixed point" anywhere in the project, mentally substitute "two stochastic samples, second one wins." If you ever try to add a third pass or to assert iteration stability, you will reproduce cycle 17's overflow. The function is named for what it does, not for what an idealised version of it would do.

The cluster also folds the geology factors that the temperature depends on (residual seismic stress, tidal-stress factor, tidal-heating factor) into each pass via `computePartialGeology`. Those factors are atmosphere- and hydrographics-independent; what they need is body-physical and orbital state, all of which is settled before climate runs. The result is that `body.Geology` is partially populated by the end of climate (TSS factors and inherent temperature in place), and tectonic plates and gas-giant residual heat are filled in by `ApplyGeology` afterwards.

## The unified Body and the iterator

Every placed thing — terrestrial planet, gas giant, planetoid belt, moon — is a `Body` (`worlds/body.go`). Moons live in `Body.Children []*Body`, parented to their planet via `Body.Parent`. The kind is encoded as `BodyKind`, not as separate types. The iterator `Universe.AllBodies()` yields each body and then its children, in placement order; `Universe.AllBodiesWithParent()` adds the parent for procedures that need parent context. The "moon path" is not a separate code path — it is yielded by the same iterator.

The trade-off this introduces, and the one a new maintainer needs to remember:

- For planets and belts, `Body.Orbit` is the orbit around the star.
- For moons, `Body.Orbit` is unset; `Body.OrbitPD` and `Body.OrbitKm` carry the moon's orbit around its parent. The orbit-around-the-star for a moon is `Body.Parent.Orbit`.
- `Body.StellarOrbit()` returns the orbit-around-the-star for any body, doing the right indirection.
- `Body.Host()` returns the body that orbits a star directly — the moon's parent for moons, the body itself otherwise. Use it whenever you need "the thing whose HZ tag, stellar orbit, or parent mass governs this procedure" without re-implementing the moon-vs-planet branch.
- `Body.MassOrDerived()` returns the body's mass in Earth masses, preferring `MassEarth` and falling back to a density-times-volume derivation. This subsumes three open-coded copies that previously existed in different stages.

A field on `Body` is a pointer (e.g. `*Atmosphere`, `*Geology`) when its presence is conditional on the stage having applied to this body. The `Has*()` predicates wrap the nil checks: `body.HasAtmosphere()`, `body.HasGeology()`. Use them; don't deref blindly. The pointer-vs-nil shape is the project's chosen way of saying "this kind of body doesn't get this kind of state."

When you introduce a new procedure, the litmus test is: does it run for every body the iterator yields, including moons of gas-giant parents? If yes, it iterates `u.AllBodies()` and switches on `Body.Kind` where the rules differ. If no, the exclusion criterion is explicit. The pre-flight checklist for this lives at `anti-patterns.md` § A.1 and is non-negotiable; it is the type-level closure of the pass-1 silent-zero recurrence.

## The Apply mutation discipline

The pipeline shape is mutation at the stage level, immutability inside procedures. `Apply*` walks the universe and writes to `*Body` fields in place; `Compute*`, `Derive*`, and `Roll*` take values, return values, and never reach into shared state. This is deliberate — pass 2 considered returning a fresh `Universe` per stage and rejected it, on grounds that the universe carries mutable per-body state through ten stages and the copy-per-stage tax buys nothing semantic.

The doc-comment on every `Apply*` cites the WBH page or the dependency-graph stage that requires the mutation. The convention is that you can read `worlds/generate.go` top-to-bottom and see the entire pipeline as a sequence of mutating stage applications followed by deterministic aggregations:

```go
ApplyDetailFrontEnd → ApplyBodyPhysical → ApplyBeltDetails → ApplyMoonRefinement →
ApplyRotationTilt → ApplyClimate → ApplyTaintTypology → ApplySurfaceDistribution →
ApplyGeology → ApplyBiology →
ApplyHabitability(no roller) → AggregateSystem(pure) → BuildIISSForms(pure)
```

Within those stages, function shapes follow a small naming protocol that you should treat as binding:

- `Generate*` rolls a sub-system from a Roller, returns a value plus an error.
- `Roll*` rolls a single value, returns it plus an error (the error is for dice-exhaustion in Scripted tests, primarily).
- `Compute*` is deterministic, returns a value with no error.
- `Derive*` is a pure formula — no Roller, no error.
- `Apply*` mutates a target. The mutation point is documented; the post-condition is tested.

These are not aesthetic. `Compute*` previously coexisted with `Roll*` confusingly; the project renamed roll-bearing computations so the discriminator is dice consumption, not vibes. If you find yourself writing a function whose name does not fit the protocol, the more likely diagnosis is that the function is doing too much, not that the protocol is wrong.

## The IISS package boundary

`iiss/` (the package, not the in-fiction service) is the boundary between system construction and rendering. The form types — `Class0IForm`, `Class23Form`, `Class4PForm` (and its `Class4PPartP` / `Class4PPartPB` bodies), `SystemForms`, `FormHeader` — live there. The Markdown renderers (`MarkdownClass0I`, `MarkdownClass23`, `MarkdownClass4P`, `MarkdownSystem`) live there. **`iiss/` does not import `worlds/`.** The boundary type passed across the seam is `iiss.SystemForms`, populated by `worlds.BuildIISSForms` and consumed by the renderers.

This is principled and pays off — package boundary decisions are cheap during design and expensive during implementation, and pass 2 made the decision before any code shipped. The cycle never had to surface.

Class IV-P used to be the one asymmetry: its body was too domain-specific to render without worlds types, so `Class4PForm` carried a `RenderBody func(*strings.Builder, FormHeader)` closure pointing back into worlds, and `any`-typed `PartP`/`PartPB` fields (issue #48). **C3 removed that.** The Class IV-P body structs (`Class4PPartP`, `Class4PPartPB`, and their sub-blocks) and their `RenderBody` methods now live in `iiss/` (`iiss/class4p.go`); `Class4PForm` holds concrete `*Class4PPartP` / `*Class4PPartPB` fields — no closure, no `any`.

So all three forms are fully owned by `iiss/` and marshal to JSON as concrete types. The worlds→iiss boundary is uniform: `worlds` **builds** the forms (`buildClass4PPlanet` / `buildClass4PBelt` in `worlds/iiss_class4p.go` read the `Universe` and fill the iiss structs), `iiss` **renders** them. The renderer-symmetry claim in `api-surface.md` § A.2 is now literally true for all three.

## What sits outside the gate

The project commits to WBH chapters 1–3 (pp.14–146). It does not commit to WBH pp.147+ (World Social Characteristics, Special Circumstances). It also does not commit to detailed special-object physics (white dwarf interiors, neutron-star magnetic fields, etc.). Those decisions are not just scope cuts — they are encoded in the type system and the error model.

When a generation path hits an out-of-scope condition, it does not silently degrade and it does not silently substitute a placeholder. It returns a typed error rooted at `stars.ErrSpecialCircumstances` with a wrapped per-site sentinel. The list includes `ErrPostStellarPrimaryUnsupported`, peculiar-primary dispatch errors, giant-companion-MAO gaps, and missing class-IV table cells. Pass 2's property tests and bulk sweeps run an `isSpecialCircumstances(err)` predicate that recognises this family and skips the seeds — about 3–4% of random seeds hit one. The 10 000-seed bulk sweep run during pass 2's "clean every run" project (`history/plan-clean-every-run.md`) reports 10 000 successes precisely because the WBH-provided Referee options for those edge cases were adopted, narrowing the out-of-scope hits to genuinely out-of-scope categories.

If you ever need to add a new out-of-scope guard, the pattern is `errors.New` for a new sentinel, `fmt.Errorf("%w: ...", ErrSpecialCircumstances, ...)` wrapping at the call site, and an entry in `isSpecialCircumstances`. Do not pattern-match on error messages. Do not catch and continue; let the error propagate to the property test or the CLI.

## The six committed inconsistencies

`wbh-inconsistencies.md` enumerates six places where the book contradicts itself. The project commits to one interpretation per case, with the interpretation encoded in code, asserted by a test marked ⚠️ in the harness, and explained in the doc. The summary table:

1. **HZCO formula vs p.42 table.** Five Class VI cells disagree by >5%. Code follows the formula; tests skip the five cells.
2. **Aab IV-d (Zed Prime) sizing.** P.58 sizing table says S; p.63 form (the canonical verification target) says Size 5. Code follows the form.
3. **Three temperature-chapter divergences.** Albedo Hyd 6+ (follow table), Terra reference greenhouse (follow book's 0.36 even though it doesn't match real Earth), Zed Prime WorstLow (compute consistently, accept divergence from sidebar).
4. **Residual seismic stress density DM.** Table says +2; worked example uses +1. Code follows the table.
5. **Compatibility "+3" addend.** Worked example shows an unsourced +3; formula box doesn't. Code follows the formula. This is the only case where the chosen interpretation does _not_ reproduce the canonical Zed Prime form value — the doc explains why (encoding the +3 would require a magic number with no procedural justification).
6. **Habitability gravity DM overlap.** Two bands overlap; footnote says "use worst at edges" but the worked example uses the narrower band. Code follows the worked example because that reproduces the form's Habitability=7.

The decision-rule heuristic across the six is "follow whichever interpretation reproduces the canonical Zed Prime IISS form on pp.141–142, unless doing so requires encoding an unsourced magic number." Entry 5 is the exception that makes this a heuristic and not a rule.

A maintainer's job, when surfacing a seventh inconsistency: do not pick a side in code without writing the seventh entry in this doc. The audit trail is the discipline.

## The four-layer test discipline

Per-procedure dice-script tests, plus property tests, plus a Markdown regression baseline, plus a 10 000-seed bulk sweep. These exist for different reasons and they catch different things; do not collapse the layers.

- **Per-procedure tests.** One per `Roll*` / `Generate*` / `Compute*`. Driven by `Scripted` rollers with the book's narrated dice. Assert outputs to the digit. These are the proof of fidelity; many predate pass 2 and were ported verbatim because the architecture didn't change what the procedure does, only how it's orchestrated.
- **Property tests.** Eight invariants × 1000 seeds each (`worlds/property_test.go`). They catch silent-zero / silent-skip across the population — the class of bug that per-procedure tests miss because each per-procedure test only sees one fixture. `TestProperty_MoonsHaveBodies` is the sentinel for anti-pattern A.1; `TestProperty_HZBodyHasClimate` catches "did atmosphere/hydro/temp populate for HZ moons too" regressions.
- **Façade fixtures.** `Generate(seed)` over 100 seeds with shape-invariant assertions (`Sol/Generate`, `Zed/Generate`). Pass 2 deliberately does **not** keep a full-pipeline gold script — pass 1's `TestZed_FullDetail` died when the pipeline reordered, and `history/spike-findings.md` § Finding 2 is the canonical statement of why this kind of fixture is anti-gold: it claims fidelity but actually asserts pipeline ordering. Pass 2's façade fixtures assert shape (no `?` in SAH triplets, every body has rotation state, mainworld picked) instead.
- **Markdown regression baseline.** Five seed snapshots at `iiss/testdata/seed_*.md`. Catches unintentional drift in renderer output. Refreshed deliberately with `-update.regression` after a reviewed change.
- **Bulk sweep.** A one-off `cmd/world-builder-bulk` runner (not in the current source tree; lived in a branch during pass 2's clean-every-run project) verifies 10 000 random seeds all produce real, fully-formed systems with no errors outside the Special-Circumstances family. The result that lets v1.0 ship.

If you introduce a new procedure, the per-procedure test is mandatory (it's where the book is reproduced); the property test is mandatory if the procedure is per-body (it's where the silent-zero is caught); the regression baseline is updated if the procedure changes any visible output.

## What changes easily, what doesn't

**Changes the system is shaped to accommodate:**

- _Adding a procedure inside an existing WBH section._ Port or write the function in the appropriate file, with WBH page citations in the doc-comment. Add a per-procedure dice-script test. If it's per-body, slot it into the relevant `Apply*` orchestrator's body loop; if it's per-allocation or per-system, slot it into the relevant aggregation. The unified `Body` and the iterator make the per-body case mostly mechanical.
- _Fixing a misinterpretation of a procedure._ Change the function, change the fixture, run `task`. Worked-example tests catch correctness regressions immediately.
- _Adding a seventh book-internal inconsistency._ Add the entry to `wbh-inconsistencies.md` with both sources cited, encode the interpretation in code with a doc-comment, flip the fixture status to ⚠️.
- _Adjusting renderer output for Class 0/I or Class II/III._ Edit `iiss/render.go`. Refresh the regression baseline if the change is intentional.
- _Adding a new output format._ Add `JSONClass0I`, `XMLClass0I`, etc., as new sibling functions in `iiss/`. No interface required — that's what `design-intent.md` § "Stop rules" rejects.

**Changes that require rethinking something fundamental:**

- _Reordering the pipeline._ The dependency graph is the invariant. If your change creates a back-edge from a later stage into an earlier one (the way TSS feeds back into temperature, the way climate would feed back into Stage 4 surface distribution if it weren't deferred), you are entering the territory of `dependency-graph.md` § "TSS back-edge" and need to decide whether to fold the new edge into the climate solver or accept a one-shot update with explicit drift bounds.
- _Replacing the Roller seam._ Every procedure consumes dice through `roller.Roller`. An effect system, monadic IO, or context-threaded RNG would be a different project. The `Scripted`-with-book-narrated-dice testing strategy is the highest-leverage thing in the codebase; touching the Roller breaks all of it.
- _Splitting moons back out of Body._ The unified iterator and the `Has*()` shape are the type-level closure of the silent-zero class. The cost of re-introducing a separate moon code path is well-documented (pass 1's four-times recurrence). Don't.
- _Bringing WBH pp.147+ into scope._ The `ErrSpecialCircumstances` family becomes a generation path. New domain entities (governments, trade codes, sophont social characteristics) appear. The shape of `Body` is fine, but the rendering surface (a fourth IISS form? a separate "society sheet"?) is open design. This is the most plausible category of "next major scope" and is post-v1 polish per `next-steps.md` if motivated.
- _Replacing per-procedure dice-script tests with full-pipeline gold scripts._ This was tried in pass 1 and abandoned. Read `history/spike-findings.md` § Finding 2 before you propose it.

A new maintainer with the theory will instinctively reach for the dependency graph and the iterator-plus-Has-pointer shape when adding behaviour. A new maintainer without the theory will reach for the chapter in the book, find the corresponding file by name match, and write code that iterates planets only or that reads a pre-climate value when it needs a post-climate one. The anti-patterns catalogue (`anti-patterns.md`) is the explicit list of the latter mistakes; A.1 (moon-path silent-zero) and A.8 (stale pre-climate inputs) are the two that recurred most.

## Where the theory is thin

These are the places where the code and the documentation are not in perfect alignment, where a maintainer should expect to update something to keep the picture clean.

- **The climate-solver vocabulary (was a lag; resolved in C2).** The evergreen docs and code comments now describe two-pass sampling consistently and no longer name `ConvergeClimate` or a "fixed-point solver". The old framing survives only in `history/` (intentionally — dated records; `history/lessons-learned.md` § L13 is the canonical account of why the convergence framing was wrong). If you find "converge" / "fixed-point" language in an evergreen doc or a code comment, treat it as a regression and fix it — substitute "two stochastic samples, second wins."
- **`stage*.go` file numbering (resolved in C5).** The `worlds/stage2.go`…`stage10.go` filenames are gone. Orchestrators now live in role-named files (`detail_frontend.go`, `physical_detail.go`, `rotation_tilt.go`, `climate.go`, `taint_surface.go`, `aggregate.go`) or alongside their feature's procedures (`geology.go`, `biology.go`, `habitability.go` each hold the `Apply*` pass). The dependency-graph _stage numbers_ still appear in doc comments and test names, where they correctly denote graph indices, not WBH chapters.
- **Class IV-P closure boundary (resolved in C3).** The `RenderBody` closure and `any`-typed `PartP`/`PartPB` are gone: the Class IV-P body structs and renderers live in `iiss/class4p.go`, and `Class4PForm` holds concrete `*Class4PPartP` / `*Class4PPartPB`. `api-surface.md`'s "all three forms return typed structs" is now literally true.
- **The bulk-sweep tool's location.** `harness.md` describes a `cmd/world-builder-bulk` runner producing 10 000 successes. That binary is not in `main` (`cmd/world-builder` is the only command). The sweep was the v1.0 readiness gate, run from a branch. If you need to re-run a bulk sweep against a current build, you will be re-creating the tool from git history (commit `328b37c`) or writing it fresh.
- **Pass 1 vs pass 2 in the docs.** `history/` contains pass-1 specs, plans, and a retrospective; the rest of `docs/` describes pass 2. The README mentions pass 1 only to point at the `pass-1-final` tag for archaeology. Where a doc-comment references "pass 1's behaviour," it is documenting a deliberate decision to match or deliberately deviate from a historical implementation. Read it as historical context, not as a TODO.
- **`MAO` now lives in `stars` (resolved in C4).** The p.39 MAO table and `stars.MAO(Star)` are in `stars/mao.go`; `GenerateSystem` computes the companion-of-giant orbit (WBH p.27) directly, and `GenerateSystemOpts` no longer carries an MAO callback. `stars/` has no inbound dependency on `worlds/`; `worlds` (available-orbits) imports `stars.MAO` / `stars.LacksP39MAORow`.

## Uncertainties and where I might be wrong

The theory above is built from the code, the design docs, and the commit history of `main`. Things I am inferring rather than knowing:

- I am inferring that the project's authors treat the heuristic in `wbh-inconsistencies.md` ("trust whichever interpretation reproduces the canonical Zed Prime form") as a decision rule of last resort, with the formula-vs-form judgment call decided case by case. The doc says as much, but the rule is also stated forcefully enough that someone might apply it mechanically. If a seventh inconsistency arrives where both interpretations reproduce some part of the form and neither reproduces all of it, the project will need to make a judgment, not consult the rule.
- I am inferring that the `BodyEmpty` constant exists as a sentinel rather than as a placed object — `dependency-graph.md` and various procedure no-ops treat empty slots that way. I have not exhaustively traced what creates `BodyEmpty` versus what filters it out. If empty slots leak into a renderer, behaviour is probably "skip silently"; this is not load-bearing for current renderers but might bite a future one.
- The 10 000-seed clean-sweep result is reported in memory entries and in `next-steps.md`; the property tests in the current source run 1000 seeds each. A maintainer who wants to verify "is the system really clean today" should re-run the sweep, not just `task test`.
- MAO now lives entirely in `stars` (C4); there is no longer a worlds↔stars split. Change the formula in `stars/mao.go`.

These are the places where I would expect a reader who runs the code to find a small surprise. They are not bugs; they are areas where the theory is provisional pending a deeper read.

## A maintainer's posture

Hold this picture loosely as a single shape:

- A book that prescribes procedures; a generator that runs those procedures.
- A seed that names the run; a Roller that gates randomness so the naming sticks.
- A graph of values, mostly forward-only, with one cycle that the code samples twice and trusts.
- A single body type walked by one iterator; pointers and `Has*()` for stage-conditional state.
- Three IISS forms rendered out of one shared aggregate, each fully owned by `iiss/` as concrete structs.
- A test suite whose primary assertion is "the book is reproduced," supplemented by population-level invariants and a regression baseline.
- A scope edge encoded as typed errors, not silent fallbacks.

When you change something, the questions to ask in order are: where is this value in the graph; does it run for every body the iterator yields; what does the book say at the cited pages; what fixture proves the chosen interpretation; what changes downstream of this stage. If you can answer those, the change will be safe. If you can't, the docs in `design-intent.md` and `dependency-graph.md` will tell you which question you're not yet positioned to answer.

The single most important habit: when you see "convergence" in older docs, mentally substitute "two stochastic samples, second wins." When you see "for each planet" in any new procedure you're writing, stop and write "for each body" instead. Those two substitutions cover most of where pass 1 went wrong and where pass 2 spent its design budget.
