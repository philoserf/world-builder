# IISS Class IV-P PART P.B (belt mainworld) — Design

**Date:** 2026-05-06
**Status:** approved by user; ready for plan
**Source material:** WBH pp. 138–139 (form template), p. 73 (belt characteristics), p. 134 (mainworld determination criteria).
**Carries forward from:** `docs/pass-1/specs/2026-05-05-world-physical-3b-final-design.md`, Q5(a) which deferred PART P.B as a stub.

## Purpose

Complete the Class IV-P renderer so it covers every mainworld type. The current implementation handles terrestrial planets and moons (Form 0407F-IV PART P); this spec adds the belt variant (Form 0407K-IV PART P.B) and extends `pickMainworld` so belts can actually be selected as the system mainworld.

This finishes the rules side of the project per the scope memo: with PART P.B in place, every body the auto-pick logic can return has a working Class IV-P renderer, and the Markdown output layer (next sub-project) can render any system without falling back to a placeholder.

## Non-goals

- **Per-body detail in the Major Bodies subtable.** The book's form template includes a repeating subtable (SAH/UWP, Orbit#, Diameter, Density, Mass per body), but our generator only computes counts of size-1 and size-S significant bodies. Generating per-body detail is a separate, larger sub-project.
- **World maps** (deferred item (o), pp. 135–137) — visual hex grids; not a textual form field.
- **Referee mainworld override** (deferred item (q)) — a `-mainworld <designation>` CLI flag.
- **Travel Zone / Initial Survey / Last Updated values.** These remain caller-supplied header metadata, identical to the existing PART P treatment.

## Scope decomposition

Two work units, both small and committable independently:

1. Extend `pickMainworld` to admit belts as candidates.
2. Implement `renderIISS4PBelt` and replace the stub dispatch.

Order matters: unit 1 makes unit 2's renderer reachable from real generated systems, but unit 2 can be tested with synthetic placements regardless. Either order works for implementation; this spec presents them in dependency order.

## Mainworld selection extension

### Current behavior

`worlds/mainworld.go:113-126` — `collect` returns early when `bodyType != BodyTerrestrial`. Belts and gas giants are never candidates. The priority chain is:

1. Sophont present (extant or extinct)
2. Highest Habitability > 0
3. Highest ResourceRating > 0
4. First terrestrial in iteration order

### New behavior

`collect` admits both `BodyTerrestrial` and `BodyPlanetoidBelt`. For a belt, the candidate is populated as:

- `habitability = 0` (belts have no `*Habitability`)
- `hasSophont = false` (belts have no `*Biology`)
- `resource = body.Belt.ResourceRating` (read from `*BeltDetails`, not `*Biology`)

The existing priority chain is unchanged. Belts naturally lose priorities 1 and 2 (no sophont, zero habitability), compete with terrestrials in priority 3 on resource rating, and remain eligible for the priority 4 fallback.

### Why this shape

WBH p. 134 lists "best refuelling location" as a mainworld criterion. Belts can satisfy this — they're literally where Traveller refuelling happens for skimming-incapable ships. The book makes mainworld picking explicitly Referee judgment, so a fall-through to belts when no terrestrial qualifies fits the book's intent.

This also closes deferred memory item (s): the priority-4 fallback no longer admits only non-habitable rock moons; it admits whatever the iteration first finds, which may be a belt.

### Edge cases

- **Belt with `Belt == nil`.** Should not happen — `runDetailPipeline` always populates `*BeltDetails` for `BodyPlanetoidBelt`. If it does happen, treat as `resource = 0` (belt loses priority 3, may still win priority 4).
- **Tie at priority 3 between a terrestrial and a belt with equal resource ratings.** First-iteration-order wins, same as the existing terrestrial-vs-terrestrial tie rule.
- **System with only belts.** Returns the belt with highest resource rating, or the first belt at priority 4 if all belts have rating 0 (which can't happen — `RollResourceRating` clamps to ≥ 2).

### Test additions

`mainworld_test.go`:

- `TestPickMainworld_BeltOnlySystem_ReturnsBeltWithHighestResource` — synth system with two belts and one barren GG; assert highest-resource belt wins.
- `TestPickMainworld_TerrestrialBeatsBelt_OnHabitability` — synth with one habitability-1 terrestrial and one resource-12 belt; assert terrestrial wins (priority 2 beats priority 3).
- `TestPickMainworld_TerrestrialAndBelt_TieOnResource_IterationOrder` — both at habitability 0, both at resource 8; assert first-iteration-order wins.

## Form rendering

### File layout

New file: `worlds/iiss_class4p_belt.go`. The single entry function is `renderIISS4PBelt`, matching the existing `renderIISS4P*` helper naming in `iiss_class4p.go`. (The exported dispatcher remains `RenderIISSClass4P`.)

### Dispatch

`RenderIISSClass4P` (existing) keys on `body.SizeCode == "0"`:

```go
if body.SizeCode == "0" {
    return renderIISS4PBelt(body, sys, mainworldDesignation)
}
```

`renderBeltStub` is deleted.

### Section layout

Plain text, six sections in book order (Header / Orbit / Composition / Resources / Major Bodies / Comments), matching the existing PART P style (section header line, indented field-value lines, blank line between sections):

```text
IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B

WORLD: <designation>   SAH/UWP: 000
SECTOR | LOCATION:    INITIAL SURVEY:    LAST UPDATED:
PRIMARY OBJECT(S): <group>    SYSTEM AGE (Gyr): <age>    TRAVEL ZONE:

ORBIT
  O#: <orbit>   AU: <au>   Span: <span> Orbit#s   Period (h): <period>

COMPOSITION
  m-type%: <m>   s-type%: <s>   c-type%: <c>   other%: <other>
  Bulk: <bulk>
  Major Bodies: Size 1 = <count1>   Size S = <countS>

RESOURCES
  Rating: <rating>

MAJOR BODIES
  Counts only: <count1> size-1 + <countS> size-S; per-body detail not generated.

COMMENTS
  [This is the system mainworld.   ← only when mainworldDesignation matches]
```

### Field sourcing

| Field                                                       | Source                                                               |
| ----------------------------------------------------------- | -------------------------------------------------------------------- |
| WORLD                                                       | `body.Designation`                                                   |
| SAH/UWP                                                     | literal `"000"` (size 0, atm 0, hydro 0 — book convention for belts) |
| Sector\|Location, Initial Survey, Last Updated, Travel Zone | empty placeholders (caller-supplied metadata, identical to PART P)   |
| Primary Object(s)                                           | `body.Group.Designation` (e.g., "Aab")                               |
| System Age (Gyr)                                            | `sys.Primary.AgeGyr`                                                 |
| O#                                                          | `body.Orbit`                                                         |
| AU                                                          | `stars.OrbitToAU(body.Orbit)`                                        |
| Span                                                        | `body.Belt.Span` (Orbit#s, decimal)                                  |
| Period (h)                                                  | `body.Period.Hours` (matches existing PART P convention)             |
| m/s/c/other %                                               | `body.Belt.Composition.{MTypePct,STypePct,CTypePct,OtherPct}`        |
| Bulk                                                        | `body.Belt.Bulk`                                                     |
| Major Bodies counts                                         | `body.Belt.SigSize1Bodies`, `body.Belt.SigSizeSBodies`               |
| Resource Rating                                             | `body.Belt.ResourceRating`                                           |
| Mainworld marker                                            | `mainworldDesignation == body.Designation`                           |

### Nil-`*BeltDetails` handling

If `body.Belt == nil` (shouldn't happen for `BodyPlanetoidBelt`, but the renderer is defensive), the COMPOSITION, RESOURCES, and MAJOR BODIES sections render as `(belt details not generated)` and the renderer returns rather than panicking. This is consistent with how the existing PART P helpers handle nil `*Atmosphere`, `*Geology`, etc.

### Test additions

New file: `worlds/iiss_class4p_belt_test.go`.

- `TestRenderIISS4PBelt_PopulatedFields` — synth a `DetailedPlacement` with all fields populated, render, assert each labeled value appears with the right number.
- `TestRenderIISS4PBelt_NilBeltDetails_DegradesGracefully` — synth without `*BeltDetails`, render, assert no panic and the placeholder strings appear.
- `TestRenderIISS4PBelt_MainworldMarker` — render with `mainworldDesignation` matching/not matching, assert COMMENTS line appears/absent.

`iiss_class4p_test.go`:

- Update `TestRenderIISSClass4P_Belt_StubRendering` (or rename to `_DispatchesToBeltRenderer`) — assert the rendered string starts with `"IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B"` (no longer the "NOT YET IMPLEMENTED" stub) and contains the COMPOSITION header.

### What does NOT need to change

- `BeltDetails` struct (`worlds/belt_details.go`) — every field the form needs is already there.
- `runDetailPipeline` and the per-belt generation chain — no upstream data changes.
- `RenderIISSClass23` (the system-level form) — the belt entry in the Object table is unchanged.
- Profile rendering (`FormatBeltProfile`) — separate concern.

## Testing strategy

The book provides no PART P.B worked example (the only Class IV-P examples on pp. 141–144 are PART P, terrestrials/moons), so this work has no fidelity test in the worked-example sense. The implementation is verified by:

1. Field-population tests on synthetic placements (described above).
2. Selection tests on synthetic systems (described above).
3. The existing acceptance test `TestZed_FullDetail_3B-final` continues to pass — Zed Prime is a moon, not a belt, so its mainworld pick is unchanged.

## Success criteria

- A system whose only `BodyPlanetoidBelt` is the highest-resource non-empty placement returns that belt's designation from `pickMainworld`.
- `RenderIISSClass4P` invoked on that belt returns a string starting with `"IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B"` and containing every section listed above.
- Synthesizing a `DetailedPlacement` with a populated `*BeltDetails` and rendering produces values matching the field-sourcing table above.
- `just check && just test` clean.
- The renderBeltStub function and its "NOT YET IMPLEMENTED" string are gone.

## Carry-forwards

These do not block this sub-project but become reachable once it ships:

- Markdown formatter for PART P.B (next sub-project, B in the three-unit plan).
- Per-body Major Bodies subtable detail (would require generating individual belt-member bodies, a chapter-of-its-own task).
- Referee mainworld override (deferred item (q)).
- Item (s) is closed by this sub-project's mainworld extension.
