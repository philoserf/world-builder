package worlds

import (
	"strings"
	"testing"

	"wbh/stars"
)

func TestRenderIISSClass23_Header(t *testing.T) {
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
	sd := SystemDetail{
		SystemPlacement: SystemPlacement{
			Counts: Counts{GasGiants: 1, PlanetoidBelts: 1, Terrestrials: 4, Total: 6},
		},
	}
	header := IISSClass23Header{
		SectorLocation:  "Sol Sector | 0801",
		InitialSurvey:   "001-001",
		LastUpdated:     "218-1061",
		IISSDesignation: "Sol (system)",
		Comments:        "Reference system.",
	}

	form := RenderIISSClass23(sd, sys, header)

	if form.IISSDesig != "Sol (system)" {
		t.Errorf("IISSDesig = %q, want \"Sol (system)\"", form.IISSDesig)
	}
	if form.SystemAgeGyr != 4.6 {
		t.Errorf("SystemAgeGyr = %v, want 4.6", form.SystemAgeGyr)
	}
	if form.GasGiants != 1 || form.PlanetoidBelts != 1 || form.Terrestrials != 4 {
		t.Errorf("counts = (%d, %d, %d), want (1, 1, 4)",
			form.GasGiants, form.PlanetoidBelts, form.Terrestrials)
	}
	if form.ClassIIIStatus {
		t.Error("ClassIIIStatus = true, want false")
	}
	if form.Comments != "Reference system." {
		t.Errorf("Comments = %q, want \"Reference system.\"", form.Comments)
	}
}

func TestRenderIISSClass23_ObjectsTerrestrialRow(t *testing.T) {
	t.Parallel()

	dps := []DetailedPlacement{
		{
			Placement: Placement{
				Body: BodyTerrestrial,
				AnomalousSlot: AnomalousSlot{Slot: Slot{
					Orbit: 1.0,
					Group: Group{Designation: "Aab"},
				}},
			},
			Designation: "Aab I",
			SizeCode:    "B",
			DiameterKm:  17600,
			Period:      Period{Years: 0.187, Days: 68.3},
			HZ:          false,
		},
	}
	sd := SystemDetail{Detailed: dps}
	form := RenderIISSClass23(sd, stars.System{}, IISSClass23Header{})
	if len(form.Objects) != 1 {
		t.Fatalf("len(Objects) = %d, want 1", len(form.Objects))
	}
	row := form.Objects[0]
	if row.Designation != "Aab I" {
		t.Errorf("Designation = %q, want \"Aab I\"", row.Designation)
	}
	if row.SAH != "B??" {
		t.Errorf("SAH = %q, want \"B??\"", row.SAH)
	}
	if row.PeriodStr != "0.187y" {
		t.Errorf("PeriodStr = %q, want \"0.187y\"", row.PeriodStr)
	}
}

func TestRenderIISSClass23_ObjectsGasGiantRow(t *testing.T) {
	t.Parallel()

	dps := []DetailedPlacement{
		{
			Placement: Placement{
				Body: BodyGasGiant,
				AnomalousSlot: AnomalousSlot{Slot: Slot{
					Orbit: 3.1,
					Group: Group{Designation: "Aab"},
				}},
			},
			Designation:    "Aab IV",
			GGClass:        GasGiantLarge,
			GGDiameterCode: "E",
			DiameterEarth:  14,
			MassEarth:      1200,
			Period:         Period{Years: 0.805, Days: 294.0},
			HZ:             true,
		},
	}
	sd := SystemDetail{Detailed: dps}
	form := RenderIISSClass23(sd, stars.System{}, IISSClass23Header{})
	row := form.Objects[0]
	if row.SAH != "GLE" {
		t.Errorf("SAH = %q, want \"GLE\"", row.SAH)
	}
	if !strings.Contains(row.Notes, "1,200⊕") {
		t.Errorf("Notes = %q, want contains \"1,200⊕\"", row.Notes)
	}
	if !strings.Contains(row.Notes, "HZ") {
		t.Errorf("Notes = %q, want contains \"HZ\"", row.Notes)
	}
}

func TestRenderIISSClass23_ObjectsBeltRow(t *testing.T) {
	t.Parallel()

	dps := []DetailedPlacement{
		{
			Placement: Placement{
				Body: BodyPlanetoidBelt,
				AnomalousSlot: AnomalousSlot{Slot: Slot{
					Orbit: 2.7,
					Group: Group{Designation: "Aab"},
				}},
			},
			Designation: "Aab PI",
		},
	}
	sd := SystemDetail{Detailed: dps}
	form := RenderIISSClass23(sd, stars.System{}, IISSClass23Header{})
	row := form.Objects[0]
	if row.SAH != "000" {
		t.Errorf("SAH = %q, want \"000\"", row.SAH)
	}
	if row.Sub != "?" {
		t.Errorf("Sub = %q, want \"?\"", row.Sub)
	}
}

func TestRenderIISSClass23_PeriodFormatting(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		p    Period
		want string
	}{
		{"sub-day → days", Period{Years: 0.005, Days: 1.826}, "1.826d"},
		{"sub-fortnight → days", Period{Years: 0.04, Days: 14.61}, "14.610d"},
		{"normal → years", Period{Years: 0.187, Days: 68.3}, "0.187y"},
		{"big → years with comma", Period{Years: 3598.0, Days: 0}, "3,598y"},
		{"medium → years 3 decimal", Period{Years: 8.627, Days: 0}, "8.627y"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := formatPeriod(tc.p)
			if got != tc.want {
				t.Errorf("formatPeriod(%v) = %q, want %q", tc.p, got, tc.want)
			}
		})
	}
}

func TestRenderIISSClass23_HZCandidateMaskedSAH(t *testing.T) {
	t.Parallel()

	dps := []DetailedPlacement{
		{
			Placement: Placement{
				Body: BodyTerrestrial,
				AnomalousSlot: AnomalousSlot{Slot: Slot{
					Orbit: 4.1,
					Group: Group{Designation: "Aab"},
				}},
			},
			Designation: "Aab VI",
			SizeCode:    "A",
			DiameterKm:  16000,
			HZ:          true,
		},
	}
	sd := SystemDetail{Detailed: dps}
	form := RenderIISSClass23(sd, stars.System{}, IISSClass23Header{})
	row := form.Objects[0]
	if row.SAH != "A??" {
		t.Errorf("HZ-candidate SAH = %q, want \"A??\"", row.SAH)
	}
}
