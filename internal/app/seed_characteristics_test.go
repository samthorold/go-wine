package app_test

import (
	"context"
	"testing"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

// seedRubric is a tiny two-grape rubric for the seed tests, independent of the
// production rubric so the test isn't coupled to specific grape values.
func seedRubric() []app.VarietySeed {
	return []app.VarietySeed{
		{Name: "Shiraz", Body: 5, Tannin: 4, Acidity: 2, Sweetness: 1, Alcohol: 5, Notes: []string{"blackberry"}},
		{Name: "Riesling", Body: 2, Tannin: 1, Acidity: 5, Sweetness: 3, Alcohol: 2, Notes: []string{"lime"}},
	}
}

func TestSeedCharacteristics_PopulatesDefaults(t *testing.T) {
	repo := memory.NewVarietyRepo()
	shiraz, _ := domain.NewVariety("Shiraz")
	riesling, _ := domain.NewVariety("Riesling")
	repo.Save(shiraz)
	repo.Save(riesling)

	if err := app.SeedCharacteristics(context.Background(), repo, seedRubric()); err != nil {
		t.Fatalf("SeedCharacteristics: %v", err)
	}

	got, _ := repo.GetCharacteristics(context.Background(), shiraz.ID)
	if got.Body.Int() != 5 || got.Provenance != domain.ProvenanceDefault {
		t.Errorf("Shiraz seeded = %+v, want body 5 / default provenance", got)
	}
}

func TestSeedCharacteristics_IsIdempotentOverDefaults(t *testing.T) {
	repo := memory.NewVarietyRepo()
	shiraz, _ := domain.NewVariety("Shiraz")
	repo.Save(shiraz)
	rubric := seedRubric()

	for i := 0; i < 3; i++ {
		if err := app.SeedCharacteristics(context.Background(), repo, rubric); err != nil {
			t.Fatalf("SeedCharacteristics pass %d: %v", i, err)
		}
	}
	got, _ := repo.GetCharacteristics(context.Background(), shiraz.ID)
	if got.Body.Int() != 5 || got.Provenance != domain.ProvenanceDefault {
		t.Errorf("re-seed should be idempotent over defaults; got %+v", got)
	}
}

// THE acceptance-criteria test: a confirmed value survives a re-seed untouched.
func TestSeedCharacteristics_NeverClobbersConfirmed(t *testing.T) {
	repo := memory.NewVarietyRepo()
	shiraz, _ := domain.NewVariety("Shiraz")
	repo.Save(shiraz)

	// The Drinker hand-edits Shiraz, confirming it.
	edit := app.NewEditCharacteristicsHandler(repo)
	if err := edit.Handle(context.Background(), app.EditCharacteristicsCommand{
		VarietyID: shiraz.ID, Body: 1, Tannin: 1, Acidity: 1, Sweetness: 1, Alcohol: 1, Notes: []string{"mine"},
	}); err != nil {
		t.Fatalf("edit: %v", err)
	}

	// A re-seed runs with different default values for Shiraz.
	if err := app.SeedCharacteristics(context.Background(), repo, seedRubric()); err != nil {
		t.Fatalf("SeedCharacteristics: %v", err)
	}

	got, _ := repo.GetCharacteristics(context.Background(), shiraz.ID)
	if got.Body.Int() != 1 || !got.IsConfirmed() {
		t.Errorf("re-seed clobbered a confirmed value: %+v", got)
	}
	if len(got.Notes) != 1 || got.Notes[0] != "mine" {
		t.Errorf("re-seed clobbered confirmed notes: %v", got.Notes)
	}
}
