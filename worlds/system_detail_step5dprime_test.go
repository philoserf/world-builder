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
