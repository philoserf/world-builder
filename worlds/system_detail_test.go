package worlds

import (
	"math"
	"testing"

	"wbh/roller"
	"wbh/stars"
)

// TestDetailedPlacement_EmbedsPlacement verifies that DetailedPlacement
// embeds Placement so all 2B fields are accessible on the embedded
// type. This is the load-bearing assumption of every subsequent 2C
// task — if the embedding chain breaks, downstream tasks won't compile.
func TestDetailedPlacement_EmbedsPlacement(t *testing.T) {
	t.Parallel()

	dp := DetailedPlacement{
		Placement: Placement{
			Body: BodyTerrestrial,
			AnomalousSlot: AnomalousSlot{
				Slot: Slot{
					StarSlot: "A1",
					Orbit:    1.0,
				},
			},
		},
		SizeCode:    "5",
		DiameterKm:  8000,
		Designation: "A I",
		Period:      Period{Years: 1.0, Days: 365.25},
		HZ:          true,
	}

	if dp.Body != BodyTerrestrial {
		t.Errorf("Body via embedding = %v, want BodyTerrestrial", dp.Body)
	}
	if dp.StarSlot != "A1" {
		t.Errorf("StarSlot via double embedding = %q, want \"A1\"", dp.StarSlot)
	}
	if dp.Orbit != 1.0 {
		t.Errorf("Orbit via double embedding = %v, want 1.0", dp.Orbit)
	}
	if dp.SizeCode != "5" {
		t.Errorf("SizeCode = %q, want \"5\"", dp.SizeCode)
	}
}

// TestSystemDetail_EmbedsSystemPlacement verifies the SystemDetail
// embedding chain: 2B fields accessible via the embedded SystemPlacement.
func TestSystemDetail_EmbedsSystemPlacement(t *testing.T) {
	t.Parallel()

	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			Counts:        Counts{GasGiants: 4, PlanetoidBelts: 2, Terrestrials: 12, Total: 18},
			BaselineN:     5,
			BaselineOrbit: 3.1,
		},
		Detailed:     []DetailedPlacement{},
		ShortProfile: "4-2-12-5-0.5",
	}

	if sd.Counts.GasGiants != 4 {
		t.Errorf("Counts.GasGiants via embedding = %d, want 4", sd.Counts.GasGiants)
	}
	if sd.BaselineN != 5 {
		t.Errorf("BaselineN via embedding = %d, want 5", sd.BaselineN)
	}
	if sd.ShortProfile != "4-2-12-5-0.5" {
		t.Errorf("ShortProfile = %q, want \"4-2-12-5-0.5\"", sd.ShortProfile)
	}
}

// TestDetailSystem_PipelineComposition is a smoke test that asserts
// DetailSystem runs end-to-end without error on a single-G2-V system
// and that each pipeline output is non-empty.
func TestDetailSystem_PipelineComposition(t *testing.T) {
	t.Parallel()

	sys := stars.System{
		Primary: stars.Compose(stars.ComposeOpts{
			Kind:            stars.KindMainSequence,
			SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
			LuminosityClass: stars.V,
			Mass:            1.0,
			Diameter:        1.0,
			Temperature:     5800,
			AgeGyr:          4.6,
		}),
		PrimaryDesignation: "A",
		AgeGyr:             4.6,
	}

	r := roller.NewSeeded(1)
	sp, err := GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatalf("GenerateSystemPlacement err: %v", err)
	}

	header := IISSClass23Header{
		SectorLocation:  "Sol Sector | 0801",
		IISSDesignation: "Sol (system)",
	}
	sd, err := DetailSystem(r, sys, sp, header)
	if err != nil {
		t.Fatalf("DetailSystem err: %v", err)
	}

	if len(sd.Detailed) != len(sp.Placements) {
		t.Errorf("len(Detailed) = %d, len(sp.Placements) = %d, want equal",
			len(sd.Detailed), len(sp.Placements))
	}
	if sd.ShortProfile == "" {
		t.Error("ShortProfile is empty, want non-empty")
	}
	if sd.LongProfile == "" {
		t.Error("LongProfile is empty, want non-empty")
	}
	if sd.Survey.IISSDesig != "Sol (system)" {
		t.Errorf("Survey.IISSDesig = %q, want \"Sol (system)\"", sd.Survey.IISSDesig)
	}

	for i, dp := range sd.Detailed {
		if dp.Body != BodyEmpty && dp.Designation == "" {
			t.Errorf("dp[%d] (body %v, orbit %v) Designation is empty", i, dp.Body, dp.Orbit)
		}
	}
}

// TestDetailSystem_TerrestrialMassPersisted asserts that after DetailSystem,
// every terrestrial body with a populated Physical block also has a non-zero
// MassEarth equal to DeriveMass(density, diameter).
//
// Regression: dp.MassEarth was set only by gas-giant sizing; terrestrial paths
// computed mass inside GenerateBodyPhysical and discarded it. Renderers that
// read dp.MassEarth directly (Class IV-P Size table) emitted 0; silent-zero
// callers (surface_tidal_effects.go) short-circuited.
func TestDetailSystem_TerrestrialMassPersisted(t *testing.T) {
	t.Parallel()

	sys := stars.System{
		Primary: stars.Compose(stars.ComposeOpts{
			Kind:            stars.KindMainSequence,
			SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
			LuminosityClass: stars.V,
			Mass:            1.0,
			Diameter:        1.0,
			Temperature:     5800,
			AgeGyr:          4.6,
		}),
		PrimaryDesignation: "A",
		AgeGyr:             4.6,
	}

	r := roller.NewSeeded(42)
	sp, err := GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatalf("GenerateSystemPlacement err: %v", err)
	}
	sd, err := DetailSystem(r, sys, sp, IISSClass23Header{})
	if err != nil {
		t.Fatalf("DetailSystem err: %v", err)
	}

	saw := 0
	for i, dp := range sd.Detailed {
		if dp.Physical == nil {
			continue
		}
		saw++
		want := DeriveMass(dp.Physical.Density, dp.DiameterKm)
		if dp.MassEarth != want {
			t.Errorf("dp[%d] %s: MassEarth = %v, want %v (DeriveMass(%v, %v))",
				i, dp.Designation, dp.MassEarth, want, dp.Physical.Density, dp.DiameterKm)
		}
	}
	if saw == 0 {
		t.Fatal("no terrestrials with Physical in this seed; test cannot validate")
	}
}

// TestDetailSystem_MoonBodyPhysicalPersisted asserts that after DetailSystem,
// every non-trivial terrestrial moon (not GG-cascade, not Size 0/R/blank) has
// Physical populated and MassEarth = DeriveMass(density, diameter).
//
// Regression: runStep5A (3A1) iterated detailed[i] only — never dp.Moons[j].
// Moons therefore had Physical=nil, MassEarth=0, and downstream callers
// silently no-opped on nil-Physical or zero-mass guards.
func TestDetailSystem_MoonBodyPhysicalPersisted(t *testing.T) {
	t.Parallel()

	sys := stars.System{
		Primary: stars.Compose(stars.ComposeOpts{
			Kind:            stars.KindMainSequence,
			SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
			LuminosityClass: stars.V,
			Mass:            1.0,
			Diameter:        1.0,
			Temperature:     5800,
			AgeGyr:          4.6,
		}),
		PrimaryDesignation: "A",
		AgeGyr:             4.6,
	}

	r := roller.NewSeeded(42)
	sp, err := GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatalf("GenerateSystemPlacement err: %v", err)
	}
	sd, err := DetailSystem(r, sys, sp, IISSClass23Header{})
	if err != nil {
		t.Fatalf("DetailSystem err: %v", err)
	}

	saw := 0
	for _, dp := range sd.Detailed {
		for j := range dp.Moons {
			m := &dp.Moons[j]
			if m.GGClass != NotGasGiant {
				continue
			}
			if m.SizeCode == "" || m.SizeCode == "0" || m.SizeCode == "R" {
				continue
			}
			saw++
			if m.Physical == nil {
				t.Errorf("moon %s (parent %s, size %s): Physical = nil, want non-nil",
					m.Designation, dp.Designation, m.SizeCode)
				continue
			}
			want := DeriveMass(m.Physical.Density, m.DiameterKm)
			if m.MassEarth != want {
				t.Errorf("moon %s: MassEarth = %v, want %v (DeriveMass(%v, %v))",
					m.Designation, m.MassEarth, want, m.Physical.Density, m.DiameterKm)
			}
		}
	}
	if saw == 0 {
		t.Fatal("no non-trivial moons in this system; test cannot validate (try a different seed)")
	}
}

// TestDetailSystem_MoonAtmosphereHydrographicsPersisted asserts that after
// DetailSystem, every non-trivial terrestrial moon of an HZ-orbit planet has
// Atmosphere and Hydrographics populated.
//
// Spec: docs/specs/2026-05-03-world-physical-3a1-design.md requires atm + hydro
// for "every HZ-orbit body and every HZ-planet moon". The companion fix in
// `46cc66e` closed body-physical for moons; atm/hydro is the remaining 3A1 gap.
func TestDetailSystem_MoonAtmosphereHydrographicsPersisted(t *testing.T) {
	t.Parallel()

	sys := stars.System{
		Primary: stars.Compose(stars.ComposeOpts{
			Kind:            stars.KindMainSequence,
			SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
			LuminosityClass: stars.V,
			Mass:            1.0,
			Diameter:        1.0,
			Temperature:     5800,
			AgeGyr:          4.6,
		}),
		PrimaryDesignation: "A",
		AgeGyr:             4.6,
	}

	saw := 0
	for iter := range 30 {
		seed := int64(iter)
		r := roller.NewSeeded(seed)
		sp, err := GenerateSystemPlacement(r, sys)
		if err != nil {
			continue
		}
		sd, err := DetailSystem(r, sys, sp, IISSClass23Header{})
		if err != nil {
			continue
		}
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if !dp.HZ || dp.GGClass != NotGasGiant {
				continue
			}
			for j := range dp.Moons {
				m := &dp.Moons[j]
				if m.GGClass != NotGasGiant {
					continue
				}
				if m.SizeCode == "" || m.SizeCode == "0" || m.SizeCode == "R" {
					continue
				}
				saw++
				if m.Atmosphere == nil {
					t.Errorf("seed %d moon %s (parent %s HZ, size %s): Atmosphere = nil",
						seed, m.Designation, dp.Designation, m.SizeCode)
				}
				if m.Hydrographics == nil {
					t.Errorf("seed %d moon %s (parent %s HZ, size %s): Hydrographics = nil",
						seed, m.Designation, dp.Designation, m.SizeCode)
				}
			}
		}
	}
	if saw == 0 {
		t.Skip("no HZ-planet terrestrial moons found across 30 seeds; test cannot validate")
	}
}

// TestBuildMoonPlacementView_PropagatesParentGroup asserts that the moonDP
// view inherits the parent's stellar Group. Without this, downstream readers
// (temperature_albedo.go, temperature.go, temperature_rederive.go) fall back
// to sys.Primary.HZCO() instead of the actual parent group's HZCO — wrong
// for multi-star systems where the parent is in a non-primary group.
func TestBuildMoonPlacementView_PropagatesParentGroup(t *testing.T) {
	t.Parallel()

	parent := &DetailedPlacement{
		SizeCode:    "8",
		DiameterKm:  12000,
		Designation: "B II",
		HZ:          true,
	}
	parent.Body = BodyTerrestrial
	parent.Group = Group{Designation: "B"}
	parent.Orbit = 1.5

	m := &Moon{
		Designation: "B II a",
		SizeCode:    "5",
		DiameterKm:  9000,
	}

	moonDP := buildMoonPlacementView(m, parent)

	if moonDP.Group.Designation != "B" {
		t.Errorf("moonDP.Group.Designation = %q, want %q (parent's Group not propagated)",
			moonDP.Group.Designation, parent.Group.Designation)
	}
}

// TestDetailSystem_ScaleHeightRefreshedAfter5E asserts that after DetailSystem,
// any body whose 3B-geology pass added inherent temperature has its
// Atmosphere.ScaleHeight in sync with the post-5E MeanK.
//
// 3A1 computes ScaleHeight from an HZCO-band midpoint estimate. 5D rederive
// refreshes it using the real MeanK from 5C. 5E then bumps MeanK via
// ApplyInherentTempAddition (when total seismic stress is high) — ScaleHeight
// must be refreshed once more or it drifts from the temperature it claims.
func TestDetailSystem_ScaleHeightRefreshedAfter5E(t *testing.T) {
	t.Parallel()

	sys := stars.System{
		Primary: stars.Compose(stars.ComposeOpts{
			Kind:            stars.KindMainSequence,
			SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
			LuminosityClass: stars.V,
			Mass:            1.0,
			Diameter:        1.0,
			Temperature:     5800,
			AgeGyr:          4.6,
		}),
		PrimaryDesignation: "A",
		AgeGyr:             4.6,
	}

	saw := 0
	for iter := range 50 {
		seed := int64(iter)
		r := roller.NewSeeded(seed)
		sp, err := GenerateSystemPlacement(r, sys)
		if err != nil {
			continue
		}
		sd, err := DetailSystem(r, sys, sp, IISSClass23Header{})
		if err != nil {
			continue
		}
		for i := range sd.Detailed {
			dp := &sd.Detailed[i]
			if dp.Atmosphere == nil || dp.Physical == nil ||
				!dp.HasTemperature() || !dp.HasGeology() {
				continue
			}
			if dp.Geology.InherentTemperatureK <= 0 {
				continue
			}
			if dp.Physical.Gravity == 0 {
				continue
			}
			saw++
			want := DeriveScaleHeight(dp.Temperature.MeanK, dp.Physical.Gravity)
			if math.Abs(dp.Atmosphere.ScaleHeight-want)/want > 0.001 {
				t.Errorf("seed %d body %s: ScaleHeight=%v, want %v (DeriveScaleHeight(MeanK=%v, G=%v))",
					seed, dp.Designation, dp.Atmosphere.ScaleHeight, want,
					dp.Temperature.MeanK, dp.Physical.Gravity)
			}
		}
	}
	if saw == 0 {
		t.Skip("no bodies with InherentTemperatureK > 0 in 50 seeds; test cannot validate")
	}
}
