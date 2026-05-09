# Insidious D/E Hazard Rule — Design

**Date:** 2026-05-09
**Sub-project:** atmosphere-taint-typology follow-up
**Predecessors:** insidious-extremely-dense-from-subtype (PR #23, merged 2026-05-09); atm-bc-pressure-from-subtype (PR #25, merged 2026-05-09)
**Closes:** issue #21

## Goal

Implement the WBH p.90 Insidious Atmosphere Hazard footnote: when the insidious subtype is D or E, a Temperature hazard automatically exists, plus an additional hazard is rolled. The current code rolls a single hazard regardless of subtype; the docstring on `RollInsidiousHazard` already flags this as "not implemented."

## Source of truth

WBH p.90 Insidious Atmosphere Hazard table:

| 2D  | Hazard        | Hazard Code |
| --- | ------------- | ----------- |
| 4-  | Biologic      | B           |
| 5   | Radioactivity | R           |
| 6   | Gas Mix       | G           |
| 7   | Gas Mix       | G           |
| 8   | Temperature   | T\*         |
| 9   | Gas Mix       | G           |
| 10  | Temperature   | T\*         |
| 11  | Radioactivity | R           |
| 12+ | Temperature   | T\*         |

WBH p.90 footnote: "If the insidious subtype is D or E, a T hazard automatically exists, roll again for an additional hazard."

WBH p.90 second footnote: "Atmosphere is extremely dense DM+2" (already implemented; honored by `RollInsidiousHazard` for atm-C subtype C/D/E).

## Decisions

### Field shape

Replace `Atmosphere.InsidiousHazard *Hazard` (single pointer) with `Atmosphere.InsidiousHazards []Hazard` (slice). Empty slice (or nil) = no hazards.

This requires renaming the struct field, updating the renderer in `atmosphere_profile.go`, and updating every test fixture and assertion that touches the old field.

### D/E semantics

For atm code 12 (Insidious):

- Subtype is `D` or `E` → 2 hazards. First entry is `Hazard{Code: "T"}` (the automatic one); second entry is the rolled hazard.
- All other subtypes (1-9, A, B, C) → 1 hazard, the rolled one.

The DM+2 ("extremely dense") still applies only to the rolled hazard, not the auto-T. The auto-T is a fixed addition; DM+2 governs which hazard the dice produce.

### Duplicate-T case

If subtype is D or E and the rolled hazard happens to also be `T` (rows 8, 10, 12+), the result is `[T, T]`. The book footnote does not direct a reroll, so the implementation does not introduce one.

### Renderer format

Today's `FormatAtmoProfileShorthand` for atm C produces `C-St<subtype>.<H>:<bar>:<gases>` where `<H>` is the single hazard code. With multiple hazards, concatenate the single-letter codes after the dot. Hazard codes are always single letters (B / R / G / T), so concatenation is unambiguous and stays compact.

Examples:

- subtype `6` with rolled `G` → `C-St6.G:1.21:N2-95...`
- subtype `D` with rolled `G` → `C-StD.TG:120:...` (auto-T then rolled-G)
- subtype `E` with rolled `T` → `C-StE.TT:5000:...` (duplicate-T case)

## Architecture

### `worlds/atmosphere_taint.go` — new orchestrator

Add a higher-level function alongside the existing `RollInsidiousHazard` (which stays as the dice primitive):

```go
// RollInsidiousHazards applies the full WBH p.90 hazard procedure for
// atm C, including the footnote rule for subtypes D and E.
//
// Behavior:
//   - Subtypes D, E: returns 2 hazards. The first is Hazard{Code: "T"}
//     per the p.90 footnote ("a T hazard automatically exists"); the
//     second is rolled via RollInsidiousHazard.
//   - All other subtypes: returns 1 rolled hazard.
//
// isExtremelyDense applies DM+2 to the rolled hazard only (the auto-T
// is fixed). Pass the same flag the caller would pass to
// RollInsidiousHazard.
//
// If subtype D/E rolls a T as the additional hazard, the result is
// [T, T] — the book footnote doesn't direct a reroll, so neither do we.
func RollInsidiousHazards(r roller.Roller, subtype string, isExtremelyDense bool) []Hazard
```

The "Not implemented" note on the existing `RollInsidiousHazard` docstring is removed (it's now the dice primitive — no missing semantics).

### `worlds/atmosphere.go` — field rename

```go
type Atmosphere struct {
    // ... existing fields ...
    Taints           []Taint
    InsidiousHazards []Hazard
}
```

The doc-comment on the type updates `InsidiousHazard` → `InsidiousHazards`.

### `worlds/system_detail_step5dprime.go` — call-site

```go
if atm.Code == 12 {
    isExtremelyDense := isExtremelyDenseSubtype(atm.Subtype)
    atm.InsidiousHazards = RollInsidiousHazards(r, atm.Subtype, isExtremelyDense)
}
```

The `atm.InsidiousHazards = nil` clear at the top of `computeBodyTaints` follows the same shape as the existing `InsidiousHazard = nil` clear.

### `worlds/atmosphere_profile.go` — renderer

```go
if atmo.Code == 12 && len(atmo.InsidiousHazards) > 0 {
    var codes strings.Builder
    for _, h := range atmo.InsidiousHazards {
        codes.WriteString(h.Code)
    }
    subtypeWithHazard = atmo.Subtype + "." + codes.String()
}
```

The doc-comment on `FormatAtmoProfileShorthand` updates the example for the C case to show the multi-hazard form.

## Out of scope

- `worlds/markdown.go` and `worlds/iiss_class4p.go`. They consume the profile shorthand string (the output of `FormatAtmoProfileShorthand`); they do not touch `InsidiousHazard(s)` directly. The shorthand format change flows through automatically.
- The `Hazard` struct shape itself — stays `{Code string}`. No severity/persistence per WBH p.89 ("hazards are inherently lethal and constant"). The slice change is what's required to represent multiple, nothing else.

## Testing strategy

### New unit tests for `RollInsidiousHazards`

`TestRollInsidiousHazards` in `worlds/atmosphere_taint_test.go`:

- Subtype `6`, no DM, scripted hazard 2D=4 → `[Hazard{B}]` (single).
- Subtype `C`, isExtremelyDense=true, scripted hazard 2D=4 → `[Hazard{G}]` (single, DM+2 fires: 4+2=6 → G).
- Subtype `D`, isExtremelyDense=true, scripted hazard 2D=4 → `[Hazard{T}, Hazard{G}]` (auto-T + rolled with DM+2).
- Subtype `E`, isExtremelyDense=true, scripted hazard 2D=2 → `[Hazard{T}, Hazard{B}]` (auto-T + rolled-B with DM+2: 2+2=4 → B).
- Duplicate case: subtype `D`, isExtremelyDense=false, scripted hazard 2D=8 → `[Hazard{T}, Hazard{T}]` (rolled lands on T row 8).

### Existing tests

`RollInsidiousHazard` (singular) tests in `worlds/atmosphere_taint_test.go` are unchanged — the dice primitive's signature and behavior are preserved.

### Test fixtures + assertions

Mechanical updates everywhere `InsidiousHazard` (singular) appears:

- `worlds/atmosphere_profile_test.go:225` fixture `InsidiousHazard: &Hazard{Code: "T"}` → `InsidiousHazards: []Hazard{{Code: "T"}}`.
- `worlds/system_detail_step5dprime_test.go:39, 40, 56, 57, 153, 154, 156` — `nil` checks become `len(...) == 0` / `len(...) > 0`; `.Code` accesses become `[0].Code`.
- `worlds/system_detail_step5dprime_test.go::TestRunStep5DPrime_ExtremelyDenseSubtypeDM` — the subtype-`D` subcase now expects 2 hazards: `[T, G]` (auto-T then rolled-G with DM+2 from 2D=4); subtype-`6` subcase expects 1 hazard `[B]`. The current test asserts only the first hazard's code; expand to also assert the slice length.

### Renderer test

Add a new case to `worlds/atmosphere_profile_test.go::TestFormatAtmoProfileShorthand` that exercises the multi-hazard concatenation: atm C subtype `D` with `InsidiousHazards: []Hazard{{Code: "T"}, {Code: "G"}}` → output should contain `C-StD.TG:`.

### Zed golden

Likely unchanged — the current golden has no atm-C body. If seed=42 happens to land one (which the existing pipeline already runs through 5D-prime), the renderer's hazard-concatenation change would shift output. Verify after implementation; refresh only if affected.

## Carry-forward

None. This closes the last open follow-up from PR #20's review.
