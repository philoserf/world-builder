# IISS Class IV-P PART P.B Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the Class IV-P renderer so it covers belt-mainworld systems, and extend `pickMainworld` so a belt can actually be selected as the system mainworld.

**Architecture:** Two surgical changes to existing files plus one new file. (1) `worlds/mainworld.go`'s `collect` admits `BodyPlanetoidBelt` and reads `ResourceRating` from `*BeltDetails` instead of `*Biology`. (2) New file `worlds/iiss_class4p_belt.go` contains a single `renderIISS4PBelt` function that mirrors the existing `renderIISS4P*` helper style and emits the WBH p.139 form. (3) `worlds/iiss_class4p.go`'s dispatch on `SizeCode == "0"` routes to the new renderer; the placeholder `renderBeltStub` is deleted.

**Tech Stack:** Go 1.26, `strings.Builder` for rendering, `go test -race ./...` via `just test`, `gofumpt` formatting.

**Spec:** `docs/history/pass-1-specs/2026-05-06-iiss-class4p-belt-design.md`

---

## File Map

| File                               | Action | What it holds                                                             |
| ---------------------------------- | ------ | ------------------------------------------------------------------------- |
| `worlds/mainworld.go`              | modify | `collect` admits `BodyPlanetoidBelt`; reads `*BeltDetails.ResourceRating` |
| `worlds/mainworld_test.go`         | modify | delete `TestPickMainworld_BeltsAndGGsOnly_EmptyString`; add 3 new tests   |
| `worlds/iiss_class4p_belt.go`      | create | `renderIISS4PBelt(body, sys, mainworldDesignation) string`                |
| `worlds/iiss_class4p_belt_test.go` | create | 3 new field-population / nil-defense / marker tests                       |
| `worlds/iiss_class4p.go`           | modify | dispatch `SizeCode == "0"` to `renderIISS4PBelt`; delete `renderBeltStub` |
| `worlds/iiss_class4p_test.go`      | modify | rename + update existing belt stub test                                   |

---

## Task 1: Extend `pickMainworld` to admit belts

**Files:**

- Modify: `worlds/mainworld.go` (function `pickMainworld`, lines 103-191)
- Modify: `worlds/mainworld_test.go` (delete lines 212-222; add three new tests at end of file)

- [ ] **Step 1.1: Delete the obsolete `BeltsAndGGsOnly_EmptyString` test**

This test asserts the _old_ behavior that belts cannot be mainworlds. The new tests in step 1.2 cover the same surface area more thoroughly.

In `worlds/mainworld_test.go`, delete lines 212-222 (the entire `TestPickMainworld_BeltsAndGGsOnly_EmptyString` function and the blank line after it).

- [ ] **Step 1.2: Add three failing tests for the new selection behavior**

Append to `worlds/mainworld_test.go`:

```go
func TestPickMainworld_BeltOnlySystem_ReturnsBeltWithHighestResource(t *testing.T) {
	belt1 := DetailedPlacement{Designation: "Belt A", SizeCode: "0", Belt: &BeltDetails{ResourceRating: 6}}
	belt1.Body = BodyPlanetoidBelt
	belt2 := DetailedPlacement{Designation: "Belt B", SizeCode: "0", Belt: &BeltDetails{ResourceRating: 9}}
	belt2.Body = BodyPlanetoidBelt
	gg := DetailedPlacement{Designation: "GG"}
	gg.Body = BodyGasGiant
	detailed := []DetailedPlacement{belt1, belt2, gg}
	got := pickMainworld(detailed)
	if got != "Belt B" {
		t.Errorf("got %q, want Belt B (highest resource)", got)
	}
}

func TestPickMainworld_TerrestrialBeatsBelt_OnHabitability(t *testing.T) {
	terr := DetailedPlacement{
		Designation:  "Terr",
		Habitability: &Habitability{Rating: 1},
	}
	terr.Body = BodyTerrestrial
	belt := DetailedPlacement{Designation: "Belt", SizeCode: "0", Belt: &BeltDetails{ResourceRating: 12}}
	belt.Body = BodyPlanetoidBelt
	detailed := []DetailedPlacement{terr, belt}
	got := pickMainworld(detailed)
	if got != "Terr" {
		t.Errorf("got %q, want Terr (habitability beats belt resource)", got)
	}
}

func TestPickMainworld_TerrestrialAndBelt_TieOnResource_IterationOrder(t *testing.T) {
	// No habitability on either; equal resource → first iteration wins.
	terr := DetailedPlacement{
		Designation:  "Terr",
		Habitability: &Habitability{Rating: 0},
		Biology:      &Biology{ResourceRating: 8},
	}
	terr.Body = BodyTerrestrial
	belt := DetailedPlacement{Designation: "Belt", SizeCode: "0", Belt: &BeltDetails{ResourceRating: 8}}
	belt.Body = BodyPlanetoidBelt
	detailed := []DetailedPlacement{belt, terr} // belt first
	got := pickMainworld(detailed)
	if got != "Belt" {
		t.Errorf("got %q, want Belt (first iteration order)", got)
	}
}
```

- [ ] **Step 1.3: Run the new tests and confirm they fail**

```bash
go test ./worlds/ -run 'TestPickMainworld_BeltOnly|TestPickMainworld_TerrestrialBeatsBelt|TestPickMainworld_TerrestrialAndBelt_Tie' -v
```

Expected (under the current implementation, before Task 1.4):

- `BeltOnlySystem_ReturnsBeltWithHighestResource` — **FAIL.** Returns `""` (no candidates because belt is filtered); expects `"Belt B"`.
- `TerrestrialAndBelt_TieOnResource_IterationOrder` — **FAIL.** Returns `"Terr"` (terrestrial is the only candidate); expects `"Belt"`.
- `TerrestrialBeatsBelt_OnHabitability` — **PASS** (passes by accident — terrestrial wins because the belt is filtered out, not because habitability beats resource). Keep this test anyway — after Task 1.4 it documents the priority correctly.

Two FAIL + one accidental PASS is the expected pre-impl state. Proceed.

- [ ] **Step 1.4: Modify `collect` and call sites to admit belts**

In `worlds/mainworld.go`, replace the existing `collect` closure (lines 113-126) with the version below, and update the two call sites (lines 130 and 135):

```go
collect := func(designation string, bodyType BodyType, h *Habitability, b *Biology, belt *BeltDetails) {
	if bodyType != BodyTerrestrial && bodyType != BodyPlanetoidBelt {
		return
	}
	c := candidate{designation: designation}
	if h != nil {
		c.habitability = h.Rating
	}
	if b != nil {
		c.resource = b.ResourceRating
		c.hasSophont = b.HasNativeSophont || b.HadExtinctSophont
	}
	if belt != nil {
		c.resource = belt.ResourceRating
	}
	candidates = append(candidates, c)
}

for i := range detailed {
	dp := &detailed[i]
	collect(dp.Designation, dp.Body, dp.Habitability, dp.Biology, dp.Belt)
	// Moons are treated as terrestrials for mainworld-pick purposes;
	// Zed Prime (WBH worked example) is itself a moon.
	for j := range dp.Moons {
		m := &dp.Moons[j]
		collect(m.Designation, BodyTerrestrial, m.Habitability, m.Biology, nil)
	}
}
```

Update the doc-comment on `pickMainworld` (lines 88-102) so the priority chain mentions belts:

```go
// pickMainworld returns the designation of the auto-picked mainworld per
// WBH p.134. Priority chain (first match wins):
//
//  1. Bodies with native sophonts (extant or extinct); among these,
//     highest Habitability; tiebreaker: highest ResourceRating;
//     final tiebreaker: iteration order.
//  2. Highest Habitability among non-sophont bodies; tiebreakers same.
//  3. Highest ResourceRating across terrestrials, moons, and belts if no
//     body has Habitability > 0.
//  4. First terrestrial-or-belt body in iteration order.
//
// Iterates detailed[i] (planets and belts), dp.Moons[j] (moons). Returns ""
// if no eligible body exists. Belts contribute resource only (no
// habitability, no sophonts) — they win priorities 3 or 4 when
// terrestrials don't qualify, satisfying WBH p.134's "best refuelling
// location" criterion.
```

- [ ] **Step 1.5: Run the full mainworld test suite to confirm everything passes**

```bash
go test ./worlds/ -run TestPickMainworld -v
```

Expected: all `TestPickMainworld_*` tests pass, including the three new ones and the pre-existing tests for sophont-wins, habitability-tied, all-zero-fallback, moon-as-mainworld, etc.

- [ ] **Step 1.6: Run `just check && just test` for the full project**

```bash
just check && just test
```

Expected: clean. If `just check` reports any modernizer drift (`go fix ./...` produced changes), inspect with `git diff` and stage those changes too — they're mandatory per project policy.

- [ ] **Step 1.7: Commit**

```bash
git add worlds/mainworld.go worlds/mainworld_test.go
git commit -m "$(cat <<'EOF'
feat(worlds): pickMainworld admits belts as candidates

Extends `collect` so BodyPlanetoidBelt placements enter the candidate
pool with resource = body.Belt.ResourceRating. Belts naturally lose
priorities 1 and 2 (no sophont, zero habitability) and compete on
resource rating in priorities 3 and 4. Closes deferred follow-up (s).

Spec: docs/history/pass-1-specs/2026-05-06-iiss-class4p-belt-design.md

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Implement `renderIISS4PBelt` (the new file plus all three tests)

**Files:**

- Create: `worlds/iiss_class4p_belt.go`
- Create: `worlds/iiss_class4p_belt_test.go`

- [ ] **Step 2.1: Write the three failing tests**

Create `worlds/iiss_class4p_belt_test.go`:

```go
package worlds

import (
	"strings"
	"testing"

	"wbh/stars"
)

func TestRenderIISS4PBelt_PopulatedFields(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyPlanetoidBelt
	body.Designation = "Aab PI"
	body.SizeCode = "0"
	body.Orbit = 2.7
	body.Period = Period{Hours: 5613.16}
	body.Group = Group{Designation: "Aab"}
	body.Belt = &BeltDetails{
		Span:           0.91,
		Composition:    BeltComposition{MTypePct: 25, STypePct: 55, CTypePct: 15, OtherPct: 5},
		Bulk:           4,
		ResourceRating: 8,
		SigSize1Bodies: 2,
		SigSizeSBodies: 19,
	}
	sys := stars.System{Primary: stars.Star{AgeGyr: 6.336}}

	got := renderIISS4PBelt(body, sys, "")

	expected := []string{
		"IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B",
		"WORLD: Aab PI   SAH/UWP: 000",
		"PRIMARY OBJECT(S): Aab",
		"SYSTEM AGE (Gyr): 6.336",
		"ORBIT",
		"Span: 0.910 Orbit#s",
		"Period (h): 5613.16",
		"COMPOSITION",
		"m-type%: 25   s-type%: 55   c-type%: 15   other%: 5",
		"Bulk: 4",
		"Major Bodies: Size 1 = 2   Size S = 19",
		"RESOURCES",
		"Rating: 8",
		"MAJOR BODIES",
		"Counts only: 2 size-1 + 19 size-S",
		"COMMENTS",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderIISS4PBelt_NilBeltDetails_DegradesGracefully(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyPlanetoidBelt
	body.Designation = "Empty Belt"
	body.SizeCode = "0"
	body.Group = Group{Designation: "A"}
	sys := stars.System{}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	got := renderIISS4PBelt(body, sys, "")

	expected := []string{
		"IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B",
		"WORLD: Empty Belt",
		"(belt details not generated)",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderIISS4PBelt_MainworldMarker(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyPlanetoidBelt
	body.Designation = "Aab PI"
	body.SizeCode = "0"
	body.Belt = &BeltDetails{}
	sys := stars.System{}

	// Marker present when mainworldDesignation matches.
	got := renderIISS4PBelt(body, sys, "Aab PI")
	if !strings.Contains(got, "This is the system mainworld.") {
		t.Errorf("expected mainworld marker, got:\n%s", got)
	}

	// Marker absent when designation does not match.
	got = renderIISS4PBelt(body, sys, "Other")
	if strings.Contains(got, "This is the system mainworld.") {
		t.Errorf("unexpected mainworld marker in:\n%s", got)
	}
}
```

- [ ] **Step 2.2: Run the new tests and confirm they fail**

```bash
go test ./worlds/ -run TestRenderIISS4PBelt -v
```

Expected: compile error — `renderIISS4PBelt` is undefined.

- [ ] **Step 2.3: Create the implementation**

Create `worlds/iiss_class4p_belt.go`:

```go
package worlds

import (
	"fmt"
	"strings"

	"wbh/stars"
)

// renderIISS4PBelt renders the WBH p.139 IISS Class IV Survey form
// (FORM 0407K-IV PART P.B) for a Size-0 planetoid belt body. Plain-text
// output with section headers matching the book's form layout.
//
// mainworldDesignation is the SystemDetail.MainworldDesignation; used
// only to mark whether THIS body is the mainworld in the COMMENTS section.
//
// Returns "" if body is nil. Defensive against nil *BeltDetails: the
// COMPOSITION, RESOURCES, and MAJOR BODIES sections render a placeholder
// rather than panicking. Header and ORBIT sections always render.
func renderIISS4PBelt(body *DetailedPlacement, sys stars.System, mainworldDesignation string) string {
	if body == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B\n\n")

	// Header
	fmt.Fprintf(&sb, "WORLD: %s   SAH/UWP: 000\n", body.Designation)
	sb.WriteString("SECTOR | LOCATION:    INITIAL SURVEY:    LAST UPDATED:\n")
	fmt.Fprintf(&sb, "PRIMARY OBJECT(S): %s    SYSTEM AGE (Gyr): %.3f    TRAVEL ZONE:\n\n",
		body.Group.Designation, sys.Primary.AgeGyr)

	// Orbit
	sb.WriteString("ORBIT\n")
	span := 0.0
	if body.Belt != nil {
		span = body.Belt.Span
	}
	fmt.Fprintf(&sb, "  O#: %.2f   AU: %.2f   Span: %.3f Orbit#s   Period (h): %.2f\n\n",
		body.Orbit, stars.OrbitToAU(body.Orbit), span, body.Period.Hours)

	// Composition
	sb.WriteString("COMPOSITION\n")
	if body.Belt == nil {
		sb.WriteString("  (belt details not generated)\n\n")
	} else {
		c := body.Belt.Composition
		fmt.Fprintf(&sb, "  m-type%%: %d   s-type%%: %d   c-type%%: %d   other%%: %d\n",
			c.MTypePct, c.STypePct, c.CTypePct, c.OtherPct)
		fmt.Fprintf(&sb, "  Bulk: %d\n", body.Belt.Bulk)
		fmt.Fprintf(&sb, "  Major Bodies: Size 1 = %d   Size S = %d\n\n",
			body.Belt.SigSize1Bodies, body.Belt.SigSizeSBodies)
	}

	// Resources
	sb.WriteString("RESOURCES\n")
	if body.Belt == nil {
		sb.WriteString("  (belt details not generated)\n\n")
	} else {
		fmt.Fprintf(&sb, "  Rating: %d\n\n", body.Belt.ResourceRating)
	}

	// Major Bodies (subtable summarized — per-body detail not generated; see spec non-goals)
	sb.WriteString("MAJOR BODIES\n")
	if body.Belt == nil {
		sb.WriteString("  (belt details not generated)\n\n")
	} else {
		fmt.Fprintf(&sb, "  Counts only: %d size-1 + %d size-S; per-body detail not generated.\n\n",
			body.Belt.SigSize1Bodies, body.Belt.SigSizeSBodies)
	}

	// Comments
	sb.WriteString("COMMENTS\n")
	if mainworldDesignation != "" && body.Designation == mainworldDesignation {
		sb.WriteString("  This is the system mainworld.\n")
	}

	return sb.String()
}
```

- [ ] **Step 2.4: Run the new tests and confirm they pass**

```bash
go test ./worlds/ -run TestRenderIISS4PBelt -v
```

Expected: all three PASS.

If `TestRenderIISS4PBelt_PopulatedFields` fails on a numeric-formatting mismatch (e.g., `Span: 0.910` vs `Span: 0.91`), match the assertion to whatever `%.3f` produces for `0.91` — Go's `fmt` formats `0.91` as `0.910` with `%.3f`, so the test as written should match.

- [ ] **Step 2.5: Run `just check` to confirm formatting and lint pass**

```bash
just check
```

Expected: clean. The new file goes through `gofumpt` and `golangci-lint`; if either complains, `just fmt` to auto-fix.

- [ ] **Step 2.6: Commit**

```bash
git add worlds/iiss_class4p_belt.go worlds/iiss_class4p_belt_test.go
git commit -m "$(cat <<'EOF'
feat(worlds): renderIISS4PBelt — Form 0407K-IV PART P.B

Implements the belt-mainworld variant of IISS Class IV-P per WBH p.139.
Six sections: Header / Orbit / Composition / Resources / Major Bodies /
Comments. Per-body detail in the Major Bodies subtable is summarized
as counts only — generating individual belt members is out of scope
(see spec non-goals).

Spec: docs/history/pass-1-specs/2026-05-06-iiss-class4p-belt-design.md

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Wire dispatch and remove the stub

**Files:**

- Modify: `worlds/iiss_class4p.go` (lines 27-30 dispatch; lines 51-57 `renderBeltStub`)
- Modify: `worlds/iiss_class4p_test.go` (lines 232-244 existing belt stub test)

- [ ] **Step 3.1: Update the existing belt-rendering test to assert the new behavior**

In `worlds/iiss_class4p_test.go`, replace the `TestRenderIISSClass4P_Belt_StubRendering` function (lines 232-244) with:

```go
func TestRenderIISSClass4P_Belt_DispatchesToBeltRenderer(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyPlanetoidBelt
	body.Designation = "Aab Belt"
	body.SizeCode = "0"
	body.Belt = &BeltDetails{}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "FORM 0407K-IV PART P.B") {
		t.Errorf("expected belt form header, got %q", got)
	}
	if !strings.Contains(got, "COMPOSITION") {
		t.Errorf("expected COMPOSITION section, got %q", got)
	}
	if !strings.Contains(got, "Aab Belt") {
		t.Errorf("missing belt designation: got %q", got)
	}
	if strings.Contains(got, "NOT YET IMPLEMENTED") {
		t.Errorf("stub marker still present: got %q", got)
	}
}
```

- [ ] **Step 3.2: Run the renamed test and confirm it fails**

```bash
go test ./worlds/ -run TestRenderIISSClass4P_Belt_DispatchesToBeltRenderer -v
```

Expected: FAIL — `RenderIISSClass4P` still routes belts to the stub, so the output contains `"NOT YET IMPLEMENTED"` and the test's last assertion trips.

- [ ] **Step 3.3: Wire the dispatch and delete the stub**

In `worlds/iiss_class4p.go`, change the belt-routing branch (lines 27-30):

Replace:

```go
	// Belt stub.
	if body.SizeCode == "0" {
		return renderBeltStub(body)
	}
```

With:

```go
	// Belt path — Form 0407K-IV PART P.B.
	if body.SizeCode == "0" {
		return renderIISS4PBelt(body, sys, mainworldDesignation)
	}
```

Update the doc-comment on `RenderIISSClass4P` (lines 12-23): replace the last paragraph

```go
// Belt bodies (Size 0) get a placeholder stub; full Form 0407K-IV PART P.B
// rendering is deferred (see spec carry-forwards).
```

with:

```go
// Belt bodies (Size 0) dispatch to renderIISS4PBelt for Form 0407K-IV
// PART P.B (WBH p.139).
```

Then delete the entire `renderBeltStub` function (lines 51-57 in the original file). Also remove the now-unused `"fmt"` import only if no other code in the file uses `fmt` — verify with `grep -n '\bfmt\.' worlds/iiss_class4p.go`. (Spoiler: every helper uses `fmt.Fprintf`, so the import stays.)

- [ ] **Step 3.4: Run the renamed test and confirm it passes**

```bash
go test ./worlds/ -run TestRenderIISSClass4P_Belt_DispatchesToBeltRenderer -v
```

Expected: PASS.

- [ ] **Step 3.5: Run the full project gate**

```bash
just check && just test
```

Expected: clean. Pay attention to the existing `TestZed_FullDetail_3B-final` acceptance test (in `worlds/worked_examples_test.go` or a similar file) — Zed Prime is a moon, not a belt, so its mainworld pick is unchanged. If that test fails, the priority-chain change in Task 1 broke the moon path; investigate before continuing.

- [ ] **Step 3.6: Commit**

```bash
git add worlds/iiss_class4p.go worlds/iiss_class4p_test.go
git commit -m "$(cat <<'EOF'
feat(worlds): RenderIISSClass4P dispatches belts to PART P.B renderer

Replaces renderBeltStub with renderIISS4PBelt in the SizeCode == "0"
branch. The "NOT YET IMPLEMENTED" placeholder is gone. Closes the
3B-final spec carry-forward Q5(a). Existing terrestrial/moon path
through Form 0407F-IV PART P is unchanged.

Spec: docs/history/pass-1-specs/2026-05-06-iiss-class4p-belt-design.md

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

## Final verification

- [ ] **Step F.1: Run the full project gate one more time**

```bash
just check && just test
```

Expected: clean.

- [ ] **Step F.2: Manual smoke test via the CLI**

```bash
# Find a seed that produces a belt-mainworld system (rare but possible).
# Most seeds produce terrestrial mainworlds; this is just to confirm
# the renderer doesn't crash when invoked.
go run ./cmd/wbh -seed 42 -format json | head -50
```

Expected: well-formed JSON output (the JSON path renders Class 0/I; Class IV-P routing is internal). Crashes here mean the dispatch broke.

- [ ] **Step F.3: Verify deferred items list is up to date**

The spec says this work closes deferred follow-up `(s)` and resolves the 3B-final carry-forward `Q5(a)`. After commits land, the user may want to update `~/.claude/projects/-Users-markayers-source-philoserf-world-builder/memory/MEMORY.md` to mark these closed.

This is informational; not a code change in this plan.
