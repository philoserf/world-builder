package worlds_test

import (
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/worlds"
)

// TestZed_ApplyStage7 exercises the Stage-7 geology orchestrator end-
// to-end. Seeded shape-invariant per spike-findings § 2.
func TestZed_ApplyStage7(t *testing.T) {
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
		for _, step := range []func(roller.Roller, *worlds.Universe) error{
			worlds.ApplyDetailFrontEnd,
			worlds.ApplyBodyPhysical,
			worlds.ApplyBeltDetails,
			worlds.ApplyMoonRefinement,
			worlds.ApplyRotationTilt,
			worlds.ApplyClimate,
			worlds.ApplyTaintTypology,
			worlds.ApplySurfaceDistribution,
			worlds.ApplyGeology,
		} {
			if err := step(r, u); err != nil {
				t.Fatalf("seed %d: step: %v", seed, err)
			}
		}

		// Every non-empty non-belt body has Geology populated. Per
		// anti-pattern A.1, every Child does too.
		for i, body := range u.Detail.Bodies {
			if body.Kind == worlds.BodyEmpty || body.SizeCode == "0" {
				continue
			}
			if !body.HasGeology() {
				t.Errorf("seed %d: bodies[%d] (%s) missing Geology",
					seed, i, body.Designation)
			}
			for j, child := range body.Children {
				if !child.HasGeology() {
					t.Errorf("seed %d: bodies[%d].Children[%d] (%s) missing Geology",
						seed, i, j, child.Designation)
				}
			}
		}
	}
}
