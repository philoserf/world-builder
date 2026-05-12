# zoneTiltAdjustment Part B (axial tilt ≥ 45°) — Design

**Date:** 2026-05-09
**Issue:** #30
**WBH refs:** pp.116-117 (latitude composition, three-zone classification)

## Problem

`worlds/temperature.go::zoneTiltAdjustment` does not correctly handle WBH p.117 Part B (worlds with axial tilt ≥ 45°). Two related bugs:

### Bug 1 — `zoneTiltAdjustment` Part B unreachable

Current switch:

```go
switch {
case latDeg <= tiltDeg:
    // Tropical: sin(45° - tilt), clamped negative to 0.
case tiltDeg >= 45 && latDeg < (90-tiltDeg):
    // Part B inner band — UNREACHABLE.
default:
    // Middle/arctic: sin(45° - lat).
}
```

For `tilt=60°, lat=50°`: case 1 fires (50 ≤ 60), the argument `45-60 = -15` is clamped to 0, returns `sin(0)=0`. The correct WBH p.117 result is the arctic-zone formula `sin(45-50) ≈ -0.087` (mild cooling).

For `tilt=60°, lat=20°`: case 1 fires, returns 0. Correct Part B inner-band result is `sin(60-45) ≈ +0.259` (the warming from being inside the always-illuminated equatorial band of a high-tilt world).

Case 2 of the switch is unreachable because case 1 always matches first whenever `lat ≤ tilt`.

### Bug 2 — `MeanBySeason` `isTropical` uses wrong boundary for Part B

Added in #4:

```go
tiltDeg := t.tropicalLatitudeBoundary()
isTropical := tiltDeg < 45 && absLat <= tiltDeg
```

The `tiltDeg < 45` clause was a deliberate scope-limit added during the #4 PR-review fix (regression caught by Copilot). For Part B worlds the no-seasonal-swing region exists but ends at `(90 − tilt)`, not at `tilt` — so currently Part B worlds always apply the seasonal swing uniformly across all latitudes (their pre-#4 behavior). This is "not wrong" in the sense it doesn't crash, but it under-reports tropical-band suppression for high-tilt worlds.

## Reference — WBH p.117 Part B

> Apply the middle and arctic zone case (Case 2) for the entire arctic zone (90° − axial tilt).
>
> The latitude luminosity adjustment for the remaining equatorial tropical zone is equal to the result of middle and arctic zone case calculation at the edge of the arctic zone (90° − axial tilt).

So for tilt ≥ 45°:

| latitude band                                     | adjustment                                     |
| ------------------------------------------------- | ---------------------------------------------- |
| `\|lat\| ≤ 90 − tilt` (inner equatorial-tropical) | `sin(45 − (90 − tilt)) = sin(tilt − 45)`       |
| `\|lat\| > 90 − tilt` (arctic)                    | `sin(45 − lat)` (same as Part A middle/arctic) |

Boundary continuity check at `tilt = 45°`:

- Part A boundary: `tilt = 45°`. Tropical formula: `sin(45 − 45) = 0`.
- Part B boundary: `90 − 45 = 45°`. Inner formula: `sin(45 − 45) = 0`.

Continuous.

## Design

### Approach — modify `tropicalLatitudeBoundary` semantics

The helper is currently named for its Part A meaning (raw axial tilt). The book's Part B reorganization preserves the "no-seasonal-swing inner band" concept but moves its outer boundary to `(90 − tilt)`. The function's role in `MeanBySeason` is to identify _that_ boundary — which differs between Part A and Part B.

Update `tropicalLatitudeBoundary` to return the part-aware boundary:

- Part A (`tilt < 45`): return `tilt`.
- Part B (`tilt ≥ 45`): return `90 − tilt`.

This lets `MeanBySeason`'s `isTropical` check simplify to `absLat <= tropicalLatitudeBoundary()` for both parts (drop the `tiltDeg < 45` gate).

The doc comment on `tropicalLatitudeBoundary` updates to reflect the Part B branch.

### Refactor — `zoneTiltAdjustment`

Restructure the switch to test Part A vs Part B first, then split by latitude. This eliminates the unreachable case and makes both parts symmetric and clear:

```go
func (t *Temperature) zoneTiltAdjustment(latDeg float64) float64 {
    rawTilt := math.Asin(t.AxialTiltFactor) * 180.0 / math.Pi
    if math.IsNaN(rawTilt) {
        rawTilt = 90
    }
    if rawTilt < 0 {
        rawTilt = -rawTilt
    }

    if latDeg < 0 {
        latDeg = -latDeg
    }
    if latDeg > 90 {
        latDeg = 90
    }

    if rawTilt >= 45 {
        // WBH p.117 Part B: middle zone disappears.
        if latDeg <= 90.0-rawTilt {
            return math.Sin((rawTilt - 45.0) * math.Pi / 180.0)
        }
        return math.Sin((45.0 - latDeg) * math.Pi / 180.0)
    }

    // WBH p.116 Part A: tropical zone is |lat| ≤ tilt; rest is middle/arctic.
    if latDeg <= rawTilt {
        return math.Sin((45.0 - rawTilt) * math.Pi / 180.0)
    }
    return math.Sin((45.0 - latDeg) * math.Pi / 180.0)
}
```

Note: `zoneTiltAdjustment` reads the raw axial tilt directly (not via `tropicalLatitudeBoundary`) because it needs to distinguish Part A from Part B. `tropicalLatitudeBoundary` returns the _no-seasonal-swing boundary_ — a different concept that happens to coincide with raw tilt under Part A.

### `MeanBySeason` simplification

Once `tropicalLatitudeBoundary` returns the correct Part B boundary, the `isTropical` check simplifies:

```go
// Before:
isTropical := tiltDeg < 45 && absLat <= tiltDeg
// After:
isTropical := absLat <= t.tropicalLatitudeBoundary()
```

The Part-B-specific comment in `MeanBySeason` (lines 543-547) gets removed since the special case is no longer skipped.

## Out of Scope

- The book's Part B language is silent on whether the **rotation factor**, **geographic factor**, or **atmospheric factor** carry forward unchanged from Part A. The implementation does not modify them — the assumption inherited from the existing code is that those are latitude-and-zone-independent. This matches the Zed worked example (Part A) and is the cheapest assumption for Part B until a high-tilt worked example proves otherwise.
- Twilight worlds (`IsTwilight=true`) short-circuit at the top of `MeanBySeason` and never reach `zoneTiltAdjustment`. This change does not interact with twilight handling.
- No callers of `MeanByLatitude` / `MeanBySeason` exist in production paths (Zed golden does not exercise them). No `cmd/wbh` markdown output impact. The fix is library-only correctness.

## Tests

New tests in `worlds/temperature_test.go`:

1. `TestTemperature_ZoneTiltAdjustment_PartB_InnerBand` — tilt 60° (factor `sin(60°)`), lat 20° → expect `sin(15°) ≈ 0.2588` (was 0).
2. `TestTemperature_ZoneTiltAdjustment_PartB_ArcticZone` — tilt 60°, lat 50° → expect `sin(-5°) ≈ -0.0872` (was 0).
3. `TestTemperature_MeanBySeason_PartB_InnerBand_NoSeasonalSwing` — tilt 60°, lat 20° (inside `90−60 = 30°` boundary), summer solstice and equinox should produce identical Kelvin.
4. `TestTemperature_MeanBySeason_PartB_ArcticZone_HasSeasonalSwing` — tilt 60°, lat 50° (outside inner band), summer should exceed winter (this is the deferred test originally listed in #4's spec).

Existing tests stay green:

- `TestTemperature_MeanByLatitude_Tropical` (Part A, tilt 23.45°) — unchanged behavior.
- `TestTemperature_MeanBySeason_TropicalZone_NoSeasonalSwing` (Part A, tilt 23.6°, lat 10°) — unchanged behavior.
- `TestTemperature_MeanBySeason_MiddleZone_SummerExceedsWinter` (Part A, tilt 23.6°, lat 45°) — unchanged behavior.
- `TestTemperature_MeanBySeason_LatitudesProduceDifferentTemps` (Part A) — unchanged.

## Acceptance Criteria

- `zoneTiltAdjustment` produces non-zero, book-correct values for tilt ≥ 45° in both inner and arctic bands.
- `MeanBySeason` applies seasonal swing to Part B arctic latitudes (lat > 90 − tilt) and skips it inside the inner band.
- All existing temperature tests stay green.
- `task check && task test` clean.
- Issue #30 closes on merge.
