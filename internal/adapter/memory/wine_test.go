package memory

import (
	"context"
	"errors"
	"testing"

	"go-wine/internal/domain"
)

func TestWineRepo_SetCompositionPersistsAndReadsBack(t *testing.T) {
	repo := NewWineRepo()
	w, _ := domain.NewWine("Guigal", "Côtes du Rhône Rouge", "GSM")
	repo.Save(w)

	g, s, m := domain.NewID(), domain.NewID(), domain.NewID()
	c, _ := domain.NewComposition([]domain.CompositionPart{
		{VarietyID: g, Proportion: 60},
		{VarietyID: s, Proportion: 25},
		{VarietyID: m, Proportion: 15},
	}, domain.ProvenanceConfirmed)
	if err := repo.SetComposition(context.Background(), w.ID, c); err != nil {
		t.Fatalf("SetComposition: %v", err)
	}

	got, err := repo.Get(context.Background(), w.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.Composition.Parts) != 3 {
		t.Fatalf("Composition parts = %d, want 3", len(got.Composition.Parts))
	}
}

func TestWineRepo_SetCompositionUnknownWineIsNotFound(t *testing.T) {
	repo := NewWineRepo()
	c, _ := domain.NewComposition([]domain.CompositionPart{{VarietyID: domain.NewID(), Proportion: 100}}, domain.ProvenanceConfirmed)
	if err := repo.SetComposition(context.Background(), domain.NewID(), c); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

// Compile-time assertion that WineRepo satisfies the port.
var _ domain.WineRepository = (*WineRepo)(nil)
