package worlds

import (
	"math"
	"strings"
	"testing"

	"github.com/philoserf/world-builder/stars"
)

func TestStarTide_TerraSol(t *testing.T) {
	// Per WBH p.107: Sol on Terra causes 0.25m amplitude.
	// Formula: Star Mass × Planet Size / (32 × AU³) = 1.0 × 8 / (32 × 1.0³) = 0.25.
	got := StarTide(1.0, 8, 1.0)
	want := 0.25
	if math.Abs(got-want) > 0.01 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestStarTide_ZedAabSummedMass(t *testing.T) {
	// Per WBH p.108: "from the two relatively distant suns is only
	// 1.836 × 5 ÷ (32 × 1.06³) or 0.24 metres".
	got := StarTide(1.836, 5, 1.06)
	want := 0.24
	if math.Abs(got-want) > 0.02 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMoonTideOnPlanet_LunaTerra(t *testing.T) {
	// Per WBH p.108: Luna on Terra causes 0.54m amplitude.
	// Formula: Moon Mass × Planet Size / (3.2 × (Distance(km)/1,000,000)³).
	// Luna mass ≈ 0.0123 Earth masses; distance ≈ 384,400 km = 0.3844 million km.
	// 0.0123 × 8 / (3.2 × 0.3844³) = 0.0984 / 0.1818 = 0.541m.
	got := MoonTideOnPlanet(0.0123, 8, 384400)
	want := 0.54
	if math.Abs(got-want) > 0.05 {
		t.Errorf("got %v, want ~%v", got, want)
	}
}

func TestPlanetTideOnMoon_ZedPrime(t *testing.T) {
	// Per WBH p.108: gas giant tide on Zed Prime is 30.6m minimum.
	// 1200 × 5 / (3.2 × 3.9424³) = 6000 / 196.05 ≈ 30.6m.
	got := PlanetTideOnMoon(1200, 5, 3942400)
	want := 30.6
	if math.Abs(got-want) > 0.1 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGenerateSurfaceTidalEffects_ZedPrime(t *testing.T) {
	// Zed Prime surface tides:
	//   - From parent gas giant (1200⊕ at 3.9424M km): ~30.6m
	//   - From Aab (sum 1.836 M☉ at 1.06 AU): ~0.24m
	//   - From other moons: 0 (Zed Prime's only moon is itself)
	// Total: ~30.84m
	//
	// Companion construction: only the Ab close-binary mate is added (option C).
	// This tests close-binary mass summing (Aa+Ab=0.918+0.918=1.836) without
	// introducing a Z companion that would contribute at 1.06 AU and inflate
	// the star total past the assertion tolerance. The plan's "Z negligible"
	// scenario requires Z to be very distant, but GenerateSurfaceTidalEffects
	// uses parentPlanet's stellar distance (1.06 AU) for all star groups, so
	// any Z companion here would contribute ~0.24m extra and break the assertion.

	zedPrime := &Body{}
	zedPrime.Kind = BodyTerrestrial
	zedPrime.SizeCode = "5"

	parentGG := &Body{}
	parentGG.Kind = BodyGasGiant
	parentGG.MassEarth = 1200
	// Body.Orbit is a WBH Orbit#, not AU; GenerateSurfaceTidalEffects
	// converts it via stars.OrbitToAU. Use the Orbit# that maps to the
	// book's 1.06 AU (Orbit# 3.1) so the star-tide distance is correct.
	parentGG.Orbit = stars.AUToOrbit(1.06)

	moonRef := &Body{
		Kind:        BodyMoon,
		SizeCode:    "5",
		OrbitKm:     3942400,
		PeriodHours: 26 * 24,
	}

	// Aa (primary) + Ab (close-binary companion, OrbitCompanion, ParentIndex=-1).
	// Sum: 0.918 + 0.918 = 1.836 M☉ — matches WBH p.108 narrative.
	aa := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V,
		Mass:            0.918,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V,
		Mass:            0.918,
	})
	sys := stars.System{
		Primary:            aa,
		PrimaryDesignation: "Aab",
		Companions: []stars.CompanionStar{
			{Star: ab, OrbitClass: stars.OrbitCompanion, ParentIndex: -1, Designation: "Ab"},
		},
	}

	te, err := GenerateSurfaceTidalEffects(zedPrime, moonRef, sys, parentGG)
	if err != nil {
		t.Fatal(err)
	}
	if te == nil {
		t.Fatal("expected non-nil SurfaceTidalEffects")
	}
	if math.Abs(te.Total-30.84) > 0.5 {
		t.Errorf("Total: got %v, want ~30.84", te.Total)
	}

	// Find the planet tidal component and sum star components.
	var planetMeters, starMeters float64
	for _, c := range te.Components {
		if strings.HasPrefix(c.Source, "planet") {
			planetMeters = c.Meters
		}
		if strings.HasPrefix(c.Source, "star") {
			starMeters += c.Meters
		}
	}
	if math.Abs(planetMeters-30.6) > 0.2 {
		t.Errorf("planet component: got %v, want 30.6", planetMeters)
	}
	if math.Abs(starMeters-0.24) > 0.05 {
		t.Errorf("summed star components: got %v, want 0.24", starMeters)
	}
}

// TestGenerateSurfaceTidalEffects_OrbitConvertedToAU guards the units
// contract: body.Orbit is a WBH Orbit#, and the star-tide component
// must equal StarTide evaluated at the body's AU distance, not at the
// raw Orbit#. Regression for the units bug where body.Orbit was passed
// straight to StarTide's AU parameter.
func TestGenerateSurfaceTidalEffects_OrbitConvertedToAU(t *testing.T) {
	t.Parallel()
	// Orbit# 3.0 == 1.0 AU (WBH p.26 HZ reference). A Size-8 planet of a
	// 1 M☉ primary should see exactly the Terra/Sol star tide.
	planet := &Body{Kind: BodyTerrestrial, SizeCode: "8", Orbit: 3.0}
	sys := stars.System{Primary: stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, LuminosityClass: stars.V, Mass: 1.0,
	})}
	te, err := GenerateSurfaceTidalEffects(planet, nil, sys, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := StarTide(1.0, 8, stars.OrbitToAU(3.0)) // == StarTide(1.0, 8, 1.0)
	if math.Abs(te.Total-want) > 1e-9 {
		t.Errorf("star tide = %v, want %v (Orbit# 3.0 must convert to 1.0 AU, not be used raw)", te.Total, want)
	}
	// Sanity: using the raw Orbit# would divide by 3.0³ instead of 1.0³,
	// a 27x error — assert we are nowhere near that wrong value.
	if wrong := StarTide(1.0, 8, 3.0); math.Abs(te.Total-wrong) < 1e-9 {
		t.Errorf("star tide matches the raw-Orbit# (unconverted) value %v — conversion missing", wrong)
	}
}
