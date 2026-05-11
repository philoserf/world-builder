# Wider `AllBodies()` Migration

## What it is

`worlds.Universe` has an `AllBodies() iter.Seq[*Body]` method (`worlds/universe.go:40`) that yields every body in the universe — planets, then each planet's children (moons), in ascending-orbit order within each star group. It is the _contract_ iterator: `LongProfile` and `AssignPlanetDesignations` rely on its ordering.

But only stages 5, 8, and 9 actually use it. Stages **3, 4, 6, and 7** still hand-walk `u.Detail.Bodies` and `body.Children` with paired loops:

```go
for i := range u.Detail.Bodies {
    body := &u.Detail.Bodies[i]
    // ... per-planet work ...
    for j := range body.Children {
        child := body.Children[j]
        // ... per-moon work ...
    }
}
```

That same paired loop appears at **stage3.go:20-32, 69-83, 88-97; stage4.go:29-48, 51-69, 73-91, 94-112; stage6.go:21-30, 59-80; stage7.go:25-37** — eight separate sites, each individually responsible for remembering to descend into `Children`.

## Why it matters

Per `worlds/CLAUDE.md` and the project memory, **anti-pattern A.1 — moon-path-diverges-from-planet-path silent-zero** is the documented #1 critical-bug class in this codebase. Across four consecutive 3A2b/3B sub-projects, the Opus final-gate review caught critical bugs where a `runStep5*` addition added planet logic without iterating moons, producing silent-zero values for moons.

The fix is `Universe.AllBodies()` — one iterator, descends into `Children` automatically. Stages 5/8/9 already use it. But stages 3/4/6/7 are the older hand-rolled pattern. As long as those eight loops exist, every future stage-addition or refactor carries an unforced-error risk that the moon descent gets forgotten or diverges.

## Why it's not a 30-line fix

Three of the four stages need _parent context_, not just the body:

- **Stage 3** (`refineParentMoons`, `generateBodyPhysicalIfTerrestrial`) — moons inherit `hostForHZ` from the parent for HZ flag and orbit. The procedure body takes `body, hostForHZ *Body`.
- **Stage 4** (tidal lock, tidal effects, day length, axial tilt) — moons take the parent's stellar orbit; some procedures need both the moon and its parent planet.
- **Stage 6** (taint typology, surface distribution) — operates per-body without parent context, but is currently structured around index iteration.
- **Stage 7** (geology) — needs parent for some computations.

So `AllBodies()` alone isn't enough. To migrate stages 3, 4, 7 cleanly, the iterator API needs a parent-aware variant: `Universe.AllBodiesWithParent() iter.Seq2[*Body, *Body]` yielding `(body, parent)` where `parent` is the planet for moons and `nil` for top-level bodies.

## Migration plan

### Phase 1 — extend the iterator API (`worlds/universe.go`)

- Add `AllBodiesWithParent() iter.Seq2[*Body, *Body]` mirroring `AllBodies()` but yielding `(body, parent)`.
- Optional: add `(b *Body) Host() *Body` returning `b.Parent` for moons and `b` for planets/belts — eliminates the recurring `host := body; if body.Kind == BodyMoon && body.Parent != nil { host = body.Parent }` idiom that already appears at `stage5.go:42-45` and `stage5.go:148-151`.

### Phase 2 — promote `bodyMassEarth` to a method (`worlds/body.go`)

- Add `(b *Body) MassOrDerived() float64` that returns `MassEarth` if non-zero, else falls back to `DeriveMass(Physical.Density, DiameterKm)`.
- Today the same fallback is open-coded at `stage3.go:103-105`, `stage7.go:111-122`, and `tidal_lock.go:357` — three divergent copies of real domain knowledge (Stage-2 fills `MassEarth` for GG only; terrestrials need Stage-3 derivation), invited by the staging confusion.

### Phase 3 — migrate stage 6 (smallest first, parent-context-free)

- `ApplyTaintTypology` and `ApplySurfaceDistribution` walk `body.Children` for no parent reason. Replace with `for body := range u.AllBodies()`.
- Lowest risk; lands the pattern without surfacing any of the parent-handling subtleties.

### Phase 4 — migrate stages 3, 4, 7 (need parent)

- Convert each to `for body, parent := range u.AllBodiesWithParent()`.
- The current `generateBodyPhysicalIfTerrestrial(r, body, hostForHZ, sys)` becomes a call inside the loop where `hostForHZ` is `parent` when `parent != nil` else `body`.
- Stage 4's tidal-effects flow has the most parent-coupling — verify each procedure against the worked-example tests (Zed, Sol/Terra) line by line.

### Phase 5 — kill the open-coded child walk in stage 10

- `pickMainworld` (post-PR #38) still has a manual `for _, child := range body.Children` loop. Migrate to `AllBodies()` for full consistency.
- Same for `buildLongProfile` (`stage10.go:114`), which iterates `u.Detail.Bodies` directly. (LongProfile is ordering-sensitive, but `AllBodies()` _is_ the contract order, so this is safe.)

## Risk

- **Test surface:** worked-example tests (Zed: `TestZed_FullDetail_3A2b`; Sol/Terra: `TestSolTerra_p35`) assert every output to the digit. Any migration that subtly reorders dice consumption breaks these. The migration must preserve the exact iteration order — same planet-first-then-children traversal, same star-group grouping.
- **Pointer stability:** `u.Detail.Bodies` is `[]Body` (value slice). Pointers from `&u.Detail.Bodies[i]` remain valid only while the slice isn't reallocated. `AllBodies()` already takes `&u.Detail.Bodies[i]`, so this is no worse than the current code — but worth checking that no migrated stage appends to `u.Detail.Bodies` while iterating.
- **Moon `Children` is `[]*Body`:** moons are heap-allocated; their pointers are stable.

## Expected payoff

- Roughly 80 LoC of paired-loop boilerplate replaced by 4–5 single-loop calls.
- Anti-pattern A.1 closed _at the type-system level_ for everything inside the migrated stages — you can no longer forget to walk moons because there's no second `body.Children` loop to forget.
- One canonical iteration order; one canonical "host vs. body" predicate; one canonical mass-derivation rule. All three are currently scattered.

## Why this is deferred from `/simplify`

Single PR, ~150-200 lines net change across 6 files, requires careful per-procedure verification against the worked-example tests. It earns its own branch with focused review — exactly the wrong shape for a `/simplify` pass that should be safe, mechanical, and atomic.
