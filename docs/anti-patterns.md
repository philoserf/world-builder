# Anti-Patterns Catalog

A living catalog of failure modes the project guards against. Each entry is a concrete bug pattern that surfaced in development, with how to detect it and how the current code prevents it. Divided into code-design, workflow, and review-process.

Read before introducing a structural change. New anti-patterns discovered post-v1 are appended here.

## A. Code-design anti-patterns

### A.1 Moon-path silent-zero (the recurring one)

**What.** A new procedure runs "for each planet, do X" but does not iterate moons. Outputs for moons remain at zero values (numeric) or nil pointers (rendered as "(not generated)").

**How it surfaced.** Recurred four times across consecutive sub-projects (3A2b-temp, 3A2b-rederive, 3B-geology, 3B-biology). Each time the Opus final-gate review caught it. Per-task review missed it because the chapter framing makes "for each planet" the natural reading.

**How to detect.** Any procedure whose pass-1 ancestor or whose WBH text says "for each planet" must explicitly answer: does this also run for moons? If yes, the implementation iterates moons, not just planets. If no, the spec calls out the exclusion explicitly.

**How prevented.** Pass 2 uses a single `applyBodyProcedures(body)` over a unified iterator that yields planets and moons uniformly. The moon-vs-planet distinction is a parameter to the procedure, not a separate code path. Any procedure that needs different behavior for moons takes a `BodyKind` or equivalent parameter; it does not get a separate moon-iteration block.

**Spec checklist.** Does this procedure run per body? If yes, is it expressed via the unified iterator? If no, why not, and what's the exclusion criterion?

### A.2 Plain-text-vs-struct renderer asymmetry

**What.** Sibling renderers return different shapes (one returns a struct, another returns a string). Downstream consumers (Markdown, JSON, golden-file tests) cannot use them uniformly.

**How it surfaced.** `RenderClass4P` returned a string while Class 0/I and Class II/III returned structs. The Markdown sub-project had to redesign mid-plan-writing to add a Q1 paragraph reconciling the asymmetry.

**How prevented.** All three IISS forms return typed structs from day one. `RenderClass0I → SurveyForm`, `RenderClass23 → IISSClass23Form`, `RenderClass4P → IISSClass4PForm`. Markdown/JSON/PlainText rendering are separate consumers of the same struct. No string-only renderer in production code.

**Spec checklist.** Does every renderer in this sub-project return a typed struct? If not, why?

### A.3 Embedding chain depth

**What.** A struct embeds another that embeds another that embeds another. Reading or setting fields requires the reader to trace the chain. Plays badly with code review and IDE navigation.

**How it surfaced.** `Slot → AnomalousSlot → Placement → DetailedPlacement` is four levels. Setting `dp.Body` traverses through to `Placement.Slot.Body`, which is correct Go but confusing for readers.

**How prevented.** Maximum two levels of embedding. Pass 2 collapses the four-level chain into a single `Body` struct that holds the placement fields directly (`Group`, `Orbit`, `Eccentricity`, `Kind`) — embedding eliminated, named composition used where structure is shared.

**Spec checklist.** Does this sub-project introduce a struct that embeds another? Is the resulting depth ≤ 2? If not, justify or refactor.

### A.4 Dead fields

**What.** A field is declared on a public struct, never written, never read, and persists across releases as latent obligation.

**How it surfaced.** `SurveyForm.ClassI bool` was declared but never written or read anywhere. Surfaced during a code-quality review long after introduction.

**How prevented.** A field is added only when its first writer and first reader are committed in the same change. Stub fields for "we'll wire this later" are forbidden. If the API surface needs a placeholder, it goes as a panic-on-call function, not a silent field.

**Spec checklist.** Does this sub-project introduce any new field? If yes, are both writer and reader in this sub-project? If no, why is the field needed?

### A.5 Structured strings (sed-bound delimiter parsing)

**What.** A field carries multiple semantic values joined by a delimiter, with downstream code splitting on the delimiter to extract them. Every consumer must know about the split convention.

**How it surfaced.** `IISSClass23Header.SectorLocation string` carried sector and location joined by `" | "`. Consumers used `splitSectorLocation` to extract them. Should have been two fields.

**How prevented.** When a field carries N semantic values, model N fields. Splitting strings is reserved for parsing user input at trust boundaries.

**Spec checklist.** Does this sub-project introduce any field that bundles multiple values into one string? If yes, why isn't it N fields?

### A.6 Misleading parameter names

**What.** A parameter is named for one concept but consumed as another. Tests that happen to pass the right value for the wrong reason perpetuate the bug.

**How it surfaced.** `RollGasMix(r, atmosphereSubtype, ...)` named its first string parameter `atmosphereSubtype` but actually expected a column letter ("A"/"B"/"C") derived from the atmosphere code. Eleven task-level reviews missed it. The fix was a single rename, but the bug shipped because nobody owned the API as a whole.

**How prevented.** Every parameter name must match what the implementation reads from it. Contract tests for misuse paths are mandatory: "if you pass a Subtype literal to a parameter that wants a column letter, you get a compile error or an explicit misuse error." If the type system can encode the distinction (e.g., `AtmosphereColumnLetter` as a typed string), it does. Pass-2's full stub interface (committed before any implementation) is reviewed for naming consistency before any procedure ships.

**Spec checklist.** For every new public function: does each parameter's name match what the function reads from it? Are there contract tests that exercise plausible misuse?

### A.7 Single-file growth without seams

**What.** A file accumulates multiple sibling concerns and grows past a reviewable size. New additions land in the same file because that's where similar code lives.

**How it surfaced.** `worlds/markdown.go` ended at 419 lines holding both Class IV-P (PART P + PART P.B) and Class II/III renderers. Code-quality review correctly flagged this as a smell.

**How prevented.** Per-form renderers ship in per-form files: `markdown_class0i.go`, `markdown_class23.go`, `markdown_class4p.go`, `markdown_system.go`. Helper functions live in the same package and are shared without forcing co-location. New consumers of the same form land in the same file; new forms get new files.

**Spec checklist.** Does this sub-project add code to an existing file > 300 lines? If yes, is the addition a clear extension of the file's single concern, or should it spawn a sibling file?

### A.8 Stale pre-fixed-point inputs

**What.** A procedure consumes a value that is later refined (climate fixed-point recovery) but does not re-run after the value stabilizes. The output is silently wrong for cases where the refinement crosses a band boundary.

**How it surfaced.** Pass 1's Stage 5B ran `GenerateSurfaceDistribution` against the preliminary hydrographics from Stage 5A; it never re-ran after Stage 5D's rederive recovery. For HZ worlds the preliminary and converged hydro happened to be identical, so the bug never fired in fixtures, but the dependency was unsound.

**How prevented.** The dependency graph (`dependency-graph.md`) is the source of truth for ordering. Any procedure that consumes climate-cluster output runs after `ConvergeClimate` returns. Surface distribution and tectonic plates move to post-climate stages.

**Spec checklist.** Does this procedure read `Atmosphere`, `Hydrographics`, `Temperature`, or any TSS-dependent value? If yes, is it scheduled after `ConvergeClimate`?

### A.9 Implicit transaction across multiple writers

**What.** Several procedures write to overlapping fields on the same struct in sequence, with later writers depending on earlier writers' completion. If any writer is added or reordered, the sequence breaks silently.

**How it surfaced.** Mostly avoided by pass 1's strict per-step ordering, but the latent risk exists in `Atmosphere` (Code is set in 5A, mutated in 5D's runaway-greenhouse, mutated in 5D-prime's promotion). The mutation-in-place pattern depends on every step running in the right order.

**How prevented.** Where possible, replace mutation-in-place with a returned-replacement value (`Atmosphere → Atmosphere`). Where mutation is unavoidable (`ApplyInherentTempAddition`), the mutator is named for what it does, the doc-comment cites the WBH page that requires the mutation, and the unit test asserts both the mutation and the post-condition.

**Spec checklist.** Does this procedure mutate state owned by a shared struct? If yes, is the mutation point named and documented? Is there a test that asserts the post-condition?

## B. Workflow anti-patterns

### B.1 Modernizer-clean-tree gate trips on uncommitted Go diff

**What.** `task check` runs `go fix ./...` first and fails if any uncommitted Go diff exists. Subagents and humans hit this and misread the error message ("ERROR: go fix ./... produced changes") as "the modernizer wants something."

**How it surfaced.** Surprised every subagent the first time. Cure was always "stage your changes first, then re-run."

**How prevented.** Either rewrite the modernize recipe to compare against the current diff (not HEAD) so it does not trip on unrelated uncommitted Go work, or document the workflow expectation prominently in the project CLAUDE.md. Pass-2 default: rewrite the recipe; keep the gate, fix the message.

**Spec checklist.** Does this sub-project touch the modernize recipe in `Taskfile.yml`? If touching, has the message been clarified?

### B.2 Toolchain migration carrying uncommitted state

**What.** A toolchain change (e.g., `just` → `task`) sits as uncommitted working-tree changes across multiple sub-projects. Subagent reports become unreliable because the underlying tools partially work.

**How it surfaced.** The `just` → `task` migration sat as uncommitted state through most of the IISS Markdown sub-project. Subagents reported `just check && just test` clean while the `justfile` was deleted.

**How prevented.** At the start of any sub-project, `git status` must be clean. Toolchain changes get their own committed sub-project before any unrelated work proceeds.

**Spec checklist.** Pre-flight: is `git status` clean? If not, what is the uncommitted change, and should it be committed first as its own sub-project?

### B.3 Test fixture inline in test files

**What.** A complex test fixture (e.g., `composeZed()` building the full Zed Prime system) lives inline in `_test.go` files. When a sibling test or a golden-file generator needs the same fixture, the choices are: copy it, refactor it invasively, or use a `package_test` external-package coupling trick.

**How it surfaced.** Zed `SystemDetail` builder was inline in `worked_examples_test.go`. The Markdown sub-project's golden-file test had to reuse it via `worlds_test` external-package access — workable but late and surprising.

**How prevented.** Extract a `BuildZedFixture(t *testing.T) (SystemDetail, stars.System)` helper in a non-`_test.go` file from the first sub-project that needs Zed. Subsequent fixtures live alongside it. Any test that needs a fully-built Zed reuses one path.

**Spec checklist.** Does this sub-project introduce a complex fixture? If yes, does it live in a non-`_test.go` helper, or is there a justification for `_test.go` placement?

## C. Review-process anti-patterns

### C.1 "Self-review found no issues" over-confidence

**What.** A subagent reports "self-review found no issues" and the controller accepts the report without verification. The review may have been thorough or may have been performative; the report doesn't distinguish.

**How it surfaced.** Strategic reflection memory called this out explicitly: "Subagent reports framed as 'self-review found no issues' should get more skepticism than they currently do."

**How prevented.** Subagent reports of "no issues" trigger a controller-side spot check: pick one assertion the report makes and verify it independently. If the spot check is clean, the report is accepted; if it's not, the report is escalated.

**Spec checklist.** N/A (process discipline).

### C.2 False-positive verification reports

**What.** A subagent reports `task check && task test` clean while the underlying tools were partially broken (e.g., toolchain mid-migration). The report is technically accurate ("the command exited 0") but operationally wrong.

**How it surfaced.** Subagents reported `just check && just test` clean while the `justfile` was deleted; `task` was running on resilient state.

**How prevented.** Spot-check at least one verification command per task in the controller's session. If the controller cannot run the verification, the report is provisional, not authoritative.

**Spec checklist.** N/A (process discipline).

### C.3 Modernizer-hint dismissal as "noise"

**What.** Tooling output (modernizer hints, LSP diagnostics, golangci-lint findings) is dismissed wholesale as "noise" without checking whether each item is signal.

**How it surfaced.** Modernizer hints (`min`/`max`, range-over-int, `new(value)`) are real Go 1.21+/1.22+ idioms but were dismissed for a session as noise until corrected.

**How prevented.** Tooling output is treated as signal until proven otherwise. Each diagnostic gets read; the response is "apply" or "explain why this is wrong." The modernize gate makes this mandatory at the project level; the discipline extends to subagent reports as well.

**Spec checklist.** N/A (process discipline; reinforced by the modernize gate).

### C.4 Code-reviewer subagent unreliability on long branches

**What.** The dedicated `code-reviewer` agent type cuts off mid-investigation on later tasks of long branches. The agent's investigation is sound; the structured report just never lands.

**How it surfaced.** Hit twice on the Markdown sub-project's later tasks.

**How prevented.** For short, mechanical follow-up reviews, inline review by the controller may be a better fit than dispatching a fresh subagent. The code picks the review mechanism per task: deep review → subagent; mechanical follow-up → controller inline. The choice is documented in the spec.

**Spec checklist.** For each task that involves a review, has the review mechanism been chosen deliberately?

## How to use this document

- Before introducing a structural change, scan the relevant section and verify the change does not regress any entry.
- For code review of substantial new work, cite the entries the work touches and confirm the response.
- New anti-patterns discovered post-v1 are appended with the same shape: name, what, how-it-surfaced, how-prevented, spec-checklist.

This document evolves. The discipline does not.
