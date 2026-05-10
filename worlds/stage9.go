package worlds

// ApplyHabitability computes the per-body Habitability rating per WBH
// p.132 for every terrestrial body in the universe. Stage-9
// orchestrator.
//
// Body filter: terrestrials only (skip belts / GGs / empty).
// Atmosphere optional — vacuum worlds get a rating via the atm-0 row.
// Per anti-pattern A.1, every terrestrial moon is walked alongside
// its parent (Zed Prime is a moon of a gas giant and is the canonical
// Habitability-7 mainworld).
//
// Pure function — no rolls. Per docs/pass-2/api-surface.md § Stage 9.
func ApplyHabitability(u *Universe) {
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if habitabilityApplies(body) {
			h := ComputeHabitability(body)
			body.Habitability = &h
		}
		for _, child := range body.Children {
			if !habitabilityApplies(child) {
				continue
			}
			ch := ComputeHabitability(child)
			child.Habitability = &ch
		}
	}
}

// habitabilityApplies reports whether body should receive a
// Habitability struct. True for terrestrial bodies (atmosphere
// optional — vacuum worlds get a rating); false for empty, belts,
// gas giants.
func habitabilityApplies(body *Body) bool {
	if body == nil || body.Kind == BodyEmpty {
		return false
	}
	if body.Kind == BodyGasGiant || body.Kind == BodyPlanetoidBelt {
		return false
	}
	if body.SizeCode == "0" {
		return false
	}
	return true
}
