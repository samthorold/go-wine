package domain

import (
	"errors"
	"testing"
)

func TestDrinker_RenameRejectsEmptyName(t *testing.T) {
	d, err := NewDrinker("Sam")
	if err != nil {
		t.Fatalf("new drinker: %v", err)
	}
	if err := d.Rename(""); !errors.Is(err, ErrValidation) {
		t.Errorf("empty rename: err = %v, want ErrValidation", err)
	}
	if d.Name != "Sam" {
		t.Errorf("a rejected rename must not mutate the name; got %q", d.Name)
	}
}

func TestDrinker_RenameKeepsIDAndUpdatesName(t *testing.T) {
	d, err := NewDrinker("Sam")
	if err != nil {
		t.Fatalf("new drinker: %v", err)
	}
	id := d.ID
	if err := d.Rename("Samuel"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if d.Name != "Samuel" {
		t.Errorf("Name = %q, want Samuel", d.Name)
	}
	if d.ID != id {
		t.Errorf("rename must keep the identity: ID = %q, want %q", d.ID, id)
	}
}
