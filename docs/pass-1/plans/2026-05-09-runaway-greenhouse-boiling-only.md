# Runaway Greenhouse "Consider Boiling" Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend `CheckRunawayGreenhouse` to evaluate atm A (10), B (11), C (12), and F+ (15+) per WBH p.79; on trigger for those codes, leave atm.Code unchanged but signal "considered boiling" so the caller's hydro re-roll uses TempBoiling DM-6 instead of TempHot DM-2.

**Architecture:** `CheckRunawayGreenhouse` keeps its `bool` return. Eligibility expands to atm 2 and above (was atm 2-9, D, E). On trigger, the function branches: atm A/B/C/F+ return `true` without mutating code; atm 2-9, D, E continue to mutate via the existing 1D table. The caller in `RederiveAtmosphereHydrographics` snapshots the pre-trigger atm code and only invokes `rerollAtmSubtypeAndPressure` when the code actually changed.

**Tech Stack:** Go 1.26, `wbh/roller`, `task` (gofumpt + go vet + golangci-lint + modernizer), `go test -race`.

---

## File Structure

- **Modify:** `worlds/runaway_greenhouse.go` — extend eligibility filter; add post-trigger branch for atm A/B/C/F+; rewrite doc-comment to drop the MVP-simplification paragraph.
- **Modify:** `worlds/temperature_rederive.go` — caller snapshots pre-trigger atm.Code; gates `rerollAtmSubtypeAndPressure` on actual mutation.
- **Modify:** `worlds/temperature_rederive_test.go` — replace `TestCheckRunawayGreenhouse_AtmAlreadyExotic_Skipped` with new boiling-only tests; add `TestRederive_AtmosphereB_RunawayBoilingOnly` integration test.
- **Possibly modify:** `worlds/testdata/zed_markdown.golden` — refresh if seed=42's Aab IV (atm B, MeanK 313, HZ) now triggers the boiling-only path and shifts hydro.

---

### Task 1: Extend `CheckRunawayGreenhouse` + caller (atomic TDD)

**Files:**

- Modify: `worlds/temperature_rederive_test.go` (replace exotic-skipped test; add boiling-only unit tests + integration test)
- Modify: `worlds/runaway_greenhouse.go` (extend eligibility, branch on outcome, update doc-comment)
- Modify: `worlds/temperature_rederive.go` (pre/post atm.Code snapshot at caller)
- Possibly modify: `worlds/testdata/zed_markdown.golden`

The function change without the caller change leaves the caller incorrectly re-rolling subtype/pressure for atm A/B/C/F+ trigger fires. To avoid an inconsistent intermediate state, both changes land in one commit driven by tests at both levels.

- [ ] **Step 1: Write the failing unit tests for the new boiling-only paths**

Open `worlds/temperature_rederive_test.go`. Find the existing test:

```go
func TestCheckRunawayGreenhouse_AtmAlreadyExotic_Skipped(t *testing.T) {
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
```

Replace with:

```go
func TestCheckRunawayGreenhouse_AtmAlreadyExtreme_BoilingOnly(t *testing.T) {
	// WBH p.79: for atm A, B, C, F+, the runaway-greenhouse trigger still
	// evaluates, but the only effect is to consider the world boiling. The
	// atm code does NOT mutate. Each subtest scripts dice that produce a
	// trigger total ≥ 12 (atm 8, age 5, size 8 → DM age+5 + boiling+4 +
	// size+0 = +9; 2D=3 + 9 = 12 → trigger).
	cases := []struct {
		name string
		code int
	}{
		{"atm A (10)", 10},
		{"atm B (11)", 11},
		{"atm C (12)", 12},
		{"atm F (15)", 15},
		{"atm G (16)", 16},
		{"atm H (17)", 17},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body := &DetailedPlacement{}
			body.Atmosphere = &Atmosphere{Code: c.code}
			body.SizeCode = "8"
			body.Temperature = &Temperature{MeanK: 400}
			sys := stars.System{Primary: stars.Star{AgeGyr: 5}}

			// 2D=3 → 3 + age+5 (ceil 5.0) + boiling+4 = 12 → trigger.
			// No 1D code-mutation roll consumed because boiling-only.
			r := roller.NewScripted(3)
			got := CheckRunawayGreenhouse(r, body, sys)
			if !got {
				t.Errorf("expected true (trigger fired with mod=12)")
			}
			if body.Atmosphere.Code != c.code {
				t.Errorf("atm code mutated for boiling-only path: got %d, want %d (unchanged)",
					body.Atmosphere.Code, c.code)
			}
		})
	}
}
```

- [ ] **Step 2: Write the failing integration test for the caller**

In the same file, append (after the existing `TestRederive_AtmosphereB_NoRunaway_NoSubtypeChange`):

```go
func TestRederive_AtmosphereB_RunawayBoilingOnly_PreservesSubtype(t *testing.T) {
	// HZ atm-B body with MeanK > 303 (boiling) and DMs that trigger the
	// runaway-greenhouse roll. Per WBH p.79, atm B's only runaway effect
	// is the "considered boiling" hydro DM — atm.Code/Subtype/Pressure
	// must NOT change.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.HZ = true
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Atmosphere = &Atmosphere{Code: 11, Subtype: "5", Pressure: 1.5, ScaleHeight: 8.5}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Density: 1.0, Gravity: 1.0}
	body.Temperature = &Temperature{MeanK: 400} // > 303, > 388 → boiling DM+4

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0, AgeGyr: 5}}

	preCode := body.Atmosphere.Code
	preSubtype := body.Atmosphere.Subtype
	prePressure := body.Atmosphere.Pressure

	// Scripted dice budget:
	//   1 roll for runaway 2D trigger (= 3 + age+5 + boiling+4 = 12 → fires)
	//   No 1D code-mutation roll (boiling-only path).
	//   No subtype/pressure re-roll (caller skips for boiling-only).
	//   Hydrographics + gas mix re-rolls consume the rest (over-script for safety).
	r := roller.NewScripted(3, 7, 5, 8, 5, 8, 5, 8, 5, 8, 5, 8, 5, 8, 5, 8)
	if err := RederiveAtmosphereHydrographics(r, body, sys, nil); err != nil {
		t.Fatal(err)
	}

	if body.Atmosphere.Code != preCode {
		t.Errorf("atm code changed under boiling-only runaway: pre=%d post=%d",
			preCode, body.Atmosphere.Code)
	}
	if body.Atmosphere.Subtype != preSubtype {
		t.Errorf("atm subtype changed under boiling-only runaway: pre=%s post=%s",
			preSubtype, body.Atmosphere.Subtype)
	}
	if body.Atmosphere.Pressure != prePressure {
		t.Errorf("atm pressure changed under boiling-only runaway: pre=%v post=%v",
			prePressure, body.Atmosphere.Pressure)
	}
}
```

- [ ] **Step 3: Run the new tests to verify they fail**

```bash
go test ./worlds/ -run "TestCheckRunawayGreenhouse_AtmAlreadyExtreme_BoilingOnly|TestRederive_AtmosphereB_RunawayBoilingOnly_PreservesSubtype" -v
```

Expected:

- `TestCheckRunawayGreenhouse_AtmAlreadyExtreme_BoilingOnly` subtests all FAIL (function returns false for atm A/B/C/F+ today; current eligibility filter rejects them).
- `TestRederive_AtmosphereB_RunawayBoilingOnly_PreservesSubtype` PASSES — current code returns false from `CheckRunawayGreenhouse` for atm 11, so no runaway fires and subtype/pressure naturally stay. (This integration test will fail later, after the function change in Step 4 makes runaway fire but before the caller change in Step 5.)

- [ ] **Step 4: Update `CheckRunawayGreenhouse`**

Open `worlds/runaway_greenhouse.go`. Replace the entire file with:

```go
package worlds

import (
	"math"

	"wbh/roller"
	"wbh/stars"
)

// CheckRunawayGreenhouse evaluates and applies WBH p.79 Optional Runaway
// Greenhouse. Triggered when:
//   - body.Atmosphere is non-nil AND atm.Code is 2 or above
//   - body.Temperature is non-nil AND MeanK > 303K
//   - 2D + DMs ≥ 12
//
// DMs:
//   - +1 per System Age Gyr (round up)
//   - +4 if mean T ≥ 388K (boiling temperature, 12+ on basic temp table)
//   - +1 if originally tainted (codes 2, 4, 7, 9)
//   - -2 if Size 2-5
//
// On trigger, the outcome depends on the original atm code:
//
//   - atm 2-9, D (13), E (14): mutate body.Atmosphere.Code via 1D table:
//     1   → A (10)
//     2-4 → B (11)
//     5+  → C (12)
//
//   - atm A (10), B (11), C (12), F+ (15+): no mutation. WBH: "the only
//     effect of a runaway greenhouse is to consider the world to be
//     boiling." The caller treats the bool return as "consider boiling"
//     and applies hydro DM-6 instead of DM-2.
//
// Returns true iff the trigger fired (regardless of outcome path).
// Caller distinguishes the mutation vs boiling-only paths by comparing
// the pre-call atm.Code to the post-call value.
func CheckRunawayGreenhouse(r roller.Roller, body *DetailedPlacement, sys stars.System) bool {
	if body.Atmosphere == nil || body.Temperature == nil {
		return false
	}
	if body.Temperature.MeanK <= 303 {
		return false
	}
	code := body.Atmosphere.Code
	// Atm 0 (None) and 1 (Trace) are not in the WBH p.79 runaway table.
	if code < 2 {
		return false
	}

	// Trigger roll: 2D + DMs.
	dm := 0
	dm += int(math.Ceil(sys.Primary.AgeGyr))
	if body.Temperature.MeanK >= 388 {
		dm += 4
	}
	if code == 2 || code == 4 || code == 7 || code == 9 {
		dm++
	}
	si := SizeAsInt(body.SizeCode)
	if si >= 2 && si <= 5 {
		dm -= 2
	}

	roll := r.Roll("2D")
	if roll+dm < 12 {
		return false
	}

	// Trigger fired. WBH p.79: for atm A, B, C, or F+, the only effect
	// is the "consider boiling" hydro DM (handled by the caller). No
	// atm code mutation, no subtype/pressure re-roll.
	if code == 10 || code == 11 || code == 12 || code >= 15 {
		return true
	}

	// Atm 2-9, D, E: mutate code via 1D table.
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

- [ ] **Step 5: Update the caller in `temperature_rederive.go`**

Open `worlds/temperature_rederive.go`. Find the call-site block (currently around lines 69-79):

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
```

Replace with:

```go
	// 1.5 (HZ-only): CheckRunawayGreenhouse — may mutate atm.Code to A/B/C
	// for atm 2-9/D/E paths, or fire boiling-only for atm A/B/C/F+ (no
	// mutation). We compare pre/post atm.Code to know which path fired
	// and only re-roll subtype/pressure when the code was mutated. Per
	// WBH p.79, the boiling-only path's "only effect" is the hydro DM-6
	// applied below.
	runawayFired := false
	if body.HZ {
		var preCode int
		if body.Atmosphere != nil {
			preCode = body.Atmosphere.Code
		}
		runawayFired = CheckRunawayGreenhouse(r, body, sys)
		if runawayFired && body.Atmosphere != nil && body.Atmosphere.Code != preCode {
			// Code was mutated (atm 2-9/D/E → A/B/C path). Re-roll subtype
			// + pressure with runawayResult=true (DM+4 to subtype).
			if err := rerollAtmSubtypeAndPressure(r, body, sys, true); err != nil {
				return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: post-runaway: %w", err)
			}
		}
	}
```

- [ ] **Step 6: Run the new tests to verify they pass**

```bash
go test ./worlds/ -run "TestCheckRunawayGreenhouse_AtmAlreadyExtreme_BoilingOnly|TestRederive_AtmosphereB_RunawayBoilingOnly_PreservesSubtype" -v
```

Expected: PASS — all 6 subtests of `_BoilingOnly` plus the integration test.

- [ ] **Step 7: Run all `CheckRunawayGreenhouse` tests**

```bash
go test ./worlds/ -run TestCheckRunawayGreenhouse -v
```

Expected: PASS — including the existing `TestCheckRunawayGreenhouse_BelowTempThreshold`, `TestCheckRunawayGreenhouse_LowAtmCode`, `TestCheckRunawayGreenhouse_LowDiceRoll`, `TestCheckRunawayGreenhouse_Triggered_AtmA/B/C`, `TestCheckRunawayGreenhouse_TaintedDM`, `TestCheckRunawayGreenhouse_SizeDM`. The replaced `_AtmAlreadyExotic_Skipped` is now `_AtmAlreadyExtreme_BoilingOnly` and is in the run.

- [ ] **Step 8: Stage Go changes**

```bash
git add worlds/runaway_greenhouse.go worlds/temperature_rederive.go worlds/temperature_rederive_test.go
```

- [ ] **Step 9: Run the full test suite**

```bash
go test -race ./...
```

Expected: PASS, except possibly `TestRenderSystemMarkdown_ZedGolden` if seed=42's Aab IV (atm B, MeanK 313, HZ, age ≥ 5 Gyr) now triggers the boiling-only path and shifts hydrographics. If that's the only failure, proceed to Step 10. If anything else fails, escalate (BLOCKED).

- [ ] **Step 10: Refresh Zed golden (only if Step 9 flagged a golden mismatch)**

```bash
go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -update
git diff worlds/testdata/zed_markdown.golden
```

Verify the diff is limited to:

- Aab IV's Hydrographics section (Code, Coverage %, Profile) shifting because the hydro re-roll now uses TempBoiling DM-6 instead of TempHot DM-2.
- Possibly downstream RNG drift on Aab IV's Life / Habitability / Resources sections (cascading).
- Aab IV's Atmosphere section MUST NOT show subtype or pressure changes (that would indicate the caller's mutation gate failed).

Anything outside that scope is a regression — escalate (BLOCKED). If clean:

```bash
git add worlds/testdata/zed_markdown.golden
go test -race ./...
```

Confirm green.

- [ ] **Step 11: Run task quality gate**

```bash
task check
```

Expected: clean.

- [ ] **Step 12: Commit**

```bash
git commit -m "$(cat <<'EOF'
fix(worlds): runaway-greenhouse boiling-only for atm A/B/C/F+ (closes #8)

WBH p.79: any atm 2-F world in HZ with MeanK > 303 evaluates the
runaway-greenhouse 2D+DMs trigger. On fire:
- atm 2-9, D, E mutate via 1D table → A/B/C (existing path)
- atm A, B, C, F+ stay at their original code; the only effect is
  the "considered boiling" hydro DM-6 the caller applies below

CheckRunawayGreenhouse keeps its bool return; the caller in
RederiveAtmosphereHydrographics snapshots the pre-call atm.Code
and only invokes rerollAtmSubtypeAndPressure when the code was
actually mutated (atm 2-9/D/E path). Atm A/B/C/F+ trigger fires
preserve subtype and pressure per WBH's "only effect" language.

Also drops the MVP-simplification note from CheckRunawayGreenhouse's
doc-comment.

Closes #8.

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: PR + close-out

**Files:** none (operational steps only).

- [ ] **Step 1: Push the branch**

```bash
git push -u origin feat/runaway-greenhouse-boiling-only
```

- [ ] **Step 2: Open the PR**

```bash
gh pr create --repo philoserf/world-builder --title "fix(worlds): runaway-greenhouse boiling-only for atm A/B/C/F+ (closes #8)" --body "$(cat <<'EOF'
## Summary

- Extends `CheckRunawayGreenhouse` eligibility from atm 2-9/D/E to atm 2 and above (was filtering out atm A/B/C/F+).
- On trigger fire, atm A (10), B (11), C (12), and F+ (15, 16, 17) keep their atm.Code unchanged per WBH p.79 ("the only effect of a runaway greenhouse is to consider the world to be boiling"). The existing 1D mutation table only runs for atm 2-9, D, E.
- Caller in `RederiveAtmosphereHydrographics` snapshots pre-call atm.Code and only invokes `rerollAtmSubtypeAndPressure` when the code was actually mutated. Boiling-only fires preserve subtype and pressure.
- Hydro re-roll continues to use TempBoiling DM-6 whenever `runawayFired` is true (both paths), which is the WBH-prescribed effect.

Closes #8.

## Spec / plan

- Spec: `docs/pass-1/specs/2026-05-09-runaway-greenhouse-boiling-only-design.md`
- Plan: `docs/pass-1/plans/2026-05-09-runaway-greenhouse-boiling-only.md`

## Test plan

- [x] `task check` clean (gofumpt, vet, golangci-lint, modernizer)
- [x] `task test` clean with race detector
- [x] New `TestCheckRunawayGreenhouse_AtmAlreadyExtreme_BoilingOnly` — 6 subtests covering atm A/B/C/F/G/H, all asserting `true` return with code unchanged
- [x] New `TestRederive_AtmosphereB_RunawayBoilingOnly_PreservesSubtype` — full rederive integration: atm-B body with MeanK 400 triggers runaway, caller doesn't re-roll subtype/pressure
- [x] Existing `TestCheckRunawayGreenhouse_*` (BelowTempThreshold, LowAtmCode, LowDiceRoll, Triggered_AtmA/B/C, TaintedDM, SizeDM) unchanged and passing
- [x] Zed golden refreshed if seed=42's Aab IV triggers boiling-only (atm-B, HZ, MeanK 313)

## Out of scope (per spec)

- Atm 0/1 trigger (book covers atm 2-F only).
- Optional non-HZ runaway with DM-2 for Temperate worlds (WBH p.79).
- Optional 303K post-temp-determination DM+1 per 10° (WBH p.111).
EOF
)"
```

- [ ] **Step 3: Stop**

Implementation complete on the branch; PR is open. Hand back to the user for review/merge.

---

## Self-review

**Spec coverage**

- Spec § Architecture (function eligibility extension, branch on outcome): Task 1 Step 4. ✓
- Spec § Architecture (caller pre/post code check): Task 1 Step 5. ✓
- Spec § Decisions (API shape — bool return, caller compares): Task 1 Steps 4 + 5. ✓
- Spec § Decisions (atm 15+ eligibility): Task 1 Step 4 (`code >= 15` branch). ✓
- Spec § Decisions (no subtype re-roll for boiling-only): Task 1 Step 5 (caller gate). ✓
- Spec § Testing strategy unit tests (atm A/B/C/F/H boiling-only): Task 1 Step 1 (6 subtests including atm G — book says F+, includes G/H). ✓
- Spec § Testing strategy integration test: Task 1 Step 2. ✓
- Spec § Zed golden refresh: Task 1 Step 10. ✓

**Placeholder scan**

No "TBD" / "TODO" / "implement later". All steps include concrete code or commands.

**Type consistency**

- `CheckRunawayGreenhouse` signature unchanged: `func(r roller.Roller, body *DetailedPlacement, sys stars.System) bool`. Same in function definition (Step 4) and call sites (Step 5, tests).
- `body.Atmosphere.Code int` — same field type throughout.
- `rerollAtmSubtypeAndPressure(r, body, sys, runawayResult bool) error` — call signature unchanged from prior code.
