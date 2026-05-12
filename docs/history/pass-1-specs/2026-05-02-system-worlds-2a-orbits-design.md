# System Worlds and Orbits — Sub-project 2A Design (Available Orbits + HZCO)

**Date:** 2026-05-02
**Status:** approved through brainstorming; pending user review of written spec
**Source material:** Mongoose Publishing, _World Builder's Handbook_ (Geir Lanesskog, 2023). PDF in repo at `Mongoose/Core Rules/World Builders Handbook.pdf`.
**Source pages:** WBH pp. 38–43.
**Parent spec:** `docs/history/pass-1-specs/2026-05-02-world-builder-design.md`.

## Purpose

Encode the Available Orbits procedures and the Habitable Zone Centre Orbit# (HZCO) computation from the _World Builder's Handbook_ chapter "System Worlds and Orbits." Sub-project 2A is the foundation that subsequent placement work (2B) and sizing/moons/survey work (2C) build on.

## Decomposition context

The "System Worlds and Orbits" chapter (WBH pp. 36–68) is too large for a single brainstorm → spec → plan cycle. It is decomposed into three sub-projects, all layered downstream of the existing Stars chapter:

| Sub-project | WBH pp.      | Scope                                                                                              |
| ----------- | ------------ | -------------------------------------------------------------------------------------------------- |
| **2A**      | 38–43        | Available Orbits + HZCO. **This spec.**                                                            |
| 2B          | 36–38, 43–52 | World counts (gas giants/belts/terrestrials) and the 8-step placement procedure. Future spec.      |
| 2C          | 54–67        | Terrestrial/gas-giant sizing, significant moons, planetary profile, IISS Class II/III survey form. |

2B consumes 2A; 2C consumes 2B. Never the reverse.

## Non-goals

- Not the Hill-sphere physics-based orbit method (WBH pp. 40–41). The chapter calls it more realistic but acknowledges the n-body problem has no general solution; the simplified 11-rule method on pp. 38–40 is the only path implemented in 2A. The Hill-sphere method may be added later as an `Options` flag once the simple paths are proven.
- Not world placement (2B), eccentricity rolls for planets (2B/Stars overlap), sizing or moons (2C), or any IISS form output beyond what already exists in `stars/survey.go`.
- Not post-stellar primary MAO support. The WBH p. 39 table explicitly defers MAO for Brown Dwarf, White Dwarf, Neutron Star, Black Hole, and Pulsar primaries to the Special Circumstances chapter.

## Architecture

### Public API

Two new functions in package `wbh/stars`:

```go
// HZCO returns the Habitable Zone Centre Orbit# for a single star,
// computed from its luminosity by the WBH p. 41 formula:
//
//	HZCO_AU    = sqrt(luminosity)
//	HZCO_Orbit = AUToOrbit(HZCO_AU)
//
// The p. 42 HZCO table is encoded as a test fixture only.
func (s Star) HZCO() float64

// CompositeHZCO returns the HZCO# for a circumbinary group of stars
// orbiting a shared barycentre. Per WBH p. 42, the luminosities of all
// stars interior to the planet's orbit are summed, then the formula
// applies to the combined luminosity.
func CompositeHZCO(stars ...Star) float64
```

One new package `wbh/worlds` with:

```go
package worlds

import "wbh/stars"

// Interval is a closed Orbit# range [Min, Max].
type Interval struct {
    Min, Max float64
}

// Group is one body or barycentric pair sharing an orbit set.
//
// Single-star group: Members has one Star, Designation is "A"/"B"/"C"/"D".
// Pair group:        Members has two Stars (parent first, companion second),
//                    Designation is "Aab"/"Cab"/...
type Group struct {
    Designation string
    Members     []stars.Star
    MAO         float64    // from p. 39 table; for pairs, raised by rule 2 if applicable
    Intervals   []Interval // disjoint, sorted ascending
}

// Total returns the sum of (Max - Min) over all intervals — the value the book
// calls "total Orbit#s" used in placement allocation (sub-project 2B).
func (g Group) Total() float64

// Contains reports whether orbit is inside any of g.Intervals.
func (g Group) Contains(orbit float64) bool

// Result is the per-group available orbits for an entire system.
type Result struct {
    Groups []Group // ordered by ascending stellar Orbit# of the group's outer member
}

// AvailableOrbits applies the 11 simplified rules from WBH pp. 38–40 to a
// stars.System and returns per-group allowed Orbit# intervals.
//
// Returns ErrPostStellarPrimaryUnsupported if the primary is a Brown Dwarf,
// White Dwarf, Neutron Star, Black Hole, or Pulsar (their MAO is in the
// Special Circumstances chapter, not yet encoded).
func AvailableOrbits(sys stars.System) (Result, error)

var ErrPostStellarPrimaryUnsupported = errors.New(
    "worlds: post-stellar primary MAO requires Special Circumstances chapter",
)
```

### Group identification

`AvailableOrbits` partitions a `stars.System` into groups:

- **Group 1 (always present):** the primary plus its `Companion`-class companion if any. Designation is `"A"` if alone, `"Aab"` if paired.
- **For each `CompanionStar` with `OrbitClass` ∈ {Close, Near, Far}:** that secondary plus its own `Companion`-class companion if any. Designations are positionally renumbered to `"B"`/`"Bab"`/`"C"`/`"Cab"`/`"D"`/`"Dab"` in the same way the existing `stars.AssignDesignations` already does.
- A `CompanionStar` with `OrbitClass == Companion` is folded into its parent's group, never its own group.

The 11-rule method then assigns each group its MAO and intervals.

### The 11 rules (WBH pp. 38–40)

Implemented as a sequence of mutations on each group's interval set, in the order the book lists them. An internal helper type `intervalSet` provides `subtract(Interval)` and `clamp(Interval)` operations. Spec numbering matches the book exactly:

1. **MAO** for each star or pair from the p. 39 table (interpolated by spectral type within a luminosity-class column).
2. **Companion eccentricity** raises the lower bound for a pair: orbits less than `0.50 + companion_eccentricity` are unavailable. If the larger star's MAO > 0.2, add the larger star's MAO to the unavailable lower zone.
3. **Primary outer bound** is Orbit# 20.
4. **For each Close/Near/Far secondary,** note its Orbit# and treat its companion (if any) as occupying the same Orbit#. Companions are ignored from this point forward; the secondary+companion pair is treated as a unit at the secondary's Orbit#.
5. **Each Close/Near/Far secondary** at Orbit# `s` excludes the open range `(s − 1, s + 1)` from the primary's available orbits. If the secondary has MAO > 0.2, add the secondary's MAO to that exclusion.
6. **Secondary eccentricity > 0.2** widens that secondary's exclusion of the primary by an additional ±1 Orbit# on each side.
7. **Close/Near (not Far) secondary eccentricity > 0.5** widens the primary's exclusion by another ±1 Orbit# on each side.
8. **Each Close/Near/Far secondary** has its own range centred on itself, extending up to (its Orbit# − 3) on each side.
9. **Adjacent-zone reduction:** any Close/Near/Far secondary with a populated adjacent zone (Close+Near or Near+Far, but not Close+Far without Near) loses 1 Orbit# from its range. The primary never triggers this for secondaries. Triggers at most once per secondary regardless of how many adjacent zones are populated.
10. **Adjacent-zone eccentricity reduction:** if a secondary or any adjacent-zone star has eccentricity > 0.2, that secondary loses another Orbit#. Triggers at most once per secondary.
11. **Self eccentricity > 0.5** further reduces that secondary's range by 1 Orbit#. Triggers at most once per secondary.

### Stars carry-forward: giant-primary support

Sub-project 2A absorbs one Stars-chapter cleanup that intersects with the MAO table:

- `stars.RollSpecialPrimary` currently returns `ErrSpecialPrimaryClassRedirect` when a Special-column roll lands on "Class III"/"IV"/"VI"/"Giants". The function signature changes from `(StarKind, error)` to `(Star, error)`, and on a class redirect it re-rolls 2D on the regular Star Type Determination table at the indicated class, returning a fully-resolved `Star`.
- `stars.GenerateSystem` no longer needs the redirect-handling branch; the deletion is ~15 lines.
- `ErrSpecialPrimaryClassRedirect` is removed from the codebase entirely.

This is included in 2A because the p. 39 MAO table covers all luminosity classes including giants (Ia/Ib/II/III), and meaningfully testing the table against giant rows requires being able to generate a giant primary.

The unrelated `Other`-descriptor wart in `stars.GenerateCompanionStar` is **not** included in 2A. It is its own small follow-up.

## File layout

```text

├── stars/
│   ├── hzco.go                    NEW   Star.HZCO(), CompositeHZCO()
│   ├── hzco_test.go               NEW   p.42 table fixture + worked examples
│   ├── peculiar.go                EDIT  absorb class-redirect into RollSpecialPrimary
│   ├── peculiar_test.go           EDIT  redirect cases
│   ├── system.go                  EDIT  drop ErrSpecialPrimaryClassRedirect handling
│   └── system_test.go             EDIT  giant-primary multi-star case
└── worlds/                        NEW PACKAGE
    ├── available_orbits.go        NEW   Group, Interval, Result, AvailableOrbits()
    ├── available_orbits_test.go   NEW   per-rule unit tests
    └── worked_examples_test.go    NEW   Sol + Zed acceptance gates
```

## Testing strategy

### `stars/hzco_test.go`

- **Property test:** for every populated cell of the p. 42 HZCO table, `Star{T, Class: C, Luminosity: tableLuminosityForCell}.HZCO()` reproduces the cell within ±5% (the book's "validating the close enough approach" tolerance).
- **Worked-example assertions** to ±0.05 Orbit#:
  - Sol (G2 V, L=1.000) → 3.0
  - Zed Aab (composite, L=1.419) → 3.3
  - Zed B (K8 V, L=0.136) → 0.92
  - Zed Cab (composite, L=0.0896) → 0.75
  - Corella Aab (composite, L=1.725) → 3.5

### `stars/peculiar_test.go` (extension)

Synthetic redirect cases (the book has no worked example with a giant primary):

- Special column rolls 2D=4 → "Class III" → re-roll 2D=7 on regular table → K-type at Class III with normally-rolled subtype.
- Special column rolls 2D=12 → "Class VI" → regular table re-roll restricted to F or A per existing Class VI rules.

### `stars/system_test.go` (extension)

- `TestGenerateSystem_GiantPrimary`: primary is Class III (constructed via `Compose`), Multiple Stars Presence still works (Class III blocks Close per existing rule), MAO for available-orbits computation reads the giant rows correctly.

### `worlds/available_orbits_test.go`

Per-rule unit tests using synthetic `stars.System` values built via `Compose`:

- Rule 2: companion eccentricity 0.11 → pair lower bound 0.61.
- Rule 3: primary outer bound 20.
- Rule 5: secondary at Orbit# 6.10 excludes `(5.10, 7.10)` from the primary.
- Rule 6: secondary ecc > 0.2 widens that exclusion by ±1.
- Rule 7: Close/Near ecc > 0.5 widens by another ±1; Far does not trigger.
- Rule 8: secondary's own range = its Orbit# − 3, centred.
- Rules 9–11: adjacent-zone reductions, adjacent-zone eccentricity reductions, self-ecc > 0.5 reductions. Each triggers at most once per secondary.
- `Group.Total()` and `Group.Contains()` behave correctly across multi-interval groups.

### `worlds/worked_examples_test.go` — acceptance gates

Both tests construct stars via `stars.Compose` (no rolls; this is a pure constraint verification, not a regression of the Stars roll sequence).

- `TestSol_AvailableOrbits`: single G2 V primary, no companions. Asserts one group `"A"` with MAO 0.03 and intervals `[[0.03, 20.00]]`.
- `TestZed_AvailableOrbits`: quintuple system. Asserts:
  - 3 groups in order: `"Aab"`, `"B"`, `"Cab"`.
  - Aab: MAO 0.61, intervals `[[0.61, 5.10], [7.10, 10.10], [14.10, 20.00]]`, Total 13.39.
  - B: MAO 0.02, intervals `[[0.02, 1.10]]`, Total 1.08.
  - Cab: MAO 0.74, intervals `[[0.74, 7.10]]`, Total 6.36.
- `TestZed_HZCO_Composite`: `CompositeHZCO(Aa, Ab) == 3.3`, `Star{B}.HZCO() == 0.92`, `CompositeHZCO(Ca, Cb) == 0.75`, all to ±0.05.

**Tolerances:** Orbit# values to ±0.01 (book gives them to two digits); HZCO to ±0.05; Total to ±0.05.

## Open questions for future sub-projects

- **Hill-sphere method.** The Hill-sphere alternate (WBH pp. 40–41) is deferred. If a future sub-project needs it, expose it as `worlds.AvailableOrbitsOpts{ Method: HillSphere }` rather than a separate function.
- **Post-stellar primary MAO.** Brown Dwarf, White Dwarf, Neutron Star, Black Hole, and Pulsar primaries get MAO from the Special Circumstances chapter, not yet encoded. `ErrPostStellarPrimaryUnsupported` is the placeholder; revisit when that chapter is built.
- **Available-orbits API for 2B placement.** 2B's allocation step needs `Group.Total()` per group plus the ability to detect which Orbit#s are blocked (`Group.Contains()`). If 2B needs additional queries (e.g., "next available Orbit# above X"), add them then rather than speculating now.
- **`Other`-descriptor wart in `GenerateCompanionStar`.** Tracked as a separate small follow-up; not part of 2A.

## Success criteria

- `stars.Star.HZCO()` and `stars.CompositeHZCO()` reproduce all five worked-example HZCO values within ±0.05 Orbit#.
- The p. 42 HZCO table is reproduced by the formula within ±5% across all populated cells.
- `worlds.AvailableOrbits(sys)` reproduces the Sol single-star case (`[[0.03, 20.00]]`) and the Zed quintuple's three groups exactly to the book's two-digit precision.
- `ErrSpecialPrimaryClassRedirect` is removed from the codebase. `RollSpecialPrimary` returns a fully-resolved `Star`.
- A fresh checkout of `` runs `just check && just test` clean (gofumpt + golangci-lint v2.12.1 + `go test -race ./...`).
- A reader with the book open can match every exported function in `stars/hzco.go` and `worlds/available_orbits.go` to a specific page or rule in WBH pp. 38–43.
