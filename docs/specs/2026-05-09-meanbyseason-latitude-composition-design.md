# MeanBySeason Latitude Composition — Design

**Date:** 2026-05-09
**Sub-project:** 3A2b-temp follow-up
**Predecessors:** 3A2b-temp (`main`)
**Closes:** issue #4

## Goal

`Temperature.MeanBySeason(latDeg, daysSinceSolstice, localYearDays)` currently returns the same value for every `latDeg` because the seasonal axial-tilt swing is applied as a global luminosity modifier without being composed with the latitude zone formula. The fix: compose seasonal axial tilt with the existing `zoneTiltAdjustment(latDeg)` per WBH p.116 — tropical zone uses only the zone adjustment (no seasonal swing); middle/arctic zone adds seasonal axial tilt to the zone adjustment.

The function's doc-comment already flags this limitation: "the seasonal swing is uniform across latitudes for a given date."

## Source of truth

WBH p.116 Temperature Scenario: Mean Temperature by Latitude. Three latitude zones based on absolute latitude vs axial tilt:

1. **Tropical zone** — `|lat| ≤ axial_tilt`. The sun is directly overhead at least part of the year.
2. **Arctic zone** — `|lat| > 90° − axial_tilt`. The sun is below the horizon for at least part of the year.
3. **Middle (temperate) zone** — everything in between.

Zone latitude adjustment formulas:

- Tropical: `sin(45° − axial_tilt)` (constant year-round)
- Middle/Arctic: `sin(45° − latitude)` (varies with latitude)

For Part B (axial tilt ≥ 45°): no middle zone; the latitude luminosity adjustment for the gap region is set to the result at the edge of the arctic zone (`90° − tilt`).

The existing helper `zoneTiltAdjustment(latDeg)` in `worlds/temperature.go` implements the three-zone classification correctly for Part A (axial tilt < 45°). `MeanByLatitude` uses it for annual-mean temperatures. The Part B branch (tilt ≥ 45°) of `zoneTiltAdjustment` has a pre-existing correctness gap — case 1 of its switch fires for any `lat ≤ tilt` and clamps the negative `45 − tilt` argument to 0 instead of producing the book's `sin(tilt − 45)` for the inner equatorial-tropical region. Fixing that is out of scope here and tracked separately as issue #30; this PR's `MeanBySeason` change scopes its no-seasonal-swing branch to Part A only so that high-tilt worlds preserve their pre-existing behavior pending the Part B fix.

WBH p.116, the key passage that prescribes seasonal composition (this is what the current `MeanBySeason` ignores):

> In the tropical zone, this modifier replaces any axial tilt factor to modify luminosity when calculating any temperature for a specific latitude. For the rest of the world, it also replaces the axial tilt factor when computing a mean annual temperature for the latitude but to determine temperatures at a specific time of the year, the zone's latitude adjustment is added to the axial tilt factor for that time period and any other factors before dividing by the atmospheric factor to determine the mean temperature at that time of the year.

WBH p.116, tropical zone is constant year-round:

> If estimating varying temperatures during the course of the year in the tropical zone, an axial tilt factor modifier is not applied to the temperature equation.

So:

- **Tropical zone, any season:** `lumMod = zoneTropicalAdj / atmFactor` (no seasonal swing).
- **Middle/arctic zone, specific season:** `lumMod = (zoneMidArcticAdj + seasonalAxialTilt) / atmFactor`.
- **Middle/arctic zone, annual mean:** `lumMod = zoneMidArcticAdj / atmFactor` (already correct in `MeanByLatitude`; equivalent to `MeanBySeason` averaged over the year).

## Decisions

### Tropical zone has no seasonal swing

The book explicitly says tropical zone is constant year-round. The fix returns the same value for every `daysSinceSolstice` when `|latDeg| ≤ axial_tilt`. This is the WBH-prescribed behavior and matches the book's "tropical temperatures have little seasonal variation" note.

### Middle/arctic composition is additive

The book says "the zone's latitude adjustment is added to the axial tilt factor for that time period." Implementation: `variance = zoneTiltAdjustment(latDeg) + seasonalAxialTilt`. The combined value is then divided by atmospheric factor and clamped to [-1, 1].

### Helper extraction

`zoneTiltAdjustment` already computes the tilt-degree threshold logic. To avoid duplicating "is this latitude tropical?" inside `MeanBySeason`, extract a small private helper on `Temperature`:

```go
func (t *Temperature) tropicalLatitudeBoundary() float64
```

returning the tilt degrees (with NaN/clamp handling). Both `zoneTiltAdjustment` and `MeanBySeason` consume it.

### Function signature unchanged

`MeanBySeason(latDeg, daysSinceSolstice, localYearDays float64) float64` stays the same. No callers exist outside the test suite (per the doc-comment) and `AtMoment` (which composes seasonal + hourly).

## Architecture

### Helper: `tropicalLatitudeBoundary`

Pure method on `Temperature` returning the tilt-degree value with NaN/range clamping, replacing the duplicated logic inside `zoneTiltAdjustment`:

```go
// tropicalLatitudeBoundary returns the absolute axial tilt in degrees
// (clamped to [0, 90]). Latitudes ≤ this are in the tropical zone per
// WBH p.116; latitudes > 90° − tilt are in the arctic zone.
func (t *Temperature) tropicalLatitudeBoundary() float64 {
    tiltDeg := math.Asin(t.AxialTiltFactor) * 180.0 / math.Pi
    if math.IsNaN(tiltDeg) {
        if t.AxialTiltFactor > 0 {
            return 90
        }
        return 0
    }
    return tiltDeg
}
```

`zoneTiltAdjustment` switches its tilt computation to call this helper. `MeanBySeason` calls it to detect the tropical-zone case.

### `MeanBySeason` rewrite

```go
func (t *Temperature) MeanBySeason(latDeg, daysSinceSolstice, localYearDays float64) float64 {
    if t.IsTwilight {
        return t.TwilightK
    }
    if localYearDays <= 0 {
        return t.MeanByLatitude(latDeg)
    }

    // Zone latitude adjustment from WBH p.116. Tropical zone uses only
    // this adjustment (no seasonal axial-tilt swing — the book is
    // explicit that "tropical temperatures have little seasonal variation").
    // Middle/arctic zones add the seasonal axial tilt to the zone
    // adjustment per the book's "the zone's latitude adjustment is added
    // to the axial tilt factor for that time period" instruction.
    zoneAdj := t.zoneTiltAdjustment(latDeg)
    absLat := math.Abs(latDeg)
    if absLat > 90 {
        absLat = 90
    }
    isTropical := absLat <= t.tropicalLatitudeBoundary()

    variance := zoneAdj
    if !isTropical {
        // Adjusted Fractional Year per WBH p.115.
        stdYearDays := 8766.0 / 24.0 // 365.25
        lagDays := 0.1 * math.Min(stdYearDays, localYearDays)
        adjFracYear := (daysSinceSolstice - 0.1*lagDays) / localYearDays
        seasonalTilt := math.Cos(adjFracYear*2*math.Pi) * t.AxialTiltFactor
        variance += seasonalTilt
    }

    lumMod := variance / t.AtmosphericFactor
    if lumMod > 1 {
        lumMod = 1
    }
    if lumMod < -1 {
        lumMod = -1
    }
    latLum := t.Luminosity * (1 + lumMod)
    if latLum < 0 {
        latLum = 0
    }
    return MeanTemperatureK(latLum, t.Albedo, t.GreenhouseFactor, t.AU)
}
```

The doc-comment loses the "KNOWN LIMITATION" paragraph and gains a description of the WBH-correct composition.

## Out of scope

- **Twilight-zone hemisphere selection in `MeanBySeason` / `AtMoment`.** Tracked separately as issue #5. Both methods continue to short-circuit to `TwilightK` for twilight worlds.
- **Geographic factor inclusion in MeanBySeason.** WBH p.114 lists `Variance Factors = Axial Tilt + Rotation + Geographic`. The current `MeanBySeason` already omits geographic (and rotation, which is correct since this is "ignoring time of day"). Adding geographic to the per-latitude variance is a separate enhancement; this PR keeps the existing scope.
- **Eccentricity-driven seasonal effects.** WBH p.116 mentions: "tropical temperatures have little seasonal variation (from axial tilt, although a large eccentricity could provide season-like effects)." Eccentricity-based seasonal modulation is out of scope.
- **`AtMoment` direct changes.** `AtMoment` calls `MeanBySeason` for the seasonal contribution; it inherits the latitude composition automatically. No direct changes to `AtMoment`.

## Testing strategy

### Unit tests for `MeanBySeason` latitude composition

Add to `worlds/temperature_test.go` (or create `worlds/temperature_seasonal_test.go` if the file is large):

- **`TestMeanBySeason_TropicalZone_NoSeasonalSwing`** — Earth-like body with axial tilt 23° at lat 0° (tropical). Solstice (day 0) and equinox (day year/4) produce the same temperature (within float tolerance) because tropical has no seasonal variation.
- **`TestMeanBySeason_MiddleZone_SeasonalSwing`** — Same body at lat 60° (middle zone). Solstice (summer) is meaningfully warmer than equinox; equinox warmer than winter solstice.
- **`TestMeanBySeason_LatitudesProduceDifferentTemps`** — Same date (e.g., summer solstice) at lat 0° vs lat 60°: temperatures differ. The current bug is exactly this — they're identical. The new test fails on current code, passes on fix.
- **`TestMeanBySeason_ArcticZone_HighTilt_PartB`** — High-tilt body (axial tilt ≥ 45°) at the gap-region latitude (between tilt and 90°-tilt). Verifies Part B fallback uses arctic-edge result.
- **`TestMeanBySeason_NoYearLength_FallsBackToMeanByLatitude`** — `localYearDays == 0` returns the same as `MeanByLatitude(latDeg)`. Existing behavior; sanity test.

### Existing test impact

`AtMoment` calls `MeanBySeason`. Existing `AtMoment` tests likely passed despite the bug because they used a single representative latitude that happened to produce reasonable values. After the fix, latitude meaningfully affects results — re-run existing tests; if any break, they need updating to reflect the corrected book behavior.

### Zed golden

No production caller of `MeanBySeason` exists per the function's own doc-comment, so the Zed golden should not shift. If it does, that's a regression — escalate.

## Closes

#4.
