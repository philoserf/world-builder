package worlds

import (
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestRunStep5G_TerrestrialPopulatesHabitability(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "5"
	dp.Designation = "Aab III"
	dp.Atmosphere = &Atmosphere{Code: 6}
	dp.Hydrographics = &Hydrographics{Code: 6}
	dp.Physical = &BodyPhysical{Gravity: 1.0}
	dp.Temperature = &Temperature{MeanK: 290, HighK: 310, LowK: 270}

	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability == nil {
		t.Fatal("Habitability is nil")
	}
	if detailed[0].Habitability.Rating < 0 || detailed[0].Habitability.Rating > 12 {
		t.Errorf("Rating %d out of [0, 12]", detailed[0].Habitability.Rating)
	}
}

func TestRunStep5G_GasGiant_NoHabitability(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyGasGiant
	dp.GGClass = GasGiantSmall
	dp.Designation = "Aab IV"
	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability != nil {
		t.Error("GG should not get Habitability")
	}
}

func TestRunStep5G_BeltSize0_NoHabitability(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyPlanetoidBelt
	dp.SizeCode = "0"
	dp.Designation = "Aab Belt"
	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability != nil {
		t.Error("Belt should not get Habitability")
	}
}

func TestRunStep5G_BodyEmpty_NoOp(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyEmpty
	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability != nil {
		t.Error("Empty body should not get Habitability")
	}
}

func TestRunStep5G_MoonRecursion(t *testing.T) {
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "8"
	dp.Designation = "Aab III"
	dp.Atmosphere = &Atmosphere{Code: 6}
	dp.Hydrographics = &Hydrographics{Code: 6}
	dp.Physical = &BodyPhysical{Gravity: 1.0}
	dp.Temperature = &Temperature{MeanK: 290, HighK: 310, LowK: 270}
	dp.Moons = []Moon{
		{
			Designation:   "Aab III a",
			SizeCode:      "5",
			Atmosphere:    &Atmosphere{Code: 6},
			Hydrographics: &Hydrographics{Code: 5},
			Physical:      &BodyPhysical{Gravity: 0.7},
			Temperature:   &Temperature{MeanK: 290, HighK: 310, LowK: 270},
		},
	}

	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability == nil {
		t.Fatal("Parent Habitability is nil")
	}
	if detailed[0].Moons[0].Habitability == nil {
		t.Fatal("Moon Habitability is nil")
	}
}

func TestRunStep5G_MoonOfGGParent_GetsHabitability(t *testing.T) {
	// Critical regression test: Zed Prime is a moon of Aab IV (a Gas Giant)
	// and is the canonical Habitability-7 mainworld per WBH p.141. The
	// previous implementation used 'continue' on the parent-applies check,
	// skipping the entire moon loop when the parent was a GG, silently
	// denying Habitability to such moons.
	dp := DetailedPlacement{}
	dp.Body = BodyGasGiant
	dp.GGClass = GasGiantSmall
	dp.Designation = "Aab IV"
	dp.Moons = []Moon{
		{
			Designation:   "Aab IV d", // Zed Prime
			SizeCode:      "5",
			Atmosphere:    &Atmosphere{Code: 6},
			Hydrographics: &Hydrographics{Code: 6},
			Physical:      &BodyPhysical{Gravity: 0.66},
			Temperature:   &Temperature{MeanK: 300, HighK: 346, LowK: 262},
		},
	}

	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	// Parent (GG) should NOT have Habitability.
	if detailed[0].Habitability != nil {
		t.Error("GG parent should not have Habitability")
	}
	// Moon should have Habitability — this is the bug being fixed.
	if detailed[0].Moons[0].Habitability == nil {
		t.Fatal("Moon of GG parent should have Habitability (Zed Prime case)")
	}
	// Zed Prime per WBH p.141 → Habitability 7.
	// DMs: Size 5=0, Atm 6=0, Hydro 6=0, TidalLock nil=0, HighK 346>323=-2,
	// MeanK 300 (273-303)=0, LowK 262>200=0, Gravity 0.66∈[0.4,0.7)=-1.
	// 10 + 0 + 0 + 0 + 0 + (-2) + 0 + (-1) = 7.
	if detailed[0].Moons[0].Habitability.Rating != 7 {
		t.Errorf("Zed Prime moon: got Habitability %d, want 7", detailed[0].Moons[0].Habitability.Rating)
	}
}

func TestRunStep5G_VacuumWorld_HasHabitability(t *testing.T) {
	// Atm nil → still gets Habitability (vacuum atm 0 DM-8 applied).
	dp := DetailedPlacement{}
	dp.Body = BodyTerrestrial
	dp.SizeCode = "5"
	dp.Designation = "Aab III"
	dp.Atmosphere = nil // vacuum
	dp.Physical = &BodyPhysical{Gravity: 1.0}

	r := roller.NewScripted()
	detailed := []DetailedPlacement{dp}
	if err := runStep5G(r, detailed, stars.System{}); err != nil {
		t.Fatal(err)
	}
	if detailed[0].Habitability == nil {
		t.Fatal("Vacuum world should still have Habitability")
	}
	// 10 - 8 (vacuum) - 4 (nil hydro) = -2 → clamp 0
	if detailed[0].Habitability.Rating != 0 {
		t.Errorf("Rating: got %d, want 0 (vacuum + nil hydro → 0 clamped)", detailed[0].Habitability.Rating)
	}
}
