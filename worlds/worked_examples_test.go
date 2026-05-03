package worlds_test

import (
	"math"
	"testing"

	"wbh/roller"
	"wbh/stars"
	"wbh/worlds"
)

// composeSol builds a single-star Sol-like system (G2 V) using
// stars.Compose so tests can construct it deterministically (no rolls).
func composeSol() stars.System {
	sol := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.000,
		Diameter:        1.000,
		Temperature:     5772,
		AgeGyr:          4.6,
	})
	return stars.System{
		Primary:            sol,
		PrimaryDesignation: "A",
		AgeGyr:             4.6,
	}
}

// composeZed builds the WBH p. 35/40 Zed quintuple system using
// stars.Compose so tests can construct it deterministically (no rolls).
func composeZed() stars.System {
	aa := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: stars.V, Mass: 0.929, Diameter: 0.967, Temperature: 5440,
	})
	ab := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 8},
		LuminosityClass: stars.V, Mass: 0.907, Diameter: 0.957, Temperature: 5360,
	})
	b := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'K', Subtype: 8},
		LuminosityClass: stars.V, Mass: 0.626, Diameter: 0.777, Temperature: 3980,
	})
	ca := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'M', Subtype: 0},
		LuminosityClass: stars.V, Mass: 0.510, Diameter: 0.728, Temperature: 3700,
	})
	cb := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindWhiteDwarf, Mass: 0.490, Diameter: 0.017, Temperature: 6700,
	})
	return stars.System{
		Primary: aa,
		Companions: []stars.CompanionStar{
			{Star: ab, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.09, Eccentricity: 0.11, ParentIndex: -1},
			{Star: b, OrbitClass: stars.OrbitNear, OrbitNumber: 6.10, Eccentricity: 0.08, ParentIndex: -1},
			{Star: ca, OrbitClass: stars.OrbitFar, OrbitNumber: 12.10, Eccentricity: 0.47, ParentIndex: -1},
			{Star: cb, OrbitClass: stars.OrbitCompanion, OrbitNumber: 0.21, Eccentricity: 0.24, ParentIndex: 2},
		},
	}
}

func TestZed_AvailableOrbits(t *testing.T) {
	t.Parallel()

	// WBH p. 40: Zed quintuple available orbits.
	//   Aab pair (G7 V + G8 V):  MAO 0.61, [[0.61, 5.10], [7.10, 10.10], [14.10, 20.00]], Total 13.39
	//   B (K8 V):                MAO 0.02, [[0.02, 1.10]], Total 1.08
	//   Cab pair (M0 V + D):     MAO 0.74, [[0.74, 7.10]], Total 6.36

	sys := composeZed()

	got, err := worlds.AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	if len(got.Groups) != 3 {
		t.Fatalf("groups = %d, want 3", len(got.Groups))
	}

	// Group Aab.
	aab := got.Groups[0]
	if aab.Designation != "Aab" {
		t.Errorf("groups[0].Designation = %q, want \"Aab\"", aab.Designation)
	}
	if math.Abs(aab.MAO-0.61) > 0.01 {
		t.Errorf("Aab MAO = %v, want 0.61", aab.MAO)
	}
	wantAab := []worlds.Interval{
		{Min: 0.61, Max: 5.10},
		{Min: 7.10, Max: 10.10},
		{Min: 14.10, Max: 20.00},
	}
	if !intervalsEqual(aab.Intervals, wantAab, 0.01) {
		t.Errorf("Aab intervals = %+v, want %+v", aab.Intervals, wantAab)
	}
	if math.Abs(aab.Total()-13.39) > 0.05 {
		t.Errorf("Aab Total = %v, want 13.39", aab.Total())
	}

	// Group B.
	bg := got.Groups[1]
	if bg.Designation != "B" {
		t.Errorf("groups[1].Designation = %q, want \"B\"", bg.Designation)
	}
	if math.Abs(bg.MAO-0.02) > 0.005 {
		t.Errorf("B MAO = %v, want 0.02", bg.MAO)
	}
	wantB := []worlds.Interval{{Min: 0.02, Max: 1.10}}
	if !intervalsEqual(bg.Intervals, wantB, 0.01) {
		t.Errorf("B intervals = %+v, want %+v", bg.Intervals, wantB)
	}
	if math.Abs(bg.Total()-1.08) > 0.01 {
		t.Errorf("B Total = %v, want 1.08", bg.Total())
	}

	// Group Cab.
	cab := got.Groups[2]
	if cab.Designation != "Cab" {
		t.Errorf("groups[2].Designation = %q, want \"Cab\"", cab.Designation)
	}
	if math.Abs(cab.MAO-0.74) > 0.01 {
		t.Errorf("Cab MAO = %v, want 0.74", cab.MAO)
	}
	wantCab := []worlds.Interval{{Min: 0.74, Max: 7.10}}
	if !intervalsEqual(cab.Intervals, wantCab, 0.01) {
		t.Errorf("Cab intervals = %+v, want %+v", cab.Intervals, wantCab)
	}
	if math.Abs(cab.Total()-6.36) > 0.01 {
		t.Errorf("Cab Total = %v, want 6.36", cab.Total())
	}
}

// intervalsEqual reports whether two interval slices match within tol on each bound.
func intervalsEqual(a, b []worlds.Interval, tol float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(a[i].Min-b[i].Min) > tol || math.Abs(a[i].Max-b[i].Max) > tol {
			return false
		}
	}
	return true
}

func TestSol_AvailableOrbits(t *testing.T) {
	t.Parallel()

	sol := stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.000, Diameter: 1.000, Temperature: 5772,
	})
	sys := stars.System{Primary: sol}
	got, err := worlds.AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	if len(got.Groups) != 1 {
		t.Fatalf("groups = %d, want 1", len(got.Groups))
	}
	g := got.Groups[0]
	if g.Designation != "A" {
		t.Errorf("Designation = %q, want \"A\"", g.Designation)
	}
	if math.Abs(g.MAO-0.03) > 0.005 {
		t.Errorf("MAO = %v, want ~0.03", g.MAO)
	}
	if len(g.Intervals) != 1 {
		t.Fatalf("intervals = %d, want 1", len(g.Intervals))
	}
	if math.Abs(g.Intervals[0].Max-20.0) > 1e-9 {
		t.Errorf("Max = %v, want 20.0", g.Intervals[0].Max)
	}

	// Sanity: full Total ≈ 19.97.
	if math.Abs(g.Total()-19.97) > 0.01 {
		t.Errorf("Total = %v, want ~19.97", g.Total())
	}
}

func TestZed_AllocateOrbitsByStar(t *testing.T) {
	t.Parallel()
	sys := composeZed()
	avail, err := worlds.AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("AvailableOrbits: %v", err)
	}
	counts := worlds.Counts{GasGiants: 4, PlanetoidBelts: 2, Terrestrials: 11, Total: 17}
	got, err := worlds.AllocateOrbitsByStar(avail, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(got) != 3 {
		t.Fatalf("allocations = %d, want 3", len(got))
	}
	// Per the book p. 44 narration:
	//   Aab: floor(13.39) = 13 (pair, no +1)
	//   B:   floor(1.08 + 1) = 2 (single, has prior allowable, no companion → +1)
	//   Cab: floor(6.36) = 6 (pair, no +1)
	//   Sum: 21
	//   Aab worlds: ceil(17×13/21) = ceil(10.52) = 11 → ROUND UP for primary
	//   B worlds:   floor(17×2/21) = floor(1.62) = 1 → round down for middle
	//   Cab worlds: 17 - 11 - 1 = 5 (remainder for last)
	wantOrbits := []int{13, 2, 6}
	wantWorlds := []int{11, 1, 5}
	for i := range got {
		if got[i].TotalStarOrbits != wantOrbits[i] {
			t.Errorf("group %d (%s) TotalStarOrbits = %d, want %d", i, got[i].Group.Designation, got[i].TotalStarOrbits, wantOrbits[i])
		}
		if got[i].AllocatedWorlds != wantWorlds[i] {
			t.Errorf("group %d (%s) AllocatedWorlds = %d, want %d", i, got[i].Group.Designation, got[i].AllocatedWorlds, wantWorlds[i])
		}
	}
}

func TestZed_RollBaselineNumber(t *testing.T) {
	t.Parallel()
	sys := composeZed()
	// Zed: companion (Ab) → DM-2; secondaries B + Ca → DM-2; total 17 → no
	// band DM (16-17 unlisted in book → 0). Primary G7 V → no class DM.
	// Net DM = -4. Book rolls 9. Result: 9 - 4 = 5.
	got, err := worlds.RollBaselineNumber(roller.NewScripted(9), sys, worlds.Counts{Total: 17})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got != 5 {
		t.Errorf("baseline = %d, want 5", got)
	}
}

func TestZed_BaselineOrbit(t *testing.T) {
	t.Parallel()
	sys := composeZed()
	avail, err := worlds.AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("%v", err)
	}
	primary := avail.Groups[0]
	// Book: baselineN=5, totalWorlds=17, HZCO Aab = 3.3, roll 5 → variance (5-7)/10 = -0.2.
	// BaselineOrbit = 3.3 + (-0.2) = 3.1.
	got, err := worlds.BaselineOrbit(roller.NewScripted(5), primary, primary.HZCO(), 5, 17)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if math.Abs(got-3.1) > 0.05 {
		t.Errorf("BaselineOrbit = %v, want 3.1", got)
	}
}

func TestZed_Spread(t *testing.T) {
	t.Parallel()
	sys := composeZed()
	avail, err := worlds.AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Zed: primary Aab MAO 0.61, baselineOrbit 3.1, baselineN 5, totalStars 3.
	// (3.1 - 0.61) / 5 = 0.498
	got := worlds.Spread(avail.Groups[0], 11, 3.1, 5, 3)
	if math.Abs(got-0.498) > 0.005 {
		t.Errorf("Spread = %v, want 0.498", got)
	}
}

func TestZed_PlaceOrbitSlots_Aab(t *testing.T) {
	t.Parallel()
	sys := composeZed()
	avail, err := worlds.AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Aab primary placement (WBH pp. 49-50). The book narrates 11 slots with
	// baselineN=5 (from Step 2 roll), baselineOrbit=3.1 and spread≈0.5
	// (exact: 0.498 per Step 5).
	//
	// Book orbit sequence (rounded): 1.0, 1.6, 2.1, 2.7, 3.1, 3.5, 4.1, 4.6,
	// 7.2, 7.8, 8.3. Slot 5 (index 4) is fixed at baselineOrbit=3.1.
	//
	// Key WBH behaviours exercised:
	//   1. Variance rolls shift orbits by ±(2D-7)×spread/10.
	//   2. Exclusion-zone widening: any slot landing in the gap (5.10, 7.10)
	//      is pushed past 7.10, producing an orbit >7.10.
	//   3. 11 slots total are placed.
	//
	// We verify slot count, the baseline slot value (index 4 = slot 5), and
	// that exclusion-zone widening fires. We do not enforce exact per-slot
	// values because the book rounds aggressively.
	allocs := []worlds.StarAllocation{{Group: avail.Groups[0], AllocatedWorlds: 11}}

	// Variance rolls per book narration (10 rolls; slot 5 is baseline, no roll):
	//   slots 1-4 pre-baseline (4 rolls), slots 6-11 post-baseline (6 rolls).
	rolls := []int{5, 9, 7, 9, 7, 7, 7, 7, 7, 7}
	got, err := worlds.PlaceOrbitSlots(roller.NewScripted(rolls...), allocs, 5, 3.1, 0.5, 0)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Must have 11 slots.
	if len(got) != 11 {
		t.Fatalf("slots = %d, want 11", len(got))
	}

	// baselineN=5, so got[4] (slot 5) must be exactly baselineOrbit=3.1.
	if math.Abs(got[4].Orbit-3.1) > 0.01 {
		t.Errorf("baseline slot (index 4) Orbit = %v, want 3.1", got[4].Orbit)
	}

	// After the exclusion zone (5.10, 7.10) at least one slot must be >7.10.
	foundWidened := false
	for _, s := range got {
		if s.Orbit > 7.10 {
			foundWidened = true
		}
	}
	if !foundWidened {
		orbits := make([]float64, len(got))
		for i, s := range got {
			orbits[i] = s.Orbit
		}
		t.Errorf("expected at least one slot >7.10 (exclusion-zone widened), got %v", orbits)
	}
}

func TestZed_GenerateCounts(t *testing.T) {
	t.Parallel()
	// WBH p. 38 Zed walkthrough — encoded against the Existence/Quantity DM split:
	//
	//   GG existence DMs (Special Circumstances only): per-post-stellar (Cb=1) = -1; 4+ stars = -1. Total -2.
	//     raw 9 + (-2) = 7 ≤9 → present.
	//   GG quantity DMs (full table): existence DMs -2 + single-Class-V (no, multi-star) = -2.
	//     raw 11 + (-2) = 9 → row 9-11 → 4 GGs. ✓
	//   Belts existence DMs (Special Circumstances only): per-post-stellar (Cb=1) = +1. Total +1.
	//     raw 7 + 1 = 8 → present.
	//   Belts quantity DMs (full table): existence DMs +1 + GGs present +1 + 2+ stars +1 = +3.
	//     raw 7 + 3 = 10 → row 7-11 → 2 belts. ✓
	//   Terrestrials DMs: -1 per post-stellar (Cb=1).
	//     raw 12 → 12 - 2 - 1 = 9 (≥3); + D3-1 (D3=3 → +2) → 11. ✓
	//   Total = 4 + 2 + 11 = 17.
	sys := composeZed()
	r := roller.NewScripted(
		9, 11, // GG existence + quantity
		7, 7, // belts existence + quantity
		12, 3, // terrestrials 2D + D3
	)
	got, err := worlds.GenerateCounts(r, sys, worlds.CountsOpts{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got.GasGiants != 4 {
		t.Errorf("GasGiants = %d, want 4", got.GasGiants)
	}
	if got.PlanetoidBelts != 2 {
		t.Errorf("PlanetoidBelts = %d, want 2", got.PlanetoidBelts)
	}
	if got.Terrestrials != 11 {
		t.Errorf("Terrestrials = %d, want 11", got.Terrestrials)
	}
	if got.Total != 17 {
		t.Errorf("Total = %d, want 17", got.Total)
	}
}

func TestZed_AddAnomalous(t *testing.T) {
	t.Parallel()
	sys := composeZed()
	avail, err := worlds.AvailableOrbits(sys)
	if err != nil {
		t.Fatalf("%v", err)
	}
	counts := worlds.Counts{GasGiants: 4, PlanetoidBelts: 2, Terrestrials: 11, Total: 17}
	allocs, err := worlds.AllocateOrbitsByStar(avail, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// After Step 4 empty distribution, B's allocation grows from 1 to 2.
	// Simulate that here for the isolated test.
	allocs[1].AllocatedWorlds = 2
	allocs[2].AllocatedWorlds = 5

	var slots []worlds.Slot
	aab := allocs[0].Group
	for _, o := range []float64{1.0, 1.6, 2.1, 2.7, 3.1, 3.5, 4.1, 4.6, 7.2, 7.8, 8.3} {
		slots = append(slots, worlds.Slot{Group: aab, Orbit: o})
	}
	// Book: anomalous=10 (1), type=10 (Retrograde), parent group D3=1 (Aab),
	// orbit raw 2D-2=5, d10=2 → 5.2.
	rolls := []int{10, 10, 1, 7, 2}
	out, newCounts, err := worlds.AddAnomalous(roller.NewScripted(rolls...), slots, allocs, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}
	last := out[len(out)-1]
	if last.Anomaly != worlds.AnomalyRetrograde {
		t.Errorf("Anomaly = %v, want Retrograde", last.Anomaly)
	}
	if math.Abs(last.Orbit-5.2) > 0.01 {
		t.Errorf("Orbit = %v, want 5.2", last.Orbit)
	}
	if last.Group.Designation != "Aab" {
		t.Errorf("Group = %v, want Aab", last.Group.Designation)
	}
	if newCounts.Terrestrials != 12 || newCounts.Total != 18 {
		t.Errorf("counts = %+v, want T=12 Total=18", newCounts)
	}
}

func TestZed_FullPlacement(t *testing.T) {
	t.Parallel()
	sys := composeZed()
	rolls := []int{
		// GenerateCounts (6 rolls):
		//   GG existence raw 9 (DMs -2 → 7 ≤9, present)
		//   GG quantity raw 11 (DMs -2 → 9 → 4 GGs)
		//   Belts existence raw 7 (DMs +1 → 8 ≥8, present)
		//   Belts quantity raw 7 (DMs +3 → 10 → 2 belts)
		//   Terrestrials 2D=12 → 12-2-1=9 (≥3) + D3-1 (D3=3 → +2) = 11
		9, 11, 7, 7, 12, 3,
		// RollBaselineNumber: raw 9 + DMs -4 = 5
		9,
		// BaselineOrbit (3a, HZCO≈3.32): 2D=5 → variance -0.2 → ≈3.12
		5,
		// RollEmptyOrbits: 10 → 1
		10,
		// PlaceOrbitSlots:
		//   Aab: 11 slots, slot 5 (index 4) is baseline-fixed; 10 variance rolls.
		5, 9, 7, 9, 7, 7, 7, 7, 7, 7,
		//   B: 2 slots (1 regular + 1 extra from emptyOrbits distributed to Near); 2 variance rolls.
		7, 7,
		//   Cab: 5 slots; 5 variance rolls.
		10, 5, 7, 9, 5,
		// AddAnomalous: count 2D=10 (→1), type 2D=10 (Retrograde),
		//   parent D3=1 (→idx 0, Aab), orbit 2D-2=5, d10=2 → orbit 5.2
		10, 10, 1, 7, 2,
		// PlaceWorlds: n=19 slots, prefixMax=ceil(19/6)=4.
		//   Order: empty → GG → belt → terrestrial.
		//   Each body: prefix roll (1D, keep ≤4) + right roll (1D, reject if idx≥19).
		//   All rolls below use prefix 1-3 (always valid) or prefix 4 right=1 (idx=18).
		// Empty (1):
		1, 1, // idx 0
		// GG (4):
		1, 2, // idx 1
		1, 3, // idx 2
		1, 4, // idx 3
		1, 5, // idx 4
		// Belt (2):
		1, 6, // idx 5
		2, 1, // idx 6
		// Terrestrial (12):
		2, 2, 2, 3, 2, 4, 2, 5, 2, 6,
		3, 1, 3, 2, 3, 3, 3, 4, 3, 5,
		3, 6, 4, 1, // idx 17, 18
		// RollPlanetEccentricities: 16 non-empty non-belt bodies (4 GG + 12 terr) × 2 rolls each.
		//   2D=7 → row 7 (or 9 for retrograde with DM+2); second roll=1.
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
	}
	got, err := worlds.GenerateSystemPlacement(roller.NewScripted(rolls...), sys)
	if err != nil {
		t.Fatalf("GenerateSystemPlacement: %v", err)
	}

	// High-level count assertions.
	if got.Counts.GasGiants != 4 {
		t.Errorf("GasGiants = %d, want 4", got.Counts.GasGiants)
	}
	if got.Counts.PlanetoidBelts != 2 {
		t.Errorf("PlanetoidBelts = %d, want 2", got.Counts.PlanetoidBelts)
	}
	if got.Counts.Terrestrials != 12 {
		t.Errorf("Terrestrials = %d, want 12 (11 from GenerateCounts + 1 from anomalous)", got.Counts.Terrestrials)
	}
	if got.Counts.Total != 18 {
		t.Errorf("Total = %d, want 18", got.Counts.Total)
	}
	if got.BaselineN != 5 {
		t.Errorf("BaselineN = %d, want 5", got.BaselineN)
	}
	if math.Abs(got.BaselineOrbit-3.1) > 0.05 {
		t.Errorf("BaselineOrbit = %v, want ~3.1", got.BaselineOrbit)
	}
	if got.EmptyOrbits != 1 {
		t.Errorf("EmptyOrbits = %d, want 1", got.EmptyOrbits)
	}
	if math.Abs(got.SystemSpread-0.50) > 0.005 {
		t.Errorf("SystemSpread = %v, want ~0.50", got.SystemSpread)
	}
	if len(got.Placements) != 19 {
		t.Fatalf("Placements = %d, want 19", len(got.Placements))
	}

	// Body-type counts.
	bodyCounts := map[worlds.BodyType]int{}
	for _, p := range got.Placements {
		bodyCounts[p.Body]++
	}
	if bodyCounts[worlds.BodyEmpty] != 1 {
		t.Errorf("Empty bodies = %d, want 1", bodyCounts[worlds.BodyEmpty])
	}
	if bodyCounts[worlds.BodyGasGiant] != 4 {
		t.Errorf("GasGiant bodies = %d, want 4", bodyCounts[worlds.BodyGasGiant])
	}
	if bodyCounts[worlds.BodyPlanetoidBelt] != 2 {
		t.Errorf("Belt bodies = %d, want 2", bodyCounts[worlds.BodyPlanetoidBelt])
	}
	if bodyCounts[worlds.BodyTerrestrial] != 12 {
		t.Errorf("Terrestrial bodies = %d, want 12", bodyCounts[worlds.BodyTerrestrial])
	}

	// The retrograde anomaly must land in group Aab at orbit ≈5.2.
	var retro *worlds.Placement
	for i := range got.Placements {
		if got.Placements[i].Anomaly == worlds.AnomalyRetrograde {
			retro = &got.Placements[i]
			break
		}
	}
	if retro == nil {
		t.Fatalf("no retrograde placement found")
	}
	if retro.Group.Designation != "Aab" {
		t.Errorf("retrograde group = %s, want Aab", retro.Group.Designation)
	}
	if math.Abs(retro.Orbit-5.2) > 0.05 {
		t.Errorf("retrograde orbit = %v, want 5.2", retro.Orbit)
	}
}

// TestSol_GenerateSystemPlacement is a single-star smoke test: assert
// the GenerateSystemPlacement pipeline runs without error on a
// single-G2-V system, produces exactly one StarAllocation, and yields
// a non-empty Placements slice. This complements TestZed_FullPlacement
// (multi-star) to cover the single-star path that 2B left untested.
//
// No book-narrated dice trail is required; this is a smoke test, not
// a worked-example regression.
func TestSol_GenerateSystemPlacement(t *testing.T) {
	t.Parallel()

	sys := composeSol()

	// Use a seeded roller; the specific values don't matter for a smoke
	// test, only that the pipeline completes without error.
	r := roller.NewSeeded(42)

	sp, err := worlds.GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatalf("GenerateSystemPlacement returned error: %v", err)
	}

	if len(sp.Allocations) != 1 {
		t.Errorf("len(Allocations) = %d, want 1 (single-star system)", len(sp.Allocations))
	}
	if sp.Allocations[0].Group.Designation != "A" {
		t.Errorf("Allocations[0].Group.Designation = %q, want \"A\"", sp.Allocations[0].Group.Designation)
	}
	if len(sp.Placements) == 0 {
		t.Error("Placements is empty, want at least one body")
	}
	if sp.Counts.Total <= 0 {
		t.Errorf("Counts.Total = %d, want > 0", sp.Counts.Total)
	}
}
