package domain

import (
	"errors"
	"testing"
)

func TestNewAxis_AcceptsTheScalePoints(t *testing.T) {
	for v := 1; v <= 5; v++ {
		a, err := NewAxis(v)
		if err != nil {
			t.Fatalf("NewAxis(%d): %v", v, err)
		}
		if a.Int() != v {
			t.Errorf("Int() = %d, want %d", a.Int(), v)
		}
	}
}

func TestNewAxis_RejectsOutOfScale(t *testing.T) {
	for _, v := range []int{0, -1, 6, 100} {
		if _, err := NewAxis(v); !errors.Is(err, ErrInvalidCharacteristics) {
			t.Errorf("NewAxis(%d) err = %v, want ErrInvalidCharacteristics", v, err)
		}
	}
}

func TestNewCharacteristics_BuildsBundleWithProvenance(t *testing.T) {
	c, err := NewCharacteristics(5, 4, 2, 1, 5, []string{"blackberry", "pepper"}, ProvenanceDefault)
	if err != nil {
		t.Fatalf("NewCharacteristics: %v", err)
	}
	if c.Body.Int() != 5 || c.Tannin.Int() != 4 || c.Acidity.Int() != 2 || c.Sweetness.Int() != 1 || c.Alcohol.Int() != 5 {
		t.Errorf("axes not set as given: %+v", c)
	}
	if len(c.Notes) != 2 || c.Notes[0] != "blackberry" {
		t.Errorf("Notes = %v, want [blackberry pepper]", c.Notes)
	}
	if c.Provenance != ProvenanceDefault {
		t.Errorf("Provenance = %v, want default", c.Provenance)
	}
}

func TestNewCharacteristics_RejectsOutOfScaleAxis(t *testing.T) {
	if _, err := NewCharacteristics(6, 1, 1, 1, 1, nil, ProvenanceDefault); !errors.Is(err, ErrInvalidCharacteristics) {
		t.Errorf("err = %v, want ErrInvalidCharacteristics", err)
	}
}

func TestCharacteristics_IsZeroWhenUnset(t *testing.T) {
	var c Characteristics
	if !c.IsZero() {
		t.Errorf("zero-value Characteristics should report IsZero")
	}
	set, _ := NewCharacteristics(3, 3, 3, 3, 3, nil, ProvenanceDefault)
	if set.IsZero() {
		t.Errorf("a populated Characteristics should not report IsZero")
	}
}

func TestCharacteristics_Confirmed(t *testing.T) {
	c, _ := NewCharacteristics(3, 3, 3, 3, 3, nil, ProvenanceDefault)
	if c.IsConfirmed() {
		t.Errorf("a default-provenance bundle is not confirmed")
	}
	c.Provenance = ProvenanceConfirmed
	if !c.IsConfirmed() {
		t.Errorf("a confirmed-provenance bundle should report IsConfirmed")
	}
}
