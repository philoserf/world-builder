package stars

import "testing"

func TestSystem_TypeFieldsZeroValue(t *testing.T) {
	var s System
	if s.PrimaryDesignation != "" {
		t.Fatalf("zero value PrimaryDesignation: %q", s.PrimaryDesignation)
	}
	if len(s.Companions) != 0 {
		t.Fatalf("zero value Companions: %v", s.Companions)
	}
}

func TestCompanionStar_TypeFieldsZeroValue(t *testing.T) {
	var c CompanionStar
	if c.Designation != "" {
		t.Fatalf("zero value Designation: %q", c.Designation)
	}
	if c.ParentIndex != 0 {
		t.Fatalf("zero value ParentIndex: %d", c.ParentIndex)
	}
}
