package worlds

import (
	"testing"

	"wbh/roller"
)

func TestRollEmptyOrbits(t *testing.T) {
	t.Parallel()
	cases := []struct {
		roll int
		want int
	}{
		{2, 0},
		{7, 0},
		{9, 0},
		{10, 1},
		{11, 2},
		{12, 3},
	}
	for _, c := range cases {
		got, err := RollEmptyOrbits(roller.NewScripted(c.roll))
		if err != nil {
			t.Fatalf("%v", err)
		}
		if got != c.want {
			t.Errorf("roll %d: empty = %d, want %d", c.roll, got, c.want)
		}
	}
}
