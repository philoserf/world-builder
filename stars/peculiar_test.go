package stars

import (
	"testing"

	"wbh/roller"
)

func TestKindFromUnusualCell(t *testing.T) {
	cases := map[string]StarKind{
		"BD": KindBrownDwarf,
		"D":  KindWhiteDwarf,
	}
	for cell, want := range cases {
		got, err := KindFromUnusualCell(cell)
		if err != nil {
			t.Fatalf("%s error: %v", cell, err)
		}
		if got != want {
			t.Fatalf("%s = %v want %v", cell, got, want)
		}
	}
}

func TestKindFromPeculiarCell(t *testing.T) {
	cases := map[string]StarKind{
		"Black Hole":   KindBlackHole,
		"Pulsar":       KindPulsar,
		"Neutron Star": KindNeutronStar,
		"Nebula":       KindNebula,
		"Protostar":    KindProtostar,
		"Star Cluster": KindStarCluster,
		"Anomaly":      KindAnomaly,
	}
	for cell, want := range cases {
		got, err := KindFromPeculiarCell(cell)
		if err != nil {
			t.Fatalf("%s error: %v", cell, err)
		}
		if got != want {
			t.Fatalf("%s = %v want %v", cell, got, want)
		}
	}
}

func TestRollSpecialPrimary_Simple(t *testing.T) {
	// 1D=3 -> Neutron Star, 1D=6 -> Black Hole.
	r := roller.NewScripted(3)
	got, err := RollSpecialPrimarySimple(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != KindNeutronStar {
		t.Fatalf("got %v want neutron star", got)
	}

	r2 := roller.NewScripted(6)
	got2, err := RollSpecialPrimarySimple(r2)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got2 != KindBlackHole {
		t.Fatalf("got %v want black hole", got2)
	}
}
