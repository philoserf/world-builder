# Atmosphere Taint Typology — Design

**Date:** 2026-05-08
**Sub-project:** atmosphere-taint-typology (post-3B follow-up)
**Predecessors:** 3A2b-rederive (`main`), 3B-biology (`main`), 3B-final (`main`)
**Closes:** issues #11 (biologic-taint Biomass promotion), #13 (low-oxygen-taint Biocomplexity DM-2)

## Goal

Implement the WBH atmosphere taint/irritant typology that the project deferred under spec tag `Q3-a` across 3B-biology, 3B-final, and 3A2b-rederive. Specifically:

- WBH p.81 — pre-existing oxygen taint promotion (atm 5/6/8 → 4/7/9 when ppO2 outside 0.10-0.50 bar; first taint subtype seeded as L or H)
- WBH p.82 — Taint Subtype table (2D + DMs; subtypes L/R/B/G/P/S/H; multi-taint reroll on result 10; max 3)
- WBH p.83 — L/H suppression on non-4-9 codes and on 2nd/3rd rolls (treat as G); ppO2/total-pressure adjustment when L or H rolled
- WBH p.84 — Taint Severity + Persistence tables (T.S.P profile)
- WBH p.85 — Exotic (A) atmosphere irritants (same Taint Subtype table)
- WBH p.89 — Corrosive (B) and Insidious (C) atmosphere irritants (same table)
- WBH p.90 — Insidious Atmosphere Hazard table (B/R/G/T; reroll on hazard subtype D/E)

After this sub-project, the three deferred biology DMs land cleanly:

- `RollBiomass`: biologic-taint (Code "B") with biomass=0 → biomass=1 (WBH p.127 Special Case 1)
- `RollBiocomplexity`: low-oxygen-taint (Code "L") → DM-2 (WBH p.129)
- `RollCompatibility`: any taint present → "or otherwise tainted" -2 qualifier (WBH p.131)

## Brainstorm decisions

| Q                                 | Decision                                                                                                |
| --------------------------------- | ------------------------------------------------------------------------------------------------------- |
| Q1: Scope                         | (iii) Subtype + Severity + Persistence + irritants on A/B/C + Insidious Hazard                          |
| Q2: Pipeline placement            | (B) New step 5D½ between rederive (5D) and geology (5E)                                                 |
| Q3: Pre-existing oxygen promotion | (A) Implement 5/6/8 → 4/7/9 promotion based on ppO2 in band                                             |
| Q4: Storage shape                 | `Atmosphere.Taints []Taint` + `Atmosphere.InsidiousHazard *Hazard`; severity/persistence always rolled  |
| Q5: Rendering                     | Extend `AtmosphereProfile.Shorthand` with taint/irritant block; existing render paths unchanged         |
| Q6: Hazard structure              | `Hazard{Code string}` only; no severity/persistence (book: hazards "automatically lethal and constant") |

## Architecture

### New file: `worlds/atmosphere_taint.go`

Holds the public types `Taint`, `Hazard`, the rolling functions, and the helper predicates biology consumes:

- `RollTaintSubtype(r, atmCode, isSecondOrLater bool) string` — rolls 2D + DMs (atm 4 → -2, atm 9 → +2; non-4-9 atms have no DM beyond suppression rule); applies L/H → G suppression for codes outside 4-9 and for 2nd/3rd rolls; returns code letter
- `RollTaintSeverity(r, taintCode, atmCode int, ppO2 float64) int` — 2D + DMs (DM+4 for L/H per p.84 footnote, with ppO2-specific severity overrides; DM+6 for atm C); returns 1-9
- `RollTaintPersistence(r, taintCode, atmCode int, severity int) int` — 2D + DMs (DM+4 for L/H, DM+6 if severity ≥ 8 or atm C); returns 2-9
- `RollAllTaints(r, body, atmCode int, ppO2 float64) []Taint` — orchestrator: handles pre-seeded L/H from p.81 promotion, multi-roll loop with result-10 reroll, max 3, calls severity + persistence per taint
- `RollInsidiousHazard(r) []string` — 2D table p.90; B/R/G/T; reroll once on T if subtype is D or E; returns hazard codes (1-2 entries)
- `HasTaintCode(taints []Taint, code string) bool` — predicate for biology consumers

### New file: `worlds/atmosphere_promotion.go`

Holds the WBH p.81 promotion logic in isolation:

- `PromoteOxygenTaint(atmCode int, ppO2 float64) (newCode int, seededTaint *Taint)` — pure function: returns promoted code and pre-seeded L/H taint when atm 5/6/8 with ppO2 outside [0.10, 0.50]; otherwise returns input code and nil

### New step: `runStep5DHalf(r, detailed, sys) error` in `worlds/system_detail_steps.go`

Slots between Step 5D (3A2b-rederive) and Step 5E (3B-geology). Single pass through `detailed`; per-body and per-moon (the moon-iteration anti-pattern check from MEMORY explicitly applies). Mutates `dp.Atmosphere.Taints` and (for atm C) `dp.Atmosphere.InsidiousHazard` in place. May also mutate `dp.Atmosphere.Code` (5/6/8 → 4/7/9 promotion) and `dp.Atmosphere.OxygenPartialPressure` / `Pressure` (L/H ppO2 adjustment per p.83).

`DetailSystem` orchestrator picks up one new line:

```go
// Step 5D½ — atmosphere taint typology: Taints + Severity + Persistence + Insidious Hazard.
if err := runStep5DHalf(r, detailed, sys); err != nil {
    return SystemDetail{}, err
}
```

### Struct extensions

`Atmosphere` (in `worlds/atmosphere.go`) gains two fields:

```go
type Atmosphere struct {
    // ... existing fields ...
    Taints          []Taint  // 0-3 entries; populated by Step 5D½
    InsidiousHazard *Hazard  // populated by Step 5D½ for atm C only
}
```

New types:

```go
// Taint — one taint or irritant condition per WBH p.82-84.
//
// Code values per WBH p.82 Taint Subtype table:
//   L = Low Oxygen
//   R = Radioactivity
//   B = Biologic
//   G = Gas Mix
//   P = Particulates
//   S = Sulphur Compounds
//   H = High Oxygen
//
// Severity (1-9) per WBH p.84 Taint Severity table.
// Persistence (2-9) per WBH p.84 Taint Persistence table.
//
// On atms outside 4-9 (A/B/C/F+), the Taint Subtype table is used for
// "irritants" with the same fields. Renderers distinguish T.S.P (taint)
// from I.S.P (irritant) by atm code, not by Taint type.
type Taint struct {
    Code        string
    Severity    int
    Persistence int
}

// Hazard — Insidious Atmosphere inherent hazard per WBH p.90.
//
// Code values:
//   B = Biologic
//   R = Radioactivity
//   G = Gas Mix
//   T = Temperature
//
// Hazards are inherently lethal and constant per WBH p.89; severity and
// persistence are not rolled.
type Hazard struct {
    Code string
}
```

## Public API

### Pre-existing oxygen taint promotion (WBH p.81)

```go
// PromoteOxygenTaint applies the WBH p.81 "tainted equivalent" rule:
// when an atm 5/6/8 has computed ppO2 outside [0.10, 0.50] bar, the code
// is promoted to its tainted equivalent (4/7/9) with low (ppO2 < 0.10)
// or high (ppO2 > 0.50) oxygen pre-seeded as the first taint subtype.
//
// For atms outside 5/6/8 or with ppO2 in band, returns (atmCode, nil).
//
// The pre-seeded Taint has Severity and Persistence == 0; Step 5D½'s
// orchestrator fills them from the severity/persistence rolls so callers
// don't have to special-case pre-seeded taints.
func PromoteOxygenTaint(atmCode int, ppO2 float64) (int, *Taint)
```

### Multi-taint orchestrator (WBH p.82-83)

```go
// RollAllTaints rolls the full taint profile for a body's atmosphere per
// WBH pp.81-84. Handles:
//   - Pre-seeded L/H taint from PromoteOxygenTaint (counts as taint #1)
//   - 2D + DMs roll on Taint Subtype table; result of 10 = particulates +
//     reroll; second 10 = third taint; max 3
//   - L/H suppression on non-4-9 codes and on 2nd/3rd rolls (→ G)
//   - Severity + Persistence per taint
//   - ppO2 / total-pressure adjustment when L/H rolled (book p.83):
//       low oxygen  → ppO2 -= 1D/100; replaced with N₂ at constant total P
//       high oxygen → ppO2 += 1D/10;  replaced with N₂ at constant total P
//
// Mutates body.Atmosphere fields directly when ppO2 adjustment fires.
// Returns the populated []Taint for assignment to body.Atmosphere.Taints.
func RollAllTaints(r roller.Roller, body *DetailedPlacement) []Taint
```

### Insidious Hazard (WBH p.90)

```go
// RollInsidiousHazard rolls the WBH p.90 Insidious Atmosphere Hazard table.
// Returns 1 hazard, or 2 hazards when the first roll is T (temperature)
// and the insidious subtype is D or E (footnote: "If the insidious subtype
// is D or E, a T hazard automatically exists, roll again for an additional
// hazard"). Returns the hazard codes; caller wraps the first into Hazard{}.
//
// DMs:
//   - "Atmosphere is extremely dense" (p.90 footnote): DM+2
//
// Returns:
//   - First call: hazard code from {B, R, G, T} per 2D + DMs
//   - Second call (only if first was T and subtype D/E): another hazard code
func RollInsidiousHazard(r roller.Roller, isExtremelyDense bool) string
```

### Predicate helpers for biology

```go
// HasTaintCode reports whether any Taint in the slice has the given code.
// Used by RollBiomass (Code "B"), RollBiocomplexity (Code "L"), and
// RollCompatibility (any taint present → "otherwise tainted" -2).
func HasTaintCode(taints []Taint, code string) bool

// HasAnyTaint reports whether the slice contains at least one Taint.
// Used by RollCompatibility for the "or otherwise tainted" qualifier.
func HasAnyTaint(taints []Taint) bool
```

## Biology hookup

Three single-line changes in `worlds/biology.go`:

### `RollBiomass` — biologic-taint Special Case 1 (WBH p.127)

After the existing biomass roll + clamp, add:

```go
// Special Case 1 (WBH p.127): biologic-taint forces biomass ≥ 1.
if biomass == 0 && body.Atmosphere != nil && HasTaintCode(body.Atmosphere.Taints, "B") {
    biomass = 1
}
```

Remove the deferral comment (lines 65-66 in current `biology.go`).

### `RollBiocomplexity` — low-oxygen DM-2 (WBH p.129)

In `RollBiocomplexity`, fold low-oxygen-taint detection into the DM accumulator:

```go
if body.Atmosphere != nil && HasTaintCode(body.Atmosphere.Taints, "L") {
    dm += -2
}
```

Remove the deferral comment (lines 207-209 in current `biology.go`).

### `RollCompatibility` — "otherwise tainted" -2 (WBH p.131)

The `compatibilityAtmDM` table already has -2 for atms 2, 4, 7, 9. Add a separate check for tainted atmospheres in codes outside that set (e.g. atm 5/6/8 that didn't get promoted but have other taints in edge cases — though promotion makes this rare, a tainted exotic/etc still counts):

```go
// "Or otherwise tainted" qualifier (WBH p.131): -2 for any tainted atm
// that didn't already get -2 from the code table.
if body.Atmosphere != nil && HasAnyTaint(body.Atmosphere.Taints) {
    code := body.Atmosphere.Code
    if !(code == 2 || code == 4 || code == 7 || code == 9) {
        dm += -2
    }
}
```

Remove the deferral comment (lines 322-323 in current `biology.go`).

## Rendering

`AtmosphereProfile.Shorthand` is extended at the point of construction (`worlds/atmosphere_profile.go`) to suffix the taint/irritant block per WBH atm-code-specific format strings:

| Atm code                  | Format                                           | Reference |
| ------------------------- | ------------------------------------------------ | --------- |
| 2, 4, 7, 9                | `A-bar-ppo:T.S.P[,T.S.P,...]`                    | WBH p.82  |
| A (10)                    | `A-St#:bar:gases I.S.P[,I.S.P,...]`              | WBH p.85  |
| B (11)                    | `B-St#:bar:gases I.S.P[,I.S.P,...]`              | WBH p.89  |
| C (12)                    | `C-St#.H:bar:gases I.S.P[,I.S.P,...]`            | WBH p.89  |
| 0, 1, 3, 5, 6, 8, D, E, F | unchanged (no taint suffix; pre-existing format) | —         |

Implementation: a private helper `formatTaintBlock(atmCode int, taints []Taint, hazard *Hazard) string` returns the suffix (or "" when no taints). Existing IISS Class IV-P render paths in `worlds/iiss_class4p.go` and `worlds/markdown.go` continue to print `Shorthand` directly — no change needed.

## Pipeline placement detail

Order within Step 5D½ (per body, per moon):

1. Read `dp.Atmosphere.Code` and `dp.Atmosphere.OxygenPartialPressure` (already finalized by Step 5D rederive).
2. Call `PromoteOxygenTaint(code, ppO2)` — may mutate `dp.Atmosphere.Code` from 5/6/8 to 4/7/9.
3. Call `RollAllTaints(r, dp)` — handles pre-seeded L/H, multi-roll, severity, persistence, ppO2 adjustment for fresh L/H rolls. Assign result to `dp.Atmosphere.Taints`.
4. If `dp.Atmosphere.Code == 12` (atm C), call `RollInsidiousHazard` once (and again if first hazard is T and subtype is D or E). Assign first hazard to `dp.Atmosphere.InsidiousHazard`.
5. Rebuild `dp.Atmosphere.Profile.Shorthand` with the taint suffix (atmosphere_profile rebuild call).

Moons follow the same path through `buildMoonPlacementView` — explicit moon iteration test required (per MEMORY anti-pattern note).

## Testing strategy

### Worked-example tests

WBH provides three explicit worked examples for taints/irritants that become regression tests:

- **WBH p.85 — Aab V d (radioactive desert moon)**: atm 4 (low-oxygen taint not auto-applied because ppO2 = 0.114 bar avoids automatic low-oxygen), then taint roll 12-2 = 10 → particulates (severity 6:hazardous, persistence 3:occasional and lingering); reroll for second taint = 5-2 = 3 → radioactivity (severity 5:serious, persistence 4:irregular). Expected `Taints == [{P, 6, 3}, {R, 5, 4}]`. Test name: `TestAabVd_TaintProfile_p85`.
- **WBH p.88 — Aab V b (exotic atmosphere)**: atm A subtype 9 (dense irritant); irritant roll 9+2 = 11 → radioactivity (severity 2:surmountable, persistence 9:constant). Expected `Atmosphere.Code == 10` (A), `Atmosphere.Subtype == "9"`, `Taints == [{R, 2, 9}]` for the irritant component. Test name: `TestAabVb_ExoticIrritant_p88`.
- **WBH p.90 — AaB VI (corrosive)**: book worked example does NOT roll an irritant (referee elects gas-mix as the corrosive cause, narratively). Our implementation always rolls — see "Book inconsistency" below. The test asserts `Atmosphere.Code == 11` (B), `Atmosphere.Subtype == "6"` (rolled 4 + DM+2 for Size 6 = standard), and that an irritant Taint is generated. Test name: `TestAaBVI_CorrosiveProfile_p90`.

#### Book inconsistency

WBH p.89 says "A corrosive or insidious atmosphere may include an irritant if indicated by type" — referee-elective phrasing. The Aab V b exotic example (p.88) rolls one; the AaB VI corrosive example (p.90) does not. For automated generation we choose **always roll** (consistency, and the more common book pattern). Document this as a feedback memo on merge.

Add to existing `TestZed_FullDetail_3A2b` acceptance test: assertions for taint typology on each touched body (per the precedent of 3B-biology / 3B-geology / 3B-final extending the same gate).

### Table integrity tests

- `TestTaintSubtypeTable_AllResults` — every (2D + DM) value in [0, 14] maps to a defined subtype
- `TestSeverityTable_RangeIs1to9`
- `TestPersistenceTable_RangeIs2to9`
- `TestInsidiousHazardTable_AllResults`

### Property tests

- `Taints` length never exceeds 3
- Severity always in [1, 9]; Persistence always in [2, 9]
- L/H never appears as 2nd or 3rd taint
- L/H never appears as a taint on atms outside 4-9 (treated as G)
- atm 5/6/8 never has Taints populated unless promotion happened (in which case Code is now 4/7/9)
- atm C always has `InsidiousHazard != nil` after Step 5D½
- atm A/B/F+ never has `InsidiousHazard != nil`

### Anti-pattern check (per MEMORY)

`TestRunStep5DHalf_MoonsVisited` — explicit assertion that moons receive Taint generation, including moons of gas-giant parents. This is the same anti-pattern the moon-path-diverges-from-planet-path memory describes.

## Carry-forward

None expected. This sub-project closes the three Q3-a deferrals identified in 3B-biology, 3B-final, and 3A2b-rederive specs. Issues #11 and #13 close on merge.

## Out of scope

- Severity/Persistence outcome strings (the "after 1D weeks acclimation" prose from p.84) — not rendered in the IISS form, no need to encode the prose.
- Optional ppO2-specific severity overrides for L/H taints (footnote on p.84) — implementation will use the default DM+4 path; the optional ppO2-specific severities can be added later if a renderer needs them.
- Exotic atmosphere gas-mix irritant special case (WBH p.86: "Treat such atmospheres as having a gas mix (G) irritant with a roll of 1D+9 on the Taint Severity table and a roll of 1D+1 on the Taint Persistence table") — this is a referee-elective rule for occasional corrosive components; out of scope for automatic generation.
- Taint Subtype to actual gas naming (per p.87 Atmospheric Gas Composition table) — the Profile already has a separate gas-mix component populated by 3A2b-rederive; this sub-project does not touch gas naming.
