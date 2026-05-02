package stars

import (
	"testing"

	"wbh/roller"
)

func TestPresenceDM_GiantClasses(t *testing.T) {
	cases := []struct {
		name  string
		class LuminosityClass
		want  int
	}{
		{"Ia", Ia, 1},
		{"Ib", Ib, 1},
		{"II", II, 1},
		{"III", III, 1},
		{"IV", IV, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := Star{Kind: KindMainSequence, LuminosityClass: tc.class, SpectralType: SpectralType{Letter: 'G', Subtype: 0}}
			if got := PresenceDM(s); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPresenceDM_VVI_HotTypes(t *testing.T) {
	for _, lc := range []LuminosityClass{V, VI} {
		for _, letter := range []SpectralLetter{'O', 'B', 'A', 'F'} {
			s := Star{Kind: KindMainSequence, LuminosityClass: lc, SpectralType: SpectralType{Letter: letter, Subtype: 0}}
			if got := PresenceDM(s); got != 1 {
				t.Fatalf("%c %s: got %d, want 1", letter, lc, got)
			}
		}
	}
}

func TestPresenceDM_VVI_MType(t *testing.T) {
	for _, lc := range []LuminosityClass{V, VI} {
		s := Star{Kind: KindMainSequence, LuminosityClass: lc, SpectralType: SpectralType{Letter: 'M', Subtype: 0}}
		if got := PresenceDM(s); got != -1 {
			t.Fatalf("M %s: got %d, want -1", lc, got)
		}
	}
}

func TestPresenceDM_GVKPrimary_NoModifier(t *testing.T) {
	// G/K class V (or VI) primary -> no DM.
	for _, letter := range []SpectralLetter{'G', 'K'} {
		s := Star{Kind: KindMainSequence, LuminosityClass: V, SpectralType: SpectralType{Letter: letter, Subtype: 0}}
		if got := PresenceDM(s); got != 0 {
			t.Fatalf("%c V: got %d, want 0", letter, got)
		}
	}
}

func TestPresenceDM_BrownDwarf(t *testing.T) {
	s := Star{Kind: KindBrownDwarf, LuminosityClass: BD}
	if got := PresenceDM(s); got != -1 {
		t.Fatalf("BD: got %d, want -1", got)
	}
}

func TestPresenceDM_WhiteDwarf(t *testing.T) {
	s := Star{Kind: KindWhiteDwarf, LuminosityClass: D}
	if got := PresenceDM(s); got != -1 {
		t.Fatalf("D: got %d, want -1", got)
	}
}

func TestPresenceDM_PostStellar(t *testing.T) {
	for _, kind := range []StarKind{KindPulsar, KindNeutronStar, KindBlackHole} {
		s := Star{Kind: kind}
		if got := PresenceDM(s); got != -1 {
			t.Fatalf("%v: got %d, want -1", kind, got)
		}
	}
}

func TestRollPresence_GiantBlocksClose(t *testing.T) {
	// Class Ia/Ib/II/III primaries cannot have Close secondaries; should
	// return false without consuming a roll.
	for _, lc := range []LuminosityClass{Ia, Ib, II, III} {
		t.Run(string(lc), func(t *testing.T) {
			s := Star{Kind: KindMainSequence, LuminosityClass: lc, SpectralType: SpectralType{Letter: 'G', Subtype: 0}}
			r := roller.NewScripted() // empty - if RollPresence rolls, it will panic
			if got := RollPresence(r, s, OrbitClose); got {
				t.Fatalf("got true, want false")
			}
		})
	}
}

func TestRollPresence_Threshold(t *testing.T) {
	s := Star{Kind: KindMainSequence, LuminosityClass: V, SpectralType: SpectralType{Letter: 'G', Subtype: 0}}
	// G V no DM -> threshold 10.
	r9 := roller.NewScripted(9)
	if RollPresence(r9, s, OrbitNear) {
		t.Fatal("9 should be below threshold")
	}
	r10 := roller.NewScripted(10)
	if !RollPresence(r10, s, OrbitNear) {
		t.Fatal("10 should be at threshold")
	}
}

func TestRollPresence_DMShifts(t *testing.T) {
	// O V primary -> DM+1; a roll of 9 + 1 = 10 succeeds.
	s := Star{Kind: KindMainSequence, LuminosityClass: V, SpectralType: SpectralType{Letter: 'O', Subtype: 0}}
	r := roller.NewScripted(9)
	if !RollPresence(r, s, OrbitNear) {
		t.Fatal("9 + DM+1 = 10 should succeed")
	}
}

func TestRollPresence_Zed(t *testing.T) {
	// WBH pp.23-24: G7 V primary (DM 0). Close=4 (false), Near=10 (true), Far=11 (true), Companion=11 (true).
	s := Star{Kind: KindMainSequence, LuminosityClass: V, SpectralType: SpectralType{Letter: 'G', Subtype: 7}}
	r := roller.NewScripted(4, 10, 11, 11)
	if RollPresence(r, s, OrbitClose) {
		t.Fatal("Close: 4 should be false")
	}
	if !RollPresence(r, s, OrbitNear) {
		t.Fatal("Near: 10 should be true")
	}
	if !RollPresence(r, s, OrbitFar) {
		t.Fatal("Far: 11 should be true")
	}
	if !RollPresence(r, s, OrbitCompanion) {
		t.Fatal("Companion: 11 should be true")
	}
}
