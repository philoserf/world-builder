# Insidious "Extremely Dense" DM from Subtype Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the WBH p.90 "Atmosphere is extremely dense → DM+2" rule actually fire on the Insidious Atmosphere Hazard roll by deriving `isExtremelyDense` from `atm.Subtype` (codes C/D/E per WBH p.89) instead of an unreachable `atm.Pressure >= 10.0` check.

**Architecture:** Add a private helper `isExtremelyDenseSubtype` keyed off the subtype letter, and replace the dead pressure check at the single call-site in `computeBodyTaints`. Two TDD cycles: helper unit test → helper implementation, then integration test → call-site wiring.

**Tech Stack:** Go 1.26, `wbh/roller` (deterministic RNG via `roller.NewScripted`), table-driven tests, `task` (gofumpt + go vet + golangci-lint), `go test -race`.

---

## File Structure

- **Modify:** `worlds/atmosphere_taint.go` — add `isExtremelyDenseSubtype` helper.
- **Modify:** `worlds/atmosphere_taint_test.go` — add `TestIsExtremelyDenseSubtype`.
- **Modify:** `worlds/system_detail_step5dprime.go` — replace the `atm.Pressure >= 10.0` check and remove the stale "unreachable" comment.
- **Modify:** `worlds/system_detail_step5dprime_test.go` — add `TestRunStep5DPrime_ExtremelyDenseSubtypeDM`.
- **Possibly modify:** `worlds/testdata/zed_markdown.golden` — refresh only if Zed's seed produces an atm-12 body with subtype in {C,D,E} that shifts the hazard outcome.

---

### Task 1: Helper + unit test (TDD)

**Files:**

- Modify: `worlds/atmosphere_taint_test.go` (add new `TestIsExtremelyDenseSubtype`)
- Modify: `worlds/atmosphere_taint.go` (add new `isExtremelyDenseSubtype`)

- [ ] **Step 1: Write the failing helper test**

Open `worlds/atmosphere_taint_test.go` and append (after the last test):

```go
func TestIsExtremelyDenseSubtype(t *testing.T) {
	cases := []struct {
		subtype string
		want    bool
	}{
		{"C", true},
		{"D", true},
		{"E", true},
		{"", false},
		{"1", false},
		{"6", false},
		{"9", false},
		{"A", false},
		{"B", false},
		{"F", false},
	}
	for _, c := range cases {
		got := isExtremelyDenseSubtype(c.subtype)
		if got != c.want {
			t.Errorf("isExtremelyDenseSubtype(%q): got %v, want %v", c.subtype, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./worlds/ -run TestIsExtremelyDenseSubtype -v
```

Expected: FAIL — compilation error `undefined: isExtremelyDenseSubtype`.

- [ ] **Step 3: Implement the helper**

Open `worlds/atmosphere_taint.go`. Below `HasAnyTaint` and above `taintEligibleAtmosphere`, add:

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

- [ ] **Step 4: Run the test to verify it passes**

Run:

```bash
go test ./worlds/ -run TestIsExtremelyDenseSubtype -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add worlds/atmosphere_taint.go worlds/atmosphere_taint_test.go
git commit -m "$(cat <<'EOF'
feat(worlds): isExtremelyDenseSubtype helper (WBH p.89)

Pure helper keyed off the Corrosive/Insidious subtype letter — codes
C/D/E map to "Extremely Dense" per the WBH p.89 subtype table. Used
in the next commit to wire the WBH p.90 hazard-roll DM+2.

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Wire call-site in `computeBodyTaints` (TDD)

**Files:**

- Modify: `worlds/system_detail_step5dprime_test.go` (add new integration test)
- Modify: `worlds/system_detail_step5dprime.go` (replace pressure check + remove stale comment)
- Possibly modify: `worlds/testdata/zed_markdown.golden` (refresh only if RNG drift)

- [ ] **Step 1: Write the failing integration test**

Open `worlds/system_detail_step5dprime_test.go` and append (after the last test):

```go
func TestRunStep5DPrime_ExtremelyDenseSubtypeDM(t *testing.T) {
	// Atm C with subtype "D" should fire the WBH p.90 DM+2 on the
	// Insidious Hazard roll; subtype "6" should not.
	//
	// RollAllTaints for atm 12 with no pre-seed consumes 3 rolls
	// (subtype 2D, severity 2D, persistence 2D) before the loop breaks
	// (rawRoll != 10). Then RollInsidiousHazard consumes 1 roll.
	//
	// Subtype roll 7 → 2D=7+0=7 → taintSubtypeFromTotal=G → atm 12 is
	// outside 4-9 so any L/H would be suppressed to G; G is already G.
	// No ppO2 adjust path. Severity 7 (atm-C DM+6) and persistence 7
	// don't matter for this test — they're stable across the two cases.
	//
	// Hazard roll 4:
	//   subtype "D" → DM+2 → 6 → G (hazardFromTotal(6) == "G")
	//   subtype "6" → DM 0  → 4 → B (hazardFromTotal(<=4) == "B")
	cases := []struct {
		name    string
		subtype string
		want    string
	}{
		{"D triggers DM+2", "D", "G"},
		{"non-CDE no DM", "6", "B"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := roller.NewScripted(7, 7, 7, 4)
			sys := stars.System{Primary: stars.Star{AgeGyr: 5.0}}
			detailed := []DetailedPlacement{{
				Placement:  Placement{Body: BodyTerrestrial},
				HZ:         true,
				SizeCode:   "8",
				Atmosphere: &Atmosphere{Code: 12, Subtype: c.subtype},
			}}
			if err := runStep5DPrime(r, detailed, sys); err != nil {
				t.Fatalf("runStep5DPrime: %v", err)
			}
			if detailed[0].Atmosphere.InsidiousHazard == nil {
				t.Fatalf("subtype %q: expected InsidiousHazard, got nil", c.subtype)
			}
			got := detailed[0].Atmosphere.InsidiousHazard.Code
			if got != c.want {
				t.Errorf("subtype %q: got hazard %q, want %q", c.subtype, got, c.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails on the "D" case**

Run:

```bash
go test ./worlds/ -run TestRunStep5DPrime_ExtremelyDenseSubtypeDM -v
```

Expected: subcase `D triggers DM+2` FAILS — got hazard `"B"`, want `"G"`. The non-CDE case may pass coincidentally (it asserts the no-DM behavior, which is the current behavior). The `D` case proving the DM is dead is what matters.

- [ ] **Step 3: Replace the dead pressure check with the helper**

Open `worlds/system_detail_step5dprime.go`. Find lines 56–68 (the `Step 3: Insidious hazard for atm C only.` block plus the stale comment above it) and replace the entire block with:

```go
	// Step 3: Insidious hazard for atm C only. WBH p.90 DM+2 fires when
	// the subtype letter is "Extremely Dense" (C/D/E per WBH p.89).
	if atm.Code == 12 {
		isExtremelyDense := isExtremelyDenseSubtype(atm.Subtype)
		hazardCode := RollInsidiousHazard(r, isExtremelyDense)
		atm.InsidiousHazard = &Hazard{Code: hazardCode}
	}
```

The exact `old_string` for the Edit is the multi-line block beginning with the comment `// Step 3: Insidious hazard for atm C only.` and ending with the closing `}` of the `if atm.Code == 12 {` block — including the stale "isExtremelyDense currently never fires" comment and the `atm.Pressure >= 10.0` line. Verify with `git diff worlds/system_detail_step5dprime.go` after the edit; the diff should remove the 6-line stale comment and replace one assignment line.

- [ ] **Step 4: Run the test to verify it passes**

Run:

```bash
go test ./worlds/ -run TestRunStep5DPrime_ExtremelyDenseSubtypeDM -v
```

Expected: PASS — both subcases.

- [ ] **Step 5: Run full test suite**

Run:

```bash
go test -race ./...
```

Expected: PASS, except possibly `TestRenderSystemMarkdown_ZedGolden` if seed=42 lands an atm-12 body with subtype in {C,D,E} and the hazard outcome shifts. If that fails, proceed to Step 6; if it passes, skip to Step 7.

- [ ] **Step 6: Refresh Zed golden (only if Step 5 flagged a golden mismatch)**

Run:

```bash
go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -update
```

Then inspect the diff:

```bash
git diff worlds/testdata/zed_markdown.golden
```

Verify the diff is limited to atm-C body hazard-suffix changes (e.g., `C-St?.B:...` → `C-St?.G:...`) or expected RNG drift downstream of the hazard roll. Anything outside that scope is a regression — investigate before continuing. Re-run `go test -race ./...` and confirm green.

- [ ] **Step 7: Run task quality gate**

Run:

```bash
task check
```

Expected: clean. (gofumpt, go vet, golangci-lint, modernizer all pass.)

- [ ] **Step 8: Commit**

```bash
git add worlds/system_detail_step5dprime.go worlds/system_detail_step5dprime_test.go
# Also add the golden if it was refreshed:
# git add worlds/testdata/zed_markdown.golden
git commit -m "$(cat <<'EOF'
fix(worlds): wire WBH p.90 extremely-dense DM from atm Subtype (closes #22)

Replace the dead atm.Pressure >= 10.0 check (insidious atms have
Pressure == 0 because AtmospherePressureRange returns "Varies") with
isExtremelyDenseSubtype keyed off subtype letters C/D/E per WBH p.89.

The DM+2 on the Insidious Atmosphere Hazard roll now fires for atm-C
bodies that land in the "Extremely Dense" subtype rows.

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: PR + carry-forward issue

**Files:** none (operational steps only).

- [ ] **Step 1: Push the branch**

Run:

```bash
git push -u origin feat/atm-c-extremely-dense-from-subtype
```

- [ ] **Step 2: Open the PR**

Run:

```bash
gh pr create --repo philoserf/world-builder --title "fix(worlds): wire WBH p.90 extremely-dense DM from atm Subtype (closes #22)" --body "$(cat <<'EOF'
## Summary

- Adds `isExtremelyDenseSubtype` helper (private) keyed off the WBH p.89 subtype table — rows 12/13/14+ map to codes C/D/E.
- Replaces the dead `atm.Pressure >= 10.0` check in `computeBodyTaints` with `isExtremelyDenseSubtype(atm.Subtype)`. The WBH p.90 Insidious Atmosphere Hazard DM+2 now fires for atm-C bodies in the Extremely Dense subtype rows.
- Removes the stale "unreachable branch" comment in `computeBodyTaints`.

Closes #22.

## Spec / plan

- Spec: `docs/specs/2026-05-09-insidious-extremely-dense-from-subtype-design.md`
- Plan: `docs/plans/2026-05-09-insidious-extremely-dense-from-subtype.md`

## Test plan

- [x] `task check` clean (gofumpt, vet, golangci-lint, modernizer)
- [x] `task test` clean with race detector
- [x] New `TestIsExtremelyDenseSubtype` — table-driven coverage of {"C","D","E"} → true and a sweep of negatives
- [x] New `TestRunStep5DPrime_ExtremelyDenseSubtypeDM` — integration: subtype "D" + scripted hazard 2D=4 → "G" (DM fires); subtype "6" + same roll → "B" (DM doesn't fire)
- [x] Existing `TestRollInsidiousHazard_ExtremelyDenseDM` and `TestRunStep5DPrime_AtmCGetsHazard` unchanged

## Out of scope (carry-forward)

- Populating `atm.Pressure` for atm B/C from the WBH p.89 subtype-keyed pressure ranges. Tracked separately.
- WBH p.90 footnote 1 (subtype D/E auto-T hazard plus additional rolled hazard). Tracked as #21.
EOF
)"
```

- [ ] **Step 3: File the carry-forward issue**

Run:

```bash
gh issue create --repo philoserf/world-builder --title "enhancement: 3B atmosphere — populate atm B/C Pressure from WBH p.89 subtype-keyed range" --body "$(cat <<'EOF'
WBH p.89 gives explicit per-subtype pressure ranges for Corrosive (B) and Insidious (C) atmospheres:

| Subtype Code | Pressure Range (bar) | Span    |
| ------------ | -------------------- | ------- |
| 1, 2, 3      | 0.1–0.42             | 0.32    |
| 4, 5         | 0.43–0.70            | 0.27    |
| 6, 7         | 0.70–1.49            | 0.79    |
| 8, 9         | 1.50–2.49            | 0.99    |
| A, B         | 2.50–10.0            | 7.50    |
| C, D, E      | 10.0+                | unbound |

Book quote: "More detailed atmospheric pressures for these atmospheres can be determined from pressure range (bars) and span using either of the total atmospheric pressure (bar) equations on page 80."

Current state (post #22): `AtmospherePressureRange(11)` and `AtmospherePressureRange(12)` both return `(0, 0)`, so `RollTotalPressure` leaves `atm.Pressure == 0` for atm B and atm C bodies. Real pressures are silently lost.

Implementation sketch:

- New table keyed by subtype letter for atm B and atm C (same ranges per the p.89 row).
- New roll path or extension to `RollTotalPressure` that consults the subtype-keyed table when atm code is 11 or 12 and a Subtype is set.
- Pipeline ordering: this must run after `RollCorrosiveInsidiousSubtype` populates Subtype. Today, `RollTotalPressure` runs before subtype is set (in `system_detail_steps.go`); a re-derivation pass post-subtype is likely required.
- "C, D, E" pressure is unbound (10.0+). Pick a deterministic distribution — e.g., 10.0 + 1D × 1000 bar, or stay with a simple low-end value; the book defers to the p.80 equations.
- RNG drift event on goldens.

Spotted while researching #22.
EOF
)"
```

- [ ] **Step 4: Stop**

Implementation complete on the branch; PR is open. Hand back to the user for review/merge.

---

## Self-review

**Spec coverage**

- Spec § Architecture: Helper added in Task 1. ✓
- Spec § Architecture: Call-site change in Task 2 Step 3. ✓
- Spec § Pre-flight verification: subtype-populated check is implicit (the test scripts subtype directly so the helper is exercised even if production wiring drifts). The "Run the live pipeline with the Zed seed" check is Step 5/6 of Task 2. ✓
- Spec § Testing strategy unit test: Task 1 Step 1. ✓
- Spec § Testing strategy integration test: Task 2 Step 1. ✓
- Spec § Carry-forward: Task 3 Step 3. ✓
- Spec § Out of scope (D/E hazard rule, atm B): both untouched.

**Placeholder scan**

No "TBD"/"TODO"/"implement later". All steps include concrete code or commands.

**Type consistency**

`isExtremelyDenseSubtype(string) bool` — same signature in helper definition (Task 1 Step 3), test call (Task 1 Step 1), and call-site (Task 2 Step 3). `Hazard.Code` field name matches existing struct in `atmosphere_taint.go`.
