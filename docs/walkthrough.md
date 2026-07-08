# World Builder Walkthrough

*2026-07-08T01:37:42Z by Showboat 0.6.1*
<!-- showboat-id: 68d6c26e-ddfa-4687-83d7-a4bbd1277509 -->

## Overview

This project generates complete Mongoose Traveller _World Builder's Handbook_ (WBH) star systems end-to-end. A seed plus deterministic dice produces a real, fully-formed system — primary star, companions, planets, moons, belts, atmospheres, life, habitability — rendered as a single **IISS Class IV Survey** document: a system-level **PART 1 — System Census** followed by a per-body **PART P** (planet/moon/gas-giant) or **PART P.B** (belt) for every notable world.

WBH pp.14–146 (Stars + System Worlds and Orbits + World Physical Characteristics) are implemented to book fidelity. WBH pp.147+ (World Social Characteristics, Special Circumstances) are out of scope.

This walkthrough follows the code top-down: from the CLI entry, through the deterministic roller, into the stars/worlds/iiss pipelines, ending with the rendered Markdown.

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
- `iiss/` — Pure renderer package. Holds the IISS form types and `MarkdownClass4Survey`. Does not import `worlds/`.
- `cmd/world-builder/` — CLI entry; three lines of pipeline plus format dispatch.

## Entry point: cmd/world-builder/run

`run` parses `-seed` / `-format` (plus a `-peculiar` column selector), generates the universe, and dispatches on format: `markdown` (default), `json`, `short`.

```bash
sed -n '26,80p' cmd/world-builder/main.go
```

```output
func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("world-builder", flag.ContinueOnError)
	fs.SetOutput(stderr)
	seed := fs.Int64("seed", 0, "random seed (omit for time-based)")
	format := fs.String("format", "markdown", "output format: markdown | json | short")
	peculiar := fs.String("peculiar", "special", "column for Special (2D=2) primary rolls: special | unusual | peculiar")
	if err := fs.Parse(args); err != nil {
		return err
	}

	column, err := stars.ParsePeculiarPath(*peculiar)
	if err != nil {
		return err
	}

	// Distinguish "flag omitted" from "-seed 0": an explicit 0 is a
	// legitimate reproducible seed, not a request for time-based.
	seedSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			seedSet = true
		}
	})
	s := *seed
	if !seedSet {
		s = time.Now().UnixNano()
	}

	u, err := worlds.GenerateWithOpts(s, worlds.GenerateOpts{PeculiarColumn: column})
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	switch *format {
	case "markdown":
		_, err := fmt.Fprint(stdout, iiss.MarkdownClass4Survey(u.Detail.SystemForms))
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
	"hash/fnv"
	"math/rand"

	"github.com/philoserf/world-builder/dice"
)

// Roller is the interface every dice-driven procedure depends on.
type Roller interface {
	// Roll executes the given dice notation (e.g. "2D", "2D-7", "d10")
	// and returns the result, including any modifier in the notation.
	Roll(notation string) int

	// Fork returns a child Roller whose stream is deterministically
	// derived from this Roller's identity and key. Forking never
	// consumes a draw from the parent, so a parent and its forks are
	// independent: rolls taken from one do not perturb the other.
	//
	// SPIKE C1 (docs/rebuild-spec.md § C1). The pipeline uses Fork to
	// give every (body, procedure-family) its own substream keyed by
	// stable body identity, so reordering a stage or re-rolling one body
	// (the tidal-lock cascade) leaves every other body's values byte-
	// identical. Seeded branches; Scripted and Fixed are transparent
	// (return a view over the same sequence / value) so per-procedure
	// worked-example fixtures that feed a flat dice list are unaffected.
	Fork(key string) Roller
}

// Seeded is a production roller backed by a seeded *math/rand.Rand.
type Seeded struct {
	seed int64 // immutable construction seed; Fork derives children from it
	rng  *rand.Rand
}
```

## worlds.Generate — the top-level pipeline

`worlds.Generate(seed)` wraps `GenerateWithRollerOpts`, which runs every stage in dependency-graph order. The stage slice is the pipeline at a glance: placement, then eleven `Apply*` passes, then aggregation and IISS-form construction. Note stage `ApplyTidalLockReEval` (WBH p.106) sits right after `ApplyClimate` — the tidal-lock cascade re-evaluates once the atmosphere is known.

```bash
sed -n '43,84p' worlds/generate.go
```

```output
func GenerateWithRoller(r roller.Roller) (Universe, error) {
	return GenerateWithRollerOpts(r, GenerateOpts{})
}

// GenerateWithRollerOpts is GenerateWithRoller with options.
func GenerateWithRollerOpts(r roller.Roller, opts GenerateOpts) (Universe, error) {
	sys, err := stars.GenerateSystem(r, stars.GenerateSystemOpts{
		WithVariance:   true,
		Accuracy:       2,
		PeculiarColumn: opts.PeculiarColumn,
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
		ApplyTidalLockReEval,
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

The first stage rolls the primary star and all companions in a fixed roll-consumption order (documented in the function's own doc-comment). A `Special` (2D=2) primary dispatches through the WBH p.15 column selected by `PeculiarColumn` — the default `Special` column yields Class VI / IV / III / Giants only, which is what makes every default seed produce a real, habitable-capable system.

```bash
sed -n '20,58p' stars/system.go
```

```output
// computes it directly.
type GenerateSystemOpts struct {
	WithVariance   bool
	Accuracy       int          // 1 or 2
	PeculiarColumn PeculiarPath // zero value = PeculiarPathSpecial
}

// GenerateSystem rolls a complete multi-star system from a Roller.
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
	// companions of present orbit-class stars in Close/Near/Far order.
	closePresent := RollPresence(r, primary, OrbitClose)
	nearPresent := RollPresence(r, primary, OrbitNear)
	farPresent := RollPresence(r, primary, OrbitFar)
	primaryHasCompanion := RollPresence(r, primary, OrbitCompanion)

	var closeCompanionPresent, nearCompanionPresent, farCompanionPresent bool
```

## Stage 1: worlds.GenerateSystemPlacement

Allocates orbit slots to stars, fixes the baseline orbit, and places bodies into slots without yet sizing them. WBH pp.36–52: counts → available orbits → allocations → baseline number → baseline orbit → empty orbits → spread → slots.

```bash
sed -n '35,68p' worlds/system_placement.go
```

```output
func GenerateSystemPlacement(r roller.Roller, sys stars.System) (SystemPlacement, error) {
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
```

## The unified Body type

Every placed object — planet, moon, belt member — is a single `Body` value. Moons are `Body{Kind: BodyMoon, Parent: <planet>}`. A unified iterator (`Universe.AllBodies()`) yields every body, so the "moon-path silent-zero" anti-pattern (a procedure that runs for planets but forgets moons) is prevented at the type level. The pointer fields are populated stage-by-stage; a nil pointer means "not applicable to this body", read through `Has*()` predicates.

```bash
sed -n '31,99p' worlds/body.go
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

	// Ring is set on a parent when WBH calls for a planetary ring in
	// place of significant moons: a moon-quantity roll of exactly 0
	// (p.55) or Hill-sphere moon removal (p.76). RingCentrePD / RingSpanPD
	// carry the ring's centre location and span in planet-diameters
	// (WBH p.77), rolled in ApplyMoonRefinement when Ring is set.
	Ring         bool
	RingCentrePD float64
	RingSpanPD   float64

	// Stage 3
	Physical *BodyPhysical
	Belt     *BeltDetails

	// Stage 4
	DayLength    *DayLength
	AxialTilt    *AxialTilt
	TidalLock    *TidalLock
	TidalEffects *SurfaceTidalEffects

	// preTidalLockSnapshot is set during Stage 4 just before
	// ApplyTidalLockEffect; consumed by ApplyTidalLockReEval after
	// Stage 5 to restore pre-tidal-lock state when the atmosphere DM
	// (WBH p.106) re-evaluates the lock. Package-private; not part of
	// the rendered output (IISS forms ignore it).
	preTidalLockSnapshot *PreTidalLockSnapshot

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
```

## The climate solver: ApplyClimatePasses

Stage 5 is the only cyclic cluster: atmosphere ↔ temperature ↔ hydrographics depend on each other. `ApplyClimatePasses` runs **two passes** and trusts the second. It is not a fixed-point solver — `RederiveAtmosphereHydrographics` consumes fresh dice each pass, so each pass is a stochastic sample, not a convergence step. There is no fixed point to reach; the function is named for what it does.

```bash
sed -n '119,170p' worlds/climate.go
```

```output
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

## Mainworld pick and IISS-form construction

After all per-body stages, `AggregateSystem` picks the mainworld (priority: HZ-terrestrial → HZ-moon → habitable rocky → any rocky/moon/belt) and computes the profile strings. `BuildIISSForms` then translates `Universe` state into the `SystemForms` boundary struct: the retained `Class0I` / `Class23` carriers (PART 1 census data), a `Class4PForms` slice with one PART P / PART P.B per non-empty body, the `Census` scalars, and the Notable Features block. It is a pure function — no rolls.

```bash
sed -n '16,28p' worlds/iiss_build.go
```

```output
func BuildIISSForms(u *Universe) {
	c0 := buildClass0I(u)
	u.Detail.Class0I = c0
	u.Detail.Class23 = buildClass23(u, c0)
	u.Detail.Class4PForms = buildClass4PForms(u, c0.FormHeader)
	u.Detail.Census = iiss.SystemCensus{
		BaselineNumber: u.Placement.BaselineN,
		BaselineOrbit:  u.Placement.BaselineOrbit,
		Spread:         u.Placement.SystemSpread,
		EmptyOrbits:    u.Placement.EmptyOrbits,
	}
	u.Detail.NotableFeatures = NotableFeatures(u)
}
```

## The Class IV Survey schema

`SystemForms` is the `iiss`↔`worlds` boundary type. `Class0I` and `Class23` are retained purely as PART 1 data carriers (the stars roster and body roster the old "short forms" used to render standalone). `Class4PForms` is one `Class4PForm` per non-empty body, in `AllBodies()` order. Each `Class4PForm` holds a concrete `*Class4PPartP` (planet/moon/gas-giant) or `*Class4PPartPB` (belt) — no `any`, no closures, so the same struct serves both the Markdown and JSON paths.

```bash
sed -n '84,111p' iiss/forms.go
```

```output
type Class4PForm struct {
	FormHeader
	// Designation is the surveyed body's designation (e.g. "Aab IV d"),
	// used to title the per-body PART P / PART P.B heading. Distinct from
	// the system-level FormHeader.IISSDesig.
	Designation string
	Variant     Class4PVariant
	PartP       *Class4PPartP  `json:",omitempty"`
	PartPB      *Class4PPartPB `json:",omitempty"`
}

// SystemForms aggregates the three IISS forms for a generated system,
// plus the system-wide profile strings, the auto-picked mainworld
// designation, and a Markdown referee-facing Notable Features summary.
// Renderer functions take SystemForms (or one of its fields) so iiss/
// does not import worlds/.
type SystemForms struct {
	// Class0I and Class23 are retained as the data carriers for the
	// Class IV Survey's PART 1 (system census) — the stars roster lives on
	// Class0I.Stars, the body roster on Class23.Objects. They are no longer
	// rendered as standalone forms.
	Class0I Class0IForm
	Class23 Class23Form
	// Class4PForms holds one PART P / PART P.B per non-empty body, in
	// AllBodies() order (ascending orbit, moons after their parent).
	Class4PForms         []Class4PForm
	Census               SystemCensus
	MainworldDesignation string
```

## Per-body builders: buildClass4PForms

`buildClass4PForms` walks `AllBodies()`, skips empty slots, and builds one PART per body — `buildClass4PBelt` for belts, `buildClass4PPlanet` for terrestrials, moons, and gas giants. The auto-picked mainworld's part is flagged (`isMW`). Gas giants take the planet builder but render a GG-appropriate PART P (class + residual heat, no terrestrial atmosphere/hydro sections).

```bash
sed -n '259,300p' worlds/iiss_build.go
```

```output
func buildClass4PForms(u *Universe, header iiss.FormHeader) []iiss.Class4PForm {
	var forms []iiss.Class4PForm
	for body := range u.AllBodies() {
		if body.Kind == BodyEmpty {
			continue
		}
		isMW := body == u.Detail.Mainworld
		form := iiss.Class4PForm{FormHeader: header, Designation: body.Designation}
		switch body.Kind {
		case BodyPlanetoidBelt:
			form.Variant = iiss.Class4PBelt
			form.PartPB = buildClass4PBelt(u, body, isMW)
		case BodyMoon:
			form.Variant = iiss.Class4PMoon
			form.PartP = buildClass4PPlanet(u, body, isMW)
		default: // BodyTerrestrial, BodyGasGiant
			form.Variant = iiss.Class4PPlanet
			form.PartP = buildClass4PPlanet(u, body, isMW)
		}
		forms = append(forms, form)
	}
	return forms
}
```

## Notable Features — the referee summary

Above PART 1, a referee-facing summary block flags conditions a Game Master wants at a glance: tidal locks, cold snaps, crush worlds (high G + high atmosphere), taint chains, and the mainworld habitability rationale. Five sections, each rendered only when non-empty. It is a scanner over `Universe.Detail.Bodies` — every fact it surfaces already exists in the survey, just dispersed across the per-body parts.

```bash
sed -n '28,60p' worlds/notable_features.go
```

```output
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
```

## The renderer: iiss.MarkdownClass4Survey

The final step. A pure function over `iiss.SystemForms`: an H1 title, optional profile lines, the Notable Features block, `markdownClass4Part1` (the system census), then one per-body part per `Class4PForms` entry. `iiss/` imports nothing from `worlds/` — every part is a concrete `iiss` struct that `BuildIISSForms` populated.

```bash
sed -n '13,35p' iiss/render.go
```

```output
func MarkdownClass4Survey(sf SystemForms) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — IISS Class IV Survey\n\n", sf.Class0I.IISSDesig)
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
	b.WriteString(markdownClass4Part1(sf))
	for _, f := range sf.Class4PForms {
		b.WriteString("\n")
		b.WriteString(markdownClass4Part(f))
	}
	return b.String()
}
```

## PART 1 — System Census

`markdownClass4Part1` folds the old short-form content into one census: system scalars (age, counts, baseline number/orbit, spread, empty orbits), a stellar roster with full orbital data, and the body roster. The stars table carries the companion Orbit#/AU/ecc/period/MAO/HZCO columns plus an **HZ Orbit#** breadth column (HZCO ± 1.0, WBH p.43) computed by `hzRange`.

```bash
sed -n '41,68p' iiss/render.go
```

```output
func markdownClass4Part1(sf SystemForms) string {
	var b strings.Builder
	c0 := sf.Class0I
	b.WriteString("## PART 1 — System Census\n\n")

	fmt.Fprintf(&b, "- System: %s\n", c0.SystemName)
	fmt.Fprintf(&b, "- Sector / Location: %s / %s\n", c0.Sector, c0.Location)
	fmt.Fprintf(&b, "- Survey: initial %s, last updated %s\n", c0.InitialSurvey, c0.LastUpdated)
	fmt.Fprintf(&b, "- System age: %.3f Gyr\n", c0.SystemAgeGyr)
	fmt.Fprintf(&b, "- Stellar count: %d\n", c0.StellarCount)
	fmt.Fprintf(&b, "- Worlds: %d gas giants, %d belts, %d terrestrials (total %d)\n",
		sf.Class23.Counts.GasGiants, sf.Class23.Counts.PlanetoidBelts,
		sf.Class23.Counts.Terrestrials, sf.Class23.Counts.Total)
	fmt.Fprintf(&b, "- Baseline: number %d, Orbit# %.2f; spread %.2f; empty orbits %d\n\n",
		sf.Census.BaselineNumber, sf.Census.BaselineOrbit, sf.Census.Spread, sf.Census.EmptyOrbits)

	if len(c0.Stars) > 0 {
		b.WriteString("### Stars\n\n")
		b.WriteString("| Component | Class | Mass | Diameter | Temp (K) | Luminosity | Orbit | AU | Ecc | Period (y) | MAO | HZCO | HZ Orbit# |\n")
		b.WriteString("| --------- | ----- | ---- | -------- | -------- | ---------- | ----- | --- | --- | ---------- | --- | ---- | --------- |\n")
		for _, s := range c0.Stars {
			fmt.Fprintf(&b, "| %s | %s | %.3f | %.3f | %.0f | %.4f | %s | %s | %s | %s | %.2f | %.2f | %s |\n",
				s.Component, s.Class, s.Mass, s.Diameter, s.Temperature, s.Luminosity,
				blankFloat(s.Orbit), blankFloat(s.AU), blankFloat(s.Eccentricity),
				blankFloat(s.PeriodYears), s.MAO, s.HZCO, hzRange(s.HZCO))
		}
		b.WriteString("\nHabitable zone breadth: ±1.0 Orbit# from HZCO (WBH p.43).\n\n")
	}
```

## Test layers

Coverage is four-layered (per `docs/harness.md`): per-procedure worked-example tests (book-narrated dice, asserted to the digit — the proof of fidelity), named worked-example fixtures (Sol, Corella, Zed, Zed Prime), property tests (invariants × 1000 seeds), and a Markdown regression baseline (`iiss/testdata/seed_*.md`) plus the Zed gold-master (`worlds/testdata/zed_gold.md`).

```bash
find . -name "*_test.go" -not -path "./docs/*" | wc -l | xargs printf "Test files: %s\n"; grep -rhc "^func Test" --include="*_test.go" . | awk "{ s+=\$1 } END { printf \"Test functions: %d\\n\", s }"
```

```output
Test files: 66
Test functions: 763
```

## Sample run

End-to-end: seed 42 as a short profile, then the first 34 lines of the Markdown survey — the H1 title, mainworld, profile strings, and the start of the Notable Features referee summary (the per-body PART P / PART P.B sections follow further down).

```bash
go run ./cmd/world-builder -seed 42 -format short
```

```output
2-1-9-8-0.7
```

```bash
go run ./cmd/world-builder -seed 42 -format markdown | head -34
```

```output
# A VII — IISS Class IV Survey

**Mainworld:** A VII

Short profile: `2-1-9-8-0.7`

Long profile: `A-7-T-G-T-T-G-T-T-T-0.7:B-0-T-T-P-T-0.7`

## Notable Features

### Tidal locks
- A I: planet → star, 1:1, twilight zone
- A II a: moon → planet, 1:1
- A II b: moon → planet, 1:1
- A III: planet → moon, 1:1
- A III a: moon → planet, 1:1
- A III b: moon → planet, 1:1
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
- B II a: moon → planet, 1:1
- B II b: moon → planet, 1:1
- B II c: moon → planet, 1:1
- B II d: moon → planet, 1:1
- B II e: moon → planet, 1:1

```

## Where to read next

- [`docs/design-intent.md`](design-intent.md) — why the architecture looks this way.
- [`docs/api-surface.md`](api-surface.md) — the public API reference.
- [`docs/dependency-graph.md`](dependency-graph.md) — every value, its inputs, the one cyclic (climate) cluster.
- [`docs/anti-patterns.md`](anti-patterns.md) — failure modes the code guards against.
- [`docs/harness.md`](harness.md) — fixture catalog + the four-layer test strategy.
- [`docs/wbh-inconsistencies.md`](wbh-inconsistencies.md) — book-internal divergences with chosen interpretations.

