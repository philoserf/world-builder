# Plan — clean every run

**Date:** 2026-05-11
**Companion doc:** [`generator-error-catalog.md`](generator-error-catalog.md) (the 10 000-seed sweep that motivated this plan)

## Intent

Eventually, every invocation of `cmd/wbh` produces a real, fully-formed system — Class 0/I + Class II/III + Class IV-P forms, no errors, no degraded output. Today, a 10 000-seed sweep returns 676 errors (6.76%), almost all rolling onto post-stellar primaries or special-primary kinds that require the WBH Special Circumstances chapter (pp.147+) — which is and remains out of project scope.

We don't yet know exactly how to reconcile "always serve a real system" with "don't implement pp.147+." The contract decision (Phase 2 below) is the load-bearing question. This plan executes the steps we already know are right while keeping the intent in mind.

## Phases

### Phase 0 — kill the non-scope errors (~16 of 676)

These errors are not Special Circumstances and must be fixed regardless of which contract we land on.

- **0a. The `\x000` descriptor bug (seed 6724, 1 error).** A NUL byte reaches the MAO lookup where a star descriptor should be. Root-cause and fix.
- **0b. Class IV / VI sizing-table gaps (15 errors across 5 subtypes: M0, M9, K5×2, F0).** WBH p.19 vs p.42 inter-table inconsistency. Apply the book-endorsed interpolation rule ("a G7 is 2/5 of the difference between G5 and K0") to fill the missing rows.

After Phase 0: the sweep returns ~660 errors, all of them typed-by-eyeball as Special Circumstances.

### Phase 1 — type the errors (closes issue #45)

Introduce typed sentinels in `worlds/`:

- `ErrSpecialCircumstances` — post-stellar primary, special-primary dispatch unimplemented, companion-of-giant requiring Plan 3+ MAO.
- `ErrUnimplementedDispatch` — for the "X dispatch not yet implemented" variants where the divergence is clearly unimplemented-stage (peculiar, giants, nebula/cluster age formulas).

Migrate the three existing string-match classifiers (`worlds/property_test.go`, `worlds/generate_test.go`, `iiss/regression_test.go`) to `errors.Is`. Phase 2 needs this predicate to know which errors are "advance" vs "abort."

Closes [#45](https://github.com/philoserf/world-builder/issues/45).

### Phase 2 — user-facing contract (open)

After Phase 1, the library returns typed errors. The CLI needs a strategy for what to do when the library returns `ErrSpecialCircumstances`. Three options sketched:

- **Re-roll forward.** `-seed N` becomes a starting hint; if N is out-of-scope, advance to N+1, N+2, … until in-scope. Output reports the actual seed used. Simple; users see the seed they passed silently shift.
- **N-th in-scope semantics.** `-seed N` means "the Nth in-scope system." Internally counts from seed 1 and returns the Nth in-scope roll. Cleaner contract; more internal bookkeeping; more expensive for large N.
- **Time-seeded only.** Drop `-seed N` from the CLI default; reserve it for test fixtures. Production runs use os time and roll until in-scope. Lowest fidelity to current contract; simplest implementation.

Decision deferred. Do not implement until chosen.

### Phase 3 — verify the loop is closed

Re-run the 10 000-seed sweep. Target: 10 000 / 10 000 produce a real, fully-rendered system. Add seed 6724 and the 5 Class IV/VI gap seeds as permanent regression fixtures so the next sweep stays clean.

## Out of scope for this plan

- Implementing WBH pp.147+ (Special Circumstances). Stated as out-of-scope in CLAUDE.md; remains so.
- Restricting the WBH stellar table to in-scope primaries. Violates WBH fidelity.
- Changing the Generate / pipeline API beyond what Phase 1 requires.
