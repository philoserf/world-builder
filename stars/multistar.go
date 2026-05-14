package stars

import (
	"errors"
	"fmt"

	"github.com/philoserf/world-builder/roller"
)

// OrbitClass is the WBH p.23 orbit class for a non-primary star.
type OrbitClass string

// OrbitClass constants (WBH p.23).
const (
	OrbitClose     OrbitClass = "close"
	OrbitNear      OrbitClass = "near"
	OrbitFar       OrbitClass = "far"
	OrbitCompanion OrbitClass = "companion"
)

// PresenceDM returns the WBH p.23 DM applied to all multi-star presence
// rolls for a system whose primary has the given properties.
func PresenceDM(primary Star) int {
	// Special-object primaries: BD, D, and post-stellar all -1.
	switch primary.Kind {
	case KindBrownDwarf, KindWhiteDwarf, KindPulsar, KindNeutronStar, KindBlackHole:
		return -1
	}
	// Class Ia/Ib/II/III/IV: +1.
	switch primary.LuminosityClass {
	case Ia, Ib, II, III, IV:
		return 1
	}
	// Class V or VI primary by spectral letter:
	switch primary.LuminosityClass {
	case V, VI:
		switch primary.SpectralType.Letter {
		case 'O', 'B', 'A', 'F':
			return 1
		case 'M':
			return -1
		}
	}
	return 0
}

// RollPresence rolls 2D + PresenceDM(primary) and returns true if the
// result meets the WBH p.23 threshold (10+).
//
// Class Ia/Ib/II/III primaries cannot have Close secondaries (WBH p.23);
// in those cases the function returns false without consuming a roll.
func RollPresence(r roller.Roller, primary Star, oc OrbitClass) bool {
	if oc == OrbitClose {
		switch primary.LuminosityClass {
		case Ia, Ib, II, III:
			return false
		}
	}
	natural := r.Roll("2D")
	return natural+PresenceDM(primary) >= MultipleStarsPresenceThreshold
}

// NonPrimaryRole names the relationship of a non-primary star to its parent.
type NonPrimaryRole int

// NonPrimaryRole constants (WBH p.29 column selectors).
const (
	RoleSecondary   NonPrimaryRole = iota // secondary orbit column
	RoleCompanion                         // companion orbit column
	RolePostStellar                       // post-stellar column
	RoleOther                             // Other column (yields "D" or "BD")
)

// RollNonPrimaryDescriptor rolls 2D into NonPrimaryStarDetermination
// (with DM-1 if the parent is Class III or IV) and returns the cell for
// the requested role: one of "Random", "Lesser", "Sibling", "Twin",
// "Other", "D", or "BD" (WBH p.29).
func RollNonPrimaryDescriptor(r roller.Roller, parent Star, role NonPrimaryRole) string {
	natural := r.Roll("2D")
	if parent.LuminosityClass == III || parent.LuminosityClass == IV {
		natural--
	}
	natural = max(2, min(12, natural))
	row := NonPrimaryStarDetermination[natural]
	switch role {
	case RoleSecondary:
		return row.Secondary
	case RoleCompanion:
		return row.Companion
	case RolePostStellar:
		return row.PostStellar
	case RoleOther:
		return row.Other
	}
	return ""
}

// GenerateCompanionStar produces a star given its parent and the
// descriptor returned by RollNonPrimaryDescriptor (WBH p.29).
//
// Descriptors:
//   - "Twin": same class, type, subtype as parent (variance off; Plan 2
//     does not apply the optional 1D-1% variance).
//   - "Sibling": parent_subtype + 1D. If subtype goes past 9, advance to
//     the next-cooler letter and subtract 10 (e.g. parent G7 + 1D=5 -> K2).
//     Note: the book text on p.29 says "Subtract 1D from the subtype" but
//     the worked example on p.30 (Zed Ab) computes G7 + 1=G8 — i.e. ADDS 1D.
//     We follow the worked example.
//   - "Lesser": one type cooler (F->G, G->K, K->M); for primary M-types
//     the lesser is also M but with subtype rerolled. Subtype rerolled
//     via RollSubtype on the new letter.
//   - "Random": roll on the regular Star Type Determination table; if
//     the result is hotter than the parent, treat as Lesser instead.
//   - "Other": for the Plan 2 simple path, return the parent's
//     descriptor on the Other column (a fresh roll on the row's "Other"
//     entry, which is "D" or "BD"). The caller passes this through.
//   - "D": white dwarf. Mass via book's white-dwarf approach; for Plan 2
//     return a Star with Kind=KindWhiteDwarf, LuminosityClass=D, and
//     placeholder physical values that the caller fills in later. Mass
//     0.6 (typical), Diameter 0.01, Temperature 8000 — book-typical
//     values; the caller (Task 11/14) overrides with full computation.
//   - "BD": brown dwarf. Kind=KindBrownDwarf, LuminosityClass=BD.
//     Placeholder mass 0.05, diameter 0.1, temperature 1500.
//
// All companion stars get their physical quantities fully recomputed
// downstream (Task 10's GenerateSystem), so the values returned here for
// non-stellar (D / BD) cases are stand-ins.
func GenerateCompanionStar(r roller.Roller, parent Star, descriptor string) (Star, error) {
	switch descriptor {
	case "Twin":
		return generateTwin(parent), nil
	case "Sibling":
		return generateSibling(r, parent)
	case "Lesser":
		return generateLesser(r, parent)
	case "Random":
		return generateRandom(r, parent)
	case "Other":
		// "Other" requires the caller to make a second roll on the Other
		// column and re-dispatch. Surface the contract: return an error
		// describing what the caller should do next.
		return Star{}, fmt.Errorf("stars: GenerateCompanionStar(\"Other\"): caller must reroll on Other column")
	case "D":
		return generateWhiteDwarfStub(parent), nil
	case "BD":
		return generateBrownDwarfStub(parent), nil
	default:
		return Star{}, fmt.Errorf("stars: unknown descriptor: %q", descriptor)
	}
}

func generateTwin(parent Star) Star {
	return Star{
		Kind:            parent.Kind,
		SpectralType:    parent.SpectralType,
		LuminosityClass: parent.LuminosityClass,
		Mass:            parent.Mass,
		Diameter:        parent.Diameter,
		Temperature:     parent.Temperature,
		Luminosity:      parent.Luminosity,
		AgeGyr:          parent.AgeGyr,
	}
}

// letterCoolerOrder lists spectral letters from hottest to coolest.
var letterCoolerOrder = []SpectralLetter{'O', 'B', 'A', 'F', 'G', 'K', 'M'}

// nextCoolerLetter returns the next-cooler spectral letter, or 0 if the
// input is already 'M' (no cooler).
func nextCoolerLetter(letter SpectralLetter) SpectralLetter {
	for i, l := range letterCoolerOrder {
		if l == letter && i+1 < len(letterCoolerOrder) {
			return letterCoolerOrder[i+1]
		}
	}
	return 0
}

func generateSibling(r roller.Roller, parent Star) (Star, error) {
	delta := r.Roll("1D")
	letter := parent.SpectralType.Letter
	subtype := parent.SpectralType.Subtype + delta
	for subtype > 9 {
		next := nextCoolerLetter(letter)
		if next == 0 {
			// Already at M9; clamp.
			subtype = 9
			break
		}
		letter = next
		subtype -= 10
	}
	// WBH p.16 letter/subtype constraints. generateSibling derives
	// subtype by addition rather than RollSubtype, so the constraints
	// that RollSubtype applies internally need to be replayed here.
	switch parent.LuminosityClass {
	case IV:
		letter = ApplyClassIVLetterConstraint(letter)
		if letter == 'K' && subtype > 4 {
			subtype -= 5
		}
	case VI:
		letter = ApplyClassVILetterConstraint(letter)
	}
	st := SpectralType{Letter: letter, Subtype: subtype}
	out := Star{
		Kind:            parent.Kind,
		SpectralType:    st,
		LuminosityClass: parent.LuminosityClass,
		AgeGyr:          parent.AgeGyr,
	}
	// Physical quantities: caller will fill in via the standard pipeline.
	return out, nil
}

func generateLesser(r roller.Roller, parent Star) (Star, error) {
	letter := nextCoolerLetter(parent.SpectralType.Letter)
	if letter == 0 {
		// Parent is already M; lesser stays M with subtype rerolled.
		letter = 'M'
	}
	switch parent.LuminosityClass {
	case IV:
		letter = ApplyClassIVLetterConstraint(letter)
	case VI:
		letter = ApplyClassVILetterConstraint(letter)
	}
	subtype, err := RollSubtype(r, letter, parent.LuminosityClass)
	if err != nil {
		return Star{}, err
	}
	return Star{
		Kind:            parent.Kind,
		SpectralType:    SpectralType{Letter: letter, Subtype: subtype},
		LuminosityClass: parent.LuminosityClass,
		AgeGyr:          parent.AgeGyr,
	}, nil
}

func generateRandom(r roller.Roller, parent Star) (Star, error) {
	letter, lc, err := RollPrimaryTypeAndClass(r)
	if errors.Is(err, ErrSpecialPrimary) {
		// Companion-side Special-primary dispatch. WBH p.15's Special
		// vs Unusual column choice applies to companions too — for now,
		// default to Special (the same as primaries). Producing
		// Unusual-column companions (BD/D/Peculiar via descriptor)
		// is what WBH p.29's NonPrimaryStarDetermination descriptors
		// already cover; the Type-column "Special" cell here is a
		// separate redirect to a non-V luminosity class.
		return generateRandomSpecial(r, parent)
	}
	if err != nil {
		return Star{}, err
	}
	// If hotter than parent, treat as Lesser instead.
	if isHotterOrEqualLetter(letter, parent.SpectralType.Letter) {
		return generateLesser(r, parent)
	}
	switch lc {
	case IV:
		letter = ApplyClassIVLetterConstraint(letter)
	case VI:
		letter = ApplyClassVILetterConstraint(letter)
	}
	subtype, err := RollSubtype(r, letter, lc)
	if err != nil {
		return Star{}, err
	}
	return Star{
		Kind:            parent.Kind,
		SpectralType:    SpectralType{Letter: letter, Subtype: subtype},
		LuminosityClass: lc,
		AgeGyr:          parent.AgeGyr,
	}, nil
}

// generateRandomSpecial handles the companion-side equivalent of
// generateSpecialPrimary: when the Type-column roll yields "Special"
// (2D=2), dispatch through the Special column to a class-redirect cell
// (Class VI / IV / III / Giants), then roll the resulting class's
// letter and subtype. WBH p.15 — Referee option; this follows the
// same Special-column default chosen for primaries. The returned star
// carries the parent's age and Kind derived from the rolled class.
func generateRandomSpecial(r roller.Roller, parent Star) (Star, error) {
	kind, lc, err := RollSpecialPrimary(r, PeculiarPathSpecial)
	if err != nil {
		return Star{}, err
	}
	if lc == "Giants" {
		giantClass, gerr := RollGiantClass(r)
		if gerr != nil {
			return Star{}, gerr
		}
		lc = giantClass
	}
	if lc == "" {
		// Special column never yields a final Kind (no BD/D/Peculiar
		// cells); this branch is unreachable today but guards against
		// future schema drift.
		return Star{}, fmt.Errorf("stars: Special column produced unexpected Kind %q", kind)
	}
	// Class-redirect: roll letter (DM+1 per WBH p.16), apply
	// class-letter constraint, then subtype.
	letter, _, err := RollPrimaryTypeAndClassDMPlus1(r)
	if err != nil {
		return Star{}, err
	}
	switch lc {
	case IV:
		letter = ApplyClassIVLetterConstraint(letter)
	case VI:
		letter = ApplyClassVILetterConstraint(letter)
	}
	subtype, err := RollSubtype(r, letter, lc)
	if err != nil {
		return Star{}, err
	}
	return Star{
		Kind:            kindForClass(lc),
		SpectralType:    SpectralType{Letter: letter, Subtype: subtype},
		LuminosityClass: lc,
		AgeGyr:          parent.AgeGyr,
	}, nil
}

func isHotterOrEqualLetter(a, b SpectralLetter) bool {
	idxA, idxB := -1, -1
	for i, l := range letterCoolerOrder {
		if l == a {
			idxA = i
		}
		if l == b {
			idxB = i
		}
	}
	if idxA == -1 || idxB == -1 {
		return false
	}
	return idxA <= idxB
}

func generateWhiteDwarfStub(parent Star) Star {
	return Star{
		Kind:            KindWhiteDwarf,
		LuminosityClass: D,
		Mass:            0.6,
		Diameter:        0.01,
		Temperature:     8000,
		AgeGyr:          parent.AgeGyr,
	}
}

func generateBrownDwarfStub(parent Star) Star {
	return Star{
		Kind:            KindBrownDwarf,
		LuminosityClass: BD,
		SpectralType:    SpectralType{Letter: 'M', Subtype: 9}, // book uses BD without spectral type; placeholder
		Mass:            0.05,
		Diameter:        0.1,
		Temperature:     1500,
		AgeGyr:          parent.AgeGyr,
	}
}

// AssignDesignations populates the Designation fields on a System
// according to WBH p.25.
//
// Primary is "A" (single) or "Aa" (with OrbitCompanion child).
// Close, Near, Far orbit classes are lettered B, C, D in presence order
// (absent classes are skipped — see Zed example p.34 where Close is
// absent and Near becomes "B"). Any star with its own OrbitCompanion
// child gets "+a" suffix; its companion gets "+b".
func AssignDesignations(sys *System) {
	// Primary first.
	primaryCompanionIdx := findCompanionByParent(sys, -1, OrbitCompanion)
	if primaryCompanionIdx >= 0 {
		sys.PrimaryDesignation = "Aa"
		sys.Companions[primaryCompanionIdx].Designation = "Ab"
	} else {
		sys.PrimaryDesignation = "A"
	}

	// Walk orbit classes in book order: Close, Near, Far.
	letters := []string{"B", "C", "D"}
	li := 0
	for _, oc := range []OrbitClass{OrbitClose, OrbitNear, OrbitFar} {
		idx := findCompanionByParent(sys, -1, oc)
		if idx < 0 {
			continue
		}
		letter := letters[li]
		li++
		// Does this star have its own OrbitCompanion child?
		childIdx := findCompanionByParent(sys, idx, OrbitCompanion)
		if childIdx >= 0 {
			sys.Companions[idx].Designation = letter + "a"
			sys.Companions[childIdx].Designation = letter + "b"
		} else {
			sys.Companions[idx].Designation = letter
		}
	}
}

// findCompanionByParent returns the index of the first companion whose
// ParentIndex matches and whose OrbitClass matches, or -1 if none.
func findCompanionByParent(sys *System, parentIdx int, oc OrbitClass) int {
	for i, c := range sys.Companions {
		if c.ParentIndex == parentIdx && c.OrbitClass == oc {
			return i
		}
	}
	return -1
}
