package worlds

import (
	"fmt"
	"strings"

	"wbh/stars"
)

// RenderSystemMarkdown is the top-level Markdown renderer. Output:
//
//   - H1 system title
//   - H2 IISS Class 0/I Survey
//   - H2 IISS Class II/III Survey
//   - H2 IISS Class IV-P Survey (omitted when sd.MainworldDesignation == "")
//
// Class 0/I is read from sd.Survey.SurveyForm (already populated by
// DetailSystem); Class II/III is read from sd.Survey; Class IV-P
// resolves the mainworld body from sd.Detailed and dispatches to the
// right variant (PART P / PART P.B) based on its SizeCode.
func RenderSystemMarkdown(sd SystemDetail, sys stars.System) string {
	var sb strings.Builder
	title := sd.Survey.IISSDesig
	if title == "" {
		title = "(unnamed)"
	}
	fmt.Fprintf(&sb, "# Star System: %s\n\n", title)

	// Class 0/I — embedded in sd.Survey.
	sb.WriteString(stars.RenderClass0IMarkdown(sd.Survey.SurveyForm))

	// Class II/III — already on sd.Survey.
	sb.WriteString(RenderClass23Markdown(sd.Survey))

	// Class IV-P — only when a mainworld was picked.
	if sd.MainworldDesignation != "" {
		if mw := findMainworld(sd, sd.MainworldDesignation); mw != nil {
			sb.WriteString(RenderClass4PMarkdown(mw, sys, sd.MainworldDesignation))
		}
	}

	return sb.String()
}

// findMainworld locates the body in sd.Detailed (or its moons) whose
// Designation matches mainworld. Returns nil if not found.
func findMainworld(sd SystemDetail, mainworld string) *DetailedPlacement {
	for i := range sd.Detailed {
		if sd.Detailed[i].Designation == mainworld {
			return &sd.Detailed[i]
		}
		for j := range sd.Detailed[i].Moons {
			m := &sd.Detailed[i].Moons[j]
			if m.Designation == mainworld {
				// Synthesize a DetailedPlacement view of the moon for
				// rendering — same pattern as buildMoonPlacementView in
				// system_detail.go.
				return buildMoonPlacementView(m, &sd.Detailed[i])
			}
		}
	}
	return nil
}
