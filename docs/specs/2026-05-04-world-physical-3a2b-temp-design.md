# World Physical Characteristics — Sub-project 3A2b-temp Design (Mean / High / Low / Scenario Temperatures)

**Date:** 2026-05-04
**Status:** approved through brainstorming; pending user review of written spec
**Source material:** Mongoose Publishing, _World Builder's Handbook_ (Geir Lanesskog, 2023). PDF at `Mongoose/Core Rules/World Builders Handbook.pdf`.

## Context

Builds on 3A2a (merged on `main` as `3fa09e7`). 3A2a's branch produced `DetailedPlacement.AxialTilt`, `DayLength`, `TidalLock`, `TidalEffects`, `SurfaceDistribution` and the `Step 5B` orchestration in `DetailSystem`. 3A2b-temp adds **`Temperature`** as `Step 5C`.

3A2b is split into two sub-projects:

| Sub-project                       | Pages   | Scope                                                                                                                                                                                                                                                                                                                                                           |
| --------------------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **3A2b-temp** _(this spec)_       | 108–126 | Mean temperature (basic table + full equation), albedo, greenhouse factor, high/low temperatures, twilight-zone temperature, altitude factor, multi-source temperature addition, scenario methods (latitude / season / time-of-day / sunlight-portion).                                                                                                         |
| **3A2b-rederive** _(future spec)_ | —       | Iteration loop re-deriving 3A1's `Atmosphere.Pressure`, `Atmosphere.ScaleHeight`, `Hydrographics.Code`, plus Hydrographics chemical formula tail (p.103), Optional Exotic Liquids (p.102), Optional Runaway Greenhouse check (p.111), tidal-lock re-eval if pressure crosses 2.5 bar. Convergence/mutation policy is the rederive sub-project's design problem. |

The 3A2a spec's "tidal heating contribution to mean temperature" carry-forward was based on a misreading: WBH's tidal-heating procedure lives on p.127 in the seismology/tectonics chapter, not the temperature chapter. **Tidal heating is now a 3B concern, not 3A2b.**

P.119 (sky magnitude — "How big/bright is that thing in the sky?") is **out of scope** for 3A2b-temp and beyond. It is a descriptive query (apparent magnitude from a body's surface), not a body characteristic, and has no consumer in the current pipeline.

## Brainstorming decisions (2026-05-04)

| Question                  | Decision                                                                  | Rationale                                                                                                                                                                                            |
| ------------------------- | ------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Q1: Sub-project structure | (b) Split into 3A2b-temp + 3A2b-rederive                                  | Temperature is one coherent computational chunk; re-derive is a fixed-point iteration with its own correctness questions and deserves a separate brainstorm.                                         |
| Q2: Mean temperature path | (b) Equation as canonical, basic table also rolled and stored             | Equation feeds high/low/scenarios. Storing both lets us flag book inconsistencies via `t.Logf` when they diverge (>10K), surfacing the same kind of WBH inter-table issue we caught on p.19 vs p.42. |
| Q3: Output struct shape   | (a) Single rich `Temperature` struct + methods                            | Discoverable scenario API (`t.MeanByLatitude(45)`); cached variance components avoid recompute per call.                                                                                             |
| Q4: Page-range scope      | (b) pp.108-126 + (d) skip p.119                                           | Includes twilight-zone, altitude, multi-source addition through the Zed Prime full worked example on p.126. P.119 sky magnitude is descriptive geometry, not a body characteristic.                  |
| Q5: Re-derive seam        | (a) Clean Step 5C; rederive is future Step 5D                             | Each sub-project owns its own design decisions. Iteration policy is a substantive 3A2b-rederive question.                                                                                            |
| Q6: Acceptance gate       | (a) Replace `TestZed_FullDetail_3A2a` with `TestZed_FullDetail_3A2b_temp` | Matches the 3A1→3A2a replacement pattern. Newest sub-project's gate inherits prior invariants and adds new ones.                                                                                     |

## Architecture

Stay flat in `worlds/`. Three new production files, one new test file:

| File                               | Status  | Responsibility                                                                                                                      |
| ---------------------------------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `worlds/temperature.go`            | **new** | `Temperature` struct, `GenerateTemperature` orchestrator, mean/high/low computation, scenario methods, multi-source addition helper |
| `worlds/temperature_albedo.go`     | **new** | `ComputeAlbedo` + per-world-type table + atmosphere/hydrographics modifiers (p.110)                                                 |
| `worlds/temperature_greenhouse.go` | **new** | `ComputeGreenhouseFactor` + atmosphere code modifier table (p.110)                                                                  |
| `worlds/temperature_test.go`       | **new** | All 3A2b-temp unit tests (Zed Prime + Terra worked examples + scenario method tests + sunlight portion + multi-source)              |
| `worlds/system_detail.go`          | modify  | Add `Temperature *Temperature` field to `DetailedPlacement`; add `HasTemperature()`; wire `Step 5C` after `Step 5B`                 |
| `worlds/moons.go`                  | modify  | Add `Temperature *Temperature` field to `Moon`; add `HasTemperature()`                                                              |
| `worlds/worked_examples_test.go`   | modify  | Replace `TestZed_FullDetail_3A2a` with `TestZed_FullDetail_3A2b_temp`                                                               |

Three production files because albedo and greenhouse are each table-driven with their own modifier rules; keeping them separate from the orchestrator keeps `temperature.go` readable. If `temperature.go` grows past ~700 lines, split scenario methods into `temperature_scenarios.go` (consider during code review, not pre-commit).

**Branch:** `feat/wbh-world-physical-3a2b-temp` — created off `main` at the merge of 3A2a (`3fa09e7`).

## Public API

### `Temperature` struct

```go
// Temperature — per-body temperature characteristics per WBH pp.108-126.
//
// Computed from currently-stored Atmosphere.Pressure, Atmosphere.ScaleHeight,
// and Hydrographics.Code. These 3A1 fields are provisional under HZCO
// temperature until 3A2b-rederive runs an iteration loop that re-derives
// them under real temperature.
type Temperature struct {
    // Headline values (all Kelvin)
    MeanK      float64 // canonical mean per equation (p.111)
    HighK      float64 // p.114 step 9
    LowK       float64 // p.114 step 9
    BasicK     float64 // basic table value (p.109) — sanity-check companion
    WorstHighK float64 // p.115 sidebar
    WorstLowK  float64 // p.115 sidebar

    // Equation inputs
    Luminosity       float64 // total in-group stellar luminosity (solar units)
    Albedo           float64 // 0.02..0.98 per p.110
    GreenhouseFactor float64 // ≥ 0; (1+G) clamped to [0.001, 1.999]
    AU               float64 // distance from primary stellar source

    // Variance components (cached so scenario methods don't recompute)
    AxialTiltFactor    float64
    RotationFactor     float64
    GeographicFactor   float64
    AtmosphericFactor  float64
    LuminosityModifier float64
    NearAU             float64
    FarAU              float64

    // Twilight zone (only populated when body is 1:1 star-locked)
    IsTwilight  bool
    TwilightK   float64 // band centerline = MeanK
    BrightSideK float64 // perpetual day
    DarkSideK   float64 // perpetual night

    // Multi-source addition (moons of GGs primarily)
    ParentRadianceK float64 // contribution from parent body's thermal IR (0 for planets)
}
```

### Public functions

```go
// Top-level pipeline orchestrator (called from Step 5C)
func GenerateTemperature(
    r roller.Roller,
    body *DetailedPlacement,
    sys stars.System,
    parent *DetailedPlacement, // nil for planets, set for moons
) (*Temperature, error)

// Component computations (each independently testable)
func ComputeAlbedo(r roller.Roller, body *DetailedPlacement) float64
func ComputeGreenhouseFactor(r roller.Roller, atm *Atmosphere) float64
func MeanTemperatureK(luminosity, albedo, greenhouse, au float64) float64
func BasicTemperatureRoll(r roller.Roller, body *DetailedPlacement, sys stars.System) (modifiedRoll int, kelvin float64)
func CombineTemperatures(temps ...float64) float64 // ⁴√(T₁⁴ + T₂⁴ + …) per p.109

// Sunlight geometry helper (p.118), exposed for Method 2 uneven-day callers
func SunlightPortion(latDeg, axialTiltDeg, daysSinceSolstice, localYearDays float64) (portion, hours float64)
```

### Methods on `*Temperature`

```go
// Annual mean at a specific latitude, ignoring season and time of day.
func (t *Temperature) MeanByLatitude(latDeg float64) float64

// Mean over a specific day at a specific latitude, ignoring time of day.
func (t *Temperature) MeanBySeason(latDeg, daysSinceSolstice, localYearDays float64) float64

// Instantaneous temperature at a specific moment (uses Method 1 even-days
// internally; callers wanting Method 2 precision should call SunlightPortion
// separately and modulate hoursSinceDawn / solarDayHours accordingly).
func (t *Temperature) AtMoment(latDeg, daysSinceSolstice, localYearDays, hoursSinceDawn, solarDayHours float64) float64

// Adjusts a previously-computed temperature for altitude (p.123-124).
// Re-derives greenhouse factor at altitude pressure (uses body Atmosphere.ScaleHeight).
func (t *Temperature) AdjustedForAltitude(baseTempK, altitudeKm float64) float64
```

**Twilight zone behavior:** When `t.IsTwilight == true`, the three scenario methods document that they return `t.BrightSideK`, `t.TwilightK`, or `t.DarkSideK` based on whether `latDeg` indicates bright-side (longitude > 90° toward star), dark-side, or band — season/time inputs are ignored. Twilight-zone worlds have no meaningful seasonal/diurnal variation per the book.

### `Has*` accessors

```go
func (dp *DetailedPlacement) HasTemperature() bool { return dp.Temperature != nil }
func (m *Moon) HasTemperature() bool                { return m.Temperature != nil }
```

## Procedure (Steps Mapped to Functions)

### Mean Temperature (pp.108-111)

| WBH step                                 | Function                   | Notes                                                                                                                                                              |
| ---------------------------------------- | -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Basic Temperature roll (p.109)           | `BasicTemperatureRoll`     | 2D + DM (HZCO-relative orbit + atmosphere DMs from the p.47 Habitable Zones Regions table). Stored as `BasicK`.                                                    |
| Albedo (p.110)                           | `ComputeAlbedo`            | Table-driven: rocky terr / icy terr / icy terr beyond HZCO+2 / gas giant base + atmosphere & hydrographics modifiers. Result clamped to `[0.02, 0.98]`.            |
| Greenhouse Factor (p.110)                | `ComputeGreenhouseFactor`  | Initial = `0.5 × √(bar)`. Modifiers per atmosphere code (1-9/D/E: +3D×0.01; A or F: ×1D-1 minimum 0.5; B/C/G/H: 1D=1-5 → ×result, 1D=6 → ×3D). Vacuum (atm 0) = 0. |
| Mean equation (p.111)                    | `MeanTemperatureK`         | `279 × ⁴√(L × (1-A) × (1+G) / AU²)` with `(1+G)` clamped to `[0.001, 1.999]`.                                                                                      |
| Multi-source addition (p.111, p.125-126) | `CombineTemperatures(...)` | `⁴√(T₁⁴ + T₂⁴ + …)`. Used inside `GenerateTemperature` to add (a) per-group stellar contributions and (b) gas-giant IR for moons.                                  |

**Divergence log:** if `|MeanK - BasicK| > 10`, emit `t.Logf` (matches the WBH-inconsistency-surfacing pattern from p.19 vs p.42).

**Multi-star luminosity grouping:** stars in the same close-binary group (`OrbitClass==stars.OrbitCompanion && ParentIndex==-1`) **sum their luminosities** → one `T_star`. Stars in _separate_ groups compute their own `T_star` and combine via `CombineTemperatures`. Mirrors 3A2a's `totalStellarMass` pattern for tidal effects.

### High / Low Temperatures (pp.112-114)

Inside `GenerateTemperature`, after `MeanK`:

| Step                             | Computation                                                                                                                                                                                                                                                    | Stored as            |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------- |
| 1: Axial Tilt Factor (p.112-113) | `sin(tilt_radians)`; tilt clamped to `[0°, 90°]` per book; if tilt < 0 take abs; if tilt > 90 use `180 - tilt`. Short-year halving when `Period.Years < 0.1`. Long-year boost (`+ 0.01 × years`, max additive `+0.25`, max value 1.0) when `Period.Years > 2`. | `AxialTiltFactor`    |
| 2: Rotation Factor (p.113)       | `√(\|solar_day_h\|) / 50`; forced to `1.0` if `solar_day > 2500 h` or 1:1 star-locked.                                                                                                                                                                         | `RotationFactor`     |
| 3: Geographic Factor (p.113)     | `(10 - Hyd) / 20`; modifier `+0.1` if SurfaceDistribution is "Very Concentrated" (9+) and `Hyd ∈ [2,8]`; `-0.1` if "Very Distributed" (1-) and `Hyd ∈ [2,8]`.                                                                                                  | `GeographicFactor`   |
| 4: Variance Factor               | `AxialTiltFactor + RotationFactor + GeographicFactor`, clamped `[0, 1]`.                                                                                                                                                                                       | (intermediate)       |
| 5: Atmospheric Factor (p.114)    | `1 + bar`.                                                                                                                                                                                                                                                     | `AtmosphericFactor`  |
| 6: Luminosity Modifier (p.114)   | `Variance / Atmospheric`, clamped `[0, 1]`.                                                                                                                                                                                                                    | `LuminosityModifier` |
| 7: High/Low Luminosity (p.114)   | `L × (1 ± LumModifier)`.                                                                                                                                                                                                                                       | (intermediate)       |
| 8: Near/Far AU (p.114)           | `AU × (1 ∓ eccentricity)`. **Moons use parent planet's eccentricity**, not the moon's own.                                                                                                                                                                     | `NearAU`, `FarAU`    |
| 9: High/Low K (p.114)            | `MeanTemperatureK(highL, A, G, NearAU)` and `MeanTemperatureK(lowL, A, G, FarAU)`.                                                                                                                                                                             | `HighK`, `LowK`      |

**Worst case (p.115 sidebar):** Recomputed with `WorstCaseLumModifier = 1 / (1 + bar/2)` substituted at step 6. Stored as `WorstHighK`, `WorstLowK`.

### Twilight Zone (p.120)

Inside `GenerateTemperature`, branch when `body.TidalLock != nil && body.TidalLock.LockRatio == "1:1" && body.TidalLock.Case == TidalLockCasePlanetToStar`:

- `IsTwilight = true`
- `BrightSideK` = recompute high-temp path with rotation factor forced to `+1.0`
- `DarkSideK` = recompute low-temp path with rotation factor forced to `-1.0`
- `TwilightK = MeanK` (band centerline; rotation factor = 0.0)

Moons locked 1:1 to their parent **planet** are NOT twilight zones — only star-locked planets/moons qualify per p.105 footnote.

### Multi-Source for Moons (p.111, p.125-126)

For moons of gas giants, the parent body radiates IR which adds to the moon's stellar temperature. Order inside `GenerateTemperature` for a moon body:

1. Compute moon's stellar-only `MeanK`/`HighK`/`LowK` per the standard procedure.
2. Determine `T_parent`: use `parent.Temperature.MeanK` (always populated for moons because Step 5C processes planets before their moons). If parent has no `Temperature` yet (defensive — shouldn't happen in normal pipeline), treat `T_parent = 0` and skip.
3. **Pragmatic MVP threshold:** if `T_parent ≤ MeanK + 30K`, set `ParentRadianceK = 0` and skip the combine. The threshold filters out cold gas giants whose IR contribution is dominated by stellar warming and isn't worth modeling.
4. Otherwise, recompute the headline values: `MeanK = CombineTemperatures(MeanK, T_parent)`; same for `HighK`/`LowK`. Set `ParentRadianceK = T_parent`.

The variance components (`AxialTiltFactor` etc.) are NOT modified by parent radiance — they describe stellar-driven variability only.

### Altitude (p.123-124)

`AdjustedForAltitude(baseTempK, altitudeKm)` method only — not stored on the struct (per-call computation):

- Re-derive greenhouse factor with `pressure_at_altitude = surface_pressure × exp(-altitude_km / scale_height_km)`.
- Substitute into `MeanTemperatureK` formula with the original `baseTempK`'s implied luminosity component back-solved.
- Returns adjusted K.

This is approximate — the book's full altitude treatment is more elaborate. We implement the principal effect (greenhouse reduction with altitude); finer effects (lapse-rate, density gradients) deferred.

### Scenario Methods (pp.115-118)

`MeanByLatitude(latDeg)`:

- Determine zone:
  - **Tropical** if `lat ≤ axial_tilt_deg`
  - **Arctic** if `lat ≥ 90 - axial_tilt_deg`
  - **Middle** otherwise (only exists when `axial_tilt < 45°`; else the tropical and arctic zones meet/overlap per Part B p.117)
- Tropical Zone Latitude Adjustment: `sin(45° - axial_tilt_deg)` replaces `AxialTiltFactor` in Lum Modifier
- Middle/Arctic Zone Adjustment: `sin(45° - latitude)` replaces `AxialTiltFactor`
- Part B (tilt ≥ 45°): no middle zone; for `lat ≥ 90 - tilt` use the arctic-case calc; for `lat < 90 - tilt` use the result at the edge `lat = 90 - tilt`.
- Recomputes Lum Modifier with zone-adjusted axial tilt → Latitude Lum → Latitude Temperature equation.

`MeanBySeason(latDeg, daysSinceSolstice, localYearDays)`:

- `lag_days = min(0.1 × stdYearHours, 0.1 × localYearHours) / 24` where `stdYearHours = 8766`
- `adjusted_fractional_year = (daysSinceSolstice - 0.1 × lag_days) / localYearDays`
- `seasonal_axial_tilt = cos(adjusted_fractional_year × 360°) × AxialTiltFactor`
- Then like `MeanByLatitude` with `seasonal_axial_tilt` substituted for the stored `AxialTiltFactor`.

`AtMoment(latDeg, daysSinceSolstice, localYearDays, hoursSinceDawn, solarDayHours)`:

- Same as `MeanBySeason` for axial tilt.
- `adjusted_fractional_day = (hoursSinceDawn / solarDayHours) + 0.15` (lag = 15% of solar day per Method 1)
- `hourly_rotation_factor = sin(adjusted_fractional_day × 360°) × RotationFactor`
- Substitute both seasonal_axial_tilt and hourly_rotation_factor into Lum Modifier → instantaneous K.

`SunlightPortion(latDeg, axialTiltDeg, daysSinceSolstice, localYearDays)`:

- `solar_declination = axialTiltDeg × cos(360° × daysSinceSolstice / localYearDays)`
- `cos_sunrise = tan(latDeg) × tan(solar_declination)` (radians)
- `cos_sunrise > 1` → polar night, return `(0, 0)`
- `cos_sunrise < -1` → polar day, return `(1.0, solarDayHours)`
- Else: `sunrise_angle_deg = arccos(cos_sunrise) × (180/π)`; `portion = sunrise_angle_deg / 180`; `hours = solarDayHours × portion`

## Pipeline Integration (Step 5C)

In `worlds/system_detail.go`, after `Step 5B`:

```go
// Step 5C — 3A2b-temp pass: per-body temperature.
for i := range detailed {
    dp := &detailed[i]
    if dp.Body == BodyEmpty {
        continue
    }
    temp, err := GenerateTemperature(r, dp, sys, nil)
    if err != nil {
        return SystemDetail{}, fmt.Errorf("worlds: temperature %s: %w", dp.Designation, err)
    }
    dp.Temperature = temp

    for j := range dp.Moons {
        m := &dp.Moons[j]
        moonDP := buildMoonPlacementView(m, dp) // existing helper from 3A2a
        moonTemp, err := GenerateTemperature(r, moonDP, sys, dp)
        if err != nil {
            return SystemDetail{}, fmt.Errorf("worlds: moon temperature %s: %w", m.Designation, err)
        }
        m.Temperature = moonTemp
    }
}
```

Step 5C runs **after** all of 5B is complete (DayLength, AxialTilt, TidalLock, TidalEffects, SurfaceDistribution must be in place — temperature consumes axial tilt, day length, tidal-lock state, hydrographics surface distribution).

## Sub-decisions

| Item                                                   | Decision                                                                                                                                                              |
| ------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Multi-star luminosity summation                        | Sum within close-binary group (`OrbitCompanion + ParentIndex==-1`); separate groups combined via `CombineTemperatures`. Mirrors 3A2a's `totalStellarMass`.            |
| Greenhouse factor clamping                             | Clamp `(1+G)` to `[0.001, 1.999]` (book p.111 thumb rule two).                                                                                                        |
| Lag in seasonal adjusted-fractional-year               | `lag = min(0.1 × stdYearHours, 0.1 × localYearHours) / 24` days.                                                                                                      |
| Solstice convention                                    | `daysSinceSolstice` parameter — caller picks the relevant hemisphere's summer solstice. Doc-comment only, no enforcement.                                             |
| Twilight zone detection                                | `body.TidalLock != nil && body.TidalLock.LockRatio == "1:1" && body.TidalLock.Case == TidalLockCasePlanetToStar`. Moon→planet 1:1 locks do NOT trigger twilight zone. |
| Provisional-input doc-comment                          | `Temperature` doc-comment notes its inputs (Atmosphere/Hydrographics) are provisional until 3A2b-rederive runs.                                                       |
| Atmosphere DMs for Basic Temperature roll              | Reuse / extract a shared helper for the p.47 Habitable Zones Regions atmosphere DM table (already used by 3A1's atmosphere generation). Small refactor noted in plan. |
| AU for moons                                           | Parent planet's AU from the star (moons co-orbit star with planet). Same fix-pattern as 3A2a's DayLength final-review correction.                                     |
| Out-of-luminosity check inside HZCO-1 / outside HZCO+1 | Basic-roll DMs (+4 / -4 per 0.5 orbit) only; equation has no such adjustment — distance does the work.                                                                |
| Worst-case temperature                                 | Always computed; stored as `WorstHighK` / `WorstLowK` fields.                                                                                                         |
| Altitude treatment                                     | Approximate — greenhouse reduction with altitude pressure (via scale height). Lapse-rate / density-gradient effects deferred.                                         |
| Parent-IR contribution to moons                        | Pragmatic MVP: skip unless `parent.Temperature.MeanK > moon.MeanK + 30K`. Stored as `ParentRadianceK` (0 if skipped).                                                 |

## Testing Strategy

### Per-file unit tests (all in `worlds/temperature_test.go`)

**Component tests:**

| Test                                           | Pin                                                            |
| ---------------------------------------------- | -------------------------------------------------------------- |
| `TestComputeAlbedo_ZedPrime`                   | 0.33 (rocky terr, atm 6, hyd 6, scripted dice from p.111 walk) |
| `TestComputeAlbedo_Terra_Reference`            | ~0.30 (formula reference; verifies modifier table)             |
| `TestComputeGreenhouseFactor_ZedPrime`         | 0.59 (initial 0.51 + 3D=8 modifier per p.111)                  |
| `TestMeanTemperatureK_Formula_ZedPrime`        | 300K from L=1.419, A=0.33, G=0.59, AU=1.06                     |
| `TestMeanTemperatureK_Formula_Terra_Reference` | 288K from L=1.0, A=0.30, G=0.36, AU=1.0                        |
| `TestBasicTemperatureRoll_Mod7`                | Modified roll 7 → 288K (table value)                           |
| `TestCombineTemperatures_TwoEqual`             | `T,T → T × ⁴√2`                                                |
| `TestCombineTemperatures_BookExample`          | per p.109 worked example if available; else 300K + 300K → 357K |

**Orchestrator tests (`GenerateTemperature` end-to-end with scripted dice):**

| Test                                              | Pin                                                                                                         |
| ------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `TestGenerateTemperature_ZedPrime_Mean`           | 300K                                                                                                        |
| `TestGenerateTemperature_ZedPrime_HighLow`        | High=346K, Low=250K (p.114)                                                                                 |
| `TestGenerateTemperature_ZedPrime_WorstCase`      | WorstHigh=359K, WorstLow=230K (p.115 sidebar)                                                               |
| `TestGenerateTemperature_Terra_HighLow_Reference` | High=312K, Low=261K (p.114 Terra comparison)                                                                |
| `TestGenerateTemperature_BasicMeanLogsDivergence` | Forces a body where MeanK − BasicK > 10K; verifies `t.Logf` fires (capture via `t.Helper()` or test logger) |

**Scenario method tests:**

| Test                                       | Pin                                                                 |
| ------------------------------------------ | ------------------------------------------------------------------- |
| `TestMeanByLatitude_Tropical`              | tilt=23.45°, lat=10° → in-zone formula validates                    |
| `TestMeanByLatitude_Arctic`                | tilt=23.45°, lat=80° → arctic-zone formula                          |
| `TestMeanByLatitude_HighTilt_NoMiddleZone` | tilt=60°, lat=30° → Part B logic                                    |
| `TestMeanBySeason_SummerSolstice`          | days=0 → max axial-tilt contribution                                |
| `TestMeanBySeason_WinterSolstice`          | days=year/2 → opposite-sign contribution                            |
| `TestAtMoment_Noon_Equator`                | hours=solar_day/2, lat=0 → max diurnal contribution                 |
| `TestAtMoment_Dawn`                        | hours=0 → near-mean (lag 0.15 means coldest moment is dawn-shifted) |
| `TestAdjustedForAltitude_8000m`            | Terra mean -10K to -15K at 8000m (sanity range)                     |

**Twilight zone tests:**

| Test                                                          | Pin                                                                                         |
| ------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `TestGenerateTemperature_TwilightZone_Detected`               | Synthetic 1:1 star-locked body has `IsTwilight=true`, `BrightSideK > TwilightK > DarkSideK` |
| `TestGenerateTemperature_TwilightZone_BrightSideAtSubstellar` | `MeanByLatitude(0)` returns `BrightSideK`                                                   |

**Multi-source tests:**

| Test                                            | Pin                                                                               |
| ----------------------------------------------- | --------------------------------------------------------------------------------- |
| `TestGenerateTemperature_GGMoon_ParentRadiance` | Synthetic moon of warm GG: `ParentRadianceK > 0` and contributes to combined mean |

**Sunlight portion:**

| Test                                      | Pin                              |
| ----------------------------------------- | -------------------------------- |
| `TestSunlightPortion_Equator_Equinox`     | 0.5 regardless of axial tilt     |
| `TestSunlightPortion_Pole_SummerSolstice` | 1.0 (polar day)                  |
| `TestSunlightPortion_Pole_WinterSolstice` | 0.0 (polar night)                |
| `TestSunlightPortion_45N_SummerSolstice`  | ~16h daylight on Terra reference |

### Composite acceptance gate (`worked_examples_test.go`)

**`TestZed_FullDetail_3A2b_temp`** replaces `TestZed_FullDetail_3A2a`. Inherits all 8 prior assertions, adds:

| #   | Assertion                                                                               |
| --- | --------------------------------------------------------------------------------------- |
| 9   | Every non-empty body has `Temperature` populated                                        |
| 10  | Every body: `LowK ≤ MeanK ≤ HighK`                                                      |
| 11  | Every body: `WorstLowK ≤ LowK` and `HighK ≤ WorstHighK`                                 |
| 12  | Every body's `Albedo ∈ [0.02, 0.98]`                                                    |
| 13  | Every body's `GreenhouseFactor ≥ 0` and `< 5` (sanity bound)                            |
| 14  | Every HZ body has `MeanK ∈ [180K, 400K]` (loose plausibility)                           |
| 15  | For 1:1 star-locked bodies: `IsTwilight=true` and `DarkSideK < TwilightK < BrightSideK` |
| 16  | Informational `t.Logf` if `\|MeanK - BasicK\| > 10K` (book inconsistency surfacing)     |

100-iteration free-dice loop with seeds 0..99. Pattern unchanged from 3A2a.

## Carry-forwards from 3A2b-temp to 3A2b-rederive

After 3A2b-temp lands, 3A2b-rederive owns:

- Iteration loop re-deriving `Atmosphere.Pressure`, `Atmosphere.ScaleHeight`, `Hydrographics.Code` under real `Temperature.MeanK`. Convergence policy (single pass vs fixed-point) and mutation policy (overwrite vs shadow fields) are 3A2b-rederive design questions.
- Hydrographics chemical formula tail (`H-D:%%:XX-##:YY-##` per p.103).
- Optional Exotic Liquids table (p.102) — choice of liquid based on real temperature.
- Optional Runaway Greenhouse check (p.111) — fires on Atmosphere 2-F with mean temp >303K; mutates atmosphere; triggers re-derive iteration.
- Tidal-lock re-eval if pressure correction crosses 2.5-bar threshold (3A2a's tidal-lock common DM uses provisional pressure; if real pressure shifts the DM across the bar threshold, lock outcome may change).

## Carry-forwards deferred beyond 3A2b

- **Tidal heating contribution.** Confirmed in seismology chapter (p.127); 3B owns it.
- **Sky magnitude / apparent magnitude** (p.119). Descriptive geometry, not a body characteristic. Defer indefinitely.
- **Twilight Zone Variability Factors** (terrain, libration, refraction — pp.121-122). Refinement layer on top of basic twilight scenario; defer.
- **Full altitude treatment** (lapse rate, density gradients). 3A2b-temp implements only the principal greenhouse-with-altitude effect.

## Branch and merge

Branch: `feat/wbh-world-physical-3a2b-temp` off `main` at `3fa09e7` (3A2a merge).

After all tasks complete:

```bash
just check && just test
git log --oneline main..HEAD
```

Expected: ~12-14 commits ahead of main, all `ok` from test, `0 issues.` from check.

Merge (after user approval):

```bash
git checkout main
git merge --no-ff feat/wbh-world-physical-3a2b-temp -m "Merge feat/wbh-world-physical-3a2b-temp: World Physical 3A2b temperature complete"
```

After merge, update `MEMORY.md` Subprojects line: 3A2b-temp complete; next is 3A2b-rederive.
