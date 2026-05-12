# Plan — clean every run

**Date:** 2026-05-11 (updated 2026-05-12)
**Status:** Phases 0 + 1 complete. Phase 2 (user-facing contract) open. Phase 3 verification complete.
**Companion doc:** [`generator-error-catalog.md`](generator-error-catalog.md) (the 10 000-seed sweep that motivated this plan)

## Intent

Eventually, every invocation of `cmd/wbh` produces a real, fully-formed system — Class 0/I + Class II/III + Class IV-P forms, no errors, no degraded output. Today, a 10 000-seed sweep returns 676 errors (6.76%), almost all rolling onto post-stellar primaries or special-primary kinds that require the WBH Special Circumstances chapter (pp.147+) — which is and remains out of project scope.

We don't yet know exactly how to reconcile "always serve a real system" with "don't implement pp.147+." The contract decision (Phase 2 below) is the load-bearing question. This plan executes the steps we already know are right while keeping the intent in mind.

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

### Phase 2 — user-facing contract (open)

After Phase 1, the library returns typed errors. The CLI needs a strategy for what to do when the library returns `ErrSpecialCircumstances`. Three options sketched:

- **Re-roll forward.** `-seed N` becomes a starting hint; if N is out-of-scope, advance to N+1, N+2, … until in-scope. Output reports the actual seed used. Simple; users see the seed they passed silently shift.
- **N-th in-scope semantics.** `-seed N` means "the Nth in-scope system." Internally counts from seed 1 and returns the Nth in-scope roll. Cleaner contract; more internal bookkeeping; more expensive for large N.
- **Time-seeded only.** Drop `-seed N` from the CLI default; reserve it for test fixtures. Production runs use os time and roll until in-scope. Lowest fidelity to current contract; simplest implementation.

Decision deferred. Do not implement until chosen.

### Phase 3 — verify the loop is closed ✅ (verification done; permanent fixtures at unit level)

Final 10 000-seed sweep (2026-05-12): 663 / 663 errors classify as Special Circumstances via `errors.Is`; zero untyped real bugs. Verification complete.

Permanent regression fixtures live at the unit level next to each fix rather than as seed-based golden tests (seed-based fixtures break when dice ordering changes upstream):

- `worlds/available_orbits_test.go::TestMAO_Protostar`
- `stars/peculiar_test.go::TestGeneratePrimaryAtClass_IV_M_to_K`
- `stars/peculiar_test.go::TestGeneratePrimaryAtClass_VI_F_to_G`

Phase 3's full target (10 000 / 10 000 produce a real system) is blocked on Phase 2's contract decision — until then, "clean" means "every error is a typed, expected Special-Circumstances class."

## Out of scope for this plan

- Implementing WBH pp.147+ (Special Circumstances). Stated as out-of-scope in CLAUDE.md; remains so.
- Restricting the WBH stellar table to in-scope primaries. Violates WBH fidelity.
- Changing the Generate / pipeline API beyond what Phase 1 requires.
