# Atm B/C Pressure From Subtype — Design

**Date:** 2026-05-09
**Sub-project:** atmosphere-taint-typology follow-up
**Predecessors:** insidious-extremely-dense-from-subtype (PR #23, merged 2026-05-09)
**Closes:** issue #24

## Goal

Populate `atm.Pressure` for Corrosive (atm B, code 11) and Insidious (atm C, code 12) atmospheres using the per-subtype pressure ranges on WBH p.89, instead of the current `(0, 0)` "Varies" short-circuit in `AtmospherePressureRange` that leaves `atm.Pressure == 0` for every B/C body.

## Source of truth

WBH p.89 Corrosive and Insidious Atmosphere Subtype table (same data as the spec for issue #22, reproduced for self-containment):

| 2D  | Code | Atmosphere Type                              | Pressure Range (bar) | Span    |
| --- | ---- | -------------------------------------------- | -------------------- | ------- |
| 1-  | 1    | Very Thin, Temperature 50K or less           | 0.1–0.42             | 0.32    |
| 2   | 2    | Very Thin, Irritant                          | 0.1–0.42             | 0.32    |
| 3   | 3    | Very Thin                                    | 0.1–0.42             | 0.32    |
| 4   | 4    | Thin, Irritant                               | 0.43–0.70            | 0.27    |
| 5   | 5    | Thin                                         | 0.43–0.70            | 0.27    |
| 6   | 6    | Standard                                     | 0.70–1.49            | 0.79    |
| 7   | 7    | Standard, Irritant                           | 0.70–1.49            | 0.79    |
| 8   | 8    | Dense                                        | 1.50–2.49            | 0.99    |
| 9   | 9    | Dense, Irritant                              | 1.50–2.49            | 0.99    |
| 10  | A    | Very Dense                                   | 2.50–10.0            | 7.50    |
| 11  | B    | Very Dense, Irritant                         | 2.50–10.0            | 7.50    |
| 12  | C    | Extremely Dense                              | 10.0+                | unbound |
| 13  | D    | Extremely Dense, Temperature 500K+           | 10.0+                | unbound |
| 14+ | E    | Extremely Dense, Temperature 500K+, Irritant | 10.0+                | unbound |

Book quote (p.89): "More detailed atmospheric pressures for these atmospheres can be determined from pressure range (bars) and span using either of the total atmospheric pressure (bar) equations on page 80."

Book quote (p.89): "Only insidious extremely dense atmospheres should have pressures exceeding 1,000 bar."

The roll formula on p.80 is the same one already used by `RollTotalPressure`:

```text
bar = minBar + span × ((1D-1)×5 + (1D-1)) / 30
```

## Project-supplied resolution for "10.0+ / unbound" cells

Subtypes C/D/E carry `10.0+` with `unbound` span. The book defers to "the equations on page 80" but does not pin a span. Per brainstorming decision (b) — tiered by subtype letter, honoring p.89's "1,000+ bar" hint:

| Subtype | minBar | spanBar | Resulting range |
| ------- | ------ | ------- | --------------- |
| C       | 10     | 90      | 10–100 bar      |
| D       | 100    | 900     | 100–1000 bar    |
| E       | 1000   | 9000    | 1000–10000 bar  |

Subtype E (insidious extremely dense + temperature 500K+ + irritant) lands above the 1,000 bar floor mentioned by the book. Subtype D (extremely dense + temperature 500K+) sits at Venus-like magnitudes. Subtype C (extremely dense alone) stays below 100 bar.

The band edges are project-supplied. Inline comments in `corrosiveInsidiousPressureRange` cite this spec.

## Architecture

### New helper: `corrosiveInsidiousPressureRange`

Pure function in `worlds/atmosphere.go`, alongside `AtmospherePressureRange`:

```go
// corrosiveInsidiousPressureRange returns (minBar, spanBar) for atm
// codes B (11) and C (12) keyed off the WBH p.89 subtype letter.
//
// Subtypes 1-B carry the explicit ranges from the p.89 table.
// Subtypes C/D/E carry "10.0+ / unbound" in the book; this function
// returns project-supplied tiered ranges per the design spec dated
// 2026-05-09 (atm-bc-pressure-from-subtype):
//
//   C: 10–100      (min=10,   span=90)
//   D: 100–1000    (min=100,  span=900)
//   E: 1000–10000  (min=1000, span=9000)
//
// Empty/unknown subtype returns (0, 0). Callers in the live pipeline
// always roll the subtype before pressure (see system_detail_steps.go
// and temperature_rederive.go).
func corrosiveInsidiousPressureRange(subtype string) (minBar, spanBar float64)
```

### `RollTotalPressure` signature change

Extend the existing function to accept a subtype:

```go
// RollTotalPressure computes total atmospheric pressure per WBH p.80,
// or per the WBH p.89 subtype-keyed range for atm B/C.
//
// For atm codes 11 (B) and 12 (C), subtype must be set (one of
// "1"-"9", "A"-"E"); otherwise pressure falls back to 0 (legacy
// "Varies" behavior). For all other codes, subtype is ignored.
//
// The roll formula is unchanged from the prior signature:
//
//   bar = minBar + span × ((1D-1)*5 + (1D-1)) / 30
//
// For atms with span == 0 (codes returning 0/0 from
// AtmospherePressureRange and a B/C with empty subtype), no rolls are
// consumed and the function returns minBar.
func RollTotalPressure(r roller.Roller, atmoCode int, subtype string) (float64, error)
```

The dispatch inside `RollTotalPressure`:

1. If `atmoCode == 11 || atmoCode == 12` and `subtype != ""`, call `corrosiveInsidiousPressureRange(subtype)`.
2. Otherwise, call `AtmospherePressureRange(atmoCode)` as today.
3. Apply the existing roll formula.

This keeps the regular-atm path untouched and routes B/C specially.

### Call-site updates

Three call sites, all already roll subtype before pressure:

- `worlds/system_detail_steps.go:95` (planet path) — pass `atmo.Subtype`.
- `worlds/system_detail_steps.go:147` (moon path) — pass `atmo.Subtype`.
- `worlds/temperature_rederive.go:203` (rederive pass) — pass `newSubtype` (already in scope).

For non-11/12 atm codes, `atmo.Subtype` will typically be `""` (unset) and is ignored.

## Out of scope

- **Atm F/G/H (15/16/17) "Varies" pressures.** The book describes these as Unusual / Helium / Hydrogen — `Varies` with no subtype-keyed table. They stay at `(0, 0)`.
- **Atm A "Exotic" subtype-keyed pressures.** Atm A subtypes exist (the spec for atmosphere-taint-typology cites Aab V b at "subtype 9, 0.55 bar"), but WBH p.85 describes exotic pressures qualitatively, not via a table. Out of scope until the book gives explicit ranges or the project decides on a derivation.
- **`ComputeGreenhouseFactor` (`temperature_greenhouse.go`).** Real B/C pressures will start producing real `0.5 × √P` initial values, but the existing `(1+G)` clamp at `[0.001, 1.999]` (WBH p.111 thumb-rule-two) saturates large values and prevents temperature blowup. No formula change.
- **WBH p.89 footnote 1 (subtype D/E auto-T hazard plus additional rolled hazard).** Tracked as #21. Independent of pressure.

## Testing strategy

### Unit tests for the helper

`TestCorrosiveInsidiousPressureRange` in `worlds/atmosphere_test.go`:

- Subtypes "1", "2", "3" → (0.1, 0.32).
- Subtypes "4", "5" → (0.43, 0.27).
- Subtypes "6", "7" → (0.70, 0.79).
- Subtypes "8", "9" → (1.50, 0.99).
- Subtypes "A", "B" → (2.50, 7.50).
- Subtype "C" → (10, 90).
- Subtype "D" → (100, 900).
- Subtype "E" → (1000, 9000).
- Empty / unknown ("Z", "", "0") → (0, 0).

### Behavior tests for `RollTotalPressure`

`TestRollTotalPressure_AtmBCWithSubtype` in `worlds/atmosphere_test.go`:

- `RollTotalPressure(r, 11, "6")` with scripted dice → value in [0.70, 1.49].
- `RollTotalPressure(r, 12, "C")` → value in [10, 100].
- `RollTotalPressure(r, 12, "E")` → value in [1000, 10000].
- `RollTotalPressure(r, 11, "")` → 0 (legacy fallback).
- `RollTotalPressure(r, 6, "")` → unchanged regular-code behavior (uses `AtmospherePressureRange(6)`).
- `RollTotalPressure(r, 6, "X")` → subtype ignored for non-11/12 codes; same as above.

Spot-check rolls: scripted dice `(1, 1)` produces `scale = 0` → `minBar`; scripted `(6, 6)` produces `scale = 1.0` → `minBar + span`.

### Existing tests

- `TestRunStep5DPrime_AtmCGetsHazard` and similar — pass-through. Subtype is set in test fixtures; signature change requires updating the existing call paths but no behavior change.
- All test files that mock `Atmosphere{Code: 11, ...}` or `Code: 12, ...` need a Subtype field added (or the test acknowledges it's testing pre-subtype state and explicitly passes "").

### Zed golden

Likely shifts. Seed=42's atm B/C bodies (Zed Prime is atm B subtype 6 per the existing golden) will get real pressure rolls instead of 0.0, which cascades into:

- Profile shorthand line for atm B (`B-St6:bar:gases ...`) gets a real bar value where it previously rendered as 0 or absent.
- Any downstream rendering or computation that read `atm.Pressure` from a B/C body shifts.

Refresh after implementation; verify the diff is limited to atm B/C bodies' pressure-derived fields and their downstream RNG drift.

## Carry-forward

None. This closes #24.

If post-merge experience shows the C/D/E band edges produce unrealistic systems (e.g., too many 5000+ bar atmospheres), revisit by adjusting the table — that's a tuning change, not a structural one.
