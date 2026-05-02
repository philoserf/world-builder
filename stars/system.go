package stars

// CompanionStar is a non-primary star with its orbital placement.
//
// Plan 2 P2-3 defines the type; later tasks (P2-4 through P2-10) populate
// the orbital fields. ParentIndex is -1 when the parent is the primary,
// or an index into System.Companions otherwise (used to encode that, e.g.,
// a Far-orbit star's own companion has the Far star as its parent).
type CompanionStar struct {
	Star         Star
	Designation  string
	OrbitClass   OrbitClass
	OrbitNumber  float64
	AU           float64
	Eccentricity float64
	Inclination  float64 // degrees
	PeriodYears  float64
	ParentIndex  int
}

// System is a star system with a primary plus zero or more companions.
//
// PrimaryDesignation is set by AssignDesignations and is "A" for a single
// primary or "Aa" if the primary has its own OrbitCompanion-class child.
type System struct {
	Primary            Star
	PrimaryDesignation string
	Companions         []CompanionStar
	AgeGyr             float64
}
