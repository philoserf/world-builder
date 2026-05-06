package worlds

import (
	"strings"
	"testing"

	"wbh/stars"
)

func TestRenderSystemMarkdown_AllSectionsWhenMainworldExists(t *testing.T) {
	sd := SystemDetail{
		Survey: IISSClass23Form{
			SurveyForm: stars.SurveyForm{
				IISSDesig:    "Test System 1234",
				SystemAgeGyr: 4.5,
				StellarCount: 1,
				Stars:        []stars.SurveyComponent{{Component: "A", Class: "G2 V", Mass: 1.0, Luminosity: 1.0}},
			},
			Terrestrials: 1,
		},
		MainworldDesignation: "A III",
	}
	sd.Detailed = []DetailedPlacement{{
		Placement: Placement{},
	}}
	sd.Detailed[0].Body = BodyTerrestrial
	sd.Detailed[0].Designation = "A III"
	sd.Detailed[0].SizeCode = "5"
	sd.Detailed[0].Group = Group{Designation: "A"}

	sys := stars.System{Primary: stars.Star{
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.0,
		Luminosity:      1.0,
		AgeGyr:          4.5,
	}}

	got := RenderSystemMarkdown(sd, sys)

	expected := []string{
		"# Star System: Test System 1234",
		"## IISS Class 0/I Survey — Form 0421B-0I",
		"## IISS Class II/III Survey — Form 0421D-II.III",
		"## IISS Class IV-P Survey — Form 0407F-IV PART P",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}

	// H2s must appear in book order: 0/I before II/III before IV-P.
	idx0I := strings.Index(got, "## IISS Class 0/I")
	idx23 := strings.Index(got, "## IISS Class II/III")
	idx4P := strings.Index(got, "## IISS Class IV-P")
	if idx0I >= idx23 || idx23 >= idx4P {
		t.Errorf("H2 sections not in book order: 0/I=%d, II/III=%d, IV-P=%d", idx0I, idx23, idx4P)
	}
}

func TestRenderSystemMarkdown_OmitsClass4PWhenNoMainworld(t *testing.T) {
	sd := SystemDetail{
		Survey: IISSClass23Form{
			SurveyForm: stars.SurveyForm{IISSDesig: "Empty System", StellarCount: 1},
		},
		MainworldDesignation: "",
	}
	sys := stars.System{Primary: stars.Star{AgeGyr: 4.5}}

	got := RenderSystemMarkdown(sd, sys)

	if !strings.Contains(got, "## IISS Class 0/I Survey") {
		t.Errorf("missing Class 0/I H2; got:\n%s", got)
	}
	if !strings.Contains(got, "## IISS Class II/III Survey") {
		t.Errorf("missing Class II/III H2; got:\n%s", got)
	}
	if strings.Contains(got, "## IISS Class IV-P Survey") {
		t.Errorf("Class IV-P H2 must be absent when MainworldDesignation == \"\"; got:\n%s", got)
	}
}

func TestRenderSystemMarkdown_BeltMainworldUsesPartPB(t *testing.T) {
	sd := SystemDetail{
		Survey: IISSClass23Form{
			SurveyForm: stars.SurveyForm{IISSDesig: "Belt System", StellarCount: 1},
		},
		MainworldDesignation: "A PI",
	}
	belt := DetailedPlacement{Placement: Placement{}}
	belt.Body = BodyPlanetoidBelt
	belt.Designation = "A PI"
	belt.SizeCode = "0"
	belt.Belt = &BeltDetails{ResourceRating: 8}
	sd.Detailed = []DetailedPlacement{belt}

	sys := stars.System{Primary: stars.Star{AgeGyr: 4.5}}

	got := RenderSystemMarkdown(sd, sys)

	if !strings.Contains(got, "Form 0407K-IV PART P.B") {
		t.Errorf("expected belt-variant Form 0407K-IV PART P.B; got:\n%s", got)
	}
	if strings.Contains(got, "Form 0407F-IV PART P\n") {
		t.Errorf("should not emit terrestrial PART P for a belt mainworld; got:\n%s", got)
	}
}
