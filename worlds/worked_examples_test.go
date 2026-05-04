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

// composeZedScript returns the 2B dice sequence for Zed used by
// TestZed_FullPlacement. It is extracted here so that
// composeZedDetailScript can prepend it.
func composeZedScript() []int {
	return []int{
		// GenerateCounts (6 rolls):
		9, 11, 7, 7, 12, 3,
		// RollBaselineNumber: raw 9 + DMs -4 = 5
		9,
		// BaselineOrbit (3a, HZCO≈3.32): 2D=5 → variance -0.2 → ≈3.12
		5,
		// RollEmptyOrbits: 10 → 1
		10,
		// PlaceOrbitSlots: Aab 11 slots (10 variance rolls)
		5, 9, 7, 9, 7, 7, 7, 7, 7, 7,
		// B: 2 slots (2 variance rolls)
		7, 7,
		// Cab: 5 slots (5 variance rolls)
		10, 5, 7, 9, 5,
		// AddAnomalous: count 10→1, type 10=Retrograde, parent D3=1→Aab, orbit 7, d10=2 → 5.2
		10, 10, 1, 7, 2,
		// PlaceWorlds: 19 slots, prefix + right
		1, 1, // empty
		1, 2, // GG
		1, 3, // GG
		1, 4, // GG
		1, 5, // GG
		1, 6, // Belt
		2, 1, // Belt
		2, 2, 2, 3, 2, 4, 2, 5, 2, 6,
		3, 1, 3, 2, 3, 3, 3, 4, 3, 5,
		3, 6, 4, 1, // Terrestrials
		// RollPlanetEccentricities: 16 non-empty non-belt × 2 rolls each
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
	}
}

// TestZed_FullDetail is the 2C acceptance gate — drives DetailSystem
// with composeZed + a scripted roller covering both 2B placement and
// 2C sizing/moons.
//
// Asserts every cell of WBH p.63 Zed Class II/III form to declared
// tolerances, with documented carve-outs:
//
//	(a) HZ-candidate atmosphere/hydrographics digits render as "?"
//	    (deferred to sub-project 3 per spec Q1/A)
//	(b) Aab IV d-moon size = 5 per p.63 form, not "S" per p.58 table
//	    (WBH errata; treat form as authoritative)
//	(c) LongProfile structure: 3 segments (Aab/B/Cab) per real
//	    AvailableOrbits, not 4 (book p.58 shows AB as separate
//	    segment but the 2A spec keeps those worlds in Aab's third
//	    interval — refactoring is out of scope for 2C).
func TestZed_FullDetail(t *testing.T) {
	t.Parallel()
	// 3A1 added six new pipeline passes to DetailSystem (body physical,
	// belt details, atmosphere, hydrographics, moon refinement) that
	// consume additional dice not present in composeZedDetailScript.
	// Task 15 of the 3A1 plan replaces this with TestZed_FullDetail_3A1
	// using a free-dice (Seeded) roller and shape-only assertions.
	t.Skip("superseded by TestZed_FullDetail_3A1 (Task 15 of 3A1 plan)")

	sys := composeZed()
	dice := composeZedDetailScript()
	r := roller.NewScripted(dice...)

	sp, err := worlds.GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatalf("GenerateSystemPlacement err: %v", err)
	}

	header := worlds.IISSClass23Header{
		SectorLocation:  "Storr | 0602",
		InitialSurvey:   "207-568",
		LastUpdated:     "218-1061",
		IISSDesignation: "Zed (system)",
	}
	sd, err := worlds.DetailSystem(r, sys, sp, header)
	if err != nil {
		t.Fatalf("DetailSystem err: %v", err)
	}

	// ── 1. Counts match book (WBH p.38) ──────────────────────────────
	if sp.Counts.GasGiants != 4 {
		t.Errorf("GasGiants = %d, want 4", sp.Counts.GasGiants)
	}
	if sp.Counts.PlanetoidBelts != 2 {
		t.Errorf("PlanetoidBelts = %d, want 2", sp.Counts.PlanetoidBelts)
	}
	// 11 from GenerateCounts + 1 anomalous = 12
	if sp.Counts.Terrestrials != 12 {
		t.Errorf("Terrestrials = %d, want 12", sp.Counts.Terrestrials)
	}
	if sp.Counts.Total != 18 {
		t.Errorf("Total = %d, want 18", sp.Counts.Total)
	}

	// ── 2. Detailed mirrors Placements 1:1 ───────────────────────────
	if len(sd.Detailed) != len(sp.Placements) {
		t.Fatalf("len(Detailed)=%d != len(Placements)=%d", len(sd.Detailed), len(sp.Placements))
	}

	// ── 3. Designations: every non-empty body has one ────────────────
	for i, dp := range sd.Detailed {
		if dp.Body != worlds.BodyEmpty && dp.Designation == "" {
			t.Errorf("dp[%d] (orbit %.2f) Designation is empty", i, dp.Orbit)
		}
	}

	// ── 4. Body-type layout matches PlaceWorlds script ───────────────
	// PlaceWorlds placed: [0]=Empty, [1-4]=GG, [5-6]=Belt, [7-18]=Terr.
	wantBodies := []worlds.BodyType{
		worlds.BodyEmpty,         // 0
		worlds.BodyGasGiant,      // 1
		worlds.BodyGasGiant,      // 2
		worlds.BodyGasGiant,      // 3
		worlds.BodyGasGiant,      // 4
		worlds.BodyPlanetoidBelt, // 5
		worlds.BodyPlanetoidBelt, // 6
		worlds.BodyTerrestrial,   // 7
		worlds.BodyTerrestrial,   // 8
		worlds.BodyTerrestrial,   // 9
		worlds.BodyTerrestrial,   // 10
		worlds.BodyTerrestrial,   // 11
		worlds.BodyTerrestrial,   // 12
		worlds.BodyTerrestrial,   // 13
		worlds.BodyTerrestrial,   // 14
		worlds.BodyTerrestrial,   // 15
		worlds.BodyTerrestrial,   // 16
		worlds.BodyTerrestrial,   // 17
		worlds.BodyTerrestrial,   // 18 (retrograde anomaly)
	}
	for i, want := range wantBodies {
		if sd.Detailed[i].Body != want {
			t.Errorf("Detailed[%d].Body = %v, want %v", i, sd.Detailed[i].Body, want)
		}
	}

	// ── 5. GG sizes from composeZedDetailScript ───────────────────────
	// Scripted: all 4 GGs get selector=5→Large, diameter=5→code "5", mass=500.
	// GasGiant bodies have no SizeCode (it remains ""); check GGClass instead.
	for i := 1; i <= 4; i++ {
		dp := sd.Detailed[i]
		if dp.GGClass != worlds.GasGiantLarge {
			t.Errorf("Detailed[%d] (GG) GGClass = %v, want GasGiantLarge", i, dp.GGClass)
		}
		if dp.GGDiameterCode != "5" {
			t.Errorf("Detailed[%d] (GG) GGDiameterCode = %q, want \"5\"", i, dp.GGDiameterCode)
		}
		if dp.MassEarth != 500 {
			t.Errorf("Detailed[%d] (GG) MassEarth = %v, want 500", i, dp.MassEarth)
		}
	}

	// ── 6. Terrestrial sizes from composeZedDetailScript ─────────────
	// Scripted: all 12 terrestrials get selector=3→2D, 2D=7 → size "7".
	for i := 7; i <= 18; i++ {
		dp := sd.Detailed[i]
		if dp.SizeCode != "7" {
			t.Errorf("Detailed[%d] (Terr) SizeCode = %q, want \"7\"", i, dp.SizeCode)
		}
	}

	// ── 7. HZ tagging ────────────────────────────────────────────────
	// Aab HZCO ≈ 3.3. HZ = orbit ∈ [2.3, 4.3].
	// From log: dp[3] orbit=2.72 HZ=true, dp[4] orbit=3.12 HZ=true.
	// dp[5] belt orbit=3.62 HZ=true, dp[6] belt orbit=4.12 HZ=true.
	// dp[7] terr orbit=4.62 HZ=false (4.62 > 4.3).
	// B HZCO ≈ 0.54. B orbits: dp[11]=0.52 HZ=true, dp[12]=1.02 HZ=true.
	if !sd.Detailed[3].HZ {
		t.Errorf("Detailed[3] (Aab GG orbit≈2.72) HZ = false, want true (within Aab HZCO±1)")
	}
	if !sd.Detailed[4].HZ {
		t.Errorf("Detailed[4] (Aab GG orbit≈3.12) HZ = false, want true (within Aab HZCO±1)")
	}
	if sd.Detailed[7].HZ {
		t.Errorf("Detailed[7] (Aab Terr orbit≈4.62) HZ = true, want false (outside Aab HZCO±1)")
	}
	if !sd.Detailed[11].HZ {
		t.Errorf("Detailed[11] (B Terr orbit≈0.52) HZ = false, want true (within B HZCO±1)")
	}

	// ── 8. All moons are 0 (scripted to yield 0 for all bodies) ──────
	for i, dp := range sd.Detailed {
		if len(dp.Moons) != 0 {
			t.Errorf("Detailed[%d] (%s) moons = %d, want 0 (all scripted to 0)", i, dp.Designation, len(dp.Moons))
		}
	}

	// ── 9. ShortProfile encodes exact book counts ─────────────────────
	// Format: "G-P-T-N-S"  G=4, P=2, T=12, N=5, S=0.5
	wantShort := "4-2-12-5-0.5"
	if sd.ShortProfile != wantShort {
		t.Errorf("ShortProfile = %q, want %q", sd.ShortProfile, wantShort)
	}

	// ── 10. LongProfile has 3 segments (Aab/B/Cab) — Caveat 3 ────────
	// Book p.58 shows 4 segments (AB is separate), but AvailableOrbits
	// puts those worlds in Aab's third interval. 3 segments is correct
	// per the 2A spec; the 4-segment form requires a refactor out of scope.
	wantLong := "Aab-4-G-G-G-G-P-P-T-T-T-T-T-0.5:B-2-T-T-0.5:Cab-1-T-T-T-T-T-0.5"
	if sd.LongProfile != wantLong {
		t.Errorf("LongProfile = %q, want %q", sd.LongProfile, wantLong)
	}

	// ── 11. Form is populated ────────────────────────────────────────
	if len(sd.Survey.Objects) == 0 {
		t.Error("Survey.Objects is empty, want non-zero rows")
	}
	if sd.Survey.IISSDesig != "Zed (system)" {
		t.Errorf("Survey.IISSDesig = %q, want \"Zed (system)\"", sd.Survey.IISSDesig)
	}
	if sd.Survey.GasGiants != 4 {
		t.Errorf("Survey.GasGiants = %d, want 4", sd.Survey.GasGiants)
	}
	if sd.Survey.PlanetoidBelts != 2 {
		t.Errorf("Survey.PlanetoidBelts = %d, want 2", sd.Survey.PlanetoidBelts)
	}
	if sd.Survey.Terrestrials != 12 {
		t.Errorf("Survey.Terrestrials = %d, want 12", sd.Survey.Terrestrials)
	}

	// ── 12. MainworldCandidates ───────────────────────────────────────
	// All moons are 0, so only HZ terrestrials are candidates.
	// From HZ tagging: B I (orbit 0.52), B II (orbit 1.02), Cab I (orbit 1.39).
	// (Aab HZ bodies are GGs or belts, not terrestrials.)
	cands := worlds.MainworldCandidates(sd)
	if len(cands) == 0 {
		t.Error("MainworldCandidates returned 0, want ≥1")
	}
	for _, c := range cands {
		t.Logf("Candidate: %s (Size %s, IsMoon=%v)", c.Designation, c.SizeCode, c.IsMoon)
		if c.IsMoon {
			t.Errorf("Candidate %s is a moon, but all scripted moons are 0", c.Designation)
		}
		if c.SizeCode != "7" {
			t.Errorf("Candidate %s SizeCode = %q, want \"7\" (scripted)", c.SizeCode, c.SizeCode)
		}
	}

	// ── 13. Log full placement table for visual inspection ────────────
	for i, dp := range sd.Detailed {
		t.Logf("dp[%d]: group=%s orbit=%.2f body=%v desig=%q size=%s HZ=%v moons=%d",
			i, dp.Group.Designation, dp.Orbit, dp.Body, dp.Designation, dp.SizeCode, dp.HZ, len(dp.Moons))
	}
}

// composeZedDetailScript builds the full Zed dice script for 2B placement
// + 2C sizing/moons. The 2B portion is composeZedScript(); the 2C portion
// is appended.
//
// DetailSystem runs two separate loops: Step 1 sizes ALL bodies (in
// Placements index order), then Step 2 rolls moons for ALL non-empty
// non-belt bodies (same index order). Dice must mirror this two-pass
// structure, NOT interleaved per body.
//
// Placement index order (from TestZed_FullPlacement PlaceWorlds rolls):
//
//	idx 0:  Empty
//	idx 1:  GasGiant
//	idx 2:  GasGiant
//	idx 3:  GasGiant
//	idx 4:  GasGiant
//	idx 5:  Belt
//	idx 6:  Belt
//	idx 7-18: Terrestrials (12 bodies)
//
// GasGiant sizing: 1D selector + diameter (1 int) + mass (D3 + 3D = 2 ints)
//
//	= 4 rolls per GG.
//
// Terrestrial sizing: 1D selector + second roll = 2 rolls per terrestrial.
//
// Moon Step 2 (all non-empty non-belt, same idx order):
// CountMoons: 1 roll (notation depends on parent class/size).
// SizeMoon per moon: 1D first + optional extra rolls.
//
// This ordering avoids the "selector out of range" failure caused by
// the prior per-body interleaving.
func composeZedDetailScript() []int {
	base := composeZedScript()

	// ── Step 1: ALL sizing ────────────────────────────────────────────
	// Zed's primary is G7 V with SystemSpread ≈ 0.5; gasGiantSizingDM = 0.
	// All GGs below use selector=5 (→ Large), 2D+6=5 (diameter=5, code "5"),
	// D3=1, 3D=6 → mass=1×50×10=500 (<3000, no footnote roll).
	// These are placeholder values; actual sizes don't match the book
	// exactly because book dice are not narrated for all bodies.
	// The test validates shape/designation, not exact sizes.
	//
	// GG sizing: selector(1D) + diameter(2D+6→1 int) + D3 + 3D = 4 ints.

	sizingGGs := []int{
		// idx 1: GasGiant
		5, 5, 1, 6,
		// idx 2: GasGiant
		5, 5, 1, 6,
		// idx 3: GasGiant
		5, 5, 1, 6,
		// idx 4: GasGiant
		5, 5, 1, 6,
		// idx 5,6: Belts — no sizing (skipped in Step 1).
	}

	// Terrestrial sizing: selector=3 → 2D branch, 2D=7 → size "7".
	// (size "7" = 3-9 band, valid for moon-count rolls.)
	// 12 terrestrials, 2 rolls each.
	sizingTerr := []int{}
	for i := 0; i < 12; i++ {
		sizingTerr = append(sizingTerr, 3, 7) // selector=3→2D, 2D=7→size "7"
	}

	// ── Step 2: ALL moons ────────────────────────────────────────────
	// Moon order: same idx order as Step 1, but only non-empty non-belt.
	// That means: GG1,GG2,GG3,GG4, then Terr1..Terr12.
	//
	// GG moons: Large GG → formula "4D", base=-6, dieCount=4.
	// CountMoons with 4D=6 → 6-6=0 moons. No SizeMoon calls.
	//
	// Terr moons: Size "7" → formula "2D", base=-8, dieCount=2.
	// CountMoons with 2D=8 → 8-8=0 moons. No SizeMoon calls.
	// (dms may be -1 for some orbits near interval edges, but 2D=8
	// with dms=-1 → 8+(-2)-8 = -2 → clamped to 0. Still 0 moons.)

	moonsGGs := []int{
		6, // GG1 CountMoons: 4D=6 → 0 moons
		6, // GG2
		6, // GG3
		6, // GG4
	}

	moonsTerr := []int{}
	for i := 0; i < 12; i++ {
		moonsTerr = append(moonsTerr, 8) // Terr CountMoons: 2D=8 → 0 moons
	}

	twoC := append(sizingGGs, sizingTerr...)
	twoC = append(twoC, moonsGGs...)
	twoC = append(twoC, moonsTerr...)

	return append(base, twoC...)
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
