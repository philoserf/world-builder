# Tidal-Lock Spec-Divergence Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Land four PRs that resolve open issues #9, #53, #54, #55, #56 against `worlds/tidal_lock.go`, bringing the tidal-lock pipeline into faithful alignment with WBH pp.105–107.

**Architecture:** Four PRs in dependency order: (1) #55 + #56 paired (Planet→Moon eligibility + sidereal-day plumbing, both needing a two-pass walk in `ApplyRotationTilt`); (2) #54 (moon-period guard on natural-12 reroll, isolated change in `ApplyTidalLockEffect`); (3) #53 (tied-DM cascade, rewrites `SelectHighestDMCase` + orchestration); (4) #9 (full re-eval cascade — capture pre-tidal-lock snapshots, after `ApplyClimate` re-evaluate tidal lock with atmosphere DM applied, re-run `ApplyClimatePasses` for affected bodies). #9 lands last because its cascade subsumes prior fixes naturally; the other three are simpler and de-risk the work.

**Tech Stack:** Go 1.21+, `task` for build (gofumpt + vet + golangci-lint + tests), `roller.NewScripted` for fidelity tests, snapshot regen via `go test ./iiss/... -update.regression -run TestRegression`.

---

## Sequencing rationale

The handoff doc recommended deciding #9 first. We chose the **full cascade** (re-eval tidal lock AND re-run Stage 5 for affected bodies) so the implementation is the most expensive of the five issues. Landing #9 last means its snapshot regeneration naturally folds in the dice-stream shifts from #54 and #53. Each prior PR will regenerate snapshots independently; by the time #9 lands, the snapshots are at their "all-other-fixes-applied" baseline, and #9's diff isolates the atmosphere-DM behavioral delta.

#55 + #56 are paired because both want a two-pass walk in `ApplyRotationTilt` (moons before planets, so Planet→Moon eligibility can check pre-locked moons). Doing them in one PR avoids redoing the refactor.

**Cross-cutting gotcha (downstream coupling):** `temperature.go` reads `body.AxialTilt` and `body.Eccentricity`; `geology.go` reads `body.Eccentricity`. Both run inside Stage 5 (`ApplyClimatePasses`). #9's faithful cascade therefore can't just recompute tidal-lock outputs — it must also re-run `ApplyClimatePasses` for any body whose tidal-lock result changed. That's why "full cascade" is in scope. We accept that the dice stream now consumes a second tidal-lock + climate roll for affected bodies; this is deterministic per seed.

---

## File map

**New files:**

- `worlds/tidal_lock_reeval.go` — the Stage-5-post re-evaluation pass (introduced in PR 4).
- `worlds/tidal_lock_reeval_test.go` — unit tests for the re-eval pass.
- `worlds/tidal_lock_snapshot.go` — `PreTidalLockSnapshot` type + capture/restore helpers (PR 4).

**Modified files (per PR — see task lists below):**

- `worlds/tidal_lock.go` — all four PRs touch this file.
- `worlds/stage4.go` — PR 1 (two-pass walk), PR 4 (snapshot capture).
- `worlds/stage5.go` — PR 4 (add `ApplyClimatePassesForBody` or expose existing per-body entry point).
- `worlds/generate.go` — PR 4 (insert re-eval step).
- `worlds/body.go` — PR 4 (extend `TidalLock` or `DetailedPlacement` with snapshot fields).
- `worlds/tidal_lock_test.go` — every PR adds unit tests.
- `worlds/stage4_test.go` (if it exists) or new — PR 1.
- `iiss/testdata/seed_*.md` — regenerated every PR.
- `docs/wbh-inconsistencies.md` — PR 4 entry noting the two-pass re-eval pattern (faithful per p.106, not in book's worked example).

---

## PR 1 — Issues #55 + #56: Planet→Moon eligibility + sidereal day source

**Goal:** Tighten the Planet→Moon eligibility check to require (a) terrestrial planet and (b) at least one already-locked moon; plumb the correct yearHours (moon's `PeriodHours`) when applying a Planet→Moon 1:1 lock.

**Files:**

- Modify: `worlds/tidal_lock.go` (EvaluateTidalLockDMs, planetToMoonDMs, ApplyTidalLockEffect, GenerateTidalLock signatures)
- Modify: `worlds/stage4.go` (sub-stage 3 → two-pass walk)
- Add tests: `worlds/tidal_lock_test.go`

### Task 1.1: Add `isTerrestrial` predicate

- [ ] **Step 1: Write the failing test** in `worlds/tidal_lock_test.go`:

```go
func TestIsTerrestrial(t *testing.T) {
    cases := []struct {
        name string
        kind BodyKind
        want bool
    }{
        {"terrestrial true", BodyTerrestrial, true},
        {"gas giant false", BodyGasGiant, false},
        {"belt false", BodyBelt, false},
        {"empty false", BodyEmpty, false},
        {"moon false (planet→moon evaluated from planet only)", BodyMoon, false},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            b := &Body{Kind: c.kind}
            if got := isTerrestrial(b); got != c.want {
                t.Errorf("isTerrestrial(%v) = %v, want %v", c.kind, got, c.want)
            }
        })
    }
}
```

- [ ] **Step 2: Run test, expect FAIL** (`isTerrestrial undefined`):

```bash
go test ./worlds/ -run TestIsTerrestrial -v
```

- [ ] **Step 3: Implement** in `worlds/tidal_lock.go` (add near `hasSignificantMoon`):

```go
// isTerrestrial reports whether a body is eligible for the Planet→Moon
// tidal-lock case per WBH p.107 (terrestrial worlds, Size 1–F).
func isTerrestrial(body *Body) bool {
    return body.Kind == BodyTerrestrial
}
```

- [ ] **Step 4: Run test, expect PASS**.

- [ ] **Step 5: Commit**:

```bash
git add worlds/tidal_lock.go worlds/tidal_lock_test.go
git commit -m "worlds: add isTerrestrial predicate for tidal-lock Planet→Moon gate"
```

### Task 1.2: Add `hasLockedMoon` predicate (depends on moons being pre-evaluated)

- [ ] **Step 1: Write the failing test**:

```go
func TestHasLockedMoon(t *testing.T) {
    locked := &Body{TidalLock: &TidalLock{LockRatio: "1:1"}}
    threeTwo := &Body{TidalLock: &TidalLock{LockRatio: "3:2"}}
    unlocked := &Body{TidalLock: &TidalLock{LockRatio: ""}}
    noLockField := &Body{}
    cases := []struct {
        name string
        body Body
        want bool
    }{
        {"no children false", Body{}, false},
        {"all unlocked false", Body{Children: []*Body{unlocked}}, false},
        {"nil TidalLock false", Body{Children: []*Body{noLockField}}, false},
        {"3:2 locked true", Body{Children: []*Body{threeTwo}}, true},
        {"1:1 locked true", Body{Children: []*Body{locked}}, true},
        {"mixed with one locked true", Body{Children: []*Body{unlocked, locked}}, true},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            if got := hasLockedMoon(&c.body); got != c.want {
                t.Errorf("hasLockedMoon = %v, want %v", got, c.want)
            }
        })
    }
}
```

- [ ] **Step 2: Run test, expect FAIL**.

- [ ] **Step 3: Implement** in `worlds/tidal_lock.go`:

```go
// hasLockedMoon reports whether body has at least one moon already in a
// 1:1 or 3:2 tidal lock with the planet, per WBH p.107 Planet→Moon
// pre-condition. Relies on the Stage 4 two-pass ordering — moons must
// be evaluated before planets so their TidalLock fields are populated.
func hasLockedMoon(body *Body) bool {
    for _, c := range body.Children {
        if c.TidalLock == nil {
            continue
        }
        if c.TidalLock.LockRatio == "1:1" || c.TidalLock.LockRatio == "3:2" {
            return true
        }
    }
    return false
}
```

- [ ] **Step 4: Run test, expect PASS**.

- [ ] **Step 5: Commit**:

```bash
git add worlds/tidal_lock.go worlds/tidal_lock_test.go
git commit -m "worlds: add hasLockedMoon predicate for Planet→Moon eligibility"
```

### Task 1.3: Tighten `EvaluateTidalLockDMs` Planet→Moon eligibility

- [ ] **Step 1: Write the failing tests** in `worlds/tidal_lock_test.go`:

```go
func TestEvaluateTidalLockDMs_PlanetToMoon_GasGiantExcluded(t *testing.T) {
    gg := &Body{
        Kind:      BodyGasGiant,
        SizeCode:  "M",
        Eccentricity: 0,
        Children: []*Body{
            {SizeCode: "5", TidalLock: &TidalLock{LockRatio: "1:1"}},
        },
    }
    dms := EvaluateTidalLockDMs(gg, stars.System{Primary: stars.Star{Mass: 1}}, nil, nil)
    if _, ok := dms[TidalLockCasePlanetToMoon]; ok {
        t.Errorf("gas giant should not have Planet→Moon case; got DMs: %v", dms)
    }
}

func TestEvaluateTidalLockDMs_PlanetToMoon_NoLockedMoonExcluded(t *testing.T) {
    p := &Body{
        Kind:      BodyTerrestrial,
        SizeCode:  "8",
        Children: []*Body{
            {SizeCode: "5", TidalLock: &TidalLock{LockRatio: ""}}, // unlocked
        },
    }
    dms := EvaluateTidalLockDMs(p, stars.System{Primary: stars.Star{Mass: 1}}, nil, nil)
    if _, ok := dms[TidalLockCasePlanetToMoon]; ok {
        t.Errorf("planet with no locked moons should not have Planet→Moon case; got DMs: %v", dms)
    }
}

func TestEvaluateTidalLockDMs_PlanetToMoon_TerrestrialWithLockedMoonIncluded(t *testing.T) {
    p := &Body{
        Kind:     BodyTerrestrial,
        SizeCode: "8",
        Children: []*Body{
            {SizeCode: "5", OrbitPD: 30, TidalLock: &TidalLock{LockRatio: "1:1"}},
        },
    }
    dms := EvaluateTidalLockDMs(p, stars.System{Primary: stars.Star{Mass: 1}}, nil, nil)
    if _, ok := dms[TidalLockCasePlanetToMoon]; !ok {
        t.Errorf("terrestrial with locked moon should have Planet→Moon case; got DMs: %v", dms)
    }
}
```

- [ ] **Step 2: Run tests, expect first two to FAIL** (current code admits gas giants and unlocked-only).

- [ ] **Step 3: Modify** `EvaluateTidalLockDMs` at `worlds/tidal_lock.go:79`:

```go
// Planet → moon: only applies if body is a TERRESTRIAL planet with at
// least one moon already in a 1:1 or 3:2 lock per WBH p.107. The
// pre-locked-moon precondition requires moon-cases to be evaluated in
// an earlier pass; the Stage 4 orchestrator is responsible for that
// ordering.
if parentPlanet == nil && moonRef == nil &&
    isTerrestrial(body) && hasSignificantMoon(body) && hasLockedMoon(body) {
    out[TidalLockCasePlanetToMoon] = common + planetToMoonDMs(body)
}
```

- [ ] **Step 4: Run tests, expect PASS**. Also re-run the existing tidal-lock test suite:

```bash
go test ./worlds/ -run TestEvaluateTidalLockDMs -v
```

- [ ] **Step 5: Commit**:

```bash
git add worlds/tidal_lock.go worlds/tidal_lock_test.go
git commit -m "worlds(tidal-lock): tighten Planet→Moon gate to terrestrial + locked-moon (#55)"
```

### Task 1.4: Restructure Stage 4 sub-stage 3 to two-pass walk

- [ ] **Step 1: Write the failing integration test** in `worlds/stage4_test.go` (create file if absent):

```go
package worlds_test

import (
    "testing"

    "github.com/philoserf/world-builder/roller"
    "github.com/philoserf/world-builder/worlds"
)

// TestApplyRotationTilt_PlanetToMoon_RequiresPreEvaluatedMoonLocks verifies
// that the two-pass walk evaluates moons BEFORE planets, so hasLockedMoon
// returns true for planets whose moon was just locked in the first pass.
func TestApplyRotationTilt_PlanetToMoon_RequiresPreEvaluatedMoonLocks(t *testing.T) {
    // Construct a Universe where (a) the moon's DMs guarantee a 1:1 lock,
    // and (b) the planet's only Planet→Moon DM comes from that moon. With
    // the single-pass walk, the planet evaluates before the moon and sees
    // hasLockedMoon=false. With two-pass, planet sees the locked moon.
    // ... fixture omitted; engineer constructs a deterministic Universe
    // using roller.NewScripted with the per-roll values required to
    // produce moon natural-12 then planet roll consuming the Planet→Moon
    // DM path.
    // Assertion: planet.TidalLock != nil && planet.TidalLock.Case ==
    // TidalLockCasePlanetToMoon.
}
```

(Engineer must construct the fixture; the comment above states the requirement explicitly.)

- [ ] **Step 2: Run test, expect FAIL**.

- [ ] **Step 3: Modify** `worlds/stage4.go:62` (sub-stage 3 loop) to use two passes:

```go
// Sub-stage 3: tidal lock. Two-pass walk per WBH p.107 — moons must be
// evaluated before planets so the Planet→Moon case can check whether
// a parent's moons are already locked.
for body, parent := range u.AllBodiesWithParent() {
    if body.Kind == BodyEmpty || parent == nil {
        continue // skip planets in pass 1
    }
    moonRef := body
    hours := body.PeriodHours
    tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, hours)
    if err != nil {
        return fmt.Errorf("worlds: stage4 tidal lock moon %s: %w", body.Designation, err)
    }
    body.TidalLock = tl
}
for body, parent := range u.AllBodiesWithParent() {
    if body.Kind == BodyEmpty || parent != nil {
        continue // skip moons in pass 2
    }
    hours := body.Period.Hours
    tl, err := GenerateTidalLock(r, body, nil, sys, nil, hours)
    if err != nil {
        return fmt.Errorf("worlds: stage4 tidal lock %s: %w", body.Designation, err)
    }
    body.TidalLock = tl
}
```

- [ ] **Step 4: Run test, expect PASS**. Run full worlds suite:

```bash
go test ./worlds/ -race
```

If existing Zed worked-example tests break, the dice-stream order has changed for that fixture (moons consumed before planets). Inspect the test; update the `roller.NewScripted` call sequence to match the new walk order. **Do not change the asserted outputs unless the book values mandate it.**

- [ ] **Step 5: Commit**:

```bash
git add worlds/stage4.go worlds/stage4_test.go
git commit -m "worlds(stage4): two-pass tidal-lock walk — moons before planets (#55)"
```

### Task 1.5: Fix Planet→Moon sidereal day source (#56)

- [ ] **Step 1: Write the failing test**:

```go
func TestApplyTidalLockEffect_PlanetToMoon_OneToOne_UsesMoonPeriod(t *testing.T) {
    planet := &Body{
        Kind:      BodyTerrestrial,
        SizeCode:  "8",
        Period:    Period{Hours: 8766}, // 1 stellar year
        DayLength: &DayLength{SiderealHours: 24},
        Children: []*Body{
            {SizeCode: "5", OrbitPD: 30, PeriodHours: 720, // 30-day moon
                TidalLock: &TidalLock{LockRatio: "1:1"}},
        },
    }
    // Drive ApplyTidalLockEffect with FinalResult ≥ 12 and case Planet→Moon.
    // Engineer constructs the scripted roller; assertion is:
    //   planet.DayLength.SiderealHours == 720  (moon's PeriodHours)
    // NOT 8766 (planet's stellar year).
}
```

- [ ] **Step 2: Run test, expect FAIL** (currently uses 8766).

- [ ] **Step 3: Modify** `worlds/stage4.go` sub-stage 3 pass-2 to resolve the moon and pass its `PeriodHours` for the planet→moon case. The cleanest seam is to compute `yearHours` inside `GenerateTidalLock` after `SelectHighestDMCase` resolves the case:

```go
// In GenerateTidalLock, after kase is selected:
yh := yearHours
if kase == TidalLockCasePlanetToMoon {
    // Resolve closest significant moon (matches planetToMoonDMs's
    // selection). Use its PeriodHours as the day-length basis.
    closest := closestSignificantMoon(body)
    if closest != nil {
        yh = closest.PeriodHours
    }
}
initialResult := RollTidalLockStatus(r, dm)
tl, err := ApplyTidalLockEffect(r, body, moonRef, kase, initialResult, yh)
```

Add helper:

```go
func closestSignificantMoon(body *Body) *Body {
    var closest *Body
    for i := range body.Children {
        if nForSizeCode(body.Children[i].SizeCode) < 1 {
            continue
        }
        if closest == nil || body.Children[i].OrbitPD < closest.OrbitPD {
            closest = body.Children[i]
        }
    }
    return closest
}
```

Also refactor `planetToMoonDMs` to use this helper (DRY — same selection logic at `tidal_lock.go:220-228`).

- [ ] **Step 4: Run test, expect PASS**. Run full suite:

```bash
go test -race ./...
```

- [ ] **Step 5: Regenerate snapshots and inspect**:

```bash
go test ./iiss/... -update.regression -run TestRegression
git diff iiss/testdata/ | head -100
```

Confirm changes are limited to bodies actually affected by #55 + #56 (Planet→Moon outcomes). If unrelated bodies drift, investigate before committing.

- [ ] **Step 6: Commit**:

```bash
git add worlds/tidal_lock.go worlds/stage4.go worlds/tidal_lock_test.go iiss/testdata/
git commit -m "worlds(tidal-lock): use moon's PeriodHours for Planet→Moon 1:1 day length (#56)"
```

### Task 1.6: Open PR, link issues #55 + #56

- [ ] Run `task check && task test` — must pass clean.
- [ ] Push branch; `gh pr create` referencing both issues; await approval per user's git rules.

---

## PR 2 — Issue #54: Moon-period guard on natural-12 verification reroll

**Goal:** Per WBH p.105 †, when a MoonToPlanet 1:1 lock survives a natural-12 verification only because the rerolled day length would exceed the moon's orbital period, the lock holds — `FinalResult` stays at the initial value.

**Files:**

- Modify: `worlds/tidal_lock.go` (`ApplyTidalLockEffect`)
- Add tests: `worlds/tidal_lock_test.go`

### Task 2.1: Add `rerolledDayLength` helper

- [ ] **Step 1: Write the failing test**:

```go
func TestRerolledDayLength(t *testing.T) {
    body := &Body{DayLength: &DayLength{SiderealHours: 24}}
    cases := []struct {
        name        string
        result      int
        dieValue    int  // value the next 1D roll would have produced
        yearHours   float64
        want        float64
    }{
        {"result 3 → 1.5× current", 3, 0, 0, 36},
        {"result 4 → 2× current", 4, 0, 0, 48},
        {"result 5 → 3× current", 5, 0, 0, 72},
        {"result 6 → 5× current", 6, 0, 0, 120},
        {"result 7 → 1D×5×24 (dieValue=3)", 7, 3, 0, 360},
        {"result 8 → 1D×20×24 (dieValue=2)", 8, 2, 0, 960},
        {"result 9 → 1D×10×24 (dieValue=4)", 9, 4, 0, 960},
        {"result 10 → 1D×50×24 (dieValue=5)", 10, 5, 0, 6000},
        {"result 11 (3:2) → 2/3 yearHours", 11, 0, 720, 480},
        {"result 12 (1:1) → yearHours", 12, 0, 720, 720},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            r := roller.NewScripted(c.dieValue)
            if got := rerolledDayLength(r, c.result, body, c.yearHours); got != c.want {
                t.Errorf("rerolledDayLength(%d, dieValue=%d, year=%g) = %g, want %g",
                    c.result, c.dieValue, c.yearHours, got, c.want)
            }
        })
    }
}
```

**Gotcha (handoff doc):** if we compute the would-be day length using the same dice that `ApplyTidalLockEffect` would consume for results 7–10, then take the "keep 1:1" branch, the dice stream is shorter than if we'd applied the reroll. The handoff calls this "correct behavior" — the book says no further effect — and accepts the dice-stream change. Tests below match this interpretation.

- [ ] **Step 2: Run test, expect FAIL** (`rerolledDayLength undefined`).

- [ ] **Step 3: Implement** in `worlds/tidal_lock.go`:

```go
// rerolledDayLength returns the day length that ApplyTidalLockEffect
// WOULD set for the given verification-reroll result, without
// committing the effect. Used by the WBH p.105 † moon-period guard:
// for MoonToPlanet, if the rerolled day length exceeds the moon's
// orbital period (yearHours), the 1:1 lock holds.
//
// Consumes one 1D roll for results 7–10 (matching ApplyTidalLockEffect).
// Caller must use the same Roller in a way that preserves total dice
// consumption when the guard fires; see ApplyTidalLockEffect's caller.
func rerolledDayLength(r roller.Roller, result int, body *Body, yearHours float64) float64 {
    current := 0.0
    if body.DayLength != nil {
        current = body.DayLength.SiderealHours
    }
    switch result {
    case 3:
        return current * 1.5
    case 4:
        return current * 2
    case 5:
        return current * 3
    case 6:
        return current * 5
    case 7:
        return float64(r.Roll("1D") * 5 * 24)
    case 8:
        return float64(r.Roll("1D") * 20 * 24)
    case 9:
        return float64(r.Roll("1D") * 10 * 24)
    case 10:
        return float64(r.Roll("1D") * 50 * 24)
    case 11:
        return yearHours * 2 / 3
    case 12, 13, 14, 15, 16:
        return yearHours
    default:
        return current
    }
}
```

- [ ] **Step 4: Run test, expect PASS**.

- [ ] **Step 5: Commit**:

```bash
git add worlds/tidal_lock.go worlds/tidal_lock_test.go
git commit -m "worlds: add rerolledDayLength helper for moon-period guard (#54 prep)"
```

### Task 2.2: Apply moon-period guard in `ApplyTidalLockEffect`

- [ ] **Step 1: Write the failing test**:

```go
func TestApplyTidalLockEffect_MoonPeriodGuard_KeepsOneToOneLock(t *testing.T) {
    // Setup: MoonToPlanet case, InitialResult = 12 (locks), natural-12
    // verification fires, rerolled result would be 8 (1D×20×24).
    // Moon's PeriodHours = 480 (a 20-day moon). Rerolled day length:
    //   1D=3 → 3×20×24 = 1440 hours. 1440 > 480 → guard fires; lock holds.
    moon := &Body{
        Kind:         BodyMoon,
        SizeCode:     "5",
        DayLength:    &DayLength{SiderealHours: 24},
        PeriodHours:  480,
    }
    // Scripted rolls: verification = 12 (2D=12 means dice=6,6)
    //                 reroll status: result 8 → 2D + 0 DM = 8 (dice = 4,4)
    //                 rerolledDayLength inspection: 1D = 3
    r := roller.NewScripted(12, 8, 3)
    tl, err := ApplyTidalLockEffect(r, moon, moon, TidalLockCaseMoonToPlanet, 12, 480)
    if err != nil {
        t.Fatalf("ApplyTidalLockEffect: %v", err)
    }
    if tl.FinalResult != 12 {
        t.Errorf("FinalResult = %d, want 12 (lock held by guard)", tl.FinalResult)
    }
    if tl.LockRatio != "1:1" {
        t.Errorf("LockRatio = %q, want 1:1", tl.LockRatio)
    }
}

func TestApplyTidalLockEffect_MoonPeriodGuard_NotApplicableToPlanetToStar(t *testing.T) {
    // Same scenario but case = PlanetToStar: guard MUST NOT fire.
    // FinalResult should reflect the rerolled value, not the initial.
    planet := &Body{Kind: BodyTerrestrial, SizeCode: "8",
        DayLength: &DayLength{SiderealHours: 24}, Period: Period{Hours: 8766}}
    r := roller.NewScripted(12, 8, 3)
    tl, err := ApplyTidalLockEffect(r, planet, nil, TidalLockCasePlanetToStar, 12, 8766)
    if err != nil { t.Fatalf("%v", err) }
    if tl.FinalResult != 8 {
        t.Errorf("FinalResult = %d, want 8 (no guard for PlanetToStar)", tl.FinalResult)
    }
}
```

- [ ] **Step 2: Run test, expect FAIL** (guard not implemented).

- [ ] **Step 3: Modify** `ApplyTidalLockEffect` natural-12 branch at `worlds/tidal_lock.go:382`:

```go
if initialResult >= 12 {
    verification := r.Roll("2D")
    if verification == 12 {
        tl.VerificationFired = true
        rerolled := RollTidalLockStatus(r, 0)
        // WBH p.105 † moon-period guard: for MoonToPlanet only, if the
        // rerolled day length would exceed the moon's orbital period,
        // the 1:1 lock holds. yearHours for a moon is its PeriodHours.
        if kase == TidalLockCaseMoonToPlanet &&
            rerolledDayLength(r, rerolled, body, yearHours) > yearHours {
            // Keep FinalResult = initialResult (1:1 lock held).
        } else {
            tl.FinalResult = rerolled
        }
    }
}
```

**Note:** `rerolledDayLength` may consume a 1D roll (for results 7–10). When the guard fires and we discard the rerolled effect, that 1D roll is consumed but its effect is not applied — accepted per Task 2.1 gotcha.

- [ ] **Step 4: Run tests, expect PASS**. Run full suite:

```bash
go test ./worlds/ -race
go test ./worlds/ -run TestZed
```

- [ ] **Step 5: Regenerate snapshots, inspect, commit**:

```bash
go test ./iiss/... -update.regression -run TestRegression
git diff iiss/testdata/ | head -200
git add worlds/tidal_lock.go worlds/tidal_lock_test.go iiss/testdata/
git commit -m "worlds(tidal-lock): add moon-period guard on natural-12 reroll (#54)"
```

### Task 2.3: Verify sanity-sweep matches the expected direction

Per the handoff doc, #54 produces "more locks" (some MoonToPlanet 1:1 outcomes that previously broke now hold).

- [ ] Run the 10-seed sweep from [`tidal-lock-handoff.md`](tidal-lock-handoff.md):

```bash
for seed in 1 2 3 4 5 42 100 200 500 1000; do
  go run ./cmd/world-builder -seed $seed -format markdown | grep -c "→ "
done
```

Capture pre-PR counts (checkout `main` first), compare against the patch. **Lock count should not decrease for any seed.** If it does, the guard is mis-firing.

### Task 2.4: Open PR

- [ ] `task check && task test` clean → push → `gh pr create --label bug` referencing #54.

---

## PR 3 — Issue #53: Tied-DM cascade

**Goal:** Per WBH p.106, when multiple cases tie on the highest DM, roll all of them (moon-case first), and apply the highest adjusted result. Currently `SelectHighestDMCase` returns one case and the others are silently dropped.

**Decision recorded:** "Roll all tied cases, take highest adjusted result" per handoff §#53 recommendation. The book's "stop at first lock" phrasing is interpreted as: the ordering matters for dice consumption (moon case rolls first), but all tied cases roll.

**Files:**

- Modify: `worlds/tidal_lock.go` (`SelectHighestDMCase` signature + `GenerateTidalLock` orchestration)
- Modify: existing callers of `SelectHighestDMCase` if any (grep first)
- Add tests: `worlds/tidal_lock_test.go`

### Task 3.1: Change `SelectHighestDMCase` to return tied cases in order

- [ ] **Step 1: Write the failing test**:

```go
func TestSelectHighestDMCases_OrderedTiedCases(t *testing.T) {
    // PlanetToStar tied with PlanetToMoon at DM=5; moon case absent.
    // Expected return order: MoonToPlanet (absent), PlanetToMoon, PlanetToStar.
    dms := map[TidalLockCase]int{
        TidalLockCasePlanetToStar: 5,
        TidalLockCasePlanetToMoon: 5,
    }
    cases, dm := SelectHighestDMCases(dms)
    if dm != 5 {
        t.Errorf("dm = %d, want 5", dm)
    }
    want := []TidalLockCase{TidalLockCasePlanetToMoon, TidalLockCasePlanetToStar}
    if !slices.Equal(cases, want) {
        t.Errorf("cases = %v, want %v", cases, want)
    }
}

func TestSelectHighestDMCases_SingleCase(t *testing.T) {
    dms := map[TidalLockCase]int{TidalLockCaseMoonToPlanet: 8}
    cases, dm := SelectHighestDMCases(dms)
    if dm != 8 || len(cases) != 1 || cases[0] != TidalLockCaseMoonToPlanet {
        t.Errorf("got (%v, %d), want ([MoonToPlanet], 8)", cases, dm)
    }
}

func TestSelectHighestDMCases_AllFiltered(t *testing.T) {
    dms := map[TidalLockCase]int{TidalLockCasePlanetToStar: -11}
    cases, _ := SelectHighestDMCases(dms)
    if len(cases) != 0 {
        t.Errorf("expected no cases, got %v", cases)
    }
}
```

- [ ] **Step 2: Run test, expect FAIL** (`SelectHighestDMCases undefined`).

- [ ] **Step 3: Implement** (alongside the existing `SelectHighestDMCase`; keep the singular for callers we don't change):

```go
// SelectHighestDMCases returns all cases tied at the highest DM, in
// p.106 priority order (MoonToPlanet, PlanetToMoon, PlanetToStar).
// Cases with DM ≤ -10 are filtered. Returns ([], 0) if no case applies.
func SelectHighestDMCases(dms map[TidalLockCase]int) ([]TidalLockCase, int) {
    priority := []TidalLockCase{
        TidalLockCaseMoonToPlanet,
        TidalLockCasePlanetToMoon,
        TidalLockCasePlanetToStar,
    }
    bestDM := -10
    for _, kase := range priority {
        if dm, ok := dms[kase]; ok && dm > bestDM {
            bestDM = dm
        }
    }
    if bestDM == -10 {
        return nil, 0
    }
    var tied []TidalLockCase
    for _, kase := range priority {
        if dm, ok := dms[kase]; ok && dm == bestDM {
            tied = append(tied, kase)
        }
    }
    return tied, bestDM
}
```

- [ ] **Step 4: Run tests, expect PASS**.

- [ ] **Step 5: Commit**:

```bash
git add worlds/tidal_lock.go worlds/tidal_lock_test.go
git commit -m "worlds(tidal-lock): add SelectHighestDMCases for tied-DM cascade (#53 prep)"
```

### Task 3.2: Rewrite `GenerateTidalLock` orchestration to roll all tied cases

- [ ] **Step 1: Write the failing test**:

```go
func TestGenerateTidalLock_TiedCases_RollsAllTakesHighest(t *testing.T) {
    // Construct DMs where Planet→Star and Planet→Moon tie at 4.
    // Scripted rolls: moon case absent; Planet→Moon rolls 2D=5 → adj 9;
    // Planet→Star rolls 2D=7 → adj 11. Highest adjusted = 11, case = PlanetToStar.
    // Verify: tl.Case == PlanetToStar, tl.InitialResult == 11.
    // Also verify dice consumption: 2 separate 2D rolls (one per tied case).
}
```

(Engineer constructs fixture using `roller.NewScripted` with dice values that produce a tie at DM evaluation, then divergent post-roll adjusted results.)

- [ ] **Step 2: Run test, expect FAIL**.

- [ ] **Step 3: Modify** `GenerateTidalLock` at `worlds/tidal_lock.go:489`:

```go
func GenerateTidalLock(
    r roller.Roller,
    body *Body,
    moonRef *Body,
    sys stars.System,
    parentPlanet *Body,
    yearHours float64,
) (*TidalLock, error) {
    if body.Kind == BodyEmpty {
        return nil, nil
    }

    dms := EvaluateTidalLockDMs(body, sys, parentPlanet, moonRef)
    tiedCases, dm := SelectHighestDMCases(dms)
    if len(tiedCases) == 0 {
        return nil, nil
    }

    // Roll all tied cases in priority order (moon-cases first per p.106);
    // take the highest adjusted result. Each case consumes its own 2D.
    bestResult := -1 << 30
    var bestCase TidalLockCase
    for _, kase := range tiedCases {
        rolled := RollTidalLockStatus(r, dm)
        if rolled > bestResult {
            bestResult = rolled
            bestCase = kase
        }
    }

    // Per #56, resolve yearHours after the case is known.
    yh := yearHours
    if bestCase == TidalLockCasePlanetToMoon {
        if closest := closestSignificantMoon(body); closest != nil {
            yh = closest.PeriodHours
        }
    }

    tl, err := ApplyTidalLockEffect(r, body, moonRef, bestCase, bestResult, yh)
    if err != nil {
        return nil, fmt.Errorf("worlds: GenerateTidalLock: %w", err)
    }
    return &tl, nil
}
```

**Gotcha:** the old `SelectHighestDMCase` is still referenced by the existing test `TestSelectHighestDMCase_*`. Either keep both functions (singular returns just the first tied case for backwards compatibility) or update the existing tests. **Preferred:** keep both, since the singular is used by tests and may be called by external code; mark singular as `// Deprecated: use SelectHighestDMCases for full tied-case handling.`

- [ ] **Step 4: Run test, expect PASS**. Run full suite + Zed:

```bash
go test -race ./worlds/...
go test ./worlds/ -run TestZed
```

- [ ] **Step 5: Regen snapshots, inspect, commit**:

```bash
go test ./iiss/... -update.regression -run TestRegression
git diff iiss/testdata/ | wc -l   # expect substantial drift
git add worlds/tidal_lock.go worlds/tidal_lock_test.go iiss/testdata/
git commit -m "worlds(tidal-lock): cascade tied DM cases per WBH p.106 (#53)"
```

### Task 3.3: 10-seed sweep direction check

Handoff predicts "more locks" from #53 (some bodies that previously rolled only the lower-priority tied case now hit a lock via the higher-priority case).

- [ ] Run sweep; lock count should not decrease.

### Task 3.4: Open PR — reference #53.

---

## PR 4 — Issue #9: Full atmosphere-DM re-eval cascade

**Goal:** Capture pre-tidal-lock body state during Stage 4; after `ApplyClimate` (Stage 5) sets atmospheric pressure, identify bodies with pressure > 2.5 bar where the atmosphere DM would change the tidal-lock outcome; for those bodies, restore pre-lock state, re-run tidal lock with the atmosphere DM, and re-run `ApplyClimatePasses`.

**Files:**

- Create: `worlds/tidal_lock_reeval.go`, `worlds/tidal_lock_snapshot.go`, `worlds/tidal_lock_reeval_test.go`
- Modify: `worlds/stage4.go` (capture snapshots), `worlds/stage5.go` (expose per-body re-run), `worlds/generate.go` (insert re-eval step), `worlds/body.go` (snapshot field)
- Modify: `docs/wbh-inconsistencies.md` (add re-eval note)

### Task 4.1: Define `PreTidalLockSnapshot` type

- [ ] **Step 1: Write the failing test** in `worlds/tidal_lock_reeval_test.go`:

```go
func TestPreTidalLockSnapshot_RoundTrip(t *testing.T) {
    body := &Body{
        Eccentricity: 0.42,
        AxialTilt:    &AxialTilt{Degrees: 27, Retrograde: false},
        DayLength:    &DayLength{SiderealHours: 24, SolarHours: 24, YearDays: 365},
    }
    snap := CapturePreTidalLockSnapshot(body)
    body.Eccentricity = 0.01
    body.AxialTilt.Degrees = 90
    body.AxialTilt.Retrograde = true
    body.DayLength.SiderealHours = 8766

    snap.RestoreInto(body)
    if body.Eccentricity != 0.42 {
        t.Errorf("Eccentricity not restored: %g", body.Eccentricity)
    }
    if body.AxialTilt.Degrees != 27 || body.AxialTilt.Retrograde {
        t.Errorf("AxialTilt not restored: %+v", body.AxialTilt)
    }
    if body.DayLength.SiderealHours != 24 {
        t.Errorf("DayLength not restored: %+v", body.DayLength)
    }
}

func TestPreTidalLockSnapshot_NilFields(t *testing.T) {
    body := &Body{Eccentricity: 0.1} // nil DayLength, nil AxialTilt
    snap := CapturePreTidalLockSnapshot(body)
    snap.RestoreInto(body) // must not panic
}
```

- [ ] **Step 2: Run test, expect FAIL** (`CapturePreTidalLockSnapshot undefined`).

- [ ] **Step 3: Implement** in `worlds/tidal_lock_snapshot.go`:

```go
package worlds

// PreTidalLockSnapshot captures the body fields that ApplyTidalLockEffect
// can mutate, for use by the WBH p.106 atmosphere-DM re-evaluation
// cascade. The Roller cannot be rewound, so the re-eval pass consumes
// fresh dice; this snapshot restores the body state that those fresh
// rolls will then operate on.
type PreTidalLockSnapshot struct {
    Eccentricity float64
    AxialTilt    *AxialTilt // value copy; nil if body had none
    DayLength    *DayLength // value copy; nil if body had none
}

// CapturePreTidalLockSnapshot snapshots the mutable fields of body
// immediately before ApplyTidalLockEffect runs.
func CapturePreTidalLockSnapshot(body *Body) PreTidalLockSnapshot {
    snap := PreTidalLockSnapshot{Eccentricity: body.Eccentricity}
    if body.AxialTilt != nil {
        v := *body.AxialTilt
        snap.AxialTilt = &v
    }
    if body.DayLength != nil {
        v := *body.DayLength
        snap.DayLength = &v
    }
    return snap
}

// RestoreInto writes the snapshot back into body, replacing the current
// values. Nil fields in the snapshot restore the body field to nil.
func (s PreTidalLockSnapshot) RestoreInto(body *Body) {
    body.Eccentricity = s.Eccentricity
    if s.AxialTilt != nil {
        v := *s.AxialTilt
        body.AxialTilt = &v
    } else {
        body.AxialTilt = nil
    }
    if s.DayLength != nil {
        v := *s.DayLength
        body.DayLength = &v
    } else {
        body.DayLength = nil
    }
}
```

- [ ] **Step 4: Run test, expect PASS**.

- [ ] **Step 5: Commit**:

```bash
git add worlds/tidal_lock_snapshot.go worlds/tidal_lock_reeval_test.go
git commit -m "worlds(tidal-lock): add PreTidalLockSnapshot type (#9 prep)"
```

### Task 4.2: Add snapshot storage on `Body`

The snapshot needs to survive from Stage 4 → Stage 5 → re-eval. Easiest: a transient field on `Body`.

- [ ] **Step 1: Modify** `worlds/body.go` near the existing Stage-4 fields:

```go
// Tidal-lock pre-eval snapshot — set during Stage 4 just before
// ApplyTidalLockEffect, consumed by ApplyTidalLockReEval after Stage 5.
// Cleared to zero value once re-eval runs. Not part of the rendered
// output (IISS forms ignore it).
preTidalLockSnapshot *PreTidalLockSnapshot
```

Add accessors only if external packages need them (they shouldn't — keep it package-private).

- [ ] **Step 2: Commit** the type change alone (so the next task's diff is clean):

```bash
git add worlds/body.go
git commit -m "worlds: add preTidalLockSnapshot field on Body (#9 prep)"
```

### Task 4.3: Capture snapshot during Stage 4 sub-stage 3

- [ ] **Step 1: Write a test** verifying that after `ApplyRotationTilt`, every body that ran tidal lock has its snapshot stored:

```go
func TestApplyRotationTilt_CapturesPreTidalLockSnapshot(t *testing.T) {
    // Construct a Universe with a body that will run tidal lock; verify
    // u.bodyByDesignation("...").preTidalLockSnapshot != nil after Stage 4.
}
```

(Engineer: use a real seeded Universe or hand-build a minimal one.)

- [ ] **Step 2: Run test, expect FAIL**.

- [ ] **Step 3: Modify** `worlds/tidal_lock.go` `GenerateTidalLock` to capture the snapshot before calling `ApplyTidalLockEffect`:

```go
// Capture pre-effect state for the WBH p.106 atmosphere-DM re-eval
// cascade. Stored on body so the post-Stage-5 pass can restore.
snap := CapturePreTidalLockSnapshot(body)
body.preTidalLockSnapshot = &snap

tl, err := ApplyTidalLockEffect(r, body, moonRef, bestCase, bestResult, yh)
```

Also store the initial DMs map on TidalLock for re-eval comparison:

```go
type TidalLock struct {
    // ... existing fields ...

    // PreEvalDMs captures the DM map computed at Stage 4 (before
    // atmosphere was known). Used by ApplyTidalLockReEval to compare
    // against the post-atmosphere DM map.
    PreEvalDMs map[TidalLockCase]int
}
```

Set `tl.PreEvalDMs = dms` before returning.

- [ ] **Step 4: Run test, expect PASS**. Run worlds suite.

- [ ] **Step 5: Commit**:

```bash
git add worlds/tidal_lock.go worlds/tidal_lock_test.go
git commit -m "worlds(tidal-lock): capture pre-eval snapshot and DM map (#9 prep)"
```

### Task 4.4: Expose per-body re-run of Stage 5

`ApplyClimate` iterates the system; for re-eval, we need a per-body entry point. `ApplyClimatePasses` already exists (`worlds/stage5.go:128`). Verify it's safe to call standalone for a body whose atmosphere/hydrographics were already populated. If not, add a clearing helper.

- [ ] **Step 1: Read** `worlds/stage5.go:128` (`ApplyClimatePasses`) and `worlds/body.go` to identify the fields populated by Stage 5 (Atmosphere, Hydrographics, Temperature, Geology populated inside ApplyClimatePasses).

- [ ] **Step 2: Write the failing test**:

```go
func TestApplyClimatePasses_ReentryClearsAndRebuilds(t *testing.T) {
    // Build a body, run ApplyClimatePasses, mutate AxialTilt, clear
    // climate fields via a new helper ClearStage5Output, re-run
    // ApplyClimatePasses; verify Atmosphere/Hydrographics/Temperature
    // were repopulated with values derived from the new AxialTilt.
}
```

- [ ] **Step 3: Implement** a `ClearStage5Output(body *Body)` helper in `worlds/tidal_lock_reeval.go`:

```go
// ClearStage5Output zeroes the per-body fields populated by
// ApplyClimatePasses (Atmosphere, Hydrographics, Temperature, Geology,
// and any taint/typology fields). Called before re-running Stage 5 for
// a body whose tidal-lock outputs changed during atmosphere-DM re-eval.
func ClearStage5Output(body *Body) {
    body.Atmosphere = nil
    body.Hydrographics = nil
    body.Temperature = nil
    body.Geology = nil
    body.Taint = nil
    // ... fill in any other Stage 5-populated fields by reading body.go
}
```

(Engineer: enumerate the Stage-5-populated fields by reading `body.go` and `stage5.go`. Don't guess — check the actual writes.)

- [ ] **Step 4: Run test, expect PASS**. Verify with `go vet`.

- [ ] **Step 5: Commit**:

```bash
git add worlds/tidal_lock_reeval.go worlds/tidal_lock_reeval_test.go
git commit -m "worlds: add ClearStage5Output for tidal-lock re-eval cascade (#9 prep)"
```

### Task 4.5: Implement `ApplyTidalLockReEval`

This is the heart of #9.

- [ ] **Step 1: Write the failing integration test**:

```go
func TestApplyTidalLockReEval_PressureCrosses2_5_TriggersReEval(t *testing.T) {
    // Build a Universe where a body has atmosphere pressure > 2.5 bar
    // AND the atmosphere -2 DM would push its tidal-lock DM into
    // genuinely-different territory (e.g., from 4 to 2 — no, that's
    // still the same case but with a lower roll target). Construct a
    // case where the new DM map changes the SELECTED case (e.g., from
    // PlanetToStar to TidalLockCaseNone, or where the new initial
    // result + verification path produces a different LockRatio).
    //
    // Run: ApplyRotationTilt → ApplyClimate → ApplyTidalLockReEval.
    // Assert: body.TidalLock reflects the post-re-eval state and
    // body.Atmosphere has been recomputed.
}
```

(Engineer: this fixture is the largest part of the work. Use `roller.NewScripted` with dice values that produce the scenario. Comments in the test should cite expected dice values at every step.)

- [ ] **Step 2: Run test, expect FAIL** (`ApplyTidalLockReEval undefined`).

- [ ] **Step 3: Implement** in `worlds/tidal_lock_reeval.go`:

```go
// ApplyTidalLockReEval performs the WBH p.106 atmosphere-DM re-evaluation
// cascade. After Stage 5 has set body.Atmosphere, this pass:
//   1. identifies bodies where atmospheric pressure > 2.5 bar AND the
//      stored pre-eval DM map would change with the -2 atmosphere DM
//      applied (the case selection or roll target changes meaningfully).
//   2. for each affected body, restores the pre-tidal-lock snapshot,
//      clears Stage-5 outputs, re-runs GenerateTidalLock (now with
//      atmosphere DM in commonTidalLockDMs), and re-runs
//      ApplyClimatePasses to regenerate atmosphere/temperature/etc.
//
// This is a Stage 5+ step; insert it in the Generate pipeline immediately
// after ApplyClimate.
func ApplyTidalLockReEval(r roller.Roller, u *Universe) error {
    for body, parent := range u.AllBodiesWithParent() {
        if body.Kind == BodyEmpty {
            continue
        }
        if body.Atmosphere == nil || body.Atmosphere.Pressure <= 2.5 {
            continue
        }
        if body.preTidalLockSnapshot == nil {
            continue // body never ran tidal lock
        }
        if !atmosphereDMWouldChangeOutcome(body, u.System, parent) {
            continue
        }

        body.preTidalLockSnapshot.RestoreInto(body)
        body.TidalLock = nil
        ClearStage5Output(body)

        // Re-run tidal lock — commonTidalLockDMs now sees body.Atmosphere
        // because we did NOT clear it... wait, we did. We need atmosphere
        // pressure for the DM. Solution: keep Atmosphere set but clear
        // Hydrographics/Temperature/Geology only. See note below.
        // ... [REVIEW POINT — this needs the engineer's careful
        // attention; the DM calculation depends on body.Atmosphere
        // being non-nil with the pre-re-eval pressure. We can either
        // (a) skip ClearStage5Output for Atmosphere until after the
        // re-eval, or (b) compute the DM ourselves with the captured
        // pressure value.]

        var moonRef *Body
        hours := body.Period.Hours
        if parent != nil {
            moonRef = body
            hours = body.PeriodHours
        }
        tl, err := GenerateTidalLock(r, body, moonRef, u.System, parent, hours)
        if err != nil {
            return fmt.Errorf("worlds: tidal-lock re-eval %s: %w", body.Designation, err)
        }
        body.TidalLock = tl

        // Re-run Stage 5 for this body only.
        if err := ApplyClimatePasses(r, body, u.System); err != nil {
            return fmt.Errorf("worlds: tidal-lock re-eval climate %s: %w", body.Designation, err)
        }
    }
    return nil
}

// atmosphereDMWouldChangeOutcome checks whether applying the -2
// atmospheric pressure DM to the stored pre-eval DM map would change
// either the selected case or the rolled target.
func atmosphereDMWouldChangeOutcome(body *Body, sys stars.System, parent *Body) bool {
    if body.TidalLock == nil || body.TidalLock.PreEvalDMs == nil {
        return false
    }
    // Re-evaluate commonTidalLockDMs with current body.Atmosphere; compare
    // case selection.
    var moonRef *Body
    if parent != nil {
        moonRef = body
    }
    newDMs := EvaluateTidalLockDMs(body, sys, parent, moonRef)
    newCases, _ := SelectHighestDMCases(newDMs)
    oldCases, _ := SelectHighestDMCases(body.TidalLock.PreEvalDMs)
    if !slices.Equal(newCases, oldCases) {
        return true
    }
    // Even if the case is the same, a 2-point DM shift could push the
    // result across the 3/7/11/12 thresholds. Always re-eval when
    // pressure > 2.5 — accept the false-positive cost.
    return true
}
```

**Review point for engineer:** the comment block in `ApplyTidalLockReEval` flags an ordering issue around `ClearStage5Output`. Resolve before commit. Recommended resolution: snapshot `body.Atmosphere.Pressure` into a local before clearing, and either (a) pass it explicitly into the re-eval, or (b) leave `body.Atmosphere` non-nil during the re-run and let `commonTidalLockDMs` consume it. Option (b) is simpler — clear Atmosphere AFTER tidal lock re-eval, not before.

- [ ] **Step 4: Run test, expect PASS**. Full suite + Zed:

```bash
go test -race ./...
go test ./worlds/ -run TestZed
```

- [ ] **Step 5: Commit** the re-eval implementation:

```bash
git add worlds/tidal_lock_reeval.go worlds/tidal_lock_reeval_test.go
git commit -m "worlds: add ApplyTidalLockReEval cascade for atmosphere DM (#9)"
```

### Task 4.6: Wire `ApplyTidalLockReEval` into the pipeline

- [ ] **Step 1: Modify** `worlds/generate.go:46` to insert the re-eval after `ApplyClimate`:

```go
for _, step := range []func(roller.Roller, *Universe) error{
    ApplyDetailFrontEnd,
    ApplyBodyPhysical,
    ApplyBeltDetails,
    ApplyMoonRefinement,
    ApplyRotationTilt,
    ApplyClimate,
    ApplyTidalLockReEval, // ← new
    ApplyTaintTypology,
    ApplySurfaceDistribution,
    ApplyGeology,
    ApplyBiology,
} {
```

- [ ] **Step 2: Run full suite + Zed worked examples**. Several Zed tests will likely break because their scripted dice sequences pre-date the re-eval pass:

```bash
go test ./worlds/ -run TestZed -v
```

For each broken Zed test: confirm Zed's body does NOT have pressure > 2.5 bar (it shouldn't — Zed is a near-Terra environment). If a Zed test breaks, the re-eval is firing when it shouldn't; check the gate condition. The Zed tests are the proof-of-fidelity safety net — they must pass.

- [ ] **Step 3: Regenerate snapshots** and inspect:

```bash
go test ./iiss/... -update.regression -run TestRegression
git diff iiss/testdata/ | wc -l
```

Expect substantial drift — this is the largest of the four PRs.

- [ ] **Step 4: 10-seed sanity sweep**:

```bash
for seed in 1 2 3 4 5 42 100 200 500 1000; do
  go run ./cmd/world-builder -seed $seed -format markdown | grep -c "→ "
done
```

Handoff predicted #9 could shift counts in either direction depending on whether atmosphere -2 pushes locks below threshold (fewer) or whether re-eval triggers verification differently (more). Document the observed direction in the PR.

- [ ] **Step 5: Commit**:

```bash
git add worlds/generate.go iiss/testdata/
git commit -m "worlds: insert ApplyTidalLockReEval after ApplyClimate (#9)"
```

### Task 4.7: Document the divergence pattern

- [ ] **Modify** `docs/wbh-inconsistencies.md` to add an entry:

```markdown
## WBH p.106 atmosphere DM — implemented via Stage-5-post re-eval cascade

The "atmospheric pressure above 2.5 bar: DM-2" line in the all-cases-common
DMs (p.106) requires atmospheric pressure, which the implementation
determines in Stage 5 (`ApplyClimatePasses`) — after the Stage 4 tidal-lock
pass. Rather than reorder the pipeline (impossible: temperature and geology
read tidal-lock outputs), the implementation runs a re-evaluation cascade:

1. Stage 4 tidal lock runs without the atmosphere DM.
2. `commonTidalLockDMs` captures the DM map onto `TidalLock.PreEvalDMs`.
3. `Body.preTidalLockSnapshot` captures pre-effect Eccentricity, AxialTilt,
   DayLength.
4. After Stage 5, `ApplyTidalLockReEval` walks bodies with atmospheric
   pressure > 2.5 bar, restores snapshots, re-runs tidal lock and
   `ApplyClimatePasses` for affected bodies.

Trade-off: the dice stream consumes both the original tidal-lock rolls and
the re-eval rolls. Deterministic per seed. Faithful to p.106 at the cost
of a second-round consumption.
```

- [ ] **Commit**:

```bash
git add docs/wbh-inconsistencies.md
git commit -m "docs: document atmosphere-DM re-eval cascade (#9)"
```

### Task 4.8: Open PR — reference #9, close on merge.

---

## Cross-cutting checklist (each PR)

Before opening any PR:

- [ ] `task check` (gofumpt + vet + golangci-lint) clean.
- [ ] `task test` (full `go test -race ./...`) clean.
- [ ] `go test ./worlds/ -run TestZed` — every worked-example test passes (these are the fidelity guarantee; if any break, the PR is wrong).
- [ ] `iiss/testdata/seed_*.md` regenerated and the diff inspected manually before committing.
- [ ] 10-seed sweep direction matches the handoff's expectation table.

PR description template:

```markdown
Closes #XX.

## What changed

[1-3 bullets]

## WBH reference

[Page + table/section]

## Determinism impact

[Snapshot diff summary; sweep counts pre/post]

## Tests added

[Test names]
```

---

## Self-review notes

Before handing off this plan:

- **Coverage:** Each of #9/#53/#54/#55/#56 maps to a labeled PR with explicit task list. ✅
- **Type consistency:** `TidalLock.PreEvalDMs` and `body.preTidalLockSnapshot` are introduced in PR 4. `SelectHighestDMCases` (plural) is introduced in PR 3 alongside the kept-as-deprecated singular. `closestSignificantMoon` is introduced in PR 1 Task 1.5 and reused in PR 3 Task 3.2.
- **Open ambiguities flagged for engineer attention:**
  - PR 4 Task 4.5: order of `ClearStage5Output` vs. tidal-lock re-roll re: `body.Atmosphere` availability. Resolution recommended in the task comment.
  - PR 4 Task 4.5: `atmosphereDMWouldChangeOutcome` currently always returns true when pressure > 2.5; refine to actually short-circuit no-op cases if the dice-stream cost matters for snapshot stability.
  - PR 1 Task 1.4: existing `roller.NewScripted` fixtures may need dice-sequence updates if their walk order changes. Inspect, don't blindly reorder.
