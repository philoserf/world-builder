# Path Forward — World Builder

**Date:** 2026-05-06
**Scope:** concrete next-step candidates from the project's stated done state. Excludes WBH pp. 147–234 (World Social, Special Circumstances) per scope decision.

## Project state

Both halves of the stated "done" criteria are now in:

1. WBH pp. 14–146 physical star-system rules — encoded.
2. `cmd/wbh -format markdown` (default) emits the full system as Markdown — all three IISS forms under H1/H2 headings.

`-format json` and `-format short` remain available. Toolchain is `task` (Taskfile.yml). Local `task check && task test` is the gate; no CI. Public repo at `github.com/philoserf/world-builder`, MIT-licensed.

Below: candidate next steps, grouped by size. None are required; the project can stop here.

---

## Polish (small, low-risk, hours)

### CLI cosmetics

- **`(unnamed)` placeholder when no IISS designation is supplied.** `cmd/wbh` passes an empty `IISSClass23Header`, so the H1 title falls through to `(unnamed)`. Could derive a default from the seed (e.g., `# Star System: seed-42`) or from the short profile. Trivial change in `cmd/wbh/main.go`.
- **Zed golden's `System Age (Gyr) | 0.000`.** `composeZed()` doesn't set `sys.AgeGyr` (only individual stars). Fixing the composer would also fix the Markdown output, then regenerate the golden via `go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -update`.

### Deferred follow-up items from `MEMORY.md`

These are flagged in working memory; pulling them into focused mini-sub-projects would close the deferred list:

- **(p)** Habitability Notes referee-color string (post-3B-final review).
- **(q)** Referee mainworld override (`-mainworld <designation>` flag in `cmd/wbh`). Small CLI work.
- **(r)** Acceptance assertion 43 silently no-ops when mainworld is a moon.
- **(d)** `BecomesRetrograde` unit test (3A2a).
- **(c)** NaN-defensive guard in `AtMoment` (3A2b-temp).

### Code-quality follow-ups

- **`SurveyForm.ClassI bool`** is declared but never written or read anywhere. Either delete the field or populate it (e.g., based on whether the system has any companions). Pre-existing dead code, surfaced during Markdown sub-project review.
- **`renderIISSClass4P` plain-text renderer is now unused production code** — only called by tests. Either delete it (and its tests) or refactor to produce a struct (see "Mid-size enhancements" below).

## Mid-size enhancements (days)

### Refactor `RenderIISSClass4P` to produce a struct

The Class 0/I and Class II/III forms are both structs (`SurveyForm`, `IISSClass23Form`); Class IV-P is plain text. Refactoring the IV-P renderer to produce an `IISSClass4PForm` struct would:

- give JSON parity (the JSON path could include Class IV-P).
- remove the now-unused plain-text renderer (replace its callers — only tests today).
- give the Markdown formatter an option to consume a struct rather than reading from `*DetailedPlacement` directly (cleaner uniformity if desired).

Trade-off: ~200 lines of new struct + builder code, plus updating tests. Worth doing only if a webservice or similar consumer materializes.

### Verbose Markdown mode

The book renders Class IV-P only for the mainworld. A `-verbose` flag could render Class IV-P-style detail for every notable terrestrial/moon/belt. Output length grows substantially (10×–30× for a typical Zed-sized system), so this is gated behind opt-in. Useful for full-system sheets a Referee wants printed.

### World Maps (deferred item (o))

WBH pp. 135–137 specifies hex-grid world maps via icosahedron tessellation. Two methods (5-hex and 7-hex variants). Visual hex rendering in Markdown is awkward; ASCII-art tessellation in monospace works. Useful for terminal output; less useful for GitHub-rendered Markdown.

### Per-body Major Bodies subtable in Class IV-P PART P.B

The belt-mainworld form (Form 0407K-IV PART P.B) has a Major Bodies subtable expecting per-body rows (SAH/UWP, Diameter, Density, Mass per body). The current implementation renders only counts (`Counts only: N size-1 + M size-S; per-body detail not generated.`). Fixing this requires generating individual belt members — a chapter-of-its-own task that the WBH treats as "Referee fills in later when developing specific bodies of interest."

### Deferred follow-up items from `MEMORY.md`

- **(a)** `MeanBySeason` latitude composition (3A2b-temp).
- **(b)** Twilight-zone hemisphere selection in scenario methods (3A2b-temp).
- **(g)** Atmosphere A/B/C/F+ "consider boiling" runaway case (deferred from 3A2b-rederive MVP).
- **(h)** Tidal-lock re-eval if pressure crosses 2.5 bar (deferred indefinitely).
- **(i)** Terrestrial-on-terrestrial tidal effects has the `MassEarth=0` silent-zero bug.
- **(j)** `ScaleHeight` not recomputed when 5E mutates `MeanK`.
- **(k)** Biologic-taint atmosphere special case for Biomass.
- **(l)** Optional rules deferred (any oxygen atmosphere = biomass≥1; Rare Earth Universe Variant).
- **(m)** Low-oxygen-taint DM-2 in Biocomplexity.

## Larger directions (weeks)

### JSON-over-HTTP webservice

The project's design has anticipated this from the start ("a future webservice may expose JSON over HTTP"). Smallest viable shape: a `cmd/wbh-server` binary serving `/system?seed=N` returning the full SystemDetail JSON, plus per-form endpoints (`/system/N/class0i`, `/system/N/class23`, `/system/N/class4p`). Class IV-P would benefit from the struct refactor above. Beyond that the work is mostly HTTP wiring.

### Full repo audit

Running `task check && task test` is the gate but has no coverage instrumentation. A one-time `go test -cover ./...` audit would identify uncovered branches. Memory entry _per-task subagent reviews catch real spec bugs_ suggests the existing tests are thorough; coverage data would either confirm that or surface gaps.

### Public-facing documentation

The repo is public on GitHub but `README.md` is minimal. A Wikipedia-style docs site (Hugo, MkDocs, or just enriched Markdown in `docs/`) describing the architecture, the WBH-traceability conventions, and the known book inconsistencies would help anyone landing on the repo cold.

## Out of scope (per scope memo)

WBH pp. 147–234 (World Social Characteristics, Special Circumstances — UPP, starport, trade codes, polities, etc.) are explicitly out of scope for current and near-term purposes. See `docs/pass-1/specs/2026-05-06-cmd-wbh-markdown-output-design.md` non-goals and the `project_world_builder_scope_narrowed` memory entry. Do not start work in those chapters; do not add code that anticipates them.

## Recommendation

**The project is at a natural stopping point.** Both halves of the stated done criteria are met. Further work is optional polish, scope expansion, or aspirational. The deferred-item list (a–r) is not load-bearing — items are real but the project functions correctly without them. If picking up again, the polish cluster (CLI cosmetics + (q) override flag + dead-code removal) is the lowest-friction next sub-project; everything else is a deliberate choice to expand scope.
