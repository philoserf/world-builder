# Plan — clean every run

**Date:** 2026-05-11 (updated 2026-05-12)
**Status:** ✅ **Complete.** All sub-phases shipped; 10000-seed sweep shows 100% successes, zero errors.
**Companion doc:** [`generator-error-catalog.md`](generator-error-catalog.md) (the 10 000-seed sweep that motivated this plan)

## Intent

Every invocation of `cmd/wbh` produces a real, fully-formed system — Class 0/I + Class II/III + Class IV-P forms, no errors, no degraded output.

## Why we're here (Phase 2 reframe, 2026-05-12)

The original plan assumed the 6.63% out-of-scope error rate was structural — that the WBH dice unavoidably roll onto primaries (white dwarfs, neutron stars, brown dwarfs, protostars, nebulae) that need the Special Circumstances chapter (pp.219+, out of scope), and we'd have to work around it at the output layer. **That premise was wrong.** The WBH itself provides three explicit Referee options that bypass the issue entirely, none of them invoking pp.219+:

- **WBH p.15 — column toggle.** The Star Type Determination table has four sub-columns (Special / Unusual / Giants / Peculiar). The book's own words: "If the Referee chooses to include [brown dwarfs and white dwarfs] as primary stars in some systems, they should choose the **Unusual column**, if not, the **Special column** does not include these stars as primary." Special-column cells are Class VI / IV / III / Giants only — all mainstream stars covered by pp.14-146.
- **WBH p.16 — 1D Peculiar fallback.** "If this result occurs for a primary star, the Referee may choose to **ignore the result and roll again** or instead resolve the unusual result with 1D, with a result of 1–5 meaning neutron star and 6 resulting in black hole." Two book-endorsed fallbacks for the Peculiar cell.
- **WBH p.27 — companion-of-giant orbit.** "Companions of giants (Ia, Ib, II or III) have Orbit# equal to **1D × MAO of the Primary star** (see page 39)." Fully specified; we just deferred implementation.

Current code uses the Unusual column, has no Peculiar fallback, and surfaces the companion-of-giant case as an explicit error. None of these were forced choices — they were early-pass simplifications that left the harder paths for later. Phase 2 finishes the job.

## Phases

### Phase 0 — kill the non-scope errors ✅

These errors were not Special Circumstances and had to be fixed regardless of which contract we land on.

- **0a. The `\x000` descriptor bug (seed 6724, 1 error).** ✅ Root cause: protostar primary has no spectral type; `MAO()` fell through to the p.39 table lookup with a zero-valued `SpectralType`. Fix: `lacksP39MAORow` predicate at the MAO gate. Commit `1d4bf6c`.
- **0b. Class IV / VI sizing-table gaps (15 errors).** ✅ Root cause was not a book inter-table inconsistency — it was the WBH p.16 letter constraints (`M→K` for Class IV, `F→G` for Class VI, plus K-IV-subtype>4 shift) being applied only in `RollSubtype`, not on the rolled letter. Fix landed on four roll paths (`generatePrimaryAtClass`, `generateLesser`, `generateRandom`, `generateSibling`) plus one-sided-missing graceful interpolation in `InterpolateClassRow`. Commit `c5aabb2`.

Result: sweep at 663 errors, all classifying as Special Circumstances.

### Phase 1 — type the errors ✅ (closed issue #45)

Introduced `stars.ErrSpecialCircumstances` as the umbrella sentinel. Every error indicating WBH pp.147+ is required wraps it via `%w`:

- `stars.ErrSpecialPrimary` (existing; rewrapped)
- `stars.ErrSpecialPrimaryGiantsDispatch` (new)
- `stars.ErrCompanionOfGiantMAO` (new)
- `stars/ages.go` "no age formula" (inline-wrapped)
- `worlds.ErrPostStellarPrimaryUnsupported` (rewrapped)

Three test classifiers migrated from `strings.Contains` lists to a single `errors.Is(err, stars.ErrSpecialCircumstances)` check. Commit `90b64d2`. Closed [#45](https://github.com/philoserf/world-builder/issues/45).

### Phase 2 — adopt the book's Referee defaults ✅

Direct implementation of the three WBH Referee options. Seed determinism preserved; no re-roll / Nth-in-scope / time-seeded gymnastics required. Sub-phases shipped:

- **2a (85d80a3)** — Primary column Unusual → Special. `PeculiarPathSpecial` constant; `GenerateSystemOpts.PeculiarColumn` field (zero value = Special). Adjacent fix: `RollPrimaryTypeAndClassDMPlus1` for class-redirect re-rolls per WBH p.16.
- **2c (f5ae866)** — Special-column Giants dispatch. `RollGiantClass` + `generatePrimaryAtClass` for III/II/Ib/Ia.
- **2d (2d6008b)** — Companion-of-giant orbit per WBH p.27: `Orbit# = 1D × MAO(primary)`. `MAO` callback on `GenerateSystemOpts`.
- **2f (25751bb)** — Post-stellar group members contribute 0 to MAO. Allows WBH-allowed BD/D companions (p.29) without invoking Special Circumstances for orbital placement.
- **2b (12d714d)** — Companion `Random` descriptor's "Special" cell dispatches through the Special column. `generateRandomSpecial` mirrors `generateSpecialPrimary` for the companion context.

Result: **10 000 / 10 000 seeds produce a real, fully-rendered system** with no contract change. Seed determinism preserved.

#### Deferred

- **2e — 1D Peculiar fallback as safety net.** Originally scoped as a catch-all for any code path that might reach the Peculiar cell. With 2a-2f shipped, the default operation reaches zero errors without it. Park as a future safety net for opt-in Unusual-column use (which currently still has the original error behavior).
- **Companion column opt-in for Unusual.** 2b hardcodes Special for `generateRandomSpecial`. Threading `PeculiarColumn` through `GenerateCompanionStar` for explicit Unusual-companions support is a future extension (needs API change to `GenerateCompanionStar`).

#### Configurations on top of Phase 2

The original three options (re-roll forward / Nth-in-scope / time-seeded) remain _available as opt-in modifiers_ for users who explicitly choose `GenerateSystemOpts.PeculiarColumn = PeculiarPathUnusual` for verisimilitude. They are not the primary mechanism. If/when these are exposed, document via CLI flags.

### Phase 3 — verify the loop is closed ✅

Final 10 000-seed sweep (2026-05-12): **10 000 / 10 000 successes, zero errors.**

Permanent unit-level regression fixtures:

- `worlds/available_orbits_test.go::TestMAO_Protostar` (0a)
- `stars/peculiar_test.go::TestGeneratePrimaryAtClass_IV_M_to_K` (0b)
- `stars/peculiar_test.go::TestGeneratePrimaryAtClass_VI_F_to_G` (0b)
- `stars/peculiar_test.go::TestRollSpecialPrimary_Special_ClassRedirects` (2a)
- `stars/system_test.go::TestGenerateSystem_SpecialPrimary_GiantsCell` (2c)

See `generator-error-catalog.md` for the full per-phase journey.

## Out of scope for this plan

- Implementing WBH pp.219+ (Special Circumstances chapter — empty-hex rogue objects, full BD/D/NS/BH primary system rules). Stated as out-of-scope in CLAUDE.md; remains so. Phase 2's reframe shows we don't need it for clean operation.
- Changing the Generate / pipeline API in disruptive ways. Adding a `Referee` opt or similar for the column toggle is in scope; rewiring the seed contract is not.
- Restricting the WBH stellar table values themselves. Phase 2 changes _which column_ we read; the table cells are book-faithful.
