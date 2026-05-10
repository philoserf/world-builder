package iiss

// Class0IForm is the IISS Class 0/I Survey form (WBH p.35 layout).
// Renders the stellar census.
type Class0IForm struct {
	FormHeader
	SystemAgeGyr float64
	StellarCount int
	Stars        []Class0IStarRow
}

// Class0IStarRow is one row of the Class 0/I Stars table.
type Class0IStarRow struct {
	Component   string
	Class       string
	Mass        float64
	Diameter    float64
	Temperature float64
	Luminosity  float64
	HZCO        float64
	MAO         float64
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
	ClassIII     bool
}

// Class23Counts holds the per-system world counts shown on the form.
type Class23Counts struct {
	GasGiants      int
	PlanetoidBelts int
	Terrestrials   int
	Total          int
}

// Class23Object is one row of the Class II/III Objects table. Fields
// land cycle-by-cycle as renderer detail is implemented.
type Class23Object struct {
	Designation string
	Notes       string
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

// Class4PPartP holds the planet / moon mainworld detail. Fields land as
// renderer work proceeds.
type Class4PPartP struct{}

// Class4PPartPB holds the belt mainworld detail. Per docs/pass-2/
// api-surface.md § Open questions, decided, this is an empty shell —
// fleshed out post-parity when the belt-mainworld fixture lands.
type Class4PPartPB struct{}

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
