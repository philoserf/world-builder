package stars

// StarTypeRow is one row of the WBH p. 15 Star Type Determination table.
//
// Default column is Type; rolls of 2 redirect to Special, rolls of 12
// redirect to Hot. Class III+ rolls go to Giants. The Unusual and
// Peculiar columns are entered only when the procedure says so.
type StarTypeRow struct {
	Type, Hot, Special, Unusual, Giants, Peculiar string
}

// StarTypeDetermination is the WBH p. 15 Star Type Determination table.
var StarTypeDetermination = map[int]StarTypeRow{
	2:  {Type: "Special", Hot: "A", Special: "Class VI", Unusual: "Peculiar", Giants: "Class III", Peculiar: "Black Hole"},
	3:  {Type: "M", Hot: "A", Special: "Class VI", Unusual: "Class VI", Giants: "Class III", Peculiar: "Pulsar"},
	4:  {Type: "M", Hot: "A", Special: "Class VI", Unusual: "Class IV", Giants: "Class III", Peculiar: "Neutron Star"},
	5:  {Type: "M", Hot: "A", Special: "Class VI", Unusual: "BD", Giants: "Class III", Peculiar: "Nebula"},
	6:  {Type: "M", Hot: "A", Special: "Class IV", Unusual: "BD", Giants: "Class III", Peculiar: "Nebula"},
	7:  {Type: "K", Hot: "A", Special: "Class IV", Unusual: "BD", Giants: "Class III", Peculiar: "Protostar"},
	8:  {Type: "K", Hot: "A", Special: "Class IV", Unusual: "D", Giants: "Class III", Peculiar: "Protostar"},
	9:  {Type: "G", Hot: "A", Special: "Class III", Unusual: "D", Giants: "Class II", Peculiar: "Protostar"},
	10: {Type: "G", Hot: "B", Special: "Class III", Unusual: "D", Giants: "Class II", Peculiar: "Star Cluster"},
	11: {Type: "F", Hot: "B", Special: "Giants", Unusual: "Class III", Giants: "Class Ib", Peculiar: "Anomaly"},
	12: {Type: "Hot", Hot: "O", Special: "Giants", Unusual: "Giants", Giants: "Class Ia", Peculiar: "Anomaly"},
}

// StarSubtypeNumeric is the WBH p. 16 Star Subtype table — Numeric column.
var StarSubtypeNumeric = map[int]int{
	2: 0, 3: 1, 4: 3, 5: 5, 6: 7, 7: 9, 8: 8, 9: 6, 10: 4, 11: 2, 12: 0,
}

// StarSubtypeMType is the WBH p. 16 Star Subtype table — M-type column (primary only).
var StarSubtypeMType = map[int]int{
	2: 8, 3: 6, 4: 5, 5: 4, 6: 0, 7: 2, 8: 1, 9: 3, 10: 5, 11: 7, 12: 9,
}

// ClassRow holds class-keyed values for tables shaped like the Mass,
// Diameter, and Luminosity tables on WBH pp. 17, 19. A nil pointer
// indicates the book leaves the cell blank ("—").
type ClassRow struct {
	Ia, Ib, II, III, IV, V, VI *float64
}

// Get returns the value for the given luminosity class, or false if
// the cell is absent.
func (r ClassRow) Get(lc LuminosityClass) (float64, bool) {
	var p *float64
	switch lc {
	case Ia:
		p = r.Ia
	case Ib:
		p = r.Ib
	case II:
		p = r.II
	case III:
		p = r.III
	case IV:
		p = r.IV
	case V:
		p = r.V
	case VI:
		p = r.VI
	}
	if p == nil {
		return 0, false
	}
	return *p, true
}

// StarMass is the WBH p. 17 Star Mass and Temperature by Class table — Mass column.
// Values are in solar masses (Sol = 1.0).
var StarMass = map[string]ClassRow{
	"O0": {Ia: new(200.0), Ib: new(150.0), II: new(130.0), III: new(110.0), V: new(90.0), VI: new(2.0)},
	"O5": {Ia: new(80.0), Ib: new(60.0), II: new(40.0), III: new(30.0), V: new(60.0), VI: new(1.5)},
	"B0": {Ia: new(60.0), Ib: new(40.0), II: new(30.0), III: new(20.0), IV: new(20.0), V: new(18.0), VI: new(0.5)},
	"B5": {Ia: new(30.0), Ib: new(25.0), II: new(20.0), III: new(10.0), IV: new(10.0), V: new(5.0), VI: new(0.4)},
	"A0": {Ia: new(20.0), Ib: new(15.0), II: new(14.0), III: new(8.0), IV: new(4.0), V: new(2.2)},
	"A5": {Ia: new(15.0), Ib: new(13.0), II: new(11.0), III: new(6.0), IV: new(2.3), V: new(1.5)},
	"F0": {Ia: new(13.0), Ib: new(12.0), II: new(10.0), III: new(4.0), IV: new(2.0), V: new(1.5)},
	"F5": {Ia: new(12.0), Ib: new(10.0), II: new(8.0), III: new(3.0), IV: new(1.5), V: new(1.3)},
	"G0": {Ia: new(12.0), Ib: new(10.0), II: new(8.0), III: new(2.5), IV: new(1.7), V: new(1.1), VI: new(0.8)},
	"G5": {Ia: new(13.0), Ib: new(11.0), II: new(10.0), III: new(2.4), IV: new(1.2), V: new(0.9), VI: new(0.7)},
	"K0": {Ia: new(14.0), Ib: new(12.0), II: new(10.0), III: new(1.1), IV: new(1.5), V: new(0.8), VI: new(0.6)},
	"K5": {Ia: new(18.0), Ib: new(13.0), II: new(12.0), III: new(1.5), V: new(0.7), VI: new(0.5)},
	"M0": {Ia: new(20.0), Ib: new(15.0), II: new(14.0), III: new(1.8), V: new(0.5), VI: new(0.4)},
	"M5": {Ia: new(25.0), Ib: new(20.0), II: new(16.0), III: new(2.4), V: new(0.16), VI: new(0.12)},
	"M9": {Ia: new(30.0), Ib: new(25.0), II: new(18.0), III: new(8.0), V: new(0.08), VI: new(0.075)},
}

// StarTemperature is the WBH p. 17 Star Mass and Temperature by Class table —
// Temperature column. Values are in Kelvin.
var StarTemperature = map[string]float64{
	"O0": 50000, "O5": 40000, "B0": 30000, "B5": 15000,
	"A0": 10000, "A5": 8000,
	"F0": 7500, "F5": 6500,
	"G0": 6000, "G5": 5600,
	"K0": 5200, "K5": 4400,
	"M0": 3700, "M5": 3000, "M9": 2400,
}

// StarDiameter is the WBH p. 19 Star Diameter by Class table.
// Values are in solar diameters (Sol = 1.0).
var StarDiameter = map[string]ClassRow{
	"O0": {Ia: new(25.0), Ib: new(24.0), II: new(22.0), III: new(21.0), V: new(20.0), VI: new(0.18)},
	"O5": {Ia: new(22.0), Ib: new(20.0), II: new(18.0), III: new(15.0), V: new(12.0), VI: new(0.18)},
	"B0": {Ia: new(20.0), Ib: new(14.0), II: new(12.0), III: new(10.0), IV: new(8.0), V: new(7.0), VI: new(0.2)},
	"B5": {Ia: new(60.0), Ib: new(25.0), II: new(14.0), III: new(6.0), IV: new(5.0), V: new(3.5), VI: new(0.5)},
	"A0": {Ia: new(120.0), Ib: new(50.0), II: new(30.0), III: new(5.0), IV: new(4.0), V: new(2.2)},
	"A5": {Ia: new(180.0), Ib: new(75.0), II: new(45.0), III: new(5.0), IV: new(3.0), V: new(2.0)},
	"F0": {Ia: new(210.0), Ib: new(85.0), II: new(50.0), III: new(5.0), IV: new(3.0), V: new(1.7)},
	"F5": {Ia: new(280.0), Ib: new(115.0), II: new(66.0), III: new(5.0), IV: new(2.0), V: new(1.5)},
	"G0": {Ia: new(330.0), Ib: new(135.0), II: new(77.0), III: new(10.0), IV: new(3.0), V: new(1.1), VI: new(0.8)},
	"G5": {Ia: new(360.0), Ib: new(150.0), II: new(90.0), III: new(15.0), IV: new(4.0), V: new(0.95), VI: new(0.7)},
	"K0": {Ia: new(420.0), Ib: new(180.0), II: new(110.0), III: new(20.0), IV: new(6.0), V: new(0.9), VI: new(0.6)},
	"K5": {Ia: new(600.0), Ib: new(260.0), II: new(160.0), III: new(40.0), V: new(0.8), VI: new(0.5)},
	"M0": {Ia: new(900.0), Ib: new(380.0), II: new(230.0), III: new(60.0), V: new(0.7), VI: new(0.4)},
	"M5": {Ia: new(1200.0), Ib: new(600.0), II: new(350.0), III: new(100.0), V: new(0.2), VI: new(0.1)},
	"M9": {Ia: new(1800.0), Ib: new(800.0), II: new(500.0), III: new(200.0), V: new(0.1), VI: new(0.08)},
}

// StarLuminosity is the WBH p. 19 Star Luminosity by Class table.
// Values are in solar luminosities (Sol = 1.0).
var StarLuminosity = map[string]ClassRow{
	"O0": {Ia: new(3_400_000.0), Ib: new(3_200_000.0), II: new(2_700_000.0), III: new(2_400_000.0), V: new(2_200_000.0), VI: new(180.0)},
	"O5": {Ia: new(1_100_000.0), Ib: new(900_000.0), II: new(730_000.0), III: new(510_000.0), V: new(330_000.0), VI: new(73.0)},
	"B0": {Ia: new(290_000.0), Ib: new(140_000.0), II: new(100_000.0), III: new(72_000.0), IV: new(46_000.0), V: new(35_000.0), VI: new(29.0)},
	"B5": {Ia: new(160_000.0), Ib: new(28_000.0), II: new(8800.0), III: new(1600.0), IV: new(1100.0), V: new(550.0), VI: new(11.0)},
	"A0": {Ia: new(130_000.0), Ib: new(22_000.0), II: new(8000.0), III: new(220.0), IV: new(140.0), V: new(43.0)},
	"A5": {Ia: new(120_000.0), Ib: new(20_000.0), II: new(7300.0), III: new(90.0), IV: new(33.0), V: new(15.0)},
	"F0": {Ia: new(120_000.0), Ib: new(20_000.0), II: new(7000.0), III: new(70.0), IV: new(25.0), V: new(8.1)},
	"F5": {Ia: new(120_000.0), Ib: new(20_000.0), II: new(6900.0), III: new(39.0), IV: new(6.0), V: new(3.5)},
	"G0": {Ia: new(120_000.0), Ib: new(20_000.0), II: new(6800.0), III: new(120.0), IV: new(10.0), V: new(1.4), VI: new(0.73)},
	"G5": {Ia: new(110_000.0), Ib: new(20_000.0), II: new(7000.0), III: new(200.0), IV: new(14.0), V: new(0.78), VI: new(0.43)},
	"K0": {Ia: new(110_000.0), Ib: new(21_000.0), II: new(7800.0), III: new(260.0), IV: new(23.0), V: new(0.52), VI: new(0.23)},
	"K5": {Ia: new(120_000.0), Ib: new(22_000.0), II: new(8400.0), III: new(530.0), V: new(0.21), VI: new(0.083)},
	"M0": {Ia: new(130_000.0), Ib: new(24_000.0), II: new(8800.0), III: new(600.0), V: new(0.082), VI: new(0.027)},
	"M5": {Ia: new(100_000.0), Ib: new(26_000.0), II: new(8800.0), III: new(720.0), V: new(0.0029), VI: new(0.00072)},
	"M9": {Ia: new(90_000.0), Ib: new(19_000.0), II: new(7300.0), III: new(1200.0), V: new(0.00029), VI: new(0.00019)},
}

// MultipleStarsPresenceThreshold is the WBH p.23 2D threshold (after
// DMs) for a star to be present in a given orbit class.
const MultipleStarsPresenceThreshold = 10

// ExistingStarLocationsBinary is the WBH p.24 Existing Star Locations table
// for binary systems, keyed by 1D. Not consulted by GenerateSystem (which
// uses the p.29 Non-Primary determination flow); retained per the
// tables-as-literals convention as a book transcription for library callers.
// "RollAgainOrCompanion" means the Referee may either reroll or treat
// the new star as a companion of an existing star with the same Class
// and Type. "RollAgain" means simply reroll on this table.
var ExistingStarLocationsBinary = map[int]string{
	1: "Companion",
	2: "Close",
	3: "Near",
	4: "Far",
	5: "RollAgainOrCompanion",
	6: "RollAgain",
}

// ExistingStarLocationsTrinaryPlus is the WBH p.24 Existing Star Locations
// table for trinary and larger systems, keyed by 1D. Like
// ExistingStarLocationsBinary, retained as a book transcription; not
// consulted by GenerateSystem.
var ExistingStarLocationsTrinaryPlus = map[int]string{
	1: "Companion",
	2: "Close",
	3: "Near",
	4: "Far",
	5: "RollAgainOrCompanion",
	6: "Far",
}

// NonPrimaryRow is one row of the WBH p.29 Non-Primary Star Determination
// table. Cells are descriptor strings:
//
//	"Random", "Lesser", "Sibling", "Twin", "Other", "D", "BD".
type NonPrimaryRow struct {
	Secondary, Companion, PostStellar, Other string
}

// NonPrimaryStarDetermination is the WBH p.29 Non-Primary Star Determination
// table, keyed by clamped 2D+DM in [2,12].
// Class III/IV primaries apply DM-1 to the 2D before lookup.
var NonPrimaryStarDetermination = map[int]NonPrimaryRow{
	2:  {Secondary: "Other", Companion: "Other", PostStellar: "Other", Other: "D"},
	3:  {Secondary: "Other", Companion: "Other", PostStellar: "Other", Other: "D"},
	4:  {Secondary: "Random", Companion: "Random", PostStellar: "Random", Other: "D"},
	5:  {Secondary: "Random", Companion: "Random", PostStellar: "Random", Other: "D"},
	6:  {Secondary: "Random", Companion: "Lesser", PostStellar: "Random", Other: "D"},
	7:  {Secondary: "Lesser", Companion: "Lesser", PostStellar: "Random", Other: "D"},
	8:  {Secondary: "Lesser", Companion: "Sibling", PostStellar: "Random", Other: "BD"},
	9:  {Secondary: "Sibling", Companion: "Sibling", PostStellar: "Lesser", Other: "BD"},
	10: {Secondary: "Sibling", Companion: "Twin", PostStellar: "Lesser", Other: "BD"},
	11: {Secondary: "Twin", Companion: "Twin", PostStellar: "Twin", Other: "BD"},
	12: {Secondary: "Twin", Companion: "Twin", PostStellar: "Twin", Other: "BD"},
}

// EccentricityRow is one row of the WBH p.27 Eccentricity Values table.
//
// The procedure: roll 2D + DMs, clamp into [5, 12], look up the row,
// then add a second-roll term (rolling SecondRoll dice) divided by Divisor.
type EccentricityRow struct {
	Base       float64
	SecondRoll string  // dice notation for the second roll, e.g. "1D" or "2D"
	Divisor    float64 // divisor applied to the second-roll result
}

// EccentricityValues is the WBH p.27 Eccentricity Values table.
// Rows 6-7 share an entry; rows 8-9 share an entry. Rolls below 5 clamp
// to row 5; rolls 12+ clamp to row 12.
var EccentricityValues = map[int]EccentricityRow{
	5:  {Base: -0.001, SecondRoll: "1D", Divisor: 1000},
	6:  {Base: 0.00, SecondRoll: "1D", Divisor: 200},
	7:  {Base: 0.00, SecondRoll: "1D", Divisor: 200},
	8:  {Base: 0.03, SecondRoll: "1D", Divisor: 100},
	9:  {Base: 0.03, SecondRoll: "1D", Divisor: 100},
	10: {Base: 0.05, SecondRoll: "1D", Divisor: 20},
	11: {Base: 0.05, SecondRoll: "2D", Divisor: 20},
	12: {Base: 0.30, SecondRoll: "2D", Divisor: 20},
}

// ----- P2-8: Orbit# ↔ AU conversion (WBH p.26) -----

// OrbitNumberRow is one row of the WBH p.26 Orbit# table.
type OrbitNumberRow struct {
	DistanceAU   float64
	DifferenceAU float64 // difference to the next-higher Orbit# (0 for Orbit# 20)
	MillionKm    float64
	Example      string
}

// OrbitNumberTable is the WBH p.26 Orbit# table mapping integer Orbit#
// 0..20 to AU distance, difference to the next orbit, kilometers, and
// the book's planetary example (where given).
var OrbitNumberTable = map[int]OrbitNumberRow{
	0:  {DistanceAU: 0, DifferenceAU: 0.4, MillionKm: 0, Example: "Companion Orbit"},
	1:  {DistanceAU: 0.4, DifferenceAU: 0.3, MillionKm: 60, Example: "Mercury"},
	2:  {DistanceAU: 0.7, DifferenceAU: 0.3, MillionKm: 105, Example: "Venus"},
	3:  {DistanceAU: 1.0, DifferenceAU: 0.6, MillionKm: 150, Example: "Terra"},
	4:  {DistanceAU: 1.6, DifferenceAU: 1.2, MillionKm: 240, Example: "Mars"},
	5:  {DistanceAU: 2.8, DifferenceAU: 2.4, MillionKm: 420, Example: "Asteroid Belt (Ceres)"},
	6:  {DistanceAU: 5.2, DifferenceAU: 4.8, MillionKm: 780, Example: "Jupiter"},
	7:  {DistanceAU: 10, DifferenceAU: 10, MillionKm: 1500, Example: "Saturn"},
	8:  {DistanceAU: 20, DifferenceAU: 20, MillionKm: 3000, Example: "Uranus"},
	9:  {DistanceAU: 40, DifferenceAU: 37, MillionKm: 6000, Example: "Kuiper Belt (Pluto)"},
	10: {DistanceAU: 77, DifferenceAU: 77, MillionKm: 11550, Example: "Scattered Disk (Eris)"},
	11: {DistanceAU: 154, DifferenceAU: 154, MillionKm: 23100},
	12: {DistanceAU: 308, DifferenceAU: 307, MillionKm: 46200},
	13: {DistanceAU: 615, DifferenceAU: 615, MillionKm: 92250, Example: "Outer Scattered Disk (Sedna)"},
	14: {DistanceAU: 1230, DifferenceAU: 1270, MillionKm: 184500},
	15: {DistanceAU: 2500, DifferenceAU: 2400, MillionKm: 375000, Example: "Inner Oort Cloud"},
	16: {DistanceAU: 4900, DifferenceAU: 4900, MillionKm: 735000, Example: "Middle Oort Cloud"},
	17: {DistanceAU: 9800, DifferenceAU: 9700, MillionKm: 1470000},
	18: {DistanceAU: 19500, DifferenceAU: 20000, MillionKm: 2925000},
	19: {DistanceAU: 39500, DifferenceAU: 39200, MillionKm: 5925000, Example: "Outer Oort Cloud"},
	20: {DistanceAU: 78700, DifferenceAU: 0, MillionKm: 11805000, Example: "> 1 light-year"},
}

// SpecialObjectAgeRow describes how to age a special-object kind (WBH p.22).
type SpecialObjectAgeRow struct {
	BaseFormula      string // "small_star" | "100m_per_2d10" | "10m_per_2d10"
	AddProgenitorAge bool   // if true, also add (2+D3) × dead-star-mass progenitor age
}

// SpecialObjectAgeByType is the WBH p.22 Special and Unusual Object Age
// by Type table.
var SpecialObjectAgeByType = map[StarKind]SpecialObjectAgeRow{
	KindBrownDwarf:  {BaseFormula: "small_star", AddProgenitorAge: false},
	KindWhiteDwarf:  {BaseFormula: "small_star", AddProgenitorAge: true},
	KindPulsar:      {BaseFormula: "100m_per_2d10", AddProgenitorAge: true},
	KindNeutronStar: {BaseFormula: "small_star", AddProgenitorAge: true},
	KindBlackHole:   {BaseFormula: "small_star", AddProgenitorAge: true},
	KindProtostar:   {BaseFormula: "10m_per_2d10", AddProgenitorAge: false},
}
