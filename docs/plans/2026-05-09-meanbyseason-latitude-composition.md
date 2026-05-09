# MeanBySeason Latitude Composition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Compose the seasonal axial-tilt swing with the existing zone-latitude adjustment in `Temperature.MeanBySeason` per WBH p.116, so latDeg actually affects the output. Tropical zone gets no seasonal swing (book-explicit); middle/arctic zones add seasonal axial tilt to the zone latitude adjustment.

**Architecture:** Extract a small private helper `tropicalLatitudeBoundary` on `Temperature` for the tilt-degree calculation already inside `zoneTiltAdjustment`. Rewrite `MeanBySeason`'s composition logic to consult both the zone adjustment and seasonal axial tilt according to the latitude band. No public-API change.

**Tech Stack:** Go 1.26, `task` (gofumpt + go vet + golangci-lint + modernizer), `go test -race`.

---

## File Structure

- **Modify:** `worlds/temperature.go` — add `tropicalLatitudeBoundary` helper; refactor `zoneTiltAdjustment` to use it; rewrite `MeanBySeason`'s body and doc-comment.
- **Modify:** `worlds/temperature_test.go` — add 3-4 latitude-composition tests; existing `TestTemperature_MeanBySeason_OppositeSolstices` and `TestTemperature_AtMoment_*` should keep passing.

No Zed golden refresh expected — `MeanBySeason` has no production caller per its doc-comment.

---

### Task 1: Latitude composition in `MeanBySeason` (TDD)

**Files:**

- Modify: `worlds/temperature_test.go` (4 new tests)
- Modify: `worlds/temperature.go` (helper extraction + `MeanBySeason` rewrite)

- [ ] **Step 1: Write the failing latitude-variation test**

Open `worlds/temperature_test.go` and append (after `TestTemperature_AtMoment_NoonExceedsDawn`):

```go
func TestTemperature_MeanBySeason_LatitudesProduceDifferentTemps(t *testing.T) {
	// Same date, different latitudes → different temperatures. The
	// existing implementation returns the same value at all latitudes
	// because the seasonal swing is applied as a global luminosity
	// modifier without composing with the latitude zone formula.
	temp := &Temperature{
		MeanK:             288,
		Luminosity:        1.0,
		Albedo:            0.3,
		GreenhouseFactor:  0.36,
		AU:                1.0,
		AxialTiltFactor:   0.40, // ~23.6° tilt → tropical zone is |lat| ≤ 23.6°
		AtmosphericFactor: 2.0,
	}
	summerSolstice := 0.0
	tropical := temp.MeanBySeason(0, summerSolstice, 365.25)  // equator
	temperate := temp.MeanBySeason(45, summerSolstice, 365.25) // middle zone
	if tropical == temperate {
		t.Errorf("equator (%v) and 45°N (%v) returned identical temps; latitude composition is not applied",
			tropical, temperate)
	}
}

func TestTemperature_MeanBySeason_TropicalZone_NoSeasonalSwing(t *testing.T) {
	// WBH p.116: "tropical temperatures have little seasonal variation
	// (from axial tilt)." A latitude inside the tropical band (|lat| ≤
	// axial tilt) should return the same value at solstice and equinox.
	temp := &Temperature{
		MeanK:             288,
		Luminosity:        1.0,
		Albedo:            0.3,
		GreenhouseFactor:  0.36,
		AU:                1.0,
		AxialTiltFactor:   0.40, // ~23.6° tilt; lat=10° is well inside tropical
		AtmosphericFactor: 2.0,
	}
	year := 365.25
	solstice := temp.MeanBySeason(10, 0, year)
	equinox := temp.MeanBySeason(10, year/4, year)
	if solstice != equinox {
		t.Errorf("tropical zone should not swing seasonally: solstice=%v equinox=%v", solstice, equinox)
	}
}

func TestTemperature_MeanBySeason_MiddleZone_SummerExceedsWinter(t *testing.T) {
	// WBH p.116: middle/arctic zones add the seasonal axial-tilt factor
	// to the zone latitude adjustment. At lat 45° (middle zone for tilt
	// 23.6°), summer solstice should be warmer than winter solstice.
	// (This was the only assertion the old buggy implementation could
	// satisfy; we keep it as a regression guard.)
	temp := &Temperature{
		MeanK:             288,
		Luminosity:        1.0,
		Albedo:            0.3,
		GreenhouseFactor:  0.36,
		AU:                1.0,
		AxialTiltFactor:   0.40,
		AtmosphericFactor: 2.0,
	}
	summer := temp.MeanBySeason(45, 0, 365.25)
	winter := temp.MeanBySeason(45, 365.25/2, 365.25)
	if summer <= winter {
		t.Errorf("middle zone: summer %v should exceed winter %v", summer, winter)
	}
}

func TestTemperature_MeanBySeason_NoYearLength_FallsBackToMeanByLatitude(t *testing.T) {
	// localYearDays == 0 short-circuits to MeanByLatitude regardless of
	// daysSinceSolstice. Pre-existing behavior; sanity guard.
	temp := &Temperature{
		MeanK:             288,
		Luminosity:        1.0,
		Albedo:            0.3,
		GreenhouseFactor:  0.36,
		AU:                1.0,
		AxialTiltFactor:   0.40,
		AtmosphericFactor: 2.0,
	}
	for _, lat := range []float64{0, 30, 60, 89} {
		want := temp.MeanByLatitude(lat)
		got := temp.MeanBySeason(lat, 999, 0)
		if got != want {
			t.Errorf("lat %v: got %v, want %v (MeanByLatitude fallback)", lat, got, want)
		}
	}
}
```

- [ ] **Step 2: Run the new tests to verify they fail (or pass for the safe ones)**

Run:

```bash
go test ./worlds/ -run TestTemperature_MeanBySeason -v
```

Expected:

- `_LatitudesProduceDifferentTemps` FAILS — current implementation returns the same value at all latitudes.
- `_TropicalZone_NoSeasonalSwing` FAILS — current implementation applies the seasonal swing globally including in the tropics.
- `_MiddleZone_SummerExceedsWinter` PASSES — current bug already produces this asymmetry (the seasonal swing is applied uniformly).
- `_NoYearLength_FallsBackToMeanByLatitude` PASSES — short-circuit unchanged.
- `_OppositeSolstices` (pre-existing) PASSES — same shape as `_MiddleZone_SummerExceedsWinter`.

- [ ] **Step 3: Add the `tropicalLatitudeBoundary` helper**

Open `worlds/temperature.go`. Find `zoneTiltAdjustment` (currently at line 457). Directly above it, insert the helper:

```go
// tropicalLatitudeBoundary returns the absolute axial tilt in degrees
// (clamped to [0, 90]) — the WBH p.116 boundary between the tropical
// zone (|lat| ≤ tilt) and the rest of the world. NaN from Asin (when
// |AxialTiltFactor| > 1) clamps to 0 or 90 by sign.
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

- [ ] **Step 4: Refactor `zoneTiltAdjustment` to use the helper**

In the same file, replace the existing `zoneTiltAdjustment`:

```go
// zoneTiltAdjustment returns the latitude-zone-adjusted axial-tilt-equivalent
// factor per WBH p.116-117 three-zone classification (tropical / middle /
// arctic). For axial tilt ≥ 45° the middle zone disappears (Part B p.117).
func (t *Temperature) zoneTiltAdjustment(latDeg float64) float64 {
	tiltDeg := math.Asin(t.AxialTiltFactor) * 180.0 / math.Pi
	if math.IsNaN(tiltDeg) {
		// |AxialTiltFactor| > 1 — clamp.
		if t.AxialTiltFactor > 0 {
			tiltDeg = 90
		} else {
			tiltDeg = 0
		}
	}
	if latDeg < 0 {
		latDeg = -latDeg
	}
	if latDeg > 90 {
		latDeg = 90
	}

	switch {
	case latDeg <= tiltDeg:
		// Tropical zone: sin(45° - axial_tilt) replaces axial tilt factor.
		adj := 45.0 - tiltDeg
		if adj < 0 {
			adj = 0
		}
		return math.Sin(adj * math.Pi / 180.0)
	case tiltDeg >= 45 && latDeg < (90-tiltDeg):
		// Part B: no middle zone; use arctic-edge result at lat=90-tilt.
		return math.Sin((45.0 - (90.0 - tiltDeg)) * math.Pi / 180.0)
	default:
		// Middle/arctic: sin(45° - latitude).
		return math.Sin((45.0 - latDeg) * math.Pi / 180.0)
	}
}
```

with the simplified version that reuses `tropicalLatitudeBoundary`:

```go
// zoneTiltAdjustment returns the latitude-zone-adjusted axial-tilt-equivalent
// factor per WBH p.116-117 three-zone classification (tropical / middle /
// arctic). For axial tilt ≥ 45° the middle zone disappears (Part B p.117).
func (t *Temperature) zoneTiltAdjustment(latDeg float64) float64 {
	tiltDeg := t.tropicalLatitudeBoundary()
	if latDeg < 0 {
		latDeg = -latDeg
	}
	if latDeg > 90 {
		latDeg = 90
	}

	switch {
	case latDeg <= tiltDeg:
		// Tropical zone: sin(45° - axial_tilt) replaces axial tilt factor.
		adj := 45.0 - tiltDeg
		if adj < 0 {
			adj = 0
		}
		return math.Sin(adj * math.Pi / 180.0)
	case tiltDeg >= 45 && latDeg < (90-tiltDeg):
		// Part B: no middle zone; use arctic-edge result at lat=90-tilt.
		return math.Sin((45.0 - (90.0 - tiltDeg)) * math.Pi / 180.0)
	default:
		// Middle/arctic: sin(45° - latitude).
		return math.Sin((45.0 - latDeg) * math.Pi / 180.0)
	}
}
```

- [ ] **Step 5: Rewrite `MeanBySeason`**

Find the current `MeanBySeason`:

```go
// MeanBySeason returns the mean temperature on a specific day at a specific
// latitude, ignoring time of day, per WBH p.115.
//
// daysSinceSolstice: 0 = summer solstice in the relevant hemisphere; year/2 = winter solstice.
// localYearDays: caller decides — for moons, use parent's stellar year (moons co-orbit star with planet).
//
// KNOWN LIMITATION: the current implementation applies the seasonal axial-tilt
// swing as a signed luminosity modifier without composing it with the latitude
// zone formula. As a result, latDeg currently does not affect the output (the
// seasonal swing is uniform across latitudes for a given date). The spec
// foresaw layering both factors, but the plan's substitution-into-zone
// approach zeroes out at lat=45° (sin(45-45)=0). Composing them properly is
// deferred to a follow-up — no production caller exists yet.
//
// Twilight worlds always return TwilightK (band centerline). Hemisphere-aware
// selection — bright/dark/twilight by latitude — is the caller's responsibility:
// read t.BrightSideK / t.TwilightK / t.DarkSideK directly. The spec foresaw
// auto-selection inside this method but the implementation defers it.
func (t *Temperature) MeanBySeason(latDeg, daysSinceSolstice, localYearDays float64) float64 {
	if t.IsTwilight {
		return t.TwilightK
	}
	if localYearDays <= 0 {
		return t.MeanByLatitude(latDeg)
	}

	// Adjusted Fractional Year per WBH p.115.
	stdYearDays := 8766.0 / 24.0 // 365.25
	lagDays := 0.1 * math.Min(stdYearDays, localYearDays)
	adjFracYear := (daysSinceSolstice - 0.1*lagDays) / localYearDays

	// Seasonal axial tilt factor: cos(adjFracYear × 360°) × stored AxialTiltFactor.
	// Positive = summer (sun higher in sky → more heat); negative = winter (less heat).
	// Apply directly as a signed luminosity modifier: the axial-tilt contribution
	// swings from +AxialTiltFactor at summer solstice to -AxialTiltFactor at winter.
	seasonalTilt := math.Cos(adjFracYear*2*math.Pi) * t.AxialTiltFactor

	lumMod := seasonalTilt / t.AtmosphericFactor
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

Replace with:

```go
// MeanBySeason returns the mean temperature on a specific day at a specific
// latitude, ignoring time of day, per WBH p.115-116.
//
// daysSinceSolstice: 0 = summer solstice in the relevant hemisphere; year/2 = winter solstice.
// localYearDays: caller decides — for moons, use parent's stellar year (moons co-orbit star with planet).
//
// Composition rule per WBH p.116 three-zone classification:
//   - Tropical zone (|lat| ≤ axial tilt): no seasonal swing — "tropical
//     temperatures have little seasonal variation (from axial tilt)."
//     The zone latitude adjustment from zoneTiltAdjustment is the sole
//     variance contributor.
//   - Middle/arctic zone: seasonal axial-tilt factor is added to the
//     zone latitude adjustment per "the zone's latitude adjustment is
//     added to the axial tilt factor for that time period."
//
// Twilight worlds always return TwilightK (band centerline). Hemisphere-aware
// selection — bright/dark/twilight by latitude — is the caller's responsibility:
// read t.BrightSideK / t.TwilightK / t.DarkSideK directly. The spec foresaw
// auto-selection inside this method but the implementation defers it (see
// issue #5).
func (t *Temperature) MeanBySeason(latDeg, daysSinceSolstice, localYearDays float64) float64 {
	if t.IsTwilight {
		return t.TwilightK
	}
	if localYearDays <= 0 {
		return t.MeanByLatitude(latDeg)
	}

	// Zone latitude adjustment from WBH p.116. Sole variance contributor
	// in the tropical zone; added to the seasonal axial tilt in middle/
	// arctic zones.
	zoneAdj := t.zoneTiltAdjustment(latDeg)
	absLat := latDeg
	if absLat < 0 {
		absLat = -absLat
	}
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

		// Seasonal axial tilt factor: cos(adjFracYear × 360°) × AxialTiltFactor.
		// +AxialTiltFactor at summer solstice; -AxialTiltFactor at winter.
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

- [ ] **Step 6: Run all `MeanBySeason` tests to verify they pass**

```bash
go test ./worlds/ -run TestTemperature_MeanBySeason -v
```

Expected: PASS — all 4 new tests plus the pre-existing `TestTemperature_MeanBySeason_OppositeSolstices`.

- [ ] **Step 7: Run all temperature tests**

```bash
go test ./worlds/ -run "TestTemperature_|TestGenerateTemperature_" -v 2>&1 | tail -40
```

Expected: PASS — including `MeanByLatitude_*`, `AtMoment_*`, and twilight tests. The `MeanByLatitude` tests should be unaffected (the helper extraction is a no-op refactor for that function). `AtMoment` calls `MeanBySeason` for the seasonal contribution — its existing tests use lat 0 (tropical at tilt 0.40 ≈ 23.6°), so the seasonal contribution drops out and the rotation factor still produces noon > dawn.

- [ ] **Step 8: Stage Go changes**

```bash
git add worlds/temperature.go worlds/temperature_test.go
```

- [ ] **Step 9: Run the full test suite**

```bash
go test -race ./...
```

Expected: PASS, including `TestRenderSystemMarkdown_ZedGolden`. `MeanBySeason` has no production caller per its doc-comment, so the Zed golden should not shift. If it does, that signals an unexpected dependency — STOP and escalate.

- [ ] **Step 10: Run task quality gate**

```bash
task check
```

Expected: clean.

- [ ] **Step 11: Commit**

```bash
git commit -m "$(cat <<'EOF'
fix(worlds): MeanBySeason composes latitude with season per WBH p.116 (closes #4)

Previously the seasonal axial-tilt swing was applied as a global
luminosity modifier without composing with the latitude zone formula,
so latDeg did not affect the output. Per WBH p.116:

- Tropical zone (|lat| ≤ axial_tilt): no seasonal swing — the book is
  explicit that "tropical temperatures have little seasonal variation
  (from axial tilt)." Variance = zoneTiltAdjustment(latDeg) alone.
- Middle/arctic zone: variance = zoneTiltAdjustment(latDeg) +
  seasonalAxialTilt per "the zone's latitude adjustment is added to
  the axial tilt factor for that time period."

Extracts a small private helper tropicalLatitudeBoundary on
Temperature to avoid duplicating the tilt-degree comparison logic
between zoneTiltAdjustment and the new branch in MeanBySeason.

No production caller of MeanBySeason exists per its doc-comment;
Zed golden unaffected. AtMoment composes via MeanBySeason and
inherits the fix automatically.

Closes #4.

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: PR + close-out

**Files:** none (operational steps only).

- [ ] **Step 1: Push the branch**

```bash
git push -u origin feat/meanbyseason-latitude-composition
```

- [ ] **Step 2: Open the PR**

```bash
gh pr create --repo philoserf/world-builder --title "fix(worlds): MeanBySeason composes latitude with season per WBH p.116 (closes #4)" --body "$(cat <<'EOF'
## Summary

Per WBH p.116, `Temperature.MeanBySeason` now composes the seasonal axial-tilt swing with the existing zone latitude adjustment from `zoneTiltAdjustment`:

- **Tropical zone** (|lat| ≤ axial tilt): no seasonal swing — \"tropical temperatures have little seasonal variation (from axial tilt).\" `MeanBySeason` returns the same value for any `daysSinceSolstice` at a given tropical latitude.
- **Middle/arctic zone**: variance = `zoneTiltAdjustment(latDeg) + seasonalAxialTilt` — the book's \"the zone's latitude adjustment is added to the axial tilt factor for that time period.\"

Extracts a small private helper `tropicalLatitudeBoundary` on `Temperature` so `zoneTiltAdjustment` and `MeanBySeason` share the tilt-degree comparison logic.

`AtMoment` composes via `MeanBySeason` and inherits the fix automatically. No production caller of `MeanBySeason` exists per its doc-comment, so the Zed golden is unaffected.

Closes #4.

## Spec / plan

- Spec: `docs/specs/2026-05-09-meanbyseason-latitude-composition-design.md`
- Plan: `docs/plans/2026-05-09-meanbyseason-latitude-composition.md`

## Test plan

- [x] \`task check\` clean (gofumpt, vet, golangci-lint, modernizer)
- [x] \`task test\` clean with race detector
- [x] New \`TestTemperature_MeanBySeason_LatitudesProduceDifferentTemps\` — equator vs 45°N at the same date now return different temps (this fails on current code; the bug we're fixing)
- [x] New \`TestTemperature_MeanBySeason_TropicalZone_NoSeasonalSwing\` — solstice and equinox at lat 10° return identical temps (book-prescribed tropical behavior)
- [x] New \`TestTemperature_MeanBySeason_MiddleZone_SummerExceedsWinter\` — middle zone still has seasonal asymmetry (regression guard)
- [x] New \`TestTemperature_MeanBySeason_NoYearLength_FallsBackToMeanByLatitude\` — \`localYearDays == 0\` short-circuit unchanged
- [x] Existing \`TestTemperature_MeanBySeason_OppositeSolstices\`, \`TestTemperature_AtMoment_*\`, \`TestTemperature_MeanByLatitude_*\` all unchanged and passing

## Out of scope

- Twilight-zone hemisphere selection in \`MeanBySeason\`/\`AtMoment\` — tracked as #5.
- Geographic factor inclusion in per-latitude variance — preserves existing scope.
- Eccentricity-driven seasonal effects (book mentions but doesn't prescribe).
EOF
)"
```

- [ ] **Step 3: Stop**

Implementation complete on the branch; PR is open. Hand back to the user for review/merge.

---

## Self-review

**Spec coverage**

- Spec § Architecture (helper extraction): Task 1 Steps 3 + 4. ✓
- Spec § Architecture (`MeanBySeason` rewrite): Task 1 Step 5. ✓
- Spec § Decisions (tropical no-swing): Task 1 Step 1 (`_TropicalZone_NoSeasonalSwing` test) + Step 5 implementation. ✓
- Spec § Decisions (middle/arctic additive): Task 1 Step 1 (`_MiddleZone_SummerExceedsWinter` + `_LatitudesProduceDifferentTemps`) + Step 5 implementation. ✓
- Spec § Decisions (function signature unchanged): Task 1 Step 5 — same `func(latDeg, daysSinceSolstice, localYearDays float64) float64`. ✓
- Spec § Testing strategy 4 unit tests: Task 1 Step 1 has all 4. ✓
- Spec § Testing strategy "no Zed golden shift": Task 1 Step 9. ✓

**Placeholder scan**

No "TBD" / "TODO" / "implement later". All steps include concrete code or commands.

**Type consistency**

- `tropicalLatitudeBoundary() float64` — same signature in helper definition (Step 3) and call sites (Step 4 `zoneTiltAdjustment`, Step 5 `MeanBySeason`).
- `MeanBySeason(latDeg, daysSinceSolstice, localYearDays float64) float64` — same signature throughout (test calls + implementation).
- `Temperature` struct fields — `AxialTiltFactor`, `AtmosphericFactor`, `Luminosity`, `Albedo`, `GreenhouseFactor`, `AU`, `IsTwilight`, `TwilightK` — all already exist on the struct; no new fields.
