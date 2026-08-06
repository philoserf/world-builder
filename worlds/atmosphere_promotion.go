package worlds

// PromoteOxygenTaint applies the WBH p.81 "tainted equivalent" rule:
// when an atm 5/6/8 has computed ppO2 outside [0.10, 0.50] bar, the
// code is promoted to its tainted equivalent (4/7/9) with low (ppO2 <
// 0.10) or high (ppO2 > 0.50) oxygen pre-seeded as the first taint
// subtype.
//
// For atms outside 5/6/8 or with ppO2 in band, returns (atmCode, nil).
//
// The pre-seeded Taint has Severity and Persistence == 0; the
// ApplyTaintTypology orchestrator fills them from the severity/persistence
// rolls so callers don't have to special-case pre-seeded taints.
func PromoteOxygenTaint(atmCode int, ppO2 float64) (int, *Taint) {
	if atmCode != 5 && atmCode != 6 && atmCode != 8 {
		return atmCode, nil
	}

	if ppO2 >= 0.10 && ppO2 <= 0.50 {
		return atmCode, nil
	}

	taintCode := "L"
	if ppO2 > 0.50 {
		taintCode = "H"
	}
	// Promotion map: 5→4, 6→7, 8→9.
	promotions := map[int]int{5: 4, 6: 7, 8: 9}
	newCode := promotions[atmCode]

	return newCode, &Taint{Code: taintCode}
}
