package worlds_test

import (
	"testing"

	"wbh/roller"
	"wbh/worlds"
)

// TestZed_ApplyStage5 exercises the Stage-5 climate orchestrator end-
// to-end. Seeded shape-invariant per spike-findings § 2.
func TestZed_ApplyStage5(t *testing.T) {
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
		if err := worlds.ApplyClimate(r, u); err != nil {
			t.Fatalf("seed %d: ApplyClimate: %v", seed, err)
		}

		// Every HZ terrestrial has Atmosphere, Hydrographics, Temperature
		// populated. Per anti-pattern A.1, every HZ-planet moon does too.
		for i, body := range u.Detail.Bodies {
			if body.Kind != worlds.BodyTerrestrial {
				continue
			}
			if !body.HZ {
				continue
			}
			if body.SizeCode == "" || body.SizeCode == "0" || body.SizeCode == "R" {
				continue
			}
			if !body.HasAtmosphere() {
				t.Errorf("seed %d: HZ terrestrial bodies[%d] (%s) missing Atmosphere",
					seed, i, body.Designation)
			}
			if !body.HasHydrographics() {
				t.Errorf("seed %d: HZ terrestrial bodies[%d] (%s) missing Hydrographics",
					seed, i, body.Designation)
			}
			if !body.HasTemperature() {
				t.Errorf("seed %d: HZ terrestrial bodies[%d] (%s) missing Temperature",
					seed, i, body.Designation)
			}
		}
	}
}
