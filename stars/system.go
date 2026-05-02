package stars

import (
	"fmt"

	"wbh/roller"
)

// GenerateSystemOpts controls GenerateSystem. WithVariance applies the
// optional ±20% variance to mass and diameter for each star (matching
// Plan 1's GenerateMainSequenceStar). Accuracy is forwarded to
// SmallStarAge for the primary's age.
type GenerateSystemOpts struct {
	WithVariance bool
	Accuracy     int // 1 or 2
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
	primary, err := GenerateMainSequenceStar(r, GenerateOpts(opts))
	if err != nil {
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

	// Orbital placement for each companion in generation order.
	for i := range sys.Companions {
		c := &sys.Companions[i]
		var parentMass float64
		if c.ParentIndex == -1 {
			parentMass = primary.Mass
		} else {
			parentMass = sys.Companions[c.ParentIndex].Star.Mass
		}

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

// isStellarKind reports whether the kind is a regular star (as opposed
// to a special/post-stellar object).
func isStellarKind(k StarKind) bool {
	switch k {
	case KindMainSequence, KindGiant, KindSubgiant, KindSupergiant, KindSubdwarf:
		return true
	}
	return false
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
