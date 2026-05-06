# World Physical 3B-Biology — Design

**Date:** 2026-05-05
**Sub-project:** 3B-biology (second sub-project of 3B; pp.127-131)
**Predecessor:** 3B-geology — merged on `main` as `2ebced4`

## Goal

Implement WBH pp.127-131: Native Lifeforms framework, Biomass Rating, Biocomplexity Rating, Native Sophonts (extant + extinct), Biodiversity Rating, Compatibility Rating, Native Lifeform Profile (MXDC eHex), and Resource Rating. Add a single new pipeline step `runStep5F` between 3B-geology and Step 6.

This is the **first sub-project after the temperature feedback edge** (3B-geology) — pure forward flow. Biology consumes finalized atm / hydro / temperature data; produces nothing the upstream pipeline reads.

## Brainstorm decisions

| Q                               | Decision                                                                                |
| ------------------------------- | --------------------------------------------------------------------------------------- |
| Q1: Scope                       | (a) Single sub-project covering all 6 ratings + Sophont bools + Profile method          |
| Q2: Body filter                 | (a) Terrestrials + HZ-planet moons (with Atmosphere); skip GGs / belts / empty          |
| Q3: Biologic-taint special case | (a) Defer entirely; document as carry-forward                                           |
| Q5: Acceptance test naming      | (b) Keep `TestZed_FullDetail_3A2b`; append assertions 32-38 (precedent from 3B-geology) |

## Architecture

### New file: `worlds/biology.go`

Holds the `Biology` struct + 7 standalone helper functions + the `Profile()` method:

- `RollBiomass(r, body) int` — `2D + DMs` capped to [-12, +4] modifier sum; depends on atm code, hydro code, age, mean/high temp; applies exotic-atm bonus when biomass ≥ 1 on atm 0/1/A/B/C/F+
- `RollBiocomplexity(r, body, biomass) int` — `2D − 7 + min(biomass, 9) + DMs`; result clamped to ≥ 1 when biomass > 0; prerequisite: biomass > 0
- `RollNativeSophont(r, biocomplexity) bool` — extant: `2D + min(biocomplexity, 9) − 7 ≥ 13`; prerequisite: biocomplexity ≥ 8
- `RollExtinctSophont(r, body, biocomplexity, sysAgeGyr) bool` — `2D + min(biocomplexity, 9) − 7 + DMs ≥ 13`; DM+1 if age > 5 Gyrs
- `RollBiodiversity(r, biomass, biocomplexity) int` — `ceil(2D − 7 + (biomass + biocomplexity) / 2)`; result clamped to ≥ 1; prerequisite: biomass > 0
- `RollCompatibility(r, body, biocomplexity) int` — `floor(2D − biocomplexity/2 + DMs)`; result clamped to ≥ 0; prerequisite: biomass > 0
- `RollResourceRating(r, body, biology) int` — `2D − 7 + Size + DMs` clamped to [2, 12]; runs for ALL terrestrial bodies (with or without biology)

### New step: `runStep5F(r, detailed, sys) error` in `worlds/system_detail_steps.go`

Slots between Step 5E (3B-geology) and Step 6 (BaselineN backfill). Single pass through `detailed`; per-body, per-moon. Mutates `dp.Biology *Biology` in place.

`DetailSystem` orchestrator picks up one new line:

```go
// Step 5F — 3B-biology pass: native lifeform ratings + resource rating.
if err := runStep5F(r, detailed, sys); err != nil {
    return SystemDetail{}, err
}
```

### Struct extensions

`DetailedPlacement` and `Moon` each get one new pointer field:

```go
Biology *Biology
```

Accessors:

```go
func (dp *DetailedPlacement) HasBiology() bool { return dp.Biology != nil }
func (m *Moon) HasBiology() bool               { return m.Biology != nil }
```

## Public API

### The `Biology` struct

```go
// Biology — native lifeform ratings + resource rating per WBH pp.127-131.
// Populated by Step 5F for terrestrial bodies (and their HZ-planet moons)
// that have Atmosphere data.
//
// Conditional applicability:
//   - Bodies with Biomass == 0: only ResourceRating populated; biology
//     ratings (Biocomplexity, Biodiversity, Compatibility) stay 0; sophont
//     bools stay false.
//   - Bodies with Biomass >= 1 but Biocomplexity < 8: sophont bools stay
//     false (prerequisite for sophont rolls is Biocomplexity >= 8).
//   - Belts (Size 0), gas giants, empty placements: biology not generated;
//     dp.Biology stays nil.
type Biology struct {
    // 2D + DMs, with combined-DM sum clamped to [-12, +4] per WBH p.127.
    // Range 0-15 (eHex 0-F); 0 = no native life.
    Biomass int

    // 2D - 7 + Biomass + DMs per WBH p.129. Zero if Biomass == 0.
    // Result < 1 promoted to 1 (when biomass > 0). Range 0-15.
    Biocomplexity int

    // True if extant native sophont species exists; 2D + Biocomplexity - 7
    // >= 13 per WBH p.130. False if Biocomplexity < 8.
    HasNativeSophont bool

    // True if evidence of an extinct native sophont species; 2D + Biocomplexity
    // - 7 + DMs >= 13 per WBH p.130. False if Biocomplexity < 8. Independent
    // of HasNativeSophont — both can be true.
    HadExtinctSophont bool

    // ceil(2D - 7 + (Biomass + Biocomplexity) / 2) per WBH p.130.
    // Zero if Biomass == 0. Result < 1 promoted to 1 (when biomass > 0).
    Biodiversity int

    // floor(2D - Biocomplexity/2 + DMs) per WBH p.130-131. Zero if Biomass == 0
    // or if rolled result <= 0. Range 0-15+ (10 = full Terran, > 10 possible).
    Compatibility int

    // 2D - 7 + Size + DMs per WBH p.131. Computed for ALL terrestrial
    // bodies regardless of biology (biology DMs only apply when applicable).
    // Range [2, 12] per WBH lower/upper bounds.
    ResourceRating int
}

// Profile returns the WBH p.131 native-lifeform-profile MXDC eHex string
// (Biomass / Biocomplexity / Biodiversity / Compatibility). Returns ""
// when Biomass is 0 (no native life to profile).
func (b *Biology) Profile() string { ... }
```

### `runStep5F` signature

```go
func runStep5F(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error
```

Same shape as `runStep5C`/`runStep5D`/`runStep5E`. Moons processed inside the loop using `buildMoonPlacementView`.

## Procedure (per body)

### Step 1 — Biomass Rating (WBH p.127-128)

```
Biomass = 2D + clamp(SumOfDMs, -12, +4)
```

DMs (sum then clamp):

| Atm code | DM  |
| -------- | --- |
| 0        | −6  |
| 1        | −4  |
| 2, 3, E  | −3  |
| 4, 5     | −2  |
| 6, 7     | 0   |
| 8, 9, D  | +2  |
| A        | −3  |
| B        | −5  |
| C        | −7  |
| F+       | −5  |

| Hydro code | DM  |
| ---------- | --- |
| 0          | −4  |
| 1−3        | −2  |
| 4−5        | 0   |
| 6−8        | +1  |
| 9, A       | +2  |

| Age (Gyrs) | DM  |
| ---------- | --- |
| < 0.2      | −6  |
| < 1        | −2  |
| > 4        | +1  |

| Temperature condition (Kelvin) | DM  |
| ------------------------------ | --- |
| HighK > 353                    | −2  |
| HighK < 273                    | −4  |
| MeanK > 353                    | −4  |
| MeanK < 273                    | −2  |
| MeanK in [279, 303]            | +2  |

**Note**: WBH provides a fallback for "if and only if detailed temperature was not determined" (DM+2 temperate / DM−2 cold / DM−6 boiling-or-frozen). Our pipeline always computes detailed temperature post-3A2b, so the fallback is never used. Document but don't implement.

**Exotic-atm bonus** (WBH "Special Case 2"): if rolled biomass ≥ 1 and atm ∈ {0, 1, A, B, C, F+}, add `(|atm DM| − 1)` to biomass:

| Atm | Bonus |
| --- | ----- |
| 0   | +5    |
| 1   | +3    |
| A   | +2    |
| B   | +4    |
| C   | +6    |
| F+  | +4    |

**Skipped** (per Q3-a): Special Case 1 (biologic-taint promotion of biomass=0 worlds). Defer pending Atmosphere taint typology.

**Worked: Zed Prime** — Atm 6 (no DM), Hydro 6 (+1), Age 6.3 (+1), MeanK 300 (+2), HighK 346 (no DM since 273 ≤ 346 ≤ 353): DMs total +4 (at cap), 2D=6 → 6+4=10 → biomass A.

### Step 2 — Biocomplexity Rating (WBH p.129)

**Prerequisite:** Biomass > 0. Otherwise: Biocomplexity = 0.

```
Biocomplexity = 2D − 7 + min(Biomass, 9) + DMs
```

DMs:

| Condition          | DM                                           |
| ------------------ | -------------------------------------------- |
| Atmosphere not 4−9 | −2                                           |
| Low oxygen taint   | DEFERRED (Q3-a — taint typology unavailable) |
| Age 3−4 Gyrs       | −2                                           |
| Age 2−3 Gyrs       | −4                                           |
| Age 1−2 Gyrs       | −8                                           |
| Age < 1 Gyr        | −10                                          |

**Edge boundary**: "If the system age is exactly at a limit between two DMs, use the worst DM." Implementation as ordered cases (each upper bound inclusive — exact-boundary ages fall into the lower-DM band):

```go
switch {
case age <= 1:
    return -10
case age <= 2:
    return -8
case age <= 3:
    return -4
case age <= 4:
    return -2
default:
    return 0
}
```

This produces: age=1.0 → −10 (worst of −10/−8); age=2.0 → −8 (worst of −8/−4); age=3.0 → −4 (worst of −4/−2); age=4.0 → −2 (worst of −2/0); age>4 → 0.

Result clamped to ≥ 1 when biomass > 0.

**Worked: Zed Prime** — Biomass A → 9 for the roll, Age 6.3 (no age DM since > 4), Atm 6 (in 4-9, no DM): 2D − 7 + 9 = 2D + 2. Roll 3 → 5.

### Step 3 — Native Sophont Rolls (WBH p.130)

**Prerequisite:** Biocomplexity ≥ 8. Otherwise both bools stay false (no dice consumed).

**Extant Native Sophont** exists if `2D + min(Biocomplexity, 9) − 7 ≥ 13`.

**Extinct Native Sophont** existed if `2D + min(Biocomplexity, 9) − 7 + DMs ≥ 13`. DM+1 if Age > 5 Gyrs (extinct only).

Independent rolls — both can be true, both can be false.

**Worked: Zed Prime** — Biocomplexity 5 < 8 → no rolls; both bools false.

### Step 4 — Biodiversity Rating (WBH p.130)

**Prerequisite:** Biomass > 0. Otherwise: Biodiversity = 0.

```
Biodiversity = ceil(2D − 7 + (Biomass + Biocomplexity) / 2)
```

The `(B + X) / 2` is float division; the final `ceil` rounds the entire expression. Result clamped to ≥ 1.

**Worked: Zed Prime** — Biomass 10, Biocomplexity 5 → `(10 + 5) / 2 = 7.5`. Roll 6 → `6 − 7 + 7.5 = 6.5` → ceil → 7.

### Step 5 — Compatibility Rating (WBH p.130-131)

**Prerequisite:** Biomass > 0. Otherwise: Compatibility = 0.

```
Compatibility = floor(2D − Biocomplexity/2 + DMs)
```

DMs:

| Condition             | DM  |
| --------------------- | --- |
| Atmosphere 0, 1, B    | −8  |
| Atmosphere 2, 4, 7, 9 | −2  |
| Atmosphere 3, 5, 8    | +1  |
| Atmosphere 6          | +2  |
| Atmosphere A or F     | −6  |
| Atmosphere C          | −10 |
| Atmosphere D, E       | −1  |
| Age > 8 Gyrs          | −2  |

Atm codes G/H mentioned in the book's table are not produced by `RollAtmoCode` (Mongoose Traveller atm range is 0-F). They cannot appear in our pipeline — no DM applied.

If `floor(...) ≤ 0`, Compatibility = 0.

**Worked per formula: Zed Prime** — Atm 6 (+2), Biocomplexity 5, 2D=7 → `7 − 2.5 + 2 = 6.5` → floor → **6**.

**Book inconsistency**: WBH p.131 worked example shows `7 + 3 − 2.5 + 2 = 9.5` → 9, with no source for the "+3" addend. Implementation follows the formula box → Zed Prime compatibility = 6 (not 9). Logged as feedback memory after merge.

### Step 6 — Native Lifeform Profile (WBH p.131)

Method `(b *Biology) Profile() string` returns 4-char eHex MXDC string from `Biomass`, `Biocomplexity`, `Biodiversity`, `Compatibility`. Returns `""` when Biomass == 0.

eHex encoding: 0-9 → "0"-"9"; 10-15 → "A"-"F"; values > 15 saturate to "F".

**Worked: Zed Prime per formula** — `Profile() = "A576"` (Biomass=10/A, Biocomplexity=5, Biodiversity=7, Compatibility=6). Book's worked profile is "A579" — divergence stems from the Compatibility book inconsistency above.

### Step 7 — Resource Rating (WBH p.131)

Runs for ALL terrestrial bodies — with or without biology.

```
ResourceRating = 2D − 7 + Size + DMs
```

DMs:

| Condition         | DM                       |
| ----------------- | ------------------------ |
| Density > 1.12    | +2                       |
| Density < 0.5     | −2                       |
| Biomass 3+        | +2                       |
| Biodiversity 8−A  | +1                       |
| Biodiversity B+   | +2                       |
| Compatibility 0−3 | −1 (only if Biomass ≥ 1) |
| Compatibility 8+  | +2                       |

Final clamp: `max(2, min(result, 12))`.

### Order of operations within `runStep5F`

For each non-empty placement:

1. Skip belts (`SizeCode == "0"`), gas giants (`Body == BodyGasGiant`), empty bodies.
2. Skip terrestrials without atmosphere data (`Atmosphere == nil`) — no biomass DM lookup possible.
3. Initialize `bio := &Biology{}`.
4. Roll Biomass.
5. If Biomass > 0:
   1. Roll Biocomplexity.
   2. If Biocomplexity ≥ 8: Roll Native Sophont (extant); Roll Extinct Sophont.
   3. Roll Biodiversity.
   4. Roll Compatibility.
6. Roll Resource Rating (always, even when Biomass = 0).
7. Assign `dp.Biology = bio`.
8. Recurse into moons that have Atmosphere; same per-moon body-filter logic.

**Per-body dice budget**: up to 7 × 2D when life is present; 2 × 2D when biomass = 0 (Biomass + Resource); 0 × 2D when body is skipped. Same per moon.

## Sub-decisions

- **Mutation policy**: `dp.Biology = &Biology{...}` (pointer assignment), consistent with 3B-geology's `dp.Geology = &Geology{...}`.
- **Atm 6 and 7**: not in the Biomass DM table → DM 0. (Atm 6 = Standard, Atm 7 = Standard Tainted.)
- **Hydrographics 4-5**: not in the Biomass DM table → DM 0.
- **Biomass and Biocomplexity clamps**: per WBH "ratings above 9 are treated as 9 for this roll" — applied via `min(value, 9)` at each downstream consumer call site.
- **Round-up for Biodiversity**: `ceil` applied to the final expression, not the intermediate `(B+X)/2`.
- **Round-down for Compatibility**: `floor` applied to the final expression. Negative results clamped to 0.
- **Sophont rolls independent**: both can be true (extant species with extinct ancestors).
- **Resource Rating runs even when Biomass = 0**: per WBH "for all other worlds, the Size of the world is the main determinant"; biology DMs simply don't fire when ratings are 0.
- **Moon biology** uses `buildMoonPlacementView` synthetic moonDP. Atmosphere/Hydrographics/Physical/Temperature are pointer-aliased from the moon (handled by buildMoonPlacementView from prior sub-projects).
- **Belt biology**: skipped. Resource analog already lives in `BeltDetails` (3A1 path).
- **GG biology**: skipped. The biomass DM table doesn't apply to GG atmospheres.
- **Empty Atmosphere on terrestrial body**: skipped. The atm DM lookup needs a code; nil means biomass DM table can't be consulted.
- **Empty Hydrographics on terrestrial body**: defensive — treat as Hydro 0 (DM−4). Per book, atm 0 / 1 worlds may not have hydrographics rolled but biomass should still compute.
- **Empty Temperature on terrestrial body**: defensive — no temperature DMs apply (treat as if temp data wasn't computed; book's fallback isn't reached because pipeline always computes temp).

## Documented WBH inconsistencies (carry-forward feedback memories)

After merge, save:

1. **WBH p.131 Compatibility worked example contradicts formula box.** Formula says `2D − Biocomplexity/2 + DMs`; worked example shows `7 + 3 − 2.5 + 2 = 9.5`. Implementation follows the formula → Zed Prime compatibility = 6 (not 9). Profile becomes "A576" (not "A579").
2. **Atmosphere codes G and H** appear in the Compatibility DM table (p.131) but don't exist in the WBH atmosphere code system (0-F). Implementation ignores them — they cannot be produced by `RollAtmoCode`. May be remnant from Cepheus Engine or earlier supplement.

## Testing

### Per-formula unit tests (`worlds/biology_test.go`)

**RollBiomass:**

- `TestRollBiomass_ZedPrime` — Atm 6, Hydro 6, Age 6.3, MeanK=300, HighK=346, 2D=6 → 10
- `TestRollBiomass_DMCap_AtPositiveCeiling` — many positive DMs sum > +4 → clamped to +4
- `TestRollBiomass_DMCap_AtNegativeFloor` — Vacuum atm + Frozen + Young → clamped to −12
- `TestRollBiomass_VacuumAtm_NegativeBaseline` — atm 0 (DM−6), 2D=12 → biomass 6 + bonus 5 = 11
- `TestRollBiomass_ExoticAtm_BonusApplied_AtmB` — atm B + rolled biomass ≥ 1 → adds 4
- `TestRollBiomass_ExoticAtm_BonusSkipped_AtmBZero` — atm B + rolled biomass 0 → bonus NOT applied (stays 0)
- `TestRollBiomass_NilAtmosphere_Zero` — defensive
- `TestRollBiomass_NilHydrographics_HydroZeroDM` — treats nil as Hydro 0
- `TestRollBiomass_NilTemperature_NoTempDMs` — defensive

**RollBiocomplexity:**

- `TestRollBiocomplexity_ZedPrime` — Biomass=10, Atm 6, Age 6.3, 2D=3 → 5
- `TestRollBiocomplexity_BiomassZero_Zero` — prerequisite fails, no dice consumed
- `TestRollBiocomplexity_BiomassClamp_Above9` — Biomass=15 used as 9
- `TestRollBiocomplexity_AgeBoundary_Exactly4_UsesWorseDM` — age=4.0 → DM−2 (in 3-4 band)
- `TestRollBiocomplexity_ResultLessThanOne_PromotedToOne`

**RollNativeSophont (extant):**

- `TestRollNativeSophont_BelowPrerequisite_False` — Biocomplexity=7, no dice
- `TestRollNativeSophont_Triggers_AtBiocomplexity9` — 2D=11 → 13 → true
- `TestRollNativeSophont_BelowThreshold` — 2D=11, Bio=8 → 12 < 13 → false
- `TestRollNativeSophont_BiocomplexityClamp_Above9` — Biocomplexity=15 used as 9

**RollExtinctSophont:**

- `TestRollExtinctSophont_BelowPrerequisite_False` — Biocomplexity=7, no dice
- `TestRollExtinctSophont_AgeOver5_DMPlusOne` — verify +1 affects threshold
- `TestRollExtinctSophont_AgeUnder5_NoDM` — verify +1 not applied

**RollBiodiversity:**

- `TestRollBiodiversity_ZedPrime` — Biomass=10, Biocomplexity=5, 2D=6 → 7
- `TestRollBiodiversity_BiomassZero_Zero` — prerequisite fails
- `TestRollBiodiversity_RoundsUp` — fractional intermediate → ceil
- `TestRollBiodiversity_ResultLessThanOne_PromotedToOne`

**RollCompatibility:**

- `TestRollCompatibility_ZedPrime_FollowsFormula` — Biocomplexity=5, Atm 6, 2D=7 → 6 (NOT 9 per book; documents WBH inconsistency inline)
- `TestRollCompatibility_BiomassZero_Zero` — prerequisite fails
- `TestRollCompatibility_NegativeResult_ClampedToZero` — high biocomplexity + bad atm → ≤ 0 → 0
- `TestRollCompatibility_AtmCRich_DMMinus10`
- `TestRollCompatibility_AgeOver8_DMMinus2`

**RollResourceRating:**

- `TestRollResourceRating_TerrestrialNoLife` — Biomass=0, no biology DMs
- `TestRollResourceRating_HighDensity_PlusTwo`
- `TestRollResourceRating_HighBiomass_PlusTwo`
- `TestRollResourceRating_HighBiodiversity_PlusOne_PlusTwo` — boundaries 8-A and B+
- `TestRollResourceRating_LowCompatibilityWithLife_MinusOne` — Compatibility 0-3 + Biomass ≥ 1
- `TestRollResourceRating_LowCompatibilityNoLife_NoDMSkipped` — Comp 0 but Biomass=0 → DM not applied
- `TestRollResourceRating_HighCompatibility_PlusTwo`
- `TestRollResourceRating_ResultBelowTwo_ClampedToTwo`
- `TestRollResourceRating_ResultAboveTwelve_ClampedToTwelve`

**Profile method:**

- `TestBiology_Profile_ZedPrime_A576` — biomass=10, biocomplexity=5, biodiversity=7, compatibility=6 → "A576"
- `TestBiology_Profile_NoLife_Empty` — biomass=0 → ""
- `TestBiology_Profile_eHexEncoding_AboveNine` — values 10-15 → A-F

### Orchestrator tests

- `TestRunStep5F_TerrestrialWithLife_PopulatesAll`
- `TestRunStep5F_TerrestrialNoLife_OnlyResource`
- `TestRunStep5F_GasGiant_NoBiology`
- `TestRunStep5F_BeltSize0_NoBiology`
- `TestRunStep5F_BodyEmpty_NoOp`
- `TestRunStep5F_MoonWithAtmosphere_GetsBiology`
- `TestRunStep5F_MoonNoAtmosphere_NoBiology`
- `TestRunStep5F_BiocomplexityBelowEight_NoSophontRolls`

### Acceptance test extension (`worlds/worked_examples_test.go`)

Append assertions 32-38 inside `TestZed_FullDetail_3A2b`'s iter loop (keep the function name per Q5-b):

- **32**: HasBiology() for terrestrial bodies with Atmosphere (and HZ-planet moons with Atmosphere); skip belts/GGs/empty
- **33**: When `Biomass > 0`: all biology fields populated (Biocomplexity > 0, Biodiversity > 0, Compatibility ≥ 0)
- **34**: When `Biomass == 0`: biology fields are 0; sophont bools false
- **35**: ResourceRating ∈ [2, 12] for all bodies with Biology
- **36**: Sophont bools: `HasNativeSophont || HadExtinctSophont` only when `Biocomplexity ≥ 8`
- **37**: For all bodies with Biology, `Profile()` returns either "" (no life) or a 4-char eHex string
- **38**: Across the 100-iter sweep, at least one body should have `Biomass ≥ 1` (smoke test against silent-zero bug, parallel to assertion 31's THF check)

Plus a 6th trailing `t.Logf` reporting count of bodies with non-zero biomass across the sweep.

## Carry-forwards (deferred to future sub-projects)

| What                                                                                  | Why                                                                                  |
| ------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Habitability rating, final mainworld determination, IISS Class IV-P form              | 3B-final's scope                                                                     |
| Biologic-taint special case (biomass=0 + biologic taint → biomass=1, biocomplexity=1) | Requires Atmosphere taint typology — out of scope                                    |
| Optional Rule (any oxygen atm = biomass ≥ 1)                                          | Referee toggle, not core procedure                                                   |
| Rare Earth Universe Variant (DM−2 on biocomplexity, restricted positive modifier)     | Referee toggle                                                                       |
| Detailed sophont generation                                                           | Sector Construction Guide p.50 — different supplement                                |
| Conditional follow-up                                                                 | GG biospheres (book doesn't address; airborne life would need a different framework) |
| Compatibility-rating-of-specific-biological-material                                  | "rating × 10% + 2D-7" formula on p.131 — derived value, not a stored field           |

## File map

| File                             | Status   | Purpose                                                                                           |
| -------------------------------- | -------- | ------------------------------------------------------------------------------------------------- |
| `worlds/biology.go`              | New      | `Biology` struct + 7 standalone helpers + `Profile()` method                                      |
| `worlds/biology_test.go`         | New      | Per-formula unit tests + orchestrator tests                                                       |
| `worlds/system_detail.go`        | Modified | Add `Biology *Biology` to `DetailedPlacement`; one new line in `DetailSystem` to call `runStep5F` |
| `worlds/system_detail_steps.go`  | Modified | Add `runStep5F` helper + `computeBiology`/`computeMoonBiology`                                    |
| `worlds/moons.go`                | Modified | Add `Biology *Biology` to `Moon` + `HasBiology()` accessor                                        |
| `worlds/worked_examples_test.go` | Modified | Append assertions 32-38 to `TestZed_FullDetail_3A2b` + 6th trailing t.Logf                        |

## Plan target

Target: 9-11 tasks, executable via subagent-driven-development with per-task subagent + spec/code review loops. Final end-to-end review on Opus before merge. Same workflow as 3B-geology.

Estimated task layout:

1. Branch + Biology struct + DetailedPlacement.Biology + Moon.Biology + HasBiology accessors
2. RollBiomass (most complex — large DM table + special-case bonus)
3. RollBiocomplexity
4. RollNativeSophont (extant)
5. RollExtinctSophont
6. RollBiodiversity
7. RollCompatibility
8. RollResourceRating
9. Biology.Profile method + runStep5F orchestrator + DetailSystem wiring
10. Acceptance test extension (assertions 32-38)
11. Final end-to-end review on Opus + merge
