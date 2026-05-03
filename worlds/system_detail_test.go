package worlds

import (
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
