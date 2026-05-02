package stars

import (
	"testing"

	"wbh/roller"
)

func TestRollPrimaryTypeAndClass_Zed(t *testing.T) {
	// WBH p. 16: "the roll for primary star is a 9, resulting in a type
	// G-type main sequence, Class V star."
	r := roller.NewScripted(9)
	letter, lc, err := RollPrimaryTypeAndClass(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if letter != 'G' {
		t.Fatalf("letter = %c, want G", letter)
	}
	if lc != V {
		t.Fatalf("class = %s, want V", lc)
	}
}

func TestRollPrimaryTypeAndClass_HotRedirect(t *testing.T) {
	// 12 -> "Hot" -> reroll on Hot column.
	// 12 followed by 5 -> Hot column at 5 = "A".
	r := roller.NewScripted(12, 5)
	letter, lc, err := RollPrimaryTypeAndClass(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if letter != 'A' {
		t.Fatalf("letter = %c, want A", letter)
	}
	if lc != V {
		t.Fatalf("class = %s, want V", lc)
	}
}

func TestRollPrimaryTypeAndClass_SpecialNotImplemented(t *testing.T) {
	// 2 -> "Special" branch is dispatched separately; for v1 (Plan 1) we
	// surface ErrSpecialPrimary so callers can route through peculiar.go.
	r := roller.NewScripted(2)
	_, _, err := RollPrimaryTypeAndClass(r)
	if err == nil {
		t.Fatal("expected ErrSpecialPrimary")
	}
}

func TestRollSubtype_Zed(t *testing.T) {
	// WBH p. 16: a 2D roll of 6 on the Numeric column -> subtype 7 (G7).
	r := roller.NewScripted(6)
	got, err := RollSubtype(r, 'G', V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 7 {
		t.Fatalf("got %d, want 7", got)
	}
}

func TestRollSubtype_MType(t *testing.T) {
	// Primary M-type uses the M column. 2D=6 on M column -> 0 (M0).
	r := roller.NewScripted(6)
	got, err := RollSubtype(r, 'M', V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 0 {
		t.Fatalf("got %d, want 0", got)
	}
}

func TestRollSubtype_KClassIVAdjustment(t *testing.T) {
	// WBH p. 16: "For a K-type Class IV star, subtract 5 (make lower)
	// any subtype result above 4."
	// 2D=7 -> Numeric[7] = 9; for K Class IV, 9 - 5 = 4.
	r := roller.NewScripted(7)
	got, err := RollSubtype(r, 'K', IV)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 4 {
		t.Fatalf("K Class IV adjustment: got %d, want 4", got)
	}
}

func TestRollSubtype_AllRollsInRange(t *testing.T) {
	for v := 2; v <= 12; v++ {
		r := roller.NewScripted(v)
		got, err := RollSubtype(r, 'G', V)
		if err != nil {
			t.Fatalf("v=%d error: %v", v, err)
		}
		if got < 0 || got > 9 {
			t.Fatalf("v=%d subtype out of range: %d", v, got)
		}
	}
}
