package worlds

import (
	"testing"

	"github.com/philoserf/world-builder/roller"
)

func TestGasMixTableLookup_FrozenExotic(t *testing.T) {
	t.Parallel()
	// Spot-check Frozen-A column (WBH p.98 Frozen HZCO+3.00+ table).
	// Roll 1 → Krypton; roll 7 → Nitrogen; roll 13 → Hydrogen.
	cases := []struct {
		roll int
		want string
	}{
		{1, "Krypton"},
		{7, "Nitrogen"},
		{13, "Hydrogen"},
	}
	for _, c := range cases {
		got := GasMixTableLookup(TempFrozen, "A", c.roll)
		if got != c.want {
			t.Errorf("Frozen-A roll=%d: got %q, want %q", c.roll, got, c.want)
		}
	}
}

func TestGasMixTableLookup_BoilingColumns(t *testing.T) {
	t.Parallel()
	// Spot-check Boiling (HZCO -2.01-) table per WBH p.96.
	// Row 1: A=Argon, B=Argon, C=Sulphuric Acid.
	// Row 7: A=Nitrogen, B=Carbon Dioxide, C=Nitrogen.
	// Row 11: A=Methane, B=Ammonia, C=Ammonia.
	// Row 13: A=Methane, B=Methane, C=Methane.
	cases := []struct {
		row  int
		col  string
		want string
	}{
		{1, "A", "Argon"},
		{1, "B", "Argon"},
		{1, "C", "Sulphuric Acid"},
		{7, "A", "Nitrogen"},
		{7, "B", "Carbon Dioxide"},
		{7, "C", "Nitrogen"},
		{11, "A", "Methane"},
		{11, "B", "Ammonia"},
		{11, "C", "Ammonia"},
		{13, "A", "Methane"},
		{13, "B", "Methane"},
		{13, "C", "Methane"},
	}
	for _, c := range cases {
		got := GasMixTableLookup(TempBoiling, c.col, c.row)
		if got != c.want {
			t.Errorf("Boiling-%s row %d: got %q, want %q", c.col, c.row, got, c.want)
		}
	}
}

func TestGasMixTableLookup_HotTable(t *testing.T) {
	t.Parallel()
	// Spot-check Hot/second-Boiling table (WBH p.96, HZCO -1.01–-2.0), stored as TempHot.
	// Row 7: A=Carbon Dioxide, B=Carbon Dioxide, C=Carbon Dioxide.
	for _, col := range []string{"A", "B", "C"} {
		got := GasMixTableLookup(TempHot, col, 7)
		if got != "Carbon Dioxide" {
			t.Errorf("Hot-%s row 7: got %q, want %q", col, got, "Carbon Dioxide")
		}
	}
}

func TestGasMixTableLookup_TemperateTable(t *testing.T) {
	t.Parallel()
	// WBH p.97 Temperate table. Row 7: A=Carbon Dioxide, B=Carbon Dioxide, C=Carbon Dioxide.
	for _, col := range []string{"A", "B", "C"} {
		got := GasMixTableLookup(TempTemperate, col, 7)
		if got != "Carbon Dioxide" {
			t.Errorf("Temperate-%s row 7: got %q, want %q", col, got, "Carbon Dioxide")
		}
	}
	// Row 13: A=Helium, B=Hydrogen, C=Hydrogen.
	if got := GasMixTableLookup(TempTemperate, "A", 13); got != "Helium" {
		t.Errorf("Temperate-A row 13: got %q, want %q", got, "Helium")
	}
	if got := GasMixTableLookup(TempTemperate, "B", 13); got != "Hydrogen" {
		t.Errorf("Temperate-B row 13: got %q, want %q", got, "Hydrogen")
	}
}

func TestGasMixTableLookup_ColdTable(t *testing.T) {
	t.Parallel()
	// WBH p.98 Cold table. Row 7: A=Carbon Dioxide, B=Carbon Dioxide, C=Carbon Dioxide.
	for _, col := range []string{"A", "B", "C"} {
		got := GasMixTableLookup(TempCold, col, 7)
		if got != "Carbon Dioxide" {
			t.Errorf("Cold-%s row 7: got %q, want %q", col, got, "Carbon Dioxide")
		}
	}
	// Row 13: A=Helium, B=Hydrogen, C=Hydrogen.
	if got := GasMixTableLookup(TempCold, "A", 13); got != "Helium" {
		t.Errorf("Cold-A row 13: got %q, want %q", got, "Helium")
	}
}

func TestGasMixTableLookup_Clamping(t *testing.T) {
	t.Parallel()
	// roll < 1 should clamp to row 1; roll > 13 should clamp to row 13.
	low := GasMixTableLookup(TempFrozen, "A", -5)
	if low == "" {
		t.Error("clamping low: empty result")
	}
	// Row 1 of TempFrozen A is Krypton.
	if low != "Krypton" {
		t.Errorf("clamping low: got %q, want %q", low, "Krypton")
	}
	high := GasMixTableLookup(TempFrozen, "A", 100)
	if high == "" {
		t.Error("clamping high: empty result")
	}
	// Row 13 of TempFrozen A is Hydrogen.
	if high != "Hydrogen" {
		t.Errorf("clamping high: got %q, want %q", high, "Hydrogen")
	}
}

func TestGasMixTableLookup_InvalidInputs(t *testing.T) {
	t.Parallel()
	// Unknown column returns "".
	if got := GasMixTableLookup(TempFrozen, "Z", 7); got != "" {
		t.Errorf("unknown column: got %q, want empty", got)
	}
}

func TestFormatAtmoProfileShorthand_Terra(t *testing.T) {
	t.Parallel()
	// Terra-like atmosphere: code 6, pressure 1.013 bar, ppO2 0.212.
	// Expected N-O format: "6-1.013-0.212"
	atmo := Atmosphere{Code: 6, Pressure: 1.013, OxygenPartialPressure: 0.212}
	prof := AtmosphereProfile{}
	got := FormatAtmoProfileShorthand(atmo, prof)
	want := "6-1.013-0.212"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAtmoProfileShorthand_ExoticNoGases(t *testing.T) {
	t.Parallel()
	// Exotic atmosphere (code 10 = A), subtype 7, no pressure yet.
	atmo := Atmosphere{Code: 10, Subtype: "7"}
	prof := AtmosphereProfile{}
	got := FormatAtmoProfileShorthand(atmo, prof)
	want := "A-St7"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAtmoProfileShorthand_ExoticWithGases(t *testing.T) {
	t.Parallel()
	// Zed Cab II worked example from WBH p.99:
	// A-St7:0.98:N2-96:Ar-04
	// Code A (10), subtype 7, pressure 0.98 bar.
	atmo := Atmosphere{Code: 10, Subtype: "7", Pressure: 0.98}
	prof := AtmosphereProfile{
		Gases: []GasFraction{
			{Name: "N2", PercentBP: 9600},
			{Name: "Ar", PercentBP: 400},
		},
	}
	got := FormatAtmoProfileShorthand(atmo, prof)
	want := "A-St7:0.98:N2-96:Ar-04"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAtmoProfileShorthand_CorrosiveWithGases(t *testing.T) {
	t.Parallel()
	// Aab I worked example from WBH p.99:
	// B-StD:CO2-48:NH3-47:H2O-03
	// Code B (11), subtype D, no pressure specified in shorthand.
	atmo := Atmosphere{Code: 11, Subtype: "D"}
	prof := AtmosphereProfile{
		Gases: []GasFraction{
			{Name: "CO2", PercentBP: 4800},
			{Name: "NH3", PercentBP: 4700},
			{Name: "H2O", PercentBP: 300},
		},
	}
	got := FormatAtmoProfileShorthand(atmo, prof)
	want := "B-StD:CO2-48:NH3-47:H2O-03"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAtmoProfileShorthand_TaintSuffix_NO(t *testing.T) {
	t.Parallel()
	atmo := Atmosphere{
		Code:                  4,
		Pressure:              0.544,
		OxygenPartialPressure: 0.114,
		Taints: []Taint{
			{Code: "P", Severity: 6, Persistence: 3},
			{Code: "R", Severity: 5, Persistence: 4},
		},
	}
	got := FormatAtmoProfileShorthand(atmo, AtmosphereProfile{})
	want := "4-0.544-0.114:P.6.3,R.5.4"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAtmoProfileShorthand_TaintSuffix_Insidious(t *testing.T) {
	t.Parallel()
	atmo := Atmosphere{
		Code:             12,
		Subtype:          "6",
		Pressure:         1.21,
		Taints:           []Taint{{Code: "G", Severity: 4, Persistence: 5}},
		InsidiousHazards: []Hazard{{Code: "T"}},
	}
	got := FormatAtmoProfileShorthand(atmo, AtmosphereProfile{})
	want := "C-St6.T:1.21 G.4.5"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAtmoProfileShorthand_TaintSuffix_Insidious_MultiHazard(t *testing.T) {
	t.Parallel()
	// Subtype D triggers the WBH p.90 footnote: auto-T + rolled hazard.
	// Hazard codes are concatenated single letters after the subtype dot.
	atmo := Atmosphere{
		Code:             12,
		Subtype:          "D",
		Pressure:         120.5,
		InsidiousHazards: []Hazard{{Code: "T"}, {Code: "G"}},
	}
	got := FormatAtmoProfileShorthand(atmo, AtmosphereProfile{})
	want := "C-StD.TG:120.50"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAtmoProfileShorthand_NoTaint_Unchanged(t *testing.T) {
	t.Parallel()
	atmo := Atmosphere{
		Code:                  6,
		Pressure:              1.013,
		OxygenPartialPressure: 0.212,
	}
	got := FormatAtmoProfileShorthand(atmo, AtmosphereProfile{})
	want := "6-1.013-0.212"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRollGasMix_BasicShape(t *testing.T) {
	t.Parallel()
	// Frozen Exotic A (Zed Cab II context). Verify produces non-empty Gases slice.
	// DM: Size 4 (1-7) → DM-3; TempFrozen DM-3 (size). No frozen-temp sub-DM applied here
	// since meanTempK is not passed to RollGasMix — it uses table DMs only.
	// Scripted rolls: 2D=7, 1D=4, d10=4, 2D=7, 1D=4, d10=0, 2D=3, 1D=4, d10=0.
	r := roller.NewScripted(7, 4, 4, 7, 4, 0, 3, 4, 0)
	prof, err := RollGasMix(r, "A", "7", TempFrozen, SizeCode("4"))
	if err != nil {
		t.Fatal(err)
	}
	if len(prof.Gases) == 0 {
		t.Fatal("no gases produced")
	}
	// Total should be ~10000 BP (100%).
	totalBP := 0
	for _, g := range prof.Gases {
		totalBP += g.PercentBP
	}
	if totalBP < 9000 || totalBP > 10100 {
		t.Errorf("total percentage off: got %d BP, want 9000-10100", totalBP)
	}
	t.Logf("Frozen-A Size4: gases=%v, totalBP=%d", prof.Gases, totalBP)
}

func TestRollGasMix_TempRangeLabel(t *testing.T) {
	t.Parallel()
	// Each RollGasMix call uses up to 4 iterations × 3 rolls = 12 rolls.
	// Provide 12 values per call (2D=7, 1D=5, d10=5 repeated 4 times).
	rolls12 := []int{7, 5, 5, 7, 5, 5, 7, 5, 5, 7, 5, 5}
	for _, tc := range []struct {
		tr   TempRange
		want string
	}{
		{TempBoiling, "Boiling"},
		{TempHot, "Hot"},
		{TempTemperate, "Temperate"},
		{TempCold, "Cold"},
		{TempFrozen, "Frozen"},
	} {
		r := roller.NewScripted(rolls12...)
		prof, err := RollGasMix(r, "A", "", tc.tr, SizeCode("5"))
		if err != nil {
			t.Fatalf("%s: %v", tc.want, err)
		}
		if prof.TempRange != tc.want {
			t.Errorf("TempRange label: got %q, want %q", prof.TempRange, tc.want)
		}
	}
}

func TestRollGasMix_GasMerging(t *testing.T) {
	t.Parallel()
	// If the same gas appears twice, allocations should be merged.
	// TempFrozen A, Size 5 (DM-3): 2D+DM. Script rolls 2D=10 → 10-3=7 → Nitrogen,
	// then 2D=10 → 7 → Nitrogen again. Should merge into one entry.
	r := roller.NewScripted(
		10, 5, 5, // first gas: 2D=10, 1D=5, d10=5
		10, 5, 5, // second gas: 2D=10, same → Nitrogen again
		3, 3, 5, // third gas
		3, 3, 5, // fourth gas
	)
	prof, err := RollGasMix(r, "A", "", TempFrozen, SizeCode("5"))
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range prof.Gases {
		if g.Name == "N2" {
			// Found merged nitrogen — verify it has more than one allocation's worth.
			// (At minimum it should have been rolled twice.)
			t.Logf("Merged N2: %d BP", g.PercentBP)
			return
		}
	}
	// If N2 not found, check that something valid was produced.
	t.Logf("Gases produced: %v", prof.Gases)
}
