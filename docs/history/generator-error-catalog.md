# Generator error catalog — 10 000-seed sweep

**Original sweep:** 2026-05-11 (pre-cleanup)
**Latest sweep:** 2026-05-12 (post Phase 2 complete)
**Tool:** one-off `cmd/wbh-bulk/main.go` (parallel goroutines, ~1.5 s wall on 8 cores; removed after each run)
**Range:** seeds 1..10 000 inclusive
**Entry point:** `worlds.Generate(seed)` — same path as `cmd/wbh`

## Final state (post Phase 2 complete) 🎯

| Bucket                  |     Count |  % of total |
| ----------------------- | --------: | ----------: |
| **Successes**           | **10000** | **100.00%** |
| Errors total            |         0 |       0.00% |
| Distinct error messages |         0 |           — |

Every invocation of `cmd/wbh -seed N` for N in 1..10000 produces a real, fully-rendered system. The intent stated in `plan-clean-every-run.md` is achieved.

## Journey

| Sweep                               | Successes | Errors | Distinct |
| ----------------------------------- | --------: | -----: | -------: |
| Initial (2026-05-11)                |      9324 |    676 |       20 |
| Phase 0a (1 real bug fixed)         |      9324 |    676 |       19 |
| Phase 0b (WBH p.16 constraints)     |      9337 |    663 |       14 |
| Phase 1 (typed errors)              |      9337 |    663 |       14 |
| Phase 2a (Special primary column)   |      9522 |    478 |       11 |
| Phase 2c (Giants dispatch)          |      9534 |    466 |       10 |
| Phase 2d (companion-of-giant orbit) |      9562 |    438 |        7 |
| Phase 2f (post-stellar group MAO=0) |      9949 |     51 |        1 |
| Phase 2b (companion Special col)    | **10000** |  **0** |    **0** |

## Phase-by-phase summary

### Phase 0 — kill non-scope errors (16 of 676)

- **0a (1d4bf6c)** — protostar primary `\x000` MAO lookup. `lacksP39MAORow` predicate at the MAO gate.
- **0b (c5aabb2)** — WBH p.16 Class IV/VI letter constraints applied across 4 roll paths + one-sided-missing graceful interpolation in `InterpolateClassRow`.

### Phase 1 — type the errors (90b64d2, closed #45)

`stars.ErrSpecialCircumstances` umbrella sentinel; every Special-Circumstances error wraps it via `%w`; test classifiers migrated from substring matching to `errors.Is`.

### Phase 2 — adopt the book's Referee defaults

- **2a (85d80a3)** — Primary dispatch defaults to WBH p.15's Special column. Adjacent fix: class redirects use DM+1 on Type-column re-roll (WBH p.16).
- **2c (f5ae866)** — Special-column Giants cell (rows 11-12) dispatches through `RollGiantClass` + `generatePrimaryAtClass`.
- **2d (2d6008b)** — Companion-of-giant orbit = `1D × MAO(primary)` per WBH p.27. `MAO` callback threaded through `GenerateSystemOpts`.
- **2f (25751bb)** — Post-stellar group members contribute 0 to MAO. WBH-allowed BD/D companions per p.29 exist in the system without invoking the Special Circumstances chapter for their orbital footprint.
- **2b (12d714d)** — Companion `Random` descriptor's Type-column "Special" cell dispatches through the Special column (mirrors 2a for companions).

## Permanent regression fixtures

Unit-level tests live next to each fix and cover the failure modes:

- `worlds/available_orbits_test.go::TestMAO_Protostar`
- `stars/peculiar_test.go::TestGeneratePrimaryAtClass_IV_M_to_K`
- `stars/peculiar_test.go::TestGeneratePrimaryAtClass_VI_F_to_G`
- `stars/peculiar_test.go::TestRollSpecialPrimary_Special_ClassRedirects`
- `stars/system_test.go::TestGenerateSystem_SpecialPrimary_GiantsCell`

## Reproducing the sweep

`cmd/wbh-bulk/` is a one-off removed after each run. To regenerate:

1. Create `cmd/wbh-bulk/main.go` invoking `worlds.Generate(seed)` in parallel for seeds 1..N, classifying errors via `errors.Is(err, stars.ErrSpecialCircumstances)`.
2. Tabulate counts + first-seed-seen per distinct `err.Error()`.
3. Delete `cmd/wbh-bulk/` after.

Spot-check one seed: `go run ./cmd/wbh -seed 67 -format short` (previously errored; now produces output).
