package worlds

// PreTidalLockSnapshot captures the body fields that ApplyTidalLockEffect
// can mutate, for use by the WBH p.106 atmosphere-DM re-evaluation
// cascade. The Roller cannot be rewound, so the re-eval pass consumes
// fresh dice; this snapshot restores the body state that those fresh
// rolls then operate on.
type PreTidalLockSnapshot struct {
	Eccentricity float64
	AxialTilt    *AxialTilt // value copy; nil if body had none
	DayLength    *DayLength // value copy; nil if body had none
}

// CapturePreTidalLockSnapshot snapshots the mutable fields of body
// immediately before ApplyTidalLockEffect runs. The returned snapshot
// is independent of the body — mutating snapshot fields does not
// affect the body, and vice versa.
func CapturePreTidalLockSnapshot(body *Body) PreTidalLockSnapshot {
	snap := PreTidalLockSnapshot{Eccentricity: body.Eccentricity}
	if body.AxialTilt != nil {
		v := *body.AxialTilt
		snap.AxialTilt = &v
	}

	if body.DayLength != nil {
		v := *body.DayLength
		snap.DayLength = &v
	}

	return snap
}

// RestoreInto writes the snapshot back into body, replacing the current
// values. Nil fields in the snapshot restore the body field to nil.
func (s PreTidalLockSnapshot) RestoreInto(body *Body) {
	body.Eccentricity = s.Eccentricity
	if s.AxialTilt != nil {
		v := *s.AxialTilt
		body.AxialTilt = &v
	} else {
		body.AxialTilt = nil
	}

	if s.DayLength != nil {
		v := *s.DayLength
		body.DayLength = &v
	} else {
		body.DayLength = nil
	}
}
