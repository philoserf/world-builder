# cmd/wbh Markdown output — Design

**Date:** 2026-05-06
**Status:** approved by user; ready for plan
**Source material:** WBH pp. 33–34 (Form 0421B-0I), pp. 138–139 (Forms 0407F-IV PART P and 0407K-IV PART P.B), pp. 140, 143, 145 (printed examples of Class II/III forms 0421D-II.III).

## Purpose

Make `cmd/wbh` emit a complete, human-readable description of a generated star system as Markdown. Today the CLI emits the Class 0/I JSON survey form only; the existing Class II/III and Class IV-P renderers (the latter merged in `feat/iiss-class4p-belt`) are not yet reachable from the CLI. This sub-project closes the rules-half-to-output-half gap and makes Markdown the default `-format`.

This satisfies the second half of the project's "done" criteria from `CLAUDE.md`: the CLI emits "a complete description of the generated system as Markdown — all three IISS forms (Class 0/I, Class II/III, Class IV-P) covering whatever world type the mainworld turns out to be."

## Non-goals

- **Refactoring `RenderIISSClass4P` to produce a struct.** It currently returns plain text and is called only from tests. Refactor when JSON parity becomes a real need (a future webservice); not now.
- **Including Class IV-P data in JSON output.** Class 0/I JSON stays as-is; future scope.
- **Verbose mode** (a Class IV-P-style form for every notable body, not just the mainworld). Different scope; would massively expand output length.
- **A `-mainworld <designation>` override flag.** Deferred memory item (q); not blocking.
- **Repointing existing tests.** The existing per-form tests (Class 0/I JSON, IV-P plain text) keep passing unchanged; this sub-project adds new tests, doesn't replace.

## Architecture

### Markdown formatters consume existing structs (where they exist)

For Class 0/I and Class II/III, the existing `stars.SurveyForm` and `worlds.IISSClass23Form` structs are already the right shape for any rendering target — they're plain Go structs with native types, not transport-encoded. `BuildSurveyForm` and `RenderIISSClass23` already do the heavy work of flattening `stars.System` and `worlds.SystemDetail` into per-row tables (with all the barycentre-composite, HZCO-source-row, MAO-post-fill, Objects-table-assembly logic that's expensive to re-derive). Markdown formatters consume these structs directly — no duplication, no new intermediate type.

For Class IV-P, no struct exists today; `RenderIISSClass4P` returns plain text. The Markdown formatter reads from `*DetailedPlacement` and `stars.System` directly (same inputs as the existing plain-text renderer), since there's no per-body flattening logic to share.

Existing renderers stay untouched. The plain-text `RenderIISSClass4P` becomes effectively dead production code (still tested) once Markdown is the CLI default — accept that, revisit when JSON parity matters.

The Markdown formatters do not parse the existing plain-text `RenderIISSClass4P` output — they read source data fresh.

### File layout

| File                        | Responsibility                                                                                                                                                                                                                                                                                                                                                                                                               |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `stars/markdown.go`         | `RenderClass0IMarkdown(SurveyForm) string` — Form 0421B-0I (consumes existing struct).                                                                                                                                                                                                                                                                                                                                       |
| `worlds/markdown.go`        | `RenderClass23Markdown(IISSClass23Form) string` — Form 0421D-II.III (consumes existing struct). Plus `RenderClass4PMarkdown(*DetailedPlacement, stars.System, mainworldDesignation) string` — Form 0407F-IV PART P or PART P.B, dispatched on `SizeCode == "0"` (reads source data; no struct exists for Class IV-P).                                                                                                        |
| `worlds/markdown_system.go` | Top-level orchestrator: `RenderSystemMarkdown(SystemDetail, stars.System) string`. Reads the Class 0/I `SurveyForm` from `sd.Survey.SurveyForm` (embedded) and the Class II/III form from `sd.Survey` (both populated upstream by `DetailSystem`). Calls the three Markdown formatters in book order. Emits H1 system title + Class 0/I H2 + Class II/III H2 + Class IV-P H2 (skipped when `sd.MainworldDesignation == ""`). |
| `cmd/wbh/main.go`           | New `case "markdown":` calling `worlds.RenderSystemMarkdown(...)`. Becomes the new default `-format`. JSON and short formats remain available.                                                                                                                                                                                                                                                                               |

### Hybrid Markdown layout

- **Real tables** for actually-tabular sections: Stars (one row per stellar component), Objects (one row per body), Subordinates (one row per moon), Major Bodies subtable (placeholder for now — counts only per the merged PART P.B).
- **2-column `Field | Value` tables** for grouped per-field sections: Orbit, Size, Atmosphere, Hydrographics, Rotation, Temperature, Seismic, Life, Resources, Habitability. One block per IV-P section.
- **H3 section headers** within each form's H2 (e.g., `### Orbit`, `### Atmosphere`, `### Comments`). Consistent across all forms.

### Document shape

```markdown
# Star System: <IISS Designation>

## IISS Class 0/I Survey — Form 0421B-0I

[per-field section as 2-column table — Sector, Initial Survey, etc.]

[Stars and Companions table — multi-row]

## IISS Class II/III Survey — Form 0421D-II.III

[per-field section as 2-column table — same metadata]

[Stars table — same data as Class 0/I, repeated for self-contained form]

[Objects table — multi-row, every body in the system]

## IISS Class IV-P Survey — Form 0407F-IV PART P

(or `Form 0407K-IV PART P.B` when mainworld is a belt)

[per-field 2-column tables: Orbit, Size, Atmosphere, Hydrographics,
Rotation, Temperature, Seismic, Life, Resources, Habitability]

[Subordinates table — moons, multi-row, when present]

### Comments

[optional mainworld marker line]
```

When `MainworldDesignation == ""` the Class IV-P H2 section is skipped entirely (book renders the form only for a mainworld; no mainworld means no form).

### Field handling

- Empty / unset string fields and zero-value numeric fields render as `—` (em-dash, the book's "—" convention for "not applicable").
- Nil pointer fields on `DetailedPlacement` (`*Habitability`, `*Geology`, etc. for non-applicable body types) render the H3 section header followed by a single-row `| Status | (not generated) |` cell. The section is never silently dropped; the absence of data is explicit.
- Numeric formatting: `%.2f` for AU and orbits, `%.3f` for Span, `%.0f` for diameter in km, `%d` for atmosphere/hydro codes and percentages. Mirrors existing PART P / PART P.B plain-text choices.
- The Stars table repetition between Class 0/I and Class II/III is intentional — each form is self-contained and can be cut/copied independently, faithful to the book's printed forms.

### CLI changes

`cmd/wbh/main.go`:

- Default `-format` flips from `json` to `markdown`.
- New `case "markdown":` that calls `worlds.RenderSystemMarkdown(sd, sys)` and writes to stdout.
- The Markdown path needs the full `SystemDetail` (not just `stars.System`), so the CLI now calls `worlds.SystemPlacement` and `worlds.DetailSystem` when format is `markdown`. The JSON path stays Class 0/I only. The `short` path keeps using `stars.ShortProfile(sys)` as today — no new pipeline calls.
- Existing `-format json` and `-format short` paths continue to work unchanged.

## Testing strategy

Three layers, scaled by what each layer actually verifies:

### 1. Per-form unit tests on synthetic data

One test file per Markdown formatter:

- `stars/markdown_test.go` — synthesizes `stars.System` with a primary + companion, asserts H2 heading, Stars-table columns, key field values via `strings.Contains`.
- `worlds/markdown_test.go` — separate tests for `RenderClass23Markdown` (synth `SystemDetail`) and `RenderClass4PMarkdown` (synth `DetailedPlacement` for both terrestrial and belt variants). Each test asserts H2 heading + the right form name + key field labels.
- `worlds/markdown_system_test.go` — orchestrator test asserting the H1 title is present, all three H2 sections appear in book order when mainworld exists, Class IV-P section is absent when mainworld is empty.

Pattern matches existing `iiss_class4p_belt_test.go` style — fast, focused, `strings.Contains` for structural assertions.

### 2. Golden-file snapshot on Zed

`worlds/testdata/zed_markdown.golden` — the Zed system's full Markdown output captured from the existing `roller.NewScripted(...)` worked-example pipeline. Test in `worlds/markdown_system_test.go` reads the golden file and compares against fresh render.

When format choices change deliberately, regenerate the golden:

```bash
go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -update
```

This buys regression coverage for "did anything in the rendering pipeline change unexpectedly" without coupling to specific field formatting like `%.3f` vs `%.2f`.

The Zed acceptance tests (`TestZed_FullDetail_*`) already prove data fidelity at the `SystemDetail` level — values are correct. The Markdown golden proves rendering is stable on top of that data.

### 3. CLI tests

`cmd/wbh/main_test.go`:

- `TestCLI_MarkdownIsDefault` — invoke with no `-format` flag, assert stdout starts with `"# Star System:"`.
- `TestCLI_MarkdownExplicit` — `-format markdown`, same assertion.
- `TestCLI_JSONStillWorks` — `-format json`, assert valid JSON parses.
- `TestCLI_ShortStillWorks` — `-format short`, assert single-line output.

## Implementation order (informational; refined in plan)

1. Class 0/I Markdown (`stars/markdown.go`) — smallest, no deps on `worlds/`.
2. Class IV-P Markdown (`worlds/markdown.go`) — both PART P and PART P.B variants.
3. Class II/III Markdown (`worlds/markdown.go`) — renders the same Stars table as Class 0/I via a separate code path; no shared helpers extracted (the Stars data is read fresh from `stars.System` each time).
4. Top-level orchestrator (`worlds/markdown_system.go`).
5. CLI wiring (`cmd/wbh/main.go`) — flip default, add `markdown` case.
6. Golden-file Zed test.

Each step lands its own tests + commit. The plan will decompose into TDD-shaped tasks.

## Success criteria

- `go run ./cmd/wbh -seed 42` (no `-format`) emits Markdown to stdout starting with `# Star System: ...` and containing three H2 form sections (or two when no mainworld).
- The Zed seed pipeline produces a stable, golden-matched Markdown rendering across all three IISS forms.
- `cmd/wbh -format json` and `-format short` continue to produce their existing output.
- `just check && just test` clean.
- A reader of the printed output can match every field to a labeled cell on the corresponding WBH form.

## Carry-forwards

These remain deferred and do not block this sub-project:

- Refactoring `RenderIISSClass4P` to produce a struct (would let JSON include Class IV-P).
- Per-body Class IV-P forms / verbose mode.
- Referee mainworld override flag (deferred item (q)).
- Per-body detail in the Class IV-P PART P.B Major Bodies subtable.
- World maps (deferred item (o)).
