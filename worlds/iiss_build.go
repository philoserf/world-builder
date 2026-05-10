package worlds

import (
	"fmt"
	"strings"

	"wbh/iiss"
	"wbh/stars"
)

// BuildIISSForms populates u.Detail.SystemForms (Class0I, Class23,
// Class4P) from the converged universe state. Pure function — no
// rolls. Called by GenerateWithRoller after AggregateSystem.
//
// Cycle-11 MVP: produces structurally-correct forms with the
// header / counts / star rows / object rows populated; cell-by-cell
// fidelity to the WBH p.35 / p.63 / pp.141-142 layouts lands in a
// post-parity sub-project.
func BuildIISSForms(u *Universe) {
	u.Detail.Class0I = buildClass0I(u)
	u.Detail.Class23 = buildClass23(u)
	u.Detail.Class4P = buildClass4P(u)
}

func buildClass0I(u *Universe) iiss.Class0IForm {
	form := iiss.Class0IForm{
		FormHeader:   buildHeader(u),
		SystemAgeGyr: u.System.Primary.AgeGyr,
		StellarCount: 1 + len(u.System.Companions),
	}
	form.Stars = append(form.Stars, starRow("A", u.System.Primary))
	for _, comp := range u.System.Companions {
		designation := comp.Designation
		if designation == "" {
			designation = "?"
		}
		form.Stars = append(form.Stars, starRow(designation, comp.Star))
	}
	return form
}

func buildClass23(u *Universe) iiss.Class23Form {
	form := iiss.Class23Form{
		FormHeader:   buildHeader(u),
		SystemAgeGyr: u.System.Primary.AgeGyr,
		StellarCount: 1 + len(u.System.Companions),
		Stars:        buildClass0I(u).Stars,
		Counts: iiss.Class23Counts{
			GasGiants:      u.Placement.Counts.GasGiants,
			PlanetoidBelts: u.Placement.Counts.PlanetoidBelts,
			Terrestrials:   u.Placement.Counts.Terrestrials,
			Total:          u.Placement.Counts.Total,
		},
	}
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if body.Kind == BodyEmpty {
			continue
		}
		form.Objects = append(form.Objects, iiss.Class23Object{
			Designation: body.Designation,
			Notes:       buildBodyNotes(body),
		})
		for _, child := range body.Children {
			form.Objects = append(form.Objects, iiss.Class23Object{
				Designation: child.Designation,
				Notes:       buildBodyNotes(child),
			})
		}
	}
	return form
}

func buildClass4P(u *Universe) iiss.Class4PForm {
	form := iiss.Class4PForm{
		FormHeader: buildHeader(u),
	}
	mainworld := findMainworld(u)
	if mainworld == nil {
		return form
	}
	switch mainworld.Kind {
	case BodyPlanetoidBelt:
		form.Variant = iiss.Class4PBelt
		form.PartPB = &iiss.Class4PPartPB{}
	case BodyMoon:
		form.Variant = iiss.Class4PMoon
		form.PartP = &iiss.Class4PPartP{}
	default:
		form.Variant = iiss.Class4PPlanet
		form.PartP = &iiss.Class4PPartP{}
	}
	return form
}

func findMainworld(u *Universe) *Body {
	if u.Detail.MainworldDesignation == "" {
		return nil
	}
	target := u.Detail.MainworldDesignation
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if body.Designation == target {
			return body
		}
		for _, child := range body.Children {
			if child.Designation == target {
				return child
			}
		}
	}
	return nil
}

func buildHeader(u *Universe) iiss.FormHeader {
	return iiss.FormHeader{
		SystemName:    u.System.PrimaryDesignation,
		Sector:        "—",
		Location:      "—",
		IISSDesig:     u.Detail.MainworldDesignation,
		InitialSurvey: "—",
		LastUpdated:   "—",
	}
}

func starRow(component string, s stars.Star) iiss.Class0IStarRow {
	return iiss.Class0IStarRow{
		Component:   component,
		Class:       formatStarClass(s),
		Mass:        s.Mass,
		Diameter:    s.Diameter,
		Temperature: s.Temperature,
		Luminosity:  s.Luminosity,
		HZCO:        s.HZCO(),
	}
}

func formatStarClass(s stars.Star) string {
	if s.Kind == stars.KindWhiteDwarf {
		return "D"
	}
	letter := s.SpectralType.Letter
	if letter == 0 {
		letter = '?'
	}
	return fmt.Sprintf("%c%d %s", letter, s.SpectralType.Subtype, formatLumClass(s.LuminosityClass))
}

func formatLumClass(c stars.LuminosityClass) string {
	switch c {
	case stars.Ia:
		return "Ia"
	case stars.Ib:
		return "Ib"
	case stars.II:
		return "II"
	case stars.III:
		return "III"
	case stars.IV:
		return "IV"
	case stars.V:
		return "V"
	case stars.VI:
		return "VI"
	case stars.D:
		return "D"
	}
	return "?"
}

func buildBodyNotes(body *Body) string {
	parts := []string{}
	switch body.Kind {
	case BodyTerrestrial:
		parts = append(parts, "Terr")
	case BodyMoon:
		parts = append(parts, "Moon")
	case BodyGasGiant:
		parts = append(parts, "GG")
	case BodyPlanetoidBelt:
		parts = append(parts, "Belt")
	}
	if body.HZ {
		parts = append(parts, "HZ")
	}
	if body.SizeCode != "" {
		parts = append(parts, fmt.Sprintf("Size %s", body.SizeCode))
	}
	if body.HasAtmosphere() && body.Atmosphere.Code > 0 {
		parts = append(parts, fmt.Sprintf("Atm %X", body.Atmosphere.Code))
	}
	if body.HasHydrographics() {
		parts = append(parts, fmt.Sprintf("Hyd %d", body.Hydrographics.Code))
	}
	if body.HasHabitability() {
		parts = append(parts, fmt.Sprintf("Hab %d", body.Habitability.Rating))
	}
	return strings.Join(parts, " ")
}
