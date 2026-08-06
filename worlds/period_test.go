package worlds

import (
	"math"
	"testing"
)

// TestPeriodFor covers the three WBH p.53 cases:
//
//	Single star:        P = sqrt(AU^3 / M☉)              → PeriodFor(au, M, 0)
//	Multiple stars:     P = sqrt(AU^3 / Σ M☉)           → PeriodFor(au, sumM, 0)
//	Large planet:       P = sqrt(AU^3 / (Σ M☉ + m⊕×ε)) → PeriodFor(au, sumM, mEarth)
//
// where ε = 0.000003 (Terra mass in solar units).
func TestPeriodFor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		au        float64
		sumMass   float64 // sum of stellar masses in M☉
		mEarth    float64 // body mass in m⊕ (0 to skip Large Planet variant)
		wantYears float64
	}{
		{
			// Sol single-star: Earth at 1.0 AU, M=1.0 → 1.000y
			name: "Sol Earth", au: 1.0, sumMass: 1.0, mEarth: 0,
			wantYears: 1.0,
		},
		{
			// Zed B I: orbit 0.52 → AU 0.208 around B alone (M=0.626)
			// sqrt(0.208^3 / 0.626) = sqrt(0.008998 / 0.626) = sqrt(0.01437) = 0.11988y
			// Form p.63 shows 0.120y.
			name: "Zed B I", au: 0.208, sumMass: 0.626, mEarth: 0,
			wantYears: 0.120,
		},
		{
			// Zed AB I: orbit 7.2 → AU 5.68 with sumStellarMass = M(Aa)+M(Ab)+M(B) = 0.929+0.907+0.626 = 2.462
			// sqrt(5.68^3 / 2.462) = sqrt(183.25/2.462) = sqrt(74.43) = 8.628y
			// Form p.63 shows 8.627y for B (Aab+B barycentre at orbit 7.2).
			name: "Zed AB I", au: 5.68, sumMass: 2.462, mEarth: 0,
			wantYears: 8.627,
		},
		{
			// Large-Planet variant smoke: a 4000⊕ planet adds 4000×0.000003 = 0.012 M☉ to sumMass.
			// At AU=5, sumMass=1.0 → without: sqrt(125/1.0)=11.180y; with mEarth=4000:
			// sqrt(125/(1.0+0.012))=sqrt(125/1.012)=sqrt(123.52)=11.114y.
			name: "Large planet variant 4000 mEarth", au: 5.0, sumMass: 1.0, mEarth: 4000,
			wantYears: 11.114,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := PeriodFor(tc.au, tc.sumMass, tc.mEarth)
			if math.Abs(got.Years-tc.wantYears) > 0.005 {
				t.Errorf("Years = %v, want %v (±0.005)", got.Years, tc.wantYears)
			}

			if math.Abs(got.Days-got.Years*365.25) > 1e-9 {
				t.Errorf("Days = %v, want Years*365.25 = %v", got.Days, got.Years*365.25)
			}
		})
	}
}
