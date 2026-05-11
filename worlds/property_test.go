package worlds_test

import (
	"strings"
	"testing"

	"wbh/worlds"
)

// generateForProperty runs the full pipeline at the given seed and
// returns the universe, or nil on Special-Circumstances errors
// (post-stellar primaries, giant-companion-MAO gaps, missing-class-IV
// table cells, peculiar-primary dispatches — all out of pass-2 scope
// per CLAUDE.md). Other errors fail the test.
func generateForProperty(t *testing.T, seed int64) *worlds.Universe {
	t.Helper()
	u, err := worlds.Generate(seed)
	if err != nil {
		if isSpecialCircumstances(err) {
			return nil
		}
		t.Fatalf("seed %d: Generate: %v", seed, err)
	}
	return &u
}

// isSpecialCircumstances classifies errors as Special-Circumstances
// chapter coverage gaps (out of pass-2 scope) vs. real bugs.
func isSpecialCircumstances(err error) bool {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "post-stellar primary"):
		return true
	case strings.Contains(msg, "special primary"):
		return true
	case strings.Contains(msg, "Special-primary"):
		return true
	case strings.Contains(msg, "giant primary requires MAO"):
		return true
	case strings.Contains(msg, "class IV missing"):
		return true
	}
	return false
}

// TestProperty_HZBodyHasClimate per harness.md § Property tests.
// Every body with HZ == true and Kind == BodyTerrestrial has non-nil
// Atmosphere / Hydrographics / Temperature after the pipeline runs
// (climate eligibility per ConvergeClimate). Vacuum / Size-S / Size-R
// bodies are exempt (they don't get climate).
func TestProperty_HZBodyHasClimate(t *testing.T) {
	t.Parallel()
	checked := 0
	for iter := range 1000 {
		seed := int64(iter)
		u := generateForProperty(t, seed)
		if u == nil {
			continue
		}
		for body := range u.AllBodies() {
			if body.Kind != worlds.BodyTerrestrial && body.Kind != worlds.BodyMoon {
				continue
			}
			// Moons inherit HZ from parent; check via the host's HZ flag.
			host := body
			if body.Kind == worlds.BodyMoon && body.Parent != nil {
				host = body.Parent
			}
			if !host.HZ {
				continue
			}
			if body.GGClass != worlds.NotGasGiant {
				continue // GG-cascade moons skip climate
			}
			switch body.SizeCode {
			case "", "0", "R":
				continue // sub-1 sizes don't get climate
			}
			checked++
			if !body.HasAtmosphere() {
				t.Errorf("seed %d: HZ body %s missing Atmosphere", seed, body.Designation)
			}
			if !body.HasHydrographics() {
				t.Errorf("seed %d: HZ body %s missing Hydrographics", seed, body.Designation)
			}
			if !body.HasTemperature() {
				t.Errorf("seed %d: HZ body %s missing Temperature", seed, body.Designation)
			}
		}
	}
	if checked < 100 {
		t.Errorf("only %d HZ bodies checked across 1000 seeds (expected >= 100)", checked)
	}
}

// TestProperty_MoonsHaveBodies per harness.md § Property tests.
// Every Body with non-empty Children has those children processed
// (not silent-zero). Specifically: each child has Kind == BodyMoon,
// non-empty Designation, and a populated Parent pointer.
func TestProperty_MoonsHaveBodies(t *testing.T) {
	t.Parallel()
	checked := 0
	for iter := range 1000 {
		seed := int64(iter)
		u := generateForProperty(t, seed)
		if u == nil {
			continue
		}
		for i := range u.Detail.Bodies {
			body := &u.Detail.Bodies[i]
			for j, child := range body.Children {
				checked++
				if child.Kind != worlds.BodyMoon {
					t.Errorf("seed %d: bodies[%d].Children[%d] Kind = %v, want BodyMoon",
						seed, i, j, child.Kind)
				}
				if child.Designation == "" {
					t.Errorf("seed %d: bodies[%d].Children[%d] missing Designation",
						seed, i, j)
				}
				if child.Parent != body {
					t.Errorf("seed %d: bodies[%d].Children[%d] Parent != &bodies[%d]",
						seed, i, j, i)
				}
			}
		}
	}
	if checked < 100 {
		t.Errorf("only %d moons checked across 1000 seeds (expected >= 100)", checked)
	}
}

// TestProperty_MainworldExists per harness.md § Property tests.
// When the system has at least one terrestrial / moon / belt body,
// the AggregateSystem mainworld pick yields a non-empty designation
// and matching pointer. (pickMainworld's priority-4 fallback returns
// the first terrestrial / moon / belt body in iteration order.)
//
// Also enforces the SystemDetail.Mainworld / MainworldDesignation
// invariant: the two are paired — both empty/nil when no candidates
// exist, both populated otherwise, and Mainworld.Designation must
// equal MainworldDesignation.
func TestProperty_MainworldExists(t *testing.T) {
	t.Parallel()
	for iter := range 1000 {
		seed := int64(iter)
		u := generateForProperty(t, seed)
		if u == nil {
			continue
		}
		hasCandidate := false
		for body := range u.AllBodies() {
			if body.Kind == worlds.BodyTerrestrial || body.Kind == worlds.BodyMoon || body.Kind == worlds.BodyPlanetoidBelt {
				hasCandidate = true
				break
			}
		}
		if hasCandidate && u.Detail.MainworldDesignation == "" {
			t.Errorf("seed %d: system has terrestrial/moon/belt candidates but no MainworldDesignation",
				seed)
		}
		if hasCandidate && u.Detail.Mainworld == nil {
			t.Errorf("seed %d: system has candidates and MainworldDesignation=%q but Mainworld pointer is nil",
				seed, u.Detail.MainworldDesignation)
		}
		if !hasCandidate && u.Detail.Mainworld != nil {
			t.Errorf("seed %d: system has no candidates but Mainworld pointer is non-nil (%s)",
				seed, u.Detail.Mainworld.Designation)
		}
		if !hasCandidate && u.Detail.MainworldDesignation != "" {
			t.Errorf("seed %d: system has no candidates but MainworldDesignation=%q is set",
				seed, u.Detail.MainworldDesignation)
		}
		if u.Detail.Mainworld == nil && u.Detail.MainworldDesignation != "" {
			t.Errorf("seed %d: Mainworld pointer is nil but MainworldDesignation=%q is set",
				seed, u.Detail.MainworldDesignation)
		}
		if u.Detail.Mainworld != nil && u.Detail.Mainworld.Designation != u.Detail.MainworldDesignation {
			t.Errorf("seed %d: Mainworld.Designation=%q does not match MainworldDesignation=%q",
				seed, u.Detail.Mainworld.Designation, u.Detail.MainworldDesignation)
		}
	}
}

// TestProperty_BiomassImpliesAtm per harness.md § Property tests.
// Every body with Biology.Biomass > 0 has non-nil Atmosphere — the
// biology pass requires atmosphere as a precondition.
func TestProperty_BiomassImpliesAtm(t *testing.T) {
	t.Parallel()
	checked := 0
	for iter := range 1000 {
		seed := int64(iter)
		u := generateForProperty(t, seed)
		if u == nil {
			continue
		}
		for body := range u.AllBodies() {
			if !body.HasBiology() {
				continue
			}
			if body.Biology.Biomass == 0 {
				continue
			}
			checked++
			if !body.HasAtmosphere() {
				t.Errorf("seed %d: body %s has Biomass=%d but no Atmosphere",
					seed, body.Designation, body.Biology.Biomass)
			}
		}
	}
	if checked < 10 {
		t.Logf("only %d biomass-bearing bodies seen across 1000 seeds (low rate, but Property invariant holds)", checked)
	}
}

// (TestProperty_HabitabilityImpliesAtm was considered but is incorrect
// per WBH p.132 — vacuum worlds with atm code 0 can have positive
// Habitability ratings for size / temperature / gravity contributions.
// The natural-language intuition "habitable → has atmosphere" does
// not hold.)

// TestProperty_GGHasMass — every gas giant has MassEarth > 0.
// Pass-2 stage-2 RollGasGiantSize always returns a positive mass per
// WBH p.55.
func TestProperty_GGHasMass(t *testing.T) {
	t.Parallel()
	checked := 0
	for iter := range 1000 {
		seed := int64(iter)
		u := generateForProperty(t, seed)
		if u == nil {
			continue
		}
		for body := range u.AllBodies() {
			if body.Kind != worlds.BodyGasGiant {
				continue
			}
			checked++
			if body.MassEarth <= 0 {
				t.Errorf("seed %d: GG %s has MassEarth = %v", seed, body.Designation, body.MassEarth)
			}
		}
	}
	if checked < 50 {
		t.Errorf("only %d gas giants seen across 1000 seeds (expected >= 50)", checked)
	}
}

// TestProperty_MoonsHaveOrbitPD — every retained moon has its OrbitPD
// populated by Stage 3's RefineMoons walk. Anti-pattern A.1 sentinel
// at a finer granularity than MoonsHaveBodies — catches "moon-path
// silent-zero" specifically in the orbital-refinement step.
func TestProperty_MoonsHaveOrbitPD(t *testing.T) {
	t.Parallel()
	checked := 0
	for iter := range 1000 {
		seed := int64(iter)
		u := generateForProperty(t, seed)
		if u == nil {
			continue
		}
		for i := range u.Detail.Bodies {
			body := &u.Detail.Bodies[i]
			for j, child := range body.Children {
				checked++
				if child.OrbitPD <= 0 {
					t.Errorf("seed %d: bodies[%d].Children[%d] (%s) OrbitPD = %v",
						seed, i, j, child.Designation, child.OrbitPD)
				}
			}
		}
	}
	if checked < 50 {
		t.Errorf("only %d moons checked across 1000 seeds (expected >= 50)", checked)
	}
}

// TestProperty_ScaleHeightPositive — every body with Atmosphere and
// Physical has Atmosphere.ScaleHeight > 0. DeriveScaleHeight reads
// post-TSS MeanK and gravity; either degenerate input drives it to 0.
func TestProperty_ScaleHeightPositive(t *testing.T) {
	t.Parallel()
	checked := 0
	for iter := range 1000 {
		seed := int64(iter)
		u := generateForProperty(t, seed)
		if u == nil {
			continue
		}
		for body := range u.AllBodies() {
			if !body.HasAtmosphere() || !body.HasPhysical() {
				continue
			}
			checked++
			if body.Atmosphere.ScaleHeight <= 0 {
				t.Errorf("seed %d: body %s has Atmosphere + Physical but ScaleHeight = %v",
					seed, body.Designation, body.Atmosphere.ScaleHeight)
			}
		}
	}
	if checked < 50 {
		t.Errorf("only %d bodies with both Atm+Physical seen across 1000 seeds (expected >= 50)", checked)
	}
}

// TestProperty_ConvergenceCompletes per harness.md § Property tests.
// Generate must complete (or fail with the documented Special-
// Circumstances primary error) for every seed in 0..999. No
// convergence-overflow, panic, or stall.
func TestProperty_ConvergenceCompletes(t *testing.T) {
	t.Parallel()
	completed := 0
	for iter := range 1000 {
		seed := int64(iter)
		_, err := worlds.Generate(seed)
		if err == nil {
			completed++
			continue
		}
		if isSpecialCircumstances(err) {
			continue
		}
		t.Errorf("seed %d: unexpected error: %v", seed, err)
	}
	if completed < 100 {
		t.Errorf("only %d / 1000 seeds completed without Special-Circumstances; expected >= 100",
			completed)
	}
}
