# World Builder — Design

**Date:** 2026-05-02
**Status:** architecture decisions approved through brainstorming; written spec pending user review
**Source material:** Mongoose Publishing, _World Builder's Handbook_ (Geir Lanesskog, 2023). PDF in repo at `Mongoose/Core Rules/World Builders Handbook.pdf`.

## Purpose

Encode the procedures of the _World Builder's Handbook_ (WBH) as a faithful Go reference implementation. Correctness, traceability, and testability take priority over user-facing polish. The library is the artifact; CLI and output formatting are secondary.

> **Language note:** This project initially started in Python (uv + ruff + pytest) and was switched to Go after the very first task hit a uv/macOS/Python 3.12 editable-install friction (`.pth` files inheriting the macOS hidden flag, which Python 3.12 then skips). Switching cost was negligible (one task of scaffolding) and Go's static binary plus zero packaging surface area outweighed Python's lighter syntax for heterogeneous tables in this case.

## Non-goals

- Not a sector generator (use the Sector Construction Guide for that).
- Not a worldbuilding sandbox with override-and-recompute semantics.
- Not a publishable workbook generator (no PDF/Markdown styling).
- Not a service. Local Python library only, used at a REPL or via tests.

## Scope and decomposition

The book has roughly five large procedural chapters. v1 of the project is the **Stars** chapter only. Each subsequent chapter is its own brainstorm → spec → plan → implementation cycle, layered on a shared core.

| Sub-project | WBH chapter                      | Status        |
| ----------- | -------------------------------- | ------------- |
| `_core`     | shared infrastructure            | built with v1 |
| `stars`     | Stars (pp. 14–35)                | **v1**        |
| `orbits`    | System Worlds and Orbits (36–68) | future        |
| `physical`  | World Physical (69–146)          | future        |
| `social`    | World Social (147–218)           | future        |
| `special`   | Special Circumstances (219–234)  | future        |

Chapters depend strictly downstream: orbits consumes stars, physical consumes orbits, etc. Never the reverse.

## Project location and packaging

- **Location:** `Traveller/` inside the existing Traveller repo. The repo is local-only (no remote) and already ignores PDFs/epubs; adding a Go subproject does not disrupt that.
- **Module path:** `wbh` (local-only module; no public hosting). Each chapter is its own package: `wbh/dice`, `wbh/roller`, `wbh/stars`, etc.
- **Toolchain:** `go` (1.22+) for build/test, `gofumpt` for formatting (stricter superset of `gofmt`), `golangci-lint` for linting, `go test ./...` for tests, `go test -race ./...` before commits. No CI.
- **Entry points:** importable packages (`import "wbh/stars"`) plus a thin `cmd/wbh/main.go` CLI that emits JSON or a text summary. The CLI is one screen of code.
- **Spec location:** this file lives at `Traveller/docs/specs/`, not inside ``, because it predates and outlives any single sub-project.

## Architecture

### Public API: pure-function pipeline

```go
import (
    "wbh/roller"
    "wbh/stars"
)

r := roller.NewSeeded(42)
sys := stars.GenerateSystem(r)
// stars.System{Primary: stars.Star{...}, Companions: []stars.Companion{...}}
```

Each `Generate*` takes upstream results and a `roller.Roller`, returns immutable value-types (no pointers escaping into the result graph). No package-level state. A seed fully determines the system. Variance options (the book's optional ±20%/±30% rolls) are passed via an `Options` struct, off by default.

### Package layout

```text

├── go.mod
├── go.sum
├── README.md
├── justfile                # check, test, fmt
├── dice/
│   ├── dice.go             # ParseDice, dice notation helpers
│   └── dice_test.go
├── roller/
│   ├── roller.go           # Roller interface + Seeded, Scripted, Fixed
│   └── roller_test.go
├── stars/
│   ├── types.go            # SpectralType, LuminosityClass, Star, StarKind
│   ├── tables.go           # all Stars-chapter tables
│   ├── primary.go          # Star Type Determination + Subtype + class restrictions
│   ├── physical.go         # mass, diameter, temperature, luminosity, interpolation
│   ├── ages.go             # main sequence / subgiant / giant / final ages
│   ├── peculiar.go         # brown dwarf, white dwarf, NS, BH, etc.
│   ├── stars.go            # public entry points (GenerateMainSequenceStar, etc.)
│   └── *_test.go
└── cmd/wbh/
    └── main.go             # CLI
```

### Determinism and the Roller

Every dice roll in the library passes through a `roller.Roller` interface. There are no package-level `math/rand` or `crypto/rand` calls; this is enforced by `go vet` plus a project lint rule (`golangci-lint` `gochecknoglobals` + manual review) banning RNG outside `roller/`.

```go
type Roller interface {
    Roll(dice string) int
}

type Seeded struct{ /* backed by *rand.Rand */ }
type Scripted struct{ /* yields preset values; panics on exhaustion */ }
type Fixed struct{ /* every Roll returns same N */ }
```

`dice` notation is parsed by `dice.Parse("2D-7")` (roll 2d6, subtract 7). Dice modifiers (`DM+1` etc.) are passed as a separate argument, not embedded in the dice string, so the audit trail stays clean.

### Traceability

Per design decision, traceability is via **source comments only** — no runtime metadata wrapper. Every procedure and table has a doc-comment citing the WBH page and table name. The verification weight sits in the test suite (see below).

### Data shape

Tables live as typed Go literals in `tables.go` files, with line comments citing the source. Heterogeneous rows (where some cells are absent — the book's "—") use `*float64` for nullable numeric cells; struct-typed rows for tables with named columns. Example:

```go
// WBH p.15 Table: Star Type Determination
type StarTypeRow struct {
    Type, Hot, Special, Unusual, Giants, Peculiar string
}

var StarTypeDetermination = map[int]StarTypeRow{
    2:  {Type: "Special", Hot: "A", Special: "Class VI", Unusual: "Peculiar", Giants: "Class III", Peculiar: "Black Hole"},
    3:  {Type: "M",       Hot: "A", Special: "Class VI", Unusual: "Class VI", Giants: "Class III", Peculiar: "Pulsar"},
    // ...
}
```

For tables with optional cells (the book uses "—" to mean "this combination does not occur"):

```go
// WBH p.17 Table: Star Mass and Temperature by Class — Mass column
// Nil pointers indicate cells the book leaves blank.
type MassRow struct {
    Ia, Ib, II, III, IV, V, VI *float64
}

var f = func(x float64) *float64 { return &x }

var StarMass = map[string]MassRow{
    "O0": {Ia: f(200), Ib: f(150), II: f(130), III: f(110), V: f(90),  VI: f(2)},
    "B0": {Ia: f(60),  Ib: f(40),  II: f(30),  III: f(20),  IV: f(20), V: f(18), VI: f(0.5)},
    // ...
}
```

No external YAML/TOML data files in v1. Tables are small, compiled into the binary, and read well next to the book. Revisit per-chapter if a future table grows unwieldy.

## Stars sub-project (v1) — what we build first

### Coverage

- **In v1 — the full Stars chapter (pp. 14–35):**
  - Primary star generation (Star Type Determination table, all six columns and redirects).
  - Star Subtype rolls (numeric and M-type variants, plus the K-type Class IV adjustment).
  - Class restrictions and re-rolls (Class IV limited to B0–K4 with M-row remapping; Class VI limited to F or A; Class III+ giants column).
  - Mass, diameter, temperature, luminosity by class and subtype, with linear interpolation between coarse table grid points (table-based and formula-based, per the book).
  - Optional ±20% (mass/diameter) and ±30% (luminosity) variance rolls, off by default.
  - Peculiar/special objects as primary stars: brown dwarfs, white dwarfs, neutron stars, black holes, pulsars, nebulae, protostars, star clusters, anomalies (limited to what the Stars chapter specifies).
  - System age (Main sequence / subgiant / giant / final age, per-class formulas; small star and large star age rolls; Special and Unusual Object Age by Type).
  - Multiple-star presence (Close/Near/Far/Companion) and the Existing Star Locations tables for binary and three-or-more-star continuation.
  - Non-Primary Star Determination (Sibling/Twin/Lesser/Random) for secondaries and companions.
  - Stellar Orbit# placement (1D-1 / 1D+5 / 1D+11 / 1D÷10+(2D-7)÷100), eccentricity (with all DM modifiers), inclination, and orbit period via Kepler's third law.
  - Star designations (Aa/Ab/B/Ca/Cb/D and barycentre composites Aab/AB/Cab/ABC).
  - Orbit# ↔ AU conversion utilities.
  - IISS Class 0/I Survey form output (Form 0421B-0I), as a structured record matching p. 33–34.

- **Deferred to later sub-projects:**
  - HZCO (Habitable Zone Centre Orbit). Physically a stellar property but only consumed by the world-orbits chapter; defer so its API is shaped by its consumer.
  - Detailed special-object physics (full white dwarf cooling tables, pulsar timing, etc., from the Special Circumstances chapter).

### Modeling the branching tables

The Star Type Determination table on p. 15 has six columns (Type / Hot / Special / Unusual / Giants / Peculiar) where some results redirect to other columns. Model each column as a pure function returning a `TypeRollResult` that is either a final spectral letter or a redirect token; a small dispatcher follows redirects to a fixed depth. This keeps every table a flat lookup readable next to the book.

### Interpolation

Mass, diameter, temperature, luminosity are tabulated at coarse subtype grid points (O0, O5, B0, B5, …). The book endorses interpolation ("a G7 is 2/5 of the difference between G5 and K0"). One shared `interpolate(table, spectral_type, luminosity_class)` helper backs all four physical-quantity functions.

## Testing strategy

This is where "faithful reference" lives.

1. **Worked-example regression tests.** The book threads an in-line example (the G7 V star "Zed" in the Storr sector) through every step of the Stars chapter. Encode that example — and every subsequent in-line example in future chapters — as a test that drives a `roller.Scripted` with the exact dice sequence the book describes, and asserts every output value matches the book to the digit. Breaking a worked example breaks `go test`. These tests are the proof of fidelity.
2. **Table integrity tests.** For each table: completeness (no missing cells in the documented domain), monotonicity where the book asserts it, column totals where applicable. Use Go's table-driven test idiom (`for _, tc := range cases { ... }`).
3. **Property tests** (lightweight, via `testing/quick` from the standard library, or `pgregory.net/rapid` if heavier shrinking is needed): for arbitrary seeds, generated stars have valid spectral type, valid class for the chosen branch, mass/diameter/luminosity in plausible ranges, etc.

## Open questions for future sub-projects

These are flagged here so the v1 design does not foreclose them:

- **Roller scope.** v1 uses one `Roller` for the whole pipeline. If reproducing a single chapter independently becomes important, we may want sub-rollers (one per chapter) seeded from the parent so `GenerateOrbits` can be re-run without re-rolling stars. Defer until orbits.
- **Caching of physical-quantity lookups.** Interpolation is cheap; not worth caching in v1. Revisit if profiling later sub-projects shows otherwise.
- **CLI output schema.** v1 emits whatever JSON falls out of `encoding/json` on the `stars.System` struct. A stable JSON schema for cross-tool consumption is a future concern.

## Success criteria

- The Zed quintuple-star worked example reproduces every value in the IISS Class 0/I Survey form on WBH p. 34, to the digit, when driven by a `ScriptedRoller` with the book's roll sequence.
- The single-star "Terra/Sol" reference example on p. 35 round-trips through the same code path with no rolls (it is fully specified) and matches.
- The two-star "Corella" example on p. 35 reproduces given the book's stated rolls.
- All tables in the Stars chapter (Star Type Determination, Star Subtype, Star Mass and Temperature, Star Diameter, Star Luminosity, Multiple Stars Presence, Existing Star Locations × 2, Non-Primary Star Determination, Eccentricity Values, Inclination, Stellar Orbit# Ranges, Special and Unusual Object Age, Orbit# table) are encoded in `tables.go` files with WBH page citations.
- A fresh checkout runs `just check && just test` clean (`golangci-lint run` + `go test -race ./...`).
- A reader with the book open can match every exported function and table in package `stars` to a specific page or table in the Stars chapter.
