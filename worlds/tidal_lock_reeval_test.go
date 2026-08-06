package worlds

import (
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
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

func TestGenerateTidalLock_CapturesSnapshot(t *testing.T) {
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

// TestApplyTidalLockReEval_LowPressureSkipped verifies that a body with
// pressure ≤ 2.5 bar is not re-evaluated. TidalLock and Atmosphere are
// unchanged after ApplyTidalLockReEval.
func TestApplyTidalLockReEval_LowPressureSkipped(t *testing.T) {
	snap := CapturePreTidalLockSnapshot(&Body{
		Eccentricity: 0.05,
		AxialTilt:    &AxialTilt{Degrees: 10},
		DayLength:    &DayLength{SiderealHours: 24},
	})
	origTL := &TidalLock{LockRatio: "1:1"}
	body := &Body{
		Kind:                 BodyTerrestrial,
		Atmosphere:           &Atmosphere{Code: 6, Pressure: 1.0}, // ≤ 2.5 bar
		TidalLock:            origTL,
		preTidalLockSnapshot: &snap,
	}
	u := &Universe{}
	u.Detail.Bodies = []Body{*body}

	r := roller.NewScripted() // no dice should be consumed
	if err := ApplyTidalLockReEval(r, u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := &u.Detail.Bodies[0]
	if got.TidalLock == nil || got.TidalLock.LockRatio != "1:1" {
		t.Errorf("TidalLock changed; want 1:1, got %v", got.TidalLock)
	}

	if got.Atmosphere == nil || got.Atmosphere.Pressure != 1.0 {
		t.Errorf("Atmosphere changed; want Pressure=1.0, got %v", got.Atmosphere)
	}

	if got.preTidalLockSnapshot == nil {
		t.Error("preTidalLockSnapshot was cleared; should be untouched")
	}
}

// TestApplyTidalLockReEval_NoSnapshotSkipped verifies that a body with no
// pre-tidal-lock snapshot is skipped even when pressure > 2.5 bar.
func TestApplyTidalLockReEval_NoSnapshotSkipped(t *testing.T) {
	origTL := &TidalLock{LockRatio: "3:2"}
	body := &Body{
		Kind:                 BodyTerrestrial,
		Atmosphere:           &Atmosphere{Code: 12, Pressure: 3.0}, // > 2.5 bar
		TidalLock:            origTL,
		preTidalLockSnapshot: nil, // Stage 4 didn't fire tidal lock
	}
	u := &Universe{}
	u.Detail.Bodies = []Body{*body}

	r := roller.NewScripted() // no dice should be consumed
	if err := ApplyTidalLockReEval(r, u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := &u.Detail.Bodies[0]
	if got.TidalLock == nil || got.TidalLock.LockRatio != "3:2" {
		t.Errorf("TidalLock changed; want 3:2, got %v", got.TidalLock)
	}

	if got.Atmosphere == nil || got.Atmosphere.Pressure != 3.0 {
		t.Errorf("Atmosphere changed; want Pressure=3.0, got %v", got.Atmosphere)
	}
}

// TestApplyTidalLockReEval_HighPressureTriggersReEval verifies the
// orchestration logic of ApplyTidalLockReEval: when a body has pressure
// > 2.5 bar AND a pre-tidal-lock snapshot, the function clears TidalLock,
// re-runs GenerateTidalLock (consuming dice), then clears Stage-5 output
// and re-runs ApplyClimatePasses.
//
// The fixture uses a body that will produce a nil TidalLock from the
// re-roll (all DMs too low after -2 atmosphere penalty), demonstrating
// that body.TidalLock is reset and Stage-5 output is cleared.
//
// We use a terrestrial body orbiting a star (no moonRef / parentPlanet),
// with an orbit and star configuration that gives a marginal DM. After
// the -2 atmosphere DM is applied on re-eval, the total DM drops below
// the threshold and no tidal lock fires — so TidalLock becomes nil and
// the body is not re-evaluated for climate (non-HZ, so ApplyClimatePasses
// is a no-op and leaves Atmosphere nil).
func TestApplyTidalLockReEval_HighPressureTriggersReEval(t *testing.T) {
	// Build a snapshot representing pre-tidal-lock state.
	snap := CapturePreTidalLockSnapshot(&Body{
		Eccentricity: 0.05,
		AxialTilt:    &AxialTilt{Degrees: 5},
		DayLength:    &DayLength{SiderealHours: 24, SolarHours: 25},
	})

	// The body is non-HZ so ApplyClimatePasses will be a no-op (no
	// atmosphere re-generated). This lets us confirm:
	//   - body.TidalLock is reset by the re-eval (re-roll returns nil
	//     when all DMs are ≤ -10 after including the atmosphere penalty).
	//   - body.Atmosphere is cleared by ClearStage5Output after re-roll.
	body := &Body{
		Kind:         BodyTerrestrial,
		SizeCode:     "5",
		Eccentricity: 0.05,
		AxialTilt:    &AxialTilt{Degrees: 5},
		DayLength:    &DayLength{SiderealHours: 24, SolarHours: 25},
		Orbit:        5.0, // orbit > 3 → DM-(5×2)=-10, total too low for a lock
		Period:       Period{Hours: 8766 * 5},
		HZ:           false, // non-HZ: ApplyClimatePasses short-circuits
		// Stage-5 output already set (simulates post-ApplyClimate state).
		Atmosphere:           &Atmosphere{Code: 12, Pressure: 3.0},
		TidalLock:            &TidalLock{LockRatio: "1:1"},
		preTidalLockSnapshot: &snap,
	}

	u := &Universe{
		System: stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}},
	}
	u.Detail.Bodies = []Body{*body}

	// When all DMs total ≤ -10 SelectHighestDMCases filters them out and
	// GenerateTidalLock returns early with nil — no dice consumed.
	// The scripted roller is empty to prove this; any roll would panic.
	r := roller.NewScripted()

	if err := ApplyTidalLockReEval(r, u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := &u.Detail.Bodies[0]

	// TidalLock must have been re-evaluated (nil if no case fired, or a
	// new TidalLock if a case fired with the scripted roll).
	// The pre-eval TidalLock was "1:1" — any mutation proves re-eval ran.
	// If the re-roll produces nil (no case), that's also fine: the old
	// "1:1" value must be gone.
	if got.TidalLock != nil && got.TidalLock.LockRatio == "1:1" {
		// The old value survived unchanged — re-eval didn't run.
		t.Error("TidalLock still has original 1:1 lock; re-eval did not run")
	}

	// Atmosphere must be nil: ClearStage5Output ran, and ApplyClimatePasses
	// is a no-op on non-HZ bodies, so no new atmosphere is generated.
	if got.Atmosphere != nil {
		t.Errorf("Atmosphere not cleared after re-eval; got %+v", got.Atmosphere)
	}

	// TidalEffects must be non-nil: reEvalBody recomputes surface tidal
	// effects from the new TidalLock before clearing Stage-5 output.
	// Even when TidalLock becomes nil (no case fired), the stellar tide
	// component is always computed for non-empty bodies.
	if got.TidalEffects == nil {
		t.Error("TidalEffects is nil after re-eval; recompute did not run")
	}
}

// TestSolFidelity_ReEval_Venus verifies that Venus's 92-bar atmosphere
// triggers the WBH p.106 atmosphere-DM re-eval cascade and shifts the
// tidal-lock outcome to a result the pre-#9 single-pass pipeline could
// not produce.
//
// Real Venus parameters used:
//
//	diameter        12,104 km → SizeCode "8" (closest anchor 12,800 km)
//	eccentricity    0.0068
//	semi-major axis 0.723 AU → Orbit# 2.077 (per stars.AUToOrbit)
//	axial tilt      177.36° (retrograde rotation)
//	pressure        92 bar (real Venus)
//	parent star     Sol (mass 1.0 M☉, age 4.6 Gyr)
//
// DM trace WITHOUT atmosphere DM (what Stage 4 produced):
//
//	commonTidalLockDMs:
//	  Size 8 → +ceil(8/3) = +3
//	  Ecc 0.0068 ≤ 0.1 → 0
//	  Tilt 177.36° > 30 → −2 (the 60–120 and 80–100 bands don't fire at 177°)
//	  No atmosphere yet → 0
//	  Age 4.6 Gyr (< 5) → 0
//	  Common = +1
//	planetToStarDMs:
//	  Base = −4
//	  Orbit# 2.077 ∈ [1, 3) → orbit-bracket DM: 2 ≤ orbit < 3 → +1
//	  Star mass 1.0 ≤ 1.0 → −1
//	  No multi-star, no moon penalty
//	  Specific = −4
//	Total = −3 → max InitialResult = 9 (retrograde rotation, 1D×10×24).
//
// DM trace WITH atmosphere DM (post-#9 re-eval sees pressure 92):
//
//	Common: +3 − 2 (tilt) − 2 (pressure > 2.5) + 0 (age) = −1
//	Specific: unchanged (−4)
//	Total = −5 → max InitialResult = 7 (random rotation, 1D×5×24).
//
// The atmosphere DM lowers the cap from result 9 to result 7 — a real
// behavioral change. Without the cascade the pre-#9 pipeline would have
// committed result 9 (retrograde); the cascade now produces result ≤ 7.
//
// Fidelity note: WBH does not reproduce Venus's *actual* spin (5,832-hour
// retrograde sidereal day, 117-day solar day). With real ecc/tilt/age the
// procedure caps at result 7 — a random rotation in the 1D×5×24-hour
// range (120–720 h), well short of Venus's actual rotation. The cascade
// is what the book prescribes; that it doesn't match observed Venus is a
// further example of the book-vs-reality divergence (see also Mercury 3:2
// and Pluto↔Charon mutual 1:1 in worlds/sol_fidelity_test.go).
func TestSolFidelity_ReEval_Venus(t *testing.T) {
	// Snapshot captures pre-tidal-lock state (post-Stage-4 axial tilt,
	// pre-Stage-4 day length).
	snap := CapturePreTidalLockSnapshot(&Body{
		Eccentricity: 0.0068,
		AxialTilt:    &AxialTilt{Degrees: 177.36, Retrograde: true, BaselineDegrees: 177.36},
		DayLength:    &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
	})

	body := &Body{
		Kind:         BodyTerrestrial,
		SizeCode:     "8",
		Eccentricity: 0.0068,
		AxialTilt:    &AxialTilt{Degrees: 177.36, Retrograde: true, BaselineDegrees: 177.36},
		DayLength:    &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		Orbit:        2.077,
		Period:       Period{Hours: 0.6152 * 8766}, // Venus 224.7-day year
		HZ:           false,                        // skip the ApplyClimatePasses re-run
		// Post-Stage-4 / post-Stage-5 state at the moment ApplyTidalLockReEval runs:
		Atmosphere:           &Atmosphere{Code: 11, Pressure: 92.0}, // real Venus pressure
		TidalLock:            &TidalLock{FinalResult: 9, BecomesRetrograde: true},
		preTidalLockSnapshot: &snap,
	}

	u := &Universe{
		System: stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 4.6}},
	}
	u.Detail.Bodies = []Body{*body}

	// Re-eval scripted dice: 2D = 12 (max possible) with DM = −5 produces
	// InitialResult = 7 (random rotation). Result 7 consumes one 1D for
	// the day-length (1D × 5 × 24 hours). 1D=3 → 360 hours.
	r := roller.NewScripted(12, 3)

	if err := ApplyTidalLockReEval(r, u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := &u.Detail.Bodies[0]

	if got.TidalLock == nil {
		t.Fatal("TidalLock is nil after re-eval; expected a re-rolled lock")
	}

	if got.TidalLock.FinalResult == 9 {
		t.Error("FinalResult is still 9 (pre-eval value); re-eval did not run")
	}

	if got.TidalLock.FinalResult > 7 {
		t.Errorf("FinalResult = %d, expected ≤ 7 (atmosphere DM −2 caps max at 7)",
			got.TidalLock.FinalResult)
	}

	if got.TidalLock.FinalResult != 7 {
		t.Errorf("FinalResult = %d, want 7 (2D=12 with DM=−5)", got.TidalLock.FinalResult)
	}
	// ClearStage5Output ran; HZ=false so ApplyClimatePasses is a no-op.
	if got.Atmosphere != nil {
		t.Errorf("Atmosphere not cleared after re-eval; got %+v", got.Atmosphere)
	}
}
