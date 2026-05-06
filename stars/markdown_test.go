package stars

import (
	"strings"
	"testing"
)

func TestRenderClass0IMarkdown_PopulatedFields(t *testing.T) {
	form := SurveyForm{
		Sector:        "Storr",
		Location:      "0602",
		IISSDesig:     "566-837 (Zed Prime)",
		InitialSurvey: "207-568",
		LastUpdated:   "218-1061",
		SystemAgeGyr:  6.336,
		StellarCount:  5,
		Stars: []SurveyComponent{
			{Component: "Aa", Class: "G7 V", Mass: 0.929, Temperature: 5440, Diameter: 0.967, Luminosity: 0.738},
			{Component: "Ab", Class: "G8 V", Mass: 0.907, Temperature: 5360, Diameter: 0.957, Luminosity: 0.681, Orbit: 0.09, AU: 0.036, Eccentricity: 0.11, PeriodYears: 0.005},
			{Component: "Aab (A)", Class: "—", Mass: 1.836, Luminosity: 1.419, Orbit: 0.09, AU: 0.036, Eccentricity: 0.11, PeriodYears: 0.005, HZCO: 3.3},
		},
	}

	got := RenderClass0IMarkdown(form)

	expected := []string{
		"## IISS Class 0/I Survey — Form 0421B-0I",
		"### Header",
		"| Field | Value |",
		"| Sector | Storr |",
		"| Location | 0602 |",
		"| IISS Designation | 566-837 (Zed Prime) |",
		"| System Age (Gyr) | 6.336 |",
		"| Stellar Count | 5 |",
		"### Stars",
		"| Component | Class | Mass | Temperature | Diameter | Luminosity | Orbit | AU | Eccentricity | Period (y) | HZCO |",
		"| Aa | G7 V | 0.929 |",
		"| Aab (A) | — | 1.836 |",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderClass0IMarkdown_EmptyMetadataFields(t *testing.T) {
	// Metadata fields left blank should render as em-dash.
	form := SurveyForm{
		SystemAgeGyr: 4.5,
		StellarCount: 1,
		Stars:        []SurveyComponent{{Component: "A", Class: "G2 V", Mass: 1.0}},
	}
	got := RenderClass0IMarkdown(form)
	if !strings.Contains(got, "| Sector | — |") {
		t.Errorf("empty Sector should render as em-dash; got:\n%s", got)
	}
	if !strings.Contains(got, "| IISS Designation | — |") {
		t.Errorf("empty IISSDesig should render as em-dash; got:\n%s", got)
	}
}
