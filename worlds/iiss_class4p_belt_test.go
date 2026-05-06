package worlds

import (
	"strings"
	"testing"

	"wbh/stars"
)

func TestRenderIISS4PBelt_PopulatedFields(t *testing.T) {
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

	got := renderIISS4PBelt(body, sys, "")

	expected := []string{
		"IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B",
		"WORLD: Aab PI   SAH/UWP: 000",
		"PRIMARY OBJECT(S): Aab",
		"SYSTEM AGE (Gyr): 6.336",
		"ORBIT",
		"Span: 0.910 Orbit#s",
		"Period (h): 5613.16",
		"COMPOSITION",
		"m-type%: 25   s-type%: 55   c-type%: 15   other%: 5",
		"Bulk: 4",
		"Major Bodies: Size 1 = 2   Size S = 19",
		"RESOURCES",
		"Rating: 8",
		"MAJOR BODIES",
		"Counts only: 2 size-1 + 19 size-S",
		"COMMENTS",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderIISS4PBelt_NilBeltDetails_DegradesGracefully(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyPlanetoidBelt
	body.Designation = "Empty Belt"
	body.SizeCode = "0"
	body.Group = Group{Designation: "A"}
	sys := stars.System{}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	got := renderIISS4PBelt(body, sys, "")

	expected := []string{
		"IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B",
		"WORLD: Empty Belt",
		"(belt details not generated)",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderIISS4PBelt_MainworldMarker(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyPlanetoidBelt
	body.Designation = "Aab PI"
	body.SizeCode = "0"
	body.Belt = &BeltDetails{}
	sys := stars.System{}

	// Marker present when mainworldDesignation matches.
	got := renderIISS4PBelt(body, sys, "Aab PI")
	if !strings.Contains(got, "This is the system mainworld.") {
		t.Errorf("expected mainworld marker, got:\n%s", got)
	}

	// Marker absent when designation does not match.
	got = renderIISS4PBelt(body, sys, "Other")
	if strings.Contains(got, "This is the system mainworld.") {
		t.Errorf("unexpected mainworld marker in:\n%s", got)
	}
}
