# Biology Oxygen-Atm Floor (WBH p.128 Optional Rule 1) — Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add opt-in `DetailOpts.OxygenAtmBiomassFloor` per WBH p.128 — partial close of issue #12 (Rule 1 of 2).

**Architecture:** Introduce a `DetailOpts` struct and a sibling `DetailSystemWithOpts` constructor. Existing `DetailSystem(...)` becomes a thin wrapper that calls `DetailSystemWithOpts(..., DetailOpts{})` so all 7 existing callers stay byte-identical. Internal pipeline plumbing: `runDetailPipeline(..., opts)` → `runStep5F(..., opts)` → `computeBiology(..., opts)`. Inside `computeBiology`, after `RollBiomass` returns, clamp biomass to ≥1 when opts opt-in is set AND the body's atmosphere code is in the oxygen-bearing set {2-9, D, E}.

**Tech Stack:** Go 1.26, `task check && task test` is the gate.

---

### Task 1: Add DetailOpts struct + DetailSystemWithOpts constructor

**Files:**

- Modify: `worlds/system_detail.go` (add struct + WithOpts function, refactor existing into wrapper)

- [ ] **Step 1: Add DetailOpts and DetailSystemWithOpts; reduce DetailSystem to a wrapper**

In `worlds/system_detail.go`, **above** the `DetailSystem` function:

```go
// DetailOpts gates opt-in WBH rules and Referee-discretion variants for
// the per-body detail pipeline. The zero value disables every opt — so
// DetailSystem(...) (the no-opts wrapper) preserves canonical-book
// behavior for all existing callers.
type DetailOpts struct {
	// OxygenAtmBiomassFloor enables WBH p.128 Optional Rule: any world
	// whose Atmosphere.Code is in the oxygen-bearing set {2-9, D, E}
	// gets a biomass floor of 1 (the rolled value is clamped up if it
	// came in below). Off by default — the book describes it as a
	// Referee opt-in.
	OxygenAtmBiomassFloor bool
}
```

Then update `DetailSystem` to be a thin wrapper, and rename the existing implementation body to `DetailSystemWithOpts`:

```go
// DetailSystem composes the full WBH pp. 53-67 procedure on top of a
// SystemPlacement (2B output). Returns a SystemDetail with sizes,
// moons, designations, periods, HZ tags, profiles, and the IISS
// Class II/III form.
//
// Equivalent to DetailSystemWithOpts with a zero-valued DetailOpts —
// canonical-book behavior, no opt-in rules.
//
// Pipeline:
//
//  1. runDetailPipeline: per-body Steps 1-5 + 5A-5G (sizing, moons,
//     designations, periods, HZ, then 3A1/3A2/3B passes).
//  2. Backfill StarAllocation.BaselineN.
//  3. ShortProfile + LongProfile.
//  4. RenderIISSClass23.
//  5. pickMainworld.
func DetailSystem(r roller.Roller, sys stars.System, sp SystemPlacement, h IISSClass23Header) (SystemDetail, error) {
	return DetailSystemWithOpts(r, sys, sp, h, DetailOpts{})
}

// DetailSystemWithOpts is the opt-aware variant of DetailSystem. See
// DetailOpts for the available opt-in rules.
func DetailSystemWithOpts(r roller.Roller, sys stars.System, sp SystemPlacement, h IISSClass23Header, opts DetailOpts) (SystemDetail, error) {
	detailed := make([]DetailedPlacement, len(sp.Placements))
	for i := range sp.Placements {
		detailed[i] = DetailedPlacement{Placement: sp.Placements[i]}
	}

	if err := runDetailPipeline(r, detailed, sys, sp, opts); err != nil {
		return SystemDetail{}, err
	}

	// ... rest of existing body unchanged through to `return sd, nil`.
}
```

Note: only the call to `runDetailPipeline` changes — pass `opts` as the 5th argument. Everything from "Step 6 — backfill StarAllocation.BaselineN" onward stays byte-identical.

- [ ] **Step 2: Compile-check**

Run: `go build ./...`

Expected: BUILD FAILS — `runDetailPipeline` signature does not yet accept `opts`. That failure is the next task's target.

---

### Task 2: Plumb opts through runDetailPipeline → runStep5F

**Files:**

- Modify: `worlds/system_detail_pipeline.go` (add `opts DetailOpts` parameter; pass through to `runStep5F`)
- Modify: `worlds/system_detail_step5f.go` (add `opts DetailOpts` parameter)

- [ ] **Step 1: Add opts parameter to runDetailPipeline and forward to runStep5F**

In `worlds/system_detail_pipeline.go`, change the signature:

```go
func runDetailPipeline(r roller.Roller, detailed []DetailedPlacement, sys stars.System, sp SystemPlacement, opts DetailOpts) error {
```

Update only the `runStep5F` call (line 116 in the current file):

```go
	// Step 5F — 3B-biology pass: native lifeform ratings + resource rating.
	if err := runStep5F(r, detailed, sys, opts); err != nil {
		return err
	}
```

All other Step 5\* calls keep their existing signatures — only Step 5F consumes opts in this loop.

- [ ] **Step 2: Add opts parameter to runStep5F and forward to computeBiology**

In `worlds/system_detail_step5f.go`, change the function signature:

```go
func runStep5F(r roller.Roller, detailed []DetailedPlacement, sys stars.System, opts DetailOpts) error {
```

Both `computeBiology(...)` call sites inside `runStep5F` get an extra `opts` argument:

```go
		if biologyApplies(dp) {
			dp.Biology = computeBiology(r, dp, sys.Primary.AgeGyr, opts)
		}
		// ...
			m.Biology = computeBiology(r, moonDP, sys.Primary.AgeGyr, opts)
```

- [ ] **Step 3: Compile-check**

Run: `go build ./...`

Expected: BUILD FAILS — `computeBiology` signature does not yet accept `opts`. Next task fixes it.

---

### Task 3: Apply the floor in computeBiology + add hasOxygenAtmosphere helper

**Files:**

- Modify: `worlds/system_detail_step5f.go` (extend `computeBiology` signature + apply floor)
- Modify: `worlds/biology.go` (add `hasOxygenAtmosphere` helper near the existing exotic-bonus helpers)

- [ ] **Step 1: Add hasOxygenAtmosphere helper in worlds/biology.go**

Insert near `exoticBiomassBonusApplies` (around line 177 of `worlds/biology.go`):

```go
// hasOxygenAtmosphere reports whether atm carries free oxygen per
// WBH p.128 Optional Rule (codes 2-9, D, E). Hex codes:
//   2-9: progressively thicker oxygen atmospheres
//   D (13): "Very Dense" oxygen atmosphere (2.50-10.0 bar per WBH p.79)
//   E (14): "Low" oxygen atmosphere (0.10-0.42 bar per WBH p.79)
//
// Excluded: 0 (None), 1 (Trace), A (10, Exotic), B (11, Corrosive),
// C (12, Insidious), F (15, Unusual).
func hasOxygenAtmosphere(atm *Atmosphere) bool {
	if atm == nil {
		return false
	}
	code := atm.Code
	return (code >= 2 && code <= 9) || code == 13 || code == 14
}
```

- [ ] **Step 2: Update computeBiology to take opts and apply the floor**

In `worlds/system_detail_step5f.go`:

```go
// computeBiology populates a Biology for the given body. Caller has
// already verified biologyApplies(dp). When opts.OxygenAtmBiomassFloor
// is set AND the body's atmosphere is in the oxygen-bearing set, a
// rolled biomass below 1 is clamped up to 1 per WBH p.128 Optional
// Rule. The floor runs before dependent rolls so an elevated biomass
// of 1 propagates naturally into Biocomplexity / Biodiversity /
// Compatibility.
func computeBiology(r roller.Roller, dp *DetailedPlacement, ageGyr float64, opts DetailOpts) *Biology {
	bio := &Biology{}
	bio.Biomass = RollBiomass(r, dp, ageGyr)
	if opts.OxygenAtmBiomassFloor && bio.Biomass < 1 && hasOxygenAtmosphere(dp.Atmosphere) {
		bio.Biomass = 1
	}
	if bio.Biomass > 0 {
		bio.Biocomplexity = RollBiocomplexity(r, dp, bio.Biomass, ageGyr)
		if bio.Biocomplexity >= 8 {
			bio.HasNativeSophont = RollNativeSophont(r, bio.Biocomplexity)
			bio.HadExtinctSophont = RollExtinctSophont(r, bio.Biocomplexity, ageGyr)
		}
		bio.Biodiversity = RollBiodiversity(r, bio.Biomass, bio.Biocomplexity)
		bio.Compatibility = RollCompatibility(r, dp, bio.Biocomplexity, ageGyr)
	}
	bio.ResourceRating = RollTerrestrialResourceRating(r, dp, bio)
	return bio
}
```

- [ ] **Step 3: Compile-check + run existing test suite**

Run: `go build ./... && go test ./worlds/ -count=1`

Expected: PASS — all existing tests stay green (the floor only fires when `opts.OxygenAtmBiomassFloor` is set, and existing callers pass the zero value).

---

### Task 4: Unit tests for the floor

**Files:**

- Modify: `worlds/biology_test.go` (append four unit tests for `computeBiology` + opts)

- [ ] **Step 1: Append the four unit tests at the end of `worlds/biology_test.go`**

```go
// --- WBH p.128 Optional Rule 1: oxygen-atm biomass floor ---

// fixedRollerForBiomassZero returns a Scripted roller whose RollBiomass
// outcome is well below 1 (rolled-zero baseline) — 2D=2 plus the heavy
// no-atmosphere/zero-hydrographics/young-system DMs typical of test
// fixtures push the modified result deeply negative, which RollBiomass
// clamps up to 0 per the existing implementation.
func fixedRollerForBiomassZero(t *testing.T) roller.Roller {
	t.Helper()
	// Biomass is a 2D roll; use 2 (= 1+1) so the result is below 1
	// before any floor logic. Subsequent dependent rolls (Biocomplexity,
	// etc.) draw from the same scripted sequence — provide enough
	// values to cover them when biomass is elevated to 1.
	return roller.NewScripted(2, 7, 7, 7, 7, 7, 7)
}

func TestComputeBiology_OxygenAtmFloor_Off_RolledZeroStaysZero(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 6} // standard-oxygen
	body.Hydrographics = &Hydrographics{Code: 5}

	bio := computeBiology(fixedRollerForBiomassZero(t), body, 5.0, DetailOpts{})
	if bio.Biomass != 0 {
		t.Errorf("opts off: biomass should stay 0, got %d", bio.Biomass)
	}
}

func TestComputeBiology_OxygenAtmFloor_On_RolledZeroBecomesOne(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 6} // standard-oxygen
	body.Hydrographics = &Hydrographics{Code: 5}

	opts := DetailOpts{OxygenAtmBiomassFloor: true}
	bio := computeBiology(fixedRollerForBiomassZero(t), body, 5.0, opts)
	if bio.Biomass != 1 {
		t.Errorf("opts on, oxygen atm, rolled zero: want biomass 1, got %d", bio.Biomass)
	}
}

func TestComputeBiology_OxygenAtmFloor_On_NonOxygenAtmStaysZero(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 10} // A — Exotic, not oxygen-bearing
	body.Hydrographics = &Hydrographics{Code: 5}

	opts := DetailOpts{OxygenAtmBiomassFloor: true}
	bio := computeBiology(fixedRollerForBiomassZero(t), body, 5.0, opts)
	if bio.Biomass != 0 {
		t.Errorf("opts on, non-oxygen atm: biomass should stay 0, got %d", bio.Biomass)
	}
}

func TestComputeBiology_OxygenAtmFloor_On_RolledPositiveUnchanged(t *testing.T) {
	// Dice that produce a positive biomass without help from the floor.
	// 2D=12, oxygen atm DM-context positive enough that biomass > 1.
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 7} // +1 DM
	body.Temperature = &Temperature{MeanK: 290}  // sweet-spot DM+2

	// 2D=12 (6+6). Successor rolls cover dependent stats.
	r := roller.NewScripted(12, 7, 7, 7, 7, 7, 7)
	opts := DetailOpts{OxygenAtmBiomassFloor: true}
	bio := computeBiology(r, body, 5.0, opts)
	if bio.Biomass < 2 {
		t.Errorf("opts on, oxygen atm, positive roll: biomass should reflect the roll (>=2), got %d", bio.Biomass)
	}
}
```

- [ ] **Step 2: Run the new tests and verify they pass**

Run: `go test ./worlds/ -run "TestComputeBiology_OxygenAtmFloor" -v`

Expected: 4/4 PASS.

If any test fails, inspect the actual `RollBiomass` modifiers (atmosphere DM, hydrographics DM, age DM, temperature DM) for the synthetic body used and adjust the dice / fixture values to match the spec's intent. The intent: tests 1–3 share dice that produce a sub-1 biomass before any floor; test 4 uses dice that produce ≥ 2 biomass naturally.

---

### Task 5: Integration test through DetailSystemWithOpts

**Files:**

- Modify: `worlds/system_detail_test.go` (append a smoke test that drives the full pipeline with opts on)

- [ ] **Step 1: Append integration test after the existing DetailSystem tests**

```go
func TestDetailSystemWithOpts_OxygenAtmFloor(t *testing.T) {
	// Smoke test: drive the full pipeline with the oxygen-atm biomass
	// floor opt-in. Verify the opt-in path threads through to Step 5F
	// and the unfloored zero-baseline path stays untouched.

	// Build a deterministic system with at least one terrestrial body
	// that ends up with an oxygen atmosphere. Reuse the test setup
	// pattern from the existing TestDetailSystem_* tests in this file.
	r := roller.NewSeeded(42)
	sys, err := stars.GenerateSystem(r, stars.GenerateSystemOpts{})
	if err != nil {
		t.Fatal(err)
	}
	sp, err := GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatal(err)
	}

	// Run twice with the SAME seed: once without the floor, once with.
	// Different opt values will produce different roller-consumption
	// patterns past the first biomass clamp, so the post-clamp diff
	// is a smoke-level signal — we only assert that opts on does not
	// crash, completes, and that any oxygen-atm body without rolled
	// life under default opts has biomass ≥ 1 with the floor on.

	rNoOpts := roller.NewSeeded(42)
	sysNo, _ := stars.GenerateSystem(rNoOpts, stars.GenerateSystemOpts{})
	spNo, _ := GenerateSystemPlacement(rNoOpts, sysNo)
	sdNo, err := DetailSystemWithOpts(rNoOpts, sysNo, spNo, IISSClass23Header{}, DetailOpts{})
	if err != nil {
		t.Fatal(err)
	}

	rOpts := roller.NewSeeded(42)
	sysOn, _ := stars.GenerateSystem(rOpts, stars.GenerateSystemOpts{})
	spOn, _ := GenerateSystemPlacement(rOpts, sysOn)
	sdOn, err := DetailSystemWithOpts(rOpts, sysOn, spOn, IISSClass23Header{}, DetailOpts{OxygenAtmBiomassFloor: true})
	if err != nil {
		t.Fatal(err)
	}

	// Sanity: the two SystemDetail outputs are well-formed.
	if len(sdNo.Detailed) == 0 || len(sdOn.Detailed) == 0 {
		t.Fatal("expected non-empty Detailed slices from both runs")
	}

	// Floor invariant: every body whose atmosphere is oxygen-bearing
	// AND has a Biology must have Biomass >= 1 in the opts-on run.
	for i := range sdOn.Detailed {
		dp := &sdOn.Detailed[i]
		if dp.Biology == nil || dp.Atmosphere == nil {
			continue
		}
		if hasOxygenAtmosphere(dp.Atmosphere) && dp.Biology.Biomass < 1 {
			t.Errorf("body %d: oxygen-atm biomass under floor: code=%d biomass=%d",
				i, dp.Atmosphere.Code, dp.Biology.Biomass)
		}
		for j := range dp.Moons {
			m := &dp.Moons[j]
			if m.Biology == nil || m.Atmosphere == nil {
				continue
			}
			if hasOxygenAtmosphere(m.Atmosphere) && m.Biology.Biomass < 1 {
				t.Errorf("body %d moon %d: oxygen-atm biomass under floor: code=%d biomass=%d",
					i, j, m.Atmosphere.Code, m.Biology.Biomass)
			}
		}
	}

	// Reference: under default opts, the same body MAY have biomass 0
	// (the rule is opt-in). Just exercise the path; do not assert
	// strict difference because seed 42 may not produce a sub-1 roll.
	_ = sdNo
}
```

- [ ] **Step 2: Run the integration test and verify it passes**

Run: `go test ./worlds/ -run "TestDetailSystemWithOpts_OxygenAtmFloor" -v`

Expected: PASS.

---

### Task 6: Full quality gate, golden, commit

**Files:** none — verification + commit only.

- [ ] **Step 1: Full test suite**

Run: `task test`

Expected: PASS — including Zed worked-example regressions and the Zed Markdown golden (which use `DetailSystem` with no opts and must be byte-identical).

- [ ] **Step 2: Modernizer + lint + format**

Run: `task check`

Expected: PASS.

- [ ] **Step 3: Commit**

Stage and commit:

```bash
git add worlds/system_detail.go worlds/system_detail_pipeline.go worlds/system_detail_step5f.go worlds/biology.go worlds/biology_test.go worlds/system_detail_test.go docs/history/pass-1-specs/2026-05-09-biology-oxygen-floor-design.md docs/history/pass-1-plans/2026-05-09-biology-oxygen-floor.md
git commit -m "$(cat <<'EOF'
feat(worlds): opt-in oxygen-atm biomass floor per WBH p.128 (partial #12)

Adds DetailOpts and DetailSystemWithOpts; existing DetailSystem becomes
a thin wrapper to preserve all 7 callers. When
DetailOpts.OxygenAtmBiomassFloor is set, computeBiology clamps a rolled
biomass below 1 up to 1 for any body whose Atmosphere.Code is in the
oxygen-bearing set {2-9, D, E} per the book's Optional Rule. Off by
default — the rule is Referee opt-in.

The floor runs after RollBiomass and before the dependent rolls
(Biocomplexity, Biodiversity, Compatibility, Sophont) so an elevated
biomass propagates naturally through the rest of the pipeline.

Closes Optional Rule 1 of #12. Rule 2 (Rare Earth Universe Variant)
mutates atmospheres mid-pipeline and stays open for a separate loop.

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 4: Verify clean tree**

Run: `git status`

Expected: clean.

---

## Self-Review

**Spec coverage:**

- `DetailOpts` exported with `OxygenAtmBiomassFloor bool` — Task 1.
- `DetailSystemWithOpts` exported; `DetailSystem` is a thin wrapper — Task 1.
- Plumbing through `runDetailPipeline` and `runStep5F` — Tasks 2.
- Floor application in `computeBiology` — Task 3.
- `hasOxygenAtmosphere` predicate matches the book's literal list {2-9, D, E} — Task 3.
- Unit tests covering off / on-with-rolled-zero / on-with-non-oxygen-atm / on-with-positive-roll — Task 4.
- Integration smoke through `DetailSystemWithOpts` — Task 5.
- All existing `DetailSystem` callers byte-identical — by construction (wrapper).

**Placeholder scan:** none.

**Type consistency:** the `DetailOpts` parameter threads through the same name (`opts`) at every level.

## Execution

Subagent-driven via `superpowers:subagent-driven-development`. Single subagent dispatch covers Tasks 1–6 (≈80 LOC of changes plus ≈100 LOC of tests).
