package stars

import (
	"errors"
	"fmt"

	"wbh/roller"
)

// GenerateSystemOpts controls GenerateSystem. WithVariance applies the
// optional ±20% variance to mass and diameter for each star (matching
// Plan 1's GenerateMainSequenceStar). Accuracy is forwarded to
// SmallStarAge for the primary's age. PeculiarColumn selects the WBH
// p.15 Referee option for "Special" (2D=2) primary rolls; zero value
// is Special (the cleaner setting — no BD/D/Peculiar primaries).
type GenerateSystemOpts struct {
	WithVariance   bool
	Accuracy       int          // 1 or 2
	PeculiarColumn PeculiarPath // zero value = PeculiarPathSpecial
}

// GenerateSystem rolls a complete multi-star system from a Roller.
//
// Roll consumption order (P2-10):
//  1. Primary star via GenerateMainSequenceStar.
//  2. Six presence rolls (2D each): Close, Near, Far, Primary companion,
//     then Close companion (if Close present), Near companion (if Near
//     present), Far companion (if Far present).
//  3. Star generation for each present non-primary slot in book order:
//     primary companion → Close → Close companion → Near → Near companion
//     → Far → Far companion.
//  4. Orbital placement for each companion in generation order: orbit#,
//     eccentricity, inclination.
//  5. AssignDesignations.
func GenerateSystem(r roller.Roller, opts GenerateSystemOpts) (System, error) {
	primary, err := GenerateMainSequenceStar(r, GenerateOpts{WithVariance: opts.WithVariance, Accuracy: opts.Accuracy})
	if errors.Is(err, ErrSpecialPrimary) {
		primary, err = generateSpecialPrimary(r, opts)
		if err != nil {
			return System{}, fmt.Errorf("special primary: %w", err)
		}
	} else if err != nil {
		return System{}, fmt.Errorf("primary: %w", err)
	}

	// Presence rolls: Close, Near, Far, primary companion first; then
	// companions of present orbit-class stars in Close/Near/Far order.
	closePresent := RollPresence(r, primary, OrbitClose)
	nearPresent := RollPresence(r, primary, OrbitNear)
	farPresent := RollPresence(r, primary, OrbitFar)
	primaryHasCompanion := RollPresence(r, primary, OrbitCompanion)

	var closeCompanionPresent, nearCompanionPresent, farCompanionPresent bool
	if closePresent {
		closeCompanionPresent = RollPresence(r, primary, OrbitCompanion)
	}
	if nearPresent {
		nearCompanionPresent = RollPresence(r, primary, OrbitCompanion)
	}
	if farPresent {
		farCompanionPresent = RollPresence(r, primary, OrbitCompanion)
	}

	sys := System{
		Primary: primary,
		AgeGyr:  primary.AgeGyr,
	}

	// genCompanion generates one non-primary star and appends it to
	// sys.Companions. Returns the index of the appended companion.
	genCompanion := func(parent Star, parentIdx int, role NonPrimaryRole, oc OrbitClass) (int, error) {
		descriptor := RollNonPrimaryDescriptor(r, parent, role)
		var star Star
		var genErr error
		if descriptor == "Other" {
			otherDesc := RollNonPrimaryDescriptor(r, parent, RoleOther)
			star, genErr = GenerateCompanionStar(r, parent, otherDesc)
		} else {
			star, genErr = GenerateCompanionStar(r, parent, descriptor)
		}
		if genErr != nil {
			return -1, fmt.Errorf("companion (descriptor %q): %w", descriptor, genErr)
		}

		// Fill in physical quantities. For stellar results, run the standard
		// Plan 1 pipeline: mass, diameter, temperature, optional variance,
		// then luminosity from formula. For special objects (D, BD, etc.),
		// keep the GenerateCompanionStar stub values but recompute age.
		if isStellarKind(star.Kind) && star.SpectralType.Letter != 0 {
			st := star.SpectralType
			lc := star.LuminosityClass
			mass, merr := ComputeMass(st, lc)
			if merr != nil {
				return -1, fmt.Errorf("companion mass: %w", merr)
			}
			diameter, derr := ComputeDiameter(st, lc)
			if derr != nil {
				return -1, fmt.Errorf("companion diameter: %w", derr)
			}
			temperature, terr := ComputeTemperature(st)
			if terr != nil {
				return -1, fmt.Errorf("companion temperature: %w", terr)
			}
			if opts.WithVariance {
				mass = ApplyVariance(mass, r, 0.20)
				diameter = ApplyVariance(diameter, r, 0.20)
			}
			star.Mass = mass
			star.Diameter = diameter
			star.Temperature = temperature
			star.Luminosity = ComputeLuminosityFromFormula(diameter, temperature)
			star.AgeGyr = sys.AgeGyr
		} else {
			// Special object: recompute age via the table.
			age, aerr := AgeSpecialObject(r, star.Kind, star.Mass)
			if aerr != nil {
				return -1, fmt.Errorf("special-object age: %w", aerr)
			}
			star.AgeGyr = age
		}

		comp := CompanionStar{
			Star:        star,
			OrbitClass:  oc,
			ParentIndex: parentIdx,
		}
		sys.Companions = append(sys.Companions, comp)
		return len(sys.Companions) - 1, nil
	}

	// Star generation in book order: primary companion, then
	// Close (+ its companion), Near (+ its companion), Far (+ its companion).
	if primaryHasCompanion {
		if _, err := genCompanion(primary, -1, RoleCompanion, OrbitCompanion); err != nil {
			return System{}, err
		}
	}

	closeIdx := -1
	if closePresent {
		idx, err := genCompanion(primary, -1, RoleSecondary, OrbitClose)
		if err != nil {
			return System{}, err
		}
		closeIdx = idx
	}
	if closePresent && closeCompanionPresent {
		if _, err := genCompanion(sys.Companions[closeIdx].Star, closeIdx, RoleCompanion, OrbitCompanion); err != nil {
			return System{}, err
		}
	}

	nearIdx := -1
	if nearPresent {
		idx, err := genCompanion(primary, -1, RoleSecondary, OrbitNear)
		if err != nil {
			return System{}, err
		}
		nearIdx = idx
	}
	if nearPresent && nearCompanionPresent {
		if _, err := genCompanion(sys.Companions[nearIdx].Star, nearIdx, RoleCompanion, OrbitCompanion); err != nil {
			return System{}, err
		}
	}

	farIdx := -1
	if farPresent {
		idx, err := genCompanion(primary, -1, RoleSecondary, OrbitFar)
		if err != nil {
			return System{}, err
		}
		farIdx = idx
	}
	if farPresent && farCompanionPresent {
		if _, err := genCompanion(sys.Companions[farIdx].Star, farIdx, RoleCompanion, OrbitCompanion); err != nil {
			return System{}, err
		}
	}

	// Orbital placement for each companion in the order they were generated.
	// Companions are inner-to-outer in the slice, so when computing
	// parentMass for Close/Near/Far stars (orbiting the inner barycentre),
	// we accumulate the mass of all earlier entries (WBH p.30: "M is the
	// sum of the stars indicated in the orbits around column").
	for i := range sys.Companions {
		c := &sys.Companions[i]
		parentMass := keplerParentMass(&sys, primary, i)

		orbit, oerr := RollStellarOrbit(r, c.OrbitClass, primary.LuminosityClass)
		if oerr != nil {
			return System{}, fmt.Errorf("companion[%d] orbit: %w", i, oerr)
		}
		c.OrbitNumber = orbit
		c.AU = OrbitToAU(orbit)

		ecc, eerr := RollEccentricity(r, EccentricityOpts{IsStar: true})
		if eerr != nil {
			return System{}, fmt.Errorf("companion[%d] eccentricity: %w", i, eerr)
		}
		c.Eccentricity = ecc

		inc, _, ierr := RollInclination(r)
		if ierr != nil {
			return System{}, fmt.Errorf("companion[%d] inclination: %w", i, ierr)
		}
		c.Inclination = inc

		c.PeriodYears = OrbitPeriodYears(c.AU, parentMass, c.Star.Mass)
	}

	AssignDesignations(&sys)
	return sys, nil
}

// keplerParentMass returns the M term for Kepler's third law (WBH p.30)
// for the companion at index i. See comment in GenerateSystem.
func keplerParentMass(sys *System, primary Star, i int) float64 {
	c := sys.Companions[i]
	if c.ParentIndex >= 0 {
		// Orbiting a specific non-primary star (e.g., Cb orbits Ca alone).
		return sys.Companions[c.ParentIndex].Star.Mass
	}
	// ParentIndex == -1: orbits the primary or its barycentre.
	if c.OrbitClass == OrbitCompanion {
		// Tight companion of the primary (e.g., Ab); orbits primary alone.
		return primary.Mass
	}
	// Close/Near/Far: orbits the cumulative inner barycentre = primary
	// plus all already-placed companions (inner-to-outer order).
	total := primary.Mass
	for j := range i {
		total += sys.Companions[j].Star.Mass
	}
	return total
}

// isStellarKind reports whether the kind is a regular star (as opposed
// to a special/post-stellar object).
func isStellarKind(k StarKind) bool {
	switch k {
	case KindMainSequence, KindGiant, KindSubgiant, KindSupergiant, KindSubdwarf:
		return true
	}
	return false
}

// generateSpecialPrimary handles the WBH p.16 "Special" primary branch
// (when the regular 2D Type roll yields 2). It walks the Unusual column
// of the Star Type Determination table and produces either a
// BD/D/NS/BH/etc. primary directly, or — for class-redirect cells
// ("Class III" / "Class IV" / "Class VI") — delegates to
// generatePrimaryAtClass to roll a giant/subgiant/subdwarf primary at
// the indicated class. The age is computed via AgeSpecialObject (WBH p.22)
// for special objects, or SmallStarAge (via generatePrimaryAtClass) for
// class redirects.
func generateSpecialPrimary(r roller.Roller, opts GenerateSystemOpts) (Star, error) {
	column := opts.PeculiarColumn
	if column == "" {
		column = PeculiarPathSpecial
	}
	kind, lc, err := RollSpecialPrimary(r, column)
	if err != nil {
		return Star{}, err
	}
	if lc == "Giants" {
		// WBH p.16: "A result of Class III+ requires a roll in the giants
		// column to determine the final luminosity class of the star,
		// followed by a roll on the type column with DM+1." RollGiantClass
		// handles the first roll; generatePrimaryAtClass handles the rest.
		giantClass, gerr := RollGiantClass(r)
		if gerr != nil {
			return Star{}, gerr
		}
		return generatePrimaryAtClass(r, giantClass, GenerateOpts{WithVariance: opts.WithVariance, Accuracy: opts.Accuracy})
	}
	if lc != "" {
		// Class redirect: re-roll on the regular Star Type Determination
		// flow at the indicated class.
		return generatePrimaryAtClass(r, lc, GenerateOpts{WithVariance: opts.WithVariance, Accuracy: opts.Accuracy})
	}
	// Build the special-object Star with stub physical values appropriate
	// to the kind. The book leaves detailed physics (white-dwarf cooling,
	// pulsar timing, etc.) to the Special Circumstances chapter (deferred).
	star := specialObjectPrimaryStub(kind)

	// Age via AgeSpecialObject. For BD and Protostar, deadStarMass is
	// unused (their formulas don't add a progenitor age). For D/NS/BH/PSR,
	// the dead-star mass is the new object's mass.
	age, err := AgeSpecialObject(r, kind, star.Mass)
	if err != nil {
		return Star{}, err
	}
	star.AgeGyr = age
	return star, nil
}

// specialObjectPrimaryStub returns a Star with placeholder physical
// values appropriate to the kind. Detailed physics (white-dwarf cooling
// curves, pulsar spin-down, etc.) is deferred to the Special Circumstances
// chapter; these stubs are good-faith book-typical values.
func specialObjectPrimaryStub(kind StarKind) Star {
	switch kind {
	case KindBrownDwarf:
		return Star{
			Kind:            KindBrownDwarf,
			LuminosityClass: BD,
			Mass:            0.05,
			Diameter:        0.10,
			Temperature:     1500,
		}
	case KindWhiteDwarf:
		return Star{
			Kind:            KindWhiteDwarf,
			LuminosityClass: D,
			Mass:            0.6,
			Diameter:        0.01,
			Temperature:     8000,
		}
	case KindNeutronStar:
		return Star{
			Kind:        KindNeutronStar,
			Mass:        1.4,
			Diameter:    0.00002, // ~20 km in solar diameters
			Temperature: 600000,
		}
	case KindBlackHole:
		return Star{
			Kind:        KindBlackHole,
			Mass:        5.0,
			Diameter:    0.0,
			Temperature: 0,
		}
	case KindPulsar:
		return Star{
			Kind:        KindPulsar,
			Mass:        1.4,
			Diameter:    0.00002,
			Temperature: 1000000,
		}
	case KindNebula:
		return Star{
			Kind: KindNebula,
		}
	case KindProtostar:
		return Star{
			Kind:        KindProtostar,
			Mass:        0.5,
			Diameter:    1.5,
			Temperature: 3000,
		}
	case KindStarCluster:
		return Star{
			Kind: KindStarCluster,
		}
	case KindAnomaly:
		return Star{
			Kind: KindAnomaly,
		}
	}
	return Star{}
}

// CompanionStar is a non-primary star with its orbital placement.
//
// Plan 2 P2-3 defines the type; later tasks (P2-4 through P2-10) populate
// the orbital fields. ParentIndex is -1 when the parent is the primary,
// or an index into System.Companions otherwise (used to encode that, e.g.,
// a Far-orbit star's own companion has the Far star as its parent).
type CompanionStar struct {
	Star         Star
	Designation  string
	OrbitClass   OrbitClass
	OrbitNumber  float64
	AU           float64
	Eccentricity float64
	Inclination  float64 // degrees
	PeriodYears  float64
	ParentIndex  int
}

// System is a star system with a primary plus zero or more companions.
//
// PrimaryDesignation is set by AssignDesignations and is "A" for a single
// primary or "Aa" if the primary has its own OrbitCompanion-class child.
type System struct {
	Primary            Star
	PrimaryDesignation string
	Companions         []CompanionStar
	AgeGyr             float64
}
