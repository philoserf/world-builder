package worlds

import (
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestClearStage5Output_ZeroesAllStage5Fields(t *testing.T) {
	body := &Body{
		// Stage 4 fields — must SURVIVE ClearStage5Output.
		DayLength:    &DayLength{SiderealHours: 24, SolarHours: 24},
		AxialTilt:    &AxialTilt{Degrees: 23},
		TidalLock:    &TidalLock{LockRatio: "1:1"},
		TidalEffects: &SurfaceTidalEffects{},
		Eccentricity: 0.05,

		// Stage 5 fields — must be cleared.
		Atmosphere:    &Atmosphere{Code: 6, Pressure: 1.0},
		Hydrographics: &Hydrographics{Code: 7},
		Temperature:   &Temperature{MeanK: 288},
		Geology:       &Geology{ResidualSeismicStress: 3},
	}

	ClearStage5Output(body)

	// Stage 4 fields must be untouched.
	if body.DayLength == nil {
		t.Error("DayLength was cleared; must survive ClearStage5Output")
	}
	if body.AxialTilt == nil {
		t.Error("AxialTilt was cleared; must survive ClearStage5Output")
	}
	if body.TidalLock == nil {
		t.Error("TidalLock was cleared; must survive ClearStage5Output")
	}
	if body.TidalEffects == nil {
		t.Error("TidalEffects was cleared; must survive ClearStage5Output")
	}
	if body.Eccentricity != 0.05 {
		t.Errorf("Eccentricity changed to %v; must survive ClearStage5Output", body.Eccentricity)
	}

	// Stage 5 fields must be zeroed.
	if body.Atmosphere != nil {
		t.Errorf("Atmosphere not cleared: %+v", body.Atmosphere)
	}
	if body.Hydrographics != nil {
		t.Errorf("Hydrographics not cleared: %+v", body.Hydrographics)
	}
	if body.Temperature != nil {
		t.Errorf("Temperature not cleared: %+v", body.Temperature)
	}
	if body.Geology != nil {
		t.Errorf("Geology not cleared: %+v", body.Geology)
	}
}

func TestPreTidalLockSnapshot_RoundTrip(t *testing.T) {
	body := &Body{
		Eccentricity: 0.42,
		AxialTilt:    &AxialTilt{Degrees: 27, Retrograde: false},
		DayLength:    &DayLength{SiderealHours: 24, SolarHours: 24, YearDays: 365},
	}
	snap := CapturePreTidalLockSnapshot(body)
	// Mutate the body after capture.
	body.Eccentricity = 0.01
	body.AxialTilt.Degrees = 90
	body.AxialTilt.Retrograde = true
	body.DayLength.SiderealHours = 8766

	snap.RestoreInto(body)
	if body.Eccentricity != 0.42 {
		t.Errorf("Eccentricity not restored: %g, want 0.42", body.Eccentricity)
	}
	if body.AxialTilt == nil {
		t.Fatal("AxialTilt is nil after restore")
	}
	if body.AxialTilt.Degrees != 27 || body.AxialTilt.Retrograde {
		t.Errorf("AxialTilt not restored: %+v, want Degrees=27 Retrograde=false", body.AxialTilt)
	}
	if body.DayLength == nil {
		t.Fatal("DayLength is nil after restore")
	}
	if body.DayLength.SiderealHours != 24 {
		t.Errorf("DayLength.SiderealHours not restored: %v, want 24", body.DayLength.SiderealHours)
	}
}

func TestPreTidalLockSnapshot_NilFields(t *testing.T) {
	body := &Body{Eccentricity: 0.1} // nil DayLength, nil AxialTilt
	snap := CapturePreTidalLockSnapshot(body)
	snap.RestoreInto(body) // must not panic; body still has nil DayLength/AxialTilt
	if body.AxialTilt != nil {
		t.Errorf("AxialTilt = %v, want nil", body.AxialTilt)
	}
	if body.DayLength != nil {
		t.Errorf("DayLength = %v, want nil", body.DayLength)
	}
	if body.Eccentricity != 0.1 {
		t.Errorf("Eccentricity = %v, want 0.1", body.Eccentricity)
	}
}

func TestPreTidalLockSnapshot_IndependentOfBody(t *testing.T) {
	// Mutating snapshot fields after capture must NOT affect the body
	// (deep copy semantics for the embedded pointer fields).
	body := &Body{
		AxialTilt: &AxialTilt{Degrees: 10},
		DayLength: &DayLength{SiderealHours: 12},
	}
	snap := CapturePreTidalLockSnapshot(body)
	// Mutate the snapshot's AxialTilt and DayLength.
	if snap.AxialTilt != nil {
		snap.AxialTilt.Degrees = 99
	}
	if snap.DayLength != nil {
		snap.DayLength.SiderealHours = 99
	}
	// Body's originals must be untouched.
	if body.AxialTilt.Degrees != 10 {
		t.Errorf("body.AxialTilt.Degrees mutated to %v via snapshot", body.AxialTilt.Degrees)
	}
	if body.DayLength.SiderealHours != 12 {
		t.Errorf("body.DayLength.SiderealHours mutated to %v via snapshot", body.DayLength.SiderealHours)
	}
}

func TestGenerateTidalLock_CapturesSnapshotAndDMs(t *testing.T) {
	// Zed Prime moon→planet fixture — known to fire (DM=+7). Use the same
	// scripted roll sequence as TestGenerateTidalLock_ZedPrime_FullPath:
	//   2D=6 → initial result 13 (≥12, verification fires)
	//   2D=12 → natural-12 → reroll with no DMs
	//   2D=4 → FinalResult=4 (day × 2, no extra dice consumed)
	// After ApplyTidalLockEffect the body's Eccentricity is unchanged (0.25)
	// because FinalResult=4 < 11 (no lock ratio), so eccentricity reroll is
	// not triggered. The snapshot must preserve the pre-effect eccentricity.
	moonRef := &Body{
		Kind:         BodyMoon,
		SizeCode:     "5",
		OrbitPD:      22,
		Retrograde:   true,
		Eccentricity: 0.25,
	}
	parent := &Body{Kind: BodyGasGiant, MassEarth: 1200, Orbit: 1.06}
	body := &Body{
		Kind:         BodyTerrestrial,
		SizeCode:     "5",
		Eccentricity: 0.25,
		AxialTilt:    &AxialTilt{Degrees: 73.65, BaselineDegrees: 73.65},
		DayLength:    &DayLength{SiderealHours: 42.37, BaselineSiderealHours: 42.37},
		Period:       Period{Years: 0.072, Hours: 0.072 * 8766},
	}
	sys := stars.System{Primary: stars.Star{Mass: 0.918, AgeGyr: 6.3}}

	r := roller.NewScripted(6, 12, 4)
	tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, body.Period.Hours)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock — fixture should hit a case")
	}

	if body.preTidalLockSnapshot == nil {
		t.Error("body.preTidalLockSnapshot is nil; snapshot was not captured")
	} else if body.preTidalLockSnapshot.Eccentricity != 0.25 {
		t.Errorf("snapshot Eccentricity = %v, want 0.25 (pre-effect)", body.preTidalLockSnapshot.Eccentricity)
	}

	if len(tl.PreEvalDMs) == 0 {
		t.Error("tl.PreEvalDMs is empty; DM map was not captured")
	}
	if _, ok := tl.PreEvalDMs[TidalLockCaseMoonToPlanet]; !ok {
		t.Errorf("tl.PreEvalDMs missing moon→planet case; got %v", tl.PreEvalDMs)
	}
}

func TestGenerateTidalLock_NoSnapshotWhenNoCase(t *testing.T) {
	// Empty body: no case applies, GenerateTidalLock returns nil immediately.
	// No snapshot should be captured and no dice consumed.
	body := &Body{Kind: BodyEmpty}
	r := roller.NewScripted()
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}
	tl, err := GenerateTidalLock(r, body, nil, sys, nil, 8766)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl != nil {
		t.Errorf("expected nil TidalLock for empty body, got %v", tl)
	}
	if body.preTidalLockSnapshot != nil {
		t.Errorf("snapshot captured for empty body: %v", body.preTidalLockSnapshot)
	}
}
