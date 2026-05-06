# World Physical 3B-Final Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement WBH pp.132-146 (Habitability rating + Final Mainworld Determination + IISS Class IV-P Survey Form rendering) as a single new pipeline step `runStep5G` between 3B-biology and Step 6, plus a system-wide mainworld pick after 5G, plus a per-body Class IV-P form renderer.

**Architecture:** A new `Habitability` struct attached as `dp.Habitability *Habitability` (and `m.Habitability` on Moon) holds the rating + reserved Notes string. `pickMainworld(detailed) string` produces a designation stored on `SystemDetail.MainworldDesignation`. `RenderIISSClass4P(body, sys, mainworldDesignation) string` renders the per-body form on demand. Body filter for 5G: terrestrials only (skip belts/GGs/empty); atmosphere optional (vacuum → atm 0 DM).

**Tech Stack:** Go 1.26, existing `wbh/roller`, `wbh/stars`, `wbh/worlds` packages. Same workflow as 3B-biology: per-task subagent (Sonnet) → spec reviewer → code reviewer → next task. Final end-to-end review on Opus before merge.

---

## File map

| File                                  | Status   | Purpose                                                                                                                                                                                                          |
| ------------------------------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `worlds/habitability.go`              | New      | `Habitability` struct + `ComputeHabitability` + DM helpers                                                                                                                                                       |
| `worlds/habitability_test.go`         | New      | Per-formula tests + boundary cases                                                                                                                                                                               |
| `worlds/mainworld.go`                 | New      | `pickMainworld(detailed) string` + priority-chain logic                                                                                                                                                          |
| `worlds/mainworld_test.go`            | New      | Priority-chain tests                                                                                                                                                                                             |
| `worlds/iiss_class4p.go`              | New      | `RenderIISSClass4P(body, sys, mainworldDesignation) string`                                                                                                                                                      |
| `worlds/iiss_class4p_test.go`         | New      | Render-shape + key-field tests                                                                                                                                                                                   |
| `worlds/system_detail_step5g.go`      | New      | `runStep5G` orchestrator + `habitabilityApplies` + `computeHabitability`                                                                                                                                         |
| `worlds/system_detail_step5g_test.go` | New      | Orchestrator tests                                                                                                                                                                                               |
| `worlds/system_detail.go`             | Modified | Add `Habitability *Habitability` to `DetailedPlacement`; add `MainworldDesignation string` to `SystemDetail`; add `HasHabitability()` accessor; one new line each for Step 5G + mainworld pick in `DetailSystem` |
| `worlds/moons.go`                     | Modified | Add `Habitability *Habitability` to `Moon` + `HasHabitability()` accessor                                                                                                                                        |
| `worlds/worked_examples_test.go`      | Modified | Append assertions 39-43 + 7th trailing t.Logf                                                                                                                                                                    |

## Reference

- **Spec:** `docs/specs/2026-05-05-world-physical-3b-final-design.md` (commit `aa327d3`)
- **WBH source:** pp.132-146
- **Predecessor:** 3B-biology merged on `main` as `b89d09b`; pre-3B-final cleanup at `3948b82`

## API gotchas (from prior sub-projects)

- `r.Roll("2D")` not `r.Roll(2, 6)`; constructor is `roller.NewScripted(...)` (variadic ints); `Roll` returns `int` with no error.
- `SizeAsInt(SizeCode) int` lives in `worlds/atmosphere.go`; converts `"0".."F"` → `0..15`.
- `dp.Atmosphere.Code` is `int` 0-15 (eHex encoded).
- `dp.Hydrographics.Code` is `int` 0-10 (10 = "A").
- `dp.Temperature.MeanK` / `HighK` / `LowK` are populated post-3A2b/3B-geology.
- `dp.Physical.Density` / `Gravity` / `Mass` are pointer fields (nil-safe).
- `dp.Body == BodyTerrestrial`, `BodyGasGiant`, `BodyEmpty`, `BodyPlanetoidBelt` (constants in `worlds/placement.go`).
- `dp.SizeCode == "0"` indicates a belt; SizeCode is `string`.
- `dp.TidalLock.IsTwilightZone` is `true` ONLY when `Case == TidalLockCasePlanetToStar && LockRatio == "1:1"` — the exact predicate for "Solar 1:1 tidally locked world" per WBH p.132. Use this field directly.
- `dp.Biology.HasNativeSophont` and `dp.Biology.HadExtinctSophont` are `bool` flags from 3B-biology.
- `dp.Biology.ResourceRating` is the `2D-7+Size+DMs` value, clamped to [2, 12].
- `buildMoonPlacementView(m, parent)` synthesizes a moon view; pointer-aliases Atmosphere/Hydrographics/Physical. Manually alias `Temperature` (precedent: 3B-biology Task 9 fix; same applies here for Habitability if needed).
- Project's `just check` runs `go fix ./...` and FAILS if it produces unstaged changes — apply any modernize rewrites; don't dismiss.
- Stale LSP "undefined" diagnostics are documented project noise — trust `just check && just test` exit codes.

## Final-review pattern

The Opus final-gate review has caught a Critical integration-level bug in EVERY recent sub-project (3A2a YearDays, 3A2b-temp MeanByLatitude, 3A2b-rederive RollGasMix column, 3B-geology terrestrial MassEarth, 3B-biology moonDP.Temperature). **Don't skip Task 10.**

---

## Task 1: Branch setup + `Habitability` struct + `MainworldDesignation` field + accessors

**Files:**

- Create: `worlds/habitability.go`
- Modify: `worlds/system_detail.go` (DetailedPlacement struct + SystemDetail struct + HasHabitability accessor)
- Modify: `worlds/moons.go` (Moon struct + HasHabitability accessor)

- [ ] **Step 1: Create the branch from main**

```bash
cd /Users/markayers/Documents/Traveller
git checkout main
git pull --ff-only 2>/dev/null || true
git checkout -b feat/wbh-world-physical-3b-final
```

- [ ] **Step 2: Create `worlds/habitability.go` with the `Habitability` struct**

```go
// Package worlds — per-body habitability rating per WBH p.132-133
// (sub-project 3B-final).
package worlds

// Habitability — a per-body habitability rating for Terragens per WBH
// p.132-133. Computed by Step 5G for any non-empty terrestrial body
// (and HZ-planet moons).
//
// Range: 0-12. The book theoretically allows higher but treats 12 as
// "very unlikely" and clamps negative results to 0.
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

- [ ] **Step 3: Add `Habitability *Habitability` field to `DetailedPlacement`**

In `worlds/system_detail.go`, find the `DetailedPlacement` struct definition. Add the new field at the END of the existing pointer-field group, after the `// 3B-biology additions` block (which ends with `Biology *Biology`):

```go
	// 3B-final additions
	Habitability *Habitability
```

- [ ] **Step 4: Add `Habitability *Habitability` field to `Moon`**

In `worlds/moons.go`, find the `Moon` struct definition. Add the new field at the END, after the existing `// 3B-biology additions`:

```go
	// 3B-final additions
	Habitability *Habitability
```

- [ ] **Step 5: Add `MainworldDesignation` field to `SystemDetail`**

In `worlds/system_detail.go`, find the `SystemDetail` struct (around line 445). Add the new field at the END, after `Survey`:

```go
	// MainworldDesignation is the auto-picked mainworld's designation per
	// WBH p.134. Priority chain: bodies with native sophonts → highest
	// habitability → highest resource → first in iteration order.
	// Empty string if no terrestrial body qualifies.
	//
	// The book explicitly says the Referee may override this pick. A future
	// sub-project may add a Referee-override mechanism; for now the
	// auto-pick is the only source.
	MainworldDesignation string
```

- [ ] **Step 6: Add `HasHabitability()` accessors**

In `worlds/system_detail.go`, after the existing `HasBiology()` accessor on `*DetailedPlacement`, add:

```go
func (dp *DetailedPlacement) HasHabitability() bool { return dp.Habitability != nil }
```

In `worlds/moons.go`, after the existing `HasBiology()` accessor on `*Moon`, add:

```go
// HasHabitability reports whether habitability data has been generated for this moon.
func (m *Moon) HasHabitability() bool { return m.Habitability != nil }
```

- [ ] **Step 7: Smoke check**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
just check
just test
```

Expected: 0 issues; all packages pass.

- [ ] **Step 8: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/habitability.go \
        worlds/system_detail.go \
        worlds/moons.go
git commit -m "feat(worlds): Habitability struct + MainworldDesignation field + accessors"
```

---

## Task 2: `ComputeHabitability` — Size + Atmosphere + Hydrographics + TidalLock DMs

**Files:**

- Modify: `worlds/habitability.go`
- Create: `worlds/habitability_test.go`

This task implements the formula skeleton + the four "easy" DM categories. Task 3 finishes with Temperature + Gravity DMs and the final clamp.

- [ ] **Step 1: Write failing tests**

Create `worlds/habitability_test.go`:

```go
package worlds

import "testing"

func TestComputeHabitability_BaselineNoDMs(t *testing.T) {
	// Size 5, Atm 6 (no DM), Hydro 5 (no DM), no tidal lock, no temp/gravity DMs.
	// Result: 10 + 0 = 10.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	if got.Rating != 10 {
		t.Errorf("got %d, want 10", got.Rating)
	}
}

func TestComputeHabitability_SmallSize_DMMinus1(t *testing.T) {
	// Size 4 (in 0-4) → DM-1.
	body := &DetailedPlacement{}
	body.SizeCode = "4"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	if got.Rating != 9 {
		t.Errorf("got %d, want 9 (Size 4 DM-1)", got.Rating)
	}
}

func TestComputeHabitability_LargeSize_DMPlus1(t *testing.T) {
	// Size 9 (in 9+) → DM+1.
	body := &DetailedPlacement{}
	body.SizeCode = "9"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	if got.Rating != 11 {
		t.Errorf("got %d, want 11 (Size 9 DM+1)", got.Rating)
	}
}

func TestComputeHabitability_AtmVacuum_DMMinus8(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 0}
	body.Hydrographics = &Hydrographics{Code: 0}
	got := ComputeHabitability(body)
	// 10 + (-8) + (-4 hydro 0) = -2 → clamp 0
	if got.Rating != 0 {
		t.Errorf("got %d, want 0 (atm 0 vacuum + hydro 0)", got.Rating)
	}
}

func TestComputeHabitability_NilAtmosphere_TreatedAsAtm0(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = nil
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	// 10 + (-8) + 0 = 2
	if got.Rating != 2 {
		t.Errorf("got %d, want 2 (nil atm → DM-8)", got.Rating)
	}
}

func TestComputeHabitability_AtmHostile_DMMinus10(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 11} // B
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	// 10 + (-10) = 0
	if got.Rating != 0 {
		t.Errorf("got %d, want 0 (atm B hostile)", got.Rating)
	}
}

func TestComputeHabitability_AtmVeryHostile_DMMinus12(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 12} // C
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	// 10 + (-12) = -2 → clamp 0
	if got.Rating != 0 {
		t.Errorf("got %d, want 0 (atm C very hostile)", got.Rating)
	}
}

func TestComputeHabitability_HydroDesert_DMMinus2(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 2}
	got := ComputeHabitability(body)
	// 10 + (-2) = 8
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (Hydro 2 desert)", got.Rating)
	}
}

func TestComputeHabitability_HydroFull_DMMinus2(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 10} // A
	got := ComputeHabitability(body)
	// 10 + (-2) = 8
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (Hydro A very-full)", got.Rating)
	}
}

func TestComputeHabitability_TidalLock1to1_DMMinus2(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.TidalLock = &TidalLock{
		Case:           TidalLockCasePlanetToStar,
		LockRatio:      "1:1",
		IsTwilightZone: true,
	}
	got := ComputeHabitability(body)
	// 10 + (-2) = 8
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (1:1 solar lock)", got.Rating)
	}
}

func TestComputeHabitability_TidalLockNot1to1_NoDM(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.TidalLock = &TidalLock{
		Case:           TidalLockCasePlanetToStar,
		LockRatio:      "3:2", // not 1:1
		IsTwilightZone: false,
	}
	got := ComputeHabitability(body)
	// 10 + 0 = 10
	if got.Rating != 10 {
		t.Errorf("got %d, want 10 (3:2 lock, no DM)", got.Rating)
	}
}

func TestComputeHabitability_NilBody_ZeroRating(t *testing.T) {
	got := ComputeHabitability(nil)
	if got.Rating != 0 {
		t.Errorf("got %d, want 0 (nil body)", got.Rating)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test -run TestComputeHabitability ./worlds/...
```

Expected: FAIL with "undefined: ComputeHabitability".

- [ ] **Step 3: Implement `ComputeHabitability` (skeleton + Size/Atm/Hydro/TidalLock DMs)**

Append to `worlds/habitability.go`:

```go
// ComputeHabitability per WBH p.132: 10 + DMs, clamped to [0, 12].
// Deterministic — no dice. Operates on body's current Atmosphere /
// Hydrographics / Temperature / Physical / SizeCode / TidalLock fields.
//
// Returns Habitability{Rating: 0} if body is nil. For bodies with
// missing pointer fields, the corresponding DMs are skipped (treated
// as 0) — defensive but documented as caller's responsibility.
//
// Skipped: low-oxygen-taint DM-2 deferred per spec Q3-a (taint
// typology not yet modeled).
func ComputeHabitability(body *DetailedPlacement) Habitability {
	if body == nil {
		return Habitability{Rating: 0}
	}
	dm := habitabilitySizeDM(SizeAsInt(body.SizeCode))
	dm += habitabilityAtmDM(body)
	dm += habitabilityHydroDM(body)
	dm += habitabilityTidalLockDM(body)

	// Temperature + Gravity DMs added in Task 3.

	rating := 10 + dm
	if rating < 0 {
		rating = 0
	}
	if rating > 12 {
		rating = 12
	}
	return Habitability{Rating: rating}
}

// habitabilitySizeDM per WBH p.132 size-DM table.
func habitabilitySizeDM(size int) int {
	switch {
	case size <= 4:
		return -1
	case size >= 9:
		return +1
	}
	return 0
}

// habitabilityAtmDM per WBH p.132 atmosphere-DM table.
// nil Atmosphere is treated as atm code 0 (vacuum) → DM-8.
func habitabilityAtmDM(body *DetailedPlacement) int {
	atmCode := 0
	if body.Atmosphere != nil {
		atmCode = body.Atmosphere.Code
	}
	switch atmCode {
	case 0, 1, 10: // 0, 1, A
		return -8
	case 2, 14: // 2, E
		return -4
	case 3, 13: // 3, D
		return -3
	case 4, 9:
		return -2
	case 5, 7, 8:
		return -1
	case 6:
		return 0 // baseline
	case 11: // B
		return -10
	case 12, 15: // C, F+
		return -12
	}
	return 0
}

// habitabilityHydroDM per WBH p.132 hydrographics-DM table.
// nil Hydrographics is treated as Hydro code 0 → DM-4.
func habitabilityHydroDM(body *DetailedPlacement) int {
	hydroCode := 0
	if body.Hydrographics != nil {
		hydroCode = body.Hydrographics.Code
	}
	switch {
	case hydroCode == 0:
		return -4
	case hydroCode >= 1 && hydroCode <= 3:
		return -2
	case hydroCode == 9:
		return -1
	case hydroCode >= 10:
		return -2
	}
	return 0 // 4-8
}

// habitabilityTidalLockDM per WBH p.132: "Solar tidally locked (1:1)
// world" → DM-2. Detection: TidalLock.IsTwilightZone (which is true
// only when Case == PlanetToStar AND LockRatio == "1:1").
func habitabilityTidalLockDM(body *DetailedPlacement) int {
	if body.TidalLock == nil {
		return 0
	}
	if body.TidalLock.IsTwilightZone {
		return -2
	}
	return 0
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestComputeHabitability ./worlds/... -v
```

Expected: all 12 tests PASS.

- [ ] **Step 5: just check && just test**

Expected: 0 issues; all packages pass.

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/habitability.go \
        worlds/habitability_test.go
git commit -m "feat(worlds): ComputeHabitability — Size + Atm + Hydro + TidalLock DMs (WBH p.132)"
```

---

## Task 3: `ComputeHabitability` — Temperature + Gravity DMs

**Files:**

- Modify: `worlds/habitability.go`
- Modify: `worlds/habitability_test.go`

This task adds the remaining DM categories (Temperature + Gravity, including the WBH p.132 gravity-bands ambiguity per spec Q3-a) and verifies the Zed Prime worked example.

- [ ] **Step 1: Write failing tests**

Append to `worlds/habitability_test.go`:

```go
func TestComputeHabitability_ZedPrime(t *testing.T) {
	// Per WBH IISS form p.141:
	// Size 5, Atm 6, Hydro 6, no tidal lock,
	// MeanK 300, HighK 346 (>323 → -2), LowK 262,
	// Gravity 0.66 (0.4-0.7 narrower band wins → -1).
	// DMs: 0 + 0 + 0 + 0 + (-2) + 0 + 0 + (-1) = -3
	// Habitability = 10 - 3 = 7.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 6}
	body.Temperature = &Temperature{MeanK: 300, HighK: 346, LowK: 262}
	body.Physical = &BodyPhysical{Gravity: 0.66}
	got := ComputeHabitability(body)
	if got.Rating != 7 {
		t.Errorf("Zed Prime: got %d, want 7", got.Rating)
	}
}

func TestComputeHabitability_TerraEquivalent(t *testing.T) {
	// Size 8, Atm 6, Hydro 7 (no DM), no temp DMs, Gravity 1.0 (no DM).
	body := &DetailedPlacement{}
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Temperature = &Temperature{MeanK: 288, HighK: 315, LowK: 255}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	got := ComputeHabitability(body)
	if got.Rating != 10 {
		t.Errorf("Terra: got %d, want 10", got.Rating)
	}
}

func TestComputeHabitability_HighTempHotBand(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 300, HighK: 330, LowK: 280}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// HighK 330 > 323 → -2
	got := ComputeHabitability(body)
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (HighK >323)", got.Rating)
	}
}

func TestComputeHabitability_HighTempColdBand(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 270, HighK: 270, LowK: 260}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// HighK 270 < 279 → -2; MeanK 270 < 273 → -2; LowK 260 ≥ 200 → 0
	got := ComputeHabitability(body)
	if got.Rating != 6 {
		t.Errorf("got %d, want 6 (HighK <279, MeanK <273)", got.Rating)
	}
}

func TestComputeHabitability_MeanTempHottest(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 350, HighK: 380, LowK: 320}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// MeanK 350 > 323 → -4; HighK 380 > 323 → -2
	got := ComputeHabitability(body)
	if got.Rating != 4 {
		t.Errorf("got %d, want 4", got.Rating)
	}
}

func TestComputeHabitability_MeanTempBoundary323(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 323, HighK: 323, LowK: 280}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// MeanK 323 is in [304, 323] → -2 (NOT -4; >323 is strict).
	// HighK 323 → no DM (>323 is strict).
	got := ComputeHabitability(body)
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (boundary 323)", got.Rating)
	}
}

func TestComputeHabitability_MeanTempBoundary324(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 324, HighK: 324, LowK: 280}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// MeanK 324 > 323 → -4; HighK 324 > 323 → -2
	got := ComputeHabitability(body)
	if got.Rating != 4 {
		t.Errorf("got %d, want 4 (boundary 324)", got.Rating)
	}
}

func TestComputeHabitability_LowTempBelow200(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 280, HighK: 290, LowK: 195}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// LowK 195 < 200 → -2
	got := ComputeHabitability(body)
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (LowK <200)", got.Rating)
	}
}

func TestComputeHabitability_GravityNarrowerBandWins(t *testing.T) {
	// Gravity 0.5 → in BOTH 0.2-0.7 (-2) AND 0.4-0.7 (-1) → narrower wins → -1.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 0.5}
	got := ComputeHabitability(body)
	if got.Rating != 9 {
		t.Errorf("got %d, want 9 (gravity 0.5 narrower band -1)", got.Rating)
	}
}

func TestComputeHabitability_GravityResidualLowBand(t *testing.T) {
	// Gravity 0.3 → in 0.2-0.7 only (NOT in 0.4-0.7) → residual band → -2.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 0.3}
	got := ComputeHabitability(body)
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (gravity 0.3 residual -2)", got.Rating)
	}
}

func TestComputeHabitability_GravityComfortable(t *testing.T) {
	// Gravity 0.8 → in 0.7-0.9 → DM+1.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 0.8}
	got := ComputeHabitability(body)
	if got.Rating != 11 {
		t.Errorf("got %d, want 11 (gravity 0.8 +1)", got.Rating)
	}
}

func TestComputeHabitability_GravityHigh(t *testing.T) {
	// Gravity 1.5 → in 1.4-2.0 → DM-3.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 1.5}
	got := ComputeHabitability(body)
	if got.Rating != 7 {
		t.Errorf("got %d, want 7 (gravity 1.5 -3)", got.Rating)
	}
}

func TestComputeHabitability_GravityCrushing(t *testing.T) {
	// Gravity 2.5 → > 2.0 → DM-6.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 2.5}
	got := ComputeHabitability(body)
	if got.Rating != 4 {
		t.Errorf("got %d, want 4 (gravity 2.5 -6)", got.Rating)
	}
}

func TestComputeHabitability_GravityVeryLow(t *testing.T) {
	// Gravity 0.1 → < 0.2 → DM-4.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 0.1}
	got := ComputeHabitability(body)
	if got.Rating != 6 {
		t.Errorf("got %d, want 6 (gravity 0.1 -4)", got.Rating)
	}
}

func TestComputeHabitability_GravityEarthBaseline(t *testing.T) {
	// Gravity 1.0 → in 0.9-1.1 → no DM.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	got := ComputeHabitability(body)
	if got.Rating != 10 {
		t.Errorf("got %d, want 10 (gravity 1.0 baseline)", got.Rating)
	}
}

func TestComputeHabitability_UndefinedGravity_Size6(t *testing.T) {
	// Physical nil, Size 6 → +1 - |6-6| = +1.
	body := &DetailedPlacement{}
	body.SizeCode = "6"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = nil
	got := ComputeHabitability(body)
	if got.Rating != 11 {
		t.Errorf("got %d, want 11 (undefined gravity Size 6 → +1)", got.Rating)
	}
}

func TestComputeHabitability_UndefinedGravity_Size0(t *testing.T) {
	// Physical nil, Size 0 → +1 - |6-0| = -5.
	// Plus Size 0 (in 0-4) → -1.
	body := &DetailedPlacement{}
	body.SizeCode = "0"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = nil
	got := ComputeHabitability(body)
	// 10 - 1 (size) - 5 (undefined gravity) = 4
	if got.Rating != 4 {
		t.Errorf("got %d, want 4 (undefined gravity Size 0 → -5)", got.Rating)
	}
}

func TestComputeHabitability_HabitabilityCannotExceed12(t *testing.T) {
	// Synthetic max-positive: Size 9 (+1), Atm 6 (0), Hydro 7 (0), Gravity 0.8 (+1).
	// 10 + 2 = 12.
	body := &DetailedPlacement{}
	body.SizeCode = "9"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Gravity: 0.8}
	got := ComputeHabitability(body)
	if got.Rating != 12 {
		t.Errorf("got %d, want 12 (max positive)", got.Rating)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run TestComputeHabitability_Zed ./worlds/...
```

Expected: FAIL (Zed Prime test fails — no temperature/gravity DMs yet).

- [ ] **Step 3: Add Temperature + Gravity DMs to `ComputeHabitability`**

In `worlds/habitability.go`, modify `ComputeHabitability` to add the two new DM categories before the clamp:

```go
func ComputeHabitability(body *DetailedPlacement) Habitability {
	if body == nil {
		return Habitability{Rating: 0}
	}
	dm := habitabilitySizeDM(SizeAsInt(body.SizeCode))
	dm += habitabilityAtmDM(body)
	dm += habitabilityHydroDM(body)
	dm += habitabilityTidalLockDM(body)
	dm += habitabilityTempDM(body)
	dm += habitabilityGravityDM(body)

	rating := 10 + dm
	if rating < 0 {
		rating = 0
	}
	if rating > 12 {
		rating = 12
	}
	return Habitability{Rating: rating}
}
```

Append the two new helper functions at the bottom of `worlds/habitability.go`:

```go
// habitabilityTempDM per WBH p.132 temperature-DM table.
// Returns 0 when Temperature is nil (defensive).
//
// Note: HighK > 323 and MeanK > 323 are strict (323 itself is in the
// [304, 323] band → -2, NOT in the >323 band → -4). Per WBH p.132 footnote,
// "use worst at edges" — but the bands as written are unambiguous at 323.
func habitabilityTempDM(body *DetailedPlacement) int {
	if body.Temperature == nil {
		return 0
	}
	dm := 0
	t := body.Temperature
	if t.HighK > 323 {
		dm += -2
	}
	if t.HighK > 0 && t.HighK < 279 {
		dm += -2
	}
	if t.MeanK > 323 {
		dm += -4
	} else if t.MeanK >= 304 && t.MeanK <= 323 {
		dm += -2
	}
	if t.MeanK > 0 && t.MeanK < 273 {
		dm += -2
	}
	if t.LowK > 0 && t.LowK < 200 {
		dm += -2
	}
	return dm
}

// habitabilityGravityDM per WBH p.132 gravity-DM table.
//
// WBH p.132 has overlapping bands (0.2-0.7 and 0.4-0.7). Per the worked
// example for Zed Prime (gravity 0.66 → DM-1, NOT -2), the narrower band
// wins. Documented as a WBH inconsistency (footnote contradicts worked
// example); implementation follows the worked example.
//
// Undefined gravity (Physical nil): per WBH "+1 - |6 - Size|".
func habitabilityGravityDM(body *DetailedPlacement) int {
	if body.Physical == nil {
		size := SizeAsInt(body.SizeCode)
		diff := 6 - size
		if diff < 0 {
			diff = -diff
		}
		return 1 - diff
	}
	g := body.Physical.Gravity
	switch {
	case g < 0.2:
		return -4
	case g >= 0.7 && g <= 0.9:
		return +1
	case g >= 0.4 && g < 0.7:
		return -1 // narrower band; wins over 0.2-0.7 per Q3-a
	case g >= 0.2 && g < 0.4:
		return -2 // residual of 0.2-0.7
	case g > 1.1 && g <= 1.4:
		return -1
	case g > 1.4 && g <= 2.0:
		return -3
	case g > 2.0:
		return -6
	}
	return 0 // 0.9-1.1 (Earth-like baseline)
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestComputeHabitability ./worlds/... -v
```

Expected: all 30 tests PASS (12 from Task 2 + 18 from Task 3).

- [ ] **Step 5: just check && just test**

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/habitability.go \
        worlds/habitability_test.go
git commit -m "feat(worlds): ComputeHabitability — Temperature + Gravity DMs (WBH p.132)"
```

---

## Task 4: `pickMainworld` — priority-chain selector

**Files:**

- Create: `worlds/mainworld.go`
- Create: `worlds/mainworld_test.go`

- [ ] **Step 1: Write failing tests**

Create `worlds/mainworld_test.go`:

```go
package worlds

import "testing"

func TestPickMainworld_SophontWins_OverHigherHabitability(t *testing.T) {
	// Body A: Habitability 8, no sophont.
	// Body B: Habitability 4, HasNativeSophont=true.
	// Sophont wins → B.
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 8}, Biology: &Biology{}},
		{Designation: "B", Habitability: &Habitability{Rating: 4}, Biology: &Biology{HasNativeSophont: true}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "B" {
		t.Errorf("got %q, want B", got)
	}
}

func TestPickMainworld_SophontTied_HigherHabitabilityWins(t *testing.T) {
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 5}, Biology: &Biology{HasNativeSophont: true}},
		{Designation: "B", Habitability: &Habitability{Rating: 8}, Biology: &Biology{HasNativeSophont: true}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "B" {
		t.Errorf("got %q, want B (higher habitability)", got)
	}
}

func TestPickMainworld_NoSophont_HighestHabitability(t *testing.T) {
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 4}},
		{Designation: "B", Habitability: &Habitability{Rating: 8}},
		{Designation: "C", Habitability: &Habitability{Rating: 6}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "B" {
		t.Errorf("got %q, want B", got)
	}
}

func TestPickMainworld_HabitabilityTied_HighestResourceWins(t *testing.T) {
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 7}, Biology: &Biology{ResourceRating: 5}},
		{Designation: "B", Habitability: &Habitability{Rating: 7}, Biology: &Biology{ResourceRating: 8}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "B" {
		t.Errorf("got %q, want B (resource tiebreaker)", got)
	}
}

func TestPickMainworld_AllZero_FirstTerrestrialFallback(t *testing.T) {
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 0}, Biology: &Biology{ResourceRating: 0}},
		{Designation: "B", Habitability: &Habitability{Rating: 0}, Biology: &Biology{ResourceRating: 0}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "A" {
		t.Errorf("got %q, want A (first terrestrial fallback)", got)
	}
}

func TestPickMainworld_BeltsAndGGsOnly_EmptyString(t *testing.T) {
	detailed := []DetailedPlacement{
		{Designation: "Belt", Body: BodyPlanetoidBelt, SizeCode: "0"},
		{Designation: "GG", Body: BodyGasGiant},
	}
	got := pickMainworld(detailed)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestPickMainworld_MoonAsMainworld(t *testing.T) {
	// Planet has habitability 4; moon has habitability 9 → moon wins.
	detailed := []DetailedPlacement{
		{
			Designation:  "Aab IV",
			Body:         BodyTerrestrial,
			Habitability: &Habitability{Rating: 4},
			Moons: []Moon{
				{Designation: "Aab IV d", Habitability: &Habitability{Rating: 9}},
			},
		},
	}
	got := pickMainworld(detailed)
	if got != "Aab IV d" {
		t.Errorf("got %q, want Aab IV d (moon)", got)
	}
}

func TestPickMainworld_EmptyDetailed_EmptyString(t *testing.T) {
	got := pickMainworld([]DetailedPlacement{})
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestPickMainworld_OnlyResourceNoHabitability(t *testing.T) {
	// All Habitability=0; pick by ResourceRating.
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 0}, Biology: &Biology{ResourceRating: 5}},
		{Designation: "B", Habitability: &Habitability{Rating: 0}, Biology: &Biology{ResourceRating: 9}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "B" {
		t.Errorf("got %q, want B (resource fallback)", got)
	}
}

func TestPickMainworld_ExtinctSophont_Counts(t *testing.T) {
	// Body A: extant sophont; Body B: extinct sophont; Body C: nothing.
	// Either A or B can win the sophont category; tiebreaker is habitability.
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 5}, Biology: &Biology{HasNativeSophont: true}},
		{Designation: "B", Habitability: &Habitability{Rating: 7}, Biology: &Biology{HadExtinctSophont: true}},
		{Designation: "C", Habitability: &Habitability{Rating: 9}, Biology: &Biology{}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	// A and B both qualify as "sophont"; B has higher habitability → B.
	if got != "B" {
		t.Errorf("got %q, want B (extinct sophont with higher habitability)", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run TestPickMainworld ./worlds/...
```

Expected: FAIL with "undefined: pickMainworld".

- [ ] **Step 3: Implement `pickMainworld`**

Create `worlds/mainworld.go`:

```go
// Package worlds — mainworld auto-pick per WBH p.134
// (sub-project 3B-final).
package worlds

// pickMainworld returns the designation of the auto-picked mainworld per
// WBH p.134. Priority chain (first match wins):
//
//  1. Bodies with native sophonts (extant or extinct); among these,
//     highest Habitability; tiebreaker: highest ResourceRating;
//     final tiebreaker: iteration order.
//  2. Highest Habitability among non-sophont bodies; tiebreakers same.
//  3. Highest ResourceRating if no body has Habitability > 0.
//  4. First terrestrial body in iteration order.
//
// Iterates BOTH detailed[i] AND dp.Moons[j]. Returns "" if no
// terrestrial body qualifies.
//
// "Best refuelling location" criterion (WBH p.134) deferred — depends on
// starport infrastructure from pp.147-234.
func pickMainworld(detailed []DetailedPlacement) string {
	type candidate struct {
		designation    string
		habitability   int
		resource       int
		hasSophont     bool
		isTerrestrial  bool
	}

	var candidates []candidate
	collect := func(designation string, body BodyType, h *Habitability, b *Biology) {
		if body != BodyTerrestrial {
			return
		}
		c := candidate{
			designation:   designation,
			isTerrestrial: true,
		}
		if h != nil {
			c.habitability = h.Rating
		}
		if b != nil {
			c.resource = b.ResourceRating
			c.hasSophont = b.HasNativeSophont || b.HadExtinctSophont
		}
		candidates = append(candidates, c)
	}

	for i := range detailed {
		dp := &detailed[i]
		collect(dp.Designation, dp.Body, dp.Habitability, dp.Biology)
		// Moons: treat as terrestrial candidates if they have at least one of
		// the relevant fields populated. Moons get Habitability + Biology
		// from the same Step 5G/5F passes that planets do.
		for j := range dp.Moons {
			m := &dp.Moons[j]
			// Moons are conceptually terrestrials for mainworld purposes.
			collect(m.Designation, BodyTerrestrial, m.Habitability, m.Biology)
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	// Priority 1: sophont present.
	best := -1
	for i, c := range candidates {
		if !c.hasSophont {
			continue
		}
		if best == -1 ||
			c.habitability > candidates[best].habitability ||
			(c.habitability == candidates[best].habitability && c.resource > candidates[best].resource) {
			best = i
		}
	}
	if best != -1 {
		return candidates[best].designation
	}

	// Priority 2: highest habitability.
	for i, c := range candidates {
		if c.habitability == 0 {
			continue
		}
		if best == -1 ||
			c.habitability > candidates[best].habitability ||
			(c.habitability == candidates[best].habitability && c.resource > candidates[best].resource) {
			best = i
		}
	}
	if best != -1 {
		return candidates[best].designation
	}

	// Priority 3: highest resource.
	for i, c := range candidates {
		if c.resource == 0 {
			continue
		}
		if best == -1 || c.resource > candidates[best].resource {
			best = i
		}
	}
	if best != -1 {
		return candidates[best].designation
	}

	// Priority 4: first terrestrial.
	return candidates[0].designation
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestPickMainworld ./worlds/... -v
```

Expected: all 10 tests PASS.

- [ ] **Step 5: just check && just test**

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/mainworld.go \
        worlds/mainworld_test.go
git commit -m "feat(worlds): pickMainworld auto-pick with priority chain (WBH p.134)"
```

---

## Task 5: `runStep5G` orchestrator + DetailSystem wiring + mainworld pick

**Files:**

- Create: `worlds/system_detail_step5g.go`
- Create: `worlds/system_detail_step5g_test.go`
- Modify: `worlds/system_detail.go`

- [ ] **Step 1: Write failing orchestrator tests**

Create `worlds/system_detail_step5g_test.go`:

```go
package worlds

import (
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestRunStep5G_TerrestrialPopulatesHabitability(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "5"
	dp.Designation = "Aab III"
	dp.Atmosphere = &Atmosphere{Code: 6}
	dp.Hydrographics = &Hydrographics{Code: 6}
	dp.Physical = &BodyPhysical{Gravity: 1.0}
	dp.Temperature = &Temperature{MeanK: 290, HighK: 310, LowK: 270}

	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability == nil {
		t.Fatal("Habitability is nil")
	}
	if detailed[0].Habitability.Rating < 0 || detailed[0].Habitability.Rating > 12 {
		t.Errorf("Rating %d out of [0, 12]", detailed[0].Habitability.Rating)
	}
}

func TestRunStep5G_GasGiant_NoHabitability(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyGasGiant
	dp.GGClass = GasGiantSmall
	dp.Designation = "Aab IV"
	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability != nil {
		t.Error("GG should not get Habitability")
	}
}

func TestRunStep5G_BeltSize0_NoHabitability(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyPlanetoidBelt
	dp.SizeCode = "0"
	dp.Designation = "Aab Belt"
	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability != nil {
		t.Error("Belt should not get Habitability")
	}
}

func TestRunStep5G_BodyEmpty_NoOp(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyEmpty
	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability != nil {
		t.Error("Empty body should not get Habitability")
	}
}

func TestRunStep5G_MoonRecursion(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "8"
	dp.Designation = "Aab III"
	dp.Atmosphere = &Atmosphere{Code: 6}
	dp.Hydrographics = &Hydrographics{Code: 6}
	dp.Physical = &BodyPhysical{Gravity: 1.0}
	dp.Temperature = &Temperature{MeanK: 290, HighK: 310, LowK: 270}
	dp.Moons = []Moon{
		{
			Designation:   "Aab III a",
			SizeCode:      "5",
			Atmosphere:    &Atmosphere{Code: 6},
			Hydrographics: &Hydrographics{Code: 5},
			Physical:      &BodyPhysical{Gravity: 0.7},
			Temperature:   &Temperature{MeanK: 290, HighK: 310, LowK: 270},
		},
	}

	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability == nil {
		t.Fatal("Parent Habitability is nil")
	}
	if detailed[0].Moons[0].Habitability == nil {
		t.Fatal("Moon Habitability is nil")
	}
}

func TestRunStep5G_VacuumWorld_HasHabitability(t *testing.T) {
	// Atm nil → still gets Habitability (vacuum atm 0 DM-8 applied).
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "5"
	dp.Designation = "Aab III"
	dp.Atmosphere = nil // vacuum
	dp.Physical = &BodyPhysical{Gravity: 1.0}

	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability == nil {
		t.Fatal("Vacuum world should still have Habitability")
	}
	// 10 - 8 (vacuum) - 4 (nil hydro) = -2 → clamp 0
	if detailed[0].Habitability.Rating != 0 {
		t.Errorf("Rating: got %d, want 0 (vacuum + nil hydro → 0 clamped)", detailed[0].Habitability.Rating)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run TestRunStep5G ./worlds/...
```

Expected: FAIL with "undefined: runStep5G".

- [ ] **Step 3: Implement `runStep5G` + helpers**

Create `worlds/system_detail_step5g.go`:

```go
package worlds

import (
	"wbh/roller"
	"wbh/stars"
)

// runStep5G applies the 3B-final habitability pass: per-body Habitability
// rating per WBH p.132. Mutates detailed in place.
//
// Body filter: terrestrials only (skip belts/GGs/empty). Atmosphere
// optional — vacuum worlds (nil atm) get DM-8 per the atm 0 row. Differs
// from Biology's filter (which required atm).
//
// Per-body dice budget: 0 dice. Habitability is deterministic. Same per moon.
//
//nolint:unparam // matches sibling runStep5* signatures (always-nil error; r unused)
func runStep5G(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error {
	_ = r   // unused — Habitability is deterministic
	_ = sys // unused
	for i := range detailed {
		dp := &detailed[i]
		if !habitabilityApplies(dp) {
			continue
		}
		h := ComputeHabitability(dp)
		dp.Habitability = &h

		for j := range dp.Moons {
			m := &dp.Moons[j]
			moonDP := buildMoonPlacementView(m, dp)
			// buildMoonPlacementView does not copy Temperature or TidalLock;
			// alias them so ComputeHabitability sees the moon's own values.
			moonDP.Temperature = m.Temperature
			moonDP.TidalLock = m.TidalLock
			if !habitabilityApplies(moonDP) {
				continue
			}
			mh := ComputeHabitability(moonDP)
			m.Habitability = &mh
		}
	}
	return nil
}

// habitabilityApplies reports whether dp should receive a Habitability struct.
// True for terrestrial bodies (atmosphere optional — vacuum worlds get a rating).
// False for empty, belts, gas giants.
func habitabilityApplies(dp *DetailedPlacement) bool {
	if dp == nil || dp.Body == BodyEmpty {
		return false
	}
	if dp.Body == BodyGasGiant || dp.Body == BodyPlanetoidBelt {
		return false
	}
	if dp.SizeCode == "0" {
		return false
	}
	return true
}
```

- [ ] **Step 4: Wire `runStep5G` + mainworld pick into `DetailSystem`**

In `worlds/system_detail.go`, find the existing block:

```go
	// Step 5F — 3B-biology pass: native lifeform ratings + resource rating.
	if err := runStep5F(r, detailed, sys); err != nil {
		return SystemDetail{}, err
	}
```

Append immediately after it:

```go
	// Step 5G — 3B-final pass: per-body habitability rating.
	if err := runStep5G(r, detailed, sys); err != nil {
		return SystemDetail{}, err
	}
```

Then find where `sd` is constructed (`sd := SystemDetail{...}`). After `sd.Survey = ...` (or wherever the SystemDetail is fully populated), add the mainworld pick:

```go
	sd.MainworldDesignation = pickMainworld(detailed)
```

(Place this BEFORE the final `return sd, nil` and AFTER all other field assignments on `sd`.)

- [ ] **Step 5: Run all tests**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
just check
just test
```

Expected: 0 issues; all packages pass. The existing `TestZed_FullDetail_3A2b` should still pass — runStep5G is wired in but Task 9's new assertions haven't landed yet.

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/system_detail_step5g.go \
        worlds/system_detail_step5g_test.go \
        worlds/system_detail.go
git commit -m "feat(worlds): runStep5G orchestrator + mainworld pick + DetailSystem wiring (WBH pp.132-134)"
```

---

## Task 6: `RenderIISSClass4P` — World/Sector/Orbit/Size/Atmosphere sections

**Files:**

- Create: `worlds/iiss_class4p.go`
- Create: `worlds/iiss_class4p_test.go`

This task implements the renderer's first 5 sections plus the function skeleton (nil-handling, body type check, header). Tasks 7 and 8 add the remaining sections.

- [ ] **Step 1: Write failing tests**

Create `worlds/iiss_class4p_test.go`:

```go
package worlds

import (
	"strings"
	"testing"

	"wbh/stars"
)

func TestRenderIISSClass4P_NilBody_Empty(t *testing.T) {
	got := RenderIISSClass4P(nil, stars.System{}, "")
	if got != "" {
		t.Errorf("got %q, want empty (nil body)", got)
	}
}

func TestRenderIISSClass4P_BodyEmpty_Empty(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyEmpty
	got := RenderIISSClass4P(body, stars.System{}, "")
	if got != "" {
		t.Errorf("got %q, want empty (BodyEmpty)", got)
	}
}

func TestRenderIISSClass4P_Header_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab IV d"
	body.SizeCode = "5"
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "IISS CLASS IV SURVEY") {
		t.Errorf("missing header: got %q", got)
	}
	if !strings.Contains(got, "Aab IV d") {
		t.Errorf("missing designation: got %q", got)
	}
}

func TestRenderIISSClass4P_OrbitSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Eccentricity = 0.10
	body.Period = Period{Hours: 365.25 * 24}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "ORBIT") {
		t.Errorf("missing ORBIT section: got %q", got)
	}
}

func TestRenderIISSClass4P_SizeSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.DiameterKm = 8163
	body.MassEarth = 0.27
	body.Physical = &BodyPhysical{Density: 1.03, Gravity: 0.66}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "SIZE") {
		t.Errorf("missing SIZE section: got %q", got)
	}
	if !strings.Contains(got, "8163") {
		t.Errorf("missing diameter 8163: got %q", got)
	}
}

func TestRenderIISSClass4P_AtmosphereSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.042, ScaleHeight: 12.88, OxygenPartialPressure: 0.292}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "ATMOSPHERE") {
		t.Errorf("missing ATMOSPHERE section: got %q", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run TestRenderIISSClass4P ./worlds/...
```

Expected: FAIL with "undefined: RenderIISSClass4P".

- [ ] **Step 3: Implement skeleton + first 5 sections**

Create `worlds/iiss_class4p.go`:

```go
// Package worlds — IISS Class IV-P Survey Form rendering per WBH pp.138-146
// (sub-project 3B-final).
package worlds

import (
	"fmt"
	"strings"

	"wbh/stars"
)

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
// rendering is deferred (see spec carry-forwards).
func RenderIISSClass4P(body *DetailedPlacement, sys stars.System, mainworldDesignation string) string {
	if body == nil || body.Body == BodyEmpty {
		return ""
	}
	// Belt stub.
	if body.SizeCode == "0" {
		return renderBeltStub(body)
	}

	var sb strings.Builder
	sb.WriteString("IISS CLASS IV SURVEY — FORM 0407F-IV PART P\n\n")
	renderIISS4PWorld(&sb, body, sys)
	renderIISS4POrbit(&sb, body)
	renderIISS4PSize(&sb, body)
	renderIISS4PAtmosphere(&sb, body)

	// Tasks 7 and 8 append the remaining sections + Comments.
	return sb.String()
}

func renderBeltStub(body *DetailedPlacement) string {
	return fmt.Sprintf(`IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B (NOT YET IMPLEMENTED)

WORLD: %s   (Belt)
Use BeltDetails for resource/composition data.
`, body.Designation)
}

func renderIISS4PWorld(sb *strings.Builder, body *DetailedPlacement, sys stars.System) {
	fmt.Fprintf(sb, "WORLD: %s\n", body.Designation)
	fmt.Fprintf(sb, "SECTOR | LOCATION:    INITIAL SURVEY:    LAST UPDATED:\n")
	fmt.Fprintf(sb, "PRIMARY OBJECT(S):    SYSTEM AGE (Gyr): %.3f    TRAVEL ZONE:\n\n",
		sys.Primary.AgeGyr)
}

func renderIISS4POrbit(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("ORBIT\n")
	fmt.Fprintf(sb, "  AU: %.2f   Eccentricity: %.2f   Period (h): %.2f\n\n",
		stars.OrbitToAU(body.Orbit), body.Eccentricity, body.Period.Hours)
}

func renderIISS4PSize(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("SIZE\n")
	density := 0.0
	gravity := 0.0
	if body.Physical != nil {
		density = body.Physical.Density
		gravity = body.Physical.Gravity
	}
	fmt.Fprintf(sb, "  Diameter (km): %.0f   Density: %.2f   Gravity: %.2f   Mass: %.2f\n\n",
		body.DiameterKm, density, gravity, body.MassEarth)
}

func renderIISS4PAtmosphere(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("ATMOSPHERE\n")
	if body.Atmosphere == nil {
		sb.WriteString("  (none — vacuum)\n\n")
		return
	}
	atm := body.Atmosphere
	fmt.Fprintf(sb, "  Code: %d   Pressure (bar): %.3f   O2 (bar): %.3f   Scale Height: %.2f\n\n",
		atm.Code, atm.Pressure, atm.OxygenPartialPressure, atm.ScaleHeight)
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestRenderIISSClass4P ./worlds/... -v
```

Expected: all 6 tests PASS (the four section presence tests + the nil/empty defenses).

- [ ] **Step 5: just check && just test**

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/iiss_class4p.go \
        worlds/iiss_class4p_test.go
git commit -m "feat(worlds): RenderIISSClass4P — World/Orbit/Size/Atmosphere sections (WBH p.138)"
```

---

## Task 7: `RenderIISSClass4P` — Hydrographics/Rotation/Temperature/Seismic sections

**Files:**

- Modify: `worlds/iiss_class4p.go`
- Modify: `worlds/iiss_class4p_test.go`

- [ ] **Step 1: Write failing tests**

Append to `worlds/iiss_class4p_test.go`:

```go
func TestRenderIISSClass4P_HydrographicsSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Hydrographics = &Hydrographics{Code: 6, Percent: 62, Profile: "H6:H2O-100"}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "HYDROGRAPHICS") {
		t.Errorf("missing HYDROGRAPHICS section: got %q", got)
	}
	if !strings.Contains(got, "62") {
		t.Errorf("missing coverage 62: got %q", got)
	}
}

func TestRenderIISSClass4P_RotationSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.DayLength = &DayLength{SiderealHours: 24, SolarHours: 24, YearDays: 365}
	body.AxialTilt = &AxialTilt{Degrees: 23.5}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "ROTATION") {
		t.Errorf("missing ROTATION section: got %q", got)
	}
}

func TestRenderIISSClass4P_TemperatureSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Temperature = &Temperature{
		MeanK: 300, HighK: 346, LowK: 262,
		Luminosity: 1.419, Albedo: 0.33, GreenhouseFactor: 0.59,
	}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "TEMPERATURE") {
		t.Errorf("missing TEMPERATURE section: got %q", got)
	}
	if !strings.Contains(got, "300") {
		t.Errorf("missing MeanK 300: got %q", got)
	}
}

func TestRenderIISSClass4P_SeismicSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Geology = &Geology{
		ResidualSeismicStress: 0,
		TidalStressFactor:     3,
		TidalHeatingFactor:    14,
		TotalSeismicStress:    17,
		TectonicPlates:        4,
	}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "SEISMIC") {
		t.Errorf("missing SEISMIC section: got %q", got)
	}
	if !strings.Contains(got, "17") {
		t.Errorf("missing TSS 17: got %q", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run "TestRenderIISSClass4P_(Hydro|Rotation|Temperature|Seismic)" ./worlds/...
```

Expected: FAIL (sections not yet rendered).

- [ ] **Step 3: Add the four new section renderers**

Modify `RenderIISSClass4P` in `worlds/iiss_class4p.go` to add four new section calls AFTER `renderIISS4PAtmosphere`:

```go
	renderIISS4PHydrographics(&sb, body)
	renderIISS4PRotation(&sb, body)
	renderIISS4PTemperature(&sb, body)
	renderIISS4PSeismic(&sb, body)
```

Append the four new helper functions to `worlds/iiss_class4p.go`:

```go
func renderIISS4PHydrographics(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("HYDROGRAPHICS\n")
	if body.Hydrographics == nil {
		sb.WriteString("  (none)\n\n")
		return
	}
	hydro := body.Hydrographics
	fmt.Fprintf(sb, "  Code: %d   Coverage (%%): %d   Profile: %s\n\n",
		hydro.Code, hydro.Percent, hydro.Profile)
}

func renderIISS4PRotation(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("ROTATION\n")
	if body.DayLength != nil {
		fmt.Fprintf(sb, "  Sidereal (h): %.2f   Solar (h): %.2f   Solar days/year: %.2f\n",
			body.DayLength.SiderealHours, body.DayLength.SolarHours, body.DayLength.YearDays)
	}
	if body.AxialTilt != nil {
		fmt.Fprintf(sb, "  Axial Tilt: %.2f°\n", body.AxialTilt.Degrees)
	}
	tidalLockText := "no"
	if body.TidalLock != nil && body.TidalLock.LockRatio != "" {
		tidalLockText = body.TidalLock.LockRatio
	}
	tidesM := 0.0
	if body.TidalEffects != nil {
		tidesM = body.TidalEffects.Total
	}
	fmt.Fprintf(sb, "  Tidal lock: %s   Tides (m): %.2f\n\n", tidalLockText, tidesM)
}

func renderIISS4PTemperature(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("TEMPERATURE\n")
	if body.Temperature == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	t := body.Temperature
	fmt.Fprintf(sb, "  High (K): %.1f   Mean (K): %.1f   Low (K): %.1f\n",
		t.HighK, t.MeanK, t.LowK)
	fmt.Fprintf(sb, "  Luminosity: %.3f   Albedo: %.2f   Greenhouse: %.2f\n\n",
		t.Luminosity, t.Albedo, t.GreenhouseFactor)
}

func renderIISS4PSeismic(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("SEISMIC\n")
	if body.Geology == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	g := body.Geology
	fmt.Fprintf(sb, "  TSS: %d   Residual: %d   TidalStress: %d   TidalHeating: %d   Plates: %d\n\n",
		g.TotalSeismicStress, g.ResidualSeismicStress, g.TidalStressFactor, g.TidalHeatingFactor, g.TectonicPlates)
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestRenderIISSClass4P ./worlds/... -v
```

Expected: all 10 tests PASS (6 from Task 6 + 4 new).

- [ ] **Step 5: just check && just test**

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/iiss_class4p.go \
        worlds/iiss_class4p_test.go
git commit -m "feat(worlds): RenderIISSClass4P — Hydro/Rotation/Temperature/Seismic sections"
```

---

## Task 8: `RenderIISSClass4P` — Life/Resources/Habitability/Subordinates/Comments

**Files:**

- Modify: `worlds/iiss_class4p.go`
- Modify: `worlds/iiss_class4p_test.go`

- [ ] **Step 1: Write failing tests**

Append to `worlds/iiss_class4p_test.go`:

```go
func TestRenderIISSClass4P_LifeSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Biology = &Biology{
		Biomass: 10, Biocomplexity: 5, HasNativeSophont: false,
		Biodiversity: 7, Compatibility: 6, ResourceRating: 11,
	}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "LIFE") {
		t.Errorf("missing LIFE section: got %q", got)
	}
	if !strings.Contains(got, "Biomass: A") {
		t.Errorf("missing Biomass: A (eHex): got %q", got)
	}
}

func TestRenderIISSClass4P_HabitabilitySection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Habitability = &Habitability{Rating: 7}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "HABITABILITY") {
		t.Errorf("missing HABITABILITY section: got %q", got)
	}
	if !strings.Contains(got, "Rating: 7") {
		t.Errorf("missing Rating: 7: got %q", got)
	}
}

func TestRenderIISSClass4P_NoMoons_NoSubordinatesSection(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	got := RenderIISSClass4P(body, stars.System{}, "")
	if strings.Contains(got, "SUBORDINATES") {
		t.Errorf("should not have SUBORDINATES section when no moons: got %q", got)
	}
}

func TestRenderIISSClass4P_WithMoons_SubordinatesRendered(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab IV"
	body.SizeCode = "5"
	body.Moons = []Moon{
		{Designation: "Aab IV a", SizeCode: "5", DiameterKm: 5000, OrbitKm: 3_920_000, Eccentricity: 0.10, PeriodHours: 624.69},
	}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "SUBORDINATES") {
		t.Errorf("missing SUBORDINATES section: got %q", got)
	}
	if !strings.Contains(got, "Aab IV a") {
		t.Errorf("missing moon designation: got %q", got)
	}
}

func TestRenderIISSClass4P_MainworldAnnotation(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab IV d"
	body.SizeCode = "5"
	got := RenderIISSClass4P(body, stars.System{}, "Aab IV d")
	if !strings.Contains(got, "system mainworld") {
		t.Errorf("missing mainworld annotation: got %q", got)
	}
}

func TestRenderIISSClass4P_NotMainworld_NoAnnotation(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab IV"
	body.SizeCode = "5"
	got := RenderIISSClass4P(body, stars.System{}, "Aab IV d")
	if strings.Contains(got, "system mainworld") {
		t.Errorf("should not have mainworld annotation: got %q", got)
	}
}

func TestRenderIISSClass4P_Belt_StubRendering(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyPlanetoidBelt
	body.Designation = "Aab Belt"
	body.SizeCode = "0"
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "NOT YET IMPLEMENTED") {
		t.Errorf("missing belt stub marker: got %q", got)
	}
	if !strings.Contains(got, "Aab Belt") {
		t.Errorf("missing belt designation: got %q", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run "TestRenderIISSClass4P_(Life|Habitability|Subordinates|Mainworld|NotMainworld|Belt)" ./worlds/...
```

Expected: FAIL (the new sections not yet rendered).

- [ ] **Step 3: Add the final section renderers**

Modify `RenderIISSClass4P` in `worlds/iiss_class4p.go` to add the final section calls AFTER `renderIISS4PSeismic`:

```go
	renderIISS4PLife(&sb, body)
	renderIISS4PResources(&sb, body)
	renderIISS4PHabitability(&sb, body)
	renderIISS4PSubordinates(&sb, body)
	renderIISS4PComments(&sb, body, mainworldDesignation)
```

Append the five new helper functions to `worlds/iiss_class4p.go`:

```go
func renderIISS4PLife(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("LIFE\n")
	if body.Biology == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	b := body.Biology
	sophontStr := "no"
	if b.HasNativeSophont {
		sophontStr = "yes"
	} else if b.HadExtinctSophont {
		sophontStr = "extinct"
	}
	fmt.Fprintf(sb, "  Biomass: %s   Biocomplexity: %d   Sophonts?: %s   Biodiversity: %d   Compatibility: %d\n\n",
		string(eHexDigit(b.Biomass)), b.Biocomplexity, sophontStr, b.Biodiversity, b.Compatibility)
}

func renderIISS4PResources(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("RESOURCES\n")
	if body.Biology == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	fmt.Fprintf(sb, "  Rating: %s\n\n", string(eHexDigit(body.Biology.ResourceRating)))
}

func renderIISS4PHabitability(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("HABITABILITY\n")
	if body.Habitability == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	fmt.Fprintf(sb, "  Rating: %d\n\n", body.Habitability.Rating)
}

func renderIISS4PSubordinates(sb *strings.Builder, body *DetailedPlacement) {
	if len(body.Moons) == 0 {
		return
	}
	sb.WriteString("SUBORDINATES\n")
	sb.WriteString("  Designation   SizeCode   Diameter (km)   Orbit (km)   Ecc   Period (h)\n")
	for _, m := range body.Moons {
		fmt.Fprintf(sb, "  %s   %s   %.0f   %d   %.3f   %.2f\n",
			m.Designation, m.SizeCode, m.DiameterKm, m.OrbitKm, m.Eccentricity, m.PeriodHours)
	}
	sb.WriteString("\n")
}

func renderIISS4PComments(sb *strings.Builder, body *DetailedPlacement, mainworldDesignation string) {
	sb.WriteString("COMMENTS\n")
	if mainworldDesignation != "" && body.Designation == mainworldDesignation {
		sb.WriteString("  This is the system mainworld.\n")
	}
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestRenderIISSClass4P ./worlds/... -v
```

Expected: all 17 tests PASS (10 prior + 7 new).

- [ ] **Step 5: just check && just test**

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/iiss_class4p.go \
        worlds/iiss_class4p_test.go
git commit -m "feat(worlds): RenderIISSClass4P — Life/Resources/Habitability/Subordinates/Comments + belt stub"
```

---

## Task 9: Acceptance test extension on `TestZed_FullDetail_3A2b`

**Files:**

- Modify: `worlds/worked_examples_test.go`

- [ ] **Step 1: Locate existing assertions and accumulator**

```bash
grep -n "Assertion 38\|totalBiomassNonzero\|3B-biology: %d body-iterations\|Compatibility formula" worlds/worked_examples_test.go | head -10
```

Identify:

- Line where Assertion 38's accumulation block ends inside the iter loop
- Line of the iter loop's closing `}`
- Lines of the existing 6 trailing `t.Logf` notes

- [ ] **Step 2: Append assertions 39-43 inside the iter loop**

After Assertion 38's accumulation block and BEFORE the iter loop's closing `}`, insert:

```go
		// 3B-final invariants (assertions 39-43).

		// Assertion 39: HasHabitability() for terrestrial bodies (and HZ-planet
		// moons). Skip belts, GGs, empty.
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.Body == worlds.BodyTerrestrial && dp.SizeCode != "0" {
				if !dp.HasHabitability() {
					t.Errorf("iter %d: terrestrial body %s missing Habitability", iter, dp.Designation)
				}
			}
			for j := range dp.Moons {
				m := &dp.Moons[j]
				if !m.HasHabitability() {
					t.Errorf("iter %d: moon %s missing Habitability", iter, m.Designation)
				}
			}
		}

		// Assertion 40: Habitability.Rating in [0, 12] for all bodies with Habitability.
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.HasHabitability() {
				r := dp.Habitability.Rating
				if r < 0 || r > 12 {
					t.Errorf("iter %d: body %s: Habitability.Rating=%d out of [0, 12]",
						iter, dp.Designation, r)
				}
			}
			for j := range dp.Moons {
				m := &dp.Moons[j]
				if m.HasHabitability() {
					r := m.Habitability.Rating
					if r < 0 || r > 12 {
						t.Errorf("iter %d: moon %s: Habitability.Rating=%d out of [0, 12]",
							iter, m.Designation, r)
					}
				}
			}
		}

		// Assertion 41: sd.MainworldDesignation != "" (Zed system has at least
		// one terrestrial body).
		if sd.MainworldDesignation == "" {
			t.Errorf("iter %d: MainworldDesignation is empty", iter)
		}

		// Assertion 42: MainworldDesignation matches an existing body or moon.
		found := false
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.Designation == sd.MainworldDesignation {
				found = true
				break
			}
			for j := range dp.Moons {
				if dp.Moons[j].Designation == sd.MainworldDesignation {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			t.Errorf("iter %d: MainworldDesignation %q matches no body or moon",
				iter, sd.MainworldDesignation)
		}

		// Assertion 43: RenderIISSClass4P for the mainworld produces non-empty
		// output containing "Habitability".
		var mainworldBody *worlds.DetailedPlacement
		for i := range sd.Detailed {
			if sd.Detailed[i].Designation == sd.MainworldDesignation {
				mainworldBody = &sd.Detailed[i]
				break
			}
		}
		if mainworldBody != nil {
			out := worlds.RenderIISSClass4P(mainworldBody, sys, sd.MainworldDesignation)
			if out == "" {
				t.Errorf("iter %d: RenderIISSClass4P returned empty string for mainworld %s",
					iter, sd.MainworldDesignation)
			}
			if out != "" && !strings.Contains(out, "HABITABILITY") {
				t.Errorf("iter %d: RenderIISSClass4P for %s missing HABITABILITY section",
					iter, sd.MainworldDesignation)
			}
		}
```

If `strings` is not yet imported in the test file, add it.

- [ ] **Step 3: Add 7th trailing `t.Logf` note**

After the existing 6 trailing `t.Logf` notes, append:

```go
	t.Logf("3B-final: Form 0407K-IV PART P.B (belt rendering) and World Maps (pp.135-137) deferred; miscellaneous habitability D3-1 referee adjustment skipped per YAGNI")
```

- [ ] **Step 4: Run the acceptance test**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test -run TestZed_FullDetail_3A2b ./worlds/... -v 2>&1 | tail -30
```

Expected: PASS across all 100 iterations.

- [ ] **Step 5: just check && just test**

```bash
just check
just test
```

Expected: 0 issues; all packages pass.

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/worked_examples_test.go
git commit -m "test(worlds): extend TestZed_FullDetail_3A2b with 3B-final assertions"
```

---

## Task 10: Final end-to-end review on Opus + merge

**Files:** none (review-only task)

- [ ] **Step 1: Verify branch state**

```bash
cd /Users/markayers/Documents/Traveller
git log --oneline main..HEAD
```

Expected: 9 commits (one per Task 1-9).

- [ ] **Step 2: Final review subagent (Opus)**

Dispatch `superpowers:code-reviewer` (Opus model) on the entire branch. Provide:

- Branch name: `feat/wbh-world-physical-3b-final`
- Spec path: `docs/specs/2026-05-05-world-physical-3b-final-design.md`
- Plan path: `docs/plans/2026-05-05-world-physical-3b-final.md`
- Diff command: `git -C /Users/markayers/Documents/Traveller diff main..feat/wbh-world-physical-3b-final -- `

Reviewer should report: spec-compliance issues, code-quality issues, cross-cutting concerns, merge readiness assessment. Specifically watch for C1-style integration silent-zero bugs (precedent: 3B-geology terrestrial MassEarth=0; 3B-biology moonDP.Temperature missing).

- [ ] **Step 3: Address review findings (if any)**

Fix any Critical/Important issues with additional commits on the branch. Re-run `just check && just test` after each fix.

- [ ] **Step 4: Confirm merge with user**

Show the user:

- Final commit count
- Brief summary of each task
- Any deviations from the spec
- Final review verdict
- Documented WBH inconsistencies to log as feedback memories after merge

Wait for explicit "merge" approval before proceeding.

- [ ] **Step 5: Merge to main**

```bash
git checkout main
git merge --no-ff feat/wbh-world-physical-3b-final -m "Merge feat/wbh-world-physical-3b-final: World Physical 3B-Final complete

WBH pp.132-146: Habitability Rating (per WBH p.132), Final Mainworld
Determination (p.134), and IISS Class IV-P Survey Form rendering
(Form 0407F-IV PART P, pp.138-146).

Implemented as new pipeline step runStep5G between 3B-biology and Step 6,
plus a system-wide mainworld pick after 5G, plus a per-body Class IV-P
form renderer.

The LAST physical-world sub-project. WBH pp.69-146 are now complete
end-to-end through DetailSystem. Next major chapter is World Social
Characteristics (pp.147-234) — different domain entirely.

Habitability struct on DetailedPlacement and Moon. MainworldDesignation
string on SystemDetail. RenderIISSClass4P called on demand by callers
(not stored on SystemDetail).

Body filter for 5G: terrestrials only (skip belts/GGs/empty); atmosphere
optional (vacuum gets atm 0 DM)."

just check && just test
```

- [ ] **Step 6: Update memory**

After merge:

1. Update `MEMORY.md` Subprojects line to mark 3B-final complete with merge SHA
2. Update `project_world_builder_3b_kickoff.md` to mark physical-world chapter complete; next is World Social Characteristics
3. Save book-inconsistency feedback memory:
   - `feedback_wbh_p132_gravity_dm_overlap.md` — gravity bands 0.2-0.7 / 0.4-0.7 overlap; footnote vs worked-example divergence; implementation follows worked example (narrower band wins)

- [ ] **Step 7: Confirm clean tree**

```bash
git status
```

Expected: clean.

---

## Self-review checklist (run after writing this plan)

- [x] **Spec coverage:** Every section of the spec maps to a task. Habitability formula → Tasks 2-3. Mainworld picker → Task 4. runStep5G + DetailSystem wiring → Task 5. Class IV-P form → Tasks 6-8 (split by section group). Acceptance test → Task 9. Final review → Task 10. Struct/field additions → Task 1.
- [x] **Placeholder scan:** No TBD/TODO/incomplete sections. All code blocks complete. Every step has runnable commands and concrete expected output.
- [x] **Type consistency:** `Habitability` struct fields (`Rating int`, `Notes string`) consistent across tasks. `ComputeHabitability(body) Habitability` signature consistent. `pickMainworld(detailed) string` signature consistent. `RenderIISSClass4P(body, sys, mainworldDesignation) string` signature consistent. `runStep5G(r, detailed, sys) error` matches sibling pattern.

## Known WBH inconsistency to log during implementation

After merge, save as feedback memory:

- **WBH p.132 gravity DM band overlap.** Bands `0.2-0.7 → DM-2` and `0.4-0.7 → DM-1` overlap. Footnote says "use worst at edges" but worked example (Zed Prime gravity 0.66) gets DM-1, contradicting the footnote. Implementation follows the worked example: narrower band wins (matches the verifiable Zed Prime IISS form on p.141).
