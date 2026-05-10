# zoneTiltAdjustment Part B Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix `zoneTiltAdjustment` and `tropicalLatitudeBoundary` to correctly handle WBH p.117 Part B (axial tilt ≥ 45°) — closes issue #30.

**Architecture:** Two-helper refactor inside `worlds/temperature.go`:

1. `tropicalLatitudeBoundary()` returns the **no-seasonal-swing boundary** — `tilt` for Part A and `(90 − tilt)` for Part B.
2. `zoneTiltAdjustment()` switch restructured to test Part A vs Part B first, then split by latitude — eliminates the previously unreachable Part B branch.
3. `MeanBySeason`'s `isTropical` check simplifies to `absLat <= t.tropicalLatitudeBoundary()` for both parts.

**Tech Stack:** Go 1.26, table-driven tests, gofumpt, `task check && task test` gate.

---

### Task 1: Failing tests for `zoneTiltAdjustment` Part B

**Files:**

- Modify: `worlds/temperature_test.go` (append after existing temperature tests)

- [ ] **Step 1: Add failing tests for Part B inner band and arctic zone**

Append to `worlds/temperature_test.go`:

```go
func TestTemperature_ZoneTiltAdjustment_PartB_InnerBand(t *testing.T) {
	// WBH p.117 Part B: for tilt ≥ 45°, the inner equatorial-tropical band
	// (|lat| ≤ 90 − tilt) returns sin(tilt − 45). For tilt 60° that is
	// sin(15°) ≈ 0.2588 — a positive (warming) factor reflecting the
	// always-illuminated equatorial band of a high-tilt world.
	temp := &Temperature{
		AxialTiltFactor: math.Sin(60 * math.Pi / 180.0), // tilt = 60°
	}
	got := temp.zoneTiltAdjustment(20)
	want := math.Sin(15 * math.Pi / 180.0)
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("got %v, want %v (sin(15°) for Part B inner band)", got, want)
	}
}

func TestTemperature_ZoneTiltAdjustment_PartB_ArcticZone(t *testing.T) {
	// WBH p.117 Part B: for tilt ≥ 45°, lat outside the inner band
	// (|lat| > 90 − tilt) uses the standard arctic formula sin(45 − lat).
	// For tilt 60°, lat 50° (outside the 30° inner band) → sin(-5°) ≈ -0.0872.
	temp := &Temperature{
		AxialTiltFactor: math.Sin(60 * math.Pi / 180.0),
	}
	got := temp.zoneTiltAdjustment(50)
	want := math.Sin(-5 * math.Pi / 180.0)
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("got %v, want %v (sin(-5°) for Part B arctic zone)", got, want)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./worlds/ -run "TestTemperature_ZoneTiltAdjustment_PartB" -v`

Expected: FAIL — current implementation returns 0 in both cases (case 1 of switch fires for both, with negative arg clamped to 0).

---

### Task 2: Failing tests for `MeanBySeason` Part B latitude composition

**Files:**

- Modify: `worlds/temperature_test.go`

- [ ] **Step 1: Add failing tests for Part B inner-band swing-skip and arctic-zone swing**

Append to `worlds/temperature_test.go`:

```go
func TestTemperature_MeanBySeason_PartB_InnerBand_NoSeasonalSwing(t *testing.T) {
	// WBH p.117 Part B: for tilt ≥ 45°, the inner equatorial-tropical band
	// (|lat| ≤ 90 − tilt) plays the role of Part A's tropical zone — no
	// seasonal swing. Tilt 60° → inner band ends at 30°. Lat 20° is inside.
	temp := &Temperature{
		MeanK:             288,
		Luminosity:        1.0,
		Albedo:            0.3,
		GreenhouseFactor:  0.36,
		AU:                1.0,
		AxialTiltFactor:   math.Sin(60 * math.Pi / 180.0),
		AtmosphericFactor: 2.0,
	}
	year := 365.25
	solstice := temp.MeanBySeason(20, 0, year)
	equinox := temp.MeanBySeason(20, year/4, year)
	if solstice != equinox {
		t.Errorf("Part B inner band should not swing seasonally: solstice=%v equinox=%v", solstice, equinox)
	}
}

func TestTemperature_MeanBySeason_PartB_ArcticZone_HasSeasonalSwing(t *testing.T) {
	// WBH p.117 Part B: for tilt ≥ 45°, latitudes outside the inner band
	// (|lat| > 90 − tilt) use the seasonal swing. Tilt 60°, lat 50° (outside
	// the 30° inner band) → summer solstice should exceed winter solstice.
	// (This was the deferred Part B test originally listed in #4's spec.)
	temp := &Temperature{
		MeanK:             288,
		Luminosity:        1.0,
		Albedo:            0.3,
		GreenhouseFactor:  0.36,
		AU:                1.0,
		AxialTiltFactor:   math.Sin(60 * math.Pi / 180.0),
		AtmosphericFactor: 2.0,
	}
	summer := temp.MeanBySeason(50, 0, 365.25)
	winter := temp.MeanBySeason(50, 365.25/2, 365.25)
	if summer <= winter {
		t.Errorf("Part B arctic zone should swing seasonally: summer=%v winter=%v", summer, winter)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./worlds/ -run "TestTemperature_MeanBySeason_PartB" -v`

Expected:

- `InnerBand_NoSeasonalSwing` FAILS — `isTropical` is false for tilt ≥ 45° (the `tiltDeg < 45` gate excludes Part B), so the seasonal swing applies and solstice ≠ equinox.
- `ArcticZone_HasSeasonalSwing` may pass or fail depending on whether the buggy `zoneTiltAdjustment` returning 0 at lat 50° still produces a measurable summer-vs-winter delta. Likely passes — this test guards against regression once Part B is properly fixed.

---

### Task 3: Refactor `zoneTiltAdjustment` and update `tropicalLatitudeBoundary`

**Files:**

- Modify: `worlds/temperature.go:454-498` (helpers and zone classifier)

- [ ] **Step 1: Replace `tropicalLatitudeBoundary` with part-aware version**

Replace the existing function (lines 454-469):

```go
// tropicalLatitudeBoundary returns the latitude (degrees, [0, 90]) at which
// the no-seasonal-swing zone ends. WBH p.116-117 reorganizes between the
// two parts:
//
//   - Part A (tilt < 45°): the tropical band runs |lat| ≤ axial_tilt.
//   - Part B (tilt ≥ 45°): the inner equatorial-tropical band runs
//     |lat| ≤ (90° − axial_tilt). Outside that band, the world enters
//     the arctic zone directly (no middle zone exists).
//
// NaN from Asin (when |AxialTiltFactor| > 1) clamps to 90; negative
// AxialTiltFactor takes the absolute value so the comparison stays
// meaningful when callers construct Temperature directly with a sign-
// flipped factor.
func (t *Temperature) tropicalLatitudeBoundary() float64 {
	tiltDeg := math.Asin(t.AxialTiltFactor) * 180.0 / math.Pi
	if math.IsNaN(tiltDeg) {
		return 90
	}
	if tiltDeg < 0 {
		tiltDeg = -tiltDeg
	}
	if tiltDeg >= 45 {
		return 90 - tiltDeg
	}
	return tiltDeg
}
```

- [ ] **Step 2: Refactor `zoneTiltAdjustment` to handle Part A and Part B symmetrically**

Replace the existing function (lines 471-498):

```go
// zoneTiltAdjustment returns the latitude-zone-adjusted axial-tilt-equivalent
// factor per WBH p.116-117 three-zone classification (tropical / middle /
// arctic). The structure differs between the book's two parts:
//
//	Part A (tilt < 45°):
//	  |lat| ≤ tilt           → sin(45° − tilt)        (tropical)
//	  |lat| > tilt           → sin(45° − lat)         (middle/arctic)
//
//	Part B (tilt ≥ 45°): the middle zone disappears; the inner equatorial-
//	tropical band uses the arctic-edge result, the rest uses arctic.
//	  |lat| ≤ (90° − tilt)   → sin(tilt − 45°)        (inner band)
//	  |lat| > (90° − tilt)   → sin(45° − lat)         (arctic)
//
// Returns are continuous at the Part A / Part B boundary (tilt = 45°): both
// formulas evaluate to 0 there.
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

	// WBH p.116 Part A: tropical band is |lat| ≤ tilt; rest is middle/arctic.
	if latDeg <= rawTilt {
		return math.Sin((45.0 - rawTilt) * math.Pi / 180.0)
	}
	return math.Sin((45.0 - latDeg) * math.Pi / 180.0)
}
```

- [ ] **Step 3: Run the new Part B tests to verify they pass**

Run: `go test ./worlds/ -run "TestTemperature_ZoneTiltAdjustment_PartB" -v`

Expected: PASS — both inner-band and arctic-zone tests now produce the correct sin(15°) and sin(−5°) values.

---

### Task 4: Simplify `MeanBySeason` `isTropical` check

**Files:**

- Modify: `worlds/temperature.go:520-576` (`MeanBySeason` body, specifically the isTropical block at lines 542-549)

- [ ] **Step 1: Replace the `isTropical` block with the part-aware version**

In `MeanBySeason`, find this block (around lines 542-549):

```go
	// "No seasonal swing" applies in the tropical zone for Part A worlds
	// (axial tilt < 45°). For Part B worlds (tilt ≥ 45°) the zone boundaries
	// reorganize per WBH p.117 and the no-swing region is (90 − tilt), not
	// the raw tilt — see issue #30. Until that lands, gate the swing-skip
	// to Part A only so high-tilt worlds keep applying the seasonal swing
	// uniformly (their pre-#4 behavior).
	tiltDeg := t.tropicalLatitudeBoundary()
	isTropical := tiltDeg < 45 && absLat <= tiltDeg
```

Replace with:

```go
	// "No seasonal swing" applies inside the tropical band — for Part A
	// (tilt < 45°) that band is |lat| ≤ tilt; for Part B (tilt ≥ 45°) it
	// is |lat| ≤ (90 − tilt). tropicalLatitudeBoundary returns the correct
	// boundary for each part per WBH p.116-117.
	isTropical := absLat <= t.tropicalLatitudeBoundary()
```

- [ ] **Step 2: Run all the new Part B tests to verify they pass**

Run: `go test ./worlds/ -run "TestTemperature_MeanBySeason_PartB" -v`

Expected: PASS — both `InnerBand_NoSeasonalSwing` and `ArcticZone_HasSeasonalSwing` pass.

- [ ] **Step 3: Run all temperature tests to verify Part A behavior is preserved**

Run: `go test ./worlds/ -run "TestTemperature_" -v`

Expected: all PASS — Part A tropical-no-swing, middle-zone summer-vs-winter, and latitude-differs tests stay green.

---

### Task 5: Full quality gate and Zed golden

**Files:** none — verification only.

- [ ] **Step 1: Run the full test suite**

Run: `task test`

Expected: PASS — including Zed worked-example regressions and the Zed golden file (no production callers of `MeanBySeason` so golden is unaffected, but verify).

- [ ] **Step 2: Run the modernizer + lint + format gate**

Run: `task check`

Expected: PASS — clean.

- [ ] **Step 3: Commit**

```bash
git add worlds/temperature.go worlds/temperature_test.go docs/pass-1/specs/2026-05-09-zonetiltadjustment-part-b-design.md docs/pass-1/plans/2026-05-09-zonetiltadjustment-part-b.md
git commit -m "$(cat <<'EOF'
fix(worlds): zoneTiltAdjustment + tropicalLatitudeBoundary handle WBH p.117 Part B (closes #30)

WBH p.117 Part B prescribes a separate latitude composition for worlds with
axial tilt ≥ 45°: the inner equatorial-tropical band runs |lat| ≤ (90 − tilt)
and uses sin(tilt − 45) (positive — always-illuminated equatorial warming);
the rest is arctic with sin(45 − lat).

Two related bugs fixed:

1. zoneTiltAdjustment switch case 2 (`tiltDeg >= 45 && latDeg < (90-tiltDeg)`)
   was unreachable because case 1 (`latDeg <= tiltDeg`) always fired first
   for high-tilt worlds. Restructured the switch to test Part A vs Part B
   first, then split by latitude.

2. MeanBySeason's isTropical check (added in #4) gated the no-swing region
   to `tiltDeg < 45` only — leaving Part B worlds applying the seasonal
   swing uniformly across all latitudes. tropicalLatitudeBoundary now
   returns the part-aware boundary (tilt for Part A, 90−tilt for Part B),
   and isTropical simplifies to absLat <= tropicalLatitudeBoundary().

No production callers of MeanBySeason / MeanByLatitude — Zed golden
unaffected. Library-only correctness fix.

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 4: Verify clean tree**

Run: `git status`

Expected: clean.

---

## Self-Review

Spec coverage:

- Bug 1 (`zoneTiltAdjustment` Part B unreachable) — Task 3 Step 2.
- Bug 2 (`MeanBySeason` `isTropical` Part B-blind) — Task 3 Step 1 + Task 4 Step 1.
- Tests for Part B inner band — Task 1, Task 2.
- Tests for Part B arctic zone — Task 1, Task 2.
- Existing Part A tests preserved — Task 4 Step 3 verification.

Placeholder scan: none.

Type consistency: helper signatures and call sites match.

## Execution

Subagent-driven via `superpowers:subagent-driven-development`. Single subagent dispatch covers all five tasks (≈40 LOC of changes, scope contained to one file + tests).
