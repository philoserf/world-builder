# Plan — clean every run

**Date:** 2026-05-11 (updated 2026-05-12 with Phase 2 reframe)
**Status:** Phases 0 + 1 complete. Phase 2 reframed; sub-phases pending. Phase 3 verification of Phases 0/1 complete.
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

### Phase 2 — adopt the book's Referee defaults (sub-phases pending)

Direct implementation of the three WBH Referee options surfaced above. Seed determinism preserved; no re-roll / Nth-in-scope / time-seeded gymnastics required.

- **2a. Primary column switch — Unusual → Special.** Add `PeculiarPathSpecial` to `stars/peculiar.go`; route `generateSpecialPrimary` through it. Eliminates ~595 post-stellar primary errors and ~53 peculiar/protostar/nebula errors (the Special column's cells are Class VI / IV / III / Giants only).
- **2b. Companion column switch — same Referee choice consistently.** Apply the Special column to companion paths (`generateRandom`, the descriptor-"Random" call site that currently bubbles `ErrSpecialPrimary`). Eliminates the 51 "companion (descriptor 'Random'): special primary; dispatch through peculiar" errors.
- **2c. Special-column Giants dispatch.** The Special-column "Giants" cell (rows 11-12) needs `RollGiantClass` + a fresh Type-column roll at DM+1 + `generatePrimaryAtClass` for III/II/Ib/Ia. `RollGiantClass` already exists. Eliminates 6 errors.
- **2d. Companion-of-giant orbit — WBH p.27 rule.** Replace the `ErrCompanionOfGiantMAO` error at `stars/orbits.go:40` with `Orbit# = 1D × MAO(primary)`. MAO is already implemented. Eliminates 5 errors.
- **2e. 1D Peculiar fallback as safety net.** For any path that still reaches the Peculiar cell (e.g. future opt-in to Unusual column), apply WBH p.16's "1D, 1-5 NS / 6 BH" rule. Catch-all so the umbrella `ErrSpecialCircumstances` is never returned in default operation. (Lower priority — 2a-2d should already reach zero.)

After 2a-2d, target is **10 000 / 10 000 seeds produce a real, fully-rendered system** with no contract change.

#### Configurations on top of Phase 2

The original three options (re-roll forward / Nth-in-scope / time-seeded) remain _available as opt-in modifiers_ for users who explicitly choose the Unusual column for verisimilitude. They are not the primary mechanism. If/when these are exposed, document via CLI flags + GenerateOpts fields.

### Phase 3 — verify the loop is closed

After Phase 0 + 1 (2026-05-12): 663 / 663 errors classify as Special Circumstances via `errors.Is`; zero untyped real bugs.

Permanent unit-level regression fixtures from Phases 0a/0b:

- `worlds/available_orbits_test.go::TestMAO_Protostar`
- `stars/peculiar_test.go::TestGeneratePrimaryAtClass_IV_M_to_K`
- `stars/peculiar_test.go::TestGeneratePrimaryAtClass_VI_F_to_G`

Phase 2 sub-phases will each add their own unit-level fixtures next to the fix. After 2a-2d the bulk-sweep target is **10 000 / 10 000 produce a real system** with no remaining `ErrSpecialCircumstances` in default operation.

## Out of scope for this plan

- Implementing WBH pp.219+ (Special Circumstances chapter — empty-hex rogue objects, full BD/D/NS/BH primary system rules). Stated as out-of-scope in CLAUDE.md; remains so. Phase 2's reframe shows we don't need it for clean operation.
- Changing the Generate / pipeline API in disruptive ways. Adding a `Referee` opt or similar for the column toggle is in scope; rewiring the seed contract is not.
- Restricting the WBH stellar table values themselves. Phase 2 changes _which column_ we read; the table cells are book-faithful.
