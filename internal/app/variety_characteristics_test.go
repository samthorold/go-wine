package app_test

import (
	"context"
	"errors"
	"testing"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

func TestGetVariety_ReturnsNameAndCharacteristics(t *testing.T) {
	repo := memory.NewVarietyRepo()
	v, _ := domain.NewVariety("Shiraz")
	repo.Save(v)
	c, _ := domain.NewCharacteristics(5, 4, 2, 1, 5, []string{"blackberry", "pepper"}, domain.ProvenanceDefault)
	_ = repo.SetCharacteristics(context.Background(), v.ID, c)

	h := app.NewGetVarietyHandler(repo)
	view, err := h.Handle(context.Background(), v.ID)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if view.Name != "Shiraz" {
		t.Errorf("Name = %q, want Shiraz", view.Name)
	}
	if view.Body != 5 || view.Tannin != 4 || view.Acidity != 2 || view.Sweetness != 1 || view.Alcohol != 5 {
		t.Errorf("axes = %+v, want 5/4/2/1/5", view)
	}
	if len(view.Notes) != 2 || view.Notes[0] != "blackberry" {
		t.Errorf("Notes = %v, want [blackberry pepper]", view.Notes)
	}
	if view.Confirmed {
		t.Errorf("Confirmed = true, want false for a default-provenance bundle")
	}
}

func TestGetVariety_UnknownIsNotFound(t *testing.T) {
	repo := memory.NewVarietyRepo()
	h := app.NewGetVarietyHandler(repo)
	if _, err := h.Handle(context.Background(), domain.NewID()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestEditCharacteristics_PersistsAndMarksConfirmed(t *testing.T) {
	repo := memory.NewVarietyRepo()
	v, _ := domain.NewVariety("Shiraz")
	repo.Save(v)

	h := app.NewEditCharacteristicsHandler(repo)
	err := h.Handle(context.Background(), app.EditCharacteristicsCommand{
		VarietyID: v.ID,
		Body:      5, Tannin: 4, Acidity: 2, Sweetness: 1, Alcohol: 5,
		Notes: []string{"blackberry", "pepper"},
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	got, _ := repo.GetCharacteristics(context.Background(), v.ID)
	if got.Body.Int() != 5 || got.Tannin.Int() != 4 {
		t.Errorf("axes not persisted: %+v", got)
	}
	if !got.IsConfirmed() {
		t.Errorf("editing should mark Provenance confirmed; got %v", got.Provenance)
	}
	if len(got.Notes) != 2 {
		t.Errorf("notes not persisted: %v", got.Notes)
	}
}

func TestEditCharacteristics_RejectsOutOfScale(t *testing.T) {
	repo := memory.NewVarietyRepo()
	v, _ := domain.NewVariety("Shiraz")
	repo.Save(v)

	h := app.NewEditCharacteristicsHandler(repo)
	err := h.Handle(context.Background(), app.EditCharacteristicsCommand{
		VarietyID: v.ID, Body: 9, Tannin: 1, Acidity: 1, Sweetness: 1, Alcohol: 1,
	})
	if !errors.Is(err, domain.ErrInvalidCharacteristics) {
		t.Errorf("err = %v, want ErrInvalidCharacteristics", err)
	}
}

func TestEditCharacteristics_UnknownVarietyIsNotFound(t *testing.T) {
	repo := memory.NewVarietyRepo()
	h := app.NewEditCharacteristicsHandler(repo)
	err := h.Handle(context.Background(), app.EditCharacteristicsCommand{
		VarietyID: domain.NewID(), Body: 3, Tannin: 3, Acidity: 3, Sweetness: 3, Alcohol: 3,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
