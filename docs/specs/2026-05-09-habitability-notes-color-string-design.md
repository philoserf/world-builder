# Habitability Notes Referee-Color String — Design

**Date:** 2026-05-09
**Sub-project:** 3B-final follow-up
**Predecessors:** 3B-final (`main`, merged 2026-05-06)
**Closes:** issue #15

## Goal

Populate the existing `Habitability.Notes` field with a referee-color synthesis of which WBH p.132 DM rules fired on a given body, and surface it through the Class IV-P renderers. Today the field exists with a doc-comment ("Currently always empty — populated by future referee-feature carry-forward") but `ComputeHabitability` leaves it empty and neither renderer reads it.

## Source of truth

WBH p.132 Habitability Rating DM table includes a description column for every DM rule. The Zed Prime worked example on p.133 confirms the intent of these descriptions as referee-color annotations: "Zed Prime is only regionally habitable, mainly because of the heat… Gravity is only 0.66, so that warrants a DM-1."

The descriptions, in book order:

| DM rule            | Description column                           |
| ------------------ | -------------------------------------------- |
| Size 0–4           | Limited surface area                         |
| Size 9+            | Additional surface area                      |
| Atm 0, 1, A        | Non-breathable atmosphere                    |
| Atm 2 or E         | Very thin, tainted, or thin, low atmospheres |
| Atm 3 or D         | Very thin or very dense atmosphere           |
| Atm 4 or 9         | Tainted thin or dense atmospheres            |
| Atm 5, 7, 8        | Thin, taint (standard), or dense Atmospheres |
| Atm B              | Hostile Atmosphere                           |
| Atm C or F+        | Very hostile Atmosphere                      |
| Hydro 0            | Lack of accessible liquid water              |
| Hydro 1–3          | Desert conditions prevalent                  |
| Hydro 9            | Little useable land surface area             |
| Hydro A            | Very little useable land surface area        |
| Tidal lock 1:1     | Very little useable land surface area        |
| High temp > 323K   | Too hot at times                             |
| High temp < 279K   | Too cold all of the time                     |
| Mean temp > 323K   | Too hot most of the time                     |
| Mean temp 304–323K | Too hot most of the time                     |
| Mean temp < 273K   | Too cold most of the time                    |
| Low temp < 200K    | Much too cold some of the time               |
| Gravity < 0.2      | Unhealthy low gravity levels                 |
| Gravity 0.2–0.7    | Very low gravity                             |
| Gravity 0.4–0.7    | Low gravity                                  |
| Gravity 0.7–0.9    | Gravity very comfortable                     |
| Gravity 1.1–1.4    | Gravity somewhat high                        |
| Gravity 1.4–2.0    | Gravity uncomfortably high                   |
| Gravity > 2.0      | Gravity too high for acclimation             |

Strings are book-quoted verbatim.

## Decisions

### Notes content

For each DM rule that fires on a body (positive or negative DM), append the book's description-column phrase. Joined with `"; "`. Empty when no DMs fire (Terra-equivalent baseline).

Positive-DM rules (Size 9+ → "Additional surface area"; Gravity 0.7–0.9 → "Gravity very comfortable") are included. The string is "referee-color" — the remarks are useful regardless of sign.

### Source-of-truth coupling

Note text and DM value live together in each helper. A future maintainer cannot update the DM band without seeing the description, and cannot rename a description without seeing the band — eliminating drift.

### Helper signatures

Each `habitability*DM` helper returns both the DM and the note text:

```go
func habitabilitySizeDM(size int) (int, string)
func habitabilityAtmDM(body *DetailedPlacement) (int, string)
func habitabilityHydroDM(body *DetailedPlacement) (int, string)
func habitabilityTidalLockDM(body *DetailedPlacement) (int, string)
func habitabilityGravityDM(body *DetailedPlacement) (int, string)

func habitabilityTempDM(body *DetailedPlacement) (int, []string)
```

The temperature helper returns `[]string` because its 5 sub-conditions (HighK > 323, HighK < 279, MeanK > 323, MeanK 304–323, MeanK < 273, LowK < 200) can fire independently and each contributes a separate description.

Empty string ("" or empty slice) means no DM fired.

### `ComputeHabitability` orchestration

Collect non-empty notes from each helper into a `[]string` in book order (Size, Atm, Hydro, TidalLock, Temp, Gravity). Join with `"; "` and store in `Habitability.Notes`.

The DM sum logic is unchanged.

### Renderer format

Both Class IV-P render paths gain a Notes row/line, shown only when `Notes != ""`.

**Markdown** (`writeClass4PHabitability` in `worlds/markdown.go`):

```
| Rating | 7 — Regionally habitable |
| Notes | Too hot most of the time; Low gravity |
```

**IISS Class 4P** (`renderIISS4PHabitability` in `worlds/iiss_class4p.go`):

```
HABITABILITY
  Rating: 7
  Notes:  Too hot most of the time; Low gravity
```

Both omit the row/line when `Notes == ""`.

## Architecture

### `worlds/habitability.go` — refactor helpers, populate Notes

Each helper changes signature as described above. `ComputeHabitability` becomes:

```go
func ComputeHabitability(body *DetailedPlacement) Habitability {
    if body == nil {
        return Habitability{Rating: 0}
    }
    var notes []string
    addNote := func(s string) {
        if s != "" {
            notes = append(notes, s)
        }
    }

    sizeDM, sizeNote := habitabilitySizeDM(SizeAsInt(body.SizeCode))
    addNote(sizeNote)
    atmDM, atmNote := habitabilityAtmDM(body)
    addNote(atmNote)
    hydroDM, hydroNote := habitabilityHydroDM(body)
    addNote(hydroNote)
    tidalDM, tidalNote := habitabilityTidalLockDM(body)
    addNote(tidalNote)
    tempDM, tempNotes := habitabilityTempDM(body)
    for _, n := range tempNotes {
        addNote(n)
    }
    gravDM, gravNote := habitabilityGravityDM(body)
    addNote(gravNote)

    rating := min(max(10+sizeDM+atmDM+hydroDM+tidalDM+tempDM+gravDM, 0), 12)
    return Habitability{Rating: rating, Notes: strings.Join(notes, "; ")}
}
```

The doc-comment on the `Notes` field updates: "Currently always empty…" → "Populated by `ComputeHabitability` from WBH p.132 DM descriptions."

### `worlds/markdown.go` — render Notes row

Conditional row in `writeClass4PHabitability`:

```go
fmt.Fprintf(sb, "| Rating | %d — %s |\n",
    body.Habitability.Rating, HabitabilityRatingName(body.Habitability.Rating))
if body.Habitability.Notes != "" {
    fmt.Fprintf(sb, "| Notes | %s |\n", body.Habitability.Notes)
}
sb.WriteString("\n")
```

### `worlds/iiss_class4p.go` — render Notes line

```go
fmt.Fprintf(sb, "  Rating: %d\n", body.Habitability.Rating)
if body.Habitability.Notes != "" {
    fmt.Fprintf(sb, "  Notes:  %s\n", body.Habitability.Notes)
}
sb.WriteString("\n")
```

## Testing strategy

### `ComputeHabitability_Notes` integration test

`TestComputeHabitability_Notes` in `worlds/habitability_test.go`:

- Terra-equivalent body (Size 8, Atm 6, Hydro 7, no tidal lock, mean temp 290K, gravity 1.0) → `Notes == ""`. (Earth-baseline produces no fired-DM descriptions.)
- Zed-Prime-like body (Size 5, Atm 6, Hydro 5, HighK 346, MeanK 290, gravity 0.66) → `Notes` contains "Too hot at times" and "Low gravity", joined with "; ". This mirrors the WBH p.133 worked example ("its high temperature of 346 exceeds 323 for a DM-2 but mean temperature is within bounds").
- Hostile body (Size 8, Atm 11/B, Hydro 7, tidal-locked, mean temp 290K, gravity 1.0) → `Notes` contains "Hostile Atmosphere" and "Very little useable land surface area".
- Multi-note temperature: HighK 350, MeanK 340, LowK 250 → notes include "Too hot at times", "Too hot most of the time" (no "Too cold most of the time" since MeanK ≥ 273).

### Per-helper unit tests

Add note-assertion to existing per-helper tests (e.g., `TestHabilitySizeDM` if it exists, or new tests):

- Size 4 → `(-1, "Limited surface area")`.
- Size 9 → `(+1, "Additional surface area")`.
- Atm 11 → `(-10, "Hostile Atmosphere")`.
- Gravity 0.66 → `(-1, "Low gravity")`.
- TidalLock IsTwilightZone → `(-2, "Very little useable land surface area")`.
- Temp HighK=350, MeanK=340, LowK=250 → `(-6, ["Too hot at times", "Too hot most of the time"])`.

### Renderer tests

- Update `TestRenderIISSClass4P_HabitabilitySection_Present` to populate Notes and assert the `Notes:` line is rendered.
- Add a markdown test asserting the Notes row appears when populated and is omitted when empty.

### Zed golden

Will shift — Aab IV (Zed Prime) currently has Habitability with no Notes; after this change, Notes will contain "Too hot most of the time; Low gravity" (or similar) per the worked example on WBH p.133. Refresh after implementation.

## Out of scope

- **Atmosphere taint DMs.** Low-oxygen-taint DM-2 from p.132 is still skipped per the existing comment in `ComputeHabitability` ("Skipped: low-oxygen-taint DM-2 deferred per spec Q3-a"). Adding it is a separate concern; the taint typology landed in PRs #20/#23/#25, but wiring it into Habitability hasn't been requested.
- **WBH p.133 miscellaneous scoring** ("D3-1 referee-elective adjustment"). That's a separate referee-input mechanism, not a DM rule with a description.
- **Other renderers.** Only the two Class IV-P render paths are touched; no other code reads `Habitability.Notes`.

## Closes

#15.
