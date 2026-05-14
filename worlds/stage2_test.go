package worlds_test

import (
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/worlds"
)

// TestZed_ApplyDetailFrontEnd exercises the Stage-2 façade end-to-end
// against a Seeded roller. Per docs/history/spike-findings.md § Finding 2,
// this is a Seeded shape-invariant fixture (not a Scripted value-exact
// gold script) — the dice consumption order for Stage 2 is determined
// by Generate / GenerateSystemPlacement / ApplyDetailFrontEnd as a
// composed pipeline, and shape-only assertions are robust to
// procedure-internal reordering.
//
// Per-procedure value-exact tests live in moons_test.go,
// sizing_terrestrial_test.go, sizing_gasgiant_test.go, period_test.go,
// and designations_test.go. This fixture verifies the orchestrator
// wires them together correctly.
func TestZed_ApplyDetailFrontEnd(t *testing.T) {
	t.Parallel()

	for iter := range 50 {
		seed := int64(iter)
		r := roller.NewSeeded(seed)
		sys := composeZed()

		sp, err := worlds.GenerateSystemPlacement(r, sys)
		if err != nil {
			t.Fatalf("seed %d: GenerateSystemPlacement: %v", seed, err)
		}

		u := &worlds.Universe{System: sys, Placement: sp}
		if err := worlds.ApplyDetailFrontEnd(r, u); err != nil {
			t.Fatalf("seed %d: ApplyDetailFrontEnd: %v", seed, err)
		}

		if len(u.Detail.Bodies) != len(sp.Placements) {
			t.Errorf("seed %d: len(Bodies) = %d, want %d (matches Placements)",
				seed, len(u.Detail.Bodies), len(sp.Placements))
		}

		// Sub-stage 1: every Body inherits Group / Orbit / Eccentricity / Kind
		// from its corresponding Placement.
		for i, body := range u.Detail.Bodies {
			p := sp.Placements[i]
			if body.Kind != p.Body {
				t.Errorf("seed %d: bodies[%d].Kind = %v, want %v", seed, i, body.Kind, p.Body)
			}
			if body.Orbit != p.Orbit {
				t.Errorf("seed %d: bodies[%d].Orbit = %v, want %v", seed, i, body.Orbit, p.Orbit)
			}
		}

		// Sub-stage 2: terrestrials and gas giants have populated sizing
		// fields; belts and empties do not.
		for i, body := range u.Detail.Bodies {
			switch body.Kind {
			case worlds.BodyTerrestrial:
				if body.SizeCode == "" {
					t.Errorf("seed %d: terrestrial bodies[%d] has empty SizeCode", seed, i)
				}
			case worlds.BodyGasGiant:
				if body.GGClass == worlds.NotGasGiant {
					t.Errorf("seed %d: gas-giant bodies[%d] has NotGasGiant", seed, i)
				}
				if body.MassEarth <= 0 {
					t.Errorf("seed %d: gas-giant bodies[%d] has MassEarth %v", seed, i, body.MassEarth)
				}
			}
		}

		// Sub-stage 3: every non-empty non-belt body has its Children
		// slice populated (possibly empty if dice say no moons; never nil
		// if count > 0). Anti-pattern A.1 guard: every child has
		// Kind == BodyMoon and Parent set.
		for i, body := range u.Detail.Bodies {
			for j, child := range body.Children {
				if child.Kind != worlds.BodyMoon {
					t.Errorf("seed %d: bodies[%d].Children[%d].Kind = %v, want BodyMoon",
						seed, i, j, child.Kind)
				}
				if child.Parent != &u.Detail.Bodies[i] {
					t.Errorf("seed %d: bodies[%d].Children[%d].Parent != &bodies[%d]",
						seed, i, j, i)
				}
			}
		}

		// Sub-stage 4: every non-empty body has a Designation.
		for i, body := range u.Detail.Bodies {
			if body.Kind != worlds.BodyEmpty && body.Designation == "" {
				t.Errorf("seed %d: bodies[%d] (kind=%v) has empty Designation",
					seed, i, body.Kind)
			}
			for j, child := range body.Children {
				if child.Designation == "" {
					t.Errorf("seed %d: bodies[%d].Children[%d] has empty Designation",
						seed, i, j)
				}
			}
		}

		// Sub-stage 5: every non-empty body has a non-zero Period
		// (Hours and Years populated by Kepler).
		for i, body := range u.Detail.Bodies {
			if body.Kind == worlds.BodyEmpty {
				continue
			}
			if body.Period.Hours <= 0 {
				t.Errorf("seed %d: bodies[%d] (%s) has Period.Hours = %v",
					seed, i, body.Designation, body.Period.Hours)
			}
		}

		// Sub-stage 6: HZ tagging — at least one HZ body is plausible
		// for Zed (Aab IV is HZCO ≈ 3.3, and Stage 1 places ~17 bodies
		// across 3 groups). Don't assert specific HZ count; just that
		// HZ flag is correctly nil/false for empties.
		for i, body := range u.Detail.Bodies {
			if body.Kind == worlds.BodyEmpty && body.HZ {
				t.Errorf("seed %d: empty bodies[%d] has HZ = true", seed, i)
			}
		}
	}
}

// TestSol_ApplyDetailFrontEnd is the single-primary smoke. Mirrors the
// Zed test with Sol's simpler topology.
func TestSol_ApplyDetailFrontEnd(t *testing.T) {
	t.Parallel()

	for iter := range 25 {
		seed := int64(iter)
		r := roller.NewSeeded(seed)
		sys := composeSol()

		sp, err := worlds.GenerateSystemPlacement(r, sys)
		if err != nil {
			t.Fatalf("seed %d: GenerateSystemPlacement: %v", seed, err)
		}

		u := &worlds.Universe{System: sys, Placement: sp}
		if err := worlds.ApplyDetailFrontEnd(r, u); err != nil {
			t.Fatalf("seed %d: ApplyDetailFrontEnd: %v", seed, err)
		}

		if len(u.Detail.Bodies) == 0 {
			t.Fatalf("seed %d: Bodies empty", seed)
		}

		// Sol is single-primary; every non-empty body lives in group "A".
		for i, body := range u.Detail.Bodies {
			if body.Kind == worlds.BodyEmpty {
				continue
			}
			if body.Group.Designation != "A" {
				t.Errorf("seed %d: bodies[%d].Group.Designation = %q, want \"A\"",
					seed, i, body.Group.Designation)
			}
		}
	}
}
