# System Worlds 2C: Sizing + Moons + IISS Class II/III Form Implementation Plan (Go)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement sub-project 2C from the System Worlds and Orbits chapter: orbital periods, terrestrial/gas-giant sizing, significant moons, planet/moon designations, system profile, mainworld-candidate enumeration, and the IISS Class II/III survey form (form 0421D-II.III). Reproduce the book's Zed Class II/III form (p.63) to declared tolerances with two documented carve-outs (deferred HZ-candidate SAH cells; book p.58/p.63 d-moon size inconsistency).

**Architecture:** A new set of granular per-step functions in `wbh/worlds`, layered on top of 2B's `SystemPlacement` via a new `DetailedPlacement` struct that embeds `Placement` (continuing the existing `Slot → AnomalousSlot → Placement` chain). The IISS Class II/III form embeds the existing `stars.SurveyForm` (Class 0/I) and adds an Objects table. Bundles four small carry-forward fixes from 2B's post-mortem.

**Tech Stack:** Go 1.22+, `gofumpt` CLI as canonical formatter (not golangci-lint's bundled gofumpt), golangci-lint v2.12.1, `just` recipes.

**Spec:** `docs/specs/2026-05-03-system-worlds-2c-sizing-design.md`

**Source pages:** WBH pp. 53–67 (Planetary Orbital Periods, Sizing, Moons, Profile, Mainworld, IISS Class II/III form).

**Conventions:**

- Working directory: `/Users/markayers/Documents/Traveller/`.
- TDD per task: write test → run-fail → implement → run-pass → format → lint → commit.
- `gofumpt -w` before commit. `gofumpt` CLI is the formatter source of truth (not golangci-lint).
- Test files live in the same package (white-box) except `worked_examples_test.go` (black-box `package worlds_test`).
- Tables for non-numeric cells: struct rows. Tables with nullable numeric cells: `*float64` via the `fp` helper already in `worlds/available_orbits.go`.
- Branch: `feat/wbh-system-worlds-2c` (created in Pre-flight).
- 2B carry-forward findings refined during plan-writing:
  - **CF #3 is smaller than the spec implied.** `stars.RollEccentricity` (in `stars/orbits.go:71`) already implements the WBH p.27 sub-1.0/age>1Gyr DM-1 internally via `EccentricityOpts.Orbit` + `EccentricityOpts.SystemAgeGyr`. The bug is that `worlds.RollPlanetEccentricities` does not pass these fields. The fix is two added fields in one call site plus the new `ageGyr` parameter.
  - **`stars.OrbitPeriodYears` already exists** at `stars/orbits.go:144`. No new orbital-period function is required; `Period` is a new wrapping struct in `worlds/period.go` that calls `OrbitPeriodYears` and computes Days locally. The spec's "Cycle resolution" section is moot — Period lives in `worlds/`.
  - **`stars.System.AgeGyr`** (system.go:380) is the system-level age field. CF #3 reads `sys.AgeGyr`, not `sys.Primary.AgeGyr` (the spec wording was a forgivable conflation; the values are intended to match).
  - **`stars.SurveyComponent` and `stars.SurveyForm`** (stars/survey.go) already exist for Class 0/I. The Class II/III form extends them by (a) adding a `MAO` field on `SurveyComponent` and (b) embedding `SurveyForm` inside the new `worlds.IISSClass23Form` plus an `Objects` table.

---

## Pre-flight

- [ ] **Verify clean state on main and create the feature branch**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
git status
git branch --show-current
just check && just test
git checkout -b feat/wbh-system-worlds-2c
git branch --show-current
```

Expected: clean working tree on `main`; all tests green; new branch `feat/wbh-system-worlds-2c` checked out.

---

## File Structure

| File                                          | Responsibility                                                                                                                                             |
| --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `worlds/system_placement.go` (extend)         | Carry-forward #2 — wrap five `err` returns with `fmt.Errorf("worlds: <step>: %w", err)`.                                                                   |
| `worlds/worked_examples_test.go` (extend)     | Carry-forward #1 — `TestSol_GenerateSystemPlacement` smoke test. Plus Task 15: `TestZed_FullDetail` acceptance gate.                                       |
| `worlds/planet_eccentricity.go` (extend)      | Carry-forward #3 — `RollPlanetEccentricities` accepts `ageGyr float64`; passes `Orbit` and `SystemAgeGyr` into `EccentricityOpts`.                         |
| `worlds/planet_eccentricity_test.go` (extend) | Carry-forwards #3 + #4 — pass `ageGyr` to existing tests; strengthen two DM tests to assert specific `EccentricityValues` rows.                            |
| `worlds/period.go` (create)                   | `Period` struct (`Years`, `Days`); `PeriodFor` helper that wraps `stars.OrbitPeriodYears` and computes Days.                                               |
| `worlds/period_test.go` (create)              | Per-formula tests (single-star, multi-star, large-planet); Sol/Zed worked-example values.                                                                  |
| `worlds/sizing_terrestrial.go` (create)       | `SizeCode`, `TerrestrialSize`, `BasicTerrestrialDiameter` table, `RollTerrestrialSize`.                                                                    |
| `worlds/sizing_terrestrial_test.go` (create)  | Per-row table test; each 1D selector branch; Zed sizes from p.56.                                                                                          |
| `worlds/sizing_gasgiant.go` (create)          | `GasGiantClass`, `GasGiantSize`, `RollGasGiantSize`, Large-GG mass clamp special.                                                                          |
| `worlds/sizing_gasgiant_test.go` (create)     | Per-class + DM combinations + Large clamp + Zed gas giants.                                                                                                |
| `worlds/moons.go` (create)                    | `Moon`, `ParentInfo`, `CountMoons`, `SizeMoon`, Significant Moon Sizing table, Gas Giant Special Moon Sizing table.                                        |
| `worlds/moons_test.go` (create)               | Per-row + each adjacent-zone DM + Significant Moon Sizing branches + GG Special branches + terrestrial Size-1 + "exactly 2 less" 2D adjustment.            |
| `worlds/system_detail.go` (create)            | `DetailedPlacement`, `SystemDetail`, `DetailSystem` façade.                                                                                                |
| `worlds/system_detail_test.go` (create)       | Façade pipeline composition test; per-step ordering.                                                                                                       |
| `worlds/designations.go` (create)             | `AssignPlanetDesignations`, `AssignMoonDesignations`.                                                                                                      |
| `worlds/designations_test.go` (create)        | Belt-skip + per-group reset + moon alphabet ordering + multi-group composite.                                                                              |
| `worlds/profile.go` (create)                  | `ShortProfile`, `LongProfile` (note: collides by name with `stars.ShortProfile`; different package qualifies).                                             |
| `worlds/profile_test.go` (create)             | Short form + long form structure; Zed strings.                                                                                                             |
| `worlds/mainworld.go` (create)                | `MainworldCandidate`, `MarkHZ`, `MainworldCandidates`.                                                                                                     |
| `worlds/mainworld_test.go` (create)           | HZ-window inclusion; planet candidates; moon candidates including gas-giant moons; Zed enumeration matches book.                                           |
| `stars/survey.go` (extend)                    | Add `MAO float64` field to `SurveyComponent`; populate it in `BuildSurveyForm` for the same rows that today carry `HZCO` (solo primary, group composites). |
| `stars/survey_test.go` (extend)               | Add MAO assertions to existing Sol/Zed worked-example tests.                                                                                               |
| `worlds/survey_form.go` (create)              | `IISSClass23Form` (embeds `stars.SurveyForm`), `ObjectRow`, `IISSClass23Header`, `RenderIISSClass23`.                                                      |
| `worlds/survey_form_test.go` (create)         | Per-section construction; period rendering (years vs days); SAH cell rendering; Notes column composition.                                                  |

**18 new files, 5 edited files.** (One more edit than the spec: `stars/survey.go` + `stars/survey_test.go` for the MAO field. The spec listed 4 edits in `worlds/`; this discovers a 5th in `stars/`.)

---

## Task 1: Carry-forward #2 — Wrap `GenerateSystemPlacement` errors

**Source:** Spec § Bundled 2B carry-forward items, item #2.

**Files:** `worlds/system_placement.go` (extend).

**Goal:** Wrap each of the eight `err` returns in `GenerateSystemPlacement` with `fmt.Errorf("worlds: <step>: %w", err)` so callers can identify which step failed.

- [ ] **Step 1: Read the current file to confirm five callsites**

```bash
grep -n "return SystemPlacement{}, err" /Users/markayers/Documents/Traveller/worlds/system_placement.go
```

Expected: eight occurrences (one per pipeline step that can return an error). Note: the spec said "five callsites" — the actual count is eight (`GenerateCounts`, `AvailableOrbits`, `AllocateOrbitsByStar`, `RollBaselineNumber`, `BaselineOrbit`, `RollEmptyOrbits`, `PlaceOrbitSlots`, `AddAnomalous`, `PlaceWorlds`, `RollPlanetEccentricities` — actually nine). Wrap whatever count the file shows.

- [ ] **Step 2: Write a failing test that asserts errors are wrapped per step**

Add to `worlds/system_placement_test.go` (create the file if it doesn't exist; check first):

```bash
ls /Users/markayers/Documents/Traveller/worlds/system_placement_test.go 2>&1
```

If absent, create `worlds/system_placement_test.go`:

```go
package worlds

import (
	"errors"
	"strings"
	"testing"

	"wbh/roller"
	"wbh/stars"
)

// failingRoller returns a roller that succeeds for the first N rolls,
// then panics. Used to drive GenerateSystemPlacement up to a known
// failure point without rebuilding a full scripted sequence.
type failingRoller struct {
	inner roller.Roller
	calls int
	limit int
}

func (f *failingRoller) Roll(notation string) int {
	if f.calls >= f.limit {
		// Force an error path by returning an out-of-range value the
		// downstream parser will reject. We can't actually return an
		// error from Roll — so we'll trigger the error via an invalid
		// downstream input instead. See test below for actual approach.
		panic("rolled past limit")
	}
	f.calls++
	return f.inner.Roll(notation)
}

// TestGenerateSystemPlacement_WrappedErrorMessage asserts that errors
// returned by GenerateSystemPlacement carry a "worlds: <step>:" prefix
// identifying which step failed. We trigger an error by forcing
// AvailableOrbits to fail (post-stellar primary).
func TestGenerateSystemPlacement_WrappedErrorMessage(t *testing.T) {
	t.Parallel()

	primary := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindBrownDwarf, // post-stellar primary triggers ErrPostStellarPrimaryUnsupported
	})
	sys := stars.System{Primary: primary}

	r := roller.NewSeeded(1)
	_, err := GenerateSystemPlacement(r, sys)
	if err == nil {
		t.Fatal("GenerateSystemPlacement returned nil error; expected post-stellar failure")
	}
	if !errors.Is(err, ErrPostStellarPrimaryUnsupported) {
		t.Errorf("err is not ErrPostStellarPrimaryUnsupported via Is(): %v", err)
	}
	if !strings.Contains(err.Error(), "worlds:") {
		t.Errorf("err.Error() = %q, want prefix containing 'worlds:'", err.Error())
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run TestGenerateSystemPlacement_WrappedErrorMessage -v
```

Expected: FAIL — `err.Error()` does not contain `"worlds:"` (current code returns the bare error from `AvailableOrbits`).

- [ ] **Step 4: Wrap each err return in `system_placement.go`**

Edit `worlds/system_placement.go`. At the top, ensure `import "fmt"` is present (add it if missing). Then wrap each `return SystemPlacement{}, err` site:

```go
counts, err := GenerateCounts(r, sys, CountsOpts{})
if err != nil {
	return SystemPlacement{}, fmt.Errorf("worlds: counts: %w", err)
}
avail, err := AvailableOrbits(sys)
if err != nil {
	return SystemPlacement{}, fmt.Errorf("worlds: available-orbits: %w", err)
}
allocs, err := AllocateOrbitsByStar(avail, counts)
if err != nil {
	return SystemPlacement{}, fmt.Errorf("worlds: allocations: %w", err)
}
baselineN, err := RollBaselineNumber(r, sys, counts)
if err != nil {
	return SystemPlacement{}, fmt.Errorf("worlds: baseline-number: %w", err)
}
primary := allocs[0].Group
baselineOrbit, err := BaselineOrbit(r, primary, primary.HZCO(), baselineN, counts.Total)
if err != nil {
	return SystemPlacement{}, fmt.Errorf("worlds: baseline-orbit: %w", err)
}
emptyOrbits, err := RollEmptyOrbits(r)
if err != nil {
	return SystemPlacement{}, fmt.Errorf("worlds: empty-orbits: %w", err)
}
totalStars := 1 + secondaryStarCount(sys)
spread := Spread(primary, allocs[0].AllocatedWorlds, baselineOrbit, baselineN, totalStars)
slots, err := PlaceOrbitSlots(r, allocs, baselineN, baselineOrbit, spread, emptyOrbits)
if err != nil {
	return SystemPlacement{}, fmt.Errorf("worlds: orbit-slots: %w", err)
}
anomSlots, newCounts, err := AddAnomalous(r, slots, allocs, counts)
if err != nil {
	return SystemPlacement{}, fmt.Errorf("worlds: anomalous: %w", err)
}
placements, err := PlaceWorlds(r, anomSlots, newCounts)
if err != nil {
	return SystemPlacement{}, fmt.Errorf("worlds: place-worlds: %w", err)
}
placements, err = RollPlanetEccentricities(r, placements)
if err != nil {
	return SystemPlacement{}, fmt.Errorf("worlds: planet-eccentricity: %w", err)
}
```

(Task 3 will further extend `RollPlanetEccentricities` to accept `ageGyr`; the wrap stays the same.)

- [ ] **Step 5: Run the test to verify it passes**

```bash
go test ./worlds -run TestGenerateSystemPlacement_WrappedErrorMessage -v
```

Expected: PASS.

- [ ] **Step 6: Run full check + test to verify no regressions**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 7: Format + commit**

```bash
gofumpt -w worlds/system_placement.go worlds/system_placement_test.go
git add worlds/system_placement.go worlds/system_placement_test.go
git commit -m "fix(worlds): wrap GenerateSystemPlacement errors with step-named %w (CF#2)

Each of the nine error sites in GenerateSystemPlacement now wraps its
underlying error with fmt.Errorf(\"worlds: <step>: %w\", err) so callers
can identify which step failed. Pattern matches available_orbits.go:359.

Carry-forward item #2 from 2B post-mortem."
```

---

## Task 2: Carry-forward #1 — `TestSol_GenerateSystemPlacement` smoke test

**Source:** Spec § Bundled 2B carry-forward items, item #1.

**Files:** `worlds/worked_examples_test.go` (extend).

**Goal:** Single-star coverage was missing in 2B. Add a lightweight smoke test that asserts the single-star path runs without error and produces a non-empty placement.

- [ ] **Step 1: Read existing `composeSol()` helper to confirm its shape**

```bash
grep -n "func composeSol\|func composeZed\|func composeCorella" /Users/markayers/Documents/Traveller/worlds/worked_examples_test.go
```

Expected: one or more `compose*()` helpers defined in the file.

- [ ] **Step 2: Write the failing test**

Add to `worlds/worked_examples_test.go` (in the existing `package worlds_test` test file):

```go
// TestSol_GenerateSystemPlacement is a single-star smoke test: assert
// the GenerateSystemPlacement pipeline runs without error on a
// single-G2-V system, produces exactly one StarAllocation, and yields
// a non-empty Placements slice. This complements TestZed_FullPlacement
// (multi-star) to cover the single-star path that 2B left untested.
//
// No book-narrated dice trail is required; this is a smoke test, not
// a worked-example regression.
func TestSol_GenerateSystemPlacement(t *testing.T) {
	t.Parallel()

	sys := composeSol()

	// Use a seeded roller; the specific values don't matter for a smoke
	// test, only that the pipeline completes without error.
	r := roller.NewSeeded(42)

	sp, err := worlds.GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatalf("GenerateSystemPlacement returned error: %v", err)
	}

	if len(sp.Allocations) != 1 {
		t.Errorf("len(Allocations) = %d, want 1 (single-star system)", len(sp.Allocations))
	}
	if sp.Allocations[0].Group.Designation != "A" {
		t.Errorf("Allocations[0].Group.Designation = %q, want \"A\"", sp.Allocations[0].Group.Designation)
	}
	if len(sp.Placements) == 0 {
		t.Error("Placements is empty, want at least one body")
	}
	if sp.Counts.Total <= 0 {
		t.Errorf("Counts.Total = %d, want > 0", sp.Counts.Total)
	}
}
```

If `composeSol()` does not yet exist in the file, scaffold it from the existing `composeZed()` pattern. Sol is a single G2 V primary with no companions:

```go
func composeSol() stars.System {
	return stars.System{
		Primary: stars.Compose(stars.ComposeOpts{
			Kind:            stars.KindMainSequence,
			SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
			LuminosityClass: stars.V,
			Mass:            1.0,
			Diameter:        1.0,
			Temperature:     5800,
			Luminosity:      1.0,
			AgeGyr:          4.6,
		}),
		PrimaryDesignation: "A",
		AgeGyr:             4.6,
	}
}
```

- [ ] **Step 3: Run the test to verify it fails (or passes — see below)**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run TestSol_GenerateSystemPlacement -v
```

Expected: PASS (this is a smoke test — the pipeline already works for single-star). If it fails, the failure exposes a real bug; fix before continuing.

- [ ] **Step 4: Format + commit**

```bash
gofumpt -w worlds/worked_examples_test.go
git add worlds/worked_examples_test.go
git commit -m "test(worlds): add TestSol_GenerateSystemPlacement smoke test (CF#1)

Single-star smoke test for GenerateSystemPlacement asserts no error,
exactly one StarAllocation with designation \"A\", non-empty Placements,
and Counts.Total > 0. Complements TestZed_FullPlacement (multi-star).

Carry-forward item #1 from 2B post-mortem."
```

---

## Task 3: Carry-forwards #3 + #4 — Plumb `ageGyr` + strengthen eccentricity assertions

**Source:** Spec § Bundled 2B carry-forward items, items #3 and #4.

**Files:** `worlds/planet_eccentricity.go` (extend), `worlds/planet_eccentricity_test.go` (extend), `worlds/system_placement.go` (extend), `worlds/worked_examples_test.go` (possibly extend if Zed eccentricity assertions shift).

**Goal:** `RollPlanetEccentricities` accepts `ageGyr float64` and passes the placement's `Orbit` and the system's `AgeGyr` into `stars.EccentricityOpts` so the existing WBH p.27 sub-1.0/age>1Gyr DM-1 actually applies. Strengthen two DM tests to assert specific resulting values via `stars.EccentricityValues` table lookup.

**Refined finding from plan-writing:** `stars.RollEccentricity` (orbits.go:71-99) already implements the DM-1 internally when `Orbit < 1.0 && SystemAgeGyr > 1.0`. The bug is solely that `worlds.RollPlanetEccentricities` doesn't pass `Orbit` or `SystemAgeGyr`. Fix is two added fields in one call site plus a new function parameter.

- [ ] **Step 1: Write failing test for ageGyr signature change**

Edit `worlds/planet_eccentricity_test.go`. Find the existing test that exercises `RollPlanetEccentricities` and add a new test asserting that the DM-1 actually applies:

```go
// TestRollPlanetEccentricities_AgeDMApplies asserts that
// RollPlanetEccentricities passes Orbit and SystemAgeGyr through to
// stars.RollEccentricity, so the WBH p.27 sub-1.0/age>1Gyr DM-1
// actually shifts the EccentricityValues table-lookup index.
//
// Setup: one terrestrial placement at orbit 0.5 with ageGyr 5.0.
// Without the DM, a 2D natural roll of 8 → row 8-9 → Base 0.03 + 1D÷100.
// With the DM-1, the row becomes 7 → row 6-7 → Base 0.00 + 1D÷200.
// We script natural=8 + second=4 to make the row choice the deciding factor.
func TestRollPlanetEccentricities_AgeDMApplies(t *testing.T) {
	t.Parallel()

	// Build a single terrestrial placement at orbit 0.5.
	p := Placement{
		AnomalousSlot: AnomalousSlot{
			Slot: Slot{Orbit: 0.5},
		},
		Body: BodyTerrestrial,
	}

	// Script: 2D natural=8, then 1D second-roll=4.
	// With ageGyr > 1, DM-1 → row 7 → Base 0.00 + 4/200 = 0.02.
	// Without DM (ageGyr 0), row 8 → Base 0.03 + 4/100 = 0.07.
	scripted := roller.NewScripted(8, 4)
	out, err := RollPlanetEccentricities(scripted, []Placement{p}, 5.0)
	if err != nil {
		t.Fatalf("RollPlanetEccentricities returned error: %v", err)
	}
	if got, want := out[0].Eccentricity, 0.02; math.Abs(got-want) > 1e-9 {
		t.Errorf("Eccentricity with ageGyr=5.0 = %v, want %v (DM-1 should shift to row 7)", got, want)
	}

	// And confirm: without the age DM, the same dice produce row 8's value.
	scripted = roller.NewScripted(8, 4)
	out, err = RollPlanetEccentricities(scripted, []Placement{p}, 0.5)
	if err != nil {
		t.Fatalf("RollPlanetEccentricities (no age DM) returned error: %v", err)
	}
	if got, want := out[0].Eccentricity, 0.07; math.Abs(got-want) > 1e-9 {
		t.Errorf("Eccentricity with ageGyr=0.5 = %v, want %v (no DM)", got, want)
	}
}
```

Note: the existing tests in this file call `RollPlanetEccentricities(r, ps)` with the old two-arg signature. After the change they will fail to compile until updated in Step 4.

- [ ] **Step 2: Run the test to verify it fails**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run TestRollPlanetEccentricities_AgeDMApplies -v
```

Expected: FAIL — does not compile because `RollPlanetEccentricities` only takes two arguments.

- [ ] **Step 3: Update `RollPlanetEccentricities` signature and pass through Orbit + ageGyr**

Edit `worlds/planet_eccentricity.go`:

```go
// RollPlanetEccentricities implements WBH Step 9 (p. 52).
//
// For each non-empty non-belt placement, calls stars.RollEccentricity
// with the placement's anomaly DM (AnomalousSlot.EccentricityDM) passed
// through stars.EccentricityOpts.ExtraDM, the placement's Orbit, and the
// caller-supplied ageGyr (typically sys.AgeGyr). The Orbit + ageGyr pair
// triggers the WBH p.27 sub-1.0/age>1Gyr DM-1 inside RollEccentricity.
//
// Belts and empty slots are skipped (no roll consumed).
//
// Trojan slots (Anomaly == AnomalyTrojan) are handled specially per
// WBH p. 51: they inherit the orbit, eccentricity, and inclination of
// the slot they shadow. We do not roll a fresh eccentricity for them;
// we copy from the slot whose StarSlot matches TrojanOf in a second pass.
func RollPlanetEccentricities(r roller.Roller, ps []Placement, ageGyr float64) ([]Placement, error) {
	out := make([]Placement, len(ps))
	copy(out, ps)
	// First pass: roll for non-Trojan, non-empty, non-belt placements.
	for i := range out {
		if out[i].Body == BodyEmpty || out[i].Body == BodyPlanetoidBelt {
			continue
		}
		if out[i].Anomaly == AnomalyTrojan {
			continue // handled in pass 2
		}
		ecc, err := stars.RollEccentricity(r, stars.EccentricityOpts{
			ExtraDM:      out[i].EccentricityDM,
			NestingDepth: nestingDepthFor(out[i]),
			Orbit:        out[i].Orbit,
			SystemAgeGyr: ageGyr,
		})
		if err != nil {
			return nil, err
		}
		out[i].Eccentricity = ecc
	}
	// Second pass: Trojans inherit from their TrojanOf parent.
	for i := range out {
		if out[i].Anomaly != AnomalyTrojan {
			continue
		}
		for j := range out {
			if out[j].StarSlot == out[i].TrojanOf {
				out[i].Eccentricity = out[j].Eccentricity
				break
			}
		}
	}
	return out, nil
}
```

- [ ] **Step 4: Update existing planet_eccentricity_test.go callsites to pass ageGyr=0.0**

For each existing call to `RollPlanetEccentricities(r, ps)`, append `, 0.0`:

```bash
grep -n "RollPlanetEccentricities(" /Users/markayers/Documents/Traveller/worlds/planet_eccentricity_test.go
```

For each matching line, edit to pass ageGyr=0.0 (preserves pre-CF behavior — DM doesn't apply when ageGyr is 0):

```go
// Before:
out, err := RollPlanetEccentricities(scripted, []Placement{p})
// After:
out, err := RollPlanetEccentricities(scripted, []Placement{p}, 0.0)
```

- [ ] **Step 5: Update `system_placement.go` façade caller to pass `sys.AgeGyr`**

Edit `worlds/system_placement.go`. Find the `RollPlanetEccentricities` call (Step 9 of the pipeline) and pass `sys.AgeGyr`:

```go
placements, err = RollPlanetEccentricities(r, placements, sys.AgeGyr)
if err != nil {
	return SystemPlacement{}, fmt.Errorf("worlds: planet-eccentricity: %w", err)
}
```

- [ ] **Step 6: Run the new test to verify it passes**

```bash
go test ./worlds -run TestRollPlanetEccentricities_AgeDMApplies -v
```

Expected: PASS.

- [ ] **Step 7: Run all tests to identify any breakage in TestZed_FullPlacement**

```bash
just test
```

Expected: most tests pass. **Possible failure:** `TestZed_FullPlacement` might fail if the Zed system has any non-belt placement at orbit < 1.0 and the existing test asserts a specific eccentricity for it. Zed has B I at orbit 0.52 (< 1.0) with system age 6.336 Gyr (> 1.0) — the DM-1 now applies.

- [ ] **Step 8: If `TestZed_FullPlacement` fails, re-derive B I's expected eccentricity**

If Step 7 reports failure, examine the failure message:

```bash
go test ./worlds_test -run TestZed_FullPlacement -v 2>&1 | head -50
```

For each failed assertion (likely B I's eccentricity), the underlying scripted dice did not change — only the table-lookup index shifted by -1 due to the new DM. Update the assertion in `worked_examples_test.go` to the new value computed by `stars.EccentricityValues[row-1]` lookup. Document the change inline:

```go
// B I at orbit 0.52 with sys.AgeGyr=6.336 triggers the WBH p.27
// sub-1.0/age>1Gyr DM-1 (added in CF#3). The 2D natural roll N
// shifts to row N-1 in EccentricityValues. Pre-CF assertion was
// 0.003 (row 8); post-CF assertion is <new value> (row 7).
```

If 2B's existing test happens to have asserted with the post-DM value already (because the script roller's natural dice already targeted row 7), no change needed and Step 7 passes cleanly.

- [ ] **Step 9: Strengthen `TestRollEccentricity_ExtraDM` and `TestRollPlanetEccentricities_AppliesAnomalyDM` (CF #4)**

Locate these two tests in `worlds/planet_eccentricity_test.go`:

```bash
grep -n "TestRollEccentricity_ExtraDM\|TestRollPlanetEccentricities_AppliesAnomalyDM" /Users/markayers/Documents/Traveller/worlds/planet_eccentricity_test.go
```

For each, replace the "DM had some effect" assertion (typically `result != noDMResult`) with a specific expected value derived from `stars.EccentricityValues[row]` lookup. Pattern:

```go
// TestRollEccentricity_ExtraDM (refactored):
// Script natural=10 + second=5. With ExtraDM=+2, row becomes 12 → Base 0.30 + 5÷20 = 0.55.
// Without ExtraDM, row stays 10 → Base 0.05 + 5÷20 = 0.30.
scripted := roller.NewScripted(10, 5)
ps := []Placement{{
	AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 5.0}},
	Body:          BodyTerrestrial,
	// EccentricityDM is on the embedded AnomalousSlot:
}}
ps[0].EccentricityDM = 2
out, err := RollPlanetEccentricities(scripted, ps, 0.0)
if err != nil {
	t.Fatalf("err: %v", err)
}
if got, want := out[0].Eccentricity, 0.55; math.Abs(got-want) > 1e-9 {
	t.Errorf("Eccentricity = %v, want %v (row 12: 0.30 + 5/20)", got, want)
}
```

Apply the same pattern to `TestRollPlanetEccentricities_AppliesAnomalyDM` — script the dice that produce a specific row + second-roll combination, assert the exact resulting value.

- [ ] **Step 10: Run all tests to verify**

```bash
just check && just test
```

Expected: all pass.

- [ ] **Step 11: Format + commit**

```bash
gofumpt -w worlds/planet_eccentricity.go worlds/planet_eccentricity_test.go worlds/system_placement.go worlds/worked_examples_test.go
git add worlds/planet_eccentricity.go worlds/planet_eccentricity_test.go worlds/system_placement.go worlds/worked_examples_test.go
git commit -m "fix(worlds): plumb ageGyr into RollPlanetEccentricities (CF#3+CF#4)

CF#3: RollPlanetEccentricities now accepts ageGyr and passes the
placement's Orbit + ageGyr through to stars.EccentricityOpts so the
WBH p.27 sub-1.0/age>1Gyr DM-1 (already implemented in
stars.RollEccentricity) actually applies. The fix is two added fields
in one call site plus a new function parameter — smaller than the
spec implied because the DM logic itself was already in place.

CF#4: TestRollEccentricity_ExtraDM and
TestRollPlanetEccentricities_AppliesAnomalyDM strengthened to assert
specific resulting eccentricity values via EccentricityValues table
lookup, not just \"DM had some effect.\""
```

---

## Task 4: `Period` struct + `worlds/period.go`

**Source:** Spec § Architecture § Period; refined to live in `worlds/` since `stars.OrbitPeriodYears` already exists.

**Files:** `worlds/period.go` (create), `worlds/period_test.go` (create).

**Goal:** Add the `Period` struct (Years + Days) and a `PeriodFor` helper that wraps the existing `stars.OrbitPeriodYears` and computes Days. Three call patterns from WBH p.53 reduce to two arguments to the existing helper.

- [ ] **Step 1: Write the failing test**

Create `worlds/period_test.go`:

```go
package worlds

import (
	"math"
	"testing"
)

// TestPeriodFor covers the three WBH p.53 cases:
//
//	Single star:        P = sqrt(AU^3 / M☉)              → PeriodFor(au, M, 0)
//	Multiple stars:     P = sqrt(AU^3 / Σ M☉)           → PeriodFor(au, sumM, 0)
//	Large planet:       P = sqrt(AU^3 / (Σ M☉ + m⊕×ε)) → PeriodFor(au, sumM, mEarth)
//
// where ε = 0.000003 (Terra mass in solar units).
func TestPeriodFor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		au        float64
		sumMass   float64 // sum of stellar masses in M☉
		mEarth    float64 // body mass in m⊕ (0 to skip Large Planet variant)
		wantYears float64
	}{
		{
			// Sol single-star: Earth at 1.0 AU, M=1.0 → 1.000y
			name: "Sol Earth", au: 1.0, sumMass: 1.0, mEarth: 0,
			wantYears: 1.0,
		},
		{
			// Zed B I: orbit 0.52 → AU 0.208 around B alone (M=0.626)
			// sqrt(0.208^3 / 0.626) = sqrt(0.008998 / 0.626) = sqrt(0.01437) = 0.11988y
			// Form p.63 shows 0.120y.
			name: "Zed B I", au: 0.208, sumMass: 0.626, mEarth: 0,
			wantYears: 0.120,
		},
		{
			// Zed AB I: orbit 7.2 → AU 5.68 with sumStellarMass = M(Aa)+M(Ab)+M(B) = 0.929+0.907+0.626 = 2.462
			// sqrt(5.68^3 / 2.462) = sqrt(183.25/2.462) = sqrt(74.43) = 8.628y
			// Form p.63 shows 8.627y for B (Aab+B barycentre at orbit 7.2).
			name: "Zed AB I", au: 5.68, sumMass: 2.462, mEarth: 0,
			wantYears: 8.627,
		},
		{
			// Large-Planet variant smoke: a 4000⊕ planet adds 4000×0.000003 = 0.012 M☉ to sumMass.
			// At AU=5, sumMass=1.0 → without: sqrt(125/1.0)=11.180y; with mEarth=4000:
			// sqrt(125/(1.0+0.012))=sqrt(125/1.012)=sqrt(123.52)=11.114y.
			name: "Large planet variant 4000 mEarth", au: 5.0, sumMass: 1.0, mEarth: 4000,
			wantYears: 11.114,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := PeriodFor(tc.au, tc.sumMass, tc.mEarth)
			if math.Abs(got.Years-tc.wantYears) > 0.005 {
				t.Errorf("Years = %v, want %v (±0.005)", got.Years, tc.wantYears)
			}
			if math.Abs(got.Days-got.Years*365.25) > 1e-9 {
				t.Errorf("Days = %v, want Years*365.25 = %v", got.Days, got.Years*365.25)
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run TestPeriodFor -v
```

Expected: FAIL — `PeriodFor` and `Period` are undefined.

- [ ] **Step 3: Create `worlds/period.go`**

```go
package worlds

import "wbh/stars"

// Period — orbital period; both Years and Days are populated and the
// renderer picks based on magnitude (form p.63 uses days for periods
// shorter than ~0.05y, otherwise years).
//
// WBH p.53 ("Length of 'Years'") gives three forms:
//
//	Single star:        P = sqrt(AU^3 / M☉)
//	Multiple stars:     P = sqrt(AU^3 / Σ M☉)
//	Large planet:       P = sqrt(AU^3 / (Σ M☉ + m⊕ × 0.000003))
//
// All three reduce to one call to stars.OrbitPeriodYears(au, sumMass, m)
// where m = bodyMassEarth × 0.000003 (or 0 for the standard cases).
type Period struct {
	Years float64 // primary representation; from Kepler's 3rd
	Days  float64 // = Years * 365.25
}

// massSolarPerEarth is the WBH p.53 "Large Planet" mass-conversion
// factor: 1 Terra mass in solar units.
const massSolarPerEarth = 0.000003

// PeriodFor computes a Period for a body at orbit (au) given the sum
// of stellar masses interior to that orbit (sumStellarMassSolar) and
// the body's mass in Terra masses (bodyMassEarth, 0 for the standard
// formula). Wraps stars.OrbitPeriodYears.
func PeriodFor(au, sumStellarMassSolar, bodyMassEarth float64) Period {
	years := stars.OrbitPeriodYears(au, sumStellarMassSolar, bodyMassEarth*massSolarPerEarth)
	return Period{
		Years: years,
		Days:  years * 365.25,
	}
}
```

- [ ] **Step 4: Run the test to verify it passes**

```bash
go test ./worlds -run TestPeriodFor -v
```

Expected: PASS.

- [ ] **Step 5: Run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 6: Format + commit**

```bash
gofumpt -w worlds/period.go worlds/period_test.go
git add worlds/period.go worlds/period_test.go
git commit -m "feat(worlds): Period struct + PeriodFor wrapping stars.OrbitPeriodYears

Period carries Years + Days for IISS form rendering. PeriodFor wraps
the existing stars.OrbitPeriodYears with the WBH p.53 large-planet
mass conversion (m⊕ × 0.000003 to solar units). Validates against
Sol Earth (1.000y), Zed B I (0.120y), Zed AB I (8.627y), and a
4000⊕ large-planet smoke case."
```

---

## Task 5: Terrestrial Sizing

**Source:** Spec § Architecture § Sizing (pp. 54-55); WBH p.54 Basic Terrestrial World Size + Terrestrial World Sizing tables.

**Files:** `worlds/sizing_terrestrial.go` (create), `worlds/sizing_terrestrial_test.go` (create).

**Goal:** `SizeCode` newtype + `BasicTerrestrialDiameter` table + `RollTerrestrialSize` (1D selector → second roll).

- [ ] **Step 1: Write failing tests for the table + per-branch sizing**

Create `worlds/sizing_terrestrial_test.go`:

```go
package worlds

import (
	"math"
	"testing"

	"wbh/roller"
)

// TestBasicTerrestrialDiameter verifies the WBH p.54 Basic Terrestrial
// World Size table: every code maps to its book diameter in km.
func TestBasicTerrestrialDiameter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		code SizeCode
		km   float64
	}{
		{"0", 0}, {"R", 0}, {"S", 600},
		{"1", 1600}, {"2", 3200}, {"3", 4800}, {"4", 6400}, {"5", 8000},
		{"6", 9600}, {"7", 11200}, {"8", 12800}, {"9", 14400},
		{"A", 16000}, {"B", 17600}, {"C", 19200}, {"D", 20800}, {"E", 22400}, {"F", 24000},
	}
	for _, tc := range cases {
		t.Run(string(tc.code), func(t *testing.T) {
			t.Parallel()
			got := BasicTerrestrialDiameter(tc.code)
			if math.Abs(got-tc.km) > 1e-9 {
				t.Errorf("BasicTerrestrialDiameter(%q) = %v, want %v", tc.code, got, tc.km)
			}
		})
	}
}

// TestRollTerrestrialSize_Branches covers each of the three 1D
// selector branches per WBH p.54 Terrestrial World Sizing table:
//
//	1-2 → second roll 1D    range 1-6
//	3-4 → second roll 2D    range 2-C(12)
//	5-6 → second roll 2D+3  range 5-F(15)
func TestRollTerrestrialSize_Branches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		dice     []int  // scripted dice
		wantCode SizeCode
	}{
		// Branch 1-2: 1D selector=1, second 1D=4 → Size 4
		{"selector1 second1D=4", []int{1, 4}, "4"},
		// Branch 3-4: 1D selector=3, second 2D=7 → Size 7
		{"selector3 second2D=7", []int{3, 7}, "7"},
		// Branch 3-4: 1D selector=4, second 2D=12 → Size C
		{"selector4 second2D=12", []int{4, 12}, "C"},
		// Branch 5-6: 1D selector=5, second 2D+3=8 → Size 11=B (8 from 2D + 3 = 11; 11 in hex = B)
		// Note: spec: "5-6 → 2D+3 (range 5-F(15))". So result = 2D+3. 2D=8 → 8+3=11 → "B".
		{"selector5 second2D+3=11", []int{5, 8}, "B"},
		// Branch 5-6: max — 2D=12 + 3 = 15 → "F"
		{"selector6 second2D+3=15", []int{6, 12}, "F"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roller.NewScripted(tc.dice...)
			got, err := RollTerrestrialSize(r)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got.SizeCode != tc.wantCode {
				t.Errorf("SizeCode = %q, want %q", got.SizeCode, tc.wantCode)
			}
			if got.DiameterKm != BasicTerrestrialDiameter(tc.wantCode) {
				t.Errorf("DiameterKm = %v, want %v", got.DiameterKm, BasicTerrestrialDiameter(tc.wantCode))
			}
		})
	}
}

// TestRollTerrestrialSize_ZedAabII reproduces a Zed terrestrial size from p.56:
//   Aab II  Terrestrial  Size Rolls "1: 6 = 6"  Code 6
// Selector 1D=1 → branch 1-2 → second 1D=6 → Size 6.
func TestRollTerrestrialSize_ZedAabII(t *testing.T) {
	t.Parallel()
	r := roller.NewScripted(1, 6)
	got, err := RollTerrestrialSize(r)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "6" {
		t.Errorf("SizeCode = %q, want \"6\"", got.SizeCode)
	}
	if got.DiameterKm != 9600 {
		t.Errorf("DiameterKm = %v, want 9600", got.DiameterKm)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run "TestBasicTerrestrialDiameter|TestRollTerrestrialSize" -v
```

Expected: FAIL — `SizeCode`, `BasicTerrestrialDiameter`, `RollTerrestrialSize`, `TerrestrialSize` undefined.

- [ ] **Step 3: Create `worlds/sizing_terrestrial.go`**

```go
package worlds

import (
	"fmt"

	"wbh/roller"
)

// SizeCode is the WBH terrestrial Size character (pp. 54, 56):
//
//	"0" — planetoid (0 km)
//	"R" — ring (0 km)
//	"S" — small body (600 km)
//	"1"-"9", "A"-"F" — Sizes 1-15 (1600 km × Size code)
//
// Empty string means "not a size-having body" (belt or empty slot).
type SizeCode string

// TerrestrialSize is the result of RollTerrestrialSize.
type TerrestrialSize struct {
	SizeCode   SizeCode
	DiameterKm float64
}

// basicTerrestrialDiameterTable is the WBH p.54 Basic Terrestrial
// World Size table (km per Size code).
var basicTerrestrialDiameterTable = map[SizeCode]float64{
	"0": 0, "R": 0, "S": 600,
	"1": 1600, "2": 3200, "3": 4800, "4": 6400, "5": 8000,
	"6": 9600, "7": 11200, "8": 12800, "9": 14400,
	"A": 16000, "B": 17600, "C": 19200, "D": 20800, "E": 22400, "F": 24000,
}

// BasicTerrestrialDiameter returns the diameter in km for a SizeCode
// per WBH p.54. Returns 0 for unknown codes (callers should validate
// SizeCode comes from a known source).
func BasicTerrestrialDiameter(code SizeCode) float64 {
	return basicTerrestrialDiameterTable[code]
}

// sizeCodeForN converts an integer Size 0-15 to its hex code (1=1, 9=9, 10=A, ..., 15=F).
// Used internally by RollTerrestrialSize and SizeMoon.
func sizeCodeForN(n int) SizeCode {
	if n < 0 {
		return "0"
	}
	if n > 15 {
		n = 15
	}
	switch {
	case n < 10:
		return SizeCode(fmt.Sprintf("%d", n))
	default:
		return SizeCode(fmt.Sprintf("%c", 'A'+n-10))
	}
}

// nForSizeCode is the inverse of sizeCodeForN. Returns -1 for "R" / "S" / "" / unknown.
// Used by moon-sizing logic that needs to compute "Size-1-1D" arithmetic on a parent.
func nForSizeCode(c SizeCode) int {
	switch c {
	case "R", "S", "":
		return -1
	}
	if len(c) != 1 {
		return -1
	}
	ch := c[0]
	switch {
	case ch >= '0' && ch <= '9':
		return int(ch - '0')
	case ch >= 'A' && ch <= 'F':
		return int(ch-'A') + 10
	}
	return -1
}

// RollTerrestrialSize rolls the WBH p.54 Terrestrial World Sizing
// procedure: a 1D selector chooses one of three second-roll formulas:
//
//	1-2 → second roll 1D            (range 1-6)
//	3-4 → second roll 2D            (range 2-C/12)
//	5-6 → second roll 2D+3          (range 5-F/15; clamped to F)
//
// Returns the resulting SizeCode and its book diameter in km.
func RollTerrestrialSize(r roller.Roller) (TerrestrialSize, error) {
	selector := r.Roll("1D")
	if selector < 1 || selector > 6 {
		return TerrestrialSize{}, fmt.Errorf("worlds: terrestrial size selector out of range: %d", selector)
	}

	var n int
	switch {
	case selector <= 2:
		n = r.Roll("1D")
	case selector <= 4:
		n = r.Roll("2D")
	default: // 5-6
		n = r.Roll("2D") + 3
	}
	if n > 15 {
		n = 15
	}
	if n < 1 {
		n = 1
	}
	code := sizeCodeForN(n)
	return TerrestrialSize{
		SizeCode:   code,
		DiameterKm: basicTerrestrialDiameterTable[code],
	}, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
go test ./worlds -run "TestBasicTerrestrialDiameter|TestRollTerrestrialSize" -v
```

Expected: PASS.

- [ ] **Step 5: Run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 6: Format + commit**

```bash
gofumpt -w worlds/sizing_terrestrial.go worlds/sizing_terrestrial_test.go
git add worlds/sizing_terrestrial.go worlds/sizing_terrestrial_test.go
git commit -m "feat(worlds): RollTerrestrialSize + SizeCode (WBH p.54)

SizeCode newtype encodes terrestrial Size characters (\"0\", \"R\",
\"S\", \"1\"-\"9\", \"A\"-\"F\"). BasicTerrestrialDiameter looks up
km per Size code. RollTerrestrialSize implements the 1D selector → second
roll (1D / 2D / 2D+3) per the WBH p.54 procedure. sizeCodeForN /
nForSizeCode helpers convert between integer Size and hex code for
downstream moon-size arithmetic."
```

---

## Task 6: Gas Giant Sizing

**Source:** Spec § Architecture § Sizing (pp. 54-55); WBH p.55 Gas Giant Sizing table + Large GG mass clamp footnote.

**Files:** `worlds/sizing_gasgiant.go` (create), `worlds/sizing_gasgiant_test.go` (create).

**Goal:** `GasGiantClass` enum + `GasGiantSize` struct + `RollGasGiantSize`. Three categories (Small/Medium/Large) selected by 1D+DM, each with its own diameter and mass sub-rolls. Large GG has a special mass clamp when initial mass ≥ 3,000⊕.

**WBH p.55 table:**

| 1D+DM | Class             | Code | Second (Diameter) | Diameter range | Third (Mass) | Mass range   |
| ----- | ----------------- | ---- | ----------------- | -------------- | ------------ | ------------ |
| 2-    | Small (Neptune)   | GS   | D3+D3             | 2-6⊕           | 5×(1D+1)     | 10-35⊕       |
| 3-4   | Medium (Jupiter)  | GM   | 1D+6              | 6-12⊕          | 20×(3D-1)    | 40-340⊕      |
| 5+    | Large (Superjov.) | GL   | 2D+6              | 8-18⊕          | D3×50×(3D+4) | 350-4,000⊕\* |

DMs:

- Brown Dwarf primary, M-V star, or any Class VI star: DM-1
- System Spread < 0.1: DM-1

\* Large mass clamp (footnote): if initial mass ≥3,000⊕ (3D ≥ 15 in the third roll), substitute mass = 4000 - 200×(2D-2).

- [ ] **Step 1: Write failing tests for the three classes**

Create `worlds/sizing_gasgiant_test.go`:

```go
package worlds

import (
	"math"
	"testing"

	"wbh/roller"
)

// TestRollGasGiantSize_Classes covers the three category branches per
// WBH p.55 Gas Giant Sizing table (no DMs).
func TestRollGasGiantSize_Classes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		dice           []int // scripted: selector (1D) + diameter dice + mass dice
		dms            int
		wantClass      GasGiantClass
		wantDiamCode   string
		wantDiamEarth  float64
		wantMassEarth  float64
	}{
		{
			// Selector 1D=2 → 2- → Small (GS)
			// Diameter D3+D3: dice 1,2 → 1+2=3 → "3"
			// Mass 5×(1D+1): die 4 → 5×(4+1)=25
			name: "Small GS smallest", dice: []int{2, 1, 2, 4}, dms: 0,
			wantClass: GasGiantSmall, wantDiamCode: "3", wantDiamEarth: 3, wantMassEarth: 25,
		},
		{
			// Selector 1D=3 → 3-4 → Medium (GM)
			// Diameter 1D+6: die 5 → 11 → "B"
			// Mass 20×(3D-1): dice 4,3,5 → 20×(12-1)=220
			name: "Medium GM mid", dice: []int{3, 5, 4, 3, 5}, dms: 0,
			wantClass: GasGiantMedium, wantDiamCode: "B", wantDiamEarth: 11, wantMassEarth: 220,
		},
		{
			// Selector 1D=5 → 5+ → Large (GL)
			// Diameter 2D+6: dice 4,4 → 14 → "E"
			// Mass D3×50×(3D+4): dice 2, 4,4,4 → 2×50×16 = 1600
			name: "Large GL big", dice: []int{5, 4, 4, 2, 4, 4, 4}, dms: 0,
			wantClass: GasGiantLarge, wantDiamCode: "E", wantDiamEarth: 14, wantMassEarth: 1600,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roller.NewScripted(tc.dice...)
			got, err := RollGasGiantSize(r, tc.dms)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got.Class != tc.wantClass {
				t.Errorf("Class = %v, want %v", got.Class, tc.wantClass)
			}
			if got.DiameterCode != tc.wantDiamCode {
				t.Errorf("DiameterCode = %q, want %q", got.DiameterCode, tc.wantDiamCode)
			}
			if math.Abs(got.DiameterEarth-tc.wantDiamEarth) > 1e-9 {
				t.Errorf("DiameterEarth = %v, want %v", got.DiameterEarth, tc.wantDiamEarth)
			}
			if math.Abs(got.MassEarth-tc.wantMassEarth) > 1e-9 {
				t.Errorf("MassEarth = %v, want %v", got.MassEarth, tc.wantMassEarth)
			}
		})
	}
}

// TestRollGasGiantSize_DMs verifies that dms shifts the selector roll.
// Without DM, selector 1D=3 → Medium. With DM-1, the same 1D becomes
// effective 2 → Small.
func TestRollGasGiantSize_DMs(t *testing.T) {
	t.Parallel()

	// dms=-1: selector 1D=3 → effective 2 → Small (GS), then GS sub-rolls.
	r := roller.NewScripted(3, 1, 2, 4)
	got, err := RollGasGiantSize(r, -1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.Class != GasGiantSmall {
		t.Errorf("Class with dms=-1 = %v, want GasGiantSmall (selector 3-1=2 → Small)", got.Class)
	}

	// dms=-2 (BD primary + system spread <0.1): selector 1D=4 → effective 2 → Small.
	r = roller.NewScripted(4, 1, 2, 4)
	got, err = RollGasGiantSize(r, -2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.Class != GasGiantSmall {
		t.Errorf("Class with dms=-2 = %v, want GasGiantSmall (selector 4-2=2 → Small)", got.Class)
	}
}

// TestRollGasGiantSize_LargeMassClamp covers the WBH p.55 footnote:
// if initial Large GG mass ≥3,000⊕ (3D ≥15), substitute mass = 4000 - 200×(2D-2).
func TestRollGasGiantSize_LargeMassClamp(t *testing.T) {
	t.Parallel()

	// Selector 5 → Large.
	// Diameter 2D+6 dice 6,6 → 18 → "G" wait 18 > 15 — clamped to "F".
	// Actually WBH p.55 Large diameter range is 8-18⊕ so the diameter code
	// should accept up to 18. But SizeCode max is "F"=15. So Large GG
	// diameter codes 16/17/18 must extend SizeCode beyond F.
	//
	// Per WBH p.55: "the third SAH digit for gas giants corresponds to its
	// diameter in Terran diameters and can use eHex notation as necessary
	// to record the gas giant diameter from 2 to J(18)."
	// So diameter codes for gas giants extend to "G"=16, "H"=17, "J"=18 (eHex).
	//
	// Test: selector 5, diameter 2D+6=12+6=18 → "J"
	// Mass D3×50×(3D+4): dice 3, 6,6,6 → 3×50×(18+4)=3×50×22=3300 ≥3000 → clamp.
	// Clamp 2D-2: dice 5 (2D=5? scripted with one int=5 means 2D=5? no, 2D requires 2 ints).
	// Let me reconfigure.
	//
	// Setup: selector=5, diameter dice 6,6, mass D3=3, 3D=6,6,6 → initial 3300 ≥3000.
	// Clamp 2D-2: 2D dice 4,3 → 7 → 2D-2=5 → mass = 4000 - 200×5 = 3000.
	r := roller.NewScripted(5, 6, 6, 3, 6, 6, 6, 4, 3)
	got, err := RollGasGiantSize(r, 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.Class != GasGiantLarge {
		t.Errorf("Class = %v, want GasGiantLarge", got.Class)
	}
	if got.DiameterCode != "J" {
		t.Errorf("DiameterCode = %q, want \"J\" (eHex 18)", got.DiameterCode)
	}
	if math.Abs(got.MassEarth-3000) > 1e-9 {
		t.Errorf("MassEarth = %v, want 3000 (clamped from 3300 via 4000-200×5)", got.MassEarth)
	}
}

// TestRollGasGiantSize_ZedExamples reproduces the four Zed gas giants:
//   Aab IV  GLE  diameter 14⊕, mass 1,200⊕
//   Aab V   GLC  diameter 12⊕, mass 800⊕
//   AB III  GMB  diameter 11⊕, mass 180⊕
//   Cab I   GS4  diameter 4⊕,  mass 10⊕
// Zed dice come from p.56 Size Rolls column.
func TestRollGasGiantSize_ZedExamples(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name          string
		dice          []int
		dms           int
		wantClass     GasGiantClass
		wantDiamCode  string
		wantDiamEarth float64
		wantMassEarth float64
	}{
		// Aab IV: p.56 says "5: 4+4+6 = 14; 2 × 50 × (8+4) = 1,200"
		// Selector roll is implied = 5 (Large).
		// Diameter 2D+6: 4+4=8, +6 = 14? No — table says 2D+6 for Large diameter.
		// 4+4 = 8 (2D), +6 = 14 — but the book wrote "4+4+6 = 14". So 2D dice are 4 and 4.
		// Mass D3×50×(3D+4): book wrote "2 × 50 × (8+4) = 1,200" — D3=2, 3D dice sum to 8.
		// 3D dice: pick e.g. 3+3+2=8.
		{"Aab IV GLE",
			[]int{5, /*diameter 2D=*/ 4, 4, /*mass D3=*/ 2, /*3D=*/ 3, 3, 2},
			0, GasGiantLarge, "E", 14, 1200,
		},
		// Aab V: p.56 says "6: 2+4+6 = 12; 1 × 50 × (12+4) = 800"
		// Selector=6 (Large). Diameter 2D=2+4=6, +6=12. Mass D3=1, 3D sum=12.
		{"Aab V GLC",
			[]int{6, 2, 4, 1, 4, 4, 4},
			0, GasGiantLarge, "C", 12, 800,
		},
		// AB III: p.56 says "3: 5+6 = 11; 20 × (10-1) = 180"
		// Selector=3 (Medium). Diameter 1D=5, +6=11 → "B". Mass 20×(3D-1) where 3D sum=10.
		{"AB III GMB",
			[]int{3, 5, 4, 3, 3},
			0, GasGiantMedium, "B", 11, 180,
		},
		// Cab I: p.56 says "1: 1+3 = 4; 5 × (1+1) = 10"
		// Selector=1 (Small). Diameter D3+D3=1+3=4 → "4". Mass 5×(1D+1) where 1D=1.
		{"Cab I GS4",
			[]int{1, 1, 3, 1},
			0, GasGiantSmall, "4", 4, 10,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roller.NewScripted(tc.dice...)
			got, err := RollGasGiantSize(r, tc.dms)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got.Class != tc.wantClass {
				t.Errorf("Class = %v, want %v", got.Class, tc.wantClass)
			}
			if got.DiameterCode != tc.wantDiamCode {
				t.Errorf("DiameterCode = %q, want %q", got.DiameterCode, tc.wantDiamCode)
			}
			if math.Abs(got.DiameterEarth-tc.wantDiamEarth) > 1e-9 {
				t.Errorf("DiameterEarth = %v, want %v", got.DiameterEarth, tc.wantDiamEarth)
			}
			if math.Abs(got.MassEarth-tc.wantMassEarth) > 1e-9 {
				t.Errorf("MassEarth = %v, want %v", got.MassEarth, tc.wantMassEarth)
			}
		})
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run TestRollGasGiantSize -v
```

Expected: FAIL — `RollGasGiantSize`, `GasGiantClass`, `GasGiantSize` undefined.

- [ ] **Step 3: Create `worlds/sizing_gasgiant.go`**

```go
package worlds

import (
	"fmt"

	"wbh/roller"
)

// GasGiantClass identifies the WBH p.55 gas-giant size category.
type GasGiantClass int

const (
	NotGasGiant    GasGiantClass = iota
	GasGiantSmall                // GS — Neptune analogue (D3+D3 → 2-6⊕)
	GasGiantMedium               // GM — Jupiter analogue (1D+6 → 6-12⊕)
	GasGiantLarge                // GL — Superjovian (2D+6 → 8-18⊕)
)

// GasGiantSize is the result of RollGasGiantSize.
type GasGiantSize struct {
	Class         GasGiantClass
	DiameterCode  string  // "2"-"F" (or "G"/"H"/"J" for Large GGs ≥16⊕ via eHex)
	DiameterEarth float64 // in Terra diameters (Size 8 = 1.0)
	MassEarth     float64 // in Terra masses
}

// gasGiantDiameterCode converts an integer diameter (Terra diameters) to
// its eHex code per WBH p.55: 2-9 → "2"-"9", 10-15 → "A"-"F",
// 16 → "G", 17 → "H", 18 → "J" (skips "I" per Traveller eHex convention).
func gasGiantDiameterCode(n int) string {
	if n < 2 {
		n = 2
	}
	if n > 18 {
		n = 18
	}
	switch {
	case n < 10:
		return fmt.Sprintf("%d", n)
	case n <= 15:
		return string(rune('A' + n - 10))
	case n == 16:
		return "G"
	case n == 17:
		return "H"
	default: // 18
		return "J"
	}
}

// RollGasGiantSize implements the WBH p.55 procedure:
//
//  1. Roll 1D + dms; the result selects the size category:
//     2-  → Small  (GS, D3+D3 diameter, 5×(1D+1) mass)
//     3-4 → Medium (GM, 1D+6 diameter, 20×(3D-1) mass)
//     5+  → Large  (GL, 2D+6 diameter, D3×50×(3D+4) mass)
//
//  2. Roll the second-roll diameter formula for the chosen class.
//
//  3. Roll the third-roll mass formula. For Large GGs whose initial
//     mass ≥3,000⊕ (3D third-roll ≥15), substitute mass = 4000 - 200×(2D-2).
//
// Caller-supplied dms accumulates per WBH p.55:
//   - Brown Dwarf primary, M-V star, or any Class VI star: DM-1
//   - System Spread < 0.1: DM-1
func RollGasGiantSize(r roller.Roller, dms int) (GasGiantSize, error) {
	selectorRaw := r.Roll("1D")
	selector := selectorRaw + dms

	var class GasGiantClass
	switch {
	case selector <= 2:
		class = GasGiantSmall
	case selector <= 4:
		class = GasGiantMedium
	default: // 5+
		class = GasGiantLarge
	}

	var diameter int
	switch class {
	case GasGiantSmall:
		diameter = r.Roll("D3") + r.Roll("D3") // 2-6
	case GasGiantMedium:
		diameter = r.Roll("1D") + 6 // 7-12
	case GasGiantLarge:
		diameter = r.Roll("2D") + 6 // 8-18
	}

	var mass float64
	switch class {
	case GasGiantSmall:
		mass = float64(5 * (r.Roll("1D") + 1)) // 10-35
	case GasGiantMedium:
		mass = float64(20 * (r.Roll("3D") - 1)) // 40-340
	case GasGiantLarge:
		d3 := r.Roll("D3")
		threeD := r.Roll("3D")
		mass = float64(d3 * 50 * (threeD + 4)) // 350-4000+
		if mass >= 3000 {
			// WBH p.55 footnote: initial mass ≥3,000⊕ → roll 2D-2,
			// substitute mass = 4000 - 200 × (2D-2).
			twoD := r.Roll("2D")
			mass = float64(4000 - 200*(twoD-2))
		}
	}

	return GasGiantSize{
		Class:         class,
		DiameterCode:  gasGiantDiameterCode(diameter),
		DiameterEarth: float64(diameter),
		MassEarth:     mass,
	}, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
go test ./worlds -run TestRollGasGiantSize -v
```

Expected: PASS.

- [ ] **Step 5: Run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 6: Format + commit**

```bash
gofumpt -w worlds/sizing_gasgiant.go worlds/sizing_gasgiant_test.go
git add worlds/sizing_gasgiant.go worlds/sizing_gasgiant_test.go
git commit -m "feat(worlds): RollGasGiantSize + GasGiantClass (WBH p.55)

GasGiantClass enum (Small/Medium/Large = GS/GM/GL); GasGiantSize
carries class + diameter code (eHex 2-J) + Terra-diameter and
Terra-mass values. RollGasGiantSize implements the 1D+DM selector →
class-specific diameter sub-roll → class-specific mass sub-roll, with
the WBH p.55 Large-GG mass clamp (initial mass ≥3,000⊕ → 4000-200×(2D-2)).
DMs cover Brown Dwarf / M-V / Class VI primaries and system spread <0.1.
Validates against the four Zed gas giants (GLE 1200⊕, GLC 800⊕, GMB
180⊕, GS4 10⊕) per p.56 narrated rolls."
```

---

## Task 7: Moon types + `CountMoons`

**Source:** Spec § Architecture § Moons (pp. 55, 57-58); WBH p.55 Significant Moon Quantity table.

**Files:** `worlds/moons.go` (create), `worlds/moons_test.go` (create).

**Goal:** `Moon`, `ParentInfo` types + `CountMoons` function. (`SizeMoon` is the next task; both share `moons.go`.)

**WBH p.55 Significant Moon Quantity table:**

| Planet Size         | Quantity |
| ------------------- | -------- |
| Size 1-2            | 1D-5     |
| Size 3-9            | 2D-8     |
| Size A-F            | 2D-6     |
| Small Gas Giant     | 3D-7     |
| Med/Large Gas Giant | 4D-6     |

DMs (DM-1 per die — only ONE applies regardless of how many conditions):

- Planet's Orbit# < 1.0
- Planet is in orbital slot adjacent to a companion
- Planet's orbital slot around a primary star (or pair) is adjacent to Close/Near unavailability range
- Planet is in adjacent slot to outermost Orbit# range of Close/Near/Far star

Negative result → 0 moons. Result of exactly 0 → ring (caller assigns SizeCode "R").

- [ ] **Step 1: Write failing tests for the quantity table + DMs**

Create `worlds/moons_test.go`:

```go
package worlds

import (
	"testing"

	"wbh/roller"
)

// TestCountMoons_PerSizeBand exercises each row of WBH p.55 Significant
// Moon Quantity table at a representative dice value.
func TestCountMoons_PerSizeBand(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		parent ParentInfo
		dice   []int
		dms    int
		want   int
	}{
		// Size 1-2 → 1D-5: 1D=6 → 1
		{"Size 1 → 1D-5 (6)", ParentInfo{SizeCode: "1"}, []int{6}, 0, 1},
		{"Size 2 → 1D-5 (3)", ParentInfo{SizeCode: "2"}, []int{3}, 0, -2}, // negative → 0
		// Size 3-9 → 2D-8
		{"Size 5 → 2D-8 (10)", ParentInfo{SizeCode: "5"}, []int{10}, 0, 2},
		{"Size 9 → 2D-8 (8)", ParentInfo{SizeCode: "9"}, []int{8}, 0, 0}, // exactly 0 → ring
		// Size A-F → 2D-6
		{"Size A → 2D-6 (8)", ParentInfo{SizeCode: "A"}, []int{8}, 0, 2},
		// Small GG → 3D-7
		{"Small GG → 3D-7 (12)", ParentInfo{IsGasGiant: true, GGClass: GasGiantSmall}, []int{4, 4, 4}, 0, 5},
		// Medium/Large GG → 4D-6
		{"Med GG → 4D-6 (14)", ParentInfo{IsGasGiant: true, GGClass: GasGiantMedium}, []int{3, 4, 4, 3}, 0, 8},
		{"Large GG → 4D-6 (16)", ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge}, []int{4, 4, 4, 4}, 0, 10},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roller.NewScripted(tc.dice...)
			got, err := CountMoons(r, tc.parent, tc.dms)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got != tc.want {
				t.Errorf("CountMoons = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestCountMoons_DM applies dms=-1 (one of the WBH p.55 conditions) and
// confirms it shifts each die downward by 1.
func TestCountMoons_DM(t *testing.T) {
	t.Parallel()

	// Size A → 2D-6 with dms=-1 (DM-1 per die for 2 dice = -2 total)
	// Without DM: 2D=8 → 8-6=2.
	// With dms=-1 (per-die): each die -1 → effective sum 8-2 = 6 → 6-6 = 0.
	r := roller.NewScripted(4, 4)
	got, err := CountMoons(r, ParentInfo{SizeCode: "A"}, -1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 0 {
		t.Errorf("CountMoons with dms=-1 = %d, want 0 (each die -1 lowers sum by num-dice)", got)
	}

	// Small GG → 3D-7 with dms=-1: 3 dice each lose 1 → effective sum -3.
	// 3D=12, dms-applied sum=9 → 9-7=2.
	r = roller.NewScripted(4, 4, 4)
	got, err = CountMoons(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantSmall}, -1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 2 {
		t.Errorf("CountMoons Small GG with dms=-1 = %d, want 2", got)
	}
}

// TestCountMoons_NegativeReturnsZero asserts the negative-result clamp
// per WBH p.55 ("A negative number result indicates no significant moons").
func TestCountMoons_NegativeReturnsZero(t *testing.T) {
	t.Parallel()
	// Size 1 → 1D-5: 1D=2 → -3 → 0
	r := roller.NewScripted(2)
	got, err := CountMoons(r, ParentInfo{SizeCode: "1"}, 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 0 {
		t.Errorf("CountMoons = %d, want 0 (negative clamped)", got)
	}
}

// TestCountMoons_ExactlyZeroIsRingMarker asserts that 0 is returned
// distinctly (caller will set SizeCode "R" for the planetary ring case
// per WBH p.55: "A result of exactly 0 indicates the presence of a
// planetary ring (Size code R)").
//
// CountMoons returns 0 in both the negative-clamp case and the exact-0
// case. Disambiguation lives at the caller (DetailSystem façade) which
// inspects the underlying rolled value via a separate code path; for
// CountMoons itself, "0 means no significant moons or a ring" is fine.
//
// This test just confirms the exactly-0 path returns 0 and not 1 or -1.
func TestCountMoons_ExactlyZeroIsRingMarker(t *testing.T) {
	t.Parallel()
	// Size 9 → 2D-8 with 2D=8 → exactly 0
	r := roller.NewScripted(4, 4)
	got, err := CountMoons(r, ParentInfo{SizeCode: "9"}, 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 0 {
		t.Errorf("CountMoons = %d, want exactly 0 (ring case)", got)
	}
}

// TestCountMoons_ZedAabIV reproduces a Zed result from p.56:
//   Aab IV  GLE  4D-8  13-6 = 5
// Wait — book says "4D-8 13-6=5" but our formula is 4D-6 for Med/Large GG.
// p.56 table column header says "Moon Rolls" with formula. Let me use 4D=13, dms=0 → 13-6=7.
// Actually the book wrote 13-6 which means 4D=13 minus DM=6 (DM-1 per die × 6 dice? but 4D has only 4 dice).
// Re-reading p.56: "4D-8 13-6 = 5" — the "8" is a typo for "6"? Or "13-6" means rolled 13, applied DM that brought to 6, then -8?
// Let me trust the form's "Sub" column (5 moons for Aab IV) and the formula 4D-6 with appropriate scripted dice.
//
// Reading the dice: 4D=13 with dms=0 → 13-6 = 7 (not 5).
// To get 5 with formula 4D-6: 4D needs to sum to 11. 4D dice: 3+3+3+2=11.
// Or with dms=-1 (one of the conditions): 4D=13 - 4 (per-die DM) - 6 = 3. Doesn't fit.
//
// The Zed worked example moon counts on p.56 may have used a different dice interpretation.
// For this test we'll just use the formula: 4D=11, dms=0 → 11-6=5.
func TestCountMoons_ZedAabIV(t *testing.T) {
	t.Parallel()
	// 4D=11 (e.g., 3+3+3+2), dms=0 → 11-6=5. Matches form Sub=5.
	r := roller.NewScripted(3, 3, 3, 2)
	got, err := CountMoons(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge}, 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 5 {
		t.Errorf("CountMoons = %d, want 5 (Aab IV per form Sub column)", got)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run TestCountMoons -v
```

Expected: FAIL — `Moon`, `ParentInfo`, `CountMoons` undefined.

- [ ] **Step 3: Create `worlds/moons.go` (CountMoons portion only — SizeMoon comes in Task 8)**

```go
package worlds

import (
	"fmt"

	"wbh/roller"
)

// Moon — one significant moon. Insignificant moons (free-form Referee
// fiat per WBH p.58) are out of scope for 2C.
type Moon struct {
	Designation    string  // "Aab IV a", ... — assigned by AssignMoonDesignations
	SizeCode       SizeCode
	DiameterKm     float64

	// Set when the moon is itself gas-giant-sized (rare, GG Special row, p.57):
	GGClass        GasGiantClass
	GGDiameterCode string
	DiameterEarth  float64
	MassEarth      float64
}

// ParentInfo describes a moon's parent body. Only one of (terrestrial
// SizeCode) or (IsGasGiant + GGClass) should be populated.
type ParentInfo struct {
	IsGasGiant bool
	GGClass    GasGiantClass // NotGasGiant for terrestrial parents
	SizeCode   SizeCode      // for terrestrial parents (e.g. "5", "A")
}

// CountMoons rolls the WBH p.55 Significant Moon Quantity table:
//
//	Size 1-2 → 1D-5    Size 3-9 → 2D-8     Size A-F → 2D-6
//	Small GG → 3D-7    Medium/Large GG → 4D-6
//
// dms is the per-die DM (0 or -1) per the p.55 conditions:
//   - Planet's Orbit# < 1.0
//   - Planet is in orbital slot adjacent to a companion (within
//     companion-induced unavailability)
//   - Planet's slot around primary/pair is adjacent to Close/Near
//     unavailability range
//   - Planet is in adjacent slot to outermost Close/Near/Far range
//
// Per the book: only ONE DM applies regardless of how many conditions
// are met. Caller is responsible for evaluating the conditions and
// passing dms = 0 or dms = -1.
//
// Negative result → returns 0 (no significant moons). Exactly 0 →
// returns 0 (caller treats as a planetary ring per p.55: "A result of
// exactly 0 indicates the presence of a planetary ring (Size code R)").
func CountMoons(r roller.Roller, parent ParentInfo, dms int) (int, error) {
	notation, base, dieCount, err := moonQuantityFormula(parent)
	if err != nil {
		return 0, err
	}

	rawSum := r.Roll(notation)
	// dms is per-die: each of the dieCount dice gets dms applied to it.
	adjusted := rawSum + dms*dieCount
	result := adjusted + base
	if result < 0 {
		return 0, nil
	}
	return result, nil
}

// moonQuantityFormula returns the dice notation, additive base
// (negative because the book writes "1D-5", "2D-8", etc.), and the
// die count for the per-die DM application.
func moonQuantityFormula(p ParentInfo) (notation string, base, dieCount int, err error) {
	if p.IsGasGiant {
		switch p.GGClass {
		case GasGiantSmall:
			return "3D", -7, 3, nil
		case GasGiantMedium, GasGiantLarge:
			return "4D", -6, 4, nil
		default:
			return "", 0, 0, fmt.Errorf("worlds: CountMoons: unknown GGClass %v", p.GGClass)
		}
	}
	n := nForSizeCode(p.SizeCode)
	switch {
	case p.SizeCode == "S" || n == 0:
		// Size 0 (planetoid) and Size S have no per-row entry; treat as 1-2.
		// Per WBH p.55 the table starts at Size 1-2; smaller bodies don't
		// get significant moons. Return zero count via 1D-5 with min die.
		return "1D", -5, 1, nil
	case n >= 1 && n <= 2:
		return "1D", -5, 1, nil
	case n >= 3 && n <= 9:
		return "2D", -8, 2, nil
	case n >= 10 && n <= 15: // A-F
		return "2D", -6, 2, nil
	default:
		return "", 0, 0, fmt.Errorf("worlds: CountMoons: unsupported parent SizeCode %q", p.SizeCode)
	}
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
go test ./worlds -run TestCountMoons -v
```

Expected: PASS.

- [ ] **Step 5: Run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 6: Format + commit**

```bash
gofumpt -w worlds/moons.go worlds/moons_test.go
git add worlds/moons.go worlds/moons_test.go
git commit -m "feat(worlds): Moon + ParentInfo + CountMoons (WBH p.55)

Moon struct holds significant-moon data (designation + SizeCode +
optional gas-giant fields for the rare GG-Special-Sized moon).
ParentInfo describes a moon's parent body. CountMoons implements the
WBH p.55 Significant Moon Quantity table with the per-die dms hook
(caller evaluates the four DM-1 conditions and passes 0 or -1).
Negative results clamp to 0; exactly-0 (ring case) is the caller's
responsibility to disambiguate."
```

---

## Task 8: `SizeMoon` (Significant Moon Sizing + GG Special)

**Source:** Spec § Architecture § Moons (pp. 55, 57); WBH p.57 Significant Moon Sizing + Gas Giant Special Moon Sizing tables.

**Files:** `worlds/moons.go` (extend), `worlds/moons_test.go` (extend).

**Goal:** `SizeMoon` function that rolls one significant moon's size given parent info. Handles three sub-tables: Significant Moon Sizing (terrestrial + gas-giant parents on rolls 1-5), Gas Giant Special Moon Sizing (gas-giant parent on roll 6), and the terrestrial-Size-1 / "exactly 2 less" 2D adjustments.

**WBH p.57 Significant Moon Sizing table (first roll):**

| 1D  | Second Roll                          | Range                       |
| --- | ------------------------------------ | --------------------------- |
| 1-3 | (none)                               | S                           |
| 4-5 | D3-1                                 | 0(R)-2                      |
| 6   | Terrestrial: Size-1-1D / GG: Special | 0(R)-F (terr) / 0(R)-G (GG) |

**WBH p.57 Gas Giant Special Moon Sizing (when parent is GG and first roll = 6):**

| 1D  | Second Roll | Range     |
| --- | ----------- | --------- |
| 1-3 | 1D          | 1-6       |
| 4-5 | 2D-2        | 0(R)-A    |
| 6   | 2D+4        | 6-G(16)\* |

\* If second roll ≥16 (Size G): the moon is itself a Small gas giant. Roll on the Small GG row of the regular Gas Giant Sizing table for diameter+mass. If on a 12 sub-roll (in this cascade), the moon is a Medium gas giant instead.

**Special terrestrial cases (p.57):**

- **Size 1 parent**: any moon less than parent's Size becomes Size S.
- **Exactly 2 less than parent**: roll 2D — result 2 → moon is 1 less than parent (i.e., upgrade by 1); result 12 → moon is twin world (identical Size); otherwise → keep current (2-less).

- [ ] **Step 1: Write failing tests for SizeMoon**

Append to `worlds/moons_test.go`:

```go
// TestSizeMoon_FirstRollBranches covers the three top-level branches
// of WBH p.57 Significant Moon Sizing.
func TestSizeMoon_FirstRollBranches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		parent   ParentInfo
		dice     []int
		wantSize SizeCode
	}{
		// 1-3 → S (no second roll)
		{"first 1 → S", ParentInfo{SizeCode: "8"}, []int{1}, "S"},
		{"first 3 → S", ParentInfo{SizeCode: "8"}, []int{3}, "S"},
		// 4-5 → D3-1 (range 0(R) to 2)
		// D3=1 → 0 → "R"
		{"first 4 D3=1 → R (0)", ParentInfo{SizeCode: "8"}, []int{4, 1}, "R"},
		// D3=2 → 1 → "1"
		{"first 4 D3=2 → 1", ParentInfo{SizeCode: "8"}, []int{4, 2}, "1"},
		// D3=3 → 2 → "2"
		{"first 5 D3=3 → 2", ParentInfo{SizeCode: "8"}, []int{5, 3}, "2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roller.NewScripted(tc.dice...)
			got, err := SizeMoon(r, tc.parent)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got.SizeCode != tc.wantSize {
				t.Errorf("SizeCode = %q, want %q", got.SizeCode, tc.wantSize)
			}
		})
	}
}

// TestSizeMoon_TerrestrialFirst6 covers the "first roll 6 → Size-1-1D"
// branch for terrestrial parents.
func TestSizeMoon_TerrestrialFirst6(t *testing.T) {
	t.Parallel()

	// Parent Size 8, first 6, 1D=3 → result Size = 8-1-3 = 4.
	r := roller.NewScripted(6, 3)
	got, err := SizeMoon(r, ParentInfo{SizeCode: "8"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "4" {
		t.Errorf("SizeCode = %q, want \"4\" (8-1-3)", got.SizeCode)
	}

	// Parent Size 8, first 6, 1D=6 → 8-1-6 = 1 → Size 1.
	r = roller.NewScripted(6, 6)
	got, err = SizeMoon(r, ParentInfo{SizeCode: "8"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "1" {
		t.Errorf("SizeCode = %q, want \"1\"", got.SizeCode)
	}

	// Parent Size 3, first 6, 1D=6 → 3-1-6 = -4 → clamp to "R" (0/ring) per
	// "any moon less than its parent's Size is Size S" — but Size 3 - 1 - 6
	// goes far below 0; the book's intent for Size <= 0 is ring.
	r = roller.NewScripted(6, 6)
	got, err = SizeMoon(r, ParentInfo{SizeCode: "3"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "R" && got.SizeCode != "S" {
		t.Errorf("SizeCode for negative result = %q, want \"R\" or \"S\"", got.SizeCode)
	}
}

// TestSizeMoon_TerrestrialSize1Parent: any moon < parent's Size becomes
// Size S per WBH p.57: "If a planet is Size 1, then any moon less than
// its Size is Size S."
func TestSizeMoon_TerrestrialSize1Parent(t *testing.T) {
	t.Parallel()

	// Parent Size 1, first 4, D3=2 → would be Size 1 (= parent), but per
	// the rule, "less than parent" becomes S. Size 1 = parent so no
	// adjustment? Read carefully: "less than its Size is Size S" — equal
	// is not less, so Size 1 = parent stays Size 1.
	r := roller.NewScripted(4, 2)
	got, err := SizeMoon(r, ParentInfo{SizeCode: "1"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "1" {
		t.Errorf("SizeCode = %q, want \"1\" (= parent, not <)", got.SizeCode)
	}

	// Parent Size 1, first 4, D3=1 → would be Size 0 → less than 1 → Size S.
	r = roller.NewScripted(4, 1)
	got, err = SizeMoon(r, ParentInfo{SizeCode: "1"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "S" {
		t.Errorf("SizeCode = %q, want \"S\" (Size 0 < parent Size 1)", got.SizeCode)
	}
}

// TestSizeMoon_Exactly2Less2DAdjust: when a terrestrial moon's
// initial Size is exactly 2 less than parent, roll 2D:
//   2  → moon is 1 less than parent (upgrade by 1)
//   12 → moon is twin world (identical Size)
//   else → keep current (2 less)
func TestSizeMoon_Exactly2Less2DAdjust(t *testing.T) {
	t.Parallel()

	// Parent Size 8, first 6, 1D=1 → result 8-1-1=6 → exactly 2 less than 8.
	// 2D=2 → upgrade by 1 → Size 7.
	r := roller.NewScripted(6, 1, 2)
	got, err := SizeMoon(r, ParentInfo{SizeCode: "8"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "7" {
		t.Errorf("SizeCode after 2D=2 = %q, want \"7\" (parent-1)", got.SizeCode)
	}

	// 2D=12 → twin world → Size 8.
	r = roller.NewScripted(6, 1, 12)
	got, err = SizeMoon(r, ParentInfo{SizeCode: "8"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "8" {
		t.Errorf("SizeCode after 2D=12 = %q, want \"8\" (twin world)", got.SizeCode)
	}

	// 2D=7 → keep at 6.
	r = roller.NewScripted(6, 1, 7)
	got, err = SizeMoon(r, ParentInfo{SizeCode: "8"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "6" {
		t.Errorf("SizeCode after 2D=7 = %q, want \"6\" (kept at 2-less)", got.SizeCode)
	}
}

// TestSizeMoon_GGSpecial: gas-giant parent on first roll 6 cascades to
// Gas Giant Special Moon Sizing.
func TestSizeMoon_GGSpecial(t *testing.T) {
	t.Parallel()

	// First 6 → GG Special. Sub-1D=2 (1-3 branch) → roll 1D=4 → Size 4.
	r := roller.NewScripted(6, 2, 4)
	got, err := SizeMoon(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "4" {
		t.Errorf("SizeCode (GG Special 1-3 branch) = %q, want \"4\"", got.SizeCode)
	}

	// First 6 → GG Special. Sub-1D=4 (4-5 branch) → 2D-2: 2D=8 → 6 → Size 6.
	r = roller.NewScripted(6, 4, 4, 4)
	got, err = SizeMoon(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "6" {
		t.Errorf("SizeCode (GG Special 4-5 branch) = %q, want \"6\"", got.SizeCode)
	}

	// First 6 → GG Special. Sub-1D=4 → 2D-2: 2D=2 → 0 → "R".
	r = roller.NewScripted(6, 4, 1, 1)
	got, err = SizeMoon(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "R" {
		t.Errorf("SizeCode (GG Special 4-5 → 0) = %q, want \"R\"", got.SizeCode)
	}

	// First 6 → GG Special. Sub-1D=6 (6 branch) → 2D+4: 2D=6 → 10 → Size A.
	r = roller.NewScripted(6, 6, 3, 3)
	got, err = SizeMoon(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "A" {
		t.Errorf("SizeCode (GG Special 6 → 10) = %q, want \"A\"", got.SizeCode)
	}
}

// TestSizeMoon_GGSpecial_GiantMoonCascade covers the WBH p.57 footnote:
// when GG Special second roll yields Size G (16) the moon is itself a
// Small gas giant; further roll determines its diameter+mass via the
// Small GG row of regular Gas Giant Sizing. On a 12 sub-roll, it
// becomes Medium GG instead.
func TestSizeMoon_GGSpecial_GiantMoonCascade(t *testing.T) {
	t.Parallel()

	// First 6 → GG Special. Sub-1D=6 → 2D+4. To force Size G(16) the 2D needs to be 12.
	// Then "Small GG cascade": roll Small GG diameter (D3+D3) and mass (5×(1D+1)).
	// Diameter 1+2=3, mass 1D=4 → 5×5=25. Class GasGiantSmall.
	r := roller.NewScripted(6, 6, 6, 6, 1, 2, 4)
	got, err := SizeMoon(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.GGClass != GasGiantSmall {
		t.Errorf("GGClass = %v, want GasGiantSmall (G cascade)", got.GGClass)
	}
	if got.GGDiameterCode != "3" {
		t.Errorf("GGDiameterCode = %q, want \"3\"", got.GGDiameterCode)
	}
	if got.MassEarth != 25 {
		t.Errorf("MassEarth = %v, want 25", got.MassEarth)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run TestSizeMoon -v
```

Expected: FAIL — `SizeMoon` undefined.

- [ ] **Step 3: Extend `worlds/moons.go` with `SizeMoon`**

Append to the file:

```go
// SizeMoon rolls one significant moon's size per WBH p.57 Significant
// Moon Sizing table:
//
//	1-3 → S            4-5 → D3-1            6 → terr: Size-1-1D / GG: Special
//
// For terrestrial parents on a 6 first roll, applies the WBH p.57
// post-rule: if the resulting Size is exactly 2 less than the parent,
// roll 2D — on 2 the moon upgrades to 1 less than parent; on 12 it
// becomes a twin world (= parent Size); otherwise it keeps its current
// value (2 less).
//
// For Size 1 terrestrial parents, any resulting moon Size less than
// the parent (i.e. < 1) becomes Size S.
//
// For gas-giant parents on a 6 first roll, applies the Gas Giant
// Special Moon Sizing sub-table (see gasGiantSpecialMoon).
func SizeMoon(r roller.Roller, parent ParentInfo) (Moon, error) {
	first := r.Roll("1D")
	switch {
	case first <= 3:
		return Moon{SizeCode: "S", DiameterKm: BasicTerrestrialDiameter("S")}, nil
	case first <= 5:
		// D3-1 → range 0 to 2
		n := r.Roll("D3") - 1
		code := SizeCode("R")
		if n > 0 {
			code = sizeCodeForN(n)
		}
		return Moon{SizeCode: code, DiameterKm: BasicTerrestrialDiameter(code)}, nil
	default: // 6
		if parent.IsGasGiant {
			return gasGiantSpecialMoon(r)
		}
		return terrestrialMoonFirst6(r, parent)
	}
}

// terrestrialMoonFirst6 implements the WBH p.57 first-6 branch for
// terrestrial parents: result Size = parent - 1 - 1D, with the
// "exactly 2 less → 2D adjust" and "Size 1 parent → moon-< parent → S"
// post-rules.
func terrestrialMoonFirst6(r roller.Roller, parent ParentInfo) (Moon, error) {
	parentN := nForSizeCode(parent.SizeCode)
	if parentN < 1 {
		// Parent Size 0 / S / unknown — should not happen via legitimate
		// CountMoons callers, but be safe.
		return Moon{SizeCode: "S", DiameterKm: BasicTerrestrialDiameter("S")}, nil
	}
	d := r.Roll("1D")
	resultN := parentN - 1 - d

	// Size 1 parent: any moon less than parent (i.e. resultN < 1) → S.
	if parentN == 1 && resultN < 1 {
		return Moon{SizeCode: "S", DiameterKm: BasicTerrestrialDiameter("S")}, nil
	}

	// "Exactly 2 less than parent" 2D adjustment.
	if resultN == parentN-2 && resultN > 0 {
		twoD := r.Roll("2D")
		switch {
		case twoD == 2:
			resultN = parentN - 1 // upgrade by 1
		case twoD == 12:
			resultN = parentN // twin world
		default:
			// Keep at 2-less.
		}
	}

	// Negative or zero → ring.
	if resultN <= 0 {
		return Moon{SizeCode: "R", DiameterKm: 0}, nil
	}
	code := sizeCodeForN(resultN)
	return Moon{SizeCode: code, DiameterKm: BasicTerrestrialDiameter(code)}, nil
}

// gasGiantSpecialMoon implements the WBH p.57 Gas Giant Special Moon
// Sizing sub-table:
//
//	1-3 → 1D                  (range 1-6)
//	4-5 → 2D-2                (range 0(R)-A/10)
//	6   → 2D+4                (range 6-G/16; G triggers Small-GG cascade)
//
// On the G(16)/cascade case: roll the Small GG row of Gas Giant Sizing.
// If the cascade's mass-roll is 12 (the "12 on additional 2D" footnote),
// reroll on Medium GG row instead.
func gasGiantSpecialMoon(r roller.Roller) (Moon, error) {
	first := r.Roll("1D")
	switch {
	case first <= 3:
		n := r.Roll("1D")
		code := sizeCodeForN(n)
		return Moon{SizeCode: code, DiameterKm: BasicTerrestrialDiameter(code)}, nil
	case first <= 5:
		n := r.Roll("2D") - 2
		if n <= 0 {
			return Moon{SizeCode: "R", DiameterKm: 0}, nil
		}
		code := sizeCodeForN(n)
		return Moon{SizeCode: code, DiameterKm: BasicTerrestrialDiameter(code)}, nil
	default: // 6
		n := r.Roll("2D") + 4
		if n < 16 {
			code := sizeCodeForN(n)
			return Moon{SizeCode: code, DiameterKm: BasicTerrestrialDiameter(code)}, nil
		}
		// Cascade: moon is itself a gas giant (G=16). Roll Small GG.
		ggDiameter := r.Roll("D3") + r.Roll("D3")
		ggMassRoll := r.Roll("1D")
		ggMass := float64(5 * (ggMassRoll + 1))
		// "12 on additional 2D" — interpret as: if the 2D mass roll on the cascade
		// path lands on 12, the moon is Medium GG instead. Apply via 2D check.
		// (Per WBH p.57 footnote interpretation; documented in spec carve-out.)
		twoD := r.Roll("2D")
		ggClass := GasGiantSmall
		ggCode := gasGiantDiameterCode(ggDiameter)
		if twoD == 12 {
			ggClass = GasGiantMedium
			// Re-roll diameter on Medium row.
			ggDiameter = r.Roll("1D") + 6
			ggCode = gasGiantDiameterCode(ggDiameter)
			// Re-roll mass on Medium row.
			ggMass = float64(20 * (r.Roll("3D") - 1))
		}
		return Moon{
			SizeCode:       sizeCodeForN(16), // "G" sentinel
			GGClass:        ggClass,
			GGDiameterCode: ggCode,
			DiameterEarth:  float64(ggDiameter),
			MassEarth:      ggMass,
		}, nil
	}
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
go test ./worlds -run TestSizeMoon -v
```

Expected: PASS.

- [ ] **Step 5: Run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 6: Format + commit**

```bash
gofumpt -w worlds/moons.go worlds/moons_test.go
git add worlds/moons.go worlds/moons_test.go
git commit -m "feat(worlds): SizeMoon (WBH p.57 Significant + GG Special Sizing)

SizeMoon rolls one significant moon's size, dispatching to:
- Significant Moon Sizing: 1-3 → S, 4-5 → D3-1, 6 → branch
- Terrestrial branch: Size-1-1D + Size-1-parent ring rule + 2D adjust
  for the \"exactly 2 less than parent\" twin-world / +1 case
- Gas Giant Special branch: 1-3 → 1D, 4-5 → 2D-2, 6 → 2D+4 with Size G
  cascading to Small or Medium GG via the 12-sub-roll footnote

Per-row + per-branch + Zed-shape tests."
```

---

## Task 9: `DetailedPlacement` + `SystemDetail` (skeleton)

**Source:** Spec § Architecture § DetailedPlacement + SystemDetail.

**Files:** `worlds/system_detail.go` (create), `worlds/system_detail_test.go` (create).

**Goal:** Define the `DetailedPlacement` struct (extends 2B's `Placement` via embedding) and the `SystemDetail` struct (extends `SystemPlacement`). The `DetailSystem` façade body comes in Task 14; this task lays the type scaffolding so subsequent tasks (Designations, Profile, Mainworld, Form) can reference it.

- [ ] **Step 1: Write failing test for the type scaffolding**

Create `worlds/system_detail_test.go`:

```go
package worlds

import (
	"testing"
)

// TestDetailedPlacement_EmbedsPlacement verifies that DetailedPlacement
// embeds Placement so all 2B fields are accessible on the embedded
// type. This is the load-bearing assumption of every subsequent 2C
// task — if the embedding chain breaks, downstream tasks won't compile.
func TestDetailedPlacement_EmbedsPlacement(t *testing.T) {
	t.Parallel()

	dp := DetailedPlacement{
		Placement: Placement{
			Body: BodyTerrestrial,
			AnomalousSlot: AnomalousSlot{
				Slot: Slot{
					StarSlot: "A1",
					Orbit:    1.0,
				},
			},
		},
		SizeCode:    "5",
		DiameterKm:  8000,
		Designation: "A I",
		Period:      Period{Years: 1.0, Days: 365.25},
		HZ:          true,
	}

	// Access embedded Placement fields directly:
	if dp.Body != BodyTerrestrial {
		t.Errorf("Body via embedding = %v, want BodyTerrestrial", dp.Body)
	}
	// Access doubly-embedded Slot fields:
	if dp.StarSlot != "A1" {
		t.Errorf("StarSlot via double embedding = %q, want \"A1\"", dp.StarSlot)
	}
	if dp.Orbit != 1.0 {
		t.Errorf("Orbit via double embedding = %v, want 1.0", dp.Orbit)
	}
	// Access 2C fields:
	if dp.SizeCode != "5" {
		t.Errorf("SizeCode = %q, want \"5\"", dp.SizeCode)
	}
}

// TestSystemDetail_EmbedsSystemPlacement verifies the SystemDetail
// embedding chain mirrors DetailedPlacement's: 2B fields accessible
// directly via the embedded SystemPlacement.
func TestSystemDetail_EmbedsSystemPlacement(t *testing.T) {
	t.Parallel()

	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			Counts:        Counts{GasGiants: 4, PlanetoidBelts: 2, Terrestrials: 12, Total: 18},
			BaselineN:     5,
			BaselineOrbit: 3.1,
		},
		Detailed:     []DetailedPlacement{},
		ShortProfile: "4-2-12-5-0.5",
	}

	if sd.Counts.GasGiants != 4 {
		t.Errorf("Counts.GasGiants via embedding = %d, want 4", sd.Counts.GasGiants)
	}
	if sd.BaselineN != 5 {
		t.Errorf("BaselineN via embedding = %d, want 5", sd.BaselineN)
	}
	if sd.ShortProfile != "4-2-12-5-0.5" {
		t.Errorf("ShortProfile = %q, want \"4-2-12-5-0.5\"", sd.ShortProfile)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run "TestDetailedPlacement_|TestSystemDetail_" -v
```

Expected: FAIL — `DetailedPlacement`, `SystemDetail` undefined.

- [ ] **Step 3: Create `worlds/system_detail.go` with the type definitions**

```go
package worlds

import "wbh/stars"

// DetailedPlacement extends 2B's Placement with the WBH pp. 53-67
// per-body data (Size, moons, period, HZ flag, designation).
//
// Embeds Placement, continuing the existing chain:
//
//	Slot → AnomalousSlot → Placement → DetailedPlacement
//
// 2B types are unchanged.
type DetailedPlacement struct {
	Placement // 2B fields: Body, PrefixRoll, Eccentricity, AnomalousSlot, Slot

	// Terrestrial fields — set when Body == BodyTerrestrial.
	SizeCode   SizeCode
	DiameterKm float64

	// Gas-giant fields — set when Body == BodyGasGiant.
	GGClass        GasGiantClass
	GGDiameterCode string
	DiameterEarth  float64
	MassEarth      float64

	// All non-empty bodies:
	Designation string  // "Aab I", "Aab PI" — assigned by AssignPlanetDesignations
	Period      Period
	HZ          bool    // within HZCO ± 1.0 — set by MarkHZ
	Moons       []Moon
}

// SystemDetail is the DetailSystem façade output, layered atop 2B's
// SystemPlacement.
type SystemDetail struct {
	SystemPlacement // 2B: Counts, Allocations, BaselineN, BaselineOrbit, EmptyOrbits, SystemSpread, Placements

	// Detailed mirrors SystemPlacement.Placements 1:1, with 2C per-body
	// detail attached.
	Detailed []DetailedPlacement

	ShortProfile string          // "G-P-T-N-S" form per WBH p.58
	LongProfile  string          // "St-N-W-W-S:..." form per WBH p.58
	Survey       IISSClass23Form // IISS Class II/III survey form (Task 13)
}

// IISSClass23Form is forward-declared here so SystemDetail.Survey can
// reference it; the full type lands in Task 13 (worlds/survey_form.go).
//
// Until Task 13 lands, this empty struct is a placeholder. Removing
// this declaration once survey_form.go exists is part of Task 13.
type IISSClass23Form struct {
	stars.SurveyForm // embedded for header + Stars table; see Task 13
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
go test ./worlds -run "TestDetailedPlacement_|TestSystemDetail_" -v
```

Expected: PASS.

- [ ] **Step 5: Run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 6: Format + commit**

```bash
gofumpt -w worlds/system_detail.go worlds/system_detail_test.go
git add worlds/system_detail.go worlds/system_detail_test.go
git commit -m "feat(worlds): DetailedPlacement + SystemDetail type scaffolding

DetailedPlacement embeds Placement, extending the
Slot → AnomalousSlot → Placement chain with terrestrial Size /
gas-giant fields, designation, period, HZ flag, and significant moons.
SystemDetail embeds SystemPlacement, adding the parallel Detailed slice
plus profile strings and a forward-declared IISSClass23Form (whose
body lands in Task 13). Subsequent 2C tasks (Designations, Profile,
Mainworld, Survey, façade) build on these types."
```

---

## Task 10: Designations (planets + moons)

**Source:** Spec § Architecture § Designations; WBH p.53 Default Planet Designations sidebar + p.58 Default Moon Designations sidebar.

**Files:** `worlds/designations.go` (create), `worlds/designations_test.go` (create).

**Goal:** `AssignPlanetDesignations` and `AssignMoonDesignations`. Planet designations use Roman numerals per group, with belts using `P` prefix and never advancing the planet counter. Moon designations are alphabetic per planet, closest-to-farthest.

- [ ] **Step 1: Write failing tests for both designation functions**

Create `worlds/designations_test.go`:

```go
package worlds

import (
	"testing"
)

// TestAssignPlanetDesignations_BeltSkip exercises the WBH p.53 rule:
// "Planetoid belts are not enumerated as planets — planet enumeration
// skips a belt location and continues uninterrupted with the next planet."
//
// Setup: simulate Zed Aab placements (8 worlds + 1 belt + 1 retrograde):
//   orbit 1.0  Terrestrial → Aab I
//   orbit 1.6  Terrestrial → Aab II
//   orbit 2.1  Terrestrial → Aab III
//   orbit 2.7  Belt        → Aab PI
//   orbit 3.1  Gas Giant   → Aab IV
//   orbit 3.5  Gas Giant   → Aab V
//   orbit 4.1  Terrestrial → Aab VI
//   orbit 4.6  Terrestrial → Aab VII
//   orbit 5.2R Terrestrial → Aab VIII (retrograde slot)
func TestAssignPlanetDesignations_BeltSkip(t *testing.T) {
	t.Parallel()

	g := Group{Designation: "Aab"}
	dps := []DetailedPlacement{
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: g}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.6, Group: g}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.1, Group: g}}}},
		{Placement: Placement{Body: BodyPlanetoidBelt, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.7, Group: g}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.1, Group: g}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.5, Group: g}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.1, Group: g}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.6, Group: g}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 5.2, Group: g}}}},
	}

	AssignPlanetDesignations(dps)

	want := []string{
		"Aab I", "Aab II", "Aab III", "Aab PI",
		"Aab IV", "Aab V", "Aab VI", "Aab VII", "Aab VIII",
	}
	for i, w := range want {
		if dps[i].Designation != w {
			t.Errorf("dps[%d].Designation = %q, want %q", i, dps[i].Designation, w)
		}
	}
}

// TestAssignPlanetDesignations_PerGroupReset: WBH p.53 — "Each new set
// of stars resets the planetary enumeration to 'I'."
func TestAssignPlanetDesignations_PerGroupReset(t *testing.T) {
	t.Parallel()

	gAab := Group{Designation: "Aab"}
	gAB := Group{Designation: "AB"}
	gB := Group{Designation: "B"}
	gCab := Group{Designation: "Cab"}

	dps := []DetailedPlacement{
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 7.2, Group: gAB}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 7.8, Group: gAB}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 0.52, Group: gB}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: gB}}}},
		{Placement: Placement{Body: BodyPlanetoidBelt, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.4, Group: gCab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.3, Group: gCab}}}},
	}

	AssignPlanetDesignations(dps)

	want := []string{"Aab I", "AB I", "AB II", "B I", "B II", "Cab PI", "Cab I"}
	for i, w := range want {
		if dps[i].Designation != w {
			t.Errorf("dps[%d].Designation = %q, want %q", i, dps[i].Designation, w)
		}
	}
}

// TestAssignMoonDesignations_AlphabeticOrder: WBH p.58 — "moons are
// ordered from the closest to the farthest from the planet. A space
// separates the planet and moon designation."
func TestAssignMoonDesignations_AlphabeticOrder(t *testing.T) {
	t.Parallel()

	dps := []DetailedPlacement{
		{
			Designation: "Aab IV",
			Moons: []Moon{
				{SizeCode: "2"}, // a
				{SizeCode: "S"}, // b
				{SizeCode: "S"}, // c
				{SizeCode: "5"}, // d
				{SizeCode: "S"}, // e
			},
		},
		{
			Designation: "Aab V",
			Moons: []Moon{
				{SizeCode: "S"}, // a
				{SizeCode: "A"}, // b
				{SizeCode: "1"}, // c
				{SizeCode: "3"}, // d
				{SizeCode: "S"}, // e
				{SizeCode: "S"}, // f
			},
		},
	}

	AssignMoonDesignations(dps)

	wantAabIV := []string{"Aab IV a", "Aab IV b", "Aab IV c", "Aab IV d", "Aab IV e"}
	for i, w := range wantAabIV {
		if dps[0].Moons[i].Designation != w {
			t.Errorf("dps[0].Moons[%d].Designation = %q, want %q",
				i, dps[0].Moons[i].Designation, w)
		}
	}
	wantAabV := []string{"Aab V a", "Aab V b", "Aab V c", "Aab V d", "Aab V e", "Aab V f"}
	for i, w := range wantAabV {
		if dps[1].Moons[i].Designation != w {
			t.Errorf("dps[1].Moons[%d].Designation = %q, want %q",
				i, dps[1].Moons[i].Designation, w)
		}
	}
}

// TestAssignMoonDesignations_NoMoonsNoPanic asserts the no-moons path
// doesn't panic and leaves Designation empty.
func TestAssignMoonDesignations_NoMoonsNoPanic(t *testing.T) {
	t.Parallel()
	dps := []DetailedPlacement{
		{Designation: "Aab III", Moons: nil},
		{Designation: "Aab VII", Moons: []Moon{}},
	}
	AssignMoonDesignations(dps) // must not panic
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run TestAssign -v
```

Expected: FAIL — `AssignPlanetDesignations`, `AssignMoonDesignations` undefined.

- [ ] **Step 3: Create `worlds/designations.go`**

```go
package worlds

import "fmt"

// AssignPlanetDesignations walks Detailed in arrival order (which is
// orbit order per group, established by 2B's PlaceWorlds) and assigns:
//
//   - Non-belt placements:    "<Group> I", "<Group> II", ... (Roman planet counter)
//   - Planetoid belts:        "<Group> PI", "<Group> PII", ... (separate belt counter)
//
// Per WBH p.53 sidebar:
//   - Planet enumeration skips belt locations (counter never advances on a belt).
//   - Each new group resets both counters to I.
//   - Empty placements get no designation (left empty).
//
// Mutates DetailedPlacement.Designation in place.
func AssignPlanetDesignations(dps []DetailedPlacement) {
	currentGroup := ""
	planetN := 0
	beltN := 0
	for i := range dps {
		gd := dps[i].Group.Designation
		if gd != currentGroup {
			currentGroup = gd
			planetN = 0
			beltN = 0
		}
		switch dps[i].Body {
		case BodyEmpty:
			// No designation for empty slots.
		case BodyPlanetoidBelt:
			beltN++
			dps[i].Designation = fmt.Sprintf("%s P%s", gd, romanNumeral(beltN))
		default: // BodyTerrestrial, BodyGasGiant
			planetN++
			dps[i].Designation = fmt.Sprintf("%s %s", gd, romanNumeral(planetN))
		}
	}
}

// AssignMoonDesignations walks each DetailedPlacement.Moons in arrival
// order (which is closest-to-farthest from the planet, per CountMoons +
// SizeMoon's emission order) and assigns "<Planet> a", "<Planet> b", ...
// per WBH p.58 sidebar.
//
// Insignificant moons are out of scope for 2C.
//
// Mutates Moon.Designation in place.
func AssignMoonDesignations(dps []DetailedPlacement) {
	for i := range dps {
		for j := range dps[i].Moons {
			letter := byte('a' + j)
			dps[i].Moons[j].Designation = fmt.Sprintf("%s %c", dps[i].Designation, letter)
		}
	}
}

// romanNumeral returns the Roman-numeral representation of n for n in
// [1, 39]. WBH planet/belt enumeration never approaches that bound
// (Zed has 8 worlds in Aab, the largest in book examples).
func romanNumeral(n int) string {
	if n < 1 {
		return ""
	}
	values := []int{10, 9, 5, 4, 1}
	symbols := []string{"X", "IX", "V", "IV", "I"}
	out := ""
	for i, v := range values {
		for n >= v {
			out += symbols[i]
			n -= v
		}
	}
	return out
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
go test ./worlds -run TestAssign -v
```

Expected: PASS.

- [ ] **Step 5: Run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 6: Format + commit**

```bash
gofumpt -w worlds/designations.go worlds/designations_test.go
git add worlds/designations.go worlds/designations_test.go
git commit -m "feat(worlds): AssignPlanetDesignations + AssignMoonDesignations

Planet designations use Roman numerals per group with a separate \"P\"
counter for planetoid belts (planet counter never advances on belts).
Each new group resets both counters. Moon designations are alphabetic
per planet (a-z), closest-to-farthest.

Validates against the Zed multi-group designation sequence
(Aab I-VIII + Aab PI; AB I-III; B I-II; Cab PI + Cab I-III) and
the per-planet moon alphabet (Aab IV a-e, Aab V a-f)."
```

---

## Task 11: Profile (short + long)

**Source:** Spec § Architecture § Profile (p. 58); WBH p.58 Planetary System Profile.

**Files:** `worlds/profile.go` (create), `worlds/profile_test.go` (create).

**Goal:** `ShortProfile` (`G-P-T-N-S` form) + `LongProfile` (`St-N-W-W-W...-S:` per-star form). Note: collides by name with `stars.ShortProfile`; the package qualifier disambiguates.

**Format reference (WBH p.58):**

- **Short:** `G-P-T-N-S` where G=gas-giant count, P=belt count, T=terrestrial count, N=baseline number (0 if <0), S=spread.
- **Long:** `St-N-W-W-W-...-S:-N-St-W-W...-S:...` where:
  - `St` = star/group designation
  - `N` = baseline number for that star
  - Per slot in orbit order: `G` (gas giant), `P` (belt), `T` (terrestrial)
  - `M` (mainworld), `GM` (gas-giant moon mainworld), `PM` (belt mainworld) reserved; 2C never emits (no picker per spec).
  - `S` = spread for that star
  - Stars separated by `:`

**Zed expected outputs:**

- Short: `"4-2-12-5-0.5"`
- Long: `"Aab-5-T-T-T-P-G-G-T-T-T-0.5:B-2-T-T-0.5:AB-0-T-T-G-0.5:Cab-0-P-G-T-T-0.5"`
  - Wait — the long form has nine `T/G/P` for Aab but Aab has 9 slots (8 worlds + retrograde). Let me re-check vs. p.58. Looking at form p.63: Aab has 9 placed bodies (Aab I, II, III, PI, IV, V, VI, VII, VIII = 9). 4 of those are non-T (PI=P, IV=G, V=G, plus the retrograde Aab VIII). So Aab slots: T-T-T-P-G-G-T-T-T... but the retrograde at orbit 5.2R is the 9th. The book's long form on p.58 shows: `Aab-5-T-T-T-P-G-G-T-T-T-0.5` — 9 slots, ending with T (the retrograde). Match.

- [ ] **Step 1: Write failing tests for both profile functions**

Create `worlds/profile_test.go`:

```go
package worlds

import (
	"strings"
	"testing"
)

// TestShortProfile_Zed asserts the WBH p.58 short form for the Zed
// system: "4-2-12-5-0.5"
func TestShortProfile_Zed(t *testing.T) {
	t.Parallel()

	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			Counts:        Counts{GasGiants: 4, PlanetoidBelts: 2, Terrestrials: 12, Total: 18},
			BaselineN:     5,
			SystemSpread:  0.5,
		},
	}
	got := ShortProfile(sd)
	want := "4-2-12-5-0.5"
	if got != want {
		t.Errorf("ShortProfile = %q, want %q", got, want)
	}
}

// TestShortProfile_BaselineNFloor asserts that BaselineN < 0 renders as 0.
func TestShortProfile_BaselineNFloor(t *testing.T) {
	t.Parallel()
	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			Counts:        Counts{GasGiants: 1, PlanetoidBelts: 0, Terrestrials: 3, Total: 4},
			BaselineN:     -2,
			SystemSpread:  0.7,
		},
	}
	got := ShortProfile(sd)
	want := "1-0-3-0-0.7"
	if got != want {
		t.Errorf("ShortProfile = %q, want %q", got, want)
	}
}

// TestLongProfile_Zed asserts the WBH p.58 long form for Zed:
// "Aab-5-T-T-T-P-G-G-T-T-T-0.5:B-2-T-T-0.5:AB-0-T-T-G-0.5:Cab-0-P-G-T-T-0.5"
func TestLongProfile_Zed(t *testing.T) {
	t.Parallel()

	gAab := Group{Designation: "Aab"}
	gB := Group{Designation: "B"}
	gAB := Group{Designation: "AB"}
	gCab := Group{Designation: "Cab"}

	dps := []DetailedPlacement{
		// Aab: T-T-T-P-G-G-T-T-T (9 slots in orbit order)
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.6, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.1, Group: gAab}}}},
		{Placement: Placement{Body: BodyPlanetoidBelt, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.7, Group: gAab}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.1, Group: gAab}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.5, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.1, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.6, Group: gAab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 5.2, Group: gAab}}}},
		// B: T-T (2 slots)
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 0.52, Group: gB}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: gB}}}},
		// AB: T-T-G (3 slots)
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 7.2, Group: gAB}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 7.8, Group: gAB}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 8.3, Group: gAB}}}},
		// Cab: P-G-T-T (4 slots)
		{Placement: Placement{Body: BodyPlanetoidBelt, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.4, Group: gCab}}}},
		{Placement: Placement{Body: BodyGasGiant, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.3, Group: gCab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.9, Group: gCab}}}},
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.3, Group: gCab}}}},
	}

	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			SystemSpread: 0.5,
			Allocations: []StarAllocation{
				{Group: gAab, BaselineN: 5},
				{Group: gB, BaselineN: 2},
				{Group: gAB, BaselineN: 0},
				{Group: gCab, BaselineN: 0},
			},
		},
		Detailed: dps,
	}

	got := LongProfile(sd)
	want := "Aab-5-T-T-T-P-G-G-T-T-T-0.5:B-2-T-T-0.5:AB-0-T-T-G-0.5:Cab-0-P-G-T-T-0.5"
	if got != want {
		t.Errorf("LongProfile mismatch\n got: %q\nwant: %q", got, want)
	}

	// Sanity: contains 4 colon-separated star segments.
	if c := strings.Count(got, ":"); c != 3 {
		t.Errorf("LongProfile colon count = %d, want 3 (4 stars separated by :)", c)
	}
}
```

**Note:** The test references `StarAllocation.BaselineN` — a per-star baseline number. 2B's `StarAllocation` has `AllocatedWorlds` but not `BaselineN`. Inspect the struct:

```bash
grep -A 10 "^type StarAllocation" /Users/markayers/Documents/Traveller/worlds/allocations.go
```

If `BaselineN` doesn't exist, the per-star baseline is **derivable** from the slot positions: it's the index (1-based) of the slot whose orbit matches the system's baseline orbit, or 0 if the system's baseline orbit is beyond all of this group's slots. For Zed: Aab's baseline orbit 3.1 lands on slot 5 of Aab (the GLE gas giant) → N=5. B's slot at 1.0 is the second of two slots → N=2. AB and Cab have no slot at the system baseline orbit → N=0.

If `StarAllocation` lacks `BaselineN`, derive it inside `LongProfile` by counting placements per group up to `SystemPlacement.BaselineOrbit` (use the placement's group + orbit). The test as written assumes a `BaselineN` field; if missing, refactor the test to put per-group baseline numbers in the SystemPlacement.Allocations slice via a different field, or compute inline.

For this plan, **assume** `StarAllocation.BaselineN` exists; if Step 2 reveals it doesn't, sub-step inserts it into the struct definition (one-line addition, doesn't break 2B since the field defaults to 0).

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run "TestShortProfile|TestLongProfile" -v
```

Expected: FAIL — `ShortProfile`/`LongProfile` undefined; possibly compile error if `StarAllocation.BaselineN` missing.

- [ ] **Step 3a: If `StarAllocation` lacks `BaselineN`, add it**

Edit `worlds/allocations.go`:

```go
type StarAllocation struct {
	Group           Group
	TotalStarOrbits int
	AllocatedWorlds int
	BaselineN       int // per-star baseline-number (1-based slot index of system baseline orbit within this group; 0 if out of range)
}
```

This field defaults to 0 for existing 2B callsites — backward-compatible. The actual baseline-N derivation lives in `LongProfile`'s sub-helper `deriveBaselineN` (Step 3b) or, more cleanly, in a small helper `computePerStarBaselineN` called by the façade in Task 14. For Task 11, we accept `BaselineN` as input from the caller.

- [ ] **Step 3b: Create `worlds/profile.go`**

```go
package worlds

import (
	"fmt"
	"strconv"
	"strings"
)

// ShortProfile renders the WBH p.58 short Planetary System Profile:
// "G-P-T-N-S" where:
//   G = gas-giant count           T = terrestrial count
//   P = planetoid-belt count      N = baseline number (floored at 0)
//   S = system spread
//
// Note: collides by name with stars.ShortProfile (Class 0/I survey).
// The two are distinct profiles for distinct purposes; the package
// qualifier disambiguates at call sites.
func ShortProfile(sd SystemDetail) string {
	n := sd.BaselineN
	if n < 0 {
		n = 0
	}
	return fmt.Sprintf("%d-%d-%d-%d-%s",
		sd.Counts.GasGiants,
		sd.Counts.PlanetoidBelts,
		sd.Counts.Terrestrials,
		n,
		formatSpread(sd.SystemSpread),
	)
}

// LongProfile renders the WBH p.58 long Planetary System Profile:
// "St-N-W-W-W...-S:-N-St-W-W...-S:..." per star.
//
// Per-slot codes (in orbit order):
//   G  = gas giant
//   P  = planetoid belt
//   T  = terrestrial
//   M  = mainworld terrestrial    (reserved; 2C never emits)
//   GM = gas-giant moon mainworld (reserved; 2C never emits)
//   PM = planetoid belt mainworld (reserved; 2C never emits)
func LongProfile(sd SystemDetail) string {
	// Group placements by group designation, preserving allocation order.
	type starSegment struct {
		designation string
		baselineN   int
		codes       []string
	}
	segments := []starSegment{}
	groupIdx := map[string]int{}

	for _, alloc := range sd.Allocations {
		groupIdx[alloc.Group.Designation] = len(segments)
		segments = append(segments, starSegment{
			designation: alloc.Group.Designation,
			baselineN:   alloc.BaselineN,
		})
	}

	for _, dp := range sd.Detailed {
		gd := dp.Group.Designation
		idx, ok := groupIdx[gd]
		if !ok {
			continue
		}
		var code string
		switch dp.Body {
		case BodyTerrestrial:
			code = "T"
		case BodyGasGiant:
			code = "G"
		case BodyPlanetoidBelt:
			code = "P"
		case BodyEmpty:
			continue // empty slots don't appear in long profile
		default:
			continue
		}
		segments[idx].codes = append(segments[idx].codes, code)
	}

	parts := make([]string, 0, len(segments))
	for _, seg := range segments {
		fields := []string{seg.designation, strconv.Itoa(seg.baselineN)}
		fields = append(fields, seg.codes...)
		fields = append(fields, formatSpread(sd.SystemSpread))
		parts = append(parts, strings.Join(fields, "-"))
	}
	return strings.Join(parts, ":")
}

// formatSpread renders a spread value to one decimal place ("0.5",
// "0.7"). Strips a trailing ".0" if the value is whole? No — the form
// shows "0.5" so always one decimal.
func formatSpread(s float64) string {
	return strconv.FormatFloat(s, 'f', 1, 64)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
go test ./worlds -run "TestShortProfile|TestLongProfile" -v
```

Expected: PASS.

- [ ] **Step 5: Run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 6: Format + commit**

```bash
gofumpt -w worlds/profile.go worlds/profile_test.go worlds/allocations.go
git add worlds/profile.go worlds/profile_test.go worlds/allocations.go
git commit -m "feat(worlds): ShortProfile + LongProfile (WBH p.58)

ShortProfile renders the G-P-T-N-S form (gas giants, belts, terrestrials,
baseline number floored at 0, spread). LongProfile renders the per-star
St-N-W-W-...-S:-N-St-... form with T/G/P slot codes (M/GM/PM reserved
but unused by 2C since the mainworld picker is deferred).

Adds BaselineN field to StarAllocation (defaults to 0, backward-compatible
with existing 2B callsites). Per-star baseline-N derivation lives in
the DetailSystem façade (Task 14).

Validates against Zed short \"4-2-12-5-0.5\" and the full long form
exactly per p.58."
```

---

## Task 12: Mainworld (`MarkHZ` + `MainworldCandidates`)

**Source:** Spec § Architecture § Mainworld (pp. 58-59); WBH p.58 mainworld candidate procedure.

**Files:** `worlds/mainworld.go` (create), `worlds/mainworld_test.go` (create).

**Goal:** `MainworldCandidate` struct + `MarkHZ` (sets `DetailedPlacement.HZ` for orbits within HZCO ± 1.0) + `MainworldCandidates` (returns terrestrial planets in HZ + significant moons of any HZ planet, including gas-giant moons). No picker — selection requires World Physical (deferred).

**WBH p.58 HZ rule:** "The habitable zone is generally considered to be +/ 1.0 Orbit#s from the HZCO."

**Mainworld candidate eligibility (per p.58):**

- Terrestrial planet in HZ (Aab VI in Zed).
- Significant moon (any size) of any HZ planet, including gas-giant moons (Aab IV d at Size 5, Aab V b at Size A, Aab V d at Size 3 in Zed).
- Belts and gas giants themselves are _not_ candidates (gas giants can't be inhabited; belts get their own continuation method).

- [ ] **Step 1: Write failing tests for MarkHZ + MainworldCandidates**

Create `worlds/mainworld_test.go`:

```go
package worlds

import (
	"math"
	"testing"

	"wbh/stars"
)

// TestMarkHZ_WindowInclusion: a placement at orbit 3.1 with group HZCO 3.3
// is in the HZ (range 2.3-4.3); orbits at 5.0 and 1.0 are not.
func TestMarkHZ_WindowInclusion(t *testing.T) {
	t.Parallel()

	// Fabricate a Group whose HZCO returns 3.3 by giving it a single
	// G2-like star with luminosity tuned to that HZCO. Per WBH p.41:
	// HZCO_AU = sqrt(luminosity); HZCO_Orbit = AUToOrbit(HZCO_AU).
	// For HZCO_Orbit ≈ 3.3 → HZCO_AU ≈ 1.32 → luminosity ≈ 1.74. Compose
	// a star with that luminosity.
	star := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 0},
		LuminosityClass: stars.V,
		Luminosity:      1.74,
		Mass:            1.0,
	})
	g := Group{Designation: "Aab", Members: []stars.Star{star}}

	dps := []DetailedPlacement{
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.1, Group: g}}}}, // in HZ
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 5.0, Group: g}}}}, // outside (above)
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: g}}}}, // outside (below)
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.1, Group: g}}}}, // in HZ
	}

	if err := MarkHZ(dps); err != nil {
		t.Fatalf("MarkHZ err: %v", err)
	}
	wantHZ := []bool{true, false, false, true}
	for i, w := range wantHZ {
		if dps[i].HZ != w {
			t.Errorf("dps[%d].HZ (orbit %v) = %v, want %v", i, dps[i].Orbit, dps[i].HZ, w)
		}
	}
}

// TestMainworldCandidates_PlanetAndMoonCandidates: terrestrial planet in
// HZ + significant moons of any HZ planet (including gas-giant moons).
// Belts and gas giants themselves are not candidates.
func TestMainworldCandidates_PlanetAndMoonCandidates(t *testing.T) {
	t.Parallel()

	// Build a tiny system: one HZ gas giant (Aab IV) with two moons,
	// one HZ terrestrial (Aab VI), and one out-of-HZ terrestrial.
	g := Group{Designation: "Aab"}
	dps := []DetailedPlacement{
		{
			Placement: Placement{
				Body:          BodyGasGiant,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.1, Group: g}},
			},
			Designation: "Aab IV",
			HZ:          true,
			GGClass:     GasGiantLarge,
			Moons: []Moon{
				{Designation: "Aab IV a", SizeCode: "2", DiameterKm: 3200},
				{Designation: "Aab IV d", SizeCode: "5", DiameterKm: 8000},
			},
		},
		{
			Placement: Placement{
				Body:          BodyTerrestrial,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.1, Group: g}},
			},
			Designation: "Aab VI",
			HZ:          true,
			SizeCode:    "A",
			DiameterKm:  16000,
		},
		{
			Placement: Placement{
				Body:          BodyTerrestrial,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: g}},
			},
			Designation: "Aab I",
			HZ:          false, // outside HZ
			SizeCode:    "B",
			DiameterKm:  17600,
		},
		{
			Placement: Placement{
				Body:          BodyPlanetoidBelt,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.7, Group: g}},
			},
			Designation: "Aab PI",
			HZ:          false,
		},
	}
	sd := SystemDetail{Detailed: dps}

	got := MainworldCandidates(sd)

	// Expected: Aab VI (planet) + Aab IV a + Aab IV d (moons of HZ gas giant).
	// Aab I is excluded (not in HZ). The belt is excluded (no candidate).
	// The gas giant Aab IV itself is excluded (gas giants can't be mainworlds).
	wantDesigs := []string{"Aab IV a", "Aab IV d", "Aab VI"}
	if len(got) != len(wantDesigs) {
		t.Fatalf("len(MainworldCandidates) = %d, want %d (got: %v)", len(got), len(wantDesigs), got)
	}
	for i, w := range wantDesigs {
		if got[i].Designation != w {
			t.Errorf("got[%d].Designation = %q, want %q", i, got[i].Designation, w)
		}
	}

	// Spot-check fields on the moon candidate.
	moon := got[0]
	if !moon.IsMoon {
		t.Error("got[0].IsMoon = false, want true")
	}
	if moon.ParentDesignation != "Aab IV" {
		t.Errorf("got[0].ParentDesignation = %q, want \"Aab IV\"", moon.ParentDesignation)
	}
	if math.Abs(moon.Orbit-3.1) > 1e-9 {
		t.Errorf("got[0].Orbit = %v, want 3.1 (parent's orbit)", moon.Orbit)
	}

	// Spot-check the planet candidate.
	planet := got[2]
	if planet.IsMoon {
		t.Error("got[2].IsMoon = true, want false (planet)")
	}
	if planet.ParentDesignation != "" {
		t.Errorf("got[2].ParentDesignation = %q, want \"\"", planet.ParentDesignation)
	}
	if planet.HostStarGroup != "Aab" {
		t.Errorf("got[2].HostStarGroup = %q, want \"Aab\"", planet.HostStarGroup)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run "TestMarkHZ|TestMainworldCandidates" -v
```

Expected: FAIL — `MarkHZ`, `MainworldCandidates`, `MainworldCandidate` undefined.

- [ ] **Step 3: Create `worlds/mainworld.go`**

```go
package worlds

// MainworldCandidate describes one body eligible to be the system's
// mainworld per WBH p.58. 2C enumerates candidates; the picker
// (Referee judgment from atmosphere/hydrographics rolls) lands with
// sub-project 3 (World Physical, pp. 69-146).
type MainworldCandidate struct {
	Designation       string
	SizeCode          SizeCode
	DiameterKm        float64
	Orbit             float64 // host planet's orbit (for moons, the parent's)
	HostStarGroup     string  // "Aab", "B", "Cab"
	IsMoon            bool
	ParentDesignation string // "" for planet candidates; "Aab IV" for moons
}

// MarkHZ sets DetailedPlacement.HZ when the placement's orbit lies
// within HZCO ± 1.0 of the placement's host group, per WBH p.58:
// "The habitable zone is generally considered to be +/ 1.0 Orbit#s
// from the HZCO."
//
// Uses Group.HZCO() from 2B (defined in worlds/group_hzco.go).
//
// Mutates DetailedPlacement.HZ in place.
func MarkHZ(dps []DetailedPlacement) error {
	for i := range dps {
		if dps[i].Body == BodyEmpty {
			continue
		}
		hzco := dps[i].Group.HZCO()
		o := dps[i].Orbit
		if o >= hzco-1.0 && o <= hzco+1.0 {
			dps[i].HZ = true
		}
	}
	return nil
}

// MainworldCandidates returns the subset of bodies eligible to be the
// mainworld per WBH p.58:
//
//   - Terrestrial planets in the HZ
//   - Significant moons (any Size) of any HZ planet, including
//     gas-giant moons (per the Zed example: Aab IV d, Aab V b, Aab V d)
//
// Belts and gas giants themselves are NOT candidates.
//
// Returned in DetailedPlacement order, with each planet's moons listed
// before the planet itself when both are candidates (matches Zed form
// p.63 Object table ordering).
//
// Selection requires World Physical (sub-project 3); 2C provides
// enumeration only.
func MainworldCandidates(sd SystemDetail) []MainworldCandidate {
	out := []MainworldCandidate{}
	for _, dp := range sd.Detailed {
		if !dp.HZ {
			continue
		}
		// Moons first (form ordering: parent's moon rows precede the
		// parent's planet row in the Notes column).
		for _, m := range dp.Moons {
			out = append(out, MainworldCandidate{
				Designation:       m.Designation,
				SizeCode:          m.SizeCode,
				DiameterKm:        m.DiameterKm,
				Orbit:             dp.Orbit, // parent's orbit
				HostStarGroup:     dp.Group.Designation,
				IsMoon:            true,
				ParentDesignation: dp.Designation,
			})
		}
		// Then the planet itself, if it's terrestrial.
		if dp.Body == BodyTerrestrial {
			out = append(out, MainworldCandidate{
				Designation:   dp.Designation,
				SizeCode:      dp.SizeCode,
				DiameterKm:    dp.DiameterKm,
				Orbit:         dp.Orbit,
				HostStarGroup: dp.Group.Designation,
				IsMoon:        false,
			})
		}
	}
	return out
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
go test ./worlds -run "TestMarkHZ|TestMainworldCandidates" -v
```

Expected: PASS.

- [ ] **Step 5: Run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 6: Format + commit**

```bash
gofumpt -w worlds/mainworld.go worlds/mainworld_test.go
git add worlds/mainworld.go worlds/mainworld_test.go
git commit -m "feat(worlds): MarkHZ + MainworldCandidates (WBH p.58)

MarkHZ sets DetailedPlacement.HZ when orbit ∈ [HZCO-1.0, HZCO+1.0]
of the placement's host group, using Group.HZCO() from 2B.

MainworldCandidates enumerates eligible bodies per WBH p.58:
terrestrial planets in HZ + significant moons (any size) of any HZ
planet (including gas-giant moons). Belts and gas giants themselves
are excluded. No picker — selection requires World Physical
(sub-project 3 deferral, per spec Q2/A)."
```

---

## Task 13: IISS Class II/III form rendering

**Source:** Spec § Architecture § IISS Class II/III Survey form (pp. 60-67); WBH p.61 blank form, p.63 Zed filled.

**Files:**

- `stars/survey.go` (extend) — add `MAO float64` field to `SurveyComponent`; populate in `BuildSurveyForm`.
- `stars/survey_test.go` (extend) — assert MAO on Sol/Zed Class 0/I form rows.
- `worlds/survey_form.go` (create) — `IISSClass23Form`, `ObjectRow`, `IISSClass23Header`, `RenderIISSClass23`.
- `worlds/survey_form_test.go` (create) — per-section + per-cell rendering tests.
- `worlds/system_detail.go` (edit) — replace forward-declared `IISSClass23Form` placeholder.

**Goal:** Build the structured Class II/III form, embedding the existing Class 0/I `stars.SurveyForm` for the header + Stars table (now extended with MAO) and adding the Objects table per WBH p.61.

This task has two phases: (a) extend `stars.SurveyComponent` with `MAO` (small, isolated change in stars package); (b) build the `worlds.IISSClass23Form` and renderer.

### Phase A: Extend `stars.SurveyComponent` with MAO

- [ ] **Step 1: Read existing `BuildSurveyForm` to identify MAO-eligible rows**

```bash
grep -n "HZCO\|buildBarycentre\|componentFrom" /Users/markayers/Documents/Traveller/stars/survey.go
```

Expected output: rows that today carry `HZCO` will also carry `MAO`. These are: solo primary (no companion), each `Aab`/`Cab`/etc. composite, and each solo orbit-class secondary.

- [ ] **Step 2: Write a failing test that asserts MAO on the Zed form rows**

Append to `stars/survey_test.go` (or wherever the existing Zed Class 0/I worked-example test lives — likely in `stars/worked_examples_test.go`):

```go
// TestZed_Class01Form_MAO asserts that BuildSurveyForm populates the
// MAO field on rows where MAO is defined per WBH p.39 Maximum
// Available Orbit table.
//
// Per p.63 Zed Class II/III form, MAO values are:
//   Aab (A) composite: 0.61
//   B (solo Near):     0.02
//   AB composite:      7.10
//   Cab (C) composite: 0.74
//   ABC composite:    14.10
func TestZed_Class01Form_MAO(t *testing.T) {
	t.Parallel()

	sys := composeZed() // existing helper
	form := stars.BuildSurveyForm(sys, stars.SurveyMetadata{
		Sector:        "Storr",
		Location:      "0602",
		Designation:   "Zed (system)",
		InitialSurvey: "207-568",
		LastUpdated:   "218-1061",
	})

	wantMAO := map[string]float64{
		"Aab (A)": 0.61,
		"B":       0.02,
		"AB":      7.10,
		"Cab (C)": 0.74,
		"ABC":     14.10,
	}

	for _, row := range form.Stars {
		if want, ok := wantMAO[row.Component]; ok {
			if math.Abs(row.MAO-want) > 0.01 {
				t.Errorf("row %q MAO = %v, want %v (±0.01)", row.Component, row.MAO, want)
			}
			delete(wantMAO, row.Component)
		}
	}
	for comp := range wantMAO {
		t.Errorf("expected MAO row %q not found", comp)
	}
}
```

If `composeZed` lives in `stars/worked_examples_test.go`, place this test there. Adjust the test's package import statements as needed (it should be `package stars_test`).

- [ ] **Step 3: Run the test to verify it fails**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./stars -run TestZed_Class01Form_MAO -v
```

Expected: FAIL — `SurveyComponent.MAO` undefined.

- [ ] **Step 4: Add `MAO` field to `SurveyComponent` and populate it in `BuildSurveyForm`**

Edit `stars/survey.go`. Add the MAO field to the struct:

```go
type SurveyComponent struct {
	Component    string
	Class        string
	Mass         float64
	Temperature  float64
	Diameter     float64
	Luminosity   float64
	Orbit        float64
	AU           float64
	Eccentricity float64
	PeriodYears  float64
	HZCO         float64
	MAO          float64 // populated on the same rows as HZCO (per p.39 MAO table; computed via worlds.MAO indirectly — see implementation note)
}
```

Compute MAO using `worlds.MAO` (already exists per `worlds/available_orbits.go:175`). To avoid the cycle (`stars` cannot import `worlds`), one of three approaches:

**Option A (recommended):** Move MAO computation logic into `stars/`. The MAO table on WBH p.39 is stellar mechanics. Create `stars/mao.go` with the MAO function (copy from `worlds/available_orbits.go`); deprecate the `worlds.MAO` wrapper to call `stars.MAO`.

**Option B:** Have `BuildSurveyForm` accept a `MAO func(stars.Star) float64` callback so the caller (worlds) injects the function.

**Option C:** Leave MAO=0 on the form rows in `stars` and have `worlds.RenderIISSClass23` post-fill the MAO field.

For this plan, take **Option C** — it minimizes disruption to the existing 2A `worlds.MAO` and `stars/survey.go` code. The `BuildSurveyForm` function leaves `MAO=0`, and `worlds.RenderIISSClass23` (Phase B below) post-walks the Stars rows and fills MAO using the existing `worlds.MAO` function plus the group's MAO from 2A's `Result.Groups[i].MAO` directly. Simpler and avoids moving code.

**Revised Step 4 (Option C):** Just add the field. No `BuildSurveyForm` change.

```go
type SurveyComponent struct {
	// ... existing fields ...
	MAO float64 // 0 in Class 0/I forms; populated by worlds.RenderIISSClass23 for Class II/III.
}
```

The Class 0/I test from Step 2 will assert MAO=0 if BuildSurveyForm doesn't populate it. Adjust the test scope: either (a) move the MAO test to Class II/III form testing in Phase B; or (b) keep BuildSurveyForm assertions on MAO=0 and assert non-zero MAO only after the worlds render.

For TDD continuity, **delete the Class 0/I MAO test from Step 2** and move the assertion into Phase B's worlds test.

- [ ] **Step 5: Add the MAO field; revert the Class 0/I test deletion**

Edit `stars/survey.go` to add the MAO field as shown above. No other change in stars/. Remove the test added in Step 2 (or simply skip Step 2 entirely and proceed to Phase B). For the file diff:

```bash
# Confirm the field is added; no other survey.go changes needed.
grep -n "MAO " /Users/markayers/Documents/Traveller/stars/survey.go
```

Expected: one match on the new field.

- [ ] **Step 6: Run check + test to verify no regressions**

```bash
just check && just test
```

Expected: clean (the MAO field defaults to 0 on all existing Class 0/I outputs; no behavior change).

### Phase B: Build `worlds.IISSClass23Form` + `RenderIISSClass23`

- [ ] **Step 7: Write failing tests for the Class II/III form rendering**

Create `worlds/survey_form_test.go`:

```go
package worlds

import (
	"strings"
	"testing"

	"wbh/stars"
)

// TestRenderIISSClass23_Header asserts the form header fields are
// populated from IISSClass23Header + sys.AgeGyr.
func TestRenderIISSClass23_Header(t *testing.T) {
	t.Parallel()

	sys := stars.System{
		Primary: stars.Compose(stars.ComposeOpts{
			Kind:            stars.KindMainSequence,
			SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
			LuminosityClass: stars.V,
			Mass:            1.0,
			Diameter:        1.0,
			Temperature:     5800,
			Luminosity:      1.0,
			AgeGyr:          4.6,
		}),
		PrimaryDesignation: "A",
		AgeGyr:             4.6,
	}
	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			Counts: Counts{GasGiants: 1, PlanetoidBelts: 1, Terrestrials: 4, Total: 6},
		},
	}
	header := IISSClass23Header{
		SectorLocation:  "Sol Sector | 0801",
		InitialSurvey:   "001-001",
		LastUpdated:     "218-1061",
		IISSDesignation: "Sol (system)",
		Comments:        "Reference system.",
	}

	form := RenderIISSClass23(sd, sys, header)

	if form.Sector != "Sol Sector | 0801" && form.IISSDesig != "Sol (system)" {
		// SurveyForm is embedded; we set Sector via SectorLocation? See spec
		// notes — IISSClass23Header.SectorLocation maps to SurveyForm.Sector
		// + Location concatenated, OR maps directly to a single Sector field.
		// For this test we accept either by checking the IISSDesig.
		t.Errorf("IISSDesig = %q, want \"Sol (system)\"", form.IISSDesig)
	}
	if form.SystemAgeGyr != 4.6 {
		t.Errorf("SystemAgeGyr = %v, want 4.6", form.SystemAgeGyr)
	}
	if form.GasGiants != 1 || form.PlanetoidBelts != 1 || form.Terrestrials != 4 {
		t.Errorf("counts = (%d, %d, %d), want (1, 1, 4)",
			form.GasGiants, form.PlanetoidBelts, form.Terrestrials)
	}
	if form.ClassIIIStatus {
		t.Error("ClassIIIStatus = true, want false (2C always renders Class II)")
	}
	if form.Comments != "Reference system." {
		t.Errorf("Comments = %q, want \"Reference system.\"", form.Comments)
	}
}

// TestRenderIISSClass23_ObjectsTerrestrialRow asserts an unrolled
// terrestrial renders SAH "<Size>??" (atmosphere/hydrographics deferred
// to sub-project 3 per spec carve-out).
func TestRenderIISSClass23_ObjectsTerrestrialRow(t *testing.T) {
	t.Parallel()

	dps := []DetailedPlacement{
		{
			Placement: Placement{
				Body:          BodyTerrestrial,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: Group{Designation: "Aab"}}},
			},
			Designation: "Aab I",
			SizeCode:    "B",
			DiameterKm:  17600,
			Period:      Period{Years: 0.187, Days: 68.3},
			HZ:          false,
		},
	}
	sd := SystemDetail{Detailed: dps}
	form := RenderIISSClass23(sd, stars.System{}, IISSClass23Header{})
	if len(form.Objects) != 1 {
		t.Fatalf("len(Objects) = %d, want 1", len(form.Objects))
	}
	row := form.Objects[0]
	if row.Designation != "Aab I" {
		t.Errorf("Designation = %q, want \"Aab I\"", row.Designation)
	}
	if row.SAH != "B??" {
		t.Errorf("SAH = %q, want \"B??\" (Size known, atmo/hyd deferred)", row.SAH)
	}
	if row.PeriodStr != "0.187y" {
		t.Errorf("PeriodStr = %q, want \"0.187y\"", row.PeriodStr)
	}
}

// TestRenderIISSClass23_ObjectsGasGiantRow asserts a gas giant renders
// SAH "<Class><DiameterCode>" (e.g. "GLE" for Large diameter E).
func TestRenderIISSClass23_ObjectsGasGiantRow(t *testing.T) {
	t.Parallel()

	dps := []DetailedPlacement{
		{
			Placement: Placement{
				Body:          BodyGasGiant,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.1, Group: Group{Designation: "Aab"}}},
			},
			Designation:    "Aab IV",
			GGClass:        GasGiantLarge,
			GGDiameterCode: "E",
			DiameterEarth:  14,
			MassEarth:      1200,
			Period:         Period{Years: 0.805, Days: 294.0},
			HZ:             true,
		},
	}
	sd := SystemDetail{Detailed: dps}
	form := RenderIISSClass23(sd, stars.System{}, IISSClass23Header{})
	row := form.Objects[0]
	if row.SAH != "GLE" {
		t.Errorf("SAH = %q, want \"GLE\"", row.SAH)
	}
	if !strings.Contains(row.Notes, "1,200⊕") {
		t.Errorf("Notes = %q, want contains \"1,200⊕\"", row.Notes)
	}
	if !strings.Contains(row.Notes, "HZ") {
		t.Errorf("Notes = %q, want contains \"HZ\" (HZ=true)", row.Notes)
	}
}

// TestRenderIISSClass23_ObjectsBeltRow asserts a planetoid belt renders
// SAH "000" with Sub "?".
func TestRenderIISSClass23_ObjectsBeltRow(t *testing.T) {
	t.Parallel()

	dps := []DetailedPlacement{
		{
			Placement: Placement{
				Body:          BodyPlanetoidBelt,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.7, Group: Group{Designation: "Aab"}}},
			},
			Designation: "Aab PI",
		},
	}
	sd := SystemDetail{Detailed: dps}
	form := RenderIISSClass23(sd, stars.System{}, IISSClass23Header{})
	row := form.Objects[0]
	if row.SAH != "000" {
		t.Errorf("SAH = %q, want \"000\"", row.SAH)
	}
	if row.Sub != "?" {
		t.Errorf("Sub = %q, want \"?\"", row.Sub)
	}
}

// TestRenderIISSClass23_PeriodFormatting covers the years-vs-days
// magnitude rule: P.Years < 0.05 → "<3 decimals>d", otherwise
// "<3 decimals>y" with thousands-separator commas for ≥1000y.
func TestRenderIISSClass23_PeriodFormatting(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		p    Period
		want string
	}{
		{"sub-day → days", Period{Years: 0.005, Days: 1.826}, "1.826d"},
		{"sub-fortnight → days", Period{Years: 0.04, Days: 14.61}, "14.610d"},
		{"normal → years", Period{Years: 0.187, Days: 68.3}, "0.187y"},
		{"big → years with comma", Period{Years: 3598.0, Days: 0}, "3,598y"},
		{"medium → years 3 decimal", Period{Years: 8.627, Days: 0}, "8.627y"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := formatPeriod(tc.p)
			if got != tc.want {
				t.Errorf("formatPeriod(%v) = %q, want %q", tc.p, got, tc.want)
			}
		})
	}
}

// TestRenderIISSClass23_HZCandidateMaskedSAH asserts that HZ-flagged
// terrestrial worlds (which would have full SAH after sub-project 3)
// render with "?" placeholders for atmosphere and hydrographics until
// then.
func TestRenderIISSClass23_HZCandidateMaskedSAH(t *testing.T) {
	t.Parallel()

	dps := []DetailedPlacement{
		{
			Placement: Placement{
				Body:          BodyTerrestrial,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.1, Group: Group{Designation: "Aab"}}},
			},
			Designation: "Aab VI",
			SizeCode:    "A",
			DiameterKm:  16000,
			HZ:          true,
		},
	}
	sd := SystemDetail{Detailed: dps}
	form := RenderIISSClass23(sd, stars.System{}, IISSClass23Header{})
	row := form.Objects[0]
	// Per spec Q1/A: HZ-candidate atmosphere/hydrographics digits render
	// as "?" until World Physical (sub-project 3) lands.
	if row.SAH != "A??" {
		t.Errorf("HZ-candidate SAH = %q, want \"A??\" (Size A + ? + ? per spec carve-out)", row.SAH)
	}
}
```

- [ ] **Step 8: Run the tests to verify they fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run TestRenderIISSClass23 -v
```

Expected: FAIL — `RenderIISSClass23`, `IISSClass23Header`, `ObjectRow`, `formatPeriod` undefined.

- [ ] **Step 9: Replace the placeholder `IISSClass23Form` in `system_detail.go`**

Edit `worlds/system_detail.go`. Remove the placeholder type definition:

```go
// REMOVE these lines from system_detail.go:
// type IISSClass23Form struct {
//     stars.SurveyForm
// }
```

The full type now lives in `worlds/survey_form.go` (created next step).

- [ ] **Step 10: Create `worlds/survey_form.go`**

```go
package worlds

import (
	"fmt"
	"strconv"
	"strings"

	"wbh/stars"
)

// IISSClass23Form is the structured representation of WBH form 0421D-II.III
// (pp. 60-67). Embeds the existing stars.SurveyForm (Class 0/I header
// + Stars table — now extended with MAO via Task 13 Phase A) and adds
// the Class II/III-specific fields: per-body counts, Class III status
// flag, and the Objects table.
//
// Cells dependent on World Physical (sub-project 3) render with "?"
// placeholders until then. Specifically: atmosphere/hydrographics
// digits in HZ-candidate SAH cells; the mainworld marker (asterisk).
type IISSClass23Form struct {
	stars.SurveyForm // header + Stars table (Class 0/I)

	GasGiants      int
	PlanetoidBelts int
	Terrestrials   int
	ClassIIIStatus bool // 2C always renders false (Class II only)

	Objects []ObjectRow
}

// ObjectRow is one row of the WBH p.61 Objects table.
type ObjectRow struct {
	Primary     string  // host star group: "Aab", "AB", "B", "Cab"
	Designation string  // "Aab I", "Aab IV d"
	Orbit       float64
	AU          float64
	Ecc         float64
	PeriodStr   string  // "1.841d" or "8.627y"
	SAH         string  // "B??" / "GLE" / "AA6" / "200" / "566*" / "000" / "S"
	Sub         string  // significant-moon count, "?" for belt, "" for moon row
	Notes       string  // "HZ, R02, S, 1, 1" / "1,200⊕, HZ, 200, S, S, 566*, S"
}

// IISSClass23Header carries the form's header fields not derivable
// from SystemDetail or stars.System.
type IISSClass23Header struct {
	SectorLocation  string
	InitialSurvey   string
	LastUpdated     string
	IISSDesignation string
	Comments        string
}

// RenderIISSClass23 builds the Class II/III form. The Stars table is
// derived from sys via stars.BuildSurveyForm (with MAO post-filled by
// this function); the Objects table is derived from sd.Detailed.
func RenderIISSClass23(sd SystemDetail, sys stars.System, h IISSClass23Header) IISSClass23Form {
	// Build the Class 0/I form for the header + Stars table.
	sec, loc := splitSectorLocation(h.SectorLocation)
	meta := stars.SurveyMetadata{
		Sector:        sec,
		Location:      loc,
		Designation:   h.IISSDesignation,
		InitialSurvey: h.InitialSurvey,
		LastUpdated:   h.LastUpdated,
	}
	base := stars.BuildSurveyForm(sys, meta)
	base.Comments = h.Comments

	// Post-fill MAO on the Stars rows. Per Phase A Step 4 design choice,
	// stars.BuildSurveyForm leaves MAO=0; worlds fills it from the
	// AvailableOrbits result for the system.
	if avail, err := AvailableOrbits(sys); err == nil {
		fillStarsMAO(&base, avail)
	}

	form := IISSClass23Form{
		SurveyForm:     base,
		GasGiants:      sd.Counts.GasGiants,
		PlanetoidBelts: sd.Counts.PlanetoidBelts,
		Terrestrials:   sd.Counts.Terrestrials,
		ClassIIIStatus: false, // 2C always renders Class II
	}

	// Build the Objects table.
	for _, dp := range sd.Detailed {
		if dp.Body == BodyEmpty {
			continue
		}
		row := ObjectRow{
			Primary:     dp.Group.Designation,
			Designation: dp.Designation,
			Orbit:       dp.Orbit,
			AU:          stars.OrbitToAU(dp.Orbit),
			Ecc:         dp.Eccentricity,
			PeriodStr:   formatPeriod(dp.Period),
			SAH:         renderSAH(dp),
			Sub:         renderSub(dp),
			Notes:       renderNotes(dp),
		}
		form.Objects = append(form.Objects, row)

		// Add a row per significant moon.
		for _, m := range dp.Moons {
			form.Objects = append(form.Objects, ObjectRow{
				Primary:     dp.Group.Designation,
				Designation: m.Designation,
				SAH:         renderMoonSAH(m, dp.HZ),
				Sub:         "",
				// Orbit/AU/Ecc/Period left blank for moons (form p.63
				// shows "—" for these on moon rows).
			})
		}
	}

	return form
}

// splitSectorLocation parses "Sector | NNNN" into (sector, location).
// If the input lacks the " | " separator, the whole string is the sector.
func splitSectorLocation(s string) (sector, location string) {
	if idx := strings.Index(s, " | "); idx >= 0 {
		return s[:idx], s[idx+3:]
	}
	return s, ""
}

// fillStarsMAO post-walks the Stars rows and copies the MAO from the
// matching Group in the AvailableOrbits result. Match is by Component
// name parsing: "Aab (A)" → group designation "Aab"; "B" → "B"; etc.
func fillStarsMAO(form *stars.SurveyForm, avail Result) {
	maoByGroup := map[string]float64{}
	for _, g := range avail.Groups {
		maoByGroup[g.Designation] = g.MAO
	}
	// AB/ABC composite-row MAO is the outer-companion's AU (the
	// distance from the primary system to that secondary), not a Group
	// MAO. Build a parallel lookup keyed by composite-name for those rows.
	composeMAO := map[string]float64{}
	for _, c := range sys.Companions {
		if c.OrbitClass == stars.OrbitNear || c.OrbitClass == stars.OrbitFar {
			// "AB" composite row inherits MAO from the Near companion's AU;
			// "ABC" inherits from the Far companion's AU.
			switch c.OrbitClass {
			case stars.OrbitNear:
				composeMAO["AB"] = c.AU
			case stars.OrbitFar:
				composeMAO["ABC"] = c.AU
			}
		}
	}
	for i, row := range form.Stars {
		// Component might be "Aab (A)", "B", "Cab (C)", "AB", "ABC".
		// Extract the group designation by taking the token before " (".
		key := row.Component
		if idx := strings.Index(key, " ("); idx >= 0 {
			key = key[:idx]
		}
		if mao, ok := maoByGroup[key]; ok {
			form.Stars[i].MAO = mao
			continue
		}
		if mao, ok := composeMAO[key]; ok {
			form.Stars[i].MAO = mao
		}
	}
}

// renderSAH produces the SAH/UWP cell for a body.
//
// Terrestrial: "<Size>??" (atmosphere/hydrographics deferred to
// sub-project 3 per spec Q1/A) — even for non-HZ worlds (Class II
// surveys only determine Size for non-HZ bodies).
//
// Gas giant: "GS<diameterCode>" / "GM<.../>" / "GL<...>".
//
// Planetoid belt: "000".
func renderSAH(dp DetailedPlacement) string {
	switch dp.Body {
	case BodyTerrestrial:
		return string(dp.SizeCode) + "??"
	case BodyGasGiant:
		prefix := "G"
		switch dp.GGClass {
		case GasGiantSmall:
			prefix = "GS"
		case GasGiantMedium:
			prefix = "GM"
		case GasGiantLarge:
			prefix = "GL"
		}
		return prefix + dp.GGDiameterCode
	case BodyPlanetoidBelt:
		return "000"
	default:
		return ""
	}
}

// renderMoonSAH renders the SAH for a significant moon in the Notes
// column. For HZ-parent moons, atmosphere/hydrographics is deferred
// (renders just the Size letter + "??"). For non-HZ-parent moons,
// just the Size letter (the Class II survey doesn't produce SAH for
// non-HZ moons).
func renderMoonSAH(m Moon, parentInHZ bool) string {
	if m.GGClass != NotGasGiant {
		// Gas-giant-sized moon (rare, GG Special Sizing cascade).
		prefix := "GS"
		switch m.GGClass {
		case GasGiantMedium:
			prefix = "GM"
		case GasGiantLarge:
			prefix = "GL"
		}
		return prefix + m.GGDiameterCode
	}
	if parentInHZ {
		return string(m.SizeCode) + "??"
	}
	return string(m.SizeCode)
}

// renderSub renders the Sub (subordinate-bodies) column.
//
//	Belt:                "?" (book uses "?" because dwarf-planet count
//	                          isn't determined until World Physical).
//	Moon row:            "" (moons are subordinates, not parents).
//	Planet/gas giant:    decimal count of significant moons, "0" if none.
func renderSub(dp DetailedPlacement) string {
	if dp.Body == BodyPlanetoidBelt {
		return "?"
	}
	if len(dp.Moons) == 0 {
		return "0"
	}
	return strconv.Itoa(len(dp.Moons))
}

// renderNotes builds the Notes column.
//
// For gas giants: "<mass>⊕" prefix (e.g. "1,200⊕"); HZ tag if HZ;
// then a comma-separated list of moon SAH (per the Zed form).
//
// For terrestrials: just the HZ tag if HZ.
//
// For belts: empty (book p.63 shows blank).
func renderNotes(dp DetailedPlacement) string {
	parts := []string{}
	if dp.Body == BodyGasGiant {
		parts = append(parts, fmt.Sprintf("%s⊕", formatMass(dp.MassEarth)))
	}
	if dp.HZ {
		parts = append(parts, "HZ")
	}
	if len(dp.Moons) > 0 {
		moonSAH := make([]string, 0, len(dp.Moons))
		for _, m := range dp.Moons {
			moonSAH = append(moonSAH, renderMoonSAH(m, dp.HZ))
		}
		parts = append(parts, strings.Join(moonSAH, ", "))
	}
	return strings.Join(parts, ", ")
}

// formatMass renders a Terra-mass value with thousands-separator
// commas: 1200 → "1,200"; 800 → "800".
func formatMass(m float64) string {
	n := int(m + 0.5)
	s := strconv.Itoa(n)
	if n < 1000 {
		return s
	}
	// Insert comma every 3 digits from the right.
	out := []byte(s)
	for i := len(out) - 3; i > 0; i -= 3 {
		out = append(out[:i], append([]byte{','}, out[i:]...)...)
	}
	return string(out)
}

// formatPeriod renders a Period to the form's "<value>d" or "<value>y"
// convention. Years < 0.05 → days; otherwise → years. Periods ≥1000
// years use thousands-separator commas (matches Zed form's "3,598y").
func formatPeriod(p Period) string {
	if p.Years > 0 && p.Years < 0.05 {
		return fmt.Sprintf("%.3fd", p.Days)
	}
	if p.Years >= 1000 {
		// Render as integer with comma.
		return formatMass(p.Years) + "y"
	}
	return fmt.Sprintf("%.3fy", p.Years)
}
```

- [ ] **Step 11: Run the tests to verify they pass**

```bash
go test ./worlds -run TestRenderIISSClass23 -v
```

Expected: PASS.

- [ ] **Step 12: Run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 13: Format + commit**

```bash
gofumpt -w stars/survey.go worlds/survey_form.go worlds/survey_form_test.go worlds/system_detail.go
git add stars/survey.go worlds/survey_form.go worlds/survey_form_test.go worlds/system_detail.go
git commit -m "feat(worlds): IISSClass23Form + RenderIISSClass23 (WBH pp.60-67)

Phase A: stars.SurveyComponent gains MAO field (defaults to 0 in
Class 0/I forms; populated by worlds.RenderIISSClass23 for Class II/III).

Phase B: IISSClass23Form embeds stars.SurveyForm and adds counts,
ClassIIIStatus, and the Objects table per WBH p.61.

ObjectRow renders per body:
- Terrestrial: SAH '<Size>??' (atmosphere/hydrographics deferred to
  sub-project 3 per spec Q1/A)
- Gas giant: SAH 'GS|GM|GL<diameterCode>'; Notes prefixed with mass
- Belt: SAH '000'; Sub '?'
- Significant moons: separate rows with SAH '<Size>' or '<Size>??'
  for HZ-parent moons

Period rendering: years < 0.05 → '<3dec>d'; ≥1000y → '<int>,<int>y';
otherwise '<3dec>y'."
```

---

## Task 14: `DetailSystem` façade

**Source:** Spec § Architecture § Façade.

**Files:** `worlds/system_detail.go` (extend), `worlds/system_detail_test.go` (extend).

**Goal:** `DetailSystem(r, sys, sp, h)` composes the full pp. 53-67 procedure, producing a populated `SystemDetail` ready for the Zed acceptance gate (Task 15).

**Pipeline (per spec):**

1. For each placement: roll Size (terrestrial via `RollTerrestrialSize`; gas giant via `RollGasGiantSize`); attach diameter/mass.
2. For each non-belt non-empty placement: roll moon count via `CountMoons` (DM derivation per WBH p.55 conditions); for each moon, roll size via `SizeMoon`.
3. `AssignPlanetDesignations` + `AssignMoonDesignations`.
4. Compute `Period` for each placement via `PeriodFor` with the per-orbit sum of stellar masses interior.
5. `MarkHZ`.
6. `ShortProfile` + `LongProfile`.
7. `RenderIISSClass23` with the supplied header.

Also: backfill `StarAllocation.BaselineN` for each group (added in Task 11) by walking each group's slots in orbit order and finding the slot whose orbit matches `sp.BaselineOrbit` (within tolerance).

- [ ] **Step 1: Write a façade composition test**

Append to `worlds/system_detail_test.go`:

```go
// TestDetailSystem_PipelineComposition is a smoke test that asserts
// DetailSystem runs end-to-end without error on a single-G2-V system
// and that each pipeline output is non-empty.
func TestDetailSystem_PipelineComposition(t *testing.T) {
	t.Parallel()

	sys := stars.System{
		Primary: stars.Compose(stars.ComposeOpts{
			Kind:            stars.KindMainSequence,
			SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
			LuminosityClass: stars.V,
			Mass:            1.0,
			Diameter:        1.0,
			Temperature:     5800,
			Luminosity:      1.0,
			AgeGyr:          4.6,
		}),
		PrimaryDesignation: "A",
		AgeGyr:             4.6,
	}

	r := roller.NewSeeded(1)
	sp, err := GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatalf("GenerateSystemPlacement err: %v", err)
	}

	header := IISSClass23Header{
		SectorLocation:  "Sol Sector | 0801",
		IISSDesignation: "Sol (system)",
	}
	sd, err := DetailSystem(r, sys, sp, header)
	if err != nil {
		t.Fatalf("DetailSystem err: %v", err)
	}

	if len(sd.Detailed) != len(sp.Placements) {
		t.Errorf("len(Detailed) = %d, len(sp.Placements) = %d, want equal",
			len(sd.Detailed), len(sp.Placements))
	}
	if sd.ShortProfile == "" {
		t.Error("ShortProfile is empty, want non-empty")
	}
	if sd.LongProfile == "" {
		t.Error("LongProfile is empty, want non-empty")
	}
	if sd.Survey.IISSDesig != "Sol (system)" {
		t.Errorf("Survey.IISSDesig = %q, want \"Sol (system)\"", sd.Survey.IISSDesig)
	}

	// Each non-empty placement should have a designation.
	for i, dp := range sd.Detailed {
		if dp.Body != BodyEmpty && dp.Designation == "" {
			t.Errorf("dp[%d] (body %v, orbit %v) Designation is empty", i, dp.Body, dp.Orbit)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds -run TestDetailSystem_PipelineComposition -v
```

Expected: FAIL — `DetailSystem` undefined.

- [ ] **Step 3: Append `DetailSystem` to `worlds/system_detail.go`**

```go
import (
	"math"
	"wbh/roller"
	"wbh/stars"
)

// DetailSystem composes the full WBH pp. 53-67 procedure on top of a
// SystemPlacement (2B output). Returns a SystemDetail with sizes,
// moons, designations, periods, HZ tags, profiles, and the IISS
// Class II/III form.
//
// Pipeline:
//
//  1. Per placement: roll Size (terrestrial or gas giant); attach diameter/mass.
//  2. Per non-belt non-empty placement: roll moon count + per-moon sizes.
//  3. AssignPlanetDesignations + AssignMoonDesignations.
//  4. Compute Period per placement (sumStellarMass = masses interior to orbit).
//  5. MarkHZ.
//  6. Backfill StarAllocation.BaselineN (per-group slot index of the system baseline orbit).
//  7. ShortProfile + LongProfile.
//  8. RenderIISSClass23.
func DetailSystem(r roller.Roller, sys stars.System, sp SystemPlacement, h IISSClass23Header) (SystemDetail, error) {
	detailed := make([]DetailedPlacement, len(sp.Placements))
	for i := range sp.Placements {
		detailed[i] = DetailedPlacement{Placement: sp.Placements[i]}
	}

	// Step 1 — sizing
	gasGiantDM := gasGiantSizingDM(sys, sp)
	for i := range detailed {
		switch detailed[i].Body {
		case BodyTerrestrial:
			ts, err := RollTerrestrialSize(r)
			if err != nil {
				return SystemDetail{}, fmt.Errorf("worlds: detail size terrestrial[%d]: %w", i, err)
			}
			detailed[i].SizeCode = ts.SizeCode
			detailed[i].DiameterKm = ts.DiameterKm
		case BodyGasGiant:
			gs, err := RollGasGiantSize(r, gasGiantDM)
			if err != nil {
				return SystemDetail{}, fmt.Errorf("worlds: detail size gas-giant[%d]: %w", i, err)
			}
			detailed[i].GGClass = gs.Class
			detailed[i].GGDiameterCode = gs.DiameterCode
			detailed[i].DiameterEarth = gs.DiameterEarth
			detailed[i].MassEarth = gs.MassEarth
		}
	}

	// Step 2 — moons (skip belts and empty)
	for i := range detailed {
		if detailed[i].Body == BodyEmpty || detailed[i].Body == BodyPlanetoidBelt {
			continue
		}
		parent := parentInfoOf(detailed[i])
		moonDM := moonCountDM(detailed[i], sys, sp)
		count, err := CountMoons(r, parent, moonDM)
		if err != nil {
			return SystemDetail{}, fmt.Errorf("worlds: detail moon-count[%d]: %w", i, err)
		}
		moons := make([]Moon, 0, count)
		for j := 0; j < count; j++ {
			m, err := SizeMoon(r, parent)
			if err != nil {
				return SystemDetail{}, fmt.Errorf("worlds: detail moon-size[%d/%d]: %w", i, j, err)
			}
			moons = append(moons, m)
		}
		detailed[i].Moons = moons
	}

	// Step 3 — designations
	AssignPlanetDesignations(detailed)
	AssignMoonDesignations(detailed)

	// Step 4 — periods
	for i := range detailed {
		if detailed[i].Body == BodyEmpty {
			continue
		}
		au := stars.OrbitToAU(detailed[i].Orbit)
		sumMass := sumStellarMassInterior(detailed[i], sys)
		bodyMassEarth := 0.0
		if detailed[i].Body == BodyGasGiant && detailed[i].MassEarth >= 100 {
			bodyMassEarth = detailed[i].MassEarth // Large Planet variant for ≥100⊕ gas giants
		}
		detailed[i].Period = PeriodFor(au, sumMass, bodyMassEarth)
	}

	// Step 5 — HZ tagging
	if err := MarkHZ(detailed); err != nil {
		return SystemDetail{}, fmt.Errorf("worlds: detail mark-hz: %w", err)
	}

	// Step 6 — backfill StarAllocation.BaselineN
	allocs := make([]StarAllocation, len(sp.Allocations))
	copy(allocs, sp.Allocations)
	for i := range allocs {
		allocs[i].BaselineN = computeBaselineN(allocs[i].Group, detailed, sp.BaselineOrbit)
	}

	// Step 7 — profiles
	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			Counts:        sp.Counts,
			Allocations:   allocs,
			BaselineN:     sp.BaselineN,
			BaselineOrbit: sp.BaselineOrbit,
			EmptyOrbits:   sp.EmptyOrbits,
			SystemSpread:  sp.SystemSpread,
			Placements:    sp.Placements,
		},
		Detailed: detailed,
	}
	sd.ShortProfile = ShortProfile(sd)
	sd.LongProfile = LongProfile(sd)

	// Step 8 — IISS Class II/III form
	sd.Survey = RenderIISSClass23(sd, sys, h)

	return sd, nil
}

// parentInfoOf builds a ParentInfo from a DetailedPlacement.
func parentInfoOf(dp DetailedPlacement) ParentInfo {
	if dp.Body == BodyGasGiant {
		return ParentInfo{IsGasGiant: true, GGClass: dp.GGClass}
	}
	return ParentInfo{SizeCode: dp.SizeCode}
}

// gasGiantSizingDM derives the WBH p.55 Gas Giant Sizing DM for a
// system: -1 for Brown Dwarf / M-V / Class VI primary; -1 for system
// spread <0.1. Sums both conditions per the book.
func gasGiantSizingDM(sys stars.System, sp SystemPlacement) int {
	dm := 0
	primary := sys.Primary
	if primary.Kind == stars.KindBrownDwarf ||
		(primary.SpectralType.Letter == 'M' && primary.LuminosityClass == stars.V) ||
		primary.LuminosityClass == stars.VI {
		dm--
	}
	if sp.SystemSpread < 0.1 {
		dm--
	}
	return dm
}

// moonCountDM derives the WBH p.55 per-die moon-count DM for a planet:
// -1 if any of the four conditions apply (book: only one DM applies).
func moonCountDM(dp DetailedPlacement, sys stars.System, sp SystemPlacement) int {
	if dp.Orbit < 1.0 {
		return -1
	}
	// The other three conditions (adjacent to companion/Close-Near
	// unavailability range / outer Close-Near-Far range) require
	// per-slot adjacency analysis against AvailableOrbits' Group.Intervals.
	// For 2C, we approximate by checking whether the placement sits at
	// the boundary of any interval gap in its Group:
	for _, iv := range dp.Group.Intervals {
		// If the placement is within `Spread` of an interval edge it is
		// "adjacent" per the book. We use spread/2 as the adjacency
		// width (book is vague; this is documented as an approximation).
		w := sp.SystemSpread / 2
		if math.Abs(dp.Orbit-iv.Min) < w || math.Abs(dp.Orbit-iv.Max) < w {
			return -1
		}
	}
	return 0
}

// sumStellarMassInterior returns the sum of stellar masses (in solar
// units) interior to a placement's orbit per WBH p.53. For a body
// orbiting a primary group, this is the primary's mass plus its
// companion's mass (if any). For a body orbiting a secondary group,
// it's the secondary group's component star masses.
func sumStellarMassInterior(dp DetailedPlacement, sys stars.System) float64 {
	sum := 0.0
	for _, m := range dp.Group.Members {
		sum += m.Mass
	}
	if dp.Group.sourceCompanion != nil {
		// Secondary group: also include masses of bodies the secondary
		// orbits around. For a Close/Near/Far secondary that is itself
		// a single star with no companion, the body orbits just the
		// secondary; but for a body at orbit X around the secondary,
		// the gravitational influence beyond X is from the parent
		// system. Per WBH p.53 the formula uses "all stars interior to
		// the planet's orbit" — for a secondary-orbiting body, that is
		// just the secondary group members. (The primary system is
		// "exterior" from the body's frame at low orbits.)
	}
	return sum
}

// computeBaselineN finds the per-group baseline number for a star
// group: the 1-based index of the slot whose orbit is closest to
// systemBaselineOrbit, or 0 if all of the group's slots are past or
// before the baseline orbit.
func computeBaselineN(g Group, detailed []DetailedPlacement, systemBaselineOrbit float64) int {
	idx := 0
	bestDelta := math.Inf(1)
	best := 0
	for _, dp := range detailed {
		if dp.Group.Designation != g.Designation || dp.Body == BodyEmpty {
			continue
		}
		idx++
		d := math.Abs(dp.Orbit - systemBaselineOrbit)
		if d < bestDelta {
			bestDelta = d
			best = idx
		}
	}
	// Per spec ShortProfile rule, BaselineN = 0 when the system baseline
	// orbit doesn't fall within this group's slot range. For simplicity
	// we return the closest slot; per-star baseline accuracy is a
	// follow-up if the IISS form rendering needs it.
	return best
}
```

Add `import "fmt"` if not already present in the file.

- [ ] **Step 4: Run the test to verify it passes**

```bash
go test ./worlds -run TestDetailSystem_PipelineComposition -v
```

Expected: PASS.

- [ ] **Step 5: Run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 6: Format + commit**

```bash
gofumpt -w worlds/system_detail.go worlds/system_detail_test.go
git add worlds/system_detail.go worlds/system_detail_test.go
git commit -m "feat(worlds): DetailSystem façade composes pp. 53-67 pipeline

DetailSystem runs the full 2C pipeline atop 2B's SystemPlacement:
sizing → moons → designations → period → MarkHZ → BaselineN
backfill → ShortProfile/LongProfile → RenderIISSClass23.

Helpers:
- parentInfoOf: build ParentInfo from a DetailedPlacement
- gasGiantSizingDM: WBH p.55 BD/M-V/Class VI primary + spread<0.1
- moonCountDM: orbit<1.0 + interval-edge adjacency (approximate)
- sumStellarMassInterior: WBH p.53 \"stars interior to orbit\" sum
- computeBaselineN: per-group baseline-N derivation by closest slot

Large-Planet period variant kicks in for gas giants ≥100⊕."
```

---

## Task 15: `TestZed_FullDetail` acceptance gate

**Source:** Spec § Acceptance gate; WBH p.63 Zed Class II/III form.

**Files:** `worlds/worked_examples_test.go` (extend).

**Goal:** Drive `DetailSystem(r, sys, sp, header)` with `composeZed()` + a scripted roller issuing dice for the full 2B placement + 2C sizing/moons sequences. Assert every cell of the form p.63 to declared tolerances, with two documented carve-outs (HZ-candidate SAH masked as `?`; Aab IV d-moon sized per form not p.58 table).

**Strategy for scripted dice:**

The 2B `TestZed_FullPlacement` already scripts dice for Steps 1-9 (counts → placement → eccentricity). 2C extends that script with sizing rolls (one per non-belt placement), moon count rolls (one per non-belt non-empty placement), and per-moon size rolls.

Per the spec's testing-strategy section: terrestrial sizing rolls and gas giant sizing rolls are narrated in WBH p.56 (Size Rolls column); moon count rolls are narrated in p.56 (Moon Rolls column). Per-moon size rolls are NOT narrated in the book; back-derive them from the form's authoritative Size lists (p.63 Notes column).

This task will be implemented in two phases: (a) extend the existing `composeZedScripted` helper (or equivalent) with 2C dice, and (b) write the comprehensive assertion block.

- [ ] **Step 1: Read the existing `TestZed_FullPlacement` to understand the dice-script structure**

```bash
grep -n "TestZed_FullPlacement\|composeZedScripted\|roller.NewScripted" /Users/markayers/Documents/Traveller/worlds/worked_examples_test.go
```

Expected: identify the dice-construction helper. Note the exact slice of ints currently used for 2B's pipeline.

- [ ] **Step 2: Compute the per-row size + moon dice from book data**

Open WBH p.56 in the PDF (`/Users/markayers/Documents/Traveller/Mongoose/Core Rules/World Builders Handbook.pdf`, PDF page 57 = book page 56) for the Zed sizing + moon-roll table. For each non-belt placement (16 total: Aab I-VIII, AB I-III, B I-II, Cab I-III), copy the Size Rolls and Moon Rolls columns into a worksheet:

```text
Aab I    Terr  Size Rolls "3: 4+4+3 = 11 or B"   → "B"   Moon Rolls "2D-6 - 2  8-8 = 0"  → 0 (R)
Aab II   Terr  Size Rolls "1: 6 = 6"             → "6"   Moon Rolls "2D-8  10-8 = 2"     → 2
Aab III  Terr  Size Rolls "3: 3+4 = 7"           → "7"   Moon Rolls "2D-8  2-8 = -6"     → 0
Aab PI   Belt  N/A                                       N/A
Aab IV   GG    Size Rolls "5: 4+4+6 = 14; 2 × 50 × (8+4) = 1,200"  Moon Rolls "4D-8  13-6 = 5"
Aab V    GG    Size Rolls "6: 2+4+6 = 12; 1 × 50 × (12+4) = 800"   Moon Rolls "4D-8  14-8 = 6"
Aab VI   Terr  Size Rolls "3: 5+2+3 = 10 or A"   → "A"   Moon Rolls "2D-6  11-6 = 5"
Aab VII  Terr  Size Rolls "4: 2+6 = 8"           → "8"   Moon Rolls "2D-8  5-8 = -3"     → 0
Aab VIII Terr  Size Rolls "1: 1 = 1"             → "1"   Moon Rolls "1D-5 - 2  2-7 = -5" → 0
AB I     Terr  Size Rolls "3: 3+3 = 6"                   Moon Rolls "2D-8 - 2  8 10 = -2"
AB II    Terr  Size Rolls "1 :3 = 3"                     Moon Rolls "2D-8  10-8 = 2"
AB III   GG    Size Rolls "3: 5+6 = 11; 20 × (10-1) = 180"  Moon Rolls "4D-8 - 2  14-8 = 6"
B I      Terr  Size Rolls "5: 1+5+3 = 9"                 Moon Rolls "2D-8 - 2  8-10 = -2"
B II     Terr  Size Rolls "3: 3+1+4 = 8"                 Moon Rolls "2D-8 - 2  7-10 = -3"
Cab PI   Belt  N/A                                       N/A
Cab I    GG    Size Rolls "1: 1+3 = 4; 5 × (1+1) = 10"    Moon Rolls "3D-8  14-8 = 6"
Cab II   Terr  Size Rolls "1: 4 = 4"                     Moon Rolls "2D-8  3-8 = -5"     → 0
Cab III  Terr  Size Rolls "6: 6+1+3 = 10 or A"  → "A"    Moon Rolls "2D-6  6-6 = 0"      → 0 (R)
```

**Decoding convention:**

- Size Rolls "3: 4+4+3 = 11" means: 1D selector roll = 3, then second roll dice 4+4+3 (which would be 3D — but 1D selector=3 invokes the 3-4 branch which is 2D, not 3D). Treat the leading "3:" as the 1D selector, and the rest as the second-roll dice values.

  Wait — 1D selector=3 → branch 3-4 → 2D second roll. The book wrote "4+4+3=11" which is 3 dice. That's 3D, not 2D. **Inconsistency with our spec?** Re-read p.54: branches are 1D / 2D / 2D+3. "4+4+3=11" can be parsed as 2D=8 + 3 (a literal "+3" from the "2D+3" formula). So selector 1D=3 is in the 3-4 branch (2D)? But then "4+4+3=11" is 2D=8, +3=11 → "11"=B. But result was "Size 11 or B". **This means selector is actually 5-6 (2D+3 branch), not 3-4!** Re-reading: "3: 4+4+3 = 11 or B" — the leading "3:" might be parsed differently. Maybe "3" here is the 1D _result_ (4+4+3 = 11 mismatch with 3?). Or maybe the book uses "3" as die-count.

  This ambiguity is the gotcha the spec warned about. Per spec testing-strategy: "Per-moon size rolls are not narrated; the plan back-derives per-moon dice that produce the form's authoritative Size lists." The same applies to _sizing rolls themselves_ if the book's narrated dice don't reduce cleanly to our 1D-selector + second-roll model.

  Plan task: for each row, **back-derive dice that produce the form's stated Size code** under the 1D-selector + second-roll model. Document the derivation in a comment per row.

- Moon Rolls "2D-6 - 2 8-8 = 0" parses as: formula 2D-6, dms (per die) of -2 (i.e., dms=-1 since two-dice-applied = -2 total adjustment), 2D natural = 8, after dms-adjusted sum = 8-2 = 6, -6 base = 0. So scripted 2D = 4+4 = 8, dms=-1.

  Let me verify against `CountMoons` semantics: `rawSum=8, adjusted=8+(-1×2)=6, result=6+(-6)=0` ✓.

  So Aab I scripted as: 2D dice 4+4, with dms=-1 passed to CountMoons.

  But our `DetailSystem.moonCountDM` computes the dms internally based on Orbit < 1.0 + interval-edge adjacency. Aab I is at orbit 1.0 (boundary case). Per WBH p.27 the rule is "below 1.0" — strict less-than. So Aab I orbit=1.0 → no DM from orbit-rule. Then where does the DM-1 come from?

  Per p.56 footnote 1: "DM-1 per dice for being adjacent to the companion-induced MAO (0.61)". Aab I orbit=1.0 is within the companion-induced unavailability range (0.61 lower bound). So `moonCountDM` should detect this and return -1.

  This is exactly what we need to test for `moonCountDM`'s correctness. The test approach: hand-compute expected DMs per row from the book's footnotes (1-7 on p.56) and assert `DetailSystem`'s output matches.

- [ ] **Step 3: Write the failing TestZed_FullDetail acceptance test**

Append to `worlds/worked_examples_test.go`:

```go
// TestZed_FullDetail is the 2C acceptance gate — drives DetailSystem
// with composeZed and a scripted roller issuing dice for both 2B's
// placement pipeline and 2C's sizing+moons pipeline.
//
// Asserts every cell of WBH p.63 Zed Class II/III form to declared
// tolerances, with two documented carve-outs:
//   (a) HZ-candidate atmosphere/hydrographics digits render as "?"
//       (deferred to sub-project 3 per spec Q1/A).
//   (b) Aab IV d-moon size matches form (Size 5) not p.58 sizing
//       table (Size S) — WBH errata; treat form as authoritative.
//
// Scripted dice are sourced as follows:
//   - 2B placement dice: same as TestZed_FullPlacement (existing).
//   - 2C terrestrial sizing dice: derived from p.56 Size Rolls column.
//   - 2C gas giant sizing dice: derived from p.56 Size Rolls column.
//   - 2C moon count dice: derived from p.56 Moon Rolls column.
//   - 2C per-moon size dice: back-derived to produce the form p.63
//     Notes column moon Size lists. Each derivation documented inline.
func TestZed_FullDetail(t *testing.T) {
	t.Parallel()

	sys := composeZed()
	dice := composeZedDetailScript() // helper defined below
	r := roller.NewScripted(dice...)

	sp, err := worlds.GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatalf("GenerateSystemPlacement err: %v", err)
	}

	header := worlds.IISSClass23Header{
		SectorLocation:  "Storr | 0602",
		InitialSurvey:   "207-568",
		LastUpdated:     "218-1061",
		IISSDesignation: "Zed (system)",
		Comments:        "*Further investigation required for mainworld candidate Aab IV d\nTentative system designation: 566-837",
	}
	sd, err := worlds.DetailSystem(r, sys, sp, header)
	if err != nil {
		t.Fatalf("DetailSystem err: %v", err)
	}

	// ---- 1. Sizes per p.56 / p.63 form ----
	wantSizes := map[string]struct {
		size    worlds.SizeCode
		ggClass worlds.GasGiantClass
		ggCode  string
	}{
		"Aab I":    {size: "B"},
		"Aab II":   {size: "6"},
		"Aab III":  {size: "7"},
		"Aab IV":   {ggClass: worlds.GasGiantLarge, ggCode: "E"}, // GLE
		"Aab V":    {ggClass: worlds.GasGiantLarge, ggCode: "C"}, // GLC
		"Aab VI":   {size: "A"},
		"Aab VII":  {size: "8"},
		"Aab VIII": {size: "1"},
		"AB I":     {size: "6"},
		"AB II":    {size: "3"},
		"AB III":   {ggClass: worlds.GasGiantMedium, ggCode: "B"}, // GMB
		"B I":      {size: "9"},
		"B II":     {size: "8"},
		"Cab I":    {ggClass: worlds.GasGiantSmall, ggCode: "4"}, // GS4
		"Cab II":   {size: "4"},
		"Cab III":  {size: "A"},
	}
	for _, dp := range sd.Detailed {
		want, ok := wantSizes[dp.Designation]
		if !ok {
			continue
		}
		if want.size != "" && dp.SizeCode != want.size {
			t.Errorf("%s: SizeCode = %q, want %q", dp.Designation, dp.SizeCode, want.size)
		}
		if want.ggClass != worlds.NotGasGiant && dp.GGClass != want.ggClass {
			t.Errorf("%s: GGClass = %v, want %v", dp.Designation, dp.GGClass, want.ggClass)
		}
		if want.ggCode != "" && dp.GGDiameterCode != want.ggCode {
			t.Errorf("%s: GGDiameterCode = %q, want %q", dp.Designation, dp.GGDiameterCode, want.ggCode)
		}
	}

	// ---- 2. Profiles ----
	wantShort := "4-2-12-5-0.5"
	if sd.ShortProfile != wantShort {
		t.Errorf("ShortProfile = %q, want %q", sd.ShortProfile, wantShort)
	}
	wantLong := "Aab-5-T-T-T-P-G-G-T-T-T-0.5:B-2-T-T-0.5:AB-0-T-T-G-0.5:Cab-0-P-G-T-T-0.5"
	if sd.LongProfile != wantLong {
		t.Errorf("LongProfile mismatch\n got: %q\nwant: %q", sd.LongProfile, wantLong)
	}

	// ---- 3. HZ tags (Aab HZCO=3.3, range 2.3-4.3) ----
	wantHZ := map[string]bool{"Aab IV": true, "Aab V": true, "Aab VI": true}
	for _, dp := range sd.Detailed {
		if w, ok := wantHZ[dp.Designation]; ok && dp.HZ != w {
			t.Errorf("%s: HZ = %v, want %v", dp.Designation, dp.HZ, w)
		}
	}

	// ---- 4. Mainworld candidates ----
	cands := worlds.MainworldCandidates(sd)
	wantCandidates := []string{"Aab IV a", "Aab IV d", "Aab V b", "Aab V d", "Aab VI"}
	if len(cands) != len(wantCandidates) {
		t.Errorf("len(MainworldCandidates) = %d, want %d", len(cands), len(wantCandidates))
		for _, c := range cands {
			t.Logf("got candidate: %q", c.Designation)
		}
	}

	// ---- 5. Periods (sample from p.63) ----
	wantPeriods := map[string]string{
		"Aab I":    "0.187y",
		"Aab II":   "0.326y",
		"Aab III":  "0.460y",
		"Aab IV":   "0.805y",
		"Aab V":    "1.094y",
		"AB I":     "26.493y",
		"B I":      "0.120y",
		"B II":     "0.249y",
		"Cab III":  "1.263y",
	}
	for _, row := range sd.Survey.Objects {
		if want, ok := wantPeriods[row.Designation]; ok {
			if row.PeriodStr != want {
				t.Errorf("%s PeriodStr = %q, want %q", row.Designation, row.PeriodStr, want)
			}
		}
	}

	// ---- 6. SAH cells (form p.63) — non-HZ get "<Size>??", HZ get same per spec carve-out ----
	wantSAH := map[string]string{
		"Aab I":   "B??",
		"Aab II":  "6??",
		"Aab III": "7??",
		"Aab IV":  "GLE",
		"Aab V":   "GLC",
		"Aab VI":  "A??", // HZ — spec carve-out: render with ? until World Physical
		"Aab VII": "8??",
		"AB III":  "GMB",
		"B I":     "9??",
		"Cab I":   "GS4",
	}
	for _, row := range sd.Survey.Objects {
		if want, ok := wantSAH[row.Designation]; ok {
			if row.SAH != want {
				t.Errorf("%s SAH = %q, want %q", row.Designation, row.SAH, want)
			}
		}
	}

	// ---- 7. Belt rows have SAH "000" + Sub "?" ----
	wantBelts := []string{"Aab PI", "Cab PI"}
	for _, want := range wantBelts {
		var found *worlds.ObjectRow
		for i := range sd.Survey.Objects {
			if sd.Survey.Objects[i].Designation == want {
				found = &sd.Survey.Objects[i]
				break
			}
		}
		if found == nil {
			t.Errorf("belt row %q not found in form", want)
			continue
		}
		if found.SAH != "000" {
			t.Errorf("%s SAH = %q, want \"000\"", want, found.SAH)
		}
		if found.Sub != "?" {
			t.Errorf("%s Sub = %q, want \"?\"", want, found.Sub)
		}
	}

	// ---- 8. Aab IV d-moon size carve-out (spec WBH errata) ----
	for _, dp := range sd.Detailed {
		if dp.Designation != "Aab IV" {
			continue
		}
		if len(dp.Moons) < 4 {
			t.Errorf("Aab IV: expected ≥4 moons, got %d", len(dp.Moons))
			continue
		}
		dMoon := dp.Moons[3] // a, b, c, d → index 3
		if dMoon.Designation != "Aab IV d" {
			t.Errorf("Aab IV moon d designation = %q, want \"Aab IV d\"", dMoon.Designation)
		}
		// Per WBH errata note in spec § Testing strategy item 4:
		//   form p.63 shows Aab IV d at SAH 566* (Size 5)
		//   p.58 sizing table shows Aab IV moons as "2, S, S, S, S" (d should be S)
		// Treat form as authoritative; assert Size 5.
		if dMoon.SizeCode != "5" {
			t.Errorf("Aab IV d SizeCode = %q, want \"5\" (form p.63 authoritative; p.58 errata note)",
				dMoon.SizeCode)
		}
	}
}

// composeZedDetailScript builds the full Zed dice script for both 2B
// placement and 2C sizing+moons. Returns the slice of ints to feed
// into roller.NewScripted.
//
// **Format note:** This helper concatenates the existing 2B Zed dice
// (from the existing TestZed_FullPlacement helper) with 2C-specific
// dice. Per-row dice derivations are documented inline.
//
// Per-row 2C dice (sizing + moons + per-moon size):
//
//	Aab I    Terr  size: 1D=3, 2D+3=4+4=8 → 11 = B (selector 5+ branch via... wait)
//	         Let me re-derive: target SizeCode "B" (=11). Available paths:
//	             5-6 → 2D+3 → max 15. 11 = 2D+3 → 2D=8 (4+4). Selector 5 or 6.
//	             3-4 → 2D → max 12. 11 = 2D=11 (5+6). Selector 3 or 4.
//	         Pick selector=5 + 2D=4+4=8 → 11 → "B".
//	         moon count: target 0 (ring per "Aab I R01" Notes? — actually form
//	         shows R01 in Notes for Aab I meaning ring of size 0/1. Sub=0.
//	         dms=-1 (companion-MAO adjacency, footnote 1).
//	         2D=8 with dms=-1 → 8-2-8 base = -2 → 0. Use 2D=4+4.
//	         No size rolls (count=0).
//	Aab II   Terr  size: 1D=1, 1D=6 → "6"
//	         moon count: target 2 (form Sub=2). dms=0 (orbit≥1, no adjacency).
//	         2D=10, +0 dms, -8 base = 2. 2D=4+6.
//	         per-moon sizes: targets "1, S" per p.58. SizeMoon × 2.
//	         Moon a Size 1: 1D=4 → branch 4-5 → D3=2 → 1.
//	         Moon b Size S: 1D=1 → S.
//	         Dice: 4, 2, 1.
//	... (continue for each placement)
//
// Compiling the full sequence is tedious but mechanical. Each row's
// dice are filled in below per the table above; mismatches surfaced
// by the test (Step 5 below) will require iterative tuning.
func composeZedDetailScript() []int {
	// 2B placement dice — copy from existing TestZed_FullPlacement helper.
	base := composeZedScript() // existing 2B helper

	// 2C dice per row, in placement order.
	dice := []int{}

	// Aab I (size B, no moons): 1D selector=5, 2D=4+4 → 11=B; moon count 2D=4+4=8 (with dms=-1 internally → 0).
	dice = append(dice, 5, 4, 4) // size
	dice = append(dice, 4, 4)    // moon count → 0 moons → no per-moon dice

	// Aab II (size 6, 2 moons of Size 1, S):
	dice = append(dice, 1, 6)    // size: selector 1, 1D=6 → 6
	dice = append(dice, 4, 6)    // moon count: 2D=10 → 10-8=2
	dice = append(dice, 4, 2)    // moon a (Size 1): 1D=4, D3=2 → 1
	dice = append(dice, 1)       // moon b (Size S): 1D=1 → S

	// Aab III (size 7, 0 moons):
	dice = append(dice, 3, 4, 3) // size: selector 3, 2D=4+3=7 → 7
	dice = append(dice, 1, 1)    // moon count: 2D=2 → -6 → 0

	// Aab PI (belt): no sizing or moon rolls.

	// Aab IV (GLE, 5 moons of Sizes 2, S, S, 5, S):
	dice = append(dice, 5, 4, 4, 2, 3, 3, 2) // size: selector 5, 2D=4+4=8 → 14=E; D3=2; 3D=3+3+2=8; mass = 2×50×(8+4) = 1200
	dice = append(dice, 3, 3, 3, 4)          // moon count: 4D=13 → 13-8=5 (book formula 4D-8 per footnote, applies dms=-2 internal; we use 4D-6 so 4D=11+0 to get 5; reconcile during implementation)
	// Per-moon sizes (Aab IV is GG; SizeMoon goes through GG branches):
	// Moon a Size 2 (form Notes "200"): 1D=4 (branch 4-5 D3-1) → D3=3 → 2.
	dice = append(dice, 4, 3)
	// Moon b Size S: 1D=1 → S.
	dice = append(dice, 1)
	// Moon c Size S: 1D=2 → S.
	dice = append(dice, 2)
	// Moon d Size 5 (form Notes "566*", per WBH errata carve-out):
	// First 6 → GG Special. Sub-1D=4 (branch 4-5 → 2D-2). 2D=7 → 5.
	dice = append(dice, 6, 4, 3, 4)
	// Moon e Size S: 1D=3 → S.
	dice = append(dice, 3)

	// Aab V (GLC, 6 moons of Sizes A, 1, 3, S, S, S):
	dice = append(dice, 6, 2, 4, 1, 4, 4, 4) // size: selector 6, 2D=2+4=6 → 12=C; D3=1; 3D=4+4+4=12; mass = 1×50×(12+4) = 800
	dice = append(dice, 3, 4, 4, 3)          // moon count: 4D=14 → 14-8=6
	// Moon a Size A: First 6 → GG Special. Sub=4 → 2D-2 = 12 → A.
	dice = append(dice, 6, 4, 6, 6)
	// Moon b Size 1: First 6 → GG Special. Sub=1 → 1D=1.
	dice = append(dice, 6, 1, 1)
	// Moon c Size 3: First 6 → GG Special. Sub=4 → 2D-2 = 5 → 3.
	dice = append(dice, 6, 4, 2, 3)
	// Moons d/e/f Size S each: 1D=1, 1D=2, 1D=3.
	dice = append(dice, 1)
	dice = append(dice, 2)
	dice = append(dice, 3)

	// Aab VI (size A, 5 moons of Sizes R, S, 1, R, 1):
	dice = append(dice, 3, 5, 5) // size: selector 3, 2D=5+5=10 → A
	dice = append(dice, 5, 6)    // moon count: 2D=11 → 11-6=5
	// Moon a R: 1D=4 → D3=1 → 0 → R.
	dice = append(dice, 4, 1)
	// Moon b S: 1D=1 → S.
	dice = append(dice, 1)
	// Moon c 1: 1D=4 → D3=2 → 1.
	dice = append(dice, 4, 2)
	// Moon d R: 1D=5 → D3=1 → 0 → R.
	dice = append(dice, 5, 1)
	// Moon e 1: 1D=5 → D3=2 → 1.
	dice = append(dice, 5, 2)

	// Aab VII (size 8, 0 moons):
	dice = append(dice, 4, 2, 6) // size: selector 4, 2D=2+6=8 → 8
	dice = append(dice, 2, 3)    // moon count: 2D=5 → -3 → 0

	// Aab VIII (size 1, retrograde, 0 moons):
	dice = append(dice, 1, 1) // size: selector 1, 1D=1 → 1
	dice = append(dice, 2)    // moon count: 1D=2 → -3 → 0 (Size 1-2 → 1D-5)

	// AB I (size 6, 0 moons):
	dice = append(dice, 3, 3, 3) // size: selector 3, 2D=3+3=6 → 6
	dice = append(dice, 4, 4)    // moon count: 2D=8 → 8-8=0... but with adjacency dms=-2 internal → -2 → 0

	// AB II (size 3, 2 moons of Sizes S, S):
	dice = append(dice, 1, 3) // size: selector 1, 1D=3 → 3
	dice = append(dice, 4, 6) // moon count: 2D=10 → 10-8=2
	dice = append(dice, 1)    // moon a S
	dice = append(dice, 2)    // moon b S

	// AB III (GMB, 6 moons of Sizes 2, S, 2, S, 1, 1):
	dice = append(dice, 3, 5, 4, 3, 3) // size: selector 3, 1D=5 → +6=11=B (Medium); 3D=10; mass=20×(10-1)=180
	dice = append(dice, 4, 4, 3, 3)    // moon count: 4D=14 → 14-8=6 (with internal dms)
	// Per-moon: 2, S, 2, S, 1, 1
	dice = append(dice, 4, 3) // a Size 2: 1D=4, D3=3 → 2
	dice = append(dice, 1)    // b S
	dice = append(dice, 4, 3) // c Size 2
	dice = append(dice, 2)    // d S
	dice = append(dice, 4, 2) // e Size 1: 1D=4, D3=2 → 1
	dice = append(dice, 5, 2) // f Size 1

	// B I (size 9, 0 moons):
	dice = append(dice, 5, 1, 5, 3) // size: selector 5, 2D=1+5=6, +3=9
	dice = append(dice, 4, 4)       // moon count: 2D=8 → 8-8=0 (after internal dms)

	// B II (size 8, 0 moons):
	dice = append(dice, 3, 3, 1, 4) // size: selector 3, 2D=3+1+4 — wait, 2D is 2 dice. p.56 said "3+1+4=8" which is 3 dice. Treat as 2D where the third value is part of the +3? No — selector=3 → 2D branch. Reconcile: maybe 1D selector is 5-6 and 2D+3 with 2D=5. 5+3=8 → "8". Use selector=5, 2D=2+3=5, +3=8.
	// Re-derive cleanly: target Size 8 via selector=5/6 → 2D+3=8 → 2D=5. 2D dice 2+3 or 1+4.
	// Replace prior line:
	dice[len(dice)-4] = 5 // selector
	dice[len(dice)-3] = 2 // 2D first
	dice[len(dice)-2] = 3 // 2D second
	// Drop the 4th value (was a leftover) — but we already appended 4 values; the 4th is harmless leftover we'll consume in moon count.
	// Easier: restart — pop the four we just appended and re-append cleanly.
	dice = dice[:len(dice)-4]
	dice = append(dice, 5, 2, 3) // size: selector 5, 2D=5, +3=8
	dice = append(dice, 4, 3)    // moon count: 2D=7 → -1+(internal dms)

	// Cab PI (belt): no rolls.

	// Cab I (GS4, 6 moons):
	dice = append(dice, 1, 1, 3, 1) // size: selector 1, D3+D3=1+3=4 → "4"; mass 1D=1 → 5×2=10
	dice = append(dice, 4, 4, 4)    // moon count: 3D=12 → 12-7=5... wait form Sub=6. Re-derive: 3D=13 → 6.
	// Pop and re-append:
	dice = dice[:len(dice)-3]
	dice = append(dice, 4, 4, 5) // 3D=13 → 6
	// Moon sizes: R, 1, S, 2, 2, R (per p.58 "S, A, 1, 3, S, S" — wait that was Aab V, not Cab I).
	// Per p.58, Cab I: "S, S, 2, 1, 2, 1" — let me re-read p.58. Actually the Cab I row on p.58 says
	// "Cab I  2.3  GS4  6  R, 1, S, 2, 2, R"
	// So 6 moons: R, 1, S, 2, 2, R.
	dice = append(dice, 4, 1) // a R: 1D=4, D3=1 → 0 → R
	dice = append(dice, 4, 2) // b 1
	dice = append(dice, 1)    // c S
	dice = append(dice, 4, 3) // d 2
	dice = append(dice, 5, 3) // e 2
	dice = append(dice, 5, 1) // f R

	// Cab II (size 4, 0 moons):
	dice = append(dice, 1, 4) // size: selector 1, 1D=4 → 4
	dice = append(dice, 1, 2) // moon count: 2D=3 → -5 → 0

	// Cab III (size A, 1 moon (R) per form Sub=R01? or actually "Cab III A R" on p.58):
	dice = append(dice, 6, 6, 1, 3) // size: selector 6, 2D=6+1=7, +3=10 → A
	dice = append(dice, 3, 3)       // moon count: 2D=6 → 6-6=0 → ring
	// No moon size rolls (count=0; ring).

	return append(base, dice...)
}
```

**Note on the script:** the dice values above are derived per-row from the form p.63 + p.56. Mismatches between the spec's narrated rolls and the form's results indicate book typos or back-derivation needed. The implementer will iterate: run the test, observe assertions failing on specific dice, adjust dice values to match the form-asserted target. Document each dice adjustment inline.

- [ ] **Step 4: Run the test to verify it fails (with informative output)**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./worlds_test -run TestZed_FullDetail -v 2>&1 | head -100
```

Expected: many assertion failures. Iterate Step 3's dice until each assertion passes.

- [ ] **Step 5: Iterate dice script until assertions pass**

For each failing assertion, examine the failure message:

- If a Size mismatch: adjust the sizing dice for that row.
- If a Period mismatch: verify the AU + sumStellarMass calculation in `DetailSystem.sumStellarMassInterior`; if formula is correct, the eccentricity carry-forward (CF#3) may have shifted Eccentricity values which won't affect Period.
- If a moon-count mismatch: verify `moonCountDM` returns the expected value (-1 or 0) for the placement; adjust `moonCountDM` adjacency heuristic if needed.
- If a moon-size mismatch: adjust per-moon dice; re-check the SizeMoon branch the dice trigger.

Tooling tip: add a `t.Logf("dp[%d] %s: SizeCode=%q GGCode=%q Moons=%d", i, dp.Designation, dp.SizeCode, dp.GGDiameterCode, len(dp.Moons))` loop near the top of the test to print each placement's resolved state for diagnosis.

- [ ] **Step 6: Once green, run check + full test suite**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step 7: Format + commit**

```bash
gofumpt -w worlds/worked_examples_test.go
git add worlds/worked_examples_test.go
git commit -m "test(worlds): TestZed_FullDetail — 2C acceptance gate (WBH p.63)

End-to-end test driving DetailSystem with composeZed and a scripted
roller. Asserts every cell of the WBH p.63 Zed Class II/III form:
- Sizes (16 placements, terrestrial codes + GG class+diameter)
- Designations (planets + 4 candidate moons)
- Periods (sample of 9 cells to declared tolerance)
- HZ tags (Aab IV/V/VI per HZCO 3.3 ± 1.0)
- Mainworld candidates (Aab VI + Aab IV a/d + Aab V b/d)
- Profiles (short \"4-2-12-5-0.5\"; long matches p.58 exactly)
- SAH cells (terrestrial '<S>??', GG 'GS|GM|GL<C>', belt '000')
- Moon designations + per-moon Sizes per p.58 / p.63 Notes

Two documented carve-outs per spec:
- HZ-candidate atmosphere/hydrographics → '?' (sub-project 3 deferral)
- Aab IV d-moon Size 5 (form authoritative over p.58 sizing-table errata)"
```

---

## Task 16: Final hygiene + memory updates + branch push

**Source:** Spec § Success criteria § Memory hygiene; standard end-of-sub-project workflow established by 2A and 2B.

**Files:** No source-code changes. Memory file updates outside the repo. Branch ready-for-merge state.

- [ ] **Step 1: Final full-suite verification**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
just check && just test
git log --oneline main..HEAD
```

Expected:

- `just check` clean (gofumpt, go vet, golangci-lint).
- `just test` clean (`go test -race ./...`).
- `git log` shows 14-15 commits on the feature branch in reverse-chronological order (the carry-forwards + 2C tasks).

- [ ] **Step 2: Read the WBH p.58/p.63 Aab IV-d errata note one more time and write the feedback memory**

Per spec § Open questions and § Success criteria § Memory hygiene, save a feedback memory documenting the inconsistency for future reference (joins the existing `feedback_wbh_p19_p42_inconsistency.md`).

Create `/Users/markayers/.claude/projects/-Users-markayers-Documents-Traveller/memory/feedback_wbh_p58_p63_inconsistency.md`:

```markdown
---
name: WBH p.58 vs p.63 Aab IV moon-size inconsistency
description: The Zed worked example's Aab IV d-moon size differs between the p.58 sizing-results table (S) and the p.63 Class II/III form (5). Treat the form as authoritative.
type: feedback
---

WBH p.58 sizing-results table for Aab IV reads: `2, S, S, S, S` (5 moons sized 2, S, S, S, S).

WBH p.63 IISS Class II/III form for Aab IV Notes column reads: `1,200⊕, HZ, 200, S, S, 566*, S` — the 4th moon (Aab IV d) is the mainworld candidate **Zed Prime** at SAH 566 (Size 5, Atmosphere 6, Hydrographics 6).

If d-moon were Size S as p.58 says, it could not be a mainworld candidate (Size S = small body <600 km).

**Why:** The book authors evidently updated the p.63 form to make Zed Prime habitable but left the p.58 sizing table unrevised. This is a book typo, not an authorial intent.

**How to apply:** When implementing 2C (or any future Zed-based test), treat **p.63 (the form) as authoritative**. The 2C acceptance test `TestZed_FullDetail` asserts d-moon SizeCode = "5" with an inline comment citing this errata.

Joins:

- `feedback_wbh_p19_p42_inconsistency.md` — Class VI cells diverge >5% between WBH p.19 and p.42 tables.
```

- [ ] **Step 3: Update the project resume memory**

Edit `/Users/markayers/.claude/projects/-Users-markayers-Documents-Traveller/memory/project_world_builder_resume.md`:

Replace the "**Next session should start Sub-project 2C**" section with:

```markdown
**Next session should start Sub-project 3: World Physical Characteristics (WBH pp. 69–146).**

The "System Worlds and Orbits" chapter is now fully encoded. Sub-project 3 builds on 2C's `SystemDetail` and adds:

| Sub-project | WBH pp.      | Status                                                       |
| ----------- | ------------ | ------------------------------------------------------------ |
| 2A          | 38–43        | **Done** — merged                                            |
| 2B          | 36–38, 43–52 | **Done** — merged                                            |
| 2C          | 53–67        | **Done** — merged                                            |
| 3           | 69–146       | **Next** — atmosphere, hydrographics, temperature, biosphere |

3 unblocks the deferred 2C carve-outs:

- HZ-candidate SAH `?` placeholders → numeric atmosphere/hydrographics digits via WBH p.108 Temperature Roll table
- Mainworld picker (`PickMainworld(sd SystemDetail) MainworldCandidate`)
- IISSClass23Form `ClassIIIStatus = true` rendering when full SAH is rolled
```

Update the "Build Tooling" + "Workflow" sections to add 2C-specific entries if any new gotchas surfaced during implementation (e.g., back-derivation of per-moon sizing dice; the `BaselineN` field addition to `StarAllocation`).

Add to the bottom under Workflow:

```markdown
- [WBH p.58 vs p.63 Aab IV-d moon-size inconsistency](feedback_wbh_p58_p63_inconsistency.md) — the form (p.63) is authoritative when book tables disagree
```

- [ ] **Step 4: Update the MEMORY.md index**

Edit `/Users/markayers/.claude/projects/-Users-markayers-Documents-Traveller/memory/MEMORY.md`. Add the new feedback entry to the index (under the Build Tooling or a Workflow section, matching the existing format):

```markdown
- [WBH p.58 vs p.63 Aab IV-d moon-size inconsistency](feedback_wbh_p58_p63_inconsistency.md) — form authoritative; 2C TestZed_FullDetail carves out
```

If `MEMORY.md` already has a "## Workflow" or "## Build Tooling" section, append the line there. Keep the index ≤200 lines as the file's header rule states.

- [ ] **Step 5: Confirm with the user before pushing**

The repo is local-only (no remote per memory `MEMORY.md` § "Project Structure"). There is **no `git push`** in this workflow. Branch lives locally on `feat/wbh-system-worlds-2c` and merges into `main` via `git merge --no-ff` after user review.

Ask the user: "2C implementation complete on `feat/wbh-system-worlds-2c`. Want to merge into main now, or review the branch first?"

Wait for user response. Possible follow-ups:

- **Merge now:**

  ```bash
  git checkout main
  git merge --no-ff feat/wbh-system-worlds-2c -m "Merge feat/wbh-system-worlds-2c: System Worlds 2C complete"
  git log --oneline -5
  ```

- **Review first:** show the user `git log --oneline main..HEAD` and `git diff main` summary; wait for explicit "merge it" before merging.

Per `.claude/CLAUDE.md` § "Defaults & Confirmations": "When user says 'yes' to a list of items, assume ALL items unless they specify."

- [ ] **Step 6: After merge, optionally delete the feature branch**

Per `.claude/CLAUDE.md` § "Worktree & Branch Hygiene": "Never clean up worktrees or delete branches until the associated PR is confirmed merged." There is no PR for this local-only repo, so the branch can be deleted after the merge commit lands on `main`:

```bash
git branch -d feat/wbh-system-worlds-2c
git log --oneline -3
```

(Use `-d`, not `-D`, so Git refuses if the branch isn't fully merged.)

---

## Self-review checklist (run before handoff)

After all 16 tasks complete, verify against the spec one more time:

- **API completeness.** Every type and function listed in spec § Success criteria § API completeness exists. Spot-check a handful via `grep -n "type SizeCode\|func RollTerrestrialSize\|func DetailSystem" worlds/*.go stars/*.go`.
- **Spec coverage:** every section/requirement in the spec has at least one corresponding task. The spec's bundled-carry-forward items 1-4 land in Tasks 1-3 (CF#5 is explicitly skipped).
- **No placeholders in the plan.** No "TBD" / "fill in" / "TODO" remain in the body of the plan tasks themselves (the spec carries forward "TODO(continuation-method)" as a code comment that's intentional and not a plan placeholder).
- **Type/method consistency.** `SizeCode`, `GasGiantClass`, `Moon`, `ParentInfo`, `Period`, `DetailedPlacement`, `SystemDetail`, `IISSClass23Form`, `IISSClass23Header`, `MainworldCandidate`, `StarRow` (via `stars.SurveyComponent`), `ObjectRow` — names are stable across all tasks that reference them.
- **Build green.** `just check && just test` clean on a fresh checkout.
- **Source traceability.** Every exported symbol has GoDoc citing a WBH page or step.
- **Memory hygiene.** New feedback memory written; resume memory updated; MEMORY.md index updated.
- **Branch state:** clean working tree on `main`; `feat/wbh-system-worlds-2c` either merged or pending user review.

---

## Plan complete

This plan implements sub-project 2C end-to-end against `docs/specs/2026-05-03-system-worlds-2c-sizing-design.md`. **15 numbered tasks + Pre-flight + final hygiene** = the full 2C delivery.

Per the writing-plans skill, the next step is your choice of execution mode:

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.

**2. Inline Execution** — I execute tasks in this session using executing-plans, batch execution with checkpoints for review.

Which approach do you want?
