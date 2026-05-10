# Runaway Greenhouse "Consider Boiling" for Atm A/B/C/F+ — Design

**Date:** 2026-05-09
**Sub-project:** 3A2b-rederive follow-up
**Predecessors:** 3A2b-rederive (`main`), atm-bc-pressure-from-subtype (PR #25, merged 2026-05-09)
**Closes:** issue #8

## Goal

Extend `CheckRunawayGreenhouse` to evaluate atmospheres of code A (10), B (11), C (12), and F+ (15, 16, 17). The current MVP skips these codes entirely; per WBH p.79, they should still trigger when the 2D+DMs roll lands on 12+, but with a different outcome: "the only effect of a runaway greenhouse is to consider the world to be boiling." The boiling consideration applies the existing `TempBoiling` DM-6 to the hydrographics re-roll instead of the original `TempHot` DM-2.

## Source of truth

WBH p.79, "Optional Rule: Runaway Greenhouse":

> An additional consideration is the eventual evaporation rate of the oceans of hotter worlds and the accelerating effect of this process caused by further rising temperatures…To simulate this, the Referee can examine any world within the habitable zone that has an Atmosphere code of 2–F and is boiling (adjusted temperature roll of 12+) or hot (10 or 11) as a result of basic generation…the Referee can determine if a runaway greenhouse occurred by rolling 2D:
>
> **Runaway Greenhouse occurred on 12+:** roll 2D + DMs

WBH p.79, outcome for already-extreme atmospheres:

> If the world already has an Atmosphere code of A, B, C or F+, then the only effect of a runaway greenhouse is to consider the world to be boiling if it was only considered hot. This can reduce the hydrographics roll by DM-6 instead of DM-2. For all other worlds, namely those with atmosphere codes 2–9, D, or E, a runaway greenhouse converts their atmosphere code to A, B or C, based on a 1D roll.

The "consider boiling" outcome is what's currently missing — the existing implementation skips the trigger evaluation entirely for atm A/B/C/F+ instead of running it and applying the boiling-only effect.

## Decisions

### API shape: keep `bool` return; caller compares pre/post atm code

`CheckRunawayGreenhouse` keeps its current signature: `func(r, body, sys) bool`. On a successful trigger:

- **Atm 2-9, D, E (existing path):** mutate `body.Atmosphere.Code` via 1D table to A/B/C. Caller detects mutation by comparing pre/post code and re-rolls subtype/pressure via `rerollAtmSubtypeAndPressure`.
- **Atm A, B, C, F+ (new path):** do not mutate; just return `true`.

In both cases, the caller forces `hydroTempRange = TempBoiling` to apply the DM-6 hydrographics modifier.

Alternative considered: encapsulate the `rerollAtmSubtypeAndPressure` call inside `CheckRunawayGreenhouse`. Rejected because the helper returns `error`, and broadening the signature to `(bool, error)` ripples through tests and callers more than the caller-side pre/post-code check, which is a one-line addition.

### Eligibility for atm 15+ (F/G/H)

Strictly follow the book's "F+": atm 15 (Unusual), 16 (Gas Helium), 17 (Gas Hydrogen) are all eligible. The book is unambiguous; physical plausibility on a hydrogen gas dwarf is questionable but not our call to make. Implementation cost is identical to F-only (one comparison).

### No subtype/pressure re-roll for already-extreme atms

WBH p.79: "**the only effect** of a runaway greenhouse is to consider the world to be boiling." We trust the "only effect" language and don't re-roll subtype or pressure for atms that didn't get their code mutated.

A nearby paragraph says "All worlds suffering a runaway greenhouse receive a DM+4 on Atmosphere code subtypes determination rolls." Reading this in context, the DM+4 applies to the _new_ subtype roll produced by the conversion table for atm 2-9/D/E paths — not to atms that already had a subtype set in 3A1.

## Architecture

### `worlds/runaway_greenhouse.go` — extend eligibility, branch on outcome

Replace the current eligibility filter:

```go
// Trigger range: atm 2-9, D (13), E (14). Skip A/B/C (10-12), F+ (15+), and 0/1.
if code < 2 || code == 10 || code == 11 || code == 12 || code >= 15 {
    return false
}
```

with:

```go
// Trigger range: atm 2 and above. Atm 0/1 (no atmosphere or trace) are skipped.
if code < 2 {
    return false
}
```

After the 2D+DMs trigger check (unchanged), branch on atm code:

```go
roll := r.Roll("2D")
if roll+dm < 12 {
    return false
}

// Trigger fired. WBH p.79: for atm A/B/C/F+, the only effect is the
// "consider boiling" hydrographics DM (handled by the caller). For
// atm 2-9, D, E: mutate the atm code via 1D roll.
if code == 10 || code == 11 || code == 12 || code >= 15 {
    return true
}

// Atm 2-9, D, E: 1D table → A/B/C.
atmRoll := r.Roll("1D")
switch {
case atmRoll == 1:
    body.Atmosphere.Code = 10 // A
case atmRoll <= 4:
    body.Atmosphere.Code = 11 // B
default:
    body.Atmosphere.Code = 12 // C
}
return true
```

The function's doc-comment loses its MVP-simplification paragraph and gains a description of the dual outcome.

### `worlds/temperature_rederive.go` — caller detects code mutation

Replace the current call-site block:

```go
runawayFired := false
if body.HZ {
    runawayFired = CheckRunawayGreenhouse(r, body, sys)
    if runawayFired {
        // Re-roll subtype + pressure with runawayResult=true (DM+4 to subtype).
        if err := rerollAtmSubtypeAndPressure(r, body, sys, true); err != nil {
            return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: post-runaway: %w", err)
        }
    }
}
```

with:

```go
runawayFired := false
if body.HZ {
    var preCode int
    if body.Atmosphere != nil {
        preCode = body.Atmosphere.Code
    }
    runawayFired = CheckRunawayGreenhouse(r, body, sys)
    if runawayFired && body.Atmosphere != nil && body.Atmosphere.Code != preCode {
        // Atm code was mutated (atm 2-9/D/E → A/B/C path). Re-roll
        // subtype + pressure with runawayResult=true (DM+4 to subtype).
        // Atm A/B/C/F+ "boiling-only" path leaves the code unchanged
        // and falls through to the hydro DM-6 below without re-rolling
        // atmospheric structure.
        if err := rerollAtmSubtypeAndPressure(r, body, sys, true); err != nil {
            return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: post-runaway: %w", err)
        }
    }
}
```

The hydro `TempBoiling` override below is unchanged — it fires on `runawayFired` regardless of mutation, which is correct for both paths.

## Out of scope

- **Atm 0/1 trigger.** WBH says atm 2-F; atms 0 (None) and 1 (Trace) are not in the runaway-greenhouse table. No change.
- **Optional rule for non-HZ worlds.** WBH p.79: "If desired, this roll can also be performed on any world closer than the HZCO, with DM-2 for a world with Temperate conditions." This optional extension is not in scope; the existing `if body.HZ` gate is preserved.
- **Optional 303K post-temp re-trigger DM+1 per 10°.** WBH p.111: "for any world where the mean temperature exceeds 303K (30°C), use DM+1 for every full 10° above 303K instead of the boiling temperature DM." This is a different DM table for a different trigger pathway and is not in scope; the current MVP uses the simpler "MeanK ≥ 388 → DM+4" rule.

## Testing strategy

### Unit tests for the new path

Add to `worlds/temperature_rederive_test.go` (where existing `TestCheckRunawayGreenhouse_*` live):

- `TestCheckRunawayGreenhouse_AtmA_BoilingOnly`: atm code 10, HZ, MeanK 400, age 5 Gyr, size 8. Expected: `true`, atm.Code unchanged at 10.
- `TestCheckRunawayGreenhouse_AtmB_BoilingOnly`: atm code 11, same conditions. Expected: `true`, atm.Code unchanged at 11.
- `TestCheckRunawayGreenhouse_AtmC_BoilingOnly`: atm code 12, same conditions. Expected: `true`, atm.Code unchanged at 12.
- `TestCheckRunawayGreenhouse_AtmF_BoilingOnly`: atm code 15, same conditions. Expected: `true`, atm.Code unchanged at 15.
- `TestCheckRunawayGreenhouse_AtmH_BoilingOnly`: atm code 17 (Gas Hydrogen), same conditions. Expected: `true`, atm.Code unchanged at 17.

### Update existing test

`TestCheckRunawayGreenhouse_AtmAlreadyExotic_Skipped` currently asserts that atm A/B/C return `false`. Replace with `TestCheckRunawayGreenhouse_AtmAlreadyExotic_BoilingOnly` that asserts the new behavior: trigger evaluates, fires when conditions met, code unchanged.

### Integration test

`TestRederiveAtmosphereHydrographics_RunawayBoilingOnly_AtmB`: HZ atm-B body with high mean temp; verify that after rederive, atm.Code stays at 11 but hydro was re-rolled with TempBoiling (i.e., the resulting hydro code reflects the DM-6 — testable by scripting RNG and comparing the hydro outcome with vs without the boiling override).

### Zed golden

May refresh. Any seed=42 HZ atm-B body whose MeanK exceeds 303K and whose 2D+DMs lands ≥12 will now trigger the boiling-only path, possibly shifting hydro. The current Zed golden has Aab IV at MeanK 313 (atm B, HZ); with age dependent DMs this may or may not trigger. Verify after implementation; refresh if affected.

## Closes

#8.
