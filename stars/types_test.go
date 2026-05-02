package stars

import (
	"testing"
)

func TestSpectralType_String(t *testing.T) {
	st := SpectralType{Letter: 'G', Subtype: 7}
	if got := st.String(); got != "G7" {
		t.Fatalf("got %q want %q", got, "G7")
	}
}

func TestSpectralType_Parse(t *testing.T) {
	cases := []struct {
		in   string
		want SpectralType
	}{
		{"G7", SpectralType{Letter: 'G', Subtype: 7}},
		{"M0", SpectralType{Letter: 'M', Subtype: 0}},
		{"M9", SpectralType{Letter: 'M', Subtype: 9}},
		{"O0", SpectralType{Letter: 'O', Subtype: 0}},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseSpectralType(tc.in)
			if err != nil {
				t.Fatalf("ParseSpectralType(%q) error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("got %+v want %+v", got, tc.want)
			}
		})
	}
}

func TestSpectralType_ParseInvalid(t *testing.T) {
	bad := []string{"", "G", "G77", "X3", "g7"}
	for _, in := range bad {
		t.Run(in, func(t *testing.T) {
			if _, err := ParseSpectralType(in); err == nil {
				t.Fatalf("ParseSpectralType(%q) succeeded; want error", in)
			}
		})
	}
}

func TestLuminosityClass_Values(t *testing.T) {
	if string(V) != "V" {
		t.Fatal("V != \"V\"")
	}
	if string(D) != "D" {
		t.Fatal("D != \"D\"")
	}
	if string(BD) != "BD" {
		t.Fatal("BD != \"BD\"")
	}
}

func TestStarKind_Values(t *testing.T) {
	for _, k := range []StarKind{KindMainSequence, KindWhiteDwarf, KindBlackHole, KindBrownDwarf} {
		if k == "" {
			t.Fatalf("empty StarKind: %v", k)
		}
	}
}
