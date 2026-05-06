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

func TestRenderClass4PMarkdown_SubordinatesTableRendersMoons(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab IV"
	body.SizeCode = "8"
	body.Group = Group{Designation: "Aab"}
	body.Moons = []Moon{
		{Designation: "Aab IV a", SizeCode: "2", DiameterKm: 3200, OrbitKm: 22000, Eccentricity: 0.015, PeriodHours: 28.5},
		{Designation: "Aab IV d", SizeCode: "5", DiameterKm: 8163, OrbitKm: 3942400, Eccentricity: 0.25, PeriodHours: 624.69},
	}
	sys := stars.System{Primary: stars.Star{AgeGyr: 6.336}}

	got := RenderClass4PMarkdown(body, sys, "")

	expected := []string{
		"### Subordinates",
		"| Designation | Size | Diameter (km) | Orbit (km) | Eccentricity | Period (h) |",
		"| Aab IV a | 2 | 3200 | 22000 | 0.015 | 28.50 |",
		"| Aab IV d | 5 | 8163 | 3942400 | 0.250 | 624.69 |",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderClass4PMarkdown_NoMoons_OmitsSubordinatesSection(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab IV"
	body.SizeCode = "8"
	body.Group = Group{Designation: "Aab"}
	// No moons.
	sys := stars.System{Primary: stars.Star{AgeGyr: 4.5}}

	got := RenderClass4PMarkdown(body, sys, "")
	if strings.Contains(got, "### Subordinates") {
		t.Errorf("Subordinates section should be absent for body with no moons; got:\n%s", got)
	}
}

func TestRenderClass4PMarkdown_BodyEmpty_ReturnsEmpty(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyEmpty
	body.Designation = "Empty Slot"

	got := RenderClass4PMarkdown(body, stars.System{}, "")
	if got != "" {
		t.Errorf("BodyEmpty should return empty string; got %q", got)
	}
}
