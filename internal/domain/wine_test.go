package domain

import (
	"errors"
	"testing"
)

func TestWine_SetComposition_AcceptsValid(t *testing.T) {
	w, _ := NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	err := w.SetComposition([]CompositionPart{{VarietyID: NewID(), Proportion: 100}}, ProvenanceConfirmed)
	if err != nil {
		t.Fatalf("SetComposition: %v", err)
	}
	if w.Composition.IsZero() {
		t.Errorf("composition should be set")
	}
}

func TestWine_SetComposition_RejectsInvalidAndLeavesWineUnchanged(t *testing.T) {
	w, _ := NewWine("Guigal", "Côtes du Rhône", "GSM")
	if err := w.SetComposition(nil, ProvenanceConfirmed); !errors.Is(err, ErrInvalidComposition) {
		t.Fatalf("err = %v, want ErrInvalidComposition", err)
	}
	if !w.Composition.IsZero() {
		t.Errorf("composition should be left unset after a rejected set")
	}
}
