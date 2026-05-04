package worlds

import (
	"testing"

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
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
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
	moonRef := &Moon{
		SizeCode:     "5",
		OrbitPD:      22,
		Retrograde:   true,
		Eccentricity: 0.25,
	}
	parent := &DetailedPlacement{}
	parent.Body = BodyGasGiant
	parent.MassEarth = 1200
	parent.Orbit = 1.06

	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
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

func TestEvaluateTidalLockDMs_PlanetToMoon_OnlyIfHasSignificantMoon(t *testing.T) {
	// Planet→moon case is absent when the planet has no significant (Size 1+) moons.
	body := &DetailedPlacement{SizeCode: "3"}
	body.Body = BodyTerrestrial
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
	body := &DetailedPlacement{SizeCode: "5"}
	body.Body = BodyTerrestrial
	body.AxialTilt = &AxialTilt{Degrees: 0}
	body.Eccentricity = 0.0
	body.Orbit = 5.0
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 5.0}}

	dms := EvaluateTidalLockDMs(body, sys, nil, nil)
	if _, ok := dms[TidalLockCaseMoonToPlanet]; ok {
		t.Errorf("moon→planet should not apply to planets, got dms=%+v", dms)
	}
}
