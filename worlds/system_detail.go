package worlds

import "wbh/stars"

// DetailedPlacement extends 2B's Placement with the WBH pp. 53-67
// per-body data (Size, moons, period, HZ flag, designation).
//
// Embeds Placement, continuing the existing chain:
//
//	Slot → AnomalousSlot → Placement → DetailedPlacement
//
// 2B types are unchanged.
type DetailedPlacement struct {
	Placement // 2B fields: Body, PrefixRoll, Eccentricity, AnomalousSlot, Slot

	// Terrestrial fields — set when Body == BodyTerrestrial.
	SizeCode   SizeCode
	DiameterKm float64

	// Gas-giant fields — set when Body == BodyGasGiant.
	GGClass        GasGiantClass
	GGDiameterCode string
	DiameterEarth  float64
	MassEarth      float64

	// All non-empty bodies:
	Designation string // "Aab I", "Aab PI" — assigned by AssignPlanetDesignations
	Period      Period
	HZ          bool // within HZCO ± 1.0 — set by MarkHZ
	Moons       []Moon
}

// SystemDetail is the DetailSystem façade output, layered atop 2B's
// SystemPlacement.
type SystemDetail struct {
	SystemPlacement // 2B: Counts, Allocations, BaselineN, BaselineOrbit, EmptyOrbits, SystemSpread, Placements

	// Detailed mirrors SystemPlacement.Placements 1:1, with 2C per-body
	// detail attached. Ordered by ascending orbit within each group,
	// matching SystemPlacement.Placements (which itself follows
	// PlaceOrbitSlots' ascending-orbit walk). LongProfile and
	// AssignPlanetDesignations both rely on this ordering — the T14
	// DetailSystem façade must preserve it when building Detailed.
	Detailed []DetailedPlacement

	ShortProfile string          // "G-P-T-N-S" form per WBH p.58
	LongProfile  string          // "St-N-W-W-S:..." form per WBH p.58
	Survey       IISSClass23Form // IISS Class II/III survey form (Task 13)
}

// IISSClass23Form is forward-declared here so SystemDetail.Survey can
// reference it; the full type lands in Task 13 (worlds/survey_form.go).
//
// Until Task 13 lands, this is a thin embedding of stars.SurveyForm.
// T13 will REPLACE this declaration with the full Class II/III form
// type (the new type definition will live in survey_form.go and this
// placeholder will be removed).
type IISSClass23Form struct {
	stars.SurveyForm // embedded for header + Stars table; see Task 13
}
