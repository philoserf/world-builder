package worlds

import (
	"strings"
	"testing"

	"wbh/stars"
)

func TestRenderClass4PMarkdown_Terrestrial(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab IV d"
	body.SizeCode = "5"
	body.DiameterKm = 8163
	body.MassEarth = 0.27
	body.Orbit = 3.1
	body.Eccentricity = 0.10
	body.Period = Period{Hours: 7050, Years: 0.805}
	body.Group = Group{Designation: "Aab"}
	body.Physical = &BodyPhysical{Density: 1.03, Gravity: 0.66}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.042, OxygenPartialPressure: 0.292, ScaleHeight: 12.88}
	body.Hydrographics = &Hydrographics{Code: 6, Percent: 62, Profile: "5"}
	sys := stars.System{Primary: stars.Star{AgeGyr: 6.336}}

	got := RenderClass4PMarkdown(body, sys, "Aab IV d")

	expected := []string{
		"## IISS Class IV-P Survey — Form 0407F-IV PART P",
		"### Header",
		"| World | Aab IV d |",
		"| Primary Object(s) | Aab |",
		"| System Age (Gyr) | 6.336 |",
		"### Orbit",
		"| AU | 1.06 |",
		"### Size",
		"| Diameter (km) | 8163 |",
		"| Density | 1.030 |",
		"| Gravity | 0.660 |",
		"### Atmosphere",
		"| Code | 6 |",
		"| Pressure (bar) | 1.042 |",
		"### Hydrographics",
		"| Code | 6 |",
		"| Coverage (%) | 62 |",
		"### Comments",
		"This is the system mainworld.",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderClass4PMarkdown_Belt(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyPlanetoidBelt
	body.Designation = "Aab PI"
	body.SizeCode = "0"
	body.Orbit = 2.7
	body.Period = Period{Hours: 5613.16}
	body.Group = Group{Designation: "Aab"}
	body.Belt = &BeltDetails{
		Span:           0.91,
		Composition:    BeltComposition{MTypePct: 25, STypePct: 55, CTypePct: 15, OtherPct: 5},
		Bulk:           4,
		ResourceRating: 8,
		SigSize1Bodies: 2,
		SigSizeSBodies: 19,
	}
	sys := stars.System{Primary: stars.Star{AgeGyr: 6.336}}

	got := RenderClass4PMarkdown(body, sys, "")

	expected := []string{
		"## IISS Class IV-P Survey — Form 0407K-IV PART P.B",
		"| World | Aab PI |",
		"| SAH/UWP | 000 |",
		"### Composition",
		"| m-type (%) | 25 |",
		"| s-type (%) | 55 |",
		"| c-type (%) | 15 |",
		"| other (%) | 5 |",
		"| Bulk | 4 |",
		"| Major Bodies (Size 1) | 2 |",
		"| Major Bodies (Size S) | 19 |",
		"### Resources",
		"| Rating | 8 |",
		"### Major Bodies",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderClass4PMarkdown_NilSubSections_RenderPlaceholders(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Bare Body"
	body.SizeCode = "5"
	body.Group = Group{Designation: "A"}
	// No Atmosphere, Hydrographics, Geology, Biology, Habitability, etc.
	sys := stars.System{}

	got := RenderClass4PMarkdown(body, sys, "")

	expected := []string{
		"### Atmosphere",
		"| Status | (not generated) |",
		"### Hydrographics",
		"### Temperature",
		"### Life",
		"### Habitability",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderClass4PMarkdown_MainworldMarker(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyPlanetoidBelt
	body.Designation = "Aab PI"
	body.SizeCode = "0"
	body.Belt = &BeltDetails{}
	sys := stars.System{}

	withMarker := RenderClass4PMarkdown(body, sys, "Aab PI")
	if !strings.Contains(withMarker, "This is the system mainworld.") {
		t.Errorf("expected mainworld marker; got:\n%s", withMarker)
	}

	withoutMarker := RenderClass4PMarkdown(body, sys, "Other")
	if strings.Contains(withoutMarker, "This is the system mainworld.") {
		t.Errorf("unexpected mainworld marker in:\n%s", withoutMarker)
	}
}
