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

func f(x float64) *float64 { return &x }

// StarMass is the WBH p. 17 Star Mass and Temperature by Class table — Mass column.
// Values are in solar masses (Sol = 1.0).
var StarMass = map[string]ClassRow{
	"O0": {Ia: f(200), Ib: f(150), II: f(130), III: f(110), V: f(90), VI: f(2)},
	"O5": {Ia: f(80), Ib: f(60), II: f(40), III: f(30), V: f(60), VI: f(1.5)},
	"B0": {Ia: f(60), Ib: f(40), II: f(30), III: f(20), IV: f(20), V: f(18), VI: f(0.5)},
	"B5": {Ia: f(30), Ib: f(25), II: f(20), III: f(10), IV: f(10), V: f(5), VI: f(0.4)},
	"A0": {Ia: f(20), Ib: f(15), II: f(14), III: f(8), IV: f(4), V: f(2.2)},
	"A5": {Ia: f(15), Ib: f(13), II: f(11), III: f(6), IV: f(2.3), V: f(1.5)},
	"F0": {Ia: f(13), Ib: f(12), II: f(10), III: f(4), IV: f(2), V: f(1.5)},
	"F5": {Ia: f(12), Ib: f(10), II: f(8), III: f(3), IV: f(1.5), V: f(1.3)},
	"G0": {Ia: f(12), Ib: f(10), II: f(8), III: f(2.5), IV: f(1.7), V: f(1.1), VI: f(0.8)},
	"G5": {Ia: f(13), Ib: f(11), II: f(10), III: f(2.4), IV: f(1.2), V: f(0.9), VI: f(0.7)},
	"K0": {Ia: f(14), Ib: f(12), II: f(10), III: f(1.1), IV: f(1.5), V: f(0.8), VI: f(0.6)},
	"K5": {Ia: f(18), Ib: f(13), II: f(12), III: f(1.5), V: f(0.7), VI: f(0.5)},
	"M0": {Ia: f(20), Ib: f(15), II: f(14), III: f(1.8), V: f(0.5), VI: f(0.4)},
	"M5": {Ia: f(25), Ib: f(20), II: f(16), III: f(2.4), V: f(0.16), VI: f(0.12)},
	"M9": {Ia: f(30), Ib: f(25), II: f(18), III: f(8), V: f(0.08), VI: f(0.075)},
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
	"O0": {Ia: f(25), Ib: f(24), II: f(22), III: f(21), V: f(20), VI: f(0.18)},
	"O5": {Ia: f(22), Ib: f(20), II: f(18), III: f(15), V: f(12), VI: f(0.18)},
	"B0": {Ia: f(20), Ib: f(14), II: f(12), III: f(10), IV: f(8), V: f(7), VI: f(0.2)},
	"B5": {Ia: f(60), Ib: f(25), II: f(14), III: f(6), IV: f(5), V: f(3.5), VI: f(0.5)},
	"A0": {Ia: f(120), Ib: f(50), II: f(30), III: f(5), IV: f(4), V: f(2.2)},
	"A5": {Ia: f(180), Ib: f(75), II: f(45), III: f(5), IV: f(3), V: f(2)},
	"F0": {Ia: f(210), Ib: f(85), II: f(50), III: f(5), IV: f(3), V: f(1.7)},
	"F5": {Ia: f(280), Ib: f(115), II: f(66), III: f(5), IV: f(2), V: f(1.5)},
	"G0": {Ia: f(330), Ib: f(135), II: f(77), III: f(10), IV: f(3), V: f(1.1), VI: f(0.8)},
	"G5": {Ia: f(360), Ib: f(150), II: f(90), III: f(15), IV: f(4), V: f(0.95), VI: f(0.7)},
	"K0": {Ia: f(420), Ib: f(180), II: f(110), III: f(20), IV: f(6), V: f(0.9), VI: f(0.6)},
	"K5": {Ia: f(600), Ib: f(260), II: f(160), III: f(40), V: f(0.8), VI: f(0.5)},
	"M0": {Ia: f(900), Ib: f(380), II: f(230), III: f(60), V: f(0.7), VI: f(0.4)},
	"M5": {Ia: f(1200), Ib: f(600), II: f(350), III: f(100), V: f(0.2), VI: f(0.1)},
	"M9": {Ia: f(1800), Ib: f(800), II: f(500), III: f(200), V: f(0.1), VI: f(0.08)},
}

// StarLuminosity is the WBH p. 19 Star Luminosity by Class table.
// Values are in solar luminosities (Sol = 1.0).
var StarLuminosity = map[string]ClassRow{
	"O0": {Ia: f(3_400_000), Ib: f(3_200_000), II: f(2_700_000), III: f(2_400_000), V: f(2_200_000), VI: f(180)},
	"O5": {Ia: f(1_100_000), Ib: f(900_000), II: f(730_000), III: f(510_000), V: f(330_000), VI: f(73)},
	"B0": {Ia: f(290_000), Ib: f(140_000), II: f(100_000), III: f(72_000), IV: f(46_000), V: f(35_000), VI: f(29)},
	"B5": {Ia: f(160_000), Ib: f(28_000), II: f(8800), III: f(1600), IV: f(1100), V: f(550), VI: f(11)},
	"A0": {Ia: f(130_000), Ib: f(22_000), II: f(8000), III: f(220), IV: f(140), V: f(43)},
	"A5": {Ia: f(120_000), Ib: f(20_000), II: f(7300), III: f(90), IV: f(33), V: f(15)},
	"F0": {Ia: f(120_000), Ib: f(20_000), II: f(7000), III: f(70), IV: f(25), V: f(8.1)},
	"F5": {Ia: f(120_000), Ib: f(20_000), II: f(6900), III: f(39), IV: f(6), V: f(3.5)},
	"G0": {Ia: f(120_000), Ib: f(20_000), II: f(6800), III: f(120), IV: f(10), V: f(1.4), VI: f(0.73)},
	"G5": {Ia: f(110_000), Ib: f(20_000), II: f(7000), III: f(200), IV: f(14), V: f(0.78), VI: f(0.43)},
	"K0": {Ia: f(110_000), Ib: f(21_000), II: f(7800), III: f(260), IV: f(23), V: f(0.52), VI: f(0.23)},
	"K5": {Ia: f(120_000), Ib: f(22_000), II: f(8400), III: f(530), V: f(0.21), VI: f(0.083)},
	"M0": {Ia: f(130_000), Ib: f(24_000), II: f(8800), III: f(600), V: f(0.082), VI: f(0.027)},
	"M5": {Ia: f(100_000), Ib: f(26_000), II: f(8800), III: f(720), V: f(0.0029), VI: f(0.00072)},
	"M9": {Ia: f(90_000), Ib: f(19_000), II: f(7300), III: f(1200), V: f(0.00029), VI: f(0.00019)},
}

// MultipleStarsPresenceThreshold is the WBH p.23 2D threshold (after
// DMs) for a star to be present in a given orbit class.
const MultipleStarsPresenceThreshold = 10

// ExistingStarLocationsBinary is the WBH p.24 Existing Star Locations table
// for binary systems, keyed by 1D.
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
// table for trinary and larger systems, keyed by 1D.
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
