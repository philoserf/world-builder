package worlds

import (
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestRunStep5DPrime_TerrestrialBodyGetsTaints(t *testing.T) {
	r := roller.NewSeeded(42)
	sys := stars.System{Primary: stars.Star{AgeGyr: 5.0}}
	detailed := []DetailedPlacement{{
		Placement:  Placement{Body: BodyTerrestrial},
		HZ:         true,
		SizeCode:   "8",
		Atmosphere: &Atmosphere{Code: 7, Pressure: 1.0, OxygenPartialPressure: 0.21},
	}}
	if err := runStep5DPrime(r, detailed, sys); err != nil {
		t.Fatalf("runStep5DPrime: %v", err)
	}
	if detailed[0].Atmosphere.Taints == nil {
		t.Errorf("expected at least one taint on atm 7, got nil")
	}
}

func TestRunStep5DPrime_AtmCGetsHazard(t *testing.T) {
	r := roller.NewSeeded(42)
	sys := stars.System{Primary: stars.Star{AgeGyr: 5.0}}
	detailed := []DetailedPlacement{{
		Placement:  Placement{Body: BodyTerrestrial},
		HZ:         true,
		SizeCode:   "8",
		Atmosphere: &Atmosphere{Code: 12, Subtype: "6", Pressure: 1.0},
	}}
	if err := runStep5DPrime(r, detailed, sys); err != nil {
		t.Fatalf("runStep5DPrime: %v", err)
	}
	if detailed[0].Atmosphere.InsidiousHazard == nil {
		t.Errorf("expected InsidiousHazard on atm C, got nil")
	}
}

func TestRunStep5DPrime_NonAtmCNoHazard(t *testing.T) {
	r := roller.NewSeeded(42)
	sys := stars.System{Primary: stars.Star{AgeGyr: 5.0}}
	detailed := []DetailedPlacement{{
		Placement:  Placement{Body: BodyTerrestrial},
		HZ:         true,
		SizeCode:   "8",
		Atmosphere: &Atmosphere{Code: 11, Pressure: 1.0}, // Corrosive, not Insidious
	}}
	if err := runStep5DPrime(r, detailed, sys); err != nil {
		t.Fatalf("runStep5DPrime: %v", err)
	}
	if detailed[0].Atmosphere.InsidiousHazard != nil {
		t.Errorf("got InsidiousHazard on atm B, want nil")
	}
}

func TestRunStep5DPrime_MoonsVisited(t *testing.T) {
	// MEMORY anti-pattern check: moons must run through 5D-prime.
	r := roller.NewSeeded(42)
	sys := stars.System{Primary: stars.Star{AgeGyr: 5.0}}
	detailed := []DetailedPlacement{{
		Placement:  Placement{Body: BodyTerrestrial},
		HZ:         true,
		SizeCode:   "8",
		Atmosphere: &Atmosphere{Code: 7, Pressure: 1.0, OxygenPartialPressure: 0.21},
		Moons: []Moon{{
			SizeCode:   "5",
			Atmosphere: &Atmosphere{Code: 4, Pressure: 0.5, OxygenPartialPressure: 0.20},
		}},
	}}
	if err := runStep5DPrime(r, detailed, sys); err != nil {
		t.Fatalf("runStep5DPrime: %v", err)
	}
	if detailed[0].Moons[0].Atmosphere.Taints == nil {
		t.Errorf("expected moon to have taints; runStep5DPrime did not visit moons")
	}
}

func TestRunStep5DPrime_GasGiantSkipped(t *testing.T) {
	r := roller.NewSeeded(42)
	sys := stars.System{Primary: stars.Star{AgeGyr: 5.0}}
	detailed := []DetailedPlacement{{
		Placement: Placement{Body: BodyGasGiant},
	}}
	if err := runStep5DPrime(r, detailed, sys); err != nil {
		t.Fatalf("runStep5DPrime: %v", err)
	}
	// No assertion needed — just no panic; atmosphere stays nil.
}

func TestRunStep5DPrime_PromotionAppliesBeforeRoll(t *testing.T) {
	// Atm 6, ppO2=0.05 → promote to 7 with L pre-seeded.
	r := roller.NewSeeded(42)
	sys := stars.System{Primary: stars.Star{AgeGyr: 5.0}}
	detailed := []DetailedPlacement{{
		Placement:  Placement{Body: BodyTerrestrial},
		HZ:         true,
		SizeCode:   "8",
		Atmosphere: &Atmosphere{Code: 6, Pressure: 1.0, OxygenPartialPressure: 0.05},
	}}
	if err := runStep5DPrime(r, detailed, sys); err != nil {
		t.Fatalf("runStep5DPrime: %v", err)
	}
	if detailed[0].Atmosphere.Code != 7 {
		t.Errorf("got promoted atm code %d, want 7", detailed[0].Atmosphere.Code)
	}
	if !HasTaintCode(detailed[0].Atmosphere.Taints, "L") {
		t.Errorf("expected L taint after promotion; got %+v", detailed[0].Atmosphere.Taints)
	}
}

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
