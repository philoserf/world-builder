# wbh — World Builder's Handbook reference implementation

Go library and CLI that generates Mongoose Traveller star systems per the procedures in the _World Builder's Handbook_ (Geir Lanesskog, 2023). Given a seed, the tool produces a complete deterministic system — stars, planets, moons, belts, atmospheres, oceans, climate, geology, biology, habitability — rendered as the three IISS Survey forms (Class 0/I + Class II/III + Class IV-P) in Markdown, JSON, or a short profile string.

## What it covers

WBH chapters 1-3 — pp.14-146:

- **Stars** (pp.14-35). Spectral types, luminosity classes, mass / diameter / temperature / luminosity, age, multi-star systems with companions, special objects (brown dwarfs, white dwarfs, etc. as primaries are stubbed; as companions they generate with type / mass / age).
- **System Worlds and Orbits** (pp.36-68). Habitable-zone central orbits, available orbits per stellar group, world counts, baseline orbit, system spread, anomalous slots, world placement, eccentricities.
- **World Physical** (pp.69-146). Per-body sizing, moons, designations, periods, HZ tagging, body physical (composition / density / gravity / mass), belt details, rotation / tilt / tidal lock / tidal effects, atmosphere / hydrographics / temperature (climate cluster), atmosphere taint typology, surface distribution, geology (residual seismic + tidal stress + tidal heating + tectonic plates), biology (biomass through compatibility), habitability rating, mainworld pick.
- **IISS Survey Forms.** Class 0/I (stellar census, p.35), Class II/III (per-body Objects table, pp.60-67), Class IV-P (mainworld detail, pp.138-146) with both PART P (planet/moon variants) and PART P.B (belt variant).

What's **not** here (out of scope): pp.147-234 (World Social Characteristics, Special Circumstances) and detailed special-object physics. The `Special Circumstances` chapter is explicitly deferred per project intent.

## Quick start

```bash
# Build
go build ./cmd/wbh

# Generate a full system as Markdown
go run ./cmd/wbh -seed 42 -format markdown

# Short profile only
go run ./cmd/wbh -seed 42 -format short
# → 2-1-9-8-0.7
# (G-P-T-N-S form per WBH p.58: gas-giants, belts, terrestrials, baseline, spread)

# Full system as JSON for downstream tooling
go run ./cmd/wbh -seed 42 -format json
```

Sample of `-format markdown` output (truncated):

```text
# A VII — System Survey

**Mainworld:** A VII

Short profile: `2-1-9-8-0.7`

Long profile: `A-7-T-G-T-T-G-T-T-T-0.7:B-0-T-T-P-T-0.7`

## Class 0/I — A VII

| Component | Class | Mass | Diameter | Temp (K) | Luminosity | HZCO |
| --------- | ----- | ---- | -------- | -------- | ---------- | ---- |
| A | A3 V | 1.780 | 1.914 | 8800 | 19.78 | 5.69 |
| B | A9 V | 1.380 | 1.830 | 7600 | 10.07 | 5.16 |

## Class II/III — A VII
...
```

Class IV-P renders only for the auto-picked mainworld (per WBH p.134's mainworld-priority chain: native sophonts → highest habitability → highest resource → first in iteration order), using whichever PART variant matches the mainworld's type (planet, moon, or belt).

## Library usage

```go
import (
    "fmt"

    "wbh/iiss"
    "wbh/worlds"
)

func main() {
    u, err := worlds.Generate(42)
    if err != nil {
        panic(err)
    }
    fmt.Print(iiss.MarkdownSystem(u.Detail.SystemForms))
}
```

Three layers:

- `dice/` and `roller/` — dice-notation parser and `Roller` interface with `Seeded`, `Scripted`, `Fixed` impls. A seed plus a procedure sequence fully determines a system; no package-level RNG anywhere.
- `stars/` — WBH pp.14-35. `stars.GenerateSystem(r, opts)` returns a multi-star `System`.
- `worlds/` — WBH pp.36-146. `worlds.Generate(seed)` returns a complete `Universe` (system + placement + per-body detail + system aggregations).
- `iiss/` — IISS form structs and renderers (Markdown / JSON / PlainText). Boundary type is `iiss.SystemForms`, populated by `worlds.BuildIISSForms` and consumed by the renderers.

Lower-level entry points (`stars.GenerateSystem`, `worlds.GenerateSystemPlacement`, individual `Apply*` stages, per-procedure `Roll*` and `Generate*` functions) remain available for finer control — see `docs/pass-2/api-surface.md`.

## Architecture

The project went through a major re-layering in 2026. Pass-1 followed WBH pagination as architecture; pass-2 inverted that — **the data dependency graph determines structure, worked-example fixtures determine the acceptance gate, the book is a citation system rather than an architecture.**

Key pass-2 design choices, all documented in `docs/pass-2/`:

- **Unified `Body` type.** Moons are `Body{Kind: BodyMoon, Parent: <planet>}` walked by the same iterator as planets. The moon-path silent-zero anti-pattern that recurred four times in pass-1 is prevented at the type level.
- **`iiss/` package boundary.** Renderers and form structs live in their own package; `iiss/` does not import `worlds/`. `iiss.SystemForms` is the boundary aggregate.
- **`ConvergeClimate` per-body solver.** Folds partial geology (Residual + TSF + THF) into each rederive pass. Originally framed as a fixed-point solver; empirical investigation revealed the climate cluster is not a fixed point in the strict sense (`RederiveAtmosphereHydrographics` samples fresh dice per call), so pass-2 settled on pass-1's pragmatic 2-pass-take-the-last behaviour, honestly named.
- **Stage orchestrators**: `ApplyDetailFrontEnd`, `ApplyBodyPhysical`, `ApplyRotationTilt`, `ApplyClimate`, `ApplyTaintTypology`, `ApplyGeology`, `ApplyBiology`, `ApplyHabitability`, `AggregateSystem` — each walks `Body` + `Body.Children` to drive per-procedure work.

The pass-1 implementation is preserved on tag `pass-1-final` for archival.

## Status

Pass 2 is on main. All gates green:

```bash
task        # check + test (modernizer + gofumpt + vet + golangci-lint + go test -race)
```

Coverage:

- Per-procedure value-exact tests for every WBH-narrated dice script (the book is the spec).
- Stage integration tests using `Seeded` shape-invariant assertions (per `docs/pass-2/spike-findings.md` § Finding 2: full-pipeline gold scripts don't survive pipeline reorders).
- Façade end-to-end tests over 100 seeds.
- Property tests (8 invariants × 1000 seeds each).
- Misuse-path contract tests for every public function.
- Markdown regression baseline (5 seeds × full output) at `iiss/testdata/seed_*.md` — refresh via `go test ./iiss/... -update.regression -run TestRegression`.

Known limitations (out of pass-2 scope):

- Post-stellar primaries (white dwarf / neutron star / black hole as the primary star) hit `stars.ErrPostStellarPrimaryUnsupported`. About 3-4% of random seeds produce them.
- Peculiar primary dispatch, giant-companion-MAO gaps, missing class-IV table cells — same family. Skipped via `isSpecialCircumstances(err)` in property tests.
- WBH pp.147+ (World Social Characteristics, Special Circumstances chapter) — explicitly out of scope.

## Documentation

- **Design pre-flight** (under `docs/pass-2/`): `design-intent.md` (the why and the cuts), `api-surface.md` (every public signature), `dependency-graph.md` (every value, its inputs, fixed-point clusters), `anti-patterns.md` (pre-flight checklist), `harness.md` (fixture catalog with status indicators), `wbh-inconsistencies.md` (six book-internal divergences with chosen interpretations), `spike-findings.md` (pre-implementation validation).
- **Post-pass retrospective**: `summary.md` (what was built), `lessons-learned.md` (13 numbered lessons from implementation), `next-steps.md` (what's open, what's closed, suggested orderings).
- **Pass-1 archive**: `docs/pass-1/{specs,plans,retrospective}/` — dated, chapter-numbered docs from the original implementation. Buildable via `git checkout pass-1-final`.

## Development

```bash
task                        # default: check + test
task check                  # modernize + gofumpt + go vet + golangci-lint
task test                   # go test -race ./...
task fmt                    # gofumpt -l -w -extra .
task tidy                   # go mod tidy

# Single test:
go test ./worlds/ -run TestZed_ApplyStage5

# Run CLI:
go run ./cmd/wbh -seed 42 -format markdown
```

`task check` runs `go fix ./...` first (modernizer pass) and fails on any diff — modernizer hints are mandatory. The gate is local; no CI.

Conventions: `gofumpt` formatting (`gofumpt -l -w -extra .`); typed `Roller` interface (no package-level RNG); WBH page citations in doc-comments; six book inconsistencies committed to specific interpretations in code (no runtime toggles for them, see `docs/pass-2/wbh-inconsistencies.md`).

## Source material

Every procedure, every table, and every formula references Mongoose Publishing's _World Builder's Handbook_ (Geir Lanesskog, 2023) as the canonical authority. WBH page numbers appear in doc-comments next to the tables and procedures they encode — that's the project's traceability mechanism. To work with this repository locally and verify procedure fidelity against the source, place a copy of the handbook PDF at `docs/World Builders Handbook.pdf` (gitignored — copyright).

Where the book contradicts itself (worked example vs formula box, table A vs table B), this implementation surfaces the divergence in `docs/pass-2/wbh-inconsistencies.md` rather than silently picking one. Each of the six known inconsistencies has a chosen interpretation with rationale and a test that asserts the chosen value.

## License

MIT — see `LICENSE`. The code is licensed for reuse. The _World Builder's Handbook_ itself is © Mongoose Publishing and is not redistributed here; users must obtain a copy of the handbook independently to verify procedure fidelity against the source.
