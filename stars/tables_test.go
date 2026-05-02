package stars

import "testing"

func TestStarTypeDetermination_Complete(t *testing.T) {
	for r := 2; r <= 12; r++ {
		row, ok := StarTypeDetermination[r]
		if !ok {
			t.Fatalf("missing row %d", r)
		}
		if row.Type == "" || row.Hot == "" || row.Special == "" ||
			row.Unusual == "" || row.Giants == "" || row.Peculiar == "" {
			t.Fatalf("row %d has empty cell: %+v", r, row)
		}
	}
}

func TestStarTypeDetermination_KnownCells(t *testing.T) {
	// WBH p. 15 spot checks.
	checks := []struct {
		row  int
		col  string
		want string
	}{
		{7, "Type", "K"},
		{7, "Hot", "A"},
		{2, "Type", "Special"},
		{2, "Peculiar", "Black Hole"},
		{12, "Type", "Hot"},
		{12, "Hot", "O"},
		{12, "Giants", "Class Ia"},
		{11, "Type", "F"},
		{11, "Peculiar", "Anomaly"},
	}
	for _, c := range checks {
		t.Run(c.col, func(t *testing.T) {
			row := StarTypeDetermination[c.row]
			var got string
			switch c.col {
			case "Type":
				got = row.Type
			case "Hot":
				got = row.Hot
			case "Special":
				got = row.Special
			case "Unusual":
				got = row.Unusual
			case "Giants":
				got = row.Giants
			case "Peculiar":
				got = row.Peculiar
			}
			if got != c.want {
				t.Fatalf("row %d %s = %q, want %q", c.row, c.col, got, c.want)
			}
		})
	}
}

func TestStarSubtype_Complete(t *testing.T) {
	for r := 2; r <= 12; r++ {
		if _, ok := StarSubtypeNumeric[r]; !ok {
			t.Fatalf("Numeric: missing row %d", r)
		}
		if _, ok := StarSubtypeMType[r]; !ok {
			t.Fatalf("M-type: missing row %d", r)
		}
	}
}

func TestStarSubtype_KnownCells(t *testing.T) {
	// WBH p. 16 — Zed primary: 2D=6 -> Numeric subtype 7 (G7).
	if got := StarSubtypeNumeric[6]; got != 7 {
		t.Fatalf("Numeric[6] = %d, want 7", got)
	}
	if got := StarSubtypeNumeric[2]; got != 0 {
		t.Fatalf("Numeric[2] = %d, want 0", got)
	}
	if got := StarSubtypeNumeric[12]; got != 0 {
		t.Fatalf("Numeric[12] = %d, want 0", got)
	}
	if got := StarSubtypeMType[6]; got != 0 {
		t.Fatalf("MType[6] = %d, want 0", got)
	}
	if got := StarSubtypeMType[12]; got != 9 {
		t.Fatalf("MType[12] = %d, want 9", got)
	}
}

func TestStarMass_KnownCells(t *testing.T) {
	// WBH p. 17 — spot checks.
	tests := []struct {
		spectral string
		class    LuminosityClass
		want     float64
	}{
		{"G0", V, 1.1},
		{"G5", V, 0.9},
		{"K0", V, 0.8},
		{"O0", Ia, 200},
		{"M9", VI, 0.075},
	}
	for _, tc := range tests {
		row, ok := StarMass[tc.spectral]
		if !ok {
			t.Fatalf("missing row %s", tc.spectral)
		}
		got, ok := row.Get(tc.class)
		if !ok {
			t.Fatalf("missing %s %s cell", tc.spectral, tc.class)
		}
		if got != tc.want {
			t.Fatalf("%s %s = %v, want %v", tc.spectral, tc.class, got, tc.want)
		}
	}
	// Class IV not available for O0; should be absent.
	if _, ok := StarMass["O0"].Get(IV); ok {
		t.Fatal("O0 IV unexpectedly present")
	}
}

func TestStarTemperature_KnownCells(t *testing.T) {
	cases := map[string]float64{
		"G0": 6000, "G5": 5600, "K0": 5200, "O0": 50000, "M9": 2400,
	}
	for s, want := range cases {
		if got := StarTemperature[s]; got != want {
			t.Fatalf("%s = %v, want %v", s, got, want)
		}
	}
}

func TestStarDiameter_KnownCells(t *testing.T) {
	tests := []struct {
		spectral string
		class    LuminosityClass
		want     float64
	}{
		{"G0", V, 1.1},
		{"G5", V, 0.95},
		{"K0", V, 0.9},
		{"O0", Ia, 25},
		{"M9", VI, 0.08},
	}
	for _, tc := range tests {
		got, ok := StarDiameter[tc.spectral].Get(tc.class)
		if !ok {
			t.Fatalf("missing %s %s", tc.spectral, tc.class)
		}
		if got != tc.want {
			t.Fatalf("%s %s = %v, want %v", tc.spectral, tc.class, got, tc.want)
		}
	}
}

func TestStarLuminosity_KnownCells(t *testing.T) {
	tests := []struct {
		spectral string
		class    LuminosityClass
		want     float64
	}{
		{"G0", V, 1.4},
		{"G5", V, 0.78},
		{"K0", V, 0.52},
		{"O0", Ia, 3_400_000},
		{"M9", VI, 0.00019},
	}
	for _, tc := range tests {
		got, ok := StarLuminosity[tc.spectral].Get(tc.class)
		if !ok {
			t.Fatalf("missing %s %s", tc.spectral, tc.class)
		}
		if got != tc.want {
			t.Fatalf("%s %s = %v, want %v", tc.spectral, tc.class, got, tc.want)
		}
	}
}

func TestMultipleStarsPresenceThreshold(t *testing.T) {
	// WBH p.23: each orbit class is "10+" on 2D + DMs.
	if MultipleStarsPresenceThreshold != 10 {
		t.Fatalf("threshold = %d, want 10", MultipleStarsPresenceThreshold)
	}
}

func TestExistingStarLocationsBinary_Complete(t *testing.T) {
	for r := 1; r <= 6; r++ {
		if _, ok := ExistingStarLocationsBinary[r]; !ok {
			t.Fatalf("binary: missing row %d", r)
		}
	}
}

func TestExistingStarLocationsTrinaryPlus_Complete(t *testing.T) {
	for r := 1; r <= 6; r++ {
		if _, ok := ExistingStarLocationsTrinaryPlus[r]; !ok {
			t.Fatalf("trinary+: missing row %d", r)
		}
	}
}

func TestNonPrimaryStarDetermination_Complete(t *testing.T) {
	for r := 2; r <= 12; r++ {
		row := NonPrimaryStarDetermination[r]
		if row.Secondary == "" || row.Companion == "" || row.PostStellar == "" || row.Other == "" {
			t.Fatalf("row %d has empty cell: %+v", r, row)
		}
	}
}

func TestNonPrimaryStarDetermination_KnownCells(t *testing.T) {
	// WBH p.29 spot checks.
	if got := NonPrimaryStarDetermination[8].Companion; got != "Sibling" {
		t.Fatalf("[8].Companion = %q want Sibling", got)
	}
	if got := NonPrimaryStarDetermination[8].Secondary; got != "Lesser" {
		t.Fatalf("[8].Secondary = %q want Lesser", got)
	}
	if got := NonPrimaryStarDetermination[11].Other; got != "BD" {
		t.Fatalf("[11].Other = %q want BD", got)
	}
}

func TestEccentricityValues_Complete(t *testing.T) {
	for r := 5; r <= 12; r++ {
		row, ok := EccentricityValues[r]
		if !ok {
			t.Fatalf("missing row %d", r)
		}
		if row.SecondRoll == "" || row.Divisor == 0 {
			t.Fatalf("row %d incomplete: %+v", r, row)
		}
	}
}

func TestEccentricityValues_KnownCells(t *testing.T) {
	// WBH p.27 spot checks.
	if got := EccentricityValues[5].Base; got != -0.001 {
		t.Fatalf("[5].Base = %v want -0.001", got)
	}
	if got := EccentricityValues[12].Base; got != 0.30 {
		t.Fatalf("[12].Base = %v want 0.30", got)
	}
	if got := EccentricityValues[10].Divisor; got != 20 {
		t.Fatalf("[10].Divisor = %v want 20", got)
	}
}

// ----- P2-8: OrbitNumberTable -----

func TestOrbitNumberTable_Complete(t *testing.T) {
	for n := 0; n <= 20; n++ {
		if _, ok := OrbitNumberTable[n]; !ok {
			t.Fatalf("missing row %d", n)
		}
	}
}

func TestOrbitNumberTable_KnownCells(t *testing.T) {
	// WBH p.26 spot checks.
	if got := OrbitNumberTable[3].DistanceAU; got != 1.0 {
		t.Fatalf("[3].DistanceAU = %v want 1.0 (Terra)", got)
	}
	if got := OrbitNumberTable[6].DistanceAU; got != 5.2 {
		t.Fatalf("[6].DistanceAU = %v want 5.2 (Jupiter)", got)
	}
	if got := OrbitNumberTable[3].Example; got != "Terra" {
		t.Fatalf("[3].Example = %q want Terra", got)
	}
}
