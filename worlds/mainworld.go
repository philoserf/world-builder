package worlds

// MainworldCandidate describes one body eligible to be the system's
// mainworld per WBH p.58. 2C enumerates candidates; the picker
// (Referee judgment from atmosphere/hydrographics rolls) lands with
// sub-project 3 (World Physical, pp. 69-146).
type MainworldCandidate struct {
	Designation       string
	SizeCode          SizeCode
	DiameterKm        float64
	Orbit             float64 // host planet's orbit (for moons, the parent's)
	HostStarGroup     string  // "Aab", "B", "Cab"
	IsMoon            bool
	ParentDesignation string // "" for planet candidates; "Aab IV" for moons
}

// MarkHZ sets DetailedPlacement.HZ when the placement's orbit lies
// within HZCO ± 1.0 of the placement's host group, per WBH p.58:
// "The habitable zone is generally considered to be +/ 1.0 Orbit#s
// from the HZCO."
//
// Uses Group.HZCO() from 2B (defined in worlds/group_hzco.go).
//
// Mutates DetailedPlacement.HZ in place.
func MarkHZ(dps []DetailedPlacement) error {
	for i := range dps {
		if dps[i].Body == BodyEmpty {
			continue
		}
		hzco := dps[i].Group.HZCO()
		o := dps[i].Orbit
		if o >= hzco-1.0 && o <= hzco+1.0 {
			dps[i].HZ = true
		}
	}
	return nil
}

// MainworldCandidates returns the subset of bodies eligible to be the
// mainworld per WBH p.58:
//
//   - Terrestrial planets in the HZ
//   - Significant moons (any Size) of any HZ planet, including
//     gas-giant moons (per the Zed example: Aab IV d, Aab V b, Aab V d)
//
// Belts and gas giants themselves are NOT candidates.
//
// Returned in DetailedPlacement order, with each planet's moons listed
// before the planet itself when both are candidates (matches Zed form
// p.63 Object table ordering).
//
// Selection requires World Physical (sub-project 3); 2C provides
// enumeration only.
func MainworldCandidates(sd SystemDetail) []MainworldCandidate {
	out := []MainworldCandidate{}
	for _, dp := range sd.Detailed {
		if !dp.HZ {
			continue
		}
		// Moons first (form ordering: parent's moon rows precede the
		// parent's planet row in the Notes column).
		for _, m := range dp.Moons {
			out = append(out, MainworldCandidate{
				Designation:       m.Designation,
				SizeCode:          m.SizeCode,
				DiameterKm:        m.DiameterKm,
				Orbit:             dp.Orbit, // parent's orbit
				HostStarGroup:     dp.Group.Designation,
				IsMoon:            true,
				ParentDesignation: dp.Designation,
			})
		}
		// Then the planet itself, if it's terrestrial.
		if dp.Body == BodyTerrestrial {
			out = append(out, MainworldCandidate{
				Designation:   dp.Designation,
				SizeCode:      dp.SizeCode,
				DiameterKm:    dp.DiameterKm,
				Orbit:         dp.Orbit,
				HostStarGroup: dp.Group.Designation,
				IsMoon:        false,
			})
		}
	}
	return out
}

// pickMainworld returns the designation of the auto-picked mainworld per
// WBH p.134. Priority chain (first match wins):
//
//  1. Bodies with native sophonts (extant or extinct); among these,
//     highest Habitability; tiebreaker: highest ResourceRating;
//     final tiebreaker: iteration order.
//  2. Highest Habitability among non-sophont bodies; tiebreakers same.
//  3. Highest ResourceRating across terrestrials, moons, and belts if no
//     body has Habitability > 0.
//  4. First terrestrial-or-belt body in iteration order.
//
// Iterates detailed[i] (planets and belts), dp.Moons[j] (moons). Returns ""
// if no eligible body exists. Belts contribute resource only (no
// habitability, no sophonts) — they win priorities 3 or 4 when
// terrestrials don't qualify, satisfying WBH p.134's "best refuelling
// location" criterion.
func pickMainworld(detailed []DetailedPlacement) string {
	type candidate struct {
		designation  string
		habitability int
		resource     int
		hasSophont   bool
	}

	var candidates []candidate

	collect := func(designation string, bodyType BodyType, h *Habitability, b *Biology, belt *BeltDetails) {
		if bodyType != BodyTerrestrial && bodyType != BodyPlanetoidBelt {
			return
		}
		c := candidate{designation: designation}
		if h != nil {
			c.habitability = h.Rating
		}
		if b != nil {
			c.resource = b.ResourceRating
			c.hasSophont = b.HasNativeSophont || b.HadExtinctSophont
		}
		if bodyType == BodyPlanetoidBelt && belt != nil {
			c.resource = belt.ResourceRating
		}
		candidates = append(candidates, c)
	}

	for i := range detailed {
		dp := &detailed[i]
		collect(dp.Designation, dp.Body, dp.Habitability, dp.Biology, dp.Belt)
		// Moons are treated as terrestrials for mainworld-pick purposes;
		// Zed Prime (WBH worked example) is itself a moon.
		for j := range dp.Moons {
			m := &dp.Moons[j]
			collect(m.Designation, BodyTerrestrial, m.Habitability, m.Biology, nil)
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	better := func(i, best int) bool {
		return candidates[i].habitability > candidates[best].habitability ||
			(candidates[i].habitability == candidates[best].habitability &&
				candidates[i].resource > candidates[best].resource)
	}

	// Priority 1: sophont present (extant or extinct).
	best := -1
	for i, c := range candidates {
		if !c.hasSophont {
			continue
		}
		if best == -1 || better(i, best) {
			best = i
		}
	}
	if best != -1 {
		return candidates[best].designation
	}

	// Priority 2: highest Habitability > 0.
	for i, c := range candidates {
		if c.habitability == 0 {
			continue
		}
		if best == -1 || better(i, best) {
			best = i
		}
	}
	if best != -1 {
		return candidates[best].designation
	}

	// Priority 3: highest ResourceRating > 0.
	for i, c := range candidates {
		if c.resource == 0 {
			continue
		}
		if best == -1 || candidates[i].resource > candidates[best].resource {
			best = i
		}
	}
	if best != -1 {
		return candidates[best].designation
	}

	// Priority 4: first terrestrial-or-belt body in iteration order.
	return candidates[0].designation
}
