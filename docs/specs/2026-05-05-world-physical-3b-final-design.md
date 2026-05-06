# World Physical 3B-Final — Design

**Date:** 2026-05-05
**Sub-project:** 3B-final (third sub-project of 3B; pp.132-146)
**Predecessor:** 3B-biology — merged on `main` as `b89d09b`
**Pre-3B-final cleanup:** `runStep5B` and `runStep5F` extracted (`3948b82`)

## Goal

Implement WBH pp.132-146: Habitability rating, Final Mainworld Determination, and IISS Class IV-P Survey Form rendering. Add a single new pipeline step `runStep5G` between 3B-biology and Step 6, plus a system-wide mainworld pick after 5G, plus a per-body Class IV-P form renderer.

This is the **last "physical world" sub-project** in WBH — after this merge, the next major chapter is World Social Characteristics (pp.147-234), a different domain.

## Brainstorm decisions

| Q                           | Decision                                                                                                                                             |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| Q1: Scope                   | (a) Single sub-project covering Habitability + Mainworld + Class IV-P form                                                                           |
| Q2: Storage                 | (b) `Habitability` struct (`Rating int; Notes string`) attached as `dp.Habitability *Habitability`                                                   |
| Q3: Gravity DM bands        | (a) Narrower band wins (matches Zed Prime worked example); document book inconsistency inline                                                        |
| Q4: Mainworld determination | (a) Auto-pick with priority chain (sophonts → habitability → resource → iteration order); single `MainworldDesignation string` field on SystemDetail |
| Q5: Class IV-P form scope   | (a) Form 0407F-IV PART P (per-body terrestrial/moon) only; belt form stub; mapping deferred                                                          |

## Architecture

### New step: `runStep5G(r, detailed, sys) error` in `worlds/system_detail_step5g.go`

Slots between Step 5F (3B-biology) and Step 6 (BaselineN backfill). Per-body, per-moon. Mutates `dp.Habitability *Habitability` in place.

After 5G, `DetailSystem` calls `pickMainworld(detailed)` and stores the result on `sd.MainworldDesignation` (new `string` field on `SystemDetail`).

### `DetailSystem` orchestrator additions

```go
// Step 5G — 3B-final pass: per-body habitability rating.
if err := runStep5G(r, detailed, sys); err != nil {
    return SystemDetail{}, err
}

// ... build sd ...

sd.MainworldDesignation = pickMainworld(detailed)
```

### Class IV-P form rendered on demand

Per-body output. NOT stored on `SystemDetail`. Callers invoke `RenderIISSClass4P(body, sys, mainworldDesignation)` for whichever body they want.

### Struct extensions

`DetailedPlacement` and `Moon` each get one new pointer field:

```go
Habitability *Habitability
```

`SystemDetail` gets:

```go
MainworldDesignation string
```

Accessors:

```go
func (dp *DetailedPlacement) HasHabitability() bool { return dp.Habitability != nil }
func (m *Moon) HasHabitability() bool               { return m.Habitability != nil }
```

### Body filter for `runStep5G`

Terrestrials only (skip belts / GGs / empty). Atmosphere optional — vacuum worlds (nil atm) get DM-8 per the atm 0 row. Differs from Biology's filter (which required atm).

## Public API

### The `Habitability` struct

```go
// Habitability — a per-body habitability rating for Terragens per WBH p.132-133.
// Computed by Step 5G for any non-empty terrestrial body (and HZ-planet moons).
//
// Range: 0-12. The book theoretically allows higher but treats 12 as "very
// unlikely" and clamps negative results to 0.
//
// Ratings interpretation (WBH p.133):
//   0       — Actively hostile world: not survivable without specialised equipment
//   1-2     — Barely habitable: full protective equipment needed
//   3-5     — Marginally survivable with proper equipment
//   6-7     — Regionally habitable: may require acclimation
//   8-9     — Suitable for human habitation with minimal equipment or acclimation
//   10-12   — Terra-equivalent garden world (10/A is the Terran baseline)
type Habitability struct {
    Rating int

    // Notes is a referee-color string visible in the Class IV-P form's
    // Habitability section (e.g., "High temperatures hinder habitability").
    // Currently always empty — populated by future referee-feature carry-forward.
    Notes string
}
```

### Pure-function APIs

```go
// ComputeHabitability per WBH p.132: 10 + DMs, clamped to [0, 12].
// Deterministic — no dice. Operates on body's current Atmosphere /
// Hydrographics / Temperature / Physical / SizeCode / TidalLock fields.
//
// Returns Habitability{Rating: 0} if body is nil. For bodies with
// missing pointer fields, the corresponding DMs are skipped (treated
// as 0) — defensive but documented as caller's responsibility.
func ComputeHabitability(body *DetailedPlacement) Habitability { ... }
```

```go
// pickMainworld returns the SAH/UWP designation of the auto-picked
// mainworld per WBH p.134. Priority chain (first match wins):
//   1. Bodies with native sophonts (extant or extinct); among these,
//      highest Habitability; tiebreaker: highest ResourceRating;
//      final tiebreaker: iteration order.
//   2. Highest Habitability among non-sophont bodies; tiebreakers same.
//   3. Highest ResourceRating if no body has Habitability > 0.
//   4. First terrestrial body in iteration order.
//
// Iterates BOTH detailed[i] AND dp.Moons[j]. Returns "" if no terrestrial
// body qualifies.
//
// "Best refuelling location" criterion (WBH p.134) deferred — depends on
// starport infrastructure from pp.147-234.
func pickMainworld(detailed []DetailedPlacement) string { ... }
```

```go
// RenderIISSClass4P renders the WBH p.138 IISS Class IV Survey form
// (FORM 0407F-IV PART P) for a single body. Plain-text output with
// section headers matching the book's form layout.
//
// mainworldDesignation is the SystemDetail.MainworldDesignation; used
// only to mark whether THIS body is the mainworld in the Comments section.
//
// Returns "" if body is nil or body.Body == BodyEmpty.
//
// Belt bodies (Size 0) get a placeholder stub; full Form 0407K-IV PART P.B
// rendering is deferred (see carry-forwards).
func RenderIISSClass4P(body *DetailedPlacement, sys stars.System, mainworldDesignation string) string { ... }
```

### `runStep5G` signature

```go
func runStep5G(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error
```

Same shape as `runStep5C/5D/5E/5F`. Always-nil error return per sibling convention. `r` parameter unused (Habitability is deterministic) but retained for sibling consistency.

## Procedure: Habitability per body

WBH p.132 formula:

```
Habitability = 10 + clamp(SumOfDMs, lower-unbounded, +2) → clamp result to [0, 12]
```

DMs (sum then clamp the final 10+sum to [0, 12]):

### Size DMs

| Size | DM  |
| ---- | --- |
| 0-4  | -1  |
| 9+   | +1  |
| 5-8  | 0   |

### Atmosphere DMs

| Atm code        | DM                                   |
| --------------- | ------------------------------------ |
| 0, 1, A (10)    | -8 (Non-breathable)                  |
| 2, E (14)       | -4 (Very thin tainted / thin low)    |
| 3, D (13)       | -3 (Very thin or very dense)         |
| 4, 9            | -2 (Tainted thin or dense)           |
| 5, 7, 8         | -1 (Thin / standard tainted / dense) |
| 6               | 0 (Standard, baseline)               |
| B (11)          | -10 (Hostile)                        |
| C (12), F+ (15) | -12 (Very hostile)                   |

**Low oxygen taint DM-2** — Deferred per same precedent as 3B-biology Q3-a (Atmosphere taint typology not yet modeled).

### Hydrographics DMs

| Hydro code | DM                                   |
| ---------- | ------------------------------------ |
| 0          | -4 (lack of accessible liquid water) |
| 1-3        | -2 (desert conditions)               |
| 4-8        | 0                                    |
| 9          | -1 (little useable land)             |
| A (10)     | -2 (very little useable land)        |

### Tidal lock DM

| Condition                      | DM  |
| ------------------------------ | --- |
| Solar 1:1 tidally locked world | -2  |

Detection: `dp.HasTidalLock() && dp.TidalLock.IsOneToOnePrimary()` — verify the existing TidalLock struct's API for the 1:1 primary check during implementation. Use whatever method exists on the TidalLock struct.

### Temperature DMs

| Condition           | DM  |
| ------------------- | --- |
| HighK > 323         | -2  |
| HighK < 279         | -2  |
| MeanK > 323         | -4  |
| MeanK in [304, 323] | -2  |
| MeanK < 273         | -2  |
| LowK < 200          | -2  |

Boundary at `MeanK == 323`: matches `[304, 323]` (DM-2) only; `> 323` is strict so 323 itself doesn't trigger -4. Single DM applies → -2. (Per book: footnote says "use worst at edges" — but the bands are `> 323` strict and `[304, 323]` inclusive, so 323 is unambiguously in the [304, 323] band.)

### Gravity DMs (per Q3-a, narrower band wins)

```go
switch {
case g < 0.2:                  return -4   // Unhealthy low gravity
case g >= 0.7 && g <= 0.9:     return +1   // Gravity very comfortable
case g >= 0.4 && g < 0.7:      return -1   // Low gravity (narrower; wins over 0.2-0.7)
case g >= 0.2 && g < 0.4:      return -2   // Very low gravity (residual of 0.2-0.7)
case g > 1.1 && g <= 1.4:      return -1   // Gravity somewhat high
case g > 1.4 && g <= 2.0:      return -3   // Gravity uncomfortably high
case g > 2.0:                  return -6   // Gravity too high for acclimation
}
return 0   // 0.9-1.1 (Earth-like baseline)
```

**Undefined gravity** (when Physical is nil): per WBH `+1 - |6 - Size|`. Vacuum-like Size 0 → `+1 - 6 = -5`. Earth-equivalent unknown Size 6 → `+1 - 0 = +1`.

### Final clamp

```
result := 10 + sumOfDMs
if result < 0  { result = 0 }
if result > 12 { result = 12 }
```

### Worked example (Zed Prime per WBH IISS form p.141)

- Size: 5 → 0
- Atmosphere: 6 → 0
- Hydrographics: 6 (62% coverage) → 0
- Tidal lock: No → 0
- HighK: 346 (> 323) → -2
- MeanK: 300 (no DM range matches) → 0
- LowK: 262 (> 200) → 0
- Gravity: 0.66 (in 0.4-0.7, narrower wins) → -1

**Sum: -3 → Habitability = 10 - 3 = 7** ✓ (matches book exactly)

### Order of operations within `runStep5G`

For each placement:

1. Skip belts (`SizeCode == "0"`), gas giants (`Body == BodyGasGiant`), empty bodies.
2. `dp.Habitability = &Habitability{Rating: ComputeHabitability(dp).Rating}` (Notes left empty per spec).
3. Recurse into moons (terrestrial moons only, including HZ-planet moons).

## Procedure: Mainworld determination

`pickMainworld` priority chain:

1. **Native sophonts** — bodies with `Biology.HasNativeSophont == true` OR `Biology.HadExtinctSophont == true`. Among these: highest Habitability; tiebreaker: highest ResourceRating; final: iteration order.

2. **Highest habitability** — among non-sophont bodies: highest `Habitability.Rating`. Tiebreakers: highest ResourceRating, then iteration order.

3. **Highest resource** — if no body has Habitability > 0: highest `Biology.ResourceRating`. Tiebreaker: iteration order.

4. **First terrestrial** — fallback. Returns "" if no terrestrials exist.

Iterates BOTH `sd.Detailed[i]` AND `dp.Moons[j]`. Returned string is the body's `.Designation` field.

### "Best refuelling location" — deferred

Depends on starport infrastructure from pp.147-234 (next chapter). Documented as carry-forward.

### Empty-system edge case

If `detailed` contains only belts/GGs/empty placements, returns `""`. Class IV-P renderer skips the mainworld annotation when designation is empty.

## Procedure: Class IV-P form rendering

WBH p.138 FORM 0407F-IV PART P. Per-body form. Plain-text rendering with section headers matching the book's form layout.

### Sections

```
IISS CLASS IV SURVEY — FORM 0407F-IV PART P

WORLD: <designation> (<SAH/UWP>)
SECTOR | LOCATION  /  INITIAL SURVEY  /  LAST UPDATED  (system metadata; empty)
PRIMARY OBJECT(S)  /  SYSTEM AGE (Gyr)  /  TRAVEL ZONE

ORBIT
  O#  /  AU  /  Eccentricity  /  Period

SIZE
  Diameter (km)  /  Composition (empty)  /  Density  /  Gravity  /  Mass  /  Esc v (kps; empty)

ATMOSPHERE
  Pressure (bar)  /  Composition (atm.Profile.Gases summary)  /  O2 (bar)
  Taints (empty per Q3-a deferral)  /  Scale Height

HYDROGRAPHICS
  Coverage (%)  /  Composition (from hydro.Profile)  /  Distribution (from SurfaceDistribution)
  Major bodies  /  Minor bodies (empty)

ROTATION
  Sidereal  /  Solar  /  Solar days/year  /  Axial Tilt
  Tidal lock?  /  Tides (TidalEffects.Total m)

TEMPERATURE
  High  /  Mean  /  Low  /  Luminosity  /  Albedo  /  Greenhouse

SEISMIC
  Stress (TSS)  /  Residual  /  Tidal Stress  /  Tidal Heating  /  Major Tectonic Plates

LIFE
  Biomass (eHex)  /  Biocomplexity  /  Sophonts?  /  Biodiversity  /  Compatibility

RESOURCES
  Rating (eHex)  /  Notes (empty)

HABITABILITY
  Rating  /  Notes (empty)

SUBORDINATES (one row per moon)
  SAH/UWP  /  Orbit (PD)  /  Orbit (km)  /  Ecc  /  Diameter  /  Density  /  Mass  /  Period (h)  /  Size (°)

COMMENTS
  "This is the system mainworld" if body.Designation == mainworldDesignation
```

### Field-availability decisions

Many form fields have no source in our pipeline. Render as empty (or a placeholder dash). Don't error on missing data; the form is meant to be populated incrementally. Specifically empty for now:

- Sector | Location, Initial Survey, Last Updated, Travel Zone (system metadata)
- Composition (Size section — would need rock/ice/metal classification)
- Esc v (kps) — would need Earth-equivalent escape velocity formula
- Atmosphere Taints — pending taint typology (Q3-a deferral)
- Major / Minor hydro body counts

### Belt rendering

Per Q5-a: belt bodies (`SizeCode == "0"`) get a stub:

```
IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B (NOT YET IMPLEMENTED)

WORLD: <designation>   (Belt)
Use BeltDetails for resource/composition data.
```

Full Form 0407K rendering is documented as a carry-forward.

## Sub-decisions

- **Habitability is deterministic** (no dice). `runStep5G` consumes 0 × 2D per body and per moon.
- **Mutation policy**: `dp.Habitability = &Habitability{...}` (pointer assignment), consistent with prior sub-projects.
- **Body filter for 5G**: terrestrials only. Atmosphere optional (vacuum → atm 0 DM). Differs from Biology's filter.
- **`MainworldDesignation` is the body's `.Designation`** (e.g., `"Aab IV d"`), NOT the SAH/UWP. Forward-compatible with future referee-override.
- **Class IV-P is per-body, NOT stored on SystemDetail.** Different from Class II/III (per-system). Callers invoke the renderer for whichever body they want.
- **Render is permissive** — empty fields render as empty (or a dash). No errors on missing data.
- **`r` parameter on `runStep5G`** retained for sibling consistency even though unused. `//nolint:unparam` matches existing pattern.

## Documented WBH inconsistencies (carry-forward feedback memories)

After merge, save:

1. **WBH p.132 gravity DM band overlap.** Bands `0.2-0.7 → DM-2` and `0.4-0.7 → DM-1` overlap. Footnote says "use worst at edges" but worked example (Zed Prime gravity 0.66) gets DM-1, contradicting the footnote. Implementation follows the worked example: narrower band wins. Same pattern as Compatibility's "+3" worked-example divergence (3B-biology) and ResidualSeismicStress density DM (3B-geology) — when the formula text and worked example disagree, the worked example tends to be the canonical reference because it's the verifiable target.

## Testing

### Per-formula unit tests (`worlds/habitability_test.go`)

- `TestComputeHabitability_ZedPrime` — Size 5, Atm 6, Hydro 6, no lock, MeanK=300, HighK=346, LowK=262, Gravity=0.66 → 7
- `TestComputeHabitability_HostileAtm_AtmB` — Atm B → DM-10; expect 0 (clamp floor)
- `TestComputeHabitability_VeryHostileAtm_AtmCorF` — Atm C → DM-12; expect 0
- `TestComputeHabitability_TerraEquivalent` — Size 8, Atm 6, Hydro 7, MeanK=288, HighK=315, LowK=255, Gravity=1.0 → 10
- `TestComputeHabitability_GravityNarrowerBandWins` — Gravity=0.5 → DM-1 (NOT -2)
- `TestComputeHabitability_GravityNonOverlap` — Gravity=0.3 → DM-2 (only 0.2-0.7 matches)
- `TestComputeHabitability_GravityBoundary_Exactly0p7` — Gravity=0.7 → in 0.7-0.9 (DM+1)
- `TestComputeHabitability_TidalLock1to1_Minus2` — 1:1 primary lock → DM-2
- `TestComputeHabitability_NilTidalLock_NoDM` — defensive
- `TestComputeHabitability_TempBoundary_MeanExactly323_TriggersHotBand` — MeanK=323 → DM-2 (in [304, 323])
- `TestComputeHabitability_TempBoundary_MeanExactly324_TriggersHotter` — MeanK=324 → DM-4 (>323)
- `TestComputeHabitability_NilAtmosphere_VacuumDM` — nil atm → DM-8
- `TestComputeHabitability_UndefinedGravity_FormulaApplied` — Physical nil, Size 6 → +1
- `TestComputeHabitability_UndefinedGravity_LowSize` — Physical nil, Size 0 → -5
- `TestComputeHabitability_NilBody_ZeroRating` — defensive
- `TestComputeHabitability_HighTemperatureBothBands_DoubleDMs` — HighK=346 + MeanK=346 → -2 -4 = -6
- `TestComputeHabitability_HabitabilityCannotExceed12` — synthetic max-positive → 12

### Mainworld picker tests (`worlds/mainworld_test.go`)

- `TestPickMainworld_SophontWins_OverHigherHabitability`
- `TestPickMainworld_SophontTied_HigherHabitabilityWins`
- `TestPickMainworld_NoSophont_HighestHabitability`
- `TestPickMainworld_HabitabilityTied_HighestResourceWins`
- `TestPickMainworld_AllZero_FirstTerrestrialFallback`
- `TestPickMainworld_BeltsAndGGsOnly_EmptyString`
- `TestPickMainworld_MoonAsMainworld`
- `TestPickMainworld_EmptyDetailed_EmptyString`

### Class IV-P render tests (`worlds/iiss_class4p_test.go`)

Strings-contains assertions:

- `TestRenderIISSClass4P_ZedPrime_KeyFieldsPresent`
- `TestRenderIISSClass4P_NilBody_Empty`
- `TestRenderIISSClass4P_BodyEmpty_Empty`
- `TestRenderIISSClass4P_Belt_StubRendering`
- `TestRenderIISSClass4P_MainworldAnnotation`
- `TestRenderIISSClass4P_NotMainworld_NoAnnotation`
- `TestRenderIISSClass4P_NoMoons_NoSubordinatesSection`
- `TestRenderIISSClass4P_WithMoons_SubordinatesRendered`

### Orchestrator tests (`worlds/system_detail_step5g_test.go`)

- `TestRunStep5G_TerrestrialPopulatesHabitability`
- `TestRunStep5G_GasGiant_NoHabitability`
- `TestRunStep5G_BeltSize0_NoHabitability`
- `TestRunStep5G_BodyEmpty_NoOp`
- `TestRunStep5G_MoonRecursion`
- `TestRunStep5G_VacuumWorld_HasHabitability`

### Acceptance test extension (`worlds/worked_examples_test.go`)

Append assertions 39-43 to `TestZed_FullDetail_3A2b`:

- **39**: HasHabitability() for terrestrial bodies (skip belts/GGs/empty)
- **40**: For all bodies with Habitability: Rating ∈ [0, 12]
- **41**: sd.MainworldDesignation != "" (Zed system has at least one habitable terrestrial)
- **42**: sd.MainworldDesignation matches an existing body's designation in sd.Detailed or any moons
- **43**: Render RenderIISSClass4P for the mainworld; assert non-empty output containing the word "Habitability"

Plus a 7th trailing `t.Logf` note about deferred Form 0407K + World Maps + miscellaneous habitability adjustments.

### Dice budget

Per body and per moon: **0 dice** (Habitability is deterministic). No new dice consumption.

## Carry-forwards (deferred to future sub-projects)

| What                                                                                                                                 | Why                                                                  |
| ------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------- |
| Form 0407K-IV PART P.B (belt Class IV-P rendering)                                                                                   | Different layout; belts get stub rendering. Full form deferred.      |
| World Maps (pp.135-137)                                                                                                              | Visual hex-grid output requires a different rendering category.      |
| Miscellaneous habitability scoring D3-1 (p.133 sidebar)                                                                              | Referee color, not deterministic.                                    |
| Referee-override for mainworld pick                                                                                                  | WBH explicitly punts; current auto-pick is the only source of truth. |
| "Best refuelling location" mainworld criterion                                                                                       | Depends on starport infrastructure from pp.147-234.                  |
| Form 0407F field population: Sector / Location, Composition (Size), Esc v, Atmospheric Taint typology, Major/Minor hydro body counts | System metadata + missing pipeline fields. Render as empty for now.  |
| Habitability `Notes` field — referee-color string                                                                                    | Currently always empty; visible in form but not generated.           |

## File map

| File                                  | Status   | Purpose                                                                                                                                                                                                          |
| ------------------------------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `worlds/habitability.go`              | New      | `Habitability` struct + `ComputeHabitability` + DM helpers                                                                                                                                                       |
| `worlds/habitability_test.go`         | New      | Per-formula tests                                                                                                                                                                                                |
| `worlds/mainworld.go`                 | New      | `pickMainworld(detailed) string`                                                                                                                                                                                 |
| `worlds/mainworld_test.go`            | New      | Priority-chain tests                                                                                                                                                                                             |
| `worlds/iiss_class4p.go`              | New      | `RenderIISSClass4P`                                                                                                                                                                                              |
| `worlds/iiss_class4p_test.go`         | New      | Render-shape tests                                                                                                                                                                                               |
| `worlds/system_detail_step5g.go`      | New      | `runStep5G` + `habitabilityApplies` + `computeHabitability`                                                                                                                                                      |
| `worlds/system_detail_step5g_test.go` | New      | Orchestrator tests                                                                                                                                                                                               |
| `worlds/system_detail.go`             | Modified | Add `Habitability *Habitability` to `DetailedPlacement`; add `MainworldDesignation string` to `SystemDetail`; add `HasHabitability()` accessor; one new line each for Step 5G + mainworld pick in `DetailSystem` |
| `worlds/moons.go`                     | Modified | Add `Habitability *Habitability` to `Moon` + `HasHabitability()` accessor                                                                                                                                        |
| `worlds/worked_examples_test.go`      | Modified | Append assertions 39-43 + 7th trailing t.Logf                                                                                                                                                                    |

## Plan target

Target: 8-11 tasks, executable via subagent-driven-development. Per-task implementer (Sonnet) → spec reviewer → code reviewer → next task. Final end-to-end review on Opus before merge. Same workflow as 3B-biology.

Estimated task layout:

1. Branch + Habitability struct + DetailedPlacement.Habitability + Moon.Habitability + HasHabitability accessors + SystemDetail.MainworldDesignation field
2. ComputeHabitability — Size + Atmosphere + Hydrographics + TidalLock DMs (formula skeleton)
3. ComputeHabitability — Temperature + Gravity DMs (per Q3-a narrower band; finalize formula and clamp)
4. pickMainworld — priority chain + tiebreakers + edge cases
5. runStep5G orchestrator + DetailSystem wiring + mainworld pick call
6. RenderIISSClass4P — World/Sector/Orbit/Size/Atmosphere sections
7. RenderIISSClass4P — Hydrographics/Rotation/Temperature/Seismic sections
8. RenderIISSClass4P — Life/Resources/Habitability/Subordinates/Comments + belt stub
9. Acceptance test extension on TestZed_FullDetail_3A2b (assertions 39-43)
10. Final end-to-end review on Opus + merge
