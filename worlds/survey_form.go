package worlds

import (
	"fmt"
	"strconv"
	"strings"

	"wbh/stars"
)

// IISSClass23Form is the structured representation of WBH form 0421D-II.III
// (pp. 60-67). Embeds the existing stars.SurveyForm (Class 0/I header
// + Stars table — extended with MAO via Task 13 Phase A) and adds
// the Class II/III-specific fields: per-body counts, Class III status
// flag, and the Objects table.
//
// Cells dependent on World Physical (sub-project 3) render with "?"
// placeholders until then. Specifically: atmosphere/hydrographics
// digits in HZ-candidate SAH cells; the mainworld marker (asterisk).
type IISSClass23Form struct {
	stars.SurveyForm // header + Stars table (Class 0/I)

	GasGiants      int
	PlanetoidBelts int
	Terrestrials   int
	ClassIIIStatus bool // 2C always renders false (Class II only)

	Objects []ObjectRow
}

// ObjectRow is one row of the WBH p.61 Objects table.
type ObjectRow struct {
	Primary     string // host star group: "Aab", "AB", "B", "Cab"
	Designation string // "Aab I", "Aab IV d"
	Orbit       float64
	AU          float64
	Ecc         float64
	PeriodStr   string // "1.841d" or "8.627y"
	SAH         string // "B??" / "GLE" / "AA6" / "200" / "566*" / "000" / "S"
	Sub         string // significant-moon count, "?" for belt, "" for moon row
	Notes       string // "HZ, R02, S, 1, 1" / "1,200⊕, HZ, 200, S, S, 566*, S"
}

// IISSClass23Header carries the form's header fields not derivable
// from SystemDetail or stars.System.
type IISSClass23Header struct {
	SectorLocation  string
	InitialSurvey   string
	LastUpdated     string
	IISSDesignation string
	Comments        string
}

// RenderIISSClass23 builds the Class II/III form. The Stars table is
// derived from sys via stars.BuildSurveyForm (with MAO post-filled by
// this function); the Objects table is derived from sd.Detailed.
func RenderIISSClass23(sd SystemDetail, sys stars.System, h IISSClass23Header) IISSClass23Form {
	sec, loc := splitSectorLocation(h.SectorLocation)
	meta := stars.SurveyMetadata{
		Sector:        sec,
		Location:      loc,
		Designation:   h.IISSDesignation,
		InitialSurvey: h.InitialSurvey,
		LastUpdated:   h.LastUpdated,
	}
	base := stars.BuildSurveyForm(sys, meta)
	base.Comments = h.Comments

	// Post-fill MAO on Stars rows. stars.BuildSurveyForm leaves MAO=0;
	// worlds fills it from the AvailableOrbits result.
	if avail, err := AvailableOrbits(sys); err == nil {
		fillStarsMAO(&base, avail, sys)
	}

	form := IISSClass23Form{
		SurveyForm:     base,
		GasGiants:      sd.Counts.GasGiants,
		PlanetoidBelts: sd.Counts.PlanetoidBelts,
		Terrestrials:   sd.Counts.Terrestrials,
		ClassIIIStatus: false,
	}

	for _, dp := range sd.Detailed {
		if dp.Body == BodyEmpty {
			continue
		}
		row := ObjectRow{
			Primary:     dp.Group.Designation,
			Designation: dp.Designation,
			Orbit:       dp.Orbit,
			AU:          stars.OrbitToAU(dp.Orbit),
			Ecc:         dp.Eccentricity,
			PeriodStr:   formatPeriod(dp.Period),
			SAH:         renderSAH(dp),
			Sub:         renderSub(dp),
			Notes:       renderNotes(dp),
		}
		form.Objects = append(form.Objects, row)

		for _, m := range dp.Moons {
			form.Objects = append(form.Objects, ObjectRow{
				Primary:     dp.Group.Designation,
				Designation: m.Designation,
				SAH:         renderMoonSAH(m, dp.HZ),
				Sub:         "",
			})
		}
	}

	return form
}

// splitSectorLocation parses "Sector | NNNN" into (sector, location).
func splitSectorLocation(s string) (sector, location string) {
	if idx := strings.Index(s, " | "); idx >= 0 {
		return s[:idx], s[idx+3:]
	}
	return s, ""
}

// fillStarsMAO post-walks Stars rows and copies MAO from AvailableOrbits.
// AB/ABC composite-row MAO is the outer-companion's AU; pulled from
// sys.Companions in a parallel lookup.
func fillStarsMAO(form *stars.SurveyForm, avail Result, sys stars.System) {
	maoByGroup := map[string]float64{}
	for _, g := range avail.Groups {
		maoByGroup[g.Designation] = g.MAO
	}
	composeMAO := map[string]float64{}
	for _, c := range sys.Companions {
		switch c.OrbitClass {
		case stars.OrbitClose, stars.OrbitNear:
			// AB composite row's MAO is the outer non-Far companion's AU.
			// Close and Near both produce an "AB" composite; use the larger
			// AU so a Close+Near system resolves to Near's (outer) distance.
			if c.AU > composeMAO["AB"] {
				composeMAO["AB"] = c.AU
			}
		case stars.OrbitFar:
			composeMAO["ABC"] = c.AU
		}
	}
	for i, row := range form.Stars {
		key := row.Component
		if idx := strings.Index(key, " ("); idx >= 0 {
			key = key[:idx]
		}
		if mao, ok := maoByGroup[key]; ok {
			form.Stars[i].MAO = mao
			continue
		}
		if mao, ok := composeMAO[key]; ok {
			form.Stars[i].MAO = mao
		}
	}
}

// renderSAH produces the SAH/UWP cell for a body.
func renderSAH(dp DetailedPlacement) string {
	switch dp.Body {
	case BodyTerrestrial:
		return string(dp.SizeCode) + "??"
	case BodyGasGiant:
		var prefix string
		switch dp.GGClass {
		case GasGiantSmall:
			prefix = "GS"
		case GasGiantMedium:
			prefix = "GM"
		case GasGiantLarge:
			prefix = "GL"
		default:
			prefix = "G"
		}
		return prefix + dp.GGDiameterCode
	case BodyPlanetoidBelt:
		return "000"
	default:
		return ""
	}
}

// renderMoonSAH renders moon SAH in the Notes column.
// HZ-parent moons get "<Size>??" (atmo/hyd deferred to sub-project 3).
// Non-HZ-parent moons get just the Size letter.
// Gas-giant-sized moons get the full GG class+code.
func renderMoonSAH(m Moon, parentInHZ bool) string {
	if m.GGClass != NotGasGiant {
		var prefix string
		switch m.GGClass {
		case GasGiantSmall:
			prefix = "GS"
		case GasGiantMedium:
			prefix = "GM"
		case GasGiantLarge:
			prefix = "GL"
		}
		return prefix + m.GGDiameterCode
	}
	if parentInHZ {
		return string(m.SizeCode) + "??"
	}
	return string(m.SizeCode)
}

// renderSub renders the Sub column.
func renderSub(dp DetailedPlacement) string {
	if dp.Body == BodyPlanetoidBelt {
		return "?"
	}
	if len(dp.Moons) == 0 {
		return "0"
	}
	return strconv.Itoa(len(dp.Moons))
}

// renderNotes builds the Notes column.
// Gas giants: "<mass>⊕" prefix; HZ tag if HZ; comma-separated moon SAHs.
// Terrestrials: HZ tag if HZ.
// Belts: empty.
func renderNotes(dp DetailedPlacement) string {
	parts := []string{}
	if dp.Body == BodyGasGiant {
		parts = append(parts, fmt.Sprintf("%s⊕", formatMass(dp.MassEarth)))
	}
	if dp.HZ {
		parts = append(parts, "HZ")
	}
	if len(dp.Moons) > 0 {
		moonSAH := make([]string, 0, len(dp.Moons))
		for _, m := range dp.Moons {
			moonSAH = append(moonSAH, renderMoonSAH(m, dp.HZ))
		}
		parts = append(parts, strings.Join(moonSAH, ", "))
	}
	return strings.Join(parts, ", ")
}

// formatMass renders Terra-mass with thousands-separator commas.
func formatMass(m float64) string {
	n := int(m + 0.5)
	s := strconv.Itoa(n)
	if n < 1000 {
		return s
	}
	out := []byte(s)
	for i := len(out) - 3; i > 0; i -= 3 {
		out = append(out[:i], append([]byte{','}, out[i:]...)...)
	}
	return string(out)
}

// formatPeriod renders Period: years < 0.05 → days; ≥1000y → comma format; else years 3-decimal.
func formatPeriod(p Period) string {
	if p.Years > 0 && p.Years < 0.05 {
		return fmt.Sprintf("%.3fd", p.Days)
	}
	if p.Years >= 1000 {
		return formatMass(p.Years) + "y"
	}
	return fmt.Sprintf("%.3fy", p.Years)
}
