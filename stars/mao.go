package stars

import (
	"errors"
	"fmt"
)

// IsPostStellar reports whether a StarKind is post-stellar (Brown Dwarf,
// White Dwarf, Neutron Star, Black Hole, Pulsar). Semantic classification
// shared by the WBH p.37 count DMs and the p.39 MAO gate; do not broaden
// it (to gate the MAO lookup use LacksP39MAORow, which also excludes
// protostars and aggregate objects).
func IsPostStellar(k StarKind) bool {
	switch k {
	case KindBrownDwarf, KindWhiteDwarf, KindNeutronStar, KindBlackHole, KindPulsar:
		return true
	}

	return false
}

// ErrPostStellarPrimaryUnsupported indicates the star has no p.39 MAO
// row — every post-stellar kind, the pre-stellar Protostar, and the
// aggregate/pre-stellar Nebula, Star Cluster, and Anomaly. MAO for these
// kinds lives in the WBH Special Circumstances chapter and is not yet
// encoded. (The name predates the non-post-stellar additions.) Wraps
// ErrSpecialCircumstances so callers can classify uniformly.
var ErrPostStellarPrimaryUnsupported = fmt.Errorf(
	"stars: post-stellar primary MAO requires Special Circumstances chapter: %w",
	ErrSpecialCircumstances,
)

// ErrNoMAOForStar reports a "—" cell in the p.39 MAO table — the
// spectral type / class combination does not exist as a star.
var ErrNoMAOForStar = errors.New("stars: spectral type / class combination has no MAO entry")

// LacksP39MAORow reports whether a StarKind has no row in the WBH p.39
// Minimum Allowable Orbit# table — every post-stellar kind, the
// pre-stellar Protostar, and the aggregate/pre-stellar Nebula, Star
// Cluster, and Anomaly. MAO for these kinds lives in the Special
// Circumstances chapter (not yet encoded).
//
// The aggregate kinds (IsAggregateObject) are included because
// specialObjectAge lets them survive the age step with AgeGyr 0; without
// this gate they reach maoCell and raise a raw table-miss error instead
// of the classifiable ErrPostStellarPrimaryUnsupported.
func LacksP39MAORow(k StarKind) bool {
	if k == KindProtostar || IsAggregateObject(k) {
		return true
	}

	return IsPostStellar(k)
}

// maoRow is one row of the WBH p.39 Minimum Allowable Orbit# table,
// keyed by luminosity class.
type maoRow struct {
	Ia, Ib, II, III, IV, V, VI *float64
}

// maoTablePage39 is the WBH p.39 MAO table, keyed by the grid spectral
// types ("O0", "O5", "B0", ..., "K5", "M0", "M5", "M9"). Off-grid types
// (e.g. G7) are interpolated by bracketSpectralType, not looked up here.
//
// nil pointer means the book leaves the cell as "—" (combination does
// not exist as a star).
var maoTablePage39 = map[string]maoRow{
	"O0": {Ia: new(0.63), Ib: new(0.60), II: new(0.55), III: new(0.53), V: new(0.5), VI: new(0.01)},
	"O5": {Ia: new(0.55), Ib: new(0.50), II: new(0.45), III: new(0.38), V: new(0.3), VI: new(0.01)},
	"B0": {Ia: new(0.50), Ib: new(0.35), II: new(0.30), III: new(0.25), IV: new(0.20), V: new(0.18), VI: new(0.01)},
	"B5": {Ia: new(1.67), Ib: new(0.63), II: new(0.35), III: new(0.15), IV: new(0.13), V: new(0.09), VI: new(0.01)},
	"A0": {Ia: new(3.34), Ib: new(1.40), II: new(0.75), III: new(0.13), IV: new(0.10), V: new(0.06)},
	"A5": {Ia: new(4.17), Ib: new(2.17), II: new(1.17), III: new(0.13), IV: new(0.07), V: new(0.05)},
	"F0": {Ia: new(4.42), Ib: new(2.50), II: new(1.33), III: new(0.13), IV: new(0.07), V: new(0.04)},
	"F5": {Ia: new(5.00), Ib: new(3.25), II: new(1.87), III: new(0.13), IV: new(0.06), V: new(0.03)},
	"G0": {Ia: new(5.21), Ib: new(3.59), II: new(2.24), III: new(0.25), IV: new(0.07), V: new(0.03), VI: new(0.02)},
	"G5": {Ia: new(5.34), Ib: new(3.84), II: new(2.67), III: new(0.38), IV: new(0.10), V: new(0.02), VI: new(0.02)},
	"K0": {Ia: new(5.59), Ib: new(4.17), II: new(3.17), III: new(0.50), IV: new(0.15), V: new(0.02), VI: new(0.02)},
	"K5": {Ia: new(6.17), Ib: new(4.84), II: new(4.00), III: new(1.00), V: new(0.02), VI: new(0.01)},
	"M0": {Ia: new(6.80), Ib: new(5.42), II: new(4.59), III: new(1.68), V: new(0.02), VI: new(0.01)},
	"M5": {Ia: new(7.20), Ib: new(6.17), II: new(5.30), III: new(3.00), V: new(0.01), VI: new(0.01)},
	"M9": {Ia: new(7.80), Ib: new(6.59), II: new(5.92), III: new(4.34), V: new(0.01), VI: new(0.01)},
}

// maoCell reads the MAO cell for an exact spectral type key (e.g. "G5")
// at a given luminosity class. Returns ErrNoMAOForStar if the cell is
// the book's "—".
func maoCell(typeKey string, lc LuminosityClass) (float64, error) {
	row, ok := maoTablePage39[typeKey]
	if !ok {
		return 0, fmt.Errorf("stars: no MAO row for %q", typeKey)
	}

	var ptr *float64

	switch lc {
	case Ia:
		ptr = row.Ia
	case Ib:
		ptr = row.Ib
	case II:
		ptr = row.II
	case III:
		ptr = row.III
	case IV:
		ptr = row.IV
	case V:
		ptr = row.V
	case VI:
		ptr = row.VI
	default:
		return 0, fmt.Errorf("stars: unknown luminosity class %q", lc)
	}

	if ptr == nil {
		return 0, ErrNoMAOForStar
	}

	return *ptr, nil
}

// MAO returns the Minimum Allowable Orbit# for a star, interpolated by
// spectral type within its luminosity-class column per the WBH p.39
// table.
//
// Kinds without a p.39 row (post-stellar, plus protostar and aggregate
// objects) return ErrPostStellarPrimaryUnsupported. Combinations the
// book lists as "—" return ErrNoMAOForStar.
func MAO(s Star) (float64, error) {
	if LacksP39MAORow(s.Kind) {
		return 0, ErrPostStellarPrimaryUnsupported
	}

	lower, upper, frac := bracketSpectralType(s.SpectralType)
	lo, errLo := maoCell(lower, s.LuminosityClass)

	hi, errHi := maoCell(upper, s.LuminosityClass)
	switch {
	case errors.Is(errLo, ErrNoMAOForStar) && errors.Is(errHi, ErrNoMAOForStar):
		return 0, ErrNoMAOForStar
	case errors.Is(errLo, ErrNoMAOForStar):
		return hi, nil
	case errors.Is(errHi, ErrNoMAOForStar):
		return lo, nil
	case errLo != nil:
		return 0, errLo
	case errHi != nil:
		return 0, errHi
	}

	return lo + (hi-lo)*frac, nil
}

// bracketSpectralType returns the two p.39 grid keys bracketing st
// within st's letter, and the fractional position from lower to upper
// (0.0 at lower, 1.0 at upper).
//
// At exact grid points (O0, O5, ..., K5, M0, M5, M9) the function
// returns frac=0, so the upper key is unused in interpolation. Only
// M5 and M9 (the table's terminal rows) return lower == upper.
func bracketSpectralType(st SpectralType) (lower, upper string, frac float64) {
	letter := string(st.Letter)
	switch {
	case st.Letter == 'M' && st.Subtype >= 5:
		// Bracket M5 → M9; subtypes 5..9.
		if st.Subtype <= 5 {
			return "M5", "M5", 0
		}

		if st.Subtype >= 9 {
			return "M9", "M9", 0
		}

		return "M5", "M9", float64(st.Subtype-5) / 4.0
	case st.Subtype < 5:
		lower = letter + "0"
		upper = letter + "5"

		return lower, upper, float64(st.Subtype) / 5.0
	default:
		// st.Subtype in [5, 9] but letter != M (handled above).
		// Bracket Letter5 → NextLetter0.
		next := nextCoolerLetter(st.Letter)
		lower = letter + "5"
		upper = string(next) + "0"

		return lower, upper, float64(st.Subtype-5) / 5.0
	}
}
