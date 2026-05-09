# Insidious "Extremely Dense" DM from Subtype — Design

**Date:** 2026-05-09
**Sub-project:** atmosphere-taint-typology follow-up
**Predecessors:** atmosphere-taint-typology (`main`, merged 2026-05-09 via PR #20)
**Closes:** issue #22

## Goal

Make the WBH p.90 "Atmosphere is extremely dense → DM+2" rule actually fire on the Insidious Atmosphere Hazard roll.

The current code at `worlds/system_detail_step5dprime.go:51` reads:

```go
isExtremelyDense := atm.Pressure >= 10.0
```

Because `AtmospherePressureRange(12)` returns `(0, 0)` (atm C is modeled as "Varies"), `RollTotalPressure` leaves `atm.Pressure == 0` for every insidious body. The threshold is unreachable in the live pipeline; the DM never applies.

## Source of truth

WBH p.89 Corrosive and Insidious Atmosphere Subtype table:

| 2D  | Code  | Atmosphere Type                                  | Pressure Range (bar) | Span    |
| --- | ----- | ------------------------------------------------ | -------------------- | ------- |
| 1-  | 1     | Very Thin, Temperature 50K or less               | 0.1–0.42             | 0.32    |
| 2   | 2     | Very Thin, Irritant                              | 0.1–0.42             | 0.32    |
| 3   | 3     | Very Thin                                        | 0.1–0.42             | 0.32    |
| 4   | 4     | Thin, Irritant                                   | 0.43–0.70            | 0.27    |
| 5   | 5     | Thin                                             | 0.43–0.70            | 0.27    |
| 6   | 6     | Standard                                         | 0.70–1.49            | 0.79    |
| 7   | 7     | Standard, Irritant                               | 0.70–1.49            | 0.79    |
| 8   | 8     | Dense                                            | 1.50–2.49            | 0.99    |
| 9   | 9     | Dense, Irritant                                  | 1.50–2.49            | 0.99    |
| 10  | A     | Very Dense                                       | 2.50–10.0            | 7.50    |
| 11  | B     | Very Dense, Irritant                             | 2.50–10.0            | 7.50    |
| 12  | **C** | **Extremely Dense**                              | **10.0+**            | unbound |
| 13  | **D** | **Extremely Dense**, Temperature 500K+           | **10.0+**            | unbound |
| 14+ | **E** | **Extremely Dense**, Temperature 500K+, Irritant | **10.0+**            | unbound |

WBH p.90 footnote 2 confirms: "Atmosphere is extremely dense DM+2".

So "extremely dense" maps directly to subtype letters **C, D, E**.

## Decision

Pick (c) from brainstorming: minimal fix only. Do not also populate `atm.Pressure` for atm B/C from the p.89 ranges in this sub-project — that is its own design with its own RNG-drift event.

## Architecture

### Helper: `isExtremelyDenseSubtype`

Private function in `worlds/atmosphere_taint.go`:

```go
// isExtremelyDenseSubtype reports whether a Corrosive/Insidious atmosphere
// subtype letter is "Extremely Dense" per WBH p.89 — rows 12/13/14+
// (codes C/D/E).
func isExtremelyDenseSubtype(subtype string) bool {
    switch subtype {
    case "C", "D", "E":
        return true
    }
    return false
}
```

### Call-site change

In `worlds/system_detail_step5dprime.go::computeBodyTaints`, replace

```go
// isExtremelyDense currently never fires: insidious atms model
// pressure as "Varies" (AtmospherePressureRange returns 0/0), so
// RollTotalPressure leaves atm.Pressure == 0 and the threshold is
// unreachable. The book's "extremely dense" criterion likely maps
// to one of the high subtype letters, but the WBH text doesn't
// pin down which — see the follow-up issue tracking that mapping.
if atm.Code == 12 {
    isExtremelyDense := atm.Pressure >= 10.0
    hazardCode := RollInsidiousHazard(r, isExtremelyDense)
    atm.InsidiousHazard = &Hazard{Code: hazardCode}
}
```

with

```go
if atm.Code == 12 {
    isExtremelyDense := isExtremelyDenseSubtype(atm.Subtype)
    hazardCode := RollInsidiousHazard(r, isExtremelyDense)
    atm.InsidiousHazard = &Hazard{Code: hazardCode}
}
```

The stale inline comment (calling the branch unreachable) is removed because the branch is no longer dead.

## Pre-flight verification (during implementation)

- Confirm `atm.Subtype` is populated for atm 12 by the time `runStep5DPrime` runs. `RollCorrosiveInsidiousSubtype` is invoked in 3A1 atmosphere generation; 5D-prime sits between 5D (3A2b-rederive) and 5E (3B-geology), so subtype must be set by then. If not, `isExtremelyDenseSubtype("")` returns false and the DM silently stays off — same failure mode as today, no regression — but worth verifying.
- Run the live pipeline with the Zed seed and inspect whether any atm-C body lands on subtype C/D/E. If yes, the hazard outcome may shift and the Zed golden refresh is required. If no, golden is untouched.

## Testing strategy

### Unit test for the helper

`TestIsExtremelyDenseSubtype` in `worlds/atmosphere_taint_test.go`:

- "C", "D", "E" → true
- "1", "6", "9", "A", "B", "F", "" → false

### Integration test for the call-site wiring

`TestRunStep5DPrime_ExtremelyDenseSubtypeDM` in `worlds/system_detail_step5dprime_test.go`:

- Subtype "D", scripted hazard roll `2D = 4`: expect hazard code `"G"` (4 + DM+2 = 6 → G).
- Subtype "6", scripted hazard roll `2D = 4`: expect hazard code `"B"` (4 + 0 = 4 → B).

This proves the DM is wired end-to-end from `atm.Subtype` through `RollInsidiousHazard`.

### Regression coverage that should keep passing

- `TestRollInsidiousHazard_ExtremelyDenseDM` (calls `RollInsidiousHazard(r, true)` directly). Unchanged.
- `TestRunStep5DPrime_AtmCGetsHazard` (uses subtype `"6"`). Unchanged — and now it indirectly asserts that subtype "6" does NOT trigger the DM.

## Carry-forward

File a new issue: **"populate `atm.Pressure` for atm B/C from subtype-keyed range per WBH p.89."**

Why deferred:

- Requires a new pressure-range table keyed by subtype letter and a new roll path.
- Pipeline-ordering check: must run after `RollCorrosiveInsidiousSubtype` and replace the current `(0,0)` short-circuit in `RollTotalPressure`.
- RNG drift event on goldens.
- Independent value: real pressures unlock downstream rendering and any future logic that conditions on atmospheric pressure for B/C bodies.

## Out of scope

- WBH p.90 footnote 1 (subtype D/E auto-T hazard plus additional rolled hazard). Tracked as issue #21; needs `Atmosphere.InsidiousHazard *Hazard` → `[]Hazard`.
- Atm B (corrosive) hazard rolls. Atm B does not roll for insidious hazards; the DM only matters for atm C.
- Pressure derivation for atm B/C (see Carry-forward).
