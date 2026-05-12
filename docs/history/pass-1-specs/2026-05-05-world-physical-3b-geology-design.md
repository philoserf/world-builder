# World Physical 3B-Geology — Design

**Date:** 2026-05-05
**Sub-project:** 3B-geology (first sub-project of 3B; pp.125-127)
**Predecessor:** 3A2b-rederive — merged on `main` as `11f9928`
**Pre-3B cleanup:** Step 5A-5D extracted (`6a157a1`), `RollGasMix` param renamed (`af3cd00`), modernize gate added (`1fa8aa8`)

## Goal

Implement WBH pp.125-127: residual seismic stress, tidal stress factor, tidal heating factor, total seismic stress, gas-giant residual heat, the post-TSS temperature recompute, and tectonic plates. Add a single new pipeline step `runStep5E` between 3A2b-rederive and Step 6.

This is the **last sub-project with a temperature feedback edge** in the entire WBH. Per the dependency-graph analysis (`project_world_builder_3b_dependency_graph.md`), the TSS → Temperature edge converges in one pass; everything after 3B-geology is pure forward pipeline.

## Brainstorm decisions

| Q                   | Decision                                                                                                                                                        |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Q1: Scope           | (b) Standard — Residual Seismic Stress + Tidal Stress Factor + Tidal Heating Factor + TSS + post-TSS temperature recompute + GG Residual Heat + Tectonic Plates |
| Q2: Step structure  | (a) Single `runStep5E` does everything per body                                                                                                                 |
| Q3: Storage         | (a) New `Geology` struct attached as `dp.Geology *Geology`                                                                                                      |
| Q4: 5D re-trigger   | (a) None. 5E is terminal; band-cross divergences logged via `t.Logf`                                                                                            |
| Q5: Acceptance test | (b) Keep `TestZed_FullDetail_3A2b` name; append new assertions                                                                                                  |

## Architecture

### New file: `worlds/geology.go`

Holds the `Geology` struct + standalone helper functions:

- `ComputeResidualSeismicStress(body) int` — `(Size − Age + DMs)²` with floor and pre-square clamp; deterministic (no dice)
- `ComputeTidalStressFactor(body) int` — `Σ TidalEffects.Total / 10`, deterministic (no dice)
- `ComputeTidalHeatingFactor(body, parent) int` — primary-mass formula divided by 3000; ignore values < 1
- `ComputeGGResidualHeat(body) float64` — GG-only: `80 × ⁴√(Mass⊕) ÷ √(Age)`
- `ApplyInherentTempAddition(temp *Temperature, addedK float64)` — applies `⁴√(T⁴ + Added⁴)` to all temperature fields
- `RollTectonicPlates(r, body) int` — `Size + Hydro − 2D + DMs` with prerequisites

### New step: `runStep5E(r, detailed, sys) error` in `worlds/system_detail_steps.go`

Slots between Step 5D (3A2b-rederive) and Step 6 (BaselineN backfill). Single pass through `detailed`; per-body, per-moon. Mutates `dp.Geology *Geology` and `dp.Temperature` in place.

`DetailSystem` orchestrator picks up one new line:

```go
// Step 5E — 3B-geology pass: seismic + GG residual heat + temp recompute + tectonic plates.
if err := runStep5E(r, detailed, sys); err != nil {
    return SystemDetail{}, err
}
```

### Struct extensions

`DetailedPlacement` gets one new pointer field:

```go
Geology *Geology
```

`Moon` gets the same pointer field (moons get their own Geology computed via `buildMoonPlacementView`-style synthetic moonDP, same pattern as 3A2a/3A2b-temp/3A2b-rederive).

Accessor:

```go
func (dp *DetailedPlacement) HasGeology() bool { return dp.Geology != nil }
```

## Public API

### The `Geology` struct

```go
// Geology — seismic activity and inherent temperature contribution per
// WBH pp.125-127. Populated by Step 5E for any non-empty, non-belt body.
//
// Conditional applicability:
//   - Terrestrials: ResidualSeismicStress, TidalStressFactor,
//     TidalHeatingFactor, TotalSeismicStress, TectonicPlates populated;
//     InherentTemperatureK == float64(TotalSeismicStress).
//   - Gas giants: only InherentTemperatureK populated (from the GG
//     residual heat formula); seismic fields and TectonicPlates remain 0.
//   - Belts (Size 0): geology not generated; dp.Geology stays nil.
type Geology struct {
    // Terrestrial-only seismic factors (0 for gas giants).
    ResidualSeismicStress int  // (Size − Age + DMs)² per WBH p.125
    TidalStressFactor     int  // Σ tidal effects ÷ 10 per WBH p.126
    TidalHeatingFactor    int  // primary-mass formula ÷ 3000 per WBH p.126
    TotalSeismicStress    int  // sum of the three above

    // Terrestrial-only tectonic plate count. Zero if prerequisites failed
    // (TSS ≤ 0 or Hydro < 1) or if the dice roll produced ≤ 1.
    TectonicPlates int

    // Inherent temperature addition in Kelvin, used in the temperature
    // recompute equation: New T = ⁴√(T⁴ + InherentTemperatureK⁴).
    // For terrestrials: equals float64(TotalSeismicStress).
    // For gas giants: equals 80 × ⁴√(MassEarth) ÷ √(AgeGyr), zero if
    // the formula produces a negligible value (< 1K).
    InherentTemperatureK float64
}
```

### `runStep5E` signature

```go
func runStep5E(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error
```

Same shape as `runStep5C` and `runStep5D`. No `parent` parameter — moons are processed inside the loop using `buildMoonPlacementView`.

## Procedure (per body)

### Step 1 — Residual Seismic Stress (terrestrials only, WBH p.125)

```
ResidualSeismicStress = (Size − Age(Gyr) + DMs)²
```

DMs:

| Condition               | DM                   |
| ----------------------- | -------------------- |
| World is a moon         | +1                   |
| World has Size-1+ moons | +1 per moon, max +12 |
| Density > 1.0           | +2                   |
| Density < 0.5           | −1                   |

Edge cases:

- Round down the inner expression _before_ squaring.
- If pre-square value is < 1, treat as 0 (so the squared result is 0, not 1).

Worked: Terra (S=8, Age=4.568, 2 moons, density=1.0) → `8 − 4.568 + 2 = 5.4322 → 5 → 25`. Luna (S=2, Age=4.568, moon, density=0.6) → `2 − 4.568 + 1 = −1.5 → 0`. Zed Prime (S=5, Age=6.3, moon, density=1.03) → `5 − 6.3 + 1 + 1 = 0.7 → 0`.

### Step 2 — Tidal Stress Factor (terrestrials only, WBH p.126)

```
TidalStressFactor = floor(TidalEffects.Total / 10)
```

Deterministic (no dice). Reads the existing `dp.TidalEffects.Total` (metres) populated in Step 5B.5.

Worked: Zed Prime, total ~30m → 3.

### Step 3 — Tidal Heating Factor (terrestrials only, WBH p.126)

```
                       (PrimaryMass⊕)² × Size⁵ × eccentricity²
TidalHeatingFactor = ─────────────────────────────────────────────────
                     3000 × Distance(Mkm)⁵ × Period(days) × WorldMass⊕
```

Floor the result; ignore values < 1 (treat as 0).

Inputs:

- **Planet around star**: `PrimaryMass⊕` = star mass × 332,946 (solar→Earth masses). Distance = `OrbitToAU(dp.Orbit) × 149.6`. Period = `dp.Period.Hours / 24`.
- **Moon around planet**: `PrimaryMass⊕` = parent planet's mass in Earth masses. Distance = `m.OrbitKm / 1_000_000`. Period = `m.PeriodHours / 24`.
- World Size = numeric Size (0-15). World Mass = body's mass in Earth masses.
- Eccentricity = body's eccentricity.

Worked: Zed Prime → 14. Io ≈ 101 (book reference). Enceladus ≈ 11 (book reference).

### Step 4 — Total Seismic Stress (terrestrials only)

```
TotalSeismicStress = ResidualSeismicStress + TidalStressFactor + TidalHeatingFactor
```

Pure addition.

### Step 5 — Gas Giant Residual Heat (gas giants only, WBH p.125)

```
InherentTemperatureK = 80 × ⁴√(Mass⊕) / √(Age in Gyr)
```

If the result is < 1K, set to 0 (negligible).

Worked: Zed Prime's GG (Mass⊕ = 1200, Age = 6.336) → `80 × ⁴√1200 / √6.336 = 80 × 5.886 / 2.517 ≈ 187K`.

### Step 6 — Inherent Temperature Recompute (all bodies, WBH p.125)

For terrestrials: `Geology.InherentTemperatureK = float64(Geology.TotalSeismicStress)`.
For gas giants: `Geology.InherentTemperatureK =` Step 5 output.

Apply to each populated temperature field on `dp.Temperature`:

```
NewT = ⁴√(OldT⁴ + InherentTemperatureK⁴)
```

Fields touched: `MeanK`, `HighK`, `LowK`, `WorstHighK`, `WorstLowK`, plus all populated scenario fields (TwilightK, etc.). The recompute is idempotent in shape — same equation, every field.

Acceptance test gates a `t.Logf` informational note when `MeanKToTempRange(preTSS_T) != MeanKToTempRange(postTSS_T)` for any body — surfaces the rare cold-rogue-world case for future attention.

### Step 7 — Tectonic Plates (terrestrials only, WBH p.127)

Prerequisites: `TotalSeismicStress > 0` AND `Hydrographics.Code ≥ 1`. If either fails: `TectonicPlates = 0` (no roll consumed).

```
TectonicPlates = Size + Hydrographics − 2D + DMs
```

DMs:

| Condition              | DM  |
| ---------------------- | --- |
| TSS between 10 and 100 | +1  |
| TSS > 100              | +2  |

If result ≤ 1: `TectonicPlates = 0` (no tectonic activity).

Worked: Zed Prime (S=5, Hydro=6, TSS=17 → DM+1, 2D=8) → `5 + 6 − 8 + 1 = 4`.

### Order of operations within `runStep5E`

For each non-empty placement:

1. If terrestrial (and not a belt):
   1. Compute ResidualSeismicStress (deterministic — no dice)
   2. Compute TidalStressFactor
   3. Compute TidalHeatingFactor
   4. Sum → TotalSeismicStress
   5. Set InherentTemperatureK = float64(TotalSeismicStress)
2. If gas giant:
   1. Compute GG Residual Heat → InherentTemperatureK
3. Apply temperature recompute via `ApplyInherentTempAddition` (mutates `dp.Temperature` fields in place)
4. If terrestrial: Roll TectonicPlates if prerequisites met (consumes 1 × 2D)
5. Recurse into moons: build moonDP via `buildMoonPlacementView`, repeat steps 1-4 for each moon, applying the same temperature recompute to `m.Temperature`.

Per-body dice budget: **1 × 2D per terrestrial that passes plate prerequisites** (Hydro ≥ 1, TSS > 0). Otherwise zero. Moons add the same conditional 2D each.

## Sub-decisions

- **Mutation policy** — In-place mutation of `dp.Temperature` fields (consistent with 3A2b-rederive). No audit flag. Tests verify pre vs post values directly.
- **Tidal heating single-source** — Per the formula's shape on p.126. Parent planet for moons, primary star for planets. No summation across multiple gravitational sources.
- **GG residual heat applies to GG bodies only** — Moons of GGs do NOT inherit the heat (deferred per book's "unlikely to matter" framing). Documented in the spec; revisit if biology/habitability surfaces a need.
- **Belts (Size 0) get no Geology** — Skip in the loop. `dp.Geology` stays nil.
- **Empty placements get no Geology** — Skip.
- **Band-cross divergence handling** — `MeanKToTempRange(preTSS_T) != MeanKToTempRange(postTSS_T)` is logged as a `t.Logf` note in the acceptance test. No 5D re-trigger. If real distributions show frequent crosses, escalate to a future "5F atm-rederive-after-TSS" sub-project.
- **Empty `dp.TidalEffects`** — If a body has no TidalEffects (e.g., temp test fixtures), TidalStressFactor = 0. Defensive but not paranoid.
- **Negative Age** — Defensive: if Age < 0 (shouldn't happen), pre-square clamp protects ResidualSeismicStress from blowing up.

## Testing

### Per-formula unit tests (`worlds/geology_test.go`)

- `TestComputeResidualSeismicStress_Terra` — S=8, Age=4.568, 2 moons, density=1.0 → 25
- `TestComputeResidualSeismicStress_Luna` — S=2, Age=4.568, moon, density=0.6 → 0
- `TestComputeResidualSeismicStress_ZedPrime` — S=5, Age=6.3, moon, density=1.03 → 0
- `TestComputeResidualSeismicStress_PreSquareClampLessThanOne` — verifies "< 1 → 0" path
- `TestComputeResidualSeismicStress_DensityMaxMoonDM` — verifies +12 cap on per-moon DM
- `TestComputeTidalStressFactor_ZedPrime` — TidalEffects.Total = 30m → 3
- `TestComputeTidalStressFactor_FloorRounding` — verifies floor semantics
- `TestComputeTidalStressFactor_NilTidalEffects_Zero` — defensive
- `TestComputeTidalHeatingFactor_ZedPrime` — full inputs → 14
- `TestComputeTidalHeatingFactor_LessThanOne_ZeroOut` — small values → 0
- `TestComputeTidalHeatingFactor_PlanetVsMoon_UnitsConvert` — verifies AU↔Mkm and years↔days conversions for the planet path
- `TestComputeGGResidualHeat_ZedPrimeGG` — Mass⊕=1200, Age=6.336 → 187 (within ±1)
- `TestComputeGGResidualHeat_OldOrLowMass_Zero` — formula < 1K → 0
- `TestApplyInherentTempAddition_AllFieldsTouched` — every populated temp field gets the equation applied
- `TestApplyInherentTempAddition_ZedPrime_Negligible` — 300K + 17K addition stays at 300K (rounded)
- `TestApplyInherentTempAddition_RogueWorld_NotNegligible` — 25K + 100K addition → ~100K
- `TestRollTectonicPlates_ZedPrime` — S=5, Hydro=6, TSS=17, 2D=8 → 4
- `TestRollTectonicPlates_TSSZero_NoActivity` — prerequisite fails, returns 0 without consuming dice
- `TestRollTectonicPlates_HydroZero_NoActivity` — same
- `TestRollTectonicPlates_ResultLessThanOrEqualOne_NoActivity` — verifies "result ≤ 1 → 0" path

### Orchestrator tests

- `TestRunStep5E_Terrestrial_PopulatesGeology` — body in, geology out, all fields set
- `TestRunStep5E_GasGiant_OnlyInherentHeat` — GG body, Geology populated with InherentTemperatureK, seismic fields stay 0
- `TestRunStep5E_BodyEmpty_NoOp` — skipped
- `TestRunStep5E_BeltSize0_NoGeology` — skipped (no geology for belts)
- `TestRunStep5E_MoonRecursion` — moons get their own Geology
- `TestRunStep5E_BandCrossLogged` — synthetic cold-world setup that crosses a band, verify the divergence is detected (test inspects log via direct helper call)

### Acceptance test extension (`worlds/worked_examples_test.go`)

Append new assertions to `TestZed_FullDetail_3A2b` (keep the name per Q5-b):

- Geology populated for terrestrials and GGs (assertion: `dp.HasGeology()` for all non-empty, non-belt bodies)
- ResidualSeismicStress: 0 (Zed Prime); 25 if Terra is in fixture
- TidalStressFactor = 3 (Zed Prime)
- TidalHeatingFactor = 14 (Zed Prime)
- TotalSeismicStress = 17 (Zed Prime)
- TectonicPlates: between 0 and 25 (theoretical max: `S=15 + H=10 + DM+2 − 2D=2 = 25`; realistic Terra-like worlds top out around 18). Range bound only — not pinning the exact dice-driven roll across all 100 iterations.
- For Zed Prime's GG: InherentTemperatureK ≈ 187 (±1)
- For all bodies: Temperature.MeanK present and finite after recompute
- Informational `t.Logf` on band-cross count across iterations

### Dice budget

Per body: 0 dice for the formulas. +1 × 2D per terrestrial that passes plate prerequisites (Hydro ≥ 1, TSS > 0). Roughly +N_terrestrials_with_hydro_and_tss dice per iteration. The existing 100-iter scripted-dice budget will need a small bump — measure the actual per-iteration consumption during the implementation.

## Carry-forwards (deferred to future sub-projects)

| What                                                                  | Why                                                                                                                             |
| --------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| GG residual heat propagation to moons (luminosity formula)            | Book says "unlikely to matter" (adds 0.0055K to Zed Prime). Revisit only if biology/habitability needs it.                      |
| Reflected light from parent body or large moon                        | Book p.125: "in most situations are too minor to be worth the calculation"                                                      |
| 5D re-trigger when temperature band crosses post-TSS                  | Not needed for HZ worlds; rare for cold rogues. Surfaced via `t.Logf`; address only if real distributions show frequent crosses |
| Earthquake/volcano frequency narrative labels (TSS thresholds → text) | UI-layer concern; address when Class IV-P form rendering is built (3B-final)                                                    |
| Tectonic-plate-driven world map detail (WBH p.127 cross-reference)    | Mapping is a separate sub-project (3B-maps); out of scope                                                                       |

## File map

| File                                                        | Status   | Purpose                                                |
| ----------------------------------------------------------- | -------- | ------------------------------------------------------ |
| `worlds/geology.go`                                         | New      | `Geology` struct + standalone helpers                  |
| `worlds/geology_test.go`                                    | New      | Per-formula unit tests                                 |
| `worlds/system_detail_steps.go`                             | Modified | Add `runStep5E`                                        |
| `worlds/system_detail.go`                                   | Modified | One new line: call `runStep5E` between 5D and Step 6   |
| `worlds/placement.go` (or wherever DetailedPlacement lives) | Modified | Add `Geology *Geology` field + `HasGeology()` accessor |
| `worlds/moon_refinement.go` (or wherever Moon lives)        | Modified | Add `Geology *Geology` field on `Moon`                 |
| `worlds/worked_examples_test.go`                            | Modified | Append new assertions to `TestZed_FullDetail_3A2b`     |

## Plan target

Target: 8-10 tasks, executable via subagent-driven-development with per-task subagent + spec/code review loops. Final end-to-end review on Opus before merge. Same workflow as 3A2b-temp and 3A2b-rederive.
