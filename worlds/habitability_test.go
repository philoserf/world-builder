package worlds

import "testing"

func TestComputeHabitability_BaselineNoDMs(t *testing.T) {
	// Size 5, Atm 6 (no DM), Hydro 5 (no DM), no tidal lock, no temp/gravity DMs.
	// Result: 10 + 0 = 10.
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	got := ComputeHabitability(body)
	if got.Rating != 10 {
		t.Errorf("got %d, want 10", got.Rating)
	}
}

func TestComputeHabitability_SmallSize_DMMinus1(t *testing.T) {
	body := &Body{}
	body.SizeCode = "4"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 1.0} // neutral gravity; isolates Size DM
	got := ComputeHabitability(body)
	if got.Rating != 9 {
		t.Errorf("got %d, want 9 (Size 4 DM-1)", got.Rating)
	}
}

func TestComputeHabitability_LargeSize_DMPlus1(t *testing.T) {
	body := &Body{}
	body.SizeCode = "9"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 1.0} // neutral gravity; isolates Size DM
	got := ComputeHabitability(body)
	if got.Rating != 11 {
		t.Errorf("got %d, want 11 (Size 9 DM+1)", got.Rating)
	}
}

func TestComputeHabitability_AtmVacuum_DMMinus8(t *testing.T) {
	body := &Body{}
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
	body := &Body{}
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
	body := &Body{}
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
	body := &Body{}
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
	body := &Body{}
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
	body := &Body{}
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
	body := &Body{}
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
	body := &Body{}
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

func TestComputeHabitability_ZedPrime(t *testing.T) {
	// Per WBH IISS form p.141:
	// Size 5, Atm 6, Hydro 6, no tidal lock,
	// MeanK 300, HighK 346 (>323 → -2), LowK 262,
	// Gravity 0.66 (0.4-0.7 narrower band wins → -1).
	// DMs: 0 + 0 + 0 + 0 + (-2) + 0 + 0 + (-1) = -3
	// Habitability = 10 - 3 = 7.
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 6}
	body.Temperature = &Temperature{MeanK: 300, HighK: 346, LowK: 262}
	body.Physical = &BodyPhysical{Gravity: 0.66}
	got := ComputeHabitability(body)
	if got.Rating != 7 {
		t.Errorf("Zed Prime: got %d, want 7", got.Rating)
	}
}

func TestComputeHabitability_TerraEquivalent(t *testing.T) {
	// Size 8, Atm 6, Hydro 7 (no DM), no temp DMs, Gravity 1.0 (no DM).
	body := &Body{}
	body.SizeCode = "8"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Temperature = &Temperature{MeanK: 288, HighK: 315, LowK: 255}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	got := ComputeHabitability(body)
	if got.Rating != 10 {
		t.Errorf("Terra: got %d, want 10", got.Rating)
	}
}

func TestComputeHabitability_HighTempHotBand(t *testing.T) {
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 300, HighK: 330, LowK: 280}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// HighK 330 > 323 → -2
	got := ComputeHabitability(body)
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (HighK >323)", got.Rating)
	}
}

func TestComputeHabitability_HighTempColdBand(t *testing.T) {
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 270, HighK: 270, LowK: 260}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// HighK 270 < 279 → -2; MeanK 270 < 273 → -2; LowK 260 ≥ 200 → 0
	got := ComputeHabitability(body)
	if got.Rating != 6 {
		t.Errorf("got %d, want 6 (HighK <279, MeanK <273)", got.Rating)
	}
}

func TestComputeHabitability_MeanTempHottest(t *testing.T) {
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 350, HighK: 380, LowK: 320}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// MeanK 350 > 323 → -4; HighK 380 > 323 → -2
	got := ComputeHabitability(body)
	if got.Rating != 4 {
		t.Errorf("got %d, want 4", got.Rating)
	}
}

func TestComputeHabitability_MeanTempBoundary323(t *testing.T) {
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 323, HighK: 323, LowK: 280}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// MeanK 323 is in [304, 323] → -2 (NOT -4; >323 is strict).
	// HighK 323 → no DM (>323 is strict).
	got := ComputeHabitability(body)
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (boundary 323)", got.Rating)
	}
}

func TestComputeHabitability_MeanTempBoundary324(t *testing.T) {
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 324, HighK: 324, LowK: 280}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// MeanK 324 > 323 → -4; HighK 324 > 323 → -2
	got := ComputeHabitability(body)
	if got.Rating != 4 {
		t.Errorf("got %d, want 4 (boundary 324)", got.Rating)
	}
}

func TestComputeHabitability_LowTempBelow200(t *testing.T) {
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Temperature = &Temperature{MeanK: 280, HighK: 290, LowK: 195}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	// LowK 195 < 200 → -2
	got := ComputeHabitability(body)
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (LowK <200)", got.Rating)
	}
}

func TestComputeHabitability_GravityNarrowerBandWins(t *testing.T) {
	// Gravity 0.5 → in BOTH 0.2-0.7 (-2) AND 0.4-0.7 (-1) → narrower wins → -1.
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 0.5}
	got := ComputeHabitability(body)
	if got.Rating != 9 {
		t.Errorf("got %d, want 9 (gravity 0.5 narrower band -1)", got.Rating)
	}
}

func TestComputeHabitability_GravityResidualLowBand(t *testing.T) {
	// Gravity 0.3 → in 0.2-0.7 only (NOT in 0.4-0.7) → residual band → -2.
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 0.3}
	got := ComputeHabitability(body)
	if got.Rating != 8 {
		t.Errorf("got %d, want 8 (gravity 0.3 residual -2)", got.Rating)
	}
}

func TestComputeHabitability_GravityComfortable(t *testing.T) {
	// Gravity 0.8 → in 0.7-0.9 → DM+1.
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 0.8}
	got := ComputeHabitability(body)
	if got.Rating != 11 {
		t.Errorf("got %d, want 11 (gravity 0.8 +1)", got.Rating)
	}
}

func TestComputeHabitability_GravityHigh(t *testing.T) {
	// Gravity 1.5 → in 1.4-2.0 → DM-3.
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 1.5}
	got := ComputeHabitability(body)
	if got.Rating != 7 {
		t.Errorf("got %d, want 7 (gravity 1.5 -3)", got.Rating)
	}
}

func TestComputeHabitability_GravityCrushing(t *testing.T) {
	// Gravity 2.5 → > 2.0 → DM-6.
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 2.5}
	got := ComputeHabitability(body)
	if got.Rating != 4 {
		t.Errorf("got %d, want 4 (gravity 2.5 -6)", got.Rating)
	}
}

func TestComputeHabitability_GravityVeryLow(t *testing.T) {
	// Gravity 0.1 → < 0.2 → DM-4.
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 0.1}
	got := ComputeHabitability(body)
	if got.Rating != 6 {
		t.Errorf("got %d, want 6 (gravity 0.1 -4)", got.Rating)
	}
}

func TestComputeHabitability_GravityEarthBaseline(t *testing.T) {
	// Gravity 1.0 → in 0.9-1.1 → no DM.
	body := &Body{}
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = &BodyPhysical{Gravity: 1.0}
	got := ComputeHabitability(body)
	if got.Rating != 10 {
		t.Errorf("got %d, want 10 (gravity 1.0 baseline)", got.Rating)
	}
}

func TestComputeHabitability_UndefinedGravity_Size6(t *testing.T) {
	// Physical nil, Size 6 → +1 - |6-6| = +1.
	body := &Body{}
	body.SizeCode = "6"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = nil
	got := ComputeHabitability(body)
	if got.Rating != 11 {
		t.Errorf("got %d, want 11 (undefined gravity Size 6 → +1)", got.Rating)
	}
}

func TestComputeHabitability_UndefinedGravity_Size0(t *testing.T) {
	// Physical nil, Size 0 → +1 - |6-0| = -5.
	// Plus Size 0 (in 0-4) → -1.
	body := &Body{}
	body.SizeCode = "0"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 5}
	body.Physical = nil
	got := ComputeHabitability(body)
	// 10 - 1 (size) - 5 (undefined gravity) = 4
	if got.Rating != 4 {
		t.Errorf("got %d, want 4 (undefined gravity Size 0 → -5)", got.Rating)
	}
}

func TestComputeHabitability_HabitabilityCannotExceed12(t *testing.T) {
	// Synthetic max-positive: Size 9 (+1), Atm 6 (0), Hydro 7 (0), Gravity 0.8 (+1).
	// 10 + 2 = 12.
	body := &Body{}
	body.SizeCode = "9"
	body.Atmosphere = &Atmosphere{Code: 6}
	body.Hydrographics = &Hydrographics{Code: 7}
	body.Physical = &BodyPhysical{Gravity: 0.8}
	got := ComputeHabitability(body)
	if got.Rating != 12 {
		t.Errorf("got %d, want 12 (max positive)", got.Rating)
	}
}

func TestComputeHabitability_Notes(t *testing.T) {
	cases := []struct {
		name      string
		setup     func() *Body
		wantNotes string
	}{
		{
			name: "Terra baseline (no DMs fire)",
			setup: func() *Body {
				body := &Body{}
				body.SizeCode = "8"
				body.Atmosphere = &Atmosphere{Code: 6}
				body.Hydrographics = &Hydrographics{Code: 7}
				body.Temperature = &Temperature{HighK: 310, MeanK: 290, LowK: 270}
				body.Physical = &BodyPhysical{Gravity: 1.0}
				return body
			},
			wantNotes: "",
		},
		{
			name: "Zed Prime (HighK > 323 + low gravity)",
			setup: func() *Body {
				body := &Body{}
				body.SizeCode = "5"
				body.Atmosphere = &Atmosphere{Code: 6}
				body.Hydrographics = &Hydrographics{Code: 5}
				body.Temperature = &Temperature{HighK: 346, MeanK: 290, LowK: 270}
				body.Physical = &BodyPhysical{Gravity: 0.66}
				return body
			},
			wantNotes: "Too hot at times; Low gravity",
		},
		{
			name: "Hostile (atm B + tidal lock 1:1)",
			setup: func() *Body {
				body := &Body{}
				body.SizeCode = "8"
				body.Atmosphere = &Atmosphere{Code: 11}
				body.Hydrographics = &Hydrographics{Code: 7}
				body.TidalLock = &TidalLock{
					Case:           TidalLockCasePlanetToStar,
					LockRatio:      "1:1",
					IsTwilightZone: true,
				}
				body.Temperature = &Temperature{HighK: 310, MeanK: 290, LowK: 270}
				body.Physical = &BodyPhysical{Gravity: 1.0}
				return body
			},
			wantNotes: "Hostile Atmosphere; Very little useable land surface area",
		},
		{
			name: "Multi-temp (HighK and MeanK both > 323)",
			setup: func() *Body {
				body := &Body{}
				body.SizeCode = "8"
				body.Atmosphere = &Atmosphere{Code: 6}
				body.Hydrographics = &Hydrographics{Code: 7}
				body.Temperature = &Temperature{HighK: 350, MeanK: 340, LowK: 250}
				body.Physical = &BodyPhysical{Gravity: 1.0}
				return body
			},
			wantNotes: "Too hot at times; Too hot most of the time",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ComputeHabitability(c.setup())
			if got.Notes != c.wantNotes {
				t.Errorf("got Notes %q, want %q", got.Notes, c.wantNotes)
			}
		})
	}
}
