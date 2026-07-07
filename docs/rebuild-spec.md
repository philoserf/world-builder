# Rebuild Specification

Specification for a from-scratch third implementation of the physical star-system
rules (WBH pp.14–146). This document is the concrete realization of the "If We
Started Over" retrospective (`history/pass-1-retrospective/2026-05-06-pass-2.md`),
written after the current `main` has been built, refined, and reviewed to v1.0.

## Framing and terminology

The repo already numbers its implementations:

- **Pass 1** — the original WBH-pagination-shaped implementation (tag `pass-1-final`).
- **Pass 2** — the data-dependency re-layering that shipped as `v1.0` (current `main`).

This document reframes **both shipped passes together as the exploratory first pass**:
the work that mapped the domain, discovered where the book contradicts itself, found
the fixed-point cluster the hard way, and closed the moon-path bug class at the type
level. That knowledge is the asset. The code on `main` is the working reference, not
the foundation.

The effort specified here is **the rebuild** — a third implementation, ground-up, in
Go, at the same scope (pp.14–146). It is not an expansion (pp.147+ stays out) and not
a patch of `main`. Where this doc says "the rebuild," read "the true second attempt,
now that we know the shape." Call the eventual tag `rebuild` or `pass-3`; this doc
avoids a bare "pass 2" to prevent collision with the shipped meaning.

**Decisions locked before writing (owner-confirmed):**

1. Language: **Go**. The data shape (many small typed lookup tables, precise numeric
   outputs) fits Go; the procedure files and dice scripts port. Not re-litigated.
2. Scope: **WBH pp.14–146, unchanged**. Cleaner internals, identical behavior surface.
   No social characteristics, no special-object physics, no new output beyond the three
   IISS forms in Markdown/JSON/short.
3. Deliverable: this committed spec.

## What the rebuild must reproduce

The rebuild is a success only if it reproduces the current `main`'s behavior on the
fidelity gate. Concretely:

- Every WBH worked example (Sol/Terra p.35, Zed, Zed Prime, Corella) reproduces to the
  digit under the same interpretation `main` commits to.
- The six committed book-inconsistencies (`wbh-inconsistencies.md`) are carried forward
  **verbatim** — same chosen interpretation, same asserted values, same citations. The
  rebuild does not reopen these decisions; it re-encodes them.
- A 10 000-seed bulk sweep produces 10 000 real, fully-formed systems with zero errors
  outside the `ErrSpecialCircumstances` family.
- Byte-identical output across machines for a given seed (determinism preserved).

The rebuild is free to produce a _different_ dice stream than `main` (the sub-roller
change below guarantees it will), so the Markdown regression baseline is re-generated,
not diffed against `main`. Fidelity is measured against **the book**, not against
`main`'s byte output.

## Invariants carried forward verbatim

These were right in pass 1, right in pass 2, and are not up for redesign. The rebuild
inherits them on day one:

1. **The Roller seam.** Every dice draw passes through `roller.Roller`. No package-level
   RNG anywhere. `Seeded` / `Scripted` / `Fixed`; `Scripted` panics on exhaustion. This
   is the architectural keystone (see change C1 for the one additive extension).
2. **Tables as compiled-in Go literals**, `*float64` for the book's "—", a doc-comment
   citing the WBH page on every table. No external YAML/TOML.
3. **Per-procedure `Scripted` worked-example tests**, driven by the book's narrated dice,
   asserted to the digit. The fidelity gate. These port unchanged — the sub-roller change
   does not touch how an isolated procedure consumes dice.
4. **Unified `Body` + one iterator.** `Body{Kind, Parent, Children}` walked by
   `AllBodies()` / `AllBodiesWithParent()`. There is no separate moon code path; the
   moon-vs-planet distinction is a `Kind` parameter. This closed anti-pattern A.1 at the
   type level (4× recurrence in pass 1 → 0 since). Keep exactly. Do not split moons back
   out.
5. **The `iiss/` package boundary.** Form structs and renderers in `iiss/`; `iiss/` does
   not import `worlds/`; `SystemForms` is the boundary type. (Change C3 removes the one
   leak in the current realization.)
6. **The typed scope-edge error model.** `ErrSpecialCircumstances` umbrella with wrapped
   per-site sentinels; `errors.Is`, never message matching. Out-of-scope paths return a
   typed error, never a silent placeholder.
7. **The naming protocol.** `Generate*` (rolls a sub-system, returns value+error),
   `Roll*` (rolls one value), `Compute*` (deterministic, no error), `Derive*` (pure
   formula), `Apply*` (mutates a `*Universe`/`*Body`, documents the mutation point,
   tests the post-condition). Discriminator = dice consumption.
8. **The six committed inconsistencies + the decision heuristic** ("reproduce the
   canonical Zed Prime IISS form on pp.141–142, unless doing so requires encoding an
   unsourced magic number"). Carried forward as data, not reopened.
9. **The four-layer test discipline**: per-procedure dice-script tests, property tests
   (per-body invariants over 1000 seeds), a Markdown regression baseline, a bulk sweep.
10. **Modernizer-as-mandatory gate**, `task check` = `go fix` clean-tree + gofumpt + vet
    - golangci-lint; `task test` = `go test -race ./...`. Local `task` is the gate; no CI.
11. **The stop rules** from `design-intent.md` § "Stop rules": no generic Procedure
    framework, no runtime DAG executor, no effect system, no uniform Renderer interface.
    Any abstraction not justified by a live call-site stays on the shelf.

## The changes the rebuild makes

Each change states the problem as it exists in `main` today, the proposal, and the
acceptance criterion. Ordered by depth of impact.

### C1 — Sub-roller tree (the load-bearing change)

> **Spike + retrofit passed (branch `c1-subroller`).** The mechanism is validated
> and has been retrofitted onto current `main`'s pipeline rather than waiting for a
> rewrite — see `docs/c1-subroller-plan.md`. Fidelity held (worked-example tests pass
> unchanged, because `Scripted.Fork` is transparent); the whole per-body suffix is now
> position-independent (isolation property test, 40 seeds × 3 shifts). This makes C2's
> cascade local and unblocks survivable full-pipeline gold fixtures. The rest of this
> section is the original spec rationale, preserved.

**Problem.** `worlds/generate.go` threads **one** `roller.Roller` through all eleven
`Apply*` stages in sequence. A single linear dice stream feeds every body in pipeline
order. Three costs follow from this one fact:

- Full-pipeline gold-script fixtures are _anti-gold_: any stage reorder or added roll
  shifts every downstream value, so the fixture asserts pipeline ordering, not book
  fidelity. Pass 1's `TestZed_FullDetail` died exactly this way
  (`history/spike-findings.md` § Finding 2, `lessons-learned.md` § L9).
- The tidal-lock re-eval cascade (C2) has to re-consume dice for one body, and because
  the stream is shared, those extra draws would perturb determinism for every body after
  it — the cascade is tolerable today only because it is confined to a rare population.
- The whole pipeline is fragile to reordering; the dependency graph is enforced socially.

**Proposal.** Extend the `Roller` interface with a deterministic fork:

```go
type Roller interface {
    Roll(notation string) int
    Fork(key string) Roller // child stream, deterministic in (parent, key)
}
```

`Seeded.Fork(key)` derives a child seed by hashing the parent's seed with `key`
(e.g. `seed64 = fnv1a(parentSeed, key)`), returning a fresh `Seeded`. `Scripted.Fork`
returns a child `Scripted` (test-supplied or a panic-on-use stub — per-procedure tests
never fork). `Fixed.Fork` returns itself.

The top-level pipeline keys sub-rollers by **stable body identity** (the orbit
designation, assigned deterministically before any physical roll) and by
**procedure family**:

```
seed
└── body "A II"        r.Fork("A II")
    ├── physical       .Fork("physical")
    ├── rotation-tidal .Fork("rotation-tidal")
    ├── climate        .Fork("climate")
    └── geology-bio    .Fork("geology-bio")
```

**What this buys.** Each `(body, family)` substream is independent. Reordering stages
does not shift any body's draws. Re-running one family for one body (the tidal-lock
cascade) touches only that body's `climate`/`rotation-tidal` substreams — no global
perturbation. **Full-pipeline gold fixtures become survivable**, which restores the
strongest possible fidelity assertion (the whole Zed Prime system, end to end, to the
digit) that pass 1 and pass 2 both had to give up.

**Cost / risk.** Stars generation and system placement run before body identities exist,
so they keep a top-level stream (fork them under fixed keys `"stars"` / `"placement"`).
The keying scheme must be pinned once and never reordered — but that is a single
design decision, not a per-stage discipline.

**Spike-gated.** This is the one genuinely new, unproven idea in the rebuild. The lesson
of L3/L13 (`ConvergeClimate` was specced with a convergence claim that the runtime
disproved) applies directly: **do not commit C1 without a spike.** The spike: implement
`Fork`, key one body's climate substream, and confirm (a) Zed Prime still reproduces to
the digit and (b) adding a throwaway roll to an unrelated stage leaves that body's
values unchanged. Only then adopt C1 as the pipeline's spine.

### C2 — Honest climate model; retire the tidal-lock cascade

**Problem A — the naming lie.** The climate cluster is called convergence / fixed-point
in ~10 docs (`api-surface.md`, `dependency-graph.md`, `harness.md`, …) and the word
survives in `worlds/stage5.go`, `worlds/tidal_lock_reeval.go`, `worlds/property_test.go`,
`worlds/stage7.go`. It is not a fixed point: `RollHydroDigit` draws fresh dice each pass,
so each pass is a stochastic sample (`lessons-learned.md` § L13, `theory.md` § "the names
lie a little"). `main` renamed the function to `ApplyClimatePasses` but the vocabulary
never fully propagated.

**Problem B — the cascade.** WBH p.106's `pressure > 2.5 bar → DM−2` tidal-lock modifier
cannot fire in a single pass, because pressure is a climate-cluster output (Stage 5) and
tidal lock runs in Stage 4. `main` resolves this with `ApplyTidalLockReEval`
(`generate.go:70`): snapshot pre-tidal state, run climate, then for affected bodies
restore the snapshot, re-run tidal lock (now the DM fires), and re-run the entire climate
cluster. That body consumes ~2× the dice of a normal body (`wbh-inconsistencies.md` § 7).
It is the ugliest mechanism in the codebase.

**Proposal.**

- **Name it honestly from day one.** `ApplyClimatePasses`, doc-commented as "two
  stochastic samples of the atmosphere/hydrographics/temperature cluster; the second is
  trusted." No "converge," no "fixed point," anywhere — code or docs. Never add a third
  pass or a stability assertion (both reproduce cycle-17's overflow).
- **Make the tidal-lock cascade a first-class, local step, not a snapshot/restore hack.**
  With C1's per-body substreams, the pressure>2.5 bar re-evaluation is a clean re-draw
  from that one body's `rotation-tidal` and `climate` substreams — no global snapshot,
  no perturbation of other bodies. Model it as a small explicit "tidal-lock depends on a
  climate output" back-edge, folded into the per-body climate solver the same way the
  TSS geology factors already are (`theory.md` § climate cluster), rather than as a
  separate pipeline stage wedged between climate and taint.
- If the spike shows the back-edge is cheaper handled by ordering (tidal lock reads a
  provisional pressure proxy, as albedo reads provisional hydro), prefer that. The
  decision is made **from a spike**, not specced blind — same lesson as C1.

**Acceptance.** `seed_500`'s `A II` (the one regression seed that hits the cascade today)
reproduces the same tidal-lock outcome; no doc or symbol contains "converge"/"fixed
point"; the cascade no longer consumes a global snapshot.

### C3 — All three IISS forms as fully-owned structs

> **Resolved (C3).** The Class IV-P body structs (`Class4PPartP`, `Class4PPartPB`, and
> their sub-blocks) and their `RenderBody` methods moved into `iiss/class4p.go`;
> `Class4PForm` now holds concrete `*Class4PPartP` / `*Class4PPartPB` — the closure and
> the `any` fields are gone. `worlds` builds the structs (`buildClass4PPlanet` /
> `buildClass4PBelt`), `iiss` renders them. Output byte-identical (Markdown + JSON). The
> rest of this section is the original spec rationale.

**Problem.** `iiss.Class4PForm` carries a `RenderBody func(*strings.Builder, FormHeader)`
closure and `any`-typed `PartP`/`PartPB` fields, because the Class IV-P body was too
domain-specific to move into `iiss/` without importing `worlds/`
(`theory.md` § "one asymmetry", issue #48). Class 0/I and Class II/III are clean structs;
Class IV-P leaks a worlds-side callback across the boundary. This also blocks Class IV-P
JSON parity.

**Proposal.** All three forms are fully-owned `iiss/` structs from day one, with every
field a concrete type. `worlds.BuildIISSForms` populates them; `iiss/` renderers (Markdown
and JSON) consume them with no callback and no `any`. The Class IV-P body content is
modeled as data on the struct, not as a closure. This is exactly the "make all three
forms structs from the start" prescription in the If-We-Started-Over doc § "Decisions
worth revising."

**Acceptance.** No `func` field and no `any` field on any `iiss/` form struct; JSON output
includes Class IV-P at parity with the other two forms; `iiss/` imports nothing from
`worlds/`.

### C4 — `stars/` is standalone

> **Resolved (C4).** MAO (the p.39 table, `MAO`, `LacksP39MAORow`, `IsPostStellar`)
> moved into `stars/mao.go`; the `GenerateSystemOpts.MAO` callback was removed and
> `GenerateSystem` computes the companion-of-giant orbit in-package. `stars/` has no
> inbound dependency on `worlds/`. Behavior byte-identical (regression + gold baselines
> unchanged). The rest of this section is the original spec rationale.

**Problem.** `worlds.MAO` is injected into `stars.GenerateSystemOpts.MAO`
(`generate.go:52`). This is the one place worlds-side knowledge flows into `stars/`
(`theory.md` § "MAO flows into stars"). It means `stars/` cannot be built or reasoned
about independently.

**Proposal.** MAO is a stars-domain concept (WBH p.39, maximum allowed orbit for an
interior body); it belongs in `stars/`. Move the MAO formula into `stars/` and delete the
Opts injection. If `worlds/` needs MAO it imports it from `stars/`, not the reverse. Audit
`GenerateSystemOpts` while here: `WithVariance`/`Accuracy` stay only if worked-example
tests still require them; otherwise cut per `design-intent.md` § "one path."

**Acceptance.** `stars/` has no inbound dependency on `worlds/`; `go test ./stars/...`
needs nothing from `worlds/`.

### C5 — File names describe dependency role, not chapter pagination

**Problem.** `worlds/stage2.go` … `worlds/stage10.go` survive from a chapter-numbered
mental model the design explicitly rejected. The numbers are dependency-graph indices, but
a reader landing in `worlds/` reads them as WBH chapters (`theory.md` § "stage*.go file
numbering").

**Proposal.** Name orchestrator files for what they do: `detail_frontend.go`,
`body_physical.go` (already exists per-feature), `rotation_tilt.go`, `climate.go`,
`taint_surface.go`, `geology.go`, `biology.go`, `habitability.go`, `aggregate.go`. The
per-feature files (`atmosphere.go`, `tidal_lock.go`, …) already follow this; the
orchestrators join them. The pipeline's canonical order lives in `generate.go`'s step
slice — that is the one place order is asserted — so filenames need not encode it.

**Acceptance.** No `stageN.go` filename in the rebuild.

### C6 — Documentation and tooling hygiene from day one

- **Vocabulary.** No "converge"/"fixed-point" language anywhere (see C2). The doc set is
  written against the honest model from the first commit, not retrofitted.
- **Bulk-sweep tool lives in the tree.** `main`'s 10 000-seed sweep tool
  (`cmd/*-bulk`) was a throwaway removed after each run and reconstructed from git history
  (`generator-error-catalog.md`, `theory.md` § "bulk-sweep tool's location"). The rebuild
  keeps a permanent `cmd/world-builder-bulk` (or a `task sweep` recipe) so "is the system
  clean today" is one command, not an archaeology exercise.
- **`BodyEmpty` semantics pinned.** `theory.md` § "Uncertainties" flags that what creates
  vs. filters `BodyEmpty` is not fully traced. The rebuild documents the sentinel's
  lifecycle explicitly and asserts (property test) that empty slots never reach a renderer.
- **Extract shared fixtures on day one.** `BuildZedFixture(t) (…)` in a non-`_test.go`
  helper from the first sub-project that needs Zed, per the If-We-Started-Over doc §
  "Test fixture extraction." No inline `composeZed()` that later needs external-test-package
  gymnastics.

## Test strategy

The four layers of `harness.md` / `theory.md` § "four-layer test discipline" carry
forward. C1 adds a fifth capability the earlier passes could not have:

- **Per-procedure dice-script tests** — unchanged; the proof of book fidelity.
- **Property tests** — per-body invariants over ≥1000 seeds; `MoonsHaveBodies` and
  `HZBodyHasClimate` remain the anti-pattern-A.1 sentinels.
- **Full-pipeline gold fixtures (new capability).** With per-body substreams (C1),
  encode Zed Prime end-to-end as a survivable gold fixture: the whole system, every
  field, to the digit, that does **not** break when an unrelated stage is reordered.
  This is the assertion pass 1 and pass 2 both had to abandon. Gate its introduction on
  the C1 spike proving substream isolation.
- **Markdown regression baseline** — re-generated for the rebuild's stream (not diffed
  against `main`).
- **Bulk sweep** — permanent tool (C6); 10 000 seeds, zero non-scope errors, as the
  readiness gate.

## Build order

Grounded in `lessons-learned.md` § L1 (signature shape predicts port cost):

1. **Foundations first, committed clean** (B.2/B.3 lessons): `dice/`, `roller/` **with
   `Fork`**, the C1 spike, `Body` + `Universe` + `AllBodies`, the naming protocol, the
   `task`/modernizer gate, `BuildZedFixture` helper, the doc skeleton (honest vocabulary).
   Nothing WBH-substantive ships until the C1 spike passes.
2. **Port the self-contained procedures verbatim.** Per L1, files that take raw inputs and
   return values port with zero adaptation: `sizing_*.go`, `period.go`, `body_physical.go`,
   `belt_details.go`, `moon_refinement.go`, `geology.go`, `biology.go`, `habitability.go`,
   `atmosphere*.go`, the `stars/` per-procedure files and their `Scripted` tests. These are
   the bulk of the code and the bulk of the fidelity; they are already correct on `main`.
3. **Rewrite the orchestrators** against the sub-roller tree (C1) and the honest climate
   model (C2). This is where the rebuild's design work concentrates — the orchestrators
   were fully rewritten pass-1→pass-2 too, for the same reason.
4. **Rebuild the `iiss/` forms as clean structs** (C3), stars standalone (C4), role-named
   files (C5).
5. **Restore the fidelity gate**: worked-example tests green, full-pipeline Zed gold
   fixture green, property tests green, bulk sweep clean, regression baseline generated.

Cadence: sub-project = dependency-graph stage, per pass-2's confirmed cycle granularity
(L4). Brainstorm → spec → plan → TDD with per-task two-stage review. `git status` clean at
every sub-project boundary (B.2).

## Non-goals and stop rules

- **No scope expansion.** WBH pp.147+ (social characteristics, special-object physics)
  stays out. The `ErrSpecialCircumstances` family remains a typed scope edge, not a
  generation path. Do not add code that anticipates pp.147+.
- **No new abstractions off the stop-list.** The sub-roller `Fork` is an additive method
  on the existing keystone interface, not a new framework — it is justified by C1's
  live need. Everything else on `design-intent.md` § "Stop rules" stays shelved.
- **No runtime toggles for book inconsistencies.** Six committed interpretations, encoded,
  asserted, cited. Unchanged.
- **The book stays the spec.** Fidelity is measured against the book's printed worked
  examples, not against `main`'s byte output.

## Open questions / risks

1. **C1 is unproven.** The sub-roller tree is the spine of every other benefit and has
   never been built here. If the spike shows keying is fragile (e.g. body identities are
   not stable early enough, or a procedure family's boundary is ambiguous), fall back to
   `main`'s single stream and drop the full-pipeline gold fixture ambition — but keep C2–C6,
   which stand on their own. **The spike decides.**
2. **C2's back-edge shape** (fold-into-solver vs. provisional-proxy ordering) is a
   spike outcome, not a spec decision. Do not commit the mechanism blind — that is the
   exact mistake L13 documents.
3. **Effort vs. value.** `main` is `v1.0`, faithful, and at a natural stopping point
   (`path-forward.md` § Recommendation). The rebuild is a craft investment, not a
   requirement. C1 is the only part that delivers something `main` cannot; C2–C6 are
   quality the current code could also reach by refactor. If the C1 spike disappoints,
   the honest recommendation may be "refactor `main` for C2–C6, skip the rebuild."

## Source documents

This spec synthesizes: `design-intent.md`, `theory.md`, `anti-patterns.md`,
`wbh-inconsistencies.md`, `history/lessons-learned.md` (§§ L1, L3, L4, L9, L13, L14),
`history/pass-1-retrospective/2026-05-06-pass-2.md`, `history/spike-findings.md`,
`history/generator-error-catalog.md`, and a read of `worlds/generate.go`,
`roller/roller.go`, and the stage orchestrators on `main` at the time of writing.
