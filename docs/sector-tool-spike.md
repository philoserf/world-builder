# Sector-Scale Companion Tool — Design Spike

**Status:** exploratory. Not a commitment to build. This document captures the design surface for a sector-scale companion tool that would generate collections of physical star systems on a Traveller hex grid, leveraging the existing world-builder library for per-system rules.

**Source material:** Mongoose Publishing's _Sector Construction Guide_ (Lanesskog 2024), specifically Phase 1 (System Location) on pp.9–16 and Sector Details on pp.17–27. Same author as WBH; explicitly a companion volume.

## Goal

Given a seed and sector dimensions, deterministically produce a sector layout: which hexes are occupied, where, with what per-hex seed. Each occupied hex is paired with a `worlds.Universe` accessible on demand. Variable stellar density across the sector (uniform default; contour-style variation as an option). Project-supplied anomaly annotations layered on top (barren hexes, no-gas-giant hexes, lost-ship zones — Referee fiat, not rolled). Optional connectivity analysis: given a jump rating and a refueling predicate, return the reachable-system graph.

The library remains the artifact; the new tool consumes it.

## Non-goals

- **No social characteristics.** Population, Government, Law Level, Starport, Tech Level all live in WBH pp.147+ and remain out of scope. See [feedback note on Trade Codes](#) — climate codes Fr/Co/Ho/Bo also belong in this excluded layer.
- **No polity/sophont generation.** SCG's polity and sophont chapters are separate machinery.
- **No three-dimensional mapping** (2300AD-style). SCG explicitly punts on this; we punt too.
- **No alternative system-creation procedures.** SCG's appendix uses Core Rulebook 2D-2 sizing with bolted variants; world-builder follows WBH p.54, which already covers Size 1–F. The sector tool consumes world-builder's existing output — it does not introduce a parallel rule set.

## Three shapes

| Shape                                                                              | Pros                                                      | Cons                                                                          |
| ---------------------------------------------------------------------------------- | --------------------------------------------------------- | ----------------------------------------------------------------------------- |
| A. `cmd/sector` in this repo                                                       | smallest delta; ships fast                                | grows `cmd/`; world-builder's "done" scope text becomes inaccurate            |
| B. New `wbh/sector` package + `cmd/wbh -sector` flag                               | library-first per project principle                       | expands world-builder beyond its stated scope (Stars + Worlds pp.14–146)      |
| C. Separate repo `philoserf/sector-builder` depending on world-builder as a module | preserves world-builder's "done" status; clean separation | two repos to maintain; cross-repo coordination if world-builder's API changes |

**The core decision is scope-political, not technical.** World-builder's CLAUDE.md says the project is done when WBH pp.14–146 are encoded and `cmd/wbh -format markdown` emits the three IISS forms — which is the current state on `main`. Adding sector machinery here (A or B) expands what "done" means. Adding it as a separate repo (C) doesn't.

**Recommendation: C.** World-builder stays focused; the sector tool consumes it via `go get wbh@vX.Y.Z`. Trade-off accepted: one more repo. The rest of this document assumes C, but everything except module path and packaging carries over to B trivially.

## API sketch

```go
// philoserf/sector-builder/sector

package sector

import "wbh/worlds" // depends on world-builder

// Generate produces a sector layout deterministically from seed + opts.
// Per-hex systems are not generated eagerly — call Hex.GenerateSystem to
// materialize one. Generating all 1,280 hexes would take minutes.
func Generate(seed int64, opts Opts) (Sector, error)

type Opts struct {
	Width, Height int         // default 32, 40 (one sector)
	Density       DensityFunc // default UniformDensity(4) — 1D ≥ 4
	Anomalies     []Anomaly   // project-supplied; layered on top of rolls
}

type DensityFunc func(x, y int) int

// Returns the threshold for system presence at (x,y). The default
// UniformDensity(4) returns 4 for every hex (the SCG-default ~50%).
// Higher value → sparser. SCG's "2D≥11" rift convention is encoded as
// a sentinel value 11+ that the roller interprets as 2D, not 1D.

func UniformDensity(threshold int) DensityFunc
func ContourDensity(regions []DensityRegion) DensityFunc

type DensityRegion struct {
	// A rectangular sub-region with its own threshold; later entries override
	// earlier ones. SCG's "contour lines" are approximated as overlapping rects.
	XMin, YMin, XMax, YMax int
	Threshold              int
}

type Anomaly struct {
	X, Y int
	Kind AnomalyKind
}

type AnomalyKind int

const (
	AnomalyBarren     AnomalyKind = iota // hex is occupied but flagged Pop 0
	AnomalyNoGasGiant                    // override world-builder's GG output
	AnomalyLostShip                      // metadata only; no rule effect
	// (extend as needed)
)

type Sector struct {
	Seed   int64
	Width  int
	Height int
	Hexes  []Hex
}

type Hex struct {
	X, Y    int   // 1-indexed per Traveller convention
	Present bool  // density roll succeeded
	Seed    int64 // per-hex seed for lazy world generation
	Anomaly AnomalyKind
}

// GenerateSystem materializes the per-hex Universe via the world-builder
// library. Stateless — call multiple times, get the same result (seed-driven).
func (h Hex) GenerateSystem() (worlds.Universe, error) {
	if !h.Present {
		return worlds.Universe{}, ErrEmptyHex
	}
	return worlds.Generate(h.Seed)
}
```

## Seed strategy (load-bearing)

A sector's `seed int64` must deterministically produce per-hex seeds AND those per-hex seeds must compose with world-builder's existing seed contract (which is "seed plus options fully determines a system" per the project's core invariant).

Cleanest approach: `hex.Seed = int64(splitMix64(sector.Seed ^ uint64(y)<<32 ^ uint64(x)))` or any analogous deterministic 2D-coord hash. Pick a specific hash function and pin it forever — changing it would invalidate every previously-generated sector.

Verify before building: round-trip `worlds.Generate(hex.Seed)` for any single hex must produce the same Universe regardless of whether the sector-level call was made.

## Density mechanics

SCG default: 1D ≥ 4 means "system present" — 50% occupancy. SCG's variant rules use higher thresholds (1D=6, 2D=11+, 2D=12) for sparser regions.

`DensityFunc` returning an integer `t` is interpreted as:

- `t ≤ 6`: roll 1D, occupied if `roll ≥ t`
- `t = 7..12`: roll 2D, occupied if `roll ≥ t`

This covers SCG's full range. The roll is deterministic from the per-hex seed.

## Anomaly handling

Anomalies are _post-roll_ overrides — Referee fiat, not procedural. The generator applies them after the density roll:

- `AnomalyBarren`: hex stays `Present=true` so it shows up on maps, but `GenerateSystem` returns a barren-flagged Universe (or a sentinel) — design decision pending.
- `AnomalyNoGasGiant`: hex generates normally, then the caller is expected to handle the override. Or the sector library wraps `worlds.Generate` and zeroes out gas giants. Simpler: leave to caller.
- `AnomalyLostShip`: pure metadata, no rule effect. Renders on the map.

Open question: should anomalies be a separate post-processing step rather than baked into `Hex`? Cleaner separation but slightly more API surface for callers.

## Connectivity (separate sub-tool)

`sector.Connectivity` is a pure function over a generated sector:

```go
type Connectivity struct {
	JumpRating int                        // 1..6
	Refuels    func(worlds.Universe) bool // gas giant? wilderness ocean?
	Graph      map[HexCoord][]HexCoord    // adjacency
}

func ComputeConnectivity(s Sector, jr int, refuels func(worlds.Universe) bool) Connectivity
```

This forces per-hex `GenerateSystem` calls for the refuel predicate, so it's the expensive operation. Cache or memoize at the caller's discretion.

Out of scope for the initial spike but mentioned because it's the most obvious thing to do once positions exist.

## Output formats

Two natural emissions:

1. **JSON.** Full sector + per-hex metadata. Per-hex `Universe` is _not_ embedded by default (huge); embed only on `--include-systems`.
2. **Markdown sector summary.** Table per subsector with hex coords, system designation, and the mainworld's short profile (one line per system — leverage the existing `cmd/wbh -format short`).

Per-system full Markdown (`cmd/wbh -format markdown`) is the existing tool's job — the sector tool would print a hex coord and let the user run `wbh -seed <hex.Seed>` for the full profile, OR expose a `--hex X,Y` shorthand that does it.

## Testing strategy

- **Determinism**: same seed + opts → byte-identical `Sector`. Property test.
- **Density distribution**: over N seeds with uniform threshold T, the observed occupancy rate matches `P(1D ≥ T) ± 3σ`. Sample size makes this robust.
- **Anomaly application**: anomaly entries override the density result correctly; barren-flag and no-gas-giant overrides reproduce.
- **Per-hex seed isolation**: hex (X,Y) under sector seed S produces the same `Universe` as `worlds.Generate(hex.Seed)` called directly. (Verifies the seed-composition contract.)
- **SCG Foreven sub-fixture (optional)**: SCG's example sector has pre-defined hexes. If we encode a few of its anomaly clusters as fixture data, we can verify the tool reproduces SCG's stated map. Not a "fidelity test" in the WBH-Zed sense — SCG doesn't have hex-by-hex worked examples — but a sanity check.

## Cost estimate (rough)

- API + density + anomaly + seed strategy: ~2 days.
- JSON + Markdown emission: ~1 day.
- Connectivity sub-tool: ~1 day.
- Testing (incl. SCG fixture if pursued): ~1–2 days.
- Cross-repo dependency wiring + CI: ~0.5 day.
- **Total: ~5–7 days of focused work.**

That's the "build it" estimate, not the "understand it" cost.

## Risks

- **Seed-composition contract.** If world-builder ever changes its seed interpretation, every previously-generated sector becomes invalid. Mitigation: pin a `worlds@v1.x.y` dependency and don't bump it without a planned sector-data migration.
- **Anomaly schema bloat.** If we let anomalies accumulate (every Referee idea becomes a new `AnomalyKind`), the API surface grows. Mitigation: keep `AnomalyKind` minimal; rich anomaly metadata lives in a separate `Notes` field per hex.
- **Connectivity performance.** A 1280-hex sector with full refuel-predicate evaluation requires 1280 calls to `worlds.Generate`. Mitigation: cache per-hex `Universe` results in memory; offer a "presence-only" connectivity mode that assumes all systems can refuel.
- **Scope creep.** The temptation to add social characteristics or polity boundaries to make the sector "feel complete." Mitigation: hard-pin the goal as "physical sector layout"; social layer is a separate project explicitly out of scope.

## Open questions to resolve before coding

1. **Shape A, B, or C?** (Recommendation: C.)
2. **Hash function for per-hex seed?** SplitMix64 is fine; commit to it forever.
3. **Density representation?** Function or data structure? (Sketch uses function; data structure would serialize cleanly.)
4. **Anomaly: baked into `Hex` or separate post-processing layer?** (Sketch bakes in; separation is more orthogonal.)
5. **Barren-hex behavior?** Does `GenerateSystem` on a barren hex return an error, a sentinel Universe, or a tagged-but-real Universe?
6. **Sector seed range vs. world-builder seed range?** Both are `int64` today; per-hex seeds are also `int64`. No conflict but worth confirming.

## Related material

- SCG pp.9–16 (System Creation, Phase 1): hex grid mechanics, density, contour lines.
- SCG pp.17–27 (Sector Details): jump connectivity, settlement waves, anomaly clusters.
- SCG p.30: tidal-lock cross-source confirmation — independently corroborates this project's PR #52 (exclude moons from planet→star) and PR #58 (Planet→Moon precondition).
- WBH pp.14–146: per-system rules (this project, complete).
- WBH pp.147–234: world social characteristics — explicitly out of scope here and for any sector-tool that follows.
