# Insidious D/E Hazard Rule Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the WBH p.90 footnote where insidious subtype D or E grants an automatic Temperature hazard plus an additional rolled hazard, requiring `Atmosphere.InsidiousHazard *Hazard` to become `Atmosphere.InsidiousHazards []Hazard`.

**Architecture:** Add `RollInsidiousHazards` (plural) as a higher-level orchestrator alongside the existing `RollInsidiousHazard` (singular dice primitive). Rename the `Atmosphere` field to a slice. Update one call-site (`computeBodyTaints`), one renderer (`FormatAtmoProfileShorthand`), and the test fixtures/assertions that touched the old field.

**Tech Stack:** Go 1.26, `wbh/roller`, `task` (gofumpt + go vet + golangci-lint + modernizer), `go test -race`.

---

## File Structure

- **Modify:** `worlds/atmosphere_taint.go` — add `RollInsidiousHazards` orchestrator; remove "Not implemented" note on `RollInsidiousHazard`.
- **Modify:** `worlds/atmosphere_taint_test.go` — add unit tests for the new orchestrator.
- **Modify:** `worlds/atmosphere.go` — rename `InsidiousHazard *Hazard` field to `InsidiousHazards []Hazard`; update doc-comment.
- **Modify:** `worlds/atmosphere_profile.go` — render concatenated hazard codes after subtype dot.
- **Modify:** `worlds/atmosphere_profile_test.go` — update existing fixture; add new multi-hazard test.
- **Modify:** `worlds/system_detail_step5dprime.go` — wire `RollInsidiousHazards` into `computeBodyTaints`.
- **Modify:** `worlds/system_detail_step5dprime_test.go` — update assertions for the new field shape and the D-subtype 2-hazard outcome.
- **Possibly modify:** `worlds/testdata/zed_markdown.golden` — refresh only if seed=42 lands an atm-C body.

---

### Task 1: `RollInsidiousHazards` orchestrator + unit tests (TDD)

**Files:**

- Modify: `worlds/atmosphere_taint_test.go` (add `TestRollInsidiousHazards`)
- Modify: `worlds/atmosphere_taint.go` (add `RollInsidiousHazards`; clean up `RollInsidiousHazard` doc-comment)

- [ ] **Step 1: Write the failing orchestrator test**

Open `worlds/atmosphere_taint_test.go` and append (after `TestRollInsidiousHazard_ExtremelyDenseDM`):

```go
func TestRollInsidiousHazards(t *testing.T) {
	cases := []struct {
		name             string
		subtype          string
		isExtremelyDense bool
		twoD             int
		want             []Hazard
	}{
		// Subtype "6" (no D/E rule, no DM): rolled-B from 2D=4.
		{"subtype 6 single", "6", false, 4, []Hazard{{Code: "B"}}},
		// Subtype "C" (no D/E rule, DM+2 from extremely-dense): rolled-G from 2D=4 + DM+2 = 6 → G.
		{"subtype C single with DM", "C", true, 4, []Hazard{{Code: "G"}}},
		// Subtype "D" (D/E rule fires, DM+2): auto-T + rolled-G (2D=4 + DM+2 = 6 → G).
		{"subtype D auto-T plus rolled", "D", true, 4, []Hazard{{Code: "T"}, {Code: "G"}}},
		// Subtype "E" (D/E rule fires, DM+2): auto-T + rolled-B (2D=2 + DM+2 = 4 → B).
		{"subtype E auto-T plus rolled", "E", true, 2, []Hazard{{Code: "T"}, {Code: "B"}}},
		// Duplicate-T case: subtype "D", no DM, 2D=8 → T as rolled, plus auto-T.
		{"subtype D duplicate T", "D", false, 8, []Hazard{{Code: "T"}, {Code: "T"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := roller.NewScripted(c.twoD)
			got := RollInsidiousHazards(r, c.subtype, c.isExtremelyDense)
			if len(got) != len(c.want) {
				t.Fatalf("len: got %d, want %d (got %+v)", len(got), len(c.want), got)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("[%d]: got %+v, want %+v", i, got[i], c.want[i])
				}
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./worlds/ -run TestRollInsidiousHazards -v
```

Expected: FAIL — compilation error `undefined: RollInsidiousHazards`.

- [ ] **Step 3: Implement the orchestrator + clean the singular doc-comment**

Open `worlds/atmosphere_taint.go`. Find the existing `RollInsidiousHazard` function and its doc-comment, currently ending with this paragraph (the "Not implemented" note):

```go
// Not implemented: the WBH p.90 footnote rule that subtype D or E grants
// an automatic T hazard plus an additional rolled hazard. Representing
// two hazards requires changing Atmosphere.InsidiousHazard from a single
// pointer to a slice; see the follow-up issue tracking that work.
func RollInsidiousHazard(r roller.Roller, isExtremelyDense bool) string {
```

Replace the trailing paragraph (above `func RollInsidiousHazard`) with:

```go
// The WBH p.90 footnote rule (subtype D/E auto-T plus additional rolled
// hazard) is handled by RollInsidiousHazards (plural); this function is
// the dice primitive used by both the single-hazard path and the rolled
// half of the D/E pair.
func RollInsidiousHazard(r roller.Roller, isExtremelyDense bool) string {
```

Then directly below the existing `RollInsidiousHazard` function (after its closing `}`), insert:

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
func RollInsidiousHazards(r roller.Roller, subtype string, isExtremelyDense bool) []Hazard {
	var hazards []Hazard
	if subtype == "D" || subtype == "E" {
		hazards = append(hazards, Hazard{Code: "T"})
	}
	rolled := RollInsidiousHazard(r, isExtremelyDense)
	hazards = append(hazards, Hazard{Code: rolled})
	return hazards
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run:

```bash
go test ./worlds/ -run TestRollInsidiousHazards -v
```

Expected: PASS — all five subcases.

Also confirm the singular tests still pass:

```bash
go test ./worlds/ -run TestRollInsidiousHazard -v
```

Expected: PASS — the existing `TestRollInsidiousHazard_AllResults` and `TestRollInsidiousHazard_ExtremelyDenseDM` are unchanged.

- [ ] **Step 5: Commit**

```bash
git add worlds/atmosphere_taint.go worlds/atmosphere_taint_test.go
git commit -m "$(cat <<'EOF'
feat(worlds): RollInsidiousHazards orchestrator (WBH p.90 footnote)

Adds a higher-level orchestrator alongside the existing dice primitive:
RollInsidiousHazards prepends an automatic Hazard{Code:"T"} for atm-C
subtypes D and E per WBH p.90 footnote ("a T hazard automatically
exists, roll again for an additional hazard"), then appends the rolled
hazard from RollInsidiousHazard. Other subtypes get a single rolled
hazard.

The "Not implemented" note on RollInsidiousHazard's doc-comment is
removed and replaced with a pointer to RollInsidiousHazards. The
existing dice primitive is unchanged.

The next commit wires this into computeBodyTaints, renames
Atmosphere.InsidiousHazard to InsidiousHazards (slice), and updates
the renderer and tests.

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Field rename + renderer + call-site + fixtures (atomic)

**Files:**

- Modify: `worlds/atmosphere.go` (rename field)
- Modify: `worlds/atmosphere_profile.go` (renderer)
- Modify: `worlds/atmosphere_profile_test.go` (existing fixture + new multi-hazard test)
- Modify: `worlds/system_detail_step5dprime.go` (call-site)
- Modify: `worlds/system_detail_step5dprime_test.go` (assertions)
- Possibly modify: `worlds/testdata/zed_markdown.golden`

This task is a model change — the field rename breaks every consumer until they're all updated. All edits land in one commit.

- [ ] **Step 1: Write the failing renderer test**

Open `worlds/atmosphere_profile_test.go`. Find the existing test `TestFormatAtmoProfileShorthand_TaintSuffix_Insidious` (the one with `InsidiousHazard: &Hazard{Code: "T"}`). Update it to use the new slice field, and append a new multi-hazard test directly after.

Replace:

```go
func TestFormatAtmoProfileShorthand_TaintSuffix_Insidious(t *testing.T) {
	t.Parallel()
	atmo := Atmosphere{
		Code:            12,
		Subtype:         "6",
		Pressure:        1.21,
		Taints:          []Taint{{Code: "G", Severity: 4, Persistence: 5}},
		InsidiousHazard: &Hazard{Code: "T"},
	}
	got := FormatAtmoProfileShorthand(atmo, AtmosphereProfile{})
	want := "C-St6.T:1.21 G.4.5"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
```

with:

```go
func TestFormatAtmoProfileShorthand_TaintSuffix_Insidious(t *testing.T) {
	t.Parallel()
	atmo := Atmosphere{
		Code:             12,
		Subtype:          "6",
		Pressure:         1.21,
		Taints:           []Taint{{Code: "G", Severity: 4, Persistence: 5}},
		InsidiousHazards: []Hazard{{Code: "T"}},
	}
	got := FormatAtmoProfileShorthand(atmo, AtmosphereProfile{})
	want := "C-St6.T:1.21 G.4.5"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAtmoProfileShorthand_TaintSuffix_Insidious_MultiHazard(t *testing.T) {
	t.Parallel()
	// Subtype D triggers the WBH p.90 footnote: auto-T + rolled hazard.
	// Hazard codes are concatenated single letters after the subtype dot.
	atmo := Atmosphere{
		Code:             12,
		Subtype:          "D",
		Pressure:         120.5,
		InsidiousHazards: []Hazard{{Code: "T"}, {Code: "G"}},
	}
	got := FormatAtmoProfileShorthand(atmo, AtmosphereProfile{})
	want := "C-StD.TG:120.50"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail (compile error)**

Run:

```bash
go test ./worlds/ -run TestFormatAtmoProfileShorthand_TaintSuffix_Insidious -v
```

Expected: FAIL — compilation error `unknown field InsidiousHazards in struct literal of type Atmosphere`.

- [ ] **Step 3: Rename the `Atmosphere` field**

Open `worlds/atmosphere.go`. Find the struct definition (around line 19-28):

```go
// Atmosphere — surface atmosphere characteristics per WBH pp.79-91.
//
// Pressure, ScaleHeight, Subtype, and Profile are populated by 3A1 with
// HZCO-bucketed provisional temperature; Step 5D (3A2b-rederive) re-derives
// these fields under the real Temperature.MeanK. Post-5D values are final
// for those fields.
//
// Taints and InsidiousHazard are populated by Step 5D-prime (post-rederive)
// per WBH pp.81-90. Taints contains 0-3 entries; InsidiousHazard is
// non-nil only for atm C (Insidious).
type Atmosphere struct {
	Code                  int
	Subtype               string
	Pressure              float64
	OxygenPartialPressure float64
	ScaleHeight           float64
	Profile               AtmosphereProfile
	Taints                []Taint
	InsidiousHazard       *Hazard
}
```

Replace with:

```go
// Atmosphere — surface atmosphere characteristics per WBH pp.79-91.
//
// Pressure, ScaleHeight, Subtype, and Profile are populated by 3A1 with
// HZCO-bucketed provisional temperature; Step 5D (3A2b-rederive) re-derives
// these fields under the real Temperature.MeanK. Post-5D values are final
// for those fields.
//
// Taints and InsidiousHazards are populated by Step 5D-prime (post-rederive)
// per WBH pp.81-90. Taints contains 0-3 entries; InsidiousHazards is
// non-empty only for atm C (Insidious) — typically 1 hazard, but 2 when
// the subtype is D or E (auto-T plus rolled, per the p.90 footnote).
type Atmosphere struct {
	Code                  int
	Subtype               string
	Pressure              float64
	OxygenPartialPressure float64
	ScaleHeight           float64
	Profile               AtmosphereProfile
	Taints                []Taint
	InsidiousHazards      []Hazard
}
```

- [ ] **Step 4: Update the renderer**

Open `worlds/atmosphere_profile.go`. Find the block at lines 516-520 (in `FormatAtmoProfileShorthand`):

```go
	// Exotic / Corrosive / Insidious
	subtypeWithHazard := atmo.Subtype
	if atmo.Code == 12 && atmo.InsidiousHazard != nil {
		subtypeWithHazard = atmo.Subtype + "." + atmo.InsidiousHazard.Code
	}
```

Replace with:

```go
	// Exotic / Corrosive / Insidious
	subtypeWithHazard := atmo.Subtype
	if atmo.Code == 12 && len(atmo.InsidiousHazards) > 0 {
		var codes strings.Builder
		for _, h := range atmo.InsidiousHazards {
			codes.WriteString(h.Code)
		}
		subtypeWithHazard = atmo.Subtype + "." + codes.String()
	}
```

Also update the doc-comment on the same function (around line 499) to show the multi-hazard example:

Find:

```go
//   - Insidious (code C=12): subtype becomes "St#.H" where H is InsidiousHazard.Code
//     e.g. "C-St6.T:1.21 G.4.5"
```

Replace with:

```go
//   - Insidious (code C=12): subtype becomes "St#.H..." where H... is the
//     concatenated single-letter codes from InsidiousHazards (1 letter for
//     most subtypes; 2 letters when subtype is D or E per WBH p.90 footnote)
//     e.g. "C-St6.T:1.21 G.4.5" or "C-StD.TG:120.50"
```

- [ ] **Step 5: Update the call-site**

Open `worlds/system_detail_step5dprime.go`. Find the clear at line 45:

```go
	atm.InsidiousHazard = nil
```

Replace with:

```go
	atm.InsidiousHazards = nil
```

Find the assignment block at lines 56-68 (the `Step 3: Insidious hazard for atm C only.` block):

```go
	// Step 3: Insidious hazard for atm C only. WBH p.90 DM+2 fires when
	// the subtype letter is "Extremely Dense" (C/D/E per WBH p.89).
	if atm.Code == 12 {
		isExtremelyDense := isExtremelyDenseSubtype(atm.Subtype)
		hazardCode := RollInsidiousHazard(r, isExtremelyDense)
		atm.InsidiousHazard = &Hazard{Code: hazardCode}
	}
```

Replace with:

```go
	// Step 3: Insidious hazards for atm C only. WBH p.90 DM+2 fires when
	// the subtype letter is "Extremely Dense" (C/D/E per WBH p.89). The
	// p.90 footnote also grants an automatic T hazard plus an additional
	// rolled hazard when the subtype is D or E — RollInsidiousHazards
	// applies that rule.
	if atm.Code == 12 {
		isExtremelyDense := isExtremelyDenseSubtype(atm.Subtype)
		atm.InsidiousHazards = RollInsidiousHazards(r, atm.Subtype, isExtremelyDense)
	}
```

Also update the doc-comment on `computeBodyTaints` at line 36 if it mentions `InsidiousHazard`:

Find:

```go
// Clears any stale Taints / InsidiousHazard before rolling so that
```

Replace with:

```go
// Clears any stale Taints / InsidiousHazards before rolling so that
```

- [ ] **Step 6: Update assertions in `system_detail_step5dprime_test.go`**

Open `worlds/system_detail_step5dprime_test.go`. There are five assertion sites:

At lines 39-41 (in `TestRunStep5DPrime_AtmCGetsHazard`), find:

```go
	if detailed[0].Atmosphere.InsidiousHazard == nil {
		t.Errorf("expected InsidiousHazard on atm C, got nil")
	}
```

Replace with:

```go
	if len(detailed[0].Atmosphere.InsidiousHazards) == 0 {
		t.Errorf("expected InsidiousHazards on atm C, got empty")
	}
```

At lines 56-58 (in `TestRunStep5DPrime_NonAtmCNoHazard`), find:

```go
	if detailed[0].Atmosphere.InsidiousHazard != nil {
		t.Errorf("got InsidiousHazard on atm B, want nil")
	}
```

Replace with:

```go
	if len(detailed[0].Atmosphere.InsidiousHazards) > 0 {
		t.Errorf("got InsidiousHazards on atm B, want empty")
	}
```

At lines 153-156 (in `TestRunStep5DPrime_ExtremelyDenseSubtypeDM`), find the entire assertion block:

```go
			if detailed[0].Atmosphere.InsidiousHazard == nil {
				t.Fatalf("subtype %q: expected InsidiousHazard, got nil", c.subtype)
			}
			got := detailed[0].Atmosphere.InsidiousHazard.Code
			if got != c.want {
				t.Errorf("subtype %q: got hazard %q, want %q", c.subtype, got, c.want)
			}
```

Replace with:

```go
			hazards := detailed[0].Atmosphere.InsidiousHazards
			if len(hazards) == 0 {
				t.Fatalf("subtype %q: expected InsidiousHazards, got empty", c.subtype)
			}
			if len(hazards) != c.wantLen {
				t.Errorf("subtype %q: got %d hazards %+v, want %d", c.subtype, len(hazards), hazards, c.wantLen)
			}
			gotCodes := ""
			for _, h := range hazards {
				gotCodes += h.Code
			}
			if gotCodes != c.want {
				t.Errorf("subtype %q: got hazard codes %q, want %q", c.subtype, gotCodes, c.want)
			}
```

The test cases also need updating to express the new expectation. Find the cases struct at the top of the test:

```go
	cases := []struct {
		name    string
		subtype string
		want    string
	}{
		{"D triggers DM+2", "D", "G"},
		{"non-CDE no DM", "6", "B"},
	}
```

Replace with:

```go
	cases := []struct {
		name    string
		subtype string
		// want is the concatenated hazard codes (e.g. "TG" = auto-T + rolled-G).
		want    string
		wantLen int
	}{
		// Subtype D fires the WBH p.90 footnote: auto-T plus rolled-G
		// (2D=4 + DM+2 = 6 → G). 2 hazards.
		{"D triggers DM+2 with auto-T", "D", "TG", 2},
		// Subtype 6: 1 rolled hazard, no DM, no auto-T.
		// 2D=4 + DM+0 = 4 → B.
		{"non-CDE no DM no auto-T", "6", "B", 1},
	}
```

- [ ] **Step 7: Run all the affected tests**

Run:

```bash
go test ./worlds/ -run "TestFormatAtmoProfileShorthand|TestRunStep5DPrime" -v
```

Expected: PASS — including the new multi-hazard renderer test, the updated insidious-hazard test, the updated step-5D-prime tests.

- [ ] **Step 8: Stage all Go changes for the modernize gate**

```bash
git add worlds/atmosphere.go worlds/atmosphere_profile.go worlds/atmosphere_profile_test.go worlds/system_detail_step5dprime.go worlds/system_detail_step5dprime_test.go
```

- [ ] **Step 9: Run the full test suite**

```bash
go test -race ./...
```

Expected: PASS, except possibly `TestRenderSystemMarkdown_ZedGolden` if seed=42 lands an atm-C body in the Zed system. If that's the only failure, proceed to Step 10. If anything else fails, escalate (BLOCKED).

- [ ] **Step 10: Refresh Zed golden (only if Step 9 flagged a golden mismatch)**

```bash
go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -update
git diff worlds/testdata/zed_markdown.golden
```

Verify the diff is limited to:

- Atm-C body's profile shorthand line, where the hazard code section may have grown from one letter to two (e.g., `C-StD.T:` → `C-StD.TG:`).

If there's no atm-C body in the seed=42 Zed system, the golden won't shift at all (Step 9 passes outright; this step is a no-op).

Anything outside that scope is a regression — escalate (BLOCKED). If clean:

```bash
git add worlds/testdata/zed_markdown.golden
go test -race ./...
```

Confirm green.

- [ ] **Step 11: Run task quality gate**

```bash
task check
```

Expected: clean.

- [ ] **Step 12: Commit**

```bash
git commit -m "$(cat <<'EOF'
fix(worlds): wire D/E hazard rule + rename to InsidiousHazards (closes #21)

Renames Atmosphere.InsidiousHazard *Hazard to InsidiousHazards []Hazard
so the WBH p.90 footnote rule can be represented: subtype D/E grants an
automatic T hazard plus an additional rolled hazard, producing 2
entries. Other subtypes produce 1.

computeBodyTaints now calls RollInsidiousHazards (plural orchestrator
added in the previous commit) instead of the dice primitive directly.

The atm-C profile shorthand renderer now concatenates single-letter
hazard codes after the subtype dot — e.g., "C-StD.TG:120.50" for
subtype D with auto-T and rolled-G — instead of rendering a single
letter.

Test fixtures and assertions updated mechanically: the old single-
pointer field becomes a slice; "InsidiousHazard != nil" checks become
"len(InsidiousHazards) > 0"; .Code accessors become [0].Code.

Closes #21.

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: PR + close-out

**Files:** none (operational steps only).

- [ ] **Step 1: Push the branch**

```bash
git push -u origin feat/insidious-de-hazard-rule
```

- [ ] **Step 2: Open the PR**

```bash
gh pr create --repo philoserf/world-builder --title "fix(worlds): insidious D/E hazard rule + rename to InsidiousHazards (closes #21)" --body "$(cat <<'EOF'
## Summary

- Adds `RollInsidiousHazards` orchestrator wrapping the existing `RollInsidiousHazard` dice primitive. For subtype D or E, prepends an automatic `Hazard{Code: "T"}` per WBH p.90 footnote, then appends the rolled hazard. Other subtypes get a single rolled hazard.
- Renames `Atmosphere.InsidiousHazard *Hazard` to `Atmosphere.InsidiousHazards []Hazard` to represent the 2-hazard case.
- Updates the renderer in `FormatAtmoProfileShorthand` to concatenate single-letter hazard codes after the subtype dot — `C-StD.TG:120.50` for subtype D with auto-T and rolled-G.
- Updates one call-site in `computeBodyTaints` and all test fixtures + assertions touched by the rename.

Closes #21.

## Spec / plan

- Spec: `docs/pass-1/specs/2026-05-09-insidious-de-hazard-rule-design.md`
- Plan: `docs/pass-1/plans/2026-05-09-insidious-de-hazard-rule.md`

## Test plan

- [x] `task check` clean (gofumpt, vet, golangci-lint, modernizer)
- [x] `task test` clean with race detector
- [x] New `TestRollInsidiousHazards` — 5 cases: subtype 6 single, subtype C with DM+2, subtype D auto-T + rolled, subtype E auto-T + rolled, duplicate-T edge case
- [x] New `TestFormatAtmoProfileShorthand_TaintSuffix_Insidious_MultiHazard` — 2-hazard concatenation rendering
- [x] Updated `TestRunStep5DPrime_ExtremelyDenseSubtypeDM` — subtype D now expects 2 hazards `[T, G]` (auto-T + rolled-G with DM+2); subtype 6 expects 1 hazard `[B]`
- [x] Updated `TestRunStep5DPrime_AtmCGetsHazard` and `TestRunStep5DPrime_NonAtmCNoHazard` — slice-shape assertions
- [x] Existing `TestRollInsidiousHazard_AllResults` and `TestRollInsidiousHazard_ExtremelyDenseDM` (singular dice primitive) unchanged
- [x] Zed golden refreshed if affected (RNG drift only — see commit body)

## Out of scope

- Other renderers (`markdown.go`, `iiss_class4p.go`) consume the profile shorthand string and don't touch `InsidiousHazard(s)` directly, so the format change flows through automatically. No changes to those files.
- The `Hazard` struct shape stays `{Code string}`; per WBH p.89, hazards are inherently lethal and constant, so no severity/persistence.
EOF
)"
```

- [ ] **Step 3: Stop**

Implementation complete on the branch; PR is open. Hand back to the user for review/merge.

---

## Self-review

**Spec coverage**

- Spec § Architecture (orchestrator): Task 1. ✓
- Spec § Architecture (field rename): Task 2 Step 3. ✓
- Spec § Architecture (call-site): Task 2 Step 5. ✓
- Spec § Architecture (renderer): Task 2 Step 4. ✓
- Spec § Decisions (D/E semantics, duplicate-T): Task 1 implementation; covered by `TestRollInsidiousHazards` cases including the duplicate-T case. ✓
- Spec § Decisions (DM+2 applies to rolled only): Task 1 implementation — `RollInsidiousHazards` calls the singular primitive with the flag, doesn't apply DM+2 to the auto-T. ✓
- Spec § Decisions (renderer format `C-StD.TG`): Task 2 Step 4 + the new multi-hazard test. ✓
- Spec § Test fixtures + assertions: Task 2 Step 1 (renderer test fixture); Task 2 Step 6 (step-5D-prime assertions). ✓
- Spec § Zed golden: Task 2 Step 10. ✓

**Placeholder scan**

No "TBD" / "TODO" / "implement later". All steps include concrete code or commands.

**Type consistency**

- `RollInsidiousHazards(r roller.Roller, subtype string, isExtremelyDense bool) []Hazard` — same signature in helper definition (Task 1 Step 3), test calls (Task 1 Step 1), and call-site (Task 2 Step 5).
- `InsidiousHazards []Hazard` — same field name in struct definition (Task 2 Step 3), renderer (Task 2 Step 4), test fixtures (Task 2 Step 1), assertions (Task 2 Step 6), and call-site (Task 2 Step 5).
- `Hazard{Code: "T"}` — same struct literal shape across all sites.
