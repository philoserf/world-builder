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

func TestNonPrimaryDescriptor_Companion_Sibling(t *testing.T) {
	// Zed Ab walkthrough p.30: G7 V parent, companion roll = 8 -> "Sibling".
	parent := Star{Kind: KindMainSequence, LuminosityClass: V, SpectralType: SpectralType{Letter: 'G', Subtype: 7}}
	r := roller.NewScripted(8)
	if got := RollNonPrimaryDescriptor(r, parent, RoleCompanion); got != "Sibling" {
		t.Fatalf("got %q, want Sibling", got)
	}
}

func TestNonPrimaryDescriptor_Secondary_Lesser(t *testing.T) {
	// Zed B p.30: G7 V parent, secondary roll = 8 -> "Lesser".
	parent := Star{Kind: KindMainSequence, LuminosityClass: V, SpectralType: SpectralType{Letter: 'G', Subtype: 7}}
	r := roller.NewScripted(8)
	if got := RollNonPrimaryDescriptor(r, parent, RoleSecondary); got != "Lesser" {
		t.Fatalf("got %q, want Lesser", got)
	}
}

func TestNonPrimaryDescriptor_ClassIII_DM(t *testing.T) {
	// Class III parent: DM-1 applied. Roll 9 -> 8 row -> Secondary "Lesser".
	parent := Star{Kind: KindMainSequence, LuminosityClass: III, SpectralType: SpectralType{Letter: 'G', Subtype: 0}}
	r := roller.NewScripted(9)
	if got := RollNonPrimaryDescriptor(r, parent, RoleSecondary); got != "Lesser" {
		t.Fatalf("got %q, want Lesser (with DM-1)", got)
	}
}

func TestGenerateCompanionStar_Twin(t *testing.T) {
	parent := Star{
		Kind:            KindMainSequence,
		SpectralType:    SpectralType{Letter: 'G', Subtype: 7},
		LuminosityClass: V,
		Mass:            0.929, Diameter: 0.967, Temperature: 5440, Luminosity: 0.738,
		AgeGyr: 6.3,
	}
	got, err := GenerateCompanionStar(roller.NewScripted(), parent, "Twin")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got.SpectralType != parent.SpectralType {
		t.Fatalf("type: got %v want %v", got.SpectralType, parent.SpectralType)
	}
	if got.Mass != parent.Mass {
		t.Fatalf("mass: got %v want %v", got.Mass, parent.Mass)
	}
}

func TestGenerateCompanionStar_Sibling_SameLetter(t *testing.T) {
	// Zed Ab: parent G7 V, 1D=1 -> G8 V (subtype 7+1=8).
	parent := Star{Kind: KindMainSequence, LuminosityClass: V, SpectralType: SpectralType{Letter: 'G', Subtype: 7}}
	r := roller.NewScripted(1)
	got, err := GenerateCompanionStar(r, parent, "Sibling")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := SpectralType{Letter: 'G', Subtype: 8}
	if got.SpectralType != want {
		t.Fatalf("got %v want %v", got.SpectralType, want)
	}
}

func TestGenerateCompanionStar_Sibling_LetterShift(t *testing.T) {
	// Parent G7 V, 1D=5 -> subtype 12 -> shift to K, subtract 10 -> K2 V.
	parent := Star{Kind: KindMainSequence, LuminosityClass: V, SpectralType: SpectralType{Letter: 'G', Subtype: 7}}
	r := roller.NewScripted(5)
	got, err := GenerateCompanionStar(r, parent, "Sibling")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := SpectralType{Letter: 'K', Subtype: 2}
	if got.SpectralType != want {
		t.Fatalf("got %v want %v", got.SpectralType, want)
	}
}

func TestGenerateCompanionStar_Sibling_M9Clamp(t *testing.T) {
	// Parent M5 V, 1D=6 -> subtype 11; cooler is missing (M is last); clamp to M9.
	parent := Star{Kind: KindMainSequence, LuminosityClass: V, SpectralType: SpectralType{Letter: 'M', Subtype: 5}}
	r := roller.NewScripted(6)
	got, err := GenerateCompanionStar(r, parent, "Sibling")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := SpectralType{Letter: 'M', Subtype: 9}
	if got.SpectralType != want {
		t.Fatalf("got %v want %v", got.SpectralType, want)
	}
}

func TestGenerateCompanionStar_Lesser(t *testing.T) {
	// Parent G7 V (Zed B walkthrough p.30): Lesser -> K letter, reroll subtype.
	// Roll 2D=8 on Numeric -> subtype 8 -> K8.
	parent := Star{Kind: KindMainSequence, LuminosityClass: V, SpectralType: SpectralType{Letter: 'G', Subtype: 7}}
	r := roller.NewScripted(8)
	got, err := GenerateCompanionStar(r, parent, "Lesser")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := SpectralType{Letter: 'K', Subtype: 8}
	if got.SpectralType != want {
		t.Fatalf("got %v want %v", got.SpectralType, want)
	}
}

func TestGenerateCompanionStar_BD(t *testing.T) {
	parent := Star{Kind: KindMainSequence, LuminosityClass: V, AgeGyr: 5.0}
	got, err := GenerateCompanionStar(roller.NewScripted(), parent, "BD")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got.Kind != KindBrownDwarf {
		t.Fatalf("kind: got %v want BrownDwarf", got.Kind)
	}
	if got.LuminosityClass != BD {
		t.Fatalf("class: got %v want BD", got.LuminosityClass)
	}
}

func TestGenerateCompanionStar_D(t *testing.T) {
	parent := Star{Kind: KindMainSequence, LuminosityClass: V, AgeGyr: 5.0}
	got, err := GenerateCompanionStar(roller.NewScripted(), parent, "D")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got.Kind != KindWhiteDwarf {
		t.Fatalf("kind: got %v want WhiteDwarf", got.Kind)
	}
	if got.LuminosityClass != D {
		t.Fatalf("class: got %v want D", got.LuminosityClass)
	}
}

func TestGenerateCompanionStar_OtherErrors(t *testing.T) {
	parent := Star{Kind: KindMainSequence, LuminosityClass: V, SpectralType: SpectralType{Letter: 'G', Subtype: 7}}
	_, err := GenerateCompanionStar(roller.NewScripted(), parent, "Other")
	if err == nil {
		t.Fatal("expected error for Other descriptor")
	}
}
