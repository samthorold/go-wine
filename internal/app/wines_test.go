package app_test

import (
	"context"
	"errors"
	"testing"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

func TestListWines_SortsByLabelAndFlagsComposition(t *testing.T) {
	wines := memory.NewWineRepo()
	a, _ := domain.NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	b, _ := domain.NewWine("Cloudy Bay", "Sauvignon Blanc", "Sauvignon Blanc")
	wines.Save(a)
	wines.Save(b)
	c, _ := domain.NewComposition([]domain.CompositionPart{{VarietyID: domain.NewID(), Proportion: 100}}, domain.ProvenanceConfirmed)
	_ = wines.SetComposition(context.Background(), a.ID, c)

	h := app.NewListWinesHandler(wines)
	views, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("len = %d, want 2", len(views))
	}
	if views[0].Label != "Cloudy Bay — Sauvignon Blanc" {
		t.Errorf("first label = %q, want Cloudy Bay sorted first", views[0].Label)
	}
	// Penfolds (second) has a Composition; Cloudy Bay does not.
	if views[0].HasComposition {
		t.Errorf("Cloudy Bay should have no composition")
	}
	if !views[1].HasComposition {
		t.Errorf("Penfolds should be flagged as having a composition")
	}
}

func TestGetWine_ResolvesVarietyNames(t *testing.T) {
	wines := memory.NewWineRepo()
	varieties := memory.NewVarietyRepo()
	w, _ := domain.NewWine("Guigal", "Côtes du Rhône Rouge", "GSM")
	wines.Save(w)
	g, _ := domain.NewVariety("Grenache")
	s, _ := domain.NewVariety("Syrah")
	varieties.Save(g)
	varieties.Save(s)
	c, _ := domain.NewComposition([]domain.CompositionPart{
		{VarietyID: g.ID, Proportion: 40},
		{VarietyID: s.ID, Proportion: 60},
	}, domain.ProvenanceConfirmed)
	_ = wines.SetComposition(context.Background(), w.ID, c)

	h := app.NewGetWineHandler(wines, varieties)
	view, err := h.Handle(context.Background(), w.ID)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if view.Label != "Guigal — Côtes du Rhône Rouge" {
		t.Errorf("Label = %q", view.Label)
	}
	if len(view.Composition) != 2 {
		t.Fatalf("composition parts = %d, want 2", len(view.Composition))
	}
	// Ordered by descending proportion: Syrah (60) leads.
	if view.Composition[0].VarietyName != "Syrah" || view.Composition[0].Proportion != 60 {
		t.Errorf("first part = %+v, want Syrah 60", view.Composition[0])
	}
}

func TestGetWine_UnknownIsNotFound(t *testing.T) {
	wines := memory.NewWineRepo()
	varieties := memory.NewVarietyRepo()
	h := app.NewGetWineHandler(wines, varieties)
	if _, err := h.Handle(context.Background(), domain.NewID()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
