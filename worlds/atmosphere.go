package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
)

// Atmosphere — surface atmosphere characteristics per WBH pp.79-91.
//
// Pressure, ScaleHeight, Subtype, and Profile are populated by 3A1 with
// HZCO-bucketed provisional temperature; ApplyClimate (Stage 5) re-derives
// these fields under the real Temperature.MeanK. Post-rederive values are
// final for those fields.
//
// Taints and InsidiousHazards are populated by ApplyTaintTypology (Stage 6)
// per WBH pp.81-90. Taints contains 0-3 entries; InsidiousHazards is
// non-empty only for atm C (Insidious) — typically 1 hazard, but 2 when
// the subtype is D or E (auto-T plus rolled, per the p.90 footnote).
type Atmosphere struct {
	Code                  int
	Subtype               string
	Pressure              float64
	OxygenPartialPressure float64
	ScaleHeight           float64
	Profile               AtmosphereProfile
	Taints                []Taint
	InsidiousHazards      []Hazard
}

// AtmosphereProfile holds the gas-mix composition detail populated in Task 12.
type AtmosphereProfile struct {
	TempRange string
	Gases     []GasFraction
	Shorthand string
}

// GasFraction is one constituent gas in an AtmosphereProfile.
type GasFraction struct {
	Name      string
	PercentBP int
}

// TempRange is a provisional temperature class derived from a body's HZCO offset
// per WBH pp.94-98.
type TempRange int

// Temperature range constants keyed to HZCO offset bands per WBH pp.94-98.
const (
	TempBoiling   TempRange = iota // offset ≤ -2.01  (mean ≥ 453 K)
	TempHot                        // offset -2.0..-1.01 (353-453 K)
	TempTemperate                  // offset -1.0..+1.0  (273-353 K)
	TempCold                       // offset +1.01..+3.0 (123-273 K)
	TempFrozen                     // offset ≥ +3.01     (≤ 123 K)
)

// HZCOOffsetToTempRange converts a body's orbital position relative to its star's
// habitable-zone centre orbit (HZCO) into a temperature class per WBH pp.94-98.
func HZCOOffsetToTempRange(orbitNumber, hzco float64) TempRange {
	offset := orbitNumber - hzco
	switch {
	case offset <= -2.01:
		return TempBoiling
	case offset <= -1.01:
		return TempHot
	case offset <= 1.0:
		return TempTemperate
	case offset <= 3.0:
		return TempCold
	default:
		return TempFrozen
	}
}

// atmosphereCodes is the single transcription of the WBH p.79
// Atmosphere Codes table: UWP hex character and label per code 0-17.
// Every code→char and code→label lookup in the package derives from
// this table — do not re-transcribe it (three independent copies once
// drifted; see docs/wbh-inconsistencies.md § atmosphere code labels).
var atmosphereCodes = map[int]struct{ Char, Name string }{
	0:  {"0", "None"},
	1:  {"1", "Trace"},
	2:  {"2", "Very Thin, Tainted"},
	3:  {"3", "Very Thin"},
	4:  {"4", "Thin, Tainted"},
	5:  {"5", "Thin"},
	6:  {"6", "Standard"},
	7:  {"7", "Standard, Tainted"},
	8:  {"8", "Dense"},
	9:  {"9", "Dense, Tainted"},
	10: {"A", "Exotic"},
	11: {"B", "Corrosive"},
	12: {"C", "Insidious"},
	13: {"D", "Very Dense"},
	14: {"E", "Low"},
	15: {"F", "Unusual"},
	16: {"G", "Gas, Helium"},
	17: {"H", "Gas, Hydrogen"},
}

// AtmosphereCompositionLabel returns the human-readable label for a WBH p.79
// atmosphere code. Returns "" for unknown codes.
func AtmosphereCompositionLabel(code int) string {
	return atmosphereCodes[code].Name
}

// SizeAsInt converts a SizeCode to its integer equivalent for arithmetic.
// "S", "R", "", "0" → 0; "1"-"9" → 1-9; "A"-"F" → 10-15.
func SizeAsInt(s SizeCode) int {
	switch s {
	case "", "0", "S", "R":
		return 0
	}
	if len(s) != 1 {
		return 0
	}
	ch := s[0]
	switch {
	case ch >= '1' && ch <= '9':
		return int(ch - '0')
	case ch >= 'A' && ch <= 'F':
		return int(ch-'A') + 10
	}
	return 0
}

// AtmospherePressureRange returns (minBar, spanBar) for the given atmosphere code per WBH p.79.
// Codes with "Varies" pressure (A/B/C/F/G/H = 10/11/12/15/16/17) return (0, 0).
func AtmospherePressureRange(code int) (minBar, spanBar float64) {
	switch code {
	case 0:
		return 0, 0.0009
	case 1:
		return 0.001, 0.089
	case 2, 3:
		return 0.1, 0.32
	case 4, 5:
		return 0.43, 0.27
	case 6, 7:
		return 0.7, 0.79
	case 8, 9:
		return 1.5, 0.99
	case 13:
		return 2.5, 7.5
	case 14:
		return 0.10, 0.32
	}
	return 0, 0
}

// corrosiveInsidiousPressureRange returns (minBar, spanBar) for atm
// codes B (11) and C (12) keyed off the WBH p.89 subtype letter.
//
// Subtypes 1-B carry the explicit ranges from the p.89 table.
// Subtypes C/D/E carry "10.0+ / unbound" in the book; we return
// project-supplied tiered ranges per the design spec dated 2026-05-09
// (atm-bc-pressure-from-subtype). The tiers apply uniformly to atm B
// and atm C — both share the p.89 subtype table — but per-code
// distribution differs in practice: insidious (atm C) gets DM+2 on
// the subtype roll (WBH p.89 footnote), so it lands subtype E and the
// 1000+ bar tier far more often than corrosive (atm B). This matches
// p.89's "Only insidious extremely dense atmospheres should have
// pressures exceeding 1,000 bar" expectation statistically; corrosive
// atmospheres can technically reach the same tier when subtype E is
// rolled, which is rare without the DM+2.
//
//	C: 10–100      (min=10,   span=90)
//	D: 100–1000    (min=100,  span=900)
//	E: 1000–10000  (min=1000, span=9000)
//
// Empty/unknown subtype returns (0, 0). Callers in the live pipeline
// always roll the subtype before pressure (see system_detail_steps.go
// and temperature_rederive.go).
func corrosiveInsidiousPressureRange(subtype string) (minBar, spanBar float64) {
	switch subtype {
	case "1", "2", "3":
		return 0.1, 0.32
	case "4", "5":
		return 0.43, 0.27
	case "6", "7":
		return 0.70, 0.79
	case "8", "9":
		return 1.50, 0.99
	case "A", "B":
		return 2.50, 7.50
	case "C":
		return 10, 90
	case "D":
		return 100, 900
	case "E":
		return 1000, 9000
	}
	return 0, 0
}

// RollTotalPressure computes total atmospheric pressure per WBH p.80,
// or per the WBH p.89 subtype-keyed range for atm B/C:
//
//	bar = MinPressureRange + Span × ((1D-1)×5 + (1D-1)) / 30
//
// For atm codes 11 (B) and 12 (C), subtype must be set (one of
// "1"-"9", "A"-"E"); otherwise pressure falls back to 0 (legacy
// "Varies" behavior). For all other codes, subtype is ignored.
//
// Returns minBar with no rolls consumed when span = 0.
func RollTotalPressure(r roller.Roller, atmoCode int, subtype string) (float64, error) {
	var minBar, span float64
	if (atmoCode == 11 || atmoCode == 12) && subtype != "" {
		minBar, span = corrosiveInsidiousPressureRange(subtype)
	} else {
		minBar, span = AtmospherePressureRange(atmoCode)
	}
	if span == 0 {
		return minBar, nil
	}
	a := r.Roll("1D")
	b := r.Roll("1D")
	scale := float64((a-1)*5+(b-1)) / 30.0
	return minBar + span*scale, nil
}

// RollOxygenFraction computes O2 fraction per WBH p.81:
//
//	OxyFraction = (1D + ageDM)/20 + (2D-7)/100 + (1D-1)/20
//
// Floored at 0.
func RollOxygenFraction(r roller.Roller, ageGyr float64) (float64, error) {
	dm := ageDMForOxygen(ageGyr)
	a := r.Roll("1D")
	b := r.Roll("2D")
	c := r.Roll("1D")
	frac := float64(a+dm)/20 + float64(b-7)/100 + float64(c-1)/20
	if frac < 0 {
		frac = 0
	}
	return frac, nil
}

// ageDMForOxygen returns the DM applied to the 1D oxygen fraction roll per WBH p.82.
// Ranges: >4 Gyr → +1; 3.0–3.5 Gyr → -1; 2.0–3.0 Gyr → -2; <2 Gyr → -4.
// Ages in the 3.5–4.0 band return 0 (gap in the book's optional DMs table).
func ageDMForOxygen(ageGyr float64) int {
	switch {
	case ageGyr > 4:
		return 1
	case ageGyr >= 3.5:
		return 0
	case ageGyr >= 3:
		return -1
	case ageGyr >= 2:
		return -2
	default:
		return -4
	}
}

// DeriveScaleHeight returns the atmospheric scale height in km per WBH p.81
// approximation: H ≈ 8.5 × (T/288) / g.
// Returns 0 if g ≤ 0.
func DeriveScaleHeight(meanTempK, gravityG float64) float64 {
	if gravityG <= 0 {
		return 0
	}
	return 8.5 * (meanTempK / 288) / gravityG
}

// RollCorrosiveInsidiousSubtype rolls on WBH p.89 Corrosive and Insidious Atmosphere Subtype table.
//
// DMs (WBH p.89):
//   - Size 2-4 → DM-3
//   - Size 8+  → DM+2
//   - Orbit < HZCO-1 → DM+4
//   - Orbit > HZCO+2 → DM-2
//   - Atmosphere is Insidious (code 12) → DM+2
//   - Runaway greenhouse result → DM+4
//
// Returns subtype letter "1"-"9", "A"-"E".
func RollCorrosiveInsidiousSubtype(
	r roller.Roller,
	sizeCode SizeCode,
	orbit, hzco float64,
	isInsidious, runawayResult bool,
) (string, error) {
	roll := r.Roll("2D")
	dm := 0
	si := SizeAsInt(sizeCode)
	if si >= 2 && si <= 4 {
		dm -= 3
	}
	if si >= 8 {
		dm += 2
	}
	if orbit < hzco-1 {
		dm += 4
	}
	if orbit > hzco+2 {
		dm -= 2
	}
	if isInsidious {
		dm += 2
	}
	if runawayResult {
		dm += 4
	}
	total := roll + dm
	switch {
	case total <= 1:
		return "1", nil
	case total <= 9:
		return fmt.Sprintf("%d", total), nil
	case total == 10:
		return "A", nil
	case total == 11:
		return "B", nil
	case total == 12:
		return "C", nil
	case total == 13:
		return "D", nil
	default:
		return "E", nil
	}
}

// HZRegionAtmosphereDM returns the WBH p.47 Habitable Zones Regions
// atmosphere DM for the given atmosphere code. Used by both the atmosphere
// generation (3A1) and the basic temperature roll (3A2b-temp p.109).
//
// Verified against the table on WBH p.47:
//
//	0, 1     → 0  (no modifier; temperature swings between day and night)
//	2, 3     → -2
//	4, 5, E  → -1
//	6, 7     → 0
//	8, 9     → +1
//	A, D, F  → +2
//	B, C     → +6
//
// Codes outside the table (G, H, others) return 0 (defensive).
func HZRegionAtmosphereDM(code int) int {
	switch code {
	case 0, 1:
		return 0
	case 2, 3:
		return -2
	case 4, 5, 14: // 4, 5, E
		return -1
	case 6, 7:
		return 0
	case 8, 9:
		return 1
	case 10, 13, 15: // A, D, F
		return 2
	case 11, 12: // B, C
		return 6
	default:
		return 0
	}
}

// RollAtmoCode rolls the unified WBH atmosphere digit formula: 2D-7+Size.
//
// Sizes 0, 1, and S return automatic atmo code 0 without consuming a roll.
// Results are clamped to the world-atmosphere range 0-15 (0-F): codes 16
// (G) and 17 (H) are gas-giant rows of the p.79 table and must never be
// produced for worlds — docs/wbh-inconsistencies.md documents "RollAtmoCode
// cannot produce them" as a package invariant, which the upper clamp
// enforces for SizeCode B-F worlds (e.g. Size F: 2D=12 → 12-7+15 = 20
// unclamped).
//
// The third argument (hzcoOffset) is reserved for non-HZ Hot/Cold table column
// selection in Task 11; it does not affect the roll formula here.
func RollAtmoCode(r roller.Roller, sizeCode SizeCode, _ float64) (int, error) {
	if sizeCode == "0" || sizeCode == "1" || sizeCode == "S" {
		return 0, nil
	}
	roll := r.Roll("2D")
	code := min(max(roll-7+SizeAsInt(sizeCode), 0), 15)
	return code, nil
}
