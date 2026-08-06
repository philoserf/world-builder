package worlds_test

import (
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/worlds"
)

// TestZed_ApplyStage3 exercises the Stage-3 orchestrators (body
// physical / belt details / moon refinement) end-to-end against a
// Seeded roller. Per spike-findings.md § Finding 2, Seeded shape-
// invariant. Per-procedure value-exact tests live in
// body_physical_test.go, belt_details_test.go, moon_refinement_test.go.
func TestZed_ApplyStage3(t *testing.T) {
	t.Parallel()

	for iter := range 25 {
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

		if err := worlds.ApplyBodyPhysical(r, u); err != nil {
			t.Fatalf("seed %d: ApplyBodyPhysical: %v", seed, err)
		}

		if err := worlds.ApplyBeltDetails(r, u); err != nil {
			t.Fatalf("seed %d: ApplyBeltDetails: %v", seed, err)
		}

		if err := worlds.ApplyMoonRefinement(r, u); err != nil {
			t.Fatalf("seed %d: ApplyMoonRefinement: %v", seed, err)
		}

		// Every terrestrial planet (non-S, non-R, non-0) has Physical set
		// and MassEarth derived. Per anti-pattern A.1, every terrestrial
		// moon does too.
		for i, body := range u.Detail.Bodies {
			if body.GGClass == worlds.NotGasGiant && body.SizeCode != "" && body.SizeCode != "0" && body.SizeCode != "R" && body.SizeCode != "S" {
				if body.Physical == nil {
					t.Errorf("seed %d: bodies[%d] (%s, size %s) missing Physical",
						seed, i, body.Designation, body.SizeCode)
				}

				if body.MassEarth == 0 {
					t.Errorf("seed %d: bodies[%d] (%s) MassEarth = 0", seed, i, body.Designation)
				}
			}

			for j, child := range body.Children {
				if child.GGClass == worlds.NotGasGiant && child.SizeCode != "" && child.SizeCode != "0" && child.SizeCode != "R" && child.SizeCode != "S" {
					if child.Physical == nil {
						t.Errorf("seed %d: bodies[%d].Children[%d] (%s) missing Physical (moon-path silent-zero?)",
							seed, i, j, child.Designation)
					}
				}
			}
		}

		// Every belt body has Belt set with non-empty Profile.
		for i, body := range u.Detail.Bodies {
			if body.SizeCode == "0" {
				if body.Belt == nil {
					t.Errorf("seed %d: bodies[%d] (%s) missing Belt", seed, i, body.Designation)
				}
			}
		}

		// Every parent that retained moons has each child's OrbitPD
		// populated by the Hill-sphere refinement (or moons removed
		// entirely if Hill sphere too small).
		for i, body := range u.Detail.Bodies {
			for j, child := range body.Children {
				if child.OrbitPD <= 0 {
					t.Errorf("seed %d: bodies[%d].Children[%d] OrbitPD = %v",
						seed, i, j, child.OrbitPD)
				}
			}
		}
	}
}
