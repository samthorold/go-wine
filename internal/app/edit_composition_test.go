package app_test

import (
	"context"
	"errors"
	"testing"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

func compositionFixture(t *testing.T) (*memory.WineRepo, *memory.VarietyRepo, domain.Wine, domain.Variety, domain.Variety) {
	t.Helper()
	wines := memory.NewWineRepo()
	varieties := memory.NewVarietyRepo()

	w, _ := domain.NewWine("Guigal", "Côtes du Rhône Rouge", "GSM")
	wines.Save(w)
	g, _ := domain.NewVariety("Grenache")
	s, _ := domain.NewVariety("Syrah")
	varieties.Save(g)
	varieties.Save(s)
	return wines, varieties, w, g, s
}

func TestEditComposition_PersistsValidComposition(t *testing.T) {
	wines, varieties, w, g, s := compositionFixture(t)
	h := app.NewEditCompositionHandler(wines, varieties)

	err := h.Handle(context.Background(), app.EditCompositionCommand{
		WineID: w.ID,
		Parts: []app.CompositionPartInput{
			{VarietyID: g.ID, Proportion: 60},
			{VarietyID: s.ID, Proportion: 40},
		},
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	got, _ := wines.Get(context.Background(), w.ID)
	if len(got.Composition.Parts) != 2 {
		t.Fatalf("composition parts = %d, want 2", len(got.Composition.Parts))
	}
}

func TestEditComposition_RejectsSumOff(t *testing.T) {
	wines, varieties, w, g, s := compositionFixture(t)
	h := app.NewEditCompositionHandler(wines, varieties)

	err := h.Handle(context.Background(), app.EditCompositionCommand{
		WineID: w.ID,
		Parts: []app.CompositionPartInput{
			{VarietyID: g.ID, Proportion: 30},
			{VarietyID: s.ID, Proportion: 30},
		},
	})
	if !errors.Is(err, domain.ErrInvalidComposition) {
		t.Errorf("err = %v, want ErrInvalidComposition", err)
	}
}

func TestEditComposition_RejectsEmpty(t *testing.T) {
	wines, varieties, w, _, _ := compositionFixture(t)
	h := app.NewEditCompositionHandler(wines, varieties)

	err := h.Handle(context.Background(), app.EditCompositionCommand{WineID: w.ID})
	if !errors.Is(err, domain.ErrInvalidComposition) {
		t.Errorf("err = %v, want ErrInvalidComposition", err)
	}
}

func TestEditComposition_RejectsUnknownWine(t *testing.T) {
	wines, varieties, _, g, _ := compositionFixture(t)
	h := app.NewEditCompositionHandler(wines, varieties)

	err := h.Handle(context.Background(), app.EditCompositionCommand{
		WineID: domain.NewID(),
		Parts:  []app.CompositionPartInput{{VarietyID: g.ID, Proportion: 100}},
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestEditComposition_RejectsUnknownVariety(t *testing.T) {
	wines, varieties, w, _, _ := compositionFixture(t)
	h := app.NewEditCompositionHandler(wines, varieties)

	err := h.Handle(context.Background(), app.EditCompositionCommand{
		WineID: w.ID,
		Parts:  []app.CompositionPartInput{{VarietyID: domain.NewID(), Proportion: 100}},
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
