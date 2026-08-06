package worlds

import (
	"fmt"
	"strings"

	"github.com/philoserf/world-builder/roller"
)

// gasMixTables maps (TempRange, column "A"/"B"/"C") → 13-row gas name table.
// Row index = (2D+DM) clamped to [1,13], stored at index [roll-1].
//
// Table sources (WBH pp.96-98):
//
//	TempBoiling   → "Boiling Atmosphere Gas Mix (HZCO -2.01-)"  — mean ≥ 453 K (p.96)
//	TempHot       → "Boiling Atmosphere Gas Mix (HZCO -1.01–-2.0)" — 353-453 K (p.96)
//	                  (the book calls this "Boiling" but it occupies the TempHot HZCO band)
//	TempTemperate → "Temperate Atmosphere Gas Mix" — 273-303 K (p.97)
//	                  (TempTemperate spans 273-353 K; Temperate table used as primary match)
//	TempCold      → "Cold Atmosphere Gas Mix" — 223-273 K (p.98)
//	TempFrozen    → "Frozen Atmosphere Gas Mix (HZCO +3.00+)" — below 123 K (p.98)
var gasMixTables = map[TempRange]map[string][13]string{
	// WBH p.96: Boiling Atmosphere Gas Mix (HZCO -2.01-)
	// Mean temperatures 453K+ (180°C+).
	// The book table has rows labelled "-2-" through "13" (16 rows).
	// Rows -2-, -1, 0 are only reachable via large negative DMs (temp 700-2000K: DM-2;
	// temp >2000K: DM-5). Since we clamp rolls to [1,13], negative-result rolls land at row 1.
	// Row 1 = book row "1"; row 13 = book row "13".
	// DMs: mean 700-2000K → DM-2; mean >2000K → DM-5; Size 1-7 → DM-1; Size A+ → DM+1.
	TempBoiling: {
		"A": {
			/* row 1  */ "Argon",
			/* row 2  */ "Sulphur Dioxide",
			/* row 3  */ "Carbon Monoxide",
			/* row 4  */ "Carbon Dioxide",
			/* row 5  */ "Nitrogen",
			/* row 6  */ "Carbon Dioxide",
			/* row 7  */ "Nitrogen",
			/* row 8  */ "Water Vapour",
			/* row 9  */ "Sulphur Dioxide",
			/* row 10 */ "Nitrogen",
			/* row 11 */ "Methane",
			/* row 12 */ "Water Vapour",
			/* row 13 */ "Methane",
		},
		"B": {
			/* row 1  */ "Argon",
			/* row 2  */ "Sulphur Dioxide",
			/* row 3  */ "Hydrogen Cyanide",
			/* row 4  */ "Formamide",
			/* row 5  */ "Carbon Dioxide",
			/* row 6  */ "Nitrogen",
			/* row 7  */ "Carbon Dioxide",
			/* row 8  */ "Sulphur Dioxide",
			/* row 9  */ "Water Vapour",
			/* row 10 */ "Nitrogen",
			/* row 11 */ "Ammonia",
			/* row 12 */ "Ammonia",
			/* row 13 */ "Methane",
		},
		"C": {
			/* row 1  */ "Sulphuric Acid",
			/* row 2  */ "Hydrochloric Acid",
			/* row 3  */ "Chlorine",
			/* row 4  */ "Fluorine",
			/* row 5  */ "Formic Acid",
			/* row 6  */ "Water Vapour",
			/* row 7  */ "Nitrogen",
			/* row 8  */ "Carbon Dioxide",
			/* row 9  */ "Sulphur Dioxide",
			/* row 10 */ "Hydrogen Cyanide",
			/* row 11 */ "Ammonia",
			/* row 12 */ "Hydrofluoric Acid",
			/* row 13 */ "Methane",
		},
	},

	// WBH p.96: Boiling Atmosphere Gas Mix (HZCO -1.01–-2.0)
	// Mean temperatures 353-453K (80-180°C). Maps to TempHot in code.
	// DMs: Size 1-7 → DM-1; Size A+ → DM+1.
	TempHot: {
		"A": {
			/* row 1 */ "Krypton",
			/* row 2 */ "Argon",
			/* row 3 */ "Sulphur Dioxide",
			/* row 4 */ "Ethane",
			/* row 5 */ "Carbon Dioxide",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Carbon Dioxide",
			/* row 8 */ "Nitrogen",
			/* row 9 */ "Water Vapour",
			/* row 10 */ "Sulphur Dioxide",
			/* row 11 */ "Methane",
			/* row 12 */ "Neon",
			/* row 13 */ "Methane",
		},
		"B": {
			/* row 1 */ "Argon",
			/* row 2 */ "Sulphur Dioxide",
			/* row 3 */ "Hydrogen Cyanide",
			/* row 4 */ "Ethane",
			/* row 5 */ "Carbon Dioxide",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Carbon Dioxide",
			/* row 8 */ "Sulphur Dioxide",
			/* row 9 */ "Water Vapour",
			/* row 10 */ "Nitrogen",
			/* row 11 */ "Ammonia",
			/* row 12 */ "Ammonia",
			/* row 13 */ "Methane",
		},
		"C": {
			/* row 1 */ "Hydrochloric Acid",
			/* row 2 */ "Chlorine",
			/* row 3 */ "Fluorine",
			/* row 4 */ "Formic Acid",
			/* row 5 */ "Water Vapour",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Carbon Dioxide",
			/* row 8 */ "Sulphur Dioxide",
			/* row 9 */ "Hydrogen Cyanide",
			/* row 10 */ "Ammonia",
			/* row 11 */ "Methane",
			/* row 12 */ "Hydrofluoric Acid",
			/* row 13 */ "Methane",
		},
	},

	// WBH p.97: Temperate Atmosphere Gas Mix
	// Mean temperatures 273-303K (0-30°C).
	// DMs: Size 1-7 → DM-1; Size A+ → DM+1.
	TempTemperate: {
		"A": {
			/* row 1 */ "Krypton",
			/* row 2 */ "Argon",
			/* row 3 */ "Sulphur Dioxide",
			/* row 4 */ "Nitrogen",
			/* row 5 */ "Carbon Monoxide",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Carbon Dioxide",
			/* row 8 */ "Ethane",
			/* row 9 */ "Nitrogen",
			/* row 10 */ "Neon",
			/* row 11 */ "Methane",
			/* row 12 */ "Methane",
			/* row 13 */ "Helium",
		},
		"B": {
			/* row 1 */ "Krypton",
			/* row 2 */ "Chlorine",
			/* row 3 */ "Argon",
			/* row 4 */ "Sulphur Dioxide",
			/* row 5 */ "Carbon Monoxide",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Carbon Dioxide",
			/* row 8 */ "Ethane",
			/* row 9 */ "Ammonia",
			/* row 10 */ "Ammonia",
			/* row 11 */ "Methane",
			/* row 12 */ "Helium",
			/* row 13 */ "Hydrogen",
		},
		"C": {
			/* row 1 */ "Argon",
			/* row 2 */ "Chlorine",
			/* row 3 */ "Fluorine",
			/* row 4 */ "Sulphur Dioxide",
			/* row 5 */ "Carbon Monoxide",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Carbon Dioxide",
			/* row 8 */ "Ethane",
			/* row 9 */ "Ammonia",
			/* row 10 */ "Ammonia",
			/* row 11 */ "Methane",
			/* row 12 */ "Helium",
			/* row 13 */ "Hydrogen",
		},
	},

	// WBH p.98: Cold Atmosphere Gas Mix
	// Mean temperatures 223-273K (-50–0°C).
	// DMs: Size 1-7 → DM-1; Size A+ → DM+1.
	TempCold: {
		"A": {
			/* row 1 */ "Krypton",
			/* row 2 */ "Argon",
			/* row 3 */ "Ethane",
			/* row 4 */ "Nitrogen",
			/* row 5 */ "Carbon Monoxide",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Carbon Dioxide",
			/* row 8 */ "Nitrogen",
			/* row 9 */ "Ethane",
			/* row 10 */ "Methane",
			/* row 11 */ "Neon",
			/* row 12 */ "Methane",
			/* row 13 */ "Helium",
		},
		"B": {
			/* row 1 */ "Krypton",
			/* row 2 */ "Chlorine",
			/* row 3 */ "Argon",
			/* row 4 */ "Nitrogen",
			/* row 5 */ "Carbon Monoxide",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Carbon Dioxide",
			/* row 8 */ "Nitrogen",
			/* row 9 */ "Ethane",
			/* row 10 */ "Ammonia",
			/* row 11 */ "Methane",
			/* row 12 */ "Helium",
			/* row 13 */ "Hydrogen",
		},
		"C": {
			/* row 1 */ "Argon",
			/* row 2 */ "Chlorine",
			/* row 3 */ "Fluorine",
			/* row 4 */ "Ethane",
			/* row 5 */ "Carbon Monoxide",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Carbon Dioxide",
			/* row 8 */ "Nitrogen",
			/* row 9 */ "Ethane",
			/* row 10 */ "Ammonia",
			/* row 11 */ "Methane",
			/* row 12 */ "Helium",
			/* row 13 */ "Hydrogen",
		},
	},

	// WBH p.98: Frozen Atmosphere Gas Mix (HZCO +3.00+)
	// Mean temperatures below 123K (-150°C).
	// DMs: mean 70-100K → DM+3; mean <70K → DM+5; Size 1-7 → DM-3; Size A+ → DM+1.
	TempFrozen: {
		"A": {
			/* row 1 */ "Krypton",
			/* row 2 */ "Argon",
			/* row 3 */ "Argon",
			/* row 4 */ "Methane",
			/* row 5 */ "Carbon Monoxide",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Nitrogen",
			/* row 8 */ "Neon",
			/* row 9 */ "Helium",
			/* row 10 */ "Helium",
			/* row 11 */ "Hydrogen",
			/* row 12 */ "Hydrogen",
			/* row 13 */ "Hydrogen",
		},
		"B": {
			/* row 1 */ "Krypton",
			/* row 2 */ "Argon",
			/* row 3 */ "Argon",
			/* row 4 */ "Methane",
			/* row 5 */ "Carbon Monoxide",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Nitrogen",
			/* row 8 */ "Neon",
			/* row 9 */ "Helium",
			/* row 10 */ "Helium",
			/* row 11 */ "Hydrogen",
			/* row 12 */ "Hydrogen",
			/* row 13 */ "Hydrogen",
		},
		"C": {
			/* row 1 */ "Krypton",
			/* row 2 */ "Argon",
			/* row 3 */ "Fluorine",
			/* row 4 */ "Methane",
			/* row 5 */ "Carbon Monoxide",
			/* row 6 */ "Nitrogen",
			/* row 7 */ "Nitrogen",
			/* row 8 */ "Neon",
			/* row 9 */ "Helium",
			/* row 10 */ "Helium",
			/* row 11 */ "Hydrogen",
			/* row 12 */ "Hydrogen",
			/* row 13 */ "Hydrogen",
		},
	},
}

// GasMixTableLookup returns the gas name at the given table coordinates.
// tempRange selects the table; subtypeColumn is "A", "B", or "C".
// roll is 2D+DM, clamped to [1, 13] before lookup.
func GasMixTableLookup(tempRange TempRange, subtypeColumn string, roll int) string {
	if roll < 1 {
		roll = 1
	}

	if roll > 13 {
		roll = 13
	}

	tbl, ok := gasMixTables[tempRange]
	if !ok {
		return ""
	}

	col, ok := tbl[subtypeColumn]
	if !ok {
		return ""
	}

	return col[roll-1]
}

// chemicalName converts a gas book name to its chemical formula for shorthand rendering.
func chemicalName(name string) string {
	m := map[string]string{
		"Hydrogen":            "H2",
		"Helium":              "He",
		"Methane":             "CH4",
		"Ammonia":             "NH3",
		"Water Vapour":        "H2O",
		"Hydrofluoric Acid":   "HF",
		"Neon":                "Ne",
		"Sodium":              "Na",
		"Nitrogen":            "N2",
		"Carbon Monoxide":     "CO",
		"Hydrogen Cyanide":    "HCN",
		"Ethane":              "C2H6",
		"Oxygen":              "O2",
		"Hydrochloric Acid":   "HCl",
		"Fluorine":            "F2",
		"Argon":               "Ar",
		"Carbon Dioxide":      "CO2",
		"Formamide":           "CH3NO",
		"Formic Acid":         "CH2O2",
		"Sulphur Dioxide":     "SO2",
		"Chlorine":            "Cl2",
		"Krypton":             "Kr",
		"Sulphuric Acid":      "H2SO4",
		"Silicates (SO, SO2)": "SiO2",
		"Metal Vapours":       "Metal",
	}
	if v, ok := m[name]; ok {
		return v
	}

	return name
}

// gasMixSizeDM returns the size-based DM for gas mix table rolls per WBH pp.96-98.
// Boiling (HZCO -2.01-): Size 1-7 → DM-1, Size A+ → DM+1 (same formula as others).
// Frozen (HZCO +3.00+): Size 1-7 → DM-3, Size A+ → DM+1.
func gasMixSizeDM(tempRange TempRange, sizeCode SizeCode) int {
	si := SizeAsInt(sizeCode)

	if tempRange == TempFrozen {
		switch {
		case si >= 1 && si <= 7:
			return -3
		case si >= 10:
			return 1
		}

		return 0
	}

	switch {
	case si >= 1 && si <= 7:
		return -1
	case si >= 10:
		return 1
	}

	return 0
}

// GasMixBoilingTempDM returns the additional DM for the HZCO -2.01- Boiling table
// based on mean temperature per WBH p.96: 700-2000K → DM-2; above 2000K → DM-5.
// Callers add this to the size DM before calling GasMixTableLookup.
func GasMixBoilingTempDM(meanTempK float64) int {
	switch {
	case meanTempK >= 700 && meanTempK <= 2000:
		return -2
	case meanTempK > 2000:
		return -5
	}

	return 0
}

// GasMixFrozenTempDM returns the additional DM for the Frozen (HZCO +3.00+) table
// based on mean temperature per WBH p.98: 70-100K → DM+3; below 70K → DM+5.
// Callers add this to the size DM before calling GasMixTableLookup.
func GasMixFrozenTempDM(meanTempK float64) int {
	switch {
	case meanTempK >= 70 && meanTempK <= 100:
		return 3
	case meanTempK < 70:
		return 5
	}

	return 0
}

// tempRangeLabel returns a human-readable label for a TempRange.
func tempRangeLabel(t TempRange) string {
	switch t {
	case TempBoiling:
		return "Boiling"
	case TempHot:
		return "Hot"
	case TempTemperate:
		return "Temperate"
	case TempCold:
		return "Cold"
	case TempFrozen:
		return "Frozen"
	}

	return ""
}

// RollGasMix rolls 2-4 gases per WBH pp.95-96 procedure:
//   - First roll: primary gas; its fraction = (1D+4)×10% with d10 variance (capped at 100%).
//   - Subsequent rolls: (1D+4)×10% of remaining fraction, up to 4 total iterations.
//   - Stop when cumulative allocations reach ≥ 95% of 10000 BP; remainder becomes "Other".
//
// atmosphereColumnLetter is the gas-mix table column selector — "A", "B", or
// "C" — derived from the atmosphere CODE (10→A, 11→B, 12→C), NOT from the
// Subtype digit/letter produced by RollCorrosiveInsidiousSubtype.
// Use gasMixColumnForAtmCode (in temperature_rederive.go) to map atm Code
// to the correct column letter.
//
// tempDM is the additional mean-temperature DM for the Boiling and
// Frozen tables (WBH p.96/p.98) — compute it via GasMixBoilingTempDM /
// GasMixFrozenTempDM and pass 0 for the other temperature ranges.
func RollGasMix(
	r roller.Roller,
	atmosphereColumnLetter string,
	tempDM int,
	tempRange TempRange,
	sizeCode SizeCode,
) (AtmosphereProfile, error) {
	prof := AtmosphereProfile{TempRange: tempRangeLabel(tempRange)}
	dm := gasMixSizeDM(tempRange, sizeCode) + tempDM

	remainingBP := 10000
	for iter := 0; iter < 4 && remainingBP > 500; iter++ {
		gasRoll := r.Roll("2D") + dm

		gas := GasMixTableLookup(tempRange, atmosphereColumnLetter, gasRoll)
		if gas == "" {
			gas = "Other"
		}

		// Primary fraction: (1D+4)×10% with ±variance from d10.
		// WBH p.96: "(1D+4) × 10% (with variance, but only up to 100%)"
		// The worked example shows: 1D=1 → 50%, d10=2 → 47% (−3% variance).
		// We model this as: base = (1D+4)×1000 BP; variance = (d10-5)×50 BP (±2.5%).
		baseFrac := (r.Roll("1D") + 4) * 1000
		variance := (r.Roll("d10") - 5) * 50
		fracOfRemaining := min(max(baseFrac+variance, 500), 10000)

		// Convert fraction-of-remaining to absolute BP.
		alloc := min(max(remainingBP*fracOfRemaining/10000, 100), remainingBP)

		// Merge if same gas already present.
		merged := false

		for i, g := range prof.Gases {
			if g.Name == chemicalName(gas) {
				prof.Gases[i].PercentBP += alloc
				merged = true

				break
			}
		}

		if !merged {
			prof.Gases = append(prof.Gases, GasFraction{Name: chemicalName(gas), PercentBP: alloc})
		}

		remainingBP -= alloc
	}

	if remainingBP > 0 {
		prof.Gases = append(prof.Gases, GasFraction{Name: "Other", PercentBP: remainingBP})
	}

	return prof, nil
}

// atmosphereCodeChar returns the UWP hex character for an atmosphere code
// integer, from the single p.79 table transcription in atmosphere.go.
func atmosphereCodeChar(code int) string {
	return atmosphereCodes[code].Char
}

// FormatAtmoProfileShorthand produces the atmosphere profile string per WBH p.82 (N-O codes),
// p.85 (exotic), p.88-89 (corrosive/insidious):
//
//   - N-O atmospheres (codes 2-9, D=13, E=14): "A-bar-ppo[:Gas-pct...][:T.S.P,...]"
//     e.g. "6-1.013-0.212" or "4-0.544-0.114:P.6.3,R.5.4"
//   - Exotic/Corrosive (codes A=10, B=11, F=15): "A-St#[:bar][:Gas-pct:Gas-pct...] [I.S.P,...]"
//     e.g. "A-St7:0.98:N2-96:Ar-04"
//   - Insidious (code C=12): subtype becomes "St#.H..." where H... is the
//     concatenated single-letter codes from InsidiousHazards (1 letter for
//     most subtypes; 2 letters when subtype is D or E per WBH p.90 footnote)
//     e.g. "C-St6.T:1.21 G.4.5" or "C-StD.TG:120.50"
func FormatAtmoProfileShorthand(atmo Atmosphere, prof AtmosphereProfile) string {
	codeChar := atmosphereCodeChar(atmo.Code)

	isNO := (atmo.Code >= 2 && atmo.Code <= 9) || atmo.Code == 13 || atmo.Code == 14
	if isNO {
		base := fmt.Sprintf("%s-%.3f-%.3f", codeChar, atmo.Pressure, atmo.OxygenPartialPressure)
		if len(prof.Gases) > 0 {
			parts := []string{base}
			for _, g := range prof.Gases {
				parts = append(parts, fmt.Sprintf("%s-%02d", g.Name, g.PercentBP/100))
			}

			base = strings.Join(parts, ":")
		}

		return appendTaintSuffix(base, atmo.Taints, ":")
	}

	// Exotic / Corrosive / Insidious
	subtypeWithHazard := atmo.Subtype
	if atmo.Code == 12 && len(atmo.InsidiousHazards) > 0 {
		var codes strings.Builder
		for _, h := range atmo.InsidiousHazards {
			codes.WriteString(h.Code)
		}

		subtypeWithHazard = atmo.Subtype + "." + codes.String()
	}

	base := fmt.Sprintf("%s-St%s", codeChar, subtypeWithHazard)
	if subtypeWithHazard == "" {
		// Any code reaching this branch with no subtype emits the bare code
		// char rather than a dangling "-St". Exotic (A, WBH p.85), Corrosive
		// (B) / Insidious (C, p.89), and most Unusual (F, p.93) rolls now
		// populate Subtype and keep the "-St" form. The empty-subtype case
		// remains for None (0) / Trace (1), which have no subtype at all, and
		// for the Unusual "Other" row (p.93 D26=26), which is intentionally
		// recorded as an empty subtype so it renders as the bare code "F".
		base = codeChar
	}

	if atmo.Pressure > 0 {
		base += fmt.Sprintf(":%.2f", atmo.Pressure)
	}

	if len(prof.Gases) > 0 {
		parts := []string{base}
		for _, g := range prof.Gases {
			parts = append(parts, fmt.Sprintf("%s-%02d", g.Name, g.PercentBP/100))
		}

		base = strings.Join(parts, ":")
	}

	return appendTaintSuffix(base, atmo.Taints, " ")
}

// appendTaintSuffix appends T.S.P or I.S.P entries comma-separated,
// joined to base with sep. Returns base unchanged when taints is empty.
func appendTaintSuffix(base string, taints []Taint, sep string) string {
	if len(taints) == 0 {
		return base
	}

	parts := make([]string, 0, len(taints))
	for _, t := range taints {
		parts = append(parts, fmt.Sprintf("%s.%d.%d", t.Code, t.Severity, t.Persistence))
	}

	return base + sep + strings.Join(parts, ",")
}
