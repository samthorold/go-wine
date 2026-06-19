package app_test

import (
	"context"
	"testing"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
	"go-wine/internal/seed"
)

// SeedStyleCompositions fills a default Composition on a Wine whose Style has a
// conventional default and whose grapes aren't set yet.
func TestSeedStyleCompositions_FillsDefaultByStyle(t *testing.T) {
	ctx := context.Background()
	varieties, byName := seedGrapes(t, "Chardonnay")
	wines := memory.NewWineRepo()
	w, _ := domain.NewWine("William Fèvre", "Chablis", "Chablis")
	wines.Save(w)

	if err := app.SeedStyleCompositions(ctx, wines, varieties, seed.StyleCompositions()); err != nil {
		t.Fatalf("SeedStyleCompositions: %v", err)
	}

	got, _ := wines.Get(ctx, w.ID)
	if got.Composition.IsZero() {
		t.Fatal("expected the Chablis default Composition to be filled")
	}
	if got.Composition.Provenance != domain.ProvenanceDefault {
		t.Errorf("Provenance = %v, want default", got.Composition.Provenance)
	}
	if got.Composition.Parts[0].VarietyID != byName["Chardonnay"] {
		t.Errorf("Chablis default should be Chardonnay; got %+v", got.Composition.Parts)
	}
}

// THE no-clobber rule end-to-end: a Wine whose Composition the Drinker confirmed
// survives a Style re-seed untouched.
func TestSeedStyleCompositions_NeverClobbersConfirmed(t *testing.T) {
	ctx := context.Background()
	varieties, byName := seedGrapes(t, "Chardonnay", "Sauvignon Blanc")
	wines := memory.NewWineRepo()
	w, _ := domain.NewWine("William Fèvre", "Chablis", "Chablis")
	wines.Save(w)

	// The Drinker insists this "Chablis" is actually 100% Sauvignon Blanc.
	confirmed, _ := domain.NewComposition(
		[]domain.CompositionPart{{VarietyID: byName["Sauvignon Blanc"], Proportion: 100}},
		domain.ProvenanceConfirmed,
	)
	_ = wines.SetComposition(ctx, w.ID, confirmed)

	if err := app.SeedStyleCompositions(ctx, wines, varieties, seed.StyleCompositions()); err != nil {
		t.Fatalf("SeedStyleCompositions: %v", err)
	}

	got, _ := wines.Get(ctx, w.ID)
	if !got.Composition.IsConfirmed() || got.Composition.Parts[0].VarietyID != byName["Sauvignon Blanc"] {
		t.Errorf("confirmed Composition was clobbered by the re-seed: %+v", got.Composition)
	}
}

// A Wine whose Style has no conventional default is left untouched.
func TestSeedStyleCompositions_SkipsUnknownStyle(t *testing.T) {
	ctx := context.Background()
	varieties, _ := seedGrapes(t, "Chardonnay")
	wines := memory.NewWineRepo()
	w, _ := domain.NewWine("Some", "Wine", "Mystery Style")
	wines.Save(w)

	if err := app.SeedStyleCompositions(ctx, wines, varieties, seed.StyleCompositions()); err != nil {
		t.Fatalf("SeedStyleCompositions: %v", err)
	}

	got, _ := wines.Get(ctx, w.ID)
	if !got.Composition.IsZero() {
		t.Errorf("a Wine with no conventional Style default should be left unset; got %+v", got.Composition)
	}
}
