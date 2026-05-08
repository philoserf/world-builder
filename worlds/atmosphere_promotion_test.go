package worlds

import "testing"

func TestPromoteOxygenTaint_NoChangeForUntaintedCodesInBand(t *testing.T) {
	for _, code := range []int{5, 6, 8} {
		newCode, seeded := PromoteOxygenTaint(code, 0.21)
		if newCode != code {
			t.Errorf("atm %d ppO2=0.21: got code %d, want %d", code, newCode, code)
		}
		if seeded != nil {
			t.Errorf("atm %d ppO2=0.21: got seeded taint, want nil", code)
		}
	}
}

func TestPromoteOxygenTaint_PromoteOnLowOxygen(t *testing.T) {
	cases := []struct {
		atmCode int
		newCode int
		ppO2    float64
	}{
		{5, 4, 0.05}, // thin → thin tainted
		{6, 7, 0.05}, // standard → standard tainted
		{8, 9, 0.05}, // dense → dense tainted
	}
	for _, c := range cases {
		newCode, seeded := PromoteOxygenTaint(c.atmCode, c.ppO2)
		if newCode != c.newCode {
			t.Errorf("atm %d ppO2=%g: got code %d, want %d", c.atmCode, c.ppO2, newCode, c.newCode)
		}
		if seeded == nil {
			t.Fatalf("atm %d ppO2=%g: got nil seeded, want L taint", c.atmCode, c.ppO2)
		}
		if seeded.Code != "L" {
			t.Errorf("atm %d ppO2=%g: got seeded code %q, want \"L\"", c.atmCode, c.ppO2, seeded.Code)
		}
	}
}

func TestPromoteOxygenTaint_PromoteOnHighOxygen(t *testing.T) {
	cases := []struct {
		atmCode int
		newCode int
		ppO2    float64
	}{
		{5, 4, 0.60}, // thin → thin tainted, high
		{6, 7, 0.60}, // standard → standard tainted, high
		{8, 9, 0.60}, // dense → dense tainted, high
	}
	for _, c := range cases {
		newCode, seeded := PromoteOxygenTaint(c.atmCode, c.ppO2)
		if newCode != c.newCode {
			t.Errorf("atm %d ppO2=%g: got code %d, want %d", c.atmCode, c.ppO2, newCode, c.newCode)
		}
		if seeded == nil {
			t.Fatalf("atm %d ppO2=%g: got nil seeded, want H taint", c.atmCode, c.ppO2)
		}
		if seeded.Code != "H" {
			t.Errorf("atm %d ppO2=%g: got seeded code %q, want \"H\"", c.atmCode, c.ppO2, seeded.Code)
		}
	}
}

func TestPromoteOxygenTaint_NoPromoteForOtherCodes(t *testing.T) {
	for _, code := range []int{0, 1, 2, 3, 4, 7, 9, 10, 11, 12, 13, 14, 15} {
		newCode, seeded := PromoteOxygenTaint(code, 0.05)
		if newCode != code {
			t.Errorf("atm %d ppO2=0.05: got code %d, want unchanged %d", code, newCode, code)
		}
		if seeded != nil {
			t.Errorf("atm %d: got seeded, want nil (not 5/6/8)", code)
		}
	}
}

func TestPromoteOxygenTaint_BoundaryValues(t *testing.T) {
	tests := []struct {
		ppO2          float64
		shouldPromote bool
		expectCode    string
	}{
		{0.10, false, ""},
		{0.50, false, ""},
		{0.0999, true, "L"},
		{0.5001, true, "H"},
	}
	for _, c := range tests {
		_, seeded := PromoteOxygenTaint(6, c.ppO2)
		if c.shouldPromote && seeded == nil {
			t.Errorf("ppO2=%g: expected promotion to %s, got none", c.ppO2, c.expectCode)
		}
		if !c.shouldPromote && seeded != nil {
			t.Errorf("ppO2=%g: expected no promotion, got seeded code %q", c.ppO2, seeded.Code)
		}
		if c.shouldPromote && seeded != nil && seeded.Code != c.expectCode {
			t.Errorf("ppO2=%g: got code %q, want %q", c.ppO2, seeded.Code, c.expectCode)
		}
	}
}
