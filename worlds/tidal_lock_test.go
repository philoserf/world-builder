package worlds

import (
	"math"
	"slices"
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
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

func TestEvaluateTidalLockDMs_PlanetToMoon_GasGiantExcluded(t *testing.T) {
	// Gas giant should not have Planet→Moon case, even if it has a locked moon.
	// The planet must be terrestrial per WBH p.107.
	gg := &Body{
		Kind:     BodyGasGiant,
		SizeCode: "M",
		Children: []*Body{
			{
				Kind:      BodyMoon,
				SizeCode:  "5",
				OrbitPD:   10,
				TidalLock: &TidalLock{LockRatio: "1:1"},
			},
		},
	}
	gg.AxialTilt = &AxialTilt{Degrees: 0}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}

	dms := EvaluateTidalLockDMs(gg, sys, nil, nil)
	if _, ok := dms[TidalLockCasePlanetToMoon]; ok {
		t.Errorf("gas giant should not have Planet→Moon case; got DMs: %v", dms)
	}
}

func TestEvaluateTidalLockDMs_PlanetToMoon_NoLockedMoonExcluded(t *testing.T) {
	// Planet with no locked moons should not have Planet→Moon case.
	// The planet must have at least one moon in 1:1 or 3:2 lock per WBH p.107.
	p := &Body{
		Kind:     BodyTerrestrial,
		SizeCode: "8",
		Children: []*Body{
			{
				Kind:      BodyMoon,
				SizeCode:  "5",
				OrbitPD:   10,
				TidalLock: &TidalLock{LockRatio: ""}, // unlocked
			},
		},
	}
	p.AxialTilt = &AxialTilt{Degrees: 0}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}

	dms := EvaluateTidalLockDMs(p, sys, nil, nil)
	if _, ok := dms[TidalLockCasePlanetToMoon]; ok {
		t.Errorf("planet with no locked moons should not have Planet→Moon case; got DMs: %v", dms)
	}
}

func TestEvaluateTidalLockDMs_PlanetToMoon_TerrestrialWithLockedMoonIncluded(t *testing.T) {
	// Terrestrial planet with a locked moon should have Planet→Moon case.
	p := &Body{
		Kind:     BodyTerrestrial,
		SizeCode: "8",
		Children: []*Body{
			{
				Kind:      BodyMoon,
				SizeCode:  "5",
				OrbitPD:   30,
				TidalLock: &TidalLock{LockRatio: "1:1"},
			},
		},
	}
	p.AxialTilt = &AxialTilt{Degrees: 0}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}

	dms := EvaluateTidalLockDMs(p, sys, nil, nil)
	if _, ok := dms[TidalLockCasePlanetToMoon]; !ok {
		t.Errorf("terrestrial with locked moon should have Planet→Moon case; got DMs: %v", dms)
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

func TestHasLockedMoon(t *testing.T) {
	locked := &Body{TidalLock: &TidalLock{LockRatio: "1:1"}}
	threeTwo := &Body{TidalLock: &TidalLock{LockRatio: "3:2"}}
	unlocked := &Body{TidalLock: &TidalLock{LockRatio: ""}}
	noLockField := &Body{}
	cases := []struct {
		name string
		body Body
		want bool
	}{
		{"no children false", Body{}, false},
		{"all unlocked false", Body{Children: []*Body{unlocked}}, false},
		{"nil TidalLock false", Body{Children: []*Body{noLockField}}, false},
		{"3:2 locked true", Body{Children: []*Body{threeTwo}}, true},
		{"1:1 locked true", Body{Children: []*Body{locked}}, true},
		{"mixed with one locked true", Body{Children: []*Body{unlocked, locked}}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasLockedMoon(&c.body); got != c.want {
				t.Errorf("hasLockedMoon = %v, want %v", got, c.want)
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
	//   2. SelectHighestDMCases picks moon→planet (no other cases applicable).
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

func TestApplyTidalLockEffect_PlanetToMoon_OneToOne_UsesMoonPeriod(t *testing.T) {
	// When ApplyTidalLockEffect lands a 1:1 lock for the Planet→Moon case,
	// the planet's new SiderealHours should equal yearHours (which
	// GenerateTidalLock will pre-compute as the moon's PeriodHours).
	// We drive ApplyTidalLockEffect directly with yearHours=720 (moon period).
	//
	// initialResult=12 → verification 2D consumed (roll=7, not 12, so no reroll).
	// Axial tilt=0° → no tilt reroll. Eccentricity=0.05 → no ecc reroll.
	// Expected: SiderealHours = 720 (moon period), not 8766 (stellar year).
	planet := &Body{
		Kind:      BodyTerrestrial,
		SizeCode:  "8",
		Period:    Period{Hours: 8766},
		DayLength: &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		AxialTilt: &AxialTilt{Degrees: 0},
	}
	// verification 2D = 7 (not 12, so no natural-12 reroll)
	r := roller.NewScripted(7)
	moonPeriodHours := 720.0
	tl, err := ApplyTidalLockEffect(r, planet, nil, TidalLockCasePlanetToMoon, 12, moonPeriodHours)
	if err != nil {
		t.Fatalf("ApplyTidalLockEffect: %v", err)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio = %q, want 1:1", tl.LockRatio)
	}
	if math.Abs(planet.DayLength.SiderealHours-720.0) > 0.01 {
		t.Errorf("planet SiderealHours = %v, want 720 (moon period), not 8766 (stellar year)",
			planet.DayLength.SiderealHours)
	}
}

func TestGenerateTidalLock_PlanetToMoon_UsesMoonPeriod(t *testing.T) {
	// Construct a planet with a locked moon; scripted dice produce a
	// Planet→Moon 3:2 lock (result=11).
	//
	// DM math:
	//   common: Size 1 → +ceil(1/3)=+1; ecc=0 no DM; tilt=0 no DM;
	//           no atmosphere; age=5Gyr → +2. Common = +3.
	//   planetToMoonDMs: base=-10; moon Size 1 → +1; pd=5 → +4 (pd≤10);
	//                    one significant moon → no extra DM. PTM = -5.
	//   Net DM = 3 + (-5) = -2.
	//
	// Roll 2D=13 → initialResult = 13 + (-2) = 11 → LockRatio "3:2".
	// Axial tilt=0° (≤3°) → no tilt reroll consumed.
	// yearHours for PlanetToMoon case = moon.PeriodHours = 720.
	// Expected SiderealHours = 720 × 2/3 = 480.
	plutoMoon := Body{
		Kind:        BodyMoon,
		SizeCode:    "1",
		OrbitPD:     5,
		PeriodHours: 720,
		TidalLock:   &TidalLock{LockRatio: "1:1"},
	}
	planet := &Body{
		Kind:      BodyTerrestrial,
		SizeCode:  "1",
		Orbit:     30,
		Period:    Period{Hours: 8766},
		DayLength: &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		AxialTilt: &AxialTilt{Degrees: 0},
		Children:  []*Body{&plutoMoon},
	}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}
	r := roller.NewScripted(13) // 2D=13 → result=11 → 3:2 lock
	tl, err := GenerateTidalLock(r, planet, nil, sys, nil, planet.Period.Hours)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock")
	}
	if tl.Case != TidalLockCasePlanetToMoon {
		t.Errorf("Case: got %v, want PlanetToMoon", tl.Case)
	}
	if tl.LockRatio != "3:2" {
		t.Errorf("LockRatio: got %q, want 3:2", tl.LockRatio)
	}
	if math.Abs(planet.DayLength.SiderealHours-480.0) > 0.01 {
		t.Errorf("planet SiderealHours = %v, want 480 (720×2/3, moon period), not %v (stellar year×2/3)",
			planet.DayLength.SiderealHours, planet.Period.Hours*2/3)
	}
}

func TestGenerateTidalLock_PlutoCharon_PlanetLockedToMoon(t *testing.T) {
	// Synthetic Pluto/Charon: small planet (Size 3) with a Size 1 moon at orbit 5 PD.
	// Pluto-side check: planet→moon case applies because the planet has a
	// significant moon. With a high-mass moon at close orbit, planet→moon DM
	// can rival or exceed planet→star, exercising the case 3 path.
	//
	// Fidelity note (project finding, #59): with REAL Pluto/Charon parameters
	// (Pluto SizeCode "1", ecc 0.2488, Charon OrbitPD 7.38, age 4.6 Gyr) the
	// PlanetToMoon DM works out to: common −1 + specific (−10 base + Size 1
	// + pd≤10 → +4) = −6. Maximum reachable result is 12 + (−6) = 6, which
	// produces a DayLengthMultiplier ×5 — not the canonical mutual 1:1 lock
	// observed in reality. WBH's procedure is structurally incapable of
	// producing Pluto↔Charon's actual mutual lock from real parameters; the
	// synthetic Size-3 / orbit-5-PD fixture here is a procedural exerciser,
	// not a real-world reproduction.
	plutoMoon := Body{
		Kind:      BodyMoon,
		SizeCode:  "1",
		OrbitPD:   5,
		TidalLock: &TidalLock{LockRatio: "1:1"}, // assume moon already locked (WBH p.107 precondition)
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

// TestClosestLockedSignificantMoon_PicksLockedNotClosest verifies that the
// function picks the locked moon over the closer unlocked moon.
// Per WBH p.107 the planet locks to one of its already-locked moons; the
// DM math and day length must reference that locked moon, not the closest.
func TestClosestLockedSignificantMoon_PicksLockedNotClosest(t *testing.T) {
	// Planet has two moons:
	//   - inner: OrbitPD=5, SizeCode="5", unlocked (TidalLock.LockRatio "")
	//   - outer: OrbitPD=30, SizeCode="3", 1:1 locked
	// The book partner is the outer (locked) moon, not the inner.
	inner := &Body{SizeCode: "5", OrbitPD: 5, TidalLock: &TidalLock{LockRatio: ""}}
	outer := &Body{SizeCode: "3", OrbitPD: 30, TidalLock: &TidalLock{LockRatio: "1:1"}}
	planet := &Body{Children: []*Body{inner, outer}}
	got := closestLockedSignificantMoon(planet)
	if got != outer {
		t.Errorf("got %v, want outer (the locked moon)", got)
	}
}

// TestClosestLockedSignificantMoon_NoLockedReturnsNil verifies nil is returned
// when no significant moon has a 1:1 or 3:2 lock.
func TestClosestLockedSignificantMoon_NoLockedReturnsNil(t *testing.T) {
	inner := &Body{SizeCode: "5", OrbitPD: 5} // nil TidalLock
	outer := &Body{SizeCode: "3", OrbitPD: 30, TidalLock: &TidalLock{LockRatio: ""}}
	planet := &Body{Children: []*Body{inner, outer}}
	if got := closestLockedSignificantMoon(planet); got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

// TestClosestLockedSignificantMoon_MultipleLockedPicksClosest verifies that
// when multiple locked moons exist, the one with the smallest OrbitPD wins.
func TestClosestLockedSignificantMoon_MultipleLockedPicksClosest(t *testing.T) {
	far := &Body{SizeCode: "3", OrbitPD: 30, TidalLock: &TidalLock{LockRatio: "1:1"}}
	near := &Body{SizeCode: "5", OrbitPD: 10, TidalLock: &TidalLock{LockRatio: "3:2"}}
	planet := &Body{Children: []*Body{far, near}}
	got := closestLockedSignificantMoon(planet)
	if got != near {
		t.Errorf("got %v (OrbitPD=%v), want near (OrbitPD=10)", got, got.OrbitPD)
	}
}

func TestApplyTidalLockEffect_MoonPeriodGuard_KeepsOneToOneLock(t *testing.T) {
	// Setup: MoonToPlanet case, initialResult ≥ 12 (locks), natural-12
	// verification fires (2D=12), rerolled status is result 8 (1D×20×24).
	// Moon's yearHours = 480 (20-day orbit). Rerolled day length:
	//   1D=3 → 3×20×24 = 1440 hours. 1440 > 480 → guard fires; lock holds.
	moon := &Body{
		Kind:        BodyMoon,
		SizeCode:    "5",
		DayLength:   &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		AxialTilt:   &AxialTilt{Degrees: 0},
		PeriodHours: 480,
	}
	// Scripted rolls (in order):
	//   verification 2D: 12 (natural twelve triggers reroll)
	//   reroll status 2D + DM=0: 8 (result 8, the 1D×20×24 branch)
	//   rerolledDayLength 1D: 3 (the would-be day-length probe)
	r := roller.NewScripted(12, 8, 3)
	tl, err := ApplyTidalLockEffect(r, moon, moon, TidalLockCaseMoonToPlanet, 12, 480)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !tl.VerificationFired {
		t.Error("VerificationFired = false, want true (natural-12 fired)")
	}
	if tl.FinalResult != 12 {
		t.Errorf("FinalResult = %d, want 12 (lock held by moon-period guard)", tl.FinalResult)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio = %q, want 1:1", tl.LockRatio)
	}
}

func TestApplyTidalLockEffect_MoonPeriodGuard_FallsThroughWhenWouldFit(t *testing.T) {
	// Setup: MoonToPlanet, verification 12, rerolled result 7 (1D×5×24).
	// 1D=2 → 2×5×24 = 240 hours. yearHours=480 → 240 ≤ 480 → guard does
	// NOT fire; FinalResult takes the rerolled value (7).
	//
	// After the fix: only 3 rolls are needed (probe = commit). The probe's
	// 1D=2 is reused as the committed day length; the effect switch skips
	// its own roll. DayLength.SiderealHours must equal 240 (2×5×24).
	moon := &Body{
		Kind:        BodyMoon,
		SizeCode:    "5",
		DayLength:   &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		AxialTilt:   &AxialTilt{Degrees: 0},
		PeriodHours: 480,
	}
	// Scripted rolls:
	//   verification 2D = 12
	//   reroll status 2D + 0 = 7
	//   probe 1D = 2 (=> 240h, not exceeding 480h — guard doesn't fire; value reused as commit)
	r := roller.NewScripted(12, 7, 2)
	tl, err := ApplyTidalLockEffect(r, moon, moon, TidalLockCaseMoonToPlanet, 12, 480)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if tl.FinalResult != 7 {
		t.Errorf("FinalResult = %d, want 7 (rerolled value committed)", tl.FinalResult)
	}
	if tl.LockRatio != "" {
		t.Errorf("LockRatio = %q, want empty (lock broken)", tl.LockRatio)
	}
	if math.Abs(moon.DayLength.SiderealHours-240) > 0.01 {
		t.Errorf("DayLength.SiderealHours = %v, want 240 (probe value committed)", moon.DayLength.SiderealHours)
	}
}

func TestApplyTidalLockEffect_MoonPeriodGuard_NotAppliedToPlanetToStar(t *testing.T) {
	// Same scenario but case = PlanetToStar — guard MUST NOT fire.
	// FinalResult should reflect the rerolled value (8), not the initial (12).
	planet := &Body{
		Kind:      BodyTerrestrial,
		SizeCode:  "8",
		DayLength: &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		AxialTilt: &AxialTilt{Degrees: 0},
		Period:    Period{Hours: 8766},
	}
	// verification 2D=12, reroll status 2D+0=8, probe 1D=3 (=> 1440h).
	// PlanetToStar: guard doesn't apply; rerolled=8 is in 7-10 range so
	// tl.NewSiderealHours=1440 is set from the probe; effect switch skips
	// its own 1D roll. Three rolls total.
	r := roller.NewScripted(12, 8, 3)
	tl, err := ApplyTidalLockEffect(r, planet, nil, TidalLockCasePlanetToStar, 12, 8766)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if tl.FinalResult != 8 {
		t.Errorf("FinalResult = %d, want 8 (no guard for PlanetToStar)", tl.FinalResult)
	}
}

// TestApplyTidalLockEffect_MoonPeriodGuard_ProbeAndCommitUseSameRoll is a
// regression test for the divergent-roll bug: previously, the guard probe
// rolled one 1D and the effect commit rolled a separate 1D, so a low probe
// (passes guard) could be followed by a high commit (would set a day length
// exceeding the moon's orbit). Now probe and commit use the SAME 1D — the
// committed day length is exactly what the guard inspected.
//
// This test scripts ONE 1D = 6 for result 7:
//   - Probe: 6 × 5 × 24 = 720h.
//   - 720 > moon period 480 → guard FIRES → 1:1 lock holds.
//   - FinalResult = 12 (initial), LockRatio = "1:1".
//   - DayLength = yearHours = 480h (1:1 lock day-length effect applies).
//
// If a second 1D were consumed by the effect switch, the scripted roller
// would panic on exhaustion — that alone proves the fix.
func TestApplyTidalLockEffect_MoonPeriodGuard_ProbeAndCommitUseSameRoll(t *testing.T) {
	moon := &Body{
		Kind:        BodyMoon,
		SizeCode:    "5",
		DayLength:   &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		AxialTilt:   &AxialTilt{Degrees: 0},
		PeriodHours: 480,
	}
	// verification 2D=12, reroll status 2D+0=7, probe 1D=6 (=> 720h > 480h → guard fires)
	r := roller.NewScripted(12, 7, 6)
	tl, err := ApplyTidalLockEffect(r, moon, moon, TidalLockCaseMoonToPlanet, 12, 480)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if tl.FinalResult != 12 {
		t.Errorf("FinalResult = %d, want 12 (guard fired, lock held)", tl.FinalResult)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio = %q, want 1:1", tl.LockRatio)
	}
	// 1:1 lock sets DayLength = yearHours (480h). Guard only prevents the rerolled
	// value from being committed — the original 1:1 effect still applies.
	if math.Abs(moon.DayLength.SiderealHours-480) > 0.01 {
		t.Errorf("DayLength.SiderealHours = %v, want 480 (1:1 lock = yearHours)", moon.DayLength.SiderealHours)
	}
}

// TestApplyTidalLockEffect_MoonPeriodGuard_LowProbeCommitsProbeValue is the
// inverse: probe rolls 1D=1 (120h ≤ 480h → guard doesn't fire), and the
// committed day length must equal 120h (the probe value reused, not a
// separately-rolled value that could have differed).
func TestApplyTidalLockEffect_MoonPeriodGuard_LowProbeCommitsProbeValue(t *testing.T) {
	moon := &Body{
		Kind:        BodyMoon,
		SizeCode:    "5",
		DayLength:   &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		AxialTilt:   &AxialTilt{Degrees: 0},
		PeriodHours: 480,
	}
	// verification 2D=12, reroll status 2D+0=7, probe 1D=1 (=> 120h ≤ 480h → guard doesn't fire)
	r := roller.NewScripted(12, 7, 1)
	tl, err := ApplyTidalLockEffect(r, moon, moon, TidalLockCaseMoonToPlanet, 12, 480)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if tl.FinalResult != 7 {
		t.Errorf("FinalResult = %d, want 7 (guard didn't fire)", tl.FinalResult)
	}
	if tl.LockRatio != "" {
		t.Errorf("LockRatio = %q, want empty (lock broken)", tl.LockRatio)
	}
	if math.Abs(moon.DayLength.SiderealHours-120) > 0.01 {
		t.Errorf("DayLength.SiderealHours = %v, want 120 (probe 1D=1 × 5 × 24, committed)", moon.DayLength.SiderealHours)
	}
}

func TestSelectHighestDMCases_OrderedTiedCases(t *testing.T) {
	// PlanetToStar tied with PlanetToMoon at DM=5; MoonToPlanet absent.
	// Expected return order per p.106 priority: PlanetToMoon (moon-cases first), PlanetToStar.
	dms := map[TidalLockCase]int{
		TidalLockCasePlanetToStar: 5,
		TidalLockCasePlanetToMoon: 5,
	}
	cases, dm := SelectHighestDMCases(dms)
	if dm != 5 {
		t.Errorf("dm = %d, want 5", dm)
	}
	want := []TidalLockCase{TidalLockCasePlanetToMoon, TidalLockCasePlanetToStar}
	if !slices.Equal(cases, want) {
		t.Errorf("cases = %v, want %v", cases, want)
	}
}

func TestSelectHighestDMCases_SingleCase(t *testing.T) {
	dms := map[TidalLockCase]int{TidalLockCaseMoonToPlanet: 8}
	cases, dm := SelectHighestDMCases(dms)
	if dm != 8 {
		t.Errorf("dm = %d, want 8", dm)
	}
	if len(cases) != 1 || cases[0] != TidalLockCaseMoonToPlanet {
		t.Errorf("cases = %v, want [MoonToPlanet]", cases)
	}
}

func TestSelectHighestDMCases_AllFiltered(t *testing.T) {
	dms := map[TidalLockCase]int{TidalLockCasePlanetToStar: -11}
	cases, _ := SelectHighestDMCases(dms)
	if len(cases) != 0 {
		t.Errorf("expected no cases, got %v", cases)
	}
}

func TestSelectHighestDMCases_AllThreeTied(t *testing.T) {
	// All three cases at DM=7 — order must be MoonToPlanet, PlanetToMoon, PlanetToStar.
	dms := map[TidalLockCase]int{
		TidalLockCasePlanetToStar: 7,
		TidalLockCaseMoonToPlanet: 7,
		TidalLockCasePlanetToMoon: 7,
	}
	cases, dm := SelectHighestDMCases(dms)
	if dm != 7 {
		t.Errorf("dm = %d, want 7", dm)
	}
	want := []TidalLockCase{TidalLockCaseMoonToPlanet, TidalLockCasePlanetToMoon, TidalLockCasePlanetToStar}
	if !slices.Equal(cases, want) {
		t.Errorf("cases = %v, want %v", cases, want)
	}
}

func TestSelectHighestDMCases_BestOnlyNotTied(t *testing.T) {
	// When DMs differ, only the highest-DM case is returned.
	dms := map[TidalLockCase]int{
		TidalLockCasePlanetToStar: 3,
		TidalLockCasePlanetToMoon: 7,
		TidalLockCaseMoonToPlanet: 5,
	}
	cases, dm := SelectHighestDMCases(dms)
	if dm != 7 {
		t.Errorf("dm = %d, want 7", dm)
	}
	if len(cases) != 1 || cases[0] != TidalLockCasePlanetToMoon {
		t.Errorf("cases = %v, want [PlanetToMoon]", cases)
	}
}

// TestGenerateTidalLock_TiedCases_RollsAllTakesHighest verifies that when two
// cases tie on DM, GenerateTidalLock rolls both (in priority order) and applies
// the highest adjusted result. Without the cascade, only one roll fires and the
// scripted roller panics on exhaustion — proving the fix.
//
// Fixture: terrestrial planet, Size 3, Orbit 2.0, star mass 1.0, age 5 Gyr,
// eccentricity 0, tilt 0, no atmosphere; one Size 1 moon at OrbitPD 10 with a
// 1:1 lock (enabling the Planet→Moon case).
//
// DM trace:
//
//	common: Size 3 → +ceil(3/3)=+1; ecc 0→0; tilt 0→0; age 5Gyr→+2; total=+3
//	planet→star: base=-4; orbit 2.0 (1<orbit<3)→+1; mass 1.0 (≤1.0)→-1;
//	             1 star, no multi-star penalty; Size 1 moon → -1; total=-5
//	planet→moon: base=-10; moon Size 1 → +1; pd=10 (≤10)→+4; total=-5
//	Both cases: +3 + (-5) = -2  (tied)
//
// Dice:
//
//	Roll 1 (PlanetToMoon, moon-case first): 2D=6 → adjusted=6+(-2)=4
//	Roll 2 (PlanetToStar):                  2D=10 → adjusted=10+(-2)=8 (higher, wins)
//	Roll 3 (1D for result 8 day length):    1D=3 → 3×20×24=1440h
//
// Expected: Case=PlanetToStar, FinalResult=8, SiderealHours=1440.
// If only one roll fires (old code), the scripted roller panics → test fails.
func TestGenerateTidalLock_TiedCases_RollsAllTakesHighest(t *testing.T) {
	moon := &Body{
		Kind:        BodyMoon,
		SizeCode:    "1",
		OrbitPD:     10,
		PeriodHours: 1200, // ~50-day orbit; long enough for 1D=2×20×24=960h day
		TidalLock:   &TidalLock{LockRatio: "1:1"},
	}
	planet := &Body{
		Kind:         BodyTerrestrial,
		SizeCode:     "3",
		Orbit:        2.0,
		Period:       Period{Hours: 8766},
		Eccentricity: 0.0,
		DayLength:    &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		AxialTilt:    &AxialTilt{Degrees: 0},
		Children:     []*Body{moon},
	}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}

	// Verify fixture: both cases must tie at DM=-2.
	dms := EvaluateTidalLockDMs(planet, sys, nil, nil)
	ptmDM, ptmOK := dms[TidalLockCasePlanetToMoon]
	ptsDM, ptsOK := dms[TidalLockCasePlanetToStar]
	if !ptmOK || !ptsOK {
		t.Fatalf("fixture missing a case: PlanetToMoon=%v(%d) PlanetToStar=%v(%d)", ptmOK, ptmDM, ptsOK, ptsDM)
	}
	if ptmDM != ptsDM {
		t.Fatalf("fixture DMs must tie: PlanetToMoon=%d PlanetToStar=%d", ptmDM, ptsDM)
	}
	if ptmDM != -2 {
		t.Fatalf("expected tied DM=-2, got %d", ptmDM)
	}

	// Scripted rolls:
	//   2D for PlanetToMoon (first, moon-cases priority): 6 → adjusted=4
	//   2D for PlanetToStar (second):                    10 → adjusted=8 (wins)
	//   1D for result-8 day length:                       3 → 1440h
	r := roller.NewScripted(6, 10, 3)
	tl, err := GenerateTidalLock(r, planet, nil, sys, nil, planet.Period.Hours)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock")
	}
	if tl.Case != TidalLockCasePlanetToStar {
		t.Errorf("Case = %v, want PlanetToStar (higher adjusted result)", tl.Case)
	}
	if tl.FinalResult != 8 {
		t.Errorf("FinalResult = %d, want 8 (10 + DM -2)", tl.FinalResult)
	}
	if tl.NewSiderealHours != 1440 {
		t.Errorf("NewSiderealHours = %v, want 1440 (1D=3 × 20 × 24)", tl.NewSiderealHours)
	}
}

// TestGenerateTidalLock_TiedCases_FirstRollWins verifies the inverse: when the
// first (moon-priority) case rolls higher, it wins and the second case's result
// is discarded (but the roll still fires — both 2D rolls must be consumed).
//
// Same fixture as above. Dice reversed:
//
//	Roll 1 (PlanetToMoon): 2D=10 → adjusted=8 (wins)
//	Roll 2 (PlanetToStar): 2D=6  → adjusted=4
//	Roll 3 (1D for result-8 day from PlanetToMoon): 1D=2 → 960h
//
// Expected: Case=PlanetToMoon, FinalResult=8, SiderealHours=960.
func TestGenerateTidalLock_TiedCases_FirstRollWins(t *testing.T) {
	moon := &Body{
		Kind:        BodyMoon,
		SizeCode:    "1",
		OrbitPD:     10,
		PeriodHours: 1200, // ~50-day orbit; long enough for 1D=2×20×24=960h day
		TidalLock:   &TidalLock{LockRatio: "1:1"},
	}
	planet := &Body{
		Kind:         BodyTerrestrial,
		SizeCode:     "3",
		Orbit:        2.0,
		Period:       Period{Hours: 8766},
		Eccentricity: 0.0,
		DayLength:    &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		AxialTilt:    &AxialTilt{Degrees: 0},
		Children:     []*Body{moon},
	}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}

	// Rolls:
	//   2D PlanetToMoon:  10 → adjusted=8 (wins)
	//   2D PlanetToStar:   6 → adjusted=4
	//   1D day (result 8): 2 → 960h
	r := roller.NewScripted(10, 6, 2)
	tl, err := GenerateTidalLock(r, planet, nil, sys, nil, planet.Period.Hours)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock")
	}
	if tl.Case != TidalLockCasePlanetToMoon {
		t.Errorf("Case = %v, want PlanetToMoon (first tied case wins when adjusted result higher)", tl.Case)
	}
	if tl.FinalResult != 8 {
		t.Errorf("FinalResult = %d, want 8 (10 + DM -2)", tl.FinalResult)
	}
	if tl.NewSiderealHours != 960 {
		t.Errorf("NewSiderealHours = %v, want 960 (1D=2 × 20 × 24)", tl.NewSiderealHours)
	}
	if planet.DayLength.YearDays < 0 {
		t.Errorf("YearDays = %v, want non-negative (PeriodHours fixture missing?)", planet.DayLength.YearDays)
	}
	if planet.DayLength.SolarHours < 0 {
		t.Errorf("SolarHours = %v, want non-negative", planet.DayLength.SolarHours)
	}
}

func TestRerolledDayLength(t *testing.T) {
	body := &Body{DayLength: &DayLength{SiderealHours: 24}}
	cases := []struct {
		name      string
		result    int
		dieValue  int // value the next 1D roll would return (0 if no 1D consumed)
		yearHours float64
		want      float64
	}{
		{"result 3 → 1.5× current", 3, 0, 0, 36},
		{"result 4 → 2× current", 4, 0, 0, 48},
		{"result 5 → 3× current", 5, 0, 0, 72},
		{"result 6 → 5× current", 6, 0, 0, 120},
		{"result 7 → 1D×5×24 (1D=3)", 7, 3, 0, 360},
		{"result 8 → 1D×20×24 (1D=2)", 8, 2, 0, 960},
		{"result 9 → 1D×10×24 (1D=4)", 9, 4, 0, 960},
		{"result 10 → 1D×50×24 (1D=5)", 10, 5, 0, 6000},
		{"result 11 (3:2) → 2/3 yearHours", 11, 0, 720, 480},
		{"result 12 (1:1) → yearHours", 12, 0, 720, 720},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var r roller.Roller
			if c.dieValue > 0 {
				r = roller.NewScripted(c.dieValue)
			} else {
				r = roller.NewScripted() // no rolls expected for 3-6, 11, 12
			}
			got := rerolledDayLength(r, c.result, body, c.yearHours)
			if got != c.want {
				t.Errorf("rerolledDayLength(%d, dieValue=%d, year=%g) = %g, want %g",
					c.result, c.dieValue, c.yearHours, got, c.want)
			}
		})
	}
}
