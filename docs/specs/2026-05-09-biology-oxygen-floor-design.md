# Biology Optional Rule 1 — Oxygen-Atm Biomass Floor — Design

**Date:** 2026-05-09
**Issue:** #12 (partial — Rule 1 of 2)
**WBH refs:** p.128 sidebar "Optional Rule" (oxygen-atmosphere biomass floor)

## Problem

WBH p.128 offers two optional Referee rules for biology generation. Issue #12 tracks both:

1. **Optional Rule 1 (oxygen-atm biomass floor):** any world with oxygen in the atmosphere (codes 2–9, D, E) has at least a biomass rating of 1.
2. **Optional Rule 2 (Rare Earth Universe Variant):** DM-2 on biocomplexity rolls, only positive biomass DMs, atm 4–9 with biocomplexity 1 → atm A + low oxygen taint.

This loop implements **Rule 1 only**. Rule 2 has multiple parts that mutate atmosphere mid-pipeline (a feedback path the current design does not support); it stays open in #12 (or moves to a successor issue) and is **out of scope here**.

## Reference — WBH p.128

> Optional Rule: The Referee may rule that any world with oxygen in the atmosphere (Atmosphere codes 2–9, D and E), has at least a biomass rating of 1. This may not be appropriate, as some non-biological processes may result in atmospheric oxygen…

The rule is opt-in by the Referee, so the implementation must default to **off**.

## Design

### Public API — `DetailOpts`

The library currently exposes `worlds.DetailSystem(r, sys, sp, h)` as the canonical entry point. Adding a 5th positional `opts` argument is a breaking change touching 7 callers; the standard Go pattern for backward-compatible opt-ins is to introduce a sibling `WithOpts` constructor:

```go
type DetailOpts struct {
    // OxygenAtmBiomassFloor enables WBH p.128 Optional Rule: any world
    // whose Atmosphere.Code is in the oxygen-bearing set {2-9, D, E}
    // gets a biomass floor of 1 (the rolled value is clamped up if it
    // came in below). Off by default — opt-in per book.
    OxygenAtmBiomassFloor bool
}

func DetailSystem(r, sys, sp, h) (SystemDetail, error) {
    return DetailSystemWithOpts(r, sys, sp, h, DetailOpts{})
}

func DetailSystemWithOpts(r, sys, sp, h, opts DetailOpts) (SystemDetail, error) {
    // existing body, plus opts plumbing
}
```

This preserves all existing callers — `DetailSystem(...)` keeps its signature and behavior unchanged.

### Internal plumbing

`DetailOpts` threads through:

```text
DetailSystemWithOpts
  → runDetailPipeline(r, detailed, sys, sp, opts)
      → runStep5F(r, detailed, sys, opts)
          → computeBiology(r, dp, ageGyr, opts)
```

`runStep5F` is the only existing pipeline step that uses opts in this loop, but `runDetailPipeline` accepts opts for forward-compat with Rule 2 and other future opt-ins.

### Floor application

Inside `computeBiology`:

```go
bio.Biomass = RollBiomass(r, dp, ageGyr)
if opts.OxygenAtmBiomassFloor && bio.Biomass < 1 && hasOxygenAtmosphere(dp.Atmosphere) {
    bio.Biomass = 1
}
if bio.Biomass > 0 {
    bio.Biocomplexity = RollBiocomplexity(r, dp, bio.Biomass, ageGyr)
    // ...
}
```

The floor runs **between the biomass roll and the dependent rolls**, so an elevated biomass of 1 propagates naturally into Biocomplexity / Biodiversity / Compatibility / Sophont rolls per the existing pipeline. No special-case wiring required.

### Oxygen-atmosphere predicate

```go
// hasOxygenAtmosphere reports whether atm carries free oxygen per
// WBH p.128 Optional Rule (codes 2-9, D, E). Hex codes:
//   2-9: oxygen-bearing thin → dense atmospheres
//   D (13): dense (oxygen-bearing in WBH's canonical mapping)
//   E (14): ellipsoidal (oxygen-bearing per WBH p.46)
func hasOxygenAtmosphere(atm *Atmosphere) bool {
    if atm == nil {
        return false
    }
    code := atm.Code
    return (code >= 2 && code <= 9) || code == 13 || code == 14
}
```

The predicate matches the book's literal list. Codes 0 (None), 1 (Trace), 10–12 (A/B/C exotic/corrosive/insidious), 15 (F+) are explicitly excluded.

### Interaction with existing Special Case 1

WBH p.131 "Special Case 1" already promotes biomass=0 → 1 with biocomplexity=1 in a different scenario (terrestrial life within an ecosystem). That code path stays unchanged — Optional Rule 1 is a separate, opt-in promotion that fires earlier (right after the biomass roll). When both fire, biomass is 1 from Rule 1 and biocomplexity is rolled normally; Special Case 1 does not re-trigger because biomass is already ≥ 1.

## Tests

New tests in `worlds/biology_test.go`:

1. `TestRollBiomass_OxygenAtmFloor_Off_RolledZeroStaysZero` — oxygen atm + low roll, opts off → biomass 0.
2. `TestRollBiomass_OxygenAtmFloor_On_RolledZeroBecomesOne` — same dice, opts on → biomass 1.
3. `TestRollBiomass_OxygenAtmFloor_On_NonOxygenAtmStaysZero` — atm A (10), opts on, low roll → biomass 0 (rule does not apply).
4. `TestRollBiomass_OxygenAtmFloor_On_RolledPositiveUnchanged` — oxygen atm, opts on, dice produce biomass 5 → biomass 5 (rule does not depress).

Plus an integration smoke test:

5. `TestDetailSystemWithOpts_OxygenAtmFloor` — drive `DetailSystemWithOpts` through Step 5F with opts on; verify a body that would have biomass 0 instead has biomass 1 and a populated biocomplexity (i.e., the dependent-roll branch ran).

The Zed worked-example regressions exercise `DetailSystem(...)` (no opts) — they must stay green.

## Acceptance Criteria

- `DetailOpts` struct exported with `OxygenAtmBiomassFloor bool`.
- `DetailSystemWithOpts` exported; `DetailSystem` is the no-opts wrapper.
- All 7 existing `DetailSystem` callers compile and behave identically.
- Optional Rule 1 fires only when opted in AND atm is oxygen-bearing AND biomass rolled below 1.
- All existing biology tests stay green; Zed golden unaffected.
- `task check && task test` clean.
- Issue #12 updated to track Rule 2 only (or closed with a successor issue filed for Rule 2).

## Out of Scope

- Optional Rule 2 (Rare Earth Universe Variant) — multi-part rule with atm-mutation feedback. Stays in #12 / future issue.
- CLI flag exposure — `cmd/wbh` continues to call `DetailSystem` (no opts). A future CLI flag can call `DetailSystemWithOpts`.
- Generalizing `DetailOpts` to other pipeline steps — only Step 5F needs an opt today; other steps will add their own opts when needed.
