package worlds

import (
	"fmt"

	"wbh/roller"
)

// Moon — one significant moon. Insignificant moons (free-form Referee
// fiat per WBH p.58) are out of scope for 2C.
type Moon struct {
	Designation string // "Aab IV a", ... — assigned by AssignMoonDesignations
	SizeCode    SizeCode
	DiameterKm  float64

	// Set when the moon is itself gas-giant-sized (rare, GG Special row, p.57):
	GGClass        GasGiantClass
	GGDiameterCode string
	DiameterEarth  float64
	MassEarth      float64

	// 3A1 additions
	Physical      *BodyPhysical
	OrbitPD       float64
	OrbitKm       int
	Eccentricity  float64
	Retrograde    bool
	PeriodHours   float64
	Atmosphere    *Atmosphere    // for HZ-planet moons only
	Hydrographics *Hydrographics // for HZ-planet moons only

	// 3A2a additions
	SurfaceDistribution *SurfaceDistribution
	DayLength           *DayLength
	AxialTilt           *AxialTilt
	TidalLock           *TidalLock
	TidalEffects        *SurfaceTidalEffects

	// 3A2b-temp additions
	Temperature *Temperature

	// 3B-geology additions
	Geology *Geology
}

// HasSurfaceDistribution reports whether surface-distribution data has been generated for this moon.
func (m *Moon) HasSurfaceDistribution() bool { return m.SurfaceDistribution != nil }

// HasDayLength reports whether day-length data has been generated for this moon.
func (m *Moon) HasDayLength() bool { return m.DayLength != nil }

// HasAxialTilt reports whether axial-tilt data has been generated for this moon.
func (m *Moon) HasAxialTilt() bool { return m.AxialTilt != nil }

// HasTidalLock reports whether tidal-lock data has been generated for this moon.
func (m *Moon) HasTidalLock() bool { return m.TidalLock != nil }

// HasTidalEffects reports whether surface tidal-effects data has been generated for this moon.
func (m *Moon) HasTidalEffects() bool { return m.TidalEffects != nil }

// HasTemperature reports whether 5C ran for this moon.
func (m *Moon) HasTemperature() bool { return m.Temperature != nil }

// HasGeology reports whether geology data has been generated for this moon.
func (m *Moon) HasGeology() bool { return m.Geology != nil }

// ParentInfo describes a moon's parent body. Only one of (terrestrial
// SizeCode) or (IsGasGiant + GGClass) should be populated.
type ParentInfo struct {
	IsGasGiant bool
	GGClass    GasGiantClass // NotGasGiant for terrestrial parents
	SizeCode   SizeCode      // for terrestrial parents (e.g. "5", "A")
}

// CountMoons rolls the WBH p.55 Significant Moon Quantity table:
//
//	Size 1-2 → 1D-5    Size 3-9 → 2D-8     Size A-F → 2D-6
//	Small GG → 3D-7    Medium/Large GG → 4D-6
//
// dms is the per-die DM (0 or -1) per the p.55 conditions:
//   - Planet's Orbit# < 1.0
//   - Planet is in orbital slot adjacent to a companion
//   - Planet's slot adjacent to Close/Near unavailability range
//   - Planet in adjacent slot to outermost Close/Near/Far range
//
// Per the book: only ONE DM applies regardless of how many conditions
// are met. Caller is responsible for evaluating conditions and passing
// dms = 0 or dms = -1.
//
// Negative result → returns 0 (no significant moons). Exactly 0 →
// returns 0 (caller treats as a planetary ring per p.55).
func CountMoons(r roller.Roller, parent ParentInfo, dms int) (int, error) {
	// Sub-1-Size terrestrial parents (SizeCode "0", "R", "S") cannot
	// host significant moons per WBH p.55 — the Quantity table starts
	// at Size 1-2. Short-circuit before consuming a die.
	if !parent.IsGasGiant {
		switch parent.SizeCode {
		case "0", "R", "S":
			return 0, nil
		}
	}

	notation, base, dieCount, err := moonQuantityFormula(parent)
	if err != nil {
		return 0, err
	}

	rawSum := r.Roll(notation)
	// dms is per-die: each of the dieCount dice gets dms applied.
	adjusted := rawSum + dms*dieCount
	result := adjusted + base
	if result < 0 {
		return 0, nil
	}
	return result, nil
}

// SizeMoon rolls one significant moon's size per WBH p.57 Significant
// Moon Sizing table:
//
//	1-3 → S            4-5 → D3-1            6 → terr: Size-1-1D / GG: Special
//
// For terrestrial parents on a 6 first roll, applies the WBH p.57
// post-rules: Size 1 parent → moon < parent forces "S"; "exactly 2 less
// than parent" 2D adjust (2 → upgrade by 1; 12 → twin world).
//
// For gas-giant parents on a 6 first roll, dispatches to gasGiantSpecialMoon.
func SizeMoon(r roller.Roller, parent ParentInfo) (Moon, error) {
	first := r.Roll("1D")
	switch {
	case first <= 3:
		return Moon{SizeCode: "S", DiameterKm: BasicTerrestrialDiameter("S")}, nil
	case first <= 5:
		// D3-1 → range 0 to 2
		n := r.Roll("D3") - 1
		// Size 1 terrestrial parent: any moon less than parent (n < 1) → "S".
		if !parent.IsGasGiant && nForSizeCode(parent.SizeCode) == 1 && n < 1 {
			return Moon{SizeCode: "S", DiameterKm: BasicTerrestrialDiameter("S")}, nil
		}
		if n <= 0 {
			return Moon{SizeCode: "R", DiameterKm: 0}, nil
		}
		code := sizeCodeForN(n)
		return Moon{SizeCode: code, DiameterKm: BasicTerrestrialDiameter(code)}, nil
	default: // 6
		if parent.IsGasGiant {
			return gasGiantSpecialMoon(r)
		}
		return terrestrialMoonFirst6(r, parent)
	}
}

// terrestrialMoonFirst6 implements the WBH p.57 first-6 branch for
// terrestrial parents.
func terrestrialMoonFirst6(r roller.Roller, parent ParentInfo) (Moon, error) {
	parentN := nForSizeCode(parent.SizeCode)
	if parentN < 1 {
		// Parent Size 0 / S / unknown — defensive (should not happen
		// via legitimate CountMoons callers since they short-circuit).
		return Moon{SizeCode: "S", DiameterKm: BasicTerrestrialDiameter("S")}, nil
	}
	d := r.Roll("1D")
	resultN := parentN - 1 - d

	// Size 1 parent: any moon less than parent (resultN < 1) → "S".
	if parentN == 1 && resultN < 1 {
		return Moon{SizeCode: "S", DiameterKm: BasicTerrestrialDiameter("S")}, nil
	}

	// "Exactly 2 less than parent" 2D adjustment.
	if resultN == parentN-2 && resultN > 0 {
		switch twoD := r.Roll("2D"); twoD {
		case 2:
			resultN = parentN - 1 // upgrade by 1
		case 12:
			resultN = parentN // twin world
		default:
			_ = twoD // keep at 2-less
		}
	}

	// Negative or zero → ring.
	if resultN <= 0 {
		return Moon{SizeCode: "R", DiameterKm: 0}, nil
	}
	code := sizeCodeForN(resultN)
	return Moon{SizeCode: code, DiameterKm: BasicTerrestrialDiameter(code)}, nil
}

// gasGiantSpecialMoon implements the WBH p.57 Gas Giant Special Moon
// Sizing sub-table:
//
//	1-3 → 1D                  (range 1-6)
//	4-5 → 2D-2                (range 0(R) to A/10)
//	6   → 2D+4                (range 6 to G/16; G triggers Small-GG cascade)
//
// On Size G(16) cascade: roll Small GG (diameter D3+D3, mass 5×(1D+1)).
// Per the WBH footnote, an additional 2D rolling 12 cascades to Medium GG.
func gasGiantSpecialMoon(r roller.Roller) (Moon, error) {
	first := r.Roll("1D")
	switch {
	case first <= 3:
		n := r.Roll("1D")
		code := sizeCodeForN(n)
		return Moon{SizeCode: code, DiameterKm: BasicTerrestrialDiameter(code)}, nil
	case first <= 5:
		n := r.Roll("2D") - 2
		if n <= 0 {
			return Moon{SizeCode: "R", DiameterKm: 0}, nil
		}
		code := sizeCodeForN(n)
		return Moon{SizeCode: code, DiameterKm: BasicTerrestrialDiameter(code)}, nil
	default: // 6
		n := r.Roll("2D") + 4
		if n < 16 {
			code := sizeCodeForN(n)
			return Moon{SizeCode: code, DiameterKm: BasicTerrestrialDiameter(code)}, nil
		}
		// Cascade: moon is itself a gas giant. Start as Small GG.
		ggDiameter := r.Roll("D3") + r.Roll("D3")
		ggMass := float64(5 * (r.Roll("1D") + 1))
		ggClass := GasGiantSmall
		ggCode := gasGiantDiameterCode(ggDiameter)
		// Footnote: additional 2D rolling 12 cascades to Medium GG.
		twoD := r.Roll("2D")
		if twoD == 12 {
			ggClass = GasGiantMedium
			ggDiameter = r.Roll("1D") + 6
			ggCode = gasGiantDiameterCode(ggDiameter)
			ggMass = float64(20 * (r.Roll("3D") - 1))
		}
		return Moon{
			SizeCode:       "G", // GG cascade — moon is itself a gas giant (Size 16)
			GGClass:        ggClass,
			GGDiameterCode: ggCode,
			DiameterEarth:  float64(ggDiameter),
			MassEarth:      ggMass,
		}, nil
	}
}

// moonQuantityFormula returns the dice notation, additive base
// (negative because the book writes "1D-5", "2D-8", etc.), and the
// die count for the per-die DM application.
func moonQuantityFormula(p ParentInfo) (notation string, base, dieCount int, err error) {
	if p.IsGasGiant {
		switch p.GGClass {
		case GasGiantSmall:
			return "3D", -7, 3, nil
		case GasGiantMedium, GasGiantLarge:
			return "4D", -6, 4, nil
		default:
			return "", 0, 0, fmt.Errorf("worlds: CountMoons: unknown GGClass %v", p.GGClass)
		}
	}
	n := nForSizeCode(p.SizeCode)
	switch {
	case n >= 1 && n <= 2:
		return "1D", -5, 1, nil
	case n >= 3 && n <= 9:
		return "2D", -8, 2, nil
	case n >= 10 && n <= 15: // A-F
		return "2D", -6, 2, nil
	default:
		return "", 0, 0, fmt.Errorf("worlds: CountMoons: unsupported parent SizeCode %q", p.SizeCode)
	}
}
