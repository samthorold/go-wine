package domain

import (
	"errors"
	"testing"
)

func TestNewCompanion_RequiresDrinkerAndName(t *testing.T) {
	d := NewID()

	if _, err := NewCompanion(d, ""); !errors.Is(err, ErrValidation) {
		t.Errorf("empty name: err = %v, want ErrValidation", err)
	}
	if _, err := NewCompanion("", "Alex"); !errors.Is(err, ErrValidation) {
		t.Errorf("empty drinker: err = %v, want ErrValidation", err)
	}
}

func TestNewCompanion_ScopedToDrinker(t *testing.T) {
	d := NewID()
	c, err := NewCompanion(d, "Alex")
	if err != nil {
		t.Fatalf("new companion: %v", err)
	}
	if c.ID == "" {
		t.Error("expected a fresh ID")
	}
	if c.DrinkerID != d {
		t.Errorf("DrinkerID = %q, want %q", c.DrinkerID, d)
	}
	if c.Name != "Alex" {
		t.Errorf("Name = %q, want Alex", c.Name)
	}
}
