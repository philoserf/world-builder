# System Worlds and Orbits — Sub-project 2B Design (World Counts + Placement)

**Date:** 2026-05-03
**Status:** approved through brainstorming; pending user review of written spec
**Source material:** Mongoose Publishing, _World Builder's Handbook_ (Geir Lanesskog, 2023). PDF in repo at `Mongoose/Core Rules/World Builders Handbook.pdf`.
**Source pages:** WBH pp. 36–38 (World Types and Quantities) and pp. 43–52 (Placement of Worlds, Steps 1–9).
**Parent spec:** `docs/history/pass-1-specs/2026-05-02-world-builder-design.md`.
**Predecessor:** `docs/history/pass-1-specs/2026-05-02-system-worlds-2a-orbits-design.md` (Available Orbits + HZCO).

## Purpose

Encode the World Counts and 8-step World Placement procedures from the _World Builder's Handbook_ chapter "System Worlds and Orbits." Sub-project 2B layers on top of 2A's Available Orbits + HZCO and feeds 2C's sizing, moons, and IISS Class II/III survey form.

## Decomposition context

The chapter is decomposed into three sub-projects, all layered downstream of the existing Stars chapter:

| Sub-project | WBH pp.      | Status                                                                                                     |
| ----------- | ------------ | ---------------------------------------------------------------------------------------------------------- |
| 2A          | 38–43        | **Done** (merged 2026-05-02). Available Orbits + HZCO.                                                     |
| **2B**      | 36–38, 43–52 | **This spec.** World counts + 8-step placement (Steps 1–9).                                                |
| 2C          | 54–67        | Future. Terrestrial/gas-giant sizing, significant moons, planetary profile, IISS Class II/III survey form. |

2B consumes 2A; 2C will consume 2B. Never the reverse.

## Non-goals

- **Continuation Method.** No pre-existing-mainworld branches. Specifically excludes: "Pre-Existing Mainworlds" (p. 36), "Known Worlds in Existing Systems" (p. 36), Planetoid Belts continuation tweak (p. 37), Terrestrial Planets continuation tweak (p. 38), Step 3d (p. 46), and Step 8 mainworld special cases (moon of GG, size-1 in belt, atmosphere-DM raw-temp reverse-engineering, pp. 51–52). A `worlds.ErrContinuationMethodUnsupported` placeholder is returned if a future `stars.System` field signals pre-existing mainworld data.
- **Optional Rules.** Per-star baseline number + spread (Step 5 sidebar, p. 49); 10–20% minimum-separation in compact systems (Step 5 sidebar, p. 49). Both are explicitly labelled "Optional Rule" in the book and the Zed worked example does not use them.
- **Referee judgment overrides.** No interface or callback for intercepting rolls. Tests follow strict dice. The single Zed worked-example divergence (the empty orbit at C1 instead of the book's narrated C2) is documented inline in `worked_examples_test.go`.
- **Sizing, moons, profile, IISS Class II/III survey form** — all 2C.
- **Hill-sphere alternate orbit method** (WBH pp. 40–41) — still deferred from 2A.
- **Post-stellar primary MAO** — still deferred from 2A.

## Architecture

### Public API — package `wbh/worlds`

Granular per-step functions + a thin façade. Each function maps to a named WBH page or step.

```go
package worlds

// ---- World Counts (pp. 36–38) ------------------------------------------

type Counts struct {
    GasGiants      int // 0–6
    PlanetoidBelts int // 0–3
    Terrestrials   int // 0–13 (13 cap from Step 7 narrative)
    Total          int // sum
}

// CountsOpts is empty for now. Gas-giant existence method is hardcoded
// to the standard CRB form ("Gas Giant Exists on 9-: roll 2D"); the
// alternate 1D 2+ form is left for a later opts addition if needed.
type CountsOpts struct{}

func GenerateCounts(r roller.Roller, sys stars.System, opts CountsOpts) (Counts, error)

// ---- Step 1: Allocations by Star (pp. 43–44) ---------------------------

type StarAllocation struct {
    Group           Group
    TotalStarOrbits int // floor(sum of intervals + (1 if no companion and >0 prior allowable))
    AllocatedWorlds int // share of Counts.Total
}

func AllocateOrbitsByStar(avail Result, counts Counts) ([]StarAllocation, error)

// ---- Step 2: System Baseline Number (pp. 44–45) ------------------------

func RollBaselineNumber(r roller.Roller, sys stars.System, counts Counts) (int, error)

// ---- Step 3: System Baseline Orbit# (pp. 45–46) ------------------------

// BaselineOrbit selects sub-case 3a / 3b / 3c by the value of baselineN
// relative to totalWorlds. hzco is the primary group's HZCO. Snaps the
// formula result to the nearest available orbit (with (2D-7)/10 direction
// variance) when it lands in an excluded zone.
func BaselineOrbit(
    r roller.Roller,
    primary Group,
    hzco float64,
    baselineN, totalWorlds int,
) (float64, error)

// ---- Step 4: Empty Orbits (p. 48) --------------------------------------

func RollEmptyOrbits(r roller.Roller) (int, error)

// ---- Step 5: System Spread (pp. 48–49) ---------------------------------

// Spread implements the base formula (BaselineOrbit - MAO) / BaselineN
// (with BaselineN treated as 1 if < 1). When the resulting spread would
// place the outermost primary slot beyond Orbit# 20, it is replaced by
// the mandatory Maximum Spread formula (p. 48):
//
//   Maximum Spread = primary.Total() / (primaryAllocated + totalStars)
//
// totalStars counts primary + secondaries; companions don't count.
func Spread(
    primary Group,
    primaryAllocated int,
    baselineOrbit float64,
    baselineN, totalStars int,
) float64

// MaximumSecondarySpread is the per-secondary cap (p. 49):
//
//   = (Outermost Allowable Orbit# - secondary.MAO) / (secondaryAllocated + 1)
//
// PlaceOrbitSlots applies this cap when the system spread would push the
// outermost slot for a secondary star past its outer allowable bound.
func MaximumSecondarySpread(secondary Group, secondaryAllocated int) float64

// ---- Step 6: Placing Orbit#s (pp. 49–50) -------------------------------

type Slot struct {
    StarSlot string // "A1", "A2", ..., "B+", "C5". The "+" suffix marks
                    // the extra slot added to a star by Step 4's empty-bump.
                    // Step 7 anomalous slots get their own "+" suffix when
                    // placed; the AnomalousSlot.Anomaly field disambiguates.
    Group    Group
    Orbit    float64
}

// PlaceOrbitSlots walks each StarAllocation in order, placing slots from
// MAO outward by (Spread + (2D-7) * Spread/10), with the baseline-N-th
// slot of the primary group fixed at baselineOrbit, and widening the
// spread by the width of any exclusion zone hit (mandatory, p. 49).
//
// Empty orbits (from Step 4) are distributed across allocations before
// placement per p. 48: Close star first, then Near, then Far, with any
// remainder going to the primary. The output Slot.Empty flag marks each
// resulting placed slot that originated from Step 4. Slot.StarSlot ids
// follow the book's convention: "A1", "A2", ..., with "+" suffix
// ("A+", "B+") marking the empty slot inserted at its insertion point.
func PlaceOrbitSlots(
    r roller.Roller,
    allocs []StarAllocation,
    baselineOrbit, spread float64,
    emptyOrbits int,
) ([]Slot, error)

// ---- Step 7: Anomalous Planets (pp. 50–51) -----------------------------

type AnomalyType int

const (
    AnomalyNone AnomalyType = iota
    AnomalyRandom
    AnomalyEccentric
    AnomalyInclined
    AnomalyRetrograde
    AnomalyTrojan
)

type AnomalousSlot struct {
    Slot
    Anomaly        AnomalyType
    InclinationDeg float64 // for AnomalyInclined
    TrojanOf       string  // StarSlot id of parent, for AnomalyTrojan
    EccentricityDM int     // additional DM applied in Step 9
}

// AddAnomalous rolls anomalous-orbits count + per-anomaly type, picks
// star groups (D3 in multi-star), rolls 2D-2 + d10/10 for orbit value,
// clamps to [MAO, 20.0] with retry. Updates terrestrial and total counts
// per the book (each anomaly adds one terrestrial).
func AddAnomalous(
    r roller.Roller,
    slots []Slot,
    allocs []StarAllocation,
    counts Counts,
) ([]AnomalousSlot, Counts, error)

// ---- Step 8: Placing Worlds (pp. 51–52) --------------------------------

type BodyType int

const (
    BodyEmpty BodyType = iota
    BodyTerrestrial
    BodyGasGiant
    BodyPlanetoidBelt
)

type Placement struct {
    AnomalousSlot
    Body       BodyType
    PrefixRoll string // "1:6", "2:3"... (audit trail)
}

// PlaceWorlds: order is empty -> gas giants -> planetoid belts ->
// terrestrials. Picks D2/D3/1D prefix die based on slot count per group.
// On 1D:1D collision, +1 to right die, then move to next slot id, then
// wrap to first empty slot.
func PlaceWorlds(
    r roller.Roller,
    slots []AnomalousSlot,
    counts Counts,
) ([]Placement, error)

// ---- Step 9: Planet Eccentricity (p. 52) -------------------------------

// RollPlanetEccentricities calls stars.RollEccentricity for each non-belt
// non-empty placement, applying anomaly DMs from AnomalousSlot.EccentricityDM.
func RollPlanetEccentricities(r roller.Roller, ps []Placement) ([]Placement, error)

// ---- Façade ------------------------------------------------------------

type SystemPlacement struct {
    Counts        Counts
    Allocations   []StarAllocation
    BaselineN     int
    BaselineOrbit float64
    EmptyOrbits   int
    SystemSpread  float64
    Placements    []Placement
}

func GenerateSystemPlacement(r roller.Roller, sys stars.System) (SystemPlacement, error)

// ---- Errors ------------------------------------------------------------

var ErrContinuationMethodUnsupported = errors.New(
    "worlds: pre-existing mainworld input requires Continuation Method, not yet encoded",
)
```

### Group extension (carry-forward refactor from 2A)

```go
type Group struct {
    Designation string
    Members     []stars.Star
    MAO         float64
    Intervals   []Interval

    // NEW: source companion star (nil for the primary group). Removes the
    // secIdx positional fragility flagged by 2A's final code review, and
    // gives 2B steps direct access to the secondary's orbit-around-primary.
    sourceCompanion *stars.CompanionStar
}

// HZCO returns the group's habitable zone centre orbit#. Single-star
// group: == Members[0].HZCO(). Pair group: == stars.CompositeHZCO(Members...).
// NEW in 2B; lives on Group because Step 3 needs it per group.
func (g Group) HZCO() float64
```

`available_orbits.go` rules 9–11 are refactored to look up the source companion via `g.sourceCompanion` instead of walking a parallel `secIdx`. The existing 2A Zed available-orbits acceptance test must continue to pass unchanged.

## File layout

```text

├── stars/
│   └── (no changes)
├── worlds/
│   ├── available_orbits.go         EDIT  attach sourceCompanion to Group; refactor rules 9–11
│   ├── available_orbits_test.go    EDIT  add sourceCompanion assertion
│   ├── group_hzco.go               NEW   Group.HZCO() method
│   ├── group_hzco_test.go          NEW   per-group HZCO unit tests
│   ├── counts.go                   NEW   Counts, CountsOpts, GenerateCounts
│   ├── counts_test.go              NEW   per-rule + Zed = 17 worlds
│   ├── allocations.go              NEW   StarAllocation, AllocateOrbitsByStar
│   ├── allocations_test.go         NEW   per-rule + Zed allocation
│   ├── baseline.go                 NEW   RollBaselineNumber, BaselineOrbit
│   ├── baseline_test.go            NEW   per-rule + Zed baseline
│   ├── empty_orbits.go             NEW   RollEmptyOrbits
│   ├── empty_orbits_test.go        NEW   table fixture
│   ├── spread.go                   NEW   Spread, MaximumSecondarySpread
│   ├── spread_test.go              NEW   per-rule + Zed spread
│   ├── orbit_slots.go              NEW   Slot, PlaceOrbitSlots
│   ├── orbit_slots_test.go         NEW   per-rule + Zed orbit#s
│   ├── anomalous.go                NEW   AnomalyType, AnomalousSlot, AddAnomalous
│   ├── anomalous_test.go           NEW   per-rule + Zed anomalous
│   ├── placement.go                NEW   BodyType, Placement, PlaceWorlds
│   ├── placement_test.go           NEW   per-rule
│   ├── planet_eccentricity.go      NEW   RollPlanetEccentricities
│   ├── planet_eccentricity_test.go NEW   reuses stars.EccentricityValues
│   ├── system_placement.go         NEW   SystemPlacement, GenerateSystemPlacement façade
│   └── worked_examples_test.go     EDIT  add TestZed_FullPlacement
└── cmd/wbh/
    └── (no changes — CLI integration deferred until 2C delivers IISS Class II/III form)
```

## Testing strategy

Same TDD pattern as 2A: per-rule unit tests against synthetic inputs first, then the Zed worked example as the integration acceptance gate.

### Per-step unit tests

Highlights — each step file gets a `_test.go` covering each rule the book lists.

- **`counts_test.go`:** Gas Giant Quantity table (every band), every Gas Giant DM (single Class V star DM+1, brown-dwarf primary DM-2, post-stellar primary DM-2, per-post-stellar DM-1, 4+ stars DM-1); Planetoid Belt Quantity table + every belt DM (≥1 GG DM+1, protostar DM+3, primordial DM+2, post-stellar DM+1, per-post-stellar DM+1, 2+ stars DM+1); Terrestrial Planets formula `2D-2 + DMs` with the `<3 → reroll D3+2` and `≥3 → +D3-1` branches; post-stellar DMs. Zed integration: returns `{4, 2, 11, 17}` with the book's stated rolls.

- **`allocations_test.go`:** Single-star case yields one allocation with all worlds. Multi-star Zed case reproduces Step 1 exactly: Aab=11, B=1, Cab=5 (from `Total Star Orbits` 13/2/6, sum 21; `Aab = ceil(17×13/21) = 11`; `B = floor(17×2/21) = 1`; `Cab = remainder = 5`). Note: the Step 4 empty-orbit allocation later bumps B from 1 to 2; that bump is exercised in `empty_orbits_test.go` and the integration test, not here.

- **`baseline_test.go`:** Step 2 DM table (companion DM-2, Ia/Ib/II DM+3, III DM+2, IV DM+1, VI DM-1, post-stellar DM-2, total-worlds bands, per-secondary DM-1). Step 3a `HZCO ≥ 1.0` and `< 1.0` paths. Step 3b `min ≥ 1.0` and `min < 1.0` paths. Step 3c both branches plus the "still negative → MAO + totalWorlds × 0.01" floor. Snap-to-nearest-available with `(2D-7)/10` direction variance when the formula lands in an excluded zone. Zed: baseline number 5 (DM −4, roll 9), HZCO 3.3, roll 5 (`2D−7 = −2`) → `3.1`.

- **`empty_orbits_test.go`:** the 9-/10/11/12 table.

- **`spread_test.go`:** Base formula `(BaselineOrbit − MAO) / BaselineN`; baseline N < 1 treated as 1; mandatory Maximum Spread cap; Maximum Secondary Spread cap. Zed: `(3.1 − 0.61) / 5 = 0.498 → 0.50`.

- **`orbit_slots_test.go`:** Inner Slot formula with `(2D-7) × Spread/10` variance; Next Slot formula; baseline-N-th slot of primary fixed at `baselineOrbit` (overrides variance); exclusion-zone widening (`nextSpread += zoneWidth` when next slot would land in an excluded interval). Zed primary (Aab) sequence: `1.0, 1.6, 2.1, 2.7, 3.1, 3.5, 4.1, 4.6, 7.2, 7.8, 8.3` reproduces the book's narrated 11-slot walkthrough. Zed B (1 slot before Step 4 empty): `0.52 → 1.0`. Zed Cab (5 slots): `1.4, 1.8, 2.3, 2.9, 3.3`.

- **`anomalous_test.go`:** Anomalous count table; Anomalous Type table (each band: 7- Random, 8 Eccentric, 9 Inclined, 10–11 Retrograde, 12 Trojan); random-orbit clamp to `[MAO, 20.0]` with retry. Zed: 1 anomalous (roll 10), retrograde (roll 10), parent group D3=1 (Aab), `2D-2 = 5` + `d10/10 = 0.2` → orbit 5.2R, marked with R suffix in slot id.

- **`placement_test.go`:** Empty → GG → belt → terrestrial order; 1D:1D collision (+1 to right die, then move to next slot id, then wrap to first empty); prefix-die selection by total slots (≤6 → 1D, 7–12 → D2, 13–18 → D3, >18 → 1D with reroll-above-N).

- **`planet_eccentricity_test.go`:** Each anomaly type applies the right DM (Random DM+2, Eccentric DM+5, Inclined DM+2, Retrograde DM+2, Trojan inherits parent). Reuses `stars.EccentricityValues`.

### Acceptance gate: `worked_examples_test.go`

`TestZed_FullPlacement` runs `GenerateSystemPlacement` with `composeZed()` and a stubbed `roller.Roller` issuing the exact dice values the book narrates across pp. 36–52. Asserts:

- `Counts` from `GenerateCounts` == `{GG: 4, Belts: 2, Terrestrials: 11, Total: 17}`; updated by `AddAnomalous` to `{4, 2, 12, 18}` after Step 7.
- Allocations: Aab=11, B=2 (post Step 4 empty), Cab=5.
- `BaselineN=5`, `BaselineOrbit=3.1`, `EmptyOrbits=1`, `SystemSpread=0.50`.
- Per-slot orbit values match the p. 52 table within ±0.05 **except** the empty slot, asserted at C1 (strict dice) with this comment: `// book p. 52 places the empty at C2 via Referee discretion; strict dice puts it at C1.`
- 18 placements total: 4 GG, 2 belts, 12 terrestrials. The retrograde slot at orbit 5.2R is in the Aab group.

Single-star integration test is dropped (Sol has no complete book-narrated roll trail; per-step tests cover the single-star path adequately).

### Tolerances

- Orbit# values: ±0.01 (book gives them to two digits).
- Spread: ±0.005 (book gives to three digits before rounding).
- Counts: exact integer match.

### Test infrastructure

- `roller.Roller` already has a stubbing pattern (used in `stars/system_test.go`). Reuse it.
- `composeZed()` from 2A's `worlds/worked_examples_test.go` is reused unchanged.

## Open questions for future sub-projects

- **Continuation Method (pre-existing mainworlds).** Step 3d, Step 8 mainworld branches, atmosphere-DM raw-temp reverse-engineering, count tweaks. Build as `worlds/continuation.go` once the clean-slate path is proven.
- **Optional Rules.** Per-star baseline number + spread (Step 5); 10–20% minimum-separation in compact systems. Add as `PlacementOpts{ PerStarBaseline: bool, MinimumSeparationPercent: float64 }` if needed.
- **Referee override hook.** Don't pre-build; expose an interface only when a real consumer needs to reproduce a book example with judgment-call modifications.
- **Hill-sphere alternate orbit method** (still deferred): `worlds.AvailableOrbitsOpts{ Method: HillSphere }`.
- **Post-stellar primary MAO** (still deferred): `ErrPostStellarPrimaryUnsupported`.
- **CLI integration.** `cmd/wbh` extension is straightforward but worth deferring until 2C lands the IISS Class II/III survey form, the natural rendering target.
- **`Other`-descriptor wart in `stars.GenerateCompanionStar`** still its own small follow-up.

## Success criteria

- All eleven new public functions (`GenerateCounts`, `AllocateOrbitsByStar`, `RollBaselineNumber`, `BaselineOrbit`, `RollEmptyOrbits`, `Spread`, `MaximumSecondarySpread`, `PlaceOrbitSlots`, `AddAnomalous`, `PlaceWorlds`, `RollPlanetEccentricities`) plus the `GenerateSystemPlacement` façade and the `Group.HZCO()` method exist in `wbh/worlds`, are exercised by per-rule unit tests, and have GoDoc comments referencing the WBH page or step they encode.
- `TestZed_FullPlacement` reproduces the book's Zed walkthrough across Steps 1–9 with the single documented C1/C2 empty-orbit divergence.
- The 2A carry-forward refactor is complete: `Group` carries `sourceCompanion`; rules 9–11 in `available_orbits.go` use it directly; the existing 2A Zed available-orbits acceptance test still passes unchanged.
- `worlds.ErrContinuationMethodUnsupported` exists and is returned when a `stars.System` carrying pre-existing mainworld data is passed in (the field doesn't exist yet, so this is a placeholder return path until a future sub-project adds the field).
- A fresh checkout of `` runs `just check && just test` clean (`gofumpt` + `golangci-lint v2.12.1` + `go test -race ./...`).
- A reader with the WBH open can match every exported symbol in the new files to a specific page or step in WBH pp. 36–52.
