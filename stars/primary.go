package stars

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
)

// ErrSpecialPrimary is returned by RollPrimaryTypeAndClass when the
// initial Type-column roll is 2 ("Special"). Callers should dispatch
// through peculiar.go to resolve the special-object branch. When the
// caller has no dispatch implementation (e.g. companion-side), it
// bubbles up and is matched by errors.Is(err, ErrSpecialCircumstances)
// since the special-companion dispatch belongs to Special Circumstances.
var ErrSpecialPrimary = fmt.Errorf(
	"stars: special primary; dispatch through peculiar: %w",
	ErrSpecialCircumstances,
)

var validLetters = map[SpectralLetter]struct{}{
	'O': {}, 'B': {}, 'A': {}, 'F': {}, 'G': {}, 'K': {}, 'M': {},
}

// RollPrimaryTypeAndClass rolls for the spectral letter and luminosity
// class of a primary star (WBH p.15) with no DM.
//
// A roll of 12 redirects to the Hot column with a fresh 2D roll.
// A roll of 2 returns ErrSpecialPrimary so the caller can route through
// the Special / Unusual / Peculiar dispatch in peculiar.go.
func RollPrimaryTypeAndClass(r roller.Roller) (SpectralLetter, LuminosityClass, error) {
	return rollPrimaryTypeAndClass(r, 0)
}

// RollPrimaryTypeAndClassDMPlus1 is the class-redirect variant per WBH
// p.16: "Any result starting with Class requires a second roll on the
// Type column with DM+1." The +1 makes the "Special" cell (row 2)
// unreachable, so the redirect cannot recursively land back on Special.
func RollPrimaryTypeAndClassDMPlus1(r roller.Roller) (SpectralLetter, LuminosityClass, error) {
	return rollPrimaryTypeAndClass(r, 1)
}

func rollPrimaryTypeAndClass(r roller.Roller, dm int) (SpectralLetter, LuminosityClass, error) {
	first := min(r.Roll("2D")+dm, 12)

	row, ok := StarTypeDetermination[first]
	if !ok {
		return 0, "", fmt.Errorf("stars: 2D out of range: %d", first)
	}

	cell := row.Type

	if cell == "Hot" {
		second := r.Roll("2D")

		hotRow, ok := StarTypeDetermination[second]
		if !ok {
			return 0, "", fmt.Errorf("stars: 2D out of range: %d", second)
		}

		cell = hotRow.Hot
	}

	if cell == "Special" {
		return 0, "", ErrSpecialPrimary
	}

	if len(cell) != 1 {
		return 0, "", fmt.Errorf("stars: unexpected primary type cell: %q", cell)
	}

	letter := SpectralLetter(cell[0])
	if _, ok := validLetters[letter]; !ok {
		return 0, "", fmt.Errorf("stars: invalid letter from table: %c", letter)
	}

	return letter, V, nil
}

// RollSubtype rolls Star Subtype (WBH p. 16) and returns 0-9.
//
// Primary M-type stars use the M column; all others use the Numeric
// column. K-type Class IV stars have any subtype > 4 reduced by 5.
func RollSubtype(r roller.Roller, letter SpectralLetter, lc LuminosityClass) (int, error) {
	if _, ok := validLetters[letter]; !ok {
		return 0, fmt.Errorf("stars: invalid letter: %c", letter)
	}

	roll := r.Roll("2D")

	var sub int

	if letter == 'M' {
		s, ok := StarSubtypeMType[roll]
		if !ok {
			return 0, fmt.Errorf("stars: 2D out of range: %d", roll)
		}

		sub = s
	} else {
		s, ok := StarSubtypeNumeric[roll]
		if !ok {
			return 0, fmt.Errorf("stars: 2D out of range: %d", roll)
		}

		sub = s
	}

	if letter == 'K' && lc == IV && sub > 4 {
		sub -= 5
	}

	return sub, nil
}

// ApplyClassIVLetterConstraint maps a spectral letter for Class IV
// limits (WBH p. 16). M -> K (subtype shift handled in RollSubtype),
// O -> B (for Hot rolls). Other letters pass through.
func ApplyClassIVLetterConstraint(letter SpectralLetter) SpectralLetter {
	switch letter {
	case 'M':
		return 'K'
	case 'O':
		return 'B'
	default:
		return letter
	}
}

// ApplyClassVILetterConstraint maps a spectral letter for Class VI
// (WBH p. 16): F -> G, A -> B. Other letters pass through.
func ApplyClassVILetterConstraint(letter SpectralLetter) SpectralLetter {
	switch letter {
	case 'F':
		return 'G'
	case 'A':
		return 'B'
	default:
		return letter
	}
}

// RollGiantClass rolls the Giants column with DM+1 to determine the
// final luminosity class for a Class III+ result (WBH p. 16).
func RollGiantClass(r roller.Roller) (LuminosityClass, error) {
	natural := r.Roll("2D")
	row := min(natural+1, 12)

	rowData, ok := StarTypeDetermination[row]
	if !ok {
		return "", fmt.Errorf("stars: 2D out of range: %d", row)
	}

	switch rowData.Giants {
	case "Class III":
		return III, nil
	case "Class II":
		return II, nil
	case "Class Ib":
		return Ib, nil
	case "Class Ia":
		return Ia, nil
	default:
		return "", fmt.Errorf("stars: unexpected Giants cell: %q", rowData.Giants)
	}
}
