package worlds

import (
	"fmt"
	"strings"

	"wbh/stars"
)

// renderIISS4PBelt renders the WBH p.139 IISS Class IV Survey form
// (FORM 0407K-IV PART P.B) for a Size-0 planetoid belt body. Plain-text
// output with section headers matching the book's form layout.
//
// mainworldDesignation is the SystemDetail.MainworldDesignation; used
// only to mark whether THIS body is the mainworld in the COMMENTS section.
//
// Returns "" if body is nil. Defensive against nil *BeltDetails: the
// COMPOSITION, RESOURCES, and MAJOR BODIES sections render a placeholder
// rather than panicking. Header and ORBIT sections always render.
func renderIISS4PBelt(body *DetailedPlacement, sys stars.System, mainworldDesignation string) string {
	if body == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("IISS CLASS IV SURVEY — FORM 0407K-IV PART P.B\n\n")

	// Header
	fmt.Fprintf(&sb, "WORLD: %s   SAH/UWP: 000\n", body.Designation)
	sb.WriteString("SECTOR | LOCATION:    INITIAL SURVEY:    LAST UPDATED:\n")
	fmt.Fprintf(&sb, "PRIMARY OBJECT(S): %s    SYSTEM AGE (Gyr): %.3f    TRAVEL ZONE:\n\n",
		body.Group.Designation, sys.Primary.AgeGyr)

	// Orbit
	sb.WriteString("ORBIT\n")
	spanStr := "(not available)"
	if body.Belt != nil {
		spanStr = fmt.Sprintf("%.3f Orbit#s", body.Belt.Span)
	}
	fmt.Fprintf(&sb, "  O#: %.2f   AU: %.2f   Span: %s   Period (h): %.2f\n\n",
		body.Orbit, stars.OrbitToAU(body.Orbit), spanStr, body.Period.Hours)

	// Composition
	sb.WriteString("COMPOSITION\n")
	if body.Belt == nil {
		sb.WriteString("  (belt details not generated)\n\n")
	} else {
		c := body.Belt.Composition
		fmt.Fprintf(&sb, "  m-type%%: %d   s-type%%: %d   c-type%%: %d   other%%: %d\n",
			c.MTypePct, c.STypePct, c.CTypePct, c.OtherPct)
		fmt.Fprintf(&sb, "  Bulk: %d\n", body.Belt.Bulk)
		fmt.Fprintf(&sb, "  Major Bodies: Size 1 = %d   Size S = %d\n\n",
			body.Belt.SigSize1Bodies, body.Belt.SigSizeSBodies)
	}

	// Resources
	sb.WriteString("RESOURCES\n")
	if body.Belt == nil {
		sb.WriteString("  (belt details not generated)\n\n")
	} else {
		fmt.Fprintf(&sb, "  Rating: %d\n\n", body.Belt.ResourceRating)
	}

	// Major Bodies (subtable summarized — per-body detail not generated; see spec non-goals)
	sb.WriteString("MAJOR BODIES\n")
	if body.Belt == nil {
		sb.WriteString("  (belt details not generated)\n\n")
	} else {
		fmt.Fprintf(&sb, "  Counts only: %d size-1 + %d size-S; per-body detail not generated.\n\n",
			body.Belt.SigSize1Bodies, body.Belt.SigSizeSBodies)
	}

	// Comments
	sb.WriteString("COMMENTS\n")
	if mainworldDesignation != "" && body.Designation == mainworldDesignation {
		sb.WriteString("  This is the system mainworld.\n")
	}

	return sb.String()
}
