package worlds

import (
	"fmt"
	"strings"

	"github.com/philoserf/world-builder/iiss"
	"github.com/philoserf/world-builder/stars"
)

// Class4PPartP is the planet/moon mainworld view consumed by the
// Class IV-P renderer and emitted as JSON. Shape matches the WBH
// p.138 Form 0407F-IV PART P layout.
type Class4PPartP struct {
	Designation  string
	SystemAgeGyr float64

	OrbitNumber  float64
	AU           float64
	Eccentricity float64
	PeriodHours  float64

	// Moon-mainworld extras (Kind == BodyMoon): the moon's own orbit
	// around its parent. Zero/empty for planet mainworlds.
	MoonOrbitKm       float64 `json:",omitempty"`
	ParentDesignation string  `json:",omitempty"`

	DiameterKm float64
	Density    float64
	Gravity    float64
	MassEarth  float64

	Atmosphere    *Class4PAtmosphere    `json:",omitempty"`
	Hydrographics *Class4PHydrographics `json:",omitempty"`

	SiderealHours    float64
	SolarHours       float64
	SolarDaysPerYear float64
	AxialTiltDeg     float64
	TidalLockRatio   string
	TidesMeters      float64

	Temperature *Class4PTemperature `json:",omitempty"`
	Seismic     *Class4PSeismic     `json:",omitempty"`
	Life        *Class4PLife        `json:",omitempty"`

	HabitabilityRating int
	HabitabilityNotes  string

	Subordinates []Class4PSubordinate `json:",omitempty"`

	IsMainworld bool
}

// Class4PAtmosphere captures the WBH p.138 ATMOSPHERE block.
type Class4PAtmosphere struct {
	Code                  int
	Pressure              float64
	OxygenPartialPressure float64
	ScaleHeight           float64
	ProfileShorthand      string
}

// Class4PHydrographics captures the WBH p.138 HYDROGRAPHICS block.
type Class4PHydrographics struct {
	Code    int
	Percent int
	Profile string
}

// Class4PTemperature captures the WBH p.138 TEMPERATURE block.
// LowK == -1 is the sentinel for "—" (degenerate-model).
type Class4PTemperature struct {
	HighK            float64
	MeanK            float64
	LowK             float64
	Luminosity       float64
	Albedo           float64
	GreenhouseFactor float64
}

// Class4PSeismic captures the WBH p.138 SEISMIC block.
type Class4PSeismic struct {
	TotalSeismicStress    int
	ResidualSeismicStress int
	TidalStressFactor     int
	TidalHeatingFactor    int
	TectonicPlates        int
}

// Class4PLife captures the WBH p.138 LIFE + RESOURCES blocks.
type Class4PLife struct {
	Biomass        int
	Biocomplexity  int
	HasSophont     bool
	HadExtinct     bool
	Biodiversity   int
	Compatibility  int
	ResourceRating int
}

// Class4PSubordinate is one row of the WBH p.138 SUBORDINATES table
// (a planet mainworld's moons).
type Class4PSubordinate struct {
	Designation  string
	SizeCode     SizeCode
	DiameterKm   float64
	OrbitKm      int
	Eccentricity float64
	PeriodHours  float64
}

// Class4PPartPB is the belt mainworld view (WBH p.139 FORM 0407K-IV
// PART P.B). Same role as Class4PPartP for the belt variant.
type Class4PPartPB struct {
	Designation  string
	PrimaryGroup string
	SystemAgeGyr float64

	OrbitNumber float64
	AU          float64
	SpanOrbits  float64
	PeriodHours float64

	MTypePct       int
	STypePct       int
	CTypePct       int
	OtherPct       int
	Bulk           int
	SigSize1Bodies int
	SigSizeSBodies int

	ResourceRating int

	IsMainworld bool
}

// buildClass4PPlanet builds the planet/moon mainworld view. The
// Markdown renderer is the RenderBody method on the result — both
// read the same struct, so adding a Class IV-P field is one struct
// edit plus one renderer edit, both in this file.
func buildClass4PPlanet(u *Universe, body *Body) *Class4PPartP {
	p := &Class4PPartP{
		Designation:  body.Designation,
		SystemAgeGyr: u.System.Primary.AgeGyr,
		OrbitNumber:  body.Orbit,
		AU:           stars.OrbitToAU(body.Orbit),
		Eccentricity: body.Eccentricity,
		PeriodHours:  body.Period.Hours,
		DiameterKm:   body.DiameterKm,
		MassEarth:    body.MassEarth,
		IsMainworld:  true,
	}
	if body.Kind == BodyMoon {
		// Body.Orbit and Body.Period are star-relative and zero for
		// moons — use the parent's stellar orbit for the AU context and
		// the moon's own parent-relative orbit and period.
		p.OrbitNumber = body.StellarOrbit()
		p.AU = stars.OrbitToAU(body.StellarOrbit())
		p.PeriodHours = body.PeriodHours
		p.MoonOrbitKm = body.OrbitKm
		if body.Parent != nil {
			p.ParentDesignation = body.Parent.Designation
		}
	}
	if body.HasPhysical() {
		p.Density = body.Physical.Density
		p.Gravity = body.Physical.Gravity
	}
	if body.HasAtmosphere() {
		atm := body.Atmosphere
		p.Atmosphere = &Class4PAtmosphere{
			Code:                  atm.Code,
			Pressure:              atm.Pressure,
			OxygenPartialPressure: atm.OxygenPartialPressure,
			ScaleHeight:           atm.ScaleHeight,
			ProfileShorthand:      FormatAtmoProfileShorthand(*atm, atm.Profile),
		}
	}
	if body.HasHydrographics() {
		hydro := body.Hydrographics
		p.Hydrographics = &Class4PHydrographics{
			Code:    hydro.Code,
			Percent: hydro.Percent,
			Profile: hydro.Profile,
		}
	}
	if body.HasDayLength() {
		p.SiderealHours = body.DayLength.SiderealHours
		p.SolarHours = body.DayLength.SolarHours
		p.SolarDaysPerYear = body.DayLength.YearDays
	}
	if body.HasAxialTilt() {
		p.AxialTiltDeg = body.AxialTilt.Degrees
	}
	p.TidalLockRatio = "no"
	if body.HasTidalLock() && body.TidalLock.LockRatio != "" {
		p.TidalLockRatio = body.TidalLock.LockRatio
	}
	if body.HasTidalEffects() {
		p.TidesMeters = body.TidalEffects.Total
	}
	if body.HasTemperature() {
		t := body.Temperature
		lowK := t.LowK
		if lowK <= 0 && t.MeanK > 0 {
			lowK = -1 // sentinel: render as "—"
		}
		p.Temperature = &Class4PTemperature{
			HighK:            t.HighK,
			MeanK:            t.MeanK,
			LowK:             lowK,
			Luminosity:       t.Luminosity,
			Albedo:           t.Albedo,
			GreenhouseFactor: t.GreenhouseFactor,
		}
	}
	if body.HasGeology() {
		g := body.Geology
		p.Seismic = &Class4PSeismic{
			TotalSeismicStress:    g.TotalSeismicStress,
			ResidualSeismicStress: g.ResidualSeismicStress,
			TidalStressFactor:     g.TidalStressFactor,
			TidalHeatingFactor:    g.TidalHeatingFactor,
			TectonicPlates:        g.TectonicPlates,
		}
	}
	if body.HasBiology() {
		bio := body.Biology
		p.Life = &Class4PLife{
			Biomass:        bio.Biomass,
			Biocomplexity:  bio.Biocomplexity,
			HasSophont:     bio.HasNativeSophont,
			HadExtinct:     bio.HadExtinctSophont,
			Biodiversity:   bio.Biodiversity,
			Compatibility:  bio.Compatibility,
			ResourceRating: bio.ResourceRating,
		}
	}
	if body.HasHabitability() {
		p.HabitabilityRating = body.Habitability.Rating
		p.HabitabilityNotes = body.Habitability.Notes
	}
	p.Subordinates = make([]Class4PSubordinate, 0, len(body.Children))
	for _, child := range body.Children {
		p.Subordinates = append(p.Subordinates, Class4PSubordinate{
			Designation:  child.Designation,
			SizeCode:     child.SizeCode,
			DiameterKm:   child.DiameterKm,
			OrbitKm:      int(child.OrbitKm),
			Eccentricity: child.Eccentricity,
			PeriodHours:  child.PeriodHours,
		})
	}
	return p
}

// RenderBody emits the planet/moon mainworld Markdown body for the
// Class IV-P form.
func (p *Class4PPartP) RenderBody(b *strings.Builder, h iiss.FormHeader) {
	fmt.Fprintf(b, "**WORLD:** %s\n", p.Designation)
	fmt.Fprintf(b, "**SECTOR | LOCATION:** %s | %s\n", h.Sector, h.Location)
	fmt.Fprintf(b, "**INITIAL SURVEY:** %s   **LAST UPDATED:** %s\n", h.InitialSurvey, h.LastUpdated)
	fmt.Fprintf(b, "**SYSTEM AGE (Gyr):** %.3f\n\n", p.SystemAgeGyr)

	b.WriteString("### ORBIT\n")
	if p.ParentDesignation != "" {
		fmt.Fprintf(b, "- AU: %.2f (via %s), Eccentricity: %.2f, Period (h): %.2f\n",
			p.AU, p.ParentDesignation, p.Eccentricity, p.PeriodHours)
		fmt.Fprintf(b, "- Moon orbit (km): %.0f around %s\n\n", p.MoonOrbitKm, p.ParentDesignation)
	} else {
		fmt.Fprintf(b, "- AU: %.2f, Eccentricity: %.2f, Period (h): %.2f\n\n",
			p.AU, p.Eccentricity, p.PeriodHours)
	}

	b.WriteString("### SIZE\n")
	fmt.Fprintf(b, "- Diameter (km): %.0f, Density: %.2f, Gravity: %.2f, Mass (Earth): %.3f\n\n",
		p.DiameterKm, p.Density, p.Gravity, p.MassEarth)

	b.WriteString("### ATMOSPHERE\n")
	if p.Atmosphere == nil {
		b.WriteString("- (none — vacuum)\n\n")
	} else {
		atm := p.Atmosphere
		fmt.Fprintf(b, "- Code: %d, Pressure (bar): %.3f, O₂ (bar): %.3f, Scale Height: %.2f\n",
			atm.Code, atm.Pressure, atm.OxygenPartialPressure, atm.ScaleHeight)
		if atm.ProfileShorthand != "" {
			fmt.Fprintf(b, "- Profile: %s\n", atm.ProfileShorthand)
		}
		b.WriteString("\n")
	}

	b.WriteString("### HYDROGRAPHICS\n")
	if p.Hydrographics == nil {
		b.WriteString("- (none)\n\n")
	} else {
		fmt.Fprintf(b, "- Code: %d, Coverage (%%): %d, Profile: %s\n\n",
			p.Hydrographics.Code, p.Hydrographics.Percent, p.Hydrographics.Profile)
	}

	b.WriteString("### ROTATION\n")
	fmt.Fprintf(b, "- Sidereal (h): %.2f, Solar (h): %.2f, Solar days/year: %.2f\n",
		p.SiderealHours, p.SolarHours, p.SolarDaysPerYear)
	fmt.Fprintf(b, "- Axial Tilt: %.2f°\n", p.AxialTiltDeg)
	fmt.Fprintf(b, "- Tidal lock: %s, Tides (m): %.2f\n\n", p.TidalLockRatio, p.TidesMeters)

	b.WriteString("### TEMPERATURE\n")
	if p.Temperature == nil {
		b.WriteString("- (not computed)\n\n")
	} else {
		t := p.Temperature
		low := fmt.Sprintf("%.1f", t.LowK)
		if t.LowK < 0 {
			low = "—"
		}
		fmt.Fprintf(b, "- High (K): %.1f, Mean (K): %.1f, Low (K): %s\n",
			t.HighK, t.MeanK, low)
		fmt.Fprintf(b, "- Luminosity: %.3f, Albedo: %.2f, Greenhouse: %.2f\n\n",
			t.Luminosity, t.Albedo, t.GreenhouseFactor)
	}

	b.WriteString("### SEISMIC\n")
	if p.Seismic == nil {
		b.WriteString("- (not computed)\n\n")
	} else {
		s := p.Seismic
		fmt.Fprintf(b, "- TSS: %d, Residual: %d, Tidal Stress: %d, Tidal Heating: %d, Plates: %d\n\n",
			s.TotalSeismicStress, s.ResidualSeismicStress, s.TidalStressFactor, s.TidalHeatingFactor, s.TectonicPlates)
	}

	b.WriteString("### LIFE\n")
	if p.Life == nil {
		b.WriteString("- (not computed)\n\n")
	} else {
		l := p.Life
		soph := "no"
		if l.HasSophont {
			soph = "yes"
		} else if l.HadExtinct {
			soph = "extinct"
		}
		fmt.Fprintf(b, "- Biomass: %d, Biocomplexity: %d, Sophonts: %s, Biodiversity: %d, Compatibility: %d\n\n",
			l.Biomass, l.Biocomplexity, soph, l.Biodiversity, l.Compatibility)
		fmt.Fprintf(b, "### RESOURCES\n- Rating: %d\n\n", l.ResourceRating)
	}

	b.WriteString("### HABITABILITY\n")
	fmt.Fprintf(b, "- Rating: %d\n", p.HabitabilityRating)
	if p.HabitabilityNotes != "" {
		fmt.Fprintf(b, "- Notes: %s\n", p.HabitabilityNotes)
	}
	b.WriteString("\n")

	if len(p.Subordinates) > 0 {
		b.WriteString("### SUBORDINATES\n")
		b.WriteString("| Designation | Size | Diameter (km) | Orbit (km) | Ecc | Period (h) |\n")
		b.WriteString("| ----------- | ---- | ------------- | ---------- | --- | ---------- |\n")
		for _, s := range p.Subordinates {
			fmt.Fprintf(b, "| %s | %s | %.0f | %d | %.3f | %.2f |\n",
				s.Designation, s.SizeCode, s.DiameterKm, s.OrbitKm, s.Eccentricity, s.PeriodHours)
		}
		b.WriteString("\n")
	}

	if p.IsMainworld {
		b.WriteString("### COMMENTS\n- This is the system mainworld.\n\n")
	}
}

// buildClass4PBelt builds the belt-mainworld view. The Markdown
// renderer is the RenderBody method on the result.
func buildClass4PBelt(u *Universe, body *Body) *Class4PPartPB {
	pb := &Class4PPartPB{
		Designation:  body.Designation,
		PrimaryGroup: body.Group.Designation,
		SystemAgeGyr: u.System.Primary.AgeGyr,
		OrbitNumber:  body.Orbit,
		AU:           stars.OrbitToAU(body.Orbit),
		PeriodHours:  body.Period.Hours,
		IsMainworld:  true,
	}
	if body.HasBelt() {
		pb.SpanOrbits = body.Belt.Span
		pb.MTypePct = body.Belt.Composition.MTypePct
		pb.STypePct = body.Belt.Composition.STypePct
		pb.CTypePct = body.Belt.Composition.CTypePct
		pb.OtherPct = body.Belt.Composition.OtherPct
		pb.Bulk = body.Belt.Bulk
		pb.SigSize1Bodies = body.Belt.SigSize1Bodies
		pb.SigSizeSBodies = body.Belt.SigSizeSBodies
		pb.ResourceRating = body.Belt.ResourceRating
	}
	return pb
}

// RenderBody emits the belt-mainworld Markdown body for the
// Class IV-P form.
func (pb *Class4PPartPB) RenderBody(b *strings.Builder, h iiss.FormHeader) {
	fmt.Fprintf(b, "**WORLD:** %s   **SAH/UWP:** 000\n", pb.Designation)
	fmt.Fprintf(b, "**SECTOR | LOCATION:** %s | %s\n", h.Sector, h.Location)
	fmt.Fprintf(b, "**INITIAL SURVEY:** %s   **LAST UPDATED:** %s\n", h.InitialSurvey, h.LastUpdated)
	fmt.Fprintf(b, "**PRIMARY OBJECT(S):** %s   **SYSTEM AGE (Gyr):** %.3f\n\n",
		pb.PrimaryGroup, pb.SystemAgeGyr)

	b.WriteString("### ORBIT\n")
	fmt.Fprintf(b, "- O#: %.2f, AU: %.2f, Span: %.3f Orbit#s, Period (h): %.2f\n\n",
		pb.OrbitNumber, pb.AU, pb.SpanOrbits, pb.PeriodHours)

	b.WriteString("### COMPOSITION\n")
	fmt.Fprintf(b, "- m-type%%: %d, s-type%%: %d, c-type%%: %d, other%%: %d\n",
		pb.MTypePct, pb.STypePct, pb.CTypePct, pb.OtherPct)
	fmt.Fprintf(b, "- Bulk: %d\n\n", pb.Bulk)

	b.WriteString("### RESOURCES\n")
	fmt.Fprintf(b, "- Rating: %d\n\n", pb.ResourceRating)

	// WBH pp.72-74 produces significant-body COUNTS by size class only;
	// the belt procedure never individuates members, so counts are the
	// complete book-faithful content of this section (not a stub).
	b.WriteString("### MAJOR BODIES\n")
	fmt.Fprintf(b, "- Size 1: %d, Size S: %d\n\n",
		pb.SigSize1Bodies, pb.SigSizeSBodies)

	if pb.IsMainworld {
		b.WriteString("### COMMENTS\n- This is the system mainworld.\n\n")
	}
}
