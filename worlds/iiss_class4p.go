// Package worlds — IISS Class IV-P Survey Form rendering per WBH pp.138-146
// (sub-project 3B-final).
package worlds

import (
	"fmt"
	"strings"

	"wbh/stars"
)

// RenderIISSClass4P renders the WBH p.138 IISS Class IV Survey form
// (FORM 0407F-IV PART P) for a single body. Plain-text output with
// section headers matching the book's form layout.
//
// The third argument is the SystemDetail.MainworldDesignation; it is
// reserved for the Comments section (Tasks 7/8) and unused here.
//
// Returns "" if body is nil or body.Body == BodyEmpty.
//
// Belt bodies (Size 0) get a placeholder stub; full Form 0407K-IV PART P.B
// rendering is deferred (see spec carry-forwards).
func RenderIISSClass4P(body *DetailedPlacement, sys stars.System, _ string) string {
	if body == nil || body.Body == BodyEmpty {
		return ""
	}
	// Belt stub.
	if body.SizeCode == "0" {
		return renderBeltStub(body)
	}

	var sb strings.Builder
	sb.WriteString("IISS CLASS IV SURVEY — FORM 0407F-IV PART P\n\n")
	renderIISS4PWorld(&sb, body, sys)
	renderIISS4POrbit(&sb, body)
	renderIISS4PSize(&sb, body)
	renderIISS4PAtmosphere(&sb, body)

	// Tasks 7 and 8 append the remaining sections + Comments.
	return sb.String()
}

func renderBeltStub(body *DetailedPlacement) string {
	return fmt.Sprintf(`IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B (NOT YET IMPLEMENTED)

WORLD: %s   (Belt)
Use BeltDetails for resource/composition data.
`, body.Designation)
}

func renderIISS4PWorld(sb *strings.Builder, body *DetailedPlacement, sys stars.System) {
	fmt.Fprintf(sb, "WORLD: %s\n", body.Designation)
	fmt.Fprintf(sb, "SECTOR | LOCATION:    INITIAL SURVEY:    LAST UPDATED:\n")
	fmt.Fprintf(sb, "PRIMARY OBJECT(S):    SYSTEM AGE (Gyr): %.3f    TRAVEL ZONE:\n\n",
		sys.Primary.AgeGyr)
}

func renderIISS4POrbit(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("ORBIT\n")
	fmt.Fprintf(sb, "  AU: %.2f   Eccentricity: %.2f   Period (h): %.2f\n\n",
		stars.OrbitToAU(body.Orbit), body.Eccentricity, body.Period.Hours)
}

func renderIISS4PSize(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("SIZE\n")
	density := 0.0
	gravity := 0.0
	if body.Physical != nil {
		density = body.Physical.Density
		gravity = body.Physical.Gravity
	}
	fmt.Fprintf(sb, "  Diameter (km): %.0f   Density: %.2f   Gravity: %.2f   Mass: %.2f\n\n",
		body.DiameterKm, density, gravity, body.MassEarth)
}

func renderIISS4PAtmosphere(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("ATMOSPHERE\n")
	if body.Atmosphere == nil {
		sb.WriteString("  (none — vacuum)\n\n")
		return
	}
	atm := body.Atmosphere
	fmt.Fprintf(sb, "  Code: %d   Pressure (bar): %.3f   O2 (bar): %.3f   Scale Height: %.2f\n\n",
		atm.Code, atm.Pressure, atm.OxygenPartialPressure, atm.ScaleHeight)
}
