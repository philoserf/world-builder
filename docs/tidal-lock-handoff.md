# Tidal-Lock Issue Handoff

**Status:** five open issues against `worlds/tidal_lock.go`, all spec divergences from WBH pp.105–107 surfaced during the post-#52 cross-check. PR #52 (`fix(worlds): exclude moons from planet→star tidal-lock case`) fixed the largest divergence; what remains is smaller in scope but still observable.

## Open issues

| #   | Summary                                                                      | Direction on lock count |
| --- | ---------------------------------------------------------------------------- | ----------------------- |
| #9  | Atmosphere DM (`P > 2.5 bar → -2`) at `tidal_lock.go:112` is dead code       | fewer (if revived)      |
| #53 | Tied-DM cascade missing (p.106 says roll both cases, take highest)           | more                    |
| #54 | Moon-period guard missing on the natural-12 verification reroll (p.105 †)    | more                    |
| #55 | Planet→Moon eligibility too loose — should require terrestrial + locked moon | fewer                   |
| #56 | Planet→Moon 1:1 sidereal day uses planet's stellar year, should be moon's PD | none (day length only)  |

Cumulative directional guess: roughly 5–15% fewer locks overall, dominated by #9 if it lands as a real fix. Confidence: moderate.

## Recommended sequence

1. **#9 — decide first.** The cheap path (remove dead code + document divergence in `wbh-inconsistencies.md`) has zero determinism impact because the DM never fires today. The faithful path (capture raw dice, re-evaluate after Stage 5) is its own multi-PR effort. The decision gates everything downstream — if revived, the Stage-5 re-eval cascade will overwrite snapshot updates from #53/#54.
2. **#55 + #56 paired.** Eligibility fix (#55) reduces Planet→Moon evaluation; sidereal-day fix (#56) only becomes observable once #55 is in. Pairing avoids a snapshot churn round-trip.
3. **#54 — narrow scope.** Single conditional branch in `ApplyTidalLockEffect`. Determinism break is contained because the triple-gate (≥12 initial, natural-12 verification, day-length > orbital period) only fires for a small slice of moons.
4. **#53 — largest blast radius.** Tied-DM cascade rewrites `SelectHighestDMCase` and the orchestration in `GenerateTidalLock`. Many seeds will see new locks. Save for last.

## Per-issue notes

### #9 — atmosphere DM dead code

**Code:** `worlds/tidal_lock.go:112` inside `commonTidalLockDMs`. `body.Atmosphere` is nil at Stage 4 because atmosphere is set in `ApplyClimate` (Stage 5).

**Cheap option:** delete the check, add an entry to `docs/wbh-inconsistencies.md` recording that the WBH common DM is intentionally not applied due to the pipeline ordering of Stage 4 (rotation/tilt/tidal) vs Stage 5 (atmosphere). Note that this is an honest divergence from p.106.

**Faithful option:** capture raw 2D + verification dice into `TidalLock`; after Stage 5 atmosphere is set, re-evaluate the DM for bodies with pressure > 2.5 bar and recompute the lock result. Has to undo any prior `ApplyTidalLockEffect` mutations to `DayLength`/`AxialTilt`/`Eccentricity`. The closed `enhancement: 3A2b — tidal-lock re-eval` comment thread (issue #9 itself) has the original cost analysis.

**Recommendation:** cheap option. The faithful path doesn't justify itself for a low-frequency, low-magnitude divergence on a project that's otherwise complete.

### #55 — Planet→Moon eligibility

**Code:** `worlds/tidal_lock.go:73` in `EvaluateTidalLockDMs`. Today's gate:

```go
if parentPlanet == nil && moonRef == nil && hasSignificantMoon(body) {
    out[TidalLockCasePlanetToMoon] = common + planetToMoonDMs(body)
}
```

**Spec (WBH p.107):** check only for terrestrial worlds (Size 1–F), and only if at least one moon is already locked to the planet.

**Implementation sketch:**

- Add a `isTerrestrial(body)` predicate (Size 1–F means `BodyKind == BodyTerrestrial` — gas giants and belts excluded).
- Sequence the Stage-4 tidal pass so MoonToPlanet runs for all moons before Planet→Moon is evaluated for any planet. Today `ApplyRotationTilt` walks bodies once and processes all cases per body; the pre-locked-moon check needs all moons of a planet to be resolved first.
- A two-pass walk in `ApplyRotationTilt` (moons first, then planets) is cleanest.

**Gotcha:** moons currently roll for `PlanetToStar` too in `EvaluateTidalLockDMs` — that was fixed in #52. But the case-selection ordering inside `SelectHighestDMCase` already prioritizes MoonToPlanet, so the two-pass scheme doesn't change moon outcomes, only the order of planet evaluation.

### #56 — Planet→Moon sidereal day source

**Code:** `worlds/stage4.go:62-77` plus `ApplyTidalLockEffect`'s `LockRatio == "1:1"` branch.

Currently `hours = body.Period.Hours` (planet's stellar year) when `body` is a planet. For Planet→Moon, should be the relevant moon's `PeriodHours`.

**Implementation sketch:** when `kase == TidalLockCasePlanetToMoon`, resolve the relevant moon (`planetToMoonDMs` already picks "closest significant moon by `OrbitPD`" — replicate that selection at the orchestration site) and pass its `PeriodHours` into `ApplyTidalLockEffect`.

**Gotcha:** `yearHours` is a single scalar passed in; the case is not known until after `SelectHighestDMCase` runs. Either resolve the moon at the call site after the case is known, or refactor `GenerateTidalLock` to compute `yearHours` internally based on the resolved case.

### #54 — moon-period guard on verification reroll

**Code:** `worlds/tidal_lock.go:376` in `ApplyTidalLockEffect`. Today's logic always uses the rerolled `FinalResult` when verification == natural 12.

**Spec (WBH p.105 †):** for MoonToPlanet only, if the rerolled day length exceeds the moon's orbital period (`PeriodHours`), the moon stays 1:1-locked.

**Implementation sketch:**

```go
if initialResult >= 12 {
    verification := r.Roll("2D")
    if verification == 12 {
        tl.VerificationFired = true
        rerolled := RollTidalLockStatus(r, 0)
        // Moon-period guard: if rerolled day length would exceed
        // the moon's orbital period, the 1:1 lock holds.
        if kase == TidalLockCaseMoonToPlanet && rerolledDayLength(rerolled, body, yearHours) > yearHours {
            // keep FinalResult = initialResult (1:1)
        } else {
            tl.FinalResult = rerolled
        }
    }
}
```

`rerolledDayLength` is a small helper that maps the rerolled result to the day length it would produce (multiplier × current, or new sidereal hours for results 7–10). Since `yearHours` for a moon is its `PeriodHours`, the comparison is straightforward.

**Gotcha:** the rerolled-result branch in `ApplyTidalLockEffect` consumes additional rolls for results 7–10 (`1D × N × 24`). If we compute the would-be day length without consuming those rolls, then take the "keep 1:1" branch, the dice stream is shorter than if we'd used the rerolled result. This is correct behavior — the book says the lock holds, no further effect — but is a subtle dice-stream change for seeds that previously hit the verification path.

### #53 — tied-DM cascade

**Code:** `worlds/tidal_lock.go:260` `SelectHighestDMCase` and the call in `GenerateTidalLock`.

**Spec (WBH p.106):** on a tie, roll the moon case first; if it doesn't lock, roll the next case; apply the highest adjusted roll.

**Implementation sketch:** `SelectHighestDMCase` needs to return an ordered list of tied cases, not a single case. `GenerateTidalLock` rolls them in order, stops once a lock is achieved (≥11), and applies the highest adjusted roll.

**Gotcha:** "if a lock condition does not occur, roll for the next case" reads as: only continue if the first roll didn't lock. But "apply the effect of highest adjusted roll" applies regardless. The simplest reading: roll all tied cases, take the highest adjusted result, apply that. The "moon first" ordering matters for which roller calls fire first (affects determinism for the verification reroll), not for the outcome.

**Determinism break:** large. Most affected seeds will see new tied-case rolls consuming dice, shifting downstream output.

## Cross-cutting concerns

### Seed determinism

Every issue except #9-cheap and #56 changes dice consumption. Expect:

- `iiss/testdata/seed_{1,7,42}.md` regression snapshots to drift on each landing.
- Zed worked-example tests (`*_p<page>` and `Test*_FullDetail_*`) are scripted with explicit dice values via `roller.NewScripted` — they don't shift unless the scripted call sequence changes. Verify by running `go test ./worlds/ -run TestZed`.

Each PR should regenerate snapshots via `go test ./iiss/... -update.regression -run TestRegression` and inspect the diff before committing.

### Pipeline ordering

#55 requires a two-pass walk in `ApplyRotationTilt`. The other issues can stay within the single-pass model. If pairing #55 + #56, do the two-pass refactor once and use it for both.

### Test additions

Each PR should add:

- A unit test in `worlds/tidal_lock_test.go` asserting the specific behavior (mirror the production call site — `body.Kind = BodyMoon` with `moonRef = body` for moon cases, per the #52 review feedback).
- For #55, a regression test verifying that gas giants and moons-without-locked-children do not get the Planet→Moon case in their DM map.
- For #53, a multi-case fixture verifying that tied DMs produce rolls for each tied case and that the highest adjusted roll is applied.

### Verification

Per-PR checklist:

```sh
task check                          # gofumpt + vet + golangci-lint
go test -race ./...                 # full suite
go test ./worlds/ -run TestZed      # worked-example fidelity
go test ./iiss/... -update.regression -run TestRegression   # snapshot refresh, inspect diff
```

Then a sanity sweep across the same 10 seeds we used for #52 verification:

```sh
for seed in 1 2 3 4 5 42 100 200 500 1000; do
  go run ./cmd/world-builder -seed $seed -format markdown | grep -c "planet → \|moon → "
done
```

Compare against the pre-fix counts to confirm directional movement matches the table at the top of this doc.

## Decision points

Before starting any PR:

- **#9:** cheap (remove + document) vs faithful (re-eval cascade). Default recommendation is cheap.
- **#55 + #56 pairing:** confirm the two-pass refactor in `ApplyRotationTilt` lands in one PR rather than split.
- **#53 cascade reading:** "stop at first lock" vs "roll all tied, take highest". Recommend the second — closer to literal spec reading and simpler to implement.

## Out of scope

- The book inconsistency at p.105 † (`(2D-2)/10` reroll) vs p.107 effect #2 ("1D on Axial Tilt table") — implementation follows p.105 †. Already noted in MEMORY; no issue filed.
- Special Circumstances chapter (WBH pp.147+) — explicitly out of project scope per `CLAUDE.md`.
