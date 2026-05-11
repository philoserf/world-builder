package iiss

// Class0IForm is the IISS Class 0/I Survey form (WBH p.35 layout).
// Renders the stellar census.
type Class0IForm struct {
	FormHeader
	SystemAgeGyr float64
	StellarCount int
	Stars        []Class0IStarRow
}

// Class0IStarRow is one row of the Class 0/I Stars table. Mirrors
// pass-1's stars.SurveyComponent so the form can render with full
// companion-orbit fidelity.
type Class0IStarRow struct {
	Component    string
	Class        string // "G7 V" / "D" / "—" for composites
	Mass         float64
	Diameter     float64 // 0 for composites
	Temperature  float64 // 0 for composites
	Luminosity   float64
	Orbit        float64
	AU           float64
	Eccentricity float64
	PeriodYears  float64
	HZCO         float64 // populated only on rows that act as a single HZCO source
	MAO          float64 // populated for Class II/III; 0 for Class 0/I header
}

// Class23Form is the IISS Class II/III Survey form (WBH form 0421D-II.III,
// pp.60-67). Extends Class 0/I with per-body counts and the Objects table.
type Class23Form struct {
	FormHeader
	SystemAgeGyr float64
	StellarCount int
	Stars        []Class0IStarRow
	Counts       Class23Counts
	Objects      []Class23Object
}

// Class23Counts holds the per-system world counts shown on the form.
type Class23Counts struct {
	GasGiants      int
	PlanetoidBelts int
	Terrestrials   int
	Total          int
}

// Class23Object is one row of the WBH p.61 Class II/III Objects
// table. Fields mirror pass-1's worlds.ObjectRow so the renderer
// reproduces the book layout.
type Class23Object struct {
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

// Class4PVariant identifies which Class IV-P variant applies to the
// auto-picked mainworld.
type Class4PVariant int

const (
	// Class4PPlanet — mainworld is a planet.
	Class4PPlanet Class4PVariant = iota
	// Class4PMoon — mainworld is a moon.
	Class4PMoon
	// Class4PBelt — mainworld is a belt.
	Class4PBelt
)

// Class4PForm is the IISS Class IV-P "Planetary Detail" Survey form.
// Renders only for the auto-picked mainworld; PartP populated for
// Planet/Moon variants, PartPB for Belt variant.
type Class4PForm struct {
	FormHeader
	Variant Class4PVariant
	PartP   *Class4PPartP
	PartPB  *Class4PPartPB
}

// Class4PPartP holds the planet / moon mainworld detail per WBH p.138
// Form 0407F-IV PART P.
type Class4PPartP struct {
	Designation  string
	SystemAgeGyr float64

	// Orbit
	OrbitNumber  float64
	AU           float64
	Eccentricity float64
	PeriodHours  float64

	// Size
	DiameterKm float64
	Density    float64
	Gravity    float64
	MassEarth  float64

	// Atmosphere — populated when the body has an atmosphere
	Atmosphere *Class4PAtmosphere

	// Hydrographics — populated when the body has hydrographics
	Hydrographics *Class4PHydrographics

	// Rotation
	SiderealHours    float64
	SolarHours       float64
	SolarDaysPerYear float64
	AxialTiltDeg     float64
	TidalLockRatio   string
	TidesMeters      float64

	// Temperature
	Temperature *Class4PTemperature

	// Seismic / Geology
	Seismic *Class4PSeismic

	// Life
	Life *Class4PLife

	// Habitability
	HabitabilityRating int
	HabitabilityNotes  string

	// Subordinates (moons of a planet mainworld; not used for moon mainworlds)
	Subordinates []Class4PSubordinate

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
type Class4PTemperature struct {
	HighK            float64
	MeanK            float64
	LowK             float64 // -1 sentinel = "—" (degenerate-model)
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
	SizeCode     string
	DiameterKm   float64
	OrbitKm      int
	Eccentricity float64
	PeriodHours  float64
}

// Class4PPartPB holds the belt mainworld detail per WBH p.139
// FORM 0407K-IV PART P.B.
type Class4PPartPB struct {
	Designation  string
	PrimaryGroup string
	SystemAgeGyr float64

	// Orbit
	OrbitNumber float64
	AU          float64
	SpanOrbits  float64
	PeriodHours float64

	// Composition
	MTypePct       int
	STypePct       int
	CTypePct       int
	OtherPct       int
	Bulk           int
	SigSize1Bodies int
	SigSizeSBodies int

	// Resources
	ResourceRating int

	IsMainworld bool
}

// SystemForms aggregates the three IISS forms for a generated system,
// plus the system-wide profile strings and the auto-picked mainworld
// designation. Renderer functions take SystemForms (or one of its
// fields) so iiss/ does not import worlds/.
type SystemForms struct {
	Class0I              Class0IForm
	Class23              Class23Form
	Class4P              Class4PForm
	MainworldDesignation string
	ShortProfile         string
	LongProfile          string
}
