# World Physical Characteristics — Sub-project 3A2a Design (Surface Distribution + Rotation + Tilt + Tidal)

**Date:** 2026-05-04
**Status:** approved through brainstorming; pending user review of written spec
**Source material:** Mongoose Publishing, _World Builder's Handbook_ (Geir Lanesskog, 2023). PDF at `Mongoose/Core Rules/World Builders Handbook.pdf`.
**Source pages:** WBH pp. 100–108.
**Parent spec:** `docs/pass-1/specs/2026-05-02-world-builder-design.md`.
**Predecessor:** `docs/pass-1/specs/2026-05-03-world-physical-3a1-design.md` (3A1 — Body Physical + Atmosphere + Hydrographics).

## Purpose

Encode the World Physical Characteristics procedures for surface feature distribution, rotation period (day length), axial tilt, tidal lock determination, and surface tidal effects. Sub-project 3A2a layers on top of 3A1's `DetailedPlacement` and produces:

1. **Surface Feature Distribution** for every terrestrial body and HZ-planet moon with hydrographics: 2D-2 distribution code (Extremely Dispersed → Extremely Concentrated) plus fundamental geography (ocean-major or land-major), including the 1D fundamental-geography rule for Hydrographics = 5 worlds.
2. **Day Length** for every non-empty body and moon: sidereal day, year days, solar day, including the 40+ hour reroll cascade and the GG/Size-0/S × 2 modifier.
3. **Axial Tilt** for every non-empty body and moon: basic tilt table + Extreme Axial Tilt table for 10+ rolls, retrograde detection, baseline preservation through tidal-lock mutations.
4. **Tidal Lock determination** for all three cases (planet→star, moon→planet, planet→moon) with the full DM stack from p.106; case selection by highest DM with p.106 tiebreakers; effect application including 3:2/1:1 lock state, day-length multipliers, prograde/retrograde rotation rerolls, axial-tilt mutation on 3:2/1:1 lock, eccentricity mutation on 1:1 lock, twilight-zone-world flagging on planet→star 1:1 lock, and the natural-12-verification reroll branch.
5. **Surface Tidal Effects** for every body: star tidal contribution(s) summed, moon tidal contribution(s), parent-planet tide on locked moons, with per-source breakdown.

3A2a does **not** complete the hydrographics-profile chemical-formula tail (`H-D:%%:XX-##:YY-##` from p.103) or the Optional Exotic Liquids table (p.102). Those are deferred to 3A2b's re-derivation pass alongside pressure / scale-height / hydrographics-code under real temperature.

## Decomposition context

The World Physical Characteristics chapter (WBH pp. 69–146) decomposes into:

| Sub-project | WBH pp.     | Status                                                                                                                                                                                                                                                                                          |
| ----------- | ----------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 3A1         | 69–100      | **Done** — merged 2026-05-04 (`bc30bc0`). Body physical, belt details, moon refinement, atmosphere (digit + subtype + profile), hydrographics.                                                                                                                                                  |
| **3A2a**    | **100–108** | **This spec.** Surface distribution, rotation, axial tilt, tidal locks, surface tidal effects.                                                                                                                                                                                                  |
| 3A2b        | 108–118     | Mean temperature (basic + greenhouse + albedo), high/low temperatures, temperature-by-season, temperature-by-latitude, temperature-by-time-of-day. Re-derives 3A1's provisional pressure/scale-height/hydrographics-code under real temperature. Completes hydrographics chemical-formula tail. |
| 3B          | 119–140     | Biosphere & habitability. Likely home for seismology (no seismology section appears in 3A2a/3A2b's pages — original resume notes erred on its location).                                                                                                                                        |
| 3C          | 141–146     | Final mainworld picker + world maps + `ClassIIIStatus = true` trigger.                                                                                                                                                                                                                          |

The original "3A2 = Temperature + Seismology, pp. 101–118" framing in the resume notes was wrong on two counts: (1) seismology is not in pp. 100–118; (2) the actual scope is 13 sub-systems and ~45% larger than 3A1, requiring decomposition. 3A2a + 3A2b is the chosen split because temperature (3A2b) depends on rotation, axial tilt, and tidal-lock state (3A2a outputs) but not vice versa.

## Non-goals

- **Mean temperature roll, greenhouse factor, high/low temperatures, latitude/season/time-of-day variations.** 3A2b owns. 3A2a continues to use 3A1's `HZCOOffsetToTempRange` / `tempRangeMidpointK` placeholders where temperature is referenced (e.g. atmosphere scale-height, currently provisional).
- **Hydrographics Profile chemical-formula tail.** Per Q1 option (c), 3A2a does not render `XX-##:YY-##` chemical formulas. Hydrographics output remains digit / range / percent. 3A2b's re-derivation pass handles water vs exotic liquid choice under real temperature.
- **Optional Exotic Liquids table (p.102).** Deferred to 3A2b alongside hydrographics chemical formula.
- **Continent counts** (book p.101 worked example: ~2 major + ~9 minor for Zed Prime). Per Q6 option (b), explicitly Referee fiat per the book; not encoded.
- **Moon-to-moon tidal effects** (p.108). Optional/Referee fiat per the book; documented as future work in 3B or later.
- **Re-derivation of tidal lock under post-3A2b corrected pressure.** 3A2a's tidal-lock DM uses 3A1's provisional pressure for the `> 2.5 bar → DM-2` rule. After 3A2b corrects pressure, the tidal-lock outcome may shift; documented as a 3A2b carry-forward.
- **Hill sphere / Roche limit / moon orbit recomputation under post-lock eccentricity** (per Q2 option c — targeted recompute). Pre-lock snapshot accepted; documented as a small drift.
- **Seismology.** Not in pp. 100–118. Likely 3B.
- **`PickMainworld()`.** Needs habitability rating from 3B. 3C.
- **Form schema changes.** Per Q5 option (a), all 3A2a outputs are struct fields only. The IISS Class II/III form is unchanged. WBH's own p.63 form sample doesn't show rotation/tilt/tidal — those are referee detail, not survey output.

## Bundled 3A1 carry-forward items

3A2a does **not** close any of the 3A1 carry-forwards. They remain on 3A2b:

- `Atmosphere.Pressure`, `Atmosphere.ScaleHeight`, `Hydrographics.Code` provisional under HZCO temperature → 3A2b re-derivation pass.
- `tempRangeMidpointK` placeholder → 3A2b real temperature.
- HZ-candidate atmosphere subtype follow-ups → 3B.
- Significant rings (R0# notation, removed-moon promotion) → 3B or later.
- `t.Logf` book inconsistencies (Aab PI resource rating, Aab PI Size-S, Cab PI profile) → standing logs; no closure planned.

## Decision log

| Question                                         | Decision                                                                           | Reason                                                                                                                                                                     |
| ------------------------------------------------ | ---------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Q1: Hydrographics Profile / Exotic Liquids scope | (c) defer entirely to 3A2b                                                         | Choice depends on real temperature; ride along with 3A2b's re-derive pass.                                                                                                 |
| Q2: Retroactive mutations from tidal lock        | (c) targeted recompute — only `Placement.Eccentricity` mutates                     | Confirmed during section 3 design that even Period doesn't actually need recomputing (Kepler's 3rd is independent of ecc); only downstream consumer is 3A2b's Near/Far AU. |
| Q3: Acceptance gate                              | (c) composite `TestZed_FullDetail_3A2a` with `t.Run` sub-tests; delete 3A1 version | Project pattern is "latest acceptance test is canonical"; 3A1 already deleted 2C's.                                                                                        |
| Q4: Tidal lock case coverage                     | (a) full generic — all three cases                                                 | Reference impl, book-faithful; case 3 is conditional and cheap.                                                                                                            |
| Q5: 3A2a output rendering                        | (a) struct data only, no form changes                                              | WBH's own form doesn't show rotation/tilt/tidal; keeps spec narrow.                                                                                                        |
| Q6: Surface Feature Distribution scope           | (b) 2D-2 roll + Hydro=5 1D fundamental-geography rule; skip continent counts       | Book explicitly defers continent counts to Referee fiat.                                                                                                                   |
| Approach: file layout                            | (b) stay flat in `worlds/`, one file per pipeline stage                            | Continues 2A/2B/2C/3A1 precedent; per-file subagent review caught 6+ bugs in 3A1.                                                                                          |

## Architecture

### File layout

| File                       | LOC est. | Scope                                                                                                                                                                                                                                                                                                                              |
| -------------------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `surface_distribution.go`  | ~150     | 2D-2 surface distribution roll; Hydro=5 1D fundamental-geography rule; description lookup.                                                                                                                                                                                                                                         |
| `day_length.go`            | ~250     | Sidereal day formula `(2D-2)×4 + 2 + 1D + DMs`; 40+ hour reroll cascade; year-days arithmetic; solar-day arithmetic; GG/Size 0/S × 2 modifier; system-age DM.                                                                                                                                                                      |
| `axial_tilt.go`            | ~200     | Basic axial tilt table (2D); Extreme Axial Tilt table (1D for 10+); retrograde detection; baseline preservation.                                                                                                                                                                                                                   |
| `tidal_lock.go`            | ~500     | Per-case DM evaluation (planet→star, moon→planet, planet→moon); case selection (highest DM, p.106 tiebreakers); roll on Tidal Lock Status table; effect application (multipliers, prograde/retrograde rerolls, 3:2/1:1 lock state, axial-tilt and eccentricity mutations, twilight-zone flagging, natural-12 verification reroll). |
| `surface_tidal_effects.go` | ~250     | Star tidal formula; moon tidal (locked or unlocked); planet tidal (moon not locked); per-component breakdown; multi-source summing.                                                                                                                                                                                                |

Each file has a colocated `_test.go` with table-driven unit coverage. Total estimate: ~1,350 new LOC + ~600 LOC tests + ~150 LOC of edits to `system_detail.go`, `moons.go`, and `worked_examples_test.go`.

### Edits to existing files

- `worlds/system_detail.go` — adds Step 5B pipeline phase wiring `5B.1`–`5B.5`; adds new pointer fields on `DetailedPlacement`; adds `Has*` accessors.
- `worlds/moons.go` — adds new pointer fields on `Moon`.
- `worlds/worked_examples_test.go` — replaces `TestZed_FullDetail_3A1` with composite `TestZed_FullDetail_3A2a`.

### Public API additions

#### `worlds/surface_distribution.go`

```go
package worlds

// SurfaceDistribution — landmass concentration per WBH p.100.
type SurfaceDistribution struct {
    Code        int                  // 0..10 (capped from 2D-2 with -1 → 0, 11+ → 10)
    Description string               // "Extremely Dispersed"|...|"Extremely Concentrated"
    Geography   FundamentalGeography // Ocean (water major) | Land (land major)
}

type FundamentalGeography int

const (
    GeographyOcean FundamentalGeography = iota
    GeographyLand
)

// RollSurfaceDistribution rolls 2D-2 and clamps to [0, 10] per WBH p.100.
func RollSurfaceDistribution(r roller.Roller) (int, error)

// DescribeSurfaceDistribution maps Code → Description from the p.100 table.
func DescribeSurfaceDistribution(code int) string

// DetermineFundamentalGeography per WBH p.101:
//   Hydrographics 6+ → Ocean
//   Hydrographics 4- → Land
//   Hydrographics 5  → 1D, 1-3 → Ocean, 4-6 → Land
func DetermineFundamentalGeography(r roller.Roller, hydroCode int) (FundamentalGeography, error)

// GenerateSurfaceDistribution orchestrates the per-body pipeline. Skipped
// (returns nil, nil) for non-HZ bodies and bodies without Hydrographics.
func GenerateSurfaceDistribution(r roller.Roller, hydro *Hydrographics) (*SurfaceDistribution, error)
```

#### `worlds/day_length.go`

```go
package worlds

// DayLength — rotation periods per WBH pp.103-104.
type DayLength struct {
    SiderealHours         float64 // post-lock final value
    SolarHours            float64 // 0 if 1:1 star lock (twilight zone)
    YearDays              float64 // local solar days = year_h / sidereal_h - 1
    BaselineSiderealHours float64 // raw roll result, pre-tidal-lock
}

// DayLengthDMs accumulates DMs for the basic rotation roll, WBH p.103.
type DayLengthDMs struct {
    SystemAgeGyr  float64 // DM+1 per 2 Gyrs (round down)
    IsGGOrSizeS   bool    // multiplies result by 2
}

// RollBasicSiderealHours: (2D-2) × 4 + 2 + 1D + DMs, with 40+ reroll cascade
// (1D ≥ 5 → add another roll, repeat).
func RollBasicSiderealHours(r roller.Roller, dms DayLengthDMs) (float64, error)

// ComputeYearDays: year_hours / sidereal_hours - 1 (per WBH p.104).
func ComputeYearDays(yearHours, siderealHours float64) float64

// ComputeSolarHours: year_hours / year_days (per WBH p.104).
func ComputeSolarHours(yearHours, yearDays float64) float64

// GenerateDayLength orchestrates per-body. Returns nil for empty bodies.
func GenerateDayLength(r roller.Roller, dp *DetailedPlacement, sys stars.System) (*DayLength, error)
```

#### `worlds/axial_tilt.go`

```go
package worlds

// AxialTilt — world's obliquity per WBH p.104.
type AxialTilt struct {
    Degrees         float64 // 0..180; >90 = retrograde
    Retrograde      bool    // 90 < tilt <= 180
    BaselineDegrees float64 // pre-lock value
}

// RollBasicAxialTilt: 2D table per p.104. Returns the rolled tilt;
// 10+ caller dispatches to RollExtremeAxialTilt.
func RollBasicAxialTilt(r roller.Roller) (float64, error)

// RollExtremeAxialTilt: 1D Extreme Axial Tilt table per p.104.
//   1-2: 10 + 1D × 10  (20-70°)
//   3:   30 + 1D × 10  (40-90°)
//   4:   90 + 1D       (91-126°, retrograde)
//   5:   180 - 1D × 1D (144-180°, extreme retrograde)
//   6:   120 + 1D × 10 (130-180°, extreme retrograde, high variance)
func RollExtremeAxialTilt(r roller.Roller) (float64, error)

// GenerateAxialTilt orchestrates per-body. Returns nil for empty bodies.
func GenerateAxialTilt(r roller.Roller, dp *DetailedPlacement) (*AxialTilt, error)
```

#### `worlds/tidal_lock.go`

```go
package worlds

// TidalLock — tidal lock state per WBH pp.105-107.
type TidalLock struct {
    Case                 TidalLockCase
    InitialResult        int    // 2D + DM (pre-verification)
    FinalResult          int    // post-verification (= InitialResult when no verification fired)
    VerificationFired    bool   // p.105 footnote: InitialResult ≥ 12 triggered 2D verification, natural 12 caused a no-DM reroll (FinalResult holds the reroll)
    LockRatio            string // "" | "3:2" | "1:1"
    IsTwilightZone       bool   // true only if Case == PlanetToStar AND LockRatio == "1:1"

    // Effect descriptors — set based on FinalResult
    DayLengthMultiplier float64 // 1.5 / 2 / 3 / 5 for FinalResult 3-6
    NewSiderealHours    float64 // for prograde/retrograde reroll FinalResult 7-10
    BecomesRetrograde   bool    // FinalResult 9-10
    EccentricityMutated bool    // 1:1 lock with old ecc > 0.1
    AxialTiltMutated    bool    // 3:2 or 1:1 lock with old tilt > 3°
}

type TidalLockCase int

const (
    TidalLockCaseNone TidalLockCase = iota // total DMs ≤ -10 → no roll
    TidalLockCasePlanetToStar
    TidalLockCaseMoonToPlanet
    TidalLockCasePlanetToMoon
)

// EvaluateTidalLockDMs returns per-case DM totals. Cases that don't apply
// (no parent planet for moon→planet, no moons for planet→moon, etc.) are
// absent from the map.
func EvaluateTidalLockDMs(
    body *DetailedPlacement,
    sys stars.System,
    parentPlanet *DetailedPlacement, // nil if body is a planet
    moonRef *Moon,                    // nil if body is a planet
) map[TidalLockCase]int

// SelectHighestDMCase returns the case to roll, applying p.106 tiebreakers:
//   - Cases with DM ≤ -10 are filtered out.
//   - On ties: moon-cases ordered first, closest moon first.
//   - Returns TidalLockCaseNone if no case applies.
func SelectHighestDMCase(dms map[TidalLockCase]int, body *DetailedPlacement) (TidalLockCase, int)

// RollTidalLockStatus rolls 2D + DM. Caller handles the natural-12-verification branch.
func RollTidalLockStatus(r roller.Roller, dm int) int

// ApplyTidalLockEffect mutates body fields and possibly Placement.Eccentricity
// based on the rolled result. Handles natural-12 verification reroll when
// result ≥ 12 (rolls 2D, on natural 12 rolls TidalLockStatus again with DM=0).
// Recomputes YearDays and SolarHours after SiderealHours mutation.
func ApplyTidalLockEffect(
    r roller.Roller,
    body *DetailedPlacement,
    moonRef *Moon, // nil if body is a planet
    kase TidalLockCase,
    result int,
    yearHours float64,
) (TidalLock, error)

// GenerateTidalLock orchestrates per-body: evaluate → select → roll → apply →
// possibly roll a tied case if the chosen case rolled < 12 (per p.106).
func GenerateTidalLock(
    r roller.Roller,
    body *DetailedPlacement,
    moonRef *Moon,
    sys stars.System,
    parentPlanet *DetailedPlacement,
    yearHours float64,
) (*TidalLock, error)
```

#### `worlds/surface_tidal_effects.go`

```go
package worlds

// SurfaceTidalEffects — tidal amplitudes from various bodies per WBH pp.107-108.
type SurfaceTidalEffects struct {
    Total      float64          // meters — sum of all components
    Components []TidalComponent
}

type TidalComponent struct {
    Source string  // "star Aab" | "planet Aab IV (1,200⊕)" | "moon Aab I a"
    Meters float64
}

// StarTide: Star Mass × Planet Size / (32 × AU³) per WBH p.107.
//   For close binary primaries, sum the masses (per Zed Prime example: 1.836).
func StarTide(starMassSolar, planetSizeN int, auFromStar float64) float64

// MoonTideOnPlanet: Moon Mass × Planet Size / (3.2 × (Moon Distance(km)/1,000,000)³) per WBH p.108.
//   Locked moons reportedly produce a 16× rise; for amplitudes treat the formula
//   as the open-ocean baseline.
func MoonTideOnPlanet(moonMassEarth float64, planetSizeN int, moonOrbitKm float64) float64

// PlanetTideOnMoon: Planet Mass × Moon Size / (3.2 × (Moon Distance(km)/1,000,000)³)
// per WBH p.108. Applied only when the moon is NOT 1:1 locked to the planet.
func PlanetTideOnMoon(planetMassEarth float64, moonSizeN int, moonOrbitKm float64) float64

// GenerateSurfaceTidalEffects orchestrates per-body, summing star contributions,
// per-moon contributions (for planets), and parent-planet contribution (for
// unlocked moons). Out of scope: moon-to-moon (Referee fiat).
func GenerateSurfaceTidalEffects(
    body *DetailedPlacement,
    moonRef *Moon,
    sys stars.System,
    parentPlanet *DetailedPlacement,
) (*SurfaceTidalEffects, error)
```

### Field additions to existing structs

```go
type DetailedPlacement struct {
    // ... existing 2C and 3A1 fields ...

    // 3A2a additions
    SurfaceDistribution *SurfaceDistribution
    DayLength           *DayLength
    AxialTilt           *AxialTilt
    TidalLock           *TidalLock
    TidalEffects        *SurfaceTidalEffects
}

type Moon struct {
    // ... existing 2C and 3A1 fields ...

    // 3A2a additions
    SurfaceDistribution *SurfaceDistribution // HZ-planet moons with Hydrographics
    DayLength           *DayLength
    AxialTilt           *AxialTilt
    TidalLock           *TidalLock
    TidalEffects        *SurfaceTidalEffects
}
```

`Has*` accessor methods on both structs follow the 3A1 pattern (`HasSurfaceDistribution`, `HasDayLength`, etc.).

## Pipeline integration

### Step 5B (NEW) — sequenced after 3A1's Step 5A

```text
For each terrestrial body or HZ-planet moon (in DetailedPlacement traversal order):

5B.1  Surface feature distribution (terrestrials + HZ moons with hydrographics)
        Inputs:  Hydrographics
        Output:  SurfaceDistribution{Code, Description, Geography}

5B.2  Day length (per body and per moon)
        Inputs:  Body type (GG/Size 0/S → ×2), Period, system age
        Output:  DayLength{SiderealHours, SolarHours, YearDays, BaselineSiderealHours}

5B.3  Axial tilt (per body and per moon)
        Inputs:  none external (rolled fresh per body)
        Output:  AxialTilt{Degrees, Retrograde, BaselineDegrees}

5B.4  Tidal lock determination (per body and per moon)
        Inputs:  Size, Eccentricity, Atmosphere.Pressure (provisional), AxialTilt,
                 sys.Stars (mass, AU), Moon orbit/mass/retrograde, system age
        Procedure (per p.106):
          a. EvaluateTidalLockDMs → per-case DM totals
          b. Filter cases with total DM ≤ -10 (no roll needed); if none remain,
             TidalLockCaseNone (skip 5B.4 entirely).
          c. Identify the highest-DM tier; if multiple cases tie at that DM,
             order them: moon-cases first, then star-case; closest moon first.
          d. Roll RollTidalLockStatus(r, dm) for each tied case in tiebreaker
             order. Track the maximum 2D+DM (the InitialResult).
          e. If InitialResult ≥ 12: roll a 2D verification per p.105 footnote;
             on natural 12, reroll RollTidalLockStatus with DM=0 → that becomes
             FinalResult; otherwise FinalResult = InitialResult.
          f. ApplyTidalLockEffect using FinalResult → mutates body fields and
             possibly Placement.Eccentricity.
        Mutation policy (per Q2 option c):
          - Mutates SiderealHours (recomputes YearDays + SolarHours)
          - Mutates AxialTilt.Degrees on 3:2 or 1:1 lock with old tilt > 3°
          - Mutates Placement.Eccentricity on 1:1 lock with old ecc > 0.1
          - Hill sphere / Roche / moon orbit values left at 3A1 snapshot

5B.5  Surface tidal effects (per body and per moon)
        Inputs:  Body mass/size, all stars, parent (for moons), all moons,
                 post-tidal-lock state
        Output:  SurfaceTidalEffects{Total, Components}
```

### Iteration order

Outer loop walks `sd.Detailed` in ascending-orbit order (existing 2C ordering); inner loop for moons walks `dp.Moons` in placement order. Each sub-step (5B.1...5B.5) completes its full traversal before the next begins. This matches 3A1's pattern and keeps scripted-dice tests deterministic.

### Effect application table

Effects key off `FinalResult` (post-verification, per the procedure above).

| FinalResult | Effect on body                                                                                                                                                               | Mutates                                                                                             |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| ≤ 2         | None                                                                                                                                                                         | nothing                                                                                             |
| 3           | Day length × 1.5                                                                                                                                                             | SiderealHours; recompute YearDays + SolarHours                                                      |
| 4           | × 2                                                                                                                                                                          | same                                                                                                |
| 5           | × 3                                                                                                                                                                          | same                                                                                                |
| 6           | × 5                                                                                                                                                                          | same                                                                                                |
| 7           | Prograde rotation, sidereal = `1D × 5 × 24` h                                                                                                                                | SiderealHours                                                                                       |
| 8           | Prograde, `1D × 20 × 24` h                                                                                                                                                   | same                                                                                                |
| 9           | Retrograde, `1D × 10 × 24` h; if old tilt < 90°, set tilt = 180 − tilt                                                                                                       | SiderealHours, AxialTilt                                                                            |
| 10          | Retrograde, `1D × 50 × 24` h; same tilt rule                                                                                                                                 | same                                                                                                |
| 11          | 3:2 lock; sidereal = orbital period × 2/3; if old tilt > 3°, reroll as `(2D-2)/10`                                                                                           | SiderealHours, LockRatio, AxialTilt (cond.)                                                         |
| 12+         | 1:1 lock; sidereal = orbital period; if star case → IsTwilightZone, SolarHours = 0; if old tilt > 3°, reroll as `(2D-2)/10`; if old ecc > 0.1, reroll ecc with DM-2 take min | SiderealHours, LockRatio, AxialTilt (cond.), Placement.Eccentricity (cond.), IsTwilightZone (cond.) |

### Multi-body cases

- **Multi-star planet→star.** DM stack computed against each star; highest DM wins. If multiple stars tie, p.106 tiebreaker prefers closer star (lower AU separation).
- **Multi-moon planet→moon.** Per moon, evaluated independently; the conditional ("planet has at least one already-locked moon") applies as preconditions before evaluation. Closest moon checked first per p.106.
- **Multi-moon moon→planet.** Each moon rolls independently against its parent planet.
- **Multi-star surface tidal effects.** Sum all star tidal contributions. For close binary primaries (Zed Aab — primary group A), the book's example sums their masses (1.836 M☉) and treats them as a single source at the system AU. Encode by `stars` group: stars within the same group (Aa + Ab as group A) sum their masses and act as one source at the group's barycentric AU; stars in different groups (Z) are treated as separate sources at their respective AUs.

## Error handling and edge cases

### Defensive defaults for missing inputs

| Missing input                        | Behavior                                            |
| ------------------------------------ | --------------------------------------------------- |
| `Atmosphere == nil` (non-HZ body)    | Tidal lock DM-2 (atmo > 2.5 bar) skipped. No error. |
| `Hydrographics == nil`               | Surface Feature Distribution skipped (returns nil). |
| Body has no moons                    | Planet→moon case dropped from DM evaluation.        |
| Body is not a moon                   | Moon→planet case dropped.                           |
| Companion star with `MassSolar == 0` | Treated as no contribution.                         |

### Boundary clamping

| Table                       | Below floor                           | Above ceiling                                                         |
| --------------------------- | ------------------------------------- | --------------------------------------------------------------------- |
| Surface Distribution (2D-2) | -2/-1 → row 0 ("Extremely Dispersed") | 11+ → row 10 ("Extremely Concentrated")                               |
| Basic Day Length            | n/a (sum always ≥ 3)                  | 40+ → reroll cascade per book                                         |
| Basic Axial Tilt            | n/a                                   | 10+ → Extreme table                                                   |
| Tidal Lock Status           | DM ≤ -10 → skip case                  | DM ≥ +10 → automatic 1:1 lock; still roll for natural-12 verification |

### Edge cases

1. **Slow rotators / negative year-days.** When sidereal day approaches or exceeds year length, `YearDays = year_h / sidereal_h - 1` can be ≤ 0. Per p.104: retrograde rotators have negative sidereal day. Encode: store negative SiderealHours for retrograde post-lock cases, allow YearDays ≤ 0; document that solar-day computation in those cases yields very large or undefined values.
2. **Tidal-locked-to-star = twilight zone world.** `IsTwilightZone = true` only when `Case == PlanetToStar AND LockRatio == "1:1"`. SolarHours = 0 (sentinel: undefined per p.107).
3. **3:2 lock arithmetic.** SiderealHours = OrbitalPeriodHours × 2/3 (per p.105 "spins three times for every two rotations").
4. **Natural-12-on-1:1-lock branch (p.105 footnote).** Result ≥ 12 → roll 2D verification; if natural 12, reroll TidalLockStatus with DM=0 and apply that result instead. Encoded in `ApplyTidalLockEffect`.
5. **Moon-locked-to-planet 1:1 with day > parent orbit (p.105 footnote).** If 1:1 fires and post-lock day length > parent orbit period, force 1:1.
6. **Provisional pressure feeds tidal lock DM.** 3A1's `Atmosphere.Pressure` is provisional under HZCO temperature. After 3A2b corrects pressure, tidal-lock DM may shift across 2.5-bar threshold. **3A2a accepts this**; 3A2b carry-forward.
7. **Eccentricity reroll on 1:1 lock (p.105 footnote).** DM-2 reroll, take min of original / new.
8. **Axial-tilt reroll on 3:2 or 1:1 lock (p.105).** Note: applies to both ratios. Reroll as `(2D-2)/10` degrees (2D-2 ranges 0–10, divided by 10 → 0.0 to 1.0°). Result is always non-negative.
9. **Single-star vs binary-primary mass summing for star tidal.** Sum masses of all stars in the same `stars` group; treat each group as a single source. Zed Aab (group A) sums Aa + Ab → 1.836 M☉ as one source; Z (separate group) treated independently.

### Book inconsistencies / referee fiat to log

`t.Logf` in Zed acceptance test:

- **p.101 continent counts.** Per Q6, not encoded. Log: "Book example would derive ~2 major + ~9 minor continents from this surface distribution + hydro percent; deferred to Referee fiat."
- **p.106 tidal lock walkthrough.** Book says "(a fudged by Referee) further roll of 12 results in a 'DM free' roll." Implementation reproduces this exactly when fed the scripted dice list; log notes that the natural-12 verification is in the book.
- **Aab Z star tidal.** Book reports 0.24m total contribution from "the two relatively distant suns" using summed mass 1.836; if Z is treated as separate, contribution is much smaller. Log clarifies the close-binary sum heuristic.

## Testing and acceptance gate

### Test files (one per new file)

| File                            | Coverage                                                                                                                                                                                                                    |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `surface_distribution_test.go`  | Table-row mapping; Hydro=5 1D rule; boundary clamping (-2/-1 → 0, 11+ → A)                                                                                                                                                  |
| `day_length_test.go`            | Basic formula, GG/S × 2 modifier, 40+ reroll cascade, year-days arithmetic, slow-rotator negative YearDays                                                                                                                  |
| `axial_tilt_test.go`            | Basic table 2-9, Extreme table 1-6, retrograde detection, baseline preservation                                                                                                                                             |
| `tidal_lock_test.go`            | DM stack per case, case selection, all result branches (2-, 3-6, 7-8, 9-10, 11, 12+), natural-12 verification reroll, 3:2 axial-tilt reroll, 1:1 axial-tilt + ecc reroll, automatic lock at DM ≥ +10, skip-case at DM ≤ -10 |
| `surface_tidal_effects_test.go` | All four formulas, multi-star summing, multi-moon summing, locked-moon planet-tide skip, Zed Aab close-binary mass-summing heuristic                                                                                        |

### Composite acceptance test

`TestZed_FullDetail_3A2a` replaces `TestZed_FullDetail_3A1` (deleted). Single scripted dice list driving full `DetailSystem`, with sub-tests:

```go
func TestZed_FullDetail_3A2a(t *testing.T) {
    dice := buildZedScriptedDice() // ~80 dice across all phases
    r := roller.NewScripted(dice...)
    sd, err := DetailSystem(r, zedSystem(), zedPlacement(), zedHeader())
    require.NoError(t, err)

    // 3A1 carry-forward sub-tests (preserved unchanged)
    t.Run("body_physical_zed_prime", ...)
    t.Run("body_physical_aab_PI", ...)
    t.Run("atmosphere_zed_prime", ...)
    t.Run("atmosphere_subtype_aab_I", ...)
    t.Run("hydrographics_zed_prime", ...)
    t.Run("belt_aab_belt_2", ...)
    t.Run("moon_refinement_zed_prime_orbit", ...)

    // 3A2a sub-tests
    t.Run("surface_distribution_zed_prime", func(t *testing.T) {
        sd := zedPrimeOf(sd).SurfaceDistribution
        require.NotNil(t, sd)
        require.Equal(t, "Mixed", sd.Description)
        require.Equal(t, GeographyOcean, sd.Geography)
    })
    t.Run("day_length_zed_prime", func(t *testing.T) {
        dl := zedPrimeOf(sd).DayLength
        require.InDelta(t, 42.37, dl.BaselineSiderealHours, 0.01)
        require.InDelta(t, 84.74, dl.SiderealHours, 0.01)
        require.InDelta(t, 85.77, dl.SolarHours, 0.01)
        require.InDelta(t, 82.27, dl.YearDays, 0.05)
    })
    t.Run("axial_tilt_zed_prime", func(t *testing.T) {
        at := zedPrimeOf(sd).AxialTilt
        require.InDelta(t, 73.65, at.BaselineDegrees, 0.5)
        // Verification reroll broke the 1:1 lock per p.106 — tilt unchanged at 73.65°.
        require.InDelta(t, 73.65, at.Degrees, 0.5)
    })
    t.Run("tidal_lock_zed_prime", func(t *testing.T) {
        tl := zedPrimeOf(sd).TidalLock
        require.Equal(t, TidalLockCaseMoonToPlanet, tl.Case)
        require.Equal(t, 13, tl.InitialResult)              // p.106 "rolling 2D gets 6 + 7 = 13"
        require.True(t, tl.VerificationFired)            // p.106 "fudged further roll of 12"
        require.Equal(t, 4, tl.FinalResult)                 // p.106 "DM free roll ... result is 4"
        require.Equal(t, "", tl.LockRatio)                  // verification broke the 1:1 lock
        require.InDelta(t, 2.0, tl.DayLengthMultiplier, 0.001) // result 4 → day × 2
        require.False(t, tl.AxialTiltMutated)               // no lock → no tilt mutation
        require.False(t, tl.EccentricityMutated)            // no lock → no ecc mutation
        require.False(t, tl.IsTwilightZone)                 // moon→planet, not star
    })
    t.Run("tidal_effects_zed_prime", func(t *testing.T) {
        te := zedPrimeOf(sd).TidalEffects
        require.InDelta(t, 30.6, componentByPrefix(te, "planet").Meters, 0.2)
        require.InDelta(t, 0.24, componentByPrefix(te, "star").Meters, 0.05)
    })

    // Property invariants (carry forward from 3A1)
    t.Run("every_HZ_body_has_full_3A2a_data", ...)
    t.Run("no_nil_pointer_panics_100_iterations", ...)

    // Referee-fiat / book-inconsistency logs
    t.Logf("p.101 continent counts deferred to Referee fiat per Q6 option (b)")
    t.Logf("p.106 tidal lock natural-12 verification reproduced from scripted list")
}
```

### Synthetic Pluto/Charon test

`TestPlanetToMoon_PlutoCharon` builds a minimal SystemDetail with a Size-3 planet and a Size-1 moon at close orbit (~5 PD), asserts that case 3 (planet→moon) fires after the moon is locked. Exercises the only path the Zed example doesn't.

### Property tests

- **No nil panics across 100 free-dice iterations.** Random multi-star systems → `DetailSystem` should not panic, no errors, every HZ body has non-nil 3A2a data.
- **3A1 numeric outputs preserved.** Body physical, atmosphere, hydrographics for all bodies still match 3A1 expected values (eccentricity-mutation accepted as the only post-3A1 drift).

### Scripted dice list management

Continues 3A1 pattern: comment-annotated slices grouped by phase, concatenated:

```go
diceBodyPhysical := []int{...}        // 3A1
diceBeltDetails := []int{...}         // 3A1
diceAtmosphere := []int{...}          // 3A1
diceHydrographics := []int{...}       // 3A1
diceMoonRefinement := []int{...}      // 3A1
diceSurfaceDistribution := []int{...} // 3A2a
diceDayLength := []int{...}           // 3A2a
diceAxialTilt := []int{...}           // 3A2a
diceTidalLock := []int{...}           // 3A2a (includes natural-12 verification + DM-free reroll)

dice := slices.Concat(
    diceBodyPhysical, diceBeltDetails, diceAtmosphere, diceHydrographics,
    diceMoonRefinement, diceSurfaceDistribution, diceDayLength, diceAxialTilt,
    diceTidalLock,
)
```

Each slice is sourced directly from book pages with line comments back to WBH page numbers.

## Carry-forwards from 3A2a to 3A2b

- `Atmosphere.Pressure`, `Atmosphere.ScaleHeight`, `Hydrographics.Code` re-derivation under real temperature.
- `tempRangeMidpointK` placeholder replacement.
- Hydrographics Profile chemical-formula tail (`XX-##:YY-##`) and Optional Exotic Liquids selection.
- Tidal-lock re-evaluation if 3A2b's pressure correction crosses the 2.5-bar threshold.
- Tidal heating contribution to mean temperature (3A2b temperature equation).

## Carry-forwards deferred beyond 3A2

- Significant rings (R0# notation, removed-moon promotion) → 3B or later.
- Continent counts → Referee fiat (no plan).
- Moon-to-moon tidal effects → Referee fiat (no plan).
- HZ-candidate atmosphere subtype follow-ups → 3B.
- Seismology → 3B (location to be confirmed when 3B begins).

## Open questions

None at present. All scope and architectural questions resolved during brainstorm.
