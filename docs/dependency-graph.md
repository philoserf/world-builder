# Pass 2 — Data Dependency Graph

This document maps every value in WBH pp.14–146 to its inputs. The graph determines pass-2's structural ordering: bodies are walked in dependency order, not in book pagination order. Where the graph cycles, pass 2 builds an explicit fixed-point solver. Where it is acyclic, pass 2 builds a topological pipeline.

The pass-1 implementation is the authoritative reference for "what actually depends on what" because it shipped working. WBH page citations come from pass-1 doc-comments. Where pass 1 worked around an edge with multi-pass iteration (atmosphere ↔ temperature ↔ hydrographics), this document treats the edge as load-bearing and pass 2 designs for it explicitly.

## At a glance

```
Stage 0: Stars                                                 (pp.14–35)
Stage 1: Counts + Allocations + Baseline + Spread + Slots      (pp.36–52)
Stage 2: Sizing + Moons + Designations + Periods + HZ          (pp.53–67)
Stage 3: Body Physical + Belt Details                          (pp.69–78, 91–93)
Stage 4: 3A2a — Surface dist + Day length + Axial tilt
         + Tidal lock + Surface tidal effects                  (pp.100–108)
Stage 5: CLIMATE FIXED-POINT                                   (pp.79, 81, 96–99,
         { Atmosphere, Hydrographics, Temperature }             102, 108–126)
Stage 6: Atmosphere taint typology                             (pp.81–90)
Stage 7: Geology — TSS + post-TSS temperature update           (pp.125–127)
Stage 8: Biology — Biomass → Biocomplexity → Sophonts
         → Biodiversity → Compatibility → Resource             (pp.127–131)
Stage 9: Habitability                                          (p.132)
Stage 10: System aggregations — BaselineN backfill + profiles
         + IISS forms + mainworld pick                         (pp.58, 132–146)
```

Stages 0–4 and 6–10 are forward-only. Stage 5 is the one fixed-point cluster. Stage 7 contains a one-shot back-edge into stage 5 (TSS-temperature-addition); pass 2 evaluates whether to fold it into the climate solver or accept the one-shot update with explicit drift bounds — see the TSS section below.

## Stage 0: Stars (WBH pp.14–35)

Generates the star system independent of any worlds.

**Inputs:** seed (Roller).

**Procedures:**

- `RollPrimaryTypeAndClass` → SpectralLetter, LuminosityClass.
- `RollSubtype` → numeric subtype.
- `ComputeMass`, `ComputeDiameter`, `ComputeTemperature` (table lookups).
- `ApplyVariance` (cut in pass 2 — see `design-intent.md` cuts list).
- `ComputeLuminosityFromFormula(diameter, temperature)`.
- `SmallStarAge` / `AgeSpecialObject` → AgeGyr.
- For each non-primary slot present (Close/Near/Far/Companion):
  - `RollPresence`
  - `RollNonPrimaryDescriptor` (with one extra "Other" reroll if needed)
  - Companion star generation (same physical pipeline as primary)
  - `RollStellarOrbit`, `RollEccentricity`, `RollInclination`
  - `OrbitPeriodYears` (Kepler's third law) — depends on cumulative inner barycentre mass.
- `AssignDesignations`.

**Output:** `System{Primary, Companions, AgeGyr, PrimaryDesignation}`.

**Notes:**

- `OrbitPeriodYears` for the i-th companion sums the masses of the primary plus all earlier-placed companions (book order is inner-to-outer). Pass 2 keeps this.
- Special objects (Brown Dwarf, White Dwarf, Neutron Star, Black Hole, Pulsar, Nebula, Protostar, Star Cluster, Anomaly) have minimum-useful values: type label, mass, age. Detailed physics — accretion, degenerate-matter equations, jet behavior — is post-parity work, but the type/mass/age trio is enough that a referee can use the body in a campaign. A pass-2 IISS form rendering "Black Hole companion: <stubbed>" is a fidelity-gate failure; "Black Hole companion: 8 M☉, 6.0 Gyr" is acceptable.
- `AgeGyr` flows downstream into atmosphere oxygen fraction, body physical age DMs, biology, geology — it is one of the most reused inputs. Compute once on the primary; companions inherit.

## Stage 1: System Worlds — Placement (WBH pp.36–52)

Allocates orbit slots to stars, fixes the baseline orbit, and places bodies in slots without yet sizing them.

**Inputs:** `System` (Stage 0).

**Sub-stages, ordered:**

1. **`GenerateCounts`** (pp.36–38) → `Counts{Total, GasGiants, Belts, ...}`.
2. **`AvailableOrbits`** (pp.38–43) → `[]Group` with HZ-relative orbit intervals per star.
3. **`AllocateOrbitsByStar`** (pp.43–44) → `[]StarAllocation`.
4. **`RollBaselineNumber`** (pp.44–45) → `BaselineN`.
5. **`BaselineOrbit`** (pp.45–46) → `BaselineOrbit` (depends on the primary group's HZCO and `BaselineN`).
6. **`RollEmptyOrbits`** (p.48) → `EmptyOrbits` count.
7. **`Spread`** (pp.48–49) → `SystemSpread` (depends on primary group, allocations, baseline, total stars).
8. **`PlaceOrbitSlots`** (pp.49–50) → `[]Slot` (ordered by ascending orbit within each group).
9. **`AddAnomalous`** (pp.50–51) → `[]AnomalousSlot`, possibly-revised `Counts`.
10. **`PlaceWorlds`** (pp.51–52) → `[]Placement` (each slot gets a `BodyKind`: Terrestrial, GasGiant, Belt, Empty).
11. **`RollPlanetEccentricities`** (p.52) → eccentricities attached to each `Placement`.

**Output:** `SystemPlacement{Counts, Allocations, BaselineN, BaselineOrbit, EmptyOrbits, SystemSpread, Placements}`.

**Notes:**

- HZCO (habitable-zone central orbit) is computed from the host group's primary star (`Group.HZCO()`); for moons it is inherited from the parent's group. Used in nearly every downstream stage.
- `Placements` are walked in ascending-orbit order within each group; `LongProfile` and `AssignPlanetDesignations` rely on this. Pass 2 preserves this iteration contract.

## Stage 2: Sizing + Moons + Designations + Periods + HZ (WBH pp.53–67)

Per-body. Depends only on `SystemPlacement` and `System`.

**Procedures, ordered:**

1. **Sizing.** `RollTerrestrialSize` for terrestrials → `SizeCode, DiameterKm`. `RollGasGiantSize(dms)` for GGs → `Class, DiameterCode, DiameterEarth, MassEarth`. The sizing DM depends on primary class (Brown Dwarf / M V / VI subtract 1) and `SystemSpread < 0.1`.
2. **Moons.** `CountMoons(parent, dm)` then `SizeMoon(parent)` per moon. Per-die DM is −1 when planet's orbit < 1.0 or adjacent to a companion interval edge.
3. **Designations.** `AssignPlanetDesignations` then `AssignMoonDesignations`. Mechanical, no rolls.
4. **Periods.** `PeriodFor(au, sumStellarMassInterior, bodyMassEarth)` per body. `bodyMassEarth` is non-zero only for gas giants (terrestrial mass is derived later in Stage 3). Belt periods use the same formula at the belt's orbit.
5. **HZ tagging.** `MarkHZ(detailed)` flags bodies in HZCO ± 1.0.

**Notes:**

- Terrestrial `MassEarth` is left zero here and derived later from `Density × DiameterKm`. This was a silent-zero-bug source in pass 1 (memory entry: `feedback_moon_path_silent_zero_pattern`). Pass 2 either (a) populates `MassEarth` here for terrestrials with a placeholder derivation or (b) makes downstream consumers explicit about the lazy derivation. Decision: see `api-surface.md`.
- `Period` uses interior stellar mass for non-gas-giant bodies; gas giants ≥ 100 Earth masses include their own mass. This non-uniformity is a real WBH rule, not a pass-1 quirk.

## Stage 3: Body Physical + Belt Details (WBH pp.69–78, 91–93)

Per-body. Runs for every terrestrial body (planets and moons of any parent type).

### BodyPhysical (terrestrials)

**Inputs:** `SizeCode`, `DiameterKm`, `BodyPhysicalDMs{SizeCode, AtHZCOOrCloser=HZ, BeyondHZCO, SystemAgeGyr}`.

**Procedures:**

- `RollComposition(dms)` → composition string.
- `RollDensity(composition)` → `Density`.
- `GenerateBodyPhysical` returns `BodyPhysical{Composition, Density, Gravity, ...}`.
- `DeriveMass(Density, DiameterKm)` → `MassEarth` (terrestrial). Backfilled into `Body.MassEarth`.

**Notes:** Moons inherit the parent's HZ-offset DMs (they share its orbit). Belts (Size 0) skip this entirely.

### BeltDetails (Size 0 only)

**Inputs:** belt orbit, `SystemSpread`, HZCO, AgeGyr, AdjacentToGG, OutermostSlot.

**Procedures:** `GenerateBeltDetails` → `BeltDetails{Span, Composition, Bulk, Resource, SigSize1, SigSizeS}`.

### Moon refinement

Runs only if the planet has moons and resolvable mass. Computes Hill-sphere moon orbit limit, may remove moons exceeding the limit, then per-moon orbit (`RollMoonOrbit`) and period (`MoonPeriodHours`).

## Stage 4: 3A2a — Rotation/Tilt/Tide pass (WBH pp.100–108)

Per-body, every planet and moon (plus HZ-planet moons get surface distribution and the others). Five sub-passes in order:

1. **`GenerateSurfaceDistribution(hydro)`** → `SurfaceDistribution`. Requires `Hydrographics` to be present. **NB:** at this point in pass 1, `Hydrographics` is the _preliminary_ Stage-5A output, not the post-fixed-point value. Pass 2 restructures so Stage 4 runs against the converged climate (see "Pass-2 sequencing" below).
2. **`GenerateDayLength(dp, sys)`** → `DayLength{SiderealHours, SolarHours, YearDays, IsLong, ...}`. For moons, uses parent's period for calendar quantities (the moon's year around the star is its parent's year).
3. **`GenerateAxialTilt(dp)`** → `AxialTilt{Degrees, IsExtreme}`.
4. **`GenerateTidalLock(dp, m, sys, parent, periodHours)`** → `TidalLock`. May reroll `Eccentricity` if 1:1 lock fires.
5. **`GenerateSurfaceTidalEffects(dp, m, sys, parent)`** → `SurfaceTidalEffects` per zone.

**Notes:**

- Tidal-lock 1:1 mutates `Eccentricity` — this is a one-shot forward update, not a loop, but it must precede any computation that reads eccentricity (notably tidal heating in Stage 7).
- Pass 1 placed Stage 4 between the preliminary atm/hydro (5A) and the temperature pass (5C). Pass 2 must decide: does Stage 4 belong before, after, or inside the climate fixed-point? Day length and axial tilt feed into temperature variance; tidal lock affects greenhouse caps. The honest answer is: **inside or alongside the climate fixed-point**, since tidal lock state can affect temperature which can affect atm. Conservative pass-2 plan: run Stage 4 once before the fixed-point, then re-run only the parts that read climate state if the fixed-point converges to a different equilibrium. Most worlds will be insensitive, but the rule must be explicit, not accidental.

## Stage 5: Climate Fixed-Point — Atmosphere ↔ Hydrographics ↔ Temperature

This is the load-bearing structural insight pass 2 corrects. Pass 1 generated atm/hydro from preliminary inputs, generated temperature from those, then re-derived atm/hydro from real temperature, then re-generated temperature, then re-derived again — a 2-pass iteration that empirically converged.

**Pass 2 designs this as an explicit fixed-point solver from day one.**

### The cycle

```
Atmosphere.Code  ─────► Greenhouse  ─────► Temperature.MeanK
       ▲                                          │
       │                                          ▼
       │                                  HZCOOffsetToTempRange
       │                                          │
       │                                          ▼
   Hydrographics  ◄──── Albedo  ◄────── (tempRange feeds hydro table)
       │                  ▲
       │                  │
       └──────────────────┘  (Hydrographics.Code feeds Albedo)
```

Concrete dependency edges (verified from pass-1 source):

- **`ComputeAlbedo`** reads `Hydrographics.Code` (`temperature_albedo.go:71`). Different hydro bands give different albedo modifiers per WBH p.110. Therefore `Temperature` depends on `Hydrographics`.
- **`ComputeGreenhouseFactor`** reads `Atmosphere.Code` and `Atmosphere.Pressure` (`temperature_greenhouse.go`). Therefore `Temperature` depends on `Atmosphere`.
- **`RederiveAtmosphereHydrographics`** reads `Temperature.MeanK` and may mutate `Atmosphere.Code` via `CheckRunawayGreenhouse` (`temperature_rederive.go`). Therefore `Atmosphere` depends on `Temperature`.
- **`GenerateHydrographics`** takes `tempRange` as input. `tempRange` is derived from `Temperature.MeanK` (post-fixed-point) — though pass 1 used `HZCOOffsetToTempRange` as a preliminary proxy in Stage 5A. Therefore `Hydrographics` depends on `Temperature`.

Three-way bidirectional dependence. Single-pass evaluation in any order produces wrong answers for non-trivial worlds. Pass 1 recovered with two iterations.

### Pass-2 solver shape

```go
// Pseudocode — to be specified in api-surface.md.
type Climate struct {
    Atmosphere    Atmosphere
    Hydrographics Hydrographics
    Temperature   Temperature
}

func ConvergeClimate(r Roller, body, sys, parent) Climate {
    // 1. Initial atm/hydro from HZ-offset proxy (pass 1's Stage 5A).
    // 2. Loop:
    //    a. Compute Temperature(atm, hydro, body, sys, parent).
    //    b. Re-derive Atmosphere from Temperature (CheckRunawayGreenhouse, ScaleHeight).
    //    c. Re-derive Hydrographics from Temperature (tempRange-driven table).
    //    d. If atm.Code, hydro.Code, and Temperature.MeanK are stable, return.
    // 3. Cap iterations at N (pass 1 used 2; pass 2 picks a hard cap and asserts convergence).
}
```

Convergence cap: pass 1 used N=2 and it sufficed for the worked examples. Pass 2 starts at N=3 and asserts convergence within the cap; if a fixture fails, the model is wrong and the failure should be loud.

### Eligibility

Climate runs for HZ-orbit terrestrials and HZ-planet moons. Vacuum worlds, gas giants, and belts skip the cycle (no atmosphere). Non-HZ terrestrials get a degenerate climate with no atmosphere; pass 2 represents this as a `Climate` with nil-pointer atmosphere/hydrographics, mirrored into the body's `Has*()` predicates.

## Stage 6: Atmosphere taint typology (WBH pp.81–90)

Forward-only. Runs after climate has converged.

**Inputs:** post-fixed-point `Atmosphere`, `OxygenPartialPressure`.

**Procedures:**

- `PromoteOxygenTaint(atmCode, ppO2)` — atm 5/6/8 may promote to 4/7/9.
- `RollAllTaints(dp, preseed)` — multi-roll up to 3 taints with reroll on result 10. Eligible codes: 2/4/7/9, A/B/C.
- `RollInsidiousHazards(subtype, isExtremelyDense)` — atm C only.

**Notes:** Mutates `Atmosphere` in place (promotion can change `Code`). The promotion does not feed back into the climate fixed-point because the changes are in atm code-grouping (still oxygen-bearing), not in greenhouse parameters that would shift temperature meaningfully. Pass 2 verifies this assumption with a fixture.

## Stage 7: Geology — TSS + post-TSS temperature update (WBH pp.125–127)

Per-body. Terrestrials get full geology; gas giants get residual heat only; belts skip entirely.

**Inputs (terrestrial):** post-climate body state including `Eccentricity` (post tidal-lock), `Period`, `MassEarth`, `Density`, `Size`, `AgeGyr`, primary-star mass.

**Procedures:**

- `ComputeResidualSeismicStress(dp, ageGyr, isMoon)` → integer.
- `ComputeTidalStressFactor(dp)` → integer (depends on `TidalEffects`, set in Stage 4.5).
- `ComputeTidalHeatingFactor(inputs)` → integer (depends on primary mass, size, eccentricity, distance, period, world mass).
- `TotalSeismicStress = Residual + TSF + THF`.
- `ApplyInherentTempAddition(temp, TSS)` — `T' = ⁴√(T⁴ + TSS⁴)`, mutates Temperature in place.
- Recompute `ScaleHeight = DeriveScaleHeight(T', Gravity)` — atmosphere needs this update.
- `RollTectonicPlates(dp, TSS)` → integer.

**Inputs (gas giant):** `MassEarth`, `AgeGyr`. `ComputeGGResidualHeat` returns InherentTemperatureK; no seismic factors or plates.

### The TSS back-edge into climate

`ApplyInherentTempAddition` mutates `Temperature.MeanK`. In principle this could re-trigger climate's atm/hydro derivation. Pass 1 chose not to chase it, with the rationale (per `project_world_builder_3b_dependency_graph` memory entry):

> For HZ worlds: TSS ~17 alters T by ~0.001K — negligible. For cold/rogue worlds (base T near 25K), TSS can dominate and add 4-5K. The math converges in one pass — no iteration needed. There is no path back into atm/hydro that would re-trigger 3A2b-rederive in practice (the temperature delta is too small to cross any band boundary except in extreme cases worth flagging as `t.Logf` divergences).

Pass-2 decision: **fold TSS into the climate fixed-point's convergence variable.** `Temperature.MeanK` post-TSS is the value atm/hydro should be derived from. The cost is computing partial geology (residual + TSF + THF) inside the convergence loop, but those quantities themselves don't depend on climate state — they depend on body physical, tidal effects, and orbital parameters, all from earlier stages. So:

```text
ConvergeClimate now includes:
   1. Compute pre-TSS Temperature from atm/hydro/body/sys/parent.
   2. Compute TSS factors (Stage 7 partial; atm/hydro-independent).
   3. Apply TSS to Temperature.
   4. Re-derive Atmosphere from Temperature.
   5. Re-derive Hydrographics from Temperature.
   6. Iterate until stable.
```

This is more honest than pass 1's "ignore the back-edge" approach, and is no more expensive in practice (TSS factors are computed once per body in pass 1 anyway; including them inside the loop costs an extra cheap call per iteration). The post-loop residual work is `RollTectonicPlates` (depends on stable TSS) and `RollGGResidualHeat` for gas giants. Both are forward-only.

If the fixture catalog reveals a case where pass 1's "ignore" and pass 2's "include" differ in the IISS form output, pass 2 documents the divergence with rationale and accepts the new value as more correct. The mid-pass strategic-reflection note about "rogue worlds where TSS can dominate" is the canonical case.

## Stage 8: Biology (WBH pp.127–131)

Forward-only. Runs after climate + geology converge.

**Eligibility:** terrestrials with `Atmosphere`; HZ-planet moons with `Atmosphere`. GGs, belts, atm-less terrestrials skip.

**Procedures, strictly ordered:**

1. `RollBiomass(dp, ageGyr)` → biomass rating.
2. If `biomass > 0`:
   a. `RollBiocomplexity(dp, biomass, ageGyr)`.
   b. If `biocomplexity ≥ 8`: `RollNativeSophont`, `RollExtinctSophont`.
   c. `RollBiodiversity(biomass, biocomplexity)`.
   d. `RollCompatibility(dp, biocomplexity, ageGyr)`.
3. `RollTerrestrialResourceRating(dp, bio)` → resource rating.

**Notes:** Resource depends on biology, so order is strict. Pass 1's optional oxygen-atm biomass floor is cut in pass 2 (see `design-intent.md` cuts list). Belts get their own resource rating in `GenerateBeltDetails` (Stage 3).

## Stage 9: Habitability (WBH p.132)

Forward-only. Deterministic — no rolls.

**Eligibility:** terrestrials (atmosphere optional — vacuum worlds get a habitability rating; the atm-0 row applies).

**Inputs:** `SizeCode`, `Atmosphere`, `Hydrographics`, `TidalLock`, `Temperature` (post-TSS), `Physical.Gravity`.

**Procedure:** `ComputeHabitability(body)` returns `Habitability{Rating, Notes}`. Sums DMs from size, atm, hydro, tidal lock, temperature bands, gravity. Pure function.

## Stage 10: System aggregations (WBH pp.58, 132–146)

Run once after every body has converged.

- `computeBaselineN(group, detailed)` per allocation — backfills `StarAllocation.BaselineN`.
- `ShortProfile(sd)` — "G-P-T-N-S" form per WBH p.58.
- `LongProfile(sd)` — "St-N-W-W-S:..." form per WBH p.58.
- `RenderClass0I`, `RenderClass23`, `RenderClass4P` — typed structs (pass-2 design; see `api-surface.md`).
- `pickMainworld(detailed)` — priority chain: native sophonts → highest habitability → highest resource → first in iteration order.

**Notes:** `pickMainworld` admits planets, moons, and belts as candidates. Class IV-P renders only for the auto-picked mainworld; PART P or PART P.B variant is selected by mainworld type.

## Per-body iteration patterns

Three patterns appear across stages:

1. **Per-placement, branching by `BodyKind`.** Sizing, period, body physical, belt details. Code shape: switch on `Body`.
2. **Per-placement and per-moon mirror.** Atmosphere, hydrographics, temperature, day length, axial tilt, tidal lock, surface tidal effects, geology, biology, habitability. **This is where pass 1's silent-zero anti-pattern fired four times.** Pass 2 uses a single `applyBodyProcedures(body)` over an iterator that yields planets and moons uniformly — the moon-vs-planet distinction becomes parameters to a procedure, not a separate code path.
3. **Per-allocation aggregation.** `computeBaselineN`, baseline orbit, spread.

Pass-2 contract: any new procedure whose pass-1 ancestor ran "for each planet, then for each of its moons" must be expressed via the unified iterator. The anti-pattern doc (`anti-patterns.md`) makes this a checklist item.

## System-wide aggregations and shared inputs

These quantities are computed once and consumed everywhere:

- **`AgeGyr`** — primary's age, propagated to companions, used in atmosphere oxygen fraction, body physical, biology, geology, planet eccentricity.
- **`HZCO`** — habitable-zone central orbit per group. Used in atm/hydro DMs, albedo, sizing DMs (gas giant), HZ tagging, baseline computation.
- **`SystemSpread`** — affects gas giant sizing DM, belt details.
- **`totalStellarLuminosity(sys)`** — close-binary group luminosity sum, used in temperature.

Pass 2 keeps these as plain function calls or fields on the `System` / `Group` types, not as global state.

## Pass-2 sequencing implications

The dependency graph reorders pass-1's pipeline as follows:

```
0. GenerateSystem (stars)
1. GenerateSystemPlacement (counts → ... → eccentricities)
2. ApplyDetailFront-end:
    - Sizing
    - Moons
    - Designations
    - Periods
    - HZ tagging
    - Body physical (terrestrials + moons)
    - Belt details (belts)
    - Moon refinement
3. ApplyRotationTilt (3A2a — surface dist gated on later climate; rest forward):
    - Day length, axial tilt, tidal lock, surface tidal effects
    - (Surface distribution deferred until after climate converges)
4. ConvergeClimate per body (the fixed-point solver; includes TSS):
    - Atmosphere, Hydrographics, Temperature, partial Geology
    - Yields final atm/hydro/temp/TSS for downstream stages
5. ApplyTaintTypology (post-climate atm mutations)
6. ApplyClimateDependentFollowups:
    - Surface distribution (now using converged hydro)
    - Tectonic plates (now using converged TSS)
    - GG residual heat (no climate dep, but lives in geology stage)
7. ApplyBiology (terrestrials with atm + HZ-planet moons with atm)
8. ApplyHabitability (terrestrials, atm optional)
9. AggregateSystem (BaselineN backfill, profiles, IISS forms, mainworld pick)
```

Differences from pass-1's pipeline:

- **`ConvergeClimate` replaces 5A-atm/hydro + 5C + 5D + the TSS-temperature-update edge of 5E.** The 2-pass iteration becomes an explicit solver with assertable convergence.
- **Surface distribution moves to step 6** because it depends on hydrographics, which is only stable post-climate. Pass 1 ran it in 5B against the preliminary hydro and never re-ran it; the result happened to be correct because the preliminary and converged hydro were nearly always identical for HZ worlds. Pass 2 makes the dependency explicit.
- **Tectonic plates moves to step 6** for the same reason — depends on stable TSS.
- **Atmosphere taint typology moves before TSS-dependent geology follow-ups** because taint promotion can mutate `atm.Code` and downstream consumers should see the post-promotion code. Pass 1 placed taint between rederive and geology; pass 2 keeps that ordering.

## What this graph commits pass 2 to

1. **One climate solver, not three sequential atm/hydro/temperature passes plus a re-derive recovery.**
2. **One per-body iterator** that walks planets and moons uniformly. The moon-vs-planet distinction is a parameter, not a code path.
3. **Forward-only stages after climate.** Geology's tectonic plates, biology, habitability, system aggregations: all single-pass.
4. **TSS folded into the climate solver's convergence variable** — pass 2 takes the more honest interpretation than pass 1's "ignore the back-edge."
5. **Ordering by data dependency, not by WBH pagination.** Pass-2 sub-projects are sized by what set of fixtures gets green next, not by what chapter is being implemented.

The graph is the structural spine. `api-surface.md` enumerates the public signatures that realize it. `harness.md` enumerates the worked-example fixtures that gate it. `wbh-inconsistencies.md` documents the divergent interpretations that were judgment calls in pass 1 and are committed code in pass 2. `anti-patterns.md` makes the unified-body-iterator a checklist item.
