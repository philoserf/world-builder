package roller

import (
	"testing"
)

func TestSeeded_Deterministic(t *testing.T) {
	a := NewSeeded(42)
	b := NewSeeded(42)
	for i := range 20 {
		ra, rb := a.Roll("2D"), b.Roll("2D")
		if ra != rb {
			t.Fatalf("seeded rollers diverged at i=%d: %d vs %d", i, ra, rb)
		}
	}
}

func TestSeeded_2DInRange(t *testing.T) {
	r := NewSeeded(1)
	for range 200 {
		v := r.Roll("2D")
		if v < 2 || v > 12 {
			t.Fatalf("2D out of range: %d", v)
		}
	}
}

func TestSeeded_Modifier(t *testing.T) {
	r := NewSeeded(1)
	for range 200 {
		v := r.Roll("2D-7")
		if v < -5 || v > 5 {
			t.Fatalf("2D-7 out of range: %d", v)
		}
	}
}

func TestSeeded_D10(t *testing.T) {
	r := NewSeeded(1)
	for range 200 {
		v := r.Roll("d10")
		if v < 1 || v > 10 {
			t.Fatalf("d10 out of range: %d", v)
		}
	}
}

func TestScripted_Order(t *testing.T) {
	r := NewScripted(7, 9, 11)
	if got := r.Roll("2D"); got != 7 {
		t.Fatalf("first roll = %d, want 7", got)
	}
	if got := r.Roll("2D"); got != 9 {
		t.Fatalf("second roll = %d, want 9", got)
	}
	if got := r.Roll("2D"); got != 11 {
		t.Fatalf("third roll = %d, want 11", got)
	}
}

func TestScripted_PanicsOnExhaustion(t *testing.T) {
	r := NewScripted(5)
	r.Roll("2D")
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on exhausted Scripted")
		}
	}()
	r.Roll("2D")
}

func TestFixed_AlwaysSame(t *testing.T) {
	r := Fixed(8)
	if r.Roll("2D") != 8 {
		t.Fatal("Fixed(8).Roll(\"2D\") != 8")
	}
	if r.Roll("1D") != 8 {
		t.Fatal("Fixed(8).Roll(\"1D\") != 8")
	}
	if r.Roll("d100") != 8 {
		t.Fatal("Fixed(8).Roll(\"d100\") != 8")
	}
}

func TestRollerInterface(_ *testing.T) {
	var _ Roller = NewSeeded(1)
	var _ Roller = NewScripted(1)
	var _ Roller = Fixed(1)
}
