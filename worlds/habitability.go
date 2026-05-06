// Package worlds — per-body habitability rating per WBH p.132-133
// (sub-project 3B-final).
package worlds

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
	// Habitability section (e.g., "High temperatures hinder habitability").
	// Currently always empty — populated by future referee-feature carry-forward.
	Notes string
}
