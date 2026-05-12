package iiss

import "strings"

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

// Class4PForm is the IISS Class IV-P "Planetary Detail" Survey form,
// rendered only for the auto-picked mainworld. PartP/PartPB carry the
// JSON payload (concretely *worlds.Class4PPartP / *worlds.Class4PPartPB,
// typed any to avoid an iiss→worlds import cycle); RenderBody is the
// Markdown renderer.
type Class4PForm struct {
	FormHeader
	Variant    Class4PVariant
	PartP      any                                `json:",omitempty"`
	PartPB     any                                `json:",omitempty"`
	RenderBody func(*strings.Builder, FormHeader) `json:"-"`
}

// SystemForms aggregates the three IISS forms for a generated system,
// plus the system-wide profile strings, the auto-picked mainworld
// designation, and a Markdown referee-facing Notable Features summary.
// Renderer functions take SystemForms (or one of its fields) so iiss/
// does not import worlds/.
type SystemForms struct {
	Class0I              Class0IForm
	Class23              Class23Form
	Class4P              Class4PForm
	MainworldDesignation string
	ShortProfile         string
	LongProfile          string
	NotableFeatures      string // pre-rendered Markdown block; rendered above Class 0/I
}
