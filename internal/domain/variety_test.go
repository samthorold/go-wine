package domain

import (
	"errors"
	"testing"
)

func TestNewVariety_AssignsIDAndName(t *testing.T) {
	v, err := NewVariety("Shiraz")
	if err != nil {
		t.Fatalf("NewVariety: %v", err)
	}
	if v.Name != "Shiraz" {
		t.Errorf("Name = %q, want Shiraz", v.Name)
	}
	if v.ID == "" {
		t.Errorf("expected a fresh ID, got empty")
	}
}

func TestNewVariety_RejectsEmptyName(t *testing.T) {
	if _, err := NewVariety(""); !errors.Is(err, ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}
