# World Physical Characteristics — Sub-project 3A1 Design (Body Physical + Atmosphere + Hydrographics)

**Date:** 2026-05-03
**Status:** approved through brainstorming; pending user review of written spec
**Source material:** Mongoose Publishing, _World Builder's Handbook_ (Geir Lanesskog, 2023). PDF in repo at `Mongoose/Core Rules/World Builders Handbook.pdf`.
**Source pages:** WBH pp. 69–100.
**Parent spec:** `docs/specs/2026-05-02-world-builder-design.md`.
**Predecessor:** `docs/specs/2026-05-03-system-worlds-2c-sizing-design.md` (System Worlds and Orbits 2C — Sizing + Moons + Form).

## Purpose

Encode the body-physical, atmosphere, atmosphere-profile, and hydrographics procedures from the _World Builder's Handbook_ chapter "World Physical Characteristics" (pp. 69–146), restricted in this sub-project to pp. 69–100. Sub-project 3A1 layers on top of 2C's `SystemDetail` and produces:

1. **Body physical** for every terrestrial body and every HZ-planet moon: composition, density, gravity, mass, escape velocity, orbital velocity, and the size-profile shorthand `S-Dkm-D-G-M`.
2. **Belt characteristics** for every Size-0 body: span, bulk, composition (m/s/c-type %), resource rating, significant Size-1 + Size-S body counts, and the belt-profile shorthand `S-CC.CC.CC.CC-B-R-#-s` (rendered into the IISS Class III form's Notes column).
3. **Moon refinements** for every planet and gas giant with moons: Hill-sphere PD, Roche limit, **moon-removal pass** (Hill Sphere Moon Limit < 1.5 PD ⇒ drop moons, optionally promote first to ring), moon-orbit determination, eccentricity, retrograde, and moon orbital period.
4. **Atmosphere** for every HZ-orbit body and every HZ-planet moon: atmosphere code (UWP digit) via Core book table for HZ + Hot/Cold Atmospheres tables for non-HZ (pp. 94–95), subtype letter (A/B/C/D/F/G/H), pressure, scale height.
5. **Atmosphere profile** (gas mix composition) via temperature-range Gas Mix tables (pp. 96–98), rendered as a profile string e.g. `B-StD:CO2-48:NH3-47:H2O-03`.
6. **Hydrographics** (p. 99): UWP digit via 2D-7+Atmo+DMs, percent range from the Hydrographics Ranges table, linear variance via d10.
7. **Form rendering updates** to `RenderIISSClass23`: full SAH triplet for HZ bodies and HZ-planet moons, `<Size>??` masking preserved for non-HZ bodies (matches book p. 63), belt profile shorthand in the Notes column.

3A1 closes two of the three 2C carry-forward items: HZ-candidate `?` SAH placeholders become numeric digits, and the form's non-HZ masking is confirmed correct. The third carry-forward (`ClassIIIStatus = true` trigger logic) remains deferred to 3C.

## Decomposition context

The "World Physical Characteristics" chapter (WBH pp. 69–146) decomposes into four sub-projects:

| Sub-project | WBH pp.    | Status                                                                                                                          |
| ----------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------- |
| **3A1**     | **69–100** | **This spec.** Body physical, belt details, moon refinements, atmosphere (digit + subtype + profile), hydrographics.            |
| 3A2         | 101–118    | Temperature roll (basic + greenhouse + latitude bands + twilight + variability + tidal heating) + seismology.                   |
| 3B          | 119–140    | Biosphere & habitability (native lifeforms, biocomplexity, sophonts, compatibility/MXDC, resource rating, habitability rating). |
| 3C          | 141–146    | Final mainworld determination (`PickMainworld`) + world maps + `ClassIIIStatus = true` trigger.                                 |

The original resume-notes plan was a single sub-project for pp. 69–146. After reading the chapter, the scope (~30 distinct subsystems, 78 pages, ~3× the size of 2C's surface) was decomposed into the four pieces above. 3A1 is the largest piece (32 pages) and was further considered for splitting into pp. 69–78 (body physical) + pp. 79–100 (atmosphere) but the two halves are tightly coupled — atmosphere's scale-height needs gravity from body physical — so they ship together.

## Non-goals

- **Taint profile.** Biological/chemical hazard layer. 3B owns it (habitability concern, not greenhouse-relevant).
- **Exotic / unusual atmosphere refinements** (the Atmosphere Subtype follow-up tables for HZCO -3.0 dense-irritant rolls and the unusual subtype 7 panthalassic rendering). 3B if needed; otherwise documented as future detail.
- **Full temperature roll** (basic temperature, greenhouse factor, mean temperature by latitude, twilight zones, variability, tidal heating). 3A2 owns it. 3A1 uses HZCO-offset bucketing as a provisional temperature for atmosphere-table selection (Boiling/Hot/Temperate/Cold/Frozen). The HZCO offsets and their mean-temperature bands are the book's own keying for the Hot/Cold Atmospheres tables (pp. 94–95) and Gas Mix tables (pp. 96–98).
- **Re-derivation of pressure / scale-height / hydrographics under full temperature.** Once 3A2 produces a real temperature, fields that depend on it (`Atmosphere.Pressure`, `Atmosphere.ScaleHeight`, `Hydrographics.Code` via the hot/boiling DMs) may shift. 3A1's outputs are documented as **provisional under HZCO temperature**; 3A2 will provide a re-derive pass.
- **Seismology** (pp. 115–118). 3A2.
- **`ClassIIIStatus = true` trigger logic.** 3C closes the form when mainworld is picked and full chapter detail is in.
- **`PickMainworld()`.** Needs habitability rating from 3B. 3C.
- **Surface feature distribution beyond hydro %.** The chapter's "Surface Feature Distribution" subsection covers land/ocean/ice/cloud breakdown for cartography. 3B if biosphere needs it; otherwise 3C with world maps.
- **Insignificant moons / minor body refinements.** Free-form Referee fiat, never encoded.
- **Continuation Method examples** (Corella). 2C deferred this; 3A1 stays excluded. `worlds.ErrContinuationMethodUnsupported` continues as the placeholder.

## Bundled 2C carry-forward items

Per the 2A→2B→2C precedent of bundling relevant carry-forwards, 3A1 closes the carry-forwards directly impacted by atmosphere/hydrographics:

- **Carry-forward #1.** HZ-candidate `?` SAH placeholders → numeric atmosphere/hydrographics digits. **Closed** by 3A1's atmosphere + hydrographics passes.
- **Carry-forward #2.** Non-HZ terrestrial SAH masking. **Confirmed correct** — book p. 63 renders non-HZ as `<Size>??`. 2C's existing rendering matches the book.
- **Carry-forward #3.** `TestZed_FullDetail` is relaxed (sized all terrestrials to "7", all GGs to GLE/500⊕ uniformly, no moons; headline-shape only). **Tightened** in 3A1: every HZ-orbit body has a 3-digit SAH (no `?`); every HZ-planet moon has a 3-digit SAH; non-HZ bodies render as `<Size>??`; belt rows include profile shorthand in Notes; `MainworldCandidates` non-empty; no panics across 100 free-dice iterations.

The remaining 2C carry-forwards (`ClassIIIStatus = true` rendering, `LongProfile` 3-segment vs book's 4-segment, `moonCountDM` adjacency heuristic, `sumStellarMassInterior` simplification, `computeBaselineN` for composite groups) are independent of 3A1 and stay deferred to their natural sub-projects.

## Architecture

### Approach: stay flat in `worlds/`

3A1 continues the 2A/2B/2C pattern: granular per-concern files in the existing `worlds/` package, grouped one file per pipeline stage (`body_physical.go`, `belt_details.go`, `moon_refinement.go`, `atmosphere.go`, `atmosphere_profile.go`, `hydrographics.go`). See the file-layout table below for full per-file scope. New fields land on `DetailedPlacement` via sub-struct pointers (`Physical`, `Belt`, `Atmosphere`, `Hydrographics`), so zero-value `nil` is distinguishable from zero-data.

Two alternative approaches were considered: a new `worlds/physical/` sub-package (cleaner isolation but bigger spec, import-cycle risk, API churn for downstream `cmd/wbh`) and a full `worlds/{placement,sizing,physical,...}/` lifecycle refactor (best long-term layout, but a massive 2A/2B/2C refactor unrelated to user goals). The flat approach is chosen because splitting later is cheap; renaming sub-packages later is expensive. The "package too big" problem is real but solvable after 3A1 ships, when the right cut points are visible from the code rather than the spec.

### Public API additions

#### `worlds/body_physical.go`

```go
package worlds

// BodyPhysical — terrestrial body or moon body physical characteristics, WBH pp. 69–72.
type BodyPhysical struct {
    Composition     string  // "Exotic Ice"|"Mostly Ice"|"Mostly Rock"|"Rock and Metal"|"Mostly Metal"|"Compressed Metal"
    Density         float64 // relative to Terra (1.0 = 5.514 g/cm³)
    Gravity         float64 // relative to Terra (1.0 G)
    EscapeVelocity  float64 // m/s
    OrbitalVelocity float64 // m/s at surface
    SizeProfile     string  // "S-Dkm-D-G-M", e.g. "5-8163-1.03-0.66-0.27"
}

// BodyPhysicalDMs — DM accumulators for the Composition roll, per WBH p. 71.
type BodyPhysicalDMs struct {
    SizeCode      SizeCode // for Size 0-4 DM-1 / Size 6-9 DM+1 / Size A-F DM+3
    AtHZCOOrCloser bool    // DM+1
    BeyondHZCO     int     // DM-1 per full Orbit#
    SystemAgeGyr   float64 // DM-1 if > 10 Gyr
}

// RollComposition: 2D + DMs → composition column on Terrestrial Composition table.
func RollComposition(r roller.Roller, dms BodyPhysicalDMs) (string, error)

// RollDensity: 2D → density value from the Terrestrial Density table column for composition.
func RollDensity(r roller.Roller, composition string) (float64, error)

// DeriveGravity: (Density × Diameter) / DiameterTerra, returned in G.
func DeriveGravity(densityRel float64, diameterKm float64) float64

// DeriveMass: Density × (Diameter / DiameterTerra)³, returned in M⊕.
func DeriveMass(densityRel float64, diameterKm float64) float64

// DeriveEscapeVelocity: √((m/M⊕) ÷ (D/D⊕)) × 11,186 m/s.
func DeriveEscapeVelocity(massEarth, diameterKm float64) float64

// DeriveOrbitalVelocity: EscV ÷ √2 (surface orbit).
func DeriveOrbitalVelocity(escapeVelocity float64) float64

// FormatSizeProfile: "S-Dkm-D-G-M" — Size, Diameter km, Density rel, Gravity G, Mass M⊕.
func FormatSizeProfile(p BodyPhysical, sizeCode SizeCode, diameterKm int) string

// GenerateBodyPhysical orchestrates the per-body pipeline; pure if r is scripted.
func GenerateBodyPhysical(r roller.Roller, sizeCode SizeCode, diameterKm int, dms BodyPhysicalDMs) (BodyPhysical, error)
```

#### `worlds/belt_details.go`

```go
package worlds

// BeltDetails — Size-0 body planetoid belt characteristics, WBH pp. 72–74.
type BeltDetails struct {
    Span            float64        // Orbit#s
    Composition     BeltComposition // m/s/c-type %
    Bulk            int
    ResourceRating  int            // typically 0-12
    SigSize1Bodies  int
    SigSizeSBodies  int
    Profile         string         // "S-CC.CC.CC.CC-B-R-#-s"
}

type BeltComposition struct {
    MTypePct int  // metallic
    STypePct int  // stony
    CTypePct int  // carbonaceous/icy
    OtherPct int  // peculiar / artificial / leftover
}

// RollBeltSpan: spread × 2D / 10. dms applies DM-1 (adjacent slot is GG) or DM+3 (outermost slot).
func RollBeltSpan(r roller.Roller, spreadOrbits float64, dms int) (float64, error)

// RollBeltComposition: 2D+DM on Belt Composition Percentages table.
// dms applies DM-4 (inside HZCO) or DM+4 (beyond HZCO+2).
func RollBeltComposition(r roller.Roller, dms int) (BeltComposition, error)

// RollBeltBulk: 2D2 + (age÷2 floor) + (cType% ÷ 10 floor). Result < 1 ⇒ 1.
func RollBeltBulk(r roller.Roller, ageGyr float64, comp BeltComposition) (int, error)

// RollResourceRating: 2D-7 + bulk + (mType ÷ 10 floor) - (cType ÷ 10 floor).
// Industrial DM-1 and TL≥8 DM-1 deferred (no consumer in 3A1).
func RollResourceRating(r roller.Roller, bulk int, comp BeltComposition) (int, error)

// RollSigSize1Bodies: 2D-12 + Bulk + DMs (HZCO+3 DM+2; span<0.1 DM-4). Negative → 0.
func RollSigSize1Bodies(r roller.Roller, bulk int, beltOrbit, hzco, span float64) (int, error)

// RollSigSizeSBodies: 2D-10 + (DM+1) × (Bulk+1). Span<0.1 ⇒ halve, round up. Negative → 0.
func RollSigSizeSBodies(r roller.Roller, bulk int, beltOrbit, hzco, span float64) (int, error)

// FormatBeltProfile: "S-CC.CC.CC.CC-B-R-#-s".
func FormatBeltProfile(b BeltDetails) string

// GenerateBeltDetails orchestrates the per-belt pipeline.
func GenerateBeltDetails(r roller.Roller, beltOrbit, spreadOrbits, hzco, ageGyr float64, beltSlotIs map[string]bool) (BeltDetails, error)
```

#### `worlds/moon_refinement.go`

```go
package worlds

// HillSphere computes the planet's Hill sphere in AU and PD, WBH p. 75.
//   Hill Sphere (AU) = AU × (1 - ecc) × ³√(m / (3 × M))
//   m = planet mass × 0.000003 (Terra masses → solar units)
//   M = total interior stellar mass in solar units
//   PD = AU × 149,597,870.9 / planet-diameter-km
func HillSphere(au, ecc, planetMassEarth, sumStellarMassSolar, planetDiameterKm float64) (auResult, pd float64)

// RocheLimit: 1.22 × planet diameter × ³√(planet density / moon density).
// Simplified to 1.537 × planet diameter when moon density ≈ planet density / 2 (book p. 76).
func RocheLimit(planetDiameterKm, planetDensityRel, moonDensityRel float64) float64

// HillSphereMoonLimit: HillSpherePD ÷ 2 (round down), prograde-moon outer bound.
func HillSphereMoonLimit(hillSpherePD float64) float64

// MoonOrbitRange: HillSphereMoonLimit (rounded down) - 2. > 200 ⇒ clamp to 200 + nMoons.
func MoonOrbitRange(hillSphereMoonLimit float64, nMoons int) int

// MoonRemovalCheck: if hillSphereMoonLimit < 1.5 PD, all moons drop; first promotes to ring.
func MoonRemovalCheck(hillSphereMoonLimit float64) (removeAll bool, promoteToRing bool)

// RollMoonOrbit: 1D + DM (MOR < 60 ⇒ DM+1) on the Moon Orbit Location table p. 76.
//   Inner:  (2D-2) × MOR ÷ 60 + 2          (rolls 1-3)
//   Middle: (2D-2) × MOR ÷ 30 + MOR÷6 + 3  (rolls 4-5)
//   Outer:  (2D-2) × MOR ÷ 20 + MOR÷6 + 2 + 4 (roll 6+)
// Optional 0.5 PD variance applied via second 1D.
func RollMoonOrbit(r roller.Roller, mor int) (orbitPD float64, err error)

// RollMoonEccentricity: roll on stars.EccentricityValues table with WBH p. 76 DMs:
//   DM-1 inner / DM+1 middle / DM+4 outer / DM+6 if exceeds MOR.
func RollMoonEccentricity(r roller.Roller, range_ MoonRange) (float64, error)

// RollMoonRetrograde: 2D + DM (same as eccentricity); ≥ 10 ⇒ retrograde.
func RollMoonRetrograde(r roller.Roller, range_ MoonRange) (bool, error)

// MoonPeriod: 0.176927 × √((PD × Size)³ / Mp) hours.
//   Size = parent Size code as integer
//   Mp = parent mass in Terra masses
// Returns *stars.Period with Days = Hours/24 / 365.25.
func MoonPeriod(orbitPD float64, parentSize int, parentMassEarth float64) Hours

// RefineMoons orchestrates the per-planet pipeline:
//   compute Hill sphere → Roche → MoonRemovalCheck → for each surviving moon
//     RollMoonOrbit → reorder → RollMoonEccentricity → RollMoonRetrograde → MoonPeriod.
// Mutates the passed *DetailedPlacement.
func RefineMoons(r roller.Roller, dp *DetailedPlacement, sumStellarMassSolar float64) error

type MoonRange int
const (
    MoonRangeInner MoonRange = iota
    MoonRangeMiddle
    MoonRangeOuter
    MoonRangeBeyondMOR
)
```

#### `worlds/atmosphere.go`

```go
package worlds

// Atmosphere — UWP atmosphere code + WBH refinements, pp. 79–95.
type Atmosphere struct {
    Code         int      // UWP digit 0-17+ (extended for exotic G/H)
    Subtype      string   // "A"|"B"|"C"|"D"|"F"|"G"|"H" or ""
    Pressure     float64  // bar
    ScaleHeight  float64  // km
    Profile      AtmosphereProfile
}

// TempRange — provisional temperature class from HZCO offset (book pp. 94–98 keying).
//   Boiling   HZCO offset ≤ -2.01  (mean ≥ 453 K)
//   Hot       HZCO offset -1.01 .. -2.0  (353-453 K)
//   Temperate HZCO offset -1.0 .. +1.0   (273-353 K nominal HZ)
//   Cold      HZCO offset +1.01 .. +3.0  (123-273 K)
//   Frozen    HZCO offset ≥ +3.01        (≤ 123 K)
type TempRange int
const (
    TempBoiling TempRange = iota
    TempHot
    TempTemperate
    TempCold
    TempFrozen
)

// HZCOOffsetToTempRange — provisional temperature for 3A1.
func HZCOOffsetToTempRange(orbitNumber, hzco float64) TempRange

// RollAtmoCodeHZ: roll on the Mongoose Core book Atmosphere table for HZ-orbit bodies.
//   DMs per WBH pp. 79-93. Implementation Task 1 (read pp. 79-93) resolves the
//   exact DM stack and updates this signature; the parameter shape may change.
func RollAtmoCodeHZ(r roller.Roller, sizeCode SizeCode, dms int) (int, error)

// RollAtmoCodeNonHZ: 2D-7 + Size on Hot Atmospheres or Cold Atmospheres tables p. 94-95.
//   HZCO offset selects column.
func RollAtmoCodeNonHZ(r roller.Roller, sizeCode SizeCode, hzcoOffset float64) (int, error)

// ResolveSubtype: from atmo code, may roll subtype follow-up (Insidious type table, etc.).
//   Subtype "A" = exotic, "B" = corrosive, "C" = insidious, "D" = very dense,
//   "F" = unusual, "G" = helium gas, "H" = hydrogen gas. Empty = no subtype.
func ResolveSubtype(r roller.Roller, atmoCode int, sizeCode SizeCode, hzcoOffset float64) (string, error)

// DerivePressure: lookup table from atmosphere code + Size, WBH pp. 79-93.
func DerivePressure(atmoCode int, subtype string, sizeCode SizeCode, tempRange TempRange) float64

// DeriveScaleHeight: physics formula H = (R × T) / (M̄ × g).
//   R = universal gas constant; T = mean temperature K (from TempRange); M̄ = mean molar mass (kg/mol);
//   g = surface gravity m/s². Returns km.
func DeriveScaleHeight(tempRange TempRange, meanMolarMassKgPerMol, gravityG float64) float64

// GenerateAtmosphere orchestrates the per-body pipeline.
func GenerateAtmosphere(r roller.Roller, dp *DetailedPlacement, hzcoOffset float64) error
```

#### `worlds/atmosphere_profile.go`

```go
package worlds

// AtmosphereProfile — gas mix composition from WBH pp. 96-98 tables.
type AtmosphereProfile struct {
    TempRange string         // "Boiling"|"Hot"|"Temperate"|"Cold"|"Frozen"
    Gases     []GasFraction  // ordered by descending percent
    Shorthand string         // "B-StD:CO2-48:NH3-47:H2O-03"
}

type GasFraction struct {
    Name       string  // "CO2"|"N2"|"O2"|"H2O"|"NH3"|"CH4"|... (chemical formula)
    PercentBP  float64 // basis points (10000 = 100%) for precision; rendered as integer %
}

// RollGasMixDMs accumulates the Gas Mix table DMs:
//   Mean temperature 700-2,000 K → DM-2; > 2,000 K → DM-5 (Boiling table only)
//   Mean temperature 70-100 K → DM+3; < 70 K → DM+5 (Frozen table only)
//   Size 1-7 → DM-1 (or DM-2/-3 for boiling/frozen edges)
//   Size A+ → DM+1
type RollGasMixDMs struct {
    TempRange   TempRange
    SizeCode    SizeCode
    Hydro0      bool // *foot-noted: H2O-only swap
}

// RollGasMix: roll 2-3 times on the appropriate temperature-range Gas Mix table,
// allocating percentages per WBH pp. 96-98 procedure:
//   First roll sets primary gas at (1D+4)×10% (with d10 variance, capped at 100%)
//   Second roll sets next gas at (1D+4)×10% of remainder
//   Continue until allocations exceed 95% or referee stops; remainder = "other"
func RollGasMix(r roller.Roller, atmoCode int, subtype string, dms RollGasMixDMs) (AtmosphereProfile, error)

// FormatAtmoProfileShorthand: "B-StD:CO2-48:NH3-47:H2O-03" (Subtype-StateSubtype:Gas-pct:...).
func FormatAtmoProfileShorthand(atmo Atmosphere, profile AtmosphereProfile) string
```

#### `worlds/hydrographics.go`

```go
package worlds

// Hydrographics — UWP hydro digit + percent refinement, WBH p. 99.
type Hydrographics struct {
    Code         int     // 0-A (10)
    PercentRange [2]int  // e.g. [56, 65] for digit 6
    Percent      int     // linear-variance-refined integer percent
}

// RollHydroDigit: 2D-7 + Atmo + DMs on book p. 99 procedure.
//   Size 0 or 1 → DM-4
//   Atmo 0, 1, A+ → DM-4
//   Hot temperature → DM-2
//   Boiling temperature → DM-6
//   Floor at 0; cap at A (10). Note: hot/boiling DMs ignored for D (very dense)
//   and F (unusual subtype 7) atmospheres.
func RollHydroDigit(r roller.Roller, atmoCode int, subtype string, sizeCode SizeCode, tempRange TempRange) (int, error)

// HydroRange: digit → percent range from Hydrographics Ranges table p. 99.
func HydroRange(digit int) [2]int

// RefineHydroPercent: linear variance via d10.
//   Hydrographics 0:  -4 + d10, results <0 treated as 0
//   Hydrographics A:  96 + d10, results >100 capped at 100
//   Otherwise:        range[0] + d10 (mod 10) when range spans 10
func RefineHydroPercent(r roller.Roller, digit int, range_ [2]int) (int, error)

// GenerateHydrographics orchestrates the per-body pipeline.
func GenerateHydrographics(r roller.Roller, dp *DetailedPlacement, tempRange TempRange) error
```

#### Extension to `worlds/system_detail.go`

```go
// DetailSystem — extended pipeline including 3A1 passes.
// Existing 2C steps run first; new 3A1 passes follow.
func DetailSystem(r roller.Roller, sys *stars.System, sp *SystemPlacement, h Histogram) (*SystemDetail, error)
```

The `DetailedPlacement` type (already exists from 2C) gains four sub-struct pointer fields. The existing `Moon` type gains six fields:

```go
// DetailedPlacement — extended for 3A1.
type DetailedPlacement struct {
    Placement
    SizeCode       SizeCode
    DiameterKm     int
    GGClass        GasGiantClass
    GGDiameterCode string
    DiameterEarth  float64
    MassEarth      float64
    Designation    string
    Period         Period
    HZ             bool
    Moons          []Moon

    // 3A1 additions — pointer = nil means "not applicable to this body type"
    Physical      *BodyPhysical
    Belt          *BeltDetails
    Atmosphere    *Atmosphere
    Hydrographics *Hydrographics
}

// Helper methods (hide pointer-juggling at form-rendering layer):
func (dp *DetailedPlacement) HasPhysical() bool       { return dp.Physical != nil }
func (dp *DetailedPlacement) HasAtmosphere() bool     { return dp.Atmosphere != nil }
func (dp *DetailedPlacement) HasHydrographics() bool  { return dp.Hydrographics != nil }
func (dp *DetailedPlacement) RenderSAH() string       { /* "<Size>??" or 3-digit */ }

// Moon — extended for 3A1.
type Moon struct {
    // 2C fields
    Designation    string
    SizeCode       SizeCode
    DiameterKm     float64
    GGClass        GasGiantClass
    GGDiameterCode string
    DiameterEarth  float64
    MassEarth      float64

    // 3A1 additions
    Physical       *BodyPhysical
    OrbitPD        float64
    OrbitKm        int
    Eccentricity   float64
    Retrograde     bool
    Period         Hours
    Atmosphere     *Atmosphere    // for HZ-planet moons only
    Hydrographics  *Hydrographics // for HZ-planet moons only
}

type Hours float64 // hour-level period for moons (vs Years for planets)
```

## Pipeline / data flow

`DetailSystem(r, sys, sp, h)` runs the existing 2C passes followed by 3A1's six new passes in order:

```text
DetailSystem
├─ [2C existing] scaffolding → designations → period → MarkHZ → MainworldCandidates
│
├─ [3A1] 1. refineDiameter        per terrestrial + per HZ-planet moon
│                                  (D3 + 1D + d100 variance per p. 69)
│
├─ [3A1] 2. generateBodyPhysical  per terrestrial + per HZ-planet moon
│         (Composition→Density→Gravity→Mass→Velocities→Profile)
│
├─ [3A1] 3. generateBeltDetails   per Size-0 belt
│         (Span→Composition→Bulk→Resources→SigBodies→Profile)
│
├─ [3A1] 4. refineMoons           per planet/GG with moons
│         (HillSphere→Roche→MoonRemoval→Orbits→Eccentricity→Retrograde→Period)
│         May drop 2C-placed moons if Hill Sphere Moon Limit < 1.5 PD.
│
├─ [3A1] 5. generateAtmosphere    HZ-orbit bodies + HZ-planet moons
│         (TempRange→Code→Subtype→Pressure→ScaleHeight→GasMix)
│
├─ [3A1] 6. generateHydrographics HZ-orbit bodies + HZ-planet moons
│         (Digit→Range→PercentVariance)
│
├─ [2C existing, updated]
│   ├─ ShortProfile               may include atmo/hydro now
│   ├─ LongProfile                may include atmo/hydro now
│   └─ RenderIISSClass23          full SAH for HZ + HZ-planet moons,
│                                 "<Size>??" for non-HZ,
│                                 belt profile in Notes column
```

### Temperature-proxy decision

Steps 5 and 6 need a temperature class for atmosphere-table selection. The book's Hot/Cold Atmospheres tables (pp. 94–95) are explicitly HZCO-keyed; the Gas Mix tables (pp. 96–98) are mean-temperature-keyed but the book sketches a direct correspondence (Boiling 453 K+ ↔ HZCO -2.01 inner, Hot 353-453 K ↔ HZCO -1.01..-2.0, Temperate 273-353 K ↔ HZ band, Cold 123-273 K ↔ HZCO +1.01..+3.0, Frozen ≤ 123 K ↔ HZCO +3.01+).

3A1 uses `HZCOOffsetToTempRange(orbitNumber, hzco)` for atmosphere/gas-mix table selection. 3A2 will replace the proxy with a true temperature roll and provide a re-derivation pass (`ReDeriveAtmosphereUnderTemperature`) for the affected fields.

This is documented in code comments, and 3A1 outputs are flagged as **provisional under HZCO temperature** in the spec, plan, and the `BodyPhysical`/`Atmosphere`/`Hydrographics` doc comments.

### File layout (new files in `worlds/`)

| File                         | Lines (estimate) | Concern                                                                       |
| ---------------------------- | ---------------- | ----------------------------------------------------------------------------- |
| `body_physical.go`           | 200              | Composition, density, gravity, mass, velocities, size profile                 |
| `body_physical_test.go`      | 250              | Tests (incl. Sol Terra worked example)                                        |
| `belt_details.go`            | 220              | Span, bulk, composition, resources, significant bodies, profile               |
| `belt_details_test.go`       | 280              | Tests (incl. Aab PI + Cab PI worked examples)                                 |
| `moon_refinement.go`         | 250              | Hill sphere, Roche, moon removal, orbits, eccentricity, period                |
| `moon_refinement_test.go`    | 320              | Tests (incl. Aab IV worked example, Zed Prime period)                         |
| `atmosphere.go`              | 230              | Code, subtype, pressure, scale height                                         |
| `atmosphere_test.go`         | 280              | Tests (incl. Aab I worked example)                                            |
| `atmosphere_profile.go`      | 200              | Gas mix tables, profile shorthand                                             |
| `atmosphere_profile_test.go` | 250              | Tests (incl. Aab I + Cab II profile examples)                                 |
| `hydrographics.go`           | 130              | Digit, range, percent variance                                                |
| `hydrographics_test.go`      | 180              | Tests                                                                         |
| Extensions to existing       | -                | `system_detail.go`, `survey_form.go`, `profile.go`, `worked_examples_test.go` |

Total new code: ~3,000 lines (production + tests). For comparison, 2C added +3,209 lines across 17 commits.

## Acceptance gates / tests

### Hybrid testing strategy

**Strict / scripted-dice tests** for deterministic outputs (formulas, shorthand strings, lookup tables, book worked examples). **Free-dice + shape tests** for the full Zed system test.

### Strict / scripted-dice tests

#### Body physical (pp. 71–72 worked example for Sol/Terra)

`TestSol_TerraPhysicalProfile` — scripted D3=2 (+600 km), 1D=4 (+300 km), d100=63, density 2D=10 with DM+1 (= 11) → expect:

- `DiameterKm = 8163`
- `Composition = "Rock and Metal"`
- `Density = 1.03` (5.68 g/cm³)
- `Gravity ≈ 0.66` G
- `Mass ≈ 0.27` M⊕
- `EscapeVelocity ≈ 7,262` m/s
- `OrbitalVelocity ≈ 5,135` m/s
- `SizeProfile = "5-8163-1.03-0.66-0.27"`

#### Belt profiles (p. 74 worked examples)

`TestZed_AabPI_BeltProfile` — scripted dice → `0.25-55.40.02.03-3-B-0-3` (Span 0.25, comp 55/40/2/3, Bulk 3, Resource B (=11), 0 Size-1 bodies, 3 Size-S bodies).

`TestZed_CabPI_BeltProfile` — scripted dice → `0.3-15.60.20.05-6-8-2-8`.

#### Moon refinement (pp. 75–77 Aab IV / Zed Prime worked examples)

`TestZed_AabIV_HillSphere` — pure derivation (no roller) → `HillSphereAU ≈ 0.083`, `HillSpherePD = 69.37`, `HillSphereMoonLimit = 34.685`. Moon-removal check returns `false, false` (≥ 1.5 PD).

`TestZed_AabIV_MoonOrbits` — scripted 1D moon-orbit values + 1D variance values from book p. 77 narrative → reordered orbits {4.5, 6.1, 14.0, 22.0, 27.9} PD. (Book gives "the five moons the results are 6.26, 4.13, 21.6, 13.6 and 28.0, rounded and reordered as 4, 6, 14, 22 and 28. Adding a variance produces 4.5, 6.1, 14.0, 22.0 and 27.9.")

`TestZedPrime_OrbitalPeriod` — pure derivation from PD=22, parent Size 8 (D=12,800 km), parent mass 1,200⊕ → period ≈ 624.69 hours = 26.03 days.

`TestMoonRemoval_TriggersOnSmallParent` — synthetic Size 1 terrestrial with very small Hill sphere → `MoonRemovalCheck` returns `true, true` and `RefineMoons` drops moons + adds ring marker.

#### Atmosphere (pp. 94, 98–99 worked examples)

`TestZed_AabI_AtmosphereExample` — scripted dice matching book p. 94 narrative:

- Size B (11) at orbit 1.0, HZCO 3.3 (offset -2.3, Boiling range)
- 2D=5, formula `5 - 7 + 11 = 9` → atmosphere code 9 (corrosive B in Hot HZCO -2.01- column)
- Subtype follow-up roll 2D=7 + DM+2 (Size 8+) + DM+4 (sunward of HZCO) = 13 → subtype D (extremely dense, 500 K+)
- Profile assertion: `Code = 9`, `Subtype = "D"`

`TestZed_CabII_AtmosphereExample` — scripted dice matching book p. 95 narrative:

- Size 4 at orbit 2.9, HZCO 0.75 (effective offset +4.4, Frozen range, HZCO +3.01+ column)
- 2D=10, formula `10 - 7 + 4 = 7` → atmosphere code 7 (exotic A subtype 7)
- Profile assertion: `Code = 7` (or expressed "exotic", per Mongoose convention), `Subtype = "7"` (the panthalassic / unusual subtype index)

#### Atmosphere profile (pp. 98–99 worked examples)

`TestZed_AabI_GasMixProfile` — scripted dice matching book p. 98 narrative:

- Boiling Atmosphere Gas Mix (HZCO -2.01-) corrosive (B) column
- 1D=11 with DM+1 → primary "ammonia" at 1D=1 with d10=2 → 47% (1+0)×10 - 100/0.. needs verification against actual % algorithm
- Subsequent rolls for CO₂ (34.5%), water vapor (13.5%), other 5%
- Final shorthand: `B-StD:CO2-48:NH3-47:H2O-03` (book's "noted profile")

`TestZed_CabII_GasMixProfile` — scripted dice matching book p. 98–99 narrative:

- Frozen Atmosphere Gas Mix (HZCO +3.00+) exotic (A) column
- DM+3 mean temp 70-100 K, DM-3 Size 4 → net DM+0
- 1D=7 → nitrogen, 64%; 1D=7 → nitrogen 89% of remainder = 32% (cumulative 96%); 1D=3 → argon 95% of remainder = 3.8%
- Final shorthand: `A-St7:0.98:N2-96:Ar-04 P.4.7`

The exact dice scripts for the gas-mix tests will be derived from the book's prose during the implementation phase, after pp. 79–93 are read for the unread atmosphere-code/subtype/pressure tables. If the book's narrative under-specifies any roll, the test loosens that assertion (e.g., assert "first gas is N₂ at 60-65%" rather than exact 64%).

#### Hydrographics (p. 99)

`TestHydrographicsTable` — table-driven assertion of digit→range mapping (0→[0,5], 1→[6,15], …, 9→[86,95], A→[96,100]).

`TestHydroPercentVariance_BoundaryCases` — scripted d10 → assert percent within range, with 0 and A clamping behavior.

`TestZed_AabVI_Hydrographics` — book form shows `AB6` for Aab VI in HZ → scripted dice rolls hydrographics digit 6.

### Free-dice shape tests

`TestZed_FullDetail_3A1` (replaces 2C's relaxed `TestZed_FullDetail`):

- All HZ-orbit bodies (Aab IV, V, VI; AB III; B II; Cab I) have full 3-character SAH triplet (no `?` placeholders).
- All HZ-planet moons (Aab IV a-d, Aab V a-d) have full 3-character SAH.
- Non-HZ bodies render as `<Size>??` (matches book p. 63).
- Belt rows include profile shorthand string in the Notes column.
- `MainworldCandidates` list non-empty (at minimum: Aab IV d, Aab V b based on 2C's enumeration).
- No panics across 100 iterations with random seeds.

### Form rendering tests

`TestRenderIISSClass23_3A1` — Zed system with mocked `*DetailedPlacement` values → output form matches the book p. 63 layout:

- HZ body row: 3-digit SAH (e.g., `AB6`)
- Non-HZ body row: `<Size>??` (e.g., `B??`, `6??`)
- Belt row: SAH-cell `000`, Notes column contains belt profile shorthand
- HZ-planet moon row at bottom: separate rows with primary designation + 3-digit SAH

`TestRenderIISSClass23_NoFormSpec_RegressionFromValid2CForm` — verify 3A1 doesn't break existing 2C form rendering for non-HZ details.

### Test files

- `body_physical_test.go` (new)
- `belt_details_test.go` (new)
- `moon_refinement_test.go` (new)
- `atmosphere_test.go` (new)
- `atmosphere_profile_test.go` (new)
- `hydrographics_test.go` (new)
- `worked_examples_test.go` (extended)
- `system_detail_test.go` (extended)
- `survey_form_test.go` (extended)

### Dice convention reminder (verbatim for plan/subagent briefs)

Per `roller/roller.go:47-50`, scripted values are **final results, one per `Roll()` call regardless of dice notation**. When the book says "2D=5 + DM+1 = 6", the scripted value is **5** (the 2D pre-DM result); the DM is applied in code. Similarly for "1D+4 = 7", the scripted value is **3** (the 1D), not 7.

This caused 4 bugs in 2C. Subagent task briefs must call this out at the top of every implementation task.

## Risks

1. **WBH pp. 79–93 unread.** This spec covers HZ atmosphere code and pressure tables on hand-wave only — the actual content (Atmosphere Type table for HZ, Atmosphere Subtype follow-up table, Pressure table) is on pages I haven't read yet. The implementation plan's first task must read pp. 79–93 and revise this spec if material is found that the design doesn't anticipate.

2. **Moon-removal cascade.** Hill Sphere Moon Limit < 1.5 PD ⇒ drop all moons + promote first to ring. Could invalidate Zed worked-example moon placements or other 2C test fixtures. Mitigation: confirm Aab IV's 5 moons survive (Hill Sphere Moon Limit = 34.685 ≫ 1.5 PD); add explicit removal test for a synthetic small Size 1 planet.

3. **Temperature-proxy fidelity.** Mapping HZCO offsets to the Gas Mix tables' Kelvin bands is the book's own keying but may not be exact at boundaries. The Gas Mix tables also reference DMs like "Mean temperature 70-100 K → DM+3" — without a true temperature, we can only apply DMs that fall cleanly within a TempRange bucket. Edge-case Size A+ + frozen Size 1-7 worlds may produce different results once 3A2 lands. Documented as carry-forward.

4. **Gas Mix table DM stack.** Size DM, mean temperature DM, hydrographics-0 footnote, HZCO offset DM stack on each gas-mix roll. High subagent-stumble risk. Mitigation: write the DM-stacking helper first as a strict-tested unit before any rolling code uses it.

5. **WBH internal inconsistencies** (per existing memory: p. 19 vs p. 42, p. 58 vs p. 63). pp. 79–100 likely has more. Mitigation: build acceptance tests around what the book says explicitly in worked examples; flag inconsistencies in `feedback_*.md` memory rather than papering over them.

6. **`DetailedPlacement` field bloat.** Approach A keeps everything in one struct via sub-struct pointers. Pointer-discipline (nil-check before deref) gets verbose at form-rendering. Mitigation: helper methods on `DetailedPlacement` (`HasAtmosphere() bool`, `RenderSAH() string`).

7. **Atmosphere Profile shorthand format may need revision.** The format `B-StD:CO2-48:NH3-47:H2O-03` is inferred from book examples; the exact delimiter/ordering rules are not formally specified. If book pp. 79–93 contradict, revise during implementation.

## Working pattern

Per 2A/2B/2C precedent:

- Brainstorm (Opus, complete) → spec (this doc) → writing-plans (Opus) → subagent-driven-development.
- Sonnet implementer subagents with per-task spec + code review loops (caught 6 bugs in 2B, 4 in 2C).
- Sonnet reviewers per task.
- Opus whole-branch review at the end (skipped in 2C; **recommended for 3A1** given larger scope and the unread pp. 79–93 risk).
- ~6-10 implementation tasks. Tentative breakdown:
  1. Read pp. 79–93 and revise spec/plan if needed (no code).
  2. `body_physical.go` + tests.
  3. `belt_details.go` + tests.
  4. `moon_refinement.go` + tests (incl. moon-removal pass).
  5. `atmosphere.go` + tests (HZ + non-HZ atmosphere code roll).
  6. `atmosphere_profile.go` + tests (gas mix).
  7. `hydrographics.go` + tests.
  8. `system_detail.go` extension (orchestration).
  9. `survey_form.go` extension (form rendering).
  10. `worked_examples_test.go` extension (`TestZed_FullDetail_3A1`).

- Read pp. 79–93 PDF directly **before** writing the implementation plan, not during execution.
- Local-only repo — never `git push`.
- Branch name: `feat/wbh-world-physical-3a1`.

## Open questions for plan phase

These can be settled during writing-plans, after reading pp. 79–93:

- Exact DM stack on `RollAtmoCodeHZ` for HZ-orbit bodies. The non-HZ formula (`2D - 7 + Size`) is on p. 94; the HZ-equivalent is on pp. 79–93.
- Pressure formula or table format. WBH mentions a "Pressure table" but the index location and DM stack are unclear without reading those pages.
- Subtype follow-up tables. The Insidious Atmosphere Type table is referenced but not extracted; the Exotic Atmosphere Subtype table likewise.
- Whether `RollGasMixDMs` includes a Hydrographics-0 swap footnote (the book hints that worlds with no surface water swap CO₂ → carbon monoxide or N₂ → nitrogen-ish in some columns).
- Atmosphere Profile shorthand exact grammar — verify via book examples before locking the format string.
