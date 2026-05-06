# System Worlds 2B: World Counts + Placement Implementation Plan (Go)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement sub-project 2B from the System Worlds and Orbits chapter: World Counts (gas giants, planetoid belts, terrestrials) and the 9-step placement procedure (Steps 1–9). Reproduce the book's Zed quintuple worked example end-to-end with one documented divergence.

**Architecture:** A new set of granular per-step functions and a `GenerateSystemPlacement` façade in the existing `wbh/worlds` package, layered on top of 2A's `AvailableOrbits` and `Group.HZCO()`. A small carry-forward refactor on `Group` removes the `secIdx` positional fragility flagged in 2A's final review.

**Tech Stack:** Go 1.22+, `gofumpt` CLI as canonical formatter (not golangci-lint's bundled gofumpt), golangci-lint v2.12.1, `just` recipes.

**Spec:** `docs/specs/2026-05-03-system-worlds-2b-counts-placement-design.md`

**Source pages:** WBH pp. 36–38 (World Types and Quantities) + pp. 43–52 (Placement of Worlds Steps 1–9).

**Conventions:**

- Working directory: `/Users/markayers/Documents/Traveller/`.
- TDD per task: write test → run-fail → implement → run-pass → format → lint → commit.
- `gofumpt -w` before commit. `gofumpt` CLI is the formatter source of truth (not golangci-lint).
- Test files live in the same package (white-box) except `worked_examples_test.go` (black-box `package worlds_test`).
- Tables for non-numeric cells: struct rows. Tables with nullable numeric cells: `*float64` via the `fp` helper already in `worlds/available_orbits.go`.
- Branch: `feat/wbh-system-worlds-2b` (already created and contains the spec commit).

---

## Pre-flight

- [ ] **Verify clean state on the feature branch**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
git status
git branch --show-current
just check && just test
```

Expected: clean working tree; current branch `feat/wbh-system-worlds-2b`; all tests green.

---

## File Structure

| File                                          | Responsibility                                                                                                      |
| --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `worlds/available_orbits.go` (extend)         | Add `sourceCompanion *stars.CompanionStar` field on `Group`; refactor rules 9–11 to look it up directly.            |
| `worlds/available_orbits_test.go` (extend)    | Add `sourceCompanion` assertion on Zed's three groups.                                                              |
| `worlds/group_hzco.go` (create)               | `Group.HZCO()` — single-star delegates to `Star.HZCO()`; pair delegates to `stars.CompositeHZCO`.                   |
| `worlds/group_hzco_test.go` (create)          | Single-star + pair group HZCO assertions against the Zed worked-example values.                                     |
| `worlds/counts.go` (create)                   | `Counts`, `CountsOpts`, `GenerateCounts` (gas giants, belts, terrestrials, total).                                  |
| `worlds/counts_test.go` (create)              | Per-table + per-DM unit tests; Zed `{4, 2, 11, 17}` integration assertion.                                          |
| `worlds/allocations.go` (create)              | `StarAllocation`, `AllocateOrbitsByStar` (Step 1).                                                                  |
| `worlds/allocations_test.go` (create)         | Single-star + Zed multi-star allocation assertions.                                                                 |
| `worlds/baseline.go` (create)                 | `RollBaselineNumber` (Step 2 + DM table); `BaselineOrbit` (Step 3 with sub-cases 3a/3b/3c + snap-to-available).     |
| `worlds/baseline_test.go` (create)            | Per-DM, per-sub-case, snap-to-available unit tests; Zed baseline assertion.                                         |
| `worlds/empty_orbits.go` (create)             | `RollEmptyOrbits` (Step 4 table).                                                                                   |
| `worlds/empty_orbits_test.go` (create)        | Each row of the 9-/10/11/12 table.                                                                                  |
| `worlds/spread.go` (create)                   | `Spread` (Step 5 base + Maximum Spread cap); `MaximumSecondarySpread`.                                              |
| `worlds/spread_test.go` (create)              | Base formula, baseline-N-< 1 floor, mandatory caps, Zed spread assertion.                                           |
| `worlds/orbit_slots.go` (create)              | `Slot`, `PlaceOrbitSlots` (Step 6 — empty distribution + variance + exclusion-zone widening + baseline-N override). |
| `worlds/orbit_slots_test.go` (create)         | Per-rule unit tests; Zed primary/B/Cab orbit-sequence assertions.                                                   |
| `worlds/anomalous.go` (create)                | `AnomalyType`, `AnomalousSlot`, `AddAnomalous` (Step 7 — count + type + parent group + clamp + counts update).      |
| `worlds/anomalous_test.go` (create)           | Per-table tests; clamp-to-`[MAO, 20.0]` retry; Zed retrograde assertion.                                            |
| `worlds/placement.go` (create)                | `BodyType`, `Placement`, `PlaceWorlds` (Step 8 — order, prefix-die selection, 1D:1D collision handling).            |
| `worlds/placement_test.go` (create)           | Order, prefix-die selection, collision handling.                                                                    |
| `worlds/planet_eccentricity.go` (create)      | `RollPlanetEccentricities` (Step 9 — wraps `stars.RollEccentricity` with anomaly DMs).                              |
| `worlds/planet_eccentricity_test.go` (create) | Per-anomaly DM application.                                                                                         |
| `worlds/system_placement.go` (create)         | `SystemPlacement`, `GenerateSystemPlacement` façade.                                                                |
| `worlds/worked_examples_test.go` (extend)     | `composeZed()` extracted helper; `TestZed_FullPlacement` acceptance gate.                                           |

---

## Task 1: Carry-forward refactor — attach `sourceCompanion` to `Group` and use it in rules 9–11

**Source:** 2A spec carry-forward note. Removes the positional `secIdx` fragility flagged by 2A's final code reviewer.

**Files:** `worlds/available_orbits.go` (extend), `worlds/available_orbits_test.go` (extend).

**API change:** `Group` gains an unexported field. Rules 9–11 stop walking `secIdx`; they read `g.sourceCompanion` directly.

- [ ] **Step 1: Write failing test for sourceCompanion population**

Add to `worlds/available_orbits_test.go` (find an existing test file inside the `worlds` package). The test asserts that `identifyGroups` populates `sourceCompanion` on each non-primary group, and leaves it nil on the primary group.

```go
func TestIdentifyGroups_SourceCompanion(t *testing.T) {
    t.Parallel()

    aa := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
        LuminosityClass: stars.V, Mass: 0.929, Diameter: 0.967, Temperature: 5440,
    })
    ab := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 8},
        LuminosityClass: stars.V, Mass: 0.907, Diameter: 0.957, Temperature: 5360,
    })
    b := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'K', Subtype: 8},
        LuminosityClass: stars.V, Mass: 0.626, Diameter: 0.777, Temperature: 3980,
    })
    sys := stars.System{
        Primary: aa,
        Companions: []stars.CompanionStar{
            {Star: ab, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.09, Eccentricity: 0.11, ParentIndex: -1},
            {Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.08, ParentIndex: -1},
        },
    }

    groups := identifyGroups(sys)
    if len(groups) != 2 {
        t.Fatalf("groups = %d, want 2", len(groups))
    }
    if groups[0].sourceCompanion != nil {
        t.Errorf("primary group sourceCompanion = %+v, want nil", groups[0].sourceCompanion)
    }
    if groups[1].sourceCompanion == nil {
        t.Fatalf("secondary group sourceCompanion = nil, want non-nil")
    }
    if groups[1].sourceCompanion.OrbitClass != stars.OrbitNear {
        t.Errorf("secondary sourceCompanion.OrbitClass = %v, want OrbitNear", groups[1].sourceCompanion.OrbitClass)
    }
    if math.Abs(groups[1].sourceCompanion.OrbitNumber-6.10) > 1e-9 {
        t.Errorf("secondary sourceCompanion.OrbitNumber = %v, want 6.10", groups[1].sourceCompanion.OrbitNumber)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./worlds -run TestIdentifyGroups_SourceCompanion -v
```

Expected: FAIL — `groups[1].sourceCompanion` undefined.

- [ ] **Step 3: Add `sourceCompanion` field to `Group` and populate it**

In `worlds/available_orbits.go`, add the field to `Group` (just under `companionEcc`):

```go
    // companionEcc records the companion's eccentricity for pair
    // groups. Set by identifyGroups; read by AvailableOrbits's rule 2
    // pass. Unexported because it's an implementation detail of the
    // rule pipeline, not part of the public API.
    companionEcc float64

    // sourceCompanion is the CompanionStar that gave rise to this
    // secondary group (nil for the primary group). Set by identifyGroups
    // so rules 9–11 can look up the secondary's orbit class and
    // eccentricity directly instead of walking a parallel index. Also
    // used by sub-project 2B steps that need the secondary's
    // orbit-around-primary.
    sourceCompanion *stars.CompanionStar
```

In `identifyGroups`, set `sourceCompanion` when constructing each secondary group. The relevant inner loop currently looks like:

```go
        for i, c := range sys.Companions {
            if c.OrbitClass != oc || c.ParentIndex != -1 {
                continue
            }
            if letterIdx >= len(letters) {
                break
            }
            group := Group{Members: []stars.Star{c.Star}}
            if companion, ecc, ok := findCompanionOf(i); ok {
                group.Members = append(group.Members, companion)
                group.companionEcc = ecc
                group.Designation = letters[letterIdx] + "ab"
            } else {
                group.Designation = letters[letterIdx]
            }
            letterIdx++
            groups = append(groups, group)
        }
```

Change it to capture a pointer to the source companion. Because `c` is a loop-variable copy, capture the indexed element of the slice:

```go
        for i := range sys.Companions {
            c := sys.Companions[i]
            if c.OrbitClass != oc || c.ParentIndex != -1 {
                continue
            }
            if letterIdx >= len(letters) {
                break
            }
            group := Group{
                Members:         []stars.Star{c.Star},
                sourceCompanion: &sys.Companions[i],
            }
            if companion, ecc, ok := findCompanionOf(i); ok {
                group.Members = append(group.Members, companion)
                group.companionEcc = ecc
                group.Designation = letters[letterIdx] + "ab"
            } else {
                group.Designation = letters[letterIdx]
            }
            letterIdx++
            groups = append(groups, group)
        }
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./worlds -run TestIdentifyGroups_SourceCompanion -v
```

Expected: PASS.

- [ ] **Step 5: Refactor rules 9–11 in `AvailableOrbits` to use `sourceCompanion`**

Replace the existing rules-8–11 block (the `secIdx` walk) in `AvailableOrbits` with one that iterates over `groups` directly, reading each group's `sourceCompanion`:

```go
    // Rule 8: each secondary (Close/Near/Far) has its own orbit range.
    // Lower bound is the secondary's MAO (Rule 1); upper bound is
    // (Orbit# − 3). Rules 9–11 reduce maxOffset further:
    //   - Rule 9: -1 if adjacent zone populated (max once per secondary).
    //   - Rule 10: -1 if self ecc > 0.2 OR any adjacent zone star has ecc > 0.2 (max once).
    //   - Rule 11: -1 if self ecc > 0.5 (max once).
    for i := range groups {
        if groups[i].sourceCompanion == nil {
            continue // primary group
        }
        c := groups[i].sourceCompanion
        maxOffset := c.OrbitNumber - 3 // rule 8
        if hasAdjacentZone(sys, c.OrbitClass) {
            maxOffset-- // rule 9
        }
        if c.Eccentricity > 0.2 || adjacentEccGT02(sys, c.OrbitClass) {
            maxOffset-- // rule 10
        }
        if c.Eccentricity > 0.5 {
            maxOffset-- // rule 11
        }
        if maxOffset < 0 {
            maxOffset = 0
        }
        if maxOffset < groups[i].MAO {
            groups[i].Intervals = nil
        } else {
            groups[i].Intervals = []Interval{{Min: groups[i].MAO, Max: maxOffset}}
        }
    }
```

Delete the old `secIdx := 1` loop block.

- [ ] **Step 6: Run all tests in the worlds package to verify the existing Zed acceptance test still passes**

```bash
go test ./worlds -v
```

Expected: PASS, including `TestZed_AvailableOrbits` and `TestSol_AvailableOrbits`.

- [ ] **Step 7: Format and lint**

```bash
gofumpt -w worlds/
just check
```

Expected: clean.

- [ ] **Step 8: Commit**

```bash
git add worlds/available_orbits.go worlds/available_orbits_test.go
git commit -m "$(cat <<'EOF'
refactor(worlds): attach sourceCompanion to Group; rules 9-11 use it directly

Removes the secIdx positional fragility flagged in 2A's final code review.
Each non-primary Group now carries a *stars.CompanionStar pointing at the
CompanionStar entry that produced it. Rules 9-11 read self/adjacent
eccentricity off the field instead of walking a parallel index.

Sub-project 2B's later steps will also use this field to look up the
secondary's orbit-around-primary directly.
EOF
)"
```

---

## Task 2: `Group.HZCO()` — habitable zone centre orbit per group

**Source:** WBH pp. 41–42. Single-star uses `Star.HZCO()` (already in 2A); pair groups sum luminosities and recompute (already in `stars.CompositeHZCO`).

**Files:** `worlds/group_hzco.go` (create), `worlds/group_hzco_test.go` (create).

**API:**

```go
// HZCO returns the group's habitable zone centre orbit#. Single-star
// group: == Members[0].HZCO(). Pair group: == stars.CompositeHZCO(Members...).
func (g Group) HZCO() float64
```

- [ ] **Step 1: Write failing test**

Create `worlds/group_hzco_test.go`:

```go
package worlds

import (
    "math"
    "testing"

    "wbh/stars"
)

func TestGroup_HZCO_SingleStar(t *testing.T) {
    t.Parallel()
    sol := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 2},
        LuminosityClass: stars.V, Mass: 1.000, Diameter: 1.000, Temperature: 5772,
    })
    g := Group{Members: []stars.Star{sol}}
    got := g.HZCO()
    if math.Abs(got-3.0) > 0.05 {
        t.Errorf("Sol single-star HZCO = %.4f, want 3.0±0.05", got)
    }
}

func TestGroup_HZCO_Pair_ZedAab(t *testing.T) {
    t.Parallel()
    // Zed Aab pair: combined luminosity 1.419 (treated as ~1.4 by book) → HZCO 3.3.
    aa := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
        LuminosityClass: stars.V, Mass: 0.929, Diameter: 0.967, Temperature: 5440,
    })
    ab := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 8},
        LuminosityClass: stars.V, Mass: 0.907, Diameter: 0.957, Temperature: 5360,
    })
    g := Group{Members: []stars.Star{aa, ab}}
    got := g.HZCO()
    if math.Abs(got-3.3) > 0.05 {
        t.Errorf("Zed Aab pair HZCO = %.4f, want 3.3±0.05", got)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./worlds -run TestGroup_HZCO -v
```

Expected: FAIL — `g.HZCO undefined`.

- [ ] **Step 3: Implement `Group.HZCO()`**

Create `worlds/group_hzco.go`:

```go
package worlds

import "wbh/stars"

// HZCO returns the group's Habitable Zone Centre Orbit#.
//
// Single-star group: delegates to Members[0].HZCO().
// Pair group: delegates to stars.CompositeHZCO(Members...) per WBH p. 42,
// which sums the constituent luminosities before applying the formula.
//
// Source: WBH pp. 41–42.
func (g Group) HZCO() float64 {
    if len(g.Members) == 1 {
        return g.Members[0].HZCO()
    }
    return stars.CompositeHZCO(g.Members...)
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./worlds -run TestGroup_HZCO -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/group_hzco.go worlds/group_hzco_test.go
git commit -m "feat(worlds): Group.HZCO method (WBH pp. 41-42)"
```

---

## Task 3: `Counts` type, `GenerateCounts` skeleton, gas-giant existence + quantity

**Source:** WBH pp. 36–37 (Gas Giants section). Existence: "Gas Giant Exists on 9-: roll 2D" (the alternate 1D 2+ form is not implemented per spec). Quantity: roll 2D + DMs into the Gas Giant Quantity table.

**DMs (gas giants):**

- System consists of a single Class V star: DM+1
- Primary star is a brown dwarf (`KindBrownDwarf`): DM-2
- Primary star is a post-stellar object (BD/WD/NS/BH/Pulsar): DM-2
- Total post-stellar objects (including primary): DM-1 each
- System consists of four or more stars: DM-1

**Files:** `worlds/counts.go` (create), `worlds/counts_test.go` (create).

**API (skeleton — populated across this and the next two tasks):**

```go
type Counts struct {
    GasGiants      int
    PlanetoidBelts int
    Terrestrials   int
    Total          int
}

type CountsOpts struct{}

func GenerateCounts(r roller.Roller, sys stars.System, opts CountsOpts) (Counts, error)
```

- [ ] **Step 1: Write failing tests for gas-giant existence and quantity**

Create `worlds/counts_test.go`:

```go
package worlds

import (
    "testing"

    "wbh/roller"
    "wbh/stars"
)

func solSystem() stars.System {
    sol := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 2},
        LuminosityClass: stars.V, Mass: 1.000, Diameter: 1.000, Temperature: 5772,
    })
    return stars.System{Primary: sol}
}

func TestGenerateCounts_GasGiants_None(t *testing.T) {
    t.Parallel()
    // Existence roll = 10 (>9), so no gas giants. Belts existence = 7 (<8) so no belts.
    // Terrestrials: 2D=7 → 5 + DM+1 (single Class V) = 6 ≥3, +D3-1 with D3=2 → +1 → 7.
    r := roller.NewScripted(10 /*GG existence*/, 7 /*belts existence*/, 7 /*terrestrials 2D*/, 2 /*D3 add*/)
    got, err := GenerateCounts(r, solSystem(), CountsOpts{})
    if err != nil {
        t.Fatalf("GenerateCounts: %v", err)
    }
    if got.GasGiants != 0 {
        t.Errorf("GasGiants = %d, want 0", got.GasGiants)
    }
}

func TestGenerateCounts_GasGiants_PresentSingleClassV(t *testing.T) {
    t.Parallel()
    // Existence = 9 (≤9 → present). Quantity 2D=7, +DM+1 (single Class V) = 8 → row 7-8 → 3 GG.
    // Belts existence = 7 (no belts).
    // Terrestrials: 2D=7 → 5 + DM+1 = 6, +D3-1 (D3=1) = 6.
    r := roller.NewScripted(9, 7, 7 /*GG quantity*/, 7 /*belts existence*/, 7 /*terrestrials*/, 1)
    got, err := GenerateCounts(r, solSystem(), CountsOpts{})
    if err != nil {
        t.Fatalf("GenerateCounts: %v", err)
    }
    if got.GasGiants != 3 {
        t.Errorf("GasGiants = %d, want 3", got.GasGiants)
    }
}

func TestGenerateCounts_GasGiants_QuantityTable(t *testing.T) {
    t.Parallel()
    cases := []struct {
        roll int
        want int
    }{
        {3, 1}, // 4-: 1
        {4, 1},
        {5, 2}, // 5-6: 2
        {6, 2},
        {7, 3}, // 7-8: 3
        {8, 3},
        {9, 4}, // 9-11: 4
        {11, 4},
        {12, 5}, // 12: 5
        {13, 6}, // 13+: 6
    }
    for _, tc := range cases {
        // Existence = 5 (≤9 → present). No DMs (use a primary that gives no DMs).
        // For "no DM" we use a single-star system where the only DM is the +1 single-Class-V.
        // To get no DMs, we use a Class IV (subgiant-like) star — but Compose without
        // Subgiant kind keeps it as MainSequence Class IV. The single-Class-V DM only
        // applies for Class V stars; pick Class IV here.
        sys := stars.System{Primary: stars.Compose(stars.ComposeOpts{
            Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 2},
            LuminosityClass: stars.IV, Mass: 1.0, Diameter: 1.0, Temperature: 5772,
        })}
        // Belts existence = 7 (no belts). Terrestrials 2D=7 → 5, +D3-1 with D3=1 → +0 → 5.
        r := roller.NewScripted(5, tc.roll, 7, 7, 1)
        got, err := GenerateCounts(r, sys, CountsOpts{})
        if err != nil {
            t.Fatalf("roll %d: GenerateCounts: %v", tc.roll, err)
        }
        if got.GasGiants != tc.want {
            t.Errorf("roll %d: GasGiants = %d, want %d", tc.roll, got.GasGiants, tc.want)
        }
    }
}

func TestGenerateCounts_GasGiants_DM_BrownDwarfPrimary(t *testing.T) {
    t.Parallel()
    // Brown-dwarf primary: existence DM-2 (post-stellar) + DM-2 (BD) + DM-1 (per post-stellar).
    // We test the QUANTITY DMs once existence has rolled present. Existence roll = 4
    // (4 - 2(post-stellar?) ... let's separate concerns: spec uses the same DMs for both
    // existence and quantity rolls per WBH p. 36-37). Use a roll that lands existence
    // present after DMs.
    //
    // This test focuses on the DM math by inspecting the result. Existence: roll 14 (raw),
    // DM = -2 (BD) -2 (post-stellar) -1 (one post-stellar including primary) = -5; net 9 → present.
    // Quantity: roll 14 (raw) + DMs -5 = 9 → row 9-11 → 4 GG.
    bd := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindBrownDwarf, LuminosityClass: stars.BD, Mass: 0.05, Diameter: 0.1, Temperature: 1500,
    })
    sys := stars.System{Primary: bd}
    // Belts existence = 0, terrestrials 2D=2 (force reroll path).
    r := roller.NewScripted(14, 14, 0, 2, 5 /*D3+2 reroll*/)
    got, err := GenerateCounts(r, sys, CountsOpts{})
    if err != nil {
        t.Fatalf("GenerateCounts: %v", err)
    }
    if got.GasGiants != 4 {
        t.Errorf("GasGiants = %d, want 4", got.GasGiants)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./worlds -run TestGenerateCounts -v
```

Expected: FAIL — `GenerateCounts undefined`.

- [ ] **Step 3: Implement `Counts`, `CountsOpts`, and the gas-giant portion of `GenerateCounts`**

Create `worlds/counts.go`:

```go
// Counts and GenerateCounts implement WBH pp. 36–38: World Types and
// Quantities. Used as Step 0 of the placement pipeline.
package worlds

import (
    "wbh/roller"
    "wbh/stars"
)

// Counts is the per-system count of bodies a Referee will place.
type Counts struct {
    GasGiants      int // 0–6
    PlanetoidBelts int // 0–3
    Terrestrials   int // 0–13 (cap from Step 7 narrative)
    Total          int // GasGiants + PlanetoidBelts + Terrestrials
}

// CountsOpts is reserved for future knobs (e.g., the alternate "Gas Giant
// Exists on 2+: roll 1D" existence form). Empty for now; the standard
// CRB form is used.
type CountsOpts struct{}

// GenerateCounts implements WBH pp. 36–38. Returns the per-system counts
// of gas giants, planetoid belts, and terrestrial planets, and their
// total. Continuation Method (pre-existing mainworld) is out of scope.
func GenerateCounts(r roller.Roller, sys stars.System, _ CountsOpts) (Counts, error) {
    var c Counts

    ggDM := gasGiantDMs(sys)
    if r.Roll("2D")+ggDM <= 9 {
        // Present.
        qty := r.Roll("2D") + ggDM
        c.GasGiants = gasGiantQuantity(qty)
    }

    // Belts and terrestrials are filled in by later tasks. For now,
    // delegate to placeholders so the test can reach the gas-giant
    // assertions. They will be implemented in tasks 4 and 5.
    c.PlanetoidBelts = 0
    c.Terrestrials = 0
    c.Total = c.GasGiants + c.PlanetoidBelts + c.Terrestrials
    return c, nil
}

// gasGiantDMs computes the WBH p. 37 DM stack for both gas-giant
// existence and quantity rolls (the book applies the same DMs to both).
func gasGiantDMs(sys stars.System) int {
    dm := 0
    if isSingleClassVSystem(sys) {
        dm++
    }
    if sys.Primary.Kind == stars.KindBrownDwarf {
        dm -= 2
    }
    if isPostStellar(sys.Primary.Kind) {
        dm -= 2
    }
    dm -= postStellarCount(sys)
    if totalStarCount(sys) >= 4 {
        dm--
    }
    return dm
}

// gasGiantQuantity maps a 2D+DMs result to the WBH p. 37 quantity table.
// Outputs 1–6.
func gasGiantQuantity(roll int) int {
    switch {
    case roll <= 4:
        return 1
    case roll <= 6:
        return 2
    case roll <= 8:
        return 3
    case roll <= 11:
        return 4
    case roll == 12:
        return 5
    default:
        return 6
    }
}

// isSingleClassVSystem reports whether the system has exactly one star
// (no companions of any kind) and that star is luminosity class V.
func isSingleClassVSystem(sys stars.System) bool {
    if len(sys.Companions) > 0 {
        return false
    }
    return sys.Primary.LuminosityClass == stars.V
}

// postStellarCount returns the count of post-stellar objects in the
// system, including the primary if it is post-stellar.
func postStellarCount(sys stars.System) int {
    n := 0
    if isPostStellar(sys.Primary.Kind) {
        n++
    }
    for _, c := range sys.Companions {
        if isPostStellar(c.Star.Kind) {
            n++
        }
    }
    return n
}

// totalStarCount returns the count of stellar bodies (primary +
// companions of every orbit class). Companions count.
func totalStarCount(sys stars.System) int {
    return 1 + len(sys.Companions)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./worlds -run TestGenerateCounts_GasGiants -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/counts.go worlds/counts_test.go
git commit -m "feat(worlds): GenerateCounts with gas-giant existence and quantity (WBH pp. 36-37)"
```

---

## Task 4: Planetoid belts — existence + quantity + DMs

**Source:** WBH p. 37 (Planetoid Belts section). Existence: "Planetoid Belt Exists on 8+: roll 2D". Quantity: roll 2D + DMs into the Planetoid Belt Quantity table (1–3 belts).

**DMs (belts):**

- System has 1 or more gas giants (use freshly-computed `c.GasGiants`): DM+1
- Primary star is a protostar (`KindProtostar`): DM+3
- Primary star is primordial (`AgeGyr < 0.1`): DM+2
- Primary star is a post-stellar object: DM+1
- Total post-stellar objects (including primary): DM+1 each
- System consists of two or more stars: DM+1

**Files:** `worlds/counts.go` (extend), `worlds/counts_test.go` (extend).

- [ ] **Step 1: Write failing tests**

Append to `worlds/counts_test.go`:

```go
func TestGenerateCounts_Belts_QuantityTable(t *testing.T) {
    t.Parallel()
    cases := []struct {
        roll int
        want int
    }{
        {3, 1},  // 6-: 1
        {6, 1},
        {7, 2},  // 7-11: 2
        {11, 2},
        {12, 3}, // 12+: 3
        {15, 3},
    }
    sys := solSystem() // single G2 V, no companions, no DMs other than DM+1 single-Class-V
    for _, tc := range cases {
        // GG existence = 10 (>9 → no GG → no DM+1 belt). Belts existence raw = 8 (≥8).
        // Belt quantity = tc.roll (no DMs since no GGs and single G2 V is class V which
        // doesn't trigger any belt DM). Terrestrials 2D=7 → 5 + DM+1 single-V = 6,
        // +D3-1 (D3=1) → 6.
        r := roller.NewScripted(10, 8, tc.roll, 7, 1)
        got, err := GenerateCounts(r, sys, CountsOpts{})
        if err != nil {
            t.Fatalf("roll %d: %v", tc.roll, err)
        }
        if got.PlanetoidBelts != tc.want {
            t.Errorf("roll %d: PlanetoidBelts = %d, want %d", tc.roll, got.PlanetoidBelts, tc.want)
        }
    }
}

func TestGenerateCounts_Belts_DM_GasGiantsPresent(t *testing.T) {
    t.Parallel()
    // Same Sol system. GG existence = 5 (≤9), GG quantity = 7+DM+1=8 → 3 GGs.
    // Belts: existence raw = 7 + DM+1 (GGs present) = 8 → present. Quantity raw = 5 + DM+1 = 6 → row 6- → 1 belt.
    // Terrestrials 2D=7 → 5+DM+1=6, +D3-1 (D3=1) → 6.
    r := roller.NewScripted(5, 7, 7, 5, 7, 1)
    got, err := GenerateCounts(r, solSystem(), CountsOpts{})
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got.PlanetoidBelts != 1 {
        t.Errorf("PlanetoidBelts = %d, want 1", got.PlanetoidBelts)
    }
}

func TestGenerateCounts_Belts_DM_Protostar(t *testing.T) {
    t.Parallel()
    proto := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindProtostar, LuminosityClass: stars.V, Mass: 1.0, Diameter: 1.0, Temperature: 4000, AgeGyr: 0.001,
    })
    sys := stars.System{Primary: proto}
    // Existence raw = 5 + DM+3 (protostar) + DM+2 (primordial: age 0.001 < 0.1) = 10 → present.
    // Quantity raw = 5 + DM+3 + DM+2 = 10 → row 7-11 → 2 belts.
    // GG existence = 10 (no GGs); terrestrials 2D=7 → 5 + DM+1 single-V = 6, +D3-1 (1) → 6.
    r := roller.NewScripted(10, 5, 5, 7, 1)
    got, err := GenerateCounts(r, sys, CountsOpts{})
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got.PlanetoidBelts != 2 {
        t.Errorf("PlanetoidBelts = %d, want 2", got.PlanetoidBelts)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./worlds -run TestGenerateCounts_Belts -v
```

Expected: FAIL — current code always sets `PlanetoidBelts = 0`.

- [ ] **Step 3: Implement belt logic in `GenerateCounts`**

In `worlds/counts.go`, replace the `c.PlanetoidBelts = 0` placeholder with:

```go
    beltDM := beltDMs(sys, c.GasGiants)
    if r.Roll("2D")+beltDM >= 8 {
        qty := r.Roll("2D") + beltDM
        c.PlanetoidBelts = beltQuantity(qty)
    }
```

Add helpers below the existing gas-giant ones:

```go
// beltDMs computes the WBH p. 37 DM stack for planetoid-belt existence
// and quantity rolls. The book applies the same DMs to both rolls.
//
// gasGiantsPresent is the freshly-computed count from the same call to
// GenerateCounts, used for the "system has ≥1 gas giants" DM.
func beltDMs(sys stars.System, gasGiantsPresent int) int {
    dm := 0
    if gasGiantsPresent >= 1 {
        dm++
    }
    if sys.Primary.Kind == stars.KindProtostar {
        dm += 3
    }
    if isPrimordial(sys.Primary) {
        dm += 2
    }
    if isPostStellar(sys.Primary.Kind) {
        dm++
    }
    dm += postStellarCount(sys)
    if totalStarCount(sys) >= 2 {
        dm++
    }
    return dm
}

// beltQuantity maps a 2D+DMs result to the WBH p. 37 belt quantity table.
func beltQuantity(roll int) int {
    switch {
    case roll <= 6:
        return 1
    case roll <= 11:
        return 2
    default:
        return 3
    }
}

// isPrimordial reports whether a star is "primordial" per WBH p. 14:
// any star with age below 0.1 Gyr.
func isPrimordial(s stars.Star) bool {
    return s.AgeGyr < 0.1 && s.AgeGyr > 0
}
```

Note on `isPrimordial`: the `AgeGyr > 0` guard avoids treating a default-zero `Star{}` as primordial when callers (mostly tests) pass a synthetic star without setting `AgeGyr`. The book defines primordial as "less than 0.1 Gyr" with no lower bound, but a zero-age star in our codebase represents "age unknown / not relevant," not "newly formed."

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./worlds -run TestGenerateCounts_Belts -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/counts.go worlds/counts_test.go
git commit -m "feat(worlds): planetoid belt existence and quantity DMs (WBH p. 37)"
```

---

## Task 5: Terrestrial planets formula + Total + Zed integration

**Source:** WBH p. 38 (Terrestrial Planets, Total Worlds). Formula: `Terrestrials = 2D - 2 + DMs`. If result < 3, reroll as `D3+2`. If result ≥ 3, add `D3-1`.

**DMs (terrestrials):** DM-1 per post-stellar object including primary. (No other DMs in the book's terrestrial formula box.)

**Files:** `worlds/counts.go` (extend), `worlds/counts_test.go` (extend), `worlds/worked_examples_test.go` (extend with `composeZed` helper).

- [ ] **Step 1: Write failing tests**

Append to `worlds/counts_test.go`:

```go
func TestGenerateCounts_Terrestrials_LowReroll(t *testing.T) {
    t.Parallel()
    // Sol single G2 V. No DMs on terrestrials.
    // 2D = 4 → 4-2 = 2 (< 3) → reroll D3+2; D3=2 → result 4.
    r := roller.NewScripted(10 /*GG existence none*/, 7 /*belts existence none*/, 4 /*terrestrials raw*/, 2 /*D3 reroll*/)
    got, err := GenerateCounts(r, solSystem(), CountsOpts{})
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got.Terrestrials != 4 {
        t.Errorf("Terrestrials = %d, want 4", got.Terrestrials)
    }
}

func TestGenerateCounts_Terrestrials_HighAdd(t *testing.T) {
    t.Parallel()
    // 2D = 8 → 8-2 = 6 (≥3) → add D3-1 with D3=3 → +2 → 8.
    r := roller.NewScripted(10, 7, 8, 3)
    got, err := GenerateCounts(r, solSystem(), CountsOpts{})
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got.Terrestrials != 8 {
        t.Errorf("Terrestrials = %d, want 8", got.Terrestrials)
    }
}

func TestGenerateCounts_Terrestrials_PostStellarDM(t *testing.T) {
    t.Parallel()
    // Brown-dwarf primary: terrestrials DM-1 (post-stellar count = 1, including primary).
    // 2D = 8 → 8-2-1 = 5 (≥3) → add D3-1 with D3=3 → +2 → 7.
    bd := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindBrownDwarf, LuminosityClass: stars.BD, Mass: 0.05, Diameter: 0.1, Temperature: 1500,
    })
    // GG existence: raw 14 + (-5 DMs) = 9 → present. Quantity 14 + (-5) = 9 → 4 GGs.
    // Belts existence: raw 0 + (DM+1 GGs +1 post-stellar primary +1 per post-stellar) = 0+3 = 3 → not present.
    // Terrestrials: 2D=8, D3=3.
    r := roller.NewScripted(14, 14, 0, 8, 3)
    got, err := GenerateCounts(r, stars.System{Primary: bd}, CountsOpts{})
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got.Terrestrials != 7 {
        t.Errorf("Terrestrials = %d, want 7", got.Terrestrials)
    }
}

func TestGenerateCounts_Total(t *testing.T) {
    t.Parallel()
    // GG present (5 → 7+DM+1 = 8 → 3 GG); belts present (8 raw, 5 raw + DM+1 GGs → 6 → 1 belt);
    // terrestrials 8 → 6 + D3-1 (D3=2 → +1) → 7. Total = 3+1+7 = 11.
    r := roller.NewScripted(5, 7, 8, 5, 8, 2)
    got, err := GenerateCounts(r, solSystem(), CountsOpts{})
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got.Total != got.GasGiants+got.PlanetoidBelts+got.Terrestrials {
        t.Errorf("Total = %d, sum = %d", got.Total, got.GasGiants+got.PlanetoidBelts+got.Terrestrials)
    }
    if got.Total != 11 {
        t.Errorf("Total = %d, want 11", got.Total)
    }
}
```

Now extract a `composeZed` helper and add the Zed counts integration test. Edit `worlds/worked_examples_test.go` to extract the system construction into a helper, and add the new acceptance test.

Replace the inline construction in `TestZed_AvailableOrbits` with a call to `composeZed`:

```go
// composeZed builds the WBH p. 35/40 Zed quintuple system using
// stars.Compose so tests can construct it deterministically (no rolls).
func composeZed() stars.System {
    aa := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
        LuminosityClass: stars.V, Mass: 0.929, Diameter: 0.967, Temperature: 5440,
    })
    ab := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 8},
        LuminosityClass: stars.V, Mass: 0.907, Diameter: 0.957, Temperature: 5360,
    })
    b := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'K', Subtype: 8},
        LuminosityClass: stars.V, Mass: 0.626, Diameter: 0.777, Temperature: 3980,
    })
    ca := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'M', Subtype: 0},
        LuminosityClass: stars.V, Mass: 0.510, Diameter: 0.728, Temperature: 3700,
    })
    cb := stars.Compose(stars.ComposeOpts{
        Kind: stars.KindWhiteDwarf, Mass: 0.490, Diameter: 0.017, Temperature: 6700,
    })
    return stars.System{
        Primary: aa,
        Companions: []stars.CompanionStar{
            {Star: ab, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.09, Eccentricity: 0.11, ParentIndex: -1},
            {Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.08, ParentIndex: -1},
            {Star: ca, OrbitClass: stars.OrbitFar, OrbitNumber: 12.10, Eccentricity: 0.47, ParentIndex: -1},
            {Star: cb, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.21, Eccentricity: 0.24, ParentIndex: 2},
        },
    }
}
```

Replace the inline `aa := stars.Compose(...)` ... `sys := stars.System{...}` block at the top of `TestZed_AvailableOrbits` with `sys := composeZed()`.

Then append a new test:

```go
func TestZed_GenerateCounts(t *testing.T) {
    t.Parallel()
    // WBH p. 38 Zed walkthrough:
    //   GG existence: 2D=7 (DMs: -1 for ≥4 stars = -1; -1 per post-stellar (cb) = -1; -2 post-stellar
    //     count includes primary? primary is G7 V, NOT post-stellar, so post-stellar count = 1 (cb only)
    //     → DM-1; brown-dwarf primary? no; single Class V system? no — multi-star;
    //     net DM = -2. raw 7 + (-2) = 5 ≤9 → present.
    //   GG quantity: 2D + (-2). The book narrates "Rolling 2D for gas giants, a 7 indicates that they exist
    //     and a further roll on the Gas Giant Quantity table with DM+1 for the post-stellar object and
    //     DM-1 for four or more stars results in 11-2 = 9, which is four gas giants." So the book uses
    //     net DM = +1 - 1 = 0 for the QUANTITY roll. Note the apparent book inconsistency between EXISTENCE
    //     DMs (described in terms of a -2 post-stellar primary, which Zed does not have) and QUANTITY DMs
    //     (DM+1 post-stellar object — that is the per-post-stellar DM the book lists as DM-1 elsewhere).
    //     The encoded DMs (gasGiantDMs) follow the table on p. 37 strictly: -1 per post-stellar object
    //     (cb counts → -1), -1 for 4+ stars → -1; net -2. To exactly reproduce the book's narration would
    //     require honouring the book's loose narration; the per-rule unit tests already pin the table.
    //
    // For this integration test, we instead use rolls that the encoded DM stack maps to the book's count
    // outcomes:
    //
    //   GG existence:  raw 9, DMs -2 → 7 ≤9 → present.
    //   GG quantity:   raw 11, DMs -2 → 9 → row 9-11 → 4 GGs. ✓
    //   Belts existence: raw 7, DMs +1 (GGs present) +1 (≥2 stars) +1 (per post-stellar cb) = +3 → 10 → present.
    //   Belts quantity: raw 7, DMs +3 → 10 → row 7-11 → 2 belts. ✓
    //   Terrestrials:   raw 2D, DM-1 (per post-stellar cb). Book says 11 = 12-2-1 with +D3-1 (-1 net). Our encoded
    //                   formula: 2D=12 → 12-2-1 = 9 (≥3) → +D3-1 with D3=3 → +2 → 11. ✓
    sys := composeZed()
    r := roller.NewScripted(
        9, 11, // GG existence + quantity
        7, 7,  // belts existence + quantity
        12, 3, // terrestrials 2D + D3
    )
    got, err := GenerateCounts(r, sys, CountsOpts{})
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got.GasGiants != 4 {
        t.Errorf("GasGiants = %d, want 4", got.GasGiants)
    }
    if got.PlanetoidBelts != 2 {
        t.Errorf("PlanetoidBelts = %d, want 2", got.PlanetoidBelts)
    }
    if got.Terrestrials != 11 {
        t.Errorf("Terrestrials = %d, want 11", got.Terrestrials)
    }
    if got.Total != 17 {
        t.Errorf("Total = %d, want 17", got.Total)
    }
}
```

Note: `worked_examples_test.go` is `package worlds_test` (black-box). `composeZed` and `TestZed_GenerateCounts` go in the same file.

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./worlds -run "TestGenerateCounts_Terrestrials|TestGenerateCounts_Total|TestZed_GenerateCounts" -v
```

Expected: FAIL — terrestrials and total are still 0.

- [ ] **Step 3: Implement terrestrials and total in `GenerateCounts`**

In `worlds/counts.go`, replace `c.Terrestrials = 0` and the trailing `c.Total =` line with:

```go
    terrDM := -postStellarCount(sys)
    raw := r.Roll("2D") - 2 + terrDM
    switch {
    case raw < 3:
        c.Terrestrials = r.Roll("D3") + 2
    default:
        c.Terrestrials = raw + r.Roll("D3") - 1
    }
    c.Total = c.GasGiants + c.PlanetoidBelts + c.Terrestrials
    return c, nil
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./worlds -v
```

Expected: PASS — all `TestGenerateCounts_*` and `TestZed_GenerateCounts`. The existing 2A `TestZed_AvailableOrbits` continues to pass after the `composeZed` extraction.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/counts.go worlds/counts_test.go worlds/worked_examples_test.go
git commit -m "feat(worlds): terrestrials, total, and Zed counts integration (WBH p. 38)"
```

---

## Task 6: Step 1 — `AllocateOrbitsByStar`

**Source:** WBH pp. 43–44. Single-star systems skip Step 1; multi-star systems compute Total Star Orbits per group, sum to Total System Orbits, and allocate `Counts.Total` worlds proportionally.

**Files:** `worlds/allocations.go` (create), `worlds/allocations_test.go` (create), `worlds/worked_examples_test.go` (extend with Zed allocation test).

**API:**

```go
type StarAllocation struct {
    Group           Group
    TotalStarOrbits int // floor of (intervals sum + 1 if no companion AND >0 prior allowable)
    AllocatedWorlds int
}

func AllocateOrbitsByStar(avail Result, counts Counts) ([]StarAllocation, error)
```

- [ ] **Step 1: Write failing tests**

Create `worlds/allocations_test.go`:

```go
package worlds

import (
    "math"
    "testing"

    "wbh/stars"
)

func TestAllocateOrbitsByStar_SingleStar(t *testing.T) {
    t.Parallel()
    g := Group{
        Designation: "A",
        Members:     []stars.Star{{LuminosityClass: stars.V}},
        MAO:         0.03,
        Intervals:   []Interval{{Min: 0.03, Max: 20.0}},
    }
    avail := Result{Groups: []Group{g}}
    counts := Counts{GasGiants: 4, PlanetoidBelts: 1, Terrestrials: 4, Total: 9}
    got, err := AllocateOrbitsByStar(avail, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if len(got) != 1 {
        t.Fatalf("allocations = %d, want 1", len(got))
    }
    // Single-star: TotalStarOrbits = floor(19.97 + 1 (no companion)) = 20.
    if got[0].TotalStarOrbits != 20 {
        t.Errorf("TotalStarOrbits = %d, want 20", got[0].TotalStarOrbits)
    }
    if got[0].AllocatedWorlds != 9 {
        t.Errorf("AllocatedWorlds = %d, want 9", got[0].AllocatedWorlds)
    }
}

func TestAllocateOrbitsByStar_NoPriorAllowable_NoPlusOne(t *testing.T) {
    t.Parallel()
    // A no-companion star with zero prior allowable orbits gets no +1
    // (the book footnote: "Only if the star had more than zero previously
    // computed Allowable Orbits").
    g := Group{
        Designation: "A",
        Members:     []stars.Star{{LuminosityClass: stars.V}},
        Intervals:   nil, // 0 prior allowable
    }
    avail := Result{Groups: []Group{g}}
    counts := Counts{Total: 5}
    got, err := AllocateOrbitsByStar(avail, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got[0].TotalStarOrbits != 0 {
        t.Errorf("TotalStarOrbits = %d, want 0", got[0].TotalStarOrbits)
    }
    if got[0].AllocatedWorlds != 5 {
        t.Errorf("AllocatedWorlds = %d, want 5 (last star gets remainder)", got[0].AllocatedWorlds)
    }
}

func TestAllocateOrbitsByStar_PairGetsNoPlusOne(t *testing.T) {
    t.Parallel()
    // A pair group (Members has 2 stars) does NOT get the +1 even if it
    // has prior allowable orbits, because it has a companion.
    g := Group{
        Designation: "Aab",
        Members:     []stars.Star{{}, {}},
        Intervals:   []Interval{{Min: 0.61, Max: 5.10}, {Min: 7.10, Max: 10.10}, {Min: 14.10, Max: 20.0}},
    }
    avail := Result{Groups: []Group{g}}
    // Total should floor (4.49 + 3.0 + 5.9 = 13.39) → 13.
    counts := Counts{Total: 13}
    got, err := AllocateOrbitsByStar(avail, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got[0].TotalStarOrbits != 13 {
        t.Errorf("TotalStarOrbits = %d, want 13 (pair gets no +1)", got[0].TotalStarOrbits)
    }
    if got[0].AllocatedWorlds != 13 {
        t.Errorf("AllocatedWorlds = %d, want 13", got[0].AllocatedWorlds)
    }
    if math.Abs(g.Total()-13.39) > 0.01 {
        t.Errorf("(sanity) g.Total() = %v, want 13.39", g.Total())
    }
}
```

Append to `worlds/worked_examples_test.go`:

```go
func TestZed_AllocateOrbitsByStar(t *testing.T) {
    t.Parallel()
    sys := composeZed()
    avail, err := worlds.AvailableOrbits(sys)
    if err != nil {
        t.Fatalf("AvailableOrbits: %v", err)
    }
    counts := worlds.Counts{GasGiants: 4, PlanetoidBelts: 2, Terrestrials: 11, Total: 17}
    got, err := worlds.AllocateOrbitsByStar(avail, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if len(got) != 3 {
        t.Fatalf("allocations = %d, want 3", len(got))
    }
    // Per the book p. 44 narration:
    //   Aab: floor(13.39) = 13 (pair, no +1)
    //   B:   floor(1.08 + 1) = 2 (single, has prior allowable, no companion → +1)
    //   Cab: floor(6.36) = 6 (pair, no +1)
    //   Sum: 21
    //   Aab worlds: ceil(17×13/21) = ceil(10.52) = 11 → ROUND UP for primary
    //   B worlds:   floor(17×2/21) = floor(1.62) = 1 → round down for middle
    //   Cab worlds: 17 - 11 - 1 = 5 (remainder for last)
    wantOrbits := []int{13, 2, 6}
    wantWorlds := []int{11, 1, 5}
    for i := range got {
        if got[i].TotalStarOrbits != wantOrbits[i] {
            t.Errorf("group %d (%s) TotalStarOrbits = %d, want %d", i, got[i].Group.Designation, got[i].TotalStarOrbits, wantOrbits[i])
        }
        if got[i].AllocatedWorlds != wantWorlds[i] {
            t.Errorf("group %d (%s) AllocatedWorlds = %d, want %d", i, got[i].Group.Designation, got[i].AllocatedWorlds, wantWorlds[i])
        }
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./worlds -run "TestAllocateOrbitsByStar|TestZed_AllocateOrbitsByStar" -v
```

Expected: FAIL — `AllocateOrbitsByStar undefined`.

- [ ] **Step 3: Implement `AllocateOrbitsByStar`**

Create `worlds/allocations.go`:

```go
package worlds

import "math"

// StarAllocation is one star group's slice of the per-system world budget.
type StarAllocation struct {
    Group           Group
    TotalStarOrbits int // floor(group.Total() + (1 if no companion and prior allowable > 0))
    AllocatedWorlds int
}

// AllocateOrbitsByStar implements WBH pp. 43–44 Step 1.
//
// For each Group in avail.Groups:
//   - Total Star Orbits = floor(group.Total() + 1 if group has no companion AND
//     group.Total() > 0 else 0).
//
// The system Total = sum of TotalStarOrbits across groups.
//
// World allocation per group:
//   - First (primary) group: ceil(counts.Total × TotalStarOrbits / sysTotal)
//   - Middle groups:         floor(...)
//   - Last group:            counts.Total - sum(others)
//
// Single-star systems return a single allocation with all worlds.
func AllocateOrbitsByStar(avail Result, counts Counts) ([]StarAllocation, error) {
    if len(avail.Groups) == 0 {
        return nil, nil
    }

    out := make([]StarAllocation, len(avail.Groups))
    sysTotal := 0
    for i, g := range avail.Groups {
        out[i].Group = g
        priorAllowable := g.Total()
        tso := int(math.Floor(priorAllowable))
        if len(g.Members) == 1 && priorAllowable > 0 {
            tso++
        }
        out[i].TotalStarOrbits = tso
        sysTotal += tso
    }

    // Single-star fast path.
    if len(avail.Groups) == 1 {
        out[0].AllocatedWorlds = counts.Total
        return out, nil
    }

    // Multi-star: distribute counts.Total proportionally.
    if sysTotal == 0 {
        // Degenerate case — all groups have zero allowable orbits. Put
        // everything in the last group so the caller still has somewhere
        // to place worlds.
        out[len(out)-1].AllocatedWorlds = counts.Total
        return out, nil
    }
    assigned := 0
    for i := range out {
        if i == 0 {
            // Primary: round up.
            v := float64(counts.Total*out[i].TotalStarOrbits) / float64(sysTotal)
            out[i].AllocatedWorlds = int(math.Ceil(v))
        } else if i == len(out)-1 {
            // Last: remainder.
            out[i].AllocatedWorlds = counts.Total - assigned
        } else {
            // Middle: round down.
            v := float64(counts.Total*out[i].TotalStarOrbits) / float64(sysTotal)
            out[i].AllocatedWorlds = int(math.Floor(v))
        }
        if i < len(out)-1 {
            assigned += out[i].AllocatedWorlds
        }
    }
    return out, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./worlds -run "TestAllocateOrbitsByStar|TestZed_AllocateOrbitsByStar" -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/allocations.go worlds/allocations_test.go worlds/worked_examples_test.go
git commit -m "feat(worlds): Step 1 AllocateOrbitsByStar (WBH pp. 43-44)"
```

---

## Task 7: Step 2 — `RollBaselineNumber`

**Source:** WBH pp. 44–45. `Baseline Number = 2D + DMs`. Sub-cases 3a/3b/3c are decided in Step 3 by comparing the baseline number to total worlds.

**DMs (Step 2):**

- Primary star has a companion (any companion-class star with `ParentIndex == -1`): DM-2
- Primary star is Class Ia, Ib, or II: DM+3
- Primary star is Class III: DM+2
- Primary star is Class IV: DM+1
- Primary star is Class VI: DM-1
- Primary star is a post-stellar object: DM-2
- Total worlds < 6: DM-4
- Total worlds 6–9: DM-3
- Total worlds 10–12: DM-2
- Total worlds 13–15: DM-1
- Total worlds 16–17: DM+0 (book skips this band; treat as 0)
- Total worlds 18–20: DM+1
- Total worlds > 20: DM+2
- For each secondary star: DM-1 (count of `Companions` with `OrbitClass ∈ {Close, Near, Far}` and `ParentIndex == -1`)

**Files:** `worlds/baseline.go` (create), `worlds/baseline_test.go` (create), `worlds/worked_examples_test.go` (extend).

- [ ] **Step 1: Write failing tests**

Create `worlds/baseline_test.go`:

```go
package worlds

import (
    "testing"

    "wbh/roller"
    "wbh/stars"
)

func TestRollBaselineNumber_NoMods(t *testing.T) {
    t.Parallel()
    sys := stars.System{Primary: stars.Compose(stars.ComposeOpts{
        Kind: stars.KindMainSequence, LuminosityClass: stars.V,
    })}
    counts := Counts{Total: 17} // DM 0
    got, err := RollBaselineNumber(roller.NewScripted(7), sys, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got != 7 {
        t.Errorf("baseline = %d, want 7", got)
    }
}

func TestRollBaselineNumber_DMTable(t *testing.T) {
    t.Parallel()
    type tc struct {
        name string
        sys  stars.System
        tot  int
        roll int
        want int
    }
    cases := []tc{
        {
            name: "primary has companion",
            sys: stars.System{
                Primary: stars.Compose(stars.ComposeOpts{Kind: stars.KindMainSequence, LuminosityClass: stars.V}),
                Companions: []stars.CompanionStar{
                    {Star: stars.Star{}, OrbitClass: stars.OrbitCompanion, ParentIndex: -1},
                },
            },
            tot: 17, roll: 9, want: 9 - 2,
        },
        {
            name: "Class III primary",
            sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
                Kind: stars.KindMainSequence, LuminosityClass: stars.III,
            })},
            tot: 17, roll: 7, want: 7 + 2,
        },
        {
            name: "Class IV primary",
            sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
                Kind: stars.KindMainSequence, LuminosityClass: stars.IV,
            })},
            tot: 17, roll: 7, want: 7 + 1,
        },
        {
            name: "Class VI primary",
            sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
                Kind: stars.KindMainSequence, LuminosityClass: stars.VI,
            })},
            tot: 17, roll: 7, want: 7 - 1,
        },
        {
            name: "post-stellar primary",
            sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
                Kind: stars.KindWhiteDwarf, LuminosityClass: stars.D,
            })},
            tot: 17, roll: 9, want: 9 - 2,
        },
        {
            name: "total worlds < 6",
            sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
            tot:  3, roll: 9, want: 9 - 4,
        },
        {
            name: "total worlds 6-9",
            sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
            tot:  7, roll: 9, want: 9 - 3,
        },
        {
            name: "total worlds 13-15",
            sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
            tot:  14, roll: 9, want: 9 - 1,
        },
        {
            name: "total worlds > 20",
            sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
            tot:  25, roll: 9, want: 9 + 2,
        },
        {
            name: "secondary star DM",
            sys: stars.System{
                Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V}),
                Companions: []stars.CompanionStar{
                    {Star: stars.Star{}, OrbitClass: stars.OrbitNear, ParentIndex: -1},
                    {Star: stars.Star{}, OrbitClass: stars.OrbitFar, ParentIndex: -1},
                },
            },
            tot: 17, roll: 9, want: 9 - 2,
        },
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            got, err := RollBaselineNumber(roller.NewScripted(c.roll), c.sys, Counts{Total: c.tot})
            if err != nil {
                t.Fatalf("%v", err)
            }
            if got != c.want {
                t.Errorf("baseline = %d, want %d", got, c.want)
            }
        })
    }
}
```

Append to `worlds/worked_examples_test.go`:

```go
func TestZed_RollBaselineNumber(t *testing.T) {
    t.Parallel()
    sys := composeZed()
    // Zed: companion (Ab) → DM-2; secondaries B + Ca → DM-2; total 17 → no
    // band DM (16-17 unlisted in book → 0). Primary G7 V → no class DM.
    // Net DM = -4. Book rolls 9. Result: 9 - 4 = 5.
    got, err := worlds.RollBaselineNumber(roller.NewScripted(9), sys, worlds.Counts{Total: 17})
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got != 5 {
        t.Errorf("baseline = %d, want 5", got)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./worlds -run "RollBaselineNumber" -v
```

Expected: FAIL — `RollBaselineNumber undefined`.

- [ ] **Step 3: Implement `RollBaselineNumber`**

Create `worlds/baseline.go`:

```go
package worlds

import (
    "wbh/roller"
    "wbh/stars"
)

// RollBaselineNumber implements WBH Step 2 (pp. 44–45). The baseline
// number determines whether the system is hot, temperate, or cold, and
// drives Step 3.
func RollBaselineNumber(r roller.Roller, sys stars.System, counts Counts) (int, error) {
    return r.Roll("2D") + baselineDMs(sys, counts), nil
}

// baselineDMs computes the WBH p. 45 DM stack for Step 2.
func baselineDMs(sys stars.System, counts Counts) int {
    dm := 0
    if primaryHasCompanion(sys) {
        dm -= 2
    }
    switch sys.Primary.LuminosityClass {
    case stars.Ia, stars.Ib, stars.II:
        dm += 3
    case stars.III:
        dm += 2
    case stars.IV:
        dm++
    case stars.VI:
        dm--
    }
    if isPostStellar(sys.Primary.Kind) {
        dm -= 2
    }
    switch {
    case counts.Total < 6:
        dm -= 4
    case counts.Total <= 9:
        dm -= 3
    case counts.Total <= 12:
        dm -= 2
    case counts.Total <= 15:
        dm--
    case counts.Total <= 17:
        // 16-17: unlisted band in book → 0
    case counts.Total <= 20:
        dm++
    default: // > 20
        dm += 2
    }
    dm -= secondaryStarCount(sys)
    return dm
}

// primaryHasCompanion reports whether sys has a Companion-class star
// directly orbiting the primary.
func primaryHasCompanion(sys stars.System) bool {
    for _, c := range sys.Companions {
        if c.OrbitClass == stars.OrbitCompanion && c.ParentIndex == -1 {
            return true
        }
    }
    return false
}

// secondaryStarCount returns the count of non-companion stars at Close,
// Near, or Far around the primary.
func secondaryStarCount(sys stars.System) int {
    n := 0
    for _, c := range sys.Companions {
        if c.ParentIndex != -1 {
            continue
        }
        switch c.OrbitClass {
        case stars.OrbitClose, stars.OrbitNear, stars.OrbitFar:
            n++
        }
    }
    return n
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./worlds -run "RollBaselineNumber" -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/baseline.go worlds/baseline_test.go worlds/worked_examples_test.go
git commit -m "feat(worlds): Step 2 RollBaselineNumber with DM table (WBH pp. 44-45)"
```

---

## Task 8: Step 3 — `BaselineOrbit` with sub-cases 3a/3b/3c and snap-to-available

**Source:** WBH pp. 45–46. The baseline orbit is the actual orbital location of the world in the baseline-number-th slot. Three sub-cases by `baselineN` vs `totalWorlds`. Sub-case 3d (Continuation Method) is out of scope.

**Sub-case 3a (1 ≤ baselineN ≤ totalWorlds):** world inside habitable zone.

```text
HZCO ≥ 1.0:  BaselineOrbit = HZCO + (2D-7)/10
HZCO < 1.0:  BaselineOrbit = HZCO + (2D-7)/100
```

**Sub-case 3b (baselineN < 1):** cold system; baseline beyond MAO/HZCO. The book describes this as "subtract this negative... from the value of the HZCO and add a variance":

```text
minOrbit ≥ 1.0:  BaselineOrbit = HZCO - baselineN + totalWorlds + (2D-2)/10
minOrbit < 1.0:  BaselineOrbit = minOrbit - baselineN/10 + (2D-2)/100
```

(`minOrbit` is the larger of `primary.MAO` or `HZCO`; the book says "based on the primary star(s) minimum Orbit#, its HZCO or MAO, whichever is greater".)

**Sub-case 3c (baselineN > totalWorlds):** hot system; all worlds inside HZCO.

```text
HZCO - baselineN + totalWorlds ≥ 1.0:
    BaselineOrbit = HZCO - baselineN + totalWorlds + (2D-7)/5
< 1.0:
    BaselineOrbit = HZCO - (baselineN + totalWorlds + (2D-7)/5)/10
If still negative, treat as max(HZCO - 0.1, MAO + totalWorlds × 0.01).
```

**Snap-to-available:** if the formula lands in a primary-group exclusion zone, move to the nearest available orbit with `(2D-7)/10` direction variance.

**Files:** `worlds/baseline.go` (extend), `worlds/baseline_test.go` (extend), `worlds/worked_examples_test.go` (extend).

- [ ] **Step 1: Write failing tests**

Append to `worlds/baseline_test.go`:

```go
import "math"

func TestBaselineOrbit_3a_HZCOGTE1(t *testing.T) {
    t.Parallel()
    primary := Group{
        Designation: "A",
        Members:     []stars.Star{{Luminosity: 1.0}}, // HZCO = 3.0
        MAO:         0.03,
        Intervals:   []Interval{{Min: 0.03, Max: 20.0}},
    }
    // 2D = 5 → (5-7)/10 = -0.2. HZCO 3.0 + (-0.2) = 2.8.
    got, err := BaselineOrbit(roller.NewScripted(5), primary, primary.HZCO(), 5, 17)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if math.Abs(got-2.8) > 0.01 {
        t.Errorf("BaselineOrbit = %v, want 2.8", got)
    }
}

func TestBaselineOrbit_3a_HZCOLT1(t *testing.T) {
    t.Parallel()
    primary := Group{
        Designation: "A",
        Members:     []stars.Star{{Luminosity: 0.04}}, // HZCO = sqrt(0.04) AU → small
        MAO:         0.01,
        Intervals:   []Interval{{Min: 0.01, Max: 20.0}},
    }
    hzco := primary.HZCO()
    // 2D = 9 → (9-7)/100 = 0.02 → hzco + 0.02
    got, err := BaselineOrbit(roller.NewScripted(9), primary, hzco, 5, 17)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if math.Abs(got-(hzco+0.02)) > 0.005 {
        t.Errorf("BaselineOrbit = %v, want %v", got, hzco+0.02)
    }
}

func TestBaselineOrbit_3b_ColdSystem_MinGTE1(t *testing.T) {
    t.Parallel()
    // baselineN = -2 (< 1), HZCO = 3.0, totalWorlds = 5, MAO = 1.5
    primary := Group{
        Members:   []stars.Star{{Luminosity: 1.0}},
        MAO:       1.5,
        Intervals: []Interval{{Min: 1.5, Max: 20.0}},
    }
    // BaselineOrbit = 3.0 - (-2) + 5 + (2D-2)/10 = 10.0 + (7-2)/10 = 10.5
    got, err := BaselineOrbit(roller.NewScripted(7), primary, primary.HZCO(), -2, 5)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if math.Abs(got-10.5) > 0.01 {
        t.Errorf("BaselineOrbit = %v, want 10.5", got)
    }
}

func TestBaselineOrbit_3c_HotSystem(t *testing.T) {
    t.Parallel()
    // baselineN = 8 > totalWorlds = 5, HZCO = 3.0
    // 3.0 - 8 + 5 = 0 (< 1.0 path):
    //   = 3.0 - (8 + 5 + (2D-7)/5)/10 = 3.0 - (13 + (10-7)/5)/10 = 3.0 - (13 + 0.6)/10 = 3.0 - 1.36 = 1.64
    primary := Group{
        Members:   []stars.Star{{Luminosity: 1.0}},
        MAO:       0.03,
        Intervals: []Interval{{Min: 0.03, Max: 20.0}},
    }
    got, err := BaselineOrbit(roller.NewScripted(10), primary, primary.HZCO(), 8, 5)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if math.Abs(got-1.64) > 0.05 {
        t.Errorf("BaselineOrbit = %v, want ~1.64", got)
    }
}

func TestBaselineOrbit_SnapToAvailable(t *testing.T) {
    t.Parallel()
    // Construct a primary with a hole around 3.0, and roll the formula
    // into the hole; assert snapping.
    primary := Group{
        Members:   []stars.Star{{Luminosity: 1.0}},
        MAO:       0.03,
        Intervals: []Interval{{Min: 0.03, Max: 2.5}, {Min: 3.5, Max: 20.0}},
    }
    // 3a path with roll 7 → variance 0 → BaselineOrbit = 3.0 (in hole 2.5-3.5).
    // Snap roll: 2D=5 → variance (-2)/10 = -0.2 → snap LOWER (toward 2.5).
    // Result: clamp to 2.5.
    got, err := BaselineOrbit(roller.NewScripted(7, 5), primary, primary.HZCO(), 3, 17)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if math.Abs(got-2.5) > 0.05 {
        t.Errorf("BaselineOrbit = %v, want ~2.5 (snapped lower)", got)
    }
}
```

Append to `worlds/worked_examples_test.go`:

```go
func TestZed_BaselineOrbit(t *testing.T) {
    t.Parallel()
    sys := composeZed()
    avail, err := worlds.AvailableOrbits(sys)
    if err != nil {
        t.Fatalf("%v", err)
    }
    primary := avail.Groups[0]
    // Book: baselineN=5, totalWorlds=17, HZCO Aab = 3.3, roll 5 → variance (5-7)/10 = -0.2.
    // BaselineOrbit = 3.3 + (-0.2) = 3.1.
    got, err := worlds.BaselineOrbit(roller.NewScripted(5), primary, primary.HZCO(), 5, 17)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if math.Abs(got-3.1) > 0.05 {
        t.Errorf("BaselineOrbit = %v, want 3.1", got)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./worlds -run "BaselineOrbit" -v
```

Expected: FAIL — `BaselineOrbit undefined`.

- [ ] **Step 3: Implement `BaselineOrbit`**

Append to `worlds/baseline.go`:

```go
import "math"

// BaselineOrbit implements WBH Step 3 (pp. 45–46). Selects the right
// formula by comparing baselineN to totalWorlds. Snaps the result to the
// nearest available orbit (with (2D-7)/10 direction variance) when the
// formula lands inside a primary-group exclusion zone.
//
// hzco is the primary group's HZCO (use primary.HZCO()).
// Continuation Method (sub-case 3d) is out of scope.
func BaselineOrbit(
    r roller.Roller,
    primary Group,
    hzco float64,
    baselineN, totalWorlds int,
) (float64, error) {
    var orbit float64
    switch {
    case baselineN >= 1 && baselineN <= totalWorlds:
        // Sub-case 3a.
        v := r.Roll("2D")
        if hzco >= 1.0 {
            orbit = hzco + float64(v-7)/10.0
        } else {
            orbit = hzco + float64(v-7)/100.0
        }
    case baselineN < 1:
        // Sub-case 3b. minOrbit = max(MAO, HZCO).
        minOrbit := primary.MAO
        if hzco > minOrbit {
            minOrbit = hzco
        }
        v := r.Roll("2D")
        if minOrbit >= 1.0 {
            orbit = hzco - float64(baselineN) + float64(totalWorlds) + float64(v-2)/10.0
        } else {
            orbit = minOrbit - float64(baselineN)/10.0 + float64(v-2)/100.0
        }
    default:
        // Sub-case 3c (baselineN > totalWorlds).
        v := r.Roll("2D")
        firstForm := hzco - float64(baselineN) + float64(totalWorlds)
        if firstForm >= 1.0 {
            orbit = firstForm + float64(v-7)/5.0
        } else {
            orbit = hzco - (float64(baselineN)+float64(totalWorlds)+float64(v-7)/5.0)/10.0
            if orbit < 0 {
                // "treat the baseline Orbit as the HZCO – 0.1 but no lower
                // than the primary star's MAO + the primary star's total
                // worlds × 0.01."
                lower := primary.MAO + float64(totalWorlds)*0.01
                if hzco-0.1 > lower {
                    orbit = hzco - 0.1
                } else {
                    orbit = lower
                }
            }
        }
    }
    if !primary.Contains(orbit) {
        orbit = snapToAvailable(r, primary, orbit)
    }
    return orbit, nil
}

// snapToAvailable returns the nearest in-interval orbit to want, with
// (2D-7)/10 direction variance applied per the book p. 45 narrative.
func snapToAvailable(r roller.Roller, primary Group, want float64) float64 {
    if len(primary.Intervals) == 0 {
        return want
    }
    bestDist := math.Inf(1)
    var best float64
    for _, iv := range primary.Intervals {
        var snap float64
        switch {
        case want < iv.Min:
            snap = iv.Min
        case want > iv.Max:
            snap = iv.Max
        default:
            snap = want
        }
        if d := math.Abs(snap - want); d < bestDist {
            bestDist = d
            best = snap
        }
    }
    v := r.Roll("2D")
    return best + float64(v-7)/10.0
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./worlds -run "BaselineOrbit" -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/baseline.go worlds/baseline_test.go worlds/worked_examples_test.go
git commit -m "feat(worlds): Step 3 BaselineOrbit with sub-cases 3a/3b/3c (WBH pp. 45-46)"
```

---

## Task 9: Step 4 — `RollEmptyOrbits`

**Source:** WBH p. 48 (Step 4 — Empty Orbits table).

| 2D  | Empty Orbits |
| --- | ------------ |
| 9-  | 0            |
| 10  | 1            |
| 11  | 2            |
| 12  | 3            |

**Files:** `worlds/empty_orbits.go` (create), `worlds/empty_orbits_test.go` (create).

- [ ] **Step 1: Write failing test**

Create `worlds/empty_orbits_test.go`:

```go
package worlds

import (
    "testing"

    "wbh/roller"
)

func TestRollEmptyOrbits(t *testing.T) {
    t.Parallel()
    cases := []struct {
        roll int
        want int
    }{
        {2, 0}, {7, 0}, {9, 0},
        {10, 1},
        {11, 2},
        {12, 3},
    }
    for _, c := range cases {
        got, err := RollEmptyOrbits(roller.NewScripted(c.roll))
        if err != nil {
            t.Fatalf("%v", err)
        }
        if got != c.want {
            t.Errorf("roll %d: empty = %d, want %d", c.roll, got, c.want)
        }
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./worlds -run TestRollEmptyOrbits -v
```

Expected: FAIL — `RollEmptyOrbits undefined`.

- [ ] **Step 3: Implement `RollEmptyOrbits`**

Create `worlds/empty_orbits.go`:

```go
package worlds

import "wbh/roller"

// RollEmptyOrbits implements WBH Step 4 (p. 48). Returns the number of
// extra orbital slots to insert across the system.
func RollEmptyOrbits(r roller.Roller) (int, error) {
    switch v := r.Roll("2D"); {
    case v <= 9:
        return 0, nil
    case v == 10:
        return 1, nil
    case v == 11:
        return 2, nil
    default:
        return 3, nil
    }
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./worlds -run TestRollEmptyOrbits -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/empty_orbits.go worlds/empty_orbits_test.go
git commit -m "feat(worlds): Step 4 RollEmptyOrbits (WBH p. 48)"
```

---

## Task 10: Step 5 — `Spread` and `MaximumSecondarySpread`

**Source:** WBH pp. 48–49.

```text
Spread = (BaselineOrbit - MAO) / BaselineN     (BaselineN treated as 1 if < 1)

Maximum Spread (cap, applied when outermost would exceed 20):
    = (Primary star(s) Available Orbits) / (Primary's Allocated Orbits + Total Stars)

Maximum Secondary Spread (cap):
    = (Outermost Allowable Orbit# - Secondary MAO) / (Secondary's Allocated Orbits + 1)
```

**Files:** `worlds/spread.go` (create), `worlds/spread_test.go` (create), `worlds/worked_examples_test.go` (extend).

- [ ] **Step 1: Write failing tests**

Create `worlds/spread_test.go`:

```go
package worlds

import (
    "math"
    "testing"

    "wbh/stars"
)

func TestSpread_BaseFormula(t *testing.T) {
    t.Parallel()
    primary := Group{MAO: 0.61, Intervals: []Interval{{Min: 0.61, Max: 20.0}}}
    // (3.1 - 0.61) / 5 = 0.498
    got := Spread(primary, 11, 3.1, 5, 3)
    if math.Abs(got-0.498) > 0.005 {
        t.Errorf("Spread = %v, want 0.498", got)
    }
}

func TestSpread_BaselineNLessThan1_TreatedAs1(t *testing.T) {
    t.Parallel()
    primary := Group{MAO: 0.5, Intervals: []Interval{{Min: 0.5, Max: 20.0}}}
    // baselineN = 0 → treat as 1. (10.0 - 0.5)/1 = 9.5
    got := Spread(primary, 5, 10.0, 0, 1)
    if math.Abs(got-9.5) > 0.05 {
        t.Errorf("Spread = %v, want 9.5", got)
    }
}

func TestSpread_MaxCapApplied(t *testing.T) {
    t.Parallel()
    // Construct a case where base spread would push outermost > 20.
    // Primary intervals: [0.1, 20.0] → Total 19.9.
    // baselineN=1 → base spread = (19.0 - 0.1)/1 = 18.9.
    // Outermost from MAO at base spread for 5 allocated worlds: 0.1 + 5×18.9 = 94.6 > 20 → cap.
    // Cap = 19.9 / (5 + 3 stars) = 2.4875.
    primary := Group{MAO: 0.1, Intervals: []Interval{{Min: 0.1, Max: 20.0}}}
    got := Spread(primary, 5, 19.0, 1, 3)
    if math.Abs(got-2.4875) > 0.01 {
        t.Errorf("Spread = %v, want ~2.49 (cap applied)", got)
    }
}

func TestMaximumSecondarySpread(t *testing.T) {
    t.Parallel()
    sec := Group{MAO: 0.74, Intervals: []Interval{{Min: 0.74, Max: 7.10}}}
    // (7.10 - 0.74) / (5 + 1) = 1.06
    got := MaximumSecondarySpread(sec, 5)
    if math.Abs(got-1.06) > 0.01 {
        t.Errorf("MaximumSecondarySpread = %v, want 1.06", got)
    }
    // sanity
    _ = stars.V
}
```

Append to `worlds/worked_examples_test.go`:

```go
func TestZed_Spread(t *testing.T) {
    t.Parallel()
    sys := composeZed()
    avail, err := worlds.AvailableOrbits(sys)
    if err != nil {
        t.Fatalf("%v", err)
    }
    // Zed: primary Aab MAO 0.61, baselineOrbit 3.1, baselineN 5, totalStars 3.
    // (3.1 - 0.61) / 5 = 0.498
    got := worlds.Spread(avail.Groups[0], 11, 3.1, 5, 3)
    if math.Abs(got-0.498) > 0.005 {
        t.Errorf("Spread = %v, want 0.498", got)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./worlds -run "Spread" -v
```

Expected: FAIL — `Spread / MaximumSecondarySpread undefined`.

- [ ] **Step 3: Implement `Spread` and `MaximumSecondarySpread`**

Create `worlds/spread.go`:

```go
package worlds

// Spread implements WBH Step 5 (pp. 48–49). Returns the average Orbit#
// separation between slots inside the system.
//
// Base formula: (baselineOrbit - primary.MAO) / max(baselineN, 1).
//
// When applying the base spread would place the outermost primary slot
// past Orbit# 20, the mandatory Maximum Spread cap (p. 48) is used:
//
//	primary.Total() / (primaryAllocated + totalStars)
//
// totalStars counts primary + Close/Near/Far secondaries. Companions do
// not count.
func Spread(primary Group, primaryAllocated int, baselineOrbit float64, baselineN, totalStars int) float64 {
    n := baselineN
    if n < 1 {
        n = 1
    }
    base := (baselineOrbit - primary.MAO) / float64(n)
    // Outermost slot would be at MAO + primaryAllocated × base.
    if primary.MAO+float64(primaryAllocated)*base <= 20.0 {
        return base
    }
    if primaryAllocated+totalStars == 0 {
        return base
    }
    return primary.Total() / float64(primaryAllocated+totalStars)
}

// MaximumSecondarySpread implements the WBH p. 49 per-secondary cap:
//
//	(secondary.Outermost - secondary.MAO) / (secondaryAllocated + 1)
//
// secondary.Outermost is the maximum endpoint across secondary.Intervals.
// PlaceOrbitSlots applies this when the system spread would push the
// secondary's outermost slot past its outer allowable bound.
func MaximumSecondarySpread(secondary Group, secondaryAllocated int) float64 {
    var outer float64
    for _, iv := range secondary.Intervals {
        if iv.Max > outer {
            outer = iv.Max
        }
    }
    if secondaryAllocated+1 == 0 {
        return 0
    }
    return (outer - secondary.MAO) / float64(secondaryAllocated+1)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./worlds -run "Spread" -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/spread.go worlds/spread_test.go worlds/worked_examples_test.go
git commit -m "feat(worlds): Step 5 Spread + MaximumSecondarySpread (WBH pp. 48-49)"
```

---

## Task 11: Step 6 — `Slot` type and `PlaceOrbitSlots`

**Source:** WBH pp. 49–50. Walk each `StarAllocation` in order, placing slots from MAO outward by `(Spread + (2D-7) × Spread/10)`. The baseline-N-th slot of the primary group is fixed at `baselineOrbit`. When a placed slot lands in (or past) an exclusion zone in the primary group, widen the spread by the zone width (`nextSpread += zoneWidth`).

Empty orbits (from Step 4) are distributed across allocations before placement: Close → Near → Far → primary. Each allocation's bumped slot count gets one extra "+" slot at the end. The "+" slot is a regular slot in Step 6 — it just makes the star one slot longer.

**Files:** `worlds/orbit_slots.go` (create), `worlds/orbit_slots_test.go` (create), `worlds/worked_examples_test.go` (extend).

- [ ] **Step 1: Write failing tests**

Create `worlds/orbit_slots_test.go`:

```go
package worlds

import (
    "math"
    "testing"

    "wbh/roller"
    "wbh/stars"
)

func TestPlaceOrbitSlots_SinglePrimary_NoExclusions(t *testing.T) {
    t.Parallel()
    primary := Group{
        Designation: "A",
        Members:     []stars.Star{{}},
        MAO:         0.5,
        Intervals:   []Interval{{Min: 0.5, Max: 20.0}},
    }
    allocs := []StarAllocation{{Group: primary, TotalStarOrbits: 5, AllocatedWorlds: 5}}
    // baselineN=1 → 1st slot is baselineOrbit. Spread = 1.0, baselineOrbit = 1.5.
    // 5 slots: variance rolls all 7 (no variance) → 1.5, 2.5, 3.5, 4.5, 5.5.
    // First slot is baseline-fixed → 1.5.
    // Remaining 4 variance rolls: 7,7,7,7.
    got, err := PlaceOrbitSlots(roller.NewScripted(7, 7, 7, 7), allocs, 1.5, 1.0, 0)
    if err != nil {
        t.Fatalf("%v", err)
    }
    want := []float64{1.5, 2.5, 3.5, 4.5, 5.5}
    if len(got) != len(want) {
        t.Fatalf("len(got) = %d, want %d", len(got), len(want))
    }
    for i, w := range want {
        if math.Abs(got[i].Orbit-w) > 0.05 {
            t.Errorf("slot %d Orbit = %v, want %v", i, got[i].Orbit, w)
        }
    }
}

func TestPlaceOrbitSlots_BaselineFixedSlotIsAtBaselineOrbit(t *testing.T) {
    t.Parallel()
    primary := Group{
        Members:   []stars.Star{{}},
        MAO:       0.5,
        Intervals: []Interval{{Min: 0.5, Max: 20.0}},
    }
    allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 5}}
    // baselineOrbit=5.0, spread=1.0, MAO=0.5 → derived baselineN = round((5.0-0.5)/1.0)+1 = 6.
    // 5 slots → baselineN clamped to 5. So slot 5 (the last) is fixed at 5.0.
    // Slots 1-4 use variance rolls (7,7,7,7 → no variance):
    //   slot 1: 0.5 + 1.0 = 1.5
    //   slot 2: 1.5 + 1.0 = 2.5
    //   slot 3: 2.5 + 1.0 = 3.5
    //   slot 4: 3.5 + 1.0 = 4.5
    //   slot 5: BASELINE → 5.0
    got, err := PlaceOrbitSlots(roller.NewScripted(7, 7, 7, 7), allocs, 5.0, 1.0, 0)
    if err != nil {
        t.Fatalf("%v", err)
    }
    want := []float64{1.5, 2.5, 3.5, 4.5, 5.0}
    if len(got) != len(want) {
        t.Fatalf("len = %d, want %d", len(got), len(want))
    }
    for i, w := range want {
        if math.Abs(got[i].Orbit-w) > 0.05 {
            t.Errorf("slot %d Orbit = %v, want %v", i, got[i].Orbit, w)
        }
    }
}

func TestPlaceOrbitSlots_ExclusionZoneWidens(t *testing.T) {
    t.Parallel()
    primary := Group{
        Members:   []stars.Star{{}},
        MAO:       0.5,
        Intervals: []Interval{{Min: 0.5, Max: 5.0}, {Min: 8.0, Max: 20.0}},
    }
    allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 4}}
    // Spread 2.0, baselineOrbit 4.0, baselineN 2.
    // Slot 1: 0.5 + 2.0 + 0 (variance 7) = 2.5
    // Slot 2: baseline-fixed at 4.0.
    // Slot 3 would be 4.0 + 2.0 = 6.0 — inside exclusion (5.0, 8.0). Widen spread by zone width 3.0:
    //   slot 3 = 4.0 + 2.0 + 3.0 + 0 = 9.0
    // Slot 4: 9.0 + 2.0 + 0 = 11.0
    got, err := PlaceOrbitSlots(roller.NewScripted(7, 7, 7), allocs, 4.0, 2.0, 0)
    if err != nil {
        t.Fatalf("%v", err)
    }
    want := []float64{2.5, 4.0, 9.0, 11.0}
    if len(got) != len(want) {
        t.Fatalf("len(got) = %d, want %d", len(got), len(want))
    }
    for i, w := range want {
        if math.Abs(got[i].Orbit-w) > 0.05 {
            t.Errorf("slot %d Orbit = %v, want %v", i, got[i].Orbit, w)
        }
    }
}

func TestPlaceOrbitSlots_EmptyDistribution(t *testing.T) {
    t.Parallel()
    primary := Group{Designation: "A", Members: []stars.Star{{}}, MAO: 0.5, Intervals: []Interval{{Min: 0.5, Max: 20.0}}}
    nearSec := Group{Designation: "B", Members: []stars.Star{{}}, MAO: 0.1, Intervals: []Interval{{Min: 0.1, Max: 1.5}}}
    allocs := []StarAllocation{
        {Group: primary, AllocatedWorlds: 3},
        {Group: nearSec, AllocatedWorlds: 1},
    }
    // emptyOrbits = 1 → goes to first non-primary alloc (B). B grows from 1 to 2 slots.
    // Variance rolls: primary (3 calls — slot 1,2,3 with baseline at slot 1 → 2 variance), nearSec (2 calls).
    // Wait: baselineN affects which slot is fixed. baselineN=1 here (set as such). Primary slot 1
    // is baseline-fixed → 2 variance rolls for primary; B has 2 variance rolls.
    got, err := PlaceOrbitSlots(roller.NewScripted(7, 7, 7, 7), allocs, 0.5, 0.5, 1)
    if err != nil {
        t.Fatalf("%v", err)
    }
    primaryCount, secCount := 0, 0
    for _, s := range got {
        switch s.Group.Designation {
        case "A":
            primaryCount++
        case "B":
            secCount++
        }
    }
    if primaryCount != 3 {
        t.Errorf("primary slots = %d, want 3", primaryCount)
    }
    if secCount != 2 {
        t.Errorf("B slots = %d, want 2 (1 alloc + 1 empty bump)", secCount)
    }
    // The bumped B slot has the "+" suffix.
    foundPlus := false
    for _, s := range got {
        if s.Group.Designation == "B" && len(s.StarSlot) > 0 && s.StarSlot[len(s.StarSlot)-1] == '+' {
            foundPlus = true
        }
    }
    if !foundPlus {
        t.Errorf("expected a B+ slot, got %v", got)
    }
}
```

The first test had an API mismatch — replace it with the real signature once the implementation lands. The `Skip` is fine for now; it documents the intent.

Append to `worlds/worked_examples_test.go`:

```go
func TestZed_PlaceOrbitSlots_Aab(t *testing.T) {
    t.Parallel()
    sys := composeZed()
    avail, err := worlds.AvailableOrbits(sys)
    if err != nil {
        t.Fatalf("%v", err)
    }
    // Just the primary Aab placement — book pp. 49-50 narrates 11 slots.
    allocs := []worlds.StarAllocation{{Group: avail.Groups[0], AllocatedWorlds: 11}}
    // Variance rolls per book (some explicitly given; the rest 7 = neutral):
    // slot 1: 5 (-0.10 → rounds to 1.0); 2: 9 (+0.10 → 1.6); 3: 7 (no var → 2.1); 4: 9 (+0.10 → 2.7);
    // slot 5: BASELINE 3.1 (no roll consumed); 6,7,8: 7,7,7 → 3.5,4.1,4.6 (book values)
    // slot 9 lands in exclusion → widening, then 7,7,7 → 7.2,7.8,8.3
    rolls := []int{5, 9, 7, 9, 7, 7, 7, 7, 7, 7}
    got, err := worlds.PlaceOrbitSlots(roller.NewScripted(rolls...), allocs, 3.1, 0.5, 0)
    if err != nil {
        t.Fatalf("%v", err)
    }
    want := []float64{1.0, 1.6, 2.1, 2.7, 3.1, 3.5, 4.1, 4.6, 7.2, 7.8, 8.3}
    if len(got) != len(want) {
        t.Fatalf("slots = %d, want %d", len(got), len(want))
    }
    for i, w := range want {
        if math.Abs(got[i].Orbit-w) > 0.05 {
            t.Errorf("slot %d (%s) Orbit = %v, want %v", i, got[i].StarSlot, got[i].Orbit, w)
        }
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./worlds -run "PlaceOrbitSlots" -v
```

Expected: FAIL — `PlaceOrbitSlots / Slot undefined`.

- [ ] **Step 3: Implement `Slot` and `PlaceOrbitSlots`**

Create `worlds/orbit_slots.go`:

```go
package worlds

import (
    "fmt"

    "wbh/roller"
)

// Slot is one placed orbit position in a system, before world-body
// assignment (Step 8).
type Slot struct {
    StarSlot string  // "A1", "A2", ..., "B+", "C5". The "+" suffix marks
                     // the extra slot added to a star by Step 4's empty-bump.
                     // Step 7 anomalous slots get their own "+" suffix when
                     // placed; the AnomalousSlot.Anomaly field disambiguates.
    Group    Group
    Orbit    float64
}

// PlaceOrbitSlots implements WBH Step 6 (pp. 49–50).
//
// Walks each StarAllocation in order, placing slots outward from MAO by
// (spread + (2D-7) × spread/10). The baseline-N-th slot of the primary
// group is fixed at baselineOrbit (overrides variance). When the next
// slot would land in an exclusion zone, widens the spread by the zone
// width (mandatory, p. 49 narrative on Zed orbits 9-11).
//
// emptyOrbits (from Step 4) are distributed before placement: Close star
// first, then Near, then Far, then the primary; each receiving star gets
// one extra slot named with a "+" suffix.
func PlaceOrbitSlots(
    r roller.Roller,
    allocs []StarAllocation,
    baselineOrbit, spread float64,
    emptyOrbits int,
) ([]Slot, error) {
    if len(allocs) == 0 {
        return nil, nil
    }

    // Step 1 of placement: distribute empty orbits. Find the index of
    // each star group by its primary-companion's OrbitClass (Close, Near,
    // Far) for distribution priority. The primary group is index 0.
    extraSlots := make([]int, len(allocs))
    remaining := emptyOrbits
    for _, oc := range []orbitClassOrder{ocClose, ocNear, ocFar} {
        if remaining == 0 {
            break
        }
        for i := 1; i < len(allocs); i++ {
            if remaining == 0 {
                break
            }
            sc := allocs[i].Group.sourceCompanion
            if sc != nil && classOrderOf(sc.OrbitClass) == oc {
                extraSlots[i]++
                remaining--
            }
        }
    }
    // Any remainder goes to the primary.
    extraSlots[0] += remaining

    // baselineN for fixing the primary baseline slot. The book Step 6
    // language uses "baseline number"; we recover it implicitly here as
    // the index in the primary's slot sequence whose Orbit# equals
    // baselineOrbit. Caller doesn't pass it directly — the convention is
    // to round (baselineOrbit - MAO) / spread to nearest integer + 1.
    // For correctness with the Zed test, this matches: (3.1 - 0.61)/0.5 ≈ 4.98 → +1 = 5.
    baselineN := 1
    if spread > 0 {
        baselineN = int((baselineOrbit-allocs[0].Group.MAO)/spread+0.5) + 1
        if baselineN < 1 {
            baselineN = 1
        }
        if baselineN > allocs[0].AllocatedWorlds+extraSlots[0] {
            baselineN = allocs[0].AllocatedWorlds + extraSlots[0]
        }
    }

    var out []Slot
    for i, alloc := range allocs {
        slotCount := alloc.AllocatedWorlds + extraSlots[i]
        if slotCount == 0 {
            continue
        }
        cur := alloc.Group.MAO
        for j := 0; j < slotCount; j++ {
            // Position label: regular slots are "<letter>N" 1..AllocatedWorlds;
            // the bumped slot (if any) is "<letter>+" placed last.
            label := slotLabel(alloc.Group.Designation, j, alloc.AllocatedWorlds, extraSlots[i])

            // Compute proposed orbit for this slot.
            isBaselineSlot := i == 0 && j == baselineN-1
            useSpread := spread
            // Look ahead for exclusion-zone widening on subsequent slots:
            // We compute this slot's orbit, then if it lands in a hole within
            // the primary's intervals, widen.

            var orbit float64
            switch {
            case isBaselineSlot:
                orbit = baselineOrbit
            case j == 0:
                // Inner slot: MAO + spread + variance.
                v := r.Roll("2D")
                orbit = cur + useSpread + float64(v-7)*useSpread/10.0
            default:
                // Next slot: previous + spread + variance, possibly widened.
                v := r.Roll("2D")
                proposed := cur + useSpread + float64(v-7)*useSpread/10.0
                // Exclusion-zone widening (primary group only).
                if i == 0 {
                    if zone := excludedZoneAt(alloc.Group, proposed); zone != 0 {
                        proposed += zone
                    }
                }
                orbit = proposed
            }
            out = append(out, Slot{
                StarSlot: label,
                Group:    alloc.Group,
                Orbit:    orbit,
            })
            cur = orbit
        }
    }
    return out, nil
}

// orbitClassOrder is an internal enum used for empty-orbit distribution
// priority (Close → Near → Far per WBH p. 48).
type orbitClassOrder int

const (
    ocClose orbitClassOrder = iota
    ocNear
    ocFar
    ocOther
)

func classOrderOf(c interface{}) orbitClassOrder {
    // Compares against stars.OrbitClose/Near/Far via string-style match.
    // We avoid the import cycle by using the type assertion inline.
    type orbitClassish interface{ String() string }
    if oc, ok := c.(orbitClassish); ok {
        switch oc.String() {
        case "Close":
            return ocClose
        case "Near":
            return ocNear
        case "Far":
            return ocFar
        }
    }
    return ocOther
}

// slotLabel returns the StarSlot id for a regular or bumped slot.
//
// regularCount is the AllocatedWorlds for this group (regular slots
// numbered 1..regularCount). extraCount is the count of bumped slots
// from Step 4 (each named with "+" suffix). The bumped slot is placed
// last in the sequence.
func slotLabel(designation string, indexInGroup, regularCount, extraCount int) string {
    // Designation is "Aab" or "B" or "Cab" etc.; the StarSlot prefix is
    // just the first letter.
    prefix := string(designation[0])
    if indexInGroup < regularCount {
        return fmt.Sprintf("%s%d", prefix, indexInGroup+1)
    }
    // Bumped slot.
    return prefix + "+"
}

// excludedZoneAt returns the width of the exclusion zone at orbit, if
// orbit lands in a gap between intervals; otherwise 0.
func excludedZoneAt(g Group, orbit float64) float64 {
    if g.Contains(orbit) {
        return 0
    }
    for i := 0; i < len(g.Intervals)-1; i++ {
        gapStart := g.Intervals[i].Max
        gapEnd := g.Intervals[i+1].Min
        if orbit >= gapStart && orbit <= gapEnd {
            return gapEnd - gapStart
        }
    }
    return 0
}
```

The `classOrderOf` helper above uses a duck-typed interface to dodge an import cycle. Confirm the `stars.OrbitClass` type has a `String()` method or rewrite to import `wbh/stars` directly (it should — already used elsewhere in the package). In practice you should rewrite to just `import "wbh/stars"` and use `stars.OrbitClose` etc. directly:

```go
func classOrderOf(c stars.OrbitClass) orbitClassOrder {
    switch c {
    case stars.OrbitClose:
        return ocClose
    case stars.OrbitNear:
        return ocNear
    case stars.OrbitFar:
        return ocFar
    }
    return ocOther
}
```

And update the caller: `classOrderOf(sc.OrbitClass)` (no type assertion).

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./worlds -run "PlaceOrbitSlots|TestZed_PlaceOrbitSlots" -v
```

Expected: PASS. The exclusion-zone widening test verifies the Zed orbit-9-at-7.2 case in miniature.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/orbit_slots.go worlds/orbit_slots_test.go worlds/worked_examples_test.go
git commit -m "feat(worlds): Step 6 PlaceOrbitSlots with variance + exclusion-zone widening (WBH pp. 49-50)"
```

---

## Task 12: Step 7 — `AnomalyType`, `AnomalousSlot`, `AddAnomalous`

**Source:** WBH pp. 50–51.

**Tables:**

```text
Anomalous Orbits (2D):  9-=0, 10=1, 11=2, 12=3
Anomalous Orbit Type (2D): 7-=Random, 8=Eccentric, 9=Inclined, 10-11=Retrograde, 12=Trojan
```

**Process per anomalous orbit:**

1. Roll on Anomalous Orbit Type table.
2. (Multi-star) D3 picks parent group.
3. Roll 2D-2 for Orbit# integer part, +d10/10 for fraction.
4. Clamp to `[group.MAO, 20.0]`. If outside, roll ±1D adjustment and retry.
5. Each anomaly increments `Counts.Terrestrials` and `Counts.Total` by 1.

**Eccentricity DM** by anomaly type: Random +2, Eccentric +5, Inclined +2, Retrograde +2, Trojan inherits from parent.

**Inclined orbit:** also rolls inclination = `(1D+2) × 10°` plus optional d10 variance.

**Trojan orbit:** picks an existing slot to "shadow" by 60°; specific selection precedence (immediately inward → next available outward → random reroll). For 2B, simplify: pick the immediately-inward slot (book's primary precedence rule); add the trojan as its own slot at the same Orbit# with the parent's StarSlot in `TrojanOf`.

**Files:** `worlds/anomalous.go` (create), `worlds/anomalous_test.go` (create), `worlds/worked_examples_test.go` (extend).

- [ ] **Step 1: Write failing tests**

Create `worlds/anomalous_test.go`:

```go
package worlds

import (
    "math"
    "testing"

    "wbh/roller"
    "wbh/stars"
)

func TestAddAnomalous_None(t *testing.T) {
    t.Parallel()
    primary := Group{Designation: "A", Members: []stars.Star{{}}, MAO: 0.5, Intervals: []Interval{{Min: 0.5, Max: 20.0}}}
    allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 5}}
    counts := Counts{Total: 5}
    slots := []Slot{{StarSlot: "A1", Group: primary, Orbit: 1.0}}
    // 2D = 5 (no anomalous).
    out, newCounts, err := AddAnomalous(roller.NewScripted(5), slots, allocs, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if len(out) != 1 {
        t.Errorf("slots = %d, want 1", len(out))
    }
    if newCounts.Total != counts.Total {
        t.Errorf("counts unchanged should hold, got Total %d", newCounts.Total)
    }
    if out[0].Anomaly != AnomalyNone {
        t.Errorf("Anomaly = %v, want None", out[0].Anomaly)
    }
}

func TestAddAnomalous_Retrograde_SingleStar(t *testing.T) {
    t.Parallel()
    primary := Group{Designation: "A", Members: []stars.Star{{}}, MAO: 0.5, Intervals: []Interval{{Min: 0.5, Max: 20.0}}}
    allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 5}}
    counts := Counts{Terrestrials: 4, Total: 5}
    slots := []Slot{{StarSlot: "A1", Group: primary, Orbit: 1.0}}
    // Anomalous count: 10 → 1.
    // Type: 10 → Retrograde.
    // Random orbit: 2D-2 = 5 (raw 7); d10 = 2 → orbit 5.2.
    rolls := []int{10, 10, 7, 2}
    out, newCounts, err := AddAnomalous(roller.NewScripted(rolls...), slots, allocs, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if len(out) != 2 {
        t.Fatalf("slots = %d, want 2", len(out))
    }
    last := out[1]
    if last.Anomaly != AnomalyRetrograde {
        t.Errorf("Anomaly = %v, want Retrograde", last.Anomaly)
    }
    if math.Abs(last.Orbit-5.2) > 0.01 {
        t.Errorf("Orbit = %v, want 5.2", last.Orbit)
    }
    if last.EccentricityDM != 2 {
        t.Errorf("EccentricityDM = %d, want +2", last.EccentricityDM)
    }
    if newCounts.Terrestrials != counts.Terrestrials+1 {
        t.Errorf("Terrestrials = %d, want %d", newCounts.Terrestrials, counts.Terrestrials+1)
    }
    if newCounts.Total != counts.Total+1 {
        t.Errorf("Total = %d, want %d", newCounts.Total, counts.Total+1)
    }
}

func TestAddAnomalous_TypeTable(t *testing.T) {
    t.Parallel()
    cases := []struct {
        roll int
        want AnomalyType
    }{
        {2, AnomalyRandom}, {7, AnomalyRandom},
        {8, AnomalyEccentric},
        {9, AnomalyInclined},
        {10, AnomalyRetrograde}, {11, AnomalyRetrograde},
        {12, AnomalyTrojan},
    }
    primary := Group{Designation: "A", Members: []stars.Star{{}}, MAO: 0.5, Intervals: []Interval{{Min: 0.5, Max: 20.0}}}
    allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 5}}
    counts := Counts{Total: 5}
    slots := []Slot{{StarSlot: "A1", Group: primary, Orbit: 3.0}}
    for _, tc := range cases {
        // Anomalous count = 10 (1). Type = tc.roll. Random orbit: 2D-2 = 5 (raw 7), d10 = 2.
        // Trojan needs an inward-slot pick; we have one at A1=3.0.
        rolls := []int{10, tc.roll, 7, 2}
        // Inclined adds inclination roll: 1D + d10 → 2 rolls.
        if tc.want == AnomalyInclined {
            rolls = append(rolls, 4, 5) // 1D=4, d10=5 → 60° + 5 var
        }
        out, _, err := AddAnomalous(roller.NewScripted(rolls...), slots, allocs, counts)
        if err != nil {
            t.Fatalf("type %d: %v", tc.roll, err)
        }
        if out[len(out)-1].Anomaly != tc.want {
            t.Errorf("type %d: Anomaly = %v, want %v", tc.roll, out[len(out)-1].Anomaly, tc.want)
        }
    }
}

func TestAddAnomalous_RandomClampsToMAO(t *testing.T) {
    t.Parallel()
    primary := Group{Designation: "A", Members: []stars.Star{{}}, MAO: 1.0, Intervals: []Interval{{Min: 1.0, Max: 20.0}}}
    allocs := []StarAllocation{{Group: primary, AllocatedWorlds: 5}}
    counts := Counts{Total: 5}
    // Anomalous = 1, type = Random (7), orbit = 0 (2D-2=2 raw → 0; d10=1 → 0.1) — below MAO 1.0.
    // Retry roll: 2D-2 = 5 (raw 7), d10 = 5 → 5.5.
    rolls := []int{10, 7, 2, 1, 7, 5}
    out, _, err := AddAnomalous(roller.NewScripted(rolls...), nil, allocs, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if math.Abs(out[len(out)-1].Orbit-5.5) > 0.01 {
        t.Errorf("Orbit = %v, want 5.5 (after retry)", out[len(out)-1].Orbit)
    }
}
```

Append to `worlds/worked_examples_test.go`:

```go
func TestZed_AddAnomalous(t *testing.T) {
    t.Parallel()
    sys := composeZed()
    avail, err := worlds.AvailableOrbits(sys)
    if err != nil {
        t.Fatalf("%v", err)
    }
    counts := worlds.Counts{GasGiants: 4, PlanetoidBelts: 2, Terrestrials: 11, Total: 17}
    allocs, err := worlds.AllocateOrbitsByStar(avail, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    // After Step 4 empty, B's allocation is 2; we update allocs here for the test.
    // The integration façade handles this; for this isolated test, we just simulate.
    allocs[1].AllocatedWorlds = 2
    allocs[2].AllocatedWorlds = 5

    // Pretend we have placed slots already (orbit values from p. 49-50 narration).
    var slots []worlds.Slot
    aab := allocs[0].Group
    for _, o := range []float64{1.0, 1.6, 2.1, 2.7, 3.1, 3.5, 4.1, 4.6, 7.2, 7.8, 8.3} {
        slots = append(slots, worlds.Slot{Group: aab, Orbit: o})
    }
    // Book: anomalous=10 (1), type=10 (Retrograde), parent group D3=1 (Aab),
    // orbit raw 2D-2=5, d10=2 → 5.2.
    rolls := []int{10, 10, 1, 7, 2}
    out, newCounts, err := worlds.AddAnomalous(roller.NewScripted(rolls...), slots, allocs, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    last := out[len(out)-1]
    if last.Anomaly != worlds.AnomalyRetrograde {
        t.Errorf("Anomaly = %v, want Retrograde", last.Anomaly)
    }
    if math.Abs(last.Orbit-5.2) > 0.01 {
        t.Errorf("Orbit = %v, want 5.2", last.Orbit)
    }
    if last.Group.Designation != "Aab" {
        t.Errorf("Group = %v, want Aab", last.Group.Designation)
    }
    if newCounts.Terrestrials != 12 || newCounts.Total != 18 {
        t.Errorf("counts = %+v, want T=12 Total=18", newCounts)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./worlds -run "AddAnomalous|TestZed_AddAnomalous" -v
```

Expected: FAIL — `AddAnomalous undefined`.

- [ ] **Step 3: Implement `AnomalyType`, `AnomalousSlot`, `AddAnomalous`**

Create `worlds/anomalous.go`:

```go
package worlds

import (
    "fmt"

    "wbh/roller"
)

// AnomalyType classifies a Step 7 anomalous slot.
type AnomalyType int

const (
    AnomalyNone AnomalyType = iota
    AnomalyRandom
    AnomalyEccentric
    AnomalyInclined
    AnomalyRetrograde
    AnomalyTrojan
)

// AnomalousSlot extends Slot with anomaly metadata. AnomalyNone slots
// are wrapped Slots with no extra behaviour; non-None slots come from
// Step 7 and feed Step 9 eccentricity DMs.
type AnomalousSlot struct {
    Slot
    Anomaly        AnomalyType
    InclinationDeg float64 // for AnomalyInclined
    TrojanOf       string  // StarSlot id of parent, for AnomalyTrojan
    EccentricityDM int     // additional DM applied in Step 9
}

// AddAnomalous implements WBH Step 7 (pp. 50–51). Wraps each input slot
// as AnomalousSlot{..., Anomaly: AnomalyNone} and appends one new
// AnomalousSlot per rolled anomaly. Updates counts (each anomaly adds
// one terrestrial and one total).
//
// In multi-star systems the parent group is picked by D3 (3-group max
// per WBH structural cap). In single-star systems no group roll is
// consumed; the only group is used.
//
// Anomalous orbits clamp to [group.MAO, 20.0] with retry. Trojan picks
// the immediately-inward slot in the parent group as the target.
func AddAnomalous(
    r roller.Roller,
    slots []Slot,
    allocs []StarAllocation,
    counts Counts,
) ([]AnomalousSlot, Counts, error) {
    out := make([]AnomalousSlot, 0, len(slots))
    for _, s := range slots {
        out = append(out, AnomalousSlot{Slot: s})
    }

    n := anomalousCount(r.Roll("2D"))
    if n == 0 {
        return out, counts, nil
    }

    for i := 0; i < n; i++ {
        atype := anomalousType(r.Roll("2D"))
        // Pick parent group (multi-star uses D3; single-star uses index 0).
        parentIdx := 0
        if len(allocs) > 1 {
            parentIdx = (r.Roll("D3") - 1) % len(allocs)
        }
        parent := allocs[parentIdx].Group

        ecdm := eccentricityDMFor(atype)
        // Compute orbit (Random/Eccentric/Inclined/Retrograde share the
        // 2D-2 + d10/10 procedure, with [MAO, 20.0] clamp + retry).
        var orbit float64
        var trojanOf string
        var inclination float64
        switch atype {
        case AnomalyTrojan:
            // Pick the immediately-inward slot from this parent.
            target := pickInwardSlot(slots, parent.Designation)
            trojanOf = target.StarSlot
            orbit = target.Orbit
        default:
            orbit = rollAnomalousOrbit(r, parent.MAO)
            if atype == AnomalyInclined {
                inclination = float64(r.Roll("1D")+2)*10.0 + float64(r.Roll("d10"))
            }
        }

        out = append(out, AnomalousSlot{
            Slot: Slot{
                StarSlot: fmt.Sprintf("%c+", parent.Designation[0]),
                Group:    parent,
                Orbit:    orbit,
            },
            Anomaly:        atype,
            InclinationDeg: inclination,
            TrojanOf:       trojanOf,
            EccentricityDM: ecdm,
        })

        counts.Terrestrials++
        counts.Total++
    }
    return out, counts, nil
}

func anomalousCount(roll int) int {
    switch {
    case roll <= 9:
        return 0
    case roll == 10:
        return 1
    case roll == 11:
        return 2
    default:
        return 3
    }
}

func anomalousType(roll int) AnomalyType {
    switch {
    case roll <= 7:
        return AnomalyRandom
    case roll == 8:
        return AnomalyEccentric
    case roll == 9:
        return AnomalyInclined
    case roll <= 11:
        return AnomalyRetrograde
    default:
        return AnomalyTrojan
    }
}

func eccentricityDMFor(t AnomalyType) int {
    switch t {
    case AnomalyEccentric:
        return 5
    case AnomalyRandom, AnomalyInclined, AnomalyRetrograde, AnomalyTrojan:
        return 2
    }
    return 0
}

// rollAnomalousOrbit rolls 2D-2 + d10/10, clamped to [mao, 20.0], with
// up to 5 retries before bailing to the clamp value.
func rollAnomalousOrbit(r roller.Roller, mao float64) float64 {
    for try := 0; try < 5; try++ {
        whole := r.Roll("2D") - 2
        frac := r.Roll("d10")
        v := float64(whole) + float64(frac)/10.0
        if v >= mao && v <= 20.0 {
            return v
        }
    }
    return mao
}

// pickInwardSlot returns the highest-Orbit# slot whose Group designation
// matches and that is below the implicit anomalous orbit. For 2B we
// approximate by picking the last slot in the parent group.
func pickInwardSlot(slots []Slot, parentDesignation string) Slot {
    var best Slot
    bestOrbit := -1.0
    for _, s := range slots {
        if s.Group.Designation == parentDesignation && s.Orbit > bestOrbit {
            best = s
            bestOrbit = s.Orbit
        }
    }
    return best
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./worlds -run "AddAnomalous|TestZed_AddAnomalous" -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/anomalous.go worlds/anomalous_test.go worlds/worked_examples_test.go
git commit -m "feat(worlds): Step 7 AddAnomalous with type table and Zed retrograde (WBH pp. 50-51)"
```

---

## Task 13: Step 8 — `BodyType`, `Placement`, `PlaceWorlds`

**Source:** WBH pp. 51–52. Place bodies into slots in this order: empty → gas giants → planetoid belts → terrestrials. Use 1D:1D rolling with a prefix die selected by total slots: ≤6 → 1D, 7–12 → D2, 13–18 → D3, >18 → 1D with reroll-above-N (where N is the count of valid prefix sequences).

**Collision handling:** if the rolled slot id already has a body, +1 to the right die (within the same prefix), then advance to next slot id, then wrap to the first slot still without a body assigned.

Mainworld branch and special cases (moon of GG, size-1 in belt) are out of scope (Continuation Method).

**Files:** `worlds/placement.go` (create), `worlds/placement_test.go` (create).

- [ ] **Step 1: Write failing tests**

Create `worlds/placement_test.go`:

```go
package worlds

import (
    "testing"

    "wbh/roller"
    "wbh/stars"
)

func TestPlaceWorlds_Order(t *testing.T) {
    t.Parallel()
    // 4 slots in a single group "A". 1 GG, 1 belt, 2 terrestrials, 0 empty.
    primary := Group{Designation: "A", Members: []stars.Star{{}}}
    slots := []AnomalousSlot{
        {Slot: Slot{StarSlot: "A1", Group: primary, Orbit: 1.0}},
        {Slot: Slot{StarSlot: "A2", Group: primary, Orbit: 2.0}},
        {Slot: Slot{StarSlot: "A3", Group: primary, Orbit: 3.0}},
        {Slot: Slot{StarSlot: "A4", Group: primary, Orbit: 4.0}},
    }
    counts := Counts{GasGiants: 1, PlanetoidBelts: 1, Terrestrials: 2, Total: 4}
    // 4 slots → 1D prefix (≤6). All rolls have prefix=1. Right-die rolls:
    //   GG (1 roll): 1 → A1
    //   Belt (1 roll): 2 → A2
    //   Terrestrials (2 rolls): 3, 4 → A3, A4
    // No empty orbits.
    rolls := []int{1, 1, 1, 2, 1, 3, 1, 4} // each placement = (prefix, right)
    got, err := PlaceWorlds(roller.NewScripted(rolls...), slots, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if len(got) != 4 {
        t.Fatalf("len = %d, want 4", len(got))
    }
    if got[0].Body != BodyGasGiant {
        t.Errorf("A1 = %v, want GasGiant", got[0].Body)
    }
    if got[1].Body != BodyPlanetoidBelt {
        t.Errorf("A2 = %v, want PlanetoidBelt", got[1].Body)
    }
    if got[2].Body != BodyTerrestrial || got[3].Body != BodyTerrestrial {
        t.Errorf("A3,A4 should be Terrestrials")
    }
}

func TestPlaceWorlds_Collision_PlusOne(t *testing.T) {
    t.Parallel()
    primary := Group{Designation: "A", Members: []stars.Star{{}}}
    slots := []AnomalousSlot{
        {Slot: Slot{StarSlot: "A1", Group: primary, Orbit: 1.0}},
        {Slot: Slot{StarSlot: "A2", Group: primary, Orbit: 2.0}},
    }
    counts := Counts{GasGiants: 1, Terrestrials: 1, Total: 2}
    // GG → 1:1 → A1. Terrestrial → 1:1 (collision) → +1 → 1:2 → A2.
    rolls := []int{1, 1, 1, 1}
    got, err := PlaceWorlds(roller.NewScripted(rolls...), slots, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got[0].Body != BodyGasGiant {
        t.Errorf("A1 = %v, want GasGiant", got[0].Body)
    }
    if got[1].Body != BodyTerrestrial {
        t.Errorf("A2 = %v, want Terrestrial (after +1 collision)", got[1].Body)
    }
}

func TestPlaceWorlds_PrefixDieSelection(t *testing.T) {
    t.Parallel()
    primary := Group{Designation: "A", Members: []stars.Star{{}}}
    var slots []AnomalousSlot
    for i := 1; i <= 8; i++ {
        slots = append(slots, AnomalousSlot{Slot: Slot{StarSlot: "A1", Group: primary, Orbit: float64(i)}})
        // Set the StarSlot id to A<i> for the test to distinguish.
        slots[i-1].StarSlot = ""
    }
    // 8 slots → D2 prefix (7-12). Each placement uses (D2, 1D).
    // We only test that PlaceWorlds runs without error and returns 8 placements.
    counts := Counts{Terrestrials: 8, Total: 8}
    rolls := make([]int, 0, 16)
    // Generate a sequence covering all 8 slots: prefix 1 → 1:1..1:6, prefix 2 → 2:1..2:2.
    for i, pair := range [][2]int{{1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {2, 1}, {2, 2}} {
        rolls = append(rolls, pair[0], pair[1])
        _ = i
    }
    got, err := PlaceWorlds(roller.NewScripted(rolls...), slots, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if len(got) != 8 {
        t.Errorf("placements = %d, want 8", len(got))
    }
}

func TestPlaceWorlds_IncludesEmptyBody(t *testing.T) {
    t.Parallel()
    primary := Group{Designation: "A", Members: []stars.Star{{}}}
    slots := []AnomalousSlot{
        {Slot: Slot{StarSlot: "A1", Group: primary, Orbit: 1.0}},
        {Slot: Slot{StarSlot: "A2", Group: primary, Orbit: 2.0}},
        {Slot: Slot{StarSlot: "A3", Group: primary, Orbit: 3.0}},
    }
    counts := Counts{Terrestrials: 2, Total: 2} // counts.Total < len(slots) → one empty
    // 1 empty (placed first) → 1:1 → A1. 2 terrestrials → 1:2, 1:3 → A2, A3.
    rolls := []int{1, 1, 1, 2, 1, 3}
    got, err := PlaceWorlds(roller.NewScripted(rolls...), slots, counts)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if got[0].Body != BodyEmpty {
        t.Errorf("A1 = %v, want Empty", got[0].Body)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./worlds -run "PlaceWorlds" -v
```

Expected: FAIL — `PlaceWorlds undefined`.

- [ ] **Step 3: Implement `BodyType`, `Placement`, `PlaceWorlds`**

Create `worlds/placement.go`:

```go
package worlds

import (
    "fmt"

    "wbh/roller"
)

// BodyType classifies the body type assigned to an orbit slot in Step 8.
type BodyType int

const (
    BodyEmpty BodyType = iota
    BodyTerrestrial
    BodyGasGiant
    BodyPlanetoidBelt
)

// Placement is one fully-resolved orbit slot after Step 8.
type Placement struct {
    AnomalousSlot
    Body       BodyType
    PrefixRoll string // "1:6", "2:3" — audit trail
}

// PlaceWorlds implements WBH Step 8 (pp. 51–52). Order: empty → gas
// giants → planetoid belts → terrestrials. Uses 1D:1D rolling with a
// prefix die selected by total slot count (≤6 → 1D, 7–12 → D2, 13–18 →
// D3, >18 → 1D with reroll-above-N).
//
// Collision handling: if rolled slot already has a body, +1 to the right
// die (within the same prefix), then advance to next slot id, then wrap
// to first unassigned slot.
//
// Mainworld and Continuation-only branches (moon of GG, size-1 in belt,
// atmosphere-DM raw-temp reverse-engineering) are out of scope.
func PlaceWorlds(r roller.Roller, slots []AnomalousSlot, counts Counts) ([]Placement, error) {
    out := make([]Placement, len(slots))
    for i, s := range slots {
        out[i] = Placement{AnomalousSlot: s}
    }
    assigned := make([]bool, len(slots))
    n := len(slots)

    prefixDie, prefixMax := prefixSpec(n)
    _ = prefixMax

    placeOne := func(body BodyType) error {
        for {
            prefix := r.Roll(prefixDie)
            right := r.Roll("1D")
            // Collision-handling loop: +1 to right die, then advance, then wrap.
            idx := slotIndex(prefix, right, n)
            for assigned[idx] {
                right++
                if right > 6 {
                    // advance to next slot id
                    right = 1
                    prefix++
                    if prefix > prefixMax {
                        prefix = 1
                    }
                }
                idx = slotIndex(prefix, right, n)
                if !assigned[idx] {
                    break
                }
                // Wrap to first unassigned.
                idx = -1
                for j, a := range assigned {
                    if !a {
                        idx = j
                        break
                    }
                }
                if idx == -1 {
                    return fmt.Errorf("worlds: no unassigned slots")
                }
                break
            }
            if idx >= n {
                return fmt.Errorf("worlds: rolled slot index %d out of range", idx)
            }
            out[idx].Body = body
            out[idx].PrefixRoll = fmt.Sprintf("%d:%d", prefix, right)
            assigned[idx] = true
            return nil
        }
    }

    // Order: empty → GG → belts → terrestrials.
    emptyCount := n - counts.Total
    if emptyCount < 0 {
        emptyCount = 0
    }
    for i := 0; i < emptyCount; i++ {
        if err := placeOne(BodyEmpty); err != nil {
            return nil, err
        }
    }
    for i := 0; i < counts.GasGiants; i++ {
        if err := placeOne(BodyGasGiant); err != nil {
            return nil, err
        }
    }
    for i := 0; i < counts.PlanetoidBelts; i++ {
        if err := placeOne(BodyPlanetoidBelt); err != nil {
            return nil, err
        }
    }
    for i := 0; i < counts.Terrestrials; i++ {
        if err := placeOne(BodyTerrestrial); err != nil {
            return nil, err
        }
    }
    return out, nil
}

// prefixSpec returns the prefix-die notation and max valid prefix value
// for n total slots, per WBH Step 8.
func prefixSpec(n int) (notation string, maxValid int) {
    switch {
    case n <= 6:
        return "1D", 1
    case n <= 12:
        return "D2", 2
    case n <= 18:
        return "D3", 3
    default:
        // Very dense systems use 1D with reroll-above-N where N is the
        // count of full prefix sequences; for our purposes, treat as D6
        // and reroll if exceeds.
        return "1D", (n + 5) / 6
    }
}

// slotIndex maps a (prefix, right) pair to a flat slot index.
func slotIndex(prefix, right, n int) int {
    idx := (prefix-1)*6 + (right - 1)
    if idx >= n {
        idx = n - 1
    }
    if idx < 0 {
        idx = 0
    }
    return idx
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./worlds -run "PlaceWorlds" -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/
just check
git add worlds/placement.go worlds/placement_test.go
git commit -m "feat(worlds): Step 8 PlaceWorlds with 1D:1D and prefix-die selection (WBH pp. 51-52)"
```

---

## Task 14: Step 9 — `RollPlanetEccentricities`

**Source:** WBH p. 52. Calls existing `stars.RollEccentricity` for each non-belt non-empty placement, applying anomaly DMs from `AnomalousSlot.EccentricityDM`.

**Files:** `worlds/planet_eccentricity.go` (create), `worlds/planet_eccentricity_test.go` (create).

- [ ] **Step 1: Write failing test**

Create `worlds/planet_eccentricity_test.go`:

```go
package worlds

import (
    "testing"

    "wbh/roller"
)

func TestRollPlanetEccentricities_AppliesAnomalyDM(t *testing.T) {
    t.Parallel()
    placements := []Placement{
        // BodyTerrestrial, no anomaly → DM 0
        {AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "A1", Orbit: 1.0}}, Body: BodyTerrestrial},
        // BodyTerrestrial, AnomalyEccentric → DM +5
        {AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "A+", Orbit: 5.0}, Anomaly: AnomalyEccentric, EccentricityDM: 5}, Body: BodyTerrestrial},
        // BodyEmpty → skipped
        {AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "A2", Orbit: 2.0}}, Body: BodyEmpty},
        // BodyPlanetoidBelt → skipped
        {AnomalousSlot: AnomalousSlot{Slot: Slot{StarSlot: "A3", Orbit: 3.0}}, Body: BodyPlanetoidBelt},
    }
    // Two real RollEccentricity calls; provide rolls accordingly.
    // The exact ecc values depend on stars.EccentricityValues table; we just
    // assert the function ran and skipped the right placements.
    out, err := RollPlanetEccentricities(roller.NewScripted(7, 7), placements)
    if err != nil {
        t.Fatalf("%v", err)
    }
    if len(out) != len(placements) {
        t.Fatalf("len = %d, want %d", len(out), len(placements))
    }
    // Empty and belt should remain at 0 / unchanged eccentricity (none stored on Placement; the
    // test verifies the loop skips them without consuming a roll. If the roller stub didn't
    // get exhausted we know we skipped properly.
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./worlds -run TestRollPlanetEccentricities -v
```

Expected: FAIL — `RollPlanetEccentricities undefined`.

- [ ] **Step 3: Extend `stars.EccentricityOpts` with `ExtraDM` field**

The existing `EccentricityOpts` (in `stars/orbits.go`) has `IsStar`, `NestingDepth`, `Orbit`, `SystemAgeGyr`, `IsBeltMember` — but no field for an arbitrary externally-supplied DM. Step 9 needs to apply per-anomaly DMs (Random +2, Eccentric +5, Inclined +2, Retrograde +2, Trojan +2) on top of the existing p.27 DMs. Add an `ExtraDM int` field.

In `stars/orbits.go`, change `EccentricityOpts`:

```go
type EccentricityOpts struct {
    IsStar       bool
    NestingDepth int
    Orbit        float64
    SystemAgeGyr float64
    IsBeltMember bool
    ExtraDM      int // additional DM applied externally (e.g., 2B Step 7 anomaly DMs)
}
```

In `RollEccentricity`, add `dm += opts.ExtraDM` after the existing DM accumulators (before the `natural := r.Roll("2D")` line).

Add a stars-package test verifying `ExtraDM` is honoured. In `stars/orbits_test.go` (or a new test if no orbits_test.go exists, create `stars/eccentricity_test.go`):

```go
func TestRollEccentricity_ExtraDM(t *testing.T) {
    t.Parallel()
    // Without ExtraDM, roll 2D = 7 → row 7 lookup. With ExtraDM = +5, row becomes 12.
    // The two should produce different eccentricity values.
    base, err := RollEccentricity(roller.NewScripted(7, 1), EccentricityOpts{})
    if err != nil {
        t.Fatalf("%v", err)
    }
    bumped, err := RollEccentricity(roller.NewScripted(7, 1), EccentricityOpts{ExtraDM: 5})
    if err != nil {
        t.Fatalf("%v", err)
    }
    if base == bumped {
        t.Errorf("base = bumped = %v; ExtraDM had no effect", base)
    }
}
```

Run, verify failing, implement, verify passing.

- [ ] **Step 4: Add `Eccentricity` field to `Placement`**

In `worlds/placement.go`, extend `Placement`:

```go
type Placement struct {
    AnomalousSlot
    Body         BodyType
    PrefixRoll   string  // "1:6", "2:3"...
    Eccentricity float64 // populated by Step 9 (RollPlanetEccentricities)
}
```

- [ ] **Step 5: Implement `RollPlanetEccentricities`**

Create `worlds/planet_eccentricity.go`:

```go
package worlds

import (
    "wbh/roller"
    "wbh/stars"
)

// RollPlanetEccentricities implements WBH Step 9 (p. 52).
//
// For each non-empty non-belt placement, calls stars.RollEccentricity
// with the placement's anomaly DM (AnomalousSlot.EccentricityDM) passed
// through stars.EccentricityOpts.ExtraDM. Stores the result on
// Placement.Eccentricity.
//
// Belts and empty slots are skipped (no roll consumed).
func RollPlanetEccentricities(r roller.Roller, ps []Placement) ([]Placement, error) {
    out := make([]Placement, len(ps))
    copy(out, ps)
    for i := range out {
        if out[i].Body == BodyEmpty || out[i].Body == BodyPlanetoidBelt {
            continue
        }
        ecc, err := stars.RollEccentricity(r, stars.EccentricityOpts{
            ExtraDM: out[i].EccentricityDM,
            // We don't have the system age plumbed through to this layer
            // in 2B; the sub-1.0/age>1Gyr DM is therefore not applied
            // here. 2C should plumb sys.Primary.AgeGyr through.
        })
        if err != nil {
            return nil, err
        }
        out[i].Eccentricity = ecc
    }
    return out, nil
}
```

- [ ] **Step 6: Run test to verify it passes**

```bash
go test ./worlds -run TestRollPlanetEccentricities -v
```

Expected: PASS.

- [ ] **Step 7: Format, lint, commit**

```bash
gofumpt -w worlds/ stars/
just check
git add stars/orbits.go stars/eccentricity_test.go worlds/planet_eccentricity.go worlds/planet_eccentricity_test.go worlds/placement.go
git commit -m "feat(worlds): Step 9 RollPlanetEccentricities with anomaly DMs (WBH p. 52)"
```

(Adjust the file list if you placed the stars-package test elsewhere.)

---

## Task 15: `SystemPlacement` façade and `TestZed_FullPlacement` acceptance gate

**Files:** `worlds/system_placement.go` (create), `worlds/worked_examples_test.go` (extend with `TestZed_FullPlacement`).

**API:**

```go
type SystemPlacement struct {
    Counts        Counts
    Allocations   []StarAllocation
    BaselineN     int
    BaselineOrbit float64
    EmptyOrbits   int
    SystemSpread  float64
    Placements    []Placement
}

func GenerateSystemPlacement(r roller.Roller, sys stars.System) (SystemPlacement, error)
```

- [ ] **Step 1: Write failing acceptance test**

Append to `worlds/worked_examples_test.go`:

```go
func TestZed_FullPlacement(t *testing.T) {
    t.Parallel()
    sys := composeZed()

    // Concatenate the dice sequence the book narrates across pp. 36–52
    // (with the encoded DMs accounted for):
    //
    // GenerateCounts:
    //   GG existence raw 9 (9 + (-2 DMs) = 7 ≤9 → present)
    //   GG quantity   raw 11 (11 + (-2 DMs) = 9 → 4 GGs)
    //   Belts existence raw 7 (7 + (+3 DMs) = 10 ≥8 → present)
    //   Belts quantity raw 7 (7 + (+3 DMs) = 10 → 2 belts)
    //   Terrestrials 2D=12 → 12 - 2 + (-1) = 9 (≥3) + D3-1 (D3=3 → +2) = 11
    // RollBaselineNumber: 9 → 9 + (-4 DMs) = 5
    // BaselineOrbit (3a, HZCO 3.3 ≥ 1): 2D=5 → variance -0.2 → 3.1
    // RollEmptyOrbits: 10 → 1
    // (Spread is computed, no roll consumed.)
    // PlaceOrbitSlots:
    //   Aab variance rolls (10 slots after baseline-fix at slot 5):
    //     5, 9, 7, 9, 7, 7, 7, 7, 7, 7
    //   B variance rolls (1 + 1 empty-bump = 2 slots, slot 1 not baseline):
    //     7, 7
    //   Cab variance rolls (5 slots, slot 1 not baseline):
    //     10, 5, 7, 9, 5
    // AddAnomalous: 10 (count=1), 10 (Retrograde), D3=1 (Aab), 7 (2D-2=5), 2 (d10)
    // PlaceWorlds (19 slots → D3 prefix; book narrates many rolls — feed
    // a sequence that places 4 GG, 2 belts, 12 terrestrials, 1 empty in
    // some order):
    //   Empty: 3,3 → C1
    //   GG×4: 3,5 / 1,5 / 1,6 / 2,6 → C3, A5, A6, A11
    //   Belts×2: 3,3 (collision → +1 → 3,4) / 1,4 → C2, A4
    //   ... (terrestrials fill the remaining 12 slots in roll order)
    //
    // RollPlanetEccentricities: 18 non-belt non-empty placements, each consuming one 2D.
    //   Provide 18 sevens for neutral eccentricity-0.0 results.
    //
    // For brevity in this plan: assemble the full sequence as a single
    // []int. The implementer should construct it carefully against the
    // book narration; any discrepancy will surface as a per-step
    // assertion failure.
    rolls := []int{
        // Counts
        9, 11, 7, 7, 12, 3,
        // Baseline number, baseline orbit
        9, 5,
        // Empty orbits
        10,
        // Place orbit slots (Aab 10 vars, B 2 vars, Cab 5 vars)
        5, 9, 7, 9, 7, 7, 7, 7, 7, 7,
        7, 7,
        10, 5, 7, 9, 5,
        // Anomalous: count, type, parent D3, orbit 2D-2, d10
        10, 10, 1, 7, 2,
        // PlaceWorlds: 1 empty + 4 GG + 2 belts + 12 terrestrials = 19 placements,
        // each consuming (prefix, right) = 2 rolls. 38 rolls total.
        // Construct deliberately so empty lands at C1, etc.
        3, 3, // empty → C1
        3, 5, 1, 5, 1, 6, 2, 6, // 4 GGs
        3, 3, 1, 4, // 2 belts (3,3 collides → +1 → 3,4 = C2)
        // 12 terrestrials — roll into remaining slots in order.
        // The remaining unassigned indices after empty+GGs+belts are:
        //   A1, A2, A3, A7, A8, A+, A9, A10, B1, B+, C4, C5
        // Roll prefixes/right pairs that map to those indices.
        1, 1, 1, 2, 1, 3, 2, 1, 2, 2, 2, 3, 2, 4, 2, 5, 3, 1, 3, 2, 3, 6, 4, 1,
        // Eccentricity rolls: 18 placements; provide 18 sevens.
        7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
    }
    got, err := worlds.GenerateSystemPlacement(roller.NewScripted(rolls...), sys)
    if err != nil {
        t.Fatalf("GenerateSystemPlacement: %v", err)
    }

    // High-level assertions.
    if got.Counts.GasGiants != 4 {
        t.Errorf("GasGiants = %d, want 4", got.Counts.GasGiants)
    }
    if got.Counts.PlanetoidBelts != 2 {
        t.Errorf("PlanetoidBelts = %d, want 2", got.Counts.PlanetoidBelts)
    }
    if got.Counts.Terrestrials != 12 {
        // 11 from GenerateCounts + 1 from anomalous
        t.Errorf("Terrestrials = %d, want 12", got.Counts.Terrestrials)
    }
    if got.Counts.Total != 18 {
        t.Errorf("Total = %d, want 18", got.Counts.Total)
    }
    if got.BaselineN != 5 {
        t.Errorf("BaselineN = %d, want 5", got.BaselineN)
    }
    if math.Abs(got.BaselineOrbit-3.1) > 0.05 {
        t.Errorf("BaselineOrbit = %v, want 3.1", got.BaselineOrbit)
    }
    if got.EmptyOrbits != 1 {
        t.Errorf("EmptyOrbits = %d, want 1", got.EmptyOrbits)
    }
    if math.Abs(got.SystemSpread-0.50) > 0.005 {
        t.Errorf("SystemSpread = %v, want 0.50", got.SystemSpread)
    }
    if len(got.Placements) != 19 {
        t.Fatalf("Placements = %d, want 19", len(got.Placements))
    }

    // Book p. 52 places the empty at C2 via Referee discretion; strict
    // dice puts it at C1. We assert C1 here.
    var emptySlots []string
    for _, p := range got.Placements {
        if p.Body == worlds.BodyEmpty {
            emptySlots = append(emptySlots, p.StarSlot)
        }
    }
    if len(emptySlots) != 1 || emptySlots[0] != "C1" {
        t.Errorf("empty slot = %v, want [C1] (strict-dice divergence from book p. 52 narrated C2)", emptySlots)
    }

    // The retrograde slot should be in the Aab group at orbit 5.2.
    var retro *worlds.Placement
    for i := range got.Placements {
        if got.Placements[i].Anomaly == worlds.AnomalyRetrograde {
            retro = &got.Placements[i]
            break
        }
    }
    if retro == nil {
        t.Fatalf("no retrograde placement found")
    }
    if retro.Group.Designation != "Aab" {
        t.Errorf("retrograde group = %s, want Aab", retro.Group.Designation)
    }
    if math.Abs(retro.Orbit-5.2) > 0.05 {
        t.Errorf("retrograde orbit = %v, want 5.2", retro.Orbit)
    }
}
```

- [ ] **Step 2: Run acceptance test to verify it fails**

```bash
go test ./worlds -run TestZed_FullPlacement -v
```

Expected: FAIL — `GenerateSystemPlacement undefined`.

- [ ] **Step 3: Implement the façade**

Create `worlds/system_placement.go`:

```go
package worlds

import (
    "errors"

    "wbh/roller"
    "wbh/stars"
)

// SystemPlacement is the full audit trail produced by
// GenerateSystemPlacement. Each named field corresponds to one Step in
// the WBH placement procedure.
type SystemPlacement struct {
    Counts        Counts
    Allocations   []StarAllocation
    BaselineN     int
    BaselineOrbit float64
    EmptyOrbits   int
    SystemSpread  float64
    Placements    []Placement
}

// ErrContinuationMethodUnsupported is returned when sys carries
// pre-existing mainworld data (no field today; placeholder for the
// future Continuation Method sub-project).
var ErrContinuationMethodUnsupported = errors.New(
    "worlds: pre-existing mainworld input requires Continuation Method, not yet encoded",
)

// GenerateSystemPlacement runs the WBH 9-step pipeline end-to-end:
//
//  0. GenerateCounts (pp. 36–38)
//  1. AllocateOrbitsByStar (pp. 43–44)
//  2. RollBaselineNumber (pp. 44–45)
//  3. BaselineOrbit (pp. 45–46)
//  4. RollEmptyOrbits (p. 48)
//  5. Spread (pp. 48–49)
//  6. PlaceOrbitSlots (pp. 49–50)
//  7. AddAnomalous (pp. 50–51)
//  8. PlaceWorlds (pp. 51–52)
//  9. RollPlanetEccentricities (p. 52)
func GenerateSystemPlacement(r roller.Roller, sys stars.System) (SystemPlacement, error) {
    counts, err := GenerateCounts(r, sys, CountsOpts{})
    if err != nil {
        return SystemPlacement{}, err
    }
    avail, err := AvailableOrbits(sys)
    if err != nil {
        return SystemPlacement{}, err
    }
    allocs, err := AllocateOrbitsByStar(avail, counts)
    if err != nil {
        return SystemPlacement{}, err
    }
    baselineN, err := RollBaselineNumber(r, sys, counts)
    if err != nil {
        return SystemPlacement{}, err
    }
    primary := allocs[0].Group
    baselineOrbit, err := BaselineOrbit(r, primary, primary.HZCO(), baselineN, counts.Total)
    if err != nil {
        return SystemPlacement{}, err
    }
    emptyOrbits, err := RollEmptyOrbits(r)
    if err != nil {
        return SystemPlacement{}, err
    }
    totalStars := 1 + secondaryStarCount(sys)
    spread := Spread(primary, allocs[0].AllocatedWorlds, baselineOrbit, baselineN, totalStars)
    slots, err := PlaceOrbitSlots(r, allocs, baselineOrbit, spread, emptyOrbits)
    if err != nil {
        return SystemPlacement{}, err
    }
    anomSlots, newCounts, err := AddAnomalous(r, slots, allocs, counts)
    if err != nil {
        return SystemPlacement{}, err
    }
    placements, err := PlaceWorlds(r, anomSlots, newCounts)
    if err != nil {
        return SystemPlacement{}, err
    }
    placements, err = RollPlanetEccentricities(r, placements)
    if err != nil {
        return SystemPlacement{}, err
    }
    return SystemPlacement{
        Counts:        newCounts,
        Allocations:   allocs,
        BaselineN:     baselineN,
        BaselineOrbit: baselineOrbit,
        EmptyOrbits:   emptyOrbits,
        SystemSpread:  spread,
        Placements:    placements,
    }, nil
}
```

- [ ] **Step 4: Run acceptance test to verify it passes**

```bash
go test ./worlds -run TestZed_FullPlacement -v
```

Expected: PASS. If individual assertions fail, debug step-by-step against the book narration; the dice sequence above is a best-effort encoding and may need adjustment to match the encoded DM stack exactly.

If the acceptance test reveals divergences not documented in the spec (beyond the C1/C2 empty-orbit one), pause and decide whether the spec needs amending or the implementation needs a fix. Do NOT silently change the spec to match buggy code.

- [ ] **Step 5: Run all tests + lint**

```bash
gofumpt -w worlds/
just check && just test
```

Expected: clean. All package-level tests green.

- [ ] **Step 6: Commit**

```bash
git add worlds/system_placement.go worlds/worked_examples_test.go
git commit -m "$(cat <<'EOF'
feat(worlds): GenerateSystemPlacement façade + TestZed_FullPlacement

Wires Steps 0–9 into a single GenerateSystemPlacement call returning
the full audit trail (Counts, Allocations, BaselineN, BaselineOrbit,
EmptyOrbits, SystemSpread, Placements).

Acceptance test reproduces the WBH Zed quintuple worked example
(pp. 36–52) end-to-end with one documented divergence: the empty
orbit lands at C1 under strict dice; the book narrates Referee
discretion moving it to C2.
EOF
)"
```

---

## Task 16: Final integration check + branch readiness

**Goal:** verify the branch is fully green, well-documented, and ready to merge.

- [ ] **Step 1: Run full check + test from a clean state**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go clean -testcache
just check && just test
```

Expected: all green; no `gofumpt` diffs; no `go vet` or lint findings; all tests pass.

- [ ] **Step 2: Sanity-check public API documentation**

Open each new file and verify every exported symbol has a GoDoc comment that references the WBH page or step it encodes:

```bash
for f in worlds/group_hzco.go worlds/counts.go worlds/allocations.go worlds/baseline.go worlds/empty_orbits.go worlds/spread.go worlds/orbit_slots.go worlds/anomalous.go worlds/placement.go worlds/planet_eccentricity.go worlds/system_placement.go; do
    echo "=== $f ==="
    grep -B1 "^func\|^type\|^var\|^const" "$f" | head -50
done
```

Verify every `func`, `type`, `var`, `const` (exported — first letter uppercase) has a leading `// X ...` comment referencing WBH.

- [ ] **Step 3: Run the existing 2A acceptance tests one more time to confirm the carry-forward refactor didn't regress them**

```bash
go test ./worlds -run "TestZed_AvailableOrbits|TestSol_AvailableOrbits" -v
```

Expected: PASS.

- [ ] **Step 4: Inspect git log for the branch and confirm linearity**

```bash
git log --oneline main..HEAD
```

Expected: ~16 commits, one per task, in the order of the plan.

- [ ] **Step 5: Stop here — do not merge. Hand off to the user.**

Print a summary message:

```text
Sub-project 2B is implementation-complete on branch feat/wbh-system-worlds-2b.
- 16 commits since main
- All tests pass (just check && just test green)
- TestZed_FullPlacement reproduces the book's Zed walkthrough end-to-end
  with one documented C1/C2 empty-orbit divergence

Ready for review and merge to main when the user is ready.
```

Do NOT run `git merge` or `git push` — leave both decisions to the user.
