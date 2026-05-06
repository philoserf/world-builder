# World Physical Characteristics — Sub-project 3A2b-rederive Design (Atmosphere / Hydrographics Re-derivation under Real Temperature)

**Date:** 2026-05-04
**Status:** approved through brainstorming; pending user review of written spec
**Source material:** Mongoose Publishing, _World Builder's Handbook_ (Geir Lanesskog, 2023). PDF at `Mongoose/Core Rules/World Builders Handbook.pdf`.

## Context

Builds on 3A2b-temp (merged on `main` as `0e4e73b`). 3A2b-temp populated `body.Temperature.MeanK` (real surface temperature) for every non-empty body. 3A2b-rederive uses that to re-derive 3A1's temperature-sensitive Atmosphere and Hydrographics fields, plus add the deferred chemical formula tails and the Optional Runaway Greenhouse check.

This is the second half of the 3A2b decomposition (3A2b-temp + 3A2b-rederive). After this sub-project lands, **3A2b is complete** and 3A1's "provisional under HZCO temperature" qualifier is finally retired.

**WBH source pages:**

- p.79 — Optional Runaway Greenhouse rule (atm 2-F, mean T > 303K, 2D + DMs ≥ 12 → atm becomes A/B/C, hydro re-rolled with DM-6)
- p.81 — Scale Height formula `≈ T_K / (M̄·g)` (project approximation: `8.5 × T/288 / g`)
- pp.94-98 — Hot/Cold/Frozen Atmospheres tables + Gas Mix tables (already implemented in 3A1 with provisional TempRange; re-derive under real TempRange)
- p.99 — Atmosphere profile composition tail format `<Code>-<Subtype>:<Gas>-<Pct>:<Gas>-<Pct>:...`
- p.102 — Possible Exotic Liquids table (15 molecules with melting/boiling points and Relative Abundance)

## Brainstorming decisions (2026-05-04)

| Question                        | Decision                                                                  | Rationale                                                                                                                                                 |
| ------------------------------- | ------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Q1: Iteration policy            | (b) Fixed 2-pass iteration                                                | Captures dominant feedback (greenhouse responds to corrected pressure) without convergence games. Most physical worlds settle in one round of correction. |
| Q2: Mutation policy             | (a) In-place mutation, no audit flag                                      | 3A1 spec already commits to post-rederive values being canonical ("provisional under HZCO temperature"). Renderers want a single value.                   |
| Q3: Optional Runaway Greenhouse | (a) Always check                                                          | Pattern consistency with prior "always compute, never punt to Referee" choices. Procedural rule with defined trigger and effect.                          |
| Q4: Exotic liquid selection     | (a) Deterministic — highest Relative Abundance wins                       | Reproducible from MeanK + atm code. Saves dice budget. Ties broken by lower BoilingK.                                                                     |
| Q5: Tidal-lock re-eval          | (b) Skip entirely                                                         | Original tidal-lock dice not captured; same-dice re-eval requires dice-capture infrastructure. Acknowledge limitation in spec.                            |
| Q6: Acceptance gate             | (a) Replace `TestZed_FullDetail_3A2b_temp` with `TestZed_FullDetail_3A2b` | Matches the 3A1 → 3A2a → 3A2b-temp pattern. Once this sub-project lands, "3A2b" is complete.                                                              |

## Architecture

Stay flat in `worlds/`. Three new production files, one new test file:

| File                                  | Status  | Responsibility                                                                                                                                    |
| ------------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `worlds/temperature_rederive.go`      | **new** | `RederiveAtmosphereHydrographics` orchestrator; `MeanKToTempRange` helper; private helpers for pressure/scale-height/profile updates              |
| `worlds/runaway_greenhouse.go`        | **new** | `CheckRunawayGreenhouse` per WBH p.79; trigger detection + atm code mutation + DM rolls                                                           |
| `worlds/exotic_liquids.go`            | **new** | `PossibleExoticLiquids` table (p.102) + `SelectExoticLiquid` deterministic picker                                                                 |
| `worlds/temperature_rederive_test.go` | **new** | All 3A2b-rederive unit tests                                                                                                                      |
| `worlds/system_detail.go`             | modify  | Wire **Step 5D** (2-pass iteration) after Step 5C                                                                                                 |
| `worlds/atmosphere.go`                | modify  | Add gas-mix Profile derivation helper if 3A1 doesn't already expose it; update `Atmosphere` doc comment to remove "provisional" qualifier post-5D |
| `worlds/hydrographics.go`             | modify  | Add `Hydrographics.Profile string` field for liquid composition tail; update doc comment                                                          |
| `worlds/body_physical.go`             | modify  | Update doc comment to remove "provisional" qualifier post-5D                                                                                      |
| `worlds/worked_examples_test.go`      | modify  | Replace `TestZed_FullDetail_3A2b_temp` with `TestZed_FullDetail_3A2b`                                                                             |

**Branch:** `feat/wbh-world-physical-3a2b-rederive` — created off `main` at the merge of 3A2b-temp (`0e4e73b`).

## Public API

### Top-level orchestrator

```go
// In worlds/temperature_rederive.go:

// RederiveAtmosphereHydrographics re-derives 3A1's temperature-sensitive
// Atmosphere/Hydrographics fields under the body's current Temperature.MeanK.
// Mutates body in place. Called twice as part of Step 5D's 2-pass iteration.
//
// Mutates:
//   - Atmosphere.Pressure       (re-rolled within current Subtype range)
//   - Atmosphere.ScaleHeight    (≈ 8.5 × T_K / 288 / gravityG)
//   - Atmosphere.Subtype        (re-derived from current TempRange for variable codes A/B/C/F)
//   - Atmosphere.Code           (only for NON-HZ bodies via Hot/Cold tables)
//   - Atmosphere.Profile        (gas composition tail per p.99 format)
//   - Hydrographics.Code        (re-rolled with current Hot/Boiling DMs)
//   - Hydrographics.Profile     (liquid composition tail; new field)
// May trigger CheckRunawayGreenhouse mutation per p.79.
//
// No-op when body.Body == BodyEmpty or body.Temperature == nil.
func RederiveAtmosphereHydrographics(
    r roller.Roller,
    body *DetailedPlacement,
    sys stars.System,
    parent *DetailedPlacement,
) error

// MeanKToTempRange buckets a real mean temperature in Kelvin into the same
// TempRange bands 3A1 used for HZCO offset bucketing:
//   ≥ 453K → Boiling
//   353-453K → Hot
//   273-353K → Temperate
//   123-273K → Cold
//   < 123K → Frozen
func MeanKToTempRange(meanK float64) TempRange
```

### Runaway greenhouse

```go
// In worlds/runaway_greenhouse.go:

// CheckRunawayGreenhouse evaluates and applies WBH p.79 Optional Runaway
// Greenhouse for HZ bodies with atm code in {2-9, D, E} (i.e., 2-14
// excluding A/B/C/F at codes 10/11/12/15) and mean T > 303K.
//
// DMs to 2D roll:
//   - +1 per System Age Gyr (round up)
//   - +4 if mean T ≥ 388K (boiling temperature, 12+ on basic table)
//
// On 12+: trigger fires. Mutates atm.Code via 1D table:
//   1 → A (10), 2-4 → B (11), 5+ → C (12)
//   - DM-2 if Size 2-5
//   - DM+1 if originally tainted (codes 2, 4, 7, 9)
// Mutates atm.Subtype with DM+4 to subtype roll.
//
// Returns true iff trigger fired. Caller re-rolls Hydrographics with DM-6
// (boiling) instead of DM-2 (hot) when this returns true.
//
// Non-HZ bodies skip per book (rule explicitly HZ-only).
//
// MVP simplification: bodies that already have atm A/B/C/F+ skip this check.
// The book's "consider boiling" case for those codes (only flips hydrographics
// DM-6 without mutating atm code) is deferred — see Carry-forwards.
func CheckRunawayGreenhouse(r roller.Roller, body *DetailedPlacement, sys stars.System) bool
```

### Exotic liquids

```go
// In worlds/exotic_liquids.go:

// ExoticLiquid is one row of the WBH p.102 Possible Exotic Liquids table.
type ExoticLiquid struct {
    Code      string  // "H2O", "CH4", "NH3", etc.
    MeltingK  float64
    BoilingK  float64
    Abundance int     // Relative Abundance, 1..100
}

// PossibleExoticLiquids — WBH p.102 table, 15 entries ordered by boiling point.
var PossibleExoticLiquids = []ExoticLiquid{
    {"F2", 53, 85, 2},
    {"O2", 54, 90, 50},
    {"CH4", 91, 113, 70},
    {"C2H6", 90, 184, 70},
    {"Cl2", 171, 239, 1},
    {"NH3", 195, 240, 30},
    {"SO2", 201, 263, 20},
    {"HF", 190, 293, 2},
    {"HCN", 260, 299, 30},
    {"HCl", 247, 321, 1},
    {"H2O", 273, 373, 100},
    {"CH2O2", 281, 374, 15},
    {"CH3NO", 275, 483, 15},
    {"H2CO3", 193, 607, 20},
    {"H2SO4", 388, 718, 20},
}

// SelectExoticLiquid returns the dominant liquid for a body with exotic
// atmosphere (Atm A-C/F: codes 10, 11, 12, 15) and non-zero hydrographics
// at the given mean temperature.
//
// Deterministic: among molecules where MeltingK ≤ meanK ≤ BoilingK, returns
// the highest-Abundance candidate. Ties broken by lower BoilingK (more
// "stable" in range). Returns "" if no candidate fits or atmCode is not exotic.
func SelectExoticLiquid(meanK float64, atmCode int) string
```

### Hydrographics.Profile (new field)

Add to existing `Hydrographics` struct in `worlds/hydrographics.go`:

```go
type Hydrographics struct {
    Code int
    // ... existing 3A1 fields
    Profile string  // 3A2b-rederive: composition tail like "H6:H2O-100" or "H4:CH4-100" (empty for vacuum/no-hydro)
}
```

### Atmosphere.Profile (verify existence)

3A1's `Atmosphere` struct has `Profile *AtmosphereProfile`. Verify that 3A2b-rederive can populate `Profile` via an existing helper, or add `DeriveAtmosphereProfile(code, subtype int, tempRange TempRange) *AtmosphereProfile` if not exposed. The Profile string format per p.99: `<Code>-<Subtype>:<Gas>-<Pct>:<Gas>-<Pct>:...` (e.g., `B-Std:N2-78:O2-21:Ar-01`).

## Procedure (Steps Mapped to Functions)

### Pass 1 / Pass 2 (RederiveAtmosphereHydrographics)

Both passes execute the same procedure; only the input `body.Temperature.MeanK` differs.

| Step                                                | Computation                                                                                                                                                   | Notes                                                                                          |
| --------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| 1. Determine current TempRange                      | `MeanKToTempRange(body.Temperature.MeanK)`                                                                                                                    | New helper using same bands as 3A1                                                             |
| 2. Re-derive `Atmosphere.ScaleHeight`               | `8.5 × meanK/288 / gravityG`                                                                                                                                  | Reuses 3A1's `DeriveScaleHeight` if signature accepts meanK; else add `DeriveScaleHeightFromK` |
| 3. Re-derive `Atmosphere.Code` (non-HZ only)        | If body is non-HZ AND new TempRange differs from prior, re-roll `RollAtmoCodeNonHZ(r, sizeCode, newTempRange)`                                                | HZ atm code is temperature-blind; left as-is. 1 dice.                                          |
| 4. Re-derive `Atmosphere.Subtype`                   | For variable-subtype atm (codes 10-12, 15), re-roll subtype with new TempRange                                                                                | Existing 3A1 helper. 1 dice if re-rolled.                                                      |
| 5. Re-derive `Atmosphere.Pressure`                  | If Subtype changed, `RollTotalPressure(r, atmCode)` re-rolls within new MinPressureRange/Span                                                                 | 2 dice (1D + 1D) if re-rolled.                                                                 |
| 6. Re-derive `Atmosphere.Profile` (gas composition) | Format `<Code>-<Subtype>:<Gas>-<Pct>:...` per p.99. Use existing 3A1 helper if available; else add `DeriveAtmosphereProfile`.                                 | Possibly new helper.                                                                           |
| 7. CheckRunawayGreenhouse                           | HZ atm 2-F + meanK > 303K + 2D + DMs ≥ 12 → mutate atm to A/B/C                                                                                               | 2D + 1D + atm DM rolls (~4 dice if triggered)                                                  |
| 8. Re-derive `Hydrographics.Code`                   | `RollHydroDigit(r, atmCode, atmSubtype, sizeCode, newTempRange)` with current TempRange. **If runaway fired in step 7, override Hot DM-2 with Boiling DM-6.** | 1 dice                                                                                         |
| 9. Generate `Hydrographics.Profile`                 | Atm 2-9/D/E + hydro > 0 → `H<code>:H2O-100`. Atm A-C/F + hydro > 0 → `H<code>:<SelectExoticLiquid>-100`. Empty otherwise.                                     | Deterministic                                                                                  |

### Step 5D orchestration

For each body and each moon (`buildMoonPlacementView` per 3A2a pattern):

```
RederiveAtmosphereHydrographics(r, body, sys, parent)         // Pass 1
body.Temperature = GenerateTemperature(r, body, sys, parent)  // Re-run with corrected atm/hydro
RederiveAtmosphereHydrographics(r, body, sys, parent)         // Pass 2 — final
```

The second `GenerateTemperature` may produce a slightly different MeanK from the first (greenhouse responds to corrected pressure). Pass 2's rederive uses that final MeanK. We do NOT re-run Temperature after Pass 2 — that's the 2-pass cap from Q1-B.

### Dice budget per body

Approximate worst-case dice consumption for Step 5D per body:

- Pass 1 rederive: up to 5 dice (atm code re-roll + subtype + 2 pressure + hydro)
- Mid-pass GenerateTemperature: ~5 dice (full 3A2b-temp budget)
- Pass 2 rederive: up to 5 dice
- Plus runaway greenhouse (if triggered): ~4 additional dice

**Total ~15 dice per body** with atmosphere and hydrographics; ~10 for vacuum bodies. Acceptance test scripts must accommodate.

### Pipeline order rationale

Step 5D runs after 5C (which produces initial `Temperature`). Within 5D, planets are processed before their moons (matches 5B/5C pattern), so `parent.Temperature` is available for the mid-pass moon `GenerateTemperature` call (multi-source addition reads `parent.Temperature.MeanK`).

### What's NOT re-derived

- **Tidal lock state** (per Q5-B): skip entirely. Documented limitation.
- **Surface distribution** (5B.1): dice consumed already; rederive shifts in Hydrographics.Code aren't physically meaningful enough to re-roll for.
- **Day length, Axial tilt** (5B.2/5B.3): independent of atmosphere/hydrographics.
- **Tidal effects** (5B.5): depend on atm pressure only via the > 2.5 bar tidal-lock DM, which we're not re-evaluating.

## Sub-decisions

| Item                                | Decision                                                                                                                                                                                                              |
| ----------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| MeanKToTempRange thresholds         | Use 3A1's bands (453K, 353K, 273K, 123K) for consistency.                                                                                                                                                             |
| HZ vs non-HZ atm code re-derivation | Only non-HZ bodies' atm code re-rolls. HZ atm codes are temperature-blind per Core Rulebook table. Runaway greenhouse can still mutate any HZ atm 2-F.                                                                |
| Surface distribution NOT re-derived | Dice consumed in 5B.1; Hydrographics.Code shifts of ±1 don't motivate a re-roll.                                                                                                                                      |
| Atm.Profile format                  | `<Code>-<Subtype>:<Gas>-<Pct>:<Gas>-<Pct>:...` per p.99. Verify 3A1 exposes a gas-mix derivation helper; add if missing.                                                                                              |
| Hydrographics.Profile format        | Single-liquid form `H<code>:<liquid>-100`. Multi-component composition deferred.                                                                                                                                      |
| Pass 2 dice budget                  | ~15 dice/body documented. Acceptance test scripts accommodate.                                                                                                                                                        |
| Moon-side mutation propagation      | Verify `buildMoonPlacementView` aliases atm/hydro pointers. If pointer-aliased, mutations propagate automatically. If value-copied, explicit write-back required (same lesson as 3A2a's `m.Eccentricity`).            |
| Convergence cap                     | 2-pass bounded; no infinite loop. Pathological cases where pass 3 would differ from pass 2 are accepted; documented.                                                                                                  |
| Doc-comment cleanup on 3A1 types    | Update `Atmosphere`, `Hydrographics`, `BodyPhysical` doc comments to replace "provisional under HZCO temperature" with "provisional until Step 5D runs; post-5D values are final." Targeted edits in the wiring task. |
| No HasRederive accessor or flag     | Per Q2-A. Pre-5D readers see provisional values; post-5D readers see final values. Documented hazard, no runtime guard.                                                                                               |

## Testing Strategy

### Per-file unit tests (all in `worlds/temperature_rederive_test.go`)

**Component tests (MeanKToTempRange, SelectExoticLiquid):**

| Test                                             | Pin                                                            |
| ------------------------------------------------ | -------------------------------------------------------------- |
| `TestMeanKToTempRange_Boundaries`                | 100K→Frozen, 200K→Cold, 288K→Temperate, 400K→Hot, 500K→Boiling |
| `TestSelectExoticLiquid_Water_Terra`             | meanK=288, atm=10 → "H2O"                                      |
| `TestSelectExoticLiquid_Methane_Cold`            | meanK=100, atm=11 → "CH4"                                      |
| `TestSelectExoticLiquid_Ethane_NotMethaneAtCold` | meanK=150, atm=11 → "C2H6" (CH4 boils at 113)                  |
| `TestSelectExoticLiquid_NoCandidate_TooHot`      | meanK=2000, atm=10 → ""                                        |
| `TestSelectExoticLiquid_TieBreakLowerBoiling`    | Tie scenario verifies lower-BoilingK wins                      |
| `TestSelectExoticLiquid_NonExoticAtm_Empty`      | atm=6 → "" (defensive)                                         |

**Runaway greenhouse tests:**

| Test                                            | Pin                                           |
| ----------------------------------------------- | --------------------------------------------- |
| `TestCheckRunawayGreenhouse_BelowTempThreshold` | meanK=300 → false                             |
| `TestCheckRunawayGreenhouse_LowAtmCode`         | atm=1 → false (out of [2, F-1])               |
| `TestCheckRunawayGreenhouse_LowDiceRoll`        | scripted 2D=2, sysAge=1 → false               |
| `TestCheckRunawayGreenhouse_Triggered_AtmA`     | scripted 2D=12, 1D=1, sysAge=5 → true, atm=10 |
| `TestCheckRunawayGreenhouse_Triggered_AtmB`     | scripted 2D=12, 1D=3 → true, atm=11           |
| `TestCheckRunawayGreenhouse_Triggered_AtmC`     | scripted 2D=12, 1D=6 → true, atm=12           |
| `TestCheckRunawayGreenhouse_TaintedDM`          | atm=7 (tainted) → +1 DM applied               |
| `TestCheckRunawayGreenhouse_SizeDM`             | size=4 → -2 DM applied                        |

**Orchestrator tests:**

| Test                                              | Pin                                                                      |
| ------------------------------------------------- | ------------------------------------------------------------------------ |
| `TestRederive_TerraLike_StableInTemperate`        | Terra-like body; rederive doesn't change atm code or hyd by more than ±1 |
| `TestRederive_ScaleHeight_TemperatureScaling`     | Same body, two meanK values; ScaleHeight scales proportionally with T    |
| `TestRederive_NonHZBody_AtmCodeRerolled`          | Non-HZ body crossing TempRange band; atm code re-rolled                  |
| `TestRederive_RunawayPath_AtmAndHydroMutate`      | Force runaway; atm ∈ {A,B,C} AND hydro re-rolled with DM-6               |
| `TestRederive_HydrographicsProfile_Water`         | atm=6, hyd=6, meanK=288 → "H6:H2O-100"                                   |
| `TestRederive_HydrographicsProfile_ExoticMethane` | atm=10, hyd=4, meanK=100 → "H4:CH4-100"                                  |
| `TestRederive_HydrographicsProfile_VacuumEmpty`   | atm=0, hyd=0 → ""                                                        |
| `TestRederive_AtmosphereProfile_TemperateOxygen`  | atm=6, subtype=Std, meanK=288 → Profile contains "N2-" prefix            |

**Pipeline-level tests:**

| Test                                 | Pin                                                               |
| ------------------------------------ | ----------------------------------------------------------------- |
| `TestStep5D_ZedPrime_2PassConverges` | Zed Prime through 5C + 5D; final MeanK within ±5K of pre-5D MeanK |
| `TestStep5D_NilTemperature_Skipped`  | Body with nil Temperature is skipped without error                |

### Composite acceptance gate

**Replace `TestZed_FullDetail_3A2b_temp` with `TestZed_FullDetail_3A2b`.** Inherits all 16 prior assertions, adds:

| #   | Assertion                                                                                                                  |
| --- | -------------------------------------------------------------------------------------------------------------------------- |
| 17  | Every body with `Atmosphere != nil && MeanK > 0` has `Atmosphere.ScaleHeight > 0` AND ≈ `8.5 × MeanK/288 / gravityG ± 20%` |
| 18  | Every body with `Atmosphere.Code > 0 && Hydrographics.Code > 0` has non-empty `Hydrographics.Profile`                      |
| 19  | Hydrographics.Profile format matches `H[0-9A]:[A-Za-z0-9]+-[0-9]+` (regex sanity)                                          |
| 20  | For Atm ∈ {10, 11, 12, 15} with hydro > 0: Profile contains a liquid from `PossibleExoticLiquids`                          |
| 21  | Pressure sanity: every body's `Atmosphere.Pressure ≥ 0`; non-gas-giants < 100 bar                                          |
| 22  | After 5D, every body's `Temperature` is non-nil (5D's mid-pass GenerateTemperature must not have produced nil)             |
| 23  | (Informational t.Logf) Count of runaway-greenhouse-fired bodies across 100 iterations                                      |
| 24  | (Informational t.Logf) Count of bodies whose atm.Code mutated from non-HZ band re-roll                                     |

**Explicitly NOT asserted:**

- "Tidal lock state correct under post-rederive pressure" — per Q5-B, deferred indefinitely
- "Pre-rederive vs post-rederive comparison" — would need instrumentation; tests can't observe pre-rederive state

100-iteration free-dice loop with seeds 0..99. Pattern unchanged.

## Carry-forwards beyond 3A2b-rederive

After 3A2b-rederive merges, **3A2b is complete** and 3A1's "provisional" qualifier is fully retired. Open carry-forwards moving to **3B (geology / biosphere)** or beyond:

- **Tidal heating contribution** (WBH p.127). Confirmed in 3A2a/3A2b-temp final reviews to live in 3B's seismology/tectonics chapter, not 3A2b.
- **Tidal lock re-eval if pressure crosses 2.5 bar.** Documented limitation; requires dice-capture infrastructure to fix correctly. Defer indefinitely or revisit in 3B.
- **Sky magnitude / apparent magnitude** (p.119). Descriptive geometry; defer.
- **Twilight Zone Variability Factors** (terrain, libration, refraction, pp.121-122). Refinement on top of basic twilight scenario; defer.
- **Full altitude treatment** (lapse rate, density gradients). 3A2b-temp implements only principal greenhouse-with-altitude effect.
- **Multi-component Hydrographics composition** (e.g., `H6:H2O-80:NH3-20`). 3A2b-rederive does single-dominant-liquid only.
- **Runaway Greenhouse "consider boiling" case for atm A/B/C/F+** (p.79 footnote). MVP skips these atm codes from the runaway check; book's effect is "consider boiling" → flip hydrographics DM-6 instead of DM-2 without mutating atm code. Defer to a later sub-project if a renderer needs hydro-percentage correction for these specific atmospheres.
- **Three scenario method spec gaps** (from 3A2b-temp final review):
  1. `MeanByLatitude` arctic clamp (fixed in 3A2b-temp follow-up; verify post-merge)
  2. `MeanBySeason` ignores latitude — composing with the latitude zone formula deferred
  3. Twilight scenario methods always return `TwilightK` — hemisphere-aware selection deferred

## Branch and merge

Branch: `feat/wbh-world-physical-3a2b-rederive` off `main` at `0e4e73b` (3A2b-temp merge).

After all tasks complete:

```bash
just check && just test
git log --oneline main..HEAD
```

Expected: ~12-14 commits ahead of main, all `ok` from test, `0 issues.` from check.

Merge (after user approval):

```bash
git checkout main
git merge --no-ff feat/wbh-world-physical-3a2b-rederive -m "Merge feat/wbh-world-physical-3a2b-rederive: World Physical 3A2b complete"
```

After merge, update `MEMORY.md` Subprojects line: 3A2b complete; next is 3B (geology / biosphere).
