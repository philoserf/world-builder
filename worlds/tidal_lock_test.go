package worlds

import (
	"math"
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestEvaluateTidalLockDMs_PlanetToStar_Mercury(t *testing.T) {
	// Mercury-like: Size 4, Orbit# 1.5, eccentricity 0.21, axial tilt 0°,
	// no atmosphere, system age ~5 Gyr, around solar-mass primary.
	// Expected DM stack:
	//   common:
	//     Size 4 → DM+ceil(4/3) = +2
	//     Eccentricity 0.21 → DM-floor(0.21×10) = -2
	//     Axial tilt 0° → no DM (not above 30°)
	//     No atmosphere → no pressure DM
	//     Age 5.0 Gyr (between 5 and 10) → DM+2
	//   common total: +2
	//   planet→star specific:
	//     Base: -4
	//     Orbit# 1.5 between 1 and 2 → DM+4
	//     Star mass 1.0 between 0.5 and 1.0 → DM-1
	//     Single star, no significant moons → 0
	//   specific total: -1
	//   Total: +2 + (-1) = +1
	body := &Body{}
	body.Kind = BodyTerrestrial
	body.SizeCode = "4"
	body.Orbit = 1.5
	body.Eccentricity = 0.21
	body.AxialTilt = &AxialTilt{Degrees: 0}

	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}

	dms := EvaluateTidalLockDMs(body, sys, nil, nil)
	got, ok := dms[TidalLockCasePlanetToStar]
	if !ok {
		t.Fatal("planet→star case missing from DM map")
	}
	want := 1 // +2 - 2 + 2 - 4 + 4 - 1 = +1
	if got != want {
		t.Errorf("planet→star DM total: got %d, want %d", got, want)
	}
}

func TestEvaluateTidalLockDMs_MoonToPlanet_ZedPrime(t *testing.T) {
	// Zed Prime per WBH p.106 narrative:
	//   common DMs:
	//     Size 5 → ceil(5/3) = +2
	//     Eccentricity 0.25 → -floor(0.25×10) = -2
	//     Tilt 73.65°: above 30° → -2; between 60°-120° → -4; total tilt DM = -6
	//     No atmosphere pointer → 0
	//     Age 6.3 Gyr (between 5 and 10) → +2
	//   common total: +2 - 2 - 6 + 2 = -4
	//   moon→planet specific:
	//     Base: +6
	//     OrbitPD 22 > 20 → -floor(22/20) = -1
	//     Retrograde → -2
	//     Planet mass 1200 ≥ 1000 → +8
	//   specific total: +11
	//   Total: -4 + 11 = +7
	moonRef := &Body{
		Kind:         BodyMoon,
		SizeCode:     "5",
		OrbitPD:      22,
		Retrograde:   true,
		Eccentricity: 0.25,
	}
	parent := &Body{}
	parent.Kind = BodyGasGiant
	parent.MassEarth = 1200
	parent.Orbit = 1.06

	body := &Body{}
	body.Kind = BodyTerrestrial
	body.SizeCode = "5"
	body.Eccentricity = 0.25
	body.AxialTilt = &AxialTilt{Degrees: 73.65}

	sys := stars.System{Primary: stars.Star{Mass: 0.918, AgeGyr: 6.3}}

	dms := EvaluateTidalLockDMs(body, sys, parent, moonRef)
	got, ok := dms[TidalLockCaseMoonToPlanet]
	if !ok {
		t.Fatal("moon→planet case missing from DM map")
	}
	want := 7
	if got != want {
		t.Errorf("moon→planet DM total for Zed Prime: got %d, want %d", got, want)
	}
}

func TestEvaluateTidalLockDMs_PlanetToStar_AbsentForMoon(t *testing.T) {
	// A moon's body.Orbit is always 0 (moons store their parent-relative
	// orbit in OrbitPD, not the star-relative Orbit#). Evaluating the
	// planet→star case for a moon would feed orbit=0 into the
	// orbit < 1 branch and award a spurious +14 close-orbit DM, making
	// the planet→star case outrank moon→planet for moons of small or
	// moderate-mass parents. Per WBH p.106 the planet→star case applies
	// to planets, not moons.
	//
	// Mirrors stage4.go: in production, GenerateTidalLock is called
	// with moonRef = body (same pointer, Kind=BodyMoon) for moons.
	body := &Body{
		Kind:     BodyMoon,
		SizeCode: "3",
		OrbitPD:  15,
	}
	body.AxialTilt = &AxialTilt{Degrees: 0}

	parent := &Body{}
	parent.Kind = BodyTerrestrial
	parent.MassEarth = 1.0
	parent.Orbit = 1.0

	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}

	dms := EvaluateTidalLockDMs(body, sys, parent, body)
	if _, ok := dms[TidalLockCasePlanetToStar]; ok {
		t.Errorf("planet→star case should not appear for a moon, got dms=%+v", dms)
	}
}

func TestEvaluateTidalLockDMs_PlanetToMoon_OnlyIfHasSignificantMoon(t *testing.T) {
	// Planet→moon case is absent when the planet has no significant (Size 1+) moons.
	body := &Body{SizeCode: "3"}
	body.Kind = BodyTerrestrial
	body.AxialTilt = &AxialTilt{Degrees: 0}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}

	// No moons → planet→moon case absent.
	dms := EvaluateTidalLockDMs(body, sys, nil, nil)
	if _, ok := dms[TidalLockCasePlanetToMoon]; ok {
		t.Errorf("planet→moon case should not appear when planet has no significant moon, got dms=%+v", dms)
	}
}

func TestEvaluateTidalLockDMs_NoMoonCases_NotAMoon(t *testing.T) {
	// A planet (parentPlanet=nil, moonRef=nil) cannot be locked to a planet.
	body := &Body{SizeCode: "5"}
	body.Kind = BodyTerrestrial
	body.AxialTilt = &AxialTilt{Degrees: 0}
	body.Eccentricity = 0.0
	body.Orbit = 5.0
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}

	dms := EvaluateTidalLockDMs(body, sys, nil, nil)
	if _, ok := dms[TidalLockCaseMoonToPlanet]; ok {
		t.Errorf("moon→planet should not apply to planets, got dms=%+v", dms)
	}
}

func TestSelectHighestDMCase_FilterBelowMinusTen(t *testing.T) {
	// Cases with DM ≤ -10 → filtered out per p.106.
	dms := map[TidalLockCase]int{
		TidalLockCasePlanetToStar: -12,
		TidalLockCaseMoonToPlanet: 5,
	}
	body := &Body{}
	kase, dm := SelectHighestDMCase(dms, body)
	if kase != TidalLockCaseMoonToPlanet {
		t.Errorf("got case %v, want MoonToPlanet (planet→star filtered as ≤-10)", kase)
	}
	if dm != 5 {
		t.Errorf("got DM %d, want 5", dm)
	}
}

func TestSelectHighestDMCase_AllFiltered_ReturnsNone(t *testing.T) {
	dms := map[TidalLockCase]int{
		TidalLockCasePlanetToStar: -15,
		TidalLockCaseMoonToPlanet: -11,
	}
	body := &Body{}
	kase, _ := SelectHighestDMCase(dms, body)
	if kase != TidalLockCaseNone {
		t.Errorf("got case %v, want None", kase)
	}
}

func TestSelectHighestDMCase_TieMoonFirst(t *testing.T) {
	// Per p.106: when tied, moon-cases roll first.
	dms := map[TidalLockCase]int{
		TidalLockCasePlanetToStar: 5,
		TidalLockCaseMoonToPlanet: 5,
	}
	body := &Body{}
	kase, _ := SelectHighestDMCase(dms, body)
	if kase != TidalLockCaseMoonToPlanet {
		t.Errorf("got case %v, want MoonToPlanet (moon-cases first on tie)", kase)
	}
}

func TestIsTerrestrial(t *testing.T) {
	cases := []struct {
		name string
		kind BodyKind
		want bool
	}{
		{"terrestrial true", BodyTerrestrial, true},
		{"gas giant false", BodyGasGiant, false},
		{"belt false", BodyPlanetoidBelt, false},
		{"empty false", BodyEmpty, false},
		{"moon false (planet→moon evaluated from planet only)", BodyMoon, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := &Body{Kind: c.kind}
			if got := isTerrestrial(b); got != c.want {
				t.Errorf("isTerrestrial(%v) = %v, want %v", c.kind, got, c.want)
			}
		})
	}
}

func TestRollTidalLockStatus_Plain2DPlusDM(t *testing.T) {
	// 2D=8, DM+3 → 11.
	r := roller.NewScripted(8)
	got := RollTidalLockStatus(r, 3)
	if got != 11 {
		t.Errorf("got %d, want 11 (2D=8 + DM+3)", got)
	}
}

func TestRollTidalLockStatus_NegativeDMs(t *testing.T) {
	// 2D=4, DM-3 → 1.
	r := roller.NewScripted(4)
	got := RollTidalLockStatus(r, -3)
	if got != 1 {
		t.Errorf("got %d, want 1 (2D=4 + DM-3)", got)
	}
}

func TestApplyTidalLockEffect_NoEffectResult2(t *testing.T) {
	body := &Body{}
	body.DayLength = &DayLength{SiderealHours: 24, BaselineSiderealHours: 24}
	body.AxialTilt = &AxialTilt{Degrees: 30}
	body.Eccentricity = 0.05

	r := roller.NewScripted()
	tl, err := ApplyTidalLockEffect(r, body, nil, TidalLockCasePlanetToStar, 2, 8766.0)
	if err != nil {
		t.Fatal(err)
	}
	if tl.LockRatio != "" {
		t.Errorf("LockRatio: got %q, want empty", tl.LockRatio)
	}
	if body.DayLength.SiderealHours != 24 {
		t.Errorf("SiderealHours mutated: %v", body.DayLength.SiderealHours)
	}
}

func TestApplyTidalLockEffect_DayMultiplier_Result4(t *testing.T) {
	body := &Body{}
	body.DayLength = &DayLength{SiderealHours: 42.37, BaselineSiderealHours: 42.37}
	r := roller.NewScripted() // result 4 doesn't roll any further dice
	tl, err := ApplyTidalLockEffect(r, body, nil, TidalLockCaseMoonToPlanet, 4, 7056.63)
	if err != nil {
		t.Fatal(err)
	}
	if tl.DayLengthMultiplier != 2.0 {
		t.Errorf("DayLengthMultiplier: got %v, want 2.0", tl.DayLengthMultiplier)
	}
	if math.Abs(body.DayLength.SiderealHours-84.74) > 0.01 {
		t.Errorf("SiderealHours: got %v, want 84.74", body.DayLength.SiderealHours)
	}
}

func TestApplyTidalLockEffect_OneToOneLock_StarCase_TwilightZone(t *testing.T) {
	body := &Body{}
	body.DayLength = &DayLength{SiderealHours: 24, BaselineSiderealHours: 24}
	body.AxialTilt = &AxialTilt{Degrees: 0, BaselineDegrees: 0}
	body.Eccentricity = 0.0
	body.Period = Period{Years: 0.5, Hours: 4383}

	// 1:1 lock, no axial-tilt reroll (tilt < 3°), no ecc reroll (ecc < 0.1).
	// Verification roll: 2D=10 (NOT natural 12) → no reroll, lock stands.
	r := roller.NewScripted(10)
	tl, err := ApplyTidalLockEffect(r, body, nil, TidalLockCasePlanetToStar, 12, 4383.0)
	if err != nil {
		t.Fatal(err)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio: got %q, want 1:1", tl.LockRatio)
	}
	if !tl.IsTwilightZone {
		t.Error("expected IsTwilightZone for star→planet 1:1 lock")
	}
	if body.DayLength.SiderealHours != 4383 {
		t.Errorf("SiderealHours: got %v, want 4383 (= year hours)", body.DayLength.SiderealHours)
	}
	if body.DayLength.SolarHours != 0 {
		t.Errorf("SolarHours: got %v, want 0 (twilight zone)", body.DayLength.SolarHours)
	}
}

func TestApplyTidalLockEffect_NaturalTwelve_BreaksLock_ZedPath(t *testing.T) {
	// Zed Prime path: InitialResult=13 (1:1 lock pending) → verification 2D=12 (natural 12)
	// → reroll TidalLockStatus with no DMs → 2D=4 → result 4 → day × 2 effect.
	body := &Body{}
	body.DayLength = &DayLength{SiderealHours: 42.37, BaselineSiderealHours: 42.37}
	body.AxialTilt = &AxialTilt{Degrees: 73.65, BaselineDegrees: 73.65}
	body.Eccentricity = 0.25

	// Verification rolls 12 (natural), then reroll status with no DMs rolls 4.
	r := roller.NewScripted(12, 4)
	tl, err := ApplyTidalLockEffect(r, body, nil, TidalLockCaseMoonToPlanet, 13, 7056.63)
	if err != nil {
		t.Fatal(err)
	}
	if !tl.VerificationFired {
		t.Error("expected VerificationFired=true")
	}
	if tl.InitialResult != 13 {
		t.Errorf("InitialResult: got %d, want 13", tl.InitialResult)
	}
	if tl.FinalResult != 4 {
		t.Errorf("FinalResult: got %d, want 4", tl.FinalResult)
	}
	if tl.LockRatio != "" {
		t.Errorf("LockRatio: got %q, want empty (lock broken by verification)", tl.LockRatio)
	}
	if math.Abs(tl.DayLengthMultiplier-2.0) > 0.001 {
		t.Errorf("DayLengthMultiplier: got %v, want 2.0", tl.DayLengthMultiplier)
	}
	if math.Abs(body.DayLength.SiderealHours-84.74) > 0.01 {
		t.Errorf("SiderealHours: got %v, want 84.74", body.DayLength.SiderealHours)
	}
	// Axial tilt unchanged (no lock means no axial-tilt mutation).
	if math.Abs(body.AxialTilt.Degrees-73.65) > 0.05 {
		t.Errorf("Degrees: got %v, want 73.65 (unchanged)", body.AxialTilt.Degrees)
	}
	if tl.AxialTiltMutated {
		t.Error("expected AxialTiltMutated=false")
	}
	if tl.EccentricityMutated {
		t.Error("expected EccentricityMutated=false")
	}
}

func TestApplyTidalLockEffect_OneToOneLock_AxialTiltReroll(t *testing.T) {
	// 1:1 lock with old tilt > 3° → reroll as (2D-2)/10. Verification doesn't reroll.
	body := &Body{}
	body.DayLength = &DayLength{SiderealHours: 24, BaselineSiderealHours: 24}
	body.AxialTilt = &AxialTilt{Degrees: 25, BaselineDegrees: 25}
	body.Eccentricity = 0.0
	body.Period = Period{Years: 1.0, Hours: 8766}

	// Verification: 2D=11 (not natural 12) → lock stands.
	// Axial tilt reroll: 2D=8 → (8-2)/10 = 0.6°.
	r := roller.NewScripted(11, 8)
	tl, err := ApplyTidalLockEffect(r, body, nil, TidalLockCasePlanetToStar, 12, 8766.0)
	if err != nil {
		t.Fatal(err)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio: got %q, want 1:1", tl.LockRatio)
	}
	if !tl.AxialTiltMutated {
		t.Error("expected AxialTiltMutated=true")
	}
	if math.Abs(body.AxialTilt.Degrees-0.6) > 0.05 {
		t.Errorf("Degrees: got %v, want 0.6", body.AxialTilt.Degrees)
	}
	if body.AxialTilt.BaselineDegrees != 25 {
		t.Errorf("BaselineDegrees should preserve original 25, got %v", body.AxialTilt.BaselineDegrees)
	}
}

func TestApplyTidalLockEffect_OneToOneLock_EccentricityReroll(t *testing.T) {
	// 1:1 lock with old ecc > 0.1 → reroll with DM-2, take min of original/new.
	// Verification: 2D=10 (not natural 12) → lock stands.
	// No axial tilt reroll (tilt < 3°).
	// Ecc reroll: 2D=5 → row = max(5, 5-2=3)=5 → SecondRoll="1D"=3 → v=-0.001+3/1000=0.002.
	body := &Body{}
	body.Eccentricity = 0.25
	body.AxialTilt = &AxialTilt{Degrees: 0, BaselineDegrees: 0}
	body.DayLength = &DayLength{SiderealHours: 24, BaselineSiderealHours: 24}
	body.Period = Period{Years: 1.0, Hours: 8766}

	r := roller.NewScripted(
		10, // verification 2D=10 (not natural 12)
		5,  // ecc table 2D=5 → row=max(5,3)=5
		3,  // ecc row-5 SecondRoll "1D"=3 → v=0.002 < 0.25
	)
	tl, err := ApplyTidalLockEffect(r, body, nil, TidalLockCasePlanetToStar, 12, 8766.0)
	if err != nil {
		t.Fatal(err)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio: got %q, want 1:1", tl.LockRatio)
	}
	if !tl.EccentricityMutated {
		t.Error("expected EccentricityMutated=true")
	}
	// stars.RollEccentricity with ExtraDM=-2 + scripted (5, 3) yields ecc ≈ 0.002.
	// Tight upper bound catches a backwards-min bug (would write back 0.25).
	if body.Eccentricity > 0.01 {
		t.Errorf("Eccentricity: got %v, want ≤ 0.01 (min(0.25, ~0.002))", body.Eccentricity)
	}
}

// TestApplyTidalLockEffect_BecomesRetrograde_LowTiltFlips covers FinalResult=9
// (the lower of the two BecomesRetrograde branches). Body starts at tilt 30°;
// the post-process should flip degrees to 180-30 = 150 and set Retrograde=true.
// FinalResult=9 also rolls 1D for NewSiderealHours = 1D × 5 × 24h.
func TestApplyTidalLockEffect_BecomesRetrograde_LowTiltFlips(t *testing.T) {
	body := &Body{}
	body.DayLength = &DayLength{SiderealHours: 24, BaselineSiderealHours: 24}
	body.AxialTilt = &AxialTilt{Degrees: 30, BaselineDegrees: 30}

	// 1D roll for NewSiderealHours: 3 → 3 × 10 × 24 = 720h.
	r := roller.NewScripted(3)
	tl, err := ApplyTidalLockEffect(r, body, nil, TidalLockCasePlanetToStar, 9, 8766.0)
	if err != nil {
		t.Fatal(err)
	}
	if !tl.BecomesRetrograde {
		t.Error("BecomesRetrograde: got false, want true (FinalResult=9)")
	}
	if tl.LockRatio != "" {
		t.Errorf("LockRatio: got %q, want empty (BecomesRetrograde branch)", tl.LockRatio)
	}
	if tl.NewSiderealHours != 720 {
		t.Errorf("NewSiderealHours: got %v, want 720 (1D=3 × 10 × 24)", tl.NewSiderealHours)
	}
	if body.AxialTilt.Degrees != 150 {
		t.Errorf("AxialTilt.Degrees: got %v, want 150 (flipped 180-30)", body.AxialTilt.Degrees)
	}
	if !body.AxialTilt.Retrograde {
		t.Error("AxialTilt.Retrograde: got false, want true (degrees > 90 after flip)")
	}
	if body.DayLength.SiderealHours != 720 {
		t.Errorf("body.DayLength.SiderealHours: got %v, want 720", body.DayLength.SiderealHours)
	}
}

// TestApplyTidalLockEffect_BecomesRetrograde_HighTiltNoFlip covers the FinalResult=10
// branch where a body already has tilt >= 90°. The post-process must NOT flip
// (no 180-degrees mutation) and must set Retrograde=true.
func TestApplyTidalLockEffect_BecomesRetrograde_HighTiltNoFlip(t *testing.T) {
	body := &Body{}
	body.DayLength = &DayLength{SiderealHours: 24, BaselineSiderealHours: 24}
	body.AxialTilt = &AxialTilt{Degrees: 120, BaselineDegrees: 120}

	// 1D roll for NewSiderealHours: 2 → 2 × 50 × 24 = 2400h (FinalResult=10 multiplier).
	r := roller.NewScripted(2)
	tl, err := ApplyTidalLockEffect(r, body, nil, TidalLockCasePlanetToStar, 10, 8766.0)
	if err != nil {
		t.Fatal(err)
	}
	if !tl.BecomesRetrograde {
		t.Error("BecomesRetrograde: got false, want true (FinalResult=10)")
	}
	if tl.NewSiderealHours != 2400 {
		t.Errorf("NewSiderealHours: got %v, want 2400 (1D=2 × 50 × 24)", tl.NewSiderealHours)
	}
	if body.AxialTilt.Degrees != 120 {
		t.Errorf("AxialTilt.Degrees: got %v, want 120 (no flip, already > 90)", body.AxialTilt.Degrees)
	}
	if !body.AxialTilt.Retrograde {
		t.Error("AxialTilt.Retrograde: got false, want true (degrees > 90)")
	}
}

// TestApplyTidalLockEffect_BecomesRetrograde_NilAxialTilt verifies the
// post-process safely no-ops when body.AxialTilt is nil — a body with no
// rolled tilt must not panic on the BecomesRetrograde flag.
func TestApplyTidalLockEffect_BecomesRetrograde_NilAxialTilt(t *testing.T) {
	body := &Body{}
	body.DayLength = &DayLength{SiderealHours: 24, BaselineSiderealHours: 24}
	// AxialTilt deliberately nil.

	r := roller.NewScripted(4)
	tl, err := ApplyTidalLockEffect(r, body, nil, TidalLockCasePlanetToStar, 9, 8766.0)
	if err != nil {
		t.Fatal(err)
	}
	if !tl.BecomesRetrograde {
		t.Error("BecomesRetrograde: got false, want true (FinalResult=9)")
	}
	if body.AxialTilt != nil {
		t.Errorf("body.AxialTilt: got %+v, want nil (should not be created by tidal lock)", body.AxialTilt)
	}
}

func TestGenerateTidalLock_ZedPrime_FullPath(t *testing.T) {
	// Zed Prime moon→planet path:
	//   1. EvaluateTidalLockDMs returns DM+7 for moon→planet case.
	//   2. SelectHighestDMCase picks moon→planet (no other cases applicable).
	//   3. RollTidalLockStatus: 2D=6 + 7 = 13 (1:1 lock pending).
	//   4. Verification: 2D=12 (natural 12) → reroll status with no DMs.
	//   5. Reroll: 2D=4 → FinalResult=4 → day × 2.
	//
	// All combined, the scripted roll list:
	//   2D for status: 6
	//   2D for verification: 12
	//   2D for status reroll (no DMs): 4

	moonRef := &Body{
		Kind:         BodyMoon,
		SizeCode:     "5",
		OrbitPD:      22,
		Retrograde:   true,
		Eccentricity: 0.25,
	}
	parent := &Body{}
	parent.Kind = BodyGasGiant
	parent.MassEarth = 1200
	parent.Orbit = 1.06

	body := &Body{}
	body.Kind = BodyTerrestrial
	body.SizeCode = "5"
	body.Eccentricity = 0.25
	body.AxialTilt = &AxialTilt{Degrees: 73.65, BaselineDegrees: 73.65}
	body.DayLength = &DayLength{SiderealHours: 42.37, BaselineSiderealHours: 42.37}
	body.Period = Period{Years: 0.072, Hours: 0.072 * 8766} // ~26 days for Zed's moon orbit

	sys := stars.System{Primary: stars.Star{Mass: 0.918, AgeGyr: 6.3}}

	r := roller.NewScripted(6, 12, 4)
	tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, body.Period.Hours)
	if err != nil {
		t.Fatal(err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock")
	}
	if tl.Case != TidalLockCaseMoonToPlanet {
		t.Errorf("Case: got %v, want MoonToPlanet", tl.Case)
	}
	if tl.InitialResult != 13 {
		t.Errorf("InitialResult: got %d, want 13", tl.InitialResult)
	}
	if !tl.VerificationFired {
		t.Error("expected VerificationFired=true")
	}
	if tl.FinalResult != 4 {
		t.Errorf("FinalResult: got %d, want 4", tl.FinalResult)
	}
	if tl.LockRatio != "" {
		t.Errorf("LockRatio: got %q, want empty (broken by verification)", tl.LockRatio)
	}
	if math.Abs(body.DayLength.SiderealHours-84.74) > 0.01 {
		t.Errorf("body day length: got %v, want 84.74", body.DayLength.SiderealHours)
	}
}

func TestGenerateTidalLock_PlutoCharon_PlanetLockedToMoon(t *testing.T) {
	// Synthetic Pluto/Charon: small planet (Size 3) with a Size 1 moon at orbit 5 PD.
	// Pluto-side check: planet→moon case applies because the planet has a
	// significant moon. With a high-mass moon at close orbit, planet→moon DM
	// can rival or exceed planet→star, exercising the case 3 path.
	plutoMoon := Body{
		Kind:     BodyMoon,
		SizeCode: "1",
		OrbitPD:  5,
	}
	pluto := &Body{}
	pluto.Kind = BodyTerrestrial
	pluto.SizeCode = "3"
	pluto.Orbit = 30 // far from sun
	pluto.Eccentricity = 0.05
	pluto.AxialTilt = &AxialTilt{Degrees: 0}
	pluto.DayLength = &DayLength{SiderealHours: 24, BaselineSiderealHours: 24}
	pluto.Period = Period{Years: 248, Hours: 248 * 8766}
	pluto.Children = []*Body{&plutoMoon}

	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}

	// DM math: planet→star = -63 (filtered ≤-10); planet→moon = +3 (common)
	// + (-10 base + Size 1 + 4 [pd≤10]) = -2. Roll 7 + (-2) = 5 → DayLengthMultiplier=3.
	r := roller.NewScripted(7)
	tl, err := GenerateTidalLock(r, pluto, nil, sys, nil, pluto.Period.Hours)
	if err != nil {
		t.Fatal(err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock for planet→moon path (DM=-2, roll=7, result=5)")
	}
	if tl.Case != TidalLockCasePlanetToMoon {
		t.Errorf("Case: got %v, want PlanetToMoon", tl.Case)
	}
	if tl.FinalResult != 5 {
		t.Errorf("FinalResult: got %d, want 5", tl.FinalResult)
	}
	if math.Abs(pluto.DayLength.SiderealHours-72.0) > 0.01 {
		t.Errorf("day length: got %v, want 72.0 (24 × 3)", pluto.DayLength.SiderealHours)
	}
}
