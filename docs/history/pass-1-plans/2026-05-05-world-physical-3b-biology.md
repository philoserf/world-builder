# World Physical 3B-Biology Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement WBH pp.127-131 (Biomass, Biocomplexity, Native Sophonts existence, Biodiversity, Compatibility, Native Lifeform Profile, Resource Rating) as a single new pipeline step `runStep5F` between 3B-geology and Step 6.

**Architecture:** A new `Biology` struct attached as `dp.Biology *Biology` (and `m.Biology *Biology` on Moon) holds all biology output. Seven standalone helper functions in `worlds/biology.go` do the rolls; a `Profile()` method derives the MXDC eHex string. `runStep5F` in `worlds/system_detail_steps.go` assembles them per body and per moon. Bodies filtered by `BodyTerrestrial` + `Atmosphere != nil`; GGs / belts / empty placements skipped.

**Tech Stack:** Go 1.26, existing `wbh/roller`, `wbh/stars`, `wbh/worlds` packages. Same workflow as 3B-geology: per-task subagent (Sonnet) → spec reviewer → code reviewer → next task. Final end-to-end review on Opus before merge.

---

## File map

| File                             | Status   | Purpose                                                                                           |
| -------------------------------- | -------- | ------------------------------------------------------------------------------------------------- |
| `worlds/biology.go`              | New      | `Biology` struct + 7 standalone helpers + `Profile()` method                                      |
| `worlds/biology_test.go`         | New      | Per-formula unit tests + orchestrator tests                                                       |
| `worlds/system_detail.go`        | Modified | Add `Biology *Biology` to `DetailedPlacement`; one new line in `DetailSystem` to call `runStep5F` |
| `worlds/system_detail_steps.go`  | Modified | Add `runStep5F` helper + `computeBiology` + `computeMoonBiology`                                  |
| `worlds/moons.go`                | Modified | Add `Biology *Biology` to `Moon` + `HasBiology()` accessor                                        |
| `worlds/worked_examples_test.go` | Modified | Append assertions 32-38 to `TestZed_FullDetail_3A2b` + 6th trailing t.Logf                        |

## Reference

- **Spec:** `docs/history/pass-1-specs/2026-05-05-world-physical-3b-biology-design.md` (commit `95c70ef`)
- **WBH source:** pp.127-131
- **Predecessor:** 3B-geology merged on `main` as `2ebced4`

## API gotchas (from prior sub-projects)

- `r.Roll("2D")` not `r.Roll(2, 6)`; constructor is `roller.NewScripted(...)` (variadic ints); `Roll` returns `int` with no error.
- `SizeAsInt(SizeCode) int` lives in `worlds/atmosphere.go`; converts `"0".."F"` → `0..15`.
- `dp.Atmosphere.Code` is `int` 0-15; `dp.Atmosphere.Subtype` is the digit/letter from the subtype roll.
- `dp.Hydrographics.Code` is `int` 0-10 (10 = "A").
- `dp.Temperature.MeanK` and `dp.Temperature.HighK` are populated post-3A2b/3B-geology (the latter via `ApplyInherentTempAddition`).
- `sys.Primary.AgeGyr` is the system age in billions of years.
- `dp.Physical.Density` is the density in Earth-units; `dp.Physical` is a pointer field (nil-safe).
- `dp.Body == BodyTerrestrial`, `BodyGasGiant`, `BodyEmpty`, `BodyPlanetoidBelt` (constants in `worlds/placement.go`).
- `dp.SizeCode == "0"` indicates a belt; SizeCode is `string` ("0".."F").
- `buildMoonPlacementView(m, parent)` (in `worlds/system_detail.go`) creates a moonDP synthetic view with Atmosphere/Hydrographics/Physical/Temperature pointer-aliased.
- Project's `just check` runs `go fix ./...` and FAILS if it produces unstaged changes — apply any modernize rewrites (don't dismiss as noise).
- Stale LSP "undefined" diagnostics are documented project noise — trust `just check && just test` exit codes.

## Final-review pattern

Per established precedent (3A2b-rederive, 3B-geology), the Opus final-gate review consistently catches integration-level Critical bugs that per-task reviews miss (silent-zero patterns, misnamed parameters, etc.). **Don't skip Task 11.** Per-task review checks the code matches the spec; Opus review checks the code matches reality.

---

## Task 1: Branch setup + `Biology` struct + struct field additions

**Files:**

- Create: `worlds/biology.go`
- Modify: `worlds/system_detail.go` (DetailedPlacement struct + HasBiology accessor)
- Modify: `worlds/moons.go` (Moon struct + HasBiology accessor)

- [ ] **Step 1: Create the branch from main**

```bash
cd /Users/markayers/Documents/Traveller
git checkout main
git pull --ff-only 2>/dev/null || true
git checkout -b feat/wbh-world-physical-3b-biology
```

- [ ] **Step 2: Create `worlds/biology.go` with the `Biology` struct**

```go
// Package worlds — biology (native lifeform ratings + resource rating)
// per WBH pp.127-131 (sub-project 3B-biology).
package worlds

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
```

- [ ] **Step 3: Add `Biology *Biology` field to `DetailedPlacement`**

In `worlds/system_detail.go`, find the `DetailedPlacement` struct definition. Add the new field at the END of the existing pointer-field group, after the `// 3B-geology additions` block:

```go
	// 3B-biology additions
	Biology *Biology
```

- [ ] **Step 4: Add `Biology *Biology` field to `Moon`**

In `worlds/moons.go`, find the `Moon` struct definition. Add the new field at the END, after the existing `// 3B-geology additions`:

```go
	// 3B-biology additions
	Biology *Biology
```

- [ ] **Step 5: Add `HasBiology()` accessors**

In `worlds/system_detail.go`, after the existing `HasGeology()` accessor, add:

```go
func (dp *DetailedPlacement) HasBiology() bool { return dp.Biology != nil }
```

In `worlds/moons.go`, after the existing `HasGeology()` accessor, add:

```go
// HasBiology reports whether biology data has been generated for this moon.
func (m *Moon) HasBiology() bool { return m.Biology != nil }
```

- [ ] **Step 6: Smoke check**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
just check
just test
```

Expected: 0 issues; all packages pass.

- [ ] **Step 7: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/biology.go \
        worlds/system_detail.go \
        worlds/moons.go
git commit -m "feat(worlds): Biology struct + DetailedPlacement.Biology + Moon.Biology"
```

---

## Task 2: `RollBiomass` (WBH p.127-128)

**Files:**

- Modify: `worlds/biology.go`
- Create: `worlds/biology_test.go`

- [ ] **Step 1: Write failing tests**

Create `worlds/biology_test.go`:

```go
package worlds

import (
	"testing"

	"wbh/roller"
)

func TestRollBiomass_ZedPrime(t *testing.T) {
	// Atm 6 (no DM), Hydro 6 (+1), Age 6.3 (+1), MeanK 300 (+2), HighK 346 (no DM).
	// DMs total +4 (at cap), 2D=6 → 6+4 = 10 → biomass A.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 6}
	body.Temperature = &Temperature{MeanK: 300, HighK: 346}
	r := roller.NewScripted(6)
	got := RollBiomass(r, body, 6.3)
	if got != 10 {
		t.Errorf("Zed Prime: got %d, want 10", got)
	}
}

func TestRollBiomass_DMCap_AtPositiveCeiling(t *testing.T) {
	// Atm 8 (+2) + Hydro A (+2) + Age > 4 (+1) + MeanK 290 (+2) = +7, clamp +4.
	// 2D=10 → 10+4 = 14 → biomass E.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 8}
	body.Hydrographics = &Hydrographics{Code: 10}
	body.Temperature = &Temperature{MeanK: 290}
	r := roller.NewScripted(10)
	got := RollBiomass(r, body, 5.0)
	if got != 14 {
		t.Errorf("got %d, want 14 (DM cap +4)", got)
	}
}

func TestRollBiomass_DMCap_AtNegativeFloor(t *testing.T) {
	// Vacuum atm 0 (-6) + Hydro 0 (-4) + Age < 0.2 (-6) + MeanK 100 (-2) + HighK 100 (-4)
	// = -22, clamp to -12.
	// 2D=2 → 2-12 = -10 → biomass clamped to 0.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 0}
	body.Hydrographics = &Hydrographics{Code: 0}
	body.Temperature = &Temperature{MeanK: 100, HighK: 100}
	r := roller.NewScripted(2)
	got := RollBiomass(r, body, 0.1)
	if got != 0 {
		t.Errorf("got %d, want 0 (DM clamp -12; result < 0 → 0)", got)
	}
}

func TestRollBiomass_ExoticAtm_BonusApplied_AtmB(t *testing.T) {
	// Atm B (-5) + Hydro 6 (+1) + Age 5 (+1) + MeanK 290 (+2) = -1, no clamp needed.
	// 2D=8 → 8 - 1 = 7 (biomass ≥ 1).
	// Exotic-atm bonus for atm B: |−5| − 1 = +4 → final 11.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 11}
	body.Hydrographics = &Hydrographics{Code: 6}
	body.Temperature = &Temperature{MeanK: 290}
	r := roller.NewScripted(8)
	got := RollBiomass(r, body, 5.0)
	if got != 11 {
		t.Errorf("got %d, want 11 (atm B bonus +4 applied)", got)
	}
}

func TestRollBiomass_ExoticAtm_BonusSkipped_AtmBZero(t *testing.T) {
	// Atm B (-5), no other DMs, 2D=2 → 2-5 = -3 → biomass 0.
	// Bonus NOT applied because rolled biomass is 0.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 11}
	body.Hydrographics = &Hydrographics{Code: 0}
	r := roller.NewScripted(2)
	got := RollBiomass(r, body, 2.0)
	if got != 0 {
		t.Errorf("got %d, want 0 (bonus skipped when biomass=0)", got)
	}
}

func TestRollBiomass_VacuumAtm_BonusApplied(t *testing.T) {
	// Atm 0 (-6) + Hydro 9 (+2) = -4, no clamp.
	// 2D=12 → 12 - 4 = 8 (biomass ≥ 1).
	// Bonus for atm 0: |−6| − 1 = +5 → final 13.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 0}
	body.Hydrographics = &Hydrographics{Code: 9}
	r := roller.NewScripted(12)
	got := RollBiomass(r, body, 5.0)
	if got != 13 {
		t.Errorf("got %d, want 13 (atm 0 bonus +5 applied)", got)
	}
}

func TestRollBiomass_NilAtmosphere_Zero(t *testing.T) {
	body := &DetailedPlacement{}
	body.Hydrographics = &Hydrographics{Code: 5}
	r := roller.NewScripted(7)
	got := RollBiomass(r, body, 5.0)
	if got != 0 {
		t.Errorf("got %d, want 0 (nil atmosphere)", got)
	}
}

func TestRollBiomass_NilHydrographics_HydroZeroDM(t *testing.T) {
	// Atm 6 (no DM), nil hydro treated as DM-4, MeanK 290 (+2), Age 5 (+1) = -1.
	// 2D=10 → 10 - 1 = 9.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Temperature = &Temperature{MeanK: 290}
	r := roller.NewScripted(10)
	got := RollBiomass(r, body, 5.0)
	if got != 9 {
		t.Errorf("got %d, want 9 (nil hydro → DM-4)", got)
	}
}

func TestRollBiomass_NilTemperature_NoTempDMs(t *testing.T) {
	// Atm 6 + Hydro 5 (no DM) + Age 5 (+1), no temp DMs (nil temp).
	// 2D=8 → 8 + 1 = 9.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	r := roller.NewScripted(8)
	got := RollBiomass(r, body, 5.0)
	if got != 9 {
		t.Errorf("got %d, want 9 (nil temp, no temp DMs)", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test -run TestRollBiomass ./worlds/...
```

Expected: FAIL with "undefined: RollBiomass".

- [ ] **Step 3: Implement `RollBiomass`**

Append to `worlds/biology.go`:

```go
// RollBiomass per WBH p.127-128: 2D + DMs (combined DM sum clamped to
// [-12, +4]). Includes the exotic-atm bonus (Special Case 2): if the
// rolled biomass is ≥ 1 AND atm code ∈ {0, 1, 10, 11, 12, 15}, add
// (|atm DM| − 1) to the result.
//
// Returns 0 if body or body.Atmosphere is nil. nil Hydrographics is
// treated as Hydro 0 (DM-4). nil Temperature contributes no temp DMs.
//
// Skipped: Special Case 1 (biologic-taint biomass=0 promotion) requires
// Atmosphere taint typology not yet modeled — deferred per spec Q3-a.
func RollBiomass(r roller.Roller, body *DetailedPlacement, ageGyr float64) int {
	if body == nil || body.Atmosphere == nil {
		return 0
	}
	atmDM := biomassAtmDM(body.Atmosphere.Code)
	hydroCode := 0
	if body.Hydrographics != nil {
		hydroCode = body.Hydrographics.Code
	}
	dm := atmDM + biomassHydroDM(hydroCode) + biomassAgeDM(ageGyr)
	if body.Temperature != nil {
		dm += biomassTempDM(body.Temperature.MeanK, body.Temperature.HighK)
	}
	dm = min(max(dm, -12), 4)

	roll := r.Roll("2D")
	biomass := roll + dm
	if biomass < 0 {
		biomass = 0
	}

	// Exotic-atm bonus (rolled biomass ≥ 1 on atm 0/1/A/B/C/F+).
	if biomass >= 1 && exoticBiomassBonusApplies(body.Atmosphere.Code) {
		biomass += exoticBiomassBonus(body.Atmosphere.Code)
	}
	return biomass
}

// biomassAtmDM per WBH p.128 atmosphere-DM table.
func biomassAtmDM(atmCode int) int {
	switch atmCode {
	case 0:
		return -6
	case 1:
		return -4
	case 2, 3, 14: // E
		return -3
	case 4, 5:
		return -2
	case 8, 9, 13: // D
		return +2
	case 10: // A
		return -3
	case 11: // B
		return -5
	case 12: // C
		return -7
	case 15: // F+
		return -5
	}
	return 0 // atm 6, 7, or unmapped
}

// biomassHydroDM per WBH p.128 hydrographics-DM table.
func biomassHydroDM(hydroCode int) int {
	switch {
	case hydroCode == 0:
		return -4
	case hydroCode >= 1 && hydroCode <= 3:
		return -2
	case hydroCode >= 6 && hydroCode <= 8:
		return +1
	case hydroCode >= 9:
		return +2
	}
	return 0 // 4-5
}

// biomassAgeDM per WBH p.128 age-DM table.
func biomassAgeDM(ageGyr float64) int {
	switch {
	case ageGyr < 0.2:
		return -6
	case ageGyr < 1:
		return -2
	case ageGyr > 4:
		return +1
	}
	return 0
}

// biomassTempDM per WBH p.128 temperature-DM table. Returns 0 when both
// MeanK and HighK are 0 (defensive — caller already gated on nil temp).
func biomassTempDM(meanK, highK float64) int {
	dm := 0
	if highK > 353 {
		dm += -2
	} else if highK > 0 && highK < 273 {
		dm += -4
	}
	if meanK > 353 {
		dm += -4
	} else if meanK > 0 && meanK < 273 {
		dm += -2
	}
	if meanK >= 279 && meanK <= 303 {
		dm += +2
	}
	return dm
}

// exoticBiomassBonusApplies reports whether atm code is in {0, 1, A, B, C, F+}.
func exoticBiomassBonusApplies(atmCode int) bool {
	return atmCode == 0 || atmCode == 1 ||
		atmCode == 10 || atmCode == 11 || atmCode == 12 || atmCode == 15
}

// exoticBiomassBonus returns |atmDM| − 1 for the exotic atm codes per WBH
// Special Case 2: "Add one less than the negative Atmosphere DM".
func exoticBiomassBonus(atmCode int) int {
	switch atmCode {
	case 0:
		return 5
	case 1:
		return 3
	case 10:
		return 2
	case 11:
		return 4
	case 12:
		return 6
	case 15:
		return 4
	}
	return 0
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestRollBiomass ./worlds/... -v
```

Expected: all 9 tests PASS.

- [ ] **Step 5: just check && just test**

```bash
just check
just test
```

Expected: 0 issues; all packages pass.

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/biology.go \
        worlds/biology_test.go
git commit -m "feat(worlds): RollBiomass (WBH p.127-128)"
```

---

## Task 3: `RollBiocomplexity` (WBH p.129)

**Files:**

- Modify: `worlds/biology.go`
- Modify: `worlds/biology_test.go`

- [ ] **Step 1: Write failing tests**

Append to `worlds/biology_test.go`:

```go
func TestRollBiocomplexity_ZedPrime(t *testing.T) {
	// Biomass=10 (clamped to 9), Atm 6 (in 4-9 → no DM), Age 6.3 (no age DM).
	// 2D=3 → 3 - 7 + 9 = 5.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(3)
	got := RollBiocomplexity(r, body, 10, 6.3)
	if got != 5 {
		t.Errorf("Zed Prime: got %d, want 5", got)
	}
}

func TestRollBiocomplexity_BiomassZero_Zero(t *testing.T) {
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted() // empty — must NOT consume dice
	got := RollBiocomplexity(r, body, 0, 6.3)
	if got != 0 {
		t.Errorf("got %d, want 0 (Biomass=0 prerequisite fails)", got)
	}
}

func TestRollBiocomplexity_BiomassClamp_Above9(t *testing.T) {
	// Biomass=15 should be clamped to 9 in the formula.
	// 2D=2, Atm 6, Age > 4 → 2 - 7 + 9 + 0 = 4.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(2)
	got := RollBiocomplexity(r, body, 15, 5.0)
	if got != 4 {
		t.Errorf("got %d, want 4 (Biomass=15 → uses 9)", got)
	}
}

func TestRollBiocomplexity_AgeBoundary_Exactly4_UsesWorseDM(t *testing.T) {
	// Age = 4.0 exactly → 3-4 band → DM-2 (the worst at the boundary).
	// Biomass=9, Atm 6, 2D=10 → 10 - 7 + 9 - 2 = 10.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(10)
	got := RollBiocomplexity(r, body, 9, 4.0)
	if got != 10 {
		t.Errorf("age=4 boundary: got %d, want 10 (DM-2 worst)", got)
	}
}

func TestRollBiocomplexity_AgeBoundary_Exactly1_UsesWorseDM(t *testing.T) {
	// Age = 1.0 exactly → < 1 band → DM-10 (the worst at the boundary).
	// Biomass=9, Atm 6, 2D=12 → 12 - 7 + 9 - 10 = 4.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(12)
	got := RollBiocomplexity(r, body, 9, 1.0)
	if got != 4 {
		t.Errorf("age=1 boundary: got %d, want 4 (DM-10 worst)", got)
	}
}

func TestRollBiocomplexity_AtmNotIn4to9_DMMinus2(t *testing.T) {
	// Atm 11 (B) → not in 4-9 → DM-2. Biomass=9, Age 5.
	// 2D=10 → 10 - 7 + 9 - 2 = 10.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 11}
	r := roller.NewScripted(10)
	got := RollBiocomplexity(r, body, 9, 5.0)
	if got != 10 {
		t.Errorf("got %d, want 10 (atm not 4-9 → DM-2)", got)
	}
}

func TestRollBiocomplexity_ResultLessThanOne_PromotedToOne(t *testing.T) {
	// Force a result < 1 with biomass > 0: 2D=2, Biomass=1, Atm 11 (DM-2), Age 0.5 (DM-10 from biocomplexity table).
	// 2 - 7 + 1 - 2 - 10 = -16 → < 1 → 1.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 11}
	r := roller.NewScripted(2)
	got := RollBiocomplexity(r, body, 1, 0.5)
	if got != 1 {
		t.Errorf("got %d, want 1 (result < 1 → promoted)", got)
	}
}

func TestRollBiocomplexity_NilAtmosphere_NoAtmDM(t *testing.T) {
	// Defensive: nil atmosphere is not in 4-9, so still gets DM-2.
	body := &DetailedPlacement{}
	body.Atmosphere = nil
	r := roller.NewScripted(10)
	got := RollBiocomplexity(r, body, 9, 5.0)
	// 10 - 7 + 9 - 2 = 10
	if got != 10 {
		t.Errorf("got %d, want 10 (nil atm → DM-2)", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run TestRollBiocomplexity ./worlds/...
```

Expected: FAIL with "undefined: RollBiocomplexity".

- [ ] **Step 3: Implement `RollBiocomplexity`**

Append to `worlds/biology.go`:

```go
// RollBiocomplexity per WBH p.129: 2D - 7 + min(Biomass, 9) + DMs.
// Returns 0 without consuming dice if biomass <= 0.
//
// DMs:
//   - Atmosphere not 4-9: -2
//   - Age 3-4 Gyrs: -2
//   - Age 2-3 Gyrs: -4
//   - Age 1-2 Gyrs: -8
//   - Age < 1 Gyr: -10
//
// At age boundaries (e.g., age=2.0), uses the worst (more negative) DM
// per WBH "If the system age is exactly at a limit between two DMs, use
// the worst DM." Implemented via ordered-case switch with inclusive
// upper bounds.
//
// Result < 1 promoted to 1 (when biomass > 0).
//
// Skipped: low-oxygen-taint DM-2 deferred per spec Q3-a (taint typology
// not yet modeled).
func RollBiocomplexity(r roller.Roller, body *DetailedPlacement, biomass int, ageGyr float64) int {
	if biomass <= 0 {
		return 0
	}
	dm := biocomplexityAgeDM(ageGyr)
	if !atmIs4to9(body) {
		dm += -2
	}
	roll := r.Roll("2D")
	bx := min(biomass, 9)
	result := roll - 7 + bx + dm
	if result < 1 {
		return 1
	}
	return result
}

// biocomplexityAgeDM per WBH p.129. Uses the worst DM at boundaries.
func biocomplexityAgeDM(ageGyr float64) int {
	switch {
	case ageGyr <= 1:
		return -10
	case ageGyr <= 2:
		return -8
	case ageGyr <= 3:
		return -4
	case ageGyr <= 4:
		return -2
	}
	return 0
}

// atmIs4to9 reports whether body.Atmosphere.Code is in [4, 9]. Returns
// false on nil atmosphere.
func atmIs4to9(body *DetailedPlacement) bool {
	if body == nil || body.Atmosphere == nil {
		return false
	}
	c := body.Atmosphere.Code
	return c >= 4 && c <= 9
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestRollBiocomplexity ./worlds/... -v
```

Expected: all 8 tests PASS.

- [ ] **Step 5: just check && just test**

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/biology.go \
        worlds/biology_test.go
git commit -m "feat(worlds): RollBiocomplexity (WBH p.129)"
```

---

## Task 4: `RollNativeSophont` (extant, WBH p.130)

**Files:**

- Modify: `worlds/biology.go`
- Modify: `worlds/biology_test.go`

- [ ] **Step 1: Write failing tests**

Append to `worlds/biology_test.go`:

```go
func TestRollNativeSophont_BelowPrerequisite_False(t *testing.T) {
	// Biocomplexity=7 < 8 → no roll, no dice consumed.
	r := roller.NewScripted() // empty
	if got := RollNativeSophont(r, 7); got {
		t.Error("got true, want false (Biocomplexity<8)")
	}
}

func TestRollNativeSophont_Triggers_AtBiocomplexity9(t *testing.T) {
	// Biocomplexity=9, 2D=11 → 11+9-7=13 ≥ 13 → true.
	r := roller.NewScripted(11)
	if got := RollNativeSophont(r, 9); !got {
		t.Error("got false, want true (mod=13)")
	}
}

func TestRollNativeSophont_BelowThreshold(t *testing.T) {
	// Biocomplexity=8, 2D=11 → 11+8-7=12 < 13 → false.
	r := roller.NewScripted(11)
	if got := RollNativeSophont(r, 8); got {
		t.Error("got true, want false (mod=12)")
	}
}

func TestRollNativeSophont_BiocomplexityClamp_Above9(t *testing.T) {
	// Biocomplexity=15 should be clamped to 9 in the formula.
	// 2D=11 → 11+9-7=13 ≥ 13 → true.
	r := roller.NewScripted(11)
	if got := RollNativeSophont(r, 15); !got {
		t.Error("got false, want true (Biocomplexity=15 → uses 9)")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run TestRollNativeSophont ./worlds/...
```

Expected: FAIL with "undefined: RollNativeSophont".

- [ ] **Step 3: Implement `RollNativeSophont`**

Append to `worlds/biology.go`:

```go
// RollNativeSophont per WBH p.130: extant native sophont exists if
// 2D + min(Biocomplexity, 9) - 7 ≥ 13.
//
// Returns false without consuming dice if biocomplexity < 8.
func RollNativeSophont(r roller.Roller, biocomplexity int) bool {
	if biocomplexity < 8 {
		return false
	}
	bx := min(biocomplexity, 9)
	roll := r.Roll("2D")
	return roll+bx-7 >= 13
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestRollNativeSophont ./worlds/... -v
```

Expected: all 4 tests PASS.

- [ ] **Step 5: just check && just test**

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/biology.go \
        worlds/biology_test.go
git commit -m "feat(worlds): RollNativeSophont (WBH p.130)"
```

---

## Task 5: `RollExtinctSophont` (WBH p.130)

**Files:**

- Modify: `worlds/biology.go`
- Modify: `worlds/biology_test.go`

- [ ] **Step 1: Write failing tests**

Append to `worlds/biology_test.go`:

```go
func TestRollExtinctSophont_BelowPrerequisite_False(t *testing.T) {
	// Biocomplexity=7 → no roll, no dice consumed.
	r := roller.NewScripted()
	if got := RollExtinctSophont(r, 7, 6.0); got {
		t.Error("got true, want false (Biocomplexity<8)")
	}
}

func TestRollExtinctSophont_AgeOver5_DMPlusOne(t *testing.T) {
	// Biocomplexity=9, Age=6 (DM+1), 2D=10 → 10+9-7+1=13 ≥ 13 → true.
	// Without the +1: 10+9-7=12 < 13 → false. So this test verifies the DM applies.
	r := roller.NewScripted(10)
	if got := RollExtinctSophont(r, 9, 6.0); !got {
		t.Error("got false, want true (age>5 DM+1 makes mod=13)")
	}
}

func TestRollExtinctSophont_AgeUnder5_NoDM(t *testing.T) {
	// Biocomplexity=9, Age=4 (no DM), 2D=10 → 10+9-7=12 < 13 → false.
	r := roller.NewScripted(10)
	if got := RollExtinctSophont(r, 9, 4.0); got {
		t.Error("got true, want false (age≤5 no DM, mod=12)")
	}
}

func TestRollExtinctSophont_HighRoll_AlwaysTrue(t *testing.T) {
	// Biocomplexity=8, Age=4, 2D=12 → 12+8-7=13 ≥ 13 → true.
	r := roller.NewScripted(12)
	if got := RollExtinctSophont(r, 8, 4.0); !got {
		t.Error("got false, want true (mod=13 even without age DM)")
	}
}

func TestRollExtinctSophont_BiocomplexityClamp_Above9(t *testing.T) {
	// Biocomplexity=15 → clamped to 9. 2D=10, Age=4 → 10+9-7=12 < 13 → false.
	r := roller.NewScripted(10)
	if got := RollExtinctSophont(r, 15, 4.0); got {
		t.Error("got true, want false (Biocomplexity=15 → 9, mod=12)")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run TestRollExtinctSophont ./worlds/...
```

Expected: FAIL with "undefined: RollExtinctSophont".

- [ ] **Step 3: Implement `RollExtinctSophont`**

Append to `worlds/biology.go`:

```go
// RollExtinctSophont per WBH p.130: evidence of an extinct native sophont
// existed if 2D + min(Biocomplexity, 9) - 7 + DMs ≥ 13.
//
// DMs: +1 if system age > 5 Gyrs.
//
// Returns false without consuming dice if biocomplexity < 8. Independent
// of RollNativeSophont — both can be true (extant species with extinct
// ancestors).
func RollExtinctSophont(r roller.Roller, biocomplexity int, ageGyr float64) bool {
	if biocomplexity < 8 {
		return false
	}
	dm := 0
	if ageGyr > 5 {
		dm = 1
	}
	bx := min(biocomplexity, 9)
	roll := r.Roll("2D")
	return roll+bx-7+dm >= 13
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestRollExtinctSophont ./worlds/... -v
```

Expected: all 5 tests PASS.

- [ ] **Step 5: just check && just test**

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/biology.go \
        worlds/biology_test.go
git commit -m "feat(worlds): RollExtinctSophont (WBH p.130)"
```

---

## Task 6: `RollBiodiversity` (WBH p.130)

**Files:**

- Modify: `worlds/biology.go`
- Modify: `worlds/biology_test.go`

- [ ] **Step 1: Write failing tests**

Append to `worlds/biology_test.go`:

```go
func TestRollBiodiversity_ZedPrime(t *testing.T) {
	// Biomass=10, Biocomplexity=5 → (10+5)/2 = 7.5. 2D=6 → 6-7+7.5 = 6.5 → ceil → 7.
	r := roller.NewScripted(6)
	got := RollBiodiversity(r, 10, 5)
	if got != 7 {
		t.Errorf("Zed Prime: got %d, want 7", got)
	}
}

func TestRollBiodiversity_BiomassZero_Zero(t *testing.T) {
	r := roller.NewScripted()
	if got := RollBiodiversity(r, 0, 5); got != 0 {
		t.Errorf("got %d, want 0 (biomass=0 prerequisite fails)", got)
	}
}

func TestRollBiodiversity_RoundsUp(t *testing.T) {
	// Biomass=4, Biocomplexity=3 → (4+3)/2 = 3.5. 2D=8 → 8-7+3.5 = 4.5 → ceil → 5.
	r := roller.NewScripted(8)
	got := RollBiodiversity(r, 4, 3)
	if got != 5 {
		t.Errorf("got %d, want 5 (ceil semantics)", got)
	}
}

func TestRollBiodiversity_ResultLessThanOne_PromotedToOne(t *testing.T) {
	// Biomass=1, Biocomplexity=1 → (1+1)/2 = 1. 2D=2 → 2-7+1 = -4 → < 1 → 1.
	r := roller.NewScripted(2)
	got := RollBiodiversity(r, 1, 1)
	if got != 1 {
		t.Errorf("got %d, want 1 (result<1 promoted)", got)
	}
}

func TestRollBiodiversity_IntegerArithmetic_NoFractional(t *testing.T) {
	// Biomass=4, Biocomplexity=4 → (4+4)/2 = 4 (integer). 2D=7 → 7-7+4 = 4 (no rounding).
	r := roller.NewScripted(7)
	got := RollBiodiversity(r, 4, 4)
	if got != 4 {
		t.Errorf("got %d, want 4 (no rounding when integer)", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run TestRollBiodiversity ./worlds/...
```

Expected: FAIL with "undefined: RollBiodiversity".

- [ ] **Step 3: Implement `RollBiodiversity`**

Append to `worlds/biology.go`:

```go
import "math"

// RollBiodiversity per WBH p.130: ceil(2D - 7 + (Biomass + Biocomplexity) / 2).
//
// Returns 0 without consuming dice if biomass <= 0.
// Result < 1 promoted to 1 (when biomass > 0).
//
// The (B + X) / 2 is float division; the final ceil rounds the entire
// expression. So Biomass=10 + Biocomplexity=5 with 2D=6 produces
// 6 - 7 + 7.5 = 6.5 → ceil → 7.
func RollBiodiversity(r roller.Roller, biomass, biocomplexity int) int {
	if biomass <= 0 {
		return 0
	}
	roll := r.Roll("2D")
	v := float64(roll) - 7 + float64(biomass+biocomplexity)/2.0
	result := int(math.Ceil(v))
	if result < 1 {
		return 1
	}
	return result
}
```

(Note: the `import "math"` at the top of `biology.go` should be added if not already present.)

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestRollBiodiversity ./worlds/... -v
```

Expected: all 5 tests PASS.

- [ ] **Step 5: just check && just test**

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/biology.go \
        worlds/biology_test.go
git commit -m "feat(worlds): RollBiodiversity (WBH p.130)"
```

---

## Task 7: `RollCompatibility` (WBH p.130-131)

**Files:**

- Modify: `worlds/biology.go`
- Modify: `worlds/biology_test.go`

- [ ] **Step 1: Write failing tests**

Append to `worlds/biology_test.go`:

```go
func TestRollCompatibility_ZedPrime_FollowsFormula(t *testing.T) {
	// Per WBH p.131 formula box: 2D - Biocomplexity/2 + DMs.
	// Atm 6 (DM+2), Biocomplexity=5, Age 6.3, 2D=7 → 7 - 2.5 + 2 = 6.5 → floor → 6.
	//
	// NOTE: WBH p.131 worked example shows 7 + 3 - 2.5 + 2 = 9.5 → 9. The
	// "+3" has no source in the formula box. Implementation follows formula
	// → Zed Prime gets 6 (book says 9). Logged as feedback memory after merge.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(7)
	got := RollCompatibility(r, body, 5, 6.3)
	if got != 6 {
		t.Errorf("Zed Prime per formula: got %d, want 6 (book worked example says 9)", got)
	}
}

func TestRollCompatibility_BiomassDependsOnPrereq_NoDirectGate(t *testing.T) {
	// The Compatibility function itself doesn't gate on biomass — caller
	// should check biomass > 0 before calling. This test verifies the
	// function still returns the formula result for any inputs.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(7)
	got := RollCompatibility(r, body, 5, 5.0)
	if got != 6 {
		t.Errorf("got %d, want 6", got)
	}
}

func TestRollCompatibility_NegativeResult_ClampedToZero(t *testing.T) {
	// Atm C (DM-10), Biocomplexity=10, 2D=2 → 2 - 5 - 10 = -13 → ≤ 0 → 0.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 12}
	r := roller.NewScripted(2)
	got := RollCompatibility(r, body, 10, 5.0)
	if got != 0 {
		t.Errorf("got %d, want 0 (negative result clamped)", got)
	}
}

func TestRollCompatibility_AtmCRich_DMMinus10(t *testing.T) {
	// Atm C (DM-10), Biocomplexity=4, 2D=12 → 12 - 2 - 10 = 0 → 0.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 12}
	r := roller.NewScripted(12)
	got := RollCompatibility(r, body, 4, 5.0)
	if got != 0 {
		t.Errorf("got %d, want 0 (atm C heavy penalty)", got)
	}
}

func TestRollCompatibility_AgeOver8_DMMinus2(t *testing.T) {
	// Atm 6 (+2), Biocomplexity=4, Age=9 (DM-2), 2D=10 → 10 - 2 + 2 - 2 = 8.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(10)
	got := RollCompatibility(r, body, 4, 9.0)
	if got != 8 {
		t.Errorf("got %d, want 8 (age>8 DM-2)", got)
	}
}

func TestRollCompatibility_FloorRounding(t *testing.T) {
	// Biocomplexity=3 → 3/2 = 1.5. 2D=10, Atm 6 (+2) → 10 - 1.5 + 2 = 10.5 → floor → 10.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	r := roller.NewScripted(10)
	got := RollCompatibility(r, body, 3, 5.0)
	if got != 10 {
		t.Errorf("got %d, want 10 (floor 10.5)", got)
	}
}

func TestRollCompatibility_NilAtmosphere_NoAtmDM(t *testing.T) {
	// Defensive: nil atm → no atm DM applied. Biocomplexity=4, Age 5, 2D=10 → 10 - 2 = 8.
	body := &DetailedPlacement{}
	r := roller.NewScripted(10)
	got := RollCompatibility(r, body, 4, 5.0)
	if got != 8 {
		t.Errorf("got %d, want 8 (nil atm)", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run TestRollCompatibility ./worlds/...
```

Expected: FAIL with "undefined: RollCompatibility".

- [ ] **Step 3: Implement `RollCompatibility`**

Append to `worlds/biology.go`:

```go
// RollCompatibility per WBH p.130-131:
//
//	floor(2D - Biocomplexity/2 + DMs)
//
// Result clamped to ≥ 0. Caller is responsible for the biomass > 0
// prerequisite — this function returns the formula result for any inputs.
//
// DMs:
//   - Atmosphere 0, 1, B: -8
//   - Atmosphere 2, 4, 7, 9: -2
//   - Atmosphere 3, 5, 8: +1
//   - Atmosphere 6: +2
//   - Atmosphere A, F: -6
//   - Atmosphere C: -10
//   - Atmosphere D, E: -1
//   - Age > 8 Gyrs: -2
//
// NOTE: WBH p.131 worked example shows 7 + 3 - 2.5 + 2 = 9.5 for Zed Prime
// (giving compatibility 9), but the formula box has no "+3" addend.
// Implementation follows the formula → Zed Prime compatibility = 6.
// This divergence is documented as a feedback memory after merge.
//
// Atm codes G/H mentioned in the book DM table don't exist in the WBH
// 0-F atm system. They cannot be produced by RollAtmoCode — no DM applied.
//
// Skipped: "or otherwise tainted" qualifier on the -2 row deferred per
// spec Q3-a (Atmosphere taint typology not yet modeled).
func RollCompatibility(r roller.Roller, body *DetailedPlacement, biocomplexity int, ageGyr float64) int {
	dm := compatibilityAtmDM(body)
	if ageGyr > 8 {
		dm += -2
	}
	roll := r.Roll("2D")
	v := float64(roll) - float64(biocomplexity)/2.0 + float64(dm)
	result := int(math.Floor(v))
	if result < 0 {
		return 0
	}
	return result
}

// compatibilityAtmDM per WBH p.131 atmosphere-DM table.
func compatibilityAtmDM(body *DetailedPlacement) int {
	if body == nil || body.Atmosphere == nil {
		return 0
	}
	switch body.Atmosphere.Code {
	case 0, 1, 11: // 0, 1, B (G and H don't exist in our system)
		return -8
	case 2, 4, 7, 9:
		return -2
	case 3, 5, 8:
		return +1
	case 6:
		return +2
	case 10, 15: // A, F
		return -6
	case 12: // C
		return -10
	case 13, 14: // D, E
		return -1
	}
	return 0
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestRollCompatibility ./worlds/... -v
```

Expected: all 7 tests PASS.

- [ ] **Step 5: just check && just test**

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/biology.go \
        worlds/biology_test.go
git commit -m "feat(worlds): RollCompatibility (WBH p.130-131)"
```

---

## Task 8: `RollResourceRating` (WBH p.131)

**Files:**

- Modify: `worlds/biology.go`
- Modify: `worlds/biology_test.go`

- [ ] **Step 1: Write failing tests**

Append to `worlds/biology_test.go`:

```go
func TestRollResourceRating_TerrestrialNoLife(t *testing.T) {
	// Size 5, Density 1.0 (no DM), no biology.
	// 2D=8 → 8 - 7 + 5 = 6.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.0}
	r := roller.NewScripted(8)
	got := RollResourceRating(r, body, &Biology{})
	if got != 6 {
		t.Errorf("got %d, want 6", got)
	}
}

func TestRollResourceRating_HighDensity_PlusTwo(t *testing.T) {
	// Size 5, Density 1.5 (>1.12 → +2). 2D=8 → 8 - 7 + 5 + 2 = 8.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.5}
	r := roller.NewScripted(8)
	got := RollResourceRating(r, body, &Biology{})
	if got != 8 {
		t.Errorf("got %d, want 8 (high density +2)", got)
	}
}

func TestRollResourceRating_LowDensity_MinusTwo(t *testing.T) {
	// Size 5, Density 0.4 (<0.5 → -2). 2D=8 → 8 - 7 + 5 - 2 = 4.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 0.4}
	r := roller.NewScripted(8)
	got := RollResourceRating(r, body, &Biology{})
	if got != 4 {
		t.Errorf("got %d, want 4 (low density -2)", got)
	}
}

func TestRollResourceRating_HighBiomass_PlusTwo(t *testing.T) {
	// Size 5, Density 1.0, Biomass=5 (≥3 → +2). 2D=8 → 8-7+5+2 = 8.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.0}
	bio := &Biology{Biomass: 5}
	r := roller.NewScripted(8)
	got := RollResourceRating(r, body, bio)
	if got != 8 {
		t.Errorf("got %d, want 8 (biomass≥3 +2)", got)
	}
}

func TestRollResourceRating_HighBiodiversity_PlusOne_8toA(t *testing.T) {
	// Size 5, Density 1.0, Biomass=1 (no biomass DM since <3),
	// Biodiversity=8 (8-A → +1). 2D=8 → 8-7+5+1 = 7.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.0}
	bio := &Biology{Biomass: 1, Biodiversity: 8}
	r := roller.NewScripted(8)
	got := RollResourceRating(r, body, bio)
	if got != 7 {
		t.Errorf("got %d, want 7 (biodiversity 8-A +1)", got)
	}
}

func TestRollResourceRating_HighBiodiversity_PlusTwo_BPlus(t *testing.T) {
	// Biodiversity=11 (B+ → +2). Size 5, Density 1.0. 2D=8 → 8-7+5+2 = 8.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.0}
	bio := &Biology{Biomass: 1, Biodiversity: 11}
	r := roller.NewScripted(8)
	got := RollResourceRating(r, body, bio)
	if got != 8 {
		t.Errorf("got %d, want 8 (biodiversity B+ +2)", got)
	}
}

func TestRollResourceRating_LowCompatibilityWithLife_MinusOne(t *testing.T) {
	// Compatibility 0-3 + Biomass ≥ 1 → -1. Size 5, Density 1.0.
	// 2D=8, Biomass=1, Compatibility=2 → 8-7+5-1 = 5.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.0}
	bio := &Biology{Biomass: 1, Compatibility: 2}
	r := roller.NewScripted(8)
	got := RollResourceRating(r, body, bio)
	if got != 5 {
		t.Errorf("got %d, want 5 (compatibility 0-3 with biomass≥1: -1)", got)
	}
}

func TestRollResourceRating_LowCompatibilityNoLife_NoDMSkipped(t *testing.T) {
	// Biomass=0 → the compatibility-0-3 -1 DM does NOT apply.
	// Size 5, Density 1.0, 2D=8 → 8-7+5 = 6.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.0}
	bio := &Biology{Biomass: 0, Compatibility: 2}
	r := roller.NewScripted(8)
	got := RollResourceRating(r, body, bio)
	if got != 6 {
		t.Errorf("got %d, want 6 (no biomass: -1 DM skipped)", got)
	}
}

func TestRollResourceRating_HighCompatibility_PlusTwo(t *testing.T) {
	// Compatibility 8+ → +2. Size 5, Density 1.0, 2D=8, Biomass=1 → 8-7+5+2 = 8.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.0}
	bio := &Biology{Biomass: 1, Compatibility: 8}
	r := roller.NewScripted(8)
	got := RollResourceRating(r, body, bio)
	if got != 8 {
		t.Errorf("got %d, want 8 (compatibility 8+ +2)", got)
	}
}

func TestRollResourceRating_ResultBelowTwo_ClampedToTwo(t *testing.T) {
	// Force result < 2: Size 1, Density 0.4 (-2), 2D=2 → 2-7+1-2 = -6 → < 2 → 2.
	body := &DetailedPlacement{}
	body.SizeCode = "1"
	body.Physical = &BodyPhysical{Density: 0.4}
	r := roller.NewScripted(2)
	got := RollResourceRating(r, body, &Biology{})
	if got != 2 {
		t.Errorf("got %d, want 2 (clamp to ≥2)", got)
	}
}

func TestRollResourceRating_ResultAboveTwelve_ClampedToTwelve(t *testing.T) {
	// Force result > 12: Size 15 (F), Density 1.5 (+2), Biomass=5 (+2),
	// Biodiversity=11 (+2), Compatibility=10 (+2), 2D=12 → 12-7+15+2+2+2+2 = 28 → > 12 → 12.
	body := &DetailedPlacement{}
	body.SizeCode = "F"
	body.Physical = &BodyPhysical{Density: 1.5}
	bio := &Biology{Biomass: 5, Biodiversity: 11, Compatibility: 10}
	r := roller.NewScripted(12)
	got := RollResourceRating(r, body, bio)
	if got != 12 {
		t.Errorf("got %d, want 12 (clamp to ≤12)", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run TestRollResourceRating ./worlds/...
```

Expected: FAIL with "undefined: RollResourceRating".

- [ ] **Step 3: Implement `RollResourceRating`**

Append to `worlds/biology.go`:

```go
// RollResourceRating per WBH p.131:
//
//	2D - 7 + Size + DMs, clamped to [2, 12]
//
// Runs for ALL terrestrial bodies regardless of biology — biology DMs
// only apply when ratings are non-zero.
//
// DMs:
//   - Density > 1.12: +2
//   - Density < 0.5: -2
//   - Biomass ≥ 3: +2
//   - Biodiversity 8-10 (8-A): +1
//   - Biodiversity ≥ 11 (B+): +2
//   - Compatibility 0-3: -1 (only if Biomass ≥ 1)
//   - Compatibility ≥ 8: +2
//
// bio may be a zero Biology{} for bodies without life.
func RollResourceRating(r roller.Roller, body *DetailedPlacement, bio *Biology) int {
	dm := 0
	if body != nil && body.Physical != nil {
		switch {
		case body.Physical.Density > 1.12:
			dm += 2
		case body.Physical.Density < 0.5:
			dm += -2
		}
	}
	if bio != nil {
		if bio.Biomass >= 3 {
			dm += 2
		}
		switch {
		case bio.Biodiversity >= 11:
			dm += 2
		case bio.Biodiversity >= 8:
			dm += 1
		}
		if bio.Compatibility >= 0 && bio.Compatibility <= 3 && bio.Biomass >= 1 {
			dm += -1
		}
		if bio.Compatibility >= 8 {
			dm += 2
		}
	}
	roll := r.Roll("2D")
	size := 0
	if body != nil {
		size = SizeAsInt(body.SizeCode)
	}
	result := roll - 7 + size + dm
	return min(max(result, 2), 12)
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test -run TestRollResourceRating ./worlds/... -v
```

Expected: all 11 tests PASS.

- [ ] **Step 5: just check && just test**

- [ ] **Step 6: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/biology.go \
        worlds/biology_test.go
git commit -m "feat(worlds): RollResourceRating (WBH p.131)"
```

---

## Task 9: `Biology.Profile()` method + `runStep5F` orchestrator + DetailSystem wiring

**Files:**

- Modify: `worlds/biology.go` (Profile method)
- Modify: `worlds/system_detail_steps.go` (runStep5F + helpers)
- Modify: `worlds/system_detail.go` (DetailSystem wiring)
- Modify: `worlds/biology_test.go` (Profile + orchestrator tests)

- [ ] **Step 1: Write failing tests for `Profile()`**

Append to `worlds/biology_test.go`:

```go
func TestBiology_Profile_ZedPrime_A576(t *testing.T) {
	// Biomass=10/A, Biocomplexity=5, Biodiversity=7, Compatibility=6.
	// Per formula (not book worked example "A579"): "A576".
	bio := &Biology{Biomass: 10, Biocomplexity: 5, Biodiversity: 7, Compatibility: 6}
	got := bio.Profile()
	if got != "A576" {
		t.Errorf("got %q, want A576", got)
	}
}

func TestBiology_Profile_NoLife_Empty(t *testing.T) {
	bio := &Biology{Biomass: 0, Biocomplexity: 0, Biodiversity: 0, Compatibility: 0}
	got := bio.Profile()
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestBiology_Profile_eHexEncoding_AboveNine(t *testing.T) {
	// Biomass=15 → F, Biocomplexity=10 → A, Biodiversity=11 → B, Compatibility=14 → E.
	bio := &Biology{Biomass: 15, Biocomplexity: 10, Biodiversity: 11, Compatibility: 14}
	got := bio.Profile()
	if got != "FABE" {
		t.Errorf("got %q, want FABE", got)
	}
}

func TestBiology_Profile_AboveFifteen_SaturatesToF(t *testing.T) {
	// Defensive: values > 15 saturate to "F".
	bio := &Biology{Biomass: 20, Biocomplexity: 16, Biodiversity: 100, Compatibility: 999}
	got := bio.Profile()
	if got != "FFFF" {
		t.Errorf("got %q, want FFFF (saturate)", got)
	}
}

func TestBiology_Profile_NilReceiver_Empty(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panicked on nil: %v", r)
		}
	}()
	var bio *Biology
	got := bio.Profile()
	if got != "" {
		t.Errorf("got %q, want empty (nil receiver)", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run TestBiology_Profile ./worlds/...
```

Expected: FAIL with "undefined: bio.Profile" or similar.

- [ ] **Step 3: Implement `Profile()` method**

Append to `worlds/biology.go`:

```go
// Profile returns the WBH p.131 native-lifeform-profile MXDC eHex string
// (Biomass / Biocomplexity / Biodiversity / Compatibility). Returns ""
// when receiver is nil or Biomass is 0 (no native life to profile).
//
// eHex encoding: 0-9 → "0"-"9"; 10-15 → "A"-"F"; values > 15 saturate to "F".
func (b *Biology) Profile() string {
	if b == nil || b.Biomass == 0 {
		return ""
	}
	return string([]byte{
		eHexDigit(b.Biomass),
		eHexDigit(b.Biocomplexity),
		eHexDigit(b.Biodiversity),
		eHexDigit(b.Compatibility),
	})
}

// eHexDigit converts an int to a single eHex character byte. 0-9 → '0'-'9';
// 10-15 → 'A'-'F'; values > 15 saturate to 'F'; negative values clamp to '0'.
func eHexDigit(n int) byte {
	if n < 0 {
		return '0'
	}
	if n > 15 {
		return 'F'
	}
	if n < 10 {
		return byte('0' + n)
	}
	return byte('A' + (n - 10))
}
```

- [ ] **Step 4: Run Profile tests**

```bash
go test -run TestBiology_Profile ./worlds/... -v
```

Expected: all 5 tests PASS.

- [ ] **Step 5: Write failing orchestrator tests**

Append to `worlds/biology_test.go`:

```go
import "wbh/stars"

func TestRunStep5F_TerrestrialWithLife_PopulatesAll(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "5"
	dp.Designation = "Aab III"
	dp.Atmosphere = &Atmosphere{Code: 6}
	dp.Hydrographics = &Hydrographics{Code: 6}
	dp.Physical = &BodyPhysical{Density: 1.0}
	dp.Temperature = &Temperature{MeanK: 290, HighK: 310}

	sys := stars.System{Primary: stars.Star{AgeGyr: 5.0}}

	// Dice budget for a body with biomass > 0 + biocomplexity ≥ 8 path:
	// Biomass + Biocomplexity + 2 sophont + Biodiversity + Compatibility + Resource
	// = 7 dice. Provide a generous budget.
	r := roller.NewScripted(10, 10, 11, 11, 8, 7, 8)
	detailed := []DetailedPlacement{dp}
	if err := runStep5F(r, detailed, sys); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Biology == nil {
		t.Fatal("Biology is nil")
	}
	bio := detailed[0].Biology
	if bio.Biomass <= 0 {
		t.Errorf("Biomass: got %d, want > 0", bio.Biomass)
	}
	// Resource Rating should be in [2, 12].
	if bio.ResourceRating < 2 || bio.ResourceRating > 12 {
		t.Errorf("ResourceRating: got %d, want in [2, 12]", bio.ResourceRating)
	}
}

func TestRunStep5F_TerrestrialNoLife_OnlyResource(t *testing.T) {
	// Atm 0 + Hydro 0 + young → biomass should be 0.
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "5"
	dp.Designation = "Aab III"
	dp.Atmosphere = &Atmosphere{Code: 0}
	dp.Hydrographics = &Hydrographics{Code: 0}
	dp.Physical = &BodyPhysical{Density: 1.0}
	dp.Temperature = &Temperature{MeanK: 100, HighK: 100}

	sys := stars.System{Primary: stars.Star{AgeGyr: 0.1}}

	// Dice budget: Biomass roll (2D) + Resource roll (2D) = 2 dice when biomass=0.
	r := roller.NewScripted(2, 8)
	detailed := []DetailedPlacement{dp}
	if err := runStep5F(r, detailed, sys); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Biology == nil {
		t.Fatal("Biology is nil")
	}
	bio := detailed[0].Biology
	if bio.Biomass != 0 {
		t.Errorf("Biomass: got %d, want 0", bio.Biomass)
	}
	if bio.Biocomplexity != 0 || bio.Biodiversity != 0 || bio.Compatibility != 0 {
		t.Errorf("non-Biomass biology fields should be 0; got %+v", bio)
	}
	if bio.HasNativeSophont || bio.HadExtinctSophont {
		t.Error("sophont bools should be false")
	}
	if bio.ResourceRating < 2 || bio.ResourceRating > 12 {
		t.Errorf("ResourceRating should still be computed; got %d", bio.ResourceRating)
	}
}

func TestRunStep5F_GasGiant_NoBiology(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyGasGiant
	dp.GGClass = GasGiantSmall
	dp.Designation = "Aab IV"
	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5F(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Biology != nil {
		t.Error("GG should not get Biology")
	}
}

func TestRunStep5F_BeltSize0_NoBiology(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyPlanetoidBelt
	dp.SizeCode = "0"
	dp.Designation = "Aab Belt"
	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5F(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Biology != nil {
		t.Error("Belt should not get Biology")
	}
}

func TestRunStep5F_BodyEmpty_NoOp(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyEmpty
	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5F(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Biology != nil {
		t.Error("Empty body should not get Biology")
	}
}

func TestRunStep5F_TerrestrialNoAtmosphere_NoBiology(t *testing.T) {
	// Terrestrial without atmosphere → skipped (atm DM lookup needs atm code).
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "5"
	dp.Designation = "Aab III"
	dp.Atmosphere = nil
	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5F(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Biology != nil {
		t.Error("Terrestrial without atmosphere should not get Biology")
	}
}

func TestRunStep5F_MoonWithAtmosphere_GetsBiology(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "8"
	dp.Designation = "Aab III"
	dp.Atmosphere = &Atmosphere{Code: 6}
	dp.Hydrographics = &Hydrographics{Code: 6}
	dp.Physical = &BodyPhysical{Density: 1.0}
	dp.Temperature = &Temperature{MeanK: 290, HighK: 310}
	dp.Moons = []Moon{
		{
			Designation:   "Aab III a",
			SizeCode:      "5",
			Atmosphere:    &Atmosphere{Code: 6},
			Hydrographics: &Hydrographics{Code: 5},
			Physical:      &BodyPhysical{Density: 1.0},
			Temperature:   &Temperature{MeanK: 290, HighK: 310},
		},
	}

	sys := stars.System{Primary: stars.Star{AgeGyr: 5.0}}

	// Dice for parent (up to 7) + moon (up to 7) — be generous.
	r := roller.NewScripted(8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8)
	detailed := []DetailedPlacement{dp}
	if err := runStep5F(r, detailed, sys); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Biology == nil {
		t.Fatal("Parent Biology is nil")
	}
	if detailed[0].Moons[0].Biology == nil {
		t.Fatal("Moon Biology is nil")
	}
}

func TestRunStep5F_MoonNoAtmosphere_NoBiology(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "8"
	dp.Designation = "Aab III"
	dp.Atmosphere = &Atmosphere{Code: 6}
	dp.Hydrographics = &Hydrographics{Code: 6}
	dp.Physical = &BodyPhysical{Density: 1.0}
	dp.Temperature = &Temperature{MeanK: 290, HighK: 310}
	dp.Moons = []Moon{
		{
			Designation: "Aab III a",
			SizeCode:    "2",
			Atmosphere:  nil, // no atmosphere
		},
	}

	sys := stars.System{Primary: stars.Star{AgeGyr: 5.0}}

	r := roller.NewScripted(8, 8, 8, 8, 8, 8, 8)
	detailed := []DetailedPlacement{dp}
	if err := runStep5F(r, detailed, sys); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Moons[0].Biology != nil {
		t.Error("Moon without atmosphere should not get Biology")
	}
}

func TestRunStep5F_BiocomplexityBelowEight_NoSophontRolls(t *testing.T) {
	// Construct a body where biomass>0 but biocomplexity<8.
	// Roll dice such that biocomplexity ends up < 8 → sophont rolls skipped.
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "5"
	dp.Designation = "Aab III"
	dp.Atmosphere = &Atmosphere{Code: 11} // atm B → DM-2 for biocomplexity (not 4-9)
	dp.Hydrographics = &Hydrographics{Code: 6}
	dp.Physical = &BodyPhysical{Density: 1.0}
	dp.Temperature = &Temperature{MeanK: 290, HighK: 310}

	sys := stars.System{Primary: stars.Star{AgeGyr: 5.0}}

	// Dice: Biomass roll (will get bonus → biomass≥1) + Biocomplexity (small)
	// + Biodiversity + Compatibility + Resource = 5 dice.
	// (NO sophont rolls if biocomplexity < 8.)
	r := roller.NewScripted(8, 2, 6, 6, 8)
	detailed := []DetailedPlacement{dp}
	if err := runStep5F(r, detailed, sys); err != nil {
		t.Fatal(err)
	}
	bio := detailed[0].Biology
	if bio == nil {
		t.Fatal("Biology is nil")
	}
	if bio.Biomass == 0 {
		t.Skip("biomass came out 0; this test only meaningful with biomass>0")
	}
	if bio.Biocomplexity >= 8 {
		t.Skip("biocomplexity came out >=8; this test only meaningful when <8")
	}
	if bio.HasNativeSophont || bio.HadExtinctSophont {
		t.Errorf("sophont bools should be false when biocomplexity<8; got native=%v extinct=%v",
			bio.HasNativeSophont, bio.HadExtinctSophont)
	}
}
```

- [ ] **Step 6: Run tests to verify they fail**

```bash
go test -run TestRunStep5F ./worlds/...
```

Expected: FAIL with "undefined: runStep5F".

- [ ] **Step 7: Implement `runStep5F` + helpers in `worlds/system_detail_steps.go`**

Append to `worlds/system_detail_steps.go`:

```go
// runStep5F applies the 3B-biology pass: native lifeform ratings (Biomass,
// Biocomplexity, Sophont bools, Biodiversity, Compatibility) + Resource
// Rating. Mutates detailed in place. WBH pp.127-131.
//
// Body filter: terrestrials with Atmosphere; HZ-planet moons with
// Atmosphere. Skip GGs, belts (Size 0), empty placements, and terrestrials
// without atmosphere data.
//
// Per-body dice budget: up to 7 × 2D when life is present; 2 × 2D when
// biomass = 0 (Biomass + Resource); 0 × 2D when body is skipped.
//
//nolint:unparam // matches sibling runStep5* signatures (always-nil error)
func runStep5F(r roller.Roller, detailed []DetailedPlacement, sys stars.System) error {
	for i := range detailed {
		dp := &detailed[i]
		if !biologyApplies(dp) {
			continue
		}
		dp.Biology = computeBiology(r, dp, sys.Primary.AgeGyr)

		for j := range dp.Moons {
			m := &dp.Moons[j]
			if m.Atmosphere == nil {
				continue
			}
			moonDP := buildMoonPlacementView(m, dp)
			m.Biology = computeBiology(r, moonDP, sys.Primary.AgeGyr)
		}
	}
	return nil
}

// biologyApplies reports whether dp should receive a Biology struct.
// True for terrestrial bodies with Atmosphere; false for empty, belts,
// gas giants, or atmosphere-less terrestrials.
func biologyApplies(dp *DetailedPlacement) bool {
	if dp == nil || dp.Body == BodyEmpty {
		return false
	}
	if dp.Body == BodyGasGiant || dp.Body == BodyPlanetoidBelt {
		return false
	}
	if dp.SizeCode == "0" {
		return false
	}
	return dp.Atmosphere != nil
}

// computeBiology populates a Biology for the given body. Caller has
// already verified biologyApplies(dp).
func computeBiology(r roller.Roller, dp *DetailedPlacement, ageGyr float64) *Biology {
	bio := &Biology{}
	bio.Biomass = RollBiomass(r, dp, ageGyr)
	if bio.Biomass > 0 {
		bio.Biocomplexity = RollBiocomplexity(r, dp, bio.Biomass, ageGyr)
		if bio.Biocomplexity >= 8 {
			bio.HasNativeSophont = RollNativeSophont(r, bio.Biocomplexity)
			bio.HadExtinctSophont = RollExtinctSophont(r, bio.Biocomplexity, ageGyr)
		}
		bio.Biodiversity = RollBiodiversity(r, bio.Biomass, bio.Biocomplexity)
		bio.Compatibility = RollCompatibility(r, dp, bio.Biocomplexity, ageGyr)
	}
	bio.ResourceRating = RollResourceRating(r, dp, bio)
	return bio
}
```

- [ ] **Step 8: Wire `runStep5F` into `DetailSystem`**

In `worlds/system_detail.go`, find the existing block:

```go
	// Step 5E — 3B-geology pass: seismic + GG residual heat + temp recompute + tectonic plates.
	if err := runStep5E(r, detailed, sys); err != nil {
		return SystemDetail{}, err
	}
```

Append immediately after it:

```go
	// Step 5F — 3B-biology pass: native lifeform ratings + resource rating.
	if err := runStep5F(r, detailed, sys); err != nil {
		return SystemDetail{}, err
	}
```

- [ ] **Step 9: Run all tests**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
just check
just test
```

Expected: 0 issues; all packages pass. The existing 100-iteration `TestZed_FullDetail_3A2b` should still pass — runStep5F is wired in but Task 10's new assertions haven't landed yet.

- [ ] **Step 10: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/biology.go \
        worlds/biology_test.go \
        worlds/system_detail_steps.go \
        worlds/system_detail.go
git commit -m "feat(worlds): Biology.Profile + runStep5F orchestrator + DetailSystem wiring (WBH pp.127-131)"
```

---

## Task 10: Acceptance test extension on `TestZed_FullDetail_3A2b`

**Files:**

- Modify: `worlds/worked_examples_test.go`

- [ ] **Step 1: Locate the existing assertions**

```bash
grep -n "Assertion 31\|3B-geology: post-TSS" worlds/worked_examples_test.go | head -3
```

Identify the location of the 31-and-prior assertions inside the iter loop, and the 5 trailing `t.Logf` notes after the loop.

- [ ] **Step 2: Append assertions 32-38 inside the iter loop**

After Assertion 31's accumulation block and BEFORE the iter loop's closing `}`, insert:

```go
		// 3B-biology invariants (assertions 32-38).

		// Assertion 32: HasBiology() for terrestrial bodies with Atmosphere
		// (and HZ-planet moons with Atmosphere). Skip belts, GGs, empty.
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.Body != BodyTerrestrial || dp.Atmosphere == nil {
				continue
			}
			if dp.SizeCode == "0" {
				continue
			}
			if !dp.HasBiology() {
				t.Errorf("iter %d: body %s missing Biology", iter, dp.Designation)
			}
		}

		// Assertion 33: When Biomass > 0, all biology fields populated.
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if !dp.HasBiology() {
				continue
			}
			b := dp.Biology
			if b.Biomass > 0 {
				if b.Biocomplexity <= 0 {
					t.Errorf("iter %d: body %s: Biomass=%d but Biocomplexity=%d",
						iter, dp.Designation, b.Biomass, b.Biocomplexity)
				}
				if b.Biodiversity <= 0 {
					t.Errorf("iter %d: body %s: Biomass=%d but Biodiversity=%d",
						iter, dp.Designation, b.Biomass, b.Biodiversity)
				}
				// Compatibility ≥ 0 is the floor; 0 is valid (incompatible).
				if b.Compatibility < 0 {
					t.Errorf("iter %d: body %s: Compatibility=%d should be ≥ 0",
						iter, dp.Designation, b.Compatibility)
				}
			}
		}

		// Assertion 34: When Biomass == 0, biology rating fields are 0; sophont bools false.
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if !dp.HasBiology() {
				continue
			}
			b := dp.Biology
			if b.Biomass == 0 {
				if b.Biocomplexity != 0 || b.Biodiversity != 0 || b.Compatibility != 0 {
					t.Errorf("iter %d: body %s: Biomass=0 but ratings non-zero (X=%d D=%d C=%d)",
						iter, dp.Designation, b.Biocomplexity, b.Biodiversity, b.Compatibility)
				}
				if b.HasNativeSophont || b.HadExtinctSophont {
					t.Errorf("iter %d: body %s: Biomass=0 but sophont bool true",
						iter, dp.Designation)
				}
			}
		}

		// Assertion 35: ResourceRating in [2, 12] for all bodies with Biology.
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if !dp.HasBiology() {
				continue
			}
			rr := dp.Biology.ResourceRating
			if rr < 2 || rr > 12 {
				t.Errorf("iter %d: body %s: ResourceRating=%d out of [2,12]",
					iter, dp.Designation, rr)
			}
		}

		// Assertion 36: Sophont bools only true when Biocomplexity ≥ 8.
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if !dp.HasBiology() {
				continue
			}
			b := dp.Biology
			if (b.HasNativeSophont || b.HadExtinctSophont) && b.Biocomplexity < 8 {
				t.Errorf("iter %d: body %s: sophont bool true but Biocomplexity=%d < 8",
					iter, dp.Designation, b.Biocomplexity)
			}
		}

		// Assertion 37: Profile() returns "" or 4-char eHex string.
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if !dp.HasBiology() {
				continue
			}
			p := dp.Biology.Profile()
			if p != "" && len(p) != 4 {
				t.Errorf("iter %d: body %s: Profile()=%q, want \"\" or 4-char string",
					iter, dp.Designation, p)
			}
		}

		// Assertion 38: smoke test against silent-zero — count bodies with Biomass ≥ 1.
		// Accumulator declared before iter loop; assertion reported after loop.
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.HasBiology() && dp.Biology.Biomass >= 1 {
				totalBiomassNonzero++
			}
			for j := range dp.Moons {
				m := &dp.Moons[j]
				if m.HasBiology() && m.Biology.Biomass >= 1 {
					totalBiomassNonzero++
				}
			}
		}
```

- [ ] **Step 3: Add the `totalBiomassNonzero` accumulator and post-loop check**

Find the existing `totalTHFNonzero` accumulator declaration (just before the iter loop). Add a sibling:

```go
	totalBiomassNonzero := 0
```

After the iter loop completes (with the existing `totalTHFNonzero` assertion + log), add:

```go
	if totalBiomassNonzero == 0 {
		t.Errorf("integration: Biomass was zero for ALL bodies across 100 iterations — likely a silent-zero bug")
	}
	t.Logf("3B-biology: %d body-iterations had non-zero Biomass across 100-iter sweep", totalBiomassNonzero)
```

- [ ] **Step 4: Add the 6th trailing `t.Logf` note**

After the existing 5 trailing `t.Logf` notes, add:

```go
	t.Logf("3B-biology: Compatibility formula follows WBH p.131 formula box; book worked example shows 9.5 for Zed Prime but lacks a source for the +3 — implementation gives 6")
```

- [ ] **Step 5: Run the acceptance test**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test -run TestZed_FullDetail_3A2b ./worlds/... -v 2>&1 | tail -30
```

Expected: PASS across all 100 iterations. Two non-zero counts reported (THF + Biomass).

If `totalBiomassNonzero == 0` even with the fix, that means the typical Zed fixture doesn't produce any worlds with biomass — investigate before degrading the assertion.

- [ ] **Step 6: just check && just test**

```bash
just check
just test
```

Expected: 0 issues; all packages pass.

- [ ] **Step 7: Commit**

```bash
cd /Users/markayers/Documents/Traveller
git add worlds/worked_examples_test.go
git commit -m "test(worlds): extend TestZed_FullDetail_3A2b with 3B-biology assertions"
```

---

## Task 11: Final end-to-end review on Opus + merge

**Files:** none (review-only task)

- [ ] **Step 1: Verify branch state**

```bash
cd /Users/markayers/Documents/Traveller
git log --oneline main..HEAD
```

Expected: 10 commits (one per Task 1-10).

- [ ] **Step 2: Final review subagent (Opus)**

Dispatch `superpowers:code-reviewer` (or `code-reviewer` agent) on the entire branch with model=opus. Provide:

- Branch name: `feat/wbh-world-physical-3b-biology`
- Spec path: `docs/history/pass-1-specs/2026-05-05-world-physical-3b-biology-design.md`
- Plan path: `docs/history/pass-1-plans/2026-05-05-world-physical-3b-biology.md`
- Diff command: `git -C /Users/markayers/Documents/Traveller diff main..feat/wbh-world-physical-3b-biology -- `

Reviewer should report: spec-compliance issues, code-quality issues, cross-cutting concerns, merge readiness assessment. Specifically watch for C1-style integration silent-zero bugs (precedent: 3B-geology's terrestrial MassEarth=0 trap).

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
git merge --no-ff feat/wbh-world-physical-3b-biology -m "Merge feat/wbh-world-physical-3b-biology: World Physical 3B-Biology complete

WBH pp.127-131: Biomass Rating, Biocomplexity Rating, Native Sophonts
(extant + extinct), Biodiversity Rating, Compatibility Rating, Native
Lifeform Profile (MXDC eHex), Resource Rating.

Implemented as a single new pipeline step runStep5F between 3B-geology
and Step 6. Biology struct attached to DetailedPlacement and Moon.
worlds/biology.go holds 7 standalone helper functions + the Profile()
method.

First sub-project after the temperature feedback edge — pure forward
flow. Pipeline now complete through pp.131."

just check && just test
```

- [ ] **Step 6: Update memory**

After merge:

1. Update `MEMORY.md` Subprojects line to mark 3B-biology complete with merge SHA
2. Update `project_world_builder_3b_kickoff.md` to mark biology done; next is 3B-final
3. Save book-inconsistency feedback memories:
   - `feedback_wbh_p131_compatibility_formula.md` — book worked example diverges from formula box ("+3" with no source)
   - (Atm codes G/H ignored — no separate memory needed; spec doc-comment captures it)

- [ ] **Step 7: Confirm clean tree**

```bash
git status
```

Expected: clean.

---

## Self-review checklist (run after writing this plan)

- [x] **Spec coverage:** Every section of the spec maps to a task. Procedure steps 1-7 → Tasks 2-8 (formula helpers) + Task 9 (Profile + orchestrator + wiring). Architecture → Task 1 (struct), Task 9 (orchestrator). Sub-decisions → addressed in tests + spec doc-comments. Testing → Tasks 2-10. Acceptance test → Task 10. Final review → Task 11.
- [x] **Placeholder scan:** No TBD/TODO/incomplete sections. All code blocks complete. Every step has runnable commands and concrete expected output.
- [x] **Type consistency:** `RollBiomass(r, body, ageGyr) int`, `RollBiocomplexity(r, body, biomass, ageGyr) int`, `RollNativeSophont(r, biocomplexity) bool`, `RollExtinctSophont(r, biocomplexity, ageGyr) bool`, `RollBiodiversity(r, biomass, biocomplexity) int`, `RollCompatibility(r, body, biocomplexity, ageGyr) int`, `RollResourceRating(r, body, bio) int`, `(b *Biology) Profile() string`, `runStep5F(r, detailed, sys) error` — all signatures consistent across tasks.

## Known WBH inconsistencies to log during implementation

After merge, save as feedback memories:

1. **WBH p.131 Compatibility worked example contradicts formula box.** Formula says `2D − Biocomplexity/2 + DMs`; worked example shows `7 + 3 − 2.5 + 2 = 9.5`. Implementation follows formula → Zed Prime gets 6 (book says 9). Profile becomes "A576" (not "A579").
2. **Atmosphere codes G and H** appear in the Compatibility DM table (p.131) but don't exist in the WBH atmosphere code system (0-F). Implementation ignores them — they cannot be produced by RollAtmoCode. Captured inline in `compatibilityAtmDM` doc-comment.
