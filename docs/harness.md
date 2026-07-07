# Harness — fixture catalog

This is the catalog of **named worked-example fixtures** that encode WBH dice scripts to the digit. The fixtures live in `*_test.go` files alongside the procedures they exercise; this doc is the index.

## Coverage strategy

The harness is one of four coverage layers, not the only one. Per-procedure tests + property tests + regression baseline + bulk-sweep together form the actual safety net:

- **Per-procedure tests.** Every `Roll*` / `Generate*` / `Compute*` function has at least one test next to it that drives book-narrated dice and asserts the expected output. These are the bulk of the coverage; many are not "Zed-named" but still encode WBH worked-example dice.
- **Named worked-example fixtures** (this catalog). Where the book threads a multi-procedure example (Sol, Corella, Zed, Zed Prime) with explicit dice, the fixture re-walks that chain end-to-end.
- **Property tests.** 8 invariants × 1000 seeds each (`worlds/property_test.go`). Smoke tests for systemic correctness that catch silent-zero / silent-skip bugs across the population.
- **Markdown regression baseline.** 5 seeds × full Markdown output at `iiss/testdata/seed_*.md`. Refreshed with `go test ./iiss/... -update.regression -run TestRegression`.
- **Bulk-sweep verification.** 10 000-seed sweep via a one-off `cmd/world-builder-bulk` runner; today produces 10 000 successes, zero errors. See [`history/generator-error-catalog.md`](history/generator-error-catalog.md).

A named worked-example fixture is only authored when the book itself narrates the dice chain across multiple procedures. For everything else, per-procedure tests cover the same surface. The original pre-v1.0 plan tracked every example slot in this catalog as 🔴 / 🟢 / ⚠️; post-v1.0 the doc is a catalog of what was authored, not a punch list of what wasn't.

Per `history/spike-findings.md` § Finding 2, full-pipeline named-example gold-script fixtures (like pass-1's abandoned `TestZed_FullDetail`) are deliberately not pursued — they died when the pipeline reordered. Pass-2 favours per-procedure Scripted gold tests where the book narrates dice, plus Seeded shape-invariant fixtures at the façade.

## Status conventions

```text
🟢 green      — fixture exists and passes (named test in *_test.go).
⚠️  divergent — fixture asserts a value that differs from the book; the
                divergence is documented in wbh-inconsistencies.md.
🚧 deferred   — fixture not yet written and not blocking; deferred per
                a specific rationale (no canonical WBH example exists).
```

## Stars (WBH pp.14–35)

The book threads three star-system examples: Sol, Corella, and Zed.

| ID                | WBH page | Status | Asserts                                                                 |
| ----------------- | -------: | :----: | ----------------------------------------------------------------------- |
| `Sol/Terra/p35`   |       35 |   🟢   | Sol primary star physical fields + Terra survey context.                |
| `Sol/SurveyForm`  |       35 |   🟢   | Sol IISS Class 0/I form full cell-by-cell.                              |
| `Corella/p35`     |       35 |   🟢   | Corella binary primary + secondary, HZCO on Aab composite row (3.5).    |
| `Zed/PrimaryOnly` |    17,21 |   🟢   | Zed primary G7 V — type, class, subtype, mass, diameter, age 4.635 Gyr. |
| `Zed/SurveyForm`  |       34 |   🟢   | Zed quintuple IISS Class 0/I form: Aa, Ab, B, Ca, Cb stars + HZCOs.     |

## System Worlds — Placement (WBH pp.36–52)

| ID                            | WBH page | Status | Asserts                                                                                                                                 |
| ----------------------------- | -------: | :----: | --------------------------------------------------------------------------------------------------------------------------------------- |
| `Zed/Counts`                  |       38 |   🟢   | `Counts.Total`, `GasGiants`, `Belts` after Step 0.                                                                                      |
| `Zed/AvailableOrbits`         |    38–43 |   🟢   | Per-group available-orbit intervals. Note: AvailableOrbits emits 5 segments where p.58 shows 4 (AB is separate per pass-1 inline note). |
| `Zed/AllocateByStar`          |    43–44 |   🟢   | `[]StarAllocation` per group.                                                                                                           |
| `Zed/BaselineN`               |    44–45 |   🟢   | Baseline number for the primary group.                                                                                                  |
| `Zed/BaselineOrbit`           |    45–46 |   🟢   | Computed baseline orbit float.                                                                                                          |
| `Zed/Spread`                  |    48–49 |   🟢   | System spread.                                                                                                                          |
| `Zed/AddAnomalous`            |    50–51 |   🟢   | Anomalous-slot insertion, revised counts.                                                                                               |
| `Zed/PlaceOrbitSlots_Aab`     |    49–50 |   🟢   | Aab group ascending-orbit slot order.                                                                                                   |
| `Zed/FullPlacement`           |    36–52 |   🟢   | End-to-end SystemPlacement matches pass-1's TestZed_FullPlacement.                                                                      |
| `Sol/GenerateSystemPlacement` |    36–52 |   🟢   | Single-primary smoke test: non-empty Placements, valid spread.                                                                          |

## Per-body sizing + HZ (WBH pp.53–67)

| ID                    | WBH page | Status | Asserts                                                                                                         |
| --------------------- | -------: | :----: | --------------------------------------------------------------------------------------------------------------- |
| `Zed/Aab IV-d/Sizing` |    58,63 |   ⚠️   | Aab IV-d size = 5 per p.63 form (not S per p.58 table). Pass-2 commits to p.63. See wbh-inconsistencies.md § 2. |

Per-procedure tests cover moon counts, periods (via `MoonPeriodHours` / `Kepler*`), and HZ tagging without named worked-example fixtures.

## Geology — TSS + tectonic plates (WBH pp.125–127)

| ID                              | WBH page | Status | Asserts                                                                                                       |
| ------------------------------- | -------: | :----: | ------------------------------------------------------------------------------------------------------------- |
| `ZedPrime/ResidualSeismic`      |  125,126 |   ⚠️   | Residual=1 (formula table density>1.0 → +2). Book worked example uses +1 → 0. See wbh-inconsistencies.md § 4. |
| `ZedPrime/TidalStressFactor`    |      126 |   🟢   | Zed Prime tidal stress factor. `TestComputeTidalStressFactor_ZedPrime`.                                       |
| `ZedPrime/TidalHeatingFactor`   |      126 |   🟢   | Zed Prime tidal heating factor (parent = Aab IV gas giant). `TestComputeTidalHeatingFactor_ZedPrime`.         |
| `ZedPrime/TotalSeismicStress`   |  125–126 |   🟢   | Sum of three factors. Covered transitively via the three per-factor `_ZedPrime` tests.                        |
| `ZedPrime/InherentTempAddition` |  125,111 |   🟢   | Post-TSS temperature update via ⁴√(T⁴ + TSS⁴). `TestApplyInherentTempAddition_ZedPrime_Negligible`.           |
| `ZedPrime/TectonicPlates`       |      127 |   🟢   | Zed Prime tectonic plate count. `TestRollTectonicPlates_ZedPrime`.                                            |
| `AabIV/GGResidualHeat`          |  126–127 |   🟢   | Aab IV (gas giant parent of Zed Prime) residual heat. `TestComputeGGResidualHeat_ZedPrimeGG`.                 |

## Biology (WBH pp.127–131)

| ID                       | WBH page | Status | Asserts                                                                                                                                                          |
| ------------------------ | -------: | :----: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ZedPrime/Biomass`       |      128 |   🟢   | Zed Prime biomass = A (10). `TestRollBiomass_ZedPrime`.                                                                                                          |
| `ZedPrime/Biocomplexity` |  128–129 |   🟢   | Zed Prime biocomplexity = 5. `TestRollBiocomplexity_ZedPrime`.                                                                                                   |
| `ZedPrime/NativeSophont` |      130 |   🟢   | Sophont present at biocomplexity 9 threshold. `TestRollNativeSophont_Triggers_AtBiocomplexity9`.                                                                 |
| `ZedPrime/Compatibility` |      131 |   ⚠️   | Compatibility = 6 per formula box. Book worked example shows 9 (unsourced "+3"). See wbh-inconsistencies.md § 5. Lifeform profile becomes `"A576"` not `"A579"`. |

## Habitability (WBH p.132)

| ID                      | WBH page | Status | Asserts                                                                                                                                 |
| ----------------------- | -------: | :----: | --------------------------------------------------------------------------------------------------------------------------------------- |
| `ZedPrime/Habitability` |      132 |   ⚠️   | Habitability = 7 (matches book p.141 form). Gravity 0.66 → DM−1 (narrower band wins, footnote ignored). See wbh-inconsistencies.md § 6. |

## Climate passes (WBH pp.79, 81, 96–99, 102, 108–126)

| ID                           | WBH page | Status | Asserts                                                                                                                                 |
| ---------------------------- | -------: | :----: | --------------------------------------------------------------------------------------------------------------------------------------- |
| `ZedPrime/Temperature/MeanK` |  108–115 |   ⚠️   | MeanK 300K. Zed Prime WorstLow = 219K (consistent Near/Far) per pass-2 choice; book sidebar says 230K. See wbh-inconsistencies.md § 3c. |

Per-procedure tests cover atmosphere/hydrographics rolls, albedo computation, greenhouse factor, and the 2-pass climate solver (`ApplyClimatePasses`).

## IISS forms — Class IV-P belt variant (WBH pp.141–142)

| ID                        | WBH page | Status | Asserts                                                                     |
| ------------------------- | -------: | :----: | --------------------------------------------------------------------------- |
| `ZedPrime/Class4P/PartPB` |  141–142 |   🚧   | PART P.B — belt variant. No belt-mainworld worked example in WBH; deferred. |

Per `docs/next-steps.md` § "Historical disposition" A4: closed as GitHub #51 won't-fix. The PART P.B renderer ships and works structurally; a value-asserting fixture would require design fiat without a book reference.

## Façade end-to-end

`Generate(seed)` and `GenerateWithRoller(r)` are the public-API top entry. These fixtures exercise the entire pipeline through the public façade.

**These are `Seeded` + shape-invariant fixtures, not `Scripted` value-exact fixtures** (per `history/spike-findings.md` § Finding 2). Drive N iterations of `roller.NewSeeded(seed)`; assert shape invariants.

| ID             | Status | Asserts                                                                                                                                               |
| -------------- | :----: | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Sol/Generate` |   🟢   | `Generate(seed)` over 100 seeds: non-empty Stars, Placements, IISS Class 0/I form populated; mainworld designation == "A" (Sol single-primary).       |
| `Zed/Generate` |   🟢   | `Generate(seed)` over 100 seeds: every HZ terrestrial has 3-char SAH (no `?`), every body has DayLength + AxialTilt + TidalEffects, mainworld picked. |

## Markdown system output

Pure regression baseline; not a named worked-example. 5 per-seed Markdown snapshots at `iiss/testdata/seed_{1,7,42,100,500}.md` with `TestRegression_MarkdownSeeds` as the drift guard. Refresh via `go test ./iiss/... -update.regression -run TestRegression` after a reviewed change.

## Misuse-path contract tests

Per `api-surface.md` § Misuse-path test pattern, every public function ships with at least one misuse-path test. All entries are covered in `worlds/misuse_test.go` — empirical behaviour for each misuse path is "does not panic; returns sensible zero or error."

| Function                                | Status | Misuse paths                                                                                         |
| --------------------------------------- | :----: | ---------------------------------------------------------------------------------------------------- |
| `RollGasMix`                            |   🟢   | (a) Subtype string passed to columnLetter; (b) atmCode outside [2..9]∪{D,E}; (c) empty columnLetter. |
| `RollAtmoCode`                          |   🟢   | (a) SizeCode "0"; (b) negative offset.                                                               |
| `RollTotalPressure`                     |   🟢   | (a) atmCode outside table; (b) Subtype required when code 11/12.                                     |
| `RollOxygenFraction`                    |   🟢   | (a) negative ageGyr.                                                                                 |
| `RollCorrosiveInsidiousSubtype`         |   🟢   | (a) atmCode not 11/12; (b) HZCO ≤ 0.                                                                 |
| `GenerateBodyPhysical`                  |   🟢   | (a) SizeCode "S"; (b) DiameterKm ≤ 0; (c) negative ageGyr.                                           |
| `GenerateBeltDetails`                   |   🟢   | (a) SizeCode not "0"; (b) negative ageGyr.                                                           |
| `GenerateHydrographics`                 |   🟢   | (a) atm.Code 0 with non-degenerate inputs; (b) tempRange invalid.                                    |
| `ApplyClimatePasses`                    |   🟢   | (a) non-HZ body with HZ flag set; (b) GG passed (should skip).                                       |
| `RollBiomass`                           |   🟢   | (a) body without atmosphere; (b) negative ageGyr.                                                    |
| `RollCompatibility`                     |   🟢   | (a) biocomplexity 0 (should not be called); (b) atm code not in DM table.                            |
| `ComputeHabitability`                   |   🟢   | (a) GG body (returns zero rating with note); (b) body with no temperature.                           |
| `pickMainworld` (via `AggregateSystem`) |   🟢   | (a) empty Universe.                                                                                  |
| `MarkdownClass0I` etc.                  |   🟢   | (a) zero-value form (asserts no panic; output may be partial).                                       |

## Property-test fixtures

8 invariants over 1000 seeds each. Smoke tests for systemic correctness; fire when a procedure silently does nothing for a class of bodies.

| ID                                 | Status | Asserts                                                                                                                            |
| ---------------------------------- | :----: | ---------------------------------------------------------------------------------------------------------------------------------- |
| `RandomSystem/GenerateCompletes`   |   🟢   | `TestProperty_GenerateCompletes` — Generate completes (or fails with Special-Circumstances) for every seed in 0..999.              |
| `RandomSystem/HZBodyHasClimate`    |   🟢   | `TestProperty_HZBodyHasClimate` — every HZ terrestrial / HZ-planet moon has atm + hydro + temp populated.                          |
| `RandomSystem/MoonsHaveBodies`     |   🟢   | `TestProperty_MoonsHaveBodies` — every Child has `Kind == BodyMoon`, populated Designation, Parent set. Anti-pattern A.1 sentinel. |
| `RandomSystem/MainworldExists`     |   🟢   | `TestProperty_MainworldExists` — non-empty mainworld designation when terrestrial/moon/belt candidates exist.                      |
| `RandomSystem/BiomassImpliesAtm`   |   🟢   | `TestProperty_BiomassImpliesAtm` — every body with `Biology.Biomass > 0` has Atmosphere.                                           |
| `RandomSystem/GGHasMass`           |   🟢   | `TestProperty_GGHasMass` — every gas giant has `MassEarth > 0`.                                                                    |
| `RandomSystem/MoonsHaveOrbitPD`    |   🟢   | `TestProperty_MoonsHaveOrbitPD` — every retained moon has Stage-3-populated `OrbitPD > 0`. Finer-grained silent-zero sentinel.     |
| `RandomSystem/ScaleHeightPositive` |   🟢   | `TestProperty_ScaleHeightPositive` — every body with both Atmosphere and Physical has `Atmosphere.ScaleHeight > 0`.                |

## How fixtures evolve

When a worked example reveals a book inconsistency, the resolution lands in `wbh-inconsistencies.md`, the fixture status flips to ⚠️, and the asserted value matches the chosen interpretation. The book's printed value goes in a comment.

When the implementer chooses to author a new named worked-example fixture (vs. relying on per-procedure coverage), add the row to the appropriate section above.
