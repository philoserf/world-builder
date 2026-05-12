# Generator error catalog — 10 000-seed sweep

**Original sweep:** 2026-05-11 (pre-cleanup)
**Latest sweep:** 2026-05-12 (post Phase 0 + Phase 1)
**Tool:** one-off `cmd/wbh-bulk/main.go` (parallel goroutines, ~1.5 s wall on 8 cores; removed after each run)
**Range:** seeds 1..10 000 inclusive
**Entry point:** `worlds.Generate(seed)` — same path as `cmd/wbh`

## Current state (post-cleanup)

| Bucket                                  | Count | % of total |
| --------------------------------------- | ----: | ---------: |
| Successes                               |  9337 |     93.37% |
| Errors total                            |   663 |      6.63% |
| ↳ Typed `stars.ErrSpecialCircumstances` |   663 |   **100%** |
| ↳ Real bugs (untyped)                   |     0 |      0.00% |
| Distinct error messages                 |    14 |          — |

Every error now wraps `stars.ErrSpecialCircumstances`; callers classify via a single `errors.Is(err, stars.ErrSpecialCircumstances)` check.

## Errors by category (current)

### 1. Out-of-scope Special-Circumstances primaries (595, 89.7% of errors)

Post-stellar primaries (white dwarf / neutron star / black hole / brown dwarf / pulsar) plus pre-stellar protostars. MAO for these kinds lives in WBH pp.147+. Out of project scope.

| Count | First seed | Message stem                                |
| ----: | ---------: | ------------------------------------------- |
|   284 |         67 | `MAO for group B: post-stellar primary MAO` |
|   226 |         56 | `post-stellar primary MAO`                  |
|    41 |        237 | `MAO for group Bab: …`                      |
|    30 |         22 | `MAO for group C: …`                        |
|    10 |        317 | `MAO for group Cab: …`                      |
|     3 |        855 | `MAO for group D: …`                        |
|     1 |       7047 | `MAO for group Dab: …`                      |

### 2. Special-primary dispatch not implemented (63, 9.5%)

Companion-side and primary-side dispatch through peculiar.go not yet implemented for various special kinds.

| Count | First seed | Message stem                                                                  |
| ----: | ---------: | ----------------------------------------------------------------------------- |
|    51 |         10 | `companion (descriptor "Random"): special primary; dispatch through peculiar` |
|     6 |        403 | `special primary: Special-primary Giants dispatch not yet implemented`        |
|     3 |       1866 | `special primary: kind "nebula" has no age formula`                           |
|     2 |       5629 | `special primary: special primary; dispatch through peculiar`                 |
|     1 |       1994 | `special primary: kind "star_cluster" has no age formula`                     |

### 3. Companion-of-giant primary MAO gap (5, 0.8%)

Known WBH gap — companions of giant primaries require Plan 3+ stellar systems, not implemented.

| Count | First seed | Message stem                                                            |
| ----: | ---------: | ----------------------------------------------------------------------- |
|     4 |        292 | `companion[0] orbit: companion of giant primary requires MAO (Plan 3+)` |
|     1 |       4141 | `companion[1] orbit: companion of giant primary requires MAO (Plan 3+)` |

## What changed from the original sweep

Original (2026-05-11): 676 errors / 20 distinct messages / 1 real bug.

Fixed:

- **Phase 0a** (commit `1d4bf6c`) — protostar primary `\x000` MAO lookup, caught by `lacksP39MAORow` guard. 1 seed → typed Special Circumstances.
- **Phase 0b** (commit `c5aabb2`) — Class IV / VI letter constraints (M→K for IV, F→G for VI, plus K-IV-subtype>4 shift) applied on all four roll paths (`generatePrimaryAtClass`, `generateLesser`, `generateRandom`, `generateSibling`) plus one-sided-missing graceful interpolation in `InterpolateClassRow`. 15 seeds resolved or rerouted.
- **Phase 1** (commit `90b64d2`) — umbrella `stars.ErrSpecialCircumstances` sentinel; 3 test classifiers migrated from `strings.Contains` to `errors.Is`. Closed [#45](https://github.com/philoserf/world-builder/issues/45).

Unit-level regression tests live next to each fix:

- `worlds/available_orbits_test.go::TestMAO_Protostar` — seed 6724 root cause.
- `stars/peculiar_test.go::TestGeneratePrimaryAtClass_IV_M_to_K` — covers seed 212 + family.
- `stars/peculiar_test.go::TestGeneratePrimaryAtClass_VI_F_to_G` — covers seed 6547.

## Reproducing the sweep

`cmd/wbh-bulk/` was a one-off removed after each run. To regenerate:

1. Create `cmd/wbh-bulk/main.go` invoking `worlds.Generate(seed)` in parallel for seeds 1..N, classifying errors via `errors.Is(err, stars.ErrSpecialCircumstances)`.
2. Tabulate counts + first-seed-seen per distinct `err.Error()`.
3. Delete `cmd/wbh-bulk/` after.

Spot-check one seed: `go run ./cmd/wbh -seed 6724 -format short`.

## What's next

Phase 0 + 1 are complete. Phase 2 — the user-facing contract decision (re-roll forward / N-th in-scope / time-seeded — see [`plan-clean-every-run.md`](plan-clean-every-run.md)) — remains open. The typed-error infrastructure from Phase 1 is the foundation for whichever path is chosen.
