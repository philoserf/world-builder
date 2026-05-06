package worlds

import "testing"

func TestComputeHabitability_BaselineNoDMs(t *testing.T) {
	// Size 5, Atm 6 (no DM), Hydro 5 (no DM), no tidal lock, no temp/gravity DMs.
	// Result: 10 + 0 = 10.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	if got.Rating != 10 {
		t.Errorf("got %d, want 10", got.Rating)
	}
}

func TestComputeHabitability_SmallSize_DMMinus1(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "4"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	if got.Rating != 9 {
		t.Errorf("got %d, want 9 (Size 4 DM-1)", got.Rating)
	}
}

func TestComputeHabitability_LargeSize_DMPlus1(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "9"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	if got.Rating != 11 {
		t.Errorf("got %d, want 11 (Size 9 DM+1)", got.Rating)
	}
}

func TestComputeHabitability_AtmVacuum_DMMinus8(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 0}
	body.Hydrographics = &Hydrographics{Code: 0}
	got := ComputeHabitability(body)
	// 10 + (-8) + (-4 hydro 0) = -2 → clamp 0
	if got.Rating != 0 {
		t.Errorf("got %d, want 0 (atm 0 vacuum + hydro 0)", got.Rating)
	}
}

func TestComputeHabitability_NilAtmosphere_TreatedAsAtm0(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = nil
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	// 10 + (-8) + 0 = 2
	if got.Rating != 2 {
		t.Errorf("got %d, want 2 (nil atm → DM-8)", got.Rating)
	}
}

func TestComputeHabitability_AtmHostile_DMMinus10(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 11} // B
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	// 10 + (-10) = 0
	if got.Rating != 0 {
		t.Errorf("got %d, want 0 (atm B hostile)", got.Rating)
	}
}

func TestComputeHabitability_AtmVeryHostile_DMMinus12(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 12} // C
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	// 10 + (-12) = -2 → clamp 0
	if got.Rating != 0 {
		t.Errorf("got %d, want 0 (atm C very hostile)", got.Rating)
	}
}

func TestComputeHabitability_HydroDesert_DMMinus2(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 2}
	got := ComputeHabitability(body)
	// 10 + (-2) = 8
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (Hydro 2 desert)", got.Rating)
	}
}

func TestComputeHabitability_HydroFull_DMMinus2(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 10} // A
	got := ComputeHabitability(body)
	// 10 + (-2) = 8
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (Hydro A very-full)", got.Rating)
	}
}

func TestComputeHabitability_TidalLock1to1_DMMinus2(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.TidalLock = &TidalLock{
		Case:           TidalLockCasePlanetToStar,
		LockRatio:      "1:1",
		IsTwilightZone: true,
	}
	got := ComputeHabitability(body)
	// 10 + (-2) = 8
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (1:1 solar lock)", got.Rating)
	}
}

func TestComputeHabitability_TidalLockNot1to1_NoDM(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.TidalLock = &TidalLock{
		Case:           TidalLockCasePlanetToStar,
		LockRatio:      "3:2", // not 1:1
		IsTwilightZone: false,
	}
	got := ComputeHabitability(body)
	// 10 + 0 = 10
	if got.Rating != 10 {
		t.Errorf("got %d, want 10 (3:2 lock, no DM)", got.Rating)
	}
}

func TestComputeHabitability_NilBody_ZeroRating(t *testing.T) {
	got := ComputeHabitability(nil)
	if got.Rating != 0 {
		t.Errorf("got %d, want 0 (nil body)", got.Rating)
	}
}
