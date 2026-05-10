# Pass 2 — Harness Fixture Catalog

The harness is the gate: every WBH worked example is encoded as a failing dice-scripted fixture **before any generation code is written**. Implementation order is "whichever fixture is closest to green next." The harness becomes the spec; passing the harness is what shipping pass 2 means.

This document is the catalog. Each entry has the fixture name, the WBH page citation, what it asserts, the status (red until implemented), and any pass-1 carry-forward notes. The actual fixture code lives in `*_test.go` files alongside the procedures it exercises.

Pass 1's worked-example tests are the seed for this catalog. Several pass-1 entries become pass-2 fixtures verbatim (the dice scripts are gold and ported as-is per `design-intent.md`). Pass 2 adds entries for in-line WBH examples that pass 1 did not encode — those are where pass-2 fidelity exceeds pass-1 fidelity.

## Status conventions

```text
🔴 red       — fixture written, asserting against unimplemented stubs
🟢 green     — fixture passing
⚠️  divergent — fixture asserts a value that differs from the book; the
                 divergence is documented in wbh-inconsistencies.md
🚧 deferred  — fixture not yet written (and not blocking; deferred per
                 some out-of-scope rationale)
```

All entries start at 🔴 when the harness lands. Implementation cycles flip them to 🟢. ⚠️ entries pass with a value that pass-2 has chosen to commit to per `wbh-inconsistencies.md`; the ⚠️ marker reminds future readers the value diverges from the printed book.

## Stars (WBH pp.14–35)

The book threads three star-system worked examples: Sol (single primary, our solar system), Corella (binary G2 V + G8 V), and Zed (quintuple, the canonical example).

| ID                | WBH page | Status | Asserts                                                                 |
| ----------------- | -------: | :----: | ----------------------------------------------------------------------- |
| `Sol/Terra/p35`   |       35 |   🔴   | Sol primary star physical fields + Terra survey context.                |
| `Sol/SurveyForm`  |       35 |   🔴   | Sol IISS Class 0/I form full cell-by-cell.                              |
| `Corella/p35`     |       35 |   🔴   | Corella binary primary + secondary, HZCO on Aab composite row (3.5).    |
| `Zed/PrimaryOnly` |    17,21 |   🔴   | Zed primary G7 V — type, class, subtype, mass, diameter, age 4.635 Gyr. |
| `Zed/SurveyForm`  |       34 |   🔴   | Zed quintuple IISS Class 0/I form: Aa, Ab, B, Ca, Cb stars + HZCOs.     |

Pass-1 carry-forwards: `TestSolTerra_p35`, `TestSolTerra_SurveyForm_p35`, `TestCorella_SurveyForm_p35`, `TestZedPrimaryOnly_p17_p21`, `TestZed_SurveyForm_p34`. Dice scripts ported verbatim.

## System Worlds — Placement (WBH pp.36–52)

Zed is the principal worked example through this chapter. Sol provides smoke-test coverage as a single-star contrast.

| ID                            | WBH page | Status | Asserts                                                                                                                                 |
| ----------------------------- | -------: | :----: | --------------------------------------------------------------------------------------------------------------------------------------- |
| `Zed/Counts`                  |       38 |   🔴   | `Counts.Total`, `GasGiants`, `Belts` after Step 0.                                                                                      |
| `Zed/AvailableOrbits`         |    38–43 |   🔴   | Per-group available-orbit intervals. Note: AvailableOrbits emits 5 segments where p.58 shows 4 (AB is separate per pass-1 inline note). |
| `Zed/AllocateByStar`          |    43–44 |   🔴   | `[]StarAllocation` per group.                                                                                                           |
| `Zed/BaselineN`               |    44–45 |   🔴   | Baseline number for the primary group.                                                                                                  |
| `Zed/BaselineOrbit`           |    45–46 |   🔴   | Computed baseline orbit float.                                                                                                          |
| `Zed/Spread`                  |    48–49 |   🔴   | System spread.                                                                                                                          |
| `Zed/AddAnomalous`            |    50–51 |   🔴   | Anomalous-slot insertion, revised counts.                                                                                               |
| `Zed/PlaceOrbitSlots_Aab`     |    49–50 |   🔴   | Aab group ascending-orbit slot order.                                                                                                   |
| `Zed/FullPlacement`           |    36–52 |   🔴   | End-to-end SystemPlacement matches pass-1's TestZed_FullPlacement.                                                                      |
| `Sol/GenerateSystemPlacement` |    36–52 |   🔴   | Single-primary smoke test: non-empty Placements, valid spread.                                                                          |

Pass-1 carry-forwards: every `TestZed_*` from `worlds/worked_examples_test.go` and `TestSol_GenerateSystemPlacement`. Dice scripts ported.

## Per-body sizing + moons + designations + periods + HZ (WBH pp.53–67)

| ID                       | WBH page | Status | Asserts                                                                                                         |
| ------------------------ | -------: | :----: | --------------------------------------------------------------------------------------------------------------- |
| `Zed/Aab IV-d/Sizing`    |    58,63 |   ⚠️   | Aab IV-d size = 5 per p.63 form (not S per p.58 table). Pass-2 commits to p.63. See wbh-inconsistencies.md § 2. |
| `Zed/MoonCounts`         |       55 |   🔴   | Moon counts per planet match the book.                                                                          |
| `Zed/Periods`            |       30 |   🔴   | Per-body orbital period via Kepler. GG-mass-≥-100 ⊕ inclusion of own mass.                                      |
| `Zed/HZTagging`          |       58 |   🔴   | HZ flag set for bodies in HZCO ± 1.0.                                                                           |
| `ZedPrime/OrbitalPeriod` |    75–77 |   🔴   | Zed Prime moon period via MoonPeriodHours.                                                                      |

Pass-1 carry-forward: `TestZedPrime_OrbitalPeriod`.

## Body Physical + Belt Details (WBH pp.69–78, 91–93)

| ID                      | WBH page | Status | Asserts                                                               |
| ----------------------- | -------: | :----: | --------------------------------------------------------------------- |
| `Sol/Terra/Physical`    |    69–77 |   🔴   | Terra composition, density 5.51, gravity 1.0, mass 1.0 ⊕.             |
| `Zed/AabPI/BeltProfile` |    91–93 |   🔴   | Aab PI belt: span, composition, bulk, resource, sig-size counts.      |
| `Zed/CabPI/BeltProfile` |    91–93 |   🔴   | Cab PI belt — secondary belt example, different stellar context.      |
| `ZedPrime/Physical`     |    69–77 |   🔴   | Zed Prime body physical: density 1.03, gravity 0.66, mass derivation. |

Pass-1 carry-forwards: `TestSol_TerraPhysicalProfile`, `TestZed_AabPI_BeltProfile`, `TestZed_CabPI_BeltProfile`.

## 3A2a — Rotation/Tilt/Tide (WBH pp.100–108)

| ID                             | WBH page | Status | Asserts                                                            |
| ------------------------------ | -------: | :----: | ------------------------------------------------------------------ |
| `ZedPrime/DayLength`           |  100–105 |   🔴   | Zed Prime sidereal hours, solar hours, year days, IsLong flag.     |
| `ZedPrime/AxialTilt`           |   105–06 |   🔴   | Zed Prime axial tilt — basic vs. extreme branch verification.      |
| `ZedPrime/TidalLock`           |  106–107 |   🔴   | Zed Prime tidal-lock state. Includes natural-12 verification path. |
| `ZedPrime/SurfaceTidalEffects` |      107 |   🔴   | Zed Prime tidal effects per zone.                                  |

Pass-1 ran these as part of `TestZed_FullDetail_3A2b`; pass 2 splits them into per-procedure fixtures so a regression in any one is named precisely.

## Climate fixed-point (WBH pp.79, 81, 96–99, 102, 108–126)

The single most important fixture group. Pass-1's `TestZed_FullDetail_3A2b` is the end-to-end gate; pass 2 adds per-stage assertions inside the convergence.

| ID                                | WBH page | Status | Asserts                                                                                                                                 |
| --------------------------------- | -------: | :----: | --------------------------------------------------------------------------------------------------------------------------------------- |
| `ZedPrime/Atmosphere/Initial`     |    79,81 |   🔴   | First-pass atmosphere code 6, oxygen partial pressure.                                                                                  |
| `ZedPrime/Hydrographics/Initial`  |    96-99 |   🔴   | First-pass hydrographics code 6.                                                                                                        |
| `ZedPrime/Temperature/MeanK`      |  108–115 |   ⚠️   | MeanK 300K. Zed Prime WorstLow = 219K (consistent Near/Far) per pass-2 choice; book sidebar says 230K. See wbh-inconsistencies.md § 3c. |
| `ZedPrime/Temperature/Albedo`     |      110 |   🔴   | Albedo 0.33. Hyd modifier formula `(2D − 4) × 0.03`; harness scripts dice 7 to reproduce. See wbh-inconsistencies.md § 3a.              |
| `ZedPrime/Temperature/Greenhouse` |      111 |   🔴   | Greenhouse factor for atm 6 + Zed Prime pressure.                                                                                       |
| `ZedPrime/Climate/Converged`      |   79+108 |   🔴   | Post-converge atm/hydro/temp stable values; convergence ≤ 3 iterations.                                                                 |
| `ZedPrime/Climate/RunawayPath`    |       79 |   🚧   | Runaway-greenhouse boiling-only path test; pass 1 deferred MVP.                                                                         |

Pass-1 carry-forward: `TestZed_FullDetail_3A2b` — refactored into the above per-procedure fixtures plus a top-level convergence assertion.

## Atmosphere taint typology (WBH pp.81–90)

Pass 1 added these as separate sub-project tests; pass 2 keeps them as named fixtures.

| ID                       | WBH page | Status | Asserts                                                           |
| ------------------------ | -------: | :----: | ----------------------------------------------------------------- |
| `AabVd/TaintProfile`     |       85 |   🔴   | Aab V-d multi-roll taint subtype + severity + persistence.        |
| `AabVb/ExoticIrritant`   |       88 |   🔴   | Aab V-b exotic-irritant atmosphere taint chain.                   |
| `AaBVI/CorrosiveProfile` |       90 |   🔴   | Aa B VI corrosive atmosphere subtype + insidious hazard branches. |

Pass-1 carry-forwards: `TestAabVd_TaintProfile_p85`, `TestAabVb_ExoticIrritant_p88`, `TestAaBVI_CorrosiveProfile_p90`.

## Geology — TSS + tectonic plates (WBH pp.125–127)

| ID                              | WBH page | Status | Asserts                                                                                                       |
| ------------------------------- | -------: | :----: | ------------------------------------------------------------------------------------------------------------- |
| `ZedPrime/ResidualSeismic`      |  125,126 |   ⚠️   | Residual=1 (formula table density>1.0 → +2). Book worked example uses +1 → 0. See wbh-inconsistencies.md § 4. |
| `ZedPrime/TidalStressFactor`    |      126 |   🔴   | Zed Prime tidal stress factor.                                                                                |
| `ZedPrime/TidalHeatingFactor`   |      126 |   🔴   | Zed Prime tidal heating factor (parent = Aab IV gas giant).                                                   |
| `ZedPrime/TotalSeismicStress`   |  125–126 |   🔴   | Sum of three factors.                                                                                         |
| `ZedPrime/InherentTempAddition` |  125,111 |   🔴   | Post-TSS temperature update via ⁴√(T⁴ + TSS⁴). Tested inside ConvergeClimate.                                 |
| `ZedPrime/TectonicPlates`       |      127 |   🔴   | Zed Prime tectonic plate count.                                                                               |
| `AabIV/GGResidualHeat`          |  126–127 |   🔴   | Aab IV (gas giant parent of Zed Prime) residual heat. No seismic factors.                                     |

Pass-1 carry-forward: covered by `TestZed_FullDetail_3A2b`'s 3B-geology block; pass 2 splits into per-factor fixtures.

## Biology (WBH pp.127–131)

| ID                        | WBH page | Status | Asserts                                                                                                                                                          |
| ------------------------- | -------: | :----: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ZedPrime/Biomass`        |      128 |   🔴   | Zed Prime biomass = A (10).                                                                                                                                      |
| `ZedPrime/Biocomplexity`  |  128–129 |   🔴   | Zed Prime biocomplexity = 5.                                                                                                                                     |
| `ZedPrime/NativeSophont`  |      130 |   🔴   | Zed Prime sophont present (book p.141).                                                                                                                          |
| `ZedPrime/Biodiversity`   |      130 |   🔴   | Zed Prime biodiversity = 7.                                                                                                                                      |
| `ZedPrime/Compatibility`  |      131 |   ⚠️   | Compatibility = 6 per formula box. Book worked example shows 9 (unsourced "+3"). See wbh-inconsistencies.md § 5. Lifeform profile becomes `"A576"` not `"A579"`. |
| `ZedPrime/ResourceRating` |      131 |   🔴   | Zed Prime resource rating.                                                                                                                                       |

## Habitability (WBH p.132)

| ID                           | WBH page | Status | Asserts                                                                                                                                 |
| ---------------------------- | -------: | :----: | --------------------------------------------------------------------------------------------------------------------------------------- |
| `ZedPrime/Habitability`      |      132 |   ⚠️   | Habitability = 7 (matches book p.141 form). Gravity 0.66 → DM−1 (narrower band wins, footnote ignored). See wbh-inconsistencies.md § 6. |
| `ZedPrime/HabitabilityNotes` |      132 |   🔴   | Notes string contains all DM contributors with Referee-color codings.                                                                   |

## System aggregations + IISS forms (WBH pp.58, 132–146)

| ID                        | WBH page | Status | Asserts                                                                     |
| ------------------------- | -------: | :----: | --------------------------------------------------------------------------- |
| `Zed/BaselineN_Backfill`  |       58 |   🔴   | Per-allocation BaselineN computed correctly.                                |
| `Zed/ShortProfile`        |       58 |   🔴   | "G-P-T-N-S" form for Zed system.                                            |
| `Zed/LongProfile`         |       58 |   🔴   | "St-N-W-W-S:..." form for Zed system.                                       |
| `Zed/MainworldPick`       |      134 |   🔴   | Mainworld designation = Zed Prime (Aab IV d).                               |
| `Sol/Class0I`             |       35 |   🔴   | Sol/Terra IISS Class 0/I form full cell-by-cell.                            |
| `Zed/Class0I`             |       34 |   🔴   | Zed quintuple IISS Class 0/I form full cell-by-cell.                        |
| `Corella/Class0I`         |       35 |   🔴   | Corella binary IISS Class 0/I form full cell-by-cell.                       |
| `Zed/Class23`             |       63 |   🔴   | Zed IISS Class II/III form full cell-by-cell.                               |
| `ZedPrime/Class4P/PartP`  |  141–142 |   🔴   | Zed Prime IISS Class IV-P PART P (planet/moon mainworld variant).           |
| `ZedPrime/Class4P/PartPB` |  141–142 |   🚧   | PART P.B — belt variant. No belt-mainworld worked example in WBH; deferred. |

Pass-1 carry-forward: `TestZed_FullDetail` (now superseded), `TestZed_FullDetail_3A2b` (acceptance gate).

## Façade end-to-end

`Generate(seed)` and `GenerateWithRoller(r)` are the public-API top entry (`api-surface.md` § The top-level façade). These fixtures exercise the entire pipeline through the public façade, ensuring the top-level shape works as a real caller would call it — not just leaf procedures in isolation.

| ID             | WBH page | Status | Asserts                                                                                                                  |
| -------------- | -------: | :----: | ------------------------------------------------------------------------------------------------------------------------ |
| `Sol/Generate` |        — |   🔴   | `GenerateWithRoller(scriptedSol)` returns Universe with non-empty Stars, Placements, IISS forms. Smoke-level shape only. |
| `Zed/Generate` |        — |   🔴   | `GenerateWithRoller(scriptedZed)` returns Universe matching the per-stage Zed assertions composed end-to-end.            |

These fixtures land in the _initial_ harness commit so the façade signature is exercised from cycle 1, not synthesized at cycle 11. They start red against the unimplemented `GenerateWithRoller` stub; they go green when the pipeline's last stage lands. Per `design-intent.md` § Risks named, this keeps the public-API shape under test from the first commit forward — pass 1's API emerged from real callers, and pass 2 must not lose that constraint by designing the façade in isolation.

## Markdown system output

| ID                   | WBH page | Status | Asserts                                                                                                                                                         |
| -------------------- | -------: | :----: | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Zed/MarkdownGolden` |        — |   🔴   | `MarkdownSystem(zedUniverse)` produces a golden-file equivalent output. Updates require explicit acknowledgement; the test is the canonical Markdown rendering. |

The golden file is `worlds/testdata/zed_markdown_golden.md` (or `iiss/testdata/...` after the package split). Pass 1 had this; pass 2 inherits.

## Misuse-path contract tests

Per `api-surface.md` § Misuse-path test pattern, every public function ships with at least one misuse-path test. The catalog below names the public function and the misuse paths to verify. During stub authoring, a Stance column is added per row capturing whether each misuse path is enforced via panic, error return, or compile-fail (typed parameter). The Stance column is the source of truth for the `api-surface.md` § Errors and misuse contracts commitment that "no procedure does both depending on caller."

| Function                        | Misuse paths                                                                                         |
| ------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `RollGasMix`                    | (a) Subtype string passed to columnLetter; (b) atmCode outside [2..9]∪{D,E}; (c) empty columnLetter. |
| `RollAtmoCode`                  | (a) SizeCode "0"; (b) negative offset.                                                               |
| `RollTotalPressure`             | (a) atmCode outside table; (b) Subtype required when code 11/12.                                     |
| `RollOxygenFraction`            | (a) negative ageGyr.                                                                                 |
| `RollCorrosiveInsidiousSubtype` | (a) atmCode not 11/12; (b) HZCO ≤ 0.                                                                 |
| `GenerateBodyPhysical`          | (a) SizeCode "S"; (b) DiameterKm ≤ 0; (c) negative ageGyr.                                           |
| `GenerateBeltDetails`           | (a) SizeCode not "0"; (b) negative ageGyr.                                                           |
| `GenerateHydrographics`         | (a) atm.Code 0 with non-degenerate inputs; (b) tempRange invalid.                                    |
| `ConvergeClimate`               | (a) non-HZ body with HZ flag set; (b) GG passed (should skip).                                       |
| `RollBiomass`                   | (a) body without atmosphere; (b) negative ageGyr.                                                    |
| `RollCompatibility`             | (a) biocomplexity 0 (should not be called); (b) atm code not in DM table.                            |
| `ComputeHabitability`           | (a) GG body (returns zero rating with note); (b) body with no temperature.                           |
| `pickMainworld`                 | (a) empty bodies slice; (b) all-non-habitable system.                                                |
| `MarkdownClass0I` etc.          | (a) zero-value form (asserts deterministic empty output, no panic).                                  |

Each entry above maps to a fixture in `*_misuse_test.go` files. The pattern: pass the misuse input, assert the documented response (compile-fail via type system, error return, or panic with a specific message — chosen per function in api-surface.md).

## Property-test fixtures

A class of fixture that runs over arbitrary seeds and asserts invariants rather than specific values. Pass 1 had a few; pass 2 keeps and expands them.

| ID                               | Status | Asserts                                                                                         |
| -------------------------------- | :----: | ----------------------------------------------------------------------------------------------- |
| `RandomSystem/Convergence`       |   🔴   | For 1000 random seeds: ConvergeClimate completes within N iterations for every body.            |
| `RandomSystem/HZBodyHasClimate`  |   🔴   | Every body with `HZ=true` and `Kind=BodyTerrestrial` has non-nil atm/hydro/temp post-pipeline.  |
| `RandomSystem/MoonsHaveBodies`   |   🔴   | Every body with non-empty `Children` has those children processed (not silent-zero).            |
| `RandomSystem/MainworldExists`   |   🔴   | Every system with at least one habitable terrestrial yields a non-empty `MainworldDesignation`. |
| `RandomSystem/BiomassImpliesAtm` |   🔴   | Every body with `Biology.Biomass > 0` has non-nil `Atmosphere`.                                 |

These are smoke tests for systemic correctness, not specific WBH values. They fire when a procedure silently does nothing for a class of bodies.

## How fixtures evolve

When a worked example reveals a new book inconsistency, the resolution lands in `wbh-inconsistencies.md`, the fixture status flips to ⚠️, and the asserted value matches the chosen interpretation. The book's printed value goes in a comment.

When a worked example reveals a procedural gap (procedure not yet implemented), the fixture stays 🔴 until implementation lands.

When a worked example reveals a bug in pass-2's design, that's a `design-intent.md` revision — the fixture stays 🔴, the design doc gets an entry, and the next implementation cycle fixes both.

## Implementation-order rule

The harness drives implementation order. After stubs land, the implementer picks the 🔴 fixture closest to green next — meaning, the one whose dependency chain has the fewest other unimplemented procedures. The dependency graph (`dependency-graph.md`) is the authority for "closest to green next."

Suggested order of first cycles (each cycle = 1–3 fixtures green):

1. `Sol/Terra/p35` + `Zed/PrimaryOnly` — Stage 0 stars only, no companion cascade.
2. `Sol/SurveyForm` — Stage 0 + Class 0/I form rendering.
3. `Zed/SurveyForm` + `Corella/SurveyForm` — Stage 0 multi-star + companion logic.
4. `Sol/AvailableOrbits` + `Zed/AvailableOrbits` — Stage 1 partial.
5. `Zed/Counts` + `Zed/AllocateByStar` + `Zed/BaselineN` — Stage 1 cumulative.
6. `Zed/FullPlacement` — Stage 1 end-to-end.
7. `Sol/Terra/Physical` — Stage 3 minimum.
8. `ZedPrime/Physical` + sizing → progress through Stages 2–3.
9. `ZedPrime/Climate/Converged` — the big one. Stages 4–5 in one cycle.
10. Geology, biology, habitability, mainworld pick — one stage per cycle.
11. `Zed/MarkdownGolden` — final integration.

Eleven cycles, ~3–7 days each per `design-intent.md` § Cadence. Soft cap: revisit at week 3 if Stage 5 (climate convergence) is not in sight.

## Closing

The harness is the contract. When every fixture in this catalog is 🟢 (and the ⚠️ entries hold their committed divergent values), pass 2 has reached behavior parity with pass 1 plus the corrections pass 1 documented but did not fix. That is the merge-to-main gate per `design-intent.md` § Behavior parity gate.
