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
