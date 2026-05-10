# System Worlds and Orbits — Sub-project 2C Design (Sizing + Moons + IISS Class II/III Form)

**Date:** 2026-05-03
**Status:** approved through brainstorming; pending user review of written spec
**Source material:** Mongoose Publishing, _World Builder's Handbook_ (Geir Lanesskog, 2023). PDF in repo at `Mongoose/Core Rules/World Builders Handbook.pdf`.
**Source pages:** WBH pp. 53–67.
**Parent spec:** `docs/pass-1/specs/2026-05-02-world-builder-design.md`.
**Predecessor:** `docs/pass-1/specs/2026-05-03-system-worlds-2b-counts-placement-design.md` (World Counts + Placement).

## Purpose

Encode the orbital-period, sizing, significant-moon, designation, profile, mainworld-candidate, and IISS Class II/III survey-form procedures from the _World Builder's Handbook_ chapter "System Worlds and Orbits." Sub-project 2C completes the chapter by layering on top of 2B's `SystemPlacement`. The Class II/III form (book p. 61 blank, p. 63 Zed, p. 65 Corella, p. 67 Terra/Sol) is the chapter's natural rendering target; 2C produces a structured representation of it.

## Decomposition context

The "System Worlds and Orbits" chapter (WBH pp. 36–67) decomposes into three sub-projects, all layered downstream of the existing Stars chapter:

| Sub-project | WBH pp.      | Status                                                                                                                                    |
| ----------- | ------------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| 2A          | 38–43        | **Done** — merged 2026-05-02. Available Orbits + HZCO.                                                                                    |
| 2B          | 36–38, 43–52 | **Done** — merged 2026-05-03. World counts + 8-step placement (Steps 1–9).                                                                |
| **2C**      | **53–67**    | **This spec.** Orbital periods, terrestrial/gas-giant sizing, moons, designations, profile, mainworld candidates, IISS Class II/III form. |

2C consumes 2B; 2C does not depend on the next chapter (sub-project 3, World Physical, pp. 69–146).

Note: 2C's source-page range extends one page earlier than the resume-notes "54–67" because **Planetary Orbital Periods** (p. 53) is needed for the IISS form's Period column and isn't covered by 2A or 2B.

## Non-goals

- **HZ atmosphere/hydrographics rolls.** The Aab IV/V/VI rows in the Zed form (p. 63) show numeric SAH like `AB6`, `AA6`, `200`, `340`, `566` for HZ-candidate worlds and moons. Rolling these requires the World Physical chapter (sub-project 3, atmosphere/hydrographics/temperature procedures, pp. 108+). 2C renders these cells with `?` placeholders for the atmosphere and hydrographics digits, retaining the Size character.
- **Mainworld picker.** 2C enumerates HZ-eligible candidates (terrestrials in the HZ + significant moons of HZ planets) but does not pick. The picker requires SAH rolls; lands with sub-project 3. No "Zed Prime" naming, no asterisk marker on the form, no `MainworldCandidate.IsPicked` field.
- **Insignificant moons** (p. 58). Free-form Referee fiat, not procedural; never encoded.
- **Continuation Method** examples (Corella, p. 65). Pre-placed mainworld branch. Already excluded from 2B; stays excluded. `worlds.ErrContinuationMethodUnsupported` continues as the placeholder.
- **Hill-sphere optional rule for moon DM-1** (p. 56 sidebar). Replaces the standard moon-count DM-1 conditions with "DM-1 for any planet with a Hill sphere of <60 planetary diameters." Defer until the Hill-sphere alternate orbit method itself is built (still deferred from 2A).
- **Insignificant ring SAH notation** beyond `R0#` for significant rings.
- **CLI integration of the new form.** `cmd/wbh` extension is straightforward but worth deferring until 2C lands the structured form, which is the natural rendering target. Likely a small follow-up after 2C merges.
- **Typst/markdown PDF rendering of the form.** Separate small project parallel to the Class 0/I form workflow; deferred.
- **Per-star baseline number + spread** (Step 5 sidebar, optional, p. 49). Still deferred from 2B.
- **10–20% minimum-separation in compact systems** (Step 5 sidebar, p. 49). Still deferred from 2B.
- **Hill-sphere alternate orbit method** (pp. 40–41). Still deferred from 2A.
- **Post-stellar primary MAO.** Brown Dwarf, White Dwarf, Neutron Star, Black Hole, Pulsar primaries. Still deferred from 2A; lives with Special Circumstances chapter (sub-project 5, pp. 219–234).
- **2B carry-forward item #5** (`PrefixRoll` audit-trail collision-fallback fix). Skipped since IISS Class II/III form has no PrefixRoll column. Revisit only if a future consumer renders this field.
- **`Other`-descriptor wart in `stars.GenerateCompanionStar`.** Tracked since 2A; still its own small follow-up.

## Bundled 2B carry-forward items

Per the 2A→2B precedent of bundling relevant carry-forwards (Group.sourceCompanion refactor in 2B), 2C absorbs four small items the 2B post-mortem identified:

- **Carry-forward #1.** Add `TestSol_GenerateSystemPlacement` smoke test to `worlds/worked_examples_test.go`. Single-star coverage was missing in 2B (`TestZed_FullPlacement` only exercised the multi-star path). The smoke test asserts no error, `len(Allocations)==1`, `len(Placements)>0`. No book-narrated dice trail required; per-step Sol tests already cover correctness.
- **Carry-forward #2.** Wrap `GenerateSystemPlacement` errors with `fmt.Errorf("worlds: <step>: %w", err)` at all five callsites in `worlds/system_placement.go`. Existing pattern from `available_orbits.go:359`. Helps when 2C extends the façade.
- **Carry-forward #3.** Plumb `sys.Primary.AgeGyr` through to `RollPlanetEccentricities`. WBH p. 27 specifies a sub-1.0/age>1 Gyr DM that today goes unapplied. Signature changes from `(r, ps)` to `(r, ps, ageGyr float64)`. Update `system_placement.go` façade caller and 2 affected tests. Existing 2B `TestZed_FullPlacement` may need eccentricity assertions updated if the DM shifts table-lookup indices.
- **Carry-forward #4.** Strengthen `TestRollEccentricity_ExtraDM` and `TestRollPlanetEccentricities_AppliesAnomalyDM` to assert specific resulting eccentricity values via `stars.EccentricityValues` table lookup at `(roll - DM)`, not just "DM had some effect."

## Architecture

### Public API

#### Package `wbh/stars` — orbital period (mirrors `Star.HZCO` placement)

```go
package stars

// Period — orbital period; both fields populated, renderer picks based on magnitude.
type Period struct {
    Years float64  // primary representation; from Kepler's 3rd
    Days  float64  // = Years * 365.25
}

// OrbitalPeriod computes one body's orbital period at orbit (in AU) given the
// sum of stellar masses interior to that orbit (in solar units), per WBH p.53.
//
//	Single star:        P = sqrt(AU^3 / M☉)
//	Multiple stars:     P = sqrt(AU^3 / Σ M☉)
//	Large planet:       P = sqrt(AU^3 / (Σ M☉ + m⊕ × 0.000003))
//
// bodyMassEarth = 0 selects the standard formula; pass the body's mass in
// Terra masses to apply the Large Planet variant.
func OrbitalPeriod(au, sumStellarMassSolar, bodyMassEarth float64) Period
```

##### Cycle resolution

`stars` cannot import `worlds` (worlds already imports stars; the reverse would create a cycle). To keep `OrbitalPeriod` in `stars/`, the `Period` type lives in `stars/period.go`. `worlds` re-exports it as a type alias for ergonomic access in 2C consumers:

```go
// worlds/period.go (alias, no new behavior)
type Period = stars.Period
```

#### Package `wbh/worlds` — sizing, moons, profile, designations, mainworld, form

```go
package worlds

// ---- Size primitives -------------------------------------------------------

// SizeCode is the WBH terrestrial Size character: "0", "R" (ring), "S" (small),
// "1"-"9", "A"-"F". Empty string means "not a size-having body" (belt / empty).
type SizeCode string

type GasGiantClass int

const (
    NotGasGiant GasGiantClass = iota
    GasGiantSmall   // GS — Neptune analogue (D3+D3 → 2-6⊕ diameter)
    GasGiantMedium  // GM — Jupiter analogue (1D+6 → 6-12⊕)
    GasGiantLarge   // GL — Superjovian (2D+6 → 8-18⊕)
)

// ---- Sizing (pp. 54–55) ----------------------------------------------------

type TerrestrialSize struct {
    SizeCode   SizeCode
    DiameterKm float64
}

// RollTerrestrialSize: 1D selector → second roll per WBH p.54 table:
//   1-2 → 1D            (range 1-6)
//   3-4 → 2D            (range 2-C/12)
//   5-6 → 2D+3          (range 5-F/15)
// DiameterKm comes from Basic Terrestrial World Size table (p.54).
func RollTerrestrialSize(r roller.Roller) (TerrestrialSize, error)

type GasGiantSize struct {
    Class         GasGiantClass
    DiameterCode  string   // "2"-"G"; the # in GS#/GM#/GL#
    DiameterEarth float64  // in Terra diameters (Size 8 = 1.0)
    MassEarth     float64  // in Terra masses
}

// RollGasGiantSize: 1D+DM category → diameter sub-roll → mass sub-roll, per WBH p.55.
// dms accumulates:
//   -1 if primary is Brown Dwarf, M-V, or any Class VI
//   -1 if system spread < 0.1
// Large-GG mass clamp: if initial mass ≥ 3,000⊕ (3D ≥ 15), roll 2D-2 and
// substitute mass = 4000 - 200 × (2D-2) per the p.55 footnote.
func RollGasGiantSize(r roller.Roller, dms int) (GasGiantSize, error)

// ---- Moons (pp. 55, 57–58) -------------------------------------------------

// Moon — one significant moon. Insignificant moons (free-form Referee) out of scope.
type Moon struct {
    Designation    string  // "Aab IV a", ... — assigned by AssignMoonDesignations
    SizeCode       SizeCode
    DiameterKm     float64

    // Set when the moon is itself gas-giant-sized (rare, GG Special row, p.57):
    GGClass        GasGiantClass
    GGDiameterCode string
    DiameterEarth  float64
    MassEarth      float64
}

type ParentInfo struct {
    IsGasGiant bool
    GGClass    GasGiantClass  // NotGasGiant for terrestrial parents
    SizeCode   SizeCode       // for terrestrial parents (e.g., "5", "A")
}

// CountMoons rolls the WBH p.55 Significant Moon Quantity table:
//   Size 1-2 → 1D-5    Size 3-9 → 2D-8     Size A-F → 2D-6
//   Small GG → 3D-7    Medium/Large GG → 4D-6
// dms is the per-die DM (0 or -1 per p.55 conditions). Negative result → 0.
// Result of exactly 0 → caller treats as a ring (set SizeCode "R").
func CountMoons(r roller.Roller, parent ParentInfo, dms int) (int, error)

// SizeMoon rolls one significant moon's size per WBH p.57.
// Significant Moon Sizing:
//   1-3 → S        4-5 → D3-1 (range 0(R)-2)
//   6   → terrestrial: Size-1-1D    gas giant: Special (sub-table)
// Gas Giant Special Moon Sizing (p.57, when parent is GG and first roll = 6):
//   1-3 → 1D       4-5 → 2D-2 (range 0(R)-A)
//   6   → 2D+4 (range 6-G); on Size G → cascade to Small GG (Medium on a 12 sub-roll)
// Terrestrial Size-1 parent: any moon < parent forces SizeCode "S".
// "Exactly 2 less than parent" 2D adjustment: result 2 → moon is 1 less than parent;
// result 12 → twin world (identical Size); otherwise → keep current.
func SizeMoon(r roller.Roller, parent ParentInfo) (Moon, error)

// ---- Designations (pp. 53, 58 sidebars) ------------------------------------

// AssignPlanetDesignations walks Detailed in orbit order per group and assigns:
//   - Non-belt planets:    "<Group> I", "<Group> II", ... (planet counter)
//   - Planetoid belts:     "<Group> PI", "<Group> PII", ...  (belt counter)
// Belt position skips the planet enumeration; planet counter never advances on a belt.
// Each new group resets both counters to I.
// Mutates DetailedPlacement.Designation in place.
func AssignPlanetDesignations(detailed []DetailedPlacement)

// AssignMoonDesignations assigns "<Planet> a", "<Planet> b", ... to each Moon
// in DetailedPlacement.Moons, in alphabetical order matching the closest-to-farthest
// moon-orbit order (which is the order CountMoons + SizeMoon produced them).
// Mutates Moon.Designation in place.
func AssignMoonDesignations(detailed []DetailedPlacement)

// ---- Mainworld (pp. 58–59) -------------------------------------------------

type MainworldCandidate struct {
    Designation       string  // "Aab VI" or "Aab IV d"
    SizeCode          SizeCode
    DiameterKm        float64
    Orbit             float64  // host planet's orbit (for moons, the parent's)
    HostStarGroup     string   // "Aab", "B", "Cab"
    IsMoon            bool
    ParentDesignation string   // "" for planet candidates; "Aab IV" for moons
}

// MarkHZ sets DetailedPlacement.HZ when the placement's orbit lies within
// HZCO ± 1.0 of the placement's host group, per WBH p.58.
// Uses Group.HZCO() from 2B.
func MarkHZ(detailed []DetailedPlacement, sys stars.System) error

// MainworldCandidates returns the subset of bodies eligible to be the mainworld
// per WBH p.58: terrestrial planets in the HZ, plus significant moons (any Size)
// of any HZ planet (including gas giant moons). Belts and gas giants themselves
// are not candidates. Selection (the "Zed Prime" pick) requires World Physical
// atmosphere/hydrographics rolls; deferred to sub-project 3.
func MainworldCandidates(sd SystemDetail) []MainworldCandidate

// ---- Profile (p. 58) -------------------------------------------------------

// ShortProfile renders the WBH p.58 short form: "G-P-T-N-S".
//   G = gas giant count, P = belt count, T = terrestrial count
//   N = baseline number (0 if < 0), S = system spread
func ShortProfile(sd SystemDetail) string

// LongProfile renders the WBH p.58 long form: "St-N-W-W-W...-S:-N-St-W-W...-S:..."
//   St = star/group designation, N = baseline number for that star, S = spread
//   W per slot, in orbit order: G (gas giant), P (belt), T (terrestrial)
//   M (mainworld) reserved; 2C never emits M (no picker).
//   GM (gas-giant moon mainworld) and PM (belt mainworld) reserved; 2C never emits.
func LongProfile(sd SystemDetail) string

// ---- DetailedPlacement + SystemDetail (façade output) ----------------------

// DetailedPlacement extends 2B's Placement via embedding (continues the
// Slot → AnomalousSlot → Placement → DetailedPlacement chain).
type DetailedPlacement struct {
    Placement                       // 2B (Body, PrefixRoll, AnomalousSlot, ...)

    // Terrestrial fields — set when Body == BodyTerrestrial
    SizeCode       SizeCode
    DiameterKm     float64

    // Gas-giant fields — set when Body == BodyGasGiant
    GGClass        GasGiantClass
    GGDiameterCode string
    DiameterEarth  float64
    MassEarth      float64

    // All non-empty bodies:
    Designation    string  // "Aab I", "Aab PI" — per p.53 sidebar
    Period         Period
    HZ             bool
    Moons          []Moon
}

// SystemDetail — façade output, layered atop 2B's SystemPlacement.
type SystemDetail struct {
    SystemPlacement                       // 2B
    Detailed         []DetailedPlacement  // 1:1 with SystemPlacement.Placements
    ShortProfile     string
    LongProfile      string
    Survey           IISSClass23Form
}

// ---- IISS Class II/III Survey form (pp. 60–67) -----------------------------

// IISSClass23Form — structured rendering of WBH form 0421D-II.III.
// Cells dependent on World Physical render with "?" placeholders until sub-project 3.
type IISSClass23Form struct {
    SectorLocation                                        string  // "Storr | 0602"
    InitialSurvey                                         string  // "207-568"
    LastUpdated                                           string  // "218-1061"
    IISSDesignation                                       string  // "Zed (system)"
    SystemAgeGyr                                          float64
    StellarCount, GasGiants, PlanetoidBelts, Terrestrials int
    ClassIIIStatus                                        bool   // 2C always renders false
    Stars                                                 []StarRow
    Notes                                                 string
    Objects                                               []ObjectRow
    Comments                                              string
}

type StarRow struct {
    Component   string  // "Aa", "Ab", "Aab (A)", "B", "AB", "Ca", "Cb", "Cab (C)", "ABC"
    Class       string  // "G7 V" (empty for composite barycentre rows)
    Mass        float64
    TempK       float64 // 0 for composite rows (renders as "—")
    DiameterRel float64 // 0 for composite rows
    Luminosity  float64
    Orbit       float64
    AU          float64
    Ecc         float64
    PeriodStr   string  // "1.841d" or "8.627y" — magnitude-aware unit
    MAO         float64 // 0 for non-group rows (renders as "—")
    HZCO        float64 // 0 for non-group rows
}

type ObjectRow struct {
    Primary     string  // host star group: "Aab", "AB", "B", "Cab"
    Designation string  // "Aab I", "Aab IV d"
    Orbit       float64
    AU          float64
    Ecc         float64
    PeriodStr   string
    SAH         string  // "B??" / "GLE" / "AA6" / "200" / "566*" / "000" / "S"
    Sub         string  // significant-moon count, "?" for belt, "" for moon row
    Notes       string  // "HZ, R02, S, 1, 1" / "1,200⊕, HZ, 200, S, S, 566*, S"
}

type IISSClass23Header struct {
    SectorLocation, InitialSurvey, LastUpdated, IISSDesignation string
    Comments                                                    string
    // SystemAgeGyr pulled from sys.Primary.AgeGyr automatically.
}

// RenderIISSClass23 builds the form from a fully-populated SystemDetail.
// Stars rows are derived from sys (primary + companions, plus barycentric
// composite rows for groups).
func RenderIISSClass23(sd SystemDetail, sys stars.System, h IISSClass23Header) IISSClass23Form

// ---- Façade ---------------------------------------------------------------

// DetailSystem composes the full pp. 53–67 procedure:
//   1. For each placement: roll Size (terrestrial or gas giant); attach diameter/mass.
//   2. For each non-belt non-empty placement: roll moon count + per-moon sizes.
//   3. AssignPlanetDesignations + AssignMoonDesignations.
//   4. Compute Period for each placement.
//   5. MarkHZ.
//   6. ShortProfile + LongProfile.
//   7. RenderIISSClass23 with the supplied header.
func DetailSystem(r roller.Roller, sys stars.System, sp SystemPlacement, h IISSClass23Header) (SystemDetail, error)
```

### Key tactical decisions

- **`Period` lives in `stars/`**, not `worlds/` — pure stellar mechanics, mirrors `Star.HZCO` placement. `worlds.Period` is a type alias to avoid cycles.
- **`SizeCode` is a `string` newtype**, not an enum — the book's codes are mixed-base (`0`, `R`, `S`, `1`-`9`, `A`-`F`); string with documented domain is friction-free for both parsing and rendering, and matches how the form's SAH column renders.
- **`GasGiantSize` returns class+diameterCode+diameter+mass together** — second and third rolls are inseparable per WBH p.55 procedure (you cannot determine mass without the sized class).
- **`AssignPlanetDesignations` and `AssignMoonDesignations` mutate in place** — designations are derived from slot/moon ordering already settled by upstream; reassigning later is undefined.
- **Approach 2 (DetailedPlacement embeds Placement)** chosen over flat-mutation or side-tables; preserves 2B types untouched and continues the existing embedding chain.
- **Stars rows + composite barycentre rows** — the form (p.63) shows individual companion stars (Aa, Ab) plus composite barycentres (Aab, AB, Cab, ABC). The renderer walks `sys.Primary` + companions and emits both per-star and composite rows.

## File layout

```text

├── stars/
│   ├── period.go                       NEW   Period struct, OrbitalPeriod()
│   └── period_test.go                  NEW   per-formula + Sol/Zed values
├── worlds/
│   ├── period.go                       NEW   type Period = stars.Period (alias)
│   ├── sizing_terrestrial.go           NEW   RollTerrestrialSize, BasicTerrestrialDiameter table
│   ├── sizing_terrestrial_test.go      NEW   per-row + each 1D selector branch + Zed sizes
│   ├── sizing_gasgiant.go              NEW   RollGasGiantSize, GS/GM/GL tables, Large mass clamp
│   ├── sizing_gasgiant_test.go         NEW   per-class + DM combinations + Large clamp + Zed GGs
│   ├── moons.go                        NEW   Moon, ParentInfo, CountMoons, SizeMoon
│   ├── moons_test.go                   NEW   per-row + each adjacent-zone DM + GG Special branches
│   ├── designations.go                 NEW   AssignPlanetDesignations, AssignMoonDesignations
│   ├── designations_test.go            NEW   belt-skip + per-group reset + moon alphabet
│   ├── profile.go                      NEW   ShortProfile, LongProfile
│   ├── profile_test.go                 NEW   per-format + Zed strings
│   ├── mainworld.go                    NEW   MarkHZ, MainworldCandidates
│   ├── mainworld_test.go               NEW   HZ-window + planet/moon candidate enumeration
│   ├── survey_form.go                  NEW   IISSClass23Form, StarRow, ObjectRow, RenderIISSClass23
│   ├── survey_form_test.go             NEW   per-section + cell rendering + period unit selection
│   ├── system_detail.go                NEW   DetailedPlacement, SystemDetail, DetailSystem façade
│   ├── system_detail_test.go           NEW   per-step composition
│   ├── worked_examples_test.go         EDIT  add TestZed_FullDetail (acceptance gate); add TestSol_GenerateSystemPlacement (carry-forward #1)
│   ├── system_placement.go             EDIT  carry-forward #2 — wrap five callsite errors with fmt.Errorf
│   ├── planet_eccentricity.go          EDIT  carry-forward #3 — accept ageGyr, apply WBH p.27 sub-1.0/age>1Gyr DM
│   └── planet_eccentricity_test.go     EDIT  carry-forwards #3+#4 — pass ageGyr, assert specific values
└── cmd/wbh/                            (no changes — CLI integration deferred)
```

**18 new files, 4 edited files.**

## Testing strategy

Same TDD pattern as 2A/2B: per-step unit tests against synthetic inputs first, then the Zed worked example as the integration acceptance gate.

### Per-step unit tests

- **`stars/period_test.go`** — single-star formula (Sol, Earth at 1.0 AU, M=1.0 → P.Years=1.000); multi-star formula (Zed AB I at orbit 7.2 around Aab+B with sumStellarMass = M(Aa)+M(Ab)+M(B) = 2.462; Zed B I at orbit 0.52 around B alone); Large-Planet variant adds bodyMassEarth × 0.000003; days/years convention (P.Years and P.Days both populated, conversion via × 365.25).

- **`worlds/sizing_terrestrial_test.go`** — every row of Basic Terrestrial Diameter table (`0`→0km, `R`→0km, `S`→600km, `1`→1600km, ..., `F`→24000km); every branch of the 1D selector (1-2 → second roll 1D → range 1-6; 3-4 → 2D → range 2-C; 5-6 → 2D+3 → range 5-F); reproduce specific Zed sizes from p.56 table.

- **`worlds/sizing_gasgiant_test.go`** — every category band (2-: GS via D3+D3 / 5×(1D+1); 3-4: GM via 1D+6 / 20×(3D-1); 5+: GL via 2D+6 / D3×50×(3D+4)); each DM (Brown Dwarf primary -1, M-V primary -1, Class VI primary -1, system spread <0.1 -1); the Large-GG mass clamp special case (initial mass ≥3,000⊕ from 3D≥15 → roll 2D-2, substitute mass = 4000 - 200×(2D-2)); reproduce Zed gas giants Aab IV (GLE, 1,200⊕), Aab V (GLC, 800⊕), AB III (GMB, 180⊕), Cab I (GS4, 10⊕).

- **`worlds/moons_test.go`** — every Moon Quantity row (Size 1-2 → 1D-5, Size 3-9 → 2D-8, Size A-F → 2D-6, GS# → 3D-7, GM#/GL# → 4D-6); each DM-1 condition (orbit <1.0, adjacent to companion-induced unavailability, adjacent to Close/Near unavailability range, in adjacent slot to outer Close/Near/Far range); negative result → 0 moons; exactly-0 → ring (`R`); Significant Moon Sizing branches (1-3 → S, 4-5 → D3-1, 6 → terrestrial Size-1-1D / GG Special); Gas Giant Special Moon Sizing branches (1-3 → 1D, 4-5 → 2D-2, 6 → 2D+4 with Size G → Small GG cascade, 12 sub-roll → Medium GG); terrestrial Size-1 parent → moon < parent forces Size S; "exactly 2 less than parent" 2D adjustment (2 → 1 less, 12 → twin world, otherwise → keep).

- **`worlds/designations_test.go`** — belt-skip example (Aab I, II, III, **PI**, IV, V, VI, VII, VIII — planet counter never advances on belt; belt counter independent); multi-group reset (after Aab VIII comes AB I); moon alphabet ordering (Aab IV a, b, c, d, e for 5 moons); composite (gas giant + significant moons + insignificant out-of-scope).

- **`worlds/profile_test.go`** — short form regex `\d+-\d+-\d+-\d+-[\d.]+`; long form per-star structure with `:` separators; Zed short = `"4-2-12-5-0.5"`; Zed long matches p.58 string exactly: `"Aab-5-T-T-T-P-G-G-T-T-T-0.5:B-2-T-T-0.5:AB-0-T-T-G-0.5:Cab-0-P-G-T-T-0.5"`.

- **`worlds/mainworld_test.go`** — HZ-window inclusion (orbit ∈ [HZCO−1.0, HZCO+1.0]); per-star HZCO (uses each `Group.HZCO()`); planet candidates (terrestrials in HZ); moon candidates (significant moons of HZ planets, including gas-giant moons per the Aab IV d / Aab V b /d pattern); Zed enumeration matches the four candidates the book lists (Aab VI plus moons Aab IV d, Aab V b, Aab V d).

- **`worlds/survey_form_test.go`** — per-section construction (header fields from `IISSClass23Header`, Stars from `stars.System` walk, Objects from `[]DetailedPlacement` walk, Notes/Comments passthrough); period rendering (Years < 0.05 → days with `d` suffix to 3 decimals, otherwise years with `y` suffix; thousands-separator commas for periods ≥1000y like `"3,598y"`); SAH cell rendering (terrestrial unrolled = `"<Size>??"`, gas giant = `"GS<Code>"`/`"GM<Code>"`/`"GL<Code>"`, belt = `"000"`); Notes column composition (`"<diameter>⊕, HZ, <moon SAHs>"` for gas giants in HZ, etc.).

- **`worlds/system_detail_test.go`** — façade pipeline: each step produces output the next step reads; `Detailed` slice 1:1 with `Placements`; `MarkHZ` runs after sizing so HZ tag covers all body kinds.

### Acceptance gate: `TestZed_FullDetail`

Drives `DetailSystem(r, sys, sp, header)` with the existing `composeZed()` plus a `roller.Scripted` issuing the exact dice sequence the book narrates across pp. 53–67. Header carries:

```go
IISSClass23Header{
    SectorLocation:  "Storr | 0602",
    InitialSurvey:   "207-568",
    LastUpdated:     "218-1061",
    IISSDesignation: "Zed (system)",
    Comments:        "*Further investigation required for mainworld candidate Aab IV d\nTentative system designation: 566-837",
}
```

Asserts:

1. **Sizes per p.56 table** — every placement's `SizeCode`, `DiameterKm` (or `GGClass`+`GGDiameterCode`+`DiameterEarth`+`MassEarth`) matches.
2. **Designations** — all 18 placements + 4 candidate moon designations match form p.63 exactly (Aab I, ..., Aab VIII for the retrograde slot; AB I-III; B I-II; Cab PI, Cab I-III; moons a-e per parent).
3. **Periods** — every Period column on form p.63 matches within ±0.001y or ±0.01d.
4. **Moons** — count and per-moon `SizeCode` match the p.58 results table, with the **documented book-typo carve-out**: form p.63 shows Aab IV d at SAH `566*` (Size 5), but the p.58 sizing table prints `2, S, S, S, S` for Aab IV. Treat the form as authoritative; assert d-moon `SizeCode = "5"`. Add a comment noting the inconsistency for the WBH errata memory.
5. **Profiles** — `ShortProfile == "4-2-12-5-0.5"`; `LongProfile == "Aab-5-T-T-T-P-G-G-T-T-T-0.5:B-2-T-T-0.5:AB-0-T-T-G-0.5:Cab-0-P-G-T-T-0.5"` exactly.
6. **HZ tags** — set on Aab IV, V, VI (Aab HZCO=3.3, range 2.3-4.3 covers orbits 3.1, 3.5, 4.1).
7. **Mainworld candidates** — `MainworldCandidates(sd)` returns 4 entries: planet Aab VI plus moons Aab IV d, Aab V b, Aab V d. Aab IV a (Size 2) is _not_ a candidate; verify by re-reading the procedure during plan.
8. **IISS form Stars table** — every row in `Survey.Stars` matches p.63 to declared tolerance (Mass ±0.001 solar, Temp ±5 K, Diameter ±0.001, Luminosity ±0.001, Orbit ±0.01, AU ±0.01, Ecc ±0.001, MAO ±0.01, HZCO ±0.01, PeriodStr matches book formatting).
9. **IISS form Objects table** — every row matches p.63 for `Primary`, `Designation`, `Orbit`, `AU`, `Period`, `Sub`, and (for non-HZ-candidate cells) `SAH`. HZ-candidate cells (Aab IV, V, VI rows; Aab IV a/d, Aab V b/d rows) render `SAH` with `?` placeholders for atmosphere/hydrographics digits — assert this explicitly. Eccentricity column matches form values to ±0.001 with carry-forward #3's age DM applied.

### Carry-forward test impact

Carry-forward #3 (plumb `ageGyr` into `RollPlanetEccentricities`) may shift values asserted by 2B's existing `TestZed_FullPlacement`. Plan task: re-derive expected eccentricity values for orbits <1.0 with Zed's `AgeGyr=6.336` applied, update 2B test assertions if needed (the underlying scripted dice rolls don't change; resulting eccentricities do, because the table-lookup index shifts by the new DM).

### Tolerances summary

| Column       | Tolerance                                          |
| ------------ | -------------------------------------------------- |
| Orbit#       | ±0.01                                              |
| AU           | ±0.01                                              |
| Mass (solar) | ±0.001                                             |
| Temp (K)     | ±5                                                 |
| Diameter rel | ±0.001                                             |
| Luminosity   | ±0.001                                             |
| Eccentricity | ±0.001                                             |
| MAO          | ±0.01                                              |
| HZCO         | ±0.01                                              |
| Period years | ±0.001                                             |
| Period days  | ±0.01                                              |
| Diameter km  | ±50                                                |
| Mass Earth   | exact integer for terrestrials; ±10 for gas giants |

### Test infrastructure

- `composeZed()` + `composeSol()` + `composeCorella()` from existing `worked_examples_test.go` reused unchanged.
- `roller.Scripted` reused; new test extends scripted dice by the 2C-specific roll sequence (sizing, moons).
- Scripted dice for 2C are sourced as follows: **terrestrial sizing rolls** are narrated in the p.56 Zed table ("Size Rolls" column); **gas giant sizing rolls** likewise narrated on p.56; **moon count rolls** are narrated in the p.56 moon-rolls table ("Moon Rolls" column). **Per-moon size rolls are not narrated** (only the resulting Size lists on p.58 / p.63 Notes). The plan back-derives per-moon dice that produce the form's authoritative Size lists, documenting each derivation inline in the test file.

## Open questions for future sub-projects

- **HZ atmosphere/hydrographics/temperature rolls.** WBH p.108 Temperature Roll table + atmosphere/hydrographics tables in the World Physical chapter (sub-project 3, pp. 69–146). Land here, then revisit 2C's `IISSClass23Form.Objects` rendering to fill the `?` placeholders.
- **Mainworld picker.** Needs the above. Add `PickMainworld(sd SystemDetail) MainworldCandidate` once SAH is rollable.
- **Insignificant moons** (p.58). Free-form by Referee, not procedural — likely never encoded.
- **Continuation Method** (Corella example, p.65). Still deferred from 2B; Corella worked-example test stays out until pre-existing-mainworld input plumbing is built.
- **Hill-sphere optional rule for moon DM-1** (p.56 sidebar). Defer until the Hill-sphere alternate orbit method is built.
- **CLI integration of the IISS Class II/III form.** Render the structured form as text (and eventually as Typst) — small follow-up after 2C lands.
- **Typst/markdown PDF rendering of the form.** Separate small project parallel to the Class 0/I form workflow.
- **Per-star baseline number + spread** (p.49 sidebar). Still deferred from 2B.
- **10–20% minimum-separation in compact systems** (p.49 sidebar). Still deferred from 2B.
- **Hill-sphere alternate orbit method** (pp. 40–41). Still deferred from 2A.
- **Post-stellar primary MAO.** Still deferred from 2A; lives with Special Circumstances chapter.
- **Carry-forward #5 from 2B** (PrefixRoll audit-trail collision-fallback fix). Skipped for 2C since the IISS form has no PrefixRoll column. Revisit only if a future consumer renders this field.
- **`Other`-descriptor wart in `stars.GenerateCompanionStar`.** Tracked since 2A; still its own small follow-up.
- **WBH p.58 sizing-table errata.** Aab IV moon list shows `2, S, S, S, S` but form p.63 Notes shows `200, S, S, 566*, S` (Size 5 d-moon). Test treats form as authoritative. Save a feedback memory entry under `feedback_wbh_p58_p63_inconsistency.md` after merging, joining `feedback_wbh_p19_p42_inconsistency.md`.

## Success criteria

- **API completeness.** All new types (`SizeCode`, `GasGiantClass`, `Moon`, `Period`, `DetailedPlacement`, `MainworldCandidate`, `SystemDetail`, `IISSClass23Form`, `IISSClass23Header`, `StarRow`, `ObjectRow`, `ParentInfo`, `TerrestrialSize`, `GasGiantSize`) and functions (`OrbitalPeriod`, `RollTerrestrialSize`, `RollGasGiantSize`, `CountMoons`, `SizeMoon`, `AssignPlanetDesignations`, `AssignMoonDesignations`, `MarkHZ`, `MainworldCandidates`, `ShortProfile`, `LongProfile`, `RenderIISSClass23`, `DetailSystem`) exist with GoDoc citing their WBH page or step.
- **Worked-example acceptance.** `TestZed_FullDetail` reproduces the book's Zed walkthrough across pp. 53–67 to declared tolerances, with two documented carve-outs: (a) HZ-candidate atmosphere/hydrographics cells render `?`; (b) Aab IV d-moon size matches form (`5`) not p.58 sizing table (`S`), per WBH errata note.
- **Carry-forward landings.** Items 1, 2, 3, 4 from 2B in place. `TestSol_GenerateSystemPlacement` smoke test passes. `GenerateSystemPlacement` errors are wrapped at all five callsites with `fmt.Errorf("worlds: <step>: %w", err)`. `RollPlanetEccentricities` accepts `ageGyr` and applies WBH p.27 sub-1.0/age>1Gyr DM; existing `TestZed_FullPlacement` updated if eccentricity values shift.
- **Test specificity.** `TestRollEccentricity_ExtraDM` and `TestRollPlanetEccentricities_AppliesAnomalyDM` assert specific resulting eccentricity values, not just "DM had effect."
- **Build green.** A fresh checkout of `` runs `just check && just test` clean (gofumpt + go vet + golangci-lint v2.12.1 + `go test -race ./...`).
- **Source traceability.** A reader with the WBH open can match every exported symbol in `stars/period.go` and the new `worlds/*.go` files to a specific page or step in WBH pp. 53–67.
- **Memory hygiene.** After merge, update `MEMORY.md` and `project_world_builder_resume.md` to reflect: 2C complete, sub-project 3 (World Physical, pp. 69–146) is next, and add a brief feedback memory for the p.58/p.63 book-data inconsistency.
