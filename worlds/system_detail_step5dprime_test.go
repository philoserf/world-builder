package worlds

import (
	"strings"
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
	if len(detailed[0].Atmosphere.InsidiousHazards) == 0 {
		t.Errorf("expected InsidiousHazards on atm C, got empty")
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
	if len(detailed[0].Atmosphere.InsidiousHazards) > 0 {
		t.Errorf("got InsidiousHazards on atm B, want empty")
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
	// End-to-end test of computeBodyTaints → RollInsidiousHazards for
	// atm-C bodies. Verifies both the WBH p.89 DM+2 ("Extremely Dense"
	// subtypes C/D/E) and the WBH p.90 footnote (subtypes D/E grant an
	// automatic T plus a rolled hazard).
	//
	// RollAllTaints for atm 12 with no pre-seed consumes 3 rolls
	// (subtype 2D, severity 2D, persistence 2D) before the loop breaks
	// (rawRoll != 10). Then RollInsidiousHazards consumes 1 roll for
	// the rolled hazard (the auto-T for D/E is free).
	//
	// Subtype roll 7 → 2D=7+0=7 → taintSubtypeFromTotal=G → atm 12 is
	// outside 4-9 so any L/H would be suppressed to G; G is already G.
	// No ppO2 adjust path. Severity 7 and persistence 7 don't matter
	// for this test — they're stable across the cases.
	//
	// Per-case hazard rolls:
	//   subtype "D", hazardRoll=4 → DM+2 → 6 → G; auto-T prepended → "TG"
	//   subtype "E", hazardRoll=2 → DM+2 → 4 → B; auto-T prepended → "TB"
	//   subtype "6", hazardRoll=4 → DM 0  → 4 → B; no auto-T          → "B"
	cases := []struct {
		name       string
		subtype    string
		hazardRoll int
		// want is the concatenated hazard codes (e.g. "TG" = auto-T + rolled-G).
		want    string
		wantLen int
	}{
		// Subtype D fires the WBH p.90 footnote: auto-T plus rolled-G
		// (2D=4 + DM+2 = 6 → G). 2 hazards.
		{"D triggers DM+2 with auto-T", "D", 4, "TG", 2},
		// Subtype E also fires the WBH p.90 footnote: auto-T plus rolled-B
		// (2D=2 + DM+2 = 4 → B). 2 hazards. Distinct integration path
		// from D since it's a separate literal in RollInsidiousHazards.
		{"E triggers DM+2 with auto-T", "E", 2, "TB", 2},
		// Subtype 6: 1 rolled hazard, no DM, no auto-T.
		// 2D=4 + DM+0 = 4 → B.
		{"non-CDE no DM no auto-T", "6", 4, "B", 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := roller.NewScripted(7, 7, 7, c.hazardRoll)
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
			hazards := detailed[0].Atmosphere.InsidiousHazards
			if len(hazards) == 0 {
				t.Fatalf("subtype %q: expected InsidiousHazards, got empty", c.subtype)
			}
			if len(hazards) != c.wantLen {
				t.Errorf("subtype %q: got %d hazards %+v, want %d", c.subtype, len(hazards), hazards, c.wantLen)
			}
			var gotCodes strings.Builder
			for _, h := range hazards {
				gotCodes.WriteString(h.Code)
			}
			if gotCodes.String() != c.want {
				t.Errorf("subtype %q: got hazard codes %q, want %q", c.subtype, gotCodes.String(), c.want)
			}
		})
	}
}
