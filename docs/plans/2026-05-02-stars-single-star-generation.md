# Stars: Single-Star Generation Implementation Plan (Go)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generate a single star with full WBH-faithful spectral class, mass, diameter, temperature, luminosity, age, and peculiar-object handling. Reproduce the Sol/Terra worked example on WBH p. 35 to the digit.

**Architecture:** Pure-function pipeline atop shared `dice` and `roller` packages. All randomness goes through a `roller.Roller` interface so a seed deterministically produces a star. Tables are typed Go literals with WBH page citations; physical quantities use linear interpolation between tabulated grid points.

**Tech Stack:** Go 1.22+, `gofumpt` for format, `golangci-lint` for lint, `go test -race ./...` for tests, `just` as the task runner.

**Scope:** Single-star generation only. Multi-star presence, secondaries, stellar orbits, eccentricity, inclination, designations, and the IISS Class 0/I survey form are deferred to Plan 2.

**Spec:** `tools/world-builder/docs/specs/2026-05-02-world-builder-design.md`

**Source pages:** WBH pp. 14–22, plus pp. 26 (Orbit#) and 35 (Terra/Sol example).

---

## File Structure

| File                                                | Responsibility                                                                                       |
| --------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `tools/world-builder/go.mod`                        | Module definition (`module wbh`, `go 1.22`)                                                          |
| `tools/world-builder/.gitignore`                    | Build artefacts                                                                                      |
| `tools/world-builder/.golangci.yml`                 | Linter config                                                                                        |
| `tools/world-builder/justfile`                      | `check`, `test`, `fmt`, `tidy` recipes                                                               |
| `tools/world-builder/README.md`                     | One-paragraph usage                                                                                  |
| `tools/world-builder/dice/dice.go`                  | `Parse` dice-notation parser                                                                         |
| `tools/world-builder/dice/dice_test.go`             | Dice parser tests                                                                                    |
| `tools/world-builder/roller/roller.go`              | `Roller` interface + `Seeded`, `Scripted`, `Fixed`                                                   |
| `tools/world-builder/roller/roller_test.go`         | Roller tests                                                                                         |
| `tools/world-builder/stars/types.go`                | `SpectralType`, `LuminosityClass`, `StarKind`, `Star`                                                |
| `tools/world-builder/stars/types_test.go`           | Type tests                                                                                           |
| `tools/world-builder/stars/tables.go`               | All Stars-chapter tables (Star Type Determination, Subtype, Mass, Temperature, Diameter, Luminosity) |
| `tools/world-builder/stars/tables_test.go`          | Table integrity tests                                                                                |
| `tools/world-builder/stars/primary.go`              | Star Type Determination + Subtype + class restrictions                                               |
| `tools/world-builder/stars/primary_test.go`         | Primary roll tests                                                                                   |
| `tools/world-builder/stars/physical.go`             | Mass, diameter, temperature, luminosity, interpolation, variance                                     |
| `tools/world-builder/stars/physical_test.go`        | Physical quantity tests                                                                              |
| `tools/world-builder/stars/ages.go`                 | Main sequence / subgiant / giant / final ages                                                        |
| `tools/world-builder/stars/ages_test.go`            | Age formula tests                                                                                    |
| `tools/world-builder/stars/peculiar.go`             | Peculiar/special object kind dispatch                                                                |
| `tools/world-builder/stars/peculiar_test.go`        | Peculiar tests                                                                                       |
| `tools/world-builder/stars/stars.go`                | Public entry points (`GenerateMainSequenceStar`, `Compose`)                                          |
| `tools/world-builder/stars/worked_examples_test.go` | Sol/Terra and Zed-primary regression tests                                                           |

---

## Conventions

- All shell commands assume working directory `/Users/markayers/Documents/Traveller/tools/world-builder/` unless otherwise noted.
- Run tests with `go test -race ./...` and lint with `golangci-lint run`.
- Format with `gofumpt -l -w .`.
- Commits use conventional commit format (`feat(stars): …`, `test(dice): …`, etc.).
- After each green step, commit. Never commit failing tests or lint errors.
- Tests live in `_test.go` files in the same package as the code they exercise (white-box). Worked-example tests use `package stars_test` (black-box).
- Use Go's table-driven test idiom: `for _, tc := range cases { t.Run(tc.name, func(t *testing.T) { ... }) }`.

---

## Task 1: Project scaffolding

**Files:**

- Create: `tools/world-builder/go.mod`
- Create: `tools/world-builder/go.sum` (will be empty initially)
- Create: `tools/world-builder/.gitignore`
- Create: `tools/world-builder/.golangci.yml`
- Create: `tools/world-builder/justfile`
- Create: `tools/world-builder/README.md`

- [ ] **Step 1: Verify parent directory and prerequisites**

```bash
ls -d "/Users/markayers/Documents/Traveller/Mongoose/Core Rules/World Builders Handbook.pdf"
go version
which gofumpt golangci-lint just
mkdir -p /Users/markayers/Documents/Traveller/tools/world-builder
```

If `gofumpt` or `golangci-lint` is not installed, install via:

```bash
go install mvdan.cc/gofumpt@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

If they fail, report DONE_WITH_CONCERNS with exact errors and continue with whatever tools are available.

- [ ] **Step 2: Initialize the module**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go mod init wbh
```

This creates `go.mod` with content:

```text
module wbh

go 1.22
```

(Go may emit `go 1.22.X` matching the local toolchain — that's fine.)

- [ ] **Step 3: Create `.gitignore`**

`tools/world-builder/.gitignore`:

```text
# Go build artefacts
*.exe
*.test
*.out
/bin/
/dist/

# Editor
.vscode/
.idea/

# Coverage
coverage.out
coverage.html
```

- [ ] **Step 4: Create `.golangci.yml`**

`tools/world-builder/.golangci.yml` (golangci-lint v2 schema; gofumpt enforced by the `just check` recipe via the CLI, not by golangci-lint's bundled gofumpt — they can disagree on import grouping):

```yaml
version: "2"

# Note: gofumpt is enforced by the `just check` recipe via the gofumpt CLI,
# not by golangci-lint's bundled gofumpt. The two disagreed on import
# grouping (CLI is more recent), so we use the CLI as the single source
# of truth and keep golangci-lint focused on lint checks.

run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - revive
    - misspell
    - unconvert
    - unparam

issues:
  exclude-use-default: false
```

- [ ] **Step 5: Create `justfile`**

`tools/world-builder/justfile`:

```just
default: check test

check:
    gofumpt -l . | tee /dev/stderr | (! grep .)
    go vet ./...
    golangci-lint run

fmt:
    gofumpt -l -w .

test:
    go test -race ./...

tidy:
    go mod tidy
```

The `check` recipe's gofumpt invocation lists files that would be reformatted; the pipeline through `tee /dev/stderr | (! grep .)` makes the recipe fail (non-zero exit) if any file is unformatted, while still surfacing the file list to the developer.

- [ ] **Step 6: Create `README.md`**

`tools/world-builder/README.md`:

````markdown
# wbh — World Builder's Handbook reference implementation

Go reference implementation of star-system generation procedures from
Mongoose Publishing's _World Builder's Handbook_ (Geir Lanesskog, 2023).

```bash
just test     # run tests
just check    # vet + lint + fmt check
just fmt      # apply formatting
```

The library is the artifact; the CLI is a thin wrapper. See
`tools/world-builder/docs/specs/2026-05-02-world-builder-design.md` at the repo
root for design rationale.
````

- [ ] **Step 7: Verify the module builds (no Go files yet — should succeed with no output)**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go build ./...
```

Expected: silent success. No packages exist yet, so nothing to build, but `go build` should not error.

- [ ] **Step 8: Commit**

```bash
git -C /Users/markayers/Documents/Traveller add tools/world-builder
git -C /Users/markayers/Documents/Traveller commit -m "feat(world-builder): scaffold Go module with golangci-lint/justfile"
```

---

## Task 2: Dice notation parser

**Files:**

- Create: `tools/world-builder/dice/dice.go`
- Create: `tools/world-builder/dice/dice_test.go`

WBH uses notations like `2D`, `1D`, `2D-7`, `1D+5`, `D3`, `D3-1`, `d10`, `d100`, `2d10`. The parser returns a `Spec{Count, Sides, Modifier}`.

- [ ] **Step 1: Write failing tests**

`tools/world-builder/dice/dice_test.go`:

```go
package dice

import (
	"testing"
)

func TestParse_Valid(t *testing.T) {
	cases := []struct {
		notation string
		want     Spec
	}{
		{"2D", Spec{Count: 2, Sides: 6, Modifier: 0}},
		{"1D", Spec{Count: 1, Sides: 6, Modifier: 0}},
		{"2D-7", Spec{Count: 2, Sides: 6, Modifier: -7}},
		{"1D+5", Spec{Count: 1, Sides: 6, Modifier: 5}},
		{"3D+2", Spec{Count: 3, Sides: 6, Modifier: 2}},
		{"D3", Spec{Count: 1, Sides: 3, Modifier: 0}},
		{"D3-1", Spec{Count: 1, Sides: 3, Modifier: -1}},
		{"d10", Spec{Count: 1, Sides: 10, Modifier: 0}},
		{"d100", Spec{Count: 1, Sides: 100, Modifier: 0}},
		{"2d10", Spec{Count: 2, Sides: 10, Modifier: 0}},
	}
	for _, tc := range cases {
		t.Run(tc.notation, func(t *testing.T) {
			got, err := Parse(tc.notation)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.notation, err)
			}
			if got != tc.want {
				t.Fatalf("Parse(%q) = %+v, want %+v", tc.notation, got, tc.want)
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	bad := []string{"", "2", "2X", "Dx", "2D-", "2D+", "D1"}
	for _, n := range bad {
		t.Run(n, func(t *testing.T) {
			if _, err := Parse(n); err == nil {
				t.Fatalf("Parse(%q) succeeded; want error", n)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests, confirm failure (build error: undefined symbols)**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
go test ./dice/...
```

Expected: build failure — `undefined: Parse`, `undefined: Spec`.

- [ ] **Step 3: Implement `Parse`**

`tools/world-builder/dice/dice.go`:

```go
// Package dice parses WBH dice notation strings.
//
// Supports the conventions used in the World Builder's Handbook:
//   - "2D" / "1D"        -> count of d6
//   - "D3", "d10", "d100", "2d10" -> arbitrary-sided dice
//   - "+N" / "-N"        -> integer modifier appended to any of the above
package dice

import (
	"fmt"
	"regexp"
	"strconv"
)

// Spec is the parsed form of a dice notation string.
type Spec struct {
	Count, Sides, Modifier int
}

var pattern = regexp.MustCompile(`^(\d*)[Dd](\d*)(?:([+-])(\d+))?$`)

// Parse converts a WBH dice notation string into a Spec.
//
// Examples:
//
//	Parse("2D")    -> Spec{2, 6, 0}
//	Parse("2D-7")  -> Spec{2, 6, -7}
//	Parse("D3-1")  -> Spec{1, 3, -1}
func Parse(notation string) (Spec, error) {
	m := pattern.FindStringSubmatch(notation)
	if m == nil {
		return Spec{}, fmt.Errorf("dice: invalid notation: %q", notation)
	}
	count := 1
	if m[1] != "" {
		c, err := strconv.Atoi(m[1])
		if err != nil {
			return Spec{}, fmt.Errorf("dice: bad count in %q: %w", notation, err)
		}
		count = c
	}
	sides := 6
	if m[2] != "" {
		s, err := strconv.Atoi(m[2])
		if err != nil {
			return Spec{}, fmt.Errorf("dice: bad sides in %q: %w", notation, err)
		}
		sides = s
	}
	if sides < 2 {
		return Spec{}, fmt.Errorf("dice: invalid sides (<2) in %q", notation)
	}
	modifier := 0
	if m[3] != "" {
		v, err := strconv.Atoi(m[4])
		if err != nil {
			return Spec{}, fmt.Errorf("dice: bad modifier in %q: %w", notation, err)
		}
		if m[3] == "-" {
			v = -v
		}
		modifier = v
	}
	return Spec{Count: count, Sides: sides, Modifier: modifier}, nil
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./dice/...
```

Expected: all tests pass.

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w dice
golangci-lint run ./dice/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/dice
git -C /Users/markayers/Documents/Traveller commit -m "feat(dice): WBH dice notation parser"
```

---

## Task 3: Roller interface and implementations

**Files:**

- Create: `tools/world-builder/roller/roller.go`
- Create: `tools/world-builder/roller/roller_test.go`

- [ ] **Step 1: Write failing tests**

`tools/world-builder/roller/roller_test.go`:

```go
package roller

import (
	"testing"
)

func TestSeeded_Deterministic(t *testing.T) {
	a := NewSeeded(42)
	b := NewSeeded(42)
	for i := range 20 {
		ra, rb := a.Roll("2D"), b.Roll("2D")
		if ra != rb {
			t.Fatalf("seeded rollers diverged at i=%d: %d vs %d", i, ra, rb)
		}
	}
}

func TestSeeded_2DInRange(t *testing.T) {
	r := NewSeeded(1)
	for range 200 {
		v := r.Roll("2D")
		if v < 2 || v > 12 {
			t.Fatalf("2D out of range: %d", v)
		}
	}
}

func TestSeeded_Modifier(t *testing.T) {
	r := NewSeeded(1)
	for range 200 {
		v := r.Roll("2D-7")
		if v < -5 || v > 5 {
			t.Fatalf("2D-7 out of range: %d", v)
		}
	}
}

func TestSeeded_D10(t *testing.T) {
	r := NewSeeded(1)
	for range 200 {
		v := r.Roll("d10")
		if v < 1 || v > 10 {
			t.Fatalf("d10 out of range: %d", v)
		}
	}
}

func TestScripted_Order(t *testing.T) {
	r := NewScripted(7, 9, 11)
	if got := r.Roll("2D"); got != 7 {
		t.Fatalf("first roll = %d, want 7", got)
	}
	if got := r.Roll("2D"); got != 9 {
		t.Fatalf("second roll = %d, want 9", got)
	}
	if got := r.Roll("2D"); got != 11 {
		t.Fatalf("third roll = %d, want 11", got)
	}
}

func TestScripted_PanicsOnExhaustion(t *testing.T) {
	r := NewScripted(5)
	r.Roll("2D")
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on exhausted Scripted")
		}
	}()
	r.Roll("2D")
}

func TestFixed_AlwaysSame(t *testing.T) {
	r := Fixed(8)
	if r.Roll("2D") != 8 {
		t.Fatal("Fixed(8).Roll(\"2D\") != 8")
	}
	if r.Roll("1D") != 8 {
		t.Fatal("Fixed(8).Roll(\"1D\") != 8")
	}
	if r.Roll("d100") != 8 {
		t.Fatal("Fixed(8).Roll(\"d100\") != 8")
	}
}

func TestRollerInterface(t *testing.T) {
	var _ Roller = NewSeeded(1)
	var _ Roller = NewScripted(1)
	var _ Roller = Fixed(1)
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./roller/...
```

Expected: build error.

- [ ] **Step 3: Implement Roller**

`tools/world-builder/roller/roller.go`:

```go
// Package roller provides dice-rolling abstractions used throughout wbh.
//
// Every random draw in the library passes through a Roller. This makes
// seeded reproducibility and scripted-test injection both straightforward.
package roller

import (
	"fmt"
	"math/rand"

	"wbh/dice"
)

// Roller is the interface every dice-driven procedure depends on.
type Roller interface {
	// Roll executes the given dice notation (e.g. "2D", "2D-7", "d10")
	// and returns the result, including any modifier in the notation.
	Roll(notation string) int
}

// Seeded is a production roller backed by a seeded *math/rand.Rand.
type Seeded struct {
	rng *rand.Rand
}

// NewSeeded constructs a Seeded roller with the given seed.
func NewSeeded(seed int64) *Seeded {
	//nolint:gosec // math/rand is intentional; we are not generating crypto material.
	return &Seeded{rng: rand.New(rand.NewSource(seed))}
}

// Roll implements the Roller interface.
func (s *Seeded) Roll(notation string) int {
	spec, err := dice.Parse(notation)
	if err != nil {
		panic(fmt.Errorf("roller.Seeded: %w", err))
	}
	total := spec.Modifier
	for range spec.Count {
		total += s.rng.Intn(spec.Sides) + 1
	}
	return total
}

// Scripted is a test roller that yields preset results in order.
//
// The scripted values are *final results* — the natural roll plus any
// modifier already applied at the call site if appropriate. This keeps
// book worked-examples readable: the book reports e.g. "a 2D roll of 9"
// and the test feeds 9 directly.
type Scripted struct {
	results []int
	idx     int
}

// NewScripted constructs a Scripted roller that returns the supplied
// values in order.
func NewScripted(results ...int) *Scripted {
	return &Scripted{results: results}
}

// Roll implements the Roller interface. Panics if the scripted sequence
// is exhausted; an exhausted Scripted in a test always indicates a bug
// in the test or the procedure being tested.
func (s *Scripted) Roll(notation string) int {
	if s.idx >= len(s.results) {
		panic(fmt.Sprintf("roller.Scripted: exhausted on Roll(%q)", notation))
	}
	v := s.results[s.idx]
	s.idx++
	return v
}

// Fixed is a roller that always returns the same value. Useful for
// property tests where you want to pin one variable while exercising
// others, or for deterministic property-style assertions.
type Fixed int

// Roll implements the Roller interface.
func (f Fixed) Roll(string) int { return int(f) }
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./roller/...
```

Expected: all pass.

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w roller
golangci-lint run ./roller/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/roller
git -C /Users/markayers/Documents/Traveller commit -m "feat(roller): Roller interface with Seeded, Scripted, Fixed"
```

---

## Task 4: Stars data types

**Files:**

- Create: `tools/world-builder/stars/types.go`
- Create: `tools/world-builder/stars/types_test.go`

- [ ] **Step 1: Write failing tests**

`tools/world-builder/stars/types_test.go`:

```go
package stars

import (
	"testing"
)

func TestSpectralType_String(t *testing.T) {
	st := SpectralType{Letter: 'G', Subtype: 7}
	if got := st.String(); got != "G7" {
		t.Fatalf("got %q want %q", got, "G7")
	}
}

func TestSpectralType_Parse(t *testing.T) {
	cases := []struct {
		in   string
		want SpectralType
	}{
		{"G7", SpectralType{Letter: 'G', Subtype: 7}},
		{"M0", SpectralType{Letter: 'M', Subtype: 0}},
		{"M9", SpectralType{Letter: 'M', Subtype: 9}},
		{"O0", SpectralType{Letter: 'O', Subtype: 0}},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseSpectralType(tc.in)
			if err != nil {
				t.Fatalf("ParseSpectralType(%q) error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("got %+v want %+v", got, tc.want)
			}
		})
	}
}

func TestSpectralType_ParseInvalid(t *testing.T) {
	bad := []string{"", "G", "G77", "X3", "g7"}
	for _, in := range bad {
		t.Run(in, func(t *testing.T) {
			if _, err := ParseSpectralType(in); err == nil {
				t.Fatalf("ParseSpectralType(%q) succeeded; want error", in)
			}
		})
	}
}

func TestLuminosityClass_Values(t *testing.T) {
	if string(V) != "V" {
		t.Fatal("V != \"V\"")
	}
	if string(D) != "D" {
		t.Fatal("D != \"D\"")
	}
	if string(BD) != "BD" {
		t.Fatal("BD != \"BD\"")
	}
}

func TestStarKind_Values(t *testing.T) {
	for _, k := range []StarKind{KindMainSequence, KindWhiteDwarf, KindBlackHole, KindBrownDwarf} {
		if k == "" {
			t.Fatalf("empty StarKind: %v", k)
		}
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/...
```

Expected: build error — undefined symbols.

- [ ] **Step 3: Implement types**

`tools/world-builder/stars/types.go`:

```go
// Package stars implements the WBH Stars chapter (pp. 14–35).
package stars

import (
	"fmt"
)

// SpectralLetter is one of O, B, A, F, G, K, M (WBH p. 14).
type SpectralLetter rune

// SpectralType is a spectral letter plus a 0-9 subtype.
type SpectralType struct {
	Letter  SpectralLetter
	Subtype int
}

// String renders SpectralType as e.g. "G7".
func (s SpectralType) String() string {
	return fmt.Sprintf("%c%d", s.Letter, s.Subtype)
}

// ParseSpectralType parses a 2-character string like "G7" into a SpectralType.
func ParseSpectralType(s string) (SpectralType, error) {
	if len(s) != 2 {
		return SpectralType{}, fmt.Errorf("stars: spectral type must be 2 chars: %q", s)
	}
	letter := SpectralLetter(s[0])
	switch letter {
	case 'O', 'B', 'A', 'F', 'G', 'K', 'M':
		// ok
	default:
		return SpectralType{}, fmt.Errorf("stars: invalid spectral letter: %q", s)
	}
	d := s[1]
	if d < '0' || d > '9' {
		return SpectralType{}, fmt.Errorf("stars: subtype must be a digit: %q", s)
	}
	return SpectralType{Letter: letter, Subtype: int(d - '0')}, nil
}

// LuminosityClass is a stellar luminosity class (WBH p. 14).
type LuminosityClass string

// Luminosity class constants. D and BD are book conventions for white
// dwarf and brown dwarf, respectively (not formal Yerkes classes).
const (
	Ia  LuminosityClass = "Ia"
	Ib  LuminosityClass = "Ib"
	II  LuminosityClass = "II"
	III LuminosityClass = "III"
	IV  LuminosityClass = "IV"
	V   LuminosityClass = "V"
	VI  LuminosityClass = "VI"
	D   LuminosityClass = "D"
	BD  LuminosityClass = "BD"
)

// StarKind is the top-level kind of stellar/post-stellar object (WBH pp. 14–22).
type StarKind string

// StarKind constants.
const (
	KindMainSequence StarKind = "main_sequence"
	KindGiant        StarKind = "giant"
	KindSubgiant     StarKind = "subgiant"
	KindSupergiant   StarKind = "supergiant"
	KindSubdwarf     StarKind = "subdwarf"
	KindBrownDwarf   StarKind = "brown_dwarf"
	KindWhiteDwarf   StarKind = "white_dwarf"
	KindNeutronStar  StarKind = "neutron_star"
	KindBlackHole    StarKind = "black_hole"
	KindPulsar       StarKind = "pulsar"
	KindNebula       StarKind = "nebula"
	KindProtostar    StarKind = "protostar"
	KindStarCluster  StarKind = "star_cluster"
	KindAnomaly      StarKind = "anomaly"
)

// Star represents a single stellar (or stellar-remnant) object.
//
// Mass, Diameter, and Luminosity are in solar units.
// Temperature is in Kelvin. AgeGyr is in giga-years.
type Star struct {
	Kind            StarKind
	SpectralType    SpectralType
	LuminosityClass LuminosityClass
	Mass            float64
	Diameter        float64
	Temperature     float64
	Luminosity      float64
	AgeGyr          float64
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

Expected: all pass.

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): SpectralType, LuminosityClass, StarKind, Star types"
```

---

## Task 5: Star Type Determination table

**Files:**

- Create: `tools/world-builder/stars/tables.go`
- Create: `tools/world-builder/stars/tables_test.go`

- [ ] **Step 1: Write failing tests**

`tools/world-builder/stars/tables_test.go`:

```go
package stars

import "testing"

func TestStarTypeDetermination_Complete(t *testing.T) {
	for r := 2; r <= 12; r++ {
		row, ok := StarTypeDetermination[r]
		if !ok {
			t.Fatalf("missing row %d", r)
		}
		if row.Type == "" || row.Hot == "" || row.Special == "" ||
			row.Unusual == "" || row.Giants == "" || row.Peculiar == "" {
			t.Fatalf("row %d has empty cell: %+v", r, row)
		}
	}
}

func TestStarTypeDetermination_KnownCells(t *testing.T) {
	// WBH p. 15 spot checks.
	checks := []struct {
		row    int
		col    string
		want   string
	}{
		{7, "Type", "K"},
		{7, "Hot", "A"},
		{2, "Type", "Special"},
		{2, "Peculiar", "Black Hole"},
		{12, "Type", "Hot"},
		{12, "Hot", "O"},
		{12, "Giants", "Class Ia"},
		{11, "Type", "F"},
		{11, "Peculiar", "Anomaly"},
	}
	for _, c := range checks {
		t.Run(c.col, func(t *testing.T) {
			row := StarTypeDetermination[c.row]
			var got string
			switch c.col {
			case "Type":
				got = row.Type
			case "Hot":
				got = row.Hot
			case "Special":
				got = row.Special
			case "Unusual":
				got = row.Unusual
			case "Giants":
				got = row.Giants
			case "Peculiar":
				got = row.Peculiar
			}
			if got != c.want {
				t.Fatalf("row %d %s = %q, want %q", c.row, c.col, got, c.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run StarTypeDetermination
```

Expected: build error — `undefined: StarTypeDetermination`.

- [ ] **Step 3: Implement table**

`tools/world-builder/stars/tables.go`:

```go
package stars

// StarTypeRow is one row of the WBH p. 15 Star Type Determination table.
//
// Default column is Type; rolls of 2 redirect to Special, rolls of 12
// redirect to Hot. Class III+ rolls go to Giants. The Unusual and
// Peculiar columns are entered only when the procedure says so.
type StarTypeRow struct {
	Type, Hot, Special, Unusual, Giants, Peculiar string
}

// StarTypeDetermination is the WBH p. 15 Star Type Determination table.
var StarTypeDetermination = map[int]StarTypeRow{
	2:  {Type: "Special", Hot: "A", Special: "Class VI", Unusual: "Peculiar", Giants: "Class III", Peculiar: "Black Hole"},
	3:  {Type: "M", Hot: "A", Special: "Class VI", Unusual: "Class VI", Giants: "Class III", Peculiar: "Pulsar"},
	4:  {Type: "M", Hot: "A", Special: "Class VI", Unusual: "Class IV", Giants: "Class III", Peculiar: "Neutron Star"},
	5:  {Type: "M", Hot: "A", Special: "Class VI", Unusual: "BD", Giants: "Class III", Peculiar: "Nebula"},
	6:  {Type: "M", Hot: "A", Special: "Class IV", Unusual: "BD", Giants: "Class III", Peculiar: "Nebula"},
	7:  {Type: "K", Hot: "A", Special: "Class IV", Unusual: "BD", Giants: "Class III", Peculiar: "Protostar"},
	8:  {Type: "K", Hot: "A", Special: "Class IV", Unusual: "D", Giants: "Class III", Peculiar: "Protostar"},
	9:  {Type: "G", Hot: "A", Special: "Class III", Unusual: "D", Giants: "Class II", Peculiar: "Protostar"},
	10: {Type: "G", Hot: "B", Special: "Class III", Unusual: "D", Giants: "Class II", Peculiar: "Star Cluster"},
	11: {Type: "F", Hot: "B", Special: "Giants", Unusual: "Class III", Giants: "Class Ib", Peculiar: "Anomaly"},
	12: {Type: "Hot", Hot: "O", Special: "Giants", Unusual: "Giants", Giants: "Class Ia", Peculiar: "Anomaly"},
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): Star Type Determination table (WBH p.15)"
```

---

## Task 6: Star Subtype tables

**Files:**

- Modify: `tools/world-builder/stars/tables.go` (append)
- Modify: `tools/world-builder/stars/tables_test.go` (append)

- [ ] **Step 1: Append failing tests**

Append to `tools/world-builder/stars/tables_test.go`:

```go
func TestStarSubtype_Complete(t *testing.T) {
	for r := 2; r <= 12; r++ {
		if _, ok := StarSubtypeNumeric[r]; !ok {
			t.Fatalf("Numeric: missing row %d", r)
		}
		if _, ok := StarSubtypeMType[r]; !ok {
			t.Fatalf("M-type: missing row %d", r)
		}
	}
}

func TestStarSubtype_KnownCells(t *testing.T) {
	// WBH p. 16 — Zed primary: 2D=6 -> Numeric subtype 7 (G7).
	if got := StarSubtypeNumeric[6]; got != 7 {
		t.Fatalf("Numeric[6] = %d, want 7", got)
	}
	if got := StarSubtypeNumeric[2]; got != 0 {
		t.Fatalf("Numeric[2] = %d, want 0", got)
	}
	if got := StarSubtypeNumeric[12]; got != 0 {
		t.Fatalf("Numeric[12] = %d, want 0", got)
	}
	if got := StarSubtypeMType[6]; got != 0 {
		t.Fatalf("MType[6] = %d, want 0", got)
	}
	if got := StarSubtypeMType[12]; got != 9 {
		t.Fatalf("MType[12] = %d, want 9", got)
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run StarSubtype
```

Expected: undefined symbols.

- [ ] **Step 3: Append tables**

Append to `tools/world-builder/stars/tables.go`:

```go
// StarSubtypeNumeric is the WBH p. 16 Star Subtype table — Numeric column.
var StarSubtypeNumeric = map[int]int{
	2: 0, 3: 1, 4: 3, 5: 5, 6: 7, 7: 9, 8: 8, 9: 6, 10: 4, 11: 2, 12: 0,
}

// StarSubtypeMType is the WBH p. 16 Star Subtype table — M-type column (primary only).
var StarSubtypeMType = map[int]int{
	2: 8, 3: 6, 4: 5, 5: 4, 6: 0, 7: 2, 8: 1, 9: 3, 10: 5, 11: 7, 12: 9,
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): Star Subtype tables (WBH p.16)"
```

---

## Task 7: Primary star type and subtype rolls

**Files:**

- Create: `tools/world-builder/stars/primary.go`
- Create: `tools/world-builder/stars/primary_test.go`

- [ ] **Step 1: Write failing tests**

`tools/world-builder/stars/primary_test.go`:

```go
package stars

import (
	"testing"

	"wbh/roller"
)

func TestRollPrimaryTypeAndClass_Zed(t *testing.T) {
	// WBH p. 16: "the roll for primary star is a 9, resulting in a type
	// G-type main sequence, Class V star."
	r := roller.NewScripted(9)
	letter, lc, err := RollPrimaryTypeAndClass(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if letter != 'G' {
		t.Fatalf("letter = %c, want G", letter)
	}
	if lc != V {
		t.Fatalf("class = %s, want V", lc)
	}
}

func TestRollPrimaryTypeAndClass_HotRedirect(t *testing.T) {
	// 12 -> "Hot" -> reroll on Hot column.
	// 12 followed by 5 -> Hot column at 5 = "A".
	r := roller.NewScripted(12, 5)
	letter, lc, err := RollPrimaryTypeAndClass(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if letter != 'A' {
		t.Fatalf("letter = %c, want A", letter)
	}
	if lc != V {
		t.Fatalf("class = %s, want V", lc)
	}
}

func TestRollPrimaryTypeAndClass_SpecialNotImplemented(t *testing.T) {
	// 2 -> "Special" branch is dispatched separately; for v1 (Plan 1) we
	// surface ErrSpecialPrimary so callers can route through peculiar.go.
	r := roller.NewScripted(2)
	_, _, err := RollPrimaryTypeAndClass(r)
	if err == nil {
		t.Fatal("expected ErrSpecialPrimary")
	}
}

func TestRollSubtype_Zed(t *testing.T) {
	// WBH p. 16: a 2D roll of 6 on the Numeric column -> subtype 7 (G7).
	r := roller.NewScripted(6)
	got, err := RollSubtype(r, 'G', V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 7 {
		t.Fatalf("got %d, want 7", got)
	}
}

func TestRollSubtype_MType(t *testing.T) {
	// Primary M-type uses the M column. 2D=6 on M column -> 0 (M0).
	r := roller.NewScripted(6)
	got, err := RollSubtype(r, 'M', V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 0 {
		t.Fatalf("got %d, want 0", got)
	}
}

func TestRollSubtype_KClassIVAdjustment(t *testing.T) {
	// WBH p. 16: "For a K-type Class IV star, subtract 5 (make lower)
	// any subtype result above 4."
	// 2D=7 -> Numeric[7] = 9; for K Class IV, 9 - 5 = 4.
	r := roller.NewScripted(7)
	got, err := RollSubtype(r, 'K', IV)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 4 {
		t.Fatalf("K Class IV adjustment: got %d, want 4", got)
	}
}

func TestRollSubtype_AllRollsInRange(t *testing.T) {
	for v := 2; v <= 12; v++ {
		r := roller.NewScripted(v)
		got, err := RollSubtype(r, 'G', V)
		if err != nil {
			t.Fatalf("v=%d error: %v", v, err)
		}
		if got < 0 || got > 9 {
			t.Fatalf("v=%d subtype out of range: %d", v, got)
		}
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run RollPrimary
```

Expected: undefined symbols.

- [ ] **Step 3: Implement**

`tools/world-builder/stars/primary.go`:

```go
package stars

import (
	"errors"
	"fmt"

	"wbh/roller"
)

// ErrSpecialPrimary is returned by RollPrimaryTypeAndClass when the
// initial Type-column roll is 2 ("Special"). Callers should dispatch
// through peculiar.go to resolve the special-object branch.
var ErrSpecialPrimary = errors.New("stars: special primary; dispatch through peculiar")

var validLetters = map[SpectralLetter]struct{}{
	'O': {}, 'B': {}, 'A': {}, 'F': {}, 'G': {}, 'K': {}, 'M': {},
}

// RollPrimaryTypeAndClass rolls for the spectral letter and luminosity
// class of a primary star (WBH p. 15).
//
// A roll of 12 redirects to the Hot column with a fresh 2D roll.
// A roll of 2 returns ErrSpecialPrimary so the caller can route through
// the Special / Unusual / Peculiar dispatch in peculiar.go.
func RollPrimaryTypeAndClass(r roller.Roller) (SpectralLetter, LuminosityClass, error) {
	first := r.Roll("2D")
	row, ok := StarTypeDetermination[first]
	if !ok {
		return 0, "", fmt.Errorf("stars: 2D out of range: %d", first)
	}
	cell := row.Type

	if cell == "Hot" {
		second := r.Roll("2D")
		hotRow, ok := StarTypeDetermination[second]
		if !ok {
			return 0, "", fmt.Errorf("stars: 2D out of range: %d", second)
		}
		cell = hotRow.Hot
	}

	if cell == "Special" {
		return 0, "", ErrSpecialPrimary
	}

	if len(cell) != 1 {
		return 0, "", fmt.Errorf("stars: unexpected primary type cell: %q", cell)
	}
	letter := SpectralLetter(cell[0])
	if _, ok := validLetters[letter]; !ok {
		return 0, "", fmt.Errorf("stars: invalid letter from table: %c", letter)
	}
	return letter, V, nil
}

// RollSubtype rolls Star Subtype (WBH p. 16) and returns 0-9.
//
// Primary M-type stars use the M column; all others use the Numeric
// column. K-type Class IV stars have any subtype > 4 reduced by 5.
func RollSubtype(r roller.Roller, letter SpectralLetter, lc LuminosityClass) (int, error) {
	if _, ok := validLetters[letter]; !ok {
		return 0, fmt.Errorf("stars: invalid letter: %c", letter)
	}
	roll := r.Roll("2D")
	var sub int
	if letter == 'M' {
		s, ok := StarSubtypeMType[roll]
		if !ok {
			return 0, fmt.Errorf("stars: 2D out of range: %d", roll)
		}
		sub = s
	} else {
		s, ok := StarSubtypeNumeric[roll]
		if !ok {
			return 0, fmt.Errorf("stars: 2D out of range: %d", roll)
		}
		sub = s
	}
	if letter == 'K' && lc == IV && sub > 4 {
		sub -= 5
	}
	return sub, nil
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): primary star type and subtype rolls (WBH pp.15-16)"
```

---

## Task 8: Class IV/VI/III+ restrictions

**Files:**

- Modify: `tools/world-builder/stars/primary.go` (append)
- Modify: `tools/world-builder/stars/primary_test.go` (append)

WBH p. 16:

- Class IV is limited to types B0–K4. Map M results to K, O results to B (for Hot rolls).
- Class VI consists of F or A populations; remap "Treat results of F as G and A as B."
- Class III+ requires a roll in the Giants column with DM+1.

- [ ] **Step 1: Append failing tests**

Append to `tools/world-builder/stars/primary_test.go`:

```go
func TestApplyClassIVLetterConstraint(t *testing.T) {
	cases := []struct{ in, want SpectralLetter }{
		{'M', 'K'},
		{'O', 'B'},
		{'G', 'G'},
		{'A', 'A'},
	}
	for _, c := range cases {
		if got := ApplyClassIVLetterConstraint(c.in); got != c.want {
			t.Fatalf("Class IV(%c) = %c, want %c", c.in, got, c.want)
		}
	}
}

func TestApplyClassVILetterConstraint(t *testing.T) {
	cases := []struct{ in, want SpectralLetter }{
		{'F', 'G'},
		{'A', 'B'},
		{'G', 'G'},
		{'M', 'M'},
	}
	for _, c := range cases {
		if got := ApplyClassVILetterConstraint(c.in); got != c.want {
			t.Fatalf("Class VI(%c) = %c, want %c", c.in, got, c.want)
		}
	}
}

func TestRollGiantClass(t *testing.T) {
	// WBH p. 16: Class III+ requires a roll in the Giants column with DM+1.
	// 2D = 7 + DM+1 = row 8. Giants column at row 8 = "Class III".
	r := roller.NewScripted(7)
	got, err := RollGiantClass(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != III {
		t.Fatalf("got %s, want %s", got, III)
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run "ApplyClass|RollGiant"
```

Expected: undefined symbols.

- [ ] **Step 3: Append implementation**

Append to `tools/world-builder/stars/primary.go`:

```go
// ApplyClassIVLetterConstraint maps a spectral letter for Class IV
// limits (WBH p. 16). M -> K (subtype shift handled in RollSubtype),
// O -> B (for Hot rolls). Other letters pass through.
func ApplyClassIVLetterConstraint(letter SpectralLetter) SpectralLetter {
	switch letter {
	case 'M':
		return 'K'
	case 'O':
		return 'B'
	default:
		return letter
	}
}

// ApplyClassVILetterConstraint maps a spectral letter for Class VI
// (WBH p. 16): F -> G, A -> B. Other letters pass through.
func ApplyClassVILetterConstraint(letter SpectralLetter) SpectralLetter {
	switch letter {
	case 'F':
		return 'G'
	case 'A':
		return 'B'
	default:
		return letter
	}
}

// RollGiantClass rolls the Giants column with DM+1 to determine the
// final luminosity class for a Class III+ result (WBH p. 16).
func RollGiantClass(r roller.Roller) (LuminosityClass, error) {
	natural := r.Roll("2D")
	row := min(natural+1, 12)
	rowData, ok := StarTypeDetermination[row]
	if !ok {
		return "", fmt.Errorf("stars: 2D out of range: %d", row)
	}
	switch rowData.Giants {
	case "Class III":
		return III, nil
	case "Class II":
		return II, nil
	case "Class Ib":
		return Ib, nil
	case "Class Ia":
		return Ia, nil
	default:
		return "", fmt.Errorf("stars: unexpected Giants cell: %q", rowData.Giants)
	}
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): Class IV/VI/III+ restrictions (WBH p.16)"
```

---

## Task 9: Mass / Diameter / Luminosity tables (shared row type) and Temperature table

**Files:**

- Modify: `tools/world-builder/stars/tables.go`
- Modify: `tools/world-builder/stars/tables_test.go`

The Mass, Diameter, and Luminosity tables (WBH pp. 17, 19) all have the same row shape: `Ia | Ib | II | III | IV | V | VI`, with some cells absent. Use a shared `ClassRow` struct with `*float64` fields. The Temperature column on p. 17 is 1D (one value per spectral type).

- [ ] **Step 1: Append failing tests**

Append to `tools/world-builder/stars/tables_test.go`:

```go
func TestStarMass_KnownCells(t *testing.T) {
	// WBH p. 17 — spot checks.
	tests := []struct {
		spectral string
		class    LuminosityClass
		want     float64
	}{
		{"G0", V, 1.1},
		{"G5", V, 0.9},
		{"K0", V, 0.8},
		{"O0", Ia, 200},
		{"M9", VI, 0.075},
	}
	for _, tc := range tests {
		row, ok := StarMass[tc.spectral]
		if !ok {
			t.Fatalf("missing row %s", tc.spectral)
		}
		got, ok := row.Get(tc.class)
		if !ok {
			t.Fatalf("missing %s %s cell", tc.spectral, tc.class)
		}
		if got != tc.want {
			t.Fatalf("%s %s = %v, want %v", tc.spectral, tc.class, got, tc.want)
		}
	}
	// Class IV not available for O0; should be absent.
	if _, ok := StarMass["O0"].Get(IV); ok {
		t.Fatal("O0 IV unexpectedly present")
	}
}

func TestStarTemperature_KnownCells(t *testing.T) {
	cases := map[string]float64{
		"G0": 6000, "G5": 5600, "K0": 5200, "O0": 50000, "M9": 2400,
	}
	for s, want := range cases {
		if got := StarTemperature[s]; got != want {
			t.Fatalf("%s = %v, want %v", s, got, want)
		}
	}
}

func TestStarDiameter_KnownCells(t *testing.T) {
	tests := []struct {
		spectral string
		class    LuminosityClass
		want     float64
	}{
		{"G0", V, 1.1},
		{"G5", V, 0.95},
		{"K0", V, 0.9},
		{"O0", Ia, 25},
		{"M9", VI, 0.08},
	}
	for _, tc := range tests {
		got, ok := StarDiameter[tc.spectral].Get(tc.class)
		if !ok {
			t.Fatalf("missing %s %s", tc.spectral, tc.class)
		}
		if got != tc.want {
			t.Fatalf("%s %s = %v, want %v", tc.spectral, tc.class, got, tc.want)
		}
	}
}

func TestStarLuminosity_KnownCells(t *testing.T) {
	tests := []struct {
		spectral string
		class    LuminosityClass
		want     float64
	}{
		{"G0", V, 1.4},
		{"G5", V, 0.78},
		{"K0", V, 0.52},
		{"O0", Ia, 3_400_000},
		{"M9", VI, 0.00019},
	}
	for _, tc := range tests {
		got, ok := StarLuminosity[tc.spectral].Get(tc.class)
		if !ok {
			t.Fatalf("missing %s %s", tc.spectral, tc.class)
		}
		if got != tc.want {
			t.Fatalf("%s %s = %v, want %v", tc.spectral, tc.class, got, tc.want)
		}
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run "StarMass|StarTemperature|StarDiameter|StarLuminosity"
```

Expected: undefined symbols.

- [ ] **Step 3: Append `ClassRow`, `Get`, and the four tables**

Append to `tools/world-builder/stars/tables.go`:

```go
// ClassRow holds class-keyed values for tables shaped like the Mass,
// Diameter, and Luminosity tables on WBH pp. 17, 19. A nil pointer
// indicates the book leaves the cell blank ("—").
type ClassRow struct {
	Ia, Ib, II, III, IV, V, VI *float64
}

// Get returns the value for the given luminosity class, or false if
// the cell is absent.
func (r ClassRow) Get(lc LuminosityClass) (float64, bool) {
	var p *float64
	switch lc {
	case Ia:
		p = r.Ia
	case Ib:
		p = r.Ib
	case II:
		p = r.II
	case III:
		p = r.III
	case IV:
		p = r.IV
	case V:
		p = r.V
	case VI:
		p = r.VI
	}
	if p == nil {
		return 0, false
	}
	return *p, true
}

func f(x float64) *float64 { return &x }

// StarMass is the WBH p. 17 Star Mass and Temperature by Class table — Mass column.
// Values are in solar masses (Sol = 1.0).
var StarMass = map[string]ClassRow{
	"O0": {Ia: f(200), Ib: f(150), II: f(130), III: f(110), V: f(90), VI: f(2)},
	"O5": {Ia: f(80), Ib: f(60), II: f(40), III: f(30), V: f(60), VI: f(1.5)},
	"B0": {Ia: f(60), Ib: f(40), II: f(30), III: f(20), IV: f(20), V: f(18), VI: f(0.5)},
	"B5": {Ia: f(30), Ib: f(25), II: f(20), III: f(10), IV: f(10), V: f(5), VI: f(0.4)},
	"A0": {Ia: f(20), Ib: f(15), II: f(14), III: f(8), IV: f(4), V: f(2.2)},
	"A5": {Ia: f(15), Ib: f(13), II: f(11), III: f(6), IV: f(2.3), V: f(1.5)},
	"F0": {Ia: f(13), Ib: f(12), II: f(10), III: f(4), IV: f(2), V: f(1.5)},
	"F5": {Ia: f(12), Ib: f(10), II: f(8), III: f(3), IV: f(1.5), V: f(1.3)},
	"G0": {Ia: f(12), Ib: f(10), II: f(8), III: f(2.5), IV: f(1.7), V: f(1.1), VI: f(0.8)},
	"G5": {Ia: f(13), Ib: f(11), II: f(10), III: f(2.4), IV: f(1.2), V: f(0.9), VI: f(0.7)},
	"K0": {Ia: f(14), Ib: f(12), II: f(10), III: f(1.1), IV: f(1.5), V: f(0.8), VI: f(0.6)},
	"K5": {Ia: f(18), Ib: f(13), II: f(12), III: f(1.5), V: f(0.7), VI: f(0.5)},
	"M0": {Ia: f(20), Ib: f(15), II: f(14), III: f(1.8), V: f(0.5), VI: f(0.4)},
	"M5": {Ia: f(25), Ib: f(20), II: f(16), III: f(2.4), V: f(0.16), VI: f(0.12)},
	"M9": {Ia: f(30), Ib: f(25), II: f(18), III: f(8), V: f(0.08), VI: f(0.075)},
}

// StarTemperature is the WBH p. 17 Star Mass and Temperature by Class table —
// Temperature column. Values are in Kelvin.
var StarTemperature = map[string]float64{
	"O0": 50000, "O5": 40000, "B0": 30000, "B5": 15000,
	"A0": 10000, "A5": 8000,
	"F0": 7500, "F5": 6500,
	"G0": 6000, "G5": 5600,
	"K0": 5200, "K5": 4400,
	"M0": 3700, "M5": 3000, "M9": 2400,
}

// StarDiameter is the WBH p. 19 Star Diameter by Class table.
// Values are in solar diameters (Sol = 1.0).
var StarDiameter = map[string]ClassRow{
	"O0": {Ia: f(25), Ib: f(24), II: f(22), III: f(21), V: f(20), VI: f(0.18)},
	"O5": {Ia: f(22), Ib: f(20), II: f(18), III: f(15), V: f(12), VI: f(0.18)},
	"B0": {Ia: f(20), Ib: f(14), II: f(12), III: f(10), IV: f(8), V: f(7), VI: f(0.2)},
	"B5": {Ia: f(60), Ib: f(25), II: f(14), III: f(6), IV: f(5), V: f(3.5), VI: f(0.5)},
	"A0": {Ia: f(120), Ib: f(50), II: f(30), III: f(5), IV: f(4), V: f(2.2)},
	"A5": {Ia: f(180), Ib: f(75), II: f(45), III: f(5), IV: f(3), V: f(2)},
	"F0": {Ia: f(210), Ib: f(85), II: f(50), III: f(5), IV: f(3), V: f(1.7)},
	"F5": {Ia: f(280), Ib: f(115), II: f(66), III: f(5), IV: f(2), V: f(1.5)},
	"G0": {Ia: f(330), Ib: f(135), II: f(77), III: f(10), IV: f(3), V: f(1.1), VI: f(0.8)},
	"G5": {Ia: f(360), Ib: f(150), II: f(90), III: f(15), IV: f(4), V: f(0.95), VI: f(0.7)},
	"K0": {Ia: f(420), Ib: f(180), II: f(110), III: f(20), IV: f(6), V: f(0.9), VI: f(0.6)},
	"K5": {Ia: f(600), Ib: f(260), II: f(160), III: f(40), V: f(0.8), VI: f(0.5)},
	"M0": {Ia: f(900), Ib: f(380), II: f(230), III: f(60), V: f(0.7), VI: f(0.4)},
	"M5": {Ia: f(1200), Ib: f(600), II: f(350), III: f(100), V: f(0.2), VI: f(0.1)},
	"M9": {Ia: f(1800), Ib: f(800), II: f(500), III: f(200), V: f(0.1), VI: f(0.08)},
}

// StarLuminosity is the WBH p. 19 Star Luminosity by Class table.
// Values are in solar luminosities (Sol = 1.0).
var StarLuminosity = map[string]ClassRow{
	"O0": {Ia: f(3_400_000), Ib: f(3_200_000), II: f(2_700_000), III: f(2_400_000), V: f(2_200_000), VI: f(180)},
	"O5": {Ia: f(1_100_000), Ib: f(900_000), II: f(730_000), III: f(510_000), V: f(330_000), VI: f(73)},
	"B0": {Ia: f(290_000), Ib: f(140_000), II: f(100_000), III: f(72_000), IV: f(46_000), V: f(35_000), VI: f(29)},
	"B5": {Ia: f(160_000), Ib: f(28_000), II: f(8800), III: f(1600), IV: f(1100), V: f(550), VI: f(11)},
	"A0": {Ia: f(130_000), Ib: f(22_000), II: f(8000), III: f(220), IV: f(140), V: f(43)},
	"A5": {Ia: f(120_000), Ib: f(20_000), II: f(7300), III: f(90), IV: f(33), V: f(15)},
	"F0": {Ia: f(120_000), Ib: f(20_000), II: f(7000), III: f(70), IV: f(25), V: f(8.1)},
	"F5": {Ia: f(120_000), Ib: f(20_000), II: f(6900), III: f(39), IV: f(6), V: f(3.5)},
	"G0": {Ia: f(120_000), Ib: f(20_000), II: f(6800), III: f(120), IV: f(10), V: f(1.4), VI: f(0.73)},
	"G5": {Ia: f(110_000), Ib: f(20_000), II: f(7000), III: f(200), IV: f(14), V: f(0.78), VI: f(0.43)},
	"K0": {Ia: f(110_000), Ib: f(21_000), II: f(7800), III: f(260), IV: f(23), V: f(0.52), VI: f(0.23)},
	"K5": {Ia: f(120_000), Ib: f(22_000), II: f(8400), III: f(530), V: f(0.21), VI: f(0.083)},
	"M0": {Ia: f(130_000), Ib: f(24_000), II: f(8800), III: f(600), V: f(0.082), VI: f(0.027)},
	"M5": {Ia: f(100_000), Ib: f(26_000), II: f(8800), III: f(720), V: f(0.0029), VI: f(0.00072)},
	"M9": {Ia: f(90_000), Ib: f(19_000), II: f(7300), III: f(1200), V: f(0.00029), VI: f(0.00019)},
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): Mass, Temperature, Diameter, Luminosity tables (WBH pp.17,19)"
```

---

## Task 10: Spectral-grid interpolation helper

**Files:**

- Create: `tools/world-builder/stars/physical.go`
- Create: `tools/world-builder/stars/physical_test.go`

The book's grid is irregular: O0, O5, B0, B5, A0, A5, F0, F5, G0, G5, K0, K5, M0, M5, M9. Every "0" / "5" / "9" subtype is tabulated; intermediate subtypes interpolate linearly between bracketing rows. The book treats K0 as if it were "G10" for interpolation: a G7 V is "2/5 of the way between G5 V and K0 V."

- [ ] **Step 1: Write failing tests**

`tools/world-builder/stars/physical_test.go`:

```go
package stars

import (
	"math"
	"testing"
)

func TestInterpolateClassRow_GridPoint(t *testing.T) {
	// G5 V is a tabulated grid point; interpolation should return the table value.
	got, err := InterpolateClassRow(StarMass, SpectralType{Letter: 'G', Subtype: 5}, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 0.9 {
		t.Fatalf("got %v want 0.9", got)
	}
}

func TestInterpolateClassRow_G7VMass(t *testing.T) {
	// WBH p. 17: G7 V mass = 0.9 + (2/5) * (0.8 - 0.9) = 0.86.
	got, err := InterpolateClassRow(StarMass, SpectralType{Letter: 'G', Subtype: 7}, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 0.86
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInterpolateClassRow_G7VDiameter(t *testing.T) {
	// WBH p. 18: G7 V diameter = 0.95 + (2/5) * (0.9 - 0.95) = 0.93.
	got, err := InterpolateClassRow(StarDiameter, SpectralType{Letter: 'G', Subtype: 7}, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 0.93
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInterpolateScalar_Temperature(t *testing.T) {
	// WBH p. 17: G7 V temperature = 5600 + 0.4 * (5200 - 5600) = 5440.
	got, err := InterpolateScalar(StarTemperature, SpectralType{Letter: 'G', Subtype: 7})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 5440.0
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInterpolateClassRow_A3III(t *testing.T) {
	// A3 III interpolates between A0 III (8) and A5 III (6) at 3/5.
	got, err := InterpolateClassRow(StarMass, SpectralType{Letter: 'A', Subtype: 3}, III)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 8 + (3.0/5.0)*(6-8)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInterpolateClassRow_M3V(t *testing.T) {
	// M3 V interpolates between M0 V (0.5) and M5 V (0.16) at 3/5.
	got, err := InterpolateClassRow(StarMass, SpectralType{Letter: 'M', Subtype: 3}, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 0.5 + (3.0/5.0)*(0.16-0.5)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInterpolateClassRow_MissingClass(t *testing.T) {
	// Class IV not tabulated for O0.
	_, err := InterpolateClassRow(StarMass, SpectralType{Letter: 'O', Subtype: 0}, IV)
	if err == nil {
		t.Fatal("expected error for missing class")
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run Interpolate
```

Expected: undefined symbols.

- [ ] **Step 3: Implement**

`tools/world-builder/stars/physical.go`:

```go
package stars

import "fmt"

// gridKeys lists the WBH-tabulated subtype grid in book order.
var gridKeys = []string{
	"O0", "O5",
	"B0", "B5",
	"A0", "A5",
	"F0", "F5",
	"G0", "G5",
	"K0", "K5",
	"M0", "M5", "M9",
}

// letterOrder maps each spectral letter to its position in book order.
var letterOrder = map[SpectralLetter]int{
	'O': 0, 'B': 1, 'A': 2, 'F': 3, 'G': 4, 'K': 5, 'M': 6,
}

// gridIndex returns (lowerIdx, upperIdx, fraction) for interpolation.
// Treats the grid as a flat ordered sequence: O0 < O5 < B0 < ... < M9.
// Subtypes between two grid points yield a fraction in [0, 1].
// Exact grid points yield (idx, idx, 0).
func gridIndex(st SpectralType) (lower, upper int, frac float64, err error) {
	lo, ok := letterOrder[st.Letter]
	if !ok {
		return 0, 0, 0, fmt.Errorf("stars: unsupported letter: %c", st.Letter)
	}
	target := lo*10 + st.Subtype

	// Compute numeric position of every grid key.
	positions := make([]int, len(gridKeys))
	for i, k := range gridKeys {
		positions[i] = letterOrder[SpectralLetter(k[0])]*10 + int(k[1]-'0')
	}

	for i, p := range positions {
		if p == target {
			return i, i, 0, nil
		}
	}

	lowerIdx, upperIdx := -1, -1
	for i, p := range positions {
		if p < target {
			lowerIdx = i
		}
		if p > target && upperIdx == -1 {
			upperIdx = i
		}
	}
	if lowerIdx == -1 || upperIdx == -1 {
		return 0, 0, 0, fmt.Errorf("stars: spectral type out of grid range: %s", st)
	}
	span := positions[upperIdx] - positions[lowerIdx]
	return lowerIdx, upperIdx, float64(target-positions[lowerIdx]) / float64(span), nil
}

// InterpolateClassRow interpolates a class-keyed quantity (Mass, Diameter,
// Luminosity) from a ClassRow table. Returns an error if the requested
// luminosity class is missing in either bracketing grid row.
func InterpolateClassRow(table map[string]ClassRow, st SpectralType, lc LuminosityClass) (float64, error) {
	lo, hi, frac, err := gridIndex(st)
	if err != nil {
		return 0, err
	}
	loRow, ok := table[gridKeys[lo]]
	if !ok {
		return 0, fmt.Errorf("stars: no row for %s", gridKeys[lo])
	}
	loVal, ok := loRow.Get(lc)
	if !ok {
		return 0, fmt.Errorf("stars: %s class %s missing", gridKeys[lo], lc)
	}
	if lo == hi {
		return loVal, nil
	}
	hiRow, ok := table[gridKeys[hi]]
	if !ok {
		return 0, fmt.Errorf("stars: no row for %s", gridKeys[hi])
	}
	hiVal, ok := hiRow.Get(lc)
	if !ok {
		return 0, fmt.Errorf("stars: %s class %s missing", gridKeys[hi], lc)
	}
	return loVal + frac*(hiVal-loVal), nil
}

// InterpolateScalar interpolates a scalar-valued quantity (Temperature)
// across the spectral grid.
func InterpolateScalar(table map[string]float64, st SpectralType) (float64, error) {
	lo, hi, frac, err := gridIndex(st)
	if err != nil {
		return 0, err
	}
	loVal, ok := table[gridKeys[lo]]
	if !ok {
		return 0, fmt.Errorf("stars: no row for %s", gridKeys[lo])
	}
	if lo == hi {
		return loVal, nil
	}
	hiVal, ok := table[gridKeys[hi]]
	if !ok {
		return 0, fmt.Errorf("stars: no row for %s", gridKeys[hi])
	}
	return loVal + frac*(hiVal-loVal), nil
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): spectral-grid interpolation helpers (WBH p.17 G7 V example)"
```

---

## Task 11: Mass / diameter / temperature lookups

**Files:**

- Modify: `tools/world-builder/stars/physical.go`
- Modify: `tools/world-builder/stars/physical_test.go`

- [ ] **Step 1: Append failing tests**

Append to `tools/world-builder/stars/physical_test.go`:

```go
func TestComputeMass_G7V(t *testing.T) {
	got, err := ComputeMass(SpectralType{Letter: 'G', Subtype: 7}, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-0.86) > 1e-9 {
		t.Fatalf("got %v want 0.86", got)
	}
}

func TestComputeDiameter_G7V(t *testing.T) {
	got, err := ComputeDiameter(SpectralType{Letter: 'G', Subtype: 7}, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-0.93) > 1e-9 {
		t.Fatalf("got %v want 0.93", got)
	}
}

func TestComputeTemperature_G7V(t *testing.T) {
	got, err := ComputeTemperature(SpectralType{Letter: 'G', Subtype: 7})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-5440) > 1e-9 {
		t.Fatalf("got %v want 5440", got)
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run "Compute"
```

Expected: undefined symbols.

- [ ] **Step 3: Append implementation**

Append to `tools/world-builder/stars/physical.go`:

```go
// ComputeMass returns the interpolated mass in solar units (WBH p. 17).
func ComputeMass(st SpectralType, lc LuminosityClass) (float64, error) {
	return InterpolateClassRow(StarMass, st, lc)
}

// ComputeDiameter returns the interpolated diameter in solar units (WBH p. 19).
func ComputeDiameter(st SpectralType, lc LuminosityClass) (float64, error) {
	return InterpolateClassRow(StarDiameter, st, lc)
}

// ComputeTemperature returns the interpolated surface temperature in
// Kelvin (WBH p. 17 — Temperature column).
func ComputeTemperature(st SpectralType) (float64, error) {
	return InterpolateScalar(StarTemperature, st)
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): mass, diameter, temperature lookups"
```

---

## Task 12: Luminosity table and formula

**Files:**

- Modify: `tools/world-builder/stars/physical.go`
- Modify: `tools/world-builder/stars/physical_test.go`

WBH p. 20 gives the closed-form luminosity: **L = (D/D⊙)² × (T/T⊙)⁴**, with T⊙ = 5772 K.

- [ ] **Step 1: Append failing tests**

Append to `tools/world-builder/stars/physical_test.go`:

```go
func TestComputeLuminosityFromTable_G0V(t *testing.T) {
	got, err := ComputeLuminosityFromTable(SpectralType{Letter: 'G', Subtype: 0}, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-1.4) > 1e-9 {
		t.Fatalf("got %v want 1.4", got)
	}
}

func TestComputeLuminosityFromFormula_Sol(t *testing.T) {
	got := ComputeLuminosityFromFormula(1.0, 5772)
	if math.Abs(got-1.0) > 1e-9 {
		t.Fatalf("got %v want 1.0", got)
	}
}

func TestComputeLuminosityFromFormula_Zed(t *testing.T) {
	// WBH p. 20: Zed (D=0.967, T=5440) -> L = (0.967)^2 * (5440/5772)^4 = 0.7378.
	got := ComputeLuminosityFromFormula(0.967, 5440)
	if math.Abs(got-0.7378) > 1e-3 {
		t.Fatalf("got %v want 0.7378", got)
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run Luminosity
```

Expected: undefined symbols.

- [ ] **Step 3: Append implementation**

Append to `tools/world-builder/stars/physical.go`:

```go
// SolTemperatureK is the WBH p. 20 reference temperature for Sol.
const SolTemperatureK = 5772.0

// ComputeLuminosityFromTable returns the interpolated luminosity in
// solar units (WBH p. 19).
func ComputeLuminosityFromTable(st SpectralType, lc LuminosityClass) (float64, error) {
	return InterpolateClassRow(StarLuminosity, st, lc)
}

// ComputeLuminosityFromFormula returns the closed-form luminosity in
// solar units (WBH p. 20):
//
//	L/L⊙ = (D/D⊙)^2 × (T/T⊙)^4
func ComputeLuminosityFromFormula(diameter, temperature float64) float64 {
	tRatio := temperature / SolTemperatureK
	return diameter * diameter * tRatio * tRatio * tRatio * tRatio
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): luminosity table lookup and physical formula"
```

---

## Task 13: Optional variance rolls

**Files:**

- Modify: `tools/world-builder/stars/physical.go`
- Modify: `tools/world-builder/stars/physical_test.go`

WBH p. 17: optional ±20% variance for mass and diameter via 2D-7. WBH p. 19: ±30% for luminosity. The factor scales linearly across [-5, +5]: `value × (1 + (2D-7)/5 × maxPct)`.

- [ ] **Step 1: Append failing tests**

Append to `tools/world-builder/stars/physical_test.go`:

```go
import (
	// ... existing imports ...
	"wbh/roller"
)

func TestApplyVariance_Zero(t *testing.T) {
	r := roller.NewScripted(0)
	got := ApplyVariance(0.86, r, 0.20)
	if got != 0.86 {
		t.Fatalf("got %v want 0.86", got)
	}
}

func TestApplyVariance_ZedMass(t *testing.T) {
	// WBH p. 17: base 0.86, 2D-7 = +2, max 0.20 -> 0.86 * (1 + 2/5 * 0.20) = 0.9288.
	r := roller.NewScripted(2)
	got := ApplyVariance(0.86, r, 0.20)
	want := 0.9288
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestApplyVariance_ZedDiameter(t *testing.T) {
	// WBH p. 18: base 0.93, 2D-7 = +1, max 0.20 -> 0.93 * (1 + 1/5 * 0.20) = 0.9672.
	r := roller.NewScripted(1)
	got := ApplyVariance(0.93, r, 0.20)
	want := 0.9672
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestApplyVariance_Negative(t *testing.T) {
	r := roller.NewScripted(-2)
	got := ApplyVariance(1.0, r, 0.20)
	want := 0.92
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}
```

Note: confirm the `_test.go` file imports `wbh/roller` — add to the import block at the top.

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run Variance
```

Expected: undefined `ApplyVariance`.

- [ ] **Step 3: Append implementation**

Append to `tools/world-builder/stars/physical.go`:

```go
import (
	// ... existing imports ...
	"wbh/roller"
)

// ApplyVariance applies the WBH optional variance roll (p. 17–19).
//
// The roll is 2D-7 (range -5 to +5). The result scales linearly between
// -maxPct and +maxPct of the base value:
//
//	adjusted = base × (1 + (2D-7)/5 × maxPct)
//
// Use maxPct=0.20 for mass/diameter, 0.30 for luminosity (Class III/V).
func ApplyVariance(base float64, r roller.Roller, maxPct float64) float64 {
	deviation := r.Roll("2D-7")
	factor := 1.0 + (float64(deviation)/5.0)*maxPct
	return base * factor
}
```

(Add `"wbh/roller"` to the imports at the top of `physical.go` if not already there.)

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): optional variance rolls for mass/diameter/luminosity"
```

---

## Task 14: Peculiar / special object kind dispatch

**Files:**

- Create: `tools/world-builder/stars/peculiar.go`
- Create: `tools/world-builder/stars/peculiar_test.go`

WBH pp. 16, 22: a "Special" primary roll leads through Unusual or Peculiar columns. The book offers a simple resolution path (1D: 1-5 = neutron star, 6 = black hole). Plan 1 implements the simple path; the full Unusual/Peculiar dispatch lands in Plan 2.

- [ ] **Step 1: Write failing tests**

`tools/world-builder/stars/peculiar_test.go`:

```go
package stars

import (
	"testing"

	"wbh/roller"
)

func TestKindFromUnusualCell(t *testing.T) {
	cases := map[string]StarKind{
		"BD": KindBrownDwarf,
		"D":  KindWhiteDwarf,
	}
	for cell, want := range cases {
		got, err := KindFromUnusualCell(cell)
		if err != nil {
			t.Fatalf("%s error: %v", cell, err)
		}
		if got != want {
			t.Fatalf("%s = %v want %v", cell, got, want)
		}
	}
}

func TestKindFromPeculiarCell(t *testing.T) {
	cases := map[string]StarKind{
		"Black Hole":    KindBlackHole,
		"Pulsar":        KindPulsar,
		"Neutron Star":  KindNeutronStar,
		"Nebula":        KindNebula,
		"Protostar":     KindProtostar,
		"Star Cluster":  KindStarCluster,
		"Anomaly":       KindAnomaly,
	}
	for cell, want := range cases {
		got, err := KindFromPeculiarCell(cell)
		if err != nil {
			t.Fatalf("%s error: %v", cell, err)
		}
		if got != want {
			t.Fatalf("%s = %v want %v", cell, got, want)
		}
	}
}

func TestRollSpecialPrimary_Simple(t *testing.T) {
	// 1D=3 -> Neutron Star, 1D=6 -> Black Hole.
	r := roller.NewScripted(3)
	got, err := RollSpecialPrimarySimple(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != KindNeutronStar {
		t.Fatalf("got %v want neutron star", got)
	}

	r2 := roller.NewScripted(6)
	got2, err := RollSpecialPrimarySimple(r2)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got2 != KindBlackHole {
		t.Fatalf("got %v want black hole", got2)
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run "Kind|Special"
```

Expected: undefined symbols.

- [ ] **Step 3: Implement**

`tools/world-builder/stars/peculiar.go`:

```go
package stars

import (
	"fmt"

	"wbh/roller"
)

// KindFromUnusualCell maps an Unusual-column cell from the Star Type
// Determination table (WBH p. 15) to a StarKind.
func KindFromUnusualCell(cell string) (StarKind, error) {
	switch cell {
	case "BD":
		return KindBrownDwarf, nil
	case "D":
		return KindWhiteDwarf, nil
	default:
		return "", fmt.Errorf("stars: unknown Unusual cell: %q", cell)
	}
}

// KindFromPeculiarCell maps a Peculiar-column cell from the Star Type
// Determination table (WBH p. 15) to a StarKind.
func KindFromPeculiarCell(cell string) (StarKind, error) {
	switch cell {
	case "Black Hole":
		return KindBlackHole, nil
	case "Pulsar":
		return KindPulsar, nil
	case "Neutron Star":
		return KindNeutronStar, nil
	case "Nebula":
		return KindNebula, nil
	case "Protostar":
		return KindProtostar, nil
	case "Star Cluster":
		return KindStarCluster, nil
	case "Anomaly":
		return KindAnomaly, nil
	default:
		return "", fmt.Errorf("stars: unknown Peculiar cell: %q", cell)
	}
}

// RollSpecialPrimarySimple resolves a "Special" primary roll using the
// simple Referee path described on WBH p. 16:
//
//	1D: 1-5 -> Neutron Star, 6 -> Black Hole
//
// The full Unusual/Peculiar dispatch (which can produce brown dwarfs,
// nebulae, protostars, star clusters, anomalies, and so on) lands in
// Plan 2 alongside the rest of the multi-star pipeline.
func RollSpecialPrimarySimple(r roller.Roller) (StarKind, error) {
	roll := r.Roll("1D")
	if roll == 6 {
		return KindBlackHole, nil
	}
	return KindNeutronStar, nil
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): peculiar/special object kind dispatch (WBH p.16)"
```

---

## Task 15: Main sequence lifespan and small-star age

**Files:**

- Create: `tools/world-builder/stars/ages.go`
- Create: `tools/world-builder/stars/ages_test.go`

WBH p. 20:

- **Main Sequence Lifespan** = 10 / Mass^2.5 (Gyr)
- **Small Star Age** = 1D × 2 + D3 - 1 Gyr (accuracy=1)
- **Small Star Age** = 1D × 2 + D3 - 2 + d10/10 Gyr (accuracy=2)

- [ ] **Step 1: Write failing tests**

`tools/world-builder/stars/ages_test.go`:

```go
package stars

import (
	"math"
	"testing"

	"wbh/roller"
)

func TestMainSequenceLifespan_Sol(t *testing.T) {
	got := MainSequenceLifespan(1.0)
	if math.Abs(got-10.0) > 1e-9 {
		t.Fatalf("got %v want 10.0", got)
	}
}

func TestMainSequenceLifespan_Zed(t *testing.T) {
	// WBH p. 21: 0.929^-2.5 * 10 = 12.022 Gyr.
	got := MainSequenceLifespan(0.929)
	if math.Abs(got-12.022) > 1e-2 {
		t.Fatalf("got %v want ~12.022", got)
	}
}

func TestSmallStarAge_Basic(t *testing.T) {
	// 1D=3, D3=1 -> 3*2 + 1 - 1 = 6 Gyr.
	r := roller.NewScripted(3, 1)
	got, err := SmallStarAge(r, 1)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-6.0) > 1e-9 {
		t.Fatalf("got %v want 6.0", got)
	}
}

func TestSmallStarAge_Zed(t *testing.T) {
	// WBH p. 21: 1D=3, D3=2 -> 3*2 + 2 - 1 = 7 Gyr.
	r := roller.NewScripted(3, 2)
	got, err := SmallStarAge(r, 1)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-7.0) > 1e-9 {
		t.Fatalf("got %v want 7.0", got)
	}
}

func TestSmallStarAge_Accuracy2(t *testing.T) {
	// WBH p. 21: 1D=3, D3=2, d10=3 -> 3*2 + 2 - 2 + 0.3 = 6.3 Gyr.
	r := roller.NewScripted(3, 2, 3)
	got, err := SmallStarAge(r, 2)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-6.3) > 1e-9 {
		t.Fatalf("got %v want 6.3", got)
	}
}

func TestSmallStarAge_InvalidAccuracy(t *testing.T) {
	r := roller.NewScripted(3, 2)
	if _, err := SmallStarAge(r, 0); err == nil {
		t.Fatal("expected error for accuracy 0")
	}
	r2 := roller.NewScripted(3, 2)
	if _, err := SmallStarAge(r2, 3); err == nil {
		t.Fatal("expected error for accuracy 3")
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run "Lifespan|SmallStarAge"
```

Expected: undefined symbols.

- [ ] **Step 3: Implement**

`tools/world-builder/stars/ages.go`:

```go
package stars

import (
	"fmt"
	"math"

	"wbh/roller"
)

// MainSequenceLifespan returns the main-sequence lifespan in Gyr (WBH p. 20):
//
//	Lifespan = 10 / Mass^2.5
func MainSequenceLifespan(mass float64) float64 {
	return 10.0 / math.Pow(mass, 2.5)
}

// SmallStarAge returns a random age for a small star in Gyr (WBH p. 21).
//
// accuracy=1: 1D × 2 + D3 - 1
// accuracy=2: 1D × 2 + D3 - 2 + d10/10
//
// Higher accuracy is referenced in the book ("Adding additional digits
// of accuracy requires additional d10 rolls"). Only 1 and 2 are
// implemented here.
func SmallStarAge(r roller.Roller, accuracy int) (float64, error) {
	if accuracy != 1 && accuracy != 2 {
		return 0, fmt.Errorf("stars: accuracy must be 1 or 2, got %d", accuracy)
	}
	oneD := r.Roll("1D")
	d3 := r.Roll("D3")
	if accuracy == 1 {
		return float64(oneD*2 + d3 - 1), nil
	}
	d10 := r.Roll("d10")
	return float64(oneD*2+d3-2) + float64(d10)/10.0, nil
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): main sequence lifespan and small-star age (WBH pp.20-21)"
```

---

## Task 16: Subgiant / giant / final-age formulas

**Files:**

- Modify: `tools/world-builder/stars/ages.go`
- Modify: `tools/world-builder/stars/ages_test.go`

WBH pp. 21–22:

- **Subgiant Lifespan** = MSL / (4 + Mass)
- **Giant Lifespan** = MSL / (10 × Mass^3)
- **Star Final Age** = (10 / Mass^2.5) × (1 + 1/(4+Mass) + 1/(10×Mass^3))

- [ ] **Step 1: Append failing tests**

Append to `tools/world-builder/stars/ages_test.go`:

```go
func TestSubgiantLifespan_Zed(t *testing.T) {
	got := SubgiantLifespan(12.022, 0.929)
	want := 12.022 / (4 + 0.929)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestGiantLifespan_Zed(t *testing.T) {
	got := GiantLifespan(12.022, 0.929)
	want := 12.022 / (10 * math.Pow(0.929, 3))
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestFinalAgeProgenitor_WhiteDwarf(t *testing.T) {
	// WBH p. 30: progenitor mass 1.47 -> 4.635 Gyr.
	got := FinalAgeProgenitor(1.47)
	if math.Abs(got-4.635) > 1e-2 {
		t.Fatalf("got %v want ~4.635", got)
	}
}

func TestFinalAgeProgenitor_UnitMass(t *testing.T) {
	// At mass=1.0, final age = 10 * (1 + 1/5 + 1/10) = 13.0.
	got := FinalAgeProgenitor(1.0)
	if math.Abs(got-13.0) > 1e-9 {
		t.Fatalf("got %v want 13.0", got)
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run "Lifespan|FinalAge"
```

Expected: undefined symbols.

- [ ] **Step 3: Append implementation**

Append to `tools/world-builder/stars/ages.go`:

```go
// SubgiantLifespan returns the subgiant phase lifespan in Gyr (WBH p. 21):
//
//	Subgiant Lifespan = Main Sequence Lifespan / (4 + Mass)
func SubgiantLifespan(mainSequenceLifespanGyr, mass float64) float64 {
	return mainSequenceLifespanGyr / (4.0 + mass)
}

// GiantLifespan returns the giant phase lifespan in Gyr (WBH p. 22):
//
//	Giant Lifespan = Main Sequence Lifespan / (10 × Mass^3)
func GiantLifespan(mainSequenceLifespanGyr, mass float64) float64 {
	return mainSequenceLifespanGyr / (10.0 * math.Pow(mass, 3))
}

// FinalAgeProgenitor returns the total elapsed lifespan of a star up to
// its post-stellar transition (WBH p. 22):
//
//	Star Final Age = (10 / Mass^2.5) × (1 + 1/(4+Mass) + 1/(10×Mass^3))
//
// progenitorMass is the original star's mass (NOT dead-star mass). For
// post-stellar age computation, multiply dead-star mass by (2 + D3) per
// the book and pass that here.
func FinalAgeProgenitor(progenitorMass float64) float64 {
	msl := MainSequenceLifespan(progenitorMass)
	return msl * (1.0 +
		1.0/(4.0+progenitorMass) +
		1.0/(10.0*math.Pow(progenitorMass, 3)))
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test -race ./stars/...
```

- [ ] **Step 5: Lint, format, commit**

```bash
gofumpt -l -w stars
golangci-lint run ./stars/...
git -C /Users/markayers/Documents/Traveller add tools/world-builder/stars
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): subgiant, giant, and post-stellar final-age formulas"
```

---

## Task 17: Public API + Sol/Terra and Zed worked examples

**Files:**

- Create: `tools/world-builder/stars/stars.go`
- Create: `tools/world-builder/stars/worked_examples_test.go`

This task ties the pieces together into a public API and proves it works against the simplest worked example: Sol/Terra (WBH p. 35), which is fully specified.

The Sol/Terra row on p. 35:

```text
Sol  G2 V  Mass 1.000  Temp 5,772  Diameter 1.000  Luminosity 1.000  Age 4.568
```

- [ ] **Step 1: Write failing tests**

`tools/world-builder/stars/worked_examples_test.go`:

```go
package stars_test

import (
	"math"
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestSolTerra_p35(t *testing.T) {
	// WBH p. 35 — Terra/Sol example (fully specified, not rolled).
	sol := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.000,
		Diameter:        1.000,
		Temperature:     5772,
		AgeGyr:          4.568,
	})
	if sol.SpectralType != (stars.SpectralType{Letter: 'G', Subtype: 2}) {
		t.Fatalf("spectral type wrong: %v", sol.SpectralType)
	}
	if sol.LuminosityClass != stars.V {
		t.Fatalf("class wrong: %v", sol.LuminosityClass)
	}
	checks := []struct {
		name string
		got  float64
		want float64
	}{
		{"mass", sol.Mass, 1.0},
		{"diameter", sol.Diameter, 1.0},
		{"temperature", sol.Temperature, 5772},
		{"luminosity", sol.Luminosity, 1.0},
		{"age", sol.AgeGyr, 4.568},
	}
	for _, c := range checks {
		if math.Abs(c.got-c.want) > 1e-9 {
			t.Errorf("%s: got %v want %v", c.name, c.got, c.want)
		}
	}
}

func TestZedPrimaryOnly_p17_p21(t *testing.T) {
	// WBH pp. 16–21 — Zed (G7 V) primary star, no companions.
	// Drive rolls verbatim from the book:
	//   2D=9 -> "G" type
	//   2D=6 -> Numeric subtype 7 (G7)
	//   2D-7=+2 mass variance -> 0.929
	//   2D-7=+1 diameter variance -> 0.967
	//   1D=3, D3=2, d10=3 -> 6.3 Gyr
	r := roller.NewScripted(
		9,  // primary type 2D
		6,  // subtype 2D
		2,  // mass variance 2D-7
		1,  // diameter variance 2D-7
		3,  // age 1D
		2,  // age D3
		3,  // age d10
	)
	star, err := stars.GenerateMainSequenceStar(r, stars.GenerateOpts{
		WithVariance: true,
		Accuracy:     2,
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if star.SpectralType != (stars.SpectralType{Letter: 'G', Subtype: 7}) {
		t.Fatalf("spectral type: got %v want G7", star.SpectralType)
	}
	if star.LuminosityClass != stars.V {
		t.Fatalf("class: got %v want V", star.LuminosityClass)
	}
	checks := []struct {
		name string
		got  float64
		want float64
		tol  float64
	}{
		{"mass", star.Mass, 0.929, 2e-3},
		{"diameter", star.Diameter, 0.967, 2e-3},
		{"temperature", star.Temperature, 5440, 2e-3},
		{"luminosity", star.Luminosity, 0.738, 2e-3},
		{"age", star.AgeGyr, 6.3, 1e-9},
	}
	for _, c := range checks {
		if math.Abs(c.got-c.want) > c.tol {
			t.Errorf("%s: got %v want %v (tol %v)", c.name, c.got, c.want, c.tol)
		}
	}
}
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
go test ./stars/... -run "SolTerra|ZedPrimary"
```

Expected: undefined `stars.Compose`, `stars.GenerateMainSequenceStar`.

- [ ] **Step 3: Implement public API**

`tools/world-builder/stars/stars.go`:

```go
package stars

import (
	"errors"
	"fmt"

	"wbh/roller"
)

// ComposeOpts holds the inputs to Compose.
type ComposeOpts struct {
	Kind            StarKind
	SpectralType    SpectralType
	LuminosityClass LuminosityClass
	Mass            float64
	Diameter        float64
	Temperature     float64
	AgeGyr          float64
}

// Compose constructs a Star from explicit physical values.
//
// Luminosity is derived from the WBH p. 20 formula:
//
//	L = (D/D⊙)^2 × (T/T⊙)^4
func Compose(o ComposeOpts) Star {
	return Star{
		Kind:            o.Kind,
		SpectralType:    o.SpectralType,
		LuminosityClass: o.LuminosityClass,
		Mass:            o.Mass,
		Diameter:        o.Diameter,
		Temperature:     o.Temperature,
		Luminosity:      ComputeLuminosityFromFormula(o.Diameter, o.Temperature),
		AgeGyr:          o.AgeGyr,
	}
}

// GenerateOpts controls GenerateMainSequenceStar.
type GenerateOpts struct {
	// WithVariance applies the optional ±20% variance roll for mass and
	// diameter (WBH p. 17). Off by default. Temperature variance is
	// intentionally omitted in Plan 1: WBH p. 17 leaves the scale of
	// temperature variance to the Referee without specifying a formula.
	WithVariance bool

	// Accuracy is 1 or 2. 1 uses 1D × 2 + D3 - 1 for age; 2 adds a d10
	// for an additional digit of accuracy.
	Accuracy int
}

// ErrNonClassVPrimary is returned by GenerateMainSequenceStar when the
// primary roll yields a non-Class-V star. Plan 1 only handles Class V;
// Plan 2 introduces full class dispatch.
var ErrNonClassVPrimary = errors.New("stars: non-Class-V primary; arrives in Plan 2")

// GenerateMainSequenceStar generates a Class V main-sequence primary
// star from rolls.
//
// Roll order (consumed from the roller in this order):
//  1. 2D for Star Type Determination (Type column)
//  2. 2D for Star Subtype
//  3. (if WithVariance) 2D-7 for mass variance
//  4. (if WithVariance) 2D-7 for diameter variance
//  5. 1D for age
//  6. D3 for age
//  7. (if Accuracy == 2) d10 for age
func GenerateMainSequenceStar(r roller.Roller, opts GenerateOpts) (Star, error) {
	letter, lc, err := RollPrimaryTypeAndClass(r)
	if err != nil {
		return Star{}, err
	}
	if lc != V {
		return Star{}, fmt.Errorf("%w: got %s", ErrNonClassVPrimary, lc)
	}
	subtype, err := RollSubtype(r, letter, lc)
	if err != nil {
		return Star{}, err
	}
	st := SpectralType{Letter: letter, Subtype: subtype}

	mass, err := ComputeMass(st, lc)
	if err != nil {
		return Star{}, err
	}
	diameter, err := ComputeDiameter(st, lc)
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

	age, err := SmallStarAge(r, opts.Accuracy)
	if err != nil {
		return Star{}, err
	}

	return Star{
		Kind:            KindMainSequence,
		SpectralType:    st,
		LuminosityClass: lc,
		Mass:            mass,
		Diameter:        diameter,
		Temperature:     temperature,
		Luminosity:      luminosity,
		AgeGyr:          age,
	}, nil
}
```

- [ ] **Step 4: Run all tests**

```bash
cd /Users/markayers/Documents/Traveller/tools/world-builder
just test
just check
```

Expected: all tests pass; check is clean.

- [ ] **Step 5: Commit**

```bash
git -C /Users/markayers/Documents/Traveller add tools/world-builder
git -C /Users/markayers/Documents/Traveller commit -m "feat(stars): public API + Sol/Terra and Zed primary worked examples"
```

---

## Plan 1 complete

After Task 17 the wbh module can:

- Roll a Class V primary star from a `roller.Roller`.
- Look up mass, diameter, temperature, luminosity (table or formula) for any tabulated combination.
- Apply optional variance.
- Compute small-star ages with two accuracy levels.
- Construct a Sol/Terra-equivalent star directly via `Compose`.

The Sol/Terra and Zed-primary worked examples pass to the digit (within stated tolerances).

**Next:** Plan 2 will add multi-star presence, secondaries, stellar orbits, eccentricity, inclination, designations, and the IISS Class 0/I survey form output. After Plan 2, the full Zed quintuple-system regression test on WBH p. 34 becomes the v1 acceptance gate.

---

## Self-review notes

Spec coverage check (against `2026-05-02-world-builder-design.md` v1 list):

- ✅ Primary star generation (Type Determination): Tasks 5, 7.
- ✅ Star Subtype: Tasks 6, 7.
- ✅ Class restrictions IV/VI/III+: Task 8.
- ✅ Mass/diameter/temperature/luminosity (table + formula): Tasks 9, 10, 11, 12.
- ✅ Variance: Task 13.
- ✅ Peculiar/special objects (kind dispatch): Task 14.
- ✅ System ages — small-star + main-sequence/subgiant/giant: Tasks 15, 16.
- ⏳ Multi-star presence, secondaries, orbits, eccentricity, inclination, designations, survey form: deferred to Plan 2.
- ⏳ Special and Unusual Object Age table: deferred to Plan 2 (Zed test needs this for the white-dwarf companion).
