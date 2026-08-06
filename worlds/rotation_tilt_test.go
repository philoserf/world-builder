package worlds_test

import (
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
	"github.com/philoserf/world-builder/worlds"
)

// TestZed_ApplyStage4 exercises the Stage-4 orchestrator (rotation /
// tilt / tide) end-to-end. Seeded shape-invariant per spike-findings § 2.
func TestZed_ApplyStage4(t *testing.T) {
	t.Parallel()

	for iter := range 25 {
		seed := int64(iter)
		r := roller.NewSeeded(seed)
		sys := composeZed()

		sp, err := worlds.GenerateSystemPlacement(r, sys)
		if err != nil {
			t.Fatalf("seed %d: GenerateSystemPlacement: %v", seed, err)
		}

		u := &worlds.Universe{System: sys, Placement: sp}
		if err := worlds.ApplyDetailFrontEnd(r, u); err != nil {
			t.Fatalf("seed %d: ApplyDetailFrontEnd: %v", seed, err)
		}

		if err := worlds.ApplyBodyPhysical(r, u); err != nil {
			t.Fatalf("seed %d: ApplyBodyPhysical: %v", seed, err)
		}

		if err := worlds.ApplyBeltDetails(r, u); err != nil {
			t.Fatalf("seed %d: ApplyBeltDetails: %v", seed, err)
		}

		if err := worlds.ApplyMoonRefinement(r, u); err != nil {
			t.Fatalf("seed %d: ApplyMoonRefinement: %v", seed, err)
		}

		if err := worlds.ApplyRotationTilt(r, u); err != nil {
			t.Fatalf("seed %d: ApplyRotationTilt: %v", seed, err)
		}

		// Every non-empty body has DayLength + AxialTilt + TidalEffects
		// populated. Per anti-pattern A.1, every child does too.
		for i, body := range u.Detail.Bodies {
			if body.Kind == worlds.BodyEmpty {
				continue
			}

			if !body.HasDayLength() {
				t.Errorf("seed %d: bodies[%d] (%s) missing DayLength", seed, i, body.Designation)
			}

			if !body.HasAxialTilt() {
				t.Errorf("seed %d: bodies[%d] (%s) missing AxialTilt", seed, i, body.Designation)
			}

			if !body.HasTidalEffects() {
				t.Errorf("seed %d: bodies[%d] (%s) missing TidalEffects", seed, i, body.Designation)
			}

			for j, child := range body.Children {
				if !child.HasDayLength() {
					t.Errorf("seed %d: bodies[%d].Children[%d] (%s) missing DayLength (moon-path silent-zero?)",
						seed, i, j, child.Designation)
				}

				if !child.HasAxialTilt() {
					t.Errorf("seed %d: bodies[%d].Children[%d] (%s) missing AxialTilt",
						seed, i, j, child.Designation)
				}

				if !child.HasTidalEffects() {
					t.Errorf("seed %d: bodies[%d].Children[%d] (%s) missing TidalEffects",
						seed, i, j, child.Designation)
				}
			}
		}
	}
}

// TestApplyRotationTilt_TwoPass verifies that sub-stage 3 evaluates moons
// before planets so that hasLockedMoon can fire for the Planet→Moon case.
//
// Setup: one terrestrial planet (Size 8, orbit 5.0 AU) with one significant
// moon (Size 5, OrbitPD 10). The scripted dice are arranged so that:
//   - Moon tidal-lock roll: 2D=4 + DM+8 = 12 → 1:1 lock (verification 2D=8, lock stands).
//   - Planet tidal-lock roll: 2D=5 + DM+0 = 5 → DayLengthMultiplier×3 via Planet→Moon case.
//
// With single-pass (planet first) the moon's TidalLock is nil when the planet
// is evaluated, hasLockedMoon returns false, and Planet→Moon is absent — the
// planet gets no tidal lock at all (Case=None). With two-pass (moon first)
// the moon's 1:1 lock is recorded before the planet is evaluated, so
// Planet→Moon fires and planet.TidalLock.Case == TidalLockCasePlanetToMoon.
//
// WBH p.107: Planet→Moon precondition — a moon must already be locked.
func TestApplyRotationTilt_TwoPass(t *testing.T) {
	t.Parallel()

	// Moon orbits planet at 10 planetary diameters, period 168 h (~7 days).
	moon := &worlds.Body{
		Kind:         worlds.BodyMoon,
		SizeCode:     "5",
		OrbitPD:      10,
		PeriodHours:  168.0,
		Eccentricity: 0.0,
	}

	// Planet: terrestrial, Size 8, mass 1.5 M⊕, orbit 5.0 AU, year ~8766 h.
	// MassEarth drives moonToPlanetDMs — needs to be non-zero.
	planet := worlds.Body{
		Kind:      worlds.BodyTerrestrial,
		SizeCode:  "8",
		MassEarth: 1.5,
		Orbit:     5.0,
		Period:    worlds.Period{Hours: 8766},
		Children:  []*worlds.Body{moon},
	}

	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 0.0}}
	u := &worlds.Universe{
		System: sys,
		Detail: worlds.SystemDetail{
			Bodies: []worlds.Body{planet},
		},
	}

	// Dice sequence for ApplyRotationTilt with one planet + one moon:
	//
	// Sub-stage 1 — DayLength (planet, then moon):
	//   planet: 2D=2 → (2-2)×4+2+1D=2+2=4 h; precision: 1D=1,d10=0,1D=1,d10=0
	//   moon:   2D=2 → same; precision: 1D=1,d10=0,1D=1,d10=0
	//
	// Sub-stage 2 — AxialTilt (planet, then moon):
	//   planet: 2D=2 (range 2-4) → 1D=1 → (1-1)/50 = 0°
	//   moon:   2D=2, 1D=1 → 0°
	//
	// Sub-stage 3 — TidalLock, TWO-PASS (moon first, then planet):
	//   moon:   2D=4 + DM+8 = 12 → 1:1; verification 2D=8 (not natural 12, lock stands)
	//   planet: 2D=5 + DM+0 = 5 → DayLengthMultiplier×3 (Planet→Moon case)
	//
	// Sub-stage 4 — SurfaceTidalEffects: no dice consumed.
	r := roller.NewScripted(
		// planet DayLength
		2, 2, 1, 0, 1, 0,
		// moon DayLength
		2, 2, 1, 0, 1, 0,
		// planet AxialTilt
		2, 1,
		// moon AxialTilt
		2, 1,
		// moon TidalLock: 2D=4+DM8=12, verification 2D=8
		4, 8,
		// planet TidalLock: 2D=5+DM0=5
		5,
	)

	if err := worlds.ApplyRotationTilt(r, u); err != nil {
		t.Fatalf("ApplyRotationTilt: %v", err)
	}

	got := u.Detail.Bodies[0]
	if got.TidalLock == nil {
		t.Fatal("planet.TidalLock is nil: Planet→Moon case did not fire (single-pass ordering bug?)")
	}

	if got.TidalLock.Case != worlds.TidalLockCasePlanetToMoon {
		t.Errorf("planet.TidalLock.Case = %v, want TidalLockCasePlanetToMoon", got.TidalLock.Case)
	}

	// Bonus: verify the moon received its 1:1 lock in pass 1.
	if moon.TidalLock == nil {
		t.Fatal("moon.TidalLock is nil: moon was not evaluated in pass 1")
	}

	if moon.TidalLock.LockRatio != "1:1" {
		t.Errorf("moon.TidalLock.LockRatio = %q, want 1:1", moon.TidalLock.LockRatio)
	}
}
