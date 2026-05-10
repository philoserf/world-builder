# System Worlds 2A: Available Orbits + HZCO Implementation Plan (Go)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement sub-project 2A from the System Worlds and Orbits chapter: HZCO (single + composite), giant-primary cleanup carry-forward, and the 11-rule simplified `worlds.AvailableOrbits` method. Reproduce the Sol single-star case and the Zed quintuple's three available-orbits groups exactly.

**Architecture:** Two new functions in `wbh/stars` (`Star.HZCO()`, `CompositeHZCO`); a new package `wbh/worlds` housing types and the rule pipeline; a focused refactor of `stars/peculiar.go` + `stars/system.go` removing `ErrSpecialPrimaryClassRedirect` by adding an internal `generatePrimaryAtClass` helper that handles Class III/IV/VI primary generation.

**Tech Stack:** Same as Stars Plan 2. Go 1.22+, gofumpt CLI as canonical formatter (not golangci-lint's bundled gofumpt), golangci-lint v2.12.1, `just` recipes.

**Spec:** `docs/pass-1/specs/2026-05-02-system-worlds-2a-orbits-design.md`

**Source pages:** WBH pp. 38–43.

**Conventions:** Same as Stars Plan 2. Briefly:

- Working directory: `/Users/markayers/Documents/Traveller/`.
- TDD per task: write test → run-fail → implement → run-pass → format → lint → commit.
- `gofumpt -w` before commit. `gofumpt` CLI is the formatter source of truth (not golangci-lint).
- Test files live in the same package (white-box) except `worked_examples_test.go` (black-box `package worlds_test`).
- Tables for non-numeric cells: struct rows. Tables with nullable numeric cells: `*float64` via the existing `f` helper in `stars/tables.go`.
- Branch: create `feat/wbh-system-worlds-2a` off `main` before Task 1.

---

## Pre-flight

- [ ] **Verify clean state on main**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
git status
git switch main
just check && just test
```

Expected: clean working tree, all tests green.

- [ ] **Create feature branch**

```bash
git switch -c feat/wbh-system-worlds-2a
```

---

## File Structure

| File                              | Responsibility                                                                                    |
| --------------------------------- | ------------------------------------------------------------------------------------------------- |
| `stars/hzco.go`                   | `Star.HZCO()`, `CompositeHZCO()` — formula path. p. 42 table is test fixture only.                |
| `stars/hzco_test.go`              | p. 42 table fixture, worked-example HZCO assertions.                                              |
| `stars/system.go` (extend)        | Remove `ErrSpecialPrimaryClassRedirect`; route redirects via new helper.                          |
| `stars/peculiar.go` (extend)      | Add `generatePrimaryAtClass` helper for Class III/IV/VI primary generation.                       |
| `stars/peculiar_test.go` (extend) | Class redirect resolution tests.                                                                  |
| `stars/system_test.go` (extend)   | Giant-primary multi-star case.                                                                    |
| `worlds/available_orbits.go`      | Package types, MAO table & lookup, group identification, the 11-rule pipeline, `AvailableOrbits`. |
| `worlds/available_orbits_test.go` | Per-rule unit tests, MAO lookup tests, group-identification tests, `Total`/`Contains` tests.      |
| `worlds/worked_examples_test.go`  | Sol single-star + Zed quintuple acceptance gates (black-box `package worlds_test`).               |

---

## Task 1: Single-star HZCO formula

**Source:** WBH p. 41.

**Files:** `stars/hzco.go` (create), `stars/hzco_test.go` (create).

**API:**

```go
// HZCO returns the Habitable Zone Centre Orbit# for a single star,
// computed from its luminosity by the WBH p. 41 formula:
//
//	HZCO_AU    = sqrt(luminosity)
//	HZCO_Orbit = AUToOrbit(HZCO_AU)
func (s Star) HZCO() float64
```

- [ ] **Step 1: Write failing test**

Create `stars/hzco_test.go`:

```go
package stars

import (
	"math"
	"testing"
)

func TestStar_HZCO_Sol(t *testing.T) {
	t.Parallel()

	sol := Compose(ComposeOpts{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: V,
		Mass:            1.000, Diameter: 1.000, Temperature: 5772, AgeGyr: 4.568,
	})
	got := sol.HZCO()
	if math.Abs(got-3.0) > 0.05 {
		t.Errorf("Sol HZCO = %.4f, want 3.0±0.05", got)
	}
}

func TestStar_HZCO_ZedB(t *testing.T) {
	t.Parallel()

	// Zed B: K8 V, L=0.136 → HZCO ≈ 0.92 (WBH p. 43).
	b := Star{Luminosity: 0.136}
	got := b.HZCO()
	if math.Abs(got-0.92) > 0.05 {
		t.Errorf("Zed B HZCO = %.4f, want 0.92±0.05", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./stars/ -run TestStar_HZCO -v
```

Expected: FAIL — `s.HZCO undefined`.

- [ ] **Step 3: Implement HZCO**

Create `stars/hzco.go`:

```go
package stars

import "math"

// HZCO returns the Habitable Zone Centre Orbit# for a single star,
// computed from its luminosity by the WBH p. 41 formula:
//
//	HZCO_AU    = sqrt(luminosity)
//	HZCO_Orbit = AUToOrbit(HZCO_AU)
//
// The p. 42 HZCO table is encoded as a test fixture only; this function
// uses the formula path which the book itself validates as the canonical
// computation.
func (s Star) HZCO() float64 {
	return AUToOrbit(math.Sqrt(s.Luminosity))
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test ./stars/ -run TestStar_HZCO -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w stars/hzco.go stars/hzco_test.go
just check && just test
git add stars/hzco.go stars/hzco_test.go
git commit -m "feat(stars): Star.HZCO() formula path (WBH p.41)"
```

---

## Task 2: Composite HZCO for circumbinary pairs

**Source:** WBH p. 42.

**Files:** `stars/hzco.go` (extend), `stars/hzco_test.go` (extend).

**API:**

```go
// CompositeHZCO returns the HZCO# for a circumbinary group of stars
// orbiting a shared barycentre. Per WBH p. 42, the luminosities of all
// stars interior to the planet's orbit are summed, then the formula
// applies to the combined luminosity.
//
// For an Aab pair (primary + companion), pass [Aa, Ab]. Empty input
// returns 0.
func CompositeHZCO(starsInterior ...Star) float64
```

- [ ] **Step 1: Write failing tests**

Append to `stars/hzco_test.go`:

```go
func TestCompositeHZCO_ZedAab(t *testing.T) {
	t.Parallel()

	// Zed Aab: combined luminosity 1.419 → HZCO ≈ 3.3 (WBH p. 42).
	aa := Star{Luminosity: 0.738}
	ab := Star{Luminosity: 0.681}
	got := CompositeHZCO(aa, ab)
	if math.Abs(got-3.3) > 0.05 {
		t.Errorf("Zed Aab CompositeHZCO = %.4f, want 3.3±0.05", got)
	}
}

func TestCompositeHZCO_ZedCab(t *testing.T) {
	t.Parallel()

	// Zed Cab: combined luminosity 0.0896 → HZCO ≈ 0.75 (WBH p. 43).
	ca := Star{Luminosity: 0.0895}
	cb := Star{Luminosity: 0.000525}
	got := CompositeHZCO(ca, cb)
	if math.Abs(got-0.75) > 0.05 {
		t.Errorf("Zed Cab CompositeHZCO = %.4f, want 0.75±0.05", got)
	}
}

func TestCompositeHZCO_CorellaAab(t *testing.T) {
	t.Parallel()

	// Corella Aab: combined luminosity 1.725 → HZCO ≈ 3.5 (WBH p. 62).
	a := Star{Luminosity: 1.045}
	b := Star{Luminosity: 0.681}
	got := CompositeHZCO(a, b)
	if math.Abs(got-3.5) > 0.05 {
		t.Errorf("Corella Aab CompositeHZCO = %.4f, want 3.5±0.05", got)
	}
}

func TestCompositeHZCO_Empty(t *testing.T) {
	t.Parallel()

	if got := CompositeHZCO(); got != 0 {
		t.Errorf("CompositeHZCO() = %v, want 0", got)
	}
}
```

- [ ] **Step 2: Run tests to verify fail**

```bash
go test ./stars/ -run TestCompositeHZCO -v
```

Expected: FAIL — `CompositeHZCO undefined`.

- [ ] **Step 3: Implement CompositeHZCO**

Append to `stars/hzco.go`:

```go
// CompositeHZCO returns the HZCO# for a circumbinary group of stars
// orbiting a shared barycentre. Per WBH p. 42, the luminosities of all
// stars interior to the planet's orbit are summed, then the formula
// applies to the combined luminosity.
//
// Empty input returns 0.
func CompositeHZCO(starsInterior ...Star) float64 {
	var totalL float64
	for _, s := range starsInterior {
		totalL += s.Luminosity
	}
	if totalL <= 0 {
		return 0
	}
	return AUToOrbit(math.Sqrt(totalL))
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test ./stars/ -run "TestStar_HZCO|TestCompositeHZCO" -v
```

Expected: PASS for all five HZCO tests.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w stars/hzco.go stars/hzco_test.go
just check && just test
git add stars/hzco.go stars/hzco_test.go
git commit -m "feat(stars): CompositeHZCO for circumbinary pairs (WBH p.42)"
```

---

## Task 3: HZCO table fixture verification

**Source:** WBH p. 42 — the 14-row × 7-column HZCO# table.

**Files:** `stars/hzco_test.go` (extend).

**Goal:** Encode the published p. 42 HZCO table as a test fixture and verify the formula reproduces every populated cell within ±5% (the book's "validating the close enough approach" tolerance).

**Strategy:** The book's table is keyed by (spectral type × luminosity class), giving an HZCO# per cell. To verify the formula, we need the _luminosity_ for each cell — which lives in the existing `StarLuminosity` table from Stars Plan 1 (`stars/tables.go`). Compute `Star{Luminosity: lookupLuminosity(type, class)}.HZCO()` and compare to the table cell.

- [ ] **Step 1: Verify the existing luminosity-lookup helpers**

```bash
grep -n "func ComputeLuminosity\|StarLuminosity" stars/*.go | head
```

Expected: `ComputeLuminosity(SpectralType, LuminosityClass) (float64, error)` exists. If not, the test will need to call into the table directly — adjust accordingly.

- [ ] **Step 2: Write the property test**

Append to `stars/hzco_test.go`:

```go
// hzcoTablePage42 is the WBH p. 42 HZCO# table, used as a verification
// fixture for the formula-based Star.HZCO(). Cells with no value (the
// book's "—") are omitted from the map.
//
// Outer key: spectral type ("O0", "B5", "G2", ...). Inner map: luminosity
// class to expected HZCO# value.
var hzcoTablePage42 = map[string]map[LuminosityClass]float64{
	"O0": {Ia: 14.5, Ib: 14.4, II: 14.3, III: 14.3, V: 14.2, VI: 7.3},
	"O5": {Ia: 13.7, Ib: 13.5, II: 13.4, III: 13.2, V: 12.9, VI: 6.7},
	"B0": {Ia: 12.8, Ib: 12.2, II: 12.0, III: 11.7, IV: 11.4, V: 11.2, VI: 6.0},
	"B5": {Ia: 12.3, Ib: 11.1, II: 10.2, III: 9.0, IV: 8.6, V: 8.2, VI: 5.2},
	"A0": {Ia: 12.2, Ib: 10.9, II: 10.2, III: 7.5, IV: 7.2, V: 6.3},
	"A5": {Ia: 12.1, Ib: 10.8, II: 10.1, III: 6.9, IV: 6.1, V: 5.5},
	"F0": {Ia: 12.1, Ib: 10.8, II: 10.1, III: 6.7, IV: 5.9, V: 5.0},
	"F5": {Ia: 12.1, Ib: 10.8, II: 10.1, III: 6.2, IV: 4.7, V: 4.2},
	"G0": {Ia: 12.1, Ib: 10.8, II: 10.1, III: 7.1, IV: 5.2, V: 3.3},
	"G5": {Ia: 12.1, Ib: 10.8, II: 10.1, III: 7.4, IV: 5.4, V: 2.6, VI: 2.5},
	"K0": {Ia: 12.1, Ib: 10.8, II: 10.2, III: 7.6, IV: 5.8, V: 2.1, VI: 1.9},
	"K5": {Ia: 12.1, Ib: 10.9, II: 10.2, III: 8.1, V: 1.2, VI: 1.3},
	"M0": {Ia: 12.2, Ib: 11.0, II: 10.2, III: 8.2, V: 0.72, VI: 0.40},
	"M5": {Ia: 12.1, Ib: 11.1, II: 10.2, III: 8.4, V: 0.13, VI: 0.07},
	"M9": {Ia: 12.0, Ib: 10.8, II: 10.1, III: 8.8, V: 0.04, VI: 0.03},
}

func TestStar_HZCO_TableFidelity(t *testing.T) {
	t.Parallel()

	const tolerance = 0.05 // ±5%

	for typeStr, row := range hzcoTablePage42 {
		st, err := ParseSpectralType(typeStr)
		if err != nil {
			t.Fatalf("ParseSpectralType(%q): %v", typeStr, err)
		}
		for lc, want := range row {
			lum, err := ComputeLuminosity(st, lc)
			if err != nil {
				continue // book-blank cells already excluded above; defensive
			}
			s := Star{
				SpectralType:    st,
				LuminosityClass: lc,
				Luminosity:      lum,
			}
			got := s.HZCO()
			rel := math.Abs(got-want) / want
			if rel > tolerance {
				t.Errorf("HZCO(%s %s) = %.4f, want %.4f (rel err %.3f > %.3f)",
					typeStr, lc, got, want, rel, tolerance)
			}
		}
	}
}
```

- [ ] **Step 3: Run test**

```bash
go test ./stars/ -run TestStar_HZCO_TableFidelity -v
```

Expected: PASS, or output cells where the formula diverges by >5% (which would surface a bug worth investigating). If any cell fails, do not relax the tolerance without first checking that the luminosity in `ComputeLuminosity` matches the book's underlying p. 19 luminosity table for that cell.

- [ ] **Step 4: Format, commit**

```bash
gofumpt -w stars/hzco_test.go
just check && just test
git add stars/hzco_test.go
git commit -m "test(stars): HZCO formula verified against WBH p.42 table"
```

---

## Task 4: `generatePrimaryAtClass` helper for Class III/IV/VI

**Source:** WBH p. 15 (Star Type Determination), p. 17 (Star Mass and Temperature by Class).

**Files:** `stars/peculiar.go` (extend), `stars/peculiar_test.go` (extend).

**Goal:** Internal helper that takes a target `LuminosityClass` (III, IV, VI, Ia, Ib, II) and rolls a complete primary at that class using the existing Type/Subtype/physical-quantity machinery. Used by `generateSpecialPrimary` to resolve Special-column class redirects.

**API (unexported):**

```go
// generatePrimaryAtClass rolls a complete primary star at a specified
// luminosity class. Used when the Special-column primary roll redirects
// to a class (e.g., "Class III"), and to generate giant primaries.
//
// Roll order (consumed from the roller in this order):
//  1. 2D for Type column (no class re-roll — class is the input)
//  2. 2D for Star Subtype, with Class IV / Class VI / giant restrictions
//     applied as RollSubtype already does.
//  3. (if WithVariance) 2D-7 for mass variance
//  4. (if WithVariance) 2D-7 for diameter variance
//  5. age rolls per Accuracy (small-star or large-star path depending
//     on the class)
//
// Returns ErrSpecialPrimaryClassRedirect-style class redirects via a
// recursive call only if the regular Type table itself emits a "Special"
// or "Peculiar" cell at the indicated class — bounded at depth 5.
func generatePrimaryAtClass(r roller.Roller, targetClass LuminosityClass, opts GenerateOpts) (Star, error)
```

- [ ] **Step 1: Verify existing helpers**

```bash
grep -n "func RollSubtype\|func ComputeMass\|func ComputeDiameter\|func ComputeTemperature\|func SmallStarAge\|func ApplyVariance" stars/*.go | head
```

Expected: All those helpers exist (from Stars Plans 1 and 2). Note: `LargeStarAge` does **not** exist; this plan uses `SmallStarAge` for all classes and documents giant-age-modeling as a future concern. The Special Circumstances chapter (and a future giant-age task) will introduce a class-aware age model.

- [ ] **Step 2: Write failing test for Class III generation**

Append to `stars/peculiar_test.go`:

```go
func TestGeneratePrimaryAtClass_III(t *testing.T) {
	t.Parallel()

	// Class III primary: 2D=7 for Type → "K"; 2D=7 for Subtype
	// (Numeric column row 7 = subtype 4 per Plan 1's Subtype table);
	// no variance; small-star age path doesn't apply for Class III, but
	// LargeStarAge does. Use a deterministic scripted roller.
	//
	// Adjust expected subtype based on the actual Subtype table row for
	// 2D=7 in the Numeric column; verify by reading stars/tables.go.
	rolls := []int{7, 7, 1, 2}
	r := roller.NewScripted(rolls...)
	got, err := generatePrimaryAtClass(r, III, GenerateOpts{Accuracy: 1})
	if err != nil {
		t.Fatalf("generatePrimaryAtClass: %v", err)
	}
	if got.LuminosityClass != III {
		t.Errorf("LuminosityClass = %s, want III", got.LuminosityClass)
	}
	if got.SpectralType.Letter != 'K' {
		t.Errorf("Letter = %c, want K", got.SpectralType.Letter)
	}
	if got.Mass <= 0 {
		t.Errorf("Mass = %v, want > 0", got.Mass)
	}
}
```

> **Note:** The exact roll sequence above assumes Numeric column row 7 maps to subtype 4 and that LargeStarAge consumes 2 rolls at Accuracy=1. **Before running**, read `stars/tables.go` for the Subtype table and `stars/ages.go` for the LargeStarAge signature, then adjust the `rolls` slice and the assertion to match. The point of the test is "given known rolls at Class III, we get a fully-resolved K-class III star with positive mass" — not the specific subtype.

- [ ] **Step 3: Run test to verify fail**

```bash
go test ./stars/ -run TestGeneratePrimaryAtClass_III -v
```

Expected: FAIL — `generatePrimaryAtClass undefined`.

- [ ] **Step 4: Implement generatePrimaryAtClass**

Append to `stars/peculiar.go`:

```go
// generatePrimaryAtClass rolls a complete primary star at a specified
// luminosity class (used for Special-column class redirects and giant
// primary generation).
//
// The Type column roll is consumed but the class is overridden to
// targetClass, since the caller already chose the class.
func generatePrimaryAtClass(r roller.Roller, targetClass LuminosityClass, opts GenerateOpts) (Star, error) {
	letter, _, err := RollPrimaryTypeAndClass(r)
	if err != nil {
		return Star{}, err
	}
	subtype, err := RollSubtype(r, letter, targetClass)
	if err != nil {
		return Star{}, err
	}
	st := SpectralType{Letter: letter, Subtype: subtype}

	mass, err := ComputeMass(st, targetClass)
	if err != nil {
		return Star{}, err
	}
	diameter, err := ComputeDiameter(st, targetClass)
	if err != nil {
		return Star{}, err
	}
	temperature, err := ComputeTemperature(st)
	if err != nil {
		return Star{}, err
	}

	if opts.WithVariance {
		mass = ApplyVariance(mass, r, 0.20)
		diameter = ApplyVariance(diameter, r, 0.20)
	}

	luminosity := ComputeLuminosityFromFormula(diameter, temperature)

	// NOTE: Plan 2 only encoded SmallStarAge. WBH has separate
	// Main-sequence / Subgiant / Giant / Final-age formulas (p. 21)
	// that giants should use. Encoding those is deferred; for 2A we
	// reuse SmallStarAge for all classes. The acceptance gate in
	// Task 5 only checks that the redirect resolves to a Class III
	// primary, not the age value.
	age, err := SmallStarAge(r, opts.Accuracy)
	if err != nil {
		return Star{}, err
	}

	return Star{
		Kind:            KindMainSequence,
		SpectralType:    st,
		LuminosityClass: targetClass,
		Mass:            mass,
		Diameter:        diameter,
		Temperature:     temperature,
		Luminosity:      luminosity,
		AgeGyr:          age,
	}, nil
}
```

- [ ] **Step 5: Run test to verify pass**

```bash
go test ./stars/ -run TestGeneratePrimaryAtClass -v
```

Expected: PASS. If FAIL on subtype mismatch, adjust the test's roll sequence and assertions to match what the existing tables actually produce; the point is the _shape_ of the result, not specific subtype values.

- [ ] **Step 6: Format, lint, commit**

```bash
gofumpt -w stars/peculiar.go stars/peculiar_test.go
just check && just test
git add stars/peculiar.go stars/peculiar_test.go
git commit -m "feat(stars): generatePrimaryAtClass helper for class redirects (WBH p.15)"
```

---

## Task 5: Wire class redirect through `generateSpecialPrimary`, remove `ErrSpecialPrimaryClassRedirect`

**Source:** WBH p. 15.

**Files:** `stars/system.go` (modify), `stars/system_test.go` (extend).

**Current code** (`stars/system.go` ~line 262):

```go
func generateSpecialPrimary(r roller.Roller) (Star, error) {
	kind, lc, err := RollSpecialPrimary(r, PeculiarPathUnusual)
	if err != nil {
		return Star{}, err
	}
	if lc != "" {
		return Star{}, fmt.Errorf("%w: class %s", ErrSpecialPrimaryClassRedirect, lc)
	}
	// ... happy path with kind
}
```

**Goal:** Replace the error-emitting redirect branch with a call to `generatePrimaryAtClass`. Delete the error sentinel.

- [ ] **Step 1: Write failing test for redirect resolution**

Append to `stars/system_test.go`:

```go
func TestGenerateSystem_SpecialPrimaryClassRedirect(t *testing.T) {
	t.Parallel()

	// Drive a Special primary roll that redirects to Class III:
	//  Roll 1: 2D=2 for the Type column (Special)
	//  Roll 2: 2D=N where the Unusual column for that 2D = "Class III"
	//      (read stars/tables.go's StarTypeDetermination Unusual column;
	//      a "Class III" cell appears at row 9 in many configurations —
	//      adjust to the actual row that produces "Class III").
	//  Rolls 3+: per generatePrimaryAtClass for Class III at the
	//      configured opts (Accuracy 1, no variance).
	//
	// Then verify the resulting primary has LuminosityClass III.
	rolls := []int{
		2,    // Special primary
		9,    // Unusual column → "Class III" (verify cell)
		7, 7, // Type letter + Subtype for Class III
		1, 2, // age rolls (LargeStarAge at Accuracy 1)
	}
	r := roller.NewScripted(rolls...)
	sys, err := GenerateSystem(r, GenerateSystemOpts{Accuracy: 1})
	if err != nil {
		t.Fatalf("GenerateSystem: %v", err)
	}
	if sys.Primary.LuminosityClass != III {
		t.Errorf("primary class = %s, want III", sys.Primary.LuminosityClass)
	}
}
```

> **Adjust the roll sequence** to match the actual `StarTypeDetermination` Unusual column. Read `stars/tables.go` for the cell mapping; pick a row where the cell is `"Class III"`. The principle of the test is: "Special → Class III redirect resolves to a Class III primary without error."

- [ ] **Step 2: Run test to verify fail**

```bash
go test ./stars/ -run TestGenerateSystem_SpecialPrimaryClassRedirect -v
```

Expected: FAIL — error returned matches `ErrSpecialPrimaryClassRedirect`.

- [ ] **Step 3: Modify `generateSpecialPrimary` and remove the sentinel**

Replace the `generateSpecialPrimary` body in `stars/system.go`:

```go
func generateSpecialPrimary(r roller.Roller, opts GenerateSystemOpts) (Star, error) {
	kind, lc, err := RollSpecialPrimary(r, PeculiarPathUnusual)
	if err != nil {
		return Star{}, err
	}
	if lc != "" {
		// Class redirect: re-roll on the regular Star Type Determination
		// flow at the indicated class.
		return generatePrimaryAtClass(r, lc, GenerateOpts{
			WithVariance: opts.WithVariance,
			Accuracy:     opts.Accuracy,
		})
	}
	// ... happy path with kind (unchanged)
}
```

Update the call site in `GenerateSystem` to pass `opts`:

```go
// Before:
primary, err = generateSpecialPrimary(r)
// After:
primary, err = generateSpecialPrimary(r, opts)
```

Delete the `ErrSpecialPrimaryClassRedirect` declaration (around line 245-252) and the `import "errors"` if it becomes unused.

- [ ] **Step 4: Run all stars tests to verify nothing broke**

```bash
go test ./stars/ -v
```

Expected: PASS — including the new redirect test and all pre-existing tests. If any pre-existing test fails referencing `ErrSpecialPrimaryClassRedirect`, update it: the redirect should now succeed instead of erroring.

- [ ] **Step 5: Verify no callers of the removed sentinel remain**

```bash
grep -rn "ErrSpecialPrimaryClassRedirect" .
```

Expected: zero matches.

- [ ] **Step 6: Format, lint, commit**

```bash
gofumpt -w stars/system.go stars/system_test.go
just check && just test
git add stars/system.go stars/system_test.go
git commit -m "feat(stars): resolve Special-primary class redirects (remove ErrSpecialPrimaryClassRedirect)"
```

---

## Task 6: `worlds/` package scaffolding

**Source:** Spec § Public API.

**Files:** `worlds/available_orbits.go` (create), `worlds/available_orbits_test.go` (create).

**Goal:** Package skeleton with the public types and `Group.Total()` / `Group.Contains()`. No `AvailableOrbits` yet; just the data shapes and helpers, fully tested.

- [ ] **Step 1: Write failing tests**

Create `worlds/available_orbits_test.go`:

```go
package worlds

import (
	"math"
	"testing"
)

func TestGroup_Total_SingleInterval(t *testing.T) {
	t.Parallel()

	g := Group{
		Designation: "A",
		MAO:         0.03,
		Intervals:   []Interval{{Min: 0.03, Max: 20.0}},
	}
	got := g.Total()
	want := 19.97
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("Total = %v, want %v", got, want)
	}
}

func TestGroup_Total_MultiInterval(t *testing.T) {
	t.Parallel()

	// Zed Aab from WBH p. 40: 0.61–5.10, 7.10–10.10, 14.10–20.00 → 13.39.
	g := Group{
		Designation: "Aab",
		MAO:         0.61,
		Intervals: []Interval{
			{Min: 0.61, Max: 5.10},
			{Min: 7.10, Max: 10.10},
			{Min: 14.10, Max: 20.00},
		},
	}
	got := g.Total()
	want := 13.39
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("Total = %v, want %v", got, want)
	}
}

func TestGroup_Total_Empty(t *testing.T) {
	t.Parallel()

	g := Group{Intervals: nil}
	if got := g.Total(); got != 0 {
		t.Errorf("Total = %v, want 0", got)
	}
}

func TestGroup_Contains(t *testing.T) {
	t.Parallel()

	g := Group{
		Intervals: []Interval{
			{Min: 0.61, Max: 5.10},
			{Min: 7.10, Max: 10.10},
		},
	}

	tests := []struct {
		orbit float64
		want  bool
	}{
		{0.5, false},
		{0.61, true},
		{3.0, true},
		{5.10, true},
		{6.0, false},
		{7.0, false},
		{7.10, true},
		{10.10, true},
		{15.0, false},
	}
	for _, tc := range tests {
		if got := g.Contains(tc.orbit); got != tc.want {
			t.Errorf("Contains(%v) = %v, want %v", tc.orbit, got, tc.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify fail**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
mkdir -p worlds
go test ./worlds/ -v
```

Expected: package compile failure — `Group`, `Interval` undefined.

- [ ] **Step 3: Create the package**

Create `worlds/available_orbits.go`:

```go
// Package worlds implements WBH world-placement procedures atop
// stars.System.
//
// Sub-project 2A covers Available Orbits and the Habitable Zone
// Centre Orbit# (HZCO). HZCO lives in the stars package as a stellar
// property; available-orbits computation lives here as a system-level
// constraint over those stars.
//
// Source: WBH pp. 38–43 (System Worlds and Orbits chapter).
package worlds

import (
	"errors"

	"wbh/stars"
)

// Interval is a closed Orbit# range [Min, Max].
type Interval struct {
	Min, Max float64
}

// Group is one body or barycentric pair sharing an orbit set.
//
// Single-star group: Members has one Star, Designation is "A"/"B"/"C"/"D".
// Pair group:        Members has two Stars (parent first, companion
//                    second), Designation is "Aab"/"Cab"/...
type Group struct {
	Designation string
	Members     []stars.Star
	MAO         float64    // p. 39 table; for pairs, raised by rule 2 if applicable
	Intervals   []Interval // disjoint, sorted ascending

	// companionEcc records the companion's eccentricity for pair
	// groups. Set by identifyGroups; read by AvailableOrbits's rule 2
	// pass. Unexported because it's an implementation detail of the
	// rule pipeline, not part of the public API.
	companionEcc float64
}

// Total returns the sum of (Max - Min) over all intervals — the value
// the book calls "total Orbit#s" used in placement allocation
// (sub-project 2B).
func (g Group) Total() float64 {
	var t float64
	for _, iv := range g.Intervals {
		t += iv.Max - iv.Min
	}
	return t
}

// Contains reports whether orbit is inside any of g.Intervals.
// Endpoints count as inside.
func (g Group) Contains(orbit float64) bool {
	for _, iv := range g.Intervals {
		if orbit >= iv.Min && orbit <= iv.Max {
			return true
		}
	}
	return false
}

// Result is the per-group available orbits for an entire system.
type Result struct {
	Groups []Group // ordered by ascending stellar Orbit# of the group's outer member
}

// ErrPostStellarPrimaryUnsupported indicates the primary star is a
// Brown Dwarf, White Dwarf, Neutron Star, Black Hole, or Pulsar —
// classes whose MAO is in the Special Circumstances chapter and not
// yet encoded.
var ErrPostStellarPrimaryUnsupported = errors.New(
	"worlds: post-stellar primary MAO requires Special Circumstances chapter",
)
```

- [ ] **Step 4: Verify go.mod doesn't need updating**

```bash
go test ./worlds/ -v
```

Expected: PASS for all four scaffold tests. If it complains about an unknown module path, ensure `wbh` is the module name in `go.mod` (it should be from earlier plans).

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/available_orbits.go worlds/available_orbits_test.go
just check && just test
git add worlds/available_orbits.go worlds/available_orbits_test.go
git commit -m "feat(worlds): package scaffolding — Interval, Group, Result types (WBH p.38)"
```

---

## Task 7: MAO table from p. 39 + lookup function

**Source:** WBH p. 39.

**Files:** `worlds/available_orbits.go` (extend), `worlds/available_orbits_test.go` (extend).

**Table:** 15 rows (O0 to M9 at standard subtype steps) × 7 columns (Ia, Ib, II, III, IV, V, VI). Cells with the book's "—" mean "this combination does not exist as a star" — represent as `nil` via `*float64`.

**API:**

```go
// MAO returns the Minimum Allowable Orbit# for a star, looked up by
// spectral type (interpolated within a luminosity-class column) per the
// WBH p. 39 table. Returns ErrNoMAOForStar if the cell is the book's
// "—" (e.g., A0 VI does not exist).
//
// Post-stellar kinds (Brown Dwarf, White Dwarf, Neutron Star, Black
// Hole, Pulsar) return ErrPostStellarPrimaryUnsupported per the spec's
// scope limitation.
func MAO(s stars.Star) (float64, error)

var ErrNoMAOForStar = errors.New("worlds: spectral type / class combination has no MAO entry")
```

- [ ] **Step 1: Write failing tests**

Append to `worlds/available_orbits_test.go`:

```go
import (
	"errors"
	// ... existing imports
	"wbh/stars"
)

func TestMAO_ZedAa(t *testing.T) {
	t.Parallel()

	// G7 V should interpolate between G5 V (0.02) and K0 V (0.02);
	// expected MAO 0.02 (book uses 0.03 for Sol G2 V — different cell).
	zedAa := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass: 0.929, Diameter: 0.967, Temperature: 5440,
	})
	got, err := MAO(zedAa)
	if err != nil {
		t.Fatalf("MAO: %v", err)
	}
	if math.Abs(got-0.02) > 1e-9 {
		t.Errorf("MAO(G7 V) = %v, want 0.02", got)
	}
}

func TestMAO_ZedB(t *testing.T) {
	t.Parallel()

	// K8 V interpolates K5 V (0.02) → M0 V (0.02); expected 0.02.
	zedB := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass: 0.626, Diameter: 0.777, Temperature: 3980,
	})
	got, err := MAO(zedB)
	if err != nil {
		t.Fatalf("MAO: %v", err)
	}
	if math.Abs(got-0.02) > 1e-9 {
		t.Errorf("MAO(K8 V) = %v, want 0.02", got)
	}
}

func TestMAO_Sol(t *testing.T) {
	t.Parallel()

	// G2 V: G0 V (0.03) → G5 V (0.02); 2/5 between → 0.03 - (0.01 × 2/5) = 0.026.
	// The Sol acceptance test requires this to be 0.03 to match the book's
	// rounded statement; pick interpolation strategy to match the book.
	sol := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass: 1.000, Diameter: 1.000, Temperature: 5772,
	})
	got, err := MAO(sol)
	if err != nil {
		t.Fatalf("MAO: %v", err)
	}
	// Book reports 0.03 for Sol on the worked example survey form.
	// Allow small interpolation variance.
	if math.Abs(got-0.03) > 0.005 {
		t.Errorf("MAO(G2 V) = %v, want ~0.03", got)
	}
}

func TestMAO_NoEntry(t *testing.T) {
	t.Parallel()

	// A0 VI is "—" in the book — no entry.
	a0vi := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'A', Subtype: 0},
		LuminosityClass: stars.VI,
		Mass: 0.5, Diameter: 0.5, Temperature: 9000,
	})
	_, err := MAO(a0vi)
	if !errors.Is(err, ErrNoMAOForStar) {
		t.Errorf("MAO(A0 VI) error = %v, want ErrNoMAOForStar", err)
	}
}

func TestMAO_PostStellar(t *testing.T) {
	t.Parallel()

	bd := stars.Star{Kind: stars.KindBrownDwarf}
	_, err := MAO(bd)
	if !errors.Is(err, ErrPostStellarPrimaryUnsupported) {
		t.Errorf("MAO(BD) error = %v, want ErrPostStellarPrimaryUnsupported", err)
	}
}
```

> **Note:** the test values for G7 V / K8 V / G2 V come directly from reading the book's p. 39 table. If the test reveals the existing project's `*float64` helper isn't available in `worlds/`, define a small private `f := func(x float64) *float64 { return &x }` inside `available_orbits.go`.

- [ ] **Step 2: Run tests to verify fail**

```bash
go test ./worlds/ -run TestMAO -v
```

Expected: FAIL — `MAO undefined`, `ErrNoMAOForStar undefined`.

- [ ] **Step 3: Implement MAO table and lookup**

Append to `worlds/available_orbits.go`:

```go
import "fmt"

// f returns a pointer to the float64 — used for nil-cells in tables.
func f(x float64) *float64 { return &x }

// maoRow is one row of the WBH p. 39 Minimum Allowable Orbit# table,
// keyed by luminosity class.
type maoRow struct {
	Ia, Ib, II, III, IV, V, VI *float64
}

// maoTablePage39 is the WBH p. 39 MAO table, keyed by spectral-type
// short code ("O0", "B5", "G7", ...).
//
// nil pointer means the book leaves the cell as "—" (combination does
// not exist as a star).
var maoTablePage39 = map[string]maoRow{
	"O0": {Ia: f(0.63), Ib: f(0.60), II: f(0.55), III: f(0.53), V: f(0.5), VI: f(0.01)},
	"O5": {Ia: f(0.55), Ib: f(0.50), II: f(0.45), III: f(0.38), V: f(0.3), VI: f(0.01)},
	"B0": {Ia: f(0.50), Ib: f(0.35), II: f(0.30), III: f(0.25), IV: f(0.20), V: f(0.18), VI: f(0.01)},
	"B5": {Ia: f(1.67), Ib: f(0.63), II: f(0.35), III: f(0.15), IV: f(0.13), V: f(0.09), VI: f(0.01)},
	"A0": {Ia: f(3.34), Ib: f(1.40), II: f(0.75), III: f(0.13), IV: f(0.10), V: f(0.06)},
	"A5": {Ia: f(4.17), Ib: f(2.17), II: f(1.17), III: f(0.13), IV: f(0.07), V: f(0.05)},
	"F0": {Ia: f(4.42), Ib: f(2.50), II: f(1.33), III: f(0.13), IV: f(0.07), V: f(0.04)},
	"F5": {Ia: f(5.00), Ib: f(3.25), II: f(1.87), III: f(0.13), IV: f(0.06), V: f(0.03)},
	"G0": {Ia: f(5.21), Ib: f(3.59), II: f(2.24), III: f(0.25), IV: f(0.07), V: f(0.03), VI: f(0.02)},
	"G5": {Ia: f(5.34), Ib: f(3.84), II: f(2.67), III: f(0.38), IV: f(0.10), V: f(0.02), VI: f(0.02)},
	"K0": {Ia: f(5.59), Ib: f(4.17), II: f(3.17), III: f(0.50), IV: f(0.15), V: f(0.02), VI: f(0.02)},
	"K5": {Ia: f(6.17), Ib: f(4.84), II: f(4.00), III: f(1.00), V: f(0.02), VI: f(0.01)},
	"M0": {Ia: f(6.80), Ib: f(5.42), II: f(4.59), III: f(1.68), V: f(0.02), VI: f(0.01)},
	"M5": {Ia: f(7.20), Ib: f(6.17), II: f(5.30), III: f(3.00), V: f(0.01), VI: f(0.01)},
	"M9": {Ia: f(7.80), Ib: f(6.59), II: f(5.92), III: f(4.34), V: f(0.01), VI: f(0.01)},
}

// ErrNoMAOForStar reports a "—" cell in the p. 39 MAO table — the
// spectral type / class combination does not exist as a star.
var ErrNoMAOForStar = errors.New("worlds: spectral type / class combination has no MAO entry")

// isPostStellar reports whether a StarKind is post-stellar (BD, D, NS,
// BH, Pulsar) — these have MAO defined in the Special Circumstances
// chapter, not yet encoded.
func isPostStellar(k stars.StarKind) bool {
	switch k {
	case stars.KindBrownDwarf, stars.KindWhiteDwarf,
		stars.KindNeutronStar, stars.KindBlackHole, stars.KindPulsar:
		return true
	}
	return false
}

// maoCell reads the MAO cell for an exact spectral type (e.g. "G5") at
// a given class. Returns ErrNoMAOForStar if the cell is the book's "—".
func maoCell(typeKey string, lc stars.LuminosityClass) (float64, error) {
	row, ok := maoTablePage39[typeKey]
	if !ok {
		return 0, fmt.Errorf("worlds: no MAO row for %q", typeKey)
	}
	var ptr *float64
	switch lc {
	case stars.Ia:
		ptr = row.Ia
	case stars.Ib:
		ptr = row.Ib
	case stars.II:
		ptr = row.II
	case stars.III:
		ptr = row.III
	case stars.IV:
		ptr = row.IV
	case stars.V:
		ptr = row.V
	case stars.VI:
		ptr = row.VI
	default:
		return 0, fmt.Errorf("worlds: unknown luminosity class %q", lc)
	}
	if ptr == nil {
		return 0, ErrNoMAOForStar
	}
	return *ptr, nil
}

// MAO returns the Minimum Allowable Orbit# for a star, interpolated by
// spectral type within its luminosity-class column per the WBH p. 39
// table.
//
// Post-stellar kinds return ErrPostStellarPrimaryUnsupported.
// Combinations the book lists as "—" return ErrNoMAOForStar.
func MAO(s stars.Star) (float64, error) {
	if isPostStellar(s.Kind) {
		return 0, ErrPostStellarPrimaryUnsupported
	}
	// Find the two grid points bracketing this subtype within the same
	// letter and interpolate linearly. Letters and grid points: O0, O5,
	// B0, B5, A0, A5, F0, F5, G0, G5, K0, K5, M0, M5, M9.
	lower, upper, frac := bracketSpectralType(s.SpectralType)
	lo, errLo := maoCell(lower, s.LuminosityClass)
	hi, errHi := maoCell(upper, s.LuminosityClass)
	switch {
	case errors.Is(errLo, ErrNoMAOForStar) && errors.Is(errHi, ErrNoMAOForStar):
		return 0, ErrNoMAOForStar
	case errors.Is(errLo, ErrNoMAOForStar):
		return hi, nil
	case errors.Is(errHi, ErrNoMAOForStar):
		return lo, nil
	case errLo != nil:
		return 0, errLo
	case errHi != nil:
		return 0, errHi
	}
	return lo + (hi-lo)*frac, nil
}

// bracketSpectralType returns the two p. 39 grid keys bracketing st
// within st's letter, and the fractional position from lower to upper.
// For exact grid hits (O0, O5, ...) lower == upper and frac is 0.
func bracketSpectralType(st stars.SpectralType) (lower, upper string, frac float64) {
	// Grid points within a letter are at subtypes 0, 5, and (for M) 9.
	letter := string(st.Letter)
	switch {
	case st.Letter == 'M' && st.Subtype >= 5:
		// Bracket M5 → M9; subtypes 5..9.
		lower = "M5"
		upper = "M9"
		if st.Subtype <= 5 {
			return "M5", "M5", 0
		}
		if st.Subtype >= 9 {
			return "M9", "M9", 0
		}
		return "M5", "M9", float64(st.Subtype-5) / 4.0
	case st.Subtype < 5:
		lower = letter + "0"
		upper = letter + "5"
		return lower, upper, float64(st.Subtype) / 5.0
	default:
		// st.Subtype in [5, 9] but letter != M (handled above).
		// Bracket Letter5 → NextLetter0.
		next := nextSpectralLetter(st.Letter)
		lower = letter + "5"
		upper = string(next) + "0"
		return lower, upper, float64(st.Subtype-5) / 5.0
	}
}

// nextSpectralLetter returns the next cooler spectral letter in O B A F G K M order.
func nextSpectralLetter(l stars.SpectralLetter) stars.SpectralLetter {
	switch l {
	case 'O':
		return 'B'
	case 'B':
		return 'A'
	case 'A':
		return 'F'
	case 'F':
		return 'G'
	case 'G':
		return 'K'
	case 'K':
		return 'M'
	default:
		return l
	}
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test ./worlds/ -run TestMAO -v
```

Expected: PASS for all five MAO tests.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/available_orbits.go worlds/available_orbits_test.go
just check && just test
git add worlds/available_orbits.go worlds/available_orbits_test.go
git commit -m "feat(worlds): MAO lookup with spectral interpolation (WBH p.39)"
```

---

## Task 8: Group identification from `stars.System`

**Source:** Spec § Group identification.

**Files:** `worlds/available_orbits.go` (extend), `worlds/available_orbits_test.go` (extend).

**API (unexported, used by AvailableOrbits):**

```go
// identifyGroups partitions a System into its barycentric orbit groups:
//   - Primary plus its Companion-class companion (if any) → "A" or "Aab"
//   - Each Close/Near/Far secondary plus its own Companion-class
//     companion (if any) → "B"/"Bab"/"C"/"Cab"/...
//
// Designations are positionally renumbered: with all of Close/Near/Far
// present they become B, C, D in order; if Close is absent, Near
// becomes B, etc.
//
// A CompanionStar with OrbitClass == Companion is folded into its
// parent's group.
func identifyGroups(sys stars.System) []Group
```

- [ ] **Step 1: Write failing tests**

Append to `worlds/available_orbits_test.go`:

```go
func TestIdentifyGroups_SinglePrimary(t *testing.T) {
	t.Parallel()

	sys := stars.System{
		Primary: stars.Compose(stars.ComposeOpts{
			Kind:            stars.KindMainSequence,
			SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
			LuminosityClass: stars.V,
			Mass:            1.0, Diameter: 1.0, Temperature: 5772,
		}),
		Companions: nil,
	}
	groups := identifyGroups(sys)
	if len(groups) != 1 {
		t.Fatalf("groups = %d, want 1", len(groups))
	}
	if groups[0].Designation != "A" {
		t.Errorf("Designation = %q, want \"A\"", groups[0].Designation)
	}
	if len(groups[0].Members) != 1 {
		t.Errorf("Members = %d, want 1", len(groups[0].Members))
	}
}

func TestIdentifyGroups_PrimaryWithCompanion(t *testing.T) {
	t.Parallel()

	primary := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass:            0.929, Diameter: 0.967, Temperature: 5440,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.907, Diameter: 0.957, Temperature: 5360,
	})
	sys := stars.System{
		Primary: primary,
		Companions: []stars.CompanionStar{
			{Star: ab, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.09, Eccentricity: 0.11, ParentIndex: -1},
		},
	}
	groups := identifyGroups(sys)
	if len(groups) != 1 {
		t.Fatalf("groups = %d, want 1 (companion folds into primary)", len(groups))
	}
	if groups[0].Designation != "Aab" {
		t.Errorf("Designation = %q, want \"Aab\"", groups[0].Designation)
	}
	if len(groups[0].Members) != 2 {
		t.Errorf("Members = %d, want 2", len(groups[0].Members))
	}
}

func TestIdentifyGroups_ZedQuintuple(t *testing.T) {
	t.Parallel()

	// Construct the Zed quintuple in pieces and verify three groups.
	// Aa + Ab → Aab; B alone → B; Ca + Cb → Cab.
	aa := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass: 0.929, Diameter: 0.967, Temperature: 5440,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V,
		Mass: 0.907, Diameter: 0.957, Temperature: 5360,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass: 0.626, Diameter: 0.777, Temperature: 3980,
	})
	ca := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'M', Subtype: 0},
		LuminosityClass: stars.V,
		Mass: 0.510, Diameter: 0.728, Temperature: 3700,
	})
	cb := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindWhiteDwarf,
		Mass: 0.490, Diameter: 0.017, Temperature: 6700,
	})
	sys := stars.System{
		Primary: aa,
		Companions: []stars.CompanionStar{
			// Index 0: Ab is companion of primary (ParentIndex == -1).
			{Star: ab, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.09, Eccentricity: 0.11, ParentIndex: -1},
			// Index 1: B is Near secondary of primary.
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.08, ParentIndex: -1},
			// Index 2: Ca is Far secondary of primary.
			{Star: ca, OrbitClass: stars.OrbitFar, OrbitNumber: 12.10, Eccentricity: 0.47, ParentIndex: -1},
			// Index 3: Cb is companion of Ca (ParentIndex == 2).
			{Star: cb, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.21, Eccentricity: 0.24, ParentIndex: 2},
		},
	}
	// Zed: Close is absent. Designations: primary group "Aab",
	// Near (only Close-relative slot left) becomes "B",
	// Far becomes "Cab" (with its own companion folded in).
	groups := identifyGroups(sys)
	if len(groups) != 3 {
		t.Fatalf("groups = %d, want 3", len(groups))
	}
	wantDesig := []string{"Aab", "B", "Cab"}
	for i, g := range groups {
		if g.Designation != wantDesig[i] {
			t.Errorf("groups[%d].Designation = %q, want %q", i, g.Designation, wantDesig[i])
		}
	}
}
```

> **Companion-pairing mechanism:** `CompanionStar` has a `ParentIndex int` field. `ParentIndex == -1` means "child of the primary"; otherwise it is an index into `sys.Companions`. `identifyGroups` MUST use this field, not slice ordering. The existing `stars.assignDesignations` function in `stars/multistar.go` (around line 291) demonstrates the pattern via its `findCompanionByParent` helper. The test in Step 1 sets `ParentIndex` correctly via the field defaults (zero value `0` would point at index 0, so each test case must set `ParentIndex` explicitly when constructing companions).

- [ ] **Step 2: Run tests to verify fail**

```bash
go test ./worlds/ -run TestIdentifyGroups -v
```

Expected: FAIL — `identifyGroups undefined`.

- [ ] **Step 3: Inspect the System / CompanionStar structure**

```bash
grep -n "type CompanionStar\|type System\|OrbitClass" /Users/markayers/Documents/Traveller/stars/system.go | head
```

Read what's actually there before writing the implementation. The pairing semantics may not be position-based.

- [ ] **Step 4: Implement identifyGroups**

Append to `worlds/available_orbits.go`. The implementation follows the structure described in the spec; pairing semantics depend on what Step 3 reveals. Below is the position-based version (companions immediately follow their parent in the slice):

```go
// identifyGroups partitions a System into its barycentric orbit groups.
// See package doc comment for the rules.
//
// Pairing uses CompanionStar.ParentIndex: -1 means "child of the
// primary"; otherwise it is an index into sys.Companions.
func identifyGroups(sys stars.System) []Group {
	groups := []Group{}

	// findCompanionOf returns the Star and CompanionStar that has parent ==
	// parentIdx and OrbitClass == Companion, or (Star{}, false) if none.
	findCompanionOf := func(parentIdx int) (stars.Star, float64, bool) {
		for _, c := range sys.Companions {
			if c.ParentIndex == parentIdx && c.OrbitClass == stars.OrbitCompanion {
				return c.Star, c.Eccentricity, true
			}
		}
		return stars.Star{}, 0, false
	}

	// Primary group: primary plus its companion (parent index -1, class
	// Companion), if any.
	primaryGroup := Group{Members: []stars.Star{sys.Primary}}
	if companion, ecc, ok := findCompanionOf(-1); ok {
		primaryGroup.Members = append(primaryGroup.Members, companion)
		primaryGroup.companionEcc = ecc
		primaryGroup.Designation = "Aab"
	} else {
		primaryGroup.Designation = "A"
	}
	groups = append(groups, primaryGroup)

	// Secondary groups: each Close/Near/Far companion of the primary
	// becomes its own group (with its own companion folded in if any).
	// Walk Close, then Near, then Far in canonical order so designations
	// are assigned positionally (B, C, D) skipping absent slots.
	letters := []string{"A", "B", "C", "D"}
	letterIdx := 1
	for _, oc := range []stars.OrbitClass{stars.OrbitClose, stars.OrbitNear, stars.OrbitFar} {
		for i, c := range sys.Companions {
			if c.OrbitClass != oc || c.ParentIndex != -1 {
				continue
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
	}

	return groups
}
```

- [ ] **Step 5: Run tests to verify pass**

```bash
go test ./worlds/ -run TestIdentifyGroups -v
```

Expected: PASS for all three identification tests.

- [ ] **Step 6: Format, lint, commit**

```bash
gofumpt -w worlds/available_orbits.go worlds/available_orbits_test.go
just check && just test
git add worlds/available_orbits.go worlds/available_orbits_test.go
git commit -m "feat(worlds): group identification from stars.System (WBH p.25)"
```

---

## Task 9: AvailableOrbits skeleton — Rule 1 (MAO assignment) + Rule 3 (primary outer 20)

**Source:** WBH p. 38, rules 1 & 3.

**Files:** `worlds/available_orbits.go` (extend), `worlds/available_orbits_test.go` (extend).

**Goal:** First end-to-end working `AvailableOrbits`. Skips rules 2, 4–11; just sets each group's MAO from the table and clamps the primary group's max to 20.

**API:**

```go
// AvailableOrbits applies the 11 simplified rules from WBH pp. 38–40 to a
// stars.System and returns per-group allowed Orbit# intervals.
//
// Returns ErrPostStellarPrimaryUnsupported if the primary is post-stellar.
func AvailableOrbits(sys stars.System) (Result, error)
```

- [ ] **Step 1: Write failing test**

Append to `worlds/available_orbits_test.go`:

```go
func TestAvailableOrbits_SoleMainSequence(t *testing.T) {
	t.Parallel()

	// Single G2 V primary, no companions. Expected: one group "A" with
	// MAO ~0.03 (interpolated G0→G5) and intervals [[MAO, 20.0]].
	sol := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.0, Diameter: 1.0, Temperature: 5772,
	})
	sys := stars.System{Primary: sol}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	if len(got.Groups) != 1 {
		t.Fatalf("groups = %d, want 1", len(got.Groups))
	}
	g := got.Groups[0]
	if g.Designation != "A" {
		t.Errorf("Designation = %q, want \"A\"", g.Designation)
	}
	if math.Abs(g.MAO-0.03) > 0.005 {
		t.Errorf("MAO = %v, want ~0.03", g.MAO)
	}
	if len(g.Intervals) != 1 {
		t.Fatalf("intervals = %d, want 1", len(g.Intervals))
	}
	iv := g.Intervals[0]
	if math.Abs(iv.Min-g.MAO) > 1e-9 {
		t.Errorf("Min = %v, want MAO %v", iv.Min, g.MAO)
	}
	if math.Abs(iv.Max-20.0) > 1e-9 {
		t.Errorf("Max = %v, want 20.0", iv.Max)
	}
}

func TestAvailableOrbits_PostStellarPrimary(t *testing.T) {
	t.Parallel()

	sys := stars.System{
		Primary: stars.Star{Kind: stars.KindWhiteDwarf, Mass: 0.5},
	}
	_, err := AvailableOrbits(sys)
	if !errors.Is(err, ErrPostStellarPrimaryUnsupported) {
		t.Errorf("err = %v, want ErrPostStellarPrimaryUnsupported", err)
	}
}
```

- [ ] **Step 2: Run test to verify fail**

```bash
go test ./worlds/ -run TestAvailableOrbits -v
```

Expected: FAIL — `AvailableOrbits undefined`.

- [ ] **Step 3: Implement AvailableOrbits skeleton**

Append to `worlds/available_orbits.go`:

```go
// AvailableOrbits applies the WBH pp. 38–40 simplified rules to a
// stars.System and returns per-group allowed Orbit# intervals.
//
// Implementation walks rules 1–11 in order, mutating each group's
// interval set. See spec for the rule list.
func AvailableOrbits(sys stars.System) (Result, error) {
	if isPostStellar(sys.Primary.Kind) {
		return Result{}, ErrPostStellarPrimaryUnsupported
	}

	groups := identifyGroups(sys)

	// Rule 1: MAO from p. 39 table for each group.
	for i := range groups {
		// Pair groups use the larger star's MAO; for now the first
		// member is the parent, which is also the larger.
		mao, err := MAO(groups[i].Members[0])
		if err != nil {
			return Result{}, fmt.Errorf("worlds: MAO for group %s: %w",
				groups[i].Designation, err)
		}
		groups[i].MAO = mao
	}

	// Rule 3: primary group can have Orbit#s up to 20.
	groups[0].Intervals = []Interval{{Min: groups[0].MAO, Max: 20.0}}

	return Result{Groups: groups}, nil
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test ./worlds/ -run TestAvailableOrbits -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/available_orbits.go worlds/available_orbits_test.go
just check && just test
git add worlds/available_orbits.go worlds/available_orbits_test.go
git commit -m "feat(worlds): AvailableOrbits skeleton with rules 1 & 3 (WBH p.38)"
```

---

## Task 10: Rule 2 — companion eccentricity raises pair lower bound

**Source:** WBH p. 38, rule 2.

**Files:** `worlds/available_orbits.go` (extend), `worlds/available_orbits_test.go` (extend).

**Rule:** For pair groups (two members), the lower bound of available orbits becomes `max(group_MAO, 0.50 + companion_eccentricity)`. If the larger star's MAO > 0.2, add the larger star's MAO to that lower bound. The companion's eccentricity comes from the matching `CompanionStar.Eccentricity` in `sys.Companions`.

- [ ] **Step 1: Extend AvailableOrbits signature internally to access eccentricities**

The current skeleton drops the `sys.Companions[i].Eccentricity` data. Restructure: identify groups along with their companion-eccentricity values, or iterate over companions while building intervals.

- [ ] **Step 2: Write failing test**

Append to `worlds/available_orbits_test.go`:

```go
func TestAvailableOrbits_Rule2_CompanionEccentricity(t *testing.T) {
	t.Parallel()

	// Aab pair: primary G7 V, companion G8 V with eccentricity 0.11.
	// Expected: pair group MAO becomes 0.50 + 0.11 = 0.61.
	aa := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass: 0.929, Diameter: 0.967, Temperature: 5440,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V,
		Mass: 0.907, Diameter: 0.957, Temperature: 5360,
	})
	sys := stars.System{
		Primary: aa,
		Companions: []stars.CompanionStar{
			{Star: ab, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.09, Eccentricity: 0.11, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	g := got.Groups[0]
	if math.Abs(g.MAO-0.61) > 1e-9 {
		t.Errorf("Aab MAO = %v, want 0.61", g.MAO)
	}
	if math.Abs(g.Intervals[0].Min-0.61) > 1e-9 {
		t.Errorf("Aab interval Min = %v, want 0.61", g.Intervals[0].Min)
	}
}
```

- [ ] **Step 3: Run test to verify fail**

```bash
go test ./worlds/ -run TestAvailableOrbits_Rule2 -v
```

Expected: FAIL — pair group has MAO ~0.02 (G7 V table value), not 0.61.

- [ ] **Step 4: Implement Rule 2**

The `Group.companionEcc` field already exists from Task 6 and is populated by `identifyGroups` from Task 8 (using `CompanionStar.Eccentricity` of the folded-in companion). This task only needs to apply rule 2 in `AvailableOrbits`.

In `AvailableOrbits`, after Rule 1 (and before Rule 3 sets the primary's intervals):

```go
// Rule 2: companion eccentricity raises pair lower bound.
for i := range groups {
	if len(groups[i].Members) < 2 {
		continue
	}
	floor := 0.50 + groups[i].companionEcc
	larger := groups[i].Members[0] // first member is parent; assume larger
	largerMAO, _ := MAO(larger)
	if largerMAO > 0.2 {
		floor += largerMAO
	}
	if floor > groups[i].MAO {
		groups[i].MAO = floor
	}
}
```

The Rule 3 line `groups[0].Intervals = []Interval{{Min: groups[0].MAO, Max: 20.0}}` from Task 9 must remain _after_ this Rule 2 block so the primary's interval picks up the raised MAO.

- [ ] **Step 5: Run tests to verify pass**

```bash
go test ./worlds/ -v
```

Expected: PASS for all worlds tests including the new Rule 2 test and prior tests.

- [ ] **Step 6: Format, lint, commit**

```bash
gofumpt -w worlds/available_orbits.go worlds/available_orbits_test.go
just check && just test
git add worlds/available_orbits.go worlds/available_orbits_test.go
git commit -m "feat(worlds): rule 2 — companion eccentricity raises pair MAO (WBH p.38)"
```

---

## Task 11: Rules 4 & 5 — secondary exclusion in primary's range

**Source:** WBH p. 38, rules 4 & 5.

**Files:** `worlds/available_orbits.go` (extend), `worlds/available_orbits_test.go` (extend).

**Rule 4:** Each Close/Near/Far secondary's companion is treated as occupying the same Orbit# (already handled by group folding in Task 8); from this point forward, companions are ignored.

**Rule 5:** For each Close/Near/Far secondary at Orbit# `s`, exclude `(s − 1, s + 1)` from the primary group's intervals. If the secondary's MAO > 0.2, add the secondary's MAO to that exclusion.

- [ ] **Step 1: Add `intervalSet` helper**

Add private interval-arithmetic helpers in `worlds/available_orbits.go`:

```go
// subtract removes the exclusion range [exMin, exMax] from a sorted
// disjoint interval list and returns the remaining intervals.
//
// Tolerates exclusions that fully cover or only partially overlap
// existing intervals.
func subtract(intervals []Interval, exMin, exMax float64) []Interval {
	if exMin >= exMax {
		return intervals
	}
	out := make([]Interval, 0, len(intervals)+1)
	for _, iv := range intervals {
		// No overlap.
		if exMax <= iv.Min || exMin >= iv.Max {
			out = append(out, iv)
			continue
		}
		// Left remainder.
		if exMin > iv.Min {
			out = append(out, Interval{Min: iv.Min, Max: exMin})
		}
		// Right remainder.
		if exMax < iv.Max {
			out = append(out, Interval{Min: exMax, Max: iv.Max})
		}
	}
	return out
}
```

- [ ] **Step 2: Write failing test**

Append to `worlds/available_orbits_test.go`:

```go
func TestAvailableOrbits_Rule5_SecondaryExclusion(t *testing.T) {
	t.Parallel()

	// Construct: primary G2 V, single secondary at Orbit# 6.10 (Near).
	// Expected: primary intervals = [[MAO, 5.10], [7.10, 20.0]].
	a := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass: 1.0, Diameter: 1.0, Temperature: 5772,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass: 0.626, Diameter: 0.777, Temperature: 3980,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.0, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	primary := got.Groups[0]
	if len(primary.Intervals) != 2 {
		t.Fatalf("primary intervals = %d, want 2: %+v", len(primary.Intervals), primary.Intervals)
	}
	if math.Abs(primary.Intervals[0].Max-5.10) > 1e-9 {
		t.Errorf("first interval Max = %v, want 5.10", primary.Intervals[0].Max)
	}
	if math.Abs(primary.Intervals[1].Min-7.10) > 1e-9 {
		t.Errorf("second interval Min = %v, want 7.10", primary.Intervals[1].Min)
	}
	if math.Abs(primary.Intervals[1].Max-20.0) > 1e-9 {
		t.Errorf("second interval Max = %v, want 20.0", primary.Intervals[1].Max)
	}
}
```

- [ ] **Step 3: Run test to verify fail**

```bash
go test ./worlds/ -run TestAvailableOrbits_Rule5 -v
```

Expected: FAIL — primary still has one interval [MAO, 20].

- [ ] **Step 4: Implement Rule 5**

In `AvailableOrbits` after Rule 2:

```go
// Rule 5: each Close/Near/Far secondary excludes (s-1, s+1) from
// primary's intervals (with secondary's MAO added if MAO > 0.2).
for _, c := range sys.Companions {
	if c.OrbitClass == stars.OrbitCompanion {
		continue
	}
	s := c.OrbitNumber
	exLow := s - 1
	exHigh := s + 1
	secMAO, _ := MAO(c.Star)
	if secMAO > 0.2 {
		exLow -= secMAO
		exHigh += secMAO
	}
	groups[0].Intervals = subtract(groups[0].Intervals, exLow, exHigh)
}
```

- [ ] **Step 5: Run tests to verify pass**

```bash
go test ./worlds/ -v
```

Expected: PASS for all worlds tests.

- [ ] **Step 6: Format, lint, commit**

```bash
gofumpt -w worlds/available_orbits.go worlds/available_orbits_test.go
just check && just test
git add worlds/available_orbits.go worlds/available_orbits_test.go
git commit -m "feat(worlds): rules 4 & 5 — secondary exclusion in primary's range (WBH p.38)"
```

---

## Task 12: Rule 6 — secondary eccentricity > 0.2 widens exclusion

**Source:** WBH p. 38, rule 6.

**Files:** `worlds/available_orbits.go` (extend), `worlds/available_orbits_test.go` (extend).

**Rule:** For each Close/Near/Far secondary with eccentricity > 0.2, widen its exclusion of the primary's range by ±1 Orbit# on each side (on top of rule 5).

- [ ] **Step 1: Write failing test**

Append to `worlds/available_orbits_test.go`:

```go
func TestAvailableOrbits_Rule6_EccentricSecondary(t *testing.T) {
	t.Parallel()

	// Zed Cab pair (Far) at Orbit# 12.10 with eccentricity 0.47 (> 0.2).
	// Expected: primary's intervals exclude (12.10 - 2, 12.10 + 2) = (10.10, 14.10).
	a := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass: 0.929, Diameter: 0.967, Temperature: 5440,
	})
	ca := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'M', Subtype: 0},
		LuminosityClass: stars.V,
		Mass: 0.510, Diameter: 0.728, Temperature: 3700,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: ca, OrbitClass: stars.OrbitFar, OrbitNumber: 12.10, Eccentricity: 0.47, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	primary := got.Groups[0]
	if len(primary.Intervals) != 2 {
		t.Fatalf("primary intervals = %d, want 2: %+v", len(primary.Intervals), primary.Intervals)
	}
	if math.Abs(primary.Intervals[0].Max-10.10) > 1e-9 {
		t.Errorf("first interval Max = %v, want 10.10", primary.Intervals[0].Max)
	}
	if math.Abs(primary.Intervals[1].Min-14.10) > 1e-9 {
		t.Errorf("second interval Min = %v, want 14.10", primary.Intervals[1].Min)
	}
}
```

- [ ] **Step 2: Run test to verify fail**

```bash
go test ./worlds/ -run TestAvailableOrbits_Rule6 -v
```

Expected: FAIL — exclusion is still ±1 (rule 5 only).

- [ ] **Step 3: Implement Rule 6**

Inside the for-loop that applies rule 5, add a widening step:

```go
for _, c := range sys.Companions {
	if c.OrbitClass == stars.OrbitCompanion {
		continue
	}
	s := c.OrbitNumber
	width := 1.0
	if c.Eccentricity > 0.2 {
		width += 1.0 // rule 6
	}
	exLow := s - width
	exHigh := s + width
	secMAO, _ := MAO(c.Star)
	if secMAO > 0.2 {
		exLow -= secMAO
		exHigh += secMAO
	}
	groups[0].Intervals = subtract(groups[0].Intervals, exLow, exHigh)
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test ./worlds/ -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/available_orbits.go worlds/available_orbits_test.go
just check && just test
git add worlds/available_orbits.go worlds/available_orbits_test.go
git commit -m "feat(worlds): rule 6 — secondary ecc > 0.2 widens exclusion (WBH p.38)"
```

---

## Task 13: Rule 7 — Close/Near eccentricity > 0.5 widens further (Far excluded)

**Source:** WBH p. 39, rule 7.

**Files:** `worlds/available_orbits.go` (extend), `worlds/available_orbits_test.go` (extend).

**Rule:** If a Close or Near secondary has eccentricity > 0.5, widen its exclusion by another ±1 Orbit#. **Far does not trigger this rule.**

- [ ] **Step 1: Write failing tests**

Append to `worlds/available_orbits_test.go`:

```go
func TestAvailableOrbits_Rule7_NearEccentricityGT05(t *testing.T) {
	t.Parallel()

	// Near at Orbit# 6.0 with ecc 0.6 (> 0.5).
	// Rule 5: ±1, Rule 6: ±1 more (= ±2), Rule 7: ±1 more (= ±3).
	// Expected exclusion: (3.0, 9.0).
	a := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass: 1.0, Diameter: 1.0, Temperature: 5772,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass: 0.626, Diameter: 0.777, Temperature: 3980,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.0, Eccentricity: 0.6, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	primary := got.Groups[0]
	if math.Abs(primary.Intervals[0].Max-3.0) > 1e-9 {
		t.Errorf("first Max = %v, want 3.0", primary.Intervals[0].Max)
	}
	if math.Abs(primary.Intervals[1].Min-9.0) > 1e-9 {
		t.Errorf("second Min = %v, want 9.0", primary.Intervals[1].Min)
	}
}

func TestAvailableOrbits_Rule7_FarDoesNotTrigger(t *testing.T) {
	t.Parallel()

	// Far at Orbit# 12 with ecc 0.6: rule 7 must NOT apply.
	// Expected exclusion: ±2 only (rules 5+6).
	a := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass: 1.0, Diameter: 1.0, Temperature: 5772,
	})
	far := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'M', Subtype: 0},
		LuminosityClass: stars.V,
		Mass: 0.510, Diameter: 0.728, Temperature: 3700,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: far, OrbitClass: stars.OrbitFar, OrbitNumber: 12.0, Eccentricity: 0.6, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	primary := got.Groups[0]
	// Exclusion (10, 14): primary intervals [[MAO, 10.0], [14.0, 20.0]].
	if math.Abs(primary.Intervals[0].Max-10.0) > 1e-9 {
		t.Errorf("first Max = %v, want 10.0", primary.Intervals[0].Max)
	}
	if math.Abs(primary.Intervals[1].Min-14.0) > 1e-9 {
		t.Errorf("second Min = %v, want 14.0", primary.Intervals[1].Min)
	}
}
```

- [ ] **Step 2: Run tests to verify fail**

```bash
go test ./worlds/ -run TestAvailableOrbits_Rule7 -v
```

Expected: FAIL — exclusion still ±2.

- [ ] **Step 3: Implement Rule 7**

Update the rule loop to add rule 7 width:

```go
width := 1.0
if c.Eccentricity > 0.2 {
	width += 1.0 // rule 6
}
if c.Eccentricity > 0.5 && (c.OrbitClass == stars.OrbitClose || c.OrbitClass == stars.OrbitNear) {
	width += 1.0 // rule 7 (not Far)
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test ./worlds/ -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/available_orbits.go worlds/available_orbits_test.go
just check && just test
git add worlds/available_orbits.go worlds/available_orbits_test.go
git commit -m "feat(worlds): rule 7 — Close/Near ecc > 0.5 widens exclusion (WBH p.39)"
```

---

## Task 14: Rule 8 — secondary's own range Orbit#-3

**Source:** WBH p. 39, rule 8.

**Files:** `worlds/available_orbits.go` (extend), `worlds/available_orbits_test.go` (extend).

**Rule:** Each Close/Near/Far secondary has its own range, centred on itself, extending up to (its Orbit# − 3) on each side. Lower bound is the secondary's MAO (or the pair MAO already computed for pair groups).

- [ ] **Step 1: Write failing test**

Append to `worlds/available_orbits_test.go`:

```go
func TestAvailableOrbits_Rule8_SecondaryOwnRange(t *testing.T) {
	t.Parallel()

	// Zed B at Orbit# 6.10: own range = [B's MAO, 6.10 - 3] = [0.02, 3.10].
	a := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass: 1.0, Diameter: 1.0, Temperature: 5772,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass: 0.626, Diameter: 0.777, Temperature: 3980,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.08, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	if len(got.Groups) != 2 {
		t.Fatalf("groups = %d, want 2", len(got.Groups))
	}
	bg := got.Groups[1]
	if bg.Designation != "B" {
		t.Errorf("Designation = %q, want \"B\"", bg.Designation)
	}
	if len(bg.Intervals) != 1 {
		t.Fatalf("intervals = %d, want 1", len(bg.Intervals))
	}
	if math.Abs(bg.Intervals[0].Min-bg.MAO) > 1e-9 {
		t.Errorf("Min = %v, want MAO %v", bg.Intervals[0].Min, bg.MAO)
	}
	if math.Abs(bg.Intervals[0].Max-3.10) > 1e-9 {
		t.Errorf("Max = %v, want 3.10", bg.Intervals[0].Max)
	}
}
```

- [ ] **Step 2: Run test to verify fail**

```bash
go test ./worlds/ -run TestAvailableOrbits_Rule8 -v
```

Expected: FAIL — secondary group has no intervals.

- [ ] **Step 3: Implement Rule 8**

Add to `AvailableOrbits` after Rule 7 logic (still inside `AvailableOrbits`):

```go
// Rule 8: each secondary group has its own range centred on its star,
// extending up to (Orbit# - 3) on each side.
//
// Walk sys.Companions to match each non-companion entry to a group
// (groups[1..]) by index. Each non-companion entry corresponds to one
// secondary group.
secIdx := 1
for _, c := range sys.Companions {
	if c.OrbitClass == stars.OrbitCompanion {
		continue
	}
	if secIdx >= len(groups) {
		break
	}
	maxOffset := c.OrbitNumber - 3
	if maxOffset < 0 {
		maxOffset = 0
	}
	groups[secIdx].Intervals = []Interval{
		{Min: groups[secIdx].MAO, Max: maxOffset},
	}
	if groups[secIdx].Intervals[0].Max < groups[secIdx].MAO {
		// Secondary too close to primary for any orbits.
		groups[secIdx].Intervals = nil
	}
	secIdx++
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test ./worlds/ -v
```

Expected: PASS.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/available_orbits.go worlds/available_orbits_test.go
just check && just test
git add worlds/available_orbits.go worlds/available_orbits_test.go
git commit -m "feat(worlds): rule 8 — secondary's own range Orbit#-3 (WBH p.39)"
```

---

## Task 15: Rules 9, 10, 11 — adjacent-zone and self-eccentricity reductions

**Source:** WBH p. 39, rules 9–11.

**Files:** `worlds/available_orbits.go` (extend), `worlds/available_orbits_test.go` (extend).

**Rules (each triggers at most once per secondary):**

- **Rule 9:** secondary loses 1 Orbit# if it has a populated adjacent zone (Close+Near, Near+Far, NOT Close+Far without Near). Primary doesn't trigger this for any secondary.
- **Rule 10:** secondary loses another 1 if it or any adjacent-zone star has eccentricity > 0.2.
- **Rule 11:** secondary loses another 1 if it has eccentricity > 0.5.

- [ ] **Step 1: Write failing test using Zed scenario**

Append to `worlds/available_orbits_test.go`:

```go
func TestAvailableOrbits_Rules9to11_ZedB(t *testing.T) {
	t.Parallel()

	// Zed B reductions (WBH p. 40):
	//  Rule 8: own range to 3.10.
	//  Rule 9: Far is adjacent → -1 → 2.10.
	//  Rule 10: Far ecc 0.47 > 0.2 → -1 → 1.10.
	//  Rule 11: B ecc 0.08 (not > 0.5) → no reduction.
	// Final B max = 1.10.
	a := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass: 0.929, Diameter: 0.967, Temperature: 5440,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass: 0.626, Diameter: 0.777, Temperature: 3980,
	})
	ca := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'M', Subtype: 0},
		LuminosityClass: stars.V,
		Mass: 0.510, Diameter: 0.728, Temperature: 3700,
	})
	sys := stars.System{
		Primary: a,
		Companions: []stars.CompanionStar{
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.08, ParentIndex: -1},
			{Star: ca, OrbitClass: stars.OrbitFar, OrbitNumber: 12.10, Eccentricity: 0.47, ParentIndex: -1},
		},
	}
	got, err := AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	if len(got.Groups) != 3 {
		t.Fatalf("groups = %d, want 3", len(got.Groups))
	}
	bg := got.Groups[1]
	if math.Abs(bg.Intervals[0].Max-1.10) > 1e-9 {
		t.Errorf("B Max = %v, want 1.10", bg.Intervals[0].Max)
	}
}
```

- [ ] **Step 2: Run test to verify fail**

```bash
go test ./worlds/ -run TestAvailableOrbits_Rules9to11 -v
```

Expected: FAIL — B max is currently 3.10 (no reductions applied).

- [ ] **Step 3: Implement Rules 9, 10, 11**

Replace the Rule 8 block with one that also applies rules 9–11. Add helper to detect adjacent-zone presence:

```go
// hasOrbitClass reports whether sys has any non-companion CompanionStar
// in the given orbit class.
func hasOrbitClass(sys stars.System, oc stars.OrbitClass) bool {
	for _, c := range sys.Companions {
		if c.OrbitClass == oc {
			return true
		}
	}
	return false
}

// adjacentEccGT02 reports whether any star in an adjacent zone has
// eccentricity > 0.2. "Adjacent" to Close is Near; to Near is Close
// and Far; to Far is Near.
func adjacentEccGT02(sys stars.System, self stars.OrbitClass) bool {
	adjacencies := map[stars.OrbitClass][]stars.OrbitClass{
		stars.OrbitClose: {stars.OrbitNear},
		stars.OrbitNear:  {stars.OrbitClose, stars.OrbitFar},
		stars.OrbitFar:   {stars.OrbitNear},
	}
	wanted := adjacencies[self]
	for _, c := range sys.Companions {
		for _, oc := range wanted {
			if c.OrbitClass == oc && c.Eccentricity > 0.2 {
				return true
			}
		}
	}
	return false
}

// hasAdjacentZone reports whether secondary in zone `self` has a
// populated adjacent zone, per rule 9 (Close+Far without Near does
// NOT count).
func hasAdjacentZone(sys stars.System, self stars.OrbitClass) bool {
	switch self {
	case stars.OrbitClose:
		return hasOrbitClass(sys, stars.OrbitNear)
	case stars.OrbitNear:
		return hasOrbitClass(sys, stars.OrbitClose) || hasOrbitClass(sys, stars.OrbitFar)
	case stars.OrbitFar:
		return hasOrbitClass(sys, stars.OrbitNear)
	}
	return false
}
```

Then apply rules 9, 10, 11 inside the secondary-group loop (replace the Rule 8 block):

```go
secIdx := 1
for _, c := range sys.Companions {
	if c.OrbitClass == stars.OrbitCompanion {
		continue
	}
	if secIdx >= len(groups) {
		break
	}
	maxOffset := c.OrbitNumber - 3 // rule 8

	// Rule 9: adjacent zone present → -1.
	if hasAdjacentZone(sys, c.OrbitClass) {
		maxOffset -= 1
	}
	// Rule 10: self ecc > 0.2 OR adjacent ecc > 0.2 → -1 more.
	if c.Eccentricity > 0.2 || adjacentEccGT02(sys, c.OrbitClass) {
		maxOffset -= 1
	}
	// Rule 11: self ecc > 0.5 → -1 more.
	if c.Eccentricity > 0.5 {
		maxOffset -= 1
	}

	if maxOffset < 0 {
		maxOffset = 0
	}
	if maxOffset < groups[secIdx].MAO {
		groups[secIdx].Intervals = nil
	} else {
		groups[secIdx].Intervals = []Interval{{Min: groups[secIdx].MAO, Max: maxOffset}}
	}
	secIdx++
}
```

- [ ] **Step 4: Run tests to verify pass**

```bash
go test ./worlds/ -v
```

Expected: PASS for all worlds tests.

- [ ] **Step 5: Format, lint, commit**

```bash
gofumpt -w worlds/available_orbits.go worlds/available_orbits_test.go
just check && just test
git add worlds/available_orbits.go worlds/available_orbits_test.go
git commit -m "feat(worlds): rules 9-11 — adjacent-zone and ecc reductions (WBH p.39)"
```

---

## Task 16: Worked example — Sol single-star

**Source:** WBH p. 67 (Sol/Terra survey form).

**Files:** `worlds/worked_examples_test.go` (create).

**Goal:** Acceptance gate for the simple single-star case.

- [ ] **Step 1: Write the test**

Create `worlds/worked_examples_test.go`:

```go
package worlds_test

import (
	"math"
	"testing"

	"wbh/stars"
	"wbh/worlds"
)

func TestSol_AvailableOrbits(t *testing.T) {
	t.Parallel()

	sol := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.000, Diameter: 1.000, Temperature: 5772,
	})
	sys := stars.System{Primary: sol}
	got, err := worlds.AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	if len(got.Groups) != 1 {
		t.Fatalf("groups = %d, want 1", len(got.Groups))
	}
	g := got.Groups[0]
	if g.Designation != "A" {
		t.Errorf("Designation = %q, want \"A\"", g.Designation)
	}
	if math.Abs(g.MAO-0.03) > 0.005 {
		t.Errorf("MAO = %v, want ~0.03", g.MAO)
	}
	if len(g.Intervals) != 1 {
		t.Fatalf("intervals = %d, want 1", len(g.Intervals))
	}
	if math.Abs(g.Intervals[0].Max-20.0) > 1e-9 {
		t.Errorf("Max = %v, want 20.0", g.Intervals[0].Max)
	}

	// Sanity: full Total ≈ 19.97.
	if math.Abs(g.Total()-19.97) > 0.01 {
		t.Errorf("Total = %v, want ~19.97", g.Total())
	}
}
```

- [ ] **Step 2: Run test**

```bash
go test ./worlds/ -run TestSol -v
```

Expected: PASS.

- [ ] **Step 3: Format, commit**

```bash
gofumpt -w worlds/worked_examples_test.go
just check && just test
git add worlds/worked_examples_test.go
git commit -m "test(worlds): Sol single-star available-orbits regression (WBH p.67)"
```

---

## Task 17: Worked example — Zed quintuple system

**Source:** WBH pp. 39–40, 63 (Zed system available orbits + IISS form).

**Files:** `worlds/worked_examples_test.go` (extend).

**Goal:** Acceptance gate for the full multi-star pipeline. Reproduce the three groups exactly.

- [ ] **Step 1: Write the test**

Append to `worlds/worked_examples_test.go`:

```go
func TestZed_AvailableOrbits(t *testing.T) {
	t.Parallel()

	// WBH p. 40: Zed quintuple available orbits.
	//   Aab pair (G7 V + G8 V):  MAO 0.61, [[0.61, 5.10], [7.10, 10.10], [14.10, 20.00]], Total 13.39
	//   B (K8 V):                MAO 0.02, [[0.02, 1.10]], Total 1.08
	//   Cab pair (M0 V + D):     MAO 0.74, [[0.74, 7.10]], Total 6.36

	aa := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass: 0.929, Diameter: 0.967, Temperature: 5440,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V,
		Mass: 0.907, Diameter: 0.957, Temperature: 5360,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V,
		Mass: 0.626, Diameter: 0.777, Temperature: 3980,
	})
	ca := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence,
		SpectralType: stars.SpectralType{Letter: 'M', Subtype: 0},
		LuminosityClass: stars.V,
		Mass: 0.510, Diameter: 0.728, Temperature: 3700,
	})
	cb := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindWhiteDwarf,
		Mass: 0.490, Diameter: 0.017, Temperature: 6700,
	})
	sys := stars.System{
		Primary: aa,
		Companions: []stars.CompanionStar{
			// Index 0: Ab is companion of primary.
			{Star: ab, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.09, Eccentricity: 0.11, ParentIndex: -1},
			// Index 1: B is Near secondary.
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.08, ParentIndex: -1},
			// Index 2: Ca is Far secondary.
			{Star: ca, OrbitClass: stars.OrbitFar, OrbitNumber: 12.10, Eccentricity: 0.47, ParentIndex: -1},
			// Index 3: Cb is companion of Ca (parent index 2).
			{Star: cb, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.21, Eccentricity: 0.24, ParentIndex: 2},
		},
	}

	got, err := worlds.AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	if len(got.Groups) != 3 {
		t.Fatalf("groups = %d, want 3", len(got.Groups))
	}

	// Group Aab.
	aab := got.Groups[0]
	if aab.Designation != "Aab" {
		t.Errorf("groups[0].Designation = %q, want \"Aab\"", aab.Designation)
	}
	if math.Abs(aab.MAO-0.61) > 0.01 {
		t.Errorf("Aab MAO = %v, want 0.61", aab.MAO)
	}
	wantAab := []worlds.Interval{
		{Min: 0.61, Max: 5.10},
		{Min: 7.10, Max: 10.10},
		{Min: 14.10, Max: 20.00},
	}
	if !intervalsEqual(aab.Intervals, wantAab, 0.01) {
		t.Errorf("Aab intervals = %+v, want %+v", aab.Intervals, wantAab)
	}
	if math.Abs(aab.Total()-13.39) > 0.05 {
		t.Errorf("Aab Total = %v, want 13.39", aab.Total())
	}

	// Group B.
	bg := got.Groups[1]
	if bg.Designation != "B" {
		t.Errorf("groups[1].Designation = %q, want \"B\"", bg.Designation)
	}
	if math.Abs(bg.MAO-0.02) > 0.005 {
		t.Errorf("B MAO = %v, want 0.02", bg.MAO)
	}
	wantB := []worlds.Interval{{Min: 0.02, Max: 1.10}}
	if !intervalsEqual(bg.Intervals, wantB, 0.01) {
		t.Errorf("B intervals = %+v, want %+v", bg.Intervals, wantB)
	}
	if math.Abs(bg.Total()-1.08) > 0.01 {
		t.Errorf("B Total = %v, want 1.08", bg.Total())
	}

	// Group Cab.
	cab := got.Groups[2]
	if cab.Designation != "Cab" {
		t.Errorf("groups[2].Designation = %q, want \"Cab\"", cab.Designation)
	}
	if math.Abs(cab.MAO-0.74) > 0.01 {
		t.Errorf("Cab MAO = %v, want 0.74", cab.MAO)
	}
	wantCab := []worlds.Interval{{Min: 0.74, Max: 7.10}}
	if !intervalsEqual(cab.Intervals, wantCab, 0.01) {
		t.Errorf("Cab intervals = %+v, want %+v", cab.Intervals, wantCab)
	}
	if math.Abs(cab.Total()-6.36) > 0.01 {
		t.Errorf("Cab Total = %v, want 6.36", cab.Total())
	}
}

// intervalsEqual reports whether two interval slices match within tol on each bound.
func intervalsEqual(a, b []worlds.Interval, tol float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(a[i].Min-b[i].Min) > tol || math.Abs(a[i].Max-b[i].Max) > tol {
			return false
		}
	}
	return true
}
```

- [ ] **Step 2: Run test**

```bash
go test ./worlds/ -run TestZed -v
```

Expected:

- **If PASS:** acceptance gate met.
- **If FAIL on Aab Total or intervals:** the Cab group's eccentricity (0.47 > 0.2) should trigger rule 6 widening to ±2 around Cab at 12.10, producing the [10.10, 14.10] gap. Verify rule 6 actually applies to Far secondaries.
- **If FAIL on Cab MAO:** rule 2 should apply with companion ecc 0.24 → 0.50 + 0.24 = 0.74. Verify rule 2's eccentricity field comes from the Cb companion, not the Ca outer.
- **If FAIL on Cab Max (book says 7.10):** Cab is Far (Orbit# 12.10), own range 12.10 - 3 = 9.10. Book example shows the actual maximum is 7.10 because of rule 9 (-1 for adjacent Near) and rule 10 (-1 for Near's ecc... wait, B ecc is 0.08 < 0.2). Re-read p. 40: "the C pair start with Orbit#s out to 9.10 (12.10 - 3) available but this is modified by the presence of star B and by the eccentricity of their orbit around the primary by a total of -2.00 leaving only Orbit#s 0.74 - 7.10 available." So the -2 comes from rule 9 (adjacent B) -1 and **self-ecc 0.47 > 0.2** → rule 10 -1. The current rule 10 implementation already handles "self ecc > 0.2". This should match.

If any failure, debug iteratively and adjust the rule implementations until all three groups match. The fix is in the rule logic, not the test expectations — the book values are authoritative.

- [ ] **Step 3: Format, commit**

```bash
gofumpt -w worlds/worked_examples_test.go
just check && just test
git add worlds/worked_examples_test.go
git commit -m "test(worlds): Zed quintuple system available-orbits regression (WBH pp.39-40)"
```

---

## Task 18: Final integration check + branch readiness

**Goal:** Run the full suite, verify all acceptance criteria from the spec are met, and prepare for merge.

- [ ] **Step 1: Run the full test suite with race detector**

```bash
just check && just test
```

Expected: 0 issues, all tests pass including `-race`.

- [ ] **Step 2: Confirm `ErrSpecialPrimaryClassRedirect` is gone**

```bash
grep -rn "ErrSpecialPrimaryClassRedirect" .
```

Expected: zero matches.

- [ ] **Step 3: Verify spec success criteria one by one**

Read `docs/pass-1/specs/2026-05-02-system-worlds-2a-orbits-design.md` § Success criteria. For each bullet:

- HZCO worked examples within ±0.05: covered by Task 1 + Task 2.
- p. 42 HZCO table reproduction within ±5%: covered by Task 3.
- Sol single-star and Zed quintuple `AvailableOrbits` reproduction: covered by Tasks 16–17.
- `ErrSpecialPrimaryClassRedirect` removed: covered by Task 5 + Step 2 above.
- `just check && just test` clean: covered by Step 1.
- Reader-traceable function-to-page mapping: ensure each new function in `stars/hzco.go` and `worlds/available_orbits.go` has a doc comment citing its WBH page.

- [ ] **Step 4: Check exported API is documented**

```bash
go doc ./worlds
go doc ./stars HZCO
go doc ./stars CompositeHZCO
```

Expected: every exported identifier has a doc comment with WBH page citation.

- [ ] **Step 5: Final commit if anything was tweaked, then announce ready for review**

If any doc-comment cleanup is needed:

```bash
gofumpt -w worlds/available_orbits.go stars/hzco.go
just check && just test
git add -u
git commit -m "docs(wbh): doc-comment WBH page citations for 2A APIs"
```

Otherwise, the branch is ready. Surface to the user that the implementation is complete and present the diff summary:

```bash
git log --oneline main..HEAD
git diff --stat main..HEAD
```

---

## Plan complete

After Task 18:

- `wbh/stars` exposes `Star.HZCO()` and `CompositeHZCO()`.
- `wbh/worlds` exposes `AvailableOrbits`, `MAO`, `Group`, `Interval`, `Result`, `ErrPostStellarPrimaryUnsupported`, `ErrNoMAOForStar`.
- `ErrSpecialPrimaryClassRedirect` is removed.
- Sol single-star and Zed quintuple available-orbits worked examples pass.
- All five HZCO worked examples (Sol, Zed Aab, Zed B, Zed Cab, Corella Aab) pass.
- The p. 42 HZCO table is reproduced by the formula within ±5%.

**v1 spec coverage check (against `2026-05-02-system-worlds-2a-orbits-design.md`):**

- ✅ `Star.HZCO()` formula path: Task 1.
- ✅ `CompositeHZCO()` for circumbinary pairs: Task 2.
- ✅ p. 42 HZCO table fixture verification: Task 3.
- ✅ `generatePrimaryAtClass` helper: Task 4.
- ✅ Class redirect resolution + `ErrSpecialPrimaryClassRedirect` removal: Task 5.
- ✅ `worlds/` package types + Total/Contains: Task 6.
- ✅ MAO table + lookup with interpolation: Task 7.
- ✅ Group identification: Task 8.
- ✅ Rules 1 & 3: Task 9.
- ✅ Rule 2: Task 10.
- ✅ Rules 4 & 5: Task 11.
- ✅ Rule 6: Task 12.
- ✅ Rule 7: Task 13.
- ✅ Rule 8: Task 14.
- ✅ Rules 9, 10, 11: Task 15.
- ✅ `ErrPostStellarPrimaryUnsupported`: Task 9 (verified by Task 9's TestAvailableOrbits_PostStellarPrimary).
- ✅ Worked examples Sol + Zed: Tasks 16 & 17.

**Remaining 2A deferrals (per spec, future sub-projects):**

- Hill-sphere alternate orbit method (future spec, opt-in via `Options`).
- Post-stellar primary MAO (Special Circumstances chapter).
- `Other`-descriptor wart in `stars.GenerateCompanionStar` (separate small follow-up).
- World counts and placement (sub-project 2B).
- Sizing, moons, IISS Class II/III survey (sub-project 2C).
