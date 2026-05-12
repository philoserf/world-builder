# Generator error catalog — 10 000-seed sweep

**Date:** 2026-05-11
**Tool:** one-off `cmd/wbh-bulk/main.go` (parallel goroutines, ~1.5 s wall on 8 cores; removed after run)
**Range:** seeds 1..10 000 inclusive
**Entry point:** `worlds.Generate(seed)` — same path as `cmd/wbh`

## Outcome

| Bucket                      | Count | % of total | % of errors |
| --------------------------- | ----: | ---------: | ----------: |
| Successes                   |  9324 |     93.24% |           — |
| Errors                      |   676 |      6.76% |     100.00% |
| **Distinct error messages** |    20 |          — |           — |

## Errors by category

### 1. Out-of-scope Special-Circumstances primaries (591, 87.4% of errors)

Post-stellar primaries (white dwarf / neutron star / black hole) require WBH pp. 147+ which is out of project scope per CLAUDE.md. Surfaces at MAO lookup for various host groups.

| Count | First seed | Message                                                                                                         |
| ----: | ---------: | --------------------------------------------------------------------------------------------------------------- |
|   282 |         67 | `placement: available-orbits: MAO for group B: post-stellar primary MAO requires Special Circumstances chapter` |
|   225 |         56 | `placement: available-orbits: post-stellar primary MAO requires Special Circumstances chapter`                  |
|    41 |        237 | `… MAO for group Bab: post-stellar primary MAO`                                                                 |
|    30 |         22 | `… MAO for group C: post-stellar primary MAO`                                                                   |
|    10 |        317 | `… MAO for group Cab: post-stellar primary MAO`                                                                 |
|     3 |        855 | `… MAO for group D: post-stellar primary MAO`                                                                   |
|     1 |       7047 | `… MAO for group Dab: post-stellar primary MAO`                                                                 |

### 2. Special-primary dispatch not implemented (62, 9.2%)

Also Special Circumstances territory — peculiar / giant / nebula / cluster star kinds dispatch but the procedural stage is not implemented.

| Count | First seed | Message                                                                              |
| ----: | ---------: | ------------------------------------------------------------------------------------ |
|    51 |         10 | `stars: companion (descriptor "Random"): special primary; dispatch through peculiar` |
|     6 |        403 | `stars: special primary: Special-primary Giants dispatch not yet implemented`        |
|     3 |       1866 | `stars: special primary: kind "nebula" has no age formula`                           |
|     2 |       5629 | `stars: special primary: special primary; dispatch through peculiar`                 |
|     1 |       1994 | `stars: special primary: kind "star_cluster" has no age formula`                     |

### 3. Class IV / VI sizing-table gaps (15, 2.2%)

WBH p.19 vs p.42 inter-table inconsistency — specific spectral subtypes are missing from one of the sizing tables. Documented in MEMORY as a book inconsistency.

| Count | First seed | Message                                       |
| ----: | ---------: | --------------------------------------------- |
|     8 |        212 | `stars: special primary: M0 class IV missing` |
|     4 |       1130 | `stars: special primary: K5 class IV missing` |
|     1 |       4342 | `stars: companion mass: K5 class IV missing`  |
|     1 |       6426 | `stars: special primary: M9 class IV missing` |
|     1 |       6547 | `stars: special primary: F0 class VI missing` |

### 4. Companion-of-giant primary MAO gap (5, 0.7%)

Known WBH gap — companions of giant primaries require Plan 3+ stellar systems, which are not implemented.

| Count | First seed | Message                                                                        |
| ----: | ---------: | ------------------------------------------------------------------------------ |
|     4 |        292 | `stars: companion[0] orbit: companion of giant primary requires MAO (Plan 3+)` |
|     1 |       4141 | `stars: companion[1] orbit: companion of giant primary requires MAO (Plan 3+)` |

### 5. Likely real bug — investigate (1, 0.1%)

| Count | First seed | Message                                                                |
| ----: | ---------: | ---------------------------------------------------------------------- |
|     1 |       6724 | `placement: available-orbits: MAO for group A: no MAO row for "\x000"` |

The `"\x000"` is a NUL byte where a star descriptor should be — not a Special-Circumstances signature. Looks like uninitialized or stomped string state upstream of the MAO lookup. Deterministic reproducer: **seed 6724**.

## Categorization summary

| Category                                              | Count | % of errors |
| ----------------------------------------------------- | ----: | ----------: |
| Out-of-scope by project scope (Special Circumstances) |   653 |       96.6% |
| Book inter-table gap (Class IV/VI missing)            |    15 |        2.2% |
| Known unimplemented WBH gap (Plan 3+ companions)      |     5 |        0.7% |
| Suspected real bug (`\x000` descriptor)               |     1 |        0.1% |

## Reproducing

The bulk runner has been removed. To regenerate:

1. Create `cmd/wbh-bulk/main.go` invoking `worlds.Generate(seed)` in parallel for seeds 1..N.
2. Tabulate `err.Error()` strings with first-seed-seen for reproducibility.
3. Delete `cmd/wbh-bulk/` after.

Or for spot-checks: `go run ./cmd/wbh -seed 6724 -format short`.
