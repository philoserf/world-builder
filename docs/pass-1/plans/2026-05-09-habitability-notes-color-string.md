# Habitability Notes Referee-Color String Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Populate `Habitability.Notes` with book-quoted descriptions of which WBH p.132 DM rules fired, joined with `"; "`, and surface it through both Class IV-P render paths.

**Architecture:** Refactor each `habitability*DM` helper (in `worlds/habitability.go`) to return both the DM and its description string (or `[]string` for the multi-condition temperature helper). `ComputeHabitability` collects non-empty notes in book order and joins them. Both renderers (`worlds/markdown.go::writeClass4PHabitability` and `worlds/iiss_class4p.go::renderIISS4PHabitability`) gain a conditional Notes row/line.

**Tech Stack:** Go 1.26, `task` (gofumpt + go vet + golangci-lint + modernizer), `go test -race`.

---

## File Structure

- **Modify:** `worlds/habitability.go` — refactor helpers' signatures; populate `Notes` in `ComputeHabitability`; add `"strings"` import; update field doc-comment.
- **Modify:** `worlds/habitability_test.go` — add `TestComputeHabitability_Notes` integration test.
- **Modify:** `worlds/markdown.go` — conditional Notes row in `writeClass4PHabitability`.
- **Modify:** `worlds/iiss_class4p.go` — conditional Notes line in `renderIISS4PHabitability`.
- **Modify:** `worlds/iiss_class4p_test.go` — extend existing habitability test; add Notes-present and Notes-empty cases.
- **Modify:** `worlds/testdata/zed_markdown.golden` — refresh; expect Aab IV (Zed Prime) Habitability section to gain a Notes row.

---

### Task 1: Refactor helpers + populate `Notes` (TDD)

**Files:**

- Modify: `worlds/habitability_test.go` (add `TestComputeHabitability_Notes`)
- Modify: `worlds/habitability.go` (refactor 6 helpers + `ComputeHabitability`; add `strings` import; update field doc-comment)

- [ ] **Step 1: Write the failing integration test**

Open `worlds/habitability_test.go` and append (after the last test in the file):

```go
func TestComputeHabitability_Notes(t *testing.T) {
	cases := []struct {
		name      string
		setup     func() *DetailedPlacement
		wantNotes string
	}{
		{
			name: "Terra baseline (no DMs fire)",
			setup: func() *DetailedPlacement {
				body := &DetailedPlacement{}
				body.SizeCode = "8"
				body.Atmosphere = &Atmosphere{Code: 6}
				body.Hydrographics = &Hydrographics{Code: 7}
				body.Temperature = &Temperature{HighK: 310, MeanK: 290, LowK: 270}
				body.Physical = &BodyPhysical{Gravity: 1.0}
				return body
			},
			wantNotes: "",
		},
		{
			name: "Zed Prime (HighK > 323 + low gravity)",
			setup: func() *DetailedPlacement {
				body := &DetailedPlacement{}
				body.SizeCode = "5"
				body.Atmosphere = &Atmosphere{Code: 6}
				body.Hydrographics = &Hydrographics{Code: 5}
				body.Temperature = &Temperature{HighK: 346, MeanK: 290, LowK: 270}
				body.Physical = &BodyPhysical{Gravity: 0.66}
				return body
			},
			wantNotes: "Too hot at times; Low gravity",
		},
		{
			name: "Hostile (atm B + tidal lock 1:1)",
			setup: func() *DetailedPlacement {
				body := &DetailedPlacement{}
				body.SizeCode = "8"
				body.Atmosphere = &Atmosphere{Code: 11}
				body.Hydrographics = &Hydrographics{Code: 7}
				body.TidalLock = &TidalLock{
					Case:           TidalLockCasePlanetToStar,
					LockRatio:      "1:1",
					IsTwilightZone: true,
				}
				body.Temperature = &Temperature{HighK: 310, MeanK: 290, LowK: 270}
				body.Physical = &BodyPhysical{Gravity: 1.0}
				return body
			},
			wantNotes: "Hostile Atmosphere; Very little useable land surface area",
		},
		{
			name: "Multi-temp (HighK and MeanK both > 323)",
			setup: func() *DetailedPlacement {
				body := &DetailedPlacement{}
				body.SizeCode = "8"
				body.Atmosphere = &Atmosphere{Code: 6}
				body.Hydrographics = &Hydrographics{Code: 7}
				body.Temperature = &Temperature{HighK: 350, MeanK: 340, LowK: 250}
				body.Physical = &BodyPhysical{Gravity: 1.0}
				return body
			},
			wantNotes: "Too hot at times; Too hot most of the time",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ComputeHabitability(c.setup())
			if got.Notes != c.wantNotes {
				t.Errorf("got Notes %q, want %q", got.Notes, c.wantNotes)
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./worlds/ -run TestComputeHabitability_Notes -v
```

Expected: FAIL — every subcase except "Terra baseline" fails because `ComputeHabitability` currently leaves `Notes == ""`. Terra baseline coincidentally passes (already empty).

- [ ] **Step 3: Refactor `habitabilitySizeDM`**

Open `worlds/habitability.go`. Replace the existing function:

```go
// habitabilitySizeDM per WBH p.132 size-DM table.
func habitabilitySizeDM(size int) int {
	switch {
	case size <= 4:
		return -1
	case size >= 9:
		return +1
	}
	return 0
}
```

with:

```go
// habitabilitySizeDM per WBH p.132 size-DM table.
// Returns the DM and the book's description-column phrase (empty when no
// DM fires, i.e., size 5–8).
func habitabilitySizeDM(size int) (int, string) {
	switch {
	case size <= 4:
		return -1, "Limited surface area"
	case size >= 9:
		return +1, "Additional surface area"
	}
	return 0, ""
}
```

- [ ] **Step 4: Refactor `habitabilityAtmDM`**

Replace:

```go
// habitabilityAtmDM per WBH p.132 atmosphere-DM table.
// nil Atmosphere is treated as atm code 0 (vacuum) → DM-8.
func habitabilityAtmDM(body *DetailedPlacement) int {
	atmCode := 0
	if body.Atmosphere != nil {
		atmCode = body.Atmosphere.Code
	}
	switch atmCode {
	case 0, 1, 10: // 0, 1, A
		return -8
	case 2, 14: // 2, E
		return -4
	case 3, 13: // 3, D
		return -3
	case 4, 9:
		return -2
	case 5, 7, 8:
		return -1
	case 6:
		return 0 // baseline
	case 11: // B
		return -10
	case 12, 15: // C, F+
		return -12
	}
	return 0
}
```

with:

```go
// habitabilityAtmDM per WBH p.132 atmosphere-DM table.
// nil Atmosphere is treated as atm code 0 (vacuum) → DM-8.
// Returns the DM and the book's description-column phrase (empty when
// no DM fires, i.e., atm 6 baseline or unhandled codes).
func habitabilityAtmDM(body *DetailedPlacement) (int, string) {
	atmCode := 0
	if body.Atmosphere != nil {
		atmCode = body.Atmosphere.Code
	}
	switch atmCode {
	case 0, 1, 10: // 0, 1, A
		return -8, "Non-breathable atmosphere"
	case 2, 14: // 2, E
		return -4, "Very thin, tainted, or thin, low atmospheres"
	case 3, 13: // 3, D
		return -3, "Very thin or very dense atmosphere"
	case 4, 9:
		return -2, "Tainted thin or dense atmospheres"
	case 5, 7, 8:
		return -1, "Thin, taint (standard), or dense Atmospheres"
	case 6:
		return 0, "" // baseline
	case 11: // B
		return -10, "Hostile Atmosphere"
	case 12, 15: // C, F+
		return -12, "Very hostile Atmosphere"
	}
	return 0, ""
}
```

- [ ] **Step 5: Refactor `habitabilityHydroDM`**

Replace:

```go
// habitabilityHydroDM per WBH p.132 hydrographics-DM table.
// nil Hydrographics is treated as Hydro code 0 → DM-4.
func habitabilityHydroDM(body *DetailedPlacement) int {
	hydroCode := 0
	if body.Hydrographics != nil {
		hydroCode = body.Hydrographics.Code
	}
	switch {
	case hydroCode == 0:
		return -4
	case hydroCode >= 1 && hydroCode <= 3:
		return -2
	case hydroCode == 9:
		return -1
	case hydroCode >= 10:
		return -2
	}
	return 0 // 4-8
}
```

with:

```go
// habitabilityHydroDM per WBH p.132 hydrographics-DM table.
// nil Hydrographics is treated as Hydro code 0 → DM-4.
// Returns the DM and the book's description-column phrase (empty when
// no DM fires, i.e., Hydro 4–8).
func habitabilityHydroDM(body *DetailedPlacement) (int, string) {
	hydroCode := 0
	if body.Hydrographics != nil {
		hydroCode = body.Hydrographics.Code
	}
	switch {
	case hydroCode == 0:
		return -4, "Lack of accessible liquid water"
	case hydroCode >= 1 && hydroCode <= 3:
		return -2, "Desert conditions prevalent"
	case hydroCode == 9:
		return -1, "Little useable land surface area"
	case hydroCode >= 10:
		return -2, "Very little useable land surface area"
	}
	return 0, "" // 4-8
}
```

- [ ] **Step 6: Refactor `habitabilityTidalLockDM`**

Replace:

```go
// habitabilityTidalLockDM per WBH p.132: "Solar tidally locked (1:1)
// world" → DM-2. Detection: TidalLock.IsTwilightZone (which is true
// only when Case == PlanetToStar AND LockRatio == "1:1").
func habitabilityTidalLockDM(body *DetailedPlacement) int {
	if body.TidalLock == nil {
		return 0
	}
	if body.TidalLock.IsTwilightZone {
		return -2
	}
	return 0
}
```

with:

```go
// habitabilityTidalLockDM per WBH p.132: "Solar tidally locked (1:1)
// world" → DM-2. Detection: TidalLock.IsTwilightZone (which is true
// only when Case == PlanetToStar AND LockRatio == "1:1").
// Returns the DM and the book's description-column phrase.
func habitabilityTidalLockDM(body *DetailedPlacement) (int, string) {
	if body.TidalLock == nil {
		return 0, ""
	}
	if body.TidalLock.IsTwilightZone {
		return -2, "Very little useable land surface area"
	}
	return 0, ""
}
```

- [ ] **Step 7: Refactor `habitabilityTempDM`**

Replace:

```go
// habitabilityTempDM per WBH p.132 temperature-DM table.
// Returns 0 when Temperature is nil (defensive).
//
// Note: HighK > 323 and MeanK > 323 are strict (323 itself is in the
// [304, 323] band → -2, NOT in the >323 band → -4). Per WBH p.132 footnote,
// "use worst at edges" — but the bands as written are unambiguous at 323.
func habitabilityTempDM(body *DetailedPlacement) int {
	if body.Temperature == nil {
		return 0
	}
	dm := 0
	t := body.Temperature
	if t.HighK > 323 {
		dm += -2
	}
	if t.HighK > 0 && t.HighK < 279 {
		dm += -2
	}
	if t.MeanK > 323 {
		dm += -4
	} else if t.MeanK >= 304 && t.MeanK <= 323 {
		dm += -2
	}
	if t.MeanK > 0 && t.MeanK < 273 {
		dm += -2
	}
	if t.LowK > 0 && t.LowK < 200 {
		dm += -2
	}
	return dm
}
```

with:

```go
// habitabilityTempDM per WBH p.132 temperature-DM table. Multiple
// sub-conditions (HighK, MeanK bands, LowK) can fire independently,
// so this helper returns a slice of fired-condition descriptions.
// Returns (0, nil) when Temperature is nil (defensive).
//
// Note: HighK > 323 and MeanK > 323 are strict (323 itself is in the
// [304, 323] band → -2, NOT in the >323 band → -4). Per WBH p.132 footnote,
// "use worst at edges" — but the bands as written are unambiguous at 323.
func habitabilityTempDM(body *DetailedPlacement) (int, []string) {
	if body.Temperature == nil {
		return 0, nil
	}
	dm := 0
	var notes []string
	t := body.Temperature
	if t.HighK > 323 {
		dm += -2
		notes = append(notes, "Too hot at times")
	}
	if t.HighK > 0 && t.HighK < 279 {
		dm += -2
		notes = append(notes, "Too cold all of the time")
	}
	if t.MeanK > 323 {
		dm += -4
		notes = append(notes, "Too hot most of the time")
	} else if t.MeanK >= 304 && t.MeanK <= 323 {
		dm += -2
		notes = append(notes, "Too hot most of the time")
	}
	if t.MeanK > 0 && t.MeanK < 273 {
		dm += -2
		notes = append(notes, "Too cold most of the time")
	}
	if t.LowK > 0 && t.LowK < 200 {
		dm += -2
		notes = append(notes, "Much too cold some of the time")
	}
	return dm, notes
}
```

- [ ] **Step 8: Refactor `habitabilityGravityDM`**

Replace:

```go
// habitabilityGravityDM per WBH p.132 gravity-DM table.
//
// WBH p.132 has overlapping bands (0.2-0.7 and 0.4-0.7). Per the worked
// example for Zed Prime (gravity 0.66 → DM-1, NOT -2), the narrower band
// wins. Documented as a WBH inconsistency (footnote contradicts worked
// example); implementation follows the worked example.
//
// Undefined gravity (Physical nil): per WBH "+1 - |6 - Size|".
func habitabilityGravityDM(body *DetailedPlacement) int {
	if body.Physical == nil {
		size := SizeAsInt(body.SizeCode)
		diff := 6 - size
		if diff < 0 {
			diff = -diff
		}
		return 1 - diff
	}
	g := body.Physical.Gravity
	switch {
	case g < 0.2:
		return -4
	case g >= 0.7 && g <= 0.9:
		return +1
	case g >= 0.4 && g < 0.7:
		return -1 // narrower band; wins over 0.2-0.7 per Q3-a
	case g >= 0.2 && g < 0.4:
		return -2 // residual of 0.2-0.7
	case g > 1.1 && g <= 1.4:
		return -1
	case g > 1.4 && g <= 2.0:
		return -3
	case g > 2.0:
		return -6
	}
	return 0 // 0.9-1.1 (Earth-like baseline)
}
```

with:

```go
// habitabilityGravityDM per WBH p.132 gravity-DM table.
//
// WBH p.132 has overlapping bands (0.2-0.7 and 0.4-0.7). Per the worked
// example for Zed Prime (gravity 0.66 → DM-1, NOT -2), the narrower band
// wins. Documented as a WBH inconsistency (footnote contradicts worked
// example); implementation follows the worked example.
//
// Undefined gravity (Physical nil): per WBH "+1 - |6 - Size|" — the
// book gives this fallback formula but no description column entry,
// so the note is empty.
//
// Returns the DM and the book's description-column phrase.
func habitabilityGravityDM(body *DetailedPlacement) (int, string) {
	if body.Physical == nil {
		size := SizeAsInt(body.SizeCode)
		diff := 6 - size
		if diff < 0 {
			diff = -diff
		}
		return 1 - diff, "" // no book description for the fallback formula
	}
	g := body.Physical.Gravity
	switch {
	case g < 0.2:
		return -4, "Unhealthy low gravity levels"
	case g >= 0.7 && g <= 0.9:
		return +1, "Gravity very comfortable"
	case g >= 0.4 && g < 0.7:
		return -1, "Low gravity" // narrower band; wins over 0.2-0.7 per Q3-a
	case g >= 0.2 && g < 0.4:
		return -2, "Very low gravity" // residual of 0.2-0.7
	case g > 1.1 && g <= 1.4:
		return -1, "Gravity somewhat high"
	case g > 1.4 && g <= 2.0:
		return -3, "Gravity uncomfortably high"
	case g > 2.0:
		return -6, "Gravity too high for acclimation"
	}
	return 0, "" // 0.9-1.1 (Earth-like baseline)
}
```

- [ ] **Step 9: Update `ComputeHabitability` to collect Notes**

Add a `"strings"` import to the file (currently has none). The `package worlds` declaration is on line 3; add the import block right after:

```go
package worlds

import "strings"

// Habitability — a per-body habitability rating for Terragens per WBH
```

(Verify: the file currently has no imports. If a Read shows otherwise, add `strings` to the existing import block.)

Replace the current `ComputeHabitability` body:

```go
func ComputeHabitability(body *DetailedPlacement) Habitability {
	if body == nil {
		return Habitability{Rating: 0}
	}
	dm := habitabilitySizeDM(SizeAsInt(body.SizeCode))
	dm += habitabilityAtmDM(body)
	dm += habitabilityHydroDM(body)
	dm += habitabilityTidalLockDM(body)
	dm += habitabilityTempDM(body)
	dm += habitabilityGravityDM(body)

	rating := min(max(10+dm, 0), 12)
	return Habitability{Rating: rating}
}
```

with:

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

	dm := sizeDM + atmDM + hydroDM + tidalDM + tempDM + gravDM
	rating := min(max(10+dm, 0), 12)
	return Habitability{Rating: rating, Notes: strings.Join(notes, "; ")}
}
```

- [ ] **Step 10: Update the `Notes` field doc-comment**

In the same file, find the field doc-comment (around lines 23-26):

```go
	// Notes is a referee-color string visible in the Class IV-P form's
	// Habitability section (e.g., "High temperatures hinder habitability").
	// Currently always empty — populated by future referee-feature carry-forward.
	Notes string
```

Replace with:

```go
	// Notes is a referee-color string visible in the Class IV-P form's
	// Habitability section. Populated by ComputeHabitability from the
	// WBH p.132 DM table's description column for whichever rules fired,
	// joined with "; ". Empty when no DMs fire (Terra-equivalent baseline).
	Notes string
```

- [ ] **Step 11: Run the test to verify it passes**

```bash
go test ./worlds/ -run TestComputeHabitability_Notes -v
```

Expected: PASS — all four subcases.

- [ ] **Step 12: Run full habitability test set**

```bash
go test ./worlds/ -run TestComputeHabitability -v
```

Expected: PASS — all existing tests still green (none assert on `Notes`, so the field-population change is invisible to them).

- [ ] **Step 13: Stage and run quality gate**

```bash
git add worlds/habitability.go worlds/habitability_test.go
task check
```

Expected: clean.

- [ ] **Step 14: Commit**

```bash
git commit -m "$(cat <<'EOF'
feat(worlds): populate Habitability.Notes from WBH p.132 descriptions

Each habitability*DM helper now returns both the DM value and its
book-quoted description-column phrase. ComputeHabitability collects
non-empty descriptions in book order (Size → Atm → Hydro → TidalLock
→ Temp → Gravity) and joins them with "; " into Habitability.Notes.

Empty when no DMs fire (Terra-equivalent baseline). Verbatim text from
WBH p.132 — e.g., "Hostile Atmosphere", "Too hot at times", "Low gravity".

The next commit wires this into the Class IV-P render paths (markdown
+ IISS).

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Render `Notes` in Class IV-P paths (TDD)

**Files:**

- Modify: `worlds/iiss_class4p_test.go` (extend habitability test, add Notes-present + Notes-empty cases)
- Modify: `worlds/iiss_class4p.go` (conditional Notes line)
- Modify: `worlds/markdown.go` (conditional Notes row)
- Possibly modify: `worlds/testdata/zed_markdown.golden` (refresh)

- [ ] **Step 1: Write the failing IISS Class 4P tests**

Open `worlds/iiss_class4p_test.go`. Find the existing test:

```go
func TestRenderIISSClass4P_HabitabilitySection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Habitability = &Habitability{Rating: 7}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "HABITABILITY") {
		t.Errorf("missing HABITABILITY section: got %q", got)
	}
	if !strings.Contains(got, "Rating: 7") {
		t.Errorf("missing Rating: 7: got %q", got)
	}
}
```

Replace with:

```go
func TestRenderIISSClass4P_HabitabilitySection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Habitability = &Habitability{Rating: 7}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "HABITABILITY") {
		t.Errorf("missing HABITABILITY section: got %q", got)
	}
	if !strings.Contains(got, "Rating: 7") {
		t.Errorf("missing Rating: 7: got %q", got)
	}
	// Empty Notes: no Notes line.
	if strings.Contains(got, "Notes:") {
		t.Errorf("got Notes: line for empty Notes: %q", got)
	}
}

func TestRenderIISSClass4P_HabitabilitySection_WithNotes(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Habitability = &Habitability{
		Rating: 7,
		Notes:  "Too hot at times; Low gravity",
	}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "Rating: 7") {
		t.Errorf("missing Rating: 7: got %q", got)
	}
	if !strings.Contains(got, "Notes:  Too hot at times; Low gravity") {
		t.Errorf("missing Notes line: got %q", got)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
go test ./worlds/ -run TestRenderIISSClass4P_HabitabilitySection -v
```

Expected: `TestRenderIISSClass4P_HabitabilitySection_WithNotes` FAILS — the renderer doesn't emit a `Notes:` line yet. The empty-Notes case in the existing test passes (renderer still doesn't emit one).

- [ ] **Step 3: Update `renderIISS4PHabitability`**

Open `worlds/iiss_class4p.go`. Find the function:

```go
func renderIISS4PHabitability(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("HABITABILITY\n")
	if body.Habitability == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	fmt.Fprintf(sb, "  Rating: %d\n\n", body.Habitability.Rating)
}
```

Replace with:

```go
func renderIISS4PHabitability(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("HABITABILITY\n")
	if body.Habitability == nil {
		sb.WriteString("  (not computed)\n\n")
		return
	}
	fmt.Fprintf(sb, "  Rating: %d\n", body.Habitability.Rating)
	if body.Habitability.Notes != "" {
		fmt.Fprintf(sb, "  Notes:  %s\n", body.Habitability.Notes)
	}
	sb.WriteString("\n")
}
```

- [ ] **Step 4: Update `writeClass4PHabitability`**

Open `worlds/markdown.go`. Find the function:

```go
func writeClass4PHabitability(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Habitability\n\n")
	if body.Habitability == nil {
		writeNotGenerated(sb)
		return
	}
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Rating | %d — %s |\n\n",
		body.Habitability.Rating, HabitabilityRatingName(body.Habitability.Rating))
}
```

Replace with:

```go
func writeClass4PHabitability(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Habitability\n\n")
	if body.Habitability == nil {
		writeNotGenerated(sb)
		return
	}
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Rating | %d — %s |\n",
		body.Habitability.Rating, HabitabilityRatingName(body.Habitability.Rating))
	if body.Habitability.Notes != "" {
		fmt.Fprintf(sb, "| Notes | %s |\n", body.Habitability.Notes)
	}
	sb.WriteString("\n")
}
```

- [ ] **Step 5: Run the IISS tests to verify they pass**

```bash
go test ./worlds/ -run TestRenderIISSClass4P_HabitabilitySection -v
```

Expected: PASS — both the existing and the new test.

- [ ] **Step 6: Stage Go changes**

```bash
git add worlds/iiss_class4p.go worlds/iiss_class4p_test.go worlds/markdown.go
```

- [ ] **Step 7: Run the full test suite**

```bash
go test -race ./...
```

Expected: PASS, except `TestRenderSystemMarkdown_ZedGolden` will fail because Aab IV (Zed Prime) now has a Notes row in the markdown output. If that's the only failure, proceed to Step 8. If anything else fails, escalate (BLOCKED).

- [ ] **Step 8: Refresh Zed golden**

```bash
go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -update
git diff worlds/testdata/zed_markdown.golden
```

Verify the diff is limited to the Aab IV body's Habitability section: a new `| Notes | ... |` row added between the `| Rating | ... |` row and the next blank line. Also possibly Notes rows on other bodies that have Habitability with non-empty Notes (any HZ-orbit terrestrial whose ComputeHabitability fired DMs).

Anything outside that scope is a regression — escalate (BLOCKED). If clean:

```bash
git add worlds/testdata/zed_markdown.golden
go test -race ./...
```

Confirm green.

- [ ] **Step 9: Run task quality gate**

```bash
task check
```

Expected: clean.

- [ ] **Step 10: Commit**

```bash
git commit -m "$(cat <<'EOF'
fix(worlds): render Habitability.Notes in Class IV-P paths (closes #15)

Both render paths (markdown writeClass4PHabitability + IISS
renderIISS4PHabitability) emit a Notes row/line below Rating when
Habitability.Notes is non-empty, and omit it otherwise.

Markdown row: "| Notes | ... |"
IISS line:    "  Notes:  ..."

Zed golden refreshed: Aab IV (Zed Prime) gains a Notes row reflecting
the DM descriptions that fire on its high-temperature, low-gravity
profile per WBH p.133 worked example.

Closes #15.

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: PR + close-out

**Files:** none (operational steps only).

- [ ] **Step 1: Push the branch**

```bash
git push -u origin feat/habitability-notes-color-string
```

- [ ] **Step 2: Open the PR**

```bash
gh pr create --repo philoserf/world-builder --title "fix(worlds): habitability notes referee-color string (closes #15)" --body "$(cat <<'EOF'
## Summary

- Each `habitability*DM` helper now returns both the DM and its book-quoted description-column phrase per WBH p.132.
- `ComputeHabitability` collects non-empty descriptions in book order (Size → Atm → Hydro → TidalLock → Temp → Gravity), joins with `"; "`, and stores in `Habitability.Notes`.
- Both Class IV-P render paths (markdown + IISS) emit a Notes row/line below Rating when `Notes != ""`, and omit it when empty.
- Zed golden refreshed: Aab IV (Zed Prime) gains a Notes row reflecting its high-temperature + low-gravity profile.

Closes #15.

## Spec / plan

- Spec: `docs/pass-1/specs/2026-05-09-habitability-notes-color-string-design.md`
- Plan: `docs/pass-1/plans/2026-05-09-habitability-notes-color-string.md`

## Test plan

- [x] `task check` clean (gofumpt, vet, golangci-lint, modernizer)
- [x] `task test` clean with race detector
- [x] New `TestComputeHabitability_Notes` — 4 cases: Terra baseline (empty), Zed-Prime-like, hostile (atm B + tidal lock), multi-temp
- [x] Updated `TestRenderIISSClass4P_HabitabilitySection_Present` — asserts no Notes line for empty Notes
- [x] New `TestRenderIISSClass4P_HabitabilitySection_WithNotes` — asserts Notes line is present when populated
- [x] Existing `TestComputeHabitability_*` tests unchanged (none assert on Notes; field population is invisible to them)
- [x] Zed golden refreshed (RNG-stable; only the new Notes row is added)

## Out of scope

- Atmosphere taint DMs (low-oxygen-taint DM-2 from WBH p.132) — still skipped per the existing `ComputeHabitability` comment about Q3-a.
- WBH p.133 miscellaneous scoring (D3-1 referee-elective adjustment) — separate feature.
- Other renderers (`short` / `json` formats) — only Class IV-P paths read `Habitability.Notes`.
EOF
)"
```

- [ ] **Step 3: Stop**

Implementation complete on the branch; PR is open. Hand back to the user for review/merge.

---

## Self-review

**Spec coverage**

- Spec § Notes content (book-quoted descriptions, joined with "; ", empty for baseline): Task 1 Step 1 (test cases) + Step 9 (orchestrator). ✓
- Spec § Source-of-truth coupling (note text and DM in same helper): Task 1 Steps 3-8. ✓
- Spec § Helper signatures: Task 1 Steps 3-8. ✓
- Spec § ComputeHabitability orchestration: Task 1 Step 9. ✓
- Spec § Renderer format (Markdown + IISS): Task 2 Steps 3-4. ✓
- Spec § Out of scope (taint DMs, miscellaneous scoring): not touched (correct). ✓
- Spec § Testing strategy ComputeHabitability_Notes integration test: Task 1 Step 1. ✓
- Spec § Testing strategy renderer tests: Task 2 Step 1. ✓
- Spec § Zed golden refresh: Task 2 Step 8. ✓

**Placeholder scan**

No "TBD" / "TODO" / "implement later". All steps include concrete code or commands.

**Type consistency**

- Helper signatures match across helper definition (Steps 3-8) and call site (Step 9). Most are `(int, string)`; temperature is `(int, []string)`.
- `Habitability{Rating: ..., Notes: ...}` literal shape consistent.
- Tests construct `Habitability{Rating, Notes}` matching the existing struct.
