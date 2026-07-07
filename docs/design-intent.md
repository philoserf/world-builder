# Design Intent

This document is the architectural foundation of the project. It records the shape of the codebase, the rules that keep it from drifting, and the WBH-fidelity decisions baked in.

The historical context — why the architecture looks this way rather than mirroring WBH pagination, what an earlier WBH-pagination-shaped implementation taught — lives in `history/lessons-learned.md` and the pass-1 retrospective trio under `history/pass-1-retrospective/`. The decisions below are committed in code.

## The structural choice

The book's pagination is not the data's dependency structure. A WBH-pagination-shaped implementation produces:

- A cyclic climate cluster (atmosphere ↔ temperature ↔ hydrographics) discovered mid-flight, recovered awkwardly with a two-pass rederive.
- A moon-path silent-zero anti-pattern that recurs every time per-body procedures aren't first-class.
- Renderer asymmetries (one form returning a struct, another returning a string) when forms are designed chapter-by-chapter.
- API gotchas (misnamed parameters, signature inconsistencies) that slip past chapter-level reviews because no chapter owns the API surface as a whole.
- Repeated "this chapter pulls forward from earlier chapters and pushes back into later ones" — pagination doesn't cleave at dependency boundaries.

The project inverts the relationship: **the data dependency graph determines structure; worked-example fixtures determine the gate; the book becomes a citation system, not an architecture.**

## What the project commits to

### Single deterministic path through the source

The code implements **one** path through WBH pp.14–146. Concretely cut from the public surface:

- **`*Opts` variance fields — partial.** `stars.GenerateSystemOpts.WithVariance` and `.Accuracy` stay load-bearing for worked-example tests that drive the book's narrated dice scripts. `worlds`-side variance fields and `AccurateAlbedo` are cut.
- **All optional rules from WBH.** Rare Earth Universe Variant, optional any-oxygen-atm = biomass ≥ 1, the Insidious DE hazard rule's optional branch, and any other section the book labels "Optional."
- **Method-of-method choices.** Where WBH offers two procedures for the same value, the code picks one. Default heuristic: prefer the formula or table over the roll when both are given (less interpretation latitude).
- **Toggles for book inconsistencies.** Each of the six documented divergences gets one chosen interpretation in code with rationale in `wbh-inconsistencies.md`. No runtime switch.
- **Referee override flags.** No `-mainworld <designation>` flag, no Opts-driven mainworld override.

The API surface is compact. Most `Generate*` functions are `(inputs..., r Roller) → result` with no Opts struct.

Several cuts are Referee-facing — Rare Earth Universe Variant, the optional any-oxygen-atm biomass floor, the `-mainworld <designation>` override — and exist in WBH because campaigns differ. They are post-v1 polish candidates if vetting motivates them; see `next-steps.md` § C2.

### Full API surface designed as a unit

Every public signature is owned by `api-surface.md` and reviewed against the full surface, not by chapter. Contract tests for misuse paths are mandatory — a misnamed parameter that compiles is caught by a misuse test, not by hoping reviewers spot it. Renderers return typed structs from day one: `RenderClass0I`, `RenderClass23`, `RenderClass4P` are sibling functions over sibling structs; Markdown / JSON rendering are separate consumers. No string-asymmetric renderer.

### Per-body procedures are first-class

One iterator walks bodies (planets, moons, belt members) uniformly via `Universe.AllBodies()`. There is no separate moon code path. The moon-path silent-zero bug is a class of bug that can't fire. Bodies are walked once; procedures take a `BodyKind` parameter where rules differ.

### Worked-example fixtures + property tests + bulk sweep as the gate

Every WBH-narrated dice script lands as a per-procedure test that asserts to the digit. Property tests over 1000 seeds catch silent-zero / silent-skip across the population. A 10 000-seed sweep verifies every invocation produces a real, fully-formed system. See `harness.md` for the catalog and the four-layer coverage strategy.

### Topic-named documents

Evergreen docs at `docs/` root are topic-named (`anti-patterns.md`, `wbh-inconsistencies.md`). No chapter-numbered or date-prefixed names. Historical artifacts (the chapter-numbered pass-1 plans/specs/retrospective) live under `docs/history/`.

## Conventions inherited

Decisions baked in from the start:

- **Go, not Python.** Static-binary distribution + compiled-in literal tables + native typing beats dynamic flexibility for this data shape.
- **Deterministic Roller as the load-bearing seam.** `roller.Roller` with `Seeded`/`Scripted`/`Fixed` impls. No package-level RNG anywhere. Seed plus options fully determines a system. `Scripted` panics on exhaustion.
- **Worked-example regression tests — per-procedure form.** Encoded with `roller.NewScripted(...)` driven by the book's exact dice; assert every output to the digit. Per `history/spike-findings.md` § Finding 2, full-pipeline gold scripts do not survive pipeline change and are not gold; the project uses per-procedure `Scripted` gold for narrow fixtures + `Seeded` + shape invariants for façade fixtures.
- **Tables as Go literals with WBH page citations.** `*float64` for nullable cells (the book's "—"). Doc-comments cite the page. No external YAML/TOML data files.
- **Brainstorm → spec → plan → TDD with subagent review.** Per-task two-stage review (spec compliance + code quality).
- **Modernizer-as-mandatory gate.** `task check` runs `go fix ./...` first and fails on any diff. Idiomatic Go 1.21+/1.22+ stays current.
- **Surfacing book inconsistencies as data, not bugs.** Six findings consolidated; the project commits to one interpretation per finding in code, with the audit trail (citations + chosen rationale) in `wbh-inconsistencies.md`.

## The six design docs at docs/ root

These documents are the design backbone. Together they define what the code commits to.

1. **`design-intent.md`** (this file) — the why and the cuts.
2. **`dependency-graph.md`** — every value, what it depends on, where the one cyclic (climate) cluster is.
3. **`api-surface.md`** — every public signature with rationale.
4. **`wbh-inconsistencies.md`** — six book-internal divergences, with the chosen interpretation per case.
5. **`anti-patterns.md`** — failure modes the code guards against.
6. **`harness.md`** — fixture catalog with status indicators.

## Stop rules — abstractions to keep off the shelf

Brooks' second-system warning applies indefinitely. Specific cleverness traps to refuse:

- **Generic procedure framework.** A `Procedure interface { Inputs []FieldRef; Outputs []FieldRef; Apply(ctx) }` with every WBH step as a Procedure. Tempting; flattens the type system. Rejected.
- **DAG executor.** Encoding the dependency graph as runtime metadata with topological sort. Tempting; loses compiler help. The graph is design documentation, not runtime data.
- **Effect system.** Making rolls and IO explicit in types. Too much ceremony for Go.
- **Renderer interface.** A uniform `Renderer{ToMarkdown, ToJSON, ToPlainText}` over forms. Sibling functions are fine.

**Stop rule:** any abstraction not justified by an existing call-site needing it goes back on the shelf.

## Deferred items

Named explicitly so they cannot quietly become "we should never do this" by absence. Status as of v1.0:

- **Notable Features Markdown block** — **shipped in v1.0** (commit `89cbcd5`). A referee-facing summary above the IISS forms: tidal-lock zones, WorstLow cold snaps, high-gravity/high-atm crush worlds, taint chains, mainworld habitability rationale.
- **Referee knobs.** Rare Earth Universe Variant, optional biomass floor, optional Insidious DE branch, `-mainworld <designation>` override. WBH ships these because campaigns differ; candidates for post-v1 polish (`next-steps.md` § C2).
- **Special-object detail.** Brown Dwarf, White Dwarf, Neutron Star, Black Hole, Pulsar, Nebula, Protostar, Star Cluster, Anomaly currently get minimum-useful values (kind / mass / age). Detailed physics — accretion, degenerate-matter equations — is post-v1 polish if motivated (`next-steps.md` § C4).
- **Belt-mainworld worked example.** No canonical WBH example exists; closed won't-fix as GitHub #51. The PART P.B renderer ships and works structurally.
