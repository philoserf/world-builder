# C1 Sub-Roller — Implementation Plan (retrofit onto `main`)

Concrete plan to land the sub-roller tree (`docs/rebuild-spec.md` § C1) on current
`main` — the one high-value change from the rebuild spec that does **not** require a
rewrite. The spike on branch `c1-subroller` validated the mechanism; this plan
generalizes it across every per-body stage.

## Status: implemented on `c1-subroller`

All per-body suffix stages (`ApplyBodyPhysical` → `ApplyBiology`, plus the tidal-lock
re-eval cascade) fork per-body substreams. `roller.Fork` added with Seeded-branches /
Scripted-and-Fixed-transparent semantics. Suite green, `gofumpt`/`go vet`/modernizer
clean. Footprint: ~200 lines across `roller/roller.go`, the stage orchestrators, and a
new `worlds/subroller.go` helper; plus `roller` Fork unit tests and
`worlds/subroller_test.go` (whole-suffix isolation over 40 seeds × 3 shifts,
seed-dependence, fork-key stability canary). Markdown baseline regenerated once. What
remains before merge to `main`: review the baseline diff, decide branch naming, and land
the follow-on Zed gold fixture (below).

## Spike verdict (branch `c1-subroller`)

- **Fidelity is free.** `Scripted.Fork` is transparent (returns a view over the one flat
  dice list), so `TestApplyRotationTilt_TwoPass` — the orchestrator-level worked-example
  test asserting book values to the digit — passes unchanged. All `worlds`/`stars`
  worked-example and property tests stay green.
- **Isolation holds.** `TestSpikeC1_*`: after a 3-roll shift of the shared stream, a
  forked stage's per-body output is byte-identical (whole-universe `DeepEqual`) across 30
  seeds, while climate (still shared-stream) diverges — proving the perturbation is real
  and the isolation non-vacuous. Rotation-tilt is seed-dependent, so not degenerate.
- **Blast radius = the 5 Markdown snapshots only** (regenerated; all still render as
  complete systems). +78 lines across `roller/roller.go` and `worlds/stage4.go`.

## The enabling decision (why this is non-invasive)

`Seeded.Fork(key)` derives the child from the roller's **immutable construction seed**,
not its live rng state. Two consequences:

1. Forking never consumes from the parent and is independent of the parent's position.
   So a stage can fork off the very `r` it is already handed — `r.Fork(id).Fork(family)`
   — and get the same substream regardless of how much earlier stages consumed from `r`.
2. **No signatures change.** Every `Apply*` keeps `(roller.Roller, *Universe) error`. No
   fork-root field on `Universe`, no threading. Direct-invocation tests are unaffected.
   Full C1 is "repeat the stage-4 body edit in each remaining per-body orchestrator."

## Structure prefix vs. per-body suffix

Body identity (`Designation`) is assigned inside `ApplyDetailFrontEnd` (stage 2), and
moons are _created_ there too. So there is a clean boundary:

- **Structure prefix — stays shared/positional:** `stars.GenerateSystem`,
  `GenerateSystemPlacement`, `ApplyDetailFrontEnd`. These build the bodies and assign
  designations; there is no stable per-body identity to key on before they run, and none
  of the C1 pain (climate resampling, the tidal-lock cascade, reorder fragility) lives
  here. Leave them on the shared `r`. Document the boundary.
- **Per-body suffix — forks per body:** every stage from `ApplyBodyPhysical` onward runs
  after designations exist. Each keys `r.Fork(bodyID).Fork(family)` per body. This is
  where all the value is.

Designations are a deterministic function of placement (group + orbit), dice-independent,
so the key is stable the moment the suffix begins.

## Fork key and family taxonomy

- **Body id:** `bodyForkID(body, parent)` — `body.Designation` for top-level bodies,
  `parent.Designation + "/" + body.Designation` for moons (guarantees system-wide
  uniqueness). Already implemented in the spike (`worlds/stage4.go`). Promote it to a
  shared helper (e.g. `worlds/subroller.go`).
- **One family key per stage** (the substream a body draws that stage's dice from):

  | Stage / orchestrator       | family key                                    |
  | -------------------------- | --------------------------------------------- |
  | `ApplyBodyPhysical`        | `"body-physical"`                             |
  | `ApplyBeltDetails`         | `"belt"`                                      |
  | `ApplyMoonRefinement`      | `"moon-refine"`                               |
  | `ApplyRotationTilt`        | `"rotation-tilt"` ✓                           |
  | `ApplyClimate`             | `"climate"`                                   |
  | `ApplyTidalLockReEval`     | `"rotation-tilt-reeval"` + `"climate-reeval"` |
  | `ApplyTaintTypology`       | `"taint"`                                     |
  | `ApplySurfaceDistribution` | `"surface"`                                   |
  | `ApplyGeology`             | `"geology"`                                   |
  | `ApplyBiology`             | `"biology"`                                   |

- **Multi-substage stages** (e.g. stage 4's day-length / axial-tilt / tidal-lock passes)
  **fork once per body and reuse** the handle across sub-stages, via the
  `sub := map[*Body]roller.Roller` pattern the spike established — never re-fork the same
  `(id, family)` mid-stage or the sub-stages collide on one seed.

## The tidal-lock cascade (the payoff)

`ApplyTidalLockReEval` re-runs tidal lock + climate for bodies with pressure > 2.5 bar.
Give the re-run its **own** family keys (`"rotation-tilt-reeval"`, `"climate-reeval"`) so
it draws fresh, independent dice for the affected body and touches **no other body's
substream**. This retires the "~2× dice, perturbs the whole downstream stream" trade-off
noted in `wbh-inconsistencies.md` § 7: the cascade becomes a local, isolated re-draw.
(The snapshot/restore of pre-tidal state stays — that is domain logic, not a stream hack.
Renaming it is a C2 concern, out of scope for C1.)

## Stage-by-stage checklist

For each suffix orchestrator, apply the spike edit:

1. At the top, build `sub := map[*Body]roller.Roller` over `u.AllBodiesWithParent()`
   (skip `BodyEmpty`): `sub[body] = r.Fork(bodyForkID(body, parent)).Fork(family)`.
2. Replace every per-body `Generate*/Roll*(r, body, …)` call with `sub[body]`.
3. Leave non-dice (`Compute*/Derive*`, roller-less `Generate*`) calls alone.

Order (each independently testable): `body_physical` → `belt_details` →
`moon_refinement` → `rotation_tilt` (done) → `climate` → `tidal_lock_reeval` →
`taint` → `surface_distribution` → `geology` → `biology`.

## Test strategy

- **Fidelity:** existing worked-example + property tests must stay green throughout
  (Scripted transparency guarantees it). Do not touch them.
- **Isolation:** generalize the spike test into one property test that runs the **full**
  pipeline twice with a k-roll shift injected right after `ApplyDetailFrontEnd`, and
  asserts the **entire** post-suffix universe is `DeepEqual` across the shift. Once every
  suffix stage is forked, the whole suffix is position-independent — so this single
  assertion covers all stages and replaces per-stage isolation tests.
- **Cascade:** a targeted test on a pressure > 2.5 bar seed (e.g. `seed_500`'s `A II`)
  asserting that forcing the re-eval leaves every _other_ body byte-identical.
- **Baseline:** regenerate `iiss/testdata/seed_*.md` **once, at the end** of the sweep
  (`go test ./iiss/... -update.regression -run TestRegression`), after review. Because a
  forked stage stops consuming from the shared stream, the baseline drifts once per stage
  if done incrementally — so land the suffix conversion as a single reviewed change (or a
  tight series) and regenerate the baseline last.

## Risks and decisions

- **Single sweep, not incremental releases.** Each converted stage shifts the shared
  stream for any still-unconverted downstream stage, so mid-sweep the baseline is in
  flux. Convert the whole suffix on one branch, regenerate the baseline once, then merge.
- **Cascade re-roll semantics (decided):** fresh isolated dice via `*-reeval` sub-keys,
  not a replay of the first-pass substream. Replaying would re-interpret the same roll
  under the corrected DM; fresh keys are cleaner and match the book's "re-roll" intent.
- **Pre-designation isolation (deferred):** the structure prefix stays shared. Isolating
  stars/placement/detail-frontend would need a placement-slot key and buys little; leave
  it, document it.
- **Fork-key stability is contract.** Never reorder or rename family keys after merge —
  doing so re-seeds every substream and silently changes all output. Pin them with a test
  that asserts a known seed's mainworld SAH is stable.
- **Full-pipeline gold fixture (follow-on, not C1):** once the suffix is position-
  independent, a survivable end-to-end Zed Prime gold fixture becomes possible. Add it as
  a separate change after C1 lands — it is the durable dividend C1 unlocks.

## Sequencing

1. Branch `c1-subroller` off `main`. Port `roller.Fork` + the `bodyForkID` helper.
2. Convert the suffix orchestrators in dependency order; keep `task test` green at each
   step except the (expected) Markdown regression, which stays red until step 5.
3. Wire the cascade's `*-reeval` keys.
4. Add the full-pipeline isolation property test and the cascade-locality test.
5. Regenerate the Markdown baseline once; review the diff (it should be pure stream drift,
   no missing/`?` fields); confirm `task check && task test` green.
6. Follow-on change: survivable Zed Prime full-pipeline gold fixture.
