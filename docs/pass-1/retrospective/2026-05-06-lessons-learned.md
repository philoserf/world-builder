# Lessons Learned — World Builder

**Date:** 2026-05-06
**Scope:** the entire project lifetime, from a one-task Python false-start through `cmd/wbh -format markdown` becoming the default.

## Context

The project encodes Mongoose Publishing's _World Builder's Handbook_ (Geir Lanesskog, 2023) — pp. 14–146 of physical star-system rules — as a faithful Go reference implementation, with `cmd/wbh` emitting all three IISS forms (Class 0/I, Class II/III, Class IV-P) as Markdown. Eight major sub-projects shipped: Stars, 2A, 2B, 2C, 3A1, 3A2a, 3A2b (split into temp + rederive), 3B (split into geology + biology + final), and the IISS Class IV-P PART P.B + Markdown output sub-projects on top. Each followed a brainstorm → spec → plan → TDD-with-subagent-review workflow. Specs and plans are dated under `docs/pass-1/specs/` and `docs/pass-1/plans/`.

## What worked

### Deterministic Roller as the load-bearing invariant

The `roller.Roller` interface (`Seeded` / `Scripted` / `Fixed`) was the project's most successful architectural decision. Every dice roll passes through it; no package-level RNG anywhere. A seed plus a sequence of options fully determines a system. `Scripted` panics on exhaustion — that always indicates a test bug, never a runtime issue, and the panic surfaces it immediately. The pattern survived ~30 sub-projects without modification.

### Worked-example regression tests as "proof of fidelity"

The book threads in-line examples through every chapter — most prominently the G7 V "Zed" quintuple. Encoding each example with `roller.NewScripted(...)` driven by the book's exact dice and asserting every output to the digit caught real bugs at every sub-project boundary. The implementation could not silently drift from the book without breaking a worked example. This was the single most valuable testing pattern.

### Brainstorm → spec → plan → subagent-driven implementation

The chapter-sized decomposition (3A1, 3A2a, 3A2b-temp, 3A2b-rederive, 3B-geology, 3B-biology, 3B-final, IISS Class IV-P PART P.B, Markdown output) gave each sub-project its own clean cycle: brainstorming surfaced scope ambiguity before code, the spec captured decisions and book-inconsistency notes, the plan decomposed into TDD-shaped tasks with full code in each step, and per-task two-stage review (spec compliance + code quality) caught real bugs before merge.

### Per-task two-stage subagent reviews

Spec-compliance review independently verified the implementer's claim that the work matched the spec; code-quality review then assessed the as-built implementation. Across this session alone the review pattern caught: a missing `BodyEmpty` guard, a silent-zero Span placeholder, a stale doc-comment, and a latent overwrite of resource ratings when both `Biology` and `Belt` were populated. Per memory, an earlier 2B run found six correctness bugs in the plan's reference code. Skipping reviews would have cost more in debugging than the reviews themselves cost in dispatch.

### Tables as Go literals with WBH page citations

Heterogeneous WBH tables live as typed Go literals, with `*float64` for nullable cells (the book's "—") and a doc-comment citing the page. The compiled-in approach kept the data next to its consumers, avoided a YAML/TOML build-time dependency, and made code review possible against the printed book. The single shared `interpolate(table, spectral_type, luminosity_class)` helper for coarse-grid physical quantities cleanly absorbed the book's own statement that interpolation is endorsed.

### Explicit placeholders over silent zeros

The "moon-path silent-zero" anti-pattern (memory entry of the same name) recurred four times across consecutive 3A2b/3B sub-projects: planet code added a step but didn't iterate moons, leaving moon outputs at zero. Each time the Opus final-gate review caught it. The fix pattern — render `(not generated)` strings or em-dash placeholders rather than `0.000` numerics — was adopted and propagated through Markdown rendering with explicit guards in every section.

### Modernizer-as-mandatory gate

Running `go fix ./...` first in the `task check` recipe and **failing if it produces any diff** turned `gopls`'s modernizer hints into a hard signal. The hints reflect Go 1.21+/1.22+ idioms that my training data predates (`min`/`max`, range-over-int, `new(value)`); without this gate I would have routinely missed them. Whenever a subagent hit the gate, the cure was always to inspect with `git diff` and stage — never override.

### Surfacing book inconsistencies, not hiding them

Six documented WBH inconsistencies (p.19 vs p.42 Class VI mass cells; p.58 vs p.63 Aab IV-d size; temperature-chapter contradictions; p.125 vs p.126 density DM; p.131 Compatibility worked example vs formula; p.132 gravity DM bands overlap). Each was captured as a memory entry, not a code workaround. Tests assert the chosen interpretation and cite the divergence. Future maintainers can adjudicate; the code is honest about which interpretation it picked and why.

## What didn't work / surfaced friction

### Python false start

The project began in Python (uv + ruff + pytest). On the very first task it hit a uv/macOS/Python 3.12 editable-install bug: `.pth` files inheriting the macOS hidden flag, which Python 3.12 then skips. Switching cost was negligible (one task of scaffolding) and Go's static-binary plus zero-packaging surface area paid for itself across heterogeneous-table-heavy work. **Lesson:** for a project whose data shape is "lots of small typed lookup tables with WBH page citations," Go is a better fit than Python. The dynamic-typing flexibility Python would have offered was never load-bearing.

### `gofumpt` CLI vs `golangci-lint`'s bundled `gofumpt`

The two disagreed on import grouping (the CLI was more recent). Picked the CLI as canonical and disabled golangci-lint's bundled copy. **Lesson:** when two tools claim authority over the same surface, declare canonical-source explicitly in tooling docs and config; don't run both.

### `just` → `task` migration mid-stream

The migration sat as uncommitted working-tree changes across most of this session. Subagents reported `just check && just test` clean while the `justfile` was deleted (likely because `task` was running on resilient state). The plan I wrote for the Markdown sub-project used `just check` and had to be sed-replaced before dispatch. **Lesson:** resolve toolchain migrations in their own commit before starting unrelated work; do not carry uncommitted toolchain changes across sub-projects.

### LSP diagnostic lag

Across multiple subagent runs, the editor's `(UndeclaredName)` diagnostics fired against newly-created functions for ~30 seconds after the file was committed. Tests passed, files existed on disk, but the LSP cache was stale. Falsely alarming and required manual verification each time. Not a project bug — a tooling artifact — but worth noting as a recurring distraction.

### Moon path silent-zero recurrence

The same anti-pattern hit four consecutive 3A2b/3B sub-projects. Per-task review caught all four; the spec template wasn't updated to call it out preemptively until late. **Lesson:** when an anti-pattern recurs, escalate it from "memory entry" to "checklist in the project CLAUDE.md / spec template" so future sub-projects guard against it before review.

### Code-reviewer subagent unreliability on long branches

The dedicated `code-reviewer` agent type cut off mid-investigation more than once on this branch's later tasks (Tasks 3 and 6 of the Markdown sub-project). The agent's investigation was sound; the structured report just never landed. Falling back to inline review by the controller worked in those cases. **Lesson:** for short, mechanical follow-up reviews, inline review by the controller may be a better fit than dispatching a fresh subagent.

## Recurring patterns

### Book inconsistencies as data, not bugs

Six independently-documented WBH inconsistencies. The book is the canonical source but is not internally consistent. Treating divergences as data (memory entries with explicit source citations) rather than as bugs to "fix" preserves traceability to the printed material. Future readers can disagree with our interpretation, but they cannot accuse the code of lying about which interpretation it chose.

### Moons mirror planets, mostly

Per WBH, moons run through nearly the same physical pipeline as their parent planet. The implementation pattern of `buildMoonPlacementView(m, parent)` synthesizing a `*DetailedPlacement` from a `Moon` works — it lets per-body procedures reuse one code path. But when a new step is added without iterating moons explicitly, the silent-zero anti-pattern fires.

### `Has*()` predicate accessors

`DetailedPlacement.HasAtmosphere()`, `HasGeology()`, `HasBiology()`, etc. wrap nil-pointer checks. Used consistently, they prevented panics across the per-body pipeline. Forgotten once or twice and caught by review. The pattern is mostly self-explanatory and survived without renaming.

## Workflow lessons

- **Brainstorm before coding, every time.** "This is too simple to need a design" is the project-wide false economy. Even one-line behavior changes benefit from 30 seconds of "what does the spec actually say."
- **Subagent reports are over-confident.** They reported `just check && just test` clean while `just` could not have run. Spot-check at least one verification command per task in the controller's session.
- **Per-task review > batch review.** Six bugs found per-task during 2B vs an unknown number that would have surfaced if the same review ran once at the end.
- **Modernizer drift is a gate, not a hint.** Treating `go fix ./...` as advisory misses real idioms. Treating it as a fail-the-build gate forced staying current.
- **Specs commit before code.** When the spec is committed, its decisions are reviewable independently of the code that implements them. Drift between spec and code is then a measurable artifact.
