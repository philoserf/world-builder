# Atm B/C Pressure From Subtype Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Populate `atm.Pressure` for atm B (code 11) and atm C (code 12) bodies using subtype-keyed pressure ranges from WBH p.89, replacing the current `(0,0)` "Varies" short-circuit.

**Architecture:** Add a private helper `corrosiveInsidiousPressureRange(subtype)` keyed off the WBH p.89 subtype letters (1–E). Extend `RollTotalPressure` to accept a subtype string and route atm 11/12 with non-empty subtype to the new helper. Three production call sites already roll subtype before pressure; pass it through.

**Tech Stack:** Go 1.26, `wbh/roller`, `task` (gofumpt + go vet + golangci-lint + modernizer), `go test -race`.

---

## File Structure

- **Modify:** `worlds/atmosphere.go` — add `corrosiveInsidiousPressureRange` helper; extend `RollTotalPressure` signature.
- **Modify:** `worlds/atmosphere_test.go` — add unit tests for the helper; add behavior tests for the new `RollTotalPressure` signature; update existing `TestRollTotalPressure_ZedPrime` for the new signature.
- **Modify:** `worlds/system_detail_steps.go` — two callers (planet path + moon path) pass subtype.
- **Modify:** `worlds/temperature_rederive.go` — one caller (rederive path) passes subtype.
- **Possibly modify:** `worlds/testdata/zed_markdown.golden` — refresh if Zed's atm-B body's pressure now differs from 0.

---

### Task 1: Helper + unit test (TDD)

**Files:**

- Modify: `worlds/atmosphere_test.go` (add `TestCorrosiveInsidiousPressureRange`)
- Modify: `worlds/atmosphere.go` (add `corrosiveInsidiousPressureRange`)

- [ ] **Step 1: Write the failing helper test**

Open `worlds/atmosphere_test.go` and append (after the last test):

```go
func TestCorrosiveInsidiousPressureRange(t *testing.T) {
	cases := []struct {
		subtype  string
		wantMin  float64
		wantSpan float64
	}{
		{"1", 0.1, 0.32},
		{"2", 0.1, 0.32},
		{"3", 0.1, 0.32},
		{"4", 0.43, 0.27},
		{"5", 0.43, 0.27},
		{"6", 0.70, 0.79},
		{"7", 0.70, 0.79},
		{"8", 1.50, 0.99},
		{"9", 1.50, 0.99},
		{"A", 2.50, 7.50},
		{"B", 2.50, 7.50},
		{"C", 10, 90},
		{"D", 100, 900},
		{"E", 1000, 9000},
		{"", 0, 0},
		{"0", 0, 0},
		{"Z", 0, 0},
	}
	for _, c := range cases {
		gotMin, gotSpan := corrosiveInsidiousPressureRange(c.subtype)
		if gotMin != c.wantMin || gotSpan != c.wantSpan {
			t.Errorf("corrosiveInsidiousPressureRange(%q): got (%g, %g), want (%g, %g)",
				c.subtype, gotMin, gotSpan, c.wantMin, c.wantSpan)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./worlds/ -run TestCorrosiveInsidiousPressureRange -v
```

Expected: FAIL — compilation error `undefined: corrosiveInsidiousPressureRange`.

- [ ] **Step 3: Implement the helper**

Open `worlds/atmosphere.go`. Below the existing `AtmospherePressureRange` function, add:

```go
// corrosiveInsidiousPressureRange returns (minBar, spanBar) for atm
// codes B (11) and C (12) keyed off the WBH p.89 subtype letter.
//
// Subtypes 1-B carry the explicit ranges from the p.89 table.
// Subtypes C/D/E carry "10.0+ / unbound" in the book; we return
// project-supplied tiered ranges per the design spec dated 2026-05-09
// (atm-bc-pressure-from-subtype) honoring p.89's "Only insidious
// extremely dense atmospheres should have pressures exceeding 1,000
// bar" hint:
//
//	C: 10–100      (min=10,   span=90)
//	D: 100–1000    (min=100,  span=900)
//	E: 1000–10000  (min=1000, span=9000)
//
// Empty/unknown subtype returns (0, 0). Callers in the live pipeline
// always roll the subtype before pressure (see system_detail_steps.go
// and temperature_rederive.go).
func corrosiveInsidiousPressureRange(subtype string) (minBar, spanBar float64) {
	switch subtype {
	case "1", "2", "3":
		return 0.1, 0.32
	case "4", "5":
		return 0.43, 0.27
	case "6", "7":
		return 0.70, 0.79
	case "8", "9":
		return 1.50, 0.99
	case "A", "B":
		return 2.50, 7.50
	case "C":
		return 10, 90
	case "D":
		return 100, 900
	case "E":
		return 1000, 9000
	}
	return 0, 0
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run:

```bash
go test ./worlds/ -run TestCorrosiveInsidiousPressureRange -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add worlds/atmosphere.go worlds/atmosphere_test.go
git commit -m "$(cat <<'EOF'
feat(worlds): corrosiveInsidiousPressureRange helper (WBH p.89)

Pure helper keyed off the Corrosive/Insidious subtype letter — codes
1-B map to the explicit p.89 table values; C/D/E map to project-
supplied tiered ranges (10–100, 100–1000, 1000–10000 bar) per the
design spec for #24. Used in the next commit to wire RollTotalPressure
for atm B/C bodies.

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Extend `RollTotalPressure` signature + update all callers (TDD)

**Files:**

- Modify: `worlds/atmosphere_test.go` (add behavior tests; update existing `TestRollTotalPressure_ZedPrime` signature call)
- Modify: `worlds/atmosphere.go` (extend `RollTotalPressure` signature)
- Modify: `worlds/system_detail_steps.go` (two call sites)
- Modify: `worlds/temperature_rederive.go` (one call site)
- Possibly modify: `worlds/testdata/zed_markdown.golden` (refresh only on RNG drift)

- [ ] **Step 1: Write the failing behavior tests**

Open `worlds/atmosphere_test.go` and append (after `TestCorrosiveInsidiousPressureRange`):

```go
func TestRollTotalPressure_AtmBCWithSubtype(t *testing.T) {
	cases := []struct {
		name    string
		atmCode int
		subtype string
		// Roll values for the formula's two 1D rolls.
		// scale = ((a-1)*5 + (b-1)) / 30; pressure = min + span*scale.
		a, b    int
		want    float64
	}{
		// Subtype 6 (Standard): min=0.70, span=0.79.
		// (1,1) → scale=0 → 0.70.
		{"atm B subtype 6 min", 11, "6", 1, 1, 0.70},
		// (6,6) → scale=1 → 0.70+0.79 = 1.49.
		{"atm B subtype 6 max", 11, "6", 6, 6, 1.49},
		// Subtype C: min=10, span=90.
		{"atm C subtype C min", 12, "C", 1, 1, 10},
		{"atm C subtype C max", 12, "C", 6, 6, 100},
		// Subtype E: min=1000, span=9000.
		{"atm C subtype E min", 12, "E", 1, 1, 1000},
		{"atm C subtype E max", 12, "E", 6, 6, 10000},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := roller.NewScripted(c.a, c.b)
			got, err := RollTotalPressure(r, c.atmCode, c.subtype)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got != c.want {
				t.Errorf("got %g, want %g", got, c.want)
			}
		})
	}
}

func TestRollTotalPressure_AtmBCEmptySubtype(t *testing.T) {
	// Empty subtype on atm 11/12 falls back to (0, 0): no rolls consumed,
	// returns 0. This preserves legacy "Varies" behavior for callers that
	// don't have a subtype yet.
	r := roller.NewScripted() // no rolls scripted; if any are consumed, panic.
	got, err := RollTotalPressure(r, 11, "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 0 {
		t.Errorf("got %g, want 0", got)
	}
}

func TestRollTotalPressure_RegularCodeIgnoresSubtype(t *testing.T) {
	// For atm codes outside 11/12, the subtype parameter is ignored.
	// Atm 6 (Standard): min=0.70, span=0.79. (1,1) → 0.70.
	r := roller.NewScripted(1, 1)
	got, err := RollTotalPressure(r, 6, "C") // subtype "C" ignored on atm 6
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 0.70 {
		t.Errorf("got %g, want 0.70", got)
	}
}
```

Update the existing `TestRollTotalPressure_ZedPrime` (currently at `worlds/atmosphere_test.go:139`) to pass the new third argument. Replace:

```go
	got, err := RollTotalPressure(r, 6)
```

with:

```go
	got, err := RollTotalPressure(r, 6, "")
```

- [ ] **Step 2: Run the tests to verify they fail (compile error)**

Run:

```bash
go test ./worlds/ -run TestRollTotalPressure -v
```

Expected: FAIL — compilation error `too many arguments in call to RollTotalPressure` (the new tests pass 3 args; signature still accepts 2).

- [ ] **Step 3: Extend `RollTotalPressure` signature**

Open `worlds/atmosphere.go`. Find the existing function:

```go
// RollTotalPressure computes total atmospheric pressure per WBH p.80:
//
//	bar = MinPressureRange + Span × ((1D-1)×5 + (1D-1)) / 30
//
// Returns minBar with no rolls consumed when span = 0.
func RollTotalPressure(r roller.Roller, atmoCode int) (float64, error) {
	minBar, span := AtmospherePressureRange(atmoCode)
	if span == 0 {
		return minBar, nil
	}
	a := r.Roll("1D")
	b := r.Roll("1D")
	scale := float64((a-1)*5+(b-1)) / 30.0
	return minBar + span*scale, nil
}
```

Replace with:

```go
// RollTotalPressure computes total atmospheric pressure per WBH p.80,
// or per the WBH p.89 subtype-keyed range for atm B/C:
//
//	bar = MinPressureRange + Span × ((1D-1)×5 + (1D-1)) / 30
//
// For atm codes 11 (B) and 12 (C), subtype must be set (one of
// "1"-"9", "A"-"E"); otherwise pressure falls back to 0 (legacy
// "Varies" behavior). For all other codes, subtype is ignored.
//
// Returns minBar with no rolls consumed when span = 0.
func RollTotalPressure(r roller.Roller, atmoCode int, subtype string) (float64, error) {
	var minBar, span float64
	if (atmoCode == 11 || atmoCode == 12) && subtype != "" {
		minBar, span = corrosiveInsidiousPressureRange(subtype)
	} else {
		minBar, span = AtmospherePressureRange(atmoCode)
	}
	if span == 0 {
		return minBar, nil
	}
	a := r.Roll("1D")
	b := r.Roll("1D")
	scale := float64((a-1)*5+(b-1)) / 30.0
	return minBar + span*scale, nil
}
```

- [ ] **Step 4: Update all production callers**

Open `worlds/system_detail_steps.go`. At line 95 (planet path), replace:

```go
			press, perr := RollTotalPressure(r, atmoCode)
```

with:

```go
			press, perr := RollTotalPressure(r, atmoCode, atmo.Subtype)
```

At line 147 (moon path), replace:

```go
				press, perr := RollTotalPressure(r, atmoCode)
```

with:

```go
				press, perr := RollTotalPressure(r, atmoCode, atmo.Subtype)
```

Open `worlds/temperature_rederive.go`. At line 203 (rederive path), replace:

```go
	newPressure, err := RollTotalPressure(r, code)
```

with:

```go
	newPressure, err := RollTotalPressure(r, code, newSubtype)
```

- [ ] **Step 5: Run the new tests to verify they pass**

Run:

```bash
go test ./worlds/ -run TestRollTotalPressure -v
```

Expected: PASS — all four tests (`TestRollTotalPressure_ZedPrime`, `TestRollTotalPressure_AtmBCWithSubtype`, `TestRollTotalPressure_AtmBCEmptySubtype`, `TestRollTotalPressure_RegularCodeIgnoresSubtype`).

- [ ] **Step 6: Stage the Go changes for the modernize gate**

The `task check` modernize gate requires staged Go changes (it runs `go fix ./...` and errors on any unstaged diff). Stage:

```bash
git add worlds/atmosphere.go worlds/atmosphere_test.go worlds/system_detail_steps.go worlds/temperature_rederive.go
```

- [ ] **Step 7: Run the full test suite**

Run:

```bash
go test -race ./...
```

Expected: PASS, except possibly `TestRenderSystemMarkdown_ZedGolden` if Zed's atm-B body's pressure now differs from 0. If that's the only failure, proceed to Step 8. If anything else fails, escalate (BLOCKED).

- [ ] **Step 8: Refresh Zed golden (only if Step 7 flagged a golden mismatch)**

Run:

```bash
go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -update
git diff worlds/testdata/zed_markdown.golden
```

Verify the diff is limited to:

- Zed Prime's atm-B `Pressure` field, switching from 0 to a value in `[0.70, 1.49]`.
- The `Profile` shorthand line that includes pressure (e.g., `B-St6:bar:gases ...`).
- Downstream RNG drift caused by the two 1D rolls now being consumed by `RollTotalPressure` for atm B (where they previously were skipped because span was 0).

Anything outside that scope is a regression — escalate (BLOCKED) before continuing. If clean, stage:

```bash
git add worlds/testdata/zed_markdown.golden
```

Re-run the full suite to confirm green:

```bash
go test -race ./...
```

- [ ] **Step 9: Run task quality gate**

Run:

```bash
task check
```

Expected: clean (gofumpt, vet, golangci-lint, modernizer all pass).

- [ ] **Step 10: Commit**

```bash
git commit -m "$(cat <<'EOF'
fix(worlds): atm B/C pressure from subtype-keyed range (closes #24)

RollTotalPressure now accepts a subtype string. For atm codes 11 (B)
and 12 (C) with a non-empty subtype, pressure is rolled from the
WBH p.89 subtype-keyed range (via corrosiveInsidiousPressureRange).
Other codes ignore the subtype parameter and use AtmospherePressureRange
as before; an empty subtype on atm 11/12 falls back to 0 (legacy
"Varies" behavior).

The three production callers (planet path, moon path, rederive path)
already roll subtype before pressure, so passing it through is
mechanical.

Closes #24.

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: PR + close-out

**Files:** none (operational steps only).

- [ ] **Step 1: Push the branch**

Run:

```bash
git push -u origin feat/atm-bc-pressure-from-subtype
```

- [ ] **Step 2: Open the PR**

Run:

```bash
gh pr create --repo philoserf/world-builder --title "fix(worlds): atm B/C pressure from subtype-keyed range (closes #24)" --body "$(cat <<'EOF'
## Summary

- Adds private `corrosiveInsidiousPressureRange` helper keyed off the WBH p.89 subtype table — subtypes 1-B map to explicit ranges from the table; subtypes C/D/E map to project-supplied tiered ranges (10–100, 100–1000, 1000–10000 bar) per the design spec.
- Extends `RollTotalPressure` to accept a subtype string. Atm codes 11 (B) / 12 (C) with non-empty subtype route to the new helper; other codes ignore the subtype.
- Updates the three production callers (planet path, moon path, rederive path) to pass `atm.Subtype` / `newSubtype`.
- Real B/C pressures now flow into downstream consumers (greenhouse formula, profile shorthand). Greenhouse output is bounded by the existing `(1+G)` clamp at `[0.001, 1.999]` per WBH p.111 thumb-rule-two — no temperature blowup.

Closes #24.

## Spec / plan

- Spec: `docs/history/pass-1-specs/2026-05-09-atm-bc-pressure-from-subtype-design.md`
- Plan: `docs/history/pass-1-plans/2026-05-09-atm-bc-pressure-from-subtype.md`

## Test plan

- [x] `task check` clean (gofumpt, vet, golangci-lint, modernizer)
- [x] `task test` clean with race detector
- [x] New `TestCorrosiveInsidiousPressureRange` — table coverage of all 14 subtype letters plus empty/unknown
- [x] New `TestRollTotalPressure_AtmBCWithSubtype` — subtype 6 / C / E min and max bounds via scripted dice
- [x] New `TestRollTotalPressure_AtmBCEmptySubtype` — empty subtype on atm 11 returns 0 with no rolls consumed
- [x] New `TestRollTotalPressure_RegularCodeIgnoresSubtype` — subtype is ignored on non-11/12 codes
- [x] Existing `TestRollTotalPressure_ZedPrime` updated for the new signature
- [x] Zed golden refreshed if required (RNG drift only — see commit body)

## Out of scope (per spec)

- Atm A "Exotic" subtype-keyed pressures (book describes qualitatively, not via a table).
- Atm F/G/H "Varies" pressures (no book table to source from).
- WBH p.90 footnote 1 (subtype D/E auto-T hazard plus additional rolled hazard) — tracked as #21.
EOF
)"
```

- [ ] **Step 3: Stop**

Implementation complete on the branch; PR is open. Hand back to the user for review/merge.

---

## Self-review

**Spec coverage**

- Spec § Architecture (helper): Task 1. ✓
- Spec § API (`RollTotalPressure` extended): Task 2 Step 3. ✓
- Spec § Call sites (3 sites): Task 2 Step 4. ✓
- Spec § Out of scope: not touched (correct). ✓
- Spec § Testing strategy unit tests: Task 1 Step 1 + Task 2 Step 1. ✓
- Spec § Testing strategy behavior tests: Task 2 Step 1. ✓
- Spec § Zed golden refresh: Task 2 Step 8. ✓

**Placeholder scan**

No "TBD" / "TODO" / "implement later". All steps include concrete code or commands. The Zed golden refresh in Step 8 has explicit verification criteria (which fields are expected to shift) so the engineer can spot regressions vs. drift.

**Type consistency**

- `corrosiveInsidiousPressureRange(subtype string) (minBar, spanBar float64)` — same signature in helper definition (Task 1 Step 3), test calls (Task 1 Step 1), and dispatch in `RollTotalPressure` (Task 2 Step 3).
- `RollTotalPressure(r, atmoCode, subtype)` — same 3-arg signature in test calls (Task 2 Step 1) and the three production updates (Task 2 Step 4).
- Variable names used in test scaffolding match the production code (`atmo.Subtype` for the planet/moon callers, `newSubtype` for the rederive caller).
