# World Builder Walkthrough

_2026-05-12T17:23:13Z by Showboat 0.6.1_
<!-- showboat-id: e1831189-dd6b-47d5-a0e2-70e2118aa7fd -->

## Overview

This project generates complete Mongoose Traveller _World Builder's Handbook_ star systems end-to-end. A seed plus deterministic dice produces a real, fully-formed system — primary star, companions, planets, moons, belts, atmospheres, life, habitability — rendered as Markdown IISS Class 0/I + II/III + IV-P forms.

WBH pp.14–146 (Stars + System Worlds and Orbits + World Physical Characteristics) are implemented to book fidelity. WBH pp.147+ (World Social Characteristics, Special Circumstances) are out of scope.

This walkthrough follows the code top-down: from the CLI entry, through the deterministic roller, into stars/worlds/iiss pipelines, ending with the rendered Markdown.

```bash
find . -maxdepth 2 -type d \( -name "cmd" -o -name "stars" -o -name "worlds" -o -name "iiss" -o -name "dice" -o -name "roller" -o -name "docs" -o -path "*/cmd/world-builder" \) | sort
```

```output
./cmd
./cmd/world-builder
./dice
./docs
./iiss
./roller
./stars
./worlds
```

## Package layout

Six Go packages plus the CLI. Dependencies point inward: `cmd/world-builder` imports `worlds` and `iiss`; `worlds` imports `stars`; `stars` and `worlds` import `roller`; `roller` imports `dice`. `iiss` does not import `worlds`.

- `dice/` — WBH dice notation parser (`"2D"`, `"2D-7"`, `"D3-1"`).
- `roller/` — `Roller` interface with `Seeded` (production), `Scripted` (test gold), and `Fixed` impls. The only RNG seam in the project.
- `stars/` — WBH pp.14–35: stellar generation, MAO, companion orbits.
- `worlds/` — WBH pp.36–146: placement, per-body procedures, the climate solver, mainworld pick, Notable Features.
- `iiss/` — Pure renderer package. Holds the IISS form types and `MarkdownSystem`. Does not import `worlds/`.
- `cmd/world-builder/` — CLI entry; three lines of pipeline plus format dispatch.

## Entry point: cmd/world-builder/main.go

The CLI takes `-seed`, `-format`, and dispatches. Three branches: `markdown` (default), `json`, `short`.

```bash
sed -n '18,65p' cmd/world-builder/main.go
```

```output
func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("world-builder", flag.ContinueOnError)
	fs.SetOutput(stderr)
	seed := fs.Int64("seed", 0, "random seed (0 = time-based)")
	format := fs.String("format", "markdown", "output format: markdown | json | short")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s := *seed
	if s == 0 {
		s = time.Now().UnixNano()
	}

	u, err := worlds.Generate(s)
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	switch *format {
	case "markdown":
		_, err := fmt.Fprint(stdout, iiss.MarkdownSystem(u.Detail.SystemForms))
		return err
	case "json":
		// Emit the full SystemForms aggregate (Class0I + Class23 + Class4P
		// plus profile strings and mainworld designation) so downstream
		// tooling has everything in one document. Per docs/next-steps.md
		// § B3.
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(u.Detail.SystemForms); err != nil {
			return fmt.Errorf("json: %w", err)
		}
		return nil
	case "short":
		_, err := fmt.Fprintln(stdout, u.Detail.ShortProfile)
		return err
	default:
		return fmt.Errorf("unknown format: %q (want markdown, json, or short)", *format)
	}
}
```

## The deterministic roller

`roller.Roller` is the single seam through which every dice roll passes. There are no package-level RNG calls anywhere; a seed plus the sequence of options fully determines a system.

Three implementations: `Seeded` (production), `Scripted` (test fixtures replay book-narrated dice), `Fixed` (single value).

```bash
sed -n '1,40p' roller/roller.go
```

```output
// Package roller provides dice-rolling abstractions used throughout the world-builder module.
//
// Every random draw in the library passes through a Roller. This makes
// seeded reproducibility and scripted-test injection both straightforward.
package roller

import (
	"fmt"
	"math/rand"

	"github.com/philoserf/world-builder/dice"
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
```

## worlds.Generate — the top-level pipeline

`worlds.Generate(seed)` wraps `GenerateWithRoller(r)`, which runs every stage in dependency-graph order. The function body is the pipeline at a glance.

```bash
sed -n '13,60p' worlds/generate.go
```

```output
func Generate(seed int64) (Universe, error) {
	return GenerateWithRoller(roller.NewSeeded(seed))
}

// GenerateWithRoller runs the entire pass-2 pipeline against any
// Roller. Tests use it with a Scripted roller for narrow per-procedure
// fixtures and with a Seeded roller for façade end-to-end fixtures
// (per docs/harness.md § Façade end-to-end). cmd/world-builder and
// Generate use it via the seed convenience.
//
// All other entry points (GenerateSystem, GenerateSystemPlacement,
// individual Apply* stages) remain available for callers that need
// finer control.
func GenerateWithRoller(r roller.Roller) (Universe, error) {
	sys, err := stars.GenerateSystem(r, stars.GenerateSystemOpts{
		WithVariance: true,
		Accuracy:     2,
		MAO:          MAO,
	})
	if err != nil {
		return Universe{}, fmt.Errorf("worlds: stars: %w", err)
	}
	sp, err := GenerateSystemPlacement(r, sys)
	if err != nil {
		return Universe{}, fmt.Errorf("worlds: placement: %w", err)
	}
	u := Universe{System: sys, Placement: sp}
	for _, step := range []func(roller.Roller, *Universe) error{
		ApplyDetailFrontEnd,
		ApplyBodyPhysical,
		ApplyBeltDetails,
		ApplyMoonRefinement,
		ApplyRotationTilt,
		ApplyClimate,
		ApplyTaintTypology,
		ApplySurfaceDistribution,
		ApplyGeology,
		ApplyBiology,
	} {
		if err := step(r, &u); err != nil {
			return u, err
		}
	}
	ApplyHabitability(&u)
	AggregateSystem(&u)
	BuildIISSForms(&u)
	return u, nil
}
```

## Stage 0: stars.GenerateSystem

The first stage rolls the primary star and all companions. Roll order:

1. Primary type (`RollPrimaryTypeAndClass`)
2. Subtype
3. Mass / Diameter / Temperature (table lookup, optional variance)
4. Six presence rolls for Close, Near, Far, and companions
5. Companion star generation in book order

The `Special` (2D=2) roll dispatches through the WBH p.15 Special column (the cleaner Referee setting) which yields Class VI / IV / III / Giants only — no BD/D/Peculiar primaries by default. This is what makes every seed produce a real system. Users can opt into the Unusual column via `GenerateSystemOpts.PeculiarColumn = PeculiarPathUnusual`.

```bash
sed -n '32,55p' stars/system.go
```

```output
//
// Roll consumption order (P2-10):
//  1. Primary star via GenerateMainSequenceStar.
//  2. Six presence rolls (2D each): Close, Near, Far, Primary companion,
//     then Close companion (if Close present), Near companion (if Near
//     present), Far companion (if Far present).
//  3. Star generation for each present non-primary slot in book order:
//     primary companion → Close → Close companion → Near → Near companion
//     → Far → Far companion.
//  4. Orbital placement for each companion in generation order: orbit#,
//     eccentricity, inclination.
//  5. AssignDesignations.
func GenerateSystem(r roller.Roller, opts GenerateSystemOpts) (System, error) {
	primary, err := GenerateMainSequenceStar(r, GenerateOpts{WithVariance: opts.WithVariance, Accuracy: opts.Accuracy})
	if errors.Is(err, ErrSpecialPrimary) {
		primary, err = generateSpecialPrimary(r, opts)
		if err != nil {
			return System{}, fmt.Errorf("special primary: %w", err)
		}
	} else if err != nil {
		return System{}, fmt.Errorf("primary: %w", err)
	}

	// Presence rolls: Close, Near, Far, primary companion first; then
```

## Stage 1: worlds.GenerateSystemPlacement

Allocates orbit slots to stars, fixes the baseline orbit, and places bodies into slots without yet sizing them. WBH pp.36–52: counts → available orbits → allocations → baseline → spread → slots → anomalous → place → eccentricities.

```bash
grep -n '^func GenerateSystemPlacement' worlds/system_placement.go
```

```output
43:func GenerateSystemPlacement(r roller.Roller, sys stars.System) (SystemPlacement, error) {
```

```bash
sed -n '43,75p' worlds/system_placement.go
```

```output
func GenerateSystemPlacement(r roller.Roller, sys stars.System) (SystemPlacement, error) {
	// TODO(continuation-method): when stars.System gains a pre-existing
	// mainworld field, return ErrContinuationMethodUnsupported here before
	// running the clean-slate pipeline.
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
	baselineN := RollBaselineNumber(r, sys, counts)
	primary := allocs[0].Group
	baselineOrbit := BaselineOrbit(r, primary, primary.HZCO(), baselineN, counts.Total)
	emptyOrbits, err := RollEmptyOrbits(r)
	if err != nil {
		return SystemPlacement{}, fmt.Errorf("worlds: empty-orbits: %w", err)
	}
	totalStars := 1 + secondaryStarCount(sys)
	spread := Spread(primary, allocs[0].AllocatedWorlds, baselineOrbit, baselineN, totalStars)
	slots, err := PlaceOrbitSlots(r, allocs, baselineN, baselineOrbit, spread, emptyOrbits)
	if err != nil {
```

## The unified Body type

Every placed object — planet, moon, belt member — is a single `Body` value. Moons are `Body{Kind: BodyMoon, Parent: <planet>}`. A unified iterator (`Universe.AllBodies()`) yields every body, so the "moon path silent-zero" anti-pattern (procedure runs for planets, forgets moons) is prevented at the type level.

```bash
sed -n '31,85p' worlds/body.go
```

```output
type Body struct {
	// Identity
	Designation string
	Kind        BodyKind
	Parent      *Body // nil for planets and belts; non-nil for moons

	// Stage 1 (placement) — for planets and belts; moons inherit via Parent.
	Group        Group // Cycle 0 uses worlds.Group; api-surface.md § Open
	Orbit        float64
	Eccentricity float64
	HZ           bool

	// Stage 2 (sizing/period)
	SizeCode       SizeCode
	DiameterKm     float64
	GGClass        GasGiantClass // NotGasGiant for non-GG bodies
	GGDiameterCode string        // GG only — eHex code matching DiameterEarth
	DiameterEarth  float64       // GG only
	MassEarth      float64
	Period         Period

	// Moon-specific (Kind == BodyMoon)
	OrbitPD     float64
	OrbitKm     float64
	PeriodHours float64
	Retrograde  bool // moon orbits its parent retrograde (anomalous slot)

	// Stage 3
	Physical *BodyPhysical
	Belt     *BeltDetails

	// Stage 4
	DayLength    *DayLength
	AxialTilt    *AxialTilt
	TidalLock    *TidalLock
	TidalEffects *SurfaceTidalEffects

	// Stage 5 — populated post-ApplyClimatePasses
	Atmosphere    *Atmosphere
	Hydrographics *Hydrographics
	Temperature   *Temperature

	// Stage 6
	SurfaceDistribution *SurfaceDistribution

	// Stages 7–9
	Geology      *Geology
	Biology      *Biology
	Habitability *Habitability

	// Sub-bodies (moons). Iterator descends into Children after the parent.
	Children []*Body
}

// HasPhysical reports whether body physical (composition / density /
```

## The climate solver: ApplyClimatePasses

Stage 5 is the only cyclic cluster: atmosphere ↔ temperature ↔ hydrographics depend on each other. `ApplyClimatePasses` runs **two passes** and trusts the second. It is not a fixed-point solver — `RederiveAtmosphereHydrographics` consumes fresh dice each pass, so each pass is a stochastic sample, not a convergence step. There is no fixed point to reach; the function is named for what it does.

```bash
sed -n '109,162p' worlds/stage5.go
```

```output
// ApplyClimatePasses is the per-body climate solver. Folds partial
// geology (Residual + TSF + THF) into each pass so Temperature
// includes the WBH p.125 inherent-temperature addition.
//
// Each pass:
//
//  1. Compute Temperature from current atm/hydro.
//  2. Compute partial-geology (atm/hydro-independent).
//  3. Apply TSS via T' = ⁴√(T⁴ + TSS⁴); refresh ScaleHeight.
//  4. Rederive atm/hydro from post-TSS Temperature.
//
// Runs exactly 2 passes (matching pass-1). The climate cluster is not
// a fixed point — RederiveAtmosphereHydrographics consumes fresh dice
// each call, so each pass is a stochastic sample, not a convergence
// step. The second sample is trusted.
//
// No-op for ineligible bodies (non-HZ, atmosphereless, gas giants,
// belts). For HZ bodies, body.Geology is also populated with the
// final TSS factors (without TectonicPlates — that's Stage 7).
func ApplyClimatePasses(r roller.Roller, body *Body, sys stars.System) error {
	atmo, eligible, err := initialAtmosphere(r, body, sys.Primary.AgeGyr)
	if err != nil {
		return err
	}
	if !eligible {
		return nil
	}
	body.Atmosphere = &atmo

	host := body
	if body.Kind == BodyMoon && body.Parent != nil {
		host = body.Parent
	}
	hzco := host.Group.HZCO()
	tempRange := HZCOOffsetToTempRange(host.Orbit, hzco)

	hydro, herr := GenerateHydrographics(r, atmo, body.SizeCode, tempRange)
	if herr != nil {
		return fmt.Errorf("hydro: %w", herr)
	}
	body.Hydrographics = &hydro

	parent := body.Parent

	// Pass 1.
	if err := climatePass(r, body, sys, parent, 1); err != nil {
		return err
	}
	// Pass 2 (final — matches pass-1's behaviour).
	if err := climatePass(r, body, sys, parent, 2); err != nil {
		return err
	}
	return nil
}
```

## Mainworld pick and aggregation

After all per-body stages, `AggregateSystem` picks the mainworld and computes system-wide profiles. The pick priority is HZ-terrestrial → HZ-moon → habitable rocky → any rocky/moon/belt. `BuildIISSForms` then translates `Universe` state into the three IISS form structs.

```bash
grep -n '^func pickMainworld\|^func AggregateSystem\|^func BuildIISSForms' worlds/*.go
```

```output
worlds/iiss_build.go:16:func BuildIISSForms(u *Universe) {
worlds/stage10.go:14:func AggregateSystem(u *Universe) {
worlds/stage10.go:153:func pickMainworld(u *Universe) (string, *Body) {
```

```bash
sed -n '14,40p' worlds/iiss_build.go
```

```output
// pp.141-142. Pure function — no rolls. Called by GenerateWithRoller
// after AggregateSystem.
func BuildIISSForms(u *Universe) {
	c0 := buildClass0I(u)
	u.Detail.Class0I = c0
	u.Detail.Class23 = buildClass23(u, c0)
	u.Detail.Class4P = buildClass4P(u, c0.FormHeader)
	u.Detail.NotableFeatures = NotableFeatures(u)
}

func buildClass0I(u *Universe) iiss.Class0IForm {
	// Delegate to pass-1's stars.BuildSurveyForm for full companion +
	// composite-barycentre fidelity, then translate to the iiss/
	// boundary type.
	meta := stars.SurveyMetadata{
		Sector:      "—",
		Location:    "—",
		Designation: u.Detail.MainworldDesignation,
	}
	sf := stars.BuildSurveyForm(u.System, meta)
	form := iiss.Class0IForm{
		FormHeader: iiss.FormHeader{
			SystemName:    u.System.PrimaryDesignation,
			Sector:        sf.Sector,
			Location:      sf.Location,
			IISSDesig:     sf.IISSDesig,
			InitialSurvey: sf.InitialSurvey,
```

## Class IV-P: a fully-owned iiss form

The Class IV-P "Planetary Detail" form is the per-body deep dive for the auto-picked mainworld. Its body structs (`Class4PPartP` for planet/moon, `Class4PPartPB` for belt) and their Markdown renderers live in `iiss/class4p.go`; `Class4PForm` holds one of them as a concrete pointer per `Variant`. `worlds` builds the struct from the `Universe` (`buildClass4PPlanet` / `buildClass4PBelt` in `worlds/iiss_class4p.go`); `iiss` renders it. No closure, no `any` — the same struct serves both the Markdown and JSON paths.

```bash
sed -n '64,87p' iiss/forms.go
```

```output
// Class4PVariant identifies which Class IV-P variant applies to the
// auto-picked mainworld.
type Class4PVariant int

const (
	// Class4PPlanet — mainworld is a planet.
	Class4PPlanet Class4PVariant = iota
	// Class4PMoon — mainworld is a moon.
	Class4PMoon
	// Class4PBelt — mainworld is a belt.
	Class4PBelt
)

// Class4PForm is the IISS Class IV-P "Planetary Detail" Survey form,
// rendered only for the auto-picked mainworld. Exactly one of PartP
// (planet/moon) or PartPB (belt) is populated, per Variant; the other is
// nil. Both are concrete iiss types, so the form is fully owned by iiss/
// and marshals to JSON without a worlds-side payload.
type Class4PForm struct {
	FormHeader
	Variant Class4PVariant
	PartP   *Class4PPartP  `json:",omitempty"`
	PartPB  *Class4PPartPB `json:",omitempty"`
}
```

## Notable Features — the referee summary

Above the IISS forms, a referee-facing summary block flags conditions a Game Master wants at a glance: tidal locks, cold snaps, crush worlds (high G + high atm), taint chains, and the mainworld habitability rationale. Five sections, each rendered only when non-empty. Data is in `Universe.Detail.Bodies`; this is a scanner + renderer.

```bash
sed -n '1,15p;22,77p' worlds/notable_features.go | head -80
```

```output
package worlds

import (
	"fmt"
	"strings"
)

// Notable-feature thresholds. Tuned for "what would a Referee want at a
// glance" — borrowed from habitability.go's DM bands where applicable
// so the flags align with the book's hazard rhetoric.
const (
	notableGravityHigh    = 1.4 // > p.132 "uncomfortably high" boundary
	notablePressureHigh   = 2.5 // bar; structures need reinforcement, breathing problematic
	notableWorstLowK      = 233 // ~ -40°C; severe frostbite, sustained cold lethal to unprotected humans
	notableMeanLivableK   = 250 // ~ -23°C; below this the mean is already too cold for a "snap" to be the story
// mainworld's habitability rationale. Returns "" if nothing notable.
//
// Inserted by BuildIISSForms above the IISS forms in cmd/world-builder's
// Markdown output. The block is informational only — every fact
// surfaced is already present in the IISS forms, just dispersed
// across them.
func NotableFeatures(u *Universe) string {
	if u == nil {
		return ""
	}
	tidals := collectTidalLocks(u)
	colds := collectColdSnaps(u)
	crushes := collectCrushWorlds(u)
	taints := collectTaintChains(u)
	mainworldNote := mainworldHabitabilityNote(u)

	if len(tidals) == 0 && len(colds) == 0 && len(crushes) == 0 && len(taints) == 0 && mainworldNote == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Notable Features\n\n")

	if len(tidals) > 0 {
		b.WriteString("### Tidal locks\n")
		for _, line := range tidals {
			fmt.Fprintf(&b, "- %s\n", line)
		}
		b.WriteString("\n")
	}
	if len(colds) > 0 {
		b.WriteString("### Cold snaps\n")
		for _, line := range colds {
			fmt.Fprintf(&b, "- %s\n", line)
		}
		b.WriteString("\n")
	}
	if len(crushes) > 0 {
		b.WriteString("### Crush worlds\n")
		for _, line := range crushes {
			fmt.Fprintf(&b, "- %s\n", line)
		}
		b.WriteString("\n")
	}
	if len(taints) > 0 {
		b.WriteString("### Taint chains\n")
		for _, line := range taints {
			fmt.Fprintf(&b, "- %s\n", line)
		}
		b.WriteString("\n")
	}
	if mainworldNote != "" {
		b.WriteString("### Mainworld habitability\n")
		fmt.Fprintf(&b, "- %s\n\n", mainworldNote)
	}

```

## The renderer: iiss.MarkdownSystem

The final step. Pure function over `iiss.SystemForms`; concatenates the four sections (notable features, Class 0/I, Class II/III, Class IV-P) under H1/H2 headings in book order. `iiss/` does not import `worlds/` — every form, including the Class IV-P body, is a concrete `iiss` struct that `worlds.BuildIISSForms` populated.

```bash
sed -n '95,120p' iiss/render.go
```

```output
// book order. Class IV-P renders only for the auto-picked mainworld.
func MarkdownSystem(sf SystemForms) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — System Survey\n\n", sf.Class0I.IISSDesig)
	if sf.MainworldDesignation != "" {
		fmt.Fprintf(&b, "**Mainworld:** %s\n\n", sf.MainworldDesignation)
	}
	if sf.ShortProfile != "" {
		fmt.Fprintf(&b, "Short profile: `%s`\n\n", sf.ShortProfile)
	}
	if sf.LongProfile != "" {
		fmt.Fprintf(&b, "Long profile: `%s`\n\n", sf.LongProfile)
	}
	if sf.NotableFeatures != "" {
		b.WriteString(sf.NotableFeatures)
		b.WriteString("\n")
	}
	b.WriteString(MarkdownClass0I(sf.Class0I))
	b.WriteString("\n")
	b.WriteString(MarkdownClass23(sf.Class23))
	b.WriteString("\n")
	if sf.MainworldDesignation != "" {
		b.WriteString(MarkdownClass4P(sf.Class4P))
		b.WriteString("\n")
	}
	return b.String()
```

## Test layers

Coverage is four-layered (per `docs/harness.md`):

1. **Per-procedure tests** — every `Roll*` / `Generate*` / `Compute*` function has at least one test that drives book-narrated dice and asserts the expected output to the digit. Bulk of coverage.
2. **Named worked-example fixtures** — Sol, Corella, Zed, Zed Prime carry book dice scripts through multi-procedure chains.
3. **Property tests** — 8 invariants × 1000 seeds each (`worlds/property_test.go`). Smoke tests for systemic correctness; fire when a procedure silently does nothing for a class of bodies.
4. **Markdown regression baseline** — 5 seeds × full Markdown output at `iiss/testdata/seed_*.md`. Refreshable with `go test ./iiss/... -update.regression -run TestRegression`.

Plus a 10 000-seed bulk-sweep verification (one-off `cmd/world-builder-bulk/` runner) confirms zero errors in default operation. See `docs/history/generator-error-catalog.md` for the journey.

```bash
find . -name "*_test.go" -not -path "./docs/*" | wc -l | xargs printf "Test files: %s\n"; grep -rhc "^func Test" --include="*_test.go" . | awk "{ s+=\$1 } END { printf \"Test functions: %d\\n\", s }"
```

```output
Test files: 60
Test functions: 692
```

## Sample run

End-to-end: generate seed 42 in short-profile form, then show the first 30 lines of the Markdown output to see the shape.

```bash
go run ./cmd/world-builder -seed 42 -format short
```

```output
2-1-9-8-0.7
```

```bash
go run ./cmd/world-builder -seed 42 -format markdown | head -30
```

```output
# A VII — System Survey

**Mainworld:** A VII

Short profile: `2-1-9-8-0.7`

Long profile: `A-7-T-G-T-T-G-T-T-T-0.7:B-0-T-T-P-T-0.7`

## Notable Features

### Tidal locks
- A I: planet → star, 1:1, twilight zone
- A II b: moon → planet, 1:1
- A III a: planet → star, 3:2
- A III b: planet → star, 1:1, twilight zone
- A III c: planet → star, 1:1, twilight zone
- A V a: moon → planet, 1:1
- A V b: moon → planet, 1:1
- A V c: moon → planet, 1:1
- A V d: moon → planet, 1:1
- A V e: moon → planet, 1:1
- A V f: moon → planet, 1:1
- A V g: moon → planet, 1:1
- A V h: moon → planet, 1:1
- A V i: moon → planet, 1:1
- A V j: moon → planet, 1:1
- B I: planet → star, 1:1, twilight zone
- B II: planet → moon, 1:1
- B II a: planet → star, 1:1, twilight zone
- B II b: planet → star, 1:1, twilight zone
```

## Where to read next

- [`docs/design-intent.md`](design-intent.md) — why the architecture looks this way.
- [`docs/api-surface.md`](api-surface.md) — the public API reference.
- [`docs/dependency-graph.md`](dependency-graph.md) — every value, its inputs, the one cyclic (climate) cluster.
- [`docs/anti-patterns.md`](anti-patterns.md) — failure modes the code guards against.
- [`docs/harness.md`](harness.md) — fixture catalog + the four-layer test strategy.
- [`docs/wbh-inconsistencies.md`](wbh-inconsistencies.md) — six book-internal divergences with chosen interpretations.
- [`docs/next-steps.md`](next-steps.md) — open post-v1.0 items.
