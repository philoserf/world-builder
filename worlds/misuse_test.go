package worlds_test

import (
	"testing"

	"wbh/roller"
	"wbh/stars"
	"wbh/worlds"
)

// TestMisuse_ConvergeClimate_GGSkipped verifies ConvergeClimate's
// eligibility check short-circuits cleanly for gas giants (climate is
// terrestrials-only). Per harness.md § Misuse-path tests, ConvergeClimate
// must skip non-HZ bodies and gas giants without panicking.
func TestMisuse_ConvergeClimate_GGSkipped(t *testing.T) {
	t.Parallel()
	body := &worlds.Body{
		Kind:        worlds.BodyGasGiant,
		GGClass:     worlds.GasGiantMedium,
		HZ:          true, // even if HZ-tagged, gas giants skip climate
		Designation: "test-gg",
	}
	r := roller.NewSeeded(42)
	if err := worlds.ConvergeClimate(r, body, stars.System{Primary: stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.0,
		Diameter:        1.0,
		Temperature:     5772,
		AgeGyr:          4.6,
	})}); err != nil {
		t.Errorf("ConvergeClimate on GG returned error: %v", err)
	}
	if body.HasAtmosphere() {
		t.Error("GG body got an atmosphere from ConvergeClimate (should be skipped)")
	}
}

// TestMisuse_ConvergeClimate_NonHZSkipped verifies non-HZ bodies skip
// climate without panicking.
func TestMisuse_ConvergeClimate_NonHZSkipped(t *testing.T) {
	t.Parallel()
	body := &worlds.Body{
		Kind:        worlds.BodyTerrestrial,
		HZ:          false,
		SizeCode:    "8",
		Designation: "test-cold",
	}
	r := roller.NewSeeded(42)
	if err := worlds.ConvergeClimate(r, body, stars.System{}); err != nil {
		t.Errorf("ConvergeClimate on non-HZ body returned error: %v", err)
	}
	if body.HasAtmosphere() {
		t.Error("non-HZ body got an atmosphere (should be skipped)")
	}
}

// TestMisuse_PickMainworld_EmptyBodies — pickMainworld is a private
// helper; AggregateSystem is the public entry. Test it with a Universe
// that has no bodies.
func TestMisuse_AggregateSystem_EmptyBodies(t *testing.T) {
	t.Parallel()
	u := &worlds.Universe{}
	worlds.AggregateSystem(u) // must not panic
	if u.Detail.MainworldDesignation != "" {
		t.Errorf("MainworldDesignation = %q, want empty for empty universe",
			u.Detail.MainworldDesignation)
	}
}

// TestMisuse_ComputeHabitability_NoTemperature verifies ComputeHabitability
// returns a sane rating when Temperature is nil (vacuum / non-HZ body
// path). Per habitabilityApplies, vacuum worlds DO get a habitability —
// the function must not panic on nil pointers.
func TestMisuse_ComputeHabitability_NoTemperature(t *testing.T) {
	t.Parallel()
	body := &worlds.Body{
		Kind:     worlds.BodyTerrestrial,
		SizeCode: "5",
		// no Atmosphere, no Hydrographics, no Temperature — vacuum cold rock.
	}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ComputeHabitability panicked on nil-pointer body: %v", r)
		}
	}()
	h := worlds.ComputeHabitability(body)
	// Just verify it returned without panic. Rating may be any value.
	_ = h
}

// TestMisuse_RollGasMix_EmptyColumn verifies RollGasMix doesn't
// panic on empty column-letter input — the misuse path flagged in
// harness.md § Misuse-path tests (RollGasMix (a)).
func TestMisuse_RollGasMix_EmptyColumn(t *testing.T) {
	t.Parallel()
	r := roller.NewSeeded(42)
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("RollGasMix panicked on empty column letter: %v", rec)
		}
	}()
	_, _ = worlds.RollGasMix(r, "", "", worlds.TempTemperate, "5")
}
