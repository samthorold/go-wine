package domain

import (
	"errors"
	"testing"
)

func TestNewComposition_AcceptsPartsSummingTo100(t *testing.T) {
	a, b := NewID(), NewID()
	c, err := NewComposition([]CompositionPart{
		{VarietyID: a, Proportion: 60},
		{VarietyID: b, Proportion: 40},
	}, ProvenanceDefault)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if len(c.Parts) != 2 {
		t.Errorf("len(Parts) = %d, want 2", len(c.Parts))
	}
}

func TestNewComposition_AcceptsSingleVarietyAt100(t *testing.T) {
	if _, err := NewComposition([]CompositionPart{{VarietyID: NewID(), Proportion: 100}}, ProvenanceDefault); err != nil {
		t.Fatalf("single-variety 100%% should be valid: %v", err)
	}
}

func TestNewComposition_AcceptsEqualThirdsWithinTolerance(t *testing.T) {
	// 33+33+33 = 99, inside tolerance — a GSM should not be rejected.
	parts := []CompositionPart{
		{VarietyID: NewID(), Proportion: 33},
		{VarietyID: NewID(), Proportion: 33},
		{VarietyID: NewID(), Proportion: 33},
	}
	if _, err := NewComposition(parts, ProvenanceDefault); err != nil {
		t.Fatalf("equal-thirds blend should be valid: %v", err)
	}
}

func TestNewComposition_RejectsEmpty(t *testing.T) {
	if _, err := NewComposition(nil, ProvenanceDefault); !errors.Is(err, ErrInvalidComposition) {
		t.Errorf("err = %v, want ErrInvalidComposition", err)
	}
}

func TestNewComposition_RejectsSumWellBelow100(t *testing.T) {
	parts := []CompositionPart{
		{VarietyID: NewID(), Proportion: 50},
		{VarietyID: NewID(), Proportion: 30},
	}
	if _, err := NewComposition(parts, ProvenanceDefault); !errors.Is(err, ErrInvalidComposition) {
		t.Errorf("err = %v, want ErrInvalidComposition", err)
	}
}

func TestNewComposition_RejectsSumWellAbove100(t *testing.T) {
	parts := []CompositionPart{
		{VarietyID: NewID(), Proportion: 70},
		{VarietyID: NewID(), Proportion: 70},
	}
	if _, err := NewComposition(parts, ProvenanceDefault); !errors.Is(err, ErrInvalidComposition) {
		t.Errorf("err = %v, want ErrInvalidComposition", err)
	}
}

func TestNewComposition_RejectsNonPositiveProportion(t *testing.T) {
	parts := []CompositionPart{
		{VarietyID: NewID(), Proportion: 100},
		{VarietyID: NewID(), Proportion: 0},
	}
	if _, err := NewComposition(parts, ProvenanceDefault); !errors.Is(err, ErrInvalidComposition) {
		t.Errorf("err = %v, want ErrInvalidComposition", err)
	}
}

func TestNewComposition_RejectsMissingVarietyID(t *testing.T) {
	parts := []CompositionPart{{VarietyID: "", Proportion: 100}}
	if _, err := NewComposition(parts, ProvenanceDefault); !errors.Is(err, ErrInvalidComposition) {
		t.Errorf("err = %v, want ErrInvalidComposition", err)
	}
}
