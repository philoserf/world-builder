package worlds_test

import (
	"testing"

	"wbh/iiss"
	"wbh/roller"
	"wbh/stars"
	"wbh/worlds"
)

// TestMisuse_ApplyClimatePasses_GGSkipped verifies ApplyClimatePasses's
// eligibility check short-circuits cleanly for gas giants (climate is
// terrestrials-only). Per harness.md § Misuse-path tests, ApplyClimatePasses
// must skip non-HZ bodies and gas giants without panicking.
func TestMisuse_ApplyClimatePasses_GGSkipped(t *testing.T) {
	t.Parallel()
	body := &worlds.Body{
		Kind:        worlds.BodyGasGiant,
		GGClass:     worlds.GasGiantMedium,
		HZ:          true, // even if HZ-tagged, gas giants skip climate
		Designation: "test-gg",
	}
	r := roller.NewSeeded(42)
	if err := worlds.ApplyClimatePasses(r, body, stars.System{Primary: stars.Compose(stars.ComposeOpts{
		Kind:            stars.KindMainSequence,
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.0,
		Diameter:        1.0,
		Temperature:     5772,
		AgeGyr:          4.6,
	})}); err != nil {
		t.Errorf("ApplyClimatePasses on GG returned error: %v", err)
	}
	if body.HasAtmosphere() {
		t.Error("GG body got an atmosphere from ApplyClimatePasses (should be skipped)")
	}
}

// TestMisuse_ApplyClimatePasses_NonHZSkipped verifies non-HZ bodies skip
// climate without panicking.
func TestMisuse_ApplyClimatePasses_NonHZSkipped(t *testing.T) {
	t.Parallel()
	body := &worlds.Body{
		Kind:        worlds.BodyTerrestrial,
		HZ:          false,
		SizeCode:    "8",
		Designation: "test-cold",
	}
	r := roller.NewSeeded(42)
	if err := worlds.ApplyClimatePasses(r, body, stars.System{}); err != nil {
		t.Errorf("ApplyClimatePasses on non-HZ body returned error: %v", err)
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

// TestMisuse_RollAtmoCode_SizeZero — SizeCode "0" (belt) — function
// should return zero atm code (or error) but not panic.
func TestMisuse_RollAtmoCode_SizeZero(t *testing.T) {
	t.Parallel()
	r := roller.NewSeeded(42)
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("RollAtmoCode panicked on SizeCode \"0\": %v", rec)
		}
	}()
	_, _ = worlds.RollAtmoCode(r, "0", 0.0)
}

// TestMisuse_RollAtmoCode_NegativeOffset — negative offset should not panic.
func TestMisuse_RollAtmoCode_NegativeOffset(t *testing.T) {
	t.Parallel()
	r := roller.NewSeeded(42)
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("RollAtmoCode panicked on negative offset: %v", rec)
		}
	}()
	_, _ = worlds.RollAtmoCode(r, "5", -2.5)
}

// TestMisuse_RollTotalPressure_OutOfTable — atmCode 100 is outside
// the WBH p.79 table; should return error or sensible zero, not panic.
func TestMisuse_RollTotalPressure_OutOfTable(t *testing.T) {
	t.Parallel()
	r := roller.NewSeeded(42)
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("RollTotalPressure panicked on atmCode 100: %v", rec)
		}
	}()
	_, _ = worlds.RollTotalPressure(r, 100, "")
}

// TestMisuse_RollOxygenFraction_NegativeAge — negative ageGyr should
// not panic; sensible behaviour is to clamp at 0 or return a zero
// fraction.
func TestMisuse_RollOxygenFraction_NegativeAge(t *testing.T) {
	t.Parallel()
	r := roller.NewSeeded(42)
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("RollOxygenFraction panicked on negative ageGyr: %v", rec)
		}
	}()
	_, _ = worlds.RollOxygenFraction(r, -1.0)
}

// TestMisuse_RollCorrosiveInsidiousSubtype_WrongAtmCode — atm code 5
// (oxygen variant) shouldn't be passed to a corrosive/insidious-only
// procedure; current behaviour should not panic.
func TestMisuse_RollCorrosiveInsidiousSubtype_WrongAtmCode(t *testing.T) {
	t.Parallel()
	r := roller.NewSeeded(42)
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("RollCorrosiveInsidiousSubtype panicked on atm code 5: %v", rec)
		}
	}()
	_, _ = worlds.RollCorrosiveInsidiousSubtype(r, "5", 3.0, 3.0, false, false)
}

// TestMisuse_GenerateBodyPhysical_SizeS — SizeCode "S" (small body)
// is technically out of the WBH p.71 Composition table range. Should
// not panic.
func TestMisuse_GenerateBodyPhysical_SizeS(t *testing.T) {
	t.Parallel()
	r := roller.NewSeeded(42)
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("GenerateBodyPhysical panicked on SizeCode S: %v", rec)
		}
	}()
	_, _ = worlds.GenerateBodyPhysical(r, "S", 600, worlds.BodyPhysicalDMs{SizeCode: "S"})
}

// TestMisuse_GenerateBeltDetails_NegativeAge — negative ageGyr should
// not panic.
func TestMisuse_GenerateBeltDetails_NegativeAge(t *testing.T) {
	t.Parallel()
	r := roller.NewSeeded(42)
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("GenerateBeltDetails panicked on negative ageGyr: %v", rec)
		}
	}()
	_, _ = worlds.GenerateBeltDetails(r, 3.0, 0.5, 3.0, -1.0, false, false)
}

// TestMisuse_GenerateHydrographics_Atm0 — atm code 0 (vacuum) with
// otherwise-normal inputs. Hydrographics shouldn't panic; returns a
// zero-coverage record.
func TestMisuse_GenerateHydrographics_Atm0(t *testing.T) {
	t.Parallel()
	r := roller.NewSeeded(42)
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("GenerateHydrographics panicked on atm code 0: %v", rec)
		}
	}()
	_, _ = worlds.GenerateHydrographics(r, worlds.Atmosphere{Code: 0}, "5", worlds.TempTemperate)
}

// TestMisuse_RollBiomass_NoAtmosphere — body without Atmosphere
// shouldn't panic; biomass roll returns 0 (per biology.go's
// nil-atmosphere guard).
func TestMisuse_RollBiomass_NoAtmosphere(t *testing.T) {
	t.Parallel()
	r := roller.NewSeeded(42)
	body := &worlds.Body{Kind: worlds.BodyTerrestrial, SizeCode: "5"}
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("RollBiomass panicked on body without atmosphere: %v", rec)
		}
	}()
	got := worlds.RollBiomass(r, body, 4.6)
	if got > 0 {
		t.Errorf("RollBiomass returned %d for atm-less body; expected 0", got)
	}
}

// TestMisuse_RollCompatibility_BiocomplexityZero — biocomplexity 0
// shouldn't trigger compatibility roll (caller's responsibility), but
// if called anyway it shouldn't panic.
func TestMisuse_RollCompatibility_BiocomplexityZero(t *testing.T) {
	t.Parallel()
	r := roller.NewSeeded(42)
	body := &worlds.Body{
		Kind:       worlds.BodyTerrestrial,
		SizeCode:   "5",
		Atmosphere: &worlds.Atmosphere{Code: 6, Pressure: 1.0},
	}
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("RollCompatibility panicked on biocomplexity 0: %v", rec)
		}
	}()
	_ = worlds.RollCompatibility(r, body, 0, 4.6)
}

// TestMisuse_Renderers_ZeroValue — Markdown/JSON/PlainText renderers
// must handle zero-value form structs without panic. Per harness.md
// § Misuse-path tests for MarkdownClass0I etc.
func TestMisuse_Renderers_ZeroValue(t *testing.T) {
	t.Parallel()
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("zero-value renderer panicked: %v", rec)
		}
	}()
	_ = iiss.MarkdownClass0I(iiss.Class0IForm{})
	_ = iiss.MarkdownClass23(iiss.Class23Form{})
	_ = iiss.MarkdownClass4P(iiss.Class4PForm{})
	_, _ = iiss.JSONClass0I(iiss.Class0IForm{})
	_, _ = iiss.JSONClass23(iiss.Class23Form{})
	_, _ = iiss.JSONClass4P(iiss.Class4PForm{})
	_ = iiss.PlainTextClass0I(iiss.Class0IForm{})
	_ = iiss.PlainTextClass23(iiss.Class23Form{})
	_ = iiss.PlainTextClass4P(iiss.Class4PForm{})
	_ = iiss.MarkdownSystem(iiss.SystemForms{})
}
