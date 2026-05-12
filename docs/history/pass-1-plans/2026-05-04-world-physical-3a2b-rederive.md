# World Physical 3A2b-rederive Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement WBH p.79 (Optional Runaway Greenhouse), p.81 (scale height under real T), pp.94-98 (atm gas-mix re-derivation under real TempRange), p.99 (atm composition tail), p.102 (Possible Exotic Liquids) — re-deriving 3A1's temperature-sensitive Atmosphere/Hydrographics fields under `body.Temperature.MeanK` (now available from 3A2b-temp). After this sub-project, **3A2b is complete**.

**Architecture:** Stay flat in `worlds/`. Three new production files (`temperature_rederive.go`, `runaway_greenhouse.go`, `exotic_liquids.go`) plus extensions to `system_detail.go`, `atmosphere.go`, `hydrographics.go`, `body_physical.go`, `worked_examples_test.go`. Step 5D wired into DetailSystem orchestrator with 2-pass iteration: rederive → re-run GenerateTemperature → rederive again.

**Tech Stack:** Go 1.22+, `wbh/roller` (scripted dice), `wbh/dice`, `wbh/stars`, `wbh/worlds` (existing 2A/2B/2C/3A1/3A2a/3A2b-temp). Justfile targets: `just check` (gofumpt + vet + golangci-lint), `just test` (`go test -race ./...`).

---

## Spec reference

`docs/history/pass-1-specs/2026-05-04-world-physical-3a2b-rederive-design.md` (committed `f2a96c3`) — read first if unfamiliar.

## Dice convention (CRITICAL — caused 4 bugs in 2C and 6+ in 3A1; per-task reviews caught more in 3A2a/3A2b-temp)

Per `roller/roller.go:47-50`, scripted values are **final results, one per `Roll()` call regardless of dice notation**. When the book says "2D=8 + DM+1 = 9", the scripted value is **8** (the 2D pre-DM); the DM is applied in code. Each `r.Roll("2D")` call consumes exactly one scripted value.

Every implementation task must call this out at the top of the subagent brief.

## Roller API

- Constructor: `roller.NewScripted(results ...int) *Scripted`
- Method: `Roll(notation string) int` (returns int, no error)
- Notations: `"2D"`, `"1D"`, `"3D"`, `"D3"`, `"d10"`, `"d100"`, `"2D+2"` (NOT `"2D2"` which parses as 2 dice of 2 sides)
- Exhaustion: panics — used as test bug indicator

## Existing types verified before plan was written

- `worlds.Atmosphere.Code` is `int`; `Atmosphere.Subtype` is `string`; `Atmosphere.Pressure float64` (bar); `Atmosphere.ScaleHeight float64` (km); `Atmosphere.Profile AtmosphereProfile` (value, not pointer)
- `worlds.AtmosphereProfile` has `Gases []GasFraction` and `TempRange string`
- `worlds.Hydrographics.Code int` (no Profile field yet — Task 5 adds)
- `worlds.BodyPhysical.Density float64`
- `dp.Temperature *Temperature` (added by 3A2b-temp Task 9)
- `m.Temperature *Temperature` (added by 3A2b-temp Task 12)
- `dp.Group.HZCO()` returns Orbit#
- `MeanK` field on `Temperature` populated by 3A2b-temp `GenerateTemperature`

**Existing 3A1 helpers reusable for rederive:**

- `worlds.DeriveScaleHeight(meanTempK, gravityG float64) float64` — already takes meanK directly (no TempRange bucketing)
- `worlds.RollTotalPressure(r, atmoCode int) (float64, error)` — re-roll pressure within current code's range
- `worlds.RollCorrosiveInsidiousSubtype(r, sizeCode, orbit, hzco, isInsidious, runawayResult bool) (string, error)` — already accepts `runawayResult bool` for the runaway DM+4
- `worlds.RollHydroDigit(r, atmoCode, atmoSubtype, sizeCode, tempRange) (int, error)` — re-roll hydrographics with current TempRange
- `worlds.RollGasMix(r, atmoSubtype, _, tempRange, sizeCode) (AtmosphereProfile, error)` — re-roll exotic atm composition
- `worlds.HZRegionAtmosphereDM(code int) int` — used by 3A2b-temp; reuse if needed
- `worlds.SizeAsInt(SizeCode) int`

**3A1 gap (not in scope for 3A2b-rederive):**

`RollAtmoCode` uses Core Rulebook's `2D-7+Size` formula for ALL bodies; Hot/Cold/Frozen Atmosphere tables (pp.94-95) for non-HZ atm code re-roll were never implemented in 3A1. This means **non-HZ atm CODE re-derivation is a no-op** for 3A2b-rederive (nothing to re-derive). The atm SUBTYPE (codes B/C) is still re-derivable via `RollCorrosiveInsidiousSubtype` if runaway greenhouse fires.

## File structure

| File                                  | New / Modified | Responsibility                                                                                                             |
| ------------------------------------- | -------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `worlds/exotic_liquids.go`            | New            | `PossibleExoticLiquids` table + `SelectExoticLiquid(meanK, atmCode) string` deterministic picker                           |
| `worlds/runaway_greenhouse.go`        | New            | `CheckRunawayGreenhouse(r, body, sys) bool` per WBH p.79                                                                   |
| `worlds/temperature_rederive.go`      | New            | `MeanKToTempRange` helper + `RederiveAtmosphereHydrographics` orchestrator + private helpers                               |
| `worlds/temperature_rederive_test.go` | New            | All 3A2b-rederive unit tests                                                                                               |
| `worlds/hydrographics.go`             | Modify         | Add `Hydrographics.Profile string` field; add `DeriveHydrographicsProfile(meanK, atmCode, hydroCode) string` helper        |
| `worlds/system_detail.go`             | Modify         | Wire **Step 5D** (2-pass iteration) after Step 5C                                                                          |
| `worlds/atmosphere.go`                | Modify         | Update `Atmosphere` doc comment to remove "provisional under HZCO temperature" qualifier; mention post-5D values are final |
| `worlds/body_physical.go`             | Modify         | Update doc comment to remove "provisional" qualifier                                                                       |
| `worlds/worked_examples_test.go`      | Modify         | Replace `TestZed_FullDetail_3A2b_temp` with `TestZed_FullDetail_3A2b`                                                      |

## Branch

`feat/wbh-world-physical-3a2b-rederive` — created off `main` at the merge of 3A2b-temp (`0e4e73b`).

---

## Task 1: Branch setup + smoke check

**Files:**

- (none modified)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
git checkout main
git checkout -b feat/wbh-world-physical-3a2b-rederive
git status
```

Expected: `On branch feat/wbh-world-physical-3a2b-rederive`, `nothing to commit, working tree clean`.

- [ ] **Step 2: Verify project is green**

```bash
just check && just test
```

Expected: `0 issues.`; all packages `ok`.

- [ ] **Step 3: No commit needed; proceed to Task 2.**

---

## Task 2: PossibleExoticLiquids table + SelectExoticLiquid (WBH p.102)

**Files:**

- Create: `worlds/exotic_liquids.go`
- Create: `worlds/temperature_rederive_test.go` (this is the project's first 3A2b-rederive test file)

### Step 1: Write the failing tests

Create `worlds/temperature_rederive_test.go`:

```go
package worlds

import (
	"testing"
)

func TestSelectExoticLiquid_Water_Terra(t *testing.T) {
	// At meanK=288 with atm A (10), water (Abundance 100) wins among candidates
	// where 273 ≤ 288 ≤ 373.
	got := SelectExoticLiquid(288, 10)
	if got != "H2O" {
		t.Errorf("got %q, want H2O", got)
	}
}

func TestSelectExoticLiquid_Methane_Cold(t *testing.T) {
	// At meanK=100 with atm B (11), methane (range 91-113, Abundance 70) wins.
	got := SelectExoticLiquid(100, 11)
	if got != "CH4" {
		t.Errorf("got %q, want CH4", got)
	}
}

func TestSelectExoticLiquid_Ethane_NotMethaneAtMid(t *testing.T) {
	// At meanK=150, methane (boils at 113) is out; ethane (range 90-184, Abundance 70) wins.
	got := SelectExoticLiquid(150, 11)
	if got != "C2H6" {
		t.Errorf("got %q, want C2H6", got)
	}
}

func TestSelectExoticLiquid_NoCandidate_TooHot(t *testing.T) {
	// At meanK=2000, all candidates' boiling points are exceeded.
	got := SelectExoticLiquid(2000, 10)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestSelectExoticLiquid_NonExoticAtm_Empty(t *testing.T) {
	// Atm 6 (standard) → defensive: caller shouldn't call; return empty.
	got := SelectExoticLiquid(288, 6)
	if got != "" {
		t.Errorf("got %q, want empty (non-exotic atm)", got)
	}
}

func TestSelectExoticLiquid_TieBreakLowerBoiling(t *testing.T) {
	// Construct a meanK where two candidates tie on Abundance — verify lower
	// BoilingK wins. CH4 (91-113, 70) and C2H6 (90-184, 70) both contain 100K.
	// Lower BoilingK is CH4 (113 < 184) → CH4 wins.
	got := SelectExoticLiquid(100, 11)
	if got != "CH4" {
		t.Errorf("got %q, want CH4 (tie-break by lower BoilingK)", got)
	}
}
```

### Step 2: Verify failure

```bash
go test -run 'TestSelectExoticLiquid' ./worlds/...
```

Expected: build error `undefined: SelectExoticLiquid`.

### Step 3: Write implementation

Create `worlds/exotic_liquids.go`:

```go
package worlds

// ExoticLiquid is one row of the WBH p.102 Possible Exotic Liquids table.
type ExoticLiquid struct {
	Code      string  // "H2O", "CH4", "NH3", etc.
	MeltingK  float64
	BoilingK  float64
	Abundance int // Relative Abundance, 1..100
}

// PossibleExoticLiquids — WBH p.102 table, 15 entries ordered by boiling point.
var PossibleExoticLiquids = []ExoticLiquid{
	{"F2", 53, 85, 2},
	{"O2", 54, 90, 50},
	{"CH4", 91, 113, 70},
	{"C2H6", 90, 184, 70},
	{"Cl2", 171, 239, 1},
	{"NH3", 195, 240, 30},
	{"SO2", 201, 263, 20},
	{"HF", 190, 293, 2},
	{"HCN", 260, 299, 30},
	{"HCl", 247, 321, 1},
	{"H2O", 273, 373, 100},
	{"CH2O2", 281, 374, 15},
	{"CH3NO", 275, 483, 15},
	{"H2CO3", 193, 607, 20},
	{"H2SO4", 388, 718, 20},
}

// isExoticAtmCode reports whether atmCode requires exotic-liquid selection
// (Atm A=10, B=11, C=12, F=15 per p.102).
func isExoticAtmCode(atmCode int) bool {
	return atmCode == 10 || atmCode == 11 || atmCode == 12 || atmCode == 15
}

// SelectExoticLiquid returns the dominant liquid for a body with exotic
// atmosphere (Atm A-C/F: codes 10, 11, 12, 15) and non-zero hydrographics
// at the given mean temperature.
//
// Deterministic: among molecules where MeltingK ≤ meanK ≤ BoilingK, returns
// the highest-Abundance candidate. Ties broken by lower BoilingK (more
// "stable" in range). Returns "" if no candidate fits or atmCode is not exotic.
func SelectExoticLiquid(meanK float64, atmCode int) string {
	if !isExoticAtmCode(atmCode) {
		return ""
	}
	bestCode := ""
	bestAbundance := -1
	bestBoiling := 0.0
	for _, l := range PossibleExoticLiquids {
		if meanK < l.MeltingK || meanK > l.BoilingK {
			continue
		}
		if l.Abundance > bestAbundance ||
			(l.Abundance == bestAbundance && l.BoilingK < bestBoiling) {
			bestCode = l.Code
			bestAbundance = l.Abundance
			bestBoiling = l.BoilingK
		}
	}
	return bestCode
}
```

### Step 4: Verify

```bash
go test -run 'TestSelectExoticLiquid' ./worlds/... -v
```

Expected: 6/6 PASS.

### Step 5: `just check && just test`

Expected green.

### Step 6: Commit

```bash
git add worlds/exotic_liquids.go worlds/temperature_rederive_test.go
git commit -m "feat(worlds): PossibleExoticLiquids table + SelectExoticLiquid (WBH p.102)"
```

---

## Task 3: MeanKToTempRange helper

**Files:**

- Create: `worlds/temperature_rederive.go`
- Append to: `worlds/temperature_rederive_test.go`

### Step 1: Append failing tests

Append to `worlds/temperature_rederive_test.go`:

```go
func TestMeanKToTempRange_Boundaries(t *testing.T) {
	cases := []struct {
		meanK float64
		want  TempRange
	}{
		{50, TempFrozen},
		{122, TempFrozen},
		{123, TempCold},
		{200, TempCold},
		{272, TempCold},
		{273, TempTemperate},
		{300, TempTemperate},
		{352, TempTemperate},
		{353, TempHot},
		{400, TempHot},
		{452, TempHot},
		{453, TempBoiling},
		{1000, TempBoiling},
	}
	for _, c := range cases {
		if got := MeanKToTempRange(c.meanK); got != c.want {
			t.Errorf("meanK=%v: got %v, want %v", c.meanK, got, c.want)
		}
	}
}
```

### Step 2: Verify failure

```bash
go test -run 'TestMeanKToTempRange' ./worlds/...
```

Expected: `undefined: MeanKToTempRange`.

### Step 3: Write implementation

Create `worlds/temperature_rederive.go`:

```go
// Package worlds — atmosphere/hydrographics re-derivation under real
// temperature per WBH p.79, p.81, pp.94-98, p.99, p.102 (sub-project 3A2b-rederive).
package worlds

// MeanKToTempRange buckets a real mean temperature in Kelvin into the same
// TempRange bands 3A1's HZCOOffsetToTempRange used (WBH pp.94-98 keying):
//
//	≥ 453K → Boiling
//	353-453K → Hot
//	273-353K → Temperate
//	123-273K → Cold
//	< 123K → Frozen
func MeanKToTempRange(meanK float64) TempRange {
	switch {
	case meanK >= 453:
		return TempBoiling
	case meanK >= 353:
		return TempHot
	case meanK >= 273:
		return TempTemperate
	case meanK >= 123:
		return TempCold
	default:
		return TempFrozen
	}
}
```

### Step 4: Verify

```bash
go test -run 'TestMeanKToTempRange' ./worlds/... -v
```

Expected: PASS (single test with 13 sub-cases).

### Step 5: `just check && just test`

Expected green.

### Step 6: Commit

```bash
git add worlds/temperature_rederive.go worlds/temperature_rederive_test.go
git commit -m "feat(worlds): MeanKToTempRange helper for rederive bucketing"
```

---

## Task 4: CheckRunawayGreenhouse standalone (WBH p.79)

**Files:**

- Create: `worlds/runaway_greenhouse.go`
- Append to: `worlds/temperature_rederive_test.go`

### Step 1: Append failing tests

Append to `worlds/temperature_rederive_test.go`:

```go
import (
	// existing: "testing"
	"wbh/roller"
	"wbh/stars"
)

// (Adjust imports as needed; if existing imports already include roller/stars, no change.)

func TestCheckRunawayGreenhouse_BelowTempThreshold(t *testing.T) {
	// meanK=300 < 303K threshold → false regardless of dice.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Temperature = &Temperature{MeanK: 300}
	sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

	r := roller.NewScripted() // no dice consumed when guard short-circuits
	if got := CheckRunawayGreenhouse(r, body, sys); got {
		t.Error("expected false (meanK below 303K)")
	}
}

func TestCheckRunawayGreenhouse_LowAtmCode(t *testing.T) {
	// atm 1 (out of [2, F-1] excluding A/B/C) → false.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 1}
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

	r := roller.NewScripted()
	if got := CheckRunawayGreenhouse(r, body, sys); got {
		t.Error("expected false (atm 1 out of trigger range)")
	}
}

func TestCheckRunawayGreenhouse_AtmAlreadyExotic_Skipped(t *testing.T) {
	// atm A, B, C, F+ skip per MVP — book's "consider boiling" deferred.
	for _, code := range []int{10, 11, 12, 15} {
		body := &DetailedPlacement{}
		body.Atmosphere = &Atmosphere{Code: code}
		body.Temperature = &Temperature{MeanK: 400}
		sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

		r := roller.NewScripted()
		if got := CheckRunawayGreenhouse(r, body, sys); got {
			t.Errorf("atm %d: expected false (exotic atm skipped per MVP)", code)
		}
	}
}

func TestCheckRunawayGreenhouse_LowDiceRoll(t *testing.T) {
	// atm 6, meanK=400, sysAge=1 → DM+1 (age round up). 2D=2 + 1 = 3 < 12 → false.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 1}}

	r := roller.NewScripted(2)
	if got := CheckRunawayGreenhouse(r, body, sys); got {
		t.Error("expected false (mod=3, below 12)")
	}
}

func TestCheckRunawayGreenhouse_Triggered_AtmA(t *testing.T) {
	// atm 6, meanK=400 (< 388 boiling, no +4 DM), sysAge=5 → DM+5.
	// 2D=7 + 5 = 12 → trigger. 1D=1 → atm becomes A (10).
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.SizeCode = "8"
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

	r := roller.NewScripted(7, 1)
	if got := CheckRunawayGreenhouse(r, body, sys); !got {
		t.Error("expected true (mod=12, triggered)")
	}
	if body.Atmosphere.Code != 10 {
		t.Errorf("atm code: got %d, want 10 (A)", body.Atmosphere.Code)
	}
}

func TestCheckRunawayGreenhouse_Triggered_AtmB(t *testing.T) {
	// Same as above but 1D=3 → atm becomes B (11).
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.SizeCode = "8"
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

	r := roller.NewScripted(7, 3)
	if got := CheckRunawayGreenhouse(r, body, sys); !got {
		t.Error("expected true")
	}
	if body.Atmosphere.Code != 11 {
		t.Errorf("atm code: got %d, want 11 (B)", body.Atmosphere.Code)
	}
}

func TestCheckRunawayGreenhouse_Triggered_AtmC(t *testing.T) {
	// 1D=6 → atm becomes C (12).
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.SizeCode = "8"
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

	r := roller.NewScripted(7, 6)
	if got := CheckRunawayGreenhouse(r, body, sys); !got {
		t.Error("expected true")
	}
	if body.Atmosphere.Code != 12 {
		t.Errorf("atm code: got %d, want 12 (C)", body.Atmosphere.Code)
	}
}

func TestCheckRunawayGreenhouse_BoilingTempDM(t *testing.T) {
	// meanK=400 < 388? No, 400 ≥ 388 → DM+4 from boiling threshold.
	// Plus sysAge=1 → DM+1. 2D=2 + 4 + 1 = 7 < 12 → still false.
	// Bump sysAge to 7 → DM+7. 2D=2 + 4 + 7 = 13 → trigger.
	body := &DetailedPlacement{}
	body.Atmosphere = &Atmosphere{Code: 6}
	body.SizeCode = "8"
	body.Temperature = &Temperature{MeanK: 400}
	sys := stars.System{Primary: stars.Star{AgeGyr: 7}}

	r := roller.NewScripted(2, 1)
	if got := CheckRunawayGreenhouse(r, body, sys); !got {
		t.Error("expected true (boiling DM+4 + age DM+7)")
	}
}
```

### Step 2: Verify failure

```bash
go test -run 'TestCheckRunawayGreenhouse' ./worlds/...
```

Expected: `undefined: CheckRunawayGreenhouse`.

### Step 3: Write implementation

Create `worlds/runaway_greenhouse.go`:

```go
package worlds

import (
	"math"

	"wbh/roller"
	"wbh/stars"
)

// CheckRunawayGreenhouse evaluates and applies WBH p.79 Optional Runaway
// Greenhouse. Triggered when:
//   - body.Atmosphere is non-nil AND atm.Code in {2-9, D=13, E=14}
//   - body.Temperature is non-nil AND MeanK > 303K
//   - 2D + DMs ≥ 12
//
// DMs:
//   - +1 per System Age Gyr (round up)
//   - +4 if mean T ≥ 388K (boiling temperature, 12+ on basic temp table)
//
// On trigger, mutates atm.Code via 1D table:
//
//	1   → A (10)
//	2-4 → B (11)
//	5+  → C (12)
//
// (Size DM-2 if Size 2-5; Tainted DM+1 if originally codes 2/4/7/9 — applied
// to the trigger roll, not the 1D atm-code roll.)
//
// Returns true iff trigger fired. Caller re-rolls Hydrographics with DM-6
// (boiling) instead of DM-2 (hot) when this returns true.
//
// MVP simplification: atm A/B/C (10/11/12) and F+ (15+) skip this check.
// The book's "consider boiling" case for those codes (only flips hydrographics
// DM without mutating atm code) is deferred — see spec carry-forwards.
func CheckRunawayGreenhouse(r roller.Roller, body *DetailedPlacement, sys stars.System) bool {
	if body.Atmosphere == nil || body.Temperature == nil {
		return false
	}
	if body.Temperature.MeanK <= 303 {
		return false
	}
	code := body.Atmosphere.Code
	// Trigger range: atm 2-9, D (13), E (14). Skip A/B/C (10-12), F+ (15+), and 0/1.
	if code < 2 || code == 10 || code == 11 || code == 12 || code >= 15 {
		return false
	}

	// Trigger roll: 2D + DMs.
	dm := 0
	// Age DM: +1 per Gyr round up.
	dm += int(math.Ceil(sys.Primary.AgeGyr))
	// Boiling DM: +4 if meanK ≥ 388K (12+ on basic temp table).
	if body.Temperature.MeanK >= 388 {
		dm += 4
	}
	// Tainted DM: +1 if originally codes 2, 4, 7, 9.
	if code == 2 || code == 4 || code == 7 || code == 9 {
		dm++
	}
	// Size DM: -2 if Size 2-5.
	si := SizeAsInt(body.SizeCode)
	if si >= 2 && si <= 5 {
		dm -= 2
	}

	roll := r.Roll("2D")
	if roll+dm < 12 {
		return false
	}

	// Trigger fired: roll 1D for new atm code.
	atmRoll := r.Roll("1D")
	switch {
	case atmRoll == 1:
		body.Atmosphere.Code = 10 // A
	case atmRoll <= 4:
		body.Atmosphere.Code = 11 // B
	default:
		body.Atmosphere.Code = 12 // C
	}
	return true
}
```

### Step 4: Verify

```bash
go test -run 'TestCheckRunawayGreenhouse' ./worlds/... -v
```

Expected: 8/8 PASS.

### Step 5: `just check && just test`

Expected green.

### Step 6: Commit

```bash
git add worlds/runaway_greenhouse.go worlds/temperature_rederive_test.go
git commit -m "feat(worlds): CheckRunawayGreenhouse with atm code mutation (WBH p.79)"
```

---

## Task 5: Hydrographics.Profile field + DeriveHydrographicsProfile helper

**Files:**

- Modify: `worlds/hydrographics.go`
- Append to: `worlds/temperature_rederive_test.go`

### Step 1: Append failing tests

Append to `worlds/temperature_rederive_test.go`:

```go
func TestDeriveHydrographicsProfile_Water(t *testing.T) {
	// Atm 6 (standard), hydro 6, meanK 288 → "H6:H2O-100".
	got := DeriveHydrographicsProfile(288, 6, 6)
	if got != "H6:H2O-100" {
		t.Errorf("got %q, want H6:H2O-100", got)
	}
}

func TestDeriveHydrographicsProfile_ExoticMethane(t *testing.T) {
	// Atm A (10), hydro 4, meanK 100 → "H4:CH4-100" (methane wins by Abundance 70).
	got := DeriveHydrographicsProfile(100, 10, 4)
	if got != "H4:CH4-100" {
		t.Errorf("got %q, want H4:CH4-100", got)
	}
}

func TestDeriveHydrographicsProfile_Vacuum_Empty(t *testing.T) {
	// Atm 0, hydro 0 → empty (no liquid).
	got := DeriveHydrographicsProfile(288, 0, 0)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestDeriveHydrographicsProfile_NoHydro_Empty(t *testing.T) {
	// Atm 6, hydro 0 → empty (no liquid surface).
	got := DeriveHydrographicsProfile(288, 6, 0)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestDeriveHydrographicsProfile_HydroA(t *testing.T) {
	// Hydro 10 renders as "A" in the formula tail.
	got := DeriveHydrographicsProfile(288, 6, 10)
	if got != "HA:H2O-100" {
		t.Errorf("got %q, want HA:H2O-100", got)
	}
}

func TestDeriveHydrographicsProfile_ExoticAtm_NoCandidate(t *testing.T) {
	// Atm A, hydro 5, meanK 2000 → no exotic liquid fits → empty.
	got := DeriveHydrographicsProfile(2000, 10, 5)
	if got != "" {
		t.Errorf("got %q, want empty (no liquid candidate at meanK 2000)", got)
	}
}
```

### Step 2: Verify failure

```bash
go test -run 'TestDeriveHydrographicsProfile' ./worlds/...
```

Expected: `undefined: DeriveHydrographicsProfile`.

### Step 3: Modify `worlds/hydrographics.go`

Add the `Profile` field to the existing `Hydrographics` struct and the helper function. First, find the struct definition (around line 8). Update it:

```go
// Hydrographics — surface liquid coverage per WBH p.99.
//
// Profile is populated by 3A2b-rederive (Step 5D); pre-5D consumers see
// the zero value. Format: "H<code>:<liquid>-100" (e.g., "H6:H2O-100",
// "H4:CH4-100"). Empty for vacuum or zero-hydrographics bodies.
type Hydrographics struct {
	Code    int
	Profile string // 3A2b-rederive composition tail
}
```

(Keep any other existing fields; only ADD `Profile string`.)

Append at the end of `worlds/hydrographics.go`:

```go
// hydroCodeChar renders a hydrographics code as its UWP character: 0-9 → "0".."9", 10 → "A".
func hydroCodeChar(code int) string {
	if code <= 9 {
		return fmt.Sprintf("%d", code)
	}
	return "A"
}

// DeriveHydrographicsProfile returns the composition tail for a body with
// the given mean temperature, atmosphere code, and hydrographics code.
//
// Format: "H<code>:<liquid>-100" — single-dominant-liquid form per spec.
// Empty for vacuum (atmCode=0), zero hydrographics (hydroCode=0), or
// when no exotic liquid fits at meanK.
//
// Selection:
//   - Atm 2-9, D, E with hydro > 0: liquid is "H2O" (water default)
//   - Atm A-C, F (10-12, 15) with hydro > 0: SelectExoticLiquid(meanK, atmCode)
//   - Otherwise: empty
func DeriveHydrographicsProfile(meanK float64, atmCode, hydroCode int) string {
	if hydroCode <= 0 || atmCode == 0 {
		return ""
	}
	var liquid string
	if isExoticAtmCode(atmCode) {
		liquid = SelectExoticLiquid(meanK, atmCode)
	} else {
		liquid = "H2O"
	}
	if liquid == "" {
		return ""
	}
	return "H" + hydroCodeChar(hydroCode) + ":" + liquid + "-100"
}
```

You will need `"fmt"` in `worlds/hydrographics.go`'s import block (verify present; add if not).

### Step 4: Verify

```bash
go test -run 'TestDeriveHydrographicsProfile' ./worlds/... -v
```

Expected: 6/6 PASS.

### Step 5: `just check && just test`

Expected green. The new `Profile` field has zero value across existing tests, so nothing else should break.

### Step 6: Commit

```bash
git add worlds/hydrographics.go worlds/temperature_rederive_test.go
git commit -m "feat(worlds): Hydrographics.Profile field + DeriveHydrographicsProfile (WBH p.102)"
```

---

## Task 6: RederiveAtmosphereHydrographics orchestrator core

**Files:**

- Modify: `worlds/temperature_rederive.go` (append)
- Append to: `worlds/temperature_rederive_test.go`

### Step 1: Append failing tests

Append to `worlds/temperature_rederive_test.go`:

```go
func TestRederive_TerraLike_StableInTemperate(t *testing.T) {
	// Terra-like body; rederive doesn't dramatically shift atm/hydro.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 6, Subtype: "5", Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 288}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	// Pre-rederive snapshot
	preCode := body.Atmosphere.Code
	preHydro := body.Hydrographics.Code

	// Scripted: 2D for runaway trigger (low so no fire), 2D for hydro re-roll.
	r := roller.NewScripted(2, 7)
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Atm code should NOT change for HZ body in temperate range, no runaway.
	if body.Atmosphere.Code != preCode {
		t.Errorf("atm code changed: pre=%d post=%d", preCode, body.Atmosphere.Code)
	}
	// Hydro code might shift ±1 due to re-roll; assert in plausible range.
	if abs(body.Hydrographics.Code-preHydro) > 2 {
		t.Errorf("hydro code shifted too much: pre=%d post=%d", preHydro, body.Hydrographics.Code)
	}
	// Hydrographics.Profile populated.
	if body.Hydrographics.Profile == "" {
		t.Error("Hydrographics.Profile should be populated")
	}
}

func TestRederive_NilTemperature_NoOp(t *testing.T) {
	// Body without Temperature → no-op.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Atmosphere = &Atmosphere{Code: 6}
	// body.Temperature is nil
	sys := stars.System{Primary: stars.Star{}}

	r := roller.NewScripted()
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// No mutations.
	if body.Hydrographics != nil && body.Hydrographics.Profile != "" {
		t.Error("expected no Profile when Temperature is nil")
	}
}

func TestRederive_BodyEmpty_NoOp(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyEmpty
	body.Temperature = &Temperature{MeanK: 288}
	sys := stars.System{Primary: stars.Star{}}

	r := roller.NewScripted()
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRederive_ScaleHeightUpdate(t *testing.T) {
	// Cold body: meanK=200 → ScaleHeight should be lower than at 288K.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 6, Subtype: "5", Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 200}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	r := roller.NewScripted(2, 7)
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// ScaleHeight ≈ 8.5 × 200/288 / 1.0 ≈ 5.9 km.
	wantApprox := 8.5 * 200 / 288 / 1.0
	if math.Abs(body.Atmosphere.ScaleHeight-wantApprox) > 0.5 {
		t.Errorf("ScaleHeight: got %v, want ≈%v", body.Atmosphere.ScaleHeight, wantApprox)
	}
}

// abs is a local int abs helper (no math.Abs for ints in stdlib).
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
```

(Add `"math"` to the test file's imports if not already present.)

### Step 2: Verify failure

```bash
go test -run 'TestRederive_TerraLike|TestRederive_NilTemperature|TestRederive_BodyEmpty|TestRederive_ScaleHeight' ./worlds/...
```

Expected: `undefined: RederiveAtmosphereHydrographics`.

### Step 3: Append implementation

Append to `worlds/temperature_rederive.go`:

```go
import (
	// existing — verify if not already present:
	"math"

	"wbh/roller"
	"wbh/stars"
)

// RederiveAtmosphereHydrographics re-derives 3A1's temperature-sensitive
// Atmosphere/Hydrographics fields under the body's current Temperature.MeanK.
// Mutates body in place. Called twice as part of Step 5D's 2-pass iteration.
//
// Currently mutates (Task 6 baseline; later tasks add more):
//   - Atmosphere.ScaleHeight   (re-derived from current MeanK + gravity)
//   - Hydrographics.Code       (re-rolled with current TempRange's Hot/Boiling DMs)
//   - Hydrographics.Profile    (composition tail per p.102)
//
// No-op when body.Body == BodyEmpty or body.Temperature == nil.
//
// Pending in subsequent tasks of this sub-project:
//   - Atm.Subtype + .Pressure re-roll for B/C atm (Task 7)
//   - Atm.Profile re-derive via RollGasMix for exotic atm (Task 8)
//   - CheckRunawayGreenhouse integration with hydro DM-6 override (Task 9)
func RederiveAtmosphereHydrographics(
	r roller.Roller,
	body *DetailedPlacement,
	sys stars.System,
	parent *DetailedPlacement,
) error {
	if body.Body == BodyEmpty || body.Temperature == nil {
		return nil
	}

	meanK := body.Temperature.MeanK
	tempRange := MeanKToTempRange(meanK)

	// 1. Atmosphere.ScaleHeight: re-derive from real meanK + gravity.
	if body.Atmosphere != nil && body.Physical != nil {
		body.Atmosphere.ScaleHeight = DeriveScaleHeight(meanK, body.Physical.Gravity)
	}

	// 2. Hydrographics.Code: re-roll with current TempRange's Hot/Boiling DMs.
	if body.Atmosphere != nil && body.Hydrographics != nil {
		newHydro, err := RollHydroDigit(r, body.Atmosphere.Code, body.Atmosphere.Subtype, body.SizeCode, tempRange)
		if err != nil {
			return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: hydro re-roll: %w", err)
		}
		body.Hydrographics.Code = newHydro
	}

	// 3. Hydrographics.Profile: derive from current code + atm + meanK.
	if body.Hydrographics != nil {
		atmCode := 0
		if body.Atmosphere != nil {
			atmCode = body.Atmosphere.Code
		}
		body.Hydrographics.Profile = DeriveHydrographicsProfile(meanK, atmCode, body.Hydrographics.Code)
	}

	_ = math.Abs // placeholder — math will be needed by later tasks; remove if golangci-lint flags
	_ = parent   // parent unused at Task 6; future tasks may use for moon-specific paths
	return nil
}
```

Add `"fmt"` to the import block if not present.

### Step 4: Verify

```bash
go test -run 'TestRederive_' ./worlds/... -v
```

Expected: 4/4 new tests PASS.

### Step 5: `just check && just test`

Expected green. Remove `_ = math.Abs` and `_ = parent` placeholders if they trigger lint warnings; keep parent in signature for Task 8 use.

### Step 6: Commit

```bash
git add worlds/temperature_rederive.go worlds/temperature_rederive_test.go
git commit -m "feat(worlds): RederiveAtmosphereHydrographics orchestrator core (ScaleHeight + Hydro)"
```

---

## Task 7: Add Atm Subtype + Pressure re-derivation to orchestrator

**Files:**

- Modify: `worlds/temperature_rederive.go`
- Append to: `worlds/temperature_rederive_test.go`

For atm codes B (11) and C (12), `RollCorrosiveInsidiousSubtype` is the existing 3A1 helper. The subtype DM table includes orbit/HZCO position; for the rederive, we re-roll subtype only if runaway-greenhouse fires (next task) — for now, leave subtype alone unless explicitly requested.

Wait — per Q1's 2-pass iteration, the subtype could shift between passes only if pass-1 runaway fires. So subtype re-roll is conditional on runaway. For Task 7, we add the **infrastructure** for subtype/pressure re-roll but only invoke it when runaway has fired (which hasn't been wired yet — Task 9).

This task adds the helper that performs subtype + pressure re-roll given a `runawayResult` flag. The orchestrator calls it always with `runawayResult=false` until Task 9.

### Step 1: Append failing tests

Append to `worlds/temperature_rederive_test.go`:

```go
func TestRederive_AtmosphereB_NoRunaway_NoSubtypeChange(t *testing.T) {
	// Atm B (11) without runaway → subtype stays unchanged.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Atmosphere = &Atmosphere{Code: 11, Subtype: "5", Pressure: 1.5, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 288}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	preSubtype := body.Atmosphere.Subtype
	prePressure := body.Atmosphere.Pressure

	r := roller.NewScripted(2, 7) // runaway trigger fails (mod 2+5+0 = 7, no runaway), then hydro re-roll
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	if body.Atmosphere.Subtype != preSubtype {
		t.Errorf("subtype changed without runaway: pre=%s post=%s", preSubtype, body.Atmosphere.Subtype)
	}
	if body.Atmosphere.Pressure != prePressure {
		t.Errorf("pressure changed without runaway: pre=%v post=%v", prePressure, body.Atmosphere.Pressure)
	}
}
```

### Step 2: Verify failure

```bash
go test -run 'TestRederive_AtmosphereB_NoRunaway' ./worlds/... -v
```

Expected: PASS without modification (since the orchestrator currently doesn't touch subtype/pressure). This test pins the no-change behavior so Task 9 doesn't accidentally touch subtype/pressure when runaway DOESN'T fire.

### Step 3: Add infrastructure (no functional change)

Append to `worlds/temperature_rederive.go`:

```go
// rerollAtmSubtypeAndPressure re-rolls Atmosphere.Subtype and Atmosphere.Pressure
// for codes B (11) and C (12), passing through runawayResult to the existing
// 3A1 helper. Called by Step 9 wiring (post-runaway-greenhouse) only.
//
// For atm codes other than B/C, no-op.
func rerollAtmSubtypeAndPressure(
	r roller.Roller,
	body *DetailedPlacement,
	sys stars.System,
	runawayResult bool,
) error {
	if body.Atmosphere == nil {
		return nil
	}
	code := body.Atmosphere.Code
	if code != 11 && code != 12 { // only B and C have variable subtypes
		return nil
	}

	hzco := 0.0
	if len(body.Group.Members) > 0 {
		hzco = body.Group.HZCO()
	} else {
		hzco = sys.Primary.HZCO()
	}

	isInsidious := code == 12 // C
	newSubtype, err := RollCorrosiveInsidiousSubtype(r, body.SizeCode, body.Orbit, hzco, isInsidious, runawayResult)
	if err != nil {
		return fmt.Errorf("worlds: rerollAtmSubtypeAndPressure: subtype: %w", err)
	}
	body.Atmosphere.Subtype = newSubtype

	newPressure, err := RollTotalPressure(r, code)
	if err != nil {
		return fmt.Errorf("worlds: rerollAtmSubtypeAndPressure: pressure: %w", err)
	}
	body.Atmosphere.Pressure = newPressure
	return nil
}
```

### Step 4: Verify

```bash
go test -run 'TestRederive' ./worlds/... -v
```

Expected: all rederive tests still PASS. The new helper is dormant (not called yet from orchestrator).

### Step 5: `just check && just test`

Expected green.

### Step 6: Commit

```bash
git add worlds/temperature_rederive.go worlds/temperature_rederive_test.go
git commit -m "feat(worlds): rerollAtmSubtypeAndPressure helper (dormant until Task 9)"
```

---

## Task 8: Add Atm.Profile re-derivation (RollGasMix for exotic)

**Files:**

- Modify: `worlds/temperature_rederive.go` (append helper + wire into orchestrator)
- Append to: `worlds/temperature_rederive_test.go`

### Step 1: Append failing tests

Append to `worlds/temperature_rederive_test.go`:

```go
func TestRederive_AtmProfile_ExoticC(t *testing.T) {
	// Atm C (12, insidious) with hydro 4 → RollGasMix should populate Atm.Profile.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Atmosphere = &Atmosphere{Code: 12, Subtype: "C", Pressure: 5.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 4}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 288}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	// Scripted: runaway trigger (no fire for atm 12 — already exotic, MVP skip), hydro re-roll, gas-mix rolls.
	// CheckRunawayGreenhouse skips atm 12 → consumes 0 dice.
	// Then hydro re-roll: 1 dice (2D).
	// Then RollGasMix: rolls 2D + DM per gas, up to 4 iterations, plus d10 variance per gas.
	// Approximate budget: 7 dice for gas mix (2 rolls × ~3 dice each).
	r := roller.NewScripted(7, 5, 6, 5, 6, 5, 6, 5, 6)
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Atm.Profile should have at least one gas.
	if len(body.Atmosphere.Profile.Gases) == 0 {
		t.Error("Atm.Profile.Gases should be populated for exotic atm with hydro")
	}
}

func TestRederive_AtmProfile_StandardAtm_NotMutated(t *testing.T) {
	// Atm 6 (standard) → Atm.Profile NOT touched by rederive (gas mix only for exotic).
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 6, Subtype: "5", Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 288}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	r := roller.NewScripted(2, 7)
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Atm.Profile.Gases should remain empty (zero-value) — RollGasMix skipped for standard atm.
	if len(body.Atmosphere.Profile.Gases) != 0 {
		t.Errorf("Atm.Profile.Gases should be empty for standard atm; got %d gases", len(body.Atmosphere.Profile.Gases))
	}
}
```

### Step 2: Verify failure

```bash
go test -run 'TestRederive_AtmProfile' ./worlds/... -v
```

Expected: `TestRederive_AtmProfile_ExoticC` fails (Atm.Profile.Gases empty); `TestRederive_AtmProfile_StandardAtm_NotMutated` passes (already empty by zero-value).

### Step 3: Wire RollGasMix into orchestrator

Modify `RederiveAtmosphereHydrographics` in `worlds/temperature_rederive.go`. After step 3 (Hydrographics.Profile), add step 4:

```go
	// 4. Atm.Profile: re-derive gas mix for exotic atm (A/B/C/F) with hydro > 0.
	if body.Atmosphere != nil && isExoticAtmCode(body.Atmosphere.Code) &&
		body.Hydrographics != nil && body.Hydrographics.Code > 0 {
		newProfile, err := RollGasMix(r, body.Atmosphere.Subtype, "", tempRange, body.SizeCode)
		if err != nil {
			return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: gas mix: %w", err)
		}
		body.Atmosphere.Profile = newProfile
	}
```

### Step 4: Verify

```bash
go test -run 'TestRederive_AtmProfile' ./worlds/... -v
```

Expected: both tests PASS.

### Step 5: `just check && just test`

Expected green.

### Step 6: Commit

```bash
git add worlds/temperature_rederive.go worlds/temperature_rederive_test.go
git commit -m "feat(worlds): RollGasMix re-derivation for exotic atm (WBH pp.96-98)"
```

---

## Task 9: Wire CheckRunawayGreenhouse into orchestrator + boiling hydro DM

**Files:**

- Modify: `worlds/temperature_rederive.go`
- Append to: `worlds/temperature_rederive_test.go`

When runaway greenhouse fires, the orchestrator needs to:

1. Call `CheckRunawayGreenhouse` → mutates atm.Code to A/B/C
2. If true, re-roll subtype with `runawayResult=true` (DM+4) via `rerollAtmSubtypeAndPressure`
3. Re-roll hydrographics with the **boiling DM-6** instead of the **hot DM-2** that the standard `RollHydroDigit` applies

For (3), we need to override the hydro re-roll with TempBoiling forcing. Easiest: call `RollHydroDigit` with `tempRange=TempBoiling` regardless of actual TempRange when runaway fired.

### Step 1: Append failing tests

Append to `worlds/temperature_rederive_test.go`:

```go
func TestRederive_RunawayFires_AtmAndHydroMutate(t *testing.T) {
	// Atm 6 + meanK=400 + sysAge=10 (DM+10) → easy trigger.
	// Body must be HZ — set body.HZ=true.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.HZ = true
	body.Atmosphere = &Atmosphere{Code: 6, Subtype: "5", Pressure: 1.0, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 400}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 10}}

	// Scripted dice:
	// 1. Runaway trigger: 2D=2 + age 10 + boiling DM 4 = 16 → trigger
	// 2. Atm code roll: 1D=3 → B (11)
	// 3. Subtype re-roll: 2D=8 (subtype calculation)
	// 4. Pressure re-roll: 2D=7 (pressure within new code's range)
	// 5. Hydrographics re-roll with TempBoiling: 2D=7 (will get DM-6 for boiling)
	// 6+. RollGasMix for new atm B with hydro: ~6 dice
	r := roller.NewScripted(2, 3, 8, 7, 7, 5, 6, 5, 6, 5, 6, 5, 6, 5, 6)
	err := RederiveAtmosphereHydrographics(r, body, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Atm code should be A, B, or C.
	if body.Atmosphere.Code != 10 && body.Atmosphere.Code != 11 && body.Atmosphere.Code != 12 {
		t.Errorf("atm code: got %d, want A/B/C (10/11/12)", body.Atmosphere.Code)
	}
	// Hydrographics.Code should be reduced (boiling DM-6 vs hot DM-2 = relative -4).
	// Pre-rederive hydro was 7; post-rederive should be lower than naive Hot reroll (≈ 5).
	if body.Hydrographics.Code > 5 {
		t.Errorf("hydro code: got %d, expected ≤ 5 (boiling DM applied)", body.Hydrographics.Code)
	}
}
```

### Step 2: Verify failure

```bash
go test -run 'TestRederive_RunawayFires' ./worlds/... -v
```

Expected: fails (runaway not yet wired).

### Step 3: Wire runaway check into orchestrator

Modify `RederiveAtmosphereHydrographics` in `worlds/temperature_rederive.go`. Insert the runaway check **before** the hydrographics re-roll (currently step 2):

```go
	// 1.5 (HZ-only): CheckRunawayGreenhouse — may mutate atm.Code to A/B/C.
	runawayFired := false
	if body.HZ {
		runawayFired = CheckRunawayGreenhouse(r, body, sys)
		if runawayFired {
			// Re-roll subtype + pressure with runawayResult=true (DM+4 to subtype).
			if err := rerollAtmSubtypeAndPressure(r, body, sys, true); err != nil {
				return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: post-runaway: %w", err)
			}
		}
	}

	// 2. Hydrographics.Code: re-roll with current TempRange's Hot/Boiling DMs.
	// If runaway fired, force TempBoiling so the DM-6 (boiling) applies instead of DM-2 (hot).
	hydroTempRange := tempRange
	if runawayFired {
		hydroTempRange = TempBoiling
	}
	if body.Atmosphere != nil && body.Hydrographics != nil {
		newHydro, err := RollHydroDigit(r, body.Atmosphere.Code, body.Atmosphere.Subtype, body.SizeCode, hydroTempRange)
		if err != nil {
			return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: hydro re-roll: %w", err)
		}
		body.Hydrographics.Code = newHydro
	}
```

(Replace the existing step 2's `RollHydroDigit` call to use `hydroTempRange` instead of `tempRange`.)

### Step 4: Verify

```bash
go test -run 'TestRederive' ./worlds/... -v
```

Expected: all tests PASS, including the new runaway test.

### Step 5: `just check && just test`

Expected green.

### Step 6: Commit

```bash
git add worlds/temperature_rederive.go worlds/temperature_rederive_test.go
git commit -m "feat(worlds): wire CheckRunawayGreenhouse + boiling-DM hydro override (WBH p.79)"
```

---

## Task 10: Step 5D pipeline wiring + 3A1 doc-comment cleanup

**Files:**

- Modify: `worlds/system_detail.go` (wire Step 5D)
- Modify: `worlds/atmosphere.go` (doc comment update)
- Modify: `worlds/body_physical.go` (doc comment update)
- Modify: `worlds/hydrographics.go` (doc comment update — already touched in Task 5; refine)

### Step 1: Wire Step 5D in `DetailSystem`

In `worlds/system_detail.go`, locate the end of Step 5C (the per-body Temperature pass added in 3A2b-temp Task 12). Insert Step 5D **between** Step 5C and Step 6:

```go
	// Step 5D — 3A2b-rederive pass: re-derive 3A1 atm/hydro under real T (2-pass iteration).
	for i := range detailed {
		dp := &detailed[i]
		if dp.Body == BodyEmpty || !dp.HasTemperature() {
			continue
		}
		// Pass 1: rederive using 3A2b-temp's MeanK
		if err := RederiveAtmosphereHydrographics(r, dp, sys, nil); err != nil {
			return SystemDetail{}, fmt.Errorf("worlds: rederive %s pass 1: %w", dp.Designation, err)
		}
		// Re-run temperature with corrected atm/hydro
		temp, err := GenerateTemperature(r, dp, sys, nil)
		if err != nil {
			return SystemDetail{}, fmt.Errorf("worlds: temperature %s pass 2: %w", dp.Designation, err)
		}
		dp.Temperature = temp
		// Pass 2: rederive using corrected MeanK (final)
		if err := RederiveAtmosphereHydrographics(r, dp, sys, nil); err != nil {
			return SystemDetail{}, fmt.Errorf("worlds: rederive %s pass 2: %w", dp.Designation, err)
		}

		// Same 2-pass for moons.
		for j := range dp.Moons {
			m := &dp.Moons[j]
			if !m.HasTemperature() {
				continue
			}
			moonDP := buildMoonPlacementView(m, dp)
			// Pass 1
			if err := RederiveAtmosphereHydrographics(r, moonDP, sys, dp); err != nil {
				return SystemDetail{}, fmt.Errorf("worlds: moon rederive %s pass 1: %w", m.Designation, err)
			}
			// Re-run moon temperature
			moonTemp, err := GenerateTemperature(r, moonDP, sys, dp)
			if err != nil {
				return SystemDetail{}, fmt.Errorf("worlds: moon temperature %s pass 2: %w", m.Designation, err)
			}
			m.Temperature = moonTemp
			// Pass 2
			if err := RederiveAtmosphereHydrographics(r, moonDP, sys, dp); err != nil {
				return SystemDetail{}, fmt.Errorf("worlds: moon rederive %s pass 2: %w", m.Designation, err)
			}
			// Propagate moon-side mutations: Atm/Hydro pointers should be aliased
			// from buildMoonPlacementView; verify and add explicit write-backs if not.
		}
	}
```

### Step 2: Verify Atm/Hydro pointer aliasing in buildMoonPlacementView

Read `buildMoonPlacementView` to confirm it sets `dp.Atmosphere = m.Atmosphere` and `dp.Hydrographics = m.Hydrographics` by pointer (not value-copy). Both fields ARE pointers (`*Atmosphere`, `*Hydrographics`), so the assignment is pointer-aliased automatically. Mutations through moonDP propagate to the Moon. No explicit write-back needed for these fields.

If `dp.Physical = m.Physical` is also pointer-aliased, ScaleHeight mutations propagate too.

### Step 3: Update doc comments — remove "provisional" qualifier

In `worlds/atmosphere.go`, find the `Atmosphere` struct doc comment. It currently mentions "provisional under HZCO temperature." Update to:

```go
// Atmosphere — surface atmosphere characteristics per WBH pp.79-91.
//
// Pressure, ScaleHeight, Subtype, and Profile are populated by 3A1 with
// HZCO-bucketed provisional temperature; Step 5D (3A2b-rederive) re-derives
// these fields under the real Temperature.MeanK. Post-5D values are final.
type Atmosphere struct {
	// ... existing fields
}
```

In `worlds/hydrographics.go`, refine the doc comment added in Task 5 to mention 5D explicitly:

```go
// Hydrographics — surface liquid coverage per WBH p.99.
//
// Code is populated by 3A1 with HZCO-bucketed provisional temperature; Step 5D
// (3A2b-rederive) re-derives Code under real Temperature.MeanK and populates
// Profile (composition tail). Post-5D values are final.
type Hydrographics struct {
	// ... existing fields
}
```

In `worlds/body_physical.go`, find the `BodyPhysical` doc comment. If it mentions "provisional," update similarly:

```go
// BodyPhysical — physical characteristics per WBH pp.71-78.
//
// Density, Gravity, Mass, DiameterKm are 3A1 outputs and are NOT temperature-
// sensitive; they remain stable across the 5D rederive pass.
type BodyPhysical struct {
	// ... existing fields
}
```

(Adjust each to match actual existing field lists; only the doc comment should change.)

### Step 4: Run all tests

```bash
just check && just test
```

Expected: `0 issues.`; all packages `ok`. The existing `TestZed_FullDetail_3A2b_temp` may now pass with rederive populated (Task 11 will replace it with `TestZed_FullDetail_3A2b`).

### Step 5: Commit

```bash
git add worlds/system_detail.go worlds/atmosphere.go worlds/hydrographics.go worlds/body_physical.go
git commit -m "feat(worlds): wire Step 5D 2-pass rederive into DetailSystem; refresh 3A1 doc comments"
```

---

## Task 11: TestZed_FullDetail_3A2b — composite acceptance gate

**Files:**

- Modify: `worlds/worked_examples_test.go`

**Reference:** Spec § Composite acceptance test. Replaces `TestZed_FullDetail_3A2b_temp`. Inherits all 16 prior assertions + adds 8 new for 3A2b-rederive.

### Step 1: Read existing `TestZed_FullDetail_3A2b_temp`

```bash
grep -n "TestZed_FullDetail_3A2b_temp\|TestZed_FullDetail_3A2b\b" worlds/worked_examples_test.go
```

Expected: shows `func TestZed_FullDetail_3A2b_temp` and its location.

### Step 2: Replace with `TestZed_FullDetail_3A2b`

The existing `TestZed_FullDetail_3A2b_temp` function in `worlds/worked_examples_test.go` already contains assertions 1-16 (3A1 + 3A2a + 3A2b-temp invariants) and the trailing `t.Logf` notes. The replacement is a **mechanical rename + append**:

1. Rename the function from `TestZed_FullDetail_3A2b_temp` to `TestZed_FullDetail_3A2b`.
2. Update the doc comment header to mention 3A2b-rederive.
3. Insert 8 new assertions (17-24) **after assertion 16** and **before the trailing `t.Logf` notes** at the end of the function.
4. Add a new `t.Logf` note about the 3A2b-rederive tidal-lock-deferral.

The replacement function looks like (showing only the doc comment + function signature + the new assertion block + the new t.Logf note; assertions 1-16 stay exactly as they are in the existing file):

```go
// TestZed_FullDetail_3A2b is the 3A2b acceptance gate. Replaces the
// 3A2b-temp gate; extends with property invariants for Step 5D rederive
// outputs (Atm.ScaleHeight under real T, Hydrographics.Profile, runaway
// greenhouse triggers, post-rederive consistency).
func TestZed_FullDetail_3A2b(t *testing.T) {
	t.Parallel()

	for iter := 0; iter < 100; iter++ {
		seed := int64(iter)
		r := roller.NewSeeded(seed)
		sys := composeZed()

		sp, err := worlds.GenerateSystemPlacement(r, sys)
		if err != nil {
			t.Fatalf("iter %d: GenerateSystemPlacement: %v", iter, err)
		}

		header := worlds.IISSClass23Header{
			SectorLocation:  "Storr | 0602",
			InitialSurvey:   "207-568",
			LastUpdated:     "218-1061",
			IISSDesignation: "Zed (system)",
		}

		sd, err := worlds.DetailSystem(r, sys, sp, header)
		if err != nil {
			t.Fatalf("iter %d: DetailSystem: %v", iter, err)
		}

		// === Assertions 1-16 stay as-is in the existing function ===
		// (3A1 invariants 1-3, survey check, 3A2a invariants 4-8,
		//  3A2b-temp invariants 9-16. Do NOT modify these.)

		// 3A2b-rederive invariants (new — append after assertion 16):

		// Assertion 17: ScaleHeight under real T.
		// 8.5 × meanK/288 / gravityG ± 20%.
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.Atmosphere == nil || dp.Atmosphere.Code == 0 || !dp.HasTemperature() {
				continue
			}
			gravity := 0.0
			if dp.Physical != nil {
				gravity = dp.Physical.Gravity
			}
			if gravity == 0 {
				continue
			}
			expected := 8.5 * dp.Temperature.MeanK / 288 / gravity
			if math.Abs(dp.Atmosphere.ScaleHeight-expected)/expected > 0.20 {
				t.Errorf("iter %d: body %s: ScaleHeight %v vs expected ≈ %v (>20%% drift)",
					iter, dp.Designation, dp.Atmosphere.ScaleHeight, expected)
			}
		}

		// Assertion 18: Hydrographics.Profile populated when atm + hydro present.
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.Atmosphere == nil || dp.Atmosphere.Code == 0 || dp.Hydrographics == nil || dp.Hydrographics.Code == 0 {
				continue
			}
			if dp.Hydrographics.Profile == "" {
				t.Errorf("iter %d: body %s: Hydrographics.Profile empty", iter, dp.Designation)
			}
		}

		// Assertion 19: Hydrographics.Profile format `H[0-9A]:[A-Za-z0-9]+-[0-9]+`.
		profileRegex := regexp.MustCompile(`^H[0-9A]:[A-Za-z0-9]+-[0-9]+$`)
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.Hydrographics == nil || dp.Hydrographics.Profile == "" {
				continue
			}
			if !profileRegex.MatchString(dp.Hydrographics.Profile) {
				t.Errorf("iter %d: body %s: Profile %q doesn't match expected format",
					iter, dp.Designation, dp.Hydrographics.Profile)
			}
		}

		// Assertion 20: Exotic-atm worlds use exotic liquid (or empty if no candidate).
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.Atmosphere == nil {
				continue
			}
			code := dp.Atmosphere.Code
			if code != 10 && code != 11 && code != 12 && code != 15 {
				continue
			}
			if dp.Hydrographics == nil || dp.Hydrographics.Code == 0 || dp.Hydrographics.Profile == "" {
				continue
			}
			// Profile contains an exotic liquid OR water (water has Abundance 100, can win).
			validLiquids := []string{"F2", "O2", "CH4", "C2H6", "Cl2", "NH3", "SO2", "HF", "HCN", "HCl", "H2O", "CH2O2", "CH3NO", "H2CO3", "H2SO4"}
			found := false
			for _, l := range validLiquids {
				if strings.Contains(dp.Hydrographics.Profile, ":"+l+"-") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("iter %d: body %s exotic atm %d: Profile %q lacks recognized liquid",
					iter, dp.Designation, code, dp.Hydrographics.Profile)
			}
		}

		// Assertion 21: Pressure sanity (≥ 0; non-gas-giants < 100 bar).
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.Atmosphere == nil {
				continue
			}
			if dp.Atmosphere.Pressure < 0 {
				t.Errorf("iter %d: body %s: negative Pressure %v", iter, dp.Designation, dp.Atmosphere.Pressure)
			}
			if dp.GGClass == worlds.NotGasGiant && dp.Atmosphere.Pressure > 100 {
				t.Errorf("iter %d: body %s (terrestrial): Pressure %v > 100 bar", iter, dp.Designation, dp.Atmosphere.Pressure)
			}
		}

		// Assertion 22: Post-5D Temperature non-nil (mid-pass GenerateTemperature must not have failed).
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.Body == worlds.BodyEmpty {
				continue
			}
			if !dp.HasTemperature() {
				t.Errorf("iter %d: body %s: Temperature nil after 5D", iter, dp.Designation)
			}
		}

		// Assertion 23: (Informational) Count of runaway-greenhouse-fired bodies.
		// (We can't directly observe trigger; approximation: bodies with HZ + atm A/B/C
		//  AND meanK > 303K are post-runaway candidates.)
		runawayCount := 0
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if !dp.HZ || dp.Atmosphere == nil || !dp.HasTemperature() {
				continue
			}
			c := dp.Atmosphere.Code
			if (c == 10 || c == 11 || c == 12) && dp.Temperature.MeanK > 303 {
				runawayCount++
			}
		}
		if runawayCount > 0 {
			t.Logf("iter %d: %d bodies post-runaway (HZ + atm A/B/C + meanK > 303K)", iter, runawayCount)
		}

		// Assertion 24: (Informational) Atm.Profile.Gases populated for exotic atm with hydro.
		profilePopulatedCount := 0
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.Atmosphere == nil || dp.Hydrographics == nil {
				continue
			}
			if (dp.Atmosphere.Code == 10 || dp.Atmosphere.Code == 11 || dp.Atmosphere.Code == 12 || dp.Atmosphere.Code == 15) &&
				dp.Hydrographics.Code > 0 && len(dp.Atmosphere.Profile.Gases) > 0 {
				profilePopulatedCount++
			}
		}
		if profilePopulatedCount > 0 {
			t.Logf("iter %d: %d exotic-atm bodies with populated Atm.Profile.Gases", iter, profilePopulatedCount)
		}
	}

	// Trailing informational logs (carried from prior gates):
	t.Logf("p.101 continent counts deferred to Referee fiat per 3A2a Q6 option (b)")
	t.Logf("p.106 tidal lock natural-12 verification implemented per spec; the book's worked example fudges it as a Referee narrative")
	t.Logf("p.115 sidebar Zed Prime WorstLow=230K (book) vs 219K (consistent Near/Far AU computation) — implementation uses consistent Near/Far AU")
	t.Logf("3A2b-rederive: tidal-lock re-eval if pressure crosses 2.5 bar deferred (Q5-B); requires dice-capture infrastructure")
}
```

You'll need to add `regexp` and `strings` imports if not already present.

### Step 3: Run the new test

```bash
go test -run 'TestZed_FullDetail_3A2b' ./worlds/... -v 2>&1 | tail -20
```

Expected: PASS across all 100 iterations.

### Step 4: Run full test suite

```bash
just check && just test
```

Expected: `0 issues.`; all packages `ok`.

### Step 5: Verify `TestZed_FullDetail_3A2b_temp` is gone

```bash
grep -n "func TestZed_FullDetail_3A2b_temp" worlds/worked_examples_test.go
```

Expected: no matches.

### Step 6: Commit

```bash
git add worlds/worked_examples_test.go
git commit -m "test(worlds): TestZed_FullDetail_3A2b — full 3A2b acceptance gate (WBH p.79-126)"
```

---

## Final verification (no commit)

After all 11 tasks, the branch should be ready to merge:

```bash
just check && just test
git log --oneline main..HEAD
```

Expected:

- 10 commits ahead of main (Task 1 has no commit; Tasks 2-11 each have one).
- All `ok` from test, `0 issues.` from check.

**Merge to main (after user approval):**

```bash
git checkout main
git merge --no-ff feat/wbh-world-physical-3a2b-rederive -m "Merge feat/wbh-world-physical-3a2b-rederive: World Physical 3A2b complete"
```

After merge, update memory:

- `MEMORY.md` Subprojects line: 3A2b complete with merge SHA; next is 3B (geology / biosphere / seismology).
- Note in memory if any book-inconsistency findings deserve their own feedback files.
