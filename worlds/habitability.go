// Package worlds — per-body habitability rating per WBH p.132-133
// (sub-project 3B-final).
package worlds

import "strings"

// Habitability — a per-body habitability rating for Terragens per WBH
// p.132-133. Computed by Step 5G for any non-empty terrestrial body
// (and HZ-planet moons).
//
// Range: 0-12. The book theoretically allows higher but treats 12 as
// "very unlikely" and clamps negative results to 0.
//
// Ratings interpretation (WBH p.133):
//
//	0       — Actively hostile world: not survivable without specialised equipment
//	1-2     — Barely habitable: full protective equipment needed
//	3-5     — Marginally survivable with proper equipment
//	6-7     — Regionally habitable: may require acclimation
//	8-9     — Suitable for human habitation with minimal equipment or acclimation
//	10-12   — Terra-equivalent garden world (10/A is the Terran baseline)
type Habitability struct {
	Rating int

	// Notes is a referee-color string visible in the Class IV-P form's
	// Habitability section. Populated by ComputeHabitability from the
	// WBH p.132 DM table's description column for whichever rules fired,
	// joined with "; ". Empty when no DMs fire (Terra-equivalent baseline).
	Notes string
}

// ComputeHabitability per WBH p.132: 10 + DMs, clamped to [0, 12].
// Deterministic — no dice. Operates on body's current Atmosphere /
// Hydrographics / Temperature / Physical / SizeCode / TidalLock fields.
//
// Returns Habitability{Rating: 0} if body is nil. For bodies with
// missing pointer fields, the corresponding DMs are skipped (treated
// as 0) — defensive but documented as caller's responsibility.
//
// Skipped: low-oxygen-taint DM-2 deferred per spec Q3-a (taint
// typology not yet modeled).
func ComputeHabitability(body *DetailedPlacement) Habitability {
	if body == nil {
		return Habitability{Rating: 0}
	}
	var notes []string
	addNote := func(s string) {
		if s != "" {
			notes = append(notes, s)
		}
	}

	sizeDM, sizeNote := habitabilitySizeDM(SizeAsInt(body.SizeCode))
	addNote(sizeNote)
	atmDM, atmNote := habitabilityAtmDM(body)
	addNote(atmNote)
	hydroDM, hydroNote := habitabilityHydroDM(body)
	addNote(hydroNote)
	tidalDM, tidalNote := habitabilityTidalLockDM(body)
	addNote(tidalNote)
	tempDM, tempNotes := habitabilityTempDM(body)
	for _, n := range tempNotes {
		addNote(n)
	}
	gravDM, gravNote := habitabilityGravityDM(body)
	addNote(gravNote)

	dm := sizeDM + atmDM + hydroDM + tidalDM + tempDM + gravDM
	rating := min(max(10+dm, 0), 12)
	return Habitability{Rating: rating, Notes: strings.Join(notes, "; ")}
}

// habitabilitySizeDM per WBH p.132 size-DM table.
// Returns the DM and the book's description-column phrase (empty when no
// DM fires, i.e., size 5–8).
func habitabilitySizeDM(size int) (int, string) {
	switch {
	case size <= 4:
		return -1, "Limited surface area"
	case size >= 9:
		return +1, "Additional surface area"
	}
	return 0, ""
}

// habitabilityAtmDM per WBH p.132 atmosphere-DM table.
// nil Atmosphere is treated as atm code 0 (vacuum) → DM-8.
// Returns the DM and the book's description-column phrase (empty when
// no DM fires, i.e., atm 6 baseline or unhandled codes).
func habitabilityAtmDM(body *DetailedPlacement) (int, string) {
	atmCode := 0
	if body.Atmosphere != nil {
		atmCode = body.Atmosphere.Code
	}
	switch atmCode {
	case 0, 1, 10: // 0, 1, A
		return -8, "Non-breathable atmosphere"
	case 2, 14: // 2, E
		return -4, "Very thin, tainted, or thin, low atmospheres"
	case 3, 13: // 3, D
		return -3, "Very thin or very dense atmosphere"
	case 4, 9:
		return -2, "Tainted thin or dense atmospheres"
	case 5, 7, 8:
		return -1, "Thin, taint (standard), or dense Atmospheres"
	case 6:
		return 0, "" // baseline
	case 11: // B
		return -10, "Hostile Atmosphere"
	case 12, 15: // C, F+
		return -12, "Very hostile Atmosphere"
	}
	return 0, ""
}

// habitabilityHydroDM per WBH p.132 hydrographics-DM table.
// nil Hydrographics is treated as Hydro code 0 → DM-4.
// Returns the DM and the book's description-column phrase (empty when
// no DM fires, i.e., Hydro 4–8).
func habitabilityHydroDM(body *DetailedPlacement) (int, string) {
	hydroCode := 0
	if body.Hydrographics != nil {
		hydroCode = body.Hydrographics.Code
	}
	switch {
	case hydroCode == 0:
		return -4, "Lack of accessible liquid water"
	case hydroCode >= 1 && hydroCode <= 3:
		return -2, "Desert conditions prevalent"
	case hydroCode == 9:
		return -1, "Little useable land surface area"
	case hydroCode >= 10:
		return -2, "Very little useable land surface area"
	}
	return 0, "" // 4-8
}

// habitabilityTidalLockDM per WBH p.132: "Solar tidally locked (1:1)
// world" → DM-2. Detection: TidalLock.IsTwilightZone (which is true
// only when Case == PlanetToStar AND LockRatio == "1:1").
// Returns the DM and the book's description-column phrase.
func habitabilityTidalLockDM(body *DetailedPlacement) (int, string) {
	if body.TidalLock == nil {
		return 0, ""
	}
	if body.TidalLock.IsTwilightZone {
		return -2, "Very little useable land surface area"
	}
	return 0, ""
}

// habitabilityTempDM per WBH p.132 temperature-DM table. Multiple
// sub-conditions (HighK, MeanK bands, LowK) can fire independently,
// so this helper returns a slice of fired-condition descriptions.
// Returns (0, nil) when Temperature is nil (defensive).
//
// Note: HighK > 323 and MeanK > 323 are strict (323 itself is in the
// [304, 323] band → -2, NOT in the >323 band → -4). Per WBH p.132 footnote,
// "use worst at edges" — but the bands as written are unambiguous at 323.
func habitabilityTempDM(body *DetailedPlacement) (int, []string) {
	if body.Temperature == nil {
		return 0, nil
	}
	dm := 0
	var notes []string
	t := body.Temperature
	if t.HighK > 323 {
		dm += -2
		notes = append(notes, "Too hot at times")
	}
	if t.HighK > 0 && t.HighK < 279 {
		dm += -2
		notes = append(notes, "Too cold all of the time")
	}
	if t.MeanK > 323 {
		dm += -4
		notes = append(notes, "Too hot most of the time")
	} else if t.MeanK >= 304 && t.MeanK <= 323 {
		dm += -2
		notes = append(notes, "Too hot most of the time")
	}
	if t.MeanK > 0 && t.MeanK < 273 {
		dm += -2
		notes = append(notes, "Too cold most of the time")
	}
	if t.LowK > 0 && t.LowK < 200 {
		dm += -2
		notes = append(notes, "Much too cold some of the time")
	}
	return dm, notes
}

// habitabilityGravityDM per WBH p.132 gravity-DM table.
//
// WBH p.132 has overlapping bands (0.2-0.7 and 0.4-0.7). Per the worked
// example for Zed Prime (gravity 0.66 → DM-1, NOT -2), the narrower band
// wins. Documented as a WBH inconsistency (footnote contradicts worked
// example); implementation follows the worked example.
//
// Undefined gravity (Physical nil): per WBH "+1 - |6 - Size|" — the
// book gives this fallback formula but no description column entry,
// so the note is empty.
//
// Returns the DM and the book's description-column phrase.
func habitabilityGravityDM(body *DetailedPlacement) (int, string) {
	if body.Physical == nil {
		size := SizeAsInt(body.SizeCode)
		diff := 6 - size
		if diff < 0 {
			diff = -diff
		}
		return 1 - diff, "" // no book description for the fallback formula
	}
	g := body.Physical.Gravity
	switch {
	case g < 0.2:
		return -4, "Unhealthy low gravity levels"
	case g >= 0.7 && g <= 0.9:
		return +1, "Gravity very comfortable"
	case g >= 0.4 && g < 0.7:
		return -1, "Low gravity" // narrower band; wins over 0.2-0.7 per Q3-a
	case g >= 0.2 && g < 0.4:
		return -2, "Very low gravity" // residual of 0.2-0.7
	case g > 1.1 && g <= 1.4:
		return -1, "Gravity somewhat high"
	case g > 1.4 && g <= 2.0:
		return -3, "Gravity uncomfortably high"
	case g > 2.0:
		return -6, "Gravity too high for acclimation"
	}
	return 0, "" // 0.9-1.1 (Earth-like baseline)
}

// HabitabilityRatingName returns the WBH p.133 banded label for a habitability
// rating. Source: WBH p.133 "Habitability Rating" remarks table — book-quoted
// descriptions copied verbatim from the bands.
func HabitabilityRatingName(r int) string {
	switch {
	case r <= 0:
		return "Actively hostile" // book: "Actively hostile world: not survivable without specialised equipment"
	case r <= 2:
		return "Barely habitable" // book: "Barely habitable world: full protective equipment often needed"
	case r <= 5:
		return "Marginally survivable" // book: "Marginally survivable world with proper equipment"
	case r <= 7:
		return "Regionally habitable" // book: "Regionally habitable world: may require acclimation"
	case r <= 9:
		return "Suitable" // book: "Suitable for human habitation with minimal equipment or acclimation"
	default: // 10+
		return "Garden world" // book: "Terra-equivalent garden world"
	}
}

// ResourceRatingName returns the WBH p.131 banded label for a resource rating.
// Source: WBH p.131 "Resource Rating" remarks — book-quoted summaries copied
// from the band-table prose.
func ResourceRatingName(r int) string {
	switch {
	case r <= 2:
		return "No economically extractable resources" // book quote
	case r <= 5:
		return "Marginal" // book: "Marginal at best; avoided by most corporations"
	case r <= 8:
		return "Worthwhile" // book: "Worthwhile with considerable effort"
	case r <= 10: // 9-A
		return "Priority target" // book: "Priority targets for both corporations and individual prospectors"
	default: // 11-12 (B-C)
		return "Resource rush" // book: "Liable to experience a resource 'rush'"
	}
}
