package iiss

import (
	"fmt"
	"strings"
)

// Class4PPartP is the planet/moon mainworld view of the IISS Class IV-P
// form, populated by worlds.BuildIISSForms and emitted as JSON. Shape
// matches the WBH p.138 Form 0407F-IV PART P layout.
type Class4PPartP struct {
	Designation  string
	SystemAgeGyr float64

	OrbitNumber  float64
	AU           float64
	Eccentricity float64
	PeriodHours  float64

	// Moon-mainworld extras (mainworld is a moon): the moon's own orbit
	// around its parent. Zero/empty for planet mainworlds.
	MoonOrbitKm       float64 `json:",omitempty"`
	ParentDesignation string  `json:",omitempty"`

	// Ring reports the WBH ring outcome (p.55 / p.76); RingCentrePD /
	// RingSpanPD carry its centre and span in planet-diameters (p.77).
	Ring         bool    `json:",omitempty"`
	RingCentrePD float64 `json:",omitempty"`
	RingSpanPD   float64 `json:",omitempty"`

	Composition    string `json:",omitempty"`
	DiameterKm     float64
	Density        float64
	Gravity        float64
	MassEarth      float64
	EscapeVelocity float64 `json:",omitempty"` // m/s
	SizeProfile    string  `json:",omitempty"` // Size-Diameter-Density-Gravity-Mass (WBH p.71)

	// Gas-giant view. When IsGasGiant is true the terrestrial atmosphere/
	// hydrographics/temperature/seismic/life/habitability sections are N/A
	// and these carry the GG-specific detail instead.
	IsGasGiant    bool    `json:",omitempty"`
	GasGiantClass string  `json:",omitempty"` // "Small" | "Medium" | "Large"
	DiameterEarth float64 `json:",omitempty"` // Terra diameters
	ResidualTempK float64 `json:",omitempty"` // WBH p.125 gas-giant residual heat

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
	Subtype               string `json:",omitempty"`
	Pressure              float64
	OxygenPartialPressure float64
	ScaleHeight           float64
	ProfileShorthand      string
	Taints                []Class4PTaint `json:",omitempty"`
	Hazards               []string       `json:",omitempty"` // Insidious atmosphere hazard codes
}

// Class4PTaint is one atmosphere taint (WBH p.83): code, severity,
// persistence.
type Class4PTaint struct {
	Code        string
	Severity    int
	Persistence int
}

// Class4PHydrographics captures the WBH p.138 HYDROGRAPHICS block.
type Class4PHydrographics struct {
	Code    int
	Percent int
	Profile string
	// Surface-feature distribution (WBH p.99), present when computed.
	Distribution string `json:",omitempty"` // "Extremely Concentrated" | … | "Extremely Dispersed"
	Geography    string `json:",omitempty"` // "Ocean" | "Land"
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
// (a planet mainworld's moons). SizeCode is the eHex size code string.
type Class4PSubordinate struct {
	Designation  string
	SizeCode     string
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

// RenderBody emits the planet/moon mainworld Markdown body for the
// Class IV-P form. Called by markdownClass4Part.
func (p *Class4PPartP) RenderBody(b *strings.Builder, h FormHeader) {
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
	if p.IsGasGiant {
		fmt.Fprintf(b, "- Class: %s gas giant\n", p.GasGiantClass)
		fmt.Fprintf(b, "- Diameter (km): %.0f (%.2f × Terra), Mass (Earth): %.3f\n",
			p.DiameterKm, p.DiameterEarth, p.MassEarth)
	} else {
		if p.Composition != "" {
			fmt.Fprintf(b, "- Composition: %s\n", p.Composition)
		}
		fmt.Fprintf(b, "- Diameter (km): %.0f, Density: %.2f, Gravity: %.2f, Mass (Earth): %.3f\n",
			p.DiameterKm, p.Density, p.Gravity, p.MassEarth)
		if p.EscapeVelocity > 0 {
			fmt.Fprintf(b, "- Escape velocity (m/s): %.0f\n", p.EscapeVelocity)
		}
		if p.SizeProfile != "" {
			fmt.Fprintf(b, "- Size profile: `%s`\n", p.SizeProfile)
		}
	}
	b.WriteString("\n")

	if !p.IsGasGiant {
		b.WriteString("### ATMOSPHERE\n")
		if p.Atmosphere == nil {
			b.WriteString("- (none — vacuum)\n\n")
		} else {
			atm := p.Atmosphere
			fmt.Fprintf(b, "- Code: %d, Pressure (bar): %.3f, O₂ (bar): %.3f, Scale Height: %.2f\n",
				atm.Code, atm.Pressure, atm.OxygenPartialPressure, atm.ScaleHeight)
			if atm.Subtype != "" {
				fmt.Fprintf(b, "- Subtype: %s\n", atm.Subtype)
			}
			if atm.ProfileShorthand != "" {
				fmt.Fprintf(b, "- Profile: %s\n", atm.ProfileShorthand)
			}
			for _, t := range atm.Taints {
				fmt.Fprintf(b, "- Taint: %s (severity %d, persistence %d)\n", t.Code, t.Severity, t.Persistence)
			}
			if len(atm.Hazards) > 0 {
				fmt.Fprintf(b, "- Insidious hazards: %s\n", strings.Join(atm.Hazards, ", "))
			}
			b.WriteString("\n")
		}

		b.WriteString("### HYDROGRAPHICS\n")
		if p.Hydrographics == nil {
			b.WriteString("- (none)\n\n")
		} else {
			h := p.Hydrographics
			fmt.Fprintf(b, "- Code: %d, Coverage (%%): %d, Profile: %s\n", h.Code, h.Percent, h.Profile)
			if h.Distribution != "" {
				fmt.Fprintf(b, "- Surface distribution: %s (%s)\n", h.Distribution, h.Geography)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("### ROTATION\n")
	fmt.Fprintf(b, "- Sidereal (h): %.2f, Solar (h): %.2f, Solar days/year: %.2f\n",
		p.SiderealHours, p.SolarHours, p.SolarDaysPerYear)
	fmt.Fprintf(b, "- Axial Tilt: %.2f°\n", p.AxialTiltDeg)
	fmt.Fprintf(b, "- Tidal lock: %s, Tides (m): %.2f\n\n", p.TidalLockRatio, p.TidesMeters)

	if p.IsGasGiant {
		b.WriteString("### GAS GIANT\n")
		if p.ResidualTempK > 0 {
			fmt.Fprintf(b, "- Residual temperature (K): %.1f (WBH p.125)\n", p.ResidualTempK)
		} else {
			b.WriteString("- Residual temperature (K): — (below 1 K, WBH p.125)\n")
		}
		b.WriteString("- No discrete surface: atmosphere, hydrographics, life, and habitability do not apply.\n\n")
	} else {
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
	}

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

	if p.IsMainworld || p.Ring {
		b.WriteString("### COMMENTS\n")
		if p.IsMainworld {
			b.WriteString("- This is the system mainworld.\n")
		}
		if p.Ring {
			if p.RingSpanPD > 0 {
				fmt.Fprintf(b, "- Has a planetary ring — R01:%.2f-%.2f (centre %.2f PD, span %.2f PD, WBH p.77).\n",
					p.RingCentrePD, p.RingSpanPD, p.RingCentrePD, p.RingSpanPD)
			} else {
				b.WriteString("- Has a planetary ring (WBH p.55/p.76).\n")
			}
		}
		b.WriteString("\n")
	}
}

// RenderBody emits the belt-mainworld Markdown body for the Class IV-P
// form. Called by markdownClass4Part.
func (pb *Class4PPartPB) RenderBody(b *strings.Builder, h FormHeader) {
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
