# Pass 2 — Design Intent

This document is the foundation pass-2 hangs off. It records the shape of the rebuild, what we're keeping from pass 1, what we're cutting, and the rules that keep the rebuild from sliding into a worse architecture than the one it replaces.

Read alongside the pass-1 retrospective (`docs/pass-1/retrospective/2026-05-06-lessons-learned.md`, `2026-05-06-pass-2.md`, `2026-05-06-path-forward.md`) and the working-memory entries the pass-1 implementation accumulated. Pass 1 is the reference implementation; main retains it. Pass 2 is greenfield on a long-lived branch and merges back only when behavior parity is reached.

## The structural error pass 2 corrects

Pass 1 confused **where the procedures live in the book** with **what depends on what**. The book's pagination is not the data's dependency structure. Following pagination produced:

- A fixed-point system (atmosphere ↔ temperature ↔ hydrographics) discovered mid-flight and recovered with a "rederive" sub-project.
- A moon-path silent-zero anti-pattern that recurred four times because per-body procedures weren't first-class.
- A late-arriving Class IV-P renderer that returned a string while siblings returned structs, forcing a Q1 redesign during the Markdown sub-project.
- API gotchas (RollGasMix's misnamed parameter, Roller's string-vs-int signature) that slipped past 11 chapter-level reviews because no chapter owned the API surface as a whole.
- Repeated "this chapter pulls forward pages from earlier chapters and pushes back into later ones" — pagination doesn't cleave at dependency boundaries.

Pass 2 inverts the relationship: **the data dependency graph determines structure; worked-example fixtures determine the acceptance gate; the book becomes a citation system, not an architecture.**

## What changes

### Single deterministic path through the source

Pass 2 implements **one** path through WBH pp.14–146. Concretely cut:

- **`*Opts` variance fields — partial.** Aspirational cut walked back during cycle-13 resolution (`next-steps.md` § A3): `stars.GenerateSystemOpts.WithVariance` stays because pass-1 worked-example tests rely on it to drive the simple-roll path the book's narrated dice use. `worlds`-side variance fields (the ones that toggled "use richer tables instead of just a roll") are still cut.
- **`*Opts` accuracy fields — partial.** Same resolution: `stars.GenerateSystemOpts.Accuracy` stays load-bearing for the ~10 `Accuracy: 1` worked-example tests (book uses the simple-1D age path). `AccurateAlbedo` and the opt-in oxygen-atm biomass floor remain cut.
- **All optional rules from WBH.** Rare Earth Universe Variant, optional any-oxygen-atm = biomass ≥ 1, the Insidious DE hazard rule's optional branch, and any other section the book labels "Optional."
- **Method-of-method choices.** Where WBH offers two procedures for the same value, pass 2 picks one. Default heuristic: prefer the formula or table over the roll when both are given (less interpretation latitude).
- **Toggles for book inconsistencies.** Each of the six documented divergences gets one chosen interpretation in code with rationale in `wbh-inconsistencies.md`. No runtime switch. No `t.Logf` divergence noise.
- **Referee override flags.** No `-mainworld <designation>` flag, no Opts-driven mainworld override.

The API surface shrinks dramatically. Most `Generate*` functions become `(inputs..., r Roller) → result`, no Opts struct. That's a much smaller surface to design up front and to test for misuse.

**The cuts are pre-parity discipline, not permanent design.** Several cuts are referee-facing — Rare Earth Universe Variant, the optional any-oxygen-atm biomass floor, the `-mainworld <designation>` override — and exist in WBH because campaigns differ. They are candidates for post-merge polish items if vetting motivates them. (Originally framed as a "Pass-3 referee knobs" cycle; that framing is retired since there is no pass 3 — pass 2 is the end-state design.)

### Full stub interface designed up front

Every public signature is written and committed as a stub before any implementation. That includes:

- All `Generate*` and per-procedure functions across `stars/` and `worlds/`.
- All public types and their accessors (`Body`, `Universe`, `SystemDetail`, the three IISS form structs, `Has*()` predicates).
- All renderers, returning typed structs from day one — **no string asymmetry like pass 1's IV-P.** `RenderClass0I`, `RenderClass23`, `RenderClass4P` are sibling functions over sibling structs, with `Markdown`/`JSON`/`PlainText` rendering as separate consumers.
- All `Roller`-consuming entry points, with documented dice expectations.

Contract tests for misuse paths are written against the stubs. Pass 1's RollGasMix bug should have been impossible: a contract test for "if you pass a Subtype, you get a misuse error or compile error" would have caught it on the first try. Misuse-path tests are mandatory, not aspirational.

If a signature can't be designed without the implementation in hand, that's the signal we don't understand the procedure yet — go back to the book.

### Per-body procedures are first-class

One `applyBodyProcedures(body)` function runs against any `Body` (planet, moon, belt member). The moon-path silent-zero bug becomes a class of bug that can't fire because there is no separate moon code path. Bodies are walked once; each procedure is parameterized over body-kind where the rules differ.

### Worked-example fixtures are the gate

Every WBH worked example becomes a failing dice-scripted fixture **before any generation code is written.** The harness is the spec. Implementation order is "whichever fixture is closest to green next." Sub-project size is "make 1–3 fixtures green," not "implement chapter N."

Pass 1 carried Zed Prime through every chapter as an end-to-end acceptance gate; that pattern survives. Pass 2 adds: every other in-line WBH example, encoded with the same `roller.NewScripted` discipline. The non-Zed examples were not load-bearing in pass 1 and are part of why some chapter-internal contradictions weren't caught earlier.

### Unnumbered, topic-named documents

Pass-1 docs were dated and chapter-numbered (`2026-05-04-world-physical-3a2b-rederive.md`). Pass-2 docs are topic-named (`atmosphere.md`, `temperature.md`, `habitability.md`). Order is git history; identity is topic. There are no sub-project dates or numbers in filenames.

## What pass 1 got right and pass 2 keeps verbatim

Lifted from `docs/pass-1/retrospective/2026-05-06-pass-2.md` § "Decisions worth keeping verbatim":

- **Go, not Python.** Static-binary distribution + compiled-in literal tables + native typing beats dynamic flexibility for this data shape.
- **Deterministic Roller as the load-bearing seam.** `roller.Roller` with `Seeded`/`Scripted`/`Fixed` impls. No package-level RNG anywhere. Seed plus options fully determines a system. `Scripted` panics on exhaustion. **Port `roller/` and `dice/` verbatim.**
- **Worked-example regression tests — the per-procedure form.** Encoded with `roller.NewScripted(...)` driven by the book's exact dice; assert every output to the digit. **These port verbatim from pass-1 source as long as procedure signatures are unchanged.** Per `spike-findings.md` § Finding 2, full-pipeline gold scripts (e.g., pass-1's `TestZed_FullDetail`) do not survive pipeline change and are not gold; pass 1 itself superseded its full-pipeline gold script with a `Seeded` + shape-invariant successor (`TestZed_FullDetail_3A2b`). Pass 2 inherits both patterns: per-procedure `Scripted` gold for narrow fixtures; `Seeded` + shape invariants for façade fixtures.
- **Tables as Go literals with WBH page citations.** `*float64` for nullable cells (the book's "—"). Doc-comments cite the page. No external YAML/TOML data files.
- **Brainstorm → spec → plan → TDD with subagent review.** Per-task two-stage review (spec compliance + code quality) caught real bugs across pass 1. Workflow stays.
- **Modernizer-as-mandatory gate.** `task check` runs `go fix ./...` first and fails on any diff. Idiomatic Go 1.21+/1.22+ stays current.
- **Surfacing book inconsistencies as data, not bugs.** Six pass-1 findings preserved; pass 2 commits to one interpretation per finding in code, but the audit trail (citations + chosen rationale) lives in `wbh-inconsistencies.md`.

## Artifacts, in order

These six documents under `docs/pass-2/` are the design backbone. They get drafted before any new generation code.

1. **`design-intent.md`** (this file) — the why and the cuts.
2. **`dependency-graph.md`** — every value, what it depends on, where the fixed-point loops are. The structural spine. Fixed-point loops are marked explicitly with the convergence pattern documented.
3. **`api-surface.md`** — every public signature enumerated and justified. The actual stubs live in `stars/` and `worlds/`; this doc is the index and the design rationale.
4. **`wbh-inconsistencies.md`** — six pass-1 findings consolidated. Each: WBH page reference, divergent sources, chosen interpretation, why.
5. **`anti-patterns.md`** — moon-path silent-zero, embedding-chain depth, plain-text-renderer asymmetry, modernizer-clean-tree workflow, etc., as a pre-flight checklist every sub-project's spec must cite.
6. **`harness.md`** — index of every WBH worked example with citation, dice script, and status (red/green). The actual fixtures live in `*_test.go` files; this doc is the catalog.

Implementation order, after the docs:

1. Port `roller/` and `dice/` verbatim from pass 1.
2. Stub every public signature per `api-surface.md` (compiles).
3. Port the worked-example dice scripts and `Seeded`-shape-invariant fixtures as failing tests (compiles against stubs, tests red). Per `spike-findings.md` § Finding 4, stubs must precede fixtures — Go test files cannot compile against unstubbed types.
4. Implement procedures, driven by which fixture is closest to green next, in dependency-graph order.
5. Per-cycle delivery target: one stage's fixtures green, ~3–7 days. See § Cadence.

## Fidelity gate (not parity)

Pass 2's merge gate is fidelity to the book, not byte-equivalence to pass 1. The TSS fold-in (`dependency-graph.md` § Stage 7) and the surface-distribution-after-converge reordering (`dependency-graph.md` § Pass-2 sequencing) are deliberate corrections that _will_ produce different IISS output from pass 1 for some seeds. Calling that gate "parity" overstates it; calling it "fidelity" names what's actually being asserted.

Pass 2 merges to `main` when:

- Every worked-example fixture passes (post-decision values, per `wbh-inconsistencies.md`).
- The cut list (above) is honored pre-merge — no resurrected variance/accuracy/optional flags during the merge cycle. (Post-parity reintroduction is named below.)

A pass-1-vs-pass-2 byte comparison was originally listed here as a third gate item. It was explicitly **dropped** during pass-2 wrap-up — see `next-steps.md` § A0. Pass-2's design intentionally diverges from pass-1 on multiple axes (TSS fold-into-climate, surface-distribution-after-converge, narrower-band-wins for gravity DM); the within-pass-2 Markdown regression baseline (`iiss/testdata/seed_*.md`) is the working substitute.

Until pass-2 merges, `main` retains pass 1 as shipped working software. There is no pressure to merge.

## Stop rules — no second-system aspirations until parity

Brooks' second-system warning applies. Specific cleverness traps for this project:

- **Generic procedure framework.** A `Procedure interface { Inputs []FieldRef; Outputs []FieldRef; Apply(ctx) }` with every WBH step as a Procedure. Tempting; flattens the type system; rejected.
- **DAG executor.** Encoding the dependency graph as runtime metadata with topological sort. Tempting; loses compiler help; rejected. The graph is design documentation, not runtime data.
- **Effect system.** Making rolls and IO explicit in types. Too much ceremony for Go.
- **Renderer interface.** A uniform `Renderer{ToMarkdown, ToJSON, ToPlainText}` over forms. Aspirational from pass 1's pass-2 doc; **not pre-parity.** Sibling functions are fine.
- **`wbh.Verifier` package as a public API.** Aspirational; not pre-parity.
- **Coverage thresholds in `task check`.** Aspirational; not pre-parity.

**Stop rule:** any abstraction not justified by an existing call-site needing it goes back on the shelf. Aspirational architecture is post-parity work, never pre-parity.

## Post-parity work named

Pass 2 is small on purpose. These items are _not_ cuts forever — they're queued for after the fidelity gate clears. Naming them here means they cannot quietly become "we should never do this" by absence. They are deferred, not refused.

- **Referee knobs.** Rare Earth Universe Variant, optional biomass floor, optional Insidious DE branch, `-mainworld <designation>` override. WBH ships these because campaigns differ; they remain candidates for post-merge polish (`next-steps.md` § C2) but no pass-3 cycle is planned.
- **Notable Features Markdown block.** A referee-facing summary above the IISS forms: tidal-lock zones, WorstLow cold snaps, high-gravity/high-atm crush worlds, taint chains, mainworld habitability rationale. Post-parity sub-project; the IISS forms alone are canon-good but referee-hostile for at-a-glance use.
- **Special-object detail.** Brown Dwarf, White Dwarf, Neutron Star, Black Hole, Pulsar, Nebula, Protostar, Star Cluster, Anomaly currently get minimum-useful values (type/mass/age). Detailed physics — accretion, degenerate-matter equations — is post-merge polish if it's wanted at all.
- **Belt-mainworld worked example.** No canonical WBH example exists; harness defers (`harness.md` § `ZedPrime/Class4P/PartPB`). Post-parity, the belt mainworld branch gets an internally-constructed example fixture so PART P.B does not ship dark.

## Risks named

- **Sunk-cost math.** Pass 1 was ~two weeks of intermittent work, ~30 sub-projects. Pass 2 with the cycle-per-stage cadence (§ Cadence) targets ~12 cycles over ~8–10 weeks. Soft cap: revisit the plan at week 8 if the fidelity gate is not in sight; hard "is this still worth it?" review at week 12.
- **API redesign in isolation.** Pass 1's API emerged from real callers. Pass 2 designs the API up front, but the harness fixtures and `cmd/wbh` are real callers from day one — they're not absent. Keep them as constraints on the design.
- **Worked-example fidelity drift.** The pass-1 dice scripts are gold. They get ported as-is, not re-derived. Re-deriving would re-introduce bugs pass 1 caught.
- **"Smaller chunks" without a number is guilt.** Concrete target: 1–3 worked-example fixtures green per cycle, ~3–7 days each. If a cycle goes longer, the chunking was wrong; split.

## Cadence

The pass-1 cadence (chapter-sized brainstorm → spec → plan → TDD-with-subagent-review) survives in form, with three changes:

- **Cycles are sized by stage, not by WBH pages or by individual fixtures.** A cycle = all fixtures for one dependency-graph stage, ~10–15 fixtures, ~3–7 days. With ten stages plus stub commit + harness commit + Markdown golden, that's roughly 12 cycles for ~8–10 weeks of work. Per `spike-findings.md` § Finding 5, the earlier "1–3 fixtures green per cycle" target gave 17–50 cycles for the full harness — not workable.
- Cycles do **not** carry forward partial state across boundaries. `git status` must be clean at every cycle start. The pass-1 `just`→`task` migration sat as uncommitted state for most of a session and corrupted multiple subagent reports — that does not recur.
- The stub commit and the harness commit are their own cycles (cycles 0 and 1) because both are PR-sized in their own right.

## Provenance

This document distills the conversation across the pass-2 design session. The retrospective trio — `docs/pass-1/retrospective/2026-05-06-{lessons-learned,pass-2,path-forward}.md` — and the working-memory entries on this project are the inputs. Where the post-completion pass-2 doc and the mid-pass strategic reflection diverged (decomposition granularity, harness-first sequencing, dependency-graph-first design, API discipline), pass 2 takes the mid-pass position deliberately.
