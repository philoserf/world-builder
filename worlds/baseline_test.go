package worlds

import (
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestRollBaselineNumber_NoMods(t *testing.T) {
	t.Parallel()
	sys := stars.System{Primary: stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, LuminosityClass: stars.V,
	})}
	counts := Counts{Total: 17}
	got, err := RollBaselineNumber(roller.NewScripted(7), sys, counts)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got != 7 {
		t.Errorf("baseline = %d, want 7", got)
	}
}

func TestRollBaselineNumber_DMTable(t *testing.T) {
	t.Parallel()
	type tc struct {
		name string
		sys  stars.System
		tot  int
		roll int
		want int
	}
	cases := []tc{
		{
			name: "primary has companion",
			sys: stars.System{
				Primary: stars.Compose(stars.ComposeOpts{Kind: stars.KindMainSequence, LuminosityClass: stars.V}),
				Companions: []stars.CompanionStar{
					{Star: stars.Star{}, OrbitClass: stars.OrbitCompanion, ParentIndex: -1},
				},
			},
			tot: 17, roll: 9, want: 9 - 2,
		},
		{
			name: "Class III primary",
			sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
				Kind: stars.KindMainSequence, LuminosityClass: stars.III,
			})},
			tot: 17, roll: 7, want: 7 + 2,
		},
		{
			name: "Class IV primary",
			sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
				Kind: stars.KindMainSequence, LuminosityClass: stars.IV,
			})},
			tot: 17, roll: 7, want: 7 + 1,
		},
		{
			name: "Class VI primary",
			sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
				Kind: stars.KindMainSequence, LuminosityClass: stars.VI,
			})},
			tot: 17, roll: 7, want: 7 - 1,
		},
		{
			name: "post-stellar primary",
			sys: stars.System{Primary: stars.Compose(stars.ComposeOpts{
				Kind: stars.KindWhiteDwarf, LuminosityClass: stars.D,
			})},
			tot: 17, roll: 9, want: 9 - 2,
		},
		{
			name: "total worlds < 6",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  3, roll: 9, want: 9 - 4,
		},
		{
			name: "total worlds 6-9",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  7, roll: 9, want: 9 - 3,
		},
		{
			name: "total worlds 10-12",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  11, roll: 9, want: 9 - 2,
		},
		{
			name: "total worlds 13-15",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  14, roll: 9, want: 9 - 1,
		},
		{
			name: "total worlds 18-20",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  19, roll: 9, want: 9 + 1,
		},
		{
			name: "total worlds > 20",
			sys:  stars.System{Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V})},
			tot:  25, roll: 9, want: 9 + 2,
		},
		{
			name: "secondary star DM",
			sys: stars.System{
				Primary: stars.Compose(stars.ComposeOpts{LuminosityClass: stars.V}),
				Companions: []stars.CompanionStar{
					{Star: stars.Star{}, OrbitClass: stars.OrbitNear, ParentIndex: -1},
					{Star: stars.Star{}, OrbitClass: stars.OrbitFar, ParentIndex: -1},
				},
			},
			tot: 17, roll: 9, want: 9 - 2,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := RollBaselineNumber(roller.NewScripted(c.roll), c.sys, Counts{Total: c.tot})
			if err != nil {
				t.Fatalf("%v", err)
			}
			if got != c.want {
				t.Errorf("baseline = %d, want %d", got, c.want)
			}
		})
	}
}
