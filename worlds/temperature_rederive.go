// Package worlds — atmosphere/hydrographics re-derivation under real
// temperature per WBH p.79, p.81, pp.94-98, p.99, p.102 (sub-project 3A2b-rederive).
package worlds

// MeanKToTempRange buckets a real mean temperature in Kelvin into the same
// TempRange bands 3A1's HZCOOffsetToTempRange used (WBH pp.94-98 keying):
//
//	≥ 453K → Boiling
//	353-453K → Hot
//	273-353K → Temperate
//	123-273K → Cold
//	< 123K → Frozen
func MeanKToTempRange(meanK float64) TempRange {
	switch {
	case meanK >= 453:
		return TempBoiling
	case meanK >= 353:
		return TempHot
	case meanK >= 273:
		return TempTemperate
	case meanK >= 123:
		return TempCold
	default:
		return TempFrozen
	}
}
