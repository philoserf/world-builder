package worlds

import (
	"math"
	"testing"

	"wbh/stars"
)

// TestMarkHZ_WindowInclusion: a placement at orbit 3.1 with group HZCO 3.3
// is in the HZ (range 2.3-4.3); orbits at 5.0 and 1.0 are not.
func TestMarkHZ_WindowInclusion(t *testing.T) {
	t.Parallel()

	// Compose a star with luminosity tuned so HZCO_Orbit ≈ 3.3.
	// HZCO_AU = sqrt(luminosity); HZCO_Orbit = AUToOrbit(HZCO_AU).
	// For HZCO_Orbit ≈ 3.3 → HZCO_AU ≈ 1.32 → luminosity ≈ 1.74.
	star := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 0},
		LuminosityClass: stars.V,
		Mass:            1.0,
	})
	// Override Luminosity directly since Compose computes it.
	star.Luminosity = 1.74
	g := Group{Designation: "Aab", Members: []stars.Star{star}}

	dps := []DetailedPlacement{
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.1, Group: g}}}}, // in HZ
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 5.0, Group: g}}}}, // outside (above)
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: g}}}}, // outside (below)
		{Placement: Placement{Body: BodyTerrestrial, AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.1, Group: g}}}}, // in HZ
	}

	if err := MarkHZ(dps); err != nil {
		t.Fatalf("MarkHZ err: %v", err)
	}
	wantHZ := []bool{true, false, false, true}
	for i, w := range wantHZ {
		if dps[i].HZ != w {
			t.Errorf("dps[%d].HZ (orbit %v) = %v, want %v", i, dps[i].Orbit, dps[i].HZ, w)
		}
	}
}

// TestMainworldCandidates_PlanetAndMoonCandidates: terrestrial planet in
// HZ + significant moons of any HZ planet (including gas-giant moons).
// Belts and gas giants themselves are not candidates.
func TestMainworldCandidates_PlanetAndMoonCandidates(t *testing.T) {
	t.Parallel()

	g := Group{Designation: "Aab"}
	dps := []DetailedPlacement{
		{
			Placement: Placement{
				Body:          BodyGasGiant,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 3.1, Group: g}},
			},
			Designation: "Aab IV",
			HZ:          true,
			GGClass:     GasGiantLarge,
			Moons: []Moon{
				{Designation: "Aab IV a", SizeCode: "2", DiameterKm: 3200},
				{Designation: "Aab IV d", SizeCode: "5", DiameterKm: 8000},
			},
		},
		{
			Placement: Placement{
				Body:          BodyTerrestrial,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 4.1, Group: g}},
			},
			Designation: "Aab VI",
			HZ:          true,
			SizeCode:    "A",
			DiameterKm:  16000,
		},
		{
			Placement: Placement{
				Body:          BodyTerrestrial,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 1.0, Group: g}},
			},
			Designation: "Aab I",
			HZ:          false, // outside HZ
			SizeCode:    "B",
			DiameterKm:  17600,
		},
		{
			Placement: Placement{
				Body:          BodyPlanetoidBelt,
				AnomalousSlot: AnomalousSlot{Slot: Slot{Orbit: 2.7, Group: g}},
			},
			Designation: "Aab PI",
			HZ:          false,
		},
	}
	sd := SystemDetail{Detailed: dps}

	got := MainworldCandidates(sd)

	// Expected: Aab IV a + Aab IV d (moons of HZ gas giant, ordered first
	// in form Notes column), then Aab VI (planet).
	// Aab I excluded (not in HZ). Belt excluded. Gas giant Aab IV itself excluded.
	wantDesigs := []string{"Aab IV a", "Aab IV d", "Aab VI"}
	if len(got) != len(wantDesigs) {
		t.Fatalf("len(MainworldCandidates) = %d, want %d (got: %v)", len(got), len(wantDesigs), got)
	}
	for i, w := range wantDesigs {
		if got[i].Designation != w {
			t.Errorf("got[%d].Designation = %q, want %q", i, got[i].Designation, w)
		}
	}

	// Spot-check moon candidate (Aab IV a).
	moon := got[0]
	if !moon.IsMoon {
		t.Error("got[0].IsMoon = false, want true")
	}
	if moon.ParentDesignation != "Aab IV" {
		t.Errorf("got[0].ParentDesignation = %q, want \"Aab IV\"", moon.ParentDesignation)
	}
	if math.Abs(moon.Orbit-3.1) > 1e-9 {
		t.Errorf("got[0].Orbit = %v, want 3.1 (parent's orbit)", moon.Orbit)
	}

	// Spot-check planet candidate (Aab VI).
	planet := got[2]
	if planet.IsMoon {
		t.Error("got[2].IsMoon = true, want false (planet)")
	}
	if planet.ParentDesignation != "" {
		t.Errorf("got[2].ParentDesignation = %q, want \"\"", planet.ParentDesignation)
	}
	if planet.HostStarGroup != "Aab" {
		t.Errorf("got[2].HostStarGroup = %q, want \"Aab\"", planet.HostStarGroup)
	}
}

func TestPickMainworld_SophontWins_OverHigherHabitability(t *testing.T) {
	// Body A: Habitability 8, no sophont.
	// Body B: Habitability 4, HasNativeSophont=true.
	// Sophont wins → B.
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 8}, Biology: &Biology{}},
		{Designation: "B", Habitability: &Habitability{Rating: 4}, Biology: &Biology{HasNativeSophont: true}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "B" {
		t.Errorf("got %q, want B", got)
	}
}

func TestPickMainworld_SophontTied_HigherHabitabilityWins(t *testing.T) {
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 5}, Biology: &Biology{HasNativeSophont: true}},
		{Designation: "B", Habitability: &Habitability{Rating: 8}, Biology: &Biology{HasNativeSophont: true}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "B" {
		t.Errorf("got %q, want B (higher habitability)", got)
	}
}

func TestPickMainworld_NoSophont_HighestHabitability(t *testing.T) {
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 4}},
		{Designation: "B", Habitability: &Habitability{Rating: 8}},
		{Designation: "C", Habitability: &Habitability{Rating: 6}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "B" {
		t.Errorf("got %q, want B", got)
	}
}

func TestPickMainworld_HabitabilityTied_HighestResourceWins(t *testing.T) {
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 7}, Biology: &Biology{ResourceRating: 5}},
		{Designation: "B", Habitability: &Habitability{Rating: 7}, Biology: &Biology{ResourceRating: 8}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "B" {
		t.Errorf("got %q, want B (resource tiebreaker)", got)
	}
}

func TestPickMainworld_AllZero_FirstTerrestrialFallback(t *testing.T) {
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 0}, Biology: &Biology{ResourceRating: 0}},
		{Designation: "B", Habitability: &Habitability{Rating: 0}, Biology: &Biology{ResourceRating: 0}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "A" {
		t.Errorf("got %q, want A (first terrestrial fallback)", got)
	}
}

func TestPickMainworld_BeltsAndGGsOnly_EmptyString(t *testing.T) {
	belt := DetailedPlacement{Designation: "Belt", SizeCode: "0"}
	belt.Body = BodyPlanetoidBelt
	gg := DetailedPlacement{Designation: "GG"}
	gg.Body = BodyGasGiant
	detailed := []DetailedPlacement{belt, gg}
	got := pickMainworld(detailed)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestPickMainworld_MoonAsMainworld(t *testing.T) {
	// Planet has habitability 4; moon has habitability 9 → moon wins.
	dp := DetailedPlacement{
		Designation:  "Aab IV",
		Habitability: &Habitability{Rating: 4},
		Moons: []Moon{
			{Designation: "Aab IV d", Habitability: &Habitability{Rating: 9}},
		},
	}
	dp.Body = BodyTerrestrial
	detailed := []DetailedPlacement{dp}
	got := pickMainworld(detailed)
	if got != "Aab IV d" {
		t.Errorf("got %q, want Aab IV d (moon)", got)
	}
}

func TestPickMainworld_EmptyDetailed_EmptyString(t *testing.T) {
	got := pickMainworld([]DetailedPlacement{})
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestPickMainworld_OnlyResourceNoHabitability(t *testing.T) {
	// All Habitability=0; pick by ResourceRating.
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 0}, Biology: &Biology{ResourceRating: 5}},
		{Designation: "B", Habitability: &Habitability{Rating: 0}, Biology: &Biology{ResourceRating: 9}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	if got != "B" {
		t.Errorf("got %q, want B (resource fallback)", got)
	}
}

func TestPickMainworld_ExtinctSophont_Counts(t *testing.T) {
	// Body A: extant sophont; Body B: extinct sophont; Body C: nothing.
	// Either A or B can win the sophont category; tiebreaker is habitability.
	detailed := []DetailedPlacement{
		{Designation: "A", Habitability: &Habitability{Rating: 5}, Biology: &Biology{HasNativeSophont: true}},
		{Designation: "B", Habitability: &Habitability{Rating: 7}, Biology: &Biology{HadExtinctSophont: true}},
		{Designation: "C", Habitability: &Habitability{Rating: 9}, Biology: &Biology{}},
	}
	for i := range detailed {
		detailed[i].Body = BodyTerrestrial
	}
	got := pickMainworld(detailed)
	// A and B both qualify as "sophont"; B has higher habitability → B.
	if got != "B" {
		t.Errorf("got %q, want B (extinct sophont with higher habitability)", got)
	}
}
