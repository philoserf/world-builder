package worlds_test

import (
	"testing"

	"wbh/roller"
	"wbh/worlds"
)

// TestZed_ApplyStage4 exercises the Stage-4 orchestrator (rotation /
// tilt / tide) end-to-end. Seeded shape-invariant per spike-findings § 2.
func TestZed_ApplyStage4(t *testing.T) {
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
		if err := worlds.ApplyRotationTilt(r, u); err != nil {
			t.Fatalf("seed %d: ApplyRotationTilt: %v", seed, err)
		}

		// Every non-empty body has DayLength + AxialTilt + TidalEffects
		// populated. Per anti-pattern A.1, every child does too.
		for i, body := range u.Detail.Bodies {
			if body.Kind == worlds.BodyEmpty {
				continue
			}
			if !body.HasDayLength() {
				t.Errorf("seed %d: bodies[%d] (%s) missing DayLength", seed, i, body.Designation)
			}
			if !body.HasAxialTilt() {
				t.Errorf("seed %d: bodies[%d] (%s) missing AxialTilt", seed, i, body.Designation)
			}
			if !body.HasTidalEffects() {
				t.Errorf("seed %d: bodies[%d] (%s) missing TidalEffects", seed, i, body.Designation)
			}
			for j, child := range body.Children {
				if !child.HasDayLength() {
					t.Errorf("seed %d: bodies[%d].Children[%d] (%s) missing DayLength (moon-path silent-zero?)",
						seed, i, j, child.Designation)
				}
				if !child.HasAxialTilt() {
					t.Errorf("seed %d: bodies[%d].Children[%d] (%s) missing AxialTilt",
						seed, i, j, child.Designation)
				}
				if !child.HasTidalEffects() {
					t.Errorf("seed %d: bodies[%d].Children[%d] (%s) missing TidalEffects",
						seed, i, j, child.Designation)
				}
			}
		}
	}
}
